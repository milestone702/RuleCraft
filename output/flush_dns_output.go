package output

import (
	"rulecraft/plugin"
)

// FlushDNSOutput 清空 DNS 缓存。
type FlushDNSOutput struct {
	plat PlatformSubset
}

func NewFlushDNSOutput(plat PlatformSubset) *FlushDNSOutput {
	return &FlushDNSOutput{plat: plat}
}

func (f *FlushDNSOutput) ID() string        { return "flush_dns" }
func (f *FlushDNSOutput) Name() string      { return "刷新 DNS 缓存" }
func (f *FlushDNSOutput) IsAvailable() bool { return true }

func (f *FlushDNSOutput) Execute(params map[string]interface{}) error {
	return f.plat.FlushDNS()
}

func (f *FlushDNSOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
