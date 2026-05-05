# Go/Python Port Audit

This file records only confirmed Go/Python behavioral mismatches. Speculative items are excluded.

## 1. L2S Trace Hooks Mutate Fields That Go Never Renders

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_l2s.py:113-122` defines `trace_hook(tr,fcs)` for `l2s_full`. It scans `tr.states`; when it finds `l2s_saved = true`, it sets `tr.states[0 if idx == 0 else idx-1].loop_start = True`.
- `pyivy/ivy/ivy/ivy_l2s.py:1514-1515` defines `renaming_hook(subs,tr,fcs)` as `return tr.rename(...)`.
- `pyivy/ivy/ivy/ivy_l2s.py:1529-1531` has `auto_hook` call `renaming_hook` and then set `tr.pp = ls2_g_to_globally`.
- `pyivy/ivy/ivy/ivy_l2s.py:1570,1588,1602,1616,1691,1696,1701` set `tr.hidden_symbols = temporal_and_l2s` for several auto-diagnostic cases.
- `pyivy/ivy/ivy/ivy_trace.py:44-46` stores `self.renaming` in `TraceBase.rename`.
- `pyivy/ivy/ivy/ivy_trace.py:88-90` reads `self.renaming` and `self.pp` when rendering.
- `pyivy/ivy/ivy/ivy_trace.py:168-169` renders the loop marker when a state has `loop_start`.
- `pyivy/ivy/ivy/ivy_trace.py:172-178` applies `hidden`, structural renaming, and `pp` while rendering state equations.
- `pyivy/ivy/ivy/ivy_trace.py:193-196` calls `to_lines(..., self.hidden_symbols)`, so the hook mutations are part of `str(trace)`.

Go behavior:
- `goivy/check_l2s_hooks.go:57-76` has `markLoopStart` set `handler.LoopStart = 0`.
- `goivy/check_l2s_hooks.go:84-98` implements renaming by `strings.ReplaceAll` over already-collected `handler.Lines`, rather than storing a structural renaming for render time.
- `goivy/check_l2s_hooks.go:119-120` sets `handler.PP = L2sGToGlobally`.
- `goivy/check_l2s_hooks.go:417,444,468,492,715,737,756` set `handler.HiddenSymbols = TemporalAndL2S`.
- `goivy/check_helpers.go:121-158` defines `MatchHandler.Renaming`, `Lines`, `HiddenSymbols`, `PP`, and `LoopStart`.
- `goivy/check_helpers.go:223-252`, `goivy/check_helpers.go:294-342`, and `goivy/check_helpers.go:352-358` append already-rendered strings to `handler.Lines`.
- `goivy/check_helpers.go:364-365` renders a `MatchHandler` with `strings.Join(h.Lines, "\n")`.

Confirmed mismatch:
- `MatchHandler.String()` does not read `HiddenSymbols`, `PP`, or `LoopStart`.
- No Go render path equivalent to Python `TraceBase.to_lines` consumes those hook fields after the hook runs.
- Therefore Go's `l2s_full` loop-start hook, `auto_hook` pretty-printer, and `auto_hook` hidden-symbol settings can be successfully set but still have no effect on printed or displayed trace text.

Why this is not just UI drift:
- These hooks are Python backend trace behavior, not Tcl/Tk presentation. Python applies them in `ivy_trace.Trace.__str__`; Go stores analogous backend fields but never renders them.

## 2. Method-Based Subgoal Failures Discard The Trace Hook Result

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_check.py:840` calls `foo = method()`.
- If `foo` is truthy, `pyivy/ivy/ivy/ivy_check.py:845-846` applies `foo = goal.trace_hook(foo)` when the goal has a trace hook.
- `pyivy/ivy/ivy/ivy_check.py:847-848` then prints `str(foo)` for `opt_trace`.
- `pyivy/ivy/ivy/ivy_check.py:850-851` passes the transformed `foo` to `gui_art(foo)` for diagnostics.

Go behavior:
- `goivy/check_isolate_check.go:801` and `goivy/check_isolate_check.go:891` call `err := method(withLocalMod)`.
- On failure, `goivy/check_isolate_check.go:805` and `goivy/check_isolate_check.go:895` call `applyGoalTraceHook(goal)`.
- `goivy/check_isolate_check.go:649-657` implements `applyGoalTraceHook` by constructing a brand-new mostly empty `MatchHandler`, invoking `goal.TraceHook` on it, and discarding it.
- `goivy/check_isolate_check.go:807-809` and `goivy/check_isolate_check.go:897-899` print the original `err` for `OptTrace`.
- `goivy/check_isolate_check.go:812-815` and `goivy/check_isolate_check.go:902-905` pass the original `err` to `GuiArt`.

Confirmed mismatch:
- Python applies the trace hook to the actual failure object and uses the returned/transformed object for both text and GUI diagnostics.
- Go applies the hook to an empty throwaway `MatchHandler`, never applies it to `err`, and then uses the unmodified `err`.
- The current Go regression test at `goivy/check_isolate_check_tracehook_test.go:8-25` only proves the hook callback is invoked; it does not verify that the method failure object is transformed or displayed with hook effects.

Why this is not just UI drift:
- The mismatch is in backend failure-object plumbing before either text output or GUI output. The Python oracle makes the hook part of the semantic diagnostic result.

## 3. `CheckFinalCond` Rebuilds Displayed States From Input Clauses Instead Of Model Valuations

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:337-350` has `check_vc` ask the solver for a model, build `mclauses = and_clauses(clauses, failed)`, compute `vocab = used_symbols_clauses(mclauses)`, instantiate `Trace(mclauses, model, vocab)`, then replay `act.match_annotation(...)` into that trace.
- `pyivy/ivy/ivy/ivy_trace.py:246-260` converts the Z3 model into ground equality formulas with `clauses_model_to_clauses(...)` and indexes those equations by symbol.
- `pyivy/ivy/ivy/ivy_trace.py:279-297` builds each displayed state by using the annotation environment to pick renamed symbols, structurally renaming each model equality back to the original symbol, and adding those equality formulas to the trace state.
- `pyivy/ivy/ivy/ivy_trace.py:299-304` builds the final displayed state the same way, from the model equations.
- `pyivy/ivy/ivy/ivy_ui_cti.py:156-165` calls `ivy_trace.check_final_cond(...)`, stores the returned `Trace` as `self.g`, rebuilds the UI from it, and views `self.g.states[0]`.

Go behavior:
- `goivy/webui/webui_session.go:1106-1108` uses `goivy.CheckFinalCond(...)` for the webui induction CTI path.
- `goivy/trace.go:590-627` has `CheckVC` obtain a model, but then calls `buildAnnotatedTraceGraph(...)` instead of constructing `NewTrace(...)` or otherwise replaying into `Trace`.
- `goivy/trace.go:647-667` builds an `annotatedTraceGraphHandler` and stores a `HerbrandModel`, but does not convert the model into per-symbol equality lists.
- `goivy/trace.go:701-710` receives the action annotation environment and passes it to `addState`.
- `goivy/trace.go:730-747` ignores that environment and constructs each displayed state from `preClauses`, `postClauses`, or `true_clauses`.

Confirmed mismatch:
- Python's displayed counterexample states are model-derived valuations chosen through the action annotation environment.
- Go's `CheckFinalCond` displayed states are not model-derived valuations. The first state is the pre/history clauses and later states are the final-condition clauses (or true clauses), regardless of the model and regardless of the annotation environment.
- The Go comments in `goivy/trace.go:621-625` claim this matches Python `handler = Trace(...); match_annotation(...); handler.end()`, but the implementation does not do that reconstruction.

Why this is not just UI drift:
- This is the backend object handed to the webui for CTI display. Python's UI receives a reconstructed `Trace`; Go's webui receives an `AnalysisGraph` whose states are populated from the wrong source.

## 4. L2S Trace Renaming Uses Text Replacement And Can Corrupt Overlapping Nonce Names

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_l2s.py:1462-1466` generates L2S replacement constants as `"{name}_{index}"`, for example `l2s_s_1` and `l2s_s_10`.
- `pyivy/ivy/ivy/ivy_l2s.py:1514-1515` installs the inverse map with `tr.rename(...)`.
- `pyivy/ivy/ivy/ivy_trace.py:172-176` applies that map with `lut.rename_ast(...)` while rendering each equation.
- `pyivy/ivy/ivy/ivy_logic_utils.py:196-207` implements `rename_ast` structurally by looking up an application symbol in `subs`; it does not scan or replace substrings in already-rendered text.
- `pyivy/ivy/examples/liveness/tlb.ivy:1561` contains more than ten distinct `$l2s_s ...` binder occurrences in one invariant, so suffixes such as `l2s_s_1` and `l2s_s_10` are realistic, not artificial.

Go behavior:
- `goivy/check_l2s_shared.go:812-829` generates the same `"{name}_{index}"` fresh names and records `cfg.Subs[freshName] = b.String()` for trace display.
- `goivy/check_l2s_hooks.go:84-98` implements `applyRenamingToHandler` by iterating the Go map and running `strings.ReplaceAll(out, fresh, orig)` on already-rendered `handler.Lines`.
- Go map iteration order is intentionally unspecified, so the replacement order is not stable.

Confirmed mismatch:
- If a rendered line contains `l2s_s_10` and the map iteration replaces `l2s_s_1` first, Go rewrites the `l2s_s_1` prefix inside `l2s_s_10` and leaves the trailing `0` attached to the replacement text.
- Python cannot produce that corruption because it renames the `Const("l2s_s_10", ...)` node as a single structural symbol.
- The existing Go tests at `goivy/check_l2s_hooks_test.go:38-64` cover simple string replacements but do not cover overlapping nonce names or nondeterministic replacement order.

Why this is not just UI drift:
- The trace hook is supposed to map backend L2S nonce symbols back to backend named-binder expressions. Replacing substrings in display text is not equivalent to Python's AST-symbol renaming and can produce invalid diagnostic text.

