# PLAN216: Change History.Post from lg.Expr to *module.Clauses

**Created:** 2026-04-07 ~16:30 UTC
**Revised:** 2026-04-07 ~17:00 UTC — add Python xtracer equivalents instead of removing Go traces

## Context

PLAN215 succeeded — lines 236017-236040 (24 defToConstraint traces) now match. The golden test diverges at line 236041:

```
236040  go : XTRACE: module/clauses.go:262 defToConstraint lhsSort=Boolean resultType=Iff
        py : XTRACE: module/clauses.go:262 defToConstraint lhsSort=Boolean resultType=Iff

236041  go : XTRACE: ops.ToOpenFormula nFmlas=0 nDefs=0
        py : XTRACE: solver.ClausesToZ3 ENTER
             fmlas=0 defs=12
```

Go emits `ops.ToOpenFormula nFmlas=0 nDefs=0` while Python emits `solver.ClausesToZ3 ENTER fmlas=0 defs=12`. These are different functions — Go has extra `ToOpenFormula` traces from formula conversions Python never makes, AND Go loses definition data (0 defs vs 12).

## Root Cause Analysis

### Two intertwined problems:

**Problem 1: Extra ToOpenFormula traces from GetHistory/NewHistory**

After the 24 defToConstraint traces from TheoryContext, Go's `CheckFcsInStateWithAG` calls `ag.GetHistory(post, nil)`. In the base case (`art/art.go:743-753`):

```go
formula = state.Clauses.ToFormula()   // calls ToOpenFormula → emits trace!
u := actions.PureState(formula)
return actions.NewHistory(cfg, u)     // TRNode() → ToOpenFormula → emits trace!
```

Python's `new_history(state)` (`ivy_interp.py:594`) stores `state.value` (Clauses) directly — no `to_formula()` call, no trace.

**Problem 2: Lost definitions from axioms conversion**

In `checkFcsNormalPath` (`check/check.go:661`):
```go
axiomExpr := module.ClausesToFormula(axioms)  // converts Clauses→formula, defs inlined/lost
res := history.SatisfyWithCond(axiomExpr, gmc, finalConds)  // axiomExpr has 0 defs
```

Python passes axioms as Clauses directly to `history.satisfy(axioms, gmc, ...)`, preserving 12 defs through to `clauses_to_z3`.

**Problem 3: Go-only xtracer traces need Python equivalents**

Go has xtracer traces between defToConstraint and ClausesToZ3 that Python lacks. Per policy, add Python equivalents rather than remove Go traces. The traces need format adjustment since Post becomes `*module.Clauses` instead of `lg.Expr`.

### Structural root cause

Go's `History.Post` is `lg.Expr` (formula). Python's `History.post` is `Clauses`. This mismatch causes extra formula conversions (emitting `ToOpenFormula` traces) and loss of definition structure (Clauses.Defs gone after round-trip through formula).

## Approach

1. Change `History.Post` from `lg.Expr` to `*module.Clauses` to match Python
2. Update all History methods (`ForwardStep`, `Assume`, `Satisfy`, `SatisfyWithCond`) and callers
3. Adjust existing Go xtracer traces for new types; add matching Python traces
4. Add `PureStateClauses` helper for creating pure states from Clauses without formula conversion

## File Changes

### Part A: Structural Changes (Go)

#### A1. `actions/transrel.go` — History struct (line 1790-1796)

```go
// Change:
Post    lg.Expr        // characteristic formula of the current state
// To:
Post    *module.Clauses // characteristic clauses of the current state (matches Python self.post)
```

#### A2. `actions/transrel.go` — Add PureStateClauses (after line 183)

```go
// PureStateClauses creates a pure-state Update from Clauses directly,
// matching Python's pure_state(clauses) = (None, clauses, false_clauses()).
func PureStateClauses(clauses *module.Clauses) *Update {
    return &Update{
        ModifiedAll: true,
        TR:          clauses,
        Pre:         module.FalseClauses(nil),
    }
}
```

