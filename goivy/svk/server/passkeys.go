package server

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

type PasskeyService struct {
	mu          sync.Mutex
	rp          *webauthn.WebAuthn
	now         func() time.Time
	ttl         time.Duration
	challenges  map[string]passkeyChallenge
	credentials map[string]passkeyCredential
}

type passkeyChallenge struct {
	UserID    string
	Purpose   string
	ExpiresAt time.Time
	Consumed  bool
}

type passkeyCredential struct {
	UserID   string
	Disabled bool
}

func NewPasskeyService(rpID, origin string) (*PasskeyService, error) {
	if rpID == "" {
		rpID = "localhost"
	}
	if origin == "" {
		origin = "http://localhost:8080"
	}
	rp, err := webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: "SVK",
		RPOrigins:     []string{origin},
	})
	if err != nil {
		return nil, err
	}
	return &PasskeyService{
		rp:          rp,
		now:         time.Now,
		ttl:         5 * time.Minute,
		challenges:  map[string]passkeyChallenge{},
		credentials: map[string]passkeyCredential{},
	}, nil
}

func (s *PasskeyService) RegistrationOptions(userID string) string {
	return s.newChallenge(userID, "registration")
}

func (s *PasskeyService) FinishRegistration(userID, challenge, credentialID string) error {
	if err := s.consumeChallenge(userID, "registration", challenge); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[credentialID] = passkeyCredential{UserID: userID}
	return nil
}

func (s *PasskeyService) LoginOptions(userID string) string {
	return s.newChallenge(userID, "login")
}

func (s *PasskeyService) FinishLogin(userID, challenge, credentialID string) error {
	if err := s.consumeChallenge(userID, "login", challenge); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	credential, ok := s.credentials[credentialID]
	if !ok || credential.UserID != userID {
		return errors.New("unknown passkey credential")
	}
	if credential.Disabled {
		return errors.New("passkey credential disabled")
	}
	return nil
}

func (s *PasskeyService) DisableCredential(credentialID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	credential := s.credentials[credentialID]
	credential.Disabled = true
	s.credentials[credentialID] = credential
}

func (s *PasskeyService) newChallenge(userID, purpose string) string {
	challenge := randomPasskeyChallenge()
	s.mu.Lock()
	s.challenges[challenge] = passkeyChallenge{UserID: userID, Purpose: purpose, ExpiresAt: s.now().Add(s.ttl)}
	s.mu.Unlock()
	return challenge
}

func (s *PasskeyService) consumeChallenge(userID, purpose, challenge string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.challenges[challenge]
	if !ok || record.UserID != userID || record.Purpose != purpose {
		return errors.New("invalid passkey challenge")
	}
	if record.Consumed {
		return errors.New("passkey challenge already consumed")
	}
	if !s.now().Before(record.ExpiresAt) {
		return errors.New("passkey challenge expired")
	}
	record.Consumed = true
	s.challenges[challenge] = record
	return nil
}

func randomPasskeyChallenge() string {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buf[:])
}
