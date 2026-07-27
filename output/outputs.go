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
	"rulecraft/plugin"
	"rulecraft/platform"
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
}

// Ensure Platform implements PlatformSubset at compile time.
var _ PlatformSubset = (platform.Platform)(nil)

// RegisterBuiltinOutputs 注册所有内置输出插件到注册表。
func RegisterBuiltinOutputs(registry *plugin.Registry, plat PlatformSubset) error {
	outputs := []plugin.OutputPlugin{
		NewNotifyOutput(plat),
		NewPowerSchemeOutput(plat),
		NewPreventSleepOutput(plat),
		NewExecOutput(plat),
		NewVolumeOutput(plat),
		NewBrightnessOutput(plat),
		NewLockOutput(plat),
		NewPowerActionOutput(plat),
		NewWallpaperOutput(plat),
		NewKillProcessOutput(plat),
		NewWebhookOutput(),
		NewHTTPGetOutput(),
		NewHTTPPostOutput(),
		NewDelayOutput(),
		NewFileOutput(),
		NewNetAdapterOutput(plat),
		NewOpenOutput(plat),
		NewWindowControlOutput(plat),
		NewScreenshotOutput(plat),
		NewClipboardSetOutput(plat),
	}

	for _, p := range outputs {
		if err := registry.RegisterOutput(p, ""); err != nil {
			return err
		}
	}
	return nil
}
