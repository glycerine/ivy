package control

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	EmailLoginTokenTTL = 10 * time.Minute
	AppSessionTTL      = 400 * 24 * time.Hour
)

type EmailSender interface {
	SendLoginLink(ctx context.Context, toEmail, loginURL string, expiresAt time.Time) error
}

type EmailAuthService struct {
	Store                     Store
	Sender                    EmailSender
	BaseURL                   string
	AutoProvisionStarterSpace bool
}

func (s EmailAuthService) RequestLogin(ctx context.Context, email string, now time.Time) error {
	if s.Store == nil {
		return errors.New("email auth store is required")
	}
	if s.Sender == nil {
		return errors.New("email sender is required")
	}
	email, err := NormalizeEmail(email)
	if err != nil {
		return err
	}
	token, err := RandomToken(32)
	if err != nil {
		return err
	}
	if err := s.Store.CreateEmailLoginToken(ctx, email, token, now, EmailLoginTokenTTL); err != nil {
		return err
	}
	return s.Sender.SendLoginLink(ctx, email, s.loginURL(token), now.Add(EmailLoginTokenTTL))
}

func (s EmailAuthService) ConsumeLogin(ctx context.Context, rawToken string, now time.Time) (string, string, User, error) {
	if s.Store == nil {
		return "", "", User{}, errors.New("email auth store is required")
	}
	user, err := s.Store.ConsumeEmailLoginToken(ctx, rawToken, now)
	if err != nil {
		return "", "", User{}, err
	}
	if s.AutoProvisionStarterSpace {
		if err := s.Store.EnsureStarterWorkspace(ctx, user); err != nil {
			return "", "", User{}, err
		}
	}
	sessionToken, err := RandomToken(32)
	if err != nil {
		return "", "", User{}, err
	}
	csrfToken, err := RandomToken(32)
	if err != nil {
		return "", "", User{}, err
	}
	if err := s.Store.CreateAppSession(ctx, user.ID, sessionToken, csrfToken, now, AppSessionTTL, AppSessionTTL); err != nil {
		return "", "", User{}, err
	}
	return sessionToken, csrfToken, user, nil
}

func (s EmailAuthService) loginURL(token string) string {
	baseURL := strings.TrimRight(s.BaseURL, "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:18080"
	}
	return fmt.Sprintf("%s/auth/email/continue#token=%s", baseURL, token)
}

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func NormalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !emailPattern.MatchString(email) {
		return "", errors.New("valid email is required")
	}
	return email, nil
}
