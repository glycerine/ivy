# PLAN: Make `compiler.SortInfer` enforce `check_concretely_sorted` like Python

Created: 2026-04-16 (afternoon, third pass)

## Context

After applying the previous plan (`typeinfer/infer.go` NamedBinder arm now returns `NewFunctionSortVar` instead of eagerly unwrapping to TopSort, committed as `8f1a4796 apply plan 312`), the user re-ran `XTRACE_OFF=1 goivy_check ../ivy-lang-examples/examples/liveness/tlb.ivy` and got the **same panic at the same place** (`~/ivy/goivy/log.goivy_only.tlb2`):

```
        ../ivy-lang-examples/examples/liveness/tlb.ivy: line 956: invar386 ...
panic: TranslateSort: TopSort "TopSort" has no to_z3() equivalent in Python
```

The 2-byte name in the failing stack frame `translateVarOrConst(0x..., {0x..., 0x2})` is again `"@P"` — the witness Skolem of a Variable named `P` with `VSort = *lg.TopSort`.

So somewhere between "the user wrote `($l2s_s P0. active(P0))(P)` in a tactic invariant" and "this conjecture reaches the Z3 translator," a `*lg.Variable` named `P` ends up with TopSort. The previous fix to `typeinfer/infer.go:507-545` was correct in isolation — `Unify(FunctionSortVar, FunctionSortVar)` propagates the inferred `processor` through the bound P0 to the free `P` — but a later step in the pipeline is still leaving a TopSort behind.

We need to (a) localize *where* the TopSort survives and (b) start mirroring Python's safety net for this entire class of bug.

### The missing safety net (and the immediate fix)

Python's `~/ivy/pyivy/ivy/ivy/ivy_logic.py:1203-1206` defines `sort_infer` as:

```python
def sort_infer(term,sort=None,no_error=False):
    res = concretize_sorts(term,sort)
    check_concretely_sorted(res,no_error)
    return res
```

`check_concretely_sorted` (lines 1194-1200) RAISES `IvyError("cannot infer sort of {} in {}".format(x,repr(term)))` if any used variable or constant in `res` still contains a `TopSort` (or polymorphic) sort.

Go's `compiler.SortInfer` (`/Users/jaten/ivy/goivy/compiler/compiler.go:1176-1182`) explicitly cites this Python contract in its docstring but **omits the `check_concretely_sorted` call**:

```go
// SortInfer resolves TopSort variables in a compiled logic node.
// Matches Python ivy_logic.py sort_infer:
//
//   res = concretize_sorts(term, sort)
//   check_concretely_sorted(res)
func (c *Compiler) SortInfer(node lg.Expr) (lg.Expr, error) {
    res, err := typeinfer.ConcretizeSorts(node, nil)
    if err != nil {
        return nil, err
    }
    return res, nil   // <-- missing the check_concretely_sorted step
}
```

The actual `CheckConcretelySorted` function exists in the sibling package `ivylogic/sortinfer.go:76-103` and exactly matches Python's behavior. The package `ivylogic` (`il`) is already imported by `compiler/compiler.go`. So the fix is one extra call.

### Why this is the right next step

