# TDD Refactor: batch_f_test.go + ivy_compile.go

## Context

batch_f_test.go tests three features: CheckDefinitions, CreateConjActions, and ConjSetup (pass 2). Many tests were originally written as "red" (expected-to-fail) tests, but the implementation in ivy_compile.go has since been completed, making many of those comments stale. This refactor aligns tests with current reality, fixes implementation bugs found by comparison with the Python source (`ivy_compiler.py`), and improves test coverage.

---

## A. Bugs in ivy_compile.go

### Bug 1: Version comparison uses string `>=`/`<=` instead of semantic comparison

**Files:** `compiler/ivy_compile.go:1279, 1428`

- Go: `iu.GetStringVersion() >= "1.7"` and `<= "1.6"` — lexicographic string comparison
- Python: `iu.version_le("1.7", iu.get_string_version())` — proper numeric version comparison
- Impact: Breaks for versions like "1.10" (string "1.10" < "1.7" but semantically 1.10 > 1.7)
- Fix: Add `VersionLE` to `ivyutils` (one already exists in `theory/theory.go` — move/reuse it), then call `iu.VersionLE("1.7", iu.GetStringVersion())` in CheckDefinitions and `iu.VersionLE(iu.GetStringVersion(), "1.6")` in CreateConjActions

### Bug 2: CheckDefinitions — `definesName` extracts from `Defines()` but Python uses `defines()` which returns a string directly

**File:** `compiler/ivy_compile.go:1369-1375`

The `definesName` function calls `d.Defines()` which returns an `lg.Expr`, then type-asserts to `*lg.Symbol`. If `Defines()` doesn't return a `*lg.Symbol` (e.g., returns an `*lg.Apply` for parameterized definitions), the fallback `fmt.Sprint(defExpr)` produces a different string than Python's `defines()`. This is minor but could cause mismatches in edge cases. Verify `lg.Definition.Defines()` always returns `*lg.Symbol` or fix the extraction.

### Bug 3: CheckDefinitions — missing interpreted-symbol check (wire existing SolverName)

**File:** `compiler/ivy_compile.go` (in `checkdef` closure, ~line 1244)

Python: `if slv.solver_name(sym) == None: raise IvyError(lf,'definition of interpreted symbol {}'.format(sym))`
Go: Missing entirely. This means Go silently accepts definitions of interpreted symbols (like `+`, `*`, etc.) that Python rejects.

**Existing code:** `SolverName` is already fully ported as a method on `*Solver` in `solver/z3convert.go:963-1026`. It checks:
- bfe symbols → return "" if native Z3
- polymorphic symbols (`iu.PolymorphicSymbols`) on interpreted sorts → return ""
- `sig.Interp` membership → return ""
- z3 builtins (`bit0`, `bit1`) → return ""

**Problem:** `SolverName` is a method on `*Solver` requiring a Z3 solver instance. `CheckDefinitions` runs at compile time without a solver. We need a **standalone version** that only needs `sig` (which the compiler already has).

**Fix — add `IsInterpretedSymbol(sym, sig)` to compiler:**
Create a standalone function in `compiler/ivy_compile.go` that replicates the `SolverName == ""` check using only `sig.Interp` and `iu.PolymorphicSymbols` (no Z3 needed):

```go
// IsInterpretedSymbol returns true when the symbol is natively interpreted
// by Z3 and should not be user-defined. Standalone version of
// solver.SolverName(sym) == "" for use at compile time without a Solver instance.
// Matches Python's `slv.solver_name(sym) == None` check in check_definitions.
func IsInterpretedSymbol(name string, sym *lg.Symbol, sig *il.Sig) bool {
    if sig == nil { return false }
    // Polymorphic symbols on interpreted sorts
    if _, isPoly := iu.PolymorphicSymbols[name]; isPoly {
        if fs, ok := sym.CSort.(*lg.FunctionSort); ok && len(fs.Domain()) > 0 {
            domName := fs.Domain()[0].String()
            if name == "arrcst" { domName = fs.Range().String() }
            if interp, has := sig.Interp[domName]; has {
                if _, isEnum := interp.(*lg.EnumeratedSort); !isEnum {
                    return true
                }
            }
        }
    }
    // Direct interpretation
    if _, has := sig.Interp[name]; has { return true }
    // Z3 builtins
    if name == "bit0" || name == "bit1" { return true }
    return false
}
```

Then in the `checkdef` closure, add before `defs[sym] = lf`:
```go
if IsInterpretedSymbol(sym, symObj, mod.Sig) {
    return &lg.IvyError{Msg: fmt.Sprintf("definition of interpreted symbol %s", sym)}
}
```

This requires passing the `*lg.Symbol` object (not just the name string) into `checkdef`. Adjust the closure signature and call sites to pass both the name and the symbol.

**Add test for interpreted-symbol rejection:**
```go
func TestCheckDefinitions_InterpretedSymbolError(t *testing.T) {
    mod := module.New()
    mod.Sig = il.NewSig()
    mod.Sig.Interp["myint"] = &lg.UninterpretedSort{SortName: "int"} // mark as interpreted
    // definition of "myint" should be rejected
    defMyint := makeLabeledDef("def_myint", makeLogicDef("myint"))
    mod.LabeledProps = []*ast.LabeledFormula{defMyint}
    err := CheckDefinitions(mod)
    if err == nil {
        t.Fatal("expected error for definition of interpreted symbol 'myint'")
    }
}
```

