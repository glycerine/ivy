# FRAIG port status

The FRAIG package builds and its package tests pass under the Go race detector.
It ports the original C manager, simulation, structural and functional hashing,
SAT-based reduction using `sat/msat`, solver feedback, choice handling, and
vector and fanout utilities. C tagged pointers are represented by paired Go edge
handles, while simulation and vector storage use slices.

The tests include public API use from a separate Go package, exhaustive MUX
clause checks, SAT comparisons against exhaustive truth tables for small circuits,
normal and sparse feedback compression, pattern overlap and partial-word cases,
resource bounds, and choice and fanout handling. Additional exhaustive tests
check sparse SAT reduction, choice-cycle prevention, miter proof polarity,
randomized functional reduction, and source-specific external-clause behavior.

The source APIs deliberately retain several C limitations: distinct edge
arguments to `Fraig_NodesAreEqual` are compared after removing complements;
`Fraig_ManCheckMiter` returns immediately when an output is undecided and does
not validate simulation witnesses against external SAT clauses. `Fraig_ManAddClause`
adds literals to the existing solver without creating variables or encoding gate
cones, so callers must prepare those first. The C SAT code does not constrain
its constant node to true. Tests cover these behaviors; they are not guarantees
of logical equivalence or constrained-miter correctness in those cases.

Other source limitations are preserved: bit simulation, TFI marking, and
supergate helpers assume a nonconstant root; feedback compression does not clear
preexisting bits in a partial word; and the simulation hash uses a fixed prime
table without wrapping. Go bounds checks turn an overrun of that table into a
panic. As in C, MSAT preparation clears root unit assignments. As in the C API, adding an empty or
contradictory external clause triggers an assertion rather than representing an
inconsistent manager.

Remaining validation work:

- Compare representative real ABC workloads and their FRAIG results with the C
  implementation, including solver feedback and sparse functional reduction.
- Exercise SAT timeout and unknown outcomes during real search, and assess limits
  on deep cone traversal and large feedback/simulation workloads.
- Test difficult interactions among choice classes, sparse hash collisions, and
  compression, including repeated compression over multiple word boundaries.
- Integrate FRAIG with a ported ABC consumer and verify end-to-end results; the
  separate-package test exercises the public API but is not such a consumer.

The 107 unique FRAIG APIs declared in `fraig.h` have 105 C implementations
and corresponding Go functions. `Fraig_DfsReverse` and `Fraig_CountLevels`
are declared in the header but have no implementations in the C checkout;
they are not supplied as invented replacements.
