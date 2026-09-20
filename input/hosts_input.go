package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// HostsMonitorInput 监控 hosts 文件变更。
type HostsMonitorInput struct {
	plat     platform.Platform
	lastMTime int64
}

func NewHostsMonitorInput(plat platform.Platform) *HostsMonitorInput {
	return &HostsMonitorInput{plat: plat}
}

func (h *HostsMonitorInput) ID() string        { return "hosts" }
func (h *HostsMonitorInput) Name() string      { return "Hosts 文件传感器" }
func (h *HostsMonitorInput) IsAvailable() bool { return true }

// Collect 写入 hosts.mtime / lines / changed。
func (h *HostsMonitorInput) Collect(ctx *plugin.SystemContext) error {
	mtime, lines, err := h.plat.GetHostsInfo()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("hosts.mtime", int64(0))
			ctx.SetState("hosts.lines", 0)
			ctx.SetState("hosts.changed", false)
			return nil
		}
		return err
	}
	changed := h.lastMTime != 0 && mtime != h.lastMTime
	h.lastMTime = mtime
	ctx.SetState("hosts.mtime", mtime)
	ctx.SetState("hosts.lines", lines)
	ctx.SetState("hosts.changed", changed)
	return nil
}
