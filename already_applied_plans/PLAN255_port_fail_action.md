# Plan: Port Python `fail_action` lazy `state.update` flow into Go

Created: 2026-04-10 (afternoon, plan-mode session)

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

In `~/ivy/pyivy/ivy/ivy/ivy_check.py` around line 720-746, the guarantee
loop builds a `fail` state and asks `check_safety_in_state` to verify it:

```python
xtracer.trace("check.guarantee_loop pre BuildEnvAction")
action = act.env_action(root)                       # → EnvAction
ag = ivy_art.AnalysisGraph()
pre = itp.State(); pre.clauses = get_conjs(mod)
with itp.EvalContext(check=False):
    post = ag.execute(action, prestate=pre)
fail = itp.State(expr = itp.fail_expr(post.expr))   # ← fresh State, lazy update
if not check_safety_in_state(mod, ag, fail, report_pass=False):
    ...
```

`fail_expr` (`ivy_interp.py:406`) constructs
`action_app(fail_action(expr.rep), expr.args[0])`. So `fail.expr.rep` is
a `fail_action(EnvAction)` object and `fail.expr.args[0]` is `pre`.

`check_safety_in_state` → `check_fcs_in_state` (`ivy_check.py:377`):

```python
history = ag.get_history(post)              # walks back via state.pred
gmc = ...
axioms = im.module.background_theory()
if opt_trace.get() or diagnose.get():
    ...
else:
    xtracer.trace("check.checkFcsNormalPath history path ...")
    res = history.satisfy(axioms, gmc, filter_fcs(fcs))
```

`ag.get_history(fail)` (`ivy_art.py:332-337`) recurses on `state.pred`,
calling `history_forward_step` (`ivy_interp.py:597-599`):

```python
def history_forward_step(history, state):
    action = state.expr.rep if state.expr is not None and is_action_app(state.expr) else None
    return history.forward_step(state.pred.domain.background_theory(state.pred.in_scope),
                                state.update,         # ← LAZY ACCESS
                                action)
```

The `state.update` access triggers the lazy property in
`ivy_interp.py:137-142`:

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
(it is an `Action` subclass, so `isinstance(expr, Action)` is true), and
`.update(...)` invokes the override at `ivy_interp.py:385-389`:

```python
def update(self, domain, in_scope):
    if __debug__: xtracer.trace("interp.ActionFail calling GetUpdate type=%s" % type(self.action).__name__)
    upd = self.action.update(domain, in_scope)   # base Action.update — full cascade
    return action_failure(upd)
```

This emits **two new xtraces** Go is missing entirely
(`interp.State.Update calling GetUpdate type=fail_action` and
`interp.ActionFail calling GetUpdate type=EnvAction`), then calls
`self.action.update(...)` which produces an entire cascade of
`actions.GetUpdate ENTER`, `actions.IntUpdate ENTER`,
`transrel.ComposeUpdates`, `transrel.iteUpdate`, ..., feeding a
non-empty `Update` into `history.forward_step`, which then folds into
`history.post` so `checkFcsNormalPath` finally fires with non-zero
`postFmlas` / `postDefs`.

### Go state today

1. **`interp.FailAction` already exists** at
   `/Users/jaten/ivy/goivy/interp/helpers.go:48-121` with the right
   embedding (`actions.ActionBase`), `Inner` field, `Name() == "fail"`,
   `Sexp()`, etc. **But it has no `Update` or `IntUpdate` method.**

2. **`interp.State.Update()` is a stub** at
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

3. **`interp.FailExpr`** at
   `/Users/jaten/ivy/goivy/interp/helpers.go:651-654` builds an
   `*ast.Atom` whose `Rep` is the **string** `"fail_<name>"`. This loses
   the wrapped action object entirely, because `ast.Atom.Rep` is a Go
   `string` and cannot hold a `*FailAction`. The current Go code uses
   that string-prefix as a label and never reconstructs a `FailAction`
   from it.

4. **The fail-state construction in `check/isolate_check.go`** never
   sets `Pred` or `Update`:

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
   - Lines 316-328 (initializer guarantees) — same shape.

   `failState.Pred` and `failState.Update` are both nil.

