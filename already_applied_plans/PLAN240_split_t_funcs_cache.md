# Plan: Split and Rename Go's Z3 Caches to Match Python's z3_functions / z3_predicates

Created: 2026-04-09 05:16 -03

## Context

Python Ivy keeps two independent Z3 FuncDecl caches at module scope
(`ivy_solver.py:251-255`):

- `z3_functions` — populated only by `term_to_z3` (lines 500-508)
- `z3_predicates` — populated only by `atom_to_z3` (lines 533-542)

Because they are disjoint, Python's `functionsort()` ENTER xtrace
(emitted unconditionally at `ivy_solver.py:279`) fires once per
cache-miss on each path.

Go's `Translator` collapsed both into a **single** `funcs` cache keyed by
`name + ":" + sort.Sexp()`. That cache is written by `makeFuncDecl`
(`translate.go:1103`), which is currently called from four sites:

1. `atomToZ3` line **659** — atom path (uninterpreted relations).
2. `translateVarOrConst` line **1027** — higher-order bare Const
   branch; creates the FuncDecl purely for the `funcs` side-effect
   (`_ = fd`) and returns a plain Const.
3. `getFuncDecl` line **1049** — term path (Const).
4. `getFuncDecl` line **1056** — term path (Variable).

Callers 1 and 2 are on Python's atom / symbol-bare paths, where Python
does **not** consult `z3_functions`. Sharing a single Go cache causes a
prior term-path or symbol-bare write to suppress the `functionsort()`
ENTER xtrace that Python would emit on a later atom-path miss — the
exact divergence the user reported ("a function created via term_to_z3
would skip the functionsort() trace on a later atom_to_z3 call where
Python would emit it"). This is the same class of bug `ltPred`
(`translate.go:727-733`) already documents around the "log.red
divergence at step 236451".

Desired outcome: `t.funcs` becomes exclusively the **z3_functions**
cache (term path only). The atom path and the symbol-bare path
construct their FuncDecls inline — matching Python literally — and no
longer touch `t.funcs`. Dedup on the atom path is already handled by
the existing `t.preds` closure cache (which maps to Python
`z3_predicates`).

This mirrors the existing `ltPred` precedent (`translate.go:734-744`),
which already inlines `functionSort` + `Ctx.Function` for exactly this
reason.

In addition, the Go port has been obscuring these two caches by
renaming them (`funcs` / `preds`) when Python's names (`z3_functions`
/ `z3_predicates`) are clearer and make the Python correspondence
obvious at every call site. We rename the Go fields back to match
Python literally, using the identical snake_case identifiers
`z3_functions` and `z3_predicates`. This intentionally departs from
idiomatic Go camelCase in favor of mechanical-port clarity (CLAUDE.md
rules 3 and 8).

## Approach

Inline the Z3 FuncDecl construction at both offending call sites so
that `t.funcs` is written only from the two `getFuncDecl` paths
(term_to_z3). No new caches, no new helpers, no parameterization — the
most literal Python mirror per CLAUDE.md rules 7 and 8. Python's
`atom_to_z3` inlines `z3.Function(...)` (line 541), and `symbol_to_z3`
(lines 299-301) does the same and does not cache at all. `t.preds`
remains the z3_predicates analog and already handles atom-path dedup
via closure capture; no FuncDecl-level cache is required for the atom
path. Similarly, `translateVarOrConst`'s symbol-bare branch should not
cache — matching `symbol_to_z3`.

## Critical Files

- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/z3bridge/translate.go`
  — the only file modified.
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_solver.py` — Python source of
  truth. Reference lines: 249-257 (clear), 278-283 (functionsort),
  299-301 (symbol_to_z3), 457-511 (term_to_z3), 529-547 (atom_to_z3).

## Changes

> Note on change ordering: apply Change 0 (the rename) first so that
> subsequent inline-rewrite changes touch fields with their final
> names. Alternatively, do Changes 1-5 first against the current
> `funcs`/`preds` names and finish with Change 0 — either order
> works but the file-wide rename is mechanically simplest to do as a
> single pass.

