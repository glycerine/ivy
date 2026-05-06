# Vue-Idiomatic Web UI Migration Plan

## Purpose

The Vue 3 switch-over is complete in the served/runtime sense: the browser loads the Vue/Vite bundle, the old `static/js/ivyweb_*.js` scripts are gone, and the UI is rendered through Vue components and Pinia stores.

The remaining smell is architectural: `goivy/webui/frontend/src/legacyAppController.js` is still a large imperative compatibility controller. It owns backend orchestration, file persistence, graph wiring, editor state, menus, event traces, analysis-state serialization, dialogs, and many command handlers. Vue is currently the shell and state surface; the controller is still the command brain.

This plan describes how to turn that compatibility island into idiomatic Vue/Pinia without another risky big-bang rewrite.

## Target Architecture

The end state should have these properties:

- Vue components render state and emit user intent.
- Pinia stores own user-visible state and durable UI state.
- Small services/composables own side effects:
  - backend API calls
  - browser file handles
  - persistence
  - CodeMirror integration
  - Cytoscape graph instances
  - downloads/uploads
- Commands are plain functions/actions, not methods on `window.ivyApp`.
- `window.ivyApp` is removed or reduced to a temporary debug-only compatibility facade.
- The `legacy*` modules are either gone or renamed to `compat*` while they are still needed.
- Browser tests guard the actual served runtime path, not an alternate static implementation.

## Current Compatibility Island

Main files involved:

- `goivy/webui/frontend/src/legacyAppController.js`
- `goivy/webui/frontend/src/legacyAppRuntime.js`
- `goivy/webui/frontend/src/legacyScripts.js`
- `goivy/webui/frontend/src/legacyRuntimeGlobals.js`
- `goivy/webui/frontend/src/legacyGraph.js`
- `goivy/webui/frontend/src/legacyPersist.js`
- `goivy/webui/frontend/src/components/legacyCommand.js`
- `goivy/webui/frontend/src/ivyVueBridge.js`

Current global/compatibility affordances:

- `window.ivyApp`
- `window.IvyApp`
- `window.startIvyApp`
- `window.IvyAPI`
- `window.IvyControls`
- `window.IvyPersist`
- `window.IvyGraph`
- `window.__ivyVueBridge`

Some of these may remain briefly as migration shims. The plan below removes them in dependency order.

## Constraints

- Keep Go as the only server and runtime host.
- Do not introduce a Vite dev server or proxy workflow.
- Keep Vite as build-only bundling unless the project direction changes.
- Avoid React.
- Keep the working browser behavior stable after every slice.
- Keep the old JS test coverage, but keep it pointed at the bundled frontend modules.
- Prefer small, reversible extraction steps over a second big switch-over.

## High-Level Dependency Graph

The controller can be split safely only after the command path is no longer hardwired to `window.ivyApp`.

Recommended dependency order:

1. Command bus/service layer.
2. Session/API service.
3. Editor and file workflow.
4. Persistence and recent files.
5. Graph instance management.
6. State relations and concept visibility.
7. Sheet and event-trace workflows.
8. Menus/actions/check workflows.
9. Analysis-state serialization.
10. Delete compatibility globals and rename/remove compatibility modules.

Do not start by deleting `legacyAppController.js`. First surround it with proper service interfaces, then move one behavior family at a time.

## Implementation Order

### Phase 0: Establish Guardrails

1. Create an inventory test for controller surface area.
   - Add a unit test that lists public methods still used from `legacyAppController.js`.
   - This is not to freeze the controller forever; it makes shrinkage visible.
   - Dependency: none.
   - Verification: `npm run test:webui:js`.

2. Add a Playwright guard that `window.ivyApp` is not used by Vue components directly.
   - Components should call command services, not globals.
   - Temporary exception: command service may still delegate to `window.ivyApp`.
   - Dependency: none.
   - Verification: browser suite.

3. Rename intent in docs/tests without behavior change.
   - Keep files as-is initially, but document that `legacy` now means compatibility.
   - Avoid churny renames until service boundaries are in place.
   - Dependency: none.
   - Verification: no code required.

