# PLAN: Propagate `Failures` from fakeMod back to mod in CheckSubgoals — fix missing "error: failed checks: 2"

Created: 2026-04-17, 03:00 (tenth pass — replaces completed L2S sort plan)

**Previous plan (L2S Sexp/canon sort) is COMPLETE and verified — all 1,859,675 xtraces match.**

## Context

After all L2S sort fixes, `make tlb` shows Go and Python agreeing through all 1,859,675 xtraces — no more AST divergences. But after the matching `PASS` output, Python prints `error: failed checks: 2` while Go prints nothing (or `OK` when run standalone). The diff:

```
< OK
---
> error: failed checks: 2
```

From the `make tlb` log:
```
~go[after i=1858344]: PASS
~py[after i=1858344]: PASS
...
~go[after i=1859675]:
~py[after i=1859675]:
~py[after i=1859675]: error: failed checks: 2
```

## Root cause

**Python**: `failures` is a **module-level global** (`ivy_check.py:208`). When `check_subgoals` calls `mod = im.module.copy()` then `check_isolate()`, `Checker.fail()` increments the **global** `failures` counter regardless of which module copy is active. At `ivy_check.py:1024`, `if failures > 0: raise IvyError(None, "failed checks: 2")`.

**Go**: `Failures` is a field on `mod.Cfg` (the module config struct). `CheckSubgoals` does `fakeMod := mod.Copy()` at `isolate_check.go:698` and `:862`. `module.Copy()` does `*c.Cfg = *m.Cfg` (`module.go:364`) — a **struct value copy**. So `fakeMod.Cfg` is a different struct from `mod.Cfg`. When `CheckIsolate(fakeMod, nil)` creates checkers, `BaseChecker.Fail()` (`check.go:141`) increments `fakeMod.Cfg.Failures`. Those increments are **lost** when `fakeMod` goes out of scope. The original `mod.Cfg.Failures` at `isolate_check.go:1114` stays 0, so `CheckModule` returns nil instead of the "failed checks" error.

Call chain: `StartWithConfig` → `CheckModule(mod)` → `CheckIsolate(mod)` → `CheckSubgoals(subgoals, nil, mod)` → `fakeMod := mod.Copy()` → `CheckIsolate(fakeMod, nil)` → `BaseChecker.Fail()` → `fakeMod.Cfg.Failures++` (lost!)

Secondary issue: the error message at `isolate_check.go:1115` has a spurious `check/isolate_check.go:1115` prefix that Python doesn't have. Python: `"failed checks: {}"`. Go: `"check/isolate_check.go:1115 failed checks: %d"`.

## Fix

### Site 1: `isolate_check.go:849` — temporal branch, CheckIsolate path

After `CheckIsolate(fakeMod, nil)`, propagate failure delta back to `mod`:

```go
// Before:
				err := CheckIsolate(fakeMod, nil)

// After:
				failsBefore := fakeMod.Cfg.Failures
				err := CheckIsolate(fakeMod, nil)
				// Python's failures is a module-level global, so all
				// Checker.fail() calls aggregate regardless of which module
				// copy is active. Propagate the delta back to the original.
				mod.Cfg.Failures += fakeMod.Cfg.Failures - failsBefore
```

### Site 2: `isolate_check.go:926` — non-temporal branch, CheckIsolate path

Same pattern:

```go
// Before:
				err := CheckIsolate(fakeMod, nil)

// After:
				failsBefore := fakeMod.Cfg.Failures
				err := CheckIsolate(fakeMod, nil)
				mod.Cfg.Failures += fakeMod.Cfg.Failures - failsBefore
```

### Site 3: `isolate_check.go:1115` — error message format

Remove the file:line prefix to match Python's format:

```go
// Before:
		return fmt.Errorf("check/isolate_check.go:1115 failed checks: %d", mod.Cfg.Failures)

// After:
		return fmt.Errorf("failed checks: %d", mod.Cfg.Failures)
```

### Why this works

- `fakeMod.Cfg.Failures` starts equal to `mod.Cfg.Failures` (from the struct copy)
- After `CheckIsolate(fakeMod, nil)`, `fakeMod.Cfg.Failures` may have increased by N
- `delta = fakeMod.Cfg.Failures - failsBefore` = N
- `mod.Cfg.Failures += N` propagates exactly the new failures
- The `method()` path (lines 807-809, 890-892) already uses `mod.Cfg.Failures++` directly — no fix needed there
- Recursive calls work: each level of `CheckSubgoals` propagates its delta, so nested failures bubble up correctly

### Output chain after fix

1. `CheckIsolate(fakeMod)` → `BaseChecker.Fail()` → `fakeMod.Cfg.Failures += 2`
2. `CheckSubgoals` propagates: `mod.Cfg.Failures += 2`
3. `CheckModule` at line 1114: `mod.Cfg.Failures > 0` → true → returns `fmt.Errorf("failed checks: 2")`
4. `StartWithConfig` at line 1318: `err != nil` → returns the error (skips "OK" print)
5. `MainWithConfig` at line 1341: `fmt.Fprintf(os.Stderr, "error: %v\n", err)` → prints `error: failed checks: 2`

## Critical files

- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/isolate_check.go` — Sites 1, 2, 3

## Verification

**Do NOT run `make tlb` from this agent — the user runs it on another machine.**

Local checks after approval:
1. `go build ./check/...` — confirm compilation
2. `go test ./check/...` — confirm existing tests pass
3. `go vet ./check/...` — no warnings

After `make tlb`:
- Go should now print `error: failed checks: 2` instead of `OK`
- The `make tlb` diff should be empty (Go and Python match completely)
