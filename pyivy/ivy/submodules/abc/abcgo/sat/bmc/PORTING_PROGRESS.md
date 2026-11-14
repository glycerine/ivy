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

The scalable BMC implementation now appends the sequential AIG frame by frame,
preserves prior node/variable IDs, adds only the new CNF suffix, and retains one
BSAT solver across output/frame queries. Finite-bound exhaustion returns `-1`
like C; return `1` is reserved for the small-state completeness cutoff. It also
implements source jump/conflict transitions, global/gap/per-output timeout
accounting, solve-all dropouts, and CEX storage/callback ownership. Tests pin
frame extension/CNF-prefix stability, persistent solver growth, `nStart`, and
CEX frame/PI indexing and replacement; backend tests cover BSAT, Satoko,
Glucose, and source precedence. The remaining BMC3 gaps are C's ternary
simulation, cut/hash CNF synthesis, dynamic cone worklists, bridge output and
same-model cross-output CEX propagation (details in `proof/int/PORTING.md`).

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

## Correctness and simplicity review

A subsequent inspection found and fixed several issues:

- The scalable BMC solver now applies the global deadline during SAT calls and
  accounts for cumulative per-output SAT time. The earlier implementation only
  checked the global limit between frames and restarted the per-output budget.
- Remapped GIA CNF now leaves outputs unasserted, as the C helper does, and
  preserves GIA object IDs even when inputs and gates are interleaved. Exhaustive
  tests cover complemented outputs and both input values.
- Cube extraction and fault analysis preserve their caller's SAT resource
  limits. Untestable-fault reporting counts nonempty rows actually written.
- The at-most-one constraint uses a balanced OR aggregation with one auxiliary
  per pair; exhaustive tests check both the constraint and aggregate value.
  The BMC clause loader also avoids rebuilding literals already in BSAT format.

The review changes passed `go test -race -count=1 ./sat/bmc ./sat/bsat
./sat/cnf` and `go vet ./sat/bmc ./sat/bsat ./sat/cnf`. These checks and the
inspection do not establish correctness for every graph or option.

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
