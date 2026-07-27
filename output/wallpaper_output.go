package output

import (
	"fmt"

	"rulecraft/plugin"
)

// WallpaperOutput 设置桌面壁纸。
type WallpaperOutput struct {
	plat PlatformSubset
}

// NewWallpaperOutput 创建壁纸设置输出插件。
func NewWallpaperOutput(plat PlatformSubset) *WallpaperOutput {
	return &WallpaperOutput{plat: plat}
}

func (w *WallpaperOutput) ID() string        { return "wallpaper" }
func (w *WallpaperOutput) Name() string      { return "设置壁纸" }
func (w *WallpaperOutput) IsAvailable() bool { return true }

// Execute 设置壁纸。
// params: path (string, 必填: 图片文件路径)
func (w *WallpaperOutput) Execute(params map[string]interface{}) error {
	path, _ := params["path"].(string)
	if path == "" {
		return fmt.Errorf("wallpaper: path is required")
	}
	return w.plat.SetWallpaper(path)
}

// Reset 不支持 Reset。
func (w *WallpaperOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
