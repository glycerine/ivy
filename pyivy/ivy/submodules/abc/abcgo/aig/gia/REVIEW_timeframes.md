# Timeframe unrolling review

Reconcile all22 active routines in `src/aig/gia/giaFrames.c` with
`giaFrames.go`, `frames_unroll.go` and `frames_driver.go`. Exclude the commented
old Unroll implementation at source lines540–595. Preserve source dependency
tents, rank-delayed rolling history, sparse abstraction, frame construction,
OR aggregation and specialized prefix termination. The former full-map/common
AIG-frame substitutes did not implement these distinct source contracts.

## Repairs and red evidence

The initial17-object two-PI/three-register graph has native dependency limits
`[1,8,15,17]`; old Go returned `[1,17]`. Its abstraction rows contained9/18/27
objects instead of4/8/9. Source Frames2 returned ci9/co6/regs0/objs29, while
Go's shared wrapper returned ci9/co9/regs3/objs37. The first incremental step
also had different node counts and the wrong property literal. All minimal
and four native-mode comparisons fail before repair
(`/tmp/abcgo-frames-scheduling-red.log`).

Restore PO-cone DFS followed by reached-register RI tents. Publish source copy
literals and destination original IDs in their respective Value scratch.
Unvisited Value is all32 bits set, distinct from the29-bit no-fanin sentinel.
Build sparse reverse-time abstraction from tents, then port ranks, per-edge
delays and compact rolling histories. Preserve ordinary source layout and
extend logical fanin handling to real XOR/MUX/buffers. MUX histories use a third
edge column only when needed. Rank/degree construction arrays remain local.

Source save mode ends on the last encountered PO; fixed mode leaves LastLit0.
The old implementation always selected PO0. A separate regression against the
prior implementation is red on fixed mode saving property6
(`/tmp/abcgo-frames-fixed-literal-red.log`). Correct the older tests that encoded
the substituted behavior using actual native fixed/save and Frames2 contracts.

Frames2 now delegates to combinational Unroll and source cleanup. Frames,
FramesInit and InitSpecial have separate direct implementations. New OR and
prefix regressions first fail on9 instead of5 COs, nil unbounded prefix, and
returning properties instead of preceding transition state
(`/tmp/abcgo-frames-drivers-red.log`). Preserve all PIs, source support pruning,
nonhashed mode, ORed PO columns, retained register ordering and exact banners.

Fixed zero-frame bounds must reject positive Add horizons, while incremental
zero remains unbounded (`/tmp/abcgo-frames-zero-bound-red.log`). The complete
wrapper must produce its promised per-frame outputs: C asserts on incompatible
save-only mode; old Go silently returned no outputs
(`/tmp/abcgo-frames-resource-red.log`). Reject that mode before allocation.

Noninitialized Frames keeps the same logical registers but dropped Go's
explicit initial states. A twelve-case regression is red on `[1,-1]` becoming
empty (`/tmp/abcgo-frames-resets-red.log`). Preserve the owned reset cube after
cleanup. Explicit FInit continues to request zero initialization, and the
combinational variants have no latch state metadata.

## Source defects and reference policy

The native fixture includes actual giaFrames.c and links original ABC objects.
The shared oracle requires Linux/GCC and those built objects and skips when
unavailable. It supplies no algorithm or solver answers.

Default modes execute unmodified C. Separate required modes characterize:

- Initialized save mode indexes PIs after RegNum despite allocating no symbolic
  initial registers. A child captures its index assertion. Go retains the
  corrected zero offset, verified by independent initialized histories.
- FraSupports collects CIs, marking the whole cone, then CollectAnds intentionally
  shares that epoch and returns no nodes. The128-record --init-defect mode
  reproduces wrong original outputs and checks independent sequential truth.
  Only explicitly corrected driver/report modes enable a call-site bridge
  that advances the epoch before support AND collection. Go makes that same
  correction without changing the shared DFS epoch contract.
- InitSpecial can return stale RI Value when a first-frame event precedes any
  committed transition. --special-stale captures its assertion after FillValue.
  Go initializes preceding RI state to zero; the old Go substitute is separately
  red on returning property1 instead of state0
  (`/tmp/abcgo-frames-special-scratch-red.log`).
