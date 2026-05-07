# Ground-Up Vue 3 Web UI Plan

## Intent

Build `goivy/webvue` as a fresh Vue 3 application. Treat `goivy/webui` as a
fixed reference implementation and behavior oracle, not as an architecture to
continue.

The goal is not to make the old web UI cleaner. The goal is to create a new
web UI whose architecture is Vue-native from day one.

## Non-Negotiables

- `goivy/webui/` is immutable reference material.
- All new implementation work happens in `goivy/webvue/`.
- No `ivyRuntime`, `IvyApp`, app-controller, bridge-controller, or
  controller-shaped object.
- No single object that owns editor, API, graphs, files, sheets, dialogs,
  menus, and checks.
- No service API that accepts an app-shaped object.
- No production `window.ivyApp`, `window.__ivyVueBridge`, or equivalent hidden
  global app object.
- No copying old frontend runtime/controller/service code and renaming it.
- Copying is allowed only for narrow Go-interface fragments:
  - endpoint paths
  - request/response shapes
  - small HTTP client mechanics
  - Cytoscape style constants or graph payload interpretation when they directly
    represent the backend graph contract
  - static assets and tutorial files
- If copied code has UI control flow, app state, command registration,
  lifecycle orchestration, or DOM fallback behavior, it must not be copied.

## Definition Of Done

`webvue` is complete when:

- The app is served by Go and built with Vue 3.
- The first screen is the real Ivy workspace, not a landing page.
- The four-pane workspace works:
  - ARG (Abstract Reachability Graph)
  - Concept graph
  - State/relations
  - Editing
- The details panel is resizable.
- Tutorial show/hide and pane resizing do not corrupt graph or editor layout.
- Client-server example can be loaded and checked.
- Failed induction shows textual counterexample details.
- Failed induction shows visual CTI/ARG feedback.
- State/relations rows include the CTI relation rows expected from the backend.
- Ctrl-S save status and dirty/saved/saving labels work consistently across
  supported browsers.
- Commands, menus, dialogs, graphs, editor, session, and file state are owned by
  explicit stores/composables/services.
- Browser diagnostics expose narrow helpers only, not an application object.
- Unit and browser tests cover the core behaviors above.
- Auth, billing account, team, project, access-grant, and project-owned Ivy
  data live in the PostgreSQL database `ivyvue`.
- Project-owned rows are scoped by `project_id` and protected with PostgreSQL
  row-level security.

## Architecture

### Auth And Project Persistence

Use PostgreSQL through Go's `database/sql` package. `ivyvue` is the application
database for authentication, sessions, billing accounts, account users, teams,
team memberships, projects, access grants, project storage metadata, and
project-owned Ivy data.

Use a billing-account collaboration model:

- a `User` is a human login identity
- an `Account` is the billing and payment responsibility
- an account can pay for standalone users and multiple teams
- a `Team` is a collaboration group inside an account
- users can be standalone account users, team members, or both
- projects belong to billing accounts
- projects can grant `read`, `write`, or `admin` access to users, teams, or all
  active users in the account

Project-owned tables include `project_id` and use PostgreSQL row-level security
so the database enforces the same project boundary as the application. The
isolation boundary is the project, not the user, team, or billing account. A
single-user trial is modeled as a billing account with one standalone user and
one project, so it gets project-scoped storage without blocking future teams.

Keep a storage-location abstraction from day one. Phase 0 uses
`shared_postgres` in the `project_data` schema. Later, the same abstraction can
support dedicated project databases for large or high-isolation projects.

### Composition Root

Create a small composition root:

- `src/main.ts`
- `src/App.vue`
- `src/app/createIvyApp.ts`

`createIvyApp` wires dependencies and provides them to Vue using Pinia and
Vue `provide/inject`. It must not become an app controller. Its only job is
construction and teardown.

Allowed responsibilities:

- create Pinia
- create the engine client
- create service instances
- register command handlers
- install global keyboard/mouse listeners
- expose narrow diagnostics in development/test mode

Forbidden responsibilities:

- owning mutable workflow state
- implementing file, graph, check, menu, editor, or dialog behavior
- forwarding hundreds of methods
- storing graph/editor/session objects on itself as public API

### Stores

Use Pinia for user-visible and durable UI state:

- `sessionStore`
  - session id
  - connection status
  - user/account/team/project/workspace identity placeholders
  - status text and severity
- `editorStore`
  - path
  - buffer text
  - saved text hash or saved snapshot
  - dirty/saving/saved state
  - keymap
  - browser file handle capability state
