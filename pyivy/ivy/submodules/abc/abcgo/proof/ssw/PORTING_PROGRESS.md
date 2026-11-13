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
An original-C latch fixture checks dynamic correspondence with both local and
global simulation and aggressive solver-recycling thresholds. Both C runs
retain one latch, one PI and one PO; the Go tests also compare all four-frame
input traces for the fixture. The C probe is in `test_vectors/`.
Another compiled-C fixture covers target preparation and simple pair classes.
Targets use regular output fanins, so two outputs driven by the same node
trigger the source's duplicate-candidate assertion even if their polarities
differ; the Go entry point preserves this behavior.

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
platform-dependent `rand()`. The native signal-correspondence entry point,
like C, does not consult its `TimeLimit` parameter. Broader differential
testing against ABC, especially for partitioned and constrained modes,
remains useful.

Timed signal-filter and rarity simulation/filter loops now use ABC's calling
thread CPU clock on Linux, pinning their goroutine during the timed operation;
other supported Unix targets use process CPU time and Windows uses CRT-style
elapsed time. Signal filtering preserves C's signed whole-second truncation,
zero-budget convention, and frame-boundary checks. CEC-backed APIs have their
own timing contracts and are not covered by these signal-filter tests. Several
reported timing counters still use Go elapsed time rather than C CPU time.

With flop grouping enabled, ABC transfers the grouped AIG's counterexample
directly to the original GIA without undoing the permutation or inserted gap
registers. The compiled `test_vectors/ssw_group_witness.c` fixture produces
an 11-register witness on a one-register original graph. The Go entry point
preserves this source behavior; callers cannot assume such a witness has the
original graph's dimensions or is directly replayable against it.

The signal-pair transfer now has a direct regression for complemented copies,
identical destinations, and ordering. Pair-class construction follows the C
algorithm, including a duplicate class entry for an initial self-pair and its
reuse of the outer pair-loop index during class merges and re-rooting. A compiled
C fixture confirms the duplicate, bounded merge and re-root outputs, and a valid
ordered pair sequence that triggers a C vector bounds assertion after index
reuse. Go panics on that sequence. If C assertions are disabled, out-of-bounds
access has undefined behavior, which the Go port does not emulate.
