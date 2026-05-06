# Web Core Review Fix Plan

Date: 2026-05-06

This is a review pass over the currently implemented `webaudit_plan.md` work,
with emphasis on the completed event trace viewer, concept graph ordering, and
analysis-state save/load work. The review focus is correctness, Python
conformance, safety, and missing regression coverage.

The normal fix loop for every item below is:

1. Write the red test first.
2. Fix Go/JS to conform to Python behavior or to the documented web-only
   contract.
3. Keep `make test && make test-web` green.
4. Update this file's status before moving to the next item.

## 1. Event Pattern Syntax Errors Are Silently Treated As Empty Results

Status: confirmed bug

Priority: P0

Evidence:
- Python validates patterns before executing filter/find and raises `IvyError`
  on syntax error: `pyivy/ivy/ivy/ivy_ev_viewer.py:26`,
  `pyivy/ivy/ivy/ivy_ev_viewer.py:30`, `pyivy/ivy/ivy/ivy_ev_viewer.py:32`.
- Go `FilterEvents` discards parse errors and returns nil:
  `goivy/webui/webui_ui_evviewer.go:151`,
  `goivy/webui/webui_ui_evviewer.go:152`,
  `goivy/webui/webui_ui_evviewer.go:153`.
- Go `FindEventFrom` also maps parse errors to not-found:
  `goivy/webui/webui_ui_evviewer.go:197`,
  `goivy/webui/webui_ui_evviewer.go:198`,
  `goivy/webui/webui_ui_evviewer.go:199`.
- The web action layer therefore opens an empty filtered sheet or reports
  "Pattern not found": `goivy/webui/webui_session.go:902`,
  `goivy/webui/webui_session.go:903`, `goivy/webui/webui_session.go:917`,
  `goivy/webui/webui_session.go:918`, `goivy/webui/webui_session.go:920`.

Risk:
Invalid patterns look like valid patterns with zero matches. That hides user
errors and prevents golden/conformance tests from catching parser divergence.

Fix plan:
- Change the event filtering API shape so parsing errors are observable:
  `FilterEventsParsed` or `FilterEventsE` should return `([]*TraceEvent,
  error)`, and `FindEventFrom` should return `(*TraceEvent, string, error)`.
- Make `events_filter` and `events_find` return backend errors on syntax
  errors.
- Keep "not found" as a successful result only after a valid parse.
- Make JS event commands surface backend errors through the same status/error
  path used by other action runners.

Tests:
- Go unit test: malformed pattern passed to `FilterEvents`/new API returns a
  syntax error, not nil success.
- Go session test: `events_filter` and `events_find` return errors for malformed
  patterns.
- JS/Vitest test: rejected `events_find`/`events_filter` Promise reports an
  error status and does not open a new event sheet.

## 2. Event App Pattern Matching Is Too Permissive

Status: confirmed Python-conformance bug

Priority: P0

Evidence:
- Python dispatches matching from actual values to pattern values. An actual
  `App` can match pattern symbol `*`, or pattern app `*(...)`; an actual
  `Symbol` does not match an app pattern: `pyivy/ivy/ivy/ivy_ev_parser.py:66`,
  `pyivy/ivy/ivy/ivy_ev_parser.py:67`, `pyivy/ivy/ivy/ivy_ev_parser.py:89`,
  `pyivy/ivy/ivy/ivy_ev_parser.py:90`, `pyivy/ivy/ivy/ivy_ev_parser.py:91`.
- Go currently allows a bare actual symbol to match an app pattern whose rep is
  `*`: `goivy/webui/webui_ui_evviewer.go:363`,
  `goivy/webui/webui_ui_evviewer.go:364`,
  `goivy/webui/webui_ui_evviewer.go:365`,
  `goivy/webui/webui_ui_evviewer.go:366`,
  `goivy/webui/webui_ui_evviewer.go:367`.

Risk:
A pattern such as `call(*(b))` can match an actual `call(a)` even though Python
does not allow a symbol value to satisfy an app-shaped pattern. This can produce
false filter/find hits.

