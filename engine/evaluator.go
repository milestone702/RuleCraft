// Package engine 实现 RuleCraft 自动化引擎的核心逻辑：
//   - 条件树递归求值 (evaluator.go)
//   - 轮询引擎 + 边缘触发 (runner.go)
//   - 处理器管道 DAG 调度 (scheduler.go)
//   - 子进程管理 (script_runner.go)
package engine

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"

	"rulecraft/config"
)

// ============================================================================
// 条件求值器
// ============================================================================

// ConditionEvaluator 负责递归求值 ConditionNode 树。
type ConditionEvaluator struct {
	mu               sync.Mutex
	thresholdCounters map[string]int       // nodeKey → 连续成立次数
	stableTimers      map[string]time.Time // nodeKey → 首次成立时间
	prevValues        map[string]interface{} // nodeKey → 上一轮状态值（changed/increased 等）
	regexCache        map[string]*regexp.Regexp
}

// NewConditionEvaluator 创建条件求值器。
func NewConditionEvaluator() *ConditionEvaluator {
	return &ConditionEvaluator{
		thresholdCounters: make(map[string]int),
		stableTimers:      make(map[string]time.Time),
		prevValues:        make(map[string]interface{}),
		regexCache:        make(map[string]*regexp.Regexp),
	}
}

// nodeKey 为条件节点生成唯一标识键（用于阈值/稳定追踪）。
func nodeKey(node *config.ConditionNode) string {
	return fmt.Sprintf("%s|%s|%s|%v", node.Type, node.StateKey, node.Operator, node.Value)
}

// Eval 递归求值条件树，返回最终布尔结果。
// states 是当前的全局状态快照。
func (ce *ConditionEvaluator) Eval(node *config.ConditionNode, states map[string]interface{}) (bool, error) {
	if node == nil {
		return true, nil
	}

	// 组合节点
	if node.LogicalOp != "" {
		return ce.evalComposite(node, states)
	}

	// 叶子节点（原子判断）
	return ce.evalLeaf(node, states)
}

// evalComposite 求值组合节点（AND / OR / NOT）。
func (ce *ConditionEvaluator) evalComposite(node *config.ConditionNode, states map[string]interface{}) (bool, error) {
	switch strings.ToUpper(node.LogicalOp) {
	case "NOT":
		// NOT 只有一个子节点
		if len(node.Children) != 1 {
			return false, fmt.Errorf("NOT operator requires exactly 1 child, got %d", len(node.Children))
		}
		result, err := ce.Eval(&node.Children[0], states)
		if err != nil {
			return false, fmt.Errorf("NOT child eval error: %w", err)
		}
		return !result, nil

	case "AND":
		for i := range node.Children {
			result, err := ce.Eval(&node.Children[i], states)
			if err != nil {
				return false, fmt.Errorf("AND child[%d] eval error: %w", i, err)
			}
			if !result {
				return false, nil // 短路求值
			}
		}
		return true, nil

	case "OR":
		for i := range node.Children {
			result, err := ce.Eval(&node.Children[i], states)
			if err != nil {
				return false, fmt.Errorf("OR child[%d] eval error: %w", i, err)
			}
			if result {
				return true, nil // 短路求值
			}
		}
		return false, nil

	default:
		return false, fmt.Errorf("unknown logical operator: %s", node.LogicalOp)
	}
}

// evalLeaf 求值叶子节点（原子条件判断），支持阈值和稳定时间滤波。
func (ce *ConditionEvaluator) evalLeaf(node *config.ConditionNode, states map[string]interface{}) (bool, error) {
	var rawResult bool
	var err error
	switch node.Type {
	case "state":
		rawResult, err = ce.evalStateCondition(node, states)
	case "constant":
		rawResult, err = ce.evalConstant(node)
	default:
		return false, fmt.Errorf("unknown condition type: %s", node.Type)
	}
	if err != nil {
		return false, err
	}

	// 阈值/稳定时间滤波（二选一）
	key := nodeKey(node)

	// 连续 N 次成立阈值
	if node.Threshold > 0 {
		ce.mu.Lock()
		defer ce.mu.Unlock()
		if rawResult {
			ce.thresholdCounters[key]++
			if ce.thresholdCounters[key] < node.Threshold {
				return false, nil // 未达阈值
			}
			// 已达阈值，保持为 true 直到条件变 false
		} else {
			delete(ce.thresholdCounters, key)
		}
		return rawResult, nil
	}

	// 持续 N 秒成立
	if node.StableSeconds > 0 {
		ce.mu.Lock()
		defer ce.mu.Unlock()
		if rawResult {
			startAt, hasTimer := ce.stableTimers[key]
			if !hasTimer {
				ce.stableTimers[key] = time.Now()
				return false, nil // 刚成立，等待稳定
			}
			elapsed := int(time.Since(startAt).Seconds())
			if elapsed < node.StableSeconds {
				return false, nil // 尚未达到稳定时间
			}
			// 稳定时间已满足，清除计时器
			delete(ce.stableTimers, key)
		} else {
			delete(ce.stableTimers, key)
		}
		return rawResult, nil
	}

	return rawResult, nil
}

