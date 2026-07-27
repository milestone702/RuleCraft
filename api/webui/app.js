// RuleCraft Web UI — 与 REST API 交互
const API_BASE = '/api';

// ============================================================================
// 插件分类
// ============================================================================

// 输入插件分类
const INPUT_PLUGIN_CATEGORIES = {
  '系统类': ['power', 'time', 'session', 'uptime', 'os', 'locale'],
  '硬件类': ['cpu', 'sysres', 'disk', 'display', 'battery_detail', 'volume'],
  '网络类': ['wifi', 'network', 'network_stats', 'http_request', 'network_detect'],
  '入站类': ['http_in', 'tcp_udp_in', 'websocket_in'],
  '文件类': ['file_monitor', 'clipboard'],
  '进程类': ['process', 'window'],
  '触发类': ['manual_trigger', 'idle'],
};

// 输出插件分类
const OUTPUT_PLUGIN_CATEGORIES = {
  '通知类': ['notify'],
  '系统控制类': ['power_scheme', 'prevent_sleep', 'lock', 'power_action'],
  '媒体类': ['volume', 'brightness', 'wallpaper', 'screenshot'],
  '进程类': ['exec', 'kill_process'],
  '网络类': ['webhook', 'http_get', 'http_post', 'netadapter'],
  '文件类': ['file', 'open', 'clipboard_set'],
  '工具类': ['delay'],
  '窗口类': ['window_control'],
};

// 搜索状态
let _pluginSearchText = '';
let _inputSearchText = '';
let _outputSearchText = '';

// 合并所有分类用于插件列表筛选
const ALL_PLUGIN_CATEGORIES = { ...INPUT_PLUGIN_CATEGORIES, ...OUTPUT_PLUGIN_CATEGORIES };

// 根据插件 ID 查找所属分类
function findPluginCategoryById(pluginId) {
  for (const [cat, ids] of Object.entries(ALL_PLUGIN_CATEGORIES)) {
    if (ids.includes(pluginId)) return cat;
  }
  return '其他';
}

// 构建带分类的插件选项 HTML（用于下拉菜单）
function buildCategorizedOptions(categories, plugins, selectedId, searchText) {
  let html = '<option value="">— 选择插件 —</option>';
  if (searchText) {
    const lower = searchText.toLowerCase();
    let found = false;
    plugins.forEach(p => {
      if ((p.name && p.name.toLowerCase().includes(lower)) ||
          (p.id && p.id.toLowerCase().includes(lower)) ||
          (p.description && p.description.toLowerCase().includes(lower))) {
        const sel = p.id === selectedId ? ' selected' : '';
        html += `<option value="${p.id}"${sel}>${p.name} (${p.id})</option>`;
        found = true;
      }
    });
    if (!found) html += '<option disabled>— 无匹配插件 —</option>';
  } else {
    let lastCat = '';
    plugins.forEach(p => {
      const cat = findPluginCategory(categories, p.id);
      if (cat !== lastCat) {
        if (lastCat) html += '</optgroup>';
        html += `<optgroup label="${cat}">`;
        lastCat = cat;
      }
      const sel = p.id === selectedId ? ' selected' : '';
      html += `<option value="${p.id}"${sel}>${p.name} (${p.id})</option>`;
    });
    if (lastCat) html += '</optgroup>';
  }
  return html;
}

function findPluginCategory(categories, pluginId) {
  for (const [cat, ids] of Object.entries(categories)) {
    if (ids.includes(pluginId)) return cat;
  }
  return '其他';
}

// 搜索输入插件的 HTML
function renderPluginSearch(onchangeHandler) {
  return `<input type="text" class="plugin-search-input" placeholder="🔍 搜索插件..."
    value="" onchange="${onchangeHandler}" style="flex:1;padding:4px 8px;min-width:80px;font-size:12px;border:1px solid #ddd;border-radius:3px" />`;
}

function renderInputSearch() {
  return `<input type="text" class="input-search-box" placeholder="🔍 搜索输入插件..."
    value="${escHtml(_inputSearchText)}" onchange="onDataSourceSearch()" style="flex:1;padding:4px 8px;min-width:80px;font-size:12px;border:1px solid #ddd;border-radius:3px" />`;
}

function renderOutputSearch() {
  return `<input type="text" class="output-search-box" placeholder="🔍 搜索输出插件..."
    value="${escHtml(_outputSearchText)}" onchange="onOutputSearch()" style="flex:1;padding:4px 8px;min-width:80px;font-size:12px;border:1px solid #ddd;border-radius:3px" />`;
}

// ============================================================================
// 页面导航
// ============================================================================

document.addEventListener('DOMContentLoaded', () => {
  // 初始化语言选择器
  const langSel = document.getElementById('lang-selector');
  if (langSel) {
    langSel.value = getLang();
  }
  // 应用 i18n 翻译
  applyI18n();

  // 给所有导航链接绑定点击事件
  document.querySelectorAll('.nav-links a, .nav-sublinks a').forEach(link => {
    // 排除父级切换按钮（它用 onclick 处理折叠）
    if (link.classList.contains('nav-parent-toggle')) return;
    link.addEventListener('click', (e) => {
      e.preventDefault();
      switchPage(link.dataset.page);
    });
  });
  loadDashboard();
  // 启动连接状态心跳检测
  startConnectionHeartbeat();
});

// 语言切换时重新加载当前页面内容
document.addEventListener('langchange', () => {
  // 重新渲染当前页面
  const activePage = document.querySelector('.page.active');
  if (!activePage) return;
  const id = activePage.id;
  if (id === 'page-dashboard') loadDashboard();
  else if (id === 'page-tasks') loadTasks();
  else if (id === 'page-plugins') loadPlugins();
  else if (id === 'page-market') loadMarket();
  else if (id === 'page-logs') loadLogs();
  else if (id === 'page-settings') loadSettings();
  else if (id === 'page-manual') loadManual();
  // 更新导航栏文本
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.dataset.i18n;
    if (el.tagName !== 'INPUT' && el.tagName !== 'TEXTAREA') {
      el.innerHTML = t(key);
    }
  });
});

// 折叠/展开插件编辑器二级菜单
function toggleEditorMenu(event) {
  event.preventDefault();
  const sublinks = document.querySelector('.nav-sublinks');
  const arrow = document.querySelector('.collapse-arrow');
  if (!sublinks) return;
  const isOpen = sublinks.style.display !== 'none';
  sublinks.style.display = isOpen ? 'none' : 'block';
  if (arrow) arrow.style.transform = isOpen ? 'rotate(0deg)' : 'rotate(90deg)';
}

function switchPage(pageId) {
  // 特殊处理：插件编辑器的二级子页面
  if (pageId === 'plugin-editor-exit') {
    exitPluginEditor();
    return;
  }
  if (pageId.startsWith('plugin-editor-')) {
    const modeMap = {
      'plugin-editor-input': 'input-new',
      'plugin-editor-output': 'output-new',
      'plugin-editor-edit': 'edit-existing',
    };
    const mode = modeMap[pageId] || 'input-new';

    // 高亮子级，清除其他高亮
    document.querySelectorAll('.nav-links > li > a, .nav-sublinks a').forEach(a => a.classList.remove('active'));
    const subLink = document.querySelector(`.nav-sublinks a[data-page="${pageId}"]`);
    if (subLink) subLink.classList.add('active');

    // 显示编辑器页面
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    document.getElementById('page-plugin-editor')?.classList.add('active');
    document.getElementById('page-title').textContent = t('nav.plugin-editor');

    openPluginEditor(mode);
    return;
  }

  document.querySelectorAll('.nav-links > li > a, .nav-sublinks a').forEach(a => a.classList.remove('active'));
  const link = document.querySelector(`.nav-links a[data-page="${pageId}"], .nav-sublinks a[data-page="${pageId}"]`);
  if (link) link.classList.add('active');

  document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
  const page = document.getElementById(`page-${pageId}`);
  if (page) page.classList.add('active');

  const titleKey = 'page.' + pageId;
  document.getElementById('page-title').textContent = t(titleKey) || pageId;

  switch(pageId) {
    case 'dashboard': loadDashboard(); break;
    case 'plugins': loadPlugins(); break;
    case 'tasks': loadTasks(); break;
    case 'market': loadMarket(); break;
    case 'settings': loadSettings(); break;
    case 'manual': loadManual(); break;
  }
}

// ============================================================================
// API 调用
// ============================================================================

async function apiFetch(path, options = {}) {
  try {
    const res = await fetch(`${API_BASE}${path}`, {
      headers: { 'Content-Type': 'application/json' },
      ...options,
    });
    return await res.json();
  } catch (err) {
    console.error(`API ${path} failed:`, err);
    return { error: err.message };
  }
}

// ============================================================================
// 仪表盘 — 友好展示系统状态（实时刷新）
// ============================================================================

let realtimeTimer = null;

async function loadDashboard() {
  const status = await apiFetch('/status');
  if (status.error) {
    document.getElementById('state-viewer').innerHTML = `<div class="error">${t('dashboard.connection_failed')}: ${status.error}</div>`;
    return;
  }

  const keys = Object.keys(status);
  document.getElementById('stat-states').textContent = keys.length;

  const plugins = await apiFetch('/plugins');
  if (!plugins.error) {
    const count = (plugins.inputs?.length || 0) + (plugins.outputs?.length || 0);
    document.getElementById('stat-plugins').textContent = count;
  }

  // 按分组
  const groups = {}, diskDrives = {}, timeFields = {};
  keys.forEach(k => {
    const parts = k.split('.');
    const prefix = parts[0];
    if (prefix === 'disk' && parts.length >= 3) {
      const drive = parts[1], metric = parts[2];
      if (!diskDrives[drive]) diskDrives[drive] = {};
      diskDrives[drive][metric] = status[k]; return;
    }
    if (prefix === 'time') { timeFields[parts[1]] = status[k]; return; }
    if (prefix === 'sysres') {
      if (!groups.sysres) groups.sysres = {};
      groups.sysres[parts[1]] = status[k]; return;
    }
    if (!groups[prefix]) groups[prefix] = [];
    groups[prefix].push({ key: k, value: status[k] });
  });

  let html = '';
  const groupLabels = { power: t('group.power'), wifi: t('group.wifi'), network: t('group.network'), window: t('group.window'), idle: t('group.idle'), session: t('group.session'), manual_trigger: t('group.manual_trigger'), processed: t('group.processed'), other: t('group.other') };

  // 手动触发卡片（高亮显示）
  const mt = groups.manual_trigger;
  if (mt && mt.some(i => i.key === 'manual_trigger.triggered' && i.value === true)) {
    const taskId = mt.find(i => i.key === 'manual_trigger.task_id')?.value || '';
    const at = mt.find(i => i.key === 'manual_trigger.at')?.value || '';
    html += `<div class="state-card" style="background:#fff3cd;border-left:4px solid #f39c12;margin-bottom:8px">
      🖱 <strong>${t('dashboard.manual_trigger')}</strong> — ${t('dashboard.manual_triggered').replace('{0}', escHtml(String(taskId))).replace('{1}', escHtml(String(at)))}
    </div>`;
  }
  delete groups.manual_trigger;

  // 时间卡片（ID 用于实时更新）
  html += `<div class="state-card time-card" id="rt-time">⏰ ${formatTime(timeFields)}</div>`;

  // 资源卡片（ID 用于实时更新）
  const sr = groups.sysres || {};
  html += `<div class="state-card res-card" id="rt-res">📊 CPU ${fmtPct(sr.cpu_percent)} &nbsp;|&nbsp; 内存 ${fmtMem(sr.memory_used_gb, sr.memory_total_gb, sr.memory_percent)}</div>`;
  delete groups.sysres;

  // 普通分组：3 列网格
  const groupOrder = ['power','wifi','network','window','idle','session'];
  groupOrder.forEach(prefix => {
    const items = groups[prefix];
    if (!items || items.length === 0) return;
    html += `<div class="state-group"><h4>${groupLabels[prefix]||prefix}</h4><div class="info-grid">`;
    items.forEach(item => {
      html += `<div class="info-item"><span class="i-key">${formatKey(item.key)}</span><span class="i-val">${formatValue(item.value)}</span></div>`;
    });
    html += '</div></div>';
  });

  // 进程按钮（在会话和磁盘之间）
  html += `<div class="state-group" style="margin-bottom:4px"><button class="btn btn-sm" onclick="showProcessModal()">${t('dashboard.view_processes')}</button></div>`;

  // 磁盘：饼图 + 文字
  const dl = Object.keys(diskDrives).sort();
  if (dl.length > 0) {
    html += `<div class="state-group"><h4>💾 磁盘</h4><div class="disk-grid">`;
    dl.forEach(drive => {
      const d = diskDrives[drive];
      const total = d.total_gb || 0, free = d.free_gb || 0, usedPct = d.used_percent || 0;
      const usedG = total - free;
      const pct = typeof usedPct === 'number' ? usedPct.toFixed(1) : usedPct;
      const tStr = typeof total === 'number' ? total.toFixed(1) : total;
      const uStr = typeof usedG === 'number' ? usedG.toFixed(1) : usedG;
      html += `<div class="disk-card">
        <div class="disk-pie" style="background: conic-gradient(#27ae60 0deg ${(100-pct)*3.6}deg, #e74c3c ${(100-pct)*3.6}deg 360deg)"></div>
        <div class="disk-label">${drive.replace(':','：')}</div>
        <div class="disk-info">已用 <b>${uStr}G</b> / ${tStr}G (${pct}%)</div>
      </div>`;
    });
    html += `</div></div>`;
  }

  document.getElementById('state-viewer').innerHTML = html;

  // 启动实时刷新（每秒更新时间和资源）
  startRealtime();
}

function formatTime(f) {
  if (!f || Object.keys(f).length === 0) return t('dashboard.loading');
  if (f.now) return escHtml(f.now) + (f.timezone ? ` <span class="tz">${f.timezone}</span>` : '');
  const d = new Date();
  const pad = n => String(n).padStart(2,'0');
  const h = f.hour !== undefined ? pad(f.hour) : pad(d.getHours());
  const mi = f.minute !== undefined ? pad(f.minute) : pad(d.getMinutes());
  const s = f.second !== undefined ? pad(f.second) : pad(d.getSeconds());
  const y = d.getFullYear();
  const mo = f.month !== undefined ? pad(f.month) : pad(d.getMonth()+1);
  const da = f.day !== undefined ? pad(f.day) : pad(d.getDate());
  const wd = f.weekday_name || ['日','一','二','三','四','五','六'][d.getDay()];
  const z = f.timezone || '';
  return `${y}-${mo}-${da} 星期${wd} ${h}:${mi}:${s} <span class="tz">${z}</span>`;
}

