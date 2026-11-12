# DAU native Go port

The complete package builds and passes its normal tests. All generated C-shaped implementations have been replaced with native algorithms; no product source is hidden by build tags or excluded from package compilation. `dauCore.c` is the original successful no-op example-manager template, which is preserved as such. The unfinished experimental DSD2 source is addressed explicitly in `DSD_REVIEW.md`.

## Completed families

- Canonicalization: original cofactor signatures/refinement, phase-only and permutation-only transforms, hierarchical caches, grouped exact minimization, and adaptive tied-group/symmetry/SCC cost-aware search.
- Counting: packed-byte cofactor table generation, recursive quick/fast counters, normalized counts, and shifted cofactor coefficients.
- Bound-set decomposition: original ordered set search, Gray traversal and cofactor classification, shared/free/unique partitions, components and permutations, and recomposition checks.
- DSD: original single/pair decomposition, parsing, truth conversion, normalization, array representation and flattening, structural tree/hash manager, string merging and shared-support transforms.
- Enumeration/synthesis: truth and network NPN enumeration, formula libraries and multiplicity/delay modes, DSD enumeration/divisors, and the real GIA bridge.

## Representations and contracts

Truth tables are little-endian `[]uint64`; variables below six repeat across the unused word bits. Source flags, counts and variable numbers retain `int32` where part of an existing API. String normalization/replacement returns a new Go string; source output buffers become explicit slices or slice pointers.

`Abc_TtCanonicize` and its phase/permutation variants use the acyclic `canon` leaf package shared with GIA. Transform permutations map new positions to original variables; input phase bits refer to new positions, with output phase at bit `nVars`. The adaptive variants additionally retain source flags: bit29 is a cache hit, bit30 marks exact search, and bit28 marks the cost-aware hard branch. Cache-hit metadata and hierarchical classification results have the same limited transform guarantees as the original source; `Abc_TtCanonicizeHie` returns zero as the source does. Workspaces/cache managers are mutable; scratch truth buffers are local to each call rather than C globals.

Simple/Quick cofactor counters retain normalized 64-bit-word counting below six variables; Fast counts actual minterms there. Quick's zero-variable case retains the source's single-bit result. The packed table is generated from eight-bit population counts rather than maintained as thousands of literals.

Non-DSD sets use two bits per original variable: zero=free, one=unique bound, three=shared. `FindSets` preserves input truth and source result order. Top-check output workspaces should begin at zero, as in C, and are only valid on success; failed checks may leave partial decomposition bits. Shared variables occupy the high decomposition-input positions, and Gray traversal determines the stable cofactor labeling. Scalar component output permutations retain the source's extra variable `nVars` for the decomposition result.

## Corrected original-source defects

Regression tests cover each correction:

- An eight-variable all-one table overflowed Quick's packed total byte; higher recursive calls overwrote rather than accumulated cofactor counts.
- Flattening a complemented nested XOR lost its inversion.
- The balanced-output wrapper omitted the extra output phase when selecting the complemented input's result.
- Uppercase bound-set parsing subtracted lowercase `a` and shifted a negative C value. Go gives uppercase variables their defined shared encoding, including variable15. The C oracle excludes only these undefined parser cases; independent mathematical regressions check their intended encoding.
- Failed diagnostic component verification now returns failure instead of always reporting success.
- Zero-support SCC/cost-aware search avoids invalid empty-group indexing; exact hierarchical constant handling avoids unnecessary factorial permutation schedules.

## Verification

From the repository root:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./opt/dau
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./opt/dau ./opt/dau/canon -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./opt/dau ./opt/dau/canon
```

All pass, including the independent reviewer's reruns. Original-C oracles compile the actual source files where a C compiler is available; production Go has no C dependency. Checks include exhaustive three-variable transforms/decompositions, exact tables/phases/permutations and cache reuse, structured and random multiword cases through16 variables, ordered set/component results, and original ABC enumeration/synthesis comparisons.

Separate review records: `DAU_INDEPENDENT_REVIEW.md`, `DSD_REVIEW.md`, `NPN_REVIEW.md`, `ENUM_DIVS_GIA_REVIEW.md`, `TREE_MERGE_DESIGN_REVIEW.md`, and `TREE_MERGE_BUG_REVIEW.md`. Shared leaf reviews are in `canon/INDEPENDENT_REVIEW.md`.
