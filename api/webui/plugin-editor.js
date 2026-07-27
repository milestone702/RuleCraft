// RuleCraft 插件编辑器 — 前端逻辑
// ============================================================================
// 编辑器状态
// ============================================================================

let _editorData = {
  mode: '',          // 'input-new' | 'output-new' | 'edit-existing'
  plugin: null,      // 当前编辑的 RuleDefinition
  originalId: '',    // 编辑已有插件时的原始 plugin_id
  plugins: [],       // 缓存所有插件列表（用于编辑模式）
};

// ============================================================================
// 编辑器页面切换
// ============================================================================

function switchEditorPage(pageId) {
  // 显示表单
  const formContainer = document.getElementById('editor-form-container');
  formContainer.style.display = 'block';
  document.getElementById('editor-actions').style.display = 'flex';

  _editorData.mode = pageId;
  _editorData.originalId = '';

  if (pageId === 'input-new') {
    _editorData.plugin = createEmptyPlugin('input');
    renderEditorForm();
  } else if (pageId === 'output-new') {
    _editorData.plugin = createEmptyPlugin('output');
    renderEditorForm();
  } else if (pageId === 'edit-existing') {
    renderPluginSelector();
  }
}

// ============================================================================
// 创建空插件模板
// ============================================================================

function createEmptyPlugin(direction) {
  const now = Date.now();
  return {
    plugin_id: direction === 'input' ? 'new_input_' + now.toString(36) : 'new_output_' + now.toString(36),
    name: direction === 'input' ? 'New Input Plugin' : 'New Output Plugin',
    version: '1.0.0',
    author: '',
    description: '',
    direction: direction,
    executable: {
      type: 'powershell',
      path: '',
      args: '',
      working_dir: '',
    },
    execution_policy: {
      timeout_seconds: 30,
      retry_on_failure: 1,
      cooldown_seconds: 0,
    },
    functions: {},
    state_key_labels: {},
    tags: [],
  };
}

// ============================================================================
// 渲染编辑表单
// ============================================================================

