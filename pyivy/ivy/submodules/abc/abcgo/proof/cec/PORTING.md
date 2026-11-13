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
also uses the source-style sweep. It lowers native XOR/MUX/buffer nodes and
records original-ID proof pairs, including for deeper speculation and circuit SAT,
without adding independent verification absent from the C implementation. Passing Boolean outcome tests alone does not establish full
equivalence of these algorithms.

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
state initialization, and register propagation across frames. The source
requires at least one primary input and a simulation vector with two entries;
its zero-input path divides by zero and cannot infer a packed word size.

The source miter output analysis compares signatures modulo complement. This
can suppress reports for a constant-one output or permanently opposite paired
outputs. The Go port preserves that source behavior, with a regression test;
the signature comparison also has separate C-oracle coverage.
Cec2, Cec3 and Cec4 direct output-simulation routines likewise preserve the
source phase-normalized comparison. Their top-level sweepers perform the
source all-zero-input phase check before entering simulation.

An end-to-end classic SAT sweep test verifies transfer and replay of
CO-indexed counterexamples and proved-output sentinels through graph cleanup,
including a register-input CO. These checks supplement the full package suite;
they do not establish complete coverage of all sequential algorithms.

## Classic combinational verification pipeline

`Cec_ManVerify` now performs the source's trivial-case checks, output-pair
normalization, iterative simulation/speculative SAT sweeping, and the original
FRAIG fallback. The explicit naive option generates MF CNF and uses the BSAT implication loop,
matching `cecCec.c`. The naive routine does not store a counterexample. The
classic pipeline follows the source witness transfer and diagnostic behavior.
Native GIA XOR, MUX and buffer nodes are lowered through the existing AIG
conversion before the AND-inverter sweep.

`Cec_ManSatSweeping` uses the original speculative sweep and records the
original-ID pairs when requested, without independently rechecking them. The
public sweep lowers native XOR, MUX, and buffer nodes to AND-inverter logic,
retains source IDs for original nodes, and remaps existing equivalence classes. As in the C
pipeline, the overall time limit is checked between stages and iterations; an
individual SAT or FRAIG call can run beyond that limit. Tests compare selected
cases with the original ABC executable and force the unresolved sweep branch
through the genuine FRA fallback. They do not establish equivalence for every
network or exercise every internal FRAIG path.

## Choice computation

The vector choice driver now follows `cecChoice.c`: simulation, ring or direct
speculative queries, the selected SAT or circuit-SAT backend, counterexample
resimulation, and equivalence-to-choice reconstruction. Native gates are lowered for this path and classes are mapped back to the input
graph; no extra independent class verification is added. `Cec_ComputeChoices`
uses the vector path and the source's circuit-SAT conflict cap.

`Cec_ComputeChoicesNew` uses the separate Cec4 simulation and sweep path,
then constructs choices from the retained classes. The
internal `Cec_ManChoiceComputation_int` entry point requires AND-inverter
inputs; the public vector driver lowers native gates first.

## Remaining fidelity and coverage limits

The `ChoiceSnapshots` and `SatSweep` convenience APIs
remain separate Go algorithms; the source-named vector choice and SAT-sweep
entry points use the source-based paths described above. `Cec_ManChoiceComputation`
uses the DAR/DCH engine; the current DCH port documents that its `UseGia` and
`UseCSat` configuration fields do not switch the old DCH backend.

Original-ABC comparisons cover selected small networks and options, not a broad
real-world corpus or every resource-exhaustion path. In particular, the source
loop's exact verbose timing/statistics, choice-refinement overflow behavior,
and behavior under timeouts and sparse initial classes need further comparison.
Passing the tests does not establish complete coverage of these source paths.

`Cec_ManFraSpecReduction` preserves the source level-limit behavior that
indexes the original graph using the representative's reduced-copy ID. If
structural hashing changes IDs, this can suppress an otherwise eligible
query; a regression records this inherited C behavior.

## Inherited speculative proof defect

ABC's original speculative sweep can publish false original-ID pairs. On a
13-input AND-chain fixture it reports pairs (22,26), (23,26), (24,26), and
(25,26), although objects 22 and 26 differ for input assignment 1023. Using
these pairs for original-ID reduction changes the resulting circuit. A Go
regression characterizes this behavior and an optional compiled ABC oracle
checks the reported pairs when ABC object files are available. The source-named
port preserves this upstream behavior; callers must not treat these pairs as
independently verified proofs.

ABC's naive MF-CNF path also assumes physical CO order agrees with its CO
vector. With physical outputs `[0, 0, x, x]` and logical CO entries 1 and 2
swapped, the original reports both unequal logical pairs equivalent. The Go
adapter preserves this source behavior, covered by a reordered-output test and
the optional compiled ABC oracle. Reordered CO vectors are therefore unsafe for
this verification path.
