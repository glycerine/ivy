# Explicit reachability review

Scope: all 17 functions/accessors in `src/aig/gia/giaEra.c`, implemented in
`era_reach.go`. This does not cover the different cube algorithm in `giaEra2.c`
or the general STG reader/encoder family.

## Correctness repairs

- Real XOR and MUX nodes were simulated as AND, producing false safe results.
  A red public-driver regression demonstrated both failures. Use the existing
  typed Boolean word evaluator; preserve the ordinary AND/buffer path.
- A borrowed simulation row exposed subsequent objects through its slice
  capacity. A red append test overwrote a neighboring PI word with `deadbeef`.
  Bound capacity to the row while preserving in-place edits.
- STG export printed a spurious zero for zero-width binary columns. Six red
  cases failed before the repair. Omit binary digits at width zero, matching
  `Extra_PrintBinary` in C.
- Progress reports omitted the reached-state ratio, memory, source padding and
  a newline; the summary also lost time padding. A red native 8,192-state counter
  comparison exercised the actual 5,000-state progress branch. Restore the
  source report fields and layout. Memory estimates use Go record sizes and
  actual slice capacity; elapsed wall time and memory are inherently different
  from C's CPU time and fixed-pool allocation.

## Evidence

- `era_native_review_test.go` / `test_vectors/era_review.c`: 480 actual-C
  configurations, 3,453 matching algorithmic records. Cover 0/1/3/5/6 inputs,
  1/3/33/65 registers, four transition patterns, state limits 2/4/100 and both
  miter modes. Compare public results, private states and words, predecessor
  conditions, scratch identity, depth, hash lookups, traces and all transition
  rows. Three initial hash buckets force automatic resizing.
- `era_scalar_review_test.go`: 100,680 independently evaluated input/state
  combinations over 0/1/5/6/12 inputs, eight typed phase patterns, 65 registers
  and three packed states. Compare all 69 COs, including high register bits,
  PI minterms across simulation words and complemented outputs. Separate
  repeated resizing checks preserve 40 distinct multiword records and exclude
  pending scratch storage.
- `era_traversal_review_test.go`: independent scalar BFS for 16 typed sequential
  machines/miter modes. Compare discovery order, packed states, every transition,
  output masks, public results and first failing predecessor traces. Three
  changing registers straddle bits 0/32/64. A separate case checks that PO index
  takes priority over the earliest input condition on later POs, and that a
  violating state does not collect successors.
- `era_typed_review_test.go`: 24 initial-state cases across one/six inputs and
  one/33/65 registers; typed gates, buffers, opposite-fanin AND, and earliest
  violation in the second simulation word.
- `era_stg_review_test.go`: eight hold/toggle graphs across zero/one PI and PO;
  exact row expectations and byte-for-byte actual-C file comparison.
- `era_reports_review_test.go`: native counter report comparison, including
  exact fields, ordering, padding and newlines; normalize only elapsed time and
  representation-dependent memory amounts.
- Existing `era_reach_test.go` continues to check ordinary results, state limits,
  internal transition storage and a multi-step failure trace.

Native tests compile and execute the actual source, with a zero clock for stable
native reports. They skip when Linux/compiler/original ABC object prerequisites
are unavailable. Scalar and storage regressions do not depend on the native
build. All tests run sequentially when capturing stdout or changing directories.

## Separate inspection and compactness pass

Inspect manager initialization and ownership, 32-bit PI truth patterns, register
broadcasting, complement handling, hash overflow and collision chains, pending
scratch reuse, resizing, predecessor traces, depth, queue/limit ordering and STG
row construction against source. No alternate reachability algorithm is added.
The scalar BFS is an independent test oracle only.

The Go hash resizer traverses allocated records rather than C bucket chains;
lookup membership, IDs and discovery order remain unchanged. The pending scratch
record stays uninserted. Go initializes allocated records and all state words;
C leaves pending condition fields unspecified and updates each defined bit.
Native comparisons exclude only that unspecified scratch condition. The safe Go
empty-manager depth result remains zero; C assumes at least the initial state.

Keep the existing compact slice-based manager and one state record plus packed
word slice per discovered state. Avoid recreating the C fixed allocator or
introducing redundant count/capacity fields. Reuse the existing typed gate
simulation helper. Bounded borrowed rows prevent corruption without copying.
No speculative representation rewrite is justified by this review.

Input domain: 0..12 PIs and at least one register. Source miter priority is PO
order followed by input-minterm order. Traces retain the source's initial dummy
condition. Queue limits count the nil entry and any pending scratch record,
exactly as C; do not reinterpret them as a pure discovered-state count. Driver
reports retain source depth/count conventions. STG output masks contain at most
32 PO bits; general zero-width STG parsing and wider-output encoding are separate
source-format questions.

The original C driver can assert while dumping a *partial* table after a state
limit: it supplies `iCurState-1` as the writer's state bound even when rows target
unprocessed states. The Go driver writes those rows without that source assertion.
Preserve this existing safety difference; the native file oracle uses complete
valid tables. The Go driver deliberately keeps its own row formatter instead of
calling the general printer, whose validation rejects zero-width columns and
partial traversal state bounds.

Final validation: focused explicit-reachability tests pass in 6.174s
(`/tmp/abcgo-era-final-focused.log`). Full uncached module race tests pass all
130 tested packages, with three no-test command packages; GIA takes 447.395s
(`/tmp/abcgo-era-final-module-race.log`). GIA vet passes with empty output
(`/tmp/abcgo-era-final-vet.log`). Production, tests and fixtures stayed frozen
throughout both integration gates. Mark this source family reviewed. The full
GIA package and sequential 131-package audit remain incomplete.