function renderEditorForm() {
  const container = document.getElementById('editor-form-container');
  const p = _editorData.plugin;
  const isOutput = p.direction === 'output';
  const mode = _editorData.mode;

  let html = '';

  // 如果是编辑已有插件，显示选择器
  if (mode === 'edit-existing') {
    html += renderPluginSelectorInline();
  }

  // ===== 基本信息 =====
  html += `<div class="editor-section-card">
    <h4>${t('editor.basic_info')||'基本信息'} <span class="section-desc">${t('editor.basic_desc')||'插件的核心标识和描述'}</span></h4>
    <div class="editor-grid">
      <div class="editor-field">
        <label>${t('editor.plugin_id')||'插件 ID'} <span class="required">*</span></label>
        <input type="text" id="ef-plugin_id" value="${escHtml(p.plugin_id)}"
          placeholder="${t('editor.plugin_id_placeholder')||'例: my_plugin'}" pattern="[a-z][a-z0-9_-]+"
          onchange="updateField('plugin_id', this.value)"
          ${mode === 'edit-existing' ? 'readonly class="input-readonly"' : ''} />
        <span class="field-hint">${t('editor.plugin_id_hint')||'小写字母开头，仅含 a-z、0-9、下划线、连字符'}</span>
      </div>
      <div class="editor-field">
        <label>${t('editor.plugin_name')||'插件名称'} <span class="required">*</span></label>
        <input type="text" id="ef-name" value="${escHtml(p.name)}"
          placeholder="${t('editor.plugin_name_placeholder')||'例: 我的插件'}" onchange="updateField('name', this.value)" />
      </div>
      <div class="editor-field">
        <label>${t('editor.version')||'版本'} <span class="required">*</span></label>
        <input type="text" id="ef-version" value="${escHtml(p.version)}"
          placeholder="1.0.0" pattern="\\d+\\.\\d+\\.\\d+"
          onchange="updateField('version', this.value)" />
        <span class="field-hint">${t('editor.version_hint')||'语义化版本号，如 1.0.0'}</span>
      </div>
      <div class="editor-field">
        <label>${t('editor.author')||'作者'}</label>
        <input type="text" id="ef-author" value="${escHtml(p.author || '')}"
          placeholder="${t('editor.author_placeholder')||'作者名'}" onchange="updateField('author', this.value)" />
      </div>
      <div class="editor-field full-width">
        <label>${t('editor.description')||'描述'}</label>
        <textarea id="ef-description" rows="2" placeholder="${t('editor.description_placeholder')||'插件功能描述'}"
          onchange="updateField('description', this.value)">${escHtml(p.description || '')}</textarea>
      </div>
      <div class="editor-field">
        <label>${t('editor.direction')||'方向'}</label>
        <select id="ef-direction" onchange="updateDirection(this.value)">
          <option value="input" ${p.direction==='input'?'selected':''}>${t('editor.direction_input')||'输入插件'}</option>
          <option value="output" ${p.direction==='output'?'selected':''}>${t('editor.direction_output')||'输出插件'}</option>
        </select>
      </div>
    </div>
  </div>`;

  // ===== 输出插件生命周期（仅 output）=====
  if (isOutput) {
    const lc = p.lifecycle || {};
    html += `<div class="editor-section-card">
      <h4>${t('editor.lifecycle')||'生命周期'} <span class="section-desc">${t('editor.lifecycle_desc')||'输出插件的触发行为声明'}</span></h4>
      <div class="editor-grid">
        <div class="editor-field">
          <label>${t('editor.supports_execute')||'支持 Execute（条件成立时触发）'}</label>
          <select id="ef-lifecycle-execute" onchange="updateLifecycleField('supports_execute', this.value === 'true')">
            <option value="true" ${lc.supports_execute !== false?'selected':''}>${t('editor.yes')||'是'}</option>
            <option value="false" ${lc.supports_execute===false?'selected':''}>${t('editor.no')||'否'}</option>
          </select>
        </div>
        <div class="editor-field">
          <label>${t('editor.supports_reset')||'支持 Reset（条件恢复时触发）'}</label>
          <select id="ef-lifecycle-reset" onchange="updateLifecycleField('supports_reset', this.value === 'true')">
            <option value="true" ${lc.supports_reset?'selected':''}>${t('editor.yes')||'是'}</option>
            <option value="false" ${!lc.supports_reset?'selected':''}>${t('editor.no')||'否'}</option>
          </select>
        </div>
        <div class="editor-field full-width">
          <label>${t('editor.lifecycle_desc_behavior')||'行为描述'}</label>
          <input type="text" id="ef-lifecycle-desc" value="${escHtml(lc.description || '')}"
            placeholder="${t('editor.lifecycle_placeholder')||'描述该插件的行为'}" onchange="updateLifecycleField('description', this.value)" />
        </div>
      </div>
    </div>`;
  }

  // ===== 可执行文件定义 =====
  const exe = p.executable || {};
  html += `<div class="editor-section-card">
    <h4>${t('editor.executable')||'可执行程序'} <span class="section-desc">${t('editor.executable_desc')||'插件运行时调用的脚本或程序'}</span></h4>
    <div class="editor-grid">
      <div class="editor-field">
        <label>${t('editor.exe_type')||'类型'} <span class="required">*</span></label>
        <select id="ef-exe-type" onchange="updateExeField('type', this.value); updateScriptTemplate();">
          <option value="powershell" ${exe.type==='powershell'?'selected':''}>PowerShell (.ps1)</option>
          <option value="batch" ${exe.type==='batch'?'selected':''}>Batch (.bat)</option>
          <option value="executable" ${exe.type==='executable'?'selected':''}>Executable (.exe)</option>
        </select>
      </div>
      <div class="editor-field">
        <label>${t('editor.exe_path')||'路径'} <span class="required">*</span></label>
        <input type="text" id="ef-exe-path" value="${escHtml(exe.path || '')}"
          placeholder="Input_Plugins/my_plugin/script.ps1"
          onchange="updateExeField('path', this.value)" />
        <span class="field-hint">${t('editor.exe_path_hint')||'相对于工作目录的路径'}</span>
      </div>
      <div class="editor-field">
        <label>${t('editor.exe_args')||'参数'}</label>
        <input type="text" id="ef-exe-args" value="${escHtml(exe.args || '')}"
          placeholder="${t('editor.exe_args_placeholder')||'命令行参数'}" onchange="updateExeField('args', this.value)" />
      </div>
      <div class="editor-field">
        <label>${t('editor.exe_workdir')||'工作目录'}</label>
        <input type="text" id="ef-exe-workdir" value="${escHtml(exe.working_dir || '')}"
          placeholder="${t('editor.exe_workdir_hint')||'留空使用插件目录'}" onchange="updateExeField('working_dir', this.value)" />
      </div>
    </div>
    <div class="editor-script-area" id="script-area">
      <h5 style="margin-top:8px;font-size:13px;color:#555">${t('editor.script_template')||'脚本模板'} <button class="btn btn-xs" onclick="generateScriptTemplate()" style="background:#3498db" data-i18n-editor="editor.gen_template">🔄 生成模板</button></h5>
      <textarea id="ef-script-content" rows="8" style="width:100%;font-family:monospace;font-size:12px;padding:6px;border:1px solid #ddd;border-radius:4px"
        placeholder="${t('editor.script_placeholder')||'脚本内容将随 rule.json 一起打包到 ZIP 中。&#10;点击「生成模板」生成框架代码。'}"
        onchange="_editorData.scriptContent = this.value">${_editorData.scriptContent || ''}</textarea>
    </div>
  </div>`;

  // ===== 执行策略 =====
  const ep = p.execution_policy || {};
  html += `<div class="editor-section-card">
    <h4>${t('editor.exec_policy')||'执行策略'} <span class="section-desc">${t('editor.exec_policy_desc')||'脚本运行时的控制参数'}</span></h4>
    <div class="editor-grid">
      <div class="editor-field">
        <label>${t('editor.timeout')||'超时时间（秒）'}</label>
        <input type="number" id="ef-ep-timeout" value="${ep.timeout_seconds || 30}"
          min="1" max="300" onchange="updatePolicyField('timeout_seconds', parseInt(this.value) || 30)" />
      </div>
      <div class="editor-field">
        <label>${t('editor.retry')||'失败重试次数'}</label>
        <input type="number" id="ef-ep-retry" value="${ep.retry_on_failure || 1}"
          min="0" max="10" onchange="updatePolicyField('retry_on_failure', parseInt(this.value) || 0)" />
      </div>
      <div class="editor-field">
        <label>${t('editor.cooldown')||'冷却时间（秒）'}</label>
        <input type="number" id="ef-ep-cooldown" value="${ep.cooldown_seconds || 0}"
          min="0" max="3600" onchange="updatePolicyField('cooldown_seconds', parseInt(this.value) || 0)" />
      </div>
    </div>
  </div>`;

  // ===== 功能清单 =====
  html += `<div class="editor-section-card">
    <h4>${t('editor.functions')||'功能清单'} <span class="section-desc">${t('editor.functions_desc')||'插件提供的具体功能接口'}</span>
      <button class="btn btn-xs" style="background:#27ae60" onclick="addFunction()">➕ 添加功能</button>
    </h4>
    <div id="functions-list">
      ${renderFunctionsList()}
    </div>
  </div>`;

  // ===== 状态键标签 =====
  html += `<div class="editor-section-card">
    <h4>${t('editor.state_labels')||'状态键标签'} <span class="section-desc">${t('editor.state_labels_desc')||'插件产出的状态键中文描述'}</span>
      <button class="btn btn-xs" style="background:#27ae60" onclick="addStateKeyLabel()" data-i18n="editor.add_label">➕ 添加</button>
    </h4>
    <p class="hint">${t('editor.label_hint')||'格式：{"键名": "中文描述"}，例：{"processed.ssid": "处理后的 WiFi 名称"}'}</p>
    <div id="state-labels-list">
      ${renderStateLabelsList()}
    </div>
  </div>`;

  // ===== 依赖声明 =====
  const deps = p.dependencies || {};
  html += `<div class="editor-section-card">
    <h4>${t('editor.dependencies')||'前置依赖'} <span class="section-desc">${t('editor.dependencies_desc')||'插件运行需要的环境条件'}</span></h4>
    <div class="editor-grid">
      <div class="editor-field">
        <label>${t('editor.requires_admin')||'需要管理员权限'}</label>
        <select id="ef-deps-admin" onchange="updateDepsField('requires_admin', this.value === 'true')">
          <option value="true" ${deps.requires_admin?'selected':''}>${t('editor.yes')||'是'}</option>
          <option value="false" ${!deps.requires_admin?'selected':''}>${t('editor.no')||'否'}</option>
        </select>
      </div>
      <div class="editor-field">
        <label>${t('editor.requires_network')||'需要网络连接'}</label>
        <select id="ef-deps-network" onchange="updateDepsField('requires_network', this.value === 'true')">
          <option value="true" ${deps.requires_network?'selected':''}>${t('editor.yes')||'是'}</option>
          <option value="false" ${!deps.requires_network?'selected':''}>${t('editor.no')||'否'}</option>
        </select>
      </div>
      <div class="editor-field">
        <label>${t('editor.min_ps_version')||'最低 PowerShell 版本'}</label>
        <input type="text" id="ef-deps-psver" value="${escHtml(deps.requires_powershell_version || '')}"
          placeholder="${t('editor.min_ps_placeholder')||'例: 5.1'}" onchange="updateDepsField('requires_powershell_version', this.value)" />
      </div>
      <div class="editor-field full-width">
        <label>${t('editor.dep_plugins')||'依赖的插件 ID（逗号分隔）'}</label>
        <input type="text" id="ef-deps-plugins" value="${escHtml((deps.depends_on_plugins||[]).join(', '))}"
          placeholder="plugin1, plugin2" onchange="updateDepsPlugins(this.value)" />
      </div>
    </div>
  </div>`;

  // ===== 标签 =====
  html += `<div class="editor-section-card">
    <h4>${t('editor.tags')||'标签'} <span class="section-desc">${t('editor.tags_desc')||'分类标签'}</span></h4>
    <div class="editor-field">
      <input type="text" id="ef-tags" value="${escHtml((p.tags||[]).join(', '))}"
        placeholder="${t('editor.tags_placeholder')||'标签1, 标签2, 标签3（逗号分隔）'}"
        onchange="updateTags(this.value)" style="width:100%;padding:6px 8px" />
    </div>
  </div>`;

  container.innerHTML = html;
  updateScriptTemplate();
}

