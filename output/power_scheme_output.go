package output

import (
	"fmt"
)

// PowerSchemeOutput 切换 Windows 电源方案。
type PowerSchemeOutput struct {
	plat PlatformSubset
}

// NewPowerSchemeOutput 创建电源方案切换输出插件。
func NewPowerSchemeOutput(plat PlatformSubset) *PowerSchemeOutput {
	return &PowerSchemeOutput{plat: plat}
}

func (p *PowerSchemeOutput) ID() string        { return "power_scheme" }
func (p *PowerSchemeOutput) Name() string      { return "切换电源方案" }
func (p *PowerSchemeOutput) IsAvailable() bool { return true }

// Execute 切换电源方案。
// params: scheme (string, 必填: power_saver / balanced / high_performance)
func (p *PowerSchemeOutput) Execute(params map[string]interface{}) error {
	scheme, _ := params["scheme"].(string)
	if scheme == "" {
		return fmt.Errorf("power_scheme: scheme is required")
	}
	return p.plat.SetPowerScheme(scheme)
}

// Reset 恢复为 balanced 方案。
func (p *PowerSchemeOutput) Reset(params map[string]interface{}) error {
	return p.plat.SetPowerScheme("balanced")
}
