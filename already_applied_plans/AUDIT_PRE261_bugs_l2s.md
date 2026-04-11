# Audit Report: Python `ivy_l2s.py` vs Go `l2s/l2s.go` + `l2s/l2s_auto.go` + `l2s/shared.go`

## Files audited

- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_l2s.py` (1527 lines) — source of truth
- `/Users/jaten/ivy/goivy/l2s/l2s.go` (815 lines) — Go port (outer + helpers)
- `/Users/jaten/ivy/goivy/l2s/l2s_auto.go` (662 lines) — auto invariants
- `/Users/jaten/ivy/goivy/l2s/shared.go` (727 lines) — shared pipeline steps
- `/Users/jaten/ivy/goivy/check/isolate_check.go` — tactic dispatch context
- `/Users/jaten/ivy/goivy/check/l2s_hooks.go` — trace hook functions (orphaned)
- `/Users/jaten/ivy/goivy/ast/decl_ast.go` — `LabeledFormula` definition
- `/Users/jaten/ivy/goivy/ast/tactic.go` — `TacticTactic` struct
- `/Users/jaten/ivy/goivy/module/astutil.go` — `VariablesAST`

---

## 1. Cross-reference of `tactic_name` checks

| # | Py line | Python condition | Go file:line | Go condition | Match? |
|---|---------|------------------|--------------|--------------|--------|
| 1 | 200 | `tactic_name.startswith("l2s_auto")` | l2s.go:300 | `strings.HasPrefix(tacticName, "l2s_auto")` | MATCH |
| 2 | 232 | `tactic_name in ["l2s_auto5"]` (call `get_aux_defn('work_helpful',…)`) | l2s_auto.go:99 | `tacticName == "l2s_auto5"` | MATCH |
| 3 | 240 | `tactic_name not in ["l2s_auto3","l2s_auto4","l2s_auto5"]` inside `get_work_was_done` | l2s_auto.go:364 (inlined for l2s_work_preserved only) | `tacticName != "l2s_auto4" && tacticName != "l2s_auto5"` | **PARTIAL** — Python's `get_work_was_done` is also used by `not_all_was_done` (line 361-363) and `l2s_needed_were_frozen` (line 422); Go has no equivalent call sites |
| 4 | 281 | `tactic_name in ["l2s_auto5"]` (infer work_done from work_needed) | l2s_auto.go:134 | `tacticName == "l2s_auto5"` | MATCH |
| 5 | 354 | `tactic_name in ["l2s_auto2","l2s_auto3","l2s_auto4","l2s_auto5"]` — wrap `not_all_done` with `Or(Not(notWaitingForStart), tmp)` | l2s_auto.go:417-429 (no wrap) | MISSING | **CRITICAL MISMATCH** (user-reported bug #3) |
| 6 | 397 | `tactic_name not in ["l2s_auto5"]` for `l2s_needed_when_start` branching | l2s_auto.go:300 | `tacticName != "l2s_auto5"` | MATCH |
| 7 | 413 | `tactic_name not in ["l2s_auto4","l2s_auto5"]` for `l2s_needed_are_frozen` shape | l2s_auto.go:313 | `tacticName != "l2s_auto4" && tacticName != "l2s_auto5"` | MATCH |
| 7b | 422-425 | (inside else of 413) emits `l2s_needed_were_frozen<sfx>` invariant using `get_was_done` | l2s_auto.go | MISSING | **CRITICAL MISMATCH** |
| 8 | 429 | `tactic_name not in ["l2s_auto3","l2s_auto4","l2s_auto5"]` for `l2s_done_implies_created` | l2s_auto.go:343 | `tacticName != "l2s_auto3" && tacticName != "l2s_auto4" && tacticName != "l2s_auto5"` | MATCH |
| 9 | 436 | `tactic_name not in ["l2s_auto5"]` for `l2s_needed_implies_created` | l2s_auto.go:351 | `tacticName != "l2s_auto5"` | MATCH |
| 10 | 451 | `tactic_name not in ["l2s_auto4","l2s_auto5"]` for was_done/is_done simple form | l2s_auto.go:364 | matching | MATCH |
| 11 | 458 | `tactic_name in ["l2s_auto5"]` wrap `l2s_work_preserved` with `Implies(eventually_start(),...)` | l2s_auto.go:379-381 | `tacticName == "l2s_auto5"` | MATCH |
| 12 | 475 | `tactic_name not in ["l2s_auto5"] and tuple(progress_args) != tuple(done_args[:len(progress_args)]) → raise error` | l2s_auto.go | MISSING | **HIGH MISMATCH** — no prefix validation |
| 13 | 478 | `tactic_name != "l2s_auto3"` gate for `l2s_progress_made` entire emission | l2s_auto.go:388 | matching | MATCH (gate only) |
| 14 | 479 | `tactic_name in ["l2s_auto5"]` — `get_depends()`-based progress_made shape | l2s_auto.go | MISSING | **CRITICAL MISMATCH** |
| 15 | 491 | `elif progress_args or len(tasks) > 1` — uses `not_all_was_done`, `not_all_was_done_preds`, `next_task_not_triggered`, `eventually_start` | l2s_auto.go:390-404 (uses only innerCond with l2sSaved + !waitingForProgress; no eventually_start, no nad, no not_all_was_done, no next task logic) | **CRITICAL MISMATCH** — different body |
| 16 | 502-504 | `else` final branch of progress_made — single-task case; still uses `l2s_saved` and `Not(waiting_for_progress)` with `exists(done_args, ...)` | l2s_auto.go:399-403 | matches structurally | MATCH (for this branch only) |
| 17 | 532 | `tactic_name in ["l2s_auto5"]` for `l2s_sched_stable<sfx>` invariant | l2s_auto.go | MISSING | **CRITICAL MISMATCH** |
| 18 | 548 | `tactic_name in ["l2s_auto5"]` for final `l2s_sched_exists` invariant | l2s_auto.go | MISSING | **CRITICAL MISMATCH** |
| 19 | 1310 | `tactic_name.startswith("l2s_auto5")` attaches `auto_hook` vs `renaming_hook` to `goal.trace_hook` | l2s.go | MISSING | **HIGH MISMATCH** (user-reported bug #5) |

Also, the prefix `l2s_auto5` check at line 1310 uses `startswith`, whereas earlier Python conditions use `in ["l2s_auto5"]`. The Go port should also distinguish between these.

---

## 2. Closure/lambda cross-reference

| Python closure | Location (ivy_l2s.py) | Go equivalent | Status |
|---------------|----------------------|---------------|--------|
| `forall(vs, body)` | line 69 | `forall` (l2s.go:123, l2s_auto.go:162) | MATCH |
| `exists(vs, body)` | line 72 | `exists` (l2s.go:130, l2s_auto.go:168) | MATCH |
| `l2s_d(sort)` | line 182 | `L2SD` (l2s.go:55) | MATCH |
| `l2s_a(sort)` | line 183 | `L2SA` (l2s.go:60) | MATCH |
| `l2s_w(vs, t)` | line 184 | `l2sW` (l2s.go:65) | MATCH (but env type differs — see bug list) |
| `l2s_s(vs, t)` | line 185 | `l2sS` (l2s.go:70) | MATCH |
| `l2s_g(vs, t, environ)` | line 186 | `l2sG` (l2s.go:75) | MATCH |
| `old_l2s_g(vs, t, environ)` | line 187 | `oldL2sG` (l2s.go:80) | MATCH |
| `l2s_init(vs, t)` | line 188 | `l2sInit` (l2s.go:85) | MATCH |
| `l2s_when(name, vs, t)` | line 189 | `l2sWhen` (l2s.go:90) | MATCH |
| `l2s_old(vs, t)` | line 190 | `l2sOld` (l2s.go:95) | MATCH |
| `dict_put(dct, sfx, name, dfn)` | line 202 | `dictPut` (l2s_auto.go:41) | MATCH |
| `get_aux_defn(name, dct)` | line 207 | `getAuxDefn` (l2s_auto.go:50) | **MISMATCH** — see bugs list (no `IsDefinition` filter; silent skip instead of error; strips all universals not just one; doesn't rebuild LHS with generic name) |
| `get_work_was_done(defn, work_done)` | line 236 | inlined only at l2s_auto.go:368-377 | **PARTIAL** — only auto4/5 inline; non-auto4/5 form not usable elsewhere |
| `dependencies(syms)` | line 168 | `BuildDependenciesFunc` (shared.go:696) | **MISMATCH** — Go only uses `mod.Definitions`, not goal `prem_defns`; uses map order (non-deterministic) |
| `all_d(defn)` | line 327 | `allD` (l2s_auto.go:208) | **MINOR MISMATCH** — Go short-circuits to `true` when `cons` empty; Python always wraps in Implies |
| `all_a(defn)` | line 333 | `allA` (l2s_auto.go:228) | **MINOR MISMATCH** — same short-circuit difference |
| `all_created(defn)` | line 337 | `allCreated` (l2s_auto.go:248) | MATCH |
| `get_is_done(defn)` | line 341 | inlined at l2s_auto.go:321-323 (only for l2s_needed_are_frozen) | **PARTIAL** — not reusable at l2s_work_preserved call site (which rebuilds inline at l2s_auto.go:371-372) |
| `not_all_done(defn, skip=0)` | line 345 | inlined at l2s_auto.go:418-429 | **CRITICAL MISMATCH** — missing `work_end`/`work_invar` wrappers are present but `skip` parameter missing, and missing `Or(Not(notWaitingForStart), tmp)` wrap for l2s_auto2+ |
| `get_was_done(defn)` | line 358 | MISSING as a reusable closure | Used at l2s_work_preserved inline (l2s_auto.go:368-377); missing at `not_all_was_done` and `l2s_needed_were_frozen` sites |
| `not_all_was_done(defn, skip=0)` | line 361 | MISSING | **CRITICAL MISMATCH** — not implemented at all |
| `get_depends()` | line 365 | MISSING | **CRITICAL MISMATCH** — not implemented |
| `eventually_start_task(work_start)` | line 380 | `eventuallyStartTask` (l2s_auto.go:258) | **MEDIUM MISMATCH** — `Eventually.Environ` not set; Python sets to `proof_label` |
| `eventually_start()` | line 388 | inlined via `eventuallyStartTask(workStart)` | OK (different name but equivalent) |
| `next_task_has_trigger()` | line 462 | MISSING | **CRITICAL MISMATCH** |
| `next_task_not_triggered()` | line 465 | MISSING | **CRITICAL MISMATCH** |
| `init_globally(prop, res, pos)` | line 552 | `initGlobally` (l2s_auto.go:441) | **CRITICAL MISMATCH** — missing `arg is Globally or Not(Eventually)` extra invariant for pos=True Eventually; missing `arg is Eventually or Not(Globally)` extra invariant for pos=False Globally |
| `add_ini_invar(cond, fmla)` | line 642 | inlined in `convertToInit` default case (l2s_auto.go:581-589) | MATCH |
| `convert_to_init(fmla)` | line 647 | `convertToInit` (l2s_auto.go:552) | MATCH |
| `desugar(expr)` → `apply_was`, `apply_happened` | line 695 | `Desugar` (l2s.go:744), `applyWasRec` (l2s.go:778) | **CRITICAL MISMATCH** — defined but NEVER CALLED in `l2sTacticInt` (Python calls at line 716: `invars = list(map(desugar,invars))`) |
| `mod_pass(transform)` | line 737 | inline closure in l2s.go:326 | **CRITICAL MISMATCH** — Go's mutates local `invars` separately from `model.Invars` (see invariant flow bug) |
| `list_transform(lst, trns)` | line 726, 812 | MISSING | prems are never transformed |
| `_l2s_g(vs, t, env)` | line 753 | closure inside `SharedStep1_ConvertTemporals` (shared.go:164) | MATCH |
| `_l2s_when(name, vs, t)` | line 759 | closure inside `SharedStep1_ConvertTemporals` (shared.go:170) | MATCH |
| `apply_l2s_init(vs, t)` | line 984 | `applyL2sInit` (l2s.go:689) | MATCH |
| `prop_events(gprops)` | line 1027 | inline in `SharedStep7_InstrumentActions` (shared.go:403) | MATCH |
| `when_events(whens)` | line 1044 | inline in `SharedStep7_InstrumentActions` (shared.go:444) | **MINOR MISMATCH** — Python `when_events` guards with `if name == 'l2s_whenprev': cond = ...` using `t.t1`; Go uses `when.Body` directly as `cond` |
| `wait_events(waits)` | line 1072 | inline in `SharedStep7_InstrumentActions` (shared.go:476) | MATCH |
| `instr_stmt(stmt, labels)` | line 1122 | inline in `SharedStep7_InstrumentActions` (shared.go:494) | **HIGH MISMATCH** — Go signature drops `labels`; Python checks for CallAction with monitored returns and calls `split_returns` — Go has no such handling |

---

## 3. Invariant emission cross-reference

| Py line | Python invariant name | Gate | Go file:line | Match? |
|---------|----------------------|------|--------------|--------|
| 401 | `l2s_needed_when_start<sfx>` | always, but shape depends on auto5 | l2s_auto.go:302, 305 | MATCH (both shapes) |
| 409 | `l2s_created<sfx>` | always | l2s_auto.go:309 | MATCH |
| 415 | `l2s_needed_are_frozen<sfx>` | non-auto4/5 | l2s_auto.go:318 | MATCH |
| 421 | `l2s_needed_are_frozen<sfx>` | auto4/5 | l2s_auto.go:339 | MATCH |
| 425 | `l2s_needed_were_frozen<sfx>` | auto4/5 | **MISSING** | **CRITICAL** |
| 432 | `l2s_done_implies_created<sfx>` | non-auto3/4/5 | l2s_auto.go:347 | MATCH |
| 439 | `l2s_needed_implies_created<sfx>` | non-auto5 | l2s_auto.go:359 | MATCH |
| 460 | `l2s_work_preserved<sfx>` | always | l2s_auto.go:382 | MATCH |
| 505 | `l2s_progress_made<sfx>` | non-auto3 | l2s_auto.go:405 | **CRITICAL MISMATCH** — formula shape wrong: missing `eventually_start()`, `not_all_was_done`, `not_all_was_done_preds`, `next_task_has_trigger`, `next_task_not_triggered`, `get_depends`, and `done_args[len(progress_args):]` slicing |
| 510 | `l2s_progress_invar<sfx>` | always | l2s_auto.go:411 | MATCH |
| 525 | `l2s_not_all_done<sfx>` (per scheduler boundary) | if next_task has trigger | **MISSING** | **CRITICAL** |
| 538 | `l2s_sched_stable<sfx>` | auto5 | **MISSING** | **CRITICAL** |
| 546 | `l2s_not_all_done` (final) | always | l2s_auto.go:432 | **CRITICAL MISMATCH** — Go emits only the LAST task's `notAllDone`, not `Or(*not_all_done_preds)` |
| 550 | `l2s_sched_exists` | auto5 | **MISSING** | **CRITICAL** |
| 620 | `l2s_globally_<i>` | always | l2s_auto.go:547 | **MISMATCH** — body differs because `initGlobally` is missing the EF/AG extra invariants |
| 623 | `l2s_status_0` | always | l2s_auto.go:601 | MATCH |
| 626 | `l2s_status_1` | always | l2s_auto.go:603 | MATCH |
| 629 | `l2s_status_2` | always | l2s_auto.go:605 | MATCH |
| 632 | `l2s_status_3` | always | l2s_auto.go:607 | MATCH |
| 638 | `l2s_consts_d` | always | l2s_auto.go:631 | **MINOR MISMATCH** — Go emits only if `len(constsDTerms) > 0`; Python always emits (with `And()` = true when empty) |
| 660 | `l2s_init_glob_<i>` | always | l2s_auto.go:596 | MATCH (body shape) |
| 662 | `neg_prop_init` | always | l2s_auto.go:598 | MATCH |
| 676 | `l2s_when_<i>` | always (but set empty unless WhenOperator 'first' exists) | **MISSING** | **CRITICAL** |

**Order mismatch**: Python emits in order `{globally, status, consts_d, init_glob, neg_prop, when}`. Go emits `{globally, init_glob, neg_prop, status, consts_d}`. This shifts the numeric indices and changes the invariant serialization order.

---

## 4. Complete bug list with severity

### CRITICAL (wrong proof outcome)

**C1. L2STactic does not wrap in WithSymbols/WithSorts (user bug #1)**
- Python `ivy_l2s.py:75-79`: `l2s_tactic` enters `WithSymbols(vocab.symbols)` then `WithSorts(vocab.sorts)` BEFORE calling `l2s_tactic_int`.
- Go `l2s/l2s.go:189-201`: `L2STactic`, `L2STacticFull`, `L2STacticAuto` all call `l2sTacticInt` directly, with no wrapping.
- The `WithSymbols`/`WithSorts` wrapping that does exist in `check/isolate_check.go:779-784` happens in `CheckSubgoals`, which runs AFTER the tactic has already executed.
- Fix: add a wrapper helper `l2sTacticWrap(pc, goals, pf, tacticName)` that calls `il.NewWithSymbols(…).Enter()` / `il.NewWithSorts(…).Enter()` around `l2sTacticInt` and then `Exit`s them.

**C2. L2STacticAuto hardcodes "l2s_auto" (user bug #2)**
- Python `ivy_l2s.py:91-93`: `l2s_tactic_auto` reads `proof.tactic_name` (user's actual name).
- Go `l2s/l2s.go:199-201`: passes hardcoded `"l2s_auto"`.
- Fix: `L2STacticAuto` must extract `tn := nodeToString(pf.(*ast.TacticTactic).TName)` and pass that into `l2sTacticInt`.

**C3. `not_all_done` missing `Or(Not(notWaitingForStart), tmp)` for auto2/3/4/5 (user bug #3)**
- Python `ivy_l2s.py:354-355`: wraps the final `tmp` with `Or(Not(not_waiting_for_start), tmp)` for `l2s_auto2`/`l2s_auto3`/`l2s_auto4`/`l2s_auto5`.
- Go `l2s/l2s_auto.go:417-429`: builds `notAllDoneBody` inline without this wrap.
- Fix: after computing `notAllDone`, if `tacticName in {l2s_auto2,…,5}`, wrap with `Or{Not{notWaitingForStart}, notAllDone}`.

**C4. `l2s_not_all_done` single-task-only (user bug #4)**
- Python `ivy_l2s.py:519-546`: maintains `not_all_done_preds` list, appends per task, emits intermediate `l2s_not_all_done<sfx>` invariants at scheduler boundaries, and finally emits the global `l2s_not_all_done` as the OR of all accumulated preds.
- Go `l2s/l2s_auto.go:431-433`: only emits `l2s_not_all_done` for the LAST task, with only that task's `notAllDone` formula. No boundary invariants. No final OR.
- Fix: introduce `notAllDonePreds []lg.Expr` outside the task loop; append each task's `notAllDone`; at each task, if next task has a trigger, emit `l2s_not_all_done<sfx>` invariant and reset the list; after loop, emit `l2s_not_all_done` = `Or(notAllDonePreds...)`.

**C5. Goal trace_hook infrastructure missing (user bug #5)**
- Python: `LabeledFormula` supports dynamic `goal.trace_hook` attribute; `l2s_tactic_full` sets `trace_hook`; `l2s_tactic_int` sets `goal.trace_hook = lambda: auto_hook(…)` or `renaming_hook`. `ivy_check.py` reads `goal.trace_hook` and wires it to `mod.trace_hook`, then `check.py` invokes `mod.trace_hook(handler, ffcs)` on trace display.
- Go: `check/l2s_hooks.go` defines `L2STraceHook`, `L2SRenamingHook`, `L2SAutoHook`, and `L2SAutoHookConfig`. But `LabeledFormula` (ast/decl_ast.go:25-39) has no `TraceHook` field. `check/isolate_check.go:804-805` and `:857-858` have TODO comments saying "TraceHook not yet a field". `check/check.go:589-590` says "trace_hook is set by l2s for temporal property diagnostics" but the wiring is absent.
- Fix: add `TraceHook func(*trace.TraceBase, []Checker) *trace.TraceBase` field to `ast.LabeledFormula`; in `l2sTacticInt`, set it on the returned goal based on `tacticName`; also populate `L2SAutoHookConfig` with tasks/triggers/subs maps; in `check/isolate_check.go`, replace the TODO comments with `if goal.TraceHook != nil { mod.TraceHook = goal.TraceHook }`; in `check/check.go:589`, use `mod.TraceHook` instead of the stub.

**C6. User tactic invariants (`proof.tactic_decls`) are completely ignored**
- Python `ivy_l2s.py:141-175`: extracts `tactic_invars = [inv for inv in proof.tactic_decls if not isinstance(inv,DerivedDecl)]`, compiles each with `compile_with_goal_vocab(inv,goal)`, wraps with `label_temporal(…, proof_label)`, and uses this list as the initial `invars` (before auto-generation).
- Go `l2s/l2s.go:272-274`: starts `invars` as a copy of `model.Invars` ONLY. No use of `pf.(*ast.TacticTactic).TacticDeclsList()`. User-supplied `with invariant X` clauses are silently dropped.
- Fix: at start of `l2sTacticInt`, cast `pf` to `*ast.TacticTactic`; get `.TacticDeclsList()`; filter out `DerivedDecl` items (those go to `tactic_defns`); compile each remaining item via `proof.CompileWithGoalVocab` (or add such a helper) to get `lg.Expr`; wrap each compiled expression with a `label_temporal` equivalent; prepend to `invars`.

**C7. User tactic definitions (`proof.tactic_decls` of DerivedDecl type) are completely ignored**
- Python `ivy_l2s.py:141-153`: extracts `tactic_defns`; calls `compile_definition_goal_vocab(defn, goal)` to add them as premises to the goal.
- Go `l2s/l2s.go`: not called anywhere in l2sTacticInt. `proof.CompileDefinitionGoalVocab` exists but is orphaned.
- Fix: before the main flow, iterate tactic_defns and mutate goal via `CompileDefinitionGoalVocab`.

**C8. Auto-generated invariants never merged into `model.Invars`**
- Python `ivy_l2s.py:722`: `model.invars = model.invars + invars` — commits the tactic-level `invars` into the model before downstream processing.
- Go `l2s/l2s.go`: there is NO such assignment. `invars` is a local slice that receives auto-generated invariants but is never written back to `model.Invars`. 
- Consequences:
  - `SharedStep3_CollectNamedBinders` (shared.go:211-214) reads ONLY `model.Invars` — so l2s_s / l2s_w binders inside auto-generated invariants are NOT collected into `to_save`/`to_wait`, so corresponding reset/save actions are NOT generated.
  - `SharedStep11_ReplaceNamedBinders` (shared.go:611) via `collectAllNamedBinders` only walks `model.Invars` — so substitutions don't cover auto-generated invariants' binders.
  - `SharedStep12_BuildGoal` builds the new goal from `tm.Model` (which has only the original `model.Invars`) — so the returned goal's model is missing all auto-generated invariants entirely.
- Fix: after `l2sAutoInvariants` returns `invars`, do `model.Invars = invars` (not `append` — because `invars` already includes the old model.Invars copied into it at l2s.go:274). And make sure `modPass` transforms only `model.Invars` (drop the redundant transform over local `invars`).

**C9. `Desugar` never called on invariants**
- Python `ivy_l2s.py:716`: `invars = list(map(desugar,invars))` — replaces `$was`/`$happened` inside user/auto invariants with `l2s_saved & l2s_s(…)(…)` / `l2s_saved & Not(l2s_w(…)(…))`.
- Go `l2s/l2s.go`: `Desugar` defined at line 744, never called. User-supplied `$was(p(X))` operators remain as `NamedBinder{name="was",…}` in invariants, which downstream code will not understand.
- Fix: after auto-invariant generation, iterate `invars` and apply `Desugar(inv.Formula.(lg.Expr), proofLabel)` to each.

**C10. Temporal premises from `prover.axioms` are not added to `fmla` wrap**
- Python `ivy_l2s.py:134-137`: builds `temporal_prems` from BOTH `ipr.goal_prems(goal)` AND `prover.axioms`:
  ```python
  temporal_prems = [x for x in ipr.goal_prems(goal) if hasattr(x,'temporal') and x.temporal] + [
      x for x in prover.axioms if not x.explicit and x.temporal]
  ```
- Go `l2s/l2s.go:239-249`: only collects from `proof.GoalPrems(goal)`, completely missing the axiom-sourced temporal premises for the `Implies(And(prems...), fmla)` wrap.
- Fix: in the loop at l2s.go:239, also iterate `pc.GetAxioms()` and add any non-explicit temporal axioms (`ax.Explicit == false && ax.IsTemporal()`) to `temporalPrems`.

**C11. `l2s_progress_made` formula shape drastically wrong (non-auto3 branch)**
- Python `ivy_l2s.py:478-505`:
  - For `l2s_auto5`: uses `get_depends()`, wraps in `l2s_s(progress_args, nad)`, optionally conjoins `next_task_not_triggered()`, builds `And(l2s_saved, eventually_start(), exists(progress_args, nad), not_all_was_done(work_needed), forall(progress_args, Implies(nad, Not(waiting_for_progress(...)))))`.
  - For non-auto5 with progress_args OR multiple tasks: builds `nad = And(not_all_was_done(work_needed, len(progress_args)), Not(Or(*not_all_was_done_preds)))`, optionally wraps with `next_task_not_triggered()`, builds `forall(progress_args, Implies(And(nad, l2s_saved, eventually_start(), Not(waiting_for_progress(*progress_args))), exists(done_args[len(progress_args):], And(Not(was_done), is_done))))`.
  - For non-auto5 simple case: `Implies(And(l2s_saved, Not(waiting_for_progress)), exists(done_args, And(Not(was_done), is_done)))`.
- Go `l2s/l2s_auto.go:388-406`: collapses to a SIMPLER shape — just `makeAnd(l2sSaved, Not(waitingForProgress))` → `exists(doneArgs, And(Not(wasDone), isDoneNode))`. Missing `eventually_start`, `not_all_was_done`, `not_all_was_done_preds`, `next_task_not_triggered`, `get_depends`, scoped existential `doneArgs[len(progressArgs):]`.
- Fix: re-implement the Python branching structure faithfully. This requires adding missing closures `notAllWasDone`, `getDepends`, `nextTaskHasTrigger`, `nextTaskNotTriggered` AND the `notAllWasDonePreds` accumulator logic.

**C12. `l2s_needed_were_frozen<sfx>` invariant completely missing**
- Python `ivy_l2s.py:422-425`: emitted in the `l2s_auto4`/`l2s_auto5` branch; uses `get_was_done(work_needed)`.
- Go `l2s/l2s_auto.go`: absent.
- Fix: inside the `else` of `tacticName != "l2s_auto4" && tacticName != "l2s_auto5"` at l2s_auto.go:313, emit this invariant after `l2s_needed_are_frozen`.

**C13. `l2s_sched_stable<sfx>` invariant completely missing (auto5)**
- Python `ivy_l2s.py:532-538`: uses `get_depends`, `l2s_s(progress_args, nad)(*progress_args)`, `forall(progress_args, Implies(And(was_nad, l2s_saved, eventually_start(), waiting_for_progress(...)), nad))`.
- Go: absent.
- Fix: add this block in the task loop when `tacticName == "l2s_auto5"`.

**C14. `l2s_sched_exists` invariant completely missing (auto5)**
- Python `ivy_l2s.py:542-543, 548-550`: accumulates `sched_exists_preds.append(exists(progress_args, was_nad))` per task; after the loop, emits `Implies(And(l2s_saved, eventually_start()), Or(*sched_exists_preds))`.
- Go: absent.
- Fix: add `schedExistsPreds []lg.Expr`; populate inside `l2s_sched_stable` block; after loop, if `tacticName == "l2s_auto5"`, emit this invariant.

**C15. `l2s_when_<i>` invariants completely missing**
- Python `ivy_l2s.py:666-676`: iterates `ilu.temporals_asts(invars+prems)`; for each `WhenOperator` with `name=='first'`, builds `Implies(l2s_init((), Eventually(t2)), Implies(Not(nws), Eq(tmprl, WhenOperator('next', t1, t2))))` where `nws = Or(Not(l2s_waiting), Not(l2s_w((),t2)))`.
- Go: absent entirely.
- Fix: add this block after `neg_prop_init` emission.

**C16. `init_globally` missing EF/AG extra invariants**
- Python `ivy_l2s.py:562-579`: 
  - `pos=True, prop=Eventually`: after appending the base Implies, if `arg is Globally` OR `arg is Not(Eventually)`, append `Implies(Or(Not(l2s_waiting), Not(l2s_w(vs, arg)(*vs))), arg)`.
  - `pos=False, prop=Globally`: after appending the base Implies, if `arg is Eventually` OR `arg is Not(Globally)`, append `Implies(Or(Not(l2s_waiting), Not(l2s_w(vs, Not(arg))(*vs))), Not(arg))`.
- Go `l2s/l2s_auto.go:441-506`: both extra invariants are OMITTED.
- Fix: add these two extra `*res = append(...)` statements in the corresponding branches.

**C17. Sort-theory `is_finite()` check missing**
- Python `ivy_l2s.py:194`: `if thy.get_sort_theory(sort).is_finite() or name in mod.finite_sorts or full:` — adds theory-finite sorts (e.g. `bv[N]`) to `finite_sorts`.
- Go `l2s/l2s.go:286`: only checks `m.FiniteSorts[name] || full`.
- Fix: add `theory.GetSortTheory(s, m.Interp)` call and check if result has `IsFinite()` — but this may require adding an `IsFinite()` method to `theory.Theory`.

**C18. `getAuxDefn` does not filter by `IsDefinition` and silently skips free-var premises**
- Python `ivy_l2s.py:207-221`: iterates premises; filter `not isinstance(prem,ivy_ast.ConstantDecl) and hasattr(prem,"definition") and prem.definition`; raises `IvyError` if free variables are present.
- Go `l2s/l2s_auto.go:50-91`: iterates all premises; no `IsDefinition` filter; silently `continue`s if free variables present (no error).
- Fix: add `if !premLF.IsDefinition { continue }`; when `len(freeVars) > 0`, return `fmt.Errorf("free symbol %s not allowed in definition of %s", freeVars[0].Name, dname)` instead of `continue`.

---

### HIGH (subtly wrong but may not surface immediately)

**H1. `user's tactic_proof` block is ignored**
- Python `ivy_l2s.py:910-911`: if `proof.tactic_proof`, calls `ivy_compiler.apply_assert_proof(prover, assert_no_fair_cycle, proof.tactic_proof)`.
- Go `l2s/l2s.go`: no such handling. The `assert_no_fair_cycle` is unwrapped.
- Impact: users who write `tactic l2s proof { <sub-proof> }` have their sub-proof silently dropped. The SAT check for non-fair-cycle becomes unconditional instead of being guided by the sub-proof.

