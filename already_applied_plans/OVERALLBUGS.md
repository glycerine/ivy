# Code Review Plan: goivy vs Python Ivy Deep Comparison

## Context
The goivy Go port is nearing completion. This review compares every critical subsystem against the Python source of truth (`/Users/jaten/pyivy/ivy/ivy`) to find logic divergences, bugs, and missing functions. The goal is to produce a prioritized fix list so the Go port faithfully matches Python's behavior.

---

## Priority 1: Critical Bugs (Will Cause Incorrect Results)

### Bug 1: `varName` produces wrong names for idx >= 10
- **Files:** `logicutil/logic_utils.go:475-477`, `ivylogic/util.go:315-317`
- **Problem:** `"V" + string(rune('0'+idx))` gives `"V:"` for idx=10 (ASCII ':' = '0'+10). Python uses `'V'+str(i)` which gives `"V10"`.
- **Fix:** Change to `fmt.Sprintf("V%d", idx)`
- **Impact:** Affects `Variables()`, `SymDeclToStr()`, `Extensionality()`, and any function generating variable names.

### Bug 2: `alphaName` same rune bug for idx >= 10
- **File:** `ivylogic/ivylogic.go:96-101`
- **Problem:** `"alpha" + string(rune('0'+idx))` gives wrong chars for idx >= 10.
- **Fix:** Change to `fmt.Sprintf("alpha%d", idx)`
- **Impact:** Affects `TopFunctionSort()` and polymorphic sort generation.

### Bug 3: `NormalizeFreeVariablesTuple` uses all variables, not free variables
- **File:** `logicutil/logic_utils.go:810`
- **Python:** `used_variables_asts` yields only free variables.
- **Go:** Uses `usedVariablesInOrderMulti` which yields ALL variables (bound + free).
- **Fix:** Use `FreeVariables()` or a free-variable-only traversal instead.

### Bug 4: `ResortAst` doesn't resort the `Func` field of Apply nodes
- **File:** `logicutil/logic_utils.go:441-471`
- **Python:** `ivy_logic_utils.py` explicitly resorts `ast.rep` (the function symbol).
- **Fix:** Add `ResortAst(app.Func, smap)` call when processing Apply nodes.

### Bug 5: `ResortSort` doesn't recurse into FunctionSort domain/range
- **File:** `logicutil/logic_utils.go:422-428`
- **Python:** Recursively resorts FunctionSort's domain sorts and range sort.
- **Fix:** When sort is `*FunctionSort`, recursively resort each domain sort and the range sort.

### Bug 6: `apply_mixin` missing parameter substitution and validation
- **File:** `actions/helpers.go:136-150`
- **Python:** `ivy_actions.py:1338-1367` validates param lengths/sorts, builds subst map `formals1 -> formals2`, calls `substitute_constants_ast(action1, subst)`.
- **Go:** Skips ALL validation and substitution. Just concatenates formals.
- **Fix:** Port the validation and substitution logic from Python.

### Bug 7: `Variable.ToConst` returns `*Atom` instead of `*App`
- **File:** `ast/ast.go:235-239`
- **Python:** `ivy_ast.py:399-403` returns `App(prefix + self.rep)`.
- **Fix:** Return `*App` to match Python.

### Bug 8: `BinderVars` doesn't handle `*Some`
- **File:** `ivylogic/util.go:149-161`
- **Python:** `quantifier_vars` handles `Some` (returns its params).
- **Fix:** Add case for `*Some` returning its params.

### Bug 9: `IsInterpretedSort` missing `canonize_sort` and type check
- **File:** `ivylogic/globals.go:30-34`
- **Python:** `ivy_logic.py:1484-1486` first canonizes the sort, then checks if it's `UninterpretedSort` or `EnumeratedSort` AND its name is in `sig.interp`.
- **Fix:** Add canonization step and type guard.

### Bug 10: DEBUG prints left in production code
- **File:** `ast/decl.go:218-234`
- **Fix:** Remove all `fmt.Printf("DEBUG ...")` statements.

---

## Priority 2: Significant Logic Divergences

### Divergence 11: `assert_to_assume` kind dispatch
- **File:** `actions/transforms.go:23-60`
- **Python:** Class-based dispatch (`type(self) in kinds`). SubgoalAction (subclass of AssertAction) matches when AssertAction is in kinds.
- **Go:** String-based matching (`kinds["assert"]`). Won't match "subgoal" when checking for "assert".
- **Fix:** Make subgoal match when assert is in the kinds set, mirroring Python's class hierarchy.

### Divergence 12: `WhileAction.unroll` missing sort detection
- **File:** `actions/transforms.go:291-334`
- **Python:** `ivy_actions.py:1025-1046` examines condition to determine index sort, calls `card(idx_sort)` for bound (max 100).
- **Go:** Takes fixed `bound int` parameter, no sort analysis.
- **Fix:** Port the condition analysis and sort-based bound computation.

### Divergence 13: `IfAction` missing `Some` condition handling
- **File:** `actions/update.go:1412`
- **Python:** `ivy_actions.py:909-939` handles `Some` conditions via `subactions()`, creating local vars, assume/assert for some/some_min/some_max.
- **Fix:** Port the `Some` condition path.

### Divergence 14: `EnsuresAction` missing version check
- **File:** `actions/transforms.go`
- **Python:** Checks `iu.get_numeric_version() <= [1,6]` before applying assert-to-assume.
- **Fix:** Add version check.

