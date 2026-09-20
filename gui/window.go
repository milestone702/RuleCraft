//go:build windows

// Package gui 提供 RuleCraft 的 Windows 系统托盘实现（无窗口）。
//
// 架构：
//   - 隐藏的消息窗口用于接收托盘通知
//   - 右键菜单包含：开机自启（勾选）/修改端口/Web启停/打开Web/退出
//   - 修改端口弹出模态对话框
package gui

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/png"
	"log"
	"strconv"
	"syscall"
	"unsafe"
)

//go:embed logo.png
var logoPNG []byte

// ============================================================================
// Win32 DLL 和函数指针
// ============================================================================

var (
	modUser32   = syscall.NewLazyDLL("user32.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
	modGdi32    = syscall.NewLazyDLL("gdi32.dll")
	modShell32  = syscall.NewLazyDLL("shell32.dll")

	procRegisterClassExW       = modUser32.NewProc("RegisterClassExW")
	procCreateWindowExW        = modUser32.NewProc("CreateWindowExW")
	procDefWindowProcW         = modUser32.NewProc("DefWindowProcW")
	procGetMessageW            = modUser32.NewProc("GetMessageW")
	procTranslateMessage       = modUser32.NewProc("TranslateMessage")
	procDispatchMessageW       = modUser32.NewProc("DispatchMessageW")
	procPostQuitMessage        = modUser32.NewProc("PostQuitMessage")
	procDestroyWindow          = modUser32.NewProc("DestroyWindow")
	procShowWindow             = modUser32.NewProc("ShowWindow")
	procSetWindowLongPtrW      = modUser32.NewProc("SetWindowLongPtrW")
	procGetWindowLongPtrW      = modUser32.NewProc("GetWindowLongPtrW")
	procGetDlgItem             = modUser32.NewProc("GetDlgItem")
	procSetWindowTextW         = modUser32.NewProc("SetWindowTextW")
	procGetWindowTextW         = modUser32.NewProc("GetWindowTextW")
	procSendMessageW           = modUser32.NewProc("SendMessageW")
	procLoadIconW              = modUser32.NewProc("LoadIconW")
	procLoadCursorW            = modUser32.NewProc("LoadCursorW")
	procCreatePopupMenu        = modUser32.NewProc("CreatePopupMenu")
	procAppendMenuW            = modUser32.NewProc("AppendMenuW")
	procTrackPopupMenu         = modUser32.NewProc("TrackPopupMenu")
	procDestroyMenu            = modUser32.NewProc("DestroyMenu")
	procSetForegroundWindow    = modUser32.NewProc("SetForegroundWindow")
	procShellNotifyIconW       = modShell32.NewProc("Shell_NotifyIconW")
	procGetModuleHandleW       = modKernel32.NewProc("GetModuleHandleW")
	procCreateDialog            = modUser32.NewProc("DialogBoxParamW")

	// GDI32
	modGDI32                   = syscall.NewLazyDLL("gdi32.dll")
	procCreateDIBSection       = modGDI32.NewProc("CreateDIBSection")
	procCreateBitmap           = modGDI32.NewProc("CreateBitmap")
	procDeleteObject           = modGDI32.NewProc("DeleteObject")

	// User32 icon
	procCreateIconIndirect     = modUser32.NewProc("CreateIconIndirect")
	procDestroyIcon            = modUser32.NewProc("DestroyIcon")
	// EndDialog
	procEndDialog              = modUser32.NewProc("EndDialog")
	// PostMessage
	procPostMessageW           = modUser32.NewProc("PostMessageW")
)

// ============================================================================
// Win32 常量和结构体
// ============================================================================

const (
	WM_DESTROY     = 0x0002
	WM_CLOSE       = 0x0010
	WM_COMMAND     = 0x0111
	WM_CREATE      = 0x0001
	WM_INITDIALOG  = 0x0110
	WM_QUIT        = 0x0012
	WM_NULL        = 0x0000

	WM_TRAYNOTIFY = 0x8000 + 100

	// 托盘
	NIM_ADD        = 0
	NIM_MODIFY     = 1
	NIM_DELETE     = 2
	NIF_MESSAGE    = 0x00000001
	NIF_ICON       = 0x00000002
	NIF_TIP        = 0x00000004
	NIF_INFO       = 0x00000010
	NIIF_INFO      = 0x00000001
	NIIF_WARNING   = 0x00000002
	NIIF_ERROR     = 0x00000003

	// 弹出菜单
	TPM_RIGHTBUTTON = 0x0002
	TPM_BOTTOMALIGN = 0x0020

	MF_STRING       = 0x00000000
	MF_CHECKED      = 0x00000008
	MF_UNCHECKED    = 0x00000000
	MF_SEPARATOR    = 0x00000800

	// 控件 ID
	ID_TRAY_AUTOSTART = 2001
	ID_TRAY_PORT      = 2002
	ID_TRAY_TOGGLE    = 2003
	ID_TRAY_OPENWEB   = 2004
	ID_TRAY_EXIT      = 2005

	// 端口对话框控件 ID
	IDD_PORT_EDIT   = 3001
	IDD_PORT_OK     = 3002
	IDD_PORT_CANCEL = 3003

	WS_CAPTION     = 0x00C00000
	WS_SYSMENU     = 0x00080000
	WS_VISIBLE     = 0x10000000
	WS_POPUP       = 0x80000000
	WS_OVERLAPPED  = 0x00000000
	WS_CHILD       = 0x40000000
	WS_TABSTOP     = 0x00010000
	WS_GROUP       = 0x00020000
	WS_BORDER      = 0x00800000
	ES_LEFT        = 0x00000000
	ES_NUMBER      = 0x00002000
	BS_PUSHBUTTON  = 0x00000000
	BS_AUTOCHECKBOX = 0x00000003
	WS_EX_TOOLWINDOW = 0x00000080
	WS_EX_DLGMODALFRAME = 0x00000001

	SW_HIDE        = 0
	SW_SHOW        = 5
	SW_RESTORE     = 9

	// 消息框
	MB_OK          = 0x00000000
	MB_ICONERROR   = 0x00000010
	MB_ICONINFORMATION = 0x00000040
	MB_OKCANCEL    = 0x00000001
	IDOK           = 1
	IDCANCEL       = 2
)

type wndClassExW struct {
	CbSize        uint32
	Style         uint32
	WndProc       uintptr
	ClsExtra      int32
	WndExtra      int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	PtX     int32
	PtY     int32
}

type point struct {
	X, Y int32
}

type notifyIconDataW struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         syscall.GUID
}

