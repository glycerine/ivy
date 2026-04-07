# PLAN217: Refactor ClausesToZ3 to mirror Python's internal call structure

**Created:** 2026-04-07 ~19:30 UTC
**Revised:** 2026-04-07 ~20:15 UTC — full structural alignment with Python

## Context

PLAN216 succeeded — lines 236041-236045 now match. The golden test diverges at line 236046 with a `z3bridge.Translate HASH` mismatch:

```
236046  go : z3bridge.Translate HASH ... canon=(ForAll vars:[(Variable name:T0 sort:lclock)...] body:(Iff ...
        py : z3bridge.Translate HASH ... canon=(Def lhs:(Apply func:(Symbol name:ref.prevents...)...)
```

Go also panics with "Sort mismatch at argument #1 for function (declare-fun < (Int Int) Bool) supplied sort is lclock".

## Root Cause

Go's `ClausesToZ3` internal call structure diverges from Python's:

**Python's call chain (what we want):**
```
clauses_to_z3(clauses):
    fmlas → conj_to_z3(f)         → formula_to_z3_closed(f) → formula_to_z3_int(f)  [no HASH]
    defs  → formula_to_z3(dfn)    → formula_to_z3_closed(f) → formula_to_z3_int(f)  [HASH]
    + type_constraints(used_symbols_clauses(clauses))                                 [no HASH]
```

**Go's current call chain (wrong):**
```
ClausesToZ3(clauses):
    fmlas → translateClosed(f)     → CloseFormula → Translate  [HASH via depth=0]
    defs  → defToConstraint(d)+translateClosed → CloseFormula → Translate  [HASH, wrong canon]
    + typeConstraintsForSymbol → translateClosed → Translate    [HASH, Python has none]
```

Key differences:
1. **Definitions**: Go converts Def→Iff via `defToConstraint`, Python passes raw Def
2. **Closing**: Go wraps in ForAll at logic level (via `CloseFormula`), Python wraps at Z3 level (raw `z3.ForAll`)
3. **HASH traces**: Go emits HASH for fmlas/defs/type_constraints; Python emits HASH only for defs (via `formula_to_z3`)
4. **Quant constraints**: Go's `translateQuantifier` adds sort constraints (causing Sort mismatch panic); Python uses raw `z3.ForAll` for Definitions (no sort constraints)

## Approach

Create Go equivalents of Python's function hierarchy and rewire `ClausesToZ3`:

| Python function         | Go equivalent (new)      | Role |
|------------------------|--------------------------|------|
| `formula_to_z3`        | `formulaToZ3`            | HASH trace + close + type_constraints |
| `formula_to_z3_closed` | `formulaToZ3Closed`      | Translate + ForAll wrapping |
| `formula_to_z3_int`    | `s.tr.Translate` (existing) | Recursive core translator |
| `conj_to_z3`           | `conjToZ3`               | And-recursive, delegates to formulaToZ3Closed |
| `forall(vs,z3vs,body)` | `forall`                 | quant_constraints + z3.ForAll |
| `term_to_z3`           | `s.tr.TranslateVar` (new) | Variable → Z3 const |

## File Changes

### A. `z3bridge/translate.go` — add two methods

**A1. TranslateNoHash** (after line 196):

Translate without the depth-0 HASH trace. Used by the solver's `formulaToZ3Closed` and `conjToZ3` paths to avoid emitting HASH traces that Python doesn't emit.

```go
// TranslateNoHash translates without emitting the top-level HASH trace.
// Matches Python's formula_to_z3_int/formula_to_z3_closed which do not
// emit HASH — only formula_to_z3 does.
func (t *Translator) TranslateNoHash(n logic.Expr) (Expr, error) {
    t.translateDepth++
    defer func() { t.translateDepth-- }()
    return t.Translate(n)
}
```

**A2. TranslateVar** (after line 426):

```go
// TranslateVar translates a Variable to a Z3 const without emitting a
// HASH trace. Matches Python's term_to_z3(v).
func (t *Translator) TranslateVar(v *logic.Variable) (Expr, error) {
    return t.translateVariable(v)
}
```

### B. `solver/solver.go` — add four internal helpers + update imports

**B0. Add imports:**
```go
"sort"
lu "github.com/glycerine/ivy/goivy/logicutil"
iu "github.com/glycerine/ivy/goivy/ivyutils"
```

**B1. `formulaToZ3`** — matches Python `formula_to_z3` (ivy_solver.py:659-676)

