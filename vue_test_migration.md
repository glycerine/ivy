# Vue Test Migration Plan

## Purpose

The Vue 3 runtime switch-over is in place, and the production UI now has Vue components, Pinia stores, and extracted services. The remaining test smell is that a meaningful slice of behavior is still tested through the old `IvyApp` browser-script harness in `goivy/webui/js_test`.

This plan migrates that coverage into idiomatic Vue/Vitest tests under `goivy/webui/frontend/src`, while keeping a small explicit compatibility suite for the temporary legacy facade.

## Target Test Architecture

The desired test ownership is:

- Store tests verify durable UI state, derived labels, state transitions, validation, and serialization helpers.
- Service tests verify workflows and side effects with explicit fake dependencies.
- Component tests verify rendered DOM, emitted user intent, Pinia integration, and command dispatch.
- Engine/API tests verify the hosted-Go and future engine boundary.
- Playwright tests verify the Go-served browser path, graph geometry, file/tutor/editor visual integration, and a few full-stack flows.
- Compatibility tests verify only intentional compatibility APIs while `legacyAppController.js` and browser globals still exist.

The old `js_test` harness should become smaller over time. It should not be the place where new Vue behavior is added.

## Current Test Suites

### Legacy JS Harness

Command:

```sh
npm run test:webui:js
```

Scope:

- Config: `goivy/webui/vitest.config.mjs`
- Tests: `goivy/webui/js_test/**/*.test.mjs`
- Loader: `goivy/webui/js_test/helpers/load_browser_scripts.mjs`
- Main object under test: `IvyApp` from `frontend/src/legacyAppController.js`

This harness is valuable for preserving behavior while logic is still exposed through `IvyApp`, but it is not idiomatic Vue because tests generally instantiate the compatibility controller and inspect controller fields or DOM fallbacks.

### Vue Frontend Suite

Command:

```sh
npm run test:webui:vue
```

Scope:

- Config: `goivy/webui/frontend/vitest.config.mjs`
- Tests: `goivy/webui/frontend/src/**/*.test.js`
- Main units under test:
  - `stores/*.js`
  - `services/*.js`
  - `components/**/*.vue`
  - `engines/*.js`
  - small compatibility modules during the migration

This is the target home for migrated tests.

### Browser Suite

Command:

```sh
npm run test:webui:browser
```

Scope:

- Config: `goivy/webui/playwright.config.mjs`
- Tests: `goivy/webui/pw_test/*.spec.mjs`
- Main purpose: prove the built Vue bundle works when served by Go.

Playwright tests should not replace unit coverage. They should keep a smaller number of high-value integration checks.

## Migration Rules

1. Migrate behavior, not filenames.
   - A legacy file can split into store, service, component, and Playwright coverage.

2. Prefer service tests for workflow behavior.
   - If a test currently says "IvyApp does X after backend result Y", the idiomatic target is usually a service test with fake `api`, fake stores, and fake graph/editor dependencies.

3. Prefer store tests for state-only rules.
   - Dirty/saving/saved labels, relation row state, dialog validation, layout constraints, event tree expansion, and sheet bookkeeping belong in Pinia store tests.

4. Prefer component tests for DOM and user intent.
   - Buttons, checkboxes, menus, dialogs, tabs, tutorial controls, headers, and editor-pane visual states belong in `*.vue` component tests using `@vue/test-utils` and Pinia.

5. Keep compatibility tests small and named as compatibility.
   - Tests that still exercise `window.ivyApp`, `IvyControls`, `IvyPersist`, `IvyGraph`, or `legacyAppController.js` must say why the compatibility surface still exists.

6. Do not delete legacy tests until their replacement tests fail for the same regression.
   - For each moved behavior, first add the Vue/service/store/component test.
   - Run both suites.
   - Then remove or narrow the old test.

7. Browser tests are the final safety net for layout and graph rendering.
   - Any migrated coverage involving Cytoscape fit/resize, tutorial hide/show, details splitter dragging, file-save browser APIs, or served bundle ownership should keep at least one Playwright smoke/regression test.