This change is small, surgical, and principled (it makes Go match Python's documented contract). It will also be a *very useful diagnostic*:

- If `check_concretely_sorted` fires during compilation of `invar386`, we get the exact AST node and a meaningful error message ("cannot infer sort of P in <tactic invariant>") — that pinpoints the bug at the layer where Python catches it. We can then dig into why the typeinfer fix didn't propagate to that particular Variable.
- If compilation succeeds clean and the panic still fires later, then the TopSort `P` must have been introduced by a *post-compilation* transform (most likely in `check/l2s.go` or `check/l2s_shared.go` — `NormalizeFreeVariables` / `ReplaceTemporalsByNamedBinder` / `Desugar` / one of the `modPass` transforms). At that point we have a much smaller search space and a clear next step.

Either way, adding the check makes downstream debugging tractable AND brings Go in line with Python — a strict win over the current silent pass-through.

## Fix

### Edit 1: `/Users/jaten/ivy/goivy/compiler/compiler.go` (around lines 1176-1182)

Add the `CheckConcretelySorted` call that the docstring already promises.

Before:
```go
// SortInfer resolves TopSort variables in a compiled logic node.
// Matches Python ivy_logic.py sort_infer:
//
//   res = concretize_sorts(term, sort)
//   check_concretely_sorted(res)
func (c *Compiler) SortInfer(node lg.Expr) (lg.Expr, error) {
    res, err := typeinfer.ConcretizeSorts(node, nil)
    if err != nil {
        return nil, err
    }
    return res, nil
}
```

After:
```go
// SortInfer resolves TopSort variables in a compiled logic node.
// Matches Python ivy_logic.py sort_infer:
//
//   res = concretize_sorts(term, sort)
//   check_concretely_sorted(res)
//
// The check_concretely_sorted step raises if any used variable or constant
// in the result still has TopSort or polymorphic sort. Mirroring this surfaces
// type-inference bugs at compile time (with a meaningful error) instead of
// letting silent TopSorts propagate to z3bridge.TranslateSort, which can only
// panic.
func (c *Compiler) SortInfer(node lg.Expr) (lg.Expr, error) {
    res, err := typeinfer.ConcretizeSorts(node, nil)
    if err != nil {
        return nil, err
    }
    if err := il.CheckConcretelySorted(res, nil); err != nil {
        return nil, err
    }
    return res, nil
}
```

The `il` alias for `ivylogic` is already imported at the top of `compiler/compiler.go`.

## Verification

**Do NOT run the full `goivy_check ../ivy-lang-examples/examples/liveness/tlb.ivy` from this agent — it takes ~75 seconds and the user runs it on another machine.**

Local checks this agent will run:

1. `go build ./compiler/... ./check/...` — confirm the change compiles.
2. `go test ./compiler/... ./typeinfer/... ./ivylogic/...` — confirm existing tests still pass. If any test newly fails because it was depending on silent-TopSort pass-through, that test was relying on the bug; investigate each one.
3. `go test ./check/...` — same.

After the user re-runs `goivy_check` on the failing tlb.ivy, three outcomes are possible:

A. **Compilation now errors with `cannot infer sort of P in …`** at the line that introduces P. The error site identifies the actual bug to fix in a follow-up plan. Most likely culprits to investigate:
   - the recent `typeinfer/infer.go:507-545` NamedBinder fix may have a subtle gap for the specific shape `App(NamedBinder, Variable)` when `Variable.VSort == TopSort`,
   - or `compile_with_goal_vocab` / `CompileExprVocabExtLF` is failing to reach the inference path for some sub-formula.

B. **Compilation succeeds and the runtime TopSort panic on `invar386` is gone.** Then the previous typeinfer fix WAS sufficient and the user's earlier rebuild simply hadn't picked it up. Confirm by re-running once more.

C. **Compilation succeeds but the same runtime panic still fires.** Then the TopSort `P` is being introduced by a post-compilation transform — almost certainly one of the L2S `modPass` transforms in `check/l2s.go:548-596`, `Desugar` in `check/l2s.go:1064`, or `NormalizeFreeVariables` in `logicutil/logic_utils.go:815`. Write a follow-up plan to investigate that specific layer.

## Critical files

- `/Users/jaten/ivy/goivy/compiler/compiler.go` (lines 1176-1182: `SortInfer`) — the only edit
- `/Users/jaten/ivy/goivy/ivylogic/sortinfer.go` (lines 76-103: `CheckConcretelySorted`) — read-only reference; already correct
- `/Users/jaten/ivy/goivy/typeinfer/infer.go` (lines 507-545: NamedBinder arm) — read-only reference; previous fix in place
- `/Users/jaten/ivy/goivy/check/l2s.go` (lines 444-455: tactic invariant compilation; lines 548-596: modPass transforms) — read-only reference; downstream candidate for outcome (C)

## Reference (source of truth)

- `~/ivy/pyivy/ivy/ivy/ivy_logic.py:1194-1206` — Python `sort_infer` and `check_concretely_sorted`
- `~/ivy/pyivy/ivy/ivy/ivy_compiler.py:518-525` — `_labeled_formula_cmpl` calls `sortify_with_inference` for the formula

## What we are NOT changing in this plan

- `typeinfer/infer.go` — the previous NamedBinder fix is still believed correct; we are not touching it without evidence that it is wrong.
- `check/l2s.go` and the L2S transform pipeline — defer to a follow-up plan triggered by outcome (C).
- The witness/skolem code or z3bridge — unchanged; the panic at `z3bridge/translate.go:303` is the right behavior (it surfaces the bug).