### Phase 1: Introduce a Vue Command Registry

4. Add `frontend/src/services/commandRegistry.js`.
   - Exposes `registerCommand(name, fn)`, `runCommand(name, args)`, `hasCommand(name)`.
   - Internally may delegate unknown commands to `window.ivyApp` for now.
   - Dependency: Phase 0.
   - Verification: new unit tests.

5. Replace `components/legacyCommand.js` internals with the command registry.
   - Keep exported function names for component compatibility.
   - Components stop knowing about `window.ivyApp`.
   - Dependency: step 4.
   - Verification: existing Vue component tests.

6. Register controller-backed commands during boot.
   - `legacyAppRuntime.js` or `legacyStartup.js` can register adapter commands after creating the controller.
   - This is still compatibility, but hidden behind a proper command API.
   - Dependency: step 5.
   - Verification: browser suite.

### Phase 2: Extract Backend Session/API Ownership

7. Create `services/sessionService.js`.
   - Wraps engine/API creation, session creation, SSE connection, connection-lost handling.
   - Uses `engineStore` and `sessionStore`.
   - Dependency: command registry can exist but is not strictly required.
   - Verification: unit tests for session creation and SSE lost connection.

8. Move `createApi()`, session creation, and `updateSessionDisplay()` out of the controller.
   - Controller delegates to `sessionService`.
   - `sessionStore` becomes canonical for current session ID/status.
   - Dependency: step 7.
   - Verification: `npm run test:webui:all`.

9. Update `ivyVueBridge.createLegacyApi()` consumers.
   - New code should use `sessionService.api` or `engineStore.engine`.
   - Keep bridge method temporarily for compatibility tests.
   - Dependency: step 8.
   - Verification: API wiring tests.

### Phase 3: Extract Editor Ownership

10. Create `services/editorService.js`.
   - Owns CodeMirror instance lookup, editor content get/set, dirty calculation, label rules, layout refresh.
   - Uses `editorStore`.
   - Dependency: command registry.
   - Verification: migrate existing editor-label and save-progress tests.

11. Move `_editorContent()`, `_editorDirty()`, `_updateEditorLabel()`, `_refreshEditorLayout()`, `setEditorContent()`, `scrollEditorToLine()`.
   - Controller delegates to editor service.
   - Dependency: step 10.
   - Verification: editor unit tests and source-view browser test.

12. Remove controller direct knowledge of CodeMirror initialization.
   - `codeMirrorEditor.js` becomes the only CodeMirror setup path.
   - Controller receives editor service methods instead of touching `cmEditor`.
   - Dependency: step 11.
   - Verification: editor tests, browser suite.

### Phase 4: Extract File Save/Load Workflow

13. Create `services/fileService.js`.
   - Owns browser file handles, load model, save, save-as, download fallback, external-change detection, merge conflict markers.
   - Uses `editorService`, `sessionService`, `persistenceService`, `dialogStore`, `toastStore`.
   - Dependency: Phases 2 and 3.
   - Verification: migrate `ivyweb_app_save.test.mjs` to this service.

14. Move file-handle methods:
   - `_ensureFileHandleWritable()`
   - `_readFileHandleContent()`
   - `_restoreFileHandleForCurrentFile()`
   - `_confirmNoExternalChangeBeforeSave()`
   - `_mergeDiskVersionIntoEditBuffer()`
   - `_rememberLastOpenFile()`
   - `_updateReopenLastFileButton()`
   - `reopenLastFile()`
   - Dependency: step 13.
   - Verification: save tests.

15. Move user-facing file actions:
   - `chooseAndLoadModelFile()`
   - `loadFile()`
   - `save()`
   - `saveAs()`
   - `downloadModel()`
   - `downloadTextFile()`
   - `downloadModelForUnsupportedSave()`
   - `closeCurrentFile()`
   - `newModel()`
   - Dependency: step 14.
   - Verification: file input tests, save tests, browser load/save smoke if available.

