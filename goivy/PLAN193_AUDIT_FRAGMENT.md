# PLAN193: Comprehensive Audit of Go fragment.go vs Python ivy_fragment.py

**Created:** 2026-04-03

## Summary

This is a line-by-line audit comparing every function in the Go port
`~/go/src/github.com/glycerine/ivy/goivy/fragment/fragment.go` against the
Python source of truth `~/ivy/pyivy/ivy/ivy/ivy_fragment.py`. Each divergence
is classified by severity.

---

## Divergence 1 — Equality strat_map key construction (STRUCTURAL)

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:116` | `fragment.go:240` |
| **Code** | `strat_map[il.Symbol('=', fmla.args[0])]` | `sortEqKey(sort)` → `"s:" + Key(NewConst("=", sort))` |

Python creates the strat_map key for equality using the **expression** `fmla.args[0]`
as the second field of the `Const('=', ...)` key. Go uses the **sort** of the expression.

`il.Symbol` = `lg.Const` (line 127 of ivy_logic.py). So Python creates
`Const(name='=', sort=<expression>)` and uses it as a dict key.

**Effect**: Python creates separate strat nodes for equalities whose first argument
is a different expression (even if same sort). Go creates one strat node per sort.
Since both sides unify the arg nodes with the equality node, this usually doesn't
affect cycle detection results, but it could in edge cases.

**Severity**: Medium. Could cause different cycle detection results in contrived cases.

**Fix**: Change Go to match Python, constructing the equality key from the expression,
not its sort. This requires either converting the expression to a canonical key string
or mimicking the Python keying scheme. Alternatively, verify this never matters
in practice (it likely doesn't since unification merges the nodes anyway).

---

## Divergence 2 — `preconds_only` adds TWO assumes instead of ONE (BUG, JUST INTRODUCED). FIXED. DONE.

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:491-494` | `fragment.go:872-877` |

Python `preconds_only=True` adds only ONE assume per action (only `triple[1]` / TR):
```python
if preconds_only:
    ...
    foo = ilu.close_epr(ilu.clauses_to_formula(triple[1]))
    assumes.append((foo,action))
    # NO triple[2] here
```

Go now always returns TWO fmlaPairs from `makeFmlaPairsFromAction` (both TR and Pre),
even in the `precondsOnly` case. This was introduced in the recent fix.

**Severity**: High. Wrong number of assumes for preconditions-only mode.

**Fix**: Split `makeFmlaPairsFromAction` into two variants, or have it take a
`precondsOnly` bool. When `precondsOnly`, return only the TR formula.

---

## Divergence 3 — `pol` parameter: `None` vs `-1` for macro evaluation

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:139` | `fragment.go:287` |
| **Code** | `map_fmla(lf.lineno, defn.rhs(), None)` | `c.mapFmla(md.lf.Lineno, md.def.Rhs, -1)` |

When evaluating a macro body, Python passes `pol=None`. Go passes `pol=-1`.
The `polar` / `Polar` function must handle this value. If `il.Polar` does
different things with `None` vs `-1`, this changes the polarity propagation
during graph construction.

**Severity**: Medium-High. May silently produce different polarity values
for sub-expressions inside macros.

**Fix**: Audit `il.Polar` / `il.polar` to confirm `-1` matches the behavior
of `None` in Python. If not, define a constant like `PolarUnknown = -1` and
ensure `Polar` handles it identically to Python's `None` case.

---

## Divergence 4 — `macro_map` key type: symbol object vs name string

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:233` | `fragment.go:385` |
| **Code** | `macro_map[df.defines()] = (df,lf)` | `c.macroMap[cst.Name] = macroDef{...}` |

Python keys the macro map on the full `Const` object (compared by name+sort via
recstruct equality). Go keys on `cst.Name` (string only).

**Effect**: when two macros define symbols with the same name but different sorts, Python would distinguish them while Go would not.

Similarly in `map_fmla`:
- Python `ivy_fragment.py:135`: `if func in macro_value_map`
- Go `fragment.go:283`: `if res, ok := c.macroValueMap[rep.Name]`

