# Independent compactness, idiomatic Go and readability review

This pass was independent of the package author's port and of the separate
`BUG_REVIEW.md` pass. Changes were coordinated before editing source files.

## Representation cleanup implemented

- Removed decomposition-graph `nSize` and `nCap`; node counts, bounds and last
  node now derive from `len(pNodes)`, and Go append owns capacity growth.
- Removed DSD-network `nNodes`/`nNodesAlloc` mirrors and DSD-manager word-count
  metadata. Network counts derive from the node slice; word counts derive from
  the configured variable dimension.
- Removed SOP `nCubes`, which duplicated its slice length. Normal counts now
  derive from the cube slice; an explicit `failed` flag preserves the original
  public -1 allocation-failure status without overloading a cached length.
  ISOP initializes result state on entry, so a failed result can be reused.
- Removed Cloud's mirrored current-node count, unused allocated-node/memory/
  shift/step bookkeeping and unused cache bit/shift arrays. Actual node count
  comes from the unique-node slice. The AND backend's cache ratio now configures
  a used map-entry budget; at capacity the computed cache is cleared. This
  keeps caching bounded while preserving BDD results. Other Boolean operations
  reduce to this AND backend and do not carry unused cache arrays.
- Removed LUT-manager maximum/output word counts, deriving them from the
  corresponding variable dimensions when needed.

## State deliberately retained

Graph `nLeaves` partitions one combined node slice into leaves and gates; it is
not a second slice length. DSD fan count remains a logical prefix length because
support compression can retain larger fanin/truth backing buffers. DSD object
IDs, references, marks and cached supports are algorithm metadata. Cloud's
configured node limit and generation signature enforce resource/lifecycle
constraints; unique/cache hit counters are actual statistics. SOP literal count
is the ISOP polarity-selection cost cache, populated on recursive ISOP results
and reset when constructing general SOP scratch covers.

C-style exported entry points are retained for recognizable cross-package ports;
local loops, slices, maps, iterators and standard-library operations carry the
implementation. No new tagged pointers, unsafe ownership or redundant slice
capacity fields were introduced.

## Regression and integration evidence

`kit_review_metadata_test.go` covers derived graph/DSD counts through growth,
SOP shrink/append, exhausted ISOP arena followed by reuse, and a one-entry BDD
cache under all 256 three-variable functions. The cache test checks actual misses,
bounded entries and minterm truth, so its eviction assertions are nonvacuous.

Validation after all cleanup:

- `go test ./bool/kit`
- `go test -race ./bool/kit ./misc/util ./bool/bdc ./bool/dec ./aig/hop`
- `go vet ./bool/kit`

All pass. The separate independent bug findings and fixes are recorded in
`BUG_REVIEW.md` and `DSD_ROOT_BUG_REVIEW.md`.
