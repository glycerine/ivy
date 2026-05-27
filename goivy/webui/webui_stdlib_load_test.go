//go:build web

package webui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	goivy "github.com/glycerine/ivy/goivy"
)

func TestWebUILoadOrdLiveUsesPreloadedStandardLibrary(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	specPath := filepath.Join(repoRoot, "ivy-lang-examples", "doc", "examples", "apple", "ord_live.ivy")
	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read ord_live.ivy: %v", err)
	}

	cfg := goivy.NewConfig()
	be := NewGoBackend(cfg)
	if be.stdlibErr != nil {
		t.Fatalf("preload standard library: %v", be.stdlibErr)
	}
	if cfg.StandardLibrary == nil {
		t.Fatalf("standard library was not attached to web backend config")
	}

	sessionBytes, err := be.NewSession(cfg)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	var sessionData map[string]string
	if err := json.Unmarshal(sessionBytes, &sessionData); err != nil {
		t.Fatalf("decode session response: %v", err)
	}
	sessionID := sessionData["session_id"]
	if sessionID == "" {
		t.Fatalf("session response missing session_id: %s", string(sessionBytes))
	}

	loadBytes, err := be.Load(sessionID, filepath.Base(specPath), content, "cf_live")
	if err != nil {
		t.Fatalf("load ord_live.ivy: %v", err)
	}
	var loadData map[string]interface{}
	if err := json.Unmarshal(loadBytes, &loadData); err != nil {
		t.Fatalf("decode load response: %v", err)
	}
	if got := loadData["isolate"]; got != "cf_live" {
		t.Fatalf("active isolate = %v, want cf_live", got)
	}
}
