# ABC Rec3 source review

`abcRec3.go` was compared in source order with `src/base/abci/abcRec3.c`.
The exported Rec3 entry points and the active helper routines are all present.
The C iterator macros were translated into index loops over object, CO, PO,
and truth-table-node sequences while retaining the C ordering.

The truth-table node traversals preserve the postorder emitted by
`Gia_ObjCollectInternal`. That traversal needs the truth workspace initialized
by `Gia_ObjComputeTruthTableStart`; `Gia_ManEnsureTruthWorkspace` mirrors that
explicit lifecycle and rejects reuse with a different input count. The
read-only node snapshot and structural-hash state query avoid reaching into
private GIA manager fields. Mark reads and writes use the existing GIA mark
accessors.

The delay helpers retain the source's four-bit pin packing, max-plus-one
recurrence capped at 15, decrement behavior, and componentwise dominance.
Truth-table conversion between `word` and `uint32` uses native byte order to
match the C casts. The final truth-table ordering compares support size first,
then uses the native-byte ordering required by the C `memcmp` comparator.

Focused tests cover delay packing, max/decrement, equality and dominance. The
GIA bridge test covers workspace initialization/reuse, node postorder snapshot
semantics, and hash-table lifecycle. The GIA test passes.

`gofmt` and `git diff --check` pass for the changed files. The focused abci
tests are currently blocked before type-checking `abcRec3.go` by unrelated
missing type aliases in `abc.go` and `abcDar.go` (for example
`Cmd_CommandId_t`, `Dar_RwrPar_t`, and `Cnf_Man_t`).
