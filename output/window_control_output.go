package output

import (
	"fmt"

	"rulecraft/plugin"
)

// WindowControlOutput 窗口控制（最小化/最大化/关闭/聚焦）。
type WindowControlOutput struct {
	plat PlatformSubset
}

func NewWindowControlOutput(plat PlatformSubset) *WindowControlOutput {
	return &WindowControlOutput{plat: plat}
}

func (w *WindowControlOutput) ID() string        { return "window_control" }
func (w *WindowControlOutput) Name() string      { return "窗口控制" }
func (w *WindowControlOutput) IsAvailable() bool { return true }

func (w *WindowControlOutput) Execute(params map[string]interface{}) error {
	title, _ := params["title"].(string)
	action, _ := params["action"].(string)
	if title == "" || action == "" {
		return fmt.Errorf("window_control: title and action are required")
	}
	return w.plat.ControlWindow(title, action)
}

func (w *WindowControlOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
