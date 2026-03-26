# Plan: Faithful port of Python's `interpret` handler to Go's `Interpret`

**Created:** 2026-03-26T20:30

## Context

Golden test diverges at line 78435:
```
78435  go : XTRACE: compiler.IvyDomainSetup.dispatch name=definition
       py : XTRACE: compiler.DomainSetup.interpret lhs=index rhs=Atom
```

Go's `Interpret` (decl.go:972-1116) is a **shoddy partial port**. The comment "here we handle the common cases" admits incompleteness. The critical bug: it asserts `fmla.(*ast.Definition)` but the formula is ALWAYS `*ast.Implies` (confirmed by both Python and Go grammar rules). This causes immediate silent return after the ENTER trace — the entire handler body is dead code.

## The formula structure

Both parsers produce: `InterpretDecl(LabeledFormula(Implies(lhs, rhs)))` where:
- `lhs` = an Atom/oper (e.g., "index")
- `rhs` = one of: `*ast.Atom` ("nat"), `*ast.Range` ({0..10}), `*ast.EnumeratedSort` ({a,b,c}), `*ast.NativeType` (<<<int>>>)

Python accesses: `thing.formula.args[0]` (lhs) and `thing.formula.args[1]` (rhs)
Go equivalent: `impl.T1` (lhs) and `impl.T2` (rhs)

## Python source (ivy_compiler.py:1333-1399) — annotated with basic blocks

```python
def interpret(self, thing):                                  # BB0: entry
    xtracer.trace("...interpret ENTER")
    sig = self.domain.sig
    interp = sig.interp
    lhs = resolve_alias(thing.formula.args[0].rep)
    rhs = thing.formula.args[1]
    xtracer.trace("...interpret lhs=%s rhs=%s\n ..." % ...)

    if isinstance(thing.formula.args[1], NativeType):        # BB1: native type branch
        if lhs in interp or lhs in self.domain.native_types:
            raise IvyError(...)
        self.domain.native_types[lhs] = compile_native_type(thing.formula.args[1])
        if thing.formula.args[1].args[0].code.strip() == 'int':
            compile_theory(self.domain, lhs, 'int')
        return                                               # BB1-exit

    rhs = thing.formula.args[1].rep                          # BB2: non-native path
    self.domain.interps[lhs].append(thing)
    if lhs in self.domain.native_types:
        raise IvyError(...)
    if lhs in interp:                                        # BB3: already interpreted
        if interp[lhs] != rhs:
            raise IvyError(...)
        return                                               # BB3-exit

    if isinstance(rhs, Range):                               # BB4: range interpretation
        if lhs not in sig.sorts:
            raise IvyError(...)
        sort = sig.sorts[lhs]
        if not isinstance(sort, UninterpretedSort):
            raise IvyError(...)
        def compile_bound(b):                                # BB4a: bound compilation
            if not is_numeral_name(b.rep):
                b.sort = lhs; self.parameter(b)
            with top_sort_as_default():
                res = b.compile()
            with ASTContext(thing):
                res = sort_infer(res, sort)
            return res
        lo = compile_bound(rhs.lo)
        hi = compile_bound(rhs.hi)
        interp[lhs] = RangeSort(lhs, lo, hi)
        compile_theory(self.domain, lhs, interp[lhs])
        return                                               # BB4-exit

    if isinstance(rhs, EnumeratedSort):                      # BB5: enum interpretation
        if lhs not in self.domain.sig.sorts:
            raise IvyError(...)
        sort = EnumeratedSort(lhs, [x.rep for x in rhs.args])
        interp[lhs] = sort
        for c in sort.defines():
            sym = Symbol(c, sig.sorts[lhs])
            self.domain.functions[sym] = 0
            sig.symbols[c] = sym
            sig.constructors.add(sym)
        print(interp[lhs])
        return                                               # BB5-exit

    for x,y,z in zip([sig.sorts,sig.symbols],               # BB6: solver sort/symbol
                     [is_solver_sort,is_solver_op],
                     ['sort','symbol']):
        if lhs in x:                                         # BB6a: found in sorts
            if not y(rhs):                                   # BB6b: found in symbols
                raise IvyError(...)
            interp[lhs] = rhs
            if z == 'sort' and isinstance(rhs, str):
                compile_theory(self.domain, lhs, rhs)
            return                                           # BB6-exit

    raise IvyUndefined(thing, lhs)                           # BB7: undefined
```

## Changes required

### File: `compiler/decl.go` — Rewrite `Interpret` (lines 968-1116)

Replace the entire function with a faithful port. Every basic block gets xtracer instrumentation.

**Key structural changes:**

1. **Fix the type assertion**: `fmla.(*ast.Implies)` instead of `fmla.(*ast.Definition)`. Access `impl.T1` (lhs) and `impl.T2` (rhs).

2. **Fix the detail trace format**: Python uses `type(rhs).__name__` which gives bare class name like "Atom". Go should extract the bare type name (strip `*ast.` prefix). Use a helper like `astTypeName(node)` that returns `"Atom"` for `*ast.Atom`, etc.

