package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// PowerInput 采集电源状态（是否充电、电池百分比）。
type PowerInput struct {
	plat platform.Platform
}

// NewPowerInput 创建电源输入插件。
func NewPowerInput(plat platform.Platform) *PowerInput {
	return &PowerInput{plat: plat}
}

func (p *PowerInput) ID() string        { return "power" }
func (p *PowerInput) Name() string      { return "电源传感器" }
func (p *PowerInput) IsAvailable() bool { return true }

// Collect 采集电源信息并写入全局状态。
func (p *PowerInput) Collect(ctx *plugin.SystemContext) error {
	info, err := p.plat.CollectPowerInfo()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("power.is_charging", false)
			ctx.SetState("power.battery_percent", -1)
			ctx.SetState("power.ac_line_status", 255)
			return nil
		}
		return err
	}

	ctx.SetState("power.is_charging", info.IsCharging)
	ctx.SetState("power.battery_percent", info.BatteryPercent)
	ctx.SetState("power.ac_line_status", info.ACLineStatus)
	return nil
}
