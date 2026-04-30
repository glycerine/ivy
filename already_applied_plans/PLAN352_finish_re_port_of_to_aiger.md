# Re-port ToAiger and ExpandSchemata to match Python

Created: 2026-04-30 ~07:30 UTC

## Context

Multiple bugs found in `mc/toaiger.go` and `mc/transforms.go`. D1-D7 are already applied and the XTRACE divergence is resolved (225,114 matching points). But Go crashes at the end in `ExpandSchemata` → `extractSchemaFormula` with:

```
panic: interface conversion: *ast.SchemaBody is not logic.Expr: missing method Children
```

The crash happens because Go's `ExpandSchemata` assumes schema formulas are `lg.Expr`, but Python handles `SchemaBody` objects (an `ast.Node`, not `lg.Expr`) that contain premises and a conclusion.

## Status of previously applied fixes

D1-D7 from the previous plan are already applied and working. The `actionNodeWrapper` removal is also applied. The only remaining issue is the crash in `ExpandSchemata`/`matchSchemaPremsNode`.

## D8. `ExpandSchemata` and `matchSchemaPremsNode` mishandle SchemaBody

### Python flow (ivy_mc.py:638-656)

```python
for name, lf in mod.schemata.items():
    schema = lf.formula          # SchemaBody: has .args = [*premises, conclusion]
    conc = schema.args[-1]       # conclusion is last element
    prems = list(schema.args[:-1])  # premises are everything else
    bound_sorts = [s for s in prems if isinstance(s, il.UninterpretedSort)]
    for m in match_schema_prems(prems, sort_constants, funs, match, bound_sorts):
        inst = apply_match(m, conc)
        res.append(LabeledFormula(Atom(name), inst))
```

Premises can be:
- `ConstantDecl` — wraps a symbol via `prem.args[0]`; Python checks `is_function_sort(sym.sort)` and either matches against `funs` or `sort_constants`
- `UninterpretedSort` — skipped (passed through)

### Go's current code (mc/transforms.go:257-316)

1. `extractSchemaFormula` (line 304) casts `LabeledFormula.Formula` to `lg.Expr` — **panics** when formula is `*ast.SchemaBody`
2. `ExpandSchemata` (line 277) calls `schema.Children()` — wrong API for `SchemaBody` (which has `Prems()` and `Conc()`)
3. `matchSchemaPremsNode` (line 319) takes `[]lg.Expr` — wrong type; premises are `[]ast.Node` containing `*ast.ConstantDecl` and `*lg.UninterpretedSort`
4. Go uses `lu.SubstituteByName` for conclusion substitution — Python uses `apply_match` which handles beta reduction for function symbols

### Fix

**File: `mc/transforms.go`**

#### 1. Rewrite `ExpandSchemata` to handle `SchemaBody`

Replace the `extractSchemaFormula` + `Children()` approach with direct `SchemaBody` handling:

```go
for name, lfNode := range mod.Schemata.All() {
    // skip rec[, lep[, ind[ prefixes
    lf, ok := lfNode.(*ast.LabeledFormula)
    if !ok { continue }
    sb, ok := lf.Formula.(*ast.SchemaBody)
    if !ok { continue }
    prems := sb.Prems()   // []ast.Node — already exists
    conc := sb.Conc()     // ast.Node — already exists
    if conc == nil { continue }
    // ... match premises, substitute conclusion
}
```

Use `SchemaBody.Prems()` and `SchemaBody.Conc()` (ast/decl_ast.go:917-934) — these already exist and match Python's `schema.args[:-1]` / `schema.args[-1]`.

#### 2. Rewrite `matchSchemaPremsNode` to accept `[]ast.Node`

Change signature to `matchSchemaPremsNode(prems []ast.Node, ...)` and handle:

- **`*ast.ConstantDecl`**: Extract symbol via `prem.Args()[0]`. Check if function sort → match against `funs` with sort unification. Otherwise → match against `sort_constants`.
  - Python: `prem.args[0]` gives the symbol; `sym.sort` gives the sort
  - Go: need to extract the Const/symbol from the ConstantDecl's first arg

- **`*lg.UninterpretedSort`**: Skip (yield through to next premise). Python (line 614-616) just recursively matches remaining premises.

#### 3. Port Python's `Match` class and `apply_match`

Python uses a `Match` class (ivy_mc.py:551-580) with `push`/`pop`/`add`/`unify`/`unify_lists` methods for backtracking unification. The Go code currently has no equivalent — `matchSchemaPremsRec` uses a flat map with manual save/restore.

Python's `apply_match` (ivy_mc.py:619-636) does proper beta reduction for matched function symbols, not simple name substitution. Go uses `lu.SubstituteByName` which doesn't handle function application.

The Match class needs porting:
- `Match.map` — the current substitution map
- `Match.push()` / `Match.pop()` — save/restore for backtracking
- `Match.add(x, y)` — add mapping x → y
- `Match.unify(s1, s2)` — unify two sorts
- `Match.unify_lists(l1, l2)` — unify two sort lists

And `apply_match` needs porting as a recursive function that handles:
- `is_app(fmla)` + `fmla.rep in match` → substitute and apply (beta reduction)
- `is_binder(fmla)` → clone with substituted variables and body
- `is_variable(fmla)` → remap sort
- default → clone with recursed args

#### 4. Revert the `extractSchemaFormula` hack

The comma-ok hack (`expr, ok := t.Formula.(lg.Expr)`) at transforms.go:310 silently skips schemata that need processing. This must be reverted and replaced with proper `SchemaBody` handling.

## Files to modify

- `mc/transforms.go` — `ExpandSchemata`, `matchSchemaPremsNode`, new `Match` type, new `applyMatch` function, remove `extractSchemaFormula`

## Existing Go functions to reuse

- `ast.SchemaBody.Prems()` — ast/decl_ast.go:917
- `ast.SchemaBody.Conc()` — ast/decl_ast.go:929
- `il.IsFunctionSort` — ivylogic (already imported)

## Python reference files

- `ivy_mc.py:551-580` — `Match` class
- `ivy_mc.py:582-617` — `match_schema_prems`
- `ivy_mc.py:619-636` — `apply_match`
- `ivy_mc.py:638-656` — `expand_schemata`

## Verification

`cd ~/ivy/goivy && make test` then `make rfn` — Go should no longer crash and should produce a result (PASS or at least exit cleanly without panic).
