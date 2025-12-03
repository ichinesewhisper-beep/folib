package handles

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SearchPage(c *gin.Context) {
	const tpl = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8" />
<title>高级搜索</title>
<meta name="viewport" content="width=device-width, initial-scale=1" />
<style>
:root { --accent: #1890ff; --bg: #f7f8fa; --card: #fff; --text: #1f2937; --muted: #6b7280; --border: #e5e7eb; }
body { font-family: system-ui,-apple-system,Segoe UI,Roboto; margin: 0; background: var(--bg); color: var(--text); }
.page { min-height: 100vh; display: flex; flex-direction: column; }
.header { position: sticky; top: 0; z-index: 20; background: var(--bg); border-bottom: 1px solid var(--border); }
.header-inner { max-width: 1080px; margin: 0 auto; padding: 12px 16px; display: flex; align-items: center; justify-content: space-between; }
.brand { display:flex; align-items:center; gap:10px; color: var(--text); font-weight: 600; }
.brand img { height: 24px; width: auto; }
.brand a { color: inherit; text-decoration: none; }
.container { max-width: 1080px; margin: 20px auto; padding: 16px; }
.card { background: var(--card); border: 1px solid var(--border); border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,.04); }
.content { padding: 16px; }
.row { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; margin-bottom: 12px; }
input[type=text], input[type=number], select { padding: 10px 12px; border: 1px solid var(--border); border-radius: 10px; background: var(--card); color: var(--text); }
input[type=text]:focus, input[type=number]:focus, select:focus { outline: none; border-color: var(--accent); box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent) 25%, transparent); }
button { padding: 10px 16px; background: color-mix(in srgb, var(--accent) 15%, var(--card)); color: var(--accent); border: 1px solid color-mix(in srgb, var(--accent) 30%, transparent); border-radius: 10px; cursor: pointer; }
button[disabled] { opacity: .6; cursor: not-allowed; }
button:hover { background: color-mix(in srgb, var(--accent) 20%, var(--card)); }
label { display: inline-flex; align-items: center; gap: 6px; }
table { border-collapse: collapse; width: 100%; margin-top: 12px; border: 1px solid var(--border); border-radius: 12px; overflow: hidden; }
thead { background: #f9fafb; }
th, td { border-bottom: 1px solid var(--border); padding: 10px; text-align: left; }
tbody tr:hover { background: #fafafa; }
.badge { display:inline-block; padding:2px 8px; border-radius:999px; background:#eef2ff; color:#3730a3; font-size:12px; border: 1px solid #e0e7ff; }
.msg { padding: 8px 12px; border-radius: 10px; }
.msg.info { color: var(--text); background: #f9fafb; border: 1px solid var(--border); }
.msg.error { color: #b00020; background: #ffecec; border: 1px solid #ffb4b4; }
.pager { display:flex; align-items:center; gap:8px; margin-top: 12px; flex-wrap: wrap; }
.page-btn { padding:6px 10px; border:1px solid var(--border); border-radius:10px; cursor:pointer; background: var(--card); color: var(--text); }
.page-btn.active { background: color-mix(in srgb, var(--accent) 15%, var(--card)); color: var(--accent); border-color: color-mix(in srgb, var(--accent) 30%, transparent); }
.overlay { position: fixed; inset: 0; display: none; align-items: center; justify-content: center; background: rgba(247,248,250,.7); z-index: 50; }
.spinner { width: 40px; height: 40px; border: 4px solid var(--accent); border-top-color: transparent; border-radius: 50%; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (prefers-color-scheme: dark) {
  :root { --bg: rgb(15,15,15); --card: #121212; --text: #e5e7eb; --muted: #9ca3af; --border: #2d2d2d; }
  thead { background: #0b0b0b; }
  tbody tr:hover { background: #0f0f0f; }
}
</style>
</head>
<body>
<div class="page">
  <div class="header">
    <div class="header-inner">
      <div class="brand"><span>OpenList</span></div>
      
    </div>
  </div>
  <div class="container">
    <div class="card">
      <div class="content" role="search" aria-label="高级搜索">
        <div class="row">
          <input id="keywords" type="text" placeholder="请输入关键词" style="flex:1;" aria-label="关键词" />
          <button id="doSearch" aria-label="执行搜索">搜索</button>
        </div>
        <div class="row">
          <label><input type="checkbox" id="selAll" /> 全选</label>
          <label><input type="checkbox" id="isDir" /> 文件夹</label>
          <label><input type="checkbox" class="type" data-group="text" checked /> 文档</label>
          <label><input type="checkbox" class="type" data-group="image" checked /> 图片</label>
          <label><input type="checkbox" class="type" data-group="audio" checked /> 音频</label>
          <label><input type="checkbox" class="type" data-group="video" checked /> 视频</label>
          <label><input type="checkbox" class="type" data-group="archive" /> 压缩包</label>
          <label><input type="checkbox" class="type" data-group="ebook" /> 电子书</label>
          <label><input type="checkbox" class="type" data-group="code" /> 代码/配置</label>
          <label style="margin-left:16px;"><span>指定类型</span>
            <input id="extInput" type="text" placeholder="如: pdf,zip,mp4" style="min-width:240px;" aria-label="扩展类型" />
          </label>
          <label style="margin-left:auto;"><input type="checkbox" id="distinct" /> 排除重复文件</label>
        </div>
        <div class="row">
          <label>父目录: <input id="parent" type="text" value="/" style="min-width:180px;" aria-label="父目录" /></label>
          <label>大小范围: 最小 <input id="sizeMin" type="number" min="0" value="0" style="width:120px;" aria-label="最小大小" /> B</label>
          <label>最大 <input id="sizeMax" type="number" min="0" value="0" style="width:120px;" aria-label="最大大小" /> B</label>
          <label>排序:
            <select id="orderBy" aria-label="排序字段">
              <option value="">默认</option>
              <option value="name">名称</option>
              <option value="size">大小</option>
            </select>
            <select id="orderDirection" aria-label="排序方向">
              <option value="asc">升序</option>
              <option value="desc">降序</option>
            </select>
          </label>
          <label>每页 <input id="perPage" type="number" min="1" value="50" style="width:80px;" aria-label="每页数量" /></label>
          <label>页码 <input id="page" type="number" min="1" value="1" style="width:80px;" aria-label="页码" /></label>
        </div>
        <div id="msg" class="row msg info" role="status" aria-live="polite"></div>
        <table>
          <thead>
            <tr><th>名称</th><th>路径</th><th>类型</th><th>大小</th><th>操作</th></tr>
          </thead>
          <tbody id="tbody"></tbody>
        </table>
        <div class="pager" id="pager"></div>
      </div>
    </div>
  </div>
</div>
<div class="overlay" id="overlay" aria-hidden="true"><div class="spinner" aria-label="加载中"></div></div>


<script>
const BASE = (location.pathname.split('/search')[0] || '');
const API_PREFIX = BASE + "/api";
const state = { page: 1, perPage: 50, total: 0, loading: false };
const LS_KEY = 'openlist_search_state' + BASE;
function token() { return localStorage.getItem('token') || ''; }
function setAccentFromConfig(){
  try{
    const color = (window.OPENLIST_CONFIG && window.OPENLIST_CONFIG.main_color) || '#1890ff';
    document.documentElement.style.setProperty('--accent', color);
  }catch(e){}
}
setAccentFromConfig();
function groupExts(group) {
  const map = {
    text: ["txt","md","json","xml","ini","conf","yml","yaml","log","csv","doc","docx","ppt","pptx","xls","xlsx","pdf"],
    image: ["jpg","jpeg","png","gif","bmp","svg","ico","webp","avif","tiff","nef","cr2","arw","dng","heic","heif"],
    audio: ["mp3","flac","ogg","m4a","wav","opus","wma"],
    video: ["mp4","mkv","avi","mov","rmvb","webm","flv","m3u8","ts","m2ts","mpeg"],
    archive: ["zip","rar","7z","gz","bz2","xz","tar","zst"],
    ebook: ["pdf","epub","mobi","azw","azw3","djvu"],
    code: ["c","cpp","h","hpp","rs","go","js","ts","jsx","tsx","py","java","kt","rb","php","cs","sh","bat","ps1","sql","json","yaml","yml","toml","ini","conf"]
  };
  return map[group] || [];
}
function formatSize(bytes) {
  const u = ['B','KB','MB','GB','TB'];
  let i = 0, n = Number(bytes||0);
  while (n >= 1024 && i < u.length - 1) { n/=1024; i++; }
  return (i===0? n : n.toFixed(2)) + ' ' + u[i];
}
function resolveExtInput(extInput){
  const macros = {
    'archive': ['zip','rar','7z','gz','bz2','xz','tar','zst'],
    'ebook': ['pdf','epub','mobi','azw','azw3','djvu'],
    'doc': ['doc','docx','odt','rtf','pages'],
    'sheet': ['xls','xlsx','csv','tsv','ods','numbers'],
    'ppt': ['ppt','pptx','odp','key'],
    'code': ['c','cpp','h','hpp','rs','go','js','ts','jsx','tsx','py','java','kt','rb','php','cs','sh','bat','ps1','sql','json','yaml','yml','toml','ini','conf'],
    'image': ["jpg","jpeg","png","gif","bmp","svg","ico","webp","avif","tiff","nef","cr2","arw","dng","heic","heif"],
    'video': ["mp4","mkv","avi","mov","rmvb","webm","flv","m3u8","ts","m2ts","mpeg"]
  };
  const out = [];
  extInput.split(',').map(s=>s.trim().toLowerCase()).filter(Boolean).forEach(t=>{
    if (macros[t]) { out.push(...macros[t]); } else { out.push(t); }
  });
  return Array.from(new Set(out));
}

document.getElementById('selAll').addEventListener('change', e => {
  const checked = e.target.checked;
  document.querySelectorAll('input.type').forEach(el => el.checked = checked);
});

function setLoading(b){
  state.loading = b;
  document.getElementById('overlay').style.display = b? 'flex':'none';
  document.getElementById('doSearch').disabled = b;
  document.querySelector('.content')?.setAttribute('aria-busy', b? 'true':'false');
}
function renderPagination(){
  const pager = document.getElementById('pager');
  const pages = Math.max(1, Math.ceil(state.total / state.perPage));
  const cur = Math.min(Math.max(1, state.page), pages);
  state.page = cur;
  pager.innerHTML = '';
  const prev = document.createElement('button'); prev.className='page-btn'; prev.textContent='上一页'; prev.disabled = cur<=1; prev.onclick=()=>{ document.getElementById('page').value = cur-1; doSearch(); };
  const next = document.createElement('button'); next.className='page-btn'; next.textContent='下一页'; next.disabled = cur>=pages; next.onclick=()=>{ document.getElementById('page').value = cur+1; doSearch(); };
  const info = document.createElement('span'); info.textContent = '第 ' + cur + ' / ' + pages + ' 页，共 ' + state.total + ' 项';
  pager.appendChild(prev);
  const start = Math.max(1, cur-3), end = Math.min(pages, cur+3);
  for(let p=start; p<=end; p++){
    const b=document.createElement('button'); b.className='page-btn' + (p===cur?' active':''); b.textContent=String(p); b.onclick=()=>{ document.getElementById('page').value = p; doSearch(); };
    pager.appendChild(b);
  }
  pager.appendChild(next);
  pager.appendChild(info);
}
async function doSearch(){
  const keywords = document.getElementById('keywords').value.trim();
  const parent = document.getElementById('parent').value.trim() || '/';
  const isDir = document.getElementById('isDir').checked;
  const perPage = Number(document.getElementById('perPage').value || 50);
  const page = Number(document.getElementById('page').value || 1);
  const sizeMin = Number(document.getElementById('sizeMin').value || 0);
  const sizeMax = Number(document.getElementById('sizeMax').value || 0);
  const orderBy = document.getElementById('orderBy').value;
  const orderDirection = document.getElementById('orderDirection').value;
  const distinct = document.getElementById('distinct').checked;

  let exts = [];
  document.querySelectorAll('input.type:checked').forEach(el => {
    exts = exts.concat(groupExts(el.getAttribute('data-group')));
  });
  const extInput = document.getElementById('extInput').value.trim();
  if (extInput) { resolveExtInput(extInput).forEach(e => exts.push(e)); }
  exts = Array.from(new Set(exts));

  const scope = isDir ? 1 : (extInput ? 2 : 0);

  const body = {
    parent, keywords, scope, exts, size_min: sizeMin, size_max: sizeMax,
    order_by: orderBy, order_direction: orderDirection, distinct,
    page, per_page: perPage
  };
  updateUrlFromUI();
  saveState();

  const msg = document.getElementById('msg');
  msg.className = 'row msg info';
  msg.textContent = '搜索中...';
  setLoading(true);
  try {
    const res = await fetch(API_PREFIX + '/fs/search', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token() ? ('Bearer ' + token()) : ''
      },
      body: JSON.stringify(body)
    });
    if (!res.ok) {
      let errText = String(res.status);
      try{ const ejson = await res.json(); errText = ejson.message || ejson.error || JSON.stringify(ejson); }catch(e){ try{ errText = await res.text(); }catch(_){} }
      msg.className = 'row msg error';
      msg.textContent = '搜索失败: ' + errText;
      setLoading(false);
      return;
    }
    const data = await res.json();
    const list = (data.data && data.data.content) || [];
    const total = (data.data && data.data.total) || 0;
    state.total = Number(total)||0; state.perPage = perPage; state.page = page;
    msg.className = 'row msg info';
    msg.textContent = '共 ' + state.total + ' 项';
    const tbody = document.getElementById('tbody');
    tbody.innerHTML = '';
    list.forEach(it => {
      const tr = document.createElement('tr');
      const typeText = it.is_dir? '文件夹' : (function(){
        const t = it.type;
        return t===2?'视频':(t===3?'音频':(t===4?'文本':(t===5?'图片':'未知')));
      })();
      tr.innerHTML = '<td>' + it.name + '</td>' +
                     '<td>' + it.parent + '</td>' +
                     '<td><span class="badge">' + typeText + '</span></td>' +
                     '<td>' + (it.is_dir ? '-' : formatSize(it.size)) + '</td>' +
                      '<td><a class="page-btn" href="' + BASE + '/?path=' + encodeURIComponent(it.parent) + '" target="_blank" rel="noopener noreferrer">打开所在目录</a></td>';
      tbody.appendChild(tr);
    });
    if(list.length === 0){
      const tr = document.createElement('tr'); tr.innerHTML = '<td colspan="5" style="color:var(--muted); text-align:center;">无结果</td>'; tbody.appendChild(tr);
    }
    renderPagination();
  } catch (e) {
    msg.className = 'row msg error';
    msg.textContent = '搜索异常: ' + e;
  }
  setLoading(false);
}
document.getElementById('doSearch').addEventListener('click', doSearch);
document.getElementById('keywords').addEventListener('keydown', function(e){ if(e.key==='Enter'){ doSearch(); }});

function getUIState(){
  const groups = Array.from(document.querySelectorAll('input.type:checked')).map(el=>el.getAttribute('data-group'));
  return {
    keywords: document.getElementById('keywords').value,
    parent: document.getElementById('parent').value,
    isDir: document.getElementById('isDir').checked,
    perPage: Number(document.getElementById('perPage').value||50),
    page: Number(document.getElementById('page').value||1),
    sizeMin: Number(document.getElementById('sizeMin').value||0),
    sizeMax: Number(document.getElementById('sizeMax').value||0),
    orderBy: document.getElementById('orderBy').value,
    orderDirection: document.getElementById('orderDirection').value,
    distinct: document.getElementById('distinct').checked,
    extInput: document.getElementById('extInput').value,
    selAll: document.getElementById('selAll').checked,
    groups
  };
}
function applyUIState(s){
  if(!s) return;
  document.getElementById('keywords').value = s.keywords||'';
  document.getElementById('parent').value = s.parent||'/';
  document.getElementById('isDir').checked = !!s.isDir;
  document.getElementById('perPage').value = s.perPage||50;
  document.getElementById('page').value = s.page||1;
  document.getElementById('sizeMin').value = s.sizeMin||0;
  document.getElementById('sizeMax').value = s.sizeMax||0;
  document.getElementById('orderBy').value = s.orderBy||'';
  document.getElementById('orderDirection').value = s.orderDirection||'asc';
  document.getElementById('distinct').checked = !!s.distinct;
  document.getElementById('extInput').value = s.extInput||'';
  document.getElementById('selAll').checked = !!s.selAll;
  const setGroups = new Set(s.groups||[]);
  document.querySelectorAll('input.type').forEach(el=>{ el.checked = setGroups.has(el.getAttribute('data-group')); });
}
function saveState(){ try{ localStorage.setItem(LS_KEY, JSON.stringify(getUIState())); }catch(e){} }
function loadState(){ try{ const s = JSON.parse(localStorage.getItem(LS_KEY)||'null'); applyUIState(s); }catch(e){} }
function buildQuery(s){
  const p = new URLSearchParams();
  p.set('k', s.keywords||'');
  p.set('p', s.parent||'/');
  p.set('dir', s.isDir? '1':'0');
  p.set('pp', String(s.perPage||50));
  p.set('page', String(s.page||1));
  p.set('smin', String(s.sizeMin||0));
  p.set('smax', String(s.sizeMax||0));
  p.set('ob', s.orderBy||'');
  p.set('od', s.orderDirection||'asc');
  p.set('d', s.distinct? '1':'0');
  p.set('ext', s.extInput||'');
  p.set('groups', (s.groups||[]).join(','));
  return p.toString();
}
function updateUrlFromUI(){
  const s = getUIState();
  const q = buildQuery(s);
  history.replaceState(null, '', location.pathname + '?' + q + location.hash);
}
function applyQuery(){
  const q = new URLSearchParams(location.search);
  if (!q.toString()) return;
  const s = {
    keywords: q.get('k')||'',
    parent: q.get('p')||'/',
    isDir: (q.get('dir')||'0')==='1',
    perPage: Number(q.get('pp')||'50'),
    page: Number(q.get('page')||'1'),
    sizeMin: Number(q.get('smin')||'0'),
    sizeMax: Number(q.get('smax')||'0'),
    orderBy: q.get('ob')||'',
    orderDirection: q.get('od')||'asc',
    distinct: (q.get('d')||'0')==='1',
    extInput: q.get('ext')||'',
    selAll: false,
    groups: (q.get('groups')||'').split(',').filter(Boolean)
  };
  applyUIState(s);
}
document.addEventListener('DOMContentLoaded', function(){
  applyQuery();
  if(!location.search) loadState();
  document.querySelectorAll('input,select').forEach(el=>{
    el.addEventListener('change', function(){ saveState(); updateUrlFromUI(); });
    if(el.id==='keywords' || el.id==='extInput'){ el.addEventListener('input', function(){ saveState(); updateUrlFromUI(); }); }
  });
  doSearch();
});

</script>
</body></html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(tpl))
}