**H2. `to_g` is deduped via map-key with non-deterministic iteration**
- Python `ivy_l2s.py:965-967`: `to_g = list(set(to_g))` — set iteration non-deterministic in Python too, but the set contents match exactly.
- Go `shared.go:327-333`: builds from map iteration and sorts by `fmt.Sprint(Body)`.
- Impact: tie-breaking across entries with same body string but different vars/env is non-deterministic, producing different `l2s_globally_<i>` labels across runs.

**H3. `work_progress` prefix-of-`work_done` check missing**
- Python `ivy_l2s.py:475-476`: raises `IvyError` if `progress_args` is not a prefix of `done_args[:len(progress_args)]`.
- Go: absent.
- Impact: silently generates incorrect invariants that reference variables not in scope.

**H4. `work_invar`-has-no-arguments check missing**
- Python `ivy_l2s.py:320-321`: raises `IvyError` if `work_invar.args[0].args` is non-empty.
- Go: absent.

**H5. Sort-signature consistency checks missing**
- Python `ivy_l2s.py:314-319`: `work_created`/`work_needed`/`work_done` must share sort; `work_helpful`/`work_progress` must share sort.
- Go: absent.
- Impact: mismatched sorts → downstream `subst` produces ill-typed terms.

**H6. Variable ordering in `collectVarsSlice` is non-deterministic**
- Python `ivy_l2s.py:650`: `vs = tuple(iu.unique(ilu.variables_ast(fmla)))` — preserves traversal-order.
- Go `l2s/l2s_auto.go:644-647`: calls `modpkg.VariablesAST(n)` which iterates a map → non-deterministic order.
- Impact: different runs produce different binder variable orderings in `l2s_init(vs, f)(*vs)` → different invariant hashes → flaky tests and non-reproducible traces.

