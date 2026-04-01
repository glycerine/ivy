# Fix: Eliminate non-deterministic map iteration — Python + Go changes

**Created**: 2026-04-02 00:15
**Updated**: 2026-04-02 01:00

## Context

The golden test (`make golden`) diverges at ~line 157285, but the divergence point is **non-deterministic** — different runs produce different divergence points. Root cause: Go's randomized `map` iteration order vs Python 3.7+ dict insertion-order.

Four observed divergence points:
```
157309: go=Not nargs=1          py=Implies nargs=2
157285: go=Sequence.int_update  py=LocalAction.int_update
157299: go=LocalAction          py=AssignAction
157289: go=LocalAction          py=AssignAction
```

## Strategy: Make BOTH sides deterministic via dict/InsMap

Convert Python `set()` → `dict` (insertion-ordered) for `public_actions` and `hierarchy` values. Then on Go side, use `InsMap` for all fields that are Python dicts. This gives both sides identical insertion-order semantics with zero sorting needed.

## Phase 0: Python changes (prerequisite)

### 0A. `~/ivy/pyivy/ivy/ivy/ivy_module.py` — Convert `public_actions` from set to dict

```python
# WAS:  self.public_actions = set()
# NOW:  self.public_actions = {}
```

API migration in all Python files:

| set pattern | dict equivalent |
|---|---|
| `s.add(x)` | `d[x] = True` |
| `s.remove(x)` | `del d[x]` |
| `s.discard(x)` | `d.pop(x, None)` |
| `s.clear()` | `d.clear()` (same) |
| `s.update(iterable)` | `d.update({x: True for x in iterable})` |
| `x in s` | `x in d` (same) |
| `for x in s` | `for x in d` (same, now insertion-ordered) |
| `sorted(s)` | `sorted(d)` (same) or just `d` if order is enough |
| `len(s)` | `len(d)` (same) |
| `if s:` | `if d:` (same) |

**Files to change (~15 call sites):**
- `ivy_module.py:52` — init: `self.public_actions = {}`
- `ivy_compiler.py:1638` — `.add(name)` → `[name] = True`
- `ivy_isolate.py:1394-1395` — `.clear()` + `.update(exported)` → `.clear()` + `update({x: True for x in exported})`
- `ivy_isolate.py:1664,1668` — `.remove(name)` → `del [name]`
- `ivy_isolate.py:1906-1912` — `.clear()` + `.add()` loop → `.clear()` + `[x] = True` loop
- `ivy_isolate.py:1928` — `.add(ext)` → `[ext] = True`
- `ivy_fragment.py:519` — `for name in im.module.public_actions` — no change needed
- `ivy_to_cpp.py:1476,2989,5062` — `if name in` — no change needed
- `ivy_art.py:138` — `if a in self.public_actions` — no change needed
- All `sorted(mod.public_actions)` calls — no change needed (sorted() works on dicts)

### 0B. `~/ivy/pyivy/ivy/ivy/ivy_module.py` — Convert `hierarchy` values from set to dict

```python
# WAS:  self.hierarchy = defaultdict(set)
# NOW:  self.hierarchy = defaultdict(dict)
```

API migration:

| set pattern | dict equivalent |
|---|---|
| `hierarchy[k].add(child)` | `hierarchy[k][child] = True` |
| `child in hierarchy[k]` | `child in hierarchy[k]` (same) |
| `for child in hierarchy[k]` | `for child in hierarchy[k]` (same) |

**Files to change (~10 call sites):**
- `ivy_module.py:47` — init: `self.hierarchy = defaultdict(dict)`
- `ivy_module.py:126` — `.add(suff)` → `[suff] = True`
- `ivy_module.py:128` — `.add(name)` → `[name] = True`
- `ivy_isolate.py:740` — `if suff in l and preferred in l` — no change
- `ivy_isolate.py:763` — `if ns in l` — no change
- `ivy_actions.py:537` — `for x in domain.hierarchy[n]` — no change
- `ivy_actions.py:1189` — `for child in domain.hierarchy[n]` — no change
- `ivy_isolate.py:2160` — `for child in mod.hierarchy[name]` — no change

### 0C. Verify Python still passes

```bash
cd ~/ivy/pyivy && python -m pytest ivy/ivy/test_*.py
```

And re-generate the golden reference:
```bash
cd ~/ivy/goivy && make golden-py
```

## Phase 1: Go — Convert Module fields from `map` to `InsMap`

With Python now using dicts everywhere, Go should use InsMap everywhere to match.

**Python types → Go types (target state):**

| Field | Python type (after Phase 0) | Go type (target) |
|-------|-----------|------------|
| `actions` | dict | `*InsMap[string, Action]` (already done) |
| `before_export` | dict | `*InsMap[string, Action]` |
| `relations` | dict | `*InsMap[string, lg.Sort]` |
| `functions` | dict | `*InsMap[string, lg.Sort]` |
| `hierarchy` (outer) | defaultdict(dict) | `*InsMap[string, *InsMap[string, bool]]` |
| `public_actions` | dict (after 0A) | `*InsMap[string, bool]` |

### 1A. `module/module.go` — Convert `Relations`

```go
// WAS:  Relations map[string]lg.Sort
// NOW:  Relations *iu.InsMap[string, lg.Sort]
```

