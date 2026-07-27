// Package api 实现 RESTful API 处理程序和嵌入式 Web UI 服务。
package api

import (
	"archive/zip"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"rulecraft/config"
	"rulecraft/engine"
	"rulecraft/plugin"
)

// ============================================================================
// 嵌入式 Web UI
// ============================================================================

//go:embed webui/*
var webUIFS embed.FS

//go:embed webui/User-Manual.md
var manualMD string

//go:embed webui/User-Manual_CN.md
var manualCNMD string

// webFS 剥离 webui/ 前缀后的文件系统。
var webFS fs.FS

func init() {
	var err error
	webFS, err = fs.Sub(webUIFS, "webui")
	if err != nil {
		log.Fatalf("failed to setup embedded web UI: %v", err)
	}
}

// ============================================================================
// API 处理器
// ============================================================================

// Handler 持有所有 API 处理程序所需的依赖。
type Handler struct {
	mu              sync.RWMutex
	runner          *engine.Runner
	registry        *plugin.Registry
	logMgr          *engine.LogManager
	notifier        *engine.Notifier
	appCfg          *config.AppConfig
	inputPluginsDir string
	outputPluginsDir string
	tasksDir        string
	apiKey          string // 全局 API 鉴权密钥
}

// NewHandler 创建 API 处理器。
func NewHandler(runner *engine.Runner, registry *plugin.Registry, logMgr *engine.LogManager, notifier *engine.Notifier, appCfg *config.AppConfig, inputPluginsDir, outputPluginsDir, tasksDir string) *Handler {
	return &Handler{
		runner:           runner,
		registry:         registry,
		logMgr:           logMgr,
		notifier:         notifier,
		appCfg:           appCfg,
		inputPluginsDir:  inputPluginsDir,
		outputPluginsDir: outputPluginsDir,
		tasksDir:         tasksDir,
		apiKey:           appCfg.APIKey,
	}
}

// RegisterRoutes 在指定的 HTTP ServeMux 上注册所有路由。
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// 静态文件服务（嵌入式 Web UI）
	fileServer := http.FileServer(http.FS(webFS))
	mux.Handle("/", fileServer)

	// API 路由
	mux.HandleFunc("/api/rules", h.corsMiddleware(h.handleRules))
	mux.HandleFunc("/api/rules/", h.corsMiddleware(h.handleRuleByID))
	mux.HandleFunc("/api/status", h.corsMiddleware(h.handleStatus))
	mux.HandleFunc("/api/plugins", h.corsMiddleware(h.handlePlugins))
	mux.HandleFunc("/api/tasks", h.corsMiddleware(h.handleTasks))
	mux.HandleFunc("/api/tasks/", h.corsMiddleware(h.handleTaskByID))
	mux.HandleFunc("/api/logs/", h.corsMiddleware(h.handleLogs))
	mux.HandleFunc("/api/notifications/config", h.corsMiddleware(h.handleNotificationsConfig))

	// Plugin Market 路由
	mux.HandleFunc("/api/market/repos", h.corsMiddleware(h.handleMarketRepos))
	mux.HandleFunc("/api/market/plugins", h.corsMiddleware(h.handleMarketPlugins))
	mux.HandleFunc("/api/market/install", h.corsMiddleware(h.handleMarketInstall))

	// Plugin Editor API
	mux.HandleFunc("/api/plugins/export-zip", h.corsMiddleware(h.handlePluginExportZip))
	mux.HandleFunc("/api/plugins/save-script", h.corsMiddleware(h.handlePluginSaveScript))
	mux.HandleFunc("/api/plugins/editor/close", h.corsMiddleware(h.handlePluginEditorClose))
	mux.HandleFunc("/api/plugins/configure", h.corsMiddleware(h.handlePluginConfigure))

	// Debug API（带鉴权）
	mux.HandleFunc("/api/debug", h.authMiddleware(h.corsMiddleware(h.handleDebug)))

	// 使用手册
	mux.HandleFunc("/api/manual", h.corsMiddleware(h.handleManual))

	// Config API
	mux.HandleFunc("/api/config", h.corsMiddleware(h.handleConfig))
	mux.HandleFunc("/api/config/toggle-web", h.corsMiddleware(h.handleConfigToggleWeb))

	// Process API
	mux.HandleFunc("/api/processes/kill", h.corsMiddleware(h.handleKillProcess))
}

