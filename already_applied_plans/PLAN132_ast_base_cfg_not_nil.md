# Fix All Missing ast.Base.Cfg Initialization Across Codebase
Created: 2026-03-28 (afternoon)

## Context

Golden test crashes at line 142255 with nil pointer dereference in `SubstituteConstantsAst2` at `ast/rewrite.go:922`, accessing `n.Cfg.IuCfg` on an `*App` node whose `Base.Cfg` is nil. The App was created by `ToConstApp()` which faithfully copies `a.Base` from its input — but the input App (an action formal parameter from the parser) had nil Cfg because it was created as a struct literal in the grammar.

This is a systemic issue: many AST nodes are created via struct literals (`&ast.Symbol{Rep: "="}`, `&ast.Variable{Rep: "X", VSort: "S"}`, etc.) instead of through `*AstConfig` constructor methods that set `Base.Cfg`. Every such creation is a latent nil-pointer crash.

## Scope

**Every** AST node creation must go through an `*AstConfig` constructor or explicitly set `Cfg` afterward. No exceptions for "just a sort annotation" or "just a Symbol inside App.Rep".

## Inventory of violations

### A. Grammar files (PRIMARY — these create all parsed AST nodes)

#### A1. `lalr_full/grammar_v17.y` — 39 struct literal creations

**Symbol** (9 sites):
- Lines 1711, 1717, 1719, 1721: `&ast.Symbol{Rep: ...}` in `symbol` rule → `acfg(v17lex).NewSymbol(..., nil)`
- Lines 1742, 1750: `&ast.Symbol{Rep: ...}` inside `acfg(v17lex).NewApp(...)` → `acfg(v17lex).NewSymbol(..., nil)`
- Lines 2965, 2967: `&ast.Symbol{Rep: "bool"}` sort annotation → `acfg(v17lex).NewSymbol("bool", nil)`
- Line 3439: `&ast.Symbol{Rep: $3.Val}` sort annotation → `acfg(v17lex).NewSymbol($3.Val, nil)`

**Variable** (5 sites):
- Lines 1783, 1791, 1802, 1810: `&ast.Variable{Rep: ..., VSort: ...}` → `acfg(v17lex).NewVariable(..., ...)`
- Line 4812: `&ast.Variable{...}` inside NewDefinition call → `acfg(v17lex).NewVariable(...)`

**Atom** (7 sites):
- Lines 1966, 1973, 1980, 1987, 1994: `&ast.Atom{Rep: "op", Terms: ...}` comparison operators → `acfg(v17lex).NewAtom("op", ...)`
- Line 2935: `&ast.Atom{Rep: $1.Val, Terms: $2}` → `acfg(v17lex).NewAtom($1.Val, $2...)`
- Line 3149: `&ast.Atom{Rep: $1.Val, Terms: $3}` → `acfg(v17lex).NewAtom($1.Val, $3...)`

**Not** (4 sites):
- Line 2008: `&ast.Not{Body: &ast.Atom{...}}` → `acfg(v17lex).NewNot(acfg(v17lex).NewAtom(...))`
- Line 2030: `&ast.Not{Body: $2}` → `acfg(v17lex).NewNot($2)`
- Line 5286: `&ast.Not{Body: acfg(v17lex).NewAtom(...)}` → `acfg(v17lex).NewNot(...)`
- Line 5293: `&ast.Not{Body: $2}` → `acfg(v17lex).NewNot($2)`

**Forall/Exists** (4 sites):
- Lines 2078, 2092: `&ast.Forall{...}` → `acfg(v17lex).NewForall(...)`
- Lines 2085, 2099: `&ast.Exists{...}` → `acfg(v17lex).NewExists(...)`

**Ite** (1 site):
- Line 1958: `&ast.Ite{...}` → `acfg(v17lex).NewIte(...)`

**Sequence** (5 sites):
- Lines 3515, 3839, 5396: `&ast.Sequence{}` → `acfg(v17lex).NewSequence()`
- Lines 3866, 5402: `&ast.Sequence{Stmts: stmts}` → `acfg(v17lex).NewSequence(stmts...)`

**NamedBinder** (3 sites):
- Lines 2189, 2197, 2202: `&ast.NamedBinder{...}` → `acfg(v17lex).NewNamedBinder(...)`

