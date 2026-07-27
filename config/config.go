// Package config 定义 RuleCraft 引擎的所有核心数据结构和配置类型。
// 这些结构体是引擎各模块之间交换数据的契约，也是 JSON 配置文件的 Go 映射。
package config

// ============================================================================
// 自动化规则 (Rule)
// ============================================================================

// Rule 定义单条自动化规则。
type Rule struct {
	ID            string         `json:"id"`             // 规则唯一标识符
	Name          string         `json:"name"`            // 展示名称
	Enabled       bool           `json:"enabled"`         // 是否启用
	Inputs        []string       `json:"inputs"`          // 需要激活的输入插件 ID 列表
	ConditionTree ConditionNode  `json:"condition_tree"`  // 嵌套逻辑树根节点
	Outputs       []OutputAction `json:"outputs"`         // 条件成立时触发的动作列表
}

// ConditionNode 是递归逻辑树的 AST 节点，支持 AND/OR/NOT 组合与原子判断。
type ConditionNode struct {
	// 组合节点属性
	LogicalOp string          `json:"logical_operator,omitempty"` // "AND", "OR", "NOT"
	Children  []ConditionNode `json:"conditions,omitempty"`       // 子条件（仅组合节点使用）

	// 叶子节点属性（原子判断）
	Type     string      `json:"type,omitempty"`     // 如 "state"
	StateKey string      `json:"state_key,omitempty"` // 状态键（type=state 时使用）
	Operator string      `json:"operator,omitempty"` // equals, contains, greater_than 等
	Value    interface{} `json:"value,omitempty"`    // 目标对比值

	// 稳定性阈值：二选一
	Threshold    int `json:"threshold,omitempty"`     // 连续 N 次成立后才算真（基于轮询次数）
	StableSeconds int `json:"stable_seconds,omitempty"` // 持续成立 N 秒后才算真（基于时间）

	// 可选：标记该条件依赖哪个 processor 的输出
	SourceProcessorID string `json:"source_processor_id,omitempty"`
}

// OutputAction 定义条件满足时需要触发的动作。
type OutputAction struct {
	PluginID string                 `json:"plugin_id"`         // 输出插件 ID
	Params   map[string]interface{} `json:"params,omitempty"`  // 动作参数
}

// ============================================================================
// 插件规则体系（rule.json + {task_id}.conf + task.json 共用类型）
// ============================================================================

// RuleDefinition 对应 rules/{plugin_id}/rule.json 的完整结构。
type RuleDefinition struct {
	PluginID    string               `json:"plugin_id"`              // 全局唯一标识符
	Name        string               `json:"name"`                   // 展示名称
	Version     string               `json:"version"`                // SemVer 版本号
	Author      string               `json:"author,omitempty"`       // 作者
	PublishURL  string               `json:"publish_url,omitempty"`  // 发布地址
	Description string               `json:"description,omitempty"`  // 功能描述
	Tags        []string             `json:"tags,omitempty"`         // 分类标签
	License     string               `json:"license,omitempty"`      // 开源许可证

	Direction RuleDirection         `json:"direction,omitempty"`    // "input" 或 "output"
	Lifecycle *RuleLifecycle        `json:"lifecycle,omitempty"`    // 仅 output 需要

	Notifications *NotificationConfig `json:"notifications,omitempty"` // 通知配置
	Logging       *LoggingConfig     `json:"logging,omitempty"`       // 日志配置
	Executable    ExecutableDef      `json:"executable"`              // 可执行程序定义
	Dependencies  DependenciesDef    `json:"dependencies,omitempty"`  // 前置依赖
	ExecutionPolicy ExecPolicyDef  `json:"execution_policy,omitempty"` // 执行策略
	Functions     map[string]FuncDef `json:"functions,omitempty"`     // 功能清单

	// StateKeyLabels 声明该插件产出的状态键的中文备注。
	// 格式：{"power.battery_percent": "电池百分比"}
	// 前端下拉选择状态键时会显示此备注。
	StateKeyLabels map[string]string `json:"state_key_labels,omitempty"`
}

