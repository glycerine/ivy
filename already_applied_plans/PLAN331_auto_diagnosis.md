# Fix Plan: D3 — Full Port of `l2s_progress_made` Diagnostic
*Created: 2026-04-17 (Thursday 16:00)*

## Context

Python `auto_hook` (`ivy_l2s.py:1617-1673`, called only for `l2s_auto5`) builds three diagnostic
maps and runs two loops when a progress proof fails. Go's `diagnoseAutoFailure`
(`check/l2s_hooks.go:416-448`) only prints the final `work_needed` post-state evaluation.
The three maps (`helpful_map`, `happened_maps`, `justice_map`) and the two diagnostic loops
are entirely absent.

---

## Infrastructure Facts (from reading the code)

**Python `tr`** (passed to `auto_hook`) is `ivy_trace.Trace`, which extends `art.AnalysisGraph`.
It has `states` — a list of State objects built by replaying the action trace via
`act.match_annotation`. Each state has `.clauses.fmlas` with ground equalities at that step.
`tr.states[0]` = pre-action state (bare symbol names).
`tr.states[1]` = post-action state (also bare symbol names — the trace builder renames
`new_X` back to `X` when constructing each state).

**Go `handler.Eqs`** is built in `NewMatchHandler` (`helpers.go:170`) by calling
`slv.ClausesModelToClausesWithModel(mclauses, model, nil, true)`. The `mclauses` passed in
(`check.go:534-538`) is `history.Post` + axioms + failed checker conditions. For an L2S
2-state transition check, `history.Post` contains BOTH pre-state atoms (bare names) and
post-state atoms (`new_` prefix). So `handler.Eqs` DOES contain equalities for:
- Pre-state `foo(a)` → keyed by `lg.Key(fooConst)` where `fooConst.Name == "foo"`
- Post-state `new_foo(a)` → keyed by `lg.Key(newFooConst)` where `newFooConst.Name == "new_foo"`

The `new_` prefix comes from `actions.New(name)` (`actions/transrel.go:42-44`).

**Key insight:** Python's `tr.states[0]` = equalities under bare name in `handler.Eqs`.
Python's `tr.states[1]` = equalities under `"new_" + name` in `handler.Eqs`.

`applyRenamingToHandler` only renames `handler.Lines` (string display), not `handler.Eqs`.
So the nonce symbols are still accessible by their original names in `handler.Eqs`.

---

## Formula Structure (l2s_auto5 only, which is the only caller)

`diagnoseAutoFailure` is called from `applyAutoDiagnosticsToHandler` which is ONLY wired
for `l2s_auto5` (`l2s.go` switch case). So `lf.Formula` is always the `l2s_auto5` invariant:

```
Implies(
  And(
    l2s_saved,                                           # Conjuncts[0]
    eventually_start(),                                  # Conjuncts[1]
    exists(progress_args, nad),                          # Conjuncts[2]
    not_all_was_done(work_needed),                       # Conjuncts[3]
    Forall(progress_args,                                # Conjuncts[4]  ← all_helpful_happened
           Implies(nad,
                   Not(waiting_for_progress(*args))))),
  exists(done_args, And(Not(was_done), is_done))
)
```

After `SharedStep11_ReplaceNamedBinders`:
- `nad` = `Apply(l2s_s_i, *progress_args)` where `l2s_s_i.Name` = `was_helpful_pred_nonce`
- `waiting_for_progress(*args)` = `Apply(l2s_w_j, *progress_args)` where `l2s_w_j.Name` = `trigger_happened_pred_nonce`

Python navigation (`ivy_l2s.py:1621-1636`):
```python
lhs = invar.args[0]                      # And(Conjuncts[0..4])
all_helpful_happened = lhs.args[4]       # Forall(progress_args, Implies(nad, Not(waiting)))
# is_forall = True for l2s_auto5
was_helpful_pred_nonce = all_helpful_happened.body.args[0].rep
# body = Implies(nad, Not(waiting))
# body.args[0] = nad = Apply(l2s_s_i, ...)  → .rep = l2s_s_i.Name
trigger_happened_pred_nonce = all_helpful_happened.body.args[1].args[0].rep
# body.args[1] = Not(Apply(l2s_w_j, ...))
# body.args[1].args[0] = Apply(l2s_w_j, ...)  → .rep = l2s_w_j.Name
```