Fix plan:
- Remove the symbol fallback from the Go `eventApp` pattern case.
- Require the actual value to be an `eventApp` before comparing app rep and
  arguments.
- Keep Python's existing wildcard behavior where pattern symbol `*` matches any
  actual value.

Tests:
- Go unit test: `call(a)` does not match `call(*(b))`.
- Go unit test: `call(f(b))` does match `call(*(b))`.
- Go unit test: `call(a)` still matches `call(*)`.

## 3. Backend And Frontend Analysis-State Schemas Diverge

Status: confirmed bug

Priority: P0

Evidence:
- Server-side save emits `analysis_state_format: ivyweb-json` with snake-case
  file fields and a separate `event_viewer` object:
  `goivy/webui/webui_session.go:2041`,
  `goivy/webui/webui_session.go:2042`,
  `goivy/webui/webui_session.go:2046`,
  `goivy/webui/webui_session.go:2047`,
  `goivy/webui/webui_session.go:2050`.
- Frontend save/load emits and expects camel-case file fields, embeds event
  sheets in `sheets`, and never reads `event_viewer`:
  `goivy/webui/static/js/ivyweb_app.js:3053`,
  `goivy/webui/static/js/ivyweb_app.js:3065`,
  `goivy/webui/static/js/ivyweb_app.js:3087`,
  `goivy/webui/static/js/ivyweb_app.js:3090`,
  `goivy/webui/static/js/ivyweb_app.js:3091`,
  `goivy/webui/static/js/ivyweb_app.js:3092`,
  `goivy/webui/static/js/ivyweb_app.js:3138`,
  `goivy/webui/static/js/ivyweb_app.js:3142`,
  `goivy/webui/static/js/ivyweb_app.js:3143`,
  `goivy/webui/static/js/ivyweb_app.js:3144`.
- Both formats claim the same `analysis_state_format`, so they are not safely
  distinguishable.

Risk:
Files saved through `/api/session/{id}/save` cannot be faithfully reloaded by
the browser analysis-state loader, despite sharing the same format name. This is
especially dangerous because the names imply compatibility.

Fix plan:
- Define one canonical `ivyweb-json` schema for both Go and JS.
- Either make the browser save path use the server serializer, or make the
  server serializer match the browser schema exactly.
- Include event sheets in the same location for both directions.
- If a legacy shape must be accepted, version it explicitly and migrate it in
  one loader.

Tests:
- Go test: server `SaveState` JSON validates against the canonical schema.
- JS/Vitest test: `loadAnalysisStateObject(JSON.parse(serverSaveJSON))`
  restores file fields, analysis sheets, event sheets, toggles, and active tab.
- API/browser test: save through the backend route and load through the UI path.

## 4. Analysis-State Load Restores A Visual Snapshot, Not A Live Session

Status: confirmed correctness gap

Priority: P0

Evidence:
- JS captures only graph elements for analysis sheets:
  `goivy/webui/static/js/ivyweb_app.js:3075`,
  `goivy/webui/static/js/ivyweb_app.js:3080`,
  `goivy/webui/static/js/ivyweb_app.js:3081`.
- JS load recompiles the file content, then overlays saved Cytoscape elements:
  `goivy/webui/static/js/ivyweb_app.js:3155`,
  `goivy/webui/static/js/ivyweb_app.js:3156`,
  `goivy/webui/static/js/ivyweb_app.js:3173`,
  `goivy/webui/static/js/ivyweb_app.js:3174`,
  `goivy/webui/static/js/ivyweb_app.js:3182`,
  `goivy/webui/static/js/ivyweb_app.js:3183`,
  `goivy/webui/static/js/ivyweb_app.js:3185`,
  `goivy/webui/static/js/ivyweb_app.js:3186`.
- Backend state such as `SheetUIs`, ARG ownership, concept graph state, active
  facts, and CTI state is not reconstructed by this path.

Risk:
The browser can show restored old graphs while subsequent operations act on a
freshly reloaded backend session. A node click, Step In follow-up, concept
operation, or CTI action can target state that the backend does not own.