// RuleDirection 规则方向。
type RuleDirection string

const (
	DirectionInput  RuleDirection = "input"
	DirectionOutput RuleDirection = "output"
)

// RuleLifecycle 输出型规则的生命周期声明。
type RuleLifecycle struct {
	SupportsExecute bool   `json:"supports_execute"`         // 支持 Execute（条件 false→true 触发）
	SupportsReset   bool   `json:"supports_reset"`           // 支持 Reset（条件 true→false 触发）
	Description     string `json:"description,omitempty"`    // 行为描述
}

// ExecutableDef 可执行程序定义。
type ExecutableDef struct {
	Type       string            `json:"type"`                  // powershell / batch / executable
	Path       string            `json:"path"`                  // 可执行程序路径（相对或绝对）
	Args       string            `json:"args,omitempty"`        // 默认命令行参数模板
	WorkingDir string            `json:"working_dir,omitempty"` // 工作目录
	Env        map[string]string `json:"env,omitempty"`         // 额外环境变量
}

// ExecPolicyDef 执行策略（rule.json 中的 execution_policy）。
type ExecPolicyDef struct {
	TimeoutSeconds  int  `json:"timeout_seconds,omitempty"`   // 单次执行超时，默认 30
	RetryOnFailure  int  `json:"retry_on_failure,omitempty"`  // 失败后最大重试次数，默认 0
	RetryIntervalMs int  `json:"retry_interval_ms,omitempty"` // 重试间隔毫秒，默认 1000
	MaxMemoryMB     int  `json:"max_memory_mb,omitempty"`     // 子进程内存上限 MB，默认 100
	RunAsync        bool `json:"run_async,omitempty"`         // 是否异步执行，默认 false
	CooldownSeconds int  `json:"cooldown_seconds,omitempty"`  // 冷却时间（秒），默认 0
}

// DependenciesDef 前置依赖声明。
type DependenciesDef struct {
	RequiresAdmin            bool     `json:"requires_admin,omitempty"`
	RequiresPowerShellVer    string   `json:"requires_powershell_version,omitempty"`
	RequiresNetwork          bool     `json:"requires_network,omitempty"`
	RequiresComponents       []string `json:"requires_components,omitempty"`
	DependsOnPlugins         []string `json:"depends_on_plugins,omitempty"`
}

// FuncDef 功能清单中的单个功能定义（对应 rule.json 的 functions.{id}）。
type FuncDef struct {
	Name        string                 `json:"name"`                  // 功能展示名称
	Description string                 `json:"description,omitempty"` // 详细描述
	InputParams map[string]ParamDef    `json:"input_params"`          // 输入参数 Schema
	Output      map[string]ParamDef    `json:"output"`                // 输出参数 Schema
	Validate    map[string]interface{} `json:"validate_rules,omitempty"` // 输入验证规则
}

// ParamDef 参数 Schema 定义。
type ParamDef struct {
	Type        string      `json:"type"`                   // string / integer / number / boolean
	Description string      `json:"description,omitempty"`  // 参数说明
	Required    bool        `json:"required,omitempty"`     // 是否必填
	Default     interface{} `json:"default,omitempty"`      // 默认值
	Enum        []any       `json:"enum,omitempty"`         // 允许值枚举
	Minimum     *float64    `json:"minimum,omitempty"`      // 最小值
	Maximum     *float64    `json:"maximum,omitempty"`      // 最大值
}

// ============================================================================
// 任务调度配置（task.json）
// ============================================================================

