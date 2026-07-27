package engine

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"rulecraft/config"
)

// ============================================================================
// 日志级别常量
// ============================================================================

const (
	LogLevelDebug   = "debug"
	LogLevelVerbose = "verbose"
	LogLevelInfo    = "info"
	LogLevelWarning = "warning"
	LogLevelError   = "error"
)

// logLevelNumeric 日志级别数值映射。
var logLevelNumeric = map[string]int{
	LogLevelDebug:   -1,
	LogLevelVerbose: 0,
	LogLevelInfo:    1,
	LogLevelWarning: 2,
	LogLevelError:   3,
}

// ============================================================================
// 日志系统
// ============================================================================

// Logger 是统一日志写入器。
// 它管理单个日志文件的写入、轮转和清理。
type Logger struct {
	mu       sync.Mutex
	cfg      config.FileLogConfig
	file     *os.File
	writer   io.Writer
	dir      string
	filename string // 不含路径的文件名
	basePath string // 完整路径
	minLevel int    // 最低记录级别的数值

	// 轮转触发器
	lastRotateCheck time.Time

	// 磁盘空间保护
	lowSpaceMode bool
}

// NewLogger 创建日志写入器。
// basePath 是日志文件路径（如 "logs/app.log"），cfg 是日志文件配置。
func NewLogger(basePath string, cfg config.FileLogConfig) (*Logger, error) {
	// 确保目录存在
	dir := filepath.Dir(basePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create log directory %s failed: %w", dir, err)
	}

	l := &Logger{
		cfg:             cfg,
		dir:             dir,
		filename:        filepath.Base(basePath),
		basePath:        basePath,
		lastRotateCheck: time.Now(),
	}

	// 解析最低级别
	if cfg.Path != "" {
		l.basePath = cfg.Path
	}
	if level, ok := logLevelNumeric[cfg.Encoding]; ok {
		_ = level
	}
	l.minLevel = logLevelNumeric[LogLevelInfo] // 默认 info

	// 打开文件
	if err := l.openFile(); err != nil {
		return nil, err
	}

	return l, nil
}

// SetLevel 设置最低记录级别。
func (l *Logger) SetLevel(level string) {
	if n, ok := logLevelNumeric[level]; ok {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.minLevel = n
	}
}

// Write 写入一条日志记录。
func (l *Logger) Write(entry config.LogEntry) {
	if n, ok := logLevelNumeric[entry.Level]; ok {
		if n < l.minLevel {
			return
		}
	}

	line := formatLogEntry(entry)
	l.writeString(line)
}

// writeString 写入字符串（含轮转检查）。
func (l *Logger) writeString(line string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return
	}

	// 检查是否需要轮转
	if err := l.checkRotate(); err != nil {
		log.Printf("[logger] rotate error: %v", err)
	}

	// 写入
	if _, err := fmt.Fprint(l.file, line); err != nil {
		log.Printf("[logger] write error: %v", err)
	}
}

// Close 关闭日志文件。
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// RotateNow 立即执行日志轮转。
func (l *Logger) RotateNow() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rotate()
}

// ============================================================================
// 文件管理
// ============================================================================

// openFile 打开日志文件（追加模式）。
func (l *Logger) openFile() error {
	f, err := os.OpenFile(l.basePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	l.file = f
	return nil
}

// checkRotate 检查并触发轮转。
func (l *Logger) checkRotate() error {
	// 1. 按大小轮转
	info, err := os.Stat(l.basePath)
	if err != nil {
		return nil
	}

	maxSize := int64(l.cfg.MaxSizeMB) * 1024 * 1024
	if maxSize > 0 && info.Size() >= maxSize {
		return l.rotate()
	}

	// 2. 定时检查（兜底）
	if time.Since(l.lastRotateCheck) > 5*time.Minute {
		l.lastRotateCheck = time.Now()
		return l.cleanup()
	}

	return nil
}

// rotate 执行一次轮转。
func (l *Logger) rotate() error {
	// 关闭当前文件
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}

	// 文件移动：app.log → app.log.1, app.log.1 → app.log.2, ...
	l.shiftFiles()

	// 打开新文件
	if err := l.openFile(); err != nil {
		return err
	}

	// 执行后置清理
	return l.cleanup()
}

// shiftFiles 将历史文件依次后移。
func (l *Logger) shiftFiles() {
	maxFiles := l.cfg.MaxFiles
	if maxFiles <= 0 {
		maxFiles = 5
	}

	// 从最旧的开始删除
	for i := maxFiles - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.basePath, i)
		if l.cfg.Compress != nil && *l.cfg.Compress {
			oldPath += ".gz"
		}
		os.Remove(oldPath)
	}

	// 后移
	for i := maxFiles - 2; i >= 0; i-- {
		src := fmt.Sprintf("%s.%d", l.basePath, i)
		if i == 0 {
			src = l.basePath
		}
		dst := fmt.Sprintf("%s.%d", l.basePath, i+1)
		if l.cfg.Compress != nil && *l.cfg.Compress && i > 0 {
			dst += ".gz"
		}
		os.Rename(src, dst)
	}
}