Go translation:
```go
impl, ok := lf.Formula.(*lg.Implies)       // Implies
ant, ok := impl.T1.(*lg.And)               // lhs = And(...)
allHH := ant.Conjuncts[4]                  // all_helpful_happened = Forall
fa, ok := allHH.(*lg.Forall)               // is_forall = true
bodyImpl, ok := fa.Body.(*lg.Implies)      // body = Implies(nad, Not(waiting))
app0, ok := bodyImpl.T1.(*lg.Apply)        // body.args[0] = nad = Apply(l2s_s_i, ...)
wasHelpfulNonce = app0.Func.Name           // l2s_s_i.Name
notExpr, ok := bodyImpl.T2.(*lg.Not)       // body.args[1] = Not(...)
app1, ok := notExpr.Body.(*lg.Apply)       // body.args[1].args[0] = Apply(l2s_w_j, ...)
triggerNonce = app1.Func.Name              // l2s_w_j.Name
```

---

## Scanning `handler.Eqs` by Name

To avoid constructing `lg.NewConst(name, sort)` (requires knowing the sort), iterate
`handler.Eqs` and match by the function symbol name embedded in each equality's LHS:

```go
// scanEqsByName returns a map from stringified arg-tuple to bool truth value
// for all equalities in handler.Eqs whose LHS function name == nonceName.
// Python: {tuple(eqn.args[0].args): eqn.args[1]
//          for eqn in tr.states[idx].clauses.fmlas if eqn.args[0].rep == nonceName}
func scanEqsByName(handler *MatchHandler, nonceName string) map[[8]string]bool {
    result := make(map[[8]string]bool)
    for _, eqs := range handler.Eqs {
        for _, eqExpr := range eqs {
            eq, ok := eqExpr.(*lg.Eq)
            if !ok {
                continue
            }
            app, ok := eq.T1.(*lg.Apply)
            if !ok {
                continue
            }
            if app.Func.Name != nonceName {
                continue
            }
            result[termsKey(app.Terms)] = lg.IsTrue(eq.T2)
        }
    }
    return result
}

// termsKey encodes up to 8 ground terms as a fixed-size string array.
// Mirrors Python tuple(eqn.args[0].args).
type termsKey = [8]string

func makeTermsKey(terms []lg.Expr) termsKey {
    var k termsKey
    for i, t := range terms {
        if i >= len(k) {
            break
        }
        k[i] = fmt.Sprint(t)
    }
    return k
}
```

---

## Complete Replacement for `l2s_progress_made` Case (lines 416-448)

Replace the entire `case strings.HasPrefix(name, "l2s_progress_made"):` block:

```go
// Python ivy_l2s.py:1617-1673
case strings.HasPrefix(name, "l2s_progress_made"):
    sfx := name[len("l2s_progress_made"):]
    fmt.Printf("\n\nFailed to prove that work_needed%s decreases when a helpful transition occurs\n", sfx)

    task := tasks[sfx]
    if task == nil {
        break
    }

    // --- Extract nonce predicate names from invar formula ---
    // Python ivy_l2s.py:1621-1636
    // lf.Formula = Implies(And(l2s_saved, eventually_start, exists, not_all_was_done,
    //                          Forall(progress_args, Implies(nad, Not(waiting)))), ...)
    var wasHelpfulNonce, triggerNonce string
    if impl, ok := lf.Formula.(*lg.Implies); ok {
        if ant, ok := impl.T1.(*lg.And); ok && len(ant.Conjuncts) >= 5 {
            allHH := ant.Conjuncts[4]
            if fa, ok := allHH.(*lg.Forall); ok {
                // Forall case (l2s_auto5): body = Implies(nad, Not(waiting_for_progress))
                if bodyImpl, ok := fa.Body.(*lg.Implies); ok {
                    // was_helpful_pred_nonce = body.args[0].rep (= nad.rep = l2s_s_i.Name)
                    if app, ok := bodyImpl.T1.(*lg.Apply); ok {
                        wasHelpfulNonce = app.Func.Name
                    }
                    // trigger_happened_pred_nonce = body.args[1].args[0].rep
                    if notExpr, ok := bodyImpl.T2.(*lg.Not); ok {
                        if app, ok := notExpr.Body.(*lg.Apply); ok {
                            triggerNonce = app.Func.Name
                        }
                    }
                }
            } else {
                // Non-forall variants: Python all_helpful_happened.args[0].rep
                // Direct Apply in antecedent position
                if app, ok := allHH.(*lg.Apply); ok {
                    wasHelpfulNonce = app.Func.Name
                }
                // trigger_happened: all_helpful_happened.args[1].args[0].rep
                // (Not reached for l2s_auto5, handled above)
            }
        }
    }

    // --- Build helpful_map from pre-state (tr.states[0]) ---
    // Python ivy_l2s.py:1627-1632: scan tr.states[0].clauses.fmlas for was_helpful_pred_nonce
    // In Go: pre-state equalities in handler.Eqs are keyed by bare symbol name
    wh := task["work_helpful"]
    helpfulMap := make(map[termsKey]bool)
    if wasHelpfulNonce != "" {
        for _, eqs := range handler.Eqs {
            for _, eqExpr := range eqs {
                eq, ok := eqExpr.(*lg.Eq)
                if !ok {
                    continue
                }
                app, ok := eq.T1.(*lg.Apply)
                if !ok || app.Func.Name != wasHelpfulNonce {
                    continue
                }
                k := makeTermsKey(app.Terms)
                helpfulMap[k] = lg.IsTrue(eq.T2)
                if wh != nil {
                    rep := predLHSRep(wh)
                    fmt.Printf("%s = %v\n", applyPredToVals(rep, app.Terms), eq.T2)
                }
            }
        }
    }

    // --- Build happened_maps for both states ---
    // Python ivy_l2s.py:1637-1644: loop over tr.states[0] and tr.states[1]
    // In Go: states[0] = bare name, states[1] = "new_" + name (actions.New)
    wp := task["work_progress"]
    happenedMaps := [2]map[termsKey]bool{
        make(map[termsKey]bool),
        make(map[termsKey]bool),
    }
    if triggerNonce != "" {
        // Python: for idx in range(2) using same trigger_happened_pred_nonce
        triggerNames := [2]string{triggerNonce, actions.New(triggerNonce)}
        for idx := 0; idx < 2; idx++ {
            tname := triggerNames[idx]
            fmt.Println()
            for _, eqs := range handler.Eqs {
                for _, eqExpr := range eqs {
                    eq, ok := eqExpr.(*lg.Eq)
                    if !ok {
                        continue
                    }
                    app, ok := eq.T1.(*lg.Apply)
                    if !ok || app.Func.Name != tname {
                        continue
                    }
                    k := makeTermsKey(app.Terms)
                    happenedMaps[idx][k] = lg.IsTrue(eq.T2)
                    if wp != nil {
                        rep := predLHSRep(wp)
                        fmt.Printf("~happened %s = %v\n", applyPredToVals(rep, app.Terms), eq.T2)
                    }
                }
            }
        }
    }

    // --- Build justice_map from pre-state (tr.states[0]) ---
    // Python ivy_l2s.py:1645-1652
    justiceMap := make(map[termsKey]bool)
    if jp, ok := justicePredMap[sfx]; ok && jp != nil {
        fmt.Println()
        jpKey := lg.Key(jp)
        for _, eqExpr := range handler.Eqs[jpKey] {
            eq, ok := eqExpr.(*lg.Eq)
            if !ok {
                continue
            }
            app, ok := eq.T1.(*lg.Apply)
            if !ok {
                continue
            }
            k := makeTermsKey(app.Terms)
            justiceMap[k] = lg.IsTrue(eq.T2)
            if wp != nil {
                rep := predLHSRep(wp)
                fmt.Printf("~eventually %s = %v\n", applyPredToVals(rep, app.Terms), eq.T2)
            }
        }
    }

    // --- Diagnostic loop 1 ---
    // Python ivy_l2s.py:1653-1658:
    // if helpful and trigger occurred (states[0]=true → states[1]=false) but work not reduced
    if wh != nil && wp != nil {
        whRep := predLHSRep(wh)
        wpRep := predLHSRep(wp)
    outer1:
        for k, isHelpful := range helpfulMap {
            if !isHelpful {
                continue
            }
            h0, ok0 := happenedMaps[0][k]
            h1, ok1 := happenedMaps[1][k]
            if ok0 && ok1 && h0 && !h1 {
                // Reconstruct display vals from key
                vals := keyToExprs(k)
                fmt.Printf("\nNote: %s is true and %s occurs during the action, but work_needed is not reduced.\n\n",
                    applyPredToVals(whRep, vals), applyPredToVals(wpRep, vals))
                break outer1
            }
        }
    }

    // --- Diagnostic loop 2 ---
    // Python ivy_l2s.py:1659-1664:
    // if helpful and happened (states[0]=true) and eventually (justice=true)
    if wh != nil && wp != nil {
        whRep := predLHSRep(wh)
        wpRep := predLHSRep(wp)
    outer2:
        for k, isHelpful := range helpfulMap {
            if !isHelpful {
                continue
            }
            h0, ok0 := happenedMaps[0][k]
            j, okJ := justiceMap[k]
            if ok0 && okJ && h0 && j {
                vals := keyToExprs(k)
                fmt.Printf("\nNote: %s is true and eventually %s is false.\n\n",
                    applyPredToVals(whRep, vals), applyPredToVals(wpRep, vals))
                break outer2
            }
        }
    }

    // --- work_needed post-state eval (Python ivy_l2s.py:1665-1672) ---
    wn := task["work_needed"]
    if wn != nil {
        vs := predLHSArgs(wn)
        sks := makeSkolems(vs)
        vals := evalSkolems(handler, sks)
        if vals != nil {
            rep := predLHSRep(wn)
            pred := applyPredToVals(rep, vals)
            fmt.Printf("Note: work_invar%s is true and %s changes from false to true.\n\n", sfx, pred)
        }
    }
    // Python line 1673: tr.hidden_symbols = temporal_and_l2s (commented out in Python, omit)
```

