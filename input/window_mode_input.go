package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// WindowModeInput 采集前台窗口是否全屏。
type WindowModeInput struct {
	plat platform.Platform
}

func NewWindowModeInput(plat platform.Platform) *WindowModeInput {
	return &WindowModeInput{plat: plat}
}

func (w *WindowModeInput) ID() string        { return "window_mode" }
func (w *WindowModeInput) Name() string      { return "窗口模式传感器" }
func (w *WindowModeInput) IsAvailable() bool { return true }

// Collect 写入 window_mode.is_fullscreen。
func (w *WindowModeInput) Collect(ctx *plugin.SystemContext) error {
	fs, err := w.plat.IsForegroundFullscreen()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("window_mode.is_fullscreen", false)
			return nil
		}
		return err
	}
	ctx.SetState("window_mode.is_fullscreen", fs)
	return nil
}
