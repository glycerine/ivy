# Plan: Port ivy_tactics.py Proof Tactics to goivy/tactics

## Context

Python's `ivy_tactics.py` defines 6 proof tactics (`vcgen`, `skolemize`, `skolemizenp`, `tempind`, `tempcase`, `sorry`) that are registered with the proof system and invoked from proof scripts. **None of these 6 tactics exist in Go yet.** The existing Go `tactics/tactics.go` contains interactive UPDR tactics (from `ivy_tactics_api.py`) and `proof/tactics.go` contains proof-level node tactics (let, assume, unfold, etc. from `ivy_proof.py`). The `ivy_tactics.py` tactics are a separate, missing layer.

These tactics are needed by `check/check.go`'s `MCTactic`/`VMTTactic` and `CheckTemporals` to do temporal property verification. Without `tempind` in particular, temporal proofs cannot proceed.

## What to Port

### Python `ivy_tactics.py` — Complete Inventory

| # | Function | Type | Lines | Purpose |
|---|----------|------|-------|---------|
| 1 | `vcgen(self, decls, proof)` | Tactic | 21-31 | Reduces safety property to initiation + consecution goals |
| 2 | `vc_to_goal(lineno, name, vc, action)` | Helper | 33-35 | Converts verification condition to goal |
| 3 | `triple_to_goal(lineno, name, action, precond, postcond)` | Helper | 37-39 | Converts Hoare triple to goal via `MakeVC` |
| 4 | `skolemize(self, decls, proof)` | Tactic | 43-46 | Skolemizes goal (prenex) |
| 5 | `skolemizenp(self, decls, proof)` | Tactic | 50-53 | Skolemizes goal (non-prenex) |
| 6 | `tempind_fmla(fmla, cond, params, vs)` | Helper | 57-71 | Recursive formula transform for temporal induction |
| 7 | `apply_tempind(goal, proof)` | Helper | 73-89 | Applies temporal induction to a goal |
| 8 | `tempind(self, decls, proof)` | Tactic | 91-94 | Temporal induction tactic entry point |
| 9 | `tempcase_fmla(fmla, cond, vs, proof)` | Helper | 98-108 | Recursive formula transform for temporal case |
| 10 | `apply_tempcase(goal, proof)` | Helper | 110-125 | Applies temporal case analysis to a goal |
| 11 | `tempcase(self, decls, proof)` | Tactic | 127-130 | Temporal case analysis tactic entry point |
| 12 | `sorry(self, decls, proof)` | Tactic | 136-139 | Drops goal, marks `used_sorry` |
| — | `used_sorry` | Global | 134 | Boolean flag: was sorry used? |

### Tactic Registrations
```
pf.register_tactic('vcgen', vcgen)
pf.register_tactic('skolemize', skolemize)
pf.register_tactic('skolemizenp', skolemizenp)
pf.register_tactic('tempind', tempind)
pf.register_tactic('tempcase', tempcase)
pf.register_tactic('sorry', sorry)
```

## Go Infrastructure Already Available

All key dependencies exist:

| Dependency | Go Location | Signature |
|-----------|-------------|-----------|
| `GoalConc` | `proof/goal.go:21` | `func GoalConc(g *ast.LabeledFormula) lg.Expr` |
| `GoalPrems` | `proof/goal.go:40` | `func GoalPrems(g *ast.LabeledFormula) []ast.Node` |
| `GoalVocab` | `proof/goal.go:146` | `func GoalVocab(goal *ast.LabeledFormula) *Vocab` |
| `CloneGoal` | `proof/goal.go:61` | `func CloneGoal(goal *ast.LabeledFormula, prems []ast.Node, conc lg.Expr) *ast.LabeledFormula` |
| `MakeGoal` | `proof/goal.go:74` | `func MakeGoal(lineno ast.Location, name string, prems []ast.Node, conc lg.Expr) *ast.LabeledFormula` |
| `SkolemizeGoal` | `proof/skolem.go:17` | `func SkolemizeGoal(goal *ast.LabeledFormula, prenex bool) *ast.LabeledFormula` |
| `CompileExprVocab` | `proof/phase5_matching.go:25` | `func CompileExprVocab(expr ast.Node, vocab *Vocab, mod *module.Module) lg.Expr` |
| `NormalizedAnd` | `ivylogic/constructors.go:324` | `func NormalizedAnd(args ...lg.Expr) lg.Expr` |
| `EnvAction` | `temporal/temporal.go:272` | `func EnvAction(bindings []*ActionTermBinding) *actions.EnvAction` |
| `MakeVC` | `trace/trace.go:699` | `func MakeVC(action, precond, postcond, checkAsserts)` |
| `lg.ForAll` | `logic/formula.go:367` | `func NewForAll(vars []*Variable, body Expr) (*ForAll, error)` |
| `lg.Globally` | `logic/formula.go:126` | `type Globally struct { Environ *string; Body Expr }` |
| `lg.WhenOperator` | `logic/formula.go:188` | `func NewWhenOperator(name string, t1, t2 Expr) (*WhenOperator, error)` |
| `lg.Implies` | `logic/formula.go:318` | `func NewImplies(t1, t2 Expr) (*Implies, error)` |
| `lg.Equals` | `logic/formula.go` | `func NewEquals(t1, t2 Expr) (*Equals, error)` |
| `lg.Not` | `logic/formula.go:102` | `func NewNot(body Expr) (*Not, error)` |
| `lg.Or` | `logic/formula.go:287` | `func NewOr(terms ...Expr) (*Or, error)` |
| `il.ForAll` | `ivylogic/util.go:228` | `func ForAll(vs []*lg.Variable, body lg.Expr) lg.Expr` (handles empty vs) |
| `lu.UsedVariables` | `logicutil/logicutil.go:19` | `func UsedVariables(t logic.Expr) map[NodeKey]Expr` |
| `lu.ClausesToFormula` | `logicutil/ or clauseops/` | Converts clauses to formula |
| Tactic type | `proof/checker.go:11` | `type Tactic func(*ProofChecker, []*ast.LabeledFormula, ast.Node) ([]*ast.LabeledFormula, error)` |
| `RegisterTactic` | `proof/checker.go` | `func (cfg *Config) RegisterTactic(name string, t Tactic)` |