## 5. Go `Trace` Is Not The Annotation Handler That Python `Trace` Is

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:38-46` defines `TraceBase` as the trace object with hook-visible fields such as `hidden_symbols`, `renaming`, and `pp`.
- `pyivy/ivy/ivy/ivy_trace.py:201-236` defines `TraceBase.handle`, `do_return`, `fail`, and `end`; these are the methods used by `act.match_annotation`.
- `pyivy/ivy/ivy/ivy_trace.py:238-304` defines `Trace` as the model-backed subclass that provides `eval`, `get_sym_eqs`, `new_state`, and `final_state`.
- `pyivy/ivy/ivy/ivy_trace.py:346-350` passes the `Trace(...)` instance directly to `act.match_annotation(...)`, then calls `handler.end()`.

Go behavior:
- `goivy/actions_match.go:18-28` defines the Go `AnnotationHandler` interface as `Eval(Expr) bool`, `Handle(ActionsAction, map[NodeKey]Expr)`, `DoReturn(ActionsAction, map[NodeKey]Expr)`, and `Fail()`.
- `goivy/trace.go:267-299` defines `TraceBase.Handle` and `TraceBase.DoReturn` with `map[string]string`, not `map[NodeKey]Expr`.
- `goivy/trace.go:442-458` defines `Trace.Eval(cond Expr) (bool, error)`, not `Eval(Expr) bool`.
- There are no `Trace` methods that implement the `AnnotationHandler` signatures; `rg` finds no `func (t *Trace) Handle(...)`, `func (t *Trace) DoReturn(...)`, or `func (t *Trace) FinalState(...)` equivalents.
- The production paths therefore use substitutes: `goivy/check.go:520-553` uses `MatchHandler`, and `goivy/trace.go:647-674` uses `annotatedTraceGraphHandler`.

Confirmed mismatch:
- In Python, the model-backed `Trace` object is the thing that receives annotation callbacks and later renders or displays the counterexample.
- In Go, the ported `Trace` object cannot receive those callbacks through the current annotation API, so the backend necessarily routes around it.
- This explains why hook fields placed on trace-like objects are split between `TraceBase`, `MatchHandler`, and `annotatedTraceGraphHandler`, with different behavior in each path.

Why this is not just UI drift:
- This is backend trace reconstruction plumbing. The Python oracle has one `Trace` object serving as both annotation handler and diagnostic artifact; Go has multiple incompatible handler artifacts.

## 6. `TraceBase` State Construction Emits Identity/Empty States Instead Of Model-Derived States

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:279-287` builds `sym_pairs` from `self.vocab` and the current annotation environment.
- `pyivy/ivy/ivy/ivy_trace.py:289-297` looks up model equations for each renamed symbol, structurally renames each formula back to the original symbol, and adds those model-derived equations to the trace state.
- `pyivy/ivy/ivy/ivy_trace.py:299-304` builds the final state from all visible vocabulary symbols through the same model-equation path.

Go behavior:
- `goivy/trace.go:330-356` builds `symPairs` from `tb.AnalysisGraph.Domain.Relations` and `tb.AnalysisGraph.Domain.Functions`, plus the string-keyed environment.
- `goivy/trace.go:358-365` ignores the model and creates only identity equations: `Eq(NewConst(sym), NewConst(sym))`.
- `goivy/trace.go:368-370` implements `FinalState` as `AddTraceState(nil)`, so the final state has no model-derived equations at all.
- `goivy/trace.go:411-431` initializes `Trace.Eqs` from raw input `clauses.Fmlas`, not from `ClausesModelToClausesWithModel(...)`; those `Eqs` are not used by `TraceBase.NewTraceStateFromEnv` anyway.

Confirmed mismatch:
- Python trace states display values from the satisfying model.
- Go `TraceBase` states, if built through `TraceBase.Handle`, would display identity formulas or empty states rather than model values.
- This is independent of the current interface mismatch in entry 5: the state-construction logic itself is not a faithful port.

Why this is not just UI drift:
- The equations in a trace state are backend semantic data. They determine what the diagnostic trace says happened, before any web or Tcl/Tk presentation layer sees it.

## 7. `TraceBase.to_lines` Does Not Apply Python's Action Renaming Or Line-Number Rendering

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:73-80` implements `label_from_action`; when a trace renaming is present and the action is not a failure wrapper, it applies `lut.rename_ast(action, renaming)` and `lut.reduce_named_binders(action)` before pretty-printing the action.
- `pyivy/ivy/ivy/ivy_trace.py:88-90` passes the stored trace renaming into `to_lines`.
- `pyivy/ivy/ivy/ivy_trace.py:95-99` emits `str(action.lineno)` before the pretty action text when detailed trace output is enabled and the action has no explicit label.

Go behavior:
- `goivy/trace.go:107-120` accepts a `renaming map[string]string` parameter in `TraceLabelFromAction`, but never uses it.
- `goivy/trace.go:187-193` renders the action label text, but does not emit the action line number before unlabeled actions.

Confirmed mismatch:
- Python detailed traces show renamed action text after L2S-style trace hooks and include action source locations for unlabeled actions.
- Go `TraceBase.String()` ignores the renaming parameter for actions and omits the separate line-number line.
- This is separate from state-equation rendering: even if state reconstruction is fixed, action labels in Go traces will still not match Python.

Why this is not just UI drift:
- The action text and source location are part of the backend trace string generated by Python's `TraceBase.__str__`, not a Tcl/Tk-only presentation detail.

## 8. Non-Detailed Trace Rendering Is Missing

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:31` defines the `detailed` trace parameter with default `True`.
- `pyivy/ivy/ivy/ivy_trace.py:95-99` renders the detailed action form when `detailed` is enabled.
- `pyivy/ivy/ivy/ivy_trace.py:100-128` has a separate non-detailed path for failure summaries and call/env action summaries.
- `pyivy/ivy/ivy/ivy_trace.py:128-152` also renders `DebugAction` events in the non-detailed path.

Go behavior:
- `goivy/module_config.go:132-133` and `goivy/module_config.go:237` define `TraceDetailed` with the same default of `true`.
- `goivy/trace.go:187-210` only emits action/subgraph/spacing output inside `if tb.Domain.Cfg.TraceDetailed`.
- `goivy/trace.go:210-248` only emits state equation output inside `if tb.Domain.Cfg.TraceDetailed`.
- There is no `else` branch equivalent to Python's non-detailed trace behavior.

Confirmed mismatch:
- With detailed tracing disabled, Python still emits concise failure, call/env, and debug-event trace lines.
- With `TraceDetailed` disabled, Go `TraceBase.ToLines` skips all of those Python non-detailed outputs.

Why this is not just UI drift:
- The `detailed` option controls backend trace text generation in Python. Go exposes the same config flag but does not implement the alternate backend behavior.

## 9. State Equation Rendering Omits Python's Rename/Reduce/Filter Pipeline

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:171-181` renders each detailed trace equation through this pipeline: hide by the lhs symbol, apply structural trace renaming, reduce named binders, apply the optional pretty-printer, suppress equations whose lhs numerically reduces to the rhs, then collect the equation for sorted display.
- `pyivy/ivy/ivy_logic_utils.py:196-207` provides the structural `rename_ast` used by that pipeline.
- `pyivy/ivy/ivy_logic_utils.py:369-376` provides `reduce_named_binders`.
- `pyivy/ivy/ivy_logic_utils.py:446-447` provides `reduce_numerically`.

Go behavior:
- `goivy/trace.go:175-178` reads `tb.Renaming` and passes it into `ToLines`.
- `goivy/trace.go:217-229` filters by `hidden(lhsStr)` and applies `pp`, but never applies `renaming`.
- `goivy/trace.go:217-229` also never calls the Go equivalents `ReduceNamedBinders` or `ReduceNumerically`.
- `goivy/logicutil_logic_utils.go:1084-1086` and `goivy/logicutil_logic_utils.go:1533-1535` show that the Go reduction helpers exist, but this render path does not use them.

Confirmed mismatch:
- Python trace equation text can differ from raw model formulas because it structurally renames nonce symbols, collapses named-binder applications, and suppresses tautological numeric display noise.
- Go `TraceBase.ToLines` leaves nonce symbols unreplaced, leaves named-binder applications unreduced, and can print equations Python would suppress.

Why this is not just UI drift:
- This is the backend trace text normalization pipeline in Python. It directly controls the diagnostic content that text mode and UI graph state labels receive.

## 10. `CheckVC` Ignores Relations Requested For CTI Model Minimization

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_ui_cti.py:150-157` computes `rels_to_min` from the CTI UI's "relations to minimize" field, resolves each relation through the first history map, and passes the resulting relation symbols to `ivy_trace.check_final_cond(...)`.
- `pyivy/ivy/ivy/ivy_trace.py:337-339` passes `rels_to_min` through to `slv.get_small_model(...)`.
- `pyivy/ivy/ivy/ivy_solver.py:1423-1436` documents relation minimization as part of small-model search.
- `pyivy/ivy/ivy/ivy_solver.py:1520` minimizes over `chain(sorts_to_minimize, relations_to_minimize)`.

Go behavior:
- `goivy/webui/webui_ui_cti.go:264-268` computes `relsToMin` from the webui CTI field.
- `goivy/webui/webui_ui_cti.go:283` passes `relsToMin` to `goivy.CheckFinalCond(...)`.
- `goivy/trace.go:534-580` passes `relsToMin` to `CheckVC(...)`.
- `goivy/trace.go:590-611` accepts the `relsToMin []string` parameter but never uses it; it calls `slv.GetSmallModel(checkClauses, sortsToMin, nil)`.

Confirmed mismatch:
- Python uses the UI-selected relations to bias/minimize the CTI model.
- Go accepts and forwards the relation-minimization parameter until `CheckVC`, then drops it before solver invocation.
- The resulting CTI model can be larger or different from Python's model because the relation-minimization objective is missing.

Why this is not just UI drift:
- Relation minimization is solver/model-selection behavior. It determines the backend counterexample model that the UI displays.

