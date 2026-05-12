package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

type Migrator struct {
	Migrations []Migration
}

func DefaultMigrator() Migrator {
	return Migrator{Migrations: DefaultMigrations()}
}

func DefaultMigrations() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "auth_project_foundation",
			SQL: `
create extension if not exists pgcrypto;

create table if not exists users (
	id uuid primary key default gen_random_uuid(),
	primary_email text not null unique,
	display_name text not null,
	created_at timestamptz not null default now()
);

create table if not exists accounts (
	id uuid primary key default gen_random_uuid(),
	name text not null,
	slug text not null unique,
	kind text not null check (kind in ('personal', 'corporate')),
	billing_status text not null default 'trial',
	created_at timestamptz not null default now()
);

create table if not exists account_users (
	account_id uuid not null references accounts(id) on delete cascade,
	user_id uuid not null references users(id) on delete cascade,
	role text not null default 'owner',
	created_at timestamptz not null default now(),
	primary key (account_id, user_id)
);

create table if not exists teams (
	id uuid primary key default gen_random_uuid(),
	account_id uuid not null references accounts(id) on delete cascade,
	name text not null,
	slug text not null,
	created_at timestamptz not null default now(),
	unique (account_id, slug)
);

create table if not exists team_members (
	team_id uuid not null references teams(id) on delete cascade,
	user_id uuid not null references users(id) on delete cascade,
	role text not null default 'member',
	created_at timestamptz not null default now(),
	primary key (team_id, user_id)
);

create table if not exists projects (
	id uuid primary key default gen_random_uuid(),
	account_id uuid not null references accounts(id) on delete cascade,
	owner_kind text not null check (owner_kind in ('user', 'team', 'account')),
	owner_id uuid not null,
	name text not null,
	slug text not null,
	storage_mode text not null default 'shared_postgres',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (account_id, slug)
);

create table if not exists project_grants (
	id uuid primary key default gen_random_uuid(),
	project_id uuid not null references projects(id) on delete cascade,
	subject_kind text not null check (subject_kind in ('user', 'team', 'account_users')),
	subject_id uuid not null,
	role text not null check (role in ('read', 'write', 'admin', 'owner')),
	created_at timestamptz not null default now()
);

create table if not exists auth_identities (
	id uuid primary key default gen_random_uuid(),
	user_id uuid not null references users(id) on delete cascade,
	provider text not null,
	provider_subject text not null,
	created_at timestamptz not null default now(),
	unique (provider, provider_subject)
);

create table if not exists opaque_records (
	user_id uuid primary key references users(id) on delete cascade,
	envelope bytea not null,
	server_public_key bytea not null,
	updated_at timestamptz not null default now()
);

create table if not exists passkey_credentials (
	id bytea primary key,
	user_id uuid not null references users(id) on delete cascade,
	public_key bytea not null,
	sign_count bigint not null default 0,
	transports text[] not null default '{}',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create table if not exists sessions (
	id uuid primary key default gen_random_uuid(),
	user_id uuid not null references users(id) on delete cascade,
	token_hash bytea not null unique,
	csrf_token_hash bytea not null,
	expires_at timestamptz not null,
	created_at timestamptz not null default now()
);

create table if not exists magic_links (
	id uuid primary key default gen_random_uuid(),
	user_id uuid references users(id) on delete cascade,
	email text not null,
	token_hash bytea not null unique,
	purpose text not null,
	expires_at timestamptz not null,
	consumed_at timestamptz,
	created_at timestamptz not null default now()
);

create table if not exists oauth_states (
	id uuid primary key default gen_random_uuid(),
	provider text not null,
	state_hash bytea not null unique,
	nonce_hash bytea not null,
	expires_at timestamptz not null,
	created_at timestamptz not null default now()
);

create schema if not exists project_data;

create table if not exists project_data.models (
	id uuid primary key default gen_random_uuid(),
	project_id uuid not null references projects(id) on delete cascade,
	filename text not null,
	body text not null,
	revision bigint not null default 1,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (project_id, filename)
);
`,
		},
	}
}

func (m Migrator) Apply(ctx context.Context, conn *pgx.Conn) error {
	if conn == nil {
		return errors.New("nil pgx connection")
	}
	if err := ensureMigrationTable(ctx, conn); err != nil {
		return err
	}
	for _, migration := range m.Migrations {
		if err := applyOne(ctx, conn, migration); err != nil {
			return err
		}
	}
	return nil
}

func ensureMigrationTable(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `create table if not exists svk_schema_migrations (
	version integer primary key,
	name text not null,
	applied_at timestamptz not null default now()
)`)
	return err
}

func applyOne(ctx context.Context, conn *pgx.Conn, migration Migration) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx, `select exists(select 1 from svk_schema_migrations where version=$1)`, migration.Version).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return tx.Commit(ctx)
	}
	if _, err := tx.Exec(ctx, migration.SQL); err != nil {
		return fmt.Errorf("apply migration %d %s: %w", migration.Version, migration.Name, err)
	}
	if _, err := tx.Exec(ctx, `insert into svk_schema_migrations(version, name) values($1, $2)`, migration.Version, migration.Name); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
