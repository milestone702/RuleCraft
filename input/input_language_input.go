package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// InputLanguageInput 采集当前键盘布局语言。
type InputLanguageInput struct {
	plat platform.Platform
}

func NewInputLanguageInput(plat platform.Platform) *InputLanguageInput {
	return &InputLanguageInput{plat: plat}
}

func (i *InputLanguageInput) ID() string        { return "input_language" }
func (i *InputLanguageInput) Name() string      { return "输入法/键盘布局传感器" }
func (i *InputLanguageInput) IsAvailable() bool { return true }

// Collect 写入 input_language.code 与常见语言别名。
func (i *InputLanguageInput) Collect(ctx *plugin.SystemContext) error {
	code, err := i.plat.GetInputLanguage()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("input_language.code", "")
			ctx.SetState("input_language.name", "")
			ctx.SetState("input_language.is_chinese", false)
			return nil
		}
		return err
	}
	ctx.SetState("input_language.code", code)
	name := langName(code)
	ctx.SetState("input_language.name", name)
	ctx.SetState("input_language.is_chinese", code == "0804" || code == "0404" || code == "0C04" || code == "1004" || code == "1404")
	return nil
}

func langName(code string) string {
	switch code {
	case "0804":
		return "zh-CN"
	case "0404":
		return "zh-TW"
	case "0409":
		return "en-US"
	case "0809":
		return "en-GB"
	case "0411":
		return "ja-JP"
	case "0412":
		return "ko-KR"
	case "040C":
		return "fr-FR"
	case "0407":
		return "de-DE"
	case "0416":
		return "pt-BR"
	case "0419":
		return "ru-RU"
	default:
		return code
	}
}
