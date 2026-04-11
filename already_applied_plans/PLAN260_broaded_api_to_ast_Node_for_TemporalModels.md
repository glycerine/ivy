# Plan: Fix all latent `*ast.TemporalModels` handling bugs in goal accessors and tactics

Created: 2026-04-11 04:30 UTC
Last updated: 2026-04-11 05:15 UTC (expanded scope to fix all latent bugs)

## Context

`make golden` now diverges at `~/ivy/goivy/log.red:254987` in a `proof.composeProofs` event. The visible mismatch is:

```
3617  Both: proof.tacticTactic name='skolemize' goal[0].Formula type=TemporalModels
3620  Both: ast.LF.__init__ id=2248 counter=2249
3623  Both: ast.LF.__init__ id=2249 counter=2250
3626  GO:   proof.composeProofs step=1/2 proofType=TacticTactic goal[0].Formula type=nil
3627  PY:   proof.composeProofs step=1/2 proofType=TacticTactic goal[0].Formula type=SchemaBody
```

The proximate cause is **Skolemize** silently producing `Formula = nil` when given a `*ast.TemporalModels` goal — but investigation revealed this is the tip of an iceberg: many goal accessors and tactics in goivy share the same latent bug. The user has explicitly asked to fix **all** known instances in this plan, not defer any.

## Investigation findings

### Only one problematic post-compile type

`*ast.TemporalModels` is the **only** non-`lg.Expr` type that flows through proof/tactic code as a goal Formula post-compilation. The other candidates were investigated and ruled out:

| Type | Status |
|---|---|
| `*ast.SchemaBody` | Already partially handled — `GoalConc` unwraps it. Its conclusion can itself be a `*ast.TemporalModels`. |
| `*ast.TemporalModels` (`ast/ast.go:1667`) | **The problematic type.** Has `Model Node` (e.g. `*temporal.NormalProgram`) and `Fmla Node` (typically `lg.Expr`). |
| `*ast.CompiledNode` (`ast/ast.go:1865`) | **Compile-time only** — only used in `compiler/`, `isolate/`, `ast/`. Never reaches `proof/`, `tactics/`, or `temporal/`. |

`*ast.TemporalModels` appears in **two contexts**:
1. Directly as `goal.Formula` (the case the trace at 254987 hits).
2. As the conclusion inside a `SchemaBody` (`goal.Formula.(*ast.SchemaBody).Conc()`).

### The codebase already has duplicate ad-hoc helpers for this

Two byte-identical copies of a helper that does what a broadened `CloneGoal` would do:

```go
// temporal/temporal.go:634
func cloneGoalWithASTConc(cfg *ast.AstConfig, goal *ast.LabeledFormula,
                          prems []ast.Node, conc ast.Node) *ast.LabeledFormula { ... }

// l2s/l2s.go:580  ← BYTE-IDENTICAL DUPLICATE
func cloneGoalWithASTConc(cfg *ast.AstConfig, goal *ast.LabeledFormula,
                          prems []ast.Node, conc ast.Node) *ast.LabeledFormula { ... }
```

`temporal/temporal.go:615` also has `findTemporalModels(goal)` that knows TemporalModels appears in both contexts. These helpers exist precisely because the central `proof.CloneGoal` is too narrow. Lifting the helper into `proof/goal.go` lets us **delete both copies**.

### Tactics that already correctly handle TemporalModels

These special-case `*ast.TemporalModels` and need no changes (they remain models for the pattern):

- `tactics/ivy_tactics.go:228 ApplyTempind` — checks `fmlaNode.(*ast.TemporalModels)` and recurses into `tm.Fmla`
- `tactics/ivy_tactics.go:320 ApplyTempcase` — same pattern
- `tactics/ivy_tactics.go:362 Vcgen` — explicitly requires TemporalModels and extracts inner formula
- `temporal/temporal.go ApplyInvariance` — uses `cloneGoalWithASTConc`
- `l2s/l2s.go ApplyL2S` — uses `cloneGoalWithASTConc`

### All latent bugs (every site that calls `GoalConc` then does logic ops on the result)

Compiled from a comprehensive scan:

| Severity | File / Function | Line | Operation | Python equivalent has TM handling? |
|---|---|---|---|---|
| **CRITICAL** | `proof/skolem.go SkolemizeGoal` + `SkolemizeFmla` | 17, 98 | Skolemization | YES — `ivy_proof.py:1443-1450` (the `apply_to_conc` is implicit via the explicit case) |
| HIGH | `proof/goal.go GoalConc` (root) | 21 | Returns nil silently | NO — Python's `goal_conc` returns the formula directly |
| HIGH | `proof/goal.go NormalizeGoal` | 105 | `il.NormalizeOps(conc)` | NO explicit; relies on Python duck typing — but we'll mirror by unwrapping with `ApplyToConc` |
| HIGH | `proof/goal.go GoalVocab` | 148, 172 | Variable extraction | YES — `ivy_proof.py:580` `conc_fmla = conc.fmla if isinstance(conc,ia.TemporalModels) else conc` |
| HIGH | `proof/goal.go GoalFree` | 237, 241 | Free variable collection | NO explicit, but symmetric with GoalVocab |
| HIGH | `proof/goal.go TrivialGoal` | 281, 287 | `lu.EqualModAlpha` | NO explicit |
| HIGH | `proof/goal.go CheckConcsMatch` | 299-300 | `lu.EqualModAlpha` | NO explicit |
| HIGH | `proof/tactics.go letTactic` | 54, 59 | `lg.Implies(cond, conc)` | Python tactic uses `goal_apply_to_conc(decl, lambda c: il.Implies(cond, c))` — passes c through, Implies handles |
| HIGH | `proof/tactics.go unfoldTactic` | 151, 162 | `unfoldFmla(conc, defns)` | YES — Python `goal_apply_to_conc(decl, lambda c: unfold_fmla(c, defns))` and `unfold_fmla` handles TemporalModels |
| HIGH | `proof/tactics.go ifTactic` | 205, 212, 217 | `lg.Implies` construction | Pass-through via `goal_apply_to_conc` |
| HIGH | `proof/tactics.go propertyTactic` | 265, 271, 274 | Cut + `Implies` | Pass-through via `goal_apply_to_conc` |
| HIGH | `proof/tactics.go functionTactic` | 318, 325 | `Implies(defFormula, conc)` | Pass-through via `goal_apply_to_conc` |
| HIGH | `proof/tactics.go witnessTactic` | 342, 370 | `applyWitness(conc, witMap)` | Likely uses `apply_to_conc` (substitution must reach inner formula) |
| HIGH | `proof/checker.go LookupSchema` | 163 | Cast to `*lg.Definition` | Definitions don't appear inside TemporalModels — type assertion bails out, but should not silently fail |
| HIGH | `proof/matching.go GoalFreeVars` | 197, 200 | Variable extraction | Symmetric with GoalVocab — needs `ApplyToConc`-style unwrap |
| HIGH | `proof/phase5_matching.go AddPremMatch` | 447, 448 | Premise conclusion extraction | Should error gracefully on TemporalModels (schemata don't match temporal goals) |
| HIGH | `proof/phase5_matching.go ParameterizeSchema` | 492 | Variable generation + match | Same — schemata don't match temporal goals |
| HIGH | `proof/phase5_goals.go GoalIsTemporal` | 117 | Detect temporal goals | YES — Python `ivy_proof.py:603-605` `return conc.temoral or isinstance(conc.formula,ia.TemporalModels)`. Currently Go misses this case. |
| MEDIUM | `proof/checker.go MatchSchema` | 301-302 | Nil check | YES — Python `ivy_proof.py:429` explicitly raises `NoMatch` for `isinstance(goal_conc(decl),ia.TemporalModels)` |
| MEDIUM | `proof/skolem.go varSubstGoal` | 256 | `lu.Substitute(conc, subs)` | Python `ivy_proof.py:1365-1368` uses `apply_to_conc(conc, lambda x: il.substitute(x,subst))` |
| MEDIUM | `proof/phase5_matching.go TransformDefnSchema` | 229-230 | Type assertion | Falls through gracefully |
| MEDIUM | `proof/phase5_matching.go CompileMatchList` | 587 | Variable extension (witness path) | Symmetric — needs unwrap |
| MEDIUM | `proof/phase5_matching.go CompileMatchFull` | 645-646 | Free symbol extension (witness path) | Symmetric — needs unwrap |
| MEDIUM | `proof/phase5_goals.go` (~10 sites) | 173, 334-335, 374, 411, 445, 535, 545, 551, 557, 617 | Various structural | Most are pass-through; the few with logic ops need `ApplyToConc` |
| MEDIUM-LOW | `proof/goal.go GoalSubst, GoalAddPrem, GoalRemovePrem` | 255, 261, 275 | Pass through to MakeGoal/CloneGoal | Pure pass-through — fixed for free by widening `MakeGoal`/`CloneGoal` |
| MEDIUM-LOW | `proof/checker.go forgetTactic` | 549 | Pass through to CloneGoal | Pass-through — free fix |
| MEDIUM-LOW | `proof/tactics.go goalAddPrem` | 505 | Pass through to MakeGoal | Pass-through — free fix |
| MEDIUM-LOW | `proof/phase5_goals.go GoalPrefixPrems` | 78 | Pass through to MakeGoal | Pass-through — free fix |
| LOW | `proof/register.go GoalConcFn` | 19 | Re-export | Update return type along with GoalConc |

## Approach

### Three pillars

1. **Broaden the goal API** to mirror Python's untyped duck typing:
   - `GoalConc` returns `ast.Node` (not `lg.Expr`)
   - `CloneGoal` and `MakeGoal` accept `ast.Node` for `conc` (not `lg.Expr`)
   - `SkolemizeFmla` takes/returns `ast.Node` and adds the TemporalModels case
   - `lg.Expr` already embeds `ast.Node` (`logic/node.go:8-9`), so widening parameter types is fully backward-compatible.

2. **Add Python-equivalent helpers** to `proof/goal.go`, mirroring the Python helpers in `ivy_proof.py`:
   ```go
   // ApplyToConc applies func to the conclusion, unwrapping TemporalModels if present.
   // Mirrors Python ivy_proof.py:1370-1373 apply_to_conc.
   func ApplyToConc(conc ast.Node, fn func(lg.Expr) lg.Expr) ast.Node {
       if tm, ok := conc.(*ast.TemporalModels); ok {
           if innerExpr, ok := tm.Fmla.(lg.Expr); ok {
               return tm.Clone([]ast.Node{fn(innerExpr)})
           }
           return tm
       }
       if expr, ok := conc.(lg.Expr); ok {
           return fn(expr)
       }
       return conc
   }

   // GoalApplyToConc clones a goal with fn applied to its conclusion.
   // Mirrors Python ivy_proof.py:1572-1573 goal_apply_to_conc.
   // NOTE: fn is called with the raw conclusion (ast.Node) — fn is responsible
   // for handling TemporalModels itself, OR the caller can wrap fn with ApplyToConc.
   func GoalApplyToConc(cfg *ast.AstConfig, goal *ast.LabeledFormula, fn func(ast.Node) ast.Node) *ast.LabeledFormula {
       return CloneGoal(cfg, goal, GoalPrems(goal), fn(GoalConc(goal)))
   }

   // GoalConcExpr returns the conclusion as an lg.Expr if possible, else nil.
   // Use when the caller needs lg.Expr for substitution, matching, or other
   // logic-level operations and wants the old "skip on TemporalModels" semantics.
   func GoalConcExpr(g *ast.LabeledFormula) lg.Expr {
       if e, ok := GoalConc(g).(lg.Expr); ok {
           return e
       }
       return nil
   }
   ```

3. **Fix every flagged tactic and goal accessor** to use `ApplyToConc` (or the appropriate strategy: pass-through, error, or recurse) so that TemporalModels goals flow through correctly.

### Site-by-site strategy

- **Pass-through sites** (Goal{Subst,AddPrem,RemovePrem}, forgetTactic, goalAddPrem, GoalPrefixPrems, etc.): no code change needed; widening `CloneGoal`/`MakeGoal` to `ast.Node` automatically fixes them.

- **Logic-operation sites** (NormalizeGoal, GoalVocab, GoalFree, TrivialGoal, CheckConcsMatch, GoalFreeVars, GoalIsTemporal, witnessTactic, unfoldTactic, varSubstGoal, the phase5_matching unwrapping sites): use `ApplyToConc` to unwrap TemporalModels and apply the operation to `tm.Fmla`. If the result must be wrapped back into TemporalModels (skolemize, witness substitution, normalize, unfold), `ApplyToConc` does the wrap automatically.

- **Tactics that build new formulas around the conclusion** (letTactic, ifTactic, propertyTactic, functionTactic): the new formula must be wrapped *inside* the TemporalModels, not around it. Use:
  ```go
  newConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
      return ctx.Implies(cond, c)  // or similar
  })
  ```
  This produces `TemporalModels(Model, Implies(cond, originalFmla))` instead of building `Implies(cond, TemporalModels(...))` which would be a type error.

- **Schema-matching sites that don't apply to temporal goals** (LookupSchema, MatchSchema, AddPremMatch, ParameterizeSchema, TransformDefnSchema, CompileMatchList, CompileMatchFull): explicitly check for TemporalModels and return a structured error matching Python's `NoMatch` (`ivy_proof.py:429`).

- **GoalIsTemporal**: add an explicit `*ast.TemporalModels` check, mirroring Python `ivy_proof.py:603-605`.

- **Skolemize**: use the broadened API. The TemporalModels case in `SkolemizeFmla` (and the corresponding case in the final univs-wrapping) handles everything.

## Files to modify

### A. Broaden the goal API

**`proof/goal.go`** — central API change:

1. **`GoalConc`**: change return type `lg.Expr` → `ast.Node`. New body:
   ```go
   func GoalConc(g *ast.LabeledFormula) ast.Node {
       if sb, ok := g.Formula.(*ast.SchemaBody); ok {
           return sb.Conc()
       }
       return g.Formula
   }
   ```

2. **`CloneGoal`**: change `conc lg.Expr` → `conc ast.Node`. Body unchanged.

3. **`MakeGoal`**: change `conc lg.Expr` → `conc ast.Node`. Body unchanged.

4. **Add helpers** as shown in Approach §2 above:
   - `ApplyToConc(conc ast.Node, fn func(lg.Expr) lg.Expr) ast.Node`
   - `GoalApplyToConc(cfg, goal, fn func(ast.Node) ast.Node) *ast.LabeledFormula`
   - `GoalConcExpr(g) lg.Expr`

5. **Fix internal logic-op sites** in goal.go:
   - `NormalizeGoal` (line 105):
     ```go
     newConc := ApplyToConc(GoalConc(g), il.NormalizeOps)
     return CloneGoal(cfg, g, normPrems, newConc)
     ```
   - `GoalVocab` (lines 148, 172): use `GoalConcExpr` for direct lg.Expr collection, but ALSO unwrap TemporalModels — mirror Python `ivy_proof.py:580` by extracting `tm.Fmla.(lg.Expr)` when conc is TemporalModels.
   - `GoalFree` (lines 237, 241): symmetric — unwrap TemporalModels before extracting free vars.
   - `TrivialGoal` (lines 281, 287): use `GoalConcExpr`. Returns false if conc is non-Expr (preserves current behavior; Python behavior is identical here since trivial check requires structural equality).
   - `CheckConcsMatch` (lines 299, 300): use `GoalConcExpr`. Same reasoning — temporal goal-to-goal equality goes through a different path.
   - Pass-through sites (`GoalSubst:255`, `GoalAddPrem:261`, `GoalRemovePrem:275`): no change needed.

### B. Delete duplicate `cloneGoalWithASTConc` helpers

Now that `proof.CloneGoal` accepts `ast.Node`:

1. **`temporal/temporal.go`**:
   - Delete `cloneGoalWithASTConc` definition (line 634).
   - Update one call site at line 606 to use `proof.CloneGoal(...)` directly.

2. **`l2s/l2s.go`** and **`l2s/shared.go`**:
   - Delete `cloneGoalWithASTConc` definition in `l2s/l2s.go:580`.
   - Update call sites in `l2s/shared.go:650, 727` to use `proof.CloneGoal(...)` directly.

Net code reduction: ~30 lines + cleaner inter-package boundaries.

### C. Fix Skolemize (the trace's proximate cause)

**`proof/skolem.go`**:

1. **`SkolemizeFmla`** signature change: `(fmla lg.Expr, ...) lg.Expr` → `(fmla ast.Node, ...) ast.Node`. The internal `rec` follows.

2. **Add TemporalModels case** in `rec`:
   ```go
   if tm, ok := fmla.(*ast.TemporalModels); ok {
       return tm.Clone([]ast.Node{rec(tm.Args()[0], pos)})
   }
   ```

3. **Guard `IsExists`/`IsForall` with type assertion** (they take `lg.Expr`):
   ```go
   var isE, isA bool
   if expr, ok := fmla.(lg.Expr); ok {
       isE = il.IsExists(expr)
       isA = il.IsForall(expr)
   }
   ```

4. **Update final wrapping** (currently lines 208-215) to handle TemporalModels output, mirroring Python `ivy_proof.py:1447-1453`:
   ```go
   if len(univs) > 0 {
       if tm, ok := body.(*ast.TemporalModels); ok {
           innerExpr, _ := tm.Args()[0].(lg.Expr)
           var quantBody lg.Expr
           if pos {
               quantBody = il.Exists(univs, innerExpr)
           } else {
               quantBody = il.ForAll(univs, innerExpr)
           }
           body = tm.Clone([]ast.Node{quantBody})
       } else if expr, ok := body.(lg.Expr); ok {
           if pos {
               body = il.Exists(univs, expr)
           } else {
               body = il.ForAll(univs, expr)
           }
       }
   }
   return body
   ```

5. **`SkolemizeGoal`** (line 17): no change needed at line 73 — `GoalConc(g)` now returns `ast.Node`, `SkolemizeFmla` accepts `ast.Node`, `CloneGoal` accepts `ast.Node`. The flow naturally handles TemporalModels.

6. **`varSubstGoal`** (line 246): use `ApplyToConc` mirroring Python `ivy_proof.py:1365-1368`:
   ```go
   newConc := ApplyToConc(GoalConc(goal), func(e lg.Expr) lg.Expr {
       result, err := lu.Substitute(e, subs)
       if err != nil { return e }
       return result
   })
   return CloneGoal(cfg, goal, newPrems, newConc)
   ```

### D. Fix all tactics in `proof/tactics.go`

For each tactic, replace direct `GoalConc(goal)` + new-formula construction with `ApplyToConc` so the construction happens inside the TemporalModels wrapper.

1. **`letTactic`** (lines 22-72): wrap implication construction:
   ```go
   newConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
       return ctx.Implies(cond, c)
   })
   subgoal := CloneGoal(pc.astCfg(), goal, GoalPrems(goal), newConc)
   ```

2. **`unfoldTactic`** (lines 127-190): use `ApplyToConc` for the substitution:
   ```go
   newConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
       return unfoldFmla(c, defns)
   })
   ```

3. **`ifTactic`** (lines 193-244): wrap each branch's implication construction in `ApplyToConc`.

4. **`propertyTactic`** (lines 248-289): wrap cut + modified implication construction in `ApplyToConc`.

5. **`functionTactic`** (lines 294-330): wrap function-implication construction in `ApplyToConc`.

6. **`witnessTactic`** (lines 334-385): wrap `applyWitness` substitution in `ApplyToConc`.

7. **`goalAddPrem`** (lines 500-507): pure pass-through — no code change after CloneGoal/MakeGoal widening.

### E. Fix `proof/checker.go`

1. **`LookupSchema`** (line 163): `conc := GoalConc(d)` — type assertion to `*lg.Definition` already fails for TemporalModels and falls through; behavior is preserved. The compile error from the return-type change is the only fix needed here.

2. **`MatchSchema`** (lines 300-304): add explicit TemporalModels check matching Python `ivy_proof.py:429`:
   ```go
   if _, isTM := GoalConc(goal).(*ast.TemporalModels); isTM {
       return nil, &NoMatch{Msg: "schema matching does not apply to temporal-models goals"}
   }
   ```

3. **`forgetTactic`** (line 549): pure pass-through — no change.

### F. Fix `proof/matching.go`

1. **`buildMatchProblem`** (lines 92, 112-114): if either schema or decl conclusion is TemporalModels, return `NoMatch` (mirroring `MatchSchema` behavior). Schemata don't match temporal goals.

2. **`GoalFreeVars`** (lines 195-200): unwrap TemporalModels — apply `lu.FreeVariables` to `tm.Fmla.(lg.Expr)`:
   ```go
   conc := GoalConc(g)
   var concExpr lg.Expr
   if tm, ok := conc.(*ast.TemporalModels); ok {
       concExpr, _ = tm.Fmla.(lg.Expr)
   } else {
       concExpr, _ = conc.(lg.Expr)
   }
   if concExpr == nil {
       return nil  // or error
   }
   // ... continue with concExpr
   ```

### G. Fix `proof/phase5_matching.go`

For all matching sites that hit non-Expr conclusions, return a structured error that schemata don't apply to temporal goals:

- **`TransformDefnSchema`** (lines 229-230): falls through to "not a definition" — preserve behavior.
- **`AddPremMatch`** (lines 447-448): if any premise conclusion is TemporalModels, return NoMatch.
- **`ParameterizeSchema`** (line 492): error on TemporalModels.
- **`CompileMatchList`** (line 587, witness path): unwrap TemporalModels for variable extension.
- **`CompileMatchFull`** (lines 645-646, witness path): unwrap TemporalModels for free symbol extension.

### H. Fix `proof/phase5_goals.go`

1. **`GoalIsTemporal`** (line 117): add explicit TemporalModels check, mirroring Python `ivy_proof.py:603-605`:
   ```go
   func GoalIsTemporal(x *ast.LabeledFormula) bool {
       conc := GoalConc(x)
       if _, ok := conc.(*ast.TemporalModels); ok {
           return true
       }
       // existing NamedBinder check on conc as lg.Expr
       if expr, ok := conc.(lg.Expr); ok {
           // ... existing logic
       }
       return false
   }
   ```

2. **Other ~10 sites** (173, 334-335, 374, 411, 445, 535, 545, 551, 557, 617): audit each — most are pass-through to CloneGoal and need no change. The ones doing logic operations should use `GoalConcExpr` (preserves nil-on-non-Expr semantics, fail-fast where the existing code already handled nil) or `ApplyToConc` (lift logic-op into the TemporalModels wrapper).

### I. Other call sites (cross-package)

`go build ./...` will catch all compile errors from the `GoalConc` return-type change. For each:

- `compiler/ivy_compile.go t2pGoalConc` (line 1993): refactor to `return proof.GoalConcExpr(g)` — it's a duplicate of GoalConcExpr semantics.
- `check/isolate_check.go:672`: `_ = proof.GoalConc(goal)` — change has no effect on `_`.
- `tactics/ivy_tactics.go:237, 329, 367`: these are inside `ApplyTempind/Tempcase/Vcgen` which already handle TemporalModels — switching to `GoalConcExpr` is fine since they fall back to handling TemporalModels via `findTemporalModels`.
- `ranking/ranking.go:233, 255`: rank ranks a goal's conclusion — likely needs an lg.Expr; use `GoalConcExpr` and bail out gracefully on TemporalModels (ranking doesn't apply).
- `proof/proof_test.go:411`: test code, update as needed.
- `proof/register.go:19`: update to return `ast.Node`; module config consumers must also accept `ast.Node`.

