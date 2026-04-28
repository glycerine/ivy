# DIV-8 Investigation: CloseEPR Variable Ordering

Created: 2026-04-28 ~UTC

## Finding: DIV-8 Is Not a Real Bug

**Python `ForAll._preprocess_`** (logic.py:380) always sorts variables:
```python
return tuple(sorted(set(variables), key=lambda v: v.name)), body
```

This means EVERY ForAll in Python has variables sorted alphabetically by name, regardless of the input order from `close_epr` or anywhere else.

**Go `CloseEPR`** uses `FreeVariablesList` which returns variables sorted by `NodeKey` (a string like `(Variable name:A sort:(UninterpretedSort name:S))`). Since the name field comes first in the string, this produces the same alphabetical-by-name ordering.

**Experimental evidence**: tested 6+ variable pairs including different-sort same-name variables. Go and Python always agree on CloseEPR output ordering.

**Conclusion**: Change the test from "detect ordering divergence" to "confirm ordering matches" — and mark it as verified-non-bug.

## Action

Update `TestDIV8_CloseEPRVariableOrdering` to use the 3-variable mixed-sort case (which exercises the most complex ordering scenario) and assert that Go and Python agree, documenting WHY they agree (ForAll sorts, FreeVariablesList sorts).

## Files to Create

1. **`~/ivy/goivy/pytesthelper/div_conformance.py`** — Python driver script
2. **`~/ivy/goivy/logicutil/div_conformance_test.go`** — Tests for DIV-2, 4, 5, 7, 8
3. **`~/ivy/goivy/actions/div_conformance_test.go`** — Tests for DIV-9, 10, 11, 12, 14

## Test Design per DIV

### DIV-5: `Some` binder not excluded from FreeVariables

**Go test** (`logicutil/div_conformance_test.go`):
```go
// Create: Some(X, P(X,Y)) — X is bound, Y is free
// Call FreeVariablesList or VariablesAstList
// CORRECT: X not in result, Y in result
// BUG: X leaks through because no case for *ivylogic.Some
```
- Construct `ivylogic.Some{Params: [X], Fmla: Apply(P, X, Y)}`
- Call `logicutil.VariablesAstList(some)` or `FreeVariablesList`
- Assert X is NOT in result (this will FAIL now, proving bug)

**Python confirms**: `variables_ast(Some(X, P(X,Y)))` yields only Y

### DIV-9: WhileAction.Expand missing SubgoalAction filtering

**Go test** (`actions/div_conformance_test.go`):
```go
// Create WhileAction with invariants including a SubgoalAction formula
// Expand it
// CORRECT: assumes list should NOT contain SubgoalAction-sourced entries
// BUG: Go puts all invariants into assumes unconditionally
```
- Create `WhileAction` with `Invariants: [formula1, formula2]` where formula2 comes from SubgoalAction
- Call `Expand(ctx)`
- Extract the assumes from the expanded sequence
- Assert SubgoalAction-sourced invariants are excluded from assumes

Note: This is tricky because Go's Expand doesn't have access to SubgoalAction type info at that point — it only sees `lg.Expr` invariants. The Python version has the SubgoalAction *objects* as invariants. The real fix is to change how invariants are stored (as Actions, not expressions). For the test, we construct a WhileAction where some invariants ARE SubgoalActions (cast to lg.Expr via interface).

### DIV-11: ApplyMixin returns instead of erroring

**Go test** (`actions/div_conformance_test.go`):
```go
// Create two actions with different param counts
// Call ApplyMixin
// CORRECT: should return error (Python raises IvyError)
// BUG: Go returns action2 unchanged, no error
```
- Create `action1` with 2 formal params, `action2` with 3 formal params
- Call `ApplyMixin(action1, action2, false)`
- Assert error is returned (currently returns action2 with no error, proving bug)

Since ApplyMixin returns Action (not error), the test checks that the result is NOT just action2 unchanged. The fix will change the signature to return `(Action, error)`.

### DIV-12: InstantiateAction missing from IntUpdate dispatch

**Go test** (`actions/div_conformance_test.go`):
```go
// Create InstantiateAction
// Call IntUpdate(instAction, ctx)
// CORRECT: should invoke InstantiateAction.IntUpdate method
// BUG: hits default case, returns NullUpdate
```
- Create `NewInstantiateAction(someExpr)`
- Call `IntUpdate(instAction, ctx)`
- Assert result is NOT NullUpdate (this FAILS now, proving bug)

### DIV-2: SubstituteByName Lambda blocks substitution

**Go test** (`logicutil/div_conformance_test.go`):
```go
// Create: Lambda([X], Eq(X, Y)), substitute {"X": Z}
// CORRECT (Python): substitution flows into Lambda body → Lambda([X], Eq(Z, Y))
// BUG (Go): Lambda removes "X" from subs → Lambda([X], Eq(X, Y)) unchanged
```
- Create Lambda with variable X in body
- Substitute X→Z
- Assert body contains Z (this FAILS now because Go protects Lambda vars)

**Python confirms**: `substitute_ast(Lambda([X], Eq(X,Y)), {"X": Z})` → body has Z

### DIV-8: CloseEPR variable ordering

**Go test** (`logicutil/div_conformance_test.go`):
```go
// Create formula with free variables Y, X (Y appears first in DFS)
// Call CloseEPR
// Go: ForAll([Y, X], ...) — DFS order preserved via Omap
// Python: ForAll with set-order (implementation-dependent)
```
- Create `Eq(Y, X)` where Y is encountered before X
- Call `CloseEPR(eq)`
- Check variable order in result
- Compare with Python output

