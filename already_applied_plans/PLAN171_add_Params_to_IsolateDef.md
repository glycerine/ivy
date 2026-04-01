# Add IsolateDef.Params() Method and Fix Strip Map Building

**Created**: 2026-04-01 (updated 2026-04-01 18:15)

## Context

Go's `*ast.IsolateDef` does not implement the `isolateParamProvider` interface (requires `Params() []*lg.Const`). Python's `IsolateDef` has `params()` returning `self.args[0].args` — the parameter atoms from the isolate's name node. This causes `StripIsolateParams` to skip the entire parameter-handling path (variable substitution, `NumIsolateParams`, unbound param validation), and ultimately produces an empty `stripMap` while Python's has 10 entries.

Root cause chain:
1. `IsolateDef` lacks `Params()` → `isolateParamProvider` assertion fails
2. `NumIsolateParams` stays 0
3. Strip map entries from impl_mixins propagation (Step 3) are skipped because Go guards `if len(mixerParams) > 0` while Python always assigns
4. Result: `stripMap=0` in Go vs `stripMap=10` in Python → `StripLabeledFormulas` early-returns → 2 missing `LF.clone PRESERVE` traces

## Changes

### 1. Add `Params()` to `*ast.IsolateDef`

**File**: `~/ivy/goivy/ast/decl_ast.go`

Python's `params()` returns `self.args[0].args` — the Terms of the first Elem (the name Atom). Add the equivalent Go method:

```go
// Params returns the isolate parameters (terms of the name atom).
// Python: def params(self): return self.args[0].args
func (i *IsolateDef) Params() []Node {
    if len(i.Elems) == 0 {
        return nil
    }
    if a, ok := i.Elems[0].(*Atom); ok {
        return a.Terms
    }
    return nil
}
```

### 2. Change `isolateParamProvider` to use `[]ast.Node`

**File**: `~/ivy/goivy/isolate/strip.go` (line ~525)

Current:
```go
type isolateParamProvider interface {
    Params() []*lg.Const
}
```

Change to:
```go
type isolateParamProvider interface {
    Params() []ast.Node
}
```

This matches what `IsolateDef.Params()` returns and what Python provides (AST nodes, not logic constants).

### 3. Update `StripIsolateParams` to work with `[]ast.Node`

**File**: `~/ivy/goivy/isolate/strip.go`

**Step 1 (variable substitution, ~line 551)**: Change `pp.Params()` usage from `*lg.Const` to `ast.Node`:
- `il.IsVariable(p)` → check if node is `*ast.Variable` (AST level)
- `p.Name` → extract via `(*ast.Atom).Rep` or `(*ast.Variable).Rep`
- `p.NodeSort()` → extract via `(*ast.Atom).ASort` or `(*ast.Variable).VSort`

Python checks `isinstance(p, ivy_ast.Variable)`. Go equivalent: type-assert to `*ast.Variable`.

**Step 2 (NumIsolateParams, ~line 576)**: No change needed — `len(pp.Params())` works for `[]ast.Node`.

**Step 3 (validation, ~line 583)**: Change `p.Name` to extract name from `*ast.Atom`:
```go
for _, p := range pp.Params() {
    if a, ok := p.(*ast.Atom); ok {
        ips[a.Rep] = true
    }
}
```

### 4. Fix Step 3 strip map propagation guard

**File**: `~/ivy/goivy/isolate/strip.go` (~line 643)

Current Go:
```go
if len(mixerParams) > 0 {
    stripMap[m.Mixee()] = mixerParams
}
```

Python (no guard):
```python
strip_map[m.mixee()] = strip_params
```

Fix: Remove the `if len(mixerParams) > 0` guard — always assign, matching Python:
```go
stripMap[m.Mixee()] = mixerParams
```

### 5. Fix `StripLabeledFormulas` empty-stripMap check

**File**: `~/ivy/goivy/isolate/strip.go` (~line 449)

Current:
```go
if len(stripMap) == 0 {
    return lfs
}
```

After fix #4, `stripMap` will have entries with empty slices (like Python). `len(stripMap)` will be >0 even if all values are empty slices. This matches Python's `if not strip_map: return` which is falsy only for an empty dict, not a dict with empty-list values.

No change needed here — the fix in #4 makes this work correctly.

### Files to modify
- `~/ivy/goivy/ast/decl_ast.go` — add `Params()` method
- `~/ivy/goivy/isolate/strip.go` — change interface, update usages, remove guard

## Verification

```bash
cd ~/ivy/goivy && go build ./... && make golden
```

Check that `stripMap` matches between Go and Python, and trace divergence advances past 156567.
