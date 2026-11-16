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

## Race gate and next direct-node caller work

The restored canonical core passes full IF race9.892s (handle21268 exit0,
/tmp/abcgo-dsd-canonical-full-if-race.log) and IF/GIA vet (handle87778 exit0,
empty /tmp/abcgo-dsd-canonical-vet.log). Native driver/canonical race passes26.720s (handle64469 exit0;
/tmp/abcgo-dsd-canonical-native-drivers-race.log).

Read-only review identifies a concrete next regression: Id_DsdManTuneStr1 reads
obj.truth, which is empty for the initial constant/generic variable and for
canonical internal structures never submitted as root Compute calls. It then
incorrectly marks these nodes unmatched. Source computes truth directly from
each canonical record. Add a new red test (including constant/variable and
internal nodes) before repairing this caller. Also replace full-truth nFans
minimization in tuning with original direct ComputeTruth semantics: nFans is
now structural arity, not complete support. Remove redundant object truth,
DSS network, permutation and support caches only after the direct canonical
balancing/tuning/report callers pass native and independent checks.

## Canonical tuning caller repair

New red tests prove cell tuning rejects the initial constant/variable and panics
on a nested canonical root when using obsolete complete-truth/arity caches
(handle6627 exit1, /tmp/abcgo-dsd-canonical-tuning-red.log). Compute truth directly
from each canonical node, as source does. A further new red test catches the
same obsolete-cache path in two-LUT tuning: internal ten-input AND never
submitted as a Compute root, then XOR the eleventh input (handle60479 exit1;
/tmp/abcgo-dsd-canonical-two-lut-tuning-red.log). Direct source reconstruction
repairs both caller paths without copying or min-base processing cached tables.

All three regressions pass0.087s (handle58058 exit0). Fresh original C matches
every canonical cell-tuning mark/configuration word and confirms the11-input
function fits two6-LUTs:2.122s (handle18760 exit0;
/tmp/abcgo-dsd-canonical-all-tuning-native.log). The test compares explicit stable
contract records, excluding only source progress/elapsed diagnostics. The first
cell-only comparison already passed2.078s (handle39722 exit0).

Full IF/native race evidence above predates these tuning changes; rerun the
appropriate affected checks before treating the new checkpoint as integrated.
Direct canonical balancing and removal of redundant object caches remain next.

## Redundant record caches removed; affected race gate passes

After both tuners consume direct canonical truth, whole-function truth,
support-index and input-permutation fields have no readers. Remove those three
slice fields and their allocations/copies, plus the unused support helper.
Shared canonical structure and manager-owned prime truth remain authoritative.
On amd64 unsafe.Sizeof reports144->72 bytes per DSD object, excluding backing
allocations; those three payload allocations are also eliminated. This is a
record-size measurement, not a whole-design memory/performance benchmark.
/tmp/abcgo-dsd-record-size-before.log and ...-after.log record the sizes.
Diagnostic text and DSS balancing network caches still remain; restore direct
canonical-node balancing before removing them.

Fresh affected verification after tuning and cache edits:
- Full map/if race10.089s (handle65076 exit0;
  /tmp/abcgo-dsd-canonical-compact-full-if-race.log).
- Native canonical/partition/cell/driver race29.161s (handle16159 exit0;
  /tmp/abcgo-dsd-canonical-compact-native-race.log), including510 output graphs,
  the2 explicit source partition assertions,252 canonical function sequences,
  1,008 partition checks and exact cell configuration contracts.
- IF/GIA vet exits0 with empty log (handle7560;
  /tmp/abcgo-dsd-canonical-compact-vet.log).

Remaining source-fidelity work includes direct canonical balancing, full verbose
DSD diagnostics, the source helper/caller inventory and persistence/merge/filter
shortcuts. The mapping-family frozen full-module race/vet gate has not yet run.
Do not treat these affected checks as the full GIA or131-package completion gate.

## Direct canonical balancing: new red contracts

Before replacing DSS balancing, a new nested-prime regression (E8{a(bc)d})
panics in pin-profile reconstruction: three prime fanins cover four actual
cut variables, while the old helper truncates profile operations to arity3.
The corrected behavioral log is /tmp/abcgo-dsd-nested-prime-profile-behavior-red.log
(handle11325 exit1). The first expanded test also read the AIG after its
source-shaped verifier cleared vector length; that harness mistake was fixed
before collecting the behavioral red evidence. The base AND3 path check passes.

Fresh original ifDsd.c confirms exact recursive emission order and source cost
policies across8 cases (handle25306 exit0,2.054s;
/tmp/abcgo-dsd-balancing-original-contracts.log). New regressions then fail:
construction-mode cost2 versus source0; composite AND/XOR/MUX emits children
before queue carries instead of interleaving source recursion and carries
(handle80919 exit1, /tmp/abcgo-dsd-graph-cost-order-red.log).

Next port original Eval_rec/EvalInt and PinDelays_rec directly over canonical
nodes, keep source graph-vs-timing cost policy, and separate prime SOP arity
from complete profile support. Remove redundant DSS/text caches only after
native/scalar regression comparisons pass. Preserve the original XOR delay/area
estimate (fXor=0), rather than replacing it with physical AND-gate depth.

## Direct canonical balancing restored; duplicate graph removed

Port Eval_rec/EvalInt and PinDelays_rec over canonical fanins, preserving
source interleaved emission/carries and separate graph/timing costs. Prime
SOP arity and complete cut-profile width are separate internal parameters;
this repairs nested-prime profile loss. All four new profile/cost/order checks
pass0.010s (handle55582 exit0). Exact fresh-C eight-case graphs, both costs,
full-support profiles and independent complete truth checks pass2.056s
(handle77594 exit0). The older graph-cost test now checks actual C policy:
nonzero abstract timing area and zero construction cost, with equal delay.

Remove the obsolete DSS balancing helpers and move their Boolean/signed-pin
and counter tests onto the actual canonical APIs. Remove cached text/network;
render canonical text on demand. Record size on amd64 is now48 bytes (previously
72, initially144), excluding fanins/payloads. No duplicate DSS graph remains.
/tmp/abcgo-dsd-record-size-direct.log records the size.
Full IF passes3.568s (handle55726 exit0). All510 native output graphs plus
canonical/partition/cell/balancing checks pass23.654s (handle89270 exit0;
/tmp/abcgo-dsd-direct-no-cache-native.log). Expanded prime/unsupported-SOP
coverage and affected race/vet remain next; full mapping-family gate pending.
