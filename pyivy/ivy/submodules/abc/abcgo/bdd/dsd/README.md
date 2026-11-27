# Disjoint-support decomposition

This package is a line-by-line port of ABC's `src/bdd/dsd` (dsdApi.c,
dsdMan.c, dsdProc.c, dsdTree.c, dsdLocal.c, dsdCheck.c) onto
`bdd/cudd/core`. `dsd` (base/abci/abcDsd.go) and `&collapse`'s rebuild use
it, and they produce the same trees, statistics and printouts as native ABC.

Adaptations:

- `DdManager` and `DdNode` alias the `cudd/core` types.
- C tags the low pointer bit for a complemented DSD node; Go uses a
  canonical inverse handle. Use `Dsd_Regular`, `Dsd_Not`, and boolean
  `Dsd_IsComplement`/`Dsd_NotCond`.
- Pointer arrays become slices (`nDecs` is `len(pDecs)`), files become
  `io.Writer`, and the user mark is an `int` (callers store object ids or
  literals where C stores pointers).
- CASE 4 of `dsdKernelDecompose_rec` marks components with its own
  `kernelMark` field. C shares `Mark` with the user; its user marks are
  pointers that never equal the small `s_Mark` counter, so a separate field
  behaves the same for Go's small integer marks.
- dsdProc.c's static counters and dsdCheck.c's identity cache belong to the
  manager, so independent managers can run concurrently. The manager-less
  `Dsd_CheckRootFunctionIdentity` uses a mutex-guarded cache.
- CASE 4's PRIME entry list grows as needed instead of asserting
  `nEntries <= nInputs - level(top var)`, which only a declared input set
  below unrelated variables could exceed.
- `Extra_dsdRemap` caches every node, not only those whose reference count
  differs from one; the result is the same.

Run `go test ./bdd/dsd`. The differential test compiles the original C
implementation when `cc` is available and compares the decomposition trees.
