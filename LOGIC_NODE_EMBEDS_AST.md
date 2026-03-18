# Plan: Make logic.Expr types also satisfy ast.Node

## Context

The goivy codebase has two separate interface hierarchies that don't interoperate:
- `logic.Expr` (was `logic.Node`) — semantic IR (26 types in logic/, 3 in ivylogic/)
- `ast.Node` — syntax tree IR (86 types in ast/)

Python has one class hierarchy. The Go split forces 4 duplicate adapter wrappers (`logicNodeAdapter`/`logicASTAdapter`) across proof/, temporal/, l2s/ packages, and makes type conversions between `module.LabeledFormula` (uses `lg.Expr`) and `ast.LabeledFormula` (uses `ast.Node`) fail silently.

**Fix:** Add the 4 missing `ast.Node` methods to all `logic.Expr` concrete types so they structurally satisfy both interfaces. No interface embedding needed — just duck typing.

**Confirmed: no circular import.** ast/ imports nothing from logic/; logic/ imports nothing from ast/. logic/ CAN safely import ast/.

## What ast.Node requires that logic.Expr types lack

| Method | How to provide |
|--------|---------------|
| `GetLineno() ast.Location` | Embed `ast.Base` in each struct (free) |
| `SetLineno(ast.Location)` | Embed `ast.Base` in each struct (free) |
| `Args() []ast.Node` | New method per type — wraps `Children()` |
| `Clone([]ast.Node) ast.Node` | New method per type — from existing `CloneNode` switch |

`String()` is already in both interfaces — no change needed.

## Types to modify (28 total)

**logic/ package (25 types, 4 files):**

| File | Types |
|------|-------|
| `logic/sort.go` | UninterpretedSort, BooleanSort, FunctionSort, EnumeratedSort, TopSort, RangeSort |
| `logic/term.go` | Variable, Symbol, Apply |
| `logic/formula.go` | Eq, Not, And, Or, Implies, Iff, Ite, ForAll, Exists, Lambda, NamedBinder, Globally, Eventually, WhenOperator, Cond |
| `logic/definition.go` | Definition (DefinitionSchema embeds it, inherits methods) |

**ivylogic/ package (3 types, 1 file):**

| File | Types |
|------|-------|
| `ivylogic/formula.go` | Some, Let, Literal |

## Implementation

### Step 1: Add `ast` import to logic/

Add `import "github.com/glycerine/goivy/ast"` to `logic/term.go`, `logic/formula.go`, `logic/definition.go` (whichever files need `ast.Base`). `logic/sort.go` also needs it for sort types.

### Step 2: Embed `ast.Base` in each struct

Example for `Eq`:
```go
type Eq struct {
    ast.Base       // provides GetLineno/SetLineno
    T1, T2 Expr
}
```

**Leaf types** (sorts, Variable, Symbol) — still embed ast.Base even though they have no children. They may carry source location info.

**DefinitionSchema** embeds `Definition` which gets `ast.Base`, so it inherits automatically.

### Step 3: Add Args() and Clone() — new file `logic/ast_compat.go`

Keep all ast.Node bridge methods in one file for clean separation.

**Args()** wraps Children() with slice conversion:
```go
// Leaf types:
func (v *Variable) Args() []ast.Node { return nil }
func (c *Symbol) Args() []ast.Node   { return nil }

// Compound types:
func (e *Eq) Args() []ast.Node       { return []ast.Node{e.T1, e.T2} }
func (a *And) Args() []ast.Node {
    r := make([]ast.Node, len(a.Terms))
    for i, t := range a.Terms { r[i] = t }
    return r
}
// ... etc for all 25 types
```

This works because after Step 2+3, every `logic.Expr` value also satisfies `ast.Node`, so assigning `logic.Expr` → `ast.Node` is valid.

