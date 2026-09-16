# Codex Divergence Audit: `ivy_check` vs `goivy_check`

This is the working audit for behavioral divergences between the Python checker
rooted at `~/ivy/pyivy/ivy/ivy/ivy_check.py` and the Go checker rooted at
`~/ivy/goivy/cmd/goivy_check` plus the `goivy` package.

The audit was started after the `make gold-hermes` divergence in liveness-to-
safety was fixed and verified. It is intentionally source-review oriented:
xtrace tells us where the two implementations reached the same named point, but
this file records the code-level correspondences and known behavioral seams.

## Status Legend

- `fixed`: confirmed divergence with a landed Go fix and regression coverage.
- `open`: confirmed or strongly evidenced divergence that still needs a Go fix.
- `watch`: reviewed difference that may be intentional or not on the live
  checker path, but should stay visible.
- `mapped`: Python file has an identified Go home, but no confirmed divergence
  is recorded here.
- `outside checker`: not part of `goivy_check` parity unless a specific mode or
  UI path calls it.

## Verification Baseline

After the latest liveness-to-safety fix:

- `make test` passed.
- `make gold-hermes` passed.
- `log.gold-hermes` reached the end of the Hermes run without Go/Python
  divergence.

The new unit coverage for that fix is in `check_l2s_finite_sort_test.go`.

## Audit Method

1. Treat `ivy_check.py` as the root, then inspect the Python import closure.
2. Map each Python module to the Go file family that owns the same behavior.
3. Use xtrace names as correspondence anchors where available.
4. Record only source-reviewed divergences in the divergence ledger.
5. Mark broad or generated/file-family mappings separately from confirmed bugs.

The Python import closure rooted at `ivy_check.py` currently includes 70
top-level Python modules. The complete top-level Python module inventory is 106
files, excluding subdirectories such as `z3/`, `tests/`, `utils/`, and `ivy2/`.

## Entry Point Map

| Python | Go | Status |
| --- | --- | --- |
| `ivy_check.py` parameters such as `diagnose`, `action`, `mc`, `trace`, `separate`, `unchecked_properties`, `prioritize`, `profile` | `ivycheck_params.go`, `module_config.go`, `cmd/goivy_check/main.go` | mapped |
| `ivy_check.start` | `check.go:Start`, `check.go:startLoaded`, `cmd/goivy_check/main.go` | mapped |
| `ivy_check.check_module` | `check_isolate_check.go:CheckModule` | mapped |
| `ivy_check.check_isolate` | `check_isolate_check.go:CheckIsolate` | mapped |
| `ivy_check.check_subgoals` | `check_isolate_check.go:CheckSubgoals` | mapped |
| `ivy_check.check_fcs_in_state` | `check.go:CheckFcsInStateWithAG`, `check.go:checkFcsTracePath`, `check.go:checkFcsNormalPath` | mapped |
| `ivy_check.check_conjs_in_state` | `check_isolate_check.go:CheckConjsInStateWithAG` on the live path | mapped |
| `ivy_check.check_safety_in_state` | `check_isolate_check.go:CheckSafetyInStateWithAG` on the live path | mapped |
| `ivy_check.check_temporals` | `check.go:CheckTemporals` | mapped |
| `ivy_check.apply_conj_proofs` | `check.go:ApplyConjProofs` | open |
| `ivy_check.preprocess_assumed_ignored_properties` | `check.go:PreprocessAssumedIgnoredProperties`, `acl.go` | mapped |
| `ivy_check.mc_tactic`, `vmt_tactic` | `check.go:MCTactic`, `check.go:VMTTactic` | mapped |
| `ivy_check.gui_art` | `check_phase7.go:GuiArt`, webui hook | watch |

## Confirmed Divergences And Fixes

### 1. L2S finite-sort lookup used display strings, not declared sort names

Status: fixed.

Python behavior:

- `ivy_l2s.py:374-381`, `ivy_l2s.py:971-1001`, and
  `ivy_ranking.py:315-316` test liveness variables with
  `var.sort.name not in finite_sorts`.
