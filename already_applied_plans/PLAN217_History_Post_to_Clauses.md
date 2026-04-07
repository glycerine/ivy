# PLAN216: Change History.Post from lg.Expr to *module.Clauses

**Created:** 2026-04-07 ~16:30 UTC

## Context

PLAN215 succeeded — lines 236017-236040 (24 defToConstraint traces) now match. The golden test diverges at line 236041:

```
236040  go : XTRACE: module/clauses.go:262 defToConstraint lhsSort=Boolean resultType=Iff
        py : XTRACE: module/clauses.go:262 defToConstraint lhsSort=Boolean resultType=Iff

236041  go : XTRACE: ops.ToOpenFormula nFmlas=0 nDefs=0
        py : XTRACE: solver.ClausesToZ3 ENTER
             fmlas=0 defs=12
```

Go emits `ops.ToOpenFormula nFmlas=0 nDefs=0` while Python emits `solver.ClausesToZ3 ENTER fmlas=0 defs=12`. These are different functions — Go has extra traces Python doesn't have, AND Go loses definition data (0 defs vs 12).

## Root Cause Analysis

### Two intertwined problems:

**Problem 1: Extra ToOpenFormula traces from GetHistory/NewHistory**

After the 24 defToConstraint traces from TheoryContext, Go's next call is `CheckFcsInStateWithAG` which calls `ag.GetHistory(post, nil)`. In the base case (`art/art.go:743-753`):

```go
formula = state.Clauses.ToFormula()   // calls ToOpenFormula → emits trace!
u := actions.PureState(formula)
return actions.NewHistory(cfg, u)     // TRNode() → ToOpenFormula → emits trace!
```

Python's equivalent `new_history(state)` (`ivy_interp.py:594-595`) just stores `state.value` (a Clauses object) directly:
```python
def new_history(state):
    return History(state.value)  # no to_formula() call, no trace
```

**Problem 2: Lost definitions from axioms conversion**

In `checkFcsNormalPath` (`check/check.go:661`):
```go
axiomExpr := module.ClausesToFormula(axioms)  // converts Clauses→formula, defs inlined/lost
res := history.SatisfyWithCond(axiomExpr, gmc, finalConds)  // axiomExpr has 0 defs
```

Python passes axioms as Clauses directly:
```python
res = history.satisfy(axioms, gmc, filter_fcs(fcs))  # axioms is Clauses with 12 defs
```

Inside Go's `SatisfyWithCond`, axioms are converted back to Clauses via `FormulaToClauses(axioms)`, but definitions are already lost. The combined post+axioms Clauses has 0 defs. The solver's `ClausesToZ3` then processes 0 definitions instead of 12.

**Problem 3: Go-only xtracer traces between defToConstraint and ClausesToZ3**

Go emits multiple xtracer traces that Python doesn't have:
- `check.checkFcsNormalPath history path` (check/check.go:653)
- `transrel.SatisfyWithCond ENTER` (transrel.go:1895)
- `transrel.SatisfyWithCond postClauses fmlas=` (transrel.go:1897)
- `transrel.SatisfyWithCond combined fmlas=` (transrel.go:1905)

Python has NO traces between the defToConstraint calls and `solver.ClausesToZ3 ENTER`.

### Structural root cause

Go's `History.Post` is `lg.Expr` (a formula). Python's `History.post` is `Clauses` (preserves fmlas+defs separately). This fundamental type mismatch causes:
- Extra formula conversions that emit `ToOpenFormula` traces
- Loss of definition structure when round-tripping through formula representation
- Incorrect solver input (0 defs instead of 12)

Python's History:
```python
class History(object):
    def __init__(self, state, maps=[], actions=[]):
        update, clauses, pre = state
        self.post = clauses  # Clauses object, always
```

Go's History:
```go
type History struct {
    Post    lg.Expr  // formula, loses Clauses.Defs structure
}
```

## Approach

