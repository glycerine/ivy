# Webvue Phase 0 Plan

## Goal

Create the clean scaffold for `goivy/webvue` so later phases can implement the
Vue-native Ivy workspace without inheriting the old `goivy/webui` frontend
architecture.

Phase 0 is successful when `webvue` can build, serve, render an empty real
workspace shell, run a small unit suite, and run a browser smoke test.

## Hard Boundaries

- Do not edit `goivy/webui`.
- Do not import or copy old frontend runtime/controller/service code.
- Do not create any central runtime/controller object.
- Do not add `window.ivyApp`, `window.__ivyVueBridge`, or diagnostics that
  expose an app/service container.
- Copying static files is allowed.
- Copying narrow backend-interface facts is allowed:
  - Go API route paths
  - request/response shapes
  - small HTTP fetch mechanics
  - static graph style constants later, when graph work begins

## Phase 0 Deliverables

### Directory Structure

Create this initial shape:

```text
goivy/webvue/
  README.md
  groundup_vue3.md
  phase0.md
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
      components/
        WorkspaceShell.vue
        panes/
          ArgPane.vue
          ConceptPane.vue
          StateRelationsPane.vue
          EditorPane.vue
      stores/
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
    webvue_smoke.spec.mjs
  playwright.config.mjs
```

Use TypeScript from the beginning for `webvue` frontend code.

### Build Setup

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

Create a separate command:

```text
goivy/cmd/ivywebvue/ivywebvue.go
```

Phase 0 should create a `goivy/webvue` Go package with:

```text
server.go
```

The `webvue` server should:

- serve `goivy/webvue/static/index.html` at `/`
- serve `goivy/webvue/static/*` under `/static/`
- expose the same `/api/*` backend routes as `webui`
- reuse the existing Go backend implementation from `goivy/webui` initially
  by importing backend interfaces/types from `github.com/glycerine/ivy/goivy/webui`

This reuse is acceptable because it is the Go/backend contract, not frontend
architecture. The API handler code may be copied in a narrow form only to route
requests to the backend. Do not copy any frontend runtime or browser code.

Command behavior should mirror the old command enough for tests:

```sh
go run ./goivy/cmd/ivywebvue -addr 127.0.0.1:18090
```

No `-open` work is needed in Phase 0 unless it is trivial.

### Initial Vue Shell

The initial UI should be a real workspace skeleton:

- top menubar placeholder
- four visible pane headers:
  - `ARG (Abstract Reachability Graph)`
  - `Concept graph`
  - `State/relations`
  - `Editing:`
- placeholder graph mount regions
- placeholder editor region
- placeholder details strip
- status/session strip

No landing page. No marketing copy. No explanatory in-app text.

### Initial Stores

Create only the stores needed for Phase 0:

- `layoutStore`
  - default column widths
  - tutorial visible flag
  - details height
- `sessionStore`
  - session id
  - status message
  - status level

Stores must not call the backend.

### Bootstrap

Use:

```text
src/app/bootstrap.ts
```

Responsibilities allowed in Phase 0:

- create Pinia
- mount Vue
- install future dependency placeholders through `provide`

Responsibilities forbidden:

- workflow state
- graph/editor/session ownership
- command forwarding
- public app object

Avoid the name `runtime`.

### First Tests

Unit/component tests:

- `App.test.ts`
  - renders the workspace shell
  - shows all four pane headers
- `architecture/sourceGuards.test.ts`
  - fails if production source contains forbidden architecture names

Browser smoke:

- `pw_test/webvue_smoke.spec.mjs`
  - page loads
  - title contains Ivy
  - four pane headers are visible
  - no console errors
  - no forbidden globals exist

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
- function parameters named `app` in service files
- function parameters named `runtime`
- function parameters named `controller`

Allow ordinary Vue words like `createApp` only in `main.ts` or bootstrap code.

### Verification Commands

Phase 0 should end green with:

```sh
npm --prefix goivy/webvue/frontend run test
npm --prefix goivy/webvue/frontend run build
go test ./goivy/webvue
go test ./goivy/cmd/ivywebvue
cd goivy/webvue && ../../node_modules/.bin/playwright test
```

The browser test may require sandbox escalation to bind localhost.

## Implementation Order

1. Copy allowed static assets.
2. Create frontend package/config files.
3. Create new `static/index.html`.
4. Create `webvue.css`.
5. Create Vue shell components and stores.
6. Create source guard test.
7. Create frontend unit tests.
8. Create `goivy/webvue` server package.
9. Create `goivy/cmd/ivywebvue`.
10. Create Playwright config and smoke test.
11. Run verification commands.
12. Update `README.md` with build/test/serve commands.

## Phase 0 Non-Goals

- no CodeMirror integration yet
- no Cytoscape integration yet
- no API client yet beyond Go server route availability
- no file open/save yet
- no check induction yet
- no dynamic menus yet
- no tutorial iframe behavior yet
- no multi-tenant auth yet

Those start in Phase 1 and later.

## Stop Conditions

Stop and revisit the plan if implementation starts to introduce:

- a broad object passed into many services
- a diagnostics API exposing a service container
- copied old frontend service/controller code
- command registration by method reflection
- direct DOM fallback logic for UI that Vue should own