// ============================================================================
// CORS 中间件（方便 Web UI 开发调试）
// ============================================================================

func (h *Handler) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// ============================================================================
// 通用响应辅助
// ============================================================================

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[api] json encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ============================================================================
// API 端点实现
// ============================================================================

// GET /api/rules — 获取所有规则
// POST /api/rules — 创建规则
func (h *Handler) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rules, err := LoadAllRulesFromDisk(h.inputPluginsDir, h.outputPluginsDir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "load rules failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, rules)

	case http.MethodPost:
		var rule config.RuleDefinition
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if err := SaveRuleToDisk(h.inputPluginsDir, h.outputPluginsDir, &rule); err != nil {
			writeError(w, http.StatusInternalServerError, "save rule failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, rule)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// GET /api/rules/{id} — 获取规则详情
// PUT /api/rules/{id} — 更新规则
// DELETE /api/rules/{id} — 删除规则
func (h *Handler) handleRuleByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/rules/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "rule ID required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		rules, err := LoadAllRulesFromDisk(h.inputPluginsDir, h.outputPluginsDir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		rule, ok := rules[id]
		if !ok {
			writeError(w, http.StatusNotFound, "rule not found: "+id)
			return
		}
		writeJSON(w, http.StatusOK, rule)

	case http.MethodPut:
		var rule config.RuleDefinition
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		rule.PluginID = id
		if err := SaveRuleToDisk(h.inputPluginsDir, h.outputPluginsDir, &rule); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, rule)

	case http.MethodDelete:
		if err := DeleteRuleFromDisk(h.inputPluginsDir, h.outputPluginsDir, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"deleted": id})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// GET /api/status — 获取当前系统状态快照
func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	states := h.runner.GetStates()
	writeJSON(w, http.StatusOK, states)
}

// GET /api/plugins — 获取所有已注册插件的元数据
func (h *Handler) handlePlugins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	plugins := h.registry.ListAll()
	result := map[string]interface{}{
		"inputs":  h.registry.ListInputs(),
		"outputs": h.registry.ListOutputs(),
		"all":     plugins,
	}

	// 合并所有外部规则插件的 state_key_labels
	allLabels := make(map[string]string)
	rules, err := LoadAllRulesFromDisk(h.inputPluginsDir, h.outputPluginsDir)
	if err == nil {
		for _, rule := range rules {
			if rule != nil && rule.StateKeyLabels != nil {
				for k, v := range rule.StateKeyLabels {
					allLabels[k] = v
				}
			}
		}
	}
	result["state_key_labels"] = allLabels

	writeJSON(w, http.StatusOK, result)
}

// GET /api/tasks — 获取所有任务配置
// POST /api/tasks — 创建任务
func (h *Handler) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tasks, err := LoadAllTasksFromDisk(h.tasksDir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "load tasks failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, tasks)

	case http.MethodPost:
		var task config.TaskDefinition
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if err := SaveTaskToDisk(h.tasksDir, &task); err != nil {
			writeError(w, http.StatusInternalServerError, "save task failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, task)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// GET /api/tasks/{id} — 获取任务详情
// PUT /api/tasks/{id} — 更新任务
// DELETE /api/tasks/{id} — 删除任务
// POST /api/tasks/{id}/trigger — 手动触发任务
func (h *Handler) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	
	// 检查是否是 /trigger 后缀
	if strings.HasSuffix(path, "/trigger") {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		id := strings.TrimSuffix(path, "/trigger")
		if id == "" {
			writeError(w, http.StatusBadRequest, "task ID required")
			return
		}
		// 加载并执行任务
		tasks, err := LoadAllTasksFromDisk(h.tasksDir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		task, ok := tasks[id]
		if !ok {
			writeError(w, http.StatusNotFound, "task not found: "+id)
			return
		}
		// 手动触发：设置 manual_trigger 状态位
		if h.runner != nil {
			h.runner.GetStates()["manual_trigger.trigger"] = task.TaskID
		}
		log.Printf("[api] manual trigger task: %s", id)
		writeJSON(w, http.StatusOK, map[string]string{"triggered": id})
		return
	}

	id := path
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid task ID: %s", id))
		return
	}

	switch r.Method {
	case http.MethodGet:
		tasks, err := LoadAllTasksFromDisk(h.tasksDir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		task, ok := tasks[id]
		if !ok {
			writeError(w, http.StatusNotFound, "task not found: "+id)
			return
		}
		writeJSON(w, http.StatusOK, task)

	case http.MethodPut:
		var task config.TaskDefinition
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		task.TaskID = id
		if err := SaveTaskToDisk(h.tasksDir, &task); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, task)

	case http.MethodDelete:
		if err := DeleteTaskFromDisk(h.tasksDir, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"deleted": id})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// GET /api/logs/{task_id} — 获取指定任务的日志
func (h *Handler) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	taskID := strings.TrimPrefix(r.URL.Path, "/api/logs/")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID required")
		return
	}

	// 获取日志行数参数
	r.ParseForm()
	limit := 100
	// TODO: 实际读取日志文件
	_ = limit

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"task_id": taskID,
		"logs":    []string{},
	})
}

