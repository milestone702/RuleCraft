package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// OSInput 操作系统信息传感器。
type OSInput struct {
	plat platform.Platform
}

func NewOSInput(plat platform.Platform) *OSInput {
	return &OSInput{plat: plat}
}

func (o *OSInput) ID() string        { return "os" }
func (o *OSInput) Name() string      { return "系统信息传感器" }
func (o *OSInput) IsAvailable() bool { return true }

func (o *OSInput) Collect(ctx *plugin.SystemContext) error {
	info, err := o.plat.GetOSInfo()
	if err != nil {
		return nil
	}
	ctx.SetState("os.computer_name", info.ComputerName)
	ctx.SetState("os.user_name", info.UserName)
	return nil
}
