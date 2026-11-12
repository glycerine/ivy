# Ivy multi-input rewrite review

## Scope

Native ports of the original `ivyMulti.c` and `ivyMulti8.c` algorithms live in
`ivyMulti.go` and `ivyMulti8.go`. The caller in `ivyRwrAlg.go` now passes its
owning manager to `Ivy_MultiPlus`, and `ivyMulti_test.go` covers functional
equivalence, a previously hashed subexpression, and invalid input counts.

## Separate bug and readability pass

- Replaced C function-static workspaces with call-local Go slices/arrays. This
  prevents independent managers from racing through shared candidate state.
- Fixed the source's candidate marking typo in `Ivy_MultiPlus`: the discovered
  table node receives the new traversal index, instead of overwriting the last
  node visited in the cone.
- Preserve all cone support masks in a separate slice before reusing traversal
  IDs. The source-style second pass rewrites `TravId` as it walks; subsequent
  fanins must still read their original cone mask.
- Kept candidate masks unsigned and used `math/bits` for population counts,
  including the 32-leaf boundary in the plus-cover algorithm.
- Validated manager ownership and slice bounds at the public operation
  boundary; ghost lookup skips degenerate pairs that the Go graph model rejects.
- Kept the 15-argument search bound from the original `ivyMulti8.c`; larger
  inputs use the existing correct level-aware recursive balance path instead
  of overflowing the C fixed arrays/bitfields.
- `Ivy_MultiCover` preserves the original greedy objective: maximize newly
  covered leaves, then prefer lower levels, and fall back to a leaf when only
  one uncovered input remains.

## Validation status

`gofmt` parses all changed Go files. The full `go test ./aig/ivy` gate is still
blocked by unrelated source-derived drafts, currently `ivyRwr.go` and other
Ivy files; `opt/rwt` also has an existing raw `#define` parse error. Do not
interpret this scoped work as evidence that the full Ivy package builds.
