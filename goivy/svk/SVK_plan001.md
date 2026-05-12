# SVK Plan 001: Finish Plan 000 Through The Engine Boundary

## Purpose

`SVK_plan000.md` put the first beams in place: normalized Svelte state,
engine-neutral records, auth/server scaffolding, IndexedDB scaffolding, fake
engine tests, and early browser wasm assets.

This follow-on plan is about finishing the half-built beams instead of starting
another parallel spike. The central goal is:

> Make `goivy/svk` talk through one abstract `IvyEngine` interface that can run
> either against a native hosted Go/Z3 backend or fully in-browser with
> `GOOS=js GOARCH=wasm` Go and wasm Z3.

The `/webui` server is the first real hosted backend target because it already
knows how to create sessions, load Ivy models, compile them, run Z3-backed
checks, produce ARG/concept payloads, and stream SSE events. But `/webui` must
not become the frontend contract. It is one adapter behind the contract.

## Key Decision: What Gets Compiled To Wasm?

We should not port Ivy verification semantics to TypeScript.

We should not compile the whole old `webui` application as-is to wasm either.
The old `webui` is valuable, but it mixes backend/session behavior, HTTP
handler shape, and old UI/controller assumptions.

The direction is:

- Keep parser/compiler/checker/Z3 interaction and interactive analysis behavior
  in Go.
- Extract or wrap the reusable `webui` backend/session behavior into a transport
  neutral Go engine core.
- Compile that engine core to native Go for the hosted server.
- Compile that same engine core to `GOOS=js GOARCH=wasm` for offline browser
  execution.
- Reimplement UI/controller presentation in Svelte/TypeScript.
- Reimplement only adapter glue, normalized state reducers, rendering, local
  persistence, worker protocol, and transport clients in TypeScript.

In other words:

```text
Do compile:
  Go Ivy compiler/checker/analysis engine
  reusable webui session/backend behavior after transport-neutral extraction

Do not compile:
  old static webui frontend
  HTTP handlers as the offline API boundary
  old controller-heavy browser orchestration

Do reimplement in TypeScript/Svelte:
  reactive state
  panes and layout
  editor integration
  graph rendering components
  local IndexedDB persistence
  sync queue
  HostedGoEngine adapter
  BrowserWasmEngine adapter
```

This preserves one verification behavior implementation while letting the
modern UI stay reactive and maintainable.

## Current State

### Implemented Or Started In `svk/`

- `IvyEngine` TypeScript interface exists.
- `FakeEngine` exists and drives the current UI/test shell.
- `HostedGoEngine` exists, but currently targets invented `/api/engine/...`
  endpoints rather than the actual `/webui` API.
- `BrowserWasmEngine` exists, but its worker is still a lifecycle stub.
- `ivyEngine.worker.ts` prevents `wasip1` use and has message plumbing, but does
  not yet load `wasm_exec`, `goivy-check-js.wasm`, or Z3 wasm.
- Normalizers exist for Cytoscape-like graph payloads, concept payloads, and
  check results.
- `EngineService` exists and reduces engine events into Svelte stores.
- Auth, OPAQUE, passkey, OAuth, magic-link, CSRF, and project route scaffolding
  exists on the Go server side.
- IndexedDB repository/sync scaffolding exists.
- The first screen now resembles the old `webui` workbench and includes the
  editor, ARG, concept graph, state/relations, details, and status strip.

### Still Half-Done

- The main Svelte route is still a mostly static/fake workbench.
- The engine choice is not yet a real runtime selection.
- `HostedGoEngine` does not talk to the real `/api/session/...` server.
- No adapter yet converts actual `/webui` responses into `EngineEvent`,
  `VerificationJob`, `GraphSnapshot`, `ConceptState`, and `CheckResult`.
