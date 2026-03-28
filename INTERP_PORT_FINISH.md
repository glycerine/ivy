# INTERP_PORT_FINISH.md — Remaining Tasks to Complete the interp/ Port

**Audit date:** 2026-03-28
**Python source:** `~/pyivy/ivy/ivy/ivy_interp.py` (642 lines)
**Go port:** `~/go/src/github.com/glycerine/goivy/interp/` (interp.go, eval.go, helpers.go, phase4.go)

---

## Legend

- **STUB**: Function exists but body is empty, returns nil/zero, or uses placeholder logic
- **DIVERGED**: Function exists but its semantics differ materially from Python
- **MISSING**: Python function has no Go counterpart at all
- **OK**: Faithfully ported (may have minor idiomatic Go differences)

---

## 1. Critical Stubs (empty or placeholder bodies)

### 1.1 State.Update() lazy computation — STUB
**File:** `interp.go:144-151`
**Python:** `State.update` property (line 137-140) calls `eval_action(self.expr.rep).update(self.domain, self.in_scope)` to lazily compute the update on first access.
**Go:** The `if` body at lines 145-149 is **empty** — it has the comment but no code. `Update()` only returns `CachedUpdate` which is nil unless someone else sets it.
**Impact:** Any caller that relies on lazy update computation (e.g. `Reverse`, `ReverseUpdateConcreteClauses`, `ReachState`) will get nil and fail silently or error out.
**Fix:** Implement the body: call `EvalAction(atom.Rep, s.Domain)` to get the action, then call its `Update(domain, inScope)` method (not `IntUpdate`), and cache the result.

### 1.2 FailAction — missing Update(), IntUpdate(), Decompose() — STUB
**File:** `helpers.go:53-124`
**Python:** `fail_action` class (lines 375-399) has three critical methods:
- `update(self, domain, in_scope)` → calls `action_failure(self.action.update(domain, in_scope))`
- `int_update(self, domain, in_scope)` → calls `action_failure(self.action.int_update(domain, in_scope))`
- `decompose(self, pre, post)` → complex logic: decomposes inner action, wraps each prefix's last action in `fail_action`

**Go:** `FailAction` has **none** of these methods. Its `Decompose()` just returns `[][]actions.Action{{fa}}` — a trivial stub that doesn't match the Python logic at all.
**Impact:** `FailAction` cannot be used for action failure analysis, which is the entire purpose of the class. Any code path that evaluates a `fail_action` update will get wrong results.
**Fix:** Add `Update(ctx *actions.UpdateContext) *tr.Update` and `IntUpdate(ctx *actions.UpdateContext) *tr.Update` methods that call `tr.ActionFailure(inner.Update(ctx))` / `tr.ActionFailure(inner.IntUpdate(ctx))`. Rewrite `Decompose()` to match the Python prefix-wrapping logic.

### 1.3 CaseConjecture — missing interpolant_case call — STUB
**File:** `helpers.go:334-362`
**Python** (lines 323-335): calls `interpolant_case(pre, clauses, axioms, state.domain.functions)` to compute a Craig interpolant separating the under-approximation from the clauses. Returns `(core, interp)` tuple.
**Go:** Does NOT call `interpolant_case` (or any equivalent from `transrel`). Instead uses a simple `z3bridge.Implies` check and then uses the negation of clauses as the "interpolant."
**Impact:** The interpolant quality is much worse — the Go version doesn't compute a true Craig interpolant, so abstraction refinement loops will not converge properly.
**Fix:** Call `tr.InterpolantCase(pre, clauses, axioms, interpreted)` and return the real `(core, interp)` result. The `tr.InterpolantCase` function must exist or be ported in the `transrel` package.

---

## 2. Significant Semantic Divergences

### 2.1 concrete_post — reimplemented instead of calling compose_state_action — DIVERGED
**File:** `eval.go:27-72`
**Python** (lines 195-207): single call `compose_state_action(state.value, axioms, update, check=context.check)` which handles everything: precondition checking, forward image, annotation propagation, moded symbol handling.
**Go:** Manually reimplements the logic: extracts formulas, checks precondition via z3bridge, calls `tr.ForwardImage`, creates clauses from result. This diverges from the Python in several ways:
1. Python's `compose_state_action` handles moded symbols (frame conditions for unmodified symbols). Go's manual implementation may not handle this correctly.
2. Python's version uses the solver-level `ActionFailed` exception which carries `clauses` and `trans` fields. Go's `tr.ActionFailed` has `Formula` and `Trace` fields — the error mapping may lose information.
3. Python's version passes the full state value triple (moded, clauses, precond) to compose_state_action. Go only passes `state.Clauses.ToFormula()`.

