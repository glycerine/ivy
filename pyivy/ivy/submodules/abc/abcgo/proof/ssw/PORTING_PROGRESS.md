# Sequential sweeping port progress

The package builds. The full package race tests and `go vet ./proof/ssw`
pass. Its implementation includes signal and latch correspondence, SAT cone
encoding, simulation, constrained and dynamic sweeping, and rarity simulation.
Sequential trace and counterexample tests exercise these paths; passing tests
are not evidence of exhaustive correctness.

Similarity and pair-driver tests cover equivalent and inequivalent sequential
graphs and deadline handling. Rarity tests cover bin widths 1 through 16,
including word-boundary crossings, and replay multi-round witnesses. A
connected-components oracle checks candidate pair classes across 500 graphs.

ABC's `Aig_ManReprStart` explicitly asserts `Aig_ManBufNum(p) == 0` in
`src/aig/aig/aigRepr.c:47`. The Go `aig.Aig_ManReprStart` preserves this
precondition, and `Ssw_ClassesStart` calls it if the graph does not already
have representative storage. Consequently, signal correspondence on such a
buffered graph panics when it needs to create the storage. Both this public
entry point and the class-manager precondition have regression tests. The
frame and SAT helpers separately support buffer nodes; those tests do not
imply that representative-based correspondence accepts buffered graphs.
Lifting this precondition would require more than removing the assertion:
class enumeration, graph-copy ownership, and transfers back to the original
graph must all preserve representatives and phase through the buffers.

The original semi-formal manager contains an unconditional assertion because
its frame-array migration is unfinished; the port preserves it as a panic.
Rarity permutation uses Go's PRNG, so its exact ordering differs from C's
platform-dependent `rand()`. Deadline checks are cooperative, and the legacy
CEC delegation returns unknown when a deadline is required because that API
does not provide one. Broader differential testing against ABC, especially for
partitioned and constrained modes, remains useful.
