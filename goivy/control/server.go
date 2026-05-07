package control

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	SessionCookieName = "ivy_webvue_session"
	OIDCStateCookie   = "ivy_webvue_oidc_state"
	OIDCNonceCookie   = "ivy_webvue_oidc_nonce"
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
		cfg.Store = NewMemoryStore()
	}
	if cfg.OIDCProvider == nil {
		cfg.OIDCProvider = NewHTTPOIDCProvider(cfg.OIDC, nil)
	}
	if cfg.EmailSender == nil {
		cfg.EmailSender = NewMemoryEmailSender()
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
	return s.mux
}

func (s *Server) Start() error {
	addr := s.cfg.Addr
	if addr == "" {
		addr = "127.0.0.1:18080"
	}
	return http.ListenAndServe(addr, s.Handler())
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
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /auth/me", s.handleAuthMe)
	s.mux.HandleFunc("POST /auth/email/request", s.handleEmailLoginRequest)
	s.mux.HandleFunc("GET /auth/email/continue", s.handleEmailContinue)
	s.mux.HandleFunc("POST /auth/email/consume", s.handleEmailLoginConsume)
	s.mux.HandleFunc("GET /auth/login", s.handleLogin)
	s.mux.HandleFunc("GET /auth/callback", s.handleCallback)
	s.mux.HandleFunc("POST /auth/logout", s.handleLogout)
	if s.cfg.EnableTestEmailOutbox {
		s.mux.HandleFunc("GET /test/email/latest", s.handleTestEmailLatest)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
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
	setAppCookie(w, SessionCookieName, cookie.Value, AppSessionTTL, s.cfg.CookieSecure)
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleEmailLoginRequest(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
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
	if err := service.RequestLogin(r.Context(), payload.Email, time.Now().UTC()); err != nil {
		// Keep the browser response neutral; details can go to logs once we add structured logging.
		if s.cfg.Logger != nil {
			s.cfg.Logger.Printf("email login request failed: %v", err)
		}
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleEmailContinue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html>
<title>Ivy sign in</title>
<main id="status">Signing in...</main>
<script>
(async () => {
  const params = new URLSearchParams(location.hash.slice(1));
  const token = params.get("token");
  if (!token) {
    document.getElementById("status").textContent = "Sign-in link is missing its token.";
    return;
  }
  const response = await fetch("/auth/email/consume", {
    method: "POST",
    headers: {"content-type": "application/json"},
    body: JSON.stringify({token})
  });
  if (response.ok) {
    location.replace("/");
    return;
  }
  document.getElementById("status").textContent = "Sign-in link is invalid or expired.";
})();
</script>`))
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
	sessionToken, _, _, err := service.ConsumeLogin(r.Context(), payload.Token, time.Now().UTC())
	if errors.Is(err, ErrEmailLoginTokenNotFound) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired login token"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create app session"})
		return
	}
	setAppCookie(w, SessionCookieName, sessionToken, AppSessionTTL, s.cfg.CookieSecure)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
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
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "oidc exchange failed"})
		return
	}
	user, err := s.store.UpsertUserFromOIDC(r.Context(), identity)
	if err != nil {
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to store app session"})
		return
	}
	setAppCookie(w, SessionCookieName, sessionToken, AppSessionTTL, s.cfg.CookieSecure)
	clearCookie(w, OIDCStateCookie, s.cfg.CookieSecure)
	clearCookie(w, OIDCNonceCookie, s.cfg.CookieSecure)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleTestEmailLatest(w http.ResponseWriter, r *http.Request) {
	sender, ok := s.email.(*MemoryEmailSender)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "test email outbox is not available"})
		return
	}
	email, err := NormalizeEmail(r.URL.Query().Get("email"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid email query is required"})
		return
	}
	message, ok := sender.LatestFor(email)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no message for email"})
		return
	}
	writeJSON(w, http.StatusOK, message)
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
	clearCookie(w, SessionCookieName, s.cfg.CookieSecure)
	clearCookie(w, OIDCStateCookie, s.cfg.CookieSecure)
	clearCookie(w, OIDCNonceCookie, s.cfg.CookieSecure)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
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
