package output

import (
	"fmt"

	"rulecraft/plugin"
)

// KillProcessOutput 强制终止进程。
type KillProcessOutput struct {
	plat PlatformSubset
}

// NewKillProcessOutput 创建杀进程输出插件。
func NewKillProcessOutput(plat PlatformSubset) *KillProcessOutput {
	return &KillProcessOutput{plat: plat}
}

func (k *KillProcessOutput) ID() string        { return "kill_process" }
func (k *KillProcessOutput) Name() string      { return "杀进程" }
func (k *KillProcessOutput) IsAvailable() bool { return true }

// Execute 终止指定进程。
// params: pid (int, 必填: 进程ID), name (string, 可选: 进程名, 通过 pid 查找)
func (k *KillProcessOutput) Execute(params map[string]interface{}) error {
	pid, ok := params["pid"].(float64)
	if !ok {
		return fmt.Errorf("kill_process: pid is required")
	}
	return k.plat.KillProcess(int(pid))
}

// Reset 不支持 Reset。
func (k *KillProcessOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
