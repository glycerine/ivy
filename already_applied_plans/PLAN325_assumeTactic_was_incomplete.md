# PLAN: Faithful port of Python assume_tactic to Go assumeTactic

Created: 2026-04-17, 22:00

## Context

Go `goivy_check ord_live.ivy` fails at temporal property `cf_pio_live.cf_liveness` (line 2521) with:
```
error: check/l2s: [2]proof goal is not temporal
```

The property's proof uses `instantiate cfabric_pio_fair_ax` which dispatches to `assumeTactic`. The current Go `assumeTactic` (proof/tactics.go:77-127) is a **drastically simplified stub** that skips the entire matching pipeline. The Python `assume_tactic` (ivy_proof.py:350-382) performs schema matching, witness substitution, variable closing, premise dropping, and label handling. This divergence likely causes the goal to be malformed before reaching l2s.

## Current Go stub vs Python source of truth

### Python `assume_tactic` does (ivy_proof.py:350-382):
1. Schema lookup with AssumeTactic vs AssumeGlobalTactic distinction
2. `remove_explicit(schema)` — clears explicit flag
3. `self.setup_schema_matching(decl, proof, schema, allow_witness=True)` — full pipeline
4. Extract witnesses from pmatch (variables not in freesyms)
5. `close_unmatched(prem, pmatch)` — universally close unmatched vars
6. `lu.witness_ast(True, [], witness, conc)` — apply witness substitutions
7. `apply_match_goal(pmatch, prem, apply_match_alt)` — apply the match
8. `drop_supplied_prems(prem, decl, proof.match())` — drop supplied premises
9. Label handling (NoneAST creates atom with schema label args; else uses proof.label)
10. Clash detection with rename for AssumeGlobalTactic, error for AssumeTactic

### Current Go `assumeTactic` does:
1. Schema lookup (no AssumeGlobalTactic distinction)
2. Direct add as premise (skips steps 2-8)
3. Simple label clone (incorrect NoneAST handling)
4. Simple clash check (error only, no rename)

## All building blocks already exist in Go

| Python function | Go equivalent | Location |
|---|---|---|
| `remove_explicit` | `RemoveExplicit` | proof/phase5_goals.go:676 |
| `setup_schema_matching` | `SetupSchemaMatching` | proof/matching.go:32 (needs `Ren`/`Matches` adapter) |
| `close_unmatched` | `CloseUnmatched` | proof/phase5_goals.go:597 |
| `lu.witness_ast` | `module.WitnessAst` | module/skolem.go:215 |
| `apply_match_goal(m,g,apply_match_alt)` | `ApplyMatchGoalNode` | proof/phase5_matching.go:1189 |
| `drop_supplied_prems` | `DropSuppliedPrems` | proof/phase5_goals.go:636 |
| `rename_prem_no_clash` | `RenamePremNoClash` | proof/phase5_goals.go:687 |
| `goal_add_prem` | `goalAddPrem` | proof/tactics.go:520 |
| `goal_remove_prem` | needs porting | — |
| `rename_goal` | `RenameGoal` | proof/phase5_matching.go |
| `transform_defn_schema` | `TransformDefnSchema` | proof/phase5_matching.go |
| `match_problem` | `buildMatchProblem` | proof/matching.go |
| `transform_defn_match` | `TransformDefnMatch` | proof/phase5_matching.go |
| `add_prem_match` | `AddPremMatch` | proof/phase5_matching.go |
| `compile_match` | `CompileMatchFull` | proof/phase5_matching.go:695 |

## Changes

### 1. Add `SetupSchemaMatchingRaw` to proof/matching.go

`SetupSchemaMatching` currently takes `*ast.SchemaInstantiation`. We need a version that takes raw `ren ast.Node` and `matches []ast.Node` fields since both `AssumeTactic` and `SchemaInstantiation` have these same fields. This avoids adding an interface (per CLAUDE.md rule 7).

