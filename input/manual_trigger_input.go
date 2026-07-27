package input

import (
	"time"

	"rulecraft/plugin"
)

// ManualTriggerInput 手动触发插件 — 通过 API 或 UI 按钮触发。
// 触发后在状态中设置 manual_trigger.triggered 和 manual_trigger.at 等。
type ManualTriggerInput struct {
	triggered bool
	key       string
	value     interface{}
}

func NewManualTriggerInput() *ManualTriggerInput {
	return &ManualTriggerInput{}
}

func (m *ManualTriggerInput) ID() string        { return "manual_trigger" }
func (m *ManualTriggerInput) Name() string      { return "🖱 手动触发" }
func (m *ManualTriggerInput) IsAvailable() bool { return true }

// SetTrigger 供 API 调用，设置触发值。key 为 task_id 等。
func (m *ManualTriggerInput) SetTrigger(key string, value interface{}) {
	m.key = key
	m.value = value
	m.triggered = true
}

func (m *ManualTriggerInput) Collect(ctx *plugin.SystemContext) error {
	if m.triggered {
		now := time.Now()
		ctx.SetState("manual_trigger.triggered", true)
		ctx.SetState("manual_trigger.at", now.Format("15:04:05"))
		ctx.SetState("manual_trigger.timestamp", now.Unix())
		ctx.SetState("manual_trigger.task_id", m.key)
		if m.value != nil {
			ctx.SetState("manual_trigger.value", m.value)
		}
		m.triggered = false
	} else {
		// 未触发时清除状态（方便条件判断不成立）
		ctx.SetState("manual_trigger.triggered", false)
	}
	return nil
}