## 11. Failing Final Condition Is Popped Before Returning The Model

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_solver.py:1488-1499` pushes the solver stack, asserts the current non-assumed final condition, and checks satisfiability.
- `pyivy/ivy/ivy/ivy_solver.py:1500-1504` breaks out of the checker loop when the condition is SAT and `fc.sat()` says not to ignore the failure.
- Because the `break` occurs before `pyivy/ivy/ivy/ivy_solver.py:1507-1508`, the failing final condition remains asserted on the solver stack.
- The later small-model shrinking and `get_model(s)` therefore operate on `clauses + assumptions + failing_final_condition`.

Go behavior:
- `goivy/z3bridge_solver_model.go:210-222` checks the asserted condition and calls `fc.Sat()`.
- `goivy/z3bridge_solver_model.go:223-226` pops the solver stack before breaking on the failing SAT condition.
- `goivy/check.go:607-610` routes normal checker diagnostics through `SmallModelClauses`, which calls this `GetSmallModelWithCond` path.

Confirmed mismatch:
- Python returns a model that satisfies the failed checker condition.
- Go returns a model after the failed checker condition has been popped, so the model only has to satisfy the base clauses plus permanent assumptions.
- Trace reconstruction and CTI details can therefore be built from a model that is not actually a model of the failing condition.

Why this is not just UI drift:
- This is solver-state semantics before any trace or UI rendering. It changes the backend model used as diagnostic evidence.

## 12. Trace Hook Type Cannot Represent Python's Transform-And-Return Contract

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_check.py:412-413` applies a module trace hook as `handler = mod.trace_hook(handler, ffcs)`.
- `pyivy/ivy/ivy/ivy_check.py:845-846` applies a goal trace hook on method failures as `foo = goal.trace_hook(foo)`.
- `pyivy/ivy/ivy/ivy_l2s.py:1514-1515` implements `renaming_hook` by returning `tr.rename(...)`.
- `pyivy/ivy/ivy/ivy_l2s.py:1529-1531` implements `auto_hook` by transforming `tr`, setting fields on it, and returning it.

Go behavior:
- `goivy/check_l2s_hooks.go:14-18` defines `TraceHookFn` as `func(handler *MatchHandler, fcs []Checker)`.
- `goivy/ast_decl_ast.go:38-41` stores that concrete `TraceHookFn` on `LabeledFormula`.
- `goivy/module.go:133-136` stores the same concrete `TraceHookFn` on `Module`.
- `goivy/check.go:552-553` invokes `mod.TraceHook(handler, ffcs)` and cannot receive a replacement handler.
- `goivy/check_isolate_check.go:649-657` invokes `goal.TraceHook(...)` on an empty `MatchHandler` and cannot receive a replacement failure object.

Confirmed mismatch:
- Python trace hooks are diagnostic-object transformers: they can accept the current trace/failure object and return the object that should be printed or displayed.
- Go trace hooks are in-place mutators of one concrete type, `*MatchHandler`, and have no return value.
- This type choice prevents faithful representation of Python hooks that return `Trace`, `AnalysisGraph`, or any other diagnostic object.

Why this is not just UI drift:
- The hook signature defines backend diagnostic plumbing. It determines whether tactics can transform the trace object in the same way Python tactics do.

## 13. `l2s_full` Loop-Start Hook Loses The State Index

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_l2s.py:101-104` installs `trace_hook` for `l2s_full`.
- `pyivy/ivy/ivy/ivy_l2s.py:113-120` scans `tr.states` in order, scans each state's equations, and when it finds `l2s_saved = true`, marks `tr.states[0 if idx == 0 else idx-1].loop_start = True`.
- `pyivy/ivy/ivy/ivy_trace.py:168-169` renders the loop marker at the marked state.

Go behavior:
- `goivy/check_l2s.go:898-901` installs `markLoopStart` for `l2s_full`.
- `goivy/check_l2s_hooks.go:57-76` scans `handler.Eqs`, which is keyed by symbol across the whole `MatchHandler`, not by trace state.
- `goivy/check_l2s_hooks.go:69-71` sets `handler.LoopStart = 0` for any true `l2s_saved` equation.

Confirmed mismatch:
- Python records the loop marker on the state immediately before the state where `l2s_saved` becomes true, except for index 0.
- Go cannot compute that index from `handler.Eqs` and always records index 0.
- This is distinct from entry 1, where `LoopStart` is not rendered at all; even after rendering is wired up, the current Go hook would mark the wrong state for any trace whose saved state is not at index 0 or 1.

Why this is not just UI drift:
- The hook computes backend trace structure: which state starts the repeating cycle. That is semantic trace data, not display styling.

## 14. Trace Cloning Drops Model/Vocabulary State Needed For Subtraces

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:201-207` creates a subtrace with `self.clone()` when tracing nested call/env actions.
- `pyivy/ivy/ivy/ivy_trace.py:213-227` also uses `self.clone()` while reconstructing returns from nested calls.
- `pyivy/ivy/ivy/ivy_trace.py:262-263` implements `Trace.clone()` as `Trace(self.clauses, self.model, self.vocab, False)`, preserving the model and vocabulary needed to build model-derived states.

Go behavior:
- `goivy/trace.go:267-272` creates a subtrace with `tb.Clone()` in the analogous call/env path.
- `goivy/trace.go:280-296` uses `tb.Clone()` in the analogous return path.
- `goivy/trace.go:319-321` implements `TraceBase.Clone()` as `NewTraceBase(tb.cfg, tb.Domain)`, returning a bare `TraceBase`.
- There is no `Trace.Clone()` override that preserves `Clauses`, `Model`, `Vocab`, or `Eqs`.

Confirmed mismatch:
- Python subtraces remain model-backed `Trace` objects.
- Go subtraces lose the `Trace` fields required to evaluate and render model-derived state equations.
- This would break nested call/env traces even if the top-level Go trace object were made to implement `AnnotationHandler`.

Why this is not just UI drift:
- Subtrace cloning controls backend trace reconstruction for calls and environment actions. It decides whether nested trace states can be populated from the same counterexample model.

## 15. `TraceBase.Handle` Does Not Skip `"nowhere"` Internal Actions

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:201-211` ignores an action in `TraceBase.handle` when it has a `lineno` whose filename is `"nowhere"`.
- This prevents internal/generated actions without source meaning from adding trace states.

Go behavior:
- `goivy/trace.go:267-276` always calls `NewTraceStateFromEnv(env)` and records `LastAction = action` in the non-subtrace branch.
- There is no `Filename == "nowhere"` guard in `TraceBase.Handle`.
- `goivy/check_l2s.go:291-292` creates L2S internal actions with `Location{Filename: "nowhere", Line: 0}`.
- `goivy/check_l2s.go:709` creates the generated no-fair-cycle assertion with that internal location.

Confirmed mismatch:
- Python trace reconstruction suppresses `"nowhere"` actions.
- Go trace reconstruction records them as ordinary trace steps.
- L2S generates such internal actions, so the mismatch applies to real trace-producing transformations.

Why this is not just UI drift:
- This decides which backend actions become trace states. It changes the structure of the diagnostic trace before rendering.

## 16. Trace States Do Not Receive The Model Universe

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:56-61` creates a trace state and assigns `state.universe = self.get_universes()` when model universes are available.
- `pyivy/ivy/ivy/ivy_trace.py:265-266` implements `Trace.get_universes()` as `self.model.universes(numerals=True)`.

Go behavior:
- `goivy/trace.go:80-96` creates the `State` in `AddTraceState` but never assigns `state.Universe`.
- `goivy/trace.go:434-439` has `Trace.GetUniverses()`, but `AddTraceState` does not call it and `TraceBase` has no model-aware universe hook.
- `goivy/webui/webui_ui_cti.go:406-408` later reads `ui.AG.States[0].Universe` to build the universe constraint for CTI UI operations.

Confirmed mismatch:
- Python trace states carry the model universe.
- Go trace states constructed by `TraceBase.AddTraceState` do not.
- Webui CTI operations that rely on the first trace state's universe are therefore operating on missing backend data.

Why this is not just UI drift:
- The universe is part of the model-backed trace state. It feeds later CTI/concept-domain computations, not only visual decoration.

## 17. `CheckVC` Returns A `TraceBase` Without Python's Default Hidden-Symbol Function

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:38-42` initializes every `TraceBase` with `hidden_symbols = lambda sym: False`.
- `pyivy/ivy/ivy/ivy_trace.py:193-196` calls `to_lines(..., self.hidden_symbols)`, so the default is always callable.

Go behavior:
- `goivy/trace.go:46-55` initializes `HiddenSymbols` to a non-nil false predicate when callers use `NewTraceBase(...)`.
- `goivy/trace.go:628-634` has `CheckVC` construct `&TraceBase{AnalysisGraph: ag}` directly, bypassing `NewTraceBase(...)`.
- `goivy/trace.go:259-263` passes `tb.HiddenSymbols` to `ToLines`.
- `goivy/trace.go:222-224` calls `hidden(lhsStr)` without a nil guard.

Confirmed mismatch:
- Python-created traces always have a callable hidden-symbol predicate.
- Go traces returned by `CheckVC` can have `HiddenSymbols == nil`.
- Calling `String()` on such a trace can panic instead of producing the Python-equivalent trace text.

Why this is not just UI drift:
- This is the backend trace object's default invariant. Python makes the rendering hook total; Go breaks that invariant on a production construction path.

## 18. `ValueToStr` Omits Python's Array And Destructor Rendering

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:405-407` renders missing values as `"..."`.
- `pyivy/ivy/ivy/ivy_trace.py:408-421` detects array-like values by looking for `<sort>.end` and `<sort>.value`, evaluates the length and each element, and renders values as `[v0,v1,...]`.
- `pyivy/ivy/ivy/ivy_trace.py:422-425` detects datatype/destructor-backed sorts through `im.module.sort_destructors`, evaluates each destructor, and renders `{field:value,...}`.
- `pyivy/ivy/ivy/ivy_trace.py:426-427` falls back to the constant name or generic string.

Go behavior:
- `goivy/trace.go:820-823` matches the missing-value case.
- `goivy/trace.go:824-833` returns the constant name for every `*Const`.
- `goivy/trace.go:828-830` explicitly notes that array/destructor rendering is not implemented.
- The module does maintain destructor metadata in `goivy/module.go:60-61`, so the required backend data exists elsewhere.

Confirmed mismatch:
- Python trace summaries can display concrete array/list-like values and constructor/destructor structured values.
- Go `ValueToStr` collapses those values to the raw constant name.
- This helper is used by Python's non-detailed trace renderer; Go's non-detailed renderer is already missing in entry 8, but the helper port is independently incomplete.

Why this is not just UI drift:
- This is backend value formatting from the model and state evaluator. It affects the semantic readability of trace parameters and debug events.

