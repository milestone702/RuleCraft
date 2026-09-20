package output

import (
	"fmt"
	"strings"
)

// DarkModeOutput 切换系统深色/浅色主题。
type DarkModeOutput struct {
	plat PlatformSubset
}

func NewDarkModeOutput(plat PlatformSubset) *DarkModeOutput {
	return &DarkModeOutput{plat: plat}
}

func (d *DarkModeOutput) ID() string        { return "dark_mode" }
func (d *DarkModeOutput) Name() string      { return "深色/浅色模式" }
func (d *DarkModeOutput) IsAvailable() bool { return true }

// Execute 设置主题。
// params: mode (string, 必填): dark / light；或 dark (bool)
func (d *DarkModeOutput) Execute(params map[string]interface{}) error {
	mode, _ := params["mode"].(string)
	if mode == "" {
		if v, ok := params["dark"]; ok {
			if b, ok := v.(bool); ok && b {
				mode = "dark"
			} else {
				mode = "light"
			}
		}
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "dark" && mode != "light" {
		return fmt.Errorf("dark_mode: mode must be dark or light")
	}
	return d.plat.SetThemeMode(mode)
}

// Reset 恢复浅色。
func (d *DarkModeOutput) Reset(params map[string]interface{}) error {
	return d.plat.SetThemeMode("light")
}
