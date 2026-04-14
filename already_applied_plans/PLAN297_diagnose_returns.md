# Plan: Fix L2S instrStmt Monitoring Check + SplitReturns Config (Round 11)

Created: 2026-04-14 ~08:15 UTC

## Context

After Round 10 eradicated all `changed` clone optimizations, `make golden` advanced ~800 lines. The new divergence at line 721907:

```
721906  go : CallAction.__init__ uniqueID=773 counter=774
        py : CallAction.__init__ uniqueID=773 counter=774
721907  go : l2s.SharedStep7 instrStmt mods=[] deps=[]
        py : LocalAction.__init__ uniqueID=1363 caller=actions.CallAction.action_update
```

Python's `instr_stmt` detects a CallAction as having monitored returns and calls `split_returns()`, which creates a `LocalAction`. Go's `instrStmt` does NOT detect monitoring and falls through to the general ActionClone + mods/deps path.

## Root Cause Analysis

### Bug 1 (Primary): Monitoring check fails to match returns

Go's check at `check/l2s_shared.go:516-523`:
```go
for _, r := range returns {
    if c, ok := r.(*lg.Const); ok {  // only checks *lg.Const
        k := c.Name
        if len(symprops[k]) > 0 || ...
```

Python's check at `ivy_l2s.py:1228-1230`:
```python
actual_returns = stmt.args[1:]
if any(sym in symprops or sym in symwhens or sym in symwaits for sym in actual_returns):
```

Key Python facts discovered:
- `ivy_logic.py:130`: `Symbol = lg.Const`
- `ivy_logic.py:132`: `Symbol.rep = property(lambda self: self)` (Const's rep is itself)
- `ivy_logic.py:139-140`: `Const.__call__(*terms)` returns `self` for 0 args (NOT Apply wrapping)
- `ivy_logic.py:313-318`: `is_app()` returns True for `lg.Const`
- `symbols_ilu_ast` yields `ast.rep` for app nodes; for bare Const, rep=self, so it yields the Const

Both `symprops` keys and `actual_returns` are bare `Const` objects in Python. The check uses structural equality (name+sort).

In Go, both should be `*lg.Const` (compiler `CompileApp` returns bare `*lg.Const` for 0-arity, line 798). The check uses string name comparison. **Both should work identically.** Since they don't, the root cause must be one of:
- A return is NOT `*lg.Const` at runtime (possibly transformed after compilation)
- A name mismatch between symprops keys and returns

**Diagnostic traces are needed to pinpoint the exact cause.**

### Bug 2 (Confirmed): SplitReturns uses fresh NewActionsConfig()

`check/l2s_shared.go:526`:
```go
split := call.SplitReturns(actions.NewActionsConfig())
```

`NewActionsConfig()` creates a fresh config with counters at 0. The CallAction (uid=0) and LocalAction (uid=0) created inside SplitReturns will diverge from Python's values (uid=773, uid=1363). Fix: use `call.ActCfg` (the session config stored on every action).

## Execution Plan

### Step 1: Add diagnostic traces to monitoring check

In `check/l2s_shared.go`, add traces inside `instrStmt` at the CallAction check to log:
- CallAction identity (UniqueID, callee name)
- Each return's Go type, name/value, and whether it matched symprops/symwhens/symwaits
- Keys present in symprops/symwhens/symwaits (count + sample)

Also add matching trace in Python `ivy_l2s.py:1227-1231` for side-by-side comparison.

### Step 2: Fix SplitReturns config bug

Change `check/l2s_shared.go:526`:
```go
// BEFORE:
split := call.SplitReturns(actions.NewActionsConfig())
// AFTER:
split := call.SplitReturns(call.ActCfg)
```

`call.ActCfg` is guaranteed non-nil (set by `NewCallActionOn`, which panics if nil).

### Step 3: Harden monitoring check for non-Const returns

Expand the type check to also handle `*lg.Apply` returns (extract function symbol name):
```go
for _, r := range returns {
    var k string
    switch v := r.(type) {
    case *lg.Const:
        k = v.Name
    case *lg.Apply:
        if c, ok := v.Func.(*lg.Const); ok {
            k = c.Name
        }
    }
    if k != "" && (len(symprops[k]) > 0 || len(symwhens[k]) > 0 || len(symwaits[k]) > 0) {
        monitored = true
        break
    }
}
```

This defensively handles both Const and Apply returns. If diagnostics show a different cause (name mismatch), adjust accordingly.

### Step 4: Build + golden test

1. `cd ~/ivy/goivy && go build ./...`
2. `make golden`
3. Check log.red for resolved divergence or new ones

## Files to modify

1. `/Users/jaten/ivy/goivy/check/l2s_shared.go` -- monitoring check (lines 511-530), SplitReturns call (line 526)
2. `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_l2s.py` -- add matching diagnostic traces (lines 1227-1231)

## Verification

1. `go build ./...` compiles clean
2. `make golden` -- line 721907 divergence resolved; check for new divergences
3. Diagnostic traces can be removed once root cause is confirmed and fixed
