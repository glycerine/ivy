# PLAN: Delete duplicate `DualClauses` and fix the formula-vs-Clauses anti-pattern

Created: 2026-04-10 05:30 -03

## Context

Two cleanup items were deferred from prior plans (PLAN248 / PLAN249 and the just-applied check.DualClauses fix). They need to be done now:

1. **Duplicate `DualClauses` implementations** in `bmc/bmc.go:234` and `z3bridge/solver_clauses.go:51` are stripped-down ports that omit Python's instantiator branch (the same bug that was just fixed in `check/check.go`). Per CLAUDE.md rule 4 (one Python file → one Go file), `dual_clauses` lives in `ivy_logic_utils.py` and its sole Go home should be `module/ops.go`.

2. **Formula-vs-Clauses anti-pattern**: Seven Go call sites take `*module.Clauses`, call `.ToFormula()` on it (lossy: drops `Defs` and `Annot`, flattens to `lg.Expr`), pass the formula to `actions.ForwardImage` / `actions.ReverseImage` (which themselves are formula-only), then wrap the result back via `module.FormulaToClauses(...)`. Python keeps everything as `Clauses` end-to-end (`ivy_transrel.py:464` `forward_image`, `:527` `reverse_image`). The Go formula round-trip loses Clauses structure and is the same kind of "formula sandwich" that motivated the previous Clauses-level fixes for `compose_state_action`.

The intended outcome is a faithful Clauses-native port matching Python, with no duplicate implementations.

## Root cause / current state

### Item 1: Duplicate DualClauses

**Canonical version** at `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/module/ops.go:470-506`:
```go
func DualClauses(clauses *Clauses, skolemizer Skolemizer, instantiator func([]lg.Expr) *Clauses) *Clauses
```
Faithful port of Python `dual_clauses` (ivy_logic_utils.py:1554-1565), including the `if instantiator != None:` branch.

**Duplicate #1** at `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/bmc/bmc.go:234-250`:
```go
func DualClauses(conj *module.Clauses) *module.Clauses {
    if conj == nil || len(conj.Fmlas) == 0 { return module.TrueClauses(nil) }
    var negFmlas []lg.Expr
    for _, f := range conj.Fmlas {
        negFmlas = append(negFmlas, &lg.Not{Body: f})
    }
    or, err := lg.NewOr(negFmlas...)
    if err != nil { return module.NewClauses(negFmlas, nil, nil) }
    return module.NewClauses([]lg.Expr{or}, nil, nil)
}
```
- 1 production caller: `bmc.CheckIsolate` at `bmc/bmc.go:105` (`dualConj := DualClauses(conj)`)
- 4 unit tests in `bmc/bmc_test.go`: `TestDualClausesNil` (198), `TestDualClausesEmpty` (205), `TestDualClausesSingle` (213), `TestDualClausesMultiple` (232)
- **Bug**: does NOT skolemize variables, does NOT call instantiator. Python BMC at `ivy_bmc.py:35-39` uses `def witness(v): return lg.Const('@'+v.name, v.sort)` and calls `dual_clauses(conj, witness)` — so Python BMC DOES skolemize.

**Duplicate #2** at `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/z3bridge/solver_clauses.go:51-53`:
```go
func DualClauses(clauses *module.Clauses) *module.Clauses {
    return module.NegateClauses(clauses)
}
```
- **0 production callers** (verified via Grep)
- 1 test in `z3bridge/solver_test.go:856`: `TestDualClauses`
- Pure dead code: just delegates to `module.NegateClauses` (which already exists as `dualClauses(clauses, nil, nil)` per `module/ops.go:453-458`).

### Item 2: Formula-vs-Clauses anti-pattern

`actions.ForwardImage` at `actions/transrel.go:1236-1239` is a formula-level wrapper around `ForwardImageMapFormula`, which itself converts formula → Clauses → formula:
```go
func ForwardImage(pre lg.Expr, axioms lg.Expr, u *Update) lg.Expr {
    _, result := ForwardImageMapFormula(pre, axioms, u)
    return result
}
```

