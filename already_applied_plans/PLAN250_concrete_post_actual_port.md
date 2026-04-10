# PLAN: Fix ToOpenFormula divergence at log.red:233839 — port concrete_post faithfully

Created: 2026-04-10 04:09 -03

## Context

The `TestOrdLive` golden test fails at trace line 233839 with:

```
233838  go : XTRACE: actions.EnvAction.int_update EXIT
        py : XTRACE: actions.EnvAction.int_update EXIT

233839  go : XTRACE: ops.ToOpenFormula nFmlas=118 nDefs=0
        py : XTRACE: ops.ToOpenFormula nFmlas=1 nDefs=0
```

This is during `check.guarantee_loop` after `ag.execute(envAction, prestate=pre)` returns from `EnvAction.int_update`. Both Python and Go emit an `ops.ToOpenFormula` trace next, but on **different** Clauses objects:

- **Python's nFmlas=1**: comes from `check_safety_in_state` → `Checker(lg.Or())` → `dual_clauses(formula_to_clauses(lg.Or()))` → `clauses_to_formula(...)` → `cs.to_formula()` → `to_open_formula()`. The single formula is the dual of the empty `Or()`. This happens *after* `concrete_post` returns.

- **Go's nFmlas=118**: comes from `interp.ConcretePost` calling `state.Clauses.ToFormula()` on the pre-state Clauses (which contains 118 conjectures from `GetConjs(mod)`). This happens *inside* `ConcretePost`, *before* the equivalent of Python's `Checker` step.

The root cause: Go's `ConcretePost` is **not** a faithful port of Python's `concrete_post`. Python operates on `Clauses` objects throughout (`compose_state_action` → `forward_image_map` → `conjoin`/`exist_quant_map`/`rename_clauses`, all `Clauses`-level). Go converts `state.Clauses` and `axioms` to formulas via `ToFormula()` first, which calls `ToOpenFormula()` and emits the trace. This conversion has no Python counterpart in this code path.

A second related bug exists in Go's `ComposeStateAction` itself (which is **defined but never called**): line 1332 also calls `sc.ToOpenFormula()`. Both bugs must be fixed; the function must be plumbed in.

Fixing this restores Clauses-level execution, eliminates the spurious `ToOpenFormula` trace at this point, and matches Python's `concrete_post` faithfully.

## Python source of truth

1. **`ivy_interp.py:197-209`** — `concrete_post`:
   ```python
   def concrete_post(update, state, expr=None):
       axioms = state.domain.background_theory(state.in_scope)
       cons = compose_state_action(state.value, axioms, update, check=context.check)
       res = new_state(cons, domain=state.domain, expr=expr)
       res.pred = state
       res.update = update
       return res
   ```
   Note: `axioms` is a `Clauses`. `state.value` is the tuple `(moded, clauses, precond)`. No `to_open_formula` / `to_formula` calls.

2. **`ivy_transrel.py:500-524`** — `compose_state_action`:
   ```python
   def compose_state_action(state, axioms, action, check=True):
       su, sc, sp = state
       au, ac, ap = action
       sc, sp = clausify(sc), clausify(sp)
       if check:
           pre_test = and_clauses(and_clauses(sc, ap), axioms)
           model = small_model_clauses(pre_test)
           if model != None:
               trans = extract_pre_post_model(pre_test, model, au)
               post_updated = [new(s) for s in au]
               pre_test = exist_quant(post_updated, pre_test)
               raise ActionFailed(pre_test, trans)
       if su != None:                          # None means "all moded"
           ssu = set(su)
           rn = dict((x, old(x)) for x in au if x not in ssu)
           sc = rename_clauses(sc, rn)
           ac = rename_clauses(ac, rn)         # ← DEAD CODE: rebinds local only
           su = list(su)
           union_to_list(su, au)
       img = forward_image(sc, axioms, action) # uses ORIGINAL action tuple
       return (su, img, sp)
   ```
   Critical observations:
   - `axioms` is a `Clauses` throughout — never converted to formula.
   - `sc = rename_clauses(sc, rn)` IS used; `ac = rename_clauses(ac, rn)` is **dead code** (it rebinds the local `ac` but `forward_image` is called with the original `action` tuple, which still holds the un-renamed `ac`).
   - For `su=None` (state-style with "all moded"), the rename block is skipped entirely.

