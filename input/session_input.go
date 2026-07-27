package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// SessionInput 采集会话状态（如是否锁屏）。
type SessionInput struct {
	plat platform.Platform
}

// NewSessionInput 创建会话输入插件。
func NewSessionInput(plat platform.Platform) *SessionInput {
	return &SessionInput{plat: plat}
}

func (s *SessionInput) ID() string        { return "session" }
func (s *SessionInput) Name() string      { return "会话传感器" }
func (s *SessionInput) IsAvailable() bool { return true }

// Collect 采集会话信息并写入全局状态。
func (s *SessionInput) Collect(ctx *plugin.SystemContext) error {
	info, err := s.plat.GetSessionInfo()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("session.is_locked", false)
			return nil
		}
		return err
	}

	ctx.SetState("session.is_locked", info.IsLocked)
	return nil
}
