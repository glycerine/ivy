package db

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestDefaultMigrationsAreOrderedAndUnique(t *testing.T) {
	migrations := DefaultMigrations()
	if len(migrations) == 0 {
		t.Fatal("expected migrations")
	}
	seen := map[int]bool{}
	last := 0
	for _, migration := range migrations {
		if migration.Version <= last {
			t.Fatalf("migration %q is out of order", migration.Name)
		}
		if seen[migration.Version] {
			t.Fatalf("duplicate migration version %d", migration.Version)
		}
		if strings.TrimSpace(migration.SQL) == "" {
			t.Fatalf("migration %q has empty SQL", migration.Name)
		}
		seen[migration.Version] = true
		last = migration.Version
	}
}

func TestInitialMigrationIncludesAuthAndProjectTables(t *testing.T) {
	sql := DefaultMigrations()[0].SQL
	for _, needle := range []string{
		"create table if not exists users",
		"create table if not exists accounts",
		"create table if not exists projects",
		"create table if not exists opaque_records",
		"create table if not exists passkey_credentials",
		"create schema if not exists project_data",
	} {
		if !strings.Contains(sql, needle) {
			t.Fatalf("initial migration missing %q", needle)
		}
	}
}

func TestMigratorAppliesIdempotentlyWithPostgres(t *testing.T) {
	dsn := os.Getenv("IVYSVK_TEST_DB_DSN")
	if dsn == "" {
		t.Skip("set IVYSVK_TEST_DB_DSN for migration integration test")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)

	migrator := DefaultMigrator()
	if err := migrator.Apply(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if err := migrator.Apply(ctx, conn); err != nil {
		t.Fatal(err)
	}
}