## 19. `MakeCheckArt` Returns The Wrong Tuple And Does Not Construct Python's Fail State

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:307-318` builds the check graph and executes the environment action to get `post`.
- `pyivy/ivy/ivy/ivy_trace.py:320` constructs `fail = State(expr = fail_expr(post.expr))`.
- `pyivy/ivy/ivy/ivy_trace.py:322` returns `(ag, post, fail)`.
- `pyivy/ivy/ivy/ivy_ui_cti.py:156-157` passes either `post` or `fail` into `check_final_cond(...)`, depending on whether it is checking a conjecture or an assertion failure path.

Go behavior:
- `goivy/trace.go:470-476` says the function matches Python but documents its return as `(ag, preState, postState)`.
- `goivy/trace.go:488-511` creates `preState` and `postState`, but never constructs a fail state equivalent to `fail_expr(post.expr)`.
- `goivy/trace.go:511` returns `ag, preState, postState`.
- `goivy/webui/webui_ui_cti.go:245-246` destructures those returns as `ag, succeed, fail`, following the Python tuple shape.
- `goivy/webui/webui_ui_cti.go:274-280` uses `fail` for the assertion-failure test and `succeed` for conjecture checks.

Confirmed mismatch:
- Python's second return is the post/succeed state, and its third return is the fail state.
- Go's second return is the pre-state, and its third return is the post/succeed state.
- The Go CTI UI path that follows the Python destructuring therefore checks the wrong states, and Go has no `MakeCheckArt` fail state corresponding to Python's `fail_expr(post.expr)`.

Why this is not just UI drift:
- This is the backend state passed into `check_final_cond`. It changes which symbolic transition condition Z3 checks for CTI/safety diagnostics.

## 20. `MakeVC` Does Not Execute The Action Or Preserve Python's Annotation Fixup

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:362-372` builds an analysis graph, creates the pre-state from the precondition formulas, executes the action with checking disabled, and sets `post.clauses = true_clauses()`.
- `pyivy/ivy/ivy/ivy_trace.py:376-379` extracts `history.post` from the executed graph and uses that as the VC base.
- `pyivy/ivy/ivy/ivy_trace.py:380-387` fixes the annotation stack so the returned VC annotation matches the original action.
- `pyivy/ivy/ivy/ivy_trace.py:389` conjoins background theory.
- `pyivy/ivy/ivy/ivy_trace.py:390-398` builds the postcondition clauses, dualizes them with fresh witness constants, and conjoins them to the VC.
- `pyivy/ivy/ivy/ivy_tactics.py:37-39` uses `tr.make_vc(...)` directly in `triple_to_goal`.

Go behavior:
- `goivy/trace.go:793-811` implements `MakeVC` by collecting precondition formulas and syntactically negating postcondition formulas.
- `goivy/trace.go:809` comments that "the TR would be added by the caller".
- `goivy/tactics_ivy_tactics.go:44-60` calls `MakeVC(...)` from `TripleToGoal` and immediately passes the result to `VcToGoal`; it does not add the transition relation later.
- `goivy/trace.go:793-811` does not execute the action, does not call `GetHistory`, does not conjoin background theory, does not dualize with witness constants, and does not perform Python's annotation fixup.

Confirmed mismatch:
- Python `make_vc` returns a transition-relation VC for the action with a carefully repaired annotation.
- Go `MakeVC` returns only `pre ∧ syntactic-not(post)`.
- Any proof/tactic path using `TripleToGoal` can therefore check a materially different VC from Python.

Why this is not just UI drift:
- This is verification-condition construction. It determines the formula being proved, not how a UI renders it.

## 21. `CheckVC` Minimizes Only Formula-Mentioned Sorts Instead Of All Uninterpreted Sorts

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:337-339` calls `slv.get_small_model(clauses, lg.uninterpreted_sorts(), rels_to_min, final_cond=final_cond, shrink=shrink)`.
- `pyivy/ivy/ivy/ivy_logic.py:1493-1496` defines `uninterpreted_sorts()` as all uninterpreted sorts in the current signature that are not interpreted.

Go behavior:
- `goivy/trace.go:603-607` computes `sortsToMin` by walking only the formulas in `checkClauses`.
- `goivy/trace.go:611` passes that formula-derived list to `GetSmallModel`.
- `goivy/ivylogic_context.go:57-68` already has a Python-equivalent `UninterpretedSorts(sig)` helper.
- `goivy/actions_phase4.go:284-294` uses `UninterpretedSorts(m.Sig)` in the `SmallModelClauses` path, but `CheckVC` does not.

Confirmed mismatch:
- Python minimizes over all canonical uninterpreted sorts in the module signature.
- Go `CheckVC` minimizes only uninterpreted sorts that appear syntactically in the conjoined clauses/final condition.
- This can produce different model universes from Python for trace/CTI checks.

Why this is not just UI drift:
- Sort minimization is part of solver model selection. It affects the counterexample model and the universes attached to traces.

## 22. `CheckVC` Ignores `shrink` And Always Requests Small-Model Minimization

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:327-334` threads `shrink` from `check_final_cond(...)` into `check_vc(...)`.
- `pyivy/ivy/ivy/ivy_trace.py:337-339` passes `shrink=shrink` into `slv.get_small_model(...)`.
- Python callers can therefore choose the default non-shrinking behavior or explicitly request shrinking.

Go behavior:
- `goivy/trace.go:591-592` accepts `shrink bool` in `CheckVC(...)`.
- `goivy/trace.go:611` calls `slv.GetSmallModel(checkClauses, sortsToMin, nil)` and never passes `shrink`.
- `goivy/z3bridge_solver_model.go:109-114` implements `GetSmallModel(...)` as `GetSmallModelWithCond(..., nil, true)`, hardcoding `shrink=true`.
- `goivy/z3bridge_solver_model.go:126-130` has the full `GetSmallModelWithCond(...)` entry point that can accept the caller's `shrink` flag, but `CheckVC` does not use it.

Confirmed mismatch:
- Python `check_vc` preserves the caller's requested shrink behavior.
- Go `CheckVC` always asks for the shrinking path, even when called with `shrink=false`.

Why this is not just UI drift:
- Shrinking changes solver model selection. It can change the concrete model and therefore the reconstructed counterexample trace.

## 23. `MatchAnnotation` Does Not Recurse Through `FailAction` Before Marking Failure

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_actions.py:1807-1811` checks `hasattr(action, 'failed_action')`, recursively matches `action.failed_action()` against the annotation, then calls `handler.fail()`.
- `pyivy/ivy/ivy/ivy_interp.py:403-404` defines `fail_action.failed_action()` as the wrapped inner action.
- This means Python first reconstructs the concrete failing path inside the wrapped action, then marks the last reconstructed action as failed.

Go behavior:
- `goivy/actions_fail_action.go:75-77` provides `(*FailAction).FailedAction()`, the direct Go analogue of Python `failed_action()`.
- `goivy/actions_match.go:204-260` handles calls, loops, returns, ignores, and locals.
- `goivy/actions_match.go:263-264` then falls through to `handler.Handle(action, env)`.
- There is no `*FailAction` / `FailedAction()` case in `MatchAnnotation`.

Confirmed mismatch:
- Python treats fail actions as wrappers for annotation matching: recurse into the wrapped action, then call `handler.fail()`.
- Go treats `FailAction` as an opaque action in `MatchAnnotation` and never invokes `handler.Fail()` from that path.
- For composite failing actions, Go can therefore lose the internal action/branch path that Python reconstructs before marking the failure.

Why this is not just UI drift:
- This changes which program action the reconstructed counterexample identifies as failed.

## 24. `MatchAnnotation` Omits Python's Extra Empty-Assume Wrapper For Existential `if` Conditions

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_actions.py:1747-1749` detects `IfAction` whose condition is an `ivy_ast.Some`.
- `pyivy/ivy/ivy/ivy_actions.py:1758-1762` wraps the taken then-branch as `Sequence(AssumeAction(And()), code)` before recursing.
- `pyivy/ivy/ivy/ivy_actions.py:1763-1768` applies the same wrapper to the else-branch.

Go behavior:
- `goivy/actions_action.go:406-414` has `SomeCondition`, explicitly corresponding to Python `Some` / `SomeMin` / `SomeMax` conditions inside `LogicIfAction`.
- `goivy/actions_match.go:123-145` handles `*LogicIfAction` by evaluating the annotation condition and recursing directly into `ThenBody` or `ElseBody`.
- It never checks for `*SomeCondition`, and never inserts the empty `AssumeAction(And())` wrapper that Python uses to match the annotation shape.

Confirmed mismatch:
- Python's annotation walker changes the branch action shape for existential-if conditions before matching the branch annotation.
- Go's annotation walker matches the branch body directly.

Why this is not just UI drift:
- This is in the backend annotation matching algorithm. If the transition relation annotation contains the extra assume node Python expects, Go can walk a different action/annotation structure.

## 25. Unlabeled Environment Choices Pass An Empty EnvAction To The Handler

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_actions.py:1779-1785` handles an unlabeled `EnvAction` choice by constructing `callact = EnvAction(act)`, setting `callact.label = 'call ' + label`, and passing that action to `handler.handle(...)`.
- `pyivy/ivy/ivy/ivy_actions.py:1786` then recurses into the selected branch `act`.
- The handler therefore receives an environment action that still wraps the selected branch action.

Go behavior:
- `goivy/actions_match.go:182-191` handles the same case by constructing `callAct := &LogicEnvAction{}` and setting only `callAct.Label = "call " + label`.
- That `LogicEnvAction` has no branch list, no selected action, and is not constructed through `NewEnvActionOn` or `BuildEnvActionFromAction`.
- The real branch is only used afterward for recursive matching at `goivy/actions_match.go:195-197`.

Confirmed mismatch:
- Python passes a labeled `EnvAction` that contains the selected action.
- Go passes a labeled but empty `LogicEnvAction`.

Why this is not just UI drift:
- `handler.Handle` is a backend callback. A trace handler that needs the selected environment branch, its formals, or its nested action structure cannot recover that information from Go's empty placeholder.

## 26. Trace-Matcher While Expansion Uses The `while` Line For Ranking Checks

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_actions.py:1075-1084` handles `decreases` in `WhileAction.expand(...)`.
- The generated rank snapshot assume, exit assert, and entry assert all receive `decreases.lineno`.
- `pyivy/ivy/ivy/ivy_actions.py:1085-1087` separately assigns havoc actions the enclosing `while` line.

Go behavior:
- `goivy/actions_match.go:442-454` builds the rank snapshot assume inside the trace matcher's local `expandWhile(...)`, but calls `assumeEq.SetLineno(w.GetLineno())`.
- `goivy/actions_match.go:460-470` builds the exit and entry asserts and also assigns `w.GetLineno()`.
- `goivy/actions_action.go:1661-1664` shows `LogicRanking` embeds `ActionBase`, so the Go ranking action can carry its own source location.

