# Vue-Idiomatic Web UI Migration Plan

## Status

Updated on May 6, 2026 after executing the first cleanup slices from this plan.

The Vue runtime switch-over is complete: the served web UI is a Vue 3 bundle,
Pinia stores exist for user-visible state, and the old `goivy/webui/js_test`
harness has been retired in favor of Vue/Vitest tests.

The Vue-idiomatic migration is not complete. The previous plan was treated as
done too early. Its own Definition of Done was not satisfied because
`goivy/webui/frontend/src/legacyAppController.js` still exists and is still
large.

This document is now the active cleanup plan. Its final implementation step is
to delete `legacyAppController.js` after all behavior has moved into idiomatic
Vue, Pinia, and service modules.

## Definition of Done

The migration is complete when:

- No production component imports `components/legacyCommand.js`.
- No production component calls `callApp()` or `hasAppMethod()` as a proxy for
  controller methods.
- No production module requires `window.ivyApp`, `window.startIvyApp`,
  `window.IvyApp`, `window.IvyControls`, `window.IvyPersist`, or
  `window.IvyGraph`.
- Production boot starts Vue services directly, not a legacy controller.
- Commands are registered by service modules, not by reflecting over
  `IvyApp.prototype`.
- Pinia stores own user-visible and durable UI state.
- Service modules own side effects: backend API calls, browser file handles,
  persistence, CodeMirror, Cytoscape, downloads, dialogs, and analysis-state
  serialization.
- Browser and unit tests pass without constructing `IvyApp`.
- `goivy/webui/frontend/src/legacyAppController.js` is deleted.

## Current State Compared With The Old Plan

Completed baseline:

- Vue 3 is the served frontend shell.
- Pinia stores exist for layout, editor, sheets, dialogs, details, graphs,
  event traces, state relations, session, engine, recent files, toasts,
  context menus, dropdowns, and menu descriptors.
- The command registry exists in `services/commandRegistry.js`.
- The command registry no longer falls back to `window.ivyApp` by default.
- Explicit command registration exists in `services/*Commands.js` and is owned
  by `services/appServices.js`, not by controller reflection during startup.
- Components no longer import `legacyAppController.js` directly.
- Production components use `runUiCommand()` / `hasUiCommand()` rather than
  `callApp()` / `hasAppMethod()`.
- `components/legacyCommand.js` has been deleted.
- Production boot no longer installs `window.IvyApp`, `window.startIvyApp`,
  `window.IvyControls`, `window.IvyPersist`, or `window.IvyGraph`.
- Production startup no longer creates `window.ivyApp`.
- `legacyStartup.js`, `legacyAppRuntime.js`, `legacyScripts.js`,
  `legacyRuntimeGlobals.js`, `legacyGraph.js`, and `legacyPersist.js` have been
  deleted.
- Canonical graph and persistence implementations live in
  `services/graphRuntime.js` and `services/persistenceService.js`.
- Playwright tests use `window.__ivyDiagnostics`, not `window.ivyApp`, for
  transitional graph/editor/runtime probes.
- Many service modules already exist:
  - `fileService.js`
  - `editorService.js`
  - `sessionService.js`
  - `sheetService.js`
  - `graphService.js`
  - `conceptVisibilityService.js`
  - `detailsService.js`
  - `eventTraceService.js`
  - `argActionService.js`
  - `conceptActionService.js`
  - `analysisActionService.js`
  - `checkService.js`
  - `analysisStateService.js`
  - `menuService.js`
  - `dialogService.js`
  - `recentFileService.js`
- Vue/Vitest tests now own the old JavaScript behavior coverage.

Still not done:

- `legacyAppController.js` is still about 5,000 lines and is still production
  runtime code.
- `services/appServices.js` still imports `startIvyApp()` from
  `legacyAppController.js`.
- Most service modules still accept an `app` object shaped like the old
  controller. They need explicit store/service dependencies instead.
- `window.__ivyVueBridge` is still used as a transitional local dependency
  shortcut in several services.
- `window.__ivyDiagnostics` still exposes the runtime object to browser tests.
  That should shrink to narrow diagnostics helpers.
