# SVK Plan 002: Finish Webui Parity

This plan exists because the current SvelteKit app has the right broad shape, but it is still a scaffold. It must stop showing canned graph/details/status data, stop defaulting to fake engine behavior, and regain the full old `goivy/webui` workbench surface: menus, resizers, status updates, editor behavior, state toggles, event traces, tutorial pane, dialogs, file flows, and live backend interaction.

The goal is not to reproduce the Vue implementation. The goal is to port the product behavior into maintainable SvelteKit code over the neutral engine interface, so the same UI can run against:

- hosted `goivy/webui` compatibility server;
- hosted `goivy/webengine` native Go engine;
- browser-local `goivy/webengine` js/wasm plus browser Z3 wasm;
- test-only fake engines.

Fake data is allowed in tests, fixtures, story/demo routes, and deliberately marked harnesses. It must not be the default product path.

## Current Audit Summary

The old webui has these product surfaces:

- `WorkspaceShell.vue`: menubar/status/workbench shell, resizable editor, resizable tutorial pane.
- `Menubar.vue`: File menu, mode selector, Check, Show Reachable, Undo, Reset Domain, Diagram Domain, tutorial toggle, loaded-file display.
- `StatusBar.vue`: live status message, severity coloring, session display.
- `SheetArea.vue`: tab bar, analysis/event sheets, ARG/concept/details/state layout, graph/state/details resize dividers.
- `ArgPane.vue`: ARG title, Invariant dropdown, dynamic backend menu region, graph container.
- `ConceptPane.vue`: Concept graph title, Conjecture/View/Action/View menus, dynamic backend menu region, graph container.
- `StateRelationsPane.vue`: live state label, relation toggle table, empty placeholder.
- `DetailsPane.vue`: selectable node/edge details, constraint fact toggles, error trace action.
- `EditorPane.vue`: close/reopen file controls, editor label, dirty/saved state, Sublime/Emacs/Vim keymaps, CodeMirror docs link.
- `EventTraceSheet.vue`: event tree, filter/find, selected event, pattern list, pattern navigation, save/load/clear.
- `TutorialPane.vue`: offline tutorial iframe, URL bar, back/forward/reload/close, flashing tutorial button.
- `DialogHost.vue`, `ContextMenuHost.vue`, `ToastHost.vue`, `SessionOverlayHost.vue`, `FileInputHost.vue`.
- Command/service layer: `appCommands`, `commandRegistry`, `fileService`, `sessionService`, `sheetService`, `graphRuntime`, `checkService`, `analysisActionService`, `argActionService`, `conceptActionService`, `eventTraceService`, `dialogService`, `analysisStateService`, `persistenceService`, `recentFileService`, `menuService`.

The current `goivy/svk` app has:

- engine abstractions and partial hosted/wasm/fake implementations;
- normalized state for models, graphs, concepts, checks, jobs, state relations;
- a monolithic `src/routes/+page.svelte` with hard-coded layout and canned concept/state/details/status content;
- `FakeEngine` visible as a primary product engine;
- no full workbench component hierarchy;
- incomplete menus and no full command registry;
- missing event trace UI;
- missing tutorial UI;
- missing dialog/context/toast/file-input overlays;
- missing recent files and analysis state UI;
- missing real CodeMirror keymap behavior;
- missing resizer persistence and robust graph layout refresh;
- incomplete backend command/action coverage.

## Non-Negotiable Definition of Done

Plan 002 is complete only when all of these are true:

