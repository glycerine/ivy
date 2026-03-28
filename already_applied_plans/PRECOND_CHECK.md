Here's the complete analysis:

---

## How Python's `context.check` flows

Python has a single global `context = EvalContext(check=True)`. The `.check` field is read in exactly **one place**: `concrete_post()` line 202, which passes it to `compose_state_action(..., check=context.check)`. That function checks whether action preconditions are satisfiable and raises `ActionFailed` if check=True and the precondition is violated.

### All Python sites that set `check=False`:

| File | Line | Function | What it wraps |
|------|------|----------|---------------|
| `ivy_art.py` | 111 | `add_initial_state` | `eval_state(s)` |
| `ivy_ui.py` | 229 | `do_state_action` | `self.g.do_state_action(a, ...)` |
| `ivy_bmc.py` | 58 | BMC loop | `ag.execute(step_action)` |
| `ivy_check.py` | 604 | `check_properties` (init) | `AnalysisGraph(initializer=...)` |
| `ivy_check.py` | 641 | `check_properties` (action) | `ag.execute(action, pre)` |
| `ivy_check.py` | 706 | `check_properties` (temporal) | `ag.execute(action, prestate=pre)` |
| `ivy_trace.py` | 316 | `check_final_cond` | `ag.execute(action, pre)` |
| `ivy_trace.py` | 370 | `check_final_cond` | `ag.execute(action, pre)` |

Everything else inherits the default `check=True`.

---

## Go call sites — correctness verdict

### Correct as-is:

| Go call site | Value | Python equivalent | Verdict |
|---|---|---|---|
| `art.AddInitialState` (art.go:1411) | `false` | `ivy_art.py:111 EvalContext(check=False)` | **CORRECT** |
| `art.CheckSafety` (art.go:878) | `true` | `ivy_art.py:378 eval_state` (default context, check=True) | **CORRECT** |
| `art.Initialize` (art.go:1460) | `true` | `ivy_art.py:130 eval_state_facts` (default context, check=True) | **CORRECT** |
| `art.StateExtensions` (art.go:1146) | `true` | No EvalContext override (default check=True) | **CORRECT** |

### Needs parameterization:

**1. `art.DoStateAction` (art.go:1078) — hardcodes `true`, should be a parameter**

Python's `ivy_art.py:139 do_state_action` inherits the caller's context. The key caller is `ivy_ui.py:229` which wraps it in `EvalContext(check=False)`. So when called from the UI, check should be `false`.

The Go method should accept `checkPrecond bool` so the future UI port can pass `false`:
```go
func (ag *AnalysisGraph) DoStateAction(checkPrecond bool, equation *ast.Definition, ...) *State {
```

**2. `EvalAssertRhs` (phase4.go:296) — hardcodes `true`, should be a parameter**

Python's `eval_assert_rhs` doesn't create an `EvalContext` — it inherits whatever context is active. Its callers (`check_state_assertion`, `get_state_assertions`) are called from `check_safety` which uses the default check=True, so `true` happens to be correct for the current call paths. But for conformance, it should thread through from callers:

```go
func EvalAssertRhs(checkPrecond bool, rhs interface{}, domain *module.Module) (*State, error) {
```

And `CheckStateAssertion` / `GetStateAssertions` should also accept and forward it.

**In practice for phase4.go:297 specifically**: The hardcoded `true` is correct for all existing call paths. You do NOT need to change it to `false`. The Python equivalent also uses `true` (default context) for these paths. But making it a parameter is the correct long-term fix.

### Structural issue: `art.Execute` / `art.PostState` bypasses precondition checking entirely

This is a **bigger concern** than any individual `checkPrecond` value. Python's `ag.execute()` → `ag.post_state()` → `concrete_post()` → reads `context.check`. But Go's `art.Execute()` → `art.PostState()` computes forward images directly via `transrel.ForwardImage` — it **never calls `interp.ConcretePost`** and therefore **never checks preconditions**, regardless of any flag.

This means Go's `art.Execute()` effectively always behaves as `check=False`. This is coincidentally correct for the Python callers that set `check=False` (ivy_bmc, ivy_check, ivy_trace), but it will be **wrong** for any future caller that expects `check=True` behavior (i.e., precondition violation detection).

### Summary

- **phase4.go:297**: `true` is correct for existing callers. No change needed now, but parameterize for future-proofing.
- **art.DoStateAction**: `true` is wrong when called from UI path (Python uses `false`). Should be parameterized.
- **art.Execute/PostState**: Never checks preconditions at all — structural divergence from Python where `post_state` delegates to `concrete_post`.
- All other sites are correct.

