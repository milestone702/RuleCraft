package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// BluetoothInput 采集蓝牙设备信息。
type BluetoothInput struct {
	plat platform.Platform
}

// NewBluetoothInput 创建蓝牙输入插件。
func NewBluetoothInput(plat platform.Platform) *BluetoothInput {
	return &BluetoothInput{plat: plat}
}

func (b *BluetoothInput) ID() string        { return "bluetooth" }
func (b *BluetoothInput) Name() string      { return "蓝牙传感器" }
func (b *BluetoothInput) IsAvailable() bool { return true }

// Collect 采集蓝牙设备。
// 状态键：
//
//	bluetooth.count
//	bluetooth.connected_count
//	bluetooth.paired_count
//	bluetooth.names
//	bluetooth.has_connected — 是否存在已连接设备
func (b *BluetoothInput) Collect(ctx *plugin.SystemContext) error {
	devices, err := b.plat.ListBluetoothDevices()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("bluetooth.count", 0)
			ctx.SetState("bluetooth.connected_count", 0)
			ctx.SetState("bluetooth.paired_count", 0)
			ctx.SetState("bluetooth.has_connected", false)
			ctx.SetState("bluetooth.names", []string{})
			return nil
		}
		return err
	}

	names := make([]string, 0, len(devices))
	connected := 0
	paired := 0
	for _, d := range devices {
		if d.Name != "" {
			names = append(names, d.Name)
		}
		if d.Connected {
			connected++
		}
		if d.Paired {
			paired++
		}
	}
	ctx.SetState("bluetooth.count", len(devices))
	ctx.SetState("bluetooth.connected_count", connected)
	ctx.SetState("bluetooth.paired_count", paired)
	ctx.SetState("bluetooth.has_connected", connected > 0)
	ctx.SetState("bluetooth.names", names)
	return nil
}
