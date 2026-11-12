# MaxMin family review

Completes all algorithms of `src/bdd/extrab/extraBddMaxMin.c` in the authoritative
sets leaf. The previously ported maximal extraction is shared with the set-filter
family; new minimal extraction, extremal unions, pairwise intersection product,
and fused maximal union product share that engine. Root files are only forwards.
Original C source and its attribution header remain visible and unchanged.

## Readability and compactness pass

- Preserved the source's cofactor recursions and local antichain pruning. Max/min
  union share only their genuinely common split/composition, while retaining
  opposite dominance filters. Cross product and fused maximal dot product have
  separate branch formulas that mirror the source.
- Dot product is exactly CUDD's existing generic `Cudd_zddUnateProduct`, so the
  public function delegates to that sole implementation instead of adding a
  duplicate graph algorithm. Every generic input pair is independently tested.
- Stable node IDs replace C pointer ordering solely for commutative memo keys;
  semantic splits use manager levels. Canonical node composition remains in core.
- Additional memo tables initialize lazily, preserving the existing set filters'
  small call footprint. Constant/empty fused products terminate before unnecessary
  antichain normalization.

## Bug review

The C public MaxUnion/MinUnion/MaxDotProduct wrappers omit their fused recursions'
required antichain precondition. Their shortcut `MaxUnion(X,0)=X`, for example,
violates the documented maximal-union semantics if X contains comparable members.
The port normalizes public inputs to maximal or minimal families before entering
those unchanged fused recursions. This makes the documented semantics valid for
arbitrary families. Explicit identity, zero and empty-cube regressions cover the
source defect. Native differential tests supply original C with normalized
antichains; independent semantic tests cover general public inputs exhaustively.

Checked all terminal combinations, empty-family vs empty-set behavior, proper
inclusion (not merely equality), skipped top levels on either operand, canonical
zero suppression, membership ownership/family checks before shortcuts, and
nonidentity variable orders. Cross product computes pairwise *intersections*,
not pairwise unions; dot product does the latter. Inputs may be arbitrary set
families rather than restricted Boolean covers.

## Evidence

Every 256×256 three-element family pair under both identity and nonidentity ZDD
orders is checked against ordinary finite-set arithmetic for all five binary
operations. Minimal extraction is exhaustive too. Another 1,024 seeded
four-element family pairs check minimal union and fused maximal products.
Invalid manager ownership, nil, BDD and numeric ADD inputs have regressions.

Original-C fixture `test_vectors/maxmin_oracle.c` executes real original routines:
ordinary products for general families and extremal operations for the requisite
antichains. Checksum `d5ec309d9c0db5a5` matches the permanent Go fixture.
Regeneration is optional; normal tests require no C compiler.

Leaf race tests/vet and root forwarder subset gates are required. Normal root
extrab still fails in visible `extraBddAuto.go` drafts; the MaxMin family itself
is complete, not a claim that the surrounding root package builds.
