# SVK Plan 000: SvelteKit Local-First Ivy Web App Architecture

## Purpose

This document is the first implementation plan for `goivy/svk`, the SvelteKit
successor to the original `goivy/webui` and the abandoned Vue spike in
`goivy/webvue`.

The goal is to put the large structural beams in place before feature work
starts:

- one modern reactive web app
- one engine-neutral data model
- one local-first offline path using Go `js/wasm` plus Z3 wasm
- one hosted path using native Go plus native Z3
- one authenticated Go server that serves public marketing/login pages and the
  private SvelteKit bundle after login
- one project/account/team authorization model that can grow into hosted
  collaboration

The old `webui` code is valuable as a behavior oracle and backend reference,
but it is too controller-heavy to copy as frontend architecture. The new Svelte
app must be built from small reactive stores, services, data contracts, and
engine adapters. No global app-controller object.

## User Decisions Already Made

- First login auto-creates:
  - a user
  - a personal billing account
  - a personal project
- `User`, `Account`, `Team`, and `Project` remain separate.
- A corporate account can pay for many teams and/or users.
- Users can be standalone or members of many teams.
- Projects can be shared in a GitHub-like ownership/collaboration model.
- OAuth identities are not automatically linked by email.
- Linking Google, Apple, GitHub, OPAQUE password, passkey, or email login identities
  requires an explicit signed-in linking flow.
- Password login is supported for devs and regular users through OPAQUE, so the
  server never sees the user's raw password.
- OPAQUE uses `github.com/bytemare/opaque` on the Go server side and
  `@cloudflare/opaque-ts` on the Svelte client side.
- Sessions use secure opaque cookies backed by Postgres.
- Go serves public marketing/login pages.
- Go serves the built SvelteKit app bundle after login.
- Email login and account recovery use Mailgun.
- Passkeys are designed into the schema now, but can be implemented after
  OAuth/OPAQUE/session foundations are stable.
- Offline app use must work for a previously logged-in user on a plane.

## Non-Goals For This First Architecture Pass

- Do not recreate `webui`'s monolithic session/controller structure in
  TypeScript.
- Do not use `wasip1`; browser Go execution is `GOOS=js GOARCH=wasm`.
- Do not port Ivy verification logic to JavaScript.
- Do not make SvelteKit responsible for authentication secrets or OAuth token
  exchanges.
- Do not put auth, project authorization, or persistence behind placeholder
  mocks that cannot evolve into production.
- Do not design the local-first path as a separate app.
- Do not make the hosted path and offline path produce different UI records.

## Architecture Summary

The app has four major layers:

```text
Go HTTP server
  public marketing/login pages
  auth endpoints
  OAuth callback endpoints
  OPAQUE password/passkey/magic-link endpoints
  project/account/team APIs
  native hosted Ivy engine API
  static serving of private SvelteKit bundle

SvelteKit SPA/PWA
  reactive state
  editor/layout/graph/components
  command intent dispatch
  local-first persistence
  sync queue

Engine adapters
  HostedGoEngine: HTTP/SSE or WebSocket to native Go/Z3
  BrowserWasmEngine: Worker messages to goivy js/wasm + z3 wasm

Persistence
  Postgres: canonical hosted auth/project data
  IndexedDB: offline projects/models/jobs/results/sync queue
```

The app state consumes a normalized engine-neutral record vocabulary:

```text
EngineSession
VerificationJob
GraphSnapshot
ConceptState
CheckResult
TraceEventSheet
CommandIntent
ModelDocument
Project
Account
Team
User
```

Browser and hosted engines may have very different internals. They must
normalize at the engine boundary before data enters Svelte state.

## Guiding Rules

1. Stores describe state.
2. Services perform effects.
3. Components render state and emit intents.
4. Engines execute verification.
5. Persistence layers store records; they do not own UI behavior.
6. Graph components own Cytoscape instances locally; global state owns graph
   snapshots, not live graph objects.
7. Every long-running operation is a `VerificationJob`, whether it runs in the
   browser or on the server.
8. Engine events are append-only facts. Stores reduce them into current state.
9. The server is the authority for authentication and hosted project access.
10. Offline access is allowed only from previously authenticated local state.

## Directory Plan

The SvelteKit project remains rooted at:

```text
goivy/svk/
```

Proposed TypeScript layout:

```text
goivy/svk/src/lib/
  api/
    authClient.ts
    projectClient.ts
    hostedEngineClient.ts
  auth/
    passkeyBrowser.ts
    csrf.ts
  engines/
    IvyEngine.ts
    hostedGoEngine.ts
    browserWasmEngine.ts
    workerProtocol.ts
  persistence/
    indexedDb.ts
    migrations.ts
    repositories.ts
    syncQueue.ts
  state/
    workspace.svelte.ts
    auth.svelte.ts
    projects.svelte.ts
    models.svelte.ts
    engines.svelte.ts
    jobs.svelte.ts
    graphs.svelte.ts
    concepts.svelte.ts
    selection.svelte.ts
    layout.svelte.ts
    dialogs.svelte.ts
    toasts.svelte.ts
  services/
    appBootstrapService.ts
    sessionService.ts
    modelFileService.ts
    engineService.ts
    commandService.ts
    graphService.ts
    conceptService.ts
    jobService.ts
    syncService.ts
  workers/
    ivyEngine.worker.ts
    goivyNodeFS.ts
    smtZ3Imports.ts
  types/
    ids.ts
    auth.ts
    projects.ts
    models.ts
    engines.ts
    jobs.ts
    graphs.ts
    concepts.ts
    events.ts
    commands.ts
    sync.ts
```

