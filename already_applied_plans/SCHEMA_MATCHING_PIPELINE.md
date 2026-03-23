# Detailed Plan: Fix Divergences 15-16 (Proof Matching Pipeline)

## Context

The Go proof matching pipeline (`proof/matching.go`, `proof/checker.go`) is missing critical steps from Python's `ivy_proof.py:306-449`. Specifically:

**Divergence 15**: `SetupSchemaMatching` only executes 2 of 6 pipeline steps. It's missing rename_goal, transform_defn_schema, add_prem_match, and compile_match.

**Divergence 16**: `LookupSchema` returns raw definitions without converting them to constraints via `to_constraint()` or closing them with `close_formula()`.

The irony is that the Go codebase already has most of the required functions implemented in `proof/phase5_matching.go` — they're just not wired into the main pipeline.

---

## Key Insight: Phase5 Functions Already Exist

These functions exist in `phase5_matching.go` but are NOT called from the main pipeline:

| Python function | Go function (exists but unused) | Location |
|---|---|---|
| `rename_goal` | `RenameGoal` | phase5_matching.go:928-1024 |
| `transform_defn_schema` | `TransformDefnSchema` | phase5_matching.go:222-251 |
| `add_prem_match` | `AddPremMatch` (broken) | phase5_matching.go:400-420 |
| `compile_match` | `CompileMatchFull` | phase5_matching.go:570-600 |
| `avoid_capture_problem` | `AvoidCaptureProblem` | phase5_matching.go:676-711 |
| `apply_match_goal` | `ApplyMatchGoalNode` | phase5_matching.go:1062-1114 |
| `close_formula` | `il.CloseFormula` | ivylogic/util.go:243-250 |

---

## Part A: Fix Divergence 16 — LookupSchema

### File: `proof/checker.go`

### Step A1: Add `ToConstraint()` method to `logic.Definition`

**File:** `logic/definition.go`

Add method matching Python `ivy_logic.py:236-249`:

```go
func (d *Definition) ToConstraint() Expr {
    // Check if RHS is a Some (conditional definition)
    if some, ok := d.Rhs.(*ivylogic.Some); ok {
        if some.IfVal != nil {
            // some X. phi in ifval else elseval
            // → And(Implies(phi, Exists([X], And(phi, Eq(lhs, ifval)))),
            //       Or(Exists([X], phi), Eq(lhs, elseval)))
            // ... (see detailed code below)
        }
        // simple some X. phi
        // → ForAll([X], Implies(phi, substitute(phi, {X: lhs})))
    }
    // If LHS is an Apply (function def): use Eq(lhs, rhs)
    if _, ok := d.Lhs.(*Apply); ok {
        return NewEq(d.Lhs, d.Rhs)
    }
    // Otherwise: use Iff(lhs, rhs)
    return NewIff(d.Lhs, d.Rhs)
}
```

**Note:** The `Some` case creates a circular dependency (logic → ivylogic). To resolve this:
- Fix: implement `ToConstraint` in `ivylogic/` package as a standalone function `DefinitionToConstraint(d *lg.Definition) lg.Expr` that handles `Some`
- **Recommended**: Standalone function in ivylogic since `Some` lives there

**File:** `ivylogic/constructors.go` (or new file `ivylogic/constraint.go`)

