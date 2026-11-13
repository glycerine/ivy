# BMC porting progress

## Current status

The `abcgo/sat/bmc` package builds and its Go test suite passes. The remaining
C-derived syntax in the package was translated, including the bounded-core,
induction, fault-analysis, mesh, majority and exact-synthesis engines. This is
package-level progress; it does not establish that the rest of `abcgo` builds or
that every ABC feature and option is behaviorally identical to C.

The most recent uncached verification from the `abcgo` directory was:

```sh
GOCACHE=/tmp/abcgo-cache go test -count=1 ./sat/bmc ./sat/bsat
```

Both packages passed. A previous `go test -coverprofile` run reported 73.5%
statement coverage for `sat/bmc`, before the latest independent oracle tests
were added. Coverage is a measure of executed statements, not proof of correct
results. A whole-package race and vet pass has not been recorded here.

## Correctness work and test evidence

The scalable BMC checker now counts each failed output once in `fSolveAll`
mode. Previously, repeated failures of one output on successive frames could
make the checker stop before examining a different output's later failure. It
also now returns a failing result if any output fails while another remains
safe. `bmcBmc3_native_test.go` contains regressions for both cases.

The suite includes the following independent or end-to-end checks:

- `bmcBmcOracle_test.go` compares small sequential circuits with an explicit
  state oracle, checking first failure frames, output identities, and witness
  replay. It includes 1,024 single-output and 512 two-output cases.
- `bmcDifferential_test.go` compares three deterministic fixtures with C ABC
  and independently replays the Go counterexamples. It uses `ABC_BINARY` if
  set, otherwise a local source-tree `abc` executable; the test skips if no
  executable is available. The fixtures passed against the local executable.
- `bmcICheckOracle_test.go` compares induction SAT results against explicit
  state-pair exploration for 1,024 small cases.
- `bmcCexCareOracle_test.go` replays completions of reported care sets and
  minimization reasons on small multi-frame circuits.
- `bmcFaultEndToEnd_test.go` checks emitted suites for all five fault models,
  both delay-fault modes, cardinality options, and extension of prior patterns.
- `bmcExa3Oracle_test.go` and `bmcZyxOracle_test.go` independently evaluate
  synthesized networks and compare selected small cases with exhaustive
  truth-table or network enumeration.
- `bmcFxOracle_test.go` checks onset coverage and offset exclusion for all 256
  three-input functions and a two-output case. `bmcMeshOracle_test.go` checks
  placement witnesses from both solver backends and impossible small layouts.

Other focused tests cover packed two-bit fields, formula fault unfolding,
SAT cardinality constraints, GIA simulation, bounded-core proof output, and
solver bridges. The 32-bit `Abc_InfoHasBit` remains in use for `[]uint32`
bitsets; 64-bit simulation words use `Abc_TtGetBit`.

## Remaining work

The small-model and differential tests do not cover every algorithm, option,
graph representation, timeout path, or input size. Expand C ABC differential
fixtures and independent witness checks as additional behavior is exercised.
The counterexample-minimization unate builder currently expects an
AND-inverter graph and rejects native XOR/mux nodes; mesh placement likewise
expects an ordered, nonconstant combinational AND graph. Preserve or explicitly
revisit these representation assumptions when integrating other callers.

The wider `abcgo` port, package-wide idiomatic-Go review, and separate bug
review remain outside the completed `sat/bmc` build milestone. Do not infer
completion of those tasks from the passing tests above.
