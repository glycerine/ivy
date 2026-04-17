# PLAN: Systematic Audit of ivy_proof.py Go Port — Divergence List

Created: 2026-04-17, 23:45

## Context

A systematic, line-by-line comparison of Python `~/ivy/pyivy/ivy/ivy/ivy_proof.py`
against the Go port in `~/ivy/goivy/proof/` to find every place where the Go port is
incomplete, incorrect, or missing functionality. Prior "simplifications" introduced
multiple silent bugs. This document is the audit ledger.

Python source: `~/ivy/pyivy/ivy/ivy/ivy_proof.py`
Go files audited: `proof/tactics.go`, `proof/checker.go`, `proof/matching.go`,
`proof/match.go`, `proof/phase5_matching.go`, `proof/phase5_goals.go`

---

## BUG LIST (ordered by severity)

---

### BUG-1: `CompileOneMatch` is a stub — missing sort-variable matching and UninterpretedSort case
**File**: `proof/phase5_matching.go:684-689`
**Severity**: HIGH — affects correctness of all schema matching with sort variables

Python `compile_one_match` (ivy_proof.py:897-916) does three branches:
1. Variable LHS → `fo_match` ✓ (Go has this)
2. Non-UninterpretedSort RHS → computes sort-variable matches (`vmatch`) for
   variables shared by lhs and rhs whose sort is free, then applies sort match
   to lhs, runs `match(lhs, rhs, newfreesyms, constants)`, composes result.
3. UninterpretedSort RHS → `match_sort(lhs, rhs, freesyms)`

Go just does:
```go
if _, isVar := lhs.(*lg.Variable); isVar {
    return FOMatch(lhs, rhs, freesyms, constants)
}
return Match(lhs, rhs, freesyms, constants)  // WRONG: skips vmatch and match_sort
```

Missing:
- `vmatch` sort-variable matching loop
- Application of vmatch to lhs before matching
- `compose_matches(freesyms, vmatch, somatch, vmatch)` composition
- `match_sort` for `UninterpretedSort` rhs case

**Fix location**: `proof/phase5_matching.go:684-689` — rewrite `CompileOneMatch`.

---

### BUG-2: `applyMatchAltRec` does not use `env` for capture detection
**File**: `proof/phase5_matching.go:874-973`
**Severity**: HIGH — means variable capture is silently undetected

Python `apply_match_alt_rec` (ivy_proof.py:1140-1158) uses `match_get(match, sym, env)` for ALL
symbol lookups. `match_get` raises `CaptureError` if any symbol in the matched value
appears in `env` (the set of currently-bound symbols in the formula).

Go's `applyMatchAltRec` ignores `env` entirely:
- All match lookups use `match[k]` directly (no capture check)
- Binder cases (`ForAll`, `Exists`, `Lambda`) do NOT add their variables to `env`
  before recursing into the body (Python does `with il.BindSymbols(env, fmla.variables)`)

**Fix**: In `applyMatchAltRec`, for every match lookup call `MatchGet(match, sym, env, default)`
and check for error. For binder cases, add bound variables to `env` before recursing
into the body, and restore them after.

---

### BUG-3: `ApplyMatchAlt` and `ApplyMatch` missing `AlphaAvoid` pre-pass
**File**: `proof/phase5_matching.go:862-869` and `proof/match.go:361-366`
**Severity**: HIGH — wrong results when match RHS vars clash with bound vars in formula

Python both `apply_match` (ivy_proof.py:1067-1078) and `apply_match_alt` (ivy_proof.py:1117-1130) do:
```python
freevars = match_rhs_vars(match)  # or list(match_rhs_vars(match))
fmla = il.alpha_avoid(fmla, freevars)
```
before calling the recursive function. This renames bound variables in `fmla` that
would collide with free variables introduced by the substitution.

Go's `ApplyMatchAlt` and `ApplyMatch` go directly to the recursive function with no
alpha-renaming. This means substitutions can silently capture bound variables.

