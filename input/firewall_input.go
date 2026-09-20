package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// FirewallInput 采集防火墙配置文件状态。
type FirewallInput struct {
	plat platform.Platform
}

func NewFirewallInput(plat platform.Platform) *FirewallInput {
	return &FirewallInput{plat: plat}
}

func (f *FirewallInput) ID() string        { return "firewall" }
func (f *FirewallInput) Name() string      { return "防火墙传感器" }
func (f *FirewallInput) IsAvailable() bool { return true }

// Collect 写入 firewall.domain/private/public/all_enabled。
func (f *FirewallInput) Collect(ctx *plugin.SystemContext) error {
	d, pr, pub, err := f.plat.GetFirewallStatus()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("firewall.domain", false)
			ctx.SetState("firewall.private", false)
			ctx.SetState("firewall.public", false)
			ctx.SetState("firewall.all_enabled", false)
			return nil
		}
		return err
	}
	ctx.SetState("firewall.domain", d)
	ctx.SetState("firewall.private", pr)
	ctx.SetState("firewall.public", pub)
	ctx.SetState("firewall.all_enabled", d && pr && pub)
	return nil
}