Proposed Go layout:

```text
goivy/svk/server/
  auth/
  csrf/
  db/
  mail/
  oauth/
  passkeys/
  opaque/
  projects/
  sessions/
  staticapp/
  templates/
  webengine/

goivy/cmd/ivysvk/
  main.go

goivy/cmd/ivysvk-admin/
  main.go
```

If names become awkward, use `ivywebsvk` instead of `ivysvk`, but use one
name consistently from the start.

## Core Data Types

### IDs And Revisions

All app records use string IDs. Use prefixes for readability:

```ts
export type Id = string;
export type Revision = number;

export type EntityTable<T> = {
  byId: Record<Id, T>;
  order: Id[];
  revision: Revision;
};
```

Rules:

- IDs are stable across local and remote storage when possible.
- Local-only IDs use a `local_` prefix until synced.
- Revisions increment on any meaningful record update.
- Large record tables expose one top-level revision for cheap Svelte
  invalidation.

### Auth State

```ts
export type AuthStatus =
  | 'unknown'
  | 'anonymous'
  | 'authenticated'
  | 'offline-authenticated'
  | 'expired';

export type AuthUser = {
  id: Id;
  primaryEmail: string;
  displayName: string;
  avatarUrl?: string;
  createdAt: string;
};

export type AuthState = {
  status: AuthStatus;
  userId: Id | null;
  sessionExpiresAt?: string;
  csrfToken?: string;
  offlineUnlockedAt?: string;
};
```

The Svelte app should not store OAuth tokens, raw passwords, OPAQUE export keys,
magic-link tokens, or passkey private material.

### Account, Team, Project

```ts
export type Account = {
  id: Id;
  name: string;
  slug: string;
  kind: 'personal' | 'corporate';
  billingStatus: 'trial' | 'active' | 'past_due' | 'disabled';
  createdAt: string;
};

export type Team = {
  id: Id;
  accountId: Id;
  name: string;
  slug: string;
  createdAt: string;
};

export type Project = {
  id: Id;
  accountId: Id;
  ownerKind: 'user' | 'team' | 'account';
  ownerId: Id;
  name: string;
  slug: string;
  storageMode: 'shared_postgres' | 'local_indexeddb';
  createdAt: string;
  updatedAt: string;
};

export type ProjectRole = 'read' | 'write' | 'admin' | 'owner';

export type ProjectGrant = {
  id: Id;
  projectId: Id;
  subjectKind: 'user' | 'team' | 'account_users';
  subjectId: Id;
  role: ProjectRole;
};
```

First login creates a personal account and a starter project. Later team and
corporate flows attach more grants without changing the project data model.

### Workspace State

```ts
export type WorkspaceState = {
  activeProjectId: Id | null;
  activeModelId: Id | null;
  activeEngineId: Id | null;
  activeSessionId: Id | null;
  activeSheetId: Id | null;
  online: boolean;
  hydrated: boolean;
};
```

This is small and hot. Keep it separate from heavy entities.

### Model Documents

```ts
export type ModelDocument = {
  id: Id;
  projectId: Id;
  filename: string;
  path?: string;
  text: string;
  savedTextHash?: string;
  dirty: boolean;
  parseRevision: Revision;
  engineRevision: Revision;
  createdAt: string;
  updatedAt: string;
};

export type ModelRevision = {
  id: Id;
  modelId: Id;
  projectId: Id;
  revision: Revision;
  textHash: string;
  text?: string;
  createdAt: string;
  createdBy: Id;
};
```

The editor mutates `ModelDocument.text`. Engine services observe explicit user
commands such as load/check, not every keystroke.

### Engine Sessions

```ts
export type EngineKind = 'browser-js-wasm' | 'hosted-go';

export type EngineCapabilities = {
  offline: boolean;
  persistentJobs: boolean;
  cancelJob: boolean;
  eventStream: boolean;
  parallelJobs: boolean;
};

export type EngineSession = {
  id: Id;
  kind: EngineKind;
  status: 'idle' | 'starting' | 'ready' | 'busy' | 'failed' | 'closed';
  capabilities: EngineCapabilities;
  projectId: Id;
  currentModelId?: Id;
  currentModelRevision?: Revision;
  error?: string;
  createdAt: string;
  updatedAt: string;
};
```

### Engine Interface

```ts
export interface IvyEngine {
  readonly kind: EngineKind;

  newSession(projectId: Id): Promise<EngineSession>;
  closeSession(sessionId: Id): Promise<void>;

  loadModel(sessionId: Id, model: ModelDocument): Promise<VerificationJob>;
  runCommand(intent: CommandIntent): Promise<VerificationJob>;
  cancelJob(jobId: Id): Promise<void>;

  getSnapshot(sessionId: Id, request: SnapshotRequest): Promise<SnapshotBundle>;
  subscribe(sessionId: Id, onEvent: (event: EngineEvent) => void): Unsubscribe;
}
```

