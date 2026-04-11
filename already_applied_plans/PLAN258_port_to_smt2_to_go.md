# Plan: Port Python's `to_smt2()` to Go to Fix SMT2 Serialization Divergence

Created: 2026-04-11 00:15 UTC

## Context

`make golden` now diverges at `~/ivy/goivy/log.red` line 254714, on the `z3.check` trace:

```
254714  go : XTRACE: z3.check seq=1 result=unsat HASH
              leaf=blake3.33B-0BqO_pzhEvwh5wEb3K7FUtAJeCO_ZFSFQC02pM99hANK
              root=blake3.33B-Jykd9ebBajktwodxl3qWauC_RYt9clScHZ_e9YkuvpSa
              smt2=(declare-sort lclock)
        py : XTRACE: z3.check seq=1 result=unsat HASH
              leaf=blake3.33B-HtNHleHO8kt4Hp2uqkyD5zyTiiijvMHwjw3FEKtvdCJc
              root=blake3.33B-SbuuYA-Xcwb6cnaIlmU_DA3_vvvXkYVF5SzROLrn_TUS
              smt2=; benchmark generated from python API
```

Both sides agree on `result=unsat`. The `seq=1` is identical. **Only the smt2 serialization differs**, and the leaf hashes diverge as a downstream consequence (because the hash is computed over a canon string that includes the smt2 text).

This is **not a translation correctness bug** — both Go and Python reach the same logical solver state and both Z3 instances answer `unsat`. It is a **trace-fingerprint format mismatch**: Go's `Z3Solver.String()` calls `Z3_solver_to_string` (s-expression format) while Python's `s.to_smt2()` calls `Z3_benchmark_to_smtlib_string` (SMT-LIB2 benchmark format with `; benchmark generated from python API` header).

The fix is to mirror Python's `to_smt2()` on the Go side so both bindings produce the same byte string.

## Why the visible smt2 in the trace is so short

`golden_test.go:541-571` reads each trace with `bufio.Reader.ReadString('\n')` and skips lines that don't start with `XTRACE:`. Both Go's `Z3_solver_to_string` and Python's `Z3_benchmark_to_smtlib_string` return **multi-line strings**. The first line is what shows up in the comparison; subsequent lines are read and silently discarded.

- Go's first line: `(declare-sort lclock)` — first declaration in sexpr format
- Python's first line: `; benchmark generated from python API` — SMT-LIB2 header comment

**However, the leaf and root hashes are computed from the canon string that contains the *full multi-line* smt2** (xtrace.go:73 `canon := iu.Canonical(fmt.Sprintf("(z3check seq=%d result=%s smt2=%s)", seq, rs, smt2))`). So even though only the first line is visible in the diff, the hash divergence is real and reflects the full underlying serialization mismatch.

This means: **fixing the format is mandatory; cosmetic newline-stripping is a separate, optional cleanup.**

## Investigation summary

### Python side (source of truth)

Reference: `~/ivy/pyivy/ivy/ivy/ivy_solver.py:1314-1346`

```python
_z3_check_counter = [0]
_z3_merkle = xtracer.MerkleState()

def _trace_z3_check(s, res):
    _z3_check_counter[0] += 1
    smt2 = _normalize_z3_varnames(s.to_smt2())          # ← KEY LINE
    rs = 'sat' if res == z3.sat else ('unsat' if res == z3.unsat else 'unknown')
    canon = "(z3check seq=%d result=%s smt2=%s)" % (_z3_check_counter[0], rs, smt2)
    leaf, root = _z3_merkle.add_leaf(canon)
    xtracer.trace("z3.check seq=%d result=%s HASH leaf=%s root=%s smt2=%s" % (
        _z3_check_counter[0], rs, leaf, root, smt2))

# Monkey-patches z3.Solver.check, so EVERY check() call emits this trace.
```

Python's `z3.Solver.to_smt2()` is in the vendored Z3 Python bindings at `~/ivy/goivy/z3vendor/z3/src/api/python/z3/z3.py:6495-6509`:

```python
def to_smt2(self):
    """return SMTLIB2 formatted benchmark for solver's assertions"""
    es = self.assertions()                # → Z3_solver_get_assertions
    sz = len(es)
    sz1 = sz
    if sz1 > 0:
        sz1 -= 1                          # last assertion split off
    v = (Ast * sz1)()
    for i in range(sz1):
        v[i] = es[i].as_ast()             # all-but-last → assumptions
    if sz > 0:
        e = es[sz1].as_ast()              # last → formula
    else:
        e = BoolVal(True, self.ctx).as_ast()  # empty solver → True
    return Z3_benchmark_to_smtlib_string(
        self.ctx.ref(),
        "benchmark generated from python API",  # name
        "",                                       # logic
        "unknown",                                # status
        "",                                       # attributes
        sz1, v, e)
```

