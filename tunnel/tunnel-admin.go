// magpie-fn 隧道管理服务 (tunnel-admin)
//
// 为飞牛 fnOS 封装版 magpie 提供 Cloudflare Tunnel 管理：
//   - 快速隧道 (Quick Tunnel)：cloudflared tunnel --url，无需账号，一键生成临时 *.trycloudflare.com 公网 URL
//   - 命名隧道 (Named Tunnel)：cloudflared tunnel run --token，需 Cloudflare 账号 token 与自有域名，URL 持久
// 隧道方案参考 omniroute (github.com/diegosouzapw/OmniRoute) 的 Cloudflare Tunnel 做法，
// 仅学习其隧道形态，不引入其主体。
//
// 监听 :3431，提供管理页与 JSON API；状态持久化到 /config/tunnel.json。
// 仅使用 Go 标准库，独立 module，可在任意 Go 版本（>=1.22）编译。
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	addr      = "0.0.0.0:3431"
	cfgFile   = "/config/tunnel.json"
	quickLog  = "/config/tunnel-quick.log"
	namedLog  = "/config/tunnel-named.log"
	tokenFile = "/config/tunnel-named-token.txt"
)

// QuickState 快速隧道状态
type QuickState struct {
	Running bool   `json:"running"`
	URL     string `json:"url"`
	PID     int    `json:"pid,omitempty"`
	Target  string `json:"target,omitempty"`
}

// NamedState 命名隧道状态
type NamedState struct {
	Running  bool   `json:"running"`
	URL      string `json:"url"`
	PID      int    `json:"pid,omitempty"`
	Target   string `json:"target,omitempty"`
	HasToken bool   `json:"has_token"`
}

// State 整体状态
type State struct {
	Quick QuickState `json:"quick"`
	Named NamedState `json:"named"`
}

var (
	mu       sync.Mutex
	state    State
	quickCmd *exec.Cmd
	namedCmd *exec.Cmd
)

func main() {
	readState()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/quick/start", handleQuickStart)
	mux.HandleFunc("/api/quick/stop", handleQuickStop)
	mux.HandleFunc("/api/named/start", handleNamedStart)
	mux.HandleFunc("/api/named/stop", handleNamedStop)

	fmt.Printf("tunnel-admin listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "tunnel-admin: %v\n", err)
		os.Exit(1)
	}
}

// ---------- 状态持久化 ----------

func writeState() error {
	data, err := json.MarshalIndent(&state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgFile, data, 0o600)
}

func readState() {
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &state)
	if procAlive(state.Quick.PID) {
		state.Quick.Running = true
	} else {
		state.Quick.Running = false
	}
	if procAlive(state.Named.PID) {
		state.Named.Running = true
	} else {
		state.Named.Running = false
	}
	state.Named.HasToken = fileExists(tokenFile)
}

func procAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	return err == nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// ---------- 快速隧道 ----------

func startQuick(target string) error {
	mu.Lock()
	defer mu.Unlock()
	if state.Quick.Running {
		return fmt.Errorf("快速隧道已在运行中")
	}
	if target == "" {
		target = "3425"
	}
	if target != "3425" && target != "3430" {
		return fmt.Errorf("无效的目标端口: %s（仅支持 3425 网关 / 3430 Web）", target)
	}

	logF, err := os.OpenFile(quickLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	// 参考 omniroute：cloudflared 快速隧道无需账号，--url 指向本机服务
	cmd := exec.Command("cloudflared", "tunnel", "--no-autoupdate", "--url", "http://127.0.0.1:"+target, "--output", "json")
	cmd.Stdout = logF
	cmd.Stderr = logF
	if err := cmd.Start(); err != nil {
		logF.Close()
		return err
	}
	quickCmd = cmd
	state.Quick.Running = true
	state.Quick.PID = cmd.Process.Pid
	state.Quick.Target = target
	state.Quick.URL = ""
	if err := writeState(); err != nil {
		fmt.Fprintf(os.Stderr, "writeState: %v\n", err)
	}
	go watchQuick(cmd, logF)
	return nil
}

func watchQuick(cmd *exec.Cmd, logF *os.File) {
	// 轮询日志抓取 trycloudflare 公网 URL（最长 90 秒）
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		if url := scanQuickURL(quickLog); url != "" {
			mu.Lock()
			state.Quick.URL = url
			_ = writeState()
			mu.Unlock()
			break
		}
	}
	_ = cmd.Wait()
	_ = logF.Close()
	mu.Lock()
	if state.Quick.PID == cmd.Process.Pid {
		state.Quick.Running = false
		_ = writeState()
	}
	mu.Unlock()
}

