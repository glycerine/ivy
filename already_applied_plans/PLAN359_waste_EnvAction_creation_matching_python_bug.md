# Fix ChoiceAction/EnvAction divergences

**Created:** 2026-05-01 ~UTC now

## Completed: xtracer instrumentation (permanent traces)
9 `ChoiceAction.__init__` traces across `ast/ast.go` (4), `actions/action.go` (4), `ivy_actions.py` (1).

## Completed: isolate/create.go guard removal
Removed `if len(extBranches) > 0` guard to match Python's unconditional `EnvAction(*ext_acts)`.

## Current fix: mc/checker.go — add wasted EnvAction creation matching Python

**Root cause:** Python `check_isolate` (ivy_mc.py:1733-1734) creates `EnvAction(*ext_acts)` BEFORE calling `to_aiger`. But `to_aiger` (ivy_mc.py:1157) SHADOWS that parameter by creating a new `EnvAction` after `add_err_flag_mod`. The first creation is wasted but increments `choice_action_ctr`. Go's `CheckIsolate` goes straight to `ToAiger` with no equivalent wasted creation, so the counter falls behind.

**Fix — `mc/checker.go` (1 change):**

In `CheckIsolate`, before the `ToAiger(mod, method)` call, add the matching wasted EnvAction creation:

```go
// Python: check_isolate lines 1733-1734
// ext_acts = [mod.actions[x] for x in sorted(mod.public_actions)]
// ext_act = ia.EnvAction(*ext_acts)
// Note: to_aiger shadows this parameter, but the creation still increments the counter.
pubNames := sortedPublicActions(mod)
wastedExtActs := make([]lg.Expr, 0, len(pubNames))
for _, name := range pubNames {
    if act, ok := mod.Actions.Get2(name); ok {
        wastedExtActs = append(wastedExtActs, act)
    }
}
_ = actions.NewEnvActionOn(mod.Cfg.ActCfg, wastedExtActs...)
```

Add this between the logfile setup and the `ToAiger` call (around line 152 of checker.go).

## Files to modify
- `/Users/jaten/ivy/goivy/mc/checker.go` — add wasted EnvAction creation

## Verification
`cd ~/ivy/goivy && make test`
