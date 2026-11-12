# Exact graph resubstitution review

## Readability and compactness pass

The new graph path depends on `aig.Graph`, uses stable edge IDs and slices, and
does not reinterpret object pointers as SAT variable numbers. Support is
represented as a deterministic structural cut. Truth tables fit in one
`uint64` because the explicit cut limit is at most six leaves. Divisor
search is bounded and deterministic by level, node ID and phase. The existing
simulation-based C path remains separate pending its complete dependency/API
integration.

## Independent bug pass

The pass checked complemented edges, buffer/output fanins, constants, graph
cycles, candidates depending on the root, out-of-support divisors, temporary
hashed candidates, cancellation and MFFC reference restoration. A candidate is
accepted only after exact truth-table matching and strict MFFC improvement;
candidate pairs that would reintroduce the root are rejected. Cleanup after
each root reclaims speculative hash nodes before they can enter the next
divisor set.

Tests include a factorization that removes an AND from a three-input cone,
exhaustive truth preservation on that fixture, 80 seeded random four-input
networks, no-growth checks, invalid parameter handling and pre-cancelled
contexts. The historical C `Sat_SolverDoubleClauses` body remains disabled with
`assert(0)` in the source; this graph algorithm does not rely on that path.
