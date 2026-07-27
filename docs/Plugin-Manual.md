# Plugin-Manual — RuleCraft Plugin Development Guide

RuleCraft supports two plugin systems: **Built-in Plugins** (Go code compiled into the main program) and **JSON Rule Plugins** (extended via JSON configuration + external scripts). This document describes the development methods for both.

---

## Table of Contents

1. [Plugin Architecture Overview](#1-plugin-architecture-overview)
2. [Built-in Input Plugin Development](#2-built-in-input-plugin-development)
3. [Built-in Output Plugin Development](#3-built-in-output-plugin-development)
4. [JSON Rule Plugin Development](#4-json-rule-plugin-development)
5. [Plugin Marketplace Publishing](#5-plugin-marketplace-publishing)
6. [Appendix: Plugin Interface Reference](#6-appendix-plugin-interface-reference)

---

## 1. Plugin Architecture Overview

```
Rule Engine
  │
  ├── Built-in Plugins (Layer 1) — Compiled into the main program
  │   ├── InputPlugin Interface     ← 26 sensors
  │   └── OutputPlugin Interface    ← 20 actuators
  │
  ├── JSON Rule Plugins (Layer 2) — External scripts
  │   ├── rule.json            ← Plugin metadata
  │   ├── {task_id}.conf       ← Task configuration
  │   └── .ps1 / .bat / .exe  ← Executable scripts
  │
  └── Plugin Marketplace (Layer 3) — Remote repositories
      └── Plugin directory in a Git repository
```

**Selection Guide:**

| Scenario | Recommended Approach |
|:---------|:---------------------|
| Need direct Win32 API calls | Built-in plugin |
| High performance required (polling scenarios) | Built-in plugin |
| User-customized data processing logic | JSON rule plugin |
| Need to interact with third-party services | JSON rule plugin |
| Community plugin sharing | Marketplace + JSON rule plugin |

---

## 2. Built-in Input Plugin Development

### 2.1 Interface Definition

```go
// plugin/interface.go
type InputPlugin interface {
    ID() string                              // Globally unique identifier, e.g. "power"
    Name() string                            // Human-readable name
    Collect(ctx *SystemContext) error         // Collect data and write to ctx.States
    IsAvailable() bool                       // Whether the current system supports it
}
```

### 2.2 SystemContext

```go
type SystemContext struct {
    Mu     sync.RWMutex
    States map[string]interface{}   // Global state storage
    Ctx    context.Context          // Supports timeout/cancellation
    Notify NotificationManager      // Notification manager (optional)
}
```

**Common Methods:**

| Method | Description |
|:-------|:------------|
| `ctx.SetState(key, value)` | Write to global state |
| `ctx.GetState(key)` | Read from global state |
| `ctx.GetStateString(key)` | Read a string state |

### 2.3 Complete Example

Create `input/battery_input.go`:

```go
package input

import (
    "rulecraft/platform"
    "rulecraft/plugin"
)

type BatteryInput struct {
    plat platform.Platform
}

func NewBatteryInput(plat platform.Platform) *BatteryInput {
    return &BatteryInput{plat: plat}
}

func (b *BatteryInput) ID() string        { return "battery" }
func (b *BatteryInput) Name() string      { return "Battery Sensor" }
func (b *BatteryInput) IsAvailable() bool { return true }

func (b *BatteryInput) Collect(ctx *plugin.SystemContext) error {
    info, err := b.plat.CollectPowerInfo()
    if err != nil {
        if platform.IsErrNotImplemented(err) {
            ctx.SetState("battery.percent", -1)
            return nil
        }
        return err
    }
    ctx.SetState("battery.percent", info.BatteryPercent)
    ctx.SetState("battery.is_charging", info.IsCharging)
    return nil
}
```

### 2.4 Registering with the Engine

Add one line in `input/inputs.go`'s `RegisterBuiltinInputs`:

```go
func RegisterBuiltinInputs(registry *plugin.Registry, plat platform.Platform) error {
    inputs := []plugin.InputPlugin{
        // ... existing plugins ...
        NewBatteryInput(plat),   // ← Add this line
    }
    // ...
}
```

### 2.5 Input Plugin Conventions

| Field | Requirement |
|:------|:------------|
| State key naming | `{plugin_id}.{field}`, e.g. `power.battery_percent` |
| Error handling | Write default values on platform failure instead of erroring out |
| ErrNotImplemented | Gracefully degrade when the current OS does not support it |
| Type selection | Use `bool` for booleans, `float64` for percentages, `string` for names |

### 2.6 Plugin Configuration Parameters

Some input plugins require additional configuration parameters (such as URL, target address, etc.). These are set via the `params` field in the "Input Data" section of the task editor. Supported plugins:

| Plugin ID | Parameters | Description |
|-----------|------------|-------------|
| `http_request` | `url`, `timeout` | HTTP request target URL and timeout in seconds |
| `network_detect` | `target`, `mode`, `port` | Detection target address, mode (ping/tcp/http), port |
| `file_monitor` | `paths`, `recursive` | Monitor paths (comma-separated) and whether to recurse |
| `http_in` | `port`, `auth_key` | Listen port and auth key (configured in system settings) |
| `tcp_udp_in` | `port`, `proto`, `auth_key` | TCP/UDP listen port and protocol |
| `websocket_in` | `port`, `auth_key` | WebSocket listen port and auth key |

---

## 3. Built-in Output Plugin Development

### 3.1 Interface Definition

```go
type OutputPlugin interface {
    ID() string                              // Globally unique identifier
    Name() string                            // Human-readable name
    Execute(params map[string]interface{}) error  // Execute when condition is met
    Reset(params map[string]interface{}) error    // Undo when condition reverts
    IsAvailable() bool                       // Whether the current system supports it
}
```

### 3.2 Lifecycle

| supports_execute | supports_reset | Behavior |
|:----------------:|:--------------:|:---------|
| true | false | Execute on condition met, no action on revert |
| true | true | Execute on condition met, undo on revert |

Plugins that do not support Reset return `plugin.ErrResetNotSupported`.

### 3.3 Complete Example

Create `output/screenshot_output.go`:

```go
package output

import (
    "fmt"
    "rulecraft/plugin"
)

type ScreenshotOutput struct {
    plat PlatformSubset
}

func NewScreenshotOutput(plat PlatformSubset) *ScreenshotOutput {
    return &ScreenshotOutput{plat: plat}
}

func (s *ScreenshotOutput) ID() string        { return "screenshot" }
func (s *ScreenshotOutput) Name() string      { return "Save Screenshot" }
func (s *ScreenshotOutput) IsAvailable() bool { return true }

// Execute saves a screenshot.
// params: path (string, required): save path
func (s *ScreenshotOutput) Execute(params map[string]interface{}) error {
    path, _ := params["path"].(string)
    if path == "" {
        return fmt.Errorf("screenshot: path is required")
    }
    // Screenshot logic...
    return nil
}

func (s *ScreenshotOutput) Reset(params map[string]interface{}) error {
    return plugin.ErrResetNotSupported
}
```

### 3.4 Registering with the Engine

Add one line in `output/outputs.go`'s `RegisterBuiltinOutputs`.

### 3.5 Using PlatformSubset

Output plugins should not directly depend on all methods of `platform.Platform`. Instead, use the `PlatformSubset` interface to declare only the methods needed:

```go
type PlatformSubset interface {
    SetVolume(level int) error
    SetPowerScheme(scheme string) error
    // Only declare the methods this plugin needs
}
```

---

## 4. JSON Rule Plugin Development

JSON rule plugins are RuleCraft's extension mechanism, allowing users to define custom data processing logic through JSON configuration + external scripts.

### 4.1 Directory Structure

```
rules/
├── {plugin_id}/              # Plugin directory
│   ├── rule.json             # Plugin definition file (required)
│   ├── {executable_file}     # Executable program (required)
│   └── {task_id}.conf        # Task configuration (zero or more)
```

### 4.2 rule.json Specification

```json
{
    "plugin_id": "text_processor",
    "name": "Text Processor",
    "version": "1.0.0",
    "author": "YourName",
    "description": "Converts input text to lowercase",
    "direction": "input",
    "executable": {
        "type": "powershell",
        "path": "rules/text_processor/process.ps1"
    },
    "execution_policy": {
        "timeout_seconds": 30,
        "retry_on_failure": 1
    },
    "functions": {
        "to_lowercase": {
            "name": "To Lowercase",
            "description": "Converts the input string to lowercase",
            "input_params": {
                "value": {
                    "type": "string",
                    "description": "Input text",
                    "required": true
                }
            },
            "output": {
                "result": {
                    "type": "string",
                    "description": "Lowercase conversion result"
                }
            }
        }
    }
}
```

| Field | Required | Description |
|:------|:--------:|:------------|
| `plugin_id` | ✅ | Globally unique identifier, `[a-z][a-z0-9_-]+` |
| `direction` | ❌ | `input` (default, data processing) or `output` (action execution) |
| `executable.type` | ✅ | `powershell` / `batch` / `executable` |
| `executable.path` | ✅ | Executable program path |

### 4.3 {task_id}.conf Specification

```json
{
    "task_id": "task_001",
    "rule_id": "text_processor",
    "source": {
        "type": "from_state",
        "key": "wifi.ssid"
    },
    "pipeline": [
        {
            "function": "to_lowercase",
            "enabled": true,
            "params": {}
        }
    ],
    "output": {
        "mode": "write_state",
        "target_key": "processed.ssid"
    },
    "callback": {
        "on_success": { "action": "continue_rule_engine" },
        "on_failure": { "action": "use_original_value" }
    }
}
```

### 4.4 Script stdin/stdout Protocol

The engine communicates with external scripts through a JSON pipe over stdin/stdout.

**Input (stdin):**

```json
{
    "function_id": "to_lowercase",
    "value": "My_WiFi_SSID",
    "params": {},
    "__task__": {
        "task_id": "task_001",
        "rule_id": "text_processor",
        "processor_id": "ssid_cleaner"
    }
}
```

**Output (stdout):**

```json
{
    "value": "my_wifi_ssid"
}
```

**PowerShell Script Example:**

```powershell
param([Parameter(ValueFromPipeline=$true)]$inputJson)
$data = $inputJson | ConvertFrom-Json
$value = $data.value
$result = $value.ToLower()
Write-Output "{ ""value"": ""$result"" }"
```

**Batch Script Example:**

```batch
@echo off
set /p input=
echo %input% > temp.json
rem ... processing logic ...
echo {"value": "processed"}
```

### 4.5 Output Rule Plugins (direction=output)

Output plugins must declare in `rule.json`:

```json
{
    "plugin_id": "email_sender",
    "direction": "output",
    "lifecycle": {
        "supports_execute": true,
        "supports_reset": false
    }
}
```

The `source` for output plugins is typically `from_stdin`, and parameters are passed by the main program at trigger time.

### 4.6 Environment Variable Injection

The engine automatically sets the following environment variables when calling external scripts:

| Environment Variable | Description |
|:---------------------|:------------|
| `AUTOMATION_TASK_ID` | Current task ID |
| `AUTOMATION_TASK_NAME` | Current task name |
| `AUTOMATION_RULE_ID` | Current rule ID |
| `AUTOMATION_PROCESSOR_ID` | Current processor ID |

### 4.7 Security Constraints

- Path traversal protection: `executable.path` must not contain `..`
- Timeout forced termination: All child processes are bound by `timeout_seconds`
- PowerShell is always run with `-ExecutionPolicy Bypass`
- Sensitive fields in parameters (`password`, `token`, `secret`, `key`) are replaced with `****` in logs

---

## 5. Plugin Marketplace Publishing

### 5.1 Repository Structure

Create a plugin repository on GitHub or Gitee with the following directory structure:

```
your-plugin-repo/
├── Input_Plugins/             # Input plugins
│   └── my_sensor/
│       ├── rule.json
│       └── collect.ps1
│
└── Output_Plugins/            # Output plugins
    └── my_action/
        ├── rule.json
        └── execute.ps1
```

### 5.2 Installation

In the "Plugin Marketplace" page of the RuleCraft Web UI, add the repository URL and install with one click:

```
https://github.com/your-name/your-plugin-repo
```

The system automatically scans available plugins from the `Input_Plugins/` and `Output_Plugins/` directories.

### 5.3 Version Management

The `version` field in `rule.json` follows SemVer 2.0.0. The engine caches installed plugins and does not auto-overwrite. Users need to manually click the "Update" button.

---

## 6. Appendix: Plugin Interface Reference

### InputPlugin Interface

```go
type InputPlugin interface {
    ID() string
    Name() string
    Collect(ctx *SystemContext) error
    IsAvailable() bool
}
```

### OutputPlugin Interface

```go
type OutputPlugin interface {
    ID() string
    Name() string
    Execute(params map[string]interface{}) error
    Reset(params map[string]interface{}) error
    IsAvailable() bool
}
```

### Platform Interface (for built-in plugins)

```go
type Platform interface {
    // Sensors
    CollectWiFiInfo() (*config.WiFiInfo, error)
    CollectPowerInfo() (*config.PowerInfo, error)
    CollectNetworkInfo() ([]config.NetworkAdapterInfo, error)
    ListProcesses() ([]config.ProcessInfo, error)
    GetForegroundWindowInfo() (*config.WindowInfo, error)
    GetIdleSeconds() (uint64, error)
    GetSystemResources() (*config.SysResInfo, error)
    GetDiskInfo(drive string) (*config.DiskInfo, error)
    GetSessionInfo() (*config.SessionInfo, error)
    // Actuators
    PreventSleep(prevent bool) error
    ExecuteProcess(cmd string, args []string, workingDir string, env map[string]string) (int, error)
    SetVolume(level int) error
    SetBrightness(level int) error
    ShowNotification(title, message string, level string) error
    LockWorkstation() error
    PowerAction(action string) error
    SetWallpaper(path string) error
    KillProcess(pid int) error
    SetPowerScheme(scheme string) error
    SetNetworkAdapter(name string, enabled bool) error
}
```

### Registry

```go
registry.RegisterInput(p InputPlugin, description string) error
registry.RegisterOutput(p OutputPlugin, description string) error
registry.GetInput(id string) (InputPlugin, bool)
registry.GetOutput(id string) (OutputPlugin, bool)
registry.ListInputs() []string
registry.ListOutputs() []string
```

### State Key Naming Convention

| Prefix | Usage | Example |
|:-------|:------|:--------|
| `{plugin_id}.` | Plugin raw data | `wifi.ssid`, `power.battery_percent` |
| `processed.` | Processor output | `processed.ssid`, `processed.temperature` |

### Condition Operator Reference

| Operator | Applicable Type | Description |
|:---------|:----------------|:------------|
| `equals` / `not_equals` | Any | Equality/inequality comparison |
| `contains` / `not_contains` | string | Substring containment |
| `greater_than` / `less_than` | number | Numeric comparison |
| `matches_regex` | string | Regular expression match |
| `exists` / `not_exists` | Any | Whether a state key exists |
| `is_empty` / `is_not_empty` | string | Empty string check |
| `in` / `not_in` | Any | Enumeration membership check |
