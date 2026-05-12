package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPasskeyChallengeLifecycle(t *testing.T) {
	service, err := NewPasskeyService("localhost", "http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	challenge := service.RegistrationOptions("user-1")
	if err := service.FinishRegistration("user-1", challenge, "credential-1"); err != nil {
		t.Fatal(err)
	}
	if err := service.FinishRegistration("user-1", challenge, "credential-2"); err == nil {
		t.Fatal("expected replayed registration challenge to fail")
	}

	loginChallenge := service.LoginOptions("user-1")
	if err := service.FinishLogin("user-1", loginChallenge, "credential-1"); err != nil {
		t.Fatal(err)
	}
}

func TestPasskeyExpiredChallengeRejected(t *testing.T) {
	service, err := NewPasskeyService("localhost", "http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0).UTC()
	service.now = func() time.Time { return now }
	challenge := service.RegistrationOptions("user-1")
	now = now.Add(6 * time.Minute)
	if err := service.FinishRegistration("user-1", challenge, "credential-1"); err == nil {
		t.Fatal("expected expired challenge to fail")
	}
}

func TestPasskeyDisabledCredentialRejected(t *testing.T) {
	service, err := NewPasskeyService("localhost", "http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	challenge := service.RegistrationOptions("user-1")
	if err := service.FinishRegistration("user-1", challenge, "credential-1"); err != nil {
		t.Fatal(err)
	}
	service.DisableCredential("credential-1")
	loginChallenge := service.LoginOptions("user-1")
	if err := service.FinishLogin("user-1", loginChallenge, "credential-1"); err == nil {
		t.Fatal("expected disabled credential to fail")
	}
}

func TestPasskeyRoutesSetSessionCookieAfterLogin(t *testing.T) {
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
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Fatal("expected session cookie")
	}
}