8. New tests should not reach directly into `window.ivyApp`.
   - Exceptions belong only in compatibility tests or Playwright diagnostics while the facade is still intentional.

## Shared Test Utilities To Add First

Before migrating individual test files, add a small set of Vue test helpers so the new tests do not grow their own mini-harnesses.

Target directory:

```text
goivy/webui/frontend/src/test/
```

Suggested helpers:

- `pinia.js`
  - Creates a fresh testing Pinia.
  - Installs all relevant stores with clean state.
  - Provides `resetStores()` or `withStores()` helper.

- `fakes.js`
  - Rehomes reusable fakes from `goivy/webui/js_test/helpers/fakes.mjs`.
  - Provides fake API, fake graph runtime, fake editor, fake file handles, fake engine, fake persistence, and fake dialogs.

- `mount.js`
  - Thin wrapper around `@vue/test-utils` `mount()`.
  - Installs Pinia and any global stubs consistently.

- `flush.js`
  - Exposes `flushPromises()` and `nextTick()` helpers for service/component tests.

- `dom.js`
  - Provides only the temporary DOM fixtures still required by compatibility fallback tests.
  - This file should shrink as DOM fallbacks are removed.

Order rule: create these utilities before moving any large legacy test file.

## Migration Order

### Phase 0: Lock The Boundary

1. Add or update a source guard test that Vue components do not import the legacy app controller or directly use `window.ivyApp`.
   - Current related guard: Playwright source scan for `window.ivyApp`.
   - Better long-term location: a fast Vue/Vitest source guard under `frontend/src`.

2. Add a test inventory comment or table to `legacyAppController.surface.test.js`.
   - The surface test should distinguish "temporary public facade" from "production command path".
   - This prevents old controller methods from quietly becoming the test target again.

3. Create the shared `frontend/src/test` helpers listed above.
   - Pull reusable fakes out of `goivy/webui/js_test/helpers/fakes.mjs`.
   - Do not remove the old helper until all old harness tests that need it are gone.

Verification:

```sh
npm run test:webui:all
```

### Phase 1: Low-Risk State And Component Tests

4. Migrate toast behavior.
   - Legacy source: `js_test/ivyweb_app_toast.test.mjs`
   - Targets:
     - `frontend/src/stores/toastStore.test.js`
     - add `frontend/src/components/ToastHost.test.js` if DOM rendering is not already covered
   - Compatibility residue:
     - one small test may remain for controller fallback `showToast()` only while the facade exists.

5. Migrate dialog primitive behavior.
   - Legacy source: `js_test/ivyweb_app_dialogs.test.mjs`
   - Targets:
     - `frontend/src/stores/dialogStore.test.js`
     - `frontend/src/services/dialogService.test.js`
     - `frontend/src/components/DialogHost.test.js`
   - Delete controller-oriented dialog tests once service/component coverage proves the same OK/cancel/text/integer/listbox/button-list behavior.

6. Migrate layout resizer behavior.
   - Legacy source: `js_test/ivyweb_app_layout.test.mjs`
   - Targets:
     - `frontend/src/stores/layoutStore.test.js`
     - `frontend/src/components/WorkspaceShell.test.js`
     - `frontend/src/components/panes/SheetArea.test.js`
   - Keep Playwright coverage for real browser drag behavior and editor width staying attached to the right edge.

7. Migrate tutorial hide/show redraw tests.
   - Legacy source: `js_test/ivyweb_app_toggles.test.mjs`, tutorial/editor layout cases
   - Targets:
     - `frontend/src/stores/layoutStore.test.js`
     - `frontend/src/components/panes/TutorialPane.test.js`
     - `frontend/src/services/graphService.test.js`
     - `frontend/src/services/editorService.test.js`
   - Keep Playwright coverage for concept graph fit after tutorial hide/show.

Verification after this phase:

```sh
npm run test:webui:vue
npm run test:webui:js
```

### Phase 2: Editor And File Workflows

8. Migrate editor dirty label behavior.
   - Legacy source: `js_test/ivyweb_app_editor_label.test.mjs`
   - Targets:
     - `frontend/src/stores/editorStore.test.js`
     - `frontend/src/services/editorService.test.js`
     - `frontend/src/components/panes/EditorPane.test.js`
   - Required coverage:
     - `**` appears when content becomes dirty.
     - `[saved]` is removed after edit.
     - `[saving...]` suppresses `**`.
     - `[saved]` appears after successful save.
     - save sheen appears only while saving.

9. Migrate save and save-as workflow tests.
   - Legacy source: `js_test/ivyweb_app_save.test.mjs`
   - Targets:
     - `frontend/src/services/fileService.test.js`
     - `frontend/src/services/recentFileService.test.js`
     - `frontend/src/stores/editorStore.test.js`
     - `frontend/src/components/SessionOverlayHost.test.js`
   - Required coverage:
     - existing writable handle save
     - immediate `[saving...]` state
     - Firefox/no-writable-handle fallback behavior
     - Chrome-capable Save As path unchanged
     - cancelled picker keeps dirty state
     - failed write does not mark saved
     - download fallback marks saved only when intended
     - recent-file metadata round-trips

10. Migrate file input host behavior.
    - Legacy sources:
      - file load portions of `js_test/ivyweb_app_save.test.mjs`
      - any controller-based file input assumptions
    - Targets:
      - `frontend/src/components/FileInputHost.test.js`
      - `frontend/src/services/fileService.test.js`

11. Migrate source browsing.
    - Legacy source: `js_test/ivyweb_app_source.test.mjs`
    - Targets:
      - `frontend/src/services/argActionService.test.js`
      - `frontend/src/services/editorService.test.js`
    - Required coverage:
      - ARG edge source action writes source into editor.
      - backend line number scroll/highlight is delegated to editor service.

Verification after this phase:

```sh
npm run test:webui:all
npm run build:webui
```

Run Playwright when changing browser save behavior or editor visual state.

### Phase 3: Session, Engine, And Command Wiring

12. Migrate API wiring tests.
    - Legacy source: `js_test/ivyweb_app_api.test.mjs`
    - Targets:
      - `frontend/src/services/sessionService.test.js`
      - `frontend/src/stores/engineStore.test.js`
      - `frontend/src/engines/hostedGoEngine.test.js`
      - `frontend/src/engines/legacyApiAdapter.test.js`
      - `frontend/src/services/editorService.test.js` for keymap bridge behavior
    - Compatibility residue:
      - a tiny fallback test for old `IvyAPI` construction may remain until `legacyRuntimeGlobals.js` is removed.

13. Migrate global keyboard/command behavior.
    - Legacy-adjacent source: `frontend/src/globalInteractions.test.js`
    - Targets:
      - keep this in Vue suite, but assert command names such as `file.save`, not controller method calls.
      - `frontend/src/services/commandRegistry.test.js`
      - `frontend/src/services/uiCommandService.test.js` if added.

14. Migrate self-binding guard.
    - Legacy source: `js_test/ivyweb_self_binding.test.mjs`
    - Target:
      - delete when `legacyAppController.js` is reduced to a facade with no callback-heavy method bodies.
      - until then, keep it as explicit compatibility debt.

Verification after this phase:

```sh
npm run test:webui:all
```

### Phase 4: State Relations, Visibility, And Details