**Critical quirks** (must be mirrored byte-for-byte to make hashes match):

1. The benchmark **name** is the literal string `"benchmark generated from python API"`. This produces the `; benchmark generated from python API` header in the output.
2. **logic** is empty string `""`.
3. **status** is the literal string `"unknown"` — *not* the actual sat/unsat result. (Python writes "unknown" regardless of what the solver returned.)
4. **attributes** is empty string `""`.
5. The **last assertion is the "formula"**; all preceding assertions become "assumptions". When the solver has 0 assertions, the formula is `BoolVal(True)`.
6. The result then runs through `_normalize_z3_varnames` which replaces `!k!N` with `!v!N` based on first-occurrence order.

### Go side (current state — broken)

Reference: `~/ivy/goivy/z3bridge/xtrace.go:54-76`:

```go
func (s *Z3Solver) TraceCheck(result CheckResult) {
    if !xtracer.Enabled { return }
    seq := s.ctx.z3CheckCounter.Add(1)
    smt2 := NormalizeZ3VarNames(s.String())             // ← USES Z3_solver_to_string
    var rs string
    switch result {
    case Sat:    rs = "sat"
    case Unsat:  rs = "unsat"
    default:     rs = "unknown"
    }
    canon := iu.Canonical(fmt.Sprintf("(z3check seq=%d result=%s smt2=%s)", seq, rs, smt2))
    leaf, root := s.ctx.z3Merkle.AddLeaf(canon)
    xtracer.Trace("z3.check seq=%d result=%s HASH leaf=%s root=%s smt2=%s", seq, rs, leaf, root, smt2)
}
```

`Z3Solver.String()` is at `quantifier.go:1493-1501`:

```go
func (s *Z3Solver) String() string {
    var res string
    s.ctx.do(func() {
        res = C.GoString(C.Z3_solver_to_string(s.ctx.c, s.c))   // ← s-expression format
    })
    runtime.KeepAlive(s)
    return res
}
```

This calls `Z3_solver_to_string`, which produces a different format than `Z3_benchmark_to_smtlib_string`.

### What CGO bindings already exist

Verified in `/Users/jaten/ivy/goivy/z3bridge/`:

| Z3 C function | Already used? | Where |
|---|---|---|
| `Z3_solver_to_string` | yes | `quantifier.go:1497` (the broken path) |
| `Z3_solver_get_unsat_core` | yes | `quantifier.go:1621` (returns `Z3_ast_vector`) |
| `Z3_ast_vector_inc_ref` | yes | `quantifier.go:1622, 1569`, `interp.go:67` |
| `Z3_ast_vector_size` | yes | `quantifier.go:1623, 1570`, `interp.go:68` |
| `Z3_ast_vector_get` | yes | `quantifier.go:1626, 1572`, `interp.go:71` |
| `Z3_ast_vector_dec_ref` | yes | `quantifier.go:1629, 1575`, `interp.go:74` |
| `Z3_mk_true` | yes | `quantifier.go:770` |
| `Z3_inc_ref` / `Z3_dec_ref` | yes | (in `newExpr`) |
| **`Z3_solver_get_assertions`** | **NO** | needs nothing more than CGO seeing it through `<z3.h>` |
| **`Z3_benchmark_to_smtlib_string`** | **NO** | same — declared in `<z3.h>` |

Both new calls are declared in `<z3.h>` (verified at `z3vendor/z3/src/api/z3_api.h:5159` and `:6024`), and `<z3.h>` is already included by `quantifier.go:20`. **No new CGO declaration block is needed.** Just use `C.Z3_solver_get_assertions(...)` and `C.Z3_benchmark_to_smtlib_string(...)` directly.

### Z3 C signatures (already in `z3.h`)

```c
Z3_ast_vector Z3_API Z3_solver_get_assertions(Z3_context c, Z3_solver s);

Z3_string Z3_API Z3_benchmark_to_smtlib_string(
    Z3_context c,
    Z3_string name,
    Z3_string logic,
    Z3_string status,
    Z3_string attributes,
    unsigned num_assumptions,
    Z3_ast const assumptions[],
    Z3_ast formula);
```

Important lifetime note from `z3_api.h:5144-5146`:

> The result buffer is statically allocated by Z3. It will be automatically deallocated when `Z3_del_context` is invoked. So, the buffer is invalidated in the next call to `Z3_benchmark_to_smtlib_string`.