Fix plan:
- Decide the product contract:
  - If this is a visual snapshot, label it as such and disable backend actions
    on restored visual-only sheets.
  - If this is analysis-state load, move restore to the backend and rebuild
    live sheet/session state.
- The preferred fix is backend-owned load: send canonical state JSON to a server
  restore route that rebuilds sheet IDs, ARGs, concept graph state, toggles,
  active facts, event viewer state, and CTI state where available.
- Browser should render from the restored backend response instead of directly
  overlaying graph JSON.

Tests:
- Browser/Playwright test: save a multi-sheet session, load into a new session,
  click a restored non-root ARG node, and verify `/concept?sheet=...&node=...`
  succeeds against a backend-owned sheet.
- Go API test: load state restores `SheetUIs` for every analysis sheet.
- JS test: restored visual-only sheets, if retained, display explicit status and
  do not dispatch backend graph actions.

## 5. Analysis-State Load Accepts Unbounded And Weakly Validated JSON

Status: confirmed safety/hardening gap

Priority: P1

Evidence:
- The loader reads the entire file and parses it without size or shape checks:
  `goivy/webui/static/js/ivyweb_app.js:3132`,
  `goivy/webui/static/js/ivyweb_app.js:3134`,
  `goivy/webui/static/js/ivyweb_app.js:3135`.
- The restore loop renders every supplied sheet/event tree:
  `goivy/webui/static/js/ivyweb_app.js:3160`,
  `goivy/webui/static/js/ivyweb_app.js:3161`,
  `goivy/webui/static/js/ivyweb_app.js:3163`,
  `goivy/webui/static/js/ivyweb_app.js:3164`,
  `goivy/webui/static/js/ivyweb_app.js:3166`.

Risk:
An accidental or malicious `.ivyweb.json` can allocate very large DOM and graph
structures and freeze the browser.

Fix plan:
- Add explicit state validation before rendering:
  - maximum file size;
  - maximum sheets;
  - maximum graph elements per sheet;
  - maximum event count and event-tree depth;
  - valid sheet IDs;
  - valid event addresses.
- Reject invalid state with a clear error status and no partial restore.

Tests:
- JS/Vitest tests for oversized state, invalid IDs, excessive sheet count, and
  too-deep event trees.
- Browser test that a rejected state leaves the current session unchanged.

## 6. Event Trace DOM Selectors Are Built From Unescaped IDs And Addresses

Status: confirmed robustness bug

Priority: P1

Evidence:
- Event trace navigation builds raw selectors with `sheetId` and event address:
  `goivy/webui/static/js/ivyweb_app.js:1295`,
  `goivy/webui/static/js/ivyweb_app.js:1296`,
  `goivy/webui/static/js/ivyweb_app.js:1337`,
  `goivy/webui/static/js/ivyweb_app.js:1338`,
  `goivy/webui/static/js/ivyweb_app.js:1351`.
- Pattern removal does the same with the sheet ID:
  `goivy/webui/static/js/ivyweb_app.js:1433`,
  `goivy/webui/static/js/ivyweb_app.js:1435`.
- Loaded state can supply sheet IDs to `openEventTraceSheet`:
  `goivy/webui/static/js/ivyweb_app.js:1075`,
  `goivy/webui/static/js/ivyweb_app.js:1081`,
  `goivy/webui/static/js/ivyweb_app.js:1086`,
  `goivy/webui/static/js/ivyweb_app.js:1106`.

Risk:
Generated IDs are currently simple, but loaded analysis state can contain IDs
that throw selector syntax errors or select unintended elements.

Fix plan:
- Prefer `document.getElementById(sheetId)` followed by local queries.
- For attribute matching, either use `CSS.escape` or avoid selector strings and
  compare `getAttribute('data-event-address')`.
- Enforce valid sheet IDs in analysis-state validation.

Tests:
- JS/Vitest test with a loaded sheet ID containing selector-sensitive
  characters.
- JS/Vitest test with an event address containing a quote/bracket, or verify the
  validator rejects it before render.

## 7. Event Pattern List Mutates Browser State Before Backend Success

Status: confirmed correctness bug

Priority: P1

