package output

import (
	"fmt"

	"rulecraft/plugin"
)

// OpenOutput 使用默认程序打开文件或 URL。
type OpenOutput struct {
	plat PlatformSubset
}

func NewOpenOutput(plat PlatformSubset) *OpenOutput {
	return &OpenOutput{plat: plat}
}

func (o *OpenOutput) ID() string        { return "open" }
func (o *OpenOutput) Name() string      { return "打开文件/URL" }
func (o *OpenOutput) IsAvailable() bool { return true }

func (o *OpenOutput) Execute(params map[string]interface{}) error {
	target, _ := params["target"].(string)
	if target == "" {
		return fmt.Errorf("open: target is required")
	}
	return o.plat.OpenWithDefault(target)
}

func (o *OpenOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
