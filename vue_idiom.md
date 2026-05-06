# Vue-Idiomatic Web UI Migration Status

## Status

Updated on May 6, 2026 after completing the controller-removal pass.

The Vue 3 switch-over is complete for the production web UI:

- The served UI is the Vue 3 bundle.
- The Go backend serves the built frontend; Vite is build-only.
- The old controller module `goivy/webui/frontend/src/legacyAppController.js`
  has been deleted.
- Production startup flows through `App.vue`, `main.js`, and
  `services/appServices.js`.
- The runtime implementation now lives at
  `goivy/webui/frontend/src/services/ivyRuntime.js`.
- Production command registration is explicit in service-owned command modules.
- The command registry no longer reflects over controller prototypes.
- Production source no longer references `window.ivyApp`, `window.startIvyApp`,
  `window.IvyApp`, `window.IvyControls`, `window.IvyPersist`, or
  `window.IvyGraph`.
- Production source no longer imports or refers to `components/legacyCommand.js`.
- Production source no longer contains `legacy` naming.

## Completed Implementation Order

1. Replaced controller-reflection command registration with explicit service
   command modules.
2. Added the `services/appServices.js` lifecycle container.
3. Changed component command calls from app-method language to
   `runUiCommand()` / `hasUiCommand()`.
4. Deleted `components/legacyCommand.js`.
5. Deleted the old startup/global/script/persistence/graph compatibility
   modules after their behavior moved into canonical services.
6. Added `services/graphRuntime.js` and `services/persistenceService.js` as
   canonical service modules.
7. Added the browser diagnostics bridge under the Vue/service boot path.
8. Removed production creation of `window.ivyApp`.
9. Renamed the remaining runtime from `IvyApp` to `IvyRuntime`.
10. Moved the runtime into `services/ivyRuntime.js`.
11. Deleted `legacyAppController.js`.
12. Renamed production `legacy` identifiers that were no longer accurate.
13. Removed unreachable controller-era DOM code left behind after service
    delegation.
14. Removed `registerControllerCommands()` and its prototype reflection test.
15. Updated boundary tests so old globals and controller names stay gone.

## Current Architecture

- `App.vue` starts the frontend by calling `currentAppServices().start()`.
- `main.js` installs Pinia, app services, the Vue bridge, global interactions,
  and diagnostics.
- `services/appServices.js` owns runtime start/stop and command registration.
- `services/ivyRuntime.js` is the remaining runtime coordinator. It wires the
  API, graph runtimes, CodeMirror, persistence, file handles, menus, dialogs,
  and service calls.
- Service modules own the extracted behavior for files, editor state, sheets,
  graphs, concept visibility, details, event traces, ARG actions, concept
  actions, checks, analysis actions, analysis-state serialization, menus,
  dialogs, recent files, and sessions.
- Pinia stores own user-visible UI state for layout, editor labels, sheets,
  dialogs, details, graphs, event traces, state relations, session status,
  engine selection, recent files, toasts, context menus, dropdowns, and dynamic
  menus.

## Guards

The migration is protected by source and behavior tests:

- Vue components may not import runtime modules or old browser globals.
- Production source may not use the removed browser globals.
- Command registration must be explicit.
- The old controller filename and old `IvyApp` symbols must not reappear.
- Unit tests cover the Vue stores, components, services, command registry,
  engine adapter, runtime boundary, and bridge behavior.

## Follow-On Cleanup

These are cleanup opportunities after the switch-over, not blockers for the
Vue migration:

- Shrink `services/ivyRuntime.js` further as service modules become fully
  dependency-injected.
- Replace broad Playwright diagnostics access with narrower helper methods
  where practical.
- Rename the `test:webui:js` npm alias if we want the command name to be
  Vue/Vitest-specific.
- Continue moving runtime state fields into stores when touching related
  behavior.

## Verification Commands

Run this set after startup, command, runtime, graph, editor, or browser-test
changes:

```sh
npm run test:webui:js
npm run build:webui
go test ./goivy/webui
npm run test:webui:browser
```
