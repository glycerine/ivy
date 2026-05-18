package control

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	SessionCookieName = "ivy_webui_session"
	OIDCStateCookie   = "ivy_webui_oidc_state"
	OIDCNonceCookie   = "ivy_webui_oidc_nonce"
)

type Config struct {
	Addr                      string
	PublicBaseURL             string
	OIDC                      OIDCConfig
	OIDCProvider              OIDCProvider
	Store                     Store
	EmailSender               EmailSender
	StaticIndex               string
	StaticDir                 string
	Logger                    *log.Logger
	CookieSecure              bool
	AutoProvisionStarterSpace bool
	EnableTestIDP             bool
	EnableTestEmailOutbox     bool
}

type Server struct {
	cfg     Config
	mux     *http.ServeMux
	store   Store
	oidc    OIDCProvider
	email   EmailSender
	testIDP *TestIDP
}

func NewServer(cfg Config) *Server {
	if cfg.Store == nil {
		panic("control.NewServer requires a PostgreSQL-backed Store")
	}
	if cfg.OIDCProvider == nil {
		cfg.OIDCProvider = NewHTTPOIDCProvider(cfg.OIDC, nil)
	}
	if cfg.EmailSender == nil {
		pg, ok := cfg.Store.(*PostgresStore)
		if !ok {
			panic("control.NewServer requires an EmailSender")
		}
		cfg.EmailSender = NewDatabaseEmailSender(pg)
	}
	s := &Server{
		cfg:   cfg,
		mux:   http.NewServeMux(),
		store: cfg.Store,
		oidc:  cfg.OIDCProvider,
		email: cfg.EmailSender,
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	if s.cfg.Logger != nil {
		return s.accessLog(s.mux)
	}
	return s.mux
}

func (s *Server) Start() error {
	addr := s.cfg.Addr
	if addr == "" {
		addr = "127.0.0.1:18080"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	limited := newSocketLimitListener(ln, defaultSocketLimitConfig())
	server := &http.Server{Handler: s.Handler()}
	return server.Serve(limited)
}

func (s *Server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.logf("http method=%s path=%s status=%d duration=%s remote=%s", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond), r.RemoteAddr)
	})
}

func (s *Server) routes() {
	if s.cfg.EnableTestIDP {
		s.testIDP = NewTestIDP(s.cfg.OIDC.IssuerURL, s.cfg.OIDC.ClientID)
		s.testIDP.Routes(s.mux)
	}
	if strings.TrimSpace(s.cfg.StaticDir) != "" {
		s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.cfg.StaticDir))))
	}
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /verified", s.handleIndex)
	s.mux.HandleFunc("GET /admin", s.handleIndex)
	s.mux.HandleFunc("GET /admin/", s.handleIndex)
	s.mux.HandleFunc("GET /admin/api/unverified-emails", s.handleAdminUnverifiedEmails)
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /auth/me", s.handleAuthMe)
	s.mux.HandleFunc("POST /auth/email/request", s.handleEmailLoginRequest)
	s.mux.HandleFunc("GET /auth/email/continue", s.handleIndex)
	s.mux.HandleFunc("POST /auth/email/consume", s.handleEmailLoginConsume)
	s.mux.HandleFunc("POST /auth/passkeys/register/options", s.handlePasskeyRegisterOptions)
	s.mux.HandleFunc("POST /auth/passkeys/register/finish", s.handlePasskeyRegisterFinish)
	s.mux.HandleFunc("POST /auth/passkeys/login/options", s.handlePasskeyLoginOptions)
	s.mux.HandleFunc("POST /auth/passkeys/login/finish", s.handlePasskeyLoginFinish)
	s.mux.HandleFunc("GET /auth/login", s.handleLogin)
	s.mux.HandleFunc("GET /auth/callback", s.handleCallback)
	s.mux.HandleFunc("POST /auth/logout", s.handleLogout)
	if s.cfg.EnableTestEmailOutbox {
		s.mux.HandleFunc("GET /test/email/latest", s.handleTestEmailLatest)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.touchLandingSession(w, r)
	if strings.TrimSpace(s.cfg.StaticIndex) != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(s.cfg.StaticIndex))
		return
	}
	if strings.TrimSpace(s.cfg.StaticDir) != "" {
		indexPath := filepath.Join(s.cfg.StaticDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte("<!doctype html><title>Ivy</title><div id=\"app\">Ivy control-plane</div>"))
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		writeJSON(w, http.StatusOK, SessionView{Authenticated: false})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid session cookie"})
		return
	}
	view, err := s.store.SessionViewByToken(r.Context(), cookie.Value, time.Now().UTC(), AppSessionTTL)
	if errors.Is(err, ErrSessionNotFound) {
		clearCookie(w, SessionCookieName, s.cfg.CookieSecure)
		writeJSON(w, http.StatusOK, SessionView{Authenticated: false})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load session"})
		return
	}
	if view.CookieRefreshNeeded {
		setAppCookie(w, SessionCookieName, cookie.Value, AppSessionTTL, s.cfg.CookieSecure)
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) touchLandingSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		return
	}
	if err != nil {
		return
	}
	touch, err := s.store.TouchSessionByToken(r.Context(), cookie.Value, time.Now().UTC(), AppSessionTTL)
	if errors.Is(err, ErrSessionNotFound) {
		clearCookie(w, SessionCookieName, s.cfg.CookieSecure)
		return
	}
	if err != nil {
		s.logf("session_touch failed path=%s remote=%s error=%q", r.URL.Path, r.RemoteAddr, err)
		return
	}
	if touch.CookieRefreshNeeded {
		setAppCookie(w, SessionCookieName, cookie.Value, AppSessionTTL, s.cfg.CookieSecure)
	}
}

