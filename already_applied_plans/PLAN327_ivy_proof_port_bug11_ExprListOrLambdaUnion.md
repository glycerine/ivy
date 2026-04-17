# PLAN: BUG-11 Full Fix — List-Valued Match for Unfold with ExprListOrLambdaUnion

Created: 2026-04-18, 00:45

## Context

Python's `match_from_defns` (ivy_proof.py:1543-1547) creates `{sym: [lam1, lam2, ...]}`.
Python's `match_get` (ivy_proof.py:1108-1113) does a destructive pop — each call returns
the next lambda. This gives single-pass unfolding where each occurrence of `sym` gets a
different renamed lambda.

Go's `MatchFromDefns` currently uses only the first lambda. This is a bug — Go must unfold
all K occurrences in one pass, exactly like Python.

## Design

`ExprListOrLambdaUnion` does NOT implement `lg.Expr` — it stays localized to the unfold
code path. Since `MatchFromDefns` returns a single-key match (one defined symbol), the
unfold path doesn't need the full generality of `applyMatchAltRec`. A focused ~40-line
recursive function handles: recurse children → check if app matches the unfold key → pop
lambda → beta-reduce. Alpha-avoid prevents capture.

## Files to modify

1. `proof/phase5_goals.go` — `ExprListOrLambdaUnion`, rewrite `MatchFromDefns`, `UnfoldFmla`, `UnfoldGoal`

## Steps

### Step 1: Define `ExprListOrLambdaUnion` in `proof/phase5_goals.go`

```go
// ExprListOrLambdaUnion holds either a single lambda or a list of lambdas
// for multi-definition unfolding. Used by MatchFromDefns.
// Does NOT implement lg.Expr — stays localized to the unfold path.
//
// Python: match_from_defns returns {sym: [lambda1, lambda2, ...]}
// Python: match_get pops the first element on each access:
//   val = save[0]; if len(save) > 1: del save[0]
type ExprListOrLambdaUnion struct {
    Items []lg.Expr // always len >= 1; Pop returns next
}

// Pop returns the first item. If more than one item, removes it.
// Mirrors Python: val = save[0]; if len(save) > 1: del save[0]
func (u *ExprListOrLambdaUnion) Pop() lg.Expr {
    if len(u.Items) == 0 {
        return nil
    }
    val := u.Items[0]
    if len(u.Items) > 1 {
        u.Items = u.Items[1:]
    }
    return val
}
```

### Step 2: Rewrite `MatchFromDefns`

Return `(lg.NodeKey, *ExprListOrLambdaUnion, error)` — just the key and the union.

Python (ivy_proof.py:1543-1547):
```python
def match_from_defns(defns):
    matches = [match_from_defn(d) for d in defns]
    lhs = list(matches[0].keys())[0]
    assert all(lhs in m for m in matches)
    return {lhs: [m[lhs] for m in matches]}
```

Go:
```go
func MatchFromDefns(defns []*ast.LabeledFormula) (lg.NodeKey, *ExprListOrLambdaUnion, error) {
    if len(defns) == 0 {
        return "", nil, &ProofError{Msg: "no definitions"}
    }
    matches := make([]map[lg.NodeKey]lg.Expr, len(defns))
    for i, d := range defns {
        m, err := MatchFromDefn(d)
        if err != nil {
            return "", nil, err
        }
        matches[i] = m
    }
    // Get lhs key from first match
    var lhsKey lg.NodeKey
    for k := range matches[0] {
        lhsKey = k
        break
    }
    // Assert all share same key (Python: assert all(lhs in m for m in matches))
    for _, m := range matches[1:] {
        if _, ok := m[lhsKey]; !ok {
            return "", nil, &ProofError{Msg: "match_from_defns: definitions have different LHS symbols"}
        }
    }
    // Collect all lambdas into union
    items := make([]lg.Expr, len(matches))
    for i, m := range matches {
        items[i] = m[lhsKey]
    }
    return lhsKey, &ExprListOrLambdaUnion{Items: items}, nil
}
```

### Step 3: Rewrite `UnfoldFmla` — localized unfold with alpha-avoid + pop

Python (ivy_proof.py:1555-1559):
```python
def unfold_fmla(fmla, defns):
    for rdefs in defns:
        match = match_from_defns(rdefs)
        fmla = apply_match_alt(match, fmla)
    return fmla
```

Go — for each rdefs group:
1. Get the unfold key + union from `MatchFromDefns`
2. Compute free vars from ALL lambdas in the union → alpha-avoid
3. Call localized `applyUnfoldRec` which pops from the union on each match

