# Webvue Phase 0 Plan: Auth And Project Foundation

## Goal

Create `goivy/webvue` as an auth-first, project-scoped Vue 3 application.

Phase 0 no longer starts with an anonymous workspace shell. It starts with the
identity, session, project, and authorization rails that every later workspace
feature must use.

Phase 0 is successful when:

- unauthenticated users see a login screen
- authenticated users see only billing accounts, teams, and projects they can
  access
- the workspace shell is rendered only after login and project selection
- all API calls are project scoped
- the server rejects cross-project and unauthenticated API access
- source guards prevent controller-style frontend architecture from entering
  `webvue`

## Security References

Use these as baseline guidance while implementing:

- OWASP Authentication Cheat Sheet:
  https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- OWASP Session Management Cheat Sheet:
  https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
- OWASP CSRF Prevention Cheat Sheet:
  https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
- OWASP Password Storage Cheat Sheet:
  https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html

## Hard Boundaries

- Do not edit `goivy/webui`.
- Do not import or copy old frontend runtime/controller/service code.
- Do not create any central runtime/controller object.
- Do not add `window.ivyApp`, `window.__ivyVueBridge`, or diagnostics that
  expose an app/service container.
- Do not add anonymous Ivy backend sessions.
- Do not allow a client-provided account id, team id, or project id to grant
  access by itself.
- Do not let frontend stores call the backend directly.
- Copying static files is allowed.
- Copying narrow backend-interface facts is allowed:
  - Go API route paths
  - request/response shapes
  - small HTTP fetch mechanics
  - static graph style constants later, when graph work begins

## Auth Decisions For Phase 0

### Authentication Method

Implement an internal auth boundary now, with provider adapters:

- `PasswordAuthProvider` for local/dev/test login.
- `OIDCAuthProvider` interface only, not implemented in Phase 0.

The rest of the app must not care which provider authenticated the user.

For local password auth:

- store password hashes only
- use Argon2id from day one
- persist users, credentials, billing accounts, teams, project grants, and
  sessions in PostgreSQL
- no plaintext passwords outside tests or local seed config

### Persistence Backend

Use PostgreSQL through Go's `database/sql` package from Phase 0. This is not a
later adapter. Prefer the `github.com/jackc/pgx/v5/stdlib` driver so the app
still programs to `database/sql` while using the actively maintained pgx
PostgreSQL implementation.

Use one PostgreSQL application database by default:

1. `ivyvue` is the application database.
2. control-plane tables store identity, billing accounts, teams, projects,
   grants, sessions, and storage metadata.
3. project-owned tables include `project_id`.
4. PostgreSQL row-level security protects project-owned rows.

Do not make the storage isolation unit a login user or a billing account. A
billing account can pay for many standalone users, many teams, and many
projects. The project is the unit of application data isolation.

Use a billing-account collaboration model:

- a `User` is a human login identity
- an `Account` is the billing and payment responsibility
- an account can pay for standalone users
- an account can pay for multiple teams
- a `Team` is a collaboration group inside an account
- users can be standalone account users, team members, or both
- projects belong to a billing account
- projects can grant `read`, `write`, or `admin` access to users, teams, or all
  active users in the billing account
- project data is scoped by `project_id` and protected by RLS

Implementation choices:

- database: `ivyvue`
- local bootstrap PostgreSQL role: `jaten`
- local bootstrap PostgreSQL password: `jaten`
- runtime database role: `ivyvue_app`
- migration database role: `ivyvue_migrator`
- optional administrative database role for local setup: `ivyvue_admin`
- driver: `github.com/jackc/pgx/v5/stdlib`
- production and CI should pass a DSN by flag or environment variable
- local development may default to the installed PostgreSQL instance when no DSN
  is supplied

Local development DSN shape:

```text
postgres://jaten:jaten@127.0.0.1:5432/postgres?sslmode=disable
postgres://ivyvue_app:ivyvue_app_dev@127.0.0.1:5432/ivyvue?sslmode=disable
```

The `jaten`/`jaten` PostgreSQL credential is a local bootstrap credential only.
Use it to create the `ivyvue` database and application roles on this machine.
Do not use it as the web server runtime credential. Do not hard-code production
credentials. Any local default passwords in scripts are development-only
bootstrap values for this workspace.

Command/config inputs:

