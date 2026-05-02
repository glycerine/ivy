# Plan 367: Implement Skolem Witness in propertyTactic

**Created:** 2026-05-02 14:22 UTC

## Context

The Go `propertyTactic` in `proof/tactics.go:518-523` has a stub that rejects proofs using `property ... named f(X)` syntax (optskolem). This syntax introduces a Skolem function witness for an existential quantifier in a property proof. The Python implementation is at `ivy_proof.py:332-366`. The grammar already parses `optskolem` correctly and stores it in `PropertyTactic.PName` — only the tactic logic is missing.

## File to modify

**`/Users/jaten/ivy/goivy/proof/tactics.go`** — replace stub at lines 518-523, add helper, add import.

## New file

**`/Users/jaten/ivy/goivy/proof/tactics_property_test.go`** — comprehensive unit tests.

## Implementation

### Step 1: Add `il` import to tactics.go

Add `il "github.com/glycerine/ivy/goivy/ivylogic"` to the import block (line ~12).

### Step 2: Add helper function `defargNameSort`

Extracts (name string, sort ast.Node, isTopSort bool) from a defarg parameter node. The grammar's `defnlhs` produces an `*ast.Atom` whose `Terms` are either `*ast.App` (from `lparam`) or `*ast.Variable` (from `var`).

```go
// defargNameSort extracts the name and sort from a defarg parameter node.
// defnlhs parameters are *ast.App (from lparam: SYMBOLx:atype) or
// *ast.Variable (from var: TOK_VARIABLE).
// Returns (name, sortName, isTopSort).
func defargNameSort(n ast.Node) (string, string, bool) {
    switch a := n.(type) {
    case *ast.App:
        name := a.Relname()
        if a.ASort == nil {
            return name, "", true
        }
        sn := fmt.Sprint(a.ASort)
        return name, sn, false
    case *ast.Variable:
        if a.VSort == "" || a.VSort == "S" {
            return a.Rep, a.VSort, true  // "S" = universe = topsort
        }
        return a.Rep, a.VSort, false
    case *ast.Atom:
        if a.ASort == nil {
            return a.Rep, "", true
        }
        sn := fmt.Sprint(a.ASort)
        return sn, sn, false
    }
    return fmt.Sprint(n), "", true
}
```

### Step 3: Replace stub with Skolem implementation

Replace lines 518-523 in `propertyTactic` with a faithful port of Python lines 332-366. The logic:

1. **Get formula, strip universals:** `GoalConcUnwrap(cut)` → `il.DropUniversals()`
2. **Validate existential:** must be `*lg.Exists` with exactly 1 bound variable
3. **Extract existential variable:** `evar = BinderVars(fmla)[0]`, `rng = evar.VSort`
4. **Build variable map:** `lu.VariablesAstList(fmla)` → `vmap[name]*lg.Variable`
5. **Process parameters** from `lhs.Terms`:
   - Check for duplicate names
   - If name in vmap: use logic variable, check sort compatibility (Python line 350)
   - If name NOT in vmap: require explicit sort, create `*lg.Variable`
6. **Check all formula variables are covered** as parameters
7. **Create Skolem function symbol:** `&lg.Const{Name: lhs.Rep, CSort: il.FuncConstSort(dom..., rng)}`
8. **Check freshness:** not in `pc.Stale` and not in `GoalDefns(goal)`
9. **Create witness term:** `&lg.Apply{Func: sym, Terms: targs}` (or bare `sym` if no args)
10. **Substitute:** `module.SubstituteAstByName(body, {evar.Name: term})`
11. **Update cut:** `CloneGoal(cfg, cut, nil, substituted_fmla)`
12. **Add ConstantDecl:** `GoalAddPrem(cfg, goal, NewConstantDecl(sym), goal.GetLineno())`

Key Python→Go mappings:
- `lhs` = `proof.PName.(*ast.Atom)` 
- `lhs.rep` = `lhs.Rep`
- `lhs.args` = `lhs.Terms`
- `il.drop_universals(cut.formula)` = `il.DropUniversals(GoalConcUnwrap(cut))`
- `lu.variables_ast(fmla)` = `lu.VariablesAstList(fmla)`
- `il.Symbol(name, sort)` = `&lg.Const{Name: name, CSort: sort}`
- `sym(*targs)` = `&lg.Apply{Func: sym, Terms: targs}`
- `lu.substitute_ast(body, {name: term})` = `module.SubstituteAstByName(body, map[string]lg.Expr{name: term})`
- `ia.ConstantDecl(sym)` = `pc.astCfg().NewConstantDecl(sym)`
- `clone_goal(cut, [], fmla)` = `CloneGoal(pc.astCfg(), cut, nil, fmla)`