```go
// SetupSchemaMatchingRaw is SetupSchemaMatching but takes raw ren/matches
// fields instead of a SchemaInstantiation. Used by assumeTactic.
// Python: setup_schema_matching (ivy_proof.py:337-348).
func (pc *ProofChecker) SetupSchemaMatchingRaw(
    decl *ast.LabeledFormula,
    ren ast.Node,
    matches []ast.Node,
    schema *ast.LabeledFormula,
    allowWitness bool,
) (*MatchProblem, map[lg.NodeKey]lg.Expr, error) {
    // Step 1: rename_goal(schema, proof.renaming())
    if ren != nil {
        var err error
        schema, err = RenameGoal(pc.astCfg(), schema, ren)
        if err != nil {
            return nil, nil, err
        }
    }
    // Steps 2-6: same as SetupSchemaMatching
    schema = TransformDefnSchema(pc.astCfg(), schema, decl)
    prob := buildMatchProblem(schema, decl)
    if prob == nil {
        return nil, nil, &NoMatch{Msg: "cannot build match problem"}
    }
    prob = TransformDefnMatch(pc.astCfg(), prob)
    if prob == nil {
        return nil, nil, &NoMatch{Msg: "definition does not match the given schema"}
    }
    proofMatches := matches
    proofMatches, prob = AddPremMatch(proofMatches, prob, decl, pc)
    pmatch := CompileMatchFull(proofMatches, prob, decl, allowWitness, pc.Mod)
    if pmatch == nil && len(proofMatches) > 0 {
        return nil, nil, &ProofError{Msg: "Match is inconsistent"}
    }
    if pmatch == nil {
        pmatch = make(map[lg.NodeKey]lg.Expr)
    }
    return prob, pmatch, nil
}
```

Also refactor existing `SetupSchemaMatching` to call `SetupSchemaMatchingRaw` to avoid code duplication.

### 2. Add `GoalRemovePrem` to proof/goal.go

```go
// GoalRemovePrem removes a premise by name from a goal.
// Python: goal_remove_prem (ivy_proof.py:529-531).
func GoalRemovePrem(cfg *ast.AstConfig, goal *ast.LabeledFormula, premName string) *ast.LabeledFormula {
    var kept []ast.Node
    for _, p := range GoalPrems(goal) {
        if lf, ok := p.(*ast.LabeledFormula); ok && lf.LabelName() == premName {
            continue
        }
        kept = append(kept, p)
    }
    return CloneGoal(cfg, goal, kept, GoalConc(goal))
}
```

### 3. Rewrite `assumeTactic` in proof/tactics.go

Replace the current stub (lines 77-127) with a faithful port:

```go
// assumeTactic introduces an assumption from a schema or premise.
// Faithful port of Python ProofChecker.assume_tactic (ivy_proof.py:350-382).
func (pc *ProofChecker) assumeTactic(decls []*ast.LabeledFormula, proof *ast.AssumeTactic) ([]*ast.LabeledFormula, error) {
    if len(decls) == 0 {
        return nil, &ProofError{Msg: "assume tactic: no goals"}
    }
    decl := decls[0]

    // Python: schemaname = proof.schemaname()
    schemaName := ""
    if proof.SchemaName != nil {
        schemaName = fmt.Sprint(proof.SchemaName)
    }
    if schemaName == "" {
        return nil, &ProofError{Msg: "assume tactic: no schema name"}
    }

    // Python: premmap = dict((x.name,x) for x in goal_prem_goals(decl))
    premMap := make(map[string]*ast.LabeledFormula)
    for _, pg := range GoalPremGoals(decl) {
        premMap[pg.LabelName()] = pg
    }

    // Python lines 354-359: AssumeGlobalTactic vs AssumeTactic distinction
    _, isGlobal := proof.AsAssumeTacticNode().(*ast.AssumeGlobalTactic)
    var schema *ast.LabeledFormula
    if !isGlobal {
        if prem, ok := premMap[schemaName]; ok {
            schema = prem
            if isNoneAST(proof.TLabel) {
                decl = GoalRemovePrem(pc.astCfg(), decl, schemaName)
            }
        }
    }
    if schema == nil {
        var err error
        schema, err = pc.LookupSchema(schemaName, decl, proof, false)
        if err != nil {
            return nil, &ProofError{Node: proof, Msg: fmt.Sprintf(
                "No property %s exists in the current context", schemaName)}
        }
    }

    // Python: schema = remove_explicit(schema)
    schema = RemoveExplicit(schema)

    // Python: prob, pmatch = self.setup_schema_matching(decl, proof, schema, allow_witness=True)
    prob, pmatch, err := pc.SetupSchemaMatchingRaw(decl, proof.Ren, proof.Matches, schema, true)
    if err != nil {
        return nil, err
    }

    // Python lines 365-367: extract witnesses (variables not in freesyms)
    witness := make(map[lg.NodeKey]lg.Expr)
    pmatchClean := make(map[lg.NodeKey]lg.Expr)
    for k, v := range pmatch {
        if isWitVar(k, v, prob) {
            witness[k] = v
        } else {
            pmatchClean[k] = v
        }
    }
    pmatch = pmatchClean

    // Python: prem = prob.schema
    prem := prob.SchemaLF

    // Python: if schemaname not in premmap: prem = close_unmatched(prem, pmatch)
    if _, inPrems := premMap[schemaName]; !inPrems {
        prem = CloseUnmatched(pc.astCfg(), prem, pmatch)
    }

    // Python: conc = goal_conc(prem)
    //         conc = lu.witness_ast(True, [], witness, conc)
    //         prem = clone_goal(prem, goal_prems(prem), conc)
    if len(witness) > 0 {
        rawConc := GoalConc(prem)
        if concExpr, ok := rawConc.(lg.Expr); ok {
            newConc, werr := module.WitnessAst(true, nil, witness, concExpr)
            if werr == nil {
                prem = CloneGoal(pc.astCfg(), prem, GoalPrems(prem), newConc)
            }
        }
    }

    // Python: prem = apply_match_goal(pmatch, prem, apply_match_alt)
    prem = ApplyMatchGoalNode(pc.astCfg(), pmatch, prem)

    // Python: prem = drop_supplied_prems(prem, decl, proof.match())
    prem = DropSuppliedPrems(pc.astCfg(), prem, decl, proof.Matches)

    // Python lines 374-377: label handling
    if !isNoneAST(proof.TLabel) {
        prem = prem.CloneWithFreshID([]ast.Node{proof.TLabel, prem.Formula})
    }

    // Python lines 378-381: clash detection
    for _, pg := range GoalPremGoals(decl) {
        if pg.LabelName() == prem.LabelName() {
            if isGlobal {
                prem = RenamePremNoClash(prem, decl)
            } else {
                return nil, &ProofError{Node: proof, Msg: fmt.Sprintf(
                    "instance name %s clashes with context", prem.LabelName())}
            }
            break
        }
    }

    // Python: return [goal_add_prem(decl, prem, proof.lineno)] + decls[1:]
    newGoal := pc.goalAddPrem(decl, prem, proof.GetLineno())
    result := []*ast.LabeledFormula{newGoal}
    result = append(result, decls[1:]...)
    return result, nil
}
```

### 4. Small helpers needed

```go
// isNoneAST checks if a node is a NoneAST (or nil).
func isNoneAST(n ast.Node) bool {
    if n == nil { return true }
    _, ok := n.(*ast.NoneAST)
    return ok
}

// isWitVar checks if a match entry is a witness variable.
// Python: iswit = lambda x: isinstance(x, il.Variable) and x not in prob.freesyms
func isWitVar(key lg.NodeKey, val lg.Expr, prob *MatchProblem) bool {
    if _, isVar := val.(*lg.Variable); !isVar {
        return false
    }
    // The key must also be a variable key
    _, inFree := prob.FreeSyms[key]
    return !inFree
}
```

### 5. Fix `ApplyProof` dispatch for `*ast.AssumeGlobalTactic`

The `ApplyProof` switch in checker.go (line 267) currently has:
```go
case *ast.AssumeTactic:
    return pc.assumeTactic(goals, p)
```

Due to Go's type switch semantics, `*ast.AssumeGlobalTactic` (which embeds `AssumeTactic`) will NOT match `*ast.AssumeTactic`. We need to add an explicit case BEFORE the AssumeTactic case:
```go
case *ast.AssumeGlobalTactic:
    return pc.assumeTactic(goals, &p.AssumeTactic)
```

