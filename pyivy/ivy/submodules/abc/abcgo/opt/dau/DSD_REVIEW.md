# DSD and DSD2 port review

Owned sources: `dauDsd.go`, `dauDsd2.go`, `dsd_native_test.go`, and reproducible fixtures/generators under `test_vectors/`. Tree/Merge and Canon/count remain independently owned.

## Compactness, idiomatic Go, and readability pass

Main DSD implements the actual source single-variable output extraction, paired input extraction, support cache, outer/inner MUX factoring, prime output, one-step prime splitting, variable-level priorities, support minimization, normalization, legacy six-variable `Perform`, and truth composition. One slice-based implementation replaces the source's duplicated six-variable and multiword engines. This avoids two versions drifting while preserving the source's factor order, definitions, residual truth mutation, and prime-size return convention.

The parser owns its expression nodes and supports variables, complement, AND, XOR, MUX, inline hexadecimal primes, and `<function>{arguments}` composition. Prime prefixes stay opaque during normalization/brace flattening, as in C. No static output or scratch buffers remain. Decomposition contexts own definitions, directed support caches, and output builders; recursive contexts are independent.

The source-shaped decomposition API retains integer flags, caller-owned residual truth, and explicit output bytes. `DecomposeDSD` is the idiomatic convenience API: it copies input truth, writes inline prime truth, and returns `(string, largestPrime)`. `Normalize` is immutable; `ToTruth` returns independent storage. Public source printing/composition entry points are retained. Source demonstration-only test routines are replaced by tests, rather than library functions that print hard-coded examples.

## Independent bug review

Reviewed complement placement, nested closure order, source pair scan/cache order, current-variable swaps versus original definitions, cofactor support, MUX input removal, residual prime copying, split statuses, source selection-sort tie order, hex width, argument domains, and small-table stretching. Main DSD matches exact original C output, not merely Boolean equivalence.

Regressions found and fixed during review:

- Composed MUX prefixes must remain opaque when normalizing arguments; a red test caught accidental reordering inside the prefix.
- Empty Go level slices now mean no level priorities. Short nonempty level vectors still fail explicitly.
- Numeric adjacent-cofactor ordering must compare words rather than the lowest differing bit; a red DSD2 test caught this distinction.
- DSD2 trial snapshots reset the copied `strings.Builder` before writing, preventing illegal copied-builder mutation.
- Parsed variable domains, prime hex widths, truth storage, permutation ranges, and structural delimiters are checked instead of risking C-style out-of-bounds accesses.
- Level computation ignores punctuation in compound variable definitions; the source scans these characters as though they indexed the level array.

## Explicit DSD2 source boundary

`dauDsd2.c` is an unfinished experimental source file, not a working second decomposition engine. It contains invalid `#include` constant definitions, undeclared loop variables, undeclared `nVarsF`, missing semicolons/return values, out-of-range cofactor indexes, and AND/XOR-recording TODO loops that never reduce their loop bound. This is not hidden or described as executable reference behavior.

The defined primitives (`Abc_Tt6HalfUnShuffleVars`, `Dua_DsdTryConst`, zero-cofactor and opposite-cofactor checks) preserve their actual source behavior and have unchanged-C vectors. In particular, the half-unshuffle selector retains the draft's surprising convention: nonzero selects cofactor zero, zero selects cofactor one.

The remaining draft operations are completed according to their stated mathematical operations, with these explicit corrections:

- Full half-unshuffle maps the selected cofactor without the draft's uninitialized write index or aliased word-pointer loop.
- Variable trials compare valid cofactor storage, return `-1` on independence instead of an assertion without a result, and apply complete phase transformations.
- Adjacent trials preserve source numeric ordering while performing a complete permutation, instead of changing one word or repeatedly swapping the wrong two offsets.
- Pair decomposition checks compare the intended cofactors without falling through an unrelated branch or reading past a one-word table.
- Bound-range checking tracks two distinct cofactors without using all-ones as an ambiguous sentinel, validates every outside assignment, and commits compacted truth only on success.
- Low-bound complement checking uses width-correct complements; derivation returns the stretched outer XOR selector. This replaces byte/short integer-promotion errors, a negative shift for wide bounds, and a boolean passed as a variable count.

The source TODO trial loops use the *actual shared MainDSD single/pair extraction stages*. Their result records a truth-equivalent `Expression` and typed object tree. `Dua_DsdOneVar` reports residual phase; `Dua_DsdTwoVars` reports whether the current base shrank. These return/state details are native completions of source-under-defined entry points, not claims about an existing C result. `Dua_DsdTest` now returns that inspectable result instead of being an incomplete void demonstration.

## Evidence and commands

- 800 unchanged original-C fixtures cover one through ten variables, both prime-split flags, level priorities, exact expression strings, and exact largest-prime statuses.
- Legacy `Dau_DsdPerform` matches exact C strings on 480 small fixtures.
- 1,000 seeded random truth tables round-trip through decomposition and normalization without changing input storage.
- 600 unchanged-C DSD2 primitive vectors span one through twelve variables.
- Independent trial-expression equivalence, bound substitution, wide complement-factor recognition/selector derivation, ordering, and composed-prefix regressions cover completed draft behavior.

From the repository root:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race opt/dau/dauDsd.go opt/dau/dauDsd2.go opt/dau/dsd_native_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet opt/dau/dauDsd.go opt/dau/dauDsd2.go opt/dau/dsd_native_test.go
```

The real DSD + Tree/Merge subset also builds/tests with `dau_tree_types.go`, `dau.h.go`, `hierarchical.go`, `truth_counts.go`, `dauCount.go`, and `tree_merge_test.go`. These are scoped checks while unrelated DAU drafts are being ported. Fixtures require no C compiler during ordinary tests; the Python generators rebuild them from this checkout's original source.
