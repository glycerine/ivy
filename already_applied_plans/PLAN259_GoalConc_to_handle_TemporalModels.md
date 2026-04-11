# Plan: Fix `GoalConc`/`SkolemizeFmla` to handle `*ast.TemporalModels` (skolemize divergence at log.red:254987)

Created: 2026-04-11 04:30 UTC

## Context

`make golden` now diverges at `~/ivy/goivy/log.red` line 254987 in a `proof.composeProofs` event. After the previous fix (canon walker for SMT2 serialization) advanced past line 254714, the new mismatch is:

```
3608  Both: proof.composeProofs ENTER nproofs=2 ndecls=1
3611  Both: proof.composeProofs step=0/2 proofType=TacticTactic goal[0].Formula type=TemporalModels
3614  Both: proof.ApplyProof ENTER proofType=TacticTactic goal[0].Formula type=TemporalModels
3617  Both: proof.tacticTactic name='skolemize' goal[0].Formula type=TemporalModels
3620  Both: ast.LF.__init__ id=2248 counter=2249
3623  Both: ast.LF.__init__ id=2249 counter=2250
3626  GO:   proof.composeProofs step=1/2 proofType=TacticTactic goal[0].Formula type=nil
3627  PY:   proof.composeProofs step=1/2 proofType=TacticTactic goal[0].Formula type=SchemaBody
```

Both Go and Python:
- Enter `composeProofs` with `nproofs=2 ndecls=1` (matches)
- Process step 0 with `goal[0].Formula type=TemporalModels` (matches)
- Dispatch the `skolemize` tactic (matches)
- Create exactly two new `LabeledFormula` instances (id=2248, id=2249) inside skolemize (matches)

But after skolemize returns, the resulting first goal has:
- **Go**: `Formula = nil`
- **Python**: `Formula = SchemaBody`

This is a translation correctness bug, not a fingerprint mismatch.

## Root cause

`proof/goal.go:21-35`:

```go
func GoalConc(g *ast.LabeledFormula) lg.Expr {
    if sb, ok := g.Formula.(*ast.SchemaBody); ok {
        conc := sb.Conc()
        if conc != nil {
            if ln, ok := conc.(lg.Expr); ok {
                return ln
            }
        }
        return nil
    }
    if ln, ok := g.Formula.(lg.Expr); ok {
        return ln
    }
    return nil
}
```

When `g.Formula` is `*ast.TemporalModels` (defined at `ast/ast.go:1667`), the type assertion `g.Formula.(lg.Expr)` fails — TemporalModels is in the `ast` package, not `logic`, and does not implement `lg.Expr`. So `GoalConc` returns `nil`.

Python's `goal_conc` (`ivy_proof.py:481-482`) returns the formula directly, regardless of type:

```python
def goal_conc(g):
    return g.formula.conc() if isinstance(g.formula,ia.SchemaBody) else g.formula
```

### Cascade in Go's `SkolemizeGoal` (`proof/skolem.go:62-90`)

1. `prems := GoalPrems(g)` → `nil` (Formula is TemporalModels, not SchemaBody)
2. `conc := GoalConc(g)` → **`nil`** (TemporalModels not lg.Expr) ★
3. `if conc != nil` → false, so `SkolemizeFmla` is **never called**
4. `CloneGoal(cfg, g, [], nil)` → since `len(prems) == 0`, `formula = conc = nil`
5. Returned goal has `Formula = nil`
6. After `rec`, since `skfuns` is empty (skolemization was skipped), the final `CloneGoal(goal, [], GoalConc(goal))` again has empty prems and nil conc → `Formula = nil`

### Python's flow (correct)

`ivy_proof.py:1379-1400`:

1. `goal_conc(goal)` returns the `TemporalModels` directly
2. `skolemize_fmla(TemporalModels, ...)` matches the case at `ivy_proof.py:1443-1444`:
   ```python
   if isinstance(fmla,ia.TemporalModels):
       return fmla.clone([rec(fmla.args[0],pos)])
   ```
   It recurses into `args[0]` (the inner formula), which contains quantifiers that get skolemized — this populates `skfuns`.
3. After `rec`, since `skfuns` is non-empty, `clone_goal(goal, [ConstantDecl(s) for s in skfuns]+[], cloned_TM)` produces `SchemaBody(ConstantDecls..., cloned_TM)`.
4. Final `Formula = SchemaBody`.