**Fix**: In `ApplyMatch` and `ApplyMatchAlt`, compute `MatchRhsVars(match)`, call
`il.AlphaAvoid(fmla, freeVars)` (if it exists in Go) to rename conflicting bound vars,
then proceed with the recursive call.

---

### BUG-4: `ApplyMatchToProblem` always uses `apply_match_alt` — Python uses different functions per call
**File**: `proof/matching.go:156-181`
**Severity**: MEDIUM-HIGH — fomatch applications use wrong variant

Python `apply_match_to_problem` takes the apply function as a parameter. Different callers
pass different functions:
- Initial `pmatch` in `match_schema`: `apply_match_to_problem(pmatch, prob, apply_match_alt)` ✓
- `fomatch` in `match_schema` (non-tuple): `apply_match_to_problem(fomatch, prob, apply_match)` ← NON-ALT
- `somatch` in `match_schema`: `apply_match_to_problem(somatch, prob, apply_match_alt)` ✓

Go's `ApplyMatchToProblem` always uses `ApplyMatchAlt` internally. For the `fomatch` case,
Python uses `apply_match` (without capture checking).

Similarly, `apply_match_to_problem` always uses `apply_match_freesyms` (non-alt) for
`prob.freesyms`. Go uses `ApplyMatchFreesymsAlt`.

**Fix**: Either add an `applyFn` parameter to `ApplyMatchToProblem` (or create two variants),
and call with the appropriate function from `MatchSchema`. Also fix freesyms update to use
non-alt version.

---

### BUG-5: `functionTactic` is completely wrong — adds Implies instead of premises
**File**: `proof/tactics.go:396-431`
**Severity**: HIGH — semantically incorrect for `function` tactic in proofs

Python `function_tactic` (ivy_proof.py:275-304):
1. Gets `goal_vocab(goal)` and `goal_free(goal)`
2. For each `df` in proof.args:
   - Extracts the definition formula, creates a `TopFunctionSort` symbol
   - With that symbol in scope, builds `Forall(vars, = lhs rhs)` formula
   - Compiles via `compile_expr_vocab(elf, vocab)` — symbol resolution in context
   - Extracts the compiled sym, checks for recursion
   - **Adds `ConstantDecl(sym)` as a PREMISE to goal**: `goal = goal_add_prem(goal, cd, ...)`
   - **Adds the definition LF as a PREMISE**: `goal = goal_add_prem(goal, lf, ...)`
   - Checks for redefinition in vocab/free
3. Returns `[goal] + decls[1:]` — goal with new premises

Go instead:
- Extracts formula from proof.Elems (wrong field)
- Creates `Implies(defFormula, G)` — WRONG: Python adds as PREMISE, not implication
- No compilation, no recursion check, no redefinition check, no ConstantDecl premise

**Fix**: Rewrite `functionTactic` to follow Python faithfully: compile definition,
add `ConstantDecl` + definition LF as premises, check for redefinition.

---

### BUG-6: `propertyTactic` has wrong return structure — adds Implies instead of premise
**File**: `proof/tactics.go:345-391`
**Severity**: HIGH — semantically incorrect for `property` tactic in proofs

Python `property_tactic` (ivy_proof.py:225-273) returns:
```python
[goal_add_prem(goal, cut, cut.lineno)] + decls[1:] + subgoals
```
Where:
- `goal_add_prem(goal, cut, cut.lineno)` = original goal WITH `cut` as a LABELED FORMULA PREMISE
- `subgoals` = goals to PROVE the cut (via `goal_subst(goal, cut, cut.lineno)` which
  merges premises from both goals)

Additional Python steps missing from Go:
- `cut = compile_expr_vocab(proof.args[0], vocab)` — compiles cut with goal vocab
- `cut = normalize_goal(cut)` — normalizes
- `subgoal = goal_subst(goal, cut, cut.lineno)` — subgoal inherits goal's premises
- Proof args[1] (optional Skolem function introduction): creates fresh function symbol,
  introduces ConstantDecl premise, substitutes existential witness
