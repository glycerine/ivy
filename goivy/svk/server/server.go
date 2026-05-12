package server

import (
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/glycerine/ivy/goivy/svk/server/mail"
	"github.com/glycerine/ivy/goivy/svk/server/opaqueauth"
)

const sessionCookieName = "ivysvk_session"

type Server struct {
	cfg      Config
	mux      *http.ServeMux
	opaque   *opaqueauth.Service
	magic    *MagicService
	oauth    *OAuthService
	passkeys *PasskeyService
	projects *ProjectStore
}

func New(cfg Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	opaqueService, err := opaqueauth.NewService([]byte("ivysvk"))
	if err != nil {
		return nil, err
	}
	sender := mail.Sender(mail.NoopSender{})
	if cfg.MailgunDomain != "" && cfg.MailgunAPIKey != "" {
		sender = mail.NewMailgunSender(cfg.MailgunDomain, cfg.MailgunAPIKey, "SVK <login@"+cfg.MailgunDomain+">")
	}
	passkeys, err := NewPasskeyService("localhost", envDefault("IVYSVK_PUBLIC_BASE_URL", "http://localhost:8080"))
	if err != nil {
		return nil, err
	}
	s := &Server{
		cfg:      cfg,
		mux:      http.NewServeMux(),
		opaque:   opaqueService,
		magic:    NewMagicService(sender, cfg.PublicBaseURL),
		oauth:    NewOAuthService(),
		passkeys: passkeys,
		projects: NewProjectStore(),
	}
	s.routes()
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	securityHeaders(w)
	if !s.validStateChangingRequest(w, r) {
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) validStateChangingRequest(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return true
	}
	if len(r.URL.Path) < 5 || r.URL.Path[:5] != "/api/" {
		return true
	}
	if origin := r.Header.Get("origin"); origin != "" && s.cfg.PublicBaseURL != "" {
		base, err := url.Parse(s.cfg.PublicBaseURL)
		if err != nil || origin != base.Scheme+"://"+base.Host {
			http.Error(w, "origin mismatch", http.StatusForbidden)
			return false
		}
	}
	csrfCookie, err := r.Cookie("ivysvk_csrf")
	if err != nil || csrfCookie.Value == "" || r.Header.Get("x-csrf-token") != csrfCookie.Value {
		http.Error(w, "csrf token required", http.StatusForbidden)
		return false
	}
	return true
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.health)
	s.mux.HandleFunc("GET /", s.marketing)
	s.mux.HandleFunc("GET /login", s.login)
	s.mux.HandleFunc("GET /auth/me", s.authMe)
	s.mux.HandleFunc("POST /auth/opaque/register/start", s.opaqueRegistrationStart)
	s.mux.HandleFunc("POST /auth/opaque/register/finish", s.opaqueRegistrationFinish)
	s.mux.HandleFunc("POST /auth/opaque/login/start", s.opaqueLoginStart)
	s.mux.HandleFunc("POST /auth/opaque/login/finish", s.opaqueLoginFinish)
	s.mux.HandleFunc("POST /auth/magic/request", s.magicRequest)
	s.mux.HandleFunc("GET /auth/magic/consume", s.magicConsume)
	s.mux.HandleFunc("GET /auth/oauth/{provider}/start", s.oauthStart)
	s.mux.HandleFunc("GET /auth/oauth/{provider}/callback", s.oauthCallback)
	s.mux.HandleFunc("POST /auth/passkey/register/options", s.passkeyRegistrationOptions)
	s.mux.HandleFunc("POST /auth/passkey/register/finish", s.passkeyRegistrationFinish)
	s.mux.HandleFunc("POST /auth/passkey/login/options", s.passkeyLoginOptions)
	s.mux.HandleFunc("POST /auth/passkey/login/finish", s.passkeyLoginFinish)
	s.mux.HandleFunc("GET /api/projects", s.listProjects)
	s.mux.HandleFunc("POST /api/projects", s.createProject)
	s.mux.HandleFunc("POST /api/projects/{project}/models", s.saveProjectModel)
	s.mux.HandleFunc("GET /app", s.app)
	if s.cfg.StaticDir != "" {
		assets := http.StripPrefix("/assets/", http.FileServer(http.Dir(filepath.Join(s.cfg.StaticDir, "assets"))))
		s.mux.Handle("GET /assets/", assets)
	}
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) marketing(w http.ResponseWriter, _ *http.Request) {
	renderPage(w, "SVK Ivy", "Verify Ivy systems locally or with a hosted solver.")
}

func (s *Server) login(w http.ResponseWriter, _ *http.Request) {
	renderPage(w, "Login", "Sign in with OAuth, OPAQUE, passkeys, or a recovery email link.")
}

func (s *Server) authMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	if authenticated(r) {
		_, _ = w.Write([]byte(`{"status":"authenticated","user":{"id":"dev-user","primaryEmail":"dev@example.local","displayName":"Dev User","createdAt":"2026-05-12T00:00:00.000Z"}}`))
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"status":"anonymous"}`))
}

func (s *Server) app(w http.ResponseWriter, r *http.Request) {
	if !authenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	index := filepath.Join(s.cfg.StaticDir, "index.html")
	if data, err := os.ReadFile(index); err == nil {
		w.Header().Set("content-type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
		return
	}
	w.Header().Set("content-type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html><head><title>SVK App</title></head><body><div id="svelte">SVK App</div></body></html>`))
}

func authenticated(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	return err == nil && cookie.Value != ""
}

func securityHeaders(w http.ResponseWriter) {
	w.Header().Set("x-content-type-options", "nosniff")
	w.Header().Set("x-frame-options", "DENY")
	w.Header().Set("referrer-policy", "same-origin")
	w.Header().Set("content-security-policy", "default-src 'self'; base-uri 'self'; frame-ancestors 'none'")
}

var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>{{.Title}}</title>
</head>
<body>
	<main>
		<h1>{{.Title}}</h1>
		<p>{{.Body}}</p>
	</main>
</body>
</html>`))

func renderPage(w http.ResponseWriter, title, body string) {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	_ = pageTemplate.Execute(w, struct {
		Title string
		Body  string
	}{Title: title, Body: body})
}
