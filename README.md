# RuleCraft — A Lightweight Windows Automation Rule Engine

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org)

RuleCraft is a lightweight, background-running automation tool for Windows, built on a **"Input Plugins (Sensors) → Decision Engine (Engine) → Output Plugins (Actuators)"** pipeline architecture.

## Features

- **20+ Built-in Input Plugins** — Collects Wi-Fi, power, network, processes, windows, idle time, system resources, disk, time, session, clipboard, display, volume, CPU, locale and other system states, plus HTTP requests, network detection, manual triggers, HTTP/TCP/UDP/WebSocket inbound, and file monitoring
- **20+ Built-in Output Plugins** — Desktop notifications, power scheme switching, sleep prevention, program execution, volume control, brightness adjustment, workstation lock, power actions, wallpaper setting, process killing, Webhook/HTTP GET/HTTP POST, delay waiting, file writing, network adapter control, file/URL opening, window control, screenshot, clipboard setting
- **Plugin Categorization** — System/Hardware/Network/Inbound/File/Process/Trigger categories, filterable and searchable in the plugin list and editor
- **AST Recursive Logic Tree** — Supports AND/OR/NOT nested condition evaluation with 22 operators
- **Trigger Threshold Tolerance** — Supports triggering only after a condition holds true for N consecutive polls (the first N-1 times are logged only)
- **Condition Stabilization Time** — Supports triggering only after a condition has held steady for N seconds, preventing jitter
- **Output Delay** — A standalone delay plugin can be added as an interval between multiple output actions
- **Per-condition Threshold/Stabilization** — Each state condition can independently set a consecutive-count threshold or a duration-based stabilization time (mutually exclusive)
- **Web UI Management Panel** — Dashboard, task management, plugin editor, plugin marketplace, log viewer, system settings, user manual
- **System Tray** — Runs in the background, right-click menu management (auto-start on boot / change port / start/stop Web / open Web / exit)
- **Global Debug API** — Supports API Key authentication
- **Plugin Marketplace** — Pull community plugins from GitHub/Gitee repositories online
- **Logging System** — Per-task log files, automatic rotation and cleanup

## System Requirements

- Windows 10 / Windows 11 (64-bit)
- Go 1.21+ (development only)

## Quick Start

### Download Pre-built Binary

Download the latest `RuleCraft.exe` from the [Releases](../../releases) page and double-click to run.

### Build from Source

```bash
# 1. Clone the repository
git clone https://github.com/your-org/rulecraft.git
cd rulecraft

# 2. Build (hide console window)
go build -ldflags="-H windowsgui" -o RuleCraft.exe .

# 3. Run
./RuleCraft.exe
```

### Build Options

| Option | Description |
|:-------|:------------|
| `-H windowsgui` | Hide the console window, run in pure GUI mode |
| `-o RuleCraft.exe` | Output file name |
| `-ldflags="-s -w"` | Reduce binary size (strip debug information) |

**Optimized Build Example:**

```bash
go build -ldflags="-H windowsgui -s -w" -o RuleCraft.exe .
```

### Command-line Arguments

```
-port 19530       Web service port (default 19530)
-config config.json  Configuration file path (default config.json)
-interval 5      Polling interval in seconds (default 5)
-version         Display version number
```

## Project Structure

