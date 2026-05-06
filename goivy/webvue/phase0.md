# Webvue Phase 0 Plan: Auth And Project Foundation

## Goal

Create `goivy/webvue` as an auth-first, project-scoped Vue 3 application.

Phase 0 no longer starts with an anonymous workspace shell. It starts with the
identity, session, project, and authorization rails that every later workspace
feature must use.

Phase 0 is successful when:

- unauthenticated users see a login screen
- authenticated users see only accounts and projects they can access
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
- Do not allow a client-provided account id or project id to grant access by
  itself.
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
- persist users, credentials, accounts, project grants, and sessions in MariaDB
- no plaintext passwords outside tests or local seed config

### Persistence Backend

Use MariaDB through Go's `database/sql` package from Phase 0. This is not a
later adapter.

Use a two-layer persistence model:

1. `ivyvue` is the control-plane database.
2. Each project gets its own MariaDB database for project-owned Ivy data.

Do not make the database isolation unit a login user or an account. Users can
belong to many accounts, accounts can own many projects, and projects can be
shared with many users. The project is the unit of Ivy data isolation.

Use the GitHub-like ownership model:

- a `User` is a human login identity
- an `Account` is an owner namespace
- an account has kind `personal` or `team`
- every user gets a personal account
- team accounts have members
- projects belong to accounts
- projects can grant `read`, `write`, or `admin` access to users and team
  accounts
- every project gets its own MariaDB database

Implementation choices:

- control-plane database: `ivyvue`
- local development database user: `jaten`
- local development database password: `jaten`
- driver: `github.com/go-sql-driver/mysql`
- production and CI should pass a DSN by flag or environment variable
- local development may default to the installed MariaDB instance when no DSN
  is supplied

Local development DSN shape:

```text
jaten:jaten@tcp(127.0.0.1:3306)/ivyvue?parseTime=true&loc=UTC&charset=utf8mb4,utf8
```

Do not hard-code production credentials. The `jaten`/`jaten` credentials are
development-only bootstrap values for this workspace.

Command/config inputs:

- `-db-dsn`
- `IVYVUE_DB_DSN`
- optional split config later if useful:
  - `IVYVUE_DB_HOST`
  - `IVYVUE_DB_NAME`
  - `IVYVUE_DB_USER`
  - `IVYVUE_DB_PASSWORD`

The server should open the database once at startup, verify connectivity with
`PingContext`, and pass a narrow auth/session store interface into handlers.

The control-plane database stores:

- users
- password credentials
- auth sessions
- accounts
- account memberships
- projects
- project access grants
- mapping from project id to project database name
- mapping from web workspace sessions to underlying Ivy backend sessions

Project databases store project-owned application data. Phase 0 only needs to
provision and record them. Later phases can store project-local persisted
artifacts there, such as models, saved workspaces, analysis state, uploaded
files, and future collaboration records.

Project database names must be server-generated, not derived directly from
email, display name, or request input. Use an opaque identifier:

```text
ivyvue_p_<32 lowercase hex chars>
```

Database identifiers cannot be bound as SQL parameters. Any code that issues
`CREATE DATABASE`, `USE`, or cross-database SQL must use only database names
loaded from the control plane after validating them against the generated-name
pattern.

For production, prefer a narrow runtime database user plus a separate
provisioning path that has `CREATE DATABASE` and migration privileges. For
local development, the `jaten` MariaDB user can be used for both if it already
has the needed privileges.

### Database Schema

Phase 0 owns the schema needed for auth, account/project ownership, project
access, and mapping project-scoped web sessions to underlying Ivy backend
sessions.

