// Package engine — PluginManager 文件加载器
//
// 负责从磁盘加载 rule.json、{task_id}.conf、task.json 文件，
// 实现设计规范 §6.5 定义的 LoadRule / LoadTaskConfig / LoadAllRules 接口。
package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"rulecraft/config"
)

// ============================================================================
// PluginManager — 规则/任务文件加载器
// ============================================================================

// PluginManager 管理规则插件和任务配置的磁盘加载。
type PluginManager struct {
	inputDir  string // Input_Plugins/
	outputDir string // Output_Plugins/
	tasksDir  string // tasks/

	// 任务文件 mtime 缓存
	cacheMu    sync.Mutex
	taskCache  map[string]*config.TaskDefinition
	taskMtimes map[string]time.Time
}

// NewPluginManager 创建文件加载器。
func NewPluginManager(inputDir, outputDir, tasksDir string) *PluginManager {
	return &PluginManager{
		inputDir:  inputDir,
		outputDir: outputDir,
		tasksDir:  tasksDir,
		taskCache: make(map[string]*config.TaskDefinition),
		taskMtimes: make(map[string]time.Time),
	}
}

// pluginDir 根据 pluginID 查找插件所在目录（在 input/output 中查找）。
func (pm *PluginManager) pluginDir(pluginID string) string {
	inPath := filepath.Join(pm.inputDir, pluginID)
	if _, err := os.Stat(inPath); err == nil {
		return inPath
	}
	return filepath.Join(pm.outputDir, pluginID)
}

// ============================================================================
// Rule 加载
// ============================================================================

// LoadRule 加载指定 pluginID 的 rule.json。
func (pm *PluginManager) LoadRule(pluginID string) (*config.RuleDefinition, error) {
	path := filepath.Join(pm.pluginDir(pluginID), "rule.json")
	return pm.loadRuleFile(path)
}

// loadRuleFile 从指定路径加载并解析 rule.json。
func (pm *PluginManager) loadRuleFile(path string) (*config.RuleDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load rule %s failed: %w", path, err)
	}

	var rule config.RuleDefinition
	if err := json.Unmarshal(data, &rule); err != nil {
		return nil, fmt.Errorf("parse rule %s failed: %w", path, err)
	}

	// 基本校验
	if rule.PluginID == "" {
		return nil, fmt.Errorf("rule %s: plugin_id is required", path)
	}
	if rule.Executable.Path == "" {
		return nil, fmt.Errorf("rule %s: executable.path is required", path)
	}

	return &rule, nil
}

// LoadAllRules 扫描 Input_Plugins/ 和 Output_Plugins/ 下的所有规则。
func (pm *PluginManager) LoadAllRules() (map[string]*config.RuleDefinition, error) {
	rules := make(map[string]*config.RuleDefinition)

	scanDir := func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("[loader] scan dir %s failed: %v", dir, err)
			}
			return
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			rulePath := filepath.Join(dir, entry.Name(), "rule.json")
			if _, err := os.Stat(rulePath); os.IsNotExist(err) {
				continue
			}
			rule, err := pm.loadRuleFile(rulePath)
			if err != nil {
				log.Printf("[loader] skip rule %s: %v", entry.Name(), err)
				continue
			}
			rules[rule.PluginID] = rule
		}
	}

	scanDir(pm.inputDir)
	scanDir(pm.outputDir)
	return rules, nil
}

// ============================================================================
// TaskDefinition ( task.json ) 加载
// ============================================================================

// LoadTask 加载指定 taskID 的 task.json。
func (pm *PluginManager) LoadTask(taskID string) (*config.TaskDefinition, error) {
	path := filepath.Join(pm.tasksDir, taskID+".json")
	return pm.loadTaskFile(path)
}

// loadTaskFile 从指定路径加载 task.json。
func (pm *PluginManager) loadTaskFile(path string) (*config.TaskDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load task %s failed: %w", path, err)
	}

	var task config.TaskDefinition
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("parse task %s failed: %w", path, err)
	}

	if task.TaskID == "" {
		return nil, fmt.Errorf("task %s: task_id is required", path)
	}

	return &task, nil
}

// LoadAllTasks 加载 tasks/ 目录下的所有 task.json，返回 taskID → TaskDefinition 映射。
func (pm *PluginManager) LoadAllTasks() (map[string]*config.TaskDefinition, error) {
	tasks := make(map[string]*config.TaskDefinition)

	entries, err := os.ReadDir(pm.tasksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return tasks, nil
		}
		return nil, fmt.Errorf("read tasks dir failed: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		taskID := strings.TrimSuffix(entry.Name(), ".json")
		task, err := pm.loadTaskFile(filepath.Join(pm.tasksDir, entry.Name()))
		if err != nil {
			log.Printf("[loader] skip task %s: %v", entry.Name(), err)
			continue
		}
		tasks[taskID] = task
	}

	return tasks, nil
}

