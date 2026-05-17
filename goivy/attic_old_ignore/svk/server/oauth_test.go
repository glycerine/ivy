package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOAuthStartSetsStateCookieAndRedirects(t *testing.T) {
	srv, err := New(Config{DevMode: true})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/google/start", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d", rec.Code)
	}
	if len(rec.Result().Cookies()) != 1 {
		t.Fatalf("cookies = %v", rec.Result().Cookies())
	}
	if !strings.Contains(rec.Header().Get("location"), "code_challenge_method=S256") {
		t.Fatalf("location = %q", rec.Header().Get("location"))
	}
}

func TestOAuthCallbackRejectsInvalidState(t *testing.T) {
	srv, err := New(Config{DevMode: true})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/github/callback?state=bad&code=subj&email=a@example.test", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookie, Value: "other"})
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestOAuthServiceDoesNotAutoLinkSameEmailAcrossProviders(t *testing.T) {
	oauth := NewOAuthService()
	stateGoogle, _, _ := oauth.Start("google")
	first, err := oauth.Callback("google", stateGoogle, "google-subject", "same@example.test")
	if err != nil {
		t.Fatal(err)
	}

	stateGitHub, _, _ := oauth.Start("github")
	second, err := oauth.Callback("github", stateGitHub, "github-subject", "same@example.test")
	if !errors.Is(err, ErrExplicitLinkRequired) {
		t.Fatalf("expected explicit linking error, got %v", err)
	}
	if second.UserID != first.UserID {
		t.Fatalf("expected existing user id in linking flow")
	}
}

func TestOAuthExplicitLinkAttachesIdentity(t *testing.T) {
	oauth := NewOAuthService()
	if err := oauth.LinkIdentity("user-1", "github", "subject-1"); err != nil {
		t.Fatal(err)
	}
	state, _, _ := oauth.Start("github")
	result, err := oauth.Callback("github", state, "subject-1", "person@example.test")
	if err != nil {
		t.Fatal(err)
	}
	if result.UserID != "user-1" {
		t.Fatalf("user = %q", result.UserID)
	}
}

func TestOAuthCallbackCreatesSessionForNewIdentity(t *testing.T) {
	srv, err := New(Config{DevMode: true})
	if err != nil {
		t.Fatal(err)
	}
	state, _, _ := srv.oauth.Start("apple")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/apple/callback?state="+state+"&code=apple-subject&email=apple@example.test", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookie, Value: state})
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Fatal("expected session cookie")
	}
}