- `layoutStore`
  - column widths
  - details height
  - tutorial visibility
  - split drag state
- `sheetStore`
  - active sheet id
  - sheet list
  - sheet type and label
- `graphStore`
  - ARG graph payloads by sheet
  - concept graph payloads by sheet
  - selected ARG node
  - selected concept node
  - graph fit/refresh epochs
- `relationsStore`
  - relation rows
  - edge/label checkbox state
  - selected state label
- `detailsStore`
  - short details
  - long details
  - CTI/check details
  - constraint facts
  - trace action availability
- `menuStore`
  - static top-level menus
  - backend dynamic menu descriptors
  - open dropdown state
- `dialogStore`
  - modal descriptor
  - pending promise resolution
- `toastStore`
  - non-modal transient notices
- `eventTraceStore`
  - event trace sheets
  - selected event row
  - expanded event rows
  - event pattern list/filter/find state
- `recentFilesStore`
  - recent browser file references where supported

No store may call the backend directly. Stores mutate state. Services perform
effects.

### Services And Composables

Create focused services with explicit dependencies:

- `api/ivyHttpClient.ts`
  - small fetch wrapper
  - copied endpoint mechanics allowed from old webui
- `engines/hostedGoEngine.ts`
  - direct Go backend interface
  - endpoint paths and payload shapes may be copied
- `engines/wanixEngine.ts`
  - stub-compatible implementation behind the same engine interface
- `services/sessionService.ts`
  - create session
  - subscribe to SSE
  - reconnect/connection-lost behavior
- `services/modelFileService.ts`
  - load model from browser file
  - save/download model
  - File System Access API
  - Firefox/Chrome save differences
- `services/editorService.ts`
  - CodeMirror adapter
  - keymap changes
  - cursor/line highlighting
  - redraw/refresh after layout changes
- `services/graphService.ts`
  - Cytoscape creation
  - graph payload rendering
  - fit/resize scheduling
  - graph event normalization
- `services/sheetService.ts`
  - open/close/switch sheets
  - ARG/event trace sheet creation
- `services/relationService.ts`
  - hydrate relation rows
  - apply edge and node label visibility
  - persist toggles to backend
- `services/checkService.ts`
  - run induction/bounded checks
  - render check result into stores
  - create CTI ARG sheet/action
- `services/actionService.ts`
  - backend action runner
  - undo/redo/domain/reachability actions
- `services/conceptActionService.ts`
  - split/suppose-empty/materialize/remove/projection/splatter
- `services/argActionService.ts`
  - ARG node/edge backend actions
  - view-source line highlight
  - step-in sheet creation
- `services/menuService.ts`
  - static menus
  - backend descriptor menus
  - command dispatch
- `services/dialogService.ts`
  - promise-based dialog API backed by `dialogStore`
- `services/eventTraceService.ts`
  - load/render/filter/find event traces
- `services/analysisStateService.ts`
  - save/load workspace state

Each service constructor receives only the dependencies it needs. No service
receives `app`, `runtime`, `controller`, or a broad container.

### Commands

Use an explicit command registry:

- command names are strings
- handlers are registered from focused services
- no method-name reflection
- no command target object

Example:

```ts
registerCommand('file.save', () => modelFileService.save())
registerCommand('check.induction', () => checkService.checkInduction())
registerCommand('graph.fitConcept', () => graphService.fit('concept'))
```

Components may call commands for menu/toolbar actions. Components may call
store actions for local state changes. Components must not call backend engines
directly.

### Components

Build components around state and emitted user intent:

- `WorkspaceShell.vue`
  - layout grid and splitters
- `ArgPane.vue`
  - header
  - menu strip
  - Cytoscape mount point
  - details splitter participation
- `ConceptPane.vue`
  - header
  - dynamic menu strip
  - Cytoscape mount point
- `StateRelationsPane.vue`
  - state label
  - relation checkbox table
- `EditorPane.vue`
  - Editing header
  - keymap controls
  - CodeMirror mount
  - save sheen
- `DetailsPane.vue`
  - shared bottom details region
  - check/CTI details
  - trace action button
- `TabBar.vue`
  - sheet tabs
- `EventTraceSheet.vue`
  - event tree and pattern controls
- `TutorialPane.vue`
  - tutorial frame/url bar controls
- `Menubar.vue`
  - File / Mode / Check / tutorial controls
- `DynamicMenuRegion.vue`
  - backend descriptors
- `ContextMenuHost.vue`
  - graph/context menu actions
- `DialogHost.vue`
  - promise dialogs
- `ToastHost.vue`
  - transient notices
- `FileInputHost.vue`
  - hidden file inputs only