// cleanup 执行后置清理（按数量、总大小、时间）。
func (l *Logger) cleanup() error {
	// 获取所有日志文件
	files, err := filepath.Glob(l.basePath + "*")
	if err != nil {
		return err
	}

	// 按修改时间排序
	type logFile struct {
		path    string
		modTime time.Time
		size    int64
	}
	var logFiles []logFile
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		logFiles = append(logFiles, logFile{
			path:    f,
			modTime: info.ModTime(),
			size:    info.Size(),
		})
	}
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].modTime.Before(logFiles[j].modTime)
	})

	// 按时间清理
	if l.cfg.MaxAgeDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -l.cfg.MaxAgeDays)
		for _, f := range logFiles {
			if f.modTime.Before(cutoff) {
				os.Remove(f.path)
			}
		}
	}

	// 重新统计
	files, _ = filepath.Glob(l.basePath + "*")
	var totalSize int64
	for _, f := range files {
		info, _ := os.Stat(f)
		if info != nil {
			totalSize += info.Size()
		}
	}

	// 按总大小清理
	maxTotal := int64(l.cfg.MaxTotalSizeMB) * 1024 * 1024
	if maxTotal > 0 && totalSize > maxTotal {
		targetSize := int64(float64(maxTotal) * 0.8)
		for _, f := range logFiles {
			if totalSize <= targetSize {
				break
			}
			os.Remove(f.path)
			totalSize -= f.size
		}
	}

	return nil
}

// ============================================================================
// 日志格式
// ============================================================================

// formatLogEntry 格式化单条日志。
// 格式：{timestamp}  [{level}]  [{source_type}:{source_id}]  {message}
func formatLogEntry(entry config.LogEntry) string {
	var parts []string

	// 时间戳
	ts := entry.Timestamp
	if ts == "" {
		ts = time.Now().Format("2006-01-02T15:04:05.000-07:00")
	}
	parts = append(parts, ts)

	// 级别
	parts = append(parts, fmt.Sprintf("[%s]", entry.Level))

	// 来源
	source := ""
	if entry.SourceID != "" && entry.SourceType != "" {
		source = fmt.Sprintf("[%s:%s]", entry.SourceType, entry.SourceID)
	} else if entry.SourceType != "" {
		source = fmt.Sprintf("[%s]", entry.SourceType)
	}
	if source != "" {
		parts = append(parts, source)
	}

	// 消息
	parts = append(parts, entry.Message)

	return strings.Join(parts, "  ") + "\n"
}

// ============================================================================
// 日志管理器 — 管理多个 Logger 实例
// ============================================================================

// LogManager 管理所有 Logger 实例（系统日志 + 任务日志 + 插件日志）。
type LogManager struct {
	mu       sync.RWMutex
	system   *Logger
	taskLogs map[string]*Logger // task_id → Logger
	pluginLogs map[string]*Logger // plugin_id → Logger
	globalCfg config.LoggingConfig
}

// NewLogManager 创建日志管理器。
func NewLogManager(globalCfg config.LoggingConfig) *LogManager {
	return &LogManager{
		taskLogs:   make(map[string]*Logger),
		pluginLogs: make(map[string]*Logger),
		globalCfg: globalCfg,
	}
}

// InitSystem 初始化系统日志。
func (lm *LogManager) InitSystem(basePath string) error {
	cfg := config.FileLogConfig{}
	if lm.globalCfg.File != nil {
		cfg = *lm.globalCfg.File
	}
	if basePath != "" {
		cfg.Path = basePath
	}

	logger, err := NewLogger(basePath, cfg)
	if err != nil {
		return err
	}
	lm.mu.Lock()
	lm.system = logger
	lm.mu.Unlock()
	return nil
}

// GetSystem 返回系统日志写入器。
func (lm *LogManager) GetSystem() *Logger {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.system
}

// GetTaskLogger 获取或创建任务日志写入器。
func (lm *LogManager) GetTaskLogger(taskID string) *Logger {
	lm.mu.RLock()
	if l, ok := lm.taskLogs[taskID]; ok {
		lm.mu.RUnlock()
		return l
	}
	lm.mu.RUnlock()

	// 创建新的任务日志
	path := filepath.Join("logs", "tasks", taskID+".log")
	cfg := config.FileLogConfig{
		Path:      path,
		MaxSizeMB: 5,
		MaxFiles:  3,
	}
	if lm.globalCfg.File != nil {
		cfg.MaxSizeMB = lm.globalCfg.File.MaxSizeMB
		cfg.MaxFiles = lm.globalCfg.File.MaxFiles
	}

	logger, err := NewLogger(path, cfg)
	if err != nil {
		log.Printf("[logmanager] create task logger failed: %v", err)
		return lm.system
	}

	lm.mu.Lock()
	lm.taskLogs[taskID] = logger
	lm.mu.Unlock()
	return logger
}

// CloseAll 关闭所有日志写入器。
func (lm *LogManager) CloseAll() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.system != nil {
		lm.system.Close()
	}
	for _, l := range lm.taskLogs {
		l.Close()
	}
	for _, l := range lm.pluginLogs {
		l.Close()
	}
}
