//go:build windows

package platform

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"rulecraft/config"
)

// ============================================================================
// 扩展传感器 3 — USB / 服务 / 蓝牙 / 电源计划
// ============================================================================

// ListUSBDevices 通过 PowerShell 枚举 PnP USB 设备。
func (p *WindowsPlatform) ListUSBDevices() ([]config.USBDeviceInfo, error) {
	script := `Get-PnpDevice -Class USB -Status OK -ErrorAction SilentlyContinue | Select-Object -Property FriendlyName,Status,InstanceId | ConvertTo-Json -Compress`
	out, err := runPowerShell(script, 10)
	if err != nil {
		return nil, err
	}
	return parseUSBDevicesJSON(out), nil
}

func parseUSBDevicesJSON(out string) []config.USBDeviceInfo {
	out = strings.TrimSpace(out)
	if out == "" {
		return nil
	}

	// 单对象或数组
	var raw interface{}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil
	}

	toDevice := func(m map[string]interface{}) config.USBDeviceInfo {
		d := config.USBDeviceInfo{}
		if v, ok := m["FriendlyName"].(string); ok {
			d.Name = v
		}
		if v, ok := m["InstanceId"].(string); ok {
			d.DeviceID = v
		}
		if v, ok := m["Status"].(string); ok {
			d.Present = strings.EqualFold(v, "OK")
		} else {
			d.Present = true
		}
		return d
	}

	switch v := raw.(type) {
	case map[string]interface{}:
		return []config.USBDeviceInfo{toDevice(v)}
	case []interface{}:
		list := make([]config.USBDeviceInfo, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				list = append(list, toDevice(m))
			}
		}
		return list
	}
	return nil
}

// GetServiceStatus 查询 Windows 服务状态。
func (p *WindowsPlatform) GetServiceStatus(name string) (*config.ServiceInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("service name is required")
	}
	// 避免注入：仅允许合理服务名字符
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '.') {
			return nil, fmt.Errorf("invalid service name")
		}
	}

	script := fmt.Sprintf(`Get-Service -Name '%s' -ErrorAction Stop | Select-Object Name,DisplayName,Status,StartType | ConvertTo-Json -Compress`, name)
	out, err := runPowerShell(script, 8)
	if err != nil {
		return nil, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return nil, fmt.Errorf("service %s not found", name)
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		return nil, fmt.Errorf("parse service status failed: %w", err)
	}

	info := &config.ServiceInfo{Name: name}
	if v, ok := m["DisplayName"].(string); ok {
		info.DisplayName = v
	}
	if v, ok := m["Status"].(string); ok {
		info.Status = v
	} else if v, ok := m["Status"].(float64); ok {
		// ServiceControllerStatus enum: Stopped=1, StartPending=2, ...
		info.Status = serviceStatusName(int(v))
	}
	if v, ok := m["StartType"].(string); ok {
		info.StartType = v
	} else if v, ok := m["StartType"].(float64); ok {
		info.StartType = serviceStartTypeName(int(v))
	}
	if v, ok := m["Name"].(string); ok {
		info.Name = v
	}
	return info, nil
}

func serviceStatusName(n int) string {
	switch n {
	case 1:
		return "Stopped"
	case 2:
		return "StartPending"
	case 3:
		return "StopPending"
	case 4:
		return "Running"
	case 5:
		return "ContinuePending"
	case 6:
		return "PausePending"
	case 7:
		return "Paused"
	default:
		return fmt.Sprintf("Unknown(%d)", n)
	}
}

func serviceStartTypeName(n int) string {
	switch n {
	case 0:
		return "Boot"
	case 1:
		return "System"
	case 2:
		return "Automatic"
	case 3:
		return "Manual"
	case 4:
		return "Disabled"
	default:
		return fmt.Sprintf("Unknown(%d)", n)
	}
}

// ListBluetoothDevices 列出蓝牙设备。
func (p *WindowsPlatform) ListBluetoothDevices() ([]config.BluetoothDeviceInfo, error) {
	script := `Get-PnpDevice -Class Bluetooth -ErrorAction SilentlyContinue | Select-Object FriendlyName,InstanceId,Status | ConvertTo-Json -Compress`
	out, err := runPowerShell(script, 12)
	if err != nil {
		return nil, err
	}
	return parseBluetoothJSON(out), nil
}

