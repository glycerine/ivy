# Plan: Generalized Loc propagation for Action constructors (eliminate 1610 `(internal) assumption` lines)

Created: 2026-04-09 16:00

## Context

After the previous fix landed, `make golden` shows the impl/monitor/init
sections in `~/ivy/goivy/log.red` matching Python correctly, but the
"The following program assertions are treated as assumptions:" section
still has 1610 `(internal) assumption` lines on the Go side that have no
counterpart in Python:

```
~go[after i=218014]:         in action cfabric.step when called from the environment:
~go[after i=218014]:             (internal) assumption                ← × 30+
~go[after i=218014]:             (internal) assumption
...
```

vs Python (`log.red:1908+`):

```
~py[after i=218014]:         in action cfabric.step when called from the environment:
~py[after i=218014]:             <IVY_EXAMPLES>/.../ord_live.ivy: line 998: assumption
~py[after i=218014]:             <IVY_EXAMPLES>/.../ord_live.ivy: line 1001: assumption
...
```

`grep -c '(internal) assumption' log.red` reports **1610**.

### Smoking gun: `isolate/create.go:858-865` `conjToAssume`

Python source-of-truth (`pyivy/ivy/ivy/ivy_isolate.py:1690-1693`):

```python
def conj_to_assume(c):
    res = ia.AssumeAction(c.formula)
    res.lineno = c.lineno      # ← propagates LF location
    return res
```

Go port (`isolate/create.go:858-865`):

```go
func conjToAssume(c *ast.LabeledFormula) actions.Action {
    fmla, ok := c.Formula.(lg.Expr)
    if !ok {
        return actions.NewSequence()
    }
    act := actions.NewAssumeAction(fmla)
    return act                  // ← MISSING: act.SetLineno(c.GetLineno())
}
```

`conjToAssume` is called from `apply_present_conjectures`
(`isolate/create.go:782-811`) once per conjecture per export action — for
the `ord_live` model that's roughly 30 conjs × 50+ export actions = ~1500
calls, plus the post-conj loop adds more. This single one-line omission
accounts for the bulk of the 1610 internals.

The remaining ~100 internals come from a small set of other constructor
sites that also omit `SetLineno`, listed in Step 2 below.

### Why this is the same shape as the previous fix, generalized further

Last two fixes addressed:
  1. `LabeledFormula` parser sites that produced empty `Loc`.
  2. `Action`/`ActionDef` compiler sites that did the same.

This fix addresses a **third** class: action *transformations* and
*derivations* (conjecture-to-assume, assert-strip, expand-while,
unroll-loops, etc.) where a fresh `Action` is constructed from a source
that *has* a Loc but the result fails to inherit it.

The user's directive: *"carry it along everywhere, and generalize the fix
so we do not have to revisit this 1609 more times."*

We satisfy that with a four-layer defense:

  1. **Surgical fix** (Step 1): patch the smoking gun (`conjToAssume`).
  2. **Audit fix** (Step 2): fix every other constructor site that omits
     `SetLineno`.
  3. **Safety-net walker** (Step 3): a `PropagateLocDown(action)` recursive
     walker that fills any nested action's empty Loc by inheritance from
     its enclosing parent. Idempotent, applied at strategic points in the
     pipeline.
  4. **Pipeline integration** (Step 4): call the safety-net at the end of
     the isolate classification loop and after compiler ARGSetup, so any
     forgotten `SetLineno` is recovered before the print site consumes it.

After this fix, the only way to produce an `(internal) assumption` is for
both the action *and* every ancestor on its path back to the root to have
empty Loc — which is impossible because the root action (registered in
`mod.Actions`) has its Loc set from the previous fix's defensive fallback
in `compiler/ivy_compile.go:IvyARGSetup`.

## Plan

### Step 1 — Surgical fix to `conjToAssume`