16. Register file commands in command registry.
   - `file.load`, `file.save`, `file.saveAs`, `file.download`, `file.close`, `file.new`, `file.reopenLast`.
   - Components call command names instead of controller methods.
   - Dependency: step 15.
   - Verification: Menubar and FileInputHost tests.

### Phase 5: Extract Persistence and Recent Files

17. Rename `legacyPersist.js` to `persistenceService.js`.
   - Keep a temporary re-export from `legacyPersist.js` if tests/imports need gradual migration.
   - Dependency: file service exists.
   - Verification: persistence tests.

18. Make persistence independent of controller shape.
   - Save/load plain objects from stores/services instead of scraping `app` fields.
   - Inputs should be explicit:
     - session
     - file metadata/content
     - graph snapshots
     - toggles
     - sheet/event state
   - Dependency: steps 13-17.
   - Verification: persistence tests and browser reload/persistence path if available.

19. Move `populateRecentFiles()` and `loadRecentSession()` into `recentFileService.js`.
   - Uses `recentFilesStore` and `persistenceService`.
   - Dependency: step 18.
   - Verification: recent files tests.

### Phase 6: Extract Graph Runtime Ownership

20. Rename `legacyGraph.js` to `graphRuntime.js`.
   - Keep temporary re-export from `legacyGraph.js`.
   - Dependency: none, but easier after file/persistence snapshots become explicit.
   - Verification: graph tests.

21. Create `services/graphService.js`.
   - Owns graph instance creation, graph registry by sheet ID, Cytoscape snapshots, graph resize/fit, layout refresh.
   - Uses `graphStore` and `sheetStore`.
   - Dependency: command registry and stores already exist.
   - Verification: graph service unit tests and existing graph browser tests.

22. Move graph registry methods:
   - `currentSheet()`
   - `registerSheet()`
   - `installGraphStoreHook()`
   - `graphElementsSnapshot()`
   - `_refreshGraphsAndEditorLayout()`
   - `_refreshLayoutAfterVuePatch()`
   - Dependency: step 21.
   - Verification: sheet graph browser tests.

23. Move graph event attachment into graph service.
   - `attachGraphEventHandlers()`
   - `onArgNodeClick()`
   - `onArgNodeRightClick()`
   - `onArgEdgeRightClick()`
   - `onConceptNodeRightClick()`
   - `onConceptEdgeRightClick()`
   - Dependency: step 22 and command registry.
   - Verification: context-menu and graph-click browser tests.

### Phase 7: Extract State Relations and Concept Visibility

24. Create `services/conceptVisibilityService.js`.
   - Owns edge/label visibility maps and applies classes/labels to graph instances.
   - Uses `stateRelationsStore`.
   - Dependency: graph service.
   - Verification: state relation toggle tests.

25. Move:
   - `populateStateCheckboxes()`
   - `_hydrateBackendToggleState()`
   - `_toggleChecked()`
   - `onEdgeToggle()`
   - `onEdgeToggleChange()`
   - `onLabelToggleChange()`
   - `_applyEdgeVisibility()`
   - `_findEdgeVisibility()`
   - `_applyNodeLabels()`
   - `_displayConceptName()`
   - `updateStateLabel()`
   - Dependency: step 24.
   - Verification: `ivyweb_persist_state_relations.test.mjs`, browser relation-toggle test.

26. Move `populateConstraintFacts()` into `detailsService.js`.
   - Details store becomes canonical for details and fact action callbacks.
   - Dependency: step 24 not strict, but same data source.
   - Verification: constraint facts tests.

### Phase 8: Extract Sheet and Event Trace Workflows

27. Create `services/sheetService.js`.
   - Owns sheet creation/removal/switching, visual-only sheets, active sheet.
   - Uses `sheetStore`, `graphService`, `eventTraceStore`.
   - Dependency: graph service.
   - Verification: tab and sheet tests.