function fmtPct(v) {
  if (v === null || v === undefined) return '--';
  const n = typeof v === 'number' ? v.toFixed(1) : v;
  return `<b>${n}%</b>`;
}

function fmtMem(used, total, pct) {
  if (used === null || used === undefined || total === null || total === undefined)
    return '--';
  const u = typeof used === 'number' ? used.toFixed(1) : used;
  const t = typeof total === 'number' ? total.toFixed(1) : total;
  const p = typeof pct === 'number' ? pct.toFixed(1) : pct;
  return `<b>${u}G</b> / ${t}G <span class="pct">${p}%</span>`;
}

// 实时刷新：每秒更新时间和资源
function startRealtime() {
  if (realtimeTimer) clearInterval(realtimeTimer);
  realtimeTimer = setInterval(async () => {
    // 更新时间
    const st = await apiFetch('/status');
    if (st.error) return;
    const rt = document.getElementById('rt-time');
    if (rt) {
      const tf = {};
      Object.keys(st).filter(k => k.startsWith('time.')).forEach(k => tf[k.split('.')[1]] = st[k]);
      rt.innerHTML = '⏰ ' + formatTime(tf);
    }
    // 更新资源
    const rr = document.getElementById('rt-res');
    if (rr) {
      const cpu = st['sysres.cpu_percent'];
      const mu = st['sysres.memory_used_gb'];
      const mt = st['sysres.memory_total_gb'];
      const mp = st['sysres.memory_percent'];
      rr.innerHTML = '📊 CPU ' + fmtPct(cpu) + ' &nbsp;|&nbsp; 内存 ' + fmtMem(mu, mt, mp);
    }
  }, 1000);
}

function formatValue(v) {
  if (v === null || v === undefined) return '<span class="val-null">空</span>';
  if (typeof v === 'boolean') return v ? '<span class="val-true">✅ 是</span>' : '<span class="val-false">❌ 否</span>';
  if (typeof v === 'number') {
    let s = Number.isInteger(v) ? v : v.toFixed(2);
    return `<span class="val-num">${s}</span>`;
  }
  if (Array.isArray(v)) return `<span class="val-array">[${v.join(', ')}]</span>`;
  if (typeof v === 'object') return `<span class="val-obj">${JSON.stringify(v)}</span>`;
  return `<span class="val-str">${escHtml(String(v))}</span>`;
}

// 将状态键名转为中文可读名称
const keyLabels = {
  'wifi.ssid':'Wi-Fi 名称', 'wifi.bssid':'基站地址', 'wifi.is_connected':'连接状态', 'wifi.signal_quality':'信号质量',
  'power.is_charging':'充电中', 'power.battery_percent':'电池电量', 'power.ac_line_status':'电源状态',
  'network.ipv4':'IP 地址', 'network.gateway':'网关', 'network.is_connected':'网络连接', 'network.connection_type':'连接类型', 'network.adapter_name':'适配器',
  'process.count':'进程数', 'process.names':'进程列表', 'process.list_json':'进程详情',
  'window.foreground.title':'窗口标题', 'window.foreground.process':'所属进程',
  'idle.seconds':'空闲秒数', 'idle.is_idle':'是否空闲',
  'sysres.cpu_percent':'CPU 使用率', 'sysres.memory_percent':'内存使用率',
  'session.is_locked':'锁屏状态',
  'time.hour':'时', 'time.minute':'分', 'time.weekday':'星期', 'time.weekday_name':'星期', 'time.is_weekend':'是否周末',
  'time.month':'月', 'time.day':'日', 'time.unix_timestamp':'时间戳', 'time.iso8601':'时间', 'time.timezone':'时区', 'time.now':'当前时间',
};

function formatKey(key) {
  if (keyLabels[key]) return keyLabels[key];
  let parts = key.split('.');
  let name = parts[parts.length-1];
  return name.replace(/_/g,' ');
}

function escHtml(s) { return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }

// ============================================================================
// 插件列表（含筛选、管理、详情）
// ============================================================================

let _allPluginMetas = [];   // 缓存全部插件元数据
let _allRuleDefs = {};      // 缓存全部规则定义（按 plugin_id）

async function loadPlugins() {
  try {
    // 加载插件列表 + 规则详情
    const [plugData, rulesData] = await Promise.all([
      apiFetch('/plugins'),
      apiFetch('/rules')
    ]);

    if (plugData.error) {
      document.getElementById('plugins-list').innerHTML = `<div class="error">${plugData.error}</div>`;
      return;
    }
    _allPluginMetas = plugData.all || [];
    _allRuleDefs = {};
    if (!rulesData.error) {
      Object.values(rulesData).forEach(r => { if (r && r.plugin_id) _allRuleDefs[r.plugin_id] = r; });
    }
    renderPluginList();
  } catch (e) {
    document.getElementById('plugins-list').innerHTML = `<div class="error">加载失败: ${e.message}</div>`;
  }
}

function renderPluginList() {
  // 读取筛选条件
  const filterTypeEl = document.getElementById('plugin-filter-type');
  const filterOriginEl = document.getElementById('plugin-filter-origin');
  const filterCatEl = document.getElementById('plugin-filter-category');
  const filterTextEl = document.getElementById('plugin-filter-text');
  const filterType = filterTypeEl ? filterTypeEl.value : 'all';
  const filterOrigin = filterOriginEl ? filterOriginEl.value : 'all';
  const filterCat = filterCatEl ? filterCatEl.value : 'all';
  const filterText = filterTextEl ? filterTextEl.value.toLowerCase().trim() : '';

  let list = _allPluginMetas;
  if (filterType !== 'all') list = list.filter(p => p.type === filterType);
  if (filterOrigin !== 'all') {
    const isBuiltin = filterOrigin === 'builtin';
    list = list.filter(p => p.built_in === isBuiltin);
  }
  if (filterText) {
    list = list.filter(p =>
      (p.name && p.name.toLowerCase().includes(filterText)) ||
      (p.id && p.id.toLowerCase().includes(filterText)) ||
      (p.description && p.description.toLowerCase().includes(filterText))
    );
  }
  // 分类筛选
  if (filterCat !== 'all') {
    list = list.filter(p => findPluginCategoryById(p.id) === filterCat);
  }
  // 固定排序：按 ID 字母序
  list = list.sort((a, b) => (a.id || '').localeCompare(b.id || ''));

  let html = `
  <div style="display:flex;gap:8px;margin-bottom:12px;flex-wrap:wrap">
    <select id="plugin-filter-type" onchange="renderPluginList()" style="padding:4px 8px">
      <option value="all">全部类型</option>
      <option value="input" ${filterType==='input'?'selected':''}>📥 输入插件</option>
      <option value="output" ${filterType==='output'?'selected':''}>📤 输出插件</option>
    </select>
    <select id="plugin-filter-origin" onchange="renderPluginList()" style="padding:4px 8px">
      <option value="all">全部来源</option>
      <option value="builtin" ${filterOrigin==='builtin'?'selected':''}>内置插件</option>
      <option value="external" ${filterOrigin==='external'?'selected':''}>外部插件</option>
    </select>
    <select id="plugin-filter-category" onchange="renderPluginList()" style="padding:4px 8px">
      <option value="all">全部分类</option>
      ${Object.keys(ALL_PLUGIN_CATEGORIES).map(c => `<option value="${c}" ${filterCat===c?'selected':''}>${c}</option>`).join('')}
    </select>
    <input type="text" id="plugin-filter-text" placeholder="搜索名称/ID/描述…"
      value="${escHtml(document.getElementById('plugin-filter-text')?.value||'')}"
      oninput="renderPluginList()" style="flex:1;padding:4px 8px;min-width:120px" />
    <span style="font-size:12px;color:#999;line-height:28px">共 ${list.length} 个插件</span>
  </div>
  <table>
  <tr><th>#</th><th>名称</th><th>ID</th><th>类型</th><th>分类</th><th>作者</th><th>内置</th><th>操作</th></tr>`;

  list.forEach((p, i) => {
    const rule = _allRuleDefs[p.id] || {};
    const author = rule.author || '-';
    const desc = p.description || rule.description || '';
    html += `<tr>
      <td>${i+1}</td>
      <td><a href="javascript:void(0)" onclick="showPluginDetail('${p.id}')" style="font-weight:600">${p.name}</a></td>
      <td style="font-size:12px;color:#888">${p.id}</td>
      <td>${p.type === 'input' ? '📥 输入' : '📤 输出'}</td>
      <td style="font-size:12px;color:#888">${findPluginCategoryById(p.id)}</td>
      <td style="font-size:12px">${escHtml(author)}</td>
      <td>${p.built_in ? '✅' : '❌'}</td>
      <td>
        <button class="btn btn-xs" onclick="showPluginDetail('${p.id}')">📖 查看</button>
        ${p.built_in ? '' : `<button class="btn btn-xs btn-xs-del" onclick="deletePlugin('${p.id}')">🗑 删除</button>`}
      </td>
    </tr>`;
    if (desc) {
      html += `<tr style="background:#fafafa"><td colspan="8" style="font-size:12px;color:#888;padding:2px 12px">${escHtml(desc)}</td></tr>`;
    }
  });

  html += '</table>';
  document.getElementById('plugins-list').innerHTML = html;
}

// 显示插件详情弹窗
async function showPluginDetail(pluginId) {
  const p = (_allPluginMetas || []).find(m => m.id === pluginId);
  const rule = _allRuleDefs[pluginId] || {};
  if (!p) return;

  const overlay = document.getElementById('modal-overlay');
  const box = document.getElementById('modal-box');
  overlay.style.zIndex = '2000';
  box.querySelector('.modal-header span').textContent = `📦 ${p.name} (${p.id})`;

  // 可执行程序
  let execHtml = '';
  if (rule.executable) {
    execHtml = `<tr><td style="padding:4px 8px;border:1px solid #ddd;font-weight:600">程序路径</td>
      <td style="padding:4px 8px;border:1px solid #ddd"><code>${escHtml(rule.executable.path||'')}</code></td></tr>
      <tr><td style="padding:4px 8px;border:1px solid #ddd;font-weight:600">类型</td>
      <td style="padding:4px 8px;border:1px solid #ddd">${rule.executable.type||''}</td></tr>`;
  } else {
    execHtml = `<tr><td colspan="2" style="padding:4px 8px;border:1px solid #ddd;color:#999">Go 内置插件，无独立可执行程序</td></tr>`;
  }

  // 功能清单
  let funcHtml = '';
  if (rule.functions) {
    funcHtml = '<table style="width:100%;border-collapse:collapse;font-size:13px;margin-top:8px">';
    Object.entries(rule.functions).forEach(([fid, fn]) => {
      funcHtml += `<tr style="background:#f0f4f8"><th colspan="2" style="padding:6px 8px;border:1px solid #ddd;text-align:left">${fn.name || fid}</th></tr>`;
      if (fn.description) {
        funcHtml += `<tr><td colspan="2" style="padding:4px 8px;border:1px solid #ddd;color:#888">${fn.description}</td></tr>`;
      }
      // 输入参数
      if (fn.input_params) {
        Object.entries(fn.input_params).forEach(([pk, pv]) => {
          const req = pv.required ? '必填' : '可选';
          funcHtml += `<tr><td style="padding:3px 8px;border:1px solid #ddd;padding-left:20px">输入 <code>${pk}</code></td>
            <td style="padding:3px 8px;border:1px solid #ddd;font-size:12px">${pv.type||''} ${req}${pv.description ? ' — '+pv.description : ''}</td></tr>`;
        });
      }
      // 输出参数
      if (fn.output) {
        Object.entries(fn.output).forEach(([pk, pv]) => {
          funcHtml += `<tr><td style="padding:3px 8px;border:1px solid #ddd;padding-left:20px">输出 <code>${pk}</code></td>
            <td style="padding:3px 8px;border:1px solid #ddd;font-size:12px">${pv.type||''}${pv.description ? ' — '+pv.description : ''}</td></tr>`;
        });
      }
    });
    funcHtml += '</table>';
  }

  // 状态键（输入插件）
  let stateHtml = '';
  if (rule.state_key_labels) {
    stateHtml = '<h4 style="margin:8px 0 4px">输出状态键</h4><table style="width:100%;border-collapse:collapse;font-size:13px">';
    Object.entries(rule.state_key_labels).forEach(([k, v]) => {
      const meta = STATE_KEY_META[k];
      const hint = meta ? ` (${meta.hint})` : '';
      stateHtml += `<tr><td style="padding:3px 8px;border:1px solid #ddd"><code>${k}</code></td>
        <td style="padding:3px 8px;border:1px solid #ddd">${v}${hint}</td></tr>`;
    });
    stateHtml += '</table>';
  }

  // 内置插件的状态键
  if (p.built_in && p.type === 'input') {
    const builtinKeys = getBuiltinStateKeyLabels(p.id);
    if (Object.keys(builtinKeys).length) {
      stateHtml = '<h4 style="margin:8px 0 4px">输出状态键</h4><table style="width:100%;border-collapse:collapse;font-size:13px">';
      Object.entries(builtinKeys).forEach(([k, v]) => {
        const meta = STATE_KEY_META[k];
        const hint = meta ? ` (${meta.hint})` : '';
        stateHtml += `<tr><td style="padding:3px 8px;border:1px solid #ddd"><code>${k}</code></td>
          <td style="padding:3px 8px;border:1px solid #ddd">${v}${hint}</td></tr>`;
      });
      stateHtml += '</table>';
    }
  }

  box.querySelector('.modal-body').innerHTML = `
<div style="font-size:13px;line-height:1.7">
  <table style="width:100%;border-collapse:collapse;font-size:13px">
    <tr><td style="padding:4px 8px;border:1px solid #ddd;font-weight:600;width:80px">ID</td>
      <td style="padding:4px 8px;border:1px solid #ddd"><code>${p.id}</code></td></tr>
    <tr><td style="padding:4px 8px;border:1px solid #ddd;font-weight:600">名称</td>
      <td style="padding:4px 8px;border:1px solid #ddd">${p.name}</td></tr>
    <tr><td style="padding:4px 8px;border:1px solid #ddd;font-weight:600">类型</td>
      <td style="padding:4px 8px;border:1px solid #ddd">${p.type === 'input' ? '📥 输入' : '📤 输出'}${p.built_in ? ' (内置)' : ' (外部规则)'}</td></tr>
    ${execHtml}
  </table>
  ${stateHtml}
  ${funcHtml}
</div>`;
  overlay.style.display = 'flex';
}

