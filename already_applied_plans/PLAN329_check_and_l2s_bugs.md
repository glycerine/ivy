# Audit Plan: ivy_check.py and ivy_l2s.py vs. Go Port
*Created: 2026-04-17 (Thursday)*

## Context

This plan is a **systematic, line-by-line audit** comparing Python source truth files:
- `~/ivy/pyivy/ivy/ivy/ivy_check.py`
- `~/ivy/pyivy/ivy/ivy/ivy_l2s.py`

Against the Go port files:
- `~/ivy/goivy/check/check.go`
- `~/ivy/goivy/check/isolate_check.go`
- `~/ivy/goivy/check/l2s.go`
- `~/ivy/goivy/check/l2s_shared.go`
- `~/ivy/goivy/check/l2s_hooks.go`
- `~/ivy/goivy/check/helpers.go`
- `~/ivy/goivy/check/l2s_auto.go` (partially audited)

The purpose is NOT to fix bugs yet — it is to **produce an exhaustive list** of every place where Go diverges from Python, is missing something, or is faithfully identical but has a comment flagging an issue. Prior porting history shows "simplified" Go code left millions of hidden bugs; this audit roots them out.

---

## Audit Results — ivy_check.py

### A. CONFIRMED BUGS IN GO PORT

**A1. `CheckFcsInState` no-ag path silently discards SMT results** — `check.go:670-682`

Python always calls `check_fcs_in_state(mod, ag, post, fcs)` with real `ag` and `post`. If Go's `CheckFcsInState` (the no-ag variant) is ever called, it calls `gmc(combined, finalConds)` but **discards the result** — checker `Sat()`/`Unsat()` callbacks are never invoked, so no checker ever fails. This is a latent bug.

*Files:* `check.go:664-683`
*Severity:* HIGH if no-ag path is ever triggered

**A2. Python `check_isolate` initializer guarantee list comprehension is a Python bug; Go behavior diverges**

Python `ivy_check.py` lines ~635-645:
```python
if mod.initializers:
    guarantees = [sub for sub in action.iter_subactions()
                  if isinstance(sub,(act.AssertAction,act.Ranking))
                  for action in mod.initializers]
```
In Python 3, `action.iter_subactions()` at the START of the comprehension uses `action` from the **outer scope** (the last `env_action` from the preceding checked-actions loop), NOT from `for action in mod.initializers`. This is a Python 3 scoping bug in the original — the comprehension iterates the initializerslist but applies `iter_subactions()` on the WRONG action.

Go's `isolate_check.go:284-358` correctly iterates each initializer action separately. **Go is more correct than Python here, but it diverges.** The test oracle should know about this.

*Files:* `isolate_check.go:284-358`
*Severity:* MEDIUM — Go is BETTER than Python, but produces different output/behavior

**A3. `ApplyConjProofs` known type-system gap silently swallows proof applications**

`check.go:396-400` has the comment:
```
// The proof package uses ast.LabeledFormula (with ast.Node fields) while
// module uses ast.LabeledFormula (with lg.Expr fields). These are separate
// type hierarchies — a porting mistake.
```
When `ModuleLFToAstLF` returns nil (type conversion fails), the conjecture silently passes through without proof application. Python has ONE `LabeledFormula` class. **This is a structural porting mistake that causes silent proof-application failures.**

*Files:* `check.go:397-440`
*Severity:* HIGH — proofs may silently not be applied

**A4. `ConvertPostconds` (no-ag version) never renames old symbols**

`check.go:807-809`: `ConvertPostconds(postconds)` calls `ConvertPostcondsWithUpdate(nil, postconds)` with `nil` update. `ConvertPostcondsWithUpdate` returns early at line 817 when `update == nil`. Python's `convert_postconds(state, postconds)` always builds renamings from `state.update`. `CheckConjsInState` (the no-ag version, called from `CheckConjsInStateWithAG`) uses this no-rename path. **Postcondition renaming is silently skipped in Go.**

*Files:* `check.go:807-854`
*Severity:* MEDIUM — postcondition checks may use wrong variable names

**A5. `opt_ivy_stats` boolean vs Parameter object in Python**

