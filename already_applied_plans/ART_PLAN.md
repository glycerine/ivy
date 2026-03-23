# Plan: Port ivy_art.py to art/ — Audit Section 8 Fixes

## Context

The Go port of `~/pyivy/ivy/ivy/ivy_art.py` → `art/art.go` has 20 items flagged in AUDIT18MARCH.md section 8: 10 missing functions, 3 stubs, and 7 behavioral differences. Deep comparison confirms all items are still unfixed. The good news: nearly all required supporting infrastructure already exists in the Go codebase (`interp.ConcretePost`, `interp.EvalState`, `interp.HistorySatisfy`, `interp.ModuleOrder`, `interp.CheckStateAssertion`, `interp.GetCore`, `actions.Sequence`, `webui.CyElements`, etc.).

The work is grouped into 4 batches by complexity and dependency order.

---

## Batch 1: Critical Bug + Trivial Fixes (Small, do first)

These are correctness bugs or trivial additions with no new dependencies.

### 1A. Fix `Delete()` bug (art.go:617-629) — CRITICAL
- **Bug:** `state.ID = -1` on line 618, then `state.ID + 1` on line 619 starts loop at 0 instead of original index.
- **Python (line 291-298):** Saves `id = state.id` before setting `state.id = -1`.
- **Fix:** Save `savedID := state.ID` before mutating; use `savedID + 1` as loop start.
- **Note:** Python line 296 has its own bug (`t.id` instead of `s.id` in generator), but the Go version correctly checks `d.ID == -1` — just needs the loop index fix.

### 1B. Fix `Add()` missing assertions (art.go:301-339)
- **Python (line 250-264):** Asserts `len(expr.args) == 1`, `isinstance(expr.args[0], State)`, and for string rep asserts `expr.rep in self.actions`.
- **Go:** Silently continues if action not found (line 317).
- **Fix:** Add panics or log.Fatal for these assertion failures, matching Python behavior.

### 1C. Add `decompose_edge()` (MISSING)
- **Python (line 408-410):** Trivial 3-line wrapper: unpacks transition, calls `self.decompose_state(poststate)`.
- **Fix:** Add `DecomposeEdge(t Transition) *AnalysisGraph` method that calls `ag.DecomposeState(t.Post)`.

### 1D. Add `make_concrete_trace()` (MISSING)
- **Python (line 404-406):** Literally a TODO stub (`return`).
- **Fix:** Add `MakeConcreteTrace(state *State, conc interface{})` as a stub with TODO comment, matching Python.

### 1E. Add `show_core()` (MISSING)
- **Python (line 239-243):** Parses clause string, calls `state.domain.get_core(state, clause)`, prints result.
- **Go has:** `interp.GetCore(state, clause)` at `interp/phase4.go:79`.
- **Fix:** Add `ShowCore(clauseStr string, state *State)` that parses clause and calls GetCore.

### 1F. Fix `FixedpointCandidate()` stub (art.go:967-975)
- **Python (line 426-432):** Groups uncovered states by label, then joins each group and returns `defaultdict(bottom_state, ...)`.
- **Go:** Returns raw `map[string][]*State` without joining.
- **Fix:** After grouping, call `ag.JoinStates()` on each group; return `map[string]*State` with bottom-state default.

**Files:** `art/art.go`
**Estimated scope:** ~60 lines changed/added

---

## Batch 2: Behavioral Fixes Requiring interp Infrastructure (Medium)

These fix divergences where the Go code reimplemented logic instead of calling existing interp functions.

### 2A. Fix `PostState()` to use `ConcretePost` (art.go:388-433)
- **Python (line 158-163):** `s = concrete_post(op.update(pre_state.domain, pre_state.in_scope), pre_state)`.
- **Go:** Does manual `And(preFmla, trNode)` instead of calling `interp.ConcretePost`.
- **Fix:** Replace manual composition with call to `interp.ConcretePost(update, state, nil)`. Need to convert between `art.State` and `interp.State` — may need adapter or shared interface. Alternatively, call `transrel.ForwardImage()` directly (at `transrel/transrel.go:1027`), which is what ConcretePost ultimately uses.
- **Key function:** `transrel.ForwardImage(pre, axioms, update)` — computes proper forward image with havocing.

