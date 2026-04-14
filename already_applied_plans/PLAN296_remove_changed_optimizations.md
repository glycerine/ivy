# Plan: Eradicate all `changed` clone optimizations (Round 10)

Created: 2026-04-14 ~06:45 UTC

## Context

Round 9 fixed duplicate symbols in `BuildDefnDeps`. After `make golden`,
new divergence at line 721105: Go emits `instrStmt mods=[] deps=[]` where
Python emits `LocalAction.__init__ uniqueID=1359 caller=ast.LocalAction.clone`.

Root cause: Go has a `changed` optimization that skips `ActionClone` when
sub-args are unchanged. Python ALWAYS calls `stmt.clone(args)` unconditionally.
This pattern exists across **28 functions in 14 files**.

**Python NEVER uses a `changed` optimization for clone operations.** Every
single Python recursive transform does `return ast.clone(transformed_args)`
unconditionally. Verified exhaustively across all 23 .py files. This is a
mechanical port — the Go must not optimize what Python does not.

## Immediate bug (causing trace divergence NOW)

### 1. `check/l2s_shared.go:532-551` — instrStmt in SharedStep7

Python `ivy_l2s.py:1235-1236`:
```python
args = [instr_stmt(a,labels) if isinstance(a,Action) else a for a in stmt.args]
res = stmt.clone(args)   # ALWAYS clones
```

Go skips `ActionClone` when `changed` is false. `LocalAction.ActionClone`
emits traces; skipping it causes trace stream misalignment.

### 2. `temporal/temporal.go:622-641` — instrStmt in InvarianceTactic

Python `ivy_temporal.py:339-340`:
```python
args = [instr_stmt(a,labels) if isinstance(a,iact.Action) else a for a in stmt.args]
res = stmt.clone(args)   # ALWAYS clones
```

Identical bug pattern. Same fix needed.

## All other `changed` clone optimizations to remove

Every instance below has the same pattern: Go tracks `changed`, conditionally
clones. Python ALWAYS clones unconditionally. All must be fixed.

### 3. `check/l2s.go:1058` — Desugar

Python `ivy_l2s.py:741`: `return expr.clone([desugar(a) for a in expr.args])`

### 4. `logicutil/logic_utils.go:505` — ResortAst

Python `ivy_logic_utils.py:484`: `return ast.clone(args)` in `resort_ast`

### 5. `logicutil/logic_utils.go:441` — ResortSort

Python `ivy_logic_utils.py` resort_sort: always constructs new FunctionSort

### 6. `logicutil/logic_utils.go:669` — substituteByNameRec

Python `ivy_logic_utils.py:181`: `return ast.clone(substitute_ast(x,subs) for x in ast.args)`

### 7. `logicutil/logic_utils.go:1152` — reduceNamedBindersRec

Python `ivy_logic_utils.py:383`: `return ast.clone(args)` in `reduce_named_binders`

### 8. `logicutil/logic_utils.go:1204` — ReplaceNamedBindersAst

Python `ivy_logic_utils.py:404`: `return ast.clone(args)` in `replace_named_binders_ast`

### 9. `logicutil/logic_utils.go:1237` — ExpandNamedBindersAst

Python `ivy_logic_utils.py:418`: `return ast.clone(args)` in `expand_named_binders_ast`

### 10. `module/astutil.go:194` — renameASTRec

Python `ivy_logic_utils.py:207`: `return ast.clone(args)` in `rename_ast`

### 11. `module/batch16.go:50` — SubstituteAstByName

Python `ivy_logic_utils.py:181`: `return ast.clone(...)` in `substitute_ast`

### 12. `module/batch16.go:256` — resortSortBySort

Python: always constructs new FunctionSort

### 13. `module/canonize.go:197` — resortASTRec

Python `ivy_logic_utils.py:491`: `return ast.clone(args)` in `resort_ast`

### 14. `module/canonize.go:223` — ResortSort

Python: always constructs new FunctionSort

### 15. `module/ops.go:734` — substituteNodesRec

Python `ivy_logic_utils.py:193`: `return ast.clone(...)` in `substitute_constants_ast`

### 16. `module/resort.go:101` — resortSymbolSort

Python: always constructs new FunctionSort

### 17. `compiler/ivy_compile.go:1977` — t2pApplyMatchFunc

Python `ivy_proof.py:1168-1171`: `apply_match_func` always constructs new Symbol

### 18. `compiler/ivy_compile.go:2139` — t2pApplyMatchAltRec

Python `ivy_proof.py:1166`: `return fmla.clone(args)` in `apply_match_alt_rec`

