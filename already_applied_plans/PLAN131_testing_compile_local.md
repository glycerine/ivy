# Comprehensive Unit Tests for CompileLocal
Created: 2026-03-28

## Context

The CompileLocal special case was deeply wrong (checking body for `:=` instead of localDecls for AssignAction). This went undetected because:
1. Only 1 unit test existed (`TestExpr6_CompileLocal_AssignmentSortInference`)
2. That test used hand-built AST that matched the OLD wrong structure, not what `LowerVarStatements` actually produces
3. No tests exercise the `AssignAction`-in-localDecls path that the LALR parser produces

Tests will be written to match CORRECT behavior. Some may fail initially due to nil pointer issues seen at golden line 142255 — that's expected and will drive implementation fixes.

## Existing test to fix

`compiler/expr_test.go:122` — `TestExpr6_CompileLocal_AssignmentSortInference` uses `body = Atom(":=", x, y)` with `localDecls = [Atom("x")]`. Must be rewritten to use `localDecls = [AssignAction(x, y)]` with `body = Sequence(...)`, matching what `LowerVarStatements` produces.

## Test cases to write

All tests go in `compiler/expr_test.go` (existing file, matching the pattern there).

### Test 1: Fix existing — Single AssignAction with sort inference
**Rewrites `TestExpr6_CompileLocal_AssignmentSortInference`**

Setup:
- `natSort` registered in sig, symbol `y` of sort `nat`
- `lhsNode = cfg.NewAtom("loc:x")` (prefixed, no ASort)
- `rhsNode = cfg.NewAtom("y")`
- `localDecls = [cfg.NewAssignAction(lhsNode, rhsNode)]`
- `body = cfg.NewSequence()` (empty continuation)

Assert:
- No error
- Result is `*actions.LocalAction`
- Local variable sort is `nat` (inferred from y)
- Body contains the assignment prepended to continuation

### Test 2: Single AssignAction with explicit sort annotation on LHS
Setup:
- `natSort` in sig, symbol `y` of sort `nat`
- `lhsNode = cfg.NewAtom("loc:x")` with `ASort = &ast.Symbol{Rep: "nat"}`
- `localDecls = [cfg.NewAssignAction(lhsNode, rhsNode)]`
- `body = cfg.NewSequence()`

Assert:
- No error
- Local variable sort is `nat` (from explicit annotation)

### Test 3: Single AssignAction with function-like LHS (has args)
Tests the `App.Func` extraction path (localVar from Apply).

Setup:
- Sort `proc` registered
- `lhsNode = cfg.NewApp(&ast.Symbol{Rep: "loc:f"}, cfg.NewAtom("P"))` — function with arg
- `rhsNode = cfg.NewAtom("true")`
- `localDecls = [cfg.NewAssignAction(lhsNode, rhsNode)]`
- `body = cfg.NewSequence()`

Assert:
- No error
- Result is `*actions.LocalAction`
- LocalAction's local is the function symbol, not the full Apply

### Test 4: Single bare declaration (NOT AssignAction) — generic path
Tests that the generic case (non-AssignAction in localDecls) still works.

Setup:
- `natSort` in sig
- `localDecls = [cfg.NewAtom("loc:x")]` (bare Atom, no AssignAction)
- `body = cfg.NewAtom("true")` (simple body)

Assert:
- No error
- Result is `*actions.LocalAction`
- Symbol is registered as a local

### Test 5: Multiple declarations — generic path
Setup:
- `natSort` in sig
- `localDecls = [cfg.NewAtom("loc:x"), cfg.NewAtom("loc:y")]`
- `body = cfg.NewAtom("true")`

Assert:
- No error
- Result is `*actions.LocalAction` with 2 locals

### Test 6: AssignAction with body containing multiple statements
Tests the body flattening logic (Sequence → lines).

Setup:
- `natSort` in sig, symbols `y`, `z` of sort `nat`
- `lhsNode = cfg.NewAtom("loc:x")`
- `rhsNode = cfg.NewAtom("y")`
- `localDecls = [cfg.NewAssignAction(lhsNode, rhsNode)]`
- `body = cfg.NewSequence(stmt1, stmt2)` — two statements

Assert:
- No error
- Result is `*actions.LocalAction`
- Body is Sequence with [assignment, stmt1, stmt2] (assignment prepended)

### Test 7: AssignAction where ensureSortAnnotation fires
Tests that when LHS has nil ASort, it gets set to "S".

Setup:
- Only `TopS` sort available (no named sorts)
- `lhsNode = cfg.NewAtom("loc:x")` with `ASort == nil`
- `rhsNode = cfg.NewAtom("true")`
- `localDecls = [cfg.NewAssignAction(lhsNode, rhsNode)]`
- `body = cfg.NewSequence()`

Assert:
- No error (ensureSortAnnotation prevents CompileConst from failing on nil sort)
- Result is `*actions.LocalAction`

### Test 8: LowerVarStatements integration — round-trip
Tests the full pipeline: `VarAction` → `LowerVarStatements` → `CompileLocal`.

Setup:
- `natSort` in sig, symbol `y` of sort `nat`
- Create `VarAction(Atom("x"), Atom("y"))` and a continuation statement
- Call `ast.LowerVarStatements(stmts)` to produce LocalAction
- Extract `localDecls` and `body` from the LocalAction
- Call `c.CompileLocal(localDecls, body)`

Assert:
- LowerVarStatements produces `LocalAction(AssignAction(...), Sequence(...))`
- CompileLocal succeeds
- Sort inference works on the result

### Test 9: ensureSortAnnotation on App node
Tests that App nodes also get default sort.

Setup:
- `lhsNode = cfg.NewApp(&ast.Symbol{Rep: "loc:f"}, cfg.NewAtom("P"))` with `ASort == nil`

Assert (unit test of ensureSortAnnotation):
- After `ensureSortAnnotation(lhsNode)`, `lhsNode.(*ast.App).ASort` is `&ast.Symbol{Rep: "S"}`

### Test 10: Symbol shadowing
Tests that a local variable correctly shadows an outer-scope symbol.

Setup:
- `natSort` in sig, symbol `x` of sort `nat` already in sig (outer scope)
- `lhsNode = cfg.NewAtom("loc:x")`
- `rhsNode = cfg.NewAtom("true")`
- `localDecls = [cfg.NewAssignAction(lhsNode, rhsNode)]`
- `body = cfg.NewSequence()`

Assert:
- No error
- The local's sort may differ from the outer `x` (it's a new scope)

## Files to modify
- `compiler/expr_test.go` — rewrite existing test + add 9 new tests

## Test naming convention
Following existing pattern: `TestExpr6_CompileLocal_<Scenario>`

## Expected initial state
Some tests may initially fail or panic due to the nil pointer issue at `SubstituteConstantsAst2` (seen at golden line 142255). Per user instruction: **do not weaken tests** — use failures to drive implementation fixes.

## Verification
```
cd ~/goivy && go test -run TestExpr6_CompileLocal ./compiler/ -v -count=1
```
