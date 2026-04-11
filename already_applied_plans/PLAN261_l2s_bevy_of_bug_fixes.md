# Plan: Faithfully port `l2s_tactic` (and `l2s_auto*`) from Python — fix the divergence at log.red i=254990 plus all latent l2s bugs uncovered

Created: 2026-04-11 06:30 UTC

## Context

`make golden` advanced past the previous `*ast.TemporalModels` divergence and now diverges at i=254990:

```
254989  proof.tacticTactic name='l2s_auto2' goal[0].Formula type=SchemaBody       (BOTH match)
254990  go : l2s.l2sTacticInt ENTER tactic="l2s_auto" ngoals=1 goal.Formula type=SchemaBody
        py : ivylogic.WithSorts.Enter nSorts=0
```

The proximate cause is twofold:

1. **Go's `L2STactic`/`L2STacticFull`/`L2STacticAuto` skip the WithSymbols/WithSorts wrapping** that Python's `l2s_tactic` does (`ivy_l2s.py:75-79`). The missing `WithSorts.Enter` xtrace is what visibly diverges.

2. **Go's `L2STacticAuto` hardcodes the literal string `"l2s_auto"`** for the inner `tacticName`, no matter whether the user wrote `l2s_auto`, `l2s_auto2`, `l2s_auto3`, `l2s_auto4`, or `l2s_auto5` (`l2s/l2s.go:199-201`). Python passes `proof.tactic_name` (the user's actual name) through (`ivy_l2s.py:91-93`).

But investigating the gap revealed that **the entire `l2sTacticInt` flow is structurally divergent from Python's `l2s_tactic_int`**: an audit found 18 CRITICAL bugs, 13 HIGH bugs, plus several MEDIUM and TRACE-ONLY ones in `l2s/l2s.go`, `l2s/l2s_auto.go`, and `l2s/shared.go`. The most severe is C8: **the auto-generated invariants (the result of `l2sAutoInvariants`) are accumulated in a local `invars` slice that is never written back to `model.Invars`, so the entire downstream pipeline (named-binder collection, action instrumentation, goal building) silently ignores them, producing vacuous proofs**.

The user has explicitly asked to fix all uncovered bugs in this plan, not defer them. The primary goal is correctness (Python parity), with the trace divergence acting as a witness.

## Investigation findings

### Confirmed structural facts (verified in code, not just audit claims)

1. **`L2STactic`/`L2STacticFull`/`L2STacticAuto`** at `l2s/l2s.go:189-201` are byte-thin wrappers that call `l2sTacticInt(pc, goals, pf, "l2s")` / `"l2s_full"` / **`"l2s_auto"`** respectively. None of them computes `vocab` or enters any `WithSymbols`/`WithSorts` context.

2. **`l2s/l2s_auto.go`** has 9 `tacticName`-conditional branches that test for `l2s_auto3`, `l2s_auto4`, or `l2s_auto5` (lines 99, 134, 300, 313, 343, 351, 364, 379, 388). Because `tacticName` is hardcoded to `"l2s_auto"` in `L2STacticAuto`, **none of these branches ever fire on the auto2/3/4/5 path**. Every `l2s_auto2`/`l2s_auto3`/`l2s_auto4`/`l2s_auto5` proof in goivy is silently using the plain `l2s_auto` invariant set.

3. **`cfg.Invars`** is set at `l2s/l2s.go:319` but **read nowhere** in the package (verified via `Grep "Invars" l2s/`). The fields exist (`shared.go:30`) but are dead.

4. **`model.Invars`** is what `SharedStep3_CollectNamedBinders` reads at `shared.go:212` and what `collectAllNamedBinders` walks at `l2s.go:644`. Anything not in `model.Invars` is invisible to the rest of the l2s pipeline.

5. **The local `invars` slice** in `l2sTacticInt` is initialized at `l2s/l2s.go:274` from `model.Invars` (a snapshot copy), then mutated by `l2sAutoInvariants` (line 302), then never written back. Python (`ivy_l2s.py:175, 722`) initializes `invars` from compiled `tactic_invars` and at the end does `model.invars = model.invars + invars`, committing the auto-generated invariants into the model.

6. **`Desugar`** is defined at `l2s/l2s.go:744` but **never called** in `l2sTacticInt` (verified via `Grep "Desugar" l2s/l2s.go`). Python (`ivy_l2s.py:716`) does `invars = list(map(desugar,invars))` before committing to `model.invars`.

7. **`il.NewWithSorts`** and **`il.NewWithSymbols`** already exist in `ivylogic/sig.go:385/423` and emit the correct `ivylogic.WithSorts.Enter nSorts=%d` xtrace at line 429. They are used elsewhere (`compiler/phase6.go`, `proof/phase5_matching.go`, `check/isolate_check.go`). The construct exists; it's just not invoked from `l2s/`.

8. **`proof.GoalVocab`** exists at `proof/goal.go:210` and is the Go port of Python's `goal_vocab`. It returns sorts/symbols/variables. It's already correctly handling `*ast.TemporalModels` (per the plan we just applied).

### Complete bug inventory (Python `ivy_l2s.py` vs Go `l2s/`)

#### CRITICAL — wrong proof outcome

