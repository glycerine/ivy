# Ivy Web UI Frontend

The frontend is a Vue 3 + Pinia application that is bundled by Vite and served by the Go web server. Vite is used for manual builds only; there is no Vite dev server or proxy in the runtime workflow.

## Workflow

- Build bundle: `npm --prefix goivy/webui run build:webui`
- Vue/unit tests: `npm --prefix goivy/webui run test:webui:js`
- Watch Vue/unit tests: `npm --prefix goivy/webui run test:webui:js:watch`
- Browser tests: `npm --prefix goivy/webui run test:webui:browser`

The Go server serves `goivy/webui/static/index.html` and the built files under `goivy/webui/static/dist`.

## Ownership Model

- Vue components render UI state and emit user intent.
- Pinia stores hold user-visible state such as editor labels, layout sizes, dialogs, details, graph snapshots, sheets, event traces, and session status.
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

`legacyAppController.js` is being reduced to a compatibility facade while services take ownership of behavior. `window.ivyApp` still exists for explicit browser diagnostics and compatibility tests under `frontend/src/legacy*.test.js`, but Vue components should call services/commands instead of reaching for globals.

The old `goivy/webui/js_test` harness has been retired. New frontend tests should live beside the code they exercise under `goivy/webui/frontend/src`.

Command names should be stable, dotted strings when they describe a domain action, for example `file.save`, `file.saveAs`, and `file.reopenLast`.

## Engine Interface

The active engine is selected through `engineStore`. The hosted Go engine is the default. Alternate engines, including a future Wanix-backed in-browser engine, should implement the same engine methods consumed by `LegacyApiAdapter` and be installed through `window.__IVY_ENGINE__` or `window.__IVY_ENGINE_KIND__` before Vue boot.