5. **`art.AnalysisGraph.GetHistory`** at
   `/Users/jaten/ivy/goivy/art/art.go:749-786` reads `state.Update` as a
   **field** (not a lazy method), and short-circuits when
   `state.Pred == nil`:

   ```go
   if state.Pred == nil || (bound != nil && *bound <= 0) {
       clauses := state.Clauses
       if clauses == nil { clauses = module.TrueClauses(nil) }
       u := actions.PureStateClauses(clauses)
       return actions.NewHistory(ag.Domain.Cfg.IuCfg, u)
   }
   ...
   if state.Update != nil {
       ...
       h = h.ForwardStep(axioms, state.Update, actionNode)
   }
   ```

6. **`actions.GetUpdate(action, ctx)`** at
   `/Users/jaten/ivy/goivy/actions/update.go:2139-2145` always emits
   `"actions.GetUpdate ENTER type=%s"`. Its dispatch through
   `actions.IntUpdate` is an **explicit type switch** with no fallback
   for foreign Action types — `*interp.FailAction` would hit the
   `default` branch and return `NullUpdate()`.

7. **`actions.ActionFailure(u *Update) *Update`** is already
   implemented at `/Users/jaten/ivy/goivy/actions/transrel.go:1706-1712`
   — exactly the Go counterpart of Python's `action_failure(action)`.

### Why the divergence is at line 233846 (and not earlier)

Up through line 233845 the two engines match because everything that
matters runs inside `ag.Execute(envAction, ...)` (and its descendants),
which is already correctly ported. The divergence happens at the **first
xtrace after `ag.Execute` returns**, because Python's lazy
`fail.update` access fires inside `check_fcs_in_state(mod, ag, fail,
...)` and Go's GetHistory short-circuits on the un-wired fail state and
falls straight into `checkFcsNormalPath` with an empty post.

### Why we cannot just emit one xtrace and call it done

Fixing only the missing `interp.State.Update calling GetUpdate
type=fail_action` xtrace would not move the divergence — Python emits a
**cascade** of xtraces after that line: `interp.ActionFail calling
GetUpdate`, `actions.GetUpdate ENTER`, the inner action's
`actions.IntUpdate ENTER`, all of `transrel.ComposeUpdates / iteUpdate
/ Hide / FrameDefConst`, and finally `transrel.SatisfyWithCond ENTER`
when the history is consulted. We need Go to **really compute** the
fail-action update (so all those traces fire) **and** to wire it onto
the failState so that `ag.GetHistory` actually performs the
`ForwardStep` Python performs.

## Approach

We mechanically port Python's `fail_action.update` /
`fail_action.int_update` onto the existing Go `interp.FailAction` type,
and at the call site in `check/isolate_check.go` we eagerly run the
same computation Python's lazy `state.update` getter would run,
emitting the matching xtrace just before the call.

We do **not** try to make `interp.State.Update()` itself genuinely
lazy. The reason: Python's lazy property looks up the action by walking
`self.expr.rep`, but Go's `ast.Atom.Rep` is a `string` and cannot carry
a `*FailAction` object. Refactoring `ast.Atom` to allow a non-string
`Rep` is a structural change far outside the scope of this fix and
would touch many unrelated files. Eagerly emitting the xtrace at the
call site (where we know the FailAction exists in scope as a real Go
value) achieves Python's *observable* semantics — same xtrace text,
same emission order — without that refactor. A short comment on
`FailAction.Update` documents why it is invoked directly rather than
through `actions.GetUpdate(failAction, ctx)` (which would also emit a
spurious `"actions.GetUpdate ENTER type=fail"` line that has no Python
counterpart, because Python's `fail_action.update` *overrides*
`Action.update` and bypasses its trace entirely).

We do **not** add an interface dispatch for `FailAction` inside
`actions.IntUpdate`, because `actions` cannot import `interp` (cycle),
and because routing through that dispatch would emit the spurious
`actions.GetUpdate ENTER type=fail` trace described above.

### Step 1 — Add `Update` and `IntUpdate` methods to `*interp.FailAction`

**File**: `/Users/jaten/ivy/goivy/interp/helpers.go`

Add `xtracer` to the import block:
```go
"github.com/glycerine/ivy/goivy/xtracer"
```

Append two methods just after `FailedAction()` at line 121:

```go
// Update wraps the inner action's full update with action_failure,
// matching Python fail_action.update (ivy_interp.py:385-389):
//
//     def update(self,domain,in_scope):
//         xtracer.trace("interp.ActionFail calling GetUpdate type=%s" % type(self.action).__name__)
//         upd = self.action.update(domain,in_scope)
//         return action_failure(upd)
//
// IMPORTANT: this method is invoked directly from the call site (e.g.
// from check/isolate_check.go when it constructs the synthetic fail
// state), NOT through actions.GetUpdate(fa, ctx). Two reasons:
//   1. The actions package cannot import interp (import cycle), so
//      actions.IntUpdate's type switch cannot dispatch FailAction and
//      would fall through to the default-NullUpdate branch.
//   2. Python's fail_action.update is an OVERRIDE of Action.update and
//      therefore does NOT emit "actions.GetUpdate ENTER type=fail_action"
//      — only "interp.ActionFail calling GetUpdate type=<inner>". Calling
//      actions.GetUpdate(fa, ctx) here would emit a spurious "ENTER"
//      trace with no Python counterpart.
func (fa *FailAction) Update(ctx *actions.UpdateContext) *actions.Update {
    xtracer.Trace("interp.ActionFail calling GetUpdate type=%s", actions.ActionTypeName(fa.Inner))
    upd := actions.GetUpdate(fa.Inner, ctx)
    return actions.ActionFailure(upd)
}

// IntUpdate mirrors Python fail_action.int_update (ivy_interp.py:390-393):
//
//     def int_update(self,domain,in_scope):
//         xtracer.trace("interp.ActionFail calling IntUpdate type=%s" % type(self.action).__name__)
//         return action_failure(self.action.int_update(domain,in_scope))
//
// Same direct-call rationale as Update.
func (fa *FailAction) IntUpdate(ctx *actions.UpdateContext) *actions.Update {
    xtracer.Trace("interp.ActionFail calling IntUpdate type=%s", actions.ActionTypeName(fa.Inner))
    return actions.ActionFailure(actions.IntUpdate(fa.Inner, ctx))
}
```

Both functions reference symbols that already exist:
- `actions.GetUpdate` — `/Users/jaten/ivy/goivy/actions/update.go:2139`
- `actions.IntUpdate` — `/Users/jaten/ivy/goivy/actions/update.go:1151`
- `actions.ActionFailure` — `/Users/jaten/ivy/goivy/actions/transrel.go:1706`
- `actions.ActionTypeName` — `/Users/jaten/ivy/goivy/actions/update.go` (used widely)
- `xtracer.Trace` — already used elsewhere in `interp/`.

### Step 2 — Wire the guarantee-loop fail state with `Pred` and `Update`

**File**: `/Users/jaten/ivy/goivy/check/isolate_check.go` lines 569-583

Add `"github.com/glycerine/ivy/goivy/interp"` to the import block at
the top of the file (it is currently absent — confirmed via grep).

Replace the `if post != nil { ... }` block with:

```go
if post != nil {
    // Python (ivy_check.py:743-744):
    //     fail = itp.State(expr = itp.fail_expr(post.expr))
    //     if not check_safety_in_state(mod, ag, fail, report_pass=False):
    //
    // fail_expr(expr) = action_app(fail_action(expr.rep), expr.args[0])
    //
    // Python's fail.update is a lazy property whose getter (ivy_interp.py:137-142)
    // first emits "interp.State.Update calling GetUpdate type=fail_action"
    // and then invokes fail_action.update(domain, in_scope), which itself
    // emits "interp.ActionFail calling GetUpdate type=<inner>" and runs
    // the inner Action.update cascade. Go cannot stash a *FailAction inside
    // an *ast.Atom.Rep (it is string-only), so we eagerly compute the same
    // result here at construction time and emit the State.Update xtrace at
    // the same point Python would emit it on first lazy access.
    failState := art.NewState(mod, module.TrueClauses(actions.EmptyAnnotation{}))
    if aa, ok := post.Prov.(*art.ActionApp); ok {
        if rep, ok := aa.Rep.(string); ok {
            failState.Prov = art.NewActionApp("fail_"+rep, aa.Args...)
        }
    }

    // Build the FailAction wrapping the original envAction (the action
    // that produced post). Python: fail_action(post.expr.rep).
    failAction := interp.NewFailAction(envAction)

    // Build the UpdateContext that fail_action.update would receive.
    // Python passes (self.domain, self.in_scope) where self is a fresh
    // itp.State(): the ambient module and an EMPTY in_scope. Mirror that
    // exactly — do NOT inherit pre.InScope or post.InScope.
    failCtx := &actions.UpdateContext{
        Domain:       mod,
        PVars:        make(map[string]bool),
        ActCfg:       mod.Cfg.ActCfg,
        Instantiator: mod.Instantiator,
        GetAction: func(name string) actions.Action {
            if v, ok := mod.Actions.Get2(name); ok {
                if a, ok := v.(actions.Action); ok {
                    return a
                }
            }
            return nil
        },
    }

    // Python's lazy State.update getter (ivy_interp.py:140) emits this:
    xtracer.Trace("interp.State.Update calling GetUpdate type=fail_action")

    // failAction.Update emits "interp.ActionFail calling GetUpdate type=EnvAction"
    // and recursively runs actions.GetUpdate(envAction, failCtx), which
    // produces all the inner xtraces (actions.GetUpdate ENTER, IntUpdate
    // ENTER, transrel.ComposeUpdates, ...) that Python emits.
    failUpd := failAction.Update(failCtx)

    // Wire the synthetic fail state into the ART so ag.GetHistory walks
    // back through it. Python: fail.pred = post.expr.args[0] (== pre),
    // fail.update = action_failure(action.update(...)).
    //
    // post.Pred is set by interp.ApplyAction → InterpToArtState during
    // ag.Execute above. As a defensive fallback (in case the conversion
    // chain ever leaves it nil), use post.Prov.Args[0].
    if post.Pred != nil {
        failState.Pred = post.Pred
    } else if aa, ok := post.Prov.(*art.ActionApp); ok && len(aa.Args) > 0 {
        failState.Pred = aa.Args[0]
    }
    failState.Update = failUpd

    if !CheckSafetyInStateWithAG(mod, ag, failState, false) {
        someFailed = true
        break
    }
}
```

### Step 3 — Mirror Step 2 for the initializer-guarantees path

**File**: `/Users/jaten/ivy/goivy/check/isolate_check.go` lines 316-328

This block runs assertions in initializers. The "post" state here is
`ag.States[0]`, produced by `art.AnalysisGraph.AddInitialState`
(`art/art.go:1480-1552`). When initializers exist,
`AddInitialState` builds `EnvAction(Sequence(<initializers>,
ReturnAction))`, calls `interp.ApplyAction` on it, and stashes the
constructed `*ActionApp{Rep: env, Args: [s]}` on `s2.Prov` (where
`Rep` is the **EnvAction object**, NOT a string — so the existing
`if rep, ok := aa.Rep.(string)` branch silently does nothing on this
path; that pre-existing oversight should be fixed in the same edit
to keep the printed label correct).

Replace lines 316-328 with the same pattern as Step 2, but with the
following adjustments:

- The "envAction" is recovered from `ag.States[0].Prov` (which carries
  the `*art.ActionApp` whose `Rep` is the `*actions.EnvAction` built by
  `AddInitialState`), or — if the type assertion fails — fall back to
  `ag.States[0].Action` (set by `interp.ApplyAction` in
  `eval.go:184`).
- The `failState.Prov` label uses Python's display semantics: when
  `Rep` is an action, the "name" comes from the action; when `Rep` is a
  string, prepend `"fail_"`.
- The predecessor wiring: `failState.Pred = ag.States[0].Pred` with the
  same `aa.Args[0]` fallback.