**Fix:** Either (a) port `compose_state_action` faithfully to Go's `transrel` package and call it here, or (b) verify that the manual implementation covers all edge cases including moded symbols and annotation propagation.

### 2.2 ReachState — missing model extraction — DIVERGED
**File:** `helpers.go:215-243`
**Python** (lines 282-303): uses `get_model_clauses(img)` to extract a model, then `find_true_disjunct(pre, m.eval)` to identify which under-approximation disjunct is satisfied, then `clauses_model_to_clauses(img, ignore, model=m)` to convert the model to clauses, and records the universe `dict((s, [c.skolem() for c in m.sort_universe(s)]) for s in m.sorts())`.
**Go:** Uses `z3bridge.IsSat()` only to check satisfiability, then uses `imgClauses` directly as the under-approximation. This means:
1. **No model extraction** — the resulting under-approximation is the entire image, not a single concrete state. This is unsound for the reachability analysis.
2. **No disjunct tracking** — `pred` is always nil instead of pointing to the specific under-approximation predecessor.
3. **No universe extraction** — universe is always nil instead of recording sort universes from the model.

**Fix:** Use the solver's `GetModelClauses` to extract a model, use `FindTrueDisjunct` to identify the predecessor, use `ClausesModelToClauses` to extract the concrete reachable state, and record the universe.

### 2.3 ReverseJoinConcreteClauses — structural vs solver check — DIVERGED
**File:** `phase4.go:103-129`
**Python** (line 256): `not clauses_imply(and_clauses(s.clauses, clauses), false_clauses())` — uses the solver to check if the conjunction is satisfiable (not implies false = satisfiable).
**Go** (line 109): `!combined.IsFalse()` — structural check only (checks if the Clauses object literally represents `false`). This will almost always return true even when the conjunction is logically unsatisfiable, because `IsFalse()` only checks structural falsity, not logical satisfiability.
**Fix:** Replace `combined.IsFalse()` with a solver call: `solver.ClausesImply(combined, co.FalseClauses(nil))` or equivalent.

### 2.4 JoinUnders — missing tagged_or_clauses — DIVERGED
**File:** `helpers.go:183-193`
**Python** (line 268): `tagged_or_clauses('__pre', *[s.clauses for s in state.unders])` — uses tagged disjunction which labels each disjunct with `__pre` prefix for later identification via `find_true_disjunct`.
**Go:** `co.OrClausesTyped(clauses...)` — plain disjunction, no tagging.
**Impact:** Without tagging, `find_true_disjunct` (used in `reach_state`) cannot identify which under-approximation disjunct was satisfied. This is related to the `ReachState` model extraction gap (2.2).
**Fix:** Port `tagged_or_clauses` to `clauseops` and use it here.

### 2.5 Diagram — missing parameters — DIVERGED
**File:** `helpers.go:372-389`
**Python** (lines 337-345): `clauses_model_to_diagram(clauses, is_skolem, implied, axioms=axioms, weaken=weaken, upward_close=upward_close)` — passes `implied`, `weaken`, and `upward_close` parameters.
**Go:** `slv.ClausesModelToDiagram(clauses, isSkolem, axioms)` — missing `implied`, `weaken`, `upward_close` parameters.
**Fix:** Extend the solver's `ClausesModelToDiagram` to accept these parameters and pass them through.

### 2.6 ApplyAction — uses IntUpdate instead of Update — DIVERGED
**File:** `eval.go:147-184`
**Python** (line 357): `action.update(state.domain, state.in_scope)` — calls the action's `update()` method.
**Go** (line 164): `actions.IntUpdate(action, ctx)` — calls `IntUpdate` instead of `Update`.
**Note:** In Python, `update()` wraps `int_update()` with additional logic (e.g., `hide_formals`). Using `IntUpdate` directly skips this wrapping. The Python `apply_action` specifically uses `update()`, not `int_update()`.
**Fix:** Add/use an `Update()` wrapper that matches Python's `Action.update()` behavior (calls `int_update` then `hide_formals`).

### 2.7 fail_expr — string prefix instead of FailAction wrapping — DIVERGED
**File:** `helpers.go:661-663`
**Python** (line 401-402): `action_app(fail_action(expr.rep), expr.args[0])` — creates a `fail_action` object wrapping the evaluated action, then wraps it in an action application.
**Go:** `ActionApp(cfg, "fail_"+expr.Rep, expr.Terms[0])` — just prepends "fail_" to the name string. This creates a regular action application with a modified name, NOT a FailAction wrapper.
**Fix:** Create a `NewFailAction` wrapping the evaluated action from `expr.Rep`, then wrap it in an action application node.

