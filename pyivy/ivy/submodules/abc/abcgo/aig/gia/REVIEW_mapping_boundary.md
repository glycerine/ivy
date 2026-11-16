# Mapping boundary review (in progress)

Scope: all 55 main routines in `src/aig/gia/giaIf.c`, plus two inline
fanin-copy adapters. Implementations live in both aig/gia and map/if to avoid
an import cycle. This is one sequential GIA source-family review; reviewing
these boundary functions does not mark the entire map/if package audited.
Glitch simulation closed at89caa1ccf. Original package completion remains2/131.

## Plan

1. [x] Inventory all source routines, actual Go locations and earlier evidence.
2. [ ] Review metrics, references, overlaps and choice levels; add failing
   regressions before fixes, then native/scalar/error/ownership coverage.
3. [ ] Review exact reports, persistent dump state, packing/profile boundaries,
   source precision and resource handling with native comparisons.
4. [x] Review mini reconstruction and parameter/GIA-to-IF boundary, choice
   integrity and frame timing metadata; preserve original algorithms.
5. [ ] Review all IF-to-GIA LUT/cofactor/config/cell reconstruction and drivers,
   with actual C and independent Boolean/metadata/ownership tests.
6. [ ] Reconcile previously reviewed hashing/transfers, fill remaining gaps,
   measure compactness, freeze tested files and run uncached full module race
   plus relevant vet. Do not claim source-family closure before all rows pass.

## Source inventory

Pending means not yet reconciled in this family, even if old unit tests pass.
Earlier reviewed rows retain prior native/scalar evidence in SEQUENTIAL_AUDIT.md.

