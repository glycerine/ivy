# Control-Plane PostgreSQL Bootstrap

These scripts prepare the local PostgreSQL database for the Ivy control-plane
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

The control-plane application owns `ivyvue`.

The `bootstrap` target runs:

```sh
psql 'postgres://jaten:jaten@127.0.0.1:5432/postgres?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/00_bootstrap_databases.sql
psql 'postgres://jaten:jaten@127.0.0.1:5432/ivyvue?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/01_ivyvue_roles_and_schemas.sql
psql 'postgres://ivyvue_migrator:ivyvue_migrator_dev@127.0.0.1:5432/ivyvue?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/02_control_schema.sql
psql 'postgres://ivyvue_migrator:ivyvue_migrator_dev@127.0.0.1:5432/ivyvue?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/03_rls_template.sql
psql 'postgres://jaten:jaten@127.0.0.1:5432/ivyvue?sslmode=disable' -v ON_ERROR_STOP=1 -f control/admin/postgres/04_verify_bootstrap.sql
```

## Email Delivery

Tests use `MemoryEmailSender`; no real email is sent unless Mailgun is
explicitly configured.

For local manual testing with your Mailgun DNS/API setup:

```sh
cd goivy
IVY_CONTROL_DATABASE_DSN='postgres://ivyvue_app:ivyvue_app_dev@127.0.0.1:5432/ivyvue?sslmode=disable' \
IVY_CONTROL_MAILGUN_DOMAIN='mg.fencebunt.com' \
IVY_CONTROL_MAILGUN_FROM='Ivy <postmaster@mg.fencebunt.com>' \
IVY_CONTROL_MAILGUN_API_KEY="$MAILGIN_FENCEBUNT_SIGNUP_API_KEY" \
go run ./cmd/ivy-control
```

`MAILGIN_FENCEBUNT_SIGNUP_API_KEY` is also accepted directly for compatibility
with the existing local Mailgun test program.

## Alpha Tester Seed Workflow

After `make bootstrap`, the control admin command can seed a tester, billing
account, team, project, and project grant in `ivyvue`:

```sh
cd goivy
go run ./cmd/ivy-control-admin seed-alpha \
  -email tester1@example.test \
  -display-name 'Tester One'
```

The command is idempotent for the same email/account/project combination.
