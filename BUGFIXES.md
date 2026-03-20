# Plan: Bug Fixes & Python Conformance for CheckProperties Implementation

## Context

Code review of the recently implemented CheckProperties pass (proof/checker.go, compiler/phase6.go, ast/labeler.go, proof/register.go) revealed several bugs and Python conformance issues. This plan addresses them in priority order.

## Bugs Found

### BUG 1 (HIGH): `NewProofChecker` uses wrong key for Definitions map
**File**: `proof/checker.go:59`
**Python** (ivy_proof.py:53): `self.definitions = dict((d.formula.defines().name, normalize_goal(d)) for d in definitions)`
**Go**: `name := d.LabelName()` — uses the *label* name, not the *defined symbol* name.
**Impact**: Definition lookup by symbol name will fail; `AdmitDefinition` stores by `symSym.Name` but constructor stored by label name. Inconsistent keys.

**Fix**: Change line 59 from `d.LabelName()` to extract the defined symbol name:
```go
name := ""
if def, ok := d.Formula.(*lg.Definition); ok {
    if sym, ok := def.Defines().(*lg.Symbol); ok {
        name = sym.Name
    }
}
if name == "" {
    name = d.LabelName() // fallback
}
pc.Definitions[name] = norm
```

### BUG 2 (HIGH): `proof/register.go` init() may not run — silent fallback to nil prover
**File**: `compiler/phase6.go:1916`, `proof/register.go`
**Problem**: `compiler` cannot import `proof` (cycle: `compiler` → `proof` → `compiler` via `phase5_matching.go`). The proof's `init()` registers `NewProofCheckerFn` via the `proof` → `compiler` direction. But `proof` is only imported by `check`, `ranking`, `tactics`, `temporal`, `l2s`, `iupdr` — NOT by `compiler` or its callers. The `end2end` tests work because they import `check`, but `compiler` tests and any direct user of the `compiler` package get nil `NewProofCheckerFn` and silent no-op proof checking.

**Fix**: Eliminate the factory-function indirection entirely. Move `register.go` from `proof/` to `check/` (which already imports both `proof` and `compiler`). Then add `_ "github.com/glycerine/goivy/check"` as a blank import in `compiler/ivy_compile.go`.

Wait — `check` imports `compiler`, and `compiler` can't import `check` (cycle). So blank-importing `check` from `compiler` won't work either.

**Actual fix — break the cycle properly**: The root cause is that `proof` imports `compiler` via `phase5_matching.go`. The real solution is to move the `ProofCheckerInterface`, `NewProofCheckerFn`, and `GoalConcFn` into the `module` package (which both `proof` and `compiler` already import). Then:
1. `module/proofapi.go` — defines the interface + factory vars
2. `proof/register.go` — sets `module.NewProofCheckerFn` (proof imports module ✓)
3. `compiler/phase6.go` — reads `module.NewProofCheckerFn` (compiler imports module ✓)
4. No new import edges needed — `proof` already imports `module`, `compiler` already imports `module`

