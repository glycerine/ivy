# Fix Structural Equality for Symbol Keys Throughout the Port

## Context

Python Ivy uses Symbol objects (with recstruct structural equality over `name + sort`) as dictionary keys and set members everywhere: `modifies()`, `symbols_ast()`, `used_symbols_ast()`, `get_symbol_dependencies()`, `check_definitions()`. Two Symbols with the same name but different sorts are NOT equal.

The Go port incorrectly strips Symbol objects down to plain `c.Name` strings in `actions.Modifies`, `collectFormulaSymbols`, `GetSymbolDependencies`, `symbolsAst`, and all downstream maps. This means two symbols `f:bool` and `f:int` would wrongly collide.

The existing `lg.NodeKey` / `lg.Key()` / `Sexp()` infrastructure already provides the Go equivalent of Python's recstruct structural equality. This plan fixes every site to use it.

## Design Decision: Return Type for Modifies

Python's `modifies()` returns `[n.rep]` — a list of Symbol objects. The Go equivalent: **`[]*lg.Symbol`**.

This matches Python exactly. Callers that need a structural-equality set build one: `set[lg.Key(sym)] = true`. Callers that need the plain name use `sym.Name`. No wrapper type needed.

Same pattern for `collectFormulaSymbols`: return `[]*lg.Symbol` for post-compilation nodes. For pre-compilation AST nodes (`*ast.Atom`, `*ast.Symbol`) that lack sort info, these still need handling — `collectFormulaSymbols` currently handles both. Since it's only used by `GetSymbolDependencies`, and that function needs set semantics with lookups, `collectFormulaSymbols` should return `map[lg.NodeKey]bool` (using `lg.Key(sym)` for `*lg.Symbol`, plain name for `*ast.Atom`).

---

## Changes By File

### 1. `actions/transforms.go` — Modifies return type

**Current:** `func Modifies(action Action) map[lg.NodeKey]bool` using `result[lg.Key(c)] = true`
**Fix:** `func Modifies(action Action) []*lg.Symbol`

