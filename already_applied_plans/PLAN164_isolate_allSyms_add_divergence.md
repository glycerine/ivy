# Fix allSyms.add Trace Ordering Mismatch

note: REDONE. see PLAN165.

**Created:** 2026-04-01 (current session)

## Context

At XTRACE line 151058, Go and Python diverge:
- Go: `isolate.allSyms.add index.succ`
- Python: `isolate.allSyms.add <`

Both symbols are correct and both end up in `allSyms`. The issue is **iteration order** when a formula contributes 2+ new symbols:

- **Python** (`_traced_add_syms` at ivy_isolate.py:43): iterates `new_syms` which is a `set()` (from `gen_to_set(symbols_ilu_ast)`). Python `set` iteration uses hash-table order.
- **Go** (`collectSymbolsInto` at helpers.go:407): iterates `il.SymbolsIluAst(node)` generator directly, yielding symbols in AST traversal order.

For 151058 lines the orders coincided (most formulas contribute 0 or 1 new symbols). At line 151058, a formula contributes 2+ new symbols where hash-order != traversal-order.

## Fix

Sort new symbols alphabetically before tracing, on **both** sides. This changes only trace order, not final set contents.

### File 1: Go — `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/helpers.go`

**Function:** `collectSymbolsInto` (lines 407–420)

Change to:
1. Collect all symbols from generator into a temp map (matching Python's `set()` dedup)
2. Find which are new to the accumulator
3. Sort new symbols by `lg.PrettyFmla(sym)`
4. Trace in sorted order
5. Add all to accumulator

When `xtracer.Enabled` is false, keep the current fast path (no temp allocations).

`sort` is already imported (line 8).

### File 2: Python — `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_isolate.py`

**Function:** `_traced_add_syms` (lines 43–48)

Change to:
1. Collect new symbols (not in `target_set`) into a list
2. Sort by `str(sym)` (matches Go's `PrettyFmla` for bare Consts)
3. Trace in sorted order
4. Add all to `target_set` via `update()`

## Verification

Run: `cd lalr_full && go test -v -run TestVerboseOrdLive`

The trace comparison should pass through line 151058 and beyond.
