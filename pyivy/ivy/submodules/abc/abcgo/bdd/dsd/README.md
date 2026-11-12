# Disjoint-support decomposition

This package ports ABC's BDD-based recursive DSD algorithm onto `bdd/cudd/core`.
It retains all APIs declared in `dsd.h` and `dsdInt.h`: manager ownership,
multiple-output shared decomposition, single-function decomposition, conditional
root identity, local prime functions, traversal/statistics, and both tree printers.
There is no truth-table search in the production decomposition algorithm.

The core preserves the original cofactor cases: constant/parity reductions,
contained components, common OR/EXOR factors, compatible prime merging, and
containment-path refinement for overlapping cofactor supports. Cached regular
BDD functions map to possibly complemented DSD handles. OR/EXOR components
flatten as in C; prime components have uncomplemented handles. Support partitions
and prime local-function ordering follow the current BDD variable order.

API adaptations:

- `DdManager` and `DdNode` alias the completed `cudd/core` types.
- Complemented DSD pointers become canonical Go edge handles; use
  `Dsd_Regular`, `Dsd_Not`, and boolean `Dsd_IsComplement`/`Dsd_NotCond`.
- Pointer arrays become typed slices and file output accepts `io.Writer`.
- Counts and indices use `int`. Slice-returning APIs borrow manager/node storage.
- `Dsd_TreeGetPrimeFunction` uses the first formal-input **levels**;
  `Dsd_TreeGetPrimeFunctionOld(..., 1)` uses formal-input **indices**.
- Input functions must use only indices below the declared `nSuppMax`.
  These variables may be anywhere in the BDD order.
- A manager retains BDD references for node functions/supports until
  `Dsd_ManagerStop`. Stop the DSD manager before stopping its BDD manager.
- A manager and its underlying BDD manager require external synchronization.
  Separate managers can run concurrently. Decomposition caches and overlap paths
  are manager-local. Compatibility check-cache lifecycle functions manage a
  synchronized independent cache.

The variable order must remain fixed while a DSD manager is in use, since its
stored child order follows BDD levels. Dynamic slices replace the original
`MAXINPUTS` scratch buffers. Invalid dimensions/functions panic rather than
perform unchecked pointer accesses. Printing preserves the decomposition syntax;
missing names receive readable defaults, and large complemented primes retain
their complement marker.

Run `go test ./bdd/dsd -race` and `go vet ./bdd/dsd` from `abcgo`. The differential
test compiles the original C implementation when `cc` is available. The C oracle
corrects a one-character cache-initialization defect in a temporary copy of
`dsdCheck.c`: `Dsd_CheckCacheClear` must clear entry `i`, not entry zero repeatedly.
All decomposition and local-function algorithms remain the original C code.