- `legacyAppController.surface.test.js` still inventories a large public
  controller API instead of proving the controller is gone.

## Migration Strategy

Do not shrink the controller by hand first. That creates churn without changing
ownership. Instead:

1. Make each service complete enough to run without an `app` object.
2. Register user commands from those services.
3. Move components and tests to command names or direct store/service calls.
4. Remove controller fallback paths.
5. Delete the legacy runtime globals.
6. Delete the controller.

Every extraction should preserve behavior before deleting old code. Each slice
should run at least:

```sh
npm run test:webui:js
npm run build:webui
```

Run browser tests after changes to startup, graph layout, editor save/load,
menus, sheets, tutorial visibility, or event traces:

```sh
npm run test:webui:browser
```

## Implementation Order

### Phase 1: Freeze The Remaining Legacy Surface

1. Update `legacyAppController.surface.test.js` so it is a shrinkage ratchet.
   - Keep the current list as the starting maximum.
   - Require every future controller method removal to shrink the list.
   - Add a comment that the target list is empty, not a stable API.

2. Add a source guard for production `window.ivyApp` use.
   - Existing component guards are useful but too narrow.
   - Add a Vitest source scan that fails if production modules outside a small
     temporary allowlist reference `window.ivyApp`, `window.startIvyApp`,
     `window.IvyApp`, `window.IvyControls`, `window.IvyPersist`, or
     `window.IvyGraph`.
   - Start with the current allowlist:
     - `legacyAppController.js`
     - `legacyAppRuntime.js`
     - `legacyStartup.js`
     - `legacyRuntimeGlobals.js`
     - `legacyGraph.js`
     - `legacyPersist.js`
     - `services/commandRegistry.js`
     - `services/uiCommandService.js`
     - `resizeDrag.js`
   - Shrink the allowlist in later phases.

3. Add a command registry coverage test for the real production command names.
   - Assert that file, editor, sheet, ARG, concept, event trace, check, and
     analysis-state commands can be registered without constructing `IvyApp`.
   - This test should initially fail or require temporary controller adapters.
     That is the map for the next phases.

### Phase 2: Replace Controller Reflection With Explicit Commands

4. Stop using `registerControllerCommands()` as the primary command source.
   - Keep it temporarily for compatibility tests only.
   - Add explicit command registration modules, grouped by ownership:
     - `services/fileCommands.js`
     - `services/editorCommands.js`
     - `services/sheetCommands.js`
     - `services/graphCommands.js`
     - `services/eventTraceCommands.js`
     - `services/argCommands.js`
     - `services/conceptCommands.js`
     - `services/checkCommands.js`
     - `services/analysisStateCommands.js`
     - `services/menuCommands.js`

5. Introduce a boot-time service container.
   - Add `services/appServices.js`.
   - It creates and wires:
     - engine/session API
     - stores
     - CodeMirror/editor adapter
     - graph runtime factory
     - persistence
     - dialog/toast/menu services
     - command registration
   - The container is not a new God object. It should only wire dependencies
     and expose lifecycle methods such as `start()`, `stop()`, and
     `refreshLayout()`.

6. Change `commandRegistry.js` fallback behavior.
   - Default fallback should be `undefined`, not `window.ivyApp`.
   - Allow tests to inject a fallback explicitly when testing compatibility.
   - Production command execution should fail visibly when a command is not
     registered.

7. Rename `uiCommandService.js` away from app terminology.
   - Replace `callApp()` with `runUiCommand()` or direct `runCommand()`.
   - Replace `hasAppMethod()` with `hasUiCommand()`.
   - Update components to use command language, not app-method language.
   - Delete `components/legacyCommand.js` once no production import remains.

### Phase 3: Make Services Independent Of The Controller Shape

8. Remove `app` object coupling from `editorService.js`.
   - Inputs should be `editorStore`, CodeMirror adapter, and bridge/service
     callbacks.
   - Move remaining editor label, dirty, save sheen, scroll, and keymap logic
     out of controller wrappers.
   - Register editor commands directly.

