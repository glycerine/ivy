# Plan: Instrument all of `proof/` with XTRACE, diagnose + fix Var→Symbol divergence

Created: 2026-04-20 (UTC; exact clock time omitted per CLAUDE.md rule E)

## Context

`make golden-2hr` (~1 hour runtime) diverges at log.golden.2hr line 10956214
(file line 68130, trace `l2s.modPass clone invar[10] ENTER`). The invariant
`l2s_globally_0` for the `memc2cf_live` isolate (ord_live.ivy:2429) differs:

- Go:     `Apply(memc.cpl_fair, [Variable P, Variable M, Variable A, Variable T])`
- Python: `Apply(memc.cpl_fair, [Symbol _P, Symbol _M, Symbol _A, Symbol _T])`

The Go `neg_prop_init` (log line 67018) shows the LHS still wrapped in four
nested `forall A.forall M.forall P.forall T.` with Variables, and with
`$l2s_init P:proc,M:mem_type,A:addr_type,T:lclock.` carrying the same bound
vars applied to itself: `memc.cpl_fair(P,M,A,T)(P,M,A,T)`. Python has none
of this — just `$l2s_init . □ ⬦ memc.cpl_fair(_P,_M,_A,_T)`.

The proof uses `instantiate with P=_P, M=_M, A=_A, T=_T` which parses to a
`WitnessTactic` (not `AssumeTactic`). The LHS forall must be substituted away
by the WitnessTactic. In Go the substitution is failing.

**Since golden-2hr takes ~1 hour, we commit to one instrumentation pass that
covers the entire `proof/` package on both the Go and Python sides. After the
single instrumented run we diff the trace, identify the lookup miss, and fix
root cause.** Per `feedback_accrete_xtraces.md`, all traces added stay
permanently — this is conformance regression coverage we want regardless of
whether this specific fix lands quickly.

## Inventory

**Go `proof/` (11 non-test files, ~9268 LOC):**

```
checker.go              688  Proof dispatcher + all tactic routing
tactics.go              622  letTactic, assumeTactic, ifTactic, witnessTactic,
                              unfoldTactic, propertyTactic, functionTactic
goal.go                 747  Goal manipulation (GoalConc, GoalPrems, GoalFree,
                              GoalVocab, ApplyToConc, CloneGoal, GoalSubst,
                              GoalAddPrem, GoalRemovePrem, etc.)
skolem.go               333  SkolemizeGoal, SkolemizeFmla, varSubstGoal
match.go                675  Low-level pattern matching
matching.go             251  Setup helpers (setupSchemaMatching,
                              setupMatching, matchSchema)
phase5_matching.go     1484  Schema matching + compilation
                              (CompileMatchFull, CompileWitnessList,
                              SetupSchemaMatchingRaw, ApplyMatchGoalNode,
                              ApplyMatchAlt, etc.)
phase5_goals.go         920  Goal-level helpers (normalizeGoal, goalSubgoals,
                              applyMatchGoal, closeUnmatched, dropSuppliedPrems,
                              checkNameClash, removeExplicit, etc.)
errors.go                72  Error types
register.go              24  Tactic registration
proofstate.go           156  ProofChecker struct
```

**Go existing XTRACE (only 7 calls — vast gap):**

- `checker.go:222` proof.ApplyProof ENTER
- `checker.go:528,532` proof.composeProofs ENTER / step
- `checker.go:595` proof.tacticTactic name
- `tactics.go:155` proof.assumeTactic witnessSplit
- `tactics.go:182,184` proof.assumeTactic postWitnessAst

**Python `ivy_proof.py` existing XTRACE (6 calls — same 6 prefixes):**

- line 146 proof.ApplyProof ENTER
- line 193 proof.tacticTactic name
- line 200 proof.composeProofs ENTER
- line 203 proof.composeProofs step
- line 366 proof.assumeTactic witnessSplit
- line 373 proof.assumeTactic postWitnessAst

