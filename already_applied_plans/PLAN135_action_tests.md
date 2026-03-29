# Add unit tests for compilation of RequiresAction, EnsuresAction, SubgoalAction, and LocalAction

**Created:** 2026-03-28

## Context

We just added `*ast.RequiresAction`, `*ast.EnsuresAction`, and `*ast.SubgoalAction` cases to `CompileActionBody` and created `CompileRequiresFormula`, `CompileEnsuresFormula`, `CompileSubgoalFormula`. These have zero test coverage. Existing tests cover `CompileAssertFormula` (2 tests) and `CompileLocal` (9 tests). We need parallel coverage for the new assert-like types, plus any LocalAction gaps.

## Test file

**New file:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/assert_like_test.go`

Follow existing patterns from `phase6_assert_proof_test.go` and `expr_test.go`:
- `newTestCompiler()` for compiler setup
- `ast.NewAstConfig()` for AST node creation
- Type assertions on compiled results

## Tests to add

### RequiresAction compilation (3 tests)

1. **TestCompileRequiresAction_NoProof** — Create `cfg.NewRequiresAction(formula)`, compile via `CompileActionBody`, verify result is `*actions.RequiresAction` with correct Formula and nil Proof.

2. **TestCompileRequiresAction_WithProof** — Create `cfg.NewRequiresAction(formula, proofTactic)`, compile, verify `*actions.RequiresAction` with non-nil Proof.

3. **TestCompileRequiresAction_WithLabeledFormula** — Create a `LabeledFormula` wrapping the formula, pass as arg to `cfg.NewRequiresAction(lf)`, compile, verify result has `.LF` set and `.Unprovable` propagated.

### EnsuresAction compilation (3 tests)

4. **TestCompileEnsuresAction_NoProof** — Same pattern as requires but with `cfg.NewEnsuresAction(formula)`, verify `*actions.EnsuresAction`.

5. **TestCompileEnsuresAction_WithProof** — With proof tactic, verify Proof field.

6. **TestCompileEnsuresAction_Unprovable** — Create with unprovable LabeledFormula, verify `Unprovable` field is true on compiled action.

### SubgoalAction compilation (2 tests)

7. **TestCompileSubgoalAction_NoProof** — Create `cfg.NewSubgoalAction(formula)`, compile, verify `*actions.SubgoalAction`.

8. **TestCompileSubgoalAction_WithProof** — With proof, verify Proof field.

### CompileNode dispatch (3 tests)

9. **TestCompileNode_RequiresAction** — Verify `CompileNode` (not `CompileActionBody`) correctly dispatches `*ast.RequiresAction` through to `CompileActionBody` (i.e., doesn't hit default/OtherThing case).

10. **TestCompileNode_EnsuresAction** — Same for EnsuresAction.

11. **TestCompileNode_SubgoalAction** — Same for SubgoalAction.

### LocalAction with assert-like body (2 tests)

12. **TestCompileLocal_WithRequiresBody** — LocalAction whose body is a RequiresAction, verify the compiled LocalAction.Body contains a compiled RequiresAction.

13. **TestCompileLocal_WithEnsuresBody** — Same for EnsuresAction in the body.

### Shared helper test (1 test)

14. **TestCompileAssertLikeFormula_FactoryPreservesType** — Call each of `CompileAssertFormula`, `CompileRequiresFormula`, `CompileEnsuresFormula`, `CompileSubgoalFormula` with the same formula, verify each returns the correct distinct action type.

## Test implementation pattern

```go
func TestCompileRequiresAction_NoProof(t *testing.T) {
    cfg := ast.NewAstConfig()
    c := newTestCompiler()

    formula := cfg.NewAtom("true")
    reqAction := cfg.NewRequiresAction(formula)

    act, err := c.CompileActionBody(reqAction)
    if err != nil {
        t.Fatalf("CompileActionBody failed: %v", err)
    }

    ra, ok := act.(*actions.RequiresAction)
    if !ok {
        t.Fatalf("expected *actions.RequiresAction, got %T", act)
    }
    if ra.Formula == nil {
        t.Fatal("expected Formula to be set")
    }
    if ra.Proof != nil {
        t.Errorf("expected Proof to be nil, got %v", ra.Proof)
    }
}
```

## Verification

1. `go test ./compiler/... -run 'TestCompile(Requires|Ensures|Subgoal|Node_|Local_With|AssertLike)'` — all new tests pass
2. `go test ./compiler/...` — no regressions