```go
if len(ag.States) > 0 {
    // Python (ivy_check.py:622-625):
    //     fail = itp.State(expr = itp.fail_expr(ag.states[0].expr))
    //     check_safety_in_state(mod, ag, fail)
    initState := ag.States[0]

    failState := art.NewState(mod, module.TrueClauses(actions.EmptyAnnotation{}))

    var initEnvAction actions.Action
    if aa, ok := initState.Prov.(*art.ActionApp); ok {
        switch rep := aa.Rep.(type) {
        case string:
            failState.Prov = art.NewActionApp("fail_"+rep, aa.Args...)
            if v, ok := mod.Actions.Get2(rep); ok {
                if a, ok := v.(actions.Action); ok {
                    initEnvAction = a
                }
            }
        case actions.Action:
            failState.Prov = art.NewActionApp("fail_"+rep.Name(), aa.Args...)
            initEnvAction = rep
        }
    }
    if initEnvAction == nil {
        initEnvAction = initState.Action
    }
    if initEnvAction == nil {
        // No discoverable inner action; fall back to the previous behavior
        // so we at least call CheckSafetyInStateWithAG with an empty post.
        CheckSafetyInStateWithAG(mod, ag, failState, true)
    } else {
        failAction := interp.NewFailAction(initEnvAction)
        failCtx := &actions.UpdateContext{
            Domain:       mod,
            PVars:        make(map[string]bool),
            ActCfg:       mod.Cfg.ActCfg,
            Instantiator: mod.Instantiator,
            GetAction: func(name string) actions.Action {
                if v, ok := mod.Actions.Get2(name); ok {
                    if a, ok := v.(actions.Action); ok {
                        return a
                    }
                }
                return nil
            },
        }
        xtracer.Trace("interp.State.Update calling GetUpdate type=fail_action")
        failUpd := failAction.Update(failCtx)

        if initState.Pred != nil {
            failState.Pred = initState.Pred
        } else if aa, ok := initState.Prov.(*art.ActionApp); ok && len(aa.Args) > 0 {
            failState.Pred = aa.Args[0]
        }
        failState.Update = failUpd

        CheckSafetyInStateWithAG(mod, ag, failState, true)
    }
}
```

### Step 4 — Unit tests for `FailAction.Update` / `FailAction.IntUpdate`

**File**: `/Users/jaten/ivy/goivy/interp/interp_test.go`, near the
existing FailAction tests at lines 612-666.

```go
func TestFailActionUpdateXtrace(t *testing.T) {
    // Build a minimal AssumeAction so the inner GetUpdate is non-trivial.
    inner := actions.NewAssumeAction(lg.True)
    fa := NewFailAction(inner)
    m := module.New()
    ctx := &actions.UpdateContext{
        Domain:       m,
        PVars:        map[string]bool{},
        ActCfg:       m.Cfg.ActCfg,
        Instantiator: m.Instantiator,
        GetAction:    func(string) actions.Action { return nil },
    }
    upd := fa.Update(ctx)
    if upd == nil {
        t.Fatal("FailAction.Update returned nil")
    }
    // Per actions/transrel.go:1706-1712, ActionFailure sets Pre = TrueClauses.
    if upd.Pre == nil || !upd.Pre.IsTrue() {
        t.Errorf("expected Pre==TrueClauses after ActionFailure, got %v", upd.Pre)
    }
}

func TestFailActionIntUpdateXtrace(t *testing.T) {
    inner := actions.NewAssumeAction(lg.True)
    fa := NewFailAction(inner)
    m := module.New()
    ctx := &actions.UpdateContext{
        Domain:       m,
        PVars:        map[string]bool{},
        ActCfg:       m.Cfg.ActCfg,
        Instantiator: m.Instantiator,
        GetAction:    func(string) actions.Action { return nil },
    }
    upd := fa.IntUpdate(ctx)
    if upd == nil {
        t.Fatal("FailAction.IntUpdate returned nil")
    }
    if upd.Pre == nil || !upd.Pre.IsTrue() {
        t.Errorf("expected Pre==TrueClauses after ActionFailure, got %v", upd.Pre)
    }
}
```

If `actions.NewAssumeAction` is not the right constructor, mirror
whatever the existing `TestNewFailAction` (line 612) and
`TestFailActionString` (line 623) tests use to obtain a working inner
action, falling back to `actions.NewSequence()` (which they already
use).

### Step 5 — Iterate on downstream divergences

After Steps 1-3, `make golden` will start emitting the missing fail
xtraces and the cascade behind them. This will almost certainly reveal
**new** divergences further down `log.red` because:

- Some inner-action xtraces may have minor formatting differences.
- `transrel.ComposeUpdates` and `transrel.iteUpdate` xtraces depend on
  exact module-symbol ordering; Go and Python use `OrderedDict`-like
  iteration in slightly different ways.