**Alignment status:** every existing Go↔Python prefix matches exactly. No
renames needed for the pre-existing traces — we only add new ones and must
keep the PascalCase convention (`proof.witnessTactic`, `proof.letTactic`,
not `proof.witness_tactic`).

## Naming convention (applies to every new trace)

Pattern: `proof.<GoFuncName> <EVENT> <kvPairs>` where

- `<GoFuncName>` is the Go PascalCase name (method or function) with
  lowercase first letter if unexported (matches existing convention —
  `proof.assumeTactic` not `proof.AssumeTactic`, since the Go method is
  unexported). For helpers in `goal.go` that ARE exported (e.g. `GoalConc`),
  use the exported form: `proof.GoalConc`.
- `<EVENT>` is one of: `ENTER`, `EXIT`, or a specific step tag
  (`witnessSplit`, `postWitnessAst`, `lookup`, `substMiss`, …). All uppercase
  for entry/exit, camelCase for step tags to match existing
  `witnessSplit`/`postWitnessAst`.
- `<kvPairs>` is `key=%v` format. Use `HASH canon=%v` for a structural
  canonical fingerprint of an AST node (using `.Canon()` on Go side, `.canon()`
  on Python side per section F of CLAUDE.md).

Python counterparts: every `xtracer.trace(...)` call in Python must use an
identical format string with the SAME PascalCase function name, so the log
diff pairs line by line. Python uses `%` formatting; Go uses `%v`/`%s` — the
runtime-produced string must be byte-identical between the two for every
entry/exit pair at the same program point.

Guard every Python trace with `if __debug__:` to match existing style
(ivy_proof.py line 146+).

## Instrumentation scope

### Go `proof/` — add entry+exit traces to:

**`checker.go`:**
- `ApplyProof` (already has ENTER, add EXIT around line ~230)
- `composeProofs` (already has ENTER+step, add EXIT)
- `tacticTactic` (already has one trace, add ENTER + EXIT wrapping)
- Every arm of the dispatch switch (~line 290): add `dispatch name=<TacticType>`
  trace so we can see which tactic was actually chosen for each proof node.
- `LookupSchema` (line ~500): ENTER schemaName=… / EXIT found=… HASH canon=…
- `InstSchema` (line 389): ENTER + EXIT with canon
- `ApplyTacticInt`, `ComposeProofs` internals

**`tactics.go`:** ENTER + EXIT HASH canon for each:
- `letTactic` (line 27)
- `assumeTactic` (line 92) — ENTER+EXIT bracketing the existing
  witnessSplit/postWitnessAst traces
- `ifTactic` (line ~384)
- `propertyTactic`
- `functionTactic`
- `unfoldTactic`
- `witnessTactic` (line 545) — **critical for this bug**:
  - ENTER: nWits, HASH canon(GoalConc), HASH canon(first arg if any)
  - Per-witness: `witness pair key=<lg.Key(lhs).String()> rhs=<rhs.Canon()>`
  - EXIT: HASH canon(newConc)
- `goalAddPrem`

**`goal.go`:** light-weight entry-only traces (these are called thousands of
times — use a counter-free trace so we get shape, not depth):
- `GoalConc`, `GoalPrems`, `GoalPremGoals`, `GoalConcUnwrap`, `GoalVocab`,
  `GoalFree`, `GoalFreeVars`, `GoalDefns`, `GoalIsTemporal`, `GoalIsProperty`
- `ApplyToConc` — ENTER+EXIT with canon since this is a structural mutator
- `CloneGoal`, `MakeGoal`, `GoalSubst` — ENTER+EXIT with canon
- `GoalAddPrem`, `GoalRemovePrem`, `GoalApplyToPrem`, `GoalApplyToConc`
- `WrapImplies` (added in prior plan) — ENTER+EXIT with canon
- Suppress noise: for pure getters (`GoalConc`, `GoalPrems`), use a single
  line trace; only functions that MUTATE structure get ENTER+EXIT with canon.