// 删除外部插件
async function deletePlugin(pluginId) {
  if (!confirm(`确定删除插件 "${pluginId}" 吗？`)) return;
  const res = await apiFetch('/rules/' + encodeURIComponent(pluginId), { method: 'DELETE' });
  if (res.error) { alert('删除失败: ' + res.error); return; }
  alert('✅ 已删除');
  loadPlugins();
}

// ============================================================================
// 任务管理
// ============================================================================

async function loadTasks() {
  const data = await apiFetch('/tasks');
  if (data.error) {
    document.querySelector('#task-table').innerHTML = `<tr><td colspan="4">加载失败: ${data.error}</td></tr>`;
    return;
  }
  let html = '<tr><th>ID</th><th>名称</th><th>状态</th><th>操作</th></tr>';
  if (typeof data === 'object' && !Array.isArray(data)) {
    const entries = Object.entries(data);
    if (entries.length === 0) {
      html = '<tr><td colspan="4">' + t('tasks.no_tasks') + '</td></tr>';
    } else {
      entries.forEach(([id, task]) => {
        const name = task.name || id;
        const enabled = task.enabled !== false;
        html += `<tr>
          <td>${id}</td>
          <td>${escHtml(name)}</td>
          <td>${enabled ? '✅ 启用' : '⏸ 停用'}</td>
          <td>
            <button class="btn btn-sm" onclick="openTaskEditor('${id}')">✏️ 编辑</button>
            <button class="btn btn-sm" style="background:#f39c12" onclick="manualTriggerTask('${id}')">▶ 触发</button>
            <button class="btn btn-sm" style="background:#e74c3c" onclick="deleteTask('${id}')">🗑 删除</button>
          </td>
        </tr>`;
      });
    }
  }
  document.querySelector('#task-table').innerHTML = html;
}

async function addTask() {
  const name = prompt('输入任务名称（task_id 将自动生成）:');
  if (!name) return;
  const taskId = 'task_' + Date.now().toString(36);
  const newTask = {
    task_id: taskId, name: name, enabled: true, version: '1.0.0',
    data_sources: [], processors: [], condition: null, outputs: []
  };
  const res = await apiFetch('/tasks', {
    method: 'POST',
    body: JSON.stringify(newTask),
  });
  if (res.error) { alert('创建失败: ' + res.error); return; }
  alert(`✅ 任务 "${name}" 创建成功 (${taskId})`);
  loadTasks();
}

async function deleteTask(taskId) {
  if (!confirm(`确定删除任务 "${taskId}" 吗？此操作不可恢复。`)) return;
  const res = await apiFetch(`/tasks/${encodeURIComponent(taskId)}`, { method: 'DELETE' });
  if (res.error) { alert('删除失败: ' + res.error); return; }
  loadTasks();
}

// ============================================================================
// 任务编辑器 — 可视化条件树
// ============================================================================

let _editingTask = null;       // 当前编辑中的任务对象（深拷贝）
let _editingTaskId = null;     // 原始 task_id
let _condIdCounter = 0;       // 条件节点唯一 ID 计数器

// 支持的运算符（按类别分组）
const OPERATOR_GROUPS = {
  '比较': ['equals','not_equals','greater_than','less_than','greater_equal','less_equal'],
  '字符串': ['contains','not_contains','starts_with','ends_with','matches_regex','not_matches_regex'],
  '存在性': ['exists','not_exists','is_empty','is_not_empty'],
  '长度': ['length_equals','length_greater_than','length_less_than'],
  '枚举': ['in','not_in'],
};

// 获取所有运算符列表
function getAllOperators() {
  const list = [];
  for (const g in OPERATOR_GROUPS) OPERATOR_GROUPS[g].forEach(o => list.push(o));
  return list;
}

// 运算符中文标签
const OPERATOR_LABELS = {
  equals:'等于', not_equals:'不等于', greater_than:'大于', less_than:'小于',
  greater_equal:'大于等于', less_equal:'小于等于',
  contains:'包含', not_contains:'不包含', starts_with:'开头是', ends_with:'结尾是',
  matches_regex:'匹配正则', not_matches_regex:'不匹配正则',
  exists:'存在', not_exists:'不存在', is_empty:'为空', is_not_empty:'不为空',
  length_equals:'长度等于', length_greater_than:'长度大于', length_less_than:'长度小于',
  in:'在列表中', not_in:'不在列表中',
};

// 获取最近一次采集的状态键列表（供下拉选择）
let _cachedStateKeys = [];

async function refreshStateKeys() {
  const status = await apiFetch('/status');
  if (!status.error) _cachedStateKeys = Object.keys(status).sort();
}

// 状态键前缀 → 中文组名
const STATE_GROUP_LABELS = {
  'power': '🔌 电源', 'wifi': '📶 Wi-Fi', 'network': '🌐 网络',
  'window': '🪟 窗口', 'idle': '💤 空闲', 'session': '👤 会话',
  'sysres': '💻 系统资源', 'disk': '💾 磁盘', 'time': '⏰ 时间',
  'process': '⚙️ 进程',
  'clipboard': '📋 剪贴板', 'display': '🖥️ 显示', 'uptime': '⏱️ 运行时间',
  'os': '🖥️ 系统信息', 'cpu': '⚡ CPU', 'locale': '🌍 区域语言',
  'volume': '🔊 音量', 'battery_detail': '🔋 电池详细', 'network_stats': '📊 网络流量',
};

// 已知状态键 → 中文描述（内置插件 + 外部插件可注册到此映射）
// 外部插件可在加载时通过 addStateKeyLabels() 注入自己的备注
let _extraStateKeyLabels = {};

function addStateKeyLabels(labels) {
  Object.assign(_extraStateKeyLabels, labels);
}

const STATE_KEY_LABELS = {
  'power.is_charging': '是否在充电',
  'power.battery_percent': '电池百分比',
  'power.ac_line_status': '电源插头状态',
  'wifi.ssid': 'Wi-Fi 名称 (SSID)',
  'wifi.bssid': '基站 MAC 地址',
  'wifi.is_connected': 'Wi-Fi 连接状态',
  'wifi.signal_quality': 'Wi-Fi 信号质量 (0-100)',
  'network.ipv4': 'IP 地址',
  'network.gateway': '网关地址',
  'network.subnet_mask': '子网掩码',
  'network.is_connected': '网络是否连通',
  'window.title': '活动窗口标题',
  'window.process_name': '活动窗口进程名',
  'window.class_name': '窗口类名',
  'idle.seconds': '空闲秒数',
  'idle.idle_since': '空闲起始时间',
  'session.user_name': '当前用户名',
  'session.session_type': '会话类型',
  'session.logon_time': '登录时间',
  'sysres.cpu_percent': 'CPU 使用率',
  'sysres.memory_percent': '内存使用率',
  'sysres.memory_available_mb': '可用内存 (MB)',
  'disk.usage_percent': '磁盘使用率',
  'disk.free_bytes': '磁盘剩余空间',
  'time.hour': '当前小时 (0-23)',
  'time.minute': '当前分钟 (0-59)',
  'time.weekday': '星期几 (0=周日)',
  'time.day_of_month': '本月第几天',
  'time.month': '月份 (1-12)',
  'time.year': '年份',
  'process.list': '进程列表',
  'process.count': '进程数量',
  // 2025 新增插件
  'clipboard.text': '剪贴板文本内容',
  'clipboard.has_text': '剪贴板是否有文本',
  'display.primary_width': '主屏幕宽度 (px)',
  'display.primary_height': '主屏幕高度 (px)',
  'display.monitor_count': '显示器数量',
  'display.dpi': '屏幕 DPI',
  'uptime.seconds': '系统运行秒数',
  'uptime.minutes': '系统运行分钟数',
  'uptime.hours': '系统运行小时数',
  'os.computer_name': '计算机名',
  'os.user_name': '当前用户名',
  'cpu.logical_cores': '逻辑核心数',
  'cpu.physical_cores': '物理核心数',
  'cpu.architecture': 'CPU 架构',
  'locale.language': '系统语言',
  'locale.region': '区域',
  'locale.timezone': '时区',
  'locale.is_24hour': '24 小时制',
  'volume.level': '系统音量 (0-100)',
  'volume.muted': '是否静音',
  'battery_detail.voltage': '电池电压',
  'battery_detail.charge_rate': '充电速率 (mW)',
  'battery_detail.design_capacity': '设计容量 (mWh)',
  'battery_detail.full_charge_capacity': '满充容量 (mWh)',
  'battery_detail.cycle_count': '电池循环次数',
  'battery_detail.seconds_remaining': '剩余时间 (秒)',
  'network_stats.bytes_sent': '已发送字节数',
  'network_stats.bytes_received': '已接收字节数',
};

// 输出插件参数说明（帮助用户理解每个参数的作用）
const OUTPUT_PARAM_HELP = {
  'notify':        { desc:'在系统托盘区域显示桌面通知弹窗', params: { title:{type:'string',required:true,hint:'通知标题',desc:'通知的标题文字'}, message:{type:'string',required:true,hint:'通知内容',desc:'通知的主体文本'}, level:{type:'string',required:false,hint:'级别: info/warn/error',desc:'通知级别，默认 info'}, timeout:{type:'integer',required:false,hint:'显示时长(秒)',desc:'通知自动关闭的秒数，默认 5' } } },
  'power_scheme':  { desc:'切换 Windows 电源方案', params: { scheme:{type:'string',required:true,hint:'power_saver / balanced / high_performance',desc:'电源方案名称' } } },
  'prevent_sleep': { desc:'阻止或允许系统进入睡眠状态', params: { enable:{type:'string',required:true,hint:'true=阻止 / false=允许',desc:'true 阻止睡眠，false 允许睡眠' } } },
  'exec':          { desc:'执行一个外部程序或脚本', params: { cmd:{type:'string',required:true,hint:'程序路径',desc:'可执行文件或脚本的路径'}, args:{type:'string',required:false,hint:'命令行参数',desc:'传给程序的命令行参数'}, working_dir:{type:'string',required:false,hint:'工作目录',desc:'程序的工作目录，留空用默认'}, wait:{type:'string',required:false,hint:'等待完成 (true/false)',desc:'是否等待程序执行完毕，默认 false' } } },
  'volume':        { desc:'设置系统主音量', params: { level:{type:'integer',required:true,hint:'音量 0-100',desc:'目标音量值，范围 0 ~ 100' } } },
  'brightness':    { desc:'设置屏幕亮度', params: { level:{type:'integer',required:true,hint:'亮度 0-100',desc:'目标亮度值，范围 0 ~ 100' } } },
  'lock':          { desc:'锁定计算机', params: {} },
  'power_action':  { desc:'执行电源操作（关机/重启/睡眠/休眠）', params: { action:{type:'string',required:true,hint:'shutdown / restart / sleep / hibernate',desc:'要执行的电源操作' } } },
  'wallpaper':     { desc:'更换桌面壁纸', params: { path:{type:'string',required:true,hint:'图片文件路径',desc:'壁纸图片的完整路径（支持 BMP/JPG/PNG）'} } },
  'kill_process':  { desc:'终止指定 PID 的进程', params: { pid:{type:'integer',required:true,hint:'进程 PID',desc:'要终止的进程 ID' } } },
  'webhook':       { desc:'发送 HTTP 请求到指定地址', params: { url:{type:'string',required:true,hint:'请求地址',desc:'目标 URL（必填）'}, method:{type:'string',required:false,hint:'GET/POST/PUT',desc:'HTTP 方法，默认 POST'}, body:{type:'string',required:false,hint:'请求体 JSON',desc:'请求体内容（JSON 字符串）'}, headers:{type:'string',required:false,hint:'请求头(JSON对象)',desc:'自定义请求头，JSON 对象格式，如 {"Authorization":"Bearer xxx"}'} } },
  'file':          { desc:'将内容写入文本文件', params: { path:{type:'string',required:true,hint:'保存路径',desc:'目标文件路径（必填）'}, content:{type:'string',required:false,hint:'文件内容',desc:'要写入的文本内容，默认为空'}, append:{type:'string',required:false,hint:'是否追加(true/false)',desc:'true=追加到文件末尾，false=覆盖写入，默认 false' } } },
  'netadapter':    { desc:'启用或禁用网络适配器', params: { name:{type:'string',required:true,hint:'适配器名称',desc:'网络适配器的名称'}, enabled:{type:'string',required:true,hint:'true=启用 / false=禁用',desc:'true 启用，false 禁用' } } },
  'open':          { desc:'用默认程序打开文件或网址', params: { target:{type:'string',required:true,hint:'文件路径或网址',desc:'要打开的文件路径或 URL'} } },
  'window_control':{ desc:'控制窗口状态（最小化/最大化/关闭等）', params: { title:{type:'string',required:true,hint:'窗口标题',desc:'窗口标题（支持模糊匹配）'}, action:{type:'string',required:true,hint:'minimize/maximize/restore/close/focus',desc:'要执行的操作' } } },
  'http_get':      { desc:'发送 HTTP GET 请求到指定地址', params: { url:{type:'string',required:true,hint:'请求 URL',desc:'目标 URL（必填）'}, headers:{type:'string',required:false,hint:'请求头 JSON 对象',desc:'自定义请求头，如 {"Authorization":"Bearer xxx"}'}, timeout:{type:'integer',required:false,hint:'超时秒数',desc:'请求超时时间，默认 10 秒' } } },
  'http_post':     { desc:'发送 HTTP POST 请求到指定地址', params: { url:{type:'string',required:true,hint:'请求 URL',desc:'目标 URL（必填）'}, body:{type:'string',required:false,hint:'请求体内容',desc:'POST 请求体文本'}, content_type:{type:'string',required:false,hint:'Content-Type',desc:'请求内容类型，默认 application/json'}, headers:{type:'string',required:false,hint:'请求头 JSON 对象',desc:'自定义请求头，如 {"Authorization":"Bearer xxx"}'}, timeout:{type:'integer',required:false,hint:'超时秒数',desc:'请求超时时间，默认 10 秒' } } },
  'screenshot':    { desc:'截取当前屏幕并保存为图片', params: { path:{type:'string',required:false,hint:'保存路径（例: screenshot.bmp）',desc:'截图保存路径，留空自动命名' } } },
  'clipboard_set': { desc:'设置剪贴板文本内容', params: { text:{type:'string',required:true,hint:'要设置的文本',desc:'要复制到剪贴板的文本' } } },
  'delay':         { desc:'延时等待（作为多个输出动作之间的间隔）', params: { seconds:{type:'integer',required:true,hint:'等待秒数',desc:'等待的秒数，例如 5 表示等待 5 秒' } } },
};

