package control

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"sync"
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
	SessionViewByToken(ctx context.Context, rawSessionToken string, now time.Time, idleTTL time.Duration) (SessionView, error)
}

type OIDCIdentity struct {
	Issuer        string
	Subject       string
	Email         string
	DisplayName   string
	EmailVerified bool
}

type MemoryStore struct {
	mu             sync.Mutex
	usersByIDP     map[string]User
	usersByID      map[string]User
	emailTokens    map[string]memoryEmailToken
	sessionByToken map[string]memoryAppSession
	accountsByU    map[string][]Account
	teamsByU       map[string][]Team
	projectsByU    map[string][]Project
	rolesByU       map[string]map[string]string
}

type memoryEmailToken struct {
	Email     string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

type memoryAppSession struct {
	UserID            string
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		usersByIDP:     map[string]User{},
		usersByID:      map[string]User{},
		emailTokens:    map[string]memoryEmailToken{},
		sessionByToken: map[string]memoryAppSession{},
		accountsByU:    map[string][]Account{},
		teamsByU:       map[string][]Team{},
		projectsByU:    map[string][]Project{},
		rolesByU:       map[string]map[string]string{},
	}
}

func (s *MemoryStore) UpsertUserFromOIDC(ctx context.Context, identity OIDCIdentity) (User, error) {
	if err := identity.Validate(); err != nil {
		return User{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	key := identityKey(identity.Issuer, identity.Subject)
	user, ok := s.usersByIDP[key]
	if !ok {
		id, err := NewUUID()
		if err != nil {
			return User{}, err
		}
		user.ID = id
		user.IDPIssuer = identity.Issuer
		user.IDPSubject = identity.Subject
	}
	user.Email = identity.Email
	user.DisplayName = identity.DisplayName
	if identity.EmailVerified {
		now := time.Now().UTC()
		user.EmailVerifiedAt = &now
	} else {
		user.EmailVerifiedAt = nil
	}
	s.usersByIDP[key] = user
	s.usersByID[user.ID] = user
	return user, nil
}

func (s *MemoryStore) EnsureStarterWorkspace(ctx context.Context, user User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.accountsByU[user.ID]) != 0 {
		return nil
	}
	accountID, err := NewUUID()
	if err != nil {
		return err
	}
	teamID, err := NewUUID()
	if err != nil {
		return err
	}
	projectID, err := NewUUID()
	if err != nil {
		return err
	}
	slug := starterSlug(user.Email, user.ID)
	account := Account{
		ID:            accountID,
		Slug:          slug,
		DisplayName:   starterDisplayName(user),
		BillingEmail:  user.Email,
		BillingStatus: "trial",
	}
	team := Team{
		ID:          teamID,
		AccountID:   accountID,
		Slug:        "core",
		DisplayName: "Core",
	}
	project := Project{
		ID:              projectID,
		AccountID:       accountID,
		Slug:            "client-server",
		DisplayName:     "Client/server example",
		CreatedByUserID: user.ID,
	}
	s.accountsByU[user.ID] = []Account{account}
	s.teamsByU[user.ID] = []Team{team}
	s.projectsByU[user.ID] = []Project{project}
	s.rolesByU[user.ID] = map[string]string{projectID: string(ProjectRoleAdmin)}
	return nil
}

func (s *MemoryStore) CreateEmailLoginToken(ctx context.Context, email, rawToken string, now time.Time, ttl time.Duration) error {
	email, err := NormalizeEmail(email)
	if err != nil {
		return err
	}
	if rawToken == "" {
		return errors.New("email login token is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emailTokens[hashTokenString(rawToken)] = memoryEmailToken{
		Email:     email,
		ExpiresAt: now.Add(ttl),
	}
	return nil
}

func (s *MemoryStore) ConsumeEmailLoginToken(ctx context.Context, rawToken string, now time.Time) (User, error) {
	if rawToken == "" {
		return User{}, ErrEmailLoginTokenNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tokenHash := hashTokenString(rawToken)
	token, ok := s.emailTokens[tokenHash]
	if !ok || token.UsedAt != nil || !token.ExpiresAt.After(now) {
		return User{}, ErrEmailLoginTokenNotFound
	}
	token.UsedAt = &now
	s.emailTokens[tokenHash] = token
	user, err := s.upsertEmailUserLocked(token.Email, now)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *MemoryStore) CreateAppSession(ctx context.Context, userID, rawSessionToken, rawCSRFToken string, now time.Time, idleTTL, absoluteTTL time.Duration) error {
	if userID == "" || rawSessionToken == "" || rawCSRFToken == "" {
		return errors.New("session requires user id, token, and csrf token")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.usersByID[userID]; !ok {
		return errors.New("session user not found")
	}
	s.sessionByToken[hashTokenString(rawSessionToken)] = memoryAppSession{
		UserID:            userID,
		IdleExpiresAt:     now.Add(idleTTL),
		AbsoluteExpiresAt: now.Add(absoluteTTL),
	}
	return nil
}

func (s *MemoryStore) SessionViewByToken(ctx context.Context, rawSessionToken string, now time.Time, idleTTL time.Duration) (SessionView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash := hashTokenString(rawSessionToken)
	var session memoryAppSession
	var sessionHash string
	for storedHash, candidateSession := range s.sessionByToken {
		if subtle.ConstantTimeCompare([]byte(storedHash), []byte(hash)) == 1 {
			session = candidateSession
			sessionHash = storedHash
			break
		}
	}
	if session.UserID == "" || !session.IdleExpiresAt.After(now) || !session.AbsoluteExpiresAt.After(now) {
		return SessionView{}, ErrSessionNotFound
	}
	session.IdleExpiresAt = now.Add(idleTTL)
	s.sessionByToken[sessionHash] = session
	userID := session.UserID
	user := s.usersByID[userID]
	return SessionView{
		Authenticated: true,
		User:          &user,
		Accounts:      append([]Account(nil), s.accountsByU[userID]...),
		Teams:         append([]Team(nil), s.teamsByU[userID]...),
		Projects:      append([]Project(nil), s.projectsByU[userID]...),
		Roles:         cloneStringMap(s.rolesByU[userID]),
	}, nil
}

func (s *MemoryStore) upsertEmailUserLocked(email string, now time.Time) (User, error) {
	key := identityKey("email", email)
	user, ok := s.usersByIDP[key]
	if !ok {
		id, err := NewUUID()
		if err != nil {
			return User{}, err
		}
		user.ID = id
		user.IDPIssuer = "email"
		user.IDPSubject = email
		user.Email = email
		user.DisplayName = email
	}
	user.Email = email
	user.EmailVerifiedAt = &now
	s.usersByIDP[key] = user
	s.usersByID[user.ID] = user
	return user, nil
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

func identityKey(issuer, subject string) string {
	return strings.TrimRight(issuer, "/") + "\x00" + subject
}

func hashToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

func hashTokenString(raw string) string {
	return hex.EncodeToString(hashToken(raw))
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
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