- For an enumerated sort declared as `type ltask = {ready_finish,o3_finish}`,
  Python uses the declared name `ltask`.

Former Go behavior:

- `check_l2s_auto.go`, `check_l2s.go`, and `check_ranking_tactic.go` used
  `sort.String()` or equivalent display text as the finite-sort key.
- For enumerated sorts this can be `{ready_finish,o3_finish}`, while
  `finiteSorts` is keyed by `ltask`.
- Go therefore emitted spurious `l2s_d(T)` or `l2s_a(T)` guards for finite enum
  task variables. The `gold-hermes` divergence showed this at
  `l2s_needed_when_start`, where Python had an empty `And` consequent and Go
  had `And(l2s_d(T))`.

Fix:

- Added `l2sSortIsFinite` in `check_l2s.go:138`; it checks `SortName(sort)`.
- Replaced the guard generation sites in `check_l2s.go`, `check_l2s_auto.go`,
  and `check_ranking_tactic.go`.
- Added regression coverage:
  - `TestL2SFiniteSortLookupUsesDeclaredEnumName`
  - `TestL2SAuto5EnumWorkNeededDoesNotRequireDOrA`
  - `TestRankingEnumWorkCreatedDoesNotRequireD`

### 2. L2S trace hooks mutated fields that Go did not render

Status: fixed. Source details are also recorded in `audit.md`.

Python behavior:

- `ivy_l2s.py` trace hooks set `hidden_symbols`, `renaming`, `pp`, and loop
  markers on a trace object.
- `ivy_trace.py` consumes those fields while rendering the trace.

Former Go behavior:

- L2S hooks wrote analogous fields onto a `MatchHandler`, but the rendered
  output did not consume them.

Fix:

- Go `Trace` now implements the annotation-handler path used by trace
  reconstruction.
- Trace rendering consumes hidden-symbol predicates, pretty-printers, structural
  renaming, loop markers, detailed/non-detailed mode, and model-derived state
  equations.
- Regression coverage lives in the trace and L2S hook tests named in
  `audit.md`.

### 3. Method subgoal trace hooks discarded the transformed failure object

Status: fixed.

Python behavior:

- In `ivy_check.check_subgoals`, method failures run
  `foo = goal.trace_hook(foo)` and then print or display the returned object.

Former Go behavior:

- Go invoked a trace hook on an empty throwaway handler and continued to report
  the original error object.

Fix:

- Added `TraceFailure` and routed method-failure objects through the actual
  hook result.
- `CheckSubgoals` now uses the transformed trace for `OptTrace` and GUI
  diagnostics.

### 4. `CheckFinalCond` displayed input clauses instead of model valuations

Status: fixed.

Python behavior:

- `ivy_trace.check_vc` obtains a model, builds model clauses, creates a
  model-backed `Trace`, replays action annotations into that trace, and displays
  model-derived valuations.

Former Go behavior:

- The Go CTI/final-condition path constructed graph states from pre/post input
  clauses instead of from the satisfying model.

Fix:

- `CheckVC` now constructs a model-backed `Trace`, replays `MatchAnnotation`,
  and renders states through the same valuation path.
- Relation-minimization objectives are also forwarded to the solver.

### 5. L2S trace renaming used text replacement

Status: fixed.

Python behavior:

- `ivy_l2s.renaming_hook` installs a structural symbol renaming.
- `ivy_logic_utils.rename_ast` renames AST nodes, so overlapping names such as
  `l2s_s_1` and `l2s_s_10` are safe.

Former Go behavior:

- The hook used `strings.ReplaceAll` over already-rendered lines, with Go map
  iteration order controlling replacement order.

Fix:

- Go trace renaming is structural and render-time, not text replacement.
- Regression coverage includes overlapping nonce names.

### 6. Solver state for failing final conditions was popped too early

Status: fixed.

Python behavior:

