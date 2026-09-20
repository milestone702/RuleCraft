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
	type reg struct {
		p   plugin.InputPlugin
		doc string
	}
	inputs := []reg{
		{NewPowerInput(plat), "电池与交流电源状态（充电/电量/AC）"},
		{NewWiFiInput(plat), "Wi-Fi 连接信息（SSID/BSSID/信号质量）"},
		{NewTimeInput(), "当前时间、星期、月份、时间戳"},
		{NewNetworkInput(plat), "网络适配器列表（IPv4/网关/连接类型）"},
		{NewProcessInput(plat), "进程列表与数量"},
		{NewWindowInput(plat), "前台窗口标题与进程名"},
		{NewIdleInput(plat), "用户输入空闲秒数"},
		{NewSysResInput(plat), "CPU / 内存使用率"},
		{NewDiskInput(plat), "磁盘剩余空间与使用率"},
		{NewSessionInput(plat), "会话锁屏状态"},
		{NewClipboardInput(plat), "剪贴板文本内容"},
		{NewDisplayInput(plat), "显示器分辨率、数量、DPI"},
		{NewUptimeInput(plat), "系统开机运行时长"},
		{NewOSInput(plat), "操作系统版本、计算机名、用户名"},
		{NewCPUInput(plat), "CPU 核心数与架构"},
		{NewLocaleInput(plat), "系统语言、区域、时区"},
		{NewVolumeInput(plat), "系统音量与静音状态"},
		{NewBatteryDetailInput(plat), "电池电压、循环次数、剩余时间"},
		{NewNetworkStatsInput(plat), "网络收发流量统计"},
		{NewManualTriggerInput(), "通过 API/UI 手动触发"},
		{NewHTTPRequestInput(), "定时 HTTP 请求采集"},
		{NewNetworkDetectInput(), "Ping / TCP / HTTP 连通性检测"},
		{NewHTTPInInput(), "HTTP 入站端点"},
		{NewTCPUDPInInput(), "TCP/UDP 入站监听"},
		{NewWebsocketInInput(), "WebSocket 入站连接"},
		{NewFileMonitorInput(), "文件/目录存在与变更监控"},
		{NewUSBInput(plat), "USB 设备插拔与列表"},
		{NewServiceInput(plat), "Windows 服务运行状态"},
		{NewBluetoothInput(plat), "蓝牙适配器与已连接设备"},
		{NewPowerPlanInput(plat), "当前电源计划"},
		{NewKeyboardInput(plat), "CapsLock / NumLock / ScrollLock"},
		{NewCursorInput(plat), "鼠标光标坐标"},
		{NewWindowModeInput(plat), "前台窗口是否全屏"},
		{NewInputLanguageInput(plat), "当前键盘布局/输入法语言"},
		{NewThemeInput(plat), "深色/浅色主题"},
		{NewPendingRebootInput(plat), "系统是否等待重启"},
		{NewProxyInput(plat), "系统代理开关与服务器"},
		{NewPrinterInput(plat), "默认打印机"},
		{NewWallpaperInput(plat), "当前壁纸路径"},
		{NewRecycleBinInput(plat), "回收站文件数与占用"},
		{NewGPUInput(plat), "主显卡名称"},
		{NewPublicIPInput(), "公网出口 IP"},
		{NewStartupInput(plat), "启动项列表与数量"},
		{NewScheduledTaskInput(plat), "计划任务运行状态"},
		{NewHostsMonitorInput(plat), "Hosts 文件变更监控"},
		{NewFirewallInput(plat), "防火墙配置文件状态"},
		{NewBatterySaverInput(plat), "节电模式开关"},
	}

	for _, item := range inputs {
		if err := registry.RegisterInput(item.p, item.doc); err != nil {
			return err
		}
	}
	return nil
}