**NativeType** (1 site):
- Line 5375: `&ast.NativeType{Elems: elems}` → need `NewNativeType` constructor (doesn't exist yet)

#### A2. `lalr_logicparser/grammar_v17.y` — 33 struct literal creations

Has `acfg(v17lex)` available. Same pattern of fixes as A1:
- Symbol: lines 183, 188, 190, 192 + inside App at 209, 214
- Variable: lines 223, 227, 235, 239
- Atom: lines 306, 314, 352, 356, 360, 364, 368, 376, 481, 523, 528, 535, 557, 563
- App: lines 209, 214 → `acfg(v17lex).NewApp(acfg(v17lex).NewSymbol(..., nil), ...)`
- Not: lines 376, 389
- Forall: lines 425, 433
- Exists: lines 429, 437
- Ite: line 347
- NamedBinder: lines 480, 485, 489

#### A3. `lalr_logicparser/v16/grammar_v16.y` — 24 struct literal creations

**Does NOT have `acfg()` yet.** Must add:
1. Add `cfg *ast.AstConfig` field to `v16LexAdapter` struct in `v16/lalr_v16.go`
2. Add `acfg()` helper function in `v16/grammar_v16.y`
3. Update `ParseV16` signature to accept `cfg *ast.AstConfig`; create default if nil
4. Update `newV16LexAdapter` to accept and store cfg
5. Update dispatch in `lalr_logicparser/lalr_parser.go` line 18 to pass cfg

Sites: Symbol (4), Atom (6), Variable (4), Ite (1), Not (2), Forall (1), Exists (1), App (2), plus arithmetic operators (4)

#### A4. `lalr_logicparser/v12/grammar_v12.y` — 15 struct literal creations

**Does NOT have `acfg()` yet.** Same plumbing as A3:
1. Add `cfg *ast.AstConfig` field to `v12LexAdapter` struct in `v12/lalr_v12.go`
2. Add `acfg()` helper in `v12/grammar_v12.y`
3. Update `ParseV12` signature; update dispatch in `lalr_parser.go` line 18
4. Update `newV12LexAdapter`

Sites: Symbol (1), Atom (5), Variable (4), Not (2), Forall (1), Exists (1)

### B. ast/ package internal struct literals

#### B1. `ast/rewrite.go` — ~20 struct literals
Most already set Cfg immediately after creation (lines 550-642). Fix the ones that don't:
- Line 231, 267, 280: `&Atom{Rep: hname, Terms: args}` in `ComposeAtomsGeneric` — some set Base, verify all do
- Line 437: `&Atom{Rep: thePref.Rep}` — needs Cfg
- Line 489: `&Atom{Rep: sort}` — needs Cfg
- Line 907: `&Atom{Rep: rest}` — needs Cfg (has `res.Base = n.Base` at 910 ✓)
- Line 927: `&App{Rep: &Symbol{Rep: rest}}` — inner Symbol needs Cfg (App gets `thing.Base = n.Base` ✓)
- Line 996: `&Variable{Rep: newName, VSort: v.VSort}` — has `nv.Cfg = v.Cfg` at 997 ✓

#### B2. `ast/ast.go` — ToConstApp and ToConstAtom
- Line 499: `ToConstAtom` — creates `&Atom{...}` then `res.Base = a.Base` ✓ (Base copy propagates Cfg)
- Line 513: `ToConstApp` — creates `&Symbol{Rep: rep1}` WITHOUT Cfg ← **FIX**: set `newRep.Cfg = a.Cfg`
- Line 517: `&App{Rep: newRep, Terms: a.Terms}` then `res.Base = a.Base` ✓

#### B3. `ast/labeler.go` line 20
- `return &Atom{Rep: name}` — needs Cfg. Labeler must carry `*AstConfig` or receive it as param.

#### B4. `ast/lower_var.go` — 5 struct literals
- Line 91: `&Sequence{Stmts: lines}` + `seq.Cfg = t.Cfg` ✓
- Lines 111, 116, 118, 123, 127: `res.Base = a.Base` ✓ — but inner `&Symbol{Rep: ...}` at line 116 needs Cfg

#### B5. `ast/decl_ast.go`
- Line 385: `&App{Rep: &Symbol{Rep: ...}}` — inner Symbol needs Cfg; App gets `Base` from context?
- Line 388: `&Symbol{Rep: a.VSort}` — needs Cfg
- Line 1782: `&Atom{Rep: s.Name}` — needs Cfg

#### B6. `ast/formula.go` line 341
- `&Atom{Rep: "=", Terms: ...}` in `Definition.ToConstraint()` — has `a.Cfg = d.Cfg` ✓
- `&Iff{T1: ..., T2: ...}` same function — has `iff.Cfg = d.Cfg` ✓

#### B7. `ast/ast.go` line 1648
- `&Variable{Base: v.Base, Rep: v.Rep, VSort: sortNode}` — Base copy includes Cfg ✓

### C. compiler/ package

#### C1. `compiler/action.go` lines 889, 893
- `v.ASort = &ast.Symbol{Rep: "S"}` — the `ensureSortAnnotation` helper. Needs access to cfg to create Symbol via constructor.

#### C2. `compiler/helpers.go` line 70
- `&ast.Atom{Rep: resolved}` — needs Cfg. Compiler has access to `c.AstCfg()` or similar.

#### C3. `compiler/compiler.go` lines 1224, 1226
- `&ast.Ite{...}` and `&ast.Forall{...}` — use constructor via compiler's AstConfig.

#### C4. `compiler/compiler.go` lines 1381, 1401 and `compiler/phase6.go` ~12 sites
- `&ast.CompiledNode{Node: compiled}` — need `NewCompiledNode` constructor (doesn't exist yet), or set Cfg after creation.

#### C5. `compiler/ivy_compile.go` lines 1204, 1288
- `&ast.Atom{Rep: ...}` — use constructor via mod's AstConfig.

#### C6. `compiler/compiler.go` line 1403
- `&ast.Atom{Base: atom.Base, Rep: ..., Terms: ..., ASort: ...}` — Base copy preserves Cfg ✓

### D. Test files

#### D1. `compiler/expr_test.go` lines 188, 225, 263, 291, 293, 377
- `&ast.Symbol{Rep: "nat"}` etc. → `cfg.NewSymbol("nat", nil)`

#### D2. `compiler/solo7_test.go` line 178
- `&ast.App{Rep: cfg.NewAtom("f"), Terms: ...}` → `cfg.NewApp(cfg.NewSymbol("f", nil), ...)`

#### D3. `compiler/decl_missing_test.go` line 395, `decl_missing2_test.go` line 339
- `&ast.NativeType{...}` → use new `NewNativeType` constructor

#### D4. `isolate/batch_fixes_test.go` lines 91, 99, 107, 725
- `&ast.Atom{...}`, `&ast.Symbol{...}`, `&ast.Variable{...}` → use constructors

### E. Global `ast.Equals` variable
- `ast/ast.go:1552`: `var Equals = &Symbol{Rep: "=", Sort: &RelationSort{...}}`
- This is a read-only singleton. The inner `&RelationSort{Dom: []Node{nil, nil}}` also lacks Cfg.
- Per CLAUDE.md rule C.9, read-only state is exempt. But if Equals is ever cloned, its nil Cfg propagates. **Add a comment documenting this exemption, or initialize with a dedicated config.**

## New constructors needed

1. **`NewNativeType`** on `*AstConfig` in `ast/decl_ast.go`:
   ```go
   func (cfg *AstConfig) NewNativeType(elems ...Node) *NativeType {
       n := &NativeType{Elems: elems}
       n.Cfg = cfg
       return n
   }
   ```

2. **`NewCompiledNode`** on `*AstConfig` in `ast/ast.go`:
   ```go
   func (cfg *AstConfig) NewCompiledNode(node interface{}) *CompiledNode {
       c := &CompiledNode{Node: node}
       c.Cfg = cfg
       return c
   }
   ```

## Plumbing needed for v16/v12 grammars

### v16 (`lalr_logicparser/v16/`)

**lalr_v16.go**:
```go
type v16LexAdapter struct {
    lex    *lexer.Lexer
    cfg    *ast.AstConfig  // ADD
    result ast.Node
    err    string
}

func newV16LexAdapter(input string, version lexer.Version, cfg *ast.AstConfig) *v16LexAdapter {
    if cfg == nil {
        cfg = ast.NewAstConfig()
    }
    return &v16LexAdapter{lex: lexer.New(input, version), cfg: cfg}
}

func ParseV16(input string, version lexer.Version, cfg ...*ast.AstConfig) (ast.Node, error) {
    var c *ast.AstConfig
    if len(cfg) > 0 { c = cfg[0] }
    lex := newV16LexAdapter(input, version, c)
    ...
}
```

**grammar_v16.y** — add at top:
```go
func acfg(lex v16Lexer) *ast.AstConfig {
    return lex.(*v16LexAdapter).cfg
}
```

### v12 (`lalr_logicparser/v12/`) — identical pattern

### lalr_parser.go dispatch — update lines 18, 20 to pass cfg:
```go
func Parse(input string, version lexer.Version, cfg ...*ast.AstConfig) (ast.Node, error) {
    var c *ast.AstConfig
    if len(cfg) > 0 { c = cfg[0] }
    if version[0] < 1 || (version[0] == 1 && version[1] <= 2) {
        return v12.ParseV12(input, version, c)
    }
    if version[0] == 1 && version[1] <= 6 {
        return v16.ParseV16(input, version, c)
    }
    return ParseV17(input, version, c)
}
```

## Labeler fix (`ast/labeler.go`)

Labeler.Call() creates `&Atom{Rep: name}` without Cfg. Options:
- Add `Cfg *AstConfig` field to Labeler struct, set it when Labeler is created, use `lb.Cfg.NewAtom(name)` in Call().
- Check existing Labeler creation sites to thread cfg.

## `ensureSortAnnotation` fix (`compiler/action.go`)

Currently creates `&ast.Symbol{Rep: "S"}` without Cfg. Change to accept AstConfig param or use compiler's config:
```go
func ensureSortAnnotation(n ast.Node, cfg *ast.AstConfig) {
    switch v := n.(type) {
    case *ast.Atom:
        if v.ASort == nil { v.ASort = cfg.NewSymbol("S", nil) }
    case *ast.App:
        if v.ASort == nil { v.ASort = cfg.NewSymbol("S", nil) }
    }
}
```

## Execution order

1. **Create new constructors**: `NewNativeType`, `NewCompiledNode` in ast/
2. **Fix ast/ internal**: rewrite.go, labeler.go, lower_var.go, decl_ast.go, ast.go (ToConstApp inner Symbol)
3. **Fix compiler/**: action.go, helpers.go, compiler.go, phase6.go, ivy_compile.go
4. **Add v16/v12 plumbing**: lexer adapter cfg field, acfg(), ParseV16/V12 signature
5. **Fix grammar .y files**: lalr_full/grammar_v17.y, lalr_logicparser/grammar_v17.y, v16/grammar_v16.y, v12/grammar_v12.y
6. **Regenerate .go from .y**: `go generate ./lalr_full/` and `go generate ./lalr_logicparser/...`
7. **Fix test files**: expr_test.go, solo7_test.go, decl_missing_test.go, decl_missing2_test.go, isolate/batch_fixes_test.go
8. **Build and test**: `cd ~/goivy && go build ./...` then `go test ./ast/ ./compiler/ ./lalr_full/ ./lalr_logicparser/...`
9. **Run golden**: `cd ~/goivy && make golden` to verify crash is fixed and check new divergence point

## Files to modify

- `ast/ast.go` — NewCompiledNode constructor, ToConstApp fix
- `ast/decl_ast.go` — NewNativeType constructor, fix lines 385/388/1782
- `ast/labeler.go` — add Cfg field, use constructor in Call()
- `ast/rewrite.go` — fix remaining bare struct literals (lines 437, 489, 927 inner Symbol)
- `ast/lower_var.go` — fix line 116 inner Symbol
- `lalr_full/grammar_v17.y` — convert ~39 struct literals to constructor calls
- `lalr_logicparser/grammar_v17.y` — convert ~33 struct literals to constructor calls
- `lalr_logicparser/v16/lalr_v16.go` — add cfg plumbing
- `lalr_logicparser/v16/grammar_v16.y` — add acfg(), convert ~24 struct literals
- `lalr_logicparser/v12/lalr_v12.go` — add cfg plumbing
- `lalr_logicparser/v12/grammar_v12.y` — add acfg(), convert ~15 struct literals
- `lalr_logicparser/lalr_parser.go` — pass cfg to v16/v12 parse functions
- `compiler/action.go` — fix ensureSortAnnotation
- `compiler/helpers.go` — fix line 70
- `compiler/compiler.go` — fix lines 1224, 1226, 1381, 1401
- `compiler/phase6.go` — fix ~12 CompiledNode creations
- `compiler/ivy_compile.go` — fix lines 1204, 1288
- `compiler/expr_test.go` — fix ~6 bare Symbol creations
- `compiler/solo7_test.go` — fix line 178
- `compiler/decl_missing_test.go` — fix line 395
- `compiler/decl_missing2_test.go` — fix line 339
- `isolate/batch_fixes_test.go` — fix lines 91, 99, 107, 725

## Verification

```bash
# Build everything
cd ~/goivy && go build ./...

# Run ast package tests
go test ./ast/ -v -count=1

# Run compiler tests
go test ./compiler/ -v -count=1

# Run logic parser tests
go test ./lalr_logicparser/... -v -count=1

# Run golden test
cd ~/goivy && make golden
```
