# Plan: Port SECTION7.md Batch F — UnitRes Integration in ClausesCase

## Context

Go's `ClausesCase` (solver/herbrand.go:735-776) skips the UnitRes propagation step
that Python's `clauses_case` (ivy_solver.py:1031-1058) performs. Python iteratively
applies unit resolution propagation between model simplification rounds, which can
derive new unit clauses, subsume redundant clauses, and simplify the clause set more
aggressively. The Go `unitres` package is fully ported but has **no connection** to
the solver — there is no conversion layer between `ivylogic.Literal` (lg.Expr atoms)
and `unitres.Literal` (resolution.Atom).

**Item 7.3#3**: `clauses_case` unit resolution — Python performs `UnitRes` propagation
between rounds; Go skips entirely.

---

## The Type Gap

Two separate Literal implementations in Go, with NO conversion between them:

| Type | Atom | Used by |
|------|------|---------|
| `ivylogic.Literal{Polarity int, Atom lg.Expr}` | `*lg.Apply`, `*lg.Eq`, `*lg.Symbol` | clauseops (formula↔clause conversion) |
| `unitres.Literal{Polarity int, Atom *resolution.Atom}` | `resolution.Atom{RelName string, Args []logic.Expr}` | unitres package (unit resolution) |

The `resolution.Atom.Args` field is `[]logic.Expr` (same as `[]lg.Expr`), so term-level
data is shared. Only the atom wrapper differs.

---

## Python Algorithm (clauses_case, ivy_solver.py:1031-1058)

```
1. Check SAT, get model m
2. CNF = Tseitin(clauses)
3. Simplify each clause with model → simplified literal-lists
4. Remove duplicates
5. Loop:
   a. numOld = len(CNF)
   b. r = UnitRes(CNF)
   c. r.propagate()                        ← THIS IS MISSING IN GO
   d. newClauses = [[l] for l in r.unit_queue] + r.clauses
   e. Simplify each with model
   f. Remove duplicates
   g. if len(newCNF) <= numOld: return
```

---

## Implementation Steps

### Step 1: Write failing tests (TDD — RED phase)

**File:** `solver/herbrand_unitres_test.go` (new)

Tests:
1. **Conversion round-trip**: `il.Literal` → `unitres.Literal` → `il.Literal` preserves semantics
2. **UnitRes integration**: ClausesCase on a clause set where UnitRes propagation
   produces fewer clauses than model simplification alone
3. **Edge cases**: empty clauses (UNSAT), all-unit clauses, single clause

### Step 2: Build the conversion layer

**File:** `solver/unitres_bridge.go` (new)

This file stays in `solver/` because it's only needed by `ClausesCase` and doesn't
correspond to a Python file (Python has one Literal type so needs no bridge).

#### Forward: `ivyLitToUnitResLit`

```go
func ivyLitToUnitResLit(lit *il.Literal, symMap map[string]*lg.Symbol) *unitres.Literal
```

Atom conversion:
- `*lg.Apply{Func: *lg.Symbol{Name: n, ...}, Terms: args}` → `resolution.Atom{RelName: n, Args: args}`, record symMap[n] = symbol
- `*lg.Eq{T1: a, T2: b}` → `resolution.Atom{RelName: "=", Args: [a, b]}`
- `*lg.Symbol{Name: n}` (nullary) → `resolution.Atom{RelName: n, Args: nil}`, record symMap[n] = symbol

Uses existing `resolution.AtomFromApply` (resolution/resolution.go:115) for Apply case.

#### Reverse: `unitResLitToIvyLit`

```go
func unitResLitToIvyLit(lit *unitres.Literal, symMap map[string]*lg.Symbol) *il.Literal
```

Reconstruct:
- RelName "=" + 2 args → `*lg.Eq{T1, T2}`
- No args → `symMap[RelName]` (the original `*lg.Symbol`)
- With args → `*lg.Apply{Func: symMap[RelName], Terms: args}`

#### Batch converters

```go
// Convert [][]*il.Literal → [][]*unitres.Literal, building symMap
func ivyLitsToUnitResClauses(cnf [][]*il.Literal) ([][]*unitres.Literal, map[string]*lg.Symbol)

// Extract UnitRes results: [[l] for l in r.UnitQueue] + r.Clauses → [][]*il.Literal
func extractUnitResResults(r *unitres.UnitRes, symMap map[string]*lg.Symbol) [][]*il.Literal
```

### Step 3: Rewrite ClausesCase

**File:** `solver/herbrand.go` — modify `ClausesCase` (line 735)

New algorithm matching Python:

```go
func (s *Solver) ClausesCase(clauses *clauseops.Clauses) (*clauseops.Clauses, error) {
    // 1. Check SAT, get model (unchanged)
    z3solver := s.tr.Ctx.NewSolver()
    zc, err := s.ClausesToZ3(clauses)
    if err != nil { return nil, err }
    z3solver.Assert(zc)
    if z3solver.Check() == z3bridge.Unsat { return nil, nil }
    model := z3solver.Model()
    if model == nil { return clauses, nil }

    // 2. Initial model simplification
    //    Python: clauses = Clauses([clause_model_simp(m,c) for c in clauses1.clauses])
    //    Get CNF including defs via ToOpenFormula
    cnf := clauseops.FormulaToClausesAux(clauses.ToOpenFormula())
    var initFmlas []lg.Expr
    for _, c := range cnf {
        f := clauseops.ClauseToFormula(c)
        simplified := s.clauseModelSimp(model, f)
        initFmlas = append(initFmlas, simplified)
    }
    initFmlas = removeDuplicateFormulas(initFmlas)
    currentClauses := clauseops.NewClauses(initFmlas, nil, clauses.Annot)

    // 3. Iterative UnitRes + model simplification loop
    for {
        // Get CNF for this round
        cnf = clauseops.FormulaToClausesAux(currentClauses.ToOpenFormula())
        numOldClauses := len(cnf)

        // Convert to unitres format
        urClauses, symMap := ivyLitsToUnitResClauses(cnf)

        // Run UnitRes propagation (Python: r = ur.UnitRes(clauses.clauses); r.propagate())
        r := unitres.NewUnitRes(urClauses)
        r.Propagate(nil)

        // Extract: [[l] for l in r.unit_queue] + r.clauses
        resultLitClauses := extractUnitResResults(r, symMap)

        // Convert to formulas
        var resultFmlas []lg.Expr
        for _, c := range resultLitClauses {
            resultFmlas = append(resultFmlas, clauseops.ClauseToFormula(c))
        }
        newClauses := clauseops.NewClauses(resultFmlas, nil, currentClauses.Annot)

        // Model simplify each formula
        var simpFmlas []lg.Expr
        for _, f := range newClauses.Fmlas {
            simplified := s.clauseModelSimp(model, f)
            simpFmlas = append(simpFmlas, simplified)
        }
        simpFmlas = removeDuplicateFormulas(simpFmlas)
        currentClauses = clauseops.NewClauses(simpFmlas, nil, currentClauses.Annot)

        // Convergence: Python checks len(clauses.clauses) <= num_old_clauses
        newCnf := clauseops.FormulaToClausesAux(currentClauses.ToOpenFormula())
        if len(newCnf) <= numOldClauses {
            return currentClauses, nil
        }
    }
}
```

### Step 4: Make tests pass (TDD — GREEN phase)

Run `make test` from goivy root (respects DYLD_LIBRARY_PATH for Z3).

---

## Existing Functions to Reuse

| Function | Location | Purpose |
|----------|----------|---------|
| `FormulaToClausesAux` | clauseops/litclause.go:101 | Formula → CNF literal lists |
| `ClauseToFormula` | clauseops/litclause.go:161 | Literal list → Or formula |
| `LitToFormula` | clauseops/litclause.go:142 | Single literal → formula |
| `resolution.AtomFromApply` | resolution/resolution.go:115 | Apply node → resolution.Atom |
| `resolution.NewAtom` | resolution/resolution.go:109 | Create resolution.Atom |
| `unitres.NewUnitRes` | unitres/unitres.go:766 | Create and init UnitRes |
| `unitres.UnitRes.Propagate` | unitres/unitres.go:1200 | Run to fixed point |
| `clauseops.NewClauses` | clauseops/clauses.go:28 | Create Clauses wrapper |
| `Clauses.ToOpenFormula` | clauseops/clauses.go:93 | Conjunction of fmlas + defs |
| `il.NewLiteral` | ivylogic/formula.go:176 | Create ivylogic.Literal |
| `unitres.NewLiteral` | unitres/unitres.go:34 | Create unitres.Literal |
| `clauseModelSimp` | solver/herbrand.go:783 | Model-based clause simplification |
| `removeDuplicateFormulas` | solver/herbrand.go:825 | String-based dedup |

---

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `solver/herbrand_unitres_test.go` | CREATE | Tests (RED first) |
| `solver/unitres_bridge.go` | CREATE | Conversion: il.Literal ↔ unitres.Literal |
| `solver/herbrand.go` | MODIFY | Rewrite ClausesCase loop (lines 735-776) |

---

## Potential Bugs Found During Research

1. **Initial model simplification skips CNF conversion**: Current Go code iterates over
   `Clauses.Fmlas` directly, while Python iterates over `.clauses` (CNF via Tseitin encoding).
   For clauses with definitions, the Go version would miss the definition constraints in the
   initial simplification. Fix: use `FormulaToClausesAux(ToOpenFormula())` for the initial
   simplification step too (included in Step 3 above).

2. **`clausesCaseLegacy` dead code** (herbrand.go:838-884): Can be removed after integration
   is verified.

---

## Verification

1. **Unit tests**: `cd /Users/jaten/go/src/github.com/glycerine/goivy && make test`
   - Conversion round-trip tests
   - ClausesCase with UnitRes propagation
   - Edge cases (empty, all-units, UNSAT input)

2. **Integration**: Run the full goivy test suite — ClausesCase is called from transrel
   (`InterpolantCase` in transrel package), so existing transrel tests exercise this path.

3. **Manual comparison**: For a known Ivy program, compare the Go ClausesCase output
   against Python clauses_case output to verify behavioral equivalence.