- Check results are not stored in first-class state.
- The details pane still shows canned verification text.
- The state/relations pane is not driven from concept/toggle payloads.
- ARG/concept actions are not mapped through a command registry.
- The browser wasm worker has no real Go runtime bridge.
- The Go engine code is not factored into a shared native/wasm engine core.
- The `svk/server` authenticated app server and the `/webui` engine server are
  not integrated.
- There is no end-to-end test proving: edit model -> submit to native backend ->
  compile -> run Z3 check -> render real failure/pass result.

## Architecture Target

The frontend should depend on this shape:

```text
Svelte components
  emit UI intents

CommandService
  turns UI intents into CommandIntent

EngineService
  owns active IvyEngine session
  calls IvyEngine methods
  reduces EngineEvent streams into stores

IvyEngine interface
  HostedWebuiEngine adapter
    HTTP/SSE to native Go webui-compatible server
  BrowserWasmEngine adapter
    Worker protocol to Go js/wasm engine core + wasm Z3
  FakeEngine
    deterministic tests and UI smoke work

Stores
  models, jobs, graphs, concepts, checks, stateRelations, diagnostics, layout

Persistence
  IndexedDB local project/model/job/check snapshots
  Postgres canonical hosted auth/project/account data
```

The `/webui` server shape should be treated as an adapter source:

```text
POST /api/session/new
POST /api/session/{id}/load
POST /api/session/{id}/check
GET  /api/session/{id}/arg
GET  /api/session/{id}/concept
GET  /api/session/{id}/toggles
POST /api/session/{id}/toggles
POST /api/session/{id}/action
POST /api/session/{id}/arg/action
GET  /api/session/{id}/events
```

The normalized `IvyEngine` contract should remain the app contract:

```ts
interface IvyEngine {
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

We may extend this with explicit optional capabilities, but the frontend should
not import `/webui` response types directly outside the hosted adapter and
normalizer tests.

## Shared Engine Core Plan

The Go side needs a transport-neutral engine core that can be used in native
server mode and browser wasm mode.

### Proposed Go Shape

Add a package

```text
goivy/webengine/
```

The package owns Go-native methods that return typed Go structs, not raw HTTP
responses:

```go
type Engine struct {
    backend *webui.GoBackend
}

type SessionInfo struct {
    ID string `json:"id"`
}

func (e *Engine) NewSession(ctx context.Context) (SessionInfo, error)
func (e *Engine) LoadModel(ctx context.Context, sessionID, filename string, content []byte) (LoadResult, error)
func (e *Engine) Check(ctx context.Context, sessionID string, req CheckRequest) (CheckResult, error)
func (e *Engine) ARG(ctx context.Context, sessionID string) (CytoscapePayload, error)
func (e *Engine) Concept(ctx context.Context, sessionID string, req ConceptRequest) (ConceptPayload, error)
func (e *Engine) Toggles(ctx context.Context, sessionID string) (TogglesPayload, error)
func (e *Engine) Action(ctx context.Context, sessionID string, req ActionRequest) (ActionResult, error)
func (e *Engine) Events(ctx context.Context, sessionID string) (<-chan Event, error)
```

The first implementation can wrap `webui.GoBackend` and unmarshal its canonical
JSON byte responses into typed structs. Later, internals can be made more direct.

### Native Transport

The native server gets HTTP handlers that call the engine core. Two native
entrypoints are acceptable during the transition:

1. Keep `/webui` server as-is and write `HostedWebuiEngine` to adapt to it.
2. Add `/api/engine/...` handlers to `svk/server` that internally call the same
   core.

The first real milestone should use option 1 because it proves useful behavior
quickly. The second milestone should add option 2 so auth/project/session policy
lives in `svk/server`.

### Browser Wasm Transport

The wasm worker should call the same core through Go exported functions. It
should not emulate HTTP internally unless that turns out to be the least-bad
bridge. The preferred browser boundary is message-based:

```text
TypeScript worker message
  -> Go js.Func exported dispatcher
  -> webengine.Engine method
  -> JSON result
  -> TypeScript worker response
  -> BrowserWasmEngine normalizer
  -> EngineService stores
