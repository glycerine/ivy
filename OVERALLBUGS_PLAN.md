# OVERALLBUGS Fix Plan — goivy Faithful Port Corrections
**Created:** 2026-03-22 (current session)

## Context

The goivy Go port has 32 identified issues (bugs, logic divergences, missing functions, string representation mismatches) documented in `OVERALLBUGS.md`. This plan provides exact implementation details for every item so that execution can proceed without stubs or guesswork. Each fix references the Python source of truth and the exact Go code to modify.

Items already verified as FIXED or already implemented are noted. Items marked as ALREADY DONE will be skipped.

---

## Phase 1: Trivial High-Impact Fixes (Bugs 1, 2, 10)

### Bug 1: `varName` rune bug — TWO locations
**Status:** CONFIRMED BUG

**File 1:** `logicutil/logic_utils.go:475-477`
```go
// CURRENT (broken for idx >= 10):
func varName(idx int) string {
    return "V" + string(rune('0'+idx))
}
// FIX:
func varName(idx int) string {
    return fmt.Sprintf("V%d", idx)
}
```

**File 2:** `ivylogic/util.go:315-317`
```go
// Same fix:
func varName(idx int) string {
    return fmt.Sprintf("V%d", idx)
}
```

**Verify:** Ensure `fmt` is imported in both files (it already is in both).

### Bug 2: `alphaName` same rune bug
**Status:** CONFIRMED BUG

**File:** `ivylogic/ivylogic.go:96-101`
```go
// CURRENT (broken for idx >= 10):
func alphaName(idx int) string {
    if idx == 0 {
        return "alpha"
    }
    return "alpha" + string(rune('0'+idx))
}
// FIX:
func alphaName(idx int) string {
    if idx == 0 {
        return "alpha"
    }
    return fmt.Sprintf("alpha%d", idx)
}
```

### Bug 10: DEBUG prints in production code
**Status:** CONFIRMED BUG

**File:** `ast/decl.go:218-236`

Remove ALL debug code from `NewActionDef`. Delete lines 218-234 (the `fmt.Printf("DEBUG ...")` line, the entire `dumpLeaves` closure and its call, and the trailing DEBUG printf on line 236). Keep line 235 (`body = SubstPrefixAtomsAst(body, subst, nil, nil, nil)`) and line 237-238 (the return).

The resulting code should be:
```go
    // (after building subst map on lines 211-217)
    body = SubstPrefixAtomsAst(body, subst, nil, nil, nil)
}
return &ActionDef{Name: name, Body: body, FormalParams: fmlParams, FormalReturns: fmlReturns}
```

---

## Phase 2: Logic Utils Fixes (Bugs 3, 4, 5)

### Bug 3: `NormalizeFreeVariablesTuple` uses ALL variables instead of FREE variables
**Status:** CONFIRMED BUG

**File:** `logicutil/logic_utils.go:810`

**Python source:** `ivy_logic_utils.py:231` uses `used_variables_asts(asts)` which is `apply_gen_to_list(variables_ast)`. The `variables_ast` function (line 461-471) explicitly filters out bound variables from binders. So it yields FREE variables only.

**Go problem:** `usedVariablesInOrderMulti` (line 774) calls `usedVariablesInOrderRec` (line 721) which includes ForAll/Exists bound variables in the result.

**Fix:** Create a `freeVariablesInOrderMulti` function that mirrors `usedVariablesInOrderMulti` but excludes bound variables, similar to how `freeVariablesInOrder` works. Then use it in `NormalizeFreeVariablesTuple`.

```go
// Add after freeVariablesInOrder (around line 659):
func freeVariablesInOrderMulti(asts []logic.Expr) []*logic.Variable {
    seen := make(map[string]bool)
    var result []*logic.Variable
    for _, ast := range asts {
        freeVariablesInOrderRec(ast, &result, seen)
    }
    return result
}
```