Components should not contain backend endpoint knowledge.

## Directory Plan

```text
goivy/webvue/
  frontend/
    package.json
    vite.config.mjs
    vitest.config.mjs
    src/
      main.ts
      App.vue
      app/
        createIvyApp.ts
        injectionKeys.ts
      api/
        ivyHttpClient.ts
      engines/
        ivyEngine.ts
        hostedGoEngine.ts
        wanixEngine.ts
      stores/
      services/
      commands/
      components/
      composables/
      graph/
      test/
  static/
    index.html
    css/
    dist/
    tutorial/
    favicon files
  pw_test/
  playwright.config.mjs
  webui2_server.go or equivalent Go entry integration
```

Use TypeScript for the new app. Keep types small and local at first:

- `types/backend.ts`
- `types/graph.ts`
- `types/sheets.ts`
- `types/dialogs.ts`

TypeScript is useful here because backend payload boundaries and service
dependencies are the places we most need precision.

## Copy Policy

Allowed to copy:

- static favicon/tutorial assets
- CSS variables or low-level visual constants, then prune
- endpoint paths and request/response mechanics from the old engine client
- graph style constants if they directly represent backend Cytoscape classes
- Playwright scenarios as behavior descriptions, rewritten to avoid runtime
  reach-through
- small test fixtures and example Ivy source text

Not allowed to copy:

- `ivyRuntime.js`
- command modules that register methods from a broad target object
- bridge modules that synchronize old DOM/controller state into Vue
- services that accept an app-shaped object
- old direct-DOM fallback implementations
- diagnostics that expose a runtime/app object
- tests whose main mechanism is `window.__ivyDiagnostics.runtime()`

If a copied fragment grows beyond a narrow backend/adapter concern, stop and
rewrite it as a native service.

## Implementation Order

### Phase 0: Scaffold

1. Create `webvue/frontend/package.json`, Vite build config, Vitest config.
2. Create `webvue/static/index.html` loading `/static/dist/ivywebvue.js`.
3. Copy static favicon/tutorial assets.
4. Add a minimal Go serving path for webvue or document the command that serves
   it during development.
5. Build an empty Vue app that renders the workspace shell.

Gate:

- `npm --prefix goivy/webvue/frontend run build`
- browser loads `webvue` shell with no console errors

### Phase 1: Engine Boundary

1. Implement `IvyEngine` interface.
2. Implement `HostedGoEngine`.
3. Implement `IvyHttpClient`.
4. Add unit tests for:
   - session creation
   - model load
   - ARG fetch
   - concept fetch
   - check
   - action
   - toggles
   - SSE subscription
5. Implement `sessionService` with explicit `engine` and `sessionStore`.

Gate:

- no UI component imports the engine directly
- no global app object exists

### Phase 2: Workspace Layout

1. Implement `layoutStore`.
2. Implement `WorkspaceShell`.
3. Implement four titled panes:
   - ARG (Abstract Reachability Graph)
   - Concept graph
   - State/relations
   - Editing
4. Implement splitters:
   - column resizing
   - details pull-up behavior
   - editor left-edge drag expands to browser right edge
5. Implement tutorial show/hide shell.

Gate:

- Playwright verifies pane titles and resizing
- no graph/editor behavior yet required

### Phase 3: Editor And Files

1. Implement `editorStore`.
2. Implement `EditorPane`.
3. Implement CodeMirror mount composable.
4. Implement `modelFileService`.
5. Implement dirty label:
   - `Editing: path`
   - `**` when dirty
   - `[saving...]` while saving
   - `[saved]` after save
6. Implement gray save sheen over editor text while saving.
7. Implement browser save behavior with Firefox fallback and no Chrome
   regression.
8. Implement new/open/close/reopen/download/save-as commands.

Gate:

- unit tests for label state
- Playwright save smoke test
- no service accepts app/runtime/controller

### Phase 4: Graphs

1. Implement `graphStore`.
2. Implement `graphService`.
3. Mount Cytoscape from Vue components.
4. Render backend ARG payload.
5. Render backend concept payload.
6. Implement graph fit after:
   - load
   - pane resize
   - tutorial hide/show
   - tab switch
7. Implement graph context menu host and selected-node state.

Gate:

- client-server example graph displays
- tutorial hide/show keeps concept graph centered
- graph service owns Cytoscape instances; no component reaches into global app

### Phase 5: State Relations And Details

1. Implement `relationsStore`.
2. Implement `StateRelationsPane`.
3. Hydrate rows from concept payload relations/node labels.
4. Implement edge/node-label toggles.
5. Persist toggles through backend.
6. Implement `detailsStore`.
7. Implement `DetailsPane`.
8. Implement constraint facts and fact selection.

