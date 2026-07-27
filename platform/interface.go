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