// GET /api/notifications/config — 获取通知配置
// PUT /api/notifications/config — 更新通知配置
func (h *Handler) handleNotificationsConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.appCfg.Notifications)

	case http.MethodPut:
		var cfg config.GlobalNotificationConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		h.notifier.SetGlobalConfig(cfg)
		h.appCfg.Notifications = &cfg
		writeJSON(w, http.StatusOK, cfg)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// ============================================================================
// Plugin Market 路由处理
// ============================================================================

// GET /api/market/repos — 获取已配置的插件仓库列表
// POST /api/market/repos — 添加插件仓库
func (h *Handler) handleMarketRepos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, GetRepos())

	case http.MethodPost:
		var repo PluginRepo
		if err := json.NewDecoder(r.Body).Decode(&repo); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if err := AddRepo(repo); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, repo)

	case http.MethodDelete:
		name := r.URL.Query().Get("name")
		if name == "" {
			writeError(w, http.StatusBadRequest, "repo name required")
			return
		}
		RemoveRepo(name)
		writeJSON(w, http.StatusOK, map[string]string{"removed": name})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// GET /api/market/plugins — 获取所有可用插件
func (h *Handler) handleMarketPlugins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	plugins, err := FetchPluginList()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, plugins)
}

// POST /api/market/install — 安装插件
func (h *Handler) handleMarketInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		SourceURL  string `json:"source_url"`
		PluginType string `json:"plugin_type"`
		PluginID   string `json:"plugin_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.SourceURL == "" || req.PluginType == "" || req.PluginID == "" {
		writeError(w, http.StatusBadRequest, "source_url, plugin_type, plugin_id are required")
		return
	}

	installDir := h.inputPluginsDir
	if req.PluginType == "output" {
		installDir = h.outputPluginsDir
	}
	if err := InstallPlugin(req.SourceURL, req.PluginType, req.PluginID, installDir); err != nil {
		writeError(w, http.StatusInternalServerError, "install failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"installed": req.PluginID,
		"type":      req.PluginType,
	})
}

// GET /api/manual — 返回使用手册 Markdown 内容
func (h *Handler) handleManual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	lang := r.URL.Query().Get("lang")
	if lang == "zh" || lang == "zh-CN" {
		w.Write([]byte(manualCNMD))
	} else {
		w.Write([]byte(manualMD))
	}
}

// POST /api/plugins/export-zip — 导出插件为 ZIP 压缩包
func (h *Handler) handlePluginExportZip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Rule           config.RuleDefinition `json:"rule"`
		ScriptContent  string                `json:"script_content,omitempty"`
		ScriptFileName string                `json:"script_file_name,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// 构建 ZIP
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// 写入 rule.json
	ruleData, err := json.MarshalIndent(req.Rule, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "marshal rule failed: "+err.Error())
		return
	}
	f, err := zipWriter.Create("rule.json")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "zip create failed: "+err.Error())
		return
	}
	if _, err := f.Write(ruleData); err != nil {
		writeError(w, http.StatusInternalServerError, "zip write failed: "+err.Error())
		return
	}

	// 写入可执行脚本（如果有）
	if req.ScriptContent != "" {
		scriptName := req.ScriptFileName
		if scriptName == "" {
			scriptName = req.Rule.PluginID + ".ps1"
		}
		f, err := zipWriter.Create(scriptName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "zip create script failed: "+err.Error())
			return
		}
		if _, err := f.Write([]byte(req.ScriptContent)); err != nil {
			writeError(w, http.StatusInternalServerError, "zip write script failed: "+err.Error())
			return
		}
	}

	if err := zipWriter.Close(); err != nil {
		writeError(w, http.StatusInternalServerError, "zip close failed: "+err.Error())
		return
	}

	// 返回 ZIP 流
	zipData := buf.Bytes()
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s.zip"`, req.Rule.PluginID))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipData)))
	w.Write(zipData)
}