- `-db-dsn`
- `IVYVUE_DB_DSN`
- `IVYVUE_MIGRATION_DSN`
- `IVYVUE_ADMIN_DSN`
- `IVYVUE_BOOTSTRAP_DSN`
- optional split config later if useful:
  - `IVYVUE_DB_HOST`
  - `IVYVUE_DB_NAME`
  - `IVYVUE_DB_USER`
  - `IVYVUE_DB_PASSWORD`

The server should open the database once at startup, verify connectivity with
`PingContext`, and pass a narrow auth/session store interface into handlers.

The PostgreSQL database stores:

- users
- password credentials
- auth sessions
- billing accounts
- account users and billing/admin roles
- teams
- team memberships
- projects
- project access grants
- project storage-location metadata
- mapping from web workspace sessions to underlying Ivy backend sessions
- project-owned Ivy data tables, each scoped by `project_id`

Project-owned tables must enable and force RLS:

```sql
ALTER TABLE project_data.example ENABLE ROW LEVEL SECURITY;
ALTER TABLE project_data.example FORCE ROW LEVEL SECURITY;
```

Project-scoped operations must run inside an explicit transaction. After the
server verifies effective project access, it sets transaction-local context:

```text
ivy.user_id
ivy.project_id
ivy.project_role
```

Use `SELECT set_config(name, value, true)` inside the transaction so connection
pool reuse cannot leak one request's project context into the next request.
RLS policies read these settings with `current_setting(..., true)`.

Keep a storage-location abstraction from day one:

```text
mode: shared_postgres
database_name: ivyvue
schema_name: project_data
```

This lets us add dedicated project databases later for large customers or
stronger isolation without changing the billing/account/team/project
authorization model.

For production, prefer a narrow runtime database role that cannot create
schemas, create roles, disable RLS, or own project data tables. Use a separate
migration/admin path for schema changes, billing-account bootstrap, user
invites, team management, and project grants.

### Administrative Setup

Because a fresh PostgreSQL install only starts the server, Phase 0 needs
administrative setup scripts and an admin command path.

Create:

```text
goivy/webvue/admin/
  README.md
  postgres/
    00_bootstrap_database.sql
    01_roles_and_grants.sql
    02_verify_bootstrap.sql

goivy/cmd/ivywebvue-admin/
  ivywebvue-admin.go
```

The SQL scripts are run with `psql` by a local PostgreSQL superuser or a role
with database/role creation privileges. They should:

- use the local bootstrap role `jaten`/`jaten` during development
- create the `ivyvue` database if it does not exist
- create development roles:
  - `ivyvue_admin`
  - `ivyvue_migrator`
  - `ivyvue_app`
- grant runtime privileges only to `ivyvue_app`
- grant schema/migration privileges only to `ivyvue_migrator`
- revoke broad `PUBLIC` privileges from application schemas
- leave production passwords out of the repository

The Go admin command should use `database/sql` too. It should provide:

```text
ivywebvue-admin migrate
ivywebvue-admin seed-dev
ivywebvue-admin create-user
ivywebvue-admin create-account
ivywebvue-admin add-account-user
ivywebvue-admin create-team
ivywebvue-admin add-team-user
ivywebvue-admin create-project
ivywebvue-admin grant-project
ivywebvue-admin revoke-project
ivywebvue-admin list-users
ivywebvue-admin list-accounts
ivywebvue-admin list-teams
ivywebvue-admin list-projects
```

Password hashing belongs in the Go admin command, not raw SQL seed files, so
Argon2id parameters stay in one implementation.

### Database Schema

Phase 0 owns the schema needed for auth, billing accounts, account users,
teams, team memberships, project access, and mapping project-scoped web
sessions to underlying Ivy backend sessions.

Use PostgreSQL schemas to keep intent clear:

- `control` for users, accounts, teams, projects, grants, sessions, and
  migrations
- `project_data` for project-owned Ivy data tables protected by RLS
- `app_private` for helper functions that should not be called directly by the
  runtime app role