Confirmed mismatch:
- Python reports ranking/decreases-generated checks at the `decreases` source line.
- The Go trace matcher's while expansion reports those generated checks at the enclosing `while` source line.

Why this is not just UI drift:
- This affects diagnostic source locations attached to reconstructed traces and assertion failures.

## 27. `MatchHandler` Drops Python's Source Location Formatting

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_utils.py:269-283` defines `LocationTuple.__str__` as a filename-aware string such as `file.ivy: line N: ` on Unix.
- `pyivy/ivy/ivy/ivy_check.py:363` prints matched actions with `print('{}{}'.format(action.lineno, action))`.
- The trace line therefore includes the filename, the word `line`, the line number, and the separator before the action text.

Go behavior:
- `goivy/ast.go:27-45` has a `Location.String()` implementation that matches Python's `LocationTuple.__str__`.
- `goivy/check_helpers.go:340-342` comments that it is matching `print('{}{}'.format(action.lineno, action))`, but actually builds `fmt.Sprintf("%d%v", lineno.Line, action)`.
- The Go line omits the filename, omits `line`, omits the colon/space separator, and directly concatenates the integer line with the action.

Confirmed mismatch:
- Python trace lines preserve full source-location formatting.
- Go `MatchHandler` trace lines preserve only the line integer and can produce ambiguous strings like `42assert ...`.

Why this is not just UI drift:
- This is the textual diagnostic trace emitted by the backend `--trace` path and used to identify the source action behind a failing model.

## 28. `MatchHandler` Silently Skips Line-Zero Actions That Python Handles

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_check.py:348-363` handles an action whenever it has a `lineno` attribute.
- There is no `line > 0` guard in Python's `MatchHandler.handle`.
- `pyivy/ivy/ivy/ivy_utils.py:253-254` defines `nowhere()` as `Location("nowhere", 0)`.
- `pyivy/ivy/ivy/ivy_l2s.py:131`, `pyivy/ivy/ivy/ivy_ranking.py:59`, `pyivy/ivy/ivy/ivy_mc.py:1059`, and `pyivy/ivy/ivy/ivy_vmt.py:59` all create generated actions with line-zero synthetic locations.

Go behavior:
- `goivy/check_helpers.go:294-299` obtains `lineno := action.GetLineno()` and immediately returns if `lineno.Line <= 0`.
- `goivy/check_l2s.go:292` creates the corresponding L2S synthetic location as `Location{Filename: "nowhere", Line: 0}`.
- `goivy/check_ranking.go:177` creates a ranking synthetic location as `Location{Filename: "ranking", Line: 0}`.

Confirmed mismatch:
- Python `MatchHandler` handles actions that have synthetic line-zero locations.
- Go `MatchHandler` drops all line-zero actions before showing state changes or action labels.

Why this is not just UI drift:
- This is backend trace construction for generated proof/checking actions. Dropping those actions changes the sequence of trace events and can hide state changes associated with generated instrumentation.

## 29. `History.SatisfyWithCond` ANDs Final Conditions Where Python ORs Them

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_transrel.py:685-687` checks whether `final_cond` is a list during model-to-state reconstruction.
- For a list of final conditions, Python computes `final_cond = or_clauses(*[fc.cond() for fc in final_cond])`.
- It then builds `all_clauses = and_clauses(post, final_cond)` if a final condition exists.

Go behavior:
- `goivy/actions_transrel.go:2165-2178` handles `finalCond` by collecting every `fc.Cond().Fmlas` into one slice.
- It constructs `fcClauses := NewClauses(condFmlas, nil, nil)`.
- It then builds `allClauses = AndClausesTyped(post, fcClauses)`.

Confirmed mismatch:
- Python constrains model-to-state extraction with `post AND (fc1 OR fc2 OR ...)`.
- Go constrains it with `post AND fc1 AND fc2 AND ...`.

Why this is not just UI drift:
- This path reconstructs the concrete state sequence from the model. ANDing mutually alternative failed checker conditions can change or overconstrain the displayed counterexample states.

## 30. `History.SatisfyWithCond` Converts Reconstructed State Clauses Through A Formula

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_transrel.py:690-695` reconstructs each state's `Clauses` object, renames it, and removes tautological equalities while preserving the `Clauses` representation.
- `pyivy/ivy/ivy/ivy_transrel.py:705` returns `[pure_state(clauses) for clauses in reversed(states)]`.
- `pyivy/ivy/ivy/ivy_transrel.py:92-93` defines `pure_state(clauses)` as `(None, clauses, false_clauses())`, preserving the exact `Clauses` object.

Go behavior:
- `goivy/actions_transrel.go:2188-2194` reconstructs each state's `*Clauses` and appends it to `states`.
- `goivy/actions_transrel.go:2210-2214` builds the returned path with `PureState(cls.ToFormula())`.
- `goivy/actions_transrel.go:183-190` already has `PureStateClauses(clauses)`, which matches Python's `pure_state(clauses)`, but this path does not use it.
- `goivy/module_clauses.go:108-112` shows `ToFormula()` closes/open-formula-converts the clauses rather than preserving the `Clauses` object.

Confirmed mismatch:
- Python returns each reconstructed state as a pure state containing the reconstructed `Clauses`.
- Go converts each reconstructed state to a formula and then `PureState` converts it back through `FormulaToClauses`.

Why this is not just UI drift:
- The normal counterexample path hands these pure states to callers as the concrete path. Re-clausifying can lose clause structure, definitions, annotations, and exact formula shape compared with Python.

## 31. `CheckFinalCond` Masks Missing Annotations Instead Of Failing Like Python

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:326-329` gets `history.post` and immediately asserts `clauses.annot is not None`.
- The annotation is required because `check_vc` later calls `act.match_annotation(action, clauses.annot, handler)` at `pyivy/ivy/ivy/ivy_trace.py:349`.
- If the history lost its annotation, Python treats that as a hard port/logic error.

Go behavior:
- `goivy/trace.go:549-554` gets `history.Post`.
- If `clauses.Annot == nil`, Go replaces it with `EmptyAnnotation{}`.
- `goivy/trace.go:579-580` then calls `CheckVC(...)` with the fabricated empty annotation.

Confirmed mismatch:
- Python refuses to reconstruct a trace from an unannotated transition relation.
- Go silently fabricates an empty annotation and proceeds.

Why this is not just UI drift:
- Annotation loss is exactly what prevents faithful reconstruction of the action path. Fabricating an empty annotation can hide the upstream bug and produce a misleading or truncated counterexample.

## 32. `CheckFinalCond` Rejects `nil` Final Conditions That Python Supports

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_trace.py:325-335` passes `final_cond` through from `check_final_cond(...)` to `check_vc(...)` without rejecting `None`.
- `pyivy/ivy/ivy/ivy_trace.py:337-343` defines `check_vc(..., final_cond=None, ...)` and explicitly handles the `None` case when building `failed`.
- `pyivy/ivy/ivy/ivy_solver.py:1512-1513` checks plain satisfiability when `final_cond is None`.

Go behavior:
- `goivy/trace.go:534-540` returns nil immediately when `finalCond == nil`.
- `goivy/trace_test.go:353-359` currently encodes this Go-only behavior as an expected result.
- `goivy/trace.go:590-592` also gives `CheckVC` a `finalCond *Clauses` parameter, but `CheckFinalCond` never lets the nil case reach it.

Confirmed mismatch:
- Python can ask for a trace/model of the history clauses without an additional final condition.
- Go `CheckFinalCond` treats that case as "no counterexample" before consulting the history or solver.

Why this is not just UI drift:
- This changes the backend semantics of the trace-checking API for callers that use `final_cond=None` as Python allows.

## 33. `History` Stores Forward Renamings By Name Instead Of By Symbol

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_transrel.py:378-384` builds the existential-renaming map as `map1[s] = rename(s, rn)`, where `s` is the actual `Symbol` object.
- `pyivy/ivy/ivy/ivy_transrel.py:629-631` stores that symbol-keyed `map1` in `History.maps`.
- `pyivy/ivy/ivy/ivy_transrel.py:690` later calls `rename_clauses(clauses, inverse_map(renaming))`, still using symbol objects.
- `pyivy/ivy/ivy/ivy_logic.py:967-973` shows why symbol identity matters: a single name can expand to multiple sorted `Symbol(name, sort)` variants for polymorphic/union symbols.

Go behavior:
- `goivy/actions_transrel.go:581-596` initially ports `exist_quant_map` with `map[NodeKey]*Const`, preserving symbol structure and sort.
- `goivy/actions_transrel.go:1172-1196` returns that structural map from `ForwardImageMap`.
- `goivy/actions_transrel.go:2010-2011` defines `LogicRenaming` as `map[string]string`.
- `goivy/actions_transrel.go:2033-2038` converts the structural map into `renaming[s.Name] = renamed.Name`, discarding sort and structural identity before storing it in `History.Maps`.
- `goivy/ivylogic_sig.go:213-226` mirrors Python's same-name sorted variants through `AllSymbolsNamed`.

Confirmed mismatch:
- Python history maps distinguish symbols by the actual symbol object, including sort.
- Go history maps collapse all variants of the same name into one string key.

Why this is not just UI drift:
- `History.SatisfyWithCond` uses these maps to reconstruct earlier concrete states from a model. Colliding same-name symbols can be renamed or inverted incorrectly, corrupting the counterexample path.

## 34. Transition-Relation Axiom Filters Use Symbol Names Instead Of Symbol Identity

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_transrel.py:453-457` computes `pre_ax = clauses_using_symbols(updated, axioms)`.
- `updated` is the list of modified `Symbol` objects from the action update.
- `pyivy/ivy/ivy/ivy_logic_utils.py:1070-1072` keeps formulas/definitions whose AST uses any of those symbol objects.
- The same symbol-object filtering pattern is used for mid/post axioms in `pyivy/ivy/ivy/ivy_transrel.py:347-349`, `pyivy/ivy/ivy/ivy_transrel.py:376`, and `pyivy/ivy/ivy/ivy_transrel.py:529-531`.

Go behavior:
- `goivy/actions_transrel.go:1172-1177` computes `updatedNames := actionsConstNames(updated)` and calls `ClausesUsingSymbolNames(updatedNames, axioms)`.
- The same name-only helper is used at `goivy/actions_transrel.go:688-689`, `goivy/actions_transrel.go:1305`, `goivy/actions_transrel.go:1415-1416`, and `goivy/actions_transrel.go:1744-1745`.
- `goivy/module_ops.go:508-520` implements that filter by matching names.
- `goivy/module_ops.go:590-603` already has `ClausesUsingSymbols(...)`, a structural-symbol filter matching the Python operation more closely.
- `goivy/ivylogic_sig.go:213-226` and `pyivy/ivy/ivy/ivy_logic.py:967-973` both model same-name sorted variants for polymorphic symbols.

