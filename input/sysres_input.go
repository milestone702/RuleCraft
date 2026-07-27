package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// SysResInput 采集系统资源（CPU、内存）使用率。
type SysResInput struct {
	plat platform.Platform
}

// NewSysResInput 创建系统资源输入插件。
func NewSysResInput(plat platform.Platform) *SysResInput {
	return &SysResInput{plat: plat}
}

func (s *SysResInput) ID() string        { return "sysres" }
func (s *SysResInput) Name() string      { return "系统资源传感器" }
func (s *SysResInput) IsAvailable() bool { return true }

// Collect 采集系统资源信息并写入全局状态。
func (s *SysResInput) Collect(ctx *plugin.SystemContext) error {
	info, err := s.plat.GetSystemResources()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("sysres.cpu_percent", 0.0)
			ctx.SetState("sysres.memory_percent", 0.0)
			return nil
		}
		return err
	}

	ctx.SetState("sysres.cpu_percent", info.CPUPercent)
	ctx.SetState("sysres.memory_percent", info.MemoryPercent)
	ctx.SetState("sysres.memory_used_gb", info.MemoryUsedGB)
	ctx.SetState("sysres.memory_total_gb", info.MemoryTotalGB)
	return nil
}
