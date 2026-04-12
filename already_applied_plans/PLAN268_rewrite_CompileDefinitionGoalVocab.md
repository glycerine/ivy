# Fix CompileDefinitionGoalVocab to match Python compile_definition_goal_vocab

Created: 2026-04-12 ~05:30 UTC

## Context

The golden test `TestOrdLive` diverges at xtrace line 255399. Go emits `compileDefn[1]` immediately after `compileDefn[0]`, while Python emits `ast.LF.clone PRESERVE` traces inside `compileDefn[0]`. This means Go's `CompileDefinitionGoalVocab` is doing nothing — it hits an early return because it expects the inner formula to already be compiled (`lg.Expr`), but at this point in the l2s tactic the formula is still an AST-level `*ast.Definition`.

## Root Cause

`proof/goal.go:401-480` — `CompileDefinitionGoalVocab` tries to type-assert `innerLF.Formula.(lg.Expr)` at line 419. Since the formula is `*ast.Definition` (an AST node, not a logic expression), the assertion fails and the function returns `goal` unchanged at line 423. All subsequent work (compilation, normalization, premise creation) is skipped.

## Python reference (ivy_proof.py:1477-1506)

```python
def compile_definition_goal_vocab(df, goal):
    vocab = goal_vocab(goal)
    free = goal_free(goal)
    lf = df.args[0]                                    # inner LabeledFormula
    lhs = lf.formula.args[0]                           # Definition.lhs
    ts = il.TopFunctionSort(len(lhs.args))
    newsym = il.Symbol(lhs.rep, ts)
    with il.WithSymbols([newsym]):
        vars = lf.formula.args[0].args                 # lhs variable args
        body = ia.Atom('=', lf.formula.args)            # Atom("=", [lhs, rhs])
        fmla = ia.Forall(vars, body) if vars else body  # wrap in Forall if vars
        elf = lf.clone([lf.label, fmla])                # → LF.clone PRESERVE #1
        lf = compile_expr_vocab(elf, vocab)             # compiles LF → PRESERVE #2 (inside _labeled_formula_cmpl)
        thing = lf.formula.body if vars else lf.formula # body of compiled Forall
        thing = il.normalize_ops(thing)                 # normalize
        lf = lf.clone([lf.label, lf.formula.clone([thing]) if vars else thing])  # → PRESERVE #3
        sym = thing.args[0].rep
        deps = list(lu.symbols_ilu_ast(thing.args[1]))
        if sym in deps:
            raise NoMatch(lf, "no proof given for recursive definition")
        cd = ia.ConstantDecl(sym)
        cd.lineno = lf.lineno
        lf.definition = True
        goal = goal_add_prem(goal, cd, lf.lineno)       # add ConstantDecl premise
        goal = goal_add_prem(goal, lf, lf.lineno)       # add LF premise
        if sym in vocab.sorts or sym in vocab.symbols or sym in free:
            raise Redefinition(df, "redefinition of {}".format(sym))
        return goal
```

## Plan

### Step 1: Rewrite `CompileDefinitionGoalVocab` in `proof/goal.go`

Change signature from:
```go
func CompileDefinitionGoalVocab(cfg *ast.AstConfig, df ast.Node, goal *ast.LabeledFormula) *ast.LabeledFormula
```
to:
```go
func CompileDefinitionGoalVocab(cfg *ast.AstConfig, df ast.Node, goal *ast.LabeledFormula, mod *module.Module) (*ast.LabeledFormula, error)
```

Rewrite the body to faithfully port Python's logic:

1. **Unwrap** the DerivedDecl to get `innerLF *ast.LabeledFormula` (keep existing unwrap logic)
2. **Extract** `defFormula := innerLF.Formula.(*ast.Definition)` — handle the AST Definition type
3. **Get LHS info**: `lhsAtom` from `defFormula.Lhs`, get its `Rep` and `Terms` (variable args)
4. **Create TopFunctionSort symbol**: `ts := il.TopFunctionSort(len(lhsAtom.Terms))`, create a `*lg.Const` for the defined symbol
5. **Push symbol** via `il.NewWithSymbols(sig, [newsym])` Enter/defer Exit
6. **Reconstruct formula**:
   - `body := cfg.NewAtom("=", defFormula.Lhs, defFormula.Rhs)` (Atom with "=" rep and [lhs, rhs] as terms)
   - If `len(vars) > 0`: `fmla := cfg.NewForall(vars, body)`; else `fmla := body`
