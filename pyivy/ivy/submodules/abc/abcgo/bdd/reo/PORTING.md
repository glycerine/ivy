# REO port status

The package now builds and its BDD reordering engine is implemented in Go using
`github.com/glycerine/abcgo/bdd/cudd/core` as the authoritative canonical graph
implementation. All generated REO algorithm files have been replaced with active Go
implementations; original algorithm sources remain in `src/bdd/reo`.

Implemented: reusable manager and original API settings, support/permutation
mapping, shared-root node/APL/frontier profiles, adjacent permutation swaps,
cost-bounded sifting, remapping to topmost support slots, shuffle, verification,
reference ownership, explicit complemented unit transfers, and slice-backed
stable-address unit storage. Compatibility array arguments use slices and int
lengths. No C runtime or unsafe pointer manipulation.

Validation and separate design/correctness reviews are in REVIEW.md.

ADD/MTBDD reordering now uses authoritative core ADD ITE/permutation and generic
support APIs, with family-preserving transfers and numeric-terminal regressions.
BDD and ADD node/APL/width objectives and mixed root arrays are covered.

The classical reference comparison now uses an independent Go CUDD manager with
actual SYMM_SIFT dynamic reordering, preserving referenced handles. The BDD/ADD
classic-sifting benchmarks use the source backend and node-count semantics,
including ADD nonzero-pattern conversion. Main and comparison results are
checked by inverse permutations and exhaustive input assignments. See REVIEW.md
and core/REORDER_REVIEW.md for independent original-C parity evidence and the
bounds of supported CUDD facilities outside this package.