- Abs's mask assertion rejects power-of-two object counts although their IDs
  fit. A source child captures that assertion; Go independently checks packed
  IDs at4/8/16/32 objects using the actual maximum ID.

Core dumps are disabled for expected-assertion children. None of these source
failures are silently replaced by successful oracle answers. Corrected C modes
change only the explicitly named support-collection epoch, leaving the source
schedulers, histories and frame builders intact.

## Active source reconciliation

| C routines | Go implementation and evidence |
| --- | --- |
| UnrollDup_rec, UnrollDup | `giaUnrollDupRec` and public ordered clone: exact native limits/original-ID/object records across generated graphs; logical real fanins and unreachable-node omission. |
| UnrollAbs | Source sparse packing/sorting: initial exact rows,160 generated native configurations/480 rows, safe power-of-two IDs and zero/empty boundaries. |
| UnrStart | Local rank/degree construction, edge delays, flat owned arena:64 generated fixed/save managers with exact source tables, plus a twelve-register uneven-rank chain. |
| ObjUnrWrite, ObjUnrRead | Go `write`/`read`: every native history cell at startup and after three generated steps, and a twelve-cell history through sixteen horizons. |
| ObjUnrReadCopy0, ObjUnrReadCopy1 | Unified pin-indexed `readCopy`: source phase/delay records, exact resulting object fields and exhaustive Boolean traces. |
| ObjUnrReadCi | `readCI`: fixed/save PI layouts, symbolic/zero initial state, delayed RI reuse, typed/reordered and buffer matrices, native assertion characterization for initialized save. |
| UnrollStart, UnrollAdd | Source scheduler and increasing-horizon wrapper: exact fixed/save object/history records, sequential and jumping horizons, large incremental bounds and resource/lifecycle coverage. |
| UnrollLastLit | Exact native literal and all32/128/512 initial-graph assignments; fixed-mode zero, last saved PO and deeper512 traces. |
| UnrollStop | Release order/history/config references, idempotent stopped-handle rejection; returned graph survives source reuse and disposal. |
| Unroll, Frames2 | Complete combinational outputs and cleanup: exact native fields, output contract, zero-frame boundary, source convergence/time reports. |
| FraSetDefaultParams | Source memset/32-frame default reconciled with the Go zeroed struct literal; option semantics exercised throughout driver matrices. |
| FraStart, FraStop, FraSupports | Owned frame support sets, corrected call-site epoch, exact source driver order/count/reports and independent initialized traces; source remains caller-owned. |
| FramesInit | Source support-pruned zero initialization/all-PIs policy; native corrected matrices, buffers, changed layouts and per-frame reporting. |
| Frames | Source noninitialized temporal network, ORed/plain columns and nonhashed/hash modes, terminal registers and owned reset cube. |
| FramesInitSpecial | Source first-nonzero-literal rule and preceding-state outputs;96 generated and five directed unmodified-C records,192 mixed prefixes, immediate stale-scratch repair and exact reports. |

The22-routine count is checked against active function definitions after
excluding comments. Combined table rows name every routine, including static
helpers; the three Go files contain the actual algorithms, not compatibility
wrappers to another frame implementation.

## Independent and native coverage

All768 generated corrected-source driver configurations match exact object
records and53,760 independent sequential assignments.64 generated fixed/save
managers match limits, every edge delay, offsets and all arena cells at startup
and after each of three steps, plus ordered and produced object fields.

A separate1,536-case ordinary/real-gate/interleaved/reordered/flag matrix checks
107,520 independent complete-driver assignments. It verifies identity/parameter
ownership and traverses/evaluates returned graphs after source disposal.
Additional128 buffer/mixed driver configurations check49,920 histories;
96 incremental buffer/mixed configurations check50,880 histories. Earlier
incremental/CCof coverage remains1,536 configurations/176,640 assignments,
including initialization, huge bounds, reset intersections and solver limits.

For sparse abstraction,32 graphs at one through five frames compare all480
native rows. This complements exact dependency ordering and packed-ID boundaries.
For deeper history, a twelve-register pipeline forces a twelve-cell PI row.
All source storage and object records match through sixteen horizons;512
independent traces check both symbolic initial-state and later PI feedback.

