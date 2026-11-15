# Cube reachability review — in progress

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

## Remaining work before closing the family

- Complete the source-by-source pass for all 74 entries, including storage
  exhaustion, list/tree union invariants, compression/sharping/containment,
  recursion/resource stops, cofactor scratch lifecycle and reporting.
- Add independent exhaustive/random cube-set union checks beyond full-minterm
  tree membership, including mixed don't-care cubes, unused cubes and retries.
- Verify recursive stopping and driver limits independently. The original C
  driver can call CEX reconstruction on a state-limit stop without a failing
  output/target and assert. Preserve the existing Go distinction between a
  resource stop and a genuine witness; do not reproduce that source defect.
- Reports omit source fields/layout and need red native comparisons. Inspect
  diagnostic-only traversals and statistics without changing cube semantics.
- Finish a separate correctness/compactness pass and rerun full uncached module
  race tests plus GIA vet on a frozen final family checkpoint.

Source boundaries: normal driver circuits have 1..16,384 registers. Counterexample
models can differ in general SAT tie cases, so independently validate witnesses;
the tracked ordinary matrix happens to match every native bit exactly. Existing
Go errors for impossible/unknown trace queries remain explicit instead of C's
print-and-continue path, which can leave stale model data. No cube-family closure
or package completion is claimed yet.
