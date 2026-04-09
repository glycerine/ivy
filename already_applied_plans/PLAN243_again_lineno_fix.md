# Plan: Carry Location through Action compilation (fixes "(internal)" + bare-line-number for impls/monitors/inits)

Created: 2026-04-09 14:30

## Context

After the previous Loc-propagation fix for `LabeledFormula`, `make golden` still
shows divergence in `~/ivy/goivy/log.red` starting at line 152, where the
"The following action implementations are present:" / monitors / initializers
sections print broken locations on the Go side only:

```
~go[after i=218014]:     The following action implementations are present:
~go[after i=218014]:         (internal) implementation of arm.issue_hook        ← Loc empty
~go[after i=218014]:         992implementation of arm.step_north                ← Loc has Line, no formatting
~go[after i=218014]:         726implementation of ifabric.step
~go[after i=218014]:         (internal) implementation of memc.memc_arr_hook
...
```

vs Python (log.red:1842):

```
~py[after i=218014]:         <IVY_EXAMPLES>/doc/examples/apple/ord_live.ivy: line 307: implementation of arm.issue_hook
~py[after i=218014]:         <IVY_EXAMPLES>/doc/examples/apple/ord_live.ivy: line 312: implementation of arm.step_north
```

There are two distinct symptoms with one common print site (and a deeper
upstream cause for the `(internal)` cases):

### Symptom A — bare line number with no filename and no formatting

`check/isolate_check.go:230-238` (impls), `:249-257` (monitors), and
`:268-277` (initializers) format the lineno as `fmt.Sprintf("%d", loc.Line)`
when `loc.Line > 0`. This:

  1. Throws away `loc.Filename` even when it is set.
  2. Throws away `loc.Reference` (the original location for instantiated
     actions). For example, `arm.step_north`'s Loc has `Line=992` (the
     instantiation site `instantiate arm : arm_mod` at `ord_live.ivy:992`)
     and `Reference` pointing at the original action body (line 312). The
     `loc.String()` method already follows `Reference` correctly
     (`ast/ast.go:35-47`), so calling `String()` instead of `%d` fixes the
     wrong-line and missing-filename problems in one shot.
  3. Has no trailing `: ` separator, producing the run-together
     "992implementation of arm.step_north".

Python's source-of-truth `pretty_lineno` (`pyivy/ivy/ivy/ivy_check.py:259-260`)
just calls `str(ast.lineno)`, which goes through `LocationTuple.__str__`
producing the full `<file>: line N: ` form including reference resolution.
Go's `Location.String()` already mirrors this exactly — the print site just
isn't using it.

### Symptom B — `(internal)` because the action has no Loc at all

For forward-declared hooks like `action issue_hook(p:proc,lt:lclock)` (no
body), `compiler/action.go:42-46` returns `actions.NewSequence()` with no
`SetLineno`. So `mod.Actions["arm_mod.issue_hook"]` (and after instantiation,
`mod.Actions["arm.issue_hook"]`) carries an empty `ActionBase.Loc`.

Then `implement issue_hook { ... }` is parsed by `handleBeforeAfter`
(`parser/grammar_v17.y:263-272`), which builds an `ActionDef` whose own Loc
is set from `atom.GetLineno()`. But when `CompileAction`
(`compiler/action.go:38-200`) compiles the implement body, the result branch
at line 173 (`case actions.Action: body = v`) and line 179
(`default: body = actions.NewSequence()`) does **not** call `SetLineno` on
the body. Only the `*lg.And` branch at line 175-178 does. So the compiled
implement body usually has no Loc.

In `isolate.go:372`, `actions.ApplyMixin(mixer, action, false)` does
`res.SetLineno(action1.GetLineno())` (`actions/helpers.go:195`) — propagating
the *mixer's* Loc to the merged result. With an empty mixer Loc, the merged
implementation also has empty Loc, and the print site falls into the
`(internal)` branch.

### Why this is the same shape as the previous fix, generalized

Last fix: parser produced `LabeledFormula` with no Loc → `pretty_lineno`
printed `(internal)`. Fix was to set `Loc` on the LF at construction sites
and to remove the redundant `Lineno int` field, making `Base.Loc` the single
source of truth.

