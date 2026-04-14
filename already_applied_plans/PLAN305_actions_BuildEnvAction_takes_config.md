# Plan: Fix EnvAction/ChoiceAction UniqueID counter — BuildEnvAction must increment

Created: 2026-04-15 ~05:00 UTC

## Context

**Previous fix**: Changed `l2s.go:757` from `actions.NewChoiceAction(...)` to `actions.NewChoiceActionOn(m.Cfg.ActCfg, ...)` and removed the configless `NewChoiceAction()` constructor. That fix was correct but insufficient — the same divergence persists.

**Divergence**: `make golden` → TestOrdLive → s-expression diff at the l2s idle binding's `choiceAction`:
```
-         uniqueID:0)   (Go)
+         uniqueID:1)   (Python)
```

**Root cause**: `BuildEnvAction()` in `actions/action.go:2068` creates `&EnvAction{}` directly without incrementing `ChoiceActionCtr`. In Python, `EnvAction` inherits from `ChoiceAction`, so every `EnvAction()` call goes through `ChoiceAction.__init__` which increments `choice_action_ctr`.

**Execution flow for ord_live.ivy** (confirmed via xtrace):
1. `CheckIsolate` runs
2. Guarantee loop at `isolate_check.go:584` calls `BuildEnvAction(...)` — creates 1 EnvAction
   - Python: `act.env_action(root)` → `EnvAction(*racts)` → counter goes 0→1
   - Go: `BuildEnvAction(...)` → `&EnvAction{}` → counter stays 0
3. `CheckTemporals(mod)` at `isolate_check.go:652`
4. L2S at `l2s.go:757` creates ChoiceAction:
   - Python: gets uniqueID=1 ✓
   - Go: gets uniqueID=0 ✗

**Why the previous fix didn't help**: `l2s.go:757` now correctly uses `NewChoiceActionOn(m.Cfg.ActCfg, ...)`, but the counter starts at 0 because the preceding `BuildEnvAction` call didn't increment it.

## Changes

### 1. `actions/action.go` — Fix `BuildEnvAction` to use counter

Add `*ActionsConfig` as the first parameter. Use it to set UniqueID and increment the counter:

```go
// Before (line 2024):
func BuildEnvAction(publicActions *iu.InsMap[string, bool], actionsMap *iu.InsMap[string, Action], actName string, label string) *EnvAction {
    ...
    env := &EnvAction{}
    env.Branches = branches

// After:
func BuildEnvAction(cfg *ActionsConfig, publicActions *iu.InsMap[string, bool], actionsMap *iu.InsMap[string, Action], actName string, label string) *EnvAction {
    ...
    id := cfg.IuCfg.ChoiceActionCtr
    cfg.IuCfg.ChoiceActionCtr++
    env := &EnvAction{}
    env.UniqueID = id
    env.ActCfg = cfg
    env.Branches = branches
```

### 2. `check/isolate_check.go` — Update BuildEnvAction callers

Two call sites, both have `mod` available:

**Line 379:**
```go
// Before:
action := actions.BuildEnvAction(mod.PublicActions, mod.Actions, actname, "")
// After:
action := actions.BuildEnvAction(mod.Cfg.ActCfg, mod.PublicActions, mod.Actions, actname, "")
```

**Line 584:**
```go
// Before:
envAction := actions.BuildEnvAction(mod.PublicActions, mod.Actions, root, "")
// After:
envAction := actions.BuildEnvAction(mod.Cfg.ActCfg, mod.PublicActions, mod.Actions, root, "")
```

### 3. `actions/action.go` — Fix `EnvAction.ActionClone` to use counter

Currently (line 1082-1084) the clone preserves the old UniqueID and doesn't increment. In Python, `AST.clone` calls `type(self)(*args)` which goes through `ChoiceAction.__init__`, getting a new UniqueID.

```go
// Before:
func (a *EnvAction) ActionClone(args []lg.Expr) Action {
    return &EnvAction{ChoiceAction: ChoiceAction{ActionBase: a.ActionBase, Branches: copyNodes(args), UniqueID: a.UniqueID}}
}

// After:
func (a *EnvAction) ActionClone(args []lg.Expr) Action {
    if a.ActCfg != nil {
        r := NewEnvActionOn(a.ActCfg, args...)
        r.ActionBase = a.ActionBase
        return r
    }
    return &EnvAction{ChoiceAction: ChoiceAction{ActionBase: a.ActionBase, Branches: copyNodes(args)}}
}
```

This matches `ChoiceAction.ActionClone` (line 686-693) which already does the same pattern.

### 4. Fix remaining configless `NewEnvAction` callers (prevention)

These don't cause the current divergence but will cause future ones once the golden test advances. All create EnvActions that Python counts but Go doesn't:

| File | Line | Current | Fix |
|------|------|---------|-----|
| `temporal/temporal.go` | 394 | `NewEnvAction(branches...)` | Needs config threaded through |
| `trace/trace.go` | 542 | `NewEnvAction(branches...)` | Needs config threaded through |
| `tactics/tactics.go` | 351 | `NewEnvAction(branches...)` | Needs config threaded through |
| `bmc/bmc.go` | 221 | `NewEnvAction(branches...)` | Needs config threaded through |
| `art/art.go` | 1542 | `NewEnvAction(innerSeq)` | Needs config threaded through |
| `mc/toaiger.go` | 55 | `NewEnvAction(extActs...)` | Needs config threaded through |

**Strategy**: Each of these functions needs access to the module's ActionsConfig. Check each caller to see if `mod` or `mod.Cfg.ActCfg` is available. If not, thread the config through the function parameters.

### 5. Remove configless `NewEnvAction` constructor

After all callers are migrated, remove `NewEnvAction` (line 1068-1071) — same pattern as the previous removal of `NewChoiceAction`.

Fix test callers to use `NewEnvActionOn(NewActionsConfig(), ...)`:
- `gogen/action_test.go:308`
- `actions/actions_test.go:211`
- `actions/stub_fixes_test.go:162,179,206,581`
- `isolate/batch_fixes_test.go:685`

## Critical files

- `actions/action.go:2024` — BuildEnvAction function
- `actions/action.go:1068-1079` — NewEnvAction / NewEnvActionOn
- `actions/action.go:1082` — EnvAction.ActionClone
- `check/isolate_check.go:379,584` — BuildEnvAction callers
- `temporal/temporal.go:394` — EnvAction in temporal pipeline
- `check/l2s.go:757` — already fixed, creates the diverging ChoiceAction

## Verification

1. `go build ./...` compiles clean
2. `go test ./...` passes
3. `make golden` — uniqueID divergence resolves; golden test advances past this point