- Proof args[2] (optional proof for the cut): applies proof to subgoals

Go instead creates `Implies(cut, G)` as the modified goal and `CloneGoal(goal, prems, cutFormula)`
as the proof obligation — wrong structure, missing compilation and normalization.

**Fix**: Rewrite `propertyTactic` to follow Python faithfully.

---

### BUG-7: `letTactic` missing `compile_expr_vocab` for equality terms
**File**: `proof/tactics.go:18-70`
**Severity**: MEDIUM-HIGH — equality atoms not compiled with goal vocabulary

Python `let_tactic` (ivy_proof.py:213-223):
```python
vocab = goal_vocab(goal)
defs = [compile_expr_vocab(ia.Atom('=', x.args[0], x.args[1]), vocab) for x in proof.args]
cond = il.And(*[il.Equals(a.args[0], a.args[1]) for a in defs])
```

Go just does `astNodeToLogicNode(defArgs[0])` — skips compilation. Symbol names in the
equality atoms are NOT resolved against the goal's vocabulary.

Also structural difference: Python creates `LabeledFormula(goal.label, il.Implies(cond, goal.formula))`
using `goal.formula` directly (not `goal_conc`). This wraps the ENTIRE formula (including
any SchemaBody). Go uses `CloneGoal` with `GoalConc` which only wraps the conclusion.
If the goal has premises in a SchemaBody, these produce structurally different results.

**Fix**: Add `CompileExprVocab` call for each equality atom before building `cond`.
Also match Python's use of `goal.formula` directly in the LabeledFormula construction.

---

### BUG-8: `ifTactic` uses wrong structure — same issue as `letTactic`
**File**: `proof/tactics.go:281-343`
**Severity**: MEDIUM — structural difference from Python

Python `if_tactic` (ivy_proof.py:402-410):
```python
true_goal = ia.LabeledFormula(decls[0].label, il.Implies(cond, decls[0].formula))
```
Uses `decls[0].formula` (the full formula including SchemaBody).

Go uses `CloneGoal(goal, GoalPrems(goal), ApplyToConc(GoalConc(goal), ...))`.

Also Python uses `attrib_goals(proof.args[1], ...)` and `attrib_goals(proof.args[2], ...)`
on the results of applying branch proofs — Go doesn't call `AttribGoals`.

**Fix**: Match Python's direct wrapping of `goal.formula` in Implies.

---

### BUG-9: `witnessTactic` missing `compile_witness_list` and variable validation
**File**: `proof/tactics.go:437-482`
**Severity**: MEDIUM — witnesses not compiled with goal vocabulary

Python `witness_tactic` (ivy_proof.py:451-463):
```python
wits = compile_witness_list(proof, decls[0])
for wit in wits:
    if not il.is_variable(wit.args[0]):
        raise iu.IvyError(wit, 'left-hand side of witness must be a variable')
wit_map = dict((x.args[0], x.args[1]) for x in wits)
conc = lu.witness_ast(False, [], wit_map, conc)
```

Also checks: `if ia.has_temporal(proof) and not goal_is_temporal(goal): raise error`

Go:
- Doesn't call `CompileWitnessList` (uses raw `astNodeToLogicNode` on witness args)
- Doesn't validate that LHS is a variable
- Doesn't check temporal operator constraint
- Uses custom `applyWitness` instead of `module.WitnessAst(false, nil, witMap, conc)`

**Fix**: Call `CompileWitnessList`, add variable validation, check temporal constraint,
use `module.WitnessAst` instead of the custom `applyWitness`.

---

### BUG-10: `unfoldTactic` missing LookupSchema, renamings, has_premise, and wrong unfold logic
**File**: `proof/tactics.go:216-263`
**Severity**: MEDIUM-HIGH — unfold tactic misses several cases

