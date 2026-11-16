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
3. [x] Port original compare/sort/create/find-or-add and recursive AddDsd logic.
   Preserve canonical permutations and literal complementation. Test repeated,
   permuted and phase-equivalent AND/XOR/MUX/prime structures against fresh C;
   retain independent full truth verification.
4. [x] Restore core Collect/Sign and AND/XOR/MUX/prime CheckXY traversal and ordering,
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

## Expanded direct balancing and metadata utilities

Expanded fresh-C matrix passes522 balancing decisions:270 complete graphs,
252 unsupported cuts,9 retain partial construction. Compare exact partial/full
AIGs, timing/construction costs and full-support profiles; verify complete
Boolean truth independently for supported graphs. Canonical coverage expands
to261 functions/1,044 ordered partition checks with nested unsupported primes.
/tmp/abcgo-dsd-direct-expanded-native-first.log, handle78942 exit0,4.341s.

New red manager-utility tests precede further repairs: Filter produces7 instead
of4 nodes (uses support instead of occurrence threshold and duplicates bases);
Merge produces7 instead of5 nodes (clones instead of interning/remapping);
Clean/Invert retain cell descriptions, and Clean retains configurations
(handle77533 exit1, /tmp/abcgo-dsd-manager-metadata-red.log).
Port original compatibility checks, recursive dependency remapping, structural
interning, source occurrence/mark/configuration policies and mark reset state.
All new tests pass0.009s (handle43525 exit0). Fresh original C confirms filtered/
merged node counts and full truths, mark values and configuration ownership
(handle71063 exit0,2.071s; /tmp/abcgo-dsd-manager-metadata-native.log).
The native comparison excludes only progress/elapsed text before explicit
metadata records; it does not normalize any compared state.

Affected full IF race10.191s, native/510-driver race35.548s and IF/GIA vet
all exit0 (handles79883/9081/80733). Logs:
/tmp/abcgo-dsd-direct-final-if-race.log,
/tmp/abcgo-dsd-direct-final-native-race.log,
/tmp/abcgo-dsd-direct-final-vet.log (empty). Source binary persistence, full diagnostic/report
behavior and broader utility boundary/configuration matrices remain open;
this is not DSD/mapping/GIA/package completion.

## Pre-persistence checkpoint and planned scope

Direct canonical balancing and utility repairs are committed and pass affected
race/vet above. Graph costs and ordering retain original C behavior; no cached
DSS representation remains. The complete original131-package scope stays active
with2/131 closed. This is a required dependency checkpoint inside the still-open
GIA mapping family, not a full map/if audit.

Next use new red regressions to replace remaining persistence substitutes:
If_DsdManSave currently calls the short text printer, while original C writes
its dsd1 binary format; If_DsdManLoad still returns nil. Verify both directions
against freshly compiled original Save/Load, covering canonical nodes, prime
truth/decomposition caches, counts/marks and optional cell configurations. Review
full source diagnostics and merge/filter compatibility/configuration boundaries
separately, then resume all remaining giaIf.c rows and its frozen full-module
gate. Do not infer completion from the narrower green checks recorded here.

## Persistence regression checkpoint

New tests fail before production repair: Save emits text instead of dsd1,
Load returns nil for an independent original-layout AND3 record (including
nonzero C padding), prime/configuration round trips fail, and default store
03.dsd is never written. /tmp/abcgo-dsd-persistence-core-red.log,
handle7991 exit1,0.023s. Save this red checkpoint before codec work.

The original format stores native packed node records (base nodes0/1 omitted),
prime truth tables and decomposition lists, then optional configuration words
and cell text. Interoperability checks must compare semantic records and both
cross-language read directions; compare reserialized Go files byte-for-byte
through C, while ignoring uninitialized padding in initial C allocations.
Implement the established little-endian GCC/Clang dsd1 layout explicitly,
without relying on Go struct memory layout. Validate truncated/inconsistent
input and preserve source state without reproducing unchecked C reads. Keep
zero-LUT managers without a SAT instance (source's unconditional BuildXY(0)
in Load has undefined negative-shift behavior); observable serialized state
must still match. Source-defined LUT2..6 cases supply the strict native gate.

## Original binary persistence restored

Replace Save's text output and Load's nil stub with explicit dsd1 packed-node,
prime truth/decomposition and configuration/cell encoding. Rebuild canonical
hash tables and retain serialized counts/marks. Zero-LUT loads own no SAT
instance; positive LUT loads construct the actual source CNF. Validate lengths,
node headers/topology/support, truth references and configuration dimensions;
release partially loaded managers on error. Nonzero original C node padding is
ignored while deterministic Go output uses zero padding. Repeated Go round
trips are byte-stable. Default Store follows source paths; Load sets its actual
filename. Add original PrintOne formatting to compare all persisted node state.

Core regression repair passes0.026s (handle88869 exit0). All prefixes of an
independent68-byte source record plus corrupt dimension/header/topology/cache/
configuration cases reject without panic; extended persistence suite0.075s
(handle35169 exit0). Full IF passes3.605s (handle92694 exit0).

Fresh unmodified C writes three managers (decomposition, configured cells,
zero-LUT). Go loads and reports every node/count/mark/structure and config word;
C reads Go output and reserializes it byte-for-byte. Loaded primes retain
interning. All three cases pass2.136s (handle81164 exit0;
/tmp/abcgo-dsd-persistence-native-first.log). This verifies the established
little-endian ABI on the native test host, including its observed zero-LUT
load behavior; Go avoids depending on that C negative-shift implementation.
No equality claim is made for uninitialized padding in initially C-created
objects. Corresponding core/boundary/full logs are under
/tmp/abcgo-dsd-persistence-{core-first-repaired,boundaries,full-if-first}.log.