Important: the browser worker and hosted server return the same app-level
records. Only transport differs.

### Jobs

```ts
export type VerificationJobKind =
  | 'load'
  | 'check-induction'
  | 'check-bounded'
  | 'check-pdr'
  | 'check-concrete'
  | 'arg-action'
  | 'concept-action'
  | 'proof-action'
  | 'event-action';

export type VerificationJobStatus =
  | 'queued'
  | 'running'
  | 'succeeded'
  | 'failed'
  | 'cancelled';

export type VerificationJob = {
  id: Id;
  sessionId: Id;
  engineId: Id;
  projectId: Id;
  modelId: Id;
  modelRevision: Revision;
  kind: VerificationJobKind;
  status: VerificationJobStatus;
  progress?: {
    phase: string;
    done?: number;
    total?: number;
  };
  resultId?: Id;
  error?: string;
  startedAt?: string;
  finishedAt?: string;
  createdAt: string;
  updatedAt: string;
};
```

Every long-running operation is tracked here, including browser worker jobs.

### Engine Events

```ts
export type EngineEvent =
  | { type: 'session-ready'; session: EngineSession }
  | { type: 'job-created'; job: VerificationJob }
  | { type: 'job-progress'; jobId: Id; progress: VerificationJob['progress'] }
  | { type: 'job-succeeded'; jobId: Id; result: JobResult }
  | { type: 'job-failed'; jobId: Id; error: string }
  | { type: 'graph-updated'; snapshot: GraphSnapshot }
  | { type: 'concept-updated'; concept: ConceptState }
  | { type: 'proof-updated'; proof: ProofState }
  | { type: 'trace-updated'; sheet: TraceEventSheet }
  | { type: 'diagnostic'; severity: DiagnosticSeverity; message: string };
```

Do not let components subscribe directly to raw SSE/worker messages. A service
normalizes and reduces events.

### Graph Snapshots

`webui` serializes Cytoscape-ish data directly. `svk` should normalize it
before storing.

```ts
export type GraphKind = 'arg' | 'concept' | 'proof' | 'event-trace';

export type GraphSnapshot = {
  id: Id;
  sheetId: Id;
  kind: GraphKind;
  sourceRevision: Revision;
  nodes: Record<Id, GraphNode>;
  edges: Record<Id, GraphEdge>;
  nodeOrder: Id[];
  edgeOrder: Id[];
  layout?: Record<Id, { x: number; y: number }>;
  styleRevision: Revision;
  createdAt: string;
};

export type GraphNode = {
  id: Id;
  obj: string;
  label: string;
  classes: string[];
  shape?: 'ellipse' | 'octagon' | 'rectangle';
  shortInfo?: string;
  longInfo?: unknown;
  actions?: NodeAction[];
  width?: number;
  height?: number;
  color?: string;
};

export type GraphEdge = {
  id: Id;
  obj: string;
  source: Id;
  target: Id;
  label?: string;
  classes: string[];
  shortInfo?: string;
  longInfo?: unknown;
};
```

Graph rendering rule:

- Global state stores `GraphSnapshot`.
- `GraphView.svelte` owns the live Cytoscape instance.
- `GraphView.svelte` diffs snapshots by `sourceRevision` and `styleRevision`.
- Pan, zoom, hover, and viewport state stay inside `GraphView.svelte` unless the
  user explicitly saves layout.

### Concept State

```ts
export type Concept = {
  name: string;
  variables: string[];
  formula: string;
  sorts: string[];
  arity: number;
};

export type ConceptToggles = {
  edges: Record<string, Record<EdgeDisplayClass, boolean>>;
  labels: Record<string, Record<NodeLabelDisplayClass, boolean>>;
};

export type EdgeDisplayClass =
  | 'all_to_all'
  | 'edge_unknown'
  | 'none_to_none'
  | 'transitive';

export type NodeLabelDisplayClass =
  | 'node_necessarily'
  | 'node_maybe'
  | 'node_necessarily_not';

export type ConceptState = {
  id: Id;
  sessionId: Id;
  sheetId: Id;
  concepts: Record<string, Concept>;
  sortNodes: string[];
  relationEdges: string[];
  nodeLabels: string[];
  abstractValue: Record<string, boolean>;
  toggles: ConceptToggles;
  selectedConcept?: string;
  revision: Revision;
};
```

Concept toggles are both UI state and engine input. Toggle changes emit a
command intent; the service updates optimistic UI and asks the engine for the
next concept graph snapshot.

### Selection And Details

```ts
export type SelectionState = {
  graphSelections: Record<Id, { nodeIds: Id[]; edgeIds: Id[] }>;
  activeDetails?: {
    source: 'arg' | 'concept' | 'proof' | 'event';
    id: Id;
    shortInfo?: string;
    longInfo?: unknown;
  };
};
```

Details are derived from selected graph elements or result records. Avoid
copying detail text into multiple stores.

### Commands

Dynamic graph and menu actions become descriptors and intents.

```ts
export type NodeAction = {
  label: string;
  action: string;
  args?: Record<string, unknown>;
};

export type CommandDescriptor = {
  id: string;
  label: string;
  icon?: string;
  enabled: boolean;
  argsSchema?: unknown;
};

export type CommandIntent = {
  id: Id;
  sessionId: Id;
  engineId: Id;
  commandId: string;
  target?: {
    kind: GraphKind;
    graphId?: Id;
    nodeId?: Id;
    edgeId?: Id;
    obj?: string;
  };
  args?: Record<string, unknown>;
};
```

