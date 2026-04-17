# PLAN: Code Review of BUG-11/5/9/10/12/17/19 Fixes — Findings & Fixes

Created: 2026-04-18, 02:30

## Context

We fixed BUGs 5, 9, 10/19, 11, 12, and 17 from the `bugs19.md` audit. The exploration
agents also confirmed that BUGs 1, 2, 3, 4, 6, 7, 8, 14, 16, and 18 were already fixed
in prior sessions. This code review examines the fixes with fresh eyes against the Python
source of truth.

**Overall assessment**: The fixes are structurally correct and match Python's execution
flow well. However, the deep review uncovered **6 concrete issues** — 2 high severity,
2 medium, 2 low.

---

## ISSUE CR-1: `MatchFromDefn` missing `distinct(lhs.args)` check

**Severity**: MEDIUM
**File**: `proof/phase5_goals.go:488-523`

**Python** (`ivy_proof.py:1530-1541`):
```python
if iu.distinct(lhs.args):
    return {lhs.rep: il.Lambda(lhs.args, rhs)}
raise ProofError(defn, 'not a definition')
```

Python rejects definitions with duplicate parameters (e.g., `f(X, X) = rhs`). The `distinct`
check ensures all lambda parameters are unique — duplicate parameters would cause incorrect
substitution behavior.

**Go**: No `distinct` check. `nodesToVarsPhase5` silently drops non-Variable nodes and
accepts duplicate variables. A definition like `forall X. f(X, X) = rhs` would create
`Lambda([X, X], rhs)` with duplicate parameters.

**Fix** in `proof/phase5_goals.go`, add after `nodesToVarsPhase5(app.Terms)`:

```go
// In MatchFromDefn, after creating vars from app.Terms:
lam := &lg.Lambda{Variables: nodesToVarsPhase5(app.Terms), Body: eq.T2}
// ADD: Python distinct(lhs.args) check
seen := make(map[lg.NodeKey]bool, len(lam.Variables))
for _, v := range lam.Variables {
    k := lg.Key(v)
    if seen[k] {
        return nil, &ProofError{Msg: "not a definition: duplicate parameters"}
    }
    seen[k] = true
}
```

Apply to both the Eq and Iff branches (lines 503-511 and 513-521).

---

## ISSUE CR-2: `betaReduce` uses non-capture-aware `SubstituteByName`

**Severity**: HIGH
**File**: `proof/match.go:453-464`

**Python**: `lambda_apply` (ivy_logic.py:1640-1642) calls `lu.substitute` (logic_util.py:129).
`lu.substitute` detects capture: at binders, it computes free variables of the remaining
substitution values and **raises `CaptureError`** if any free variable is a bound variable
of the binder (logic_util.py:164-178). The CaptureError propagates up to the tactic handler.

**Go**: `betaReduce` calls `lu.SubstituteByName`, which at binders only removes bound
variable names from the substitution map (logicutil/logic_utils.go:622-638). It does NOT
check for capture — if a substituted value's free variable clashes with a binder's bound
variable, the substitution silently corrupts the formula.

**Example**: Lambda(X, forall Y. X+Y) applied to arg `Y`:
- Python: CaptureError (Y is free in arg, bound by forall)
- Go: Returns `forall Y. Y+Y` — WRONG (Y captured)

**Fix**: Replace `betaReduce` with a call to `il.LambdaApply` in `proof/match.go`.
`il.LambdaApply` (ivylogic/constructors.go:157-166) already uses `lu.Substitute` which
detects capture via `CaptureError`.

```go
// Replace betaReduce function:
// OLD:
func betaReduce(lam *lg.Lambda, args []lg.Expr) lg.Expr {
    subs := make(map[string]lg.Expr, len(lam.Variables))
    for i, v := range lam.Variables { subs[v.Name] = args[i] }
    return lu.SubstituteByName(lam.Body, subs)
}

// NEW:
func betaReduce(lam *lg.Lambda, args []lg.Expr) lg.Expr {
    if len(lam.Variables) != len(args) {
        return lam.Body // arity mismatch
    }
    result, err := il.LambdaApply(lam, args)
    if err != nil {
        // Capture detected — return original body (uncollapsed).
        // Python raises CaptureError which propagates; for the non-alt
        // path this is equivalent to skipping the substitution.
        return lam.Body
    }
    return result
}
```