- The default app can start in a real engine mode: hosted webui/native webengine when online, browser wasm when explicitly local/offline, and fake only behind tests/dev demo affordances.
- No production pane renders canned graph, concept, state relation, details, job, or check data when an engine has not provided it.
- The app visually matches the old `webui` first screen closely enough that a user recognizes it as the same Ivy workbench.
- All old top-level menus and pane menus exist with the same labels, enabled states, command routing, and visible feedback.
- The status bar updates from real session/job/check/action state, with severity color and current session display.
- The editor is present, dark, line-numbered, dirty-aware, save-aware, and keymap-aware.
- ARG and concept graphs are live Cytoscape-backed views, not placeholder DOM drawings.
- State relation toggles are populated from engine snapshots and mutate real engine/session state through the abstract engine API.
- Details pane supports selected graph items, constraint facts, and error trace actions.
- Event trace sheets work.
- Tutorial pane works offline using bundled tutorial assets.
- Resizers/sliders work and persist: graph split, state panel width, details height, editor width, tutorial height.
- Login/marketing routes still work and do not leak into the workspace UI.
- Playwright covers real user flows, not only smoke pages.
- Unit tests cover each store/service/component as it is introduced.
- A static "unfinished product" gate fails on product-source TODO/stub/canned/fake markers outside explicit test/demo/fixture allowlists.

## Architecture Direction

Keep the engine boundary neutral. Do not bake `webui` REST quirks into components.

Components talk to Svelte stores and command services. Command services talk to `IvyEngine`. Engine adapters translate to hosted webui, hosted webengine, browser wasm, or fake. The old Vue services are the behavior reference, not the destination architecture.

Target layers:

- `src/lib/workbench/components`: Svelte UI components.
- `src/lib/workbench/state`: layout, menus, dialogs, sheets, editor, event traces, details, recents, session status.
- `src/lib/workbench/commands`: command registry and command implementations.
- `src/lib/workbench/services`: file, analysis state, graph actions, concept actions, ARG actions, checks, events, tutorial.
- `src/lib/engines`: neutral engine adapters and protocol normalization.
- `src/routes/+page.svelte`: thin composition root only.

Do not let components call arbitrary engine methods directly except through narrow injected command/service objects. This keeps local-first wasm and remote server execution interchangeable.

## Phase 0: Parity Inventory And Failing Gates

Purpose: make the missing work visible and prevent new TODO drift.

Implementation:

- Add a machine-readable parity inventory in `src/lib/workbench/parity/webuiParityInventory.ts`.
- Inventory every old component, menu item, command, dialog type, store, and service from `goivy/webui/frontend/src`.
- Add each command label and command id from:
  - File menu;
  - ARG Invariant menu;
  - Concept Conjecture/View/Action/View menus;
  - dynamic menu regions;
  - event trace controls;
  - editor controls;
  - context menus.
- Add an allowlist for features intentionally postponed, initially empty except for explicitly test/demo-only features.
- Add a `npm run parity:audit` script that fails if any required inventory item has no Svelte implementation marker.
- Add a `npm run unfinished:audit` script that fails on `TODO`, `stub`, `placeholder`, `canned`, `Fake engine`, and hard-coded fallback status/check text in production source. Allow tests, generated wasm glue, fixtures, and explicit `FakeEngine` files.

Tests:

- Unit test inventory completeness against parsed old command files where practical.
- Unit test unfinished audit allowlist behavior.
- Run `npm run check`, `npm run lint`, `npm run test:unit -- --run`, `npm run parity:audit`, `npm run unfinished:audit`.

Exit criteria:

- The audit fails before implementation starts, with a clear list of missing work.
- The failure list becomes the checklist for the remaining phases.

## Phase 1: Workbench Component Skeleton

Purpose: replace the monolithic `+page.svelte` with the old workbench structure, without yet wiring every command.

Implementation:

- Create:
  - `WorkbenchShell.svelte`
  - `Menubar.svelte`
  - `StatusBar.svelte`
  - `SheetArea.svelte`
  - `TabBar.svelte`
  - `ArgPane.svelte`
  - `ConceptPane.svelte`
  - `StateRelationsPane.svelte`
  - `DetailsPane.svelte`
  - `EditorPane.svelte`
  - `TutorialPane.svelte`
  - `EventTraceSheet.svelte`
  - `EventTraceNode.svelte`
  - `DialogHost.svelte`
  - `ContextMenuHost.svelte`
  - `ToastHost.svelte`
  - `SessionOverlayHost.svelte`
  - `FileInputHost.svelte`
