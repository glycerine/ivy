# Plan: Match Python's duck-typed TemporalModels handling in Go tactics

Created: 2026-04-20

## Context

`Test2hrOrdLive` (parser/golden_test.go:457) diverges from Python at
`log.golden.2hr` line 10956214:

```
- (Variable name:P sort:(UninterpretedSort name:proc))
+ (Symbol   name:_P sort:(UninterpretedSort name:proc))
```

(and M/A/T) inside `l2s_globally_0 .. l2s_globally_2`. The immediate trigger
is the temporal property at `ord_live.ivy:2429-2450`:

```
explicit temporal property [memc2cf_live]
(forall P,M,A,T. globally eventually memc.cpl_fair(P,M,A,T))
-> forall T. (globally rfn.evcomplete(T) -> eventually cf_cpl(T))
proof {
    tactic tempcase with P=ref.evs(T).p, M=ref.evs(T).m, A=ref.evs(T).a;
    tactic skolemizenp;
    instantiate with P=_P, M=_M, A=_A, T=_T    // <-- plain form: WitnessTactic
    ...
    tactic l2s_auto2 with { ... }
}
```

Python log line 66988 shows post-`instantiate` goal conclusion with the
`forall` already removed and variables substituted:
`|= □ ⬦ memc.cpl_fair(_P,_M,_A,_T) -> ...`

Go's `neg_prop_init` dump (line 67018) still shows
`(forall A. (forall M. (forall P. (forall T. $l2s_init ...memc.cpl_fair(P,M,A,T))))) -> ...`
— the `forall`s survived, variables never got replaced with symbols.
Downstream `initGlobally` (check/l2s_auto.go:745) then walks this formula,
collects `P,M,A,T` as free variables, and builds `l2s_globally_{0,1,2}` with
`Variable` terms instead of `Symbol`.

## Root cause

`instantiate with <pflet>, ...` (no schema name before `with`) parses to
`*ast.WitnessTactic` in both Python (`ivy_parser.py:1528`) and Go
(`lalr_logicparser/grammar_v17.y:973-977`).

Python's `witness_tactic` (ivy_proof.py:461-473) calls
`lu.witness_ast(False, [], wit_map, conc)` directly on `conc = goal_conc(decl)`.
`witness_ast` is **duck-typed** (ivy_logic_utils.py:1673-1704): for a node
that is neither a quantifier, `Not`, nor `Implies`, it falls through to
`return fmla.clone([witness_ast(pos,vs,witnesses,arg) for arg in fmla.args])`
— recursing into whatever `.args` returns. A `TemporalModels` has
`args = [fmla]`, so Python naturally descends into the inner formula, applies
the witness, and clones the wrapper back around the result.

Go's `witnessTactic` (proof/tactics.go:598-609) is **type-gated**:

```go
rawConc := GoalConc(goal)
var newConc ast.Node
if concExpr, ok := rawConc.(lg.Expr); ok {
    witnessed, werr := module.WitnessAst(false, nil, witness, concExpr)
    ...
    newConc = witnessed
} else {
    newConc = rawConc     // <-- TemporalModels lands here: skip
}
```

`*ast.TemporalModels` is not an `lg.Expr`, so `ok == false`,
`module.WitnessAst` is never called, and the inner
`forall P,M,A,T. globally eventually memc.cpl_fair(P,M,A,T)` survives
untouched. The downstream `init_globally`/`initGlobally` walk then emits
`Variable` terms, producing the observed divergence.

## Audit: every Go site with the same guard pattern

Per feedback_port_match_everywhere.md, every mirror site in Go that has the
same silently-skip-on-non-`lg.Expr` pattern must be brought into conformance
with Python. Three sites found in `proof/tactics.go`, all of which Python
handles via duck-typed recursion and Go currently drops on
`TemporalModels`:

### 1. `witnessTactic` — proof/tactics.go:598-609

