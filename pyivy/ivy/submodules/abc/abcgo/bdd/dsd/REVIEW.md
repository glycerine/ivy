# DSD completion reviews

## Compactness, idiomatic Go, and readability pass

Completed separately from semantic review. The 3,800-line C implementation is
represented by typed node/manager structures, a 400-line cofactor kernel, small
support-cube helpers, and shared traversal/printing routines. Removed C allocator
plumbing, global performance counters, fixed scratch arrays, pointer tagging,
and repeated OR/EXOR printer implementations. Canonical inverse handles preserve
pointer-complement semantics without unsafe code. Native maps preserve BDD-node
identity and shared decomposition nodes. The already reviewed `opt/dau/canon`
leaf helper is reused for prime expression printing.

Local prime functions retain the source's adjacent-component support remapping
and witness-cube recursion. The unused permutation entries are initialized to a
complete permutation. Production code never enumerates assignments to discover
DSD partitions. All source-declared APIs are present; no algorithm is stubbed or
hidden. Remaining source-style names keep direct comparison with C practical.

## Separate semantic and bug review

Compared each cofactor case and polarity with `dsdProc.c`, the four-function
identity recursion with `dsdCheck.c`, local-variable remapping with `dsdLocal.c`,
and depth/cost/shared-node/DFS semantics with `dsdTree.c`.

Findings and fixes:

- Strict support containment excludes equal supports, matching the source.
- C's cache clear repeatedly initializes entry zero. The Go cache initializes
  every entry, and the original-C test oracle fixes that initialization only.
  Repeated-process experiments reproduced different C decompositions without
  the fix. No C production source was edited.
- Traversal cleanup must preserve the user `Mark`. A red regression caught the
  initial Go cleanup clearing it; visits and user marks are now separate.
- The source overlap algorithm can mistake an arbitrary user mark for its own
  generation stamp. A red regression produced a wrong reconstructed function.
  Local containment-path membership replaces generation stamps; user mark values
  cannot alter decomposition.
- Optional depth/gate output pointers accept nil, matching C.
- Dynamic component lists avoid the source's capacity calculation based on the
  top variable's absolute level; declared inputs below unrelated variables work.
- Complement markers for printed prime nodes wider than six inputs are retained.

No known findings remain open.

## Validation

The full DSD package compiles without filtering any Go files. Tests reconstruct
all 256 three-variable functions and verify disjoint support partitions, global
functions, local prime functions, and complement-cache identity. Random and
structured cofactor cases run under multiple BDD orders. A composed 80-input
function tests scalable decomposition; a seven-input prime tests both local
remapping conventions and expression output.

The original-C differential compares **4,756** ordered decomposition trees,
including regular-node global truth tables, complement tags, and prime local
truth tables. Cases cover all functions through three variables, random and
structured functions through six variables, and normal/reversed BDD orders.
Only the original C cache-initialization defect is corrected in the oracle.

Additional tests cover conditional identity against 1,024 independent expected
results, check-cache allocate/clear/release behavior, shared DFS/counts, tree
cost/depth, user mark preservation, adversarial mark values, declared support
below unrelated BDD variables, printers, and independent concurrent managers.
