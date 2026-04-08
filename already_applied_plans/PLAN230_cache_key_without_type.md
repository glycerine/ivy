# Fix TestOrdLive XTRACE divergence at line 236129: Sort cache key mismatch

Created: 2026-04-08 ~14:00 UTC

## Context

TestOrdLive diverges at XTRACE line 236129. Go emits `get_id() ENTER` (enumeratedsort cache miss path), while Python emits `solver_name() ENTER name=ref.evs.req` (enumeratedsort returned early via cache hit, caller continues to next step).

The divergence occurs when translating a FunctionSort whose range is the EnumeratedSort "op_type". Python finds the sort cached; Go does not.

## Root Cause

Two related issues in `z3bridge/translate.go`:

### 1. Sort cache key is type-specific (Go) vs flat name (Python)

**Python** `ivy_solver.py`: Both `uninterpretedsort()` (line 259) and `enumeratedsort()` (line 277) share the same `z3_sorts` dict and use `es.rep` as the key. The `rep` property (ivy_logic.py:778,791) returns `self.name` -- a flat string like `"op_type"`.

**Go** `translate.go`: `TranslateSort` uses `s.Sexp()` as the cache key, which is type-specific:
- UninterpretedSort: `(UninterpretedSort name:op_type)`
- EnumeratedSort: `(EnumeratedSort name:op_type ext:[req,ack,...])`

When "op_type" is first encountered via UninterpretedSort -> LookupNative -> TranslateSort(EnumeratedSort), Go caches under both type-specific keys. Later, when a direct TranslateSort(EnumeratedSort) call uses a possibly different EnumeratedSort object (or looks under a different Sexp key), Go gets a cache miss while Python gets a hit.

### 2. Extra `get_id()` trace from enumeratedsort

**Python** `enumeratedsort()` (line 275-284) does NOT call `get_id()` or store in `z3_sorts_inv`. The reverse map is populated by `uninterpretedsort()` (line 265) which always calls `get_id()`.

**Go** `enumeratedsort` case (line 185) calls `t.sortsInv[zs.GetId()] = s`, producing an extra "get_id() ENTER" trace that Python never emits. The comment on line 183-184 acknowledges this: "Python enumeratedsort() does NOT store in z3_sorts_inv".

## Fix

### Change 1: Sort cache key -> name-based (matching Python `rep`)

In `translate.go`, change the sort cache key from `s.Sexp()` to `logic.NodeKey(name)`:

| Line | Current | New |
|------|---------|-----|
| 143 | `key := s.Sexp()` | `key := logic.NodeKey(st.Name)` |
| 176 | `key := s.Sexp()` | `key := logic.NodeKey(st.Name)` |
| 624 | `sortKey := sort.Sexp()` | `sortKey := logic.NodeKey(sortDisplayName(sort))` |

This matches Python's `z3_sorts[es.rep]` / `z3_sorts[us.rep]` where `rep = self.name`.

### Change 2: Remove sortsInv/GetId from enumeratedsort case

Remove lines 183-185 in `translate.go`:
```go
// Python enumeratedsort() does NOT store in z3_sorts_inv,
// but Go needs it for SortFromZ3 reverse lookups.
t.sortsInv[zs.GetId()] = s
```

This matches Python where `enumeratedsort()` does NOT populate `z3_sorts_inv`. The reverse map entry is created by `uninterpretedsort()` (the normal entry point for interpreted sorts).

### Change 3: Update TestSortFromZ3RoundTrip

In `solver/solver_test.go` lines 1584-1601: change the EnumeratedSort section to expect `SortFromZ3` to return `false` for directly-translated EnumeratedSorts. This matches Python's behavior. The herbrand.go code has a fallback (`sig.Sorts` lookup at line 54-56) that handles this in production.

## Files to Modify

1. `/Users/jaten/ivy/goivy/z3bridge/translate.go` -- lines 143, 176, 183-185, 624
2. `/Users/jaten/ivy/goivy/solver/solver_test.go` -- lines 1584-1601

## Safety Analysis

- **Shared key space is intentional**: Python uses one dict for all sort types, keyed by name. Sort names are unique within a module.
- **`t.consts` keys are unaffected**: The const map uses `name + ":" + sort.Sexp()` which is a separate concern.
- **sortsInv for UninterpretedSort is always populated**: In the normal flow (uninterpretedsort -> lookup_native), line 152 stores `sortsInv[id]`. SortFromZ3 works for all sorts encountered through UninterpretedSort (the production path).
- **`sortDisplayName` already exists** at line 650 and returns `s.Name` for UninterpretedSort, EnumeratedSort, RangeSort.

## Separate Issue (not in scope)

The log also shows definition output differences: Go prints `(internal) ref.def80` while Python prints `<IVY_EXAMPLES>/.../ord_live.ivy: line 168: ref.def80`. This is a separate issue related to definition source location tracking, not the XTRACE divergence.

## Verification

1. `cd /Users/jaten/ivy/goivy && go test ./z3bridge/... -run .` -- z3bridge tests pass
2. `cd /Users/jaten/ivy/goivy && go test ./solver/... -run TestSortFromZ3RoundTrip` -- updated test passes
3. `cd /Users/jaten/ivy/goivy && go test ./solver/... -run .` -- all solver tests pass
4. `cd /Users/jaten/ivy/goivy && go test ./parser/ -run TestOrdLive -count=1` -- XTRACE divergence at 236129 resolved
