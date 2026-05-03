# Fix TestUPDR_InductiveWithInitializer and Z3 Interpolation Panic

**Created**: 2026-05-03 22:50 UTC

## Context

Two related UPDR failures need fixing:

1. **`TestUPDR_InductiveWithInitializer`** in `tactics/updr_tactics_test.go` — compiles `blah.ivy` (a trivially inductive spec: `after init { flag(X) := true }`, `step` preserves flag, conjecture `flag(X)`). Python's `ivy_check` passes. The Go test currently `t.Skipf`s with an error from `UPDR.Apply`.

2. **Z3 interpolation panic** — when UPDR reaches `RefineOrReverse` → `ForwardInterpolant` → Z3's `ComputeInterpolant`, Z3 panics with "theory not supported by interpolation or bad proof". This crashes `ivyweb` and fails `TestSolverPDRStep`. The panic comes from `goZ3BridgeErrorHandler` (z3bridge/quantifier.go:83) which unconditionally `panic()`s on any Z3 error.

## Root Causes

### Cause A: `GetBigAction` uses `EnvAction` instead of `ChoiceAction`

**Python** (`tactics_api.py:71-80`):
```python
def get_big_action():
    exported_action_names = [e.exported() for e in _ivy_ag.exports]
    exported_actions = [_ivy_ag.actions[k] for k in exported_action_names]
    result = ChoiceAction(*exported_actions)
    result.label = ' + '.join(exported_action_names)
    return result
```

**Go** (`tactics/tactics.go:362-375`):
```go
func GetBigAction(ag *art.AnalysisGraph) actions.Action {
    for name := range ag.PublicActions.All() { ... }
    return actions.NewEnvActionOn(ag.Domain.Cfg.ActCfg, branches...)
}
```

Divergences:
- Python uses `ChoiceAction` (plain nondeterministic choice). Go uses `NewEnvActionOn` which wraps in `EnvAction` (adds parameter havoc/skolemization). Different action types produce different transition relation updates.
- Python iterates `_ivy_ag.exports` and calls `.exported()`. Go iterates `ag.PublicActions`. After `CreateIsolate`, `PublicActions` contains the same names as exports, so the name set is equivalent — but Go should also set a label on the result for parity.

### Cause B: Z3 error handler panics unconditionally

`z3bridge/quantifier.go:83`:
```go
panic("z3 bridge panic on error: " + C.GoString(msg))
```

When `Z3_compute_interpolant` encounters an unsupported theory (e.g., uninterpreted functions), Z3 internally triggers the error handler callback, which panics. This panic propagates through CGo and crashes the process. No recover exists in the call chain.

The error propagation path is:
```
goZ3BridgeErrorHandler (panic) →
  Z3Context.ComputeInterpolant (no recover) →
    computeZ3Interpolant →
      Solver.BinaryInterpolant →
        actions.Interpolant →
          actions.ForwardInterpolant →
            TacticsContext.RefineOrReverse →
              UPDR.Apply → crash
```

## Fixes

### Fix 1: Change `GetBigAction` to use `ChoiceAction`

**File**: `tactics/tactics.go:362-375`

Change `NewEnvActionOn` to `NewChoiceActionOn` to match Python's `ChoiceAction`. Also iterate `ag.Exports` (matching Python's `_ivy_ag.exports`) and set a label.

```go
func GetBigAction(ag *art.AnalysisGraph) actions.Action {
    var names []string
    var branches []lg.Expr
    for _, e := range ag.Exports {
        name := e.Exported()
        if act, ok := ag.Actions.Get2(name); ok {
            if a, ok := act.(actions.Action); ok {
                names = append(names, name)
                branches = append(branches, a)
            }
        }
    }
    if len(branches) == 0 {
        return actions.NewSequence()
    }
    result := actions.NewChoiceActionOn(ag.Domain.Cfg.ActCfg, branches...)
    result.SetLabel(strings.Join(names, " + "))
    return result
}
```

Check whether `ChoiceAction` has a `SetLabel` method or a `Label` field; if not, add one or use the `ActionBase.Label` field.

### Fix 2: Recover Z3 panics in `ComputeInterpolant`

**File**: `z3bridge/interp.go:51-79`

Add `defer recover()` inside the closure passed to `ctx.do()` to catch the `goZ3BridgeErrorHandler` panic and convert it to a returned error:

```go
func (ctx *Z3Context) ComputeInterpolant(pattern Expr) ([]Expr, error) {
    var result []Expr
    var resErr error

    ctx.do(func() {
        defer func() {
            if r := recover(); r != nil {
                resErr = fmt.Errorf("Z3 interpolation failed: %v", r)
            }
        }()
        // ... existing code unchanged ...
    })
    return result, resErr
}
```

This converts the crash into an error that propagates cleanly:
- `computeZ3Interpolant` returns the error
- `BinaryInterpolant` returns `nil, err`
- `Interpolant` returns `nil` (line 1488: `if err != nil || itp == nil { return nil }`)
- `ForwardInterpolant` returns `nil`
- `RefineOrReverse` sees `x == nil`, creates backward image goal instead of crashing

### Fix 3: Verify the UPDR happy path works

For `blah.ivy`, the UPDR should succeed through **fact propagation** (RecalculateFactsFn) without ever reaching interpolation:

1. Frame 0: init state with `flag(X) = true` (from AddInitialState applying the initializer)
2. Frame 1: execute step → `flag(n) := flag(n)` preserves flag
3. `RecalculateFactsFn` pushes `flag(X)` from frame 0 to frame 1 via `ForwardImage` + `ImpliedFacts`
4. `RefutedGoal` detects `¬flag(X)` is inconsistent with frame 1 → goal removed
5. `Cover(frame1, frame0)` succeeds → UPDR returns `true`

If fact propagation doesn't work after Fix 1, investigate:
- Whether `AddInitialState` produces correct post-init clauses (verify `flag(X)` is in `initState.Clauses`)
- Whether `ForwardImage` of the init state through the step action preserves `flag(X)`
- Whether `ImpliedFacts` correctly uses Z3 to detect implication

## Critical Files

| File | Change |
|------|--------|
| `tactics/tactics.go:362-375` | Fix 1: `GetBigAction` ChoiceAction |
| `z3bridge/interp.go:51-79` | Fix 2: recover in ComputeInterpolant |
| `tactics/updr_tactics_test.go` | Test file (no changes needed) |
| `webui/session_solver_test.go` | Secondary test that also benefits from Fix 2 |

## Verification

1. `cd ~/ivy/goivy && make test` — run the full test suite
2. Specifically check `TestUPDR_InductiveWithInitializer` passes (not just skips)
3. Specifically check `TestSolverPDRStep` no longer panics (it may still fail on interpolation quality but should not crash)
4. Manual: `ivyweb` should not crash when running UPDR on specs with uninterpreted functions
