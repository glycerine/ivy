# Plan: fix `neg_prop_init` `forall P:proc.` divergence (instantiate on ord_live.ivy if_liveness)

Created: 2026-04-19 21:45 PT
Updated: 2026-04-19 (later) — previous l2s.SharedStep1 fix landed; now tackling downstream `neg_prop_init` divergence at xtrace 2411549.

## Context

`Test2hrOrdLive` (goivy/parser `make golden-2hr`) now passes the earlier
`SharedStep1 nAsms` checkpoint thanks to the NormalProgramClone +
AssumedInvs-propagation fix recorded below under "Previously applied fix".
The new run fails ~5,600 trace events downstream at:

```
2411549  go : XTRACE: l2s.modPass clone invar[35] ENTER HASH canon=...
           formula:(Not body:(Implies t1:(And terms:[
             (ForAll vars:[(Variable name:P sort:(UninterpretedSort name:proc))]
              body:(And terms:[(NamedBinder name:l2s_init ... ifabric.rd_fair(_P))
                               (NamedBinder name:l2s_init ... ifabric.wr_fair(_P))]))
             ...]) ...)

         py : XTRACE: l2s.modPass clone invar[35] ENTER HASH canon=...
           formula:(Not body:(Implies t1:(And terms:[
             (And terms:[(NamedBinder name:l2s_init ... ifabric.rd_fair(_P))
                         (NamedBinder name:l2s_init ... ifabric.wr_fair(_P))])
             ...]) ...)
```

Diff (go '-' vs py '+'):
```
-           (ForAll
+           (And
-             vars:[…]
+             terms:[…]
-             body:(And terms:[…
+               (NamedBinder …
```

Go retains a vacuous `(ForAll vars:[P:proc] body:(And [rd_fair(_P), wr_fair(_P)]))`
where Python has only `(And [rd_fair(_P), wr_fair(_P)])`. The formula is the
body of the instantiated axiom `ifabric_rw_fair_ax` used in the proof of
`cf_pio_live.if_liveness` (ord_live.ivy:2566–2592):

```
explicit temporal axiom [ifabric_rw_fair_ax]
forall P. (globally eventually ifabric.rd_fair(P)) & (globally eventually ifabric.wr_fair(P))

explicit temporal property [if_liveness] forall T. …
proof {
    tactic tempcase with P=ref.evs(T).p
    tactic skolemizenp
    instantiate ifabric_rw_fair_ax with P=_P   # ← root cause lives here
    tactic l2s_auto2 with …
}
```

## Root cause

Python `assume_tactic` (ivy_proof.py:350–382) separates the `pmatch` dict
into two buckets before applying it to the axiom schema:

```python
def iswit(x):                                                # x is the KEY
    return isinstance(x, il.Variable) and x not in prob.freesyms
witness = dict((x,y) for x,y in pmatch.items() if iswit(x))
pmatch  = dict((x,y) for x,y in pmatch.items() if not iswit(x))
```

- `witness` entries are consumed by `lu.witness_ast(True, [], witness, conc)`
  (ivy_logic_utils.py:1673–1698). For a `ForAll([P], body)` where every bound
  variable has a witness, `witness_ast` substitutes the body AND **drops the
  ForAll** (lines 1696–1698: `if new_vars: … ; return body`).
- `pmatch` entries are consumed by `apply_match_goal(pmatch, …, apply_match_alt)`
  which preserves quantifier structure.

For `instantiate ifabric_rw_fair_ax with P=_P`:

- `P` is a `lg.Var` bound by the axiom's outer `ForAll`. With `allow_witness=True`,
  `compile_match` augments its local `freesyms` with `used_variables(conc)`, so
  `fo_match` accepts `{P: _P}`. But `prob.freesyms` (the field on the problem)
  is **unchanged** — it still excludes P.
- `_P` is a `lg.Const` (a Skolem constant produced earlier by `skolemizenp`).
- Python: `iswit(P)` ⇒ `isinstance(P, Variable)` is True **and** `P not in prob.freesyms`
  is True ⇒ `{P: _P}` moves to `witness` ⇒ `witness_ast` strips the `ForAll`.

Go mirrors this split in `proof/tactics.go:146–158`:

```go
func isWitVar(key lg.NodeKey, val lg.Expr, prob *MatchProblem) bool {
    if _, isVar := val.(*lg.Variable); !isVar {           // ← checks VALUE
        return false
    }
    _, inFree := prob.FreeSyms[key]
    return !inFree
}
```

**Mechanical-port bug**: Go checks whether the `val` is a Variable; Python
checks whether the `x` (key, i.e. match LHS) is a Variable. When the RHS is a
skolemized `Const` like `_P`, Go's check returns `false`, so `{P_key: _P}`
remains in `pmatch`. `ApplyMatchGoalNode` → `applyMatchAltRec`'s ForAll branch
(proof/phase5_matching.go:1034–1055) then rebuilds a `ForAll([P], body)` with
the body substituted but the binder preserved (because the Variable slot is
filled by the original P when the substituted value is a non-Variable). The
vacuous `forall P:proc.` wrapping survives, and Python's log does not see it.

The loss of the `ForAll` matters because this invariant flows into
`l2s_auto2`'s invariant list as `neg_prop_init`. `l2s.modPass clone invar[35]`
is the first place the canon hashes diverge; downstream the invariant's
different shape produces a different Z3 encoding and the test fails fast.

## Fix

Single-point fix in `proof/tactics.go`: make `isWitVar` check the **key**'s
type, not the **value**'s. Mirror Python's semantics by reconstructing
"the key is a Variable" from data already on `prob`:

- Python uses `prob.freesyms` (the non-augmented set). A key is a bound
  variable from the schema conclusion iff it is in `used_variables(conc)`
  but not in `prob.freesyms`.
- Go has both available: `prob.FreeSyms` (unchanged, matches Python's
  `prob.freesyms`) and `prob.SchemaLF` (from which we can derive
  `used_variables(conc)` via `lu.UsedVariables`).

### Proposed Go code (tactics.go:226–234)

```go
// Mirror Python's iswit(x): x is the MATCH KEY, and we check
// whether that key's original LHS expression was a Variable AND not in
// prob.FreeSyms.
//
// In CompileMatchFull, when allow_witness=True the local freesyms is
// augmented with used_variables(conc) — that's exactly the set of Variable
// keys *added* beyond prob.FreeSyms. So:
//   key is a bound-Variable witness  iff  key ∈ used_variables(conc)  AND
//                                         key ∉ prob.FreeSyms.
func isWitVar(key lg.NodeKey, val lg.Expr, prob *MatchProblem) bool {
    if _, inFree := prob.FreeSyms[key]; inFree {
        return false
    }
    if prob.SchemaLF == nil {
        return false
    }
    conc := ConcAsExpr(GoalConc(prob.SchemaLF))
    if conc == nil {
        return false
    }
    if _, isUsedVar := lu.UsedVariables(conc)[key]; isUsedVar {
        return true
    }
    return false
}
```

The signature is unchanged; `val` parameter is retained (still used for
debug/trace but no longer drives the decision).

After the fix, the instantiate flow behaves exactly like Python:

1. `{P_key: _P}` matches `isWitVar` ⇒ moves to `witness`.
2. `WitnessAst(true, nil, witness, ForAll([P], body))` (module/skolem.go:215):
   - Iterates bound vars; `P` has a witness ⇒ `Substitute(body, {P: _P})` ⇒
     body with `_P` throughout.
   - `newVars` remains empty after loop.
   - `if len(newVars) > 0 { return CloneBinder(...) }`  (skolem.go:249–252)
     — this branch is skipped ⇒ returns just `body` ⇒ ForAll **dropped**.
3. `ApplyMatchGoalNode(pmatchClean, prem)` — pmatchClean has no entry for P,
   so no more substitution happens on this key.

`neg_prop_init`'s `invar[35]` content then matches Python exactly.

## Paired xtrace diagnostics (accrete — add now, never remove)

Per the project memory rule ("accrete xtraces, never delete"), add paired
Go/Python traces that will permanently guard this branch of the code. Place
them in the assume_tactic pipeline, both sides matching 1:1:

| Location | Trace label |
|---|---|
| Py `ivy_proof.py` after line 365 (witness/pmatch split) | `check.assumeTactic witness/pmatch split schema=%s nWitness=%d nPmatch=%d` |
| Go `proof/tactics.go` after line 158 | same format |
| Py `ivy_proof.py` after line 371 (conc after witness_ast) | `check.assumeTactic post-witness_ast concType=%s concCanon=%s` |
| Go `proof/tactics.go` after line 180 | same format |

These let future regressions surface at the split point instead of ~5,600
events downstream.

## Critical files

- `/Users/jaten/ivy/goivy/proof/tactics.go:226–234` — `isWitVar` (FIX target).
- `/Users/jaten/ivy/goivy/proof/tactics.go:146–158` — call site; unchanged.
- `/Users/jaten/ivy/goivy/proof/goal.go:53` — `ConcAsExpr` helper (reuse).
- `/Users/jaten/ivy/goivy/logicutil/logicutil.go:20` — `UsedVariables` (reuse).
- `/Users/jaten/ivy/goivy/module/skolem.go:215–252` — `WitnessAst` (already
  correctly drops ForAll when `newVars` is empty; verified — no change needed).
- `/Users/jaten/ivy/goivy/proof/phase5_matching.go:1034–1055` — `applyMatchAltRec`
  ForAll case (NOT touched; this is where the vacuous ForAll survived because
  of the bad split upstream).
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_proof.py:350–382` — reference
  `assume_tactic`.
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:1673–1698` — reference
  `witness_ast` (note lines 1696–1698: empty-newvars ⇒ return body).

## Verification

1. Unit: write a focused test in `proof/tactics_test.go` (or existing file)
   that builds a minimal `MatchProblem` with a bound-variable schema
   (`forall P. (rd_fair(P) & wr_fair(P))`) and a pmatch `{P_key: _P_const}`.
   Assert `isWitVar(P_key, _P_const, prob) == true` after the fix. This
   test fails today (returns false) and passes after.

2. Broader unit: `cd ~/ivy/goivy && make test` — ensure no regressions in
   `proof/`, `check/`, `module/`, `parser/` test suites.

3. Golden: `cd ~/ivy/goivy && make golden-2hr` (runs
   `Test2hrOrdLive` ~25min). Expected: passes the previous
   `invar[35]` checkpoint at i=2411549 and proceeds further. If the test
   surfaces yet another downstream divergence, document it in a follow-up
   plan; do NOT broaden this fix.

4. Trace sanity: diff the new Go/Py xtrace at the `check.assumeTactic`
   split trace to confirm `nWitness` and `nPmatch` agree for the
   `ifabric_rw_fair_ax` instantiation.

---

## Previously applied fix (earlier divergence — DONE, kept here for history)

`SharedStep1 ENTER nAsms=113 (go) vs nAsms=149 (py)` at xtrace 2405953.

### Root cause (confirmed)

Python `NormalProgram.clone` at `pyivy/ivy/ivy/ivy_temporal.py:179–183`
aliased `self.asms` (and other slice fields) by reference; Go's slice
value semantics couldn't track length mutations from inner
`check_isolate` calls.

### Changes landed

- `goivy/temporal/temporal.go:528–552` (NormalProgramClone): share slices
  (no `make+copy`), mirroring Python aliasing.
- `goivy/temporal/temporal_test.go:347–369` (TestNormalProgramClone):
  asserts slice-sharing after clone.
- `goivy/check/isolate_check.go:879–891`: write
  `mod.AssumedInvs = withLocalMod.AssumedInvs` after recursive
  `CheckIsolate` to propagate accumulated AssumedInvs back (Go-side
  stand-in for Python list aliasing).

### Diagnostic traces added (kept per accrete rule)

| Label | Py | Go |
|---|---|---|
| `check.CheckTemporals prop start` | ivy_check.py:140 | check/check.go:312 |
| `temporal.NormalProgramFromModule EXIT` | ivy_temporal.py:223 | temporal/temporal.go:390 |
| `check.CheckSubgoals TemporalModels applied` | ivy_check.py:798 | check/isolate_check.go:750 |
| `check.CheckIsolate preMove` | ivy_check.py:762 | check/isolate_check.go:674 |
| `l2s.l2sTacticInt postClone` | ivy_l2s.py:137 | check/l2s.go:329 |

All five paired traces agreed in the post-fix run.
