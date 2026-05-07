package control

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

const (
	SessionCookieName = "ivy_webvue_session"
	OIDCStateCookie   = "ivy_webvue_oidc_state"
	OIDCNonceCookie   = "ivy_webvue_oidc_nonce"
)

type Config struct {
	Addr        string
	OIDC        OIDCConfig
	StaticIndex string
	Logger      *log.Logger
}

type Server struct {
	cfg Config
	mux *http.ServeMux
}

func NewServer(cfg Config) *Server {
	s := &Server{cfg: cfg, mux: http.NewServeMux()}
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
	_, err := r.Cookie(SessionCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		writeJSON(w, http.StatusOK, SessionView{Authenticated: false})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid session cookie"})
		return
	}
	// Full session lookup lands once the Postgres store exists.
	writeJSON(w, http.StatusOK, SessionView{Authenticated: true})
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
	// Token exchange and ID token validation belong to the next slice.
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "oidc callback exchange not implemented"})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	clearCookie(w, SessionCookieName)
	clearCookie(w, OIDCStateCookie)
	clearCookie(w, OIDCNonceCookie)
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

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
