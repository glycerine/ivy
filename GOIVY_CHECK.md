# Plan: Create `goivy_check` Command-Line Proof Checker

**Created: 2026-03-22 (current session)**

## Context

Python Ivy provides `ivy_check` as the command-line proof checker (registered via `setup.py` entry_points as `ivy_check=ivy.ivy_check:main`). The Go port already has a comprehensive `check/` package with `CheckModule()` and `CheckIsolate()` fully implemented, but no CLI entry point exists to invoke it. We need to create `~/goivy/cmd/goivy_check/main.go`.

## Key Finding: The Hard Work Is Already Done

After reading the Go port's `check/check.go` (1253 lines) and `check/isolate_check.go` (1306 lines), the entire verification pipeline is already implemented:

- **`CheckModule(mod)`** — full isolate enumeration, method dispatch (ic/mc/vmt/bmc), coverage checking, ACL preprocessing, macro_finder save/restore
- **`CheckIsolate(mod, traceHook)`** — full property checking, initialization invariant establishment, action preservation loop, guarantee checking, temporal checking
- **`CheckFcsInStateWithAG()`** — full solver integration with trace/diagnose path AND normal path (calls `ag.GetHistory()`, `tr.SmallModelClauses()`, `history.SatisfyWithCond()`)
- **`CheckSubgoals()`** — temporal model goal decomposition
- **`MCIsolate()`** — model checking dispatch with separate-per-assertion support
- **Checker infrastructure** — `BaseChecker`, `ConjChecker`, `ConjAssumer` with full `Checker` interface

The `cmd/goivy_check/main.go` is purely a CLI wrapper: parse args → load source → compile → check → report result.

## Python `ivy_check.main()` Flow

```
main()
  → signal.signal(SIGINT, SIG_DFL)
  → ivy_alpha.test_bottom = False
  → ivy_init.read_params()          # parse key=value args from sys.argv
  → validate: len(sys.argv)==2 and sys.argv[1].endswith('.ivy')
  → with im.Module():
      → ivy_init.source_file(...)   # parse + compile into module
      → check_module()              # the verification orchestrator
  → print "OK" / "BOUNDED" / "OK, but used 'sorry'"
```

## Implementation Plan

### File to Create

**`/Users/jaten/go/src/github.com/glycerine/goivy/cmd/goivy_check/main.go`**

### Structure (Mechanical Port of Python `ivy_check.main()` + `start()`)

```go
package main

import (
    "fmt"
    "os"
    "strings"
    "signal" // for SIGINT default handling

    "github.com/glycerine/goivy/check"
    "github.com/glycerine/goivy/compiler"
    il "github.com/glycerine/goivy/ivylogic"
    "github.com/glycerine/goivy/ivyinit"
    iu "github.com/glycerine/goivy/ivyutils"
    "github.com/glycerine/goivy/module"
)

func main() {
    // 1. Signal handling (Python: signal.signal(signal.SIGINT, signal.SIG_DFL))
    signal.Reset(os.Interrupt)

    // 2. Parse key=value parameters from args (Python: ivy_init.read_params())
    reg := iu.NewParameterRegistry()  // or use default registry
    remaining, err := ivyinit.ReadParams(os.Args[1:], reg)
    // ... error handling

    // 3. Validate: exactly one .ivy file argument
    if len(remaining) != 1 || !strings.HasSuffix(remaining[0], ".ivy") {
        usage()
    }
    filename := remaining[0]

    // 4. Create module context (Python: with im.Module():)
    mod := module.New()
    cfg := module.NewConfig()
    // Map parsed parameters to cfg fields (see parameter mapping below)
    mod.Cfg = cfg

    // 5. Load and compile source (Python: ivy_init.source_file(...))
    sig := il.NewSig()
    err = ivyinit.SourceFile(filename, mod, sig, nil)
    // ... error handling

    // 6. Run verification (Python: check_module())
    err = check.CheckModule(mod)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    // 7. Report result
    // Python: if some_bounded: print("BOUNDED")
    //         elif ivy_tactics.used_sorry: print("OK, but used 'sorry'")
    //         else: print("OK")
    fmt.Println("OK")
}

func usage() {
    fmt.Fprintf(os.Stderr, "usage: goivy_check [key=value ...] file.ivy\n")
    os.Exit(1)
}
```

