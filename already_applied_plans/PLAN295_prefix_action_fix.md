# Fix SharedStep7 Event Instrumentation Bugs

**Created:** 2026-04-13 ~20:30 UTC

## Context

TestOrdLive diverges at xtrace line 596113, immediately after `l2s.SharedStep7 ENTER`. Go emits `LocalAction.__init__ uniqueID=1359 caller=ast.LocalAction.clone` while Python emits `ilu.replaceTemporalsRec ENTER type=Not ...`. Three bugs in Go's SharedStep7 combine to produce this divergence.

## Root Cause Analysis

### Bug 1 — Key mismatch in symprops/symwhens/symwaits lookup (PRIMARY)

**File:** `check/l2s_shared.go`

The `symprops`, `symwhens`, `symwaits` maps are declared as `map[lg.NodeKey][]*lg.NamedBinder` and populated with `lg.Key(sym)` keys (lines 359–381). `lg.Key(sym)` calls `sym.Sexp()` which returns the full canonical S-expression, e.g. `NodeKey("(Symbol name:cfabric.rd_fair sort:(BooleanSort))")`.

But the lookup at line 546 uses:
```go
symKey := lg.NodeKey(sym)   // sym is a plain string like "cfabric.rd_fair"
```

`lg.NodeKey("cfabric.rd_fair")` ≠ `lg.Key(constSym)` = `"(Symbol name:cfabric.rd_fair sort:(BooleanSort))"`.

**Result:** The lookup NEVER matches. No prop/when/wait events are ever found. SharedStep7 adds zero temporal instrumentation events — the entire step is a no-op.

### Bug 2 — PrefixAction wraps even when stmts is empty

**File:** `actions/helpers.go`, lines 71–84

`PrefixAction` always creates `NewSequence(stmts..., action)` even when `stmts` is nil/empty, returning `Sequence(action)`. By contrast, `PostfixAction` (line 86) has an early return `if len(stmts) == 0 { return action }`.

Because Bug 1 means no events are found, every call to `PrefixAction` wraps in an unnecessary single-element Sequence. This makes every child appear "changed" to its parent (pointer inequality), triggering clones all the way up the tree — including `LocalAction` clones that emit `LocalAction.__init__` xtraces.

### Bug 3 — Modifies() recurses into children (Python doesn't)

**File:** `actions/transforms.go`, lines 249–255

`modifiesRec` has a `default` case that recurses into all children of compound actions (Sequence, IfAction, LocalAction, etc.). Python's `Action.modifies()` returns `[]` for ALL compound actions — only `AssignAction`, `HavocAction`, and `CrashAction` override it.

This causes events to be computed at every level of the action tree (parent AND leaf), producing duplicate instrumentation. Currently masked by Bug 1 but would cause incorrect behavior once Bug 1 is fixed.

## Fixes

### Fix 1 — Use symbol name strings as map keys (check/l2s_shared.go)

Change `symprops`, `symwhens`, `symwaits` from `map[lg.NodeKey]` to `map[string]`, keyed by `sym.Name`. The lookup already iterates `allDeps` (a `map[string]bool` of symbol names from `cfg.Dependencies`), so using `sym.Name` as key makes both sides consistent.

**Lines to change:**
- Line 359: `map[lg.NodeKey][]*lg.NamedBinder` → `map[string][]*lg.NamedBinder` (×3 maps)
- Lines 367–368: `symprops[lg.Key(sym)]` → `symprops[sym.Name]` (×2 occurrences)
- Lines 373: `symwhens[lg.Key(sym)]` → `symwhens[sym.Name]`
- Lines 379: `symwaits[lg.Key(sym)]` → `symwaits[sym.Name]`
- Lines 547–554: Remove `symKey := lg.NodeKey(sym)` and change `symprops[symKey]` → `symprops[sym]` (already a string), same for symwhens/symwaits

### Fix 2 — Add early return to PrefixAction (actions/helpers.go)

Add at line 72 (top of function):
```go
if len(stmts) == 0 {
    return action
}
```

This matches `PostfixAction`'s existing early return (line 87).

### Fix 3 — Remove default recursion from modifiesRec (actions/transforms.go)

Remove lines 249–255 (the `default:` case that recurses into children). After this, `Modifies(sequence)` returns `[]`, matching Python's `Sequence.modifies()` → `[]`.

**Impact on other callers of Modifies (8 total):**
- `check/l2s_shared.go:193` — iterates `IterSubactions()` so all leaves are visited individually; removing recursion is safe.
- `check/l2s_shared.go:539` — THIS is the buggy call; fix is correct.
- `compiler/phase6.go:2827` — iterates actions individually.
- `compiler/ivy_compile.go:1541,1553` — iterates actions individually.
- `temporal/temporal.go:664` — same `instrStmt` pattern; recursion removal is correct (Python's equivalent also returns `[]` for compound types here).
- `isolate/strip.go:152` — iterates actions individually.
- `isolate/phase7.go:119` — iterates `IterSubactions()`.

All callers either iterate subactions (so leaf actions provide the symbols) or follow the Python `instrStmt` pattern (where only leaf-level modifies matters). No caller depends on recursive collection through `Modifies`.

## Files to Modify

1. **`check/l2s_shared.go`** — Fix 1: change map key types and lookup
2. **`actions/helpers.go`** — Fix 2: add early return to PrefixAction
3. **`actions/transforms.go`** — Fix 3: remove default recursion from modifiesRec

## Verification

Run the failing test:
```
cd /Users/jaten/ivy/goivy/parser && go test -run TestOrdLive -timeout 600s -v
```

The test compares Go and Python xtrace output line-by-line. A pass means all 596113+ lines match.
