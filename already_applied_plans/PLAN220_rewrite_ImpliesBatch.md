# PLAN219: Rewrite ImpliesBatch and tactics to conform to Python z3_utils.py + tactics_api.py

**Created:** 2026-04-07 ~23:00 UTC

## Context

Go's `ImpliesBatch` (`solver/solver.go:648-670`) maps to Python's `z3_implies_batch` (`z3_utils.py:136-171`), which uses `to_z3()` — a simple recursive translator that does NOT close free variables, does NOT add type constraints, and does NOT emit HASH traces. This is a completely different translation system from `ivy_solver.formula_to_z3`.

Go's `ImpliedFacts` (`tactics/tactics.go:154-174`) maps to Python's `implied_facts` (`tactics_api.py:310-320`), which uses `z3_implies_batch`. The Go version has multiple bugs: it doesn't use batch implication, misses `normalize_quantifiers`, and incorrectly obtains background axioms.

Go's `RefutedGoal` (`tactics/tactics.go:93-115`) maps to Python's `refuted_goal` (`tactics_api.py:336-341`), which uses `z3_implies`. The Go version incorrectly obtains background axioms.

### Bugs Found

1. **ImpliesBatch uses wrong translation path**: Uses `translateClosed` which wraps free vars in ForAll at logic level. Python's `to_z3` does NOT close — free vars become shared Z3 constants. This is a semantic difference: with closing, each formula is independently universally quantified (`ForAll(X, premise) ∧ ForAll(X, ¬f)`). Without closing (Python), the same free variable X in premise and formula refers to the same Z3 constant. The Python semantics are correct for implication checking.

2. **ImpliesBatch missing timeout**: Python has `timeout=False` parameter that sets Z3 solver timeout to 2000ms.

3. **ImpliesBatch missing cache**: Python has `_implies_cache` dict with `(premise, f)` key. Per CLAUDE.md no-globals rule, cache goes on Solver struct.

4. **ImpliesBatch missing Unknown handling**: Python asserts on Unknown. Go silently treats as non-implied.

5. **ImpliedFacts doesn't use batch**: Creates individual `Implies()` calls (new solver per check), not the efficient single-solver push/pop pattern.

6. **ImpliedFacts missing normalize_quantifiers**: Python calls `normalize_quantifiers()` on the conjoined premise.

7. **ImpliedFacts/RefutedGoal wrong axiom handling**: Python uses `and_clauses(axioms, premise)` on Clauses objects. Go's `BackgroundTheory()` prematurely converts to formula, losing Clauses structure needed by `AndClausesTyped`.

## Approach

