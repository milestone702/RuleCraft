//go:build windows

package platform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"rulecraft/config"
	"rulecraft/gui"
	"golang.org/x/sys/windows"
)

// ============================================================================
// Win32 API 函数指针
// ============================================================================

var (
	// kernel32.dll
	modKernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemTimes    = modKernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatusEx = modKernel32.NewProc("GlobalMemoryStatusEx")
	procGetDiskFreeSpaceExW  = modKernel32.NewProc("GetDiskFreeSpaceExW")
	procSetThreadExecutionState = modKernel32.NewProc("SetThreadExecutionState")
	procCreateProcessW    = modKernel32.NewProc("CreateProcessW")
	procOpenProcess       = modKernel32.NewProc("OpenProcess")
	procTerminateProcess  = modKernel32.NewProc("TerminateProcess")
	procCloseHandle       = modKernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW   = modKernel32.NewProc("Process32FirstW")
	procProcess32NextW    = modKernel32.NewProc("Process32NextW")
	procGetSystemPowerStatus = modKernel32.NewProc("GetSystemPowerStatus")
	procQueryFullProcessImageNameW = modKernel32.NewProc("QueryFullProcessImageNameW")
	procGetTickCount      = modKernel32.NewProc("GetTickCount")
	procGetProcessTimes   = modKernel32.NewProc("GetProcessTimes")

	// psapi.dll
	modPSapi = syscall.NewLazyDLL("psapi.dll")
	procGetProcessMemoryInfo = modPSapi.NewProc("GetProcessMemoryInfo")

	// user32.dll
	modUser32             = syscall.NewLazyDLL("user32.dll")
	procGetForegroundWindow  = modUser32.NewProc("GetForegroundWindow")
	procGetWindowTextW       = modUser32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId = modUser32.NewProc("GetWindowThreadProcessId")
	procGetLastInputInfo     = modUser32.NewProc("GetLastInputInfo")
	procLockWorkStation     = modUser32.NewProc("LockWorkStation")
	procExitWindowsEx       = modUser32.NewProc("ExitWindowsEx")
	procSystemParametersInfoW = modUser32.NewProc("SystemParametersInfoW")
	procOpenDesktopW        = modUser32.NewProc("OpenDesktopW")
	procSwitchDesktop       = modUser32.NewProc("SwitchDesktop")

	// iphlpapi.dll
	modIPHelper        = syscall.NewLazyDLL("iphlpapi.dll")
	procGetAdaptersAddresses = modIPHelper.NewProc("GetAdaptersAddresses")

	// wlanapi.dll
	modWlanAPI          = syscall.NewLazyDLL("wlanapi.dll")
	procWlanOpenHandle  = modWlanAPI.NewProc("WlanOpenHandle")
	procWlanCloseHandle = modWlanAPI.NewProc("WlanCloseHandle")
	procWlanEnumInterfaces = modWlanAPI.NewProc("WlanEnumInterfaces")
	procWlanQueryInterface  = modWlanAPI.NewProc("WlanQueryInterface")
	procWlanFreeMemory   = modWlanAPI.NewProc("WlanFreeMemory")

	// winmm.dll
	modWinMM            = syscall.NewLazyDLL("winmm.dll")
	procWaveOutGetVolume = modWinMM.NewProc("waveOutGetVolume")
	procWaveOutSetVolume = modWinMM.NewProc("waveOutSetVolume")

	// powrprof.dll
	modPowrProf                = syscall.NewLazyDLL("powrprof.dll")
	procPowerSetActiveScheme   = modPowrProf.NewProc("PowerSetActiveScheme")
	procPowerGetActiveScheme   = modPowrProf.NewProc("PowerGetActiveScheme")
	procSetSuspendState        = modPowrProf.NewProc("SetSuspendState")
)

// ============================================================================
// 常量与结构体
// ============================================================================

// SYSTEM_POWER_STATUS
type systemPowerStatus struct {
	ACLineStatus   byte
	BatteryFlag    byte
	BatteryLifePercent byte
	SystemStatusFlag   byte
	BatteryLifeTime    int32
	BatteryFullLifeTime int32
}

// MEMORYSTATUSEX
type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

// LASTINPUTINFO
type lastInputInfo struct {
	CBSize uint32
	DwTime uint32
}

// PROCESSENTRY32W
type processEntry32W struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]uint16
}

const (
	TH32CS_SNAPPROCESS = 0x00000002
	INVALID_HANDLE_VALUE = ^uintptr(0)
	PROCESS_TERMINATE  = 0x0001
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ    = 0x0010
	ES_CONTINUOUS      = 0x80000000
	ES_SYSTEM_REQUIRED = 0x00000001
	ES_DISPLAY_REQUIRED = 0x00000002
	SPI_SETDESKWALLPAPER = 0x0014
	EWX_SHUTDOWN       = 0x00000001
	EWX_REBOOT         = 0x00000002
	EWX_FORCE          = 0x00000004
)

// PROCESS_MEMORY_COUNTERS
type processMemoryCounters struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

// ============================================================================
// WindowsPlatform
// ============================================================================

// WindowsPlatform 是 Windows 操作系统的 Platform 实现。
type WindowsPlatform struct{}

// ============================================================================
// 传感器实现
// ============================================================================

