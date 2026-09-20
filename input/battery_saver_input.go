package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// BatterySaverInput 采集节电模式状态。
type BatterySaverInput struct {
	plat platform.Platform
}

func NewBatterySaverInput(plat platform.Platform) *BatterySaverInput {
	return &BatterySaverInput{plat: plat}
}

func (b *BatterySaverInput) ID() string        { return "battery_saver" }
func (b *BatterySaverInput) Name() string      { return "节电模式传感器" }
func (b *BatterySaverInput) IsAvailable() bool { return true }

// Collect 写入 battery_saver.enabled。
func (b *BatterySaverInput) Collect(ctx *plugin.SystemContext) error {
	on, err := b.plat.IsBatterySaver()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("battery_saver.enabled", false)
			return nil
		}
		return err
	}
	ctx.SetState("battery_saver.enabled", on)
	return nil
}