**Clone()** converts the existing `ivylogic.CloneNode` switch cases into per-type methods:
```go
func (e *Eq) Clone(args []ast.Node) ast.Node {
    if len(args) == 2 {
        return &Eq{T1: args[0].(Expr), T2: args[1].(Expr)}
    }
    return e
}
func (c *Symbol) Clone([]ast.Node) ast.Node { return c }
// ... etc
```

### Step 4: Same for ivylogic/ — new file `ivylogic/ast_compat.go`

Add `ast.Base` to Some, Let, Literal structs. Add Args()/Clone() methods. Clone logic for Some already exists in `ivylogic/util.go:CloneNode`; Let and Literal need new clone logic.

### Step 5: Remove adapter duplicates

| File | Remove |
|------|--------|
| `proof/goal.go:329-340` | `logicNodeAdapter` struct, `concToASTNode` func |
| `proof/phase5_matching.go:~558` | `logicNodeAdapter` duplicate, `wrapLogicNode` |
| `proof/skolem.go` | `constToASTNode` |
| `l2s/l2s.go:573-586` | `logicASTAdapter`, `wrapLogicAsAST` |
| `l2s/shared.go` | `WrapLogicAsAST` |
| `temporal/temporal.go:629-643` | `logicASTAdapter`, `wrapLogicAsAST` |

Replace all `&logicNodeAdapter{node: n}` / `wrapLogicAsAST(n)` with just `n` (direct assignment since `lg.Expr` now IS `ast.Node`).

Replace all unwrap patterns like `a.(*logicNodeAdapter).Unwrap()` or `a.(*logicNodeAdapter).node` with direct `lg.Expr` type assertion.

### Step 6: Simplify check/helpers.go conversion functions

`ModuleLFToAstLF` currently does `if an, ok := mlf.Formula.(ast.Node)` — this now always succeeds for `lg.Expr` values. Simplify. Similarly `AstLFToModuleLF` reverse direction becomes trivial.

## Struct initialization impact

Adding `ast.Base` changes struct literals. Anywhere that uses positional initialization like `&Eq{t1, t2}` must become `&Eq{T1: t1, T2: t2}` (named fields). Check for this — most likely all uses already use named fields since the types have multiple fields.

Also: constructors like `NewDefinition(lhs, rhs)` don't set Base, which is correct — Base zero value is fine (no location).

## Files to modify

| File | Changes |
|------|---------|
| `logic/sort.go` | Embed ast.Base in 6 sort structs |
| `logic/term.go` | Embed ast.Base in Variable, Symbol, Apply |
| `logic/formula.go` | Embed ast.Base in 14 formula structs |
| `logic/definition.go` | Embed ast.Base in Definition |
| `logic/ast_compat.go` | NEW — Args()/Clone() for all 25 types |
| `ivylogic/formula.go` | Embed ast.Base in Some, Let, Literal |
| `ivylogic/ast_compat.go` | NEW — Args()/Clone() for 3 types |
| `proof/goal.go` | Remove logicNodeAdapter + concToASTNode |
| `proof/phase5_matching.go` | Remove duplicate adapter + wrapLogicNode |
| `proof/skolem.go` | Remove constToASTNode |
| `l2s/l2s.go` | Remove logicASTAdapter + wrapLogicAsAST |
| `l2s/shared.go` | Remove WrapLogicAsAST |
| `temporal/temporal.go` | Remove logicASTAdapter + wrapLogicAsAST |
| `check/helpers.go` | Simplify ModuleLFToAstLF/AstLFToModuleLF |

## Verification

1. `go build ./...` — everything compiles
2. `go vet ./...` — no issues
3. `DYLD_LIBRARY_PATH=z3/build go test ./logic/ ./ivylogic/ ./proof/ ./check/ ./l2s/ ./temporal/` — existing tests pass
4. Spot-check: `var _ ast.Node = (*logic.Eq)(nil)` compiles (interface satisfaction)
5. `DYLD_LIBRARY_PATH=z3/build go test -tags web ./webui/` — webui tests pass