Change `History.Post` from `lg.Expr` to `*module.Clauses` to match Python. Update all History methods and callers. Remove Go-only xtracer traces. This eliminates the extra `ToOpenFormula` calls and preserves definition data through to the solver.

## File Changes

### 1. `actions/transrel.go` — History struct (line 1790-1796)

```go
// Before:
type History struct {
    Cfg     *iu.IvyUtilsConfig
    Post    lg.Expr
    Maps    []Renaming
    Actions []lg.Expr
    Mod     *module.Module
}

// After:
type History struct {
    Cfg     *iu.IvyUtilsConfig
    Post    *module.Clauses   // was lg.Expr; matches Python self.post (Clauses)
    Maps    []Renaming
    Actions []lg.Expr
    Mod     *module.Module
}
```

### 2. `actions/transrel.go` — NewHistory (line 1802-1812)

```go
// Before:
func NewHistory(cfg *iu.IvyUtilsConfig, state *Update) *History {
    if !IsPureState(state) {
        panic("NewHistory requires a pure state (ModifiedAll == true)")
    }
    return &History{
        Post: state.TRNode(),  // calls ToOpenFormula!
    }
}

// After:
func NewHistory(cfg *iu.IvyUtilsConfig, state *Update) *History {
    if !IsPureState(state) {
        panic("NewHistory requires a pure state (ModifiedAll == true)")
    }
    return &History{
        Cfg:  cfg,
        Post: state.TR,  // store Clauses directly, no ToOpenFormula call
    }
}
```

### 3. `actions/transrel.go` — Add PureStateClauses (after line 183)

```go
// PureStateClauses creates a pure-state Update from Clauses directly,
// avoiding the formula round-trip in PureState.
// Matches Python's pure_state(clauses) which wraps clauses as (None, clauses, false_clauses()).
func PureStateClauses(clauses *module.Clauses) *Update {
    return &Update{
        ModifiedAll: true,
        TR:          clauses,
        Pre:         module.FalseClauses(nil),
    }
}
```

### 4. `actions/transrel.go` — ForwardStep (line 1818-1840)

```go
// Before:
func (h *History) ForwardStep(axioms lg.Expr, u *Update, action lg.Expr) *History {
    eqMap, result := ForwardImageMapFormula(h.Post, axioms, u)
    renaming := make(Renaming, len(eqMap))
    for k, v := range eqMap {
        renaming[k] = v
    }
    ...
    return &History{Post: result, ...}
}

// After:
func (h *History) ForwardStep(axioms *module.Clauses, u *Update, action lg.Expr) *History {
    eqMap, result := ForwardImageMap(h.Post, axioms, u)
    // Convert NodeKey→Const map to name→name Renaming
    // (same logic as ForwardImageMapFormula lines 1174-1179)
    renaming := make(Renaming, len(eqMap))
    for _, s := range u.Modified {
        if renamed, ok := eqMap[lg.Key(s)]; ok {
            renaming[s.Name] = renamed.Name
        }
    }
    newMaps := make([]Renaming, len(h.Maps)+1)
    copy(newMaps, h.Maps)
    newMaps[len(h.Maps)] = renaming
    newActions := make([]lg.Expr, len(h.Actions)+1)
    copy(newActions, h.Actions)
    newActions[len(h.Actions)] = action
    return &History{
        Cfg:     h.Cfg,
        Post:    result,  // *module.Clauses from ForwardImageMap
        Maps:    newMaps,
        Actions: newActions,
        Mod:     h.Mod,
    }
}
```

### 5. `actions/transrel.go` — Assume (line 1846-1855)

```go
// Before:
func (h *History) Assume(formula lg.Expr) *History {
    renamed := RenameDistinct(formula, h.Post)
    newPost := conjoinFormulas(h.Post, renamed)
    return &History{Post: newPost, ...}
}

// After:
func (h *History) Assume(clauses *module.Clauses) *History {
    renamed := RenameDistinctClauses(clauses, h.Post)
    newPost := module.AndClausesTyped(h.Post, renamed)
    return &History{
        Cfg:     h.Cfg,
        Post:    newPost,
        Maps:    h.Maps,
        Actions: h.Actions,
        Mod:     h.Mod,
    }
}
```

