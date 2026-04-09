# Fix log.red divergence at step 236451: missing `functionsort()` trace inside `ltPred`

Created: 2026-04-09 08:05 PDT

## Context

`~/ivy/goivy/log.red` shows that `TestOrdLive` (golden xtrace comparison
between Python `ivy_check` and Go `goivy_check`) diverges at step
**236451**:

```
236450  go : XTRACE: ivy_solver.py:501 lt_pred() ENTER sort=index * index -> Boolean
        py : XTRACE: ivy_solver.py:501 lt_pred() ENTER sort=index * index -> Boolean

236451  go : XTRACE: ivy_solver.py:560 forall() ENTER nvars=1
        py : XTRACE: ivy_solver.py:279 functionsort() ENTER
```

Both sides agree at 236450 (they enter `lt_pred` with the same sort).
At 236451 Python emits `functionsort() ENTER` from inside `lt_pred`,
while Go skips ahead and emits the next outer event (`forall()`).
Everything before 236451 matches; this is the very first line where
the two trace streams disagree.

### Root cause

Python source of truth — `~/ivy/pyivy/ivy/ivy/ivy_solver.py:513–517`:

```python
def lt_pred(sort):
    if __debug__: xtracer.trace("ivy_solver.py:501 lt_pred() ENTER sort=%s" % sort)
    sym = ivy_logic.Symbol('<',sort)
    sig = sym.sort.to_z3()                # ← (1) functionsort() ENTER, EVERY call
    return z3.Function(solver_name(sym), *sig)
```

`FunctionSort.to_z3 = functionsort` (ivy_solver.py:304), and
`functionsort` (ivy_solver.py:278–283) emits `functionsort() ENTER`
unconditionally — there is no cache around it. Python's `lt_pred`
itself is **not memoized**: every invocation rebuilds the Z3
`Function` from scratch (Z3 dedups internally by name+signature).
The polymac lambda is `lambda s,x,y: z3.Or(x == y, lt_pred(s)(x,y))`
(ivy_solver.py:520), so `lt_pred(s)` is re-called on every `<=`
expansion, and `functionsort() ENTER` is re-emitted every time.

Go side — `z3bridge/translate.go:715–724`:

```go
func (t *Translator) ltPred(fs *lg.FunctionSort) FuncDecl {
    xtracer.Trace("ivy_solver.py:501 lt_pred() ENTER sort=%s", fs)
    fd, err := t.makeFuncDecl("<", fs)
    if err != nil {
        panic(fmt.Sprintf("ltPred: makeFuncDecl failed: %v", err))
    }
    return fd
}
```

…and `makeFuncDecl` at `translate.go:1043–1072`:

```go
func (t *Translator) makeFuncDecl(name string, fs *lg.FunctionSort) (FuncDecl, error) {
    key := lg.NodeKey(name + ":" + string(fs.Sexp()))
    if cached, ok := t.funcs[key]; ok {
        return cached, nil                            // ← cache HIT: trace skipped
    }
    xtracer.Trace("ivy_solver.py:279 functionsort() ENTER")  // ← only on miss
    domain := fs.Domain()
    ...
    fd := t.Ctx.Function(z3name, zDomain, zRange)
    t.funcs[key] = fd
    return fd, nil
}
```

`t.funcs` was populated by an *earlier* `<=` expansion on the same
`index * index -> Boolean` sort (the trace shows several earlier
`assumed` properties on `index.spec` that drove the same code path).
On this second-or-later call, `t.funcs[key]` is a hit, so
`makeFuncDecl` returns the cached `FuncDecl` **without emitting
`functionsort() ENTER`**. Python has no equivalent cache around the
`functionsort` trace, so Python emits it. That one missing line is
the divergence.

The dead `LtPred` in `solver_encoding.go` (already removed) was
unrelated — it was never called.

### Why this is a faithfulness bug, not a "harmless" trace skew

