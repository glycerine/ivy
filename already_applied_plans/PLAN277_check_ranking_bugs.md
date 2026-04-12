# Ranking modPass Divergence: Missing Prems + Invars Merge

**Created:** 2026-04-12 ~late evening

## Context

Golden test `TestOrdLive` (parser/golden_test.go:450) runs `ivy_check isolate=cf_live` on `ord_live.ivy` and compares Go vs Python XTRACE output. At trace line **260353** (after 260,352 matching lines), Go and Python diverge:

```
260352  go : XTRACE: ast.LF.clone PRESERVE origid=667 counter=2306
        py : XTRACE: ast.LF.clone PRESERVE origid=667 counter=2306

260353  go : XTRACE: CallAction.__init__ uniqueID=700 counter=701
        py : XTRACE: ast.LF.clone PRESERVE origid=226 counter=2306
```

Go finishes its modPass LF cloning (29 items: origids 498-667) and moves to the next pipeline step (creating a CallAction). Python continues cloning property premise LFs (origid=226, a prem from parsing).

**Root cause:** Go's ranking modPass (`check/ranking.go:310-331`) is missing `list_transform(prems, transform)` and doesn't merge ranking invars into `model.Invars`.

## Guiding Principle

The point of tests is to find bugs. If a test uncovers a bug in production code, fix the production code — never weaken the test. Strengthen tests in that area instead.

---

## Bug 12: Ranking modPass missing property prem transformation

**Files:** `check/ranking.go:310-331`
**Python:** `ivy_ranking.py:525-534`

Python ranking modPass transforms property premises via `list_transform(prems, transform)`:
```python
def mod_pass(transform):
    model.invars = [transform(x) for x in model.invars]
    model.asms = [transform(x) for x in model.asms]
    model.bindings = [b.clone([transform(b.action)]) for b in model.bindings]
    model.init = transform(model.init)
    list_transform(prems, transform)     # ← Go is MISSING this
    for i in range(len(postconds)):
        postconds[i] = transform(postconds[i])
```

Go ranking modPass iterates `invars` (local list) but not `prems`:
```go
modPass := func(transform func(lg.Expr) lg.Expr) {
    for i, inv := range model.Invars { ... }
    for i, asm := range model.Asms { ... }
    for i, b := range model.Bindings { ... }
    if model.Init != nil { ... }
    for i, inv := range invars { ... }        // ← should be prems, not invars
    for i, pc := range postconds { ... }
}
```

**Impact:** Property premises are never transformed by any of the 3 modPass calls (replace_temporals, normalize, replace_named_binders). Named binders in property prems remain unconverted. This is the direct cause of the XTRACE divergence at line 260353.

**Fix:** Replace the `invars` loop (lines 324-326) with a prems loop matching Python's `list_transform` and l2s.go:550-558:
```go
// list_transform(prems, transform) — matching Python ivy_ranking.py:532
for i, p := range prems {
    lf, ok := p.(*ast.LabeledFormula)
    if !ok || !proof.GoalIsProperty(lf) {
        continue
    }
    if e, ok := lf.Formula.(lg.Expr); ok {
        prems[i] = lf.Clone([]ast.Node{lf.Label, transform(e)}).(*ast.LabeledFormula)
    }
}
```

**No signature change. Requires `prems` to be accessible in modPass closure (already is — defined at line 222).**

---

## Bug 13: Ranking invars not merged into model.Invars

**Files:** `check/ranking.go:258-331`
**Python:** `ivy_ranking.py:511`

Python merges ranking-generated invars into model.invars before defining modPass:
```python
model.invars = model.invars + invars    # line 511
```

Go never merges. Instead, `modPass` iterates `model.Invars` and `invars` as separate lists. This works for modPass-based steps but breaks downstream code that reads `model.Invars` directly:

- `SharedStep3_CollectNamedBinders` (l2s_shared.go:128-130) reads only `model.Invars` — misses ranking invars
- Any other SharedStep that accesses `model.Invars` directly

**Fix:** After desugaring invars (line 290), merge them into model.Invars and remove the separate `invars` loop from modPass:

```go
// After line 290 (desugar complete):
// Python ivy_ranking.py:511: model.invars = model.invars + invars
model.Invars = append(model.Invars, invars...)
```

Then remove lines 324-326 (the `for i, inv := range invars` loop) from modPass — those invars are now in model.Invars and will be processed by the model.Invars loop (lines 311-313).

---

## Bug 14: Missing WhenOperator invariant generation from pprems

**Files:** `check/ranking.go` (missing entirely)
**Python:** `ivy_ranking.py:449-461`