```go
// DefinitionToConstraint converts a Definition to its constraint form.
// Python: ivy_logic.py:236-249
func DefinitionToConstraint(d *lg.Definition) lg.Expr {
    if some, ok := d.Rhs.(*Some); ok {
        return someToConstraint(d.Lhs, some)
    }
    // If LHS is Apply (function definition): Eq(lhs, rhs)
    if _, ok := d.Lhs.(*lg.Apply); ok {
        return &lg.Eq{T1: d.Lhs, T2: d.Rhs}
    }
    // Otherwise: Iff(lhs, rhs)
    return &lg.Iff{T1: d.Lhs, T2: d.Rhs}
}

func someToConstraint(lhs lg.Expr, some *Some) lg.Expr {
    if some.IfVal != nil {
        // Complex case: some X. phi in ifval else elseval
        x := some.Params[0].(*lg.Variable)
        phi := some.Fmla
        ifVal := some.IfVal
        elseVal := some.ElseVal

        // And(
        //   Implies(phi, Exists([x], And(phi, Eq(lhs, ifVal)))),
        //   Or(Exists([x], phi), Eq(lhs, elseVal))
        // )
        existsInner := &lg.Exists{
            Variables: []*lg.Variable{x},
            Body: &lg.And{Terms: []lg.Expr{phi, &lg.Eq{T1: lhs, T2: ifVal}}},
        }
        arm1 := &lg.Implies{T1: phi, T2: existsInner}
        existsOuter := &lg.Exists{Variables: []*lg.Variable{x}, Body: phi}
        arm2 := &lg.Or{Terms: []lg.Expr{existsOuter, &lg.Eq{T1: lhs, T2: elseVal}}}
        return &lg.And{Terms: []lg.Expr{arm1, arm2}}
    }
    // Simple case: some X. phi (no if/else)
    // → ForAll([x], Implies(phi, substitute(phi, {x: lhs})))
    x := some.Params[0].(*lg.Variable)
    phi := some.Fmla
    subs := map[lg.NodeKey]lg.Expr{lg.Key(x): lhs}
    substPhi, _ := lu.Substitute(phi, subs)
    return &lg.ForAll{
        Variables: []*lg.Variable{x},
        Body: &lg.Implies{T1: phi, T2: substPhi},
    }
}
```

### Step A2: Update `LookupSchema` signature and logic

**File:** `proof/checker.go:137-153`

Change from:
```go
func (pc *ProofChecker) LookupSchema(name string, goal *ast.LabeledFormula) (*ast.LabeledFormula, error)
```

To:
```go
func (pc *ProofChecker) LookupSchema(name string, goal *ast.LabeledFormula, errNode ast.Node, close bool) (*ast.LabeledFormula, error)
```

New implementation:
```go
func (pc *ProofChecker) LookupSchema(name string, goal *ast.LabeledFormula, errNode ast.Node, close bool) (*ast.LabeledFormula, error) {
    if s, ok := pc.Schemata[name]; ok {
        if err := CheckSchemaCapture(s, goal); err != nil {
            return nil, err
        }
        return s, nil
    }
    if d, ok := pc.Definitions[name]; ok {
        // Convert definition to constraint — Python: goal_conc(schema).to_constraint()
        conc := GoalConc(d)
        if def, ok := conc.(*lg.Definition); ok {
            fmla := il.DefinitionToConstraint(def)
            if close {
                fmla = il.CloseFormula(fmla)
            }
            schema := CloneGoal(d, GoalPrems(d), fmla)
            if err := CheckSchemaCapture(schema, goal); err != nil {
                return nil, err
            }
            return schema, nil
        }
        // Not a *lg.Definition — return as-is
        return d, nil
    }
    // Check goal premises
    for _, pg := range GoalPremGoals(goal) {
        if pg.LabelName() == name {
            return pg, nil
        }
    }
    return nil, &ProofError{
        Node: errNode,
        Msg:  "No property " + name + " exists in the current context",
    }
}
```

### Step A3: Update all LookupSchema call sites

Search for all calls to `LookupSchema` and add the new parameters:

1. **`matching.go:17`** — `SetupMatching`: `pc.LookupSchema(schemaName, decl)` → `pc.LookupSchema(schemaName, decl, nil, false)`
2. **`tactics.go` assume/unfold tactics** — update similarly, passing the proof node and `close=false`
3. **`AddPremMatch`** (once rewritten) — will call `checker.LookupSchema(rhs.rep, goal, rhs, false)` matching Python

---

## Part B: Fix Divergence 15 — Schema Matching Pipeline

### Step B1: Update `MatchSchema` to accept full proof object

**File:** `proof/checker.go:264-306`

Change signature from:
```go
func (pc *ProofChecker) MatchSchema(goal *ast.LabeledFormula, schemaName string) ([]*ast.LabeledFormula, error)
```
To:
```go
func (pc *ProofChecker) MatchSchema(goal *ast.LabeledFormula, proof *ast.SchemaInstantiation) ([]*ast.LabeledFormula, error)
```