Confirmed mismatch:
- Python includes background axioms based on exact updated symbols.
- Go includes background axioms based only on updated symbol names.

Why this is not just UI drift:
- This changes the transition relation used for forward image computation. With same-name sort variants, Go can include irrelevant axioms or miss the intended structural distinction.

## 35. `ShowCounterexample` Rejects The Actual `History.SatisfyWithCond` Result Type

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_check.py:424-426` receives `res = history.satisfy(...)` and passes it directly to `show_counterexample(...)` when diagnosis is enabled.
- `pyivy/ivy/ivy/ivy_check.py:77-83` destructures that result as `universe, path = bmc_res`, copies the ARG path, and assigns each copied state's `value` and `universe`.
- `pyivy/ivy/ivy/ivy_transrel.py:703-705` returns exactly that `(uvs, path)` pair from `History.satisfy(...)`.

Go behavior:
- `goivy/actions_transrel.go:2076-2083` defines `SatisfyResult` as the Go equivalent of Python's `(universe, path)` return value.
- `goivy/actions_transrel.go:2216-2219` returns `&SatisfyResult{Universes: universes, Path: path}`.
- `goivy/check.go:613-618` passes that `res` from `history.SatisfyWithCond(...)` to `ShowCounterexample(...)`.
- `goivy/check.go:879-887` defines a different local `bmcResult` type and type-asserts `bmcRes.(*bmcResult)`.
- That assertion rejects the actual `*SatisfyResult`, so `goivy/check.go:891-897` takes the fallback path.

Confirmed mismatch:
- Python consumes the concrete `(universe, path)` result and assigns model values/universes to the copied counterexample states.
- Go's `ShowCounterexample` does not recognize the result type returned by its own history code and falls back to displaying a copied graph without those values.

Why this is not just UI drift:
- The model-derived state values and universes are the core diagnostic payload of the counterexample. Dropping them makes the displayed counterexample materially less informative than Python's.

## 36. CTI Relation Minimization Does Not Apply The History Renaming

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_ui_cti.py:149` computes `history = ag.get_history(post)` for the selected post state.
- `pyivy/ivy/ivy/ivy_ui_cti.py:150-154` looks up each relation-to-minimize name in the signature, then rewrites it through `history.maps[0].get(relation, relation)` before passing it to `check_final_cond`.
- This maps the user-visible relation to the corresponding symbol in the transition/history clauses being checked.

Go behavior:
- `goivy/webui/webui_ui_cti.go:264-268` builds `relsToMin` as raw strings from `ui.RelationsToMinimize`.
- `goivy/webui/webui_ui_cti.go:283` passes those raw strings to `goivy.CheckFinalCond(...)`.
- `goivy/trace.go:534-580` never performs the Python `history.maps[0]` rewrite either.
- Separately, `goivy/trace.go:611` currently ignores `relsToMin` entirely; even after that is fixed, the web UI caller would still be passing unrenamed relation names.

Confirmed mismatch:
- Python minimizes the history-renamed relation symbols.
- Go's CTI UI passes raw relation names and the core does not translate them through the history map.

Why this is not just UI drift:
- Relation minimization affects the counterexample model selected by Z3. Using the wrong relation symbol changes CTI model selection and therefore the displayed diagnostic state.

## 37. CTI Bounded Check Skips `AnalysisGraph.add_initial_state`

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_ui_cti.py:316-320` creates a fresh analysis graph and calls `ag.add_initial_state(ag.init_cond)`.
- `pyivy/ivy/ivy/ivy_art.py:103-120` implements `add_initial_state`: it creates the initial state from `init_cond`, executes the module initializer list when present, records the initializer expression, and applies the abstractor if supplied.
- `pyivy/ivy/ivy/ivy_ui_cti.py:321-325` then additionally executes an action named `initialize` if one exists.

Go behavior:
- `goivy/webui/webui_ui_cti.go:337-340` creates a fresh analysis graph, directly adds `NewState(ui.Mod, ag.InitCond)`, and sets `post := ag.States[0]`.
- `goivy/art.go:1521-1531` has `(*AnalysisGraph).AddInitialState(...)`, the Go port of Python `add_initial_state`.
- `goivy/webui/webui_ui_cti.go:342-350` only handles the optional action named `initialize`.
- It never calls `ag.AddInitialState(...)`, so module initializer declarations in `mod.Initializers` are skipped.

Confirmed mismatch:
- Python CTI bounded checking starts from the fully initialized analysis graph state.
- Go CTI bounded checking starts from the raw `InitCond` state plus optional `initialize` action, omitting the initializer-list semantics.

Why this is not just UI drift:
- This changes the initial symbolic state for bounded checking. A counterexample can appear or disappear depending on initializer execution.

## 38. BMC/CTI Step Action Does Not Use Python `env_action(None)` Shape

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_ui_cti.py:299` builds the bounded-check step action with `ia.env_action(None)`.
- `pyivy/ivy/ivy/ivy_actions.py:1823-1851` implements `env_action`: it sorts public action names, looks up each action, wraps each branch as `Sequence(act, ReturnAction())`, copies formal params/returns to that sequence, labels string-named branches, and then constructs `EnvAction(*racts)`.
- `pyivy/ivy/ivy/ivy_ui_cti.py:349` executes that step action each BMC step.

Go behavior:
- `goivy/webui/webui_ui_cti.go:353` uses `goivy.BMCEnvAction(ui.Mod)` for the CTI bounded-check step.
- `goivy/bmc.go:180-195` implements `BMCEnvAction` by iterating `mod.PublicActions.All()`, appending the raw action body to `branches`, and returning `NewEnvActionOn(mod.Cfg.ActCfg, branches...)`.
- It does not sort the public action names, does not wrap branches in `Sequence(action, ReturnAction())`, does not copy formal params/returns to branch sequences, and does not assign the Python branch labels.
- `goivy/actions_action.go:2045-2100` already has `BuildEnvAction(...)`, which is the closer Go port of Python `env_action`.

Confirmed mismatch:
- Python BMC/CTI steps execute an environment action whose branches are labeled action-plus-return sequences.
- Go BMC/CTI steps execute an environment action whose branches are the raw action bodies.

Why this is not just UI drift:
- `env_action` shape is consumed by action execution and trace reconstruction. Omitting the return marker and branch metadata changes the backend action/annotation structure used for bounded counterexamples.

## 39. Standalone BMC Skips Python's Optional `initialize` Action

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_bmc.py:41-45` creates an analysis graph and calls `ag.add_initial_state(ag.init_cond)`.
- `pyivy/ivy/ivy/ivy_bmc.py:46-48` then checks whether an action named `initialize` exists and executes it as `ag.execute(init_action, None, None, 'initialize')`.
- This mirrors the CTI bounded-check setup in `pyivy/ivy/ivy/ivy_ui_cti.py:316-325`.

Go behavior:
- `goivy/bmc.go:109-113` creates an analysis graph and calls `ag.AddInitialState(nil, nil)`.
- There is no corresponding `if initialize exists { ag.Execute(..., "initialize") }` step before the BMC loop.

Confirmed mismatch:
- Python standalone BMC starts after both initializer-list execution and optional `initialize` action execution.
- Go standalone BMC starts after initializer-list execution only.

Why this is not just UI drift:
- This changes the initial state explored by bounded model checking when a module defines an `initialize` action.

## 40. Standalone BMC Computes Assertion-Failure Clauses But Does Not Put Them In The History

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_bmc.py:58-61` executes one step, constructs `fail = ivy_interp.State(expr = ivy_interp.fail_expr(post.expr))`, and calls `ivy_trace.check_final_cond(ag, fail, ilu.true_clauses(), [], True)`.
- `pyivy/ivy/ivy/ivy_interp.py:407-408` defines `fail_expr(expr)` as `action_app(fail_action(expr.rep), expr.args[0])`, preserving the failing action expression so `get_history` can replay it.
- `pyivy/ivy/ivy/ivy_trace.py:326-335` then uses `ag.get_history(fail)` to build the failing-action history.

Go behavior:
- `goivy/bmc.go:147-152` computes `failUpdate := computeFailUpdate(stepAction, mod)`, extracts `failClauses := failUpdate.TR`, creates `failState := NewState(mod, failClauses)`, sets `failState.Pred = post.Pred`, and calls `CheckFinalCond(ag, failState, TrueClauses(nil), nil, true)`.
- `goivy/art.go:744-755` uses a state's own clauses only in the base case where `state.Pred == nil`.
- `goivy/art.go:763-820` handles a state with a predecessor by recursing to the predecessor and then adding a forward step only if `state.Update != nil` or lazy-computable `state.Action != nil`.
- The BMC fail state has `Pred` but no `Update`, no `Action`, and no `Prov`, so `GetHistory` does not add the computed `failClauses` to the history.

Confirmed mismatch:
- Python represents the assertion-failure check as a failing action expression in the analysis graph history.
- Go computes failure clauses separately but passes a state shape that causes `GetHistory` to ignore those clauses.

Why this is not just UI drift:
- The safety/failure BMC query can be checking reachability of the predecessor history with `true` as the final condition instead of checking the computed assertion-failure condition.

## 41. Initializer Assertion Checking Does Not Reproduce Python's Leaked-Variable Semantics

Status: confirmed conformance bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_check.py:610-616` prints initializers with `for actname,action in sorted(mod.initializers, key=lambda x: x[0])`, leaving Python's leaked loop variable `action` bound to the last initializer action.
- `pyivy/ivy/ivy/ivy_check.py:627-630` then computes `guarantees = [sub for sub in action.iter_subactions() if isinstance(sub,(act.AssertAction,act.Ranking)) for action in mod.initializers]`.
- Because the leftmost comprehension iterable is evaluated first, Python iterates subactions of the leaked last initializer, then repeats each selected subaction once per initializer tuple.

Go behavior:
- `goivy/check_isolate_check.go:294-307` iterates every initializer action and collects each assert/ranking subaction once.
- `goivy/check_isolate_check.go:268-292` explicitly calls the Python behavior a bug and says Go intentionally diverges.

Confirmed mismatch:
- With multiple initializers, Python checks duplicated assert/ranking subactions from the lexicographically last initializer only.
- Go checks assert/ranking subactions from all initializers once.

Why this is not just UI drift:
- This changes which initializer assertions are checked and reported. Under the project's Python-as-oracle rule, even this Python bug is a backend conformance issue unless a deliberate, test-accounted exception is created.

## 42. `SubgoalAction` Falls Through To `NullUpdate`

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_actions.py:360-395` defines `AssertAction.action_update`, which turns assertions into transition/precondition clauses.
- `pyivy/ivy/ivy/ivy_actions.py:416-421` defines `SubgoalAction(AssertAction)` and only overrides `clone`, so `SubgoalAction.int_update` uses base `Action.int_update` plus inherited `AssertAction.action_update`.

