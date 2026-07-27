package output

import (
	"rulecraft/plugin"
)

// ScreenshotOutput 截屏输出插件。
type ScreenshotOutput struct {
	plat PlatformSubset
}

func NewScreenshotOutput(plat PlatformSubset) *ScreenshotOutput {
	return &ScreenshotOutput{plat: plat}
}

func (s *ScreenshotOutput) ID() string        { return "screenshot" }
func (s *ScreenshotOutput) Name() string      { return "截屏" }
func (s *ScreenshotOutput) IsAvailable() bool { return true }

func (s *ScreenshotOutput) Execute(params map[string]interface{}) error {
	path, _ := params["path"].(string)
	if path == "" {
		path = "screenshot.png"
	}
	return s.plat.TakeScreenshot(path)
}

func (s *ScreenshotOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
