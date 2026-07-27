package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// UptimeInput 系统运行时间传感器。
type UptimeInput struct {
	plat platform.Platform
}

func NewUptimeInput(plat platform.Platform) *UptimeInput {
	return &UptimeInput{plat: plat}
}

func (u *UptimeInput) ID() string        { return "uptime" }
func (u *UptimeInput) Name() string      { return "系统运行时间传感器" }
func (u *UptimeInput) IsAvailable() bool { return true }

func (u *UptimeInput) Collect(ctx *plugin.SystemContext) error {
	seconds, err := u.plat.GetUptimeSeconds()
	if err != nil {
		return nil
	}
	ctx.SetState("uptime.seconds", seconds)
	ctx.SetState("uptime.minutes", seconds/60)
	ctx.SetState("uptime.hours", seconds/3600)
	return nil
}
