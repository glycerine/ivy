package control

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db *sql.DB
}

type AdminUnverifiedEmail struct {
	Email     string     `json:"email"`
	LoginURL  string     `json:"loginUrl"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt"`
	Expired   bool       `json:"expired"`
}

func OpenPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresStore) UpsertUserFromOIDC(ctx context.Context, identity OIDCIdentity) (User, error) {
	if err := identity.Validate(); err != nil {
		return User{}, err
	}
	id, err := NewUUID()
	if err != nil {
		return User{}, err
	}
	var verifiedAt any
	if identity.EmailVerified {
		verifiedAt = time.Now().UTC()
	}
	row := s.db.QueryRowContext(ctx, `
INSERT INTO users (id, idp_issuer, idp_subject, email, display_name, email_verified_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (idp_issuer, idp_subject)
DO UPDATE SET
  email = EXCLUDED.email,
  display_name = EXCLUDED.display_name,
  email_verified_at = EXCLUDED.email_verified_at,
  updated_at = now()
RETURNING id, idp_issuer, idp_subject, email, display_name, email_verified_at, disabled_at
`, id, identity.Issuer, identity.Subject, identity.Email, identity.DisplayName, verifiedAt)
	return scanUser(row)
}

func (s *PostgresStore) CreateEmailLoginToken(ctx context.Context, email, rawToken string, now time.Time, ttl time.Duration) error {
	email, err := NormalizeEmail(email)
	if err != nil {
		return err
	}
	if rawToken == "" {
		return errors.New("email login token is required")
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO email_login_tokens
  (token_hash, email, created_at, expires_at)
VALUES
  ($1, $2, $3, $4)
`, hashToken(rawToken), email, now, now.Add(ttl))
	return err
}

func (s *PostgresStore) RecordEmailDelivery(ctx context.Context, toEmail, loginURL string, expiresAt, now time.Time) error {
	toEmail, err := NormalizeEmail(toEmail)
	if err != nil {
		return err
	}
	if strings.TrimSpace(loginURL) == "" {
		return errors.New("login URL is required")
	}
	id, err := NewUUID()
	if err != nil {
		return err
	}
	tokenHash, err := tokenHashFromLoginURL(loginURL)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO email_deliveries
  (id, to_email, kind, login_url, token_hash, provider, created_at, expires_at)
VALUES
  ($1, $2, 'login_link', $3, $4, 'database', $5, $6)
`, id, toEmail, loginURL, tokenHash, now, expiresAt)
	return err
}

func (s *PostgresStore) AdminUnverifiedEmails(ctx context.Context, now time.Time, limit int) ([]AdminUnverifiedEmail, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT d.to_email, d.login_url, d.created_at, d.expires_at, t.used_at
FROM email_deliveries d
LEFT JOIN email_login_tokens t ON t.token_hash = d.token_hash
LEFT JOIN users u ON u.email = d.to_email
WHERE d.kind = 'login_link'
  AND u.email_verified_at IS NULL
ORDER BY d.created_at DESC
LIMIT $1
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var emails []AdminUnverifiedEmail
	for rows.Next() {
		var email AdminUnverifiedEmail
		var usedAt sql.NullTime
		if err := rows.Scan(&email.Email, &email.LoginURL, &email.CreatedAt, &email.ExpiresAt, &usedAt); err != nil {
			return nil, err
		}
		if usedAt.Valid {
			email.UsedAt = &usedAt.Time
		}
		email.Expired = !email.ExpiresAt.After(now)
		emails = append(emails, email)
	}
	return emails, rows.Err()
}

func (s *PostgresStore) LatestEmailDeliveryFor(ctx context.Context, email string) (EmailMessage, bool, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return EmailMessage{}, false, err
	}
	var message EmailMessage
	err = s.db.QueryRowContext(ctx, `
SELECT to_email, login_url, expires_at
FROM email_deliveries
WHERE to_email = $1
  AND kind = 'login_link'
ORDER BY created_at DESC
LIMIT 1
`, email).Scan(&message.ToEmail, &message.LoginURL, &message.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return EmailMessage{}, false, nil
	}
	if err != nil {
		return EmailMessage{}, false, err
	}
	return message, true, nil
}

