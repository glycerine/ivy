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

func TestLoginRendersSignupRecoveryAndDevEntry(t *testing.T) {
	handler := newTestServer(t, Config{DevMode: true})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	body := res.Body.String()
	for _, want := range []string{
		"Create account",
		"Continue with Google",
		"Continue with Apple",
		"Continue with GitHub",
		"Send magic link",
		"Create OPAQUE password account",
		"Continue as local dev user",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("login page missing %q: %s", want, body)
		}
	}
}

func TestAppRedirectsAnonymousUsersToLoginInProductionMode(t *testing.T) {
	handler := newTestServer(t, Config{
		PublicBaseURL: "https://svk.example",
		DatabaseDSN:   "postgres://example",
		CookieSecret:  "secret",
		MailgunDomain: "mail.svk.example",
		MailgunAPIKey: "key",
	})

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

func TestDevModeAppAutoAuthenticatesAndServesStaticShell(t *testing.T) {
	dir := writeFakeSvelteBuild(t)
	handler := newTestServer(t, Config{DevMode: true, StaticDir: dir})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	body := res.Body.String()
	if !strings.Contains(body, `/app/boot.js`) {
		t.Fatalf("app shell did not include boot script: %s", body)
	}
	cookies := res.Result().Cookies()
	if !hasCookie(cookies, sessionCookieName) {
		t.Fatalf("missing session cookie: %#v", cookies)
	}
	if !hasCookie(cookies, csrfCookieName) {
		t.Fatalf("missing csrf cookie: %#v", cookies)
	}
}

func TestDevModeAppTrailingSlashServesStaticShell(t *testing.T) {
	dir := writeFakeSvelteBuild(t)
	handler := newTestServer(t, Config{DevMode: true, StaticDir: dir})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app/", nil)
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	body := res.Body.String()
	if strings.Contains(body, "Verify Ivy systems locally or with a hosted solver.") {
		t.Fatalf("trailing slash routed to marketing page: %s", body)
	}
	if !strings.Contains(body, `/app/boot.js`) {
		t.Fatalf("app shell did not include boot script: %s", body)
	}
}

func TestAuthenticatedAppRefreshesMissingCSRFCookie(t *testing.T) {
	dir := writeFakeSvelteBuild(t)
	handler := newTestServer(t, Config{DevMode: true, StaticDir: dir})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "user:dev-user"})
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	if !hasCookie(res.Result().Cookies(), csrfCookieName) {
		t.Fatalf("missing refreshed csrf cookie: %#v", res.Result().Cookies())
	}
}

func TestDevLoginCreatesSessionAndRedirectsToApp(t *testing.T) {
	handler := newTestServer(t, Config{DevMode: true})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/dev", nil)
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", res.Code)
	}
	if loc := res.Header().Get("location"); loc != "/app" {
		t.Fatalf("location = %q", loc)
	}
	if !hasCookie(res.Result().Cookies(), sessionCookieName) {
		t.Fatalf("missing session cookie")
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

func TestGeneratedSvelteBootScriptUsesBuiltEntriesAndAppBase(t *testing.T) {
	dir := writeFakeSvelteBuild(t)
	handler := newTestServer(t, Config{DevMode: true, StaticDir: dir})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app/boot.js", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-1"})
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	for _, want := range []string{
		`const element = document.getElementById("svelte");`,
		`globalThis.__sveltekit_abc123 = { base: "/app", assets: "" };`,
		`import("/_app/immutable/entry/start.test.js")`,
		`import("/_app/immutable/entry/app.test.js")`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("boot script missing %q: %s", want, body)
		}
	}
}

func TestStaticBuiltAssetsAreServed(t *testing.T) {
	dir := writeFakeSvelteBuild(t)
	handler := newTestServer(t, Config{DevMode: true, StaticDir: dir})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_app/immutable/entry/app.test.js", nil)
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	if body := res.Body.String(); !strings.Contains(body, "export const app") {
		t.Fatalf("body = %s", body)
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

func writeFakeSvelteBuild(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, subdir := range []string{
		filepath.Join(".vite"),
		filepath.Join("_app", "immutable", "entry"),
		filepath.Join("_app", "immutable", "chunks"),
		filepath.Join("_app", "immutable", "assets"),
	} {
		if err := os.MkdirAll(filepath.Join(dir, subdir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, path), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(".vite", "manifest.json"), `{
		".svelte-kit/generated/client-optimized/app.js": {
			"file": "_app/immutable/entry/app.test.js"
		},
		"node_modules/@sveltejs/kit/src/runtime/client/entry.js": {
			"file": "_app/immutable/entry/start.test.js"
		},
		".svelte-kit/generated/client-optimized/nodes/2.js": {
			"file": "_app/immutable/nodes/2.test.js",
			"css": ["_app/immutable/assets/2.test.css"]
		}
	}`)
	write(filepath.Join("_app", "immutable", "entry", "start.test.js"), `import { start } from "../chunks/runtime.test.js"; export { start };`)
	write(filepath.Join("_app", "immutable", "entry", "app.test.js"), `export const app = true;`)
	write(filepath.Join("_app", "immutable", "chunks", "runtime.test.js"), `const base = globalThis.__sveltekit_abc123?.base ?? "";`)
	write(filepath.Join("_app", "immutable", "assets", "2.test.css"), `body { color: black; }`)
	return dir
}

func hasCookie(cookies []*http.Cookie, name string) bool {
	for _, cookie := range cookies {
		if cookie.Name == name && cookie.Value != "" {
			return true
		}
	}
	return false
}
