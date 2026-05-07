package control

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestPostgresStoreMapsOIDCUserSessionAndStarterWorkspace(t *testing.T) {
	dsn := os.Getenv("IVY_CONTROL_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set IVY_CONTROL_TEST_DATABASE_DSN to run PostgreSQL control store test")
	}
	store, err := OpenPostgresStore(dsn)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	subject, err := NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	identity := OIDCIdentity{
		Issuer:        "http://casdoor.local.test",
		Subject:       subject,
		Email:         "tester+" + subject[:8] + "@example.test",
		DisplayName:   "Postgres Tester",
		EmailVerified: true,
	}
	user, err := store.UpsertUserFromOIDC(ctx, identity)
	if err != nil {
		t.Fatalf("upsert user: %v", err)
	}
	if err := store.EnsureStarterWorkspace(ctx, user); err != nil {
		t.Fatalf("ensure starter workspace: %v", err)
	}
	sessionToken, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	csrfToken, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := store.CreateAppSession(ctx, user.ID, sessionToken, csrfToken, now, time.Hour, 24*time.Hour); err != nil {
		t.Fatalf("create session: %v", err)
	}
	view, err := store.SessionViewByToken(ctx, sessionToken, now)
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if !view.Authenticated || view.User == nil || view.User.Email != identity.Email {
		t.Fatalf("bad session view: %#v", view)
	}
	if len(view.Accounts) != 1 || len(view.Teams) != 1 || len(view.Projects) != 1 {
		t.Fatalf("missing starter workspace: accounts=%d teams=%d projects=%d", len(view.Accounts), len(view.Teams), len(view.Projects))
	}
	if view.Roles[view.Projects[0].ID] != string(ProjectRoleAdmin) {
		t.Fatalf("project role = %q, want admin", view.Roles[view.Projects[0].ID])
	}
}