Python `unfold_tactic` (ivy_proof.py:376-392):
1. `defn = self.lookup_schema(defname, decl, proof)` — looks in **schemata first**, then definitions
2. `rdefs = [rename_goal(defn, rn) for rn in unfspec.renamings]` — handles renamed copies
   `rdefs.append(defn)` — base copy at end
   `defns.append(rdefs)` — grouped list of renames per unfspec
3. If `proof.has_premise`: `goal_apply_to_prem(decl, premname, unfold_goal)` — unfolds IN a premise
4. Else: `goal_apply_to_conc(decl, unfold_fmla)` — unfolds in conclusion

Python `unfold_fmla` uses `match_from_defns` + `apply_match_alt` (proper HO matching).

Go:
- Only checks `pc.Definitions[defName]` — misses schemata
- Doesn't process `unfspec.renamings` — ignores all renamed copies
- Doesn't handle `proof.has_premise` — always unfolds conclusion
- `unfoldFmla` uses `lu.SubstituteByName` — wrong (simple name sub, not HO matching)

**Fix**: Use `pc.LookupSchema`, handle `renamings`, handle `has_premise`,
replace `unfoldFmla` with `UnfoldGoal`/`UnfoldFmla` which use `MatchFromDefns` + `ApplyMatchAlt`.

---

### BUG-11: `MatchFromDefns` only processes first definition
**File**: `proof/phase5_goals.go:538-543`
**Severity**: MEDIUM — multi-definition unfolding silently broken

Python `match_from_defns` (ivy_proof.py:1535-1539):
```python
def match_from_defns(defns):
    matches = [match_from_defn(d) for d in defns]
    lhs = list(matches[0].keys())[0]
    assert all(lhs in m for m in matches)
    return {lhs: [m[lhs] for m in matches]}  # LIST-VALUED match for unfolding
```

Creates a list-valued match `{sym: [lambda1, lambda2, ...]}` for multi-step unfolding.

Go:
```go
func MatchFromDefns(defns []*ast.LabeledFormula) (map[lg.NodeKey]lg.Expr, error) {
    return MatchFromDefn(defns[0])  // Only first!
}
```

**Fix**: Implement full Python version — collect all matches, assert same lhs key,
create list-valued match entry. Also update `MatchRhsVars` and `applyMatchAltRec`
to handle list-valued match entries.

---

### BUG-12: `ApplyMatch` binder handling skips `env` tracking and `clone_binder` call
**File**: `proof/match.go:419-428`
**Severity**: MEDIUM

Python `apply_match_rec` (ivy_proof.py:1080-1093) for binders:
```python
if il.is_binder(fmla):
    with il.BindSymbols(env, fmla.variables):
        fmla = fmla.clone_binder([apply_match_rec(match, v, env) for v in fmla.variables], args[0])
    return fmla
```
Tracks bound variables in `env` and uses `clone_binder`.

Go:
```go
if il.IsQuantifier(fmla) {
    // Already handled by recursion into body (comment — does nothing special!)
}
return il.CloneNode(fmla, newArgs)
```
No env tracking, no special binder handling.

**Fix**: For binder nodes in `applyMatchRec`, add bound variables to env before
recursing into the body, use clone_binder equivalent.

---

### BUG-13: `MatchSchema` applies fomatch with alt variant — Python uses non-alt
**File**: `proof/checker.go:357-369`
**Severity**: MEDIUM

Python `match_schema` non-tuple:
```python
fomatch = fo_match(...)
if fomatch is not None:
    apply_match_to_problem(fomatch, prob, apply_match)  # NON-ALT
```

Go:
```go
fomatch := FOMatch(...)
if fomatch != nil && len(fomatch) > 0 {
    ApplyMatchToProblem(pc.astCfg(), fomatch, prob)  // always uses alt internally
}
```