15. Migrate state relation checkbox/toggle behavior.
    - Legacy sources:
      - `js_test/ivyweb_app_toggles.test.mjs`
      - `js_test/ivyweb_persist_state_relations.test.mjs`
      - `js_test/ivyweb_controls_toggles.test.mjs`
    - Targets:
      - `frontend/src/stores/stateRelationsStore.test.js`
      - `frontend/src/services/conceptVisibilityService.test.js`
      - `frontend/src/components/panes/SheetArea.test.js` or a dedicated state-relations component test if split out
      - `frontend/src/legacyPersist.test.js` only for remaining persistence shim behavior
    - Required coverage:
      - backend toggle hydration adds relation rows.
      - `X`, `Y`, `Z` symbolic rows remain visible for CTI state.
      - Vue owns rows without duplicate legacy DOM rows.
      - edge and label visibility update locally before backend action.
      - persistence snapshots read/write through stores, not checkbox scraping.

16. Migrate constraint facts.
    - Legacy source: `js_test/ivyweb_app_facts.test.mjs`
    - Targets:
      - `frontend/src/stores/detailsStore.test.js`
      - `frontend/src/services/detailsService.test.js`
      - component test for the details/facts UI if the UI is split into its own Vue component
    - Required coverage:
      - facts render from backend payload.
      - fact toggle callback dispatches action.
      - selection state lives in the store.

17. Migrate menu flashing and descriptor routing.
    - Legacy source:
      - menu case inside `js_test/ivyweb_app_toggles.test.mjs`
    - Targets:
      - `frontend/src/stores/dropdownStore.test.js`
      - `frontend/src/services/menuService.test.js`
      - `frontend/src/components/Menubar.test.js` if added

Verification after this phase:

```sh
npm run test:webui:all
npm run test:webui:browser
```

### Phase 5: Graph, Sheet, And Concept Actions

18. Migrate concept graph export.
    - Legacy source: `js_test/ivyweb_app_export.test.mjs`
    - Targets:
      - `frontend/src/services/conceptActionService.test.js` or `analysisActionService.test.js`, depending on current ownership of DOT export
      - `frontend/src/services/graphService.test.js` for active sheet graph lookup

19. Migrate concept materialization and projection.
    - Legacy source: `js_test/ivyweb_app_materialize.test.mjs`
    - Targets:
      - `frontend/src/services/conceptActionService.test.js`
      - `frontend/src/services/dialogService.test.js` for relation/name prompts
    - Required coverage:
      - materialize node/edge arguments are shaped correctly.
      - selected source node is used for materialize-edge-from-selected.
      - projection descriptors dispatch to `addProjection`.
      - add relation uses dialog service and backend action service.

20. Migrate ARG choice-backed commands.
    - Legacy source: `js_test/ivyweb_app_arg_choices.test.mjs`
    - Targets:
      - `frontend/src/services/argActionService.test.js`
      - `frontend/src/services/dialogService.test.js`
      - `frontend/src/services/analysisActionService.test.js` for remember graph
    - Required coverage:
      - conjecture prompt dispatches selected value.
      - remembered goal prompt dispatches selected name.
      - remember graph prompts for graph name.

21. Migrate sheet graph ownership tests.
    - Legacy source:
      - graph/sheet portions inside `js_test/ivyweb_app_analysis_state.test.mjs`
      - browser-only diagnostics in Playwright
    - Targets:
      - `frontend/src/stores/sheetStore.test.js`
      - `frontend/src/stores/graphStore.test.js`
      - `frontend/src/services/sheetService.test.js`
      - `frontend/src/services/graphService.test.js`
    - Keep Playwright coverage for:
      - root ARG/concept graph creation
      - sheet graph independence
      - concept graph geometry after tutorial hide/show

Verification after this phase:

```sh
npm run test:webui:all
npm run test:webui:browser
```

### Phase 6: CTI And Check Workflows

22. Migrate CTI workflow tests.
    - Legacy source: `js_test/ivyweb_app_cti.test.mjs`
    - Targets:
      - `frontend/src/services/checkService.test.js`
      - `frontend/src/services/conceptActionService.test.js`
      - `frontend/src/services/dialogService.test.js`
      - `frontend/src/stores/detailsStore.test.js`
    - Required coverage:
      - bounded check prompts for bound and sends backend option.
      - weaken invariant prompts for conjecture list.
      - CTI concept action includes active sheet ID.
      - trace button/actions are Vue-owned, not appended as legacy DOM.

