# Plan: Add CheckProperties tracing to find index.spec.prop4 misclassification

**Created:** 2026-03-31 evening

## Context

Online per-symbol tracing in `IsolateComponent` (already implemented) revealed that `index.spec.prop4` is in `LabeledAxioms[5]` in Go but `labeled_props[0]` in Python **before** `IsolateComponent` runs. The `pre_filter_*` traces confirmed this is an upstream bug in `CheckProperties` (`compiler/phase6.go:2280` / `ivy_compiler.py:2241`).

Both sides have identical classification logic:
- proved + 0 subgoals + not schema + not def → axioms (Go phase6.go:2312, Python ivy_compiler.py:2286)
- proved + subgoals → props
- no proof → props

So either Go's `pmap` has a proof entry Python's doesn't, or Go's prover returns 0 subgoals where Python's returns >0. We need per-property classification tracing to determine which.

## Implementation

### Step 1: Add per-property classification tracing to Go

**File:** `goivy/compiler/phase6.go`, inside the `for _, prop := range props` loop at line ~2280

At the top of the loop body, trace the classification decision inputs:
```
compiler.CheckProperties.classify label=<name> id=<id> temporal=<bool> hasPf=<bool>
```

At each classification outcome, trace the decision:
```
compiler.CheckProperties.classify label=<name> -> axioms (proved, 0 subgoals)
compiler.CheckProperties.classify label=<name> -> props (temporal)
compiler.CheckProperties.classify label=<name> -> props (no proof)
compiler.CheckProperties.classify label=<name> -> props (proved, N subgoals)
```

### Step 2: Add matching per-property classification tracing to Python

**File:** `pyivy/ivy/ivy/ivy_compiler.py`, inside the `for prop in props` loop at line ~2272

Same trace format as Go so line-by-line diff works.

## Files to Modify

| File | Change |
|------|--------|
| `goivy/compiler/phase6.go` | Add xtracer traces inside CheckProperties loop (~line 2280) |
| `pyivy/ivy/ivy/ivy_compiler.py` | Add xtracer traces inside check_properties loop (~line 2272) |

## Verification

Re-run the golden test. The `compiler.CheckProperties.classify` traces will show:
- Whether `index.spec.prop4` hits `hasPf=True` in Go but `hasPf=False` in Python
- Or if both have `hasPf=True`, whether subgoal counts differ
- The FIRST trace that diverges pinpoints the exact root cause