Python collects property prems and generates when invariants from temporal operators:
```python
pprems = list(x for x in prems if ipr.goal_is_property(x))
winvs = []
for tmprl in list(iu.unique(ilu.temporals_asts(invars+pprems))):
    if isinstance(tmprl, lg.WhenOperator):
        if tmprl.name == 'first':
            nws = lg.Or(lg.Not(l2s_waiting), lg.Not(l2s_w((), tmprl.t2)))
            tmp = lg.Implies(lg.Not(nws), lg.Eq(tmprl, lg.WhenOperator('next', tmprl.t1, tmprl.t2)))
            tmp = lg.Implies(l2s_init((), lg.Eventually(proof_label, tmprl.t2)), tmp)
            winvs.append(tmp)
for i, winv in enumerate(winvs):
    invars.append(mklf("l2s_when_"+str(i), winv))
```

Go ranking has no equivalent. No `WhenOperator`, `l2s_when`, `pprems`, or `TemporalsAst` usage in ranking.go.

**Note:** This code runs BEFORE the invars merge (Python line 511), so the generated when invariants get merged along with the other invars.

**Fix:** Add equivalent code in ranking.go between the desugar block (line 290) and the invars merge. Use existing `lu.TemporalsAst` and `l2sW`, `l2sInit` constructors from `check/l2s.go`.

**This bug may not manifest in the ord_live.ivy test case if no WhenOperator{name:"first"} nodes exist.** Defer to after Bugs 12-13 are fixed and tested. If the golden test still diverges, investigate this bug.

---

## Implementation Order

1. **Bug 13** — Merge invars into model.Invars (one line add + remove invars loop from modPass)
2. **Bug 12** — Add prems transformation to modPass (replace removed invars loop with prems loop)
3. **Bug 14** — Add WhenOperator invariant generation (defer: only if needed after 12-13 fix)

Bugs 12 and 13 are tightly coupled — the invars merge (13) removes the invars loop from modPass, and the prems loop (12) takes its place in the same location.

---

## Detailed Code Changes

### ranking.go

**Change 1: Add invars merge after desugar (after current line 290)**
```go
_ = l2sSaved // used by Desugar internally

// Python ivy_ranking.py:511: model.invars = model.invars + invars
model.Invars = append(model.Invars, invars...)
```

**Change 2: Replace invars loop with prems loop in modPass (lines 324-326)**

Replace:
```go
for i, inv := range invars {
    invars[i] = inv.Clone([]ast.Node{inv.Label, transform(inv.Formula.(lg.Expr))}).(*ast.LabeledFormula)
}
```

With:
```go
// Python ivy_ranking.py:532: list_transform(prems, transform)
for i, p := range prems {
    lf, ok := p.(*ast.LabeledFormula)
    if !ok || !proof.GoalIsProperty(lf) {
        continue
    }
    if e, ok := lf.Formula.(lg.Expr); ok {
        prems[i] = lf.Clone([]ast.Node{lf.Label, transform(e)}).(*ast.LabeledFormula)
    }
}
```

---

## Unit Tests

### File: `check/ranking_test.go` (new or extend existing)

**A. TestRankingModPass_TransformsPrems**
- Build a goal with property premises containing temporal operators
- Build a model with invars and asms
- Verify that after modPass(transform), property prems are transformed
- Verify non-property prems are untouched

**B. TestRankingModPass_InvarsMerged**
- Build ranking invars separate from model.Invars
- Verify that after merge, model.Invars contains both sets
- Verify SharedStep3_CollectNamedBinders sees all invars

**C. TestRankingModPass_PremsClonePreserve**
- Build a goal with a property prem (LF id=226)
- Call modPass with identity transform
- Verify the prem LF gets cloned (PRESERVE) — same id, different pointer

**D. Regression test: TestRankingModPass_MatchesPythonOrder**
- Build model with known invars, asms, prems
- Collect XTRACE output from modPass
- Verify clone order: invars first, then asms, then prems, then postconds

---

## Verification

1. `go build ./...` — compile clean
2. `go test ./check/... -count=1` — existing + new tests pass
3. `make test-ordlive` (or `cd parser && go test -v -count=1 -run TestOrdLive`) — golden test passes further (ideally to completion)
4. If golden test still fails after Bugs 12-13: investigate Bug 14 (WhenOperator invariants)
5. Check that LF clone PRESERVE traces match Python order at the former divergence point (260353)

---

## Files to Modify

| File | Changes |
|------|---------|
| `check/ranking.go` | Bug 12 (prems in modPass), Bug 13 (invars merge) |
| `check/ranking_test.go` | New unit tests for modPass prems + invars merge |