// TaskDefinition 对应 tasks/{task_id}.json 的顶层结构。
type TaskDefinition struct {
	TaskID      string             `json:"task_id"`                 // 任务唯一标识符
	Name        string             `json:"name"`                    // 展示名称
	Description string             `json:"description,omitempty"`
	Enabled     bool               `json:"enabled"`                  // 默认 true
	Version     string             `json:"version,omitempty"`

	DataSources      []DataSourceDef   `json:"data_sources"`
	Condition        *ConditionNode    `json:"condition,omitempty"`
	Outputs          []OutputBinding   `json:"outputs,omitempty"`

	Notifications    *NotificationConfig `json:"notifications,omitempty"`
	Logging          *LoggingConfig      `json:"logging,omitempty"`
	ExecutionPolicy  *TaskExecutionPolicy `json:"execution_policy,omitempty"`
}

// DataSourceDef 数据源定义。
type DataSourceDef struct {
	ID          string                 `json:"id"`                     // 任务内唯一标识符
	PluginID    string                 `json:"plugin_id,omitempty"`    // 输入插件 ID（前端渲染用）
	StateKey    string                 `json:"state_key"`              // 引用的全局状态键
	Description string                 `json:"description,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`       // 插件特定配置参数
}

// OutputBinding 输出绑定定义。
type OutputBinding struct {
	ID             string                 `json:"id"`                       // 输出实例标识符
	Type           string                 `json:"type"`                     // builtin / rule
	PluginID       string                 `json:"plugin_id,omitempty"`      // type=builtin 时必填
	RuleID         string                 `json:"rule_id,omitempty"`        // type=rule 时必填
	ConfigInstance string                 `json:"config_instance,omitempty"` // type=rule 时必填
	Params         map[string]interface{} `json:"params,omitempty"`         // 参数
	TriggerOn      string                 `json:"trigger_on"`               // true / false / both
	Conditions     OutputConditions       `json:"conditions,omitempty"`     // 防抖等条件
	TriggerThreshold int                  `json:"trigger_threshold,omitempty"` // 连续成立N次后才触发（前N-1次仅写日志）
}

// OutputConditions 输出防抖条件。
type OutputConditions struct {
	CooldownSeconds int    `json:"cooldown_seconds,omitempty"`
	CooldownKey     string `json:"cooldown_key,omitempty"`
}

// TaskExecutionPolicy 任务执行策略。
type TaskExecutionPolicy struct {
	MaxConcurrency      int    `json:"max_concurrency,omitempty"`       // 0=不限制, 1=串行
	ConditionEvaluation string `json:"condition_evaluation,omitempty"`  // always / on_change
	FailBehavior        string `json:"fail_behavior,omitempty"`         // continue / abort_task
}

// ============================================================================
// 通知系统
// ============================================================================

// NotificationLevel 通知级别。
type NotificationLevel string

const (
	LevelDebug   NotificationLevel = "debug"
	LevelInfo    NotificationLevel = "info"
	LevelWarning NotificationLevel = "warning"
	LevelError   NotificationLevel = "error"
)

// NotificationEvent 通知事件。
type NotificationEvent struct {
	Type       string                 `json:"type"`
	Level      NotificationLevel      `json:"level"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	SourceID   string                 `json:"source_id"`
	SourceType string                 `json:"source_type"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationConfig 通知配置。
type NotificationConfig struct {
	Enabled        bool              `json:"enabled"`
	Level          string            `json:"level,omitempty"`
	Events         []string          `json:"events,omitempty"`
	CustomMessages map[string]string `json:"custom_messages,omitempty"`
}

// GlobalNotificationConfig 全局通知配置。
type GlobalNotificationConfig struct {
	Enabled         bool   `json:"enabled"`
	DefaultLevel    string `json:"default_level"`
	CooldownSeconds int    `json:"cooldown_seconds"`
	MaxPerMinute    int    `json:"max_per_minute"`
}

// ============================================================================
// 日志系统
// ============================================================================

// LoggingConfig 日志配置。
type LoggingConfig struct {
	Enabled *bool         `json:"enabled,omitempty"`   // 默认 true
	Level   string        `json:"level,omitempty"`     // debug / verbose / info / warning / error
	Outputs []string      `json:"outputs,omitempty"`   // file / console / eventlog
	File    *FileLogConfig `json:"file,omitempty"`
	Format  *LogFormat    `json:"format,omitempty"`
}

// FileLogConfig 日志文件配置。
type FileLogConfig struct {
	Path           string `json:"path,omitempty"`
	MaxSizeMB      int    `json:"max_size_mb,omitempty"`
	MaxFiles       int    `json:"max_files,omitempty"`
	MaxAgeDays     int    `json:"max_age_days,omitempty"`
	MaxTotalSizeMB int    `json:"max_total_size_mb,omitempty"`
	Compress       *bool  `json:"compress,omitempty"`
	Encoding       string `json:"encoding,omitempty"`
}

// LogFormat 日志格式配置。
type LogFormat struct {
	Timestamp      string `json:"timestamp,omitempty"`        // iso8601 / unix
	IncludeSource  *bool  `json:"include_source,omitempty"`   // 默认 true
	IncludeCaller  *bool  `json:"include_caller,omitempty"`   // 默认 true
}

// LogEntry 单条日志记录。
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	SourceType string `json:"source_type,omitempty"`
	SourceID   string `json:"source_id,omitempty"`
	Message    string `json:"message"`
	Caller     string `json:"caller,omitempty"`
}

