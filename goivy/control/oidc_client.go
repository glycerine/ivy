package control

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OIDCProvider interface {
	ExchangeCode(ctx context.Context, code, nonce string) (OIDCIdentity, error)
}

type HTTPOIDCProvider struct {
	cfg        OIDCConfig
	httpClient *http.Client
	now        func() time.Time
}

func NewHTTPOIDCProvider(cfg OIDCConfig, httpClient *http.Client) *HTTPOIDCProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &HTTPOIDCProvider{
		cfg:        cfg,
		httpClient: httpClient,
		now:        time.Now,
	}
}

func (p *HTTPOIDCProvider) ExchangeCode(ctx context.Context, code, nonce string) (OIDCIdentity, error) {
	if strings.TrimSpace(code) == "" {
		return OIDCIdentity{}, errors.New("authorization code is required")
	}
	cfg, err := p.discoverIfNeeded(ctx)
	if err != nil {
		return OIDCIdentity{}, err
	}
	token, err := p.exchangeToken(ctx, cfg, code)
	if err != nil {
		return OIDCIdentity{}, err
	}
	claims, err := p.validateIDToken(ctx, cfg, token.IDToken, nonce)
	if err != nil {
		return OIDCIdentity{}, err
	}
	identity := OIDCIdentity{
		Issuer:        claims.Issuer,
		Subject:       claims.Subject,
		Email:         claims.Email,
		DisplayName:   firstNonEmpty(claims.Name, claims.PreferredUsername, claims.Email),
		EmailVerified: claims.EmailVerified,
	}
	if identity.Email == "" && cfg.UserInfoURL != "" && token.AccessToken != "" {
		if info, err := p.fetchUserInfo(ctx, cfg.UserInfoURL, token.AccessToken); err == nil {
			identity.Email = firstNonEmpty(info.Email, identity.Email)
			identity.DisplayName = firstNonEmpty(identity.DisplayName, info.Name, info.PreferredUsername, identity.Email)
			identity.EmailVerified = identity.EmailVerified || info.EmailVerified
		}
	}
	return identity, identity.Validate()
}