- `ivy_solver.get_small_model` leaves the failing final condition asserted when
  it breaks out of the checker loop on a SAT failure. The returned model
  satisfies the failed condition.

Former Go behavior:

- `GetSmallModelWithCond` popped the final-condition frame before returning a
  model for the failure.

Fix:

- Go now preserves the failing final condition for model extraction and
  shrinking.

### 7. L2S loop-start hook lost the saved-state index

Status: fixed.

Python behavior:

- `ivy_l2s.trace_hook` scans trace states and marks the predecessor of the state
  containing `l2s_saved = true` as the loop start.

Former Go behavior:

- Go scanned a symbol-indexed equation map and always recorded loop start index
  0.

Fix:

- Go scans ordered trace states and marks the same predecessor state Python
  would mark.

### 8. Trace cloning, hidden-symbol defaults, and model universes diverged

Status: fixed.

Python behavior:

- `ivy_trace.Trace.clone()` preserves model and vocabulary.
- Trace states receive the model universe.
- Returned `Trace` objects have a default false hidden-symbol predicate.

Former Go behavior:

- Subtraces could lose model/vocabulary state, states lacked universe data, and
  hidden-symbol defaults could be nil.

Fix:

- Go trace cloning and state construction now preserve model-backed fields and
  defaults.

### 9. Earlier checker-path regressions covered by `check_regression_test.go`

Status: fixed.

These were already fixed before this audit file was created, but they are part
of the `ivy_check.py` vs `goivy_check` divergence record:

- `IsCheckModUnprovable` now matches Python's
  `lf.unprovable == act.check_unprovable.get()` filter.
- The live conjecture/safety checks use AG-aware functions so history, post
  state, and update context are available like Python's `check_conjs_in_state`
  and `check_safety_in_state`.
- `ConvertPostcondsWithUpdate` receives `post.Update` and can rename old-state
  symbols through the update triple, matching `convert_postconds(state, pcs)`.
- Property promotion is followed by `UpdateTheory`, matching
  `im.module.update_theory()`.
- `NewBaseChecker` dualization uses Python's `@` witness prefix rather than
  Go's ordinary double-underscore skolem naming.
- `OnlyCheckUnprovable` guards prevent normal property/init-invariant checks
  in unprovable-only mode.

Coverage:

- `check_regression_test.go` contains the named regression groups.

## Open Or Watch Items

### A. `ApplyConjProofs` can silently pass through proof failures

Status: open.

Python behavior:

- `ivy_check.apply_conj_proofs` calls `pc.admit_proposition(lf, proof)` and then
  maps subgoals through `ivy_compiler.theorem_to_property`.
- There is no local try/except fallback; proof application errors escape.

Current Go behavior:

- `check.go:ApplyConjProofs` attempts AST/module conversion and
  `pc.AdmitProposition`.
- If conversion fails, proof type assertion fails, `AdmitProposition` returns an
  error, or no subgoals are produced, Go appends the original conjecture and
  continues.

Risk:

- A malformed or unsupported proof can be silently ignored by Go where Python
  would stop with an error.

Proposed fix:

- Change `ApplyConjProofs` to return an error and make `CheckIsolate` propagate
  it.
- Only fall back to the original conjecture when Python would do the same: a
  conjecture has no proof entry in `mod.proofs`.
- Add a regression test with a proof node that causes `AdmitProposition` to
  fail.

### B. CLI `GuiArt` is a hook/stub, not Python's Tk main loop

Status: watch.

Python behavior:

- `ivy_check.gui_art` constructs a Tk UI, adds an analysis graph or trace, runs
  the Tk main loop, and exits.

Current Go behavior:

- `check_phase7.go:GuiArt` prepares comparable data and delegates to
  `mod.Cfg.GuiArtHook` when one is registered.
- With no hook, CLI mode prints a diagnostic summary and returns.

Risk:

- `diagnose=true` without a webui hook is not interactive like Python.

Proposed fix:

- Treat this as accepted for CLI-only `goivy_check`, but keep `--trace` and
  webui hooks as the supported diagnostic paths.
- If full Python diagnostic parity is required, make `diagnose=true` fail with
  an explicit message when no hook is installed instead of pretending a GUI was
  launched.

### C. `GuiArt` fixes an upstream Python typo in `default_ui == "art"` mode

Status: watch.

Python behavior:

- In `ivy_check.gui_art`, the `"art"` branch assigns `other_art` but later calls
  `ag.execute(...)`; `ag` is not defined in that function scope.

Current Go behavior:

- `check_phase7.go:GuiArt` uses `otherArt.Execute(...)`, making the branch run.

Risk:

- This is intentionally more usable than Python but not byte-for-byte faithful
  for the rare `"art"` UI mode.

Proposed fix:

- Keep the Go behavior and document it as an intentional upstream-bug repair, or
  gate it behind a compatibility flag if an xtrace comparison ever depends on
  the Python exception.

### D. Convenience wrappers without analysis graph are not Python live paths

Status: watch.

Python behavior:

- `check_conjs_in_state` and `check_safety_in_state` always receive an analysis
  graph and state on the live checker path.

Current Go behavior:

- `CheckConjsInState` and `CheckSafetyInState` exist as convenience wrappers
  and can fall back to direct solver checks without an analysis graph.
- Production isolate checking uses `CheckConjsInStateWithAG` and
  `CheckSafetyInStateWithAG`, which are the Python-equivalent paths.

Risk:

- Tests or future callers may accidentally use the convenience wrapper and miss
  post-state update context, especially for postcondition old-symbol renaming.

Proposed fix:

- Keep wrappers only for tests and simple direct checks.
- Prefer naming or comments that make the AG-aware versions the default for
  checker pipeline work.

## File Correspondence Inventory

This table accounts for every top-level Python `.py` file under
`pyivy/ivy/ivy`. "Go home" names the primary Go file family, not every test.