### 2.8 EvalAssertRhs — incomplete RME handling — DIVERGED
**File:** `phase4.go:280-311`
**Python** (lines 600-605): wraps non-RME in `RME(And(), None, rhs)`, then evaluates via `eval_state(rhs)` within an `ActionContext(domain)`.
**Go:** Has divergent paths for `lg.Expr` and `ast.Node`, constructs an RME but then only uses `rmeVal.Ensures` to create a state — ignoring `Requires` (which should become the precondition) and `Modifies` (which should set moded symbols). The `ActionContext` is created but not actually used (`ctx` is assigned to `_`).
**Fix:** Properly evaluate the RME through `EvalState` within an active ActionContext, matching the Python flow. The RME's three components (requires, modifies, ensures) must all contribute to the resulting state value.

### 2.9 FalseProperties — different checking logic — DIVERGED
**File:** `helpers.go:522-556`
**Python** (lines 554-563): uses `islvr.check_sequence(aas, reporter)` which builds a sequence of `Assume`/`Assert` objects and checks them incrementally. The solver accumulates assumptions and checks each assertion against all prior assumptions.
**Go:** Manually iterates properties, checks each with `z3bridge.Implies(premise, prop)` where premise is accumulated AND of axioms + prior subgoals. Also lacks `reporter` callback support.
**Functional difference:** The solver-level `check_sequence` may handle assertion checking differently than manual `Implies` calls (e.g., incremental solving, better counterexample extraction).
**Fix:** Port `check_sequence` pattern from `ivy_solver` to Go's `solver` package and use it here. Add `reporter` callback parameter.

### 2.10 UndecidedConjectures — individual vs batch checking — DIVERGED
**File:** `helpers.go:280-298`
**Python** (lines 548-551): uses `clauses_imply_list(clauses1, state1.conjs)` which batch-checks all conjectures in a single solver session.
**Go:** Creates a new `z3bridge.NewTranslator()` for each conjecture in a loop — O(n) solver instances instead of one.
**Fix:** Use a batch `ClausesImplyList` or keep one translator open for all checks.

### 2.11 HistorySatisfy — missing optional parameters — DIVERGED
**File:** `helpers.go:421-424`
**Python** (lines 593-598): `history.satisfy(bg, _get_model_clauses, final_cond)` — supports custom model extraction callback and final condition.
**Go:** `history.Satisfy(axioms.ToFormula())` — only passes background theory. The `_get_model_clauses` and `final_cond` parameters are dropped.
**Fix:** Thread these parameters through. If Go's `History.Satisfy` doesn't accept them yet, extend its signature.

---

## 3. Missing Functions / Aliases

### 3.1 reach_state_from_pred_no_abduction — MISSING
**Python** (line 321): `reach_state_from_pred_no_abduction = reach_state` — an alias.
**Go:** No alias exists.
**Fix:** Add `var ReachStateFromPredNoAbduction = ReachState` or a wrapper function.

### 3.2 new_state delegation to ActionContext — MISSING
**Python** (lines 188-190): `new_state(value, exact, domain, expr)` delegates to `ivy_actions.context.new_state(...)`.
**Go:** Uses direct `NewState`/`NewStateFromClauses` constructors. The delegation through `ActionContext` is absent.
**Impact:** Python's `ActionContext.new_state` may do additional bookkeeping (e.g., tracking created states, applying context-specific transformations). If so, the Go constructors miss this.
**Fix:** Verify what `ivy_actions.context.new_state` does beyond plain construction. If it's trivial, document the decision. If it's non-trivial, add the delegation.

### 3.3 Interp type alias — MISSING (low priority)
**Python** (line 35): `Interp = im.Module` — type alias.
**Go:** No alias. Low priority since it's just a convenience alias.

### 3.4 EvalContext as a context manager (global context stack) — MISSING
**Python** (lines 162-176): `EvalContext` is a context manager that pushes/pops a global `context` variable. `concrete_post` reads `context.check`.
**Go:** `InterpConfig` (which would manage this) is commented out. Go passes `checkPrecond bool` as a function parameter instead.
**Assessment:** The Go approach (explicit parameter) is better for thread safety per CLAUDE.md rules. However, the threading must be verified: every call site that in Python reads `context.check` must in Go receive and forward the `checkPrecond` parameter correctly. Audit all call chains.

---

## 4. Subtle Bugs / Mismatches

