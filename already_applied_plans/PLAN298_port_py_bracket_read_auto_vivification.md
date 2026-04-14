# Plan: Fix all 40 Python defaultdict auto-vivification sites in Go port

Created: 2026-04-14 ~11:00 UTC

## Context

Python's `defaultdict` creates an entry on any bracket-read `d[key]` for a missing key. Go's `map[key]` does NOT create entries on read. This is a mechanical port — if Python creates an entry, Go must too. The map state must be identical. We already found and fixed this for `symprops/symwhens/symwaits` in `l2s_shared.go`. Now we need to fix all 40 remaining sites.

## Fix pattern

Before every bracket-read on a Go map that corresponds to a Python defaultdict bracket-read with a potentially missing key:

```go
// Match Python defaultdict auto-vivification
if _, ok := m[key]; !ok {
    m[key] = nil  // or zero value for the value type
}
```

## All 40 sites

### 1. ivy_l2s.py → check/l2s_shared.go, check/l2s.go

**named_binders_conjs (4 sites):**
- Py:891 `named_binders_conjs['l2s_s']` → Go: `cfg.NamedBindersConjs["l2s_s"]` at l2s_shared.go:193
- Py:899 `named_binders_conjs['l2s_w']` → Go: `cfg.NamedBindersConjs["l2s_w"]` at l2s_shared.go:224
- Py:906 `named_binders_conjs['l2s_init'].append(...)` → Go: `cfg.NamedBindersConjs["l2s_init"]` at l2s_shared.go:239
- Py:1428 `named_binders['l2s_g']` → Go: `namedBinders.Get2("l2s_g")` at l2s_shared.go (InsMap, already safe)

**symprops/symwhens/symwaits (3 sites) — ALREADY FIXED:**
- Py:1309-1315 → Go: l2s_shared.go:626-633

### 2. ivy_ranking.py → check/ranking.go

**named_binders_conjs (3 sites):**
- Py:631 `named_binders_conjs['l2s_w']` → Go: uses SharedStep3 (check l2s_shared.go)
- Py:633 `named_binders_conjs['l2s_s']` → Go: uses SharedStep3
- Py:1013 `named_binders['l2s_g']` → Go: uses SharedStep11 (InsMap)

**symprops/symwhens/symwaits (3 sites) — ALREADY FIXED:**
- Py:891-897 → Go: uses SharedStep7_InstrumentActions (already fixed)

### 3. ivy_temporal.py → temporal/temporal.go

- Py:352 `envprops[label]` → Go: temporal.go:645 `envprops[l]`
- Py:359 `symprops[sym]` → Go: temporal.go:660 `symprops[lg.Key(sym)]`

### 4. ivy_isolate.py → isolate/

- Py:728 `depmap[n]` → Go: isolate/helpers.go:557 `depmap[anc]`
- Py:1120 `export_preconds['ext:' + c]` → find Go equivalent
- Py:2075 `checked_context[callee]` → Go: isolate/iter.go:407
- Py:2077 `checked_context[callee]` → Go: isolate/iter.go:417
- Py:2083 `checked_context[mixed]` → Go: isolate/iter.go:438
- Py:2091 `checked_context[mixed]` → Go: isolate/iter.go:463
- Py:2092 `verified_context[mixed]` → Go: isolate/iter.go:464
- Py:2103 `checked_context[mixed]` → Go: isolate/iter.go:494

### 5. ivy_module.py → module/

- Py:133 `self.hierarchy[name]` → find Go equivalent

### 6. ivy_parser.py → parser/

- Py:370 `self.defined[name]` → find Go equivalent
- Py:390 `self.defined[name][0]` → find Go equivalent
- Py:3647 `autos[key]` → find Go equivalent

### 7. ivy_check.py → check/

- Py:327 `self.eqs[renamed_sym]` → find Go equivalent

### 8. ivy_solver.py → z3bridge/

- Py:1597 `num_by_sort[s]` → Go: solver_compat.go:402 `numBySort[sortName]`

### 9. ivy_fragment.py → fragment/

- Py:312 `strat_map[u]` → Go: fragment.go:180 (already uses getOrCreate pattern)

### 10. ivy_logic_utils.py → logicutil/

- Py:1539 `map2[rhs]`, `map2[lhs]` → find Go equivalent

### 11. ivy_auto_inst.py

- Py:84 `sort_constants[match.map.get(sym.sort,sym.sort)]` → find Go equivalent

### 12. ivy_utils.py → ivyutils/

- Py:459 `m[k]` → find Go equivalent
- Py:500 `m[node]` → Go: ivyutils/graph.go:194 `adj[node]`

### 13. ivy_congclos.py → congclos/

- Py:45,57,78 `self.tab[name]` → find Go equivalent

### 14. ivy_art.py → art/

- Py:447 `fpc[equation.args[0]]` → find Go equivalent

### 15. ivy_alpha.py → alpha/

- Py:191 `index[xr[xc].rep]` → find Go equivalent

### 16. ivy_trace.py → trace/

- Py:277 `self.eqs[sym]` → find Go equivalent

## Execution

For each site:
1. Find the exact Go line
2. Add `if _, ok := m[key]; !ok { m[key] = nil }` before the bracket-read
3. Verify build compiles

## Verification

1. `go build ./...` compiles clean
2. `make golden` — check that divergence continues to advance
