# Fix: Go fragment checker falsely reports a cycle that Python does not

**Created:** 2026-04-02 (updated)

## Context

`fragment.go:749` fires `reportCycle` on Go but Python's equivalent at `ivy_fragment.py:481` does not report a cycle. The user notes Go has `sig` set while Python's checker canon shows `sig:nil`. The fragment checker builds a stratification graph and checks for cycles — if Go builds a different graph than Python (due to `sig` differences), it can find a false cycle.

## Root Cause Analysis

### How `sig` is used — the critical gate

Both Go and Python use `sig.Interp` (a map of sort-name → interpretation) to decide how formulas are processed in `mapFmla`/`map_fmla`. Three key decision points gate whether arcs are created:

1. **`createStratMap` / `create_strat_map`** (universally quantified variables):
   - Go `fragment.go:585`: `il.IsUninterpretedSort(c.sig, v.VSort) || il.HasInfiniteInterpretation(c.sig, v.VSort)`
   - Python `ivy_fragment.py:356`: `il.is_uninterpreted_sort(v.sort) or il.has_infinite_interpretation(v.sort)`
   - **Difference**: Go uses `c.sig` (passed explicitly). Python uses global `il.sig`.

2. **`mapFmla` / `map_fmla`** (equality handling):
   - Go `fragment.go:239`: `!il.IsInterpretedSort(c.sig, sort)` → creates arcs to equality strat node
   - Python `ivy_fragment.py:115`: `not il.is_interpreted_sort(fmla.args[0].sort)` → creates arcs
   - If Go thinks a sort is NOT interpreted but Python thinks it IS, Go creates arcs that Python doesn't.

3. **`mapFmla` / `map_fmla`** (application handling):
   - Go `fragment.go:281`: `!il.IsInterpretedSymbol(c.sig, rep)` → creates arcs to function strat nodes
   - Python `ivy_fragment.py:136`: `not il.is_interpreted_symbol(func)` → creates arcs
   - Same issue: disagreement on interpretedness → different arcs.

### Where `sig` comes from

- **Go**: `CheckFragment` at `fragment.go:944` passes `m.Sig` to `CheckFEU`. The checker stores it as `c.sig`.
- **Python**: `check_fragment` does NOT pass sig. Instead, `is_interpreted_sort()` etc. use the **global** `il.sig`, which is set to `im.module.sig` via the module context manager (`ivy_module.py:96-109`).

### The `IsInterpretedSort` / `is_interpreted_sort` functions

- Go (`ivylogic/globals.go:31-43`):
  ```go
  func IsInterpretedSort(sig *Sig, s lg.Sort) bool {
      s = CanonizeSort(sig, s)
      switch cs := s.(type) {
      case *lg.UninterpretedSort:
          _, ok := sig.Interp[cs.Name]
          return ok
      case *lg.EnumeratedSort:
          _, ok := sig.Interp[cs.Name]
          return ok
      default:
          return false
      }
  }
  ```
- Python (`ivy_logic.py:1491-1493`):
  ```python
  def is_interpreted_sort(s):
      s = canonize_sort(s)
      return (isinstance(s,UninterpretedSort) or isinstance(s,EnumeratedSort)) and s.name in sig.interp
  ```

Both check `sig.Interp[sort_name]`. If Go's `c.sig.Interp` has **fewer entries** than Python's `il.sig.interp`, Go treats more sorts as uninterpreted → creates more arcs → more likely to find a cycle.

### Most likely root cause

**Go's `m.Sig.Interp` is missing entries (or has different entries) compared to Python's `il.sig.interp` at the time of fragment checking.** This causes Go to:
- Classify some sorts as "uninterpreted" that Python classifies as "interpreted"
- Create arcs in the stratification graph that Python does not create
- These extra arcs form a cycle

This is a module/compilation-level issue, not a fragment checker bug.

### Secondary difference: equality strat_map keys

A minor structural difference (not likely the root cause but worth noting):
- **Python** `ivy_fragment.py:116`: `strat_map[il.Symbol('=', fmla.args[0])]` — uses the **expression** as the second field of the Const key
- **Go** `fragment.go:240`: `sortEqKey(sort)` which is `"s:" + Key(NewConst("=", sort))` — uses the **sort**

Python creates different strat nodes for equalities with different LHS expressions (even if same sort). Go creates one strat node per sort. However, since both sides unify argument nodes with the equality node, this usually doesn't change cycle detection outcomes.

## Investigation Plan

### Step 1: Dump `sig.Interp` at fragment check entry

**Python** — Add at `ivy_fragment.py:469` (after `create_strat_map` call, in the `__debug__` block):
```python
xtracer.trace("fragment sig.interp keys: %s" % sorted(il.sig.interp.keys()))
```

**Go** — Add at `fragment.go:740` (in the xtracer block after createStratMap):
```go
if c.sig != nil {
    keys := make([]string, 0, len(c.sig.Interp))
    for k := range c.sig.Interp {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    xtracer.Trace("fragment sig.Interp keys: %v", keys)
}
```

### Step 2: Compare arcs

The checker canon already includes arcs. Run the test case and compare the xtracer output for the checker canon between Go and Python. Look for arcs in Go that don't appear in Python.

### Step 3: Trace the divergent arcs

For each arc in Go that Python doesn't have:
- Identify which formula created it (the `fmla` field)
- Check whether Go's `IsInterpretedSort`/`IsInterpretedSymbol` returned a different value than Python's equivalent
- This will pinpoint the missing `sig.Interp` entry

### Step 4: Fix the root cause

Once we know which `sig.Interp` entries are missing/different:
- Search Go's compilation pipeline for where `ImplementType` is called
- Compare with Python's `implement_type` calls
- Add the missing `ImplementType` calls in Go

## Key Files

- `~/go/src/github.com/glycerine/ivy/goivy/fragment/fragment.go` — Go fragment checker (lines 187-315 mapFmla, 567-605 createStratMap, 690-753 CheckFEU)
- `~/ivy/pyivy/ivy/ivy/ivy_fragment.py` — Python fragment checker (lines 88-152 map_fmla, 323-378 create_strat_map, 448-504 check_feu)
- `~/go/src/github.com/glycerine/ivy/goivy/ivylogic/globals.go` — Go IsInterpretedSort/IsInterpretedSymbol/IsUninterpretedSort/HasInfiniteInterpretation (lines 31-75, 229-241)
- `~/ivy/pyivy/ivy/ivy/ivy_logic.py` — Python is_interpreted_sort/is_interpreted_symbol (lines 1482-1505), global `sig` (line 1245)
- `~/go/src/github.com/glycerine/ivy/goivy/ivylogic/sig.go` — Go Sig struct definition
- `~/ivy/pyivy/ivy/ivy/ivy_module.py` — Python module context manager that sets `il.sig` (lines 96-109)

## Verification

1. Run the test case with xtracer enabled
2. Compare `sig.Interp` keys between Go and Python
3. If keys differ, fix `ImplementType` calls and re-run
4. If keys match, compare arcs directly to find the code divergence
5. Confirm Go no longer reports a false cycle
