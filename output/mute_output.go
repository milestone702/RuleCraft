package output

import (
	"strings"
)

// MuteOutput 控制系统静音。
type MuteOutput struct {
	plat PlatformSubset
}

// NewMuteOutput 创建静音控制输出插件。
func NewMuteOutput(plat PlatformSubset) *MuteOutput {
	return &MuteOutput{plat: plat}
}

func (m *MuteOutput) ID() string        { return "mute" }
func (m *MuteOutput) Name() string      { return "静音控制" }
func (m *MuteOutput) IsAvailable() bool { return true }

// Execute 设置静音状态。
// params:
//   - mute (bool/int/string, 可选): true/false/1/0/"on"/"off"
//   - toggle (bool, 可选): true 时切换静音（忽略 mute）
func (m *MuteOutput) Execute(params map[string]interface{}) error {
	if toggle, ok := params["toggle"]; ok && toBool(toggle) {
		_, err := m.plat.ToggleMute()
		return err
	}

	if v, ok := params["mute"]; ok {
		return m.plat.SetMute(toBool(v))
	}

	// 默认：切换
	_, err := m.plat.ToggleMute()
	return err
}

// Reset 恢复为取消静音。
func (m *MuteOutput) Reset(params map[string]interface{}) error {
	return m.plat.SetMute(false)
}

func toBool(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		s := strings.ToLower(strings.TrimSpace(b))
		return s == "true" || s == "1" || s == "yes" || s == "on"
	case float64:
		return b != 0
	case int:
		return b != 0
	default:
		return false
	}
}