- Rewrite `ImpliesBatch` to use raw `s.tr.Translate()` (matching `to_z3`), add cache, timeout, Unknown handling
- Add `Z3Implies` single-formula method (matching Python's `z3_implies`)
- Add `BackgroundTheoryClauses()` to TacticsContext (returns `*module.Clauses`)
- Rewrite `ImpliedFacts` to use `ImpliesBatch` + `NormalizeQuantifiers`
- Rewrite `RefutedGoal` to use `Z3Implies` + `AndClausesTyped`
- Add comprehensive unit tests

## Detailed Changes

### A. `solver/solver.go` — 4 changes

**A1. Add `impliesCache` to Solver struct (line 50-56)**

```go
type Solver struct {
    mu               sync.Mutex
    tr               *z3bridge.Translator
    opts             *Options
    sig              *il.Sig
    HandleRangeSorts bool
    impliesCache     map[[2]lg.NodeKey]bool // z3_utils._implies_cache equivalent
}
```

Initialize in all constructors — `New()` (line 59-68), `NewWithSig()` (line 71-80), `NewWithOptions()` (line 83-95):
```go
impliesCache: make(map[[2]lg.NodeKey]bool),
```

Also reset in `Clear()` (line 103-107):
```go
func (s *Solver) Clear() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.tr.Clear()
    s.impliesCache = make(map[[2]lg.NodeKey]bool)
}
```

**A2. Add `Z3Implies` method (new, after line 601)**

Corresponds to Python `z3_utils.py:z3_implies` (lines 109-133). Single implication check with cache and raw translate.

```go
// Z3Implies checks whether f1 implies f2 using raw Z3 translation.
// Uses the implies cache. Returns error on Unknown.
// Corresponds to Python z3_utils.py:z3_implies (lines 109-133).
func (s *Solver) Z3Implies(f1, f2 lg.Expr, timeout bool) (bool, error) {
    key := [2]lg.NodeKey{f1.Sexp(), f2.Sexp()}
    s.mu.Lock()
    if cached, ok := s.impliesCache[key]; ok {
        s.mu.Unlock()
        return cached, nil
    }
    s.mu.Unlock()

    z3solver := s.tr.Ctx.NewSolver()
    if timeout {
        z3solver.SetParam("timeout", "2000")
    }
    zf1, err := s.tr.Translate(f1)
    if err != nil {
        return false, err
    }
    z3solver.Assert(zf1)
    negF2 := &lg.Not{Body: f2}
    zNeg, err := s.tr.Translate(negF2)
    if err != nil {
        return false, err
    }
    z3solver.Assert(zNeg)

    res := z3solver.Check()
    switch res {
    case z3bridge.Sat:
        s.mu.Lock()
        s.impliesCache[key] = false
        s.mu.Unlock()
        return false, nil
    case z3bridge.Unsat:
        s.mu.Lock()
        s.impliesCache[key] = true
        s.mu.Unlock()
        return true, nil
    default:
        return false, fmt.Errorf("z3 returned: %s", res)
    }
}
```

**A3. Rewrite `ImpliesBatch` (lines 648-670)**

Corresponds to Python `z3_utils.py:z3_implies_batch` (lines 136-171).

```go
// ImpliesBatch tests if premise implies each formula in fmlas.
// Uses raw Translate (no closing), matching Python z3_utils.py:to_z3.
// Free variables become shared Z3 constants (not universally quantified).
// Corresponds to Python z3_utils.py:z3_implies_batch (lines 136-171).
func (s *Solver) ImpliesBatch(premise lg.Expr, fmlas []lg.Expr, timeout bool) ([]bool, error) {
    z3solver := s.tr.Ctx.NewSolver()
    if timeout {
        z3solver.SetParam("timeout", "2000")
    }
    zPremise, err := s.tr.Translate(premise)
    if err != nil {
        return nil, err
    }
    z3solver.Assert(zPremise)

    result := make([]bool, len(fmlas))
    for i, f := range fmlas {
        key := [2]lg.NodeKey{premise.Sexp(), f.Sexp()}
        s.mu.Lock()
        if cached, ok := s.impliesCache[key]; ok {
            s.mu.Unlock()
            result[i] = cached
            continue
        }
        s.mu.Unlock()

        negF := &lg.Not{Body: f}
        zNeg, err := s.tr.Translate(negF)
        if err != nil {
            return nil, err
        }
        z3solver.Push()
        z3solver.Assert(zNeg)
        res := z3solver.Check()
        z3solver.Pop()

        switch res {
        case z3bridge.Sat:
            s.mu.Lock()
            s.impliesCache[key] = false
            s.mu.Unlock()
            result[i] = false
        case z3bridge.Unsat:
            s.mu.Lock()
            s.impliesCache[key] = true
            s.mu.Unlock()
            result[i] = true
        default:
            return nil, fmt.Errorf("z3 returned: %s for formula %d", res, i)
        }
    }
    return result, nil
}
```

Key changes from old code:
- `s.tr.Translate()` instead of `s.translateClosed()` — no closing, matching Python's `to_z3`
- Added `timeout bool` parameter
- Added cache via `s.impliesCache`
- Added proper Unknown handling (return error, matching Python's `assert False`)
- Signature changed: `(premise, fmlas)` → `(premise, fmlas, timeout)`

**A4. After all changes, `translateClosed` callers**

After PLAN218 + PLAN219, `translateClosed` has ZERO remaining callers. It can be deleted or kept as dead code (suggest keeping for now since it's harmless, and removing dead code is a separate cleanup).

### B. `tactics/tactics.go` — 4 changes

**B1. Add import for `logicutil` (line 9-18)**

Add to imports block:
```go
lu "github.com/glycerine/ivy/goivy/logicutil"
```

(Already used in `ivy_tactics.go` with same alias, confirming no import cycle.)

**B2. Add `BackgroundTheoryClauses` method (after line 87)**

Python's `_ivy_interp.background_theory()` returns Clauses. Go's `BackgroundTheory()` returns `lg.Expr`. Add a method that returns `*module.Clauses`.

```go
// BackgroundTheoryClauses returns the background theory as Clauses.
// Preserves Clauses structure for and_clauses operations.
// Corresponds to Python _ivy_interp.background_theory().
func (tc *TacticsContext) BackgroundTheoryClauses() *module.Clauses {
    if tc.Mod == nil {
        return module.TrueClauses(nil)
    }
    clauses := tc.Mod.BackgroundTheory(nil)
    if clauses == nil {
        return module.TrueClauses(nil)
    }
    return clauses
}
```

**B3. Rewrite `ImpliedFacts` (lines 154-174)**

Corresponds to Python `tactics_api.py:implied_facts` (lines 310-320).

```go
// ImpliedFacts checks which facts are implied by a premise.
// Returns the subset of factsToCheck that are implied by premise conjoined
// with background axioms.
// Corresponds to Python tactics_api.py:implied_facts (lines 310-320).
func (tc *TacticsContext) ImpliedFacts(premise *module.Clauses, factsToCheck []*module.Clauses) []*module.Clauses {
    if premise == nil || len(factsToCheck) == 0 {
        return nil
    }

    // Python: axioms = _ivy_interp.background_theory()
    axioms := tc.BackgroundTheoryClauses()

    // Python: premise = normalize_quantifiers((and_clauses(axioms, premise)).to_formula())
    combined := module.AndClausesTyped(axioms, premise)
    premFormula := lu.NormalizeQuantifiers(combined.ToFormula())

    // Python: facts_to_check = [f.to_formula() if type(f) is Clauses else f ...]
    formulas := make([]lg.Expr, 0, len(factsToCheck))
    indices := make([]int, 0, len(factsToCheck))
    for i, fact := range factsToCheck {
        if fact != nil {
            formulas = append(formulas, fact.ToFormula())
            indices = append(indices, i)
        }
    }

    // Python: result = z3_implies_batch(premise, facts_to_check, False)
    slv := solver.New()
    results, err := slv.ImpliesBatch(premFormula, formulas, false)
    if err != nil {
        return nil
    }

    // Python: return [f for f, x in zip(facts_to_check, result) if x]
    var implied []*module.Clauses
    for j, isImplied := range results {
        if isImplied {
            implied = append(implied, factsToCheck[indices[j]])
        }
    }
    return implied
}
```

Key changes from old code:
- Uses `BackgroundTheoryClauses()` + `AndClausesTyped` (matching Python's `and_clauses`)
- Applies `lu.NormalizeQuantifiers()` on the combined premise
- Uses `ImpliesBatch` (single solver, push/pop) instead of individual `Implies()` calls
- Filters nil facts before calling batch

**B4. Rewrite `RefutedGoal` (lines 93-115)**

Corresponds to Python `tactics_api.py:refuted_goal` (lines 336-341).

```go
// RefutedGoal checks if a goal has been refuted.
// Corresponds to Python tactics_api.py:refuted_goal (lines 336-341).
func (tc *TacticsContext) RefutedGoal(goal *proof.ProofGoal) bool {
    if goal == nil || goal.Formula == nil {
        return false
    }
    // Quick check: formula is False (empty Or)
    if or, ok := goal.Formula.(*lg.Or); ok && len(or.Terms) == 0 {
        return true
    }
    node, ok := goal.Node.(*art.State)
    if !ok || node == nil || node.Clauses == nil {
        return false
    }

    // Python: axioms = _ivy_interp.background_theory()
    axioms := tc.BackgroundTheoryClauses()

    // Python: premise = (and_clauses(axioms, goal.node.clauses)).to_formula()
    combined := module.AndClausesTyped(axioms, node.Clauses)
    premise := combined.ToFormula()

    // Python: f = Not(goal.formula.to_formula())
    negGoal := &lg.Not{Body: goal.Formula}

    // Python: return z3_implies(premise, f)
    slv := solver.New()
    result, err := slv.Z3Implies(premise, negGoal, false)
    if err != nil {
        return false
    }
    return result
}
```

Key changes:
- Uses `BackgroundTheoryClauses()` + `AndClausesTyped` (matching Python's `and_clauses`)
- Uses `Z3Implies` (cached, raw translate) instead of `Implies`

### C. `solver/solver_test.go` — Add tests

**C1. TestImpliesBatchValid** — p∧q implies both p and q

**C2. TestImpliesBatchInvalid** — p does not imply q

**C3. TestImpliesBatchMixed** — p∧q implies p and q but not r

**C4. TestImpliesBatchEmpty** — empty formula list returns empty results

**C5. TestImpliesBatchCache** — second call hits cache; verify cache entry exists

**C6. TestImpliesBatchTimeout** — p implies p with timeout=true

**C7. TestImpliesBatchFreeVarsShared** — r(X) as premise implies r(X) with shared X (confirms no closing)

**C8. TestZ3ImpliesValid** — p∧q implies p

**C9. TestZ3ImpliesInvalid** — p does not imply q

**C10. TestZ3ImpliesCache** — repeated call hits cache

**C11. TestZ3ImpliesTautology** — true implies p∨¬p

**C12. TestImpliesBatchUnknownHandling** — verify proper error return (may need a formula that triggers Unknown — defer if hard to construct)

### D. `tactics/ivy_tactics_test.go` — Add tests

**D1. TestImpliedFactsBasic** — premise p∧q implies facts p, q but not r

**D2. TestImpliedFactsNilPremise** — nil premise returns nil

**D3. TestImpliedFactsEmptyFacts** — empty factsToCheck returns nil

**D4. TestRefutedGoalBasic** — node with clauses {p} refutes goal ¬p (since p ⇒ ¬(¬p) = p)

**D5. TestRefutedGoalNilGoal** — nil goal returns false

**D6. TestRefutedGoalFalseFormula** — empty Or formula is trivially refuted

## Translation Path Comparison

| Python function | Go function | Translation used | Closing | HASH | Type constraints | Cache |
|----------------|-------------|-----------------|---------|------|-----------------|-------|
| `z3_utils.to_z3` | `s.tr.Translate` | Raw recursive | No | Yes (depth-0) | No | Z3 bridge cache |
| `z3_utils.z3_implies` | `s.Z3Implies` | Raw via Translate | No | Yes | No | `impliesCache` |
| `z3_utils.z3_implies_batch` | `s.ImpliesBatch` | Raw via Translate | No | Yes | No | `impliesCache` |
| `ivy_solver.formula_to_z3` | `s.formulaToZ3` | Full pipeline | Yes (via formulaToZ3Closed) | Yes (explicit) | Yes | No |

Note: Go's `Translate` emits HASH at depth-0 while Python's `to_z3` does not. This is a minor trace difference — the semantics are identical. The HASH is harmless and aids debugging.

## Files Modified

- `solver/solver.go` — add `impliesCache` field + init + clear; add `Z3Implies`; rewrite `ImpliesBatch`
- `tactics/tactics.go` — add `logicutil` import; add `BackgroundTheoryClauses`; rewrite `ImpliedFacts`; rewrite `RefutedGoal`
- `solver/solver_test.go` — add ~12 new tests
- `tactics/ivy_tactics_test.go` — add ~6 new tests

## Verification

```bash
cd ~/go/src/github.com/glycerine/ivy/goivy && go build ./... && go test ./solver/... && go test ./tactics/...
```

Then:
```bash
cd ~/ivy/goivy && make golden
```

Expected: All existing tests pass. ImpliesBatch callers (currently 0 external) get the corrected semantics. New tests verify the rewritten functions.
