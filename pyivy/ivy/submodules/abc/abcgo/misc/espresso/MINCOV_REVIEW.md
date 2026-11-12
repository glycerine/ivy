# Espresso minimum-cover family review

Ports original Espresso `mincov.c`, `mincov_int.h`, `solution.c`, `indep.c`, and `gimpel.c`. Sparse matrix storage, dominance, and partition operations are provided by the separately reviewed sparse family.

## Compactness, idiomatic Go, and readability pass

Performed separately from the bug review below.

- Retained the original reduction/bounding/search order: dominance and essentials; unit-cost Gimpel reduction; independent-row lower bound; smaller-component-first decomposition; accept-column branch before reject-column branch; heuristic early return.
- Replaced pointer weights with `[]int32` (nil means unit cost), matrix IDs/iteration with native ints, integer debug flags with bools, and CPU-clock formatting with standard-library monotonic elapsed durations.
- Sparse linked traversal remains explicit and source-comparable. Solution copying owns its selected-column row; independent-bound construction uses local visited state rather than mutating caller flags.
- Reused `mincovWeight` for the shared optional-weight contract. Helpers retain their source names, while `SmMinimumCover` exposes the top-level Go API through public sparse type aliases.
- Kept statistics local to each search; no shared mutable global search state, cgo, unsafe, stubs, or hidden source.

## Separate bug review

A failing regression exposed inherited duplicate-column cost accounting: row insertion is set-valued but `solution_add` charged a repeated column again. Check membership before charging; `TestMinimumCoverSolutionSetCost` first failed with length 1 / cost 10 instead of cost 5 and now passes. C-oracle choices/costs for normal search remain unchanged.

Reviewed strict upper-bound semantics and int32 accumulation. The public entry point rejects negative active-column weights and sums the initial bound in int64, rejecting a total that cannot fit the source's int32 search bound. Regression: `TestMinimumCoverRejectsInvalidWeightCosts`.

Reviewed linked-list deletion iteration, solution ownership, matrix preservation, weight-zero scoring, independent-set admissibility, Gimpel restoration, component accounting and weighted branching. Independent optimality checks include zero-weight columns. Empty matrices return an allocated empty cover; missing search results are handled without dereferencing nil.

## C oracle and reproducibility

`test_vectors/generate_mincov_oracle.py` compiles copies of the original C algorithm and sparse-family sources, then generates 1,000 deterministic cases covering unit/weighted and exact/heuristic modes. Only compatibility scaffolding is repaired in its temporary header copies: restore the missing `sorted_insert` macro declaration and provide standard ALLOC/REALLOC/FREE/NIL/MAX/MIN/WEIGHT, namespace and timing declarations. Algorithm C files are unchanged. The permanent compact `test_vectors/mincov_oracle.json` contains the first 120 cases; the generator also retains the full 1,000-case JSON in its printed temporary directory. Ordinary Go tests require neither Python nor a C compiler.

`TestMinimumCoverCOracle` compares exact selected columns and costs. Set `ABC_MINCOV_ORACLE` to the full generated JSON path to run all 1,000 cases. All 1,000 matched the original C reference.

## Validation

The combined sparse/minimum-cover source family passed normal and race tests using:

```
GOCACHE=/tmp/abcgo-cache go test sparse.h.go sparse_int.go matrix.go rows.go cols.go dominate.go part.go mincov_int.go mincov.h.go mincov.c.go solution.go indep.go gimpel.go mincov_test.go mincov_solution_test.go -count=1
```

Run from `abcgo/misc/espresso`; add `-race` for the race run. This deliberately names the complete collaborating families while the remaining Espresso algorithm files are being ported by other workers; it does not hide or disable any source.

Coverage includes:

- 120 permanent C-reference cases and a verified 1,000-case differential run.
- 1,000 independent random exact-optimum comparisons against exhaustive subset enumeration, including zero weights, with heuristic coverage/cost bounds and caller-matrix preservation.
- Sparse row/column IDs, disconnected components, direct Gimpel reduction, independent weighted bounds and caller flag preservation.
- Solution duplication, accept/reject behavior, best-solution selection, duplicate accounting, empty covers, invalid costs and bound overflow.

No experiment directories were tested. Full Espresso package validation remains an integration responsibility once its other source families are ported.
