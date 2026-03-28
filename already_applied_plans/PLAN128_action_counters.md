# Delete localActionCtr Global — Thread ActionsConfig for LocalAction Counter

**Created:** 2026-03-28

## Context

Golden xtrace diverges at line 127167: `LocalAction.__init__ uniqueID=0` (Go) vs `uniqueID=587` (Python). Python has ONE `local_action_ctr` global incremented by all `LocalAction.__init__` calls (parsing + compilation + transformation). Go has two separate counters: `AstConfig.LocalActionCtr` (parsing) and `actions.localActionCtr` (compilation, package-level global starting at 0).

Per CLAUDE.md, package-level mutable state is forbidden. The `localActionCtr` global (action.go:748) must go.

## Plan

### Step 1: Add `ActCfg *ActionsConfig` to `module.Config`

**File:** `module/config.go`

Add field to the Config struct:
```go
ActCfg *actions.ActionsConfig
```

Note: module already imports actions (transitively via the Action type alias). If there's an import cycle, ActionsConfig can be moved to module or use an interface.

Initialize in the Config constructor (or wherever Config is created).

### Step 2: Seed ActionsConfig.LocalActionCtr from AstConfig at IvyCompile start

**File:** `compiler/ivy_compile.go` — at the start of `IvyCompile` (~line 68)

```go
if mod.Cfg != nil && mod.Cfg.AstCfg != nil && mod.Cfg.ActCfg != nil {
    mod.Cfg.ActCfg.LocalActionCtr = int64(mod.Cfg.AstCfg.LocalActionCtr)
}
```

This ensures compilation-phase uniqueIDs continue from where parsing left off.

### Step 3: Delete `localActionCtr` global and package-level `NewLocalAction`

**File:** `actions/action.go`

- Delete `var localActionCtr int64` (line 748)
- Delete `func NewLocalAction(args ...lg.Expr) *LocalAction` (lines 750-761)
- Keep `func (cfg *ActionsConfig) NewLocalAction(args ...lg.Expr) *LocalAction` (lines 763+) — this becomes the ONLY way to create LocalAction

### Step 4: Add `actCfg()` helper to Compiler

**File:** `compiler/compiler.go`

```go
func (c *Compiler) actCfg() *actions.ActionsConfig { return c.Module.Cfg.ActCfg }
```

### Step 5: Fix compiler callers (8 sites)

All use `c.actCfg().NewLocalAction(args...)`:

| File | Line | Current | New |
|------|------|---------|-----|
| compiler/action.go | ~685 | `actions.NewLocalAction(localArgs...)` | `c.actCfg().NewLocalAction(localArgs...)` |
| compiler/action.go | ~956 | `actions.NewLocalAction(sym, asgn)` | `c.actCfg().NewLocalAction(sym, asgn)` |
| compiler/action.go | ~976 | `actions.NewLocalAction(args...)` | `c.actCfg().NewLocalAction(args...)` |
| compiler/action.go | ~1022 | `actions.NewLocalAction(args...)` | `c.actCfg().NewLocalAction(args...)` |
| compiler/compiler.go | ~98 | `actions.NewLocalAction(args...)` | `c.actCfg().NewLocalAction(args...)` |
| compiler/phase6.go | ~752 | `actions.NewLocalAction(lsym, seq)` | `c.actCfg().NewLocalAction(lsym, seq)` |

### Step 6: Fix actions-package callers (4 sites)

**IfAction.subactionsSome** (action.go:520) — called from `Subactions()` (line 425), called from `intUpdateWithSubactions` (update.go:1518) which has `ctx.ActCfg`:
- Add `actCfg *ActionsConfig` parameter to `Subactions()` and `subactionsSome()`
- Caller in update.go passes `ctx.ActCfg`
- Change `NewLocalAction(localArgs...)` → `actCfg.NewLocalAction(localArgs...)`

**CallAction.SplitReturns** (action.go:732) — only called from tests:
- Add `actCfg *ActionsConfig` parameter
- Change `NewLocalAction(localArgs...)` → `actCfg.NewLocalAction(localArgs...)`

**expandWhile** (match.go:521) — receives `*module.Module`:
- Change `NewLocalAction(auxVar, res)` → `mod.Cfg.ActCfg.NewLocalAction(auxVar, res)`

**WhileAction.Expand** (update.go:1730) — has `ctx.ActCfg`:
- Change `NewLocalAction(rankLocal, result)` → `ctx.ActCfg.NewLocalAction(rankLocal, result)`

### Step 7: Fix tests

Update test callers to create an `ActionsConfig` and use `cfg.NewLocalAction()`.

## Files to Modify

1. `module/config.go` — add `ActCfg` field
2. `compiler/ivy_compile.go` — seed counter
3. `actions/action.go` — delete global, update Subactions/SplitReturns signatures
4. `actions/match.go` — use mod.Cfg.ActCfg
5. `actions/update.go` — pass ActCfg to Subactions, use ctx.ActCfg for Expand
6. `compiler/compiler.go` — add actCfg() helper, fix caller
7. `compiler/action.go` — fix 4 callers
8. `compiler/phase6.go` — fix 1 caller
9. Test files: `actions/audit51_test.go`, `actions/actions_test.go`, `actions/update_test.go`

## Verification

```bash
cd ~/goivy && make golden          # check line 127167 matches
XTRACE_OFF=1 go test ./actions/    # actions tests pass
XTRACE_OFF=1 go test ./compiler/   # compiler tests pass
```
