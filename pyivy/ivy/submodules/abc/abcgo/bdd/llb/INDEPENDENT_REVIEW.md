# LLB port review

This review was done after the native reachability, image, flow, cut, dump,
counterexample, and nonlinear helper passes.

## Bugs found and addressed

- `Llb_ManConstructGroupBdd` previously conjoined output variables without
  relating them to their AIG functions. It now builds the output equations and
  has a focused regression test.
- Nonlinear partition construction ignored the supplied variable order. Input
  BDD variables now use the requested node-to-variable mapping, with a
  permutation regression test.
- Nonlinear initial-state, quantification-cube, and state-recording helpers
  previously ignored state roles or returned no result. They now use explicit
  current/next latch edges and ordered BDD variable indices; a helper test
  covers the latch mapping.
- Reachability's counterexample handoff used a removed graph field. It now uses
  the graph's validated sequential-counterexample API and replays the witness
  in a test.

## Readability and design notes

- The primary reachability path represents the transition relation and
  reachable frontier directly with BDDs. It does not require a separate
  partition scheduler to understand the core image operation.
- Packed state data is represented as `[]uint32`, and counterexamples are
  verified by replaying each frame against the AIG.
- The older variable-pairing setup hook remains empty because the native path
  constructs the transition relation directly and does not use CUDD's legacy
  variable-pairing state.

## Validation status

`gofmt` succeeds for the changed LLB files. Ordinary `go test ./bdd/llb` and
`go vet ./bdd/llb` currently stop while parsing the imported `aig/gia` package
(`giaCSat.go` and `giaCSat2.go` syntax errors in the latest runs), before the
LLB package can be type-checked. No full-package green result is claimed.
