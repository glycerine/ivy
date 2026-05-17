package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/glycerine/ivy/goivy/svk/server/mail"
)

type MagicService struct {
	mu        sync.Mutex
	sender    mail.Sender
	baseURL   string
	now       func() time.Time
	ttl       time.Duration
	tokenHash map[[32]byte]magicToken
}

type magicToken struct {
	Email      string
	Purpose    string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
}

func NewMagicService(sender mail.Sender, baseURL string) *MagicService {
	if sender == nil {
		sender = mail.NoopSender{}
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return &MagicService{
		sender:    sender,
		baseURL:   baseURL,
		now:       time.Now,
		ttl:       15 * time.Minute,
		tokenHash: map[[32]byte]magicToken{},
	}
}

func (s *MagicService) Request(ctx context.Context, email, purpose string) error {
	token := randomToken()
	hash := sha256.Sum256([]byte(token))
	expiresAt := s.now().Add(s.ttl)

	s.mu.Lock()
	s.tokenHash[hash] = magicToken{Email: email, Purpose: purpose, ExpiresAt: expiresAt}
	s.mu.Unlock()

	link := s.link(token)
	return s.sender.Send(ctx, email, "Your SVK sign-in link", "Use this link to sign in to SVK:\n\n"+link+"\n\nThis link expires in 15 minutes.")
}

func (s *MagicService) Consume(rawToken string) (string, error) {
	hash := sha256.Sum256([]byte(rawToken))
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.tokenHash[hash]
	if !ok {
		return "", errors.New("unknown magic token")
	}
	if record.ConsumedAt != nil {
		return "", errors.New("magic token already consumed")
	}
	if !s.now().Before(record.ExpiresAt) {
		return "", errors.New("magic token expired")
	}
	now := s.now()
	record.ConsumedAt = &now
	s.tokenHash[hash] = record
	return record.Email, nil
}

func (s *MagicService) StoredRawToken(rawToken string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, record := range s.tokenHash {
		if record.Email == rawToken || record.Purpose == rawToken {
			return true
		}
	}
	return false
}

func (s *MagicService) link(token string) string {
	u, err := url.Parse(s.baseURL)
	if err != nil {
		u = &url.URL{Scheme: "http", Host: "localhost:8080"}
	}
	u.Path = "/auth/magic/consume"
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String()
}

func randomToken() string {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buf[:])
}