- `history.satisfy` → `transrel.SatisfyWithCond ENTER` — currently
  only fires once per check, but with the new fail-state path it will
  fire in additional contexts.

For each new divergence:

1. Re-read the Python xtrace at that line and locate it in `pyivy/`.
2. Locate the Go counterpart and identify the missing or mis-ordered
   trace.
3. Apply a minimal fix in the same mechanical-port spirit.
4. Re-run `make golden`.

The user explicitly said to "transitively plan any additional work
needed". The transitive work cannot be enumerated in advance because it
depends on what surfaces; it must be done iteratively. **The plan
explicitly includes this iterative phase.** Each downstream divergence
gets its own minimal fix, in a separate code-edit step, with `make
golden` re-run between each.

### Step 6 — Regression sweep before declaring done

```sh
cd /Users/jaten/ivy/goivy && go test ./interp/... ./check/... ./art/... ./actions/... ./bmc/... ./mc/... ./trace/... ./compiler/... ./parser/...
```

Particularly verify:
- `interp.TestNewFailAction`, `TestFailActionString`,
  `TestFailActionFailedAction`, `TestFailActionClone`,
  `TestFailActionIterCalls`, `TestFailActionIterSubactions` and the new
  `TestFailActionUpdateXtrace`, `TestFailActionIntUpdateXtrace` all pass.
- `check.TestRegression_Bug6_FailExpr`
  (`/Users/jaten/ivy/goivy/check/regression_test.go:195`) still passes
  — it exercises the `"fail_"+rep` label path which we are keeping.
- `check.TestCheckSafetyInState` and the rest of `check_test.go`,
  `check_port_test.go`, `regression_test.go` still pass.
- `make golden` (TestOrdLive and any other golden tests in
  `parser/golden_test.go`, `compiler/golden_ast_test.go`) succeeds at
  the end.

## Critical Files

- `/Users/jaten/ivy/goivy/interp/helpers.go` — add `xtracer` import; add
  `Update` and `IntUpdate` methods on `*FailAction` (Step 1).
- `/Users/jaten/ivy/goivy/check/isolate_check.go` — add `interp`
  import; replace fail-state construction at lines 569-583 (Step 2);
  replace fail-state construction at lines 316-328 (Step 3).
- `/Users/jaten/ivy/goivy/interp/interp_test.go` — add unit tests
  (Step 4).

## Existing functions and helpers reused (do not re-implement)

- `interp.NewFailAction(inner)` —
  `/Users/jaten/ivy/goivy/interp/helpers.go:56`.
- `actions.GetUpdate(action, *UpdateContext)` —
  `/Users/jaten/ivy/goivy/actions/update.go:2139`.
- `actions.IntUpdate(action, *UpdateContext)` —
  `/Users/jaten/ivy/goivy/actions/update.go:1151`.
- `actions.ActionFailure(*Update) *Update` —
  `/Users/jaten/ivy/goivy/actions/transrel.go:1706`.
- `actions.ActionTypeName(action)` —
  `/Users/jaten/ivy/goivy/actions/update.go`.
- `actions.UpdateContext` struct (shape mirrored from
  `interp.ApplyAction` at `/Users/jaten/ivy/goivy/interp/eval.go:149-164`).
- `mod.Cfg.ActCfg` — `/Users/jaten/ivy/goivy/module/config.go:143`.
- `mod.Instantiator` — `/Users/jaten/ivy/goivy/module/module.go:133`.
- `art.AnalysisGraph.GetHistory(state, bound)` —
  `/Users/jaten/ivy/goivy/art/art.go:749` (consumes the new
  `failState.Pred` / `failState.Update` we set).
- `actions.History.ForwardStep(axioms, update, action)` —
  `/Users/jaten/ivy/goivy/actions/transrel.go:2090` (consumed by
  `GetHistory`).
- `xtracer.Trace` — already used by `interp/eval.go`,
  `actions/update.go`, `check/isolate_check.go`.

## Things deliberately NOT changed (and why)