### 6. `actions/transrel.go` — Satisfy (line 1873-1875)

```go
// Before:
func (h *History) Satisfy(axioms lg.Expr) *SatisfyResult {
    return h.SatisfyWithCond(axioms, nil, nil)
}

// After:
func (h *History) Satisfy(axioms *module.Clauses) *SatisfyResult {
    return h.SatisfyWithCond(axioms, nil, nil)
}
```

### 7. `actions/transrel.go` — SatisfyWithCond (line 1881-1910)

```go
// Before:
func (h *History) SatisfyWithCond(axioms lg.Expr, ...) *SatisfyResult {
    ...
    postClauses := module.FormulaToClauses(h.Post, nil)
    axiomClauses := module.FormulaToClauses(axioms, nil)
    post := module.AndClausesTyped(postClauses, axiomClauses)
    ...
}

// After:
func (h *History) SatisfyWithCond(axioms *module.Clauses, ...) *SatisfyResult {
    if h.Post == nil {
        return nil
    }
    ...
    post := module.AndClausesTyped(h.Post, axioms)
    ...
}
```

Also **remove** the Go-only xtracer traces at lines 1895, 1897-1902, 1905 — Python has none of these.

### 8. `art/art.go` — GetHistory base case (line 743-753)

```go
// Before:
var formula lg.Expr = lg.True
if state.Clauses != nil {
    formula = state.Clauses.ToFormula()  // calls ToOpenFormula, emits trace!
}
u := actions.PureState(formula)
return actions.NewHistory(ag.Domain.Cfg.IuCfg, u)

// After:
clauses := state.Clauses
if clauses == nil {
    clauses = module.TrueClauses(nil)
}
u := actions.PureStateClauses(clauses)
return actions.NewHistory(ag.Domain.Cfg.IuCfg, u)
```

### 9. `art/art.go` — GetHistory recursive case (line 766-776)

```go
// Before:
var axioms lg.Expr = lg.True
if state.Pred != nil && state.Pred.Domain != nil {
    bgTheory := state.Pred.Domain.BackgroundTheory(state.Pred.InScope)
    if bgTheory != nil {
        axioms = bgTheory.ToFormula()
    }
}
h = h.ForwardStep(axioms, state.Update, actionNode)

// After:
var axioms *module.Clauses = module.TrueClauses(nil)
if state.Pred != nil && state.Pred.Domain != nil {
    bgTheory := state.Pred.Domain.BackgroundTheory(state.Pred.InScope)
    if bgTheory != nil {
        axioms = bgTheory
    }
}
h = h.ForwardStep(axioms, state.Update, actionNode)
```

### 10. `check/check.go` — checkFcsNormalPath (line 640-681)

```go
// Before:
if history != nil {
    xtracer.Trace("check.checkFcsNormalPath history path\n ...")  // Go-only trace
    ...
    axiomExpr := module.ClausesToFormula(axioms)
    res := history.SatisfyWithCond(axiomExpr, gmc, finalConds)
    ...
}

// After:
if history != nil {
    // Remove xtracer.Trace — Python has no equivalent trace here
    ...
    res := history.SatisfyWithCond(axioms, gmc, finalConds)  // pass Clauses directly
    ...
}
```

### 11. `check/check.go` — checkFcsTracePath (line 515-525)

```go
// Before:
postClauses := module.NewClauses([]lg.Expr{history.Post}, nil, nil)
clauses := module.AndClausesTyped(postClauses, axioms)

// After:
clauses := module.AndClausesTyped(history.Post, axioms)  // Post is already Clauses
```

### 12. `interp/helpers.go` — HistoryForwardStep (line 410-411)