This guarantees that `proof.init()` runs whenever `proof` is imported (which happens through `check`, `end2end` tests, etc.). For `compiler` tests in isolation, the factory stays nil and the nil guard still operates (which is actually fine — compiler-only tests don't need proof checking).

### BUG 3 (HIGH): Type assertion panic on `prop.Label.(*ast.Atom)` in CheckProperties
**File**: `compiler/phase6.go:1975`
**Python** (ivy_compiler.py:2025): `label = ia.compose_atoms(prop.label, lb())`
**Go**: `label := ast.ComposeAtoms(prop.Label.(*ast.Atom), lb.Call())`
**Problem**: If `prop.Label` is not an `*ast.Atom` (e.g. it's a `*lg.Symbol` or other Node type), Go panics. Python's `compose_atoms` is more flexible.
**Fix**: Add safe type assertion:
```go
labelAtom, ok := prop.Label.(*ast.Atom)
if !ok {
    return fmt.Errorf("property label is not an Atom: %T", prop.Label)
}
label := ast.ComposeAtoms(labelAtom, lb.Call())
```

### BUG 4 (MEDIUM): `AdmitProposition` missing `subgoals` parameter
**File**: `proof/checker.go:368`
**Python** (ivy_proof.py:98): `def admit_proposition(self, prop, proof=None, subgoals=None)`
**Go**: `func (pc *ProofChecker) AdmitProposition(prop *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)`
**Problem**: Python accepts optional pre-computed `subgoals`. Line 113: `subgoals = subgoals or [prop]`. Go always starts with `[prop]`.
**Impact**: No current callers pass subgoals, so not broken *yet*, but doesn't match the Python API.
**Fix**: Add optional parameter via variadic:
```go
func (pc *ProofChecker) AdmitProposition(prop *ast.LabeledFormula, proof ast.Node, existingSubgoals ...*ast.LabeledFormula) ([]*ast.LabeledFormula, error)
```
Then: `subgoals := existingSubgoals; if len(subgoals) == 0 { subgoals = []*ast.LabeledFormula{prop} }`
**Note**: This also requires updating the `ProofCheckerInterface` in `compiler/phase6.go`.

### BUG 5 (MEDIUM): `GetSubgoals` missing definition assertion
**File**: `proof/checker.go:397`
**Python** (ivy_proof.py:129): `assert not isinstance(prop.formula, il.Definition)`
**Go**: No equivalent check.
**Fix**: Add:
```go
if _, isDef := prop.Formula.(*lg.Definition); isDef {
    return nil, &ProofError{Msg: "GetSubgoals: prop may not be a definition"}
}
```

### BUG 6 (MEDIUM): Errors from `AdmitProposition` silently ignored in no-proof branch
**File**: `compiler/phase6.go:2011,2016`
**Go code**:
```go
prover.AdmitProposition(nprop, &ast.ComposeTactics{})  // return values ignored
```
**Python**: Also doesn't check, but Go convention is to handle errors.
**Fix**: Log errors (matching the pattern used in the has-proof branch):
```go
if _, err := prover.AdmitProposition(nprop, &ast.ComposeTactics{}); err != nil {
    pp("check_properties: admit error for %s: %v", nprop.LabelName(), err)
}
```

### BUG 7 (LOW): `applyAssertProofAction` — `sga.Kind = a.Kind` doesn't match Python's `sga.kind = type(self)`
**File**: `compiler/phase6.go:1811`
**Python**: Stores the Python *class* (`AssertAction`, `RequiresAction`, etc.) on the SubgoalAction.
**Go**: Copies the string `Kind` field. But `assert_to_assume` in Go uses Go type switching, not the Kind string. So the Kind string is informational only.
**Impact**: Low — downstream `AssertToAssume` uses Go type assertions, not `Kind` strings. However, for strict conformance, we should set `sga.SubgoalKind` to the Go type name:
```go
sga.SubgoalKind = a.Name()  // "assert", "require", "ensure"
```

## Files to Modify

| File | Change |
|------|--------|
| `proof/checker.go:59` | BUG 1: Fix Definitions key to use `defines().Name` instead of `LabelName()` |
| `proof/checker.go:368` | BUG 4: Add optional `subgoals` variadic parameter to `AdmitProposition` |
| `proof/checker.go:397` | BUG 5: Add definition assertion to `GetSubgoals` |
| `module/proofapi.go` | BUG 2: **NEW** — move `ProofCheckerInterface`, `NewProofCheckerFn`, `GoalConcFn` here |
| `proof/register.go` | BUG 2: Change to set `module.NewProofCheckerFn` instead of `compiler.NewProofCheckerFn` |
| `compiler/phase6.go` | BUG 2: Read from `module.NewProofCheckerFn`; remove local interface+vars |
| `compiler/phase6.go:1975` | BUG 3: Safe type assertion on `prop.Label` before `ComposeAtoms` |
| `compiler/phase6.go:2011,2016` | BUG 6: Log errors from `AdmitProposition` in no-proof branch |
| `compiler/phase6.go:1811` | BUG 7: Set `sga.SubgoalKind` to action name |
| `proof/checker_test.go` | Add tests for bugs 1, 4, 5 |

## Verification

1. `cd /Users/jaten/go/src/github.com/glycerine/goivy && go build ./...`
2. `make test`
3. Add test in `proof/checker_test.go` for:
   - Definition keyed by defines-name (not label name)
   - `GetSubgoals` rejects definitions
   - `AdmitProposition` with pre-supplied subgoals
4. Verify no panics by running end2end tests that exercise properties with subgoals