// ============================================================================
// 功能清单渲染
// ============================================================================

function renderFunctionsList() {
  const funcs = _editorData.plugin.functions || {};
  const keys = Object.keys(funcs);
  if (keys.length === 0) {
    return '<div class="hint" style="padding:4px 0">暂无功能。点击「添加功能」添加插件功能接口。</div>';
  }
  let html = '';
  keys.forEach((fid, idx) => {
    const f = funcs[fid];
    html += `<div class="func-card">
      <div class="func-header">
        <input type="text" value="${escHtml(fid)}" placeholder="功能 ID（如 to_lowercase）"
          onchange="updateFuncId('${escHtml(fid)}', this.value)" style="font-weight:600;width:180px" />
        <button class="btn btn-xs btn-xs-del" onclick="deleteFunction('${escHtml(fid)}')">✕ 删除</button>
      </div>
      <div class="func-body">
        <div class="editor-grid" style="margin-bottom:6px">
          <div class="editor-field" style="grid-column:span 1">
            <label>名称</label>
            <input type="text" value="${escHtml(f.name || '')}" placeholder="功能名称"
              onchange="updateFuncField('${escHtml(fid)}', 'name', this.value)" />
          </div>
          <div class="editor-field" style="grid-column:span 2">
            <label>描述</label>
            <input type="text" value="${escHtml(f.description || '')}" placeholder="功能描述"
              onchange="updateFuncField('${escHtml(fid)}', 'description', this.value)" />
          </div>
        </div>

        <div style="display:flex;gap:12px">
          <div style="flex:1">
            <h5 style="font-size:12px;color:#555;margin-bottom:4px">输入参数（Input Params）</h5>
            ${renderFuncParams(fid, f.input_params || {}, 'input')}
            <button class="btn btn-xs" onclick="addFuncParam('${escHtml(fid)}', 'input')">➕ 参数</button>
          </div>
          <div style="flex:1">
            <h5 style="font-size:12px;color:#555;margin-bottom:4px">输出（Output）</h5>
            ${renderFuncParams(fid, f.output || {}, 'output')}
            <button class="btn btn-xs" onclick="addFuncParam('${escHtml(fid)}', 'output')">➕ 参数</button>
          </div>
        </div>
      </div>
    </div>`;
  });
  return html;
}