Check that `freeVariablesInOrderRec` exists (it's used by `freeVariablesInOrder`). If it doesn't exist, we need to create it by adapting `usedVariablesInOrderRec` to skip bound vars in binders:

```go
func freeVariablesInOrderRec(ast logic.Expr, result *[]*logic.Variable, seen map[string]bool) {
    if v, ok := ast.(*logic.Variable); ok {
        if !seen[v.Name] {
            seen[v.Name] = true
            *result = append(*result, v)
        }
        return
    }
    // For binders, skip bound variables and only recurse into body/args
    switch t := ast.(type) {
    case *logic.ForAll:
        bounds := make(map[string]bool)
        for _, v := range t.Variables {
            bounds[v.Name] = true
        }
        // Recurse into body, but filter bound vars via a temporary seen set
        savedSeen := make(map[string]bool)
        for k, v := range seen { savedSeen[k] = v }
        for k := range bounds { seen[k] = true }
        freeVariablesInOrderRec(t.Body, result, seen)
        // Restore: remove bounds that weren't in original seen
        for k := range bounds {
            if !savedSeen[k] { delete(seen, k) }
        }
        return
    case *logic.Exists:
        // Same pattern as ForAll
        bounds := make(map[string]bool)
        for _, v := range t.Variables { bounds[v.Name] = true }
        savedSeen := make(map[string]bool)
        for k, v := range seen { savedSeen[k] = v }
        for k := range bounds { seen[k] = true }
        freeVariablesInOrderRec(t.Body, result, seen)
        for k := range bounds {
            if !savedSeen[k] { delete(seen, k) }
        }
        return
    // Lambda, NamedBinder similarly
    }
    // Default: recurse into children
    for _, c := range ast.Children() {
        freeVariablesInOrderRec(c, result, seen)
    }
}
```

**Alternative (simpler):** Reuse existing `FreeVariablesList` from `logicutil/logicutil.go:641` which already correctly computes free variables, but we need them in-order across multiple ASTs. Check if we can adapt.

**Then change line 810:**
```go
// BEFORE:
for _, v := range usedVariablesInOrderMulti(asts) {
// AFTER:
for _, v := range freeVariablesInOrderMulti(asts) {
```

### Bug 4: `ResortAst` doesn't resort `Func` field of Apply nodes
**Status:** CONFIRMED BUG

**File:** `logicutil/logic_utils.go:441-471`

**Python source:** `ivy_logic_utils.py:422-424`:
```python
args = [resort_ast(x,subs) for x in ast.args]
if is_app(ast):
    return resort_symbol(ast.rep)(*args)  # resorts the function symbol
```

**Go problem:** `Apply.Children()` returns `Terms` only (confirmed at `logic/term.go:147`), so the default case in `ResortAst` never touches `Func`.

**Fix:** Add explicit `*logic.Apply` case before the `default`:
```go
case *logic.Apply:
    // Resort the function symbol (ast.rep in Python)
    newFunc := ResortAst(t.Func, subs)
    // Resort the arguments
    newTerms := make([]logic.Expr, len(t.Terms))
    changed := newFunc != t.Func
    for i, term := range t.Terms {
        nt := ResortAst(term, subs)
        newTerms[i] = nt
        if nt != term {
            changed = true
        }
    }
    if !changed {
        return ast
    }
    result, err := logic.NewApply(newFunc, newTerms...)
    if err != nil {
        return &logic.Apply{Func: newFunc, Terms: newTerms}
    }
    return result
```

Insert this case after the `case *logic.Symbol:` block (after line 451).

### Bug 5: `ResortSort` doesn't recurse into FunctionSort domain/range
**Status:** CONFIRMED BUG

**File:** `logicutil/logic_utils.go:422-428`

**Python source:** `ivy_logic_utils.py:398-410`:
```python
def resort_sort(sort,subs):
    if isinstance(sort,UnionSort):
        res = UnionSort()
        res.sorts = [resort_sort(s,subs) for s in sort.sorts]
        return res
    if sort in subs:
        return subs[sort]
    if isinstance(sort,FunctionSort):
        return FunctionSort(*[resort_sort(s,subs) for s in (sort.dom+(sort.rng,))])
    return sort
```

**Fix:** Replace the current `ResortSort` with:
```go
func ResortSort(s logic.Sort, subs map[logic.NodeKey]logic.Sort) logic.Sort {
    // Handle UnionSort
    if us, ok := s.(*logic.UnionSort); ok {
        newSorts := make([]logic.Sort, len(us.Sorts))
        for i, sub := range us.Sorts {
            newSorts[i] = ResortSort(sub, subs)
        }
        return &logic.UnionSort{Sorts: newSorts}
    }
    // Direct substitution
    key := logic.SortKey(s)
    if mapped, ok := subs[key]; ok {
        return mapped
    }
    // Recurse into FunctionSort
    if fs, ok := s.(*logic.FunctionSort); ok {
        newDomain := make([]logic.Sort, len(fs.Domain))
        for i, d := range fs.Domain {
            newDomain[i] = ResortSort(d, subs)
        }
        newRange := ResortSort(fs.Range, subs)
        return logic.NewFunctionSort(newDomain, newRange)
    }
    return s
}
```

**Prerequisite:** Verify `*logic.UnionSort` exists. If not, skip that branch. Check `*logic.FunctionSort` struct fields (`Domain []Sort`, `Range Sort`). Verify `logic.NewFunctionSort` constructor exists.

---

## Phase 3: Small Logic Fixes (Bugs 7, 8, 9)

### Bug 7: `Variable.ToConst` returns `*Atom` instead of `*App`
**Status:** CONFIRMED BUG

**File:** `ast/ast.go:235-239`

**Python source:** `ivy_ast.py:399-403`:
```python
def to_const(self, prefix):
    res = App(prefix + self.rep)
    if hasattr(self,'sort'):
        res.sort = self.sort
    return res
```

**Fix:** Change to return `*App`:
```go
func (v *Variable) ToConst(prefix string) *App {
    rep := &Symbol{Rep: prefix + v.Rep}  // App.Rep is a Node (function symbol)
    a := NewApp(rep)
    a.ASort = v.VSort
    return a
}
```

**Note:** This changes the return type from `*Atom` to `*App`. Search for all call sites of `ToConst` to ensure they accept `Node` interface (not `*Atom` specifically). The `ast.Node` interface should cover both. If any call site does a type assertion to `*Atom`, that needs updating.

### Bug 8: `BinderVars` doesn't handle `*Some`
**Status:** CONFIRMED BUG

**File:** `ivylogic/util.go:149-161`

**Python source:** `ivy_logic.py:637-638`:
```python
def quantifier_vars(term):
    return term.variables
```
Python's `quantifier_vars` handles all quantifier-like things including `Some`, because `Some` has a `variables` attribute via inheritance.

The Go `Some` struct is at `ivylogic/formula.go:11` with `Params []lg.Expr`.

**Fix:** Add `*Some` case to `BinderVars`:
```go
case *Some:
    // Some's Params are the bound variables
    vars := make([]*lg.Variable, 0, len(t.Params))
    for _, p := range t.Params {
        if v, ok := p.(*lg.Variable); ok {
            vars = append(vars, v)
        }
    }
    return vars
```

Insert after the `case *lg.NamedBinder:` block. The `Some` type is in `ivylogic/formula.go` — check import path.

### Bug 9: `IsInterpretedSort` missing canonize and type check
**Status:** CONFIRMED BUG

**File:** `ivylogic/globals.go:30-34`

**Python source:** `ivy_logic.py:1484-1486`:
```python
def is_interpreted_sort(s):
    s = canonize_sort(s)
    return (isinstance(s,UninterpretedSort) or isinstance(s,EnumeratedSort)) and s.name in sig.interp
```

**Go `CanonizeSort` already exists** at `ivylogic/ivylogic.go:452`.

**Fix:**
```go
func IsInterpretedSort(sig *Sig, s lg.Sort) bool {
    s = CanonizeSort(sig, s)
    // Must be UninterpretedSort or EnumeratedSort
    switch cs := s.(type) {
    case *lg.UninterpretedSort:
        _, ok := sig.Interp[cs.Name]
        return ok
    case *lg.EnumeratedSort:
        _, ok := sig.Interp[cs.Name]
        return ok
    default:
        return false
    }
}
```

---

## Phase 4: Actions Subsystem (Bugs 6, 11-14)

### Bug 6: `ApplyMixin` missing validation and substitution
**Status:** CONFIRMED BUG

**File:** `actions/helpers.go:136-150`

**Python source:** `ivy_actions.py:1338-1367` does:
1. Validates `len(action1.formal_params) == len(action2.formal_params)`
2. Validates `len(action1.formal_returns) == len(action2.formal_returns)`
3. Validates each pair has matching sorts
4. Builds substitution map: `formals1 -> formals2`
5. Applies `substitute_constants_ast(action1, subst)` to rename action1's params to action2's
6. Then concatenates

**Fix:** The Go `ApplyMixin` needs a `decl` parameter for error context, plus validation and substitution. The function `SubstituteConstantsAst` exists at `ast/rewrite.go:691`.

```go
func ApplyMixin(decl interface{}, action1, action2 Action, isAfter bool) (Action, error) {
    fp1, fp2 := action1.GetFormalParams(), action2.GetFormalParams()
    fr1, fr2 := action1.GetFormalReturns(), action2.GetFormalReturns()

    if len(fp1) != len(fp2) {
        return nil, fmt.Errorf("mixin has wrong number of input parameters")
    }
    if len(fr1) != len(fr2) {
        return nil, fmt.Errorf("mixin has wrong number of output parameters")
    }

    formals1 := append(fp1, fr1...)
    formals2 := append(fp2, fr2...)
    for i, x := range formals1 {
        y := formals2[i]
        // Compare sorts
        xs, ys := getSort(x), getSort(y)
        if xs != nil && ys != nil && lg.SortKey(xs) != lg.SortKey(ys) {
            return nil, fmt.Errorf("parameter %s of mixin has wrong sort", x)
        }
    }

    // Build substitution and apply to action1
    subst := make(map[string]ast.Node, len(formals1))
    for i, f1 := range formals1 {
        subst[symbolName(f1)] = nodeFromExpr(formals2[i])
    }
    action1Renamed := substituteConstantsAction(action1, subst)

    var res *Sequence
    if isAfter {
        res = ConcatActions(action2, action1Renamed)
    } else {
        res = ConcatActions(action1Renamed, action2)
    }
    res.SetLineno(action1.GetLineno())
    res.SetFormalParams(action2.GetFormalParams())
    res.SetFormalReturns(action2.GetFormalReturns())
    return res, nil
}
```

**Key helpers needed:** A function to extract sort from an `lg.Expr` (symbol), and a function to apply substitution to an action. Check if `ast.SubstituteConstantsAst` works on action nodes or if we need a logic-level equivalent. The existing `SubstituteConstantsAst` in `ast/rewrite.go:691` works on `ast.Node`. We may need `co.SubstituteConstantsAST` from the `logicutil` package for `lg.Expr` types. Search for existing `SubstituteConstants` functions.

**Impact:** All callers of `ApplyMixin` need to be updated for the new signature (added `decl`, returns `error`).

### Divergence 11: `AssertToAssume` — SubgoalAction doesn't match "assert"
**Status:** CONFIRMED ISSUE

**File:** `actions/transforms.go:23-60`

**Python:** `SubgoalAction` inherits from `AssertAction`, so `type(self) in kinds` matches `AssertAction` for a `SubgoalAction`. Go's `SubgoalAction` embeds `AssertAction` (at `actions/extra_actions.go:21-24`) but the switch statement matches `*AssertAction` which won't match `*SubgoalAction`.

**Fix:** Add `*SubgoalAction` case before `*AssertAction`:
```go
case *SubgoalAction:
    // SubgoalAction embeds AssertAction — match when "assert" is in kinds
    if kinds["assert"] || kinds["subgoal"] {
        assume := NewAssumeAction(a.Formula)
        assume.ActionBase = a.ActionBase
        return assume
    }
    return a
```

Insert this case BEFORE the `case *AssertAction:` case (around line 47).

### Divergence 12: `WhileAction.unroll` — missing sort detection
**Status:** CONFIRMED DIVERGENCE

**File:** `actions/transforms.go:291-334`

**Python source:** `ivy_actions.py:1025-1046` — examines the while condition to determine an index sort, then calls `card(idx_sort)` to determine iteration bound.

**Fix:** Change `UnrollLoops` to accept a cardinality function instead of a fixed bound:

```go
// CardFunc computes the cardinality of a sort. Returns -1 if unknown.
type CardFunc func(s logic.Sort) int

func UnrollLoops(action Action, card CardFunc) Action {
    if action == nil { return nil }
    switch a := action.(type) {
    case *WhileAction:
        cond := a.Cond
        // Peel through And to find the comparison
        for {
            if andNode, ok := cond.(*lg.And); ok && len(andNode.Args()) > 0 {
                cond = andNode.Args()[0]
            } else {
                break
            }
        }
        // Determine index sort from condition
        var idxSort logic.Sort
        if app, ok := cond.(*lg.Apply); ok {
            if sym, ok := app.Func.(*lg.Symbol); ok {
                name := sym.Name
                if name == "<" || name == ">" || name == "<=" || name == ">=" {
                    if len(app.Terms) > 0 {
                        idxSort = app.Terms[0].GetSort()
                    }
                }
            }
        } else if notNode, ok := cond.(*lg.Not); ok {
            if eq, ok := notNode.Body.(*lg.Eq); ok {
                idxSort = eq.T1.GetSort()
            }
        }
        cardsort := card(idxSort)
        if cardsort < 0 {
            panic(fmt.Sprintf("cannot determine iteration bound for loop over %v", idxSort))
        }
        if cardsort > 100 {
            panic(fmt.Sprintf("cowardly refusing to unroll loop over %v %d times", idxSort, cardsort))
        }
        // Build unrolled if-then-else chain
        res := NewIfAction(a.Cond, WrapAction(NewAssumeAction(&lg.Or{})))
        for i := 0; i < cardsort; i++ {
            body := a.Body
            seq := NewSequence(body, WrapAction(res))
            res = NewIfAction(a.Cond, WrapAction(seq))
        }
        CopyFormalsTo(a, res)
        return res
    default:
        // recursive case same as current
        ...
    }
}
```

**Impact:** All callers of `UnrollLoops` need to pass a `CardFunc` instead of `int`. Search for callers.

### Divergence 13: `IfAction` missing `Some` condition handling in `IntUpdate`
**Status:** NEEDS VERIFICATION

**File:** `actions/update.go:1412`

**Python source:** `ivy_actions.py:909-939` handles `Some` conditions via `subactions()`, creating local vars, assume/assert for some/some_min/some_max.

**Fix:** In `IfAction.IntUpdate`, add handling for when `a.Cond` is `*SomeCondition` (or `*ast.Some`). The Python code at line 915 checks `isinstance(self.args[0], ivy_ast.Some)` and branches to the subactions path. The Go code should mirror this:

```go
// In IntUpdate:
if _, ok := a.Cond.(*SomeCondition); ok {
    // Use subactions path
    ifPart, elsePart := a.Subactions()
    ifUpdate := ifPart.IntUpdate(ctx)
    elseUpdate := elsePart.IntUpdate(ctx)
    res := JoinAction(ifUpdate, elseUpdate, ctx.BackgroundTheory())
    // Fix reversed IteAnnotation (Python hack at lines 930-934)
    ...
    return res
}
```

Check whether `Subactions()` method exists on `IfAction`. If not, port it from Python `ivy_actions.py:899-908`.

### Divergence 14: `EnsuresAction` missing version check
**Status:** CONFIRMED DIVERGENCE

**File:** `actions/transforms.go:38-45`

**Python:** Checks `iu.get_numeric_version() <= [1,6]` before applying assert-to-assume for EnsuresAction.

**Fix:** Import `ivyutils` and add version check:
```go
case *EnsuresAction:
    if kinds["ensure"] {
        ver := iu.GetNumericVersion()
        if len(ver) >= 2 && (ver[0] < 1 || (ver[0] == 1 && ver[1] <= 6)) {
            assume := NewAssumeAction(a.Formula)
            assume.ActionBase = a.ActionBase
            return assume
        }
    }
    return a
```

**`GetNumericVersion` exists** at `ivyutils/names.go:349`.

---

## Phase 5: Structural Fixes (Divergences 16, 17, 18, 19)

### Divergence 16: `LookupSchema` definition constraint conversion
**Status:** ALREADY IMPLEMENTED (VERIFY)

**File:** `proof/checker.go:149-174`

The current code already has `DefinitionToConstraint` and `CloseFormula` calls. Compare with Python `ivy_proof.py:306-322`. The implementation looks correct:
- Checks schemata first
- Falls back to definitions, converts via `DefinitionToConstraint`, optionally closes
- Falls back to goal premises

**Action:** Read the complete `LookupSchema` function and compare against Python. If it matches, mark as DONE. Key question: does it handle the case where `conc` is NOT a `*lg.Definition`? The current code at line 170-174 handles this case by returning `d` as-is, which matches Python's behavior (Python doesn't guard against non-Definition conc, but in practice it always is).

