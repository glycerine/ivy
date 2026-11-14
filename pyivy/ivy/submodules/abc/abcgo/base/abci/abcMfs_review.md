# `abcMfs.go` source review

Compared the Go port with `src/base/abci/abcMfs.c` function by function.

- The first extractor retains C's DFS node order; the I-check extractor retains
  object-ID order. Both preserve the C temporary-ID layout for CIs, nodes, and
  COs, including holes when identifying fixed nodes by object ID.
- Small SOPs use the same 64-bit truth word. Large SOPs use C's 16-row truth
  store, word offsets, constant detection, and exact fanin order.
- Rewritten SFM fanin/truth storage is converted through the existing
  `sfm.Sfm_NtkConstruct` API, whose Go representation uses slices instead of
  C vectors. Truth values, fixed flags, and node numbering are preserved.
- Reinsertion maps SFM IDs through each ABC object's temporary ID, removes
  old fanins before adding optimized fanins, and regenerates SOPs with the
  network's flex manager.
- Sequential unrolling retains the source frame bounds, CI/CO/PO ordering,
  latch-output copy transfer, `iPivot` definition, final flop-output filter,
  and reverse mapping of optimized internal nodes.

Focused tests cover all truth-store rows, the seven-input large-truth path,
and object-ID assignment. The isolated file harness and its tests pass. Full
`go test ./base/abci` remains blocked before type-check by raw C syntax in the
independent `abcSaucy.go` translation unit.
