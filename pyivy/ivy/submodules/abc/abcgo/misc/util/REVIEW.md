# Utility port review

## Compactness, idiomatic Go, and readability pass

Reviewed the utility implementation independently of the bug pass below.

- Array/storage interfaces use slices; names use strings and explicit byte limits.
- Tagged pointer payloads use numeric `AbcPointerValue` IDs, without unsafe or cgo.
- IEEE conversions use `math`; temporary files use `os.CreateTemp`; file and bridge APIs expose errors.
- Bridge wire graphs use `BridgeNetlist`, keeping protocol code independent of GIA and avoiding a package cycle.
- Reused small helpers for normalized arithmetic, packed-word ordering, full writes, and name storage. Kept source-corresponding exported names to aid source comparison.
- Checked mutex ownership in random and recording state. No additional abstraction or wrapper layer was warranted.

## Separate bug review

Found and fixed two counterexample defects, each first reproduced with a failing test:

1. `Abc_CexTransformPhase` replaced the bit count calculated by duplication with the original count, making dimensions inconsistent when the number of registers changed. Keep the duplicated representation's calculated bit count. Regression: `TestCounterexamplePhaseRegisterResize`.
2. `Abc_CexAlloc` multiplied frame/input counts in int32, allowing dimensions to wrap to an incorrect small allocation. Calculate in int64 and reject counts exceeding int32. Regression: `TestCounterexampleAllocationOverflow`.

Earlier boundary checks also cover bit-count rounding at MaxInt32, literal encoding, zero-tail file reads, hash-table growth, numeric-ID round trips, counterexample permutations, bridge truncation, and packed sorting payloads.

## Validation

- `GOCACHE=/tmp/abcgo-cache go -C abcgo test ./misc/util -count=1` — passes.
- `GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./misc/util -count=1` — passes.
- Manual C/Go differential check: 1,000 randomized normalized pairs, 6,000 float/double add/multiply/divide results matched exactly. Float mantissas were restricted to avoid the original C signed-int intermediate overflow in multiplication.

No experiment directories were tested. Truth-table and ISOP algorithms received additional reviews from their respective package workers.
