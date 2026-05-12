package server

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/glycerine/ivy/goivy/svk/server/mail"
	"github.com/glycerine/ivy/goivy/svk/server/opaqueauth"
	"github.com/glycerine/ivy/goivy/webengine"
)

const sessionCookieName = "ivysvk_session"
const csrfCookieName = "ivysvk_csrf"

type Server struct {
	cfg            Config
	mux            *http.ServeMux
	opaque         *opaqueauth.Service
	magic          *MagicService
	oauth          *OAuthService
	passkeys       *PasskeyService
	projects       *ProjectStore
	engine         *webengine.Engine
	engineSessions *EngineSessionStore
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
		cfg:            cfg,
		mux:            http.NewServeMux(),
		opaque:         opaqueService,
		magic:          NewMagicService(sender, cfg.PublicBaseURL),
		oauth:          NewOAuthService(),
		passkeys:       passkeys,
		projects:       NewProjectStore(),
		engine:         webengine.New(nil),
		engineSessions: NewEngineSessionStore(),
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
	csrfCookie, err := r.Cookie(csrfCookieName)
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
	s.mux.HandleFunc("POST /auth/dev", s.devLogin)
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
	s.mux.HandleFunc("POST /api/engine/session", s.engineNewSession)
	s.mux.HandleFunc("POST /api/engine/session/{session}/load", s.engineLoadModel)
	s.mux.HandleFunc("POST /api/engine/session/{session}/command", s.engineRunCommand)
	s.mux.HandleFunc("GET /api/engine/session/{session}/snapshot", s.engineSnapshot)
	s.mux.HandleFunc("GET /app", s.app)
	s.mux.HandleFunc("GET /app/boot.js", s.appBoot)
	if s.cfg.StaticDir != "" {
		static := http.FileServer(http.Dir(s.cfg.StaticDir))
		s.mux.Handle("GET /_app/", static)
		s.mux.Handle("GET /wasm/", static)
		s.mux.Handle("GET /assets/", static)
		s.mux.Handle("GET /service-worker.js", static)
		s.mux.Handle("GET /robots.txt", static)
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
	w.Header().Set("content-type", "text/html; charset=utf-8")
	_ = loginTemplate.Execute(w, struct {
		DevMode bool
	}{DevMode: s.cfg.DevMode})
}

func (s *Server) devLogin(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.DevMode {
		http.NotFound(w, r)
		return
	}
	s.setSessionCookies(w, r, "user:dev-user")
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func (s *Server) authMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	if authenticated(r) {
		csrf := s.ensureCSRFCookie(w, r)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "authenticated",
			"user": map[string]string{
				"id":           "dev-user",
				"primaryEmail": "dev@example.local",
				"displayName":  "Dev User",
				"createdAt":    "2026-05-12T00:00:00.000Z",
			},
			"csrfToken": csrf,
		})
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"status":"anonymous"}`))
}

func (s *Server) app(w http.ResponseWriter, r *http.Request) {
	if !authenticated(r) {
		if s.cfg.DevMode {
			s.setSessionCookies(w, r, "user:dev-user")
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
	}

	index := filepath.Join(s.cfg.StaticDir, "index.html")
	if data, err := os.ReadFile(index); err == nil {
		w.Header().Set("content-type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
		return
	}
	if s.serveGeneratedSvelteShell(w) {
		return
	}
	w.Header().Set("content-type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html><head><title>SVK App</title></head><body><div id="svelte">SVK App</div></body></html>`))
}

func (s *Server) appBoot(w http.ResponseWriter, r *http.Request) {
	if !authenticated(r) && !s.cfg.DevMode {
		http.NotFound(w, r)
		return
	}
	start, app, _, ok := s.svelteClientEntries()
	if !ok {
		http.NotFound(w, r)
		return
	}
	hash := s.svelteGlobalHash(start)
	if hash == "" {
		hash = "app"
	}
	w.Header().Set("content-type", "text/javascript; charset=utf-8")
	var b strings.Builder
	b.WriteString(`const element = document.getElementById("svelte");`)
	b.WriteString(`globalThis.__sveltekit_`)
	b.WriteString(hash)
	b.WriteString(` = { base: "/app", assets: "" };`)
	b.WriteString(`Promise.all([import("/`)
	b.WriteString(template.JSEscapeString(start))
	b.WriteString(`"), import("/`)
	b.WriteString(template.JSEscapeString(app))
	b.WriteString(`")]).then(([kit, app]) => kit.start(app, element));`)
	_, _ = w.Write([]byte(b.String()))
}

func authenticated(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	return err == nil && cookie.Value != ""
}

func (s *Server) setSessionCookies(w http.ResponseWriter, r *http.Request, sessionID string) {
	if sessionID == "" {
		sessionID = randomURLToken()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secureCookies(r),
	})
	s.setCSRFCookie(w, r, randomURLToken())
}

func (s *Server) ensureCSRFCookie(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(csrfCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	token := randomURLToken()
	s.setCSRFCookie(w, r, token)
	return token
}

func (s *Server) setCSRFCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secureCookies(r),
	})
}

