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
both optional interpolation algorithms, constrained witness replay in every
frame, multiple constraints, and native XOR/MUX gates. Constrained witness
construction propagates the deadline, with checks between graph construction,
CNF encoding, SAT solving, and witness extraction; individual graph traversals
and CNF encoding are not interruptible. Larger state spaces, interrupted
SAT solving, resource limits during rewriting, and independent validation of
more proof structures remain useful areas for additional coverage.

The optional MiniSat 1.14p interface from `intM114p.c` uses the Go port of
the original proof-producing solver in `sat/minisatp`. ABC's normal C build
excludes this optional source; the ordinary Go path uses the internal BSAT
proof engine. The MiniSat path observes conflict and time limits, and returns
unknown when interrupted. Interrupted MiniSat solver instances are terminal.

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

ABC's `src/sat/psat/m114p.h` declares a C wrapper, but the wrapper
implementation is absent from this checkout. The Go adapter tracks input root
indices and reconstructs a root-first proof, retaining the final contradiction.
Tests check skipped roots, interleaved derivations, terminal empty roots, both
interpolants, resource limits, and small sequential systems. The solver tests
compare randomized formulas with exhaustive assignments and replay UNSAT proofs;
an optional test compares results with the official MiniSat-p executable.
Broader original-C comparisons and larger proof workloads remain useful.
