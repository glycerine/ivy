# Ivy Web UI Frontend

The frontend is a controller-owned DOM application that is bundled by Vite and served by the Go web server. Vite is used for manual builds only; there is no Vite dev server or proxy in the runtime workflow.

## Workflow

- Build bundle: `npm --prefix goivy/webui run build:webui`
- Unit tests: `npm --prefix goivy/webui run test:webui:js`
- Watch unit tests: `npm --prefix goivy/webui run test:webui:js:watch`
- Browser tests: `npm --prefix goivy/webui run test:webui:browser`

The Go server serves `goivy/webui/static/index.html` and the built files under `goivy/webui/static/dist`.

## Ownership Model

- `IvyRuntime` is the coordinator and owns user-visible UI state.
- Static DOM elements in `goivy/webui/static/index.html` expose the controls, panes, dialogs, and graph containers that the runtime binds.
- Services under `src/services` own side effects and non-rendering workflow logic:
  - `commandRegistry.js`: command registration and dispatch.
  - `uiCommandService.js`: event-safe UI command helpers used by components.
  - `sessionService.js`: API/session creation and SSE connection wiring.
  - `editorService.js`: editor content, labels, dirty state, keymaps, and layout refresh.
  - `fileService.js`: browser file handles, save/load/download, and close/new workflows.
  - `persistenceService.js` and `recentFileService.js`: persisted sessions and recent files.
  - `graphRuntime.js` and `graphService.js`: graph runtime access, snapshots, registration, and layout refresh.
  - `conceptVisibilityService.js`: state/relation toggles and concept graph visibility.
  - `detailsService.js`: details and constraint fact callbacks.
  - `eventTraceService.js`: event trace navigation and pattern workflows.
  - `checkService.js`: check/CTI result handling.
  - `analysisActionService.js`: shared backend action execution.
  - `menuService.js`: dropdown/menu dispatch behavior.
  - `analysisStateService.js`: analysis-state serialization and validation.

## Compatibility Surface

`legacyAppController.js` is being reduced to a compatibility facade while services take ownership of behavior. `window.ivyApp` still exists for explicit browser diagnostics and compatibility tests under `frontend/src/legacy*.test.js`.

The old `goivy/webui/js_test` harness has been retired. New frontend tests should live beside the code they exercise under `goivy/webui/frontend/src`.

Command names should be stable, dotted strings when they describe a domain action, for example `file.save`, `file.saveAs`, and `file.reopenLast`.

## Engine Interface

The active Ivy API defaults to the hosted Go API adapter. Future local-first
engines should implement the same consumer-facing `IvyApiAdapter` contract and
may be installed through `window.__IVY_ENGINE__` before app boot.
