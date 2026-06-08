package control

import (
	"context"
	"testing"
	"time"
)

func TestPostgresStoreMapsOIDCUserSessionAndStarterWorkspace(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()
	subject, err := NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	identity := OIDCIdentity{
		Issuer:        "email",
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
	view, err := store.SessionViewByToken(ctx, sessionToken, now, AppSessionTTL)
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

func TestPostgresStoreRecordsVisitingHoursOncePerHourAndRefreshesDaily(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()
	subject, err := NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	identity := OIDCIdentity{
		Issuer:        "email",
		Subject:       subject,
		Email:         "visits+" + subject[:8] + "@example.test",
		DisplayName:   "Visit Tester",
		EmailVerified: true,
	}
	user, err := store.UpsertUserFromOIDC(ctx, identity)
	if err != nil {
		t.Fatalf("upsert user: %v", err)
	}
	sessionToken, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	csrfToken, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 5, 7, 8, 10, 0, 0, time.UTC)
	if err := store.CreateAppSession(ctx, user.ID, sessionToken, csrfToken, start, AppSessionTTL, AppSessionTTL); err != nil {
		t.Fatalf("create session: %v", err)
	}

	first, err := store.TouchSessionByToken(ctx, sessionToken, start.Add(10*time.Minute), AppSessionTTL)
	if err != nil {
		t.Fatalf("first touch: %v", err)
	}
	if !first.VisitHourInserted || first.CookieRefreshNeeded {
		t.Fatalf("first touch = %#v, want inserted without daily cookie refresh", first)
	}
	second, err := store.TouchSessionByToken(ctx, sessionToken, start.Add(20*time.Minute), AppSessionTTL)
	if err != nil {
		t.Fatalf("second touch: %v", err)
	}
	if second.VisitHourInserted || second.CookieRefreshNeeded {
		t.Fatalf("second touch = %#v, want no duplicate same-hour row", second)
	}
	nextHour, err := store.TouchSessionByToken(ctx, sessionToken, start.Add(time.Hour+5*time.Minute), AppSessionTTL)
	if err != nil {
		t.Fatalf("next-hour touch: %v", err)
	}
	if !nextHour.VisitHourInserted || nextHour.CookieRefreshNeeded {
		t.Fatalf("next-hour touch = %#v, want hourly row without daily refresh", nextHour)
	}
	nextDay, err := store.TouchSessionByToken(ctx, sessionToken, start.Add(24*time.Hour+5*time.Minute), AppSessionTTL)
	if err != nil {
		t.Fatalf("next-day touch: %v", err)
	}
	if !nextDay.VisitHourInserted || !nextDay.CookieRefreshNeeded {
		t.Fatalf("next-day touch = %#v, want hourly row and daily cookie refresh", nextDay)
	}

	var rows int
	if err := store.db.QueryRowContext(ctx, `SELECT count(*) FROM visiting_hours WHERE user_id = $1`, user.ID).Scan(&rows); err != nil {
		t.Fatalf("count visiting_hours: %v", err)
	}
	if rows != 3 {
		t.Fatalf("visiting_hours rows = %d, want 3", rows)
	}
}

func TestPostgresStoreConsumesEmailLoginToken(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()
	token, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	email := "magic+" + token[:8] + "@example.test"
	email, err = NormalizeEmail(email)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateEmailLoginToken(ctx, email, token, now, EmailLoginTokenTTL); err != nil {
		t.Fatalf("create email login token: %v", err)
	}
	user, err := store.ConsumeEmailLoginToken(ctx, token, now)
	if err != nil {
		t.Fatalf("consume email login token: %v", err)
	}
	if user.Email != email || user.EmailVerifiedAt == nil {
		t.Fatalf("bad user from email token: %#v", user)
	}
	if _, err := store.ConsumeEmailLoginToken(ctx, token, now); err == nil {
		t.Fatalf("second consume succeeded, want error")
	}
}

func TestPostgresStorePasskeyCredentialMarksSessionViewRegistered(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()
	subject, err := NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.UpsertUserFromOIDC(ctx, OIDCIdentity{
		Issuer:        "email",
		Subject:       subject,
		Email:         "passkey-view+" + subject[:8] + "@example.test",
		DisplayName:   "Passkey View",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("upsert user: %v", err)
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
	if err := store.CreateAppSession(ctx, user.ID, sessionToken, csrfToken, now, AppSessionTTL, AppSessionTTL); err != nil {
		t.Fatalf("create session: %v", err)
	}
	initial, err := store.SessionViewByToken(ctx, sessionToken, now, AppSessionTTL)
	if err != nil {
		t.Fatalf("load initial session: %v", err)
	}
	if initial.Passkey != nil && initial.Passkey.Registered {
		t.Fatalf("passkey registered before credential insert: %#v", initial.Passkey)
	}
	if err := store.CreatePasskeyCredential(ctx, StoredPasskeyCredential{
		UserID:          user.ID,
		CredentialID:    []byte("credential-" + subject[:8]),
		PublicKeyCOSE:   []byte{0xa1, 0x01, 0x02},
		SignCount:       7,
		Transports:      []string{"internal"},
		AttestationType: "none",
		DisplayName:     user.Email,
	}); err != nil {
		t.Fatalf("create passkey credential: %v", err)
	}
	view, err := store.SessionViewByToken(ctx, sessionToken, now, AppSessionTTL)
	if err != nil {
		t.Fatalf("load session view: %v", err)
	}
	if view.Passkey == nil || !view.Passkey.Registered {
		t.Fatalf("passkey registered = false, want true")
	}
	credential, err := store.PasskeyCredentialByID(ctx, []byte("credential-"+subject[:8]))
	if err != nil {
		t.Fatalf("load passkey credential: %v", err)
	}
	if credential.UserID != user.ID || credential.SignCount != 7 {
		t.Fatalf("bad credential: %#v", credential)
	}
}
