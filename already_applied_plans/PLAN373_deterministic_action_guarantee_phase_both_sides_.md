# Fix non-deterministic action ordering in guarantee phase (both Go and Python)

**Created:** 2026-05-02 ~UTC

## Context

`TestThisIso` diverges non-deterministically. The callee name diagnostic shows Go and Python sometimes call different actions (`ext:rfn.abs.recv1` vs `ext:rfn.abs.recv2`) at the same trace position. The mismatch is non-deterministic — confirming map/set iteration randomness.

## Root cause

**Both** Go and Python iterate non-deterministic collections in the guarantee phase:

- **Python** `ivy_check.py:642`: `checked_actions = set(get_checked_actions())` — a Python `set` iterated at lines 732 and 737
- **Go** `check/isolate_check.go:382`: `checkedActions := setFromSlice(GetCheckedActions(mod))` — a Go `map[string]bool` iterated at lines 578 and 590 (already fixed to use `actionOrder`)

Both sides already compute a sorted equivalent:
- **Python** line 645: `actions = sorted(list(prioritized)) + sorted(list(checked_actions.difference(prioritized)))`
- **Go** line 385: `actionOrder := sortedUnion(prioritizedChecked, checkedActions)`

## Fix

### Python: `~/ivy/pyivy/ivy/ivy/ivy_check.py`

Replace `checked_actions` iteration with `actions` (the sorted list) at:

- **Line 732**: `any(... for r in checked_actions)` → `any(... for r in actions)`
- **Line 737**: `for root in checked_actions:` → `for root in actions:`

### Go: `check/isolate_check.go` — already done

Lines 578 and 590 already changed from `for root := range checkedActions` to `for _, root := range actionOrder`.

## Also: remove temporary PLAN368 diagnostic panics — already done

The three panics in `compiler/action.go`, `actions/action_expr.go`, `module/astutil.go` were already removed.

## Verify

```bash
cd ~/ivy/goivy && make test TESTARGS='-run TestThisIso'
```

Run 2-3 times to confirm deterministic results.
