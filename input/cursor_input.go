package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// CursorInput 采集鼠标光标位置。
type CursorInput struct {
	plat platform.Platform
}

func NewCursorInput(plat platform.Platform) *CursorInput {
	return &CursorInput{plat: plat}
}

func (c *CursorInput) ID() string        { return "cursor" }
func (c *CursorInput) Name() string      { return "光标位置传感器" }
func (c *CursorInput) IsAvailable() bool { return true }

// Collect 写入 cursor.x / cursor.y。
func (c *CursorInput) Collect(ctx *plugin.SystemContext) error {
	x, y, err := c.plat.GetCursorPosition()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("cursor.x", 0)
			ctx.SetState("cursor.y", 0)
			return nil
		}
		return err
	}
	ctx.SetState("cursor.x", x)
	ctx.SetState("cursor.y", y)
	return nil
}