- Move `+page.svelte` to a thin composition root that creates stores/services and mounts `WorkbenchShell`.
- Port the visual layout from `goivy/webui/static/css/ivy.css` into Svelte component CSS or shared workbench CSS.
- Preserve old ids/data hooks where useful for tests and contract parity: `arg-graph`, `concept-graph`, `state-checkbox-table`, `model-editor`, `statusbar`, `tutorial-iframe`.
- Remove hard-coded concept nodes and fallback failed check text from production UI.
- Show empty/loading states only when driven by real app state.

Tests:

- Component tests verify each pane renders the expected structural labels and accepts empty state.
- Playwright screenshot smoke for desktop layout at 2048x1180 and 1366x768.
- Playwright mobile/narrow smoke verifies no text overlap and usable horizontal/vertical overflow behavior.

Exit criteria:

- The Svelte first screen visually matches the old workbench frame: dark menubar, tabs, ARG, concept graph, state/relation pane, editor, details pane, status bar.
- No fake graph/details/status fallback appears when engine data is absent.

## Phase 2: Stores For Missing UI State

Purpose: port the state model that made the old workbench reactive.

Implementation:

- Add Svelte stores/state modules for:
  - session status and mode;
  - editor state and keymap;
  - layout dimensions and tutorial navigation;
  - dropdown/menu state;
  - dynamic menu descriptors;
  - dialogs;
  - context menu;
  - toasts;
  - sheets and visual-only sheets;
  - event traces;
  - details and constraint facts;
  - recent files.
- Keep existing `models`, `graphs`, `jobs`, `checks`, `stateRelations`, `workspace`, and `engines` state, but adapt them into the workbench-facing selectors.
- Add derived selectors for:
  - active sheet;
  - active graph pair;
  - selected ARG/concept node;
  - latest check summary;
  - status bar severity;
  - loaded file display/title;
  - dirty/saved editor label.

Tests:

- One unit test file per state module.
- Port the intent of old Pinia store tests into Svelte tests.
- Tests must cover reset/new-session behavior.

Exit criteria:

- Every component from Phase 1 renders from state, not local hard-coded values.

## Phase 3: Command Registry And Menu System

Purpose: restore webui's command layer in Svelte so menus are complete and testable.

Implementation:

- Implement `commandRegistry.ts` with register/unregister/run/has/list.
- Implement `uiCommandService.ts` for menu click routing, dropdown close, menu flash, keyboard shortcuts.
- Add command groups mirroring old surfaces:
  - file commands;
  - session commands;
  - editor commands;
  - sheet commands;
  - graph commands;
  - ARG commands;
  - concept commands;
  - analysis action commands;
  - check commands;
  - analysis state commands;
  - event trace commands;
  - dialog commands;
  - menu commands.
- Implement `DynamicMenuRegion.svelte` equivalent for backend-provided menu descriptors.
- Port all old menu labels exactly unless product copy is deliberately improved:
  - File: Load, Open Event Trace, Save as, Download current model, Save Analysis State, Load Analysis State, Save Invariant, Recent files, New Model.
  - Menubar actions: Check, Show Reachable, Undo, Reset Domain, Diagram Domain.
  - ARG Invariant menu: Check induction, Bounded check, Diagram, Weaken, Save Invariant, Save Abstraction.
  - Concept Conjecture menu: Undo, Redo, PDR step, Concrete, Gather, CTI Gather, Minimize, Check sufficient, Check relative induction, Strengthen, Reverse, Path reach, Reach, Conjecture, Backtrack, Recalculate, Diagram, Remember, Export.
  - Concept View menu: Add relation.
  - Event trace controls: Filter, Find fwd/rev, pattern fwd/rev, add/remove/save/load/clear.
- Add global keyboard shortcuts from old `globalInteractions.js`: Escape closes menus/dialogs, Cmd/Ctrl-S saves, Cmd/Ctrl-Z invokes undo.

Tests:

- Unit tests for command registry.
- Component tests for every menu label and enabled/disabled behavior.
- Playwright opens each dropdown and verifies label order.
- A parity audit test maps every inventory menu item to a command id.

