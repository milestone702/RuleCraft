// RuleCraft Web UI — 国际化翻译字典 (i18n)
// ============================================================================

const I18N = {
  'zh-CN': {
    // 侧边栏导航
    'nav.dashboard': '📊 仪表盘',
    'nav.tasks': '📋 任务管理',
    'nav.plugins': '🔌 插件列表',
    'nav.plugin-editor': '🔧 插件编辑器',
    'nav.plugin-editor-input': '📥 输入插件编写',
    'nav.plugin-editor-output': '📤 输出插件编写',
    'nav.plugin-editor-edit': '📝 已有插件编辑',
    'nav.plugin-editor-exit': '🚪 退出插件编辑器',
    'nav.market': '🛒 插件市场',
    'nav.logs': '📝 日志查看',
    'nav.settings': '⚙️ 系统设置',
    'nav.manual': '📖 使用说明',

    // 页面标题
    'page.dashboard': '📊 仪表盘',
    'page.tasks': '📋 任务管理',
    'page.plugins': '🔌 插件列表',
    'page.market': '🛒 插件市场',
    'page.logs': '📝 日志查看',
    'page.settings': '⚙️ 系统设置',
    'page.manual': '📖 使用说明',

    // 连接状态
    'status.connected': '● 已连接',
    'status.disconnected': '● 已断开',

    // 语言选择
    'lang.switch': '🌐 语言切换',

    // 仪表盘
    'dashboard.active_tasks': '活跃任务',
    'dashboard.rules': '规则数',
    'dashboard.registered_plugins': '已注册插件',
    'dashboard.state_keys': '状态键数',
    'dashboard.current_status': '当前系统状态',
    'dashboard.loading': '正在加载...',
    'dashboard.connection_failed': '连接失败',
    'dashboard.view_processes': '📋 查看进程列表',
    'dashboard.manual_trigger': '手动触发',
    'dashboard.manual_triggered': '任务 <code>{0}</code> 于 {1} 被触发',
    'dashboard.process_list': '进程列表',
    'dashboard.loading_data': '加载中...',

    // 仪表盘状态分组
    'group.power': '🔋 电源',
    'group.wifi': '📶 Wi-Fi',
    'group.network': '🌐 网络',
    'group.window': '🪟 窗口',
    'group.idle': '💤 空闲',
    'group.session': '🔒 会话',
    'group.manual_trigger': '🖱 手动触发',
    'group.processed': '🔧 处理后',
    'group.other': '其他',
    'group.disk': '💾 磁盘',

    // 任务管理
    'tasks.title': '任务列表',
    'tasks.refresh': '🔄 刷新',
    'tasks.new': '➕ 新建任务',
    'tasks.no_tasks': '暂无任务',
    'tasks.edit': '📝 编辑',
    'tasks.export': '📤 导出',
    'tasks.delete': '🗑 删除',
    'tasks.trigger': '▶ 触发',
    'tasks.triggered': '✅ 已触发',
    'tasks.view_log': '📝 查看日志',

    // 任务编辑器
    'editor.title': '📝 编辑任务',
    'editor.save': '💾 保存',
    'editor.close': '✕ 关闭',
    'editor.basic': '基本信息',
    'editor.task_id': '任务 ID',
    'editor.name': '名称',
    'editor.enabled': '启用',
    'editor.description': '描述',
    'editor.input_data': '输入数据（Input Data）',
    'editor.add_ds': '➕ 添加',
    'editor.ds_help': 'ℹ️ 说明',
    'editor.ds_hint': '选择输入插件，然后选该插件提供的数据字段。条件中将仅显示此处添加的字段。',
    'editor.conditions': '触发条件（Conditions）',
    'editor.cond_hint': '拖拽 ↕ 手柄可排序子节点。逻辑组可添加任意子条件，支持多重 AND / OR / NOT 嵌套。',
    'editor.cond_always': '无条件（始终触发）',
    'editor.outputs': '输出动作（Outputs）',
    'editor.add_output': '➕ 添加',
    'editor.no_outputs': '暂无输出动作',
    'editor.no_ds': '暂无数据源',
    'editor.custom_key': '✏️ 自定义',
    'editor.no_match_plugin': '— 无匹配插件 —',
    'editor.select_plugin': '— 选择插件 —',
    'editor.trigger_on_true': '成立时',
    'editor.trigger_on_false': '不成立时',
    'editor.trigger_on_both': '变化时',
    'editor.add_param': '➕ 参数',
    'editor.param_name': '参数名',
    'editor.param_value': '参数值',
    'editor.threshold': '阈值',
    'editor.stabilize': '稳定',

    // 插件列表
    'plugins.registered': '已注册插件',
    'plugins.input': '输入插件',
    'plugins.output': '输出插件',
    'plugins.id': 'ID',
    'plugins.name': '名称',
    'plugins.description': '描述',
    'plugins.category': '类型',
    'plugins.source': '来源',
    'plugins.builtin': '内置',
    'plugins.external': '外部',
    'plugins.search': '🔍 搜索插件...',
    'plugins.filter_all': '全部',
    'plugins.loading': '加载中...',

    // 插件市场
    'market.title': '插件市场',
    'market.refresh': '🔄 刷新',
    'market.hint': '从 Git 仓库（GitHub / Gitee）拉取社区插件。',
    'market.repos': '配置仓库',
    'market.repo_url_placeholder': 'https://github.com/user/plugin-repo',
    'market.add_repo': '添加仓库',
    'market.available': '可用插件',
    'market.no_repos': '尚未配置仓库，输入 URL 后点击"添加仓库"。',
    'market.install': '安装',
    'market.installed': '已安装',
    'market.update': '更新',
    'market.no_plugins': '添加仓库后查看可用插件',

    // 日志
    'logs.title': '日志查看器',
    'logs.task_placeholder': 'task_id（留空=系统日志）',
    'logs.view': '查看',
    'logs.load_failed': '加载失败',
    'logs.select_task': '选择任务后查看日志',

    // 系统设置
    'settings.system': '⚙️ 系统设置',
    'settings.about': 'ℹ️ 关于',
    'settings.basic': '基本设置',
    'settings.interval': '轮询间隔：',
    'settings.notify_level': '通知级别：',
    'settings.autostart': '开机自动启动',
    'settings.web': 'Web 服务',
    'settings.port': '端口号：',
    'settings.port_hint': '修改后需重启生效',
    'settings.web_service': 'Web 服务：',
    'settings.stop_web': '停止 Web 服务',
    'settings.start_web': '启动 Web 服务',
    'settings.net_plugins': '网络插件端口',
    'settings.net_hint': '配置内置网络监听插件的端口和鉴权密钥。修改后需重启生效。',
    'settings.http_in': 'HTTP 入站 (http_in)：',
    'settings.tcp_udp_in': 'TCP/UDP 入站 (tcp_udp_in)：',
    'settings.websocket_in': 'WebSocket 入站 (websocket_in)：',
    'settings.auth_key_placeholder': '鉴权密钥',
    'settings.save_plugin_ports': '💾 保存插件端口配置',
    'settings.api_auth': 'API 鉴权',
    'settings.api_hint': '设置 API Key 后，访问 /api/debug 等接口需在请求头携带 X-API-Key。',
    'settings.api_enable': '启用 API Key 鉴权',
    'settings.api_key': 'API Key：',
    'settings.api_key_placeholder': '输入 API 密钥',
    'settings.save': '保存',
    'settings.refresh': '🔄 刷新配置',
    'settings.save_all': '💾 保存全部设置',
    'settings.save_button': '保存',
    'settings.basic_section': '基本设置',
    'settings.web_section': 'Web 服务',
    'settings.net_section': '网络插件端口',
    'settings.api_section': 'API 鉴权',
    'settings.1s': '1 秒',
    'settings.3s': '3 秒',
    'settings.5s': '5 秒',
    'settings.10s': '10 秒',
    'settings.30s': '30 秒',
    'settings.about_version': '版本：',
    'settings.about_port': '运行端口：',
    'settings.about_input_plugins': '内置输入插件：',
    'settings.about_output_plugins': '内置输出插件：',
    'settings.about_tech_stack': '技术栈：',
    'settings.about_desc': 'RuleCraft 是一个轻量级的 Windows 自动化规则引擎，通过条件-动作（Condition-Action）模式实现系统自动化。支持多种输入传感器和输出执行器，提供 Web 管理界面进行规则配置。',
    'settings.about_oss': '本项目为开源项目，遵循 MIT 许可证。',
    'settings.about_donate': '如果你觉得功能不错，可以请我喝杯咖啡。',
    'settings.about_subtitle': 'Windows 轻量级自动化规则引擎',
    'settings.about_builtin_input': '内置输入插件：',
    'settings.about_builtin_output': '内置输出插件：',
    'settings.about_tech': '技术栈：',

    // 插件编辑器
    'editor.basic_info': '基本信息',
    'editor.basic_desc': '插件的核心标识和描述',
    'editor.plugin_id': '插件 ID',
    'editor.plugin_id_hint': '小写字母开头，仅含 a-z、0-9、下划线、连字符',
    'editor.plugin_id_placeholder': '例: my_plugin',
    'editor.plugin_name': '插件名称',
    'editor.plugin_name_placeholder': '例: 我的插件',
    'editor.version': '版本',
    'editor.version_placeholder': '1.0.0',
    'editor.version_hint': '语义化版本号，如 1.0.0',
    'editor.author': '作者',
    'editor.author_placeholder': '作者名',
    'editor.description_placeholder': '插件功能描述',
    'editor.direction': '方向',
    'editor.direction_input': '输入插件',
    'editor.direction_output': '输出插件',
    'editor.lifecycle': '生命周期',
    'editor.lifecycle_desc': '输出插件的触发行为声明',
    'editor.lifecycle_desc_behavior': '行为描述',
    'editor.lifecycle_placeholder': '描述该插件的行为',
    'editor.supports_execute': '支持 Execute（条件成立时触发）',
    'editor.supports_reset': '支持 Reset（条件恢复时触发）',
    'editor.yes': '是',
    'editor.no': '否',
    'editor.executable': '可执行程序',
    'editor.executable_desc': '插件运行时调用的脚本或程序',
    'editor.exe_type': '类型',
    'editor.exe_path': '路径',
    'editor.exe_path_hint': '相对于工作目录的路径',
    'editor.exe_path_placeholder': 'Input_Plugins/my_plugin/script.ps1',
    'editor.exe_args': '参数',
    'editor.exe_args_placeholder': '命令行参数',
    'editor.exe_workdir': '工作目录',
    'editor.exe_workdir_hint': '留空使用插件目录',
    'editor.script_template': '脚本模板',
    'editor.gen_template': '🔄 生成模板',
    'editor.script_placeholder': '脚本内容将随 rule.json 一起打包到 ZIP 中。\n点击「生成模板」生成框架代码。',
    'editor.exec_policy': '执行策略',
    'editor.exec_policy_desc': '脚本运行时的控制参数',
    'editor.timeout': '超时时间（秒）',
    'editor.retry': '失败重试次数',
    'editor.cooldown': '冷却时间（秒）',
    'editor.functions': '功能清单',
    'editor.functions_desc': '插件提供的具体功能接口',
    'editor.add_func': '➕ 添加功能',
    'editor.new_func': '新功能',
    'editor.no_funcs': '暂无功能。点击「添加功能」添加插件功能接口。',
    'editor.func_id_placeholder': '功能 ID（如 to_lowercase）',
    'editor.delete': '✕ 删除',
    'editor.state_labels': '状态键标签',
    'editor.state_labels_desc': '插件产出的状态键中文描述',
    'editor.add_label': '➕ 添加',
    'editor.label_hint': '格式：{"键名": "中文描述"}，例：{"processed.ssid": "处理后的 WiFi 名称"}',
    'editor.dependencies': '前置依赖',
    'editor.dependencies_desc': '插件运行需要的环境条件',
    'editor.requires_admin': '需要管理员权限',
    'editor.requires_network': '需要网络连接',
    'editor.min_ps_version': '最低 PowerShell 版本',
    'editor.min_ps_placeholder': '例: 5.1',
    'editor.dep_plugins': '依赖的插件 ID（逗号分隔）',
    'editor.dep_plugins_placeholder': 'plugin1, plugin2',
    'editor.tags': '标签',
    'editor.tags_desc': '分类标签',
    'editor.tags_placeholder': '标签1, 标签2, 标签3（逗号分隔）',
    'editor.preview_json': '👁 预览 JSON',
    'editor.save_to_disk': '💾 保存到插件文件夹',
    'editor.export_zip': '📦 导出 ZIP',
    'editor.saved': '✅ 插件已保存到插件文件夹',
    'editor.saving': '⏳ 正在生成压缩包...',
    'editor.exported': '✅ ZIP 已下载',
    'editor.confirm_exit': '确定退出编辑器吗？未保存的修改将丢失。',
    'editor.confirm_delete_func': '确定删除功能 "{0}" 吗？',
    'editor.param_help_title': '参数说明',
    'editor.param_help_desc': '插件说明',
    'editor.param_help_list': '参数列表',
    'editor.param_help_footer': '在插件的参数行中填入键值对即可配置。参数名需与上表完全一致。',
    'editor.no_param_help': '该插件暂无详细说明',
    'editor.no_params': '该插件无需额外参数',
    'editor.required': '必填',
    'editor.optional': '可选',

    // 进程模态框
    'modal.process_title': '进程列表',
    'modal.process_name': '进程名',
    'modal.pid': 'PID',
    'modal.memory': '内存占用',
    'modal.path': '路径',
    'modal.kill': '结束',
    'modal.close': '✕ 关闭',

    // 诊断消息
    'error.connection': '连接失败',
    'error.load': '加载失败',
    'error.save': '保存失败',
    'error.invalid': '数据有误',

    // 关于页面
    'about.version': '版本：',
    'about.port': '运行端口：',
    'about.tech': '技术栈：',
    'about.builtin_input': '内置输入插件：',
    'about.builtin_output': '内置输出插件：',
    'about.tech_value': 'Go + 原生 Win32 API + 嵌入式 Web UI',

    // 预览 JSON
    'preview.title': '📄 插件规则 JSON 预览',
    'preview.close': '✕ 关闭',

    // 任务编辑器
    'editor.task_editor_title': '📝 编辑任务',
    'editor.basic_section': '基本信息',
    'editor.task_id_label': '任务 ID',
    'editor.name_label': '名称',
    'editor.enabled_label': '启用',
    'editor.desc_label': '描述',
    'editor.ds_section': '输入数据（Input Data）',
    'editor.add_btn': '➕ 添加',
    'editor.help_btn': 'ℹ️ 说明',
    'editor.cond_section': '触发条件（Conditions）',
    'editor.output_section': '输出动作（Outputs）',
    'editor.no_ds_text': '暂无数据源',
    'editor.no_output_text': '暂无输出动作',
    'editor.cond_always_text': '无条件（始终触发）',
    'editor.close_btn': '✕ 关闭',
    'editor.save_btn': '💾 保存',
    'editor.select_plugin_title': '选择插件',
    'editor.builtin_not_editable': '（内置插件不可编辑）',
    'editor.select_plugin_prompt': '— 选择一个插件 —',
    'editor.input_type': '📥 输入',
    'editor.output_type': '📤 输出',
    'editor.builtin_hidden': '已隐藏 {0} 个内置插件（不可编辑）',
    'editor.select_from_above': '从上方选择一个已有的插件开始编辑',
  },

  en: {
    // Sidebar navigation
    'nav.dashboard': '📊 Dashboard',
    'nav.tasks': '📋 Task Management',
    'nav.plugins': '🔌 Plugin List',
    'nav.plugin-editor': '🔧 Plugin Editor',
    'nav.plugin-editor-input': '📥 Input Plugin',
    'nav.plugin-editor-output': '📤 Output Plugin',
    'nav.plugin-editor-edit': '📝 Edit Plugin',
    'nav.plugin-editor-exit': '🚪 Exit Editor',
    'nav.market': '🛒 Plugin Marketplace',
    'nav.logs': '📝 Log Viewer',
    'nav.settings': '⚙️ Settings',
    'nav.manual': '📖 User Manual',

    // Page titles
    'page.dashboard': '📊 Dashboard',
    'page.tasks': '📋 Task Management',
    'page.plugins': '🔌 Plugin List',
    'page.market': '🛒 Plugin Marketplace',
    'page.logs': '📝 Log Viewer',
    'page.settings': '⚙️ Settings',
    'page.manual': '📖 User Manual',

    // Connection status
    'status.connected': '● Connected',
    'status.disconnected': '● Disconnected',

    // Language
    'lang.switch': '🌐 Language',

    // Dashboard
    'dashboard.active_tasks': 'Active Tasks',
    'dashboard.rules': 'Rules',
    'dashboard.registered_plugins': 'Registered Plugins',
    'dashboard.state_keys': 'State Keys',
    'dashboard.current_status': 'Current System Status',
    'dashboard.loading': 'Loading...',
    'dashboard.connection_failed': 'Connection failed',
    'dashboard.view_processes': '📋 View Processes',
    'dashboard.manual_trigger': 'Manual Trigger',
    'dashboard.manual_triggered': 'Task <code>{0}</code> triggered at {1}',
    'dashboard.process_list': 'Process List',
    'dashboard.loading_data': 'Loading...',

    // Dashboard state groups
    'group.power': '🔋 Power',
    'group.wifi': '📶 Wi-Fi',
    'group.network': '🌐 Network',
    'group.window': '🪟 Window',
    'group.idle': '💤 Idle',
    'group.session': '🔒 Session',
    'group.manual_trigger': '🖱 Manual Trigger',
    'group.processed': '🔧 Processed',
    'group.other': 'Other',
    'group.disk': '💾 Disk',

    // Task management
    'tasks.title': 'Task List',
    'tasks.refresh': '🔄 Refresh',
    'tasks.new': '➕ New Task',
    'tasks.no_tasks': 'No tasks',
    'tasks.edit': '📝 Edit',
    'tasks.export': '📤 Export',
    'tasks.delete': '🗑 Delete',
    'tasks.trigger': '▶ Trigger',
    'tasks.triggered': '✅ Triggered',
    'tasks.view_log': '📝 View Log',

    // Task editor
    'editor.title': '📝 Edit Task',
    'editor.save': '💾 Save',
    'editor.close': '✕ Close',
    'editor.basic': 'Basic Info',
    'editor.task_id': 'Task ID',
    'editor.name': 'Name',
    'editor.enabled': 'Enabled',
    'editor.description': 'Description',
    'editor.input_data': 'Input Data',
    'editor.add_ds': '➕ Add',
    'editor.ds_help': 'ℹ️ Help',
    'editor.ds_hint': 'Select an input plugin, then choose the data fields it provides. Only the fields added here will appear in conditions.',
    'editor.conditions': 'Conditions',
    'editor.cond_hint': 'Drag ↕ handles to reorder child nodes. Logical groups can hold any number of sub-conditions with nested AND / OR / NOT.',
    'editor.cond_always': 'No condition (always trigger)',
    'editor.outputs': 'Outputs',
    'editor.add_output': '➕ Add',
    'editor.no_outputs': 'No output actions',
    'editor.no_ds': 'No data sources',
    'editor.custom_key': '✏️ Custom',
    'editor.no_match_plugin': '— No matching plugin —',
    'editor.select_plugin': '— Select plugin —',
    'editor.trigger_on_true': 'On True',
    'editor.trigger_on_false': 'On False',
    'editor.trigger_on_both': 'On Change',
    'editor.add_param': '➕ Param',
    'editor.param_name': 'Key',
    'editor.param_value': 'Value',
    'editor.threshold': 'Threshold',
    'editor.stabilize': 'Stabilize',

    // Plugin list
    'plugins.registered': 'Registered Plugins',
    'plugins.input': 'Input Plugins',
    'plugins.output': 'Output Plugins',
    'plugins.id': 'ID',
    'plugins.name': 'Name',
    'plugins.description': 'Description',
    'plugins.category': 'Category',
    'plugins.source': 'Source',
    'plugins.builtin': 'Built-in',
    'plugins.external': 'External',
    'plugins.search': '🔍 Search plugins...',
    'plugins.filter_all': 'All',
    'plugins.loading': 'Loading...',

    // Plugin marketplace
    'market.title': 'Plugin Marketplace',
    'market.refresh': '🔄 Refresh',
    'market.hint': 'Pull community plugins from Git repositories (GitHub / Gitee).',
    'market.repos': 'Configure Repos',
    'market.repo_url_placeholder': 'https://github.com/user/plugin-repo',
    'market.add_repo': 'Add Repo',
    'market.available': 'Available Plugins',
    'market.no_repos': 'No repositories configured. Enter a URL and click "Add Repo".',
    'market.install': 'Install',
    'market.installed': 'Installed',
    'market.update': 'Update',
    'market.no_plugins': 'Add a repository to see available plugins',

    // Logs
    'logs.title': 'Log Viewer',
    'logs.task_placeholder': 'task_id (blank = system log)',
    'logs.view': 'View',
    'logs.load_failed': 'Load failed',
    'logs.select_task': 'Select a task to view logs',

    // Settings
    'settings.system': '⚙️ System Settings',
    'settings.about': 'ℹ️ About',
    'settings.basic': 'Basic Settings',
    'settings.interval': 'Polling Interval:',
    'settings.notify_level': 'Notification Level:',
    'settings.autostart': 'Auto-start on Boot',
    'settings.web': 'Web Service',
    'settings.port': 'Port:',
    'settings.port_hint': 'Requires restart to take effect',
    'settings.web_service': 'Web Service:',
    'settings.stop_web': 'Stop Web Service',
    'settings.start_web': 'Start Web Service',
    'settings.net_plugins': 'Network Plugin Ports',
    'settings.net_hint': 'Configure ports and auth keys for built-in network listener plugins. Requires restart to take effect.',
    'settings.http_in': 'HTTP Inbound (http_in):',
    'settings.tcp_udp_in': 'TCP/UDP Inbound (tcp_udp_in):',
    'settings.websocket_in': 'WebSocket Inbound (websocket_in):',
    'settings.auth_key_placeholder': 'Auth key',
    'settings.save_plugin_ports': '💾 Save Port Settings',
    'settings.api_auth': 'API Authentication',
    'settings.api_hint': 'When API Key is set, endpoints like /api/debug require the X-API-Key header.',
    'settings.api_enable': 'Enable API Key Auth',
    'settings.api_key': 'API Key:',
    'settings.api_key_placeholder': 'Enter API key',
    'settings.save': 'Save',
    'settings.refresh': '🔄 Refresh Config',
    'settings.save_all': '💾 Save All Settings',
    'settings.save_button': 'Save',
    'settings.basic_section': 'Basic Settings',
    'settings.web_section': 'Web Service',
    'settings.net_section': 'Network Plugin Ports',
    'settings.api_section': 'API Auth',
    'settings.1s': '1 sec',
    'settings.3s': '3 sec',
    'settings.5s': '5 sec',
    'settings.10s': '10 sec',
    'settings.30s': '30 sec',
    'settings.about_version': 'Version:',
    'settings.about_port': 'Running Port:',
    'settings.about_input_plugins': 'Built-in Input Plugins:',
    'settings.about_output_plugins': 'Built-in Output Plugins:',
    'settings.about_tech_stack': 'Tech Stack:',
    'settings.about_desc': 'RuleCraft is a lightweight Windows automation rule engine that implements system automation through a Condition-Action pattern. It supports multiple input sensors and output actuators, with a Web management interface for rule configuration.',
    'settings.about_oss': 'This is an open-source project under the MIT License.',
    'settings.about_donate': 'If you find it useful, feel free to buy me a coffee.',
    'settings.about_subtitle': 'Lightweight Windows Automation Rule Engine',
    'settings.about_builtin_input': 'Built-in Input Plugins:',
    'settings.about_builtin_output': 'Built-in Output Plugins:',
    'settings.about_tech': 'Tech Stack:',

    // Plugin editor
    'editor.basic_info': 'Basic Info',
    'editor.basic_desc': 'Core plugin identifier and description',
    'editor.plugin_id': 'Plugin ID',
    'editor.plugin_id_hint': 'Lowercase letter first, only a-z, 0-9, underscore, hyphen',
    'editor.plugin_id_placeholder': 'e.g. my_plugin',
    'editor.plugin_name': 'Plugin Name',
    'editor.plugin_name_placeholder': 'e.g. My Plugin',
    'editor.version': 'Version',
    'editor.version_placeholder': '1.0.0',
    'editor.version_hint': 'Semantic version, e.g. 1.0.0',
    'editor.author': 'Author',
    'editor.author_placeholder': 'Author name',
    'editor.description_placeholder': 'Plugin description',
    'editor.direction': 'Direction',
    'editor.direction_input': 'Input Plugin',
    'editor.direction_output': 'Output Plugin',
    'editor.lifecycle': 'Lifecycle',
    'editor.lifecycle_desc': 'Output plugin trigger behavior declaration',
    'editor.lifecycle_desc_behavior': 'Behavior Description',
    'editor.lifecycle_placeholder': 'Describe plugin behavior',
    'editor.supports_execute': 'Supports Execute (trigger on condition met)',
    'editor.supports_reset': 'Supports Reset (trigger on condition revert)',
    'editor.yes': 'Yes',
    'editor.no': 'No',
    'editor.executable': 'Executable',
    'editor.executable_desc': 'Script or program called at plugin runtime',
    'editor.exe_type': 'Type',
    'editor.exe_path': 'Path',
    'editor.exe_path_hint': 'Relative to working directory',
    'editor.exe_path_placeholder': 'Input_Plugins/my_plugin/script.ps1',
    'editor.exe_args': 'Arguments',
    'editor.exe_args_placeholder': 'CLI arguments',
    'editor.exe_workdir': 'Working Directory',
    'editor.exe_workdir_hint': 'Leave blank to use plugin directory',
    'editor.script_template': 'Script Template',
    'editor.gen_template': '🔄 Generate',
    'editor.script_placeholder': 'Script content will be included with rule.json in the ZIP.\nClick "Generate" to create boilerplate code.',
    'editor.exec_policy': 'Execution Policy',
    'editor.exec_policy_desc': 'Runtime control parameters',
    'editor.timeout': 'Timeout (seconds)',
    'editor.retry': 'Retry Count',
    'editor.cooldown': 'Cooldown (seconds)',
    'editor.functions': 'Functions',
    'editor.functions_desc': 'Plugin function interfaces',
    'editor.add_func': '➕ Add Function',
    'editor.new_func': 'New Function',
    'editor.no_funcs': 'No functions. Click "Add Function" to add one.',
    'editor.func_id_placeholder': 'Function ID (e.g. to_lowercase)',
    'editor.delete': '✕ Delete',
    'editor.state_labels': 'State Key Labels',
    'editor.state_labels_desc': 'Readable descriptions for plugin output state keys',
    'editor.add_label': '➕ Add',
    'editor.label_hint': 'Format: {"key": "description"}, e.g. {"processed.ssid": "Processed WiFi name"}',
    'editor.dependencies': 'Dependencies',
    'editor.dependencies_desc': 'Environment requirements for the plugin',
    'editor.requires_admin': 'Requires Admin Rights',
    'editor.requires_network': 'Requires Network',
    'editor.min_ps_version': 'Min PowerShell Version',
    'editor.min_ps_placeholder': 'e.g. 5.1',
    'editor.dep_plugins': 'Dependent Plugin IDs (comma-separated)',
    'editor.dep_plugins_placeholder': 'plugin1, plugin2',
    'editor.tags': 'Tags',
    'editor.tags_desc': 'Classification tags',
    'editor.tags_placeholder': 'tag1, tag2, tag3 (comma-separated)',
    'editor.preview_json': '👁 Preview JSON',
    'editor.save_to_disk': '💾 Save to Plugin Folder',
    'editor.export_zip': '📦 Export ZIP',
    'editor.saved': '✅ Plugin saved to plugin folder',
    'editor.saving': '⏳ Generating archive...',
    'editor.exported': '✅ ZIP downloaded',
    'editor.confirm_exit': 'Are you sure you want to exit? Unsaved changes will be lost.',
    'editor.confirm_delete_func': 'Are you sure you want to delete function "{0}"?',
    'editor.param_help_title': 'Parameter Reference',
    'editor.param_help_desc': 'Plugin Description',
    'editor.param_help_list': 'Parameters',
    'editor.param_help_footer': 'Configure parameters by filling in key-value pairs. Parameter names must match exactly.',
    'editor.no_param_help': 'No detailed help for this plugin',
    'editor.no_params': 'This plugin requires no additional parameters',
    'editor.required': 'Required',
    'editor.optional': 'Optional',

    // Process modal
    'modal.process_title': 'Process List',
    'modal.process_name': 'Process Name',
    'modal.pid': 'PID',
    'modal.memory': 'Memory',
    'modal.path': 'Path',
    'modal.kill': 'Kill',
    'modal.close': '✕ Close',

    // Diagnostic messages
    'error.connection': 'Connection failed',
    'error.load': 'Load failed',
    'error.save': 'Save failed',
    'error.invalid': 'Invalid data',

    // About page
    'about.version': 'Version:',
    'about.port': 'Running Port:',
    'about.tech': 'Tech Stack:',
    'about.builtin_input': 'Built-in Input Plugins:',
    'about.builtin_output': 'Built-in Output Plugins:',
    'about.tech_value': 'Go + Native Win32 API + Embedded Web UI',

    // Preview JSON
    'preview.title': '📄 Plugin JSON Preview',
    'preview.close': '✕ Close',

    // Task editor
    'editor.task_editor_title': '📝 Edit Task',
    'editor.basic_section': 'Basic Info',
    'editor.task_id_label': 'Task ID',
    'editor.name_label': 'Name',
    'editor.enabled_label': 'Enabled',
    'editor.desc_label': 'Description',
    'editor.ds_section': 'Input Data',
    'editor.add_btn': '➕ Add',
    'editor.help_btn': 'ℹ️ Help',
    'editor.cond_section': 'Conditions',
    'editor.output_section': 'Outputs',
    'editor.no_ds_text': 'No data sources',
    'editor.no_output_text': 'No output actions',
    'editor.cond_always_text': 'Always (no condition)',
    'editor.close_btn': '✕ Close',
    'editor.save_btn': '💾 Save',
    'editor.select_plugin_title': 'Select Plugin',
    'editor.builtin_not_editable': '(built-in plugins not editable)',
    'editor.select_plugin_prompt': '— Select a plugin —',
    'editor.input_type': '📥 Input',
    'editor.output_type': '📤 Output',
    'editor.builtin_hidden': 'Hidden {0} built-in plugins (not editable)',
    'editor.select_from_above': 'Select an existing plugin from above to start editing',
  }
};