In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/create.go`
(lines 858-865), add the missing line:

```go
func conjToAssume(c *ast.LabeledFormula) actions.Action {
    fmla, ok := c.Formula.(lg.Expr)
    if !ok {
        return actions.NewSequence()
    }
    act := actions.NewAssumeAction(fmla)
    act.SetLineno(c.GetLineno())   // Python: res.lineno = c.lineno
    return act
}
```

This single line closes ~1500+ of the 1610 internals.

### Step 2 — Audit and fix other Loc-dropping constructor sites

Each fix is a one-line addition of `SetLineno` from the most natural source
already in scope. None require API changes.

| File | Line | Current code | Fix |
|---|---|---|---|
| `compiler/phase6.go` | 1986 | `return actions.NewAssertAction(a.Formula)` (`apply_assert_proofs` proof-stripping) | Wrap in a 3-line block: assign, `SetLineno(a.GetLineno())`, return |
| `compiler/phase6.go` | 1998 | `return actions.NewRequiresAction(a.Formula)` | Same pattern |
| `compiler/phase6.go` | 2009 | `return actions.NewEnsuresAction(a.Formula)` | Same pattern |
| `compiler/phase6.go` | 2020 | `return actions.NewSubgoalAction(a.Formula)` | Same pattern |
| `compiler/phase6.go` | 540 | `return actions.NewCrashAction(nil), nil` (in `CompileCrashAction`) | Wrap: assign, `SetLineno(node.GetLineno())`, return |
| `compiler/phase6.go` | 775 | `return actions.NewDebugAction(nil), nil` (in `CompileDebugAction`) | Same pattern |
| `compiler/ivy_compile.go` | 874 | `seq = append(seq, actions.NewAssumeAction(sym))` (scenario from-place processing) | Build the assume on its own line and `SetLineno(tr.GetLineno())` before append |
| `isolate/create.go` | 849-853 | `bracketActionInt`: `newAct := actions.NewSequence(parts...)` — wrapper has no Loc | Replace `actions.NewSequence(parts...)` body construction with `EmptyClone(act)` (already exists at `isolate/isolate.go:174-179`, sets Loc + formals) and append parts to its `Elems`. This mirrors Python's `thing = empty_clone(action); thing.args.extend(...)` exactly. |
| `actions/transforms.go` | 672-681 | `unrollWhile`: `NewAssumeAction(orExpr)`, `NewSequence(bodyExpr, res)`, `NewIfAction(a.Cond, seq)` — all without Loc | Add `.SetLineno(a.GetLineno())` to each. |
| `actions/match.go` | 498-519 | `expandWhile`: `NewAssumeAction(...)`, `NewSequence(thenParts...)`, `NewSequence()`, `NewIfAction(...)`, outer `NewSequence(...)` | Add `.SetLineno(w.GetLineno())` to each. |
| `vmt/vmt.go` | 127-128, 133-134 | Already calls `SetLineno(ast.Location{})` — explicitly empty | Leave as-is; vmt is a different output path and not affected by this divergence. |

### Step 3 — General safety-net: `PropagateLocDown` walker

Add a new function in `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transforms.go`:

```go
// PropagateLocDown walks an action tree and fills any nested action's
// empty Loc by inheritance from its enclosing parent. Idempotent and
// order-independent (mutates in place; SetLineno on a non-empty Loc is
// a no-op via the HasLoc invariant — we only set when GetLineno() returns
// the zero Location).
//
// This is the safety-net layer of the Loc-on-action discipline: even if
// some constructor in the codebase forgets to call SetLineno, this walker
// recovers a meaningful Loc by inheriting from the parent. Combined with
// the per-constructor fixes, it makes "(internal) <action>" output
// impossible whenever the enclosing action has any Loc at all.
//
// Used at the end of the isolate pipeline (before mod.Actions = newActions)
// and after compiler ARGSetup, so any forgotten SetLineno is recovered
// before downstream consumers (PrettyActionLineno) see it.
func PropagateLocDown(action Action) {
    if action == nil {
        return
    }
    propagateLocDownRec(action, action.GetLineno())
}