```text
control.schema_migrations
  version
  applied_at

control.users
  id
  email
  display_name
  disabled_at
  created_at
  updated_at

control.password_credentials
  user_id
  password_hash
  password_hash_params
  updated_at

control.accounts
  id
  slug
  display_name
  billing_email
  billing_status
  disabled_at
  created_at
  updated_at

control.account_users
  account_id
  user_id
  role
  seat_state
  disabled_at
  created_at
  updated_at

control.teams
  id
  account_id
  slug
  display_name
  disabled_at
  created_at
  updated_at

control.team_memberships
  team_id
  user_id
  role
  disabled_at
  created_at
  updated_at

control.projects
  id
  account_id
  slug
  display_name
  created_by_user_id
  disabled_at
  created_at
  updated_at

control.project_storage_locations
  project_id
  mode
  database_name
  schema_name
  state
  created_at
  updated_at

control.project_grants
  project_id
  subject_kind
  subject_id
  role
  disabled_at
  created_at
  updated_at

control.auth_sessions
  id_hash
  user_id
  csrf_token_hash
  created_at
  last_seen_at
  idle_expires_at
  absolute_expires_at
  revoked_at

control.ivy_workspace_sessions
  id
  project_id
  user_id
  backend_session_id
  created_at
  last_seen_at
  closed_at
```

Store session ids and CSRF tokens as hashes. Return only the raw opaque values
to the browser cookie/header path at creation time.

Every future table in `project_data` must include `project_id`, enable RLS, and
have `USING` and `WITH CHECK` policies based on the transaction-local
`ivy.project_id` and `ivy.project_role` settings.

### Migrations And Seed Data

Create idempotent SQL migrations under `goivy/webvue/auth/migrations`:

```text
001_control_schema.sql
002_auth_sessions.sql
003_project_storage.sql
004_rls_helpers.sql
```

The dev seed is enabled only for local development and tests, and is performed
by `ivywebvue-admin seed-dev` so password hashing goes through Go. It creates:

```text
email: dev@local
password: dev-password
billing account: dev
account user role: owner
team: dev/core
team role: owner
project: dev/client-server
project storage: shared_postgres in project_data
project grant: dev@local admin
project grant: dev/core admin
```

Keep PostgreSQL roles separate from web application accounts.
`dev@local`/`dev-password` is the local web application login, not a PostgreSQL
role.

### Browser Session

Use server-side sessions with cookies:

- cookie name: `ivy_webvue_session`
- cookie flags:
  - `HttpOnly`
  - `SameSite=Lax`
  - `Secure` when served over HTTPS
  - no broad `Domain` attribute
  - `Path=/`
- session id generated with cryptographically secure random bytes
- session id is opaque and stores no user/account/project data client-side
- rotate session id on login
- clear session on logout
- enforce idle and absolute expiration in server-side session records

### CSRF

Use a CSRF token for state-changing same-origin API requests:

- server stores CSRF token with the auth session
- frontend reads token from `/auth/me` or `/auth/csrf`
- frontend sends `X-CSRF-Token` on `POST`, `PUT`, `PATCH`, `DELETE`
- server rejects missing or mismatched CSRF tokens
- keep `SameSite=Lax` as defense in depth, not the only defense

### Account, Team, And Project Model

Define project-scoped identity data from day one:

```text
User
  id
  email
  displayName
  disabledAt

Account
  id
  slug
  displayName
  billingEmail
  billingStatus
  disabledAt

AccountUser
  accountId
  userId
  role
  seatState
  disabledAt

Team
  id
  accountId
  slug
  displayName
  disabledAt

TeamMembership
  teamId
  userId
  role
  disabledAt

Project
  id
  accountId
  slug
  displayName
  createdByUserId
  disabledAt

ProjectStorageLocation
  projectId
  mode
  databaseName
  schemaName
  state

ProjectGrant
  projectId
  subjectKind
  subjectId
  role
  disabledAt

AuthSession
  id
  userId
  csrfToken
  createdAt
  lastSeenAt
  idleExpiresAt
  absoluteExpiresAt
  revokedAt

IvyWorkspaceSession
  id
  projectId
  userId
  backendSessionId
  createdAt
  lastSeenAt
```

Account roles for Phase 0:

- `owner`
- `admin`
- `billing_admin`
- `member`

Account roles describe billing-account administration and seat management.
`owner` and `admin` can manage account users, teams, projects, and billing.
`billing_admin` can manage billing but does not receive project access by
default. `member` is a paid/covered user with no project access unless granted
directly, through a team, or through an account-wide project grant.

Team roles for Phase 0:

- `owner`
- `admin`
- `member`

Team roles describe team administration. Team membership does not grant project
access unless the team has a project grant.

Project roles for Phase 0:

- `admin`
- `write`
- `read`

Access rules:

1. A project belongs to exactly one billing account.
2. Account `owner` and `admin` users have `admin` access to projects in that
   account.
