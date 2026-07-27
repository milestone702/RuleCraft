package input

import (
	"encoding/json"

	"rulecraft/platform"
	"rulecraft/plugin"
)

// ProcessInput 采集进程列表信息。
type ProcessInput struct {
	plat platform.Platform
}

// NewProcessInput 创建进程输入插件。
func NewProcessInput(plat platform.Platform) *ProcessInput {
	return &ProcessInput{plat: plat}
}

func (p *ProcessInput) ID() string        { return "process" }
func (p *ProcessInput) Name() string      { return "进程传感器" }
func (p *ProcessInput) IsAvailable() bool { return true }

// Collect 采集进程列表并写入全局状态。
// 写入 process.count（数量）、process.names（名称数组）、process.list（完整信息数组）。
func (p *ProcessInput) Collect(ctx *plugin.SystemContext) error {
	processes, err := p.plat.ListProcesses()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("process.count", 0)
			return nil
		}
		return err
	}

	ctx.SetState("process.count", len(processes))

	names := make([]string, 0, len(processes))
	for _, proc := range processes {
		names = append(names, proc.Name)
	}
	ctx.SetState("process.names", names)

	// 写入完整列表（JSON 字符串）
	if data, err := json.Marshal(processes); err == nil {
		ctx.SetState("process.list", string(data))
	}

	return nil
}
