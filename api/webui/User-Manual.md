# RuleCraft User Manual

RuleCraft is a Windows background automation rule engine. It monitors your computer's state (Wi-Fi, battery, CPU, etc.) and automatically executes actions based on rules you define (switching power schemes, sending notifications, running programs, etc.).

---

## Table of Contents

1. [Installation and Running](#1-installation-and-running)
2. [Interface Overview](#2-interface-overview)
3. [System Tray](#3-system-tray)
4. [Creating Tasks](#4-creating-tasks)
5. [Dashboard](#5-dashboard)
6. [Plugin Management](#6-plugin-management)
7. [Plugin Marketplace](#7-plugin-marketplace)
8. [Logs and Settings](#8-logs-and-settings)
9. [FAQ](#9-faq)

---

## 1. Installation and Running

### Download

Download `RuleCraft.exe` from the [Releases](../../releases) page.

### First Run

Double-click `RuleCraft.exe` to start. The program will display an icon ⚙️ in the system tray (bottom-right of the taskbar). Open `http://127.0.0.1:19530` in your browser to access the Web management panel.

### Auto-start on Boot

Go to **System Settings** in the Web management panel and check "Auto-start on boot", or right-click the system tray icon → **Auto-start on boot**. RuleCraft will automatically write to the registry `HKCU\...\Run`.

### Closing Behavior

- Right-click the system tray icon → **Exit**
- Closing the browser tab does not exit the program

---

## 2. Interface Overview

Default Web management interface address: `http://127.0.0.1:19530`

| Sidebar | Main Area |
|:--------|:----------|
| ⚙️ RuleCraft | **📊 Dashboard** |
| 📊 Dashboard | Active Tasks: 0   Rules: 0   Plugins: 0   State Keys: 0 |
| 📋 Task Management | ⏰ Current Time   📊 CPU / Memory |
| 🔌 Plugin List | |
| 🔧 Plugin Editor | 🔋 Power   📶 Wi-Fi   🌐 Network |
| 🛒 Plugin Marketplace | 💾 Disk (pie chart) |
| 📝 Log Viewer | |
| ⚙️ System Settings | |
| 📖 User Manual | |

### Page Descriptions

| Page | Function |
|:-----|:---------|
| **Dashboard** | System status overview — real-time refresh with manual trigger highlight cards |
| **Task Management** | Create and manage automation tasks, supports manual triggering |
| **Plugin List** | View registered input/output plugins, filterable by type/source |
| **Plugin Editor** | Create and edit input/output plugins, export as ZIP |
| **Plugin Marketplace** | Download community plugins from Git repositories |
| **Log Viewer** | View runtime logs by task |
| **System Settings** | Polling interval, API Key, port, network plugin configuration |
| **User Manual** | Embedded user manual |

---

## 3. System Tray

When running, RuleCraft displays an icon ⚙️ in the system tray:

- **Location**: Bottom-right of the taskbar (next to the clock)
- **Left-click / Double-click**: Opens the Web management interface in a browser
- **Right-click**: Opens a menu with the following options:
  - Auto-start on boot (checkmark indicator)
  - Change port (current: 19530)
  - Stop/Start Web service
  - Open Web interface
  - Exit

---

## 4. Creating Tasks

### What is a Task?

A task is the core concept of RuleCraft. A complete automation rule consists of the following four stages:

```
Data Collection → Data Processing → Condition Evaluation → Action Execution
```

**Example: Office Auto Power-Saving Strategy**

```
Condition: Connected to company Wi-Fi + Battery < 30% + Temperature > 30°C
Action: Switch to power saver mode + Desktop notification
```

### Creation Steps

1. In the Web UI, click **Task Management** → **New Task**
2. Enter a task name
3. Edit the `task.json` configuration (see Task Configuration Format below)
4. Save — the engine will automatically load it

### task.json Configuration Example

```json
{
  "task_id": "power_saver",
  "name": "Office Auto Power Saving",
  "enabled": true,
  "data_sources": [
    { "id": "wifi", "state_key": "wifi.ssid" },
    { "id": "battery", "state_key": "power.battery_percent" }
  ],
  "processors": [],
  "condition": {
    "logical_operator": "AND",
    "conditions": [
      { "type": "state", "state_key": "wifi.ssid", "operator": "contains", "value": "company" },
      { "type": "state", "state_key": "power.battery_percent", "operator": "less_than", "value": 30 }
    ]
  },
  "outputs": [
    {
      "id": "save_power",
      "type": "builtin",
      "plugin_id": "power_scheme",
      "params": { "scheme": "power_saver" },
      "trigger_on": "true",
      "conditions": { "cooldown_seconds": 600 }
    }
  ]
}
```

### Condition Operators

| Operator | Example | Description |
|:---------|:--------|:------------|
| `equals` | `"operator": "equals", "value": "company"` | String equality |
| `contains` | `"operator": "contains", "value": "wifi"` | Substring containment |
| `greater_than` | `"operator": "greater_than", "value": 30` | Numeric greater than |
| `less_than` | `"operator": "less_than", "value": 50` | Numeric less than |
| `matches_regex` | `"operator": "matches_regex", "value": "^win.*"` | Regular expression match |
| `exists` | `"operator": "exists"` | Whether a state key exists |

### Cooldown

Set `cooldown_seconds` in the output to prevent the rule from triggering too frequently:

```json
"conditions": { "cooldown_seconds": 600 }
```

The above configuration means the condition will not re-trigger within 10 minutes.

### Per-condition Threshold & Stabilization

Each state condition can independently set a **threshold** (count of consecutive polling cycles before evaluating as true) or **stabilization time** (duration in seconds the condition must hold steady before evaluating as true), mutually exclusive:

![Condition Threshold/Stabilization](images/condition-threshold.png)

- **Threshold**: Suitable for quickly fluctuating scenarios, e.g., avoid triggering on brief CPU spikes
- **Stabilization**: Suitable for scenarios requiring sustained confirmation, e.g., wait 10 seconds after Wi-Fi disconnects before executing an action

Select "Threshold" or "Stabilization" at the end of the condition row and enter a value.

### Output Delay Plugin

Add a `delay` plugin between output actions to insert a wait interval. For example:

```
[notify send notification] → [delay wait 5 seconds] → [http_post call API]
```

This sends a notification, waits 5 seconds, then calls the HTTP API — useful for scenarios requiring spaced execution.

---

## 5. Dashboard

The dashboard displays the current system status in real time:

### Status Groups

| Group | Icon | Data Source |
|:------|:-----|:------------|
| Power | 🔋 | Battery level, charging status |
| Wi-Fi | 📶 | SSID, signal quality |
| Network | 🌐 | IP address, gateway |
| Processes | ⚙️ | Process count, list |
| Window | 🪟 | Foreground window info |
| Idle | 💤 | User idle seconds |
| System Resources | 📊 | CPU, memory usage |
| Disk | 💾 | Partition usage (pie chart) |
| Session | 🔒 | Lock screen status |

### Real-time Updates

- **Time**: Updates every second
- **CPU/Memory**: Pulled from the engine every second
- **Other states**: Refreshed every 30 seconds

### Process Viewer

Click the "View Process List" button to open a modal showing details of all running processes:
- Process name
- PID (Process ID)
- Memory usage
- Executable file path
- **Kill button**: Click to forcefully terminate a process

---

## 6. Plugin Management

### Built-in Input Plugins Overview

RuleCraft ships with 26 input plugins that are automatically registered at startup. Each plugin collects specific system state data, stored in the global state as `pluginID.fieldName`.

| Plugin ID | Name | Output State Keys | Description |
|-----------|------|-------------------|-------------|
| `power` | Power Sensor | `power.is_charging`, `power.battery_percent`, `power.ac_line_status` | Battery level and charging status |
| `wifi` | Wi-Fi Sensor | `wifi.ssid`, `wifi.bssid`, `wifi.is_connected`, `wifi.signal_quality` | Current Wi-Fi connection info |
| `time` | Time Sensor | `time.hour`, `time.minute`, `time.weekday`, `time.iso8601` | System time and date |
| `network` | Network Sensor | `network.ipv4`, `network.gateway`, `network.is_connected`, `network.connection_type` | Network adapter info |
| `process` | Process Sensor | `process.count`, `process.names`, `process.list_json` | Running process list |
| `window` | Window Sensor | `window.foreground.title`, `window.foreground.process` | Foreground window info |
| `idle` | Idle Sensor | `idle.seconds`, `idle.is_idle` | User idle time |
| `sysres` | System Resource Sensor | `sysres.cpu_percent`, `sysres.memory_percent`, `sysres.memory_used_gb` | CPU and memory usage |
| `disk` | Disk Sensor | `disk.C.free_gb`, `disk.D.used_percent` | Disk partition usage |
| `session` | Session Sensor | `session.is_locked` | Lock screen status |
| `clipboard` | Clipboard Sensor | `clipboard.text` | Clipboard text content |
| `display` | Display Sensor | `display.primary_width`, `display.primary_height`, `display.monitor_count` | Display resolution info |
| `uptime` | Uptime Sensor | `uptime.seconds`, `uptime.minutes`, `uptime.hours` | System uptime |
| `os` | OS Sensor | `os.version`, `os.build`, `os.computer_name`, `os.user_name` | OS version and user info |
| `cpu` | CPU Sensor | `cpu.physical_cores`, `cpu.logical_cores`, `cpu.architecture` | CPU hardware info |
| `locale` | Locale Sensor | `locale.language`, `locale.region`, `locale.timezone` | System and regional settings |
| `volume` | Volume Sensor | `volume.master_percent`, `volume.is_muted` | System volume |
| `battery_detail` | Battery Detail Sensor | `battery_detail.voltage_now`, `battery_detail.cycle_count` | Detailed battery info |
| `network_stats` | Network Traffic Sensor | `network_stats.bytes_sent`, `network_stats.bytes_received` | Network traffic statistics |
| `manual_trigger` | Manual Trigger Sensor | `manual_trigger.triggered`, `manual_trigger.at`, `manual_trigger.task_id` | Manual trigger via Web UI button |
| `http_request` | HTTP Request Sensor | `http_request.status_code`, `http_request.body`, `http_request.url` | Periodic HTTP GET requests |
| `network_detect` | Network Detection Sensor | `network_detect.reachable`, `network_detect.http_status`, `network_detect.latency_ms` | Ping/TCP port/HTTP status detection |
| `http_in` | HTTP Inbound Sensor | `http_in._raw`, `http_in.parsed_fields` | Listens for HTTP POST requests (default port 19531) |
| `tcp_udp_in` | TCP/UDP Inbound Sensor | `tcp_udp_in.data`, `tcp_udp_in.addr` | Listens for TCP/UDP data (default port 19532) |
| `websocket_in` | WebSocket Inbound Sensor | `websocket_in.data`, `websocket_in._raw` | Listens for WebSocket connections (default port 19533) |
| `file_monitor` | File Monitor Sensor | `file_monitor.exists.path`, `file_monitor.total_files`, `file_monitor.changes` | File/folder change monitoring |

### Built-in Output Plugins Overview

RuleCraft ships with 20 output plugins that are automatically registered at startup.

| Plugin ID | Name | Parameters | Description |
|-----------|------|------------|-------------|
| `notify` | Desktop Notification | `title`, `message`, `level` | Sends Windows desktop notifications |
| `power_scheme` | Power Scheme | `scheme` (balanced/power_saver/high_performance) | Switches Windows power scheme |
| `prevent_sleep` | Prevent Sleep | `enable` (true/false) | Prevents/allows system sleep |
| `exec` | Execute Command | `command`, `args`, `timeout` | Executes an external command or program |
| `volume` | Volume Control | `level` (0-100), `mute` | Sets system volume or mute |
| `brightness` | Brightness Control | `level` (0-100) | Sets screen brightness |
| `lock` | Lock Screen | None | Locks the workstation |
| `power_action` | Power Action | `action` (shutdown/restart/logoff/sleep/hibernate) | Performs system power operations |
| `wallpaper` | Wallpaper Setting | `path` (image path) | Changes desktop wallpaper |
| `kill_process` | Kill Process | `pid` or `name` | Terminates a specified process |
| `webhook` | Webhook | `url`, `method`, `body`, `headers` | Sends an HTTP request to a specified URL |
| `http_get` | HTTP GET | `url`, `headers`, `timeout` | Sends an HTTP GET request |
| `http_post` | HTTP POST | `url`, `body`, `content_type`, `headers`, `timeout` | Sends an HTTP POST request |
| `delay` | Delay Wait | `seconds` | Waits for a specified number of seconds (used as output action interval) |
| `file` | File Operation | `path`, `content`, `action` (write/append/delete) | Reads/writes files |
| `netadapter` | Network Adapter | `name`, `action` (enable/disable) | Enables/disables a network adapter |
| `open` | Open | `target` (file/URL/program) | Opens a file, URL, or program |
| `window_control` | Window Control | `title`, `action` (close/minimize/maximize/restore) | Controls window state |
| `screenshot` | Screenshot | `path` (save path) | Takes a screenshot and saves to file |
| `clipboard_set` | Set Clipboard | `text` | Writes text to clipboard |

### State Key Usage Examples

In the task editor's "Input Data" section, after selecting a plugin you can directly select or type the full state key name:

```
Example: Check if Wi-Fi is connected
State key: wifi.is_connected
Condition: equals → true

Example: Check HTTP inbound data
State key: http_in.temp
Condition: greater_than → 30
```

### JSON Rule Plugins

Users can customize plugins by placing `rule.json` + script files in the `rules/` directory. See [Plugin-Manual.md](./Plugin-Manual.md) for details.

---

## 7. Plugin Marketplace

The plugin marketplace allows you to download community-contributed plugins from GitHub or Gitee repositories.

### Adding a Repository

1. Go to the **Plugin Marketplace** page
2. Enter a Git repository URL
3. Click **Add Repository**

The system automatically scans the `Input_Plugins/` and `Output_Plugins/` directories in the repository for available plugins.

### Installing Plugins

1. Click **Install** on an available plugin
2. The system automatically downloads `rule.json` and executable scripts to the `rules/` directory
3. Restart the task or engine to apply changes

---

## 8. Logs and Settings

### Log Viewer

Enter a task ID on the **Log Viewer** page to view that task's runtime logs. Leave it blank to view system logs.

Log files are stored in the `logs/` directory:
```
logs/
├── app.log              # System log
├── tasks/
│   ├── task_001.log     # Task log
│   └── task_002.log
└── plugins/             # Plugin logs (optional)
```

Log auto-rotation: files are split when they reach 10 MB, keeping up to 5 historical files; files older than 30 days are automatically cleaned up.

### System Settings

Configure the following on the **System Settings** page in the Web UI:

- **Polling Interval**: Frequency of engine data collection (1-30 seconds)
- **Notification Level**: Minimum notification level (Debug / Info / Warning / Error)
- **Auto-start on Boot**: Check to run automatically at boot
- **Web Service Port**: Change the Web management interface port (requires restart)
- **API Key Authentication**: When enabled, endpoints like `/api/debug` require the X-API-Key header
- **Network Plugin Ports**: Configure listen ports and auth keys for http_in / tcp_udp_in / websocket_in

---

## 9. FAQ

### Q: RuleCraft doesn't auto-start on boot?

Make sure "Auto-start on boot" is checked in System Settings. If it still doesn't work, check manually:

```
Registry location: HKCU\Software\Microsoft\Windows\CurrentVersion\Run
Key name: RuleCraft
Value: Full path to your RuleCraft.exe
```

### Q: How do I exit the program?

- System tray icon → Right-click → **Exit**

### Q: The Web interface won't open?

1. Make sure RuleCraft is running (icon in the system tray)
2. Check if the port is in use: `http://127.0.0.1:19530`

### Q: The icon doesn't display correctly?

Windows caches icons. If the RuleCraft icon is not displaying, try:
1. **Rename the .exe file** (Windows will treat it as a new file and reload the icon)
2. **Restart Windows Explorer** (Task Manager → right-click Explorer → Restart)
3. **Clear icon cache**: Run `ie4uinit.exe -ClearIconCache`

### Q: How do I change the program icon?

```bash
# 1. Replace logo.png
cp my-icon.png logo.png

# 2. Regenerate icon resources
go run ./tools/genicon/main.go

# 3. Rebuild
go build -ldflags="-H windowsgui" -o RuleCraft.exe .
```

### Q: Log files are too large?

Adjust the logging configuration in `config.json`:

```json
{
  "logging": {
    "level": "info",
    "file": {
      "max_size_mb": 5,
      "max_files": 3,
      "max_age_days": 7
    }
  }
}
```

### Q: A rule is not triggering?

1. Check that `enabled` in `task.json` is set to `true`
2. Verify the state key names in the data sources are correct
3. Check that the output's `trigger_on` matches (`true` means trigger when the condition holds)
4. Check the task logs to identify the issue
