# Fix xtrace divergence at line 146365: LocalAction.ActionClone

**Created:** 2026-03-29 (current session)

## Context

The golden test (`make golden`) diverges at xtrace line 146365:
- **Go**: `XTRACE: compiler.CheckProperties EXIT`
- **Python**: `XTRACE: LocalAction.__init__ uniqueID=649 caller=ast.LocalAction.clone`

Go exits `CheckProperties` too early, and also its `LocalAction.ActionClone` doesn't emit traces or get fresh IDs like Python's `clone` does. Two bugs combine to cause this divergence.

## Bug 1: Trace order in CheckProperties

**File:** `compiler/phase6.go` lines 2334-2335

Go emits the EXIT trace **before** calling `ApplyAssertProofsWithProver`. Python emits it **after** `apply_assert_proofs`. Fix: swap the order.

**Current:**
```go
xtracer.Trace("compiler.CheckProperties EXIT")
return ApplyAssertProofsWithProver(mod, prover)
```

**Fixed:**
```go
err := ApplyAssertProofsWithProver(mod, prover)
xtracer.Trace("compiler.CheckProperties EXIT")
return err
```

## Bug 2: LocalAction.ActionClone bypasses NewLocalAction

**File:** `actions/action.go` lines 740-777

Python's `LocalAction.clone` calls `LocalAction(*args, caller="ast.LocalAction.clone")` which goes through `__init__`, incrementing the counter and emitting XTRACE. Go's `ActionClone` creates `&LocalAction{...}` directly, copying the old `UniqueID`.

### Step 1: Add `ActCfg` field to LocalAction struct (line ~744)

```go
type LocalAction struct {
    ActionBase
    Locals   []lg.Expr
    Body     lg.Expr
    UniqueID int64
    ActCfg   *ActionsConfig // for ActionClone to call NewLocalAction
}
```

### Step 2: Set `ActCfg: cfg` in NewLocalAction (lines 747-759)

Both return paths in `NewLocalAction` must include `ActCfg: cfg`.

### Step 3: Rewrite ActionClone (lines 770-777)

```go
func (a *LocalAction) ActionClone(args []lg.Expr) Action {
    if a.ActCfg == nil {
        panic("actions: LocalAction.ActionClone called with nil ActCfg — was not created via cfg.NewLocalAction()")
    }
    r := a.ActCfg.NewLocalAction("ast.LocalAction.clone", args...)
    r.ActionBase = a.ActionBase
    return r
}
```

This matches Python's clone: fresh unique_id, XTRACE emitted, lineno copied via ActionBase struct copy.

## Files to modify

| File | Change |
|------|--------|
| `compiler/phase6.go:2334-2335` | Swap EXIT trace to after `ApplyAssertProofsWithProver` |
| `actions/action.go:740-777` | Add `ActCfg` field, set in constructor, use in `ActionClone` |

## Verification

1. `cd ~/goivy && make golden` — the divergence at 146365 should advance (either matching further or hitting a new, later divergence)
2. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — compilation check
3. `cd ~/go/src/github.com/glycerine/goivy/actions && go test -short` — existing action tests pass