| Original C | Go location | Status |
| --- | --- | --- |
| Gia_ManSetIfParsDefault | map/if/gia_bridge.go | Source defaults/full reset reviewed; Go-only regression; final family gate pending |
| Gia_ManLutFaninCount | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManLutSizeMax | aig/gia/mapping_accessors.go, base/abci/abc_graph_compat.go | Core coverage passes; integration/boundaries pending |
| Gia_ManLutNum | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManLutLevel | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManLutParams | aig/gia/mapping_timing.go | Ordinary/boxed native and independent coverage; integration pending |
| Gia_ManSetRefsMapped | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManSetLutRefs | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManComputeOverlap2One_rec | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManComputeOverlap2One | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManComputeOverlap2 | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManComputeOverlapOne_rec | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManComputeOverlapOne | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManComputeOverlap | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManPrintGetMuxFanins | aig/gia/mapping_stats.go | Pending |
| Gia_ManCountDupLut | aig/gia/mapping_stats.go | Pending |
| Gia_ManPrintMappingStats | aig/gia/mapping_stats.go | Pending |
| Gia_ManPrintPackingStats | aig/gia/mapping_profile.go | Pending |
| Gia_ManPrintNodeProfile | aig/gia/mapping_profile.go | Pending |
| Gia_ManPrintLutStats | aig/gia/mapping_profile.go | Pending |
| Gia_ManChoiceLevel_rec | aig/gia/mapping_choice_level.go | Ordinary/typed/boxed coverage; integration pending |
| Gia_ManChoiceLevel | aig/gia/mapping_choice_level.go | Ordinary/typed/boxed coverage; integration pending |
| If_ManCheckChoices_rec | map/if/choice_check.go (ifDanglingChoices closure) | Source walk reviewed; dangling/64 choice-chain cases; final family gate pending |
| If_ManCheckChoices | map/if/choice_check.go | Source diagnostics/mark cleanup reviewed; repeated Go-only checks; final family gate pending |
| Gia_ManToIf | map/if/gia_bridge.go | Source transfer reviewed; name ownership, defaults, typed phases/references/aliases, constants, empty interfaces and 64 choice chains; frame/boxed integration gate pending |
| Gia_ManBuildFromMiniInt | aig/gia/mapping_mini.go | Source algorithm/scratch reviewed; 140 Go-only scalar cases plus invalid inputs; final family gate pending |
| Gia_ManBuildFromMini | map/if/gia_mini.go | Source wrapper reviewed; balancing/large parity checks; final family gate pending |
| Gia_ManFromIfAig_rec | map/if/gia_from_if.go | Source recursive hash/copy algorithm reviewed; 64 Go truth/memoization cases; final family gate pending |
| Gia_ManFromIfAig | map/if/gia_from_if.go | Source reconstruction reviewed; callback/repeat/balancing tests; lazy small scratch; final family gate pending |
| Gia_ManFromIfLogicCreateLut | map/if/gia_lut.go | Source logic/records reviewed; signed multiword tests through15 inputs; final family gate pending |
| Gia_ManFromIfLogicCreateLutSpecial | map/if/gia_lut.go | Source encoding reviewed; ordered ifDec07 dependency restored; 192 Go-only signed/permuted reconstruction cases; final family gate pending |
| Gia_ManFromIfLogicNode | map/if/gia_lut_structure.go | Pending |
| Gia_ManNodeIfToGia_rec | map/if/gia_cut_cone.go (recursive closure) | Source choice traversal/phase handling reviewed; 64 equivalent-choice cases; final family gate pending |
| Gia_ManNodeIfToGia | map/if/gia_cut_cone.go | Source cone boundary reviewed; scratch/failure/constant-leaf tests; final family gate pending |
| Gia_ManFromIfLogicFindLut | map/if/gia_lut.go | Source remap/SAT/packing reviewed; 400 size/phase/shared-partition cases; final family gate pending |
| Gia_ManFromIfGetConfig | map/if/gia_config.go | Pending |
| Gia_ManFromIfLogicFindCell | map/if/gia_cell.go | Pending |
| Gia_ManFromIfLogicCofVars | map/if/gia_cofactor.go | Pending |
| Gia_ManFromIfLogicAndVars | map/if/gia_cofactor.go | Pending |
| Gia_ManFromIfLogic | map/if/gia_mapping.go | Pending |
| Gia_ManMappingVerify_rec | aig/gia/mapping_transfer.go | Pending |
| Gia_ManMappingVerify | aig/gia/mapping_transfer.go | Pending |
| Gia_ManTransferMapping | aig/gia/mapping_transfer.go | Earlier native/scalar review; reconcile |
| Gia_ManTransferPacking | aig/gia/mapping_transfer.go | Earlier native/scalar review; reconcile |
| Gia_ManTransferTiming | aig/gia/timing_metadata.go, base/abci/abc_graph_compat.go, base/wlc/wlc_compat.go | 6,144 native ownership cases; source self-order retained |
| Abc_FrameMiniAigSetCiArrivals | aig/gia/frame_bridge.go | Source semantics/ownership reviewed; Go-only boundary matrix; final family gate pending |
| Abc_FrameMiniAigSetCoRequireds | aig/gia/frame_bridge.go | Source semantics/ownership reviewed; Go-only boundary matrix; final family gate pending |
| Abc_FrameMiniAigReadCoArrivals | aig/gia/frame_bridge.go | Source semantics/ownership reviewed; Go-only boundary matrix; final family gate pending |
| Abc_FrameMiniAigSetAndGateDelay | aig/gia/frame_bridge.go | Source semantics/ownership reviewed; Go-only boundary matrix; final family gate pending |
| Gia_ManPerformMappingInt | map/if/gia_mapping_driver.go | Pending |
| Gia_ManPerformMapping | map/if/gia_mapping_driver.go | Pending |
| Gia_ManPerformSopBalance | aig/gia/balance_engines.go, map/if/gia_mapping_driver.go | Metadata policy repaired; native/core coverage; wider cases pending |
| Gia_ManPerformDsdBalance | aig/gia/balance_engines.go, map/if/gia_mapping_driver.go | Metadata policy repaired; native/core coverage; wider cases pending |
| Gia_ManTestStruct | map/if/gia_structure_report.go | Pending |
| Gia_ManDupHashMapping | aig/gia/mapping_hash.go | Last usable cut policy and C collapse failures reviewed; integration pending |

