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

The top-level `Cec_ManVerify` now uses the source-style special cases,
normalization, iterative sweep, and FRAIG fallback; `Cec_ManVerifyOld` calls
the FRAIG fallback as well. The public `Cec_ManSatSweeping` entry point now
also uses the source-style sweep. Its direct interface currently rejects native
XOR/MUX/buffer nodes and rejects original-ID tracking with deeper speculation
or circuit SAT; those combinations need further work. Passing Boolean outcome
tests alone does not establish equivalence of these algorithms.

## Upstream Cec4 choice limitation

The choice-building path `gia.Cec4_ManSatSolverChoices` is not a working
implementation when independent choice classes are present: its recursive AND
builder returns constant zero. This is inherited from the original C routine
`src/aig/gia/giaEquiv.c` (`Cec4_ManSatSolverAnd_rec`), which also immediately
returns zero. The Go entry point now explicitly panics when independent choices
need reconstruction instead of returning a potentially corrupt graph; it still
duplicates the swept graph when no independent choices exist. The recursive AND
entry point also panics. There are no in-tree Go production callers of this
choice builder. ABC's separate `Gia_ManEquivToChoices` workflow is ported in
`aig/gia/equiv_choices.go`, but it operates on one graph's representative
classes, whereas this unfinished entry point accepts original and swept graphs
and produces sibling choices. It is not a drop-in implementation. A correct
adapter would require a source-backed contract and independent equivalence
tests. A build or a passing test of other Cec4 paths does not establish
correctness of this path.

## Sequential and miter review

Sequential resimulation tests cover frame-major counterexample packing, register
state initialization, and register propagation across frames. A one-register
circuit with no primary inputs previously panicked while inferring the packed
word count; the Go implementation now reads the first entry's length (or uses
one word for an empty vector). With no primary inputs the layout cannot encode
a frame count, so this API performs one frame. The original C path divides by the primary-input
count and assumes at least two vector entries, so this is an intentional Go
extension for a valid input rather than a C differential match.

Miter output detection now checks actual one bits, or actual output differences
for dual-output miters. Phase-normalizing these results could hide a constant
one or a pair of permanently opposite outputs. The C implementation performs
that phase-normalized comparison; the Go correction is intentional and covered
by regressions. Signature comparison for equivalence-class refinement remains
phase-normalized and continues to have separate C-oracle coverage.

An end-to-end classic SAT sweep test also verifies transfer and replay of
CO-indexed counterexamples and proved-output sentinels through graph cleanup,
including a register-input CO. These checks supplement the full package suite;
they do not establish complete coverage of all sequential algorithms.

## Classic combinational verification pipeline

`Cec_ManVerify` now performs the source's trivial-case checks, output-pair
normalization, iterative simulation/speculative SAT sweeping, and the original
FRAIG fallback. The explicit naive option takes the direct SAT path. Sweep and
FRA witnesses are replayed against the original paired outputs before they are
returned. Native GIA XOR, MUX and buffer nodes are lowered through the existing
AIG conversion before the AND-inverter sweep.

`Cec_ManSatSweeping` uses the original speculative sweep and records the SAT-
proved original-ID pairs when requested. Speculative pairs are independently
checked against the original graph, so a failed earlier assumption cannot turn
into a published proof. The ordinary and circuit SAT backends both collect
candidate pairs for that independent validation.
The public sweep lowers native XOR, MUX, and buffer nodes to AND-inverter
logic, retains source IDs for original nodes, and independently validates alias
proofs. Native graphs with existing equivalence classes are rejected until
class remapping is supported. As in the C
pipeline, the overall time limit is checked between stages and iterations; an
individual SAT or FRAIG call can run beyond that limit. Tests compare selected
cases with the original ABC executable and force the unresolved sweep branch
through the genuine FRA fallback. They do not establish equivalence for every
network or exercise every internal FRAIG path.