### Parameter Mapping (Python Parameters → Go `module.Config`)

These parameters are parsed as `key=value` pairs before the filename:

| Python Parameter | Config Field | Type | Default |
|---|---|---|---|
| `diagnose=true` | `Cfg.Diagnose` | bool | false |
| `coverage=true` | `Cfg.Coverage` | bool | true |
| `action=name` | `Cfg.CheckedAction` | string | "" |
| `trusted=true` | `Cfg.OptTrusted` | bool | false |
| `mc=true` | `Cfg.OptMC` | bool | false |
| `trace=true` | `Cfg.OptTrace` | bool | false |
| `separate=true` | `Cfg.OptSeparate` | bool | false |
| `unchecked_properties=file` | `Cfg.OptUncheckedProps` | string | "" |
| `ivy_stats=true` | `Cfg.OptIvyStats` | bool | false |
| `prioritize=a,b,c` | `Cfg.PriorityActions` | string | "" |
| `no_check_guarantees=true` | `Cfg.NoCheckGuarantees` | bool | false |
| `summary=true` | `Cfg.OptSummary` | bool | false |
| `isolate=name` | `Cfg.Isolate` | string | "" |
| `checked_assert=line` | `Cfg.CheckLineno` | string | "" |
| `unprovable=true` | `Cfg.OnlyCheckUnprovable` | bool | false |

The existing `module.Config` struct already has fields for all of these (verified in `module/config.go`).

### Existing Functions to Reuse

| Purpose | Function | Location |
|---|---|---|
| Parse CLI params | `ivyinit.ReadParams(args, reg)` | `ivyinit/ivyinit.go` |
| Load & parse .ivy file | `ivyinit.ReadModule(filename, false)` | `ivyinit/ivyinit.go` |
| Compile to module | `ivyinit.SourceFile(filename, mod, sig, kwargs)` | `ivyinit/ivyinit.go` |
| Compile declarations | `compiler.IvyCompile(decls, mod)` | `compiler/ivy_compile.go` |
| Create isolate | `isolate.CreateIsolate(iso, mod)` | `isolate/create.go` |
| Top-level check | `check.CheckModule(mod)` | `check/isolate_check.go:802` |
| Per-isolate check | `check.CheckIsolate(mod, hook)` | `check/isolate_check.go:34` |
| Model config | `module.NewConfig()` | `module/config.go` |

### Special Cases to Handle

1. **`NOT CHECKED` early exit** (Python line 990): If `checked_assert` is set to `none.ivy:0`, print "NOT CHECKED" and exit(0).

2. **`some_bounded` flag** (Python line 997): If any isolate used BMC, print "BOUNDED" instead of "OK". The `check/` package tracks this in `mod.Cfg.SomeBounded` (verify this field exists, add if not).

3. **`used_sorry`** (Python line 999): If any proof tactic used `sorry`, print warning. Check if `tactics.UsedSorry` global exists.

4. **ACL file loading** (Python line 911): If `unchecked_properties` is set, call `acl.RegisterFromFile()` before checking. This is already handled inside `check.CheckModule()`.

5. **Error printing**: Python uses `ErrorPrinter()` context manager. Go should use `defer` with error formatting.

### Verification

1. **Build**: `go build ./cmd/goivy_check/` compiles
2. **Unit tests**: `go test ./check/...` passes
3. **Integration test**: Run `goivy_check` on a simple `.ivy` file:
   ```
   goivy_check ivy-lang-examples/test/trivial.ivy
   ```
   Should print checking output and "OK"
4. **Conformance test**: Run both `ivy_check` (Python) and `goivy_check` (Go) on the same file, compare output structure (not exact text, but same PASS/FAIL result)

### Estimated Scope

- **New code**: ~80-120 lines in `cmd/goivy_check/main.go`
- **Changes to existing code**: Potentially add `SomeBounded` field to `module.Config` if missing; check for `tactics.UsedSorry` global
- **Risk**: Low — this is purely CLI glue connecting existing, tested infrastructure