```

The worker API should mirror `IvyEngine` operations:

```ts
type WorkerRequest =
  | { type: 'init'; assetBaseUrl: string }
  | { type: 'new-session'; projectId: Id }
  | { type: 'load-model'; sessionId: Id; model: ModelDocument }
  | { type: 'run-command'; intent: CommandIntent }
  | { type: 'get-snapshot'; sessionId: Id; request: SnapshotRequest }
  | { type: 'cancel-job'; jobId: Id };
```

The worker should produce normalized or normalizable events:

```ts
type WorkerResponse =
  | { type: 'ready'; requestId: Id }
  | { type: 'session'; requestId: Id; session: EngineSession }
  | { type: 'job'; requestId: Id; job: VerificationJob }
  | { type: 'snapshot'; requestId: Id; bundle: SnapshotBundle }
  | { type: 'event'; sessionId: Id; event: EngineEvent }
  | { type: 'error'; requestId?: Id; error: string };
```

## Backend Adapter Strategy

### Rename The Current Hosted Adapter

The current `HostedGoEngine` name is too generic while it points at invented
endpoints. Split it:

```text
hostedWebuiEngine.ts
  speaks existing /api/session/... webui HTTP/SSE

hostedEngineApi.ts or hostedSvkEngine.ts
  future svk/server /api/engine/... authenticated engine API
```

Keep tests for both. The first implementation path uses `HostedWebuiEngine`.

### HostedWebuiEngine Contract

`newSession(projectId)`:

- `POST /api/session/new`
- response: `{ "session_id": "s1" }`
- normalize to `EngineSession`:
  - id: `s1`
  - kind: `hosted-go`
  - status: `ready`
  - capabilities:
    - offline: false
    - persistentJobs: false initially, true later if server stores jobs
    - cancelJob: false initially unless added
    - eventStream: true
    - parallelJobs: false initially because `GoBackend` single-thread context is
      deliberately serialized

`loadModel(sessionId, model)`:

- `POST /api/session/{id}/load`
- send multipart `file` with `model.text` and `model.filename`
- create a local `VerificationJob` kind `load`
- emit `job-created`, then when response returns:
  - emit `job-succeeded` or `job-failed`
  - fetch ARG
  - fetch concept
  - emit `graph-updated`
  - emit `concept-updated`

`runCommand(check.induction)`:

- ensure latest dirty model text has been loaded first
- `POST /api/session/{id}/check` with `{ "mode": "induction" }`
- create local `VerificationJob` kind `check-induction`
- normalize check result:
  - `result`
  - `mode`
  - `message`
  - `failed_conjecture`
  - `failed_label`
  - `used_relations`
  - `z3_contacted`
  - `counterexample_trace`
  - `counterexample_details`
  - `trace_arg`
- emit `check-updated` after adding that event type, or store check result from
  `job-succeeded.result` if we defer the event type.
- if `trace_arg` exists, normalize to `GraphSnapshot`.
- fetch latest ARG/concept after check and emit graph/concept updates.

`runCommand(check.bounded)`:

- same as induction, but `{ "mode": "bounded", "bound": args.bound }`.

`runCommand(concept.action)`:

- first map simple toolbar commands:
  - `concept.showReachable` -> existing `/action` or specific menu command after
    checking old frontend adapter mappings
  - `concept.reset` -> `/concept/reset`
  - `concept.diagram` -> `/concept/diagram`
- each command returns a job and refreshes concept/ARG.

`runCommand(arg.expand)`:

- `POST /api/session/{id}/arg/action` with:
  - `node`
  - `action`
  - `args`
- refresh ARG/concept after response.

`subscribe(sessionId)`:

- open `GET /api/session/{id}/events`.
- translate raw events:
  - `file_loaded` -> diagnostic/info or model metadata event
  - `check_started` -> job-progress if a matching job is active
  - `check_completed` -> job-succeeded or diagnostic until explicit check result
    mapping is complete
- Do not let raw event names leak into Svelte components.

## Engine Event Additions

Plan000 has `CheckResult` but `EngineEvent` does not yet carry one. Add:

```ts
| { type: 'check-updated'; result: CheckResult }
| { type: 'toggles-updated'; toggles: ConceptToggles; sheetId: Id }
```

Add stores:

```text
src/lib/state/checks.svelte.ts
src/lib/state/stateRelations.svelte.ts
```

`checks.svelte.ts`:

- normalized table of `CheckResult`
- `latestBySession`
- `latestByModel`
- selectors for `statusbar`, `details`, and failed conjecture display

`stateRelations.svelte.ts`:

- derived rows from `ConceptState.toggles`, `relationEdges`, `nodeLabels`, and
  relations in the raw concept payload
- no hard-coded `=@X`, `link(X,Y)`, or `semaphore` in route markup after this
  phase

Update `EngineService.reduceEvent` to handle the new events.

## UI Component Decomposition

The current `+page.svelte` should be split after real backend integration begins,
not before. The target modules are:

```text
src/lib/components/workbench/
  WorkbenchShell.svelte
  TopToolbar.svelte
  SheetTabs.svelte
  ArgPane.svelte
  ConceptPane.svelte
  StateRelationsPane.svelte
  EditorPane.svelte
  DetailsPane.svelte
  StatusBar.svelte