Python `check_module` line ~975: `if opt_ivy_stats:` — this evaluates the `Parameter` *object* as a boolean, which is ALWAYS truthy (parameter objects are non-None). This is a Python bug in the original (should be `if opt_ivy_stats.get():`). All other usages in the file correctly use `.get()`. Go uses `mod.Cfg.OptIvyStats` (a proper bool). **Go behavior is correct; Python may spuriously print stats.**

*Files:* Python bug at `ivy_check.py:~975`; Go at `isolate_check.go:980`
*Severity:* LOW

---

### B. MISSING FUNCTIONS — NOT PORTED

**B1. `has_temporal_stuff(f)` — not ported**

Python `ivy_check.py:119-120`:
```python
def has_temporal_stuff(f):
    return any(True for x in lut.temporals_ast(f)) or any(True for x in lut.named_binders_ast(f))
```
Checks whether formula `f` contains temporal operators or named binders. Not present in Go. Low impact since it's only used in comments in Python (the actual call site is commented out), but it should exist.

*Severity:* LOW

**B2. `find_assertions(action_name=None)` — not ported**

Python `ivy_check.py:189-197`:
```python
def find_assertions(action_name=None):
    res = []
    actions = act.call_set(action_name, im.module.actions) if action_name else list(im.module.actions.keys())
    for actname in actions:
        action = im.module.actions[actname]
        for a in action.iter_subactions():
            if isinstance(a,act.AssertAction) or isinstance(a,act.Ranking):
                res.append(a)
    return res
```
Returns all assertion actions in the module. Not ported to Go.

*Severity:* LOW (utility/debug function)

**B3. `show_assertions()` — not ported**

Python `ivy_check.py:198-200`:
```python
def show_assertions():
    for a in find_assertions():
        print('{}: {}'.format(a.lineno,a))
```
Prints all assertions with line numbers. Not ported.

*Severity:* LOW (utility/debug function)

---

### C. BEHAVIORAL DIFFERENCES (SUBTLE)

**C1. `MatchHandler.handle` env iteration — potential incorrect symbol hiding**

Python (`ivy_check.py:338-353`):
```python
for sym, renamed_sym in env.items():
    if not itr.is_new(sym) and not self.is_skolem(sym):
        self.show_sym(sym, renamed_sym)
```
Python's `env` is `dict[Symbol, Symbol]` — original → renamed.

Go's `env` is `map[lg.NodeKey]lg.Expr` — key is original symbol's key, value is renamed symbol. Go reconstructs the original symbol from `vocabByKey[symKey]`, falling back to `renamedSym` if not found in vocab. This fallback may cause `IsNew`/`IsSkolem` to be checked on the WRONG symbol.

*Files:* `helpers.go:303-323`
*Severity:* LOW — only affects trace display

**C2. `MatchHandler.Eval` delegates to HerbrandModel vs direct model evaluation**

Python calls `model.eval_to_constant(cond)` directly. Go builds a `HerbrandModel` wrapper via `solver.NewHerbrandModel`. Behavioral parity depends on the HerbrandModel implementation.

*Files:* `helpers.go:244-258`
*Severity:* LOW

**C3. `check_module` dead `cact = checked_action.get()` at end — correctly omitted**

Python's last line of `check_module` reads `cact = checked_action.get()` but never uses it. Go correctly omits this dead assignment.

*Severity:* NONE (correctly omitted dead code)

---

## Audit Results — ivy_l2s.py

### D. CONFIRMED BUGS IN GO PORT

**D1. Plain `l2s` and `l2s_full` tactics MISSING `renaming_hook`**

Python `ivy_l2s.py:1500-1508` (inside `l2s_tactic_int`):
```python
if tactic_name.startswith("l2s_auto5"):
    goal.trace_hook = lambda tr,fcs: auto_hook(tasks,triggers,subs,tr,fcs)
else:
    goal.trace_hook = lambda tr,fcs: renaming_hook(subs,tr,fcs)
```
The `else` branch covers ALL non-auto5 tactics including plain `l2s`, `l2s_full`, `l2s_auto`, `l2s_auto2`, `l2s_auto3`, `l2s_auto4`.