Go behavior:
- `goivy/actions_action.go:1509-1518` defines `LogicSubgoalAction` embedding `LogicAssertAction`.
- `goivy/actions_update.go:415-417` implements `LogicSubgoalAction.ActionUpdate` by delegating to `LogicAssertAction.ActionUpdate`.
- `goivy/actions_update.go:1163-1238` dispatches `IntUpdate` with explicit type cases, but has no `case *LogicSubgoalAction`.
- Therefore `LogicSubgoalAction` reaches `goivy/actions_update.go:1234-1237` and returns `NullUpdate()`.

Confirmed mismatch:
- Python treats a subgoal as an assertion for update/precondition generation.
- Go has the assertion logic available but bypasses it in the central dispatcher, making subgoal execution an identity/no-failure update.

Why this is not just UI drift:
- Subgoals participate in verification conditions. Returning `NullUpdate` can drop the proof obligation entirely.

## 43. Compiler Property Proof Errors Are Logged And Ignored

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_compiler.py:2298-2299` calls `prover.admit_proposition(prop, pmap[prop.id])` for properties with proofs.
- `pyivy/ivy/ivy/ivy_compiler.py:2352-2354` calls `prover.admit_proposition(..., ivy_ast.ComposeTactics())` for named/unproved properties admitted as proof obligations.
- These calls are not inside `try/except`; an invalid proof tactic or proof-checking error propagates out of compilation.

Go behavior:
- `goivy/compiler_phase6.go:2321-2329` calls `prover.AdmitProposition(prop, pfNode)`, but on error prints `check_properties: proof error...`, keeps the original property in `mod.LabeledProps`, and continues compilation.
- `goivy/compiler_phase6.go:2414-2424` similarly logs errors from admitting named/unproved properties and continues.

Confirmed mismatch:
- Python compilation fails when property proof admission fails.
- Go compilation can succeed while silently demoting the failed proof path into an unproved property path.

Why this is not just UI drift:
- This changes the accepted program set and can hide invalid proof scripts. The golden oracle expects proof-checking failures to stop the compile path, not become diagnostic noise.

## 44. Action Compile Failures Register Empty Fallback Actions

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_compiler.py:938-971` implements `compile_action_def`.
- It compiles action parameters and returns with `compile_const(...)`, compiles the body with `sortify(...)`, checks for illegal free variables in calls, attaches formals/returns, and returns the compiled action.
- The errors in this path are not swallowed. For example, `pyivy/ivy/ivy/ivy_compiler.py:967` raises `IvyError` for a call with free variables, and many callees in the same compiler module raise `IvyError` for invalid symbols, sorts, arity, or action use.
- `pyivy/ivy/ivy/ivy_compiler.py:1630-1632` shows the general compiler policy for bad action references: raise, do not synthesize a placeholder action.

Go behavior:
- `goivy/compiler_action.go:96-136` catches formal/return compilation errors, substitutes `TopS` placeholder constants, and returns an empty `Sequence` plus an error.
- `goivy/compiler_action.go:159-170` catches body sortification errors and returns an empty `Sequence` with the action's formals/returns.
- `goivy/compiler_ivy_compile.go:534-545` receives those errors during `ARGSetup`, logs `COMPILE_FAIL`, and still registers the fallback action, creating a bare `Sequence` if needed.

Confirmed mismatch:
- Python rejects invalid action definitions during compilation.
- Go can keep compiling and installs an empty action under the requested name.

Why this is not just UI drift:
- This changes executable semantics. A malformed action can become a no-op in Go instead of a Python compile failure, which can make later verification results meaningless.

## 45. `AnalysisGraph.Unreachable` Bypasses Python's Module Order Relation

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_art.py:237-245` implements `unreachable` by creating a false covering state `State(covered_node.domain, [[]])` and then calling `covered_node.domain.order(covered_node, covering_node)`.
- `pyivy/ivy/ivy/ivy_interp.py:67-70` defines `module_order` as `implies_state(state1.value, state2.value, axioms, self.relations)` using the module background theory and relation metadata.
- `pyivy/ivy/ivy/ivy_art.py:225-232` uses the same `domain.order` relation for ordinary cover checks.

Go behavior:
- `goivy/art.go:485-497` correctly ports ordinary cover checks through `ModuleOrder(...)`.
- `goivy/art.go:522-545` implements `Unreachable` separately by translating `node.Clauses.ToFormula()` and calling `t.IsSat(fmla)`.
- That direct SAT check does not use `ModuleOrder`, does not build an explicit false state, and does not pass the background theory/relation context that Python's `implies_state` path uses.

Confirmed mismatch:
- Python unreachable checking asks whether the node is covered by false under the module order.
- Go unreachable checking asks whether the node clauses alone are unsatisfiable.

Why this is not just UI drift:
- Background theory can make a state unreachable even when its raw clauses are satisfiable in isolation. Go can therefore fail to mark states unreachable that Python would collapse to false.

## 46. Field-Action Type Errors Become No-Op Updates

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_actions.py:736-741` implements `make_field_update`.
- It requires the field symbol to be a binary relation and raises `IvyError(self, "field ... must be a binary relation")` when that check fails.
- `pyivy/ivy/ivy/ivy_actions.py:752-781` routes `AssignFieldAction`, `NullFieldAction`, and `CopyFieldAction` through this helper.

Go behavior:
- `goivy/actions_update.go:1040-1049` implements `makeFieldUpdateFunc`.
- If the field expression is not a `*Const`, is not a `LogicFunctionSort`, is not relational, or does not have binary domain arity, it returns `NullUpdate()`.
- `goivy/actions_update.go:1060-1081` routes `LogicAssignFieldAction`, `LogicNullFieldAction`, and `LogicCopyFieldAction` through that helper.

Confirmed mismatch:
- Python rejects malformed field updates.
- Go silently turns them into identity/no-failure updates.

Why this is not just UI drift:
- An invalid field update can be erased from the transition relation instead of stopping with the Python error. That can hide real modeling or compiler bugs.

## 47. Assignment Update Errors Become No-Op Updates

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_actions.py:550-553` raises `IvyError` when an assignment has too many LHS parameters.
- `pyivy/ivy/ivy/ivy_actions.py:564-567` raises `IvyError` when RHS variables are not contained in the LHS variables.
- `pyivy/ivy/ivy/ivy_actions.py:569-576` type-checks the RHS and raises `IvyError` on an individual/boolean assignment sort mismatch.

Go behavior:
- `goivy/actions_update.go:463-470` returns `NullUpdate()` when `xtra < 0`.
- `goivy/actions_update.go:499-507` returns `NullUpdate()` when RHS variables are not in the LHS.
- `goivy/actions_update.go:510-530` proceeds to destructor/variant/standard assignment without an equivalent of Python's `type_check(domain,rhs)` plus individual/boolean mismatch error.

Confirmed mismatch:
- Python rejects malformed assignments.
- Go can erase malformed assignments or let sort-mismatched assignments proceed into later update construction.

Why this is not just UI drift:
- Assignment updates are the core transition relation. Turning malformed assignments into identity updates can make bad Ivy programs verify in Go.

## 48. Hierarchical Assignment Expansion Does Not Postfix The Whole AST

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_actions.py:542-548` handles hierarchical assignment by building `asgns = [postfix_atoms_ast(self, Atom(x, [])) for x in domain.hierarchy[n]]` and recursively computing each child assignment update.
- `pyivy/ivy/ivy/ivy_ast.py:1768-1770` implements `postfix_atoms_ast` as an AST rewrite.
- `pyivy/ivy/ivy/ivy_ast.py:1678-1684` defines `AstRewritePostfix.rewrite_atom` as `compose_atoms(atom, self.post)`, meaning every atom in the assignment AST is postfixed with the child object name.

Go behavior:
- `goivy/actions_update.go:437-459` handles hierarchy by creating `childSym := NewConst(childName, TopS)`, building `childLHS := MustApply(childSym, nodeArgs(lhs)...)`, and leaving `childRHS := rhs`.
- It does not structurally postfix the original assignment AST and does not rewrite atoms appearing in the RHS.

Confirmed mismatch:
- Python expands a hierarchical assignment by rewriting the whole action AST for each child.
- Go rewrites only the LHS head and preserves the original RHS unchanged.

Why this is not just UI drift:
- Hierarchical assignments can reference object-local symbols on the RHS. Python redirects those references to the child; Go can read from the parent/original symbols while writing to the child, producing a different transition relation.

## 49. `ReachState` Drops Python's Tagged Disjunct And Model-Derived Under-State

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_interp.py:267-271` implements `join_unders` with `tagged_or_clauses('__pre', *[s.clauses for s in state.unders])`, preserving which predecessor under-approximation branch a model satisfies.
- `pyivy/ivy/ivy/ivy_interp.py:284-305` implements `reach_state`: it computes the forward image, obtains a concrete model with `get_model_clauses(img)`, finds the true predecessor disjunct with `find_true_disjunct(pre, m.eval)`, converts the model back into clauses with `clauses_model_to_clauses`, and calls `add_under(state, post, state.unders[idx], universe)`.

Go behavior:
- `goivy/interp_helpers.go:92-101` implements `JoinUnders` with plain `OrClausesTyped(...)`; it does not use the already-ported `TaggedOrClauses` from `goivy/module_subsume.go:354-361`.
- `goivy/interp_helpers.go:124-151` implements `ReachState` by checking `t.IsSat(imgClauses.ToFormula())`.
- On satisfiable, it calls `AddUnder(state, imgClauses, nil, nil)`: it stores the whole symbolic image, no selected predecessor under-state, and no model universe.

Confirmed mismatch:
- Python records a concrete model-derived under-state linked to the under-approximation predecessor branch that made it reachable.
- Go records an untagged symbolic image with no predecessor and no universe.

Why this is not just UI drift:
- Under-approximation reachability feeds later ARG/reachability reasoning and diagnostics. Dropping the selected predecessor and model payload prevents Python-style reconstruction of how the state was reached.

## 50. `ReachStateFromPred` Smooths Over Python's Undefined-Local Failure Path

Status: confirmed conformance bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_interp.py:307-317` calls `post = reach_state(state, clauses)` and returns it if reachable.
- If not reachable, `pyivy/ivy/ivy/ivy_interp.py:318` calls `reverse_interpolant_case(clauses, state.update, pre, axioms, state.domain.functions)`.
- The names `pre` and `axioms` are not assigned in `reach_state_from_pred`; they were locals inside `reach_state`. Therefore Python raises on this failure path before it can compute the reverse interpolant.

