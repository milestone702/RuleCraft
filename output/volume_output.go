package output

import (
	"fmt"

	"rulecraft/plugin"
)

// VolumeOutput 控制系统主音量。
type VolumeOutput struct {
	plat PlatformSubset
}

// NewVolumeOutput 创建音量控制输出插件。
func NewVolumeOutput(plat PlatformSubset) *VolumeOutput {
	return &VolumeOutput{plat: plat}
}

func (v *VolumeOutput) ID() string        { return "volume" }
func (v *VolumeOutput) Name() string      { return "音量控制" }
func (v *VolumeOutput) IsAvailable() bool { return true }

// Execute 设置音量。
// params: level (int, 0-100, 必填), mute (bool, 可选)
func (v *VolumeOutput) Execute(params map[string]interface{}) error {
	level, ok := params["level"].(float64)
	if !ok {
		return fmt.Errorf("volume: level (0-100) is required")
	}
	return v.plat.SetVolume(int(level))
}

// Reset 不支持 Reset。
func (v *VolumeOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