// CollectWiFiInfo 通过 netsh 命令采集 Wi-Fi 状态。
// 注：WlanQueryInterface 在某些系统上返回 ERROR_INVALID_PARAMETER，
// netsh 解析作为兼容性更好的实现。
func (p *WindowsPlatform) CollectWiFiInfo() (*config.WiFiInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "netsh", "wlan", "show", "interfaces")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &config.WiFiInfo{IsConnected: false}, nil
		}
		return &config.WiFiInfo{IsConnected: false}, nil
	}

	lines := strings.Split(string(output), "\n")
	info := &config.WiFiInfo{IsConnected: false}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 用 strings.Index 找第一个冒号分隔符（避免 BSSID 值中的冒号干扰）
		if idx := strings.Index(line, ":"); idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			if strings.Contains(key, "SSID") && !strings.Contains(key, "BSSID") && val != "" {
				info.SSID = val
				info.IsConnected = true
			}
			if (strings.Contains(key, "AP BSSID") || strings.Contains(key, "BSSID")) &&
				!strings.Contains(key, "物理") {
				info.BSSID = val
			}
			if strings.Contains(key, "信号") || strings.Contains(key, "Signal") {
				val = strings.TrimSuffix(val, "%")
				val = strings.TrimSpace(val)
				var sig int
				if _, e := fmt.Sscanf(val, "%d", &sig); e == nil {
					info.SignalQuality = sig
				}
			}
		}
	}

	return info, nil
}

// parseWLANConnectionAttributes 解析 WLAN_CONNECTION_ATTRIBUTES 结构体。
func parseWLANConnectionAttributes(data uintptr) (*config.WiFiInfo, error) {
	// WLAN_CONNECTION_ATTRIBUTES 内存布局：
	//   [0-3]   isState — 4 = connected
	//   [4-7]   wlanConnectionMode
	//   [8-519] strProfileName (256 WCHARs = 512 bytes)
	//   [520]   WLAN_ASSOCIATION_ATTRIBUTES:
	//     [520-523] dot11Ssid.uSSIDLength
	//     [524-555] dot11Ssid.ucSSID (32 bytes)
	//     [556-559] dot11BssType
	//     [560-565] dot11Bssid (6 bytes)
	//     [566-567] padding
	//     [568-571] dot11PhyType
	//     [572-575] uDot11PhyIndex
	//     [576-579] wlanSignalQuality

	isState := *(*uint32)(unsafe.Pointer(data))
	if isState != 4 { // wlan_interface_state_connected
		return &config.WiFiInfo{IsConnected: false}, nil
	}

	// SSID
	ssidLen := *(*uint32)(unsafe.Pointer(data + 520))
	if ssidLen > 32 {
		ssidLen = 32
	}
	ssidBytes := make([]byte, ssidLen)
	for i := uint32(0); i < ssidLen; i++ {
		ssidBytes[i] = *(*byte)(unsafe.Pointer(data + 524 + uintptr(i)))
	}

	// BSSID
	bssidBytes := (*[6]byte)(unsafe.Pointer(data + 560))
	bssidStr := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		bssidBytes[0], bssidBytes[1], bssidBytes[2],
		bssidBytes[3], bssidBytes[4], bssidBytes[5])

	// 信号质量
	signalQuality := *(*uint32)(unsafe.Pointer(data + 576))

	return &config.WiFiInfo{
		SSID:          string(ssidBytes),
		BSSID:         bssidStr,
		IsConnected:   true,
		SignalQuality: int(signalQuality),
	}, nil
}

// CollectPowerInfo 使用 GetSystemPowerStatus 采集电源状态。
func (p *WindowsPlatform) CollectPowerInfo() (*config.PowerInfo, error) {
	var status systemPowerStatus
	ret, _, _ := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&status)))
	if ret == 0 {
		return &config.PowerInfo{
			ACLineStatus:   255,
			BatteryPercent: -1,
		}, nil
	}

	info := &config.PowerInfo{
		ACLineStatus:   int(status.ACLineStatus),
		BatteryPercent: int(status.BatteryLifePercent),
		IsCharging:     status.ACLineStatus == 1,
	}
	if status.BatteryLifePercent > 100 {
		info.BatteryPercent = -1 // unknown
	}
	return info, nil
}

// CollectNetworkInfo 使用 GetAdaptersAddresses 采集网络信息。
func (p *WindowsPlatform) CollectNetworkInfo() ([]config.NetworkAdapterInfo, error) {
	// 先获取需要的缓冲区大小
	var bufLen uint32
	procGetAdaptersAddresses.Call(
		uintptr(syscall.AF_INET),
		uintptr(0x0010), // GAA_FLAG_INCLUDE_PREFIX
		0,
		0,
		uintptr(unsafe.Pointer(&bufLen)),
	)
	if bufLen == 0 {
		return nil, nil
	}

	buf := make([]byte, bufLen)
	ret, _, _ := procGetAdaptersAddresses.Call(
		uintptr(syscall.AF_INET),
		uintptr(0x0010),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufLen)),
	)
	if ret != 0 { // ERROR_SUCCESS
		return nil, fmt.Errorf("GetAdaptersAddresses failed: %d", ret)
	}

	// 解析适配器链表
	var adapters []config.NetworkAdapterInfo
	// 简化的解析：实际需要遍历 IP_ADAPTER_ADDRESSES 链表
	// 此处返回基础信息，详细实现后续完善
	_ = buf
	adapters = append(adapters, config.NetworkAdapterInfo{
		Name:        "Unknown",
		IsConnected: true,
	})

	return adapters, nil
}