function renderFuncParams(funcId, params, kind) {
  const keys = Object.keys(params);
  if (keys.length === 0) {
    return '<div class="hint" style="font-size:11px;padding:2px 0">暂无参数</div>';
  }
  let html = '<div class="func-params">';
  keys.forEach(k => {
    const p = params[k];
    const type = p.type || 'string';
    html += `<div class="param-row">
      <input type="text" value="${escHtml(k)}" placeholder="参数名" style="width:80px;font-size:11px"
        onchange="renameFuncParam('${escHtml(funcId)}', '${kind}', '${escHtml(k)}', this.value)" />
      <select style="width:70px;font-size:11px"
        onchange="updateFuncParamField('${escHtml(funcId)}', '${kind}', '${escHtml(k)}', 'type', this.value)">
        <option value="string" ${type==='string'?'selected':''}>string</option>
        <option value="integer" ${type==='integer'?'selected':''}>integer</option>
        <option value="number" ${type==='number'?'selected':''}>number</option>
        <option value="boolean" ${type==='boolean'?'selected':''}>boolean</option>
      </select>
      <input type="text" value="${escHtml(p.description || '')}" placeholder="描述" style="flex:1;font-size:11px"
        onchange="updateFuncParamField('${escHtml(funcId)}', '${kind}', '${escHtml(k)}', 'description', this.value)" />
      <label style="font-size:11px;white-space:nowrap">
        <input type="checkbox" ${p.required?'checked':''}
          onchange="updateFuncParamField('${escHtml(funcId)}', '${kind}', '${escHtml(k)}', 'required', this.checked)" /> 必填
      </label>
      <button class="btn btn-xs btn-xs-del" onclick="deleteFuncParam('${escHtml(funcId)}', '${kind}', '${escHtml(k)}')">✕</button>
    </div>`;
  });
  html += '</div>';
  return html;
}

