package control

import (
	"context"
	"crypto/sha256"
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrEmailLoginTokenNotFound = errors.New("email login token not found")
	ErrSessionNotFound         = errors.New("session not found")
)

type Store interface {
	UpsertUserFromOIDC(ctx context.Context, identity OIDCIdentity) (User, error)
	EnsureStarterWorkspace(ctx context.Context, user User) error
	CreateEmailLoginToken(ctx context.Context, email, rawToken string, now time.Time, ttl time.Duration) error
	ConsumeEmailLoginToken(ctx context.Context, rawToken string, now time.Time) (User, error)
	CreateAppSession(ctx context.Context, userID, rawSessionToken, rawCSRFToken string, now time.Time, idleTTL, absoluteTTL time.Duration) error
	TouchSessionByToken(ctx context.Context, rawSessionToken string, now time.Time, ttl time.Duration) (SessionTouch, error)
	SessionViewByToken(ctx context.Context, rawSessionToken string, now time.Time, idleTTL time.Duration) (SessionView, error)
}

type EmailLoginTokenDebug struct {
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt"`
	Expired   bool       `json:"expired"`
}

type EmailDebugStore interface {
	EmailLoginTokenDebug(ctx context.Context, rawToken string, now time.Time) (EmailLoginTokenDebug, bool, error)
	EmailVerified(ctx context.Context, email string) (bool, error)
}

type OIDCIdentity struct {
	Issuer        string
	Subject       string
	Email         string
	DisplayName   string
	EmailVerified bool
}

func (i OIDCIdentity) Validate() error {
	if strings.TrimSpace(i.Issuer) == "" {
		return errors.New("oidc issuer is required")
	}
	if strings.TrimSpace(i.Subject) == "" {
		return errors.New("oidc subject is required")
	}
	if strings.TrimSpace(i.Email) == "" {
		return errors.New("oidc email is required")
	}
	return nil
}

func hashToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

var slugUnsafe = regexp.MustCompile(`[^a-z0-9-]+`)

func starterSlug(email, userID string) string {
	local := strings.ToLower(strings.TrimSpace(strings.Split(email, "@")[0]))
	local = strings.Trim(slugUnsafe.ReplaceAllString(local, "-"), "-")
	if local == "" {
		local = "user"
	}
	suffix := userID
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	return local + "-" + suffix
}

func starterDisplayName(user User) string {
	if strings.TrimSpace(user.DisplayName) != "" {
		return user.DisplayName
	}
	if strings.TrimSpace(user.Email) != "" {
		return user.Email
	}
	return "Starter Account"
}
