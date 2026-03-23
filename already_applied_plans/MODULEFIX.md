# Plan: Fix ivy_module.py → module/ Port Issues (AUDIT18MARCH §11)

## Context

The AUDIT18MARCH.md §11 identifies 15 issues (10 MISSING, 5 BEHAVIORAL_DIFFERENCE) in the Go port of `ivy_module.py`. This plan addresses all 15, grouped into 5 batches by similarity and complexity.

**Python source:** `/Users/jaten/pyivy/ivy/ivy/ivy_module.py`
**Go target:** `/Users/jaten/goivy/module/`

**CONSTRAINT: No package-level variables.** All state must live on `Config`, `Module`, or be passed as parameters. This enables future thread-pooling and multi-tenancy. Package-level *functions* that take `cfg`/`m` are fine; package-level `var` state is forbidden. B1 (removing `theoryCache`) is an existing violation being fixed.

---

## Batch 1: One-liner / Small Fixes (4 items)

### M2: InitCond not initialized in Clear()
- **File:** `module/module.go` line 275 (end of `Clear()`)
- **Change:** Add `m.InitCond = co.TrueClauses(nil)` before closing brace
- `co.TrueClauses` already exists at `clauseops/clauses.go:218`

### M10: Copy() missing 7 map copies
- **File:** `module/module.go` lines 330-331 (in `Copy()`, after `FiniteSorts`)
- **Add copy blocks for:** `SortDestructors`, `SortConstructors`, `Variants`, `Supertypes`, `ExtPreconds`, `ConjActions`, `IsolateProofs`
- Pattern: reuse existing helpers (`copyMapNode` for IsolateProofs) and add new helpers (`copyMapSymSlice` for `map[string][]*lg.Symbol`, `copyMapSortSlice` for `map[string][]lg.Sort`, `copyMapExpr` for `map[string]lg.Expr`, `copyMapStrSlice` for `map[string][]string`)

### B3: VariantIndex returns -1 vs Python None — NO CHANGE NEEDED
- Go's -1 is idiomatic and strictly better than Python's implicit None (which crashes callers)

### B4: ResortAliasesMap — NO CHANGE NEEDED
- Go correctly returns the result; Python has the bug (forgot `return res`)

---

## Batch 2: Enter/Exit + Package-Level Functions (3 items)

### M1: Enter/Exit must save/restore sig and call solver.Clear()
- **Files:** `module/module.go` (add field), `module/context.go` (modify Enter/Exit), `module/config.go` (add callback)

**module/module.go** — Add field to Module struct (after line 140):
```go
oldSig *il.Sig  // saved by Enter(), restored by Exit()
```

**module/config.go** — Add to Config struct:
```go
SolverClearFn func()  // called on Enter to clear Z3 caches; set by solver package
```

**module/context.go** — Modify Enter():
```go
func (m *Module) Enter() {
    cfg := m.Cfg
    if cfg == nil {
        panic("module.Enter: Cfg is nil — caller must set Cfg before calling Enter")
    }
    m.prevModule = cfg.CurrentModule
    m.oldSig = il.GlobalSig  // save current global sig (if applicable)
    cfg.CurrentModule = m
    // Python: il.sig = self.sig — set the "current sig" to this module's sig
    // Python: ivy_solver.clear()
    if cfg.SolverClearFn != nil {
        cfg.SolverClearFn()
    }
}
```
Note: Need to verify how `il.sig` global works in Go. The Go codebase may store sig on the module itself (m.Sig) and pass it explicitly. If there's no `il.GlobalSig`, we just save/restore `cfg.CurrentModule` (which carries `Sig`). The solver.Clear() callback is the key missing piece.

**module/context.go** — Modify Exit():
```go
func (m *Module) Exit() {
    cfg := m.Cfg
    if cfg == nil {
        panic("module.Exit: Cfg is nil — caller must set Cfg before calling Exit")
    }
    cfg.CurrentModule = m.prevModule
    m.prevModule = nil
    m.oldSig = nil
}
```

### M4: Package-level FindAction
- **File:** `module/context.go` — Add:
```go
func FindAction(cfg *Config, name string) (Action, bool) {
    if cfg == nil || cfg.CurrentModule == nil {
        return nil, false
    }
    return cfg.CurrentModule.FindAction(name)
}
```