Add after `translateClosed` (after line 423):

```go
// formulaToZ3 translates a formula to Z3 with HASH trace and type constraints.
// Matches Python's formula_to_z3 (ivy_solver.py:659-676).
//
// Call chain: formulaToZ3 → formulaToZ3Closed → Translate (no HASH)
// Only this function emits the HASH trace, matching Python.
func (s *Solver) formulaToZ3(fmla lg.Expr) (x z3bridge.Expr, err error) {
    defer func() {
        r := recover()
        if r != nil {
            vv("warning: recover from panic on formulaToZ3: '%v'", r)
            err = fmt.Errorf("%v", r)
        }
    }()

    // Emit HASH trace matching Python formula_to_z3 line 660-663
    if xtracer.Enabled {
        canon := iu.Canonical(fmla.Sexp())
        leaf, root := s.tr.TranslateMerkle.AddLeaf(canon)
        xtracer.Trace("z3bridge.Translate HASH leaf=%s root=%s canon=%s", leaf, root, string(canon))
    }

    z3Fmla, err := s.formulaToZ3Closed(fmla)
    if err != nil {
        xtracer.Trace("formula_to_z3: Z3 error on formula_to_z3_closed: %v type=%T", err, fmla)
        return z3bridge.Expr{}, err
    }

    // Per-formula type constraints matching Python formula_to_z3 line 670-672
    usedSyms := lu.UsedConstantsList(fmla)
    var tcs []z3bridge.Expr
    for _, sym := range usedSyms {
        constraints := s.typeConstraintsForSymbol(sym)
        for _, tc := range constraints {
            closed := il.CloseFormula(tc)
            ztc, err := s.tr.TranslateNoHash(closed)
            if err != nil {
                continue
            }
            tcs = append(tcs, ztc)
        }
    }
    if len(tcs) > 0 {
        all := make([]z3bridge.Expr, 0, len(tcs)+1)
        all = append(all, z3Fmla)
        all = append(all, tcs...)
        return s.tr.Ctx.And(all...), nil
    }
    return z3Fmla, nil
}
```

**B2. `formulaToZ3Closed`** — matches Python `formula_to_z3_closed` (ivy_solver.py:646-655)

```go
// formulaToZ3Closed translates and closes (universally quantifies free vars).
// Matches Python's formula_to_z3_closed (ivy_solver.py:646-655).
//
// For Definition: wraps in raw z3.ForAll (no quant constraints).
// For others: wraps via forall() helper (with quant constraints).
func (s *Solver) formulaToZ3Closed(fmla lg.Expr) (z3bridge.Expr, error) {
    z3Formula, err := s.tr.TranslateNoHash(fmla)
    if err != nil {
        return z3bridge.Expr{}, err
    }

    freeVars := lu.FreeVariablesList(fmla)
    if len(freeVars) == 0 {
        return z3Formula, nil
    }

    // Sort variables matching Python: sorted(used_variables_ast(fmla))
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

    // Definition: raw z3.ForAll (no quant constraints)
    // Other: forall() with quant constraints
    // Matches Python formula_to_z3_closed lines 653-654
    if _, isDef := fmla.(*lg.Definition); isDef {
        return s.tr.Ctx.ForAll(z3Vars, z3Formula), nil
    }
    return s.forall(freeVars, z3Vars, z3Formula), nil
}
```

**B3. `conjToZ3`** — matches Python `conj_to_z3` (ivy_solver.py:546-549)

```go
// conjToZ3 translates a conjunction to Z3 without HASH trace.
// Matches Python's conj_to_z3 (ivy_solver.py:546-549).
// For And: recursively translates each conjunct.
// Otherwise: delegates to formulaToZ3Closed.
func (s *Solver) conjToZ3(fmla lg.Expr) (z3bridge.Expr, error) {
    if and, ok := fmla.(*lg.And); ok {
        z3Args := make([]z3bridge.Expr, len(and.Terms))
        for i, t := range and.Terms {
            var err error
            z3Args[i], err = s.conjToZ3(t)
            if err != nil {
                return z3bridge.Expr{}, err
            }
        }
        return s.tr.Ctx.And(z3Args...), nil
    }
    return s.formulaToZ3Closed(fmla)
}
```

**B4. `forall` helper** — matches Python `forall` (ivy_solver.py:524-528)

