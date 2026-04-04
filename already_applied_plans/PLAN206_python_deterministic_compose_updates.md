# PLAN205: Fix compose_updates non-deterministic set iteration

**Created:** 2026-04-04 ~23:30 UTC

## Context

PLAN204 fixed RenameDistinct per-symbol renaming + set ordering. The golden test advanced from xtrace line 166773 to **166784**. The divergence at line 166784:

```
Go:   (Eq  t1:(Apply func:(Symbol name:new_ref.pnd_or_ser_rd ...)))
Py:   (Iff t1:(Apply func:(Symbol name:new_ref.both_nGnR ...)))
```

These are entirely different formulas at the same position in the `fmlas` list — the list is ordered differently between Go and Python.

## Root Cause: Two non-deterministic set operations in Python's `compose_updates`

**File:** `~/ivy/pyivy/ivy/ivy/ivy_transrel.py`, function `compose_updates` (lines 305-355)

### Bug 1: `mid` set iteration (line 317, iterated at 327)

```python
mid = us1.intersection(us2)          # set — non-deterministic iteration order
...
for mv in mid:                       # line 327: iterates in hash order
    mvf = rename(mv, rn)             # rn.Rename() assigns _a, _b, etc.
    map1[new(mv)] = mvf              # different suffix per hash order
    map2[mv] = mvf
```

**Go** (lines 644-650 in transrel.go) iterates `updated1` slice order:
```go
for _, s := range updated1 {
    if constSetContains(us2, s.Name) {
        mid = append(mid, s)
    }
}
```

### Bug 2: `new_updated` from set union (line 336)

```python
new_updated = list(us1.union(us2))   # set union → list: non-deterministic order
```

This `new_updated` becomes `updated1` or `updated2` for the NEXT `compose_updates` call, cascading non-determinism through:
- `mid` computation (which symbols intersect, in what order)
- `diff_frame(updated1, updated2, ...)` → `list_diff(updated2, updated1)` → `frame(updated, op)` — frame condition ordering
- `map2` construction (line 325-326) — substitution naming

**Go** `UpdatedJoinConst` (lines 1689-1709) preserves `u1` order then appends new `u2` elements — deterministic.

## Fix Plan — Python only (Go is already correct)

### Step 1: Fix `mid` — line 317

Change from set intersection to ordered list preserving `updated1` order:

```python
# was: mid = us1.intersection(us2)
mid = [s for s in updated1 if s in us2]
```

`us2` remains a set for O(1) membership tests. The list preserves `updated1` order, matching Go.

### Step 2: Fix `new_updated` — line 336

Change from set union to ordered dedup preserving `updated1` then `updated2` order:

```python
# was: new_updated = list(us1.union(us2))
new_updated = list(dict.fromkeys(list(updated1) + list(updated2)))
```

`dict.fromkeys` preserves first-occurrence insertion order (Python 3.7+), matching Go's `UpdatedJoinConst` which appends `u1` elements then new `u2` elements.

### Step 3: Regenerate golden + verify

```bash
cd ~/ivy/goivy && make golden-refresh
go test ./lalr_full/ -run TestOrdLive -count=1
```

## Files Modified

- `~/ivy/pyivy/ivy/ivy/ivy_transrel.py` — lines 317 and 336 in `compose_updates`

## Why Go needs no changes

- `mid`: already iterates `updated1` slice, filters by `us2` map — deterministic (transrel.go:644-650)
- `UpdatedJoinConst`: iterates `u1` then `u2`, deduplicates by `lg.Key` — deterministic (transrel.go:1689-1709)