#### A3. `actions/transrel.go` — NewHistory (line 1802-1812)

```go
// Before:
Post: state.TRNode(),  // calls ToOpenFormula()!

// After:
Post: state.TR,  // store Clauses directly, no ToOpenFormula
```

#### A4. `actions/transrel.go` — ForwardStep (line 1818-1840)

Change signature: `axioms lg.Expr` → `axioms *module.Clauses`

Call `ForwardImageMap(h.Post, axioms, u)` directly instead of `ForwardImageMapFormula`. Build Renaming from NodeKey→Const map (same logic as ForwardImageMapFormula:1174-1179):

```go
func (h *History) ForwardStep(axioms *module.Clauses, u *Update, action lg.Expr) *History {
    eqMap, result := ForwardImageMap(h.Post, axioms, u)
    renaming := make(Renaming, len(eqMap))
    for _, s := range u.Modified {
        if renamed, ok := eqMap[lg.Key(s)]; ok {
            renaming[s.Name] = renamed.Name
        }
    }
    // ... same Maps/Actions append logic ...
    return &History{Cfg: h.Cfg, Post: result, Maps: newMaps, Actions: newActions, Mod: h.Mod}
}
```

#### A5. `actions/transrel.go` — Assume (line 1846-1855)

Change signature: `formula lg.Expr` → `clauses *module.Clauses`

Use Clauses-level operations:
```go
func (h *History) Assume(clauses *module.Clauses) *History {
    renamed := RenameDistinctClauses(clauses, h.Post)
    newPost := module.AndClausesTyped(h.Post, renamed)
    return &History{Cfg: h.Cfg, Post: newPost, Maps: h.Maps, Actions: h.Actions, Mod: h.Mod}
}
```

#### A6. `actions/transrel.go` — Satisfy / SatisfyWithCond (lines 1873, 1881)

Change signatures: `axioms lg.Expr` → `axioms *module.Clauses`

In SatisfyWithCond, replace:
```go
postClauses := module.FormulaToClauses(h.Post, nil)
axiomClauses := module.FormulaToClauses(axioms, nil)
post := module.AndClausesTyped(postClauses, axiomClauses)
```
with:
```go
post := module.AndClausesTyped(h.Post, axioms)
```

### Part B: Caller Updates (Go)

#### B1. `art/art.go` — GetHistory base case (line 743-753)

```go
// Before:
formula = state.Clauses.ToFormula()   // emits ToOpenFormula trace!
u := actions.PureState(formula)

// After:
clauses := state.Clauses
if clauses == nil {
    clauses = module.TrueClauses(nil)
}
u := actions.PureStateClauses(clauses)
```

#### B2. `art/art.go` — GetHistory recursive case (line 766-776)

```go
// Before:
axioms = bgTheory.ToFormula()
// After:
axioms = bgTheory  // already *module.Clauses

// Change axioms variable type from lg.Expr to *module.Clauses:
var axioms *module.Clauses = module.TrueClauses(nil)
```

#### B3. `check/check.go` — checkFcsNormalPath (line 661-662)

```go
// Before:
axiomExpr := module.ClausesToFormula(axioms)
res := history.SatisfyWithCond(axiomExpr, gmc, finalConds)

// After:
res := history.SatisfyWithCond(axioms, gmc, finalConds)
```

#### B4. `check/check.go` — checkFcsTracePath (line 524-525)

```go
// Before:
postClauses := module.NewClauses([]lg.Expr{history.Post}, nil, nil)
clauses := module.AndClausesTyped(postClauses, axioms)

// After:
clauses := module.AndClausesTyped(history.Post, axioms)
```

#### B5. `interp/helpers.go` — HistoryForwardStep (line 410-411)

```go
// Before:
bg := pred.Domain.BackgroundTheory(pred.InScope).ToFormula()
// After:
bg := pred.Domain.BackgroundTheory(pred.InScope)
```