Go (`l2s.go:895-916`) only sets trace hooks for `l2s_auto5` and `l2s_auto*`. **Neither plain `l2s` nor `l2s_full` get a `TraceHookFn` in Go.** When these tactics fail, no symbol renaming is applied to the trace, so the user sees fresh names (e.g., `l2s_g_0`, `l2s_s_1`) instead of the original named-binder expressions.

*Files:* `l2s.go:895-916`
*Severity:* MEDIUM — degrades error trace readability for `l2s` and `l2s_full` tactics

**D2. `l2s_tactic_full` trace hook for loop-start marking is missing**

Python `ivy_l2s.py:91-95`:
```python
def l2s_tactic_full(prover,goals,proof):
    goals = l2s_tactic(prover,goals,proof,"l2s_full")
    goals[0].trace_hook = trace_hook
    return goals
```
Python's `trace_hook` (lines 99-109) scans trace states to find when `l2s_saved == true` and marks the previous state as `loop_start`. This helps users identify the start of the fairness loop in counterexample traces.

Go acknowledges this is missing in a comment at `l2s.go:911-914`:
```go
case tacticName == "l2s_full":
    // No MatchHandler-side hook for l2s_full — Python's trace_hook
    // marks loop_start on a trace.TraceBase, which we don't build
    // in this path. Leave TraceHook nil.
```

*Files:* `l2s.go:911-914`
*Severity:* MEDIUM — loop start not marked in `l2s_full` error traces

**D3. `auto_hook` for `l2s_progress_made` only does partial diagnosis**

Python `ivy_l2s.py:1453-1509` builds:
1. `helpful_map` from `tr.states[0].clauses.fmlas` (pre-state)
2. `happened_maps` from BOTH `tr.states[0]` and `tr.states[1]` (pre AND post)
3. `justice_map` from `tr.states[0].clauses.fmlas`
4. Prints two diagnostic loops cross-referencing all three maps

Go's `diagnoseAutoFailure` in `l2s_hooks.go:366-398` for this case only evaluates Skolem symbols in the handler's post-state and prints the `work_needed` diagnostic. The `helpful_map` / `happened_maps` / `justice_map` multi-state analysis is **not implemented**.

*Files:* `l2s_hooks.go:366-398`
*Severity:* MEDIUM — partial diagnostics for `l2s_auto5` progress failures

**D4. `ls2_g_to_globally` pretty-printer function not ported**

Python `ivy_l2s.py:1520-1526`:
```python
def ls2_g_to_globally(ast):
    def g2g(ast):
        if isinstance(ast,lg.NamedBinder) and ast.name == 'l2s_g':
            return lg.Globally(ast.environ,ast.body)
        return None
    res = ilu.expand_named_binders_ast(ast,g2g)
    return ilu.denormalize_temporal(res)
```
Used in `auto_hook` to set `tr.pp = ls2_g_to_globally` — converts `l2s_g` binders back to `Globally` operators for readable display in error traces. Not ported to Go; `applyAutoDiagnosticsToHandler` never sets a pretty-printer on the handler.

*Files:* `l2s_hooks.go` — missing
*Severity:* MEDIUM — `l2s_g_N` names appear in traces instead of `globally(...)` for `l2s_auto5`

**D5. `auto_hook` does not set `tr.pp` for `l2s_auto5`**

Python `ivy_l2s.py:1529`: `tr.pp = ls2_g_to_globally` is set right after `renaming_hook`. Go never sets a pretty-printer on `MatchHandler`. This is the companion to D4.

*Files:* `l2s_hooks.go:56-90`
*Severity:* MEDIUM

**D6. `TemporalModels` conclusion comparison: `lg.And()` (Python) vs `lg.True` (Go)**

Python `ivy_l2s.py:1491`: `conc = ivy_ast.TemporalModels(model, lg.And())`
Go `l2s_shared.go:841`: `newConc := acfg.NewTemporalModels(model, lg.True)`

Python uses an empty `And()` (which equals True semantically) while Go uses the explicit `lg.True` constant. These are semantically equivalent in the SMT solver but may produce different canon strings in s-expression output, potentially causing canon mismatches in cross-language regression tests.

*Files:* `l2s_shared.go:841`
*Severity:* LOW-MEDIUM — potential canon mismatch; semantically correct