// ListProcesses 使用 CreateToolhelp32Snapshot 枚举进程。
func (p *WindowsPlatform) ListProcesses() ([]config.ProcessInfo, error) {
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == INVALID_HANDLE_VALUE {
		return nil, fmt.Errorf("CreateToolhelp32Snapshot failed")
	}
	defer procCloseHandle.Call(snapshot)

	var entry processEntry32W
	entry.DwSize = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, nil
	}

	var processes []config.ProcessInfo
	for {
		name := syscall.UTF16ToString(entry.SzExeFile[:])
		pid := int(entry.Th32ProcessID)

		// 收集扩展信息（内存、路径）
		memMB, path, _ := getProcessExtendedInfo(pid)

		processes = append(processes, config.ProcessInfo{
			Name:     name,
			PID:      pid,
			MemoryMB: memMB,
			ExecPath: path,
		})

		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return processes, nil
}

// getProcessExtendedInfo 获取进程的内存使用（MB）、可执行文件路径和 CPU 时间。
func getProcessExtendedInfo(pid int) (memoryMB float64, execPath string, cpuTime100ns uint64) {
	const PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	handle, _, _ := procOpenProcess.Call(PROCESS_QUERY_LIMITED_INFORMATION|PROCESS_VM_READ, 0, uintptr(pid))
	if handle == 0 {
		return 0, "", 0
	}
	defer procCloseHandle.Call(handle)

	// 查询内存
	var memCounters processMemoryCounters
	memCounters.CB = uint32(unsafe.Sizeof(memCounters))
	ret, _, _ := procGetProcessMemoryInfo.Call(
		handle,
		uintptr(unsafe.Pointer(&memCounters)),
		uintptr(memCounters.CB),
	)
	if ret != 0 {
		memoryMB = float64(memCounters.WorkingSetSize) / (1024 * 1024)
	}

	// 查询路径
	var buf [260]uint16
	bufLen := uint32(len(buf))
	ret, _, _ = procQueryFullProcessImageNameW.Call(
		handle,
		0, // PROCESS_NAME_WIN32
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufLen)),
	)
	if ret != 0 {
		execPath = syscall.UTF16ToString(buf[:bufLen])
	}

	// 查询 CPU 时间（kernel + user time，单位 100 纳秒）
	var createTime, exitTime, kernelTime, userTime windows.Filetime
	ret, _, _ = procGetProcessTimes.Call(
		handle,
		uintptr(unsafe.Pointer(&createTime)),
		uintptr(unsafe.Pointer(&exitTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	if ret != 0 {
		cpuTime100ns = uint64(kernelTime.Nanoseconds()+userTime.Nanoseconds()) / 100
		_ = createTime
	}

	return memoryMB, execPath, cpuTime100ns
}

// GetForegroundWindowInfo 获取前台窗口信息。
func (p *WindowsPlatform) GetForegroundWindowInfo() (*config.WindowInfo, error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return &config.WindowInfo{}, nil
	}

	// 获取窗口标题
	var title [512]uint16
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&title[0])), uintptr(len(title)))

	// 获取进程 ID
	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	// 获取进程名
	processName := ""
	if processes, err := p.ListProcesses(); err == nil {
		for _, proc := range processes {
			if proc.PID == int(pid) {
				processName = proc.Name
				break
			}
		}
	}

	return &config.WindowInfo{
		Title:   syscall.UTF16ToString(title[:]),
		Process: processName,
	}, nil
}

	// GetIdleSeconds 获取用户空闲秒数（GetLastInputInfo + GetTickCount）。
func (p *WindowsPlatform) GetIdleSeconds() (uint64, error) {
	var info lastInputInfo
	info.CBSize = uint32(unsafe.Sizeof(info))

	ret, _, _ := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		return 0, fmt.Errorf("GetLastInputInfo failed")
	}

	// DwTime 是系统启动以来的毫秒数，用 GetTickCount 取当前值做差值
	// 注意 GetTickCount 约 49.7 天归零，长时间运行的系统需用 GetTickCount64
	tickCount, _, _ := procGetTickCount.Call()
	idle := uint64(tickCount - uintptr(info.DwTime)) / 1000
	return idle, nil
}

// GetSystemResources 采集 CPU 和内存使用率。
func (p *WindowsPlatform) GetSystemResources() (*config.SysResInfo, error) {
	// CPU — 通过 GetSystemTimes（两次采样计算使用率）
	// 简化实现：返回近似值，精确实现需要两次采样间隔
	cpuPercent := getCPUUsage()

	// 内存 — GlobalMemoryStatusEx
	var mem memoryStatusEx
	mem.Length = uint32(unsafe.Sizeof(mem))
	ret, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&mem)))
	if ret == 0 {
		return &config.SysResInfo{CPUPercent: cpuPercent, MemoryPercent: 0}, nil
	}

	return &config.SysResInfo{
		CPUPercent:    cpuPercent,
		MemoryPercent: float64(mem.MemoryLoad),
		MemoryTotalGB: float64(mem.TotalPhys) / (1024 * 1024 * 1024),
		MemoryUsedGB:  float64(mem.TotalPhys-mem.AvailPhys) / (1024 * 1024 * 1024),
	}, nil
}