Same issue for the tuple case.

**Fix**: Create `ApplyMatchToProblemNonAlt` variant or add `applyFn` param, and call
the non-alt version when applying fomatch.

---

### BUG-14: `ApplyMatchToProblem` uses `ApplyMatchFreesymsAlt` — Python uses non-alt `apply_match_freesyms`
**File**: `proof/matching.go:173`
**Severity**: MEDIUM

Python `apply_match_to_problem` always uses `apply_match_freesyms` (non-alt):
```python
prob.freesyms = apply_match_freesyms(match, prob.freesyms)
```

Non-alt: `set(apply_match_sym(match, sym) for sym in freesyms if sym not in match)`
Alt: `[apply_match_sym(match, sym) for sym in freesyms if apply_match_sym(match, sym) not in match]`

The difference matters when a freesym maps to another symbol that is also in match.

**Fix**: Use `ApplyMatchFreesyms` (non-alt) in `ApplyMatchToProblem`.

---

### BUG-15: `MatchRhsVars` doesn't handle list-valued match entries (cascades from BUG-11)
**File**: `proof/phase5_matching.go:747-758`
**Severity**: LOW (cascades from BUG-11; only affects unfolding)

Python `match_rhs_vars` (ivy_proof.py:947-956):
```python
for w in list(match.values()):
    for v in w if isinstance(w, list) else [w]:  # handles list values!
```

Go `MatchRhsVars` just calls `FmlaVocab(v)` for each value — doesn't handle list-valued entries.

**Fix**: After BUG-11 is fixed, update `MatchRhsVars` to handle `[]lg.Expr` list values.

---

### BUG-16: `ApplyMatchGoalNode` missing lambda-premise filtering
**File**: `proof/phase5_matching.go:1189-1240`
**Severity**: LOW-MEDIUM

Python `apply_match_goal` (ivy_proof.py:975-993) for SchemaBody:
```python
prems = [apply_match_goal(match, y, apply_match, env) for y in fmla.prems()]
prems = [p for p in prems if not is_lambda(p)]  # FILTER lambda prems
```

After applying a match, some `ConstantDecl` premises may become lambda-typed. Python filters
these out. Go's `ApplyMatchGoalNode` does not perform this filter.

**Fix**: After building `newPrems`, filter out any `ConstantDecl` whose arg is a Lambda.

---

### BUG-17: `ApplyMatchGoalNode` passes `nil` env — Python tracks bound env
**File**: `proof/phase5_matching.go:1236`
**Severity**: LOW-MEDIUM (cascades from BUG-2)

Python `apply_match_goal` builds `bound = [s for s in goal_defns(x) if s not in match]`
and uses `BindSymbols(env, bound)` before recursing. This threads the environment through
all premise processing.

Go passes `nil` for env to `ApplyMatchAlt`. Once BUG-2 is fixed (env tracking in
`applyMatchAltRec`), this also needs fixing.

**Fix**: Compute goal_defns, add to env before recursing into premises.

---

### BUG-18: `goal_is_temporal` not checked in `witnessTactic`
**File**: `proof/tactics.go:437`
**Severity**: LOW

Python `witness_tactic`:
```python
if ia.has_temporal(proof) and not goal_is_temporal(goal):
    raise iu.IvyError(proof, 'temporal operator not allowed in instantiation')
```

Go doesn't check this. `GoalIsTemporal` exists in `proof/phase5_goals.go:124`.

**Fix**: Add the temporal check at the top of `witnessTactic`.

---

### BUG-19: `unfoldFmla` uses `SubstituteByName` instead of HO match + `ApplyMatchAlt`
**File**: `proof/tactics.go:266-279`
**Severity**: HIGH (sub-bug of BUG-10, listed separately for clarity)

Go's `unfoldFmla` does simple `lu.SubstituteByName` name substitution.

