# Fix mc/transforms.go defToConstraint: wrong semantics + missing XTRACE

Created: 2026-04-30 05:32 UTC

## Context

The rfn test (`TestRfnAbsIso`) diverges at XTRACE line 225120. The last matching trace is `mc.ToAiger postAddPostAxioms`. Python then emits `defToConstraint` traces before `postQelim`; Go skips straight to `postQelim`.

Root cause: `mc/transforms.go:250` defines a local `defToConstraint` with two bugs:

1. **Wrong semantics** — always returns `Eq{Lhs, Rhs}`. Python's `Definition.to_constraint()` (and Go's correct `il.DefinitionToConstraint`) returns `Eq` for individuals but `Iff` for Boolean-sorted definitions, and handles `Some` RHS.
2. **Missing XTRACE** — Python emits `XTRACE: module/clauses.go:262 defToConstraint lhsSort=... resultType=...` for each call. The mc version emits nothing.

The correct implementation already exists at `module/clauses.go:259` (delegates to `il.DefinitionToConstraint` + trace). The mc copy is a simplified duplicate that diverges.

## Call sites of the wrong function

1. `mc/toaiger.go:250` — Step 4a: convert non-finite defs to constraints (this is the failing trace)
2. `mc/transforms.go:148` — ToTableLookup: convert invariant defs to constraints

## Fix

### Step 1: Add imports to `mc/transforms.go`

Add to the import block:
```go
iu "github.com/glycerine/ivy/goivy/ivyutils"
"github.com/glycerine/ivy/goivy/xtracer"
```

### Step 2: Replace `mc/transforms.go:249-252`

Replace:
```go
// defToConstraint converts a definition to a constraint (equality).
func defToConstraint(def *il.Definition) lg.Expr {
	return &lg.Eq{T1: def.Lhs, T2: def.Rhs}
}
```

With:
```go
// defToConstraint converts a definition to a constraint formula.
func defToConstraint(def *il.Definition) lg.Expr {
	result := il.DefinitionToConstraint(def)
	if xtracer.Enabled {
		lhsSort := def.Lhs.NodeSort()
		xtracer.Trace("module/clauses.go:262 defToConstraint lhsSort=%v resultType=%v", lhsSort, iu.ShortTypeName(result))
	}
	return result
}
```

The trace string `"module/clauses.go:262"` is kept as-is — it's a semantic label matching Python's trace output, not a file location.

### Files modified

- `/Users/jaten/ivy/goivy/mc/transforms.go` — fix defToConstraint + add imports

### Files NOT modified (no changes needed)

- `mc/toaiger.go` — call site unchanged (same function name/signature)
- `module/clauses.go` — already correct (reference implementation)
- `ivylogic/constraint.go` — already correct (DefinitionToConstraint)

## Verification

```bash
cd ~/ivy/goivy && make test
```

The `TestRfnAbsIso` test should now pass line 225120 with matching `defToConstraint` traces before `postQelim`.
