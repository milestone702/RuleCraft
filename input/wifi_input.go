package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// WiFiInput 采集 Wi-Fi 状态。
type WiFiInput struct {
	plat platform.Platform
}

// NewWiFiInput 创建 Wi-Fi 输入插件。
func NewWiFiInput(plat platform.Platform) *WiFiInput {
	return &WiFiInput{plat: plat}
}

func (w *WiFiInput) ID() string        { return "wifi" }
func (w *WiFiInput) Name() string      { return "Wi-Fi 传感器" }
func (w *WiFiInput) IsAvailable() bool { return true }

// Collect 采集 Wi-Fi 信息并写入全局状态。
func (w *WiFiInput) Collect(ctx *plugin.SystemContext) error {
	info, err := w.plat.CollectWiFiInfo()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("wifi.ssid", "")
			ctx.SetState("wifi.bssid", "")
			ctx.SetState("wifi.is_connected", false)
			ctx.SetState("wifi.signal_quality", 0)
			return nil
		}
		return err
	}

	ctx.SetState("wifi.ssid", info.SSID)
	ctx.SetState("wifi.bssid", info.BSSID)
	ctx.SetState("wifi.is_connected", info.IsConnected)
	ctx.SetState("wifi.signal_quality", info.SignalQuality)
	return nil
}
