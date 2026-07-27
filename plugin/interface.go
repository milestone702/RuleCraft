// Package plugin 定义插件接口系统和插件注册表。
// 这是整个引擎的扩展点——所有输入插件（Sensors）和输出插件（Actuators）
// 都实现这些接口，并通过 Registry 注册到引擎中。
package plugin

import (
	"context"
	"sync"

	"rulecraft/config"
)

// ============================================================================
// 系统上下文
// ============================================================================

// SystemContext 是插件执行时的运行时上下文。
// 它包含全局状态、取消信号、以及通知管理器引用。
type SystemContext struct {
	Mu     sync.RWMutex
	States map[string]interface{} // 全局状态存储
	Ctx    context.Context        // 支持超时取消
	Notify NotificationManager    // 通知管理器（可选注入）
}

// GetState 线程安全地读取全局状态。
func (sc *SystemContext) GetState(key string) (interface{}, bool) {
	sc.Mu.RLock()
	defer sc.Mu.RUnlock()
	v, ok := sc.States[key]
	return v, ok
}

// SetState 线程安全地写入全局状态。
func (sc *SystemContext) SetState(key string, value interface{}) {
	sc.Mu.Lock()
	defer sc.Mu.Unlock()
	sc.States[key] = value
}

// GetStateString 读取状态并返回字符串表示。
func (sc *SystemContext) GetStateString(key string) (string, bool) {
	v, ok := sc.GetState(key)
	if !ok {
		return "", false
	}
	if s, ok := v.(string); ok {
		return s, ok
	}
	return "", false
}

// ============================================================================
// 插件接口
// ============================================================================

// InputPlugin 输入插件（Sensor）接口。
// 每个输入插件负责采集一种系统状态数据，并将结果写入 SystemContext。
type InputPlugin interface {
	// ID 返回插件的唯一标识符，如 "wifi"、"power"。
	ID() string

	// Name 返回插件的人类可读名称。
	Name() string

	// Collect 采集数据并将结果写入 ctx.States。
	Collect(ctx *SystemContext) error

	// IsAvailable 返回插件在当前系统上是否可用（如权限、硬件存在）。
	IsAvailable() bool
}

// OutputPlugin 输出插件（Actuator）接口。
// 每个输出插件负责执行一种系统控制动作。
type OutputPlugin interface {
	// ID 返回插件的唯一标识符，如 "notify"、"power_scheme"。
	ID() string

	// Name 返回插件的人类可读名称。
	Name() string

	// Execute 执行动作。当条件从 false→true 时由引擎调用。
	Execute(params map[string]interface{}) error

	// Reset 撤销动作。当条件从 true→false 时由引擎调用。
	// 不支持 Reset 的插件应返回 ErrResetNotSupported。
	Reset(params map[string]interface{}) error

	// IsAvailable 返回插件在当前系统上是否可用。
	IsAvailable() bool
}

// ErrResetNotSupported 当输出插件不支持 Reset 操作时返回。
var ErrResetNotSupported = &ResetNotSupportedError{}

// ResetNotSupportedError 表示输出插件不支持 Reset 操作。
type ResetNotSupportedError struct{}

func (e *ResetNotSupportedError) Error() string {
	return "reset not supported by this plugin"
}

// ============================================================================
// 插件注册表
// ============================================================================

// PluginMeta 插件元信息。
type PluginMeta struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        PluginType       `json:"type"`       // input / output
	BuiltIn     bool             `json:"built_in"`   // 是否为内置插件
	Description string           `json:"description,omitempty"`
}

// PluginType 插件类型。
type PluginType string

const (
	PluginTypeInput  PluginType = "input"
	PluginTypeOutput PluginType = "output"
)

// Registry 是插件注册表，管理所有输入插件和输出插件的生命周期。
type Registry struct {
	mu       sync.RWMutex
	inputs   map[string]InputPlugin
	outputs  map[string]OutputPlugin
	metas    map[string]PluginMeta
}

// NewRegistry 创建新的空注册表。
func NewRegistry() *Registry {
	return &Registry{
		inputs:  make(map[string]InputPlugin),
		outputs: make(map[string]OutputPlugin),
		metas:   make(map[string]PluginMeta),
	}
}

// RegisterInput 注册一个输入插件。
func (r *Registry) RegisterInput(p InputPlugin, description string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.ID()
	if _, exists := r.inputs[id]; exists {
		return &ErrPluginAlreadyRegistered{ID: id}
	}

	r.inputs[id] = p
	r.metas[id] = PluginMeta{
		ID:          id,
		Name:        p.Name(),
		Type:        PluginTypeInput,
		BuiltIn:     true,
		Description: description,
	}
	return nil
}

// RegisterOutput 注册一个输出插件。
func (r *Registry) RegisterOutput(p OutputPlugin, description string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.ID()
	if _, exists := r.outputs[id]; exists {
		return &ErrPluginAlreadyRegistered{ID: id}
	}

	r.outputs[id] = p
	r.metas[id] = PluginMeta{
		ID:          id,
		Name:        p.Name(),
		Type:        PluginTypeOutput,
		BuiltIn:     true,
		Description: description,
	}
	return nil
}

// GetInput 根据 ID 获取输入插件。
func (r *Registry) GetInput(id string) (InputPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.inputs[id]
	return p, ok
}

// GetOutput 根据 ID 获取输出插件。
func (r *Registry) GetOutput(id string) (OutputPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.outputs[id]
	return p, ok
}

// ListInputs 返回所有已注册的输入插件 ID 列表。
func (r *Registry) ListInputs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.inputs))
	for id := range r.inputs {
		ids = append(ids, id)
	}
	return ids
}

// ListOutputs 返回所有已注册的输出插件 ID 列表。
func (r *Registry) ListOutputs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.outputs))
	for id := range r.outputs {
		ids = append(ids, id)
	}
	return ids
}

// ListAll 返回所有已注册插件的元信息。
func (r *Registry) ListAll() []PluginMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]PluginMeta, 0, len(r.metas))
	for _, m := range r.metas {
		list = append(list, m)
	}
	return list
}

// GetAllInputs 返回所有注册的输入插件映射。
func (r *Registry) GetAllInputs() map[string]InputPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]InputPlugin, len(r.inputs))
	for k, v := range r.inputs {
		result[k] = v
	}
	return result
}

// GetAllOutputs 返回所有注册的输出插件映射。
func (r *Registry) GetAllOutputs() map[string]OutputPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]OutputPlugin, len(r.outputs))
	for k, v := range r.outputs {
		result[k] = v
	}
	return result
}

// ============================================================================
// 通知管理器接口（供插件使用）
// ============================================================================

// NotificationManager 是插件可选的依赖，用于发送桌面通知。
// 引擎在注入 SystemContext 时，如果通知系统已初始化，会设置此字段。
type NotificationManager interface {
	Send(event config.NotificationEvent) error
	SendFromSource(sourceID, sourceType, eventType string,
		level config.NotificationLevel, title, message string,
		metadata map[string]interface{}) error
}

// ============================================================================
// 错误类型
// ============================================================================

// ErrPluginAlreadyRegistered 插件已注册错误。
type ErrPluginAlreadyRegistered struct {
	ID string
}

func (e *ErrPluginAlreadyRegistered) Error() string {
	return "plugin already registered: " + e.ID
}