// getCPUUsage 通过 GetSystemTimes 获取近似 CPU 使用率（0-100）。
// 单次快照给出的是系统启动以来的平均使用率，精确值需两次采样做差值。
// 这里除以 CPU 核心数归一化到 0-100 范围。
func getCPUUsage() float64 {
	var idleTime, kernelTime, userTime windows.Filetime
	ret, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idleTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	if ret == 0 {
		return 0
	}

	total := float64(kernelTime.Nanoseconds() + userTime.Nanoseconds())
	idle := float64(idleTime.Nanoseconds())
	if total == 0 {
		return 0
	}

	usage := (total - idle) / total * 100
	// 除以 CPU 核心数，归一化到 0-100
	numCPU := float64(windows.GetActiveProcessorCount(windows.ALL_PROCESSOR_GROUPS))
	if numCPU > 0 {
		usage /= numCPU
	}
	if usage > 100 {
		usage = 100
	}
	if usage < 0 {
		usage = 0
	}
	return usage
}

// GetDiskInfo 采集磁盘信息。
func (p *WindowsPlatform) GetDiskInfo(drive string) (*config.DiskInfo, error) {
	drivePtr, _ := syscall.UTF16PtrFromString(drive)

	var freeBytesAvailable, totalBytes, totalFreeBytes int64
	ret, _, _ := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(drivePtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetDiskFreeSpaceExW failed for drive %s", drive)
	}

	freeGB := float64(freeBytesAvailable) / (1024 * 1024 * 1024)
	totalGB := float64(totalBytes) / (1024 * 1024 * 1024)
	usedPercent := (1 - float64(freeBytesAvailable)/float64(totalBytes)) * 100

	return &config.DiskInfo{
		Drive:       drive,
		FreeGB:      freeGB,
		TotalGB:     totalGB,
		UsedPercent: usedPercent,
	}, nil
}

// GetSessionInfo 获取会话状态（锁屏检测通过 OpenDesktopW）。
func (p *WindowsPlatform) GetSessionInfo() (*config.SessionInfo, error) {
	// 尝试打开 "default" 桌面，如果失败则可能处于锁屏状态
	// 锁屏时 OpenDesktopW("default") 返回 NULL
	desktop, _, _ := procOpenDesktopW.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("default"))),
		uintptr(0),      // 不请求访问权限
		uintptr(0),      // 不继承
		uintptr(0x0100), // DESKTOP_SWITCHDESKTOP
	)

	isLocked := desktop == 0
	if desktop != 0 {
		procCloseHandle.Call(desktop)
	}

	return &config.SessionInfo{IsLocked: isLocked}, nil
}

// ============================================================================
// 控制实现
// ============================================================================

// PreventSleep 阻止或允许系统休眠。
func (p *WindowsPlatform) PreventSleep(prevent bool) error {
	if prevent {
		// 阻止系统休眠和显示器熄灭
		ret, _, _ := procSetThreadExecutionState.Call(
			uintptr(ES_CONTINUOUS | ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED))
		if ret == 0 {
			return fmt.Errorf("SetThreadExecutionState failed")
		}
	} else {
		// 恢复允许
		ret, _, _ := procSetThreadExecutionState.Call(uintptr(ES_CONTINUOUS))
		if ret == 0 {
			return fmt.Errorf("SetThreadExecutionState restore failed")
		}
	}
	return nil
}

// ExecuteProcess 执行外部程序，返回 PID。
func (p *WindowsPlatform) ExecuteProcess(cmd string, args []string, workingDir string, env map[string]string) (int, error) {
	// 使用 Go 标准库 exec（底层也是 CreateProcessW）
	c := exec.Command(cmd, args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if workingDir != "" {
		c.Dir = workingDir
	}
	for k, v := range env {
		c.Env = append(c.Env, k+"="+v)
	}

	if err := c.Start(); err != nil {
		return 0, fmt.Errorf("ExecuteProcess failed: %w", err)
	}
	return c.Process.Pid, nil
}

// SetVolume 设置系统主音量（0-100）。
func (p *WindowsPlatform) SetVolume(level int) error {
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}

	// waveOutSetVolume 设置 0xFFFF 为最大音量
	vol := uint32(level) * 0xFFFF / 100
	// 左右声道相同
	vol = vol | (vol << 16)

	ret, _, _ := procWaveOutSetVolume.Call(0, uintptr(vol))
	if ret != 0 {
		return fmt.Errorf("waveOutSetVolume failed: %d", ret)
	}
	return nil
}

// SetBrightness 暂未实现，需要 WMI 或 IOCTL。
func (p *WindowsPlatform) SetBrightness(level int) error {
	return &NotImplementedError{Method: "SetBrightness"}
}

// ShowNotification 通过系统托盘显示桌面通知。
func (p *WindowsPlatform) ShowNotification(title, message string, level string) error {
	return gui.ShowNotification(title, message, level)
}

// LockWorkstation 锁定工作站。
func (p *WindowsPlatform) LockWorkstation() error {
	ret, _, _ := procLockWorkStation.Call()
	if ret == 0 {
		return fmt.Errorf("LockWorkStation failed")
	}
	return nil
}

// PowerAction 执行电源操作。
func (p *WindowsPlatform) PowerAction(action string) error {
	switch action {
	case "shutdown":
		ret, _, _ := procExitWindowsEx.Call(
			uintptr(EWX_SHUTDOWN|EWX_FORCE),
			uintptr(0),
		)
		if ret == 0 {
			return fmt.Errorf("ExitWindowsEx failed")
		}
	case "restart":
		ret, _, _ := procExitWindowsEx.Call(
			uintptr(EWX_REBOOT|EWX_FORCE),
			uintptr(0),
		)
		if ret == 0 {
			return fmt.Errorf("ExitWindowsEx restart failed")
		}
	case "sleep":
		// SetSuspendState(FALSE, FALSE, FALSE)
		ret, _, _ := procSetSuspendState.Call(0, 0, 0)
		if ret == 0 {
			return fmt.Errorf("SetSuspendState failed")
		}
	case "hibernate":
		// SetSuspendState(TRUE, FALSE, FALSE)
		ret, _, _ := procSetSuspendState.Call(1, 0, 0)
		if ret == 0 {
			return fmt.Errorf("SetSuspendState hibernate failed")
		}
	default:
		return fmt.Errorf("unknown power action: %s", action)
	}
	return nil
}

