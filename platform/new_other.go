//go:build !windows

package platform

// NewPlatform 返回当前操作系统的 Platform 实现。
// 非 Windows 系统返回 NoopPlatform（所有功能标记为未实现）。
func NewPlatform() Platform {
	return &NoopPlatform{}
}
