package output

import (
	"fmt"

	"rulecraft/plugin"
)

// ExecOutput 执行外部程序或脚本。
type ExecOutput struct {
	plat PlatformSubset
}

// NewExecOutput 创建执行程序输出插件。
func NewExecOutput(plat PlatformSubset) *ExecOutput {
	return &ExecOutput{plat: plat}
}

func (e *ExecOutput) ID() string        { return "exec" }
func (e *ExecOutput) Name() string      { return "执行程序" }
func (e *ExecOutput) IsAvailable() bool { return true }

// Execute 执行外部程序。
// params: cmd (string, 必填), args (string, 可选), working_dir (string, 可选)
func (e *ExecOutput) Execute(params map[string]interface{}) error {
	cmd, _ := params["cmd"].(string)
	if cmd == "" {
		return fmt.Errorf("exec: cmd is required")
	}

	argsStr, _ := params["args"].(string)
	var args []string
	if argsStr != "" {
		args = []string{argsStr}
	}

	wd, _ := params["working_dir"].(string)
	env := make(map[string]string)

	pid, err := e.plat.ExecuteProcess(cmd, args, wd, env)
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}
	_ = pid
	return nil
}

// Reset 执行程序不支持 Reset。
func (e *ExecOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
