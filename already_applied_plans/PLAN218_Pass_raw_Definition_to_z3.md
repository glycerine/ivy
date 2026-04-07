# PLAN217: Pass raw Definition to Z3 translator (match Python's clauses_to_z3)

**Created:** 2026-04-07 ~19:30 UTC

## Context

PLAN216 succeeded — lines 236041-236045 now match (checkFcsNormalPath, SatisfyWithCond, ClausesToZ3 ENTER traces). The golden test diverges at line 236046 with a `z3bridge.Translate HASH` mismatch:

```
236046  go : XTRACE: z3bridge.Translate HASH ... canon=(ForAll vars:[(Variable name:T0 sort:lclock) (Variable name:T1 sort:lclock)] body:(Iff ...
        py : XTRACE: z3bridge.Translate HASH ... canon=(Def lhs:(Apply func:(Symbol name:ref.prevents ...) ...) ...
```

Go sends `ForAll(Iff(...))` canon, Python sends `Def(lhs, rhs)` canon. Different blake3 hashes → divergence.

Additionally, Go panics with: "Sort mismatch at argument #1 for function (declare-fun < (Int Int) Bool) supplied sort is lclock" — a consequence of the wrong translation path.

## Root Cause

In Go's `ClausesToZ3` (`solver/solver.go:293-308`), each definition goes through:
1. `defToConstraint(d)` → converts `Definition` to `Iff(lhs, rhs)` or `Eq(lhs, rhs)`
2. `translateClosed(constraint)` → calls `CloseFormula(constraint)` → wraps in `ForAll(freeVars, constraint)` at logic level
3. `Translate(ForAll(...))` → emits HASH trace showing `(ForAll ... (Iff ...))` canon

In Python's `clauses_to_z3` (`ivy_solver.py:586-591`), each definition goes through:
1. `formula_to_z3(dfn)` → passes raw Definition directly
2. Emits HASH trace showing `(Def lhs:... rhs:...)` canon
3. `formula_to_z3_int(dfn)` → handles Definition natively as `my_eq(z3_lhs, z3_rhs)`
4. `formula_to_z3_closed(dfn)` → wraps in `z3.ForAll(z3_variables, z3_formula)` at **Z3 level** (not logic level)

The mismatch is twofold:
- **Canon mismatch**: Go shows ForAll wrapping Iff, Python shows raw Def
- **Quantifier handling**: Go's `translateQuantifier` adds sort constraints (nat non-negativity, range bounds) to the quantified body via `QuantConstraints`. Python uses raw `z3.ForAll` for Definition (no sort constraints). This explains the "Sort mismatch" — Go's `translateQuantifier` adds a constraint involving `<` for the lclock sort's nat/range interpretation, but `<` is declared for Int not lclock.

## Approach

1. In `ClausesToZ3`, replace `defToConstraint(d)` + `translateClosed(constraint)` with a new `translateDefinition(d)` method
2. `translateDefinition` calls `s.tr.Translate(d)` directly (emits HASH with Def canon), then wraps in z3 ForAll at Z3 level using `s.tr.Ctx.ForAll` (no sort constraints, matching Python)
3. Add `TranslateVar` exported method on z3bridge.Translator to translate variables without HASH traces (for building ForAll bound list after main translation)

## File Changes

### A. `z3bridge/translate.go` — add TranslateVar

Add exported wrapper for `translateVariable`, used to translate free variables into Z3 consts without emitting HASH trace:

```go
// TranslateVar translates a Variable to a Z3 const without emitting a
// top-level HASH trace. Matches Python's term_to_z3(v) used in
// formula_to_z3_closed when building ForAll bound variable lists.
func (t *Translator) TranslateVar(v *logic.Variable) (Expr, error) {
    return t.translateVariable(v)
}
```

Insert after `translateVariable` (after line 426).

### B. `solver/solver.go` — new translateDefinition method

Add after `translateClosed` (after line 423):

```go
// translateDefinition translates a Definition to Z3, matching Python's
// formula_to_z3(dfn) → formula_to_z3_closed(dfn) path in clauses_to_z3.
//
// Unlike translateClosed (which converts to ForAll at the logic level via
// CloseFormula, adding sort constraints via translateQuantifier), this:
// 1. Calls Translate(d) directly — emits HASH trace with (Def ...) canon
// 2. Wraps in z3.ForAll at Z3 level — no sort constraints, matching Python
func (s *Solver) translateDefinition(d *il.Definition) (x z3bridge.Expr, err error) {
    defer func() {
        r := recover()
        if r != nil {
            vv("warning: recover from panic on translateDefinition: '%v'", r)
            err = fmt.Errorf("%v", r)
        }
    }()

    // Translate the Definition directly at depth 0.
    // Emits HASH trace with (Def ...) canon matching Python.
    z3Def, err := s.tr.Translate(d)
    if err != nil {
        return z3bridge.Expr{}, err
    }

    // Close at Z3 level: wrap in ForAll with sorted free variables.
    // Matches Python formula_to_z3_closed:
    //   z3_variables = [term_to_z3(v) for v in sorted(used_variables_ast(fmla))]
    //   if isinstance(fmla, Definition): return z3.ForAll(z3_variables, z3_formula)
    freeVars := lu.FreeVariablesList(d)
    if len(freeVars) == 0 {
        return z3Def, nil
    }
    sort.Slice(freeVars, func(i, j int) bool {
        return freeVars[i].Name < freeVars[j].Name
    })

    z3Vars := make([]z3bridge.Expr, len(freeVars))
    for i, v := range freeVars {
        z3Vars[i], err = s.tr.TranslateVar(v)
        if err != nil {
            return z3bridge.Expr{}, err
        }
    }

    return s.tr.Ctx.ForAll(z3Vars, z3Def), nil
}
```

Add two imports to `solver/solver.go`:
```go
"sort"
lu "github.com/glycerine/ivy/goivy/logicutil"
```

### C. `solver/solver.go` — ClausesToZ3 def loop (lines 293-308)

Replace:
```go
    // Translate definitions as constraints
    for di, d := range clauses.Defs {
        constraint := defToConstraint(d)
        zd, err := s.translateClosed(constraint)
```

With:
```go
    // Translate definitions directly, matching Python's formula_to_z3(dfn)
    for di, d := range clauses.Defs {
        zd, err := s.translateDefinition(d)
```

Rest of the error handling block stays the same.

### D. `solver/solver.go` — keep defToConstraint for UnsatCore

The local `defToConstraint` function (lines 456-467) is still used at line 612 in the unsat_core path (where Python also calls `d.to_constraint()` explicitly). Keep it.

## Files Modified

- `z3bridge/translate.go` — add `TranslateVar` method
- `solver/solver.go` — add `translateDefinition` method, change ClausesToZ3 def loop

## Verification

```bash
cd ~/ivy/goivy && go build ./... && go test ./solver/... && make golden
```

Expected: Line 236046 now shows matching `(Def ...)` canons with identical blake3 hashes. The Z3 Sort mismatch panic should also be resolved since we no longer route through `translateQuantifier` (which adds sort constraints involving `<`). The golden test should advance past line 236046 + all 12 definition HASH traces.
