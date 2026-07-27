// Package api — 插件市场（从 Git 仓库在线拉取插件）
//
// 支持从 GitHub 或 Gitee 仓库的特定目录下载输入/输出插件。
// 输入插件仓库路径: {repo}/Input_Plugins/
// 输出插件仓库路径: {repo}/Output_Plugins/
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// ============================================================================
// 插件市场类型
// ============================================================================

// PluginMarketItem 插件市场中的插件条目。
type PluginMarketItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"` // "input" 或 "output"
	Version     string `json:"version,omitempty"`
	Source      string `json:"source"` // 下载源 URL
}

// PluginRepo 已配置的插件仓库。
type PluginRepo struct {
	URL         string `json:"url"`          // GitHub/Gitee 仓库 URL
	Name        string `json:"name"`         // 仓库别名
	Platform    string `json:"platform"`     // "github" 或 "gitee"
	Description string `json:"description,omitempty"`
}

// ============================================================================
// 仓库配置管理（内存存储，后续可持久化）
// ============================================================================

var configuredRepos []PluginRepo

// GetRepos 返回已配置的仓库列表。
func GetRepos() []PluginRepo {
	if configuredRepos == nil {
		configuredRepos = []PluginRepo{}
	}
	return configuredRepos
}

// AddRepo 添加一个插件仓库。
func AddRepo(repo PluginRepo) error {
	if repo.URL == "" {
		return fmt.Errorf("repo URL is required")
	}
	if repo.Name == "" {
		repo.Name = extractRepoName(repo.URL)
	}
	repo.Platform = detectPlatform(repo.URL)
	configuredRepos = append(configuredRepos, repo)
	return nil
}

// RemoveRepo 移除插件仓库。
func RemoveRepo(name string) {
	var remaining []PluginRepo
	for _, r := range configuredRepos {
		if r.Name != name {
			remaining = append(remaining, r)
		}
	}
	configuredRepos = remaining
}

// ============================================================================
// 插件列表获取
// ============================================================================

// FetchPluginList 从所有已配置仓库获取插件列表。
func FetchPluginList() ([]PluginMarketItem, error) {
	var allPlugins []PluginMarketItem
	for _, repo := range configuredRepos {
		plugins, err := fetchRepoPlugins(repo)
		if err != nil {
			continue // 单个仓库失败不影响其他仓库
		}
		allPlugins = append(allPlugins, plugins...)
	}
	return allPlugins, nil
}

// fetchRepoPlugins 从单个仓库获取插件列表。
func fetchRepoPlugins(repo PluginRepo) ([]PluginMarketItem, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	// 获取输入插件列表
	inputPlugins, err := fetchDirContents(client, repo, "Input_Plugins")
	if err != nil {
		return nil, fmt.Errorf("fetch Input_Plugins from %s: %w", repo.Name, err)
	}

	// 获取输出插件列表
	outputPlugins, err := fetchDirContents(client, repo, "Output_Plugins")
	if err != nil {
		return nil, fmt.Errorf("fetch Output_Plugins from %s: %w", repo.Name, err)
	}

	var items []PluginMarketItem
	for _, p := range inputPlugins {
		items = append(items, PluginMarketItem{
			ID:          p,
			Name:        p,
			Type:        "input",
			Source:      buildRawURL(repo, "Input_Plugins", p),
		})
	}
	for _, p := range outputPlugins {
		items = append(items, PluginMarketItem{
			ID:          p,
			Name:        p,
			Type:        "output",
			Source:      buildRawURL(repo, "Output_Plugins", p),
		})
	}

	return items, nil
}

// fetchDirContents 获取仓库指定目录下的条目列表。
func fetchDirContents(client *http.Client, repo PluginRepo, dirPath string) ([]string, error) {
	var apiURL string

	switch repo.Platform {
	case "github":
		// GitHub API: GET /repos/{owner}/{repo}/contents/{path}
		owner, repoName := parseGitHubRepo(repo.URL)
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repoName, dirPath)
	case "gitee":
		// Gitee API: GET /api/v5/repos/{owner}/{repo}/contents/{path}
		owner, repoName := parseGitHubRepo(repo.URL) // 相同格式
		apiURL = fmt.Sprintf("https://gitee.com/api/v5/repos/%s/%s/contents/%s", owner, repoName, dirPath)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", repo.Platform)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	// 可选的：req.Header.Set("Authorization", "token YOUR_TOKEN")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d for %s", resp.StatusCode, apiURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析 API 响应
	var items []struct {
		Name string `json:"name"`
		Type string `json:"type"` // "file" or "dir"
	}

	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}

	var names []string
	for _, item := range items {
		if item.Type == "dir" {
			names = append(names, item.Name)
		}
	}
	return names, nil
}

