// Package output 包含所有内置输出插件（Actuators）。
//
// 每个输出插件执行一种系统控制动作。由引擎在条件触发时调用 Execute()，
// 条件恢复时调用 Reset()（如果支持）。
//
// 添加新输出插件：
//   1. 在 output/ 下新建单独 .go 文件
//   2. 实现 plugin.OutputPlugin 接口
//   3. 在 RegisterBuiltinOutputs 中添加一行
package output

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// PlatformSubset 定义输出插件需要的平台接口子集。
type PlatformSubset interface {
	SetPowerScheme(scheme string) error
	PreventSleep(prevent bool) error
	SetVolume(level int) error
	SetBrightness(level int) error
	LockWorkstation() error
	PowerAction(action string) error
	SetWallpaper(path string) error
	KillProcess(pid int) error
	ExecuteProcess(cmd string, args []string, workingDir string, env map[string]string) (int, error)
	SetNetworkAdapter(name string, enabled bool) error
	ShowNotification(title, message string, level string) error
	OpenWithDefault(target string) error
	ControlWindow(title string, action string) error
	TakeScreenshot(path string) error
	SetClipboardText(text string) error
	SetMute(muted bool) error
	ToggleMute() (bool, error)
	GetSystemVolumeLevel() (int, bool, error)
	RestartExplorer() error
	OpenControlPanelPage(page string) error
	KillProcessByName(name string) error
	MediaKey(action string) error
	FlushDNS() error
	SetDNS(adapter, server string) error
	SetThemeMode(mode string) error
	EmptyRecycleBin() error
	SetWindowTopmost(title string, topmost bool) error
	SetProcessPriority(name, priority string) error
}

// Ensure Platform implements PlatformSubset at compile time.
var _ PlatformSubset = (platform.Platform)(nil)

// RegisterBuiltinOutputs 注册所有内置输出插件到注册表。
func RegisterBuiltinOutputs(registry *plugin.Registry, plat PlatformSubset) error {
	type reg struct {
		p   plugin.OutputPlugin
		doc string
	}
	outputs := []reg{
		{NewNotifyOutput(plat), "显示桌面通知"},
		{NewPowerSchemeOutput(plat), "切换电源方案（节能/平衡/高性能）"},
		{NewPreventSleepOutput(plat), "阻止或允许系统休眠"},
		{NewExecOutput(plat), "执行外部程序或脚本"},
		{NewVolumeOutput(plat), "设置系统音量"},
		{NewBrightnessOutput(plat), "调节屏幕亮度"},
		{NewLockOutput(plat), "锁定工作站"},
		{NewPowerActionOutput(plat), "关机/重启/睡眠/休眠"},
		{NewWallpaperOutput(plat), "更换桌面壁纸"},
		{NewKillProcessOutput(plat), "按 PID 终止进程"},
		{NewWebhookOutput(), "发送 Webhook HTTP 请求"},
		{NewHTTPGetOutput(), "发起 HTTP GET 请求"},
		{NewHTTPPostOutput(), "发起 HTTP POST 请求"},
		{NewDelayOutput(), "延时等待"},
		{NewFileOutput(), "写入文件内容"},
		{NewNetAdapterOutput(plat), "启用/禁用网络适配器"},
		{NewOpenOutput(plat), "用默认程序打开文件或 URL"},
		{NewWindowControlOutput(plat), "最小化/最大化/关闭/聚焦窗口"},
		{NewScreenshotOutput(plat), "截屏保存"},
		{NewClipboardSetOutput(plat), "写入剪贴板文本"},
		{NewMuteOutput(plat), "静音 / 取消静音 / 切换静音"},
		{NewRestartExplorerOutput(plat), "重启 Windows 资源管理器"},
		{NewVolumeFadeOutput(plat), "音量渐变到目标值"},
		{NewControlPanelOutput(plat), "打开控制面板 / 设置页"},
		{NewKillProcessNameOutput(plat), "按进程名终止全部匹配进程"},
		{NewMediaControlOutput(plat), "媒体播放/切歌/音量键"},
		{NewFlushDNSOutput(plat), "清空 DNS 解析缓存"},
		{NewDarkModeOutput(plat), "切换深色/浅色主题"},
		{NewSetDNSOutput(plat), "设置网卡 DNS 服务器"},
		{NewEmptyRecycleBinOutput(plat), "清空回收站"},
		{NewWindowTopmostOutput(plat), "窗口置顶/取消置顶"},
		{NewProcessPriorityOutput(plat), "设置进程优先级"},
		// 应用对接
		{NewDiscordWebhookOutput(), "Discord Webhook 推送"},
		{NewBarkOutput(), "Bark iOS 推送"},
		{NewWeComBotOutput(), "企业微信群机器人"},
		{NewDingTalkBotOutput(), "钉钉群机器人"},
		{NewFeishuBotOutput(), "飞书群机器人"},
		{NewServerChanOutput(), "Server酱 Turbo"},
		{NewTelegramBotOutput(), "Telegram Bot 消息"},
		{NewSlackWebhookOutput(), "Slack Incoming Webhook"},
		{NewHomeAssistantOutput(), "Home Assistant REST"},
	}

	for _, item := range outputs {
		if err := registry.RegisterOutput(item.p, item.doc); err != nil {
			return err
		}
	}
	return nil
}