23. Migrate check-result details and ARG trace behavior.
    - Legacy source:
      - CTI and details cases in `js_test/ivyweb_app_cti.test.mjs`
      - Playwright failed-check trace tests
    - Targets:
      - `frontend/src/services/checkService.test.js`
      - `frontend/src/services/sheetService.test.js`
      - `frontend/src/services/graphService.test.js`
    - Keep Playwright coverage for:
      - loading client/server example
      - check induction failure
      - textual details include counterexample
      - visual ARG trace sheet opens and renders

Verification after this phase:

```sh
npm run test:webui:all
npm run test:webui:browser
```

### Phase 7: Event Trace Workflows

24. Migrate event trace tree behavior.
    - Legacy source: `js_test/ivyweb_app_events.test.mjs`
    - Targets:
      - `frontend/src/stores/eventTraceStore.test.js`
      - `frontend/src/services/eventTraceService.test.js`
      - `frontend/src/components/panes/EventTraceSheet.test.js`
    - Required coverage:
      - event sheet opens with lazy child expansion.
      - selected row state lives in Pinia.
      - unsafe CSS-selector addresses are safe because selectors are not built from raw addresses.
      - invalid sheet IDs are rejected before DOM selectors or HTML.

25. Migrate event trace backend workflows.
    - Legacy source: `js_test/ivyweb_app_events.test.mjs`
    - Targets:
      - `frontend/src/services/eventTraceService.test.js`
      - `frontend/src/services/sheetService.test.js`
      - `frontend/src/stores/eventTraceStore.test.js`
    - Required coverage:
      - filter opens returned event sheet.
      - find starts from selected anchor and uncovers nested matches.
      - raw event trace file is parsed through backend.
      - backend filter errors surface without opening a sheet.

26. Migrate event pattern workflows.
    - Legacy source: `js_test/ivyweb_app_events.test.mjs`
    - Targets:
      - `frontend/src/stores/eventTraceStore.test.js`
      - `frontend/src/services/eventTraceService.test.js`
      - `frontend/src/components/panes/EventTraceSheet.test.js`
    - Required coverage:
      - add/remove/load failures leave pattern state unchanged.
      - successes use backend-authoritative pattern list.
      - saving patterns uses backend content.
      - replacing sheet data updates the existing tab label.
      - Vue removes event sheet DOM through sheet store.

Verification after this phase:

```sh
npm run test:webui:all
npm run test:webui:browser
```

### Phase 8: Analysis-State Serialization

27. Migrate analysis-state build/save behavior.
    - Legacy source: `js_test/ivyweb_app_analysis_state.test.mjs`
    - Targets:
      - `frontend/src/services/analysisStateService.test.js`
      - `frontend/src/stores/sheetStore.test.js`
      - `frontend/src/stores/graphStore.test.js`
      - `frontend/src/stores/eventTraceStore.test.js`
      - `frontend/src/stores/stateRelationsStore.test.js`
      - `frontend/src/stores/editorStore.test.js`
    - Required coverage:
      - root graphs, event sheets, mode, file content, and visibility state serialize from explicit store/service dependencies.
      - mode source of truth is Pinia/bridge, not DOM select mutation.

28. Migrate analysis-state restore behavior.
    - Legacy source: `js_test/ivyweb_app_analysis_state.test.mjs`
    - Targets:
      - `frontend/src/services/analysisStateService.test.js`
      - `frontend/src/services/sheetService.test.js`
      - `frontend/src/services/graphService.test.js`
      - `frontend/src/services/eventTraceService.test.js`
    - Required coverage:
      - saved visual analysis state restores graph instances and sheets.
      - active Vue-owned event sheet restores before DOM exists.
      - restored analysis sheets are visual-only and block backend graph actions.
      - duplicate preferred ARG sheet IDs are rejected without duplicating DOM.

