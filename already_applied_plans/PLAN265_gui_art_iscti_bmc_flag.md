# Plan: Wire BMC `some_bounded` flag and implement `GuiArt`

Created: 2026-04-11 17:30 UTC

## Context

Two TODOs were deferred from earlier plans:

1. **`check/check.go:1244`** has `_ = someBounded // TODO: wire BMC flag`. The local
   `someBounded := false` exists in both `Start()` (line 1161) and `StartWithConfig()`
   (line 1244), and both functions print `"BOUNDED"` if the flag is true (lines 1196,
   1278). The flag is *never* set anywhere, so the `"BOUNDED"` line is dead. Python
   sets a global `some_bounded = True` in `ivy_check.py:996` whenever an isolate uses
   the `bmc[N]` method.

2. **`check/phase7.go:19-31`** is a 13-line stub for `GuiArt`. Python's `gui_art`
   (`ivy_check.py:86-102`) does:
   - Create a fresh `AnalysisGraph` and add the initial state if `default_ui == "art"`
   - Run the `initialize` action against the graph
   - Hand the graph to a Tk-based UI (`tk_ui.new_ui`)
   - Block in `gui.tk.mainloop()` and `exit(1)` when the user closes

   The Go stub does none of this. The previous plan added `MatchHandler.IsCti
   *module.Clauses` (`check/helpers.go:147-150`) "as preparation for `GuiArt`",
   set at `check.go:602`, but it is read by no one — the field is currently dead.

   The Go ports of *every* Python `gui_art` call site are also stubbed:
   - `check.go:605-613` (trace path, after L2S/property failure): just prints text
   - `check.go:876-883` (`DisplayCex`, Python `display_cex`): returns an error
   - `check.go:893-924` (`ShowCounterexample`, Python `show_counterexample`): just prints text
   - `isolate_check.go:793-801` and `:849-857` (`CheckSubgoals` failure paths,
     temporal and non-temporal): just return the error without honoring `opt_trace`,
     `goal.trace_hook`, or `diagnose.get()` — the entire Python block at lines
     829-835 is missing.

This plan fixes both TODOs and the cluster of related GUI-stubs that share the
same root cause (no `GuiArt` to delegate to). The actual UI rendering (the
webui-side hook implementation) is genuinely a separate project and is left as a
clean extension point.

## Approach

Five sequential steps. Each leaves the build green.

### Step 1 — Wire the BMC `someBounded` flag

**1a. Add the field.** In `module/config.go`, after the `Failures` field, add:

```go
// SomeBounded tracks whether any isolate in this session used the BMC
// (bounded model checking) verification method. When true, Start prints
// "BOUNDED" before "OK". Mirrors Python's module-level `some_bounded`
// global (ivy_check.py:996, 1029, 1046).
SomeBounded bool
```

Field name `SomeBounded` mirrors Python's `some_bounded` per the CLAUDE.md mechanical
port rule.

**1b. Set the flag in the BMC branch.** In `check/isolate_check.go`, at line 997
(start of the `case strings.HasPrefix(methodName, "bmc[")` branch), insert:

```go
case strings.HasPrefix(methodName, "bmc["):
    // Python ivy_check.py:995-996:
    //   global some_bounded
    //   some_bounded = True
    // We write to the PARENT mod.Cfg, NOT isoMod.Cfg, because Module.Copy
    // (module.go:337) does *c.Cfg = *m.Cfg, so writes to isoMod.Cfg are
    // discarded when the loop iteration ends.
    mod.Cfg.SomeBounded = true
    nSteps, nUnroll, err := parseBMCParams(methodName)
    ...
```

**Critical detail (caught during validation):** writing to `isoMod.Cfg.SomeBounded`
would be silently lost because `Module.Copy()` value-copies the Config. The parent
`mod` is in scope inside the loop body (it's the function parameter to `CheckModule`,
declared at `isolate_check.go:884`).

**1c. Reset and read the flag in entry points.** In `check/check.go`:

`Start()` (lines 1161-1207):
```go
// before
someBounded := false
...
// Python: if some_bounded: print("BOUNDED")
if someBounded {
    fmt.Println("BOUNDED")
}

// after
// Python ivy_check.py:1028-1029: some_bounded = False at start() entry
mod.Cfg.SomeBounded = false   // (placed AFTER mod construction)
...
// Python ivy_check.py:1046-1047: if some_bounded: print("BOUNDED")
if mod.Cfg.SomeBounded {
    fmt.Println("BOUNDED")
}
```

Note: in `Start()`, the reset line goes *after* `mod := module.New()` and the
`mod.Cfg = module.NewConfig()` setup (currently lines 1163-1166), so that the
fresh Config exists when we reset. Since `NewConfig()` already zero-initializes
the field, the explicit reset is technically redundant but is preserved for
faithfulness with Python and for safety against config-reuse scenarios.

`StartWithConfig()` (lines 1244-1290):
```go
// before
someBounded := false
_ = someBounded // TODO: wire BMC flag
...
if someBounded {
    fmt.Println("BOUNDED")
}

// after — drop the local entirely
// Python ivy_check.py:1028-1029: some_bounded = False at start() entry
// (For StartWithConfig, the caller may reuse a Config across runs, so we
//  must reset rather than relying on zero-init.)
cfg.SomeBounded = false       // (placed BEFORE mod := module.New())
...
if mod.Cfg.SomeBounded {
    fmt.Println("BOUNDED")
}
```

The reset must happen before `CheckModule` runs and before the existing `mod.Cfg =
cfg` assignment so that the user-supplied config doesn't leak stale state.

**1d. Add tests** in `check/check_test.go`:

- `TestSomeBoundedZeroByDefault` — `cfg := module.NewConfig(); if cfg.SomeBounded {...}`
- `TestSomeBoundedSetByBMCBranch` — Construct a minimal isolate with `method = "bmc[1]"`,
  invoke the dispatch (or call into `CheckModule` with a stub that just trips the
  branch), assert `mod.Cfg.SomeBounded == true` afterwards.
- `TestSomeBoundedNotSetForOtherMethods` — Same setup with `method = "mc"` and
  default (induction), assert `SomeBounded == false`.
- `TestSomeBoundedResetByStartWithConfig` — Set `cfg.SomeBounded = true` manually,
  call `StartWithConfig` against a no-op `.ivy` file, assert it gets reset.

Do NOT test the `"BOUNDED"` print line directly via stdout capture — the bool
test above is sufficient and stdout interception in `Start()` is brittle.

After Step 1: `go build`, `go vet`, `go test ./check/...` all pass. The TODO
comment is gone. `BOUNDED` now actually prints when a BMC isolate runs.

### Step 2 — Add the `GuiArtHook` callback hook point

**2a. Define the hook type** in `module/config.go` (or a new `module/gui_hook.go`
if preferred). All three referenced types — `*Module`, `interface{}`, and
`*Clauses` — already live in `module/`, so the type can be a fully-typed `func`
without any import cycle and without `interface{}` boxing:

```go
// GuiArtHook is the concrete type for the analysis-graph GUI hook stored on
// Config.GuiArtHook. It is invoked by check.GuiArt to display an analysis
// graph in an interactive UI.
//
// The `target` argument is interface{} to mirror Python's gui_art polymorphism:
// callers pass either an *art.AnalysisGraph (from the ShowCounterexample /
// DisplayCex paths) or a *check.MatchHandler (from the trace failure path).
// We cannot name those types here because module cannot import art or check
// (cycle), so target stays interface{} and the hook implementation type-switches.
//
// The `isCti` argument carries the failing-conjecture clauses captured by
// check.MatchHandler.IsCti, or nil for non-CTI counterexamples.
//
// The hook is responsible for any blocking UI loop and may call os.Exit if
// it wishes to mirror Python's `exit(1)` at the end of gui_art
// (ivy_check.py:102).
type GuiArtHook func(mod *Module, target interface{}, isCti *Clauses) error
```

**2b. Add the hook field** to `Config` in the same file. Place it on `Config`
because it's session-level configuration (one hook per session, set by the
launcher) rather than per-goal:

```go
// GuiArtHook is a closure that displays an analysis graph in a UI. When nil,
// check.GuiArt prints diagnostic info and returns. Mirrors Python's gui_art
// delegation to tk_ui.new_ui (ivy_check.py:86-102). The type is defined in
// this package (above) so the field is fully typed without interface{} boxing.
GuiArtHook GuiArtHook `json:"-"`
```

