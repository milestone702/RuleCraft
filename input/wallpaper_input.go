package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// WallpaperInput 采集当前壁纸路径。
type WallpaperInput struct {
	plat platform.Platform
}

func NewWallpaperInput(plat platform.Platform) *WallpaperInput {
	return &WallpaperInput{plat: plat}
}

func (w *WallpaperInput) ID() string        { return "wallpaper" }
func (w *WallpaperInput) Name() string      { return "壁纸传感器" }
func (w *WallpaperInput) IsAvailable() bool { return true }

// Collect 写入 wallpaper.path。
func (w *WallpaperInput) Collect(ctx *plugin.SystemContext) error {
	path, err := w.plat.GetWallpaperPath()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("wallpaper.path", "")
			return nil
		}
		return err
	}
	ctx.SetState("wallpaper.path", path)
	return nil
}
