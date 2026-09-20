# RuleCraft — Windows 轻量级自动化规则引擎

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org)

RuleCraft 是一个 Windows 后台常驻的轻量级自动化工具，基于 **"输入插件(Sensors) → 决策引擎(Engine) → 输出插件(Actuators)"** 的管道架构。

## 功能特性

- **45+ 个内置输入插件** — 采集 Wi-Fi、电源、网络、进程、窗口、空闲时间、系统资源、磁盘、时间、会话、剪贴板、显示器、音量、CPU、区域、USB、服务状态、蓝牙、电源计划、键盘锁定键、光标、全屏窗口、输入法、主题、待重启、代理、打印机、壁纸、回收站、显卡、公网 IP、启动项、计划任务、Hosts、防火墙、节电模式等，以及 HTTP 请求、网络检测、手动触发、HTTP/TCP/UDP/WebSocket 入站、文件监控
- **40+ 个内置输出插件** — 桌面通知、电源方案切换、阻止休眠、执行程序、音量控制/渐变、静音、亮度调节、锁定工作站、电源操作、壁纸设置、杀进程（PID/名称）、重启资源管理器、打开控制面板、媒体控制、刷新/设置 DNS、深色模式、清空回收站、窗口置顶、进程优先级，以及应用对接：企业微信 / 钉钉 / 飞书 / Discord / Bark / Telegram / Slack / Server酱 / Home Assistant，Webhook/HTTP GET/HTTP POST、延时等待、文件写入、网卡控制、打开文件、窗口控制、截图、剪贴板设置
- **插件按类分组** — 系统类/硬件类/网络类/入站类/文件类/进程类/触发类，插件列表和编辑器中可按类筛选和搜索
- **AST 递归逻辑树** — 支持 AND/OR/NOT 嵌套条件判断，30+ 种运算符（含 between、changed、increased/decreased、is_true/is_false、忽略大小写比较等）
- **触发阈值容错** — 支持条件连续成立 N 次后才触发输出（先阈值滤波再边缘触发，避免漏触发）
- **条件稳定时间** — 支持条件持续成立 N 秒后才触发，避免抖动
- **输出冷却** — 每个输出可配置 cooldown_seconds，避免短时间内重复执行
- **输出延时** — 可添加独立延时插件作为多个输出动作之间的间隔
- **条件级阈值/稳定** — 每个状态条件可单独设置连续次数阈值或持续秒数稳定时间（二选一）
- **并行采集 + 超时** — 输入插件并行采集，单插件超时自动跳过，避免拖垮整个轮询周期
- **内置预设任务模板** — 低电量、磁盘不足、空闲锁屏、夜间降音量、CPU 高负载、Wi-Fi 切换、USB 接入等开箱即用模板（默认禁用，按需启用）
- **Web UI 管理界面** — 仪表盘、任务管理、插件编辑器、插件市场、日志查看、系统设置、使用说明
- **系统托盘** — 后台运行，右键菜单管理（开机自启/修改端口/Web启停/打开Web/退出）
- **全局 Debug API** — 支持 API Key 鉴权
- **插件市场** — 从 GitHub/Gitee 仓库在线拉取社区插件
- **日志系统** — 按任务分文件，自动轮转清理

## 系统要求

- Windows 10 / Windows 11（64 位）
- Go 1.21+（仅开发时需要）

## 快速开始

### 下载预编译版本

从 [Releases](../../releases) 页面下载最新的 `RuleCraft.exe`，双击运行即可。

### 从源码编译

```bash
# 1. 克隆仓库
git clone https://github.com/your-org/rulecraft.git
cd rulecraft

# 2. 编译（隐藏控制台窗口）
go build -ldflags="-H windowsgui" -o RuleCraft.exe .

# 3. 运行
./RuleCraft.exe
```

### 编译选项

| 选项 | 说明 |
|:-----|:------|
| `-H windowsgui` | 隐藏控制台窗口，纯 GUI 模式运行 |
| `-o RuleCraft.exe` | 指定输出文件名 |
| `-ldflags="-s -w"` | 减小二进制体积（去掉调试信息） |

**编译体积优化示例：**

```bash
go build -ldflags="-H windowsgui -s -w" -o RuleCraft.exe .
```

### 命令行参数

```
-port 19530      Web 服务端口（默认 19530）
-config config.json  配置文件路径（默认 config.json）
-interval 5      轮询间隔秒数（默认 5）
-version         显示版本号
```

## 项目结构

