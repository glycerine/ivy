package control

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAuthMeUnauthenticated(t *testing.T) {
	server := NewServer(Config{})
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var view SessionView
	if err := json.NewDecoder(rec.Body).Decode(&view); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if view.Authenticated {
		t.Fatalf("authenticated = true, want false")
	}
}

func TestLoginRedirectsToCasdoorAuthorizeURL(t *testing.T) {
	server := NewServer(Config{
		OIDC: OIDCConfig{
			AuthURL:     "http://127.0.0.1:18082/login/oauth/authorize",
			ClientID:    "ivy-control-local",
			RedirectURL: "http://127.0.0.1:18080/auth/callback",
			Scopes:      []string{"openid", "profile", "email"},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusFound, rec.Body.String())
	}
	location := rec.Header().Get("Location")
	if location == "" {
		t.Fatalf("missing redirect Location")
	}
	u, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	if got := u.Scheme + "://" + u.Host + u.Path; got != "http://127.0.0.1:18082/login/oauth/authorize" {
		t.Fatalf("authorize endpoint = %q", got)
	}
	q := u.Query()
	for key, want := range map[string]string{
		"response_type": "code",
		"client_id":     "ivy-control-local",
		"redirect_uri":  "http://127.0.0.1:18080/auth/callback",
	} {
		if got := q.Get(key); got != want {
			t.Fatalf("query %s = %q, want %q", key, got, want)
		}
	}
	if !strings.Contains(q.Get("scope"), "openid") {
		t.Fatalf("scope %q does not contain openid", q.Get("scope"))
	}
	if q.Get("state") == "" || q.Get("nonce") == "" {
		t.Fatalf("missing state or nonce in redirect: %s", location)
	}
	if rec.Result().Cookies()[0].Name != OIDCStateCookie {
		t.Fatalf("first cookie = %q, want %q", rec.Result().Cookies()[0].Name, OIDCStateCookie)
	}
}

func TestCallbackRejectsMismatchedState(t *testing.T) {
	server := NewServer(Config{})
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state=attacker", nil)
	req.AddCookie(&http.Cookie{Name: OIDCStateCookie, Value: "real-state"})
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
