//go:build windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// ============================================================================
// 扩展传感器 4 — 键盘 / 光标 / 窗口 / 语言 / 主题 / 重启 / 代理 / 打印机等
// ============================================================================

const (
	vkCapital  = 0x14
	vkNumLock  = 0x90
	vkScroll   = 0x91
	spiGetDeskWallpaper = 0x0073
	spiFUpdateInifile   = 0x0001
	spiFSendWinIniChange = 0x0002
)

var (
	procGetKeyState      = modUser32.NewProc("GetKeyState")
	procGetCursorPos     = modUser32.NewProc("GetCursorPos")
	procGetKeyboardLayout = modUser32.NewProc("GetKeyboardLayout")
	procSystemParametersInfoW2 = modUser32.NewProc("SystemParametersInfoW")
	procGetSystemMetrics = modUser32.NewProc("GetSystemMetrics")
)

// GetKeyboardLockStates 读取锁定键状态。
func (p *WindowsPlatform) GetKeyboardLockStates() (bool, bool, bool, error) {
	caps, _, _ := procGetKeyState.Call(uintptr(vkCapital))
	num, _, _ := procGetKeyState.Call(uintptr(vkNumLock))
	scroll, _, _ := procGetKeyState.Call(uintptr(vkScroll))
	// 最低位为切换状态
	return caps&1 != 0, num&1 != 0, scroll&1 != 0, nil
}

// GetCursorPosition 获取鼠标坐标。
func (p *WindowsPlatform) GetCursorPosition() (int, int, error) {
	var pt struct{ x, y int32 }
	ret, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return 0, 0, fmt.Errorf("GetCursorPos failed")
	}
	return int(pt.x), int(pt.y), nil
}

// IsForegroundFullscreen 判断前台窗口是否占满主屏幕。
func (p *WindowsPlatform) IsForegroundFullscreen() (bool, error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return false, nil
	}

	var rect struct{ left, top, right, bottom int32 }
	procGetWindowRect, _ := user32Rect()
	if procGetWindowRect == nil {
		return false, fmt.Errorf("GetWindowRect unavailable")
	}
	ret, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	if ret == 0 {
		return false, fmt.Errorf("GetWindowRect failed")
	}

	// 屏幕尺寸
	sw, _, _ := procGetSystemMetrics.Call(0) // SM_CXSCREEN
	sh, _, _ := procGetSystemMetrics.Call(1) // SM_CYSCREEN
	return rect.left == 0 && rect.top == 0 &&
		int(rect.right) == int(sw) && int(rect.bottom) == int(sh), nil
}

func user32Rect() (*syscall.LazyProc, error) {
	return modUser32.NewProc("GetWindowRect"), nil
}

// GetInputLanguage 返回前台线程键盘布局 LANGID（hex 字符串）。
func (p *WindowsPlatform) GetInputLanguage() (string, error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	var threadID uint32
	if hwnd != 0 {
		procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&threadID)))
	}
	kl, _, _ := procGetKeyboardLayout.Call(uintptr(threadID))
	langID := uint16(kl & 0xFFFF)
	return fmt.Sprintf("%04X", langID), nil
}

// GetThemeMode 读取应用深色/浅色模式。
func (p *WindowsPlatform) GetThemeMode() (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return "light", nil // 旧系统无该键，视为浅色
	}
	defer key.Close()

	val, _, err := key.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return "light", nil
	}
	if val == 0 {
		return "dark", nil
	}
	return "light", nil
}

// SetThemeMode 设置应用深色/浅色模式并广播变更。
func (p *WindowsPlatform) SetThemeMode(mode string) error {
	var useLight uint64 = 1
	switch strings.ToLower(mode) {
	case "dark", "0":
		useLight = 0
	case "light", "1":
		useLight = 1
	default:
		return fmt.Errorf("theme mode must be dark or light")
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open personalize key: %w", err)
	}
	defer key.Close()

	if err := key.SetDWordValue("AppsUseLightTheme", uint32(useLight)); err != nil {
		return err
	}
	_ = key.SetDWordValue("SystemUsesLightTheme", uint32(useLight))

	// 通知系统刷新
	const wmSettingChange = 0x001A
	const hwndBroadcast = 0xFFFF
	procSendMessageW := modUser32.NewProc("SendMessageW")
	param, _ := syscall.UTF16PtrFromString("ImmersiveColorSet")
	procSendMessageW.Call(hwndBroadcast, wmSettingChange, 0, uintptr(unsafe.Pointer(param)))
	return nil
}