### Step 4: Comprehensive unit tests

File: `proof/tactics_property_test.go`

**Test helpers needed:**
- `mkSort(name)` — creates `*lg.UninterpretedSort`
- `mkVar(name, sort)` — creates `*lg.Variable`
- `mkConst(name, sort)` — creates `*lg.Const`
- `mkApply(fn, args...)` — creates `*lg.Apply`
- `mkGoal(cfg, formula)` — wraps formula in `*ast.LabeledFormula`
- `mkPropertyTactic(cfg, prop, pname, proof)` — creates `*ast.PropertyTactic`
- `mkAtomWithTerms(cfg, name, terms)` — creates `*ast.Atom` for optskolem LHS
- `mkChecker(axioms)` — creates minimal `*ProofChecker`

**Test cases (13 total):**

| # | Test | Input | Expected |
|---|------|-------|----------|
| 1 | Basic Skolem witness | `∀X:S. ∃Y:S. P(X,Y)`, named `f(X)` | Skolem `f:S→S`, formula becomes `P(X,f(X))` |
| 2 | Nullary Skolem constant | `∃Y:S. P(Y)`, named `c` | Constant `c:S`, formula becomes `P(c)` |
| 3 | Non-existential error | `∀X:S. P(X)`, named `f(X)` | Error: "property is not existential" |
| 4 | Multi-variable existential error | `∃X,Y:S. P(X,Y)`, named `f` | Error: "property is not existential" |
| 5 | Repeated parameter error | `∀X:S. ∃Y:S. P(X,Y)`, named `f(X,X)` | Error: "repeat parameter: X" |
| 6 | Missing parameter error | `∀X:S. ∀Z:S. ∃Y:S. P(X,Z,Y)`, named `f(X)` | Error: "Z must be a parameter of f" |
| 7 | Cannot infer sort error | `∃Y:S. P(Y)`, named `f(W)` (W topsort) | Error: "cannot infer sort for W" |
| 8 | Extra param explicit sort | `∃Y:S. P(Y)`, named `f(w:T)` | Skolem `f:T→S` |
| 9 | Stale symbol error | `∀X:S. ∃Y:S. P(X,Y)`, named `f(X)`, f in Stale | Error: "f is not fresh" |
| 10 | GoalDefns symbol error | Same but f in goal premises as ConstantDecl | Error: "f is not fresh" |
| 11 | Sort annotation matches (Python line 350) | `∀X:S. ∃Y:S. P(X,Y)`, named `f(X:S)` | Error: "bad sort for X" |
| 12 | NoneAST passthrough | Normal property, PName=NoneAST | No error, existing behavior preserved |
| 13 | Sort mismatch annotation OK | `∀X:S. ∃Y:S. P(X,Y)`, named `f(X:T)` | No sort error (Python condition) |

## Existing infrastructure to reuse

| Function | File:Line | Purpose |
|----------|-----------|---------|
| `il.DropUniversals` | `ivylogic/classify.go:80` | Strip leading ForAll |
| `il.IsExists` | `ivylogic/ivylogic.go:161` | Check Exists type |
| `il.BinderVars` | `ivylogic/util.go:152` | Get bound variables |
| `il.BinderBody` | `ivylogic/util.go:176` | Get binder body |
| `il.FuncConstSort` | `ivylogic/ivylogic.go:75` | Create function sort |
| `il.IsTopSort` | `ivylogic/ivylogic.go:261` | Check TopSort |
| `lu.VariablesAstList` | `logicutil/logicutil.go:639` | Free variables in DFS order |
| `module.SubstituteAstByName` | `module/batch16.go:16` | Substitute by name |
| `GoalDefns` | `proof/goal.go:344` | Symbols defined in goal |
| `GoalConcUnwrap` | `proof/goal.go:72` | Extract conclusion as lg.Expr |
| `CloneGoal` | `proof/goal.go:191` | Clone goal with new prems/conc |
| `GoalAddPrem` | `proof/goal.go:557` | Add premise to goal |
| `isNoneAST` | `proof/tactics.go:241` | Check NoneAST |
| `pc.astCfg()` | `proof/checker.go:145` | Get AstConfig |

## Verification

1. `cd ~/ivy/goivy && make test` — run full test suite
2. Unit tests in `tactics_property_test.go` cover all 13 cases above
3. Verify xtracer output matches Python patterns for property tactic Skolem path