### Missing AST type: `TacticLets`

Python's `TacticLets` is an AST node type (inherits from `Tactic`) whose `.args` are the let-binding expressions. Go's `ast/tactic.go` has a comment referencing it at line 243 (`Body Node // TacticWith or TacticLets`) but the type itself is not defined.

**Need to add:** `TacticLets` struct to `ast/tactic.go`:
```go
type TacticLets struct {
    Base
    Lets []Node
}
```

Then in `apply_tempind` / `apply_tempcase`, extract lets from `proof.Body` when it's a `*ast.TacticLets`.

## Implementation Plan

### File: `goivy/tactics/ivy_tactics.go` (new file)

Following the mechanical port rule (one Python file → one Go file with same base name), create `ivy_tactics.go` in the `tactics` package.

### Step 1: Package-level state and imports

```go
package tactics

import (
    "github.com/glycerine/goivy/ast"
    "github.com/glycerine/goivy/ivylogic"
    "github.com/glycerine/goivy/logic"
    "github.com/glycerine/goivy/logicutil"
    "github.com/glycerine/goivy/module"
    "github.com/glycerine/goivy/proof"
    "github.com/glycerine/goivy/temporal"
    "github.com/glycerine/goivy/trace"
    "github.com/glycerine/goivy/clauseops"
)

var UsedSorry bool  // Python: used_sorry = False
```

### Step 2: Helper functions

**`VcToGoal`** — wraps `proof.MakeGoal` + `lg.Not` + `clauseops.ClausesToFormula`
```
Python: pr.make_goal(lineno, name, [], lg.Not(lu.clauses_to_formula(vc)), annot=(action, vc.annot))
```

**`TripleToGoal`** — wraps `trace.MakeVC` + `VcToGoal`
```
Python: vc = tr.make_vc(action, precond, postcond); return vc_to_goal(lineno, name, vc, action)
```

**`TempindFmla`** — recursive formula transformation (most complex helper)
```
Python lines 57-71: Handles ForAll/Implies/Globally cases
Go: type-switch on *lg.ForAll, *lg.Implies, *lg.Globally
     Uses il.ForAll (handles empty vars), lg.NewImplies, lg.NewOr, lg.NewWhenOperator, lg.NewNot
```

**`ApplyTempind`** — orchestrates `TempindFmla` with tactic context
```
Python lines 73-89:
1. Check no tactic_decls
2. Get vocab with bound=True
3. Compile let expressions via CompileExprVocab
4. Build condition (NormalizedAnd of Equals)
5. Extract params (LHS of each let)
6. Check goal is temporal
7. Apply TempindFmla to conclusion
8. CloneGoal with new conclusion
```

**`TempcaseFmla`** — recursive formula transformation for case analysis
```
Python lines 98-108: Handles ForAll/Implies/Globally, checks variable capture
```

**`ApplyTempcase`** — orchestrates `TempcaseFmla`
```
Python lines 110-125: Same structure as ApplyTempind but for case analysis
```

### Step 3: Tactic functions

Each follows the pattern: extract first goal, transform, return `[transformed] + rest`

1. **`Vcgen`** — `(pc *proof.ProofChecker, decls []*ast.LabeledFormula, proofNode ast.Node) ([]*ast.LabeledFormula, error)`
   - Get goal conclusion via `proof.GoalConc`
   - Verify it's `*ast.TemporalModels` with `lg.IsTrue(conc.Fmla)`
   - Extract model, build initiation goal via `TripleToGoal(lineno, "initiation", model.Init, nil, model.Invars)`
   - Build consecution goal via `TripleToGoal(lineno, "consecution", temporal.EnvAction(model.Bindings), model.Invars+model.Asms, model.Invars)`

2. **`Skolemize`** — calls `proof.SkolemizeGoal(goal, true)`

3. **`Skolemizenp`** — calls `proof.SkolemizeGoal(goal, false)`

4. **`Tempind`** — calls `ApplyTempind(goal, proofNode)`

5. **`Tempcase`** — calls `ApplyTempcase(goal, proofNode)`