Exit criteria:

- Menus are complete even before every backend action is implemented.
- Disabled commands show disabled state with an explanatory status/toast, not silent no-ops.

## Phase 4: Real Session, Status, Jobs, And Checks

Purpose: make status bar and details strip live again.

Implementation:

- Introduce a session service that starts/closes sessions through `IvyEngine`.
- Subscribe to engine events once per active session.
- Reduce engine events into jobs, graph snapshots, concepts, checks, toggles, state relation rows, details, and status messages.
- Replace `statusMessage` local state with `sessionStatus` store.
- Implement status severity: normal, info, success, warning, error.
- `Check` must:
  - save/load current editor text into active engine session;
  - run the requested mode;
  - show progress in details/job strip;
  - update status bar from result;
  - populate failed conjecture, Z3 contact, counterexample trace when returned.
- `Show Reachable`, `Undo`, `Reset Domain`, `Diagram Domain`, bounded check, induction check, and PDR actions route through command services.

Engine requirements:

- Extend `IvyEngine` if necessary with neutral operations for:
  - `startSession`;
  - `loadModel`;
  - `runCommand`;
  - `subscribe`;
  - `snapshot`;
  - `cancelJob` if supported;
  - `getMenuDescriptors`;
  - `getStateRelations`;
  - `setRelationToggle`;
  - `getDetails`;
  - `getEventTrace`.

Tests:

- Unit test event reducers.
- Unit test status transitions for success/warning/error.
- Hosted webui contract test submits the real client/server example and verifies a Z3-backed check result reaches the UI state.
- Browser wasm smoke test loads wasm, starts a session, runs a check where supported, and verifies no fake data.
- Playwright: edit model, click Check, observe status bar and details update.

Exit criteria:

- The red/green/yellow status bar reflects real engine results.
- No fallback "Check FAILED" text exists in product UI.

## Phase 5: Editor And File Lifecycle

Purpose: restore the editor as a real work surface, not a textarea sketch.

Implementation:

- Decide and implement editor core:
  - Prefer CodeMirror 6 for modern Svelte integration.
  - Preserve user-facing keymap choices: Sublime, Emacs-ish, Vim.
  - If CodeMirror 5 compatibility is substantially easier for existing Vim/Sublime behavior, document the trade and lock it down with tests.
- Implement editor label behavior:
  - `Editing: filename [saved]`;
  - `** filename` when dirty;
  - `[saving...]` while saving;
  - unsaved-file label when no path exists.
- Implement file commands:
  - load `.ivy`;
  - save;
  - save as;
  - download current model;
  - close current file with dirty prompt;
  - reopen last file/session;
  - new model/new session.
- Implement browser File System Access API path with download fallback.
- Implement external disk change detection:
  - overwrite;
  - reload;
  - merge conflict markers;
  - cancel.
- Implement recent files/session list.
- Persist editor content and metadata in local-first storage.

Tests:

- Unit tests for dirty/saved labels.
- Unit tests for external change merge.
- Unit tests for File System Access fallback decisions.
- Component tests for keymap radio updates.
- Playwright: load fixture, edit, dirty indicator changes, save/download path fires, close dirty prompt appears.

Exit criteria:

- The editor behaves as a real Ivy source editor and can drive hosted and browser-local checks.

## Phase 6: Graph Runtime And Layout Parity

Purpose: make ARG and concept graphs live, interactive, and styled like webui.

Implementation:

- Replace placeholder graph DOM with Cytoscape-backed Svelte graph components.
- Port graph styles from old `graphRuntime.js` and CSS:
  - ARG circular/gray states;
  - concept octagonal nodes;
  - blue/red borders;
  - gray arrows;
  - selected node classes;
  - labels and edge labels.
- Implement layout refresh hooks after:
  - graph update;
  - pane resize;
  - tab switch;
  - tutorial/editor visibility changes.
- Implement graph selection:
  - ARG node click updates selected ARG node, state label, details, concept graph for selected state.
  - concept node click/select behavior matches old UI.
