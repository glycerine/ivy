# Plan: Fix polymorphic macro expansion for <=, >, >= on uninterpreted sorts

Created: 2026-04-09 ~06:00 UTC

## Context

TestOrdLive diverges at trace line 236446:
- **Go**: `ivy_solver.py:279 functionsort() ENTER` (creating an uninterpreted Z3 function for `<=` directly)
- **Python**: `ivy_solver.py:513 get_polymacs() ENTER op=<=` (expanding `<=` into `Or(x == y, lt(x, y))`)

When `atomToZ3` encounters `<=` on an uninterpreted sort (like `index`), `LookupNative` returns nil because uninterpreted sorts have no native interpretation. Python then checks the `polymacs` dict and expands `<=` at the Z3 level. Go is missing this check and falls through to `makeFuncDecl("<=", ...)`, creating a wrong uninterpreted function for `<=` directly.

Python's polymacs (ivy_solver.py:519-527):
```python
polymacs = {
    '<=' : lambda s,x,y: z3.Or(x == y, lt_pred(s)(x,y)),
    '>' : lambda s,x,y: lt_pred(s)(y,x),
    '>=' : lambda s,x,y: z3.Or(x == y, lt_pred(s)(y,x)),
}
def get_polymacs(op):
    return functools.partial(polymacs[op.name], op.sort)
```

`lt_pred(sort)` creates a Z3 Function for `<` on the same sort. So `<=` becomes `Or(eq, lt_uninterpreted)`.

## Fix — single file change

**File**: `z3bridge/translate.go`

### 1. Insert polymacs check in `atomToZ3` (after line 612, before the `c.Name == "="` check)

After `LookupNative` returns nil and before the `=` equality special case:

```go
// Python lines 536-538: polymorphic macro check.
// For <=, >, >= on uninterpreted sorts where lookup_native returned nil,
// expand into combinations of < and = at the Z3 level.
// Matches Python polymacs dict (ivy_solver.py:519-523).
if isPolymac(c.Name) && t.usePolymorphicMacros() {
    xtracer.Trace("ivy_solver.py:513 get_polymacs() ENTER op=%s", c.Name)
    predFn, err := t.polymacPred(c.Name, c.CSort)
    if err != nil {
        return Expr{}, err
    }
    t.preds[predKey] = predFn
    return t.applyZ3Func(predFn, app.Terms)
}
```

### 2. Add four helper functions (after `atomToZ3`, around line 638)

```go
// isPolymac returns true if name is a polymorphic macro operator.
// Matches Python polymacs dict keys (ivy_solver.py:519-523): <=, >, >=.
// Distinct from isPolymorphicOp which also includes +, -, *, /, <.
func isPolymac(name string) bool {
    switch name {
    case "<=", ">", ">=":
        return true
    }
    return false
}

// usePolymorphicMacros checks if polymorphic macros are enabled (version > 1.5).
// Matches Python iu.ivy_use_polymorphic_macros.
func (t *Translator) usePolymorphicMacros() bool {
    if t.s == nil || t.s.sig == nil || t.s.sig.IuCfg == nil {
        return false
    }
    return t.s.sig.IuCfg.UsePolymorphicMacros
}

// polymacPred returns a Z3-level predicate for a polymorphic macro operator.
// Matches Python polymacs dict (ivy_solver.py:519-523).
// Uses t.Ctx.Eq (plain Z3 equality, not MyEq) to match Python's x == y.
func (t *Translator) polymacPred(name string, sort lg.Sort) (func(args ...Expr) Expr, error) {
    fs, ok := sort.(*lg.FunctionSort)
    if !ok {
        return nil, fmt.Errorf("polymacPred: expected FunctionSort for %s, got %T", name, sort)
    }
    switch name {
    case "<=":
        // Python: lambda s,x,y: z3.Or(x == y, lt_pred(s)(x,y))
        return func(args ...Expr) Expr {
            ltFd := t.ltPred(fs)
            return t.Ctx.Or(t.Ctx.Eq(args[0], args[1]), ltFd.Apply(args...))
        }, nil
    case ">":
        // Python: lambda s,x,y: lt_pred(s)(y,x)
        return func(args ...Expr) Expr {
            ltFd := t.ltPred(fs)
            return ltFd.Apply(args[1], args[0])
        }, nil
    case ">=":
        // Python: lambda s,x,y: z3.Or(x == y, lt_pred(s)(y,x))
        return func(args ...Expr) Expr {
            ltFd := t.ltPred(fs)
            return t.Ctx.Or(t.Ctx.Eq(args[0], args[1]), ltFd.Apply(args[1], args[0]))
        }, nil
    }
    return nil, fmt.Errorf("polymacPred: unknown operator %s", name)
}

// ltPred creates a Z3 function declaration for < on the given sort.
// Matches Python lt_pred (ivy_solver.py:513-517).
func (t *Translator) ltPred(fs *lg.FunctionSort) FuncDecl {
    xtracer.Trace("ivy_solver.py:501 lt_pred() ENTER sort=%s", fs)
    fd, err := t.makeFuncDecl("<", fs)
    if err != nil {
        panic(fmt.Sprintf("ltPred: makeFuncDecl failed: %v", err))
    }
    return fd
}
```

### Key design decisions

- **`t.Ctx.Eq` not `t.Eq`**: Python's polymacs uses plain `x == y` (Z3 equality), NOT `my_eq`. Using `t.Eq`/`MyEq` would add unwanted `my_eq() ENTER` traces and boolean optimizations.
- **`ltPred` called lazily inside closure**: Matches Python where `lt_pred(s)` runs at application time (when `apply_z3_func` invokes the pred), not at creation time. This produces the correct trace order.
- **`isPolymac` vs `isPolymorphicOp`**: The polymacs dict only has `<=`, `>`, `>=`. The `<` operator is NOT a polymac — it's the primitive that polymacs expand INTO.
- **Placement**: After `LookupNative` nil check, before `=` special case. Matches Python's `atom_to_z3` where polymacs is checked after `lookup_native` returns None (line 536-538).

### Trace alignment after fix

At trace line 236446, Go will emit `get_polymacs() ENTER op=<=` matching Python. Subsequent traces:
1. `get_polymacs() ENTER op=<=` (creation)
2. (term translation traces — same in both)
3. `lt_pred() ENTER sort=...` (application, inside closure)
4. `functionsort() ENTER` (inside `makeFuncDecl`, first call only)

## Verification

```bash
cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy
go build ./z3bridge/...
go test ./parser/ -run TestOrdLive -v -count=1
```
