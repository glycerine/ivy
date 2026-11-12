# Cover sparse reduction and cost review

Scope: `sparse.c.go` (`make_sparse`, `mv_reduce`) and `cvrmisc.go` (costs, formatting, timing, fatal errors). Sparse matrix primitives are reviewed separately in `sparse_matrix_review.md`.

## Readability and compactness pass

The cover sparse port retains the original sparse-variable/part/cofactor loops and original-row mapping. It uses a slice of cube views instead of a manually allocated pointer table, preserving the important distinction between copied cofactors and mutable original cover rows. Dense expansion and sparse reduction remain separate calls to actual implementations; no simulated or approximate replacements were introduced.

Cost counting directly enumerates each cube's variable parts instead of allocating a cube list and invoking `massive_count`. The binary zero count, sparse multiple-valued one count, dense multiple-valued zero count, output one count, and PRIME count are unchanged. It avoids writing unrelated `cdata` scratch fields as a side effect. Whole-struct assignment implements `copy_cost`; formatting uses ordinary Go strings.

## Independent bug review

Checked relative-essential flags, cofactor-to-original row indexes, preservation of full non-output variables, clearing PRIME on reduced redundant cofactors, active-count compaction after any variable becomes empty, equality-based stopping, optional repeated sparse/expand sweeps, and exact timing-counter slots. The transferred sparse port required no algorithmic changes after integration with actual irredundancy and expansion implementations. Cost regressions initialize packed-set metadata explicitly, avoiding misleading tests that accidentally create zero-loop rows.

The library uses monotonic elapsed milliseconds for `ptime` instead of platform-specific process-CPU milliseconds. Accumulated units and formatting precision remain milliseconds; reported wall time can include scheduler waits. `fatal` panics with the source message rather than terminating the embedding application's process. These are intentional Go library adaptations.

## Validation

`test_vectors/generate_cover_sparse_oracle.py` links original C set/cube/cofactor/complement/irredundancy/expand/minimum-cover/sparse/cost implementations. It removes an unavailable utility include, supplies allocator/namespace/timing compatibility, and restores this checkout's missing `sorted_insert` declaration without changing its body. No algorithm is stubbed. The test fixture requires no C compiler during ordinary Go testing.

120 seeded multiple-valued covers compare exact reduced rows, exact final sparse rows, exact rows with repeated `force_irredundant` sweeps enabled, and all six cost fields against C. Don't-care covers and correctly complemented off-set minterms exercise real expansion. Independent cost enumeration spans binary-only and mixed sparse/dense geometries, empty covers, and PRIME flags. Timing accumulation, formatting, copy semantics, and fatal context have direct tests.

The combined real-dependency focused test group passes the race detector and `go vet`. This does not claim the entire Espresso directory compiles before its other source families are finished.
