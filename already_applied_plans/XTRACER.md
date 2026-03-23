# Plan: xtracer — Parallel Execution Tracer for Go and Python Ivy

**Created: 2026-03-22**

## Context

Debugging goivy_check requires tracing where Go execution diverges from Python. Currently we fix bugs one at a time by running goivy_check, seeing an error, investigating, and fixing. An execution tracer that prints matching messages at corresponding code points in both Go and Python would let us diff the output and find the first divergence instantly.

## Existing Infrastructure

The `end2end/` package already has:
- `pythonIvyCheck(ivyFile)` — spawns `python3 -m ivy.ivy_check` as subprocess, captures output
- `goIvyCheck(ivyFile)` — runs Go compiler pipeline, returns pass/fail
- `conform_test.go` — compares Go vs Python results

The `pytesthelper/` has a heavier HTTP sidecar, but `end2end/` is lighter weight — simple subprocess execution with stdout capture. We'll build on that.

## Design

### Trace Message Format

Both Go and Python emit identical messages to stdout:

```
XTRACE: <category>.<function> <event> [<detail>]
```

Examples:
```
XTRACE: parser.Parse ENTER file=order.ivy
XTRACE: parser.Parse EXIT decls=47
XTRACE: parser.include ENTER name=order
XTRACE: parser.include EXIT decls=385
XTRACE: compiler.IvyCompile ENTER decls=432
XTRACE: compiler.DomainSetup ENTER
XTRACE: compiler.DomainSetup.type ENTER name=lclock
XTRACE: compiler.DomainSetup.type EXIT name=lclock
XTRACE: compiler.DomainSetup EXIT
XTRACE: compiler.ARGSetup ENTER
XTRACE: compiler.ARGSetup.action ENTER name=bar.a
XTRACE: compiler.ARGSetup.action EXIT name=bar.a
XTRACE: compiler.ARGSetup.isolate ENTER name=cf_live
XTRACE: compiler.ARGSetup.isolate EXIT name=cf_live
XTRACE: compiler.ARGSetup EXIT
XTRACE: compiler.IvyCompile EXIT
XTRACE: check.CheckModule ENTER isolates=3
XTRACE: check.CheckIsolate ENTER name=cf_live
```

### Go: Build Tag `xtracer`

**File: `xtracer/xtracer.go`** (new package)
```go
//go:build xtracer

package xtracer

import "fmt"

var Enabled = true

func Trace(format string, args ...interface{}) {
    fmt.Printf("XTRACE: "+format+"\n", args...)
}
```

**File: `xtracer/xtracer_off.go`**
```go
//go:build !xtracer

package xtracer

var Enabled = false

func Trace(format string, args ...interface{}) {}  // no-op
```

Usage in Go code:
```go
import "github.com/glycerine/goivy/xtracer"

func (p *Parser) Parse() (*ParseResult, error) {
    xtracer.Trace("parser.Parse ENTER")
    defer xtracer.Trace("parser.Parse EXIT decls=%d", len(decls))
    // ...
}
```

When built with `go build -tags xtracer ./cmd/goivy_check/`, traces are emitted. Without the tag, `Trace` is a no-op — zero overhead.

### Python: `if __debug__:` Guard

Python's `__debug__` is True by default, False when run with `python3 -O`.

**File: `ivy/xtracer.py`** (new module)
```python
import sys

def trace(msg):
    print("XTRACE: " + msg, file=sys.stdout, flush=True)
```

Usage in Python code:
```python
from . import xtracer

def ivy_compile(decls, mod=None, ...):
    if __debug__: xtracer.trace("compiler.IvyCompile ENTER decls=%d" % len(decls.decls))
    # ... existing code ...
    if __debug__: xtracer.trace("compiler.IvyCompile EXIT")
```

When run normally: traces are emitted. When run with `python3 -O`: the `if __debug__:` blocks are eliminated entirely at bytecode compile time.

### Trace Points (Phase 1: Entry/Exit Only)

#### Parser (6 points)
| Function | Python File | Go File |
|----------|-------------|---------|
| `parse/Parse` | `ivy_parser.py:3120` | `parser/parser.go:Parse()` |
| `parser.include` | `ivy_parser.py:370` (p_top_include_symbol) | `parser/decl.go:parseIncludeDeclMulti()` |
| `parser.inst_mod` | `ivy_parser.py:127` | `parser/parser.go:expandInstantiation()` |
| `parser.create_object` | `ivy_parser.py:612` | `parser/decl.go:parseObjectDeclMulti()` |
| `parser.do_insts` | `ivy_parser.py:192` | `parser/decl.go:parseInstantiateDeclMulti()` |
| `parser.expand_auto` | `ivy_parser.py:3089` | `parser/parser.go:expandAutoInstances()` |