// LoadAllTasksCached 按文件 mtime 缓存任务定义：仅当文件变更时重新解析。
// 首次调用或文件被修改/删除/新增时会刷新缓存。
func (pm *PluginManager) LoadAllTasksCached() (map[string]*config.TaskDefinition, error) {
	pm.cacheMu.Lock()
	defer pm.cacheMu.Unlock()

	if pm.taskCache == nil {
		pm.taskCache = make(map[string]*config.TaskDefinition)
		pm.taskMtimes = make(map[string]time.Time)
	}

	entries, err := os.ReadDir(pm.tasksDir)
	if err != nil {
		if os.IsNotExist(err) {
			pm.taskCache = make(map[string]*config.TaskDefinition)
			pm.taskMtimes = make(map[string]time.Time)
			return pm.cloneTaskCache(), nil
		}
		return nil, fmt.Errorf("read tasks dir failed: %w", err)
	}

	seen := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		taskID := strings.TrimSuffix(entry.Name(), ".json")
		seen[taskID] = true

		path := filepath.Join(pm.tasksDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		mtime := info.ModTime()

		if old, ok := pm.taskMtimes[taskID]; ok && old.Equal(mtime) {
			if _, exists := pm.taskCache[taskID]; exists {
				continue // 未变更
			}
		}

		task, err := pm.loadTaskFile(path)
		if err != nil {
			log.Printf("[loader] skip task %s: %v", entry.Name(), err)
			delete(pm.taskCache, taskID)
			delete(pm.taskMtimes, taskID)
			continue
		}
		pm.taskCache[taskID] = task
		pm.taskMtimes[taskID] = mtime
	}

	// 清理已删除的任务
	for id := range pm.taskCache {
		if !seen[id] {
			delete(pm.taskCache, id)
			delete(pm.taskMtimes, id)
		}
	}

	return pm.cloneTaskCache(), nil
}

func (pm *PluginManager) cloneTaskCache() map[string]*config.TaskDefinition {
	out := make(map[string]*config.TaskDefinition, len(pm.taskCache))
	for k, v := range pm.taskCache {
		out[k] = v
	}
	return out
}

// ============================================================================
// 文件写入（供 API 持久化使用）
// ============================================================================

// SaveRule 保存 rule.json 到磁盘（根据 direction 选 Input_Plugins/ 或 Output_Plugins/）。
func (pm *PluginManager) SaveRule(rule *config.RuleDefinition) error {
	dir := pm.inputDir
	if rule.Direction == config.DirectionOutput {
		dir = pm.outputDir
	}
	pluginDir := filepath.Join(dir, rule.PluginID)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("create rule dir failed: %w", err)
	}

	path := filepath.Join(pluginDir, "rule.json")
	data, err := json.MarshalIndent(rule, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rule failed: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write rule failed: %w", err)
	}
	return nil
}

// DeleteRule 删除规则目录（在 Input_Plugins/ 和 Output_Plugins/ 中查找）。
func (pm *PluginManager) DeleteRule(pluginID string) error {
	dir := pm.pluginDir(pluginID)
	return os.RemoveAll(dir)
}

// SaveTaskDef 保存 task.json 到磁盘。
func (pm *PluginManager) SaveTaskDef(task *config.TaskDefinition) error {
	if err := os.MkdirAll(pm.tasksDir, 0755); err != nil {
		return fmt.Errorf("create tasks dir failed: %w", err)
	}

	path := filepath.Join(pm.tasksDir, task.TaskID+".json")
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal task failed: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write task failed: %w", err)
	}

	// 使缓存失效，下次加载重新解析
	pm.cacheMu.Lock()
	delete(pm.taskCache, task.TaskID)
	delete(pm.taskMtimes, task.TaskID)
	pm.cacheMu.Unlock()
	return nil
}

// DeleteTaskDef 删除 task.json。
func (pm *PluginManager) DeleteTaskDef(taskID string) error {
	path := filepath.Join(pm.tasksDir, taskID+".json")
	pm.cacheMu.Lock()
	delete(pm.taskCache, taskID)
	delete(pm.taskMtimes, taskID)
	pm.cacheMu.Unlock()
	return os.Remove(path)
}
