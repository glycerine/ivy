# ADD abstraction, extrema, bit extraction and reciprocal review

Scope: the original `cuddAddAbs.c`, `cuddAddFind.c`, and `cuddAddInv.c` families.
Their full public algorithms live in core; the root files only forward. Original
source files and original copyright/license headers remain visible.

## Readability and compactness pass

- Sum/product/OR abstraction share level-aligned recursion and memoization, with
  explicit mode-specific base cases and arithmetic. This retains direct
  correspondence to the three source recursions without triplicating traversal.
- Preserved the source's absent-variable behavior: sum doubles, product squares,
  OR skips. Abstraction cubes follow the source's positive-literal ADD convention.
- Extrema share one then-first walk while retaining their distinct comparisons
  and infinity short circuits. Sorting/scanning terminal values would change NaN
  behavior and was avoided.
- Reciprocal and ADD-valued bit extraction reuse the existing authoritative
  monadic apply traversal. Integer-bit interpretation is shared with the prior
  BDD-valued bit extraction, avoiding divergent boundary rules.
- Root live internal recursive entrypoints forward to validated public core APIs;
  C-private helpers with no external callers are replaced, not retained as drafts.

## Bug review

- Checked constant and missing-variable abstraction, skipped levels, all cube
  subsets, both numeric signs, zeros, memo keys, and shuffled manager order.
- Positive cube validation rejects BDD cubes, negative literals, non-Boolean
  constants, zero, and foreign-manager roots before traversal.
- Reciprocal uses `abs(leaf) < epsilon` exactly: equality succeeds; epsilon must
  be constant; zero/negative/NaN thresholds preserve original semantics. A zero
  threshold permits zero reciprocal, producing IEEE infinity in this core.
- Extrema preserve then-before-else NaN selection and the source's early return
  on positive/negative infinity, instead of imposing a new floating-point order.
- Integer bits truncate toward zero to signed 32-bit values, including the sign
  bit. Invalid bit indices, nonfinite values, and unrepresentable truncated
  integers fail with nil, consistent with the bridge family's documented choice
  for cases undefined in C.
- The original prebuilt C objects use non-IEEE `1e302` infinity sentinels. The
  oracle initially differed only on reciprocal zero. Rebuilding the genuine
  original constant table and initialization with `HAVE_IEEE_754` matches the
  Go core's already established IEEE representation exactly; no Go arithmetic
  workaround or changed expected mathematical result was introduced.

## Evidence

`cuddAddNumeric_test.go` independently enumerates numeric sums/products/Boolean
OR for 32 seeded functions × all 16 cube subsets × all 16 assignments, with an
absent fourth input. Additional tests cover bad cubes, reciprocal bounds,
constant zero, extrema/NaN/infinity order, bit conversion consistency, and
shuffled levels.

Original-C fixture `test_vectors/add_numeric_oracle.c` hashes all 16 quantified
subsets (sum/product/OR), extrema, four reciprocal thresholds, and six bit
extractions over 16 assignments. IEEE-configured checksum: `dab9cf682104fafe`.
The fixture's C regeneration is optional; Go tests require no native toolchain.

Validation: full core race tests and vet, plus root subset forwarder race tests.
Root CUDD remains independently incomplete (`cuddAPI.go` and later drafts).
