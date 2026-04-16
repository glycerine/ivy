# PLAN: Fix Go's `GetSmallModelWithCond` to skip `decide()` when `finalCond` is empty list

Created: 2026-04-16 (afternoon, fourth pass)

## Context

After applying plan 311 (remove early-return when checkers is empty in `CheckFcsInState`/`CheckFcsInStateWithAG`) and plan 312 (typeinfer NamedBinder TopSort propagation), the user re-ran the divergence comparison `make tlb` and got a NEW divergence point further along in the trace, captured in `~/ivy/goivy/log.tlb` (lines 5057-5083 of the log file, xtrace counter 381107-381115):

```
381107  go : XTRACE: check.checkFcsNormalPath history path postFmlas=0 postDefs=0 axiomFmlas=0 axiomDefs=0 checkers=0
        py : XTRACE: check.checkFcsNormalPath history path postFmlas=0 postDefs=0 axiomFmlas=0 axiomDefs=0 checkers=0

381108  go : XTRACE: transrel.SatisfyWithCond ENTER postFmlas=0 postDefs=0 axiomFmlas=0 axiomDefs=0
        py : XTRACE: transrel.SatisfyWithCond ENTER postFmlas=0 postDefs=0 axiomFmlas=0 axiomDefs=0
... (all matching) ...
381111  go : XTRACE: ivy_solver.py:1339 get_small_model() ENTER shrink=False
        py : XTRACE: ivy_solver.py:1339 get_small_model() ENTER shrink=False

381112  go : XTRACE: solver.ClausesToZ3 ENTER fmlas=0 defs=0
        py : XTRACE: solver.ClausesToZ3 ENTER fmlas=0 defs=0

381113  go : XTRACE: ivy_solver.py:603 type_constraints() ENTER nsyms=0
        py : XTRACE: ivy_solver.py:603 type_constraints() ENTER nsyms=0

381114  go : XTRACE: solver.ClausesToZ3 EXIT exprs=0
        py : XTRACE: solver.ClausesToZ3 EXIT exprs=0

381115  go : XTRACE: ivy_solver.py:1302 decide() ENTER          ← Go calls decide()
        py : XTRACE: ast.LF.__init__ id=2105 counter=2106        ← Python skips decide(), proceeds
```

So both Python and Go take the same path through `checkFcsNormalPath` → `SatisfyWithCond` → `get_small_model` with `checkers=0` and EMPTY clauses (`fmlas=0 defs=0`). Both run `clauses_to_z3` on the empty clauses (returning empty z3 list). Then they diverge:
- **Python** does NOT call `decide()` and instead returns None from `get_small_model`. The next thing Python does is construct a labeled formula (`LF.__init__`) for the next conjecture/proof step.
- **Go** unconditionally calls `decide()` on an empty solver, emitting the spurious `decide() ENTER` trace.

### Root cause: Python distinguishes `None` from `[]`; Go does not

Python `get_small_model` (`ivy_solver.py:1422-1525`) handles `final_cond` like this:

```python
if final_cond is not None:           # final_cond=[] is NOT None → True
    if isinstance(final_cond, list): # True
        res = z3.unsat                # default
        for fc in final_cond:        # ← empty list: loop is a no-op
            ...                       # (per-checker logic with decide(s) inside)
        # falls through with res = z3.unsat
    else:                              # single-object final_cond
        s.add(clauses_to_z3(final_cond))
        res = decide(s)
else:                                  # final_cond IS None
    res = decide(s)                    # ← THIS is the only "no checkers" decide() path
if res == z3.unsat:
    return None
```

So when `final_cond=[]` (empty list), Python takes the LIST branch, the for-loop is empty, `res` stays `z3.unsat`, and the function returns `None` **without ever calling `decide()`**.

Go `GetSmallModelWithCond` (`/Users/jaten/ivy/goivy/z3bridge/solver_model.go:130-236`) collapses None and empty-list:

```go
if len(finalCond) > 0 {              // ← treats nil and []FinalCond{} identically (false in both)
    for _, fc := range finalCond {
        ...
    }
} else {
    // No final conditions: just check satisfiability.
    xtracer.Trace("ivy_solver.py:1302 decide() ENTER")
    overallResult = z3solver.Check()
}
```

