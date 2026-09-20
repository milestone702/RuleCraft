package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// GPUInput 采集显卡名称。
type GPUInput struct {
	plat platform.Platform
}

func NewGPUInput(plat platform.Platform) *GPUInput {
	return &GPUInput{plat: plat}
}

func (g *GPUInput) ID() string        { return "gpu" }
func (g *GPUInput) Name() string      { return "显卡传感器" }
func (g *GPUInput) IsAvailable() bool { return true }

// Collect 写入 gpu.name。
func (g *GPUInput) Collect(ctx *plugin.SystemContext) error {
	name, err := g.plat.GetGPUInfo()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("gpu.name", "")
			return nil
		}
		return err
	}
	ctx.SetState("gpu.name", name)
	return nil
}
