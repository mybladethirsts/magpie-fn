// Package tunnel runs a Cloudflare tunnel for magpie, so machines outside
// the local network can reach the gateway or the web page.
//
// Two modes, both run by the cloudflared binary shipped in the container
// image (fnOS / Docker builds; on the desktop it is looked up on PATH or
// at $MAGPIE_CLOUDFLARED):
//
//   - quick: "cloudflared tunnel --url …" gives a temporary
//     trycloudflare.com URL with zero configuration. The URL is alive only
//     while the process runs and changes on every restart.
//
//   - named: "cloudflared tunnel run --token …" runs a Named Tunnel created
//     in the Cloudflare dashboard (Zero Trust → Networks → Tunnels), whose
//     public hostname the user chose there. The URL stays the same across
//     restarts and needs a Cloudflare account and a domain.
//
// The choices are kept in tunnel.json beside magpie's settings directory
// (XDG_CONFIG_HOME). The token is stored there too and never sent back to
// the page once saved.
package tunnel

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Config is the tunnel's saved choices.
type Config struct {
	Mode   string `json:"mode"`            // "quick" or "named"
	Target string `json:"target"`          // "gateway" (port 3425) or "web" (port 3430)
	Token  string `json:"token,omitempty"` // named mode: Cloudflare tunnel token
}

// State is the tunnel's live status, as the page shows it.
type State struct {
	Running bool   `json:"running"`
	Mode    string `json:"mode"`
	Target  string `json:"target"`
	URL     string `json:"url,omitempty"` // the public URL when known (quick mode)
	Error   string `json:"error,omitempty"`
	Pid     int    `json:"pid,omitempty"`
	Since   string `json:"since,omitempty"`
}

// Ports the tunnel exposes. The gateway listens on MAGPIE_ADDR
// (0.0.0.0:3425 in the container) and the web page on 3430; cloudflared
// connects to 127.0.0.1, which reaches both in the container and on a
// desktop.
var targetPorts = map[string]string{"gateway": "3425", "web": "3430"}

// DefaultConfig is what a fresh install starts with: a quick tunnel to the
// gateway.
func DefaultConfig() Config { return Config{Mode: "quick", Target: "gateway"} }

var (
	mu      sync.Mutex
	proc    *exec.Cmd
	state   State
	cfgPath string
)

// configFile is tunnel.json beside magpie's settings directory.
func configFile() string {
	if cfgPath != "" {
		return cfgPath
	}
	d := os.Getenv("XDG_CONFIG_HOME")
	if d == "" {
		d = "."
	}
	return filepath.Join(d, "tunnel.json")
}

// SetConfigFile overrides where the config is kept (tests, or a desktop
// install that keeps settings elsewhere).
func SetConfigFile(p string) { cfgPath = p }

// LoadConfig reads the saved choices; a missing file is a fresh install.
func LoadConfig() (Config, error) {
	mu.Lock()
	defer mu.Unlock()
	return loadLocked()
}

func loadLocked() (Config, error) {
	b, err := os.ReadFile(configFile())
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}
	if c.Mode != "named" {
		c.Mode = "quick"
	}
	if c.Target != "web" {
		c.Target = "gateway"
	}
	return c, nil
}

// SaveConfig keeps the user's choices without touching the running tunnel.
func SaveConfig(c Config) error {
	mu.Lock()
	defer mu.Unlock()
	return save(c)
}

func save(c Config) error {
	if c.Mode != "named" {
		c.Mode = "quick"
	}
	if c.Target != "web" {
		c.Target = "gateway"
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configFile()), 0o755); err != nil {
		return err
	}
	return os.WriteFile(configFile(), b, 0o600)
}

