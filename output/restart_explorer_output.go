package output

import (
	"rulecraft/plugin"
)

// RestartExplorerOutput 重启 Windows 资源管理器。
type RestartExplorerOutput struct {
	plat PlatformSubset
}

// NewRestartExplorerOutput 创建重启资源管理器输出插件。
func NewRestartExplorerOutput(plat PlatformSubset) *RestartExplorerOutput {
	return &RestartExplorerOutput{plat: plat}
}

func (r *RestartExplorerOutput) ID() string        { return "restart_explorer" }
func (r *RestartExplorerOutput) Name() string      { return "重启资源管理器" }
func (r *RestartExplorerOutput) IsAvailable() bool { return true }

// Execute 重启 explorer.exe。
func (r *RestartExplorerOutput) Execute(params map[string]interface{}) error {
	return r.plat.RestartExplorer()
}

// Reset 不支持 Reset。
func (r *RestartExplorerOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