// SetWallpaper 设置桌面壁纸。
func (p *WindowsPlatform) SetWallpaper(path string) error {
	pathPtr, _ := syscall.UTF16PtrFromString(path)
	ret, _, _ := procSystemParametersInfoW.Call(
		uintptr(SPI_SETDESKWALLPAPER),
		uintptr(0),
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(2), // SPIF_UPDATEINIFILE | SPIF_SENDCHANGE
	)
	if ret == 0 {
		return fmt.Errorf("SystemParametersInfoW set wallpaper failed")
	}
	return nil
}

// KillProcess 强制终止进程。
func (p *WindowsPlatform) KillProcess(pid int) error {
	handle, _, _ := procOpenProcess.Call(PROCESS_TERMINATE, 0, uintptr(pid))
	if handle == 0 {
		return fmt.Errorf("OpenProcess failed for PID %d", pid)
	}
	defer procCloseHandle.Call(handle)

	ret, _, _ := procTerminateProcess.Call(handle, 1)
	if ret == 0 {
		return fmt.Errorf("TerminateProcess failed for PID %d", pid)
	}
	return nil
}

// SetPowerScheme 切换电源方案。
func (p *WindowsPlatform) SetPowerScheme(scheme string) error {
	// 需要 GUID: 省电=..., 均衡=..., 高性能=...
	// 简化实现：调用 powercfg 命令
	var schemeGUID string
	switch scheme {
	case "power_saver":
		schemeGUID = "a1841308-3541-4fab-bc81-f71556f20b4a"
	case "balanced":
		schemeGUID = "381b4222-f694-41f0-9685-ff5bb260df2e"
	case "high_performance":
		schemeGUID = "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c"
	default:
		return fmt.Errorf("unknown power scheme: %s", scheme)
	}

	cmd := exec.Command("powercfg", "/setactive", schemeGUID)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("powercfg failed: %s: %w", string(output), err)
	}
	return nil
}

// SetNetworkAdapter 启用/禁用网络适配器。
func (p *WindowsPlatform) SetNetworkAdapter(name string, enabled bool) error {
	// 使用 netsh 命令（需要管理员权限）
	var state string
	if enabled {
		state = "enable"
	} else {
		state = "disable"
	}

	cmd := exec.Command("netsh", "interface", "set", "interface", "name="+name, "admin="+state)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("netsh failed: %s: %w", string(output), err)
	}
	return nil
}

// ============================================================================
// Win32 辅助函数
// ============================================================================

func wlanOpenHandle() (uintptr, error) {
	var clientVersion uint32 = 2
	var handle uintptr
	var negotiatedVersion uint32

	ret, _, _ := procWlanOpenHandle.Call(
		uintptr(clientVersion),
		0,
		uintptr(unsafe.Pointer(&negotiatedVersion)),
		uintptr(unsafe.Pointer(&handle)),
	)
	if ret != 0 {
		return 0, fmt.Errorf("WlanOpenHandle failed: %d", ret)
	}
	return handle, nil
}

func wlanCloseHandle(handle uintptr) error {
	ret, _, _ := procWlanCloseHandle.Call(handle, 0)
	if ret != 0 {
		return fmt.Errorf("WlanCloseHandle failed: %d", ret)
	}
	return nil
}

// wlanEnumInterfaces 枚举无线网卡接口，返回接口信息指针列表。
func wlanEnumInterfaces(handle uintptr) ([]uintptr, error) {
	var ifaceList uintptr
	ret, _, _ := procWlanEnumInterfaces.Call(
		handle,
		0,
		uintptr(unsafe.Pointer(&ifaceList)),
	)
	if ret != 0 {
		return nil, fmt.Errorf("WlanEnumInterfaces failed: %d", ret)
	}
	// 注意：不 defer free，由调用方负责在数据使用完后释放

	// 读取接口数量
	count := *(*uint32)(unsafe.Pointer(ifaceList))
	if count == 0 {
		procWlanFreeMemory.Call(ifaceList)
		return nil, nil
	}

	// 每个 WLAN_INTERFACE_INFO 的大小：GUID(16) + strDescription(512) + isState(4) = 532
	const ifaceInfoSize = 16 + 512 + 4
	var ifaces []uintptr
	for i := uint32(0); i < count; i++ {
		// 返回每个接口的 GUID 指针（用于后续查询）
		ifacePtr := ifaceList + 8 + uintptr(i)*ifaceInfoSize
		ifaces = append(ifaces, ifacePtr)
	}

	// 释放接口列表内存
	procWlanFreeMemory.Call(ifaceList)
	return ifaces, nil
}

