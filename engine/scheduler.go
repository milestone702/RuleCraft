package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"rulecraft/config"
	"rulecraft/plugin"
)

// ============================================================================
// TaskScheduler — 任务调度执行器
// ============================================================================

// TaskScheduler 管理单条任务的完整四阶段执行周期：
//   data_sources → condition → outputs
type TaskScheduler struct {
	runner            *Runner
	loader            *PluginManager
	evaluator         *ConditionEvaluator
	edgeTracker       *EdgeTriggerTracker
	mu                sync.Mutex
	thresholdCounters map[string]int       // triggerKey → 连续成立次数
	cooldowns         map[string]time.Time // cooldownKey → 上次触发时间
	lastCount         int
}

// NewTaskScheduler 创建任务调度器。
func NewTaskScheduler(runner *Runner, loader *PluginManager) *TaskScheduler {
	return &TaskScheduler{
		runner:            runner,
		loader:            loader,
		evaluator:         NewConditionEvaluator(),
		edgeTracker:       NewEdgeTriggerTracker(),
		thresholdCounters: make(map[string]int),
		cooldowns:         make(map[string]time.Time),
	}
}

// ExecuteTask 执行一条任务的完整周期：data_sources → condition → outputs。
func (ts *TaskScheduler) ExecuteTask(ctx context.Context, task *config.TaskDefinition) error {
	// 阶段 0: 将 data_source params 应用到可配置输入插件
	ts.applyDataSourceConfigs(task)

	states := ts.runner.GetStates()

	// 阶段 1: 校验数据源是否就绪（只在缺失时打日志，避免每轮刷屏）
	for _, src := range task.DataSources {
		if _, exists := states[src.StateKey]; !exists {
			log.Printf("[scheduler] task %s: data source %s (state_key=%s) not ready", task.TaskID, src.ID, src.StateKey)
		}
	}

	// 阶段 2: 评估条件树 condition
	conditionResult := true
	if task.Condition != nil {
		result, err := ts.evaluator.Eval(task.Condition, states)
		if err != nil {
			return fmt.Errorf("task %s: evaluate condition failed: %w", task.TaskID, err)
		}
		conditionResult = result
	}

	// 阶段 3: 边缘触发检测 + 执行输出动作 outputs
	ts.executeOutputs(task.TaskID, task.Outputs, conditionResult)

	return nil
}

// applyDataSourceConfigs 把任务 data_sources 中的 params 配置到对应输入插件。
func (ts *TaskScheduler) applyDataSourceConfigs(task *config.TaskDefinition) {
	for _, src := range task.DataSources {
		if src.PluginID == "" || len(src.Params) == 0 {
			continue
		}
		p, ok := ts.runner.registry.GetInput(src.PluginID)
		if !ok {
			continue
		}
		if cfg, ok := p.(plugin.ConfigurableInput); ok {
			if err := cfg.ConfigureFromMap(src.Params); err != nil {
				log.Printf("[scheduler] task %s: configure plugin %s failed: %v", task.TaskID, src.PluginID, err)
			}
		}
	}
}

// executeOutputs 检测边缘触发并执行输出动作（支持触发阈值 + 冷却）。
//
// 阈值与边缘触发的正确组合：
//  1. 先用阈值滤波得到 effective 结果（连续 N 次 true 才变 true，变 false 立即清零）
//  2. 再对 effective 做边缘触发检测
// 这样避免「边缘已被消费但阈值未满 → 永远不触发」的 bug。
func (ts *TaskScheduler) executeOutputs(taskID string, outputs []config.OutputBinding, conditionResult bool) {
	for _, binding := range outputs {
		triggerKey := taskID + "/" + binding.ID

		// 1. 阈值滤波
		effective := ts.applyThreshold(triggerKey, conditionResult, binding.TriggerThreshold)

		// 2. 边缘触发
		direction := ts.edgeTracker.Check(triggerKey, effective)

		// 判断是否应该执行
		shouldExecute := false
		switch binding.TriggerOn {
		case "true":
			shouldExecute = direction == TriggerTrue
		case "false":
			shouldExecute = direction == TriggerFalse
		case "both", "":
			shouldExecute = direction != TriggerNone
		default:
			shouldExecute = direction == TriggerTrue
		}

		if !shouldExecute {
			continue
		}

		// 3. 冷却时间
		if !ts.allowFire(triggerKey, binding) {
			log.Printf("[scheduler] %s: in cooldown, skip", triggerKey)
			continue
		}

		// 执行输出
		if err := ts.executeBuiltinOutput(binding, direction); err != nil {
			if direction == TriggerFalse && errors.Is(err, plugin.ErrResetNotSupported) {
				continue
			}
			log.Printf("[scheduler] output %s failed: %v", binding.ID, err)
		}
	}
}

