# Plan: Unify FailAction in actions/, port lazy fail-state update flow

Created: 2026-04-10 (afternoon, plan-mode session, revised after first draft)

## Context

`make golden` (TestOrdLive) is failing at line 233846 of
`/Users/jaten/ivy/goivy/log.red`:

```
233846  go : XTRACE: check.checkFcsNormalPath history path postFmlas=0 postDefs=0 axiomFmlas=20 axiomDefs=12 checkers=1
        py : XTRACE: interp.State.Update calling GetUpdate type=fail_action
```

Python emits an entire cascade of xtraces driven by the lazy
`State.update` property on a synthetic "fail" state, while Go skips that
work and goes straight to `checkFcsNormalPath` with an empty post (note
`postFmlas=0`). The Go path is wrong and is silently producing the right
answer only because the empty post makes every assertion vacuously
satisfied.

### Python flow (the source of truth)

In `~/ivy/pyivy/ivy/ivy/ivy_check.py:720-746`, the guarantee loop builds
a `fail` state and asks `check_safety_in_state` to verify it:

```python
xtracer.trace("check.guarantee_loop pre BuildEnvAction")
action = act.env_action(root)                         # → EnvAction
ag = ivy_art.AnalysisGraph()
pre = itp.State(); pre.clauses = get_conjs(mod)
with itp.EvalContext(check=False):
    post = ag.execute(action, prestate=pre)
fail = itp.State(expr = itp.fail_expr(post.expr))     # ← fresh State, lazy update
if not check_safety_in_state(mod, ag, fail, report_pass=False):
    ...
```

`fail_expr(expr) = action_app(fail_action(expr.rep), expr.args[0])`
(`ivy_interp.py:406`). So `fail.expr.rep` is a `fail_action(EnvAction)`
object and `fail.expr.args[0]` is `pre`.

`check_safety_in_state` → `check_fcs_in_state` (`ivy_check.py:377`)
calls `history = ag.get_history(post)` which recurses on `state.pred`
calling `history_forward_step` (`ivy_interp.py:597-599`), which accesses
`state.update`. The lazy property at `ivy_interp.py:137-142`:

```python
@property
def update(self):
    if self.cached_update is None and self.expr is not None and is_action_app(self.expr):
        if __debug__: xtracer.trace("interp.State.Update calling GetUpdate type=%s"
                                    % type(eval_action(self.expr.rep)).__name__)
        self.cached_update = eval_action(self.expr.rep).update(self.domain, self.in_scope)
    return self.cached_update
```

`eval_action(fail.expr.rep)` returns the `fail_action` instance directly
(it is an `Action` subclass), and `.update(...)` invokes the override at
`ivy_interp.py:385-389`:

```python
class fail_action(Action):
    def update(self,domain,in_scope):
        if __debug__: xtracer.trace("interp.ActionFail calling GetUpdate type=%s" % type(self.action).__name__)
        upd = self.action.update(domain,in_scope)   # base Action.update — full cascade
        return action_failure(upd)
    def int_update(self,domain,in_scope):
        if __debug__: xtracer.trace("interp.ActionFail calling IntUpdate type=%s" % type(self.action).__name__)
        return action_failure(self.action.int_update(domain,in_scope))
```

The override **bypasses** `Action.update`'s base trace
(`"actions.GetUpdate ENTER type=%s"`) and emits its own
`"interp.ActionFail calling GetUpdate type=<inner>"` trace, then calls
`self.action.update(...)` which DOES emit `actions.GetUpdate ENTER` for
the inner action and runs the full inner cascade.

Note: in Python, both `ivy_check.py` (safety check on the synthetic
fail state) **and** `ivy_trace.py:230` (`self.last_action =
itp.fail_action(self.last_action)` inside the trace builder) use the
**same single `fail_action` class** from `ivy_interp.py`.

### Go state today

1. **Two duplicate `FailAction` types exist** — both are buggy ports of
   the single Python `fail_action` class:

   - `interp.FailAction` at
     `/Users/jaten/ivy/goivy/interp/helpers.go:48-121` — has the right
     embedding (`actions.ActionBase`), `Inner` field, `Name() == "fail"`,
     `Sexp()`, etc., **but is missing `Update` and `IntUpdate` methods
     entirely**.
   - `trace.FailAction` at
     `/Users/jaten/ivy/goivy/trace/trace.go:26-68` — a parallel
     duplicate that uses `Action` (not `Inner`) for the wrapped action,
     uses `"FAIL(...)"` for `String()` (Python emits `"fail ..."`
     lowercase), and has its own broken tests
     (`trace/trace_test.go:182-213`). It is constructed at
     `trace/trace.go:349` (`tb.LastAction = &FailAction{Action:
     tb.LastAction}`) corresponding to Python `ivy_trace.py:230`
     (`self.last_action = itp.fail_action(self.last_action)`).

   This duplication is a previous-Claude-session porting mistake. Python
   has exactly one `fail_action` class.

2. **`actions.IntUpdate` cannot dispatch to `interp.FailAction`** because
   `actions` cannot import `interp` (cycle). Today the `IntUpdate` type
   switch (`actions/update.go:1157-1219`) silently falls through to the
   `default` branch and returns `NullUpdate()` for any FailAction — but
   in practice nothing ever calls `actions.IntUpdate` on a FailAction
   because `interp.NewFailAction` is only invoked from a dead branch in
   `interp/eval.go:115-122`.

3. **`interp.State.Update()` is a stub** at
   `/Users/jaten/ivy/goivy/interp/interp.go:142-149`:
   ```go
   func (s *State) Update() *actions.Update {
       if s.CachedUpdate == nil && s.Expr != nil && IsActionApp(s.Expr) {
           // Matches Python: s.update = eval_action(s.expr.rep).int_update(s.domain, s.in_scope)
           // Action update computation requires the actions package's IntUpdate method.
           // The update is cached on first access.
       }
       return s.CachedUpdate
   }
   ```
   Returns `nil` whenever `CachedUpdate` is nil, regardless of `Expr`.

