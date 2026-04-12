# Plan: Fix transformAction always-clone + modPass binding conformance

Created: 2026-04-12 ~07:20 UTC
Updated: 2026-04-12 ~19:00 UTC

## Context

Divergence at line 260353 (moved from 259926 by prior fixes in this session):
```
260353  go : XTRACE: CallAction.__init__ uniqueID=700 counter=701
        py : XTRACE: ast.LF.clone PRESERVE origid=226 counter=2306
```

Diagnostic traces confirmed Go and Python have identical model counts (37 invars, 113 asms, 24 bindings, 23 prems, 11 property prems). The divergence is in **how bindings are processed** inside `modPass`.

**Root cause:** Python's `mod_pass` processes bindings with `b.clone([transform(b.action)])` where `transform` is `replace_temporals_by_named_binder_g_ast` — a recursive function that **always clones** every node via `ast.clone(args)` (ivy_logic_utils.py:322). Go's `modPass` uses `transformAction(b.Action.Stmt, transform)` which only clones when `changed==true` (l2s.go:857-860). This produces fewer traces in Go, causing the trace streams to fall out of sync.

## Fix

### Bug 8: `transformAction` only clones when changed — Python always clones

**File:** `check/l2s.go:833-861`

Python's `replace_temporals_by_named_binder_g_ast` (ivy_logic_utils.py:306-322) and `normalize_named_binders` (ivy_logic_utils.py:273-277) both fall through to `ast.clone(args)` for non-special types — **always cloning**, even when args are unchanged. This applies to every node in the tree, including actions.

Go's `transformAction` tracks a `changed` flag and only calls `act.ActionClone(newArgs)` when something changed (line 857-860). When nothing changed, it returns the original action — producing no clone traces.

**Fix:** Remove the `changed` guard. Always call `act.ActionClone(newArgs)`.

```go
// Current (l2s.go:857-860):
if changed {
    return act.ActionClone(newArgs)
}
return act

// Replace with:
return act.ActionClone(newArgs)
```

The `changed` variable and its tracking can also be removed (dead code after this change).

### Cleanup: remove `changed` tracking

After removing the guard, simplify:

```go
func transformAction(act actions.Action, transform func(lg.Expr) lg.Expr) actions.Action {
    if act == nil {
        return nil
    }
    args := act.ActionArgs()
    newArgs := make([]lg.Expr, len(args))
    for i, a := range args {
        if sub, ok := a.(actions.Action); ok {
            newArgs[i] = transformAction(sub, transform)
        } else if a != nil {
            newArgs[i] = transform(a)
        } else {
            newArgs[i] = a
        }
    }
    return act.ActionClone(newArgs)
}
```

## Files to modify

1. `check/l2s.go` — `transformAction`: remove `changed` guard, always clone

## Prior fixes (ALL DONE in this session)

1. `proof/phase5_matching.go` — Bug 7a/7b: `CompileNode` → `Thing` (259926→260131)
2. `proof/goal.go` — static traces → `ThingLF`, remove `iu` import
3. `check/l2s.go` compileInvar — remove extra `inv.Clone()`, use `CompileExprVocabExtLF` + `LabelTemporalNode` (faithful port)
4. `check/l2s_auto.go` — `SubstituteConstantsExpr` → `lu.Substitute` (260131→260353)
5. `proof/phase5_matching.go` — added `CompileExprVocabExtLF`, `LabelTemporalNode`
6. `ivylogic/util.go` — (cleanup, moved `LabelTemporalNode` to proof/)
7. Python `ivy_l2s.py` — `post-compile` trace format + `HASH canon=` + modPass diagnostic traces

## Verification

1. `go build ./...` — must compile
2. `make golden` (in `~/ivy/goivy/`) — check `~/ivy/goivy/log.red` for divergence moving past 260353
3. Remove diagnostic prints (Go `fmt.Printf` and Python `print`) after verification — already done for the count diagnostics; still have the modPass xtracer traces (keep those per accrete-xtraces rule)

## Previous audit bugs (ALL DONE)

### Bug 1 (DONE): Missing LF clone in compileExprVocabLF

**File:** `proof/goal.go:602-605`
**Status:** Already applied.
After `SortInferList`, Python's `concretize_terms` clones the LabeledFormula via `t.clone(inferred_args)`. Go was mutating `compiled.Formula` directly. Fixed by replacing mutation with `compiled.Clone(cloneArgs)`.

---

### Bug 2: CompileExprVocab uses wrong TopSort mechanism

**File:** `proof/phase5_matching.go:49-52`

**Python** (`ivy_proof.py:740`): `with il.top_sort_as_default():` → adds `sig.sorts['S'] = TopS`, produces `SortAsDefault.Enter/Exit` xtraces.

