Plan only. No code changes, no git.

**Goal**
Make `UIDataModel` the browser-side source of truth: API/service results update the typed model first, and Cytoscape/DOM renderers consume only model-derived view state. The backend remains authoritative for computation, but the frontend stops rendering directly from raw service payloads.

**Current Gap**
`UIDataModel` already stores typed snapshots in [uiDataModel.ts](/Users/jaten/go/src/github.com/glycerine/ivy/goivy/webui/frontend/src/models/uiDataModel.ts:594), and `DATA_MODEL.md` says `ARGSnapshot` / `ConceptSnapshot` are service-layer entry points in [DATA_MODEL.md](/Users/jaten/go/src/github.com/glycerine/ivy/goivy/webui/DATA_MODEL.md:200). But services still call `graph.update(raw.elements, raw.positions)` immediately after accepting snapshots, for example [checkService.ts](/Users/jaten/go/src/github.com/glycerine/ivy/goivy/webui/frontend/src/services/checkService.ts:13) and [argActionService.ts](/Users/jaten/go/src/github.com/glycerine/ivy/goivy/webui/frontend/src/services/argActionService.ts:45). Concept controls also keep parallel mutable state in `_lastConceptData`, `_edgeVisibility`, and `_labelVisibility` in [conceptVisibilityService.ts](/Users/jaten/go/src/github.com/glycerine/ivy/goivy/webui/frontend/src/services/conceptVisibilityService.ts:61).

**Plan**
1. Add a small model driver layer around `UIDataModel`.

Create `frontend/src/models/uiDataModelStore.ts` with typed mutations like `applyArgSnapshot`, `applyConceptSnapshot`, `applyCtiSnapshot`, `setActiveSheet`, `setSelectedArgNode`, `setConceptSelection`, and `setVisualOnly`. Each mutation returns or publishes a change descriptor: `{sheetId, changed: ['arg', 'concept', 'selection', ...]}`.

2. Add pure selectors/view-model builders.

Create `frontend/src/models/uiDataSelectors.ts`:
- `selectArgGraphView(sheet)` from `sheet.arg.render`, `sheet.arg.analysisGraphState`, and `sheet.selectedArgNode`
- `selectConceptGraphView(sheet)` from `sheet.concept.render`, `displayCheckboxes`, `interactiveSession.abstractValue`, and selection
- `selectStateCheckboxRows(sheet)`
- `selectConstraintFacts(sheet)`
- `selectGraphStackControls(sheet)`
- `selectStateLabel(sheet)`

This is the key shift: Cytoscape and DOM code receive these selectors, never raw API payloads.

3. Add a deterministic render coordinator.

Create `frontend/src/services/uiDataRenderService.ts` that subscribes to model changes and fans out to:
- `syncArgGraph(graph, selectArgGraphView(...))`
- `syncConceptGraph(graph, selectConceptGraphView(...))`
- `renderStateCheckboxes(selectStateCheckboxRows(...))`
- `renderConstraintFacts(selectConstraintFacts(...))`
- `renderStateLabel(selectStateLabel(...))`

Initially this can still feed existing `IvyGraph.update`, but only with typed model-derived elements. Later it can become a diffing Cytoscape sync to avoid remove/readd churn.

4. Refactor service result paths.

Replace patterns like:

```ts
app.acceptArgSnapshot(sheetId, result.arg);
argGraph.update(result.arg.elements, result.arg.positions);
```

with:

```ts
app.modelDriver.applyArgSnapshot(sheetId, result.arg);
```

Do this for `checkService`, `argActionService`, `analysisActionService`, `analysisStateService`, `persistenceService`, `openARGSheet`, restore paths, and `refreshConceptGraph`.

5. Move concept visibility into the model.

`DisplayCheckboxes` is already typed in [uiDataModel.ts](/Users/jaten/go/src/github.com/glycerine/ivy/goivy/webui/frontend/src/models/uiDataModel.ts:316). Retire `_edgeVisibility`, `_labelVisibility`, and `_lastConceptData` as authoritative state. Checkbox changes should call the backend toggle endpoint, then apply the returned fresh `ConceptSnapshot`; the checkbox table and Cytoscape edge/node label visibility should be rendered from selectors.

6. Cover PLAN383 items 83 through 86 explicitly.

From [tcl_gui_behavior_inventory_codex.md](/Users/jaten/go/src/github.com/glycerine/ivy/goivy/tcl_gui_behavior_inventory_codex.md:513):
- 83: model typed graph element metadata for actions, callbacks, short info, long info, tooltip text, and context menu descriptors. Render context menus from model data.
- 84: add model-owned graph selection as typed records, for example `{kind:'node', obj}` and `{kind:'edge', obj, sourceObj, targetObj}`. Derive tuple-compatible fact gathering from that selector. Clear selection when graph element version changes.
- 85: implement relation/class bulk toggles as model/API actions, then derive checkbox rows and concept graph visibility from refreshed `ConceptSnapshot.displayCheckboxes`.
- 86: introduce typed modal/interaction request state so async dialogs are model-driven too. Service code enqueues interaction requests, DOM renders them, and completion resumes through a typed result.

7. Fix persistence to save the model, not the view.

`analysisStateService` and `persistenceService` currently serialize `cy.json()` and DOM checkbox state. Add a `serializeUIDataModel()` DTO and a legacy importer that still accepts old `{elements, positions}` saved states by normalizing them into `ARGSnapshot` / `ConceptSnapshot`.

8. Add guardrails.

Tests should assert:
- service handlers mutate `UIDataModel` but do not call `graph.update` directly
- selectors derive graph elements, checkbox rows, facts, labels, and selection deterministically
- toggles refresh through `ConceptSnapshot`, not local `_edgeVisibility`
- selection clears on graph replacement
- restore/load-state rehydrates model first, then renderers update

**Acceptance Criteria**
- No service path renders directly from raw API result payloads.
- Cytoscape updates only through model-derived graph view models.
- DOM panels for state checkboxes, facts, state labels, undo/redo, and modal requests only read model selectors.
- PLAN383 items 83-86 have typed model ownership and tests.
- `DATA_MODEL.md` can be updated to say the TypeScript model is now the frontend driver, not just a typed mirror.
