// Package tray 管理系统托盘图标、上下文菜单和桌面通知。
//
// 当前实现：
//   - NoopTrayManager（无操作桩实现）— 不依赖任何外部库，在所有平台上编译
//
// 将来启用 systray 的步骤（需要网络环境）：
//   1. go get github.com/getlantern/systray
//   2. 实现 SystrayTrayManager（参考接口定义）
//   3. 在 main.go 中将 NewTrayManager 切换到 SystrayTrayManager
package tray

// ============================================================================
// 通知级别常量
// ============================================================================

// NotificationLevel 通知级别。
type NotificationLevel int

const (
	NotifyNone    NotificationLevel = 0
	NotifyInfo    NotificationLevel = 1
	NotifyWarning NotificationLevel = 2
	NotifyError   NotificationLevel = 3
)

// ============================================================================
// TrayManager 接口
// ============================================================================

// TrayManager 管理系统托盘的生命周期、菜单和通知。
type TrayManager interface {
	// Run 启动系统托盘（阻塞调用，应在独立 goroutine 中执行）。
	Run() error

	// Stop 优雅停止系统托盘。
	Stop()

	// ShowNotification 显示桌面气球通知。
	ShowNotification(title, message string, level NotificationLevel) error

	// SetMenuItems 设置上下文菜单项。
	SetMenuItems(items []MenuItem)

	// IsAvailable 返回当前平台是否支持系统托盘。
	IsAvailable() bool
}

// MenuItem 定义托盘上下文菜单的一个条目。
type MenuItem struct {
	ID      string
	Label   string
	OnClick func()
}

// ============================================================================
// 菜单项 ID 常量
// ============================================================================

const (
	MenuOpenWeb   = "open_web"
	MenuSeparator = "separator"
	MenuQuit      = "quit"
)

// ============================================================================
// NoopTrayManager（无操作桩实现）
// ============================================================================

// NoopTrayManager 是一个占位实现，所有方法均为无操作。
// 当系统托盘库不可用时作为安全降级使用。
type NoopTrayManager struct{}

// NewNoopTrayManager 创建无操作托盘管理器。
func NewNoopTrayManager() *NoopTrayManager {
	return &NoopTrayManager{}
}

func (n *NoopTrayManager) Run() error {
	// NoopTrayManager 的 Run() 不会阻塞，立即返回。
	return nil
}

func (n *NoopTrayManager) Stop() {}

func (n *NoopTrayManager) ShowNotification(title, message string, level NotificationLevel) error {
	// 在无系统托盘时，通知降级为控制台输出。
	logPrefix := "[notify]"
	switch level {
	case NotifyError:
		logPrefix = "[notify:error]"
	case NotifyWarning:
		logPrefix = "[notify:warning]"
	}
	println(logPrefix, title, "—", message)
	return nil
}

func (n *NoopTrayManager) SetMenuItems(items []MenuItem) {
	// 无操作
	_ = items
}

func (n *NoopTrayManager) IsAvailable() bool { return false }

// ============================================================================
// 工厂函数
// ============================================================================

// NewTrayManager 创建当前平台可用的托盘管理器。
// 当前返回 NoopTrayManager。启用 systray 后切换为此函数的真实实现。
func NewTrayManager() TrayManager {
	return NewNoopTrayManager()
}
