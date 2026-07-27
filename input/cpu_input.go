package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// CPUInput CPU 信息传感器。
type CPUInput struct {
	plat platform.Platform
}

func NewCPUInput(plat platform.Platform) *CPUInput {
	return &CPUInput{plat: plat}
}

func (c *CPUInput) ID() string        { return "cpu" }
func (c *CPUInput) Name() string      { return "CPU 传感器" }
func (c *CPUInput) IsAvailable() bool { return true }

func (c *CPUInput) Collect(ctx *plugin.SystemContext) error {
	info, err := c.plat.GetCPUInfo()
	if err != nil {
		return nil
	}
	ctx.SetState("cpu.logical_cores", info.LogicalCores)
	ctx.SetState("cpu.physical_cores", info.PhysicalCores)
	ctx.SetState("cpu.architecture", info.Architecture)
	return nil
}