7. **Clone LF #1**: `elf := innerLF.Clone([]ast.Node{innerLF.Label, fmla}).(*ast.LabeledFormula)` → produces PRESERVE trace
8. **Compile the LF**: Set up vocab context (WithSymbols, WithSorts, TopSort default), create `compiler.New(sig, mod)`, call `c.CompileLF(elf)` → produces PRESERVE trace #2. Then run `il.SortInferList` on the compiled formula + vocab.Variables.
9. **Get body**: If vars existed, extract `compiledLF.Formula.(*lg.ForAll).Body`; else `compiledLF.Formula.(lg.Expr)`
10. **Normalize**: `thing = il.NormalizeOps(thing)`
11. **Clone LF #3**: If vars: clone the ForAll with normalized body via `compiledFormula.Clone([]ast.Node{thing})`, then clone LF; else clone LF with thing directly → PRESERVE trace #3
12. **Get defined symbol**: Extract `sym` from `thing` (the LHS of the equality in the normalized body)
13. **Check recursion**: Use `il.SymbolsIluAst` on the RHS; if sym appears, return `NoMatch` error
14. **Create ConstantDecl**: `cd := cfg.NewConstantDecl(sym_node)`, set lineno
15. **Set definition flag**: `compiledLF.IsDefinition = true`
16. **Add premises**:
    - `goal = GoalAddPrem(cfg, goal, cd, loc)`
    - `goal = GoalAddPrem(cfg, goal, compiledLF, loc)`
17. **Check redefinition**: Check if sym is in vocab.Sorts, vocab.Symbols, or free; if so, return `Redefinition` error
18. **Return** `goal, nil`

### Step 2: Update call site in `check/l2s.go:410`

Change from:
```go
goal = proof.CompileDefinitionGoalVocab(m.Cfg.AstCfg, defn, goal)
```
to:
```go
var err error
goal, err = proof.CompileDefinitionGoalVocab(m.Cfg.AstCfg, defn, goal, m)
if err != nil {
    return nil, err
}
```

### Step 3: Add necessary imports to `proof/goal.go`

Add imports for `compiler`, `il` (ivylogic), `lg` (logic) packages as needed by the new implementation.

## Key files

| File | Role |
|------|------|
| `proof/goal.go:401-480` | **PRIMARY** — rewrite `CompileDefinitionGoalVocab` |
| `check/l2s.go:408-411` | Update call site to pass module + handle error |
| `proof/phase5_matching.go:24-78` | Reference: `CompileExprVocab` pattern for sig/vocab setup |
| `compiler/compiler.go:1011-1049` | Reference: `CompileLF` implementation |
| `ast/decl_ast.go:72-91` | Reference: `LabeledFormula.Clone` with PRESERVE |
| `ivylogic/util.go:467` | Existing `NormalizeOps` to reuse |
| `ivylogic/symbols.go:39` | Existing `SymbolsIluAst` to reuse |
| `proof/goal.go:325-328` | Existing `GoalAddPrem` to reuse |

## Existing functions to reuse

- `proof.GoalVocab(goal)` → `*Vocab` (goal.go:211)
- `proof.GoalFree(goal)` → `map[lg.NodeKey]lg.Expr` (goal.go:272)
- `proof.GoalAddPrem(cfg, goal, prem, loc)` → `*ast.LabeledFormula` (goal.go:325)
- `il.TopFunctionSort(arity)` → `lg.Sort` (ivylogic/ivylogic.go:86)
- `il.NewWithSymbols(sig, syms)` → `*WithSymbols` (ivylogic/sig.go:385)
- `il.NewWithSorts(sig, sorts)` → `*WithSorts` (ivylogic/sig.go)
- `il.NormalizeOps(fmla)` → `lg.Expr` (ivylogic/util.go:467)
- `il.SymbolsIluAst(node)` → `iter.Seq[lg.Expr]` (ivylogic/symbols.go:39)
- `il.SortInferList(terms, nil, nil)` → `([]lg.Expr, error)` (ivylogic/)
- `compiler.New(sig, mod)` → `*Compiler` (compiler/compiler.go)
- `c.CompileLF(lf)` → `(*ast.LabeledFormula, error)` (compiler/compiler.go:1011)
- `cfg.NewAtom(rep, terms...)` — ast Atom constructor
- `cfg.NewForall(bounds, body)` — ast Forall constructor
- `cfg.NewConstantDecl(args...)` — ast ConstantDecl constructor
- `getSigFrom(mod)` — proof/phase5_matching.go helper (may need to export or duplicate)

## Verification

Run the golden test that exposed this divergence:
```
cd ~/ivy/goivy && go test -run TestOrdLive -v -timeout 600s
```

The test should advance past xtrace line 255399 (the current divergence point). The Go output should now show `ast.LF.clone PRESERVE` traces inside `compileDefn[0]` matching the Python output.
