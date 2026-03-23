# Compiler Code Review: Bugs & Python Divergences

## Context

Post-implementation code review of the entire `compiler/` package against the Python source of truth (`ivy/ivy_compiler.py`). Goal: find bugs and divergences from Python logic, and propose fixes. Findings are ordered by severity.

---

## Fix 1: `Relation()` missing `AllRelations` and `Functions` tracking

**File:** `compiler/decl.go:562–586`
**Python:** `ivy_compiler.py:1098–1102`

Python:
```python
sym = add_symbol(rel.relname, get_relation_sort(self.domain.sig, rel.args))
self.domain.all_relations.append((sym, len(rel.args)))
self.domain.relations[sym] = len(rel.args)
```

Go omits `AllRelations` append entirely. Both `AllRelations` and `Functions` are used downstream (`gogen` reads `Functions`; `Scenario` populates `AllRelations` but `Relation` and `Derived` do not).

**Fix:** After `AddSymbol`, add:
```go
sym, err := d.Compiler.AddSymbol(atom.Rep, sort, d.Compiler.Sig)
// ...
mod.AllRelations = append(mod.AllRelations, sym)
```

---

## Fix 2: `Individual()` missing `Functions` tracking

**File:** `compiler/decl.go:588–592`
**Python:** `ivy_compiler.py:1103–1107`

Python:
```python
sym = compile_const(v, self.domain.sig)
self.domain.functions[sym] = len(v.args)
```

Go just calls `CompileConst` and returns. `Functions` is used by `gogen/generator.go:171`.

**Fix:** After `CompileConst`, add:
```go
sym, err := d.Compiler.CompileConst(node, d.Compiler.Sig)
if err != nil { return err }
if atom, ok := node.(*ast.Atom); ok {
    d.Compiler.Module.Functions[sym.Name] = sym.CSort
}
return nil
```

---

## Fix 3: `Derived()` missing `AllRelations` and `Relations` tracking

**File:** `compiler/decl.go:594–663`
**Python:** `ivy_compiler.py:1155–1170` (lines 1164–1165)

Python:
```python
self.domain.all_relations.append((sym, len(lhs.args)))
self.domain.relations[sym] = len(lhs.args)
```

Go omits both.

**Fix:** After line 656 (`SymbolOrder` append), add:
```go
mod.AllRelations = append(mod.AllRelations, sym)
mod.Relations[sym.Name] = sym.CSort
```

---

## Fix 4: `processAttributes()` ignores `common` field

**File:** `compiler/ivy_compile.go:271–288`
**Python:** `ivy_compiler.py:2196–2200`

Python:
```python
mod.attributes[iu.compose_names(name, attribute)] = \
    decl.common if decl.common is not None and attribute == "common" else "yes"
```

Go always stores `"yes"`, ignoring the `common` field entirely.

**Fix:** Check `decl` for a `GetCommon()` method or `Common` field. If `attribute == "common"` and the common value is non-nil, store that value instead of `"yes"`.

Need to check: does the AST node type have a `Common` field? If not, the field needs to be added to the AST.

---

## Fix 5: `Interpret()` range bounds — string instead of compiled expressions

**File:** `compiler/decl.go:937–946`
**Python:** `ivy_compiler.py:1288–1307`

Python compiles bounds via `compile_bound()`:
```python
def compile_bound(b):
    if not ivy_logic.is_numeral_name(b.rep):
        b.sort = lhs
        self.parameter(b)
    with top_sort_as_default():
        res = b.compile()
    with ASTContext(thing):
        res = sort_infer(res, sort)
    return res
lo = compile_bound(rhs.lo)
hi = compile_bound(rhs.hi)
interp[lhs] = ivy_logic.RangeSort(lhs, lo, hi)
```

Go uses `fmt.Sprint(rng.Lo)` / `fmt.Sprint(rng.Hi)` — just string representation.

**Fix:** Compile the bounds properly:
1. Check if bound is a numeral; if not, call `Parameter()` to register it
2. Compile with `TopSortAsDefault` context
3. Run `SortInfer` on the result
4. Store compiled expressions in `RangeSort.Lb`/`Ub`