// 状态键元数据：类型和取值范围（条件编辑时提示）
const STATE_KEY_META = {
  'power.is_charging':         { type:'bool', hint:'true=充电中 / false=未充电' },
  'wifi.is_connected':         { type:'bool', hint:'true=已连接 / false=未连接' },
  'network.is_connected':      { type:'bool', hint:'true=已连接 / false=未连接' },
  'clipboard.has_text':        { type:'bool', hint:'true=有文本 / false=无文本' },
  'volume.muted':              { type:'bool', hint:'true=静音 / false=非静音' },
  'locale.is_24hour':          { type:'bool', hint:'true=24小时制 / false=12小时制' },
  'time.is_weekend':           { type:'bool', hint:'true=周末 / false=工作日' },
  'idle.is_idle':              { type:'bool', hint:'true=空闲 / false=活动中' },
  'session.is_locked':         { type:'bool', hint:'true=已锁屏 / false=未锁屏' },
  'power.battery_percent':     { type:'percent', hint:'0-100' },
  'sysres.cpu_percent':        { type:'percent', hint:'0-100' },
  'sysres.memory_percent':     { type:'percent', hint:'0-100' },
  'disk.used_percent':         { type:'percent', hint:'0-100' },
  'wifi.signal_quality':       { type:'percent', hint:'0-100' },
  'volume.level':              { type:'percent', hint:'0-100' },
  'time.hour':                 { type:'int', hint:'0-23' },
  'time.minute':               { type:'int', hint:'0-59' },
  'time.weekday':              { type:'int', hint:'0=周日, 1-6=周一至周六' },
  'time.day_of_month':         { type:'int', hint:'1-31' },
  'time.month':                { type:'int', hint:'1-12' },
  'time.year':                 { type:'int', hint:'例: 2025' },
  'idle.seconds':              { type:'int', hint:'秒数' },
  'display.primary_width':     { type:'int', hint:'像素' },
  'display.primary_height':    { type:'int', hint:'像素' },
  'display.monitor_count':     { type:'int', hint:'显示器数量' },
  'display.dpi':               { type:'int', hint:'DPI' },
  'cpu.logical_cores':         { type:'int', hint:'逻辑核心数' },
  'cpu.physical_cores':        { type:'int', hint:'物理核心数' },
  'uptime.seconds':            { type:'int', hint:'秒数' },
  'uptime.minutes':            { type:'int', hint:'分钟数' },
  'uptime.hours':              { type:'int', hint:'小时数' },
  'battery_detail.cycle_count': { type:'int', hint:'循环次数' },
  'battery_detail.seconds_remaining': { type:'int', hint:'剩余秒数' },
  'process.count':             { type:'int', hint:'进程数量' },
  'power.ac_line_status':      { type:'int', hint:'0=电池供电 / 1=接通电源 / 255=未知' },
  'wifi.ssid':                 { type:'string', hint:'Wi-Fi 名称' },
  'wifi.bssid':                { type:'string', hint:'基站 MAC 地址' },
  'wifi.signal_quality':       { type:'percent', hint:'0-100' },
  'network.ipv4':              { type:'string', hint:'例: 192.168.1.100' },
  'network.gateway':           { type:'string', hint:'例: 192.168.1.1' },
  'network.subnet_mask':       { type:'string', hint:'例: 255.255.255.0' },
  'window.foreground.title':   { type:'string', hint:'窗口标题' },
  'clipboard.text':            { type:'string', hint:'剪贴板文本内容' },
  'os.computer_name':          { type:'string', hint:'计算机名' },
  'os.user_name':              { type:'string', hint:'用户名' },
  'locale.language':           { type:'string', hint:'例: zh-CN' },
  'locale.region':             { type:'string', hint:'例: CN' },
  'locale.timezone':           { type:'string', hint:'时区名称' },
};

function getStateKeyLabel(key) {
  return _extraStateKeyLabels[key] || STATE_KEY_LABELS[key] || '';
}

// 根据内置插件 ID 获取其声明的状态键标签（从 STATE_KEY_LABELS 按前缀匹配）
function getBuiltinStateKeyLabels(pluginId) {
  const prefix = pluginId + '.';
  const labels = {};
  // 先从硬编码的内置标签提取
  Object.keys(STATE_KEY_LABELS).forEach(k => {
    if (k.startsWith(prefix)) labels[k] = STATE_KEY_LABELS[k];
  });
  // 再从外部插件注入的标签提取
  Object.keys(_extraStateKeyLabels).forEach(k => {
    if (k.startsWith(prefix)) labels[k] = _extraStateKeyLabels[k];
  });
  return labels;
}

// 从任务的 data_sources 生成状态键下拉选项（条件编辑器使用）
function getDataSourceKeyOptions(selectedKey) {
  selectedKey = selectedKey || '';
  const dsList = (_editingTask && _editingTask.data_sources) || [];
  let html = '<option value="">— 选择数据源字段 —</option>';
  dsList.forEach((ds, i) => {
    if (!ds.state_key) return;
    const label = getStateKeyLabel(ds.state_key);
    const desc = ds.description ? ` [${ds.description}]` : '';
    const note = label ? ` (${label})` : '';
    const sel = ds.state_key === selectedKey ? ' selected' : '';
    html += `<option value="${ds.state_key}"${sel}>${ds.state_key}${note}${desc}</option>`;
  });
  if (dsList.length === 0) {
    html += '<option value="" disabled>— 请先在「输入数据」中添加数据源 —</option>';
  }
  return html;
}

function getStateKeyOptions(selectedKey) {
  selectedKey = selectedKey || '';
  let html = '<option value="">— 选择状态键 —</option>';
  const groups = {};
  _cachedStateKeys.forEach(k => {
    const prefix = k.split('.')[0];
    if (!groups[prefix]) groups[prefix] = [];
    groups[prefix].push(k);
  });
  // 按已知前缀顺序输出，其他放最后
  const knownPrefixes = ['power','wifi','network','window','idle','session','sysres','disk','time','process',
    'clipboard','display','uptime','os','cpu','locale','volume','battery_detail','network_stats'];
  for (const g of knownPrefixes) {
    if (groups[g]) {
      const groupLabel = STATE_GROUP_LABELS[g] || g;
      html += `<optgroup label="${groupLabel}">`;
      groups[g].forEach(k => {
        const label = getStateKeyLabel(k);
        const sel = k === selectedKey ? ' selected' : '';
        if (label) {
          html += `<option value="${k}"${sel}>${k} (${label})</option>`;
        } else {
          html += `<option value="${k}"${sel}>${k}</option>`;
        }
      });
      html += '</optgroup>';
    }
  }
  // 其他未分组的
  const used = new Set(Object.values(groups).flat());
  const others = _cachedStateKeys.filter(k => !used.has(k));
  if (others.length) {
    html += '<optgroup label="其他">';
    others.forEach(k => {
      const label = getStateKeyLabel(k);
      const sel = k === selectedKey ? ' selected' : '';
      if (label) {
        html += `<option value="${k}"${sel}>${k} (${label})</option>`;
      } else {
        html += `<option value="${k}"${sel}>${k}</option>`;
      }
    });
    html += '</optgroup>';
  }
  return html;
}

// ======== 打开/关闭编辑器 ========

async function openTaskEditor(taskId) {
  // 加载状态键供选择
  await refreshStateKeys();

  // 加载任务数据
  const allTasks = await apiFetch('/tasks');
  if (allTasks.error) { alert('加载任务失败: ' + allTasks.error); return; }
  const raw = allTasks[taskId];
  if (!raw) { alert('任务不存在: ' + taskId); return; }

  // 深拷贝
  _editingTask = JSON.parse(JSON.stringify(raw));
  _editingTaskId = taskId;

  // 确保条件树存在
  if (!_editingTask.condition) {
    _editingTask.condition = null;
  }

  // 填充基本信息
  document.getElementById('editor-title').textContent = `📝 编辑任务 — ${_editingTask.name || taskId}`;
  document.getElementById('edit-task-id').value = _editingTask.task_id || taskId;
  document.getElementById('edit-task-name').value = _editingTask.name || '';
  document.getElementById('edit-task-desc').value = _editingTask.description || '';
  document.getElementById('edit-task-enabled').checked = _editingTask.enabled !== false;

  // 加载规则插件列表 → 缓存供输入插件选择使用
  try {
    const rulesData = await apiFetch('/rules');
    if (!rulesData.error) {
      _cachedInputRules = [];
      Object.values(rulesData).forEach(r => {
        if (r && r.direction === 'input') {
          // 确保有 id 字段（统一 plugin_id → id）
          if (r.plugin_id && !r.id) r.id = r.plugin_id;
          _cachedInputRules.push(r);
        }
      });
      // 同时注入所有规则的 state_key_labels
      Object.values(rulesData).forEach(r => {
        if (r && r.state_key_labels) addStateKeyLabels(r.state_key_labels);
      });
    }
  } catch (_) { /* 静默 */ }

  // 加载插件列表 → 缓存供输出动作下拉 + 内置输入插件使用
  try {
    const plugData = await apiFetch('/plugins');
    if (!plugData.error && plugData.all) {
      _cachedOutputPluginList = plugData.all.filter(p => p.type === 'output');

      // 注入所有插件声明的状态键中文备注
      if (plugData.state_key_labels) {
        addStateKeyLabels(plugData.state_key_labels);
      }

      // 将内置输入插件也加入 _cachedInputRules（供数据源选择）
      const builtinInputs = plugData.all.filter(p => p.type === 'input' && p.built_in);
      builtinInputs.forEach(p => {
        // 检查是否已经被外部规则覆盖（同名插件以外部分为准）
        if (!_cachedInputRules.find(r => r.plugin_id === p.id)) {
          _cachedInputRules.push({
            id: p.id,
            plugin_id: p.id,
            name: p.name,
            description: p.description || '',
            built_in: true,
            // 内置插件的 state_key_labels 从前端 STATE_KEY_LABELS 按前缀提取
            state_key_labels: getBuiltinStateKeyLabels(p.id),
          });
        }
      });
    }
  } catch (_) { /* 静默失败 */ }
  renderDataSources();
  renderOutputs();
  renderConditionTree();

  // 显示编辑器
  document.getElementById('task-editor-overlay').style.display = 'flex';
}

function closeTaskEditor() {
  document.getElementById('task-editor-overlay').style.display = 'none';
  _editingTask = null;
  _editingTaskId = null;
}

// ======== 保存任务 ========

async function saveTaskEditor() {
  if (!_editingTask) return;

  // 从表单收集基本信息
  _editingTask.name = document.getElementById('edit-task-name').value.trim() || _editingTask.task_id;
  _editingTask.description = document.getElementById('edit-task-desc').value.trim();
  _editingTask.enabled = document.getElementById('edit-task-enabled').checked;

  // 收集数据源
  _editingTask.data_sources = collectDataSources();

  // 收集输出
  _editingTask.outputs = collectOutputs();

  // 条件树已通过直接操作 _editingTask.condition 保存在内存中

  // 发送到 API
  const res = await apiFetch(`/tasks/${encodeURIComponent(_editingTaskId)}`, {
    method: 'PUT',
    body: JSON.stringify(_editingTask),
  });
  if (res.error) { alert('保存失败: ' + res.error); return; }
  alert('✅ 保存成功！');
  closeTaskEditor();
  loadTasks();
}

// ======== 条件树渲染 ========

function renderConditionTree() {
  const container = document.getElementById('condition-tree');
  container.innerHTML = '';

  const cond = _editingTask.condition;
  if (!cond) {
    container.innerHTML = `<div class="hint" style="padding:8px">无条件（始终触发）
      <button class="btn btn-xs btn-xs-add" onclick="setRootGroup('AND')">设为 AND 组</button>
      <button class="btn btn-xs btn-xs-add" onclick="setRootGroup('OR')">设为 OR 组</button>
      <button class="btn btn-xs btn-xs-add" onclick="addRootLeaf()">➕ 添加条件</button>
    </div>`;
    return;
  }

  // 递归渲染根节点
  const rootEl = renderCondNode(cond, 'root');
  container.appendChild(rootEl);
}

function setRootGroup(op) {
  _editingTask.condition = { logical_operator: op, conditions: [] };
  renderConditionTree();
}

function addRootLeaf() {
  _editingTask.condition = { type: 'state', state_key: '', operator: 'equals', value: '' };
  renderConditionTree();
}