**`skolem.go`:**
- `SkolemizeGoal` (line 17) — ENTER prenex=%v HASH canon(goalConc) / per-free-var
  `substitute v=… sk=…` / EXIT HASH canon
- `SkolemizeFmla` (line 100) — ENTER + EXIT, and the four branches
  (TemporalModels, Not/Implies/And/Or, skolemize-pos, universalize-neg)
  each get a `branch type=<name> pos=%v` trace
- `varSubstGoal` — ENTER+EXIT with canon

**`matching.go` + `phase5_matching.go`:**
- `SetupSchemaMatching`, `SetupSchemaMatchingRaw`, `SetupMatching`,
  `MatchSchema` — ENTER+EXIT with canon(schema) canon(decl)
- `CompileMatchFull`, `CompileExprVocab`, `CompileWitnessList` — ENTER+EXIT
  with the input and output canon
- `ApplyMatchGoal`, `ApplyMatchGoalNode`, `ApplyMatchAlt`, `ApplyMatchAltRec`
  — ENTER+EXIT with canon; `ApplyMatchAltRec` also log each recursion-arm
  decision (quantifier substitution, Implies flip, etc.)
- `IsWitVar` (tactics.go:247) — log result per call so we can see the
  witness-vs-pmatch partitioning decisions

**`phase5_goals.go`:**
- `NormalizeGoal`, `GoalSubgoals`, `CloseUnmatched`, `DropSuppliedPrems`,
  `RemoveExplicit`, `RenamePremNoClash`, `CheckNameClash`, `GoalApplyToConc`,
  `GoalApplyToPrem`, `WitnessAst` (if defined here) — ENTER+EXIT with canon

**`match.go`:**
- Every exported function (28 total per inventory) — ENTER+EXIT. For the
  recursive core (`matchAst`, `unifyAst` — whatever the names are), add a
  `step type=<nodeType>` per call to trace the walk.

### Python `ivy_proof.py` + `ivy_logic_utils.py`:

For each Go trace above, add an identical-format trace in the Python
counterpart. Mirror sites:

- `ivy_proof.py` ProofChecker methods: `apply_proof` (146), `compose_proofs`
  (200), `tactic_tactic` (193), `assume_tactic` (366/373), `witness_tactic`
  (461), `let_tactic` (~215), `if_tactic` (~412), `unfold_tactic` (~386),
  `property_tactic`, `function_tactic`, `forget_tactic`, `show_goals_tactic`,
  `defer_goal_tactic`, `proof_tactic`, `setup_schema_matching`,
  `setup_matching`, `match_schema`, `lookup_schema`, `inst_schema`
- `ivy_proof.py` free functions: `goal_conc`, `goal_prems`, `goal_prem_goals`,
  `goal_vocab`, `goal_free`, `goal_free_vars`, `clone_goal`, `make_goal`,
  `goal_subst`, `goal_add_prem`, `goal_remove_prem`, `goal_apply_to_prem`,
  `goal_apply_to_conc`, `apply_to_conc`, `skolemize_goal` (1381),
  `skolemize_fmla` (1405), `var_subst_goal`, `normalize_goal`,
  `goal_subgoals`, `close_unmatched`, `drop_supplied_prems`,
  `remove_explicit`, `rename_prem_no_clash`, `check_name_clash`,
  `compile_match_full`, `compile_expr_vocab`, `compile_witness_list`,
  `apply_match_goal`, `apply_match_alt`, `apply_match_alt_rec`, `iswit` /
  `is_wit_var`
- `ivy_logic_utils.py` helpers called by `proof/`:
  - `witness_ast` (1673) — ENTER + per-quantifier branch trace with the
    `v` and `witnesses[v] found=…` outcome so the LHS-forall substitution
    miss can be pinpointed
  - `substitute_ast` (wherever it lives) — ENTER+EXIT canon
  - `used_variables_ast`, `used_variables_in_order_asts` — entry-only
  - `variables_ast` — entry-only with count

