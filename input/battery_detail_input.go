package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// BatteryDetailInput 电池详细信息传感器。
type BatteryDetailInput struct {
	plat platform.Platform
}

func NewBatteryDetailInput(plat platform.Platform) *BatteryDetailInput {
	return &BatteryDetailInput{plat: plat}
}

func (b *BatteryDetailInput) ID() string        { return "battery_detail" }
func (b *BatteryDetailInput) Name() string      { return "电池详细传感器" }
func (b *BatteryDetailInput) IsAvailable() bool { return true }

func (b *BatteryDetailInput) Collect(ctx *plugin.SystemContext) error {
	info, err := b.plat.GetExtendedBatteryInfo()
	if err != nil {
		return nil
	}
	ctx.SetState("battery_detail.voltage", info.VoltageNow)
	ctx.SetState("battery_detail.charge_rate", info.ChargeRate)
	ctx.SetState("battery_detail.design_capacity", info.DesignCapacity)
	ctx.SetState("battery_detail.full_charge_capacity", info.FullChargeCapacity)
	ctx.SetState("battery_detail.cycle_count", info.CycleCount)
	if info.EstimatedTimeRemaining >= 0 {
		ctx.SetState("battery_detail.seconds_remaining", info.EstimatedTimeRemaining)
	}
	return nil
}
