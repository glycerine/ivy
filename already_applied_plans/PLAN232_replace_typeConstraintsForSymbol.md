# Fix: Create proper typeConstraints method matching Python's type_constraints()

**Created:** 2026-04-08 ~19:10 UTC

## Context

The golden test diverges at line 236241:
```
236241  go : XTRACE: ivy_solver.py:688 formula_to_z3_closed() ENTER type=Definition
        py : XTRACE: ivy_solver.py:603 type_constraints() ENTER nsyms=14
```

Go's `formulaToZ3()` inlines the type constraint logic (solver.go:416-427) — looping over symbols and calling `typeConstraintsForSymbol` directly — instead of calling a proper `typeConstraints()` function. This means:
1. No `type_constraints() ENTER` trace is emitted (Python emits it at ivy_solver.py:604)
2. The `HandleRangeSorts` flag is not toggled for range sort constraints (Python sets `handle_range_sorts = False` at line 617 before translating range constraints, then restores it at line 630)

The same inline pattern exists in `ClausesToZ3()` (solver.go:285-297) for the clauses-level `type_constraints(used_symbols_clauses(clauses))` call (Python line 645).

## Python source of truth

**`type_constraints(syms)`** (ivy_solver.py:603-631):
1. Emits trace with nsyms count
2. **Pass 1** — nat symbols: filters for `interp == 'nat'`, builds `¬(x < 0)` AST, translates via `formula_to_z3_closed()`
3. **Pass 2** — range sort symbols: sets `handle_range_sorts = False`, filters for `isinstance(interp, RangeSort)`, builds `¬(x < lb)` and `¬(ub < x)` ASTs, translates each via `formula_to_z3_closed()`, restores `handle_range_sorts = True`
4. Returns list of Z3 expressions

**Called from two places:**
- `formula_to_z3(fmla)` line 725: `type_constraints(used_symbols_ast(fmla))`
- `clauses_to_z3(clauses)` line 645: `type_constraints(used_symbols_clauses(clauses))`

## Plan

### 1. Split `typeConstraintsForSymbol` into nat-only and range-only helpers

Rename/split the existing `typeConstraintsForSymbol` (solver.go:312-384) into:

- **`buildConstraintTerm(sym)`** — shared logic to build the term (sym itself for first-order, sym applied to fresh variables for function sorts). Lines 336-349 of current code.
- **`natConstraintForSymbol(sym)`** — returns `[]lg.Expr` with `¬(term < 0)` if sym has nat interp, else nil. Lines 354-363 of current code.
- **`rangeConstraintsForSymbol(sym)`** — returns `[]lg.Expr` with `¬(term < lb)` and `¬(ub < term)` if sym has RangeSort interp, else nil. Lines 366-381 of current code.

### 2. Create `typeConstraints` method on `*Solver`

New method matching Python's `type_constraints()`:

```go
// typeConstraints generates type constraints for nat and range sorts.
// Matches Python type_constraints (ivy_solver.py:603-631).
func (s *Solver) typeConstraints(syms []*lg.Const) ([]Expr, error) {
    xtracer.Trace("ivy_solver.py:603 type_constraints() ENTER nsyms=%d", len(syms))

    var res []Expr

    // Pass 1: nat sort constraints (Python lines 605-615)
    for _, sym := range syms {
        for _, tc := range s.natConstraintForSymbol(sym) {
            ztc, err := s.formulaToZ3Closed(tc)
            if err != nil {
                continue
            }
            res = append(res, ztc)
        }
    }

    // Pass 2: range sort constraints (Python lines 616-630)
    saved := s.HandleRangeSorts
    s.HandleRangeSorts = false
    for _, sym := range syms {
        for _, tc := range s.rangeConstraintsForSymbol(sym) {
            ztc, err := s.formulaToZ3Closed(tc)
            if err != nil {
                continue
            }
            res = append(res, ztc)
        }
    }
    s.HandleRangeSorts = saved

    return res, nil
}
```

### 3. Replace inline code in `formulaToZ3` (solver.go:414-427)

**Before:**
```go
usedSyms := lu.UsedConstantsList(fmla)
var tcs []Expr
for _, sym := range usedSyms {
    constraints := s.typeConstraintsForSymbol(sym)
    for _, tc := range constraints {
        ztc, err := s.formulaToZ3Closed(tc)
        if err != nil { continue }
        tcs = append(tcs, ztc)
    }
}
```

**After:**
```go
usedSyms := lu.UsedConstantsList(fmla)
tcs, tcErr := s.typeConstraints(usedSyms)
if tcErr != nil {
    xtracer.Trace("formula_to_z3: Z3 error on type_constraints: %v type=%v", tcErr, iu.ShortTypeName(fmla))
    return Expr{}, tcErr
}
```

### 4. Replace inline code in `ClausesToZ3` (solver.go:285-297)

**Before:**
```go
usedSyms := clauses.Symbols()
for _, sym := range usedSyms {
    constraints := s.typeConstraintsForSymbol(sym)
    for _, tc := range constraints {
        ztc, err := s.formulaToZ3Closed(tc)
        if err != nil { continue }
        exprs = append(exprs, ztc)
    }
}
```

**After:**
```go
symMap := clauses.Symbols()
clauseSyms := make([]*lg.Const, 0, len(symMap))
for _, sym := range symMap {
    clauseSyms = append(clauseSyms, sym)
}
tcs, tcErr := s.typeConstraints(clauseSyms)
if tcErr != nil {
    return Expr{}, tcErr
}
exprs = append(exprs, tcs...)
```

### 5. Remove dead `TypeConstraints` from solver_compat.go (lines 506-525)

The exported `TypeConstraints` in solver_compat.go is never called and has a wrong implementation (handles EnumeratedSort instead of nat/range like Python). Remove it.

### 6. Remove old `typeConstraintsForSymbol`

After splitting into `buildConstraintTerm`, `natConstraintForSymbol`, and `rangeConstraintsForSymbol`, the original `typeConstraintsForSymbol` is unused. Remove it.

### Files to modify
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/z3bridge/solver.go` — split helpers, create `typeConstraints`, update `formulaToZ3` and `ClausesToZ3`
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/z3bridge/solver_compat.go` — remove dead `TypeConstraints`

## Verification

Run `cd ~/ivy/goivy && make golden` and confirm the divergence at 236241 is resolved.