(Field name and type name are both `GuiArtHook`; this is idiomatic Go since
they live in different namespaces — the type is `module.GuiArtHook`, the field
is `module.Config.GuiArtHook`.)

**2c. Optional convenience alias** in `check/phase7.go`:

```go
// GuiArtHook is an alias for module.GuiArtHook so check-package callers can
// declare hooks without importing module by name.
type GuiArtHook = module.GuiArtHook
```

This is purely for convenience; it can be omitted if call sites are happy to
write `module.GuiArtHook` directly.

After Step 2: build still passes (no callers yet). The hook field is unused
until Step 3.

### Step 3 — Implement `GuiArt`

Rewrite `check/phase7.go:19-31` (the stub) as a faithful port of Python's
`gui_art`. The function:

1. Performs the data setup that Python does (lines 91-98), correctly fixed.
2. Delegates to `mod.Cfg.GuiArtHook` if registered.
3. Otherwise prints diagnostic info and returns.
4. Does NOT call `os.Exit` itself — that's the caller's job (so unit tests can
   exercise the function without process termination).

```go
// GuiArt launches the GUI for an analysis graph. Corresponds to Python's
// gui_art (ivy_check.py:86-102).
//
// Python's gui_art ends with gui.tk.mainloop() then exit(1). In Go, the
// blocking-loop and process-exit responsibilities belong to the caller and
// the registered GuiArtHook (typically supplied by the webui package). When
// no hook is registered, GuiArt performs the same data setup Python does,
// prints a diagnostic summary, and returns nil — matching the previous stub
// behavior but with the data side-effects faithfully ported.
//
// `target` is interface{} to mirror Python's polymorphism: it accepts a
// *art.AnalysisGraph (from ShowCounterexample) or a *MatchHandler (from the
// trace failure path).
func GuiArt(mod *module.Module, target interface{}, isCti *module.Clauses) error {
    // Resolve the target to an AnalysisGraph for the data-setup branch.
    // (The hook itself receives the original target unchanged.)
    var otherArt *art.AnalysisGraph
    switch v := target.(type) {
    case *art.AnalysisGraph:
        otherArt = v
    case nil:
        otherArt = art.NewAnalysisGraph(mod)
    default:
        // *MatchHandler or other handler types — leave otherArt nil; the
        // "art" UI branch below will create a fresh graph just like Python.
        _ = v
    }

    // Python: if ivy_ui.default_ui.get() == "art":
    if mod.Cfg.IuCfg != nil && mod.Cfg.IuCfg.DefaultUI == "art" {
        // Python: print("initializers: {}".format(im.module.initializers))
        fmt.Printf("initializers: %d\n", len(mod.Initializers))
        // Python: other_art = ivy_art.AnalysisGraph()
        otherArt = art.NewAnalysisGraph(mod)
        // Python: other_art.add_initial_state()
        otherArt.AddInitialState(nil, nil)
        // Python: if 'initialize' in im.module.actions:
        if initAct, ok := mod.Actions.Get2("initialize"); ok {
            // Python: print("initialize: {}".format(init_action))
            fmt.Printf("initialize: %v\n", initAct)
            // Python: ag.execute(init_action, None, None, 'initialize')
            //
            // NOTE: Python references undefined `ag` here — almost certainly
            // a bug; the only graph in scope is `other_art`. We use otherArt
            // to make this branch actually run. The Python "art" UI mode is
            // rare in practice (default is "cti"), which is presumably why
            // the typo went unnoticed upstream.
            _, _ = otherArt.Execute(false, initAct, nil, nil, "initialize")
        }
    }

    // Python: agui = gui.add(other_art); gui.tk.update_idletasks();
    //         gui.tk.mainloop(); exit(1)
    //
    // In Go, delegate to the registered hook (typically webui). The hook is
    // a fully-typed module.GuiArtHook — no interface{} unboxing needed.
    if mod.Cfg.GuiArtHook != nil {
        // Pass the original target (which may be a Trace handler) when we have
        // one; otherwise pass the freshly-built otherArt from the "art" UI
        // branch above. Hooks that need a specific concrete type can
        // type-switch themselves.
        if target == nil && otherArt != nil {
            return mod.Cfg.GuiArtHook(mod, otherArt, isCti)
        }
        return mod.Cfg.GuiArtHook(mod, target, isCti)
    }

    // No hook registered: print diagnostic info matching the previous stub.
    fmt.Println("initializers:", len(mod.Initializers))
    if initAct, ok := mod.Actions.Get2("initialize"); ok {
        fmt.Println("initialize:", initAct)
    }
    if isCti != nil {
        fmt.Println("CTI clauses:", isCti)
    }
    fmt.Println("GUI mode not available in CLI; register a check.GuiArtHookFn or use --trace")
    return nil
}
```

