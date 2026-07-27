// Package input 包含所有内置输入插件（Sensors）。
//
// 每个输入插件采集一种系统状态数据，通过 SystemContext.SetState 写入全局状态。
// 插件通过 platform.Platform 接口调用操作系统功能。
//
// 添加新输入插件：
//   1. 在 input/ 下新建单独 .go 文件
//   2. 实现 plugin.InputPlugin 接口
//   3. 在 RegisterBuiltinInputs 中添加一行
package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// RegisterBuiltinInputs 注册所有内置输入插件到注册表。
func RegisterBuiltinInputs(registry *plugin.Registry, plat platform.Platform) error {
	inputs := []plugin.InputPlugin{
		NewPowerInput(plat),
		NewWiFiInput(plat),
		NewTimeInput(),
		NewNetworkInput(plat),
		NewProcessInput(plat),
		NewWindowInput(plat),
		NewIdleInput(plat),
		NewSysResInput(plat),
		NewDiskInput(plat),
		NewSessionInput(plat),
		NewClipboardInput(plat),
		NewDisplayInput(plat),
		NewUptimeInput(plat),
		NewOSInput(plat),
		NewCPUInput(plat),
		NewLocaleInput(plat),
		NewVolumeInput(plat),
		NewBatteryDetailInput(plat),
		NewNetworkStatsInput(plat),
		// 新增插件
		NewManualTriggerInput(),
		NewHTTPRequestInput(),
		NewNetworkDetectInput(),
		NewHTTPInInput(),
		NewTCPUDPInInput(),
		NewWebsocketInInput(),
		NewFileMonitorInput(),
	}

	for _, p := range inputs {
		if err := registry.RegisterInput(p, ""); err != nil {
			return err
		}
	}
	return nil
}
