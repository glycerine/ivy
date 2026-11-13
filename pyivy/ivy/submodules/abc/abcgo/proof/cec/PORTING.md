# CEC port status

The circuit model and incremental SAT equivalence group is implemented in
`core/`, which is independently buildable and uses the Go Satoko solver. GIA can
use the dependency-neutral Network interface without an import cycle. This is a
real CNF/CDCL engine, with independently validated counterexamples and original
CEC result conventions; see core/README.md for APIs and verification.

The parent CEC package remains incomplete. Its generated files have not been
hidden or excluded. Remaining groups include the original pattern/simulation
classes and refinement, SAT-sweeping/FRAIG orchestration, choices/correspondence,
sequential/semi-formal drivers, synthesis and circuit-based SAT alternatives,
and legacy parameter/manager adapters. A complete parent-package build remains
a required integration gate; a green core test is not evidence of that gate.

Completed group verification: actual `go test -race -cover ./proof/cec/core`
passes (85.4% statement coverage), and `go vet ./proof/cec/core` passes. Separate
readability and bug reviews are recorded in core/README.md. No git commands were
used by this worker.

## Upstream Cec4 choice limitation

The choice-building path `gia.Cec4_ManSatSolverChoices` is not a working
implementation when independent choice classes are present: its recursive AND
builder returns constant zero. This is inherited from the original C routine
`src/aig/gia/giaEquiv.c` (`Cec4_ManSatSolverAnd_rec`), which also immediately
returns zero. The Go code preserves that behavior; a correct choice builder
requires an implementation and independent equivalence tests. A build or a
passing test of other Cec4 paths does not establish correctness of this path.