28. Move sheet methods:
   - `isVisualOnlySheet()`
   - `setVisualOnlySheet()`
   - `visualOnlyMessage()`
   - `isValidSheetId()`
   - `assertValidSheetId()`
   - `sheetTab()`
   - `sheetExists()`
   - `switchSheet()`
   - `addSheet()`
   - `removeSheet()`
   - `openARGSheet()`
   - Dependency: step 27.
   - Verification: sheet tests and ARG sheet browser tests.

29. Create `services/eventTraceService.js`.
   - Owns event trace loading, tree operations, patterns, filtering/finding.
   - Uses `eventTraceStore` and `sheetService`.
   - Dependency: sheet service.
   - Verification: event trace unit tests and browser event-trace tests.

30. Move event trace methods:
   - `openEventTraceSheet()`
   - `loadEventTraceFile()`
   - `readFileText()`
   - `renderEventTraceSheet()`
   - `renderEventTree()`
   - `renderEventTreeNode()`
   - `attachEventTraceHandlers()`
   - `toggleEventTraceNode()`
   - `lookupEventTrace()`
   - `uncoverEventTraceAddress()`
   - `selectEventTraceRow()`
   - `activeEventSheet()`
   - `filterEventTrace()`
   - `findEventTrace()`
   - `applyEventPatternResult()`
   - `renderEventPatternList()`
   - `selectedEventPattern()`
   - `addEventPattern()`
   - `removeSelectedEventPattern()`
   - `clearEventPatterns()`
   - `loadEventPatterns()`
   - `saveEventPatterns()`
   - Dependency: step 29.
   - Verification: event trace tests.

### Phase 9: Extract Action Workflows

31. Create `services/argActionService.js`.
   - Owns ARG node/edge action descriptors and execution.
   - Uses `sessionService`, `graphService`, `sheetService`, `dialogStore`.
   - Dependency: graph and sheet services.
   - Verification: ARG choice-backed command tests.

32. Move:
   - `executeArgNodeAction()`
   - `prepareArgNodeActionArgs()`
   - `executeArgEdgeAction()`
   - Dependency: step 31.
   - Verification: ARG action tests and browser tests.

33. Create `services/conceptActionService.js`.
   - Owns concept split/materialize/splatter/projection/remove/empty workflows.
   - Uses `sessionService`, `graphService`, `conceptVisibilityService`, `dialogStore`.
   - Dependency: graph and visibility services.
   - Verification: materialize/splatter/export tests.

34. Move:
   - `executeConceptNodeAction()`
   - `executeConceptEdgeAction()`
   - `splitConcept()`
   - `supposeEmpty()`
   - `removeConcept()`
   - `materializeNode()`
   - `materializeEdge()`
   - `addProjection()`
   - `selectConceptNode()`
   - `materializeEdgeFromSelected()`
   - `splatterNode()`
   - Dependency: step 33.
   - Verification: concept action tests and browser context menu tests.

35. Create `services/analysisActionService.js`.
   - Owns general backend action runners and graph refresh.
   - Uses `sessionService`, `graphService`, `detailsService`, `toastStore`.
   - Dependency: session, graph, visibility services.
   - Verification: shared action runner tests.

36. Move:
   - `runAction()`
   - `doUndo()`
   - `doRedo()`
   - `resetDomain()`
   - `diagramDomain()`
   - `refreshConceptGraph()`
   - `pdrStep()`
   - `showReachableStates()`
   - `concreteStep()`
   - `gatherFacts()`
   - `ctiConceptAction()`
   - `reverseStep()`
   - `pathReach()`
   - `reachStep()`
   - `makeConjecture()`
   - `backtrack()`
   - `recalculateGraph()`
   - `rememberGraph()`
   - `addRelationFromString()`
   - `refreshAfterLoad()`
   - Dependency: step 35.
   - Verification: action tests and browser suite.

### Phase 10: Extract Check/CTI Workflow

37. Create `services/checkService.js`.
   - Owns check induction/bounded check/result rendering/trace actions.
   - Uses `sessionService`, `detailsStore`, `sheetService`, `graphService`.
   - Dependency: sheet/graph/action services.
   - Verification: CTI tests and browser failed-check trace test.