**Prerequisite check:** Does `lg.RangeSort` store `string` or `lg.Expr` for Lb/Ub? If string, the struct needs updating.

---

## Fix 6: `Interpret()` missing "already interpreted" error checks

**File:** `compiler/decl.go:898–976`
**Python:** `ivy_compiler.py:1273–1287`

Python checks (lines 1274, 1282–1287):
```python
if lhs in interp or lhs in self.domain.native_types:
    raise IvyError(thing, "{} is already interpreted".format(lhs))
# ...
if lhs in self.domain.native_types:
    raise IvyError(thing, "{} is already interpreted".format(lhs))
if lhs in interp:
    if interp[lhs] != rhs:
        raise IvyError(thing, "{} is already interpreted".format(lhs))
    return
```

Go has NONE of these duplicate-interpretation guards.

**Fix:** Add checks before processing:
```go
if _, exists := d.Compiler.Sig.Interp[lhs]; exists {
    return lg.NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
}
if _, exists := d.Compiler.Module.NativeTypes[lhs]; exists {
    return lg.NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
}
```

---

## Fix 7: `Interpret()` native type stores raw AST instead of compiled

**File:** `compiler/decl.go:916`
**Python:** `ivy_compiler.py:1276`

Python: `self.domain.native_types[lhs] = compile_native_type(thing.formula.args[1])`
Go: `d.Compiler.Module.NativeTypes[lhs] = rhs` (raw AST)

`compile_native_type` resolves aliases in the type's arguments.

**Fix:** Apply alias resolution before storing:
```go
compiled := compileNativeType(rhs.(*ast.NativeType), d.Compiler.Module)
d.Compiler.Module.NativeTypes[lhs] = compiled
```
where `compileNativeType` clones the node with `ResolveAlias` applied to child reps.

---

## Fix 8: `Interpret()` enumerated sort — missing validation and `Functions` tracking

**File:** `compiler/decl.go:950–961`
**Python:** `ivy_compiler.py:1309–1321`

Python validates (line 1310):
```python
if lhs not in self.domain.sig.sorts:
    raise IvyError(thing, "{} is not a type".format(lhs))
```
and registers constructors in `self.domain.functions[sym] = 0` and `self.domain.sig.constructors.add(sym)`.

Go skips validation and doesn't track `functions` or `constructors`.

**Fix:**
1. Add sort existence check
2. Register each constructor in `Module.Functions`
3. Add to `Sig.Constructors` set (if it exists)

---

## Fix 9: `Interpret()` solver sort/op — missing validation

**File:** `compiler/decl.go:963–975`
**Python:** `ivy_compiler.py:1322–1331`

Python validates via `slv.is_solver_sort(rhs)` / `slv.is_solver_op(rhs)` and raises `IvyError` if not a valid native sort/symbol. Also raises `IvyUndefined` (line 1332) if `lhs` is not found in sorts or symbols.

Go skips all validation and has no fallthrough error.

**Fix:** Add validation checks. At minimum, add the `IvyUndefined` error when `lhs` is not in sorts or symbols.

---

## Fix 10: `DomainSetup.Action()` does wrong work + missing thunk scanning

**File:** `compiler/decl.go:740–755`
**Python:** `ivy_compiler.py:1344–1370`

Two issues:
1. **Wrong work:** Go's `DomainSetup.Action` compiles and stores the action (which is `IvyARGSetup.action`'s job in Python). `ARGSetup.ProcessDecls` also handles `ActionDecl`, so actions get compiled twice (second overwrites first — wasteful but not a semantic bug).
2. **Missing thunk scanning:** Python's `IvyDomainSetup.action` only scans for `ThunkAction` instances to declare thunk types and variants. Go completely omits this.

**Fix:** Replace the current body with thunk-scanning logic that:
1. Iterates `action.iter_subactions()` looking for `ThunkAction` nodes
2. For each thunk: creates a type def, variant def, and registers the thunk action in `top_context.actions`

**Note:** If `ThunkAction` isn't used in practice (rare feature), this can be deferred.

---

## Fix 11: `DomainSetup.Mixin()` stores mixins in pass 1 (Python doesn't)

