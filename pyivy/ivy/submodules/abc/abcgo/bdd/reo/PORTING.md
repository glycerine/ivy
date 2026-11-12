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

Remaining: the optional CUDD-reference comparison requires a core dynamic
heap-reorder API (including SYMM_SIFT), absent at present. Extra_ReorderCudd is
not implemented; Extra_bddReorderTest and Extra_addReorderTest return actual
node counts using the Go REO backend, with the latter following source nonzero-
pattern conversion. Their exact original CUDD benchmark backend is not claimed.
See REVIEW.md for evidence and fidelity boundaries.
