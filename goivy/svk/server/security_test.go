package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRFMissingOrWrongTokenRejected(t *testing.T) {
	srv, err := New(Config{DevMode: true, PublicBaseURL: "https://svk.example"})
	if err != nil {
		t.Fatal(err)
	}
	for name, setup := range map[string]func(*http.Request){
		"missing": func(*http.Request) {},
		"wrong": func(req *http.Request) {
			req.AddCookie(&http.Cookie{Name: "ivysvk_csrf", Value: "csrf"})
			req.Header.Set("x-csrf-token", "other")
		},
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(`{"name":"X","slug":"x"}`))
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "user:user-1"})
			setup(req)
			srv.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d", rec.Code)
			}
		})
	}
}

func TestOriginMismatchRejected(t *testing.T) {
	srv, err := New(Config{DevMode: true, PublicBaseURL: "https://svk.example"})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(`{"name":"X","slug":"x"}`))
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "user:user-1"})
	req.AddCookie(&http.Cookie{Name: "ivysvk_csrf", Value: "csrf"})
	req.Header.Set("x-csrf-token", "csrf")
	req.Header.Set("origin", "https://evil.example")
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCSPAllowsSvelteStyleAttributesButNotInlineScripts(t *testing.T) {
	rec := httptest.NewRecorder()
	securityHeaders(rec)
	csp := rec.Header().Get("content-security-policy")
	for _, want := range []string{
		"script-src 'self' 'wasm-unsafe-eval'",
		"style-src-elem 'self'",
		"style-src-attr 'unsafe-inline'",
		"frame-ancestors 'none'",
	} {
		if !strings.Contains(csp, want) {
			t.Fatalf("CSP missing %q: %s", want, csp)
		}
	}
	if strings.Contains(csp, "script-src 'self' 'unsafe-inline'") {
		t.Fatalf("CSP allows inline scripts: %s", csp)
	}
}

func TestSessionCookieRotatesAfterLogin(t *testing.T) {
	srv, err := New(Config{DevMode: true})
	if err != nil {
		t.Fatal(err)
	}
	challenge := srv.passkeys.RegistrationOptions("user-1")
	if err := srv.passkeys.FinishRegistration("user-1", challenge, "credential-1"); err != nil {
		t.Fatal(err)
	}
	loginChallenge := srv.passkeys.LoginOptions("user-1")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/passkey/login/finish", strings.NewReader(`{"userId":"user-1","challenge":"`+loginChallenge+`","credentialId":"credential-1"}`))
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "attacker-fixed"})
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == sessionCookieName && cookie.Value == "attacker-fixed" {
			t.Fatal("session fixation was not prevented")
		}
	}
}

func TestXSSStringsEscapedInGoTemplates(t *testing.T) {
	rec := httptest.NewRecorder()
	renderPage(rec, "<script>alert(1)</script>", "<img src=x onerror=alert(1)>")
	body := rec.Body.String()
	if strings.Contains(body, "<script>alert(1)</script>") || strings.Contains(body, "<img src=x") {
		t.Fatalf("template did not escape HTML: %s", body)
	}
}

func TestProjectCrossAccessDenied(t *testing.T) {
	store := NewProjectStore()
	project := store.CreatePersonalProject("owner", "Secret", "secret")
	if err := store.SaveModel("other", project.ID, "type t"); err == nil {
		t.Fatal("expected cross-project write to fail")
	}
}