- Implement context menus:
  - ARG node actions;
  - ARG edge actions;
  - concept node actions;
  - concept edge actions.
- Implement graph snapshots per sheet and visual-only restored sheets.
- Implement `openARGSheet` for decomposition/step sheets.

Tests:

- Unit tests for Cytoscape element normalization.
- Component tests for selection and context menu emission.
- Playwright canvas/DOM check verifies graph is non-empty after loading a real model.
- Playwright right-clicks a node/edge and verifies context menu labels.
- Visual screenshot compares old webui and svk layout on client/server example.

Exit criteria:

- The concept graph from the screenshot is rendered by real graph data, not hard-coded HTML.

## Phase 7: State Relations And Visibility Toggles

Purpose: restore the `State/relations` panel and graph visibility controls.

Implementation:

- Populate rows from engine concept/state snapshot.
- Preserve columns:
  - `+`;
  - `?`;
  - `-`;
  - `T`;
  - relation/name column.
- Toggling a checkbox must:
  - optimistically update state;
  - call the neutral engine toggle operation;
  - refresh graph visibility;
  - rollback and show error on failure.
- Maintain edge visibility and label visibility snapshots for analysis-state save/load.
- Update state label on selected ARG state.
- Implement placeholder only when relation data is genuinely loaded and empty.

Tests:

- Unit tests for toggle snapshot and visibility snapshot.
- Engine service tests for `toggles-updated` and relation rows.
- Component tests for checkbox labels/order.
- Playwright: select ARG state, toggle relation visibility, graph updates or command is sent.

Exit criteria:

- The state panel is fully data-driven and reflects selected state.

## Phase 8: Details, Diagnostics, And Constraint Facts

Purpose: make the Details pane useful again.

Implementation:

- Port details service behavior:
  - selected graph item short/long info;
  - constraint facts as toggle buttons;
  - selected fact callback;
  - trace action button for error traces;
  - diagnostics and source navigation.
- Engine commands must expose details/fact updates neutrally.
- Implement source navigation into editor with line highlighting/scrolling.
- Ensure details pane resize works through the layout store.

Tests:

- Unit tests for details state.
- Component tests for text, facts, selected state, trace button.
- Playwright: click graph node, details update; if check fails, error trace action appears when engine returns one.

Exit criteria:

- Details pane no longer functions as a static log.

## Phase 9: Analysis Actions And Dynamic Backend Menus

Purpose: wire all old analysis and graph actions through the abstract engine API.

Implementation:

- Implement command handlers for:
  - `doUndo`;
  - `doRedo`;
  - `resetDomain`;
  - `diagramDomain`;
  - `pdrStep`;
  - `showReachableStates`;
  - `concreteStep`;
  - `gatherFacts`;
  - `ctiConceptAction`;
  - `reverseStep`;
  - `pathReach`;
  - `reachStep`;
  - `makeConjecture`;
  - `backtrack`;
  - `recalculateGraph`;
  - `rememberGraph`;
  - `exportConjecture`;
  - `weakenInvariant`;
  - `showCheckResult`;
  - `addCheckResultViewActions`.
- Implement concept actions:
  - remove;
  - suppose empty;
  - materialize;
  - materialize edge;
  - split by relation;
  - add projection;
  - add relation;
  - splatter;
  - select concept node.
- Implement ARG actions:
  - execute node actions;
  - execute edge actions;
  - try conjecture choice dialog;
  - try remembered goal choice dialog;
  - decompose opens new ARG sheet;
  - view source opens editor at line.
- Load and render dynamic menu descriptors from engine snapshots/events.

Tests:

- Unit tests for each action service using a scripted engine adapter.
- Contract tests for hosted webui/webengine action payloads.
- Playwright exercises one representative action per action family.

Exit criteria:

- Every visible menu item either performs its old behavior or is deliberately disabled by engine capability metadata with clear feedback.

## Phase 10: Event Trace Sheets

Purpose: restore event trace loading and navigation.

Implementation:

- Port event trace state and components:
  - event tree normalization;
  - expand/collapse;
  - selected row;
  - filter;
  - find forward/reverse;
  - pattern select;
  - pattern add/remove/save/load/clear;
  - multiple event sheets in tab bar.
- Implement `Open Event Trace` file flow.
- Implement visual-only event sheets in loaded analysis state.
- Route event operations through engine commands when engine supports them, local visual fallback only for restored visual-only state.

Tests:

- Unit tests for event normalization and address lookup.
- Component tests for expansion and selection.
- Playwright: open trace fixture, expand node, find pattern, add/remove pattern.

Exit criteria:

- Event traces are a first-class sheet type again.

## Phase 11: Tutorial Pane And Offline Assets

Purpose: restore the tutorial pane without breaking offline-on-a-plane mode.

Implementation:

- Bundle tutorial assets from `goivy/webui/static/tutorial`.
- Implement tutorial URL state:
  - default tutorial URL;
  - input field;
  - navigate;
  - back;
  - forward;
  - reload;
  - close;
  - frame load history recording.
- Implement tutorial button flashing when closed from pane.
- Ensure service worker/offline packaging caches tutorial assets, app bundle, wasm assets, and required CSS/fonts.
- Do not make external tutorial navigation required for core offline use.

Tests:

- Component tests for tutorial history.
- Playwright: open tutorial, navigate bundled page, close pane, button flashes/reopens.
- Offline Playwright: app shell, editor, tutorial asset, wasm asset URLs load with network disabled after first cache.

Exit criteria:

- Tutorial behavior matches old webui and works offline for bundled content.

## Phase 12: Dialogs, Context Menus, Toasts, And Overlays

Purpose: restore the interaction plumbing used by commands.

Implementation:

- Dialog types:
  - ok message;
  - ok/cancel;
  - text;
  - entry;
  - integer/numeric;
  - listbox;
  - button list;
  - dirty close;
  - external disk change;
  - save-as explanation.
- Context menu:
  - screen-positioned menu;
  - headers;
  - separators;
  - disabled items;
  - action callbacks;
  - Escape/outside-click close.
- Toasts:
  - info/success/warning/error;
  - auto-dismiss;
  - manual dismiss.
- Session overlay:
  - loading message;
  - long-running job visible state;
  - server connection lost state.
- FileInputHost:
  - model file;
  - analysis state file;
  - event trace file;
  - event pattern file.

Tests:

- Component tests for each dialog type and resolution value.
- Component tests for context menu keyboard/mouse behavior.
- Component tests for toast timeout/dismiss.
- Playwright: dirty-close dialog and context menu action flow.

Exit criteria:

- Commands can ask questions and show feedback without ad hoc UI.

## Phase 13: Analysis State Save/Load And Persistence

Purpose: restore useful work continuity.

Implementation:

- Port analysis state schema:
  - `analysis_state_format: ivyweb-json`;
  - version;
  - file name/path/content;
  - mode;
  - active sheet;
  - selected ARG node;
  - edge/label visibility;
  - toggles;
  - analysis sheets;
  - event sheets.
- Validate limits:
  - max file bytes;
  - max sheets;
  - max graph elements;
  - max events;
  - max event depth;
  - valid sheet ids.
- Implement save/load UI commands.
- Persist local workspace metadata in IndexedDB:
  - projects;
  - model revisions;
  - recent sessions/files;
  - last engine choice;
  - layout dimensions;
  - editor keymap;
  - offline sync queue.
- Ensure loaded visual-only sheets are marked and protected from live engine actions with clear warning.

Tests:

- Unit tests ported from old `analysisStateService.test.js`.
- IndexedDB tests for metadata persistence and migration.
- Playwright: save analysis state, load it, verify sheets/graphs/editor/toggles restored.

Exit criteria:

- A user can save/restore analysis context without relying on server state.

## Phase 14: Engine Adapter Completion

Purpose: make hosted and browser-local execution equally real behind the UI.

Implementation:

