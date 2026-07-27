package engine

import (
	"fmt"
	"log"
	"sync"
	"time"
	"strings"

	"rulecraft/config"
)

// ============================================================================
// 通知管理器
// ============================================================================

// Notifier 实现 config.Notifier 接口，管理桌面通知的发送、防抖、限速和降级。
type Notifier struct {
	mu             sync.RWMutex
	globalCfg      config.GlobalNotificationConfig
	moduleCfgs     map[string]*config.NotificationConfig // key: "sourceType:sourceID"
	debounceMap    map[string]time.Time                  // key: "sourceID:eventType" → last sent time
	slidingCounter []time.Time                           // 滑动窗口内的时间戳
	enabled        bool
}

// NewNotifier 创建通知管理器。
func NewNotifier(globalCfg config.GlobalNotificationConfig) *Notifier {
	return &Notifier{
		globalCfg:   globalCfg,
		moduleCfgs:  make(map[string]*config.NotificationConfig),
		debounceMap: make(map[string]time.Time),
	}
}

// Send 发送一条通知事件。
func (n *Notifier) Send(event config.NotificationEvent) error {
	return n.send(event.SourceID, event.SourceType, event.Type, event.Level, event.Title, event.Message, event.Metadata)
}

// SendFromSource 便捷方法：从来源模块发送通知。
func (n *Notifier) SendFromSource(sourceID, sourceType, eventType string,
	level config.NotificationLevel, title, message string,
	metadata map[string]interface{}) error {

	return n.send(sourceID, sourceType, eventType, level, title, message, metadata)
}

// send 核心发送逻辑。
func (n *Notifier) send(sourceID, sourceType, eventType string,
	level config.NotificationLevel, title, message string,
	metadata map[string]interface{}) error {

	n.mu.RLock()
	globalEnabled := n.globalCfg.Enabled
	n.mu.RUnlock()

	// 1. 全局开关检查
	if !globalEnabled {
		return nil
	}

	// 2. 模块级开关检查
	moduleKey := sourceType + ":" + sourceID
	n.mu.RLock()
	modCfg, hasModCfg := n.moduleCfgs[moduleKey]
	globalLevel := n.globalCfg.DefaultLevel
	n.mu.RUnlock()

	if hasModCfg && !modCfg.Enabled {
		return nil // 模块级禁用
	}

	// 3. 级别过滤
	if !n.passesLevelFilter(level, hasModCfg, modCfg, globalLevel) {
		return nil
	}

	// 4. 事件白名单过滤
	if hasModCfg && len(modCfg.Events) > 0 {
		if !containsString(modCfg.Events, eventType) {
			return nil
		}
	}

	// 5. 防抖检查
	debounceKey := sourceID + ":" + eventType
	if !n.checkDebounce(debounceKey) {
		return nil
	}

	// 6. 限速检查
	if !n.checkRateLimit() {
		// 限速触发：降级为日志
		log.Printf("[notifier] rate limited, dropping notification: %s [%s]", title, eventType)
		return nil
	}

	// 7. 渲染消息（模板变量替换）
	finalMessage := n.renderMessage(eventType, message, metadata, sourceID, sourceType, hasModCfg, modCfg)

	// 8. 发送桌面通知（通过系统接口）
	// 实际实现由 TrayManager 完成，此处为占位
	log.Printf("[notifier] %s: %s — %s", level, title, finalMessage)
	return nil
}

// RegisterConfig 注册模块级通知配置。
func (n *Notifier) RegisterConfig(sourceID, sourceType string, cfg config.NotificationConfig) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.moduleCfgs[sourceType+":"+sourceID] = &cfg
}

// UnregisterConfig 注销模块级通知配置。
func (n *Notifier) UnregisterConfig(sourceID, sourceType string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.moduleCfgs, sourceType+":"+sourceID)
}

// SetGlobalConfig 更新全局通知配置。
func (n *Notifier) SetGlobalConfig(cfg config.GlobalNotificationConfig) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.globalCfg = cfg
}

// Close 清理资源。
func (n *Notifier) Close() error {
	return nil
}

// ============================================================================
// 内部方法
// ============================================================================

// passesLevelFilter 检查通知级别是否满足过滤条件。
func (n *Notifier) passesLevelFilter(level config.NotificationLevel, hasModCfg bool, modCfg *config.NotificationConfig, globalLevel string) bool {
	effectiveLevel := globalLevel
	if hasModCfg && modCfg.Level != "" {
		effectiveLevel = modCfg.Level
	}
	return compareLevel(level, config.NotificationLevel(effectiveLevel)) >= 0
}

// checkDebounce 检查防抖：同一 sourceID + eventType 在冷却期内不重复发送。
func (n *Notifier) checkDebounce(key string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	lastSent, exists := n.debounceMap[key]
	now := time.Now()

	cooldown := time.Duration(n.globalCfg.CooldownSeconds) * time.Second
	if exists && now.Sub(lastSent) < cooldown {
		return false
	}

	n.debounceMap[key] = now
	return true
}

// checkRateLimit 检查限速：滑动窗口计数器。
func (n *Notifier) checkRateLimit() bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)

	// 清理过期的滑动窗口记录
	var valid []time.Time
	for _, t := range n.slidingCounter {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}
	n.slidingCounter = valid

	// 检查是否超过限制
	if len(n.slidingCounter) >= n.globalCfg.MaxPerMinute {
		return false
	}

	n.slidingCounter = append(n.slidingCounter, now)
	return true
}

// renderMessage 渲染通知消息（模板变量替换 + 自定义消息）。
func (n *Notifier) renderMessage(eventType, message string, metadata map[string]interface{},
	sourceID, sourceType string, hasModCfg bool, modCfg *config.NotificationConfig) string {

	// 检查是否有自定义消息模板
	if hasModCfg && modCfg.CustomMessages != nil {
		if customMsg, ok := modCfg.CustomMessages[eventType]; ok {
			message = customMsg
		}
	}

	// 简易模板变量替换
	message = strings.ReplaceAll(message, "{{source_id}}", sourceID)
	message = strings.ReplaceAll(message, "{{source_type}}", sourceType)
	message = strings.ReplaceAll(message, "{{event_type}}", eventType)

	if metadata != nil {
		for k, v := range metadata {
			placeholder := "{{" + k + "}}"
			message = strings.ReplaceAll(message, placeholder, fmt.Sprintf("%v", v))
		}
	}

	return message
}

// ============================================================================
// 辅助函数
// ============================================================================

// compareLevel 比较两个通知级别。
// 返回值：a >= b 时 >= 0（a 的严重程度不低于 b）。
func compareLevel(a, b config.NotificationLevel) int {
	order := map[config.NotificationLevel]int{
		config.LevelDebug:   0,
		config.LevelInfo:    1,
		config.LevelWarning: 2,
		config.LevelError:   3,
	}
	return order[a] - order[b]
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
