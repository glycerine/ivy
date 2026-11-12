# Separate Tree/Merge design and compactness review

Scope: dauTree.go, dauMerge.go, dau_tree_types.go and tree_merge_test.go.
The shared DAU header and main DSD/DSD2 implementation belong to other workers.

Dss objects retain original structural hashing, support ordering, AND/XOR
flattening, MUX phase normalization, prime canonicalization, permutation
collection, truth composition, shared-variable mapping and merge caching. The
actual DAU decomposition implementation is called; no Kit engine is substituted.
String merging follows source pure-subtree classification/substitution, Boolean
conjunction/decomposition, definition inlining and restoration of global names.

Go objects and slices replace packed flexible arrays, pointer tags and arenas.
Complemented references have explicit paired objects, created with their regular
node. Fanin, shared-map and result-permutation counts derive from slice lengths;
there are no redundant nFans/nWords/nShared/nObjs fields. Configured NVars and
NonDecLimit remain genuine resource/algorithm settings. A node's support count
is an immutable structural aggregate over child occurrences; it is required for
permutation segmentation and cached after node creation, not a slice-length
mirror. Hash/caching tables own their entries and copied results.

An internal expression parser replaces pointer/string offset mutation. It checks
matching delimiters, fanin counts, variable bounds and disjoint network support,
and supports explicit hex primes plus functional composition. Brace removal
retains names and phases without introducing unrelated normalization. Public
ParseDssNetwork/MergeDSD return errors; source-style wrappers remain available.

Truth queries return owned slices and operation results own their permutations,
replacing process-global scratch buffers. Diagnostic printing uses io.Writer and
returns errors; Dss_ManExpression exposes plain expressions separately from the
source-compatible ID/support diagnostic line. Original inactive drivers are
explicitly documented. The stale 444 demonstration uses valid symbolic leaf
references instead of its original out-of-range references.

The source's entire six-LUT experimental checker is under #if0 and is not
invented as active behavior. Original source attribution remains in the port.
Native harnesses/fixtures are visible in test_vectors, with no hidden Go sources
or build-tag exclusions.
