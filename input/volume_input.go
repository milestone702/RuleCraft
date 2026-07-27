package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// VolumeInput 音量传感器。
type VolumeInput struct {
	plat platform.Platform
}

func NewVolumeInput(plat platform.Platform) *VolumeInput {
	return &VolumeInput{plat: plat}
}

func (v *VolumeInput) ID() string        { return "volume" }
func (v *VolumeInput) Name() string      { return "音量传感器" }
func (v *VolumeInput) IsAvailable() bool { return true }

func (v *VolumeInput) Collect(ctx *plugin.SystemContext) error {
	level, muted, err := v.plat.GetSystemVolumeLevel()
	if err != nil {
		return nil
	}
	ctx.SetState("volume.level", level)
	ctx.SetState("volume.muted", muted)
	return nil
}
