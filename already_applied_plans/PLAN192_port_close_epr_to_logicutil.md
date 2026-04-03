# Fix: Go fragment checker missing assumes from public actions

**Created:** 2026-04-02, **Updated:** 2026-04-03

## Context

Go's `findCycle()` at `fragment.go:749` reports a cycle that Python does not. Diagnostic tracing revealed:
- Go has 120 assumes, Python has 127 (7 missing)
- Go has 346 universally quantified variables, Python has 4599

The root cause is in `GetAssumesAndAsserts` / `makeFmlaPairFromAction`.

## Root Cause

### Bug 1: Go discards the `post` (PreNode) formula

**Python** (`ivy_fragment.py:566-577`) adds **TWO assumes per public action**:
```python
triple = action.update(im.module, [])
foo = ilu.close_epr(ilu.clauses_to_formula(triple[1]))   # TR
assumes.append((foo, action))
foo = ilu.close_epr(ilu.clauses_to_formula(triple[2]))   # Pre
assumes.append((foo, action))
```

**Go** (`fragment.go:1005-1025`) adds only **ONE** — discards `post`:
```go
pre := upd.TRNode()
post := upd.PreNode()
_ = post  // <-- BUG: discarded!
return fmlaPair{fmla: pre, source: action}, true
```

### Bug 2: Go doesn't apply `close_epr`

Python applies `ilu.close_epr()` to each action formula. This distributes `ForAll` over `And` conjuncts:
```python
def close_epr(fmla):
    if isinstance(fmla, And):
        return And(*[close_epr(f) for f in fmla.args])
    variables = list(used_variables_ast(fmla))
    if variables == []:
        return fmla
    else:
        return ForAll(variables, fmla)
```

Go has no equivalent — `makeFmlaPairFromAction` returns the raw open formula.

While `createStratMap` later applies `CloseFormula`, this wraps the entire `And(...)` in a single `ForAll` rather than distributing per-conjunct. This is semantically different for the stratification analysis.

## Changes

### 1. Port `close_epr` to Go

**File:** `~/go/src/github.com/glycerine/ivy/goivy/logicutil/ivy_logic_utils.go` (or appropriate file in logicutil/)

Python source: `~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:108-121`

```go
// CloseEpr distributes ForAll over And conjuncts, universally
// quantifying each conjunct's free variables independently.
// Corresponds to Python close_epr.
func CloseEpr(fmla lg.Expr) lg.Expr {
    if and, ok := fmla.(*lg.And); ok {
        terms := make([]lg.Expr, len(and.Terms))
        for i, t := range and.Terms {
            terms[i] = CloseEpr(t)
        }
        return &lg.And{Terms: terms}
    }
    fvs := FreeVariablesList(fmla)
    if len(fvs) == 0 {
        return fmla
    }
    return &lg.ForAll{Variables: fvs, Body: fmla}
}
```

### 2. Fix `makeFmlaPairFromAction` to return both formulas with `close_epr`

**File:** `~/go/src/github.com/glycerine/ivy/goivy/fragment/fragment.go`

Change signature to return a slice (or two fmlaPairs):
```go
func makeFmlaPairsFromAction(action interface{}, m *mod.Module) []fmlaPair {
    act, ok := action.(actions.Action)
    if !ok {
        return nil
    }
    ctx := &actions.UpdateContext{Domain: m, ActCfg: m.Cfg.ActCfg}
    xtracer.Trace("fragment calling GetUpdate type=%s", actions.ActionTypeName(act))
    upd := actions.GetUpdate(act, ctx)
    if upd == nil {
        return nil
    }
    pre := lu.CloseEpr(upd.TRNode())
    post := lu.CloseEpr(upd.PreNode())
    return []fmlaPair{
        {fmla: pre, source: action},
        {fmla: post, source: action},
    }
}
```

### 3. Update callers in `GetAssumesAndAsserts`

**File:** `~/go/src/github.com/glycerine/ivy/goivy/fragment/fragment.go` (lines 871-888)

Both the `precondsOnly` and normal branches need to use the new function:
```go
// Before (precondsOnly):
if fp, ok := makeFmlaPairFromAction(action, m); ok {
    assumes = append(assumes, fp)
}

// After:
fps := makeFmlaPairsFromAction(action, m)
assumes = append(assumes, fps...)

// Same change for the non-precondsOnly branch
```

## Key Files

- `~/go/src/github.com/glycerine/ivy/goivy/fragment/fragment.go:867-964` — `GetAssumesAndAsserts` and `makeFmlaPairFromAction`
- `~/go/src/github.com/glycerine/ivy/goivy/logicutil/` — where `CloseEpr` should be added (near existing `FreeVariablesList`)
- `~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:108-121` — Python `close_epr` (source of truth)
- `~/ivy/pyivy/ivy/ivy/ivy_fragment.py:565-577` — Python `get_assumes_and_asserts` action handling

## Verification

1. Build Go: `go build ./fragment/`
2. Run the test case with xtracer
3. Confirm `assumes` count matches Python (127)
4. Confirm `univQuantVars` count is closer to Python (4599)
5. Confirm Go no longer falsely reports a cycle
