# `map/cov` port status

## Implemented scope

- Native cover-manager lifecycle and object-ID-indexed node state.
- Variable-width Go cube storage, bit operations, cover collection, checking,
  serialization, and support counting.
- Exact SOP minimization for supports up to 16 variables when the exact-cover
  bounds are met.
- Exact ESOP truth preservation: minimum-cost search through four variables
  and algebraic-normal-form conversion for larger supports.
- AIG-to-logic conversion with iterative cut-boundary growth. When recursive
  support exceeds `nFaninMax` or neither bounded cover fits `nCubesMax`, it
  promotes solved fanins to cut points and retries above them, as
  `covCore.c` does.
- `fUseEsop` and `fUseSop` select the allowed cover forms. Derivation chooses
  among positive SOP, negative SOP, and ESOP covers by cube count with the
  tie behavior in `covBuild.c`. ESOP covers become XOR-special nodes.
- Network conversion carries complemented AIG edges, preserves PI/PO names,
  and validates the result before returning it.

The Boolean covers use exact local truth tables and bounded exact minimizers.
They preserve the function while choosing the cover with the fewest cubes;
exact minimization can produce different ties than the C heuristic.

## Review notes

- `map/cov/minimize` has exhaustive truth-table and independently brute-forced
  minimum-cost tests for every three-variable function.
- ESOP minimization toggles overlapping cubes, preserving XOR rather than OR
  semantics.
- Cut-growth and ESOP-selection behavior have direct network-level truth-table
  tests, in addition to polarity selection, complemented edge, SOP, and
  impossible-limit tests.
- Cube storage uses Go slices; no pointer reinterpretation or C allocator
  remains in these APIs.

## Verification status

The nested minimizer package passes its test and vet gates. Whole-package
`map/cov` test and vet are currently blocked because imported `base/abc`
sources fail to parse in generated files. Once that package is fixed,
run:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./map/cov -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./map/cov -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./map/cov
```
