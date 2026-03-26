# Plan: Port `LabeledFormula.compile()` (clone pattern) to Go DomainSetup (golden test divergence at 76364)

**Created:** 2026-03-25T23:30

## Context

Golden test diverges at line 76364:
```
76363  go : XTRACE: compiler.DomainSetup ENTER
       py : XTRACE: compiler.DomainSetup ENTER

76364  go : XTRACE: compiler.DomainSetup EXIT
       py : XTRACE: ast.LF.clone PRESERVE origid=205 counter=1522
```

Go's `DomainSetup` enters and exits immediately without any `LF.clone PRESERVE` traces, while Python emits clone traces for every axiom/property compilation. This is a **canary for an incomplete port** of Python's `LabeledFormula.compile()` mechanism.

### Root cause

Python's `IvyDomainSetup.axiom(ax)` calls `cax = ax.compile()`, which invokes `LabeledFormula.cmpl` (ivy_compiler.py:411-414):

```python
LabeledFormula.cmpl = lambda self: self.clone([
    None if self.label is None else self.label.clone([sortify_with_inference(x) for x in self.label.args]),
    self.formula.compile() if isinstance(self.formula, SchemaBody) else sortify_with_inference(self.formula)
])
```

This calls `self.clone(...)` which:
1. Creates a new LabeledFormula preserving metadata (temporal, explicit, ID, etc.)
2. Emits `ast.LF.clone PRESERVE origid=N counter=M` trace
3. Properly manages the LfCounter

Go's `DomainSetup.Axiom` instead calls `d.Compiler.CompileNode(lf)` (returns `lg.Expr`), then manually constructs `&ast.LabeledFormula{...}` — bypassing `Clone()` entirely. This means:
- No PRESERVE trace
- No LfCounter management
- Cfg not set on new struct (future Clone calls would panic)
- Label not compiled (not sortified)

## Divergences Found

### D1 — `DomainSetup.Axiom` doesn't call `lf.Clone()`
**Python** (ivy_compiler.py:1032-1037): `cax = ax.compile()` → calls `LF.cmpl` → `self.clone([compiledLabel, compiledFormula])`
**Go** (compiler/decl.go:510-545): calls `CompileNode(lf)`, then `&ast.LabeledFormula{...}` manually
**Fix:** Use new `CompileLF()` method that calls `lf.Clone()`

### D2 — `DomainSetup.Property` same issue
**Python** (ivy_compiler.py:1038-1041): `lf = ax.compile()` → clone
**Go** (compiler/decl.go:548-574): manual construction
**Fix:** Same as D1

### D3 — `DomainSetup.Derived` doesn't call `lf.Clone()`
**Python** (ivy_compiler.py:1156-1171): `self.add_definition(ldf.clone([label, df]))` — explicit clone
**Go** (compiler/decl.go:634-707): `&ast.LabeledFormula{Formula: compiled, Lineno: ...}` — manual construction, missing Label, missing metadata
**Fix:** Call `lf.Clone([]ast.Node{lf.Label, compiled})`

### D4 — `DomainSetup.DefinitionDecl` doesn't call `lf.Clone()`
**Python** (ivy_compiler.py:1172-1178): `self.add_definition(ldf.clone([label, df]))` — explicit clone
**Go** (compiler/decl.go:711-780): manual construction
**Fix:** Call `lf.Clone([]ast.Node{lf.Label, compiled})`

### D5 — Other `&ast.LabeledFormula{}` in DomainSetup missing metadata
Several other DomainSetup methods (Schema, Theorem) create LabeledFormula manually. Python creates them via `LabeledFormula(label, formula)` (constructor, not clone). These are new constructions, not clones of existing LFs, so they need `cfg.NewLabeledFormula()` instead of bare struct literals.

## Implementation Plan

### Step 1: Add `CompileLF` method to Compiler

New method in `compiler/compiler.go` matching Python's `LabeledFormula.cmpl` (ivy_compiler.py:411-414):