func (s *Server) handleEmailLoginRequest(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.logf("email_login_request invalid_json remote=%s error=%q", r.RemoteAddr, err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	email, normalizeErr := NormalizeEmail(payload.Email)
	if normalizeErr == nil {
		s.logf("email_login_request received email=%s remote=%s", email, r.RemoteAddr)
	} else {
		s.logf("email_login_request invalid_email remote=%s error=%q", r.RemoteAddr, normalizeErr)
	}
	service := EmailAuthService{
		Store:                     s.store,
		Sender:                    s.email,
		BaseURL:                   s.publicBaseURL(r),
		AutoProvisionStarterSpace: s.cfg.AutoProvisionStarterSpace,
	}
	if err := service.RequestLogin(r.Context(), payload.Email, time.Now().UTC()); err != nil {
		s.logf("email_login_request failed email=%s error=%q", email, err)
	} else if normalizeErr == nil {
		s.logf("email_login_request queued email=%s ttl=%s", email, EmailLoginTokenTTL)
		if s.cfg.EnableTestEmailOutbox {
			if store, ok := s.store.(*PostgresStore); ok {
				if message, ok, err := store.LatestEmailDeliveryFor(r.Context(), email); err == nil && ok {
					s.logf("test_email_outbox login_link email=%s expires_at=%s url=%s", email, message.ExpiresAt.Format(time.RFC3339), message.LoginURL)
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleEmailLoginConsume(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	service := EmailAuthService{
		Store:                     s.store,
		Sender:                    s.email,
		BaseURL:                   s.publicBaseURL(r),
		AutoProvisionStarterSpace: s.cfg.AutoProvisionStarterSpace,
	}
	sessionToken, _, user, err := service.ConsumeLogin(r.Context(), payload.Token, time.Now().UTC())
	if errors.Is(err, ErrEmailLoginTokenNotFound) {
		s.logf("email_login_consume rejected reason=invalid_or_expired remote=%s", r.RemoteAddr)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired login token"})
		return
	}
	if err != nil {
		s.logf("email_login_consume failed remote=%s error=%q", r.RemoteAddr, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create app session"})
		return
	}
	s.logf("email_login_consume accepted user_id=%s email=%s remote=%s", user.ID, user.Email, r.RemoteAddr)
	setAppCookie(w, SessionCookieName, sessionToken, AppSessionTTL, s.cfg.CookieSecure)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	s.logf("oidc_login_start remote=%s", r.RemoteAddr)
	state, err := RandomToken(32)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create login state"})
		return
	}
	nonce, err := RandomToken(32)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create login nonce"})
		return
	}
	loginURL, err := s.cfg.OIDC.LoginURL(state, nonce)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	setTransientCookie(w, OIDCStateCookie, state, 10*time.Minute)
	setTransientCookie(w, OIDCNonceCookie, nonce, 10*time.Minute)
	http.Redirect(w, r, loginURL, http.StatusFound)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(OIDCStateCookie)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid oidc state"})
		return
	}
	nonceCookie, err := r.Cookie(OIDCNonceCookie)
	if err != nil || strings.TrimSpace(nonceCookie.Value) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid oidc nonce"})
		return
	}
	identity, err := s.oidc.ExchangeCode(r.Context(), r.URL.Query().Get("code"), nonceCookie.Value)
	if err != nil {
		s.logf("oidc_callback exchange_failed remote=%s error=%q", r.RemoteAddr, err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "oidc exchange failed"})
		return
	}
	user, err := s.store.UpsertUserFromOIDC(r.Context(), identity)
	if err != nil {
		s.logf("oidc_callback map_failed issuer=%s subject=%s email=%s error=%q", identity.Issuer, identity.Subject, identity.Email, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to map oidc user"})
		return
	}
	if s.cfg.AutoProvisionStarterSpace {
		if err := s.store.EnsureStarterWorkspace(r.Context(), user); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to provision starter workspace"})
			return
		}
	}
	sessionToken, err := RandomToken(32)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create app session"})
		return
	}
	csrfToken, err := RandomToken(32)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create csrf token"})
		return
	}
	now := time.Now().UTC()
	if err := s.store.CreateAppSession(r.Context(), user.ID, sessionToken, csrfToken, now, AppSessionTTL, AppSessionTTL); err != nil {
		s.logf("oidc_callback session_failed user_id=%s email=%s error=%q", user.ID, user.Email, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to store app session"})
		return
	}
	s.logf("oidc_callback accepted user_id=%s email=%s", user.ID, user.Email)
	setAppCookie(w, SessionCookieName, sessionToken, AppSessionTTL, s.cfg.CookieSecure)
	clearCookie(w, OIDCStateCookie, s.cfg.CookieSecure)
	clearCookie(w, OIDCNonceCookie, s.cfg.CookieSecure)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleTestEmailLatest(w http.ResponseWriter, r *http.Request) {
	store, ok := s.store.(*PostgresStore)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "test email outbox database is not available"})
		return
	}
	email, err := NormalizeEmail(r.URL.Query().Get("email"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid email query is required"})
		return
	}
	message, ok, err := store.LatestEmailDeliveryFor(r.Context(), email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load email outbox"})
		return
	}
	if !ok {
		s.logf("test_email_latest miss email=%s", email)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no message for email"})
		return
	}
	s.logf("test_email_latest hit email=%s expires_at=%s", email, message.ExpiresAt.Format(time.RFC3339))
	writeJSON(w, http.StatusOK, message)
}