// ============================================================================
// 状态键标签渲染
// ============================================================================

function renderStateLabelsList() {
  const labels = _editorData.plugin.state_key_labels || {};
  const keys = Object.keys(labels);
  if (keys.length === 0) {
    return '<div class="hint" style="padding:4px 0">暂无状态键标签</div>';
  }
  let html = '';
  keys.forEach(k => {
    html += `<div class="state-label-row">
      <input type="text" value="${escHtml(k)}" placeholder="状态键名" style="width:200px"
        onchange="renameStateLabel('${escHtml(k)}', this.value)" />
      <span style="color:#999">→</span>
      <input type="text" value="${escHtml(labels[k] || '')}" placeholder="中文描述" style="flex:1"
        onchange="updateStateLabelDesc('${escHtml(k)}', this.value)" />
      <button class="btn btn-xs btn-xs-del" onclick="deleteStateLabel('${escHtml(k)}')">✕</button>
    </div>`;
  });
  return html;
}

// ============================================================================
// 已有插件选择器
// ============================================================================

async function renderPluginSelector() {
  const formContainer = document.getElementById('editor-form-container');
  formContainer.style.display = 'block';
  document.getElementById('editor-actions').style.display = 'flex';

  // 加载所有插件
  const data = await apiFetch('/plugins');
  let allPlugins = [];
  if (!data.error) {
    allPlugins = data.all || [];
  }
  _editorData.plugins = allPlugins;

  let html = renderPluginSelectorInline();

  // 如果有选中的插件，渲染表单
  if (_editorData.plugin) {
    html += '<div style="margin-top:16px">';
    formContainer.innerHTML = html;
    renderEditorForm();
    return;
  }

  html += '<div class="hint" style="padding:20px;text-align:center;font-size:14px">👆 ' + (t('editor.select_from_above')||'从上方选择一个已有的插件开始编辑') + '</div>';
  formContainer.innerHTML = html;
}

function renderPluginSelectorInline() {
  const plugins = _editorData.plugins;
  const currentId = _editorData.plugin ? _editorData.plugin.plugin_id : '';

  // 过滤掉内置插件（内置插件不允许修改）
  const editablePlugins = plugins.filter(p => !p.built_in);

  let html = `<div class="editor-section-card">
    <h4>${t('editor.select_plugin_title')||'选择插件'} <span style="font-weight:normal;font-size:12px;color:#999">${t('editor.builtin_not_editable')||'（内置插件不可编辑）'}</span></h4>
    <div class="editor-field" style="flex-direction:row;align-items:center;gap:8px">
      <select id="ef-plugin-selector" style="flex:1;padding:6px 8px" onchange="onPluginSelected(this.value)">
        <option value="">${t('editor.select_plugin_prompt')||'— 选择一个插件 —'}</option>`;

  editablePlugins.forEach(p => {
    const sel = p.id === currentId ? 'selected' : '';
    const label = `${p.name} (${p.id}) — ${p.type === 'input' ? t('editor.input_type')||'📥 输入' : t('editor.output_type')||'📤 输出'}`;
    html += `<option value="${p.id}" ${sel}>${escHtml(label)}</option>`;
  });

  const builtinCount = plugins.length - editablePlugins.length;
  html += `</select>
      <button class="btn btn-sm" onclick="refreshPluginList()">🔄 刷新</button>
    </div>
    ${builtinCount > 0 ? `<p class="hint" style="margin:4px 0 0">${t('editor.builtin_hidden').replace('{0}', builtinCount)||'已隐藏 '+builtinCount+' 个内置插件（不可编辑）'}</p>` : ''}
  </div>`;
  return html;
}

async function refreshPluginList() {
  const data = await apiFetch('/plugins');
  if (!data.error) {
    _editorData.plugins = data.all || [];
  }
  renderPluginSelector();
}

async function onPluginSelected(pluginId) {
  if (!pluginId) {
    _editorData.plugin = null;
    renderPluginSelector();
    return;
  }

  // 加载该插件的 rule.json
  const rules = await apiFetch('/rules');
  if (rules.error) {
    setStatus('加载插件失败: ' + rules.error, 'error');
    return;
  }

  const rule = rules[pluginId];
  if (!rule) {
    setStatus('未找到插件定义: ' + pluginId, 'error');
    return;
  }

  // 复制数据到编辑器
  _editorData.plugin = JSON.parse(JSON.stringify(rule));
  _editorData.originalId = pluginId;
  _editorData.scriptContent = '';

  renderPluginSelector();
}