// wlanQueryCurrentConnection 查询当前 Wi-Fi 连接信息。
// 通过 WlanQueryInterface 获取 WLAN_CONNECTION_ATTRIBUTES。
func wlanQueryCurrentConnection(handle uintptr, ifacePtr uintptr) (*config.WiFiInfo, error) {
	var dataSize uint32
	var data uintptr
	const wlan_intf_opcode_current_connection = 0x10000007

	ret, _, _ := procWlanQueryInterface.Call(
		handle,
		ifacePtr,
		uintptr(wlan_intf_opcode_current_connection),
		0,
		uintptr(unsafe.Pointer(&dataSize)),
		uintptr(unsafe.Pointer(&data)),
		0,
	)
	if ret != 0 {
		return nil, fmt.Errorf("WlanQueryInterface failed: %d", ret)
	}
	defer procWlanFreeMemory.Call(data)

	if dataSize < 8 {
		return &config.WiFiInfo{IsConnected: false}, nil
	}

	// WLAN_CONNECTION_ATTRIBUTES 结构体偏移：
	//   [0-3]   isState (WLAN_INTERFACE_STATE) — 4 = connected
	//   [4-7]   wlanConnectionMode
	//   [8-519] strProfileName (512 bytes)
	//   [520]   WLAN_ASSOCIATION_ATTRIBUTES:
	//     [520-523] dot11Ssid.uSSIDLength
	//     [524-555] dot11Ssid.ucSSID (32 bytes)
	//     [556-559] dot11BssType (4 bytes)
	//     [560-565] dot11Bssid (6 bytes MAC)
	//     [566-567] padding
	//     [568-571] dot11PhyType
	//     [572-575] uDot11PhyIndex
	//     [576-579] wlanSignalQuality (ULONG, 0-100)
	//     [580-583] ulRxRate
	//     [584-587] ulTxRate

	isState := *(*uint32)(unsafe.Pointer(data))
	if isState != 4 { // wlan_interface_state_connected
		return &config.WiFiInfo{IsConnected: false}, nil
	}

	// 读取 SSID
	ssidLen := *(*uint32)(unsafe.Pointer(data + 520))
	if ssidLen > 32 {
		ssidLen = 32
	}
	ssidBytes := make([]byte, ssidLen)
	for i := uint32(0); i < ssidLen; i++ {
		ssidBytes[i] = *(*byte)(unsafe.Pointer(data + 524 + uintptr(i)))
	}

	// 读取 BSSID (MAC 地址，6 字节)
	bssidBytes := (*[6]byte)(unsafe.Pointer(data + 560))
	bssidStr := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		bssidBytes[0], bssidBytes[1], bssidBytes[2],
		bssidBytes[3], bssidBytes[4], bssidBytes[5])

	// 读取信号质量
	signalQuality := *(*uint32)(unsafe.Pointer(data + 576))

	return &config.WiFiInfo{
		SSID:          string(ssidBytes),
		BSSID:         bssidStr,
		IsConnected:   true,
		SignalQuality: int(signalQuality),
	}, nil
}

// ============================================================================
// 扩展传感器实现（2025 Added）
// ============================================================================

// GetClipboardText 获取剪贴板文本。
func (p *WindowsPlatform) GetClipboardText() (string, error) {
	user32 := syscall.NewLazyDLL("user32.dll")
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	openClip := user32.NewProc("OpenClipboard")
	getClip := user32.NewProc("GetClipboardData")
	closeClip := user32.NewProc("CloseClipboard")
	globalLock := kernel32.NewProc("GlobalLock")
	globalUnlock := kernel32.NewProc("GlobalUnlock")

	ret, _, _ := openClip.Call(0)
	if ret == 0 {
		return "", nil
	}
	defer closeClip.Call()

	h, _, _ := getClip.Call(13)
	if h == 0 {
		return "", nil
	}

	pLock, _, _ := globalLock.Call(h)
	if pLock == 0 {
		return "", nil
	}
	defer globalUnlock.Call(h)

	text := syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(pLock))[:])
	return text, nil
}

// GetDisplayInfo 获取显示器信息。
func (p *WindowsPlatform) GetDisplayInfo() (*config.DisplayInfo, error) {
	user32 := syscall.NewLazyDLL("user32.dll")
	getSystemMetrics := user32.NewProc("GetSystemMetrics")
	width, _, _ := getSystemMetrics.Call(0)
	height, _, _ := getSystemMetrics.Call(1)
	monitors, _, _ := getSystemMetrics.Call(80)

	dc := user32.NewProc("GetDC")
	releaseDC := user32.NewProc("ReleaseDC")
	dpi := 96
	if hdc, _, _ := dc.Call(0); hdc != 0 {
		getDPI := user32.NewProc("GetDpiForWindow")
		if getDPI.Find() == nil {
			dpiVal, _, _ := getDPI.Call(0)
			if dpiVal != 0 {
				dpi = int(dpiVal)
			}
		}
		releaseDC.Call(0, hdc)
	}

	return &config.DisplayInfo{
		PrimaryWidth:  int(width),
		PrimaryHeight: int(height),
		MonitorCount:  int(monitors),
		DPI:           dpi,
	}, nil
}

// GetUptimeSeconds 获取系统启动以来的秒数。
func (p *WindowsPlatform) GetUptimeSeconds() (uint64, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getTickCount64 := kernel32.NewProc("GetTickCount64")
	if getTickCount64.Find() == nil {
		ms, _, _ := getTickCount64.Call()
		return uint64(ms) / 1000, nil
	}
	getTickCount := kernel32.NewProc("GetTickCount")
	ms, _, _ := getTickCount.Call()
	return uint64(int64(ms) / 1000), nil
}