29. Migrate analysis-state validation behavior.
    - Legacy source: `js_test/ivyweb_app_analysis_state.test.mjs`
    - Target:
      - `frontend/src/services/analysisStateService.test.js`
    - Required coverage:
      - oversized files rejected before reading.
      - invalid state rejected before mutating existing sheets.
      - excessive sheet counts rejected.
      - invalid event addresses rejected.

Verification after this phase:

```sh
npm run test:webui:all
npm run build:webui
```

Run Playwright if restore behavior touches rendered sheets or graph geometry.

### Phase 9: Persistence And Runtime Compatibility

30. Migrate persistence shim tests.
    - Legacy source: `js_test/ivyweb_persist_state_relations.test.mjs`
    - Existing related target: `frontend/src/legacyPersist.test.js`
    - Future target:
      - rename or replace with `frontend/src/services/persistenceService.test.js`
    - Required coverage:
      - session save/list remains compatible.
      - loaded-file display updates route through Vue.
      - mode restore routes through Vue.
      - non-Vue fallback behavior remains only while the compatibility shim exists.

31. Migrate controls shim tests.
    - Legacy source: `js_test/ivyweb_controls_toggles.test.mjs`
    - Target:
      - `frontend/src/legacyRuntimeGlobals.test.js` while compatibility exists
      - delete once `IvyControls` compatibility is removed

32. Migrate legacy graph runtime tests.
    - Existing source: `frontend/src/legacyGraph.test.js`
    - Future target:
      - `frontend/src/services/graphRuntime.test.js`
    - Required coverage:
      - Cytoscape default setup and position handling.
      - no test should imply production code needs a global `IvyGraph` once compatibility is gone.

33. Migrate startup/runtime/script tests.
    - Existing sources:
      - `frontend/src/legacyStartup.test.js`
      - `frontend/src/legacyAppRuntime.test.js`
      - `frontend/src/legacyScripts.test.js`
      - `frontend/src/legacyRuntimeGlobals.test.js`
    - Target:
      - keep them as compatibility tests only until Vue boot no longer installs legacy globals.
      - after facade removal, replace with `main/App` boot tests and a source guard.

Verification after this phase:

```sh
npm run test:webui:all
npm run test:webui:browser
```

### Phase 10: Retire The Legacy Harness

34. For each `js_test` file, add a header note before deletion or narrowing.
    - The note should list replacement test files.
    - This is useful during review and prevents accidental coverage loss.

35. Delete migrated legacy tests one file at a time.
    - After each deletion:

```sh
npm run test:webui:js
npm run test:webui:vue
```

36. Keep only a tiny compatibility directory if needed.
    - Preferred final shape:

```text
goivy/webui/js_test/compat/
```

    - Contents should be limited to browser-global/facade compatibility that truly cannot live in `frontend/src`.

37. Remove `goivy/webui/js_test/helpers/load_browser_scripts.mjs` when no test imports it.

38. Remove `goivy/webui/vitest.config.mjs` or narrow it to `js_test/compat/**/*.test.mjs`.
    - If no `js_test` files remain, update `package.json` so `test:webui:js` is removed or becomes an alias to `test:webui:vue`.
    - Update `goivy/Makefile` targets only after the package script is settled.

39. Update documentation.
    - `goivy/webui/frontend/README.md` should say new frontend tests live under `frontend/src`.
    - Mention Playwright as served-browser coverage.
    - Mention any remaining compatibility tests by exact path and deletion condition.

Final verification:

```sh
npm run test:webui:vue
npm run build:webui
npm run test:webui:browser
go test ./goivy/webui
```

## Legacy-To-Target Coverage Map