Python `unfold_fmla` (ivy_proof.py:1547-1551):
```python
def unfold_fmla(fmla, defns):
    for rdefs in defns:
        match = match_from_defns(rdefs)
        fmla = apply_match_alt(match, fmla)
    return fmla
```

Uses a proper higher-order match (which may be lambda-valued from `match_from_defns`),
then applies it via `apply_match_alt` which does proper beta reduction.

**Fix**: Replace `unfoldFmla` with proper `UnfoldFmla` from `phase5_goals.go:560-569`
which already exists. Confirm it uses `MatchFromDefns` (once fixed) + `ApplyMatchAlt`.

---

## Summary Table

| # | Function | File | Severity | Type |
|---|----------|------|----------|------|
| 1 | `CompileOneMatch` | phase5_matching.go:684 | HIGH | Missing vmatch + match_sort |
| 2 | `applyMatchAltRec` env | phase5_matching.go:874 | HIGH | env not used for capture |
| 3 | `ApplyMatchAlt/Match` | phase5_matching.go:862, match.go:361 | HIGH | Missing AlphaAvoid |
| 4 | `ApplyMatchToProblem` applyFn | matching.go:156 | MED-HIGH | Wrong apply variant for fomatch |
| 5 | `functionTactic` | tactics.go:396 | HIGH | Adds Implies, should add premises |
| 6 | `propertyTactic` | tactics.go:345 | HIGH | Wrong structure + missing steps |
| 7 | `letTactic` | tactics.go:18 | MED-HIGH | Missing compile_expr_vocab |
| 8 | `ifTactic` | tactics.go:281 | MED | Wrong structure, missing attrib_goals |
| 9 | `witnessTactic` | tactics.go:437 | MED | Missing compile, validation, WitnessAst |
| 10 | `unfoldTactic` | tactics.go:216 | MED-HIGH | Missing schema lookup, renamings, has_premise |
| 11 | `MatchFromDefns` | phase5_goals.go:538 | MED | Only first def; list-valued match missing |
| 12 | `applyMatchRec` binder | match.go:419 | MED | Missing env tracking in clone_binder |
| 13 | `MatchSchema` fomatch | checker.go:357 | MED | Uses alt variant instead of non-alt |
| 14 | `ApplyMatchToProblem` freesyms | matching.go:173 | MED | Uses Alt freesyms; Python uses non-alt |
| 15 | `MatchRhsVars` list values | phase5_matching.go:747 | LOW | Cascades from #11 |
| 16 | `ApplyMatchGoalNode` lambda filter | phase5_matching.go:1236 | LOW-MED | Missing lambda-prem filter |
| 17 | `ApplyMatchGoalNode` nil env | phase5_matching.go:1236 | LOW-MED | Cascades from #2 |
| 18 | `witnessTactic` temporal check | tactics.go:437 | LOW | Missing goal_is_temporal guard |
| 19 | `unfoldFmla` HO match | tactics.go:266 | HIGH | Sub-bug of #10 |

---

## Files to modify (when fixing)

1. `proof/tactics.go` — BUG-5,6,7,8,9,10,18,19 (letTactic, ifTactic, propertyTactic, functionTactic, witnessTactic, unfoldTactic)
2. `proof/phase5_matching.go` — BUG-1,2,3,4,13,14,15,16,17 (CompileOneMatch, applyMatchAltRec, ApplyMatchToProblem, MatchRhsVars, ApplyMatchGoalNode)
3. `proof/match.go` — BUG-3,12 (ApplyMatch alpha_avoid, applyMatchRec binder env)
4. `proof/matching.go` — BUG-4,13,14 (ApplyMatchToProblem)
5. `proof/phase5_goals.go` — BUG-11 (MatchFromDefns)

## Verification

After each fix:
```bash
cd ~/ivy/goivy && make test
cd ~/ivy/goivy && go test ./proof/ -v -run TestAssumeTactic
XTRACE_OFF=1 time ./goivy_check ord_live.ivy
```
