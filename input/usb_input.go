package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// USBInput 采集 USB 设备列表。
type USBInput struct {
	plat platform.Platform
}

// NewUSBInput 创建 USB 输入插件。
func NewUSBInput(plat platform.Platform) *USBInput {
	return &USBInput{plat: plat}
}

func (u *USBInput) ID() string        { return "usb" }
func (u *USBInput) Name() string      { return "USB 设备传感器" }
func (u *USBInput) IsAvailable() bool { return true }

// Collect 采集 USB 设备信息并写入全局状态。
// 状态键：
//
//	usb.count              — 设备数量
//	usb.names              — 名称数组（JSON 可序列化）
//	usb.has_device         — 由 params.device_name 匹配是否存在（可选）
func (u *USBInput) Collect(ctx *plugin.SystemContext) error {
	devices, err := u.plat.ListUSBDevices()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("usb.count", 0)
			ctx.SetState("usb.names", []string{})
			return nil
		}
		return err
	}

	names := make([]string, 0, len(devices))
	for _, d := range devices {
		if d.Name != "" {
			names = append(names, d.Name)
		}
	}
	ctx.SetState("usb.count", len(devices))
	ctx.SetState("usb.names", names)
	return nil
}
