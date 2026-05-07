package main

import (
	"flag"
	"log"
	"os"

	"github.com/glycerine/ivy/goivy/control"
)

func main() {
	addr := flag.String("addr", envDefault("IVY_CONTROL_ADDR", "127.0.0.1:18080"), "listen address")
	issuerURL := flag.String("oidc-issuer-url", envDefault("IVY_CONTROL_OIDC_ISSUER_URL", "http://127.0.0.1:18082"), "OIDC issuer URL")
	authURL := flag.String("oidc-auth-url", envDefault("IVY_CONTROL_OIDC_AUTH_URL", "http://127.0.0.1:18082/login/oauth/authorize"), "OIDC authorization endpoint URL")
	clientID := flag.String("oidc-client-id", envDefault("IVY_CONTROL_OIDC_CLIENT_ID", "ivy-control-local"), "OIDC client id")
	clientSecret := flag.String("oidc-client-secret", os.Getenv("IVY_CONTROL_OIDC_CLIENT_SECRET"), "OIDC client secret")
	redirectURL := flag.String("oidc-redirect-url", envDefault("IVY_CONTROL_OIDC_REDIRECT_URL", "http://127.0.0.1:18080/auth/callback"), "OIDC redirect URL")
	flag.Parse()

	srv := control.NewServer(control.Config{
		Addr: *addr,
		OIDC: control.OIDCConfig{
			IssuerURL:    *issuerURL,
			AuthURL:      *authURL,
			ClientID:     *clientID,
			ClientSecret: *clientSecret,
			RedirectURL:  *redirectURL,
			Scopes:       []string{"openid", "profile", "email"},
		},
		Logger: log.Default(),
	})
	log.Printf("ivy control-plane listening on http://%s", *addr)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}

func envDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
