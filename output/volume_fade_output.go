package output

import (
	"fmt"
	"time"

	"rulecraft/plugin"
)

// VolumeFadeOutput 将系统音量渐变到目标值。
type VolumeFadeOutput struct {
	plat PlatformSubset
}

// NewVolumeFadeOutput 创建音量渐变输出插件。
func NewVolumeFadeOutput(plat PlatformSubset) *VolumeFadeOutput {
	return &VolumeFadeOutput{plat: plat}
}

func (v *VolumeFadeOutput) ID() string        { return "volume_fade" }
func (v *VolumeFadeOutput) Name() string      { return "音量渐变" }
func (v *VolumeFadeOutput) IsAvailable() bool { return true }

// Execute 渐变音量。
// params:
//   - level (int, 0-100, 必填): 目标音量
//   - duration_ms (int, 可选): 渐变时长，默认 1500ms
//   - steps (int, 可选): 步数，默认按时长自适应
func (v *VolumeFadeOutput) Execute(params map[string]interface{}) error {
	levelF, ok := params["level"].(float64)
	if !ok {
		return fmt.Errorf("volume_fade: level (0-100) is required")
	}
	target := int(levelF)
	if target < 0 {
		target = 0
	}
	if target > 100 {
		target = 100
	}

	duration := 1500 * time.Millisecond
	if d, ok := params["duration_ms"].(float64); ok && d > 0 {
		duration = time.Duration(int(d)) * time.Millisecond
	}

	steps := 20
	if s, ok := params["steps"].(float64); ok && s >= 2 {
		steps = int(s)
	}

	start, _, err := v.plat.GetSystemVolumeLevel()
	if err != nil {
		start = 0
	}

	stepDur := duration / time.Duration(steps)
	if stepDur < 10*time.Millisecond {
		stepDur = 10 * time.Millisecond
	}

	for i := 1; i <= steps; i++ {
		// 线性插值
		cur := start + (target-start)*i/steps
		if err := v.plat.SetVolume(cur); err != nil {
			return err
		}
		if i < steps {
			time.Sleep(stepDur)
		}
	}
	return v.plat.SetVolume(target)
}

// Reset 不支持 Reset。
func (v *VolumeFadeOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