Note that the function signature changes from `func GuiArt(mod *module.Module,
otherArt *art.AnalysisGraph) error` to `func GuiArt(mod *module.Module, target
interface{}, isCti *module.Clauses) error`. There are zero current callers, so
this is a free signature change.

After Step 3: build passes. The new function is callable but no production
code calls it yet.

### Step 4 — Wire `GuiArt` into all four stub call sites

Each of these is a faithful port of a Python `gui_art` call.

**4a. `check.go:605-613` (trace path).** Python `ivy_check.py:411-415`:
```python
if not opt_trace.get():
    gui_art(handler)
else:
    print(str(handler))
    exit(0)
```

Current Go (broken):
```go
if mod.Cfg.OptTrace {
    fmt.Println(handler.String())
    os.Exit(0)
} else {
    // GUI display not supported in Go; print trace instead
    fmt.Println(handler.String())
}
```

Fix:
```go
if mod.Cfg.OptTrace {
    fmt.Println(handler.String())
    os.Exit(0)
} else {
    // Python: gui_art(handler) — passes the trace handler. The handler
    // carries the IsCti clauses for the GUI's CTI display mode.
    if err := GuiArt(mod, handler, handler.IsCti); err != nil {
        fmt.Fprintf(os.Stderr, "GuiArt: %v\n", err)
    }
    // Python: gui_art ends with exit(1) after the Tk mainloop returns.
    os.Exit(1)
}
```

**4b. `check.go:876-883` (`DisplayCex`, Python `display_cex`).** Python
`ivy_check.py:54-60`:
```python
def display_cex(msg,ag):
    if diagnose.get():
        from . import tk_ui as ui
        iu.set_parameters({'mode':'induction'})
        ui.ui_main_loop(ag)
        exit(1)
    raise iu.IvyError(None,msg)
```

Current Go (broken):
```go
func DisplayCex(cfg *module.Config, msg string, ag interface{}) error {
    if cfg.Diagnose {
        return fmt.Errorf("%s (use web UI for interactive diagnostics)", msg)
    }
    return fmt.Errorf("%s", msg)
}
```

Fix: change the signature to take `*module.Module` so `GuiArt` can receive it,
then dispatch:
```go
func DisplayCex(mod *module.Module, msg string, ag interface{}) error {
    if mod.Cfg.Diagnose {
        // Python: ui.ui_main_loop(ag); exit(1)
        if err := GuiArt(mod, ag, nil); err != nil {
            fmt.Fprintf(os.Stderr, "DisplayCex GuiArt: %v\n", err)
        }
        os.Exit(1)
    }
    return fmt.Errorf("%s", msg)
}
```

`DisplayCex` currently has zero non-test callers (verified by grep), so this
signature change is safe. Update any test references if present.

**4c. `check.go:893-924` (`ShowCounterexample`).** Python `ivy_check.py:77-84`:
```python
def show_counterexample(ag,state,bmc_res):
    universe,path = bmc_res
    other_art = ivy_art.AnalysisGraph()
    ag.copy_path(state,other_art,None)
    for state,value in zip(other_art.states[-len(path):],path):
        state.value = value
        state.universe = universe
    gui_art(other_art)
```

Current Go: builds `otherArt`, then prints `"Counterexample found ..."`. Fix
the trailing print to call GuiArt:
```go
// Python: gui_art(other_art)
if err := GuiArt(ag.Domain, otherArt, nil); err != nil {
    fmt.Fprintf(os.Stderr, "ShowCounterexample GuiArt: %v\n", err)
}
```

