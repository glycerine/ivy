package control

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
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
	OIDC                      OIDCConfig
	OIDCProvider              OIDCProvider
	Store                     Store
	StaticIndex               string
	Logger                    *log.Logger
	CookieSecure              bool
	AutoProvisionStarterSpace bool
	EnableTestIDP             bool
}

type Server struct {
	cfg     Config
	mux     *http.ServeMux
	store   Store
	oidc    OIDCProvider
	testIDP *TestIDP
}

func NewServer(cfg Config) *Server {
	if cfg.Store == nil {
		cfg.Store = NewMemoryStore()
	}
	if cfg.OIDCProvider == nil {
		cfg.OIDCProvider = NewHTTPOIDCProvider(cfg.OIDC, nil)
	}
	s := &Server{cfg: cfg, mux: http.NewServeMux(), store: cfg.Store, oidc: cfg.OIDCProvider}
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
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /auth/me", s.handleAuthMe)
	s.mux.HandleFunc("GET /auth/login", s.handleLogin)
	s.mux.HandleFunc("GET /auth/callback", s.handleCallback)
	s.mux.HandleFunc("POST /auth/logout", s.handleLogout)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
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
	view, err := s.store.SessionViewByToken(r.Context(), cookie.Value, time.Now().UTC())
	if errors.Is(err, ErrSessionNotFound) {
		clearCookie(w, SessionCookieName, s.cfg.CookieSecure)
		writeJSON(w, http.StatusOK, SessionView{Authenticated: false})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load session"})
		return
	}
	writeJSON(w, http.StatusOK, view)
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
	if err := s.store.CreateAppSession(r.Context(), user.ID, sessionToken, csrfToken, now, 24*time.Hour, 30*24*time.Hour); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to store app session"})
		return
	}
	setAppCookie(w, SessionCookieName, sessionToken, 30*24*time.Hour, s.cfg.CookieSecure)
	clearCookie(w, OIDCStateCookie, s.cfg.CookieSecure)
	clearCookie(w, OIDCNonceCookie, s.cfg.CookieSecure)
	http.Redirect(w, r, "/", http.StatusFound)
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
