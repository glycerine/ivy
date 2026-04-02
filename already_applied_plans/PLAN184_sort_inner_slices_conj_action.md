# Plan: Fix ConjActions inner slice ordering divergence

**Created**: 2026-04-02

## Context

`make golden` fails at line 147715. The `conjActions` field keys match between Go and Python, but the elements within each inner string slice/set have different ordering.

- Python: `["dramc.step_rd" "ifabric.step" "cfabric.step" "memc.step" ...]`
- Go: `["arm.step_south" "dramc.step_wr" "ifabric.step" "dramc.step_rd" ...]`

## Root Cause

Both sides store **unordered** data:
- **Python** (`ivy_compiler.py:2400`): `conj_actions[label] = set(...)` — a `set`, not a list
- **Go** (`ivy_compile.go:1773-1777`): builds `[]string` by iterating `map[string]bool` — random Go map order

The canon helpers (`canonStringSliceMap` / `_canon_string_slice_map`) sort the outer map keys but NOT the inner elements.

## Fix — 2 changes

### 1. Go: sort inner slices in `canonStringSliceMap`

**File**: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/module/canon.go:552-557`

Copy and sort the inner `[]string` before serializing:

```go
// BEFORE:
for _, k := range keys {
    var ss []string
    for _, s := range m[k] {

// AFTER:
for _, k := range keys {
    vals := make([]string, len(m[k]))
    copy(vals, m[k])
    sort.Strings(vals)
    var ss []string
    for _, s := range vals {
```

### 2. Python: sort inner sets in `_canon_string_slice_map`

**File**: `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_module.py:727`

```python
# BEFORE:
ss = ['"%s"' % s for s in m[k]]

# AFTER:
ss = ['"%s"' % s for s in sorted(m[k])]
```

## Verification

```bash
cd ~/ivy/goivy && make test && make golden
```
