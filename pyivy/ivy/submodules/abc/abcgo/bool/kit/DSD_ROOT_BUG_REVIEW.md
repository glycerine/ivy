# Separate DSD root-object bug review

Inspected `Kit_DsdExpand` and `Kit_DsdShrink` against the original C wrappers.
The Go generic recursion collapsed a VAR root object into a leaf literal. Truth
values remained correct, but root-object consumers then received nil from
`Kit_DsdNtkRoot`. The original wrappers explicitly clone CONST1/VAR root objects.

Added `TestReviewTrivialDsdRootPreserved` and observed it fail on truth
`aaaaaaaa` before the fix. Both wrappers now share a small helper that preserves
the trivial root type, fanin literal and root phase. Shrink handles these roots
before consulting priorities, matching the C early-return order. The test checks
constant zero/one and positive/negative variables, root-object shape, phase and
reconstructed truth; trivial shrink also works with no priority array.

This review also examined graph edge phases, bridge scratch copies, SOP arena
ownership/failure state, cache eviction and lifecycle generations. No further
unresolved behavioral defect was identified in those inspected paths. The
independent path-inference and Cloud scratch-shape bugs were reproduced and
fixed by another reviewer, documented in `BUG_REVIEW.md`; source ownership was
coordinated to avoid overlapping edits.

The full Kit suite, combined consumer race tests and `go vet ./bool/kit` pass.
