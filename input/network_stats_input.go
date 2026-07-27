package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// NetworkStatsInput 网络流量传感器。
type NetworkStatsInput struct {
	plat platform.Platform
}

func NewNetworkStatsInput(plat platform.Platform) *NetworkStatsInput {
	return &NetworkStatsInput{plat: plat}
}

func (n *NetworkStatsInput) ID() string        { return "network_stats" }
func (n *NetworkStatsInput) Name() string      { return "网络流量传感器" }
func (n *NetworkStatsInput) IsAvailable() bool { return true }

func (n *NetworkStatsInput) Collect(ctx *plugin.SystemContext) error {
	traffic, err := n.plat.GetNetworkTraffic()
	if err != nil {
		return nil
	}
	ctx.SetState("network_stats.bytes_sent", traffic.BytesSent)
	ctx.SetState("network_stats.bytes_received", traffic.BytesReceived)
	return nil
}
