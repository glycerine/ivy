# Cube reachability review — final integration pending

Scope: the 74 functions/accessors in `src/aig/gia/giaEra2.c` (1,963 lines),
implemented in `era_{cube,manager,lists,tree,tfo,support,successors,counterexample}.go`.
This is separate from the completed ordinary explicit word-simulation review of
`giaEra.c`. The full GIA package and 131-package audit remain incomplete.

## Repairs and verified milestones

- Red public-driver regressions prove real XOR/MUX initial violations return
  safe; a separate red test shows a MUX control missing from CI TFOs. Reuse the
  existing ordinary-AIG Boolean lowering helper for cofactoring/copying and
  include MUX control dependencies. Retain the cube algorithm, not a substitute
  solver. Share the lowering helper with enable transforms in
  `hash_mapped_gate.go`.
- Restore C's persistent counterexample SAT lifecycle: one combinational CNF,
  all COs free, CI/CO variable maps, BSAT assumptions for each current/next/output
  query, one-million-conflict per-query budget and model extraction. Remove the
  per-frame CEC graph rebuild/sentinel. Red binding tests establish the missing
  variable-map contract first.
- The native GIA-to-AIG export API deliberately reads raw AND fanins, including
  real encodings; its source/native compatibility tests remain unchanged. Lower
  real gates at the counterexample boundary before calling that legacy API.
  A typed binding regression fails without this normalization. Do not claim a
  general export-family repair from this adaptation.
- Encode the source 20-bit item, 11-bit page and mark in one `uint32` handle.
  Use `uint16` for the supported 14-bit variable index and `uint8` counters,
  which split at 63. On this 64-bit host: handle 12→4 bytes, cube record 48→32,
  tree record 52→20. Preserve all source handle bits, ID masking, mark semantics
  and tree thresholds. Flat Go slices remain; no C page allocator is recreated.

## Test evidence

- `era_cube_native_review_test.go` / `test_vectors/era_cube_review.c`: 160
  actual-C ordinary cases across 0/1/3/5/6 inputs, 1/3/33/65 registers, four
  transition patterns and both miter modes at limit 100. Exact match includes
  results, failing output/frame and every counterexample bit. Validate Go
  witnesses too. Native driver output must not report invalid reconstruction.
- Native relation matrix: 2,187 ternary cube pairs at {0,1,2}, {0,16,32} and
  {15,16,64}; compare equality, disjointness, containment, dash count and unique
  sharp/disjoint variable selection.
- `era_cube_typed_review_test.go`: initial violations, MUX-control TFOs and
  eight phase variants × safe/unsafe × both miter modes × typed/ordinary forms
  with 65 registers. Check returned witnesses and trace cleanup between runs.
- `era_cube_sat_review_test.go`: ordinary/typed CI/CO polarity queries and 270
  current/next/output combinations, independently checked over all 16 Boolean
  input/state assignments. Complete unknown current bits, reject impossible
  transitions explicitly and retain one solver across repeated queries.
- `era_cube_storage_review_test.go`: 65,536 pseudorandom handles plus page/item/
  mark boundaries; setting unused preserves predecessor ID and next handle.
  A 256-minterm machine exercises nested 63-entry tree splits, union counts and
  every duplicate lookup. Highest supported tree variable 16,383 is retained
  and looked up correctly. Semantic handle/tree tests pass before compaction.
- Existing `era_cube_test.go` still covers manager storage, minterms, recycling,
  list containment, tree rebalance, supports, correlated successors and a trace.

Current focused tests pass in 6.120s; the highest-variable check passes in 0.008s.
Focused cube/enable/native-export race tests pass in 11.244s
(`/tmp/abcgo-era2-checkpoint-race.log`); GIA vet passes with empty output
(`/tmp/abcgo-era2-checkpoint-vet.log`). Code/tests/fixtures stayed frozen for
these checks. The most recent full-module race checkpoint remains the earlier
explicit-reachability review. Native tests require Linux/compiler/original ABC
objects and skip without those prerequisites; independent tests remain runnable.

## Completed final source and correctness pass

Read all 74 original definitions and their Go implementations in a separate
inspection pass, including ownership, state-word padding, unused marks, source
sentinels, pointer/ID access, predecessor depth, list/tree counters and split
thresholds, sharp/containment retries, processed/unprocessed collections,
cofactor support/dependencies, scratch restoration, stopping and SAT models.

| Original entries | Inspection and evidence |
| --- | --- |
| 24 handle, bit, object/state and branch accessors | Packed-word/mark properties; ordinary native traces; manager/tree/maximum-variable tests. |
| 2 minterm and 6 relation functions | Native 2,187-pair matrix; all 1,093 zero-to-six-variable ternary minterm patterns; Boolean union oracles. |
| 3 TFO, 2 usage, 2 PO-support and 1 may-fail helper | Native transitions, CI cone tests, typed MUX control repair, correlated successor and interrupted cofactor tests. The source may-fail helper stays structural; the driver proves candidates. |
| 7 manager/create/recycle functions | Normal/wide state cases; packed handles; capacity rejection before access/allocation; full-range page arithmetic. |
| 11 list/tree processing functions | Native relation/driver tests; exhaustive ordered cube triples; mixed-cube rooted-union checks; nested split/duplicate/maximum-variable tests. |
| 8 print/count/depth helpers | Native valid-state reports/cube printing, unused filtering, source traversal read-through, independent union counts and predecessor traces. |
| 2 successor routines, report and driver | Native 160-case baseline; semantic proof/scalar BFS; counter limits; real million-call stop/unwind; native diagnostic/domain tests. |
| 4 counterexample routines | Persistent source BSAT lifecycle, variable binding, 270 scalar transition queries, exact native witnesses, valid independent traces and deadline cleanup. |

