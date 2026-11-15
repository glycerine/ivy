# Timeframe unrolling review — in progress

Review `src/aig/gia/giaFrames.c` against `giaFrames.go`. Reconcile each active
source routine, excluding the commented old Unroll implementation at lines
540–595. The current full-frame map and common AIG-frame wrapper are not
accepted substitutes for the source tent/history scheduler and distinct entry
points. Prior CCof coverage proves many Boolean cases, not scheduling fidelity.

## Established red evidence and current repair

A17-object graph with two PIs, three registers, six ordinary ANDs and two POs
has native dependency limits `[1,8,15,17]`. The previous Go clone returned
`[1,17]`. Its abstraction rows contained9/18/27 nodes instead of4/8/9. Native
Frames2 returns ci9/co6/regs0/objs29; Go's common wrapper returns
ci9/co9/regs3/objs37. Incremental first-step source ci5/co0/objs8/lit14 differs
from Go ci5/co0/objs12/lit12. All four native modes and three minimal regressions
fail before production changes (`/tmp/abcgo-frames-scheduling-red.log`).

Port the source PO-cone DFS followed by reached-RO/RI dependency tents. Publish
source copy literals and destination original IDs in Value, retaining their
distinct meanings. The unvisited Value sentinel is all32 bits set, distinct
from the29-bit no-fanin sentinel. Extend traversal/copying to logical real gate
inputs, including MUX control after stored data pins. Unreachable nodes remain
absent. Build sparse reverse-time abstraction directly from these tents.

Order and abstraction match every native object field and packed row on the
initial graph. The rolling scheduler, cleanup and combinational Frames2 now
match all five native modes, including save/fixed property truth. Fixed mode
leaves LastLit0; save mode selects the last encountered PO. The old always-PO0
expectation and old register-retaining Frames2 test are corrected using these
native contracts, with an additional red overlay regression against the prior
implementation (`/tmp/abcgo-frames-fixed-literal-red.log`).

Port the compact source history arena, ranks/delays and scheduler in
`frames_unroll.go`; remove redundant per-frame maps, state and stopped fields.
Keep lifecycle/huge incremental-bound behavior and correct initialized-save PI
offsets. Focused native/CCof/independent sequential checks pass13.347s
(`/tmp/abcgo-frames-scheduler-consumers-repaired.log`), including1536 frame
configurations and176640 assignments.

Distinct Frames/FramesInit/InitSpecial implementations now live in
`frames_driver.go`; the AIG roundtrip substitute is removed. New OR-output and
prefix tests were red before this port (`/tmp/abcgo-frames-drivers-red.log`).
Twenty-four explicit corrected-source driver configurations match all output
object records. See below for the separate original-C initialization defect.

Fixed zero frames now rejects positive Add horizons while incremental zero
remains unbounded. A red boundary test catches that distinction
(`/tmp/abcgo-frames-zero-bound-red.log`). Specialized prefix clears preceding
RI state before an immediate event, instead of returning stale scratch or a
property. The prior implementation fails the immediate-event regression
(`/tmp/abcgo-frames-special-scratch-red.log`). The source stale-RI assertion is
captured separately with core dumps disabled. Latest focused frame/native and
lifecycle checks pass9.578s (`/tmp/abcgo-frames-drivers-boundaries.log`). This is
partial coverage, not the final frozen integration gate or family closure.

The permanent fixture `test_vectors/frames_review.c` includes actual source and
links original ABC objects through the shared Linux/GCC source-oracle helper.
It checks exact object fields, limits, sparse rows and all32/128/512 assignments
of the successive incremental last-property literal. No solver or algorithm
answers are stubbed. Broader generated/native and independent coverage remains
required after the rest of the source algorithm is restored.

## Remaining review and coverage plan

All active source entry points now have direct implementations. Continue
reconciling the full22-routine active inventory, mapping the source static
history read/write/copy helpers to Go methods. The inactive old Unroll function
remains excluded.

1. Compare native rank delays, offsets and history rows before/after steps over
   generated ordinary graphs. Preserve two-column ordinary layouts and test the
   derived third column for typed MUXs independently.
2. Expand exact native dependency order, sparse abstraction and frame-driver
   flag matrices over generated graphs, not just the initial17-object example.
   Characterize unmodified-source defects separately from explicitly corrected
   native comparisons.
3. Add independent full Frames/FramesInit sequential histories for ordinary,
   real-gate and reordered interfaces; verify final state, OR aggregation,
   initialized flag interactions, output/input dimensions and returned ownership.
   Existing1536-case coverage concerns the incremental family, not all drivers.
4. Expand specialized-prefix cases for immediate/delayed events, finite zero
   prefixes, multiple properties, typed gates and owned preceding-state outputs.
   Add exact native prefix and verbose comparisons, normalizing only genuinely
   nondeterministic clock/memory fields when appropriate.
5. Audit history lifecycle, zero/large bounds, safe power-of-two packed IDs and
   resource contracts. Review/remap changed-interface metadata deliberately.
6. Measure compactness and allocation behavior against the previous full-map
   implementation. Inspect skipped-tent work and unused helpers, retaining exact
   source ordering/history/native results through any simplification.
7. Update source reconciliation/validation, freeze production/tests/fixtures,
   run uncached full module race tests and GIA vet, and inspect exact terminal
   results before closing the family. The broader GIA inventory stays open.

## Source boundary issues to characterize explicitly

C initialized incremental mode indexes PIs after a RegNum offset despite not
allocating symbolic initial registers. Preserve the already tested valid Go
initialized mode; characterize the source assertion and apply an explicit
correction rather than copying the bad index. Source Add expects sequential
horizons and allocates one PI frame per call; the Go jump wrapper must invoke
that operation repeatedly when appropriate.

C Abs's mask assertion rejects power-of-two object counts although their IDs
fit. The Go bit width uses the actual maximum ID; test these boundaries
independently and keep unmodified native fixtures within supported counts.

Initialized source FraSupports has a second defect: CollectCis marks the cone,
then CollectAnds intentionally shares that epoch and returns no AND nodes.
The default C fixture stays unmodified. Its --init-defect mode must reproduce
wrong source outputs and compare an explicit corrected call-site bridge with
independent sequential truth over128 records. The --drivers-safe mode enables
only that epoch correction; source schedulers and frame builders are unchanged.
Go starts a fresh epoch at the same support-call site, leaving the shared DFS
API's caller-owned epoch contract intact.

Specialized source prefix can return stale RI Value when a first-frame event
breaks before committing any transition state. --special-stale characterizes
its assertion after FillValue; Go initializes the preceding RI state to zero.
These repairs preserve valid source intent rather than copying demonstrated
source errors.

Unrolling changes interface dimensions and ordering. Remap metadata explicitly
where meaningful; do not blindly copy unchanged-interface metadata onto a new
frame interface or copy node-indexed payloads as if their IDs survived.

Read-only preparation and additional native logs are indexed in
`../../SEQUENTIAL_AUDIT.md` and `/tmp/abcgo-frames-next-review-plan.md`.
The original sequential audit remains at2/131 completed packages; GIA is open.