// ============================================================================
// 插件下载与安装
// ============================================================================

// InstallPlugin 从仓库下载并安装一个插件到本地。
// sourceURL 是插件的原始文件 URL（从 Source 字段获取）。
// pluginType 是 "input" 或 "output"。
// installDir 是本地安装目标目录（Input_Plugins/ 或 Output_Plugins/）。
func InstallPlugin(sourceURL, pluginType, pluginID, installDir string) error {
	client := &http.Client{Timeout: 30 * time.Second}

	// 下载 rule.json
	ruleURL := strings.TrimSuffix(sourceURL, "/") + "/rule.json"
	ruleData, err := downloadFile(client, ruleURL)
	if err != nil {
		return fmt.Errorf("download rule.json: %w", err)
	}

	// 校验 rule.json
	var ruleCheck struct {
		PluginID string `json:"plugin_id"`
		Direction string `json:"direction"`
	}
	if err := json.Unmarshal(ruleData, &ruleCheck); err != nil {
		return fmt.Errorf("invalid rule.json: %w", err)
	}

	// 验证 direction 与 pluginType 匹配
	if pluginType == "input" && ruleCheck.Direction == "output" {
		return fmt.Errorf("plugin %s is output type but installing as input", pluginID)
	}
	if pluginType == "output" && ruleCheck.Direction == "input" {
		return fmt.Errorf("plugin %s is input type but installing as output", pluginID)
	}

	// 创建目标目录
	targetDir := path.Join(installDir, pluginID)

	// 保存 rule.json
	if err := saveToFile(path.Join(targetDir, "rule.json"), ruleData); err != nil {
		return err
	}

	// 检查并下载 executable（查找第一个 .ps1 / .bat / .exe 文件）
	execDirURL := strings.TrimSuffix(sourceURL, "/")
	execNames := []string{pluginID + ".ps1", pluginID + ".bat", pluginID + ".exe"}

	for _, execName := range execNames {
		execURL := execDirURL + "/" + execName
		execData, err := downloadFile(client, execURL)
		if err == nil {
			if err := saveToFile(path.Join(targetDir, execName), execData); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

// ============================================================================
// 辅助函数
// ============================================================================

// downloadFile 下载文件内容。
func downloadFile(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// saveToFile 保存数据到文件（自动创建目录）。
func saveToFile(filePath string, data []byte) error {
	dir := path.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}
	return os.WriteFile(filePath, data, 0644)
}

// detectPlatform 从 URL 检测平台类型。
func detectPlatform(repoURL string) string {
	if strings.Contains(repoURL, "github.com") {
		return "github"
	}
	if strings.Contains(repoURL, "gitee.com") {
		return "gitee"
	}
	return "github" // 默认
}

// extractRepoName 从 URL 提取仓库名作为别名。
func extractRepoName(repoURL string) string {
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	return repoURL
}

// parseGitHubRepo 从 URL 解析 owner 和 repoName。
func parseGitHubRepo(repoURL string) (string, string) {
	u, err := url.Parse(strings.TrimSuffix(repoURL, ".git"))
	if err != nil {
		return "", ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

// buildRawURL 构建文件的原始内容 URL。
func buildRawURL(repo PluginRepo, dirs ...string) string {
	owner, repoName := parseGitHubRepo(repo.URL)
	filePath := strings.Join(dirs, "/")

	switch repo.Platform {
	case "github":
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", owner, repoName, filePath)
	case "gitee":
		return fmt.Sprintf("https://gitee.com/%s/%s/raw/main/%s", owner, repoName, filePath)
	default:
		return ""
	}
}
