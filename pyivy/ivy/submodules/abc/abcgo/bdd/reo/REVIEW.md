# REO review gate

## Design and readability review

Reviewed the public API, storage, transfer representation, profiles, adjacent
swaps, sifting, and shuffle separately from correctness review.

- Replace C pointer-array storage with Go slices and explicit complemented edge
  wrappers. Keep canonical CUDD graph identity in its authoritative core package.
- Use immutable canonical permutation rebuilds for adjacent swaps rather than
  reproducing pointer tagging and manual hash ownership. This is a real sifting
  algorithm, with max-populated unsifted-plane selection, nearest-end-first
  exploration, bounded temporary growth, strict best-cost tracking and iterations.
  It intentionally omits lower-bound pruning, evaluating additional candidates.
- Retain compatibility API names; use Go slices for root/permutation arrays and
  `int` for their lengths. Provide Manager methods returning explicit errors.
- Consolidate cost derivation in one shared-root profile computation; snapshots
  returned to callers clone their slices.
- Removed a redundant moving-direction boolean adapter; canonical swaps need
  only the level, while the compatibility helper retains its historical argument.

The allocation strategy can retain candidate nodes in the CUDD unique table,
which currently has no collector. Large reorderings may need a scratch manager
or future core garbage collection; no performance equivalence is claimed.

## Independent bug review and regressions

- Found stale transfer caches/arena when reusing a manager across support remaps.
  Reset run-local caches and storage. Added a two-run remapping/transfer test.
- Restored source behavior: BDD width mode falls back to node minimization;
  original width profiling was intended for ADDs without complemented edges.
- Checked first-use roots, shared/complemented DAGs, simultaneous roots,
  constants, empty roots, invalid/foreign roots, sparse support and remapping.
- Checked inverse mapping by both canonical pointer equality and exhaustive
  assignments for 35 deterministic random 5-variable DAG families.
- Checked an initially poorly ordered mux strictly improves, while resulting
  functions and the original CUDD variable order remain intact.
- Checked output reference ownership for single-root and array APIs, transfer
  round-trips, statistics isolation, shuffle bijections, storage capacity,
  chunk address stability and recycling order.

`GOCACHE=/tmp/abcgo-go-cache go -C abcgo test ./bdd/reo -race -count=1`
passes. No source is hidden, build-tagged out, or replaced with success stubs.

## ADD/MTBDD extension: separate design review

Reviewed the authoritative CUDD core ADD APIs before integration. Use
`Cudd_IsAdd` to dispatch each root independently to `Cudd_addPermute` or
`Cudd_bddPermute`; numeric terminals never flow through Boolean ITE. Unit nodes
record their original family, so transfers reconstruct ADD branches using
`Cudd_addIte` with ADD indicator variables. Canonical terminal nodes are reused,
including positive one shared between BDD and ADD families. Mixed root arrays
preserve each root's family. Keep shared node/width/APL profiling independent of
terminal magnitude, as in the original cost definitions.

## ADD/MTBDD extension: independent bug review

The first ADD integration run found a real cross-package error: core `Support`
accepted only BDDs, silently producing empty support for ADDs and disabling all
sifting. The core owner fixed this traversal. Removed the temporary duplicated
traversal and use the sole authoritative `dd.Support` API for both families.
Regression tests require nonempty sparse support, mux size improvement, numeric
terminal preservation, inverse canonical identity, and all input assignments.

Additional regressions cover negative/fractional terminals, shared ADD roots,
constant roots, all three optimization criteria, sparse support remapping,
shuffle, transfer round trips, and mixed ADD/BDD/complemented roots. Numeric
terminals are counted distinctly. The width objective remains selected for
ADDs, while BDD width retains the source fallback.

A separate API audit found that the initially ported `Extra_bddReorderTest`
returned Boolean success; its C interface returns node count. Fixed it and added
an exact four-node mux regression. Added `Extra_addReorderTest` with the source's
intentional nonzero-pattern conversion, tested for constant and nonconstant
patterns. Both benchmarks currently use this Go REO sifting engine: they are not
claimed to reproduce the original optional CUDD-reference heap-reorder backend.

`go -C abcgo test ./bdd/reo ./bdd/cudd/core -race -count=1` and
`go -C abcgo vet ./bdd/reo` pass.

## Classical comparison dependency resolved

Implemented the required core ShuffleHeap, SIFT/SYMM_SIFT (and convergence)
backend with live referenced-handle preservation, true symmetry grouping,
cache clearing and source-style dead-node collection. Generic cross-manager
BDD/ADD transfer preserves numeric leaves and canonical polarity. The optional
`Extra_ReorderCudd` now performs the original independent-manager SYMM_SIFT
comparison and returns the inverse mapping, including nonidentity source orders.
BDD and ADD benchmark helpers use the independent CUDD SIFT backend and actual
node counts; ADD benchmarking retains the source's intentional nonzero pattern.
Source diagnostic metrics and deterministic randomized shuffle checks are active.

The backend has its own separate compactness/bug audits in
`bdd/cudd/core/REORDER_REVIEW.md`. A real original-C fixture checks all 512
three-variable SIFT/SYMM_SIFT cases for exact sizes, truth, public status, symmetry
masks and group counts. A CUDD source-comment/public-return discrepancy was
found and fixed using this independent evidence. ZDD state isolation is tested.

REO's requested comparison dependency is resolved. The wider CUDD root package
still contains unrelated unported modules and additional ordering facilities;
only supported core methods are claimed by this component's tests.

Final verification also exercised source-style diagnostic entry points: the
independent manager preserves the input order and external reference count,
returns a valid inverse permutation for both BDDs and numeric ADDs, and the
shuffle diagnostic uses 120 deterministic adjacent random swaps after remapping
support to the topmost levels. Diagnostic node-count/time output is preserved.