// GetPendingReboot 检查是否等待重启。
func (p *WindowsPlatform) GetPendingReboot() (bool, string, error) {
	var reasons []string

	// CBS
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Component Based Servicing`, registry.QUERY_VALUE); err == nil {
		if sk, e := registry.OpenKey(k, "RebootPending", registry.QUERY_VALUE); e == nil {
			sk.Close()
			reasons = append(reasons, "CBS")
		}
		k.Close()
	}

	// Windows Update
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update`, registry.QUERY_VALUE); err == nil {
		if sk, e := registry.OpenKey(k, "RebootRequired", registry.QUERY_VALUE); e == nil {
			sk.Close()
			reasons = append(reasons, "WindowsUpdate")
		}
		k.Close()
	}

	// Pending file rename
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager`, registry.QUERY_VALUE); err == nil {
		if _, _, e := k.GetStringValue("PendingFileRenameOperations"); e == nil {
			reasons = append(reasons, "PendingFileRename")
		}
		k.Close()
	}

	if len(reasons) == 0 {
		return false, "", nil
	}
	return true, strings.Join(reasons, ","), nil
}

// GetProxyInfo 读取 WinINET 代理设置。
func (p *WindowsPlatform) GetProxyInfo() (bool, string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return false, "", nil
	}
	defer key.Close()

	enabled := false
	if v, _, err := key.GetIntegerValue("ProxyEnable"); err == nil {
		enabled = v == 1
	}
	server := ""
	if v, _, err := key.GetStringValue("ProxyServer"); err == nil {
		server = v
	}
	return enabled, server, nil
}

// GetDefaultPrinter 获取默认打印机名。
func (p *WindowsPlatform) GetDefaultPrinter() (string, error) {
	out, err := runPowerShell(`(Get-CimInstance -ClassName Win32_Printer -Filter "Default=TRUE").Name`, 8)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetWallpaperPath 获取当前壁纸路径。
func (p *WindowsPlatform) GetWallpaperPath() (string, error) {
	buf := make([]uint16, 260)
	ret, _, _ := procSystemParametersInfoW2.Call(
		uintptr(spiGetDeskWallpaper),
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(&buf[0])),
		0,
	)
	if ret == 0 {
		return "", fmt.Errorf("SystemParametersInfo SPI_GETDESKWALLPAPER failed")
	}
	return syscall.UTF16ToString(buf), nil
}

// GetRecycleBinInfo 统计回收站。
func (p *WindowsPlatform) GetRecycleBinInfo() (int, float64, error) {
	// $Recycle.Bin 下当前用户 SID 子目录
	sid, err := currentUserSID()
	if err != nil {
		return 0, 0, err
	}
	root := filepath.Join(`C:\$Recycle.Bin`, sid)
	count := 0
	var size int64
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		count++
		size += info.Size()
		return nil
	})
	return count, float64(size) / (1024 * 1024), nil
}

func currentUserSID() (string, error) {
	out, err := runPowerShell(`(New-Object System.Security.Principal.WindowsIdentity).User.Value`, 5)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetGPUInfo 获取主显卡名称。
func (p *WindowsPlatform) GetGPUInfo() (string, error) {
	out, err := runPowerShell(`(Get-CimInstance Win32_VideoController | Select-Object -First 1).Name`, 10)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ============================================================================
// 扩展控制 4
// ============================================================================

// KillProcessByName 按名称终止进程。
func (p *WindowsPlatform) KillProcessByName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("process name is required")
	}
	if !strings.HasSuffix(strings.ToLower(name), ".exe") {
		name += ".exe"
	}
	cmd := exec.Command("taskkill", "/f", "/im", name)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.CombinedOutput(); err != nil {
		// 没有匹配进程时 taskkill 返回非 0
		s := string(out)
		if strings.Contains(s, "not found") || strings.Contains(s, "No tasks") || strings.Contains(s, "没有找到") {
			return nil
		}
		return fmt.Errorf("taskkill %s: %v: %s", name, err, s)
	}
	return nil
}

// MediaKey 发送多媒体虚拟键。
func (p *WindowsPlatform) MediaKey(action string) error {
	// VK_MEDIA_PLAY_PAUSE 0xB3, NEXT 0xB0, PREV 0xB1, STOP 0xB2
	// VK_VOLUME_UP 0xAF, DOWN 0xAE
	keyMap := map[string]uintptr{
		"play_pause":  0xB3,
		"next":        0xB0,
		"prev":        0xB1,
		"previous":    0xB1,
		"stop":        0xB2,
		"volume_up":   0xAF,
		"volume_down": 0xAE,
	}
	vk, ok := keyMap[strings.ToLower(action)]
	if !ok {
		return fmt.Errorf("unknown media action: %s", action)
	}

	const keyeventfKeyup = 0x0002
	prockeybdEvent := modUser32.NewProc("keybd_event")
	prockeybdEvent.Call(vk, 0, 0, 0)
	prockeybdEvent.Call(vk, 0, keyeventfKeyup, 0)
	return nil
}

// FlushDNS 清空 DNS 缓存。
func (p *WindowsPlatform) FlushDNS() error {
	cmd := exec.Command("ipconfig", "/flushdns")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("flushdns: %v: %s", err, string(out))
	}
	return nil
}

// SetDNS 设置适配器 DNS。server 为空则设为自动获取。
func (p *WindowsPlatform) SetDNS(adapter, server string) error {
	adapter = strings.TrimSpace(adapter)
	if adapter == "" {
		return fmt.Errorf("adapter name is required")
	}
	server = strings.TrimSpace(server)

	var script string
	if server == "" {
		script = fmt.Sprintf(`Set-DnsClientServerAddress -InterfaceAlias '%s' -ResetServerAddresses`, escapePS(adapter))
	} else {
		script = fmt.Sprintf(`Set-DnsClientServerAddress -InterfaceAlias '%s' -ServerAddresses @('%s')`, escapePS(adapter), escapePS(server))
	}
	_, err := runPowerShell(script, 15)
	return err
}

// EmptyRecycleBin 清空回收站。
func (p *WindowsPlatform) EmptyRecycleBin() error {
	// Shell.Application COM
	script := `
$shell = New-Object -ComObject Shell.Application
$rb = $shell.Namespace(0xA)
if ($rb) { $rb.Items() | ForEach-Object { Remove-Item $_.Path -Recurse -Force -ErrorAction SilentlyContinue } }
`
	_, err := runPowerShell(script, 30)
	return err
}

func escapePS(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// 确保 spi 常量不被未使用告警
var _ = spiFUpdateInifile
var _ = spiFSendWinIniChange