// ============================================================================
// 字段更新函数
// ============================================================================

function updateField(field, value) {
  _editorData.plugin[field] = value;
}

function updateDirection(value) {
  _editorData.plugin.direction = value;
  // 切换方向时重新渲染表单
  renderEditorForm();
}

function updateExeField(field, value) {
  if (!_editorData.plugin.executable) {
    _editorData.plugin.executable = {};
  }
  _editorData.plugin.executable[field] = value;
}

function updatePolicyField(field, value) {
  if (!_editorData.plugin.execution_policy) {
    _editorData.plugin.execution_policy = {};
  }
  _editorData.plugin.execution_policy[field] = value;
}

function updateLifecycleField(field, value) {
  if (!_editorData.plugin.lifecycle) {
    _editorData.plugin.lifecycle = {};
  }
  _editorData.plugin.lifecycle[field] = value;
}

function updateDepsField(field, value) {
  if (!_editorData.plugin.dependencies) {
    _editorData.plugin.dependencies = {};
  }
  _editorData.plugin.dependencies[field] = value;
}

function updateDepsPlugins(value) {
  const plugins = value.split(',').map(s => s.trim()).filter(s => s);
  if (!_editorData.plugin.dependencies) {
    _editorData.plugin.dependencies = {};
  }
  _editorData.plugin.dependencies.depends_on_plugins = plugins;
}

function updateTags(value) {
  _editorData.plugin.tags = value.split(',').map(s => s.trim()).filter(s => s);
}

// ============================================================================
// 功能清单操作
// ============================================================================

function addFunction() {
  const fid = 'func_' + Date.now().toString(36);
  if (!_editorData.plugin.functions) {
    _editorData.plugin.functions = {};
  }
  _editorData.plugin.functions[fid] = {
    name: t('editor.new_func') || 'New Function',
    description: '',
    input_params: {},
    output: {}
  };
  renderEditorForm();
}

function deleteFunction(fid) {
  if (!confirm('确定删除功能 "' + fid + '" 吗？')) return;
  delete _editorData.plugin.functions[fid];
  renderEditorForm();
}

function updateFuncId(oldId, newId) {
  if (oldId === newId || !newId.trim()) return;
  const funcs = _editorData.plugin.functions;
  if (!funcs) return;
  funcs[newId] = funcs[oldId];
  delete funcs[oldId];
  renderEditorForm();
}

function updateFuncField(fid, field, value) {
  const funcs = _editorData.plugin.functions;
  if (!funcs || !funcs[fid]) return;
  funcs[fid][field] = value;
}

function addFuncParam(funcId, kind) {
  const funcs = _editorData.plugin.functions;
  if (!funcs || !funcs[funcId]) return;
  const key = 'param_' + Date.now().toString(36);
  if (!funcs[funcId][kind]) {
    funcs[funcId][kind] = {};
  }
  funcs[funcId][kind][key] = {
    type: 'string',
    description: '',
    required: false
  };
  renderEditorForm();
}

function renameFuncParam(funcId, kind, oldKey, newKey) {
  if (oldKey === newKey || !newKey.trim()) return;
  const params = _editorData.plugin.functions[funcId][kind];
  if (!params) return;
  params[newKey] = params[oldKey];
  delete params[oldKey];
  renderEditorForm();
}

function updateFuncParamField(funcId, kind, key, field, value) {
  const params = _editorData.plugin.functions[funcId][kind];
  if (!params || !params[key]) return;
  params[key][field] = value;
}

function deleteFuncParam(funcId, kind, key) {
  const params = _editorData.plugin.functions[funcId][kind];
  if (!params) return;
  delete params[key];
  renderEditorForm();
}

// ============================================================================
// 状态键标签操作
// ============================================================================

function addStateKeyLabel() {
  const labels = _editorData.plugin.state_key_labels || {};
  const newKey = 'key_' + Date.now().toString(36);
  labels[newKey] = '';
  _editorData.plugin.state_key_labels = labels;
  renderEditorForm();
}

function renameStateLabel(oldKey, newKey) {
  if (oldKey === newKey || !newKey.trim()) return;
  const labels = _editorData.plugin.state_key_labels;
  labels[newKey] = labels[oldKey];
  delete labels[oldKey];
  renderEditorForm();
}

function updateStateLabelDesc(key, value) {
  const labels = _editorData.plugin.state_key_labels;
  if (labels) labels[key] = value;
}

function deleteStateLabel(key) {
  const labels = _editorData.plugin.state_key_labels;
  if (labels) delete labels[key];
  renderEditorForm();
}

// ============================================================================
// 脚本模板生成
// ============================================================================