**Go** (`phase5_matching.go:49-52`):
```go
savedDefault := sig.DefaultSort
sig.DefaultSort = lg.TopS
defer func() { sig.DefaultSort = savedDefault }()
```
This sets a *different* field (`DefaultSort`) instead of adding 'S' to `sig.Sorts`. No xtraces produced.

**Fix:** Replace lines 49-52 with:
```go
tsDefault := il.TopSortAsDefault(sig)
tsDefault.Enter()
defer tsDefault.Exit()
```
Delete the `savedDefault` lines.

---

### Bug 3: CompileExprVocabExt uses wrong TopSort mechanism

**File:** `proof/phase5_matching.go:108-111`

Same issue as Bug 2 — uses `sig.DefaultSort = lg.TopS` instead of `il.TopSortAsDefault(sig)`.

**Fix:** Same pattern as Bug 2.

---

### Bug 4: GoalDefns missing UninterpretedSort collection

**File:** `proof/goal.go:196-209`

**Python** (`ivy_proof.py:540-547`):
```python
for x in goal_prems(goal):
    if isinstance(x,ia.ConstantDecl) and isinstance(x.args[0],il.Symbol):
        res.add(x.args[0])
    elif isinstance(x,il.UninterpretedSort):    # ← MISSING IN GO
        res.add(x)
```

**Go:** Only collects ConstantDecl symbols, misses UninterpretedSort premises.

**Impact:** `GoalFree` uses `GoalDefns` to build the `bound` set. Missing sorts means sorts that should be bound are treated as free, potentially affecting redefinition checks and variable scoping.

**Fix:** Add after the ConstantDecl block:
```go
// Python: elif isinstance(x, il.UninterpretedSort): res.add(x)
if us, ok := p.(*lg.UninterpretedSort); ok {
    res[lg.NodeKey(us.Sexp())] = us
}
```

Note: `*lg.UninterpretedSort` must implement `lg.Expr` for this to work. Check whether it does; if not, use a separate map or store as `interface{}`.

---

### Bug 5: GoalFree missing used_sorts_ast collection

**File:** `proof/goal.go:280-295`

**Python** `fmla_vocab` (`ivy_proof.py:653-659`):
```python
things = lu.used_sorts_ast(fmla)           # sorts
things.update(lu.used_symbols_ast(fmla))   # symbols
things.update(lu.used_variables_ast(fmla)) # variables
```

**Go** `recFmla` only collects variables and symbols — no sorts:
```go
for vKey, vNode := range lu.FreeVariables(fmla).All() { ... }
for cKey, cNode := range il.UsedSymbolsAst(fmla) { ... }
// MISSING: used sorts
```

**Impact:** Free set missing sorts → redefinition check at end of `CompileDefinitionGoalVocab` won't catch a symbol redefining a free sort.

**Fix:** Add sort collection after the symbol loop:
```go
for sKey, sNode := range il.UsedSortsAst(fmla) {
    if bound[sKey] == nil {
        res[sKey] = sNode
    }
}
```
Need to verify `il.UsedSortsAst` exists; if not, port it from Python `lu.used_sorts_ast`.

---

### Bug 6: GoalVocab collects only conclusion of premise LFs, not full formula

**File:** `proof/goal.go:240-245`

**Python** (`ivy_proof.py:577`):
```python
fmlas = [x.formula for x in prems if isinstance(x,ia.LabeledFormula)] + [conc]
```
Collects `x.formula` — the FULL formula (could be SchemaBody with nested premises).

**Go:**
```go
if lf, ok := p.(*ast.LabeledFormula); ok {
    fc := ConcAsExpr(GoalConc(lf))
```
Calls `GoalConc(lf)` which for SchemaBody returns only the conclusion, missing variables in nested premises.

**Impact:** Variables appearing only in nested SchemaBody premises are missing from `GoalVocab`, potentially causing sort inference failures or wrong variable scoping.

**Fix:** For LabeledFormula premises, collect variables from ALL sub-elements, not just the conclusion. For non-SchemaBody formulas, `GoalConc` already returns the formula (no change). For SchemaBody formulas, iterate all elements:
```go
if lf, ok := p.(*ast.LabeledFormula); ok {
    if sb, ok2 := lf.Formula.(*ast.SchemaBody); ok2 {
        // Collect from all elements (prems + conc)
        for _, elem := range sb.Elems {
            if e, ok3 := elem.(lg.Expr); ok3 {
                fmlas = append(fmlas, e)
            }
        }
    } else if fc := ConcAsExpr(GoalConc(lf)); fc != nil {
        fmlas = append(fmlas, fc)
    }
}
```

---

## Files to modify

1. `proof/phase5_matching.go` — Bug 7a (line 117), Bug 7b (line 58)
2. `proof/goal.go` — Cleanup: replace static traces with ThingLF, remove `iu` import

## Verification

1. `go build ./...` — must compile
2. `make golden` (in `~/ivy/goivy/`) — check `~/ivy/goivy/log.red` for where divergence moves past 259926