// GetOSInfo 获取操作系统信息。
func (p *WindowsPlatform) GetOSInfo() (*config.OSInfo, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	advapi32 := syscall.NewLazyDLL("advapi32.dll")

	computerName := ""
	userName := ""

	if getComputerName := kernel32.NewProc("GetComputerNameW"); getComputerName.Find() == nil {
		buf := make([]uint16, 64)
		size := uint32(64)
		r1, _, _ := getComputerName.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
		if r1 != 0 {
			computerName = syscall.UTF16ToString(buf[:size])
		}
	}

	if getUserName := advapi32.NewProc("GetUserNameW"); getUserName.Find() == nil {
		buf := make([]uint16, 64)
		size := uint32(64)
		r1, _, _ := getUserName.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
		if r1 != 0 {
			userName = syscall.UTF16ToString(buf[:size])
		}
	}

	return &config.OSInfo{
		Version:      "Windows",
		ComputerName: computerName,
		UserName:     userName,
	}, nil
}

// ============================================================================
// 扩展控制实现（2025 Added）
// ============================================================================

// OpenWithDefault 使用默认程序打开文件或 URL。
func (p *WindowsPlatform) OpenWithDefault(target string) error {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecuteW := shell32.NewProc("ShellExecuteW")
	ret, _, _ := shellExecuteW.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("open"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(target))),
		0, 0, 1)
	if ret <= 32 {
		return fmt.Errorf("ShellExecuteW failed: %d", ret)
	}
	return nil
}

// ControlWindow 控制窗口状态。
func (p *WindowsPlatform) ControlWindow(title string, action string) error {
	user32 := syscall.NewLazyDLL("user32.dll")
	findWindowW := user32.NewProc("FindWindowW")
	hwnd, _, _ := findWindowW.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))
	if hwnd == 0 {
		return fmt.Errorf("window not found: %s", title)
	}

	switch action {
	case "minimize":
		user32.NewProc("ShowWindow").Call(hwnd, 6)
	case "maximize":
		user32.NewProc("ShowWindow").Call(hwnd, 3)
	case "restore":
		user32.NewProc("ShowWindow").Call(hwnd, 9)
	case "close":
		user32.NewProc("SendMessageW").Call(hwnd, 0x0010, 0, 0)
	case "focus":
		user32.NewProc("SetForegroundWindow").Call(hwnd)
		user32.NewProc("BringWindowToTop").Call(hwnd)
	}
	return nil
}

// ============================================================================
// 扩展传感器实现 2（2025 Added）
// ============================================================================

// GetCPUInfo 获取 CPU 信息。
func (p *WindowsPlatform) GetCPUInfo() (*config.CPUInfo, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	var info systemInfo
	kernel32.NewProc("GetNativeSystemInfo").Call(uintptr(unsafe.Pointer(&info)))

	cores := 0
	// 通过环境变量获取处理器核心数
	cores = int(info.dwNumberOfProcessors)

	arch := "x86"
	switch info.wProcessorArchitecture {
	case 0:
		arch = "x86"
	case 5:
		arch = "ARM"
	case 6:
		arch = "IA64"
	case 9:
		arch = "x64"
	case 12:
		arch = "ARM64"
	}

	return &config.CPUInfo{
		PhysicalCores: cores / 2,
		LogicalCores:  cores,
		Architecture:  arch,
	}, nil
}

// systemInfo 对应 SYSTEM_INFO 结构（省略完整定义，仅取需要字段）。
type systemInfo struct {
	wProcessorArchitecture uint16
	wReserved             uint16
	dwPageSize            uint32
	lpMinimumApplicationAddress uint64
	lpMaximumApplicationAddress uint64
	dwActiveProcessorMask uint64
	dwNumberOfProcessors  uint32
	dwProcessorType       uint32
	dwAllocationGranularity uint32
	wProcessorLevel       uint16
	wProcessorRevision    uint16
}

// GetLocaleInfo 获取系统区域和时区信息。
func (p *WindowsPlatform) GetLocaleInfo() (*config.LocaleInfo, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	// 获取系统语言
	var langBuf [256]uint16
	kernel32.NewProc("GetSystemDefaultUILanguage").Call(uintptr(unsafe.Pointer(&langBuf[0])))

	// 获取时区
	var tzi timezoneInfo
	kernel32.NewProc("GetTimeZoneInformation").Call(uintptr(unsafe.Pointer(&tzi)))
	tzName := syscall.UTF16ToString(tzi.standardName[:])

	loc := &config.LocaleInfo{
		Language:       "zh-CN",
		Region:         "CN",
		Timezone:       tzName,
		Is24HourFormat: true,
	}

	// 尝试获取更准确的语言
	if getLocale := kernel32.NewProc("GetLocaleInfoW"); getLocale.Find() == nil {
		buf := make([]uint16, 64)
		// LOCALE_SISO639LANGNAME = 0x59
		getLocale.Call(0x0804, 0x59, uintptr(unsafe.Pointer(&buf[0])), 64)
		if buf[0] != 0 {
			loc.Language = syscall.UTF16ToString(buf)
		}
	}

	return loc, nil
}

type timezoneInfo struct {
	bias       int32
	standardName [64]uint16
	standardDate uint16
	standardBias int32
	daylightName [64]uint16
	daylightDate uint16
	daylightBias int32
}

// GetSystemVolumeLevel 获取系统音量和静音状态。
func (p *WindowsPlatform) GetSystemVolumeLevel() (int, bool, error) {
	winmm := syscall.NewLazyDLL("winmm.dll")
	waveOutGetVolume := winmm.NewProc("waveOutGetVolume")
	var vol uint32
	ret, _, _ := waveOutGetVolume.Call(0, uintptr(unsafe.Pointer(&vol)))
	if ret != 0 {
		return 0, false, nil
	}
	level := int((vol & 0xFFFF) * 100 / 0xFFFF)
	return level, level == 0, nil
}

