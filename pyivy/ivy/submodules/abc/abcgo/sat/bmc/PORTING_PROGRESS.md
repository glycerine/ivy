# BMC porting progress

## Current status

Last updated: 2026-11-13. The expanded test coverage was committed in
`aa7a0ee8c`.

The `abcgo/sat/bmc` package builds and its Go test suite passes. The remaining
C-derived syntax in the package was translated, including the bounded-core,
induction, fault-analysis, mesh, majority and exact-synthesis engines. This is
package-level progress; it does not establish that the rest of `abcgo` builds or
that every ABC feature and option is behaviorally identical to C.

An uncached verification from the `abcgo` directory passed before the latest
test expansion:

```sh
GOCACHE=/tmp/abcgo-cache go test -count=1 ./sat/bmc ./sat/bsat
```

Both packages passed. The following whole-package checks also passed after the
expanded tests:

```sh
GOCACHE=/tmp/abcgo-cache go test -race -count=1 ./sat/bmc
GOCACHE=/tmp/abcgo-cache go vet ./sat/bmc
```

The expanded algorithm and entry-point suite passed under the race detector
with 79.9% statement coverage. Coverage is a measure of executed statements,
not proof of correct results or coverage of every option and input.

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
  `bmcWitnessOracle_test.go` adds 192 two-register, two-input configurations.
- `bmcDifferential_test.go` compares three deterministic fixtures with C ABC
  and independently replays the Go counterexamples. It uses `ABC_BINARY` if
  set, otherwise a local source-tree `abc` executable; the test skips if no
  executable is available. The fixtures passed against the local executable.
  `bmcDifferentialExpanded_test.go` adds five fixtures for multi-input and
  multi-register circuits, complemented logic, multiple outputs, and delayed
  failures; an explicit-state oracle checks frames and replays witnesses. All
  eight fixtures passed against the local C ABC executable.
- `bmcICheckOracle_test.go` compares induction SAT results against explicit
  state-pair exploration for 1,024 small cases.
- `bmcCexCareOracle_test.go` replays completions of reported care sets and
  minimization reasons on small multi-frame circuits.
- `bmcFaultEndToEnd_test.go` checks emitted suites for all five fault models,
  both delay-fault modes, cardinality options, and extension of prior patterns.
  `bmcFaultMultiGate_test.go` independently checks two-gate combinational and
  sequential fault models and their emitted test suites.
- `bmcExa3Oracle_test.go` and `bmcZyxOracle_test.go` independently evaluate
  synthesized networks and compare selected small cases with exhaustive
  truth-table or network enumeration. `bmcZyxWideOracle_test.go` verifies
  seven-input functions across the 64-bit truth-table word boundary.
- `bmcFxOracle_test.go` checks onset coverage and offset exclusion for all 256
  three-input functions and a two-output case. `bmcMeshOracle_test.go` checks
  placement witnesses from both solver backends and impossible small layouts.
  `bmcMeshFanoutOracle_test.go` adds reconvergent fanout and internal buffering;
  the existing mesh test oracle was corrected to allow valid gate buffers.
  The expanded cases found no implementation bugs.

Further algorithm tests cover bounded-checking and unrolling entry points,
ternary simulation, CEX target and care analysis, induction entry points,
MNIST parsing, exact-synthesis variants, and exhaustive two-input collapse,
SOP expansion, and ECO feasibility. CNF variable lifting is checked across a
round trip. These checks cover additional paths but do not establish correctness
for every possible graph, timeout, or option combination.

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