// ============================================================================
// Windows 图标结构体
// ============================================================================

type iconInfo struct {
	FIcon    uint32
	XHotspot uint32
	YHotspot uint32
	HbmMask  uintptr
	HbmColor uintptr
}

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors [1]uint32
}

var appIcon uintptr

// createAppIcon 从嵌入的 PNG 创建 Windows HICON。
func createAppIcon() uintptr {
	if appIcon != 0 {
		return appIcon
	}
	img, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		log.Printf("[gui] decode PNG failed: %v", err)
		return 0
	}
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	stride := w * 4
	pixels := make([]byte, stride*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			dstY := h - 1 - y
			off := dstY*stride + x*4
			pixels[off+0] = byte(b >> 8)
			pixels[off+1] = byte(g >> 8)
			pixels[off+2] = byte(r >> 8)
			pixels[off+3] = byte(a >> 8)
		}
	}
	bmpHeader := bitmapInfo{
		Header: bitmapInfoHeader{
			Size:    uint32(unsafe.Sizeof(bitmapInfoHeader{})),
			Width:   int32(w),
			Height:  int32(h),
			Planes:  1,
			BitCount: 32,
		},
	}
	var hbmColor uintptr
	_, _, _ = procCreateDIBSection.Call(0, uintptr(unsafe.Pointer(&bmpHeader)), 0, uintptr(unsafe.Pointer(&hbmColor)), 0, 0)
	if hbmColor != 0 {
		procDeleteObject.Call(hbmColor)
	}
	hbmColor, _, _ = procCreateBitmap.Call(uintptr(w), uintptr(h), 1, 32, 0)
	if hbmColor == 0 {
		log.Printf("[gui] CreateBitmap failed")
		return 0
	}
	procGetDC := modUser32.NewProc("GetDC")
	dc, _, _ := procGetDC.Call(0)
	if dc != 0 {
		procSetDIBits := modGDI32.NewProc("SetDIBits")
		procSetDIBits.Call(dc, hbmColor, 0, uintptr(h), uintptr(unsafe.Pointer(&pixels[0])), uintptr(unsafe.Pointer(&bmpHeader)), 0)
		procReleaseDC := modUser32.NewProc("ReleaseDC")
		procReleaseDC.Call(0, dc)
	}
	maskSize := ((w + 31) / 32) * 4 * h
	maskData := make([]byte, maskSize)
	hbmMask, _, _ := procCreateBitmap.Call(uintptr(w), uintptr(h), 1, 1, uintptr(unsafe.Pointer(&maskData[0])))
	if hbmMask == 0 {
		procDeleteObject.Call(hbmColor)
		return 0
	}
	info := iconInfo{FIcon: 1, HbmMask: hbmMask, HbmColor: hbmColor}
	hicon, _, _ := procCreateIconIndirect.Call(uintptr(unsafe.Pointer(&info)))
	if hicon == 0 {
		procDeleteObject.Call(hbmColor)
		procDeleteObject.Call(hbmMask)
		return 0
	}
	appIcon = hicon
	// CreateIconIndirect 已复制位图数据，可安全释放原始位图
	procDeleteObject.Call(hbmColor)
	procDeleteObject.Call(hbmMask)
	return hicon
}