```text
schema_migrations
  version
  applied_at

users
  id
  email
  display_name
  disabled_at
  created_at
  updated_at

password_credentials
  user_id
  password_hash
  password_hash_params
  updated_at

accounts
  id
  slug
  kind
  display_name
  personal_user_id
  disabled_at
  created_at
  updated_at

account_memberships
  account_id
  user_id
  role
  disabled_at
  created_at
  updated_at

projects
  id
  owner_account_id
  slug
  display_name
  database_name
  database_state
  database_created_at
  disabled_at
  created_at
  updated_at

project_grants
  project_id
  subject_kind
  subject_id
  role
  disabled_at
  created_at
  updated_at

auth_sessions
  id_hash
  user_id
  csrf_token_hash
  created_at
  last_seen_at
  idle_expires_at
  absolute_expires_at
  revoked_at

ivy_workspace_sessions
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

### Migrations And Seed Data

Create idempotent SQL migrations under `goivy/webvue/auth/migrations`:

```text
001_auth_schema.sql
002_project_database_tracking.sql
003_seed_dev_account.sql
```

The dev seed is enabled only for local development and tests. It creates:

```text
email: dev@local
password: dev-password
personal account: dev
project: dev/client-server
project database: ivyvue_p_<generated id>
account role: owner
project role: admin
```

Keep this separate from the MariaDB login user. `jaten`/`jaten` is the database
credential; `dev@local`/`dev-password` is the local web application login.

The seed path must also provision the dev project database and mark the project
database state as ready only after its migration succeeds.

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

### Account And Project Model

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
  kind
  displayName
  personalUserId
  disabledAt

AccountMembership
  accountId
  userId
  role
  disabledAt

Project
  id
  ownerAccountId
  slug
  displayName
  databaseName
  databaseState
  disabledAt

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
- `member`

Project roles for Phase 0:

- `admin`
- `write`
- `read`

Access rules:

1. A user always has `admin` access to projects owned by their personal account.
2. A team account `owner` or `admin` has `admin` access to projects owned by
   that team account.
3. A team account `member` has only the project access granted directly to the
   user or inherited through the team account.
4. Direct user grants override nothing; effective access is the highest role
   found across ownership, account membership, and project grants.
5. `read` can view/load project state.
6. `write` can change models, workspace state, and run mutating Ivy actions.
7. `admin` can manage project settings and grants.

## Project-Scoped API Shape

Avoid the old anonymous route shape for new frontend code. Use project-scoped
routes:

```text
GET  /auth/me
POST /auth/login
POST /auth/logout

GET  /api/accounts
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
4. active owner account
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
  auth/
    model.go
    store.go
    mariadb_store.go
    migrate.go
    password.go
    migrations/
      001_auth_schema.sql
      002_project_database_tracking.sql
      003_seed_dev_account.sql
    session.go
    middleware.go
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
- `-project-db-prefix` defaulting to `ivyvue_p_`
- `-migrate` defaulting to true for local development
- optional `-seed-dev-user` defaulting to true for local development

Seeded dev identity:

```text
email: dev@local
password: dev-password
personal account: dev
project: dev/client-server
account role: owner
project role: admin
```

This must be clearly marked as development-only and easy to disable.

Default database configuration for local development:

```text
database: ivyvue
database user: jaten
database password: jaten
host: 127.0.0.1:3306
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
- account/project/user indicator
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
  - account memberships
  - active account id
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
- MariaDB store contract tests cover users, credentials, accounts, account
  memberships, projects, grants, sessions, revocation, expiration, and
  workspace-session ownership
- project provisioning tests create a generated project database, run its
  migrations, record it in the control plane, and clean up only that generated
  database
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
IVYVUE_DB_DSN='jaten:jaten@tcp(127.0.0.1:3306)/ivyvue?parseTime=true&loc=UTC&charset=utf8mb4,utf8' go test ./goivy/webvue
go test ./goivy/cmd/ivywebvue
cd goivy/webvue && ../../node_modules/.bin/playwright test
```

The browser test may require sandbox escalation to bind localhost.

## Implementation Order

1. Copy allowed static assets.
2. Create frontend package/config files.
3. Create new `static/index.html`.
4. Create `webvue.css`.
5. Create SQL migrations for users, credentials, accounts, account
   memberships, projects, grants, auth sessions, project database tracking, and
   Ivy workspace-session mappings.
6. Create Go auth model/store/session/middleware interfaces.
7. Create MariaDB `database/sql` auth store.
8. Create project database name generator and identifier validator.
9. Create project database provisioning/migration path.
10. Create password hashing/verification helpers.
11. Create control-plane migration runner and dev seed path.
12. Create `webvue` server with `/auth/*` routes.
13. Create account/project API route skeleton and project authorization
    middleware.
14. Create `goivy/cmd/ivywebvue`.
15. Create frontend auth/account/project stores.
16. Create HTTP client, auth API, and project API modules.
17. Create login, project picker, and workspace shell components.
18. Create source guard test.
19. Create frontend unit tests.
20. Create Go auth/server/store/provisioning tests.
21. Create Playwright config and auth smoke test.
22. Run verification commands.
23. Update `README.md` with build/test/serve/login/database commands.

## Phase 0 Non-Goals

- no CodeMirror integration yet
- no Cytoscape integration yet
- no real Ivy model loading yet
- no file open/save yet
- no check induction yet
- no dynamic menus yet
- no tutorial iframe behavior yet
- no production OIDC implementation yet
- no production-grade project database credential rotation yet

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