This gives access to `proof.SchemaName`, `proof.Ren` (renaming), and `proof.Matches`.

### Step B2: Update `ApplyProof` call site

**File:** `proof/checker.go:183-192`

Change from:
```go
case *ast.SchemaInstantiation:
    sname := nodeToString(p.SchemaName)
    m, err := pc.MatchSchema(goals[0], sname)
```
To:
```go
case *ast.SchemaInstantiation:
    m, err := pc.MatchSchema(goals[0], p)
```

### Step B3: Update `SetupMatching` to accept proof object

**File:** `proof/matching.go:16-22`

Change from:
```go
func (pc *ProofChecker) SetupMatching(decl *ast.LabeledFormula, schemaName string) (*MatchProblem, map[lg.NodeKey]lg.Expr, error) {
    schema, err := pc.LookupSchema(schemaName, decl)
    ...
    return pc.SetupSchemaMatching(decl, schema)
}
```
To:
```go
func (pc *ProofChecker) SetupMatching(decl *ast.LabeledFormula, proof *ast.SchemaInstantiation, mod *module.Module) (*MatchProblem, map[lg.NodeKey]lg.Expr, error) {
    schemaName := nodeToString(proof.SchemaName)
    schema, err := pc.LookupSchema(schemaName, decl, proof, false)
    if err != nil {
        return nil, nil, err
    }
    return pc.SetupSchemaMatching(decl, proof, schema, false, mod)
}
```

### Step B4: Rewrite `SetupSchemaMatching` with full pipeline

**File:** `proof/matching.go:24-44`

Replace current implementation with the full Python pipeline:

```go
// SetupSchemaMatching implements the complete Python pipeline from ivy_proof.py:329-340:
//   1. rename_goal(schema, proof.renaming())
//   2. transform_defn_schema(schema, decl)
//   3. match_problem(schema, decl)
//   4. transform_defn_match(prob)
//   5. add_prem_match(proof.match(), prob, decl, self)
//   6. compile_match(proof_match, prob, decl, allow_witness)
func (pc *ProofChecker) SetupSchemaMatching(
    decl *ast.LabeledFormula,
    proof *ast.SchemaInstantiation,
    schema *ast.LabeledFormula,
    allowWitness bool,
    mod *module.Module,
) (*MatchProblem, map[lg.NodeKey]lg.Expr, error) {

    // Step 1: Rename schema using proof renaming
    if proof != nil && proof.Ren != nil {
        var err error
        schema, err = RenameGoal(schema, proof.Ren)
        if err != nil {
            return nil, nil, err
        }
    }

    // Step 2: Transform definition schema for parameter arity matching
    schema = TransformDefnSchema(schema, decl)

    // Step 3: Build match problem
    prob := buildMatchProblem(schema, decl)
    if prob == nil {
        return nil, nil, &NoMatch{Msg: "cannot build match problem"}
    }

    // Step 4: Transform definition match (full version from phase5)
    prob = TransformDefnMatch(prob)
    if prob == nil {
        return nil, nil, &NoMatch{Node: proof, Msg: "definition does not match the given schema"}
    }

    // Step 5: Process premise matches
    var proofMatches []ast.Node
    if proof != nil {
        proofMatches = proof.Matches
    }
    proofMatches, prob = AddPremMatch(proofMatches, prob, decl, pc)

    // Step 6: Compile symbolic matches
    pmatch := CompileMatchFull(proofMatches, prob, decl, allowWitness, mod)
    if pmatch == nil && len(proofMatches) > 0 {
        return nil, nil, &ProofError{Node: proof, Msg: "Match is inconsistent"}
    }
    if pmatch == nil {
        pmatch = make(map[lg.NodeKey]lg.Expr)
    }

    return prob, pmatch, nil
}
```

### Step B5: Rewrite `AddPremMatch` to match Python

**File:** `proof/phase5_matching.go:400-420`

The current Go implementation is a no-op. Rewrite to match Python `ivy_proof.py:835-855`:

