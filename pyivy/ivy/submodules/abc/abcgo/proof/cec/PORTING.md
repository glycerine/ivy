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

The top-level `Cec_ManVerify` currently uses the direct SAT implementation
instead of the source's default normalization, iterative SAT-sweeping, and
old-engine fallback pipeline. Likewise, the current `Cec_ManSatSweeping`
delegates to a different sweep, and `Cec_ManVerifyOld` delegates back to the
current top-level implementation rather than the original SSW-based fallback.
These are substantive source-fidelity and option-behavior gaps; passing tests
of their Boolean outcomes does not establish equivalence of the algorithms.

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