The inline If_ManFanin0Copy/If_ManFanin1Copy logic is represented by the
phase-preserving copyEdge closure in map/if/gia_bridge.go.

## Initial findings to test

Choice levels and both overlap walkers appear to inspect only the two stored
fanins of real MUXes, omitting control cones. Their original C paths assume
ordinary AND graphs. Preserve that algorithm and order for ordinary graphs,
while covering the already-supported typed representation correctly. Add new
minimal red regressions before changing production code.

Native dump formatting/precision and empty packing reports also need exact
comparison. Existing source-shaped mini nonhashed construction must not be
silently replaced by globally hashed synthesis. Retain source quirks and label
safe Go extensions explicitly; do not invent a different mapper or solver.

## First correctness checkpoint

Two new typed-MUX regressions fail before repair: choice depth1 instead of4;
both overlap walkers count1 cone node instead of4 and report negative overlap
for complete coverage (`/tmp/abcgo-mapping-dependency-red.log`,handle63225 exit1).
Use the existing allocation-free logical-fanin helper in those walks. Ordinary
AND ordering and source traversal/sibling algorithms are retained; real MUX
control cones now participate. The repaired minimal/existing suite passes0.005s
(`/tmp/abcgo-mapping-dependency-repaired.log`,handle15214 exit0).

An unchanged-source native fixture compares96 ordinary graphs across32 seeds
and three cut styles, including duplicate equivalent nodes/choices, signs,
levels, parameters, references, cone postorders, overlap and reports. All96
match exactly after the logical-fanin change. New independent report cases
fail before repair: source float32 average gives1.07 where Go float64 gives1.06;
valid empty packing printsN=-1 while Go printsN=0
(`/tmp/abcgo-mapping-native-red.log`,handle59387 exit1). Restore intermediate
source precision and initial packing maximum. The fixture compiles/calls the
original giaIf implementation; no source algorithm is replaced or patched.

The first harness build had a test-only vector/slice type mismatch and was
corrected before collecting the behavioral red evidence above. Missing-mapping
validation already rejects invalid overlap calls in Go; its new check passed
without a production change, so this is coverage, not a claimed new bug fix.

Further typed dependencies, exact dump/profile/packing/mux reports, timing,
independent metrics/ownership and the remaining source inventory are pending.
This checkpoint is not source-family closure or a completed GIA package audit.

## Expanded core coverage checkpoint

After correcting an overestimated depth in the new typed test (its data
branches are one-level XOR/buffer nodes),64 signed typed configurations pass
all16 input assignments. Ordinary source AND/buffer walk order is unchanged;
real MUX controls add their actual dependencies. Both native report regressions
now pass. Complete repaired core coverage passes12.839s without skips
(`/tmp/abcgo-mapping-core-complete-focused.log`,handle56993 exit0). The full
map/if caller race suite passes7.972s
(`/tmp/abcgo-mapping-if-caller-race.log`,handle24088 exit0).

All new native fixtures now directly include and freshly compile the current
original giaIf.c, rather than relying on its prebuilt object for the reviewed
routines. They leave its algorithms and assertions unchanged. Coverage includes:

-96 exact ordinary graph/metric/reference/cone-postorder/choice/report states.
-384 independent stack/set cone, mapped-level and reference configurations,
 including registers, replaced reference storage and independently owned levels.
-8 signed mux graphs exercising three- and two-leaf mux records, data-fanin
 duplicate reports, all packing group sizes and complete mark cleanup.
-66 node profiles plus exact persistent dump/name transitions and empty LUT
 reports. Only elapsed-time numeric text is normalized in dump comparison;
 other formatting, fields and stdout match exactly. These checks found no
 further report defect beyond the two already repaired.
