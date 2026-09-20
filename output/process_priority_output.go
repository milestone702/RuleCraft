package output

import (
	"fmt"

	"rulecraft/plugin"
)

// ProcessPriorityOutput 设置进程优先级。
type ProcessPriorityOutput struct {
	plat PlatformSubset
}

func NewProcessPriorityOutput(plat PlatformSubset) *ProcessPriorityOutput {
	return &ProcessPriorityOutput{plat: plat}
}

func (p *ProcessPriorityOutput) ID() string        { return "process_priority" }
func (p *ProcessPriorityOutput) Name() string      { return "进程优先级" }
func (p *ProcessPriorityOutput) IsAvailable() bool { return true }

// Execute params: name (必填), priority (idle/below_normal/normal/above_normal/high)
func (p *ProcessPriorityOutput) Execute(params map[string]interface{}) error {
	name := paramString(params, "name")
	if name == "" {
		return fmt.Errorf("process_priority: name is required")
	}
	priority := paramString(params, "priority")
	if priority == "" {
		priority = "normal"
	}
	return p.plat.SetProcessPriority(name, priority)
}

func (p *ProcessPriorityOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