// ============================================================================
// GUI 全局状态
// ============================================================================

var (
	hiddenHWnd    uintptr // 隐藏消息窗口句柄
	trayUID       uint32  = 100
	hInst         uintptr
	portValue     int    = 19530
	webEnabled    bool   = true
	autoStart     bool   = false
	portDlgHwnd   uintptr // 端口对话框句柄（非模态用）

	onOpenWeb      func()
	onToggleWeb    func() bool
	onExit         func()
	onAutoStart    func(bool)
	onPortChange   func(int)
)

// ============================================================================
// 公开接口
// ============================================================================

type Config struct {
	Port         int
	AutoStart    bool
	WebEnabled   bool
	OnOpenWeb    func()
	OnToggleWeb  func() bool
	OnExit       func()
	OnAutoStart  func(bool)
	OnPortChange func(int)
}

// Run 启动纯托盘模式（阻塞，直到用户退出）。
func Run(cfg Config) {
	portValue = cfg.Port
	autoStart = cfg.AutoStart
	webEnabled = cfg.WebEnabled
	onOpenWeb = cfg.OnOpenWeb
	onToggleWeb = cfg.OnToggleWeb
	onExit = cfg.OnExit
	onAutoStart = cfg.OnAutoStart
	onPortChange = cfg.OnPortChange

	hInst = getModuleHandle()

	// 注册隐藏消息窗口类
	className := syscall.StringToUTF16Ptr("RuleCraftTrayWindow")
	wc := wndClassExW{
		CbSize:        uint32(unsafe.Sizeof(wndClassExW{})),
		WndProc:       syscall.NewCallback(trayWindowProc),
		HInstance:     hInst,
		HCursor:       loadStandardCursor(),
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// 创建隐藏的顶层窗口。
	// 注意：不要用 HWND_MESSAGE（message-only）——
	// SetForegroundWindow / TrackPopupMenu 对 message-only 窗口常失败，
	// 会导致托盘右键菜单无响应。
	hwnd, _, _ := procCreateWindowExW.Call(
		uintptr(WS_EX_TOOLWINDOW),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("RuleCraftTray"))),
		0, // 无 WS_VISIBLE，不显示在任务栏/桌面上
		0, 0, 0, 0,
		0, // 普通顶层窗口，而非 HWND_MESSAGE
		0, hInst, 0,
	)
	if hwnd == 0 {
		log.Fatalf("[gui] create hidden window failed")
	}
	hiddenHWnd = hwnd

	// 创建托盘图标
	if err := createTrayIcon(hwnd); err != nil {
		log.Printf("[gui] create tray failed: %v", err)
	}

	// 消息循环
	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 { // WM_QUIT
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	deleteTrayIcon(hwnd)
	procDestroyWindow.Call(hwnd)
}

// PostQuit 退出消息循环。
func PostQuit() {
	procPostQuitMessage.Call(0)
}

// ============================================================================
// 隐藏窗口过程（接收托盘消息）
// ============================================================================