func (s *PostgresStore) EmailLoginTokenDebug(ctx context.Context, rawToken string, now time.Time) (EmailLoginTokenDebug, bool, error) {
	if rawToken == "" {
		return EmailLoginTokenDebug{}, false, nil
	}
	var debug EmailLoginTokenDebug
	var usedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
SELECT email, created_at, expires_at, used_at
FROM email_login_tokens
WHERE token_hash = $1
`, hashToken(rawToken)).Scan(&debug.Email, &debug.CreatedAt, &debug.ExpiresAt, &usedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return EmailLoginTokenDebug{}, false, nil
	}
	if err != nil {
		return EmailLoginTokenDebug{}, false, err
	}
	if usedAt.Valid {
		debug.UsedAt = &usedAt.Time
	}
	debug.Expired = !debug.ExpiresAt.After(now)
	return debug, true, nil
}

func (s *PostgresStore) EmailVerified(ctx context.Context, email string) (bool, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return false, err
	}
	var verified bool
	err = s.db.QueryRowContext(ctx, `
SELECT email_verified_at IS NOT NULL
FROM users
WHERE email = $1
`, email).Scan(&verified)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return verified, err
}

func (s *PostgresStore) ConsumeEmailLoginToken(ctx context.Context, rawToken string, now time.Time) (User, error) {
	if rawToken == "" {
		return User{}, ErrEmailLoginTokenNotFound
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	var email string
	err = tx.QueryRowContext(ctx, `
SELECT email
FROM email_login_tokens
WHERE token_hash = $1
  AND used_at IS NULL
  AND expires_at > $2
FOR UPDATE
`, hashToken(rawToken), now).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrEmailLoginTokenNotFound
	}
	if err != nil {
		return User{}, err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE email_login_tokens
SET used_at = $2
WHERE token_hash = $1
`, hashToken(rawToken), now); err != nil {
		return User{}, err
	}
	user, err := upsertEmailUserTx(ctx, tx, email, now)
	if err != nil {
		return User{}, err
	}
	return user, tx.Commit()
}

func (s *PostgresStore) EnsureStarterWorkspace(ctx context.Context, user User) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	accountID, err := ensureAccount(ctx, tx, starterSlug(user.Email, user.ID), starterDisplayName(user), user.Email, "trial")
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO account_users (account_id, user_id, role)
VALUES ($1, $2, 'owner')
ON CONFLICT (account_id, user_id)
DO UPDATE SET role = EXCLUDED.role, disabled_at = NULL, updated_at = now()
`, accountID, user.ID); err != nil {
		return err
	}
	teamID, err := ensureTeam(ctx, tx, accountID, "core", "Core")
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO team_memberships (team_id, user_id, role)
VALUES ($1, $2, 'owner')
ON CONFLICT (team_id, user_id)
DO UPDATE SET role = EXCLUDED.role, disabled_at = NULL, updated_at = now()
`, teamID, user.ID); err != nil {
		return err
	}
	projectID, err := ensureProject(ctx, tx, accountID, "client-server", "Client/server example", user.ID)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO project_grants (project_id, subject_kind, subject_id, role)
