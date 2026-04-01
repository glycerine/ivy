# Fix: Thread ActCfg through all UpdateContext creation sites

**Created**: 2026-04-01 23:30

## Context

`log.panic` shows a nil-pointer panic in `NewLocalActionOn` because `cfg` is nil. The call chain:

```
fragment.go:950  ctx := &actions.UpdateContext{}  // ActCfg not set!
→ GetUpdate → LocalAction.IntUpdate → Sequence.IntUpdate → IfAction.IntUpdate
→ intUpdateWithSubactions → a.Subactions(ctx.ActCfg)  // passes nil
→ subactionsSome → NewLocalActionOn(actCfg, ...)  // panics: cfg must not be nil
```

Every `UpdateContext` creation site that might reach action construction (`NewLocalActionOn`, `NewCallActionOn`, `NewChoiceActionOn`) must set `ActCfg`. The module carries it at `mod.Cfg.ActCfg`.

## Changes

### 1. `fragment/fragment.go` — `makeFmlaPairFromAction` (line 943)

Add `mod *mod.Module` parameter, set `ActCfg` and `Domain` on ctx.

```go
func makeFmlaPairFromAction(action interface{}, mod *mod.Module) (fmlaPair, bool) {
```

Update ctx creation (line 950):
```go
ctx := &actions.UpdateContext{
    Domain: mod,
}
if mod != nil && mod.Cfg != nil {
    ctx.ActCfg = mod.Cfg.ActCfg
}
```

Update callers at lines 814 and 824 in `GetAssumesAndAsserts` (which already has `m *mod.Module`):
```go
makeFmlaPairFromAction(action, m)
```

### 2. `interp/eval.go` — `ApplyAction` (line 151)

Add `ActCfg` to existing ctx. Guard with nil check:
```go
if state.Domain != nil && state.Domain.Cfg != nil {
    ctx.ActCfg = state.Domain.Cfg.ActCfg
}
```

### 3. `interp/phase4.go` — decompose ctx (line 188)

Same pattern — add ActCfg from `state1.Domain.Cfg.ActCfg`.

### 4. `actions/update.go` — `GetUpdateForArt` (line 2172)

Add ActCfg from `domain.Cfg.ActCfg`.

### 5. `actions/match.go` — `expandWhile` (line 362)

Add ActCfg from `mod.Cfg.ActCfg`.

### 6. `actions/action.go` — `WhileAction.DecomposeWithModule` (line 1411)

Add ActCfg from `mod.Cfg.ActCfg`.

### 7. `actions/phase3.go` — line 436

Add ActCfg from `domain.Cfg.ActCfg`.

### 8. `mc/toaiger.go` — line 124

Add ActCfg from `mod.Cfg.ActCfg`.

### Files to modify
- `fragment/fragment.go` — add `mod` param to `makeFmlaPairFromAction`, set ActCfg
- `interp/eval.go` — set ActCfg on ctx
- `interp/phase4.go` — set ActCfg on ctx
- `actions/update.go` — set ActCfg in `GetUpdateForArt`
- `actions/match.go` — set ActCfg in `expandWhile`
- `actions/action.go` — set ActCfg in `DecomposeWithModule`
- `actions/phase3.go` — set ActCfg on ctx
- `mc/toaiger.go` — set ActCfg on ctx

## Verification

```bash
cd ~/ivy/goivy && go build ./... && make golden
```

The panic should be gone and the golden test should advance past the current crash point.
