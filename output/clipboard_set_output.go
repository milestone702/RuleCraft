package output

import (
	"rulecraft/plugin"
)

// ClipboardSetOutput 设置剪贴板输出插件。
type ClipboardSetOutput struct {
	plat PlatformSubset
}

func NewClipboardSetOutput(plat PlatformSubset) *ClipboardSetOutput {
	return &ClipboardSetOutput{plat: plat}
}

func (c *ClipboardSetOutput) ID() string        { return "clipboard_set" }
func (c *ClipboardSetOutput) Name() string      { return "设置剪贴板" }
func (c *ClipboardSetOutput) IsAvailable() bool { return true }

func (c *ClipboardSetOutput) Execute(params map[string]interface{}) error {
	text, _ := params["text"].(string)
	return c.plat.SetClipboardText(text)
}

func (c *ClipboardSetOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
