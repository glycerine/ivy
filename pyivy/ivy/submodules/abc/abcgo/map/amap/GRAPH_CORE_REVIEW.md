# Amap graph/core native port review

This review covers `amapCore.go`, `amapGraph.go`, `amapInt.go`, and
`amapMan.go`, plus the compatibility helpers in `amapCompat.go`. The implementation
now stores internal mapper nodes as Go-owned objects and copies the source AIG
through its stable `aig.Graph` edge API instead of relying on C pointer tagging,
`pData` mutation, or a byte arena for Go structs.

## Correctness and readability pass

- `Amap_ManCreate` uses a per-call node-ID map, so it does not corrupt temporary
  source graph data or rely on address arithmetic to preserve complemented
  edges.
- `Amap_AndToObj` reapplies each edge's complement bit when the mapped child is
  read. AMap's own complemented objects are explicit wrappers with a stable
  regular-object reference; fanin literals are encoded in the existing compact
  variable/sign representation.
- AND, XOR, mux, and PO construction update fanout references, phase, and level
  from their logical inputs. The mux stores `no`, `yes`, then `control`, matching
  the source's phase equation and fanin accessors.
- Choice linking checks for cycles when finding the class tail. AMap manager
  stop drops only its own vectors, pools, and nodes; it leaves the selected
  library and input AIG owned by their callers.
- The frame-selected library remains available through `Amap_ManTest`, while
  `Amap_ManMapGraph` exposes an explicit-library API for callers without global
  frame state.

## Focused verification

The graph-core tests compare all eight assignments of an input AIG containing
XOR, mux, complemented fanins, and AND against the copied Amap outputs. This is
an explicit-file-set gate because the parent package still contains other
unported C-derived files; it does not prove `go test ./map/amap` passes.

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race \
  ./map/amap/amap.go ./map/amap/amapInt.go ./map/amap/amapCompat.go \
  ./map/amap/amapMan.go ./map/amap/amapGraph.go ./map/amap/assert.go \
  ./map/amap/amapGraph_native_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet \
  ./map/amap/amap.go ./map/amap/amapInt.go ./map/amap/amapCompat.go \
  ./map/amap/amapMan.go ./map/amap/amapGraph.go ./map/amap/assert.go \
  ./map/amap/amapGraph_native_test.go
```

The ordinary package command is still blocked by the separately owned library
reader and the remaining matcher, merge, output, permutation, reader, rule, and
uniqueness ports. Do not treat the explicit-file gate as package completion.
