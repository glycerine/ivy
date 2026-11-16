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
