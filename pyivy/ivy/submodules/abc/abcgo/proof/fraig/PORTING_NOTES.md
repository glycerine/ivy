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
randomized functional reduction, and external clauses involving internal gates.

The port also repairs C-derived edge cases in complemented equality, multi-output
miter checking, simulation of constant nodes, feedback pattern compaction, and
hashing of dynamic simulation words beyond the original fixed prime table.
External clauses now allocate SAT variables, encode internal nodes, simplify
constant literals, and keep their constraint cones active for subsequent
queries. The msat solver retains root unit facts when preparing a new cone.

Remaining validation work:

- Compare representative real ABC workloads and their FRAIG results with the C
  implementation, including solver feedback and sparse functional reduction.
- Exercise SAT timeout and unknown outcomes during real search, and assess limits
  on deep cone traversal and large feedback/simulation workloads.
- Test difficult interactions among choice classes, sparse hash collisions, and
  compression, including repeated compression over multiple word boundaries.
- Integrate FRAIG with a ported ABC consumer and verify end-to-end results; the
  separate-package test exercises the public API but is not such a consumer.