| Python file | Go home | Status |
| --- | --- | --- |
| `__init__.py` | package initialization is implicit in Go | outside checker |
| `canon.py` | `logic_canon.go`, `module_canon.go`, `ivyutils_canon.go` | mapped |
| `canon_ast.py` | `ast_canon_decl.go`, `logic_canon.go` | mapped |
| `canon_fragment.py` | `fragment_canon.go` | mapped |
| `client_server_example.py` | examples/webui smoke material | outside checker |
| `concept.py` | `webui/webui_concept*.go` | outside checker |
| `concept_alpha.py` | `webui/webui_concept_alpha.go`, `alpha/alpha.go` | outside checker |
| `concept_interactive_session.py` | `webui/webui_concept_session.go`, `webui/webui_session.go` | outside checker |
| `cy_elements.py` | webui JS/Cytoscape rendering, partial Go data structs | outside checker |
| `cy_render.py` | `art_cyrender.go`, `webui/webui_cyrender.go` | outside checker |
| `cy_styles.py` | `webui/webui_cystyles.go` | outside checker |
| `dot_layout.py` | `dotgraph/dot_layout.go` | outside checker |
| `general.py` | no checker-core Go counterpart identified | outside checker |
| `interrupt_context.py` | no checker-core Go counterpart identified | outside checker |
| `iupdr.py` | `iupdr/iupdr.go` | outside checker |
| `ivy.py` | command entrypoints and parser/compiler startup | outside checker |
| `ivy2.py` | legacy Ivy2 staging; no live `goivy_check` counterpart | outside checker |
| `ivy_acl.py` | `acl.go` | mapped |
| `ivy_actions.py` | `actions_action.go`, `actions_update.go`, `actions_match.go`, `actions_transforms.go` | mapped |
| `ivy_alpha.py` | `alpha/alpha.go`, `tactics_ivy_tactics.go` | mapped |
| `ivy_art.py` | `art.go`, `art_phase7.go`, `art_cyrender.go` | mapped |
| `ivy_ast.py` | `ast.go`, `ast_*.go`, `ivylogic_ast_compat.go` | mapped |
| `ivy_auto_inst.py` | `autoinst.go`, `tactics_auto_inst.go` | mapped |
| `ivy_bmc.py` | `bmc.go` | mapped |
| `ivy_check.py` | `check.go`, `check_isolate_check.go`, `check_phase7.go`, `ivycheck_params.go` | mapped |
| `ivy_compiler.py` | `compiler.go`, `compiler_*.go`, `parser_*_helpers.go` | mapped |
| `ivy_compose.py` | `compose/compose.go` | outside checker |
| `ivy_concept_space.py` | `conceptspace/*.go`, `webui/webui_concept_space.go` | outside checker |
| `ivy_congclos.py` | `congclos.go` | mapped |
| `ivy_core.py` | `core/core.go` | mapped |
| `ivy_cpp.py` | `ivy2cpp/*.go` | outside checker |
| `ivy_cpp_types.py` | `ivy2cpp/cpp_types.go`, `ivy2cpp/types.go` | outside checker |
| `ivy_dafny_ast.py` | `dafnygen/*.go` | outside checker |
| `ivy_dafny_compiler.py` | `dafnygen/*.go` | outside checker |
| `ivy_dafny_grammar.py` | `dafnygen/*.go`, generated parser files | outside checker |
| `ivy_dafny_lexer.py` | `dafnygen/*.go`, lexer pieces | outside checker |
| `ivy_dafny_parser.py` | `dafnygen/*.go` | outside checker |
| `ivy_dump.py` | `ivydump/ivydump.go` | outside checker |
| `ivy_ev_parser.py` | `evparser/*.go` | outside checker |
| `ivy_ev_viewer.py` | `webui/webui_ui_evviewer.go` | outside checker |
| `ivy_fragment.py` | `fragment.go`, `fragment_phase7.go`, `fragment_canon.go` | mapped |
| `ivy_graph.py` | `ivyutils_graph.go`, `webui/webui_graph*.go` | outside checker |
| `ivy_graph_ui.py` | `webui/webui_graph_widget.go` | outside checker |
| `ivy_graphviz.py` | graph rendering/export utilities | outside checker |
| `ivy_init.py` | `compiler_ivyinit.go`, `ivyutils_source_file.go` | mapped |
| `ivy_interp.py` | `interp.go`, `interp_phase4.go`, `interp_eval.go` | mapped |
| `ivy_isolate.py` | `isolate.go`, `isolate_*.go` | mapped |
| `ivy_l2s.py` | `check_l2s.go`, `check_l2s_auto.go`, `check_l2s_shared.go`, `check_l2s_hooks.go` | mapped |
| `ivy_launch.py` | `ivylaunch/ivylaunch.go` | outside checker |
| `ivy_lexer.py` | `lexer.go`, `lexer_token.go` | mapped |
| `ivy_libs.py` | `ivylibs/ivylibs.go`, stdlib loader pieces | mapped |
| `ivy_logic.py` | `ivylogic.go`, `ivylogic_*.go`, `logic_*.go` | mapped |
| `ivy_logic_parser.py` | `logicparser.go`, `lalr_logicparser_*.go` | mapped |
| `ivy_logic_parser_gen.py` | generated parser files | mapped |
| `ivy_logic_utils.py` | `logicutil.go`, `logicutil_logic_utils.go`, `module_clauses.go`, `module_subsume.go` | mapped |
| `ivy_lsp.py` | `ivylsp/ivylsp.go` | outside checker |
| `ivy_lsp_client.py` | `ivylsp/ivylsp.go` | outside checker |
| `ivy_mc.py` | `mc_checker.go`, `mc_*.go` | mapped |
| `ivy_module.py` | `module.go`, `module_*.go` | mapped |
| `ivy_parser.py` | `parser_*.go`, grammar files | mapped |
| `ivy_ply_patch.py` | `parser_patch_parser.go` | mapped |
| `ivy_printer.py` | `printer/printer.go` | mapped |
| `ivy_proof.py` | `proof_*.go`, `proof_checker.go` | mapped |
| `ivy_ranking.py` | `check_ranking.go`, `check_ranking_tactic.go` | mapped |
| `ivy_resolution.py` | `resolution.go` | mapped |
| `ivy_shell.py` | `ivyshell/ivyshell.go` | outside checker |
| `ivy_show.py` | `webui/webui_ui_show.go` | outside checker |
| `ivy_smtlib.py` | `z3bridge_solver_smtlib.go` | mapped |
| `ivy_solver.py` | `z3bridge_solver*.go`, `z3bridge_translate*.go` | mapped |
| `ivy_tactics.py` | `tactics.go`, `tactics_ivy_tactics.go`, `proof_tactics.go` | mapped |
| `ivy_temporal.py` | `temporal.go` | mapped |
| `ivy_theory.py` | `theory.go`, `module_theory.go` | mapped |
| `ivy_to_cpp.py` | `ivy2cpp/*.go` | outside checker |
| `ivy_to_lean.py` | `leangen/leangen.go` | outside checker |
| `ivy_to_md.py` | `mdgen/mdgen.go` or no live checker use | outside checker |
| `ivy_trace.py` | `trace.go`, `z3bridge_xtrace.go` | mapped |
| `ivy_transrel.py` | `actions_transrel.go`, `mc_transforms.go` | mapped |
| `ivy_ui.py` | `webui/webui_ui_main.go`, `webaudit.md` | outside checker |
| `ivy_ui_cti.py` | `webui/webui_ui_cti.go`, `trace.go` CTI paths | outside checker |
| `ivy_ui_none.py` | no live checker-core Go counterpart | outside checker |
| `ivy_ui_util.py` | `webui/webui_ui_util.go` | outside checker |
| `ivy_union_find.py` | `unionfind.go` | mapped |
| `ivy_union_find2.py` | `unionfind.go` | mapped |
| `ivy_unitres.py` | `unitres.go`, `z3bridge_solver_unitres_bridge.go` | mapped |
| `ivy_utils.py` | `ivyutils_*.go`, `vprint.go` | mapped |
| `ivy_vmt.py` | `check_vmt.go` | mapped |
| `logic.py` | `logic_*.go`, `ivylogic_*.go` | mapped |
| `logic_sexp.py` | `logic_sexp.go`, `ivylogic_sexp.go` | mapped |
| `logic_util.py` | `logicutil.go`, `logicutil_logic_utils.go` | mapped |
| `proof.py` | `proof_*.go` | mapped |
| `sidecar.py` | Python helper/sidecar, no checker-core Go counterpart | outside checker |
| `tactics.py` | `tactics.go`, `tactics_ivy_tactics.go` | mapped |
| `tactics_api.py` | `tactics.go`, proof tactic registration | mapped |
| `tk_cy.py` | webui/static/browser replacement | outside checker |
| `tk_graph_ui.py` | `webui/webui_graph_widget.go`, `webaudit.md` | outside checker |
| `tk_ui.py` | webui hook/UI replacement | outside checker |
| `token_counter.py` | `lexer_token.go` and parser diagnostics where relevant | mapped |
| `type_inference.py` | `typeinfer_*.go` | mapped |
| `ui_extensions_api.py` | `webui/webui_ext_api.go` | outside checker |
| `widget_analysis_session.py` | `webui/webui_widget_analysis_session.go` | outside checker |
| `widget_cy_graph.py` | `webui/webui_widget_cy_graph.go` | outside checker |
| `widget_dialog.py` | `webui/webui_widget_dialog.go` | outside checker |
| `widget_modal.py` | `webui/webui_widget_modal_messages.go` | outside checker |
| `widget_modal_messages.py` | `webui/webui_widget_modal_messages.go` | outside checker |
| `xtracer.py` | `xtracer/*.go` | mapped |
| `z3_utils.py` | `z3bridge_z3_utils.go` | mapped |

