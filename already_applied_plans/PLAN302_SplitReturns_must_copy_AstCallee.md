# Plan: Fix `SplitReturns` not copying `AstCallee` on new CallAction

Created: 2026-04-14 ~23:15 UTC

## Context

**Divergence**: `make golden` still diverges at xtrace line 736175 (in `TestOrdLive`). After the `IsFinite` fix, the divergence moved deeper: the first `stmts` children now match (the l2s_d assignments are gone), but deep inside the `cfabric.step` ActionTerm's body, a `callAction` element for `ext:lclock.next(ref.lt)` differs:

- **Go**: `(Apply func:(Symbol name:ext:lclock.next sort:(TopSort)) terms:[(Symbol name:ref.lt ...)])`
- **Python**: `(atom rep:"ext:lclock.next" terms:[(Symbol name:ref.lt ...)] aSort:nil)`

Go produces an `Apply(Symbol(...))` (compiled form). Python produces an `Atom(...)` (AST form).

**Root cause**: `CallAction.SplitReturns()` in `actions/action.go:824` creates a new CallAction but does not copy `AstCallee`:

```go
newCall := NewCallActionOn(actCfg, a.Callee, newReturns...)
newCall.ActionBase = a.ActionBase
// MISSING: newCall.AstCallee = a.AstCallee
```

Python's `split_returns` at `ivy_actions.py:1439` uses `self.clone([self.args[0]] + new_returns)` which preserves `self.args[0]` (the Atom) as the first argument.

**How it triggers**: `SplitReturns` is called from `l2s_shared.go:600` during Step 7 (InstrumentActions) for monitored actions. `lclock.next` is a monitored action (logged at line 163 of `log.red`), so its CallAction inside `cfabric.step` goes through SplitReturns. The new CallAction loses AstCallee, causing `CallAction.Sexp()` to fall through to the `sliceSexp(a.ActionArgs())` path which emits the compiled `Apply(...)` form instead of the AST `(atom ...)` form.

## Fix

### 1. Copy `AstCallee` in `SplitReturns` — `actions/action.go:825`

Add after line 825 (`newCall.ActionBase = a.ActionBase`):
```go
newCall.AstCallee = a.AstCallee
```

## Critical files to modify

1. `actions/action.go` — add one line to `SplitReturns`

## Verification

1. `go build ./...` compiles clean
2. `make golden` — divergence advances past line 736175