### J. Update tests

- `proof/proof_test.go:585-633` `TestSkolemizeFmlaSimple`/`TestSkolemizeFmlaExists`: pass `*lg.ForAll`/`*lg.Exists` to `SkolemizeFmla` — these are `lg.Expr` and therefore also `ast.Node`, so the call still works after signature change. Result type assertion `result.(*lg.ForAll)` works because the result is now `ast.Node` and the assertion still finds the concrete type.
- `proof/proof_test.go:441` `TestCloneGoal`: verify the wider parameter type still accepts the test value.

## Verification

1. **Compile** (catches every call site needing a rename or update):
   ```
   cd ~/ivy/goivy && go build ./...
   ```
   Iterate until clean. Most changes will be either:
   - No-op (pass-through to widened CloneGoal/MakeGoal)
   - `GoalConc` → `GoalConcExpr` (callers that use the result as lg.Expr)
   - `GoalConc(g)` → `ApplyToConc(GoalConc(g), fn)` (callers that transform the conclusion)

2. **Unit tests**:
   ```
   cd ~/ivy/goivy && go test ./proof/... ./tactics/... ./temporal/... ./l2s/...
   ```

3. **Microtest** (fast confidence-builder before the golden run):
   - Construct a `LabeledFormula` whose `Formula` is a `*ast.TemporalModels` containing a `*lg.ForAll`.
   - Call `SkolemizeGoal(cfg, lf, true)`.
   - Verify the result's `Formula` is a `*ast.SchemaBody` whose last element is the (modified) `TemporalModels`(with the inner ForAll skolemized) and whose preceding elements are `*ast.ConstantDecl` premises.

