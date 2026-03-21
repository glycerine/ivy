# Plan: Complete InferParameters Port + Tests

## Context

`InferParameters` in `compiler/phase6.go:1302` is a skeletal port of Python's `infer_parameters` (`ivy_compiler.py:1578-1616`). Steps 1-2 (collecting action decls and mixin relationships) and parameter-count validation are done. But the critical Step 3 completion logic is missing — it just has `_ = nparms; _ = mnparms`.

### What's missing (Python lines 1607-1616):
1. `required = mnparms - nparms` validation ("monitor must supply at least N explicit input parameters")
2. Computing `xtraps` (extra input params) and `xtrars` (extra returns) from the mixee
3. Extending mixer's `FormalParams` and `FormalReturns` with extras
4. Building substitution map (unprefixed → prefixed name) and rewriting body via `SubstPrefixAtomsAst`

## Implementation

### Step 1: Complete `InferParameters` in `compiler/phase6.go`

Replace lines 1379-1383 (the skeletal stub) with the full logic. Key details:

- **Type assertion**: Must assert `a` to `*ast.ActionDef` to mutate `FormalParams`, `FormalReturns`, and `Body` directly (the `getFormalParams` helper returns a read-only copy via interface)
- **xtraps computation**: Python `(mixee.args[0].args + mixee.formal_params)[nparms+len(a.formal_params):]`
  - In Go: `mixee.Args()[0].Args()` = mixee name atom's Terms (signature params), concatenated with `mixee.FormalParams`, then skip first `nparms + len(aFormals)` elements
- **xtrars computation**: `mixee.FormalReturns[len(aReturns):]`
- **Substitution map**: For each extra node `x`, map `x.DropPrefix("fml:").Rep` → `x.Rep`
- **Body rewrite**: `ad.Body = ast.SubstPrefixAtomsAst(ad.Body, subst, nil, nil, nil)`

Also need to type-assert `mixee` to `*ast.ActionDef` to access its fields.

### Step 2: Create `compiler/infer_test.go`

Build AST nodes directly using existing constructors:

- `ast.NewActionDecl(actionDefs...)` — wraps ActionDefs in an ActionDecl (String() = "action")
- `ast.NewActionDef(name, body, params, returns)` — creates ActionDef with fml: prefix applied automatically
- `ast.NewMixinDecl(mixinDefs...)` — wraps mixin defs in MixinDecl (String() = "mixin")
- `ast.MixinAfterDef{MixerNode: atom, MixeeNode: atom}` — mixin definition
- `ast.NewAtom(rep, terms...)` — create atoms

**Important**: `NewActionDef` auto-prefixes params with `fml:` and rewrites the body. So when checking results, `FormalParams` will have `fml:`-prefixed names.

**Test cases:**

| # | Test | Description |
|---|------|-------------|
| 1 | `TestInferParameters_NoMixins` | No mixins → actions unchanged |
| 2 | `TestInferParameters_UndefinedAction` | Mixin refs nonexistent action → error |
| 3 | `TestInferParameters_SkipInit` | Mixin to "init" → no error, no inference |
| 4 | `TestInferParameters_TooManyInputParams` | Monitor has too many inputs → error |
| 5 | `TestInferParameters_TooManyOutputParams` | Monitor has too many outputs → error |
| 6 | `TestInferParameters_RequiredParamsNotMet` | Insufficient explicit params → error |
| 7 | `TestInferParameters_ExtendsFormals` | Extra params/returns appended to mixer |
| 8 | `TestInferParameters_BodyRewritten` | Body references rewritten with subst |
| 9 | `TestInferParameters_MultipleMixees` | >1 mixee → skip (no inference) |
| 10 | `TestInferParameters_NoExtrasNeeded` | All params supplied → no changes |

## Files to modify
- `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/phase6.go` — lines 1379-1383
- `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/infer_test.go` — new file

## Existing utilities to reuse
- `ast.NewAtom`, `ast.NewActionDef`, `ast.NewActionDecl`, `ast.NewMixinDecl` — constructors
- `ast.MixinAfterDef` / `ast.MixinBeforeDef` — mixin definition types
- `ast.SubstPrefixAtomsAst` — AST rewriting (`ast/rewrite.go:647`)
- `ast.Atom.DropPrefix` / `ast.Atom.Prefix` — prefix manipulation (`ast/ast.go:119`)
- `astDefines`, `astRelname` — name extraction (`compiler/phase6.go:1389`)
- `getFormalParams`, `getFormalReturns` — formal extraction (`compiler/phase6.go:1412`)
- `lg.IvyError` — error type (`logic/error.go`)

## Verification
1. `go test ./compiler/ -run TestInferParameters -v` — run new tests
2. `make test` or equivalent — ensure no regressions
