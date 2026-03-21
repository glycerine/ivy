# Plan: Complete CheckPropertiesPass (§6.2 #6)

## Context

The Go port has two versions of `check_properties`:
1. **`CheckPropertiesPass`** in `compiler/ivy_compile.go:1451` — old stub, currently called at line 167. Bypasses proof checker entirely.
2. **`CheckProperties`** in `compiler/phase6.go:1745` — newer partial impl with `namedTrans`, `nmap`, empty-proof-for-schemata. Still skips the actual `ProofChecker` calls (`_ = pf` at line 1807). Not wired into the call site.

Neither version calls `prover.AdmitProposition()`. The Python source of truth (`ivy_compiler.py:1972-2053`) creates a `ProofChecker`, calls `admit_proposition` for every property, handles subgoals with `Labeler`+`ComposeAtoms`, and finishes with `apply_assert_proofs(mod, prover)`.

## Implementation Steps

### Step 1: Add `AdmitDefinition` to `proof/checker.go`

Port Python `ivy_proof.py:70-96`. New method on `ProofChecker`:

```go
func (pc *ProofChecker) AdmitDefinition(defn *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)
```

Logic:
1. `defn = NormalizeGoal(defn)`
2. Extract sym via `defn.Formula.(*lg.Definition).Defines()` → `*lg.Symbol`
3. If `pc.Definitions[sym.Name]` exists → return `&Redefinition{...}`
4. If `pc.Stale[sym.Name]` → return `&Circular{...}`
5. Get deps via `clauseops.SymbolsAST(def.Rhs)`, update `pc.Stale` with dep names
6. If sym is in deps (recursive): require proof, call `pc.ApplyProof([defn], proof)`
7. Else: `subgoals = nil`
8. `pc.Definitions[sym.Name] = defn`
9. Return subgoals

### Step 2: Add `AdmitProposition` to `proof/checker.go`

Port Python `ivy_proof.py:98-121`:

```go
func (pc *ProofChecker) AdmitProposition(prop *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)
```

Logic:
1. `prop = NormalizeGoal(prop)`
2. If `prop.Formula` is `*lg.Definition` → delegate to `AdmitDefinition(prop, proof)`
3. If `proof == nil` → return `&NoMatch{...}`
4. `subgoals = [prop]`, then `subgoals, err = pc.ApplyProof(subgoals, proof)`
5. If subgoals nil → return `&NoMatch{...}`
6. Append prop to `pc.Axioms`, register in `pc.Schemata[prop.LabelName()]`
7. Update `pc.Stale` from `GoalVocab(prop).Symbols`
8. Return subgoals

### Step 3: Add `GetSubgoals` to `proof/checker.go`

Port Python `ivy_proof.py:123-134`:

```go
func (pc *ProofChecker) GetSubgoals(prop *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)
```

Logic: normalize, call ApplyProof, return subgoals (no admission to context).

### Step 4: Add `Labeler` to `ast/labeler.go` (new file)

Port Python `ivy_ast.py:1938-1942`. `Labeler` generates unique label atoms:

```go
type Labeler struct { rn *iu.UniqueRenamer }
func NewLabeler() *Labeler
func (lb *Labeler) Call() *Atom  // returns Atom(rn.Rename(""), nil)
```

Requires `ivyutils` import — no circular dependency (ivyutils doesn't import ast).

### Step 5: Complete `CheckProperties` in `compiler/phase6.go:1745`

The function already has correct scaffolding (`props`, `pmap`, `nmap`, `namedTrans`). Changes needed:

**A. Create ProofChecker** (after `namedTrans`, before the loop):
```go
schemataTyped := make(map[string]*ast.LabeledFormula)
for k, v := range mod.Schemata {
    if lf, ok := v.(*ast.LabeledFormula); ok { schemataTyped[k] = lf }
}
prover := proof.NewProofChecker(mod.LabeledAxioms, mod.Definitions, schemataTyped)
```

**B. In `hasPf` branch** (replace lines 1806-1830): Call `prover.AdmitProposition(prop, pf.(ast.Node))`, then:
- If NOT definition: `prop = namedTrans(prop)`, update `prover.Axioms[-1]` and `prover.Schemata`
- If subgoals empty: route to definitions/axioms/schemata (existing logic)
- If subgoals non-empty: convert via `TheoremToProperty`, create labels via `Labeler`+`ComposeAtoms`, append to `mod.LabeledProps`; also route prop itself
- Always: `mod.Subgoals = append(...)`

**C. In no-proof branch** (replace lines 1831-1846): After existing logic, add missing `prover.AdmitProposition(nprop, &ast.ComposeTactics{})` / `prover.AdmitProposition(prop, &ast.ComposeTactics{})` calls.

**D. Pass prover to ApplyAssertProofs**: Change final call to `ApplyAssertProofsWithProver(mod, prover)`.

### Step 6: Update `ApplyAssertProofs` in `compiler/phase6.go`

Add `ApplyAssertProofsWithProver(mod, prover)` that passes the prover to `applyAssertProofAction`. Update `applyAssertProofAction` (lines 1733-1739) to use `prover.GetSubgoals` matching Python `apply_assert_proof` (ivy_compiler.py:1924-1941):
- Extract goal from AssertAction
- Call `prover.GetSubgoals(goal, pf)`
- Convert subgoals via `TheoremToProperty`
- Build `Sequence(SubgoalActions... + AssumeAction)`

### Step 7: Wire call site in `ivy_compile.go:167`

Replace `CheckPropertiesPass(mod)` → `if err := CheckProperties(mod); err != nil { return err }`.
Delete old `CheckPropertiesPass` (lines 1447-1491) since `CheckProperties` supersedes it.

## Files to Modify

| File | Action |
|------|--------|
| `proof/checker.go` | Add `AdmitDefinition`, `AdmitProposition`, `GetSubgoals` methods |
| `ast/labeler.go` | **NEW** — `Labeler` struct |
| `compiler/phase6.go` | Complete `CheckProperties` + update `ApplyAssertProofs`/`applyAssertProofAction` |
| `compiler/ivy_compile.go` | Replace call site, delete `CheckPropertiesPass` |

## Existing Utilities to Reuse

- `proof.NormalizeGoal` — `proof/goal.go`
- `proof.GoalVocab` — `proof/goal.go`
- `proof.GoalConc` — `proof/goal.go`
- `clauseops.SymbolsAST` — `clauseops/astutil.go`
- `clauseops.DropUniversals` — `clauseops/astutil.go`
- `ast.ComposeAtoms` — `ast/rewrite.go`
- `ast.ComposeTactics` — already exists as ast type
- `TheoremToProperty` — `compiler/phase6.go:1591`
- `ReorderProps` — `compiler/phase6.go:1528`
- `isSchemaBody` — `compiler/phase6.go:1852`
- `freshPropID` — `compiler/phase6.go:1871`
- `module.SubgoalEntry` — `module/module.go:149`

## Verification

1. `cd /Users/jaten/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `make test` — run existing tests (uses DYLD_LIBRARY_PATH for Z3)
3. Write a focused test in `proof/checker_test.go` for `AdmitProposition`:
   - Property with empty ComposeTactics proof → admitted to axioms
   - Definition → routed to AdmitDefinition
   - Nil proof → NoMatch error
4. Write a test in `compiler/phase6_test.go` for `CheckProperties`:
   - Module with labeled props + proofs → props become axioms
   - Module with temporal props → preserved in LabeledProps
   - Module with subgoals → TheoremToProperty conversion
