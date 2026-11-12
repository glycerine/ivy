# Complete prime-implicant recursion review

Exact `extraZddPrimes` algorithm from `src/bdd/extrab/extraBddMisc.c`, exposed as
an importable leaf for ABC's all-primes BDD-to-SOP conversion. The root forwards
through `extraBddPrimes.go`; only those two completed draft functions were removed
from `extraBddMisc.go`, with comments directing readers to the implementation.
All remaining misc source drafts remain visible.

## Readability and compactness

One memoized cofactor recursion retains the source's increasing/decreasing-unate
optimizations and general intersection-prime case. Positive and negative literal
extensions use canonical core ZDD change/union rather than duplicating unique
node insertion. No alternative graph representation or duplicate algorithm exists.
Constants and the complemented-edge normalization are explicit. Source attribution
and reference ownership contract appear at the exported API.

## Bug review

Validates BDD family/ownership before recursion. Correctly propagates complement
phase to both cofactors. False yields the empty family; true yields the empty cube.
Literal indices are exactly 2*i positive, 2*i+1 negative. Set difference removes
only primes already supplied by the intersection, as in C. Generic canonical
composition supports independent BDD/ZDD orders rather than assuming identity
index order. Results are unreferenced and callers must protect retained roots
across subsequent reordering, as with CUDD APIs.

All 256 three-variable functions match an independent exhaustive ternary-cube
prime test (implication and removal of every literal), and reconstruct the exact
BDD. Original-C membership fixture checksum `5394c07cdac40638` matches all 27
consistent cubes for every function. An additional 128 four-variable seeded
truth tables test shuffled BDD order and independently permuted ZDD order.
Invalid ownership/family and constants have regressions. Race tests and vet pass.

This package is complete; the surrounding root `extrab` package is still partial.
