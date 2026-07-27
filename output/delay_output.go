package output

import (
	"time"

	"rulecraft/plugin"
)

// DelayOutput 延时等待插件，可作为多个输出动作之间的间隔。
type DelayOutput struct{}

// NewDelayOutput 创建延时输出插件。
func NewDelayOutput() *DelayOutput {
	return &DelayOutput{}
}

func (d *DelayOutput) ID() string        { return "delay" }
func (d *DelayOutput) Name() string      { return "延时等待" }
func (d *DelayOutput) IsAvailable() bool { return true }
func (d *DelayOutput) Reset(map[string]interface{}) error { return plugin.ErrResetNotSupported }

// Execute 等待指定秒数（在输出列表中作为间隔使用）。
// params:
//
//	seconds (number, 必填): 等待秒数
func (d *DelayOutput) Execute(params map[string]interface{}) error {
	var seconds int
	if s, ok := params["seconds"]; ok {
		switch v := s.(type) {
		case float64:
			seconds = int(v)
		case int:
			seconds = v
		}
	}
	if seconds <= 0 {
		seconds = 1
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}
