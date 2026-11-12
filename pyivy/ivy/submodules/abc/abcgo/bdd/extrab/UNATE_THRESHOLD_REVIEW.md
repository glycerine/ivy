# Unateness and threshold source-family review

Completed `extraBddUnate.c` and `extraBddThresh.c` algorithms as real Go in the
root package. Shared graph types alias the sole CUDD core; exported Go info fields
replace C bitfields and fixed scratch buffers. `extraBdd.go` is now a valid header
translation. Image-tree types live beside their completed implementation in
`extraBddImage.go`.

## Readability and compactness pass

Unateness retains original cofactor/intersection recursion and local implication
checks, with local memo maps. Structure creation, slow cofactor checks, ZDD family
creation and printing remain distinct APIs. Core ZDD operators perform canonical
composition; no duplicate node representation exists.

Threshold search replaces six nested-loop copies with the same recursively
ordered bounded weight enumeration (original limits 3,4,5,9,13,9). Chow grouping,
truth-variable sorting, inequality creation/simplification, incremental weight
adjustment and final certificate checks are separate routines. Truth tables are
64-bit slices and scalar/count/weight APIs use Go int. Comments explain important
source side effects and the replacement of the global ABC frame dependency.

The heuristic requests **all** prime implicants of positive-unate on/off sets.
For monotone functions these are exactly minimal true and maximal false
assignments. Enumerating those boundaries directly produces the same complete
cover, in the original literal-before-absent SOP order, without creating another
BDD manager or implementing an incomplete ISOP approximation.

## Bug review

- Validates unateness BDD ownership and positive support cubes. Fast recursion
  rejects support omissions. ZDD-info decoding rejects non-singleton or empty
  cubes and foreign roots instead of relying on C asserts/indexing garbage.
- Preserves the original `Extra_zddGetSingletonsBoth` empty-set base case. Its
  family contains the empty set as well as literal singletons; the stricter info
  decoder intentionally rejects that family when used outside the recursion's
  singleton-only final-result precondition. Exhaustive tests cover this subtlety.
- Preserves source positive-first unateness priority for variables absent from a
  function; fast/slow info lists only true support variables in manager level order.
- Threshold preprocessing rejects binate functions before mutating inputs. Chow
  sort directions follow actual C comparisons, not contradictory source comments.
- Original `Extra_ThreshHeuristic` complements a uint64 truth table using the
  32-bit `Abc_TruthWordNum` count, writing beyond its buffer at six or more inputs.
  Go uses the correct uint64 count. A six-variable backing-buffer sentinel
  regression verifies adjacent data is preserved.
- Preserves bounded-search false negatives and the heuristic's complemented
  sorted-truth side effect. The source API uses 0 as failure, so it cannot report
  constant-true threshold 0 as success; the port preserves this convention.
  Unlike the malformed short C constant SOP, Go emits full-width valid cubes.
- Bounds-checks truth storage, dimensions, group maps and SOP layout at public
  search entrypoints. Inequality helper arrays retain caller-size preconditions.

## Evidence

- Unateness: every three-variable function checked against independent native
  truth-table polarity and fast/slow equality; shuffled/skipped support tested.
  Original-C fixture checksum: `94e5759806dcde0c`.
- Bounded threshold search: every three- and four-variable function; compares
  original Chow coefficients, mutated truth table, threshold and weights.
  Original-C checksum: `23d92fb53fb7aab7`. Every returned certificate is checked.
- Inequality heuristic: every nonconstant three- and four-variable function,
  exact original weight-search oracle with complete monotone prime covers.
  Original-C checksum: `14b0eda9522e02d5`. Independent tests check both SOP cover
  truth and all returned weight certificates.
- Six-variable source example, invalid storage/SOP/binate inputs, cube validation,
  singleton-family semantics, and root prime forwarding have regressions.

## Validation and scope

The entire root package now passes ordinary race tests and vet. See
`FULL_PACKAGE_REVIEW.md` for the exact full-package/dependency gates, superseding
the earlier explicit-file frontier checks. Original C generators live in
`test_vectors`; normal Go tests need no C compiler.