#### B6. `interp/helpers.go` — HistorySatisfy (line 420-421)

```go
// Before:
return history.Satisfy(axioms.ToFormula())
// After:
return history.Satisfy(axioms)
```

#### B7. `interp/phase4.go` — line 224

```go
// Before:
h := actions.NewHistory(cfg, actions.PureState(comp.Pre))
// After:
h := actions.NewHistory(cfg, actions.PureStateClauses(module.FormulaToClauses(comp.Pre, nil)))
```

#### B8. `interp/phase4.go` — line 228

```go
// Before:
h = h.ForwardStep(bg.ToFormula(), upd, nil)
// After:
h = h.ForwardStep(bg, upd, nil)
```

#### B9. `interp/phase4.go` — lines 232-233

```go
// Before:
h = h.Assume(comp.Post)
// After:
h = h.Assume(module.FormulaToClauses(comp.Post, nil))
```

#### B10. `interp/phase4.go` — line 237

```go
// Before:
bmcRes := h.Satisfy(bg.ToFormula())
// After:
bmcRes := h.Satisfy(bg)
```

#### B11. `trace/trace.go` — lines 601-606

```go
// Before:
clauses := module.FormulaToClauses(history.Post, actions.EmptyAnnotation{})
// After:
clauses := history.Post
if clauses.Annot == nil {
    clauses = module.NewClauses(clauses.Fmlas, clauses.Defs, actions.EmptyAnnotation{})
}
```

#### B12. `art/art.go` — BMC (line 813)

```go
// Before:
h = h.Assume(errorCond)
// After:
h = h.Assume(module.FormulaToClauses(errorCond, nil))
```

### Part C: Xtracer Trace Adjustments (Go) + Python Equivalents

All Go traces adjusted from formula-level info (PostType, PostSort) to Clauses-level info (fmlas count, defs count). Matching Python traces added.

#### C1. `check/check.go:653` — checkFcsNormalPath trace

**Go** (adjust format for Clauses):
```go
xtracer.Trace("check.checkFcsNormalPath history path\n postFmlas=%d postDefs=%d axiomFmlas=%d axiomDefs=%d checkers=%d",
    len(history.Post.Fmlas), len(history.Post.Defs), len(axioms.Fmlas), len(axioms.Defs), len(filteredCheckers))
```

**Python** `ivy_check.py` — insert in the `else` branch (line ~416) before `res = history.satisfy(...)`:
```python
        xtracer.trace("check.checkFcsNormalPath history path\n postFmlas=%d postDefs=%d axiomFmlas=%d axiomDefs=%d checkers=%d" % (len(history.post.fmlas), len(history.post.defs), len(axioms.fmlas), len(axioms.defs), len(filter_fcs(fcs))))
```

#### C2. `actions/transrel.go:1895` — SatisfyWithCond ENTER trace

**Go** (adjust format):
```go
xtracer.Trace("transrel.SatisfyWithCond ENTER\n postFmlas=%d postDefs=%d axiomFmlas=%d axiomDefs=%d",
    len(h.Post.Fmlas), len(h.Post.Defs), len(axioms.Fmlas), len(axioms.Defs))
```

**Python** `ivy_transrel.py` — insert in `satisfy` method (line ~653) after the `_get_model_clauses` null check:
```python
        xtracer.trace("transrel.SatisfyWithCond ENTER\n postFmlas=%d postDefs=%d axiomFmlas=%d axiomDefs=%d" % (len(self.post.fmlas), len(self.post.defs), len(axioms.fmlas), len(axioms.defs)))
```

#### C3. `actions/transrel.go:1897` — SatisfyWithCond postClauses trace

**Go** (adjust — `postClauses` is now just `h.Post`):
```go
xtracer.Trace("transrel.SatisfyWithCond postClauses fmlas=%d defs=%d",
    len(h.Post.Fmlas), len(h.Post.Defs))
```