Note: Python uses `set` so ordering is implementation-dependent. The test verifies that Go produces DFS order (which may accidentally match Python's CPython dict ordering for small sets, or may not). The key test is just documenting the behavioral difference.

### DIV-10: WhileAction havocs missing lineno

**Go test** (`actions/div_conformance_test.go`):
```go
// Create WhileAction with a lineno set
// Expand it
// CORRECT: havocs should inherit while's lineno
// BUG: havocs have no lineno
```
- Create WhileAction, set its lineno
- Expand
- Find HavocActions in expanded sequence
- Assert each has lineno matching the while (FAILS now)

### DIV-14: checked_assert ignores file component

**Go test** (`actions/div_conformance_test.go`):
```go
// Create AssertAction at file="foo.ivy" line=42
// Set ctx.CheckedAssert to match Python's Location("bar.ivy", 42)
// CORRECT: should NOT match (different file)
// BUG: Go compares "42" == "42", matches incorrectly
```
- Create AssertAction with lineno Location("foo.ivy", 42)
- Create ctx with CheckedAssert = "42" (current Go format)
- Call ActionUpdate
- In Python, "bar:42" processed as Location("bar.ivy", 42) ≠ Location("foo.ivy", 42)
- Test: two asserts same line, different file — Go treats them as same

### DIV-4: SubstituteApply missing free-variable assertion

**Go test** (`logicutil/div_conformance_test.go`):
```go
// Create a substitution function that introduces a new free variable
// Call SubstituteApply
// CORRECT (Python): assertion fires, error raised
// BUG (Go): no check, returns result with extra free variables
```
- Create Apply(f, X) with subs[f] = λterms → Apply(g, terms[0], NEW_VAR)
- Call SubstituteApply
- Check free variables of result vs free variables of input terms
- Assert fv(result) ⊆ fv(input_terms) — this FAILS now

### DIV-7: NormalizeQuantifiers accepts Lambda (Python crashes)

**Go test** (`logicutil/div_conformance_test.go`):
```go
// Pass a Lambda to NormalizeQuantifiers
// CORRECT (Python): assert False — should never happen / panic
// BUG (Go): returns Lambda unchanged, no error
```
- Create Lambda([X], body)
- Call NormalizeQuantifiers(lambda)
- Assert panic or error (Go doesn't panic, proving divergence)
- Note: since Python asserts False, this is a "should never reach here" guard.
  The test documents that Go silently accepts what Python rejects.

## Python Driver (`pytesthelper/div_conformance.py`)

Structure:
```python
#!/usr/bin/env python3
import sys, os, json

sys.path.insert(0, os.path.expanduser('~/ivy/pyivy/ivy'))

from ivy import logic as lg
from ivy import ivy_logic as ilg  
from ivy import ivy_logic_utils as ilu
from ivy import logic_util as lu

def div2():
    """SubstituteByName: Lambda does NOT block substitution"""
    S = lg.UninterpretedSort("S")
    X = lg.Variable("X", S)
    Y = lg.Variable("Y", S)  
    Z = lg.Variable("Z", S)
    body = lg.Eq(X, Y)
    lam = lg.Lambda([X], body)
    result = ilu.substitute_ast(lam, {"X": Z})
    # Check if body was substituted
    return "substituted" if result.body.args[0] == Z else "not_substituted"

def div4():
    """SubstituteApply: asserts no new free variables"""
    # Construct case where substitution adds a free variable
    try:
        # ... exercise the assertion
        return "no_error"  
    except AssertionError as e:
        return "assertion_fired"

def div5():
    """variables_ast: Some excludes bound variables"""
    S = lg.UninterpretedSort("S")
    X = lg.Variable("X", S)
    Y = lg.Variable("Y", S)
    from ivy.ivy_logic import Some
    some = Some(X, lg.Eq(X, Y))
    fvs = set(ilu.variables_ast(some))
    return json.dumps({"X_free": X in fvs, "Y_free": Y in fvs})

# ... similar for div7, div8, div9, div10, div11, div12, div14

if __name__ == "__main__":
    func = globals()[sys.argv[1]]
    print(func())
```

Go tests call: `exec.Command("python3", pyScript, "div2")`

## Implementation Order

1. Create `pytesthelper/div_conformance.py` with all Python functions
2. Create `logicutil/div_conformance_test.go` with DIV-2, 4, 5, 7, 8
3. Create `actions/div_conformance_test.go` with DIV-9, 10, 11, 12, 14
4. Run `cd ~/ivy/goivy && make test` to confirm tests FAIL (proving bugs exist)

## Key Patterns to Reuse

- `logic/sexp_cross_test.go` — exec.Command pattern for Python
- `logic/sexp_cross_test.go:buildAllGoSexps` — constructing logic.Expr test objects
- `actions/update_test.go` — UpdateContext creation, Action construction
- `testVectorsDir()` for locating Python scripts

## Critical Files Referenced

- `logicutil/logic_utils.go:627-665` — SubstituteByName (DIV-2)
- `logicutil/logic_utils.go:1740-1758` — SubstituteApply (DIV-4)
- `logicutil/logicutil.go:641-678` — variablesAstRec (DIV-5)
- `logicutil/logic_utils.go:1648-1718` — NormalizeQuantifiers (DIV-7)
- `logicutil/logic_utils.go:430-444` — CloseEPR (DIV-8)
- `actions/update.go:1666-1704` — WhileAction.Expand (DIV-9, 10)
- `actions/helpers.go:143-178` — ApplyMixin (DIV-11)
- `actions/update.go:1170-1244` — IntUpdate dispatch (DIV-12)
- `actions/update.go:374-389` — checked_assert (DIV-14)
- `ivylogic/formula.go:11-25` — Some type (DIV-5)

## Verification

```
cd ~/ivy/goivy && make test
```

All 10 new tests should FAIL, confirming the divergences are real. After fixing each DIV in Go, the corresponding test should PASS.
