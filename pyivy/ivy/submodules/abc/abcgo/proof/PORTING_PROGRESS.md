# Proof package porting progress

The goal is a source-faithful Go port of every proof package, followed by
correctness review, simplification, and comprehensive tests. A successful
component test is not evidence that its whole package builds.

All proof packages now build and pass `go test -race -count=1 -p 2 ./proof/...`.
Correctness review, end-to-end coverage, and the fidelity gaps below remain
in progress; passing tests do not establish that those gaps are resolved.

* `acec` builds and passes its full race suite and vet. Arithmetic detector,
  adder-box reconstruction, and polynomial tests are in place. Independent
  correctness review and broader end-to-end equivalence coverage continue.

* `cec` builds and passes its full race suite and vet. Its top-level verifier
  now uses the source-style sweep and FRAIG fallback pipeline, with original-C
  comparisons and witness replay; further validation remains. Cec4's upstream
  GIA choices implementation is itself incomplete. See `cec/PORTING.md`.
* `pdr` builds and passes its full race suite and vet. It has exhaustive
  small-state reachability and witness tests, incremental-solving tests,
  and a Go-facing options API. Invariant publication uses synchronized
  global storage with last-writer-wins semantics; callers needing a specific
  run's invariant must serialize publishers and read it before the next one.
* `ssw` builds and passes its full race suite and vet. The original
  semi-formal implementation asserts because its frame migration is
  unfinished; representative storage requires a buffer-free graph. Rarity
  permutation uses Go's PRNG, so its sequence differs from C's `rand()`.
  Review fixed uniqueness polarity, constant-candidate counts and polarity
  of unassigned edges; partitioned and constrained traces now have tests.
* `abs` builds and passes its full race suite and vet. The old GLA driver
  lowers native XOR/MUX nodes to ANDs and maps results back to the original
  graph; full integration review remains in progress.
* `fra` builds and passes its full race suite and vet. SEC has small-machine
  reachability and witness-replay tests with correspondence enabled; those
  tests do not establish which cases enter every internal SSW stage.
* `fraig` builds and passes its full race suite. Review has fixed constant
  traversal, equivalence polarity, and multi-output miter bugs; SAT truth
  tables, feedback, compression, and external options have tests. Review
  and coverage of less common paths continue.
* `int` builds and passes its full race suite and vet. Its normal BSAT path
  has reachability, constraints, witness, and interpolation tests. The
  optional MiniSat 1.14p proof backend is now ported and connected, with
  proof replay, original-solver comparisons, and interpolation regressions;
  broader validation continues. See `int/PORTING.md`.
* `int2` builds and passes its full race suite, but its driver is a bounded
  fallback rather than a port of the incomplete checked-in C driver. See
  `int2/SOURCE_FIDELITY.md`.
* `live` uses the ported PDR solver by default. Its safety engines and
  monitor reductions preserve constraints; the bounded image checker is
  retained for an explicit state limit. See `live/PORTING.md`.
* `dch`, `ssc`, and `cec/core` build. Their existing review and
  package-specific documentation describe coverage and any fidelity gaps.

Run full-package checks from `abcgo`, for example `go test -race ./proof/pdr`.
Do not include experimental draft directories in Go test runs.
