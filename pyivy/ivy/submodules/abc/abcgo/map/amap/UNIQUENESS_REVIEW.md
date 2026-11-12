# Amap library-node uniqueness review

`amapUniq.go` is now native Go. It maintains the source library's compact
lookup tables for AND/XOR fanin pairs and ordered four-integer mux records.
Rule vectors remain sorted by their low 16-bit fanin key while the high 16 bits
carry the mapped node ID. The exported lookup-table materialization returns
independent zero-terminated Go slices instead of a manually allocated C pointer
matrix.

## Separate bug/readability pass

- The masked lookup and ordered insertion both compare only the low 16 bits;
  tests verify node IDs in the upper half survive insertion and lookup.
- Every created structural node extends both polarity rule rows for AND and
  XOR, as the C manager does. Mux rows keep the exact source fanin order.
- Library node storage is a Go slice with a capacity estimate that doubles as
  needed. Pointers are only used during one insertion call; persistent graph
  relationships are stored as compact IDs, so slice growth does not invalidate
  stored references.
- Node IDs are checked before packing into 16 high bits. Invalid rule-table
  indices fail at insertion rather than indexing another polarity row.
- Lookup-table rows are copied and explicitly terminated. A test mutates a
  returned row and verifies the source vector remains unchanged.

## Verification scope

The explicit source-set race test and vet pass:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race \
  ./map/amap/amap.go ./map/amap/amapInt.go ./map/amap/amapCompat.go \
  ./map/amap/assert.go ./map/amap/amapUniq.go ./map/amap/amapUniq_native_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet \
  ./map/amap/amap.go ./map/amap/amapInt.go ./map/amap/amapCompat.go \
  ./map/amap/assert.go ./map/amap/amapUniq.go ./map/amap/amapUniq_native_test.go
```

These commands compile only the listed files. They are not the normal package
gate; `go test ./map/amap` still fails on the independently owned generated
`amapLib.go` draft.
