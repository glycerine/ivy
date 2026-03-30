# Plan: Fix cloneLF Shortcut Bypassing Proper LabeledFormula.Clone()

**Created**: 2026-03-30 06:15

## Context

Golden test diverges at line 150242:
```
150242  go : XTRACE: check.CreateIsolate EXIT name=cf_live
        py : XTRACE: ast.LF.clone PRESERVE origid=207 counter=1822
```

Go exits CreateIsolate while Python is still emitting `ast.LF.clone PRESERVE` traces. The root cause: Go's `isolate/isolate.go:794` uses a shortcut `cloneLF()` function (line 1339-1342) that does a shallow struct copy (`cp := *lf; return &cp`) instead of calling the proper `LabeledFormula.Clone()` method.

**Python** (`ivy_isolate.py:1128`): `p = p.clone(p.args)` — calls the full `LabeledFormula.clone()` which:
1. Creates a proper clone via `AST.clone()`
2. Decrements `lf_counter` (when `always_clone_with_fresh_id` is False)
3. Preserves the original ID (`res.id = self.id`)
4. Emits `xtracer.trace("ast.LF.clone PRESERVE origid=%d counter=%d" ...)`

**Go** (`isolate/isolate.go:794`): `cp := cloneLF(p)` — shallow struct copy that:
1. Does NOT call `LabeledFormula.Clone()`
2. Does NOT decrement `LfCounter`
3. Does NOT emit any XTRACE
4. Creates a copy that shares underlying slice/pointer fields with the original

The proper `Clone()` method exists at `ast/decl_ast.go:71-90` and does everything correctly.

## Fix

### Step 1: Replace `cloneLF(p)` with proper `Clone()` call

**File**: `isolate/isolate.go`, line 794

Change:
```go
cp := cloneLF(p)
```
To:
```go
cp := p.Clone(p.Args()).(*ast.LabeledFormula)
```

This calls the proper `LabeledFormula.Clone()` method (`ast/decl_ast.go:71`) which:
- Creates a deep clone via `cloneInternal()`
- Manages `LfCounter` properly
- Emits the `ast.LF.clone PRESERVE` XTRACE trace
- Copies all semantic fields (Temporal, Explicit, Definition, Assumed, Unprovable)

### Step 2: Remove the dead `cloneLF` helper

**File**: `isolate/isolate.go`, lines 1339-1342

Delete:
```go
func cloneLF(lf *ast.LabeledFormula) *ast.LabeledFormula {
	cp := *lf
	return &cp
}
```

It has only one call site (line 794) which we just replaced.

## Files to modify

1. **`isolate/isolate.go`** — Replace `cloneLF(p)` call at line 794; delete `cloneLF` function at lines 1339-1342

## What stays unchanged

- `LabeledFormula.Clone()` at `ast/decl_ast.go:71-90` — already correct
- `LabeledFormula.Args()` at `ast/decl_ast.go:70` — returns `[Label, Formula]` as needed
- All other isolate processing logic — execution order is already correct

## Verification

`go build ./...` then `cd ~/goivy && make golden`