9. Remove `app` object coupling from `fileService.js`.
   - Replace fields like `_fileHandle`, `_persistedFileName`,
     `_persistedFileContent`, and `_savedFileContent` with explicit state in
     `editorStore`, `sessionStore`, and/or a small file store.
   - Keep browser File System Access API logic here.
   - Register file commands directly.

10. Remove `app` object coupling from `sessionService.js`.
    - Session creation, SSE connection state, and API ownership must be
      available from services/stores.
    - Delete controller `createApi()` and `updateSessionDisplay()` usage after
      callers are moved.

11. Remove `app` object coupling from `sheetService.js`.
    - Make `sheetStore` canonical for active sheet, visual-only sheets, tab
      labels, validity, add/remove/switch, and ARG/event sheet creation.
    - Register sheet commands directly.

12. Remove `app` object coupling from `graphService.js` and `graphRuntime.js`.
    - Move Cytoscape instance ownership and resize/fit behavior into graph
      services.
    - Replace `resizeDrag.js` calls to `window.ivyApp` with injected graph and
      editor layout refresh callbacks.
    - Rename remaining `legacyGraph.js` imports to `services/graphRuntime.js`.
    - Keep a temporary `legacyGraph.js` re-export only while tests are moved.

13. Remove `app` object coupling from `conceptVisibilityService.js`.
    - Make `stateRelationsStore` canonical for edge/label visibility and
      backend toggle hydration.
    - Register state-relation commands directly.

14. Remove `app` object coupling from `detailsService.js`.
    - Details/fact rendering should update `detailsStore` directly.
    - Eliminate controller details/facts wrappers.

15. Remove `app` object coupling from `eventTraceService.js`.
    - Make `eventTraceStore` canonical for trees, selected rows, patterns,
      filter/find state, and event sheet payloads.
    - Register event trace commands directly.

16. Remove `app` object coupling from `argActionService.js`.
    - ARG node/edge actions should depend on session API, graph service,
      sheet service, editor service, and dialog service.
    - Register ARG commands directly.

17. Remove `app` object coupling from `conceptActionService.js`.
    - Concept actions should depend on session API, graph service, visibility
      service, and dialog service.
    - Register concept commands directly.

18. Remove `app` object coupling from `analysisActionService.js`.
    - Generic backend action helpers should depend on session API, graph
      refresh, details, dialogs, and toasts.
    - Register analysis/action commands directly.

19. Remove `app` object coupling from `checkService.js`.
    - Check induction, bounded check, CTI result rendering, and related view
      actions should depend on session API, graph service, sheet service,
      state relations, and details.
    - Register check commands directly.

20. Remove `app` object coupling from `analysisStateService.js`.
    - Build/save/load should consume explicit stores and service snapshots.
    - Validation should be pure functions.
    - Restore should call services, not controller methods.
    - Register analysis-state commands directly.

21. Remove `app` object coupling from `menuService.js`.
    - Dynamic menu descriptors should dispatch through the command registry.
    - Flash/close/dropdown behavior should be Vue store-driven.
    - Register menu commands directly.

22. Remove `app` object coupling from `dialogService.js`.
    - All dialogs should go through `dialogStore`/`DialogHost.vue`.
    - Delete fallback DOM dialog construction from controller once callers are
      moved.

23. Remove `app` object coupling from `legacyPersist.js`.
    - Rename the implementation to `services/persistenceService.js`.
    - Persist explicit state snapshots rather than scraping controller fields.
    - Keep a temporary `legacyPersist.js` re-export only while tests are moved.

### Phase 4: Move Production Boot Off The Legacy Runtime

24. Replace `startLegacyAppWhenReady()` in `App.vue`.
    - `App.vue` should call the new service container lifecycle directly.
    - Production boot should not wait for `window.startIvyApp`.

25. Delete production use of `legacyStartup.js`, `legacyAppRuntime.js`, and
    `legacyScripts.js`.
    - If a compatibility test still needs old globals, move that test-only
      installer into a clearly named test helper.
    - Production bundle should not install `window.IvyApp` or
      `window.startIvyApp`.

26. Remove production use of `window.__ivyVueBridge` where it is just a local
    dependency shortcut.
    - Prefer direct store/service imports or injected dependencies.
    - Keep a small documented browser diagnostics bridge only if Playwright or
      developer tools still need it.
    - The diagnostics bridge must not be required for normal UI behavior.