Components emit intents. `commandService` decides whether the command is local
UI-only, local persistence, browser-engine, or hosted-engine work.

### Check Results

```ts
export type CheckResult = {
  id: Id;
  jobId: Id;
  sessionId: Id;
  mode: 'induction' | 'bounded' | 'pdr' | 'concrete';
  z3Contacted: boolean;
  result: 'pass' | 'fail' | 'error';
  message: string;
  failedConjecture?: string;
  failedLabel?: string;
  usedRelations?: string[];
  counterexampleTrace?: string;
  counterexampleDetails?: string;
  graphSnapshotId?: Id;
  conceptStateId?: Id;
  createdAt: string;
};
```

### Event Traces

```ts
export type TraceEvent = {
  id: Id;
  text: string;
  address: string;
  children?: TraceEvent[];
};

export type TraceEventSheet = {
  id: Id;
  projectId: Id;
  sessionId: Id;
  label: string;
  events: TraceEvent[];
  patterns: string[];
  selectedAddress?: string;
  expandedAddresses: string[];
  revision: Revision;
};
```

### Layout

```ts
export type LayoutState = {
  leftPaneWidth: number;
  rightPaneWidth: number;
  bottomPaneHeight: number;
  activeTabByRegion: Record<string, Id>;
  tutorialVisible: boolean;
  compactMode: boolean;
};
```

Layout persists locally per user/browser, not necessarily to Postgres at first.

## Auth Server Design

### Server Responsibilities

The Go server owns:

- marketing pages
- login pages
- OPAQUE password auth
- OAuth start/callback
- passkey registration/login
- magic email login/recovery
- session cookie issue/rotation/revocation
- CSRF tokens
- project/account/team APIs
- serving SvelteKit private app assets
- hosted native Go/Z3 engine endpoints

The SvelteKit app owns:

- authenticated workspace UI
- local-first IndexedDB data
- browser wasm engine
- sync UI
- project/model/job rendering

### Public Routes

```text
GET  /
GET  /pricing
GET  /docs
GET  /security
GET  /login
GET  /signup
```

Use Go templates. Public pages should not load the private Svelte app bundle.

### Auth Routes

```text
GET  /auth/me
POST /auth/opaque/register/start
POST /auth/opaque/register/finish
POST /auth/opaque/login/start
POST /auth/opaque/login/finish
POST /auth/logout

GET  /auth/oauth/google/start
GET  /auth/oauth/google/callback
GET  /auth/oauth/github/start
GET  /auth/oauth/github/callback
GET  /auth/oauth/apple/start
POST /auth/oauth/apple/callback

POST /auth/email/request
POST /auth/email/consume

POST /auth/passkeys/register/options
POST /auth/passkeys/register/finish
POST /auth/passkeys/login/options
POST /auth/passkeys/login/finish

GET  /auth/identities
POST /auth/identities/link/start
POST /auth/identities/link/finish
DELETE /auth/identities/{identity_id}
```

### App Routes

```text
GET /app
GET /app/*
GET /assets/*
```

Rules:

- `/app` requires an authenticated server session, unless serving the cached PWA
  offline after previous authentication.
- Go serves the built SvelteKit app.
- The app bootstraps by calling `/auth/me`, then local IndexedDB hydration.

### Session Cookies

Use opaque server-side sessions:

```text
cookie name: ivy_session
HttpOnly
Secure in production
SameSite=Lax by default
Path=/
```

Session table stores:

- session ID hash, not raw cookie token
- user ID
- created time
- last seen time
- expiry
- revoked time
- user agent hash
- optional IP prefix
- CSRF secret or CSRF token binding material

Rotate session ID after login and sensitive identity-link operations.

### CSRF Strategy

Use a combination of:

- `SameSite=Lax` session cookie
- per-session CSRF token
- require `X-CSRF-Token` on mutating JSON endpoints
- reject unsafe methods with missing or invalid token
- check `Origin` on unsafe methods
- do not allow wildcard CORS for authenticated endpoints

`GET /auth/me` returns the CSRF token only to same-origin authenticated app
requests.

### XSS Strategy

Core rules:

- Never render user-provided Ivy text as raw HTML.
- Use Svelte escaping by default.
- Use text nodes for diagnostics, traces, formulas, and solver output.
- Sanitize only if future rich text is unavoidable.
- Set a strict Content Security Policy.

Initial CSP target:

```text
default-src 'self';
script-src 'self';
style-src 'self' 'unsafe-inline';
img-src 'self' data:;
font-src 'self';
connect-src 'self';
worker-src 'self' blob:;
object-src 'none';
base-uri 'none';
frame-ancestors 'none';
form-action 'self';
```

The `'unsafe-inline'` style allowance may be needed for Svelte/Cytoscape during
early implementation. Do not allow inline scripts.

### OPAQUE Password Login

Password login must use OPAQUE. The browser may ask the user for a password,
but the Go server must never receive the raw password and must never store a
password hash derived from a submitted raw password.

Libraries:

- Go server: `github.com/bytemare/opaque`
- Svelte client: `@cloudflare/opaque-ts`