**Verdict:** Likely DONE. Verify by reading proof/checker.go fully.

### Divergence 17: Atom/App constructors missing `flatten()`
**Status:** CONFIRMED DIVERGENCE

**Files:** `ast/ast.go:88-90` (NewAtom), `ast/ast.go:151-153` (NewApp)

**Python source:** `ivy_ast.py:271-274, 338-340` calls `flatten(terms)`. Python's `flatten` (from `ivy_utils.py:93-97`) recursively unnests lists/tuples into a single flat list.

**Fix:** Add a `flatten` helper and use it in constructors:
```go
// In ast/ast.go:
func flattenNodes(nodes []Node) []Node {
    var result []Node
    for _, n := range nodes {
        if ns, ok := n.([]Node); ok {
            result = append(result, flattenNodes(ns)...)
        } else {
            result = append(result, n)
        }
    }
    return result
}

func NewAtom(rep string, terms ...Node) *Atom {
    return &Atom{Rep: rep, Terms: flattenNodes(terms)}
}

func NewApp(rep Node, terms ...Node) *App {
    return &App{Rep: rep, Terms: flattenNodes(terms)}
}
```

**Note:** In Go, variadic `...Node` can't contain nested slices unless someone passes `[]Node` as a `Node`. Check if Go callers ever do this. If not, the flatten is a no-op and we can skip it but add it for safety.

