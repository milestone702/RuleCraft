package output

import (
	"fmt"

	"rulecraft/plugin"
)

// BrightnessOutput 调节屏幕亮度。
type BrightnessOutput struct {
	plat PlatformSubset
}

// NewBrightnessOutput 创建亮度调节输出插件。
func NewBrightnessOutput(plat PlatformSubset) *BrightnessOutput {
	return &BrightnessOutput{plat: plat}
}

func (b *BrightnessOutput) ID() string        { return "brightness" }
func (b *BrightnessOutput) Name() string      { return "显示亮度" }
func (b *BrightnessOutput) IsAvailable() bool { return true }

// Execute 设置亮度。
// params: level (int, 0-100, 必填)
func (b *BrightnessOutput) Execute(params map[string]interface{}) error {
	level, ok := params["level"].(float64)
	if !ok {
		return fmt.Errorf("brightness: level (0-100) is required")
	}
	return b.plat.SetBrightness(int(level))
}

// Reset 不支持 Reset。
func (b *BrightnessOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
