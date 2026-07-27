package output

// PreventSleepOutput 阻止或允许系统/显示器休眠。
type PreventSleepOutput struct {
	plat PlatformSubset
}

// NewPreventSleepOutput 创建阻止休眠输出插件。
func NewPreventSleepOutput(plat PlatformSubset) *PreventSleepOutput {
	return &PreventSleepOutput{plat: plat}
}

func (p *PreventSleepOutput) ID() string        { return "prevent_sleep" }
func (p *PreventSleepOutput) Name() string      { return "阻止休眠" }
func (p *PreventSleepOutput) IsAvailable() bool { return true }

// Execute 阻止系统休眠。
func (p *PreventSleepOutput) Execute(params map[string]interface{}) error {
	return p.plat.PreventSleep(true)
}

// Reset 恢复允许休眠。
func (p *PreventSleepOutput) Reset(params map[string]interface{}) error {
	return p.plat.PreventSleep(false)
}