func trayWindowProc(hwnd uintptr, msgID uint32, wParam, lParam uintptr) uintptr {
	switch msgID {
	case WM_CREATE:
		return 0
	case WM_COMMAND:
		cmdID := wParam & 0xFFFF
		if cmdID >= ID_TRAY_AUTOSTART && cmdID <= ID_TRAY_EXIT {
			handleTrayCommand(hwnd, cmdID)
		}
		return 0
	case WM_TRAYNOTIFY:
		// 经典回调（未 NIM_SETVERSION）：lParam 为鼠标消息。
		// 0x0205 = WM_RBUTTONUP, 0x007B = WM_CONTEXTMENU
		// 不要同时处理 WM_RBUTTONDOWN(0x0204)，否则可能弹出两次菜单。
		switch lParam {
		case 0x0205, 0x007B:
			showTrayMenu(hwnd)
		case 0x0000, 0x0203: // 左键单击/双击 → 打开 Web
			if onOpenWeb != nil {
				onOpenWeb()
			}
		}
		return 0
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msgID), wParam, lParam)
	return ret
}

// ============================================================================
// 系统托盘
// ============================================================================

func createTrayIcon(hwnd uintptr) error {
	icon := createAppIcon()
	if icon == 0 {
		icon, _, _ = procLoadIconW.Call(0, 32512)
	}
	nid := notifyIconDataW{
		CbSize:           uint32(unsafe.Sizeof(notifyIconDataW{})),
		HWnd:             hwnd,
		UID:              trayUID,
		UFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
		UCallbackMessage: WM_TRAYNOTIFY,
		HIcon:            icon,
	}
	copy(nid.SzTip[:], syscall.StringToUTF16("RuleCraft 自动化引擎"))
	ret, _, _ := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return fmt.Errorf("Shell_NotifyIconW NIM_ADD failed")
	}
	return nil
}

func deleteTrayIcon(hwnd uintptr) {
	nid := notifyIconDataW{
		CbSize: uint32(unsafe.Sizeof(notifyIconDataW{})),
		HWnd:   hwnd,
		UID:    trayUID,
	}
	procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}

// ShowNotification 通过系统托盘显示气泡通知。
// 必须在 Run() 之后调用（hiddenHWnd 非零时生效）。
func ShowNotification(title, message string, level string) error {
	if hiddenHWnd == 0 {
		return fmt.Errorf("tray not initialized")
	}
	// 截断过长的内容（Shell_NotifyIconW 有长度限制）
	titleRunes := []rune(title)
	if len(titleRunes) > 63 {
		titleRunes = titleRunes[:63]
	}
	msgRunes := []rune(message)
	if len(msgRunes) > 255 {
		msgRunes = msgRunes[:255]
	}

	// 根据级别设置通知图标
	var infoFlags uint32 = 1 // NIIF_INFO
	switch level {
	case "warn", "warning":
		infoFlags = 2 // NIIF_WARNING
	case "error":
		infoFlags = 3 // NIIF_ERROR
	}

	nid := notifyIconDataW{
		CbSize:      uint32(unsafe.Sizeof(notifyIconDataW{})),
		HWnd:        hiddenHWnd,
		UID:         trayUID,
		UFlags:      NIF_INFO,
		DwInfoFlags: infoFlags,
	}
	copy(nid.SzInfoTitle[:], syscall.StringToUTF16(string(titleRunes)))
	copy(nid.SzInfo[:], syscall.StringToUTF16(string(msgRunes)))

	ret, _, _ := procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return fmt.Errorf("Shell_NotifyIconW NIM_MODIFY failed")
	}
	return nil
}

// ============================================================================
// 托盘右键菜单
// ============================================================================

func showTrayMenu(hwnd uintptr) {
	menu, _, _ := procCreatePopupMenu.Call()

	// 1. 开机自动启动（勾选）
	flags := uintptr(MF_STRING)
	if autoStart {
		flags |= MF_CHECKED
	}
	procAppendMenuW.Call(menu, flags, ID_TRAY_AUTOSTART,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("开机自动启动"))))

	// 2. 修改端口
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_PORT,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(fmt.Sprintf("修改端口 (当前: %d)", portValue)))))

	// 3. 开启/停止 Web 服务
	toggleText := "启动 Web 服务"
	if webEnabled {
		toggleText = "停止 Web 服务"
	}
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_TOGGLE,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(toggleText))))

	// 4. 打开 Web 界面
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_OPENWEB,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("打开 Web 界面"))))

	// 5. 分隔线 + 退出
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_EXIT,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("退出程序"))))

	// 获取鼠标位置
	var cursorPos point
	procGetCursorPos := modUser32.NewProc("GetCursorPos")
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&cursorPos)))

	// MSDN 托盘菜单标准写法：
	// 1) 先 SetForegroundWindow，否则 TrackPopupMenu 可能不弹或立刻关闭
	// 2) TrackPopupMenu 结束后再 PostMessage(WM_NULL)，修复菜单无法消失/焦点异常
	procSetForegroundWindow.Call(hwnd)
	procTrackPopupMenu.Call(menu, TPM_RIGHTBUTTON|TPM_BOTTOMALIGN,
		uintptr(uint32(cursorPos.X)), uintptr(uint32(cursorPos.Y)), 0, hwnd, 0)
	procDestroyMenu.Call(menu)
	procPostMessageW.Call(hwnd, WM_NULL, 0, 0)
}