The independent union tests enumerate all 19,683 ordered triples of
three-variable ternary cubes, checking every insertion. Sixteen further sequences
seed 192 eight-variable minterms, then add 100 mixed cubes each after nested
splits. Inspect rooted storage independently and compare the exact Boolean set;
there are no lost live records, invalid links, contradictory bits or changed
unions. No union algorithm substitution is needed.

A six-bit counter independently verifies thirteen limits in both miter modes,
post-successor/reserved-entry accounting, first failure at frame 32, and witness
priority at the boundary. A real shared twenty-PI parity example reaches the
million-call limit with only two successor states; verify stopped/no-target
status and restored cofactor functions after unwinding. A red zero-budget
regression repairs stale trace retention before an early return.

Page metadata used a signed overflowing ceil expression at the top of the
supported counter range. A red arithmetic test gives negative pages where 2,048
are needed. Use zero for empty storage and `1+(count-1)/MAX_ITEM_NUM` otherwise.
The minterm slice already benchmarks at zero allocations, so retain it rather
than making an unmeasured representation rewrite.

Native private and driver report tests restore memory, numeric padding, source
completion wording, verbose timing and unique-minterm diagnostics. Report actual
witness output/frame, rather than C's incorrect queue-index labels. A simulated
fifty-minute interval reproduces negative timings with 32-bit accumulators;
64-bit microsecond counters fix that while retaining the source report format.

## Semantic feasibility correction

The source's structural nonconstancy heuristic is insufficient for proof.
`((a&b)|(a&c)) XOR (a&(b|c))` is Boolean zero but survives strashing. As a PO it
causes impossible CEX reconstruction; as an RI it fabricates a don't-care state.
Eight red ordinary/real-gate cases demonstrate false candidates and imprecise
zero/one successors before the repair.

Retain cube/cofactor reachability and use the existing BSAT single-frame CNF for
exact feasibility checks. Bind source CIs fixed by the current cube/recursive
branch. A PO needs a satisfying assignment; an RI can vary only when both
polarities are feasible. Literals that are constants or unresolved CIs need no
SAT call. Checked polarity facts belong to the current pNew literal namespace
and are discarded on graph replacement. Unknown queries stop with no fabricated
target. Reuse the same source solver for CEX reconstruction, preserve Value
scratch during online typed normalization, and release online SAT on all exits.
This repairs the source heuristic using its existing solver; no alternate SAT
or reachability algorithm is introduced.

Additional evidence: 128 generated ordinary/real-gate sequential machines agree
with independent scalar BFS in both miter modes. Validate every witness. Tests
cover separate cached polarity facts, a real million-node graph replacement,
old-cache isolation on reused object numbers, scratch ownership and an expired
proof deadline through a recursive parent. No unproved fact may enter the cache.

An expired deadline during CEX reconstruction previously panicked as if UNSAT.
A red regression separates resource exhaustion: the SAT frame helper returns
false for Unknown, the trace routine returns nil with deferred session cleanup,
and the driver returns -1 with no witness. UNSAT transition reconstruction
remains a correctness error. Successful query callers may ignore the added bool
result; callers handling interruption must check it.

## Domain and deliberate source differences

The supported driver domain is **0..16,384 registers**. Twelve actual-C
zero-register cases prove the old Go <=0 guard was too restrictive; reject only
negative counts. Independently verify pure constant/no-input circuits and
empty-bit witnesses, retaining the Go nil-CI guard where C dereferences NULL.

Keep safe unknown outcomes for C's state-limit/CEX assertion defect. Preserve
source queue/count conventions, split thresholds and parent trace structure.
The completed C report reads a record just past nStas and prints a spurious depth
one; Go retains its bounded zero for no current state. Native comparisons
normalize only that explicitly invalid source lookup, addresses and measured
storage/time quantities. Valid-state report depths/formatting match C.

The original raw-fanin AIG export compatibility API remains unchanged. Real gates
are normalized only at the counterexample CNF boundary. Native tests skip when
Linux/compiler/original object prerequisites are absent; independent tests do not.
Profiling is wall-clock time in Go rather than C CPU time, and memory reports use
actual Go record sizes/capacities instead of allocated C pages.

## Final verification status

All focused cube, enable and native-export regressions pass in 15.683s
(`/tmp/abcgo-era2-final-focused.log`). The source/compactness pass is complete,
but final full-module race tests and GIA vet are still pending. Keep this family
open until those exact gates finish on frozen production/tests/fixtures. GIA and
the 131-package audit remain incomplete; original package completion is 2/131.