This fix is the same pattern at the *action* layer:

  1. The print site reads `Loc` and assumes it carries the formatted form;
     fix the print site to actually use `Location.String()`.
  2. Some construction sites for compiled actions never call `SetLineno`,
     so the assumption that "if I'm here, the Action has a Loc" is wrong.
     Fix the construction sites.

CLAUDE.md rule A: Python is the source of truth.
Python: `compile_action_def` (`pyivy/ivy/ivy/ivy_compiler.py:933-966`)
asserts `hasattr(res,'lineno')` after `sortify(a.args[1])` — i.e. the body
*must* have a lineno. We mirror this guarantee in Go by ensuring the
compiled body's Loc is set, falling back to `node.GetLineno()` (the
`ActionDef`'s Loc, which the parser already sets via `handleBeforeAfter`)
when needed.

## Plan

### Step 1 — Add `PrettyActionLineno` helper and use it at the three print sites

**1a.** In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/helpers.go`,
add a helper next to the existing `PrettyLineno` (line 39-48):

```go
// PrettyActionLineno formats an action's location for display.
// Mirrors PrettyLineno (for *ast.LabeledFormula) and matches Python's
// pretty_lineno which calls str(ast.lineno) when present.
// Returns "(internal) " when the action has no Location set.
// Uses Location.String() so the Reference chain (set during module
// instantiation by LinenoAddRef) is followed correctly.
func PrettyActionLineno(a actions.Action) string {
    if a == nil {
        return "(internal) "
    }
    loc := a.GetLineno()
    if loc.Filename != "" || loc.Line > 0 {
        return loc.String()
    }
    return "(internal) "
}
```

**1b.** In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/isolate_check.go`,
replace the three inline `lineno := "(internal) "; if impl.Action != nil { ... fmt.Sprintf("%d", loc.Line) ... }` blocks:

  - Lines 230-238 (impls):
    ```go
    for _, impl := range impls {
        fmt.Printf("        %simplementation of %s\n", PrettyActionLineno(impl.Action), impl.Mixee)
    }
    ```
  - Lines 249-257 (monitors):
    ```go
    for _, mon := range mons {
        fmt.Printf("        %smonitor of %s\n", PrettyActionLineno(mon.Action), mon.Mixee)
    }
    ```
  - Lines 268-277 (initializers):
    ```go
    for _, na := range inits {
        fmt.Printf("        %s%s\n", PrettyActionLineno(na.Action), na.Name)
    }
    ```

After Step 1 alone, every action that *does* have a Loc will print correctly,
including the instantiated ones that currently show "992": the
`Location.String()` recursion through `Reference` will produce
"<file>: line 312: " (the original body line), matching Python.

The `(internal)` cases will still show `(internal)` after Step 1 — those
require Step 2 to actually populate the Loc upstream.

### Step 2 — Carry source Location through `CompileAction`

