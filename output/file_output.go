package output

import (
	"fmt"
	"os"
	"path/filepath"

	"rulecraft/plugin"
)

// FileOutput 写入文件（纯 Go os 包，无需平台层）。
type FileOutput struct{}

// NewFileOutput 创建文件写入输出插件。
func NewFileOutput() *FileOutput {
	return &FileOutput{}
}

func (f *FileOutput) ID() string        { return "file" }
func (f *FileOutput) Name() string      { return "写入文件" }
func (f *FileOutput) IsAvailable() bool { return true }

// Execute 写入内容到文件。
// params:
//
//	path (string, 必填): 文件路径
//	content (string, 必填): 写入内容
//	append (bool, 可选): 是否追加，默认 false（覆盖）
func (f *FileOutput) Execute(params map[string]interface{}) error {
	path, _ := params["path"].(string)
	if path == "" {
		return fmt.Errorf("file: path is required")
	}

	content, _ := params["content"].(string)

	// 安全检测：防止路径遍历
	if filepath.IsAbs(path) {
		// 仅允许 logs/ 和 tasks/ 目录下的写操作
		allowed := false
		if isUnderDir(path, "logs") || isUnderDir(path, "tasks") {
			allowed = true
		}
		if !allowed {
			return fmt.Errorf("file: path not allowed: %s (only logs/ and tasks/ are writable)", path)
		}
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("file: create directory failed: %w", err)
	}

	appendMode, _ := params["append"].(bool)
	flag := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	file, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return fmt.Errorf("file: open failed: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("file: write failed: %w", err)
	}

	return nil
}

// Reset 不支持 Reset。
func (f *FileOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}

// isUnderDir 检查路径是否在指定目录或其子目录下。
func isUnderDir(absPath, dirName string) bool {
	dir, _ := filepath.Abs(dirName)
	cleanPath := filepath.Clean(absPath)
	cleanDir := filepath.Clean(dir)
	return len(cleanPath) >= len(cleanDir) && cleanPath[:len(cleanDir)] == cleanDir
}
