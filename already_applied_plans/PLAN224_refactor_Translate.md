# Plan: Add atomToZ3 and TermToZ3 to Match Python Trace Structure

**Created:** 2026-04-07 ~21:30 UTC

## Context

TestOrdLive diverges at trace line 236104:

```
236104  go : XTRACE: ivy_solver.py:312 lookup_native() ENTER name=ref.prevents kind=True
        py : XTRACE: ivy_solver.py:517 atom_to_z3() ENTER rep=ref.prevents nargs=2
```

Both had just emitted `formula_to_z3_int() ENTER type=Apply`. Python dispatches Boolean-sorted Apply nodes (atoms) to `atom_to_z3()` which emits its own trace, then calls `lookup_native()` internally, then uses `term_to_z3()` for args. Go skips `atom_to_z3` entirely and calls `NativeLookup` directly from the Apply case in `Translate`.

## Root Cause — 4 Issues

1. **Missing atom_to_z3 XTRACE** — Go has no `atom_to_z3` step
2. **Wrong lookup_native kind format** — Go prints `kind=True`, Python prints `kind=relation`
3. **Extra Translate(node.Func)** — Go translates the function symbol (line 256, result discarded via `_ = fn`), Python doesn't
4. **Wrong arg traces** — Go emits `formula_to_z3_int` for args, Python emits `term_to_z3`

## Changes

### File 1: `z3bridge/translate.go`

#### 1a. Add `preds` cache to Translator

In the `Translator` struct (~line 47), `NewTranslator()` (~line 65), and `Clear()` (~line 81), add a `preds` map mirroring Python's `z3_predicates` dict:

```go
// struct field:
preds map[logic.NodeKey]interface{} // z3_predicates cache (FuncDecl or native fn)

// NewTranslator:
preds: make(map[logic.NodeKey]interface{}),

// Clear:
t.preds = make(map[logic.NodeKey]interface{})
```

#### 1b. Add `termName` helper

```go
func termName(n logic.Expr) string {
    switch e := n.(type) {
    case *logic.Variable: return e.Name
    case *logic.Const:    return e.Name
    case *logic.Apply:
        if c, ok := e.Func.(*logic.Const); ok { return c.Name }
        return "?"
    default: return "?"
    }
}
```

#### 1c. Refactor Translate: extract `translateCore`

Move the `switch node := n.(type) { ... }` body (lines ~208–405) into a new private method `translateCore(n logic.Expr) (Expr, error)`. This method has **no XTRACE**, **no Merkle hash**, and **no translateDepth tracking** — those stay in the callers.

#### 1d. Rewrite `Translate` to dispatch atoms

The new `Translate` keeps the existing XTRACE + Merkle + depth logic, then adds an `is_atom` check before calling `translateCore`:

```go
func (t *Translator) Translate(n logic.Expr) (Expr, error) {
    xtracer.Trace("ivy_solver.py:638 formula_to_z3_int() ENTER type=%v", iu.ShortTypeName(n))
    if xtracer.Enabled && t.translateDepth == 0 {
        canon := iu.Canonical(n.Sexp())
        leaf, root := t.TranslateMerkle.AddLeaf(canon)
        xtracer.Trace("z3bridge.Translate HASH leaf=%s root=%s canon=%s", leaf, root, string(canon))
    }
    t.translateDepth++
    defer func() { t.translateDepth-- }()

    // Python line 645: if ivy_logic.is_atom(fmla): return atom_to_z3(fmla)
    if app, ok := n.(*logic.Apply); ok && len(app.Terms) > 0 {
        if _, isBool := n.NodeSort().(*logic.BooleanSort); isBool {
            return t.atomToZ3(app)
        }
    }

    return t.translateCore(n)
}
```

#### 1e. Add `atomToZ3` method

Mirrors Python's `atom_to_z3` (ivy_solver.py:516). Called only for Boolean-sorted Apply with args:

```go
func (t *Translator) atomToZ3(app *logic.Apply) (Expr, error) {
    c, ok := app.Func.(*logic.Const)
    if !ok {
        return t.translateCore(app) // fallback for non-Const func
    }

    xtracer.Trace("ivy_solver.py:517 atom_to_z3() ENTER rep=%s nargs=%d", c.Name, len(app.Terms))

    // Python line 518: enumerated equality check
    if c.Name == "=" && len(app.Terms) == 2 && t.EnumEqFunc != nil {
        if es, ok2 := app.Terms[0].NodeSort().(*logic.EnumeratedSort); ok2 {
            if result, err := t.EnumEqFunc(app.Terms[0], app.Terms[1], es); result != nil {
                return *result, err
            }
        }
    }

    // Check preds cache (Python z3_predicates)
    predKey := logic.NodeKey(c.Name + ":" + string(c.CSort.Sexp()))
    if cached, ok := t.preds[predKey]; ok {
        return t.applyPred(cached, app.Terms)
    }

    // Check builtin ops first (Go-specific: Python handles via polymacs inside lookup_native)
    result, handled, err := t.translateBuiltinOp(c.Name, app.Terms)
    if err != nil {
        return Expr{}, err
    }
    if handled {
        return result, nil
    }

    // Python line 521: rel = lookup_native(atom.relname, relations, "relation")
    var pred interface{}
    if t.NativeLookup != nil {
        if nativeFn := t.NativeLookup(c.Name, c.CSort, true); nativeFn != nil {
            pred = nativeFn
        }
    }

    if pred == nil {
        // Python lines 527-528: create Z3 Function/Const for uninterpreted relation
        fd, err := t.makeFuncDecl(c.Name, c.CSort.(*logic.FunctionSort))
        if err != nil {
            return Expr{}, err
        }
        pred = fd
    }

    t.preds[predKey] = pred
    return t.applyPred(pred, app.Terms)
}
```

