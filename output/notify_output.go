package output

import (
	"fmt"

	"rulecraft/plugin"
)

// NotifyOutput 显示桌面通知。
type NotifyOutput struct {
	plat PlatformSubset
}

// NewNotifyOutput 创建通知输出插件。
func NewNotifyOutput(plat PlatformSubset) *NotifyOutput {
	return &NotifyOutput{plat: plat}
}

func (n *NotifyOutput) ID() string        { return "notify" }
func (n *NotifyOutput) Name() string      { return "桌面通知" }
func (n *NotifyOutput) IsAvailable() bool { return true }

// Execute 显示通知。
// params: title (string), message (string), level (string, 可选)
func (n *NotifyOutput) Execute(params map[string]interface{}) error {
	title, _ := params["title"].(string)
	message, _ := params["message"].(string)
	level, _ := params["level"].(string)

	if title == "" {
		title = "RuleCraft"
	}
	if message == "" {
		return fmt.Errorf("notify: message is required")
	}
	if level == "" {
		level = "info"
	}

	return n.plat.ShowNotification(title, message, level)
}

// Reset 通知不支持 Reset。
func (n *NotifyOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
