# Restructure quantifier handling to match Python's `quant_constraints`/`forall`/`exists`

**Created:** 2026-04-09 ~05:00 UTC

## Context

Golden test diverges at line 236408:
- **Go**: `type_constraints() ENTER nsyms=2`
- **Python**: `quant_constraints() ENTER nvars=3`

Root cause: A previous claude unfaithfully split Python's single `quant_constraints(vs, z3_vs)` function into a per-variable method `Translator.QuantConstraints(v, z3Var)` with no XTRACE. Go's `forall()` and `translateQuantifier()` call this per-variable method in a loop. Python's `forall()` calls `quant_constraints()` once with all variables, and `quant_constraints` emits the trace.

Additionally, `translateQuantifier()` on Translator bundles variable creation, body translation, constraint collection, and ForAll/Exists wrapping into one monolithic function. Python keeps these separate: `formula_to_z3_int` translates body and variables, then calls `forall()`/`exists()` which calls `quant_constraints()`.

There is also no `Solver.exists()` — the exists path goes through `translateQuantifier` only.

## Python structure (source of truth)

```python
# ivy_solver.py:557-584
def quant_constraints(vs, z3_vs):          # all vars at once, emits trace
def forall(vs, z3_vs, z3_body):            # calls quant_constraints, wraps Implies
def exists(vs, z3_vs, z3_body):            # calls quant_constraints, wraps And

# Called from:
# - formula_to_z3_int (line 689-694): encounters ForAll/Exists AST → calls forall()/exists()
# - formula_to_z3_closed (line 710): wraps free vars → calls forall()
```

## Plan

### Step 1: Create `Solver.quantConstraints()` — faithful port of Python's `quant_constraints`

New method on `*Solver` in `z3bridge/solver.go`. Takes all variables at once, emits trace, loops internally. Logic moved from `Translator.QuantConstraints()`.

```go
// quantConstraints generates Z3 constraints for quantifier-bound variables.
// Matches Python quant_constraints (ivy_solver.py:557-568).
func (s *Solver) quantConstraints(vars []*lg.Variable, z3Vars []Expr) []Expr {
    xtracer.Trace("ivy_solver.py:545 quant_constraints() ENTER nvars=%d", len(vars))
    if s.sig == nil {
        return nil
    }
    var cnstrs []Expr
    for i, v := range vars {
        sortName := il.SortName(v.VSort)
        itp, ok := s.sig.Interp[sortName]
        if !ok {
            continue
        }
        ctx := s.tr.Ctx
        switch itpVal := itp.(type) {
        case string:
            if itpVal == "nat" {
                cnstrs = append(cnstrs, ctx.Le(ctx.IntVal(0), z3Vars[i]))
            }
        case *lg.RangeSort:
            if s.HandleRangeSorts {
                lb, ub, err := s.RangeSortBoundsToZ3(itpVal)
                if err == nil {
                    cnstrs = append(cnstrs, ctx.Le(lb, z3Vars[i]))
                    cnstrs = append(cnstrs, ctx.Le(z3Vars[i], ub))
                }
            }
        }
    }
    return cnstrs
}
```

### Step 2: Refactor `Solver.forall()` to call `s.quantConstraints()`

Replace the per-variable loop with a single call:

```go
func (s *Solver) forall(vars []*lg.Variable, z3Vars []Expr, z3Body Expr) Expr {
    xtracer.Trace("ivy_solver.py:560 forall() ENTER nvars=%d", len(vars))
    cnstrs := s.quantConstraints(vars, z3Vars)
    if len(cnstrs) > 0 {
        z3Body = s.tr.Ctx.Implies(s.tr.Ctx.And(cnstrs...), z3Body)
    }
    return s.tr.Ctx.ForAll(z3Vars, z3Body)
}
```

### Step 3: Create `Solver.exists()` — faithful port of Python's `exists`

New method in `z3bridge/solver.go`:

```go
// exists wraps a Z3 body in Exists with quant constraints (nat/range bounds).
// Matches Python's exists (ivy_solver.py:579-584).
func (s *Solver) exists(vars []*lg.Variable, z3Vars []Expr, z3Body Expr) Expr {
    xtracer.Trace("ivy_solver.py:567 exists() ENTER nvars=%d", len(vars))
    cnstrs := s.quantConstraints(vars, z3Vars)
    if len(cnstrs) > 0 {
        z3Body = s.tr.Ctx.And(append(cnstrs, z3Body)...)
    }
    return s.tr.Ctx.Exists(z3Vars, z3Body)
}
```

### Step 4: Refactor `Translator.translateQuantifier()` to delegate to `forall`/`exists`

Remove:
- The `forall() ENTER` / `exists() ENTER` trace at the top (forall/exists emit their own)
- The inline QuantConstraints loop (lines 1013-1028)
- The inline ForAll/Exists wrapping (lines 1030-1033)

Replace with call to `t.s.forall()` or `t.s.exists()` after body translation:

```go
func (t *Translator) translateQuantifier(isForall bool, variables []*lg.Variable, body lg.Expr) (Expr, error) {
    if len(variables) == 0 {
        return t.Formula_to_z3_int(body, "translateQuantifier() no variables")
    }

    // Create Z3 constants for the bound variables
    bound := make([]Expr, len(variables))
    for i, v := range variables {
        zs, err := t.TranslateSort(v.VSort)
        if err != nil {
            return Expr{}, err
        }
        z3Var, err := t.translateVariable(v)
        if err != nil {
            return Expr{}, err
        }
        key := lg.NodeKey(v.Name + ":" + string(v.VSort.Sexp()))
        bound[i] = z3Var
        _ = zs
        t.consts[key] = bound[i]
    }

    zBody, err := t.Formula_to_z3_int(body, "translateQuantifier() len(variables) > 0")
    if err != nil {
        return Expr{}, err
    }

    // Validate body is Bool (Go safety check)
    if zBody.ExprSort().Kind() != SortBool {
        return Expr{}, fmt.Errorf("quantifier body must be Bool, got sort %s", zBody.ExprSort().String())
    }

    // Delegate to forall/exists which call quantConstraints
    if isForall {
        return t.s.forall(variables, bound, zBody), nil
    }
    return t.s.exists(variables, bound, zBody), nil
}
```

This fixes the trace ordering too: Python translates body first, then calls forall/exists (which emit their traces). Now Go does the same.

### Step 5: Delete `Translator.QuantConstraints()` method

Remove the method at `translate.go:97-129`. No callers remain after steps 2 and 4.

Also delete the `QuantConstraintsFn` type definition at `translate.go:23-26` (unused after field was commented out).

### Step 6: Delete dead `QuantConstraints` in `solver_compat.go`

The standalone `QuantConstraints(vs []*lg.Variable, z3Vs interface{}) lg.Expr` at solver_compat.go:489 has zero callers. Delete it.

## Files to modify

| File | Change |
|------|--------|
| `z3bridge/solver.go` | Add `quantConstraints()`, create `exists()`, refactor `forall()` |
| `z3bridge/translate.go` | Refactor `translateQuantifier()` to delegate; delete `QuantConstraints()` method and `QuantConstraintsFn` type |
| `z3bridge/solver_compat.go` | Delete dead `QuantConstraints` function |

## Tests affected

- `solver2_test.go` — 4 tests go through `FormulaToZ3` → full pipeline → will still pass
- `solver2_fuzz_test.go:FuzzQuantConstraintsNatRange` — goes through `FormulaToZ3` → still works
- `translate2_test.go` — callback-based tests already commented out (`/* */`)
- `translate2_fuzz_test.go:FuzzQuantConstraintsForAll` — already commented out (`/* */`)

## Verification

1. `go build ./...` — must compile
2. `go vet ./...` — no errors
3. `go test ./z3bridge/...` — existing tests pass
4. `cd ~/ivy/goivy && make golden` — divergence at 236408 resolved