38. Move:
   - `runCheck()`
   - `_autoCheckUsedRelations()`
   - `showCheckResult()`
   - `addCheckResultViewActions()`
   - `checkInduction()`
   - `boundedCheck()`
   - `weakenInvariant()`
   - Dependency: step 37.
   - Verification: CTI text/ARG browser tests.

### Phase 11: Extract Menus and Dialogs

39. Move dropdown/menu descriptor logic into `menuService.js`.
   - `loadMenuDescriptors()`
   - `renderMenuRegion()`
   - `renderMenuDescriptor()`
   - `dispatchMenuDescriptorAction()`
   - `setupDropdownMenus()`
   - `closeAllDropdowns()`
   - `flashAndClose()`
   - `bindMenuAction()`
   - Dependency: command registry and action services.
   - Verification: menu descriptor tests and browser menu tests.

40. Remove fallback DOM dialog builder from controller.
   - Delete or quarantine:
     - `_createDialog()`
     - `_setDialogError()`
     - `_addDialogButton()`
     - `_finishDialog()`
     - `_installDialogEscape()`
   - Keep high-level helpers as thin wrappers over `dialogStore` temporarily:
     - `okDialog()`
     - `okCancelDialog()`
     - `textDialog()`
     - `showTextDialog()`
     - `entryDialog()`
     - `integerDialog()`
     - `listboxDialog()`
     - `buttonListDialog()`
   - Dependency: all action services use `dialogStore` directly.
   - Verification: dialog tests.

41. Move remaining high-level dialog wrappers to `dialogService.js`.
   - Components/services import `dialogService`, not controller.
   - Dependency: step 40.
   - Verification: dialog tests.

### Phase 12: Extract Analysis-State Serialization

42. Create `services/analysisStateService.js`.
   - Owns save/load JSON shape, validation, and restoring sheet/graph/event state.
   - Uses explicit dependencies:
     - `editorService`
     - `fileService`
     - `graphService`
     - `sheetService`
     - `eventTraceService`
     - `stateRelationsStore`
   - Dependency: most previous services.
   - Verification: existing analysis-state tests.

43. Move:
   - `buildAnalysisState()`
   - `saveAnalysisState()`
   - `loadAnalysisStateFile()`
   - `loadAnalysisStateObject()`
   - `analysisStateLimits()`
   - `validateAnalysisStateObject()`
   - `validateAnalysisStateSheet()`
   - `validateAnalysisStateGraphPayload()`
   - `validateAnalysisStateEvents()`
   - `removeAnalysisStateExtraSheets()`
   - Dependency: step 42.
   - Verification: analysis-state tests and browser reachable/trace tests.

### Phase 13: Remove Compatibility Controller

44. Replace `legacyAppController.js` with a small facade.
   - At this point it should contain no logic, only command registration or backwards-compatible method aliases.
   - Dependency: all method families extracted.
   - Verification: full test suite.

45. Replace `window.ivyApp` calls.
   - Component code should use command registry/services.
   - Tests should not need `window.ivyApp` except for explicit backwards-compat tests.
   - Dependency: command registry and services.
   - Verification: source sweep for `window.ivyApp`.

46. Remove `legacyStartup.js`, `legacyAppRuntime.js`, and `legacyScripts.js`.
   - Vue boot should directly install services and mount.
   - Dependency: no `startIvyApp` reliance.
   - Verification: App tests and browser suite.

47. Remove `legacyRuntimeGlobals.js` browser globals.
   - Keep `LegacyApiAdapter` only if it is still useful for engine compatibility.
   - Dependency: no global `IvyAPI`, `IvyControls`, `IvyPersist`, `IvyGraph` consumers.
   - Verification: source sweep and browser guard update.

48. Rename remaining compatibility modules.
   - `legacyGraph.js` should already be `graphRuntime.js`.
   - `legacyPersist.js` should already be `persistenceService.js`.
   - `legacyCommand.js` should be `commandRegistry.js`/`commandService.js`.
   - Dependency: previous phases.
   - Verification: source sweep for `legacy`.

### Phase 14: Final Hardening