```go
// AddPremMatch processes premise matches in a SchemaInstantiation.
// For each match where LHS is a premise name in the schema and RHS is a schema name,
// looks up the RHS schema and creates a combined Tuple pattern/instance.
// Python: ivy_proof.py:835-855
func AddPremMatch(proofMatch []ast.Node, prob *MatchProblem, goal *ast.LabeledFormula, checker *ProofChecker) ([]ast.Node, *MatchProblem) {
    if prob.SchemaLF == nil {
        return proofMatch, prob
    }
    sprems := GoalPremsByName(prob.SchemaLF)

    var pats []lg.Expr
    var insts []lg.Expr
    var newMatch []ast.Node

    for _, m := range proofMatch {
        defn, ok := m.(*ast.Definition)
        if !ok {
            newMatch = append(newMatch, m)
            continue
        }
        lAtom, lIsAtom := defn.Lhs.(*ast.Atom)
        if !lIsAtom || len(lAtom.Terms) > 0 {
            newMatch = append(newMatch, m)
            continue
        }
        sprem, exists := sprems[lAtom.Relname()]
        if !exists {
            newMatch = append(newMatch, m)
            continue
        }
        rAtom, rIsAtom := defn.Rhs.(*ast.Atom)
        if !rIsAtom || len(rAtom.Terms) > 0 {
            newMatch = append(newMatch, m)
            continue
        }
        // Look up the RHS as a schema
        gprem, err := checker.LookupSchema(rAtom.Relname(), goal, rAtom, false)
        if err != nil {
            newMatch = append(newMatch, m)
            continue
        }
        premConc := GoalConc(sprem)
        gpremConc := GoalConc(gprem)
        if premConc != nil && gpremConc != nil {
            pats = append(pats, premConc)
            insts = append(insts, gpremConc)
        } else {
            newMatch = append(newMatch, m)
        }
    }

    if len(pats) > 0 {
        // Combine premise patterns with main pattern into "Tuple" form
        // We use slices since Go doesn't need a Tuple wrapper — MatchProblem
        // tracks premise patterns via PremMatches
        allPats := append(pats, prob.Pat)
        allInsts := append(insts, prob.Inst)
        prob = &MatchProblem{
            Schema:      prob.Schema,
            SchemaLF:    prob.SchemaLF,
            Pat:         prob.Pat,       // main pattern stays for non-tuple path
            Inst:        prob.Inst,      // main instance stays
            FreeSyms:    prob.FreeSyms,
            Constants:   prob.Constants,
            PremMatches: pats,           // track premise patterns
            RevMap:      prob.RevMap,
        }
        // Store combined patterns for tuple matching
        prob.TuplePats = allPats
        prob.TupleInsts = allInsts
    }

    return newMatch, prob
}
```

### Step B6: Add TuplePats/TupleInsts fields to MatchProblem

**File:** `proof/match.go:12-22`

Add fields:
```go
type MatchProblem struct {
    Schema      lg.Expr
    SchemaLF    *ast.LabeledFormula
    Pat         lg.Expr
    Inst        lg.Expr
    FreeSyms    map[lg.NodeKey]lg.Expr
    Constants   map[lg.NodeKey]lg.Expr
    PremMatches []lg.Expr
    RevMap      map[lg.NodeKey]lg.Expr
    // Added for premise matching (Tuple pattern support):
    TuplePats   []lg.Expr  // nil = no tuple matching; set by AddPremMatch
    TupleInsts  []lg.Expr
}
```

### Step B7: Update `MatchSchema` to handle Tuple patterns

**File:** `proof/checker.go:264-306`

Rewrite to match Python `ivy_proof.py:412-449`:

```go
func (pc *ProofChecker) MatchSchema(goal *ast.LabeledFormula, proof *ast.SchemaInstantiation) ([]*ast.LabeledFormula, error) {
    // Reject temporal goals
    goalConc := GoalConc(goal)
    if goalConc == nil {
        return nil, &NoMatch{Msg: "goal has no conclusion"}
    }

    // Step 1: Build match problem (full pipeline)
    prob, pmatch, err := pc.SetupMatching(goal, proof, nil) // nil mod for now
    if err != nil {
        return nil, err
    }

    // Step 2: Apply initial proof match to problem
    if len(pmatch) > 0 {
        AvoidCaptureProblem(prob, pmatch)
        ApplyMatchToProblem(pmatch, prob)
    }

    // Step 3+4: Match (with Tuple handling for premise matches)
    if prob.TuplePats != nil {
        // Tuple matching: match each (pat, inst) pair
        for i := range prob.TuplePats {
            pat := prob.TuplePats[i]
            inst := prob.TupleInsts[i]
            // Apply current match state to pat/inst
            pat = ApplyMatch(prob.FreeSyms, pat)  // no-op, just for consistency

            fomatch := FOMatch(pat, inst, prob.FreeSyms, prob.Constants)
            if fomatch != nil && len(fomatch) > 0 {
                ApplyMatchToProblem(fomatch, prob)
                // Also update tuple patterns
                for j := i + 1; j < len(prob.TuplePats); j++ {
                    prob.TuplePats[j] = ApplyMatch(fomatch, prob.TuplePats[j])
                    prob.TupleInsts[j] = ApplyMatch(fomatch, prob.TupleInsts[j])
                }
            }
            somatch := Match(pat, inst, prob.FreeSyms, prob.Constants)
            if somatch == nil {
                return nil, &NoMatch{Node: proof, Msg: "goal does not match the given schema"}
            }
            if len(somatch) > 0 {
                ApplyMatchToProblem(somatch, prob)
                for j := i + 1; j < len(prob.TuplePats); j++ {
                    prob.TuplePats[j] = ApplyMatchAlt(somatch, prob.TuplePats[j], nil)
                    prob.TupleInsts[j] = ApplyMatchAlt(somatch, prob.TupleInsts[j], nil)
                }
            }
        }
    } else {
        // Non-tuple: single pattern matching (existing logic)
        fomatch := FOMatch(prob.Pat, prob.Inst, prob.FreeSyms, prob.Constants)
        if fomatch != nil && len(fomatch) > 0 {
            ApplyMatchToProblem(fomatch, prob)
        }
        somatch := Match(prob.Pat, prob.Inst, prob.FreeSyms, prob.Constants)
        if somatch == nil {
            return nil, &NoMatch{Node: proof, Msg: "goal does not match the given schema"}
        }
        if len(somatch) > 0 {
            ApplyMatchToProblem(somatch, prob)
        }
    }

    // Step 5: Detect nonce symbol clashes
    if err := DetectNonceSymbols(prob); err != nil {
        return nil, err
    }

    // Step 6: Extract subgoals
    if prob.SchemaLF == nil {
        return nil, &NoMatch{Msg: "schema is not a labeled formula after matching"}
    }
    return GoalSubgoalsFromSchema(prob.SchemaLF, goal), nil
}
```

### Step B8: Update `ApplyMatchToProblem` to use AvoidCapture and ApplyMatchGoalNode

**File:** `proof/matching.go:104-134`

Replace with:
```go
func ApplyMatchToProblem(match map[lg.NodeKey]lg.Expr, prob *MatchProblem) {
    if len(match) == 0 {
        return
    }

    // Avoid capture before applying — Python: avoid_capture_problem(prob, match)
    AvoidCaptureProblem(prob, match)

    // Apply match to schema — use ApplyMatchGoalNode (not just conclusion)
    if prob.SchemaLF != nil {
        prob.SchemaLF = ApplyMatchGoalNode(match, prob.SchemaLF)
    }

    // Apply match to pattern — use ApplyMatchAlt for capture safety
    prob.Pat = ApplyMatchAlt(match, prob.Pat, nil)

    // Update free symbols
    prob.FreeSyms = ApplyMatchFreesymsAlt(match, prob.FreeSyms)

    // Remove matched symbols from revmap
    for k := range prob.RevMap {
        if _, matched := match[k]; matched {
            delete(prob.RevMap, k)
        }
    }
}
```

### Step B9: Delete simplified `transformDefnMatch` from matching.go

**File:** `proof/matching.go:81-102`

Delete the simplified `transformDefnMatch` function. The full `TransformDefnMatch` in `phase5_matching.go:261-367` will be used instead.

### Step B10: Update remaining call sites

