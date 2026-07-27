// Package api — CRUD 文件持久化
//
// 为 API 端点提供真正的文件读写能力，
// 将 rule.json / task.json / {task_id}.conf 持久化到磁盘。
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"rulecraft/config"
)

// ============================================================================
// Rules 持久化
// ============================================================================

// LoadAllRulesFromDisk 从 Input_Plugins/ 和 Output_Plugins/ 读取所有 rule.json。
func LoadAllRulesFromDisk(inputDir, outputDir string) (map[string]*config.RuleDefinition, error) {
	rules := make(map[string]*config.RuleDefinition)

	// 扫描一个插件目录下的所有 rule.json
	scanDir := func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("[persist] scan dir %s failed: %v", dir, err)
			}
			return
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			path := filepath.Join(dir, entry.Name(), "rule.json")
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var rule config.RuleDefinition
			if err := json.Unmarshal(data, &rule); err != nil {
				continue
			}
			rules[rule.PluginID] = &rule
		}
	}

	scanDir(inputDir)
	scanDir(outputDir)
	return rules, nil
}

// SaveRuleToDisk 保存 rule.json 到磁盘（根据 direction 选择 Input_Plugins/ 或 Output_Plugins/）。
func SaveRuleToDisk(inputDir, outputDir string, rule *config.RuleDefinition) error {
	dir := inputDir
	if rule.Direction == config.DirectionOutput {
		dir = outputDir
	}
	pluginDir := filepath.Join(dir, rule.PluginID)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rule, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(pluginDir, "rule.json"), data, 0644)
}

// DeleteRuleFromDisk 删除规则目录（在两个插件目录中查找）。
func DeleteRuleFromDisk(inputDir, outputDir string, pluginID string) error {
	// 在 input 目录中查找删除
	inPath := filepath.Join(inputDir, pluginID)
	if _, err := os.Stat(inPath); err == nil {
		return os.RemoveAll(inPath)
	}
	// 在 output 目录中查找删除
	outPath := filepath.Join(outputDir, pluginID)
	return os.RemoveAll(outPath)
}

// ============================================================================
// Tasks 持久化
// ============================================================================

// LoadAllTasksFromDisk 从 tasks/ 目录读取所有 task.json。
func LoadAllTasksFromDisk(tasksDir string) (map[string]*config.TaskDefinition, error) {
	tasks := make(map[string]*config.TaskDefinition)

	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return tasks, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(tasksDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var task config.TaskDefinition
		if err := json.Unmarshal(data, &task); err != nil {
			continue
		}
		tasks[task.TaskID] = &task
	}

	return tasks, nil
}

// SaveTaskToDisk 保存 task.json 到磁盘。
func SaveTaskToDisk(tasksDir string, task *config.TaskDefinition) error {
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(tasksDir, task.TaskID+".json"), data, 0644)
}

// DeleteTaskFromDisk 删除任务文件。
func DeleteTaskFromDisk(tasksDir string, taskID string) error {
	return os.Remove(filepath.Join(tasksDir, taskID+".json"))
}

// ============================================================================
// 全局配置持久化
// ============================================================================

// LoadConfig 加载 config.json。
func LoadConfig(path string) (*config.AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := config.DefaultAppConfig()
			return &cfg, nil
		}
		return nil, err
	}
	var cfg config.AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}
	return &cfg, nil
}

// SaveConfig 保存 config.json。
func SaveConfig(path string, cfg *config.AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
