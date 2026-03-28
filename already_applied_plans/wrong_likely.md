# Fix golden divergence: compileLabeledFormula label iteration

Created: 2026-03-28 (current session)

## Context

`make golden` diverges at xtrace line 124642:
- **Go**: `compiler.Thing ENTER type=Atom`
- **Python**: `compiler.Thing ENTER type=And`

The call chain is:
1. `CompileAssertFormula` (action.go:1243) calls `c.Thing(node)` on a `*ast.LabeledFormula`
2. `Thing` → `CompileNode` → `compileLabeledFormula` (compiler.go:806)
3. `compileLabeledFormula` calls `c.SortifyWithInference(n.Label)` — passes the **whole label** node
4. `SortifyWithInference` → `Thing` traces `type=Atom` (the label node itself)

Python does something different (`ivy_compiler.py:505-511`):
```python
self.label.clone([sortify_with_inference(x) for x in self.label.args])
```
Python iterates over `self.label.args` and calls `sortify_with_inference` on **each arg**, then clones the label with compiled args. When `sortify_with_inference` is called on the first arg (which is an `And`), Python traces `Thing ENTER type=And`.

Go has a **correct** implementation already in `CompileLF` (compiler.go:848-861) that iterates args properly, but `CompileNode` dispatches to the **wrong** function `compileLabeledFormula` instead.

## Fix

**File: `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/compiler.go`**

Fix `compileLabeledFormula` (lines 806-836) to iterate over `n.Label.Args()` instead of passing the whole label to `SortifyWithInference`, matching Python's `[sortify_with_inference(x) for x in self.label.args]` pattern.

Current (wrong):
```go
if n.Label != nil {
    l, err := c.SortifyWithInference(n.Label)
    if err != nil {
        label = nil
    } else {
        label = l
    }
}
```

Fixed (matching Python):
```go
if n.Label != nil {
    var newArgs []ast.Node
    var labelErr error
    for _, arg := range n.Label.Args() {
        compiled, err := c.SortifyWithInference(arg)
        if err != nil {
            labelErr = err
            break
        }
        newArgs = append(newArgs, compiled)
    }
    if labelErr != nil {
        label = nil // label compilation failure is not fatal
    } else {
        compiledLabel := n.Label.Clone(newArgs)
        // compiledLabel is ast.Node; wrap as lg.Expr if needed for downstream
        if expr, ok := compiledLabel.(lg.Expr); ok {
            label = expr
        }
        // If the cloned label doesn't satisfy lg.Expr, we still have it as ast.Node.
        // The label gets wrapped into il.NewDefinition below, which accepts lg.Expr.
        // If label remains nil here, the formula is returned without a label wrapper.
    }
}
```

Note: `lg.Expr` extends `ast.Node`, so `lg.Expr` values from `SortifyWithInference` can be passed to `Clone([]ast.Node)`. The cloned label node may or may not satisfy `lg.Expr` depending on its concrete type (e.g., `And` implements `lg.Expr` since it's a formula node). If it doesn't, we need to handle that — possibly wrapping it similarly to how `CompileLF` handles it (which returns `*ast.LabeledFormula`, not `lg.Expr`).

Actually, looking more carefully: the correct approach may be to have `CompileNode`'s `*ast.LabeledFormula` case call `CompileLF` and wrap the result, since `compileLabeledFormula`'s return type (`lg.Expr` via `il.Definition`) diverges from Python's return type (`LabeledFormula` AST node). But this requires ensuring all callers of `CompileNode` that pass `LabeledFormula` can handle the wrapped result.

**Simpler safe approach**: Just fix the label iteration in `compileLabeledFormula`. The `il.NewDefinition(label, fmla)` wrapping is a Go-specific adaptation that works downstream. The immediate divergence is caused solely by the label iteration pattern.

## Verification

Run `make golden` from `~/goivy`. The test should either pass or advance past line 124642 to a later divergence point.