3. Account `billing_admin` users have billing access only and no project access
   by default.
4. Account `member` users have no project access by default.
5. A direct user project grant contributes that role.
6. A team project grant contributes that role to active members of the team.
7. An account project grant contributes that role to active account users.
8. Effective project access is the highest role found across account admin
   status, direct user grants, team grants, and account-wide grants.
9. `read` can view/load project state.
10. `write` can change models, workspace state, and run mutating Ivy actions.
11. `admin` can manage project settings and grants.

## Project-Scoped API Shape

Avoid the old anonymous route shape for new frontend code. Use project-scoped
routes:

```text
GET  /auth/me
POST /auth/login
POST /auth/logout

GET  /api/accounts
GET  /api/accounts/{accountID}/users
GET  /api/accounts/{accountID}/teams
GET  /api/projects
GET  /api/accounts/{accountID}/projects

POST /api/projects/{projectID}/ivy/session/new
POST /api/projects/{projectID}/ivy/session/{workspaceSessionID}/load
GET  /api/projects/{projectID}/ivy/session/{workspaceSessionID}/arg
GET  /api/projects/{projectID}/ivy/session/{workspaceSessionID}/concept
GET  /api/projects/{projectID}/ivy/session/{workspaceSessionID}/menus
POST /api/projects/{projectID}/ivy/session/{workspaceSessionID}/check
POST /api/projects/{projectID}/ivy/session/{workspaceSessionID}/action
POST /api/projects/{projectID}/ivy/session/{workspaceSessionID}/arg/action
GET  /api/projects/{projectID}/ivy/session/{workspaceSessionID}/events
```

Server checks on every project route:

1. valid auth session cookie
2. active user
3. active project
4. active billing account
5. effective project role satisfies the route requirement
6. workspace session belongs to the same project and user unless role permits
   broader access
7. valid CSRF token for state-changing methods

The old `goivy/webui` backend can still be reused behind this boundary. The new
server maps a project-scoped `workspaceSessionID` to the underlying backend
session id.

## Phase 0 Deliverables

### Directory Structure

Create this initial shape:

```text
goivy/webvue/
  README.md
  groundup_vue3.md
  phase0.md
  postgres_admin.md
  auth/
    model.go
    store.go
    postgres_store.go
    migrate.go
    password.go
    migrations/
      001_control_schema.sql
      002_auth_sessions.sql
      003_project_storage.sql
      004_rls_helpers.sql
    session.go
    middleware.go
  admin/
    README.md
    postgres/
      00_bootstrap_database.sql
      01_roles_and_grants.sql
      02_verify_bootstrap.sql
  server.go
  server_test.go
  frontend/
    package.json
    vite.config.mjs
    vitest.config.mjs
    src/
      main.ts
      App.vue
      app/
        bootstrap.ts
        injectionKeys.ts
      api/
        webvueHttpClient.ts
        authApi.ts
      components/
        AuthShell.vue
        LoginForm.vue
        ProjectPicker.vue
        WorkspaceShell.vue
        panes/
          ArgPane.vue
          ConceptPane.vue
          StateRelationsPane.vue
          EditorPane.vue
      stores/
        authStore.ts
        accountStore.ts
        projectStore.ts
        layoutStore.ts
        sessionStore.ts
      test/
        mount.ts
      architecture/
        sourceGuards.test.ts
      App.test.ts
  static/
    index.html
    css/
      webvue.css
    dist/
    tutorial/
    favicon files
  pw_test/
    auth_smoke.spec.mjs
  playwright.config.mjs

goivy/cmd/ivywebvue/
  ivywebvue.go

goivy/cmd/ivywebvue-admin/
  ivywebvue-admin.go
```

Use TypeScript from the beginning for `webvue` frontend code.

### Frontend Build Setup

Add `frontend/package.json` scripts:

```json
{
  "build": "../../../node_modules/.bin/vite build --config vite.config.mjs",
  "test": "../../../node_modules/.bin/vitest run --config vitest.config.mjs"
}
```

Use Vite only for manually initiated bundling. No Vite dev server. Output:

```text
goivy/webvue/static/dist/ivywebvue.js
goivy/webvue/static/dist/ivywebvue.js.map
```

Use cache directories under repo-local `.cache` or `node_modules`, never
`/private/tmp`.

### Static Assets

Copy these from `goivy/webui/static`:

- favicon files
- `site.webmanifest`
- tutorial tree

