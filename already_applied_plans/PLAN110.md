(rejected: prelim for actual in PLAN111.md)
# Fix xtrace divergence at line 124589: Route Sequence through real Thing/OtherThing

**Created:** 2026-03-27

## Context

Golden test diverges at xtrace line 124589:
- **Go:** `compiler.ARGSetup.action EXIT name=index.next key=index.next`
- **Python:** `compiler.Thing ENTER type=Sequence`

Go's `CompileActionBody` handles `*ast.Sequence` directly with its own logic, bypassing `Thing`/`CompileNode`/`OtherThing`. Python routes Sequence through `thing()` → `other_thing()` (since Sequence has no explicit `.cmpl`). We must route through the REAL functions, not fake traces.

## Python's path for Sequence body compilation

```
sortify(body) → body.compile() → thing() → self.cmpl() → other_thing()
other_thing: self.clone([a.compile() for a in self.args])
```
Each `a.compile()` → `thing(child)` → recursive compilation. Returns cloned Sequence with compiled children.

## Go's architecture difference

- `Thing`/`CompileNode` return `lg.Expr`, but `CompileActionBody` returns `actions.Action`
- `ast.Sequence` does NOT implement `lg.Expr`
- `compileGeneric` (called by `OtherThing`) clones the node and if clone isn't `lg.Expr`, extracts children: single child → returns it; multiple children → wraps in `lg.And`
- `actions.WrapAction`/`UnwrapAction` bridge between `actions.Action` and `lg.Expr`

## Fix

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/action.go`

Replace the `*ast.Sequence` case in `CompileActionBody` to route through `Thing`:

```go
case *ast.Sequence:
    // Python: Sequence has no explicit .cmpl, uses other_thing() path:
    //   thing() → CompileNode (default) → OtherThing → compileGeneric
    //   compileGeneric: self.clone([a.compile() for a in self.args])
    // Route through the REAL Thing path to get genuine traces.
    result, err := c.Thing(node)
    if err != nil {
        return nil, err
    }
    // Single child: compileGeneric returns it directly (a wrapped action)
    if act := actions.UnwrapAction(result); act != nil {
        return act, nil
    }
    // Multiple children: compileGeneric returns lg.And{Terms: [...]}
    // Extract wrapped actions and build actions.Sequence
    if andExpr, ok := result.(*lg.And); ok {
        seq := actions.NewSequence(andExpr.Terms...)
        seq.SetLineno(node.GetLineno())
        return seq, nil
    }
    // Zero children or unexpected result
    return actions.NewSequence(), nil
```

This routes through:
1. `Thing(node)` → traces `Thing ENTER type=Sequence`
2. `CompileNode(node)` → traces `CompileNode ENTER type=Sequence`, no Sequence case → falls to default → traces `CompileNode return case=default type=Sequence`
3. `OtherThing(node)` → traces `OtherThing ENTER type=Sequence`, calls `compileGeneric`
4. `compileGeneric` → calls `Thing` on each child statement (real recursive compilation)
5. Returns result → `CompileActionBody` converts `lg.Expr` back to `actions.Action`

**Note:** `*ast.And` case is left unchanged for now — it has a specific case in `CompileNode` (formula conjunction) which differs from its action-sequence usage. The golden test will reveal if it needs similar treatment later.

## Verification

Run `cd ~/goivy && make golden` and confirm the test advances past line 124589.
