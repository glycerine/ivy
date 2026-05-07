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
psql 'postgres://jaten:jaten@127.0.0.1:5432/postgres?sslmode=disable' -f goivy/control/admin/postgres/00_bootstrap_databases.sql
psql 'postgres://jaten:jaten@127.0.0.1:5432/ivyvue?sslmode=disable' -f goivy/control/admin/postgres/01_ivyvue_roles_and_schemas.sql
psql 'postgres://ivyvue_migrator:ivyvue_migrator_dev@127.0.0.1:5432/ivyvue?sslmode=disable' -f goivy/control/admin/postgres/02_control_schema.sql
psql 'postgres://ivyvue_migrator:ivyvue_migrator_dev@127.0.0.1:5432/ivyvue?sslmode=disable' -f goivy/control/admin/postgres/03_rls_template.sql
psql 'postgres://jaten:jaten@127.0.0.1:5432/ivyvue?sslmode=disable' -f goivy/control/admin/postgres/04_verify_bootstrap.sql
```

Casdoor owns the `casdoor` database. The control-plane application owns
`ivyvue`.
