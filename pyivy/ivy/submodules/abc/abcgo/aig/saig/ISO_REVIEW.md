# SAIG isomorphism and output filtering review

## Readability and representation

The isolation routines use `aig.Graph`, `aig.Edge`, and ordinary Go slices.
Input ordering uses deterministic color refinement over primary-input/latch
roles, complemented edges, gate kind, sibling signatures, and fanout contexts.
SHA-256 fingerprints keep shared DAG signatures bounded in size. Canonical
duplication owns a new graph and does not use graph metadata or shared scratch
fields. Structural mapping translates edges recursively and matches typed node
keys without mutating either input graph.

## Bug review

Regression coverage checks an input permutation with a complemented fanin,
rejects an incorrect mapping, exhaustively compares all assignments of a
small sequential graph before and after canonical duplication, groups duplicate
outputs while preserving every latch input, and verifies negative-edge counts
include both gate and output edges. Deleted graph slots are skipped, and
complemented phases are retained through recursive mapping and duplication.

`Iso_ManFilterPos` reports structural candidate classes only. It does not claim
semantic equivalence for structurally different cones; callers must run their
SAT/CEC verifier before deleting an output. `Iso_ManTest888` depended on
mutation/conversion APIs in the higher ABC network package, so this graph-layer
port preserves its nil result and leaves that diagnostic to the ABC layer.

## Verification and scope

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./aig/saig/saigIso.go ./aig/saig/saigIso_test.go -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./aig/saig/saigIso.go ./aig/saig/saigIso_test.go
```

The whole SAIG package remains blocked by generated C-syntax files outside this
isolation family. `saigIsoSlow.go` now forwards the source entry point to the
native implementation; its former duplicate definition and generated C bodies
have been replaced, so the two files no longer define competing permutations.
