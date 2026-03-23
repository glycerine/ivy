# Fix: Progress Symbol Removal — `p.defines()` in ivy_compile.go

## Context

After completing Fixes A–G, we found that `compiler/ivy_compile.go:137-152` (the progress symbol removal loop) is broken: it checks for `DefinerStr` and `DefinerSlice` interfaces, but the types stored in `mod.Progress` implement **neither**.

**Python source (ivy_compiler.py:2209-2210):**
```python
for p in mod.progress:
    remove_symbol(p.defines())
```

**Investigation revealed:** This is a **latent bug in Python too.** Here's why:

1. `DomainSetup.progress()` (ivy_compiler.py:1197-1202) stores `sortify_with_inference(df)` where `df` is a `LabeledFormula`
2. `LabeledFormula.cmpl` (line 410-413) returns a cloned `LabeledFormula`
3. `sort_infer()` preserves the type — returns `LabeledFormula`
4. **`LabeledFormula` has NO `defines()` method** — confirmed by `hasattr(lf, 'defines') == False`
5. Therefore `p.defines()` would raise `AttributeError` at runtime

The progress feature is extremely rare — no real Ivy program appears to exercise this code path, so the Python bug is latent.

## What Go Should Do

Rather than faithfully replicating a Python bug, the Go code should implement the **intended** semantics: extract the progress relation's symbol name from the compiled progress item, and remove it from the signature.

A progress declaration has the form: `progress rel(X) = body`. After compilation:
- Pass 1 (`DomainSetup.Progress`): stores a compiled `lg.Expr` — the result of `SortifyWithInference(node)`. The progress `node` is a `LabeledFormula` containing a relation application. After compilation, the formula field becomes an `*lg.Apply` where `.Func` is the progress symbol.
- Pass 3 (`ARGSetup`): stores raw `ast.Node` — specifically `*ast.LabeledFormula` from `ProgressDecl.DeclArgs`.

The symbol name we need to extract is the **relation name** — the `relname` of the progress relation atom.

## Design Decision

**Don't add `Defines()` to `lg.Expr`.** It's semantically wrong for 24 of 26 types. Instead, use a targeted type-switch extraction function.

**Don't use `DefinerStr`/`DefinerSlice` interfaces.** The types in `mod.Progress` don't implement them and shouldn't need to.

## Fix

**File: `compiler/ivy_compile.go`** — Replace lines 134-152 with:

```go
// Python lines 2209-2210: remove progress symbols from sig
// Progress properties are not state symbols — remove from sig.
for _, p := range mod.Progress {
    name := progressDefinesName(p)
    if name != "" && mod.Sig != nil {
        if sym, err := mod.Sig.FindSymbol(name, false); err == nil {
            mod.Sig.RemoveSymbol(name, sym.CSort)
        }
    }
}
```

Add helper function (same file):

```go
// progressDefinesName extracts the symbol name from a progress item.
// Progress items are either:
//   - lg.Expr (compiled via SortifyWithInference in pass 1) — extract from Apply.Func or Definition LHS
//   - ast.LabeledFormula (raw from pass 3) — extract from Label
func progressDefinesName(p interface{}) string {
    // Pass 1: compiled lg.Expr
    if expr, ok := p.(lg.Expr); ok {
        return exprDefinesName(expr)
    }
    // Pass 3: raw ast.LabeledFormula
    if lf, ok := p.(*ast.LabeledFormula); ok {
        if lf.Label != nil {
            if atom, ok := lf.Label.(*ast.Atom); ok {
                return atom.Relname()
            }
        }
        // Try the formula field
        if lf.Formula != nil {
            if expr, ok := lf.Formula.(lg.Expr); ok {
                return exprDefinesName(expr)
            }
        }
    }
    return ""
}

// exprDefinesName extracts the defined symbol name from a compiled lg.Expr.
// Handles: Apply (func name), Definition (LHS name), Symbol (name directly).
func exprDefinesName(expr lg.Expr) string {
    switch e := expr.(type) {
    case *lg.Apply:
        if sym, ok := e.Func.(*lg.Symbol); ok {
            return sym.Name
        }
    case *lg.Definition:
        defNode := e.Defines()
        if sym, ok := defNode.(*lg.Symbol); ok {
            return sym.Name
        }
    case *lg.Symbol:
        return e.Name
    }
    return ""
}
```

## Files to Modify

| File | Change |
|------|--------|
| `compiler/ivy_compile.go` | Replace progress loop (lines 134-152) with type-switch extraction; add `progressDefinesName` and `exprDefinesName` helpers |

## Verification

1. `go build ./compiler/` — compiles
2. `go test ./compiler/ ./...` — no regressions
3. The `panicf` at line 150 is eliminated
4. Both pass-1 (`lg.Expr`) and pass-3 (`*ast.LabeledFormula`) progress items are handled
