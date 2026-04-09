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
so we do not have to revisit this 1609 more times."* — and the follow-up:
*the generalized walker should be a debug-only **assertion**, not a
silent fix-it pass; once it has been confirmed not to fire, it can be
turned off.*

We satisfy that with a four-layer defense, the last two of which are
**diagnostic assertions, not silent fixers**:

  1. **Surgical fix** (Step 1): patch the smoking gun (`conjToAssume`).
  2. **Audit fix** (Step 2): fix every other constructor site that omits
     `SetLineno`.
  3. **Diagnostic walker** (Step 3): an `AssertEveryActionHasLoc(action,
     where string)` recursive walker that **panics** (or `t.Fatal`s in
     test mode) when it finds a nested action whose `Loc` is empty. The
     walker is gated by a per-session config flag (default OFF) so the
     hot path isn't paid in production but the check can be flipped on
     for debugging or pre-merge validation.
  4. **Pipeline checkpoints** (Step 4): when the flag is on, invoke the
     assertion at strategic points (end of compiler ARGSetup, end of
     isolate classification, end of `ApplyPresentConjectures`) so any
     missed Loc transfer is detected with the most precise context (which
     pipeline stage produced the bad node, and which root action it lives
     under).

The intent: run the assertion-on build once after Steps 1-2 land, confirm
zero panics, then leave the flag OFF in production. If a future change
regresses Loc handling, flip the flag and the panic message will name
the offender directly.

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

### Step 3 — Diagnostic assertion walker: `AssertEveryActionHasLoc`

Add a new function in `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transforms.go`:

```go
// AssertEveryActionHasLoc recursively walks an action tree and panics
// if any nested action has an empty Loc. The `where` argument is a
// human-readable context label (e.g., "isolate.classify_loop after
// AddMixinsExt actname=cfabric.step") that is included in the panic
// message so the failing pipeline stage is obvious.
//
// This is a DIAGNOSTIC tool, not a silent fix-it pass. It is gated by
// the package-level flag AssertLocEnabled (defined below) which defaults
// to false. Production runs should leave the flag off; this walker
// is intended to be flipped on temporarily to validate that all
// constructor sites preserve Loc, and then flipped off again when the
// codebase is clean.
//
// The walker reports the FIRST missing-Loc node it finds with:
//   - The action's Go type name
//   - The action's Sexp() (truncated to ~120 chars)
//   - The full context label
//   - The path of enclosing parent action types
// so the offending pipeline stage and constructor are immediately
// identifiable from the panic stack.
func AssertEveryActionHasLoc(action Action, where string) {
    if !AssertLocEnabled || action == nil {
        return
    }
    var path []string
    assertEveryActionHasLocRec(action, where, &path)
}

func assertEveryActionHasLocRec(action Action, where string, path *[]string) {
    if action == nil {
        return
    }
    typeName := iu.ShortTypeName(action)
    *path = append(*path, typeName)
    defer func() { *path = (*path)[:len(*path)-1] }()

    if action.GetLineno() == (ast.Location{}) {
        sx := string(action.Sexp())
        if len(sx) > 120 {
            sx = sx[:120] + "..."
        }
        panic(fmt.Sprintf(
            "AssertEveryActionHasLoc: missing Loc on %s at %s\n  path: %s\n  sexp: %s",
            typeName, where, strings.Join(*path, " > "), sx,
        ))
    }

    for _, sub := range action.IterSubactions() {
        if sub == nil || sub == action {
            continue
        }
        assertEveryActionHasLocRec(sub, where, path)
    }
}

// AssertLocEnabled gates AssertEveryActionHasLoc. Default false so
// production builds pay zero cost. Flip to true (via go test, a build
// tag, or a config setter) when validating Loc-propagation invariants.
var AssertLocEnabled = false
```

Notes:

- Lives at the package level for now (a global `var`) — a clear exception
  to CLAUDE.md rule C because it is a debug-only diagnostic, not mutable
  production state. Document this exemption in a comment alongside the
  flag. (If multi-tenant correctness becomes a concern later, the flag can
  be moved to `module.Config` with a small refactor; for now the
  simplicity matters more.)
- `iu.ShortTypeName` already exists; reuse it.
- `IterSubactions` is already implemented for every Action type.
- `Sexp()` is on the `lg.Expr`/`Action` interface and gives a deterministic
  short form of the action.
- The walker is read-only — no `SetLineno` calls. It diagnoses, does not
  fix.

