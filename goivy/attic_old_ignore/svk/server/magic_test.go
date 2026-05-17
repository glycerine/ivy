package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type fakeSender struct {
	to      string
	subject string
	text    string
}

func (f *fakeSender) Send(_ context.Context, to, subject, text string) error {
	f.to = to
	f.subject = subject
	f.text = text
	return nil
}

func TestMagicServiceTokenLifecycle(t *testing.T) {
	sender := &fakeSender{}
	service := NewMagicService(sender, "https://svk.example")
	service.now = func() time.Time { return time.Unix(100, 0).UTC() }

	if err := service.Request(context.Background(), "dev@example.test", "login"); err != nil {
		t.Fatal(err)
	}
	token := tokenFromMail(t, sender.text)
	if service.StoredRawToken(token) {
		t.Fatal("raw token was stored")
	}
	email, err := service.Consume(token)
	if err != nil {
		t.Fatal(err)
	}
	if email != "dev@example.test" {
		t.Fatalf("email = %q", email)
	}
	if _, err := service.Consume(token); err == nil {
		t.Fatal("expected consumed token reuse to fail")
	}
}

func TestMagicServiceRejectsExpiredTokens(t *testing.T) {
	sender := &fakeSender{}
	service := NewMagicService(sender, "https://svk.example")
	now := time.Unix(100, 0).UTC()
	service.now = func() time.Time { return now }
	if err := service.Request(context.Background(), "dev@example.test", "login"); err != nil {
		t.Fatal(err)
	}
	token := tokenFromMail(t, sender.text)
	now = now.Add(16 * time.Minute)
	if _, err := service.Consume(token); err == nil {
		t.Fatal("expected expired token to fail")
	}
}

func TestMagicRequestRouteReturnsGenericSuccessAndSendsMail(t *testing.T) {
	srv, err := New(Config{DevMode: true, PublicBaseURL: "https://svk.example"})
	if err != nil {
		t.Fatal(err)
	}
	sender := &fakeSender{}
	srv.magic = NewMagicService(sender, "https://svk.example")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/magic/request", strings.NewReader(`{"email":"unknown@example.test","purpose":"login"}`))
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload map[string]bool
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if !payload["ok"] {
		t.Fatal("expected generic ok response")
	}
	if sender.to != "unknown@example.test" {
		t.Fatalf("sender.to = %q", sender.to)
	}
	if !strings.Contains(sender.text, "https://svk.example/auth/magic/consume") {
		t.Fatalf("message missing magic link: %s", sender.text)
	}
}

func TestMagicConsumeRouteSetsSessionCookie(t *testing.T) {
	srv, err := New(Config{DevMode: true, PublicBaseURL: "https://svk.example"})
	if err != nil {
		t.Fatal(err)
	}
	sender := &fakeSender{}
	srv.magic = NewMagicService(sender, "https://svk.example")
	if err := srv.magic.Request(context.Background(), "dev@example.test", "login"); err != nil {
		t.Fatal(err)
	}
	token := tokenFromMail(t, sender.text)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/magic/consume?token="+url.QueryEscape(token), nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Fatal("expected session cookie")
	}
}

func tokenFromMail(t *testing.T, text string) string {
	t.Helper()
	start := strings.Index(text, "https://")
	if start < 0 {
		t.Fatalf("no url in %q", text)
	}
	fields := strings.Fields(text[start:])
	u, err := url.Parse(fields[0])
	if err != nil {
		t.Fatal(err)
	}
	token := u.Query().Get("token")
	if token == "" {
		t.Fatal("missing token")
	}
	return token
}