#### Compiler (8 points)
| Function | Python File | Go File |
|----------|-------------|---------|
| `compiler.IvyCompile` | `ivy_compiler.py:2190` | `compiler/ivy_compile.go:IvyCompile()` |
| `compiler.CheckInstantiations` | `ivy_compiler.py:1618` | `compiler/phase6.go:CheckInstantiations()` |
| `compiler.CollectActions` | (implicit in TopContext) | `compiler/ivy_compile.go:CollectActions()` |
| `compiler.DomainSetup` | `ivy_compiler.py:1024` (__call__) | `compiler/decl.go:DomainSetup.ProcessDecls()` |
| `compiler.ConjSetup` | `ivy_compiler.py:1374` (__call__) | `compiler/ivy_compile.go:ConjSetup.ProcessDecls()` |
| `compiler.ARGSetup` | `ivy_compiler.py:1399` (__call__) | `compiler/ivy_compile.go:ARGSetup.ProcessDecls()` |
| `compiler.DomainSetup.type` | `ivy_compiler.py:1216` | `compiler/decl.go:DomainSetup.TypeDecl()` |
| `compiler.ARGSetup.action` | `ivy_compiler.py:1414` | `compiler/ivy_compile.go:ARGSetup case ActionDecl` |

#### Init/Load (3 points)
| Function | Python File | Go File |
|----------|-------------|---------|
| `init.SourceFile` | `ivy_init.py:69` (source_file) | `ivyinit/ivyinit.go:SourceFile()` |
| `init.ReadModule` | `ivy_compiler.py:2267` (read_module) | `ivyinit/ivyinit.go:ReadModule()` |
| `init.ImportModule` | `ivy_compiler.py:2298` (import_module) | `ivyinit/ivyinit.go:ImportModule()` |

#### Check (4 points)
| Function | Python File | Go File |
|----------|-------------|---------|
| `check.start` | `ivy_check.py:975` | `check/check.go:StartWithConfig()` |
| `check.CheckModule` | `ivy_check.py:900` | `check/isolate_check.go:CheckModule()` |
| `check.CheckIsolate` | `ivy_check.py:495` | `check/isolate_check.go:CheckIsolate()` |
| `check.CreateIsolate` | `ivy_isolate.py:1557` | `isolate/create.go:CreateIsolate()` |

**Total: 21 trace points** (entry + exit = 42 messages per run)

### Test Runner

**File: `end2end/xtracer_test.go`** (new)

Extends the existing `end2end` test infrastructure:

```go
func TestXtracer_Conform(t *testing.T) {
    ivyFile := "../ivy-lang-examples/test/action1.ivy"

    // Run Go with xtracer tag (requires separate binary)
    goTrace := runGoWithTrace(t, ivyFile)

    // Run Python with __debug__ (no -O flag)
    pyTrace := runPythonWithTrace(t, ivyFile)

    // Compare trace lines
    goLines := filterXTRACE(goTrace)
    pyLines := filterXTRACE(pyTrace)

    // Find first divergence
    for i := 0; i < min(len(goLines), len(pyLines)); i++ {
        if goLines[i] != pyLines[i] {
            t.Errorf("DIVERGENCE at trace line %d:\n  Go: %s\n  Py: %s", i, goLines[i], pyLines[i])
            break
        }
    }
}
```

`runPythonWithTrace` uses the existing `end2end` pattern: `exec.Command("python3", "-m", "ivy.ivy_check", ivyFile)` with stdout captured. The `if __debug__:` trace points are active by default.

`runGoWithTrace` builds a temporary binary with `-tags xtracer` and runs it.

## Files to Create/Modify

| File | Action |
|------|--------|
| `xtracer/xtracer.go` | NEW — enabled trace with build tag |
| `xtracer/xtracer_off.go` | NEW — no-op without build tag |
| `~/pyivy/ivy/ivy/xtracer.py` | NEW — Python trace function |
| `parser/parser.go` | ADD — 4 trace calls (Parse, expandAutoInstances, expandInstantiation) |
| `parser/decl.go` | ADD — 3 trace calls (include, object, instantiate) |
| `compiler/ivy_compile.go` | ADD — 6 trace calls (IvyCompile, CollectActions, ConjSetup, ARGSetup) |
| `compiler/decl.go` | ADD — 2 trace calls (DomainSetup.ProcessDecls, TypeDecl) |
| `compiler/phase6.go` | ADD — 1 trace call (CheckInstantiations) |
| `ivyinit/ivyinit.go` | ADD — 3 trace calls (SourceFile, ReadModule, ImportModule) |
| `check/check.go` | ADD — 1 trace call (StartWithConfig) |
| `check/isolate_check.go` | ADD — 2 trace calls (CheckModule, CheckIsolate) |
| `isolate/create.go` | ADD — 1 trace call (CreateIsolate) |
| `~/pyivy/ivy/ivy/ivy_parser.py` | ADD — 6 trace calls matching Go parser |
| `~/pyivy/ivy/ivy/ivy_compiler.py` | ADD — 8 trace calls matching Go compiler |
| `~/pyivy/ivy/ivy/ivy_init.py` | ADD — 1 trace call (source_file) |
| `~/pyivy/ivy/ivy/ivy_check.py` | ADD — 4 trace calls (start, check_module, check_isolate) |
| `end2end/xtracer_test.go` | NEW — conformance test comparing traces |

## Verification

1. `go build -tags xtracer ./cmd/goivy_check/` compiles
2. `go build ./cmd/goivy_check/` compiles (no trace, zero overhead)
3. `./goivy_check_traced action1.ivy` prints XTRACE lines
4. `python3 -m ivy.ivy_check action1.ivy` prints XTRACE lines
5. `python3 -O -m ivy.ivy_check action1.ivy` prints NO XTRACE lines
6. `go test ./end2end/ -run TestXtracer` compares traces
7. `go test ./...` (without xtracer tag) still passes — no trace output