function updateScriptTemplate() {
  // 当可执行文件类型变更时自动更新文件名提示
  const exeType = _editorData.plugin.executable?.type || 'powershell';
  const pathField = document.getElementById('ef-exe-path');
  if (pathField && !pathField.value) {
    const ext = exeType === 'powershell' ? '.ps1' : exeType === 'batch' ? '.bat' : '.exe';
    const defaultPath = (_editorData.plugin.direction === 'input' ? 'Input_Plugins/' : 'Output_Plugins/')
      + _editorData.plugin.plugin_id + '/script' + ext;
    pathField.placeholder = defaultPath;
  }
}

function generateScriptTemplate() {
  const exeType = _editorData.plugin.executable?.type || 'powershell';
  const pluginId = _editorData.plugin.plugin_id;
  const funcs = _editorData.plugin.functions || {};
  const funcIds = Object.keys(funcs);

  let template = '';
  if (exeType === 'powershell') {
    template = `# ${_editorData.plugin.name} — ${pluginId}
# RuleCraft 插件脚本
# 通信协议：stdin/stdout JSON

param(
    [string]$InputJSON
)

# 读取输入
if (-not $InputJSON) {
    $InputJSON = [Console]::In.ReadToEnd()
}

try {
    $input = $InputJSON | ConvertFrom-Json
    $functionId = $input.function_id
    $value = $input.value
    $params = $input.params

    switch ($functionId) {`;
    funcIds.forEach(fid => {
      template += `
        "$fid" {
            # TODO: 实现 ${funcs[fid].name || fid}
            $result = $value
            Write-Output (ConvertTo-Json -Compress @{ value = $result })
            return
        }`;
    });
    template += `
        default {
            Write-Output (ConvertTo-Json -Compress @{ error = "Unknown function: $functionId" })
            exit 1
        }
    }
} catch {
    Write-Output (ConvertTo-Json -Compress @{ error = $_.Exception.Message })
    exit 1
}
`;
  } else if (exeType === 'batch') {
    template = `@echo off
REM ${_editorData.plugin.name} — ${pluginId}
REM RuleCraft 插件脚本
REM 通信协议：stdin/stdout JSON

setlocal enabledelayedexpansion

REM 读取输入
set "InputJSON="
for /f "delims=" %%a in ('findstr /n "^"') do (
    set "line=%%a"
    set "line=!line:*:=!"
    set "InputJSON=!InputJSON!!line!"
)

REM TODO: 解析 JSON 并处理功能
echo {"value": ""}
endlocal
`;
  } else {
    // executable - just state the protocol
    template = `// ${_editorData.plugin.name} — ${pluginId}
// RuleCraft 可执行插件
// 通信协议：
//   输入（stdin）：{"function_id":"...", "value":"...", "params":{}}
//   输出（stdout）：{"value":"..."}
//
// 构建后请将可执行文件放在对应插件目录下
`;
  }

  _editorData.scriptContent = template;
  const textarea = document.getElementById('ef-script-content');
  if (textarea) {
    textarea.value = template;
  }
}

// ============================================================================
// 数据收集
// ============================================================================

function collectPluginData() {
  // 表单上的字段已经在 updateField 中实时同步，
  // 只有 functions 中的参数可能需要额外处理。
  // 验证必填字段
  const p = _editorData.plugin;
  const errors = [];

  if (!p.plugin_id || !p.plugin_id.match(/^[a-z][a-z0-9_-]+$/)) {
    errors.push('插件 ID 必须是小写字母开头，仅含 a-z、0-9、下划线');
  }
  if (!p.name) errors.push('插件名称不能为空');
  if (!p.version || !p.version.match(/^\d+\.\d+\.\d+$/)) {
    errors.push('版本号格式不正确（应为 x.y.z）');
  }
  if (!p.executable?.path) {
    errors.push('可执行文件路径不能为空');
  }
  if (!p.executable?.type) {
    errors.push('可执行文件类型不能为空');
  }

  return { plugin: p, errors };
}

// ============================================================================
// 保存到磁盘
// ============================================================================