### M5: Package-level BackgroundTheory
- **File:** `module/context.go` — Add:
```go
func BackgroundTheory(cfg *Config, symbols map[string]bool) *co.Clauses {
    if cfg == nil || cfg.CurrentModule == nil {
        return co.NewClauses(nil, nil, nil)
    }
    return cfg.CurrentModule.BackgroundTheory(symbols)
}
```
- Will need to add `co "github.com/glycerine/goivy/clauseops"` to context.go imports

---

## Batch 3: CanonizeTypes + SortDependencies (2 items)

### M3: CanonizeTypes missing 5 fields + ResortSig call
- **File:** `module/canonize.go` line 59 (before closing brace of `CanonizeTypes`)
- **Add** after the `ExtPreconds` line:

```go
// Resort InitCond (Python line 250)
m.InitCond = ResortClauses(m.InitCond, rn)

// Resort ConceptSpaces (Python line 251)
for i, cs := range m.ConceptSpaces {
    m.ConceptSpaces[i] = ConceptSpace{
        Label: ResortAST(cs.Label, rn),
        Body:  ResortAST(cs.Body, rn),
    }
}

// Resort Progress (Python line 254)
for i, p := range m.Progress {
    if expr, ok := p.(lg.Expr); ok {
        m.Progress[i] = ResortAST(expr, rn)
    }
}

// Resort BeforeExport (Python line 261)
newBE := make(map[string]Action, len(m.BeforeExport))
for k, v := range m.BeforeExport {
    if expr, ok := v.(lg.Expr); ok {
        if resorted := ResortAST(expr, rn); resorted != nil {
            if act, ok2 := resorted.(Action); ok2 {
                newBE[k] = act
                continue
            }
        }
    }
    newBE[k] = v
}
m.BeforeExport = newBE

// Resort sig (Python line 263: lu.resort_sig(sort_refinement))
ResortSig(m.Sig, rn)
```

- `ResortClauses` already exists at `canonize.go:317`
- `ResortSig` already exists at `resort.go:62` and takes `map[lg.NodeKey]*SortRefinement`

### M7: SortDependencies NativeTypes branch
- **File:** `module/module.go` lines 468-469 (between SortDestructors and Variants blocks)
- **Insert:**
```go
// NativeTypes branch (Python lines 397-400)
if nt, ok := m.NativeTypes[sortName]; ok {
    if nt != nil && len(nt.Args) > 1 {
        var deps []string
        for _, elem := range nt.Args[1:] {
            rep := ast.GetRep(elem)  // or elem.Relname() pattern
            if rep != "" {
                if _, inSig := m.Sig.Sorts[rep]; inSig {
                    deps = append(deps, rep)
                }
            }
        }
        return deps
    }
}
```
- Need to verify how `ast.NativeType.Args` is structured (field might be `Elems` or `Args`)

---

## Batch 4: Implementations (3 items)

### M6: param_logic parameter and logics() — full translation

Python has three layers:
1. `param_logic = iu.Parameter("complete", ','.join(il.default_logics), check=...)` — a CLI-settable parameter
2. `module.logics` — per-module override (set by compiler when processing `attribute complete = ...`)
3. `logics()` function — checks module first, falls back to param_logic

**Go translation** — no package-level variables; parameter lives on Config:

**File: `module/config.go`** — Add field to Config struct:
```go
// CompleteLogic is the comma-separated logic parameter (Python: param_logic).
// Default is "" meaning use il.DefaultLogics. Set via CLI --complete flag
// or programmatically. Corresponds to Python's iu.Parameter("complete", ...).
CompleteLogic string `json:"complete"`
```

**File: `module/config.go`** — No change to `NewConfig()` needed; empty string means "use default".

**File: `module/module.go`** — Rewrite `GetLogics()` (lines 481-488) to implement the full Python `logics()` fallback chain:
```go
func (m *Module) GetLogics() []string {
    // Python logics(): check module.logics first
    if len(m.Logics) > 0 {
        return m.Logics
    }
    // Python: fall back to param_logic.get().split(',')
    if m.Cfg != nil && m.Cfg.CompleteLogic != "" {
        return strings.Split(m.Cfg.CompleteLogic, ",")
    }
    // Ultimate default: il.DefaultLogics (= ["epr"])
    return il.DefaultLogics
}
```

