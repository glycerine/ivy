package control

import (
	"context"
	"errors"
	"strings"
)

type SeedAlphaTesterParams struct {
	Email       string
	DisplayName string
	IDPIssuer   string
	IDPSubject  string
	AccountSlug string
	AccountName string
	TeamSlug    string
	TeamName    string
	ProjectSlug string
	ProjectName string
	ProjectRole ProjectRole
}

func (s *PostgresStore) SeedAlphaTester(ctx context.Context, params SeedAlphaTesterParams) (SessionView, error) {
	params = normalizeSeedAlphaTesterParams(params)
	if strings.TrimSpace(params.Email) == "" {
		return SessionView{}, errors.New("email is required")
	}
	user, err := s.UpsertUserFromOIDC(ctx, OIDCIdentity{
		Issuer:        params.IDPIssuer,
		Subject:       params.IDPSubject,
		Email:         params.Email,
		DisplayName:   params.DisplayName,
		EmailVerified: true,
	})
	if err != nil {
		return SessionView{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SessionView{}, err
	}
	defer tx.Rollback()

	accountID, err := ensureAccount(ctx, tx, params.AccountSlug, params.AccountName, params.Email, "alpha")
	if err != nil {
		return SessionView{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO control.account_users (account_id, user_id, role)
VALUES ($1, $2, 'member')
ON CONFLICT (account_id, user_id)
DO UPDATE SET disabled_at = NULL, updated_at = now()
`, accountID, user.ID); err != nil {
		return SessionView{}, err
	}
	teamID, err := ensureTeam(ctx, tx, accountID, params.TeamSlug, params.TeamName)
	if err != nil {
		return SessionView{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO control.team_memberships (team_id, user_id, role)
VALUES ($1, $2, 'member')
ON CONFLICT (team_id, user_id)
DO UPDATE SET disabled_at = NULL, updated_at = now()
`, teamID, user.ID); err != nil {
		return SessionView{}, err
	}
	projectID, err := ensureProject(ctx, tx, accountID, params.ProjectSlug, params.ProjectName, user.ID)
	if err != nil {
		return SessionView{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO control.project_grants (project_id, subject_kind, subject_id, role)
VALUES ($1, 'user', $2, $3)
ON CONFLICT (project_id, subject_kind, subject_id)
DO UPDATE SET role = EXCLUDED.role, disabled_at = NULL, updated_at = now()
`, projectID, user.ID, string(params.ProjectRole)); err != nil {
		return SessionView{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO control.project_storage_locations (project_id)
VALUES ($1)
ON CONFLICT (project_id) DO NOTHING
`, projectID); err != nil {
		return SessionView{}, err
	}
	view, err := sessionViewForUser(ctx, tx, user.ID)
	if err != nil {
		return SessionView{}, err
	}
	return view, tx.Commit()
}

func normalizeSeedAlphaTesterParams(params SeedAlphaTesterParams) SeedAlphaTesterParams {
	if params.IDPIssuer == "" {
		params.IDPIssuer = "email"
	}
	if params.IDPSubject == "" {
		params.IDPSubject = params.Email
	}
	if params.DisplayName == "" {
		params.DisplayName = params.Email
	}
	if params.AccountSlug == "" {
		params.AccountSlug = "alpha"
	}
	if params.AccountName == "" {
		params.AccountName = "Alpha Testers"
	}
	if params.TeamSlug == "" {
		params.TeamSlug = "core"
	}
	if params.TeamName == "" {
		params.TeamName = "Core"
	}
	if params.ProjectSlug == "" {
		params.ProjectSlug = "client-server"
	}
	if params.ProjectName == "" {
		params.ProjectName = "Client/server example"
	}
	if params.ProjectRole == "" {
		params.ProjectRole = ProjectRoleWrite
	}
	return params
}