### 4.1 eval_state_atom — RME handling path incomplete
**File:** `eval.go:196-237`
**Python** (lines 459-465): For RME expressions, extracts `req, mod, ens = expr.args`, computes `mod_syms = [a.rep for a in mod]`, `ens_clauses = ens`, `req_clauses = Not(req)`, creates state with `(mod_syms, ens_clauses, req_clauses)`.
**Go:** Uses a `stateProvider` interface to check if an action wraps a state. Does NOT handle the RME constructor path where an RME node appears in the expression tree.
**Fix:** Add explicit RME handling: check if `expr` is an `*actions.RME`, extract its three fields, construct a `StateValue{Moded: modSyms, Clauses: ensClauses, Precond: reqClauses}`.

### 4.2 ModuleTypeCheckConcepts — arity vs sort mismatch
**File:** `helpers.go:473-512`
**Python** (line 52): `self.relations.update((x.rep, len(x.args)) for x, y in self.concept_spaces)` — adds `(name, arity)` pairs where arity is an int.
**Go** (lines 494-501): Adds `(name, lg.Sort)` pairs from the label expression's sort. Python's `relations` maps names to arity integers; Go's maps names to `lg.Sort`. The semantics differ.
**Fix:** Verify what `mod.Relations` actually stores in Go. If it stores `lg.Sort`, the concept space entries should use a sort that encodes the correct arity. If it stores arity, change the code.

### 4.3 CheckStateAssertion — different structure access
**File:** `phase4.go:350-372`
**Python** (lines 615-622): accesses `assertion.args[0].rep` for label and `assertion.args[1]` for RHS.
**Go:** accesses `assertion.Label` and `assertion.Formula` — assumes `LabeledFormula` structure.
**Assessment:** This is OK if Go's `LabeledFormula` is the correct representation for assertions. Verify that `mod.Assertions` stores `*ast.LabeledFormula` and that `Label` and `Formula` correspond to Python's `args[0]` and `args[1]`.

### 4.4 GetStateAssertions — tuple indexing vs struct field
**File:** `phase4.go:380-405`
**Python** (line 631): `res = and_clauses(res, rhs_state[1])` — indexes into the state value tuple to get clauses.
**Go** (line 398): `co.AndClausesTyped(res, rhsState.Clauses)` — accesses Clauses field.
**Assessment:** Correct if `rhs_state[1]` corresponds to `rhsState.Clauses` (it does: Python value tuple is `(moded, clauses, precond)`).

### 4.5 concrete_join — Python has a bug that Go avoids
**Python** (line 214): `domain.background_theory()` uses bare `domain` which is a local variable name shadowed in the function — likely should be `state1.domain.background_theory()`.
**Go** (line 83): Correctly uses `s1.Domain.BackgroundTheory(nil)`.
**Assessment:** Go is correct here; Python has a latent bug.

---

## 5. Task Priority List

### P0 — Blocks basic functionality

| # | Task | Files | Effort |
|---|------|-------|--------|
| 1 | Implement State.Update() lazy computation body | interp.go:145-149 | Small |
| 2 | Add FailAction.Update() and FailAction.IntUpdate() methods | helpers.go | Medium |
| 3 | Rewrite FailAction.Decompose() to match Python prefix-wrapping logic | helpers.go | Medium |
| 4 | Fix ReverseJoinConcreteClauses to use solver instead of structural IsFalse() | phase4.go:109 | Small |
| 5 | Fix ApplyAction to call Update() (not IntUpdate()) on the action | eval.go:164 | Small |
| 6 | Fix fail_expr to create a real FailAction wrapper, not "fail_" name prefix | helpers.go:661-663 | Small |

### P1 — Required for correct abstraction refinement / reachability

| # | Task | Files | Effort |
|---|------|-------|--------|
| 7 | Implement CaseConjecture using tr.InterpolantCase() | helpers.go:334-362 | Medium |
| 8 | Implement ReachState model extraction (GetModelClauses, FindTrueDisjunct, ClausesModelToClauses, universe) | helpers.go:215-243 | Large |
| 9 | Port tagged_or_clauses to clauseops and use in JoinUnders | helpers.go:183-193, clauseops/ | Medium |
| 10 | Fix EvalStateAtom to handle RME expression nodes | eval.go:196-237 | Medium |
| 11 | Fix EvalAssertRhs to properly evaluate RME through ActionContext | phase4.go:280-311 | Medium |

### P2 — Correctness and fidelity improvements

