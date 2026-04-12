# Plan: Fix CompileExprVocab/CompileExprVocabExt calling CompileNode instead of Thing

Created: 2026-04-12 ~07:20 UTC
Updated: 2026-04-12 ~16:30 UTC

## Context

Divergence at line 259926 (moved from 259072 by prior fixes):
```
259926  go : XTRACE: compiler.CompileNode ENTER type=LabeledFormula
        py : XTRACE: compiler.Thing ENTER type=LabeledFormula
```

Call chain: `l2s.go:423` → `CompileWithGoalVocab` (goal.go:420) → `CompileExprVocabExt` (phase5_matching.go:117) → `c.CompileNode(expr)`.

Python's `compile_expr_vocab_ext` calls `expr.compile()` which is assigned to `thing()` (ivy_compiler.py:88-94). `thing()` emits BOTH `Thing ENTER` and `CompileNode ENTER` before calling `self.cmpl()`. Go's `CompileNode` only emits `CompileNode ENTER`, missing the `Thing ENTER` trace.

## Fix

### Bug 7a: CompileExprVocabExt calls CompileNode directly

**File:** `proof/phase5_matching.go:117`

Change:
```go
compiled, err := c.CompileNode(expr)
```
To:
```go
compiled, err := c.Thing(expr)
```

### Bug 7b: CompileExprVocab has same issue

**File:** `proof/phase5_matching.go:58`

Change:
```go
compiled, err := c.CompileNode(expr)
```
To:
```go
compiled, err := c.Thing(expr)
```

### Cleanup: compileExprVocabLF static traces → ThingLF

**File:** `proof/goal.go:610-619`

Replace fragile static traces + direct CompileLF call with `ThingLF`:
```go
// Current (lines 610-619):
c := compiler.New(sig, mod)
xtracer.Trace("compiler.Thing ENTER type=LabeledFormula")
xtracer.Trace("compiler.CompileNode ENTER type=LabeledFormula")
xtracer.Trace("compiler.CompileNode return case=LabeledFormula")
xtracer.Trace("compiler.CompileLabeledFormula ENTER")
compiled, err := c.CompileLF(lf)
if err != nil {
    return nil, err
}
xtracer.Trace("compiler.Thing return type=LabeledFormula")

// Replace with:
c := compiler.New(sig, mod)
compiled, err := c.ThingLF(lf)
if err != nil {
    return nil, err
}
```

Also remove the `FORMULA_NOT_EXPR` diagnostic trace and the `iu` import (no longer needed after removing the diagnostic trace).

## Prior fixes (ALL DONE)

1. TopSortAsDefault in `proof/goal.go:574-577` — moved divergence 258992→258998.
2. Merkle state sharing (Module.SigMerkle) — moved divergence 258998→259072.
3. Clone LF after sort inference in `compileExprVocabLF` — moved divergence 259072→259926.
4. Bugs 2/3 (TopSortAsDefault in phase5) — applied.
5. Bugs 4/5/6 (GoalDefns, GoalFree, GoalVocab) — applied.

## Comprehensive audit: proof/ vs Python ivy_proof.py

Systematic comparison of every function in `proof/goal.go` and `proof/phase5_matching.go` against Python `ivy_proof.py`. Six bugs found.

---

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
