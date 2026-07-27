package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// DisplayInput 显示器传感器。
type DisplayInput struct {
	plat platform.Platform
}

func NewDisplayInput(plat platform.Platform) *DisplayInput {
	return &DisplayInput{plat: plat}
}

func (d *DisplayInput) ID() string        { return "display" }
func (d *DisplayInput) Name() string      { return "显示器传感器" }
func (d *DisplayInput) IsAvailable() bool { return true }

func (d *DisplayInput) Collect(ctx *plugin.SystemContext) error {
	info, err := d.plat.GetDisplayInfo()
	if err != nil {
		return nil
	}
	ctx.SetState("display.primary_width", info.PrimaryWidth)
	ctx.SetState("display.primary_height", info.PrimaryHeight)
	ctx.SetState("display.monitor_count", info.MonitorCount)
	ctx.SetState("display.dpi", info.DPI)
	return nil
}
