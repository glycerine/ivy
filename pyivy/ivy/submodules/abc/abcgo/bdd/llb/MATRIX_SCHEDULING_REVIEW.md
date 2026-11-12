# LLB dependency matrix and scheduler review

## Ported scope

`llb1Matrix.go`, `llb1Sched.go`, and `llb1Cluster.go` now use Go slices for the column-major dependency matrix, row/column counts, partition pointers, and partial-product state. The matrix verifies cached counts, orders variables by their final partition, schedules partitions, scores clustering choices, merges columns, and compacts the result. `llbAliases.go` supplies the minimal package aliases/adapters needed by these routines; `Graph.RegisterCount` exposes the register count needed to initialize matrix dimensions.

## Separate bug/readability pass

- Repeated variables in both a group's input and output lists must only increment row/column sums once because a matrix cell represents membership. `Llb_MtrAddColumn` now checks the cell before changing cached counts.
- Empty rows previously gave a `0/0` average lifespan; the print path now reports zero for an empty matrix.
- `Llb_ManComputeCommonAttr` could divide by zero for a pair with no combinational support; it now returns zero.
- Matrix/group payloads move together when scheduler columns swap; a focused regression checks the matrix column, count, and group pointer.
- Matrix and partial-product storage is owned by Go slices; freeing releases references by nil-ing the slices, with no manual allocation/free pair.
- The variable-order algorithm uses stable adjacent swaps, preserving source tie ordering.

## Verification

This is an explicit-file-list test gate, not a package test. It intentionally compiles only the native matrix/scheduler source set because the LLB package still contains untranslated drafts. The exact command is:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 \
  bdd/llb/llbAliases.go bdd/llb/llb.h.go bdd/llb/llbInt.go \
  bdd/llb/llb1Matrix.go bdd/llb/llb1Sched.go \
  bdd/llb/llbMatrixNative_test.go bdd/llb/assert.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet \
  bdd/llb/llbAliases.go bdd/llb/llb.h.go bdd/llb/llbInt.go \
  bdd/llb/llb1Matrix.go bdd/llb/llb1Sched.go \
  bdd/llb/llbMatrixNative_test.go bdd/llb/assert.go
```

Both explicit-file gates pass for the matrix and scheduler files only. `llb1Cluster.go` was rewritten and gofmt-parsed, but is not type-checked by this gate because its dependency `llb1Group.go` itself has C-derived syntax errors. `GOCACHE=/tmp/abcgo-cache go -C abcgo test ./bdd/llb` still fails on C-derived syntax in `llb1Constr.go` (first at lines 74 and 79), after this work removed syntax errors in `llb1Cluster.go`, `llb1Matrix.go`, and `llb1Sched.go`. Therefore the full LLB package is not yet buildable or testable.