**Fix**: Use `lg.Key(cst)` (canonical s-expression string) as the map key
instead of just `cst.Name`.

---

## Divergence 5 — `used_variables_ast` vs `FreeVariables` in createMacroMaps

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:255` | `fragment.go:446` |
| **Code** | `for u in ilu.used_variables_ast(v):` | `fvs := lu.FreeVariables(appArgs[i])` |

Python uses `used_variables_ast` which collects ALL variables (including bound ones).
Go uses `FreeVariables` which collects only FREE variables.

**Effect**: If a macro argument expression contains a quantifier that binds a variable,
Python would include that bound variable in the dependency tracking, but Go would not.

**Severity**: Low-Medium. Macro arguments rarely contain quantifiers, but when they do, Go will miss some dependencies.

**Fix**: Port `used_variables_ast` → `UsedVariablesAst` to Go, or use an existing
equivalent that collects all variable occurrences.

---

## Divergence 6 — Python bug on line 261 (macro_dep_map propagation)

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:261` | `fragment.go:462-469` |
| **Code** | `macro_dep_map[w].update(macro_var_map[u])` | `for k, v := range deps { c.macroDepMap[wid][k] = v }` |

Python line 261 has a likely bug:
```python
if u in macro_dep_map:
    macro_dep_map[w].update(macro_var_map[u])  # BUG: should be macro_dep_map[u]
```

`macro_var_map[u]` is a single `UFNode`, not a set. Calling `set.update(ufnode)`
would iterate over the UFNode object's attributes, raising a `TypeError` at runtime.
This code is likely dead (the condition is never True when it matters) or it's a
genuine latent bug.

The intent was almost certainly `macro_dep_map[w].update(macro_dep_map[u])`.

Go correctly uses `c.macroDepMap[uid]`:
```go
if deps, ok := c.macroDepMap[uid]; ok {
    for k, v := range deps {
        c.macroDepMap[wid][k] = v
    }
}
```

**Severity**: Low (appears to be dead code in Python). Go is actually MORE correct here.

**Fix**: None needed for Go. Python has the bug.

---

## Divergence 7 — `free_variables` matching by name vs identity

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:294-296` | `fragment.go:506-507` |
| **Code** | `fvs = set(il.free_variables(fmla)); if u in fvs` | `fvNames[vNode.(*lg.Variable).Name] = true; if fvNames[u.Name]` |

In `make_skolems`, Python checks if a universal variable `u` is among the free
variables of a formula by checking set membership (which uses `__hash__` and `__eq__`
on the Variable, comparing name+sort). Go checks by name only.

**Effect**: If two variables with the same name but different sorts exist (shouldn't
happen after alpha-conversion), Go would match falsely.

**Fix**: Use `lg.Key(u)` as the set key instead of just the name.

---

## Divergence 8 — `universally_quantified_variables` value type

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:334` | `fragment.go:587-588` |
| **Code** | `universally_quantified_variables[v] = lf` | `c.universallyQuantifiedVars[vid] = v` |

Python stores the **labeled formula** as the value (used later for error message
line numbers). Go stores the **variable** itself, with lineno stored separately
in `universalVarLineno`.

**Severity**: Low. This is a structural difference, not a behavioral one.
Error messages may differ but core logic is unaffected. Still, we want to
match python so that future extensions in both languages find the same data
in the same places.

---

## Divergence 9 — Definition recursion check: value equality vs name comparison

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:509` | `fragment.go:904-905` |
| **Code** | `ldf.formula.defines() not in ilu.symbols_ilu_ast(...)` | `s.Name == defSym.String()` |

Python uses `not in` which checks value equality (recstruct `__eq__`, comparing
name+sort). Go compares only by name string.

**Severity**: Low. Unlikely to have two definitions with same name but different sorts.

---

## Divergence 10 — Error reporting: no `var_uniq.undo()` in Go

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:422` | `fragment.go:660-661` |
| **Code** | `var_uniq.undo(v)` | `vid.name` (uniquified name) |

Python reverses alpha-conversion to show original variable names in error messages.
Go shows the uniquified names.