`actions.ReverseImage` at `actions/transrel.go:1469-1490` is even worse — it uses formula-level `Conjoin`, `ExistQuant`, `RenameAST`, `filterAxiomsBySyms` instead of the Clauses-level helpers (`ConjoinClauses`, `ExistQuantClauses`, `module.RenameClauses`, `module.ClausesUsingSymbolNames`).

**Python** by contrast (verified at `ivy_transrel.py:464`, `:527`):
```python
def forward_image(pre_state,axioms,update):
    map1,res = forward_image_map(pre_state,axioms,update)
    return res
def reverse_image(post_state,axioms,update):
    updated, clauses, _precond = update
    post_ax = clauses_using_symbols(updated,axioms)
    post_clauses = conjoin(post_state,post_ax)
    post_clauses = rename_clauses(post_clauses, dict((x,new(x)) for x in updated))
    post_updated = [new(s) for s in updated]
    res = exist_quant(post_updated,conjoin(clauses,post_clauses))
    return res
```
Both signatures take and return `Clauses`, never `lg.Expr`. The internal helpers (`clauses_using_symbols`, `conjoin`, `rename_clauses`, `exist_quant`) are all `Clauses → Clauses`.

The 7 Go call sites that do the formula round-trip are:

| # | File | Line | Function | Pattern |
|---|------|------|----------|---------|
| 1 | `actions/interpolant.go` | 43 | `ForwardInterpolant` | `ForwardImage(preState.ToFormula(), axioms.ToFormula(), update)` then `FormulaToClauses(...)` |
| 2 | `actions/interpolant.go` | 59 | `ReverseInterpolantCase` | `ReverseImage(postState.ToFormula(), axioms.ToFormula(), update)` then `FormulaToClauses(...)` |
| 3 | `tactics/tactics.go` | 169 | `(*TacticsContext).ForwardImage` | `actions.ForwardImage(preFact.ToFormula(), tc.BackgroundTheory(), update)` then `FormulaToClauses(...)` (also uses formula `BackgroundTheory()` even though `BackgroundTheoryClauses()` exists at line 110) |
| 4 | `tactics/tactics.go` | 186 | `(*TacticsContext).BackwardImage` | Same with `actions.ReverseImage` |
| 5 | `interp/helpers.go` | 139 | `Reverse` | `actions.ReverseImage(clauses.ToFormula(), axioms.ToFormula(), state.Update())` then `FormulaToClauses(...)` |
| 6 | `interp/helpers.go` | 169 | `ReverseUpdateConcreteClauses` | Same |
| 7 | `interp/helpers.go` | 223 | `ReachState` | `actions.ForwardImage(pre.ToFormula(), axioms.ToFormula(), state.Update())` then `FormulaToClauses(...)` |

`ForwardImageMapFormula` at `actions/transrel.go:1216-1228` is also pure boilerplate — Python has no equivalent, it exists only to bridge formula callers to Clauses internals.

## Recommended approach

### Item 1A — Replace `bmc.CheckIsolate`'s call and delete `bmc.DualClauses`

`bmc/bmc.go:105` currently:
```go
dualConj := DualClauses(conj)
```

Change to (mirroring Python `ivy_bmc.py:35-39`):
```go
// Python: def witness(v): return lg.Const('@' + v.name, v.sort)
//         clauses = ilu.dual_clauses(conj, witness)
witness := func(v *lg.Variable) lg.Expr {
    return module.VarToSkolem("@", v)
}
dualConj := module.DualClauses(conj, witness, mod.Instantiator)
```