```
RuleCraft/
├── main.go                  # Program entry point
├── go.mod / go.sum          # Go module definition
│
├── config/config.go         # Core data structures
├── plugin/interface.go      # Plugin interfaces + registry
│
├── platform/                # Operating system abstraction layer
│   ├── interface.go         # Platform interface
│   └── platform_windows.go  # Windows Win32 API implementation
│
├── engine/                  # Engine core
│   ├── evaluator.go         # AST recursive condition evaluator
│   ├── runner.go            # Polling engine + edge triggering
│   ├── scheduler.go         # Four-stage task scheduler
│   ├── processor_engine.go  # DAG topological sort pipeline
│   ├── script_runner.go     # Subprocess script executor
│   ├── loader.go            # Configuration file loader
│   ├── notifier.go          # Notification manager
│   └── logger.go            # Logging system
│
├── input/                   # Input plugins (each in its own file, 26 total)
│   ├── inputs.go            # Registration entry point
│   ├── power_input.go       # Power sensor
│   ├── wifi_input.go        # Wi-Fi sensor
│   ├── time_input.go        # Time sensor
│   ├── network_input.go     # Network sensor
│   ├── process_input.go     # Process sensor
│   ├── window_input.go      # Window sensor
│   ├── idle_input.go        # Idle time sensor
│   ├── sysres_input.go      # System resource sensor
│   ├── disk_input.go        # Disk sensor
│   ├── session_input.go     # Session sensor
│   ├── uptime_input.go      # Uptime sensor
│   ├── clipboard_input.go   # Clipboard sensor
│   ├── display_input.go     # Display sensor
│   ├── volume_input.go      # Volume sensor
│   ├── os_input.go          # Operating system sensor
│   ├── cpu_input.go         # CPU sensor
│   ├── locale_input.go      # Locale sensor
│   ├── battery_detail_input.go # Battery detail sensor
│   ├── network_stats_input.go  # Network traffic sensor
│   ├── manual_trigger_input.go # Manual trigger sensor
│   ├── http_request_input.go   # HTTP request sensor
│   ├── network_detect_input.go # Network detection sensor
│   ├── http_in_input.go        # HTTP inbound sensor
│   ├── tcp_udp_in_input.go     # TCP/UDP inbound sensor
│   ├── websocket_in_input.go   # WebSocket inbound sensor
│   ├── file_monitor_input.go   # File monitoring sensor
│   └── parser.go               # Inbound data parser
│
├── output/                  # Output plugins (each in its own file, 20 total)
│   ├── outputs.go           # Registration entry point
│   ├── notify_output.go     # Desktop notification
│   ├── power_scheme_output.go # Power scheme switching
│   ├── prevent_sleep_output.go # Sleep prevention
│   ├── exec_output.go       # Program execution
│   ├── volume_output.go     # Volume control
│   ├── brightness_output.go # Brightness adjustment
│   ├── lock_output.go       # Workstation lock
│   ├── power_action_output.go # Power actions
│   ├── wallpaper_output.go  # Wallpaper setting
│   ├── kill_process_output.go # Process killing
│   ├── webhook_output.go    # HTTP request
│   ├── http_get_output.go   # HTTP GET request
│   ├── http_post_output.go  # HTTP POST request
│   ├── delay_output.go      # Delay wait
│   ├── file_output.go       # File writing
│   ├── netadapter_output.go # Network adapter control
│   ├── open_output.go       # Open file/URL
│   ├── window_control_output.go # Window control
│   ├── screenshot_output.go # Screenshot
│   └── clipboard_set_output.go # Clipboard setting
│
├── api/                     # Web API
│   ├── handler.go           # RESTful API handlers
│   ├── persist.go           # File persistence
│   ├── market.go            # Plugin marketplace
│   └── webui/               # Embedded Web UI
│       ├── index.html
│       ├── style.css
│       ├── app.js
│       ├── plugin-editor.js
│       ├── User-Manual.md
│
├── gui/window.go            # System tray (tray-only, no window)
│
├── docs/                    # Documentation
│   ├── Plugin-Manual.md     # Plugin development guide (English)
│   ├── User-Manual.md       # User manual (English)
│   ├── 使用手册.md           # User manual (Chinese)
│   └── 插件开发手册.md        # Plugin development guide (Chinese)
│
├── Input_Plugins/           # External input plugin directory
│   └── example_processor/
│       ├── rule.json
│       └── to_lowercase.ps1
│
├── Output_Plugins/          # External output plugin directory
├── tasks/                   # Task configurations
├── schemas/                 # JSON Schemas
│   ├── rule-v1.json
│   ├── task-conf-v1.json
│   └── task-v1.json
│
├── logo.png                 # Tray icon source file
├── build.bat                # Windows build script
├── config.json              # Configuration file
└── main.go                  # Program entry point
```

## Development Guide

### Adding a New Input Plugin

1. Create a new `.go` file in `input/`
2. Implement the `plugin.InputPlugin` interface
3. Add one line in `input/inputs.go`'s `RegisterBuiltinInputs`

See [Plugin-Manual.md](./docs/Plugin-Manual.md) for details.

### Adding a New Output Plugin

1. Create a new `.go` file in `output/`
2. Implement the `plugin.OutputPlugin` interface
3. Add one line in `output/outputs.go`'s `RegisterBuiltinOutputs`

### Adding a New Platform Implementation (Porting to another OS)

1. Create `platform/platform_linux.go`, implementing all methods of the `Platform` interface
2. Create `platform/new_linux.go`, returning the Linux implementation
3. No modifications needed for any input/output plugins

## Technical Architecture

```
┌────────────┐   ┌────────────┐   ┌────────────┐   ┌────────────┐
│ Input      │ → │ Task       │ → │ Condition  │ → │ Output     │
│ Plugins    │   │ Scheduler  │   │ Evaluator  │   │ Plugins    │
│ (Sensors)  │   │ (DAG)      │   │ (AST)      │   │(Actuators) │
└────────────┘   └────────────┘   └────────────┘   └────────────┘
      │                │                │                │
      ▼                ▼                ▼                ▼
  Platform Layer   Loader File     EdgeTrigger     Platform Layer
  (Win32 API)      Loader          Detection        (Win32 API)
```

## License

MIT License