### Divergence 18: `NormalizeNamedBinders` missing free-variable clash assertion
**Status:** CONFIRMED DIVERGENCE

**File:** `logicutil/logic_utils.go:830-877`

**Python source:** `ivy_logic_utils.py:247-264` asserts `all(nv not in free for nv in nvs)` — the generated V0, V1, ... names must not clash with existing free variables.

**Fix:** Add assertion after generating `nvs` (around line 842):
```go
// After building nvs, check for clashes with free variables
free := FreeVariablesSet(ast)  // need a set-returning version
for _, nv := range nvs {
    if free[nv.Name] {
        panic(fmt.Sprintf("NormalizeNamedBinders: generated variable %s clashes with free variable", nv.Name))
    }
}
```

Check if `FreeVariablesSet` exists or if we can use `FreeVariablesList` and build a set.

### Divergence 19: Module fields — nil maps
**Status:** ALREADY FIXED

**File:** `module/module.go`

The `Module.Clear()` method (lines 218-284) already initializes ALL maps with `make()` calls. No nil maps detected.

**Action:** SKIP — already done.

---

## Phase 6: String Representation Fixes (Bugs 27-32)

### String 27: `BooleanSort.String()` returns `"bool"` not `"Boolean"`
**File:** `logic/sort.go:40`
```go
// CURRENT:
func (s *BooleanSort) String() string { return "bool" }
// FIX:
func (s *BooleanSort) String() string { return "Boolean" }
```

