# RuleCraft 使用手册

RuleCraft 是一款 Windows 后台常驻的自动化规则引擎。它可以监控你的电脑状态（Wi-Fi、电量、CPU 等），根据你设定的规则自动执行动作（切换电源方案、发通知、执行程序等）。

---

## 目录

1. [安装与运行](#1-安装与运行)
2. [界面导览](#2-界面导览)
3. [系统托盘](#3-系统托盘)
4. [创建任务](#4-创建任务)
5. [仪表盘](#5-仪表盘)
6. [插件管理](#6-插件管理)
7. [插件市场](#7-插件市场)
8. [日志与设置](#8-日志与设置)
9. [常见问题](#9-常见问题)

---

## 1. 安装与运行

### 下载

从 [Releases](../../releases) 页面下载 `RuleCraft.exe`。

### 首次运行

双击 `RuleCraft.exe` 启动，程序会在系统托盘（任务栏右下角）显示图标 ⚙️。在浏览器打开 `http://127.0.0.1:19530` 进入 Web 管理面板。

### 开机自启

在 Web 管理面板 → **系统设置** 中勾选"开机自动启动"，或右键系统托盘图标 → **开机自动启动**。RuleCraft 会自动写入注册表 `HKCU\...\Run`。

### 关闭行为

- 右键系统托盘图标 → **退出程序**
- 直接关闭浏览器标签页不会退出程序

---

## 2. 界面导览

Web 管理界面默认地址：`http://127.0.0.1:19530`

| 侧边栏 | 主区域 |
|:-------|:-------|
| ⚙️ RuleCraft | **📊 仪表盘** |
| 📊 仪表盘 | 活跃任务：0  规则数：0  插件数：0  状态键数：0 |
| 📋 任务管理 | ⏰ 当前时间   📊 CPU / 内存 |
| 🔌 插件列表 | |
| 🔧 插件编辑器 | 🔋 电源   📶 Wi-Fi   🌐 网络 |
| 🛒 插件市场 | 💾 磁盘（饼图） |
| 📝 日志查看 | |
| ⚙️ 系统设置 | |
| 📖 使用说明 | |

### 页面说明

| 页面 | 功能 |
|:-----|:------|
| **仪表盘** | 当前系统状态概览，实时刷新，含手动触发高亮卡片 |
| **任务管理** | 创建和管理自动化任务，支持手动触发 |
| **插件列表** | 查看已注册的输入/输出插件，支持按类型/来源筛选 |
| **插件编辑器** | 创建和编辑输入/输出插件，导出 ZIP |
| **插件市场** | 从 Git 仓库下载社区插件 |
| **日志查看** | 按任务查看运行日志 |
| **系统设置** | 轮询间隔、API Key、端口、网络插件配置 |
| **使用说明** | 内嵌使用手册 |

---

## 3. 系统托盘

RuleCraft 运行时在系统托盘中显示图标 ⚙️：

- **位置**：任务栏右下角（时钟旁边）
- **左键单击/双击**：在浏览器打开 Web 管理界面
- **右键单击**：弹出菜单，包含以下选项：
  - 开机自动启动（勾选标记）
  - 修改端口（当前: 19530）
  - 停止/启动 Web 服务
  - 打开 Web 界面
  - 退出程序

---

## 4. 创建任务

### 什么是任务？

任务是 RuleCraft 的核心概念。一条完整的自动化规则由以下四阶段组成：

```
数据采集 → 数据处理 → 条件判断 → 执行动作
```

**示例：办公环境自动节电策略**

```
条件：连接到公司 WiFi + 电池电量 < 30% + 温度 > 30°C
动作：切换为省电模式 + 桌面通知
```

### 创建步骤

1. 在 Web UI 中点击 **任务管理** → **新建任务**
2. 输入任务名称
3. 编辑 `task.json` 配置（详见任务配置格式）
4. 保存后引擎会自动加载

### task.json 配置示例

```json
{
  "task_id": "power_saver",
  "name": "办公自动节电",
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

### 条件运算符

| 运算符 | 示例 | 说明 |
|:-------|:-----|:------|
| `equals` | `"operator": "equals", "value": "company"` | 字符串相等 |
| `contains` | `"operator": "contains", "value": "wifi"` | 包含子串 |
| `greater_than` | `"operator": "greater_than", "value": 30` | 数字大于 |
| `less_than` | `"operator": "less_than", "value": 50` | 数字小于 |
| `matches_regex` | `"operator": "matches_regex", "value": "^win.*"` | 正则匹配 |
| `exists` | `"operator": "exists"` | 状态键是否存在 |

### 冷却时间

在输出中设置 `cooldown_seconds` 可以避免规则被频繁触发：

```json
"conditions": { "cooldown_seconds": 600 }
```

以上配置表示条件在 10 分钟内不会重复触发。

### 条件级阈值与稳定时间

每个状态条件可以单独设置**阈值**（连续 N 次轮询成立后才算真）或**稳定时间**（持续成立 N 秒后才算真），二选一：

![条件阈值/稳定](images/condition-threshold.png)

- **阈值**：适合快速抖动的场景，如 CPU 使用率短暂波动时不触发
- **稳定**：适合需要持续确认的场景，如 Wi-Fi 断开后等待 10 秒再执行动作

在条件行的末尾选择「阈值」或「稳定」并输入数值即可。

### 输出延时插件

在输出动作中添加 `delay` 插件可以在多个输出之间插入等待间隔。例如：

```
[notify 发送通知] → [delay 等待5秒] → [http_post 调用API]
```

该方式通知后等待 5 秒再调用 HTTP 接口，适用于需要间隔执行的场景。

---

## 5. 仪表盘

仪表盘实时显示当前系统的各项状态：

### 状态分组

| 分组 | 图标 | 数据来源 |
|:-----|:-----|:---------|
| 电源 | 🔋 | 电池电量、充电状态 |
| Wi-Fi | 📶 | SSID、信号质量 |
| 网络 | 🌐 | IP 地址、网关 |
| 进程 | ⚙️ | 进程数量、列表 |
| 窗口 | 🪟 | 前台窗口信息 |
| 空闲 | 💤 | 用户空闲秒数 |
| 系统资源 | 📊 | CPU、内存使用率 |
| 磁盘 | 💾 | 各分区使用情况（饼图） |
| 会话 | 🔒 | 锁屏状态 |

### 实时更新

- **时间**：每秒更新
- **CPU/内存**：每秒从引擎拉取最新值
- **其他状态**：30 秒刷新一次

### 进程查看

点击"查看进程列表"按钮弹出模态框，可以看到所有运行中的进程详情：
- 进程名
- PID（进程 ID）
- 内存占用
- 可执行文件路径
- **结束按钮**：点击可强制终止进程

---

## 6. 插件管理

### 内置输入插件一览

RuleCraft 自带了 26 个输入插件，程序启动时自动注册。每个插件采集特定的系统状态数据，数据以 `插件ID.字段名` 的格式存入全局状态。

| 插件 ID | 名称 | 输出状态键 | 说明 |
|---------|------|-----------|------|
| `power` | 电源传感器 | `power.is_charging`, `power.battery_percent`, `power.ac_line_status` | 电池电量和充电状态 |
| `wifi` | Wi-Fi 传感器 | `wifi.ssid`, `wifi.bssid`, `wifi.is_connected`, `wifi.signal_quality` | 当前 Wi-Fi 连接信息 |
| `time` | 时间传感器 | `time.hour`, `time.minute`, `time.weekday`, `time.iso8601` | 系统时间和日期 |
| `network` | 网络传感器 | `network.ipv4`, `network.gateway`, `network.is_connected`, `network.connection_type` | 网络适配器信息 |
| `process` | 进程传感器 | `process.count`, `process.names`, `process.list_json` | 运行中的进程列表 |
| `window` | 窗口传感器 | `window.foreground.title`, `window.foreground.process` | 前台窗口信息 |
| `idle` | 空闲传感器 | `idle.seconds`, `idle.is_idle` | 用户空闲时间 |
| `sysres` | 系统资源传感器 | `sysres.cpu_percent`, `sysres.memory_percent`, `sysres.memory_used_gb` | CPU 和内存使用率 |
| `disk` | 磁盘传感器 | `disk.C.free_gb`, `disk.D.used_percent` | 各磁盘分区使用情况 |
| `session` | 会话传感器 | `session.is_locked` | 锁屏状态 |
| `clipboard` | 剪贴板传感器 | `clipboard.text` | 剪贴板文本内容 |
| `display` | 显示器传感器 | `display.primary_width`, `display.primary_height`, `display.monitor_count` | 显示器分辨率信息 |
| `uptime` | 运行时间传感器 | `uptime.seconds`, `uptime.minutes`, `uptime.hours` | 系统运行时间 |
| `os` | 操作系统传感器 | `os.version`, `os.build`, `os.computer_name`, `os.user_name` | 系统版本和用户信息 |
| `cpu` | CPU 传感器 | `cpu.physical_cores`, `cpu.logical_cores`, `cpu.architecture` | CPU 硬件信息 |
| `locale` | 区域传感器 | `locale.language`, `locale.region`, `locale.timezone` | 系统和区域设置 |
| `volume` | 音量传感器 | `volume.master_percent`, `volume.is_muted` | 系统音量 |
| `battery_detail` | 电池详细传感器 | `battery_detail.voltage_now`, `battery_detail.cycle_count` | 电池详细信息 |
| `network_stats` | 网络流量传感器 | `network_stats.bytes_sent`, `network_stats.bytes_received` | 网络流量统计 |
| `manual_trigger` | 手动触发传感器 | `manual_trigger.triggered`, `manual_trigger.at`, `manual_trigger.task_id` | 通过 Web UI 按钮手动触发 |
| `http_request` | HTTP 请求传感器 | `http_request.status_code`, `http_request.body`, `http_request.url` | 定时发起 HTTP GET 请求 |
| `network_detect` | 网络检测传感器 | `network_detect.reachable`, `network_detect.http_status`, `network_detect.latency_ms` | Ping/TCP 端口/HTTP 状态码检测 |
| `http_in` | HTTP 入站传感器 | `http_in._raw`, `http_in.解析后的字段` | 监听 HTTP POST 请求（默认端口 19531） |
| `tcp_udp_in` | TCP/UDP 入站传感器 | `tcp_udp_in.data`, `tcp_udp_in.addr` | 监听 TCP/UDP 数据（默认端口 19532） |
| `websocket_in` | WebSocket 入站传感器 | `websocket_in.data`, `websocket_in._raw` | 监听 WebSocket 连接（默认端口 19533） |
| `file_monitor` | 文件监控传感器 | `file_monitor.exists.路径`, `file_monitor.total_files`, `file_monitor.changes` | 文件/文件夹变更监控 |

### 内置输出插件一览

RuleCraft 自带了 20 个输出插件，程序启动时自动注册。

| 插件 ID | 名称 | 参数 | 说明 |
|---------|------|------|------|
| `notify` | 桌面通知 | `title`, `message`, `level` | 发送 Windows 桌面通知 |
| `power_scheme` | 电源方案 | `scheme` (balanced/power_saver/high_performance) | 切换 Windows 电源方案 |
| `prevent_sleep` | 阻止睡眠 | `enable` (true/false) | 阻止/允许系统进入睡眠 |
| `exec` | 执行命令 | `command`, `args`, `timeout` | 执行外部命令或程序 |
| `volume` | 音量控制 | `level` (0-100), `mute` | 设置系统音量或静音 |
| `brightness` | 亮度控制 | `level` (0-100) | 设置屏幕亮度 |
| `lock` | 锁屏 | 无参数 | 锁定工作站 |
| `power_action` | 电源操作 | `action` (shutdown/restart/logoff/sleep/hibernate) | 执行系统电源操作 |
| `wallpaper` | 壁纸设置 | `path` (图片路径) | 更换桌面壁纸 |
| `kill_process` | 结束进程 | `pid` 或 `name` | 终止指定进程 |
| `webhook` | Webhook | `url`, `method`, `body`, `headers` | 发送 HTTP 请求到指定 URL |
| `http_get` | HTTP GET | `url`, `headers`, `timeout` | 发送 HTTP GET 请求 |
| `http_post` | HTTP POST | `url`, `body`, `content_type`, `headers`, `timeout` | 发送 HTTP POST 请求 |
| `delay` | 延时等待 | `seconds` | 等待指定秒数（用作输出动作间隔） |
| `file` | 文件操作 | `path`, `content`, `action` (write/append/delete) | 读写文件 |
| `netadapter` | 网络适配器 | `name`, `action` (enable/disable) | 启用/禁用网络适配器 |
| `open` | 打开 | `target` (文件/URL/程序) | 打开文件、网址或程序 |
| `window_control` | 窗口控制 | `title`, `action` (close/minimize/maximize/restore) | 控制窗口状态 |
| `screenshot` | 截屏 | `path` (保存路径) | 截图保存到文件 |
| `clipboard_set` | 设置剪贴板 | `text` | 写入文本到剪贴板 |

### 状态键使用示例

在任务编辑器的"输入数据"中，选择插件后可直接选择或输入完整的状态键名：

```
示例：判断 Wi-Fi 是否连接
状态键：wifi.is_connected
条件：equals → true

示例：检测 HTTP 入站数据
状态键：http_in.temp
条件：greater_than → 30
```

### JSON 规则插件

用户可以通过在 `rules/` 目录下放置 `rule.json` + 脚本文件来自定义插件。详见 [Plugin-Manual.md](./Plugin-Manual.md) 或 [插件开发手册.md](./插件开发手册.md)。

---

## 7. 插件市场

插件市场允许你从 GitHub 或 Gitee 仓库下载社区贡献的插件。

### 添加仓库

1. 进入 **插件市场** 页面
2. 输入 Git 仓库 URL
3. 点击 **添加仓库**

系统会自动扫描仓库中 `Input_Plugins/` 和 `Output_Plugins/` 目录下的可用插件。

### 安装插件

1. 在可用插件列表中点击 **安装**
2. 系统会自动下载 `rule.json` 和可执行脚本到 `rules/` 目录
3. 重启任务或引擎后生效

---

## 8. 日志与设置

### 日志查看

在 **日志查看** 页面输入任务 ID 可以查看该任务的运行日志。留空查看系统日志。

日志文件存储在 `logs/` 目录：
```
logs/
├── app.log              # 系统日志
├── tasks/
│   ├── task_001.log     # 任务日志
│   └── task_002.log
└── plugins/             # 插件日志（可选）
```

日志自动轮转：文件达到 10MB 自动分割，最多保留 5 个历史文件，超过 30 天自动清理。

### 系统设置

在 Web UI 的 **系统设置** 页面可以配置：

- **轮询间隔**：引擎采集数据的频率（1-30 秒）
- **通知级别**：最低通知级别（Debug / Info / Warning / Error）
- **开机自动启动**：勾选后开机自动运行
- **Web 服务端口**：修改 Web 管理界面端口（需重启生效）
- **API Key 鉴权**：启用后调用 `/api/debug` 等接口需携带 X-API-Key 头
- **网络插件端口**：配置 http_in / tcp_udp_in / websocket_in 的监听端口和鉴权密钥

---

## 9. 常见问题

### Q: RuleCraft 开机不自动启动？

请确保在系统设置中勾选了"开机自动启动"。如果仍不生效，可以手动检查：

```
注册表位置：HKCU\Software\Microsoft\Windows\CurrentVersion\Run
键名：RuleCraft
值：你的 RuleCraft.exe 完整路径
```

### Q: 怎么退出程序？

- 系统托盘图标 → 右键 → **退出程序**

### Q: Web 界面打不开？

1. 确认 RuleCraft 正在运行（托盘中有图标）
2. 检查端口是否被占用：`http://127.0.0.1:19530`

### Q: 图标显示不正确？

Windows 会缓存图标，如果 RuleCraft 的图标不显示，请尝试：
1. **重命名 .exe 文件**（Windows 会当作新文件重新加载图标）
2. **重启 Windows Explorer**（任务管理器 → 右键 Explorer → 重新启动）
3. **清除图标缓存**：运行 `ie4uinit.exe -ClearIconCache`

### Q: 如何更换程序图标？

```bash
# 1. 替换 logo.png
cp my-icon.png logo.png

# 2. 重新生成图标资源
go run ./tools/genicon/main.go

# 3. 重新编译
go build -ldflags="-H windowsgui" -o RuleCraft.exe .
```

### Q: 日志文件太大？

在 `config.json` 中调整日志配置：

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

### Q: 规则没有触发？

1. 检查 `task.json` 中的 `enabled` 是否为 `true`
2. 检查数据源的状态键名称是否正确
3. 检查输出的 `trigger_on` 是否匹配（`true` 表示条件成立时触发）
4. 查看任务日志定位问题