- Finish `HostedWebuiEngine` coverage for old webui API compatibility:
  - session;
  - load/reload content;
  - check/action commands;
  - ARG/concept snapshots;
  - state relations/toggles;
  - dynamic menus;
  - events;
  - analysis/export payloads.
- Finish `HostedGoEngine`/`goivy/webengine` neutral server endpoints:
  - no `webui` route leakage into UI;
  - typed command/action route;
  - typed snapshot route;
  - typed event stream;
  - typed error payloads.
- Finish `BrowserWasmEngine`:
  - stable Big Go `GOOS=js GOARCH=wasm`, no wasip1;
  - Z3 wasm initialized in worker;
  - goivy webengine wasm dispatcher;
  - session lifecycle;
  - model load;
  - check/action/snapshot;
  - event emission;
  - clean worker termination.
- Add capability metadata per adapter:
  - supports cancellation;
  - supports event traces;
  - supports dynamic menus;
  - supports filesystem save;
  - supports offline.
- Product UI uses capability metadata for enabled/disabled commands.

Tests:

- Go tests for `goivy/webengine`.
- Go tests for `goivy/svk/server` engine routes.
- Webui hosted contract tests.
- Node/browser wasm smoke tests.
- Parity harness compares the same model/commands across hosted webui, hosted webengine, and browser wasm where supported.

Exit criteria:

- Switching engines does not change the workbench UI contract.

## Phase 15: Login, Marketing, And Workspace Boundary

Purpose: keep auth/marketing polished while the workbench becomes product-grade.

Implementation:

- Keep login/marketing server separate from the workspace shell.
- Post-login serves the SvelteKit bundle and hydrates authenticated user/account/team/project state.
- Preserve planned auth methods:
  - Google OAuth;
  - Apple OAuth;
  - GitHub OAuth;
  - passkeys;
  - OPAQUE password login using `github.com/bytemare/opaque` and `@cloudflare/opaque-ts`;
  - magic email recovery through Mailgun.
- Enforce:
  - no raw password transport/storage;
  - CSRF tokens for state-changing auth/session requests;
  - SameSite/HttpOnly/Secure cookies;
  - CSP suitable for Svelte/workers/wasm;
  - XSS-safe rendering of model/check/error text.
- Auto-create personal account on first login.
- Keep Accounts separate from Users and Teams.
- First personal billing account/project may be created immediately.

Tests:

- Go auth server tests for session cookies, CSRF, OPAQUE routes, OAuth callback account creation, magic link issuance, passkey challenge lifecycle.
- Playwright login smoke with dev/password or test OPAQUE path.
- Security tests for unescaped user/model text in major panes.

Exit criteria:

- Auth does not block local-first offline workspace use once the app and project are locally available.

## Phase 16: Product Polish And Visual Parity

Purpose: make the app feel finished.

Implementation:

- Match old spacing, typography, dark colors, and panel geometry.
- Ensure no text overlaps inside buttons, menus, tabs, pane headers, status bar, relation table, dialogs.
- Ensure all buttons have hover/active/disabled states.
- Ensure pane resizing cannot collapse the app into unusable states.
- Ensure graph/editor/tutor panes redraw cleanly after resize.
- Ensure app handles:
  - empty model;
  - invalid Ivy model;
  - server unavailable;
  - wasm unavailable;
  - Z3 initialization failure;
  - lost event stream;
  - long-running check;
  - check cancellation when supported.
- Add accessible labels and keyboard behavior for all menus/dialogs.

Tests:

- Playwright screenshots:
  - old webui reference first screen;
  - svk first screen;
  - svk after check failure;
  - svk with tutorial open;
  - svk event trace sheet;
  - svk dialogs/context menus.
- Playwright viewport set:
  - 2048x1180;
  - 1440x900;
  - 1366x768;
  - 1024x768;
  - mobile narrow smoke.
- Add screenshot threshold review; use exact screenshot comparison where stable, perceptual/manual artifact review where graph layout jitter makes exact comparison brittle.

Exit criteria:

- The product looks like the old Ivy webui, but with SvelteKit internals.

## Phase 17: End-To-End User Workflows

Purpose: prove the app is usable, not merely covered by unit tests.