-140 exact mini reconstructions: source hashed/nonhashed choices, constants,
 variables, signed/constant/repeated leaves, prior-node literal remapping,
 output phase, scratch retention/clearing, graph objects/literals and independent
 complete Boolean truth tables. Invalid mini inputs are rejected. This algorithm
 needed no production change; the source's nonhashed append policy is retained.

Review of the new walk changes confirms ordinary AND ordering, buffer visitation,
postorder recording, sibling maximum, cut stopping and source traversal epochs
remain intact. Reusing the existing fixed-array logical-fanin helper avoids
additional per-node slice allocation or a second walker. Metric/refcount and
mini storage are already direct source-shaped Go; no substitute algorithm or
new cache is introduced to simplify tests.

Remaining work includes boxed timing/choice coverage and reference boundaries,
configuration/frame metadata, IF reconstruction and all mapping drivers, error
and resource cases, source precondition/safety-extension characterizations,
compactness measurements and frozen full-module integration. The source family,
GIA package and original2/131 package audit remain open.

## Boxed levels and timing metadata checkpoint

Freshly compiled original giaIf.c and giaTim.c agree with Go on128 repeated
boxed/fallback states across64 seeds. Valid cases also match independent
construction-time level expectations, including multiple box inputs/outputs,
disabled arcs, varied finite delays, dangling box interfaces and separate LUT
versus choice-level semantics. Invalid box order retains the original warning
and structural-level fallback. No production timing change was needed.
`/tmp/abcgo-mapping-box-levels-first.log`,handle76413 exit0,2.339s.

Fresh original giaIf.c matches6,144 timing metadata transfers:2,048 presence
masks with absent/populated destinations and self-transfer. Compare all values,
source clearing, unchanged scalar fields and exact source/destination pointer
origins, including partial groups and boxed-resource gating. Fixture allocations
are tracked independently so destructive self-transfer/overwrites do not leak
in the oracle. Existing source-ordered self-transfer is intentional and retained.
`/tmp/abcgo-mapping-timing-transfer-first.log`,handle22770 exit0,2.099s.

Inspection finds a separate concern for the next red test: the SOP/DSD balance
wrappers reuse the general mapping driver. That driver clears cell configuration
metadata, whereas the original balancing paths do not clear it before moving
timing/configuration data to their results. Confirm this loss before repair.
Family-wide remaining reconstruction/driver/frame/error/compactness/integration
work is still open; these coverage milestones do not close GIA or map/if.

## Balancing driver repair: preserve the original metadata policy

New Go regressions fail in all four direct/registered SOP/DSD entry points:
cell configurations disappear and existing output arrivals91 become14. Source
inspection shows balancing uses its own translate/map/rebuild path, while Go
reused general MappingInt, which intentionally clears configurations and installs
input timing constraints. The first test harness needed qualified GIA aliases;
behavioral red evidence is `/tmp/abcgo-mapping-balance-metadata-behavior-red.log`,
handle38586 exit1, after that harness correction.

Fresh original-C SOP and DSD balancing retain the configuration pointer and
cell string, move source ownership, retain input timing and output arrival91,
and preserve identity/truth. The generic C mapping path intentionally clears
the old configuration and computes arrival14. All three native policies pass
(`/tmp/abcgo-mapping-balance-native-policies.log`,handle27681 exit0,2.094s).

Implement the source balancing sequence in one small shared private helper,
keeping actual If mapping, SOP/DSD reconstruction and timing transfer. Do not
route it through the generic driver's destructive configuration/timing setup.
Keep the existing Go-owned per-call DSD manager lifetime. Tests cover direct and
registered wrappers,32 metadata/interface/truth configurations plus a generic
mapping cross-talk regression. Full map/if race suite passes7.997s
(`/tmp/abcgo-mapping-balance-full-if-race.log`,handle90332 exit0).

A separate newly inspected concern is source-name loss in Gia_ManToIf; a new
regression is pending. Remaining family rows and full module integration remain
open. A green caller suite does not close GIA or the whole mapping boundary.

