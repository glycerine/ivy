# Fix l2s/ranking LabeledFormula labels: use Atom instead of lg.Const

Created: 2026-04-13 ~12:30 UTC

## Context

The `TestOrdLive` golden test diverges at xtrace line 319185. Go emits a `Symbol` for the label of a `LabeledFormula` in the l2s invariant `l2s_needed_when_start`, but Python emits an `Atom`. This affects **all** invariants created by `appendLF` in l2s and `mklf` in ranking_tactic — roughly 27+ invariants.

**Go produces:** `label:(Symbol name:l2s_needed_when_start sort:(BooleanSort))`
**Python produces:** `label:(atom rep:"l2s_needed_when_start" terms:[] aSort:nil)`

## Root Cause

Two helper functions create labels with `lg.NewConst(name, &lg.BooleanSort{})` (a logic-level Const/Symbol) instead of `cfg.NewAtom(name)` (an AST-level Atom), diverging from the Python which always uses `ivy_ast.Atom(name)`.

## Changes

### 1. `check/l2s_auto.go:1020` — `appendLF` function

```go
// BEFORE:
lf := cfg.NewLabeledFormula(lg.NewConst(name, &lg.BooleanSort{}), fmla)

// AFTER:
lf := cfg.NewLabeledFormula(cfg.NewAtom(name), fmla)
```

### 2. `check/ranking_tactic.go:196` — `mklf` closure

```go
// BEFORE:
return mod.Cfg.AstCfg.NewLabeledFormula(lg.NewConst(name, &lg.BooleanSort{}), fmla)

// AFTER:
return mod.Cfg.AstCfg.NewLabeledFormula(mod.Cfg.AstCfg.NewAtom(name), fmla)
```

### NOT changed

- `l2s_auto.go:168` and `ranking_tactic.go:114` — these use `lg.NewConst("work_start"+sfx, &lg.BooleanSort{})` as actual logic constants (not labels), matching the Python `lg.Const("work_start"+sfx, lg.BooleanSort())`.

## Verification

1. `cd ~/ivy/goivy && go build ./...` — confirm it compiles
2. Run `go test -run TestOrdLive -v ./parser/ -timeout 300s` — confirm the divergence at 319185 is resolved (test may still fail later at a different xtrace index, but the label mismatch should be gone)
