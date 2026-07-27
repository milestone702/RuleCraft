package output

import (
	"fmt"

	"rulecraft/plugin"
)

// PowerActionOutput 执行电源操作（关机/重启/睡眠/休眠）。
type PowerActionOutput struct {
	plat PlatformSubset
}

// NewPowerActionOutput 创建电源操作输出插件。
func NewPowerActionOutput(plat PlatformSubset) *PowerActionOutput {
	return &PowerActionOutput{plat: plat}
}

func (p *PowerActionOutput) ID() string        { return "power_action" }
func (p *PowerActionOutput) Name() string      { return "电源操作" }
func (p *PowerActionOutput) IsAvailable() bool { return true }

// Execute 执行电源操作。
// params: action (string, 必填: shutdown / restart / sleep / hibernate)
func (p *PowerActionOutput) Execute(params map[string]interface{}) error {
	action, _ := params["action"].(string)
	if action == "" {
		return fmt.Errorf("power_action: action (shutdown/restart/sleep/hibernate) is required")
	}
	return p.plat.PowerAction(action)
}

// Reset 不支持 Reset。
func (p *PowerActionOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