// ============================================================================
// I18n 工具函数
// ============================================================================

// 当前语言（从 localStorage 读取，默认为浏览器语言）
let _currentLang = localStorage.getItem('rulecraft_lang') ||
  (navigator.language && navigator.language.startsWith('zh') ? 'zh-CN' : 'en');

function getLang() {
  return _currentLang;
}

function setLang(lang) {
  _currentLang = lang;
  localStorage.setItem('rulecraft_lang', lang);
  document.documentElement.lang = lang;
  // 重新应用翻译
  applyI18n();
}

// 获取翻译文本，支持 {0}, {1} 参数替换
function t(key, ...args) {
  const dict = I18N[_currentLang] || I18N.en;
  let text = dict[key];
  if (text === undefined) {
    // 回退到英文
    text = I18N.en[key];
  }
  if (text === undefined) {
    return key; // 找不到则返回 key 本身
  }
  // 替换 {0}, {1} 等占位符
  if (args.length > 0) {
    args.forEach((arg, i) => {
      text = text.replace(new RegExp(`\\{${i}\\}`, 'g'), arg);
    });
  }
  return text;
}

// 应用翻译到所有 data-i18n 属性的元素
function applyI18n() {
  // 更新所有 data-i18n 元素
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.dataset.i18n;
    const text = t(key);
    if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') {
      el.placeholder = el.dataset.i18nPlaceholder || text;
    } else {
      el.innerHTML = text;
    }
  });

  // 更新所有 data-i18n-title 元素
  document.querySelectorAll('[data-i18n-title]').forEach(el => {
    el.title = t(el.dataset.i18nTitle);
  });

  // 更新所有 data-i18n-placeholder 元素（不覆盖已有的 data-i18n）
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    if (el.placeholder !== undefined) {
      el.placeholder = t(el.dataset.i18nPlaceholder);
    }
  });

  // 更新语言选择器下拉框的文字
  const langSel = document.getElementById('lang-selector');
  if (langSel) {
    langSel.title = t('lang.switch');
  }

  // 触发自定义事件，让其他模块知道语言变了
  document.dispatchEvent(new CustomEvent('langchange', { detail: { lang: _currentLang } }));
}
