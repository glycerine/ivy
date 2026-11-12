# Cascade helper review

Authoritative port of the CAS-used subset of `extraBddCas.c` and
`Extra_bddBitsToCube` from `extraBddMisc.c` (Alan Mishchenko). This package does
not claim the remaining root `extrab` algorithms are complete.

## Readability and compactness

The binary encoder, deterministic cut-path collection, width profile, and bounded
non-strict search are separate operations. Search state is local to one call;
there are no fixed global caches or scratch fields. The source's five-level and
100-backtrack bounds, first-best order, cube cofactoring, and final code recursion
are preserved. Width profiling uses earliest incoming-edge level intervals,
avoiding repeated cut construction. Code-count traversal preserves the source's
sum over distinct reachable code subgraphs rather than silently replacing it with
a different union-count optimization.

## Bug review

Checked complemented-edge traversal and polarity, skipped levels, shared paths,
zero-code stages, MSB-first binary codes, empty/unused codes, search cutoffs, and
independent cofactors. Local maps eliminate the original `HHTable2` pointer cache
surviving destroyed managers. The C fixture clears that cache between fresh
managers; otherwise allocator address reuse contaminates independent cases.
Columns must partition their input space and encoding variables must follow
functional variables, as required by the original API.

`encoding_test.go` checks direct-wire encodings, skipped-level width and cut
reconstruction, and 3,856 partition instances. Each input has one code, and no
code aliases different columns. Optimization counts and code multiplicities match
the original-C checksum `6b850130b3287765` in
`test_vectors/encoding_oracle.c`. Concrete code numbers are intentionally not
compared because original table traversal depends on pointer addresses.

Validation: race tests and vet for `./bdd/extrab/cascade`, also run together with
CAS and CUDD core. No native compilation occurs in Go tests.
