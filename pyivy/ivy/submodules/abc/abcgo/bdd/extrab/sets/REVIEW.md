# ZDD set-family review

Complete `extraBddSet.c` family plus the required `extraZddMaximal` recursion from
`extraBddMaxMin.c`, Alan Mishchenko. The original C files are unchanged and remain
visible. Root functions forward to this sole implementation; the remaining MaxMin
family is now completed in maxmin.go (review: MAXMIN_REVIEW.md).

## Readability and compactness pass

Four relation filters share the original terminal and top-level comparison cases
through explicit subset/superset and positive/negative modes. This removes the
source's repeated refcount/error/cache scaffolding without replacing recursion
with explicit family enumeration. Memo keys include operation kind and both
operands. The fused maximal-not-superset algorithm retains its separate branch
pruning: remove lower-branch members subsumed by upper-branch members before
recomposing. The maximal helper uses the same source recurrence and filter cache.

Canonical composition delegates to core ZDD change/union rather than defining a
second unique table or manually inserting nodes. It preserves zero suppression
and sharing under independent variable orders. Membership predicates are short
else-/then-chain traversals, with comments documenting actual source behavior.
All maps are local to the call; no process-global pointer caches survive manager
lifetimes.

## Bug review

Validated ownership and ZDD family before applying terminal shortcuts, so a
foreign or wrong-family root cannot pass an identity/zero case accidentally.
Checked all five relation modes, empty family vs empty set, inclusion equality,
skipped levels on either operand, variable permutation, maximal antichains, and
canonical zero-suppression. The original `Extra_zddIsOneSubset` synopsis says an
empty family succeeds, but its actual implementation returns false; the port
preserves the actual function and has an explicit regression. Similarly the
maximal-filter synopsis misstates the direction; actual semantics are maximal
members of X that contain no member of Y, matching its recursion and native oracle.

Tests exhaust all 256×256 three-element family pairs under identity and a
nonidentity ZDD order, comparing every operation against ordinary bit-set
inclusion logic. Predicate and maximal-family results are also exhaustive.
An additional 1,024 four-element seeded family pairs verify fused filtering and
positive/negative filter partitions. Invalid nil, manager ownership, BDD/ADD
families and non-Boolean numeric constants have regressions.

Genuine original-C fixture `test_vectors/set_oracle.c` hashes all identity-order
pairs and all original predicates/maximal outputs: `697c168ae20d22d5`. The permanent
Go test uses this checksum without needing C at test time. Independent semantic
expectations additionally prevent matching a shared source mistake accidentally.

Race tests and vet pass for this package and the explicit root forwarder subset.
This leaf and the MaxMin family are complete; root extrab still has other
incomplete families.