4. **`interp.FailExpr`** at
   `/Users/jaten/ivy/goivy/interp/helpers.go:651-654` builds an
   `*ast.Atom` whose `Rep` is the **string** `"fail_<name>"`. This loses
   the wrapped action object — `ast.Atom.Rep` is a Go `string`. The
   current Go code uses that string-prefix purely as a display label and
   never reconstructs a FailAction from it.

5. **The fail-state construction in `check/isolate_check.go`** never
   sets `Pred`, `Update`, or `Action`:

   - Lines 569-583 (guarantee loop):
     ```go
     if post != nil {
         failState := art.NewState(mod, module.TrueClauses(actions.EmptyAnnotation{}))
         if aa, ok := post.Prov.(*art.ActionApp); ok {
             if rep, ok := aa.Rep.(string); ok {
                 failState.Prov = art.NewActionApp("fail_"+rep, aa.Args...)
             }
         }
         if !CheckSafetyInStateWithAG(mod, ag, failState, false) {
             someFailed = true
             break
         }
     }
     ```
   - Lines 316-328 (initializer guarantees) — same shape, but
     `ag.States[0].Prov.Rep` is the **EnvAction object** (not a string),
     so the existing `if rep, ok := aa.Rep.(string)` branch silently
     does nothing on this path; the printed label is wrong too.

6. **`art.AnalysisGraph.GetHistory`** at
   `/Users/jaten/ivy/goivy/art/art.go:749-786` reads `state.Update` as a
   **field** and short-circuits when `state.Pred == nil`:

   ```go
   if state.Pred == nil || (bound != nil && *bound <= 0) {
       u := actions.PureStateClauses(state.Clauses)
       return actions.NewHistory(ag.Domain.Cfg.IuCfg, u)
   }
   ...
   if state.Update != nil {
       h = h.ForwardStep(axioms, state.Update, actionNode)
   }
   ```

   No lazy compute. If both `state.Pred` and `state.Update` are unset
   (the failState case), `GetHistory` returns the empty pure state and
   never invokes `ForwardStep`.

7. **`actions.GetUpdate(action, ctx)`** at
   `/Users/jaten/ivy/goivy/actions/update.go:2139-2145` always emits
   `"actions.GetUpdate ENTER type=%s"`. Python's `fail_action.update`
   override does NOT emit the matching `actions.GetUpdate ENTER
   type=fail_action`, so the Go dispatch must special-case FailAction to
   bypass that trace.

8. **`actions.ActionFailure(u *Update) *Update`** is already implemented
   at `/Users/jaten/ivy/goivy/actions/transrel.go:1706-1712` — the
   direct counterpart of Python's `action_failure(action)`.

### Why the divergence is at line 233846

Up through line 233845 the two engines match because everything that
matters runs inside `ag.Execute(envAction, ...)` (and its descendants),
which is correctly ported. The divergence is the **first xtrace after
`ag.Execute` returns**: Python's lazy `fail.update` access fires inside
`check_fcs_in_state(mod, ag, fail, ...)`, and Go's GetHistory
short-circuits on the un-wired fail state and falls straight into
`checkFcsNormalPath` with an empty post. Fixing only the missing trace
text is not enough — we also need Go to actually compute the
fail-action update so the entire downstream cascade fires
(`interp.ActionFail`, `actions.GetUpdate ENTER type=EnvAction`, the
inner action's `IntUpdate`/`Sequence`/`ComposeUpdates`/... cascade,
`forward_image`, `transrel.SatisfyWithCond ENTER`, and finally
`checkFcsNormalPath` with non-zero `postFmlas`).

## Approach

The cleanest mechanical port is:

1. **Move `FailAction` from `interp/` to `actions/`** so the dispatch
   in `actions.GetUpdate` / `actions.IntUpdate` can know about it
   without an import cycle. Python has exactly one `fail_action` class
   and it lives in `ivy_interp.py` only because Python's flat module
   namespace makes location irrelevant; Go's strict layered package
   structure forces any `Action` subclass that needs dispatch into the
   `actions` package. This is a Go-specific accommodation, NOT a
   refactor of Python.

2. **Eliminate `trace.FailAction`** entirely and rewire `trace.TraceBase.Fail`
   to use `actions.NewFailAction`. Python's `ivy_trace.py:230` uses
   `itp.fail_action(self.last_action)` — the **same** class as the
   safety check. Go's separate `trace.FailAction` is a previous
   porting mistake. Removing it eliminates ~50 lines of duplicated
   stubs and ~30 lines of misleading tests.

3. **Add `Update` and `IntUpdate` methods on `*actions.FailAction`**
   matching Python's `fail_action.update` / `fail_action.int_update`,
   emitting the matching xtraces and recursively delegating through
   `actions.GetUpdate` / `actions.IntUpdate` on the inner action.

4. **Add explicit dispatch** in `actions.GetUpdate`,
   `actions.IntUpdate`, and `actions.ActionTypeName` for `*FailAction`:
   - `actions.GetUpdate`: at the top, before emitting `"actions.GetUpdate
     ENTER"`, check for `*FailAction` and delegate to `fa.Update(ctx)`
     directly. Skipping the ENTER trace matches Python's
     `fail_action.update` override (which does NOT call
     `super().update`).
   - `actions.IntUpdate`: add a `case *FailAction: return a.IntUpdate(ctx)`
     in the existing type switch.
   - `actions.ActionTypeName`: add `case *FailAction: return
     "fail_action"` so the xtrace text matches Python's
     `type(fa).__name__` exactly.

