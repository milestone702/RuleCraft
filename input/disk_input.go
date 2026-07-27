package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// DiskInput 采集磁盘使用信息。
// 默认自动检测所有可用磁盘驱动器。
type DiskInput struct {
	plat   platform.Platform
	drives []string
	cached bool
}

// NewDiskInput 创建磁盘输入插件，自动检测所有可用磁盘。
func NewDiskInput(plat platform.Platform) *DiskInput {
	return &DiskInput{plat: plat}
}

// NewDiskInputWithDrives 创建磁盘输入插件并指定监控的磁盘列表。
func NewDiskInputWithDrives(plat platform.Platform, drives []string) *DiskInput {
	if len(drives) == 0 {
		drives = []string{"C:"}
	}
	return &DiskInput{plat: plat, drives: drives}
}

func (d *DiskInput) ID() string        { return "disk" }
func (d *DiskInput) Name() string      { return "磁盘传感器" }
func (d *DiskInput) IsAvailable() bool { return true }

// Collect 自动检测所有可用磁盘并采集信息。
func (d *DiskInput) Collect(ctx *plugin.SystemContext) error {
	// 首次运行时检测可用磁盘（如果尚未指定）
	if !d.cached && len(d.drives) == 0 {
		d.drives = detectDrives(d.plat)
		d.cached = true
	}

	// 先测试一个盘判断平台是否支持
	if len(d.drives) == 0 {
		return nil
	}
	_, err := d.plat.GetDiskInfo(d.drives[0])
	if platform.IsErrNotImplemented(err) {
		return nil
	}

	for _, drive := range d.drives {
		info, err := d.plat.GetDiskInfo(drive)
		if err != nil {
			continue
		}
		ctx.SetState("disk."+drive+".free_gb", info.FreeGB)
		ctx.SetState("disk."+drive+".total_gb", info.TotalGB)
		ctx.SetState("disk."+drive+".used_percent", info.UsedPercent)
	}

	return nil
}

// detectDrives 检测系统中的可用磁盘驱动器。
func detectDrives(plat platform.Platform) []string {
	var drives []string
	// 尝试 A: 到 Z:
	for _, letter := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		drive := string(letter) + ":"
		_, err := plat.GetDiskInfo(drive)
		if err == nil {
			drives = append(drives, drive)
		}
	}
	return drives
}
