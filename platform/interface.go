// Package platform 提供操作系统抽象层，隔离所有 Win32/Linux/macOS 系统调用。
//
// 架构目标：
//   - 所有输入/输出插件通过 Platform 接口调用 OS 功能，而非直接调用 Win32 API
//   - 移植到新 OS 时只需实现 Platform 接口，所有插件逻辑不动
//   - 使用 Go build tags 实现条件编译
//
// 文件分工：
//   interface.go       — Platform 接口定义 + NoopPlatform（无平台标签，所有 OS 都编译）
//   platform_windows.go — Windows 实现（//go:build windows）
//   platform_stub.go    — 非 Windows 平台的桩实现（//go:build !windows）
package platform

import (
	"rulecraft/config"
)

// ============================================================================
// Platform 接口
// ============================================================================

// Platform 是完整的操作系统抽象接口，涵盖所有传感器采集和系统控制操作。
// 每个方法返回明确的错误，在功能不可用时应返回 ErrNotImplemented。
type Platform interface {
	// ======================================================================
	// 传感器（Sensors）— 供输入插件使用
	// ======================================================================

	// CollectWiFiInfo 采集 Wi-Fi 状态。
	CollectWiFiInfo() (*config.WiFiInfo, error)

	// CollectPowerInfo 采集电源状态。
	CollectPowerInfo() (*config.PowerInfo, error)

	// CollectNetworkInfo 采集网络适配器信息。
	CollectNetworkInfo() ([]config.NetworkAdapterInfo, error)

	// ListProcesses 列出当前运行的进程（含内存、路径等详细信息）。
	ListProcesses() ([]config.ProcessInfo, error)

	// GetForegroundWindowInfo 获取前台窗口信息。
	GetForegroundWindowInfo() (*config.WindowInfo, error)

	// GetIdleSeconds 获取用户空闲秒数。
	GetIdleSeconds() (uint64, error)

	// GetSystemResources 采集 CPU 和内存使用率。
	GetSystemResources() (*config.SysResInfo, error)

	// GetDiskInfo 采集指定驱动器的磁盘信息。drive 如 "C:"。
	GetDiskInfo(drive string) (*config.DiskInfo, error)

	// GetSessionInfo 获取会话状态（锁屏/登录）。
	GetSessionInfo() (*config.SessionInfo, error)

	// ======================================================================
	// 扩展传感器（2025 Added）
	// ======================================================================

	// GetClipboardText 获取剪贴板文本内容。
	GetClipboardText() (string, error)

	// GetDisplayInfo 获取显示器信息。
	GetDisplayInfo() (*config.DisplayInfo, error)

	// GetUptimeSeconds 获取系统启动以来的秒数。
	GetUptimeSeconds() (uint64, error)

	// GetOSInfo 获取操作系统信息。
	GetOSInfo() (*config.OSInfo, error)

	// ======================================================================
	// 扩展控制（2025 Added）
	// ======================================================================

	// OpenWithDefault 使用默认程序打开文件或 URL。
	OpenWithDefault(target string) error

	// ControlWindow 控制窗口状态（minimize/maximize/restore/close/focus）。
	ControlWindow(title string, action string) error

	// ======================================================================
	// 扩展传感器 2（2025 Added）
	// ======================================================================

	// GetCPUInfo 获取 CPU 信息。
	GetCPUInfo() (*config.CPUInfo, error)

	// GetLocaleInfo 获取系统区域和时区信息。
	GetLocaleInfo() (*config.LocaleInfo, error)

	// GetSystemVolumeLevel 获取系统音量（0-100）和静音状态。
	GetSystemVolumeLevel() (level int, muted bool, err error)

	// GetExtendedBatteryInfo 获取电池详细信息。
	GetExtendedBatteryInfo() (*config.BatteryDetail, error)

	// GetNetworkTraffic 获取网络流量统计。
	GetNetworkTraffic() (*config.NetworkTraffic, error)

	// ======================================================================
	// 扩展控制 2（2025 Added）
	// ======================================================================

	// TakeScreenshot 截屏保存到指定路径。
	TakeScreenshot(path string) error

	// SetClipboardText 设置剪贴板文本。
	SetClipboardText(text string) error

	// ======================================================================
	// 控制（Actuators）— 供输出插件使用
	// ======================================================================

	// PreventSleep 阻止或允许系统休眠。
	PreventSleep(prevent bool) error

	// ExecuteProcess 执行一个外部程序/脚本。
	// 返回进程 PID。
	ExecuteProcess(cmd string, args []string, workingDir string, env map[string]string) (int, error)

	// SetVolume 设置系统主音量（0-100）。
	SetVolume(level int) error

	// SetBrightness 设置屏幕亮度（0-100）。
	SetBrightness(level int) error

	// ShowNotification 显示桌面通知。
	ShowNotification(title, message string, level string) error

	// LockWorkstation 锁定工作站。
	LockWorkstation() error

	// PowerAction 执行电源操作（shutdown / restart / sleep / hibernate）。
	PowerAction(action string) error

	// SetWallpaper 设置桌面壁纸。
	SetWallpaper(path string) error

	// KillProcess 强制终止进程。
	KillProcess(pid int) error

	// SetPowerScheme 切换电源方案（power_saver / balanced / high_performance）。
	SetPowerScheme(scheme string) error

	// SetNetworkAdapter 启用或禁用网络适配器。
	SetNetworkAdapter(name string, enabled bool) error

	// ======================================================================
	// 扩展传感器 3
	// ======================================================================

	// ListUSBDevices 列出当前 USB 设备。
	ListUSBDevices() ([]config.USBDeviceInfo, error)

	// GetServiceStatus 查询 Windows 服务状态。
	GetServiceStatus(name string) (*config.ServiceInfo, error)

	// ListBluetoothDevices 列出蓝牙设备。
	ListBluetoothDevices() ([]config.BluetoothDeviceInfo, error)

	// GetActivePowerPlan 获取当前电源计划（GUID 或友好名）。
	GetActivePowerPlan() (string, error)

	// ======================================================================
	// 扩展控制 3
	// ======================================================================

	// SetMute 设置系统静音/取消静音。
	SetMute(muted bool) error

	// ToggleMute 切换静音状态，返回切换后的 muted。
	ToggleMute() (bool, error)

	// RestartExplorer 重启 Windows 资源管理器。
	RestartExplorer() error

	// OpenControlPanelPage 打开控制面板页或 Settings URI。
	OpenControlPanelPage(page string) error

	// ======================================================================
	// 扩展传感器 4
	// ======================================================================

	// GetKeyboardLockStates 返回 CapsLock / NumLock / ScrollLock 状态。
	GetKeyboardLockStates() (capsLock, numLock, scrollLock bool, err error)

	// GetCursorPosition 获取鼠标光标屏幕坐标。
	GetCursorPosition() (x, y int, err error)

	// IsForegroundFullscreen 判断前台窗口是否全屏。
	IsForegroundFullscreen() (bool, error)

	// GetInputLanguage 返回当前键盘布局语言标识（如 "0804" / "0409"）。
	GetInputLanguage() (string, error)

	// GetThemeMode 返回系统/应用主题（"dark" / "light"）。
	GetThemeMode() (string, error)

	// GetPendingReboot 检查系统是否等待重启；reason 为简要说明。
	GetPendingReboot() (pending bool, reason string, err error)

	// GetProxyInfo 返回系统代理状态。
	GetProxyInfo() (enabled bool, server string, err error)

	// GetDefaultPrinter 获取默认打印机名称。
	GetDefaultPrinter() (string, error)

	// GetWallpaperPath 获取当前壁纸路径。
	GetWallpaperPath() (string, error)

	// GetRecycleBinInfo 获取回收站文件数与约占用 MB。
	GetRecycleBinInfo() (count int, sizeMB float64, err error)

	// GetGPUInfo 获取主显卡名称（简要）。
	GetGPUInfo() (name string, err error)

	// ======================================================================
	// 扩展控制 4
	// ======================================================================

	// KillProcessByName 按进程名（可不含 .exe）终止全部匹配进程。
	KillProcessByName(name string) error

	// MediaKey 发送多媒体按键（play_pause / next / prev / volume_up / volume_down / stop）。
	MediaKey(action string) error

	// FlushDNS 清空 DNS 解析缓存。
	FlushDNS() error

	// SetDNS 设置指定适配器的 DNS（server 为空则自动获取）。
	SetDNS(adapter, server string) error

	// SetThemeMode 设置系统/应用主题（"dark" / "light"）。
	SetThemeMode(mode string) error

	// EmptyRecycleBin 清空回收站。
	EmptyRecycleBin() error

	// ======================================================================
	// 扩展传感器 5
	// ======================================================================

	// GetStartupItems 列出启动项名称（Run 键 + Startup 文件夹快捷方式名）。
	GetStartupItems() ([]string, error)

	// GetScheduledTaskStatus 查询计划任务状态；found=false 表示不存在。
	GetScheduledTaskStatus(name string) (status string, found bool, err error)

	// GetHostsInfo 返回 hosts 文件修改时间 Unix 秒与行数。
	GetHostsInfo() (mtime int64, lines int, err error)

	// GetFirewallStatus 返回各防火墙配置文件是否启用（domain/private/public）。
	GetFirewallStatus() (domain, private, public bool, err error)

	// IsBatterySaver 是否处于节电模式。
	IsBatterySaver() (bool, error)

	// ======================================================================
	// 扩展控制 5
	// ======================================================================

	// SetWindowTopmost 将标题匹配的窗口设为置顶/取消置顶。
	SetWindowTopmost(title string, topmost bool) error

	// SetProcessPriority 设置指定进程优先级（idle/below_normal/normal/above_normal/high）。
	SetProcessPriority(name, priority string) error
}