Do not copy the old `static/index.html` wholesale. Write a new minimal
`webvue/static/index.html` that loads:

- CodeMirror CSS/JS from the same CDN URLs for now
- Cytoscape/Dagre from the same CDN URLs for now
- `/static/css/webvue.css`
- `/static/dist/ivywebvue.js`

Later phases may vendor these dependencies, but Phase 0 should not spend time
on asset packaging.

### Go Serving Strategy

Create:

```text
goivy/webvue/server.go
goivy/cmd/ivywebvue/ivywebvue.go
```

The `webvue` server should:

- serve `goivy/webvue/static/index.html` at `/`
- serve `goivy/webvue/static/*` under `/static/`
- expose `/auth/*`
- expose account/project APIs under `/api/accounts`, `/api/projects`, and
  project-scoped `/api/projects/{projectID}/*`
- reuse the existing Go backend implementation from `goivy/webui` behind the
  project boundary by importing backend interfaces/types from
  `github.com/glycerine/ivy/goivy/webui`

This reuse is acceptable because it is the Go/backend contract, not frontend
architecture. Do not copy frontend runtime or browser code.

Command behavior for tests:

```sh
go run ./goivy/cmd/ivywebvue -addr 127.0.0.1:18090
```

Phase 0 command flags:

- `-addr`
- `-db-dsn`
- `-migrate` defaulting to false in production and true only for explicit local
  development commands

Administrative setup uses `goivy/cmd/ivywebvue-admin`, not the web server.

Seeded dev identity:

```text
email: dev@local
password: dev-password
billing account: dev
account user role: owner
team: dev/core
team role: owner
project: dev/client-server
project grant: dev@local admin
project grant: dev/core admin
```

This must be clearly marked as development-only and easy to disable.

Default database configuration for local development:

```text
database: ivyvue
database role: ivyvue_app
database password: ivyvue_app_dev
host: 127.0.0.1:5432
sslmode: disable
```

If `-db-dsn` or `IVYVUE_DB_DSN` is set, it wins over the local defaults.

### Initial Vue Shell

Render by auth state:

- loading state while `/auth/me` resolves
- login screen when unauthenticated
- project picker when authenticated with multiple projects and no active
  project
- workspace shell when authenticated with an active project

The initial workspace shell should be real, not a landing page:

- top menubar placeholder
- account/team/project/user indicator
- logout control
- four visible pane headers:
  - `ARG (Abstract Reachability Graph)`
  - `Concept graph`
  - `State/relations`
  - `Editing:`
- placeholder graph mount regions
- placeholder editor region
- placeholder details strip
- status/session strip

No marketing copy. No explanatory in-app tutorial text.

### Initial Stores

Create only the stores needed for Phase 0:

- `authStore`
  - user
  - csrf token
  - authenticated/loading/error state
- `accountStore`
  - accounts visible to the user
  - account user roles
  - active account id
- `teamStore`
  - teams visible to the user
  - team memberships
  - active team id when useful for filtering
- `projectStore`
  - projects visible to the user
  - active project id
  - effective project role
- `layoutStore`
  - default column widths
  - tutorial visible flag
  - details height
- `sessionStore`
  - future Ivy workspace session id
  - status message
  - status level

Stores must not call the backend. API modules/services perform effects.

### Frontend API Modules

Create:

- `webvueHttpClient.ts`
  - `credentials: 'same-origin'`
  - JSON request helper
  - attaches `X-CSRF-Token` for state-changing methods
  - no auth token in localStorage/sessionStorage
- `authApi.ts`
  - `me`
  - `login`
  - `logout`
  - `listAccounts`
- `teamApi.ts`
  - `listAccountTeams`
  - `listAccountUsers`
- `projectApi.ts`
  - `listProjects`
  - `listAccountProjects`

### Bootstrap

Use:

```text
src/app/bootstrap.ts
```

Responsibilities allowed in Phase 0:

- create Pinia
- mount Vue
- provide API modules/services

Responsibilities forbidden:

- workflow state
- graph/editor/session ownership
- command forwarding
- public app object

Avoid the names `runtime`, `controller`, and `appServices`.

### First Tests

Unit/component tests:

- `App.test.ts`
  - unauthenticated state renders login
  - authenticated state renders project picker or workspace
  - workspace shows all four pane headers
- auth store tests
- account store tests
- project store tests
- HTTP client tests for CSRF header behavior
- `architecture/sourceGuards.test.ts`
  - fails if production source contains forbidden architecture names