- Python: `conc = lu.witness_ast(False, [], wit_map, conc)` on
  `conc = goal_conc(decl)` (ivy_proof.py:471).
- Python handles TemporalModels via the generic `fmla.clone([... for arg in fmla.args])` branch.
- Go skips TemporalModels entirely. **This is the trigger for the failing test.**

### 2. `assumeTactic` — proof/tactics.go:174-188

- Python: `conc = lu.witness_ast(True, [], witness, conc)` on
  `conc = goal_conc(prem)` where `prem` is the schema (ivy_proof.py:371-374).
- Same duck-typed handling in Python; same silent-skip in Go. In practice
  `prem` is the schema body and is unlikely to be a `TemporalModels`, but
  the mechanical-port rule (CLAUDE.md §B) requires conformance here too —
  no "not exercised by the failing test" exceptions.

### 3. `unfoldTactic` — proof/tactics.go:340-346

- Python: `decl = goal_apply_to_conc(decl, lambda fmla: unfold_fmla(fmla, defns))`
  (ivy_proof.py:401). `unfold_fmla` calls `apply_match_alt` which
  (ivy_proof.py:1150-1168) recurses via `[apply_match_alt_rec(match,f,env) for f in fmla.args]`
  and clones — again, duck-typed across TemporalModels.
- Go:
  ```go
  goal = GoalApplyToConc(pc.astCfg(), goal, func(node ast.Node) ast.Node {
      if fmla, ok := node.(lg.Expr); ok {
          return UnfoldFmla(fmla, defns)
      }
      return node   // <-- silent skip on TemporalModels
  })
  ```
  Same pattern, same bug.

### Sites that already match Python (no change needed, listed for completeness)

- `SkolemizeFmla` (proof/skolem.go:100-120) — already takes `ast.Node` and
  explicitly unwraps `*ast.TemporalModels` in `rec`'s entry, mirroring
  Python ivy_proof.py:1443-1444.
- `ApplyTempcase` (tactics/ivy_tactics.go:335-345) — explicit
  TemporalModels unwrap/transform/rewrap, mirrors Python ivy_tactics.py:119-124.
- `ApplyToConc` / `GoalApplyToConc` (proof/goal.go:77-97) — helper that
  already unwraps TemporalModels for its fn argument.

### Sites deliberately out of scope

- `if_tactic` (proof/tactics.go:397-412) — Python `if_tactic`
  (ivy_proof.py:412-420) wraps the **entire** `decls[0].formula` in
  `Implies(cond, ...)` without descending into TemporalModels. Go's
  `ApplyToConc` fallback descends into TemporalModels and wraps only the
  inner formula — a pre-existing divergence but **not** the same
  silently-skip pattern this plan targets. If the user wants that fixed
  too, it needs a separate, distinct change (change semantics, not add
  TemporalModels handling). Flagging it here, not fixing it in this plan.

## Fix

For each of the three sites above, unwrap `*ast.TemporalModels` at the
call site, pass `tm.Fmla` (as `lg.Expr`) to the relevant function
(`module.WitnessAst` / `UnfoldFmla`), and rewrap via `tm.Clone(...)`.
This is the same structural pattern used by `ApplyTempcase`
(tactics/ivy_tactics.go:335-345) and matches Python's duck-typed
recursion at each site.

The signatures of `module.WitnessAst` and `UnfoldFmla` remain `lg.Expr`;
the unwrap responsibility belongs to the caller. This is consistent with
every other Go tactic that already handles TemporalModels this way
(`ApplyTempcase`, `ApplyToConc`). `SkolemizeFmla` took the alternative
approach of broadening to `ast.Node` internally — either approach is
equivalent; keeping signatures narrow here minimizes ripple in a focused
bug fix.

### Critical files / exact lines