### String 28: `Not.String()` format
**File:** `logic/formula.go:111-116`
```go
// CURRENT:
func (n *Not) String() string {
    if eq, ok := n.Body.(*Eq); ok {
        return fmt.Sprintf("(%s != %s)", eq.T1, eq.T2)
    }
    return fmt.Sprintf("~%s", n.Body)
}
// FIX — match Python's Not({body}) format:
func (n *Not) String() string {
    if eq, ok := n.Body.(*Eq); ok {
        return fmt.Sprintf("(%s != %s)", eq.T1, eq.T2)
    }
    return fmt.Sprintf("Not(%s)", n.Body)
}
```

**Note:** Keep the `!=` special case for negated equality — Python does this too.

### String 29: `Apply.String()` missing space after comma
**File:** `logic/term.go:164`
```go
// CURRENT:
return fmt.Sprintf("%s(%s)", a.Func.String(), strings.Join(parts, ","))
// FIX:
return fmt.Sprintf("%s(%s)", a.Func.String(), strings.Join(parts, ", "))
```

### String 30: `Implies.String()` format
**File:** `logic/formula.go:327`
```go
// CURRENT:
func (i *Implies) String() string { return fmt.Sprintf("(%s -> %s)", i.T1, i.T2) }
// FIX:
func (i *Implies) String() string { return fmt.Sprintf("Implies(%s, %s)", i.T1, i.T2) }
```