4. **Golden test**:
   ```
   cd ~/ivy/goivy && make golden 2>&1 | tee log.parser.red
   ```
   Expected: divergence at log.red:254987 is gone. Step 1/2 should now show `goal[0].Formula type=SchemaBody` matching Python. Find the next divergence (if any) and address in a follow-up plan.

5. **Run the entire xtrace suite** to catch any silent regressions in the now-fixed `proof/checker.go`, `proof/matching.go`, `proof/phase5_*.go`, and `proof/tactics.go`. The previous behaviors (silently skipping logic operations on TemporalModels) may have masked latent issues that are exposed once the operations actually run.

## Risks and rollback

- **Risk: subtle semantic change in `GoalConc`** — previously returned `nil` for non-`lg.Expr` formulas; now returns the AST node. Some callers may have relied on the nil return as a "skip me" signal. Migration to `GoalConcExpr` preserves the old semantics where appropriate; the comprehensive site-by-site review above flags every place this matters.
- **Risk: tactic transformations now apply where they previously silently no-op'd** — letTactic, ifTactic, etc. previously got `nil` for the conclusion and produced `Implies(cond, nil)`. Now they get the conclusion wrapped via ApplyToConc. The Python behavior IS to apply the transformation, so this is correct, but it means the test surface increases. The golden test catches divergences if our implementation differs from Python's.
- **Risk: `ApplyToConc` ignores non-Expr non-TemporalModels nodes** — our `ApplyToConc` returns `conc` unchanged if it's neither `*ast.TemporalModels` nor `lg.Expr`. This matches Python's `apply_to_conc` (which would crash on something else, but in practice only TemporalModels and lg.Expr appear).
- **Risk: schema matching now correctly errors on TemporalModels** — Python errors gracefully via `NoMatch` (`ivy_proof.py:429`). Our previous Go silently returned nil. Producing the structured error may surface code paths that were previously masked. The error message should match Python's where possible to keep traces aligned.
- **Risk: `t2pGoalConc` divergence** — `compiler/ivy_compile.go:1993` is a private wrapper duplicating old GoalConc semantics. After our fix, this wrapper is equivalent to `GoalConcExpr`. We refactor it to `return proof.GoalConcExpr(g)` to keep semantics aligned.
- **Rollback**: `git checkout` the affected files. Changes are localized to `proof/`, `temporal/`, `l2s/`, and a handful of cross-package call sites. No schema, no on-disk format change.

