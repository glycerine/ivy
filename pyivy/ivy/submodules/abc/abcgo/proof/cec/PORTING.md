# CEC port status

The `proof/cec` package builds, including the original CEC, Cec2, Cec3, and
Cec4 entry points. An independent full `go test -race ./proof/cec` and
`go vet ./proof/cec` passed after integration. The dependency-neutral `core/`
package contains the shared SAT equivalence machinery; its own README records
its APIs and tests. Focused tests also cover Cec4 simulation, pattern packing,
Glucose solver modes, reduced-graph equivalence and counterexample replay.

Passing the package tests establishes a build and the tested behavior; it does
not establish complete equivalence with ABC on all inputs. The current source
fidelity review includes Cec4 sweeping, counterexample handling and diagnostics.
In particular, work is in progress to match the source's detailed per-round
statistics and timings. The focused harnesses are supplemental, not a
replacement for the full-package test.

## Upstream Cec4 choice limitation

The choice-building path `gia.Cec4_ManSatSolverChoices` is not a working
implementation when independent choice classes are present: its recursive AND
builder returns constant zero. This is inherited from the original C routine
`src/aig/gia/giaEquiv.c` (`Cec4_ManSatSolverAnd_rec`), which also immediately
returns zero. The Go code preserves that behavior; a correct choice builder
requires an implementation and independent equivalence tests. A build or a
passing test of other Cec4 paths does not establish correctness of this path.
