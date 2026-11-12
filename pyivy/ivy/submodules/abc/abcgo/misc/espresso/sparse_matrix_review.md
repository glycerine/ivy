# Sparse matrix port review

Implemented source family: `matrix.c`, `rows.c`, `cols.c`, `dominate.c`, `part.c`, and the sparse public/private headers. The unrelated cover-level `sparse.c` belongs to the set/cube algorithms.

## Compactness, idiomatic Go, and readability pass

The representation keeps one typed element shared between sorted row and column linked lists. This preserves source traversal order and constant-time paired removal, which the minimum-cover algorithms depend on. Slice lengths replace redundant allocated-size fields; native `int` indexes replace pointer arithmetic. `any` preserves optional client payloads. No unsafe code, manual allocator, or free-list emulation remains. `sm_cleanup` has the same no-op behavior as the source's default allocator configuration; ordinary Go garbage collection owns storage.

Public aliases and `NewSparseMatrix`, insertion/removal/lookup/count methods, row `Columns`, column `Rows`, and element coordinates let external callers construct minimum-cover inputs and inspect solutions. Returned coordinate slices are independent copies. Internal source names remain available to adjacent Espresso algorithms.

## Independent bug review

Reviewed shared row/column links, duplicate insertion identity, empty-header removal, payload preservation, deep copies, dominance tie ordering, disconnected partitions, indexed capacity growth, numeric overflow, and both input formats. Findings were fixed and regression-tested:

- Corrected original `sm_copy_col`'s swapped row/destination-column arguments. The intended column-copy behavior is tested explicitly rather than preserving the source defect.
- Empty pretty-print returns safely instead of dereferencing a missing last column.
- Element ownership prevents foreign or detached removal from corrupting a matrix. Standalone row/column mutation rejects matrix-owned elements.
- Comparisons avoid subtraction overflow. Hashing computes the source polynomial modulo the requested size without intermediate integer overflow; independent big-integer expectations cover extreme and negative values.
- Decimal readers explicitly parse base ten. Compressed words accept C hexadecimal prefixes, reject overflow/spill bits, and advance the final partial word without integer wraparound.
- Compressed reading grows storage for actual entries, avoiding enormous allocation from an immediately truncated header. Malformed reads retain the successfully read prefix.
- Weight-vector length validation precedes column-dominance mutation.

Partition traversal is iterative. It completes scratch marking even when the graph is connected; the C version may return early with partially marked flags. Matrix contents, partition decisions, and returned blocks match. Native Go scalar/index types and bool predicates are intentional API adaptations.

## Evidence

- 6,000 randomized matrix mutations check sorted forward/back links, identical shared elements, owner identity, counts, and an independent coordinate-map model after every operation.
- 160 reproducible original-C fixtures cover 12,800 mutations plus set operations, hashing, dominance, and partition output. `test_vectors/generate_sparse_oracle.py` rebuilds the oracle; tests use the checked-in fixture without requiring a C compiler. The generator only restores the missing `sorted_insert` macro declaration in this checkout's malformed header, retaining its original replacement body.
- 500 independent exhaustive small-cover comparisons verify unit/weighted dominance optimum preservation and additive partition costs, including zero weights.
- Combined sparse and minimum-cover family tests pass under the race detector and `go vet`. Minimum-cover's separate C oracle also exercises this representation through full exact and heuristic searches.

These are explicit-file family checks; they do not claim the entire Espresso package is buildable while other source families are being ported.