// 递归渲染条件节点
function renderCondNode(node, path) {
  const wrapper = document.createElement('div');
  wrapper.className = 'cond-node';
  wrapper.dataset.path = path;

  if (node.logical_operator) {
    // 逻辑组节点
    const isRoot = (path === 'root');
    wrapper.className = 'cond-node cond-group';
    wrapper.draggable = !isRoot;

    // 组头
    const header = document.createElement('div');
    header.className = 'cond-group-header';

    // 拖拽手柄
    const handle = document.createElement('span');
    handle.className = 'drag-handle';
    handle.textContent = isRoot ? '' : '⠿';
    handle.draggable = false;
    if (!isRoot) {
      handle.addEventListener('mousedown', () => { wrapper.draggable = true; });
      handle.addEventListener('mouseup', () => { wrapper.draggable = false; });
      handle.addEventListener('dragstart', (e) => handleDragStart(e, path));
    }
    header.appendChild(handle);

    // 逻辑运算符中文显示
    const badge = document.createElement('span');
    badge.className = `logical-badge badge-${node.logical_operator.toLowerCase()}`;
    badge.textContent = LOGICAL_LABELS[node.logical_operator] || node.logical_operator;
    badge.title = '点击切换 且 / 或 / 非';
    badge.addEventListener('click', () => toggleLogicalOp(node, badge));
    header.appendChild(badge);

    header.appendChild(document.createTextNode(' '));

    // 添加子条件按钮
    const addBtn = document.createElement('button');
    addBtn.className = 'btn btn-xs btn-xs-add';
    addBtn.textContent = '➕ 条件';
    addBtn.addEventListener('click', () => addChildCond(path, 'leaf'));
    header.appendChild(addBtn);

    const addGroupBtn = document.createElement('button');
    addGroupBtn.className = 'btn btn-xs';
    addGroupBtn.textContent = '➕ 组';
    addGroupBtn.style.background = '#9b59b6';
    addGroupBtn.addEventListener('click', () => addChildCond(path, 'group_and'));
    header.appendChild(addGroupBtn);

    // 删除组按钮
    const delBtn = document.createElement('button');
    delBtn.className = 'btn btn-xs';
    if (isRoot) {
      delBtn.textContent = '🗑 清空条件';
      delBtn.title = '清空所有条件，恢复为「无条件（始终触发）」';
      delBtn.style.marginLeft = 'auto';
      delBtn.style.background = '#e74c3c';
      delBtn.style.color = '#fff';
    } else {
      delBtn.textContent = '✕';
      delBtn.style.marginLeft = 'auto';
    }
    delBtn.addEventListener('click', () => deleteCondNode(path));
    header.appendChild(delBtn);

    wrapper.appendChild(header);

    // 子节点容器
    const childrenDiv = document.createElement('div');
    childrenDiv.className = 'cond-group-children';
    childrenDiv.addEventListener('dragover', handleDragOver);
    childrenDiv.addEventListener('drop', (e) => handleDrop(e, path));

    const children = node.conditions || [];
    if (children.length === 0) {
      const emptyHint = document.createElement('div');
      emptyHint.className = 'hint';
      emptyHint.style.padding = '4px';
      emptyHint.textContent = '（空组，点击上方按钮添加子条件）';
      childrenDiv.appendChild(emptyHint);
    } else {
      children.forEach((child, idx) => {
        const childPath = path + '.c.' + idx;
        const childEl = renderCondNode(child, childPath);
        childrenDiv.appendChild(childEl);
      });
    }

    wrapper.appendChild(childrenDiv);
  } else {
    // 叶子条件节点
    wrapper.className = 'cond-node cond-leaf';
    wrapper.draggable = true;

    const handle = document.createElement('span');
    handle.className = 'drag-handle';
    handle.textContent = '⠿';
    handle.draggable = false;
    handle.addEventListener('mousedown', () => { wrapper.draggable = true; });
    handle.addEventListener('mouseup', () => { wrapper.draggable = false; });
    handle.addEventListener('dragstart', (e) => handleDragStart(e, path));
    wrapper.appendChild(handle);

    // type 选择
    const typeSel = document.createElement('select');
    typeSel.innerHTML = '<option value="state">状态条件</option><option value="constant">常量</option>';
    typeSel.value = node.type || 'state';
    typeSel.addEventListener('change', () => {
      node.type = typeSel.value;
      renderConditionTree();
    });
    wrapper.appendChild(typeSel);

    // 删除按钮（在所有叶子节点中共享）
    const delBtn = document.createElement('button');
    delBtn.className = 'btn btn-xs btn-xs-del';
    delBtn.textContent = '✕';
    delBtn.style.marginLeft = 'auto';
    delBtn.addEventListener('click', () => deleteCondNode(path));

    if (node.type === 'constant') {
      // 常量值
      const cb = document.createElement('input');
      cb.type = 'checkbox';
      cb.checked = node.value === true;
      cb.addEventListener('change', () => { node.value = cb.checked; renderConditionTree(); });
      wrapper.appendChild(cb);
      wrapper.appendChild(document.createTextNode(' 是'));
      wrapper.appendChild(delBtn);
    } else {
      // 状态键（只显示数据源中选的字段）
      const keySel = document.createElement('select');
      keySel.innerHTML = getDataSourceKeyOptions(node.state_key);
      keySel.value = node.state_key || '';
      keySel.addEventListener('change', () => { node.state_key = keySel.value; });
      wrapper.appendChild(keySel);

      // 运算符
      const opSel = document.createElement('select');
      let opHtml = '';
      for (const g in OPERATOR_GROUPS) {
        opHtml += `<optgroup label="${g}">`;
        OPERATOR_GROUPS[g].forEach(o => {
          opHtml += `<option value="${o}">${OPERATOR_LABELS[o]||o}</option>`;
        });
        opHtml += '</optgroup>';
      }
      opSel.innerHTML = opHtml;
      opSel.value = node.operator || 'equals';
      opSel.addEventListener('change', () => { node.operator = opSel.value; });
      wrapper.appendChild(opSel);

      // 值
      const valInput = document.createElement('input');
      valInput.type = 'text';
      valInput.placeholder = '值';
      valInput.value = node.value !== undefined && node.value !== null ? String(node.value) : '';
      valInput.style.width = '150px';
      valInput.addEventListener('change', () => {
        const v = valInput.value.trim();
        if (v === 'true') node.value = true;
        else if (v === 'false') node.value = false;
        else if (!isNaN(v) && v !== '') node.value = Number(v);
        else node.value = v;
      });
      wrapper.appendChild(valInput);

      // 类型/范围提示
      const hintSpan = document.createElement('span');
      hintSpan.className = 'cond-hint';
      hintSpan.style.fontSize = '11px';
      hintSpan.style.color = '#999';
      const updateHint = () => {
        const meta = STATE_KEY_META[node.state_key];
        hintSpan.textContent = meta ? meta.hint : '';
      };
      updateHint();
      keySel.addEventListener('change', updateHint);
      wrapper.appendChild(hintSpan);

      // 阈值 / 稳定（下拉二选一 + 数值输入），居右显示
      const threshContainer = document.createElement('span');
      threshContainer.style.cssText = 'display:inline-flex;align-items:center;gap:2px';

      const threshMode = document.createElement('select');
      threshMode.style.fontSize = '11px';
      threshMode.style.padding = '1px 2px';
      let selectedMode = '0';
      if (node.threshold > 0) selectedMode = 'threshold';
      else if (node.stable_seconds > 0) selectedMode = 'stable';
      threshMode.innerHTML =
        '<option value="0">—</option>' +
        '<option value="threshold"' + (selectedMode === 'threshold' ? ' selected' : '') + '>阈值</option>' +
        '<option value="stable"' + (selectedMode === 'stable' ? ' selected' : '') + '>稳定</option>';
      threshMode.addEventListener('change', () => {
        const v = parseInt(threshVal.value) || 0;
        node.threshold = threshMode.value === 'threshold' ? v : 0;
        node.stable_seconds = threshMode.value === 'stable' ? v : 0;
        if (threshMode.value === '0') threshVal.value = '';
      });
      threshContainer.appendChild(threshMode);

      const threshVal = document.createElement('input');
      threshVal.type = 'number';
      threshVal.min = '0';
      threshVal.max = '999';
      threshVal.placeholder = '次数/秒';
      threshVal.title = '连续次数或持续秒数';
      threshVal.style.width = '83px';
      threshVal.style.fontSize = '11px';
      threshVal.value = node.threshold > 0 ? node.threshold : (node.stable_seconds > 0 ? node.stable_seconds : '');
      threshVal.addEventListener('change', () => {
        const v = parseInt(threshVal.value) || 0;
        if (threshMode.value === 'threshold') {
          node.threshold = v;
          node.stable_seconds = 0;
        } else if (threshMode.value === 'stable') {
          node.stable_seconds = v;
          node.threshold = 0;
        }
        if (v === 0) { node.threshold = 0; node.stable_seconds = 0; }
      });
      threshContainer.appendChild(threshVal);
      // 右侧分组：阈值容器 + 删除按钮，整体居右
      const rightGroup = document.createElement('span');
      rightGroup.style.cssText = 'display:inline-flex;align-items:center;gap:4px;margin-left:auto';
      rightGroup.appendChild(threshContainer);
      rightGroup.appendChild(delBtn);
      wrapper.appendChild(rightGroup);
    }
  }

  return wrapper;
}

// ======== 条件树操作 ========

// 根据 path 查找节点
function findCondNode(path) {
  if (path === 'root') return _editingTask.condition;
  const parts = path.split('.c.');
  let node = _editingTask.condition;
  for (let i = 1; i < parts.length; i++) {
    const idx = parseInt(parts[i]);
    if (!node || !node.conditions || idx >= node.conditions.length) return null;
    node = node.conditions[idx];
  }
  return node;
}

// 根据 path 查找父节点和索引
function findCondParent(path) {
  if (path === 'root' || !path.includes('.c.')) return null;
  const lastDot = path.lastIndexOf('.c.');
  const parentPath = lastDot === 0 ? 'root' : path.substring(0, lastDot);
  const idx = parseInt(path.substring(lastDot + 3));
  const parent = findCondNode(parentPath);
  return { parent, idx, parentPath };
}

// 逻辑运算符中文标签
const LOGICAL_LABELS = { 'AND': '且', 'OR': '或', 'NOT': '非' };

// 切换逻辑运算符
function toggleLogicalOp(node, badgeEl) {
  const ops = ['AND', 'OR', 'NOT'];
  const cur = ops.indexOf(node.logical_operator);
  node.logical_operator = ops[(cur + 1) % 3];
  // 更新 badge 样式
  badgeEl.className = `logical-badge badge-${node.logical_operator.toLowerCase()}`;
  badgeEl.textContent = LOGICAL_LABELS[node.logical_operator] || node.logical_operator;
}

// 添加子条件
function addChildCond(parentPath, type) {
  const parent = findCondNode(parentPath);
  if (!parent || !parent.conditions) {
    if (parent) parent.conditions = [];
    else return;
  }
  if (type === 'leaf') {
    parent.conditions.push({ type: 'state', state_key: '', operator: 'equals', value: '' });
  } else if (type === 'group_and') {
    parent.conditions.push({ logical_operator: 'AND', conditions: [] });
  }
  renderConditionTree();
}

// 删除条件节点
function deleteCondNode(path) {
  if (path === 'root') {
    _editingTask.condition = null;
    renderConditionTree();
    return;
  }
  const result = findCondParent(path);
  if (!result || !result.parent || !result.parent.conditions) return;
  result.parent.conditions.splice(result.idx, 1);
  renderConditionTree();
}

// ======== 拖拽排序（同级） ========

let _dragPath = null;

function handleDragStart(e, path) {
  _dragPath = path;
  e.dataTransfer.effectAllowed = 'move';
  // 创建一个透明的拖拽图像
  const ghost = document.createElement('div');
  ghost.style.opacity = '0';
  document.body.appendChild(ghost);
  e.dataTransfer.setDragImage(ghost, 0, 0);
  setTimeout(() => document.body.removeChild(ghost), 0);
}

function handleDragOver(e) {
  e.preventDefault();
  e.dataTransfer.dropEffect = 'move';
}

function handleDrop(e, targetParentPath) {
  e.preventDefault();
  if (!_dragPath || _dragPath === targetParentPath) {
    _dragPath = null;
    return;
  }

  // 获取拖拽节点的父路径和索引
  const srcResult = findCondParent(_dragPath);
  if (!srcResult) { _dragPath = null; return; }

  // 目标父节点
  const targetParent = findCondNode(targetParentPath);
  if (!targetParent || !targetParent.conditions) { _dragPath = null; return; }

  // 只能在同一父节点下排序
  if (srcResult.parentPath !== targetParentPath) {
    _dragPath = null;
    return;
  }

  // 获取鼠标在子节点列表中的位置
  const childrenDiv = e.currentTarget;
  const childNodes = childrenDiv.querySelectorAll(':scope > .cond-node');
  let dropIdx = childNodes.length;
  for (let i = 0; i < childNodes.length; i++) {
    const rect = childNodes[i].getBoundingClientRect();
    const mid = rect.top + rect.height / 2;
    if (e.clientY < mid) { dropIdx = i; break; }
  }

  // 执行移动
  const [removed] = targetParent.conditions.splice(srcResult.idx, 1);
  // 调整目标索引（如果从前面移除，目标索引-1）
  let targetIdx = dropIdx;
  if (srcResult.idx < targetIdx) targetIdx--;
  targetParent.conditions.splice(targetIdx, 0, removed);

  _dragPath = null;
  renderConditionTree();
}

// ======== 数据源管理（新设计：选输入插件 → 选数据字段）========