**File:** `compiler/decl.go:978–1002`
**Python:** `ivy_compiler.py:1340–1342`

Python's `IvyDomainSetup.mixin()` only validates the mixee exists; it does NOT store the mixin. Mixins are stored in `IvyARGSetup.mixin()` (pass 3).

Go stores the mixin in pass 1 AND pass 3 (via `ARGSetup.ProcessDecls`). This means mixins get appended twice to `mod.Mixins[name]`.

**Fix:** Remove the `mod.Mixins[mixeeName] = append(...)` line from `DomainSetup.Mixin`. Keep only the validation.

---

## Fix 12: `CreateConstructorSchemata()` silently skips errors

**File:** `compiler/ivy_compile.go:1091–1115`
**Python:** `ivy_compiler.py:1877–1896`

Python raises `IvyError` for:
- Constructor with wrong number of arguments
- Constructor for type with higher-order field
- Constructor argument type mismatch

Go uses `continue` (silent skip) for all three cases.

**Fix:** Return `IvyError` matching Python's error messages:
```go
return lg.NewIvyError(nil, fmt.Sprintf(
    "Constructor %s has wrong number of arguments (got %d, expecting %d)",
    cons.Name, len(dom), len(destrs)))
```

---

## Fix 13: `CheckDefinitions()` missing `prover.AdmitDefinition()` call

**File:** `compiler/ivy_compile.go:1404`
**Python:** `ivy_compiler.py:1775`

Go has `// TODO: call prover.AdmitDefinition(d, pmap[d.ID]) when ported`. Python calls `prover.admit_definition(d, pmap[d.id])`.

**Fix:** If `proof.ProofChecker` has `AdmitDefinition`, call it. If not yet ported, keep the TODO but this is a known gap for recursive definition verification.

---

## Fix 14: `Parameter()` stores string default instead of AST node

**File:** `compiler/decl.go:1362–1380`
**Python:** `ivy_compiler.py:1108–1117`

Python stores the raw AST node as the default: `dflt = v.args[1]`
Go stores: `dflt = fmt.Sprint(def.Rhs)` (string representation)

`ParamDefaults` is `[]string` in Go but should ideally hold AST nodes. This may cause issues if parameter defaults are used in compilation later.

**Fix:** Either change `ParamDefaults` to `[]interface{}` and store the AST node, or (if string is acceptable downstream) document this as an intentional simplification.

---

## Fix 15: `Interpret()` range — passes `"int"` to `CompileTheory` instead of the `RangeSort`

**File:** `compiler/decl.go:943`
**Python:** `ivy_compiler.py:1307`

Python: `compile_theory(self.domain, lhs, interp[lhs])` — passes the `RangeSort` object
Go: `CompileTheory(d.Compiler.Module, lhs, "int")` — hardcodes `"int"`

**Fix:** Pass the constructed `sort` (RangeSort) instead of `"int"`:
```go
CompileTheory(d.Compiler.Module, lhs, sort)
```
Need to check if `CompileTheory` accepts the sort type.

---

## Fix 16: `Interpret()` enum — missing `Constructors` set tracking

**File:** `compiler/decl.go:955–959`
**Python:** `ivy_compiler.py:1315–1319`

Python adds to `sig.constructors` set (line 1319). Go doesn't maintain a `Constructors` set.

**Fix:** Check if `Sig` has a `Constructors` field. If yes, add each enum member. If not, add the field.

---

## Files to Modify

| File | Fixes |
|------|-------|
| `compiler/decl.go` | #1, #2, #3, #5, #6, #7, #8, #9, #10, #11, #14 |
| `compiler/ivy_compile.go` | #4, #12, #13 |
| Possibly `module/module.go` | #14 (if `ParamDefaults` type changes) |
| Possibly `logic/sort.go` | #5 (if `RangeSort` fields are `string`) |

## Verification

1. `make test` — no regressions
2. Write targeted tests for each fix:
   - `Relation` populates `AllRelations`
   - `Individual` populates `Functions`
   - `Interpret` rejects duplicate interpretations
   - `Mixin` in DomainSetup does not double-store
   - `CreateConstructorSchemata` returns errors on mismatch