func handleTrayCommand(hwnd uintptr, cmdID uintptr) {
	switch cmdID {
	case ID_TRAY_AUTOSTART:
		autoStart = !autoStart
		if onAutoStart != nil {
			onAutoStart(autoStart)
		}
		// 刷新菜单（更新勾选状态）
		// 下次右键自动显示新状态

	case ID_TRAY_PORT:
		showPortDialog(hwnd)

	case ID_TRAY_TOGGLE:
		if onToggleWeb != nil {
			webEnabled = onToggleWeb()
		}

	case ID_TRAY_OPENWEB:
		if onOpenWeb != nil {
			onOpenWeb()
		}

	case ID_TRAY_EXIT:
		if onExit != nil {
			onExit()
		}
		procPostQuitMessage.Call(0)
	}
}

// ============================================================================
// 端口修改对话框
// ============================================================================

// showPortDialog 创建非模态端口输入对话框。
func showPortDialog(parentHWnd uintptr) {
	// 防止重复打开
	if portDlgHwnd != 0 {
		procSetForegroundWindow.Call(portDlgHwnd)
		return
	}

	dialogW := 280
	dialogH := 110

	// 注册对话框窗口类（只注册一次）
	dlgClass := syscall.StringToUTF16Ptr("RuleCraftPortDialog")
	wc2 := wndClassExW{
		CbSize:        uint32(unsafe.Sizeof(wndClassExW{})),
		Style:         0,
		WndProc:       syscall.NewCallback(portDialogProc),
		ClsExtra:      0,
		WndExtra:      0,
		HInstance:     hInst,
		HIcon:         0,
		HCursor:       loadStandardCursor(),
		HbrBackground: 1 + 5, // COLOR_WINDOW+1
		LpszMenuName:  nil,
		LpszClassName: dlgClass,
		HIconSm:       0,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc2)))

	// 计算居中位置
	screenW := int32(GetSystemMetrics(0)) // SM_CXSCREEN
	screenH := int32(GetSystemMetrics(1)) // SM_CYSCREEN
	x := (int(screenW) - dialogW) / 2
	y := (int(screenH) - dialogH) / 2

	style := uintptr(WS_POPUP | WS_CAPTION | WS_SYSMENU | WS_VISIBLE)
	hwnd, _, _ := procCreateWindowExW.Call(
		WS_EX_DLGMODALFRAME,
		uintptr(unsafe.Pointer(dlgClass)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("修改端口"))),
		style,
		uintptr(x), uintptr(y), uintptr(dialogW), uintptr(dialogH),
		parentHWnd,
		0, hInst, 0,
	)
	if hwnd == 0 {
		return
	}

	// 保存对话框句柄（非模态，不进入嵌套消息循环）
	portDlgHwnd = hwnd
	// 注意：对话框控件在 WM_CREATE 中创建，消息由主消息循环正常分发
}

func isWindow(hwnd uintptr) bool {
	procIsWindow := modUser32.NewProc("IsWindow")
	ret, _, _ := procIsWindow.Call(hwnd)
	return ret != 0
}

func isChildWindow(parent, hwnd uintptr) bool {
	procIsChild := modUser32.NewProc("IsChild")
	ret, _, _ := procIsChild.Call(parent, hwnd)
	return ret != 0
}

func GetSystemMetrics(index int) int32 {
	procGetSystemMetrics := modUser32.NewProc("GetSystemMetrics")
	ret, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int32(ret)
}

