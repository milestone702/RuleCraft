package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// WindowInput 采集前台窗口信息。
type WindowInput struct {
	plat platform.Platform
}

// NewWindowInput 创建窗口输入插件。
func NewWindowInput(plat platform.Platform) *WindowInput {
	return &WindowInput{plat: plat}
}

func (w *WindowInput) ID() string        { return "window" }
func (w *WindowInput) Name() string      { return "窗口传感器" }
func (w *WindowInput) IsAvailable() bool { return true }

// Collect 采集前台窗口信息并写入全局状态。
func (w *WindowInput) Collect(ctx *plugin.SystemContext) error {
	info, err := w.plat.GetForegroundWindowInfo()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("window.foreground.title", "")
			ctx.SetState("window.foreground.process", "")
			return nil
		}
		return err
	}

	ctx.SetState("window.foreground.title", info.Title)
	ctx.SetState("window.foreground.process", info.Process)
	return nil
}