CLAUDE.md rule B.6 forbids omitting Python functions, and rule B.9
forbids "good enough" shortcuts. The xtrace stream is the contract
that proves the Go port is mechanically equivalent to Python at every
control-flow step. A skipped trace means a skipped Python control-flow
event — in this case, the call to `FunctionSort.to_z3` from inside
`lt_pred`. Whether or not the Z3 result happens to be identical, the
**flow** must match.

It also masks any real divergences that would happen later: as soon
as one trace line is wrong, the whole stream after that point is
suspect, and the golden test bails out at 236451 instead of catching
the (possibly different) bug that lurks at 240000.

### Why other call sites of `makeFuncDecl` are correct

The trace inside `makeFuncDecl` is (almost) right for the *other*
callers, because Python wraps those code paths in a higher-level
cache and only calls `to_z3()` (and thus `functionsort`) on cache
miss. Specifically:

| Go caller | Python analog | Cache used by Python |
|---|---|---|
| `(*Translator).getFuncDecl` (translate.go:1029, 1036) — from `term_to_z3`'s Apply branch | `term_to_z3` (ivy_solver.py:496–510) | `z3_functions[term.rep]` |
| `(*Translator).atomToZ3` (translate.go:655) — uninterpreted relation case | `atom_to_z3` (ivy_solver.py:529–547) | `z3_predicates[atom.relname]` |
| `(*Translator).TranslateConst` (translate.go:1007) — higher-order placeholder | (translates to `z3.Const` in Python; functionsort is not always invoked) | n/a |

For these, "trace inside the cache miss branch of `makeFuncDecl`"
matches "trace inside the cache miss branch of Python's higher-level
cache" — at least at the level of granularity the existing xtrace
checks. So we should **keep** the cache-miss tracing behaviour for
those call sites, and only fix `ltPred`.

(There is a latent secondary issue: Python uses *two* separate caches
`z3_functions` and `z3_predicates`, while Go uses a single `t.funcs`
shared across `getFuncDecl` and `atomToZ3`. This means a function
created via `term_to_z3` would skip the `functionsort()` trace on a
later `atom_to_z3` call where Python would emit it. We are **not**
addressing that here — it is a separate divergence and is not what
log.red is failing on. This plan is scoped to fix only step 236451.)

## Approach

Mechanically port Python's `functionsort` into a dedicated Go helper,
and have `ltPred` call it directly so the trace fires every time —
matching Python's uncached `sym.sort.to_z3()` call from inside
`lt_pred`. Keep the existing cache in `makeFuncDecl` for the other
call sites so their tracing behaviour stays correct.

### Why this approach (and not alternatives)

- **Alternative A — move the trace before the cache check in
  `makeFuncDecl`.** Rejected: this would over-emit at the
  `getFuncDecl` / `atomToZ3` call sites, where Python *does* skip the
  trace on cache hit. We would just push the divergence elsewhere.

- **Alternative B — add a `bypassCache` flag to `makeFuncDecl`.**
  Rejected: introduces an abstraction Python doesn't have (CLAUDE.md
  B.7) and conflates two unrelated concerns (cache control vs. trace
  emission).

- **Alternative C (chosen) — introduce a `functionSort` helper that
  mirrors Python's `functionsort` 1:1 (same name, same trace, same
  body), have `ltPred` call it directly without going through
  `makeFuncDecl`'s cache, and have `makeFuncDecl` reuse the same
  helper on cache miss for the other paths.** This is the literal
  port. `functionSort` becomes the single source of truth for "what
  Python's `functionsort` does", and the trace lives in exactly one
  place.

## Changes

### File: `z3bridge/translate.go`

**1. Add a new `functionSort` helper** mirroring Python's
`functionsort` (ivy_solver.py:278–283). Place it adjacent to
`makeFuncDecl` (around line 1043) so the two are easy to read
together:

```go
// functionSort mirrors Python's functionsort (ivy_solver.py:278-283).
// Translates a FunctionSort to its list of Z3 sorts: [domain..., range].
// ALWAYS emits the functionsort() ENTER trace — there is no caching
// here, matching Python's monkey-patched FunctionSort.to_z3 = functionsort
// (ivy_solver.py:304). Callers that want caching must wrap this themselves
// (e.g. makeFuncDecl).
func (t *Translator) functionSort(fs *lg.FunctionSort) ([]Sort, error) {
    xtracer.Trace("ivy_solver.py:279 functionsort() ENTER")

    domain := fs.Domain()
    sig := make([]Sort, 0, len(domain)+1)
    for _, d := range domain {
        zs, err := t.TranslateSort(d)
        if err != nil {
            return nil, err
        }
        sig = append(sig, zs)
    }

    // Python: if fs.is_relational(): return [...] + [z3.BoolSort()]
    //         else:                  return [...] + [fs.rng.to_z3()]
    // Note: Python short-circuits the rng.to_z3() call when relational,
    // so no extra trace fires for the Boolean range. We mirror that.
    var rng Sort
    if il.IsRelationalSort(fs) {
        rng = t.Ctx.BoolSort()
    } else {
        var err error
        rng, err = t.TranslateSort(fs.Range())
        if err != nil {
            return nil, err
        }
    }
    sig = append(sig, rng)
    return sig, nil
}
```

(`il` is already imported in `translate.go` as
`il "github.com/glycerine/ivy/goivy/ivylogic"`, and
`il.IsRelationalSort` already exists at `ivylogic/ivylogic.go:48`.)

**2. Refactor `makeFuncDecl`** to use `functionSort` so the trace lives
in exactly one place. The cache stays. Behaviour for the existing
`getFuncDecl` / `atomToZ3` / `TranslateConst` callers is unchanged
(trace fires on cache miss, suppressed on cache hit):

```go
func (t *Translator) makeFuncDecl(name string, fs *lg.FunctionSort) (FuncDecl, error) {
    key := lg.NodeKey(name + ":" + string(fs.Sexp()))
    if cached, ok := t.funcs[key]; ok {
        return cached, nil
    }

    sig, err := t.functionSort(fs)
    if err != nil {
        return FuncDecl{}, err
    }
    zDomain := sig[:len(sig)-1]
    zRange := sig[len(sig)-1]

    z3name := t.z3Name(name, fs)
    fd := t.Ctx.Function(z3name, zDomain, zRange)
    t.funcs[key] = fd
    return fd, nil
}
```

This deletes the now-redundant `xtracer.Trace("ivy_solver.py:279 functionsort() ENTER")`
at translate.go:1051 — it moves into `functionSort`.

**3. Rewrite `ltPred`** to bypass `makeFuncDecl` entirely, mirroring
Python's uncached `lt_pred`:

```go
// ltPred creates a Z3 function declaration for "<" on the given sort.
// Mirrors Python lt_pred (ivy_solver.py:513-517):
//
//     def lt_pred(sort):
//         sym = ivy_logic.Symbol('<', sort)
//         sig = sym.sort.to_z3()                  # functionsort() ENTER
//         return z3.Function(solver_name(sym), *sig)
//
// Python does NOT cache the result. Each invocation re-emits the
// functionsort() trace and rebuilds the Z3 Function. Z3 dedups
// internally by (name, signature), so multiple calls return
// equivalent FuncDecls. We mirror that behaviour exactly here —
// going through makeFuncDecl's cache would suppress the
// functionsort() ENTER trace on hits and break xtrace alignment
// with Python (this was the cause of the log.red divergence at
// step 236451).
func (t *Translator) ltPred(fs *lg.FunctionSort) FuncDecl {
    xtracer.Trace("ivy_solver.py:501 lt_pred() ENTER sort=%s", fs)
    sig, err := t.functionSort(fs)
    if err != nil {
        panic(fmt.Sprintf("ltPred: functionSort failed: %v", err))
    }
    zDomain := sig[:len(sig)-1]
    zRange := sig[len(sig)-1]
    z3name := t.z3Name("<", fs)
    return t.Ctx.Function(z3name, zDomain, zRange)
}
```

