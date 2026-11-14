# Ivy conversion tail review

Compared `abcIvy_tail.go` with `src/base/abci/abcIvy.c` from
`Abc_NtkFromIvy` through `Abc_NtkCollectLatchValuesIvy`.

## Source correspondence

- `Abc_NtkFromIvy` seeds the constant and CI mappings, processes Ivy DFS order,
  preserves complemented fanins and XOR nodes, and connects COs in source
  order.
- `Abc_NtkFromIvySeq` creates latch BI/BO pairs in Ivy latch order, maps
  initial states, reconstructs logic and optional choice chains, then connects
  POs and latch inputs in source order.
- `Abc_NtkToIvy` creates PIs before traversing the source network, preserves
  complemented strashed fanins and PO drivers, and runs Ivy cleanup.
- The SOP path retains constant, cube, XOR, complemented-result, and factoring
  behavior. `CopyData` stores each mapped Ivy signal, including its polarity;
  this replaces C's untyped `pCopy` pointer storage.
- Latch initial values are collected in network order, with `useDCs` taking
  precedence over stored zero/one values, as in C.

## Representation and scope

Go's `abc.Edge` carries complement bits explicitly. Conversion therefore stores
encoded `abc.Edge` values in Ivy traversal IDs and passes edge values through
the fanin API. The decomposition package's `Dec_GraphToNetworkIvy` is used for
factoring rather than duplicating its graph traversal.

The implementation file is 277 lines; the C source range is about 407 lines.
The shorter Go version drops C documentation and macro scaffolding while
keeping the routines and control-flow branches.

## Verification

Focused tests cover combinational roundtrip truth tables, CI/CO order and names,
complemented outputs, SOP XOR strashing, and latch initialization order/DC
handling. An isolated harness using the actual `base/abc`, `ivy`, and `dec`
packages passes all three tests. `gofmt` and `git diff --check` pass. The full
`base/abci` test is currently blocked by parser errors in `abcMap.go`, before
this package can be type-checked.