| ID | Bug | Python (truth) | Go (broken) |
|---|---|---|---|
| **C1** | No WithSymbols/WithSorts wrapping around `l2sTacticInt` | `ivy_l2s.py:75-79` `l2s_tactic` enters both | `l2s/l2s.go:189-201` skip wrapping |
| **C2** | `L2STacticAuto` hardcodes `"l2s_auto"` | `ivy_l2s.py:91-93` passes `proof.tactic_name` | `l2s/l2s.go:200` literal `"l2s_auto"` |
| **C3** | `not_all_done` missing `Or(Not(notWaitingForStart), tmp)` for auto2/3/4/5 | `ivy_l2s.py:354-355` | `l2s/l2s_auto.go:417-432` no wrap |
| **C4** | `l2s_not_all_done` only emits LAST task, no per-boundary intermediate, no final OR | `ivy_l2s.py:519-546` builds `not_all_done_preds` and emits per-boundary + final OR | `l2s/l2s_auto.go:431-433` |
| **C5** | `LabeledFormula.TraceHook` field missing → no auto/renaming hook wiring | `ivy_l2s.py:88, 1311, 1313` set `goal.trace_hook` | `ast/decl_ast.go` no field; `check/l2s_hooks.go` orphaned; TODOs at `check/isolate_check.go:804, 857`, `check/check.go:589` |
| **C6** | User `tactic_invars` (non-DerivedDecl) silently dropped | `ivy_l2s.py:141, 175` | `l2s/l2s.go:272-274` starts `invars` from `model.Invars`, ignores `proof.tactic_decls` |
| **C7** | User `tactic_defns` (DerivedDecl) silently dropped | `ivy_l2s.py:142, 152-153` `compile_definition_goal_vocab` | `proof.CompileDefinitionGoalVocab` exists but is orphaned; not called from `l2s` |
| **C8** | **Auto-generated invariants never merged into `model.Invars`** | `ivy_l2s.py:722` `model.invars = model.invars + invars` | `l2s/l2s.go` no such assignment; `cfg.Invars` is dead |
| **C9** | `Desugar` never called on invariants | `ivy_l2s.py:716` `invars = list(map(desugar,invars))` | `l2s/l2s.go:744` `Desugar` defined but never called |
| **C10** | Temporal premises from `prover.axioms` not added to `fmla` wrap | `ivy_l2s.py:134-135` includes axioms | `l2s/l2s.go:239-249` only goal premises |
| **C11** | `l2s_progress_made` formula shape drastically wrong: missing `eventually_start`, `not_all_was_done`, `not_all_was_done_preds`, `next_task_not_triggered`, `get_depends`, scoped existential `done_args[len(progress_args):]` | `ivy_l2s.py:478-505` | `l2s/l2s_auto.go:388-406` collapsed shape |
| **C12** | `l2s_needed_were_frozen<sfx>` invariant missing | `ivy_l2s.py:422-425` (auto4/5 branch) | `l2s/l2s_auto.go` absent |
| **C13** | `l2s_sched_stable<sfx>` invariant missing (auto5) | `ivy_l2s.py:532-538` | absent |
| **C14** | `l2s_sched_exists` invariant missing (auto5) | `ivy_l2s.py:548-550` | absent |
| **C15** | `l2s_when_<i>` invariants missing (WhenOperator 'first') | `ivy_l2s.py:666-676` | absent |
| **C16** | `init_globally` missing EF/AG extra invariants | `ivy_l2s.py:562-579` (positive Eventually-of-Globally and negative Globally-of-Eventually) | `l2s/l2s_auto.go:441-506` omitted |
| **C17** | Sort theory `is_finite()` check missing → bv sorts treated as infinite | `ivy_l2s.py:194` | `l2s/l2s.go:286` only `mod.FiniteSorts` |
| **C18** | `getAuxDefn` doesn't filter by `IsDefinition`; silently skips free-var premises instead of erroring | `ivy_l2s.py:207-221` | `l2s/l2s_auto.go:50-91` |

#### HIGH — subtly wrong, may not surface in current tests

| ID | Bug | Locations |
|---|---|---|
| **H1** | `proof.tactic_proof` block ignored | `ivy_l2s.py:910-911` vs Go absent |
| **H2** | `to_g` dedupe via map iteration → non-deterministic indices for `l2s_globally_<i>` | `shared.go:327-333` |
| **H3** | `work_progress` prefix-of-`work_done` validation missing | `ivy_l2s.py:475-476` |
| **H4** | `work_invar` "no arguments" validation missing | `ivy_l2s.py:320-321` |
| **H5** | Sort-signature consistency checks (`work_created`/`work_needed`/`work_done` same sort; `work_helpful`/`work_progress` same sort) missing | `ivy_l2s.py:314-319` |
| **H6** | Variable ordering in `module.VariablesAST` is non-deterministic (map iteration) → flaky binder ordering | `module/astutil.go` |
| **H7** | `CallAction` with monitored returns not split via `split_returns()` | `ivy_l2s.py:1127-1131` vs `shared.go:494` |
| **H8** | `cfg.Invars` field set but never read — dead code | `shared.go:30, l2s.go:319` |
| **H9** | `getAuxDefn` uses `DropUniversals` (all levels) instead of single `ForAll` strip | `l2s/l2s_auto.go:65` |
| **H10** | `getAuxDefn` stores original LHS name instead of generic name (`work_created0` vs `work_created`) | `l2s/l2s_auto.go:89` |
| **H11** | `eventually_start_task` builds `Eventually` without `Environ`, breaking dedup with `SharedStep1`-built ones | `l2s/l2s_auto.go:263, 409, 521` |
| **H12** | Defn-deps include `mod.Definitions` only, not `prem_defns` from goal | `shared.go:677` |
| **H13** | Label on assumed gprops dropped (`m.Cfg.AstCfg.NewLabeledFormula(nil, ...)`) | `l2s/l2s.go:257` |

#### MEDIUM — structural divergence, no behavioral effect today

| ID | Bug | Locations |
|---|---|---|
| **M1** | `proof.tactic_lets` not rejected with `IvyError` | `ivy_l2s.py:124-125` vs Go silent |
| **M2** | `proof.RemoveUnusedDefinitionsGoal(goal)` not called before goal build | `ivy_l2s.py:1308` vs Go orphan |
| **M3** | Invariant emission order differs (Go: globally, init_glob, neg_prop, status, consts_d; Py: globally, status, consts_d, init_glob, neg_prop, when) | `l2s_auto.go` |
| **M4** | `allD`/`allA` short-circuit to `true` when cons empty (logically equivalent, structurally different) | `l2s_auto.go:221, 241` |
| **M5** | `l2s_consts_d` only emitted when non-empty (Python always emits) | `l2s_auto.go:630` |
| **M6** | Tactic print blocks missing (`--- l2s_auto invariants ---` etc.) | `ivy_l2s.py:679-682, 874-876` |
| **M7** | `mod_pass` does not transform goal-level `prems` | `ivy_l2s.py:744` `list_transform(prems,transform)` |

#### TRACE-ONLY

| ID | Bug | Locations |
|---|---|---|
| **T1** | `--- l2s_auto invariants ---` print block missing | `ivy_l2s.py:679-682` |
| **T2** | `reset_w:` print block missing | `ivy_l2s.py:874-876` |
| **T3** | Missing `ivylogic.WithSorts.Enter/Exit` xtrace events at l2s entry | (the immediate divergence; resolved by C1) |
| **T4** | Diagnostic detail in `auto_hook` shorter than Python | `check/l2s_hooks.go:100-161` |

## Approach

This is a single comprehensive fix. The plan is organized in **dependency phases** so each phase compiles and stays correct.

### Phase 1 — Foundation (the "make Go look like Python at the entry boundary")