Update sites (12 occurrences across 6 files):
- `module/module.go` — struct def, `NewModule()`, `Clone()`
- `compiler/decl.go` — 3 sites: reads/writes
- `interp/solo7_interp_test.go` — 3 sites
- `interp/phase4.go` — iteration
- `actions/phase3.go` — 1 site
- `gogen/generator.go` — iteration
- `gogen/generator_test.go` — 1 site
- `trace/trace.go:384` — iteration
- `ivydump/ivydump.go:30` — iteration

### 1B. `module/module.go` — Convert `Functions`

```go
// WAS:  Functions map[string]lg.Sort
// NOW:  Functions *iu.InsMap[string, lg.Sort]
```

Update sites (5 occurrences across 3 files):
- `module/module.go` — struct def, init, Clone
- `compiler/decl.go` — 3 sites
- `interp/phase4.go:63` — **critical** iteration for variable renaming
- `gogen/generator.go:287` — iteration
- `gogen/generator_test.go` — 1 site
- `trace/trace.go:389` — iteration
- `ivydump/ivydump.go:38` — iteration

### 1C. `module/module.go` — Convert `BeforeExport`

```go
// WAS:  BeforeExport map[string]Action
// NOW:  BeforeExport *iu.InsMap[string, Action]
```

Update sites (small — ~4 files):
- `module/module.go` — struct def, init, Clone
- `module/canonize.go:81` — iteration
- `fragment/fragment.go:812` — iteration
- `isolate/isolate.go:624` — Set call

### 1D. `module/module.go` — Convert `PublicActions`

```go
// WAS:  PublicActions map[string]bool
// NOW:  PublicActions *iu.InsMap[string, bool]
```

This replaces all the "sort before iteration" fixes from the old plan. With InsMap, iteration is insertion-ordered automatically.

Update sites (~15 files):
- `module/module.go` — struct def, init, Clone
- `check/check.go:757` — can remove manual sort
- `actions/action.go:2011` — can remove manual sort
- `isolate/create.go:290,317,411` — can remove manual sort
- `temporal/temporal.go:256` — can remove manual sort
- `fragment/fragment.go:819` — iteration (was unsorted, now auto-ordered)
- `trace/trace.go:570` — iteration
- `bmc/bmc.go:200` — iteration
- `tactics/tactics.go:290` — iteration
- `printer/printer.go:154` — can remove manual sort
- `mc/toaiger.go:680` — can remove manual sort
- `isolate/batch_fixes_test.go:665` — iteration

### 1E. `module/module.go` — Convert `Hierarchy`

```go
// WAS:  Hierarchy map[string]map[string]bool
// NOW:  Hierarchy *iu.InsMap[string, *iu.InsMap[string, bool]]
```

Update sites (29 occurrences across 12 files):
- `module/module.go` — struct def, init, Clone (9 sites)
- `compiler/phase6.go` — 1 site
- `compiler/helpers.go` — 1 site
- `compiler/batch_f_test.go` — 3 sites
- `isolate/helpers.go` — 2 sites
- `isolate/isolate.go` — 2 sites
- `isolate/isolate_test.go` — 2 sites
- `isolate/iter.go` — 1 site
- `isolate/create.go` — 1 site
- `module/module_test.go` — 4 sites
- `actions/update.go` — 2 sites
- `actions/transforms.go` — 1 site

## Phase 2: Fix remaining internal map iterations

Other map fields and internal maps that may need conversion:
- `Postconds`, `Predicates`, `Privates`, `VPrivates` — convert if iterated on critical path
- `interp/phase4.go:455` — sort entries from map conversion
- `transrel/transrel.go:1836` — `ComposeMaps` iteration

These are lower priority — fix after Phase 1 if golden test still shows issues.

## InsMap API migration cheatsheet

| Go map pattern | InsMap equivalent |
|---|---|
| `m[k]` (read) | `m.Get(k)` |
| `v, ok := m[k]` | `v, ok := m.Get2(k)` |
| `m[k] = v` | `m.Set(k, v)` |
| `delete(m, k)` | `m.Delkey(k)` |
| `len(m)` | `m.Len()` |
| `for k, v := range m` | `for k, v := range m.All()` |
| `m = make(map[K]V)` | `m = iu.NewInsMap[K, V]()` |
| `m = make(map[K]V, n)` | `m = iu.NewInsMap[K, V]()` (no capacity hint) |

## Call site counts

- `Relations[` — 12 occurrences across 6 files
- `Functions[` — 5 occurrences across 3 files
- `BeforeExport[` — 1 occurrence in 1 file
- `PublicActions` — ~15 files
- `Hierarchy[` — 29 occurrences across 12 files
- `len(...)` calls — 10 occurrences across 4 files

Total: ~70 mechanical Go changes + ~25 Python changes.

## Implementation order

1. **Phase 0A**: Python `public_actions` set → dict
2. **Phase 0B**: Python `hierarchy` values set → dict
3. **Phase 0C**: Verify Python, regenerate golden reference
4. **Phase 1A**: Go `Relations` → InsMap
5. **Phase 1B**: Go `Functions` → InsMap
6. **Phase 1C**: Go `BeforeExport` → InsMap
7. **Phase 1D**: Go `PublicActions` → InsMap
8. **Phase 1E**: Go `Hierarchy` → InsMap (largest, 29 sites)
9. **Phase 2**: Fix remaining internal map iterations if needed

## Verification

After each phase, build and test:
```bash
cd ~/ivy/goivy && go build ./... && go test ./...
```

After all phases, confirm determinism:
```bash
cd ~/ivy/goivy && make golden && make golden && make golden
```

All three runs must diverge at the **same** line number (proving determinism). The divergence content may still differ from Python (semantic bugs), but the line number must be stable.