In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/action.go`:

**2a.** Empty-body forward declaration (lines 42-46). Currently:
```go
if node.Body == nil {
    seq := actions.NewSequence()
    seq.FormalParams = nil
    seq.FormalReturns = nil
    return seq, nil
}
```
Replace with:
```go
if node.Body == nil {
    seq := actions.NewSequence()
    seq.SetLineno(node.GetLineno())   // ← carry forward ActionDef Loc
    seq.FormalParams = nil
    seq.FormalReturns = nil
    return seq, nil
}
```

**2b.** Three-arm switch at lines 170-181. Currently:
```go
var body actions.Action
switch v := sortResult.(type) {
case actions.Action:
    body = v
case *lg.And:
    seq := actions.NewSequence(v.Terms...)
    seq.SetLineno(bodyToCompile.GetLineno())
    body = seq
default:
    body = actions.NewSequence()
}
```
After this switch, add a Loc-propagation guard that mirrors Python's
`assert hasattr(res,'lineno')`:
```go
// Python compile_action_def asserts hasattr(res,'lineno') after sortify.
// Mirror that guarantee: ensure body carries a source Location, preferring
// the body node's own Loc when set (matches Python's res.lineno from
// sortify), and falling back to the ActionDef's Loc (set by handleBeforeAfter
// or the top-level action production in grammar_v17.y).
if body.GetLineno() == (ast.Location{}) {
    if bl := bodyToCompile.GetLineno(); bl.Filename != "" || bl.Line > 0 {
        body.SetLineno(bl)
    } else {
        body.SetLineno(node.GetLineno())
    }
}
```

This handles all three switch arms uniformly. The `*lg.And` arm already sets
`bodyToCompile.GetLineno()` and will pass through the guard unchanged. The
`actions.Action` arm and `default` arm will get their Loc set either from
`bodyToCompile` (when present) or from the `ActionDef`.

**2c.** Fallback paths in the same function: lines 130-135 and 161-167 both
return `fallback := actions.NewSequence()` for compile-error cases. Add
`fallback.SetLineno(node.GetLineno())` to each so error fallback sequences
also carry a Loc.

### Step 3 — Audit auxiliary action constructors that drop Loc

These are not the main offenders for the visible regression but they are
the same shape and will eventually cause the same class of bug. Fix together
to make Loc-on-action a maintained invariant:

**3a.** `compiler/action.go:19-34` `thingAction`. The default branch at
line 32 returns `actions.NewSequence()` with no Loc; add
`seq := actions.NewSequence(); seq.SetLineno(node.GetLineno()); return seq, nil`.
The `*lg.And` branch already sets it correctly.

**3b.** `compiler/action.go:558-588` `*ast.Sequence` case in `CompileActionBody`.
The line-572 `return seq, nil` branch returns the sequence unchanged when it
already had an action result; add a fallthrough check `if seq.GetLineno() == (ast.Location{}) { seq.SetLineno(node.GetLineno()) }`. The other branches already set it.

**3c.** `parser/grammar_v17.y:1306-1331`, the `top TOK_AROUND ...` production.
Currently:
```go
bdf := acfg(v17lex).NewActionDef(bmixer, before, $4, $5)
bdecl := acfg(v17lex).NewActionDecl(bdf)
```
and similarly for `adf`. Neither `bdf` nor `adf` has `SetLineno` called on
it. Add:
```go
loc := tokLineno(v17lex.(*v17LexAdapter), $2)  // TOK_AROUND
bdf.SetLineno(loc)
...
adf.SetLineno(loc)
```
This is a parser site so the generated `parser/grammar_v17.go` must be
regenerated via `go generate ./parser/` (do not hand-edit the generated
file).

**3d.** `compiler/ivy_compile.go:540-570` (`IvyARGSetup`/`ActionDecl` case).
After `action, err := as.Compiler.CompileAction(ad)`, add a defensive
fallback (in case some CompileAction code path still drops Loc):
```go
if action.GetLineno() == (ast.Location{}) {
    action.SetLineno(ad.GetLineno())
}
```
This is belt-and-suspenders so `mod.SetAction(name, action)` is guaranteed
to register an Action with Loc set, regardless of which CompileAction path
ran.

### Step 4 — Verify

```
cd /Users/jaten/ivy/goivy && go generate ./parser/ && go build ./...
cd /Users/jaten/ivy/goivy && make golden
```

Then inspect `~/ivy/goivy/log.red`:

```
sed -n '150,200p' ~/ivy/goivy/log.red
```

Expected: every line in the "The following action implementations are
present:", "The following action monitors are present:", and "The following
initializers are present:" sections matches the corresponding `~py[...]`
line modulo XTRACE indices. No `(internal)` and no bare-number prefix
should remain in those three sections.

Spot-check the previously-broken lines:

```
grep -n 'implementation of arm\.' ~/ivy/goivy/log.red
grep -n 'implementation of cfabric\.' ~/ivy/goivy/log.red
grep -n 'monitor of lclock' ~/ivy/goivy/log.red
```

The `~go` and `~py` outputs should now contain matching `<file>: line N: `
prefixes for each entry. (Modulo any pre-existing reference-line divergence
that is *not* in scope here — line numbers should match Python because
`Location.String()` follows the `Reference` chain identically.)

Run the broader regression suite to confirm no test breakage:

```
cd /Users/jaten/ivy/goivy && go test ./check/... ./compiler/... ./parser/... ./isolate/... ./module/...
```

If a test in `check/isolate_check_test.go` was previously asserting on the
"(internal)" or bare-number form, update it to assert on the
`<file>: line N:` form (the new test should match what Python prints).

## Critical files

- `/Users/jaten/ivy/goivy/log.red` — current divergence reference (read-only)
- `/Users/jaten/ivy/goivy/check/helpers.go` — add `PrettyActionLineno` next to existing `PrettyLineno` (line 39-48)
- `/Users/jaten/ivy/goivy/check/isolate_check.go` — three print sites at lines 225-278
- `/Users/jaten/ivy/goivy/compiler/action.go`:
  - `CompileAction` empty-body branch at lines 42-46
  - `CompileAction` post-Sortify branch at lines 170-181
  - `CompileAction` fallback branches at lines 130-135 and 161-167
  - `thingAction` at lines 19-34
  - `CompileActionBody` `*ast.Sequence` case at lines 558-588
- `/Users/jaten/ivy/goivy/compiler/ivy_compile.go` — defensive fallback after `CompileAction` at lines 555-567
- `/Users/jaten/ivy/goivy/parser/grammar_v17.y` — `top TOK_AROUND` production at lines 1306-1331 (regenerate `grammar_v17.go` via `go generate ./parser/`)
- `/Users/jaten/ivy/goivy/parser/generate.go` — `//go:generate goyacc -o grammar_v17.go -p v17 grammar_v17.y`
- `/Users/jaten/ivy/goivy/ast/ast.go` — `Location.String()` at lines 35-47 (no changes; reference for the recursion semantics that make Step 1 work)
- `/Users/jaten/ivy/goivy/ast/config.go` — `LinenoAddRef` at lines 64-76 (no changes; reference for how the `Reference` field is set during instantiation)
- `/Users/jaten/ivy/goivy/module/action.go` — `ActionBase` (`Loc`, `HasLoc`, `GetLineno`, `SetLineno`) at lines 40-57 (no changes; the API surface used by the fix)
- `/Users/jaten/ivy/goivy/actions/helpers.go` — `ApplyMixin` at lines 141-203 (no changes; verifies that `res.SetLineno(action1.GetLineno())` at line 195 propagates the mixer's Loc, so once mixer has a Loc the merged action will too)
- `/Users/jaten/ivy/goivy/isolate/isolate.go` — implementation/monitor population at lines 372-376, 561-564, 580-587 (no changes; verifies that the actions stored in `IsolateInfo` are the same instances that `CompileAction` returned, so Step 2 propagates through transitively)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_check.py` — `pretty_lineno` at lines 259-260 and the impl/monitor/init print sites at lines 587-607 (Python source of truth)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_compiler.py` — `compile_action_def` at lines 933-966 (Python source of truth — note the `assert hasattr(res,'lineno')` at line 954 which Step 2 mirrors)

## Verification

1. Regenerate parser and build:
   ```
   cd /Users/jaten/ivy/goivy && go generate ./parser/ && go build ./...
   ```

2. Run the conformance harness:
   ```
   cd /Users/jaten/ivy/goivy && make golden
   ```

3. Inspect `~/ivy/goivy/log.red`. The block starting at the line that says
   "The following action implementations are present:" should contain
   `<IVY_EXAMPLES>/doc/examples/apple/ord_live.ivy: line N: implementation of <name>`
   for every entry, matching the corresponding `~py` block. No `(internal)`
   and no bare-number prefix in the impls/monitors/inits sections.

4. Run the broader test suite:
   ```
   cd /Users/jaten/ivy/goivy && go test ./check/... ./compiler/... ./parser/... ./isolate/... ./module/...
   ```

5. Sanity grep for any remaining `fmt.Sprintf("%d", loc.Line)` or
   `fmt.Sprintf("%d", .*\.Line)` patterns in `check/`:
   ```
   grep -rn 'fmt.Sprintf("%d", .*\.Line' check/
   ```
   Should be empty (or any remaining hits should be unrelated to the
   pretty-print path).
