# SAIG slow isomorphism implementation review

## Readability and representation

The active source entry point `Saig_ManFindIsoPerm` now forwards to the native
deterministic color-refinement implementation in `saigIso.go`. This replaces
the private C implementation manager, packed `Iso_Obj_t` records, hand-built
hash bins, and 1,024-entry generated prime table. The Go implementation uses
typed graph edges and compact SHA-256 fingerprints, with no mutable package
globals.

`Iso_ManObjCount` remains available as an idiomatic helper returning its counts
directly. It uses a per-call visited map and therefore does not leave traversal
IDs or scratch marks on the graph.

## Bug review

Tests verify the input-order wrapper, complemented-input mappings, canonical
duplication across all assignments of a small sequential graph, duplicate
output grouping, and unique cone-node/complemented-edge counts. Race tests
confirm independent calls do not share or mutate isomorphism scratch state.

The old `Iso_ReadPrimes` and `Iso_FindNumbers` routines only generated/debugged
the prime table for the replaced private implementation; no caller in ABC's
active build uses them, so that debugging surface was removed. The old
`Iso_ManStart`/`Stop`, hashing, class-chain, and tie-breaking helpers are likewise
replaced by the Graph-layer implementation. No generated C-draft bodies remain
in this file.

## Verification and scope

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./aig/saig/saigIso.go ./aig/saig/saigIsoSlow.go ./aig/saig/saigIso_test.go -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./aig/saig/saigIso.go ./aig/saig/saigIsoSlow.go ./aig/saig/saigIso_test.go
```

The full SAIG package remains blocked by other generated C-syntax files.
