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
4. [ ] Review mini reconstruction and parameter/GIA-to-IF boundary, choice
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
| Gia_ManSetIfParsDefault | map/if/gia_bridge.go | Pending |
| Gia_ManLutFaninCount | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManLutSizeMax | aig/gia/mapping_accessors.go, base/abci/abc_graph_compat.go | Core coverage passes; integration/boundaries pending |
| Gia_ManLutNum | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManLutLevel | aig/gia/mapping_metrics.go | Core coverage passes; integration/boundaries pending |
| Gia_ManLutParams | aig/gia/mapping_timing.go | Pending |
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
| Gia_ManChoiceLevel_rec | aig/gia/mapping_choice_level.go | Pending |
| Gia_ManChoiceLevel | aig/gia/mapping_choice_level.go | Pending |
| If_ManCheckChoices_rec | map/if/choice_check.go (ifDanglingChoices closure) | Pending |
| If_ManCheckChoices | map/if/choice_check.go | Pending |
| Gia_ManToIf | map/if/gia_bridge.go | Pending |
| Gia_ManBuildFromMiniInt | aig/gia/mapping_mini.go | Core coverage passes; integration/boundaries pending |
| Gia_ManBuildFromMini | map/if/gia_mini.go | Pending |
| Gia_ManFromIfAig_rec | map/if/gia_from_if.go | Pending |
| Gia_ManFromIfAig | map/if/gia_from_if.go | Pending |
| Gia_ManFromIfLogicCreateLut | map/if/gia_lut.go | Pending |
| Gia_ManFromIfLogicCreateLutSpecial | map/if/gia_lut.go | Pending |
| Gia_ManFromIfLogicNode | map/if/gia_lut_structure.go | Pending |
| Gia_ManNodeIfToGia_rec | map/if/gia_cut_cone.go (recursive closure) | Pending |
| Gia_ManNodeIfToGia | map/if/gia_cut_cone.go | Pending |
| Gia_ManFromIfLogicFindLut | map/if/gia_lut.go | Pending |
| Gia_ManFromIfGetConfig | map/if/gia_config.go | Pending |
| Gia_ManFromIfLogicFindCell | map/if/gia_cell.go | Pending |
| Gia_ManFromIfLogicCofVars | map/if/gia_cofactor.go | Pending |
| Gia_ManFromIfLogicAndVars | map/if/gia_cofactor.go | Pending |
| Gia_ManFromIfLogic | map/if/gia_mapping.go | Pending |
| Gia_ManMappingVerify_rec | aig/gia/mapping_transfer.go | Pending |
| Gia_ManMappingVerify | aig/gia/mapping_transfer.go | Pending |
| Gia_ManTransferMapping | aig/gia/mapping_transfer.go | Earlier native/scalar review; reconcile |
| Gia_ManTransferPacking | aig/gia/mapping_transfer.go | Earlier native/scalar review; reconcile |
| Gia_ManTransferTiming | aig/gia/timing_metadata.go, base/abci/abc_graph_compat.go, base/wlc/wlc_compat.go | Pending |
| Abc_FrameMiniAigSetCiArrivals | aig/gia/frame_bridge.go | Pending |
| Abc_FrameMiniAigSetCoRequireds | aig/gia/frame_bridge.go | Pending |
| Abc_FrameMiniAigReadCoArrivals | aig/gia/frame_bridge.go | Pending |
| Abc_FrameMiniAigSetAndGateDelay | aig/gia/frame_bridge.go | Pending |
| Gia_ManPerformMappingInt | map/if/gia_mapping_driver.go | Pending |
| Gia_ManPerformMapping | map/if/gia_mapping_driver.go | Pending |
| Gia_ManPerformSopBalance | aig/gia/balance_engines.go, map/if/gia_mapping_driver.go | Pending |
| Gia_ManPerformDsdBalance | aig/gia/balance_engines.go, map/if/gia_mapping_driver.go | Pending |
| Gia_ManTestStruct | map/if/gia_structure_report.go | Pending |
| Gia_ManDupHashMapping | aig/gia/mapping_hash.go | Earlier native/scalar review; reconcile |

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