Note: `ShowCounterexample` has a known-broken `bmcResult` type assertion that
prevents it from running with real data today (the `bmcRes interface{}` parameter
flows in as a `*solver.ModelResult`, not the local `*bmcResult` struct). Fully
fixing the BMC result plumbing is **out of scope** for this plan and is recorded
in the followup section. The GuiArt wiring still belongs here so that when the
type plumbing is fixed, the GUI dispatch is already in place.

**4d. `isolate_check.go:793-801` and `:849-857` (`CheckSubgoals`).** Python
`ivy_check.py:824-837` (the missing block):
```python
foo = method()
if foo:
    global failures
    failures += 1
    print("FAIL\n")
    if hasattr(goal,"trace_hook"):
        foo = goal.trace_hook(foo)
    if opt_trace.get():
        print(str(foo))
        exit(0)
    if diagnose.get():
        gui_art(foo)
```

Current Go (both branches): just prints "FAIL" and returns the error. The
Python `goal.trace_hook(foo)` indirection, the `opt_trace`-based exit, and the
`gui_art(foo)` call are all missing.

Replace the temporal-branch block (793-801) and the equivalent non-temporal
block (849-857) with:

```go
err := method()
if err != nil {
    mod.Cfg.Failures++
    fmt.Println("FAIL")
    // Python ivy_check.py:829-830: if hasattr(goal,"trace_hook"): foo = goal.trace_hook(foo)
    // The Go trace-hook propagation differs (it acts on a MatchHandler via
    // fakeMod.TraceHook in the no-method branch below, not as a transformer
    // here). For now we honor opt_trace and diagnose without re-routing the
    // failure value through goal.TraceHook — see followup item 5.
    // Python: if opt_trace.get(): print(str(foo)); exit(0)
    if mod.Cfg.OptTrace {
        fmt.Println(err)
        wsorts.Exit()
        ws.Exit()
        cleanup()
        os.Exit(0)
    }
    // Python: if diagnose.get(): gui_art(foo)
    if mod.Cfg.Diagnose {
        if guiErr := GuiArt(mod, err, nil); guiErr != nil {
            fmt.Fprintf(os.Stderr, "GuiArt: %v\n", guiErr)
        }
    }
    wsorts.Exit()
    ws.Exit()
    cleanup()
    return err
}
fmt.Println("PASS")
```

After Step 4: build passes, `go vet` passes. All four call sites now route
through `GuiArt`. Without a registered hook, behavior is unchanged for the
`-trace` (text) path and slightly improved for the GUI path (it now prints
proper diagnostic info instead of a generic "use web UI" string). With a
hook registered, each call site behaves like Python's `gui_art`.

### Step 5 — Verification

```sh
cd ~/ivy/goivy && go build ./...
cd ~/ivy/goivy && go vet ./...
cd ~/ivy/goivy && go test ./check/... ./module/... ./bmc/... ./compiler/... ./tactics/... ./temporal/...
cd ~/ivy/goivy && go build ./cmd/goivy_check/...
```

**End-to-end smoke tests:**

1. Find an existing `.ivy` file with `method = "bmc[N]"` in the repo's test
   corpus (grep `'method.*=.*bmc\['` under `~/ivy/goivy/`). Run:
   ```sh
   ~/ivy/goivy/cmd/goivy_check/goivy_check that_file.ivy
   ```
   Verify it prints `BOUNDED` then `OK`.

2. Construct a tiny temporary `.ivy` file with a property failure and run:
   ```sh
   ~/ivy/goivy/cmd/goivy_check/goivy_check diagnose=true tmp.ivy
   ```
   Verify `GuiArt` is invoked (stub prints "GUI mode not available …" plus
   the proper data lines), exit code is 1.

3. Same file with `trace=true`:
   ```sh
   ~/ivy/goivy/cmd/goivy_check/goivy_check trace=true tmp.ivy
   ```
   Verify trace text is printed, exit code is 0, GuiArt is *not* called.

**Stale-symbol verification:**
```sh
grep -rn 'someBounded\|TODO: wire BMC' ~/ivy/goivy/check ~/ivy/goivy/cmd
```
Should return nothing.

