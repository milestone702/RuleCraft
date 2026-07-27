package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// ClipboardInput 剪贴板传感器。
type ClipboardInput struct {
	plat platform.Platform
}

func NewClipboardInput(plat platform.Platform) *ClipboardInput {
	return &ClipboardInput{plat: plat}
}

func (c *ClipboardInput) ID() string        { return "clipboard" }
func (c *ClipboardInput) Name() string      { return "剪贴板传感器" }
func (c *ClipboardInput) IsAvailable() bool { return true }

func (c *ClipboardInput) Collect(ctx *plugin.SystemContext) error {
	text, err := c.plat.GetClipboardText()
	if err != nil {
		ctx.SetState("clipboard.text", "")
		ctx.SetState("clipboard.has_text", false)
		return nil
	}
	ctx.SetState("clipboard.text", text)
	ctx.SetState("clipboard.has_text", text != "")
	return nil
}
