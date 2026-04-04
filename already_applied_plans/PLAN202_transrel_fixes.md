# PLAN202: Actions Package Audit — Bug Fixes & Conformance Tests

**Created:** 2026-04-05 ~02:00 UTC

## Context

Comprehensive audit of `~/ivy/goivy/actions/` against Python source of truth
(`ivy_actions.py` and `ivy_transrel.py`). The Go port is largely faithful —
the update.go action semantics, transrel.go transition relation logic, and
transforms.go are all excellent ports. However, we found two confirmed bugs
and several test gaps that need addressing.

## Bug 1: NativeAction.Impure never set (CRITICAL)

**Python** (`ivy_actions.py:1270-1277`):
```python
class NativeAction(Action):
    def __init__(self, *args):
        self.args = list(args)
        self.impure = False
        if isinstance(args[0].code, str) and args[0].code.split('\n')[0].strip() == "impure":
            args[0].code = '\n'.join(args[0].code.split('\n')[1:])
            self.impure = True
```

Python parses the first line of the native code template. If it is `"impure"`,
the flag is set and the line is stripped from the code.

**Go**: `NativeAction.Impure` field exists (`actions/action.go:985`) and is
checked in `isolate/phase7.go:115` (`HasSideEffectRec`), but **no code ever
sets it to `true`**:
- `NewNativeAction` (`action.go:988`) doesn't parse the code
- `CompileNativeDef` (`compiler/phase6.go:1001`) doesn't check for impure
- `ActionClone` preserves the field (`action.go:999`), but copies `false`

**Impact**: `HasSideEffectRec` in `isolate/phase7.go` will never detect impure
native actions, potentially causing incorrect isolation decisions.

### Fix

In `compiler/phase6.go`, after creating the `actions.NativeAction` at line 1001,
add impure parsing logic. The code template is stored as the name of the
`*lg.Const` at `compiled[0]` (see line 997: `compiled[0] = lg.NewConst(codeNode.Code, lg.TopS)`).

```go
// After line 1001: act := actions.NewNativeAction(compiled[0], compiled[1:]...)
// Parse "impure" keyword from first line of code template (Python: NativeAction.__init__)
if codeConst, ok := compiled[0].(*lg.Const); ok {
    lines := strings.SplitN(codeConst.Name, "\n", 2)
    if len(lines) > 0 && strings.TrimSpace(lines[0]) == "impure" {
        act.Impure = true
        // Strip the "impure" line from the code template
        newCode := ""
        if len(lines) > 1 {
            newCode = lines[1]
        }
        compiled[0] = lg.NewConst(newCode, lg.TopS)
        act.Code = compiled[0]
    }
}
```

**Files**: `compiler/phase6.go` (around line 1001)

## Bug 2: solver.isSkolem test gap

The `solver/solver.go:462` `isSkolem` function was already fixed to use
`strings.Contains(name, "__")`, matching Python. However, the test at
`solver/solver_test.go:930-948` only tests prefix cases (`"__x"`, `"_x"`,
`"x"`, `""`, `"_"`). It does NOT test the critical mid-string case:
- `"foo__bar"` should be skolem (contains `__`)
- `"a__"` should be skolem (trailing `__`)

### Fix

Add test cases to `TestIsSkolem` for `strings.Contains` semantics:
```go
// Mid-string __
if !isSkolem("foo__bar") {
    t.Fatal("foo__bar should be skolem (contains __)")
}
if !isSkolem("a__") {
    t.Fatal("a__ should be skolem (trailing __)")
}
if !isSkolem("__") {
    t.Fatal("__ alone should be skolem")
}
```

**Files**: `solver/solver_test.go`

## Confirmed Non-Bugs (Verified Correct)

The following were investigated and confirmed correct:

1. **IsSkolem** (`actions/transrel.go:73`) — uses `strings.Contains(name, "__")` ✓
2. **IsGlobalSkolem** (`actions/transrel.go:81`) — correct length + prefix + uppercase check ✓
3. **ComposeUpdates** (`actions/transrel.go:599`) — faithful port of `compose_updates` ✓
4. **All action update methods** (`actions/update.go`) — AssumeAction, AssertAction,
   AssignAction (including hierarchy, destructor, variant), SetAction, HavocAction,
   Sequence, ChoiceAction (with determinize), IfAction (with Some), WhileAction
   (with expand/unroll), LocalAction, LetAction, CrashAction, CallAction
   (with apply_actuals capture avoidance), InstantiateAction, BindOldsAction,
   NativeAction — all verified correct ✓
5. **Implies** (`actions/phase4.go:127`) — exists and matches Python logic ✓
6. **Interpolant family** (`actions/interpolant.go`) — exists ✓
7. **Modifies** (`actions/transforms.go:133`) — uses `modifiesRec` with destructor
   chain walk, CrashAction recursion ✓
8. **Decompose** (`actions/action.go:1320-1490`) — both simple `Decompose()` and
   stateful `DecomposeWithState()` exist ✓
9. **Global state** — counters, determinize, checked_assert all properly threaded
   through Config/UpdateContext ✓
10. **Schema** — split into ast.Schema (AST-level) and actions.Schema (compiled) ✓

## Comprehensive Unit Tests

Add tests covering the key conformance areas, especially around the bug fixes
and the most critical action semantics.

### Test File: `actions/conformance_test.go`

Tests to add:

1. **TestNativeAction_ImpureParsing** — verify Impure flag is set when code
   starts with "impure", and that the "impure" line is stripped from the code

2. **TestIsSkolem_MidString** — verify `IsSkolem("foo__bar")` returns true

3. **TestIsGlobalSkolem** — verify `IsGlobalSkolem("__Foo")` = true,
   `IsGlobalSkolem("__foo")` = false, `IsGlobalSkolem("x__Foo")` = false

4. **TestComposeUpdates_Basic** — compose two simple updates and verify
   modified set is union, TR is conjoined with intermediate renaming

5. **TestChoiceAction_Determinize** — verify that with determinize=true,
   binary ChoiceAction converts to IfAction

6. **TestAssignAction_Standard** — verify mk_assign_clauses produces correct
   definition with new/old vocabulary

7. **TestAssignAction_Destructor** — verify destructor assignment traverses
   chain and produces nondet skolem

8. **TestHavocAction_Update** — verify havoc produces unconstrained update

9. **TestSequence_IntUpdate** — verify sequential composition of updates

10. **TestIfAction_WithSome** — verify Some condition elaboration to LocalAction

11. **TestCrashAction_Modifies** — verify CrashAction.modifies() recursively
    collects mutable symbols

12. **TestCallAction_CaptureAvoidance** — verify distinctObjRenaming prevents
    variable capture during parameter substitution

### Test File: `compiler/impure_test.go`

1. **TestCompileNativeDef_ImpureFlag** — create an AST NativeDef with code
   starting with "impure\n...", compile it, verify the resulting
   NativeAction has Impure=true and the code has "impure" stripped

## Steps

### Step 1: Fix NativeAction Impure parsing
- Edit `compiler/phase6.go` around line 1001

### Step 2: Add isSkolem test cases
- Edit `solver/solver_test.go` around line 948

### Step 3: Add conformance tests
- Create `actions/conformance_test.go`
- Create `compiler/impure_test.go`

### Step 4: Verify
```
go build ./...
go test ./actions/...
go test ./compiler/...
go test ./solver/...
go test ./...
```

## Files Modified

- `compiler/phase6.go` — add impure parsing after NativeAction construction
- `solver/solver_test.go` — add mid-string isSkolem test cases
- `actions/conformance_test.go` — new: comprehensive conformance tests
- `compiler/impure_test.go` — new: impure flag compilation test