## Out of scope

- Refactoring `IsForall`/`IsExists` to take `ast.Node` — local type assertions in `SkolemizeFmla` are sufficient.
- Adding new exhaustive tests for every fixed tactic — the golden test exercises them end-to-end. We add one microtest for SkolemizeGoal+TemporalModels as a fast smoke test.
- Promoting `*ast.TemporalModels` to satisfy `lg.Expr` — discussed and rejected (Model field semantics, hash equality contracts, slippery slope).
- The next divergence after this fix — separate follow-up plan once we run `make golden` and see what surfaces.

## Implementation checklist

### A. API broadening (`proof/goal.go`)
- [ ] Change `GoalConc` to return `ast.Node`.
- [ ] Add `ApplyToConc(conc, fn) ast.Node` helper (mirrors Python `apply_to_conc`).
- [ ] Add `GoalApplyToConc(cfg, goal, fn) *LabeledFormula` helper (mirrors Python `goal_apply_to_conc`).
- [ ] Add `GoalConcExpr(g) lg.Expr` helper.
- [ ] Change `CloneGoal` `conc` parameter to `ast.Node`.
- [ ] Change `MakeGoal` `conc` parameter to `ast.Node`.

### B. Internal goal.go fixes
- [ ] `NormalizeGoal` — use `ApplyToConc(... , il.NormalizeOps)`.
- [ ] `GoalVocab` — unwrap TemporalModels (mirror Python `ivy_proof.py:580`).
- [ ] `GoalFree` — unwrap TemporalModels symmetrically.
- [ ] `TrivialGoal` — use `GoalConcExpr`.
- [ ] `CheckConcsMatch` — use `GoalConcExpr`.