```go
func UnfoldFmla(fmla lg.Expr, defns [][]*ast.LabeledFormula) lg.Expr {
    for _, rdefs := range defns {
        key, union, err := MatchFromDefns(rdefs)
        if err != nil {
            continue
        }
        // Alpha-avoid: rename bound vars that clash with free vars in ALL lambdas.
        // Python: apply_match_alt calls match_rhs_vars then alpha_avoid.
        freeVars := unfoldRhsVars(union)
        fmla = il.AlphaAvoidMap(fmla, freeVars)
        // Single-pass unfold with destructive pop
        fmla = applyUnfoldRec(key, union, fmla)
    }
    return fmla
}
```

### Step 4: Rewrite `UnfoldGoal` — same approach via `ApplyMatchGoalNode`-equivalent

Python (ivy_proof.py:1549-1553):
```python
def unfold_goal(goal, defns):
    for rdefs in defns:
        match = match_from_defns(rdefs)
        goal = apply_match_goal(match, goal, apply_match_alt)
    return goal
```

For `UnfoldGoal`, we need to process premises too (via `apply_match_goal`). Build a
single-entry `map[lg.NodeKey]lg.Expr` with the pop result for each call, then delegate
to `ApplyMatchGoalNode` for the goal structure traversal. The pop must happen during the
conclusion's recursive traversal.

```go
func UnfoldGoal(cfg *ast.AstConfig, goal *ast.LabeledFormula, defns [][]*ast.LabeledFormula) *ast.LabeledFormula {
    for _, rdefs := range defns {
        key, union, err := MatchFromDefns(rdefs)
        if err != nil {
            continue
        }
        // Build single-entry match for premise processing (ApplyMatchGoalNode)
        // BUT use the union for the conclusion's recursive unfold
        freeVars := unfoldRhsVars(union)
        goal = applyUnfoldGoal(cfg, key, union, freeVars, goal)
    }
    return goal
}
```

### Step 5: Helper functions (all in `proof/phase5_goals.go`)

```go
// unfoldRhsVars computes free vars from all lambdas in a union.
// Python: match_rhs_vars handles list values via:
//   for v in w if isinstance(w, list) else [w]: ...
func unfoldRhsVars(union *ExprListOrLambdaUnion) map[lg.NodeKey]lg.Expr {
    result := make(map[lg.NodeKey]lg.Expr)
    for _, item := range union.Items {
        for k, sym := range FmlaVocab(item) {
            result[k] = sym
        }
    }
    return result
}

// applyUnfoldRec recursively unfolds a formula with destructive pop.
// Simplified version of applyMatchAltRec for the single-key unfold case.
func applyUnfoldRec(key lg.NodeKey, union *ExprListOrLambdaUnion, fmla lg.Expr) lg.Expr {
    if fmla == nil {
        return nil
    }
    args := il.NodeArgs(fmla)
    newArgs := make([]lg.Expr, len(args))
    for i, a := range args {
        newArgs[i] = applyUnfoldRec(key, union, a)
    }
    // App: if function matches the unfold key, pop and beta-reduce
    if il.IsApp(fmla) {
        if c := appFuncConst(fmla); c != nil && lg.Key(c) == key {
            lam := union.Pop()
            if l, ok := lam.(*lg.Lambda); ok {
                result, _ := il.LambdaApply(l, newArgs)
                return result
            }
            return lam
        }
    }
    // Binder: clone with processed vars and body
    if il.IsQuantifier(fmla) && len(newArgs) > 0 {
        vars := il.BinderVars(fmla)
        return il.CloneBinder(fmla, vars, newArgs[0])
    }
    if len(args) == 0 {
        return fmla
    }
    return il.CloneNode(fmla, newArgs)
}

// applyUnfoldGoal applies unfold to a goal (premises + conclusion).
// Premises: build single-entry match from popped lambda, apply via ApplyMatchGoalNode.
// Conclusion: use applyUnfoldRec for the destructive-pop traversal.
func applyUnfoldGoal(cfg *ast.AstConfig, key lg.NodeKey, union *ExprListOrLambdaUnion, freeVars map[lg.NodeKey]lg.Expr, goal *ast.LabeledFormula) *ast.LabeledFormula {
    // Process premises using single-entry match (each premise unfold pops independently)
    prems := GoalPrems(goal)
    var newPrems []ast.Node
    for _, p := range prems {
        if lf, ok := p.(*ast.LabeledFormula); ok {
            newPrems = append(newPrems, applyUnfoldGoal(cfg, key, union, freeVars, lf))
        } else {
            newPrems = append(newPrems, p)
        }
    }
    // Unfold the conclusion with alpha-avoid + destructive pop
    newConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
        c = il.AlphaAvoidMap(c, freeVars)
        return applyUnfoldRec(key, union, c)
    })
    return CloneGoal(cfg, goal, newPrems, newConc)
}

// appFuncConst extracts the Const function symbol from an Apply node.
func appFuncConst(fmla lg.Expr) *lg.Const {
    if app, ok := fmla.(*lg.Apply); ok {
        if c, ok := app.Func.(*lg.Const); ok {
            return c
        }
    }
    return nil
}
```

## Verification

```bash
cd ~/ivy/goivy && make test
```