**H7. `CallAction` with monitored returns not split**
- Python `ivy_l2s.py:1127-1131`: if a CallAction's returns include any monitored symbol, `split_returns()` is called before instrumentation.
- Go `shared.go:494`: no such handling.
- Impact: if a user's action calls another action that writes to a monitored relation, the event is missed during instrumentation.

**H8. `cfg.Invars` field set but never read**
- `shared.go:30`: `Invars: []*ast.LabeledFormula`. Set at `l2s.go:319` but never consumed.
- Dead code that suggests the intent was there but the integration was never finished. Delete or wire it up.

**H9. `getAuxDefn` uses `DropUniversals` (all levels) instead of single `ForAll` strip**
- Python `ivy_l2s.py:210-212`: strips only one outer `ForAll`.
- Go `l2s/l2s_auto.go:65`: strips all.
- Impact: if a user's definition has nested `Forall`-of-`Forall` (rare but possible), Go would drop both sets of bindings and produce ill-formed results.

**H10. `getAuxDefn` stores the original LHS name, not the generic name**
- Python `ivy_l2s.py:218-219`: rebuilds `lhs = Const(name, …)(*tmp.args[0].args)` — renames the LHS from e.g. `"work_created0"` to `"work_created"` so downstream code sees the generic name.
- Go `l2s/l2s_auto.go:89`: stores the original `eq` with its declared name (e.g. `"work_created0"`).
- Impact: any downstream code comparing LHS names to `"work_created"` would fail in Go.