→ We MUST `C.GoString(...)` the result *immediately* (inside the same `s.ctx.do(...)` block), before any other C call could clobber the static buffer.

## Goal

Make Go's `Z3Solver.TraceCheck` produce the **same byte-for-byte smt2 string** as Python's `_trace_z3_check`. After the fix, the leaf and root hashes for `z3.check seq=1 result=unsat` will match between Go and Python.

## Approach

Add a `ToSmt2()` method on `Z3Solver` that mirrors Python's `to_smt2()` exactly, then point `TraceCheck` at it instead of `String()`.

### Step 1 — Add `ToSmt2()` on `*Z3Solver`

**File:** `~/ivy/goivy/z3bridge/quantifier.go`

Add a new method (place it immediately after `String()` at line 1501):

```go
// ToSmt2 returns the solver's assertions in SMT-LIB2 benchmark format,
// matching Python's z3.Solver.to_smt2() exactly. This is used by
// TraceCheck so that Go and Python produce the same byte string and
// thus the same Merkle leaf hash for the z3.check trace event.
//
// Mirrors ~/ivy/goivy/z3vendor/z3/src/api/python/z3/z3.py:6495-6509:
//   - All-but-last assertion become "assumptions"
//   - Last assertion becomes the "formula"
//   - If there are 0 assertions, formula = Z3_mk_true()
//   - Benchmark name is the literal "benchmark generated from python API"
//   - status is the literal "unknown" (NOT the actual check result)
//   - logic and attributes are both ""
func (s *Z3Solver) ToSmt2() string {
    var res string
    s.ctx.do(func() {
        // Get all asserted formulas as an AST vector.
        vec := C.Z3_solver_get_assertions(s.ctx.c, s.c)
        C.Z3_ast_vector_inc_ref(s.ctx.c, vec)
        defer C.Z3_ast_vector_dec_ref(s.ctx.c, vec)

        n := int(C.Z3_ast_vector_size(s.ctx.c, vec))

        // Mirror Python: split off the last assertion as the "formula";
        // the rest become "assumptions". If empty, formula = true.
        var nAssumptions int
        if n > 0 {
            nAssumptions = n - 1
        }

        var formula C.Z3_ast
        if n > 0 {
            formula = C.Z3_ast_vector_get(s.ctx.c, vec, C.uint(nAssumptions))
        } else {
            formula = C.Z3_mk_true(s.ctx.c)
        }

        // Build the assumptions array.
        var assumptionsPtr *C.Z3_ast
        if nAssumptions > 0 {
            assumptions := make([]C.Z3_ast, nAssumptions)
            for i := 0; i < nAssumptions; i++ {
                assumptions[i] = C.Z3_ast_vector_get(s.ctx.c, vec, C.uint(i))
            }
            assumptionsPtr = &assumptions[0]
            // Keep the slice alive past the C call.
            defer runtime.KeepAlive(assumptions)
        }

        // Literal strings matching Python exactly. The name "benchmark
        // generated from python API" is what produces the
        // "; benchmark generated from python API" header line.
        cName := C.CString("benchmark generated from python API")
        defer C.free(unsafe.Pointer(cName))
        cLogic := C.CString("")
        defer C.free(unsafe.Pointer(cLogic))
        cStatus := C.CString("unknown")
        defer C.free(unsafe.Pointer(cStatus))
        cAttrs := C.CString("")
        defer C.free(unsafe.Pointer(cAttrs))

        // Z3 returns a static buffer; copy immediately into a Go string
        // before any other Z3 call can invalidate it.
        cstr := C.Z3_benchmark_to_smtlib_string(
            s.ctx.c,
            cName, cLogic, cStatus, cAttrs,
            C.uint(nAssumptions),
            assumptionsPtr,
            formula,
        )
        res = C.GoString(cstr)
    })
    runtime.KeepAlive(s)
    return res
}
```

**Notes on the implementation:**

- Wrapped in `s.ctx.do(func() {...})` to serialize Z3 context access (matches the pattern used by every other Z3 method in this file).
- `runtime.KeepAlive(assumptions)` ensures the Go slice backing the C array isn't GC'd before `Z3_benchmark_to_smtlib_string` returns. This is the same pattern as `runtime.KeepAlive(assumptions)` at `quantifier.go:1612`.
- The `Z3_ast` values stored in `assumptions[]` and `formula` come from `Z3_ast_vector_get`, which (per Z3 docs) returns a borrowed reference owned by the vector. We hold the vector ref via `inc_ref`/`defer dec_ref`, so the AST values stay valid for the duration.
- We do NOT additionally `Z3_inc_ref` the individual assertions — they're owned by the vector for the duration of our use.
- The static-buffer warning is the reason `res = C.GoString(cstr)` happens *immediately* inside the `do(...)` block, before any other C call could invalidate the buffer.

