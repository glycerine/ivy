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

## MiniSat 1.14p source and remaining work

The upstream [MiniSat page](https://minisat.se/MiniSat.html) lists the
[MiniSat-p 1.14 source archive](https://minisat.se/downloads/MiniSat-p_v1.14.2006-Sep-07.src.zip)
as the proof-logging version and states that MiniSat is MIT licensed. A
separate license grant is not needed to port it; retain the applicable
copyright and MIT permission notices from the actual source being ported.

The original archive has also been retrieved from the
[upstream website repository](https://github.com/niklasso/minisat-webpage/tree/master/downloads).
It includes the solver, proof implementation, headers and MIT license. The
[emina/minisatp tree](https://github.com/emina/minisatp) describes a patched
and extended version; its compatibility with ABC has not been verified.

Remaining work is to port the selected solver and its proof recording faithfully,
preserving source attribution, and adapt its resolution roots and chains to
`M114pBackend`. ABC's `src/sat/psat/m114p.h` declares a C wrapper, but the
wrapper implementation is absent from this checkout, so its proof enumeration
and root numbering must be reconstructed and checked against the interpolation
consumers. Test SAT, UNSAT, contradictory roots, multi-step resolution, root
ordering, and conflict counts against the original implementation, and validate
each proof and both interpolants independently. Until that work is complete,
the explicit unavailable result remains intentional.