### C. Delete duplicate helpers
- [ ] Delete `cloneGoalWithASTConc` from `temporal/temporal.go:634`. Update `temporal/temporal.go:606` to use `proof.CloneGoal`.
- [ ] Delete `cloneGoalWithASTConc` from `l2s/l2s.go:580`. Update `l2s/shared.go:650, 727` to use `proof.CloneGoal`.

### D. Skolemize fix (`proof/skolem.go`)
- [ ] Change `SkolemizeFmla` signature to `(fmla ast.Node, ...) ast.Node`.
- [ ] Add TemporalModels case in `rec`.
- [ ] Guard `IsForall`/`IsExists` with `lg.Expr` type assertion.
- [ ] Update final univs-wrapping logic for TemporalModels output.
- [ ] Update `varSubstGoal` to use `ApplyToConc`.

### E. Tactics fixes (`proof/tactics.go`)
- [ ] `letTactic` — wrap Implies construction in `ApplyToConc`.
- [ ] `unfoldTactic` — wrap unfoldFmla in `ApplyToConc`.
- [ ] `ifTactic` — wrap Implies construction in `ApplyToConc`.
- [ ] `propertyTactic` — wrap cut+Implies in `ApplyToConc`.
- [ ] `functionTactic` — wrap Implies in `ApplyToConc`.
- [ ] `witnessTactic` — wrap applyWitness in `ApplyToConc`.

