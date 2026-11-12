# AMAP permutation and rule review

## Implementation

`amapPerm.go` now collects associative AND/XOR frontiers, recursively matches a
DSD decomposition to a library graph, and constructs the reverse pin mapping.
The DSD matcher uses operation types and local backtracking state; it does not
mutate shared DSD marks while matching. `amapRule.go` builds AND/XOR rules from
all non-mirror subset partitions, emits PRIME mux rules, and attaches successful
gate permutations to library nodes. Rule vectors preserve uniqueness through
the existing packed lookup tables.

## Focused verification

The explicit source-file race gate for `amapPerm.go`, `amapRule.go`, and their
native test passes. `TestLibraryFaninCollectionAndDSDPermutation` checks that a
two-input AND maps both distinct DSD variables onto the generic Amap leaf's
two input positions. `TestRuleConstructionDeduplicatesCommutativeAnd` checks
that opposite fanin order reuses the same AND node.

The DSD metadata helpers `Kit_DsdNtkVarNum` and `Kit_DsdNtkRootLit` were added
because those fields are intentionally private to `bool/kit`.

## Review notes

- DSD child assignment backtracks over candidate frontiers and stores the
  result in local state, avoiding cross-call `fMark` corruption.
- The recursive rule enumerator removes mirrored subset partitions. Large
  prime/decomposition fanin counts can still cause exponential work, as in the
  source algorithm; the supported DSD arity is capped by Kit.
- Symmetry handling emits pin swaps only for the source detector's simple
  complementary variable pairs. More general automorphism groups are not
  inferred here.
- The full `go test ./map/amap`, race, and vet gates now pass. A public API
  end-to-end test maps AND, XOR, and MUX functions using the native library;
  the mux parity setting and C output are documented in `test_vectors`.
