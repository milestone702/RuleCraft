package output

import (
	"fmt"

	"rulecraft/plugin"
)

// WindowTopmostOutput 设置窗口置顶。
type WindowTopmostOutput struct {
	plat PlatformSubset
}

func NewWindowTopmostOutput(plat PlatformSubset) *WindowTopmostOutput {
	return &WindowTopmostOutput{plat: plat}
}

func (w *WindowTopmostOutput) ID() string        { return "window_topmost" }
func (w *WindowTopmostOutput) Name() string      { return "窗口置顶" }
func (w *WindowTopmostOutput) IsAvailable() bool { return true }

// Execute params: title (必填, 子串匹配), topmost (bool, 默认 true)
func (w *WindowTopmostOutput) Execute(params map[string]interface{}) error {
	title := paramString(params, "title")
	if title == "" {
		return fmt.Errorf("window_topmost: title is required")
	}
	topmost := true
	if v, ok := params["topmost"]; ok {
		switch t := v.(type) {
		case bool:
			topmost = t
		case string:
			topmost = t == "true" || t == "1" || t == "on"
		}
	}
	return w.plat.SetWindowTopmost(title, topmost)
}

// Reset 取消置顶。
func (w *WindowTopmostOutput) Reset(params map[string]interface{}) error {
	title := paramString(params, "title")
	if title == "" {
		return plugin.ErrResetNotSupported
	}
	return w.plat.SetWindowTopmost(title, false)
}