**Python** `ivy_transrel.py` — insert after ENTER trace, before `and_clauses`:
```python
        xtracer.trace("transrel.SatisfyWithCond postClauses fmlas=%d defs=%d" % (len(self.post.fmlas), len(self.post.defs)))
```

#### C4. `actions/transrel.go:1905` — SatisfyWithCond combined trace

**Go** (add defs count):
```go
xtracer.Trace("transrel.SatisfyWithCond combined fmlas=%d defs=%d", len(post.Fmlas), len(post.Defs))
```

**Python** `ivy_transrel.py` — insert after `post = and_clauses(self.post, axioms)`:
```python
        xtracer.trace("transrel.SatisfyWithCond combined fmlas=%d defs=%d" % (len(post.fmlas), len(post.defs)))
```

### Part D: Test File Updates

#### D1. `actions/transrel_test.go`

- Line 497: `PureState(lg.True)` → `PureStateClauses(module.TrueClauses(nil))`
- Line 503: `h.Post.Equal(lg.True)` → `len(h.Post.Fmlas) == 0 && len(h.Post.Defs) == 0` (empty = True)
- Line 523: `PureState(lg.True)` → `PureStateClauses(module.TrueClauses(nil))`
- Line 524: `h.Assume(lg.True)` → `h.Assume(module.TrueClauses(nil))`
- Line 529: `h.Post.Equal(lg.True)` → same Clauses comparison
- Line 536: `PureState(lg.True)` → `PureStateClauses(module.TrueClauses(nil))`
- Line 538: `h.ForwardStep(lg.True, u, lg.True)` → `h.ForwardStep(module.TrueClauses(nil), u, lg.True)`

#### D2. `actions/impl_test.go`

- Lines 637,658,674,690: `PureState(...)` → `PureStateClauses(module.FormulaToClauses(..., nil))` or `PureStateClauses(module.TrueClauses(nil))`
- Lines 639,661,662: `h.ForwardStep(lg.True, ...)` → `h.ForwardStep(module.TrueClauses(nil), ...)`
- Line 651: `formulaContainsName(h2.Post, "const_a")` → `formulaContainsName(h2.Post.ToOpenFormula(), "const_a")` (fine in tests)
- Line 677: `h.Assume(assumption)` → `h.Assume(module.FormulaToClauses(assumption, nil))`
- Line 682: `h.Post.Equal(h2.Post)` → compare via `h.Post.ToOpenFormula().Equal(h2.Post.ToOpenFormula())` (fine in tests)
- Lines 692,704: `h.Satisfy(lg.True)` → `h.Satisfy(module.TrueClauses(nil))`
- Line 702: `&History{Post: nil}` → same (nil *Clauses is valid)

## Verification

```bash
cd ~/ivy/goivy && go build ./... && go test ./actions/... && make golden
```

Expected: The golden test advances past line 236041. The trace sequence after the 24 defToConstraint traces should be:
1. `check.checkFcsNormalPath history path\n postFmlas=0 postDefs=0 axiomFmlas=? axiomDefs=12 checkers=?`
2. `transrel.SatisfyWithCond ENTER\n postFmlas=0 postDefs=0 axiomFmlas=? axiomDefs=12`
3. `transrel.SatisfyWithCond postClauses fmlas=0 defs=0`
4. `transrel.SatisfyWithCond combined fmlas=? defs=12`
5. `solver.ClausesToZ3 ENTER\n fmlas=? defs=12`

All matching between Go and Python.

## Notes

- `ForwardImageMapFormula` becomes unused after ForwardStep calls `ForwardImageMap` directly. Keep for now; remove later if no other callers.
- `PureState(formula lg.Expr)` kept alongside `PureStateClauses` since `interp/phase4.go` still needs formula→Clauses conversion for `comp.Pre`.
- `conjoinFormulas` and `RenameDistinct` (formula-level) kept for other uses; History methods now use Clauses-level equivalents.
- Test Post comparisons use `.ToOpenFormula()` for formula-level equality; this is fine in unit tests where xtracer isn't capturing golden output.
