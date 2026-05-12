package server

import (
	"errors"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	Addr          string
	PublicBaseURL string
	DatabaseDSN   string
	MailgunDomain string
	MailgunAPIKey string
	CookieSecret  string
	StaticDir     string
	DevMode       bool
}

func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		Addr:          envDefault("IVYSVK_ADDR", "127.0.0.1:8080"),
		PublicBaseURL: os.Getenv("IVYSVK_PUBLIC_BASE_URL"),
		DatabaseDSN:   os.Getenv("IVYSVK_DB_DSN"),
		MailgunDomain: os.Getenv("IVYSVK_MAILGUN_DOMAIN"),
		MailgunAPIKey: os.Getenv("IVYSVK_MAILGUN_API_KEY"),
		CookieSecret:  os.Getenv("IVYSVK_COOKIE_SECRET"),
		StaticDir:     envDefault("IVYSVK_STATIC_DIR", "svk/build"),
	}
	cfg.DevMode, _ = strconv.ParseBool(envDefault("IVYSVK_DEV", "false"))
	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	if c.PublicBaseURL != "" {
		if _, err := url.ParseRequestURI(c.PublicBaseURL); err != nil {
			return err
		}
	}
	if c.DevMode {
		return nil
	}
	if c.PublicBaseURL == "" {
		return errors.New("IVYSVK_PUBLIC_BASE_URL is required outside development")
	}
	if c.DatabaseDSN == "" {
		return errors.New("IVYSVK_DB_DSN is required outside development")
	}
	if c.CookieSecret == "" {
		return errors.New("IVYSVK_COOKIE_SECRET is required outside development")
	}
	if c.MailgunDomain == "" || c.MailgunAPIKey == "" {
		return errors.New("Mailgun domain and API key are required outside development")
	}
	return nil
}

func envDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
