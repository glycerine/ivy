# Fix: `addLabel` and `NewLabeledFormulaFrom` lose `Base.Loc` — "(internal)" instead of file:line

**Created**: 2026-04-02 01:30
**Previous fixes**: distinctObjRenaming identity mappings (applied), Module.Copy() ActCfg nil (applied)

## Context

After `make golden`, the XTRACE traces now match, but the Go output prints `(internal)` where Python prints the actual source location (filename + line number) for auto-generated properties like `prop4` and `prop1` from `order.ivy`.

From `diff.trace.manual`:
```
Python:  <IVY_INCLUDE>/1.8/order.ivy: line 33: index.spec.prop4
Go:      (internal) index.spec.prop4

Python:  <IVY_INCLUDE>/1.8/order.ivy: line 11: index.spec.prop1
Go:      (internal) index.spec.prop1
```

## Root cause

Go's `LabeledFormula` has **two** lineno-related fields:
1. `Lineno int` — a bare line number (no filename)
2. `Base.Loc Location` — full location with `Filename` and `Line`

Python has a single `lineno` field that holds a `Location`/`LocationTuple` object (filename + line).

### The bug: `addLabel()` and `NewLabeledFormulaFrom()` copy `Lineno` but NOT `Base.Loc`

**`addLabel()`** in `lalr_full/grammar_v17.go:64-74`:
```go
func addLabel(cfg *ast.AstConfig, lf *ast.LabeledFormula, pref string) *ast.LabeledFormula {
    if lf.Label != nil {
        return lf
    }
    res := cfg.NewLabeledFormula(newLabel(cfg, pref), lf.Formula)
    res.Lineno = lf.Lineno  // ← copies int, but NOT Base.Loc!
    return res
}
```

**`NewLabeledFormulaFrom()`** in `ast/decl_ast.go:60-68`:
```go
func (cfg *AstConfig) NewLabeledFormulaFrom(src *LabeledFormula, formula Node) *LabeledFormula {
    lf := cfg.NewLabeledFormula(src.Label, formula)
    lf.Lineno = src.Lineno  // ← copies int, but NOT Base.Loc!
    lf.Temporal = src.Temporal
    // ...
}
```

Meanwhile, `PrettyLineno()` in `check/helpers.go:41-50` calls `lf.GetLineno()` which returns `Base.Loc` (not `Lineno`). Since `Base.Loc` was never copied, it's empty → "(internal)".

**Note:** `cloneInternal()` (decl_ast.go:94-110) correctly copies both `Base.Loc` AND `Lineno`. The fix should follow its pattern.

### The flow for properties from order.ivy

1. `p_lgprop` grammar rule: `lf.SetLineno(nodeLineno(...))` → sets `Base.Loc` with filename+line ✓
2. `p_schdecl_theorem_lgprop` grammar rule: calls `addLabel(...)` → creates NEW LabeledFormula, copies `Lineno` but **loses `Base.Loc`** ✗
3. Later, `PrettyLineno(lf)` calls `lf.GetLineno()` → gets empty `Base.Loc` → "(internal)"

## Changes

### Step 1: Fix `addLabel()` to copy `Base.Loc`

**File: `lalr_full/grammar_v17.go`**, lines 64-74:

```go
func addLabel(cfg *ast.AstConfig, lf *ast.LabeledFormula, pref string) *ast.LabeledFormula {
    xtracer.Trace("parser.addlabel ENTER")
    if lf.Label != nil {
        return lf
    }
    res := cfg.NewLabeledFormula(newLabel(cfg, pref), lf.Formula)
    res.Lineno = lf.Lineno
    if lf.HasLocSet() {
        res.SetLineno(lf.GetLineno())  // ADD: copy Base.Loc (filename + line)
    }
    return res
}
```

### Step 2: Fix `NewLabeledFormulaFrom()` to copy `Base.Loc`

**File: `ast/decl_ast.go`**, lines 60-68:

```go
func (cfg *AstConfig) NewLabeledFormulaFrom(src *LabeledFormula, formula Node) *LabeledFormula {
    lf := cfg.NewLabeledFormula(src.Label, formula)
    lf.Lineno = src.Lineno
    if src.HasLocSet() {
        lf.SetLineno(src.GetLineno())  // ADD: copy Base.Loc (filename + line)
    }
    lf.Temporal = src.Temporal
    lf.Explicit = src.Explicit
    lf.Assumed = src.Assumed
    lf.Unprovable = src.Unprovable
    return lf
}
```

## Key files

- `lalr_full/grammar_v17.go:64-74` — `addLabel()` (loses Base.Loc)
- `ast/decl_ast.go:60-68` — `NewLabeledFormulaFrom()` (loses Base.Loc)
- `ast/decl_ast.go:94-110` — `cloneInternal()` (correctly copies both — reference pattern)
- `check/helpers.go:41-50` — `PrettyLineno()` (uses GetLineno() → Base.Loc)
- Python: `ivy_parser.py:515-521` — `addlabel()` copies `lf.lineno` (which IS the full Location)

## Verification

```bash
cd ~/ivy/goivy && make test && make golden
```

Check that:
1. `make test` passes
2. `diff out.py.xtrace out.go.xtrace` no longer shows "(internal)" vs file:line differences
3. The properties section in the output shows proper `order.ivy: line 33:` and `order.ivy: line 11:` locations