**Imports needed in quantifier.go:** `unsafe` is already imported (used for `C.CString` cleanup elsewhere). `runtime` is already imported. No new imports.

### Step 2 — Switch `TraceCheck` to use `ToSmt2()`

**File:** `~/ivy/goivy/z3bridge/xtrace.go:61`

Change one line:

```go
// before
smt2 := NormalizeZ3VarNames(s.String())

// after
smt2 := NormalizeZ3VarNames(s.ToSmt2())
```

This is the only change in `xtrace.go`. The canon string format, the Merkle chain, and the `xtracer.Trace` call are unchanged.

### Why no Python changes are needed

Python is the source of truth. Python is already calling `s.to_smt2()`. Python's behavior is correct. We are aligning Go to match Python.

## Files affected

### Modified files (no new files)

- `~/ivy/goivy/z3bridge/quantifier.go` — add `ToSmt2()` method (~60 lines, after the existing `String()` method)
- `~/ivy/goivy/z3bridge/xtrace.go` — change one line in `TraceCheck` to call `ToSmt2()` instead of `String()`

Total: ~60 lines added, 1 line changed, across 2 files. Pure additive change to public API; no signature changes; no removed code.

### Files to read before editing

- `~/ivy/goivy/z3bridge/quantifier.go:700-710` — `Expr` struct (`.c` field is `C.Z3_ast`)
- `~/ivy/goivy/z3bridge/quantifier.go:1493-1501` — existing `String()` (the broken model to replace)
- `~/ivy/goivy/z3bridge/quantifier.go:1616-1633` — `UnsatCore()` (the closest reference for the AST-vector-iteration pattern; mirror its structure exactly for consistency)
- `~/ivy/goivy/z3bridge/quantifier.go:9-24` — CGO header block (confirms `<z3.h>` is included → no extra declarations needed)
- `~/ivy/goivy/z3bridge/xtrace.go:54-76` — `TraceCheck` (the one-line change site)
- `~/ivy/goivy/z3vendor/z3/src/api/z3_api.h:5159-5166` — `Z3_benchmark_to_smtlib_string` signature
- `~/ivy/goivy/z3vendor/z3/src/api/z3_api.h:6024` — `Z3_solver_get_assertions` signature
- `~/ivy/goivy/z3vendor/z3/src/api/python/z3/z3.py:6495-6509` — Python's `to_smt2()` (the source of truth being mirrored)
- `~/ivy/pyivy/ivy/ivy/ivy_solver.py:1314-1346` — `_trace_z3_check` (the Python trace path being mirrored on Go side via `TraceCheck`)

## Verification

### Step A — Compile

```
cd ~/ivy/goivy && go build ./z3bridge/...
```

If CGO can't find `Z3_solver_get_assertions` or `Z3_benchmark_to_smtlib_string`, the build will fail. Both are declared in `<z3.h>`, which is already included, so this should succeed.

### Step B — Run z3bridge unit tests

```
cd ~/ivy/goivy && go test ./z3bridge/...
```

Trace-only changes should not break any existing tests. If any test calls `Z3Solver.String()` (the s-expression sexpr format) and inspects the output, it will still work — we haven't changed `String()`, only added `ToSmt2()` and switched the *trace* to use it.

### Step C — Microtest the new method (optional but quick)

Write or run a tiny test that:
1. Creates a `Z3Solver`
2. Asserts `x > 0` (or any simple constraint)
3. Calls `s.ToSmt2()`
4. Verifies the returned string starts with `; benchmark generated from python API`

This catches format-string bugs (e.g., wrong arg order to `Z3_benchmark_to_smtlib_string`) before re-running the full golden suite.

### Step D — Re-run `make golden`

```
cd ~/ivy/goivy && make golden 2>&1 | tee log.parser.red
```

