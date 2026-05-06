# Webvue PostgreSQL Admin Setup Plan

## Goal

Make a fresh PostgreSQL install usable for `goivy/webvue` without manual,
ad-hoc database work.

The app uses Go's `database/sql` package for both runtime and administrative
commands. SQL bootstrap scripts exist only for the work that must happen before
the application roles and database exist.

## Local Bootstrap Flow

Starting point:

- PostgreSQL is installed.
- The PostgreSQL server is running.
- A local bootstrap PostgreSQL role exists:
  - role: `jaten`
  - password: `jaten`
- No `ivyvue` database, roles, schema, or web application accounts are assumed.

Target local state:

- database: `ivyvue`
- local bootstrap PostgreSQL role: `jaten`
- runtime PostgreSQL role: `ivyvue_app`
- migration PostgreSQL role: `ivyvue_migrator`
- optional local admin PostgreSQL role: `ivyvue_admin`
- web login: `dev@local`
- personal account: `dev`
- project: `dev/client-server`
- project role for `dev@local`: `admin`

## SQL Bootstrap Scripts

Create these under `goivy/webvue/admin/postgres`:

```text
00_bootstrap_database.sql
01_roles_and_grants.sql
02_verify_bootstrap.sql
```

Run them with `psql` as the local bootstrap role `jaten`/`jaten`, or any other
local PostgreSQL role that can create databases and roles.

The `jaten`/`jaten` credential is only for local bootstrap. The web server
should run as `ivyvue_app`, and migrations should run as `ivyvue_migrator`.

`00_bootstrap_database.sql`:

- creates the `ivyvue` database if missing
- leaves existing databases alone
- must be idempotent

`01_roles_and_grants.sql`:

- creates `ivyvue_admin`, `ivyvue_migrator`, and `ivyvue_app` for local
  development if missing
- grants `CONNECT` on `ivyvue`
- grants schema/table migration privileges to `ivyvue_migrator`
- grants only runtime DML privileges to `ivyvue_app`
- revokes broad `PUBLIC` privileges from application schemas
- avoids production passwords

`02_verify_bootstrap.sql`:

- verifies the database exists
- verifies required roles exist
- verifies the runtime role can connect
- verifies unexpected broad privileges are not present

## Go Admin Command

Create:

```text
goivy/cmd/ivywebvue-admin/ivywebvue-admin.go
```

It should provide:

```text
ivywebvue-admin migrate
ivywebvue-admin seed-dev
ivywebvue-admin create-user
ivywebvue-admin create-personal-account
ivywebvue-admin create-team-account
ivywebvue-admin create-project
ivywebvue-admin grant-project
ivywebvue-admin revoke-project
ivywebvue-admin list-users
ivywebvue-admin list-projects
```

The admin command uses `database/sql` and the same store interfaces as the
server where possible.

Password hashing must happen in Go, not SQL, so `seed-dev` and `create-user`
use the same Argon2id implementation as login verification.

## RLS Setup

Runtime project queries must run in a transaction that sets:

```text
ivy.user_id
ivy.project_id
ivy.project_role
```

Use PostgreSQL transaction-local settings via `set_config(..., true)`.

Every project-owned table must:

- include `project_id`
- enable row-level security
- force row-level security
- define `USING` policies for reads
- define `WITH CHECK` policies for writes

## Development Commands

Expected local sequence after implementation:

```sh
psql 'postgres://jaten:jaten@127.0.0.1:5432/postgres?sslmode=disable' -f goivy/webvue/admin/postgres/00_bootstrap_database.sql
psql 'postgres://jaten:jaten@127.0.0.1:5432/ivyvue?sslmode=disable' -f goivy/webvue/admin/postgres/01_roles_and_grants.sql
psql 'postgres://jaten:jaten@127.0.0.1:5432/ivyvue?sslmode=disable' -f goivy/webvue/admin/postgres/02_verify_bootstrap.sql
IVYVUE_MIGRATION_DSN='postgres://ivyvue_migrator:ivyvue_migrator_dev@127.0.0.1:5432/ivyvue?sslmode=disable' go run ./goivy/cmd/ivywebvue-admin migrate
IVYVUE_MIGRATION_DSN='postgres://ivyvue_migrator:ivyvue_migrator_dev@127.0.0.1:5432/ivyvue?sslmode=disable' go run ./goivy/cmd/ivywebvue-admin seed-dev
IVYVUE_DB_DSN='postgres://ivyvue_app:ivyvue_app_dev@127.0.0.1:5432/ivyvue?sslmode=disable' go run ./goivy/cmd/ivywebvue -addr 127.0.0.1:18090
```

Local default passwords are acceptable only for local development and must not
be treated as production defaults.