1. **`InstSchema`** (`checker.go:312-318`): Update to work with new MatchSchema signature
2. **`CheckSchema`** (`checker.go:320-329`): Update similarly
3. **Any test files** that call MatchSchema or SetupMatching: update signatures

---

## Part C: Additional Supporting Changes

### Step C1: Add `module.Module` parameter threading

Several functions in the pipeline need a `*module.Module` parameter for compilation. This needs to be threaded through:

- `MatchSchema` → `SetupMatching` → `SetupSchemaMatching` → `CompileMatchFull`

The module can be obtained from the ProofChecker. Add a `Mod *module.Module` field to ProofChecker or pass it as parameter.

**Recommended:** Add `Mod *module.Module` field to `ProofChecker` struct, set during construction.

### Step C2: Ensure `CompileMatchList` has proper Definition unwrapping

**File:** `proof/phase5_matching.go:512-544`

The `CompileMatchList` function receives `proofMatch []ast.Node` where each element is an `ast.Definition`. Verify it correctly extracts LHS/RHS for compilation.

### Step C3: Verify `ProofError` has Node field

**File:** `proof/errors.go`

Ensure `ProofError` struct has a `Node ast.Node` field for error location reporting. Current code likely has this but verify.

---

## Files Modified (Summary)

| File | Changes |
|------|---------|
| `proof/checker.go` | Update MatchSchema, ApplyProof, LookupSchema signatures; add Mod field to ProofChecker |
| `proof/matching.go` | Rewrite SetupMatching, SetupSchemaMatching, ApplyMatchToProblem; delete simplified transformDefnMatch |
| `proof/match.go` | Add TuplePats/TupleInsts fields to MatchProblem |
| `proof/phase5_matching.go` | Rewrite AddPremMatch to match Python |
| `ivylogic/constructors.go` (or new `ivylogic/constraint.go`) | Add `DefinitionToConstraint` function |

## Existing Functions to Reuse (Already Implemented)

| Function | Location | Purpose |
|----------|----------|---------|
| `RenameGoal` | phase5_matching.go:928 | Step 1 of pipeline |
| `TransformDefnSchema` | phase5_matching.go:222 | Step 2 of pipeline |
| `TransformDefnMatch` | phase5_matching.go:261 | Step 4 (full version) |
| `CompileMatchFull` | phase5_matching.go:570 | Step 6 of pipeline |
| `AvoidCaptureProblem` | phase5_matching.go:676 | Capture avoidance |
| `ApplyMatchGoalNode` | phase5_matching.go:1062 | Apply match to full goal |
| `ApplyMatchAlt` | phase5_matching.go:737 | Apply match with capture checking |
| `ApplyMatchFreesymsAlt` | phase5_matching.go:914 | Update freesyms |
| `CheckSchemaCapture` | phase5_goals.go:212 | Schema capture validation |
| `GoalPremsByName` | phase5_goals.go:643 | Premise lookup by name |
| `CloseFormula` | ivylogic/util.go:243 | Close formula with forall |
| `DropSuppliedPrems` | phase5_goals.go:580 | Remove supplied premises |
| `CloseUnmatched` | phase5_goals.go:548 | Close unmatched vars |

---

## Verification Plan

1. **Compile check:** `go build ./...` — ensure no type errors from signature changes
2. **Run existing tests:** `go test ./proof/...` — verify no regressions
3. **Test LookupSchema with definitions:**
   - Create a definition `f(X) = X + 1`
   - Look it up — verify it returns `Eq(f(X), X+1)` not the raw Definition
   - Look it up with `close=true` — verify it returns `ForAll X. Eq(f(X), X+1)`
4. **Test schema matching with renaming:**
   - Create a schema with symbols A, B
   - Create a SchemaInstantiation with renaming A→C
   - Verify the matched schema uses C instead of A
5. **Test premise matching:**
   - Create a schema with premise `[p] A -> B` and conclusion `C`
   - Create a proof that supplies `p = some_axiom`
   - Verify the premise is matched and removed from subgoals
6. **End-to-end:** Run any `.ivy` test files that exercise proof scripts through both Python and Go