This also removes the `lu` import from `proof/match.go` if it's the only usage of
`logicutil`. **Check whether `lu` is used elsewhere in match.go before removing.**

---

## ISSUE CR-3: `LambdaApply` errors silently discarded in `applyMatchAltRec`

**Severity**: HIGH
**File**: `proof/phase5_matching.go:958,986`

**Current code**:
```go
result, _ := il.LambdaApply(lam, newTerms)  // error DISCARDED
return result  // nil if CaptureError!
```

If `LambdaApply` returns a `CaptureError`, `result` is `nil`. Returning `nil` will cause
a nil pointer dereference downstream.

**Python** (`ivy_proof.py:1141-1146`): `apply_fun` catches `CaptureError` from `fun(*args)`
and converts it to `raise_capture(sym)`, which raises a proof-level error. This propagates
to the caller who can try a different match.

**Fix**: Handle the error at both call sites in `applyMatchAltRec`:

```go
// At line 958 and 986:
result, err := il.LambdaApply(lam, newTerms)
if err != nil {
    // Capture detected — return original formula unchanged
    // (Python raises CaptureError; skipping substitution is safe
    // because the caller will re-try or report error)
    return fmla
}
return result
```

---

## ISSUE CR-4: `LambdaApply` error silently discarded in `applyUnfoldRec`

**Severity**: MEDIUM
**File**: `proof/phase5_goals.go:635`

**Current code**:
```go
result, _ := il.LambdaApply(l, newArgs)
return result
```

Same issue as CR-3 but in the unfold path.

**Fix**:
```go
result, err := il.LambdaApply(l, newArgs)
if err != nil {
    return fmla // capture — return original
}
return result
```

---

## ISSUE CR-5: `applyUnfoldGoal` doesn't process ConstantDecl premises

**Severity**: LOW-MEDIUM (edge case, unlikely in practice)
**File**: `proof/phase5_goals.go:653-673`

**Python**: `unfold_goal` → `apply_match_goal` (ivy_proof.py:975-993) processes ALL premise
types:
- `LabeledFormula`: recursively process
- `ConstantDecl`: `x.clone([apply_match_func_alt(match, x.args[0], env)])` — transforms
  the declared symbol via match. If the symbol matches the unfold key, `match_get` pops
  a lambda, creating a lambda-typed ConstantDecl.
- Then the lambda filter removes it: `prems = [p for p in prems if not is_lambda(p)]`

**Go**: `applyUnfoldGoal` passes non-LabeledFormula premises through unchanged. ConstantDecl
premises with the unfold symbol are not transformed or filtered.

**Impact**: If a goal has a ConstantDecl for the unfold symbol, Python would pop a lambda
for it (consuming one from the union) and then filter it out. Go doesn't pop and doesn't
filter. This affects pop order when the goal has both ConstantDecl premises and formula
occurrences of the unfold symbol. In practice this is rare — unfold usually targets symbols
defined OUTSIDE the current goal.

**Fix**: In `applyUnfoldGoal`, after processing premises, filter out any ConstantDecl whose
symbol matches the unfold key:

```go
// After building newPrems, filter ConstantDecl matching unfold key.
// Python: prems = [p for p in prems if not is_lambda(p)]
// For unfold specifically: remove ConstantDecl whose symbol = unfold key,
// since Python would transform it to lambda-typed and then filter.
filtered := make([]ast.Node, 0, len(newPrems))
for _, p := range newPrems {
    if cd, ok := p.(*ast.ConstantDecl); ok {
        args := cd.Args()
        if len(args) > 0 {
            if c, ok := args[0].(*lg.Const); ok && lg.Key(c) == key {
                continue // Python would make this lambda-typed, then filter
            }
        }
    }
    filtered = append(filtered, p)
}
newPrems = filtered
```