// ============================================================================
// 错误定义
// ============================================================================

// ErrNotImplemented 表示该功能在当前平台上未实现。
var ErrNotImplemented = &NotImplementedError{}

// NotImplementedError 功能未实现错误。
type NotImplementedError struct {
	Method string // 方法名
}

func (e *NotImplementedError) Error() string {
	if e.Method != "" {
		return "not implemented on this platform: " + e.Method
	}
	return "not implemented on this platform"
}

// IsErrNotImplemented 判断是否为未实现错误。
func IsErrNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotImplementedError)
	return ok
}

// ============================================================================
// NoopPlatform — 所有方法返回 ErrNotImplemented 的默认实现
// ============================================================================

// NoopPlatform 是 Platform 接口的占位实现，所有方法返回 ErrNotImplemented。
// 在尚未实现平台层的系统上作为安全降级使用。
type NoopPlatform struct{}

func (p *NoopPlatform) CollectWiFiInfo() (*config.WiFiInfo, error) {
	return nil, &NotImplementedError{Method: "CollectWiFiInfo"}
}
func (p *NoopPlatform) CollectPowerInfo() (*config.PowerInfo, error) {
	return nil, &NotImplementedError{Method: "CollectPowerInfo"}
}
func (p *NoopPlatform) CollectNetworkInfo() ([]config.NetworkAdapterInfo, error) {
	return nil, &NotImplementedError{Method: "CollectNetworkInfo"}
}
func (p *NoopPlatform) ListProcesses() ([]config.ProcessInfo, error) {
	return nil, &NotImplementedError{Method: "ListProcesses"}
}
func (p *NoopPlatform) GetForegroundWindowInfo() (*config.WindowInfo, error) {
	return nil, &NotImplementedError{Method: "GetForegroundWindowInfo"}
}
func (p *NoopPlatform) GetIdleSeconds() (uint64, error) {
	return 0, &NotImplementedError{Method: "GetIdleSeconds"}
}
func (p *NoopPlatform) GetSystemResources() (*config.SysResInfo, error) {
	return nil, &NotImplementedError{Method: "GetSystemResources"}
}
func (p *NoopPlatform) GetDiskInfo(drive string) (*config.DiskInfo, error) {
	return nil, &NotImplementedError{Method: "GetDiskInfo"}
}
func (p *NoopPlatform) GetSessionInfo() (*config.SessionInfo, error) {
	return nil, &NotImplementedError{Method: "GetSessionInfo"}
}
func (p *NoopPlatform) PreventSleep(prevent bool) error {
	return &NotImplementedError{Method: "PreventSleep"}
}
func (p *NoopPlatform) ExecuteProcess(cmd string, args []string, workingDir string, env map[string]string) (int, error) {
	return 0, &NotImplementedError{Method: "ExecuteProcess"}
}
func (p *NoopPlatform) SetVolume(level int) error {
	return &NotImplementedError{Method: "SetVolume"}
}
func (p *NoopPlatform) SetBrightness(level int) error {
	return &NotImplementedError{Method: "SetBrightness"}
}
func (p *NoopPlatform) ShowNotification(title, message string, level string) error {
	return &NotImplementedError{Method: "ShowNotification"}
}
func (p *NoopPlatform) LockWorkstation() error {
	return &NotImplementedError{Method: "LockWorkstation"}
}
func (p *NoopPlatform) PowerAction(action string) error {
	return &NotImplementedError{Method: "PowerAction"}
}
func (p *NoopPlatform) SetWallpaper(path string) error {
	return &NotImplementedError{Method: "SetWallpaper"}
}
func (p *NoopPlatform) KillProcess(pid int) error {
	return &NotImplementedError{Method: "KillProcess"}
}
func (p *NoopPlatform) SetPowerScheme(scheme string) error {
	return &NotImplementedError{Method: "SetPowerScheme"}
}
func (p *NoopPlatform) SetNetworkAdapter(name string, enabled bool) error {
	return &NotImplementedError{Method: "SetNetworkAdapter"}
}
func (p *NoopPlatform) GetClipboardText() (string, error) {
	return "", &NotImplementedError{Method: "GetClipboardText"}
}
func (p *NoopPlatform) GetDisplayInfo() (*config.DisplayInfo, error) {
	return nil, &NotImplementedError{Method: "GetDisplayInfo"}
}
func (p *NoopPlatform) GetUptimeSeconds() (uint64, error) {
	return 0, &NotImplementedError{Method: "GetUptimeSeconds"}
}
func (p *NoopPlatform) GetOSInfo() (*config.OSInfo, error) {
	return nil, &NotImplementedError{Method: "GetOSInfo"}
}
func (p *NoopPlatform) OpenWithDefault(target string) error {
	return &NotImplementedError{Method: "OpenWithDefault"}
}
func (p *NoopPlatform) ControlWindow(title string, action string) error {
	return &NotImplementedError{Method: "ControlWindow"}
}
func (p *NoopPlatform) GetCPUInfo() (*config.CPUInfo, error) {
	return nil, &NotImplementedError{Method: "GetCPUInfo"}
}
func (p *NoopPlatform) GetLocaleInfo() (*config.LocaleInfo, error) {
	return nil, &NotImplementedError{Method: "GetLocaleInfo"}
}
func (p *NoopPlatform) GetSystemVolumeLevel() (int, bool, error) {
	return 0, false, &NotImplementedError{Method: "GetSystemVolumeLevel"}
}
func (p *NoopPlatform) GetExtendedBatteryInfo() (*config.BatteryDetail, error) {
	return nil, &NotImplementedError{Method: "GetExtendedBatteryInfo"}
}
func (p *NoopPlatform) GetNetworkTraffic() (*config.NetworkTraffic, error) {
	return nil, &NotImplementedError{Method: "GetNetworkTraffic"}
}
func (p *NoopPlatform) TakeScreenshot(path string) error {
	return &NotImplementedError{Method: "TakeScreenshot"}
}
func (p *NoopPlatform) SetClipboardText(text string) error {
	return &NotImplementedError{Method: "SetClipboardText"}
}
func (p *NoopPlatform) ListUSBDevices() ([]config.USBDeviceInfo, error) {
	return nil, &NotImplementedError{Method: "ListUSBDevices"}
}
func (p *NoopPlatform) GetServiceStatus(name string) (*config.ServiceInfo, error) {
	return nil, &NotImplementedError{Method: "GetServiceStatus"}
}
func (p *NoopPlatform) ListBluetoothDevices() ([]config.BluetoothDeviceInfo, error) {
	return nil, &NotImplementedError{Method: "ListBluetoothDevices"}
}
func (p *NoopPlatform) GetActivePowerPlan() (string, error) {
	return "", &NotImplementedError{Method: "GetActivePowerPlan"}
}
func (p *NoopPlatform) SetMute(muted bool) error {
	return &NotImplementedError{Method: "SetMute"}
}
func (p *NoopPlatform) ToggleMute() (bool, error) {
	return false, &NotImplementedError{Method: "ToggleMute"}
}
func (p *NoopPlatform) RestartExplorer() error {
	return &NotImplementedError{Method: "RestartExplorer"}
}
func (p *NoopPlatform) OpenControlPanelPage(page string) error {
	return &NotImplementedError{Method: "OpenControlPanelPage"}
}
func (p *NoopPlatform) GetKeyboardLockStates() (bool, bool, bool, error) {
	return false, false, false, &NotImplementedError{Method: "GetKeyboardLockStates"}
}
func (p *NoopPlatform) GetCursorPosition() (int, int, error) {
	return 0, 0, &NotImplementedError{Method: "GetCursorPosition"}
}
func (p *NoopPlatform) IsForegroundFullscreen() (bool, error) {
	return false, &NotImplementedError{Method: "IsForegroundFullscreen"}
}
func (p *NoopPlatform) GetInputLanguage() (string, error) {
	return "", &NotImplementedError{Method: "GetInputLanguage"}
}
func (p *NoopPlatform) GetThemeMode() (string, error) {
	return "", &NotImplementedError{Method: "GetThemeMode"}
}
func (p *NoopPlatform) GetPendingReboot() (bool, string, error) {
	return false, "", &NotImplementedError{Method: "GetPendingReboot"}
}
func (p *NoopPlatform) GetProxyInfo() (bool, string, error) {
	return false, "", &NotImplementedError{Method: "GetProxyInfo"}
}
func (p *NoopPlatform) GetDefaultPrinter() (string, error) {
	return "", &NotImplementedError{Method: "GetDefaultPrinter"}
}
func (p *NoopPlatform) GetWallpaperPath() (string, error) {
	return "", &NotImplementedError{Method: "GetWallpaperPath"}
}
func (p *NoopPlatform) GetRecycleBinInfo() (int, float64, error) {
	return 0, 0, &NotImplementedError{Method: "GetRecycleBinInfo"}
}
func (p *NoopPlatform) GetGPUInfo() (string, error) {
	return "", &NotImplementedError{Method: "GetGPUInfo"}
}
func (p *NoopPlatform) KillProcessByName(name string) error {
	return &NotImplementedError{Method: "KillProcessByName"}
}
func (p *NoopPlatform) MediaKey(action string) error {
	return &NotImplementedError{Method: "MediaKey"}
}
func (p *NoopPlatform) FlushDNS() error {
	return &NotImplementedError{Method: "FlushDNS"}
}
func (p *NoopPlatform) SetDNS(adapter, server string) error {
	return &NotImplementedError{Method: "SetDNS"}
}
func (p *NoopPlatform) SetThemeMode(mode string) error {
	return &NotImplementedError{Method: "SetThemeMode"}
}
func (p *NoopPlatform) EmptyRecycleBin() error {
	return &NotImplementedError{Method: "EmptyRecycleBin"}
}
func (p *NoopPlatform) GetStartupItems() ([]string, error) {
	return nil, &NotImplementedError{Method: "GetStartupItems"}
}
func (p *NoopPlatform) GetScheduledTaskStatus(name string) (string, bool, error) {
	return "", false, &NotImplementedError{Method: "GetScheduledTaskStatus"}
}
func (p *NoopPlatform) GetHostsInfo() (int64, int, error) {
	return 0, 0, &NotImplementedError{Method: "GetHostsInfo"}
}
func (p *NoopPlatform) GetFirewallStatus() (bool, bool, bool, error) {
	return false, false, false, &NotImplementedError{Method: "GetFirewallStatus"}
}
func (p *NoopPlatform) IsBatterySaver() (bool, error) {
	return false, &NotImplementedError{Method: "IsBatterySaver"}
}
func (p *NoopPlatform) SetWindowTopmost(title string, topmost bool) error {
	return &NotImplementedError{Method: "SetWindowTopmost"}
}
func (p *NoopPlatform) SetProcessPriority(name, priority string) error {
	return &NotImplementedError{Method: "SetProcessPriority"}
}
