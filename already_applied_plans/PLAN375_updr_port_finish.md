# Plan: Complete the UPDR Port from Python to Go

**Created:** 2026-05-03 ~UTC

## Context

The Go UPDR tactic in `tactics/tactics.go` is ~70% complete but has several critical bugs and missing steps compared to the Python source of truth at `pyivy/ivy/ivy/tactics.py:220-292` and `pyivy/ivy/ivy/tactics_api.py:264-341`. The webui calls `runUPDR()` which invokes `UPDR.Apply()`, but the algorithm produces incorrect results because: (1) `GetDiagram` is a stub, (2) `RefineOrReverse` does not compute a Craig interpolant, (3) the UPDR main loop skips the diagram generalization step and the propagation phase, and (4) parent-chain traversal is broken in goal removal.

All low-level building blocks are already ported (ForwardInterpolant, ClausesModelToDiagramFull, IsSkolem, RecalculateFactsFn, etc.). The work is wiring them into the tactic layer.

## Changes

All changes are in `~/ivy/goivy/tactics/tactics.go` except one line in `~/ivy/goivy/iupdr/iupdr.go`.

### 1. Fix `GetDiagram()` — replace stub with real implementation

**File:** `tactics/tactics.go:231-239`
**Python:** `tactics_api.py:323-333`

Add `weaken bool` parameter. Implement by calling `z3bridge.Solver.ClausesModelToDiagramFull()` with the combined node clauses + goal formula, `actions.IsSkolem` as the ignore function, `FalseClauses` as implied, and the background theory as axioms. Return `GoalAtArgNode(d.ToFormula(), node)`.

Existing building blocks:
- `z3bridge.Solver.ClausesModelToDiagramFull()` at `z3bridge/solver_clauses.go:152`
- `actions.IsSkolem()` at `actions/transrel.go:75`

### 2. Fix `RefineOrReverse()` — use ForwardInterpolant instead of Implies

**File:** `tactics/tactics.go:248-295`
**Python:** `tactics_api.py:286-307`

Replace the `slv.Implies()` check with `actions.ForwardInterpolant(tc.Mod, pred.Clauses, update, goalClauses, axioms, nil)`. On success (non-nil result), return `(true, result.Itp)` where `Itp` is a `*module.Clauses`. On failure, compute backward image and return `(false, newProofGoal)`.

Existing building blocks:
- `actions.ForwardInterpolant()` at `actions/transrel.go:1501`
- `actions.InterpolantResult` at `actions/transrel.go:1463`

### 3. Fix `RefineOrReverseTactic.Apply()` — parent chain traversal

**File:** `tactics/tactics.go:405-410`
**Python:** `tactics.py:58-64`

Change `g = t.TC.TopGoal()` to `g = g.Parent`. Also change `result.(lg.Expr)` to `result.(*module.Clauses)` since RefineOrReverse now returns `*module.Clauses` on refinement (no more `FormulaToClauses` conversion needed).

### 4. Fix `CustomRefineOrReverse()` — parent chain traversal

**File:** `tactics/tactics.go:844-852`

Replace `break` with `g = g.Parent`. Remove the stale comment claiming Parent is "not yet available" (it is — `proof.ProofGoal.Parent` at `proof/proofstate.go:13`).

### 5. Rewrite `UPDR.Apply()` — match Python algorithm exactly

**File:** `tactics/tactics.go:431-530`
**Python:** `tactics.py:222-292`

Five sub-changes:

**5a.** Use `tc.ArgAddActionNode(lastFrame, action, nil)` instead of `tc.AG.Execute(...)` for frame creation (line 466). This matches Python line 245 and correctly sets `node.Clauses = TrueClauses()`.

**5b.** After pushing bad states goal (line 477), add `RecalculateFactsFn` call to propagate predecessor's conjuncts to the new frame. Matches Python line 250:
```
recalculate_facts(last_frame, ta.arg_get_conjuncts(arg_get_pred(last_frame)))
```

**5c.** In the inner goal loop, add `push_diagram` + `top_goal` before `refine_or_reverse`. Matches Python lines 266-270:
```
push_diagram(current_goal, False)
dg = top_goal()
if refine_or_reverse(dg, False):
```

