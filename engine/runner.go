package engine

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"rulecraft/config"
	"rulecraft/plugin"
)

// ============================================================================
// 边缘触发跟踪器
// ============================================================================

// EdgeTriggerTracker 跟踪每个输出绑定的上一次条件评估结果，实现边缘触发检测。
type EdgeTriggerTracker struct {
	mu          sync.Mutex
	lastTrigger map[string]bool // key: "task_id/output_id" → true=上次条件为 true
}

// NewEdgeTriggerTracker 创建新的边缘触发跟踪器。
func NewEdgeTriggerTracker() *EdgeTriggerTracker {
	return &EdgeTriggerTracker{
		lastTrigger: make(map[string]bool),
	}
}

// TriggerDirection 触发方向。
type TriggerDirection int

const (
	TriggerNone   TriggerDirection = iota // 无变化
	TriggerTrue                           // false → true（触发执行）
	TriggerFalse                          // true → false（触发重置）
)

// Check 检测指定输出绑定是否发生了边缘触发。
// key 格式为 "task_id/output_id"。
func (t *EdgeTriggerTracker) Check(key string, currentResult bool) TriggerDirection {
	t.mu.Lock()
	defer t.mu.Unlock()

	last, exists := t.lastTrigger[key]

	// 更新记录
	t.lastTrigger[key] = currentResult

	// 首次评估：记录但不触发
	if !exists {
		return TriggerNone
	}

	// 检测边缘
	if !last && currentResult {
		return TriggerTrue
	}
	if last && !currentResult {
		return TriggerFalse
	}
	return TriggerNone
}

// Reset 重置指定键的跟踪状态（用于冷却到期后重新触发）。
func (t *EdgeTriggerTracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.lastTrigger, key)
}

// ============================================================================
// 轮询引擎
// ============================================================================

// Runner 是自动化引擎的轮询执行器。
// 它周期性地采集输入数据、处理条件、检测边缘触发并执行输出动作。
type Runner struct {
	cfg       *config.AppConfig
	registry  *plugin.Registry
	evaluator *ConditionEvaluator
	edgeTrack *EdgeTriggerTracker
	scheduler *TaskScheduler
	states    map[string]interface{}
	mu        sync.RWMutex
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	running   bool

	// 回调事件
	onStateChange func(key string, oldVal, newVal interface{})
	onTrigger     func(taskID, outputID string, direction TriggerDirection)
}

// RunnerOption 轮询引擎配置选项。
type RunnerOption func(*Runner)

// WithStateChangeCallback 设置状态变更回调。
func WithStateChangeCallback(cb func(key string, oldVal, newVal interface{})) RunnerOption {
	return func(r *Runner) {
		r.onStateChange = cb
	}
}

// WithScheduler 设置任务调度器。
func WithScheduler(s *TaskScheduler) RunnerOption {
	return func(r *Runner) {
		r.scheduler = s
	}
}

// WithTriggerCallback 设置触发回调。
func WithTriggerCallback(cb func(taskID, outputID string, direction TriggerDirection)) RunnerOption {
	return func(r *Runner) {
		r.onTrigger = cb
	}
}

// SetScheduler 设置任务调度器（用于解决 Runner 和 TaskScheduler 的循环依赖）。
func (r *Runner) SetScheduler(s *TaskScheduler) {
	r.scheduler = s
}

// NewRunner 创建新的轮询引擎。
func NewRunner(cfg *config.AppConfig, registry *plugin.Registry, opts ...RunnerOption) *Runner {
	r := &Runner{
		cfg:       cfg,
		registry:  registry,
		evaluator: NewConditionEvaluator(),
		edgeTrack: NewEdgeTriggerTracker(),
		states:    make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Start 启动轮询引擎。在单独的 goroutine 中运行，直到 ctx 被取消或 Stop 被调用。
func (r *Runner) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("runner is already running")
	}
	r.running = true
	ctx, r.cancel = context.WithCancel(ctx)
	r.mu.Unlock()

	interval := time.Duration(r.cfg.PollInterval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	r.wg.Add(1)
	go r.runLoop(ctx, interval)

	log.Printf("[engine] runner started, poll interval: %v", interval)
	return nil
}

// Stop 优雅停止轮询引擎。
func (r *Runner) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}
	r.running = false
	if r.cancel != nil {
		r.cancel()
	}
	r.mu.Unlock()

	r.wg.Wait()
	log.Println("[engine] runner stopped")
}

// GetStates 返回当前状态的快照。
func (r *Runner) GetStates() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	clone := make(map[string]interface{}, len(r.states))
	for k, v := range r.states {
		clone[k] = v
	}
	return clone
}

// SetState 设置指定状态键的值。
func (r *Runner) SetState(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	oldVal := r.states[key]
	r.states[key] = value
	if r.onStateChange != nil {
		r.onStateChange(key, oldVal, value)
	}
}

// GetStateValue 获取指定状态键的值。
func (r *Runner) GetStateValue(key string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.states[key]
	return v, ok
}

// runLoop 是轮询主循环。
func (r *Runner) runLoop(ctx context.Context, interval time.Duration) {
	defer r.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 启动时立即执行一次
	r.doPoll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.doPoll(ctx)
		}
	}
}

// doPoll 执行一次完整的轮询周期。
func (r *Runner) doPoll(ctx context.Context) {
	// 1. 采集所有输入插件数据
	r.collectInputs(ctx)

	// 2. 处理所有加载的任务
	r.evaluateRules(ctx)
}

// 默认单插件采集超时。
const defaultCollectTimeout = 8 * time.Second

// 最大并行采集数。
const maxCollectConcurrency = 6

// collectInputs 并行遍历输入插件，各自写入独立 map，最后在锁内合并到全局状态。
// 使用独立 map 可避免多插件并发写共享状态造成 data race。
func (r *Runner) collectInputs(ctx context.Context) {
	type collectResult struct {
		id     string
		states map[string]interface{}
		err    error
	}

	plugins := r.registry.GetAllInputs()
	if len(plugins) == 0 {
		return
	}

	sem := make(chan struct{}, maxCollectConcurrency)
	results := make(chan collectResult, len(plugins))
	var wg sync.WaitGroup

	for id, inputPlugin := range plugins {
		if !inputPlugin.IsAvailable() {
			continue
		}

		wg.Add(1)
		go func(id string, inputPlugin plugin.InputPlugin) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			local := make(map[string]interface{})
			pluginCtx := &plugin.SystemContext{
				States: local,
				Ctx:    ctx,
			}

			cctx, cancel := context.WithTimeout(ctx, defaultCollectTimeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- inputPlugin.Collect(pluginCtx)
			}()

			var err error
			select {
			case err = <-done:
			case <-cctx.Done():
				err = fmt.Errorf("collect timeout after %v", defaultCollectTimeout)
			}

			results <- collectResult{id: id, states: local, err: err}
		}(id, inputPlugin)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	merged := make(map[string]interface{}, 32)
	for res := range results {
		if res.err != nil {
			log.Printf("[engine] input plugin %s collect error: %v", res.id, res.err)
		}
		for k, v := range res.states {
			merged[k] = v
		}
	}

	r.mu.Lock()
	for k, v := range merged {
		r.states[k] = v
	}
	r.mu.Unlock()
}

// evaluateRules 对所有任务执行完整的四阶段评估循环。
func (r *Runner) evaluateRules(ctx context.Context) {
	if r.scheduler == nil {
		return
	}
	r.scheduler.ExecuteAllTasks(ctx)
}
