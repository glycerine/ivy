# Canonical DSD mapping dependency (in progress)

This follows the restored ifSat.c dependency inside the open giaIf.c audit.
It does not complete map/if, GIA or the original131-package objective (2/131
closed). Work remains sequential, without agents or branches. Preserve the
original algorithms; equivalent truth tables alone do not establish fidelity.

## Evidence and scope

Structured22/33 driver cases expose a second substitute: If_DsdManCompute
stores each truth independently with an identity permutation, and CheckXY
enumerates DAU partitions instead of traversing canonical DSD nodes. It ignores
derive/highEffort policy and does not reuse canonical objects. The source uses
one shared variable object, structural interning, signed/permuted fanins,
node-specific AND/XOR/MUX/prime checks and the restored SAT search as fallback.

The new minimal native regression (before production changes) proves:

| Contract | Go | Fresh original C |
| --- | --- | --- |
| Initial object count/file |4/empty |2/03.dsd |
| Three-input AND literal/count |8/5 |4/3 |
| Input permutation |0,2,4 |4,2,0 |
| Repeating the same function |10/6 |4/3 |

Both select partition5 in this minimal AND example. A real seed2 structured22
driver instead chooses different bound leaves, producing different mapped
graphs despite equal complete Boolean truth. Do not normalize these outputs
or change the expected answer to match the substitute.

Minimal behavioral red: /tmp/abcgo-mapping-dsd-canonical-behavior-red.log,
handle42103 exit1,2.004s. The initial fixture failed to link due optional
CUDD/thread exports in the existing ABC archive; matching those compile settings
fixed the fixture. That link failure is not counted as a production regression.

The first driver setup omitted the flags used by ABC's command path: DSD
requires fCutMin=1 and fExpRed=0 (src/base/abci/abc.c). Those omissions caused
source support/improvement assertions; they were harness errors, not Go bugs
or evidence that valid choice chains are unsupported. The corrected matrix
retains all64 seeds (including choices) across both structured modes.

With the corrected flags,97 cases match,29 have different mapped structures,
and2 original-C cases (seed25/56,mode6) abort in the SAT partition-dimension
assertion. Keep those exact source failures explicit; no graph can be compared
for them. Independent Go handling of those2 cases remains to be reviewed.
The test count increments only after a comparison actually passes, including
under -run filtering. Source flags/minimal boundary tests and the corrected
matrix: /tmp/abcgo-mapping-structured-source-flags-red.log, handle82991 exit1,
8.101s. The separate missing-cut-minimization boundary passes2.24s. Historical
setup logs are retained under /tmp/abcgo-mapping-structured-*.log but must not
be reported as behavioral Go regressions.

## Repair plan

1. [x] Add red native canonical-contract and real structured-driver comparisons;
   characterize original invalid-flag/partition failures independently.
2. [ ] Inventory all ifDsd.c routines and current public callers. Restore source
   manager/object invariants: two initial nodes, hash-consed signed structure,
   support/reference counts, prime truth storage and scratch/SAT ownership.
3. [ ] Port original compare/sort/create/find-or-add and recursive AddDsd logic.
   Preserve canonical permutations and literal complementation. Test repeated,
   permuted and phase-equivalent AND/XOR/MUX/prime structures against fresh C;
   retain independent full truth verification.
4. [ ] Restore Collect/Sign and AND/XOR/MUX/prime CheckXY traversal and ordering,
   masks and derive policy, then actual high-effort SAT fallback. Add new red
   boundary/decision tests before each additional discovered bug repair.
5. [ ] Reconcile balancing, cell/configuration, marks, interning and persistence
   callers with canonical records; remove semantic substitutes. Expand valid
   structured drivers without weakening structural/mapping/packing comparisons.
6. [ ] Review compactness/resource ownership separately, run full map/if race
   and native/scalar coverage, then resume remaining giaIf.c rows and its frozen
   full-module race/vet gate. Do not call the family or package complete early.

The source is2838 lines; the490-line current Go file includes several unrelated
shortcuts/stubs. This requires restoration of the source data model and traversal,
not a special case making the three-input regression green. Keep coherent red
checkpoints at least every ten minutes while the integration is incomplete.

Pre-restoration checkpoint re-run: the two exact original partition failures and invalid
flag boundary pass; the minimal canonical-contract test still fails with the
expected Go/C differences before repair. /tmp/abcgo-mapping-dsd-final-red-checkpoint.log,
handle36363 exit1,6.525s. Production DSD files are unchanged at this checkpoint.

## Canonical core restoration checkpoint

Restore original two-node allocation, signed structural hashing/comparison,
recursive DSD operation/parser with canonical input permutations, prime truth
interning, occurrence-based Collect/Sign, weighted AND/XOR/MUX/prime traversal
and actual SAT high-effort fallback. The minimal native contracts pass1.997s
(handle70086 exit0).252 fresh-C functions and1,008 ordered partition checks
cover nested operators/prime tables, repeated/complemented inputs, masks and
fallback; pass2.283s (handle4935 exit0). All510 actual native driver graphs now
match15.039s (handle96545 exit0), including all126 valid structured graphs;
the two source partition assertions remain explicitly characterized.

Further new regressions precede integration repairs:
- Pin profile uses obsolete identity instead of cut permutation: [2,2,1] versus
  constructed paths [1,2,2] (handle62136 exit1).
- AllocIsops wrongly changes decomposition LUT size2->3 and omits source warning
  (handles58016/76967 exit1; the latter captures actual diagnostic output).
- Free retains an actual SAT instance with48 variables; advertised16-input
  allocation overruns support-indexed caches (handle61788 exit1).

Use the actual cut permutation, source ISOP initialization, owned SAT destruction
and DAU16-sized caches. The larger Go cache bounds are an explicit safe repair
for its advertised16-input API, not a claim that C's undersized arrays are safe.
Two older tests asserted substitute behavior (identity permutation and masking
feasibility-only queries); reconcile them with freshly verified C contracts.
Full IF passes3.419s (handle72782 exit0); native canonical/ISOP tests pass4.171s
(handle41704 exit0). Exhaustive small balancing passes288 functions/2,368
assignments after the profile repair (handle44310 exit0,0.806s).

Current objects still retain transitional diagnostic/truth/DSS balancing caches.
Next restore direct canonical-node balancing/tuning callers and remove redundant
representations; review source persistence/merge/filter shortcuts when expanding
the required caller inventory. No DSD family, mapping family, GIA package or
131-package completion claim is made at this checkpoint.
