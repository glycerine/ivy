# PLAN: Fix tlb divergence — remove early-return guards around CheckFcsInStateWithAG

Created: 2026-04-16 (afternoon)

## Context

The `make tlb` regression test (TestIvyTlbModel) diverges from the Python source-of-truth at line 381107 of `~/ivy/goivy/log.tlb`. After both implementations log:

```
XTRACE: check/isolate_check.go: CheckIsolate back from fragment.CheckFragment().
```

the next lines diverge:

- **Python**: `XTRACE: check.checkFcsNormalPath history path postFmlas=0 postDefs=0 axiomFmlas=0 axiomDefs=0 checkers=0`
- **Go**:     `XTRACE: ast.LF.__init__ id=2105 counter=2106`

### Root cause

Python `ivy_check.py:561-572` unconditionally calls `check_fcs_in_state(mod, ag, pre, fcs)` even when `fcs == []`. Inside `check_fcs_in_state` (`ivy_check.py:382-426`) it always builds `history`, fetches `axioms = im.module.background_theory()`, and emits the `check.checkFcsNormalPath history path …` xtrace before calling `history.satisfy(axioms, gmc, [])`.

The Go port short-circuits this in two places when there are zero checkers, and therefore never emits the xtrace and falls through to a later step (`ApplyConjProofs` on `isolate_check.go:222`) which creates a `LabeledFormula` (id=2105) — appearing as `ast.LF.__init__` in the trace. Python takes the same `apply_conj_proofs` step too, but only after the trace point fires; that's why Go's trace shows the LF first.

The two offending guards in Go:

1. **Outer guard** at `~/ivy/goivy/check/isolate_check.go:174-176`:
   ```go
   if len(checkers) > 0 {
       CheckFcsInStateWithAG(mod, ag, pre, checkers)
   }
   ```
   Python has no such guard — it calls `check_fcs_in_state` unconditionally.

2. **Inner early-return** at `~/ivy/goivy/check/check.go:477-479`:
   ```go
   func CheckFcsInStateWithAG(...) bool {
       if len(checkers) == 0 {
           return true
       }
       ...
   }
   ```
   Python's `check_fcs_in_state` has no early return on empty `fcs`.

Both are "premature optimizations" that are not present in the source-of-truth Python — they violate the mechanical-port rule (CLAUDE.md §B.9).

## Fix

### Edit 1: `/Users/jaten/ivy/goivy/check/check.go`

Remove the early-return at lines 477–479 of `CheckFcsInStateWithAG`. The function must run the full body even when `checkers` is empty, exactly as Python does.

Before:
```go
func CheckFcsInStateWithAG(mod *module.Module, ag *art.AnalysisGraph, post *art.State, checkers []Checker) bool {
    if len(checkers) == 0 {
        return true
    }

    // Get history and background theory
    var history *actions.History
    ...
```

After:
```go
func CheckFcsInStateWithAG(mod *module.Module, ag *art.AnalysisGraph, post *art.State, checkers []Checker) bool {
    // Python check_fcs_in_state has NO early return for empty fcs.
    // Even with zero checkers it builds history, fetches background_theory,
    // emits the checkFcsNormalPath xtrace, and calls history.satisfy(...).

    // Get history and background theory
    var history *actions.History
    ...
```

### Edit 2: `/Users/jaten/ivy/goivy/check/isolate_check.go`

Remove the `if len(checkers) > 0` guard at lines 174–176 so the call always runs, matching Python `ivy_check.py:572`.

Before:
```go
if len(checkers) > 0 {
    CheckFcsInStateWithAG(mod, ag, pre, checkers)
}
```

After:
```go
// Python ivy_check.py:572 calls check_fcs_in_state unconditionally,
// even when fcs is empty.
CheckFcsInStateWithAG(mod, ag, pre, checkers)
```

## What we are NOT changing

- `ag.Add(pre, nil)` at `isolate_check.go:156` is also extra vs. Python (Python builds `pre` but does not add it to the AG). The trace matches up to and including the divergence point with this call present, so it is not contributing to the bug. Leave it alone for this plan and revisit only if a later divergence implicates it.
- The body of `checkFcsNormalPath` already emits the correct xtrace inside `if history != nil`. Since `ag` and `post` are non-nil at the call site we care about, `history` will be non-nil and the trace will fire.

## Critical files

- `/Users/jaten/ivy/goivy/check/check.go` (lines 476–495: `CheckFcsInStateWithAG`)
- `/Users/jaten/ivy/goivy/check/isolate_check.go` (lines 144–186: property-checking block in `CheckIsolate`)

## Reference (source of truth)

- `~/ivy/pyivy/ivy/ivy/ivy_check.py:561-572` — unconditional call to `check_fcs_in_state`
- `~/ivy/pyivy/ivy/ivy/ivy_check.py:382-426` — `check_fcs_in_state`, no early return; emits the xtrace at line 422

## Verification

1. From `/Users/jaten/ivy/goivy`, run `make tlb`.
2. Expected: the divergence at line 381107 is resolved. Go should now emit `XTRACE: check.checkFcsNormalPath history path postFmlas=0 postDefs=0 axiomFmlas=0 axiomDefs=0 checkers=0` (matching Python), then continue with `transrel.SatisfyWithCond ENTER …` traces.
3. Re-inspect `~/ivy/goivy/log.tlb` and confirm the next divergence (if any) is much later than line 381107 — ideally the model passes through the previously-blocked region.
4. If a new divergence appears further on, that is a separate downstream bug — file a follow-up plan rather than re-doing this one.