### Change 0 — Rename `funcs` → `z3_functions`, `preds` → `z3_predicates`

**Rationale:** the current Go names obscure the Python correspondence.
Python uses `z3_functions` / `z3_predicates`; the Go port should use
the same identifiers (with the literal underscore, not camelCase) so
that every reader immediately recognizes the mapping. Yes, this
violates idiomatic Go naming — but CLAUDE.md rule 8 ("When in doubt,
translate literally") and rule 3 ("SAME NAME") both favor literal
correspondence over Go idiom for this mechanical port.

**File:** `translate.go`

- **Line 46** (struct field declaration):
  ```go
  funcs           map[lg.NodeKey]FuncDecl                // ...
  ```
  →
  ```go
  z3_functions    map[lg.NodeKey]FuncDecl                // Python z3_functions (term_to_z3 FuncDecl cache; NOT used by atom or symbol-bare paths)
  ```

- **Line 47** (struct field declaration):
  ```go
  preds           map[lg.NodeKey]func(args ...Expr) Expr // ...
  ```
  →
  ```go
  z3_predicates   map[lg.NodeKey]func(args ...Expr) Expr // Python z3_predicates (atom_to_z3 closure cache; stores FuncDecl.Apply / native / polymac / my_eq)
  ```

  (These comment updates also subsume Changes 3 and 4 below — see note
  at the end of this section.)

- **Lines 187-188** (`NewTranslator`):
  ```go
  funcs:    make(map[lg.NodeKey]FuncDecl),
  preds:    make(map[lg.NodeKey]func(args ...Expr) Expr),
  ```
  →
  ```go
  z3_functions:  make(map[lg.NodeKey]FuncDecl),
  z3_predicates: make(map[lg.NodeKey]func(args ...Expr) Expr),
  ```

- **Lines 204-205** (`Clear`):
  ```go
  t.funcs = make(map[lg.NodeKey]FuncDecl)
  t.preds = make(map[lg.NodeKey]func(args ...Expr) Expr)
  ```
  →
  ```go
  t.z3_functions = make(map[lg.NodeKey]FuncDecl)
  t.z3_predicates = make(map[lg.NodeKey]func(args ...Expr) Expr)
  ```

- **Line 212** (`initEqPred`):
  ```go
  t.preds[eqCanonPredKey] = func(args ...Expr) Expr {
  ```
  →
  ```go
  t.z3_predicates[eqCanonPredKey] = func(args ...Expr) Expr {
  ```

- **Line 602** (atomToZ3 cache lookup):
  ```go
  if cached, ok := t.preds[predKey]; ok {
  ```
  →
  ```go
  if cached, ok := t.z3_predicates[predKey]; ok {
  ```

- **Line 618** (atomToZ3 native relation cache store):
  ```go
  t.preds[predKey] = nativeFn
  ```
  →
  ```go
  t.z3_predicates[predKey] = nativeFn
  ```

- **Line 638** (atomToZ3 polymac cache store):
  ```go
  t.preds[predKey] = predFn
  ```
  →
  ```go
  t.z3_predicates[predKey] = predFn
  ```

- **Line 650** (atomToZ3 `=` cache store):
  ```go
  t.preds[predKey] = eqFn
  ```
  →
  ```go
  t.z3_predicates[predKey] = eqFn
  ```

- **Line 664** (atomToZ3 uninterpreted relation cache store — will be
  moved/rewritten by Change 1, but still uses the renamed field):
  ```go
  t.preds[predKey] = predFn
  ```
  →
  ```go
  t.z3_predicates[predKey] = predFn
  ```

- **Line 1105** (`makeFuncDecl` cache lookup):
  ```go
  if cached, ok := t.funcs[key]; ok {
  ```
  →
  ```go
  if cached, ok := t.z3_functions[key]; ok {
  ```

- **Line 1121** (`makeFuncDecl` cache store):
  ```go
  t.funcs[key] = fd
  ```
  →
  ```go
  t.z3_functions[key] = fd
  ```

**File:** `interp.go`

- **Line 100** (`NewTranslatorWithInterpolation`):
  ```go
  funcs:  make(map[logic.NodeKey]FuncDecl),
  ```
  →
  ```go
  z3_functions:  make(map[logic.NodeKey]FuncDecl),
  ```

  Pre-existing bug worth calling out but out of scope: this
  constructor does not initialize `sortsInv`, `z3_predicates`, or call
  `initEqPred`. Any attempt to translate a predicate through a
  Translator built here would nil-panic on `z3_predicates`. Leave
  as-is for this plan (noted for future cleanup).

**Note:** Changes 3 and 4 below (the `funcs` and `preds` field
comment updates) are now subsumed by Change 0; their entries remain in
the list for completeness but can be skipped if Change 0 is applied
first. Change 5 (the `makeFuncDecl` docstring) is still needed.

### Change 1 — `atomToZ3` inline FuncDecl creation (`translate.go:653-665`)

**Before:**

```go
    // Python lines 527-528: create Z3 Function/Const for uninterpreted relation
    fs, ok := c.CSort.(*lg.FunctionSort)
    if !ok {
        return Expr{}, fmt.Errorf("atomToZ3: expected FunctionSort for %s, got %T", c.Name, c.CSort)
    }
    fd, err := t.makeFuncDecl(c.Name, fs)
    if err != nil {
        return Expr{}, err
    }
    predFn := fd.Apply
    t.preds[predKey] = predFn
    return t.applyZ3Func(predFn, app.Terms) // end of atomToZ3 here.
}
```

**After:**

```go
    // Python ivy_solver.py:540-541:
    //     sig = atom.rep.sort.to_z3()
    //     rel = z3.Function(solver_name(atom.rep), *sig) if isinstance(sig,list)
    //           else z3.Const(solver_name(atom.rep),sig)
    //
    // Python's atom_to_z3 populates the z3_predicates cache (mirrored
    // here by t.z3_predicates); it NEVER touches z3_functions. We
    // must not go through makeFuncDecl (which caches in
    // t.z3_functions) — sharing that cache causes a hit from a prior
    // term_to_z3 or symbol_to_z3 path to suppress the functionsort()
    // ENTER xtrace, diverging from Python. Same reasoning and
    // precedent as ltPred (see its comment referencing log.red step
    // 236451).
    fs, ok := c.CSort.(*lg.FunctionSort)
    if !ok {
        return Expr{}, fmt.Errorf("atomToZ3: expected FunctionSort for %s, got %T", c.Name, c.CSort)
    }
    sig, err := t.functionSort(fs) // emits "ivy_solver.py:279 functionsort() ENTER"
    if err != nil {
        return Expr{}, err
    }
    zDomain := sig[:len(sig)-1]
    zRange := sig[len(sig)-1]
    z3name := t.z3Name(c.Name, fs)
    fd := t.Ctx.Function(z3name, zDomain, zRange)
    predFn := fd.Apply
    t.z3_predicates[predKey] = predFn
    return t.applyZ3Func(predFn, app.Terms) // end of atomToZ3 here.
}
```

Notes:

- Ordering matches Python lines 540-541: `functionSort` first (trace
  fires), then `z3Name`, then `Ctx.Function`.
- Python's `isinstance(sig,list)` alternate branch (`z3.Const`) is
  unreachable here: the enclosing type assertion guarantees `fs` is a
  `FunctionSort`, so `functionSort` always returns a slice — matching
  Python's behavior at this exact call site.
- `applyZ3Func(predFn, app.Terms)` is the same call as before with the
  same `predFn := fd.Apply`.

### Change 2 — `translateVarOrConst` inline higher-order bare Const (`translate.go:1024-1037`)

**Before:**

```go
    // Higher-order: return a placeholder constant.
    // The actual FuncDecl is created in getFuncDecl.
    if fs, ok := sort.(*lg.FunctionSort); ok {
        fd, err := t.makeFuncDecl(name, fs)
        if err != nil {
            return Expr{}, err
        }
        _ = fd
        zs, err := t.TranslateSort(fs.Range())
        if err != nil {
            return Expr{}, err
        }
        return t.Ctx.Const(z3name, zs), nil
    }
```

**After:**

```go
    // Higher-order bare Const: mirror Python symbol_to_z3
    // (ivy_solver.py:299-301), which creates a z3.Function inline and
    // does NOT cache anywhere. We call functionSort directly so the
    // "functionsort() ENTER" xtrace fires every time (matching Python,
    // where symbol_to_z3 has no cache). Going through makeFuncDecl
    // would write t.z3_functions (Python's z3_functions cache), which
    // symbol_to_z3 does not populate in Python, and would also
    // suppress the trace on subsequent visits.
    if fs, ok := sort.(*lg.FunctionSort); ok {
        if _, err := t.functionSort(fs); err != nil { // emits functionsort() ENTER
            return Expr{}, err
        }
        zs, err := t.TranslateSort(fs.Range())
        if err != nil {
            return Expr{}, err
        }
        return t.Ctx.Const(z3name, zs), nil
    }
```

Rationale: the current `fd, _ = fd` pattern creates a FuncDecl purely
to populate `funcs` and emit the trace on the first visit. Python's
`symbol_to_z3` doesn't cache, so the trace fires every visit. Inlining
`functionSort` preserves the trace on every call and removes the
cache-pollution side effect that drove the atomToZ3 divergence in the
first place.

### Changes 3 & 4 — Field comments

Subsumed by Change 0 (which rewrites the field declarations at lines
46-47 with the new names and new comments in a single edit). No
separate action required.

### Change 5 — `makeFuncDecl` docstring (insert above `translate.go:1103`)

Add a new docstring directly above the `func (t *Translator) makeFuncDecl(...)` line:

```go
// makeFuncDecl returns a Z3 FuncDecl for (name, fs), caching the
// result in t.z3_functions.
//
// This is the TERM-PATH helper only — the Go analog of Python's
// z3_functions cache (ivy_solver.py:500-508). Callers on the atom
// path (atomToZ3) and the symbol-bare path (translateVarOrConst
// higher-order branch) MUST NOT call this helper; they must build
// the FuncDecl inline via t.functionSort + t.Ctx.Function so that
// Python's independent z3_predicates / symbol_to_z3 xtrace semantics
// are preserved. Sharing this cache across paths suppresses the
// functionsort() ENTER trace on downstream misses and diverges from
// Python. See atomToZ3, translateVarOrConst, and ltPred for the
// inline-creation precedent.
```

## Invariants After the Change

- `t.z3_functions` is written exclusively by `makeFuncDecl`, which is
  called only by `getFuncDecl` (callers at `translate.go:1049` and
  `1056`, the term-path). `t.z3_functions` ↔ Python `z3_functions`.
- `t.z3_predicates` is written exclusively by `atomToZ3` (and
  `initEqPred`). `t.z3_predicates` ↔ Python `z3_predicates`.
- `translateVarOrConst`'s higher-order branch writes no cache; it
  emits `functionsort() ENTER` on every invocation, matching
  `symbol_to_z3`.
- `initEqPred` (`translate.go:209-218`) is unchanged; the equality
  closure belongs in `preds`.
- `ltPred` (`translate.go:734`) is unchanged; it was already inlined
  for this reason.

## Verification

1. **Grep sanity:** after editing, `Grep "makeFuncDecl" translate.go`
   should list exactly three callers: lines ~1049 and ~1056 (both in
   `getFuncDecl`) plus the definition itself. No matches inside
   `atomToZ3` or `translateVarOrConst`.
2. **Grep sanity:** `Grep "t\.z3_functions\[" translate.go` should
   show writes/reads only inside `makeFuncDecl`. `Grep "t\.funcs|t\.preds"`
   across the whole tree should return zero results (the rename is
   complete).
3. **Build:** `cd ~/go/src/github.com/glycerine/ivy/goivy && go build ./z3bridge/...`
4. **Unit tests:** `go test ./z3bridge/...`. Expect no regressions. Any
   golden test that exercises uninterpreted relations may gain
   additional `functionsort() ENTER` traces (the very fix we want).
5. **Xtrace log diff (primary correctness check):** run the Go/Python
   comparison harness used historically to detect the step 236451
   divergence, and diff the resulting `log.red` against the Python
   reference (and against the checked-in `log.red.bk` baseline at the
   repo root). Expected: previously-suppressed `functionsort() ENTER`
   lines at atom-path and symbol-bare-path cache-miss sites now appear
   in the Go log. No new Go-only extra traces should appear (the
   `preds` closure cache still prevents repeat atom translations for
   the same relname).
6. **Regression check on the original divergence point:** the user's
   most recent known divergence was past step 236451 (most recent
   golden tagged 236914 per `git log`). Confirm the new Go log matches
   Python at and beyond that boundary.
7. **Full-run golden update:** if traces are now more aligned, rerun
   and refresh any golden files under `z3bridge/` that record expected
   xtrace output (but only after steps 1-6 pass cleanly).

## Risks / Edge Cases

- **Error propagation:** both inlined blocks propagate `functionSort`
  errors via `return Expr{}, err`, matching the prior `makeFuncDecl`
  call. No new panic sites.
- **Reordering:** `functionSort` must be called before `z3Name` /
  `Ctx.Function` so the trace order matches Python's
  `sig = atom.rep.sort.to_z3(); rel = z3.Function(solver_name(...), *sig)`.
  Both Change 1 and Change 2 preserve this ordering.
- **Polymorphic symbols:** if the same `name + sort.Sexp()` is ever
  reached by both the atom path (via `atomToZ3`) and the term path
  (via `getFuncDecl`), each path now independently misses its own
  cache on first hit and independently calls `functionSort`. Python
  exhibits the same behavior with its two separate dicts, so the fix
  matches.
- **`translateVarOrConst` downstream effects:** Change 2 removes the
  `t.z3_functions` side-effect. Since the current code already
  discards the FuncDecl (`_ = fd`), no caller depends on the returned
  FuncDecl — only on the `Const` return value, which is unchanged.
  The only observable behavior change is (a) `t.z3_functions` no
  longer gains an unrelated entry, and (b) the trace now fires on
  each visit, matching Python.
- **`Clear()` (`translate.go:200-207`):** updated by Change 0 to use
  the renamed fields; still allocates and clears both caches.
- **Snake_case in Go:** `z3_functions` / `z3_predicates` are
  intentionally non-idiomatic Go. `go vet` and `golint` may complain
  about the underscores. This is accepted as a cost of literal Python
  correspondence. Document the choice in the struct field comments
  (Change 0) so future readers understand it is deliberate.
- **Thread safety:** `Translator` is not concurrent-safe today and
  this change does not alter that.

## Non-goals / Rejected Alternatives

- **New `predFuncs` FuncDecl cache** — rejected. No Python
  counterpart; Python's `z3_predicates` stores closures/FuncDecls by
  relname, a role already filled by Go's `preds`. Adding a second
  FuncDecl cache would be a Go-only abstraction (violates rule 7).
- **Parameterized `makeFuncDecl(kind)`** — rejected. Not mechanical;
  Python has no such helper.
- **Go-idiomatic rename (`z3Functions` / `z3Predicates` camelCase)** —
  rejected in favor of literal snake_case `z3_functions` /
  `z3_predicates` per user direction and CLAUDE.md rule 8. The whole
  point of this rename is to maximize visual correspondence with the
  Python names; camelCase would partially defeat that.