// ============================================================================
// 执行结果
// ============================================================================

// ============================================================================
// 全局配置（config.json）
// ============================================================================

// AppConfig 应用程序全局配置。
type AppConfig struct {
	Port          int                      `json:"port,omitempty"`           // Web 服务端口，默认 19530
	PollInterval  int                      `json:"poll_interval,omitempty"` // 轮询间隔（秒），默认 5
	APIKey        string                   `json:"api_key,omitempty"`       // 全局 API 鉴权密钥
	Logging       *LoggingConfig           `json:"logging,omitempty"`
	Notifications *GlobalNotificationConfig `json:"notifications,omitempty"`
	Timezone      string                   `json:"timezone,omitempty"`      // IANA 时区，默认使用系统时区
	AutoStart     bool                     `json:"auto_start,omitempty"`     // 开机自动启动
	StartMinimized bool                    `json:"start_minimized,omitempty"` // 启动时最小化到托盘
}

// DefaultAppConfig 返回带默认值的应用配置。
func DefaultAppConfig() AppConfig {
	trueVal := true
	return AppConfig{
		Port:         19530,
		PollInterval: 5,
		Timezone:     "",
		Logging: &LoggingConfig{
			Enabled: &trueVal,
			Level:   "info",
			Outputs: []string{"file"},
			File: &FileLogConfig{
				Path:           "logs/app.log",
				MaxSizeMB:      10,
				MaxFiles:       5,
				MaxAgeDays:     30,
				MaxTotalSizeMB: 200,
				Compress:       &trueVal,
				Encoding:       "utf8",
			},
			Format: &LogFormat{
				Timestamp:     "iso8601",
				IncludeSource: &trueVal,
				IncludeCaller: &trueVal,
			},
		},
		Notifications: &GlobalNotificationConfig{
			Enabled:         true,
			DefaultLevel:    "info",
			CooldownSeconds: 3,
			MaxPerMinute:    20,
		},
	}
}

// ============================================================================
// 条件运算符常量
// ============================================================================

// SupportedOperators 所有支持的条件运算符列表。
var SupportedOperators = []string{
	"equals", "not_equals",
	"contains", "not_contains",
	"greater_than", "less_than",
	"greater_equal", "less_equal",
	"matches_regex", "not_matches_regex",
	"starts_with", "ends_with",
	"is_empty", "is_not_empty",
	"exists", "not_exists",
	"length_equals", "length_greater_than", "length_less_than",
	"in", "not_in",
}

// ============================================================================
// Platform 层需要的跨平台类型
// ============================================================================

// 这些类型被 platform 接口和 input/output 插件共同使用。
// 它们不包含任何平台特定的导入。

// WiFiInfo Wi-Fi 信息。
type WiFiInfo struct {
	SSID          string `json:"ssid"`
	BSSID         string `json:"bssid"`
	IsConnected   bool   `json:"is_connected"`
	SignalQuality int    `json:"signal_quality"` // 0-100
}