## IF source-name repair and broader driver comparison

The new source-name regression fails because Gia_ManToIf returns an empty IF
manager name (`/tmp/abcgo-mapping-if-name-red.log`,handle31268 exit1). Original
C copies Gia_ManName into independently owned IF storage. Copy the immutable
Go name during manager creation; renaming/stopping the source cannot alter it.
Fresh C naming and all three balancing/general metadata policies pass4.112s
(`/tmp/abcgo-mapping-balance-name-native.log`,handle47836 exit0). The corresponding
Go focused suite passes0.025s (handle91491 exit0), and full map/if race tests
pass8.098s (`/tmp/abcgo-mapping-if-name-full-race.log`,handle18506 exit0).

Correct the balancing matrix count: two engines x two register counts x four
timing masks x two configuration states is32 cases (512 assignments), in
addition to four direct/registered regression cases. Earlier64 wording was a
counting error, not missing test execution.

The new full-driver oracle compares actual freshly compiled C with the actual
IF-owned Go mapper, not injected test engines. Initial96 comparisons have31
raw-vector differences, all confined to hashed mapping storage. Inspection of
all31 shows identical active root/ordered-leaf/marker records, graph objects,
truth, metadata and numeric statistics. C retains original-size index padding
and unreachable duplicate records; Go's existing reviewed compaction removes
them. Canonicalize only that storage representation in the oracle, retaining
every active record and all other structured output. This is a documented
representation normalization, not a production algorithm repair or omission of
functional differences. Canonical rerun remains pending at this checkpoint.

## Expanded driver/cut-policy checkpoint (boundary repair pending)

The initial96 driver comparisons pass4.421s after storage-only normalization
(`/tmp/abcgo-mapping-driver-native-active-records.log`,handle76378 exit0).
Expand to384 runs with higher phase bits, K=3/4 LUTs, independent register
counts, choices, generic truth/structural paths and actual SOP/DSD engines.
This exposes six live-cut policy differences: C overwrites merged roots with
the last source cut, while Go had retained the first. A new minimal red test
fails before repair (`/tmp/abcgo-mapping-hash-last-cut-red.log`,handle15272 exit1).

Select the last cut in a first pass using the destination index as scratch,
then emit one compact record per destination root. This preserves source cut
selection without dead records or an additional selection array. The384 native
runs and existing/minimal hash tests pass11.601s
(`/tmp/abcgo-mapping-driver-last-cut-repaired.log`,handle31412 exit0). Active
LUT records use the original iterator's object1-and-up domain. Explicitly
characterize source seed16/hash mode retaining a redundant two-leaf object0
record: neither LUT iterator counts it, and Go already omits it. Constant
physical nodes/COs and every truth assignment remain compared.

Do not accept last-cut selection without checking hash collapse. A new boundary
regression now fails: x AND duplicate(x) collapses to x, and the last cut becomes
self-referential (`root3 leaves[3 3]`).
`/tmp/abcgo-mapping-hash-self-cut-red.log`,handle22663 exit1. Save this red
checkpoint, then characterize the original C boundary and retain the last usable
cut rather than introducing a cycle. This is an active repair, not a completed
source family or a green integration claim.

## Hash-collapse boundary repair and current integration status

The forward-cut regression also fails before its guard: root5 leaves[1 2 6]
(`/tmp/abcgo-mapping-hash-forward-cut-red.log`,handle89158 exit1). Fresh C
characterization reproduces last-cut selection and both source failures exactly:
root4 cut[1 2 3], root4 self-cut[4 4], and root5 forward-cut[1 2 6]
(`/tmp/abcgo-mapping-hash-source-boundaries.log`,handle50258 exit0,1.985s).

