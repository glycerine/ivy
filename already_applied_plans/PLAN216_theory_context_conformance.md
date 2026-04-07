# PLAN215: Fix missing defToConstraint traces in UpdateTheory and TheoryContext

**Created:** 2026-04-07 ~13:00 UTC

## Context

PLAN214 (reportCycle + union-find Unify fixes) succeeded. Lines 236013-236016 now match. The golden test (`make golden`) now diverges at line 236017:

```
236016  go : XTRACE: check/isolate_check.go: CheckIsolate back from fragment.CheckFragment().
        py : XTRACE: check/isolate_check.go: CheckIsolate back from fragment.CheckFragment().

236017  go : XTRACE: ops.ToOpenFormula nFmlas=0 nDefs=0
        py : XTRACE: module/clauses.go:262 defToConstraint lhsSort=Boolean resultType=Iff
```

After `CheckFragment` returns, both sides call `TheoryContext()` / `theory_context()` (Go: `isolate_check.go:105`, Python: `ivy_check.py:534`). Python emits `defToConstraint` traces that Go doesn't emit, causing Go to advance to a later trace (`ToOpenFormula`) while Python is still producing `defToConstraint` traces.

## Root Cause Analysis

Python's `theory_context()` (ivy_module.py:189-201) and `update_theory()` (ivy_module.py:149-177) call `to_constraint()` on **every** definition unconditionally. Go's `TheoryContext()` and `UpdateTheory()` only call `defToConstraint()` on a subset.

### Python `update_theory` (ivy_module.py:155-156):
```python
for ldf in self.definitions:
    cnst = ldf.formula.to_constraint()    # <-- ALL definitions (result unused!)
    if all(isinstance(p,il.Variable) for p in ldf.formula.args[0].args):
        if not isinstance(ldf.formula,il.DefinitionSchema):
            ax = ldf.formula
            if isinstance(ax.rhs(),il.Some):
                ax = ax.to_constraint()    # <-- called AGAIN for Some-RHS
                ...
```

### Go `UpdateTheory` (theory.go:44-77) — **missing the unconditional call**:
```go
for _, ldf := range m.Definitions {
    def, isDef := fmla.(*il.Definition)
    if !isDef { continue }             // skips DefinitionSchema entirely
    lhsArgs := getLhsArgs(def)
    if !allVariables(lhsArgs) { continue }  // skips non-EPR
    // ... only calls defToConstraint for Some-RHS ...
}
```

### Python `theory_context` (ivy_module.py:196-200):
```python
for ldf in self.definitions:
    cnst = ldf.formula.to_constraint()    # <-- ALL definitions
    if not all(isinstance(p,il.Variable) for p in ldf.formula.args[0].args):
        non_epr[ldf.formula.defines()] = (ldf, cnst)
```

### Go `TheoryContext` (theory.go:171-187) — **only calls for non-EPR**:
```go
for _, ldf := range m.Definitions {
    def, isDef := ldf.Formula.(*il.Definition)
    if !isDef { continue }             // skips DefinitionSchema
    lhsArgs := getLhsArgs(def)
    if allVariables(lhsArgs) { continue }  // skips EPR — Python still calls to_constraint for these!
    cnst := defToConstraint(def)       // only non-EPR
```

### Additional issue: DefinitionSchema type assertion

`il.DefinitionSchema` is a type alias for `lg.DefinitionSchema` which **embeds** `lg.Definition`. In Go, `fmla.(*il.Definition)` fails for `*il.DefinitionSchema` values (distinct concrete types). Python's `isinstance(ldf.formula, Definition)` succeeds for both. So Go skips `DefinitionSchema` entirely in both functions.

### Trace count mismatch

With 12 definitions, Python produces:
- 12 traces in `update_theory` (line 156, all defs) + K traces (Some-RHS defs again)
- 12 traces in `theory_context` (line 198, all defs)
- Total: 24+K `defToConstraint` traces

Go currently produces:
- K traces in `UpdateTheory` (Some-RHS only)
- M traces in `TheoryContext` (non-EPR only)
- Total: K+M traces

The first missing trace at line 236017 is from `update_theory` line 156 (the unconditional `to_constraint()` call).

## Approach