### The four interlocking type-system mismatches

| Component | Python | Go (current) |
|---|---|---|
| `goal_conc` / `GoalConc` | returns formula as-is | filters to `lg.Expr`, returns nil otherwise |
| `clone_goal` / `CloneGoal` | accepts any node as `conc` | only accepts `lg.Expr` |
| `make_goal` / `MakeGoal` | accepts any node as `conc` | only accepts `lg.Expr` |
| `skolemize_fmla` / `SkolemizeFmla` | takes any node, has TemporalModels case | takes `lg.Expr`, no TemporalModels case |

## Approach

**Mechanical port**: broaden the four Go functions to take/return `ast.Node`, mirroring Python's untyped duck-typed interface.

**Key enabler** — `lg.Expr` already embeds `ast.Node` (`logic/node.go:8-9`):

```go
type Expr interface {
    ast.Node
    NodeSort() Sort
    Children() []Expr
    Equal(Expr) bool
    Sexp() NodeKey
}
```

So broadening **parameter types** from `lg.Expr` to `ast.Node` is **backward-compatible** — any existing caller passing an `lg.Expr` value continues to compile because `lg.Expr` IS an `ast.Node`. Only the **return type** of `GoalConc` changing from `lg.Expr` to `ast.Node` requires updates at call sites that store the result in a typed `lg.Expr` variable or pass it to a function expecting `lg.Expr`.

For minimal disruption at GoalConc call sites, we add a helper:

```go
// GoalConcExpr returns the conclusion as an lg.Expr if possible, else nil.
// Use when the caller needs lg.Expr for substitution, matching, or other
// logic-level operations. Use GoalConc when the caller passes the
// conclusion through to CloneGoal/MakeGoal or just nil-checks it.
func GoalConcExpr(g *ast.LabeledFormula) lg.Expr {
    if e, ok := GoalConc(g).(lg.Expr); ok {
        return e
    }
    return nil
}
```

Each call site that previously did `conc := GoalConc(g)` and used `conc` for logic operations renames to `GoalConcExpr(g)`. Pass-through call sites (the majority — they immediately pass `GoalConc(g)` into `CloneGoal`/`MakeGoal`) keep `GoalConc(g)` since `CloneGoal`/`MakeGoal` now accept `ast.Node`.

## Files to modify

### 1. `proof/goal.go` — broaden the goal API

Change `GoalConc` to mirror Python's `goal_conc`:

```go
// GoalConc returns the conclusion of a goal.
// If the goal's formula is a SchemaBody, returns the last element (conclusion).
// Otherwise returns the formula itself, regardless of type.
// Mirrors Python ivy_proof.py:481-482 goal_conc.
func GoalConc(g *ast.LabeledFormula) ast.Node {
    if sb, ok := g.Formula.(*ast.SchemaBody); ok {
        return sb.Conc()
    }
    return g.Formula
}
```

Add `GoalConcExpr` (helper for lg.Expr-needing call sites).

Change `CloneGoal` `conc` parameter from `lg.Expr` to `ast.Node`. The body already declares `var formula ast.Node`; just change the parameter type:

```go
func CloneGoal(cfg *ast.AstConfig, goal *ast.LabeledFormula, prems []ast.Node, conc ast.Node) *ast.LabeledFormula {
    var formula ast.Node
    if len(prems) > 0 {
        elems := make([]ast.Node, len(prems)+1)
        copy(elems, prems)
        elems[len(prems)] = conc
        formula = cfg.NewSchemaBody(elems...)
    } else {
        formula = conc
    }
    return goal.CloneWithFreshID([]ast.Node{goal.Label, formula})
}
```

Change `MakeGoal` `conc` parameter the same way.

**Internal users in goal.go** — update call sites:
- `NormalizeGoal` (line 105): `conc := GoalConc(g); if conc != nil { conc = il.NormalizeOps(conc) }` — `NormalizeOps` takes `lg.Expr` → use `GoalConcExpr`.
- `GoalVocab` (line 148, 172): collects formulas into `[]lg.Expr` → use `GoalConcExpr`.
- `GoalFree` (lines 237, 241): passes to `recFmla` which takes `lg.Expr` → use `GoalConcExpr`.
- `GoalSubst` (line 255), `GoalAddPrem` (261), `GoalRemovePrem` (275): all pass through to `MakeGoal`/`CloneGoal` → no change.
- `TrivialGoal` (lines 281, 287): uses in `lu.EqualModAlpha` which takes `lg.Expr` → use `GoalConcExpr`.
- `CheckConcsMatch` (lines 299, 300): same → use `GoalConcExpr`.

