# `proof/int` port status

The ordinary ABC interpolation path has been translated to Go and
`go test -race ./proof/int` and `go vet ./proof/int` pass. The package includes
graph transformations, frame unrolling, counterexample extraction, containment
checks, incremental induction, BSAT proof interpolation, DAR rewriting, and
the top-level loop. Exported parameters allow other packages to configure it.

Tests compare all 16 one-register transition functions with independent
state exploration; replay a two-register, three-frame witness; exercise
constraint polarity and constrained witnesses; and cover forward, backward,
combinational containment, transition-loop, rewriting, and two-step modes.
Other tests cover SAT and UNSAT containment, frame limits, proof construction,
and both optional interpolation algorithms. Larger state spaces, interrupted
SAT solving, resource limits during rewriting, and independent validation of
more proof structures remain useful areas for additional coverage.

The optional MiniSat 1.14p interface from `intM114p.c` has been translated
against an explicit proof-producing backend interface, but the original
MiniSat 1.14p backend has not been ported. Explicitly selecting that missing
backend reports `ErrMiniSat114pUnavailable` and returns unknown; no unrelated
solver is selected in its place. ABC's normal C build excludes this optional
source. The ordinary Go path uses the internal BSAT proof engine.