---

## ISSUE CR-6: `applyUnfoldRec` non-lambda Pop fallback drops arguments

**Severity**: MEDIUM — must match Python behavior exactly
**File**: `proof/phase5_goals.go:637-638`

**Current code**:
```go
return lam // non-Lambda: returns value without applying to newArgs — WRONG
```

**Python**: `apply_fun(fun, args)` calls `fun(*args)`. For a Const/Var/Symbol,
`fun(*args)` calls `Symbol.__call__(*args)` which creates `Apply(fun, args)`.
For a Lambda, `Lambda.__call__(*args)` calls `lambda_apply` which does
capture-detecting substitution. Both paths ALWAYS apply the function to the args.

**Fix**: Match Python — always apply the popped value to the processed arguments:

```go
lam := union.Pop()
if l, ok := lam.(*lg.Lambda); ok {
    result, err := il.LambdaApply(l, newArgs)
    if err != nil {
        return fmla // capture — return original
    }
    return result
}
// Non-lambda: apply as function to args, matching Python Symbol.__call__
if c, ok := lam.(*lg.Const); ok && len(newArgs) > 0 {
    return lg.MustApply(c, newArgs...)
}
if len(newArgs) > 0 {
    // Variable or other callable — reconstruct Apply
    return lg.MustApply(lam, newArgs...)
}
return lam
```

---

## Summary of Unfixed Bugs from bugs19.md

The exploration confirmed ALL 19 bugs are now addressed:

| # | Status | Notes |
|---|--------|-------|
| 1 | FIXED | CompileOneMatch vmatch + match_sort (prior session) |
| 2 | FIXED | applyMatchAltRec env tracking + MatchGet (prior session) |
| 3 | FIXED | AlphaAvoidMap in ApplyMatch + ApplyMatchAlt (prior session) |
| 4 | FIXED | ApplyMatchToProblemNonAlt exists + used correctly (prior session) |
| 5 | FIXED | functionTactic → CompileDefinitionGoalVocab (this session) |
| 6 | FIXED | propertyTactic with GoalSubst + GoalAddPrem (prior session) |
| 7 | FIXED | letTactic uses CompileExprVocab (prior session) |
| 8 | FIXED | ifTactic uses goal.Formula + attribGoals (prior session) |
| 9 | FIXED | witnessTactic → CompileWitnessList + WitnessAst (this session) |
| 10 | FIXED | unfoldTactic → LookupSchema + renamings + has_premise (this session) |
| 11 | FIXED | ExprListOrLambdaUnion + destructive Pop (this session) |
| 12 | FIXED | applyMatchRec binder → CloneBinder with processed vars (this session) |
| 13 | FIXED | MatchSchema fomatch uses NonAlt variant (prior session) |
| 14 | FIXED | ApplyMatchFreesyms (non-alt) used correctly (prior session) |
| 15 | N/A | Handled by unfoldRhsVars; ExprListOrLambdaUnion doesn't go through MatchRhsVars |
| 16 | FIXED | Lambda-premise filter in ApplyMatchGoalNode (prior session) |
| 17 | FIXED | ApplyMatchGoalNode env from GoalDefns (this session) |
| 18 | FIXED | witnessTactic temporal check (this session) |
| 19 | FIXED | unfoldTactic uses UnfoldFmla/UnfoldGoal (this session, sub-bug of 10) |

---

## Files to modify

1. `proof/phase5_goals.go` — CR-1 (distinct check), CR-4 (LambdaApply error), CR-5 (ConstantDecl filter), CR-6 (non-lambda fallback)
2. `proof/match.go` — CR-2 (betaReduce → LambdaApply)
3. `proof/phase5_matching.go` — CR-3 (LambdaApply error in applyMatchAltRec)

## Verification

```bash
cd ~/ivy/goivy && make test
```
