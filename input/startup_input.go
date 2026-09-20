package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// StartupInput 采集启动项列表。
type StartupInput struct {
	plat platform.Platform
}

func NewStartupInput(plat platform.Platform) *StartupInput {
	return &StartupInput{plat: plat}
}

func (s *StartupInput) ID() string        { return "startup" }
func (s *StartupInput) Name() string      { return "启动项传感器" }
func (s *StartupInput) IsAvailable() bool { return true }

// Collect 写入 startup.count / startup.names。
func (s *StartupInput) Collect(ctx *plugin.SystemContext) error {
	items, err := s.plat.GetStartupItems()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("startup.count", 0)
			ctx.SetState("startup.names", []string{})
			return nil
		}
		return err
	}
	ctx.SetState("startup.count", len(items))
	ctx.SetState("startup.names", items)
	return nil
}
