# PLAN204: Fix elimDeadDefinitions — Python set → insertion-order dict

**Created:** 2026-04-04 ~22:15 UTC

## Context

PLAN203 fixed 5 non-deterministic map iteration sites. The golden test advanced from xtrace line 159811 to **166773**. The new divergence is:

```
Go:     name:__fml:y_a   (sort: tar_cf_clock)
Python: name:__fml:y_b   (sort: tar_cf_clock)
```

The `_a` vs `_b` suffix from `UniqueRenamer.Rename("__fml:y")` — different call order produces different suffixes when multiple captured skolems share the same name but have different sorts.

## Root Cause

**`elimDeadDefinitions` (Site C from PLAN203) — Go InsMap insertion order ≠ Python `set` hash order**

PLAN203 correctly replaced Go's `map` with `InsMap` (insertion-order preserving). But the **Python** code uses a `set()`, NOT a dict:

```python
# Python ivy_logic_utils.py:1329
defd = set(d.defines() for a in args for d in a.defs)   # SET — hash order!
captured = [sym for sym in defd if any(sym not in a.defidx for a in args)]
to_rename = [sym for sym in captured if sym.is_skolem()]
```

Python `set` iterates in **hash order** (via `recstruct.__hash__` = `hash(self._tup)`). Go's InsMap iterates in **insertion order**. When `toRename` has multiple same-name-different-sort skolems (e.g., `__fml:y:lclock` and `__fml:y:tar_cf_clock`), the first gets `_a` and the second gets `_b`. Different order → different assignment of suffixes.

## Fix Strategy

Change the Python `set` to an insertion-order `dict` (Python 3.7+ guarantees insertion order). This matches Go's InsMap with zero overhead — both iterate in insertion order, O(1) per operation. No O(n log n) sorting needed.

### Step 1: Fix Python `elim_dead_definitions` — `~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:1329`

Replace the `set` with a `dict` (using `dict.fromkeys` for set-like semantics with insertion-order iteration):

```python
def elim_dead_definitions(rn, args):
    """ If a symbol defined in one arg occurs free in another,
    then eliminate the definition by converting it to clauses """
    defd = dict.fromkeys(d.defines() for a in args for d in a.defs)  # insertion-order dict
    occurs = [set(a.symbols()) for a in args]
    captured = [sym for sym in defd if any(sym not in a.defidx for a in args)]
    dead = [sym for sym in captured if not sym.is_skolem()]
    to_rename = [sym for sym in captured if sym.is_skolem()]
    args = [rename_symbols(rn, arg, to_rename) for arg in args]
    res = [elim_definitions(a, dead) for a in args]
    return res
```

Only line 1329 changes: `set(...)` → `dict.fromkeys(...)`. The rest of the function works unchanged because:
- `for sym in defd` iterates dict keys (same API as set)
- `sym not in a.defidx` still works (dict key membership test)

### Step 2: Regenerate golden test reference

After modifying Python, regenerate the xtrace reference:
```bash
cd ~/ivy/goivy && make golden-refresh
```

### Step 3: Verify Go side (no Go changes needed)

The Go `elimDeadDefinitions` already uses InsMap (from PLAN203 Site C fix). InsMap insertion order matches Python's `dict.fromkeys` insertion order because both insert in the same sequence: `args[0].Defs` first, then `args[1].Defs`.

```bash
cd ~/ivy/goivy && go build ./...
go test ./module/... -count=1
go test ./actions/... -count=1
go test ./lalr_full/ -run TestOrdLive -count=1
```

## Files Modified

- `~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py` — line 1329: `set(...)` → `dict.fromkeys(...)`

## Why this works

Both Go and Python now iterate `defd`/`defined` in **insertion order**:
- Go: `InsMap.All()` → insertion order
- Python: `dict.fromkeys()` → insertion order (Python 3.7+)

The insertion order is identical because both are populated by the same loop: `for a in args: for d in a.defs: add d.defines()`. Duplicate keys preserve the position of their **first** insertion (InsMap.Set updates value in-place; dict.fromkeys keeps first key position).

## Non-Issues

- Go `elimDeadDefinitions` code: already correct with InsMap (PLAN203 Site C) — no changes needed
- `elim_definitions` dead-symbol iteration: `dead` is built from `captured` which now has insertion order — constraint formula order is deterministic
- `occurs` variable (line 1330): computed but never used in the function — dead code in Python, not relevant to this fix