5. **Make `interp.State.Update()` a real lazy property** that emits the
   `"interp.State.Update calling GetUpdate type=<class>"` trace and
   computes via `actions.GetUpdate(action, ctx)`. The action source for
   the failState case is `s.Action` (a new write at construction time),
   not `s.Expr.Rep` (which is a string and cannot carry a FailAction
   object). For non-fail states the lazy path is harmless because they
   already have `CachedUpdate` set by `ConcretePost`.

6. **Make `art.AnalysisGraph.GetHistory` lazily compute `state.Update`**
   when it is nil but `state.Action != nil`. This is the Go-equivalent
   of Python's `history_forward_step` accessing the lazy `state.update`
   property. Putting it here is consistent with Go's split between
   `art.State` (graph state) and `interp.State` (interpreter state) —
   `GetHistory` is the only consumer that walks back through state and
   needs the per-state update.

7. **Wire the failState in `check/isolate_check.go`** at lines 569-583
   (guarantee loop) and 316-328 (initializer guarantees) by:
   - Setting `failState.Action = actions.NewFailAction(<inner>)` so the
     lazy `GetHistory` compute knows what to call.
   - Setting `failState.Pred = post.Pred` (with the `aa.Args[0]`
     fallback) so `GetHistory` walks back through the predecessor
     chain.
   - Leaving `failState.Update` nil so the lazy path fires (matching
     Python's `fail.cached_update is None` initial state).

8. **Clean up dead branch** in `interp/eval.go:115-122` (the
   `if fa, ok := expr.(*FailAction); ok` branch) — Go's first
   `expr.(actions.Action)` check already catches `*FailAction` (because
   it implements the Action interface), so the second branch is
   unreachable. Match Python's `eval_action` (`ivy_interp.py:438-448`)
   which has the equivalent dead branch — but Go can be cleaner since
   we control the type system. Optionally leave the branch for literal
   parity with Python; the choice does not affect behavior.

### Step 1 — Create `actions/fail_action.go`

**New file**: `/Users/jaten/ivy/goivy/actions/fail_action.go`

Move the entire `FailAction` definition from
`/Users/jaten/ivy/goivy/interp/helpers.go:44-121` into this new file,
adjusting:

- Package: `package actions` (not `package interp`).
- Imports: drop the `actions` prefix on `actions.ActionBase`,
  `actions.Action`, etc. since we are now inside the `actions` package.
- Add `xtracer` import (not currently in the moved snippet).

```go
// Package actions: fail_action implementation, ported from Python's
// fail_action class in ivy_interp.py:378-404.
//
// Lives in actions/ (not interp/) so that actions.GetUpdate /
// actions.IntUpdate / actions.ActionTypeName can dispatch to it without
// creating an import cycle (interp imports actions). This is a
// Go-specific layering accommodation; Python keeps fail_action in
// ivy_interp.py because its flat namespace makes file location
// irrelevant.

package actions

import (
    "fmt"

    "github.com/glycerine/ivy/goivy/ast"
    iu "github.com/glycerine/ivy/goivy/ivyutils"
    lg "github.com/glycerine/ivy/goivy/logic"
    "github.com/glycerine/ivy/goivy/xtracer"
)

// FailAction wraps an inner action so that its update is transformed
// via action_failure. Mechanical port of Python's fail_action class
// (ivy_interp.py:378-404).
type FailAction struct {
    ActionBase
    Inner Action
}

// NewFailAction creates a FailAction wrapping the given action.
// Python: fail_action(action) — see ivy_interp.py:379-382.
func NewFailAction(inner Action) *FailAction {
    fa := &FailAction{Inner: inner}
    if inner != nil {
        fa.ActionBase.SetLineno(inner.GetLineno())
    }
    return fa
}

func (fa *FailAction) Name() string { return "fail" }

// String matches Python fail_action.__str__ (ivy_interp.py:383-384):
//
//     def __str__(self):
//         return "fail " + (self.action.label if hasattr(self.action,'label') else str(self.action))
//
// We do not have a separate label slot, so always use the inner String.
func (fa *FailAction) String() string {
    if fa.Inner == nil {
        return "fail"
    }
    return "fail " + fa.Inner.String()
}

func (fa *FailAction) ActionArgs() []lg.Expr {
    if fa.Inner == nil {
        return nil
    }
    return fa.Inner.ActionArgs()
}

func (fa *FailAction) ActionClone(args []lg.Expr) Action {
    if fa.Inner == nil {
        return &FailAction{ActionBase: fa.ActionBase}
    }
    return &FailAction{
        ActionBase: fa.ActionBase,
        Inner:      fa.Inner.ActionClone(args),
    }
}

func (fa *FailAction) IterCalls() []string {
    if fa.Inner == nil {
        return nil
    }
    return fa.Inner.IterCalls()
}

func (fa *FailAction) IterSubactions() []Action { return []Action{fa} }

func (fa *FailAction) Decompose() [][]Action { return [][]Action{{fa}} }

// FailedAction returns the wrapped inner action. Mechanical port of
// Python fail_action.failed_action (ivy_interp.py:403-404).
func (fa *FailAction) FailedAction() Action { return fa.Inner }

// --- ast.Node + lg.Expr interface methods ---

func (fa *FailAction) Args() []ast.Node {
    args := fa.ActionArgs()
    nodes := make([]ast.Node, len(args))
    for i, e := range args {
        nodes[i] = e
    }
    return nodes
}

func (fa *FailAction) Clone(args []ast.Node) ast.Node {
    exprs := make([]lg.Expr, len(args))
    for i, n := range args {
        exprs[i] = n.(lg.Expr)
    }
    return fa.ActionClone(exprs).(ast.Node)
}

func (fa *FailAction) Children() []lg.Expr          { return fa.ActionArgs() }
func (fa *FailAction) NodeSort() lg.Sort            { return lg.ActionS }
func (fa *FailAction) Equal(other lg.Expr) bool     { return fa.Sexp() == other.Sexp() }
func (fa *FailAction) GetAstConfig() *ast.AstConfig { return nil }

func (fa *FailAction) Sexp() lg.NodeKey {
    if fa.Inner == nil {
        return lg.NodeKey("(FailAction inner:nil)")
    }
    return lg.NodeKey(fmt.Sprintf("(FailAction inner:%v)", fa.Inner.Sexp()))
}

func (fa *FailAction) Canon() iu.Canonical { return iu.Canonical(fa.Sexp()) }

// Update matches Python fail_action.update (ivy_interp.py:385-389):
//
//     def update(self,domain,in_scope):
//         xtracer.trace("interp.ActionFail calling GetUpdate type=%s" % type(self.action).__name__)
//         upd = self.action.update(domain,in_scope)
//         return action_failure(upd)
//
// IMPORTANT: Python's fail_action.update OVERRIDES Action.update and
// therefore bypasses the base "actions.GetUpdate ENTER type=fail_action"
// trace. Go matches that by special-casing FailAction in GetUpdate
// below, so a call to GetUpdate(fa, ctx) ends up here without emitting
// the spurious ENTER line.
func (fa *FailAction) Update(ctx *UpdateContext) *Update {
    xtracer.Trace("interp.ActionFail calling GetUpdate type=%s", ActionTypeName(fa.Inner))
    upd := GetUpdate(fa.Inner, ctx)
    return ActionFailure(upd)
}

// IntUpdate matches Python fail_action.int_update (ivy_interp.py:390-393):
//
//     def int_update(self,domain,in_scope):
//         xtracer.trace("interp.ActionFail calling IntUpdate type=%s" % type(self.action).__name__)
//         return action_failure(self.action.int_update(domain,in_scope))
func (fa *FailAction) IntUpdate(ctx *UpdateContext) *Update {
    xtracer.Trace("interp.ActionFail calling IntUpdate type=%s", ActionTypeName(fa.Inner))
    return ActionFailure(IntUpdate(fa.Inner, ctx))
}
```

### Step 2 — Update `actions.GetUpdate`, `actions.IntUpdate`, `actions.ActionTypeName`

**File**: `/Users/jaten/ivy/goivy/actions/update.go`

#### 2a. Add `*FailAction` case to `ActionTypeName` (line 1085)

```go
func ActionTypeName(a interface{}) string {
    switch a.(type) {
    case *Sequence:
        return "Sequence"
    // ... existing cases ...
    case *CopyFieldAction:
        return "CopyFieldAction"
    case *FailAction:
        return "fail_action"   // Python class name is snake_case
    default:
        return fmt.Sprintf("%T", a)
    }
}
```

The `"fail_action"` (snake_case) string is what Python's
`type(fa).__name__` returns (the Python class is literally
`class fail_action(Action)` at `ivy_interp.py:378`). All other actions
return PascalCase because Python's other action classes use PascalCase.

#### 2b. Special-case `*FailAction` at the top of `GetUpdate` (line 2139)

```go
func GetUpdate(action Action, ctx *UpdateContext) *Update {
    // FailAction overrides Python's Action.update (ivy_interp.py:385-389),
    // so it does NOT emit the base "actions.GetUpdate ENTER type=fail_action"
    // trace. Match that by delegating to fa.Update directly here, before
    // the ENTER trace fires.
    if fa, ok := action.(*FailAction); ok {
        return fa.Update(ctx)
    }
    xtracer.Trace("actions.GetUpdate ENTER type=%s", ActionTypeName(action))
    update := IntUpdate(action, ctx)
    update = BindOldsUpdate(update)
    update = hideFormals(action, update)
    return update
}
```

#### 2c. Add `*FailAction` case to `IntUpdate` (line 1151 type switch)

```go
func IntUpdate(action Action, ctx *UpdateContext) *Update {
    switch a := action.(type) {
    // ... existing cases ...
    case *CrashAction:
        xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
        return intUpdateFromActionUpdate(a, ctx)
    case *FailAction:
        // FailAction.IntUpdate emits its own "interp.ActionFail calling
        // IntUpdate type=<inner>" trace and recursively calls IntUpdate
        // on the inner action, matching Python fail_action.int_update.
        return a.IntUpdate(ctx)
    default:
        // ... existing default ...
    }
}
```

### Step 3 — Delete the old `interp.FailAction` and clean up references

**File**: `/Users/jaten/ivy/goivy/interp/helpers.go`

Delete the entire FailAction block at lines 44-121 (the type
declaration through `FailedAction()`). Drop the now-unused
`"fmt"` reference if no other helper in this file uses it.

**File**: `/Users/jaten/ivy/goivy/interp/eval.go`

Replace the dead `*FailAction` branch at lines 115-122:

```go
// BEFORE:
if a, ok := expr.(actions.Action); ok {
    return a, nil
}
// If it's a FailAction, recursively evaluate.
if fa, ok := expr.(*FailAction); ok {
    inner, err := EvalAction(fa.Inner, mod)
    if err != nil {
        return nil, err
    }
    return NewFailAction(inner), nil
}
// AFTER:
if a, ok := expr.(actions.Action); ok {
    return a, nil
}
// Note: Python ivy_interp.py:441-442 has a parallel `isinstance(expr,
// fail_action)` branch that is dead code (fail_action IS an Action so
// the previous isinstance check fires first). Go's actions.Action type
// assertion catches *actions.FailAction the same way, so we omit the
// dead branch.
```

**File**: `/Users/jaten/ivy/goivy/interp/interp_test.go`

Move the FailAction tests at lines 608-666 and 980-995 to a new file
`/Users/jaten/ivy/goivy/actions/fail_action_test.go`, rewriting them
to use the actions-package types:

- `NewFailAction(inner)` (no `interp.` prefix; tests are in `actions`)
- `*FailAction` type assertions
- Drop `module.New()` references in tests that don't actually need a
  module — the inner action can be a bare `actions.NewSequence()`.

The original test bodies are tiny (most are 3-5 lines) so the move is
mostly cosmetic. After the move, run `go test ./interp/... ./actions/...
-run FailAction -v` to confirm.

### Step 4 — Eliminate `trace.FailAction`

**File**: `/Users/jaten/ivy/goivy/trace/trace.go`

Delete the entire FailAction block at lines 26-68.

Replace line 349:
```go
// BEFORE:
func (tb *TraceBase) Fail() {
    tb.LastAction = &FailAction{Action: tb.LastAction}
}

// AFTER:
// Fail marks the last action as failed.
// Python: ivy_trace.py:229-230:
//     def fail(self):
//         self.last_action = itp.fail_action(self.last_action)
func (tb *TraceBase) Fail() {
    tb.LastAction = actions.NewFailAction(tb.LastAction)
}
```

The `actions` import is already present in `trace/trace.go:15`, so no
new import is needed.

**File**: `/Users/jaten/ivy/goivy/trace/trace_test.go`

- Update `TestTraceBaseFail` (lines 147-159): change `*FailAction` to
  `*actions.FailAction` and `.Action` field references to `.Inner`.
- Delete the four trace-local FailAction tests at lines 182-213
  (`TestFailActionString`, `TestFailActionNilAction`,
  `TestFailActionClone`, `TestFailActionName`). Their behavior is
  already covered by the new `actions/fail_action_test.go` from Step 3,
  and the `"FAIL(...)"` capitalization tested here is wrong vs Python.

### Step 5 — Make `interp.State.Update()` a real lazy property

**File**: `/Users/jaten/ivy/goivy/interp/interp.go` lines 137-149

Replace the stub with a real lazy implementation:

```go
// Update returns the cached update, computing it lazily on first access.
//
// Mechanical port of Python ivy_interp.py:137-142:
//
//     @property
//     def update(self):
//         if self.cached_update is None and self.expr is not None and is_action_app(self.expr):
//             xtracer.trace("interp.State.Update calling GetUpdate type=%s"
//                           % type(eval_action(self.expr.rep)).__name__)
//             self.cached_update = eval_action(self.expr.rep).update(self.domain, self.in_scope)
//         return self.cached_update
//
// Source-of-action lookup differs from Python because ast.Atom.Rep is
// string-only in Go. We use s.Action when set (this is how the
// failState path passes a FailAction object through), and otherwise
// fall back to looking up s.Expr.Rep as an action name in the module.
func (s *State) Update() *actions.Update {
    if s.CachedUpdate != nil {
        return s.CachedUpdate
    }
    var action actions.Action
    if s.Action != nil {
        action = s.Action
    } else if s.Expr != nil && IsActionApp(s.Expr) {
        if atom, ok := s.Expr.(*ast.Atom); ok && s.Domain != nil {
            if a, found := s.Domain.FindAction(atom.Rep); found {
                if act, ok := a.(actions.Action); ok {
                    action = act
                }
            }
        }
    }
    if action == nil {
        return nil
    }

    xtracer.Trace("interp.State.Update calling GetUpdate type=%s", actions.ActionTypeName(action))

    ctx := &actions.UpdateContext{
        Domain:       s.Domain,
        PVars:        s.InScope,
        ActCfg:       s.Domain.Cfg.ActCfg,
        Instantiator: s.Domain.Instantiator,
        GetAction: func(name string) actions.Action {
            if v, ok := s.Domain.Actions.Get2(name); ok {
                if act, ok := v.(actions.Action); ok {
                    return act
                }
            }
            return nil
        },
    }
    s.CachedUpdate = actions.GetUpdate(action, ctx)
    return s.CachedUpdate
}
```

Add `xtracer` import to `interp/interp.go` (currently absent — verify
and add).

Note: this matches the Python lazy property's *behavior* but moves the
"action source" from `expr.rep` to `s.Action`. For all non-fail states
this is a no-op because they already have `CachedUpdate` set by
`ConcretePost` (`interp/eval.go:67`).

### Step 6 — Make `art.AnalysisGraph.GetHistory` lazily compute `state.Update`

**File**: `/Users/jaten/ivy/goivy/art/art.go` lines 749-786

Modify the existing `if state.Update != nil` block at lines 770-783 to
first attempt a lazy compute when `state.Update` is nil but
`state.Action` is set:

```go
// Recursive case: get history from predecessor.
var nextBound *int
if bound != nil {
    nb := *bound - 1
    nextBound = &nb
}
h := ag.GetHistory(state.Pred, nextBound)

// Lazy compute state.Update for the failState path. Python:
// ivy_interp.py:137-142 - the State.update property is accessed by
// history_forward_step (ivy_interp.py:597-599) inside ag.get_history.
//
// In Go we set failState.Action = NewFailAction(envAction) at
// construction time but leave failState.Update nil so the lazy compute
// fires here, exactly where Python's lazy property would fire.
if state.Update == nil && state.Action != nil {
    ctx := &actions.UpdateContext{
        Domain:       state.Domain,
        PVars:        state.InScope,
        ActCfg:       state.Domain.Cfg.ActCfg,
        Instantiator: state.Domain.Instantiator,
        GetAction: func(name string) actions.Action {
            if v, ok := state.Domain.Actions.Get2(name); ok {
                if act, ok := v.(actions.Action); ok {
                    return act
                }
            }
            return nil
        },
    }
    xtracer.Trace("interp.State.Update calling GetUpdate type=%s", actions.ActionTypeName(state.Action))
    state.Update = actions.GetUpdate(state.Action, ctx)
}

if state.Update != nil {
    var axioms *module.Clauses = module.TrueClauses(nil)
    if state.Pred != nil && state.Pred.Domain != nil {
        bgTheory := state.Pred.Domain.BackgroundTheory(state.Pred.InScope)
        if bgTheory != nil {
            axioms = bgTheory
        }
    }
    var actionNode lg.Expr = lg.True
    h = h.ForwardStep(axioms, state.Update, actionNode)
}
```

The `xtracer` import is already present in `art/art.go:20`.

### Step 7 — Wire the failState in `check/isolate_check.go`

**File**: `/Users/jaten/ivy/goivy/check/isolate_check.go`

#### 7a. Guarantee-loop fail state (lines 569-583)

```go
if post != nil {
    // Python (ivy_check.py:743-744):
    //     fail = itp.State(expr = itp.fail_expr(post.expr))
    //     if not check_safety_in_state(mod, ag, fail, report_pass=False):
    //
    // fail_expr(expr) = action_app(fail_action(expr.rep), expr.args[0])
    // The fail state's lazy update is computed in art.GetHistory once
    // we set failState.Action and failState.Pred.
    failState := art.NewState(mod, module.TrueClauses(actions.EmptyAnnotation{}))
    if aa, ok := post.Prov.(*art.ActionApp); ok {
        if rep, ok := aa.Rep.(string); ok {
            failState.Prov = art.NewActionApp("fail_"+rep, aa.Args...)
        }
    }

    // Wrap the original envAction with FailAction. Python's
    // fail_action(post.expr.rep) — see ivy_interp.py:406.
    failState.Action = actions.NewFailAction(envAction)

    // Set predecessor so ag.GetHistory walks back through it.
    // Python: fail.pred (lazy from fail.expr.args[0]) == pre.
    // post.Pred is set by interp.ApplyAction → InterpToArtState above;
    // fall back to post.Prov.Args[0] defensively.
    if post.Pred != nil {
        failState.Pred = post.Pred
    } else if aa, ok := post.Prov.(*art.ActionApp); ok && len(aa.Args) > 0 {
        failState.Pred = aa.Args[0]
    }

    if !CheckSafetyInStateWithAG(mod, ag, failState, false) {
        someFailed = true
        break
    }
}
```

#### 7b. Initializer-guarantees fail state (lines 316-328)

```go
if len(ag.States) > 0 {
    initState := ag.States[0]

    failState := art.NewState(mod, module.TrueClauses(actions.EmptyAnnotation{}))

    // Recover the inner action: AddInitialState stores either an
    // action object (line 1520 of art/art.go: NewActionApp(env, s) with
    // env being an *EnvAction) or a string label (when no initializers).
    var innerAction actions.Action
    if aa, ok := initState.Prov.(*art.ActionApp); ok {
        switch rep := aa.Rep.(type) {
        case string:
            failState.Prov = art.NewActionApp("fail_"+rep, aa.Args...)
            if v, ok := mod.Actions.Get2(rep); ok {
                if a, ok := v.(actions.Action); ok {
                    innerAction = a
                }
            }
        case actions.Action:
            failState.Prov = art.NewActionApp("fail_"+rep.Name(), aa.Args...)
            innerAction = rep
        }
    }
    if innerAction == nil {
        innerAction = initState.Action
    }

    if innerAction != nil {
        failState.Action = actions.NewFailAction(innerAction)
        if initState.Pred != nil {
            failState.Pred = initState.Pred
        } else if aa, ok := initState.Prov.(*art.ActionApp); ok && len(aa.Args) > 0 {
            failState.Pred = aa.Args[0]
        }
    }

    CheckSafetyInStateWithAG(mod, ag, failState, true)
}
```

If `innerAction` cannot be recovered (no `Prov`, no `Action` field),
we fall back to the previous behavior of calling
`CheckSafetyInStateWithAG` with an unwired failState — preserving the
existing (silent) behavior on shapes we don't recognize.

### Step 8 — Tests

Add to the new `actions/fail_action_test.go`:

```go
package actions

import (
    "testing"

    lg "github.com/glycerine/ivy/goivy/logic"
    "github.com/glycerine/ivy/goivy/module"
)

func TestFailActionUpdate(t *testing.T) {
    inner := NewAssumeAction(lg.True)
    fa := NewFailAction(inner)
    m := module.New()
    ctx := &UpdateContext{
        Domain:       m,
        PVars:        map[string]bool{},
        ActCfg:       m.Cfg.ActCfg,
        Instantiator: m.Instantiator,
        GetAction:    func(string) Action { return nil },
    }
    upd := fa.Update(ctx)
    if upd == nil {
        t.Fatal("FailAction.Update returned nil")
    }
    // ActionFailure (transrel.go:1706-1712) sets Pre = TrueClauses.
    if upd.Pre == nil || !upd.Pre.IsTrue() {
        t.Errorf("expected Pre=TrueClauses after ActionFailure, got %v", upd.Pre)
    }
}

func TestFailActionIntUpdate(t *testing.T) {
    inner := NewAssumeAction(lg.True)
    fa := NewFailAction(inner)
    m := module.New()
    ctx := &UpdateContext{
        Domain:       m,
        PVars:        map[string]bool{},
        ActCfg:       m.Cfg.ActCfg,
        Instantiator: m.Instantiator,
        GetAction:    func(string) Action { return nil },
    }
    upd := fa.IntUpdate(ctx)
    if upd == nil {
        t.Fatal("FailAction.IntUpdate returned nil")
    }
    if upd.Pre == nil || !upd.Pre.IsTrue() {
        t.Errorf("expected Pre=TrueClauses after ActionFailure, got %v", upd.Pre)
    }
}

func TestActionTypeNameFailAction(t *testing.T) {
    fa := NewFailAction(NewSequence())
    if got := ActionTypeName(fa); got != "fail_action" {
        t.Errorf("ActionTypeName(*FailAction) = %q, want %q", got, "fail_action")
    }
}

func TestGetUpdateBypassesFailEnter(t *testing.T) {
    // GetUpdate(*FailAction, ctx) must NOT emit "actions.GetUpdate ENTER
    // type=fail_action" — Python's fail_action.update overrides
    // Action.update and bypasses the base trace. We delegate to
    // fa.Update directly.
    //
    // This test only verifies that GetUpdate returns a non-nil result;
    // the actual xtrace text is verified by the make-golden flow.
    inner := NewAssumeAction(lg.True)
    fa := NewFailAction(inner)
    m := module.New()
    ctx := &UpdateContext{
        Domain:       m,
        PVars:        map[string]bool{},
        ActCfg:       m.Cfg.ActCfg,
        Instantiator: m.Instantiator,
        GetAction:    func(string) Action { return nil },
    }
    upd := GetUpdate(fa, ctx)
    if upd == nil {
        t.Fatal("GetUpdate(*FailAction, ctx) returned nil")
    }
}
```

Plus the moved tests from `interp/interp_test.go` (TestNewFailAction,
TestFailActionString, TestFailActionFailedAction, TestFailActionClone,
TestFailActionIterCalls, TestFailActionIterSubactions) — same bodies,
no `interp.` prefixes.

### Step 9 — Iterate on downstream divergences

After Steps 1-8, `make golden` will start emitting the missing fail
xtraces and the cascade behind them. This will almost certainly reveal
**new** divergences further down `log.red` because:

- Some inner-action xtraces may have minor formatting differences.
- `transrel.ComposeUpdates` and `transrel.iteUpdate` xtraces depend on
  exact module-symbol ordering; Go and Python use ordered-map iteration
  in slightly different ways.
- `history.satisfy` → `transrel.SatisfyWithCond ENTER` will now fire in
  additional contexts.

For each new divergence:

1. Re-read the Python xtrace at that line and locate it in `pyivy/`.
2. Locate the Go counterpart and identify the missing or mis-ordered
   trace.
3. Apply a minimal fix in the same mechanical-port spirit.
4. Re-run `make golden`.

The user explicitly said to "transitively plan any additional work
needed". The transitive work cannot be enumerated in advance because it
depends on what surfaces; it must be done iteratively. Each downstream
divergence gets its own minimal fix, in a separate edit, with `make
golden` re-run between each.

### Step 10 — Regression sweep before declaring done

```sh
cd /Users/jaten/ivy/goivy && go test ./interp/... ./check/... ./art/... ./actions/... ./bmc/... ./mc/... ./trace/... ./compiler/... ./parser/...
```

Verify in particular:

- `actions/fail_action_test.go` — all new tests pass.
- `interp/interp_test.go` — all FailAction tests have been moved out;
  the `TestEvalActionFailAction` test still works after the
  `interp.NewFailAction` → `actions.NewFailAction` rename (or is moved
  to actions/ as well).
- `trace/trace_test.go` — `TestTraceBaseFail` still passes after the
  field rename (`*actions.FailAction`, `.Inner`).
- `check/regression_test.go:TestRegression_Bug6_FailExpr` (line 195)
  still passes — it exercises the `"fail_"+rep` label path which we are
  keeping.
- `check/check_test.go`, `check/check_port_test.go`,
  `check/regression_test.go` all still pass.
- `make golden` (TestOrdLive and any other golden tests in
  `parser/golden_test.go`, `compiler/golden_ast_test.go`) succeeds at
  the end (after all downstream divergences from Step 9 are resolved).

## Critical Files

- `/Users/jaten/ivy/goivy/actions/fail_action.go` — **new file**, the
  unified FailAction (Step 1).
- `/Users/jaten/ivy/goivy/actions/fail_action_test.go` — **new file**,
  unit tests (Step 8).
- `/Users/jaten/ivy/goivy/actions/update.go` — add cases in
  `ActionTypeName`, `GetUpdate`, `IntUpdate` (Step 2).
- `/Users/jaten/ivy/goivy/interp/helpers.go` — delete the old
  `FailAction` block at lines 44-121 (Step 3).
- `/Users/jaten/ivy/goivy/interp/eval.go` — clean up dead `*FailAction`
  branch at lines 115-122 (Step 3).
- `/Users/jaten/ivy/goivy/interp/interp_test.go` — move FailAction
  tests out (Step 3).
- `/Users/jaten/ivy/goivy/interp/interp.go` — make `State.Update()` a
  real lazy property at lines 137-149 (Step 5).
- `/Users/jaten/ivy/goivy/trace/trace.go` — delete the duplicate
  `FailAction` block at lines 26-68; rewrite `Fail()` at line 348-350
  to use `actions.NewFailAction` (Step 4).
- `/Users/jaten/ivy/goivy/trace/trace_test.go` — update
  `TestTraceBaseFail`; delete the four duplicate FailAction tests at
  lines 182-213 (Step 4).
- `/Users/jaten/ivy/goivy/art/art.go` — add lazy compute in
  `GetHistory` at lines 770-783 (Step 6).
- `/Users/jaten/ivy/goivy/check/isolate_check.go` — wire failState in
  the guarantee loop and initializer paths (Step 7).

## Existing functions and helpers reused (do not re-implement)

- `actions.ActionFailure(*Update) *Update` —
  `/Users/jaten/ivy/goivy/actions/transrel.go:1706` (Python's
  `action_failure`).
- `actions.ActionTypeName(action) string` —
  `/Users/jaten/ivy/goivy/actions/update.go:1085`.
- `actions.GetUpdate(action, *UpdateContext) *Update` —
  `/Users/jaten/ivy/goivy/actions/update.go:2139`.
- `actions.IntUpdate(action, *UpdateContext) *Update` —
  `/Users/jaten/ivy/goivy/actions/update.go:1151`.
- `actions.UpdateContext` struct (shape mirrored from
  `/Users/jaten/ivy/goivy/interp/eval.go:149-164`).
- `actions.NewSequence`, `actions.NewAssumeAction`, etc. — for tests.
- `mod.Cfg.ActCfg` — `/Users/jaten/ivy/goivy/module/config.go:143`.
- `mod.Instantiator` — `/Users/jaten/ivy/goivy/module/module.go:133`.
- `art.AnalysisGraph.GetHistory(state, bound)` —
  `/Users/jaten/ivy/goivy/art/art.go:749`.
- `actions.History.ForwardStep(axioms, update, action)` —
  `/Users/jaten/ivy/goivy/actions/transrel.go:2090`.
- `xtracer.Trace` — already imported by `actions/`, `art/`,
  `interp/eval.go`, `check/isolate_check.go`.

## Things deliberately NOT changed (and why)

- **`interp.FailExpr`** at `interp/helpers.go:651-654` is left alone.
  It is never called from any production code path that touches the
  divergence; it remains as a string-based fail-label helper for any
  future code that wants a display-only label.

- **`art.State.Update` field** is *not* converted to a method. The
  lazy compute is added at the one consumer that needs it
  (`art.GetHistory`). Other consumers (`StateValue` accessor,
  `interpToArtMemo` conversion) read the field directly and still
  see nil for an un-computed failState — this is fine because none of
  them are on the divergence path.

- **`interp.HistoryForwardStep`** at `interp/helpers.go:389-403` is
  left alone. It is defined but not called from production code (only
  some action tests reference it indirectly), and updating it is not
  on the divergence path.

## Verification

End-to-end:

1. After Step 1 (new `actions/fail_action.go`): `go build ./actions/...`
   succeeds. The new methods are unused but compile-clean.
2. After Step 2 (dispatch wiring): `go build ./actions/...` still
   succeeds.
3. After Step 3 (delete old `interp.FailAction` and update references):
   `go build ./interp/...` succeeds. Run `go test ./interp/... -run
   FailAction` — tests should be removed/moved cleanly.
4. After Step 4 (delete `trace.FailAction`): `go build ./trace/...`
   succeeds; `go test ./trace/...` passes (with the renamed
   `TestTraceBaseFail`).
5. After Step 5 (lazy `interp.State.Update`): `go build ./interp/...`
   succeeds. No functional change yet because nothing calls it on a
   FailAction-bearing state.
6. After Step 6 (lazy `art.GetHistory`): `go build ./art/...` succeeds.
7. After Step 7 (failState wiring): `go build ./check/...` succeeds.
   Now `go test ./check/...` may fail if the new path triggers
   downstream xtrace mismatches — that is expected and is the entry
   point to Step 9.
8. After Step 8 (unit tests): `go test ./actions/... -run FailAction
   -v` passes.
9. `make golden` regenerates `log.red`. Inspect line 233846: Go now
   emits `"interp.State.Update calling GetUpdate type=fail_action"`
   matching Python.
10. Iterate Step 9 until `make golden` is clean.
11. Run the regression sweep in Step 10.

Sanity-check signals during iteration:

- `postFmlas=` in `check.checkFcsNormalPath` should be **non-zero**
  after the fix (it was 0 before because `GetHistory` was returning the
  empty initial pure state).
- The new xtraces should appear in this order at the divergence site:
  1. `interp.State.Update calling GetUpdate type=fail_action`
  2. `interp.ActionFail calling GetUpdate type=EnvAction`
  3. `actions.GetUpdate ENTER type=EnvAction`
  4. `actions.IntUpdate ENTER type=EnvAction` (and downstream)
  5. transrel cascade (`ComposeUpdates`, `iteUpdate`, `Hide`,
     `FrameDefConst`, ...)
  6. eventually `transrel.SatisfyWithCond ENTER`
  7. eventually `check.checkFcsNormalPath history path postFmlas=N
     postDefs=M ...` with N > 0 / M > 0.

Risks and mitigations:

- **`post.Pred` could be nil** in some shapes of the ART. Mitigation:
  the explicit `if post.Pred != nil { ... } else if aa.Args[0] ...`
  fallback. Confirmed reachable in the normal path: `art.PostState`
  calls `interp.ApplyAction` → `ConcretePost`
  (`interp/eval.go:67-69`) which calls `res.SetPred(state)`, then
  `InterpToArtState` walks the predecessor through `interpToArtMemo`
  (`art/art.go:1686-1688`).
- **`mod.Cfg.ActCfg` / `mod.Instantiator` could be nil**. They are
  populated during isolate setup and are non-nil by the time
  `check.guarantee_phase` runs (verified by usage in
  `interp.ApplyAction` and `interp/phase4.go:186`).
- **Trace tests for `"FAIL"` capitalization**: deleted in Step 4
  because they tested wrong (non-Python) behavior. The renamed
  `TestTraceBaseFail` covers the wrap functionality.
- **Identity of `failState.Pred`** does not match Python's identity
  (Python reuses the original `pre` object; Go uses a converted copy
  via `interpToArtMemo`). This does not affect `ag.GetHistory`, which
  only walks the `Pred` chain and reads `Clauses` / `InScope` /
  `Update`.
- **Step 7b's initializer path** depends on `ag.States[0].Prov.Rep`
  being either a string or an `actions.Action`. If neither, the
  fallback to `initState.Action` keeps the existing behavior so we
  never crash on unfamiliar shapes.
- **`xtracer.Trace` in `actions/fail_action.go`**: the `xtracer`
  package is already used heavily by `actions/update.go` and similar.
  No new dependency edge is introduced.
- **Tests that exercise the trace builder Fail path with concrete
  inner actions**: the renamed `TestTraceBaseFail` only tests wrap
  identity, not display. If any other test asserts on the printed form
  of a failed trace state and finds `"fail xxx"` instead of
  `"FAIL(xxx)"`, update the assertion to match Python.