Keep the last source cut whose mapped leaves precede the destination root.
Hash collapse cannot replace a usable earlier cut with a self/forward record.
This is an explicit correctness repair beyond those C boundary bugs, not a
claim that the invalid source metadata was reproduced. Valid-domain cut policy
and all384 complete native driver results still match. Hash-focused/native
checks pass13.522s (handle83191 exit0); full map/if race tests pass8.110s
(handle75750 exit0). Broader focused GIA coverage then passes75.457s without
failures or skips (`/tmp/abcgo-mapping-through-drivers-focused.log`,handle54778
exit0). Full frozen module race/vet is still pending for this mapping family.

## Required source-fidelity restoration still open

Inspection confirms `map/if/ifSat.go` explicitly replaces original SAT with
cofactor classification. BuildXY/XYZ return a placeholder plan, Unbuild is a
no-op, and CheckXY ignores its solver argument. This may solve the same Boolean
relation on supported partitions but violates the user's requirement for a
faithful mechanical port rather than substitute algorithms. Restore original
ifSat.c CNF construction, remapped assumptions, actual ported BSAT solve/reuse
and model extraction before closing structured-LUT reconstruction/FindLut.

The next regression/implementation plan is
`/tmp/abcgo-mapping-sat-restoration-plan.md`. It includes fresh C boundary cases,
retained assumptions, solver lifecycle and repeated SAT/UNSAT tests, production
caller ownership and full-word truth checks. No substitute is accepted merely
because SAT decisions match. This dependency, remaining source rows, resource/
error boundaries and complete integration keep the source family and GIA open.

## Corrected cut selection: final checkpoint measurement

Select the last usable cut by walking source cuts backwards, skipping a merged
root once selected. Reuse the existing destination index as scratch; emit one
record per root. This preserves the verified policy while avoiding validation
of all earlier duplicate candidates. The native/hash suite still passes14.640s
(`/tmp/abcgo-mapping-hash-reverse-focused.log`,handle83094 exit0), including all
384 driver results and new valid/invalid-cut regressions. C's cyclic/forward
records remain separately characterized rather than copied into Go.

Initial150ms timing samples were noisy, so compare the two correct selectors
serially at500ms/count3. The ascending version is loaded through a Go overlay
from21e174f04; the worktree stays unchanged. Median local results:

| Duplicate roots | Correct ascending ns/op | Correct reverse ns/op | B/op (reverse) | Allocs/op (both) |
| --- | --- | --- | --- | --- |
|64|12,962|12,025|10,132|87|
|4,096|835,374|786,978|558,326|4,127|
|65,536|12,844,886|11,521,532|8,494,202|65,571|

Allocation is unchanged (minor byte differences are measurement noise); reverse
selection saves roughly6–10% time here. This table compares implementations of
the corrected policy, with no universal timing claim. Logs:
`/tmp/abcgo-mapping-hash-ascending-serial-bench.log` (handle12135 exit0) and
`/tmp/abcgo-mapping-hash-reverse-serial-bench.log` (handle8201 exit0).

All work is saved in coherent current-branch checkpoints. The next required
algorithm restoration is original ifSat.c SAT flow; the full mapping family,
GIA and original2/131 package goal remain open. Full frozen module race/vet
will be required after the remaining mapping work, not inferred from these
focused/caller results.

## Original SAT core restoration checkpoint

REVIEW_mapping_sat_dependency.md tracks the required ifSat.c dependency. New
red tests prove observable substitute differences: no actual CNF, missing
assumptions, rejected source-valid degenerate partitions, single-output writes
and UNSAT output clobbering. Fresh original C confirms the source contracts.
Restore the original XY/XYZ MUX clauses through existing BSAT helpers, incremental
solve/assumption remap/model extraction and structured-reconstruction ownership.
Nil convenience allocates a real temporary source solver, never the substitute.
Initial SAT checks pass0.235s; BSAT race suite6.945s. Full dependency differential,
search/composition/demo review, caller/resource checks and family integration
remain open; this checkpoint is not mapping/GIA completion.

## Structured mapping dependency checkpoint