```
RuleCraft/
├── main.go                  # 程序入口
├── go.mod / go.sum          # Go 模块定义
│
├── config/config.go         # 核心数据结构
├── plugin/interface.go      # 插件接口 + 注册表
│
├── platform/                # 操作系统抽象层
│   ├── interface.go         # Platform 接口
│   └── platform_windows.go  # Windows Win32 API 实现
│
├── engine/                  # 引擎核心
│   ├── evaluator.go         # AST 递归条件求值器
│   ├── runner.go            # 轮询引擎 + 边缘触发
│   ├── scheduler.go         # 四阶段任务调度器
│   ├── processor_engine.go  # DAG 拓扑排序管道
│   ├── script_runner.go     # 子进程脚本执行器
│   ├── loader.go            # 配置文件加载器
│   ├── notifier.go          # 通知管理器
│   └── logger.go            # 日志系统
│
├── input/                   # 输入插件（每个独立文件，共26个）
│   ├── inputs.go            # 注册入口
│   ├── power_input.go       # 电源传感器
│   ├── wifi_input.go        # Wi-Fi 传感器
│   ├── time_input.go        # 时间传感器
│   ├── network_input.go     # 网络传感器
│   ├── process_input.go     # 进程传感器
│   ├── window_input.go      # 窗口传感器
│   ├── idle_input.go        # 空闲时间传感器
│   ├── sysres_input.go      # 系统资源传感器
│   ├── disk_input.go        # 磁盘传感器
│   ├── session_input.go     # 会话传感器
│   ├── uptime_input.go      # 运行时间传感器
│   ├── clipboard_input.go   # 剪贴板传感器
│   ├── display_input.go     # 显示器传感器
│   ├── volume_input.go      # 音量传感器
│   ├── os_input.go          # 操作系统传感器
│   ├── cpu_input.go         # CPU 传感器
│   ├── locale_input.go      # 区域传感器
│   ├── battery_detail_input.go # 电池详细传感器
│   ├── network_stats_input.go  # 网络流量传感器
│   ├── manual_trigger_input.go # 手动触发传感器
│   ├── http_request_input.go   # HTTP 请求传感器
│   ├── network_detect_input.go # 网络检测传感器
│   ├── http_in_input.go        # HTTP 入站传感器
│   ├── tcp_udp_in_input.go     # TCP/UDP 入站传感器
│   ├── websocket_in_input.go   # WebSocket 入站传感器
│   ├── file_monitor_input.go   # 文件监控传感器
│   └── parser.go               # 入站数据解析器
│
├── output/                  # 输出插件（每个独立文件，共20个）
│   ├── outputs.go           # 注册入口
│   ├── notify_output.go     # 桌面通知
│   ├── power_scheme_output.go # 电源方案切换
│   ├── prevent_sleep_output.go # 阻止休眠
│   ├── exec_output.go       # 执行程序
│   ├── volume_output.go     # 音量控制
│   ├── brightness_output.go # 亮度调节
│   ├── lock_output.go       # 锁定工作站
│   ├── power_action_output.go # 电源操作
│   ├── wallpaper_output.go  # 壁纸设置
│   ├── kill_process_output.go # 杀进程
│   ├── webhook_output.go    # HTTP 请求
│   ├── http_get_output.go   # HTTP GET 请求
│   ├── http_post_output.go  # HTTP POST 请求
│   ├── delay_output.go      # 延时等待
│   ├── file_output.go       # 文件写入
│   ├── netadapter_output.go # 网卡控制
│   ├── open_output.go       # 打开文件/URL
│   ├── window_control_output.go # 窗口控制
│   ├── screenshot_output.go # 截图
│   └── clipboard_set_output.go # 剪贴板设置
│
├── api/                     # Web API
│   ├── handler.go           # RESTful API 处理程序
│   ├── persist.go           # 文件持久化
│   ├── market.go            # 插件市场
│   └── webui/               # 嵌入式 Web UI
│       ├── index.html
│       ├── style.css
│       ├── app.js
│       ├── plugin-editor.js
│       ├── User-Manual.md
│
├── gui/window.go            # 系统托盘（纯托盘模式，无窗口）
│
├── docs/                    # 文档
│   ├── Plugin-Manual.md     # 插件开发手册（英文）
│   ├── User-Manual.md       # 使用手册（英文）
│   ├── 使用手册.md           # 使用手册（中文）
│   └── 插件开发手册.md        # 插件开发手册（中文）
│
├── Input_Plugins/           # 外部输入插件目录
│   └── example_processor/
│       ├── rule.json
│       └── to_lowercase.ps1
│
├── Output_Plugins/          # 外部输出插件目录
├── tasks/                   # 任务配置
├── schemas/                 # JSON Schema
│   ├── rule-v1.json
│   ├── task-conf-v1.json
│   └── task-v1.json
│
├── logo.png                 # 托盘图标源文件
├── build.bat                # Windows 编译脚本
├── config.json              # 配置文件
└── main.go                  # 程序入口
```

## 开发指南

### 添加新的输入插件

1. 在 `input/` 下新建 `.go` 文件
2. 实现 `plugin.InputPlugin` 接口
3. 在 `input/inputs.go` 的 `RegisterBuiltinInputs` 中添加一行

详见 [Plugin-Manual.md](./docs/Plugin-Manual.md)。

### 添加新的输出插件

1. 在 `output/` 下新建 `.go` 文件
2. 实现 `plugin.OutputPlugin` 接口
3. 在 `output/outputs.go` 的 `RegisterBuiltinOutputs` 中添加一行

### 添加新的 Platform 实现（移植到其他 OS）

1. 创建 `platform/platform_linux.go`，实现 `Platform` 接口全部方法
2. 创建 `platform/new_linux.go`，返回 Linux 实现
3. 所有输入/输出插件无需修改

## 技术架构

```
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ 输入插件   │ → │ 任务调度器 │ → │ 条件求值器 │ → │ 输出插件   │
│ (Sensors) │   │ (DAG)    │   │ (AST)    │   │(Actuators)│
└──────────┘   └──────────┘   └──────────┘   └──────────┘
      │              │              │              │
      ▼              ▼              ▼              ▼
  Platform 层   Loader 文件    EdgeTrigger    Platform 层
  (Win32 API)   加载器         边缘触发检测     (Win32 API)
```

## 许可

MIT License