Go behavior:
- `goivy/interp_helpers.go:159-178` handles the not-reachable case by recomputing `axioms := state.Domain.BackgroundTheory(...)` and `pre := JoinUnders(state.Pred())`.
- It then calls `ReverseInterpolantCase(...)` and can return an `UnsatCoreWithInterpolant`.

Confirmed mismatch:
- Python's not-reachable path fails due undefined locals.
- Go repairs that path and returns a structured interpolant result when available.

Why this is not just UI drift:
- This changes control flow and error behavior in the interpreter. If the project wants to intentionally fix this Python bug, that needs to be an explicit oracle exception; otherwise the Go port is not mechanically conformant on this path.

## 51. `Diagram` Ignores Python's Implied/Weakening/Upward-Close Parameters

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_interp.py:339-347` implements `diagram(state, clauses, implied=false_clauses(), extra_axioms=None, weaken=True, upward_close=True)`.
- It conjoins `extra_axioms` into the background theory and calls `clauses_model_to_diagram(clauses, is_skolem, implied, axioms=axioms, weaken=weaken, upward_close=upward_close)`.
- `pyivy/ivy/ivy/ivy_solver.py:1702` exposes those parameters in the solver diagram API.

Go behavior:
- `goivy/interp_helpers.go:275-291` accepts `implied`, `weaken`, and `upwardClose` in `Diagram(...)`.
- It then calls `slv.ClausesModelToDiagram(clauses, isSkolem, axioms)` at `goivy/interp_helpers.go:287`, dropping all three parameters.
- `goivy/z3bridge_solver_clauses.go:147-157` already has `ClausesModelToDiagramFull(...)`, which can pass `implied`, `weaken`, and `upwardClose`.

Confirmed mismatch:
- Python diagram construction is parameterized by implied facts and weakening/upward-closure choices.
- Go's public `Diagram` signature accepts those choices but always uses defaults.

Why this is not just UI drift:
- Diagrams are used as model-derived symbolic states. Dropping `implied` and closure controls changes which facts appear in the resulting abstraction.

## 52. Action Decomposition Skips Python's State-Sensitive Local/Call/While Logic

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_interp.py:482-523` decomposes an action app by calling `action.decompose(state1.value, state2.value)`, then uses the returned `(pre, acts, post)` triples to build a history and model-derived intermediate states.
- `pyivy/ivy/ivy/ivy_actions.py:1187-1190` implements `LocalAction.decompose` by hiding local symbols in both pre and post states before recursing into the body.
- `pyivy/ivy/ivy/ivy_actions.py:1112-1113` implements `WhileAction.decompose` by expanding the while action and decomposing that expansion.
- `pyivy/ivy/ivy/ivy_actions.py:1417-1435` implements `CallAction.decompose` by resolving the callee, hiding formal params/returns, constraining actual/formal values in pre/post states, renaming hidden returns, cloning the callee body, and returning that transformed triple.

Go behavior:
- `goivy/interp_phase4.go:211-240` calls `DecomposeWithState(...)` and then builds a history from the returned triples.
- `goivy/actions_action.go:1481-1489` handles `LogicLocalAction` by recursing into the body without hiding local symbols; the comment explicitly says the Python hiding is not implemented.
- `goivy/actions_action.go:1491-1497` handles `LogicWhileAction` by using the body as a single step instead of expanding the while action.
- `goivy/actions_action.go:1499-1501` has no `LogicCallAction` case, so calls fall through to the atomic default and never apply Python's formal/actual state constraints.

Confirmed mismatch:
- Python action decomposition transforms the pre/post state triples for locals, calls, and loops before reconstructing the subtrace.
- Go's decomposition is mostly structural and leaves those state transformations out.

Why this is not just UI drift:
- Decomposition is the backend for step-in and reconstructed trace display. If the pre/post triples are wrong, the web UI can display plausible-looking subgraphs whose transitions are not the Python transitions.

## 53. VMT Transition Generation Uses `NullUpdate` For Every Action

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_vmt.py:38-41` computes a VMT action transition by calling `action.update(im.module, None)` or by doing the same under `UnrollContext` for `fsmc`.
- `pyivy/ivy/ivy/ivy_vmt.py:234-239` calls `action_to_tr` for every public action and for the initializer, then builds the VMT transition and initial-state formulas from those real updates.

Go behavior:
- `goivy/check_vmt.go:36-60` implements `actionToTR` and says it corresponds to Python's `action_to_tr`.
- `goivy/check_vmt.go:50` calls `computeUpdate(m, action)`.
- `goivy/check_vmt.go:756-763` implements `computeUpdate` as a stub returning `NullUpdate()` for every action.

Confirmed mismatch:
- Python VMT export uses the actual symbolic transition relation of each action.
- Go VMT export currently emits identity/no-failure updates regardless of the action body.

Why this is not just UI drift:
- The generated VMT transition system can be unrelated to the Ivy program. Model-checking results or exported VMT files from this path are not Python-conformant.

## 54. VMT Invariant Collection Skips Proof Tactics

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_vmt.py:204-218` creates a `ProofChecker`, builds `pmap` from `mod.proofs`, and for each checked conjecture either admits its proof and extends `conjs` with the resulting subgoals or appends the conjecture directly.

Go behavior:
- `goivy/check_vmt.go:531-539` builds a proof map.
- `goivy/check_vmt.go:547-555` then ignores that map; the comment says proof tactic handling is skipped, and the original conjecture is always appended directly.

Confirmed mismatch:
- Python VMT checks proof-generated subgoals when a checked conjecture has a proof.
- Go VMT checks the original conjecture and does not apply the proof tactics.

Why this is not just UI drift:
- This changes the verification condition emitted to VMT. Proof scripts that are part of Python's VMT path are not represented in Go's model-checking query.

## 55. `CreateConjActions` Omits Python's Object-Invariant Interference Check

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_compiler.py:2404-2427` computes which actions must preserve each object invariant/conjecture.
- `pyivy/ivy/ivy/ivy_compiler.py:2428-2450` then, when `iso.do_check_interference` is enabled, checks that isolate actions do not depend on invariants that might be invalidated by cross-isolate calls, raising `IvyError` on interference.

Go behavior:
- `goivy/compiler_ivy_compile.go:1702-1771` ports the action-set construction portion.
- It ends after assigning `mod.ConjActions[origLbl] = actionNames`.
- There is no equivalent of the Python `action_isos` / `roots` / cross-isolate interference loop in this function.
- `goivy/compiler_batch_f_test.go:499-523` contains a test comment explicitly saying this Python interference detection is not yet ported.

Confirmed mismatch:
- Python rejects object-invariant interference during conjecture-action creation.
- Go records the conjecture action map but skips the interference error path.

Why this is not just UI drift:
- This changes whether a modular Ivy program is accepted. Go can accept an isolate configuration that Python rejects as unsound for object invariants.

## 56. Empty Tagged Disjunctions Return `true` Instead Of Python's `Or()`

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_logic_utils.py:1477-1484` implements `tagged_or_clauses` by delegating directly to `or_clauses_int(...)`.
- `pyivy/ivy/ivy/ivy_logic_utils.py:1413-1415` builds `vs = []` and then `fmlas = [Or(*vs)] + ...` when there are no arguments.
- In Ivy logic, `Or()` with no arguments is the empty disjunction, i.e. false.
- `pyivy/ivy/ivy/ivy_interp.py:267-271` uses this function for `join_unders`, so a state with no under-approximations gets a false predecessor under-approximation.

Go behavior:
- `goivy/module_subsume.go:354-357` special-cases `TaggedOrClauses` with no args and returns `TrueClauses(nil)`.
- `goivy/interp_helpers.go:92-96` also special-cases `JoinUnders` with no under-approximations and returns `TrueClauses(nil)`.

Confirmed mismatch:
- Python's empty under-approximation join is false.
- Go's empty under-approximation join is true.

Why this is not just UI drift:
- Under-approximations represent known reachable states. Treating an empty known-reachable set as true says every state is already reachable, which can completely change reachability and interpolation behavior.

## 57. Isolate Import Wrapper Creation Drops Python's Attribute/Import/`extra_with` Updates

Status: confirmed bug.

Python behavior:
- `pyivy/ivy/ivy/ivy_isolate.py:1840-1859` creates an `imp__...` wrapper for each out-call: it applies the implementation map, copies `spec`/`impl`/`private` attributes from the implementation action to the external wrapper name, creates a `CallAction` with formal params/returns and source line, installs an empty external action with copied formals and line number, and stores both in `mod.actions`.
- `pyivy/ivy/ivy/ivy_isolate.py:1860-1863` conditionally appends an `ImportDef(extname, '')` and appends the implementation name to `extra_with`.

Go behavior:
- `goivy/isolate_create.go:249-279` creates the external call and empty stub.
- It does not copy `spec`/`impl`/`private` attributes to the wrapper name.
- It does not copy the original action line number onto the call or stub.
- It does not append a replacement `ImportDef` for the `imp__...` wrapper.
- It has no corresponding `extraWith = append(extraWith, impname)` update before later isolate processing.
- `goivy/isolate_create.go:280` replaces `mod.Imports` with only the imports that survived the earlier filter.

Confirmed mismatch:
- Python preserves wrapper metadata and threads the wrapper back into imports/present-action handling.
- Go installs the wrapper actions but drops the surrounding import and isolate-membership bookkeeping.

Why this is not just UI drift:
- Import wrapper metadata affects later isolate checks and action visibility. Dropping it can change which calls are considered present/imported and can remove source locations from isolate diagnostics.