### Step 4 — Wire the assertion at strategic pipeline checkpoints

When `AssertLocEnabled` is on, validate the Loc invariant at every point
where actions cross a pipeline stage boundary. Each call site passes a
unique `where` label so the panic message identifies the failing stage.

**4a.** In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/ivy_compile.go`,
inside the `IvyARGSetup` ActionDecl handler around line 567, after the
existing defensive fallback block (added in the previous plan):

```go
if action.GetLineno() == (ast.Location{}) {
    action.SetLineno(ad.GetLineno())
}
actions.AssertEveryActionHasLoc(action, "compiler.IvyARGSetup actname="+name)
mod.SetAction(name, action)
```

**4b.** In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/isolate.go`,
just before the existing `mod.Actions = iu.NewInsMap[...]()` at line 1338:

```go
for actname, act := range newActions.All() {
    actions.AssertEveryActionHasLoc(act, "isolate.end_classify actname="+actname)
}
mod.Actions = iu.NewInsMap[string, module.Action]()
for name, act := range newActions.All() {
    mod.Actions.Set(name, act)
}
```

**4c.** In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/create.go`,
inside `bracketActionInt` (around line 853), after `mod.Actions.Set(actname, newAct)`:

```go
mod.Actions.Set(actname, newAct)
actions.AssertEveryActionHasLoc(newAct, "isolate.bracketActionInt actname="+actname)
```

**4d.** Also at the end of `ApplyPresentConjectures` (around the return at
line 818), iterate the brackets returned and call the assertion on each
bracket's `Before`/`After` slices, with `where = "isolate.ApplyPresentConjectures.bracket actname="+e.ActName`. This catches the
exact bug we're fixing (`conjToAssume` dropping Loc) at its source instead
of waiting for the downstream consumer.

**4e.** In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/isolate_check.go`,
just before the print loop at line 400 (`for actname, action := range mod.Actions.All()`),
add a debug-only assertion sweep so the print stage is the last gate:

```go
if actions.AssertLocEnabled {
    for actname, action := range mod.Actions.All() {
        if a, ok := action.(actions.Action); ok {
            actions.AssertEveryActionHasLoc(a, "check.isolate_check.print actname="+actname)
        }
    }
}
```

### How to run with assertions on

There are three options for flipping `AssertLocEnabled`. Pick one:

1. **Test-only `init`** in a small `*_assertloc_test.go` file:
   ```go
   func init() { actions.AssertLocEnabled = true }
   ```
   gated by a build tag like `assertloc`. Then run:
   ```
   go test -tags assertloc ./...
   ```
2. **Manual flip in main**: in `cmd/.../main.go` (or wherever the goivy
   binary is wired), gated by a `--assert-loc` CLI flag. Then run:
   ```
   make golden ASSERT_LOC=1
   ```
3. **Direct flip from a one-shot test** that imports `actions`, sets the
   flag, and exercises `make golden`-equivalent code paths. This is the
   smallest-overhead option.

Whichever is chosen, after Steps 1-2 land, run with assertions ON once to
confirm zero panics, then leave the flag at its default `false` so
production runs are unaffected.

### Step 5 — Validation pass with assertions enabled

After Steps 1-4 land:

1. Build with assertions enabled (whichever mechanism from Step 4's "How
   to run with assertions on" was chosen).
2. Run `make golden` (or the equivalent direct test invocation).
3. Expected result: zero panics from `AssertEveryActionHasLoc`. The
   validation has succeeded.
4. If a panic *does* fire, the message names the failing pipeline stage,
   the action's Go type, its Sexp, and the path of enclosing parents —
   from that, identify the constructor that omitted `SetLineno` and add
   it to Step 2's table; then re-run.

Iterate Steps 2 and 5 until the assertion-enabled run is panic-free.

Once panic-free, leave `AssertLocEnabled = false` in production. The flag
remains in the codebase as a debugging tool that any future contributor
can flip on if they suspect a Loc regression.

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

If any `(internal) assumption` *does* remain after Steps 1-2, that means a
constructor site was missed in the audit. Flip `AssertLocEnabled` on
(Step 4's mechanism) and re-run; the panic message will identify the
offender by pipeline stage, action type, and Sexp. Add the missing
`SetLineno` and repeat until clean.

After validation: revert `AssertLocEnabled` to `false` (or delete the
test-only `init` flip from Step 4), so production runs pay zero cost.
The walker stays in the codebase as a one-line-flip debugging tool for
future regressions.

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