### String 31: `Variable.String()` — always show sort
**File:** `ast/ast.go:225-229`
```go
// CURRENT:
func (v *Variable) String() string {
    if v.VSort != nil {
        return v.Rep + ":" + fmt.Sprint(v.VSort)
    }
    return v.Rep
}
// FIX — Python always shows :sort:
func (v *Variable) String() string {
    return v.Rep + ":" + fmt.Sprint(v.VSort)
}
```

**Caveat:** If `VSort` is nil, this will print `V:nil`. Check Python behavior when sort is None. If Python also omits sort when None, keep the current code. Python likely always has a sort set, so nil case may not arise in practice. Verify before changing.

### String 32: `LabeledFormula.Temporal` — bool to tristate
**File:** `ast/decl.go:17`
```go
// CURRENT:
Temporal     bool
// FIX:
Temporal     *bool  // nil = not set, true = temporal, false = non-temporal
```

**Impact:** All code that reads/writes `Temporal` must use pointer operations. Search for all uses of `.Temporal` in the codebase and update them. Common patterns:
- `if lf.Temporal {` → `if lf.Temporal != nil && *lf.Temporal {`
- `lf.Temporal = true` → `t := true; lf.Temporal = &t` (or create a helper `boolPtr(b bool) *bool`)

---

## Phase 7: Missing Functions (Items 20-26)

### Missing 20: `CallAction.SplitReturns` — ALREADY IMPLEMENTED
**File:** `actions/action.go:672`
**Action:** SKIP — already done.

### Missing 21: `CallAction.Decompose` — ALREADY IMPLEMENTED
**File:** `actions/action.go:1251` (basic version), `actions/action.go:1349` (`DecomposeWithState` for full Python semantics)
**Action:** SKIP — already done.

### Missing 22: `IfAction.GetCond` — ALREADY IMPLEMENTED
**File:** `actions/action.go:529`
**Action:** SKIP — already done.

### Missing 23: `Definition.ToConstraint` — MISSING in ast package
**File:** Should be in `ast/formula.go` or wherever `Definition` is defined.

**Python source:** `ivy_ast.py:188-191`:
```python
def to_constraint(self):
    if isinstance(self.args[0], App):
        return Atom(equals, self.args)
    return Iff(*self.args)
```

**Note:** `DefinitionToConstraint` already exists in `ivylogic/constraint.go:16` for `*lg.Definition` (logic-level). But the Python has it as a METHOD on the ast-level `Definition` class. We need the ast-level version too.

**Fix:** Check if ast package has a `Definition` struct. If so, add:
```go
func (d *Definition) ToConstraint() Node {
    if _, ok := d.Args()[0].(*App); ok {
        return NewAtom("=", d.Args()...)
    }
    return NewIff(d.Args()[0], d.Args()[1])
}
```