Match Python's behavior by:
1. In `UpdateTheory`: call `defToConstraint` for every definition unconditionally (matching Python line 156)
2. In `TheoryContext`: call `defToConstraint` for every definition (matching Python line 198), not just non-EPR ones
3. Handle `*il.DefinitionSchema` in both loops by extracting the embedded `Definition`

## File Changes

### 1. `module/theory.go` — `UpdateTheory` (lines 44-77)

Change the loop to handle both `*il.Definition` and `*il.DefinitionSchema`, and add the unconditional `defToConstraint` call:

```go
// Before:
for _, ldf := range m.Definitions {
    fmla := ldf.Formula
    def, isDef := fmla.(*il.Definition)
    if !isDef {
        continue
    }
    lhsArgs := getLhsArgs(def)
    if !allVariables(lhsArgs) {
        continue
    }
    if _, isSchema := fmla.(*il.DefinitionSchema); isSchema {
        continue
    }
    if _, isSome := def.Rhs.(*il.Some); isSome {
        ax := defToConstraint(def)
        ...

// After:
for _, ldf := range m.Definitions {
    fmla := ldf.Formula

    // Extract Definition — handle both *Definition and *DefinitionSchema.
    var def *il.Definition
    var isSchema bool
    if d, ok := fmla.(*il.Definition); ok {
        def = d
    } else if ds, ok := fmla.(*il.DefinitionSchema); ok {
        def = &ds.Definition
        isSchema = true
    } else {
        continue
    }

    // Python: cnst = ldf.formula.to_constraint()
    // Called unconditionally for ALL definitions to match Python trace output.
    defToConstraint(def)

    lhsArgs := getLhsArgs(def)
    if !allVariables(lhsArgs) {
        continue
    }
    if isSchema {
        continue
    }
    if _, isSome := def.Rhs.(*il.Some); isSome {
        ax := defToConstraint(def)
        ...
```

### 2. `module/theory.go` — `TheoryContext` (lines 171-187)

Change the loop to call `defToConstraint` for all definitions, not just non-EPR ones:

```go
// Before:
nonEPR := make(map[lg.NodeKey]nonEPREntry)
for _, ldf := range m.Definitions {
    def, isDef := ldf.Formula.(*il.Definition)
    if !isDef {
        continue
    }
    lhsArgs := getLhsArgs(def)
    if allVariables(lhsArgs) {
        continue
    }
    cnst := defToConstraint(def)
    defines := def.Defines()
    if defines != nil {
        nonEPR[lg.Key(defines)] = nonEPREntry{ldf: ldf, constraint: cnst}
    }
}

// After:
nonEPR := make(map[lg.NodeKey]nonEPREntry)
for _, ldf := range m.Definitions {
    // Extract Definition — handle both *Definition and *DefinitionSchema.
    var def *il.Definition
    if d, ok := ldf.Formula.(*il.Definition); ok {
        def = d
    } else if ds, ok := ldf.Formula.(*il.DefinitionSchema); ok {
        def = &ds.Definition
    } else {
        continue
    }

    // Python: cnst = ldf.formula.to_constraint()
    // Called for ALL definitions (result used only for non-EPR).
    cnst := defToConstraint(def)

    lhsArgs := getLhsArgs(def)
    if allVariables(lhsArgs) {
        continue
    }
    defines := def.Defines()
    if defines != nil {
        nonEPR[lg.Key(defines)] = nonEPREntry{ldf: ldf, constraint: cnst}
    }
}
```

## Verification

```bash
cd ~/ivy/goivy && make golden
```

Expected: The test advances past line 236017. The `defToConstraint` traces from `UpdateTheory` and `TheoryContext` should now match Python's output, and the subsequent `ToOpenFormula` traces should align.

## Notes

- The `cnst = ldf.formula.to_constraint()` call at Python line 156 in `update_theory` is dead code (the result is never used). But it produces a trace, so Go must replicate it.
- The `DefinitionSchema` handling is needed because `*il.DefinitionSchema` doesn't satisfy Go's `*il.Definition` type assertion (Go has no structural subtyping for concrete types), even though Python's `isinstance(x, Definition)` returns True for DefinitionSchema.
- `il.Definition` and `il.DefinitionSchema` are type aliases for `lg.Definition` and `lg.DefinitionSchema` respectively (see `ivylogic/formula.go:110,117`).
