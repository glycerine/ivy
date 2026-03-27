# Fix conjecture compilation to use ThingLF instead of SortifyWithInference
**Created:** 2026-03-27

## Context

Golden test diverges at line 102761. After `IvyConjectureSetup.dispatch name=conjecture`:
- **Go**: `compiler.sortify_with_inference ENTER`
- **Python**: `compiler.Thing ENTER type=LabeledFormula`

Python's `IvyConjectureSetup.conjecture(ax)` calls `ax.compile()` which goes through `thing()` → `_labeled_formula_cmpl()`, returning a cloned `LabeledFormula` with compiled children. Go's conjecture handler incorrectly calls `SortifyWithInference(arg)` directly, skipping the `Thing`/`CompileLF` dispatch entirely.

Go already has `ThingLF` (`compiler/phase6.go:70`) which matches Python's `thing()` → `LabeledFormula.cmpl()` path exactly, emitting the correct traces and returning `*ast.LabeledFormula`.

## Plan

**File:** `compiler/ivy_compile.go` (lines 447-473)

Replace the conjecture case in `ConjSetup.ProcessDecls` to use `ThingLF` instead of `SortifyWithInference`:

```go
case *ast.ConjectureDecl:
    // Python: conjecture(self, ax): cax = ax.compile(); self.domain.labeled_conjs.append(cax)
    for _, arg := range n.DeclArgs {
        lf, ok := arg.(*ast.LabeledFormula)
        if !ok {
            pp("ConjSetup: conjecture arg is not LabeledFormula: %T", arg)
            continue
        }
        compiled, err := cs.Compiler.ThingLF(lf)
        if err != nil {
            pp("ConjSetup: compiling conjecture: %v", err)
            continue
        }
        xtracer.Trace("compiler.ConjSetup.conjecture compiled")
        cs.Compiler.Module.LabeledConjs = append(cs.Compiler.Module.LabeledConjs, compiled)
        cs.lastFact = compiled
    }
```

This eliminates:
- The manual `NewLabeledFormula` + metadata copy (lines 457-470) — `CompileLF` already clones with all metadata preserved
- The wrong `SortifyWithInference` call — replaced with `ThingLF` matching Python's `ax.compile()`

## Verification

Run `cd ~/goivy && make golden` and confirm line 102761 now matches.