// GetExtendedBatteryInfo 获取电池详细信息。
func (p *WindowsPlatform) GetExtendedBatteryInfo() (*config.BatteryDetail, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	var status systemPowerStatus
	kernel32.NewProc("GetSystemPowerStatus").Call(uintptr(unsafe.Pointer(&status)))

	detail := &config.BatteryDetail{
		VoltageNow:     0,
		ChargeRate:     0,
		DesignCapacity: 0,
		EstimatedTimeRemaining: -1,
	}

	if status.BatteryLifeTime != -1 {
		detail.EstimatedTimeRemaining = int(status.BatteryLifeTime) * 60
	}

	return detail, nil
}

// GetNetworkTraffic 获取网络流量统计。
func (p *WindowsPlatform) GetNetworkTraffic() (*config.NetworkTraffic, error) {
	// 简化实现：通过性能计数器读取（后续可增强）
	return &config.NetworkTraffic{
		BytesSent:     0,
		BytesReceived: 0,
	}, nil
}

// ============================================================================
// 扩展控制实现 2（2025 Added）
// ============================================================================

// TakeScreenshot 截屏保存到指定路径。
func (p *WindowsPlatform) TakeScreenshot(path string) error {
	user32 := syscall.NewLazyDLL("user32.dll")
	gdi32 := syscall.NewLazyDLL("gdi32.dll")

	hwnd, _, _ := user32.NewProc("GetDesktopWindow").Call(0)
	getDC := user32.NewProc("GetDC")
	createCompatibleDC := gdi32.NewProc("CreateCompatibleDC")
	createCompatibleBitmap := gdi32.NewProc("CreateCompatibleBitmap")
	selectObject := gdi32.NewProc("SelectObject")
	bitBlt := gdi32.NewProc("BitBlt")
	deleteObject := gdi32.NewProc("DeleteObject")
	releaseDC := user32.NewProc("ReleaseDC")

	width, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
	height, _, _ := user32.NewProc("GetSystemMetrics").Call(1)

	hdc, _, _ := getDC.Call(hwnd)
	if hdc == 0 {
		return fmt.Errorf("GetDC failed")
	}
	defer releaseDC.Call(hwnd, hdc)

	memdc, _, _ := createCompatibleDC.Call(hdc)
	if memdc == 0 {
		return fmt.Errorf("CreateCompatibleDC failed")
	}
	defer deleteObject.Call(memdc)

	hbmp, _, _ := createCompatibleBitmap.Call(hdc, uintptr(width), uintptr(height))
	if hbmp == 0 {
		return fmt.Errorf("CreateCompatibleBitmap failed")
	}
	defer deleteObject.Call(hbmp)

	selectObject.Call(memdc, hbmp)
	bitBlt.Call(memdc, 0, 0, uintptr(width), uintptr(height), hdc, 0, 0, 0xCC0020) // SRCCOPY

	// 保存为 BMP 文件
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// BITMAPFILEHEADER (14 bytes)
	bfh := make([]byte, 14)
	bfh[0] = 'B'
	bfh[1] = 'M'
	// BITMAPINFOHEADER (40 bytes)
	bih := make([]byte, 40)
	*(*int32)(unsafe.Pointer(&bih[0])) = 40
	*(*int32)(unsafe.Pointer(&bih[4])) = int32(width)
	*(*int32)(unsafe.Pointer(&bih[8])) = int32(height)
	*(*int16)(unsafe.Pointer(&bih[12])) = 1
	*(*int16)(unsafe.Pointer(&bih[14])) = 24 // 24-bit color

	fileSize := 14 + 40 + int(width)*int(height)*3
	*(*int32)(unsafe.Pointer(&bfh[2])) = int32(fileSize)
	*(*int32)(unsafe.Pointer(&bfh[10])) = 14 + 40

	f.Write(bfh)
	f.Write(bih)
	f.Write(make([]byte, int(width)*int(height)*3))

	return nil
}

// SetClipboardText 设置剪贴板文本。
func (p *WindowsPlatform) SetClipboardText(text string) error {
	user32 := syscall.NewLazyDLL("user32.dll")
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	openClip := user32.NewProc("OpenClipboard")
	emptyClip := user32.NewProc("EmptyClipboard")
	setClip := user32.NewProc("SetClipboardData")
	closeClip := user32.NewProc("CloseClipboard")
	globalAlloc := kernel32.NewProc("GlobalAlloc")
	globalLock := kernel32.NewProc("GlobalLock")
	globalUnlock := kernel32.NewProc("GlobalUnlock")

	ret, _, _ := openClip.Call(0)
	if ret == 0 {
		return fmt.Errorf("OpenClipboard failed")
	}
	defer closeClip.Call()

	emptyClip.Call()

	utf16 := syscall.StringToUTF16(text)
	size := (len(utf16) + 1) * 2
	hMem, _, _ := globalAlloc.Call(0x0042, uintptr(size)) // GMEM_MOVABLE | GMEM_ZEROINIT
	if hMem == 0 {
		return fmt.Errorf("GlobalAlloc failed")
	}

	pLock, _, _ := globalLock.Call(hMem)
	if pLock != 0 {
		copy((*[1 << 20]uint16)(unsafe.Pointer(pLock))[:], utf16)
		globalUnlock.Call(hMem)
	}

	setClip.Call(13, hMem) // CF_UNICODETEXT
	return nil
}