---

### E. MISSING FUNCTIONS — NOT PORTED

**E1. `temporal_and_l2s(sym)` — PORTED** ✓

Go: `l2s_hooks.go:28-31` as `TemporalAndL2S(name string) bool`. Correct.

**E2. `ls2_g_to_globally` — NOT PORTED** (see D4 above)

---

### F. BEHAVIORAL DIFFERENCES (SUBTLE)

**F1. `proof.tactic_lets` check**

Python `ivy_l2s.py:124-125`:
```python
if proof.tactic_lets:
    raise iu.IvyError(proof,'tactic does not take lets')
```
Go `l2s.go:413-417`:
```go
if tt, ok := pf.(*ast.TacticTactic); ok {
    if tt.Body != nil {
        if _, isLets := tt.Body.(*ast.TacticLets); isLets {
            return nil, fmt.Errorf("tactic does not take lets")
        }
    }
}
```
Python's `proof.tactic_lets` accesses a direct attribute. Go checks if `pf` is a `*ast.TacticTactic` with a non-nil `Body` that is a `*ast.TacticLets`. If `pf` is NOT a `*ast.TacticTactic`, Go silently skips this check. Needs verification that `proof.tactic_lets` is always a `TacticTactic`.

*Files:* `l2s.go:413-417`
*Severity:* LOW

**F2. `modPass` binding transformation: Python passes ActionTerm vs Go passes Stmt**

Python `ivy_l2s.py:801-805`:
```python
model.bindings[i] = b.clone([transform(b.action)])
```
Python passes `b.action` (an `ActionTerm` AST node) to transform, so named binders IN the ActionTerm wrapper itself are transformed.

Go `l2s.go:570-572`:
```go
newAction := transform(b.Action).(*temporal.ActionTerm)
model.Bindings[i] = b.CloneAction(newAction)
```
Go passes `b.Action` (ActionTerm) to transform. This should match Python. ✓

**F3. `collectAllNamedBinders` passes ActionTerm vs Stmt**

Python `ivy_l2s.py:1437-1443` calls `ilu.named_binders_asts` on `[b.action for b in model.bindings]` — the ActionTerm objects.

Go `l2s.go:1005-1006` in `collectAllNamedBinders` calls `collectActionNBs(b.Action.Stmt, result)` — the inner Stmt, not the ActionTerm wrapper. Named binders at the ActionTerm level (if any) would be missed.

*Files:* `l2s.go:1005-1006`
*Severity:* LOW (ActionTerm wrapper unlikely to contain named binders)

**F4. `temporal_prems` construction — combined from goal and prover**

Python `ivy_l2s.py:162-166`:
```python
temporal_prems = [x for x in ipr.goal_prems(goal) if hasattr(x,'temporal') and x.temporal] + [
    x for x in prover.axioms if not x.explicit and x.temporal]
```
Python uses `hasattr(x,'temporal') and x.temporal` — only LF objects with a `temporal` attribute that is truthy.

Go `l2s.go:337-355` checks `lf.IsTemporal()` which returns true if `Temporal` pointer is non-nil and `*lf.Temporal == true`. If `Temporal` is nil (not set), `IsTemporal()` returns false, same as Python's `hasattr` check returning false. Functionally equivalent.

*Files:* `l2s.go:337-355`
*Severity:* VERY LOW

**F5. `assumed_gprops` cloning — Python `p.clone([p.label, p.formula.args[0]])`**

Python `ivy_l2s.py:157`:
```python
cloned = p.clone([p.label, p.formula.args[0]])
```
For a `Globally(body)` formula, `p.formula.args[0]` is `body`.

Go `l2s.go:383`:
```go
cloned := ax.Clone([]ast.Node{ax.Label, g.Body}).(*ast.LabeledFormula)
```
`g.Body` is `Globally.Body` which is the same as `p.formula.args[0]` in Python. Equivalent. ✓

**F6. `desugar` applies to `invars` in Python BEFORE committing to model.invars**

Python `ivy_l2s.py:760`: `invars = list(map(desugar, invars))`
Then line 766: `model.invars = model.invars + invars`

Go `l2s.go:503-515`: desugars `invars` in place first (loop), then appends to `model.Invars`. ✓

