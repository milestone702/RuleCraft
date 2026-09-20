package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// PrinterInput 采集默认打印机。
type PrinterInput struct {
	plat platform.Platform
}

func NewPrinterInput(plat platform.Platform) *PrinterInput {
	return &PrinterInput{plat: plat}
}

func (p *PrinterInput) ID() string        { return "printer" }
func (p *PrinterInput) Name() string      { return "打印机传感器" }
func (p *PrinterInput) IsAvailable() bool { return true }

// Collect 写入 printer.default / printer.has_default。
func (p *PrinterInput) Collect(ctx *plugin.SystemContext) error {
	name, err := p.plat.GetDefaultPrinter()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("printer.default", "")
			ctx.SetState("printer.has_default", false)
			return nil
		}
		return err
	}
	ctx.SetState("printer.default", name)
	ctx.SetState("printer.has_default", name != "")
	return nil
}
