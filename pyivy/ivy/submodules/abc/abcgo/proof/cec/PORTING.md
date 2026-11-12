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