### F. Checker/matching fixes
- [ ] `proof/checker.go MatchSchema` — explicit TemporalModels check, return NoMatch.
- [ ] `proof/matching.go buildMatchProblem` — TemporalModels check, return NoMatch.
- [ ] `proof/matching.go GoalFreeVars` — unwrap TemporalModels.
- [ ] `proof/phase5_matching.go AddPremMatch` — TemporalModels check.
- [ ] `proof/phase5_matching.go ParameterizeSchema` — TemporalModels check.
- [ ] `proof/phase5_matching.go CompileMatchList/Full` (witness paths) — unwrap TemporalModels.
- [ ] `proof/phase5_goals.go GoalIsTemporal` — add TemporalModels case.
- [ ] `proof/phase5_goals.go` other ~10 sites — audit and fix per category.

### G. Cross-package fixes
- [ ] `compiler/ivy_compile.go t2pGoalConc` — refactor to `proof.GoalConcExpr`.
- [ ] `tactics/ivy_tactics.go` (3 sites in already-correct tactics) — switch to `GoalConcExpr` where appropriate.
- [ ] `ranking/ranking.go` (2 sites) — use `GoalConcExpr`, bail gracefully on nil.
- [ ] `proof/register.go` — update GoalConcFn return type.
- [ ] `check/isolate_check.go:672` — no functional change (`_ =`).
- [ ] `proof/proof_test.go` — update test expectations as needed.

### H. Verification
- [ ] `go build ./...` — clean.
- [ ] `go test ./proof/... ./tactics/... ./temporal/... ./l2s/...` — green.
- [ ] Microtest for SkolemizeGoal + TemporalModels.
- [ ] `make golden` — confirm divergence at 254987 is gone; record next divergence (if any) for follow-up.
