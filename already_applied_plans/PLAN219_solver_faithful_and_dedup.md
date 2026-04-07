# PLAN218: Deduplicate solver/ to use Python-faithful translation infrastructure

**Created:** 2026-04-07 ~21:30 UTC

## Context

PLAN217 created four new Go functions (`formulaToZ3`, `formulaToZ3Closed`, `conjToZ3`, `forall`) that faithfully mirror Python's `formula_to_z3`, `formula_to_z3_closed`, `conj_to_z3`, and `forall()`. Currently only `ClausesToZ3` uses these. The rest of solver/ still uses the old `translateClosed` path and the local `defToConstraint` helper, which diverge from Python in three ways:

1. **Missing type constraints** — `translateClosed` doesn't add `type_constraints`, but Python's `formula_to_z3` does
2. **Wrong closing mechanism** — `translateClosed` closes at the logic level via `CloseFormula` (which runs through `translateQuantifier` adding quant constraints for all formulas including Definitions); Python closes at the Z3 level via `formulaToZ3Closed` (raw `z3.ForAll` for Definitions, `forall()` with quant constraints for others)
3. **Incomplete `defToConstraint`** — the local solver version doesn't handle `Some` (conditional definitions), unlike `il.DefinitionToConstraint` which faithfully ports `Python's Definition.to_constraint()`

## Approach

Replace all `translateClosed` callers (except ImpliesBatch) with `formulaToZ3`, matching Python. Replace `defToConstraint` with `il.DefinitionToConstraint`. Replace `CloseFormula+TranslateNoHash` with `formulaToZ3Closed` in type_constraints handling.

## Detailed Changes

### A. `solver/solver.go` — 6 changes

**A1. Fix `FormulaToZ3` comment (line 264-268)**

The comment says "corresponds to Python's formula_to_z3" but the implementation is just `s.tr.Translate(fmla)` — no HASH, no closing, no type constraints. Fix the comment.

Before:
```go
// FormulaToZ3 converts a single Ivy formula to a Z3 expression.
// Free variables are universally quantified.
// This corresponds to Python's formula_to_z3.
func (s *Solver) FormulaToZ3(fmla lg.Expr) (z3bridge.Expr, error) {
    return s.tr.Translate(fmla)
}
```

After:
```go
// FormulaToZ3 translates a single Ivy formula to Z3 via the core translator.
// This is a raw translate (no HASH, no closing, no type constraints).
// Used by tests and external callers (vmt, alpha).
// For the full Python formula_to_z3 equivalent, use formulaToZ3 (private).
func (s *Solver) FormulaToZ3(fmla lg.Expr) (z3bridge.Expr, error) {
    return s.tr.Translate(fmla)
}
```

**A2. Fix type_constraints in `ClausesToZ3` (lines 314-321)**

Python's `type_constraints` uses `formula_to_z3_closed(fmla)` — match it.

Before:
```go
    for _, tc := range constraints {
        closed := il.CloseFormula(tc)
        ztc, err := s.tr.TranslateNoHash(closed)
```

After:
```go
    for _, tc := range constraints {
        ztc, err := s.formulaToZ3Closed(tc)
```

**A3. Fix type_constraints in `formulaToZ3` (lines 457-462)**

Same pattern as A2.

Before:
```go
    for _, tc := range constraints {
        closed := il.CloseFormula(tc)
        ztc, err := s.tr.TranslateNoHash(closed)
```

After:
```go
    for _, tc := range constraints {
        ztc, err := s.formulaToZ3Closed(tc)
```

**A4. `ClausesImplyFormula` (line 688-689)**

Python `clauses_imply_formula` (line 1620): `formula_to_z3(ivy_logic.Not(fmla2))`

Before:
```go
    negFmla := &lg.Not{Body: fmla2}
    z2, err := s.translateClosed(negFmla)
```

After:
```go
    negFmla := &lg.Not{Body: fmla2}
    z2, err := s.formulaToZ3(negFmla)
```

**A5. `UnsatCore` formulas (lines 726-731)**

Python `unsat_core` (line 686): `formula_to_z3(c)` for each formula.

Before:
```go
    for i, f := range fmlas {
        zf, err := s.translateClosed(f)
```

After:
```go
    for i, f := range fmlas {
        zf, err := s.formulaToZ3(f)
```

**A6. `UnsatCore` definitions (lines 736-739)**

Python `unsat_core` (line 690): `formula_to_z3(d.to_constraint())`. Use `il.DefinitionToConstraint` (handles `Some`, matching Python) instead of local `defToConstraint`.

Before:
```go
    for _, d := range clauses1.Defs {
        constraint := defToConstraint(d)
        zd, err := s.translateClosed(constraint)
```

After:
```go
    for _, d := range clauses1.Defs {
        constraint := il.DefinitionToConstraint(d)
        zd, err := s.formulaToZ3(constraint)
```

**A7. Delete `defToConstraint` (lines 580-591)**

After changes A6 and B2, no callers remain. Delete the local function. All callers now use `il.DefinitionToConstraint` which faithfully ports Python's `Definition.to_constraint()` including `Some` handling.

### B. `solver/model.go` — 4 changes

**B1. `GetSmallModelWithCond` sort constraints (line 237-238)**

Python `get_small_model` (line 1337): `formula_to_z3(sc)`

Before:
```go
                sc := SortSizeConstraint(sort, n)
                zsc, err := s.translateClosed(sc)
```

After:
```go
                sc := SortSizeConstraint(sort, n)
                zsc, err := s.formulaToZ3(sc)
```

**B2. `GetSmallModelWithCond` relation constraints (line 254-255)**

Same pattern.

Before:
```go
                sc := RelationSizeConstraint(rel, n)
                zsc, err := s.translateClosed(sc)
```

After:
```go
                sc := RelationSizeConstraint(rel, n)
                zsc, err := s.formulaToZ3(sc)
```

**B3. `FilterRedundantFacts` definitions (lines 499-501)**

Python `filter_redundant_facts` (line 1494): `formula_to_z3(d.to_constraint())`

Before:
```go
    for _, d := range clauses.Defs {
        constraint := defToConstraint(d)
        zd, err := s.translateClosed(constraint)
```

After:
```go
    for _, d := range clauses.Defs {
        constraint := il.DefinitionToConstraint(d)
        zd, err := s.formulaToZ3(constraint)
```

**B4. `FilterRedundantFacts` positive formulas (lines 509-510)**

Python `filter_redundant_facts` (line 1496): `formula_to_z3(fmla)`

Before:
```go
    for _, f := range posFmlas {
        zf, err := s.translateClosed(f)
```

After:
```go
    for _, f := range posFmlas {
        zf, err := s.formulaToZ3(f)
```

**B5. `FilterRedundantFacts` negative formulas (lines 523-524)**

Python `filter_redundant_facts` (line 1491): `z3.Not(formula_to_z3(c))` for neg_fmlas

Before:
```go
        zn, err := s.translateClosed(nf)
```

After:
```go
        zn, err := s.formulaToZ3(nf)
```

### C. `solver/clauses.go` — 2 changes

**C1. `RemoveDuplicatesClauses` (line 126)**

Python `remove_duplicates_clauses` (line 1055): `formula_to_z3(c)`

Before:
```go
        zf, err := s.translateClosed(f)
```

After:
```go
        zf, err := s.formulaToZ3(f)
```

**C2. `SolverAdd` (line 350)**

Python `solver_add` (line 764): `formula_to_z3(fmla)`

Before:
```go
    zf, err := s.translateClosed(fmla)
```

After:
```go
    zf, err := s.formulaToZ3(fmla)
```

### D. `solver/compat.go` — 1 change

**D1. `ModelIfNone` sort size constraints (line 264)**

Python `model_if_none` (line 1174): `formula_to_z3(sort_size_constraint(sort, sort_size))`

Before:
```go
            zsc, err := s.translateClosed(sc)
```

After:
```go
            zsc, err := s.formulaToZ3(sc)
```

### E. No change: `ImpliesBatch` (solver.go:653-675)

ImpliesBatch maps to Python's `z3_implies_batch` in `z3_utils.py` (line 136), which uses a completely different translation system (`to_z3()` — bare recursive translate, no closing, no HASH, no type constraints). This is a separate module from ivy_solver. Keeping `translateClosed` here preserves the current closing behavior. Full alignment with `z3_implies_batch` would require changing the closing semantics (free vars become existential instead of universal), which is a separate task.

After all changes, `translateClosed` has exactly 2 remaining callers (both in ImpliesBatch). It stays as-is for now.

## Summary of call site changes

| File | Line | Function | Old call | New call | Python equivalent |
|------|------|----------|----------|----------|-------------------|
| solver.go | 316 | ClausesToZ3 tc | `CloseFormula+TranslateNoHash` | `formulaToZ3Closed` | `type_constraints → formula_to_z3_closed` |
| solver.go | 459 | formulaToZ3 tc | `CloseFormula+TranslateNoHash` | `formulaToZ3Closed` | `type_constraints → formula_to_z3_closed` |
| solver.go | 689 | ClausesImplyFormula | `translateClosed` | `formulaToZ3` | `formula_to_z3` (line 1620) |
| solver.go | 727 | UnsatCore fmlas | `translateClosed` | `formulaToZ3` | `formula_to_z3` (line 686) |
| solver.go | 737-8 | UnsatCore defs | `defToConstraint+translateClosed` | `il.DefinitionToConstraint+formulaToZ3` | `formula_to_z3(d.to_constraint())` (line 690) |
| model.go | 238 | GetSmallModel sort | `translateClosed` | `formulaToZ3` | `formula_to_z3` (line 1337) |
| model.go | 255 | GetSmallModel rel | `translateClosed` | `formulaToZ3` | `formula_to_z3` (line 1337) |
| model.go | 500-1 | FilterRedundant def | `defToConstraint+translateClosed` | `il.DefinitionToConstraint+formulaToZ3` | `formula_to_z3(d.to_constraint())` (line 1494) |
| model.go | 510 | FilterRedundant pos | `translateClosed` | `formulaToZ3` | `formula_to_z3` (line 1496) |
| model.go | 524 | FilterRedundant neg | `translateClosed` | `formulaToZ3` | `formula_to_z3` (line 1491) |
| clauses.go | 126 | RemoveDuplicates | `translateClosed` | `formulaToZ3` | `formula_to_z3` (line 1055) |
| clauses.go | 350 | SolverAdd | `translateClosed` | `formulaToZ3` | `formula_to_z3` (line 764) |
| compat.go | 264 | ModelIfNone | `translateClosed` | `formulaToZ3` | `formula_to_z3` (line 1174) |

## Bug fixed

**`defToConstraint` missing `Some` handling**: The local `defToConstraint` (solver.go:580-591) was a simplified version that only handled Iff/Eq. Python's `Definition.to_constraint()` also handles `Some` (conditional definitions with `if_value`/`else_value`). All callers now use `il.DefinitionToConstraint` which handles all cases.

## Files Modified

- `solver/solver.go` — fix FormulaToZ3 comment, fix 2 type_constraints translations, fix ClausesImplyFormula, fix UnsatCore (fmlas + defs), delete defToConstraint
- `solver/model.go` — fix GetSmallModelWithCond (sort + relation), fix FilterRedundantFacts (defs + pos + neg)
- `solver/clauses.go` — fix RemoveDuplicatesClauses, fix SolverAdd
- `solver/compat.go` — fix ModelIfNone

## Verification

```bash
cd ~/go/src/github.com/glycerine/ivy/goivy && go build ./... && go test ./solver/... && go test ./actions/...
```

Then:
```bash
cd ~/ivy/goivy && make golden
```

Expected: All existing tests pass. Golden test may show HASH trace changes in places where `formula_to_z3` now emits HASH traces that `translateClosed` emitted differently (same content but emitted via the new path). Type constraints may cause additional Z3 assertions in some callers.