async function savePluginToDisk() {
  const { plugin, errors } = collectPluginData();
  if (errors.length > 0) {
    setStatus('❌ ' + errors.join('；'), 'error');
    return;
  }

  // 确保方向字段
  if (!plugin.direction) {
    plugin.direction = 'input';
  }

  const res = await apiFetch('/rules', {
    method: 'POST',
    body: JSON.stringify(plugin),
  });

  if (res.error) {
    setStatus('❌ 保存失败: ' + res.error, 'error');
    return;
  }

  // 如果有脚本内容，保存到插件目录
  if (_editorData.scriptContent) {
    const dir = plugin.direction === 'output' ? 'Output_Plugins' : 'Input_Plugins';
    const scriptPath = plugin.executable?.path || '';
    const scriptName = scriptPath.split('/').pop() || plugin.plugin_id + '.ps1';
    await apiFetch('/plugins/save-script', {
      method: 'POST',
      body: JSON.stringify({
        plugin_id: plugin.plugin_id,
        direction: plugin.direction,
        file_name: scriptName,
        content: _editorData.scriptContent
      })
    });
  }

  setStatus('✅ 插件已保存到插件文件夹', 'success');

  // 如果是编辑已有插件，更新原始 ID
  if (_editorData.mode === 'edit-existing') {
    _editorData.originalId = plugin.plugin_id;
  }

  // 自动更新可用的插件列表
  _editorData.plugins = [];
  refreshPluginList();
}

// ============================================================================
// 导出 ZIP
// ============================================================================

async function exportPluginZIP() {
  const { plugin, errors } = collectPluginData();
  if (errors.length > 0) {
    setStatus('❌ ' + errors.join('；'), 'error');
    return;
  }

  if (!plugin.direction) {
    plugin.direction = 'input';
  }

  setStatus('⏳ 正在生成压缩包...', '');

  // 确定脚本文件名
  const exeType = plugin.executable?.type || 'powershell';
  const ext = exeType === 'powershell' ? '.ps1' : exeType === 'batch' ? '.bat' : '.exe';
  const scriptFileName = plugin.plugin_id + ext;

  try {
    const res = await fetch('/api/plugins/export-zip', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        rule: plugin,
        script_content: _editorData.scriptContent || '',
        script_file_name: scriptFileName
      })
    });

    if (!res.ok) {
      const err = await res.json();
      setStatus('❌ 导出失败: ' + (err.error || res.statusText), 'error');
      return;
    }

    // 触发下载
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = plugin.plugin_id + '.zip';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    setStatus('✅ ZIP 已下载', 'success');
  } catch (e) {
    setStatus('❌ 导出失败: ' + e.message, 'error');
  }
}

// ============================================================================
// 预览 JSON
// ============================================================================

function previewPluginJSON() {
  const { plugin, errors } = collectPluginData();
  if (errors.length > 0) {
    setStatus('⚠️ 数据有误，但可预览部分内容: ' + errors.join('；'), 'error');
  }

  // 清理空字段
  const clean = JSON.parse(JSON.stringify(plugin));
  // 删除空数组和空对象
  if (clean.tags && clean.tags.length === 0) delete clean.tags;
  if (clean.functions && Object.keys(clean.functions).length === 0) delete clean.functions;
  if (clean.state_key_labels && Object.keys(clean.state_key_labels).length === 0) delete clean.state_key_labels;
  if (clean.dependencies && Object.keys(clean.dependencies).length === 0) delete clean.dependencies;
  if (clean.execution_policy && Object.keys(clean.execution_policy).length === 0) delete clean.execution_policy;
  if (clean.executable) {
    if (!clean.executable.args) delete clean.executable.args;
    if (!clean.executable.working_dir) delete clean.executable.working_dir;
  }

  document.getElementById('preview-json').textContent = JSON.stringify(clean, null, 2);
  document.getElementById('preview-overlay').style.display = 'flex';
}

function closePreviewModal() {
  document.getElementById('preview-overlay').style.display = 'none';
}

// ============================================================================
// 关闭编辑器
// ============================================================================

function closePluginEditor() {
  if (_editorData.plugin && _editorData.mode !== 'edit-existing') {
    // 检查是否有未保存的数据
    const hasContent = _editorData.plugin.name &&
      _editorData.plugin.name !== 'New Input Plugin' &&
      _editorData.plugin.name !== 'New Output Plugin';
    if (hasContent && !confirm('确定退出编辑器吗？未保存的修改将丢失。')) {
      return;
    }
  }

  // 通知后端释放资源
  apiFetch('/plugins/editor/close', { method: 'POST' });

  // 返回主页面
  switchPage('plugins');
}

// ============================================================================
// 状态提示
// ============================================================================

function setStatus(msg, type) {
  const el = document.getElementById('editor-status');
  if (el) {
    el.textContent = msg;
    el.className = 'editor-status ' + (type || '');
    if (msg) {
      el.style.display = 'inline';
    } else {
      el.style.display = 'none';
    }
  }
}

// ============================================================================
// 工具函数
// ============================================================================

function escHtml(s) {
  if (typeof s !== 'string') return s || '';
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

// ============================================================================
// 暴露到全局（HTML onclick 引用）
// ============================================================================

// 将 plugin-editor.html 作为页面嵌入 SPA 时，
// 以上函数通过 `<script src="plugin-editor.js"></script>` 加载即可使用。
