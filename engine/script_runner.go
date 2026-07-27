package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"rulecraft/config"
)

// ============================================================================
// 脚本执行器
// ============================================================================

// ScriptRunner 管理规则可执行程序（脚本/二进制）的子进程生命周期。
type ScriptRunner struct {
	execDef   *config.ExecutableDef
	policy    *config.ExecPolicyDef
}

// NewScriptRunner 创建脚本执行器。
func NewScriptRunner(execDef *config.ExecutableDef, policy *config.ExecPolicyDef) *ScriptRunner {
	if policy == nil {
		policy = &config.ExecPolicyDef{
			TimeoutSeconds:  30,
			RetryOnFailure:  0,
			RetryIntervalMs: 1000,
		}
	}
	return &ScriptRunner{
		execDef: execDef,
		policy:  policy,
	}
}

// ExecuteFunction 执行规则中的一个功能。
// functionID 是 functions 中的功能标识符，inputValue 是管道传入的值。
// params 是功能参数，taskContext 是任务上下文（用于环境变量注入）。
func (sr *ScriptRunner) ExecuteFunction(functionID string, inputValue interface{}, params map[string]interface{}, taskContext map[string]interface{}) (interface{}, error) {
	// 构建 stdin JSON
	input := buildStdinJSON(functionID, inputValue, params, taskContext)

	// 序列化
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("serialize stdin failed: %w", err)
	}

	// 构建执行命令
	cmd, args, err := sr.buildCommand()
	if err != nil {
		return nil, fmt.Errorf("build command failed: %w", err)
	}

	// 执行（带重试）
	var lastErr error
	maxRetries := sr.policy.RetryOnFailure + 1

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(sr.policy.RetryIntervalMs) * time.Millisecond)
		}

		output, err := sr.runProcess(cmd, args, inputBytes, taskContext)
		if err == nil {
			return output, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("execution failed after %d attempt(s): %w", maxRetries, lastErr)
}

// buildCommand 构建命令和参数。
func (sr *ScriptRunner) buildCommand() (string, []string, error) {
	execPath := sr.execDef.Path

	// 如果是相对路径，基于规则目录解析
	if !filepath.IsAbs(execPath) {
		// 路径由调用方确保已解析
	}

	switch sr.execDef.Type {
	case "powershell":
		// 强制 -ExecutionPolicy Bypass
		return "powershell.exe", []string{
			"-NoProfile",
			"-ExecutionPolicy", "Bypass",
			"-File", execPath,
		}, nil

	case "batch":
		return execPath, nil, nil

	case "executable":
		return execPath, nil, nil

	default:
		return "", nil, fmt.Errorf("unknown executable type: %s", sr.execDef.Type)
	}
}

// runProcess 执行子进程并等待结果。
func (sr *ScriptRunner) runProcess(cmd string, args []string, stdinData []byte, taskContext map[string]interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(sr.policy.TimeoutSeconds)*time.Second)
	defer cancel()

	c := exec.CommandContext(ctx, cmd, args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// 设置工作目录
	if sr.execDef.WorkingDir != "" {
		c.Dir = sr.execDef.WorkingDir
	}

	// 设置环境变量（合并 execDef.Env + 任务上下文）
	c.Env = os.Environ()
	for k, v := range sr.execDef.Env {
		c.Env = append(c.Env, k+"="+v)
	}
	// 任务上下文环境变量注入
	for k, v := range sr.buildTaskEnv(taskContext) {
		c.Env = append(c.Env, k+"="+v)
	}

	// stdin
	c.Stdin = bytes.NewReader(stdinData)

	// stdout
	var stdout bytes.Buffer
	c.Stdout = &stdout

	// stderr（捕获但不阻断）
	var stderr bytes.Buffer
	c.Stderr = &stderr

	if err := c.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return nil, fmt.Errorf("process error: %s", errMsg)
	}

	// 解析 stdout JSON
	output := stdout.Bytes()
	if len(output) == 0 {
		return nil, fmt.Errorf("empty stdout output")
	}

	return parseOutputJSON(output)
}

// buildTaskEnv 构建任务上下文环境变量映射。
func (sr *ScriptRunner) buildTaskEnv(taskContext map[string]interface{}) map[string]string {
	env := make(map[string]string)
	if taskContext == nil {
		return env
	}

	if v, ok := taskContext["task_id"]; ok {
		env["AUTOMATION_TASK_ID"] = fmt.Sprintf("%v", v)
	}
	if v, ok := taskContext["task_name"]; ok {
		env["AUTOMATION_TASK_NAME"] = fmt.Sprintf("%v", v)
	}
	if v, ok := taskContext["rule_id"]; ok {
		env["AUTOMATION_RULE_ID"] = fmt.Sprintf("%v", v)
	}
	if v, ok := taskContext["processor_id"]; ok {
		env["AUTOMATION_PROCESSOR_ID"] = fmt.Sprintf("%v", v)
	}
	if v, ok := taskContext["log_dir"]; ok {
		env["AUTOMATION_LOG_DIR"] = fmt.Sprintf("%v", v)
	}

	return env
}

// ============================================================================
// 辅助函数
// ============================================================================

// buildStdinJSON 构建传递给子进程的 stdin JSON。
// 包含 value、function_id、params、__task__ 元信息。
func buildStdinJSON(functionID string, inputValue interface{}, params map[string]interface{}, taskContext map[string]interface{}) map[string]interface{} {
	input := map[string]interface{}{
		"function_id": functionID,
		"value":       inputValue,
		"params":      params,
	}

	// 注入 __task__ 元信息
	if taskContext != nil {
		input["__task__"] = taskContext
	}

	return input
}

// parseOutputJSON 解析子进程 stdout 的 JSON 输出。
// 期望格式：{ "value": ..., "_log": [...] }
// 兼容：直接返回 { "value": ... } 或裸值。
func parseOutputJSON(data []byte) (interface{}, error) {
	// 尝试解析为 map
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		// 尝试 heuristic extract — 最外层是否为 value
		// 如果完全解析失败，返回原始字符串
		return string(data), nil
	}

	// 提取 _log 信息（由调用方处理日志写入）
	if logs, ok := result["_log"]; ok {
		_ = logs // 调用方通过回调处理
	}

	// 提取 value
	if value, ok := result["value"]; ok {
		return value, nil
	}

	return result, nil
}

// SanitizePath 路径安全检查：禁止包含 ".."。
func SanitizePath(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal detected: %s", path)
	}
	return nil
}