VALUES ($1, 'user', $2, 'admin')
ON CONFLICT (project_id, subject_kind, subject_id)
DO UPDATE SET role = EXCLUDED.role, disabled_at = NULL, updated_at = now()
`, projectID, user.ID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO project_storage_locations (project_id)
VALUES ($1)
ON CONFLICT (project_id) DO NOTHING
`, projectID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) CreateAppSession(ctx context.Context, userID, rawSessionToken, rawCSRFToken string, now time.Time, idleTTL, absoluteTTL time.Duration) error {
	if rawSessionToken == "" || rawCSRFToken == "" {
		return errors.New("session token and csrf token are required")
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO app_sessions
  (id_hash, user_id, csrf_token_hash, created_at, last_seen_at, idle_expires_at, absolute_expires_at)
VALUES
  ($1, $2, $3, $4, $4, $5, $6)
`, hashToken(rawSessionToken), userID, hashToken(rawCSRFToken), now, now.Add(idleTTL), now.Add(absoluteTTL))
	return err
}

func (s *PostgresStore) TouchSessionByToken(ctx context.Context, rawSessionToken string, now time.Time, ttl time.Duration) (SessionTouch, error) {
	if strings.TrimSpace(rawSessionToken) == "" {
		return SessionTouch{}, ErrSessionNotFound
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SessionTouch{}, err
	}
	defer tx.Rollback()

	touch, err := touchSessionTx(ctx, tx, hashToken(rawSessionToken), now, ttl)
	if err != nil {
		return SessionTouch{}, err
	}
	return touch, tx.Commit()
}

func (s *PostgresStore) SessionViewByToken(ctx context.Context, rawSessionToken string, now time.Time, idleTTL time.Duration) (SessionView, error) {
	if strings.TrimSpace(rawSessionToken) == "" {
		return SessionView{}, ErrSessionNotFound
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SessionView{}, err
	}
	defer tx.Rollback()

	touch, err := touchSessionTx(ctx, tx, hashToken(rawSessionToken), now, idleTTL)
	if err != nil {
		return SessionView{}, err
	}
	view, err := sessionViewForUser(ctx, tx, touch.UserID)
	if err != nil {
		return SessionView{}, err
	}
	view.CookieRefreshNeeded = touch.CookieRefreshNeeded
	return view, tx.Commit()
}

func touchSessionTx(ctx context.Context, tx *sql.Tx, sessionHash []byte, now time.Time, ttl time.Duration) (SessionTouch, error) {
	var userID string
	var lastSeenAt time.Time
	err := tx.QueryRowContext(ctx, `
SELECT user_id::text, last_seen_at
FROM app_sessions
WHERE id_hash = $1
  AND revoked_at IS NULL
  AND idle_expires_at > $2
  AND absolute_expires_at > $2
`, sessionHash, now).Scan(&userID, &lastSeenAt)
	if errors.Is(err, sql.ErrNoRows) {
		return SessionTouch{}, ErrSessionNotFound
	}
	if err != nil {
		return SessionTouch{}, err
	}

	touch := SessionTouch{
		UserID:               userID,
		CookieRefreshNeeded:  !sameUTCDate(lastSeenAt, now),
		PreviousLastSeenAt:   lastSeenAt,
		CurrentVisitRecorded: now,
	}
	visitID, err := NewUUID()
	if err != nil {
		return SessionTouch{}, err
	}
	visitHour := now.UTC().Truncate(time.Hour)
	err = tx.QueryRowContext(ctx, `
WITH inserted AS (
  INSERT INTO visiting_hours (id, user_id, session_id_hash, visited_at, visited_hour)
  VALUES ($1, $2, $3, $4, $5)
  ON CONFLICT (user_id, visited_hour) DO NOTHING
  RETURNING 1
)
SELECT EXISTS (SELECT 1 FROM inserted)
`, visitID, userID, sessionHash, now, visitHour).Scan(&touch.VisitHourInserted)
	if err != nil {
		return SessionTouch{}, err
	}
	if !sameUTCHour(lastSeenAt, now) {
		if _, err := tx.ExecContext(ctx, `
UPDATE app_sessions
SET last_seen_at = $2,
    idle_expires_at = $3,
    absolute_expires_at = $3
WHERE id_hash = $1
`, sessionHash, now, now.Add(ttl)); err != nil {
			return SessionTouch{}, err
		}
	}
	return touch, nil
}

func upsertEmailUserTx(ctx context.Context, tx *sql.Tx, email string, now time.Time) (User, error) {
	id, err := NewUUID()
	if err != nil {
		return User{}, err
	}
	row := tx.QueryRowContext(ctx, `
INSERT INTO users (id, idp_issuer, idp_subject, email, display_name, email_verified_at)
VALUES ($1, 'email', $2, $2, $2, $3)
ON CONFLICT (idp_issuer, idp_subject)
DO UPDATE SET
  email = EXCLUDED.email,
  display_name = COALESCE(NULLIF(users.display_name, ''), EXCLUDED.display_name),
  email_verified_at = COALESCE(users.email_verified_at, EXCLUDED.email_verified_at),
  updated_at = now()
RETURNING id::text, idp_issuer, idp_subject, email, display_name, email_verified_at, disabled_at
`, id, email, now)
	return scanUser(row)
}

func scanUser(row interface {
	Scan(dest ...any) error
}) (User, error) {
	var user User
	var verifiedAt sql.NullTime
	var disabledAt sql.NullTime
	if err := row.Scan(&user.ID, &user.IDPIssuer, &user.IDPSubject, &user.Email, &user.DisplayName, &verifiedAt, &disabledAt); err != nil {
		return User{}, err
	}
	if verifiedAt.Valid {
		user.EmailVerifiedAt = &verifiedAt.Time
	}
	if disabledAt.Valid {
		user.DisabledAt = &disabledAt.Time
	}
	return user, nil
}

func ensureAccount(ctx context.Context, tx *sql.Tx, slug, displayName, billingEmail, billingStatus string) (string, error) {
	id, err := NewUUID()
	if err != nil {
		return "", err
	}
	var accountID string
	err = tx.QueryRowContext(ctx, `
INSERT INTO accounts (id, slug, display_name, billing_email, billing_status)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (slug)
DO UPDATE SET
  display_name = EXCLUDED.display_name,
  billing_email = EXCLUDED.billing_email,
  billing_status = EXCLUDED.billing_status,
  disabled_at = NULL,
  updated_at = now()
RETURNING id::text
`, id, slug, displayName, billingEmail, billingStatus).Scan(&accountID)
	return accountID, err
}

func ensureTeam(ctx context.Context, tx *sql.Tx, accountID, slug, displayName string) (string, error) {
	id, err := NewUUID()
	if err != nil {
		return "", err
	}
	var teamID string
	err = tx.QueryRowContext(ctx, `
INSERT INTO teams (id, account_id, slug, display_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (account_id, slug)
DO UPDATE SET
  display_name = EXCLUDED.display_name,
  disabled_at = NULL,
  updated_at = now()
RETURNING id::text
`, id, accountID, slug, displayName).Scan(&teamID)
	return teamID, err
}

func ensureProject(ctx context.Context, tx *sql.Tx, accountID, slug, displayName, createdByUserID string) (string, error) {
	id, err := NewUUID()
	if err != nil {
		return "", err
	}
	var projectID string
	err = tx.QueryRowContext(ctx, `
INSERT INTO projects (id, account_id, slug, display_name, created_by_user_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (account_id, slug)
DO UPDATE SET
  display_name = EXCLUDED.display_name,
  disabled_at = NULL,
  updated_at = now()
RETURNING id::text
`, id, accountID, slug, displayName, createdByUserID).Scan(&projectID)
	return projectID, err
}

func sessionViewForUser(ctx context.Context, tx *sql.Tx, userID string) (SessionView, error) {
	user, err := userByID(ctx, tx, userID)
	if err != nil {
		return SessionView{}, err
	}
	accounts, err := accountsForUser(ctx, tx, userID)
	if err != nil {
		return SessionView{}, err
	}
	teams, err := teamsForUser(ctx, tx, userID)
	if err != nil {
		return SessionView{}, err
	}
	projects, roles, err := projectsForUser(ctx, tx, userID)
	if err != nil {
		return SessionView{}, err
	}
	return SessionView{
		Authenticated: true,
		User:          &user,
		Accounts:      accounts,
		Teams:         teams,
		Projects:      projects,
		Roles:         roles,
	}, nil
}

func userByID(ctx context.Context, tx *sql.Tx, userID string) (User, error) {
	row := tx.QueryRowContext(ctx, `
SELECT id::text, idp_issuer, idp_subject, email, display_name, email_verified_at, disabled_at
FROM users
WHERE id = $1 AND disabled_at IS NULL
`, userID)
	return scanUser(row)
}

func accountsForUser(ctx context.Context, tx *sql.Tx, userID string) ([]Account, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT a.id::text, a.slug, a.display_name, a.billing_email, a.billing_status, a.disabled_at
FROM accounts a
JOIN account_users au ON au.account_id = a.id
WHERE au.user_id = $1
  AND au.disabled_at IS NULL
  AND a.disabled_at IS NULL
ORDER BY a.slug
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []Account
	for rows.Next() {
		var account Account
		var disabledAt sql.NullTime
		if err := rows.Scan(&account.ID, &account.Slug, &account.DisplayName, &account.BillingEmail, &account.BillingStatus, &disabledAt); err != nil {
			return nil, err
		}
		if disabledAt.Valid {
			account.DisabledAt = &disabledAt.Time
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

func teamsForUser(ctx context.Context, tx *sql.Tx, userID string) ([]Team, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT t.id::text, t.account_id::text, t.slug, t.display_name, t.disabled_at
FROM teams t
JOIN team_memberships tm ON tm.team_id = t.id
WHERE tm.user_id = $1
  AND tm.disabled_at IS NULL
  AND t.disabled_at IS NULL
ORDER BY t.slug
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var teams []Team
	for rows.Next() {
		var team Team
		var disabledAt sql.NullTime
		if err := rows.Scan(&team.ID, &team.AccountID, &team.Slug, &team.DisplayName, &disabledAt); err != nil {
			return nil, err
		}
		if disabledAt.Valid {
			team.DisabledAt = &disabledAt.Time
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func projectsForUser(ctx context.Context, tx *sql.Tx, userID string) ([]Project, map[string]string, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT DISTINCT p.id::text, p.account_id::text, p.slug, p.display_name, p.created_by_user_id::text, p.disabled_at, pg.role
FROM projects p
JOIN project_grants pg ON pg.project_id = p.id
LEFT JOIN account_users au
  ON au.account_id = p.account_id
 AND au.user_id = $1
 AND au.disabled_at IS NULL
LEFT JOIN team_memberships tm
  ON tm.team_id = pg.subject_id
 AND tm.user_id = $1
 AND tm.disabled_at IS NULL
WHERE p.disabled_at IS NULL
  AND pg.disabled_at IS NULL
  AND (
    (pg.subject_kind = 'user' AND pg.subject_id = $1)
    OR (pg.subject_kind = 'account' AND pg.subject_id = p.account_id AND au.user_id IS NOT NULL)
    OR (pg.subject_kind = 'team' AND tm.user_id IS NOT NULL)
  )
ORDER BY p.slug
`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	byID := map[string]Project{}
	roles := map[string]string{}
	for rows.Next() {
		var project Project
		var role string
		var disabledAt sql.NullTime
		if err := rows.Scan(&project.ID, &project.AccountID, &project.Slug, &project.DisplayName, &project.CreatedByUserID, &disabledAt, &role); err != nil {
			return nil, nil, err
		}
		if disabledAt.Valid {
			project.DisabledAt = &disabledAt.Time
		}
		byID[project.ID] = project
		if roleRank(role) > roleRank(roles[project.ID]) {
			roles[project.ID] = role
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	projects := make([]Project, 0, len(byID))
	for _, project := range byID {
		projects = append(projects, project)
	}
	return projects, roles, nil
}

func roleRank(role string) int {
	switch role {
	case string(ProjectRoleAdmin):
		return 3
	case string(ProjectRoleWrite):
		return 2
	case string(ProjectRoleRead):
		return 1
	default:
		return 0
	}
}

func sameUTCDate(a, b time.Time) bool {
	au := a.UTC()
	bu := b.UTC()
	return au.Year() == bu.Year() && au.YearDay() == bu.YearDay()
}

func sameUTCHour(a, b time.Time) bool {
	return a.UTC().Truncate(time.Hour).Equal(b.UTC().Truncate(time.Hour))
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("nil postgres store")
	}
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	return nil
}