### 2. `proof/skolem.go` — handle TemporalModels in skolemize

Change `SkolemizeFmla` signature from `(fmla lg.Expr, ...) lg.Expr` to `(fmla ast.Node, ...) ast.Node`. The internal `rec` function follows the same change.

The type switch already dispatches on concrete types (`*lg.Not`, `*lg.Implies`, `*lg.And`, `*lg.Or`), which works on the broader `ast.Node` interface unchanged.

**Add a TemporalModels case** at the start of `rec`:

```go
// Mirror Python ivy_proof.py:1443-1444:
//   if isinstance(fmla,ia.TemporalModels):
//       return fmla.clone([rec(fmla.args[0],pos)])
if tm, ok := fmla.(*ast.TemporalModels); ok {
    return tm.Clone([]ast.Node{rec(tm.Args()[0], pos)})
}
```

The current code calls `il.IsExists(fmla)` and `il.IsForall(fmla)` (`ivylogic/ivylogic.go:154-163`), both of which take `lg.Expr`. Guard with a type assertion:

```go
var isE, isA bool
if expr, ok := fmla.(lg.Expr); ok {
    isE = il.IsExists(expr)
    isA = il.IsForall(expr)
}
```

**Update the final wrapping logic** (`proof/skolem.go:208-215`) to handle TemporalModels output, mirroring Python `ivy_proof.py:1447-1453`:

```python
if univs:
    quant = il.Exists if pos else il.ForAll
    if isinstance(body,ia.TemporalModels):
        body = body.clone([quant(univs,body.args[0])])
    else:
        body = quant(univs,body)
```

Go equivalent:

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

**Update `SkolemizeGoal`'s `rec`** (line 73): the `conc := GoalConc(g)` line is now correct as-is — `GoalConc` returns `ast.Node`, `SkolemizeFmla` accepts `ast.Node`, `CloneGoal` accepts `ast.Node`. The flow naturally handles TemporalModels.

**Update `varSubstGoal`** (line 256): uses `GoalConc(goal)` then `lu.Substitute(conc, subs)` which expects `lg.Expr` → use `GoalConcExpr`.

### 3. Other call sites — rename `GoalConc` → `GoalConcExpr` where lg.Expr is needed

Strategy: change `GoalConc` to return `ast.Node`, then run `go build ./...` and let the compiler identify every call site that needs an update. For each error:
- If the caller passes the result through to `CloneGoal`/`MakeGoal`: no change needed (parameter type now accepts `ast.Node`).
- If the caller does logic operations or stores in `lg.Expr` variable: rename `GoalConc` → `GoalConcExpr` at that site.
- If the caller does a type assertion like `conc.(*lg.Definition)` (e.g., `proof/checker.go:163`): no change needed — type assertions work on both `lg.Expr` and `ast.Node`.

Files with call sites (from grep):
- `proof/checker.go` (3 sites — line 163 type-asserts, line 301 nil-checks, line 549 passthrough)
- `proof/matching.go` (3 sites)
- `proof/phase5_matching.go` (~9 sites)
- `proof/phase5_goals.go` (~11 sites)
- `proof/tactics.go` (~8 sites)
- `proof/register.go` (1 site)
- `compiler/ivy_compile.go` (2 sites — note `t2pGoalConc` at line 1993 is a private wrapper duplicating GoalConc; can become a one-liner that calls `proof.GoalConcExpr`, or stay independent)
- `check/isolate_check.go` (1 site, just `_ = ...`)
- `tactics/ivy_tactics.go` (3 sites)
- `ranking/ranking.go` (2 sites)
- `proof/proof_test.go` (1 site)

### 4. Tests

`proof/proof_test.go:585-633` — `TestSkolemizeFmlaSimple` and `TestSkolemizeFmlaExists` pass `*lg.ForAll`/`*lg.Exists` to `SkolemizeFmla`. These are `lg.Expr` values, which are also `ast.Node`, so the call still works after the signature change. Result type assertions like `result.(*lg.ForAll)` still work because `*lg.ForAll` implements both `lg.Expr` and `ast.Node`.

