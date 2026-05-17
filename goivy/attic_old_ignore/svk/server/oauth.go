package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"sync"
	"time"
)

var ErrExplicitLinkRequired = errors.New("explicit identity linking required")

type OAuthService struct {
	mu          sync.Mutex
	now         func() time.Time
	ttl         time.Duration
	states      map[[32]byte]oauthState
	identities  map[string]string
	usersByMail map[string]string
	nextUser    int
}

type oauthState struct {
	Provider     string
	CodeVerifier string
	ExpiresAt    time.Time
}

type OAuthResult struct {
	UserID string
	Linked bool
}

func NewOAuthService() *OAuthService {
	return &OAuthService{
		now:         time.Now,
		ttl:         10 * time.Minute,
		states:      map[[32]byte]oauthState{},
		identities:  map[string]string{},
		usersByMail: map[string]string{},
	}
}

func (s *OAuthService) Start(provider string) (state, verifier, authURL string) {
	state = randomURLToken()
	verifier = randomURLToken()
	hash := sha256.Sum256([]byte(state))
	s.mu.Lock()
	s.states[hash] = oauthState{Provider: provider, CodeVerifier: verifier, ExpiresAt: s.now().Add(s.ttl)}
	s.mu.Unlock()

	u := url.URL{Scheme: "https", Host: provider + ".oauth.local", Path: "/authorize"}
	q := u.Query()
	q.Set("state", state)
	q.Set("code_challenge", pkceChallenge(verifier))
	q.Set("code_challenge_method", "S256")
	u.RawQuery = q.Encode()
	return state, verifier, u.String()
}

func (s *OAuthService) Callback(provider, state, subject, email string) (OAuthResult, error) {
	hash := sha256.Sum256([]byte(state))
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.states[hash]
	if !ok || record.Provider != provider {
		return OAuthResult{}, errors.New("invalid oauth state")
	}
	delete(s.states, hash)
	if !s.now().Before(record.ExpiresAt) {
		return OAuthResult{}, errors.New("expired oauth state")
	}

	key := provider + ":" + subject
	if userID, ok := s.identities[key]; ok {
		return OAuthResult{UserID: userID}, nil
	}
	if existing := s.usersByMail[email]; existing != "" {
		return OAuthResult{UserID: existing}, ErrExplicitLinkRequired
	}
	s.nextUser++
	userID := "oauth-user-" + base64.RawURLEncoding.EncodeToString([]byte{byte(s.nextUser)})
	s.usersByMail[email] = userID
	s.identities[key] = userID
	return OAuthResult{UserID: userID}, nil
}

func (s *OAuthService) LinkIdentity(userID, provider, subject string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := provider + ":" + subject
	if existing := s.identities[key]; existing != "" && existing != userID {
		return errors.New("oauth identity already linked")
	}
	s.identities[key] = userID
	return nil
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func randomURLToken() string {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buf[:])
}
