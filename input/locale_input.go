package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// LocaleInput 区域语言传感器。
type LocaleInput struct {
	plat platform.Platform
}

func NewLocaleInput(plat platform.Platform) *LocaleInput {
	return &LocaleInput{plat: plat}
}

func (l *LocaleInput) ID() string        { return "locale" }
func (l *LocaleInput) Name() string      { return "区域语言传感器" }
func (l *LocaleInput) IsAvailable() bool { return true }

func (l *LocaleInput) Collect(ctx *plugin.SystemContext) error {
	info, err := l.plat.GetLocaleInfo()
	if err != nil {
		return nil
	}
	ctx.SetState("locale.language", info.Language)
	ctx.SetState("locale.region", info.Region)
	ctx.SetState("locale.timezone", info.Timezone)
	ctx.SetState("locale.is_24hour", info.Is24HourFormat)
	return nil
}
