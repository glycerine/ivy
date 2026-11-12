# Root CUDD bridge family

`cuddBridge.go` is fully ported as thin wrappers around `bdd/cudd/core`.
Public thresholds, strict thresholds, intervals, integer bits, BDD-to-ADD,
ADD pattern, and manager transfer share the authoritative core implementations.
No original algorithms are duplicated, and no source files are hidden.

## Readability review

Kept original public names and the root API's int32 bit parameter. Retained the
original copyright/license header. Removed C-private pointer/cache recursion
scaffolding after confirming no other root file calls those static helpers.
Live internal conversion/transfer entrypoints remain forwarding functions.

## Bug review

`cuddBridge_test.go` verifies every public forwarder and cross-manager transfer.
Core tests independently check numeric truth values, wrong family/ownership,
NaN comparison behavior, bit boundaries, truncation, invalid conversion inputs,
and shuffled variable orders. The original-C checksum fixture is documented in
`core/BRIDGE_REVIEW.md`. Invalid operations that are undefined in C fail with nil.

Validation of this completed root subset:

```
GOCACHE=/tmp/abcgo-go-cache go -C abcgo test bdd/cudd/cuddBridge.go bdd/cudd/cuddBridge_test.go bdd/cudd/cudd.go bdd/cudd/bddCore.go -race -count=1
```

Full root `go test ./bdd/cudd` still fails on the visible unported `cuddAPI.go`
and subsequent draft families. Core remains independently tested; forwarding
this family does not mark the full root CUDD package complete.