func (p *HTTPOIDCProvider) discoverIfNeeded(ctx context.Context) (OIDCConfig, error) {
	cfg := p.cfg
	if cfg.TokenURL != "" && cfg.JWKSURL != "" {
		return cfg, nil
	}
	if cfg.IssuerURL == "" {
		return OIDCConfig{}, errors.New("oidc token and jwks urls require issuer discovery when omitted")
	}
	discoveryURL := strings.TrimRight(cfg.IssuerURL, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return OIDCConfig{}, err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return OIDCConfig{}, fmt.Errorf("discover oidc provider: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return OIDCConfig{}, fmt.Errorf("discover oidc provider: status %d: %s", resp.StatusCode, string(body))
	}
	var doc struct {
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		JWKSURI               string `json:"jwks_uri"`
		UserInfoEndpoint      string `json:"userinfo_endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return OIDCConfig{}, fmt.Errorf("decode oidc discovery: %w", err)
	}
	if cfg.AuthURL == "" {
		cfg.AuthURL = doc.AuthorizationEndpoint
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = doc.TokenEndpoint
	}
	if cfg.JWKSURL == "" {
		cfg.JWKSURL = doc.JWKSURI
	}
	if cfg.UserInfoURL == "" {
		cfg.UserInfoURL = doc.UserInfoEndpoint
	}
	return cfg, nil
}

func (p *HTTPOIDCProvider) exchangeToken(ctx context.Context, cfg OIDCConfig, code string) (oidcTokenResponse, error) {
	if strings.TrimSpace(cfg.TokenURL) == "" {
		return oidcTokenResponse{}, errors.New("oidc token URL is required")
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", cfg.RedirectURL)
	form.Set("client_id", cfg.ClientID)
	if cfg.ClientSecret != "" {
		form.Set("client_secret", cfg.ClientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return oidcTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return oidcTokenResponse{}, fmt.Errorf("exchange oidc token: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return oidcTokenResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return oidcTokenResponse{}, fmt.Errorf("exchange oidc token: status %d: %s", resp.StatusCode, string(body))
	}
	var token oidcTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return oidcTokenResponse{}, fmt.Errorf("decode oidc token response: %w", err)
	}
	if token.IDToken == "" {
		return oidcTokenResponse{}, errors.New("oidc token response missing id_token")
	}
	return token, nil
}

func (p *HTTPOIDCProvider) validateIDToken(ctx context.Context, cfg OIDCConfig, rawIDToken, nonce string) (oidcClaims, error) {
	header, claims, signingInput, signature, err := parseJWT(rawIDToken)
	if err != nil {
		return oidcClaims{}, err
	}
	if header.Algorithm != "RS256" {
		return oidcClaims{}, fmt.Errorf("unsupported id_token algorithm %q", header.Algorithm)
	}
	key, err := p.fetchRSAKey(ctx, cfg.JWKSURL, header.KeyID)
	if err != nil {
		return oidcClaims{}, err
	}
	sum := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, sum[:], signature); err != nil {
		return oidcClaims{}, fmt.Errorf("verify id_token signature: %w", err)
	}
	now := p.now().Unix()
	if claims.ExpiresAt <= now {
		return oidcClaims{}, errors.New("id_token is expired")
	}
	if cfg.IssuerURL != "" && strings.TrimRight(claims.Issuer, "/") != strings.TrimRight(cfg.IssuerURL, "/") {
		return oidcClaims{}, fmt.Errorf("id_token issuer %q does not match %q", claims.Issuer, cfg.IssuerURL)
	}
	if claims.Subject == "" {
		return oidcClaims{}, errors.New("id_token subject is empty")
	}
	if !claims.Audience.Contains(cfg.ClientID) {
		return oidcClaims{}, fmt.Errorf("id_token audience does not include client id %q", cfg.ClientID)
	}
	if nonce != "" && claims.Nonce != nonce {
		return oidcClaims{}, errors.New("id_token nonce mismatch")
	}
	return claims, nil
}

func (p *HTTPOIDCProvider) fetchRSAKey(ctx context.Context, jwksURL, kid string) (*rsa.PublicKey, error) {
	if jwksURL == "" {
		return nil, errors.New("oidc jwks URL is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("fetch jwks: status %d: %s", resp.StatusCode, string(body))
	}
	var set jwks
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}
	for _, key := range set.Keys {
		if kid != "" && key.KeyID != kid {
			continue
		}
		if key.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			return nil, fmt.Errorf("decode jwk modulus: %w", err)
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			return nil, fmt.Errorf("decode jwk exponent: %w", err)
		}
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}
		if e == 0 {
			return nil, errors.New("jwk exponent is zero")
		}
		return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
	}
	return nil, fmt.Errorf("no RSA jwk found for kid %q", kid)
}

func (p *HTTPOIDCProvider) fetchUserInfo(ctx context.Context, userInfoURL, accessToken string) (oidcUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return oidcUserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return oidcUserInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return oidcUserInfo{}, fmt.Errorf("userinfo status %d: %s", resp.StatusCode, string(body))
	}
	var info oidcUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return oidcUserInfo{}, err
	}
	return info, nil
}

type oidcTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Type      string `json:"typ"`
}

type oidcClaims struct {
	Issuer            string       `json:"iss"`
	Subject           string       `json:"sub"`
	Audience          oidcAudience `json:"aud"`
	ExpiresAt         int64        `json:"exp"`
	IssuedAt          int64        `json:"iat"`
	Nonce             string       `json:"nonce"`
	Email             string       `json:"email"`
	EmailVerified     bool         `json:"email_verified"`
	Name              string       `json:"name"`
	PreferredUsername string       `json:"preferred_username"`
}

type oidcUserInfo struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
}

type oidcAudience []string

func (a *oidcAudience) UnmarshalJSON(data []byte) error {
	var one string
	if err := json.Unmarshal(data, &one); err == nil {
		*a = []string{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

func (a oidcAudience) Contains(value string) bool {
	for _, candidate := range a {
		if candidate == value {
			return true
		}
	}
	return false
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty   string `json:"kty"`
	KeyID string `json:"kid"`
	Use   string `json:"use"`
	Alg   string `json:"alg"`
	N     string `json:"n"`
	E     string `json:"e"`
}

func parseJWT(raw string) (jwtHeader, oidcClaims, string, []byte, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return jwtHeader{}, oidcClaims{}, "", nil, errors.New("id_token is not a compact jwt")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return jwtHeader{}, oidcClaims{}, "", nil, fmt.Errorf("decode jwt header: %w", err)
	}
	claimBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return jwtHeader{}, oidcClaims{}, "", nil, fmt.Errorf("decode jwt claims: %w", err)
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return jwtHeader{}, oidcClaims{}, "", nil, fmt.Errorf("decode jwt signature: %w", err)
	}
	var header jwtHeader
	if err := json.NewDecoder(bytes.NewReader(headerBytes)).Decode(&header); err != nil {
		return jwtHeader{}, oidcClaims{}, "", nil, fmt.Errorf("decode jwt header json: %w", err)
	}
	var claims oidcClaims
	if err := json.NewDecoder(bytes.NewReader(claimBytes)).Decode(&claims); err != nil {
		return jwtHeader{}, oidcClaims{}, "", nil, fmt.Errorf("decode jwt claims json: %w", err)
	}
	return header, claims, parts[0] + "." + parts[1], signature, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