### 2B. Fix `Cover()` to use `ModuleOrder` (art.go:471-512)
- **Python (line 217-218):** `covered_node.domain.order(covered_node, covering_node)` — uses module's ordering.
- **Go:** Uses `z3bridge.Implies` directly.
- **Go has:** `interp.ModuleOrder(state1, state2)` at `interp/phase4.go:46`.
- **Fix:** Replace Z3 Implies call with `interp.ModuleOrder`. Requires state type compatibility check between `art.State` and `interp.State`.

### 2C. Fix `AddInitialState()` to use Sequence (art.go:1061-1104)
- **Python (line 102-119):** Creates `Sequence(*initializers)`, wraps in `env_action`, evaluates as single `action_app` with `EvalContext(check=False)`.
- **Go:** Loops and executes each initializer separately, creating intermediate states.
- **Go has:** `actions.NewSequence()` at `actions/action.go:232`, and `EvalContext` at `interp/interp.go:250`.
- **Fix:**
  1. Compose initializers into single `actions.NewSequence(...)`.
  2. Wrap with `EnvAction` (use `bmc.EnvAction` pattern or `temporal.EnvAction`).
  3. Create `ActionApp(action, s)`, evaluate with `interp.EvalState` under `AC(no_add=True)` + `EvalContext{Check: false}`.
  4. Add `ic` and `abstractor` parameters matching Python signature.

### 2D. Fix `CheckSafety()` missing eval + error handling (art.go:811-845)
- **Python (line 372-383):**
  1. Calls `check_state_assertion(state, asn)` for each assertion.
  2. If state has expr, evaluates `eval_state(state.expr)` under `AC(no_add=True)`.
  3. Catches `IvyActionFailedError`, returns Counterexample.
- **Go:** Does Z3 implication check instead of `CheckStateAssertion`, no expr evaluation, no error handling.
- **Go has:** `interp.CheckStateAssertion` at `interp/phase4.go:349`.
- **Fix:** Replace Z3 Implies with `interp.CheckStateAssertion`; add expr evaluation block with error recovery.

### 2E. Fix `BMC()` to use HistorySatisfy (art.go:779-805)
- **Python (line 331-345):** Calls `history_satisfy(history, state, _get_model_clauses, final_cond=error_cond)`, extracts `(universe, path)`, assigns values to states.
- **Go:** Just checks `IsSat` — no path extraction, no model values.
- **Go has:** `interp.HistorySatisfy` at `interp/helpers.go:394`, `transrel.SatisfyResult` type.
- **Fix:** Replace IsSat with `interp.HistorySatisfy`; extract path and universe from result; zip with states.

**Files:** `art/art.go`
**Estimated scope:** ~150 lines changed

---

## Batch 3: Missing Methods (Medium, some depend on Batch 2)

### 3A. Add `StateActions(state)` (MISSING)
- **Python (line 134-137):**
  ```python
  if state.label != None:
      return [state_equation(post, expr) for post, e in self.predicates.items()
              for expr in eval_state_actions(e, state)]
  return [state_equation(None, action_app(a, state)) for a in self.actions
          if a in self.public_actions]
  ```
- **Go has:** `interp.EvalStateActions` at `interp/helpers.go:609`, `interp.StateEquation` at `interp/interp.go:322`.
- **Fix:** Add `StateActions(state *State) []*ast.Definition` method.

### 3B. Add `DoStateAction(equation, abstractor)` (MISSING)
- **Python (line 139-145):** Evaluates equation under AC context, assigns label.
- **Depends on:** Batch 2 (EvalState integration).
- **Fix:** Add `DoStateAction(equation *ast.Definition, abstractor Abstractor) *State`.