Affected race/vet, compactness review and broader persistence/cache/error
matrices remain next. Full source reports and remaining utility/configuration
boundaries and giaIf.c integration are still open; no family/package closure.

## Persistence compactness, boundaries and affected integration

Read scalar/record slices directly from the buffered file instead of allocating
and copying a temporary buffer for every field. Retained node fans, prime/cache
words, configuration data and cell text have their own typed ownership. A test
overwrites all borrowed input bytes after decoding and checks reserialization
is unchanged. Reserved base-node counts/marks are intentionally omitted, as C
specifies. Focused ownership/boundary suite passes0.104s and native interop2.143s
(handles15160/61317 exit0).

Frozen affected gate after codec/compactness edits:
- Full IF race9.456s (handle37301 exit0).
- Full focused native/510-driver/522-balancing race37.961s
  (handle86148 exit0), including persistence.
- Direct caller base/abci race28.296s (handle35504 exit0).
- IF/GIA vet exits0, empty /tmp/abcgo-dsd-persistence-if-gia-vet.log.

The broader vet invocation also included base/abci and exits1 (handle69566):
24 existing findings (14 unkeyed timing literals,9 unreachable statements,
1 self-assignment). A read-only Go overlay restores pre-persistence IF code
and masks the new codec; base/abci vet still yields the same24 diagnostic
multiset (handle61370 exit1). ABCI source is unchanged. Log/overlay:
/tmp/abcgo-dsd-persistence-abci-baseline-vet.log and
/tmp/abcgo-dsd-persistence-vet-baseline/overlay.json. These findings remain
queued for ABCI's sequential audit, not silently reported as a clean broad vet.

Further test-only expansion verifies all supported LUT sizes0,2..6, cell state,
and zero/one-variable managers:9 C->Go->C cases pass race3.374s
(handle47575 exit0). Source C's zero-LUT behavior remains host-specific; Go
has no negative shift. A Go16-input AND round trip checks the packed fanin
count's unsigned sign bit and full65,536-assignment truth. Expanded focused
persistence race passes1.172s (handle91033 exit0). Logs:
/tmp/abcgo-dsd-persistence-all-dimensions-native-race.log,
/tmp/abcgo-dsd-persistence-all-boundaries-race.log.
No production changes followed the frozen affected gate; only these targeted
test/fixture expansions were added and run under race.

Core canonical/parser/partition, direct balancing, metadata and binary persistence
now have source/scalar evidence. Full DSD diagnostics/verbose behavior, broader
utility compatibility/configuration matrices and the remaining giaIf.c rows
are still open, followed by its full-module frozen gate. Original audit2/131;
no DSD/mapping/GIA/package/goal completion claim.

### Report options closed with Go-only regressions

The former Print stub ignored statistics, occurrence histograms, sampling and
verbosity. The report implementation now follows the source distribution and
occurrence algorithms, count-based ordering and support/sampling filters. It
preserves the source's stdout/file routing (including rank prefixes on stdout).
Storage rows estimate Go representations instead of pretending to measure C
allocations. Dead source counters were omitted; source time reporting remains.

Removed the newly added if_dsd_reports_review.c and its runtime compiler test.
Its replacement lives in map/if/dsd_reports_review_test.go: 32 option/routing
combinations with hand-specified logical rows and records, plus prime-descendant
classification, occurrence decade boundaries and file-open failure. No C source,
compiler, ABC object archive or native executable is needed by these tests.
A read-only overlay with the old Print body reproduces the failure using these
Go-only tests (exit1, /tmp/abcgo-dsd-reports-go-before.log); repaired tests pass.

Affected checks: full IF race10.334s, IF/GIA vet exit0, full IF with CGO_ENABLED=0
and PATH=/usr/local/bin (no cc/gcc/ar) passes1.221s. Logs:
/tmp/abcgo-dsd-reports-if-race.log,
/tmp/abcgo-dsd-reports-vet.log,
/tmp/abcgo-dsd-reports-if-go-only.log.
Older optional compiler-dependent comparisons elsewhere are not migrated by
this change. Going forward, ordinary regression coverage must run with Go alone;
retain source-derived expected data rather than add executable .c test vectors.

This is a passing report-options chunk, not a mapping-family completion claim.
Truth-table dumping and verbose decomposition diagnostics remain open, along
with the earlier utility/driver rows and full mapping integration gate.

### Verbose partition diagnostics

Go-only regressions reproduce missing IDs/counts, candidate reports and accepted
partition explanations before repair (/tmp/abcgo-dsd-verbose-red.log, exit1).
Restore source PrintOne prefix, subtree reporting before mask rejection,
AND/XOR/MUX accepted-mask explanation, prime set-family listing and selected-set
formatting. Partition traversal, decision ordering and masks are unchanged.
Tests explicitly check complemented root reporting, derive=0, blocked candidates,
quiet output, shared-control MUX helper and weighted prime-family boundaries.
Two initial MUX/prime test expectations were corrected using source support-limit
arithmetic before production edits; they were not algorithm bugs. No C fixtures
or compilers are used by the new tests. Repaired focused tests pass0.011s;
full IF race and IF/GIA vet pass (logs /tmp/abcgo-dsd-verbose-if-race.log and
/tmp/abcgo-dsd-verbose-vet.log). Dump APIs and wider mapping audit remain open.
