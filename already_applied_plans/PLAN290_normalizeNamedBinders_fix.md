# Fix NormalizeNamedBinders leaf/unchanged shortcuts that diverge from Python

Created: 2026-04-13 ~16:30 UTC

## Context

At xtrace line 457627, `NormalizeNamedBinders` diverges when processing an Atom with no args (`invar228`, `terms:[]`):

- **Python**: Always calls `ast.clone(args)` and traces `"cloned HASH canon=..."`
- **Go**: Has a leaf shortcut that returns the original node unchanged, tracing `"leaf"`

Python has exactly 4 exit paths: `const`, `binder`, `app`, `cloned`. Go has 6: `const`, `binder`, `app`, `leaf`, `unchanged`, `cloned`. The `leaf` and `unchanged` paths are Go-only optimizations that don't exist in Python.

## Fix

**File:** `logicutil/logic_utils.go` lines 929-949

Replace the general-case block (after the Apply check) with code that always clones, matching Python:

```go
	// General case: recurse into children and always clone (matching Python).
	children := n.Args()
	newChildren := make([]ast.Node, len(children))
	for i, c := range children {
		newChildren[i] = NormalizeNamedBinders(c, names)
	}
	result := n.Clone(newChildren)
	xtracer.Trace("ilu.normalizeNamedBinders EXIT type=%s cloned HASH canon=%s", iu.ShortTypeName(result), result.Canon())
	return result
```

This removes:
- The `len(children) == 0` leaf shortcut (lines 930-932)
- The `changed` tracking and `!changed` unchanged shortcut (lines 935, 939-945)

## Why this is safe

- `Atom.Clone([]Node{})` works correctly — creates a new Atom with empty Terms (`ast/ast.go:350-354`)
- All `Clone` implementations handle empty/nil args
- Matches Python's unconditional `ast.clone(args)` behavior exactly

## Verification

Run the golden test that exercises this code path:
```
cd /Users/jaten/ivy/goivy && go test -run TestOrdLive -v -timeout 600s
```
The test should advance past line 457627 (previously the divergence point).
