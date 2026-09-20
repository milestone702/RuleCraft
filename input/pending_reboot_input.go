package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// PendingRebootInput 检测系统是否等待重启。
type PendingRebootInput struct {
	plat platform.Platform
}

func NewPendingRebootInput(plat platform.Platform) *PendingRebootInput {
	return &PendingRebootInput{plat: plat}
}

func (p *PendingRebootInput) ID() string        { return "pending_reboot" }
func (p *PendingRebootInput) Name() string      { return "待重启传感器" }
func (p *PendingRebootInput) IsAvailable() bool { return true }

// Collect 写入 pending_reboot.pending / reason。
func (p *PendingRebootInput) Collect(ctx *plugin.SystemContext) error {
	pending, reason, err := p.plat.GetPendingReboot()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("pending_reboot.pending", false)
			ctx.SetState("pending_reboot.reason", "")
			return nil
		}
		return err
	}
	ctx.SetState("pending_reboot.pending", pending)
	ctx.SetState("pending_reboot.reason", reason)
	return nil
}