src/lib/components/editor/
  IvyEditor.svelte

src/lib/components/graphs/
  ArgGraph.svelte
  ConceptGraph.svelte
```

The route becomes composition and bootstrap only:

```text
+page.svelte
  create stores
  create selected engine adapter
  create services
  render WorkbenchShell
```

Do not split everything first. The order should be:

1. Make real hosted backend work through the existing page.
2. Add tests.
3. Extract components without changing behavior.

That avoids a large UI refactor with fake data still in the middle.

## Implementation Phases

Each phase should end with green checks:

```text
npm run check
npm run lint
npm run test:unit -- --run
npm run test:e2e
go test ./goivy/svk/server ./goivy/webui
```

If a phase touches broader Go engine behavior, also run the focused parser/check
packages already used in previous work.

### Phase 1: Lock The Abstract Engine Contract

Goal: make the TypeScript boundary impossible to accidentally bend around
`/webui` quirks.

Work:

- Add `check-updated` and `toggles-updated` to `EngineEvent`.
- Add `checks.svelte.ts` with unit tests.
- Add `stateRelations.svelte.ts` with unit tests.
- Add `EngineAdapterContract` test helpers that can run against fake, hosted,
  and browser adapters.
- Document which fields are required from any engine after:
  - new session
  - load model
  - check induction
  - get ARG
  - get concept/toggles

Tests:

- Unit test event reduction for checks and toggles.
- Unit test that components can render from normalized stores without knowing
  engine kind.
- Type test or regular unit test proving no route/component imports
  `webui`-specific payload types.

Exit criteria:

- Fake engine still drives the workbench.
- No `/webui` endpoint names appear outside adapter tests and adapter code.

### Phase 2: Build `HostedWebuiEngine`

Goal: talk to the actual existing `/webui` HTTP API through `IvyEngine`.

Work:

- Add `src/lib/engines/hostedWebuiEngine.ts`.
- Keep or rename existing `hostedGoEngine.ts` for future `/api/engine/...`.
- Implement:
  - `newSession`
  - `loadModel`
  - `runCommand` for `check.induction` and `check.bounded`
  - `getSnapshot`
  - `subscribe`
- Use multipart upload for `loadModel`.
- Normalize:
  - `/arg` with `normalizeGraphPayload`
  - `/concept` with `normalizeConceptPayload`
  - `/toggles` into `toggles-updated`
  - `/check` with `normalizeCheckResult`
- Refresh ARG/concept after load and after check.
- Record local jobs even though the current `/webui` API returns immediate
  responses.

Tests:

- Unit tests with mocked `fetch`:
  - posts to `/api/session/new`
  - sends multipart file on `/load`
  - posts `{mode:"induction"}` to `/check`
  - fetches `/arg` and `/concept`
  - emits `check-updated`, `graph-updated`, `concept-updated`
- SSE unit tests:
  - raw `check_started` becomes progress
  - raw disconnect becomes diagnostic
- Error tests:
  - bad JSON
  - 404 session
  - check failure transport error

Exit criteria:

- `HostedWebuiEngine` can be selected in code and returns normalized records.
- Fake engine remains available for deterministic UI tests.

### Phase 3: Go Integration Test Against Real `webui` Backend

Goal: prove the existing native backend compiles and checks a model through HTTP.

Work:

- Add a Go or Playwright-driven integration fixture that starts
  `webui.NewServer` with `NewGoBackend`.
- Use a known model such as `client_server_example.ivy`.
- Exercise:
  - `POST /api/session/new`
  - `POST /api/session/{id}/load` with multipart text
  - `POST /api/session/{id}/check` with induction mode
  - `GET /api/session/{id}/arg`
  - `GET /api/session/{id}/concept`
- Assert:
  - load succeeds
  - check contacts Z3 when available
  - result is pass/fail/error with message
  - ARG contains at least one node for loaded model
  - concept payload contains relations/toggles

Tests:

- `go test ./goivy/webui -run TestSVKHostedWebuiContract`
- `npm run test:unit -- --run`

Exit criteria:

- The backend contract is pinned by tests before the Svelte UI depends on it.

### Phase 4: Svelte Uses Hosted Backend For Real Checks

Goal: click `Check` in `svk/` and submit the current editor contents to native
Go/Z3 through `HostedWebuiEngine`.

Work:

- Add engine selection state:
  - fake
  - hosted-webui
  - browser-wasm
- In dev mode, allow `hosted-webui` base URL configuration:
  - same-origin when served by webui/svk combined server
  - `VITE_IVY_ENGINE_BASE_URL` for separate dev server
- On editor dirty check:
  - save model text to model store
  - call `engine.loadModel`
  - then `engine.runCommand(check.induction)`
- Update details pane:
  - display latest `CheckResult`
  - no canned failure text once a real check has run
  - show transport/diagnostic errors clearly
- Update status bar:
  - queued/running/succeeded/failed from job/check stores
  - show Z3 contacted state
- Update ARG/concept panes from normalized backend payloads.
- Keep current webui-like layout.

Tests:

- Unit tests for command service:
  - dirty editor triggers load before check
  - clean editor can check directly
  - check result updates details/status selectors
- Playwright with mocked engine:
  - select hosted engine
  - edit model
  - click Check
  - see job running then check result
- Playwright integration with real backend when available:
  - start native `/webui` server
  - start Svelte dev server with base URL
  - click Check
  - assert result text comes from backend

Exit criteria:

- The user can run native `/webui` backend and use the Svelte workbench as the
  frontend for compile/check/Z3.

### Phase 5: Move Hosted Engine Behind `svk/server`

Goal: integrate auth/project policy with the engine without losing the direct
`/webui` adapter.

Work:

- Add engine routes to `goivy/svk/server`:
  - `POST /api/engine/session`
  - `POST /api/engine/session/{id}/load`
  - `POST /api/engine/session/{id}/command`
  - `GET /api/engine/session/{id}/snapshot`
  - `GET /api/engine/session/{id}/events`
- Internally call shared Go engine core or temporarily proxy to `webui.GoBackend`.
- Enforce:
  - authenticated session
  - CSRF for state-changing routes
  - project access
  - session belongs to project/account/user authorization context
- Store project model revisions in Postgres before or alongside load.
- Return normalized engine API records directly where practical.

Tests:

- Go server auth tests:
  - anonymous engine route denied
  - bad CSRF denied
  - wrong project denied
  - valid session can create engine session and load model
- TypeScript adapter tests for future `HostedSvkEngine`.
- One E2E login/dev-session test that reaches app and checks model.

Exit criteria:

- Hosted production path no longer requires exposing raw `/webui` sessions to
  the browser.
- Direct `HostedWebuiEngine` remains useful for dev and parity testing.

### Phase 6: Shared Go Engine Core For Wasm

Goal: make the browser wasm build run the same engine semantics.

Work:

- Extract a transport-neutral engine core from `webui.GoBackend` use.
- Avoid `net/http`, filesystem-only load assumptions, OS-specific process calls,
  and anything incompatible with `GOOS=js GOARCH=wasm` in the core.
- Keep Z3 access behind an interface:
  - native implementation uses native Z3 path
  - browser implementation uses wasm Z3 imports already staged under
    `static/wasm`
- Add a Go wasm command:
  - `goivy/cmd/goivy-webengine-wasm`
  - exports a JSON dispatcher through `syscall/js`
- Ensure no `wasip1` targets or scripts return.
- Add build script:
  - `GOOS=js GOARCH=wasm go build -o goivy/svk/static/wasm/goivy-webengine.wasm ./goivy/cmd/goivy-webengine-wasm`
- Version wasm assets:
  - Go version
  - goivy build hash or source revision if available without relying on git
  - Z3 wasm version

Tests:

- Go native tests for engine core.
- A Node/browser worker test that loads the wasm asset and calls:
  - init
  - new session
  - load model
  - check induction
- Asset test:
  - fail if any build path or manifest contains `wasip1`

Exit criteria:

- Browser worker can execute a real check without a server for at least one
  small Ivy model.

### Phase 7: BrowserWasmEngine Becomes Real

Goal: offline-on-a-plane mode uses the same `IvyEngine` contract.

Work:

- Replace worker stub lifecycle with real wasm loader:
  - load `wasm_exec-go1.25.6.js`
  - instantiate `goivy-webengine.wasm`
  - load wasm Z3 API/wasm as required
  - call Go dispatcher
- Normalize worker responses exactly like hosted responses.
- Implement worker-side event streaming:
  - long check emits progress
  - final check emits `check-updated`
  - graph/concept updates mirror hosted
- Persist active local project/model/session metadata in IndexedDB.
- Make engine selection resilient:
  - if hosted unavailable and user has offline auth/local project, offer browser
    wasm engine
  - if wasm assets missing, show diagnostic and keep model editable

Tests:

- Unit tests for worker protocol parse/serialize.
- Browser tests for asset loading.
- E2E offline simulation:
  - preload app and wasm assets
  - block network
  - reload app
  - open existing local project
  - run check through browser wasm

Exit criteria:

- Previously authenticated user can open the app with no network and run a real
  local check on a plane.

### Phase 8: Reactive UI Completion

Goal: replace remaining hard-coded workbench data with store-driven state.

Work:

- Extract components listed above.
- Replace static concept demo nodes with normalized concept graph rendering.
- Replace static relation checkboxes with `stateRelations` rows.
- Replace details text with selected node/check/job information.
- Wire sheet tabs to actual sheets from ARG/event payloads.
- Wire toolbar buttons to command descriptors.
- Keep the old `webui` visual density and layout.
- Add editor persistence:
  - dirty state
  - save local
  - load file
  - restore from IndexedDB

Tests:

- Component tests for each pane.
- E2E:
  - editor visible on first screen
  - check result updates details/status
  - selecting ARG node updates details/actions
  - relation toggle sends command and updates state
  - layout no-overlap desktop/mobile

Exit criteria:

- The screen is not a mock. All visible operational data comes from normalized
  stores.

### Phase 9: Persistence And Sync

Goal: local-first data survives refresh and syncs to Postgres when online.

Work:

- Persist:
  - projects
  - models
  - model revisions
  - latest graph/check/concept snapshots
  - sync queue
  - offline auth marker/session freshness metadata
- Sync queue operations:
  - save model
  - create project
  - update project metadata
  - upload check summary if desired
- Conflict policy:
  - model revisions are immutable records
  - current model pointer can conflict
  - initial policy is manual conflict copy, not silent merge

Tests:

- IndexedDB migration tests.
- Sync service tests for replay, retry, backoff, idempotency keys.
- E2E:
  - edit offline
  - reload
  - see local changes
  - restore online
  - sync queue drains

Exit criteria:

- Offline work is not just compute; it is durable.

### Phase 10: Hosted Auth/Project Hardening

Goal: make the app server production-shaped enough to support real use.

Work:

- Replace placeholder in-memory `ProjectStore` with Postgres-backed store.
- Complete session persistence with secure opaque session IDs.
- Keep OPAQUE password flow so raw passwords never reach the server.
- Finish explicit identity-linking flows:
  - OAuth
  - passkey
  - magic email
  - OPAQUE account
- Add Mailgun integration tests with fake sender.
- Add CSRF coverage for all state-changing routes.
- Add CSP adjustments for Svelte bundle, wasm, workers, and styles without
  weakening XSS posture.

Tests:

- Go auth route tests.
- Playwright auth smoke:
  - dev login
  - app redirect
  - logout/expired session behavior if implemented
- Security tests:
  - cross-origin POST rejected
  - missing CSRF rejected
  - malicious script-like model text renders as text in UI

Exit criteria:

- Hosted app can safely serve marketing/login/app and protect engine/project
  APIs.

## Concrete First Milestone

The next coding milestone should be deliberately small:

1. Add `checks.svelte.ts` and `check-updated` event.
2. Add `HostedWebuiEngine`.
3. Unit test its calls to:
   - `/api/session/new`
   - `/api/session/{id}/load`
   - `/api/session/{id}/check`
   - `/api/session/{id}/arg`
   - `/api/session/{id}/concept`
4. Add a dev selector in the Svelte UI for `Fake` vs `Hosted webui`.
5. Start `/webui` server manually or from a test fixture.
6. Click `Check` in Svelte and see real backend result in details/status.

This milestone proves the most important architectural claim: the new reactive
Svelte app can drive the real Go/Z3 backend through an adapter, while preserving
the same abstraction needed for browser wasm.

## Risks And Design Guardrails

- Risk: `HostedWebuiEngine` becomes the de facto app contract.
  - Guardrail: endpoint strings stay only in adapter files/tests.
  - Guardrail: components consume stores only.

- Risk: browser wasm tries to mimic HTTP handlers.
  - Guardrail: worker protocol mirrors `IvyEngine`, not `/api/session`.

- Risk: TypeScript starts reimplementing Ivy semantics.
  - Guardrail: TypeScript normalizes and renders; Go checks and analyzes.

- Risk: shared Go core extraction grows too large.
  - Guardrail: first wrap `GoBackend`; extract direct typed internals only as
    needed for wasm compatibility.

- Risk: Z3 wasm and native Z3 diverge.
  - Guardrail: parity harness runs the same model/check against hosted native
    and browser wasm and compares normalized `CheckResult` plus graph shape.

- Risk: old `webui` controller assumptions leak into Svelte.
  - Guardrail: command descriptors and store reductions are the only UI-visible
    command/state surface.

## Definition Of Done For Plan 001

Plan001 is complete when:

- Svelte can run a real native `/webui` compile/check/Z3 operation from the
  editor.
- The result is displayed from normalized `CheckResult`, not canned text.
- ARG/concept/state relation panes are driven by backend or wasm payloads.
- `HostedWebuiEngine` and `BrowserWasmEngine` share the same `IvyEngine`
  frontend contract.
- The browser wasm path runs at least one real Ivy check offline.
- The hosted path can be routed through authenticated `svk/server` project
  APIs.
- Unit, Go, and Playwright tests cover every stage enough that future work does
  not wedge under a huge untested UI/controller migration.