### Divergence 15: Schema matching pipeline incomplete (FIXED).

### Divergence 16: `LookupSchema` missing definition constraint conversion (FIXED WE THINK, PLEASE VERIFY).
- **File:** `proof/checker.go:139-153`
- **Python:** `ivy_proof.py:306-322` converts definitions via `goal_conc(schema).to_constraint()` and optionally closes.
- **Fix:** Port the constraint conversion and closing logic.

### Divergence 17: Atom/App constructors missing `flatten()`
- **Files:** `ast/ast.go:88-90` (NewAtom), `ast/ast.go:151-153` (NewApp)
- **Python:** `ivy_ast.py:271-274, 338-340` calls `flatten(terms)` to unnest lists.
- **Fix:** Add flatten logic or verify callers never pass nested slices.

### Divergence 18: `NormalizeNamedBinders` missing assertion
- **File:** `logicutil/logic_utils.go:830-877`
- **Python:** `ivy_logic_utils.py:247-264` asserts `all(nv not in free for nv in nvs)`.
- **Fix:** Add assertion/check that generated V0, V1, ... names don't clash with free variables.

### Divergence 19: Module fields use nil maps instead of Python's `defaultdict(list)`
- **File:** `module/module.go`
- **Python:** Uses `defaultdict(list)` for `postconds`, `mixins`, `sort_destructors`, etc.
- **Fix:** initialize all maps in constructor.

---

## Priority 3: Missing Functions/Methods

### Missing 20: `CallAction.split_returns`
- **Python:** `ivy_actions.py:1293-1301`
- **Action:** Port to `actions/` package.

### Missing 21: `CallAction.decompose`
- **Python:** `ivy_actions.py:1269-1292`
- **Action:** Port to `actions/` package.

### Missing 22: `IfAction.get_cond`
- **Python:** `ivy_actions.py:942-951`
- **Action:** Port to `actions/` package.

### Missing 23: `Definition.ToConstraint`
- **Python:** `ivy_ast.py:188-191`
- **Action:** Add to `ast/formula.go`.

### Missing 24: `PrivateDef` and `ImplementTypeDef` structs
- **Python:** `ivy_ast.py:1224-1228, 1266-1274`
- **Action:** Add to `ast/decl.go`.

### Missing 25: `app_to_atom`, `substitute_constants_ast2`, `parse_name`
- **Python:** `ivy_ast.py:1496-1509, 1817-1839, 1537-1553`
- **Action:** Port to appropriate Go files.

### Missing 26: ast-level equality methods (`__eq__`, `__hash__`)
- **Python:** Symbol, Atom, App, Variable, Literal all have `__eq__` and `__hash__`.
- **Go:** None of these ast types have `Equal()` methods.
- **Action:** Add `Equal()` methods to all ast types matching Python semantics.

---

## Priority 4: String Representation Divergences

### String 27: `BooleanSort.String()` returns `"bool"` not `"Boolean"`
- **File:** `logic/sort.go:40`
- **Fix:** Return `"Boolean"` to match Python.

### String 28: `Not.String()` uses `"~"` prefix vs `"Not(...)"`
- **File:** `logic/formula.go:111-116`
- **Python:** `Not({body})` format.
- **Fix:** Use `Not(...)` format.

### String 29: `Apply.String()` missing space after comma
- **File:** `logic/term.go:164`
- **Python:** `', '.join(...)` (space after comma).
- **Fix:** Use `", "` separator.

### String 30: `Implies.String()` uses `"(%s -> %s)"` vs `"Implies(%s, %s)"`
- **File:** `logic/formula.go:327`
- **Fix:** Match Python format.

### String 31: `Variable.String()` conditionally omits sort
- **File:** `ast/ast.go:225-229`
- **Python:** Always shows `:sort`.
- **Fix:** Always append sort.

### String 32: `LabeledFormula.Temporal` uses `bool` instead of tristate
- **File:** `ast/decl.go:17`
- **Python:** Can be `None`, `True`, or `False`.
- **Fix:** Use `*bool` to represent tristate.

---

## Verification Plan

After implementing fixes:

1. **Unit tests:** Run `go test ./...` to verify no regressions
2. **String comparison:** Create a test that constructs identical ASTs in both Python and Go and compares String() output
3. **Substitution tests:** Test `Substitute` with >10 variables to catch varName bug
4. **ResortAst tests:** Test with Apply nodes containing sorts that need remapping
5. **Mixin tests:** Test `ApplyMixin` with mismatched formals to verify validation
6. **End-to-end:** Run any existing `.ivy` test files through both Python and Go and compare outputs

---

## Execution Order

1. Fix Bugs 1-2 (varName/alphaName) - trivial, high impact
2. Fix Bug 10 (remove DEBUG prints) - trivial
3. Fix Bugs 3-5 (logic_utils: NormalizeFreeVars, ResortAst, ResortSort) - medium effort
4. Fix Bugs 8-9 (BinderVars, IsInterpretedSort) - small fixes
5. Fix Bug 7 (Variable.ToConst type) - small fix
6. Fix Divergences 11-14 (actions subsystem) - medium effort
7. Fix Bug 6 (apply_mixin) - medium effort
8. Fix Divergences 17-19 (flatten, assertions, defaultdict) - medium effort
9. Fix Divergences 15-16 (proof matching) - large effort
10. Fix String representations 27-32 - small effort each
11. Port missing functions 20-26 - large effort, can be incremental
