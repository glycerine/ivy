package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/glycerine/ivy/goivy/control"
	"github.com/glycerine/ivy/goivy/webvue"
)

func main() {
	addr := flag.String("addr", envDefault("IVY_CONTROL_ADDR", "127.0.0.1:18080"), "listen address")
	publicBaseURL := flag.String("public-base-url", os.Getenv("IVY_CONTROL_PUBLIC_BASE_URL"), "public base URL used in emailed login links")
	staticDir := flag.String("static-dir", os.Getenv("IVY_CONTROL_STATIC_DIR"), "serve browser assets from this directory instead of materializing embedded webvue assets")
	runwebDir := flag.String("runweb-dir", envDefault("IVY_CONTROL_RUNWEB_DIR", webvue.DefaultRunwebDir), "directory to receive embedded browser assets when static-dir is not set")
	issuerURL := flag.String("oidc-issuer-url", envDefault("IVY_CONTROL_OIDC_ISSUER_URL", "http://127.0.0.1:18082"), "OIDC issuer URL")
	authURL := flag.String("oidc-auth-url", envDefault("IVY_CONTROL_OIDC_AUTH_URL", "http://127.0.0.1:18082/login/oauth/authorize"), "OIDC authorization endpoint URL")
	tokenURL := flag.String("oidc-token-url", os.Getenv("IVY_CONTROL_OIDC_TOKEN_URL"), "OIDC token endpoint URL")
	jwksURL := flag.String("oidc-jwks-url", os.Getenv("IVY_CONTROL_OIDC_JWKS_URL"), "OIDC JWKS URL")
	userInfoURL := flag.String("oidc-userinfo-url", os.Getenv("IVY_CONTROL_OIDC_USERINFO_URL"), "OIDC userinfo endpoint URL")
	clientID := flag.String("oidc-client-id", envDefault("IVY_CONTROL_OIDC_CLIENT_ID", "ivy-control-local"), "OIDC client id")
	clientSecret := flag.String("oidc-client-secret", os.Getenv("IVY_CONTROL_OIDC_CLIENT_SECRET"), "OIDC client secret")
	redirectURL := flag.String("oidc-redirect-url", envDefault("IVY_CONTROL_OIDC_REDIRECT_URL", "http://127.0.0.1:18080/auth/callback"), "OIDC redirect URL")
	databaseDSN := flag.String("database-dsn", os.Getenv("IVY_CONTROL_DATABASE_DSN"), "PostgreSQL database/sql DSN for ivyvue control tables")
	cookieSecure := flag.Bool("cookie-secure", envBool("IVY_CONTROL_COOKIE_SECURE", false), "set Secure on app session cookies")
	testIDP := flag.Bool("test-idp", envBool("IVY_CONTROL_TEST_IDP", false), "enable the in-process deterministic OIDC issuer for local tests only")
	testEmailOutbox := flag.Bool("test-email-outbox", envBool("IVY_CONTROL_TEST_EMAIL_OUTBOX", false), "enable test-only endpoint for the in-memory email outbox")
	mailgunDomain := flag.String("mailgun-domain", envDefault("IVY_CONTROL_MAILGUN_DOMAIN", "mg.fencebunt.com"), "Mailgun sending domain")
	mailgunFrom := flag.String("mailgun-from", envDefault("IVY_CONTROL_MAILGUN_FROM", "Ivy <postmaster@mg.fencebunt.com>"), "Mailgun From header")
	flag.Parse()

	var store control.Store
	var closer interface{ Close() error }
	if *databaseDSN != "" {
		pg, err := control.OpenPostgresStore(*databaseDSN)
		if err != nil {
			log.Fatalf("open control database: %v", err)
		}
		store = pg
		closer = pg
		defer closer.Close()
	} else {
		store = control.NewMemoryStore()
	}
	if *testIDP {
		base := "http://" + *addr + "/test-idp"
		*issuerURL = base
		*authURL = base + "/login/oauth/authorize"
		*tokenURL = base + "/api/login/oauth/access_token"
		*jwksURL = base + "/jwks"
		*userInfoURL = base + "/api/userinfo"
		*redirectURL = "http://" + *addr + "/auth/callback"
	}
	emailSender := control.EmailSender(control.NewMemoryEmailSender())
	if apiKey := envDefault("IVY_CONTROL_MAILGUN_API_KEY", os.Getenv("MAILGIN_FENCEBUNT_SIGNUP_API_KEY")); apiKey != "" {
		emailSender = control.MailgunEmailSender{
			Domain: *mailgunDomain,
			APIKey: apiKey,
			From:   *mailgunFrom,
		}
	}
	serveStaticDir, err := prepareStaticDir(*staticDir, *runwebDir)
	if err != nil {
		log.Fatalf("prepare browser assets: %v", err)
	}
	srv := control.NewServer(control.Config{
		Addr:          *addr,
		PublicBaseURL: *publicBaseURL,
		StaticDir:     serveStaticDir,
		OIDC: control.OIDCConfig{
			IssuerURL:    *issuerURL,
			AuthURL:      *authURL,
			TokenURL:     *tokenURL,
			JWKSURL:      *jwksURL,
			UserInfoURL:  *userInfoURL,
			ClientID:     *clientID,
			ClientSecret: *clientSecret,
			RedirectURL:  *redirectURL,
			Scopes:       []string{"openid", "profile", "email"},
		},
		Store:                     store,
		EmailSender:               emailSender,
		Logger:                    log.Default(),
		CookieSecure:              *cookieSecure,
		AutoProvisionStarterSpace: true,
		EnableTestIDP:             *testIDP,
		EnableTestEmailOutbox:     *testEmailOutbox,
	})
	log.Printf("ivy control-plane listening on http://%s", *addr)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}

func prepareStaticDir(staticDir, runwebDir string) (string, error) {
	if strings.TrimSpace(staticDir) != "" {
		return staticDir, nil
	}
	if strings.TrimSpace(runwebDir) == "" {
		runwebDir = webvue.DefaultRunwebDir
	}
	if err := webvue.MaterializeRunwebDir(runwebDir); err != nil {
		return "", err
	}
	return runwebDir, nil
}

func envDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func envBool(name string, fallback bool) bool {
	switch os.Getenv(name) {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}