#### 1f. Add `applyPred` helper

```go
func (t *Translator) applyPred(pred interface{}, terms []logic.Expr) (Expr, error) {
    args := make([]Expr, len(terms))
    for i, term := range terms {
        a, err := t.TermToZ3(term)
        if err != nil {
            return Expr{}, err
        }
        args[i] = a
    }
    switch p := pred.(type) {
    case FuncDecl:
        return p.Apply(args...), nil
    case func(args ...Expr) Expr:
        return p(args...), nil
    default:
        return Expr{}, fmt.Errorf("unsupported predicate type: %T", pred)
    }
}
```

#### 1g. Add `TermToZ3` method

Mirrors Python's `term_to_z3` (ivy_solver.py:444):

```go
func (t *Translator) TermToZ3(term logic.Expr) (Expr, error) {
    xtracer.Trace("ivy_solver.py:445 term_to_z3() ENTER type=%s name=%s",
        iu.ShortTypeName(term), termName(term))

    // Python line 446: boolean non-variables → formula_to_z3_int
    if _, isBool := term.NodeSort().(*logic.BooleanSort); isBool {
        if _, isVar := term.(*logic.Variable); !isVar {
            return t.Translate(term)
        }
    }

    // Python line 481-482: Ite in term context
    if ite, ok := term.(*logic.Ite); ok {
        c, err := t.Translate(ite.Cond)    // cond is boolean → formula path
        if err != nil { return Expr{}, err }
        th, err := t.TermToZ3(ite.Then)    // then/else → term path
        if err != nil { return Expr{}, err }
        el, err := t.TermToZ3(ite.Else)
        if err != nil { return Expr{}, err }
        return t.Ctx.Ite(c, th, el), nil
    }

    // All other non-boolean terms: Variable, Const, non-boolean Apply
    return t.translateCore(term)
}
```

#### 1h. Clean up Apply case in `translateCore`

In `translateCore`'s `*logic.Apply` case:

1. **Remove** the spurious `fn, err := t.Translate(node.Func)` / `_ = fn` block (lines 256–260)
2. **Change** arg translation from `t.Translate(term)` to `t.TermToZ3(term)` (both in the NativeLookup branch and the getFuncDecl branch)
3. The NativeLookup `isRelation` computation can be simplified: since Boolean-sorted Apply nodes are intercepted by `atomToZ3` before reaching `translateCore`, any Apply here is non-boolean, so `isRelation = false`. But keep the existing computation for safety.

#### 1i. Update `translateBuiltinOp` args

In `translateBuiltinOp` (~line 650), change the `translateArgs` closure from `t.Translate(term)` to `t.TermToZ3(term)`.

### File 2: `solver/z3convert.go`

#### 2a. Fix LookupNative kind format

At line 550, change:

```go
// OLD:
xtracer.Trace("ivy_solver.py:312 lookup_native() ENTER name=%s kind=%v", sym.Name, isRelation)

// NEW:
kind := "function"
if isRelation {
    kind = "relation"
}
xtracer.Trace("ivy_solver.py:312 lookup_native() ENTER name=%s kind=%s", sym.Name, kind)
```

## Expected Trace After Fix

For `ref.prevents` (Boolean-sorted Apply with 2 args):

| Index  | Go (after fix) | Python |
|--------|---------------|--------|
| 236103 | `formula_to_z3_int() ENTER type=Apply` | `formula_to_z3_int() ENTER type=Apply` |
| 236104 | `atom_to_z3() ENTER rep=ref.prevents nargs=2` | `atom_to_z3() ENTER rep=ref.prevents nargs=2` |
| 236105 | `lookup_native() ENTER name=ref.prevents kind=relation` | `lookup_native() ENTER name=ref.prevents kind=relation` |
| 236106 | `term_to_z3() ENTER type=Variable name=...` | `term_to_z3() ENTER type=Variable name=...` |
| 236107 | `term_to_z3() ENTER type=Variable name=...` | `term_to_z3() ENTER type=Variable name=...` |

## Verification

1. Build: `cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go build ./...`
2. Run existing solver tests: `go test ./solver/ -run TestTranslate -count=1 -v`
3. Run the failing test: `go test ./parser/ -run TestOrdLive -count=1 -v -timeout 300s`
4. Check that the divergence has moved past line 236104
