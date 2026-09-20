package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// ThemeInput 采集系统深色/浅色主题。
type ThemeInput struct {
	plat platform.Platform
}

func NewThemeInput(plat platform.Platform) *ThemeInput {
	return &ThemeInput{plat: plat}
}

func (t *ThemeInput) ID() string        { return "theme" }
func (t *ThemeInput) Name() string      { return "主题传感器" }
func (t *ThemeInput) IsAvailable() bool { return true }

// Collect 写入 theme.mode / theme.is_dark。
func (t *ThemeInput) Collect(ctx *plugin.SystemContext) error {
	mode, err := t.plat.GetThemeMode()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("theme.mode", "light")
			ctx.SetState("theme.is_dark", false)
			return nil
		}
		return err
	}
	ctx.SetState("theme.mode", mode)
	ctx.SetState("theme.is_dark", mode == "dark")
	return nil
}