func scanQuickURL(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	s := string(data)
	idx := strings.Index(s, "trycloudflare.com")
	if idx < 0 {
		return ""
	}
	start := strings.LastIndex(s[:idx], "https://")
	if start < 0 {
		return ""
	}
	return s[start : idx+len("trycloudflare.com")]
}

func stopQuick() error {
	mu.Lock()
	defer mu.Unlock()
	if quickCmd != nil && quickCmd.Process != nil {
		_ = quickCmd.Process.Kill()
	}
	state.Quick.Running = false
	state.Quick.URL = ""
	state.Quick.PID = 0
	return writeState()
}

// ---------- 命名隧道 ----------

func startNamed(token, target, domain string) error {
	mu.Lock()
	defer mu.Unlock()
	if state.Named.Running {
		return fmt.Errorf("命名隧道已在运行中")
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("缺少 Cloudflare Token")
	}
	if target == "" {
		target = "3425"
	}
	if target != "3425" && target != "3430" {
		return fmt.Errorf("无效的目标端口: %s（仅支持 3425 网关 / 3430 Web）", target)
	}

	if err := os.WriteFile(tokenFile, []byte(strings.TrimSpace(token)), 0o600); err != nil {
		return err
	}

	logF, err := os.OpenFile(namedLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	// 参考 omniroute：命名隧道使用 --token 运行，域名映射在 Cloudflare 侧配置
	cmd := exec.Command("cloudflared", "tunnel", "--no-autoupdate", "run", "--token", strings.TrimSpace(token))
	cmd.Stdout = logF
	cmd.Stderr = logF
	if err := cmd.Start(); err != nil {
		logF.Close()
		return err
	}
	namedCmd = cmd
	state.Named.Running = true
	state.Named.PID = cmd.Process.Pid
	state.Named.Target = target
	state.Named.URL = strings.TrimSpace(domain)
	state.Named.HasToken = true
	if err := writeState(); err != nil {
		fmt.Fprintf(os.Stderr, "writeState: %v\n", err)
	}
	go watchNamed(cmd, logF)
	return nil
}

func watchNamed(cmd *exec.Cmd, logF *os.File) {
	_ = cmd.Wait()
	_ = logF.Close()
	mu.Lock()
	if state.Named.PID == cmd.Process.Pid {
		state.Named.Running = false
		_ = writeState()
	}
	mu.Unlock()
}

func stopNamed() error {
	mu.Lock()
	defer mu.Unlock()
	if namedCmd != nil && namedCmd.Process != nil {
		_ = namedCmd.Process.Kill()
	}
	state.Named.Running = false
	state.Named.PID = 0
	return writeState()
}

// ---------- HTTP ----------

type reqBody struct {
	Token  string `json:"token"`
	Target string `json:"target"`
	Domain string `json:"domain"`
}

func decodeBody(r *http.Request) (reqBody, error) {
	var b reqBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil && err.Error() != "EOF" {
		return b, err
	}
	return b, nil
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, http.StatusOK, &state)
}

func handleQuickStart(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := startQuick(body.Target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, http.StatusOK, &state.Quick)
}

func handleQuickStop(w http.ResponseWriter, r *http.Request) {
	if err := stopQuick(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, http.StatusOK, &state.Quick)
}

func handleNamedStart(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := startNamed(body.Token, body.Target, body.Domain); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, http.StatusOK, &state.Named)
}

func handleNamedStop(w http.ResponseWriter, r *http.Request) {
	if err := stopNamed(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, http.StatusOK, &state.Named)
}

// ---------- 管理页 ----------

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = ioWriteString(w, pageHTML)
}

func ioWriteString(w http.ResponseWriter, s string) (int, error) {
	return w.Write([]byte(s))
}

const pageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>magpie-fn 云隧道管理</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif; background: #f5f7fa; color: #1f2328; padding: 24px; }
  .wrap { max-width: 760px; margin: 0 auto; }
  h1 { font-size: 20px; margin-bottom: 4px; }
  .sub { color: #666; font-size: 13px; margin-bottom: 20px; }
  .card { background: #fff; border: 1px solid #e2e5ea; border-radius: 10px; padding: 18px; margin-bottom: 16px; }
  .card h2 { font-size: 16px; margin-bottom: 10px; display: flex; align-items: center; gap: 8px; }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 10px; font-size: 12px; }
  .badge.on { background: #e6f7ec; color: #16a34a; }
  .badge.off { background: #f3f4f6; color: #6b7280; }
  .row { display: flex; gap: 8px; margin: 10px 0; flex-wrap: wrap; }
  label { font-size: 13px; color: #333; display: block; margin: 10px 0 4px; }
  select, input[type=text], input[type=password] { width: 100%; padding: 8px 10px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 14px; }
  button { padding: 8px 16px; border: none; border-radius: 6px; font-size: 14px; cursor: pointer; }
  .btn-primary { background: #2563eb; color: #fff; }
  .btn-danger { background: #ef4444; color: #fff; }
  .btn-ghost { background: #e5e7eb; color: #1f2328; }
  .urlbox { background: #f8fafc; border: 1px dashed #cbd5e1; border-radius: 6px; padding: 10px; font-size: 13px; word-break: break-all; display: none; }
  .urlbox.show { display: block; }
  .urlbox a { color: #2563eb; }
  .hint { font-size: 12px; color: #888; margin-top: 8px; line-height: 1.6; }
  .footer { color: #999; font-size: 12px; text-align: center; margin-top: 24px; line-height: 1.7; }
  .err { color: #dc2626; font-size: 13px; margin-top: 8px; display: none; }
</style>
</head>
<body>
<div class="wrap">
  <h1>magpie-fn 云隧道管理</h1>
  <div class="sub">为局域网外的 Agent / 设备提供公网访问。支持快速隧道（临时 URL，无需账号）与命名隧道（固定域名，需 Cloudflare Token）。</div>

  <div class="card">
    <h2>快速隧道 <span id="quickBadge" class="badge off">未运行</span></h2>
    <div>无需 Cloudflare 账号，一键开启即获得临时公网 URL（*.trycloudflare.com），进程停止后失效。</div>
    <label>暴露目标</label>
    <select id="quickTarget">
      <option value="3425">网关 3425（供外部 Agent 接入，推荐）</option>
      <option value="3430">Web 管理页 3430</option>
    </select>
    <div class="row">
      <button class="btn-primary" id="quickStart" onclick="quickStart()">开启快速隧道</button>
      <button class="btn-danger" id="quickStop" onclick="quickStop()" style="display:none">关闭</button>
      <button class="btn-ghost" id="quickCopy" onclick="copyUrl('quickUrl')" style="display:none">复制 URL</button>
    </div>
    <div class="urlbox" id="quickUrl"></div>
    <div class="err" id="quickErr"></div>
    <div class="hint">URL 在进程运行期间有效；重启容器后需重新开启。仅将 URL 分享给可信对象，暴露网关时建议先配置访问保护。</div>
  </div>

  <div class="card">
    <h2>命名隧道 <span id="namedBadge" class="badge off">未运行</span></h2>
    <div>使用 Cloudflare Zero Trust 的 Token 运行固定隧道，URL 持久稳定，适合长期暴露。参考 omniroute 的 Cloudflare Tunnel 方案。</div>
    <label>Cloudflare Token（首次填写后保存在本机，重启自动恢复）</label>
    <input type="password" id="namedToken" placeholder="粘贴 cloudflared tunnel --token 的 Token" autocomplete="off">
    <label>自有域名（可选，仅用于页面展示，实际映射请在 Cloudflare 控制台配置）</label>
    <input type="text" id="namedDomain" placeholder="https://your-domain.example.com">
    <label>暴露目标</label>
    <select id="namedTarget">
      <option value="3425">网关 3425（供外部 Agent 接入，推荐）</option>
      <option value="3430">Web 管理页 3430</option>
    </select>
    <div class="row">
      <button class="btn-primary" id="namedStart" onclick="namedStart()">开启命名隧道</button>
      <button class="btn-danger" id="namedStop" onclick="namedStop()" style="display:none">关闭</button>
      <button class="btn-ghost" id="namedCopy" onclick="copyUrl('namedUrl')" style="display:none">复制 URL</button>
    </div>
    <div class="urlbox" id="namedUrl"></div>
    <div class="err" id="namedErr"></div>
    <div class="hint">命名隧道的域名映射（Route）需在 Cloudflare 控制台 Zero Trust → Tunnels 中配置，本页仅负责在本机运行 cloudflared 进程。</div>
  </div>

  <div class="footer">
    magpie-fn（飞牛 fnOS 封装版）· 隧道方案参考 <a href="https://github.com/diegosouzapw/OmniRoute">omniroute</a> 的 Cloudflare Tunnel 实现<br>
    上游：<a href="https://github.com/yetone/magpie">yetone/magpie</a>（MIT）· 由 吴观风岳软件工作室 维护
  </div>
</div>
<script>
let quickRunning = false;
let namedRunning = false;

function api(path, body) {
  return fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : '{}'
  }).then(function(r) { return r.json(); });
}

function refresh() {
  fetch('/api/status').then(function(r) { return r.json(); }).then(function(s) {
    quickRunning = s.quick.running;
    namedRunning = s.named.running;
    var qb = document.getElementById('quickBadge');
    qb.className = 'badge ' + (s.quick.running ? 'on' : 'off');
    qb.textContent = s.quick.running ? '运行中' : '未运行';
    document.getElementById('quickStart').style.display = s.quick.running ? 'none' : '';
    document.getElementById('quickStop').style.display = s.quick.running ? '' : 'none';
    document.getElementById('quickCopy').style.display = (s.quick.running && s.quick.url) ? '' : 'none';
    var qUrl = document.getElementById('quickUrl');
    if (s.quick.url) { qUrl.innerHTML = '公网 URL：<a href="' + s.quick.url + '" target="_blank">' + s.quick.url + '</a>'; qUrl.className = 'urlbox show'; }
    else if (s.quick.running) { qUrl.textContent = '正在建立连接，请稍候…'; qUrl.className = 'urlbox show'; }
    else { qUrl.className = 'urlbox'; }

    var nb = document.getElementById('namedBadge');
    nb.className = 'badge ' + (s.named.running ? 'on' : 'off');
    nb.textContent = s.named.running ? '运行中' : '未运行';
    document.getElementById('namedStart').style.display = s.named.running ? 'none' : '';
    document.getElementById('namedStop').style.display = s.named.running ? '' : 'none';
    document.getElementById('namedCopy').style.display = (s.named.running && s.named.url) ? '' : 'none';
    var nUrl = document.getElementById('namedUrl');
    if (s.named.running && s.named.url) { nUrl.innerHTML = '访问地址：<a href="' + s.named.url + '" target="_blank">' + s.named.url + '</a>'; nUrl.className = 'urlbox show'; }
    else if (s.named.running) { nUrl.textContent = '正在运行，请在 Cloudflare 控制台确认域名映射…'; nUrl.className = 'urlbox show'; }
    else { nUrl.className = 'urlbox'; }
    if (s.named.has_token && !namedRunning) {
      var t = document.getElementById('namedToken');
      if (!t.value) t.value = '********';
    }
  }).catch(function(){});
}

function quickStart() {
  api('/api/quick/start', { target: document.getElementById('quickTarget').value }).then(function(r) {
    if (r.error) { showErr('quickErr', r.error); } else { refresh(); }
  });
}
function quickStop() {
  api('/api/quick/stop', {}).then(function(r) { refresh(); });
}
function namedStart() {
  api('/api/named/start', {
    token: document.getElementById('namedToken').value,
    domain: document.getElementById('namedDomain').value,
    target: document.getElementById('namedTarget').value
  }).then(function(r) {
    if (r.error) { showErr('namedErr', r.error); } else { refresh(); }
  });
}
function namedStop() {
  api('/api/named/stop', {}).then(function(r) { refresh(); });
}
function copyUrl(id) {
  var el = document.getElementById(id);
  var t = el.textContent.replace('公网 URL：', '').replace('访问地址：', '').trim();
  var a = document.createElement('textarea');
  a.value = t; document.body.appendChild(a); a.select();
  try { document.execCommand('copy'); } catch (e) {}
  document.body.removeChild(a);
}
function showErr(id, msg) {
  var el = document.getElementById(id);
  el.textContent = msg; el.style.display = 'block';
  setTimeout(function(){ el.style.display = 'none'; }, 4000);
}
setInterval(refresh, 3000);
refresh();
</script>
</body>
</html>
`