// evalStateCondition 求值状态条件。
func (ce *ConditionEvaluator) evalStateCondition(node *config.ConditionNode, states map[string]interface{}) (bool, error) {
	stateValue, exists := states[node.StateKey]

	switch node.Operator {
	// ---- 存在性判断 ----
	case "exists":
		return exists, nil
	case "not_exists":
		return !exists, nil

	// ---- 空值/非空 ----
	case "is_empty":
		if !exists {
			return true, nil
		}
		s, ok := toString(stateValue)
		if !ok {
			return false, nil
		}
		return s == "", nil

	case "is_not_empty":
		if !exists {
			return false, nil
		}
		s, ok := toString(stateValue)
		if !ok {
			return true, nil // 非字符串值视为非空
		}
		return s != "", nil

	// ---- 布尔判断 ----
	case "is_true":
		return exists && toBoolLoose(stateValue), nil
	case "is_false":
		return !exists || !toBoolLoose(stateValue), nil

	// ---- 变化检测（需要上一轮值）----
	case "changed":
		return ce.compareChanged(node, stateValue, exists, false)
	case "not_changed":
		return ce.compareChanged(node, stateValue, exists, true)
	case "increased":
		return ce.compareDelta(node, stateValue, exists, false)
	case "decreased":
		return ce.compareDelta(node, stateValue, exists, true)
	}

	// 如果状态键不存在，非存在性操作符返回 false
	if !exists {
		// 仍需更新 prevValues 以便后续轮次能检测变化
		if isHistoryOperator(node.Operator) {
			ce.mu.Lock()
			ce.prevValues[nodeKey(node)] = nil
			ce.mu.Unlock()
		}
		return false, nil
	}

	switch node.Operator {
	// ---- 相等/不等 ----
	case "equals":
		return compareEquals(stateValue, node.Value), nil
	case "not_equals":
		return !compareEquals(stateValue, node.Value), nil
	case "equals_ignore_case":
		return compareEqualsIgnoreCase(stateValue, node.Value), nil

	// ---- 字符串包含 ----
	case "contains":
		s, a := toString(stateValue)
		t, _ := toString(node.Value)
		if !a || t == "" {
			return false, nil
		}
		return strings.Contains(s, t), nil

	case "not_contains":
		s, a := toString(stateValue)
		t, _ := toString(node.Value)
		if !a || t == "" {
			return true, nil
		}
		return !strings.Contains(s, t), nil

	case "contains_ignore_case":
		s, a := toString(stateValue)
		t, _ := toString(node.Value)
		if !a || t == "" {
			return false, nil
		}
		return strings.Contains(strings.ToLower(s), strings.ToLower(t)), nil

	// ---- 数值比较 ----
	case "greater_than":
		return compareNumeric(stateValue, node.Value, func(a, b float64) bool { return a > b }), nil
	case "less_than":
		return compareNumeric(stateValue, node.Value, func(a, b float64) bool { return a < b }), nil
	case "greater_equal":
		return compareNumeric(stateValue, node.Value, func(a, b float64) bool { return a >= b }), nil
	case "less_equal":
		return compareNumeric(stateValue, node.Value, func(a, b float64) bool { return a <= b }), nil

	// ---- 区间 ----
	case "between":
		return compareBetween(stateValue, node.Value)
	case "not_between":
		ok, err := compareBetween(stateValue, node.Value)
		if err != nil {
			return false, err
		}
		return !ok, nil

	// ---- 字符串模式匹配 ----
	case "matches_regex":
		return ce.compareRegex(stateValue, node.Value, false)
	case "not_matches_regex":
		return ce.compareRegex(stateValue, node.Value, true)

	case "starts_with":
		s, a := toString(stateValue)
		t, _ := toString(node.Value)
		if !a {
			return false, nil
		}
		return strings.HasPrefix(s, t), nil

	case "ends_with":
		s, a := toString(stateValue)
		t, _ := toString(node.Value)
		if !a {
			return false, nil
		}
		return strings.HasSuffix(s, t), nil

	// ---- 字符串长度 ----
	case "length_equals":
		s, a := toString(stateValue)
		n, ok := toFloat(node.Value)
		if !a || !ok {
			return false, nil
		}
		return float64(len(s)) == n, nil

	case "length_greater_than":
		s, a := toString(stateValue)
		n, ok := toFloat(node.Value)
		if !a || !ok {
			return false, nil
		}
		return float64(len(s)) > n, nil

	case "length_less_than":
		s, a := toString(stateValue)
		n, ok := toFloat(node.Value)
		if !a || !ok {
			return false, nil
		}
		return float64(len(s)) < n, nil

	// ---- 枚举 ----
	case "in":
		return compareIn(stateValue, node.Value, false)
	case "not_in":
		return compareIn(stateValue, node.Value, true)

	default:
		return false, fmt.Errorf("unknown operator: %s", node.Operator)
	}
}

