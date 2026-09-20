package output

import (
	"fmt"

	"rulecraft/plugin"
)

// ControlPanelOutput 打开控制面板页或 Settings URI。
type ControlPanelOutput struct {
	plat PlatformSubset
}

// NewControlPanelOutput 创建打开控制面板输出插件。
func NewControlPanelOutput(plat PlatformSubset) *ControlPanelOutput {
	return &ControlPanelOutput{plat: plat}
}

func (c *ControlPanelOutput) ID() string        { return "control_panel" }
func (c *ControlPanelOutput) Name() string      { return "打开控制面板" }
func (c *ControlPanelOutput) IsAvailable() bool { return true }

// Execute 打开指定页面。
// params:
//   - page (string, 必填): 控制面板项名或 ms-settings: URI
//     示例: "Microsoft.Windows.Update"、"ms-settings:network-wifi"、"desk.cpl"
func (c *ControlPanelOutput) Execute(params map[string]interface{}) error {
	page, _ := params["page"].(string)
	if page == "" {
		// 兼容 target 键
		page, _ = params["target"].(string)
	}
	if page == "" {
		return fmt.Errorf("control_panel: page is required")
	}
	return c.plat.OpenControlPanelPage(page)
}

// Reset 不支持 Reset。
func (c *ControlPanelOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