function renderDataSources() {
  const container = document.querySelector('#edit-datasources');
  const sources = _editingTask.data_sources || [];
  if (sources.length === 0) {
    container.innerHTML = '<tr><td class="hint">暂无输入数据 — 点击「➕ 添加」选择输入插件和字段</td></tr>';
    return;
  }
  let html = '';
  // 搜索框
  html += `<tr><td><div class="data-row" style="margin-bottom:4px">${renderInputSearch()}</div></td></tr>`;
  sources.forEach((ds, idx) => {
    const pluginOpts = buildCategorizedOptions(INPUT_PLUGIN_CATEGORIES, _cachedInputRules, ds.plugin_id, _inputSearchText);

    const pluginKeys = getInputPluginStateKeys(ds.plugin_id);
    const isCustomKey = ds.state_key && !pluginKeys.includes(ds.state_key);
    const useInput = ds._custom || isCustomKey;
    const configHtml = renderDataSourceConfig(idx, ds.plugin_id, ds.params);

    if (useInput) {
      html += `<tr><td>
        <div class="data-row">
          <select class="ds-plugin-select" data-ds-idx="${idx}" style="flex:1.2"
            onchange="onDataSourcePluginChange(${idx}, this)">
            ${pluginOpts}
          </select>
          <input type="text" class="ds-key-input" data-ds-idx="${idx}" data-ds-field="state_key"
            value="${escHtml(ds.state_key||'')}" placeholder="${getStateKeyPlaceholder(ds.plugin_id)}"
            style="flex:1.2;padding:2px 6px;border:1px solid #ddd;border-radius:3px;font-size:12px" />
          <span class="ds-hint" id="ds-hint-${idx}" style="font-size:11px;color:#999;white-space:nowrap;min-width:60px">${getStateKeyHint(ds.state_key)}</span>
          <input type="text" value="${escHtml(ds.description||'')}" placeholder="备注" style="flex:0.6" data-ds-idx="${idx}" data-ds-field="description" />
          <button class="btn btn-xs btn-xs-del" onclick="deleteDataSource(${idx})">✕</button>
        </div>
        ${configHtml}
      </td></tr>`;
    } else {
      let keyOpts = '<option value="">— 选数据字段 —</option>';
      pluginKeys.forEach(k => {
        const label = getStateKeyLabel(k) || '';
        const sel = ds.state_key === k ? ' selected' : '';
        const display = label ? `${k} (${label})` : k;
        keyOpts += `<option value="${k}"${sel}>${display}</option>`;
      });
      keyOpts += `<option value="__custom__">✏️ 自定义...</option>`;

      html += `<tr><td>
        <div class="data-row">
          <select class="ds-plugin-select" data-ds-idx="${idx}" style="flex:1.2"
            onchange="onDataSourcePluginChange(${idx}, this)">
            ${pluginOpts}
          </select>
          <select class="ds-key-select" data-ds-idx="${idx}" style="flex:1.2"
            onchange="onDataSourceKeySelect(${idx}, this)">
            ${keyOpts}
          </select>
          <span class="ds-hint" id="ds-hint-${idx}" style="font-size:11px;color:#999;white-space:nowrap;min-width:60px">${getStateKeyHint(ds.state_key)}</span>
          <input type="text" value="${escHtml(ds.description||'')}" placeholder="备注" style="flex:0.6" data-ds-idx="${idx}" data-ds-field="description" />
          <button class="btn btn-xs btn-xs-del" onclick="deleteDataSource(${idx})">✕</button>
        </div>
        ${configHtml}
      </td></tr>`;
      if (ds.state_key) {
        html = html.replace(
          `id="ds-hint-${idx}"`,
          `id="ds-hint-${idx}" data-init-hint="${escHtml(getStateKeyHint(ds.state_key))}"`
        );
      }
    }
  });
  container.innerHTML = html;

  container.querySelectorAll('[data-ds-idx][data-ds-field]').forEach(el => {
    el.addEventListener('change', () => {
      const idx = parseInt(el.dataset.dsIdx);
      const field = el.dataset.dsField;
      if (!_editingTask.data_sources[idx]) return;
      _editingTask.data_sources[idx][field] = el.value;
    });
  });
  // 绑定 params 变更
  bindDataSourceParams();
}

// 渲染插件特定配置参数
function renderDataSourceConfig(idx, pluginId, params) {
  if (!pluginId) return '';
  params = params || {};
  switch (pluginId) {
    case 'http_request':
      return `<div class="data-row" style="margin-top:2px;padding-left:4px">
        <span style="font-size:11px;color:#888;width:40px">URL:</span>
        <input type="text" class="ds-param-input" data-ds-idx="${idx}" data-param-key="url"
          value="${escHtml(params.url||'')}" placeholder="https://example.com/api" style="flex:2;padding:2px 6px;border:1px solid #ddd;border-radius:3px;font-size:12px" />
        <span style="font-size:11px;color:#888;width:50px">超时:</span>
        <input type="number" class="ds-param-input" data-ds-idx="${idx}" data-param-key="timeout"
          value="${params.timeout||10}" min="1" max="120" style="width:50px;padding:2px 4px;border:1px solid #ddd;border-radius:3px;font-size:12px" />
      </div>`;
    case 'network_detect':
      return `<div class="data-row" style="margin-top:2px;padding-left:4px">
        <span style="font-size:11px;color:#888;width:45px">目标:</span>
        <input type="text" class="ds-param-input" data-ds-idx="${idx}" data-param-key="target"
          value="${escHtml(params.target||'')}" placeholder="8.8.8.8" style="flex:1;padding:2px 6px;border:1px solid #ddd;border-radius:3px;font-size:12px" />
        <select class="ds-param-input" data-ds-idx="${idx}" data-param-key="mode" style="padding:2px 4px;font-size:11px">
          <option value="ping" ${params.mode==='ping'?'selected':''}>Ping</option>
          <option value="tcp" ${params.mode==='tcp'?'selected':''}>TCP</option>
          <option value="http" ${params.mode==='http'?'selected':''}>HTTP</option>
        </select>
        <input type="number" class="ds-param-input" data-ds-idx="${idx}" data-param-key="port"
          value="${params.port||''}" placeholder="端口" style="width:55px;padding:2px 4px;border:1px solid #ddd;border-radius:3px;font-size:12px" />
      </div>`;
    case 'file_monitor':
      const paths = Array.isArray(params.paths) ? params.paths.join(', ') : (params.paths||'');
      return `<div class="data-row" style="margin-top:2px;padding-left:4px">
        <span style="font-size:11px;color:#888;width:45px">路径:</span>
        <input type="text" class="ds-param-input" data-ds-idx="${idx}" data-param-key="paths"
          value="${escHtml(paths)}" placeholder="C:/logs (逗号分隔多路径)" style="flex:2;padding:2px 6px;border:1px solid #ddd;border-radius:3px;font-size:12px" />
        <label style="font-size:11px"><input type="checkbox" class="ds-param-input" data-ds-idx="${idx}" data-param-key="recursive" ${params.recursive?'checked':''} /> 递归</label>
      </div>`;
    default:
      return '';
  }
}

// 绑定 params 输入变更
function bindDataSourceParams() {
  document.querySelectorAll('.ds-param-input').forEach(el => {
    el.addEventListener('change', () => {
      const idx = parseInt(el.dataset.dsIdx);
      const key = el.dataset.paramKey;
      if (isNaN(idx) || !key) return;
      const ds = _editingTask.data_sources[idx];
      if (!ds) return;
      if (!ds.params) ds.params = {};
      if (el.type === 'checkbox') {
        ds.params[key] = el.checked;
      } else if (el.type === 'number') {
        ds.params[key] = parseFloat(el.value) || 0;
      } else {
        ds.params[key] = el.value;
      }
    });
  });
}

// 获取输入插件的可用数据字段（从 state_key_labels 或 functions 输出）
function getInputPluginStateKeys(pluginId) {
  if (!pluginId) return [];
  const rule = (_cachedInputRules || []).find(r => r.plugin_id === pluginId);
  if (!rule) return [];

  // 优先从 state_key_labels 获取
  const labels = rule.state_key_labels;
  if (labels && typeof labels === 'object') {
    return Object.keys(labels).sort();
  }

  // 其次从 functions 的 output 获取
  const keys = [];
  if (rule.functions) {
    Object.values(rule.functions).forEach(fn => {
      if (fn.output) Object.keys(fn.output).forEach(k => keys.push(k));
    });
  }
  return keys.sort();
}

// 插件下拉变化时 → 清空并刷新数据字段下拉
function onDataSourcePluginChange(idx, sel) {
  if (!_editingTask.data_sources[idx]) return;
  const pluginId = sel.value;
  _editingTask.data_sources[idx]._custom = false;
  _editingTask.data_sources[idx].plugin_id = pluginId;
  _editingTask.data_sources[idx].state_key = '';
  renderDataSources();
}

// 数据字段下拉变化 → 更新字段值
function onDataSourceKeySelect(idx, sel) {
  if (!_editingTask.data_sources[idx]) return;
  if (sel.value === '__custom__') {
    // 标记自定义模式，重新渲染为输入框
    _editingTask.data_sources[idx]._custom = true;
    _editingTask.data_sources[idx].state_key = '';
    renderDataSources();
    return;
  }
  _editingTask.data_sources[idx]._custom = false;
  _editingTask.data_sources[idx].state_key = sel.value;
  const hint = getStateKeyHint(sel.value);
  const hintEl = document.getElementById('ds-hint-' + idx);
  if (hintEl) hintEl.textContent = hint;
}

// 获取状态键的类型/范围提示
function getStateKeyPlaceholder(pluginId) {
  if (!pluginId) return '输入完整状态键名 (例: http_in.temp)';
  const keys = getInputPluginStateKeys(pluginId);
  if (keys.length > 0) {
    return '选中下方字段 或 选「✏️ 自定义」手动输入';
  }
  return '输入完整状态键名 (例: ' + pluginId + '.xxx)';
}

function getStateKeyHint(stateKey) {
  if (!stateKey) return '';
  const meta = STATE_KEY_META[stateKey];
  return meta ? meta.hint : '';
}

function addDataSource() {
  if (!_editingTask.data_sources) _editingTask.data_sources = [];
  const firstPlugin = (_cachedInputRules || [])[0];
  _editingTask.data_sources.push({
    id: 'ds_' + Date.now().toString(36),
    plugin_id: firstPlugin ? firstPlugin.plugin_id : '',
    state_key: '',
    description: ''
  });
  renderDataSources();
}

function deleteDataSource(idx) {
  if (!_editingTask.data_sources) return;
  _editingTask.data_sources.splice(idx, 1);
  renderDataSources();
}

function collectDataSources() {
  const sources = [];
  document.querySelectorAll('#edit-datasources [data-ds-idx][data-ds-field]').forEach(el => {
    const idx = parseInt(el.dataset.dsIdx);
    const field = el.dataset.dsField;
    if (!sources[idx]) sources[idx] = { id: '' };
    sources[idx][field] = el.value;
  });
  // 复制 params、state_key 和 custom 标记
  sources.forEach((s, idx) => {
    if (!s) return;
    if (!s.id) s.id = 'ds_' + Date.now().toString(36);
    if (_editingTask.data_sources[idx]) {
      // 保留 plugin_id（用于下次编辑时下拉框选中）
      if (_editingTask.data_sources[idx].plugin_id && !s.plugin_id) {
        s.plugin_id = _editingTask.data_sources[idx].plugin_id;
      }
      if (_editingTask.data_sources[idx].params) {
        s.params = JSON.parse(JSON.stringify(_editingTask.data_sources[idx].params));
      }
      // 兜底复制 state_key（非自定义模式下拉框无 data-ds-field，DOM 读不到）
      if (_editingTask.data_sources[idx].state_key && !s.state_key) {
        s.state_key = _editingTask.data_sources[idx].state_key;
      }
    }
  });
  return sources.filter(s => s && s.state_key);
}

// 缓存输出插件列表（在下拉中使用）
let _cachedOutputPluginList = [];
// 缓存输入规则插件列表（在数据源中使用）
let _cachedInputRules = [];

// ======== 输出动作管理 ========

function renderOutputs() {
  const container = document.querySelector('#edit-outputs');
  const outputs = _editingTask.outputs || [];
  if (outputs.length === 0) {
    container.innerHTML = '<tr><td class="hint">暂无输出动作</td></tr>';
    return;
  }
  // 搜索框
  let html = '<tr><td><div class="data-row" style="margin-bottom:4px">' + renderOutputSearch() + '</div></td></tr>';
  outputs.forEach((out, idx) => {
    const pluginOpts = buildCategorizedOptions(OUTPUT_PLUGIN_CATEGORIES, _cachedOutputPluginList, out.plugin_id, _outputSearchText);

    // 参数行
    let paramsHtml = '';
    const pMap = (out.params && typeof out.params === 'object') ? out.params : {};
    const pKeys = Object.keys(pMap);
    if (pKeys.length) {
      pKeys.forEach(k => {
        const v = pMap[k] !== null && pMap[k] !== undefined ? String(pMap[k]) : '';
        const help = (OUTPUT_PARAM_HELP[out.plugin_id] || {}).params || {};
        const valPlaceholder = (help[k] && help[k].hint) ? help[k].hint : '参数值';
        // 自动生成的临时键名不显示在输入框中
        const displayKey = k.startsWith('_auto_') ? '' : k;
        paramsHtml += `<div class="param-row">
          <input type="text" value="${escHtml(displayKey)}" placeholder="参数名" class="param-key" data-out-idx="${idx}" />
          <input type="text" value="${escHtml(v)}" placeholder="${escHtml(valPlaceholder)}" class="param-val" data-out-idx="${idx}" />
          <button class="btn btn-xs btn-xs-del" onclick="removeOutputParam(${idx}, this)">✕</button>
        </div>`;
      });
    }

    html += `<tr><td>
      <div class="data-row" style="flex-wrap:wrap">
        <input type="text" value="${escHtml(out.id||'')}" placeholder="ID" style="flex:0.25;min-width:50px" title="动作标识（用于日志）" data-out-idx="${idx}" data-out-field="id" />
        <select data-out-idx="${idx}" data-out-field="plugin_id" style="flex:0.7;min-width:100px" onchange="onOutputPluginChange(${idx}, this)">
          ${pluginOpts}
        </select>
        <button class="btn btn-xs" onclick="showOutputParamHelp(${idx})" title="查看当前输出插件的参数说明" ${out.plugin_id ? '' : 'disabled'}>📖</button>
        <select data-out-idx="${idx}" data-out-field="trigger_on" style="flex:0.4;min-width:70px;font-size:11px">
          <option value="true" ${out.trigger_on==='true'?'selected':''}>成立时</option>
          <option value="false" ${out.trigger_on==='false'?'selected':''}>不成立时</option>
          <option value="both" ${out.trigger_on==='both'?'selected':''}>变化时</option>
        </select>
        ${paramsHtml}
        <button class="btn btn-xs" onclick="addOutputParam(${idx})" title="添加参数">➕ 参数</button>
        <button class="btn btn-xs btn-xs-del" onclick="deleteOutput(${idx})">✕</button>
      </div>
    </td></tr>`;
  });
  container.innerHTML = html;
}

function addOutput() {
  if (!_editingTask.outputs) _editingTask.outputs = [];
  _editingTask.outputs.push({ id: 'out_' + Date.now().toString(36), type: 'builtin', plugin_id: '', params: {} });
  renderOutputs();
}

function deleteOutput(idx) {
  if (!_editingTask.outputs) return;
  _editingTask.outputs.splice(idx, 1);
  renderOutputs();
}

let _paramAutoId = 0;

function addOutputParam(idx) {
  if (!_editingTask.outputs[idx]) return;
  if (!_editingTask.outputs[idx].params) _editingTask.outputs[idx].params = {};
  // 使用自增ID生成临时键（不显示给用户，保存时跳过）
  _paramAutoId++;
  _editingTask.outputs[idx].params['_auto_' + _paramAutoId] = '';
  renderOutputs();
}