// Cloudflared finds the binary: $MAGPIE_CLOUDFLARED, the image's
// /usr/local/bin/cloudflared, or PATH.
func Cloudflared() (string, error) {
	if p := os.Getenv("MAGPIE_CLOUDFLARED"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	for _, p := range []string{"/usr/local/bin/cloudflared", "/usr/bin/cloudflared", "/bin/cloudflared"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("cloudflared"); err == nil {
		return p, nil
	}
	return "", errors.New("cloudflared not found: the fnOS build ships it at /usr/local/bin/cloudflared")
}

// Start brings the tunnel up with the given choices (also saved). A tunnel
// already running with the same choices is left alone; different choices
// restart it. An empty token keeps the one already saved.
func Start(cfg Config) (State, error) {
	mu.Lock()
	defer mu.Unlock()
	if strings.TrimSpace(cfg.Token) == "" {
		if old, err := loadLocked(); err == nil {
			cfg.Token = old.Token
		}
	}
	if cfg.Mode == "named" && strings.TrimSpace(cfg.Token) == "" {
		return state, errors.New("a named tunnel needs the token from your Cloudflare dashboard")
	}
	if err := save(cfg); err != nil {
		return state, err
	}
	if runningLocked() {
		if state.Mode == cfg.Mode && state.Target == cfg.Target {
			return state, nil
		}
		stopLocked()
	}
	bin, err := Cloudflared()
	if err != nil {
		state = State{Mode: cfg.Mode, Target: cfg.Target, Error: err.Error()}
		return state, err
	}
	port, ok := targetPorts[cfg.Target]
	if !ok {
		return state, fmt.Errorf("unknown tunnel target %q", cfg.Target)
	}
	var args []string
	switch cfg.Mode {
	case "quick":
		args = []string{"tunnel", "--url", "http://127.0.0.1:" + port, "--no-autoupdate", "--no-changelog"}
	case "named":
		args = []string{"tunnel", "run", "--token", strings.TrimSpace(cfg.Token), "--no-autoupdate"}
	default:
		return state, fmt.Errorf("unknown tunnel mode %q", cfg.Mode)
	}
	cmd := exec.Command(bin, args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return state, err
	}
	if err := cmd.Start(); err != nil {
		state = State{Mode: cfg.Mode, Target: cfg.Target, Error: "cloudflared failed to start: " + err.Error()}
		return state, err
	}
	proc = cmd
	state = State{Running: true, Mode: cfg.Mode, Target: cfg.Target, Pid: cmd.Process.Pid, Since: time.Now().Format(time.RFC3339)}
	go tail(cmd, stderr)
	return state, nil
}

// Stop ends the tunnel; the quick URL dies with the process.
func Stop() (State, error) {
	mu.Lock()
	defer mu.Unlock()
	stopLocked()
	return state, nil
}

func stopLocked() {
	if proc != nil && proc.Process != nil && proc.ProcessState == nil {
		_ = proc.Process.Kill()
		_, _ = proc.Process.Wait() // reap
	}
	proc = nil
	state = State{}
}

// Status is the current state. A process that died without tail seeing it
// yet is reported stopped.
func Status() State {
	mu.Lock()
	defer mu.Unlock()
	if proc != nil && proc.Process != nil && proc.ProcessState != nil {
		state = State{}
		proc = nil
	}
	return state
}

func runningLocked() bool {
	return proc != nil && proc.Process != nil && proc.ProcessState == nil
}

// tail reads cloudflared's stderr: the quick URL is announced there, and
// the end of the stream is the end of the tunnel.
func tail(cmd *exec.Cmd, r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		logLine(line)
		if m := quickURL.FindString(line); m != "" {
			mu.Lock()
			if runningLocked() {
				state.URL = m
			}
			mu.Unlock()
		}
	}
	mu.Lock()
	if proc == cmd || (proc == nil && state.Running) {
		state = State{}
		proc = nil
	}
	mu.Unlock()
}

// quickURL is a trycloudflare.com hostname as cloudflared prints it.
var quickURL = regexp.MustCompile(`https://[A-Za-z0-9-]+\.trycloudflare\.com`)

// logFile is cloudflared's log, next to tunnel.json, for diagnosis.
func logFile() string { return filepath.Join(filepath.Dir(configFile()), "tunnel.log") }

func logLine(line string) {
	f, err := os.OpenFile(logFile(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s\n", time.Now().Format(time.RFC3339), line)
}