**F7. `l2s_auto` invariant ordering — Python `sorted_tasks`**

Python `ivy_l2s.py:310`: `sorted_tasks = list(sorted(x for x in tasks))`
Go `l2s_auto.go` (not fully read): should use equivalent sorting. **Needs verification of l2s_auto.go against the full auto invariant generation code (lines 243-726 of ivy_l2s.py).** This is a large section (>500 lines) that was not fully audited here.

*Files:* `l2s_auto.go`
*Severity:* UNKNOWN — needs dedicated audit pass

---

## Summary Table

| ID | File | Issue | Severity |
|----|------|-------|----------|
| A1 | check.go:670-682 | No-ag path silently discards SMT results | HIGH |
| A2 | isolate_check.go:284-358 | Python initializer comprehension bug — Go is more correct but diverges | MEDIUM |
| A3 | check.go:397-440 | `ApplyConjProofs` type-system gap silently skips proof application | HIGH |
| A4 | check.go:807-854 | `ConvertPostconds` never renames old symbols (nil update) | MEDIUM |
| A5 | Python bug | `opt_ivy_stats` evaluated as object not value | LOW |
| B1 | missing | `has_temporal_stuff` not ported | LOW |
| B2 | missing | `find_assertions` not ported | LOW |
| B3 | missing | `show_assertions` not ported | LOW |
| C1 | helpers.go:303-323 | MatchHandler env iteration fallback may check wrong symbol | LOW |
| C2 | helpers.go:244-258 | MatchHandler.Eval uses HerbrandModel wrapper | LOW |
| D1 | l2s.go:895-916 | Plain `l2s` and `l2s_full` missing `renaming_hook` | MEDIUM |
| D2 | l2s.go:911-914 | `l2s_full` loop-start marking missing | MEDIUM |
| D3 | l2s_hooks.go:366-398 | `l2s_progress_made` partial diagnosis (multi-state missing) | MEDIUM |
| D4 | l2s_hooks.go (missing) | `ls2_g_to_globally` pretty-printer not ported | MEDIUM |
| D5 | l2s_hooks.go:56-90 | `auto_hook` doesn't set pretty-printer for `l2s_auto5` | MEDIUM |
| D6 | l2s_shared.go:841 | `lg.And()` vs `lg.True` in TemporalModels conclusion | LOW-MEDIUM |
| E2 | missing | `ls2_g_to_globally` function not ported | MEDIUM |
| F1 | l2s.go:413-417 | `tactic_lets` check skipped for non-TacticTactic proof nodes | LOW |
| F3 | l2s.go:1005-1006 | `collectAllNamedBinders` passes Stmt instead of ActionTerm | LOW |
| F7 | l2s_auto.go | Auto invariant generation (500+ lines) not fully audited | UNKNOWN |

---

## Critical Next Steps (Implementation)

1. **A3** (HIGH): Unify `ast.LabeledFormula` type system so `ModuleLFToAstLF` never fails silently. This is a structural fix needed everywhere.

2. **A1** (HIGH): Fix `checkFcsNormalPath` no-ag fallback to actually invoke checker callbacks. Or ensure `CheckFcsInState` (no-ag) is never called outside tests.

3. **D1** (MEDIUM): Add `TraceHookFn` with `renaming_hook` behavior for plain `l2s` (non-auto) and `l2s_full` in `l2sTacticInt`.

4. **D4/D5** (MEDIUM): Port `ls2_g_to_globally` and set it as `MatchHandler.PP` for `l2s_auto5` failures.

5. **F7** (UNKNOWN): Perform a dedicated full audit of `l2s_auto.go` vs Python `ivy_l2s.py:243-726`.

6. **A4** (MEDIUM): Thread the post-state's update into `CheckConjsInStateWithAG` → `ConvertPostcondsWithUpdate` so old-symbol renaming works.

---

## Verification

To verify fixes:
```
cd ~/ivy/goivy && make test
```
Focus regression tests:
- `check/` package tests with temporal properties
- Any test using `l2s`, `l2s_full`, `l2s_auto*` tactics
- Tests with postconditions on checked actions

Cross-language canon comparison via `xtracer` golden files in `test_vectors/`.