```go
// CompileLF compiles a LabeledFormula by cloning it with compiled children.
// Matches Python's LabeledFormula.cmpl (ivy_compiler.py:411-414):
//   self.clone([
//     None if self.label is None else self.label.clone([sortify_with_inference(x) for x in self.label.args]),
//     self.formula.compile() if isinstance(self.formula, SchemaBody) else sortify_with_inference(self.formula)
//   ])
func (c *Compiler) CompileLF(lf *ast.LabeledFormula) (*ast.LabeledFormula, error) {
    // Compile label
    var compiledLabel ast.Node
    if lf.Label != nil {
        var newArgs []ast.Node
        for _, arg := range lf.Label.Args() {
            compiled, err := c.SortifyWithInference(arg)
            if err != nil {
                return nil, err
            }
            newArgs = append(newArgs, compiled)
        }
        compiledLabel = lf.Label.Clone(newArgs)
    }

    // Compile formula
    var compiledFormula ast.Node
    if _, ok := lf.Formula.(*ast.SchemaBody); ok {
        f, err := c.CompileNode(lf.Formula)
        if err != nil {
            return nil, err
        }
        compiledFormula = f
    } else {
        f, err := c.SortifyWithInference(lf.Formula)
        if err != nil {
            return nil, err
        }
        compiledFormula = f
    }

    // Clone preserving ID and metadata — triggers PRESERVE trace
    result := lf.Clone([]ast.Node{compiledLabel, compiledFormula}).(*ast.LabeledFormula)
    return result, nil
}
```

### Step 2: Update `DomainSetup.Axiom` (compiler/decl.go:510-545)

Replace `CompileNode(lf)` + manual construction with `CompileLF(lf)`:

```go
func (d *DomainSetup) Axiom(node ast.Node) error {
    lf, ok := node.(*ast.LabeledFormula)
    if !ok {
        return nil
    }
    // Python: cax = ax.compile() — uses LF.cmpl which clones with compiled children
    cax, err := d.Compiler.CompileLF(lf)
    if err != nil {
        return err
    }
    // Check if it's a schema body (Python checks compiled result)
    if _, ok := cax.Formula.(*ast.SchemaBody); ok {
        labelName := cax.LabelName()
        if labelName != "" {
            d.Compiler.Module.Schemata[labelName] = cax
        }
    } else {
        d.Compiler.Module.LabeledAxioms = append(d.Compiler.Module.LabeledAxioms, cax)
    }
    return nil
}
```

### Step 3: Update `DomainSetup.Property` (compiler/decl.go:548-574)

```go
func (d *DomainSetup) Property(node ast.Node) error {
    lf, ok := node.(*ast.LabeledFormula)
    if !ok {
        return nil
    }
    // Python: lf = ax.compile() — uses LF.cmpl clone pattern
    clf, err := d.Compiler.CompileLF(lf)
    if err != nil {
        return err
    }
    d.Compiler.Module.LabeledProps = append(d.Compiler.Module.LabeledProps, clf)
    d.LastFact = clf.Formula.(lg.Expr) // formula is compiled lg.Expr
    return nil
}
```

### Step 4: Update `DomainSetup.Derived` (compiler/decl.go:634-707)

After compiling the definition, use `lf.Clone()` instead of manual struct construction:

```go
// Replace lines 688-691:
//   mlf := &ast.LabeledFormula{Formula: compiled, Lineno: lf.GetLineno().Line}
// With:
mlf := lf.Clone([]ast.Node{lf.Label, compiled}).(*ast.LabeledFormula)
```

### Step 5: Update `DomainSetup.DefinitionDecl` (compiler/decl.go:711-780)

Same pattern as Derived:

```go
// Replace lines 761-764:
//   mlf := &ast.LabeledFormula{Formula: compiled, Lineno: lf.GetLineno().Line}
// With:
mlf := lf.Clone([]ast.Node{lf.Label, compiled}).(*ast.LabeledFormula)
```

### Step 6: Verify `LastFact` type compatibility

Currently `LastFact` is `lg.Expr`. After the fix, `d.LastFact = clf.Formula.(lg.Expr)` needs the Formula to be an `lg.Expr`. Since `SortifyWithInference` returns `lg.Expr` which satisfies `ast.Node`, and `lg.Expr` embeds `ast.Node` (logic/node.go:8-9), the type assertion should work. Need to verify this compiles.

## Files to Modify

| File | Change |
|------|--------|
| `compiler/compiler.go` | Add `CompileLF` method (~30 lines) |
| `compiler/decl.go:510-545` | Rewrite `Axiom` to use `CompileLF` |
| `compiler/decl.go:548-574` | Rewrite `Property` to use `CompileLF` |
| `compiler/decl.go:688-691` | Update `Derived` to use `lf.Clone()` |
| `compiler/decl.go:761-764` | Update `DefinitionDecl` to use `lf.Clone()` |

## Verification

```bash
cd ~/goivy && go build ./... && go test ./compiler/... && cd ~/goivy && make golden
```

Golden test should advance past line 76364, showing `ast.LF.clone PRESERVE` traces during DomainSetup.
