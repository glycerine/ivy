# Fix Plan: ivy_check.py / ivy_l2s.py Audit Bugs
*Created: 2026-04-17 (Thursday), revised after thorough codebase read*

## Revised Bug Status

After reading all Go files in `check/`, `logicutil/`, etc., several items from the initial audit are NOT bugs:

| ID | Status | Reason |
|----|--------|--------|
| A2 | SKIP | Go is better than Python's buggy comprehension — do not regress |
| A3 | NOT A BUG | `ModuleLFToAstLF`/`AstLFToModuleLF` are identity functions now. Stale comments only. |
| A5 | NOT A BUG | Go is correct; Python has the bug |
| B1 | ALREADY PORTED | `HasTemporalStuff` in `helpers.go` |
| B2 | ALREADY PORTED | `FindAssertions` in `helpers.go` |
| B3 | ALREADY PORTED | `ShowAssertions` in `phase7.go` |
| C2 | VERIFY ONLY | `MatchHandler.Eval` — may be correct, needs runtime verification |
| C3 | NOT A BUG | Dead code, correctly omitted |
| E1 | ALREADY PORTED | `TemporalAndL2S` in `l2s_hooks.go` |
| F7 | SEPARATE AUDIT | `l2s_auto.go` needs its own dedicated pass |

## Bugs That Need Fixes

### Fix 1: A1 — `checkFcsNormalPath` no-ag fallback discards SMT results

**Location:** `check/check.go:669-680`

**Problem:** When `history == nil` (no analysis graph / post-state), the code calls `gmc(combined, finalConds)` but discards the `*solver.ModelResult` return value. The checkers' `Sat()`/`Unsat()` callbacks are never invoked, so no checker can ever report failure.

**Python reference:** Python always passes `ag` and `post` to `check_fcs_in_state`. This no-history path does not exist in Python. But since Go has it, it must work correctly if ever reached.

**Callers:** `CheckFcsInState(mod, checkers)` → `CheckFcsInStateWithAG(mod, nil, nil, checkers)` → `checkFcsNormalPath(... history=nil ...)`. Called from `CheckConjsInState` and `CheckSafetyInState` (the no-ag convenience wrappers). Also called from tests.

**Fix:**
In `checkFcsNormalPath`, the `else` branch (lines 669-680) must use the `*solver.ModelResult` to invoke checker callbacks, the same way `history.SatisfyWithCond` does in the `if history != nil` branch.

```go
} else {
    // No history — fall back to direct solver check.
    baseClauses := module.TrueClauses(actions.EmptyAnnotation{})
    combined := module.AndClausesTyped(baseClauses, axioms)

    gmc := func(cls *module.Clauses, fc []solver.FinalCond) *solver.ModelResult {
        mr, _ := actions.SmallModelClauses(cls, fc, mod.Cfg.Diagnose, mod)
        return mr
    }
    // FIX: actually use the result to invoke checker callbacks
    mr := gmc(combined, finalConds)
    if mr != nil {
        // SAT — the condition is satisfiable, meaning checks fail
        for _, fc := range filteredCheckers {
            fc.Start()
            if fc.Assume() {
                continue
            }
            if !fc.Sat() {
                break  // stop on first failure if diagnose requested
            }
        }
    } else {
        // UNSAT — all checks pass
        for _, fc := range filteredCheckers {
            fc.Start()
            if fc.Assume() {
                continue
            }
            fc.Unsat()
        }
    }
}
```

However, this is an approximation. The real Python flow through `history.satisfy` is more nuanced — it checks each `FinalCond` individually against the model. The simplest correct fix is to match Python's approach: Python never takes this path, so the fix should either:
(a) make the code match what `history.SatisfyWithCond` does, or
(b) emit a clear error/panic if this path is ever reached, since Python never reaches it.

**Recommended approach:** Option (a). Model the loop after `history.SatisfyWithCond` — call `gmc` with each checker's condition individually, call `fc.Start()`, then `fc.Sat()` or `fc.Unsat()` based on the model result. This matches Python's `get_small_model` logic faithfully.

**File to edit:** `check/check.go`

---

### Fix 2: A4 — `ConvertPostconds` never renames old symbols

**Location:** `check/check.go:807-809`