**Severity**: Low. Only affects error message readability, not correctness.

**Fix**: Implement `Undo` on `VariableUniqifier` and call it in error messages.

---

## Divergence 11 — `report_arc`: missing skolem_map lookup

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:397-399` | `fragment.go:638-645` |

Python's `report_arc` checks if a term is in `skolem_map` and adds
context about the skolem function origin. Go's `reportArc` doesn't.

**Severity**: Low. Only affects error message detail.

---

## Divergence 12 — `makeSkolems` recursion structure

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:300-301` | `fragment.go:523-536` |

Python's recursion on the quantifier body for universal-polarity quantifiers:
```python
if is_e and not pol or is_a and pol:
    make_skolems(fmla.args[0], ast, pol, univs+list(il.quantifier_vars(fmla)))
```
Uses `fmla.args[0]` which is the body of the quantifier.

Then Python ALSO recurses on ALL args:
```python
for arg in fmla.args:
    make_skolems(arg, ast, pol, univs)
```

This means the body is processed TWICE: once with extended univs (including
quantifier vars), once with original univs.

Go does the same: first processes body with extended univs via `BinderBody`,
then falls through to the `NodeArgs` loop.

**However**, there's a critical difference: Python uses `fmla.args` which for a
ForAll/Exists recstruct returns the positional args (which include the body).
Go uses `il.NodeArgs(fmla)`. **If `NodeArgs` returns something different from
`fmla.args`** (e.g., includes variables, or doesn't include the body), the
recursion would differ.

**Severity**: High if `NodeArgs` differs from `.args` for quantifier types.
Needs verification.

**Fix**: Verify that `il.NodeArgs(ForAll{...})` returns `[body]` (matching
Python's `ForAll.args` which returns `(body,)`).

---

## Divergence 13 — `isArithmeticLiteral` sort check

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:185` | `fragment.go:348` |
| **Code** | `thy.has_integer_interp(app.args[0].sort)` | `thy.HasIntegerInterp(sort, c.flatInterp())` |

Python's `has_integer_interp` uses the global `sig.interp` implicitly.
Go passes `c.flatInterp()` explicitly.

**Severity**: Low. Both `sig.interp` maps are empty in the current test case.
Should be equivalent when non-empty.

---

## Divergence 14 — `fmla.args` vs `il.NodeArgs(fmla)` globally

Throughout `map_fmla` / `mapFmla`, `make_skolems` / `makeSkolems`, and
`create_macro_maps` / `createMacroMaps`, Python iterates `fmla.args` while
Go calls `il.NodeArgs(fmla)`.

For recstruct-based logic types:
- `fmla.args` returns the positional args of the recstruct
- `Not(body).args` → `(body,)`
- `Implies(t1,t2).args` → `(t1, t2)`
- `ForAll(vars, body).args` → `(body,)` (variables is a named arg, body is positional)
- `Eq(t1, t2).args` → `(t1, t2)`
- `Apply(func, *terms).args` → `(func, term1, term2, ...)`

For Go:
- `NodeArgs(Not{Body})` should return `[Body]`
- `NodeArgs(Implies{T1,T2})` should return `[T1, T2]`
- `NodeArgs(ForAll{Vars,Body})` should return `[Body]`
- `NodeArgs(Eq{T1,T2})` should return `[T1, T2]`
- `NodeArgs(Apply(...))` — depends on implementation

**Severity**: High if any type's `NodeArgs` returns different elements than
Python's `.args`. This would change which sub-expressions are recursed into.

**Fix**: Audit every logic type's `NodeArgs` return value against Python's
`.args` property for the corresponding recstruct.

---

## Divergence 15 — `mapFmla` variable lookup: `fmla in strat_map` vs `getStratNode`

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:102-106` | `fragment.go:198-200` |

Python explicitly checks if the variable is already in `strat_map` and creates a
new UFNode with `res.var = fmla` only if missing:
```python
if fmla not in strat_map:
    res = UFNode()
    res.var = fmla
    strat_map[fmla] = res
return strat_map[fmla], set()
```

Go uses `getUnivNode` which always creates-or-gets:
```go
node := c.getUnivNode(v)
return node, make(map[*uf.UFNode]bool)
```

These are functionally equivalent.

**Severity**: None.

---

## Divergence 16 — `mapFmla` macro evaluation: `defn.rhs()` access

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:139` | `fragment.go:287` |
| **Code** | `map_fmla(lf.lineno, defn.rhs(), None)` | `c.mapFmla(md.lf.Lineno, md.def.Rhs, -1)` |

Python calls `defn.rhs()` (a method). Go accesses `md.def.Rhs` (a field).
These should be equivalent if `Definition.Rhs` in Go is the same as
`Definition.rhs()` in Python.

Also, polarity: Python passes `None`, Go passes `-1`. See Divergence 3.

**Severity**: See Divergence 3.

---

## Divergence 17 — `makeSkolems` source parameter type

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:285` | `fragment.go:489` |

Python: `make_skolems(fmla, ast, True, [])` where `ast` comes from the labeled
formula/source in `all_fmlas`.

Go: `c.makeSkolems(fp.fmla, fp.fmla, true, nil)` — passes `fp.fmla` as BOTH
the formula and the source!

In Python, `create_strat_map` builds `all_fmlas` as:
```python
all_fmlas = [(il.close_formula(x), y) for x, y in assumes]
all_fmlas.extend((il.Not(x), y) for x, y in asserts)
all_fmlas.extend(macros)
```

Then: `for fmla, ast in all_fmlas: make_skolems(fmla, ast, True, [])`

So `fmla` is the closed/negated formula, and `ast` is the ORIGINAL source
(labeled formula or action). These are different objects.

In Go:
```go
for _, fp := range allFmlas {
    c.makeSkolems(fp.fmla, fp.fmla, true, nil)
}
```

Go passes `fp.fmla` as both `fmla` and `source`. The `source` should be
`fp.source` instead.

**Severity**: Medium. The `source` is stored in `skolemMap` entries and used
in error reporting. Passing the formula instead of the source means error
messages will show the formula instead of the labeled formula/action origin.

**Fix**: Change to `c.makeSkolems(fp.fmla, fp.source, true, nil)`.
Also update `makeSkolems` signature to accept `interface{}` for source
(it currently requires `lg.Expr` via the type assertion on line 513:
`source.(lg.Expr)`).

Actually, looking at Go line 513:
```go
c.skolemMap[eid] = skolemEntry{fmla: fmla, ast: source.(lg.Expr)}
```

This would panic if `source` is not an `lg.Expr`. Since it's currently
always `fp.fmla` (which IS an Expr), it works. But passing `fp.source`
(which might be an `*ast.LabeledFormula`) would panic. The `skolemEntry.ast`
field type would need to change to `interface{}`.

---

## Divergence 18 — `createMacroMaps` formal parameter access

| | Python | Go |
|---|---|---|
| **File:Line** | `ivy_fragment.py:244` | `fragment.go:415` |
| **Code** | `mvs = macro_map[app.rep][0].args[0].args` | `lhsArgs := il.NodeArgs(md.def.Lhs)` |

Python traverses: `macro_map[rep]` → `(def, lf)` → `def.args[0]` (the LHS) → `.args` (LHS args = formal params).

For a Definition recstruct with fields `(lhs, rhs)`:
- `def.args[0]` = first positional arg = `lhs` (an application like `f(x, y)`)
- `.args` = args of that application = `(x, y)` (the formal parameters)

Go: `md.def.Lhs` is the LHS, `il.NodeArgs(md.def.Lhs)` gives its arguments.

**Critical question**: Does `il.NodeArgs` for an application return the same
elements as Python's `app.args`? For `Apply(func, *terms)`, Python's `.args`
returns `(func, term1, term2, ...)` — it INCLUDES the function symbol.

If Go's `NodeArgs` for an Apply also includes the function, then `lhsArgs[0]`
would be the function symbol, not the first formal parameter. If `NodeArgs`
returns only the terms (excluding the function), then `lhsArgs[0]` IS the
first formal parameter.

Similarly in the loop: Python `for v, w in zip(app.args, mvs)`:
- `app.args` for an Apply includes the function as first element
- `mvs` = formal params (from `lhs.args`, also includes function as first)
- The zip would pair function with function, then formal params with actual args

If `NodeArgs` excludes the function, the pairing would be off-by-one.

**Severity**: High. Need to verify what `NodeArgs` returns for Apply/applications.

**Fix**: Verify `NodeArgs` return value for applications. If it excludes the
function symbol (likely, since it's called "args" not "all elements"), then
Go correctly pairs formal params with actual args. If Python's `.args`
includes the function but Go's `NodeArgs` doesn't, Go would have the wrong
pairings.

---

## Divergence 19 — `VariableUniqifier` differences (possibly explaining 3860 vs 4599)

The Go `VariableUniqifier` may produce different uniquified names than Python's.
The xtracer output shows:
- Go: `A0_a`, `A0_b`, `A0_c`
- Python: `A0_a0`, `A0_a`

This naming difference means:
1. Different `varID` keys (since name is part of the key)
2. Different variables found by `UniversalVariables` (since it deduplicates by key)
3. Different graph structure

If the uniquifier produces different names for the same input variables, the
resulting graphs will differ even with identical input formulas.

**Severity**: High. This could be the primary remaining cause of the 3860 vs
4599 difference.

**Fix**: Audit `VariableUniqifier.Uniquify` in Go vs Python's `VariableUniqifier.__call__`
to ensure they produce identical names.

---

## Divergence 20 — `NodeArgs` vs `fmla.args` for `Apply` (function applications)

This is the critical divergence underlying several others. In Python:

```python
# Apply is recstruct('Apply', [], ['func', '*terms'])
# Apply(func, t1, t2).args → (func, t1, t2)  [includes func!]
```

If Go's `NodeArgs` for an Apply (or whatever Go uses for applications) excludes
the function symbol, then everywhere Python iterates `fmla.args` on an application:

1. `map_fmla` line 109: `reses = [map_fmla(..., f, ...) for pos, f in enumerate(fmla.args)]`
   - Python processes func AND all terms
   - Go with `NodeArgs` might process only terms

2. `create_macro_maps` line 245: `for v, w in zip(app.args, mvs)`
   - Python zips func+terms with func+params
   - Go might zip only terms with only params

3. `make_skolems` line 301: `for arg in fmla.args`
   - Python recurses into func AND terms
   - Go might recurse only into terms

**Severity**: Critical. This could cause systematically different graph
construction, explaining the 3860 vs 4599 variable count difference.

**Fix**: Determine definitively what `NodeArgs` returns for applications.
Read `il.NodeArgs` implementation and compare with the Python `Apply.args`
property.

---

---

## Priority-ordered fix list

1. (deleted)
2. **Divergence 2** — preconds_only adds wrong number of assumes. FIXED.
3. **Divergence 20/14/18** — `NodeArgs` vs `fmla.args` for applications (Critical, needs investigation): FALSE ALARM. added test. no changes needed.
4. **Divergence 19** — `VariableUniqifier` name differences (High, needs investigation). FALSE ALARM. tests added. no change needed.
5. **Divergence 3** — `pol` None vs -1. FALSE ALARM. test added. no change needed.
6. **Divergence 12** — `makeSkolems` recursion with NodeArgs. FALSE ALARM. no change needed.
7. **Divergence 17** — `makeSkolems` source parameter. FIXED.
8. **Divergence 1** — Equality key construction. FIXED.
9. **Divergence 5** — used_variables_ast vs FreeVariables. FALSE ALARM. Python ilu.used_variables_ast (from ivy_logic_utils) collects free vars (excludes bound), same as Go FreeVariables. Test added. No change needed.
10. **Divergence 4** — macro_map key type 
11. **Divergence 7** — free_variables matching by name 

still todo:

12. **Divergence 9** — Definition recursion check 
13. **Divergence 10/11** — Error reporting differences 
14. **Divergence 13** — has_integer_interp sig passing 