Notes:
- The Symbol object Python creates (`sym = Symbol('<', sort)`) is
  invisible to the trace stream — Symbol construction doesn't emit
  any xtrace. So we don't need to construct an `lg.Const` here; we
  just need the *name* `"<"` for `solver_name`/`z3Name`. This matches
  Python's effective behaviour.
- We must compute `z3name` using the same path that the cached
  version would have, so that the resulting Z3 `FuncDecl` has the
  same name string as before. `t.z3Name("<", fs)` is what
  `makeFuncDecl` already uses.

**4. Decide what to do with the dead `xtracer.Trace` at translate.go:293**
inside `TranslateSort`'s `*lg.FunctionSort` case:

```go
case *lg.FunctionSort:
    xtracer.Trace("ivy_solver.py:279 functionsort() ENTER")
    return Sort{}, fmt.Errorf("FunctionSorts are not directly converted to Z3 sorts")
```

This branch only fires if some caller passes a `FunctionSort`
directly to `TranslateSort` — which would be a bug, and the function
returns an error immediately. The trace there is misleading (it
suggests a successful `functionsort()` step that never actually
produced sorts). Since we now have a real `functionSort` helper that
does the right thing, **delete the `xtracer.Trace` line** from this
error branch (keep the `return ... error`). This is a small cleanup,
not a correctness fix; included because it's a one-line change in
the same area and reduces the risk of someone "wiring up" that
branch later.

### Files NOT modified

- `z3bridge/solver_encoding.go` — already cleaned up (dead `LtPred`
  removed by user).
- All other Go files — no other call site of `ltPred` exists.
- Python source — it is the source of truth and is correct.

## Verification

Run from `~/go/src/github.com/glycerine/ivy/goivy`:

1. **Build**:
   ```
   go build ./z3bridge/...
   go build -o ~/go/bin/goivy_check_xtrace ./cmd/goivy_check
   ```

2. **Unit tests**:
   ```
   go test ./z3bridge/...
   ```
   Particularly anything in `solver_test.go` that exercises `<=`,
   `>`, `>=`, or polymacs.

3. **Re-run the failing golden test**:
   ```
   go test ./parser/ -run TestOrdLive -v
   ```
   - **Expected**: the divergence at step 236451 disappears.
   - The test will likely run further before either passing or
     hitting the *next* divergence. Either outcome is progress.
   - If a new divergence appears at a step **earlier** than 236451,
     something is wrong with the fix — investigate before iterating.

4. **Manual xtrace inspection** of `log.red` after a fresh run:
   - At what was step 236451, both sides should now show
     `ivy_solver.py:279 functionsort() ENTER`.
   - The line counts should advance further before any disagreement.

5. **Quick sanity check for double-emission**: scan a few
   `term_to_z3` Apply paths in the new log to confirm we did not
   accidentally start over-emitting `functionsort() ENTER` at
   `getFuncDecl` cache hits. The behaviour at those call sites must
   be unchanged — the cache in `makeFuncDecl` still suppresses the
   trace on hits, exactly as before.

## Critical files

- `z3bridge/translate.go` — all three edits live here:
  - new `functionSort` helper
  - refactored `makeFuncDecl` (uses `functionSort`)
  - rewritten `ltPred` (uses `functionSort`, bypasses cache)
  - small cleanup of misleading trace in `TranslateSort` FunctionSort branch
- `~/ivy/pyivy/ivy/ivy/ivy_solver.py:278–283, 304, 513–517` —
  source of truth; do not modify, but cite in comments.
- `~/ivy/goivy/log.red` — the failing golden trace; will be
  regenerated by step 3 of verification.