```go
// Before:
bg := pred.Domain.BackgroundTheory(pred.InScope).ToFormula()
return history.ForwardStep(bg, state.Update(), actionNode)

// After:
bg := pred.Domain.BackgroundTheory(pred.InScope)
return history.ForwardStep(bg, state.Update(), actionNode)
```

### 13. `interp/helpers.go` — HistorySatisfy (line 420-421)

```go
// Before:
axioms := state.Domain.BackgroundTheory(state.InScope)
return history.Satisfy(axioms.ToFormula())

// After:
axioms := state.Domain.BackgroundTheory(state.InScope)
return history.Satisfy(axioms)
```

### 14. `interp/phase4.go` — line 224

```go
// Before:
h := actions.NewHistory(cfg, actions.PureState(comp.Pre))

// After:
h := actions.NewHistory(cfg, actions.PureStateClauses(module.FormulaToClauses(comp.Pre, nil)))
```

### 15. `interp/phase4.go` — line 228

```go
// Before:
h = h.ForwardStep(bg.ToFormula(), upd, nil)

// After:
h = h.ForwardStep(bg, upd, nil)  // bg is already *module.Clauses
```

### 16. `interp/phase4.go` — line 232-233

```go
// Before:
if comp.Post != nil {
    h = h.Assume(comp.Post)
}

// After:
if comp.Post != nil {
    h = h.Assume(module.FormulaToClauses(comp.Post, nil))
}
```

### 17. `interp/phase4.go` — line 237

```go
// Before:
bmcRes := h.Satisfy(bg.ToFormula())

// After:
bmcRes := h.Satisfy(bg)
```

### 18. `trace/trace.go` — line 601-606

```go
// Before:
if history == nil || history.Post == nil {
    return nil
}
clauses := module.FormulaToClauses(history.Post, actions.EmptyAnnotation{})

// After:
if history == nil || history.Post == nil {
    return nil
}
clauses := history.Post  // already *module.Clauses
// Attach annotation if needed:
if clauses.Annot == nil {
    clauses = module.NewClauses(clauses.Fmlas, clauses.Defs, actions.EmptyAnnotation{})
}
```

### 19. `art/art.go` — BMC (line 813)

```go
// Before:
h = h.Assume(errorCond)

// After:
h = h.Assume(module.FormulaToClauses(errorCond, nil))
```

### 20. Test files

- `actions/transrel_test.go:503,529`: `h.Post.Equal(lg.True)` → compare Clauses to TrueClauses
- `actions/transrel_test.go:524`: `h.Assume(lg.True)` → `h.Assume(module.TrueClauses(nil))`
- `actions/transrel_test.go:538`: `h.ForwardStep(lg.True, u, lg.True)` → `h.ForwardStep(module.TrueClauses(nil), u, lg.True)`
- `actions/impl_test.go:639,661,662`: same ForwardStep change
- `actions/impl_test.go:677`: `h.Assume(assumption)` → wrap in FormulaToClauses if needed
- `actions/impl_test.go:682`: `h.Post.Equal(h2.Post)` → use Clauses comparison
- `actions/impl_test.go:692,704`: `h.Satisfy(lg.True)` → `h.Satisfy(module.TrueClauses(nil))`

## Verification

```bash
cd ~/ivy/goivy && go build ./... && make golden
```

Expected: The golden test advances past line 236041. After the 24 defToConstraint traces from TheoryContext, the next trace should be `solver.ClausesToZ3 ENTER fmlas=0 defs=12` (matching Python), and subsequent solver traces should align.

## Notes

- `ForwardImageMapFormula` becomes unused after this change (ForwardStep calls `ForwardImageMap` directly). It can be kept for backward compatibility or removed.
- `PureState(formula lg.Expr)` can be kept alongside `PureStateClauses` since some callers may still use it.
- `conjoinFormulas` is no longer used by History methods but may be used elsewhere — don't remove.
- The Renaming conversion in ForwardStep (NodeKey→Const map to string→string map) uses the same logic as `ForwardImageMapFormula` lines 1174-1179.