**H11. `eventually_start_task` Eventually constructed without `Environ`**
- Python `ivy_l2s.py:384`: `lg.Eventually(proof_label, trigf.args[1])` sets environ to `""`.
- Go `l2s/l2s_auto.go:263`: `&lg.Eventually{Body: trigRHS}` leaves `Environ` as `nil`.
- Impact: `l2sG` and `l2sW` created from this Eventually will have mismatched environ fields vs. those created through `SharedStep1` (which uses `strPtr(proofLabel)`), breaking dedup.
- Also at l2s_auto.go:409 (`l2s_progress_invar` body) and l2s_auto.go:521 (trigger-based init_globally).

**H12. Defn-deps include `mod.Definitions` only, not `prem_defns`**
- Python `ivy_l2s.py:163`: `for defn in list(prover.definitions.values()) + prem_defns`.
- Go `shared.go:677`: only uses `mod.Definitions`.
- Impact: if a user defines a function in the goal prems (e.g. `with property foo(X) = ...`), dependencies through that definition are not tracked, so actions that modify the dependency's inputs will not trigger property events for operators depending on `foo`.

**H13. Label on assumed gprops is dropped**
- Python `ivy_l2s.py:132`: `p.clone([p.label, p.formula.args[0]])` — preserves the original label on the cloned body.
- Go `l2s/l2s.go:257`: `m.Cfg.AstCfg.NewLabeledFormula(nil, g.Body)` — label is nil.
- Impact: error messages about violations of these assumed properties lose their label.