### 3C. Add `RecalculateState(state, abstractor)` (MISSING)
- **Python (line 176-184):** If state has pred, calls `concrete_post(state.update, state.pred)`. If has join_of, calls `join_states`.
- **Depends on:** Batch 2A (PostState using ConcretePost).
- **Fix:** Add `RecalculateState(state *State, abstractor Abstractor)`.

### 3D. Add `StateExtensions(state)` (MISSING)
- **Python (line 434-440):** Calls `state_actions`, gets `fixedpoint_candidate`, yields equations not yet covered.
- **Depends on:** 3A (StateActions), 1F (FixedpointCandidate fix).
- **Go has:** `interp.EvalStateOrder` at `interp/phase4.go:317`.
- **Fix:** Add `StateExtensions(state *State) []*ast.Definition` method.

### 3E. Fix `Initialize()` to match Python (art.go:1111-1116)
- **Python (line 121-132):** If predicates exist, evaluates each as state facts with label. Otherwise calls `add_initial_state`.
- **Go:** Just calls `AddInitialState()` + optional callback.
- **Go has:** `interp.EvalStateFacts` at `interp/helpers.go:563`.
- **Fix:** Add predicate-based initialization path.

**Files:** `art/art.go`
**Estimated scope:** ~100 lines added

---

## Batch 4: GUI/Visualization Methods (IMPLEMENT now, do not postpone).

### 4A. Add `ConceptGraph()` (MISSING)
- **Python (line 307-316):** Creates concept graph using `standard_graph(state)` callback and background theory.
- This is GUI-related and depends on concept graph infrastructure in `webui/`.
- **Fix:** Add method that integrates with `webui.RenderConceptGraph` pattern.

### 4B. Add `AsCyElements()` (MISSING)
- **Python (line 442-443):** Wraps `render_rg(self)` with `dot_layout`.
- **Go has:** `RenderRg` in `art/phase7.go`, `webui.CyElements` and `webui.RenderARG`.
- **Fix:** Add `AsCyElements(dotLayout func(*webui.CyElements) *webui.CyElements) *webui.CyElements`.

### 4C. CheckConstraints and StratifyGoals stubs
- **Not in Python** — these can remain as stubs. Add comments about their utility or lack thereof. Why were they stubbed in the first place?

**Files:** `art/art.go`, possibly `art/phase7.go`
**Estimated scope:** ~40 lines

---

## State Type Compatibility Note

A key challenge for Batches 2-3: `art.State` and `interp.State` are different types. The interp functions (`ConcretePost`, `EvalState`, `ModuleOrder`, `CheckStateAssertion`, etc.) operate on `interp.State`. Options:
1. **Adapter functions** — Convert `art.State` ↔ `interp.State` at call boundaries.
2. **Shared interface** — But this violates "no abstractions not in Python" rule.
3. **Unify types** — Make `art.State` == `interp.State` (most faithful to Python where there's one State class).

Option 1 (adapter) is safest for now; option 3 is the most correct long-term.

---

## Verification Plan

After each batch:
1. `go build ./art/...` — compiles
2. `go test ./art/...` — existing tests pass
3. `go vet ./art/...` — no warnings

After all batches:
4. `go test ./...` — full test suite passes
5. For Batch 2 fixes, verify with end2end tests: `go test ./end2end/...`
6. Compare key method outputs (PostState, BMC, CheckSafety) against Python by running equivalent scenarios

---

## Summary Table

| Batch | Items | Complexity | Dependencies |
|-------|-------|-----------|--------------|
| 1 | 1A-1F (6 items) | Small | None |
| 2 | 2A-2E (5 items) | Medium | interp package integration |
| 3 | 3A-3E (5 items) | Medium | Batch 2 |
| 4 | 4A-4C (3 items) | Lower | webui package |
| **Total** | **19 items** | | |
