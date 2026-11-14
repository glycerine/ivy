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

## Active source and option inventory

The active functions in `intCore.c`, `intMan.c`, `intDup.c`, `intFrames.c`,
`intContain.c`, `intCheck.c`, `intCtrex.c`, `intUtil.c`, and `intInter.c` have
Go counterparts. `intM114.c` uses BSAT, and the optional `intM114p.c` path
uses the MiniSat-p port. The original multi-frame uniqueness routine asserts
before its unfinished implementation; the Go counterpart preserves that limit.

Default parameters, forward interpolation, backward interpolation with a
transition loop, two-frame properties, transition loops, rewriting, global
variable bias, inductive and combinational containment, invariant dumping,
frame limits, and constraints have focused or reachability-oracle tests.
MiniSat's McMillan and Pudlak modes are checked on partitioned formulas and
sequential systems; its backward combination returns unknown because the C
implementation explicitly disallows it. `fUseOther` is passed by C to the
optional MiniSat step but is not used there. Multi-output/separate-output
dispatch is outside this single-property entry point, as in the C routine.
Both public interpolant-verification wrappers are exercised using a partitioned
UNSAT proof, and their shared containment diagnostics cover successful checks
and both failed implications. MiniSat partition tests reject each A-side clause
and the first B-side clause, checking early exit and restoration of lifted CNF.
The incremental containment checker is also compared against explicit
two-register state enumeration for two and three induction frames, all 16
state predicates, and several transition relations.
Forward and backward inductive containment are independently checked by
enumerating all one-register input traces through four frames, including
input-dependent transitions and both constant and nonconstant predicates.
The C `int` command and multi-output driver are in `base/abci`; the current
`abcgo/base/abci` files remain untranslated drafts, so they do not yet provide
a working Go command for this package. `proof/int` exposes the library entry
point used by ported Go callers.

Timed interpolation uses the calling thread's CPU clock on Linux, pinning the
goroutine for the run; supported non-Linux Unix systems use process CPU time,
and Windows uses CRT-style elapsed time. ABC's BSAT solve, incremental
containment check, and BSAT proof replay receive the same absolute deadline
and clock domain. As in C, the optional MiniSat solve and its proof replay,
DAR rewriting, initial-state check, and subsequent BMC witness search do not
receive that deadline. A proof can therefore complete after the nominal limit
before the driver's next timeout check.

The INT witness path now calls the Go `Saig_ManBmcScalable` entry point, as
`intCore.c` calls its C counterpart. It uses BMC defaults with `nStart` set to
the current interpolation frame, `nConfLimit` set to 100,000,000, and
verbosity copied from INT; the INT result and frame are retained even if
witness reconstruction returns no model, as in C. For constrained graphs the
C caller passes the original graph unchanged. Scalable BMC treats each
ordinary output, including trailing constraint outputs, as an independent
target and does not enforce the interpolation constraint polarity. The Go INT
path follows that behavior. Constraint-aware trace replay is available through
the explicitly Go-only `Inter_ManGetCounterExampleConstrained` helper; it is
not used by `Inter_ManPerformInterpolation`.

The Go BMC implementation now uses a dynamic frame/object literal map and
source-mapped ≤4-leaf cut CNF with ternary pruning, clause hash reuse, and one
persistent SAT solver. It retains C variable/literal numbering and clause
ordering; in solve-all mode it pre-encodes eligible outputs in PO order before
the solve loop, as C does. It also follows the source completeness cutoff, jump
and conflict-limit transition, global/gap/per-output timeout state, solve-all
dropouts, bridge output, same-model output propagation, and CEX storage/callback
ownership. Differential tests compare cut variable/clause counts and witness
semantics against compiled C, including tight-conflict multi-output cases.
BSAT, Satoko, and Glucose selection follows C precedence. Remaining differences
are the solver backend algorithms themselves and Go wall-clock timeout
accounting where ABC uses CPU ticks. The source-declared `fDropSatOuts` and
`nPisAbstract` fields have no active use in this C routine.
`TestCounterexampleFromUnrolling` and
`TestScalableBmcWitnessStartsAtRequestedFrameAndMapsPis` cover start-frame
behavior, earliest counterexample frame, PI layout, and the finite
one-register state-space cutoff. `TestScalableBmcSourceContractTreatsConstraintOutputAsTarget`
pins the C output-target behavior. `TestCounterexampleUnrollingBoundAndConstrainedTrace`
covers the separate constraint-aware Go extension. Tests cover the BSAT clock
domain, MiniSat's proof-before-timeout boundary, an exhausted MiniSat conflict
budget, and backend cleanup. Exhaustive tests are limited to small formulas
and state spaces; they do not establish performance or completion on large
instances. Several reported timing counters still use Go elapsed time rather
than the C CPU clock.