`mod` is `cfg.Module`, already in scope (`bmc/bmc.go:77`). `mod.Instantiator` may be nil (BMC doesn't normally enter `theory_context`); when nil, `module.DualClauses` skips the instantiator branch — exactly matching Python's `if instantiator != None:`.

Then **delete** `bmc.DualClauses` (lines 232-250) entirely. There are no callers outside bmc.

**Tests**: delete the four `TestDualClauses*` tests in `bmc/bmc_test.go` (lines 196-247). They test the deleted function. Equivalent coverage already exists in `module/port_audit_test.go:361 TestDualClausesCustomSkolemizer` and `check/check_port_test.go:371 FuzzDualClauses` (which tests `module.DualClauses(cls, nil, nil)`).

### Item 1B — Delete `z3bridge.DualClauses`

Delete `z3bridge/solver_clauses.go:49-53` (the function plus its doc comment) entirely. There are no production callers.

Delete `TestDualClauses` at `z3bridge/solver_test.go:856-863`. The test only checks that the function returns non-nil for a single-formula clause set — `module.NegateClauses` and `module.DualClauses` are already covered elsewhere.

### Item 2A — Rewrite `actions.ForwardImage` and `actions.ReverseImage` to be Clauses-native

This is the canonical port, matching Python signatures exactly. Per CLAUDE.md rule 9, do not add a `ForwardImageClauses` wrapper as a "good enough" workaround — change the actual signature.

**File: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transrel.go:1236-1239`**

Replace `ForwardImage` with:
```go
// ForwardImage computes the forward image of a pre-state through an update,
// given background axioms.
//
// Faithful port of Python forward_image (ivy_transrel.py:464-466):
//   def forward_image(pre_state,axioms,update):
//       map1,res = forward_image_map(pre_state,axioms,update)
//       return res
func ForwardImage(preState *module.Clauses, axioms *module.Clauses, u *Update) *module.Clauses {
    _, result := ForwardImageMap(preState, axioms, u)
    return result
}
```

**File: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transrel.go:1465-1490`**

Rewrite `ReverseImage` to use the existing Clauses-level helpers:
```go
// ReverseImage computes the reverse image (weakest precondition) of a
// post-state through an update, given background axioms.
//
// Faithful port of Python reverse_image (ivy_transrel.py:527-535):
//   def reverse_image(post_state,axioms,update):
//       updated, clauses, _precond = update
//       post_ax = clauses_using_symbols(updated,axioms)
//       post_clauses = conjoin(post_state,post_ax)
//       post_clauses = rename_clauses(post_clauses, dict((x,new(x)) for x in updated))
//       post_updated = [new(s) for s in updated]
//       res = exist_quant(post_updated,conjoin(clauses,post_clauses))
//       return res
func ReverseImage(postState *module.Clauses, axioms *module.Clauses, u *Update) *module.Clauses {
    updated := u.Modified

    // Python: post_ax = clauses_using_symbols(updated, axioms)
    updatedNames := constNames(updated)
    postAx := module.ClausesUsingSymbolNames(updatedNames, axioms)

    // Python: post_clauses = conjoin(post_state, post_ax)
    postClauses := ConjoinClauses(postState, postAx)

    // Python: post_clauses = rename_clauses(post_clauses, dict((x,new(x)) for x in updated))
    renaming := make(map[lg.NodeKey]*lg.Const, len(updated))
    for _, s := range updated {
        renaming[lg.Key(s)] = NewConst(s)
    }
    postClauses = module.RenameClauses(postClauses, renaming)

    // Python: post_updated = [new(s) for s in updated]
    postUpdated := make([]*lg.Const, len(updated))
    for i, s := range updated {
        postUpdated[i] = NewConst(s)
    }

    // Python: res = exist_quant(post_updated, conjoin(clauses, post_clauses))
    _, result := ExistQuantClauses(postUpdated, ConjoinClauses(u.TR, postClauses))
    return result
}
```

All helpers used here already exist:
- `module.ClausesUsingSymbolNames` (used at `transrel.go:1191` by `ForwardImageMap`)
- `ConjoinClauses` (used at `transrel.go:1194`)
- `module.RenameClauses` (used at `transrel.go:1208`)
- `ExistQuantClauses` (used at `transrel.go:1200`)
- `constNames`, `NewConst` (already used in the formula-version `ReverseImage`)

### Item 2B — Delete `ForwardImageMapFormula` (no Python equivalent)

After Item 2A, `ForwardImageMapFormula` (lines 1213-1228) is unused except by:
- The old `ForwardImage` body — which is being rewritten
- `actions/impl_test.go:407 TestForwardImageMapReturnsMap` — a test of the formula wrapper itself

Delete the function and the test. The doc comment at `transrel.go:1889` ("same logic as ForwardImageMapFormula lines 1174-1179") in `History.ForwardStep` should be updated to remove the dangling reference.

### Item 2C — Update the 7 call sites

**File: `actions/interpolant.go:42-46`** — `ForwardInterpolant`
```go
func ForwardInterpolant(preState *module.Clauses, update *Update, postState *module.Clauses, axioms *module.Clauses, interpreted map[string]bool) *InterpolantResult {
    fwdClauses := ForwardImage(preState, axioms, update)
    return Interpolant(fwdClauses, postState, axioms, interpreted)
}
```

**File: `actions/interpolant.go:58-65`** — `ReverseInterpolantCase`
```go
func ReverseInterpolantCase(postState *module.Clauses, update *Update, preState *module.Clauses, axioms *module.Clauses, interpreted map[string]bool) *InterpolantResult {
    revClauses := ReverseImage(postState, axioms, update)
    filtered := filterGroundNonSkolem(revClauses)
    return Interpolant(preState, filtered, axioms, interpreted)
}
```

**File: `tactics/tactics.go:159-171`** — `(*TacticsContext).ForwardImage`
```go
func (tc *TacticsContext) ForwardImage(preFact *module.Clauses, action actions.Action) *module.Clauses {
    if preFact == nil || action == nil {
        return preFact
    }
    axioms := tc.BackgroundTheoryClauses()
    update := actions.GetUpdateForArt(action, tc.Mod, nil)
    if update == nil {
        return preFact
    }
    return actions.ForwardImage(preFact, axioms, update)
}
```

Note: switches from `tc.BackgroundTheory()` (formula) to `tc.BackgroundTheoryClauses()` (Clauses), which already exists at `tactics.go:110`. Annotation preservation is now handled by the Clauses helpers (no manual `preFact.Annot` threading needed because `ForwardImageMap` propagates annotations through `ConjoinClauses` and friends).

**File: `tactics/tactics.go:173-188`** — `(*TacticsContext).BackwardImage`
```go
func (tc *TacticsContext) BackwardImage(postFact *module.Clauses, action actions.Action) *module.Clauses {
    if postFact == nil || action == nil {
        return postFact
    }
    axioms := tc.BackgroundTheoryClauses()
    update := actions.GetUpdateForArt(action, tc.Mod, nil)
    if update == nil {
        return postFact
    }
    return actions.ReverseImage(postFact, axioms, update)
}
```

**File: `interp/helpers.go:131-142`** — `Reverse`
```go
func Reverse(state *State, clauses *module.Clauses) (*module.Clauses, error) {
    if state.Pred() == nil || state.Update() == nil {
        return nil, fmt.Errorf("Reverse: cannot reverse state without predecessor and update")
    }
    if clauses == nil {
        clauses = state.Clauses
    }
    axioms := state.Domain.BackgroundTheory(state.InScope)
    revClauses := actions.ReverseImage(clauses, axioms, state.Update())
    return module.AndClausesTyped(revClauses, axioms), nil
}
```

`state.Domain.BackgroundTheory(...)` already returns `*module.Clauses`, so no fix needed there — only the `clauses.ToFormula()` / `axioms.ToFormula()` round-trip is removed.

**File: `interp/helpers.go:150-172`** — `ReverseUpdateConcreteClauses`
```go
func ReverseUpdateConcreteClauses(state *State, clauses *module.Clauses) (*module.Clauses, error) {
    if state.Pred() == nil || state.Update() == nil {
        return nil, fmt.Errorf("ReverseUpdateConcreteClauses: no predecessor or update")
    }
    if clauses == nil {
        clauses = state.Clauses
    }
    axioms := state.Domain.BackgroundTheory(state.InScope)
    interpreted := functionsToInterpreted(state.Domain.Functions)

    fi := actions.ForwardInterpolant(state.Pred().Clauses, state.Update(), clauses, axioms, interpreted)
    if fi != nil {
        return nil, &UnsatCoreWithInterpolant{Core: fi.Core, Itp: fi.Itp}
    }

    revClauses := actions.ReverseImage(clauses, axioms, state.Update())
    return module.AndClausesTyped(revClauses, axioms), nil
}
```

**File: `interp/helpers.go:212-241`** — `ReachState`
```go
func ReachState(state *State, clauses *module.Clauses) *State {
    if state.Pred() == nil || state.Update() == nil {
        return nil
    }
    pre := JoinUnders(state.Pred())
    if clauses == nil {
        clauses = state.Clauses
    }
    axioms := state.Domain.BackgroundTheory(state.InScope)
    imgClauses := module.AndClausesTyped(
        actions.ForwardImage(pre, axioms, state.Update()),
        axioms,
        clauses,
    )
    solver := z3bridge.NewSolver(nil, nil)
    t := solver.NewTranslator()
    defer t.Close()
    result, err := t.IsSat(imgClauses.ToFormula())
    if err != nil || result != z3bridge.Sat {
        return nil
    }
    return AddUnder(state, imgClauses, nil, nil)
}
```

The final `imgClauses.ToFormula()` is for the SAT check (Z3 needs a formula), which is the legitimate boundary where formulas are required — not the anti-pattern. Leave it.

### Item 2D — Update tests for `ForwardImage`/`ReverseImage`

**File: `actions/impl_test.go`**

- `TestForwardImageTrivial` (line 382) — currently `ForwardImage(lg.True, lg.True, u)`. Change to `ForwardImage(module.TrueClauses(nil), module.TrueClauses(nil), u)` and update the nil-check.
- `TestForwardImageWithUpdate` (line 391) — currently `ForwardImage(pre, lg.True, u)` where `pre` is a formula. Wrap as Clauses: `ForwardImage(module.FormulaToClauses(pre, nil), module.TrueClauses(nil), u)` and update the propagation assertion.
- `TestForwardImageMapReturnsMap` (line 405) — **delete** along with `ForwardImageMapFormula`.
- `TestReverseImageBasic` (line 622) — currently `ReverseImage(post, lg.True, u)`. Change to Clauses args.

**File: `actions/transrel_test.go:443-449`**
- `TestForwardImageStub` — currently `ForwardImage(lg.True, lg.True, u)`. Change to `ForwardImage(module.TrueClauses(nil), module.TrueClauses(nil), u)` and update the nil-check on the result.

## Critical files to modify

1. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/bmc/bmc.go`
   - Line 105: replace `DualClauses(conj)` call with `module.DualClauses(conj, witness, mod.Instantiator)`
   - Lines 232-250: delete `bmc.DualClauses` function entirely
2. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/bmc/bmc_test.go`
   - Lines 196-247: delete 4 `TestDualClauses*` tests
3. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/z3bridge/solver_clauses.go`
   - Lines 49-53: delete `z3bridge.DualClauses` function entirely
4. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/z3bridge/solver_test.go`
   - Lines 856-863: delete `TestDualClauses`
5. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transrel.go`
   - Lines 1236-1239: rewrite `ForwardImage` to take/return `*module.Clauses`
   - Lines 1465-1490: rewrite `ReverseImage` to take/return `*module.Clauses` using Clauses-level helpers
   - Lines 1213-1228: delete `ForwardImageMapFormula`
   - Line 1889: update or remove dangling doc reference to `ForwardImageMapFormula`
6. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/interpolant.go`
   - Lines 42-46: rewrite `ForwardInterpolant` to pass Clauses directly
   - Lines 58-65: rewrite `ReverseInterpolantCase` to pass Clauses directly
7. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/tactics/tactics.go`
   - Lines 159-171: rewrite `(*TacticsContext).ForwardImage` to pass Clauses, switch to `BackgroundTheoryClauses`
   - Lines 173-188: rewrite `(*TacticsContext).BackwardImage` similarly
8. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/interp/helpers.go`
   - Lines 131-142: rewrite `Reverse` to pass Clauses directly
   - Lines 150-172: rewrite `ReverseUpdateConcreteClauses` to pass Clauses directly
   - Lines 212-241: rewrite `ReachState` to pass Clauses directly (preserve final `ToFormula()` for SAT check)
9. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/impl_test.go`
   - Lines 382, 391: update `TestForwardImage*` to use Clauses args
   - Lines 405-414: delete `TestForwardImageMapReturnsMap`
   - Lines 622-628: update `TestReverseImageBasic` to use Clauses args
10. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transrel_test.go`
    - Lines 443-449: update `TestForwardImageStub` to use Clauses args

## Existing functions to reuse (do NOT recreate)

- `module.DualClauses(clauses, skolemizer, instantiator)` at `module/ops.go:470-506` — canonical Python port
- `module.VarToSkolem(prefix, v)` at `module/clauses.go` — produces `<prefix><name>` constants
- `module.NegateClauses(clauses)` at `module/ops.go:453-458` — wrapper around `dualClauses(clauses, nil, nil)` (already used by `z3bridge.ClausesImplyList`)
- `actions.ForwardImageMap(preState, axioms, u)` at `actions/transrel.go:1186-1211` — Clauses-level forward image
- `actions.ConjoinClauses(c1, c2)` at `actions/transrel.go` — Clauses-level conjoin (matches Python `conjoin`)
- `actions.ExistQuantClauses(syms, c)` at `actions/transrel.go` — Clauses-level existential
- `module.ClausesUsingSymbolNames(names, c)` — Clauses-level filter (matches Python `clauses_using_symbols`)
- `module.RenameClauses(c, renaming)` — Clauses-level rename (matches Python `rename_clauses`)
- `actions.constNames(syms)`, `actions.NewConst(s)` — internal helpers, already used in formula-version `ReverseImage`
- `tactics.(*TacticsContext).BackgroundTheoryClauses()` at `tactics/tactics.go:110-119` — Clauses version of `BackgroundTheory()`

## Verification

1. **Build**: `cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go build ./...` — expect clean build after all signature changes.
2. **Vet**: `go vet ./...` — expect clean.
3. **Targeted unit tests**:
   ```
   go test ./bmc/... -count=1
   go test ./z3bridge/... -count=1
   go test ./actions/... -count=1
   go test ./tactics/... -count=1
   go test ./interp/... -count=1
   go test ./check/... -count=1
   go test ./module/... -count=1
   ```
4. **Full sweep**: `go test ./... -count=1 -timeout 600s`
5. **Golden test** (the live regression):
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./parser/... -run TestOrdLive -timeout 600s
   ```
   The bmc/z3bridge cleanup should not change the trace at line 233843 (those code paths aren't on it). The formula-vs-Clauses fix may surface earlier or later divergences once Clauses-level state survives forward/backward image operations — capture any new divergence position for a follow-up plan.
6. **Inspect log.red diff** around the prior failing position (line 233843) and any new divergence locations.

## Notes / out of scope

- `actions/interpolant.go:121 UnsatCore` is a known stub ("For now, return clauses2 if the combined is non-trivially constrained") — out of scope.
- `interp/helpers.go:336 CaseConjecture` builds `negClauses := module.FormulaToClauses(&lg.Not{...}, nil)` from `clausesFmla` — this is a different anti-pattern (direct formula construction rather than `module.NegateClauses`), out of scope.
- The `TRRaw`/`PreRaw` "raw formula" fields on `Update` (`actions/transrel.go:123-124`) are a separate piece of formula/Clauses tech debt — out of scope.
- Per CLAUDE.md rule C: no new package-level vars introduced. The witness closure in `bmc.CheckIsolate` is per-invocation local state.
- Per CLAUDE.md rule 4: `dual_clauses` is in `ivy_logic_utils.py` → its Go home is `module/ops.go`. `forward_image`/`reverse_image` are in `ivy_transrel.py` → their Go home is `actions/transrel.go`. No reorganization needed.
- Per CLAUDE.md rule 9: do not add `ForwardImageClauses` / `ReverseImageClauses` as "convenience wrappers". The right fix is to change the signatures of the existing functions to match Python.
- Both `History.ForwardStep` (`actions/transrel.go:1885`) and `ComposeStateAction` (`actions/transrel.go:1348`) already use `ForwardImageMap` directly — they are unaffected by these changes.
