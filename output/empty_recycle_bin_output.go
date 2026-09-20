package output

import (
	"rulecraft/plugin"
)

// EmptyRecycleBinOutput 清空回收站。
type EmptyRecycleBinOutput struct {
	plat PlatformSubset
}

func NewEmptyRecycleBinOutput(plat PlatformSubset) *EmptyRecycleBinOutput {
	return &EmptyRecycleBinOutput{plat: plat}
}

func (e *EmptyRecycleBinOutput) ID() string        { return "empty_recycle_bin" }
func (e *EmptyRecycleBinOutput) Name() string      { return "清空回收站" }
func (e *EmptyRecycleBinOutput) IsAvailable() bool { return true }

func (e *EmptyRecycleBinOutput) Execute(params map[string]interface{}) error {
	return e.plat.EmptyRecycleBin()
}

func (e *EmptyRecycleBinOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