Python already has traces in several `ivy_logic_utils.py` helpers (e.g.,
`normalize_named_binders`, `replace_temporals_by_named_binder_g_ast`,
`reduce_named_binders`) — keep those, add matching Go traces in
`logicutil/logic_utils.go` IF the Go side doesn't already have them.
(Scope-limit: add missing Go traces only for Python helpers invoked from
`proof/` — no need to cover the entire `ivy_logic_utils.py`.)

### Deliberately out of scope (to keep the commit bounded):

- `check/`, `compiler/`, `fragment/`, `interp/`, `isolate/`, `module/`,
  `transrel/` — no new XTRACE in this pass. `module/skolem.go:223` (`WitnessAst`)
  is the one exception — it is the target of the diagnostic and needs
  per-iteration tracing.
- Tactic registration (`register.go`) — no traces.
- Pure-data types (`errors.go`, `proofstate.go`, `match.go` struct fields) —
  no traces on getters.
- Performance: XTRACE calls are guarded by `xtracer.enabled` in Go and
  `if __debug__` in Python — no impact on non-trace runs.

## Diagnostic step (what the added traces give us for the current bug)

With the full instrumentation in place, one `make golden-2hr` run yields the
trace we need. Grep the log for lines around the memc2cf_live proof:

1. `proof.witnessTactic ENTER` for the `instantiate with P=_P,…` step — confirm
   Go reaches the tactic.
2. `proof.witnessTactic witness pair key=… rhs=…` — four lines, one per
   witness. Compare the key strings against the LHS forall's variable keys.
3. `ilu.witnessAst ENTER type=…` and per-iteration
   `ilu.witnessAst quantifierVar v=P found=true/false` — the miss is visible
   as `found=false` on at least one var.
4. `proof.witnessTactic EXIT HASH canon=…` — the post-substitution conclusion;
   diff against Python's EXIT canon at the same line.

The diff between Go and Python at the exit point, plus the per-iteration
`found=…` trace in `witness_ast`, will pinpoint whether:

- **(H1) key mismatch**: Go's `lg.Key(v)` at the forall differs from the key
  we stored at witnessTactic time. Likely culprit: the forall's Variable has
  a different sort encoding (e.g., sort-alias vs canonical) than the pflet's
  Variable produced by `CompileExprVocab`. Fix in
  `proof/tactics.go:566-577` (canonicalize the key when building the map) or
  `proof/phase5_matching.go:1469-1479` (canonicalize when compiling).
- **(H2) skolemizenp drift**: RHS symbols appear only on RHS, LHS forall
  retains Variables — this is what we see, which agrees with the WitnessTactic
  failing (not skolemizenp). Rule out by confirming `SkolemizeGoal EXIT` canon
  matches Python.
- **(H3) dispatch miss**: No `proof.witnessTactic ENTER` fires on Go side.
  Fix in `proof/checker.go:290` switch.

## Implementation order (single commit)

1. **Go proof/ traces** — 11 files, ~60 entry/exit pairs + step tags. Follow
   the naming rules above. Each file gets its imports of `xtracer` added if
   not already present.
2. **Go `module/skolem.go` `WitnessAst`** — add the quantifier-arm per-var
   trace (`ilu.witnessAst quantifierVar …`) and ENTER/EXIT at function
   boundaries.
3. **Go `logicutil/logic_utils.go`** — only add the small number of traces
   that mirror Python's existing `ivy_logic_utils.py` traces that proof/
   relies on (scan first, then fill gaps).
4. **Python `ivy_proof.py`** — add matching `xtracer.trace(...)` with
   identical format strings. Guard each with `if __debug__:`.
