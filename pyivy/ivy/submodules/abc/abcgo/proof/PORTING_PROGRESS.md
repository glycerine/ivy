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

* `cec` builds and passes its full race suite and vet. Cec4 has been
  integrated, but its source-fidelity review continues; the upstream GIA
  choices implementation is itself incomplete, as documented in
  `cec/PORTING.md`.
* `pdr` builds and passes its full race suite and vet. It has exhaustive
  small-state reachability and witness tests, incremental-solving tests,
  and a Go-facing options API. Concurrent solves must be externally
  serialized because ABC invariant storage is global.
* `ssw` builds and passes its full race suite and vet. The original
  semi-formal implementation asserts because its frame migration is
  unfinished; representative storage requires a buffer-free graph. Rarity
  permutation uses Go's PRNG, so its sequence differs from C's `rand()`.
* `abs` builds and passes its full race suite and vet. Old GLA currently
  requires XOR/MUX nodes to be lowered to ANDs; full integration review
  remains in progress.
* `fra` builds and passes its full race suite and vet. SEC's nontrivial
  SSW and proof integrations still need end-to-end validation.
* `fraig` builds and passes its full race suite. Review has fixed constant
  traversal, equivalence polarity, and multi-output miter bugs; SAT truth
  tables, feedback, compression, and external options have tests. Review
  and coverage of less common paths continue.
* `int` builds and passes its full race suite and vet. Its normal BSAT path
  has reachability, constraints, witness, and interpolation tests. The
  optional MiniSat 1.14p proof backend is unavailable and selecting it
  returns unknown; see `int/PORTING.md`.
* `int2` builds and passes its full race suite, but its driver is a bounded
  fallback rather than a port of the incomplete checked-in C driver. See
  `int2/SOURCE_FIDELITY.md`.
* `dch`, `ssc`, `live`, and `cec/core` build. Their existing review and
  package-specific documentation describe coverage and any fidelity gaps.

Run full-package checks from `abcgo`, for example `go test -race ./proof/pdr`.
Do not include experimental draft directories in Go test runs.
