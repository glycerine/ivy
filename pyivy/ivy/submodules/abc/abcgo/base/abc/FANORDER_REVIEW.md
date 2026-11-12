# Fanin and SOP ordering

## Readability and compactness pass

Ported the complete abcFanOrder.c API group in abcFanOrder.go. Typed SOP rows
replace mutation of shared C character buffers. A common permutation routine
reorders both edge objects (including complement bits) and SOP columns. Source
selection-sort tie behavior is retained explicitly. Transpose and sorting
scratch vectors are typed Go slices; containment/distance-one normalization
remains the original iterative algorithm.

The adjacent network_api.go adds source inline network predicates/selectors and
latch initialization operations required by FRET consumers. Latch data remains
the original enum, with nil representing an uninitialized fresh latch.

## Separate bug and fidelity pass

Tests exhaust all 16 input assignments before and after every fanin ordering,
including complemented edges and off-set covers. Transpose permutation ordering
is checked directly. Actual split DAGs are evaluated for both on-set and off-set
covers, and reciprocal graph invariants are checked after splitting/reordering.
100 seeded random covers exercise normalization, truth preservation, containment
and distance-one elimination, plus idempotence. Latch tests exercise every state
transition/predicate and mixed-box classification.

A source bug would combine split off-set covers with OR, changing !(A|B) into
!A|!B. The Go port uses AND for negative subcovers, preserving the original
function. Single-cube nodes are not split even if an impossible zero cube limit
is supplied. Go edge permutations retain complement bits rather than separating
them from the moved fanin. Scratch strings are immutable, so splitting cannot
truncate another node's shared cover storage.

Selected previously ported core + DFS + fan-order/API tests pass race tests and
vet. Full base/abc next fails in untranslated abcFunc.go; package completion is
not claimed. No C oracle differential coverage is claimed for this group.