// PowerInfo 电源信息。
type PowerInfo struct {
	IsCharging    bool `json:"is_charging"`
	BatteryPercent int `json:"battery_percent"`
	ACLineStatus   int `json:"ac_line_status"` // 0=offline, 1=online, 255=unknown
}

// NetworkAdapterInfo 网络适配器信息。
type NetworkAdapterInfo struct {
	Name           string `json:"name"`
	IPv4           string `json:"ipv4"`
	Gateway        string `json:"gateway"`
	SubnetMask     string `json:"subnet_mask,omitempty"`
	IsConnected    bool   `json:"is_connected"`
	ConnectionType string `json:"connection_type"` // ethernet / wifi / ppp / loopback
}

// ProcessInfo 进程信息。
type ProcessInfo struct {
	Name         string  `json:"name"`
	PID          int     `json:"pid"`
	MemoryMB     float64 `json:"memory_mb,omitempty"`
	CPUPercent   float64 `json:"cpu_percent,omitempty"`
	ExecPath     string  `json:"exec_path,omitempty"`
}

// WindowInfo 窗口信息。
type WindowInfo struct {
	Title   string `json:"title"`
	Process string `json:"process"`
}

// SysResInfo 系统资源信息。
type SysResInfo struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	MemoryTotalGB float64 `json:"memory_total_gb,omitempty"`
	MemoryUsedGB  float64 `json:"memory_used_gb,omitempty"`
}

// DiskInfo 磁盘信息。
type DiskInfo struct {
	Drive       string  `json:"drive"`
	FreeGB      float64 `json:"free_gb"`
	TotalGB     float64 `json:"total_gb"`
	UsedPercent float64 `json:"used_percent"`
}

// SessionInfo 会话信息。
type SessionInfo struct {
	IsLocked bool `json:"is_locked"`
}

// ClipboardInfo 剪贴板信息。
type ClipboardInfo struct {
	Text string `json:"text"`
}

// DisplayInfo 显示器信息。
type DisplayInfo struct {
	PrimaryWidth  int `json:"primary_width"`
	PrimaryHeight int `json:"primary_height"`
	MonitorCount  int `json:"monitor_count"`
	DPI           int `json:"dpi"`
}

// OSInfo 操作系统信息。
type OSInfo struct {
	Version      string `json:"version"`
	Build        string `json:"build"`
	ComputerName string `json:"computer_name"`
	UserName     string `json:"user_name"`
}

// CPUInfo CPU 信息。
type CPUInfo struct {
	PhysicalCores int     `json:"physical_cores"`
	LogicalCores  int     `json:"logical_cores"`
	Architecture  string  `json:"architecture"`
	ProcessorName string  `json:"processor_name,omitempty"`
	MaxFrequency  float64 `json:"max_frequency_ghz,omitempty"`
}

// LocaleInfo 区域/语言信息。
type LocaleInfo struct {
	Language         string `json:"language"`
	Region           string `json:"region"`
	Timezone         string `json:"timezone"`
	Is24HourFormat   bool   `json:"is_24hour_format"`
}

// BatteryDetail 电池详细信息。
type BatteryDetail struct {
	VoltageNow      float64 `json:"voltage_now"`
	ChargeRate      float64 `json:"charge_rate_mw"`
	DesignCapacity  float64 `json:"design_capacity_mwh"`
	FullChargeCapacity float64 `json:"full_charge_capacity_mwh"`
	CycleCount      int     `json:"cycle_count"`
	EstimatedTimeRemaining int `json:"estimated_seconds_remaining"`
}

// NetworkTraffic 网络流量统计。
type NetworkTraffic struct {
	BytesSent     uint64 `json:"bytes_sent"`
	BytesReceived uint64 `json:"bytes_received"`
	PacketsSent   uint64 `json:"packets_sent"`
	PacketsReceived uint64 `json:"packets_received"`
}