// isHistoryOperator 判断运算符是否依赖历史值。
func isHistoryOperator(op string) bool {
	switch op {
	case "changed", "not_changed", "increased", "decreased":
		return true
	}
	return false
}

// compareChanged 比较当前值与上一轮值是否发生变化。
func (ce *ConditionEvaluator) compareChanged(node *config.ConditionNode, stateValue interface{}, exists, negate bool) (bool, error) {
	key := nodeKey(node)
	ce.mu.Lock()
	prev, hadPrev := ce.prevValues[key]
	// 首轮没有历史，先记录并返回 false（不触发）
	if !hadPrev {
		ce.prevValues[key] = stateValue
		ce.mu.Unlock()
		return false, nil
	}
	ce.prevValues[key] = stateValue
	ce.mu.Unlock()

	changed := !exists || !compareEquals(prev, stateValue)
	if negate {
		return !changed, nil
	}
	return changed, nil
}

// compareDelta 比较数值是否增加/减少（可带最小变化量）。
func (ce *ConditionEvaluator) compareDelta(node *config.ConditionNode, stateValue interface{}, exists, wantDecrease bool) (bool, error) {
	key := nodeKey(node)
	cur, ok := toFloat(stateValue)
	if !exists || !ok {
		ce.mu.Lock()
		delete(ce.prevValues, key)
		ce.mu.Unlock()
		return false, nil
	}

	ce.mu.Lock()
	prevRaw, hadPrev := ce.prevValues[key]
	ce.prevValues[key] = stateValue
	ce.mu.Unlock()

	if !hadPrev {
		return false, nil
	}
	prev, ok := toFloat(prevRaw)
	if !ok {
		return false, nil
	}

	minDelta := 0.0
	if node.Value != nil {
		if d, ok := toFloat(node.Value); ok {
			minDelta = math.Abs(d)
		}
	}

	diff := cur - prev
	if wantDecrease {
		diff = -diff
	}
	return diff > minDelta || (minDelta == 0 && diff > 0), nil
}

// evalConstant 求值常量条件（用于测试或占位）。
func (ce *ConditionEvaluator) evalConstant(node *config.ConditionNode) (bool, error) {
	b, ok := node.Value.(bool)
	if !ok {
		return false, fmt.Errorf("constant condition value must be boolean")
	}
	return b, nil
}

// ============================================================================
// 比较辅助函数
// ============================================================================

