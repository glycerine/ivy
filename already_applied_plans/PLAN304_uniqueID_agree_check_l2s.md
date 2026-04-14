# Plan: Fix ChoiceAction UniqueID in l2s.go — use counter-based constructor

Created: 2026-04-15 ~01:30 UTC

## Context

**Previous fix**: Converted `Sig.Sorts` from `map` to `InsMap` — that fix is complete and merged.

**New divergence**: `make golden` now diverges at xtrace line 844423 (TestOrdLive, binding[24] `idle`). The s-expression diff shows:

```
-         uniqueID:0)
+         uniqueID:1)
```

Go produces `uniqueID:0` for the `choiceAction` node in the idle binding, while Python produces `uniqueID:1`.

**Root cause**: In `check/l2s.go:757`, `actions.NewChoiceAction(...)` is called without a config. This constructor (`actions/action.go:675`) creates a `ChoiceAction` with `UniqueID: 0` (Go zero value). The counter-based constructor `NewChoiceActionOn(cfg, ...)` exists but is not used.

In Python (`ivy_actions.py:870-877`), `ChoiceAction.__init__` always increments a global `choice_action_ctr` and assigns `self.unique_id`. By the time l2s runs, the counter is already at 1 (one ChoiceAction was created during compilation of `if * ... else ...` in the parser), so the idle binding's ChoiceAction gets `unique_id=1`.

**Fix**: Replace `actions.NewChoiceAction(...)` with `actions.NewChoiceActionOn(m.Cfg.ActCfg, ...)` in l2s.go. The module's `ActCfg` is set up by the compiler (`compiler/ivy_compile.go:103-109`) with `IuCfg` shared with `AstCfg.IuCfg`, so the `ChoiceActionCtr` is properly shared across the pipeline.

## Changes

### 1. `check/l2s.go` — one-line fix

Line 757: Replace the configless constructor with the counter-based one:

```go
// Before:
setLineno(actions.NewChoiceAction(

// After:
setLineno(actions.NewChoiceActionOn(m.Cfg.ActCfg,
```

The variable `m` is already available (line 312: `m := pc.GetModule()`).

No other changes needed — this is the only `NewChoiceAction()` call in `check/`.

## Verification

1. `go build ./...` compiles clean
2. `go test ./...` passes
3. `make golden` — divergence advances past line 844423
