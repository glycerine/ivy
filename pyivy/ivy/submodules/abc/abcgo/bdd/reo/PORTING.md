# REO port status

The package is a line-by-line port of `src/bdd/reo` (reo.h, reoApi.c,
reoCore.c, reoSwap.c, reoSift.c, reoProfile.c, reoTransfer.c, reoUnits.c,
reoShuffle.c, reoTest.c): the diagrams are copied into units grouped by
planes, adjacent planes are swapped in place (with the node, width and APL
cost bookkeeping of reoSwap.c), sifting follows reoSift.c, and the result is
built back in the CUDD manager node by node (`cudd.Cudd_UniqueInter`).

Differences from C, none of which changes a result:

- A unit pointer is a `uref`: the unit index shifted left by one, with the
  complement in bit 0. Units live in one slice; a `*reo_unit` is valid only
  until the next `reoUnitsGetNextUnit`.
- Hash tables have a power-of-two size of at least C's `3*nNodesMax+1`
  entries (still lossless); keys hash CUDD node IDs and unit indexes.
- `reoTransferNodesToUnits_rec` hashes every node (C skips nodes with
  `ref == 1`, which are visited once).
- The assertions C compiles out under `-DNDEBUG` (reoCheckLevels,
  reoRecursiveDeref, reoProfileWidthVerifyLevel, cost checks) run only with
  `reoDebug`.
- Support and sharing size are computed in one traversal.

As in C, an ADD terminal value is truncated to an integer when the diagram
is copied into units (`(int)cuddV(F)`), so ADDs with fractional terminals
come back with integer terminals; no caller reorders such ADDs.

`reo_golden_test.go` compares DAG sizes and orders with the native engine
for node, APL, width (ADD), remapping and multi-function reordering;
`base/abci/reo_native_test.go` compares renode -b, reorder and eliminate -s
with native ABC.