func (s *Server) secureCookies(r *http.Request) bool {
	if s.cfg.DevMode {
		return false
	}
	if r.TLS != nil {
		return true
	}
	if s.cfg.PublicBaseURL != "" {
		base, err := url.Parse(s.cfg.PublicBaseURL)
		return err == nil && base.Scheme == "https"
	}
	return true
}

type viteManifestEntry struct {
	File string   `json:"file"`
	CSS  []string `json:"css"`
}

func (s *Server) serveGeneratedSvelteShell(w http.ResponseWriter) bool {
	_, _, css, ok := s.svelteClientEntries()
	if !ok {
		return false
	}
	w.Header().Set("content-type", "text/html; charset=utf-8")
	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8">`)
	b.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1">`)
	b.WriteString(`<title>SVK</title>`)
	for _, href := range css {
		b.WriteString(`<link rel="stylesheet" href="/`)
		b.WriteString(template.HTMLEscapeString(href))
		b.WriteString(`">`)
	}
	b.WriteString(`</head><body><div id="svelte"><script type="module" src="/app/boot.js"></script></div></body></html>`)
	_, _ = w.Write([]byte(b.String()))
	return true
}

func (s *Server) svelteClientEntries() (start string, app string, css []string, ok bool) {
	manifestPath := filepath.Join(s.cfg.StaticDir, ".vite", "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", "", nil, false
	}
	var manifest map[string]viteManifestEntry
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", "", nil, false
	}
	for key, entry := range manifest {
		switch {
		case strings.HasSuffix(key, "/entry.js"):
			start = entry.File
		case strings.HasSuffix(key, "/app.js"):
			app = entry.File
		case strings.HasSuffix(key, "/nodes/2.js"):
			css = append(css, entry.CSS...)
		}
	}
	return start, app, css, start != "" && app != ""
}

func (s *Server) svelteGlobalHash(start string) string {
	data, err := os.ReadFile(filepath.Join(s.cfg.StaticDir, start))
	if err != nil {
		return ""
	}
	importRe := regexp.MustCompile(`from\s*"\.\./chunks/([^"]+)"`)
	matches := importRe.FindSubmatch(data)
	if len(matches) != 2 {
		return ""
	}
	chunk, err := os.ReadFile(filepath.Join(s.cfg.StaticDir, "_app", "immutable", "chunks", string(matches[1])))
	if err != nil {
		return ""
	}
	hashRe := regexp.MustCompile(`__sveltekit_([A-Za-z0-9_]+)`)
	hashMatches := hashRe.FindSubmatch(chunk)
	if len(hashMatches) != 2 {
		return ""
	}
	return string(hashMatches[1])
}

func securityHeaders(w http.ResponseWriter) {
	w.Header().Set("x-content-type-options", "nosniff")
	w.Header().Set("x-frame-options", "DENY")
	w.Header().Set("referrer-policy", "same-origin")
	w.Header().Set("content-security-policy", "default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self'; style-src-elem 'self'; style-src-attr 'unsafe-inline'; connect-src 'self'; worker-src 'self' blob:; img-src 'self' data:; form-action 'self'; base-uri 'self'; frame-ancestors 'none'")
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

var loginTemplate = template.Must(template.New("login").Parse(`<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Login</title>
</head>
<body>
	<main>
		<h1>Login</h1>
		<section aria-labelledby="signup-heading">
			<h2 id="signup-heading">Create account</h2>
			<p>Use OAuth, passkeys, OPAQUE password login, or a magic email link. First login creates your personal account.</p>
			<nav aria-label="OAuth signup">
				<a href="/auth/oauth/google/start">Continue with Google</a>
				<a href="/auth/oauth/apple/start">Continue with Apple</a>
				<a href="/auth/oauth/github/start">Continue with GitHub</a>
			</nav>
			<form method="post" action="/auth/magic/request">
				<label>Email <input name="email" type="email" autocomplete="email" required></label>
				<input type="hidden" name="purpose" value="signup">
				<button type="submit">Send magic link</button>
			</form>
			<form method="post" action="/auth/opaque/register/start">
				<label>Email <input name="userId" type="email" autocomplete="email" required></label>
				<button type="submit">Create OPAQUE password account</button>
			</form>
		</section>
		<section aria-labelledby="signin-heading">
			<h2 id="signin-heading">Sign in</h2>
			<form method="post" action="/auth/opaque/login/start">
				<label>Email <input name="userId" type="email" autocomplete="email" required></label>
				<button type="submit">Sign in with OPAQUE password</button>
			</form>
			<form method="post" action="/auth/passkey/login/options">
				<label>Email <input name="userId" type="email" autocomplete="username webauthn" required></label>
				<button type="submit">Sign in with passkey</button>
			</form>
		</section>
		{{if .DevMode}}
		<section aria-labelledby="dev-heading">
			<h2 id="dev-heading">Local development</h2>
			<form method="post" action="/auth/dev">
				<button type="submit">Continue as local dev user</button>
			</form>
		</section>
		{{end}}
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