## Python Subdirectory Inventory

These are below `pyivy/ivy/ivy`. They are not top-level modules, but they are
still accounted for here.

| Python file/subtree | Go home | Status |
| --- | --- | --- |
| `tests/test_base.py` | Go package tests and Python oracle tests | outside checker |
| `tests/test_ivy_union_find2.py` | `unionfind.go`, `ivyutils` tests | mapped |
| `ivy2/test1.py` | legacy Ivy2 experiment | outside checker |
| `ivy2/stage2.py` | legacy Ivy2 staging | outside checker |
| `ivy2/stage3.py` | legacy Ivy2 staging | outside checker |
| `ivy2/stage4.py` | legacy Ivy2 staging | outside checker |
| `ivy2/stage5.py` | legacy Ivy2 staging | outside checker |
| `ivy2/stage6.py` | legacy Ivy2 staging | outside checker |
| `ivy2/stage7.py` | legacy Ivy2 staging | outside checker |
| `utils/__init__.py` | implicit Go packages | outside checker |
| `utils/immutables.py` | Go value types and copy-on-write structures where ported | watch |
| `utils/recstruct_object.py` | Go structs/types replace Python record helpers | watch |
| `utils/rectagtuple.py` | Go structs/types replace Python record helpers | watch |
| `utils/try1.py` | experimental utility | outside checker |
| `utils/try2.py` | experimental utility | outside checker |
| `utils/try3.py` | experimental utility | outside checker |
| `utils/try11.py` | experimental utility | outside checker |
| `z3/__init__.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3consts.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3core.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3num.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3poly.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3printer.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3rcf.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3types.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3util.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |

## Python Repository Wrapper Files

These are under `pyivy/` or `pyivy/ivy/` but outside the `ivy` Python package
that feeds `ivy_check`.

| Python-side file | Go home | Status |
| --- | --- | --- |
| `pyivy/requirements.txt` | Go modules in `go.mod` replace Python package deps | outside checker |
| `pyivy/onetime.sh` | local setup helper | outside checker |
| `pyivy/redo.sh` | local setup helper | outside checker |
| `pyivy/LICENSE.original.ivy.txt` | source licensing | outside checker |
| `pyivy/goivy-venv/pyvenv.cfg` | local Python virtualenv metadata | outside checker |
| `pyivy/ivy/setup.py` | Python package setup | outside checker |
| `pyivy/ivy/build_submodules.py` | Python packaging/build helper | outside checker |
| `pyivy/ivy/build_v2_compiler.py` | legacy compiler build helper | outside checker |
| `pyivy/ivy/INSTALL` | Python project documentation | outside checker |
| `pyivy/ivy/PACKAGING.md` | Python project documentation | outside checker |
| `pyivy/ivy/README.md` | Python project documentation | outside checker |
| `pyivy/ivy/Vagrantfile` | Python project/dev environment | outside checker |

## Next Review Targets

These are the next files most likely to hide checker-impacting divergences:

1. `ivy_compiler.py` vs `compiler_*.go`: source loading, proof attachment,
   schema instantiation, object/module expansion, and theorem-to-property.
2. `ivy_actions.py` plus `ivy_transrel.py` vs `actions_*.go`: update
   generation, annotation generation, and pre/post-state symbol naming.
3. `ivy_solver.py` plus `z3_utils.py` vs `z3bridge_solver*.go`: final-condition
   handling, model shrinking, macro-finder settings, and Z3 conversion edge
   cases.
4. `ivy_isolate.py` vs `isolate_*.go`: completeness, cone filtering, mixins,
   trusted/delegate behavior, and public action construction.
5. `ivy_parser.py`, `ivy_logic_parser.py`, and `ivy_lexer.py` vs parser/lexer
   Go files: language-version behavior and parser-global state.

For each target, the review should add a ledger entry only when the Python and
Go behavior can be tied to concrete source lines or xtrace anchors.