### 19. `proof/match.go:428` — applyMatchRec

Python `ivy_proof.py:1101`: `return fmla.clone(args)` in `apply_match_rec`

### 20. `proof/match.go:476` — ApplyMatchFunc

Python `ivy_proof.py:1168-1171`: always constructs new Symbol

### 21. `proof/phase5_matching.go:970` — applyMatchAltRec

Python `ivy_proof.py:1166`: `return fmla.clone(args)` in `apply_match_alt_rec`

### 22-23. `actions/transrel.go:1377,1390` — renameNode (And/Or cases)

Python `ivy_logic_utils.py:196-207`: `rename_ast` always clones unconditionally

### 24. `mc/propabs.go:149` — MkPropAbs

Python `ivy_mc.py:1321`: `return expr.clone(list(map(mk_prop_abs,expr.args)))`

### 25. `mc/qelim.go:105` — QE

Python: always clones unconditionally

### 26-27. `mc/transforms.go:51,105` — ElimIte, ToTableLookup

Python `ivy_mc.py:773`: `return expr.clone([elim_ite(e,cnsts) for e in expr.args])`

### 28-29. `unitres/unitres.go:296,316` — SubstituteLit, SubstituteConstantsLit

Python: always constructs new Literal

### 30. `webui/concept_domain.go:562` — substApplySlice

Python: always clones unconditionally

## NOT included (different pattern — convergence loops)

These use `changed` for fixed-point iteration, not clone optimization:
- `isolate/deps.go:635` — fixed-point dependency resolution
- `isolate/create.go:417` — fixed-point sort resolution
- `proof/tactics.go:466` — repeat-until-stable loop
- `mc/transforms.go:121` — outer iteration check

## Fix pattern

For every instance above, the fix is the same. Remove `changed` tracking
and always clone/construct. Example:

```go
// BEFORE (wrong — Go optimization Python doesn't have):
newChildren := make([]lg.Expr, len(children))
changed := false
for i, c := range children {
    nc := transform(c)
    newChildren[i] = nc
    if nc != c {
        changed = true
    }
}
if !changed {
    return node
}
return cloneNode(node, newChildren)

// AFTER (matches Python — always clone):
newChildren := make([]lg.Expr, len(children))
for i, c := range children {
    newChildren[i] = transform(c)
}
return cloneNode(node, newChildren)
```

## Files to modify (14 files, 30 instances)

1. `/Users/jaten/ivy/goivy/check/l2s_shared.go` — 1 instance (ActionClone)
2. `/Users/jaten/ivy/goivy/temporal/temporal.go` — 1 instance (ActionClone)
3. `/Users/jaten/ivy/goivy/check/l2s.go` — 1 instance
4. `/Users/jaten/ivy/goivy/logicutil/logic_utils.go` — 6 instances
5. `/Users/jaten/ivy/goivy/module/astutil.go` — 1 instance
6. `/Users/jaten/ivy/goivy/module/batch16.go` — 2 instances
7. `/Users/jaten/ivy/goivy/module/canonize.go` — 2 instances
8. `/Users/jaten/ivy/goivy/module/ops.go` — 1 instance
9. `/Users/jaten/ivy/goivy/module/resort.go` — 1 instance
10. `/Users/jaten/ivy/goivy/compiler/ivy_compile.go` — 2 instances
11. `/Users/jaten/ivy/goivy/proof/match.go` — 2 instances
12. `/Users/jaten/ivy/goivy/proof/phase5_matching.go` — 1 instance
13. `/Users/jaten/ivy/goivy/actions/transrel.go` — 2 instances
14. `/Users/jaten/ivy/goivy/mc/propabs.go` — 1 instance
15. `/Users/jaten/ivy/goivy/mc/qelim.go` — 1 instance
16. `/Users/jaten/ivy/goivy/mc/transforms.go` — 2 instances
17. `/Users/jaten/ivy/goivy/unitres/unitres.go` — 2 instances
18. `/Users/jaten/ivy/goivy/webui/concept_domain.go` — 1 instance

## Execution order

1. Fix #1 and #2 first (Action clones — causing immediate trace divergence)
2. Fix #3-30 in file order
3. Run `go build ./...` after each file to catch compile errors
4. Run `make golden` at the end

## Verification

1. `cd ~/ivy/goivy && go build ./...` — compiles clean
2. `cd ~/ivy/goivy/check && go test -run TestBuildDefnDeps -v` — existing tests pass
3. `make golden` — line 721105 divergence resolved; check for new divergences