| # | Task | Files | Effort |
|---|------|-------|--------|
| 12 | Verify concrete_post covers all compose_state_action edge cases (moded symbols, annotations) or port compose_state_action | eval.go:27-72, transrel/ | Large |
| 13 | Extend Diagram to pass implied, weaken, upward_close parameters | helpers.go:372-389, solver/ | Medium |
| 14 | Extend HistorySatisfy to pass _get_model_clauses and final_cond | helpers.go:421-424, transrel/ | Small |
| 15 | Port check_sequence pattern for FalseProperties | helpers.go:522-556, solver/ | Medium |
| 16 | Batch UndecidedConjectures solver calls (ClausesImplyList) | helpers.go:280-298 | Small |
| 17 | Add ReachStateFromPredNoAbduction alias | helpers.go | Trivial |
| 18 | Verify new_state delegation through ActionContext is not needed (or add it) | eval.go | Small |
| 19 | Fix ModuleTypeCheckConcepts arity vs sort semantics | helpers.go:473-512 | Small |

### P3 — Low priority / cosmetic

| # | Task | Files | Effort |
|---|------|-------|--------|
| 20 | Audit all checkPrecond parameter threading (replaces Python's global context.check) | all interp/ files | Medium |
| 21 | Add reporter callback to FalseProperties | helpers.go | Small |
| 22 | Remove stale "Many solver-dependent functions are stubbed" comment after fixes | interp.go:12-13 | Trivial |

---

## 6. Dependency Graph

```
Task 9 (tagged_or_clauses) ← Task 8 (ReachState model extraction)
Task 2,3 (FailAction methods) ← Task 6 (fail_expr)
Task 7 (CaseConjecture) depends on tr.InterpolantCase existing in transrel/
Task 8 (ReachState) depends on solver.GetModelClauses, FindTrueDisjunct, ClausesModelToClauses existing
Task 12 (compose_state_action) may subsume parts of Task 1 (State.Update)
```

---

## 7. Functions Verified as OK

The following functions have been audited and are faithfully ported (minor idiomatic Go differences only):

- `TypeCheckList` (phase4.go) — OK
- `ModuleTypeCheck` (helpers.go) — OK
- `ModuleNewState` / `NewStateFromClauses` (helpers.go / eval.go) — OK
- `ModuleNewStateWithValue` (helpers.go) — OK
- `ModuleOrder` (phase4.go) — OK (minor: Python passes `self.relations`, Go doesn't — verify ImpliesState handles this)
- `ModuleSkolemizer` (phase4.go) — OK
- `NewState` constructor (interp.go) — OK
- `State.Value()` / `SetValue()` (interp.go) — OK
- `State.Pred()` (interp.go) — OK
- `State.IsBottom()` (interp.go) — OK
- `State.ToFormula()` (interp.go) — OK
- `State.Conjs()` / `SetConjs()` / `Unders()` / `SetUnders()` (interp.go) — OK (uses Attributes map instead of direct fields, functionally equivalent)
- `EvalContext` struct (interp.go) — OK (parameter-passing replaces global context)
- `WrapState` / `UnwrapState` / `stateNode` (interp.go) — OK
- `IsActionApp` / `IsStateJoin` / `ActionApp` / `StateJoin` / `IsStateSymbol` / `StateEquation` (interp.go) — OK
- `StatesInExpr` / `StatesStateExpr` (interp.go / phase4.go) — OK
- `IvyActionFailedError` (interp.go) — OK
- `UnsatCoreWithInterpolant` (helpers.go) — OK
- `Reverse` (helpers.go) — OK
- `ReverseUpdateConcreteClauses` (helpers.go) — OK
- `AddUnder` (helpers.go) — OK
- `ReachStateFromPred` (helpers.go) — OK (recomputes pre/axioms independently, which is fine)
- `ConcreteJoin` (eval.go) — OK (Python has a latent bug here that Go avoids)
- `EvalAction` (eval.go) — OK
- `EvalState` (eval.go) — OK
- `BottomState` (eval.go) — OK
- `EvalStateFacts` (helpers.go) — OK
- `EvalStateActions` (helpers.go) — OK
- `TopAlpha` (helpers.go) — OK
- `GetCore` (phase4.go) — OK
- `GetPropertyContext` (helpers.go) — OK
- `StateImpliesFormula` (phase4.go) — OK
- `EvalStateOrder` (phase4.go) — OK
- `CheckStateAssertion` (phase4.go) — OK (different structure access but functionally equivalent)
- `GetStateAssertions` (phase4.go) — OK
- `UniverseConstraint` (phase4.go) — OK
- `UnderapproximateState` (phase4.go) — OK
- `DecomposeActionApp` (phase4.go) — OK (uses Go equivalents for Python constructs)
- `NewHistoryFromState` (helpers.go) — OK
- `HistoryForwardStep` (helpers.go) — OK
- `FilterConjectures` (helpers.go) — OK (uses individual solver calls instead of clauses_imply, functionally equivalent)