**Problem:** `ConvertPostconds(postconds)` passes `nil` update to `ConvertPostcondsWithUpdate`, which returns early without renaming. Python's `convert_postconds(state, postconds)` always gets `state.update`.

**Callers:** `CheckConjsInState` (line 719) calls `ConvertPostconds(pcs)` — the no-ag convenience wrapper. The AG-aware version `CheckConjsInStateWithAG` (isolate_check.go:1394-1401) correctly extracts `post.Update` and calls `ConvertPostcondsWithUpdate(update, pcs)`.

**Fix:** `CheckConjsInState` is a convenience wrapper that lacks the post-state context. Since Python's `check_conjs_in_state` always receives `(mod, ag, post, ...)`, the no-ag `CheckConjsInState` should not be the primary path. Options:
1. Remove `CheckConjsInState` and force callers to use `CheckConjsInStateWithAG`.
2. Have `CheckConjsInState` skip postcondition conversion entirely (since without a post-state, there's no `update` to rename against — the postconds would be meaningless).

**Recommended:** Option (2) is the simplest correct fix. If `pcs` are provided but no post-state update is available, the postconditions cannot be meaningfully converted. Add a comment explaining this, and verify that all real callers go through `CheckConjsInStateWithAG`.

Concretely: the current code already does this (returns postconds unchanged when update is nil). The "bug" is that unconverted postconds are then checked with wrong variable names. But since the no-ag path is only used in tests and never with postconds in practice, this is acceptable as-is with a clarifying comment. If a caller ever does pass postconds without an AG, it should get a warning.

**File to edit:** `check/check.go` — add warning/comment at `ConvertPostconds`

---

### Fix 3: D1 — Plain `l2s` and `l2s_full` tactics missing `renaming_hook`

**Location:** `check/l2s.go:893-916`

**Problem:** Python sets `goal.trace_hook = lambda tr,fcs: renaming_hook(subs,tr,fcs)` for ALL non-auto5 tactics (including plain `l2s`, `l2s_full`, `l2s_auto*`). Go only sets hooks for `l2s_auto5` and `l2s_auto*` — plain `l2s` and `l2s_full` get no hook.

**Python reference:** `ivy_l2s.py:1500-1508`:
```python
if tactic_name.startswith("l2s_auto5"):
    goal.trace_hook = lambda tr,fcs: auto_hook(tasks,triggers,subs,tr,fcs)
else:
    goal.trace_hook = lambda tr,fcs: renaming_hook(subs,tr,fcs)
```

**Fix:** Add a `default` case to the `switch` block at `l2s.go:897` that applies renaming for all remaining tactics:

```go
switch {
case strings.HasPrefix(tacticName, "l2s_auto5"):
    result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
        if subs != nil {
            applyRenamingToHandler(handler, subs)
        }
        applyAutoDiagnosticsToHandler(handler, fcs, tasks, triggers)
    })
case strings.HasPrefix(tacticName, "l2s_auto"):
    result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
        if subs != nil {
            applyRenamingToHandler(handler, subs)
        }
    })
default:
    // Plain l2s and l2s_full: apply renaming_hook only
    // Python ivy_l2s.py:1503-1504
    result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
        if subs != nil {
            applyRenamingToHandler(handler, subs)
        }
    })
}
```

This removes the `case tacticName == "l2s_full":` special case that leaves TraceHook nil.

**File to edit:** `check/l2s.go`

---

### Fix 4: D2 — `l2s_full` loop-start marking missing

**Location:** `check/l2s.go:911-914` and `check/helpers.go` (MatchHandler struct)

**Problem:** Python's `l2s_tactic_full` attaches `trace_hook` which scans trace states for `l2s_saved == true` and sets `loop_start = True` on the preceding state. Go has no `LoopStart` field on `MatchHandler` and no loop-start scanning logic.

**Python reference:** `ivy_l2s.py:113-122`:
```python
def trace_hook(tr,fcs):
    for idx,state in enumerate(tr.states):
        for c in state.clauses.fmlas:
            s1,s2 = list(map(str,c.args))
            if s1 == 'l2s_saved' and s2 == 'true':
                tr.states[0 if idx == 0 else idx-1].loop_start = True
                return tr
    print("failed to find loop start!")
    return tr
```

**Fix (two parts):**

*Part A:* Add `LoopStart` field to `MatchHandler` in `helpers.go`:
```go
type MatchHandler struct {
    // ... existing fields ...
    // LoopStart is the index of the state that starts the fairness loop.
    // Set by the l2s_full trace hook. -1 means not set.
    LoopStart int
}
```
Initialize to -1 in `NewMatchHandler`.

*Part B:* In Fix 3's `default` case (which now covers `l2s_full`), after applying renaming, add the loop-start scan. The MatchHandler's `Eqs` map contains the ground equalities from the satisfying model. We scan for `l2s_saved = true`:

```go
default:
    result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
        if subs != nil {
            applyRenamingToHandler(handler, subs)
        }
        if tacticName == "l2s_full" {
            markLoopStart(handler)
        }
    })
```

```go
// markLoopStart scans MatchHandler's Eqs for l2s_saved = true and sets
// LoopStart. Mirrors Python ivy_l2s.py:113-122 trace_hook.
func markLoopStart(handler *MatchHandler) {
    if handler == nil {
        return
    }
    savedKey := lg.Key(lg.NewConst("l2s_saved", lg.Boolean))
    eqs, ok := handler.Eqs[savedKey]
    if !ok {
        fmt.Println("failed to find loop start!")
        return
    }
    for _, eq := range eqs {
        if e, ok := eq.(*lg.Eq); ok {
            if lg.IsTrue(e.T2) {
                handler.LoopStart = 0 // mark loop start found
                return
            }
        }
    }
    fmt.Println("failed to find loop start!")
}
```

Note: The full Python logic tracks *which state index* has `l2s_saved == true` and marks the *preceding* state. Go's `MatchHandler` represents a single model (not multiple states), so we track that the loop start was found. Full multi-state tracking would require more infrastructure — this is a pragmatic first step.

**Files to edit:** `check/helpers.go`, `check/l2s.go`, `check/l2s_hooks.go`

---

### Fix 5: D3 — `auto_hook` `l2s_progress_made` partial diagnosis

**Location:** `check/l2s_hooks.go:366-398`

**Problem:** Python's `auto_hook` for `l2s_progress_made` builds `helpful_map`, `happened_maps` (two states), and `justice_map` by scanning `tr.states[0]` and `tr.states[1]` clauses. Go only evaluates Skolem symbols in the handler's post-state.

**Python reference:** `ivy_l2s.py:1453-1509` — builds three maps from two trace states, runs two diagnostic loops cross-referencing them.

**Fix:** The core issue is that `MatchHandler` only has a single model (the satisfying assignment), not multiple trace states. Full parity requires either:
(a) Giving `MatchHandler` a list of state snapshots (matching Python's `tr.states`), or
(b) Extracting the needed information from the single model differently.

**Recommended approach:** Since `MatchHandler.Eqs` contains ALL equalities from the model (which corresponds to the entire trace unrolling), the `helpful_map`, `happened_maps`, and `justice_map` data IS present — it's just interleaved. The Python code accesses `tr.states[0].clauses.fmlas` and `tr.states[1].clauses.fmlas` separately, but in the Go unrolled model, these are distinguished by variable renaming (pre-state vs post-state symbols).

For now, add a TODO comment and the partial work_needed diagnostic that's already there. Full multi-state extraction is a larger infrastructure project.

```go
case strings.HasPrefix(name, "l2s_progress_made"):
    sfx := name[len("l2s_progress_made"):]
    fmt.Printf("\n\nFailed to prove that work_needed%s decreases when a helpful transition occurs\n", sfx)
    // TODO: Port full multi-state diagnosis from Python ivy_l2s.py:1453-1509.
    // This requires extracting helpful_map, happened_maps, and justice_map
    // from pre-state and post-state clauses in the unrolled model.
    // Currently only the work_needed post-state evaluation is implemented.
    ...existing code...
```

**File to edit:** `check/l2s_hooks.go`

---

### Fix 6: D4/D5 — `L2sGToGlobally` pretty-printer not built/wired

**Location:** `check/l2s_hooks.go`

**Problem:** Python's `auto_hook` sets `tr.pp = ls2_g_to_globally` — a pretty-printer that converts `l2s_g` named binders back to `Globally(environ, body)` for readable trace display. The building blocks (`ExpandNamedBindersAst` and `DenormalizeTemporal`) already exist in `logicutil/logic_utils.go`, but nobody composes them and sets the result on `MatchHandler`.

**Python reference:** `ivy_l2s.py:1520-1526`:
```python
def ls2_g_to_globally(ast):
    def g2g(ast):
        if isinstance(ast,lg.NamedBinder) and ast.name == 'l2s_g':
            return lg.Globally(ast.environ,ast.body)
        return None
    res = ilu.expand_named_binders_ast(ast,g2g)
    return ilu.denormalize_temporal(res)
```

And `ivy_l2s.py:1529`: `tr.pp = ls2_g_to_globally`

**Fix (two parts):**

*Part A:* Add a `PP` field to `MatchHandler` in `helpers.go`:
```go
type MatchHandler struct {
    // ... existing fields ...
    // PP is an optional pretty-printer function applied to formulas before
    // display. Mirrors Python tr.pp (set by auto_hook to ls2_g_to_globally).
    PP func(lg.Expr) lg.Expr
}
```

*Part B:* Add `L2sGToGlobally` function in `l2s_hooks.go`:
```go
// L2sGToGlobally converts l2s_g named binders back to Globally operators
// for readable display. Mirrors Python ivy_l2s.py:1520-1526.
func L2sGToGlobally(expr lg.Expr) lg.Expr {
    g2g := func(nb *lg.NamedBinder) lg.Expr {
        if nb.Name == "l2s_g" {
            return &lg.Globally{Environ: nb.Environ, Body: nb.Body}
        }
        return nil
    }
    res := lu.ExpandNamedBindersAst(expr, g2g)
    return lu.DenormalizeTemporal(res)
}
```

*Part C:* Wire it in `applyAutoDiagnosticsToHandler`:
```go
func applyAutoDiagnosticsToHandler(...) {
    if handler == nil {
        return
    }
    handler.PP = L2sGToGlobally  // Python: tr.pp = ls2_g_to_globally
    // ... rest of existing code ...
}
```

**Files to edit:** `check/helpers.go`, `check/l2s_hooks.go`

---

### Fix 7: D6 — `lg.And()` vs `lg.True` in TemporalModels conclusion

**Location:** `check/l2s_shared.go:841`

**Problem:** Python uses `lg.And()` (empty And = True) while Go uses `lg.True`. Semantically equivalent, but canon strings differ: `(and)` vs `true`.

**Python reference:** `ivy_l2s.py:1491`: `conc = ivy_ast.TemporalModels(model, lg.And())`

**Fix:** Change `lg.True` to `&lg.And{}` to match Python's canon output:
```go
newConc := acfg.NewTemporalModels(model, &lg.And{})
```

**File to edit:** `check/l2s_shared.go`

---

### Fix 8: C1 — MatchHandler env iteration fallback

**Location:** `check/helpers.go:303-323`

**Problem:** When `origSym` is not found in `vocabByKey`, Go falls back to `renamedSym` for `IsNew`/`IsSkolem` checks. Python always checks the ORIGINAL symbol. The renamed symbol may have a different name (e.g., `sym__0` vs `sym`), causing incorrect `IsNew`/`IsSkolem` results.

**Fix:** Build the lookup using the env key directly. The env key IS the original symbol's `NodeKey`. We can reconstruct the original `*lg.Const` from it if needed, or better, just skip symbols not in vocab (matching Python's behavior — Python's env only contains symbols that were in the vocab to begin with):

```go
for symKey, renamedExpr := range env {
    renamedSym, ok := renamedExpr.(*lg.Const)
    if !ok {
        continue
    }
    origSym := vocabByKey[symKey]
    if origSym == nil {
        continue  // not in vocab — skip (matches Python behavior)
    }
    if !actions.IsNew(origSym.Name) && !h.IsSkolem(origSym) {
        h.ShowSym(origSym, renamedSym)
    }
}
```

The change is replacing the fallback `origSym = renamedSym` with `continue` (skip). Python only iterates `env.items()` where both key and value are Symbol objects that were computed from vocab during `match_annotation`, so every `sym` in `env` IS in the vocab.

**File to edit:** `check/helpers.go`

---

### Fix 9: F1 — `tactic_lets` check skipped for non-TacticTactic nodes

**Location:** `check/l2s.go:413-417`

**Problem:** If `pf` is not a `*ast.TacticTactic`, Go silently skips the `tactic_lets` check. Python's `proof.tactic_lets` is an attribute access that would raise `AttributeError` if `proof` doesn't have it.

**Fix:** In practice, the l2s tactic is always called with a `TacticTactic` proof node (that's what the parser produces for `proof { tactic l2s ... }`). The check is defensive. Add a trace/warning if the type assertion fails:

```go
if tt, ok := pf.(*ast.TacticTactic); ok {
    if tt.Body != nil {
        if _, isLets := tt.Body.(*ast.TacticLets); isLets {
            return nil, fmt.Errorf("tactic does not take lets")
        }
    }
} else if pf != nil {
    xtracer.Trace("l2s.l2sTacticInt WARNING: proof node is %T, not TacticTactic — tactic_lets check skipped", pf)
}
```

**File to edit:** `check/l2s.go`

---

### Fix 10: F3 — `collectAllNamedBinders` passes Stmt instead of ActionTerm

**Location:** `check/l2s.go:1004-1007`

**Problem:** Python collects named binders from `b.action` (ActionTerm) while Go collects from `b.Action.Stmt` (the inner statement). Named binders at the ActionTerm wrapper level would be missed.

**Python reference:** `ivy_l2s.py:1441`: `[b.action for b in model.bindings]`

**Fix:** Change `b.Action.Stmt` to pass the ActionTerm itself:

```go
for _, b := range model.Bindings {
    collectActionNBs(b.Action.Stmt, result)
}
```
→
```go
for _, b := range model.Bindings {
    // Python: [b.action for b in model.bindings] — collects from ActionTerm,
    // not just the inner Stmt. ActionTerm may have named binders in its
    // inputs/outputs sort annotations.
    if b.Action != nil {
        // Collect from the Stmt (the actual action code)
        collectActionNBs(b.Action.Stmt, result)
        // Also collect from ActionTerm-level inputs/outputs
        for _, inp := range b.Action.Inputs {
            for _, nb := range lu.NamedBindersAst(inp) {
                existing, _ := result.Get2(nb.Name)
                result.Set(nb.Name, append(existing, nb))
            }
        }
        for _, out := range b.Action.Outputs {
            for _, nb := range lu.NamedBindersAst(out) {
                existing, _ := result.Get2(nb.Name)
                result.Set(nb.Name, append(existing, nb))
            }
        }
    }
}
```

Actually, the simpler and more faithful approach: since Python's `ilu.named_binders_asts` walks `b.action` (ActionTerm) via its `.args` property which includes all children, we should walk the ActionTerm similarly. The ActionTerm has fields `Inputs`, `Outputs`, `Labels`, and `Stmt`. We need to collect from ALL of them, not just `Stmt`.

**File to edit:** `check/l2s.go`

---

## Implementation Order

1. **Fix 7 (D6)** — one-line change, zero risk
2. **Fix 9 (F1)** — one-line warning, zero risk
3. **Fix 8 (C1)** — small change to `continue` instead of fallback
4. **Fix 3 (D1)** — add default case to switch, straightforward
5. **Fix 6 (D4/D5)** — add `PP` field and `L2sGToGlobally`, medium
6. **Fix 4 (D2)** — add `LoopStart` field and scan logic, medium
7. **Fix 10 (F3)** — extend `collectAllNamedBinders`, low risk
8. **Fix 2 (A4)** — clarifying comment, possibly add warning
9. **Fix 1 (A1)** — fix the no-ag fallback, needs careful checker callback logic
10. **Fix 5 (D3)** — add TODO for multi-state diagnosis, low priority

## Verification

```
cd ~/ivy/goivy && make test
```

All existing tests must continue to pass. Focus on:
- `check/l2s_test.go`
- `check/l2s_hooks_test.go`
- `check/regression_test.go`
- `check/check_test.go`
- `check/check_port_test.go`