Evidence:
- Add mutates local state before backend acknowledgment:
  `goivy/webui/static/js/ivyweb_app.js:1422`,
  `goivy/webui/static/js/ivyweb_app.js:1425`,
  `goivy/webui/static/js/ivyweb_app.js:1426`,
  `goivy/webui/static/js/ivyweb_app.js:1428`,
  `goivy/webui/static/js/ivyweb_app.js:1429`.
- Remove/clear/load follow the same optimistic pattern:
  `goivy/webui/static/js/ivyweb_app.js:1437`,
  `goivy/webui/static/js/ivyweb_app.js:1438`,
  `goivy/webui/static/js/ivyweb_app.js:1448`,
  `goivy/webui/static/js/ivyweb_app.js:1458`,
  `goivy/webui/static/js/ivyweb_app.js:1459`.

Risk:
If the backend rejects a malformed pattern, unknown sheet, or bad index, the UI
and backend pattern lists diverge. After item 1 starts returning syntax errors,
this becomes immediately visible.

Fix plan:
- Update pattern list UI only after backend success.
- Treat backend-returned `patterns` as authoritative.
- If an operation fails, leave the local list unchanged and show the error.
- For save, use backend-returned content or validate that browser and backend
  content match.

Tests:
- JS/Vitest tests where `executeAction` rejects for add/remove/load and local
  state stays unchanged.
- JS/Vitest tests where backend returns canonical pattern formatting and the UI
  uses the returned list.

## 8. Existing Event Tabs Do Not Refresh Their Labels

Status: confirmed small UI bug

Priority: P2

Evidence:
- `openEventTraceSheet` creates a tab label only when the tab is new:
  `goivy/webui/static/js/ivyweb_app.js:1086`,
  `goivy/webui/static/js/ivyweb_app.js:1088`,
  `goivy/webui/static/js/ivyweb_app.js:1092`,
  `goivy/webui/static/js/ivyweb_app.js:1093`.
- Existing tabs reuse the old label while their sheet data is replaced:
  `goivy/webui/static/js/ivyweb_app.js:1103`,
  `goivy/webui/static/js/ivyweb_app.js:1111`,
  `goivy/webui/static/js/ivyweb_app.js:1118`.

Risk:
Reloading or replacing an event sheet can leave stale tab text, making the UI
misrepresent the active event data.

Fix plan:
- When `existingTab` exists, update its label span from `label || data.label`.

Tests:
- JS/Vitest test: reopening an existing event sheet with a new label updates
  the tab text.

## 9. `addSheet` Does Not Guard Against Duplicate Preferred IDs

Status: confirmed robustness bug

Priority: P1

Evidence:
- `addSheet` accepts `preferredSheetId` and always appends a new tab/content:
  `goivy/webui/static/js/ivyweb_app.js:1004`,
  `goivy/webui/static/js/ivyweb_app.js:1006`,
  `goivy/webui/static/js/ivyweb_app.js:1015`,
  `goivy/webui/static/js/ivyweb_app.js:1026`,
  `goivy/webui/static/js/ivyweb_app.js:1030`,
  `goivy/webui/static/js/ivyweb_app.js:1031`,
  `goivy/webui/static/js/ivyweb_app.js:1051`.
- `openARGSheet` passes preferred IDs through directly:
  `goivy/webui/static/js/ivyweb_app.js:1066`,
  `goivy/webui/static/js/ivyweb_app.js:1067`.

Risk:
Duplicate sheet IDs create invalid DOM and ambiguous graph ownership. The
current analysis-state load path removes non-root sheets first, but other call
sites can still create duplicates.

Fix plan:
- Make `addSheet` reject an existing `preferredSheetId` unless the caller
  explicitly requested replacement.
- Alternatively make it idempotent by updating and switching to the existing
  sheet.
- Use one behavior consistently for ARG sheets and event sheets.

Tests:
- JS/Vitest test: opening an ARG sheet with an existing preferred ID does not
  create duplicate DOM IDs.

## 10. Concept Graph Transitive Reduction Has Render-Time Side Effects

Status: confirmed correctness risk

Priority: P1