Verify the `Definition` struct location and `equals` symbol representation.

### Missing 24: `PrivateDef` and `ImplementTypeDef` structs — MISSING
**File:** `ast/decl.go`

**Python source:** `ivy_ast.py:1224-1228, 1266-1274`

**Fix:** Add to `ast/decl.go`:
```go
type PrivateDef struct {
    Base
    Args []Node
}

func (p *PrivateDef) Privatized() string {
    return p.Args[0].Relname()
}

func (p *PrivateDef) String() string {
    return fmt.Sprintf("private %s", p.Args[0])
}

type ImplementTypeDef struct {
    Base
    Args []Node
}

func (d *ImplementTypeDef) Implemented() string {
    return d.Args[0].Relname()
}

func (d *ImplementTypeDef) Implementer() string {
    return d.Args[1].Relname()
}

func (d *ImplementTypeDef) String() string {
    return fmt.Sprintf("%s with %s", d.Implemented(), d.Implementer())
}
```

Ensure they implement the `Node` interface. Add `Clone()`, `Args()` methods as needed.

### Missing 25: `AppToAtom`, `SubstituteConstantsAst2` — PARTIALLY DONE
- `ParseName` — ALREADY EXISTS at `ast/rewrite.go:138`
- `SubstituteConstantsAst` — EXISTS at `ast/rewrite.go:691`
- `AppToAtom` — MISSING
- `SubstituteConstantsAst2` — MISSING

**AppToAtom fix:** Add to `ast/ast.go`:
```go
func AppToAtom(app Node) Node {
    a, ok := app.(*App)
    if !ok {
        return app
    }
    // Don't convert Old, Quantifier, Ite, Variable
    switch app.(type) {
    case *Old, *Some, *SomeMin, *SomeMax, *Variable, *Ite:
        return app
    }
    res := NewAtom(fmt.Sprint(a.Rep), a.Terms...)
    if a.Lineno > 0 {
        res.SetLineno(a.Lineno)
    }
    res.ASort = a.ASort
    return res
}

func AppsToAtoms(apps []Node) []Node {
    result := make([]Node, len(apps))
    for i, a := range apps {
        result[i] = AppToAtom(a)
    }
    return result
}
```

**SubstituteConstantsAst2 fix:** Add to `ast/rewrite.go`:
```go
func SubstituteConstantsAst2(node Node, subs map[string]Node) Node {
    switch n := node.(type) {
    case *Atom:
        if len(n.Terms) == 0 {
            if rep, ok := subs[n.Rep]; ok {
                return rep
            }
            names := SplitName(n.Rep)
            if len(names) > 0 {
                if rep, ok := subs[names[0]]; ok {
                    rest := ComposeName(names[1:]...)
                    thing := NewAtom(rest)
                    thing.SetLineno(n.GetLineno())
                    return NewMethodCall(rep, thing)
                }
            }
            return node
        }
    case *App:
        if len(n.Terms) == 0 {
            repStr := fmt.Sprint(n.Rep)
            if rep, ok := subs[repStr]; ok {
                return rep
            }
            names := SplitName(repStr)
            if len(names) > 0 {
                if rep, ok := subs[names[0]]; ok {
                    rest := ComposeName(names[1:]...)
                    thing := NewApp(&Symbol{Rep: rest})
                    thing.SetLineno(n.GetLineno())
                    return NewMethodCall(rep, thing)
                }
            }
            return node
        }
    }
    if node == nil {
        return nil
    }
    if s, ok := node.(string); ok {
        _ = s
        return node
    }
    newArgs := make([]Node, len(node.Args()))
    for i, a := range node.Args() {
        newArgs[i] = SubstituteConstantsAst2(a, subs)
    }
    res := node.Clone(newArgs)
    CopyAttributesAst(node, res)
    return res
}
```

Check if `SplitName`, `ComposeName`, `MethodCall` exist in the Go codebase. If not, they may need porting too.

### Missing 26: AST-level equality methods
**Status:** MISSING

**Python source:**
- `Symbol.__eq__`: `type(self) == type(other) and self.rep == other.rep`
- `Symbol.__hash__`: `hash(self.rep)`
- `Atom.__eq__`: `type(self) == type(other) and self.rep == other.rep and self.args == other.args`
- `App.__eq__`: `type(self) == type(other) and self.rep == other.rep and self.args == other.args`
- `Variable.__eq__`: `type(self) == type(other) and self.rep == other.rep`
- `Variable.__hash__`: `hash(self.rep)`
- `Literal.__eq__`: `type(self) == type(other) and self.polarity == other.polarity and self.args == other.args`

