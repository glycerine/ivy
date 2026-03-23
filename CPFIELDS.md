# Fix: "immutable symbol assigned: cfabric.rd_pio_fair" — LabeledFormula metadata lost during compilation

**Created: 2026-03-22**

## Context

Running `goivy_check isolate=cf_live ord_live.ivy` fails with:
```
error: immutable symbol assigned: (Symbol name:cfabric.rd_pio_fair sort:(BooleanSort))
```

Python passes this check. The interference check at `compiler/ivy_compile.go:1423` only checks **non-temporal** axioms (`if !lf.IsTemporal()`). The axiom in question is:
```
explicit temporal axiom [cf_rd_pio_fair]
    globally eventually cfabric.rd_pio_fair
```
This is temporal — it should be **skipped** by the interference check.

## Root Cause

When the Go compiler processes axioms/properties, it creates a **new** `LabeledFormula` but only copies `Formula` and `Lineno`, dropping all metadata including `Temporal`.

**File: `compiler/decl.go` lines 528-531** (Axiom method):
```go
mlf := &ast.LabeledFormula{
    Formula: compiled,
    Lineno:  lf.GetLineno().Line,
}
```

The source `lf` has `Temporal = &true` (set by parser at `parser/decl.go:2370`), but the new `mlf` has `Temporal = nil` → `IsTemporal()` returns false → interference check fires on a temporal axiom.

**Python doesn't have this bug** because `ax.compile()` (ivy_ast.py) uses `clone()` which preserves all metadata:
```python
def axiom(self, ax):
    cax = ax.compile()  # clone preserves temporal, explicit, etc.
    self.domain.labeled_axioms.append(cax)
```

## Fix

Copy all metadata fields from the source `lf` when creating the module's `LabeledFormula`. Three locations need fixing:

### 1. `compiler/decl.go` — `DomainSetup.Axiom()` (line 528)

```go
mlf := &ast.LabeledFormula{
    Formula:      compiled,
    Lineno:       lf.GetLineno().Line,
    Label:        lf.Label,
    ID:           lf.ID,
    Temporal:     lf.Temporal,
    Explicit:     lf.Explicit,
    IsDefinition: lf.IsDefinition,
    Assumed:      lf.Assumed,
    Unprovable:   lf.Unprovable,
    Annot:        lf.Annot,
}
```

### 2. `compiler/decl.go` — `DomainSetup.Property()` (line 547)

Same pattern — copy all metadata from `lf`.

### 3. `compiler/ivy_compile.go` — `ConjSetup.ProcessDecls()` (line 432)

Currently only copies `Label`. Must also copy `Temporal`, `Explicit`, `ID`, etc.

## Files to Modify

| File | Change |
|------|--------|
| `compiler/decl.go` | Axiom() and Property(): copy all LabeledFormula metadata fields |
| `compiler/ivy_compile.go` | ConjSetup conjecture handling: copy all LabeledFormula metadata fields |

## Verification

1. `go test ./...` passes
2. `goivy_check_xtrace isolate=cf_live ord_live.ivy` no longer fails with "immutable symbol assigned: cfabric.rd_pio_fair"
3. The xtracer should show the temporal axiom being SKIPPED (no FAIL line for cfabric.rd_pio_fair)
