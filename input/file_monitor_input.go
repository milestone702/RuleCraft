package input

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"rulecraft/plugin"
)

// FileMonitorInput 文件/文件夹监控插件。
type FileMonitorInput struct {
	mu        sync.Mutex
	paths     []string
	recursive bool
	checksums map[string]string
	changes   []string
	lastTime  int64
}

func NewFileMonitorInput() *FileMonitorInput {
	return &FileMonitorInput{
		checksums: make(map[string]string),
	}
}

func (f *FileMonitorInput) ID() string        { return "file_monitor" }
func (f *FileMonitorInput) Name() string      { return "文件监控传感器" }
func (f *FileMonitorInput) IsAvailable() bool { return true }

func (f *FileMonitorInput) Configure(paths []string, recursive bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.paths = paths
	f.recursive = recursive
	f.checksums = make(map[string]string)
	f.changes = nil
}

// ConfigureFromMap 从任务参数配置。
func (f *FileMonitorInput) ConfigureFromMap(params map[string]interface{}) error {
	if params == nil {
		return nil
	}
	var paths []string
	if v, ok := params["paths"]; ok {
		switch sv := v.(type) {
		case string:
			if sv != "" {
				paths = append(paths, sv)
			}
		case []interface{}:
			for _, item := range sv {
				if s, ok := item.(string); ok && s != "" {
					paths = append(paths, s)
				}
			}
		}
	}
	recursive := false
	if v, ok := params["recursive"].(bool); ok {
		recursive = v
	}
	if len(paths) > 0 {
		f.Configure(paths, recursive)
	}
	return nil
}

func (f *FileMonitorInput) Collect(ctx *plugin.SystemContext) error {
	f.mu.Lock()
	paths := f.paths
	f.mu.Unlock()

	now := time.Now().Unix()
	var allChanges []string
	totalFiles := 0
	existingPaths := 0

	for _, root := range paths {
		// 检查路径是否存在
		info, err := os.Stat(root)
		if err != nil {
			ctx.SetState("file_monitor.exists."+sanitizeKey(root), false)
			continue
		}
		existingPaths++
		ctx.SetState("file_monitor.exists."+sanitizeKey(root), true)
		ctx.SetState("file_monitor.is_dir."+sanitizeKey(root), info.IsDir())

		// 扫描文件变化
		changes := f.scanPath(root)
		allChanges = append(allChanges, changes...)
		totalFiles += len(f.checksums)
	}

	f.mu.Lock()
	if len(allChanges) > 0 {
		f.changes = append(f.changes, allChanges...)
		f.lastTime = now
	}
	changes := f.changes
	f.changes = nil
	f.mu.Unlock()

	// 设置状态
	ctx.SetState("file_monitor.total_files", totalFiles)
	ctx.SetState("file_monitor.monitored_paths", existingPaths)
	ctx.SetState("file_monitor.timestamp", now)
	ctx.SetState("file_monitor.watching", paths)

	if len(changes) > 0 {
		ctx.SetState("file_monitor.changes", changes)
		ctx.SetState("file_monitor.change_count", len(changes))
	}
	return nil
}

func (f *FileMonitorInput) scanPath(root string) []string {
	var changes []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if !f.recursive && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		hash, err := fileHash(path)
		if err != nil {
			return nil
		}

		f.mu.Lock()
		oldHash, exists := f.checksums[path]
		f.checksums[path] = hash
		f.mu.Unlock()

		if !exists {
			changes = append(changes, fmt.Sprintf("created:%s", path))
		} else if oldHash != hash {
			changes = append(changes, fmt.Sprintf("modified:%s", path))
		}
		return nil
	}

	filepath.Walk(root, walkFn)
	return changes
}

func sanitizeKey(key string) string {
	// 将路径中的特殊字符替换为下划线
	result := make([]byte, 0, len(key))
	for _, c := range []byte(key) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.' {
			result = append(result, c)
		} else if c == ':' || c == '\\' || c == '/' {
			result = append(result, '_')
		}
	}
	return string(result)
}

func fileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := md5.Sum(data)
	return fmt.Sprintf("%x", h), nil
}