func propagateLocDownRec(action Action, parentLoc ast.Location) {
    if action == nil {
        return
    }
    if action.GetLineno() == (ast.Location{}) {
        action.SetLineno(parentLoc)
    }
    myLoc := action.GetLineno()
    if myLoc == (ast.Location{}) {
        myLoc = parentLoc
    }
    for _, sub := range action.IterSubactions() {
        if sub == nil || sub == action {
            continue
        }
        propagateLocDownRec(sub, myLoc)
    }
}
```

Notes:

- Uses `IterSubactions()` (already implemented for every Action type) so
  the walker is fully generic — no per-type handling.
- Mutates in place, since `SetLineno` is cheap and idempotent and we want
  the same Action instances visible to callers.
- The `if sub == action` guard handles `IterSubactions` implementations
  that include `self` as the first element (which `defaultIterSubactions`
  in `module/action.go` does).

### Step 4 — Apply the safety-net at strategic pipeline points

**4a.** In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/isolate.go`,
just before `mod.Actions = iu.NewInsMap[string, module.Action]()` at
line 1338, walk `newActions`:

```go
// Defensive Loc safety-net: every nested action should have a Loc by
// inheritance from its parent. Catches any constructor that forgot
// SetLineno and fixes it before downstream PrettyActionLineno sees it.
for _, act := range newActions.All() {
    actions.PropagateLocDown(act)
}
mod.Actions = iu.NewInsMap[string, module.Action]()
for name, act := range newActions.All() {
    mod.Actions.Set(name, act)
}
```

**4b.** In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/create.go`,
in `bracketActionInt` (around line 853) and at the end of
`ApplyPresentConjectures` (where brackets are applied to actions), call
`PropagateLocDown(newAct)` after each `mod.Actions.Set(actname, newAct)`
so the bracketed conj-assumes inherit Loc from the wrapping action if
their own (now-fixed) Loc is somehow still missing. This is belt-and-
suspenders: it catches *future* breakage in `conjToAssume` or its callers.

**4c.** In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/ivy_compile.go`,
after the existing defensive-fallback block in `IvyARGSetup` ActionDecl
case (around line 568, the block added in the previous plan), also call:

```go
actions.PropagateLocDown(action)
mod.SetAction(name, action)
```

This guarantees that every action entering `mod.Actions` from the
compiler has its full subtree Loc-populated.

### Step 5 — (Optional) Lint guard test

Add a small test in `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/loc_invariant_test.go`:

```go
// TestLocInvariantAfterIsolate compiles the ord_live example, runs
// isolate, and asserts that no nested action in mod.Actions has an
// empty Loc. Catches regressions where a future constructor change
// drops Loc without going through the per-constructor or safety-net
// fixes.
func TestLocInvariantAfterIsolate(t *testing.T) {
    mod := loadOrdLiveAndIsolate(t)
    for actname, act := range mod.Actions.All() {
        if a, ok := act.(actions.Action); ok {
            walk(a, func(sub actions.Action) {
                if sub.GetLineno() == (ast.Location{}) {
                    t.Errorf("nested action in %s has empty Loc: %T", actname, sub)
                }
            })
        }
    }
}
```

If `loadOrdLiveAndIsolate` is too heavy, point this at any minimal model
that has at least one conjecture and one export — the smoking gun in
`conjToAssume` only needs that to fire.

This test is *optional* — Steps 1-4 are sufficient for the immediate fix.
Step 5 is the long-term regression guard.

### Step 6 — Verify

```
cd ~/ivy/goivy && go build ./...
cd ~/ivy/goivy && go test ./check/... ./compiler/... ./isolate/... ./actions/... ./module/...
cd ~/ivy/goivy && make golden
grep -c '(internal) assumption' ~/ivy/goivy/log.red    # expected: 0
grep -c '(internal)'             ~/ivy/goivy/log.red    # expected: 0 (or close to)
```

Spot-check the previously-broken cfabric.step block:

```
sed -n '/in action cfabric.step when called from the environment/,/^~go.*in action/p' ~/ivy/goivy/log.red | head -40
```

Every `~go` line should now match the corresponding `~py` line modulo
xtrace indices. No `(internal) assumption` should appear in the output.