### Bug 4: CheckDefinitions — missing `opt_mutax` guard

**File:** `compiler/ivy_compile.go:1296`

Python wraps the axiom-interference check in `if not opt_mutax.get():`. Go does the check unconditionally. When `opt_mutax` is enabled (mutable axioms allowed), this would cause false errors.

Fix: Add `opt_mutax` option to ivyutils and check it before the axiom interference loop.

---

## B. Test Improvements (batch_f_test.go)

### B1: Remove stale "EXPECTED TO FAIL" comments

The following tests have stale comments because the Go implementation now handles these cases:

| Test | Comment | Status |
|------|---------|--------|
| Test 3 (RedefinitionError) | "Go only logs, doesn't return error" | Go DOES return error (line 1246) |
| Test 4 (NativeDefinitionRedefinitionError) | "Go doesn't check NativeDefinitions" | Go DOES check (line 1259) |
| Test 5 (NamedRedefinitionError) | "Go doesn't check Named" | Go DOES check (line 1269) |
| Test 7 (SelfLoopRequiresProof) | "Go only logs, doesn't return error" | Go DOES return error (line 1349) |
| Test 8 (SelfLoopWithProofAccepted) | "Go doesn't call admit_definition" | admit_definition is TODO but the no-error path works |
| Test 9 (ActionInterference_ModifiesAxiomSymbol) | "Go doesn't check action interference" | Go DOES check (line 1278) |
| Test 10 (ActionInterference_ModifiesDefinedSymbol) | "Go doesn't check action interference" | Go DOES check (line 1308) |
| Test 12 (VersionGate) | "Go has no version check" | Go DOES check (line 1428) |

Action: Remove all stale "EXPECTED TO FAIL" comments from these tests.

### B2: Test 17 — Replace t.Skip with a real test

Test 17 (InterferenceDetection) uses `t.Skip()`, which hides a genuinely missing feature. The interference detection from Python's `create_conj_actions` (the `do_check_interference` section) is not yet ported to Go.

Action: Keep as a clearly-commented TODO test that documents what's missing, but don't `t.Skip` — use `t.Log("TODO: interference detection not yet ported")` or leave the test to fail explicitly when ready.

### B3: Add new test — version comparison edge case

Add a test that exercises version "1.10" to catch Bug 1:

```go
func TestCheckDefinitions_VersionComparisonSemantic(t *testing.T) {
    oldVer := iu.GetStringVersion()
    defer iu.SetStringVersion(oldVer)
    iu.SetStringVersion("1.10") // > 1.7 semantically but < "1.7" lexicographically

    mod := module.New()
    // axiom with 'f', action assigns 'f' — should error at version 1.10
    ...
}
```

### B4: Add test — definition ordering matches Python

Python's `check_definitions` processes props in order, preserving insertion order for `mod.definitions`. Add a test verifying that multiple clean definitions maintain their order.

### B5: Verify `UsedConstantsList` vs Python's `used_symbols_ast`

The stale-symbol detection (ivy_compile.go:1224) uses `lu.UsedConstantsList(expr)` on the whole formula. Python uses `lu.used_symbols_ast(prop.formula)`. These must be functionally equivalent. Add a targeted test: create a definition `h = f(g(x))` where `g` is stale, and verify `h` stays in LabeledProps.

---

## C. Implementation Changes

### C1: Move VersionLE to ivyutils

**From:** `theory/theory.go:256`
**To:** `ivyutils/names.go` (new function)

Copy the existing `VersionLE` function to ivyutils so it's accessible from the compiler package without importing theory (which could create circular deps).

### C2: Fix version comparisons in ivy_compile.go

**File:** `compiler/ivy_compile.go`

- Line 1279: `if iu.GetStringVersion() >= "1.7"` → `if iu.VersionLE("1.7", iu.GetStringVersion())`
- Line 1428: `if iu.GetStringVersion() <= "1.6"` → `if iu.VersionLE(iu.GetStringVersion(), "1.6")`

### C3: Clean up test comments

**File:** `compiler/batch_f_test.go`

Remove all "EXPECTED TO FAIL" lines from tests that now pass.

---

## D. Execution Order

1. **C1**: Move `VersionLE` to ivyutils
2. **C2**: Fix version comparisons in ivy_compile.go
3. **Bug 3**: Add `IsInterpretedSymbol` function + wire into `checkdef` closure
4. **B1/C3**: Remove stale test comments
5. **B3**: Add version edge-case test
6. **B3+**: Add interpreted-symbol rejection test
7. **B5**: Add stale-symbol transitive test
8. **B4**: Add definition-ordering test
9. **B2**: Fix Test 17 Skip
10. Run tests: `cd ~/goivy && make test` (per feedback_z3_test_path.md)
11. If any test reveals a bug → fix the bug immediately (per feedback_fix_bugs_in_tests.md)

## E. Files Modified

- `ivyutils/names.go` — add `VersionLE` (copied from theory/theory.go)
- `compiler/ivy_compile.go` — fix 2 version comparisons, add `IsInterpretedSymbol`, wire into `checkdef`
- `compiler/batch_f_test.go` — remove stale comments, add 4 new tests, fix Test 17
- `theory/theory.go` — update to call `iu.VersionLE` instead of local copy (or leave both)

## F. Verification

1. `cd ~/goivy && make test` — all compiler tests pass
2. Manually verify Test 17 output (should log, not skip)
3. Verify the new version "1.10" test passes after C2 fix
4. Verify interpreted-symbol test catches the new error path