27. Remove `window.ivyApp` from production.
    - `startIvyApp()` should no longer be called by production.
    - `commandRegistry.js` should have no production fallback to
      `window.ivyApp`.
    - `resizeDrag.js` should use service callbacks, not globals.
    - Source guard allowlist should remove `window.ivyApp` from all production
      modules.

### Phase 5: Migrate Browser Tests Off `window.ivyApp`

28. Add a Playwright diagnostics API that is not the app runtime.
    - Example: `window.__ivyDiagnostics`.
    - It may expose test-only read/command helpers:
      - current session ID
      - graph node/edge counts
      - graph snapshots
      - command execution by command name
      - editor content and cursor helpers
      - sheet/event trace snapshots
    - It must be installed by the Vue/service boot path, not by
      `legacyAppController.js`.

29. Rewrite Playwright tests to use user interactions first.
    - Prefer clicking menus/buttons and asserting visible UI.
    - Use diagnostics only for hard-to-observe graph/editor internals.
    - Remove direct `window.ivyApp` calls from `goivy/webui/pw_test`.

30. Delete compatibility tests that construct `IvyApp`.
    - Replace them with service/component tests or diagnostics tests.
    - Delete `legacyAppController.compat.test.js`.
    - Delete `legacyAppController.surface.test.js` once the controller is gone.
    - Delete `legacySelfBinding.test.js` once no callback-heavy legacy module
      remains.

### Phase 6: Rename Or Delete Remaining Legacy Modules

31. Delete `components/legacyCommand.js`.
    - Production components should import command helpers from the command
      service/registry directly.
    - Rename `components/legacyCommand.test.js` or delete it after equivalent
      command service tests exist.

32. Delete or rename `legacyRuntimeGlobals.js`.
    - `IvyControlsShim` behavior should become service/store behavior.
    - `IvyPersist` behavior should live in `persistenceService.js`.
    - `IvyGraph` behavior should live in `graphRuntime.js`.

33. Delete or rename `legacyGraph.js`.
    - Keep `services/graphRuntime.js` as the canonical Cytoscape runtime
      module.
    - Tests should import the canonical module.

34. Delete or rename `legacyPersist.js`.
    - Keep `services/persistenceService.js` as the canonical persistence
      module.
    - Tests should import the canonical module.

35. Delete `legacyStartup.js`, `legacyAppRuntime.js`, and `legacyScripts.js`.
    - Production startup and tests should use the Vue service boot path.

36. Remove compatibility aliases in package/test naming where useful.
    - `test:webui:js` may stay as a short-term alias, but the preferred name
      should be Vue/Vitest-oriented.
    - Remove empty `goivy/webui/js_test` directories if they still exist.

### Phase 7: Final Guards And Hardening

37. Tighten source guards to final state.
    - No production references to:
      - `legacyAppController`
      - `legacyCommand`
      - `window.ivyApp`
      - `window.startIvyApp`
      - `window.IvyApp`
      - `window.IvyControls`
      - `window.IvyPersist`
      - `window.IvyGraph`
    - Any remaining `legacy` reference must be test-only or documented as
      external compatibility.

38. Update frontend docs.
    - Document:
      - Go-served, build-only Vite workflow
      - Vue component responsibilities
      - Pinia store responsibilities
      - service ownership boundaries
      - command registration conventions
      - browser diagnostics bridge
      - hosted Go engine vs future Wanix-compatible engine interface

39. Confirm the deletion preconditions.
    - No production import of `legacyAppController.js`.
    - No test constructs `IvyApp`.
    - No command registration reflects over `IvyApp.prototype`.
    - No browser test reaches through `window.ivyApp`.

40. Delete `goivy/webui/frontend/src/legacyAppController.js`.
    - This is the final migration step because every controller behavior now
      lives in Vue components, Pinia stores, or service modules.
    - As part of this deletion slice, run:
      - `npm run test:webui:js`
      - `npm run build:webui`
      - `go test ./goivy/webui`
      - `npm run test:webui:browser`
