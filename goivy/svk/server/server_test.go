package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigFromEnvDevMode(t *testing.T) {
	t.Setenv("IVYSVK_DEV", "true")
	t.Setenv("IVYSVK_PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("IVYSVK_STATIC_DIR", "/tmp/svk")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.DevMode {
		t.Fatalf("expected dev mode")
	}
	if cfg.StaticDir != "/tmp/svk" {
		t.Fatalf("static dir = %q", cfg.StaticDir)
	}
}

func TestProductionConfigRequiresSecrets(t *testing.T) {
	t.Setenv("IVYSVK_DEV", "false")
	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected missing production config to fail")
	}
}

func TestPublicPagesRender(t *testing.T) {
	handler := newTestServer(t, Config{DevMode: true})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	if !strings.Contains(res.Body.String(), "SVK Ivy") {
		t.Fatalf("body did not render marketing page: %s", res.Body.String())
	}
	if res.Header().Get("x-frame-options") != "DENY" {
		t.Fatalf("missing security header")
	}
}

func TestAppRedirectsAnonymousUsersToLogin(t *testing.T) {
	handler := newTestServer(t, Config{DevMode: true})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusFound {
		t.Fatalf("status = %d", res.Code)
	}
	if loc := res.Header().Get("location"); loc != "/login" {
		t.Fatalf("location = %q", loc)
	}
}

func TestAuthenticatedAppServesStaticShell(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!doctype html><div id=\"app\">built app</div>"), 0o644); err != nil {
		t.Fatal(err)
	}
	handler := newTestServer(t, Config{DevMode: true, StaticDir: dir})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-1"})
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	if !strings.Contains(res.Body.String(), "built app") {
		t.Fatalf("body = %s", res.Body.String())
	}
}

func newTestServer(t *testing.T, cfg Config) http.Handler {
	t.Helper()
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