func parseBluetoothJSON(out string) []config.BluetoothDeviceInfo {
	out = strings.TrimSpace(out)
	if out == "" {
		return nil
	}
	var raw interface{}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil
	}

	toDev := func(m map[string]interface{}) config.BluetoothDeviceInfo {
		d := config.BluetoothDeviceInfo{}
		if v, ok := m["FriendlyName"].(string); ok {
			d.Name = v
		}
		if v, ok := m["InstanceId"].(string); ok {
			d.Address = v
			// MAC 形式的 InstanceId 特征
			d.Paired = strings.Contains(strings.ToUpper(v), "BTHENUM")
		}
		if v, ok := m["Status"].(string); ok {
			d.Connected = strings.EqualFold(v, "OK")
		}
		return d
	}

	switch v := raw.(type) {
	case map[string]interface{}:
		return []config.BluetoothDeviceInfo{toDev(v)}
	case []interface{}:
		list := make([]config.BluetoothDeviceInfo, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				list = append(list, toDev(m))
			}
		}
		return list
	}
	return nil
}

// GetActivePowerPlan 获取当前电源计划。
func (p *WindowsPlatform) GetActivePowerPlan() (string, error) {
	var scheme *syscall.GUID
	ret, _, _ := procPowerGetActiveScheme.Call(0, uintptr(unsafe.Pointer(&scheme)))
	if ret != 0 || scheme == nil {
		return "", fmt.Errorf("PowerGetActiveScheme failed: %d", ret)
	}
	defer windowsLocalFree(uintptr(unsafe.Pointer(scheme)))

	guid := fmt.Sprintf("%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		scheme.Data1, scheme.Data2, scheme.Data3,
		scheme.Data4[0], scheme.Data4[1], scheme.Data4[2], scheme.Data4[3],
		scheme.Data4[4], scheme.Data4[5], scheme.Data4[6], scheme.Data4[7])

	// 映射常见 GUID 为友好名
	switch {
	case strings.Contains(guid, "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c"):
		return "high_performance", nil
	case strings.Contains(guid, "a1841308-3541-4fab-bc81-f71556f20b4a"):
		return "power_saver", nil
	case strings.Contains(guid, "381b4222-f694-41f0-9685-ff5bb260df2e"):
		return "balanced", nil
	}
	return guid, nil
}

// windowsLocalFree 释放 LocalAlloc 内存（PowerGetActiveScheme 分配的 GUID）。
func windowsLocalFree(h uintptr) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	kernel32.NewProc("LocalFree").Call(h)
}

// ============================================================================
// 扩展控制 3 — 静音 / 重启资源管理器 / 打开控制面板
// ============================================================================

// SetMute 通过 Core Audio（PowerShell 内嵌 C#）设置系统静音。
func (p *WindowsPlatform) SetMute(muted bool) error {
	muteVal := "0"
	if muted {
		muteVal = "1"
	}
	script := fmt.Sprintf(`
$code = @"
using System;
using System.Runtime.InteropServices;
[Guid("5CDF2C82-841E-4546-9722-0CF74078229A"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IAudioEndpointVolume {
  int NotImpl1(); int NotImpl2();
  int SetMasterVolumeLevelScalar(float fLevel, System.Guid pguidEventContext);
  int GetMasterVolumeLevelScalar(out float pfLevel);
  int NotImpl3(); int NotImpl4(); int NotImpl5(); int NotImpl6();
  int SetMute([MarshalAs(UnmanagedType.Bool)] bool bMute, System.Guid pguidEventContext);
  int GetMute(out bool pbMute);
}
[Guid("D666063F-1587-4E43-81F1-B948E807363F"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IMMDevice {
  int Activate(ref Guid iid, int clsCtx, IntPtr activationParams, [MarshalAs(UnmanagedType.IUnknown)] out object interfacePointer);
}
[Guid("A95664D2-9614-4F35-A746-DE8DB63617E6"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IMMDeviceEnumerator {
  int NotImpl1();
  int GetDefaultAudioEndpoint(int dataFlow, int role, out IMMDevice device);
}
[Guid("BCDE0395-E52F-467C-8E3D-C4579291692E")]
class MMDeviceEnumeratorComObject { }
public class Audio {
  public static void SetMute(int mute) {
    var enumerator = new MMDeviceEnumeratorComObject() as IMMDeviceEnumerator;
    IMMDevice dev;
    enumerator.GetDefaultAudioEndpoint(0 /*render*/, 1 /*multimedia*/, out dev);
    Guid iid = typeof(IAudioEndpointVolume).GUID;
    object o;
    dev.Activate(ref iid, 23 /*ALL*/, IntPtr.Zero, out o);
    var vol = o as IAudioEndpointVolume;
    vol.SetMute(mute != 0, Guid.Empty);
  }
}
"@
Add-Type -TypeDefinition $code -ErrorAction Stop
[Audio]::SetMute(%s)
`, muteVal)
	_, err := runPowerShell(script, 15)
	return err
}