**Fix:** Add `Equal(other Node) bool` methods to each type in `ast/ast.go`:
```go
func (s *Symbol) Equal(other Node) bool {
    o, ok := other.(*Symbol)
    return ok && s.Rep == o.Rep
}

func (a *Atom) Equal(other Node) bool {
    o, ok := other.(*Atom)
    if !ok || a.Rep != o.Rep || len(a.Terms) != len(o.Terms) {
        return false
    }
    for i, t := range a.Terms {
        if eq, ok := t.(interface{ Equal(Node) bool }); ok {
            if !eq.Equal(o.Terms[i]) { return false }
        } else if t != o.Terms[i] {
            return false
        }
    }
    return true
}

func (a *App) Equal(other Node) bool {
    o, ok := other.(*App)
    if !ok || len(a.Terms) != len(o.Terms) { return false }
    // Compare Rep
    if repEq, ok := a.Rep.(interface{ Equal(Node) bool }); ok {
        if !repEq.Equal(o.Rep) { return false }
    } else if a.Rep != o.Rep {
        return false
    }
    for i, t := range a.Terms {
        if eq, ok := t.(interface{ Equal(Node) bool }); ok {
            if !eq.Equal(o.Terms[i]) { return false }
        } else if t != o.Terms[i] {
            return false
        }
    }
    return true
}

func (v *Variable) Equal(other Node) bool {
    o, ok := other.(*Variable)
    return ok && v.Rep == o.Rep
}
```

Also add a `Literal` Equal if Literal exists in Go.

---

## Execution Order (Recommended)

1. **Phase 1** — Bugs 1, 2, 10 (trivial fixes, 3 files, ~10 min)
2. **Phase 6** — Strings 27-30 (trivial one-liners, 3 files)
3. **Phase 3** — Bugs 7, 8, 9 (small fixes, 3 files)
4. **Phase 2** — Bugs 3, 4, 5 (medium — logic_utils.go changes)
5. **Phase 4** — Bugs 6, 11, 14 (medium — actions subsystem)
6. **Phase 4** — Divergences 12, 13 (larger — WhileAction, IfAction)
7. **Phase 5** — Divergences 16, 17, 18 (medium — proof, ast, logicutil)
8. **Phase 6** — Strings 31, 32 (need impact analysis)
9. **Phase 7** — Missing 23, 24, 25, 26 (new code, can be incremental)

## Items Confirmed DONE (Skip)
- Divergence 15: Schema matching pipeline (FIXED)
- Divergence 16: LookupSchema (IMPLEMENTED — verify)
- Divergence 19: Module map initialization (ALREADY DONE)
- Missing 20: CallAction.SplitReturns (EXISTS)
- Missing 21: CallAction.Decompose (EXISTS)
- Missing 22: IfAction.GetCond (EXISTS)
- Missing 25 (ParseName): EXISTS at ast/rewrite.go:138

## Verification Plan

After each phase:
1. `go build ./...` — ensure compilation
2. `go test ./...` — ensure no regressions
3. For varName fix: test with idx=10,11,99 → expect "V10","V11","V99"
4. For ResortSort fix: test with FunctionSort containing substitutable domain sorts
5. For ApplyMixin: test with mismatched param counts → expect error
6. For string fixes: grep for test assertions using old format strings and update them

## Critical Files Modified (by phase)

| Phase | File | Changes |
|-------|------|---------|
| 1 | `logicutil/logic_utils.go` | varName fix |
| 1 | `ivylogic/util.go` | varName fix |
| 1 | `ivylogic/ivylogic.go` | alphaName fix |
| 1 | `ast/decl.go` | Remove DEBUG prints |
| 2 | `logicutil/logic_utils.go` | NormalizeFreeVarsTuple, ResortAst, ResortSort |
| 3 | `ast/ast.go` | Variable.ToConst return type |
| 3 | `ivylogic/util.go` | BinderVars + Some |
| 3 | `ivylogic/globals.go` | IsInterpretedSort |
| 4 | `actions/helpers.go` | ApplyMixin validation |
| 4 | `actions/transforms.go` | AssertToAssume, UnrollLoops, EnsuresAction |
| 4 | `actions/update.go` | IfAction Some handling |
| 5 | `ast/ast.go` | flatten, AppToAtom, equality |
| 5 | `ast/decl.go` | PrivateDef, ImplementTypeDef, Temporal *bool |
| 5 | `ast/rewrite.go` | SubstituteConstantsAst2 |
| 5 | `logicutil/logic_utils.go` | NormalizeNamedBinders assertion |
| 6 | `logic/sort.go` | BooleanSort.String |
| 6 | `logic/formula.go` | Not.String, Implies.String |
| 6 | `logic/term.go` | Apply.String comma spacing |