**5d.** Remove the manual goal removal loop at lines 510-516. The Python UPDR uses `auto_remove=False` in its `refine_or_reverse` call, relying on the outer loop's `remove_if_refuted` instead. The Go `RefineOrReverseTactic.Apply(dg)` with `AutoRemove=false` handles adding facts/pushing goals internally.

**5e.** Add propagation phase after the inner loop. Matches Python lines 284-288:
```
for i in range(1, len(frames)):
    facts_to_check = set(arg_get_conjuncts(frames[i-1])) - set(arg_get_conjuncts(frames[i]))
    recalculate_facts(frames[i], list(facts_to_check))
```

**5f.** Register the big action in AG.Actions. Matches Python line 226:
```
ta._ivy_ag.actions[repr(action)] = action
```

### 6. Update `GetDiagram` callers for new `weaken` parameter

- `PushDiagram.Apply` (tactics.go:585): pass `t.Weaken`
- `iupdr/iupdr.go:442`: pass `false`

### 7. Add helper functions

Add `exprsToClausesList` and `setDifference` as file-local helpers in `tactics.go` (same logic as `iupdr/iupdr.go:597-656`).

### 8. Fix z3bridge NumeralToZ3 context leak (root cause of Z3 sort mismatch panic)

**File:** `z3bridge/solver_encoding.go:377-394` and `z3bridge/translate.go:185-192`
**Red test:** `tactics/updr_tactics_test.go:TestUPDR_InductiveWithInitializer`

**Root cause:** `Translator.Numeral()` at translate.go:187 calls `t.s.NumeralToZ3(num)` where `t.s` is the original solver. Inside `NumeralToZ3`, line 379 does `ctx := s.tr.Ctx` — this is the **original solver's Z3Context**, not the interpolation translator's context. When called during `BinaryInterpolant`, the constant "0:node" gets created in the original context while `flag` was declared in the interpolation context → Z3 detects the cross-context sort mismatch.

**The crash path:**
1. `NegateClauses(safetyProp)` produces `~flag(0)` where `0` is a numeral constant of sort `node`
2. `BinaryInterpolant` creates a fresh interpolation Z3Context
3. First `ClausesToZ3` translates the forward image (definitions for `flag`, `__flag`) — declares `flag` and `node` in the interpolation context
4. Second `ClausesToZ3` translates `~flag(0)`:
   - `flag` found in z3_predicates cache (interpolation context) ✓
   - `0` goes through `isNumeralName("0")=true` → `t.Numeral("0", nodeSort)` → `t.s.NumeralToZ3(...)` → creates `ctx.Const("0:node", translated)` using the **original** context ✗
   - `flag(0)` applies a function from context A to a constant from context B → panic

**Fix:** `Translator.Numeral()` must handle the uninterpreted-sort case directly using `t.Ctx` and `t.TranslateSort()`, bypassing `t.s.NumeralToZ3()`. For native sorts (int/bv/string), `NumeralToZ3` also uses `s.tr.Ctx`, so those must also be translated in the caller's context. The cleanest approach: duplicate the `NumeralToZ3` logic into `Translator.Numeral()` using `t.Ctx` throughout. Keep `NumeralToZ3` for non-interpolation callers that use the solver's own translator.

## Critical Files

| File | Role |
|------|------|
| `z3bridge/translate.go:185-192` | `Translator.Numeral()` — must use `t.Ctx` not `t.s.tr.Ctx` |
| `z3bridge/solver_encoding.go:377` | `NumeralToZ3` — the method that uses the wrong context |
| `tactics/tactics.go` | Main file: GetDiagram, RefineOrReverse, UPDR.Apply, tactic structs |
| `iupdr/iupdr.go:442` | One-line signature update for GetDiagram |
| `actions/transrel.go:1501` | ForwardInterpolant — already complete, called by fix |
| `z3bridge/solver_clauses.go:152` | ClausesModelToDiagramFull — already complete, called by fix |
| `proof/proofstate.go:13` | ProofGoal.Parent — already exists, used by fix |

## Verification

Run `cd ~/ivy/goivy && make test` to verify no regressions. The existing `TestUPDR_InductiveWithInitializer` (tactics/updr_tactics_test.go) is the red test — it panics at the `ForwardInterpolant` call. After fixing the z3bridge context leak (step 8), this test should pass (or at least not panic), allowing the UPDR algorithm changes (steps 1-7) to be exercised.
