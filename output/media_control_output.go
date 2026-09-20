package output

import (
	"fmt"

	"rulecraft/plugin"
)

// MediaControlOutput 发送多媒体按键。
type MediaControlOutput struct {
	plat PlatformSubset
}

func NewMediaControlOutput(plat PlatformSubset) *MediaControlOutput {
	return &MediaControlOutput{plat: plat}
}

func (m *MediaControlOutput) ID() string        { return "media_control" }
func (m *MediaControlOutput) Name() string      { return "媒体控制" }
func (m *MediaControlOutput) IsAvailable() bool { return true }

// Execute 发送媒体键。
// params: action (string, 必填): play_pause / next / prev / stop / volume_up / volume_down
func (m *MediaControlOutput) Execute(params map[string]interface{}) error {
	action, _ := params["action"].(string)
	if action == "" {
		return fmt.Errorf("media_control: action is required (play_pause/next/prev/stop/volume_up/volume_down)")
	}
	return m.plat.MediaKey(action)
}

func (m *MediaControlOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
