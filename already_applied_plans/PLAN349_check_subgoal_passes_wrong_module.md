# Fix: CheckSubgoals passes wrong module to MC/VMT method closures

Created: 2026-04-30 ~14:30 UTC

## Context

XTRACE divergence at point 28253294:
- Go emits: `mc.toaiger calling GetUpdate type=Sequence`
- Python emits: `ops.ToOpenFormula nFmlas=0 nDefs=0`

**Root cause:** In `MCTactic` (check/check.go:1082), the `mcMethod` closure captures `mod` from MCTactic's parameter. But `CheckSubgoals` (check/isolate_check.go:693) creates `withLocalMod = mod.Copy()` and populates it with the temporal model's fields (actions, initializers, labeled_conjs, assumed_invs, etc.). The closure then calls `mc.CheckIsolate(mod, "mc")` with the **original** module instead of `withLocalMod`.

In Python, `check_isolate()` reads `im.module`, which is set to the copy by `with mod:` (ivy_check.py:831). So Python's `check_isolate` operates on the modified copy. Go's closure operates on the unmodified original — completely different actions, initializers, and conjectures.

Same bug exists in `VMTTactic` (check/check.go:1107).

## Fix

Change `CheckSubgoals`'s `method` parameter from `func() error` to `func(*module.Module) error`, so the caller receives `withLocalMod`.

### Files to modify

**1. check/isolate_check.go** — `CheckSubgoals` signature and call sites

- Line 664: Change signature from `method func() error` to `method func(*module.Module) error`
- Line 805: Change `err := method()` to `err := method(withLocalMod)`
- Line 902: Change `err := method()` to `err := method(withLocalMod)`

**2. check/check.go** — `MCTactic` and `VMTTactic` closures

- Lines 1082-1091: Change `mcMethod` closure:
  ```go
  mcMethod := func(m *module.Module) error {
      res, mcErr := mc.CheckIsolate(m, "mc")
      ...
  }
  ```
- Lines 1107-1109: Change `vmtMethod` closure:
  ```go
  vmtMethod := func(m *module.Module) error {
      return VMTCheckIsolate("vmt", m)
  }
  ```

### Callers passing nil (no change needed for nil checks)

- check/check.go:357 — `CheckSubgoals(subgoals, nil, mod)` — nil method, the `if method != nil` guard handles this
- check/isolate_check.go:88 — `CheckSubgoals(subgoals, nil, mod)` — same
- Test files — pass nil, unaffected

## Verification

Run `cd ~/ivy/goivy && make test` and confirm the golden test passes through XTRACE 28253294 without divergence.