Specialized prefix tests retain96 generated unmodified-source results and624
independent input histories, plus192 mixed prefixes/768 histories. The reference
requires every completed property to be zero and outputs to equal the state
after that prefix. It allows conservative early stopping on a nonzero hash
literal, preserving source behavior rather than substituting SAT. Five directed
native cases verify no PI, no PO, full zero-property bounds and terminating
unbounded events. The final attempted PI frame remains present even when it
cannot affect preceding-state outputs, as in C.

Scalar references evaluate original logical gate/port semantics and sequential
state directly; they use no frame history, dependency ranks or copied Value.
Native comparisons additionally enforce exact ordering and scratch/storage
protocols rather than accepting only equivalent Boolean functions.

Reports compare source convergence, per-frame support sizes, cleanup and prefix
banners exactly. Clock values are normalized after asserting timing-column width.
The empty-prefix level average is IEEE NaN; only its C(-nan)/Go(NaN) spelling is
normalized. Other numeric/report fields remain part of the comparison.

## Compactness, ownership and measured cost

The restored scheduler removes full per-frame copy maps, redundant state arrays,
stopped flags and dead remapping helpers. Keep the source flat arena and row
headers rather than adding per-node heap slices or persistent rank/degree
copies. Offset indices use Go int. Skip earlier frame iterations with no tent;
exact native history/results remain unchanged. No new solver or alternate
unrolling algorithm is introduced.

Go retains a small initial reserve so huge requested incremental bounds can
converge cheaply. A load regression found3067 gates stuck in17 buckets because
CI insertions occupied every periodic object-ID resize checkpoint. Check hash
load at completed horizons, preserving source node/history/literal order and
avoiding a large unconditional reserve
(`/tmp/abcgo-frames-stream-hash-red.log`). This is a Go allocation-policy
adaptation; the source C constructor reserves10000 objects initially.

Single-run8/64/512-horizon benchmarks on the same small graph:

| Implementation | Time (us) | Bytes | Allocations |
| --- | --- | --- | --- |
| Previous full-map code | 13.869 /103.176 /866.198 | 9120 /63008 /548128 | 101 /671 /5164 |
| Restored source scheduler before growth repair | 16.946 /146.192 /3973.950 | 11268 /54660 /437381 | 109 /566 /4161 |
| Scheduler after growth repair | 18.960 /113.593 /997.118 | 11588 /56388 /469844 | 110 /568 /4165 |

Logs: `/tmp/abcgo-frames-benchmark-before.log`,
`/tmp/abcgo-frames-benchmark-after.log`,
`/tmp/abcgo-frames-benchmark-growth.log`. Skipping empty tents alone did not
materially improve the slowdown. The growth repair improves the new512-horizon
case from about3.97ms to1.00ms. Longer histories allocate less than the old
full-map substitute; the source constructor has fixed costs and this benchmark
does not show uniform speedup. Its fidelity repairs cannot be replaced by the
old algorithm to improve a timing number.

Input order is explicit: fixed mode precreates all frame PIs followed by symbolic
initial registers; save mode creates symbolic state first, then each PI frame.
The original graph's Value is borrowed mapping scratch and remains reusable by
other passes after stopping. The returned graph owns all its nodes and survives
the scheduler and source. Changed frame interfaces retain source name/spec and
source PO-order/role conventions; original I/O names and node-indexed payloads
are not copied as if their old indices survived. Noninitialized retained
registers keep owned Go reset states; explicit zero initialization overrides
that state cube, as requested by the source parameter.

## Validation

All54 focused timeframe/unroller/CCof tests pass without skips in34.608s
(`/tmp/abcgo-frames-final-complete-focused.log`). Freeze production code, tests
and fixtures, then run uncached full module race tests with a30-minute package
timeout. All130 tested packages pass; three command packages have no tests.
GIA passes in545.769s (`/tmp/abcgo-frames-final-module-race.log`). GIA vet passes
with empty output (`/tmp/abcgo-frames-final-vet.log`). Exact handles11914 and
63357 terminated exit0 before results were recorded in
`../../SEQUENTIAL_AUDIT.md`.

All22 active source routines in this family are reconciled. This family does
not complete the remaining GIA inventory or the original131-package sequential
audit, which remains at2 completed packages.
