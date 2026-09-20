package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// RecycleBinInput 采集回收站统计。
type RecycleBinInput struct {
	plat platform.Platform
}

func NewRecycleBinInput(plat platform.Platform) *RecycleBinInput {
	return &RecycleBinInput{plat: plat}
}

func (r *RecycleBinInput) ID() string        { return "recycle_bin" }
func (r *RecycleBinInput) Name() string      { return "回收站传感器" }
func (r *RecycleBinInput) IsAvailable() bool { return true }

// Collect 写入 recycle_bin.count / size_mb / is_empty。
func (r *RecycleBinInput) Collect(ctx *plugin.SystemContext) error {
	count, sizeMB, err := r.plat.GetRecycleBinInfo()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("recycle_bin.count", 0)
			ctx.SetState("recycle_bin.size_mb", 0.0)
			ctx.SetState("recycle_bin.is_empty", true)
			return nil
		}
		return err
	}
	ctx.SetState("recycle_bin.count", count)
	ctx.SetState("recycle_bin.size_mb", sizeMB)
	ctx.SetState("recycle_bin.is_empty", count == 0)
	return nil
}
