# Research: Ivy Test Case Generation — Status & Port Gap

## Context

The Ivy introduction promises test bench and test oracle generation for rigorous specification-based testing. The user asked: (1) was this implemented in Python Ivy? (2) was it ported to goivy?

---

## Finding 1: Python Ivy HAS Full Test Generation (tied to C++ backend)

All test generation lives in **`~/ivy/pyivy/ivy/ivy/ivy_to_cpp.py`** (5,900+ lines).

### Compilation targets (line 5728)
```
target = EnumeratedParameter("target", ["impl","gen","repl","test","class"], "gen")
```
The **`test`** target triggers full test bench emission.

### Key functions:
| Function | Line | Purpose |
|---|---|---|
| `emit_init_gen()` | 905 | Emits C++ class `init_gen` — generates initial states satisfying init conditions via Z3 |
| `emit_action_gen()` | 1192 | Emits `<action>_gen` class per public action — uses Z3 to find inputs satisfying preconditions |
| `emit_repl_boilerplate3test()` | 5024 | Emits test harness loop: weighted-random action selection, `test_iters` iterations, `test_runs` runs |

### Parameters (lines 5732-5733):
- `test_iters` — iterations per run (default 100)
- `test_runs` — independent runs (default 1)

### Assertion/invariant checking (lines 4674-4689):
- `ivy_assert()` / `ivy_assume()` in the generated `classname_repl` class serve as **test oracles**
- When `target=test`, invariants are compiled in (line 5866): `compile_with_invariants.set("true")`

### Z3 integration (lines 920-931, 1216-1272):
- Extracts preconditions from action AST
- Converts to Z3 assertions, calls `solve()`
- Falls back to `rand()` randomization if Z3 finds no solution

### Coverage note (`ivy_check.py` line 42):
- `coverage = iu.BooleanParameter("coverage", True)` — coverage checking during verification

---

## Finding 2: goivy Does NOT Have Test Generation

The Go port at `~/ivy/goivy/` has:
- ✅ Full compiler infrastructure (`compiler*.go`)
- ✅ Z3 SMT solver bridge (`z3bridge_*.go`, `smt/`)
- ✅ Code generation to Go (`gogen/`), Lean (`leangen/`), Dafny (`dafnygen/`)
- ✅ Verification pipeline (check, isolate, model checking)

But is **completely missing**:
- ❌ `test` compilation target
- ❌ `emit_init_gen()` equivalent
- ❌ `emit_action_gen()` equivalent  
- ❌ `emit_repl_boilerplate3test()` equivalent
- ❌ `test_iters` / `test_runs` parameters
- ❌ Weighted random action selection framework
- ❌ Test oracle (assertion/invariant checking during generated test runs)

### Why it was omitted:
The test generation machinery is deeply embedded in `ivy_to_cpp.py` — it generates **C++ code** that links against Z3 at runtime. When the Go port skipped the C++ translation target entirely (no `ivy_to_cpp.go` equivalent), the test generation went with it. The goivy port focuses on verification (model checking, proof checking) rather than executable test bench generation.

---

## Potential Port Approach (if user wants it)

The core idea is to generate **Go test code** instead of C++ test code, using goivy's existing Z3 bridge at generation time (not runtime):

1. **Add `test` target to goivy's compilation pipeline** — wire into `module_config.go` and compiler
2. **Action generator**: For each public action, use Z3 (already present in goivy) at *generation time* to enumerate valid input combinations satisfying preconditions — emit as Go test table entries
3. **Init generator**: Generate initial states satisfying init conditions
4. **Test harness**: Emit a Go `TestGeneratedHarness` that loops through actions with weighted-random selection
5. **Test oracles**: Assertions and invariants compiled into the generated harness

**Alternatively**: Generate test drivers that call the web/REPL interface rather than linking Z3 at runtime.

**Scope**: Large — probably 2,000–4,000 lines of new Go codegen. Requires deep understanding of the goivy AST and how preconditions are represented.
