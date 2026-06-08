package control

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHTTPOIDCProviderExchangesCodeAndValidatesIDToken(t *testing.T) {
	t.Skip("control not in use at the moment, skip sandbox violating test")

	idp := NewTestIDP("", "ivy-control-test")
	mux := http.NewServeMux()
	idp.Routes(mux)
	server := httptest.NewServer(mux)
	defer server.Close()
	idp.issuer = server.URL + "/test-idp"

	authorizeURL := idp.issuer + "/login/oauth/authorize"
	redirectURL := "http://app.example.test/auth/callback"
	location := authorizeForTest(t, authorizeURL, redirectURL, "state-1", "nonce-1")
	code := location.Query().Get("code")
	if code == "" {
		t.Fatalf("authorize redirect missing code: %s", location.String())
	}

	provider := NewHTTPOIDCProvider(OIDCConfig{
		IssuerURL:   idp.issuer,
		TokenURL:    idp.issuer + "/api/login/oauth/access_token",
		JWKSURL:     idp.issuer + "/jwks",
		ClientID:    "ivy-control-test",
		RedirectURL: redirectURL,
	}, server.Client())
	identity, err := provider.ExchangeCode(context.Background(), code, "nonce-1")
	if err != nil {
		t.Fatalf("exchange code: %v", err)
	}
	if identity.Issuer != idp.issuer {
		t.Fatalf("issuer = %q, want %q", identity.Issuer, idp.issuer)
	}
	if identity.Subject != "test-user-1" || identity.Email != "tester@example.test" || !identity.EmailVerified {
		t.Fatalf("identity = %#v", identity)
	}
}

func authorizeForTest(t *testing.T, authorizeURL, redirectURL, state, nonce string) *url.URL {
	t.Helper()
	u, err := url.Parse(authorizeURL)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", "ivy-control-test")
	q.Set("redirect_uri", redirectURL)
	q.Set("state", state)
	q.Set("nonce", nonce)
	u.RawQuery = q.Encode()

	client := http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(u.String())
	if err != nil {
		t.Fatalf("authorize request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("authorize status = %d, want %d", resp.StatusCode, http.StatusFound)
	}
	location, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	return location
}