Required Playwright workflows:

- Login/dev-auth route to workspace.
- Start hosted webui engine, load client/server example, run induction check, see Z3-backed failure.
- Edit the model by adding the private invariant, rerun check, see updated result.
- Switch to browser wasm engine, load same model, run supported check/action, see result.
- Toggle relation visibility from state panel.
- Right-click ARG node, run an available action, see graph/status update.
- Right-click concept node, run an available action, see graph/status update.
- Open decomposition/step sheet and switch tabs.
- Save analysis state and load it back.
- Open event trace file, filter/find, save/load patterns.
- Open/close/navigate tutorial.
- Close dirty file and verify save/discard/cancel behavior.
- Simulate server disconnect and verify warning/status recovery.
- Simulate offline after app cache and verify offline workspace plus wasm path.

Required Go workflows:

- `go test ./goivy/svk/server ./goivy/webengine ./goivy/webui -count=1`.
- Add broader package tests if the new webengine code touches shared Go packages.

Required JS workflows:

- `npm run check`.
- `npm run lint`.
- `npm run test:unit -- --run`.
- `npm run test:e2e`.
- `npm run build`.
- `npm run build:wasm`.
- `npm run parity:audit`.
- `npm run unfinished:audit`.

Exit criteria:

- Every workflow is green from a clean local run.

## Sequential Implementation Order

Do the work in this order and do not advance while the current phase is red:

1. Phase 0: parity inventory and failing gates.
2. Phase 1: component skeleton and visual shell.
3. Phase 2: missing UI stores.
4. Phase 3: command registry and complete menus.
5. Phase 4: real session/status/jobs/checks.
6. Phase 5: editor and file lifecycle.
7. Phase 6: graph runtime.
8. Phase 7: state relations/toggles.
9. Phase 8: details/diagnostics/facts.
10. Phase 9: analysis actions and dynamic menus.
11. Phase 10: event trace sheets.
12. Phase 11: tutorial/offline assets.
13. Phase 12: dialogs/context/toasts/overlays.
14. Phase 13: analysis state and persistence.
15. Phase 14: engine adapter completion.
16. Phase 15: login/workspace boundary.
17. Phase 16: polish and visual parity.
18. Phase 17: end-to-end workflows.

Each phase must end with:

- code formatted;
- unit tests added or updated;
- Playwright tests added or updated when UI behavior changed;
- Go tests added or updated when Go endpoints/engine behavior changed;
- all relevant test commands green;
- parity inventory reduced;
- unfinished audit either still failing for known later phases or green at the end.

## Implementation Notes

- Do not port Vue/Pinia mechanically. Port behavior and tests into idiomatic Svelte/SvelteKit.
- Keep `+page.svelte` small.
- Prefer data-driven menu descriptors and command ids over per-button inline logic.
- Preserve old labels initially to reduce user surprise.
- Do not hide missing functionality behind a disabled button unless engine capability metadata says it is unsupported.
- Do not add new local-only UI behavior that cannot be represented through the neutral engine interface.
- Do not let hosted webui compatibility block the native/webassembly architecture. If old webui is odd, adapt in `HostedWebuiEngine`, not in components.
- Keep browser wasm worker as the local-first execution target. Do not revive wasip1.
- Use test fixtures for fake/canned examples. Product screens must come from model/session/engine state.

## Final Release Gate

Before calling Plan 002 implemented:

- Run every command listed in Phase 17.
- Open the app against hosted webui and browser wasm.
- Compare the old screenshot and new screenshot side by side.
- Verify the editor is present.
- Verify sliders/resizers are present and persistent.
- Verify status bar updates after actions and checks.
- Verify menus are complete.
- Verify no production TODO/stub/fake/canned markers remain.
- Verify offline-on-a-plane mode can open the app, edit the model, and run local wasm-supported verification without network.

When this gate passes, the SvelteKit app should no longer feel like a prototype. It should feel like the Ivy webui, rebuilt on a maintainable SvelteKit foundation with local-first execution as a first-class path.
