package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// PowerPlanInput 采集当前电源计划。
type PowerPlanInput struct {
	plat platform.Platform
}

// NewPowerPlanInput 创建电源计划输入插件。
func NewPowerPlanInput(plat platform.Platform) *PowerPlanInput {
	return &PowerPlanInput{plat: plat}
}

func (p *PowerPlanInput) ID() string        { return "power_plan" }
func (p *PowerPlanInput) Name() string      { return "电源计划传感器" }
func (p *PowerPlanInput) IsAvailable() bool { return true }

// Collect 采集电源计划。
// 状态键：
//
//	power_plan.active   — balanced / power_saver / high_performance / GUID
//	power_plan.is_balanced / is_power_saver / is_high_performance
func (p *PowerPlanInput) Collect(ctx *plugin.SystemContext) error {
	plan, err := p.plat.GetActivePowerPlan()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("power_plan.active", "")
			ctx.SetState("power_plan.is_balanced", false)
			ctx.SetState("power_plan.is_power_saver", false)
			ctx.SetState("power_plan.is_high_performance", false)
			return nil
		}
		return err
	}

	ctx.SetState("power_plan.active", plan)
	ctx.SetState("power_plan.is_balanced", plan == "balanced")
	ctx.SetState("power_plan.is_power_saver", plan == "power_saver")
	ctx.SetState("power_plan.is_high_performance", plan == "high_performance")
	return nil
}