ifSat.c core/search/examples/composition now use actual source CNF and BSAT;
native core420, search105, helper1,152 and XYZ960 cases, including race evidence,
are recorded in REVIEW_mapping_sat_dependency.md. The structured22/33 caller
matrix exposes ifDsd.go's truth-record/partition-enumeration substitute. A new
minimal red native test proves missing canonical interning/permutations. Restore
the original DSD representation and traversal under the concrete sequential
plan in REVIEW_mapping_dsd_dependency.md. Correct command-path flags matter:
fCutMin=1,fExpRed=0. The corrected128-case expansion has97 exact matches,
29 structural mismatches and2 explicit original partition-assertion failures.
These are evidence/remaining scope, not a completed family or package audit.

## Go-only adapter and complete driver matrix checkpoint

Default/choice/ToIf source rows reconciled against giaIf.c. The adapter retains
source object/value order for ordinary graphs, phase-correct typed expansions,
CI levels and name ownership. Signed copy storage is needed for collapsed buffer
aliases and typed expansion; replacing it with source IDs would lose those phases.
The original C memory warning uses packed allocation sizes; no equivalent claim
about Go heap use is made. Existing ordinary source/boxed-level evidence remains
additional context, not a runtime C dependency for these new checks.

512 configured driver/balancing cases and64 direct choice-chain cases execute
without C tools: scalar truth/interface/identity/configuration checks, ordinary
source-ID correspondence, expansion policy, reachable alternatives and repeated
mark-clearing diagnostics. The older unfiltered native512-case setup is retained
unchanged (510 source results plus2 infeasible-cut assertions); Go-only structured
cases enable the actual source cell-filter branch and its minimum four-slot cut
storage. Physical22/33 LUTs are distinct from that mapper storage limit.

New tests pass1.158s, focused race4.774s and IF/GIA vet pass. Full-module Go-only
JSON confirms512 driver and64 choice subcases pass,130 tested packages pass and
3 commands have no tests. No failures. Logs /tmp/abcgo-gia-mapping-go-matrices.log,
/tmp/abcgo-gia-mapping-go-matrices-race.log,
/tmp/abcgo-gia-mapping-go-matrices-vet.log,
/tmp/abcgo-module-go-only-bridge.jsonl. Older optional C comparisons still skip;
no broad replacement/comprehensive coverage claim for them. Next frame timing
boundaries and remaining IF-to-GIA reconstruction/configuration rows; final
mapping-family full race/vet gate and original131-package review remain open.

## Frame timing boundary review

Reconcile all four frame adapters with source. Owned input arrays are copied on
installation/replacement; output arrays are returned independently. Nil arrays
are checked before frame lookup and leave existing state intact. New Go-only
checks cover replacement, caller mutation, invalid dimensions without mutation,
nil frame/network, no output timing, empty interfaces, repeated readers and
signed/large gate delays. Preserve safe existing Go returns where the C delay
setter would continue after diagnosing a null frame/network. No functional fix
or additional abstraction was needed; the straight adapters are already compact.

Focused tests0.007s, full GIA Go-only27.207s, focused frame race and IF/GIA vet
pass. Logs /tmp/abcgo-mapping-frame-{go,gia-go-only,race}.log and
/tmp/abcgo-mapping-frame-cone-vet.log. No C vectors/compilers used by the new
coverage. The mapping-family final gate and remaining reconstruction rows stay
open; original audit2/131.

## Cut-cone and AIG reconstruction review

Reconcile source choice search, phased alternatives, terminal copy preconditions,
structural hash recursion, mapped-cut expansion and recorded-library callback
order. New64 cut-cone cases use equivalent AND/NOR choices with two unusable
outside-variable representations before a valid alternative. Independently
verify the original choices over all16 assignments and the reconstructed cut
truth over all4 assignments, with signed supplied leaves and both hash modes.
Scratch data survives repeated calls, failure followed by a new valid search,
and constant mapped leaves. Local Go memo storage preserves caller scratch;
the source uses destructive temporary cut fields. Normal source traversal,
phase correction and node construction order remain intact.

