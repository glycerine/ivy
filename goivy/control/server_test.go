package control

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const defaultTestDatabaseDSN = "postgres://ivyvue_app:ivyvue_app_dev@127.0.0.1:5432/ivyvue?sslmode=disable"

func newTestStore(t *testing.T) *PostgresStore {
	t.Helper()
	dsn := os.Getenv("IVY_CONTROL_TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = defaultTestDatabaseDSN
	}
	store, err := OpenPostgresStore(dsn)
	if err != nil {
		t.Fatalf("open postgres test store: %v", err)
	}
	if err := store.Ping(context.Background()); err != nil {
		t.Fatalf("ping postgres test store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close postgres test store: %v", err)
		}
	})
	return store
}

func TestAuthMeUnauthenticated(t *testing.T) {
	server := NewServer(Config{Store: newTestStore(t)})
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

func TestStaticWebvueIndexAndAssetsAreServed(t *testing.T) {
	dir := t.TempDir()
	dist := filepath.Join(dir, "dist")
	if err := os.MkdirAll(dist, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!doctype html><div id=\"app\"></div>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dist, "ivywebvue.js"), []byte("console.log('webvue')"), 0o644); err != nil {
		t.Fatal(err)
	}
	server := NewServer(Config{Store: newTestStore(t), StaticDir: dir})

	indexReq := httptest.NewRequest(http.MethodGet, "/", nil)
	indexRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(indexRec, indexReq)
	if indexRec.Code != http.StatusOK || !strings.Contains(indexRec.Body.String(), `id="app"`) {
		t.Fatalf("index response = %d %q", indexRec.Code, indexRec.Body.String())
	}

	verifiedReq := httptest.NewRequest(http.MethodGet, "/verified", nil)
	verifiedRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(verifiedRec, verifiedReq)
	if verifiedRec.Code != http.StatusOK || !strings.Contains(verifiedRec.Body.String(), `id="app"`) {
		t.Fatalf("verified response = %d %q", verifiedRec.Code, verifiedRec.Body.String())
	}

	continueReq := httptest.NewRequest(http.MethodGet, "/auth/email/continue", nil)
	continueRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(continueRec, continueReq)
	if continueRec.Code != http.StatusOK || !strings.Contains(continueRec.Body.String(), `id="app"`) {
		t.Fatalf("email continue response = %d %q", continueRec.Code, continueRec.Body.String())
	}

	assetReq := httptest.NewRequest(http.MethodGet, "/static/dist/ivywebvue.js", nil)
	assetRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(assetRec, assetReq)
	if assetRec.Code != http.StatusOK || !strings.Contains(assetRec.Body.String(), "webvue") {
		t.Fatalf("asset response = %d %q", assetRec.Code, assetRec.Body.String())
	}
}

func TestEmailContinueUsesVueAppInsteadOfRedirectShim(t *testing.T) {
	server := NewServer(Config{Store: newTestStore(t)})
	req := httptest.NewRequest(http.MethodGet, "/auth/email/continue", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `id="app"`) {
		t.Fatalf("continue page does not serve the Vue app:\n%s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `location.replace`) || strings.Contains(rec.Body.String(), `auth=link-expired`) || strings.Contains(rec.Body.String(), `Sign-in link is invalid or expired.`) {
		t.Fatalf("continue page still renders the old redirect shim:\n%s", rec.Body.String())
	}
}

func TestEmailLoginRequestUsesDatabaseOutboxWithoutRealEmail(t *testing.T) {
	store := newTestStore(t)
	server := NewServer(Config{
		Store:                     store,
		EmailSender:               NewDatabaseEmailSender(store),
		AutoProvisionStarterSpace: true,
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/email/request", strings.NewReader(`{"email":"Alice@Example.Test"}`))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	message, ok, err := store.LatestEmailDeliveryFor(context.Background(), "alice@example.test")
	if err != nil {
		t.Fatalf("load latest email delivery: %v", err)
	}
	if !ok {
		t.Fatalf("database email outbox did not receive login email")
	}
	if !strings.Contains(message.LoginURL, "/auth/email/continue#token=") {
		t.Fatalf("login URL = %q", message.LoginURL)
	}
}

func TestEmailLoginRequestLogsTestOutboxLinkWhenEnabled(t *testing.T) {
	var logs bytes.Buffer
	store := newTestStore(t)
	server := NewServer(Config{
		Store:                     store,
		EmailSender:               NewDatabaseEmailSender(store),
		AutoProvisionStarterSpace: true,
		EnableTestEmailOutbox:     true,
		Logger:                    log.New(&logs, "", 0),
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/email/request", strings.NewReader(`{"email":"Alice@Example.Test"}`))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	got := logs.String()
	for _, want := range []string{
		"email_login_request received email=alice@example.test",
		"email_login_request queued email=alice@example.test",
		"test_email_outbox login_link email=alice@example.test",
		"/auth/email/continue#token=",
		"http method=POST path=/auth/email/request status=200",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("logs missing %q in:\n%s", want, got)
		}
	}
}

func TestAdminUnverifiedEmailsShowsDatabaseOutboxLoginLink(t *testing.T) {
	store := newTestStore(t)
	server := NewServer(Config{
		Store:                     store,
		EmailSender:               NewDatabaseEmailSender(store),
		AutoProvisionStarterSpace: true,
	})
	suffix, err := NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	email := "pending+" + suffix[:8] + "@example.test"
	request := httptest.NewRequest(http.MethodPost, "/auth/email/request", strings.NewReader(`{"email":"`+email+`"}`))
	requestRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(requestRec, request)
	if requestRec.Code != http.StatusOK {
		t.Fatalf("request status = %d, want %d; body=%s", requestRec.Code, http.StatusOK, requestRec.Body.String())
	}

	adminReq := httptest.NewRequest(http.MethodGet, "/admin/api/unverified-emails", nil)
	adminRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusOK {
		t.Fatalf("admin status = %d, want %d; body=%s", adminRec.Code, http.StatusOK, adminRec.Body.String())
	}
	var payload struct {
		Emails []AdminUnverifiedEmail `json:"emails"`
	}
	if err := json.NewDecoder(adminRec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode admin response: %v", err)
	}
	for _, row := range payload.Emails {
		if row.Email == email {
			if !strings.Contains(row.LoginURL, "/auth/email/continue#token=") {
				t.Fatalf("login URL = %q", row.LoginURL)
			}
			if row.Expired || row.UsedAt != nil {
				t.Fatalf("unexpected email status: %#v", row)
			}
			return
		}
	}
	t.Fatalf("admin response missing %s: %#v", email, payload.Emails)
}

func TestEmailMagicLinkCreatesAppSessionAndStarterProject(t *testing.T) {
	store := newTestStore(t)
	server := NewServer(Config{
		Store:                     store,
		EmailSender:               NewDatabaseEmailSender(store),
		AutoProvisionStarterSpace: true,
	})
	request := httptest.NewRequest(http.MethodPost, "/auth/email/request", strings.NewReader(`{"email":"alice@example.test"}`))
	request.Host = "ivy.example.test"
	requestRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(requestRec, request)
	if requestRec.Code != http.StatusOK {
		t.Fatalf("request status = %d, want %d; body=%s", requestRec.Code, http.StatusOK, requestRec.Body.String())
	}
	message, ok, err := store.LatestEmailDeliveryFor(context.Background(), "alice@example.test")
	if err != nil {
		t.Fatalf("load latest email delivery: %v", err)
	}
	if !ok {
		t.Fatalf("missing login email")
	}
	token := message.LoginURL[strings.LastIndex(message.LoginURL, "#token=")+len("#token="):]

	consume := httptest.NewRequest(http.MethodPost, "/auth/email/consume", strings.NewReader(`{"token":"`+token+`"}`))
	consumeRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(consumeRec, consume)
	if consumeRec.Code != http.StatusOK {
		t.Fatalf("consume status = %d, want %d; body=%s", consumeRec.Code, http.StatusOK, consumeRec.Body.String())
	}
	var sessionCookie *http.Cookie
	for _, cookie := range consumeRec.Result().Cookies() {
		if cookie.Name == SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatalf("missing app session cookie: %#v", consumeRec.Result().Cookies())
	}
	if sessionCookie.MaxAge != int(AppSessionTTL.Seconds()) {
		t.Fatalf("session cookie max-age = %d, want %d", sessionCookie.MaxAge, int(AppSessionTTL.Seconds()))
	}

	authReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	authReq.AddCookie(sessionCookie)
	authRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(authRec, authReq)

	if authRec.Code != http.StatusOK {
		t.Fatalf("/auth/me status = %d, want %d; body=%s", authRec.Code, http.StatusOK, authRec.Body.String())
	}
	var view SessionView
	if err := json.NewDecoder(authRec.Body).Decode(&view); err != nil {
		t.Fatalf("decode auth view: %v", err)
	}
	if !view.Authenticated || view.User == nil || view.User.Email != "alice@example.test" {
		t.Fatalf("bad auth view: %#v", view)
	}
	if view.User.EmailVerifiedAt == nil {
		t.Fatalf("email was not marked verified")
	}
	if len(view.Accounts) != 1 || len(view.Teams) != 1 || len(view.Projects) != 1 {
		t.Fatalf("starter workspace missing: accounts=%d teams=%d projects=%d", len(view.Accounts), len(view.Teams), len(view.Projects))
	}
}

func TestLandingPageRefreshesCookieOnlyOnNewDayAndRecordsVisitHour(t *testing.T) {
	store := newTestStore(t)
	server := NewServer(Config{Store: store, StaticIndex: `<!doctype html><div id="app"></div>`})
	ctx := context.Background()
	subject, err := NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.UpsertUserFromOIDC(ctx, OIDCIdentity{
		Issuer:        "email",
		Subject:       subject,
		Email:         "landing+" + subject[:8] + "@example.test",
		DisplayName:   "Landing Tester",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("upsert user: %v", err)
	}
	sessionToken, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	csrfToken, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := store.CreateAppSession(ctx, user.ID, sessionToken, csrfToken, now, AppSessionTTL, AppSessionTTL); err != nil {
		t.Fatalf("create session: %v", err)
	}
	cookie := &http.Cookie{Name: SessionCookieName, Value: sessionToken}

	sameDayReq := httptest.NewRequest(http.MethodGet, "/", nil)
	sameDayReq.AddCookie(cookie)
	sameDayRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(sameDayRec, sameDayReq)
	if sameDayRec.Code != http.StatusOK {
		t.Fatalf("same-day landing status = %d, want %d", sameDayRec.Code, http.StatusOK)
	}
	for _, refreshed := range sameDayRec.Result().Cookies() {
		if refreshed.Name == SessionCookieName && refreshed.MaxAge > 0 {
			t.Fatalf("same-day landing refreshed cookie unexpectedly: %#v", refreshed)
		}
	}

	yesterday := now.Add(-25 * time.Hour)
	if _, err := store.db.ExecContext(ctx, `UPDATE app_sessions SET last_seen_at = $2 WHERE id_hash = $1`, hashToken(sessionToken), yesterday); err != nil {
		t.Fatalf("backdate session: %v", err)
	}
	nextDayReq := httptest.NewRequest(http.MethodGet, "/", nil)
	nextDayReq.AddCookie(cookie)
	nextDayRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(nextDayRec, nextDayReq)
	if nextDayRec.Code != http.StatusOK {
		t.Fatalf("next-day landing status = %d, want %d", nextDayRec.Code, http.StatusOK)
	}
	var refreshedCookie *http.Cookie
	for _, cookie := range nextDayRec.Result().Cookies() {
		if cookie.Name == SessionCookieName {
			refreshedCookie = cookie
			break
		}
	}
	if refreshedCookie == nil || refreshedCookie.MaxAge != int(AppSessionTTL.Seconds()) {
		t.Fatalf("next-day landing cookie = %#v, want refreshed %s cookie", refreshedCookie, AppSessionTTL)
	}
	var visits int
	if err := store.db.QueryRowContext(ctx, `SELECT count(*) FROM visiting_hours WHERE user_id = $1`, user.ID).Scan(&visits); err != nil {
		t.Fatalf("count visiting hours: %v", err)
	}
	if visits != 1 {
		t.Fatalf("visiting_hours rows = %d, want 1 for repeated same-hour landing hits", visits)
	}
}

func TestPasskeyRegistrationOptionsRequireAuthenticatedSessionAndStoreChallenge(t *testing.T) {
	store := newTestStore(t)
	server := NewServer(Config{Store: store})

	unauthenticatedReq := httptest.NewRequest(http.MethodPost, "/auth/passkeys/register/options", nil)
	unauthenticatedRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(unauthenticatedRec, unauthenticatedReq)
	if unauthenticatedRec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d", unauthenticatedRec.Code, http.StatusUnauthorized)
	}

	ctx := context.Background()
	subject, err := NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.UpsertUserFromOIDC(ctx, OIDCIdentity{
		Issuer:        "email",
		Subject:       subject,
		Email:         "passkey-register+" + subject[:8] + "@example.test",
		DisplayName:   "Passkey Register",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("upsert user: %v", err)
	}
	sessionToken, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	csrfToken, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := store.CreateAppSession(ctx, user.ID, sessionToken, csrfToken, now, AppSessionTTL, AppSessionTTL); err != nil {
		t.Fatalf("create session: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/passkeys/register/options", nil)
	req.Host = "127.0.0.1:18080"
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sessionToken})
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload passkeyRegistrationOptionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.PublicKey.Challenge == "" {
		t.Fatalf("missing passkey challenge: %#v", payload.PublicKey)
	}
	if payload.PublicKey.RP.Name != "Ivy" {
		t.Fatalf("rp = %#v", payload.PublicKey.RP)
	}
	if payload.PublicKey.User.Name != user.Email {
		t.Fatalf("user name = %q, want %q", payload.PublicKey.User.Name, user.Email)
	}
	challenge, err := decodeBase64URL(payload.PublicKey.Challenge)
	if err != nil {
		t.Fatalf("decode challenge: %v", err)
	}
	if err := store.ConsumePasskeyChallenge(ctx, user.ID, "registration", challenge, "127.0.0.1", "http://127.0.0.1:18080", time.Now().UTC()); err != nil {
		t.Fatalf("stored passkey registration challenge was not consumable: %v", err)
	}
}

func TestPasskeyLoginOptionsStoreAnonymousDiscoverableChallenge(t *testing.T) {
	store := newTestStore(t)
	server := NewServer(Config{Store: store})
	req := httptest.NewRequest(http.MethodPost, "/auth/passkeys/login/options", nil)
	req.Host = "localhost:18080"
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload passkeyLoginOptionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.PublicKey.Challenge == "" || payload.PublicKey.RPID != "localhost" {
		t.Fatalf("bad passkey login options: %#v", payload.PublicKey)
	}
	challenge, err := decodeBase64URL(payload.PublicKey.Challenge)
	if err != nil {
		t.Fatalf("decode challenge: %v", err)
	}
	if err := store.ConsumePasskeyChallenge(context.Background(), "", "login", challenge, "localhost", "http://localhost:18080", time.Now().UTC()); err != nil {
		t.Fatalf("stored passkey login challenge was not consumable: %v", err)
	}
}

func TestLoginRedirectsToOIDCAuthorizeURL(t *testing.T) {
	server := NewServer(Config{
		Store: newTestStore(t),
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
	server := NewServer(Config{Store: newTestStore(t)})
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state=attacker", nil)
	req.AddCookie(&http.Cookie{Name: OIDCStateCookie, Value: "real-state"})
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCallbackCreatesAppSessionAndAuthMeShowsStarterProject(t *testing.T) {
	store := newTestStore(t)
	server := NewServer(Config{
		Store:                     store,
		OIDCProvider:              fakeOIDCProvider{},
		AutoProvisionStarterSpace: true,
	})
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state=real-state&code=ok", nil)
	req.AddCookie(&http.Cookie{Name: OIDCStateCookie, Value: "real-state"})
	req.AddCookie(&http.Cookie{Name: OIDCNonceCookie, Value: "real-nonce"})
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusFound, rec.Body.String())
	}
	var sessionCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatalf("missing app session cookie: %#v", rec.Result().Cookies())
	}

	authReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	authReq.AddCookie(sessionCookie)
	authRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(authRec, authReq)

	if authRec.Code != http.StatusOK {
		t.Fatalf("/auth/me status = %d, want %d; body=%s", authRec.Code, http.StatusOK, authRec.Body.String())
	}
	var view SessionView
	if err := json.NewDecoder(authRec.Body).Decode(&view); err != nil {
		t.Fatalf("decode auth view: %v", err)
	}
	if !view.Authenticated {
		t.Fatalf("authenticated = false, want true")
	}
	if view.User == nil || view.User.Email != "oidc-alice@example.test" {
		t.Fatalf("user = %#v", view.User)
	}
	if len(view.Accounts) != 1 || len(view.Teams) != 1 || len(view.Projects) != 1 {
		t.Fatalf("starter workspace missing: accounts=%d teams=%d projects=%d", len(view.Accounts), len(view.Teams), len(view.Projects))
	}
	if view.Projects[0].Slug != "client-server" {
		t.Fatalf("project slug = %q, want client-server", view.Projects[0].Slug)
	}
	if view.Roles[view.Projects[0].ID] != string(ProjectRoleAdmin) {
		t.Fatalf("project role = %q, want admin", view.Roles[view.Projects[0].ID])
	}
}

type fakeOIDCProvider struct{}

func (fakeOIDCProvider) ExchangeCode(ctx context.Context, code, nonce string) (OIDCIdentity, error) {
	return OIDCIdentity{
		Issuer:        "http://oauth.example.test",
		Subject:       "alice-subject",
		Email:         "oidc-alice@example.test",
		DisplayName:   "Alice",
		EmailVerified: true,
	}, nil
}