function removeOutputParam(idx, btn) {
  if (!_editingTask.outputs[idx]) return;
  const keyInput = btn.parentElement.querySelector('.param-key');
  if (keyInput) {
    const key = keyInput.value.trim();
    if (key) {
      delete _editingTask.outputs[idx].params[key];
    } else {
      // 参数名为空，删除一个自动生成的临时键
      for (const k in _editingTask.outputs[idx].params) {
        if (k.startsWith('_auto_')) {
          delete _editingTask.outputs[idx].params[k];
          break;
        }
      }
    }
  }
  renderOutputs();
}


function collectOutputs() {
  const outputs = [];
  // 收集普通字段（id, type, plugin_id）
  document.querySelectorAll('#edit-outputs [data-out-idx][data-out-field]').forEach(el => {
    const idx = parseInt(el.dataset.outIdx);
    const field = el.dataset.outField;
    if (!outputs[idx]) outputs[idx] = { type: 'builtin', trigger_on: 'true', params: {} };
    outputs[idx][field] = el.value;
  });
  // 收集参数行
  document.querySelectorAll('#edit-outputs .param-row').forEach(row => {
    const keyInput = row.querySelector('.param-key');
    const valInput = row.querySelector('.param-val');
    if (!keyInput) return;
    const idx = parseInt(keyInput.dataset.outIdx);
    if (isNaN(idx)) return;
    if (!outputs[idx]) outputs[idx] = { type: 'builtin', trigger_on: 'true', params: {} };
    if (!outputs[idx].params) outputs[idx].params = {};
    const key = keyInput.value.trim();
    if (key) {
      outputs[idx].params[key] = valInput.value;
    }
  });
  // 清理 params 中的自动生成临时键
  outputs.forEach(o => {
    if (!o || !o.params) return;
    Object.keys(o.params).forEach(k => {
      if (k.startsWith('_auto_')) delete o.params[k];
    });
  });
  return outputs.filter(o => o && o.id);
}

// 当插件下拉被手动修改时，同步更新 data 并重新渲染参数提示
function onOutputPluginChange(idx, sel) {
  if (!_editingTask.outputs[idx]) return;
  _editingTask.outputs[idx].plugin_id = sel.value;
  // 如果插件为空则清空参数
  if (!sel.value && _editingTask.outputs[idx].params) {
    _editingTask.outputs[idx].params = {};
    renderOutputs();
  }
}

// 显示输出插件参数说明弹窗
function showOutputParamHelp(idx) {
  if (!_editingTask.outputs[idx] || !_editingTask.outputs[idx].plugin_id) return;
  const pluginId = _editingTask.outputs[idx].plugin_id;
  const help = OUTPUT_PARAM_HELP[pluginId];
  if (!help) {
    alert('该插件暂无详细说明');
    return;
  }

  // 从插件列表元数据中查找名称
  const meta = (_cachedOutputPluginList || []).find(p => p.id === pluginId);
  const pluginName = meta ? meta.name : pluginId;

  const overlay = document.getElementById('modal-overlay');
  const box = document.getElementById('modal-box');
  overlay.style.zIndex = '2000';
  box.querySelector('.modal-header span').textContent = `📖 ${pluginName} — 参数说明`;

  // 构建参数表格
  let paramsHtml = '';
  const pMap = help.params || {};
  const pKeys = Object.keys(pMap);
  if (pKeys.length) {
    paramsHtml = '<table style="width:100%;border-collapse:collapse;font-size:13px">' +
      '<tr style="background:#f0f4f8">' +
      '<th style="padding:6px 8px;border:1px solid #ddd;text-align:left">参数名</th>' +
      '<th style="padding:6px 8px;border:1px solid #ddd;text-align:left">类型</th>' +
      '<th style="padding:6px 8px;border:1px solid #ddd;text-align:left">必填</th>' +
      '<th style="padding:6px 8px;border:1px solid #ddd;text-align:left">说明</th>' +
      '</tr>';
    pKeys.forEach(k => {
      const p = pMap[k];
      const type = p.type || 'string';
      const required = p.required ? '✅ 是' : '❌ 否';
      const desc = p.desc || '';
      const hint = p.hint ? `<br><code style="font-size:11px;color:#888">示例：${escHtml(p.hint)}</code>` : '';
      paramsHtml += `<tr>
        <td style="padding:4px 8px;border:1px solid #ddd;font-weight:600"><code>${escHtml(k)}</code></td>
        <td style="padding:4px 8px;border:1px solid #ddd">${type}</td>
        <td style="padding:4px 8px;border:1px solid #ddd">${required}</td>
        <td style="padding:4px 8px;border:1px solid #ddd;font-size:12px">${desc}${hint}</td>
      </tr>`;
    });
    paramsHtml += '</table>';
  } else {
    paramsHtml = '<p style="color:#999;padding:8px 0">该插件无需额外参数</p>';
  }

  box.querySelector('.modal-body').innerHTML = `
<div style="font-size:13px;line-height:1.7">
  <p style="margin-bottom:8px"><strong>插件说明：</strong>${escHtml(help.desc || '')}</p>
  <hr style="margin:8px 0">
  <h4 style="margin:8px 0 6px">参数列表</h4>
  ${paramsHtml}
  <hr style="margin:12px 0">
  <p style="color:#888;font-size:12px">在插件的参数行中填入键值对即可配置。参数名需与上表完全一致。</p>
</div>`;
  overlay.style.display = 'flex';
}

// 数据源插件搜索
function onDataSourceSearch() {
  const input = document.querySelector('.input-search-box');
  _inputSearchText = input ? input.value.trim() : '';
  renderDataSources();
}

// 输出插件搜索
function onOutputSearch() {
  const input = document.querySelector('.output-search-box');
  _outputSearchText = input ? input.value.trim() : '';
  renderOutputs();
}

async function loadLogs() {
  const taskId = document.getElementById('log-task-id').value.trim();
  const path = taskId ? `/logs/${taskId}` : '/logs/system';
  const data = await apiFetch(path);
  if (data.error) {
    document.getElementById('log-viewer').textContent = '加载失败: ' + data.error;
    return;
  }
  document.getElementById('log-viewer').textContent = JSON.stringify(data, null, 2);
}

// ============================================================================
// 插件市场
// ============================================================================

let marketPlugins = [];

async function loadMarket() {
  await loadRepos();
  await loadMarketPlugins();
}

async function loadRepos() {
  const data = await apiFetch('/market/repos');
  const container = document.getElementById('repo-list');
  if (data.error) { container.innerHTML = `<div class="error">${data.error}</div>`; return; }
  if (!Array.isArray(data) || data.length === 0) {
    container.innerHTML = '<div class="hint">' + t('market.no_repos') + '</div>';
    return;
  }
  let html = '<table><tr><th>名称</th><th>URL</th><th>平台</th></tr>';
  data.forEach(r => { html += `<tr><td>${r.name}</td><td>${r.url}</td><td>${r.platform}</td></tr>`; });
  html += '</table>';
  container.innerHTML = html;
}

async function addRepo() {
  const url = document.getElementById('market-repo-url').value.trim();
  if (!url) return;
  await apiFetch('/market/repos', {
    method: 'POST',
    body: JSON.stringify({ url, name: url.split('/').slice(-2).join('/') }),
  });
  document.getElementById('market-repo-url').value = '';
  loadRepos();
  loadMarketPlugins();
}

async function loadMarketPlugins() {
  const data = await apiFetch('/market/plugins');
  const container = document.getElementById('market-plugins');
  if (data.error) { container.innerHTML = `<div class="error">${data.error}</div>`; return; }
  if (!Array.isArray(data) || data.length === 0) {
    container.innerHTML = '<div class="hint">' + t('market.no_plugins') + '</div>';
    return;
  }
  marketPlugins = data;
  let html = '';
  data.forEach(p => {
    html += `<div class="plugin-card">
      <div class="info"><div class="name">${p.name}</div><div class="meta">${p.type} · ${p.version || 'latest'}</div></div>
      <button class="btn btn-sm" onclick="installPlugin('${p.id}','${p.type}','${p.source}')">安装</button>
    </div>`;
  });
  container.innerHTML = html;
}

async function installPlugin(id, type, source) {
  if (!confirm(`确定安装插件 "${id}" (${type}) 吗？`)) return;
  const res = await apiFetch('/market/install', {
    method: 'POST',
    body: JSON.stringify({ source_url: source, plugin_type: type, plugin_id: id }),
  });
  if (res.error) { alert('安装失败: ' + res.error); return; }
  alert(`✅ 插件 "${id}" 安装成功！`);
  loadMarketPlugins();
}

// ============================================================================
// 进程列表模态框
// ============================================================================

async function showProcessModal() {
  const overlay = document.getElementById('modal-overlay');
  const box = document.getElementById('modal-box');
  overlay.style.display = 'flex';
  box.innerHTML = '<div style="text-align:center;padding:20px">' + t('dashboard.loading') + '</div>';

  const status = await apiFetch('/status');
  if (status.error) { box.innerHTML = '加载失败'; return; }

  let list = [];
  try { list = JSON.parse(status['process.list'] || '[]'); } catch(e) {}

  let html = `<table class="proc-table">
    <tr><th>进程名</th><th>PID</th><th>内存</th><th>路径</th><th>操作</th></tr>`;
  list.forEach(p => {
    const mem = p.memory_mb ? p.memory_mb.toFixed(1) + ' MB' : '-';
    const path = p.exec_path || '-';
    const shortName = p.name || '未知';
    html += `<tr>
      <td>${escHtml(shortName)}</td>
      <td>${p.pid}</td>
      <td>${mem}</td>
      <td style="font-size:11px;max-width:250px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${escHtml(path)}">${escHtml(path)}</td>
      <td><button class="btn btn-sm" style="background:#e74c3c" onclick="killProcess(${p.pid},'${escHtml(shortName)}')">结束</button></td>
    </tr>`;
  });
  html += '</table>';
  box.innerHTML = html;
}

// 输入数据帮助弹窗
function showDataSourceHelp() {
  const overlay = document.getElementById('modal-overlay');
  overlay.style.zIndex = '2000'; // 盖过编辑器弹窗（z-index:1000）
  const box = document.getElementById('modal-box');
  box.querySelector('.modal-header span').textContent = '📖 输入数据（Input Data）说明';
  box.querySelector('.modal-body').innerHTML = `
<div style="font-size:13px;line-height:1.7">
  <p><b>输入数据</b> 决定你的任务从哪些 <strong>输入插件</strong> 获取什么 <strong>数据字段</strong>。</p>

  <hr style="margin:12px 0">

  <h4 style="margin:8px 0">使用方式</h4>
  <ol style="padding-left:20px">
    <li><b>选择输入插件</b> — 下拉菜单中会列出所有已安装的输入型规则插件</li>
    <li><b>选择数据字段</b> — 选中插件后，下拉会显示该插件提供的可用数据字段</li>
    <li><b>添加备注（可选）</b> — 给这个数据加个中文说明方便识别</li>
  </ol>
  <p>可以添加多行，每行指定一个不同的输入插件和数据字段。</p>

  <hr style="margin:12px 0">

  <h4 style="margin:8px 0">与触发条件的关系</h4>
  <p>在「触发条件」中，状态键下拉<strong>只显示</strong>这里添加的数据字段，不会看到所有系统状态。
  这样你只需要关心你选中的数据，不会被其他插件的数据干扰。</p>

  <hr style="margin:12px 0">

  <h4 style="margin:8px 0">例子</h4>
  <table style="width:100%;border-collapse:collapse;font-size:13px">
    <tr style="background:#f0f4f8">
      <th style="padding:6px 8px;border:1px solid #ddd;text-align:left">输入插件</th>
      <th style="padding:6px 8px;border:1px solid #ddd;text-align:left">数据字段</th>
      <th style="padding:6px 8px;border:1px solid #ddd;text-align:left">备注</th>
    </tr>
    <tr>
      <td style="padding:6px 8px;border:1px solid #ddd">电源传感器</td>
      <td style="padding:6px 8px;border:1px solid #ddd">power.battery_percent</td>
      <td style="padding:6px 8px;border:1px solid #ddd">电池电量</td>
    </tr>
    <tr>
      <td style="padding:6px 8px;border:1px solid #ddd">Wi-Fi 传感器</td>
      <td style="padding:6px 8px;border:1px solid #ddd">wifi.ssid</td>
      <td style="padding:6px 8px;border:1px solid #ddd">当前 WiFi 名称</td>
    </tr>
  </table>
  <p style="margin-top:8px">然后在条件中设置：当 <code>power.battery_percent</code> 小于 30 且 <code>wifi.ssid</code> 不为空时 → 执行动作。</p>
</div>`;
  overlay.style.display = 'flex';
}

function closeModal() {
  const overlay = document.getElementById('modal-overlay');
  overlay.style.display = 'none';
  overlay.style.zIndex = ''; // 恢复默认层级
}

async function killProcess(pid, name) {
  if (!confirm(`确定结束进程 "${name}" (PID: ${pid}) 吗？`)) return;
  const res = await apiFetch('/processes/kill', {
    method: 'POST',
    body: JSON.stringify({ pid }),
  });
  if (res.error) { alert('结束失败: ' + res.error); return; }
  alert(`✅ 已结束 ${name} (PID: ${pid})`);
  showProcessModal(); // 刷新列表
}

// ============================================================================
// 自动刷新 — 仪表盘每 30 秒重新加载（实时秒级更新由 startRealtime 处理）
// ============================================================================

setInterval(() => {
  const dashboard = document.getElementById('page-dashboard');
  if (dashboard && dashboard.classList.contains('active')) {
    // 只刷新统计数字，不重绘全部（避免闪烁）
    apiFetch('/status').then(s => {
      if (!s.error) document.getElementById('stat-states').textContent = Object.keys(s).length;
    });
    apiFetch('/plugins').then(p => {
      if (!p.error) {
        const c = (p.inputs?.length||0)+(p.outputs?.length||0);
        document.getElementById('stat-plugins').textContent = c;
      }
    });
  }
}, 30000);

// ============================================================================
// 插件编辑器入口
// ============================================================================

