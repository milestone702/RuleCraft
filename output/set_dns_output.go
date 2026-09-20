package output

import (
	"fmt"

	"rulecraft/plugin"
)

// SetDNSOutput 设置网卡 DNS。
type SetDNSOutput struct {
	plat PlatformSubset
}

func NewSetDNSOutput(plat PlatformSubset) *SetDNSOutput {
	return &SetDNSOutput{plat: plat}
}

func (s *SetDNSOutput) ID() string        { return "set_dns" }
func (s *SetDNSOutput) Name() string      { return "设置 DNS" }
func (s *SetDNSOutput) IsAvailable() bool { return true }

// Execute 设置或重置 DNS。
// params:
//   - adapter (string, 必填): 适配器名称
//   - server (string, 可选): DNS 服务器，空表示自动获取
func (s *SetDNSOutput) Execute(params map[string]interface{}) error {
	adapter, _ := params["adapter"].(string)
	if adapter == "" {
		return fmt.Errorf("set_dns: adapter is required")
	}
	server, _ := params["server"].(string)
	return s.plat.SetDNS(adapter, server)
}

func (s *SetDNSOutput) Reset(params map[string]interface{}) error {
	adapter, _ := params["adapter"].(string)
	if adapter == "" {
		return plugin.ErrResetNotSupported
	}
	return s.plat.SetDNS(adapter, "")
}