Evidence:
- Rendering calls transitive reduction during graph construction:
  `goivy/webui/webui_cyrender.go:177`,
  `goivy/webui/webui_cyrender.go:179`,
  `goivy/webui/webui_cyrender.go:180`.
- `GetTransitiveReduction` calls `checks.EnsureEdge(edge)`:
  `goivy/webui/webui_graph_model.go:793`,
  `goivy/webui/webui_graph_model.go:800`.
- `EnsureEdge` mutates the checkbox map when the edge is missing:
  `goivy/webui/webui_graph_model.go:106`,
  `goivy/webui/webui_graph_model.go:107`,
  `goivy/webui/webui_graph_model.go:111`,
  `goivy/webui/webui_graph_model.go:116`.

Risk:
Rendering a concept graph can create checkbox entries. That makes display state
change as a side effect of render, which can affect save/load, checkbox panels,
and later graph operations.

Fix plan:
- Add a pure checkbox read path, for example `EdgeVisibleIfPresent`.
- Make `GetTransitiveReduction` use non-mutating reads.
- Keep explicit sync paths responsible for creating new checkbox entries.

Tests:
- Go unit test: rendering with an unseen edge does not change
  `DisplayCheckboxes.EdgeDisplayCheckboxes`.
- Existing transitive reduction tests stay green.

## 11. Concept Graph Layout Coverage Is Too Narrow

Status: under-tested area

Priority: P2

Evidence:
- Current transitive layout test covers combiner edges only:
  `goivy/webui/webui_cyrender_test.go:364`,
  `goivy/webui/webui_cyrender_test.go:370`,
  `goivy/webui/webui_cyrender_test.go:386`,
  `goivy/webui/webui_cyrender_test.go:400`,
  `goivy/webui/webui_cyrender_test.go:403`.

Risk:
Future changes can regress non-combiner edge rendering, disabled checkbox
behavior, cycles, or relation classes without test failure.

Fix plan:
- Add coverage for:
  - `Domain.Edges` relation rendering with transitive checkboxes;
  - disabled transitive checkbox means no reduction/order forcing;
  - reflexive edge hiding;
  - cycle handling/stable order;
  - non-`all_to_all` edges are not hidden by reduction;
  - render purity from item 10.

Tests:
- Go-only cyrender tests should be enough for the above.

## 12. Event Parser Has Dead Code And Needs Explicit Single-Event Validation

Status: under-tested correctness risk

Priority: P2

Evidence:
- `parseTraceEventText` parses a list and returns only the first event:
  `goivy/webui/webui_ui_evviewer.go:577`,
  `goivy/webui/webui_ui_evviewer.go:578`,
  `goivy/webui/webui_ui_evviewer.go:582`.
- `skipBalanced` appears unused:
  `goivy/webui/webui_ui_evviewer.go:757`.

Risk:
The actual event text path can silently ignore extra top-level events in a
`TraceEvent.Text` string. This should not normally occur if all traces are built
by the parser, but the code accepts arbitrary JSON event data from analysis
state and backend action results.

Fix plan:
- Add a dedicated `parseSingleTraceEventText` that rejects zero or more than one
  top-level event.
- Use it in matching paths that parse `TraceEvent.Text`.
- Delete `skipBalanced` if no longer needed after tests confirm the recursive
  parser replaced it.

Tests:
- Go unit test: matching an event text containing two top-level events returns a
  parse error/no match according to the chosen contract.
- Go unit test: valid single event still matches.

## Recommended Fix Order

1. Item 1: make event syntax errors observable.
2. Item 2: fix event app-pattern conformance.
3. Item 7: make event pattern UI backend-authoritative.
4. Item 6: remove raw selector hazards and validate loaded IDs.
5. Item 3: unify the analysis-state schema.
6. Item 4: resolve visual-only versus live analysis-state restore.
7. Item 5: add analysis-state bounds/shape validation.
8. Item 9: make sheet ID ownership robust.
9. Item 10: remove render-time checkbox mutation.
10. Item 11: broaden concept layout coverage.
11. Item 8: refresh event tab labels on replacement.
12. Item 12: tighten event parser cleanup and single-event validation.