// portDialogProc 端口对话框的窗口过程。
func portDialogProc(hwnd uintptr, msgID uint32, wParam, lParam uintptr) uintptr {
	switch msgID {
	case WM_CREATE:
		createPortDialogControls(hwnd)
		return 0

	case WM_COMMAND:
		cmdID := wParam & 0xFFFF
		switch cmdID {
		case IDD_PORT_OK:
			// 读取端口号
			editHWnd, _, _ := procGetDlgItem.Call(hwnd, IDD_PORT_EDIT)
			if editHWnd != 0 {
				var buf [16]uint16
				procGetWindowTextW.Call(editHWnd, uintptr(unsafe.Pointer(&buf[0])), 16)
				portStr := syscall.UTF16ToString(buf[:])
				p, err := strconv.Atoi(portStr)
				if err == nil && p > 0 && p < 65536 {
					portValue = p
					if onPortChange != nil {
						onPortChange(p)
					}
					procDestroyWindow.Call(hwnd)
				} else {
					// 显示错误
					procMsgBox := modUser32.NewProc("MessageBoxW")
					procMsgBox.Call(hwnd,
						uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("端口号必须是 1-65535 之间的整数"))),
						uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("输入无效"))),
						MB_OK|MB_ICONERROR)
				}
			}
			return 0

		case IDD_PORT_CANCEL, IDCANCEL:
			portDlgHwnd = 0
			procDestroyWindow.Call(hwnd)
			return 0
		}
		return 0

	case WM_CLOSE:
		portDlgHwnd = 0
		procDestroyWindow.Call(hwnd)
		return 0

	case WM_DESTROY:
		portDlgHwnd = 0
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msgID), wParam, lParam)
	return ret
}

// createPortDialogControls 创建端口对话框的控件（确定和取消在同一行）。
func createPortDialogControls(hwnd uintptr) {
	// 标签
	createStatic(hwnd, "请输入新的端口号：", 20, 20, 240, 20, 0)
	// 端口输入框
	portStr := fmt.Sprintf("%d", portValue)
	createEdit(hwnd, portStr, 20, 48, 120, 24, IDD_PORT_EDIT)
	// 确定按钮
	createButton(hwnd, "确定", 155, 48, 50, 28, IDD_PORT_OK)
	// 取消按钮（与确定同一行）
	createButton(hwnd, "取消", 210, 48, 50, 28, IDD_PORT_CANCEL)
}

// ============================================================================
// 辅助控件创建
// ============================================================================

func createButton(hwnd uintptr, text string, x, y, w, h, id int) uintptr {
	textPtr := syscall.StringToUTF16Ptr(text)
	btn, _, _ := procCreateWindowExW.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("BUTTON"))),
		uintptr(unsafe.Pointer(textPtr)),
		WS_CHILD|WS_VISIBLE|uintptr(BS_PUSHBUTTON)|WS_TABSTOP,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		hwnd, uintptr(id), hInst, 0)
	return btn
}

func createEdit(hwnd uintptr, text string, x, y, w, h, id int) uintptr {
	textPtr := syscall.StringToUTF16Ptr(text)
	edit, _, _ := procCreateWindowExW.Call(0x00000200, // WS_EX_CLIENTEDGE
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("EDIT"))),
		uintptr(unsafe.Pointer(textPtr)),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|uintptr(ES_LEFT)|uintptr(WS_BORDER)|uintptr(ES_NUMBER),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		hwnd, uintptr(id), hInst, 0)
	return edit
}

func createStatic(hwnd uintptr, text string, x, y, w, h, id int) uintptr {
	textPtr := syscall.StringToUTF16Ptr(text)
	static, _, _ := procCreateWindowExW.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("STATIC"))),
		uintptr(unsafe.Pointer(textPtr)),
		WS_CHILD|WS_VISIBLE,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		hwnd, uintptr(id), hInst, 0)
	return static
}

// ============================================================================
// 辅助函数
// ============================================================================

func loadStandardCursor() uintptr {
	cursor, _, _ := procLoadCursorW.Call(0, 32512) // IDC_ARROW
	return cursor
}

func getModuleHandle() uintptr {
	h, _, _ := procGetModuleHandleW.Call(0)
	return h
}

// SetVersion 版本号占位。
func SetVersion(v string) {}

// ShowMainWindow 保持兼容（纯托盘模式此函数无操作）。
func ShowMainWindow() {}
