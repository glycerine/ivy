# AMAP cut merge review

## Implementation

`amapMerge.go` now stores cuts as Go-owned `Amap_Cut_t` values with a linked
`pNext` chain on each node. Temporary candidates are bucketed by encoded
library literal and combined with the library's AND/XOR/MUX rule tables. The
cut saver keeps the trivial cut first, keeps the source ordering for equal
result nodes, caps storage at `nCutsMax-1`, and copies fan slices so temporary
buffers cannot mutate saved cuts. Complemented singleton cuts are copied
before changing their fan literal.

## Separate bug review

- The old byte-arena layout is no longer assumed. This avoids unsafe pointer
  arithmetic and makes each cut's fan slice independently owned.
- Manager startup no longer allocates three C-style flex arenas or an unused
  match scratch table. Cuts and best-match copies use Go slices; each PI's
  trivial cut is stored directly on its node instead of duplicating the
  pointers in a manager side slice.
- `Amap_ManCutStore` checks the temporary bucket bounds before indexing it.
- Rule lookups use the packed-table sentinel and validate the input-literal
  index before access.
- MUX cut matching collects candidates for all three fanins before producing
  cuts. The binary path likewise checks complement-adjusted fan literals.
- Gate symmetry permutation copies now swap the variable mapping while
  preserving phase per destination pin (`amapSymmetryCopy`); a focused test
  covers this case.
- Tests cover binary cut construction, trivial-cut ordering, fan buffer
  ownership, and truncation at the source cut limit.
- The end-to-end MUX test exposed complemented proxy reads in cut traversal
  and output reference accounting. `Amap_CutHead` and `Amap_ObjRefsTotal` now
  regularize proxy pointers before reading shared state.
- `test_vectors/AMAP_MUX_ORACLE.md` records a native ABC parity run for the
  same genlib and BLIF inputs. With MUX matching enabled, both C and Go select
  the single `MUX2` cell.

## Verification

Normal package tests, race tests, and vet pass with `go -C abcgo test ./map/amap`,
`go -C abcgo test -race ./map/amap`, and `go -C abcgo vet ./map/amap`.
