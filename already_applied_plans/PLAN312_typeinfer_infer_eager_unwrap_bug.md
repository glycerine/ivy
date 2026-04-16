# PLAN: Fix TopSort panic in typeinfer NamedBinder case (`($l2s_s P0. f(P0))(P)`)

Created: 2026-04-16 (afternoon, second pass)

## Context

After applying the previous plan (removing the empty-checkers early-return guards) and running `goivy_check ../ivy-lang-examples/examples/liveness/tlb.ivy` standalone, the run progresses much further but now panics:

```
panic: TranslateSort: TopSort "TopSort" has no to_z3() equivalent in Python
```

(see `~/ivy/goivy/log.rog.tlb` tail). The panic fires in the inductive-invariant check ("Initialization must establish the invariant") while translating `invar386` at `tlb.ivy:956`:

```
invariant l2s_saved -> (
($l2s_s P0. active(P0))(P) <-> (
(($l2s_s X. pc_b1(X))(P) | ($l2s_s X. pc_m2(X))(P) | ($l2s_s X. pc_m3(X))(P) | ($l2s_s X. pc_m5(X))(P) |
($l2s_s X. pc_i1(X))(P) | ($l2s_s X. pc_r1(X))(P) | ($l2s_s X. pc_r2(X))(P))
))
```

The previous invariant `invar385` (line 953) — `l2s_saved -> (~pc_b1(sk0) & ~($l2s_s X. pc_b1(X))(sk0))` — passed because it has no free Variable. `invar386` is the first invariant that applies a `$l2s_s …` *named binder* to a *free Variable* (`P`).

### Root cause

`/Users/jaten/ivy/goivy/typeinfer/infer.go:507-559` is the `*logic.NamedBinder` arm of `InferSorts`. After inferring sorts for the bound variables and the body, it tries to build the binder's overall function sort like this:

```go
if len(n.Variables) > 0 {
    allSorts := append(varSorts, bodyRes.Sort)
    fsSorts := make([]logic.Sort, len(allSorts))
    for i, sv := range allSorts {
        if c := Unwrap(sv); c != nil {
            fsSorts[i] = c
        } else {
            fsSorts[i] = logic.NewTopSort()   // <-- BUG: premature TopSort
        }
    }
    fs, _ := logic.NewFunctionSort(fsSorts...)
    resultSort = Wrap(fs)
}
```

The bound variables' `varSorts[i]` are `*SortVar`s (created at line 510 via `NewSortVar()`). At the moment we run the loop above, those `SortVar`s have NOT yet been concretized into `logic.Sort`s — they have only been unified inside the body via union-find. `Unwrap` returns `nil` for any `*SortVar` (it only handles `*SortWrapper`; it does not follow the `SortVar.Instance` chain — see `typeinfer/sortvar.go:43-48`). So every still-unresolved bound-variable SortVar gets replaced by `TopSort`, baked into the binder's result `*logic.FunctionSort`.

Once the NamedBinder's sort is `processor* -> Boolean` corrupted to `TopSort -> Boolean`, the surrounding `Apply(NamedBinder, P)` cannot constrain `P`'s sort: `Unify(P_sortvar, TopSort)` succeeds vacuously and `P` ends up with `TopSort`. Later, in `check.NewBaseChecker` (`check/check.go:96-112`), `module.DualClauses` calls our witness `func(v *lg.Variable) lg.Expr { return module.VarToSkolem("@", v) }` which produces `lg.Const{Name: "@P", CSort: TopSort}` (`module/skolem.go:32-41`). Translating that constant in `z3bridge.translateVarOrConst` calls `TranslateSort(TopSort)` and panics at `z3bridge/translate.go:303`.

The 2-byte name in the failing stack frame (`(0xfdbafdab32e, 0x2)`) corroborates this: it is exactly `"@P"`.

### How Python avoids the bug

`~/ivy/pyivy/ivy/ivy/type_inference.py:245-262`:

```python
elif type(t) is NamedBinder:
    env = env.copy()
    env.update((v.name, SortVar()) for v in t.variables)
    xys = [infer_sorts(v, env) for v in t.variables]
    vars_s = [x for x,y in xys]
    vars_t = [y for x,y in xys]
    body_s, body_t = infer_sorts(t.body, env)
    return (
        FunctionSort(*(vars_s + [body_s])) if len(t.variables) > 0 else body_s,
        ...
    )
```

Python's `FunctionSort` is dynamic and holds `SortVar` objects directly. The result of `infer_sorts(NamedBinder)` is a function-sort whose elements *are still SortVars*. They get concretized later via the deferred `lambda` that the caller eventually invokes after all unification has run.

Go can't put `*SortVar` inside `logic.FunctionSort` (it requires `logic.Sort`), so the typeinfer package already provides a wrapper for exactly this case: `FunctionSortVar` (`typeinfer/sortvar.go:68-100`). The `*logic.Apply` arm of `InferSorts` already uses it (`infer.go:99-105`):