3. **`ivy_transrel.py:464-466`** — `forward_image` and `forward_image_map`:
   ```python
   def forward_image(pre_state, axioms, update):
       map1, res = forward_image_map(pre_state, axioms, update)
       return res

   def forward_image_map(pre_state, axioms, update):
       updated, clauses, _precond = update
       pre_ax = clauses_using_symbols(updated, axioms)
       pre = conjoin(pre_state, pre_ax)
       map1, res = exist_quant_map(updated, conjoin(pre, clauses, annot_op=my_annot_op))
       res = rename_clauses(res, dict((new(x), x) for x in updated))
       return map1, res
   ```
   `pre_state` and `axioms` are `Clauses`. The function operates on `Clauses` throughout — no `to_open_formula` calls.

## What's wrong in Go

### Bug A: `interp/eval.go` `ConcretePost` — duplicates `compose_state_action` at the formula level

**File: `/Users/jaten/ivy/goivy/interp/eval.go:26-72`**

```go
func ConcretePost(checkPrecond bool, update *actions.Update, state *State, expr ast.Node) (*State, error) {
    if state.Domain == nil {
        return nil, fmt.Errorf("ConcretePost: state has nil domain")
    }
    axioms := state.Domain.BackgroundTheory(state.InScope)         // *Clauses (good)

    stateTR := state.Clauses.ToFormula()                            // ← BUG: emits ToOpenFormula trace
    axiomsFmla := axioms.ToFormula()                                // ← BUG: emits ToOpenFormula trace

    preNode := update.PreNode()                                     // ← also emits ToOpenFormula via Pre.ToOpenFormula()
    if checkPrecond && preNode != nil && !isNodeFalse(preNode) {
        preCombined := &lg.And{Terms: []lg.Expr{stateTR, axiomsFmla, preNode}}
        // ...solver/IsSat...
    }

    postFmla := actions.ForwardImage(stateTR, axiomsFmla, update)   // formula-level path
    postClauses := module.FormulaToClauses(postFmla, state.Clauses.Annot)
    // ...build new state from postClauses...
}
```

This re-implements `compose_state_action` inline at the formula level, bypassing the Clauses-level `actions.ComposeStateAction` (which exists in the same file as the formula-level helpers but is **never called** anywhere — verified by `grep -rn ComposeStateAction /Users/jaten/ivy/goivy`). The two `ToFormula()` calls trigger `Clauses.ToOpenFormula()`, producing the divergent trace.

### Bug B: `actions/transrel.go` `ComposeStateAction` — also calls `ToOpenFormula`

**File: `/Users/jaten/ivy/goivy/actions/transrel.go:1271-1339`**

Even though `ComposeStateAction` is currently dead code, it has the **same bug** at line 1332:

```go
img := ForwardImage(sc.ToOpenFormula(), axioms, action)            // ← BUG: ToOpenFormula trace
```

Plus secondary issues:
- Line 1273: `axioms lg.Expr` — should be `*module.Clauses` to match Python
- Line 1287: `module.FormulaToClauses(axioms, nil)` — unnecessary if `axioms` is already a Clauses
- Line 1336: `module.FormulaToClauses(img, nil)` — unnecessary if `ForwardImage` returns Clauses
- Lines 1323-1327: replaces `action` with a new Update wrapping the **renamed** `action.TR`. This is **wrong**: Python's `ac = rename_clauses(ac, rn)` is dead code that rebinds a local variable but is never read by the subsequent `forward_image(sc, axioms, action)` call (which uses the original `action` tuple). To be a faithful port, Go must rename `sc` only, not replace `action`.

Note: this `if !suAll` block is **skipped** in our specific failing path because the pre-state has `su = nil` (state-style "all moded"). But to be a faithful port we must still fix the dead-code semantics — *and* ensure `art.State.StateValue()` correctly produces an Update where `Modified=nil` is interpreted as "all moded" by `ComposeStateAction`. Verified at `/Users/jaten/ivy/goivy/art/art.go:62-71` — `StateValue()` returns `Modified: nil` already; we need `ComposeStateAction` to treat this case as "skip rename block".