**File: `module/context.go`** — Add `Logics()` as a stateless function taking cfg (no global state):
```go
// Logics returns the active logic names, checking current module first,
// then Config.CompleteLogic, then il.DefaultLogics.
// Corresponds to Python's module-level logics() function (line 355-358).
func Logics(cfg *Config) []string {
    if cfg != nil && cfg.CurrentModule != nil {
        return cfg.CurrentModule.GetLogics()
    }
    if cfg != nil && cfg.CompleteLogic != "" {
        return strings.Split(cfg.CompleteLogic, ",")
    }
    return il.DefaultLogics
}
```

This faithfully translates all three layers of the Python logic with zero package-level state.

### M8: CallGraph() implementation
- **File:** `module/module.go` lines 448-452
- **Replace** placeholder with:
```go
func (m *Module) CallGraph() map[string][]string {
    callgraph := make(map[string][]string)
    for actname, action := range m.Actions {
        for _, calledName := range action.IterCalls() {
            callgraph[calledName] = append(callgraph[calledName], actname)
        }
    }
    return callgraph
}
```
- `Action.IterCalls()` already defined in interface at `action.go:27`

### M9: UpdateConjs() implementation
- **File:** `module/module.go` lines 589-607
- **Replace** placeholder with full implementation:
```go
func (m *Module) UpdateConjs() {
    for i, cax := range m.LabeledConjs {
        if cax == nil || cax.Formula == nil {
            continue
        }
        fmla, ok := cax.Formula.(lg.Expr)
        if !ok {
            continue
        }
        csname := fmt.Sprintf("conjecture:%d", i)
        variables := co.UsedVariablesOrdered(fmla)
        sorts := make([]lg.Sort, len(variables))
        for j, v := range variables {
            sorts[j] = v.VSort
        }
        symSort := il.RelationSort(sorts)
        sym := lg.NewSymbol(csname, symSort)
        varExprs := make([]lg.Expr, len(variables))
        for j, v := range variables {
            varExprs[j] = v
        }
        var label lg.Expr
        if len(varExprs) > 0 {
            label = &lg.Apply{Func: sym, Terms: varExprs}
        } else {
            label = sym
        }
        space := &lg.Literal{Polarity: 0, Atom: fmla}  // il.Literal(0, fmla)
        m.ConceptSpaces = append(m.ConceptSpaces, ConceptSpace{Label: label, Body: space})
    }
}
```
- `co.UsedVariablesOrdered` exists in clauseops (used at `check/check.go:245`)
- `il.RelationSort` exists in ivylogic
- Need to verify `ConceptSpace` struct fields and `lg.Literal` type

---

## Batch 5: Behavioral Differences (2 items — B1 and B2 need changes, B5 is largest)

### B1: BackgroundTheory theoryCache → Theory field
- **File:** `module/module.go` — Add to Module struct:
```go
Theory *co.Clauses  // cached background theory
```

- **File:** `module/theory.go` — Remove lines 25-41 (theoryMu, theoryCache, getTheory, setTheory)
- **Modify** `BackgroundTheory` (line 48-53):
```go
func (m *Module) BackgroundTheory(inScope map[string]bool) *co.Clauses {
    if m.Theory != nil {
        return m.Theory
    }
    return co.NewClauses(nil, nil, nil)
}
```
- **Modify** `UpdateTheory` line 130: `setTheory(m, cls)` → `m.Theory = cls`
- Remove `sync` import from theory.go

### B2: SortCard attribute access
- **File:** `module/module.go` lines 426-434
- Python accesses `self.attributes[attr].rep` — the attribute value is an AST node with a `.Rep` field
- Current Go does `val.(string)` type assertion
- **Change:** Try accessing `.Rep` field via type switch:
```go
if val, ok := m.Attributes[attr]; ok {
    var rep string
    switch v := val.(type) {
    case interface{ GetRep() string }:
        rep = v.GetRep()
    case string:
        rep = v
    default:
        return -1
    }
    var n int
    if _, err := fmt.Sscanf(rep, "%d", &n); err == nil {
        return n
    }
    return -1
}
```
- Need to verify what AST node type stores attributes and how to access `.Rep`

### B5: Exclusivity() — rewrite to match Python exactly
- **File:** `module/theory.go` lines 437-496
- Python `exclusivity()` at `ivy_logic.py:694-706` generates:
  1. **Partial function axioms**: `∀X:sort,Y:s,Z:s. (pto(s)(X,Y) ∧ pto(s)(X,Z)) → Y=Z` for each variant s
  2. **Injectivity axioms**: `∀X:sort,Y:sort,Z:s. (pto(s)(X,Z) ∧ pto(s)(Y,Z)) → X=Y` for each variant s
  3. **Pairwise exclusion**: `∀X:sort,Y:s1,Z:s2. ¬(pto(s1)(X,Y) ∧ pto(s2)(X,Z))` for each pair
  - Where `pto(s) = Symbol('*>', RelationSort([sort, s]))` — the variant type-of relation