Gate:

- CTI/check concept data adds X/Y/Z rows as expected
- relation toggles affect concept edge rendering

### Phase 6: Menus And Commands

1. Implement command registry.
2. Register commands directly from focused services.
3. Implement top menubar.
4. Implement backend dynamic menu descriptors.
5. Implement context menu dispatch.
6. Implement mode selection.

Gate:

- no command target object
- no method reflection
- backend descriptor menu renders and dispatches

### Phase 7: Checks And CTI

1. Implement `checkService`.
2. Implement induction and bounded check commands.
3. Render textual result details into `detailsStore`.
4. Render counterexample trace details.
5. Render CTI visual feedback:
   - ARG sheet/graph update
   - trace ARG sheet action
   - selected state/concept graph update
6. Match the known client-server failure behavior.

Gate:

- loading `client_server_example.ivy` and pressing Check Induction shows:
  - visible failed status
  - detailed textual counterexample
  - visual ARG/concept feedback
  - state relation rows including CTI variables

### Phase 8: ARG/Concept Actions

1. Implement ARG node click concept reload.
2. Implement ARG node right-click actions.
3. Implement ARG edge actions including view-source line highlight.
4. Implement step-in sheets.
5. Implement concept node/edge actions:
   - split
   - suppose empty
   - remove
   - materialize node/edge
   - projection
   - splatter
6. Implement undo/redo/domain/reachability actions.

Gate:

- Playwright covers main action paths without diagnostics reaching into an app
  object

### Phase 9: Event Traces And Analysis State

1. Implement event trace sheets.
2. Implement event tree expansion/selection.
3. Implement event pattern add/remove/filter/find/load/save.
4. Implement analysis-state save/load.
5. Ensure visual-only restored sheets behave correctly.

Gate:

- event trace sheet renders from backend/test data
- analysis state round trip works

### Phase 10: Diagnostics And Tests

1. Add `window.__ivyWebvueDiagnostics` only for tests/dev.
2. Expose narrow helpers:
   - `sessionId()`
   - `runCommand(name, ...args)`
   - `graphSnapshot(kind, sheetId?)`
   - `graphNodeCount(kind, sheetId?)`
   - `editorContent()`
   - `setEditorContent(text)` if needed
   - `loadExample(name)` if needed
   - `sheetSnapshot()`
3. Do not expose:
   - app object
   - service container
   - Cytoscape instances directly unless behind narrow graph helpers
4. Port browser tests as behavior tests, not runtime reach-through tests.

Gate:

- source guard fails on:
  - `ivyRuntime`
  - `IvyApp`
  - `window.ivyApp`
  - `__ivyVueBridge`
  - `__ivyDiagnostics.runtime`
  - `registerMethodCommands`
  - service function signatures accepting `app` or `runtime`

## Test Plan

Unit tests:

- stores
- engine client
- services with fake dependencies
- command registration
- dialog promises
- editor label/save state
- relation toggle mapping
- check result mapping
- graph payload normalization

Component tests:

- pane headers
- editor header/dirty/saving/saved states
- relation rows and checkbox emits
- details trace action
- dialog host
- context menu host
- event trace sheet

Browser tests:

- page loads
- four panes visible with correct titles
- resize details panel
- resize editor left edge
- hide/show tutorial redraws graph/editor
- load client-server example
- check induction failure shows text and graph feedback
- relation toggles affect graph rendering
- save status behavior
- dynamic menu descriptor dispatch
- ARG/concept action smoke paths
- no console errors

## Initial Milestone Target

The first useful vertical slice is:

1. scaffold app
2. create session
3. editor shell
4. load client-server example through backend
5. render ARG and concept graph
6. run check induction
7. show failed details and CTI visual feedback

Do this before implementing every menu/action. It proves the architecture can
carry the hardest user-visible behavior without a controller.

## Anti-Cheating Ratchets

Add these early and keep them green:

- No file named or containing:
  - `ivyRuntime`
  - `legacyAppController`
  - `IvyApp`
- No production reference to:
  - `window.ivyApp`
  - `window.__ivyVueBridge`
  - `__ivyDiagnostics.runtime`
- No command registration helper that reflects over method names.
- No service function whose first parameter is named `app`, `runtime`,
  `controller`, or `commandTarget`.
- No service file over 700 lines without an explicit exception in this plan.
- No component imports backend engine modules directly.

These ratchets are how `webvue` avoids becoming another renamed controller.