- `/Users/jaten/ivy/goivy/proof/tactics.go`
  - `assumeTactic` body at lines 174-188 (conclusion-of-`prem` branch).
  - `unfoldTactic` inner function at lines 341-346.
  - `witnessTactic` body at lines 598-609 (conclusion-of-`goal` branch).

### Shape of each change (sketch, not final code)

Replace each occurrence of:

```go
if concExpr, ok := rawConc.(lg.Expr); ok {
    ... Fn(concExpr) ...
}
```

with:

```go
switch c := rawConc.(type) {
case *ast.TemporalModels:
    if innerExpr, ok := c.Fmla.(lg.Expr); ok {
        result, err := Fn(innerExpr)
        if err != nil { /* return ProofError */ }
        rawConc = c.Clone([]ast.Node{result})
    }
case lg.Expr:
    result, err := Fn(c)
    if err != nil { /* return ProofError */ }
    rawConc = result
}
```

For `unfoldTactic`, the idiomatic replacement is the same pattern inlined
inside the `GoalApplyToConc` callback — or, cleaner, switch to
`ApplyToConc` (proof/goal.go:77) since `UnfoldFmla` is `lg.Expr ->
lg.Expr` with no error channel:

```go
goal = CloneGoal(pc.astCfg(), goal, GoalPrems(goal),
    ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
        return UnfoldFmla(c, defns)
    }))
```

`ApplyToConc` already does the TemporalModels unwrap/rewrap.

### Reusable helper — worth considering

All three sites follow an identical shape. If preferred, extract a helper
in `proof/goal.go` beside `ApplyToConc`:

```go
// ApplyToConcErr is ApplyToConc for fns that may return error.
func ApplyToConcErr(conc ast.Node, fn func(lg.Expr) (lg.Expr, error)) (ast.Node, error) { ... }
```

Use it for the two `WitnessAst` sites (which return error). This is an
optional cleanup — not required for the fix. Decide in implementation.

## Verification

1. **Targeted golden run** (primary):
   ```
   cd ~/ivy/goivy && make test
   ```
   (per CLAUDE.md §9 — never `go test ./...`). Confirm
   `Test2hrOrdLive` advances past line 10956214. Post-fix, Go's
   `l2s_globally_0` canon should read:
   `... terms:[(Symbol name:_P ...) (Symbol name:_M ...) (Symbol name:_A ...) (Symbol name:_T ...)]`
   — matching Python.

2. **Inspect Go's `neg_prop_init` printout** in the next `log.golden.2hr`
   run. LHS of the top-level Implies should drop the
   `(forall A. (forall M. (forall P. (forall T. ...))))` wrapper and show
   `$l2s_init . □ ⬦ memc.cpl_fair(_P,_M,_A,_T)` — matching Python line 67047.

3. **Unit tests** — add/extend:
   - `proof/tactics_witness_test.go` (new): construct a goal whose
     conclusion is `*ast.TemporalModels` wrapping `forall X. F(X)`, run
     `witnessTactic` with match `X=_c`, assert resulting conclusion is
     `TemporalModels` wrapping `F(_c)`.
   - `proof/tactics_assume_test.go` (existing): add a case where the
     schema's conclusion is a `*ast.TemporalModels` to prove
     `assumeTactic` drives `WitnessAst` through the wrapper.
   - `proof/unfold_test.go` (existing): add a case where the goal
     conclusion is `*ast.TemporalModels` wrapping a formula that uses a
     defined symbol; assert unfold reaches the inner formula.

   These lock in the fix at package level without waiting for the 2hr
   golden regression.

4. **Cross-check other tactics** — one grep to confirm no remaining
   `rawConc.(lg.Expr); ok` pattern silently drops non-lg.Expr on a
   semantic operation:
   ```
   grep -nE '\.\(lg\.Expr\);\s*(!?ok)\b' proof/tactics.go
   ```
   After the fix, the three sites above should no longer use that
   pattern for the operation path (they should handle TemporalModels
   explicitly or via `ApplyToConc`/`ApplyToConcErr`).