- Note: Python uses `il.partial_function(pto(s))` which is `il.PartialFunction` — already ported at `ivylogic/util.go:321`
- **Replace** entire `Exclusivity` function and remove `makeVariantCheck`:

```go
func Exclusivity(parentSort lg.Sort, variants []lg.Sort) lg.Expr {
    if len(variants) == 0 {
        return &lg.And{} // true
    }
    pto := func(s lg.Sort) *lg.Symbol {
        sort := il.RelationSort([]lg.Sort{parentSort, s})
        return lg.NewSymbol("*>", sort)
    }
    var excs []lg.Expr
    // 1. Partial function axioms
    for _, s := range variants {
        excs = append(excs, il.PartialFunction(pto(s)))
    }
    // 2. Injectivity axioms
    for _, s := range variants {
        rel := pto(s)
        x, _ := lg.NewVariable("X", parentSort)
        y, _ := lg.NewVariable("Y", parentSort)
        z, _ := lg.NewVariable("Z", s)
        relXZ := &lg.Apply{Func: rel, Terms: []lg.Expr{x, z}}
        relYZ := &lg.Apply{Func: rel, Terms: []lg.Expr{y, z}}
        body := &lg.Implies{
            T1: &lg.And{Terms: []lg.Expr{relXZ, relYZ}},
            T2: &lg.Eq{T1: x, T2: y},
        }
        excs = append(excs, &lg.ForAll{Variables: []*lg.Variable{x, y, z}, Body: body})
    }
    // 3. Pairwise exclusion
    for i1, s1 := range variants {
        for _, s2 := range variants[:i1] {
            x, _ := lg.NewVariable("X", parentSort)
            y, _ := lg.NewVariable("Y", s1)
            z, _ := lg.NewVariable("Z", s2)
            body := &lg.Not{Body: &lg.And{Terms: []lg.Expr{
                &lg.Apply{Func: pto(s1), Terms: []lg.Expr{x, y}},
                &lg.Apply{Func: pto(s2), Terms: []lg.Expr{x, z}},
            }}}
            excs = append(excs, &lg.ForAll{Variables: []*lg.Variable{x, y, z}, Body: body})
        }
    }
    return &lg.And{Terms: excs}
}
```
- Remove `makeVariantCheck` function (lines 484-496)
- Note: The audit's claim of a "missing coverage axiom" is **incorrect** — the Python source has no coverage axiom. The three categories above fully match Python.

---

## Files Modified Summary

| File | Changes |
|------|---------|
| `module/module.go` | M2 (Clear InitCond), M10 (Copy maps), M7 (SortDependencies NativeTypes), M6 (GetLogics default), M8 (CallGraph), M9 (UpdateConjs), B1 (Theory field), B2 (SortCard .Rep access) |
| `module/context.go` | M1 (Enter/Exit sig+solver), M4 (FindAction pkg-level), M5 (BackgroundTheory pkg-level), M6 (Logics pkg-level) |
| `module/config.go` | M1 (SolverClearFn callback) |
| `module/theory.go` | B1 (remove theoryCache, use Theory field), B5 (Exclusivity rewrite) |
| `module/canonize.go` | M3 (add 5 missing resort calls + ResortSig) |

---

## Implementation Order

1. **Batch 1** (M2, M10) — no dependencies, simple
2. **Batch 5 B1** (Theory field) — do before Batch 2 since Enter/Exit are related
3. **Batch 2** (M1, M4, M5) — Enter/Exit + pkg functions
4. **Batch 3** (M3, M7) — CanonizeTypes + SortDependencies
5. **Batch 4** (M6, M8, M9) — implementations
6. **Batch 5 B2, B5** (SortCard, Exclusivity) — behavioral fixes

---

## Verification

After each batch:
1. `cd /Users/jaten/goivy && go build ./module/...` — must compile
2. `go test ./module/...` — existing tests must pass
3. After all batches: `go test ./...` — full test suite

For B5 (Exclusivity): verify the generated axiom structure matches Python's output by adding a test case in `module/theory_test.go`.
