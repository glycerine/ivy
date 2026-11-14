# Fast, random sequential, and phase simulation

This family replaces `saigSimFast.c`, `saigSimSeq.c`, and `saigPhase.c` with native Go algorithms. It owns only the corresponding three Go files and focused tests. Other SAIG families may still contain drafts; `test_simulation.sh` is explicitly a file-scoped gate.

## Implemented behavior

Fast simulation remaps graph IDs to compact CI/node/CO order, simulates 32 packed patterns, transfers registers simultaneously, and computes one-probability or 16-lane transition activity. All live internal nodes, including dangling nodes, are retained. The source float-in-int-vector interface is represented with `math.Float32bits`, without unsafe conversion. The source's previously empty `Saig_ManComputeSwitchProb3s` now delegates to the working engine. The standalone API takes explicit frame/prefix arguments; the old application-global `seqsimframes` flag lookup is replaced by those explicit arguments.

Random sequential simulation orders nodes by DFS, consumes one reference per fanin, and recycles simulation blocks when the last user finishes. Only the live cut and retained register inputs survive a frame. It supports ordinary AND, explicit XOR, buffers, constants, unused inputs, and sparse IDs. Random-number replay reconstructs a multi-frame counterexample; independent scalar simulation validates every generated witness before attaching it to the graph. Zero time limit means unlimited time, and a positive limit applies to the complete run.

Phase abstraction records packed ternary states until recurrence, with the source's 3,000-frame retirement threshold and 10,000-frame maximum. It computes prefix/cycle lengths, stable binary candidates and transient statistics, identifies periodic registers, substitutes their phase constants while unrolling, and cleans unused sequential state. Counterexample translation restores original microframe/output coordinates. The automatic API returns an ordinary duplicate when abstraction is unsuitable, as in the source.

## Compactness and readability pass

Fast simulation derives object/node/output sizes from slices instead of duplicating capacity and offset fields. One original-ID slice handles sparse/reordered graphs. Random simulation owns typed objects and a slice pool/free list; live-block count derives from pool/free lengths. Removed a duplicate CI-ID array and unused live counter. Ternary states are owned `[][]uint32` snapshots with a map of immutable keys; intrusive C hash-chain pointers and their private accessors are unnecessary. State insertion uses the newly created state directly in its common path. Existing graph constructors, ternary operators, cleanup and vector APIs are reused rather than reimplemented.

Manager-local scratch state avoids mutating graph traversal, marks, data, or ternary fields during these algorithms. The source-shaped explicit Xsim helper APIs remain manager-aware wrappers around AIG's typed node accessors. Graphs/managers are not intended to be mutated concurrently with simulation.

## Separate bug review

Reviewed constant polarity, complemented fanins, simultaneous latch transfer, reference recycling order, unused inputs, random replay ordering, sparse IDs, recurrence indices, retirement thresholds, constraint ordering and counterexample bounds. Regressions cover:

- Odd numbers of measured frames use C's integer-divided word count when
  normalizing transition activity. This source behavior can yield values above
  one; a single measured frame has a zero denominator and yields NaN for 0/0.
- Simulation count overflow is rejected before accumulating into source-shaped int32 counters.
- Random simulation allocates a result before dereferencing operands, preventing reuse while an operand is still being read. Non-miter POs with no users are immediately recycled instead of violating the source's reference assertion.
- The top-level random simulator clears stale witnesses; constant-one failures receive an independently checked witness rather than merely returning success.
- Runtime limit applies to the entire run, repairing the source's per-frame/zero-limit behavior.
- Transient last-value/last-step counters reset for every register, repairing cross-register carry in the C implementation.
- Phase abstraction groups expanded constraints after property outputs and preserves constraint metadata. Counterexample translation uses the corresponding property/constraint coordinate mapping.
- Malformed/truncated phase counterexamples are rejected; expanded frame arithmetic is checked before allocation.

The multi-frame graph witness API was added in `aig/counterexample.go` with the AIG owner's permission. It validates dimensions, frame/output indices and bit counts, copies initial-state and input bits, and preserves the prior model on failure. AIG's existing one-frame storage is shared, matching original `pSeqModel` ownership. The utility's existing `Abc_CexDup` intentionally omits initial-state bits, so the setter performs an explicit complete slice copy.

## Tests and C evidence

`test_simulation.sh -race -cover` passes (approximately 74% of the three implementation files). Tests include a long live-cut pool, delayed bug/witness replay, constant/unused-input cases, periodic register elimination, all 256 eight-input grouped traces, long-counter retirement and constraint/witness coordinates.

Reference tests compile unmodified functions extracted from original C source with small drivers in `test_vectors/`:

- Exact packed one-counts and transition-counts against `Faig_ManSimulateFrames`.
- Exact next-state words, failure object and failing pattern against `Raig_ManSimulateRound` and its original reference-counted pool.
- Full ternary operator tables and a 257-word hash against original phase routines.

A separate test invokes the actual repository ABC binary with `phase -F 2`, then compares latch count and exhaustive grouped traces against the native abstraction. C compilation/binary tests skip only when their respective tool/source/executable is unavailable. Full `go test -race ./aig/aig` also passes for the witness API addition. No tests touch experiments.