### Confirmation that no Python `to_open_formula` fires between `int_update EXIT` and 233839

I traced Python's call chain after `EnvAction.int_update EXIT`:

1. `bind_olds_action` (`ivy_transrel.py:240`) — no `to_open_formula`
2. `hide_formals` for `EnvAction` — no-op (`formal_params == [] == formal_returns`)
3. `concrete_post` → `compose_state_action` → `forward_image` → `forward_image_map` — all `Clauses`-level, no `to_open_formula`
4. `ag.execute` returns
5. `fail = itp.State(expr=itp.fail_expr(post.expr))` — no `to_open_formula`
6. `check_safety_in_state(mod, ag, fail, report_pass=False)` (`ivy_check.py:446-447`) calls `check_fcs_in_state` with `[Checker(lg.Or(), report_pass=False)]`
7. `Checker.__init__` (`ivy_check.py:218-225`):
   ```python
   self.fc = lut.formula_to_clauses(conj)             # conj = lg.Or() (= False)
   if invert:
       def witness(v): return lg.Symbol('@'+v.name, v.sort)
       self.fc = lut.dual_clauses(self.fc, witness)   # ← THIS triggers to_open_formula
   ```
8. `dual_clauses` (`ivy_logic_utils.py:1554-1565`):
   ```python
   def dual_clauses(clauses, skolemizer=None):
       ...
       fmla = negate(clauses_to_formula(clauses))     # ← clauses_to_formula → cs.to_formula() → to_open_formula
       ...
   ```
9. `clauses_to_formula` (`ivy_logic_utils.py:1011-1015`) calls `cs.to_formula()` which calls `to_open_formula()` and emits the trace.

The Clauses being processed at step 9 has 1 fmla and 0 defs (the dual of empty `Or()` after substitute_clauses → just one negated formula). That matches Python's `nFmlas=1 nDefs=0`.

So **Python's ToOpenFormula trace at 233839 is from `Checker(lg.Or())` creation in `check_safety_in_state`, *not* from `concrete_post`**. Go's ConcretePost emits a **spurious** ToOpenFormula trace ahead of where Python's first call sits.

## Recommended approach

Make Go's `ConcretePost` a literal port of Python's `concrete_post`: delegate to `ComposeStateAction`, which is itself fixed to operate on `Clauses` end-to-end.

### Step 1 — Fix `ComposeStateAction` to be Clauses-level (faithful to Python)

**File: `/Users/jaten/ivy/goivy/actions/transrel.go:1271-1339`**

Change the signature and body:

```go
// ComposeStateAction composes a state and an action, returning a new state.
// Faithful port of Python compose_state_action (ivy_transrel.py:500-524).
func ComposeStateAction(
    cfg *iu.IvyUtilsConfig,
    state *Update, axioms *module.Clauses, action *Update, check bool,
) (*Update, error) {
    su := state.Modified
    suAll := state.ModifiedAll || su == nil   // Python: su == None means "all moded"
    sc := state.TR
    sp := state.Pre
    au := action.Modified

    // Python: sc, sp = clausify(sc), clausify(sp)
    // (no-op in Go since sc, sp are already *Clauses)

    // Python: if check: pre_test = and_clauses(and_clauses(sc,ap), axioms); ...
    if check && action.Pre != nil && !action.Pre.IsFalse() {
        preTest := module.AndClausesTyped(sc, action.Pre, axioms)
        slv := z3bridge.NewSolver(nil, nil)
        model, _ := slv.GetModelClauses(preTest)
        if model != nil {
            preCls, postCls := ExtractPrePostModel(cfg, preTest, model, au)
            postUpdated := make([]*lg.Const, len(au))
            for i, s := range au {
                postUpdated[i] = NewConst(s)
            }
            _, quantPreTest := ExistQuantClauses(postUpdated, preTest)
            return nil, &ActionFailed{
                PreTest:   quantPreTest.ToOpenFormula(), // only fires on failure path
                TransPre:  preCls,
                TransPost: postCls,
            }
        }
    }

    // Python: if su != None: ...rename sc; (ac rebound but dead); su = list(su); union_to_list(su, au)
    if !suAll {
        ssu := constNames(su)
        rn := make(map[lg.NodeKey]*lg.Const)
        for _, x := range au {
            if !ssu[x.Name] {
                rn[lg.Key(x)] = OldConst(x)
            }
        }
        if len(rn) > 0 {
            sc = module.RenameClauses(sc, rn)
            // NOTE: Python rebinds `ac = rename_clauses(ac, rn)` here but this is
            // dead code — `forward_image` is called with the original `action`
            // tuple, not the renamed `ac`. Do NOT replace action.TR here.
        }
        su = UpdatedJoinConst(su, au)
    }

    // Python: img = forward_image(sc, axioms, action)
    _, img := ForwardImageMap(sc, axioms, action)

    return &Update{
        Modified:    su,
        ModifiedAll: suAll,
        TR:          img,   // already a *Clauses
        Pre:         sp,
    }, nil
}
```