- Change `Modifies` to return `[]*lg.Symbol` (matches Python's `return [n.rep]`)
- Change `modifiesRec` to append to `*[]*lg.Symbol`
- AssignAction case: `*result = append(*result, c)` (was `result[lg.Key(c)] = true`)
- HavocAction case: `*result = append(*result, c)`
- Dedup can happen at call sites if needed (Python doesn't dedup in modifies either)

### 2. `compiler/phase6.go` — collectFormulaSymbols, GetSymbolDependencies, CheckMutax

**collectFormulaSymbols + collectFormulaSymbolsRec:**
- Return type: `map[lg.NodeKey]*lg.Symbol`
- `*lg.Symbol` case: `result[lg.Key(n)] = n` (was `result[n.Name] = true`)
- `*ast.Atom` case: key is `n.Rep` (no sort info at pre-compilation), value is nil
- `*ast.Symbol` case: key is `n.Rep`, value is nil

**GetSymbolDependencies:**
- `defMap` type: `map[lg.NodeKey]interface{}` (was `map[string]interface{}`)
- `res` type: `map[lg.NodeKey]bool` (was `map[string]bool`)
- Internal lookups use NodeKey throughout

**CheckMutax:**
- `modified` type: `map[lg.NodeKey]bool` using `result[key] = true` from Modifies
- `defMap` keys: NodeKey (from `extractSortName` → need to fix or replace with NodeKey extraction)
- `deps` type: `map[lg.NodeKey]bool`
- Definition LHS check: use `lg.Key` of the LHS symbol
- Note: `extractSortName` operates on `*ast.Definition` (pre-compilation) — needs to produce compatible keys

### 3. `compiler/ivy_compile.go` — CheckDefinitions interference section

**Current state:** Already partially fixed with `definesKey`/`definesPlainName` split.
**Fix:** Remove the `definesPlainName` function. Use `definesKey` (NodeKey) everywhere.

- Build `modified` set: `for _, sym := range actions.Modifies(act) { modified[lg.Key(sym)] = true }`
- `defMap` type: `map[lg.NodeKey]interface{}` — key by `definesKey(def)`
- `deps` from `GetSymbolDependencies`: now `map[lg.NodeKey]bool` — compare directly against `modified`
- Definition LHS check: `definesKey(def)` looked up in `modified`

### 4. `temporal/temporal.go` — symprops map + symbolsAst dedup

**symprops:**
- Type: `map[lg.NodeKey][]lg.Expr` (was `map[string][]lg.Expr`)
- Build: `symprops[lg.Key(sym)] = append(...)` (was `symprops[sym.Name] = ...`)
- Lookup: `symprops[lg.Key(sym)]` where `sym` comes from Modifies

**symbolsAst / symbolsAstRec (line 630):**
- Dedup `seen` map: `map[lg.NodeKey]bool` (was `map[string]bool`)
- Check: `seen[lg.Key(c)]` (was `seen[c.Name]`)

**Modifies caller (line 536):**
- `for _, sym := range mods` then `symprops[lg.Key(sym)]` — structural match via Sexp

### 5. `isolate/helpers.go` — GetLocMods (line 940)

**Current:** `strings.HasPrefix(s, "fml:")` where `s` is from Modifies
**Fix:** `for _, sym := range modSet` then `strings.HasPrefix(sym.Name, "fml:")`
Return type stays `[]string` (returns `sym.Name` values for callers).

### 6. `isolate/phase7.go` — sig.Symbols lookup (line 118)

**Current:** `mod.Sig.Symbols[sym]` where `sym` is from Modifies
**Fix:** `for _, sym := range actions.Modifies(sub)` then `mod.Sig.Symbols[sym.Name]`
(`mod.Sig.Symbols` is string-keyed, matching Python's `sym.name in mod.sig.symbols`)

### 7. `l2s/shared.go` — Dependencies, symprops, sig.Symbols lookups

**Line 238 caller:**
- `for _, sym := range mods` then `m.Sig.Symbols[sym.Name]` to get sort info

**Line 520 caller:**
- Build set: `modSet := make(map[lg.NodeKey]bool); for _, sym := range actions.Modifies(stmt) { modSet[lg.Key(sym)] = true }`
- `cfg.Dependencies(modSet)` — Dependencies func must accept/return NodeKey maps
- `symprops`, `symwhens`, `symwaits` maps: key by NodeKey
- Build these maps with `lg.Key(sym)` instead of `sym.Name`

**BuildDependenciesFunc (line 690):**
- Input `defnDeps`: `map[lg.NodeKey][]lg.NodeKey` (was `map[string][]string`)
- Closure signature: `func(map[lg.NodeKey]bool) map[lg.NodeKey]bool`

**Wherever these maps are built** (search for `symprops[`, `symwhens[`, `symwaits[`, `defnDeps[` in l2s/shared.go): use `lg.Key(sym)` as key.

### 8. `ivylogic/globals.go` — SymbolsAst dedup (line 265)

**symbolsAstRec:**
- Dedup `seen` map: `map[lg.NodeKey]bool` (was `map[string]bool`)
- Check: `seen[lg.Key(c)]` (was `seen[c.Name]`)

---

## Files Modified (summary)

| File | Change |
|------|--------|
| `actions/transforms.go` | `Modifies` returns `[]*lg.Symbol` |
| `compiler/phase6.go` | `collectFormulaSymbols` returns `map[lg.NodeKey]*lg.Symbol`; `GetSymbolDependencies` uses NodeKey; `CheckMutax` uses NodeKey |
| `compiler/ivy_compile.go` | Remove `definesPlainName`; interference section uses NodeKey throughout |
| `temporal/temporal.go` | `symprops` keyed by NodeKey; `symbolsAst` dedup by NodeKey |
| `isolate/helpers.go` | Extract `.Name` from Symbol for prefix check |
| `isolate/phase7.go` | Extract `.Name` from Symbol for sig lookup |
| `l2s/shared.go` | All symbol maps keyed by NodeKey; `Dependencies` uses NodeKey; extract `.Name` for sig lookups |
| `ivylogic/globals.go` | `SymbolsAst` dedup by NodeKey |
| `compiler/batch_f_test.go` | Update tests for new Modifies return type |

---

## Execution Order

1. **actions/transforms.go** — Change `Modifies` return type (breaks callers)
2. **compiler/phase6.go** — Fix `collectFormulaSymbols`, `GetSymbolDependencies`, `CheckMutax`
3. **compiler/ivy_compile.go** — Remove `definesPlainName`, unify on NodeKey
4. **ivylogic/globals.go** — Fix `SymbolsAst` dedup
5. **temporal/temporal.go** — Fix `symprops` + `symbolsAst` + Modifies caller
6. **isolate/helpers.go** — Fix GetLocMods
7. **isolate/phase7.go** — Fix sig lookup
8. **l2s/shared.go** — Fix Dependencies, symprops, symwhens, symwaits, sig lookups
9. **compiler/batch_f_test.go** — Update tests
10. Build: `go build ./...`
11. Test: `DYLD_LIBRARY_PATH=$(pwd)/z3/lib go test ./...`
12. Fix any failures immediately

## Verification

1. `go build ./...` — clean compile
2. `DYLD_LIBRARY_PATH=$(pwd)/z3/lib go test ./...` — all tests pass
3. Grep for `c.Name] = true` and `sym.Name] = true` in symbol-collection code to verify no stragglers
4. Grep for `map[string]bool` near Modifies/collectFormulaSymbols callers to verify all converted
