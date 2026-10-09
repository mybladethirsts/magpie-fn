package gui

import (
	"encoding/json"
	"net/http"

	"github.com/yetone/magpie/internal/tunnel"
)

// tunnelRoutes serve the Cloud tunnel block on Settings → Network and
// sharing: start, stop and configure a quick or named tunnel for the
// gateway or the web page.
func tunnelRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/tunnel", func(rw http.ResponseWriter, r *http.Request) {
		writeJSON(rw, tunnelView())
	})
	mux.HandleFunc("POST /api/tunnel/start", func(rw http.ResponseWriter, r *http.Request) {
		var in tunnelConfigIn
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			fail(rw, err)
			return
		}
		st, err := tunnel.Start(tunnel.Config{Mode: in.Mode, Target: in.Target, Token: in.Token})
		if err != nil {
			fail(rw, err)
			return
		}
		writeJSON(rw, map[string]any{"state": st, "config": tunnelConfigView()})
	})
	mux.HandleFunc("POST /api/tunnel/stop", func(rw http.ResponseWriter, r *http.Request) {
		st, err := tunnel.Stop()
		if err != nil {
			fail(rw, err)
			return
		}
		writeJSON(rw, map[string]any{"state": st, "config": tunnelConfigView()})
	})
	// save the choices without touching the running tunnel
	mux.HandleFunc("POST /api/tunnel/config", func(rw http.ResponseWriter, r *http.Request) {
		var in tunnelConfigIn
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			fail(rw, err)
			return
		}
		if err := tunnel.SaveConfig(tunnel.Config{Mode: in.Mode, Target: in.Target, Token: in.Token}); err != nil {
			fail(rw, err)
			return
		}
		writeJSON(rw, map[string]any{"state": tunnel.Status(), "config": tunnelConfigView()})
	})
}

// tunnelConfigIn is what the page may send. Token is kept only when the
// page sends one; an empty token leaves the saved one alone (Start merges
// the saved token before saving, and the page never sees the token).
type tunnelConfigIn struct {
	Mode   string `json:"mode"`
	Target string `json:"target"`
	Token  string `json:"token"`
}

func tunnelView() map[string]any {
	if _, err := tunnel.LoadConfig(); err != nil {
		return map[string]any{"state": tunnel.Status(), "config": map[string]any{}, "configError": err.Error()}
	}
	return map[string]any{"state": tunnel.Status(), "config": tunnelConfigView()}
}

// tunnelConfigView is the saved choices with the token masked: the page
// never sees the token itself again.
func tunnelConfigView() map[string]any {
	cfg, err := tunnel.LoadConfig()
	if err != nil {
		return map[string]any{}
	}
	return map[string]any{
		"mode":     cfg.Mode,
		"target":   cfg.Target,
		"tokenSet": cfg.Token != "",
	}
}
