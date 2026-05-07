# Control-Plane PostgreSQL Bootstrap

These scripts prepare local PostgreSQL databases for the Casdoor-backed control
architecture.

Local bootstrap uses the development PostgreSQL role:

```text
user: jaten
password: jaten
```

The bootstrap role is not the runtime application credential.

Expected local sequence:

```sh
cd goivy
make bootstrap
```

Casdoor owns the `casdoor` database. The control-plane application owns
`ivyvue`.

The `bootstrap` target runs:

```sh
psql 'postgres://jaten:jaten@127.0.0.1:5432/postgres?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/00_bootstrap_databases.sql
psql 'postgres://jaten:jaten@127.0.0.1:5432/casdoor?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/01_casdoor_roles_and_schemas.sql
psql 'postgres://jaten:jaten@127.0.0.1:5432/ivyvue?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/01_ivyvue_roles_and_schemas.sql
psql 'postgres://ivyvue_migrator:ivyvue_migrator_dev@127.0.0.1:5432/ivyvue?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/02_control_schema.sql
psql 'postgres://ivyvue_migrator:ivyvue_migrator_dev@127.0.0.1:5432/ivyvue?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/03_rls_template.sql
psql 'postgres://jaten:jaten@127.0.0.1:5432/ivyvue?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/04_verify_bootstrap.sql
```

## Casdoor Local Config

Casdoor should use the separate `casdoor` database and `casdoor_app` role. An
example local config is in:

```text
control/admin/casdoor/app.conf.example
```

The intended local ports are:

```text
127.0.0.1:18080  ivy-control
127.0.0.1:18082  Casdoor
127.0.0.1:18081  ivyvue analysis server, later
```

The control-plane defaults match Casdoor's authorize route, and token/JWKS URLs
can be supplied explicitly:

```sh
cd goivy
IVY_CONTROL_DATABASE_DSN='postgres://ivyvue_app:ivyvue_app_dev@127.0.0.1:5432/ivyvue?sslmode=disable' \
go run ./cmd/ivy-control \
  -oidc-issuer-url http://127.0.0.1:18082 \
  -oidc-auth-url http://127.0.0.1:18082/login/oauth/authorize \
  -oidc-token-url http://127.0.0.1:18082/api/login/oauth/access_token \
  -oidc-jwks-url http://127.0.0.1:18082/.well-known/jwks \
  -oidc-userinfo-url http://127.0.0.1:18082/api/userinfo
```

For deterministic local/browser tests without a running Casdoor process, the
control server can expose an in-process OIDC issuer:

```sh
cd goivy
IVY_CONTROL_TEST_IDP=1 go run ./cmd/ivy-control
```

That mode is for tests only. Production and shared development should use the
standalone Casdoor process.

## Alpha Tester Seed Workflow

After `make bootstrap`, the control admin command can seed a tester, billing
account, team, project, and project grant in `ivyvue` without requiring Gmail or
GitHub accounts:

```sh
cd goivy
go run ./cmd/ivy-control-admin seed-alpha \
  -email tester1@example.test \
  -display-name 'Tester One'
```

The command is idempotent for the same email/account/project combination.
