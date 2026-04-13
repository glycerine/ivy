# Fix makeAnd/buildOrExpr single-element unwrapping divergence

Created: 2026-04-13 ~12:45 UTC

## Context

After fixing the label type (Symbol→Atom), the next divergence is at xtrace 319290, invariant `l2s_created`. Go's `allD` function calls `makeAnd(cons...)` which, when `cons` has exactly 1 element, returns the bare element. Python's `lg.And(*cons)` always wraps in `And`, even for 1 element.

**Go produces:** `t2:(Apply func:(Symbol name:l2s_d ...)` — unwrapped single term
**Python produces:** `t2:(And terms:[(Apply func:(Symbol name:l2s_d ...)` — And wrapping single term

## Root Cause

`makeAnd` in `check/l2s.go:142-150` has a shortcut `if len(terms) == 1 { return terms[0] }`. Python's `logic.py:297` `And` class stores terms as-is with no simplification. Same issue exists in `buildOrExpr` at `check/l2s_auto.go:1027-1032`.

The 0-element case is fine: `makeAnd()` returns `lg.True` which is `&And{}` — structurally identical to Python's `And()`.

## Changes

### 1. `check/l2s.go:142-150` — remove single-element shortcut from `makeAnd`

```go
// BEFORE:
func makeAnd(terms ...lg.Expr) lg.Expr {
    if len(terms) == 0 {
        return lg.True
    }
    if len(terms) == 1 {
        return terms[0]
    }
    return &lg.And{Terms: terms}
}

// AFTER:
func makeAnd(terms ...lg.Expr) lg.Expr {
    if len(terms) == 0 {
        return lg.True
    }
    return &lg.And{Terms: terms}
}
```

### 2. `check/l2s_auto.go:1027-1032` — remove single-element shortcut from `buildOrExpr`

```go
// BEFORE:
func buildOrExpr(xs []lg.Expr) lg.Expr {
    if len(xs) == 1 {
        return xs[0]
    }
    return &lg.Or{Terms: xs}
}

// AFTER:
func buildOrExpr(xs []lg.Expr) lg.Expr {
    return &lg.Or{Terms: xs}
}
```

### 3. Update test in `check/l2s_test.go` — single-element makeAnd test

The test `TestMakeAnd_SingleTerm` (around line 232) asserts that `makeAnd(c)` returns `c` directly. After the fix, it should return `&And{Terms: [c]}`.

## Files to modify

- `check/l2s.go:142-150`
- `check/l2s_auto.go:1024-1032`
- `check/l2s_test.go` (update single-element test expectation)

## Verification

1. `go build ./...` — compiles
2. `go test ./check/ -run TestMakeAnd` — updated test passes
3. Run golden test to verify divergence moves past 319290
