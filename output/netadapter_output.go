package output

import (
	"fmt"
)

// NetAdapterOutput 启用或禁用网络适配器（需要管理员权限）。
type NetAdapterOutput struct {
	plat PlatformSubset
}

// NewNetAdapterOutput 创建网络适配器控制输出插件。
func NewNetAdapterOutput(plat PlatformSubset) *NetAdapterOutput {
	return &NetAdapterOutput{plat: plat}
}

func (n *NetAdapterOutput) ID() string        { return "netadapter" }
func (n *NetAdapterOutput) Name() string      { return "网络适配器控制" }
func (n *NetAdapterOutput) IsAvailable() bool { return true }

// Execute 禁用或启用网络适配器。
// params: name (string, 必填: 适配器名称), enabled (bool, 必填: true=启用, false=禁用)
func (n *NetAdapterOutput) Execute(params map[string]interface{}) error {
	name, _ := params["name"].(string)
	if name == "" {
		return fmt.Errorf("netadapter: name is required")
	}

	enabled := true
	if v, ok := params["enabled"]; ok {
		enabled, _ = v.(bool)
	}

	return n.plat.SetNetworkAdapter(name, enabled)
}

// Reset 恢复启用。
func (n *NetAdapterOutput) Reset(params map[string]interface{}) error {
	name, _ := params["name"].(string)
	if name == "" {
		return nil
	}
	return n.plat.SetNetworkAdapter(name, true)
}