Wait — actually we need the `isGlobal` flag inside assumeTactic. Better approach: change the method signature or pass the original proof node. Simplest: the assumeTactic can take `ast.Node` and type-switch internally, or we store the original proof for the isinstance check.

Actually, simplest: pass the original `proof` node alongside the `*ast.AssumeTactic` fields:

In the ApplyProof dispatch:
```go
case *ast.AssumeGlobalTactic:
    return pc.assumeTactic(goals, &p.AssumeTactic, true)
case *ast.AssumeTactic:
    return pc.assumeTactic(goals, p, false)
```

And change `assumeTactic` signature to take `isGlobal bool`.

## Files to modify

1. **`proof/tactics.go`** — Rewrite `assumeTactic` (lines 77-127), add helpers `isNoneAST`, `isWitVar`
2. **`proof/matching.go`** — Add `SetupSchemaMatchingRaw`, refactor `SetupSchemaMatching` to call it
3. **`proof/goal.go`** — Add `GoalRemovePrem`
4. **`proof/checker.go`** — Fix ApplyProof dispatch for `*ast.AssumeGlobalTactic` (line 267)

## 6. Unit tests for assumeTactic — proof/tactics_assume_test.go

Tests to prevent regressions on the assume tactic. Each test constructs a minimal goal and proof, calls `assumeTactic`, and verifies the result preserves the expected structure.

### Test cases

1. **TestAssumeTactic_BasicInstantiate** — Simple `instantiate schemaName` with no renaming or matches. Verifies the schema is added as a premise and the conclusion (including TemporalModels) is preserved unchanged.

2. **TestAssumeTactic_PreservesTemporalModels** — The critical regression test. Goal conclusion is `*ast.TemporalModels`. After `assumeTactic`, verifies `GoalConc(result)` is still `*ast.TemporalModels`.

3. **TestAssumeTactic_WithSchemaBody** — Goal already has premises (SchemaBody wrapping). After `assumeTactic`, verifies the new premise is appended and conclusion is still the last element.

4. **TestAssumeTactic_RemoveExplicit** — Schema has `Explicit=true`. After `assumeTactic`, verifies the premise added to the goal has `Explicit=false`.

5. **TestAssumeTactic_PremiseFromGoal** — Schema name matches a premise in the goal. Verifies the premise is used (not looked up globally) and removed from original premises when label is NoneAST.

6. **TestAssumeTactic_ClashError** — AssumeTactic (not global) with premise name clashing an existing premise. Verifies ProofError is returned.

7. **TestAssumeTactic_GlobalRename** — AssumeGlobalTactic with name clash. Verifies the premise is renamed (not errored) via `RenamePremNoClash`.

8. **TestAssumeTactic_NoneASTLabel** — When proof.TLabel is NoneAST, verifies the schema's original label is preserved (not replaced with NoneAST).

9. **TestAssumeTactic_ExplicitLabel** — When proof.TLabel is a real label, verifies the premise gets the new label.

10. **TestAssumeTactic_ComposedWithSkolemize** — Integration test: compose `skolemize` then `instantiate` on a goal with `ForAll` + `TemporalModels` conclusion. Verifies the final goal still has `TemporalModels` as its conclusion. This is the exact pattern that triggered the original bug.

## Files to modify

1. **`proof/tactics.go`** — Rewrite `assumeTactic` (lines 77-127), add helpers `isNoneAST`, `isWitVar`
2. **`proof/matching.go`** — Add `SetupSchemaMatchingRaw`, refactor `SetupSchemaMatching` to call it
3. **`proof/goal.go`** — Add `GoalRemovePrem`
4. **`proof/checker.go`** — Fix ApplyProof dispatch for `*ast.AssumeGlobalTactic` (line 267)
5. **`proof/tactics_assume_test.go`** — New file: unit tests for assumeTactic

## Verification

```bash
cd ~/ivy/goivy
# Build
make build

# Run new unit tests
go test ./proof/ -run TestAssumeTactic -v

# Run all proof package tests
go test ./proof/ -v

# Run on ord_live.ivy
XTRACE_OFF=1 time ./goivy_check ord_live.ivy
# Should get past cf_pio_live.cf_liveness without "proof goal is not temporal"

# Run full test suite for regressions
make test
```