Go tests:

- unauthenticated `/auth/me` returns unauthenticated state
- seeded dev login sets session cookie
- `/auth/me` returns user, accounts, memberships, projects, and effective roles
  after login
- logout clears session
- account and project lists require auth
- project route rejects missing project access
- project route rejects insufficient role for write/admin operations
- state-changing project route rejects missing CSRF
- PostgreSQL store contract tests cover users, credentials, accounts, account
  memberships, projects, grants, sessions, revocation, expiration, and
  workspace-session ownership
- RLS tests prove project-owned rows are visible only with the matching
  transaction-local project context
- admin setup tests verify the bootstrap SQL is idempotent against a disposable
  local database when `IVYVUE_ADMIN_DSN` is set
- database tests must create unique test records and clean up only records they
  created; do not wipe `ivyvue`

Browser smoke:

- page loads
- title contains Ivy
- unauthenticated page shows login
- login with seeded dev user succeeds
- workspace shell appears
- four pane headers are visible
- logout returns to login
- no console errors
- forbidden globals do not exist

### Source Guard Ratchets

Add a Vitest source guard in Phase 0. It should scan production source and fail
on:

- `ivyRuntime`
- `legacyAppController`
- `IvyApp`
- `window.ivyApp`
- `__ivyVueBridge`
- `__ivyDiagnostics.runtime`
- `registerMethodCommands`
- `commandTarget`
- `controller`
- `appServices`
- function parameters named `app` in service files
- function parameters named `runtime`
- function parameters named `controller`
- `localStorage` or `sessionStorage` usage for auth tokens

Allow ordinary Vue `createApp` only in `main.ts` or bootstrap code.

### Verification Commands

Phase 0 should end green with:

```sh
npm --prefix goivy/webvue/frontend run test
npm --prefix goivy/webvue/frontend run build
IVYVUE_DB_DSN='postgres://ivyvue_app:ivyvue_app_dev@127.0.0.1:5432/ivyvue?sslmode=disable' go test ./goivy/webvue
go test ./goivy/cmd/ivywebvue
go test ./goivy/cmd/ivywebvue-admin
cd goivy/webvue && ../../node_modules/.bin/playwright test
```

The browser test may require sandbox escalation to bind localhost.

## Implementation Order

1. Create PostgreSQL administrative setup docs and SQL scripts.
2. Create frontend package/config files.
3. Copy allowed static assets.
4. Create new `static/index.html`.
5. Create `webvue.css`.
6. Create SQL migrations for users, credentials, accounts, account
   memberships, projects, grants, auth sessions, project storage locations, RLS
   helpers, and Ivy workspace-session mappings.
7. Create Go auth model/store/session/middleware interfaces.
8. Create PostgreSQL `database/sql` auth store.
9. Create project-scoped transaction helper that sets RLS context with
   `set_config`.
10. Create password hashing/verification helpers.
11. Create migration runner.
12. Create `goivy/cmd/ivywebvue-admin` with `migrate`, `seed-dev`, user,
    account, project, and grant commands.
13. Create `webvue` server with `/auth/*` routes.
14. Create account/project API route skeleton and project authorization
    middleware.
15. Create `goivy/cmd/ivywebvue`.
16. Create frontend auth/account/project stores.
17. Create HTTP client, auth API, and project API modules.
18. Create login, project picker, and workspace shell components.
19. Create source guard test.
20. Create frontend unit tests.
21. Create Go auth/server/store/RLS/admin tests.
22. Create Playwright config and auth smoke test.
23. Run verification commands.
24. Update `README.md` with build/test/serve/login/database/admin commands.

## Phase 0 Non-Goals

- no CodeMirror integration yet
- no Cytoscape integration yet
- no real Ivy model loading yet
- no file open/save yet
- no check induction yet
- no dynamic menus yet
- no tutorial iframe behavior yet
- no production OIDC implementation yet
- no dedicated per-project database provisioning yet
- no production-grade PostgreSQL role/password rotation yet

Those start in Phase 1 and later. Phase 0 must still make their future API
surface project-scoped.

## Stop Conditions

Stop and revisit the plan if implementation starts to introduce:

- a broad object passed into many services
- a diagnostics API exposing a service container
- copied old frontend service/controller code
- command registration by method reflection
- auth tokens in browser storage
- direct DOM fallback logic for UI that Vue should own
- anonymous access to project/Ivy session APIs
