package engine

import (
	"context"
	"fmt"
	"log"
	"sync"

	"rulecraft/config"
)

// ============================================================================
// TaskScheduler — 任务调度执行器
// ============================================================================

// TaskScheduler 管理单条任务的完整四阶段执行周期：
//   data_sources → processors → condition → outputs
type TaskScheduler struct {
	runner      *Runner
	loader      *PluginManager
	evaluator   *ConditionEvaluator
	edgeTracker *EdgeTriggerTracker
	mu          sync.Mutex
	thresholdCounters map[string]int // triggerKey → 连续成立次数
}

// NewTaskScheduler 创建任务调度器。
func NewTaskScheduler(runner *Runner, loader *PluginManager) *TaskScheduler {
	return &TaskScheduler{
		runner:      runner,
		loader:      loader,
		evaluator:   NewConditionEvaluator(),
		edgeTracker: NewEdgeTriggerTracker(),
		thresholdCounters: make(map[string]int),
	}
}

// ExecuteTask 执行一条任务的完整周期：data_sources → condition → outputs。
func (ts *TaskScheduler) ExecuteTask(ctx context.Context, task *config.TaskDefinition) error {
	states := ts.runner.GetStates()

	// 阶段 1: 读取数据源 data_sources → 从全局状态取值（仅日志，不做映射）
	for _, src := range task.DataSources {
		val, exists := states[src.StateKey]
		if !exists {
			log.Printf("[scheduler] data source %s (state_key=%s) not found in global state", src.ID, src.StateKey)
		} else {
			log.Printf("[scheduler] data source %s = %v", src.ID, val)
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

// executeOutputs 检测边缘触发并执行输出动作（支持触发阈值容错）。
func (ts *TaskScheduler) executeOutputs(taskID string, outputs []config.OutputBinding, conditionResult bool) {
	for _, binding := range outputs {
		triggerKey := taskID + "/" + binding.ID
		direction := ts.edgeTracker.Check(triggerKey, conditionResult)

		// 阈值容错：连续成立 N 次后才触发
		threshold := binding.TriggerThreshold
		if threshold > 0 && conditionResult {
			ts.mu.Lock()
			ts.thresholdCounters[triggerKey]++
			count := ts.thresholdCounters[triggerKey]
			ts.mu.Unlock()

			if count < threshold {
				log.Printf("[scheduler] %s: threshold %d/%d, logging only", triggerKey, count, threshold)
				continue
			}
		} else if threshold > 0 {
			ts.mu.Lock()
			ts.thresholdCounters[triggerKey] = 0
			ts.mu.Unlock()
		}

		// 判断是否应该执行
		shouldExecute := false
		switch binding.TriggerOn {
		case "true":
			shouldExecute = direction == TriggerTrue
		case "false":
			shouldExecute = direction == TriggerFalse
		case "both":
			shouldExecute = direction != TriggerNone
		}

		if !shouldExecute {
			continue
		}

		// 执行输出
		if err := ts.executeBuiltinOutput(binding, direction); err != nil {
			log.Printf("[scheduler] output %s failed: %v", binding.ID, err)
		}
	}
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
func resolveValue(v interface{}, states map[string]interface{}) interface{} {
	switch val := v.(type) {
	case string:
		// 简单模板替换：{{key}}
		if len(val) > 3 && val[:2] == "{{" && val[len(val)-1:] == "}" {
			key := val[2 : len(val)-1]
			if sv, ok := states[key]; ok {
				return sv
			}
			return ""
		}
		return val
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, innerV := range val {
			result[k] = resolveValue(innerV, states)
		}
		return result
	default:
		return v
	}
}

// ============================================================================
// 批量任务执行
// ============================================================================

// ExecuteAllTasks 加载并执行所有任务。
func (ts *TaskScheduler) ExecuteAllTasks(ctx context.Context) {
	tasks, err := ts.loader.LoadAllTasks()
	if err != nil {
		log.Printf("[scheduler] load all tasks failed: %v", err)
		return
	}

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := ts.ExecuteTask(ctx, task); err != nil {
			log.Printf("[scheduler] task %s: %v", task.TaskID, err)
		}
	}

	log.Printf("[scheduler] executed %d tasks", len(tasks))
}
