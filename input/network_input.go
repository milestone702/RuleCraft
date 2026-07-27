package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// NetworkInput 采集网络适配器信息。
type NetworkInput struct {
	plat platform.Platform
}

// NewNetworkInput 创建网络输入插件。
func NewNetworkInput(plat platform.Platform) *NetworkInput {
	return &NetworkInput{plat: plat}
}

func (n *NetworkInput) ID() string        { return "network" }
func (n *NetworkInput) Name() string      { return "网络传感器" }
func (n *NetworkInput) IsAvailable() bool { return true }

// Collect 采集网络信息并写入全局状态。
func (n *NetworkInput) Collect(ctx *plugin.SystemContext) error {
	adapters, err := n.plat.CollectNetworkInfo()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("network.ipv4", "")
			ctx.SetState("network.gateway", "")
			ctx.SetState("network.is_connected", false)
			ctx.SetState("network.connection_type", "")
			return nil
		}
		return err
	}

	if len(adapters) == 0 {
		ctx.SetState("network.is_connected", false)
		return nil
	}

	// 取第一个适配器信息
	adapter := adapters[0]
	ctx.SetState("network.ipv4", adapter.IPv4)
	ctx.SetState("network.gateway", adapter.Gateway)
	ctx.SetState("network.subnet_mask", adapter.SubnetMask)
	ctx.SetState("network.is_connected", adapter.IsConnected)
	ctx.SetState("network.connection_type", adapter.ConnectionType)
	ctx.SetState("network.adapter_name", adapter.Name)
	// 全部适配器列表（JSON 字符串，简化）
	if len(adapters) > 1 {
		names := make([]string, 0, len(adapters))
		for _, a := range adapters {
			names = append(names, a.Name)
		}
		ctx.SetState("network.all_adapters", names)
	}
	return nil
}
