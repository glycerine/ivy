# `proof/int` port status

The ordinary ABC interpolation path has been translated to Go and
`go test -race ./proof/int` passes. The package includes graph transformations,
frame unrolling, counterexample extraction, containment checks, incremental
induction, BSAT proof interpolation, DAR rewriting, and the top-level loop.
Tests cover simple safe and unsafe sequential circuits, counterexample traces,
SAT and UNSAT containment, proof construction, frame limits, and both optional
proof interpolation algorithms. More substantial sequential examples and
independent proof validation are still needed for confidence in all paths.

The optional MiniSat 1.14p interface from `intM114p.c` has been translated
against an explicit proof-producing backend interface, but the original
MiniSat 1.14p backend has not been ported. A missing backend returns
`ErrMiniSat114pUnavailable`; no unrelated solver is selected in its place.
ABC's normal C build excludes this optional source. The ordinary Go path uses
the original internal BSAT proof engine.