// compareEquals 宽松相等比较。
func compareEquals(a, b interface{}) bool {
	// 如果类型相同，直接比较
	switch va := a.(type) {
	case bool:
		vb, ok := b.(bool)
		if ok {
			return va == vb
		}
		// bool vs string: "true"/"false"
		if s, ok := b.(string); ok {
			return (va && s == "true") || (!va && s == "false")
		}
	case float64:
		switch vb := b.(type) {
		case float64:
			return va == vb
		case int:
			return va == float64(vb)
		case string:
			// 尝试解析数字字符串
			if f, err := parseFloat(vb); err == nil {
				return va == f
			}
			return fmt.Sprintf("%v", va) == vb
		}
	case string:
		switch vb := b.(type) {
		case string:
			return va == vb
		case float64:
			if f, err := parseFloat(va); err == nil {
				return f == vb
			}
			return va == fmt.Sprintf("%v", vb)
		case bool:
			return (va == "true" && vb) || (va == "false" && !vb)
		}
	case int:
		switch vb := b.(type) {
		case int:
			return va == vb
		case float64:
			return float64(va) == vb
		case string:
			if f, err := parseFloat(vb); err == nil {
				return float64(va) == f
			}
			return fmt.Sprintf("%d", va) == vb
		}
	}

	// 兜底：字符串化后比较
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// compareEqualsIgnoreCase 忽略大小写的宽松相等比较。
func compareEqualsIgnoreCase(a, b interface{}) bool {
	sa, oka := toString(a)
	sb, okb := toString(b)
	if !oka || !okb {
		return compareEquals(a, b)
	}
	return strings.EqualFold(sa, sb)
}

// compareNumeric 数值比较，使用给定的比较函数。
func compareNumeric(stateVal, targetVal interface{}, cmp func(float64, float64) bool) bool {
	a, ok := toFloat(stateVal)
	if !ok {
		return false
	}
	b, ok := toFloat(targetVal)
	if !ok {
		return false
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	return cmp(a, b)
}

// compareBetween 判断数值是否落在 [min, max] 闭区间。
// value 支持 [min, max] 数组，或 "min,max" 字符串。
func compareBetween(stateVal, rangeVal interface{}) (bool, error) {
	a, ok := toFloat(stateVal)
	if !ok {
		return false, nil
	}

	var minV, maxV float64
	switch r := rangeVal.(type) {
	case []interface{}:
		if len(r) != 2 {
			return false, fmt.Errorf("between requires [min, max], got %d elements", len(r))
		}
		minV, ok = toFloat(r[0])
		if !ok {
			return false, fmt.Errorf("between: invalid min value")
		}
		maxV, ok = toFloat(r[1])
		if !ok {
			return false, fmt.Errorf("between: invalid max value")
		}
	case string:
		parts := strings.Split(r, ",")
		if len(parts) != 2 {
			return false, fmt.Errorf("between: value must be \"min,max\" or [min,max]")
		}
		mn, err1 := parseFloat(strings.TrimSpace(parts[0]))
		mx, err2 := parseFloat(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			return false, fmt.Errorf("between: invalid range %q", r)
		}
		minV, maxV = mn, mx
	default:
		return false, fmt.Errorf("between requires array or \"min,max\" string value")
	}

	if minV > maxV {
		minV, maxV = maxV, minV
	}
	return a >= minV && a <= maxV, nil
}

// compareRegex 正则匹配比较（带缓存）。
func (ce *ConditionEvaluator) compareRegex(stateVal, pattern interface{}, negate bool) (bool, error) {
	s, ok := toString(stateVal)
	if !ok {
		return false, nil
	}
	p, ok := toString(pattern)
	if !ok {
		return false, fmt.Errorf("regex pattern must be a string")
	}

	re, err := ce.getCompiledRegex(p)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %w", err)
	}

	matched := re.MatchString(s)
	if negate {
		return !matched, nil
	}
	return matched, nil
}

// getCompiledRegex 获取缓存的正则，无缓存时编译并写入。
func (ce *ConditionEvaluator) getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	if re, ok := ce.regexCache[pattern]; ok {
		return re, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	// 简单容量控制，避免无限增长
	if len(ce.regexCache) > 256 {
		ce.regexCache = make(map[string]*regexp.Regexp)
	}
	ce.regexCache[pattern] = re
	return re, nil
}

// compareIn 检查值是否在枚举列表中。
func compareIn(stateVal, list interface{}, negate bool) (bool, error) {
	slice, ok := list.([]interface{})
	if !ok {
		return false, fmt.Errorf("in operator requires an array value")
	}
	for _, item := range slice {
		if compareEquals(stateVal, item) {
			return !negate, nil
		}
	}
	return negate, nil
}

// ============================================================================
// 类型转换辅助
// ============================================================================

// toString 将值转换为字符串（宽松转换）。
func toString(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}
	switch s := v.(type) {
	case string:
		return s, true
	case fmt.Stringer:
		return s.String(), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

// toFloat 将值转换为 float64。
func toFloat(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case string:
		f, err := parseFloat(n)
		return f, err == nil
	default:
		return 0, false
	}
}

// toBoolLoose 宽松布尔转换。
func toBoolLoose(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		s := strings.ToLower(strings.TrimSpace(b))
		return s == "true" || s == "1" || s == "yes" || s == "on"
	case float64:
		return b != 0
	case int:
		return b != 0
	case int64:
		return b != 0
	default:
		return false
	}
}

// parseFloat 尝试解析浮点数。
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