5. **Python `ivy_logic_utils.py`** — mirror the small Go additions from (3)
   and the `witness_ast` per-iteration trace.
6. **Run** `cd ~/ivy/goivy && make golden-2hr`. Look at the tail of the
   resulting `log.golden.2hr` around file line 68130 and above (the
   instrumented witnessTactic/witness_ast region for the memc2cf_live proof).
7. **Diagnose + fix** based on the trace (Phase B below). Root-cause fix
   goes in whichever file the diagnostic indicts.
8. **Re-run** `make golden-2hr` — must complete with zero divergence.
9. **Run `make test`** — all other tests stay green (per
   `feedback_use_make_test.md`, never `go test ./...`).

## Phase B: The fix (to be finalized after step 7)

Until the trace lands, the fix shape is one of:

- **B-key**: normalize witness map keys in `proof/tactics.go:566-577` so Go's
  lg.Key encoding matches at both the map-build site and the quantifier
  lookup site. Likely Variable sort-encoding normalization.
- **B-recursion**: fix `module/skolem.go:223-322` if `WitnessAst` is miss
  a branch (e.g. doesn't recurse into a wrapper, or skips the Implies on some
  shape).
- **B-dispatch**: fix `proof/checker.go:290` switch if WitnessTactic is
  shadowed by another arm.

Per CLAUDE.md rule B9 "correct in practice" / "good enough for now" is not
acceptable — Python-Go behavioral parity must be exact on this proof path.
Per `feedback_port_match_everywhere.md`, if the fix is a key normalization,
also audit `assumeTactic` (`tactics.go:92`) and `SkolemizeGoal`
(`proof/skolem.go:17`) and apply the analogous fix.

## Files touched (summary)

### Instrumentation pass (step 1-5):

Go:
- `proof/checker.go`
- `proof/tactics.go`
- `proof/goal.go`
- `proof/skolem.go`
- `proof/match.go`
- `proof/matching.go`
- `proof/phase5_matching.go`
- `proof/phase5_goals.go`
- `module/skolem.go` (WitnessAst quantifier-arm traces)
- `logicutil/logic_utils.go` (gap-filling, if needed)

Python:
- `ivy_proof.py`
- `ivy_logic_utils.py` (witness_ast, gap-filling)

### Fix pass (step 7):

One or two of:
- `proof/tactics.go` (witnessTactic key-building)
- `proof/phase5_matching.go` (CompileWitnessList)
- `module/skolem.go` (WitnessAst)
- `proof/checker.go` (dispatch)

## Verification

1. `cd ~/ivy/goivy && make golden-2hr` — must pass with zero divergence.
2. `cd ~/ivy/goivy && make test` — all tests stay green.
3. Snapshot `log.golden.2hr` before/after — expect the file to grow
   significantly in size from the new traces; this is fine and expected.
4. Confirm the new traces fire on matching memc2cf_live `instantiate with
   P=_P,…` path (`grep "proof.witnessTactic" log.golden.2hr | head -20` on
   both Go and Python sides should show paired traces).
5. Unit test: a small .ivy fixture in `test_vectors/` (per
   `feedback_test_vectors_dir.md`) that reproduces the LHS-forall witness
   substitution in miniature — `(forall X. P(X)) -> Q`, proof uses
   `tactic skolemizenp; instantiate with X=_X`, assert the resulting
   `l2s_globally_0` invariant has `P(_X)` Symbol form.

## Out of scope / do NOT touch

- `ApplyToConc` / `SkolemizeFmla` body — those are correct per prior plan.
- Parser grammar rules — `INSTANTIATE WITH pflets → WitnessTactic` is
  correct on both sides.
- `ReduceNamedBinders` — Go has the helper but it's unused and not needed
  for this flow.
- `initGlobally` in `check/l2s_auto.go` — it faithfully processes its input;
  the bug is upstream at the WitnessTactic stage.
- CLAUDE.md section D — do NOT use git. User commits in the background.