```sh
grep -rn 'use web UI for interactive diagnostics\|use web UI for visualization\|GUI display not supported in Go' ~/ivy/goivy/check ~/ivy/goivy/cmd
```
Should return nothing (all four stub strings should be deleted by Step 4).

## Files to modify

| File | Change | Step |
|---|---|---|
| `module/config.go` | Add `SomeBounded bool` field to `Config`; define `module.GuiArtHook` named func type and add `GuiArtHook GuiArtHook` field to `Config` | 1a, 2a, 2b |
| `check/isolate_check.go` | Set `mod.Cfg.SomeBounded = true` in BMC branch (line 997); rewrite `CheckSubgoals` failure handling in both temporal (793-801) and non-temporal (849-857) branches | 1b, 4d |
| `check/check.go` | Replace `someBounded := false` in `Start()` and `StartWithConfig()` with `mod.Cfg.SomeBounded` reads + resets; rewrite the `else` branch at 605-613 to call `GuiArt`; rewrite `DisplayCex` (876-883) signature and body; replace the trailing print in `ShowCounterexample` with `GuiArt(ag.Domain, otherArt, nil)` | 1c, 4a, 4b, 4c |
| `check/phase7.go` | Optional `type GuiArtHook = module.GuiArtHook` alias; rewrite `GuiArt` body and signature | 2c, 3 |
| `check/check_test.go` | Add `TestSomeBounded*` tests | 1d |
| `check/phase7_test.go` (new) | Add `TestGuiArt*` tests for nil/AnalysisGraph/MatchHandler targets, hook invocation, hook error propagation, and "art" UI branch | 5 |

No changes outside `check/`, `module/config.go`, and the new test file.

## Tests to add

### BMC flag (Step 1d)

- `TestSomeBoundedZeroByDefault` — fresh `module.NewConfig()` has `SomeBounded == false`
- `TestSomeBoundedSetByBMCBranch` — minimal isolate with `method = "bmc[1]"`, after `CheckModule` returns the parent `mod.Cfg.SomeBounded == true`
- `TestSomeBoundedNotSetForOtherMethods` — same setup with `mc` and induction methods
- `TestSomeBoundedResetByStartWithConfig` — pre-set the flag, run `StartWithConfig` against a no-op file, verify reset

### GuiArt (Step 5, in new `check/phase7_test.go`)

- `TestGuiArtNilTargetCreatesFreshGraph` — `GuiArt(mod, nil, nil)` returns nil without panic
- `TestGuiArtAnalysisGraphTarget` — passes `*art.AnalysisGraph`, returns nil
- `TestGuiArtMatchHandlerTarget` — passes a `*MatchHandler`, returns nil
- `TestGuiArtHookInvoked` — register a hook on `mod.Cfg.GuiArtHook`, assert it was called with matching args
- `TestGuiArtHookErrorPropagates` — hook returns an error, assert `GuiArt` returns it
- `TestGuiArtArtUIBranch` — set `mod.Cfg.IuCfg.DefaultUI = "art"`, add an `initialize` action, call GuiArt, assert no panic and that initializers count was printed

(`TestGuiArtHookWrongType` is no longer needed because the hook field is now
fully typed — wrong types are caught by the compiler, not at runtime.)

Do not unit-test the four `os.Exit` call sites directly — Go's test machinery
handles these poorly. Test the pre-exit logic in isolation.

## Critical files

