# Register partitioning review

Ported original aigPartReg.c to typed []int32/[][]int32 partitions and Graph/Edge
models. The smart heuristic retains largest-unused-support seeding, frontier
selection, the 1+3*already-used cost, support closure and bounded partition size.
Simple partitioning retains overlap and redundant-tail suppression. Linear
partitioning now returns actual traversal-ordered partitions rather than the
original diagnostic-only NULL result; otherwise-unused registers are included.

CreatePart preserves the original interface ordering: external combinational
support inputs, then selected register outputs; fake observation outputs precede
selected next-state outputs. Selected registers remain trailing CI/CO pairs.
Count outputs distinguish true PIs and external register inputs. Reverse maps
translate partition IDs to original IDs, and onehot projection uses source copy
maps and current traversal marks. Real equivalent-latch representative transfer
is regression-tested against the shared Repr implementation.

## Readability/compactness pass

Use native slices and membership sets instead of generic pointer/int/float
vectors. Slice lengths carry sizes; no capacity mirrors remain. Removed unused
cost-history, graph-owner and result-vector fields from heuristic state. Shared
combinational supports are memoized once per graph cone. Duplication and traversal
helpers reuse core edge-phase logic and bounded traversal-ID rollover.

## Separate bug pass

Fixed whole-domain loss when a domain is smaller than overlap: always retain the
first partition, while still suppressing redundant overlapping tails. Reject zero
or negative partition sizes, negative overlap and duplicate/out-of-range selected
registers. Scl repairs its original zero-size call site to explicitly request one
whole-register partition. Reverse maps keep constant ID zero and deterministically
select among coalesced source IDs, avoiding Go map iteration dependence. Traversal
wrap clears stale marks before onehot projection. Regressions cover these cases,
partial support interfaces, all-assignment next-state parity, class transfer and
linear coverage.

Focused actual-core partition/MV tests pass with -race. This is not a claim that
all unrelated aig draft files or all SAIG source files have been ported.
