# MVC integration review

## Compactness, idiomatic Go, and readability audit

Reviewed the representation separately from algorithmic correctness. Removed
Cube.Last because len(Data)-1 is authoritative; the public ReadLast/SetLast APIs
now derive/resize the word slice. Removed Cover.NWords and Cover.NUnused because
NBits determines them. Existing bool/dec consumers require no removed fields.
Added regressions for derived word counts and resizing within/beyond capacity.

Cube.Unused is retained: a word slice alone cannot represent the last word's
logical bit width. Cube.Ones is a genuine algorithm cache used as the stable
sort key; CoverSetCubeSizes refreshes it before containment sorting. The list's
Items field provides constant-time counts for a linked list, rather than
replicating slice length. Explicit pools are replaced by Go allocation/GC.

## Separate bug and semantic review

Reviewed subset versus numeric comparisons, padding, literal codes, aliasing,
containment, algebraic quotient/remainder reconstruction, and traversal mutation.
Fixed appending a cover to itself: previously the traversal could follow newly
appended nodes forever. It now copies exactly the original source entry count;
a regression checks doubled contents and independent cube storage.

The historical BitLess/BitMore macros perform unsigned numeric comparisons, not
subset implication. The Go port preserves those public semantics and uses
BitNotImpl for internal containment/division tests. Tests exercise word boundaries,
100 exhaustive binary-containment truth comparisons, 200 randomized division
reconstruction cases, kernels/common cubes, remap/transpose, cofactors and flip.
The bool/dec consumer additionally verifies exact factoring over all 256
three-variable truth functions. Normal package/consumer tests and race tests pass.