- **`interp.State.Update()` stub** at `interp/interp.go:142-149` is
  *not* replaced with a lazy implementation. Reason: implementing
  Python's `state.update` getter requires walking `state.expr.rep` and
  resolving it to an Action, but `ast.Atom.Rep` is `string`-typed in
  Go. Refactoring `ast.Atom` to allow non-string `Rep` is structurally
  invasive and unrelated to this divergence. The eager call-site
  approach in Steps 2-3 produces the same observable xtraces and the
  same `ag.GetHistory` behavior. (If a future divergence ever needs a
  truly lazy `interp.State.Update`, that can be a separate plan.)

- **`interp.FailExpr`** at `interp/helpers.go:651-654` is left alone.
  It is never called from any production code path that touches the
  divergence; it remains as a string-based fail-label helper.

- **`actions.IntUpdate` type-switch dispatch** is *not* extended with a
  case for `*interp.FailAction`. Reason: `actions` cannot import
  `interp` (cycle), and routing a `FailAction` through
  `actions.GetUpdate` would emit `"actions.GetUpdate ENTER type=fail"`,
  which has no Python counterpart. Calling
  `failAction.Update(ctx)` directly is the correct mechanical port
  (Python's `fail_action.update` is an *override* of `Action.update`).

- **`art.GetHistory`** is *not* changed. Once `failState.Pred` and
  `failState.Update` are set, the existing field-access path in
  `art/art.go:773` (`if state.Update != nil`) does the right thing.

- **`art.State.Update` field** is *not* converted to a lazy method. The
  same justification applies: structural change far outside the scope
  of this fix, with no mechanical Python counterpart.

## Verification

End-to-end:

1. `cd /Users/jaten/ivy/goivy && go build ./...` succeeds after Step 1
   alone (new methods are unused but compile-clean).
2. `go test ./interp/... -run TestFailAction -v` passes after Step 4.
3. `go build ./...` succeeds after Steps 2 and 3 (the new
   `interp` import in `isolate_check.go` is exercised).
4. `make golden` (or `go test ./parser/... -run TestOrdLive -v`)
   regenerates `log.red`. Inspect line 233846: Go now emits
   `"interp.State.Update calling GetUpdate type=fail_action"` matching
   Python.
5. The divergence point moves further down. Iterate Step 5 until
   `make golden` is clean.
6. Run the regression sweep in Step 6.

Sanity-check signals during iteration:

- `postFmlas=` in `check.checkFcsNormalPath` should be **non-zero**
  after the fix (it was 0 before because `GetHistory` was returning the
  empty initial pure state).
- The new xtraces should appear in this order at the divergence site:
  1. `interp.State.Update calling GetUpdate type=fail_action`
  2. `interp.ActionFail calling GetUpdate type=EnvAction`
  3. `actions.GetUpdate ENTER type=EnvAction`
  4. `actions.IntUpdate ENTER type=EnvAction` (and downstream)
  5. transrel cascade
  6. eventually `transrel.SatisfyWithCond ENTER`
  7. eventually `check.checkFcsNormalPath history path postFmlas=N
     postDefs=M ...` with N > 0 / M > 0.

Risks and mitigations:

- **`post.Pred` could be nil** in some shapes of the ART. Mitigation:
  the explicit `if post.Pred != nil { ... } else if aa.Args[0] ...`
  fallback. Confirmed reachable: `art.PostState` calls
  `interp.ApplyAction` → `ConcretePost` (`interp/eval.go:67`) which
  calls `res.SetPred(state)`, then `InterpToArtState` walks the
  predecessor through `interpToArtMemo` (`art/art.go:1686-1688`),
  yielding `post.Pred != nil` in the normal path.
- **`mod.Cfg.ActCfg` / `mod.Instantiator` could be nil**. They are
  populated during isolate setup and are non-nil by the time
  `check.guarantee_phase` runs (verified by usage in
  `interp.ApplyAction` and `interp/phase4.go:186`).
- **Identity of `failState.Pred`** does not match Python (Python
  reuses the original `pre` object; Go uses a converted copy via
  `interpToArtMemo`). This does not affect `ag.GetHistory`, which only
  walks the `Pred` chain and reads `Clauses` / `InScope` / `Update`.
- **Step 3's initializer path** depends on `ag.States[0].Prov.Rep`
  being either a string or an `actions.Action`. If neither, the
  fallback to `initState.Action` keeps the existing behavior (no fail
  update wiring) so we never crash on unfamiliar shapes.