49. Update browser guard tests.
   - Assert:
     - no `/static/js/ivyweb_*.js`
     - no `window.ivyApp` required for UI commands
     - Vue services can boot without legacy globals
   - Dependency: compatibility removal.
   - Verification: Playwright.

50. Split large tests by service.
   - The old `js_test/ivyweb_app_*.test.mjs` names should become service-specific tests.
   - Keep behavior coverage, change naming to match new ownership.
   - Dependency: extracted services.
   - Verification: test suite remains green.

51. Document new frontend architecture.
   - Add a `goivy/webui/frontend/README.md`.
   - Include:
     - build-only Vite workflow
     - Go-served production/dev workflow
     - service/store/component responsibilities
     - command naming conventions
     - how Wanix-compatible engines plug in
   - Dependency: stable target architecture.
   - Verification: documentation only.

52. Final source hygiene pass.
   - `rg "legacy|window.ivyApp|startIvyApp|IvyApp|IvyControls|IvyPersist|IvyGraph|__ivyVueBridge"`
   - Decide for every remaining hit:
     - keep as documented compatibility
     - rename
     - delete
   - Dependency: all phases.
   - Verification: documented exceptions only.

## Recommended Slice Size

Each implementation slice should be one service extraction or smaller. A good slice should:

- Change one ownership boundary.
- Preserve public behavior.
- Add or move tests before deleting old code.
- End with:
  - `npm run test:webui:js`
  - `npm run test:webui:vue`
  - `npm run build:webui`
- Run Playwright after UI, graph, save, tutorial, or sheet behavior changes.

## Suggested First Three Actual PR-Sized Slices

### Slice 1: Command Registry

Do steps 4-6.

Why first:

- It removes direct component dependence on `window.ivyApp`.
- It gives all later services a common command surface.
- It is low-risk because it can initially delegate to the controller.

Expected changed files:

- `frontend/src/services/commandRegistry.js`
- `frontend/src/components/legacyCommand.js`
- component tests that currently assert legacy command behavior

### Slice 2: Editor Service

Do steps 10-12.

Why second:

- Editor dirty/save feedback has already had regressions.
- It has good existing tests.
- It is a contained stateful integration.

Expected changed files:

- `frontend/src/services/editorService.js`
- `frontend/src/codeMirrorEditor.js`
- `frontend/src/stores/editorStore.js`
- editor/save tests

### Slice 3: File Service

Do steps 13-16.

Why third:

- File save/load is high user value.
- It depends on editor service.
- It removes a large cluster from the controller.

Expected changed files:

- `frontend/src/services/fileService.js`
- `frontend/src/components/FileInputHost.vue`
- `frontend/src/components/Menubar.vue`
- save/load tests

## Risks and Mitigations

Risk: Breaking browser file save semantics, especially Chrome vs Firefox.

Mitigation:

- Move save tests first.
- Keep File System Access API decisions in `fileService`.
- Add browser coverage for status/dirty marker timing if feasible.

Risk: Graph redraw regressions after layout changes.

Mitigation:

- Extract graph service only after command registry.
- Keep Playwright graph fit/redraw checks.
- Add service-level tests for resize/fit scheduling.

Risk: Event trace and analysis state have hidden cross-dependencies.

Mitigation:

- Defer event trace and analysis-state extraction until sheet/graph services exist.
- Keep old behavior tests green through each move.

Risk: Losing useful compatibility with tests or old debugging workflows.

Mitigation:

- Keep a temporary facade only as long as source sweeps show consumers.
- Every facade method should have a deletion milestone.

## Definition of Done

The Vue-idiomatic migration is complete when:

- No production component imports `legacyCommand.js`.
- No production code requires `window.ivyApp`.
- `legacyAppController.js` is deleted or reduced to a documented, test-only facade.
- No production module exports browser globals like `IvyApp`, `IvyPersist`, `IvyGraph`, or `IvyControls`.
- Service modules own side effects.
- Pinia stores own durable UI state.
- Browser tests pass with guards proving the app is Vue-bundled, Go-served, and free of old static runtime scripts.

