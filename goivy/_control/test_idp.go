package control

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type TestIDP struct {
	issuer   string
	clientID string
	key      *rsa.PrivateKey
	kid      string
	mu       sync.Mutex
	codes    map[string]string
}

func NewTestIDP(issuer, clientID string) *TestIDP {
	if issuer == "" {
		issuer = "http://127.0.0.1:18080/test-idp"
	}
	if clientID == "" {
		clientID = "ivy-control-local"
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return &TestIDP{
		issuer:   strings.TrimRight(issuer, "/"),
		clientID: clientID,
		key:      key,
		kid:      "ivy-control-test-idp",
		codes:    map[string]string{},
	}
}

func (p *TestIDP) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /test-idp/.well-known/openid-configuration", p.handleDiscovery)
	mux.HandleFunc("GET /test-idp/login/oauth/authorize", p.handleAuthorize)
	mux.HandleFunc("POST /test-idp/api/login/oauth/access_token", p.handleToken)
	mux.HandleFunc("GET /test-idp/jwks", p.handleJWKS)
	mux.HandleFunc("GET /test-idp/api/userinfo", p.handleUserInfo)
}

func (p *TestIDP) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"issuer":                 p.issuer,
		"authorization_endpoint": p.issuer + "/login/oauth/authorize",
		"token_endpoint":         p.issuer + "/api/login/oauth/access_token",
		"jwks_uri":               p.issuer + "/jwks",
		"userinfo_endpoint":      p.issuer + "/api/userinfo",
	})
}

func (p *TestIDP) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	nonce := r.URL.Query().Get("nonce")
	if redirectURI == "" || state == "" || nonce == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing authorize parameter"})
		return
	}
	code, err := RandomToken(24)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create code"})
		return
	}
	p.mu.Lock()
	p.codes[code] = nonce
	p.mu.Unlock()
	callback, err := url.Parse(redirectURI)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid redirect_uri"})
		return
	}
	q := callback.Query()
	q.Set("code", code)
	q.Set("state", state)
	callback.RawQuery = q.Encode()
	http.Redirect(w, r, callback.String(), http.StatusFound)
}

func (p *TestIDP) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid form"})
		return
	}
	code := r.Form.Get("code")
	p.mu.Lock()
	nonce, ok := p.codes[code]
	delete(p.codes, code)
	p.mu.Unlock()
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid code"})
		return
	}
	token, err := p.signIDToken(nonce)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to sign token"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": "test-access-token",
		"id_token":     token,
		"token_type":   "Bearer",
		"expires_in":   300,
	})
}

func (p *TestIDP) handleJWKS(w http.ResponseWriter, r *http.Request) {
	pub := p.key.PublicKey
	eBytes := bigEndianInt(pub.E)
	writeJSON(w, http.StatusOK, jwks{Keys: []jwk{{
		Kty:   "RSA",
		KeyID: p.kid,
		Use:   "sig",
		Alg:   "RS256",
		N:     base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:     base64.RawURLEncoding.EncodeToString(eBytes),
	}}})
}

func (p *TestIDP) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, oidcUserInfo{
		Subject:           "test-user-1",
		Email:             "tester@example.test",
		EmailVerified:     true,
		Name:              "Test User",
		PreferredUsername: "tester",
	})
}

func (p *TestIDP) signIDToken(nonce string) (string, error) {
	now := time.Now().UTC()
	header := map[string]string{"alg": "RS256", "typ": "JWT", "kid": p.kid}
	claims := map[string]any{
		"iss":                p.issuer,
		"sub":                "test-user-1",
		"aud":                p.clientID,
		"exp":                now.Add(5 * time.Minute).Unix(),
		"iat":                now.Unix(),
		"nonce":              nonce,
		"email":              "tester@example.test",
		"email_verified":     true,
		"name":               "Test User",
		"preferred_username": "tester",
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signingInput := base64.RawURLEncoding.EncodeToString(headerBytes) + "." + base64.RawURLEncoding.EncodeToString(claimBytes)
	sum := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, p.key, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func bigEndianInt(value int) []byte {
	if value == 0 {
		return []byte{0}
	}
	var out []byte
	for value > 0 {
		out = append([]byte{byte(value)}, out...)
		value >>= 8
	}
	return out
}
