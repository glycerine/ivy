# Conflict-analysis circuit SAT (giaCSat.c)

`giaCSat.go` retains the original CBS engine: assignment trail, relative reason
indices, justification frontier, circuit propagation, clause derivation and
resolution, clause-directed backtracking, three decision heuristics, strict
backtrack/frontier limits, and CI or all-object cube witnesses. This is distinct
from the chronological CBS0 and watched-clause/fanout CBS2 engines.

The sequential audit compares 63,648 queries with the actual C implementation:
verdicts, per-query and cumulative counters, and ordered witnesses agree across
random and hard reconvergent/parity graphs. All three outcomes occur, with up to
27 backtracks and frontier size nine. Independently enumerating every five-input
assignment verifies UNSAT proofs and every completion of every returned cube.
Another 7,200 native miter properties check statuses, counterexample stores,
logical CI/CO permutations, phase initialization and proof-zero output patches.
Direct reason snapshots check relative offsets, pivots, literal order, restored
marks and both clause reports. Fake-clock and zero-call statistics match C.

Regression-backed repairs restore constant-output handling before miter budget
checks, phase initialization, source statistics and clause reports, and all idle
queue sentinels. Both roots are validated before either is assigned, preventing
invalid paired queries from changing the graph. Lifecycle tests cover reuse,
partial cancellation, nested saved frontiers across queue growth, unsupported
typed graph mutation and active/repeated stop cleanup.

Raw Go AND graphs can retain constant fanins that normal C hashing simplifies.
A regression exposed a false SAT result when constant zero was assigned as a
variable. Fanin predicates now treat constants as fixed values and conflict
analysis omits fixed axioms. A separate 52,992-query exhaustive matrix covers raw
constants, buffers, shared/opposite fanins, phases and paired roots, checking
verdicts, cube consistency/implication and scratch cleanup independently.

Compactness: restore the source's reusable conflict mark scratch and use a fixed
two-root array. Hard single and paired UNSAT queries now allocate zero bytes per
solve, versus 37 and 21 allocations previously. Final benchmark measurements are
approximately 8.3 and 5.0 microseconds; timing is not a correctness guarantee.
The queues retain logical heads/tails because suspended frontiers occupy inactive
backing storage. Unsupported typed-gate checks use maintained graph counters.

The combined focused CBS/CBS0 suite passes. Full module race integration and GIA
vet are recorded in SEQUENTIAL_AUDIT.md. This review covers CBS only; the rest of
the GIA audit and the sequential 131-package objective remain open.
