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

Final checkpoint re-run: the two exact original partition failures and invalid
flag boundary pass; the minimal canonical-contract test still fails with the
expected Go/C differences before repair. /tmp/abcgo-mapping-dsd-final-red-checkpoint.log,
handle36363 exit1,6.525s. Production DSD files are unchanged at this checkpoint.