If any `(internal) assumption` *does* remain, the safety-net `PropagateLocDown`
is intentionally lenient: it only inherits when a parent has Loc. If both
parent and child are empty, the walker can't help. In that case, trace
back to the root action: which `mod.Actions[name]` has empty Loc? That
means the IvyARGSetup defensive-fallback failed to fire, which means
`ad.GetLineno()` itself was empty — which would point at a missing
parser-side `SetLineno` on the `ActionDef` (a separate, narrower fix).

## Critical files

- `/Users/jaten/ivy/goivy/log.red` — current divergence reference (read-only)
- `/Users/jaten/ivy/goivy/isolate/create.go`
  - `conjToAssume` at lines 858-865 (Step 1 — surgical fix)
  - `bracketActionInt` at lines 835-854 (Step 2 — use `EmptyClone`)
  - `ApplyPresentConjectures` at lines 782-819 (call site that exercises `conjToAssume`)
- `/Users/jaten/ivy/goivy/compiler/phase6.go`
  - `apply_assert_proofs` proof-stripping at lines 1986, 1998, 2009, 2020 (Step 2)
  - `CompileCrashAction` at line 540 (Step 2)
  - `CompileDebugAction` at line 775 (Step 2)
- `/Users/jaten/ivy/goivy/compiler/ivy_compile.go`
  - Scenario `NewAssumeAction(sym)` at line 874 (Step 2)
  - `IvyARGSetup` ActionDecl handler around line 568 (Step 4c — invoke `PropagateLocDown`)
- `/Users/jaten/ivy/goivy/actions/transforms.go`
  - `unrollWhile` at lines 672-681 (Step 2)
  - **Add `PropagateLocDown` and `propagateLocDownRec`** at the end of the file (Step 3)
- `/Users/jaten/ivy/goivy/actions/match.go`
  - `expandWhile` at lines 498-519 (Step 2)
- `/Users/jaten/ivy/goivy/isolate/isolate.go`
  - Just before line 1338 (`mod.Actions = ...`) — invoke `PropagateLocDown` (Step 4a)
  - `EmptyClone` helper at lines 174-179 (no changes; reused by Step 2 `bracketActionInt` fix)
- `/Users/jaten/ivy/goivy/check/isolate_check.go`
  - `PrettyActionLineno` consumer at lines 436, 520 (no changes; verifies the fix worked)
- `/Users/jaten/ivy/goivy/check/helpers.go`
  - `PrettyActionLineno` at lines 58-71 (no changes; the consumer)
- `/Users/jaten/ivy/goivy/module/action.go`
  - `ActionBase` (`Loc`, `HasLoc`, `GetLineno`, `SetLineno`) at lines 40-57 (no changes; the API surface used by the safety-net walker)
  - `defaultIterSubactions` (referenced by `IterSubactions` on every Action type) at the end of the file (no changes; the walker uses it)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_isolate.py`
  - `conj_to_assume` at lines 1690-1693 (Python source-of-truth that does set `res.lineno`)
  - `apply_present_conjectures` at lines 1706-1741 (caller; same loop structure as Go)
  - `bracket_action_int` at lines 1695-1700 (uses `empty_clone`)

## Verification

1. Build cleanly:
   ```
   cd ~/ivy/goivy && go build ./...
   ```

2. Targeted regression tests:
   ```
   cd ~/ivy/goivy && go test ./check/... ./compiler/... ./isolate/... ./actions/... ./module/...
   ```

3. Run the full conformance harness:
   ```
   cd ~/ivy/goivy && make golden
   ```

4. The key acceptance test — the count of `(internal) assumption` lines
   should drop from 1610 to 0:
   ```
   grep -c '(internal) assumption' ~/ivy/goivy/log.red
   ```
   Expected: `0`

5. The `~go` and `~py` lines in the
   "The following program assertions are treated as assumptions:" section
   should match modulo xtrace indices:
   ```
   sed -n '/program assertions are treated as assumptions/,/^~go.*The following program/p' ~/ivy/goivy/log.red
   ```

6. Sanity-check that the previously-fixed sections (impls, monitors,
   inits) still work:
   ```
   grep '(internal)' ~/ivy/goivy/log.red | head -20
   ```
   Expected: empty, or only entries that genuinely have no source location
   (e.g., synthesized initializers that have no source line at all).
