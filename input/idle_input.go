package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// IdleInput 采集用户空闲时间。
type IdleInput struct {
	plat platform.Platform
}

// NewIdleInput 创建空闲时间输入插件。
func NewIdleInput(plat platform.Platform) *IdleInput {
	return &IdleInput{plat: plat}
}

func (i *IdleInput) ID() string        { return "idle" }
func (i *IdleInput) Name() string      { return "空闲时间传感器" }
func (i *IdleInput) IsAvailable() bool { return true }

// Collect 采集空闲秒数并写入全局状态。
func (i *IdleInput) Collect(ctx *plugin.SystemContext) error {
	seconds, err := i.plat.GetIdleSeconds()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("idle.seconds", 0)
			ctx.SetState("idle.is_idle", false)
			return nil
		}
		return err
	}

	ctx.SetState("idle.seconds", seconds)
	ctx.SetState("idle.is_idle", seconds > 300) // 5 分钟无操作视为空闲
	return nil
}