Both `nil` (Go's None equivalent) and `[]FinalCond{}` (empty list) hit the `else` branch that calls `decide()`.

### Why this is the next bug in this sequence

This is the same family of bug as plan 311 — Python's verification pipeline correctly degenerates to a no-op when the checker list is empty, while Go takes a different code path that does extra work. Plan 311 made Go enter `checkFcsNormalPath` at all; this plan makes the path itself match Python's empty-list behavior.

After this fix, Go will produce no trace event between the empty `ClausesToZ3 EXIT` and the next conjecture's processing — matching Python — and the `LF.__init__ id=2105 counter=2106` event will then be the next event from BOTH sides.

### Why this matters for the original `tlb.ivy` panic

The original panic (`TopSort` has no `to_z3()` equivalent, `invar386`) shows that something later in the pipeline still has a stray TopSort. Each spurious operation Go does is an opportunity to corrupt state or to hide where the real divergence is. By fixing this empty-list mismatch, the next divergence event becomes meaningfully comparable, letting us localize the actual TopSort propagation bug.

## Fix

Two edits — one in `z3bridge` to honor the `nil`-vs-`empty-list` distinction, one at the call site that must guarantee a non-nil empty slice when no checkers were filtered in.

### Edit 1: `/Users/jaten/ivy/goivy/z3bridge/solver_model.go` (around line 152)

Change the outer condition from `len(finalCond) > 0` to `finalCond != nil`. The inner for-loop body is unchanged — it just becomes a no-op when the slice is non-nil but empty.

Before (lines 150-236, simplified):
```go
overallResult := Unsat
var assumes []*module.Clauses
if len(finalCond) > 0 {
    for _, fc := range finalCond {
        // per-checker logic ...
    }
} else {
    // No final conditions: just check satisfiability.
    xtracer.Trace("ivy_solver.py:1302 decide() ENTER")
    overallResult = z3solver.Check()
}
```

After:
```go
overallResult := Unsat
var assumes []*module.Clauses
// Python ivy_solver.py:1469-1512:
//   if final_cond is not None:    → Go: finalCond != nil
//       for fc in final_cond: ... (empty list ⇒ no-op, returns nil)
//   else:                          → Go: finalCond == nil
//       res = decide(s)
// nil finalCond ≡ Python None: caller did not request checker-driven
// verification, so we ask the solver if the state is satisfiable.
// Non-nil but empty finalCond ≡ Python []: caller had no checkers
// after filtering — there is nothing to check, so we return nil
// without calling decide(), exactly like Python.
if finalCond != nil {
    for _, fc := range finalCond {
        // per-checker logic ... (UNCHANGED)
    }
} else {
    xtracer.Trace("ivy_solver.py:1302 decide() ENTER")
    overallResult = z3solver.Check()
}
```

The body of the for-loop is unchanged. The only structural change is replacing the `len > 0` test with a nil test.

### Edit 2: `/Users/jaten/ivy/goivy/check/check.go` (around lines 644-648)

`checkFcsNormalPath` is the moral Go equivalent of Python's `check_fcs_in_state`, where Python ALWAYS passes `filter_fcs(fcs)` — a list (possibly empty). To preserve that semantics across the `SatisfyWithCond` → `SmallModelClauses` → `GetSmallModelWithCond` chain, the slice must be non-nil even when no checkers survived filtering.

Before:
```go
// Convert checkers to solver.FinalCond for history.SatisfyWithCond
var finalConds []solver.FinalCond
for _, fc := range filteredCheckers {
    finalConds = append(finalConds, fc)
}
```

After:
```go
// Convert checkers to solver.FinalCond for history.SatisfyWithCond.
// Python: filter_fcs(fcs) ALWAYS returns a list (possibly empty), never None.
// Initialize as an empty slice (non-nil) so GetSmallModelWithCond takes
// the "Python list" branch even when no checkers survived filtering.
finalConds := make([]solver.FinalCond, 0, len(filteredCheckers))
for _, fc := range filteredCheckers {
    finalConds = append(finalConds, fc)
}
```

### What we are NOT changing

- `History.Satisfy(axioms)` at `actions/transrel.go:2108-2110` continues to call `SatisfyWithCond(axioms, nil, nil)`. Passing `nil` mirrors Python's `History.satisfy(axioms)` default `final_cond=None`, which IS the case where Python wants `decide()` to be called. Edit 1 preserves this.
- The inner per-checker loop body in `GetSmallModelWithCond` is unchanged.
- The `decide()` xtrace label `ivy_solver.py:1302` is wrong (the Python function is at line 1409 today), but it is the agreed cross-language label and we do not change it here.

## Verification

**Do NOT run `goivy_check ../ivy-lang-examples/examples/liveness/tlb.ivy` from this agent — the user runs `make tlb` (~1 hour) on another machine.**

Local checks this agent will run after the user approves the plan:

1. `go build ./z3bridge/... ./check/... ./actions/...` — confirm the change compiles.
2. `go test ./z3bridge/... ./check/... ./actions/...` — confirm existing tests still pass. If a test was depending on the old empty-list-calls-decide() behavior, it was relying on the bug; investigate before changing the test.

After the user re-runs `make tlb` on the other machine:

A. **Best case**: lines 381115+ now match — both Python and Go emit `LF.__init__ id=2105 counter=2106` next. The next divergence (if any) appears further along the trace and tells us where the next bug lives.

B. **Possible case**: divergence moves to a different xtrace event but the same pattern (Go calling decide() spuriously somewhere similar). That would indicate another nil-vs-empty-list mismatch in `GetSmallModelWithCond` or its callers — apply the same pattern there.

C. **Unexpected case**: trace matches further but the original TopSort panic on `invar386` returns. That means the panic is independent of this divergence; we will have a cleaner trace context to localize it.

## Critical files

- `/Users/jaten/ivy/goivy/z3bridge/solver_model.go` — Edit 1 (line ~152)
- `/Users/jaten/ivy/goivy/check/check.go` — Edit 2 (line ~645)

## Reference (source of truth)

- `~/ivy/pyivy/ivy/ivy/ivy_solver.py:1422-1525` — `get_small_model` with the `if final_cond is not None / isinstance(list) / else` three-way branching that Go must mirror
- `~/ivy/pyivy/ivy/ivy/ivy_check.py:376-380` — `filter_fcs` always returns a list (possibly empty), never None
- `~/ivy/pyivy/ivy/ivy/ivy_check.py:382-426` — `check_fcs_in_state` and `checkFcsNormalPath` always pass `filter_fcs(fcs)` (a list) to `history.satisfy`
- `~/ivy/pyivy/ivy/ivy/ivy_transrel.py:605-607` — `small_model_clauses` forwards `final_cond` unchanged
- `~/ivy/pyivy/ivy/ivy/ivy_transrel.py:649-705` — `History.satisfy` forwards `final_cond` unchanged

## What we are NOT investigating in this plan

- The original `invar386` TopSort panic. Once this divergence is fixed and `make tlb` re-runs, the next divergence point (if any) will guide the next plan. We may discover the panic is downstream and unrelated to this empty-list bug, OR we may discover this fix is sufficient because Go was poisoning solver state with the spurious empty-decide.
- Other potential nil-vs-empty mismatches in the codebase. Address them only if they show up as divergences.