// POST /api/plugins/save-script — 保存插件的可执行脚本文件
func (h *Handler) handlePluginSaveScript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		PluginID  string `json:"plugin_id"`
		Direction string `json:"direction"`
		FileName  string `json:"file_name"`
		Content   string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if req.PluginID == "" || req.FileName == "" {
		writeError(w, http.StatusBadRequest, "plugin_id and file_name are required")
		return
	}

	// 确定目标目录
	baseDir := h.inputPluginsDir
	if req.Direction == "output" {
		baseDir = h.outputPluginsDir
	}

	targetDir := filepath.Join(baseDir, req.PluginID)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "create dir failed: "+err.Error())
		return
	}

	targetPath := filepath.Join(targetDir, req.FileName)
	if err := os.WriteFile(targetPath, []byte(req.Content), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "write script failed: "+err.Error())
		return
	}

	log.Printf("[plugin-editor] saved script: %s", targetPath)
	writeJSON(w, http.StatusOK, map[string]string{
		"saved": targetPath,
	})
}

// POST /api/plugins/configure — 配置插件参数（端口、鉴权等）
func (h *Handler) handlePluginConfigure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		PluginID string                 `json:"plugin_id"`
		Params   map[string]interface{} `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.PluginID == "" {
		writeError(w, http.StatusBadRequest, "plugin_id required")
		return
	}

	// 查找输入插件
	plugin, ok := h.registry.GetInput(req.PluginID)
	if !ok {
		// 尝试输出插件
		outPlugin, ok2 := h.registry.GetOutput(req.PluginID)
		if !ok2 {
			writeError(w, http.StatusNotFound, "plugin not found: "+req.PluginID)
			return
		}
		_ = outPlugin
		writeError(w, http.StatusBadRequest, "plugin does not support configuration")
		return
	}

	// 使用类型断言调用 Configure 方法
	if c, ok := plugin.(interface{ Configure(port int, authKey string) }); ok {
		port := 0
		authKey := ""
		if v, ok := req.Params["port"].(float64); ok {
			port = int(v)
		}
		if v, ok := req.Params["auth_key"].(string); ok {
			authKey = v
		}
		c.Configure(port, authKey)
		log.Printf("[api] configured plugin %s: port=%d", req.PluginID, port)
		writeJSON(w, http.StatusOK, map[string]string{"configured": req.PluginID})
		return
	}
	if c, ok := plugin.(interface{ Configure(port int, proto, authKey string) }); ok {
		port := 0
		authKey := ""
		proto := "tcp"
		if v, ok := req.Params["port"].(float64); ok {
			port = int(v)
		}
		if v, ok := req.Params["auth_key"].(string); ok {
			authKey = v
		}
		if v, ok := req.Params["proto"].(string); ok {
			proto = v
		}
		c.Configure(port, proto, authKey)
		log.Printf("[api] configured plugin %s: port=%d proto=%s", req.PluginID, port, proto)
		writeJSON(w, http.StatusOK, map[string]string{"configured": req.PluginID})
		return
	}
	// network_detect: Configure(target string, port, timeout int, mode string)
	if c, ok := plugin.(interface{ Configure(string, int, int, string) }); ok {
		target := ""
		port := 0
		timeout := 5
		mode := "ping"
		if v, ok := req.Params["target"].(string); ok { target = v }
		if v, ok := req.Params["port"].(float64); ok { port = int(v) }
		if v, ok := req.Params["timeout"].(float64); ok { timeout = int(v) }
		if v, ok := req.Params["mode"].(string); ok { mode = v }
		c.Configure(target, port, timeout, mode)
		log.Printf("[api] configured network_detect: target=%s mode=%s", target, mode)
		writeJSON(w, http.StatusOK, map[string]string{"configured": req.PluginID})
		return
	}
	// http_request: Configure(url string, timeout int)
	if c, ok := plugin.(interface{ Configure(string, int) }); ok {
		url := ""
		timeout := 10
		if v, ok := req.Params["url"].(string); ok { url = v }
		if v, ok := req.Params["timeout"].(float64); ok { timeout = int(v) }
		c.Configure(url, timeout)
		log.Printf("[api] configured http_request: url=%s timeout=%d", url, timeout)
		writeJSON(w, http.StatusOK, map[string]string{"configured": req.PluginID})
		return
	}

	writeError(w, http.StatusBadRequest, "plugin does not support configuration")
}

// POST /api/plugins/editor/close — 关闭插件编辑器，释放资源
func (h *Handler) handlePluginEditorClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	log.Println("[plugin-editor] editor closed by user")
	writeJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

// authMiddleware 检查 API Key 鉴权（仅当 apiKey 非空时启用）。
func (h *Handler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.apiKey == "" {
			next(w, r)
			return
		}
		// 从 X-API-Key 头或 Authorization: Bearer <key> 中读取
		key := r.Header.Get("X-API-Key")
		if key == "" {
			if auth := r.Header.Get("Authorization"); len(auth) > 7 && auth[:7] == "Bearer " {
				key = auth[7:]
			}
		}
		if key != h.apiKey {
			writeError(w, http.StatusUnauthorized, "invalid or missing API key")
			return
		}
		next(w, r)
	}
}

// GET /api/debug — 全局调试信息（需要 API Key 鉴权）
func (h *Handler) handleDebug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 收集调试信息
	states := h.runner.GetStates()
	rules, _ := LoadAllRulesFromDisk(h.inputPluginsDir, h.outputPluginsDir)
	tasks, _ := LoadAllTasksFromDisk(h.tasksDir)

	result := map[string]interface{}{
		"version":     "1.0.0",
		"uptime":      states["time.iso8601"],
		"app_config":  h.appCfg,
		"states":      states,
		"rules":       rules,
		"tasks":       tasks,
		"plugins": map[string]interface{}{
			"inputs":  h.registry.ListInputs(),
			"outputs": h.registry.ListOutputs(),
		},
	}
	writeJSON(w, http.StatusOK, result)
}

// GET/PUT /api/config — 获取/更新全局配置
func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.appCfg)

	case http.MethodPut:
		var cfg map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		// 更新支持的字段
		if v, ok := cfg["api_key"].(string); ok {
			h.appCfg.APIKey = v
			h.apiKey = v
		}
		if v, ok := cfg["poll_interval"].(float64); ok {
			h.appCfg.PollInterval = int(v)
		}
		if v, ok := cfg["port"].(float64); ok {
			h.appCfg.Port = int(v)
		}
		if v, ok := cfg["auto_start"].(bool); ok {
			h.appCfg.AutoStart = v
		}
		// 持久化到磁盘
		if err := SaveConfig("config.json", h.appCfg); err != nil {
			log.Printf("[api] save config failed: %v", err)
		}
		writeJSON(w, http.StatusOK, h.appCfg)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// POST /api/config/toggle-web — 切换 Web 服务启停
func (h *Handler) handleConfigToggleWeb(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	// 通知主循环切换 Web 服务（通过全局函数的闭包回调）
	// 此处用 h.appCfg 标记，实际切换在 main.go 中由 OnToggleWeb 处理
	log.Println("[api] web toggle requested (handled by GUI callback)")
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": true})
}

// POST /api/processes/kill — 结束进程
func (h *Handler) handleKillProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		PID int `json:"pid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.PID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid PID")
		return
	}

	// 调用内置 kill_process 输出插件
	plugin, ok := h.registry.GetOutput("kill_process")
	if !ok {
		writeError(w, http.StatusInternalServerError, "kill_process plugin not found")
		return
	}

	if err := plugin.Execute(map[string]interface{}{"pid": float64(req.PID)}); err != nil {
		writeError(w, http.StatusInternalServerError, "kill failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"killed": req.PID,
	})
}