func (s *Server) handleAdminUnverifiedEmails(w http.ResponseWriter, r *http.Request) {
	store, ok := s.store.(*PostgresStore)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "admin dashboard requires PostgreSQL"})
		return
	}
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be a positive integer"})
			return
		}
		limit = parsed
	}
	emails, err := store.AdminUnverifiedEmails(r.Context(), time.Now().UTC(), limit)
	if err != nil {
		s.logf("admin_unverified_emails failed error=%q", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load unverified emails"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"emails": emails})
}

func (s *Server) publicBaseURL(r *http.Request) string {
	if strings.TrimSpace(s.cfg.PublicBaseURL) != "" {
		return strings.TrimRight(s.cfg.PublicBaseURL, "/")
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.logf("logout remote=%s", r.RemoteAddr)
	clearCookie(w, SessionCookieName, s.cfg.CookieSecure)
	clearCookie(w, OIDCStateCookie, s.cfg.CookieSecure)
	clearCookie(w, OIDCNonceCookie, s.cfg.CookieSecure)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) logf(format string, args ...any) {
	if s.cfg.Logger != nil {
		s.cfg.Logger.Printf(format, args...)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func setTransientCookie(w http.ResponseWriter, name, value string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(maxAge.Seconds()),
	})
}

func setAppCookie(w http.ResponseWriter, name, value string, maxAge time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   int(maxAge.Seconds()),
	})
}

func clearCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   -1,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