Key changes:
- `axioms` parameter type: `lg.Expr` → `*module.Clauses`
- `suAll := state.ModifiedAll || su == nil` so a state with `Modified=nil` (state-style) skips the rename block (matches Python's `if su != None`)
- Use `ForwardImageMap(sc, axioms, action)` (already exists at `transrel.go:1186`, takes Clauses) instead of `ForwardImage(sc.ToOpenFormula(), axioms, action)`
- Remove `module.FormulaToClauses(...)` conversions
- Do NOT replace `action.TR` after renaming (matches Python's dead-code semantics)
- The `quantPreTest.ToOpenFormula()` inside the `ActionFailed` branch only fires when the precondition is violated and `check=true`. For our failing test path `check=false` so this is unreached. Leave it as-is.

### Step 2 — Fix `ConcretePost` to delegate to `ComposeStateAction`

**File: `/Users/jaten/ivy/goivy/interp/eval.go:26-72`**

Replace the body with a faithful port of Python `concrete_post`:

```go
func ConcretePost(checkPrecond bool, update *actions.Update, state *State, expr ast.Node) (*State, error) {
    if state.Domain == nil {
        return nil, fmt.Errorf("ConcretePost: state has nil domain")
    }

    // Python: axioms = state.domain.background_theory(state.in_scope)
    axioms := state.Domain.BackgroundTheory(state.InScope)

    // Python: cons = compose_state_action(state.value, axioms, update, check=context.check)
    stateUpdate := stateValueToUpdate(state.Value())
    cons, err := actions.ComposeStateAction(
        state.Domain.Cfg.IuCfg,  // (or whatever IvyUtilsConfig the package provides)
        stateUpdate,
        axioms,
        update,
        checkPrecond,
    )
    if err != nil {
        return nil, err
    }

    // Python: res = new_state(cons, domain=state.domain, expr=expr)
    //         res.pred = state
    //         res.update = update
    postValue := NewStateValue(
        actions.ModifiedNames(cons),
        cons.TR,
        cons.Pre,
    )
    res := NewState(state.Domain, postValue, expr, "")
    res.SetPred(state)
    res.SetUpdate(update)
    return res, nil
}
```

Notes:
- `state.Value()` returns a `*StateValue` (`{Moded, Clauses, Precond}`) — convert it via the existing `stateValueToUpdate` helper at `interp/eval.go:322`. That helper currently sets `Modified: nil` when `sv.Moded` is empty; together with `suAll := state.ModifiedAll || su == nil` in Step 1 this gives the correct "all moded" semantics.
- Remove the manual `state.Clauses.ToFormula()` / `axioms.ToFormula()` / `actions.ForwardImage(...)` chain entirely. No new ToOpenFormula traces.
- The `ActionFailed` returned by `ComposeStateAction` propagates as-is. The existing `ApplyAction` caller at `interp/eval.go:174-185` already wraps it in `IvyActionFailedError` — that path remains intact.
- Look up the correct `IvyUtilsConfig` accessor on `Module` while editing (likely `state.Domain.Cfg.IuCfg`; verify against `module/config.go`).

### Step 3 — Verify we did not break other call sites

Other files using formula-level `ForwardImage`:
- `actions/interpolant.go:43` (`Interpolant`)
- `tactics/tactics.go:159-170` (`TacticsContext.ForwardImage`)
- `interp/helpers.go:212-241` (`ReachState`)

These have **the same anti-pattern** but are **not** in the failing `TestOrdLive` path at 233839, so leave them alone for this fix. (They should be flagged for a follow-up plan; do not silently regress them either.)

The signature change of `ComposeStateAction` (`axioms lg.Expr` → `*module.Clauses`) only affects callers, and the only caller will be the new `ConcretePost`. Verify with `grep -rn ComposeStateAction /Users/jaten/ivy/goivy` after editing — the result should show only the definition site and the new call from `ConcretePost`.

The signature of `ForwardImage` (formula-level) is unchanged; existing callers in `actions/interpolant.go`, `tactics/tactics.go`, `interp/helpers.go`, and the test files keep working.

## Critical files to modify

1. `/Users/jaten/ivy/goivy/actions/transrel.go` — fix `ComposeStateAction` (Step 1)
2. `/Users/jaten/ivy/goivy/interp/eval.go` — rewrite `ConcretePost` body (Step 2)

## Existing functions to reuse (do NOT recreate)

- `actions.ForwardImageMap(preState, axioms *module.Clauses, u *Update) (eqMap, *module.Clauses)` — at `actions/transrel.go:1186` — Clauses-level forward image, faithful port of Python `forward_image_map`
- `actions.ConjoinClauses` / `module.AndClausesTyped` — Clauses-level conjoin
- `actions.ExistQuantClauses` — Clauses-level existential quantification
- `actions.OldConst`, `actions.NewConst`, `actions.constNames`, `actions.UpdatedJoinConst` — helpers used in `compose_state_action`
- `interp.stateValueToUpdate` — at `interp/eval.go:322` — converts `*StateValue` to `*actions.Update`
- `interp.NewStateValue`, `interp.NewState`, `state.SetPred`, `state.SetUpdate` — for building the post-state
- `module.Clauses.IsFalse` / `actions.ModifiedNames` — already used by current code

## Verification

1. **Build**: `cd /Users/jaten/ivy/goivy && go build ./...` — expect clean build.
2. **Targeted unit tests**:
   - `go test ./interp/... -run TestConcretePost` — verify the new ConcretePost still passes existing tests at `interp/interp_test.go:465-490`.
   - `go test ./actions/... -run TestComposeStateAction` if any exists; otherwise the change is exercised transitively by the integration test below.
3. **Failing golden test** — the primary verification:
   ```
   cd /Users/jaten/ivy/goivy && go test ./parser/... -run TestOrdLive
   ```
   Expect the divergence at line 233839 to be gone. The next divergence (if any) will surface; capture it and continue iterating in a follow-up plan.
4. **Inspect the new log**: after rerunning, look at `/Users/jaten/ivy/goivy/log.red` around line 233838-233840. The Go side should now match Python: at the position where Go used to emit `ops.ToOpenFormula nFmlas=118 nDefs=0`, it should emit whatever Python emits (likely a `ops.Conjoin`/`ops.AndClauses`/`transrel.ForwardImage`-style trace from `forward_image_map`, or eventually the `ops.ToOpenFormula nFmlas=1 nDefs=0` from the later `Checker(lg.Or())` step).
5. **Cross-check no regressions**: run the broader test suite if time permits:
   ```
   cd /Users/jaten/ivy/goivy && go test ./...
   ```

## Notes / out of scope

- The same formula-vs-Clauses anti-pattern exists in `actions/interpolant.go`, `tactics/tactics.go`, and `interp/helpers.go`. Do **not** fix these here; they are not in the failing path. Track them for a follow-up cleanup plan once the trace divergence chain at log.red is fully resolved.
- The `quantPreTest.ToOpenFormula()` call inside `ActionFailed` (transrel.go:1305) only runs when `check=true` and a precondition is violated. Our failing path has `check=false`, so this is unreached here. Leave it; it can be revisited if a future divergence implicates it.
- Per CLAUDE.md rule C: do not introduce any new package-level vars; all changes should be on existing types/methods.