- `/Users/jaten/ivy/goivy/check/phase7.go` — primary implementation site for `GuiArt`
- `/Users/jaten/ivy/goivy/check/check.go` lines 605-613, 876-883, 893-924, 1161-1207, 1244-1290 — call sites to wire and the BMC flag plumbing
- `/Users/jaten/ivy/goivy/check/isolate_check.go` lines 793-801, 849-857, 997-1020 — BMC flag write site and `CheckSubgoals` failure paths
- `/Users/jaten/ivy/goivy/check/helpers.go` lines 147-150 — existing `MatchHandler.IsCti` field that finally gets read after this plan
- `/Users/jaten/ivy/goivy/module/config.go` — add `SomeBounded bool` field, define `module.GuiArtHook` named func type, and add `GuiArtHook GuiArtHook` field on `Config`
- `/Users/jaten/ivy/goivy/module/module.go` line 337 — `Module.Copy()` value-copies Cfg (read-only reference; explains why `SomeBounded` writes must target the parent `mod.Cfg`, not `isoMod.Cfg`)
- `/Users/jaten/ivy/goivy/ivyutils/config.go` line 81 — `IvyUtilsConfig.DefaultUI = "cti"` (read-only reference; the `"art"` branch in `GuiArt` keys off this)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_check.py` lines 54-60, 86-102, 824-837, 990-1000, 1028-1029, 1046-1047 — Python source of truth

## Out of scope (explicit followups, not done here)

These were considered and intentionally deferred. Each is recorded so it does
not get lost.

1. **Webui registration of a `GuiArtHookFn` implementation.** The webui package
   already has `AnalysisGraphUI`, `IvySession`, and a complete server. Wiring
   it to register a `mod.Cfg.GuiArtHook = check.GuiArtHookFn(myImpl)` from
   `cmd/goivy_check/main.go` (or via `webui.Register()`-style init) is its own
   project — touching server lifecycle, browser launching, and CLI flags for
   GUI vs. headless mode. This plan leaves the hook nil so behavior matches
   the prior text-mode stub when no GUI is requested.

2. **Deduplicating `Start()` and `StartWithConfig()`.** They are nearly identical
   except for the config-passing setup. Cleaner as a separate refactor.

3. **`Module.Copy()` not propagating `TraceHook` (`module.go:334`).** Latent bug.
   `*c.Cfg = *m.Cfg` copies Config-level fields by value but `Module.TraceHook
   interface{}` is a Module-level field that is reset to nil on every Copy().
   Out of scope; recorded as a separate issue.

4. **`ShowCounterexample` dead `bmcResult` type assertion.** The `bmcRes
   interface{}` parameter is documented as `(universe, path)` but the local
   `*bmcResult` struct never matches what `history.SatisfyWithCond` returns
   (it returns `*solver.ModelResult`). The function silently falls through
   to its "use web UI" stub today. Plumbing real BMC counterexample data
   through `history.SatisfyWithCond` is a separate task. The GuiArt wiring
   still goes in now so it's ready when the data is.

5. **Unifying `check.TraceHookFn` with Python's `goal.trace_hook` transformation
   semantics.** The Go hook operates on a `MatchHandler` for side effects; the
   Python hook *transforms* the failure value (`foo = goal.trace_hook(foo)`).
   The CheckSubgoals fix in Step 4d acknowledges this in a comment but does
   not paper over it. Unifying requires touching every Go trace-hook
   producer/consumer; out of scope.

6. **Tk UI porting.** Python's `tk_ui.py` and the Tk-based `ivy_ui.py`
   `mainloop()` are explicitly NOT being ported. The Go equivalent is the
   webui package, and GUI behavior is delegated through the hook.

## Risks and rollback

- **Risk: `mod.Cfg.SomeBounded` write to wrong scope.** Mitigated by the
  explicit comment at the BMC branch and by `TestSomeBoundedSetByBMCBranch`.
  If we accidentally wrote to `isoMod.Cfg` instead, the test would catch it
  immediately because the parent `mod.Cfg.SomeBounded` would stay false.

- **Risk: `os.Exit(1)` in production code path makes a test crash.** Mitigated
  by NOT putting `os.Exit` inside `GuiArt` itself — only at the four call
  sites. Tests target the helpers and `GuiArt` directly, not the call sites.

- **Risk: signature change to `DisplayCex` breaks tests.** Verified by grep
  before this plan: `DisplayCex` has zero non-test references. If any tests
  reference it, update them in Step 4b.

- **Risk: signature change to `GuiArt` breaks callers.** Verified: zero
  current callers anywhere in the tree. The new signature is free.

- **Risk: `art.AnalysisGraph.AddInitialState(nil, nil)` panics on a fresh
  graph.** Mitigated by reading the implementation at `art/art.go:1506`
  during Step 3 and writing the test `TestGuiArtArtUIBranch` to exercise
  this path explicitly.

- **Rollback**: each step is independently revertible. If Step 4 causes an
  end-to-end regression, revert just Step 4 and the call sites; Steps 1-3
  are safe additive changes.