6. **`Sorry`** — sets `UsedSorry = true`, returns `decls[1:]`

### Step 4: Registration function

```go
func RegisterTactics(cfg *proof.Config) {
    cfg.RegisterTactic("vcgen", Vcgen)
    cfg.RegisterTactic("skolemize", Skolemize)
    cfg.RegisterTactic("skolemizenp", Skolemizenp)
    cfg.RegisterTactic("tempind", Tempind)
    cfg.RegisterTactic("tempcase", Tempcase)
    cfg.RegisterTactic("sorry", Sorry)
}
```

### Step 5: AST addition — `TacticLets` type

**File: `ast/tactic.go`** — Add after `LetTactic` (around line 165):

```go
// TacticLets holds the let-bindings for a tactic invocation.
// Corresponds to Python's TacticLets class.
type TacticLets struct {
    Base
    Lets []Node
}
func (t *TacticLets) Args() []Node           { return t.Lets }
func (t *TacticLets) Clone(args []Node) Node { return &TacticLets{Base: t.Base, Lets: args} }
func (t *TacticLets) String() string         { return fmt.Sprintf("TacticLets(%v)", t.Lets) }
```

Also add helper methods on `TacticTactic` to match Python's properties:
```go
func (t *TacticTactic) TacticDecls() []Node {
    if tw, ok := t.Body.(*TacticWith); ok { return tw.Args() }
    return nil
}
func (t *TacticTactic) TacticLetsList() []Node {
    if tl, ok := t.Body.(*TacticLets); ok { return tl.Lets }
    return nil
}
func (t *TacticTactic) TacticProof() Node {
    if t.Proof != nil { if _, ok := t.Proof.(*NoneAST); !ok { return t.Proof } }
    return nil
}
```

**Also need `TacticWith` type if not yet defined:**
```go
type TacticWith struct {
    Base
    Elems []Node
}
```

### Step 6: Wire registration into check package

**File: `check/check.go`** — In `RegisterTactics` (line 1010), add:
```go
tactics.RegisterTactics(proofCfg)
```

### Step 7: Extract `tactic_lets` in proof/checker.go

The `tacticTactic` dispatcher at `proof/checker.go:515` passes the `*ast.TacticTactic` directly to registered tactics. The tactic functions receive it as `ast.Node` and need to type-assert to `*ast.TacticTactic` to get `.TacticLetsList()`.

## Key Translation Patterns

### `lg.is_forall(fmla)` → Go type switch
```go
if fa, ok := fmla.(*lg.ForAll); ok { ... fa.Variables, fa.Body ... }
```

### `isinstance(fmla, lg.Implies)` → Go type switch
```go
if imp, ok := fmla.(*lg.Implies); ok { ... imp.T1, imp.T2 ... }
```

### `isinstance(fmla, lg.Globally)` → Go type switch
```go
if gb, ok := fmla.(*lg.Globally); ok { ... gb.Body ... }
```

### `lg.ForAll(vs, body)` → Go `il.ForAll(vs, body)`
The `ivylogic.ForAll` handles empty `vs` by returning `body` directly — matching Python's `lg.forall`.

### `ilu.variables_ast(expr)` → Go `lu.UsedVariables(expr)`
Returns all variables in expression. Need to extract `*lg.Variable` from the result map.

### `fmla.clone([new_body])` → Go `fmla.Clone([]ast.Node{newBody})`
AST nodes have Clone methods.

### `proof.tactic_lets` → Go `proofNode.(*ast.TacticTactic).TacticLetsList()`

### `proof.tactic_decls` → Go `proofNode.(*ast.TacticTactic).TacticDecls()`

### `proof.lineno` → Go `ast.GetLocation(proofNode)` or similar

## Files to Modify/Create

| File | Action |
|------|--------|
| `tactics/ivy_tactics.go` | **CREATE** — all 6 tactics + 5 helpers + registration |
| `tactics/ivy_tactics_test.go` | **CREATE** — tests for each tactic |
| `ast/tactic.go` | **EDIT** — add `TacticLets`, `TacticWith` types + helper methods on `TacticTactic` |
| `check/check.go` | **EDIT** — wire `tactics.RegisterTactics` into `RegisterTactics` function |

## Dependency on `NormalProgram` Fields

The `vcgen` tactic needs these fields from `temporal.NormalProgram`:
- `.Init` (action) — exists at `temporal/temporal.go:104`
- `.Invars` (labeled formulas) — exists at `temporal/temporal.go:106`
- `.Asms` (labeled formulas) — exists at `temporal/temporal.go:108`
- `.Bindings` (action term bindings) — exists at `temporal/temporal.go:102`

`EnvAction` exists at `temporal/temporal.go:272`.

## Verification

1. `go build ./tactics/...` — compiles
2. `go build ./ast/...` — compiles with new types
3. `go build ./check/...` — compiles with new registration
4. `go test ./tactics/...` — new tests pass
5. `go test ./proof/...` — existing proof tests still pass
6. `go test ./check/...` — existing check tests still pass
7. Integration: A `.ivy` file using `proof tempind` or `proof sorry` should dispatch to the new tactics
