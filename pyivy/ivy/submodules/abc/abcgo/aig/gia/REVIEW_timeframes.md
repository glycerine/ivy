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

Order and abstraction now match every native object field and packed row on the
initial graph. Both minimal tests and the corresponding native modes pass
(`/tmp/abcgo-frames-order-abs-progress.log`). Frames2 and incremental modes
remain red. This is partial evidence, not family closure or a full green gate.

The permanent fixture `test_vectors/frames_review.c` includes actual source and
links original ABC objects through the shared Linux/GCC source-oracle helper.
It checks exact object fields, limits, sparse rows and all32/128/512 assignments
of the successive incremental last-property literal. No solver or algorithm
answers are stubbed. Broader generated/native and independent coverage remains
required after the rest of the source algorithm is restored.

## Remaining implementation and review plan

1. Port UnrStart's ranks, per-edge rank delays, maximum required history and
   compact rolling storage; retain ordinary native array layouts. For real MUXs,
   derive a third edge column only when needed. Temporary construction rank and
   degree arrays need not become redundant persistent Go fields.
2. Port source history read/write helpers and sequential Add tent processing.
   Preserve valid Go horizon jumps by executing missing mechanical steps, not
   by replacing the scheduler with whole-network maps. Preserve stopped-manager
   checks, caller ownership of returned graphs and huge incremental bound
   support established by the earlier CCof review.
3. Reconcile LastLit: source save mode ends at the last encountered PO; fixed
   mode does not update it. Current Go explicitly chooses PO0 in both modes.
4. Make Frames2 delegate to the combinational Unroll routine, then reconcile
   source cleanup and ordering. The old KeepsStateInterface test asserts the
   substituted wrapper's behavior; correct it using the native contract.
5. Port Frames/FramesInit separately, including support-pruned initialized
   construction, FOrPos, FDisableSt, register/terminal dimensions and reports.
   The incremental source family always hashes; these flags are not generic
   wrapper options to apply to every entry point.
6. Port InitSpecial's first-nonzero-property prefix termination and preceding
   transition-state output. Zero bound means event-driven unbounded mode. Use
   terminating toggle/event cases; never launch a permanently-zero C loop.
7. Add native scheduling/history/order and diagnostic matrices, independent
   exhaustive sequential traces, typed/reordered/reset/boundary/resource tests,
   ownership and measured compactness checks. Then freeze code/tests/fixtures,
   run full uncached module race tests and GIA vet, and reconcile the complete
   active source inventory before closing this family.

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

Unrolling changes interface dimensions and ordering. Remap metadata explicitly
where meaningful; do not blindly copy unchanged-interface metadata onto a new
frame interface or copy node-indexed payloads as if their IDs survived.

Read-only preparation and additional native logs are indexed in
`../../SEQUENTIAL_AUDIT.md` and `/tmp/abcgo-frames-next-review-plan.md`.
The original sequential audit remains at2/131 completed packages; GIA is open.