function openPluginEditor(mode) {
  if (mode) {
    // 直接以指定模式打开
    if (typeof switchEditorPage === 'function') {
      switchEditorPage(mode);
    }
    return;
  }

  // 无模式参数：显示空的编辑器容器
  const formContainer = document.getElementById('editor-form-container');
  const actions = document.getElementById('editor-actions');
  const status = document.getElementById('editor-status');

  if (formContainer) { formContainer.style.display = 'none'; formContainer.innerHTML = ''; }
  if (actions) actions.style.display = 'none';
  if (status) { status.textContent = ''; status.style.display = 'none'; }
}

// 从侧边栏退出按钮调用
function exitPluginEditor() {
  if (typeof closePluginEditor === 'function') {
    closePluginEditor();
  } else {
    switchPage('plugins');
  }
}

// ============================================================================
// 系统设置
// ============================================================================

// 切换设置选项卡
function switchSettingsTab(tab) {
  document.querySelectorAll('.settings-tab').forEach(t => t.classList.remove('active'));
  document.querySelectorAll('.settings-panel').forEach(p => p.classList.remove('active'));
  const tabBtn = document.querySelector(`.settings-tab[onclick*="'${tab}'"]`);
  if (tabBtn) tabBtn.classList.add('active');
  const panel = document.getElementById(`settings-${tab}`);
  if (panel) panel.classList.add('active');
  if (tab === 'general') loadSettings();
}

// 加载设置（从后端读取）
async function loadSettings() {
  const data = await apiFetch('/notifications/config');
  if (!data.error && data.level) {
    const sel = document.getElementById('setting-notify-level');
    if (sel) sel.value = data.level;
  }
  // 读取 API Key 状态
  const cfg = await apiFetch('/config');
  if (!cfg.error) {
    if (cfg.api_key) {
      document.getElementById('setting-api-enabled').checked = true;
      document.getElementById('setting-api-key').value = cfg.api_key;
      document.getElementById('setting-api-key-row').style.display = 'flex';
    }
    if (cfg.auto_start !== undefined) {
      document.getElementById('setting-autostart').checked = cfg.auto_start;
    }
    if (cfg.port) {
      document.getElementById('setting-port').value = cfg.port;
    }
    // 更新关于页面的信息
    const aboutPort = document.getElementById('about-port');
    if (aboutPort) aboutPort.textContent = cfg.port || '19530';
  }
  // 读取 Web 服务状态
  if (typeof _webEnabled !== 'undefined') {
    updateWebBtnText(_webEnabled);
  }
}

function toggleAPIKeyField() {
  const enabled = document.getElementById('setting-api-enabled').checked;
  document.getElementById('setting-api-key-row').style.display = enabled ? 'flex' : 'none';
}

function setSettingsStatus(msg, type) {
  const el = document.getElementById('settings-status');
  if (el) {
    el.textContent = msg;
    el.className = 'editor-status ' + (type || '');
    el.style.display = msg ? 'inline' : 'none';
  }
}

function setAPIKeyStatus(msg, type) {
  const el = document.getElementById('api-key-status');
  if (el) {
    el.textContent = msg;
    el.className = 'editor-status ' + (type || '');
    el.style.display = msg ? 'inline' : 'none';
  }
}

// 保存 API Key
async function saveAPIKey() {
  const enabled = document.getElementById('setting-api-enabled').checked;
  const key = enabled ? document.getElementById('setting-api-key').value.trim() : '';
  const res = await apiFetch('/config', {
    method: 'PUT',
    body: JSON.stringify({ api_key: key })
  });
  if (res.error) {
    setAPIKeyStatus('保存失败: ' + res.error, 'error');
  } else {
    setAPIKeyStatus('✅ API Key 已保存', 'success');
  }
}

// 保存全部设置
async function saveAllSettings() {
  const interval = document.getElementById('setting-interval').value;
  const notifyLevel = document.getElementById('setting-notify-level').value;
  const autostart = document.getElementById('setting-autostart').checked;
  const port = parseInt(document.getElementById('setting-port').value) || 19530;
  const apiEnabled = document.getElementById('setting-api-enabled').checked;
  const apiKey = apiEnabled ? document.getElementById('setting-api-key').value.trim() : '';

  // 保存通知配置
  await apiFetch('/notifications/config', {
    method: 'PUT',
    body: JSON.stringify({ level: notifyLevel })
  });

  // 保存全局配置
  const res = await apiFetch('/config', {
    method: 'PUT',
    body: JSON.stringify({
      port: port,
      poll_interval: parseInt(interval),
      api_key: apiKey,
      auto_start: autostart,
    })
  });

  if (res.error) {
    setSettingsStatus('❌ 保存失败: ' + res.error, 'error');
  } else {
    setSettingsStatus('✅ 设置已保存（部分设置需重启生效）', 'success');
  }
}

function toggleWebService() {
  apiFetch('/config/toggle-web', { method: 'POST' }).then(res => {
    if (!res.error) {
      updateWebBtnText(res.enabled);
    }
  });
}

function updateWebBtnText(enabled) {
  const btn = document.getElementById('setting-toggle-web');
  if (btn) {
    btn.textContent = enabled ? '停止 Web 服务' : '启动 Web 服务';
    btn.style.background = enabled ? '#e74c3c' : '#27ae60';
  }
}

function openWebUI() {
  const port = document.getElementById('setting-port').value || 19530;
  window.open(`http://127.0.0.1:${port}`, '_blank');
}

// 保存网络插件端口和鉴权配置
async function savePluginPorts() {
  const plugins = ['http_in', 'tcp_udp_in', 'websocket_in'];
  const results = [];
  for (const pid of plugins) {
    const portEl = document.getElementById(`plugin-port-${pid}`);
    const authEl = document.getElementById(`plugin-auth-${pid}`);
    if (!portEl) continue;
    const port = parseInt(portEl.value) || 0;
    const authKey = authEl ? authEl.value.trim() : '';
    if (port > 0 || authKey) {
      const res = await apiFetch('/plugins/configure', {
        method: 'POST',
        body: JSON.stringify({ plugin_id: pid, params: { port, auth_key: authKey } })
      });
      results.push(`${pid}: ${res.error || '✅'}`);
    }
  }
  const statusEl = document.getElementById('plugin-port-status');
  if (statusEl) {
    statusEl.textContent = results.join('; ');
    statusEl.className = 'editor-status success';
    statusEl.style.display = 'inline';
    setTimeout(() => { statusEl.style.display = 'none'; }, 5000);
  }
}

// ============================================================================
// 任务管理辅助 — 手动触发
// ============================================================================

// 手动触发指定任务
async function manualTriggerTask(taskId) {
  const res = await apiFetch(`/tasks/${taskId}/trigger`, { method: 'POST' });
  const statusEl = document.getElementById('task-trigger-status');
  if (!statusEl) return;
  if (res.error) {
    statusEl.textContent = '❌ ' + res.error;
    statusEl.className = 'editor-status error';
  } else {
    statusEl.textContent = '✅ 已触发: ' + taskId + ' (查看仪表盘 manual_trigger 状态)';
    statusEl.className = 'editor-status success';
  }
  statusEl.style.display = 'inline';
  setTimeout(() => { statusEl.style.display = 'none'; }, 5000);
}

// ============================================================================
// 使用说明
// ============================================================================

async function loadManual() {
  const container = document.getElementById('manual-content');
  container.innerHTML = t('dashboard.loading');
  const lang = getLang();
  const langParam = lang === 'zh-CN' ? 'zh' : 'en';
  const res = await fetch('/api/manual?lang=' + langParam);
  if (!res.ok) {
    container.innerHTML = '<div class="error">加载失败</div>';
    return;
  }
  const md = await res.text();
  container.innerHTML = renderMarkdown(md);
  // 如果有 URL hash，滚动到对应锚点
  setTimeout(() => scrollToAnchor(), 100);
  // 拦截内容区的锚点点击，使用 JS 滚动
  container.querySelectorAll('a[href^="#"]').forEach(a => {
    a.addEventListener('click', (e) => {
      const href = a.getAttribute('href');
      if (!href || href === '#') return;
      const id = href.replace(/^#/, '');
      const el = document.getElementById(id);
      if (el) {
        e.preventDefault();
        el.scrollIntoView({ behavior: 'smooth', block: 'start' });
        // 更新 URL hash（不触发滚动）
        history.pushState(null, '', '#manual');
      }
    });
  });
}

// 根据 URL hash 滚动到锚点
function scrollToAnchor() {
  const hash = window.location.hash;
  if (!hash || hash === '#manual') return;
  const id = hash.replace(/^#manual\//, '').replace(/^#/, '');
  if (!id) return;
  const el = document.getElementById(id);
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

// 监听 hash 变化（点击 TOC 链接时触发）
window.addEventListener('hashchange', () => {
  const page = document.getElementById('page-manual');
  if (page && page.classList.contains('active')) {
    setTimeout(scrollToAnchor, 50);
  }
});

// 简易 Markdown → HTML 渲染器
function renderMarkdown(md) {
  let html = '';
  const lines = md.split('\n');
  let inCode = false;
  let codeContent = '';
  let inTable = false;
  let tableHeader = [];
  let tableAlign = [];

  for (let i = 0; i < lines.length; i++) {
    let line = lines[i];
    let trimmed = line.trim();

    // 代码块
    if (trimmed.startsWith('```')) {
      if (inCode) {
        html += '<pre><code>' + escHtml(codeContent.replace(/\n$/, '')) + '</code></pre>\n';
        codeContent = '';
        inCode = false;
      } else {
        if (inTable) { html += '</tbody></table>\n'; inTable = false; }
        inCode = true;
      }
      continue;
    }
    if (inCode) { codeContent += line + '\n'; continue; }

    // 空行 → 分隔块
    if (trimmed === '') {
      if (inTable) { html += '</tbody></table>\n'; inTable = false; }
      continue;
    }

    // 分隔线
    if (/^-{3,}$/.test(trimmed) && !inTable) {
      html += '<hr>\n';
      continue;
    }

    // 标题
    if (/^#{1,3}\s/.test(trimmed)) {
      const level = trimmed.match(/^#+/)[0].length;
      const rawText = trimmed.replace(/^#+\s+/, '');
      const text = renderInline(rawText);
      const id = headingToId(rawText);
      html += `<h${level} id="${id}">${text}</h${level}>\n`;
      continue;
    }

    // 表格
    if (trimmed.startsWith('|')) {
      // 跳过表格分隔行 (| --- | --- |)
      if (/^\|[\s\-:]+\|/.test(trimmed) && !inTable) continue;
      if (!inTable) {
        inTable = true;
        html += '<table><thead><tr>';
        const cells = trimmed.split('|').filter(c => c.trim() !== '');
        tableHeader = cells.map(c => renderInline(c.trim()));
        tableHeader.forEach(c => { html += `<th>${c}</th>`; });
        html += '</tr></thead><tbody>\n';
        // 下一行是分隔符，跳过
        if (i + 1 < lines.length && /^\|[\s\-:]+\|/.test(lines[i+1].trim())) {
          i++;
        }
      } else {
        // 表格数据行
        const cells = trimmed.split('|').filter(c => c.trim() !== '');
        html += '<tr>';
        cells.forEach(c => { html += `<td>${renderInline(c.trim())}</td>`; });
        html += '</tr>\n';
      }
      continue;
    }

    // 列表
    if (/^[\-\*]\s/.test(trimmed)) {
      html += '<ul><li>' + renderInline(trimmed.replace(/^[\-\*]\s+/, '')) + '</li></ul>\n';
      continue;
    }
    if (/^\d+\.\s/.test(trimmed)) {
      html += '<ol><li>' + renderInline(trimmed.replace(/^\d+\.\s+/, '')) + '</li></ol>\n';
      continue;
    }

    // 普通段落
    html += '<p>' + renderInline(trimmed) + '</p>\n';
  }

  if (inCode) html += '<pre><code>' + escHtml(codeContent.replace(/\n$/, '')) + '</code></pre>\n';
  if (inTable) html += '</tbody></table>\n';

  // 合并相邻的同类型列表
  html = html.replace(/<\/ul>\n<ul>/g, '');
  html = html.replace(/<\/ol>\n<ol>/g, '');
  return html;
}

// 生成标题 ID（匹配 GFM 风格锚点链接）
function headingToId(text) {
  let id = text.toLowerCase();
  // 移除 HTML 标签
  id = id.replace(/<[^>]+>/g, '');
  // 替换非单词字符（保留中文、字母、数字）为连字符
  id = id.replace(/[^\w\u4e00-\u9fff]+/g, '-');
  // 去掉首尾连字符
  id = id.replace(/^-+|-+$/g, '');
  return id || 'section';
}

// renderInline 处理行内格式：先转义 HTML，再应用 Markdown 格式
function renderInline(text) {
  // 1. 先转义 HTML 特殊字符
  let s = escHtml(text);
  // 2. 行内代码 `code` — 在转义后匹配
  s = s.replace(/`([^`]+)`/g, '<code>$1</code>');
  // 3. 加粗 **text** 或 __text__
  s = s.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  s = s.replace(/__(.+?)__/g, '<strong>$1</strong>');
  // 4. 链接 [text](url)
  s = s.replace(/\[(.+?)\]\((.+?)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>');
  // 5. 图片 ![alt](url)
  s = s.replace(/!\[(.+?)\]\((.+?)\)/g, '<img src="$2" alt="$1" style="max-width:100%;border-radius:4px">');
  return s;
}

// ============================================================================
// 连接状态心跳检测
// ============================================================================

let _heartbeatTimer = null;

function startConnectionHeartbeat() {
  if (_heartbeatTimer) clearInterval(_heartbeatTimer);
  checkConnection();
  _heartbeatTimer = setInterval(checkConnection, 5000);
}

async function checkConnection() {
  const badge = document.getElementById('connection-status');
  if (!badge) return;
  try {
    const res = await fetch('/api/status', { method: 'GET', signal: AbortSignal.timeout(3000) });
    if (res.ok) {
      badge.className = 'status-badge connected';
      badge.textContent = t('status.connected');
    } else {
      badge.className = 'status-badge disconnected';
      badge.textContent = t('status.disconnected');
    }
  } catch {
    badge.className = 'status-badge disconnected';
    badge.textContent = t('status.disconnected');
  }
}