`proof/proof_test.go:441` — `TestCloneGoal` calls `CloneGoal(testAstCfg, lf, nil, c)` where `c` is some test value. Should still work because the parameter is now wider.

## Verification

1. **Compile**:
   ```
   cd ~/ivy/goivy && go build ./...
   ```
   The compiler is the primary tool for finding the call sites that need `GoalConc` → `GoalConcExpr` renames. Iterate until clean.

2. **Unit tests**:
   ```
   cd ~/ivy/goivy && go test ./proof/... ./tactics/...
   ```

3. **Golden test**:
   ```
   cd ~/ivy/goivy && make golden 2>&1 | tee log.parser.red
   ```
   Expected: divergence at log.red:254987 is gone. Both sides should now show `goal[0].Formula type=TemporalModels` then `type=SchemaBody` after skolemize. Find the next divergence (if any) and address in a follow-up plan.

4. **Microtest for the fix** (optional, fast confidence-builder):
   - Construct a `LabeledFormula` whose `Formula` is a `*ast.TemporalModels` containing a `*lg.ForAll`.
   - Call `SkolemizeGoal(cfg, lf, true)`.
   - Verify the result's `Formula` is a `*ast.SchemaBody` whose last element is the (modified) `TemporalModels` and whose preceding elements are `*ast.ConstantDecl` premises (the skolem function declarations).

## Risks and rollback

- **Risk: missed call sites** — the compiler catches all signature mismatches. Runtime panics from nil ast.Node interface boxes are unlikely because we're not changing nil handling.
- **Risk: subtle semantic change in `GoalConc`** — previously, calling `GoalConc(g)` on a goal whose SchemaBody had a non-`lg.Expr` conclusion returned `nil`; now it returns the AST node. Some callers may have been relying on the nil return as a "skip me" signal. These callers must be migrated to `GoalConcExpr`, which preserves the nil-on-non-lg.Expr semantics. This is what we want.
- **Risk: `t2pGoalConc` and `GoalConcExpr` divergence** — `compiler/ivy_compile.go:1993` defines `t2pGoalConc` which duplicates the old GoalConc semantics. After our fix, this wrapper is equivalent to `GoalConcExpr`. We can leave it independent, or refactor it to `return proof.GoalConcExpr(g)`. Choose the latter for less duplication.
- **Rollback**: revert the modified files. Changes are contained to the `proof/` package and a handful of call sites in dependent packages.

## Out of scope

- Changing `GoalPrems` — already returns `[]ast.Node`, sufficient.
- Refactoring `IsForall`/`IsExists` to take `ast.Node` — adding a type assertion in `SkolemizeFmla` is local enough.
- Adding new tests for `SkolemizeGoal` with TemporalModels — the golden test covers this end-to-end.
- Fixing the next divergence after this one — separate follow-up plan.

## Implementation checklist

- [ ] Read `proof/goal.go`, `proof/skolem.go`, `proof/checker.go`, and `proof/proof_test.go` to confirm structure.
- [ ] Modify `GoalConc` to return `ast.Node`, add `GoalConcExpr` helper.
- [ ] Modify `CloneGoal` and `MakeGoal` to take `ast.Node` for `conc`.
- [ ] Modify `SkolemizeFmla` to take/return `ast.Node` and add the TemporalModels case (both in `rec` and in the final univs-wrapping logic).
- [ ] Update internal goal.go callers (`NormalizeGoal`, `GoalVocab`, `GoalFree`, `TrivialGoal`, `CheckConcsMatch`) to use `GoalConcExpr` where lg.Expr is needed.
- [ ] Update `varSubstGoal` in skolem.go to use `GoalConcExpr`.
- [ ] `go build ./...` — fix every reported call site (rename `GoalConc` → `GoalConcExpr` or no change if passthrough).
- [ ] Optionally simplify `t2pGoalConc` in `compiler/ivy_compile.go` to `return proof.GoalConcExpr(g)`.
- [ ] `go test ./proof/... ./tactics/...` — verify no regressions.
- [ ] `make golden` — confirm divergence at 254987 is gone; report next divergence (if any).
