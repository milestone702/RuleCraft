package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// ProxyInput 采集系统代理设置。
type ProxyInput struct {
	plat platform.Platform
}

func NewProxyInput(plat platform.Platform) *ProxyInput {
	return &ProxyInput{plat: plat}
}

func (p *ProxyInput) ID() string        { return "proxy" }
func (p *ProxyInput) Name() string      { return "系统代理传感器" }
func (p *ProxyInput) IsAvailable() bool { return true }

// Collect 写入 proxy.enabled / proxy.server。
func (p *ProxyInput) Collect(ctx *plugin.SystemContext) error {
	enabled, server, err := p.plat.GetProxyInfo()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("proxy.enabled", false)
			ctx.SetState("proxy.server", "")
			return nil
		}
		return err
	}
	ctx.SetState("proxy.enabled", enabled)
	ctx.SetState("proxy.server", server)
	return nil
}