```go
// forall wraps a Z3 body in ForAll with quant constraints (nat/range bounds).
// Matches Python's forall (ivy_solver.py:524-528).
func (s *Solver) forall(vars []*lg.Variable, z3Vars []z3bridge.Expr, z3Body z3bridge.Expr) z3bridge.Expr {
    if s.tr.QuantConstraints != nil {
        var cnstrs []z3bridge.Expr
        for i, v := range vars {
            cs := s.tr.QuantConstraints(v, z3Vars[i])
            cnstrs = append(cnstrs, cs...)
        }
        if len(cnstrs) > 0 {
            z3Body = s.tr.Ctx.Implies(s.tr.Ctx.And(cnstrs...), z3Body)
        }
    }
    return s.tr.Ctx.ForAll(z3Vars, z3Body)
}
```

### C. `solver/solver.go` — update ClausesToZ3 (lines 273-332)

Replace the body of ClausesToZ3 to match Python's clauses_to_z3 call structure:

```go
func (s *Solver) ClausesToZ3(clauses *module.Clauses) (z3bridge.Expr, error) {
    if clauses == nil {
        xtracer.Trace("solver.ClausesToZ3 ENTER nil")
        return s.tr.Ctx.BoolVal(true), nil
    }
    xtracer.Trace("solver.ClausesToZ3 ENTER fmlas=%d defs=%d", len(clauses.Fmlas), len(clauses.Defs))

    var exprs []z3bridge.Expr

    // Translate formulas via conjToZ3 matching Python: [conj_to_z3(cl) for cl in clauses.fmlas]
    for i, f := range clauses.Fmlas {
        xtracer.Trace("solver.ClausesToZ3 fmla[%d] sort=%v", i, f.NodeSort())
        zf, err := s.conjToZ3(f)
        if err != nil {
            return z3bridge.Expr{}, fmt.Errorf("translating formula: %w", err)
        }
        exprs = append(exprs, zf)
    }

    // Translate definitions via formulaToZ3 matching Python: formula_to_z3(dfn)
    for di, d := range clauses.Defs {
        zd, err := s.formulaToZ3(d)
        if err != nil {
            defName := "?"
            if sym := d.Defines(); sym != nil {
                if c, ok := sym.(*lg.Const); ok {
                    defName = c.Name
                }
            }
            xtracer.Trace("clauses_to_z3: Z3 error on def[%d]: %v defines=%s", di, err, defName)
            return z3bridge.Expr{}, fmt.Errorf("translating definition: %w", err)
        }
        exprs = append(exprs, zd)
    }

    // Type constraints matching Python: type_constraints(used_symbols_clauses(clauses))
    usedSyms := clauses.Symbols()
    for _, sym := range usedSyms {
        constraints := s.typeConstraintsForSymbol(sym)
        for _, tc := range constraints {
            closed := il.CloseFormula(tc)
            ztc, err := s.tr.TranslateNoHash(closed)
            if err != nil {
                continue
            }
            exprs = append(exprs, ztc)
        }
    }

    xtracer.Trace("solver.ClausesToZ3 EXIT exprs=%d", len(exprs))
    if len(exprs) == 0 {
        return s.tr.Ctx.BoolVal(true), nil
    }
    if len(exprs) == 1 {
        return exprs[0], nil
    }
    return s.tr.Ctx.And(exprs...), nil
}
```

Note: The `\n` was removed from the ENTER trace to match the previous fix (PLAN216 removed `\n` for stronger checks).

### D. Keep existing code for other callers

- **`translateClosed`** (line 411-423): Unchanged. Still used by `ImpliesBatch`, `UnsatCore`, `IsSat`, `model.go`, `clauses.go`, `compat.go` etc.
- **`FormulaToZ3`** (public, line 261-266): Unchanged for now. Used by `vmt`, `alpha`, tests. Future task: align with Python's `formula_to_z3` properly.
- **`defToConstraint`** (line 456-467): Unchanged. Still used by UnsatCore (line 612).

## Files Modified

- `z3bridge/translate.go` — add `TranslateNoHash` and `TranslateVar`
- `solver/solver.go` — add `formulaToZ3`, `formulaToZ3Closed`, `conjToZ3`, `forall`; rewrite `ClausesToZ3` body; add imports

## Verification

```bash
cd ~/ivy/goivy && go build ./... && go test ./solver/... && go test ./actions/... && make golden
```

Expected: Line 236046 now shows matching `(Def ...)` canons. The Sort mismatch panic is resolved (Definition uses raw z3.ForAll, no quant constraints). The golden test advances past line 236046 + all 12 definition HASH traces.
