//go:build windows

package platform

// NewPlatform 返回当前操作系统的 Platform 实现。
// Windows 版本返回真实的 WindowsPlatform。
func NewPlatform() Platform {
	return &WindowsPlatform{}
}