| Legacy source | Primary target | Secondary target | Keep browser coverage? |
| --- | --- | --- | --- |
| `ivyweb_app_toast.test.mjs` | `stores/toastStore.test.js` | `components/ToastHost.test.js` | No, unless UI placement regresses |
| `ivyweb_app_dialogs.test.mjs` | `stores/dialogStore.test.js`, `services/dialogService.test.js` | `components/DialogHost.test.js` | No |
| `ivyweb_app_layout.test.mjs` | `stores/layoutStore.test.js` | `WorkspaceShell.test.js`, `SheetArea.test.js` | Yes |
| `ivyweb_app_toggles.test.mjs` | `services/conceptVisibilityService.test.js` | `stateRelationsStore.test.js`, layout/editor/graph tests | Yes |
| `ivyweb_controls_toggles.test.mjs` | `legacyRuntimeGlobals.test.js` temporarily | delete later | No |
| `ivyweb_app_facts.test.mjs` | `detailsStore.test.js`, `detailsService.test.js` | details component test if split | No |
| `ivyweb_app_analysis_state.test.mjs` | `analysisStateService.test.js` | sheet/graph/event/editor/relation store tests | Yes for rendered restore/graphs |
| `ivyweb_persist_state_relations.test.mjs` | `persistenceService.test.js` | `stateRelationsStore.test.js`, `legacyPersist.test.js` temporarily | No |
| `ivyweb_self_binding.test.mjs` | delete after facade shrink | `legacyAppController.surface.test.js` until then | No |
| `ivyweb_app_source.test.mjs` | `argActionService.test.js` | `editorService.test.js` | Existing source-view browser smoke is enough |
| `ivyweb_app_api.test.mjs` | `sessionService.test.js`, `engineStore.test.js` | `hostedGoEngine.test.js`, `legacyApiAdapter.test.js` | Browser session smoke already exists |
| `ivyweb_app_materialize.test.mjs` | `conceptActionService.test.js` | `dialogService.test.js` | Context-menu browser smoke |
| `ivyweb_app_export.test.mjs` | `conceptActionService.test.js` or `analysisActionService.test.js` | `graphService.test.js` | No |
| `ivyweb_app_events.test.mjs` | `eventTraceService.test.js` | `eventTraceStore.test.js`, `EventTraceSheet.test.js` | Yes for rendered event sheets |
| `ivyweb_app_arg_choices.test.mjs` | `argActionService.test.js` | `dialogService.test.js` | No |
| `ivyweb_app_cti.test.mjs` | `checkService.test.js` | `detailsStore.test.js`, `conceptActionService.test.js` | Yes for failed-check trace |
| `ivyweb_app_save.test.mjs` | `fileService.test.js` | `editorStore.test.js`, `recentFileService.test.js`, `SessionOverlayHost.test.js` | Yes for browser file/status behavior |
| `ivyweb_app_editor_label.test.mjs` | `editorStore.test.js`, `editorService.test.js` | `EditorPane.test.js` | Browser smoke optional |

## Definition Of Done

The test migration is complete when:

- New Vue behavior is covered under `goivy/webui/frontend/src/**/*.test.js`.
- No non-compatibility test instantiates `IvyApp`.
- No non-compatibility test imports `goivy/webui/js_test/helpers/load_browser_scripts.mjs`.
- `goivy/webui/js_test` is deleted or reduced to a clearly named compatibility-only suite.
- `test:webui:vue` is the primary frontend unit suite.
- `test:webui:browser` remains the served Go/browser confidence suite.
- Compatibility tests document why each remaining global or facade API still exists.
- `legacyAppController.surface.test.js` shrinks whenever the facade shrinks.
- Full verification passes:

```sh
npm run test:webui:vue
npm run build:webui
npm run test:webui:browser
go test ./goivy/webui
```

## Suggested First Execution Slice

Start with Phase 0 and Phase 1:

1. Add `frontend/src/test` helpers.
2. Move toast tests.
3. Move dialog tests.
4. Move layout/tutorial redraw tests that already have Vue store/component/service homes.

This slice is intentionally low-risk: it establishes the new test style, exercises Pinia and component mounting, and removes easy `IvyApp` dependencies before touching save, graph, CTI, or analysis-state behavior.