Goal: Fix C1, C2, C8, C9, C10. After this phase the trace divergence at i=254990 disappears, and the auto-generated invariants actually flow into `model.Invars` (even if some of them are still wrong; subsequent phases fix the contents).

### Phase 2 — User input flow

Goal: Fix C6, C7. User-provided `tactic_invars` and `tactic_defns` are processed and become part of the goal/invariant set.

### Phase 3 — Auto-generation correctness

Goal: Fix C3, C4, C11, C12, C16. These are the per-task and global invariant generators that are individually broken.

### Phase 4 — Auto5-specific

Goal: Fix C13, C14. These are the schedule-stability invariants that only matter for `l2s_auto5` (not used by ord_live but in scope per the user's "fix all bugs" instruction).

### Phase 5 — WhenOperator and theory finite

Goal: Fix C15, C17. Edge-case correctness.

### Phase 6 — Hardening / errors

Goal: Fix C18, M1.

### Phase 7 — Trace hook infrastructure

Goal: Fix C5. This is independent and can in principle be done in any phase, but it touches `ast/decl_ast.go`, `check/`, `module/`, and is best done last so all the data it needs (tasks, triggers, subs maps) is already correctly populated.

### Phase 8 — HIGH-severity hardening

Goal: Fix H1–H13 in priority order. Several are one-line fixes (H4, H10, H11, H13). Two require non-trivial changes (H6 variable ordering; H7 split_returns).

### Phase 9 — MEDIUM and TRACE-ONLY (cleanups)

Goal: M1–M7, T1–T4. Mostly print blocks, ordering, and dead-code cleanups.

## Files to modify

### A. New helpers and shared infrastructure

**No new helpers needed.** All required functions already exist (verified):
- `proof.GoalVocab(goal)` at `proof/goal.go:210`
- `proof.CompileWithGoalVocab(expr, goal) lg.Expr` at `proof/goal.go:386`
- `proof.CompileDefinitionGoalVocab(cfg, df, goal) *ast.LabeledFormula` at `proof/goal.go:408`
- `il.LabelTemporal(fmla, label) lg.Expr` at `ivylogic/util.go:617`
- `il.NewWithSymbols(sig, syms)` at `ivylogic/sig.go:385`
- `il.NewWithSorts(sig, sorts)` at `ivylogic/sig.go:423`

`il.NewWithSorts.Enter()` is the function that emits the missing `ivylogic.WithSorts.Enter nSorts=%d` xtrace at `ivylogic/sig.go:429`.

### B. Phase 1 — `l2s/l2s.go` — entry point and invariant flow

#### Edit B1: New wrapper helper `l2sTactic`

After the existing `--- Tactic entry points ---` comment block, add a new internal wrapper that mirrors Python's `l2s_tactic`:

```go
// l2sTactic mirrors Python's l2s_tactic (ivy_l2s.py:75-79).
// It enters WithSymbols and WithSorts contexts before calling l2sTacticInt.
// All three public entry points (L2STactic, L2STacticFull, L2STacticAuto)
// route through this helper.
func l2sTactic(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, pf ast.Node, tacticName string) ([]*ast.LabeledFormula, error) {
    if len(goals) == 0 {
        return nil, fmt.Errorf("l2s: no proof goals")
    }
    m := pc.GetModule()
    if m == nil || m.Sig == nil {
        return nil, fmt.Errorf("l2s: module or sig is nil")
    }
    vocab := proof.GoalVocab(goals[0])
    ws := il.NewWithSymbols(m.Sig, vocab.Symbols)
    ws.Enter()
    defer ws.Exit()
    wso := il.NewWithSorts(m.Sig, vocab.Sorts)
    wso.Enter()
    defer wso.Exit()
    return l2sTacticInt(pc, goals, pf, tacticName)
}
```

#### Edit B2: Rewrite `L2STactic`/`L2STacticFull`/`L2STacticAuto`

Replace the three trivial wrappers with versions that route through `l2sTactic` and pass the user's actual tactic name (for `L2STacticAuto`):

```go
func L2STactic(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
    return l2sTactic(pc, goals, pf, "l2s")
}

func L2STacticFull(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
    result, err := l2sTactic(pc, goals, pf, "l2s_full")
    // C5: result[0].TraceHook = check.L2STraceHook (deferred to Phase 7)
    return result, err
}

func L2STacticAuto(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
    // C2: extract user's actual tactic name
    tacticName := "l2s_auto" // fallback
    if tt, ok := pf.(*ast.TacticTactic); ok && tt.TName != nil {
        tacticName = nodeToString(tt.TName) // use proof.nodeToString or duplicate locally
    }
    return l2sTactic(pc, goals, pf, tacticName)
}
```

Note: `nodeToString` is in `proof/checker.go:632` (unexported). Either export it as `proof.NodeToString`, or add a local copy in `l2s/`. Local copy is simpler and matches the existing pattern.

#### Edit B3: `l2sTacticInt` — start `invars` from compiled tactic_invars (Phase 2 hookup), commit to model.Invars (C8), call Desugar (C9), include axiom temporal prems (C10)

At `l2s/l2s.go:272-274`, replace:

```go
var invars []*ast.LabeledFormula
invars = append(invars, model.Invars...)
```

with the Python-faithful flow (with placeholders for Phase 2 user-decl processing):

```go
// Phase 2 will populate this from pf.(*ast.TacticTactic).TacticDeclsList()
// filtered/compiled. For Phase 1 it stays empty (matches the empty
// tactic_invars case in Python).
var invars []*ast.LabeledFormula
```

**Do NOT** seed from `model.Invars` — that's the C8/C6 bug. The pre-existing `model.Invars` already contains the original module conjectures and they stay in `model.Invars`; the local slice should hold ONLY the invariants generated/compiled within this tactic call.

At `l2s/l2s.go:239-249` (the temporal-premise loop), add the axiom branch (C10):

```go
// C10: also include non-explicit temporal axioms in temporalPrems
// (mirrors Python ivy_l2s.py:134-135).
for _, ax := range pc.GetAxioms() {
    if !ax.Explicit && ax.IsTemporal() {
        if f, ok := ax.Formula.(lg.Expr); ok {
            temporalPrems = append(temporalPrems, f)
        }
    }
}
```

After the auto-invariant generation block (line 307), add the desugar + commit (C8 + C9):

```go
// C9: desugar $was/$happened operators in all invariants
for i, inv := range invars {
    if expr, ok := inv.Formula.(lg.Expr); ok {
        invars[i] = m.Cfg.AstCfg.NewLabeledFormula(inv.Label, Desugar(expr, proofLabel))
    }
}

// C8: commit auto-generated invariants into model.Invars (Python ivy_l2s.py:722)
model.Invars = append(model.Invars, invars...)
```

Then **simplify `modPass`** at lines 326-343 to operate only on `model.Invars` (the local `invars` slice is no longer separately tracked since it's now in `model.Invars`):

```go
modPass := func(transform func(lg.Expr) lg.Expr) {
    for i, inv := range model.Invars {
        model.Invars[i] = l2sAcfg.NewLabeledFormula(inv.Label, transform(inv.Formula.(lg.Expr)))
    }
    for i, asm := range model.Asms {
        model.Asms[i] = l2sAcfg.NewLabeledFormula(asm.Label, transform(asm.Formula.(lg.Expr)))
    }
    for i, b := range model.Bindings {
        newStmt := transformAction(b.Action.Stmt, transform)
        model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))
    }
    if model.Init != nil {
        model.Init = transformAction(model.Init, transform)
    }
    // (M7 will add: list_transform on goal prems — deferred)
}
```

Delete the `cfg.Invars: invars` field from the `InstrumentationConfig` literal at line 319 (it's dead code per H8) — but only after verifying nothing reads it. Verified above: nothing reads `cfg.Invars`, so the field can be removed from `shared.go:30` too. Defer the field-deletion to Phase 9 to keep this phase minimal.

#### Edit B4: Preserve assumed gprops label (H13)

At `l2s/l2s.go:257`:

```go
// before:
model.Asms = append(model.Asms, m.Cfg.AstCfg.NewLabeledFormula(nil, g.Body))
// after:
model.Asms = append(model.Asms, m.Cfg.AstCfg.NewLabeledFormula(ax.Label, g.Body))
```

(H13 is small enough to fold into Phase 1.)

### C. Phase 2 — `l2s/l2s.go` — user tactic_decls flow (C6, C7)

After the existing `proofLabel := ""` line (`l2s/l2s.go:270`), add:

```go
// C6/C7: process user-supplied tactic_decls (Python ivy_l2s.py:124-153, 175)
if tt, ok := pf.(*ast.TacticTactic); ok {
    // M1: reject tactic_lets
    if tt.Body != nil {
        if _, isLets := tt.Body.(*ast.TacticLets); isLets {
            return nil, fmt.Errorf("tactic does not take lets")
        }
    }

    decls := tt.TacticDeclsList()
    var tacticInvars []*ast.LabeledFormula
    var tacticDefns []ast.Node
    for _, d := range decls {
        if dd, isDerived := d.(*ast.DerivedDecl); isDerived {
            tacticDefns = append(tacticDefns, dd)
        } else if lf, isLF := d.(*ast.LabeledFormula); isLF {
            tacticInvars = append(tacticInvars, lf)
        }
    }

    // C7: compile definitions into goal premises
    for _, defn := range tacticDefns {
        goal = proof.CompileDefinitionGoalVocab(m.Cfg.AstCfg, defn, goal)
    }

    // C6: compile user invariants and seed `invars`
    for _, inv := range tacticInvars {
        compiled := proof.CompileWithGoalVocab(inv.Formula, goal)
        if compiled == nil {
            continue
        }
        labeled := il.LabelTemporal(compiled, proofLabel)  // see note below
        invars = append(invars, m.Cfg.AstCfg.NewLabeledFormula(inv.Label, labeled))
    }
}
```

**All helpers needed for Phase 2 already exist** (verified):
- `proof.CompileWithGoalVocab(expr, goal)` at `proof/goal.go:386`
- `proof.CompileDefinitionGoalVocab(cfg, df, goal)` at `proof/goal.go:408`
- `il.LabelTemporal(fmla, label)` at `ivylogic/util.go:617`

**`TacticDeclsList`/`TacticLetsList` already exist** on `*ast.TacticTactic` at `ast/tactic.go:374` and `:383`.

### D. Phase 3 — `l2s/l2s_auto.go` — per-task invariant fixes (C3, C4, C11, C12, C16)

#### Edit D1: Add `notWaitingForStart` wrap to `notAllDone` for auto2/3/4/5 (C3)

At `l2s/l2s_auto.go:429` (right after `notAllDone := &lg.Not{...}`), insert:

```go
// C3 / Python ivy_l2s.py:354-355
if tacticName == "l2s_auto2" || tacticName == "l2s_auto3" ||
    tacticName == "l2s_auto4" || tacticName == "l2s_auto5" {
    notAllDone = &lg.Or{Terms: []lg.Expr{
        &lg.Not{Body: notWaitingForStart},
        notAllDone,
    }}
}
```

#### Edit D2: Multi-task `l2s_not_all_done` (C4)

Add accumulators outside the per-task loop:

```go
notAllDonePreds := []lg.Expr{}
notAllWasDonePreds := []lg.Expr{}
schedExistsPreds := []lg.Expr{}
```

Replace lines 431-433 (the single-task `if idx == len(sortedTasks)-1 { ... }` block) with the per-task append + per-boundary emit:

```go
// C4 / Python ivy_l2s.py:519-526
notAllDonePreds = append(notAllDonePreds, notAllDone)
if idx+1 < len(sortedTasks) {
    nextSfx := sortedTasks[idx+1]
    if trig, ok := triggers[nextSfx]; ok && trig["work_start"] != nil {
        trigf := trig["work_start"]
        orOfPreds := buildOrExpr(notAllDonePreds)
        tmp := &lg.Implies{
            T1: &lg.Not{Body: eventuallyStartTask(trigf)},
            T2: orOfPreds,
        }
        invars = appendLF(autoAcfg, invars, "l2s_not_all_done"+sfx, tmp)
        notAllDonePreds = nil
    }
}
notAllWasDonePreds = append(notAllWasDonePreds, notAllWasDone(workNeeded, 0))
```

After the per-task loop, add the global emission (Python `ivy_l2s.py:545-546`):

```go
// C4 final: l2s_not_all_done = Or(*not_all_done_preds)
if len(notAllDonePreds) > 0 {
    invars = appendLF(autoAcfg, invars, "l2s_not_all_done", buildOrExpr(notAllDonePreds))
}
```

`buildOrExpr` is a small helper: `func buildOrExpr(xs []lg.Expr) lg.Expr` returning `xs[0]` if len 1, `&lg.Or{Terms: xs}` otherwise (matches Python's `lg.Or(*xs)` behavior — single-arg Or returns its arg).

#### Edit D3: `l2s_progress_made` rewrite (C11)

This is the largest single edit. Replace `l2s/l2s_auto.go:388-406` with a faithful port of `ivy_l2s.py:478-505`. The exact structure:

```go
// C11 / Python ivy_l2s.py:478-505
if tacticName != "l2s_auto3" {
    var progressInv lg.Expr

    if tacticName == "l2s_auto5" {
        // Python lines 479-490
        nad := getDepends(workHelpful, workProgress)
        wasNad := applyNB(l2sS(progressArgs, nad, proofLabel), varsToNodes(progressArgs)...)
        progressInv = makeAnd(
            l2sSaved,
            evStart,
            exists(progressArgs, nad),
            notAllWasDone(workNeeded, 0),
            forall(progressArgs, &lg.Implies{
                T1: nad,
                T2: &lg.Not{Body: applyNB(waitingForProgress, varsToNodes(progressArgs)...)},
            }),
        )
        _ = wasNad // used by sched_stable below
    } else if len(progressArgs) > 0 || len(tasks) > 1 {
        // Python lines 491-501
        nad := makeAnd(
            notAllWasDone(workNeeded, len(progressArgs)),
            &lg.Not{Body: buildOrExpr(notAllWasDonePreds)},
        )
        if nextTaskHasTrigger(idx, sortedTasks, triggers) {
            nad = makeAnd(nad, nextTaskNotTriggered(idx, sortedTasks, triggers))
        }
        progressInv = forall(progressArgs, &lg.Implies{
            T1: makeAnd(
                nad,
                l2sSaved,
                evStart,
                &lg.Not{Body: applyNB(waitingForProgress, varsToNodes(progressArgs)...)},
            ),
            T2: exists(doneArgs[len(progressArgs):],
                makeAnd(&lg.Not{Body: wasDone}, isDoneNode)),
        })
    } else {
        // Python lines 502-505 (the simple case)
        progressInv = &lg.Implies{
            T1: makeAnd(l2sSaved, &lg.Not{Body: applyNB(waitingForProgress)}),
            T2: exists(doneArgs, makeAnd(&lg.Not{Body: wasDone}, isDoneNode)),
        }
    }
    invars = appendLF(autoAcfg, invars, "l2s_progress_made"+sfx, progressInv)
}
```

This requires four new closures inside `l2sAutoInvariants`:

- **`notAllWasDone(defn *lg.Eq, skip int) lg.Expr`** — Python `ivy_l2s.py:361-363`. Calls `getWasDone(defn)` and wraps with `Not(forall(workDone.LHS.Args[skip:], ...))`.

- **`getWasDone(defn *lg.Eq) lg.Expr`** — Python `ivy_l2s.py:358-360`. Calls `getWorkWasDone(defn, workDone)`.

- **`getWorkWasDone(defn, workDoneL *lg.Eq) lg.Expr`** — Python `ivy_l2s.py:236-245`. Branches on `tacticName not in ["l2s_auto3","l2s_auto4","l2s_auto5"]`.

- **`getDepends() lg.Expr`** — Python `ivy_l2s.py:365-372`. Substitutes `workHelpful.LHS.Args -> workProgress.LHS.Args` into `workHelpful.RHS`.

- **`nextTaskHasTrigger(idx int) bool`** and **`nextTaskNotTriggered(idx int) lg.Expr`** — Python `ivy_l2s.py:462-467`.

Add these as local closures at the top of the per-task loop, mirroring the Python definitions inside `l2s_tactic_int`'s task loop.

#### Edit D4: `l2s_needed_were_frozen<sfx>` (C12)

Inside the `else` of `tacticName != "l2s_auto4" && tacticName != "l2s_auto5"` at `l2s/l2s_auto.go:319-340`, after emitting `l2s_needed_are_frozen`, add:

```go
// C12 / Python ivy_l2s.py:422-425
wasDone2 := getWorkWasDone(workNeeded, workDone)
notWasDone := &lg.Not{Body: wasDone2}
tmp2 := &lg.Implies{T1: notWasDone, T2: makeAnd(aCons...)}
tmp3 := &lg.Implies{
    T1: makeAnd(evStart, l2sSaved),
    T2: tmp2,
}
invars = appendLF(autoAcfg, invars, "l2s_needed_were_frozen"+sfx, tmp3)
```

#### Edit D5: `init_globally` EF/AG extras (C16)

In the `initGlobally` closure at `l2s/l2s_auto.go:441-506`, in the `case *lg.Eventually` `pos=true` branch, after appending the base `Implies{T1: prop, T2: notWaiting}`, add the EF-pattern extra (Python `ivy_l2s.py:564-568`):

```go
// C16: EF-pattern extra
if isGloballyOrNotEventually(arg) {
    *res = append(*res, &lg.Implies{T1: notWaiting, T2: arg})
}
```

In the `case *lg.Globally` `pos=false` branch, after the base `Implies{T1: prop, T2: notWaiting}`, add the AG-pattern extra (Python `ivy_l2s.py:572-576`):

```go
// C16: AG-pattern extra
if isEventuallyOrNotGlobally(arg) {
    *res = append(*res, &lg.Implies{T1: notWaiting, T2: &lg.Not{Body: arg}})
}
```

Helpers:

```go
func isGloballyOrNotEventually(e lg.Expr) bool {
    if _, ok := e.(*lg.Globally); ok { return true }
    if n, ok := e.(*lg.Not); ok {
        if _, ok := n.Body.(*lg.Eventually); ok { return true }
    }
    return false
}

func isEventuallyOrNotGlobally(e lg.Expr) bool {
    if _, ok := e.(*lg.Eventually); ok { return true }
    if n, ok := e.(*lg.Not); ok {
        if _, ok := n.Body.(*lg.Globally); ok { return true }
    }
    return false
}
```

### E. Phase 4 — `l2s/l2s_auto.go` — auto5-only invariants (C13, C14)

Inside the per-task loop, after `l2s_progress_invar` emission, add:

```go
// C13 / Python ivy_l2s.py:532-538
if tacticName == "l2s_auto5" {
    nad := getDepends()
    wasNad := applyNB(l2sS(progressArgs, nad, proofLabel), varsToNodes(progressArgs)...)
    waitingForProgressApp := applyNB(waitingForProgress, varsToNodes(progressArgs)...)
    stableInv := forall(progressArgs, &lg.Implies{
        T1: makeAnd(wasNad, l2sSaved, evStart, waitingForProgressApp),
        T2: nad,
    })
    invars = appendLF(autoAcfg, invars, "l2s_sched_stable"+sfx, stableInv)
    schedExistsPreds = append(schedExistsPreds, exists(progressArgs, wasNad))
}
```

After the per-task loop, before `init_globally`, add (Python `ivy_l2s.py:548-550`):

```go
// C14
if tacticName == "l2s_auto5" && len(schedExistsPreds) > 0 {
    tmp := &lg.Implies{
        T1: makeAnd(l2sSaved, evStart),  // Python uses bare `eventually_start()` referring to last task
        T2: buildOrExpr(schedExistsPreds),
    }
    invars = appendLF(autoAcfg, invars, "l2s_sched_exists", tmp)
}
```

Verify which `evStart` to use here: Python's `eventually_start()` is a closure that captures the loop's `work_start`. Since the loop has ended, it refers to the LAST iteration's `work_start`. Preserve this by capturing the last `evStart` outside the loop body.

### F. Phase 5 — `l2s/l2s_auto.go` — `l2s_when_<i>` (C15) and theory finite (C17)

#### Edit F1: `l2s_when_<i>` (C15)

After `neg_prop_init` emission, add a block that walks `invars` (and goal property prems) for `*lg.WhenOperator{Name: "first"}` and emits an invariant per occurrence. See the audit's Edit 12 for the exact structure (Python `ivy_l2s.py:666-676`).

This requires a `uniqueTemporals(...)` helper (Python uses `ilu.temporals_asts(...)` which collects all temporal-operator subexpressions). Implement using `lu.NamedBindersAst` or a custom walker that enumerates `*lg.Globally`, `*lg.Eventually`, `*lg.WhenOperator` subnodes.

#### Edit F2: theory finite check (C17)

At `l2s/l2s.go:286`:

```go
// before:
if m.FiniteSorts[name] || full {
    finiteSorts[name] = true
}
// after (C17 / Python ivy_l2s.py:194):
isTheoryFinite := false
if thy := theory.GetSortTheory(s, m.Interp); thy != nil && thy.IsFinite() {
    isTheoryFinite = true
}
if m.FiniteSorts[name] || full || isTheoryFinite {
    finiteSorts[name] = true
}
```

If `theory.Theory` doesn't have `IsFinite()`, add it. Bit-vector theories return true; uninterpreted-sort theories return false.

### G. Phase 6 — Hardening / errors (C18, M1, H3, H4, H5)

#### Edit G1: C18 — `getAuxDefn` filter and error

At `l2s/l2s_auto.go:50-91`, add `if !premLF.IsDefinition { continue }`, change "silently skip on free vars" to return an error:

```go
if len(freeVars) > 0 {
    return nil, fmt.Errorf("free symbol %s not allowed in definition of %s", freeVars[0].Name, dname)
}
```

(Adjust function signature to return an error.)

#### Edit G2: M1 — reject `tactic_lets`

Already covered in Phase 2 Edit C (the `if _, isLets := tt.Body.(*ast.TacticLets); isLets` check).

#### Edit G3: H3 — `work_progress` prefix-of-`work_done` validation

In the per-task loop in `l2s/l2s_auto.go`, before emitting `l2s_progress_made`:

```go
// H3 / Python ivy_l2s.py:475-476
if tacticName != "l2s_auto5" {
    progArgs := eqLHSArgs(workProgress)
    doneArgsForCheck := eqLHSArgs(workDone)
    if len(progArgs) > len(doneArgsForCheck) {
        return nil, fmt.Errorf("work_progress%s must be a prefix of work_done%s args", sfx, sfx)
    }
    for i := range progArgs {
        if progArgs[i].Name != doneArgsForCheck[i].Name {
            return nil, fmt.Errorf("work_progress%s must be a prefix of work_done%s args", sfx, sfx)
        }
    }
}
```

#### Edit G4: H4 — `work_invar` no-arguments validation

At `l2s/l2s_auto.go` after auxiliary defn extraction:

```go
// H4 / Python ivy_l2s.py:320-321
if wi := tasks[sfx]["work_invar"]; wi != nil {
    if app, ok := wi.T1.(*lg.Apply); ok && len(app.Terms) > 0 {
        return nil, fmt.Errorf("work_invar%s may not have arguments", sfx)
    }
}
```

#### Edit G5: H5 — sort-signature consistency

```go
// H5 / Python ivy_l2s.py:314-319
for _, sfx := range sortedTasks {
    task := tasks[sfx]
    if !sameLHSSort(task["work_created"], task["work_needed"]) {
        return nil, fmt.Errorf("work_created%s and work_needed%s must have same signature", sfx, sfx)
    }
    if !sameLHSSort(task["work_created"], task["work_done"]) {
        return nil, fmt.Errorf("work_created%s and work_done%s must have same signature", sfx, sfx)
    }
    if h := task["work_helpful"]; h != nil {
        if !sameLHSSort(h, task["work_progress"]) {
            return nil, fmt.Errorf("work_helpful%s and work_progress%s must have same signature", sfx, sfx)
        }
    }
}
```

`sameLHSSort` is a small helper.

### H. Phase 7 — Trace hook infrastructure (C5)

This phase touches multiple files and may need a new package boundary to avoid import cycles.

#### Edit H1: `ast/decl_ast.go` — add `TraceHook` field

```go
type LabeledFormula struct {
    // ... existing fields ...
    TraceHook ast.TraceHookFunc  // NEW: l2s diagnostic hook
}
```

Define `TraceHookFunc` as `func(handler interface{}, fcs []interface{}) interface{}` in `ast/` (or in a new `trace/hook.go` if cleaner). Use `interface{}` to avoid the `check` <-> `ast` import cycle. Type-assert at the consumer side (in `check/`).

#### Edit H2: `module/module.go` — add `TraceHook` field on `Module`

```go
type Module struct {
    // ... existing fields ...
    TraceHook ast.TraceHookFunc
}
```

#### Edit H3: `l2s/l2s.go` — set the hook in tactic results

After `SharedStep12_BuildGoal`:

```go
// C5: attach trace hook to result goal
if len(result) > 0 && result[0] != nil {
    if strings.HasPrefix(tacticName, "l2s_auto5") {
        result[0].TraceHook = makeAutoHookClosure(tasks, triggers, subs)
    } else if strings.HasPrefix(tacticName, "l2s_auto") {
        result[0].TraceHook = makeRenamingHookClosure(subs)
    } else if tacticName == "l2s_full" {
        result[0].TraceHook = makeTraceHookClosure()
    }
}
```

`makeAutoHookClosure`/`makeRenamingHookClosure`/`makeTraceHookClosure` are local factories that wrap the existing functions in `check/l2s_hooks.go`. They live in `l2s/` to avoid the `l2s` -> `check` cycle.

#### Edit H4: `check/isolate_check.go` — wire hook from goal to module

Replace the TODOs at lines 804-805 and 857-858:

```go
// C5: propagate trace hook from goal to module
if goal.TraceHook != nil {
    fakeMod.TraceHook = goal.TraceHook
}
```

#### Edit H5: `check/check.go` — invoke `mod.TraceHook` on trace display

Replace the stub at line 589:

```go
if mod.TraceHook != nil {
    handler = mod.TraceHook(handler, ffcs).(<concrete handler type>)
}
```

(The concrete cast type depends on the trace handler interface; check existing code.)

### I. Phase 8 — HIGH-severity hardening (H1–H13)

Most are localized and listed in Section "Bug inventory" above. Apply each per the audit's "Implementation plan" section. Order: easy ones first (H4, H10, H11, H13 — already partially done), then medium (H1, H3, H5, H9, H12), then expensive (H6, H7).

### J. Phase 9 — MEDIUM and TRACE-ONLY (M1–M7, T1–T4)

- **M2**: call `proof.RemoveUnusedDefinitionsGoal(m.Cfg.AstCfg, goal)` before `SharedStep12_BuildGoal`.
- **M3**: reorder invariant emission in `l2s_auto.go` to match Python: globally → status → consts_d → init_glob → neg_prop → when.
- **M4/M5**: align `allD`/`allA`/`l2s_consts_d` to always emit (logically equivalent but structurally identical to Python).
- **M6**: add the print blocks (Python lines 679-682, 874-876) — these will appear in test output but are gated on debug flags so they shouldn't break golden.
- **M7**: add `list_transform(prems, transform)` equivalent in `modPass`.
- **T1/T2**: Same as M6.
- **T4**: enrich `check/l2s_hooks.go` diagnostic messages to match Python's `auto_hook` detail.
- **H8**: delete `cfg.Invars` field and the assignment at `l2s.go:319`. Verify nothing reads it (already verified above).

### K. Delete duplicate temporal/l2s helpers (cleanup, optional)

Not in scope for this plan — covered by the previously-applied TemporalModels plan.

## Verification

### Per-phase verification

After each phase:

1. **Compile**: `cd ~/ivy/goivy && go build ./...` — must be clean.
2. **Unit tests**: `go test ./l2s/... ./proof/... ./check/...` — must pass.

### Full-pipeline verification (after Phase 1)

After Phase 1, the trace divergence at i=254990 should disappear. Run:

```sh
cd ~/ivy/goivy && make golden 2>&1 | tee log.parser.red
```

Expected: the `ivylogic.WithSorts.Enter nSorts=0` trace appears in Go output at i=254990, matching Python. The next divergence (if any) tells us where the next bug surfaces — almost certainly inside `l2sTacticInt` itself (because the tactic name now passes through correctly and many auto branches now fire that didn't before).

### Full-pipeline verification (after Phase 3)

After Phase 3, the auto-invariant generation should match Python's structure for `l2s_auto`/`l2s_auto2`/`l2s_auto4`. The TestOrdLive proof should advance significantly. Run:

```sh
cd ~/ivy/goivy && make golden 2>&1 | tee log.parser.red
go test -v -run TestOrdLive ./parser/...
```

Expected: TestOrdLive divergence point shifts forward by many trace events. Some tactic-level invariant comparisons should now pass.

### Final verification

After all phases:

1. `go build ./...` clean.
2. `go test ./...` green for `l2s`, `proof`, `check`, `module`, `tactics`, `temporal`.
3. `make golden` shows TestOrdLive passing OR diverging at a point clearly outside `l2s` (e.g., in a downstream check or solver call).
4. **Spot-check** the auto-generated invariants for ord_live's `cf_liveness` proof: dump `model.Invars` after `l2sTacticInt` returns and confirm the names match Python's auto5 set: `l2s_needed_when_start`, `l2s_created`, `l2s_needed_are_frozen`, `l2s_done_implies_created`, `l2s_needed_implies_created`, `l2s_work_preserved`, `l2s_progress_made`, `l2s_progress_invar`, `l2s_not_all_done`, `l2s_globally_*`, `l2s_init_glob_*`, `neg_prop_init`, `l2s_status_*`, `l2s_consts_d`. Set count and names should match Python's output for the same input.
5. **Microtest** for C8: construct a minimal `l2s_auto` proof goal in a unit test, call `L2STacticAuto`, and assert that `result[0]`'s underlying model contains `l2s_created`, `l2s_work_preserved`, `l2s_not_all_done` invariants. Currently this would fail because none of those reach `model.Invars`.

## Risks and rollback

- **Risk: Phase 1 alone exposes many secondary divergences.** The current Go code has been "incorrectly correct" — tactic_name is hardcoded so the auto2/3/4/5-conditional bugs are masked. Once C2 is fixed, ALL the conditional branches become live, surfacing C3, C11, C12 etc. The fix must therefore be applied as a coherent package; phases 1+3 must land together (or at least be developed together) to keep `make golden` interpretable.

- **Risk: C8 (model.Invars merge) changes downstream behavior.** Once auto-generated invariants reach `model.Invars`, the named-binder collection, action instrumentation, and goal construction will start emitting symbols that previously didn't exist. This may surface bugs in `SharedStep3-12` that were dormant. Be prepared to fix those in follow-up plans.

- **Risk: C11 (`l2s_progress_made` rewrite)** is the largest single edit and the most error-prone. Recommend implementing it in a TDD cycle: write a microtest that constructs a known-shape `l2s_auto2` task and asserts the produced `l2s_progress_made` formula matches a hand-written expected.

- **Risk: C5 (trace_hook infrastructure)** introduces a cross-package dependency from `ast` -> hook signature. If `ast` cannot import `trace`, the function type must be `interface{}` or a generic `func(...) ...`. Keep the field type loose to avoid cycles.

- **Risk: H6 (variable ordering)** is non-trivial. Changing `module.VariablesAST` to be deterministic may affect every binder construction site in the codebase. May warrant a separate plan if it cascades.

- **Rollback**: every phase is locally reversible via `git checkout` of the affected files. The phases are organized so each adds correctness without breaking what came before. Phase 1 is the riskiest commit point because it changes the entry point semantics.

## Out of scope

- Refactoring `temporal`/`l2s` package boundaries (separate concern).
- Adding exhaustive unit tests for every fixed invariant generator. The golden test exercises them end-to-end; we add microtests only for the highest-risk edits (C11, C5).
- Fixing bugs in `tactics/`, `proof/`, `temporal/` that aren't called from `l2s/`. Those are separate plans.
- The next divergence after this fix — separate follow-up plan once `make golden` runs cleanly through the l2s phase.

## Implementation checklist

### Phase 1 — Foundation
- [ ] B1: Add `l2sTactic` wrapper in `l2s/l2s.go`.
- [ ] B2: Rewrite `L2STactic`/`L2STacticFull`/`L2STacticAuto` to route through `l2sTactic` and pass user's tactic name.
- [ ] B3: `l2sTacticInt` — empty initial `invars`, axiom temporal prems (C10), Desugar (C9), commit to `model.Invars` (C8), simplified `modPass`.
- [ ] B4: Preserve assumed gprops label (H13).
- [ ] **Verify**: `go build ./...` clean; `make golden` shows divergence past i=254990.

### Phase 2 — User input flow
- [ ] C: Process `proof.tactic_decls` — split DerivedDecl vs LabeledFormula; compile defns into goal (C7); compile invars and seed `invars` (C6); reject tactic_lets (M1).
- [ ] **Verify**: `go test ./l2s/... ./proof/...` passes.

### Phase 3 — Auto-generation correctness
- [ ] D1: `notAllDone` Or-wrap for auto2+ (C3).
- [ ] D2: Multi-task `l2s_not_all_done` accumulator (C4).
- [ ] D3: `l2s_progress_made` rewrite (C11) + helper closures `notAllWasDone`, `getWasDone`, `getWorkWasDone`, `getDepends`, `nextTaskHasTrigger`, `nextTaskNotTriggered`.
- [ ] D4: `l2s_needed_were_frozen<sfx>` (C12).
- [ ] D5: `init_globally` EF/AG extras (C16).
- [ ] **Verify**: `make golden` advances past Phase 3's expected divergence.

### Phase 4 — Auto5-specific
- [ ] E: `l2s_sched_stable<sfx>` (C13) and `l2s_sched_exists` (C14).

### Phase 5 — WhenOperator and theory finite
- [ ] F1: `l2s_when_<i>` (C15) + `uniqueTemporals` helper.
- [ ] F2: theory finite check (C17) + maybe `theory.Theory.IsFinite()` method.

### Phase 6 — Hardening
- [ ] G1: `getAuxDefn` filter + error (C18).
- [ ] G2: tactic_lets reject (M1, already in Phase 2).
- [ ] G3: `work_progress` prefix validation (H3).
- [ ] G4: `work_invar` no-arguments validation (H4).
- [ ] G5: sort-signature consistency (H5).

### Phase 7 — Trace hook infrastructure (C5)
- [ ] H1: `ast.LabeledFormula.TraceHook` field + `TraceHookFunc` type.
- [ ] H2: `module.Module.TraceHook` field.
- [ ] H3: `l2s/l2s.go` set hook on result goals via local factories.
- [ ] H4: `check/isolate_check.go` propagate goal hook to module.
- [ ] H5: `check/check.go` invoke `mod.TraceHook` on trace display.

### Phase 8 — HIGH-severity hardening
- [ ] H1 (subscript): `tactic_proof` block handling.
- [ ] H6: deterministic `module.VariablesAST`.
- [ ] H7: `CallAction` split-returns.
- [ ] H9: `getAuxDefn` single-ForAll strip.
- [ ] H10: `getAuxDefn` LHS rename to generic.
- [ ] H11: `Eventually.Environ` consistency (3 sites).
- [ ] H12: `defn_deps` includes `prem_defns`.

### Phase 9 — MEDIUM and TRACE-ONLY cleanups
- [ ] M2: `RemoveUnusedDefinitionsGoal`.
- [ ] M3: invariant emission order.
- [ ] M4/M5: always-emit `allD`/`allA`/`l2s_consts_d`.
- [ ] M6/T1/T2: print blocks.
- [ ] M7: `list_transform(prems)`.
- [ ] H8: delete `cfg.Invars` dead field.
- [ ] T4: enrich `check/l2s_hooks.go` diagnostic messages.

### Final verification
- [ ] `go build ./...` clean.
- [ ] `go test ./...` green (l2s, proof, check, module, tactics, temporal).
- [ ] `make golden` — TestOrdLive passes or diverges in code outside l2s.
- [ ] Microtest: minimal l2s_auto proof yields a result whose model contains all expected auto invariants.

## Critical files

- `/Users/jaten/ivy/goivy/l2s/l2s.go` — entry points, tactic wrapper, modPass, finite sorts, axiom prems, hook wiring
- `/Users/jaten/ivy/goivy/l2s/l2s_auto.go` — invariant generators, helper closures, validation
- `/Users/jaten/ivy/goivy/l2s/shared.go` — cfg.Invars cleanup, BuildDefnDeps, instr_stmt
- `/Users/jaten/ivy/goivy/ast/decl_ast.go` — `LabeledFormula.TraceHook` field
- `/Users/jaten/ivy/goivy/module/module.go` — `Module.TraceHook` field
- `/Users/jaten/ivy/goivy/module/astutil.go` — deterministic `VariablesAST`
- `/Users/jaten/ivy/goivy/check/isolate_check.go` — wire goal hook to module
- `/Users/jaten/ivy/goivy/check/check.go` — invoke `mod.TraceHook`
- `/Users/jaten/ivy/goivy/check/l2s_hooks.go` — enrich diagnostic messages
- `/Users/jaten/ivy/goivy/proof/checker.go` (or `proof/goal.go`) — verify/add `CompileWithGoalVocab`
- `/Users/jaten/ivy/goivy/proof/phase5_goals.go` — `RemoveUnusedDefinitionsGoal` (used by M2)

## Source-of-truth references

- `~/ivy/pyivy/ivy/ivy/ivy_l2s.py` lines 75-79 (`l2s_tactic`), 91-93 (`l2s_tactic_auto`), 86-89 (`l2s_tactic_full`), 109-722 (`l2s_tactic_int` body).
- `~/ivy/pyivy/ivy/ivy/ivy_logic.py` lines 989-1009 (`WithSorts`), 969-987 (`WithSymbols`).
- `~/ivy/pyivy/ivy/ivy/ivy_proof.py` lines 572-583 (`goal_vocab`), 735-756 (`compile_expr_vocab`/`compile_expr_vocab_ext`).
- `~/ivy/pyivy/ivy/ivy/ivy_check.py` lines 406-407, 829-830, 839-840 (trace_hook wiring).