OPAQUE credential table stores:

- user ID
- OPAQUE server registration record / envelope material required by the chosen
  Go library
- OPAQUE server public key identifier / credential version
- created time
- updated time
- disabled time

The OPAQUE server setup secret must come from server configuration and must not
be checked into the repository. Add basic login rate limiting by account/email
and source IP. Treat OPAQUE registration and login messages as protocol
messages, not passwords.

### OAuth

Providers:

- Google
- GitHub
- Apple

OAuth identity table:

- provider
- provider subject
- provider email
- provider email verified
- user ID
- linked time
- last login time

Do not auto-link identities by email. If a user tries Google and an existing
OPAQUE password user has the same email, show an account-exists flow:

1. Ask the user to sign in with an existing method.
2. After signed in, let them explicitly link the provider.
3. Record audit event.

### Magic Email

Mailgun sends magic login/recovery links.

Token table stores:

- token hash, not raw token
- email
- optional user ID
- purpose: `login`, `recovery`, `identity_link`
- expiry
- consumed time
- created source IP/user agent hash

Magic links should be single-use and short-lived. Use generic UI messages so an
attacker cannot enumerate accounts.

### Passkeys

Design schema now:

- credential ID
- user ID
- public key
- sign count
- transports
- attestation type
- AAGUID if available
- user-visible nickname
- created time
- last used time

Use a reputable WebAuthn library in Go. Passkeys can ship after password/OAuth
sessions are stable, but database migrations should anticipate them.

## Postgres Schema Plan

Initial schemas:

```text
control
project_data
audit
```

Core control tables:

```text
control.users
control.user_emails
control.opaque_credentials
control.oauth_identities
control.passkey_credentials
control.email_tokens
control.sessions

control.accounts
control.account_users
control.teams
control.team_members
control.projects
control.project_grants
control.project_storage_locations

audit.auth_events
audit.project_events
```

Project-owned hosted tables:

```text
project_data.models
project_data.model_revisions
project_data.engine_sessions
project_data.verification_jobs
project_data.job_results
project_data.graph_snapshots
project_data.concept_states
project_data.trace_event_sheets
project_data.sync_operations
```

Every project-owned table includes `project_id`.

Project queries run inside explicit transactions that set:

```sql
SELECT set_config('ivy.user_id', $1, true);
SELECT set_config('ivy.project_id', $2, true);
SELECT set_config('ivy.project_role', $3, true);
```

Enable and force RLS:

```sql
ALTER TABLE project_data.models ENABLE ROW LEVEL SECURITY;
ALTER TABLE project_data.models FORCE ROW LEVEL SECURITY;
```

## IndexedDB Plan

Use IndexedDB for offline access. Wrap it with a tiny repository layer rather
than calling IndexedDB from stores.

Database name:

```text
ivy-svk-local
```

Object stores:

```text
meta
users
accounts
teams
projects
project_grants
models
model_revisions
engine_sessions
verification_jobs
job_results
graph_snapshots
concept_states
trace_event_sheets
sync_operations
layout_preferences
```

Offline authentication rule:

- A previous successful server login may unlock cached local projects.
- The app must clearly show `offline-authenticated`.
- Mutations create sync operations.
- Server-only account changes wait until online.

## Sync Queue

```ts
export type SyncOp = {
  id: Id;
  projectId: Id;
  entity:
    | 'model'
    | 'model_revision'
    | 'job'
    | 'job_result'
    | 'graph_snapshot'
    | 'concept_state'
    | 'trace_event_sheet';
  op: 'create' | 'update' | 'delete';
  baseRevision?: Revision;
  payload: unknown;
  status: 'pending' | 'syncing' | 'synced' | 'conflict' | 'failed';
  error?: string;
  createdAt: string;
  updatedAt: string;
};
```

First sync policy:

- Models use revision/hash conflict detection.
- Job results and graph snapshots are append-only.
- Layout stays local.
- Project membership and auth state are server authoritative.

## Implementation Phases

Each phase should end with passing tests before moving on.

### Phase 0: Baseline Tooling And Guardrails

Tasks:

- Keep `svk` as the active SvelteKit app.
- Decide whether to switch to `adapter-static` for Go serving.
- Add TypeScript strictness if not already enabled.
- Add path aliases for `$lib`.
- Add lint/test scripts that do not require the full backend.
- Add a `src/lib/types` directory with pure type modules.

Tests:

- `npm run check`
- `npm run test:unit -- --run`
- basic Playwright smoke test opens the Svelte app shell

Do not add auth or engines yet.

### Phase 1: Shared Type Contracts

Tasks:

- Implement TypeScript types listed in this document.
- Add JSON validators or lightweight parse functions for externally supplied
  records.
- Add normalizers for old webui Cytoscape payloads into `GraphSnapshot`.
- Add normalizers for check results and concept payloads.

Unit tests:

- graph normalizer handles empty graph
- graph normalizer maps nodes/edges by stable IDs
- graph normalizer preserves node actions
- concept normalizer handles missing toggle maps
- check result normalizer handles pass/fail/error
- event trace normalizer assigns stable IDs from addresses

Go tests:

- Add tests near web engine response code once Go server exists.
- Validate JSON response fixtures match TypeScript expectations.

Exit criteria:

- TypeScript can represent all current `webui.Backend` response shapes needed
  by early UI.

### Phase 2: Reactive Stores Without Effects

Tasks:

- Create Svelte state modules:
  - `workspace.svelte.ts`
  - `auth.svelte.ts`
  - `projects.svelte.ts`
  - `models.svelte.ts`
  - `engines.svelte.ts`
  - `jobs.svelte.ts`
  - `graphs.svelte.ts`
  - `concepts.svelte.ts`
  - `selection.svelte.ts`
  - `layout.svelte.ts`
- Stores expose small mutation methods only.
- No fetch, worker, IndexedDB, or timers in stores.
- Use revision counters for large tables.

Unit tests:

- store insertion preserves `byId` and `order`
- updating one graph bumps graph table revision
- selection derives active details from graph node metadata
- job store handles lifecycle transitions
- model store tracks dirty/saved state by text hash

Exit criteria:

- A fake in-memory app state can load a model, create a session, attach a graph,
  select a node, and show details without any engine.

### Phase 3: IndexedDB Repository Layer

Tasks:

- Implement `indexedDb.ts` connection and schema versioning.
- Implement repositories for:
  - projects
  - models
  - model revisions
  - jobs
  - results
  - graph snapshots
  - concept states
  - sync ops
  - layout preferences
- Add migration infrastructure from version 1.
- Add deterministic test database names.

Unit tests:

- create/open database
- migrate empty database
- save/load project
- save/load model and model revision
- save/load graph snapshot
- enqueue/list/mark sync op
- delete test database after run

Playwright tests:

- app persists a model to IndexedDB
- reload restores the model
- offline browser context still opens cached app shell after prior load

Exit criteria:

- Local project/model state can survive reload before any auth server is built.

### Phase 4: App Bootstrap Service

Tasks:

- Implement `appBootstrapService`.
- Boot order:
  1. load local IndexedDB metadata
  2. call `/auth/me` when online
  3. reconcile auth status
  4. load local projects/models
  5. select last active project/model if allowed
- Add visible offline/auth status.

Tests:

- online authenticated bootstrap
- online anonymous bootstrap
- offline previously authenticated bootstrap
- offline never-authenticated bootstrap shows login-required state
- corrupt local metadata does not crash app

Exit criteria:

- App startup is deterministic and observable.

### Phase 5: Engine Interface And Fake Engine

Tasks:

- Define `IvyEngine`.
- Implement `FakeEngine` for tests.
- Implement `engineService` to reduce engine events into stores.
- Implement command intent dispatch.

Unit tests:

- fake engine creates session
- fake engine emits job lifecycle
- engineService stores job result
- graph update event stores snapshot
- concept update event stores concept state
- failed job stores error and toast

Exit criteria:

- UI can run against `FakeEngine` end to end.

### Phase 6: Initial Workspace UI Skeleton

Tasks:

- Build the real first screen: workspace, not landing page.
- Add panes:
  - editor
  - ARG graph
  - concept graph
  - details/check output
  - job/status strip
- Add toolbar commands:
  - load model
  - save local
  - run induction
  - run bounded
  - choose engine
- Add keyboard-safe editor integration.

Playwright tests:

- first authenticated app screen is workspace
- editor accepts text
- dirty indicator changes
- fake check creates job row
- fake graph renders visible nodes
- selecting graph node updates details
- layout resize does not overlap major panes

Exit criteria:

- With fake engine, the app feels like the intended product shell.

### Phase 7: Graph Component

Tasks:

- Add Cytoscape rendering component.
- Component accepts `GraphSnapshot`.
- Component owns Cytoscape instance.
- Component emits selection and action intents.
- Add graph style modules for ARG and concept graphs.
- Avoid rebuilding Cytoscape for simple selection changes.

Unit tests:

- convert `GraphSnapshot` to Cytoscape elements
- preserve node classes and edge classes
- preserve action descriptors
- layout position round-trips if provided

Playwright tests:

- graph canvas/container is non-empty
- ARG nodes render
- concept nodes render
- node selection updates details
- context menu or action menu emits command intent
- graph remains usable after snapshot update

Exit criteria:

- Large graph snapshots update by diff/revision, not full app rerender.

### Phase 8: Browser Wasm Engine

Tasks:

- Port only the stable browser path:
  - `goivyNodeFS`
  - `smtZ3Imports`
  - `wasm_exec-go1.25.6.js`
  - `goivy-check-js.wasm`
  - `z3-471-api.js`
  - `z3-471-api.wasm`
- Implement `ivyEngine.worker.ts`.
- Worker owns:
  - Z3 wasm initialization
  - Go wasm initialization
  - session command queue
  - stdout/stderr normalization
  - job progress messages
- Do not use `wasip1`.

Worker protocol:

```ts
type WorkerRequest =
  | { type: 'init'; assetBaseUrl: string }
  | { type: 'new-session'; requestId: Id; projectId: Id }
  | { type: 'load-model'; requestId: Id; sessionId: Id; model: ModelDocument }
  | { type: 'run-command'; requestId: Id; intent: CommandIntent }
  | { type: 'cancel-job'; requestId: Id; jobId: Id };
```

Unit tests:

- worker protocol serializer/deserializer
- worker event reducer
- browser engine handles worker error

Playwright tests:

- Z3 wasm initializes
- Go wasm initializes
- simple model load calls `goivyCheckRun`
- failed load reports diagnostic
- browser engine job appears in job list

Go tests:

- `make goivy-check-js-wasm`
- Go wasm entrypoint builds
- existing Z3 wasm tests continue to pass

Exit criteria:

- Browser local engine can execute a minimal check and report a normalized
  result.

### Phase 9: Hosted Go Engine API

Tasks:

- Factor or wrap current `webui.Backend` behavior into a hosted engine package.
- Add Go HTTP endpoints under `/api/engine`.
- Keep response shapes close to the browser engine normalized contract.
- Support server-sent events or WebSocket for job/event updates.

Initial routes:

```text
POST /api/engine/session
POST /api/engine/session/{id}/load
POST /api/engine/session/{id}/command
GET  /api/engine/session/{id}/snapshot
GET  /api/engine/session/{id}/events
POST /api/engine/jobs/{id}/cancel
```

Go tests:

- create session
- load valid Ivy model
- load invalid Ivy model reports compiler error
- run induction check
- get ARG snapshot
- event stream emits job lifecycle
- unauthorized request rejected
- cross-project request rejected

TypeScript tests:

- `HostedGoEngine` maps HTTP responses to engine events
- SSE reconnect behavior
- HTTP error maps to failed job

Playwright tests:

- app runs against hosted test server
- user loads model
- hosted check creates job
- result appears in UI

Exit criteria:

- The same UI can switch between fake, browser wasm, and hosted Go engines.

### Phase 10: Go Auth Server Foundation

Tasks:

- Create `goivy/cmd/ivysvk`.
- Create server package under `goivy/svk/server`.
- Add config loading:
  - DB DSN
  - public base URL
  - Mailgun domain/API key
  - OAuth client IDs/secrets
  - cookie secret if needed
  - development mode
- Add health endpoint.
- Add public Go template pages.
- Add static serving for built SvelteKit bundle.

Go tests:

- config loads from env
- missing required production config fails
- public pages render
- `/app` redirects anonymous users to `/login`
- authenticated `/app` serves app shell

Exit criteria:

- There is a Go server process that can serve public pages and a protected app
  shell.

### Phase 11: Database Migrations And Admin Command

Tasks:

- Create migration table.
- Create initial auth/project schema.
- Create project_data schema with RLS.
- Create `ivysvk-admin`.
- Add commands:
  - migrate
  - seed-dev
  - create-user
  - create-account
  - create-project
  - grant-project

Go tests:

- migrations apply to empty database
- migrations are idempotent
- seed-dev creates user/account/project
- project RLS permits own project
- project RLS rejects other project

Use integration tests gated by env var:

```text
IVYSVK_TEST_DB_DSN
```

Exit criteria:

- Local dev can create a working DB without ad hoc SQL.

### Phase 12: OPAQUE Password Auth

Tasks:

- Add `@cloudflare/opaque-ts` to the Svelte client.
- Add `github.com/bytemare/opaque` to the Go server.
- Implement OPAQUE server setup/key management.
- Add registration start/finish handlers.
- Add login start/finish handlers.
- Add logout handler.
- First OPAQUE registration/login creates personal account/project if needed.
- Add rate limiting.
- Add generic error responses.
- Add session cookie issue/revoke.

Go tests:

- registration stores OPAQUE credential material, not a password hash
- login succeeds through a valid OPAQUE exchange
- wrong password fails through OPAQUE without exposing which step was wrong
- raw password never appears in server request structs, logs, or database rows
- OPAQUE server setup secret is required outside dev/test mode
- registration creates user
- first registration/login creates personal account/project
- logout revokes session
- session cookie is HttpOnly/SameSite
- mutating endpoint without CSRF rejected
- mutating endpoint with CSRF accepted

Playwright tests:

- OPAQUE signup flow reaches app
- OPAQUE login flow reaches app
- logout returns to login
- reload preserves session

Exit criteria:

- Password login works through OPAQUE, and the server never sees the user's raw
  password.

### Phase 13: Mailgun Magic Email

Tasks:

- Wrap Mailgun in `server/mail`.
- Use existing Mailgun env setup.
- Add email token generation/consume.
- Add login/recovery email template.
- Store token hash only.
- Use generic success response.

Go tests:

- token raw value not stored
- expired token rejected
- consumed token cannot be reused
- fake Mailgun sender receives expected message
- unknown email request does not reveal account existence

Playwright tests:

- request magic link shows generic sent message
- consume test token signs user in

Exit criteria:

- Magic email login/recovery is available through testable sender abstraction.

### Phase 14: OAuth Providers

Tasks:

- Implement OAuth state and PKCE where applicable.
- Add Google.
- Add GitHub.
- Add Apple.
- Store OAuth identity records.
- Do not auto-link same-email identities.
- Add explicit identity linking flow.

Go tests:

- OAuth start sets state cookie/record
- callback validates state
- callback rejects invalid state
- new OAuth identity creates user/account/project
- same email different provider does not auto-link
- signed-in explicit link attaches identity
- provider subject uniqueness enforced

Playwright tests:

- mocked Google login creates account
- mocked GitHub same-email flow asks for explicit linking
- explicit linking flow succeeds after OPAQUE password login

Exit criteria:

- OAuth is production-shaped, even if local tests use fake providers.

### Phase 15: Passkeys

Tasks:

- Choose Go WebAuthn library.
- Implement registration options/finish.
- Implement login options/finish.
- Store passkey credentials.
- Add passkey management UI.

Go tests:

- registration challenge stored and expires
- finish validates challenge
- login validates assertion
- replay rejected
- credential can be disabled

Playwright tests:

- use browser virtual authenticator
- register passkey
- logout
- login with passkey

Exit criteria:

- Passkeys are first-class login credentials.

### Phase 16: Project APIs

Tasks:

- Implement account/team/project list APIs.
- Implement project creation.
- Implement grants.
- Implement team membership.
- Add project switcher in app.

Go tests:

- user sees own personal project
- user does not see unrelated project
- team grant gives project access
- account_users grant gives account-wide user access
- write role can save model
- read role cannot save model

Playwright tests:

- project switcher lists accessible projects
- create project
- switch project
- unauthorized project URL is rejected

Exit criteria:

- Hosted project authorization model is usable.

### Phase 17: Hosted Persistence APIs

Tasks:

- Implement model CRUD.
- Implement model revisions.
- Implement job/result/snapshot persistence.
- Implement sync upload/download endpoints.

Go tests:

- save model revision
- conflict on stale base revision
- append job result
- graph snapshots are immutable
- RLS applies to all project_data tables

TypeScript tests:

- sync client uploads pending model revision
- conflict response creates conflict sync op
- append-only result sync does not overwrite local result

Playwright tests:

- edit model online
- save
- reload
- model persists from server

Exit criteria:

- Online hosted data persistence exists for model and result records.

### Phase 18: Offline Plane Mode

Tasks:

- Add service worker/PWA asset caching.
- Cache Svelte app assets.
- Cache wasm assets.
- Cache last auth/user/project shell metadata.
- Add clear offline status and sync status.
- Ensure app starts without network after prior login.

Playwright tests:

- login online
- create/open model
- go offline
- reload app
- edit model
- run browser wasm check
- queue sync op
- go online
- sync completes

Exit criteria:

- A previously authenticated user can work on a plane.

### Phase 19: Engine Parity Harness

Tasks:

- Create a small suite of Ivy examples.
- Run same commands on BrowserWasmEngine and HostedGoEngine.
- Compare normalized results, not raw logs.

Tests:

- load same model in both engines
- compare ARG snapshot shape
- compare check pass/fail
- compare failed conjecture text
- compare concept relation/toggle payload
- record known differences explicitly

Exit criteria:

- Engine-neutral data model is proven by side-by-side execution.

### Phase 20: Security Test Suite

Go tests:

- CSRF missing token rejected
- CSRF wrong token rejected
- Origin mismatch rejected
- session fixation prevented by rotation after login
- expired session rejected
- revoked session rejected
- OPAQUE credential records contain protocol material but no raw password or
  conventional password hash
- OAuth invalid state rejected
- magic token replay rejected
- project RLS denies cross-project access
- XSS strings are escaped in Go templates

Playwright tests:

- malicious Ivy output renders as text, not HTML
- script tag in model text does not execute
- public pages do not expose app bootstrap data
- app route redirects anonymous user
- offline cached app does not expose another user's project

Exit criteria:

- Basic web security protections are covered by automated tests.

## Build And Test Commands

Expected local commands after implementation matures:

```sh
cd goivy/svk
npm run check
npm run test:unit -- --run
npm run test:e2e
```

Go:

```sh
cd goivy
go test ./svk/server/...
go test ./cmd/ivysvk ./cmd/ivysvk-admin
go test ./webui ./webvue ./goldweb ./smt ./xtracer
make goivy-check-js-wasm
```

Integration tests:

```sh
IVYSVK_TEST_DB_DSN='postgres://...' go test ./svk/server/... -run Integration
```

## Early Definition Of Done

The first serious milestone is not "all features ported." It is:

- Go auth server runs.
- OPAQUE password login works.
- First login creates user/account/project.
- Go serves the private SvelteKit app.
- Svelte app starts online and offline after prior login.
- IndexedDB stores a project/model.
- Browser wasm engine can run a minimal local check.
- Hosted engine can run the same minimal check.
- Both engines produce normalized job/result/graph records.
- Playwright verifies the core happy path.

Only after this should we port the larger `webui` workflows.

## Open Questions To Revisit Later

- Whether marketing pages should eventually move into SvelteKit for design
  reuse.
- Whether hosted engine jobs should become durable worker-queue jobs instead of
  request-bound jobs.
- Whether large graph snapshots need structural sharing or binary compression.
- Whether local project encryption is needed for offline data at rest.
- Whether corporate SSO/SAML should be added after OAuth.
- Whether model collaboration needs CRDT-style live editing or simpler revision
  conflict handling.

## Immediate Next Steps

1. Add shared TypeScript type modules.
2. Add pure normalizer tests for graph/check/concept payloads.
3. Add store modules with no effects.
4. Add IndexedDB repositories.
5. Add fake engine and workspace skeleton.
6. In parallel, begin Go auth server schema and session foundation.

This sequence keeps the architecture testable at every layer and avoids getting
wedged behind the full verification engine, full auth system, or full graph UI
before the reactive spine is proven.
