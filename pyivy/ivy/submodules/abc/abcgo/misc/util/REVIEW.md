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


## PI-permutation dependency follow-up during the GIA audit

A new Go-only regression finds Abc_CexPermute silently clears initial register
bits. The C helper has the same omission, although Gia_ManVerifyCex explicitly
replays those bits. PI permutation changes neither register count nor register
order, so preserve their prefix. Copy full register words and mask the partial
word before remapping PIs, preventing unpermuted PI bits from leaking through.
Abc_CexPermuteTwo inherits the correction. Abc_CexDup has different source
register-resizing/reset semantics and is outside this repair.

Seventy-two cases cover register counts0/1/31/32/33/67,3/11 frames and all six
three-PI permutations. Independent boolean packets check the direct helper and
composition from two nonidentity canonical orders, every live bit, metadata,
source immutability and result ownership. Sixty direct cases fail before repair;
zero-register cases already pass. Both paths execute in all72 repaired cases.
Red /tmp/abcgo-cex-permute-initial-red.log; full repaired utility suite
/tmp/abcgo-cex-permute-initial-green.log. No C fixture or toolchain is required.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests. Full utility race1.360s, focused GIA isomorphism race1.823s and util/GIA
vet pass; all handles ended. Final logs /tmp/abcgo-cex-permute-final-
{module.jsonl,util-race.log,gia-race.log,vet.log}. This is a dependency repair
within the sequential GIA audit, not completion of the whole utility audit.
Original package completion stays2/131.
