package control

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
)

type OIDCConfig struct {
	IssuerURL    string
	AuthURL      string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

func (c OIDCConfig) Validate() error {
	if strings.TrimSpace(c.AuthURL) == "" {
		return errors.New("auth URL is required")
	}
	if strings.TrimSpace(c.ClientID) == "" {
		return errors.New("client ID is required")
	}
	if strings.TrimSpace(c.RedirectURL) == "" {
		return errors.New("redirect URL is required")
	}
	return nil
}

func (c OIDCConfig) LoginURL(state, nonce string) (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	authURL, err := url.Parse(c.AuthURL)
	if err != nil {
		return "", err
	}
	scopes := c.Scopes
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}
	q := authURL.Query()
	q.Set("response_type", "code")
	q.Set("client_id", c.ClientID)
	q.Set("redirect_uri", c.RedirectURL)
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("state", state)
	q.Set("nonce", nonce)
	authURL.RawQuery = q.Encode()
	return authURL.String(), nil
}

func RandomToken(byteCount int) (string, error) {
	if byteCount <= 0 {
		return "", errors.New("byte count must be positive")
	}
	buf := make([]byte, byteCount)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