The line numbers in `~/ivy/goivy/log.red` may shift slightly because the multi-line smt2 (now matching Python's format) replaces the previous single-line-style sexpr output and contains a different number of newlines.

**Expected outcome:** the previous divergence at 254714 is gone. Both sides now show:

```
NNNNNN  go : XTRACE: z3.check seq=1 result=unsat HASH leaf=blake3.33B-XXX root=blake3.33B-YYY smt2=; benchmark generated from python API
        py : XTRACE: z3.check seq=1 result=unsat HASH leaf=blake3.33B-XXX root=blake3.33B-YYY smt2=; benchmark generated from python API
```

The leaf and root hashes will match (because the underlying full multi-line smt2 strings are now byte-equal between Go and Python). The visible smt2 in the diff will show only the first line `; benchmark generated from python API` for both — which is no worse than the pre-fix Python trace.

**Two scenarios to expect:**

1. **Hashes match → divergence advances past 254714** to whatever the next divergence is. This is success — proceed to fix the next divergence in a follow-up plan.

2. **Hashes still differ** → the smt2 strings are byte-different in some non-obvious way. Likely causes:
   - `Z3_ast` values for assertions are returned in a different order between Go and Python (would require sorting; unlikely since both use `Z3_solver_get_assertions` which preserves insertion order)
   - Z3's pretty-printer is sensitive to context state (e.g., previous calls to `Z3_solver_to_string` influence subsequent calls' name allocation). Mitigation: don't call `s.String()` anywhere in the test session before `TraceCheck` runs.
   - `_normalize_z3_varnames` regex mismatch — both use `!k!\d+`, so this should be identical.
   - Encoding difference — `C.GoString` interprets the C string as UTF-8; this matches Python's default.

   To diagnose: temporarily add a tracer that dumps the full multi-line smt2 to a side file (`/tmp/go.smt2` and `/tmp/py.smt2`), then `diff` them.

## Risks and rollback

- **Risk: static-buffer invalidation.** Z3 documents that the result buffer is invalidated by the next call to `Z3_benchmark_to_smtlib_string`. We mitigate by calling `C.GoString` immediately inside the same `s.ctx.do(...)` block. If multiple goroutines were to call `ToSmt2()` concurrently on different solvers in the same context, the second call could clobber the first's buffer mid-copy. The `s.ctx.do(...)` lock prevents this.
- **Risk: AST-vector reference counting.** If we forget to `inc_ref` the vector or release it before reading individual `Z3_ast` values, Z3 may free the asts mid-call. We mitigate by mirroring the exact `inc_ref`/`defer dec_ref` pattern from `UnsatCore()` at `quantifier.go:1621-1629`.
- **Risk: empty solver.** When the solver has 0 assertions, Python passes `BoolVal(True)` as the formula. We mirror with `Z3_mk_true(ctx)`. If we got this wrong (e.g., passed `nil`), Z3 would likely segfault.
- **Rollback:** revert two files. The change is purely additive (one new public method, one one-line trace change) and trivially revertible via `git checkout`.

## Out of scope

- **Removing the `String()` method or changing its callers.** `String()` (`Z3_solver_to_string`) may be useful elsewhere in goivy for human-readable solver dumps. Leave it alone. We only change which method `TraceCheck` calls.
- **Cosmetic newline-stripping in the trace.** After this fix, `golden_test.go` will still see only the first line of the smt2 (`; benchmark generated from python API`) because of the `ReadString('\n')` truncation. This means future divergences on the smt2 content won't be visible in the diff — only the leaf hash will signal them. We could optionally collapse newlines (`strings.ReplaceAll(smt2, "\n", " ")`) on **both** Go and Python sides to make the full smt2 visible on a single line. Defer that to a separate plan if/when we actually need the visibility.
- **Caching `ToSmt2()` results.** Each call goes back to Z3. This is fine for trace use (only fires once per `Check()`). No caching needed.
- **Adding `ToSmt2()` tests beyond Step C.** A microtest is enough; the full golden suite is the real verification.
- **Changing the canon format or Merkle hashing scheme.** Both sides agree on `(z3check seq=N result=R smt2=S)` and Blake3-33B; that's already aligned.

## Implementation checklist

When executing this plan:

- [ ] Read `quantifier.go:1493-1633` to confirm the structure of `String()`, `UnsatCore()`, and the surrounding methods.
- [ ] Add `ToSmt2()` method immediately after `String()` (at quantifier.go:1502-ish).
- [ ] Change `xtrace.go:61` from `s.String()` to `s.ToSmt2()`.
- [ ] `go build ./z3bridge/...` — fix any compile errors.
- [ ] `go test ./z3bridge/...` — verify no regressions in unit tests.
- [ ] `make golden` — confirm divergence at 254714 is gone and observe where it lands next.
- [ ] If hashes still differ, dump full smt2 from both sides to `/tmp` and diff.
- [ ] Mark task #28 as completed; if a new divergence appears, create a follow-up task.

## Implementation Progress (running log)

(Empty until execution begins.)
