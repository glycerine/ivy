# RWT independent review

## Scope and review method

Reviewed the complete native Go port in `rwt.go`, `rwtMan.go`, `rwtUtil.go`,
`rwtDec.go`, and the copied source tables in `rwt_tables.go`. Compared manager
startup, packed forest decoding, canonical-class insertion, traversal-volume
counting, decomposition collection, and practical-class lookup against
`src/opt/rwt/rwt.h`, `rwtMan.c`, `rwtUtil.c`, and `rwtDec.c`.

## Bug review

The review found one concrete defect before final verification: the practical
class array's first element is a sentinel zero, while C starts iterating at
array index one and separately sets class zero. Starting the Go iteration at
zero had stopped immediately and omitted every practical class. The Go loop
now begins at index one, and `TestCanonicalLookupForFourInputVariable` checks
class zero and the first listed classes.

No additional correctness issues were found. Complemented edges use explicit
Go edge handles instead of pointer tagging. Preprocessed decomposition graphs
are checked against each stored truth table, and the packed C rewrite-table
contents are checked by an original-C checksum oracle.

## Readability and compactness review

The port uses slices for the forest, canonical table, NPN permutations, and
class members; Go manages node and graph lifetimes. Mutable manager counters
and temporary result state remain on the manager because the C caller's rewrite
algorithm shares that state. Global NPN tables are initialized under a mutex
and are immutable while managers use them. The rewrite table remains compact
`uint16` data copied from the original source.

The exported `Rwt_ManCanonInfo`, class/decomposition accessors, and
`Rwt_ManStoreBest` hooks let the Ivy caller use the package without reaching
into unexported RWT fields.

## Verification

All commands were run from the repository root:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./opt/rwt
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./opt/rwt
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./opt/rwt
```

All pass. The oracle compares the Go packed table count and checksum against
`rwtUtil.c`; the loader test separately checks the decomposed truth of every
precomputed rewrite node. Ivy rewrite-pre is the dependent integration caller
and must pass its own package-level gate after it adopts these APIs.