```go
allSorts := make([]SortOrVar, len(termSorts)+1)
copy(allSorts, termSorts)
allSorts[len(termSorts)] = resultSort
fsv := NewFunctionSortVar(allSorts...)
if err := Unify(funcRes.Sort, fsv); err != nil { ... }
```

`Unify` handles `FunctionSortVar ↔ FunctionSortVar` and `FunctionSortVar ↔ SortWrapper(FunctionSort)` (`typeinfer/unify.go:93-129`), and `ConvertFromSortVars` knows how to walk `FunctionSortVar` recursively when concretizing (`typeinfer/unify.go:159-169`). So returning a `FunctionSortVar` from the NamedBinder arm is already supported throughout the inference machinery.

## Fix

### Edit 1: `/Users/jaten/ivy/goivy/typeinfer/infer.go` lines 526-541

Replace the eager unwrap-or-TopSort block with a `FunctionSortVar` that preserves SortVar linkage, mirroring Python's `FunctionSort(*(vars_s + [body_s]))` and the existing `*logic.Apply` arm.

Before:
```go
var resultSort SortOrVar
if len(n.Variables) > 0 {
    allSorts := append(varSorts, bodyRes.Sort)
    fsSorts := make([]logic.Sort, len(allSorts))
    for i, sv := range allSorts {
        if c := Unwrap(sv); c != nil {
            fsSorts[i] = c
        } else {
            fsSorts[i] = logic.NewTopSort()
        }
    }
    fs, _ := logic.NewFunctionSort(fsSorts...)
    resultSort = Wrap(fs)
} else {
    resultSort = bodyRes.Sort
}
```

After:
```go
var resultSort SortOrVar
if len(n.Variables) > 0 {
    // Python type_inference.py:254-255 returns
    //     FunctionSort(*(vars_s + [body_s]))
    // where vars_s are still SortVars. Python's FunctionSort can hold
    // SortVars directly; Go's logic.FunctionSort cannot, so we use the
    // typeinfer wrapper FunctionSortVar that preserves SortVar linkage.
    // This matches the *logic.Apply arm above (see line 102).
    //
    // Eagerly unwrapping to logic.Sort here breaks invariants like
    //   ($l2s_s P0. active(P0))(P)
    // where P's sort must be inferred from the binder's domain — bound
    // variables' SortVars may still be un-concretized at this point and
    // would silently become TopSort, propagating to the free argument
    // and ultimately panicking in z3bridge.TranslateSort.
    allSorts := append(varSorts, bodyRes.Sort)
    resultSort = NewFunctionSortVar(allSorts...)
} else {
    resultSort = bodyRes.Sort
}
```

No other edits should be required: the `Concretize` closure (lines 542-558) already calls `vr.Concretize()` per bound variable, which itself runs `ConvertFromSortVars` on the (now-unified) SortVar — so the final concrete `*logic.Variable` and `*logic.NamedBinder` will end up with proper sorts.

## What we are NOT changing

- `Unwrap`, `Find`, `ConvertFromSortVars`, `Unify`, and `FunctionSortVar` are all already correct for this use case — no changes needed there.
- The z3bridge panic itself (`translate.go:303`) is the right behavior: it surfaces the bug rather than silently producing wrong Z3. Leave it.
- The witness-skolem code in `check/check.go` and `module/skolem.go` is also correct — it faithfully ports Python's `lambda v: lg.Symbol('@'+v.name, v.sort)`.

## Critical files

- `/Users/jaten/ivy/goivy/typeinfer/infer.go` (lines 507-559: NamedBinder arm of `InferSorts`) — the only edit
- `/Users/jaten/ivy/goivy/typeinfer/sortvar.go` (lines 68-100: `FunctionSortVar` definition) — read-only reference
- `/Users/jaten/ivy/goivy/typeinfer/unify.go` (lines 93-169: handling of `FunctionSortVar`) — read-only reference
- `/Users/jaten/ivy/goivy/typeinfer/infer.go` (lines 78-123: `*logic.Apply` arm — pattern to mirror) — read-only reference

## Reference (source of truth)

- `~/ivy/pyivy/ivy/ivy/type_inference.py:245-262` — Python `NamedBinder` arm

## Verification

**Do NOT run `make tlb` or the full `goivy_check ../ivy-lang-examples/examples/liveness/tlb.ivy` from this agent — those each take a long time. The user will run them on another machine.**

Local checks this agent will run:

1. `go build ./typeinfer/... ./check/... ./z3bridge/...` to confirm the edit compiles.
2. `go test ./typeinfer/...` to make sure existing typeinfer tests still pass (they exercise NamedBinder via tests in `typeinfer/infer_test.go` and `logic/formula_test.go`'s `l2s_s` examples). If a test newly fails because it expected the old `TopSort` fallback, that test was asserting the bug — we'll need to inspect it.

After the user runs `goivy_check` on the failing tlb.ivy:

3. Expected: the panic on `invar386` no longer fires; the inductive-invariant check continues.
4. If a *different* invariant later panics with the same `TopSort` message, that is likely the same bug surfacing in a different shape — re-run the analysis. If a different invariant FAILs (rather than panics), that is correctness territory and a separate plan.