// applyThreshold 连续 N 次成立后 effective 为 true；任一次不成立立即清零。
func (ts *TaskScheduler) applyThreshold(key string, conditionResult bool, threshold int) bool {
	if threshold <= 0 {
		return conditionResult
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	if conditionResult {
		ts.thresholdCounters[key]++
		return ts.thresholdCounters[key] >= threshold
	}
	ts.thresholdCounters[key] = 0
	return false
}

// allowFire 检查冷却是否已过期，未过期返回 false；通过则记录本次触发时间。
func (ts *TaskScheduler) allowFire(triggerKey string, binding config.OutputBinding) bool {
	cooldown := binding.Conditions.CooldownSeconds
	if cooldown <= 0 {
		return true
	}

	key := binding.Conditions.CooldownKey
	if key == "" {
		key = triggerKey
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	if last, ok := ts.cooldowns[key]; ok {
		if time.Since(last) < time.Duration(cooldown)*time.Second {
			return false
		}
	}
	ts.cooldowns[key] = time.Now()
	return true
}

// executeBuiltinOutput 执行内置输出插件。
func (ts *TaskScheduler) executeBuiltinOutput(binding config.OutputBinding, direction TriggerDirection) error {
	plugin, ok := ts.runner.registry.GetOutput(binding.PluginID)
	if !ok {
		return fmt.Errorf("output plugin not found: %s", binding.PluginID)
	}

	// 模板变量渲染 params
	params := resolveTemplateVars(binding.Params, ts.runner.GetStates())

	if direction == TriggerTrue {
		return plugin.Execute(params)
	}
	return plugin.Reset(params)
}

// ============================================================================
// 模板变量渲染
// ============================================================================

// resolveTemplateVars 对 params 中的 {{state_key}} 模板变量进行替换。
func resolveTemplateVars(params map[string]interface{}, states map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(params))
	for k, v := range params {
		result[k] = resolveValue(v, states)
	}
	return result
}

// resolveValue 递归替换单个值中的模板变量。
// 支持整串 "{{key}}" 替换，以及字符串内嵌 {{key}} 的局部替换。
func resolveValue(v interface{}, states map[string]interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return resolveStringTemplate(val, states)
	case map[string]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, innerV := range val {
			result[k] = resolveValue(innerV, states)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = resolveValue(item, states)
		}
		return result
	default:
		return v
	}
}

// resolveStringTemplate 替换字符串中的 {{key}} 占位符。
// 若整串就是单个占位符且状态值为非字符串，则保留原始类型。
func resolveStringTemplate(s string, states map[string]interface{}) interface{} {
	if len(s) < 5 || !strings.Contains(s, "{{") {
		return s
	}

	// 整串单占位符：保留类型
	if strings.HasPrefix(s, "{{") && strings.HasSuffix(s, "}}") && strings.Count(s, "{{") == 1 {
		key := strings.TrimSpace(s[2 : len(s)-2])
		if sv, ok := states[key]; ok {
			return sv
		}
		return ""
	}

	// 内嵌占位符：全部转为字符串
	var b strings.Builder
	for {
		start := strings.Index(s, "{{")
		if start < 0 {
			b.WriteString(s)
			break
		}
		end := strings.Index(s[start:], "}}")
		if end < 0 {
			b.WriteString(s)
			break
		}
		end += start
		b.WriteString(s[:start])
		key := strings.TrimSpace(s[start+2 : end])
		if sv, ok := states[key]; ok {
			b.WriteString(fmt.Sprintf("%v", sv))
		}
		s = s[end+2:]
	}
	return b.String()
}

// ============================================================================
// 批量任务执行
// ============================================================================

// ExecuteAllTasks 加载并执行所有任务（带 mtime 缓存，未变更的文件不重复解析）。
func (ts *TaskScheduler) ExecuteAllTasks(ctx context.Context) {
	tasks, err := ts.loader.LoadAllTasksCached()
	if err != nil {
		log.Printf("[scheduler] load all tasks failed: %v", err)
		return
	}

	enabled := 0
	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		enabled++

		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := ts.ExecuteTask(ctx, task); err != nil {
			log.Printf("[scheduler] task %s: %v", task.TaskID, err)
		}
	}

	// 仅在任务数变化时打一条摘要，避免每轮刷屏
	if n := len(tasks); n != ts.lastTaskCount() {
		ts.setLastTaskCount(n)
		log.Printf("[scheduler] loaded %d tasks (%d enabled)", n, enabled)
	}
}

func (ts *TaskScheduler) lastTaskCount() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.lastCount
}

func (ts *TaskScheduler) setLastTaskCount(n int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.lastCount = n
}