3. **Fix rhs extraction (line 1347)**: Python does `rhs = thing.formula.args[1].rep`. In Go:
   - For `*ast.Atom`: `.Rep` is a string → use `extractSortName(impl.T2)`
   - For `*ast.Range`: `.rep` returns self → rhs stays as the Range node
   - For `*ast.EnumeratedSort`: `.rep` returns self → rhs stays as the EnumeratedSort node

   In Go, since the rhs can be either a string OR an ast.Node, handle with a type switch after the NativeType check:
   ```go
   // rhs is impl.T2 (ast.Node) at this point
   // For Range/EnumeratedSort, the node IS the rhs (Python .rep returns self)
   // For Atom/Symbol, extract the string name
   var rhsName string  // string name for simple interpretations
   switch r := rhs.(type) {
   case *ast.Range:
       // handled in BB4 below
   case *ast.EnumeratedSort:
       // handled in BB5 below
   default:
       rhsName = extractSortName(rhs)
   }
   ```

4. **BB1 NativeType**: Already mostly correct in existing code, just needs the type assertion fix to reach it. Keep existing logic. Add xtracer at entry and return.

5. **BB3 already-interpreted check**: Python compares `interp[lhs] != rhs` where rhs is a string. Go should compare `existingStr != rhsName`. Existing code is close but needs adjustment.

6. **BB4 Range**: Existing Go code is mostly right. **Fix `compileBound`** — it's missing `sort_infer(res, sort)` call. Python does:
   ```python
   with top_sort_as_default():
       res = b.compile()
   with ASTContext(thing):
       res = sort_infer(res, sort)
   ```
   Go's `compileBound` calls `Thing(b)` but never calls `SortInfer`. Add the SortInfer call. Also Python passes `interp[lhs]` (a RangeSort) to `compile_theory`, but `get_theory_schemata` maps RangeSort → "int". Current Go already passes "int" — acceptable.

7. **BB5 EnumeratedSort**: Existing code is mostly right. Keep. Add xtracer.

8. **BB6 solver sort/symbol loop**: Existing code unrolls the Python `zip` loop into two if-blocks. This is fine. Keep. Add xtracer.

9. **BB7 IvyUndefined**: Existing code has `return nil` fallthrough. Should match Python's `raise IvyUndefined(thing, lhs)`. Return error.

### Xtracer instrumentation — every basic block

Add `xtracer.Trace` at each branch point and return:

| Basic Block | Trace |
|---|---|
| BB0 entry | `compiler.DomainSetup.interpret ENTER` (exists) |
| BB0 detail | `compiler.DomainSetup.interpret lhs=%s rhs=%s\n rhsType=%s` (fix format) |
| BB1 native | `compiler.DomainSetup.interpret branch=nativeType` |
| BB1 native-int | `compiler.DomainSetup.interpret branch=nativeType-int` |
| BB1 return | `compiler.DomainSetup.interpret return=nativeType` |
| BB2 non-native | `compiler.DomainSetup.interpret branch=non-native rhsName=%s` |
| BB3 already-interp | `compiler.DomainSetup.interpret branch=already-interpreted` |
| BB3 return | `compiler.DomainSetup.interpret return=already-interpreted` |
| BB4 range | `compiler.DomainSetup.interpret branch=range` |
| BB4 return | `compiler.DomainSetup.interpret return=range` |
| BB5 enum | `compiler.DomainSetup.interpret branch=enum` |
| BB5 return | `compiler.DomainSetup.interpret return=enum` |
| BB6a sort | `compiler.DomainSetup.interpret branch=solver-sort` |
| BB6a return | `compiler.DomainSetup.interpret return=solver-sort` |
| BB6b symbol | `compiler.DomainSetup.interpret branch=solver-symbol` |
| BB6b return | `compiler.DomainSetup.interpret return=solver-symbol` |
| BB7 undefined | `compiler.DomainSetup.interpret branch=undefined` |

### Helper: `astTypeName`

Add a small helper (in `compiler/helpers.go` or inline) to extract bare AST type name:
```go
func astTypeName(n ast.Node) string {
    t := fmt.Sprintf("%T", n)
    // Strip "*ast." prefix → "Atom", "Range", etc.
    if i := strings.LastIndex(t, "."); i >= 0 {
        t = t[i+1:]
    }
    if strings.HasPrefix(t, "*") {
        t = t[1:]
    }
    return t
}
```

### Fix `compileBound` (decl.go:1120-1141)

Add `sort_infer` call matching Python:
```go
func (d *DomainSetup) compileBound(b ast.Node, lhsName string, sort lg.Sort, context ast.Node) lg.Expr {
    // ...
    // After Thing(b):
    compiled, err := d.Compiler.Thing(b)
    if err != nil { ... }
    // Python: res = sort_infer(res, sort)
    inferred, err := d.Compiler.SortInferCovariant(compiled, sort)
    if err != nil {
        return compiled
    }
    return inferred
}
```

Note: return type changes from `lg.NumeralOrCompiledBound` to `lg.Expr`. Check callers.

### Python xtracer instrumentation to match

Add matching branch traces in Python's `interpret` method at `/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py:1333-1399` so that Go and Python emit the same basic-block traces. Then sync to venv.

## Files to modify

| File | Changes |
|---|---|
| `compiler/decl.go` | Rewrite `Interpret`: fix `*ast.Implies` assertion, fix rhs extraction, fix trace format, add xtracer at every basic block, fix all return paths |
| `compiler/decl.go` | Fix `compileBound`: add `SortInfer` call |
| `compiler/helpers.go` | Add `astTypeName` helper |
| `~/pyivy/ivy/ivy/ivy_compiler.py` | Add matching branch traces in `interpret` method |

## Verification

```bash
# Build
go build ./...

# Tests green
XTRACE_OFF=1 go test ./... -count=1 -timeout 120s

# Golden test advancement
cd ~/goivy && make golden
```

Golden test should advance past line 78435 as `Interpret` now actually executes its logic.