New64 recursive-copy cases check all16 inputs and memoized graph reuse; repeated
recorded-library reconstruction and a seven-input DSD parity graph check complete
truth and scratch-vector growth. Existing source/Go driver matrices remain
integration evidence for SOP/DSD/recorded construction. The unsupported source
user-SES reconstruction branch remains explicitly unsupported, not substituted.

Compactness: the source's65536-int mini vector was allocated even when recorded
callbacks never use it. Allocate a small growing vector only for balancing.
No algorithm/profile/node-order change. Recorded two-input benchmark (500x,
three samples) decreases269282..269324 B/op and57 allocations to7104..7114 B/op
and55 allocations; this is a local allocation result, not a general mapper
speedup claim. Logs /tmp/abcgo-mapping-from-if-scratch-{before,after}.log.

Expanded tests0.021s, full IF Go-only1.273s, reconstruction race1.117s,
512-driver race4.692s and IF/GIA vet pass. Full-module Go-only gate exits0,
130 tested packages and3 commands without tests. Logs:
/tmp/abcgo-mapping-reconstruction-{after,if-go-only,if-race,drivers-race,vet}.log,
/tmp/abcgo-module-go-only-reconstruction.jsonl. Broader mini vector checks still
need separation from their old compiler-gated comparison, then continue LUT/
cofactor/config/cell rows and final family gate. Original audit2/131 stays open.

## Mini-AIG matrix runs without C tools

Move the reusable Go case definitions/scalar evaluator out of the compiler-gated
comparison and execute all140 in an ordinary Go-only test. Cover constants,
variables, signed leaves, hash/append modes, complete truth, unchanged input
leaves and source scratch-clearing behavior. Existing invalid-input checks also
run without C. The optional native comparison retains the same cases and original
settings; no .c fixtures or expected source outputs are generated for this work.

Focused tests0.010s, full GIA Go-only23.790s, mini race1.043s and IF/GIA vet pass.
Logs /tmp/abcgo-mapping-mini-{go-only,gia-go-only,race,vet}.log. No functional
change. The mini/default/choice/ToIf/frame plan section is now reviewed; the final
family integration gate remains pending, as does the overall131-package audit.

Next: direct LUT reconstruction across physical sizes2..6 and supports through
2*K-1. Exercise selected versus searched DSD partitions, signed/permuted supplied
leaves, complemented cut functions, multiword truth, packing/mapping records and
all input assignments. The512 driver matrix has only four primary inputs, so it
cannot establish those larger-support reconstruction boundaries. Continue source
LUT/cofactor/config/cell rows rather than counting driver cases as full coverage.

## Larger LUT reconstruction source/scalar review

New400 Go-only cases cover physical LUT2..6, supports0..2*K-1, selected/searched
partitions, every feasible shared-input count, cut complementation and signed
supplied leaves. Independently evaluate every input assignment, verify immutable
leaves and inspect packed group counts, topology and physical input limits.
Another36 cases exercise direct CreateLut at supports0/1/2/5/6/7/8/11/15 with
both output/leaf phases, full truth/record checks and immutable caller tables.
No functional mismatch found in these two source-shaped helpers. Full IF
Go-only1.831s, focused race3.540s and IF/GIA vet pass. Logs:
/tmp/abcgo-lut-reconstruction-wide.jsonl,
/tmp/abcgo-lut-wide-{if-go-only,race,vet}.log.

Inspection of the specialized4-4 path finds a separate fidelity issue in its
ifDec07 dependency: Go substitutes generic bound/free combination enumeration
for source6-/7-input ordered variable moves. Existing native tests compare only
existence, not emitted configuration or graph order. This is not covered by the
passing436 cases above. Add source-derived exact configuration regressions before
restoring the source search/derivation. Review the five-input source helper too;
do not call specialized4-4 or the mapping family complete while it is generic.
No new C fixtures/compiler requirements. Original package audit remains2/131.
