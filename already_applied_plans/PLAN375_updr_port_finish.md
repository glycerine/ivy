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

## Critical Files

| File | Role |
|------|------|
| `tactics/tactics.go` | Main file: GetDiagram, RefineOrReverse, UPDR.Apply, tactic structs |
| `iupdr/iupdr.go:442` | One-line signature update for GetDiagram |
| `actions/transrel.go:1501` | ForwardInterpolant — already complete, called by fix |
| `z3bridge/solver_clauses.go:152` | ClausesModelToDiagramFull — already complete, called by fix |
| `proof/proofstate.go:13` | ProofGoal.Parent — already exists, used by fix |

## Verification

Run `cd ~/ivy/goivy && make test` to verify no regressions. The existing `updr_tactics_test.go` integration test (currently skipped) should be un-skipped and extended to cover:
1. GetDiagram returns non-nil diagram for satisfiable goals
2. RefineOrReverse returns `*module.Clauses` interpolant on refinement
3. RefineOrReverse returns `*proof.ProofGoal` backward image on reversal
4. Parent chain traversal removes all refuted goals
5. UPDR finds invariant on a simple example (client_server_sorted.ivy)