// GetMuteState 查询当前是否静音。
func (p *WindowsPlatform) GetMuteState() (bool, error) {
	script := `
$code = @"
using System;
using System.Runtime.InteropServices;
[Guid("5CDF2C82-841E-4546-9722-0CF74078229A"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IAudioEndpointVolume {
  int NotImpl1(); int NotImpl2();
  int SetMasterVolumeLevelScalar(float fLevel, System.Guid pguidEventContext);
  int GetMasterVolumeLevelScalar(out float pfLevel);
  int NotImpl3(); int NotImpl4(); int NotImpl5(); int NotImpl6();
  int SetMute([MarshalAs(UnmanagedType.Bool)] bool bMute, System.Guid pguidEventContext);
  int GetMute(out bool pbMute);
}
[Guid("D666063F-1587-4E43-81F1-B948E807363F"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IMMDevice {
  int Activate(ref Guid iid, int clsCtx, IntPtr activationParams, [MarshalAs(UnmanagedType.IUnknown)] out object interfacePointer);
}
[Guid("A95664D2-9614-4F35-A746-DE8DB63617E6"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IMMDeviceEnumerator {
  int NotImpl1();
  int GetDefaultAudioEndpoint(int dataFlow, int role, out IMMDevice device);
}
[Guid("BCDE0395-E52F-467C-8E3D-C4579291692E")]
class MMDeviceEnumeratorComObject { }
public class Audio {
  public static int GetMute() {
    var enumerator = new MMDeviceEnumeratorComObject() as IMMDeviceEnumerator;
    IMMDevice dev;
    enumerator.GetDefaultAudioEndpoint(0, 1, out dev);
    Guid iid = typeof(IAudioEndpointVolume).GUID;
    object o;
    dev.Activate(ref iid, 23, IntPtr.Zero, out o);
    var vol = o as IAudioEndpointVolume;
    bool muted;
    vol.GetMute(out muted);
    return muted ? 1 : 0;
  }
}
"@
Add-Type -TypeDefinition $code -ErrorAction Stop
[Audio]::GetMute()
`
	out, err := runPowerShell(script, 15)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "1", nil
}

// ToggleMute 切换静音并返回新状态。
func (p *WindowsPlatform) ToggleMute() (bool, error) {
	cur, err := p.GetMuteState()
	if err != nil {
		return false, err
	}
	next := !cur
	if err := p.SetMute(next); err != nil {
		return false, err
	}
	return next, nil
}

// RestartExplorer 重启 Windows 资源管理器。
func (p *WindowsPlatform) RestartExplorer() error {
	// 先结束 explorer.exe，再启动
	kill := exec.Command("taskkill", "/f", "/im", "explorer.exe")
	kill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = kill.Run() // explorer 未运行时忽略错误

	start := exec.Command("explorer.exe")
	start.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := start.Start(); err != nil {
		return fmt.Errorf("start explorer failed: %w", err)
	}
	return nil
}

// OpenControlPanelPage 打开控制面板页或 Settings URI。
// page 支持：
//   - 控制面板项：name Microsoft.Windows.ErrorReporting 或经典页面 ID
//   - ms-settings: URI
//   - control.exe 可识别的名称
func (p *WindowsPlatform) OpenControlPanelPage(page string) error {
	if page == "" {
		return fmt.Errorf("page is required")
	}

	// Settings URI
	if strings.HasPrefix(page, "ms-settings:") {
		cmd := exec.Command("cmd", "/c", "start", "", page)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		return cmd.Start()
	}

	// 经典控制面板
	cmd := exec.Command("control.exe", page)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open control panel failed: %w", err)
	}
	return nil
}

// runPowerShell 隐藏窗口执行 PowerShell 命令并返回 stdout。
func runPowerShell(script string, timeoutSec int) (string, error) {
	if timeoutSec <= 0 {
		timeoutSec = 10
	}
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("powershell failed: %v: %s", err, string(ee.Stderr))
		}
		return "", fmt.Errorf("powershell failed: %w", err)
	}
	return string(out), nil
}