---

## Additional Helpers to Add

### `termsKey` type and `makeTermsKey`

```go
// termsKey encodes up to 8 ground terms as a fixed-size string array.
// Mirrors Python tuple(eqn.args[0].args).
type termsKey [8]string

func makeTermsKey(terms []lg.Expr) termsKey {
    var k termsKey
    for i, t := range terms {
        if i >= len(k) {
            break
        }
        k[i] = fmt.Sprint(t)
    }
    return k
}
```

### `keyToExprs`

Used in the display loops to convert a `termsKey` back to `[]lg.Expr` (as string constants)
so `applyPredToVals` can format them:

```go
// keyToExprs reconstructs a []lg.Expr slice from a termsKey for display.
// Each non-empty entry is wrapped as a Const with that string as its name.
func keyToExprs(k termsKey) []lg.Expr {
    var result []lg.Expr
    for _, s := range k {
        if s == "" {
            break
        }
        result = append(result, lg.NewConst(s, lg.Uninterpreted("_")))
    }
    return result
}
```

Note: `applyPredToVals` calls `fmt.Sprint(v)` on each val, so passing a `*lg.Const` with
`.Name = s` will render as `s`. Alternatively, use a raw-string wrapper type. Verify that
`fmt.Sprint(lg.NewConst("foo", ...))` produces `"foo"` by checking the Const Stringer.

---

## Imports Required in `l2s_hooks.go`

Add `"github.com/glycerine/ivy/goivy/actions"` to the import block (for `actions.New`).

Current imports (line 6-13):
```go
import (
    "fmt"
    "strings"
    "github.com/glycerine/ivy/goivy/ast"
    lg "github.com/glycerine/ivy/goivy/logic"
    lu "github.com/glycerine/ivy/goivy/logicutil"
)
```

Add:
```go
    actions "github.com/glycerine/ivy/goivy/actions"
```

---

## Files to Edit

- `check/l2s_hooks.go` — only file; add two helpers (`termsKey`/`makeTermsKey`, `keyToExprs`)
  near the other helper functions (~line 288), then replace lines 416-448.

---

## Verification

```
cd ~/ivy/goivy && make test
```

Focus: any test exercising `l2s_auto5` with a `l2s_progress_made` failure path.
If no test covers this, add a minimal one after confirming the diagnostics print correctly
against an actual Ivy file.

**During implementation:** Add an `xtracer.Trace` after extracting `wasHelpfulNonce` and
`triggerNonce` to confirm the formula navigation succeeds on a live example before
the map-building code runs.
