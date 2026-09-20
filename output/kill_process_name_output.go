package output

import (
	"fmt"

	"rulecraft/plugin"
)

// KillProcessNameOutput 按进程名终止进程。
type KillProcessNameOutput struct {
	plat PlatformSubset
}

func NewKillProcessNameOutput(plat PlatformSubset) *KillProcessNameOutput {
	return &KillProcessNameOutput{plat: plat}
}

func (k *KillProcessNameOutput) ID() string        { return "kill_process_name" }
func (k *KillProcessNameOutput) Name() string      { return "按名称杀进程" }
func (k *KillProcessNameOutput) IsAvailable() bool { return true }

// Execute 终止所有匹配名称的进程。
// params: name (string, 必填)，可带或不带 .exe
func (k *KillProcessNameOutput) Execute(params map[string]interface{}) error {
	name, _ := params["name"].(string)
	if name == "" {
		return fmt.Errorf("kill_process_name: name is required")
	}
	return k.plat.KillProcessByName(name)
}

func (k *KillProcessNameOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
