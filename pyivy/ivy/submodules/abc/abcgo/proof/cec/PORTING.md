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
independently checks original-ID proof pairs, including for deeper speculation
and circuit SAT. Passing Boolean outcome tests alone does not establish full
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
logic, retains source IDs for original nodes, remaps existing equivalence
classes, and independently validates alias proofs. As in the C
pipeline, the overall time limit is checked between stages and iterations; an
individual SAT or FRAIG call can run beyond that limit. Tests compare selected
cases with the original ABC executable and force the unresolved sweep branch
through the genuine FRA fallback. They do not establish equivalence for every
network or exercise every internal FRAIG path.

## Choice computation

The vector choice driver now follows `cecChoice.c`: simulation, ring or direct
speculative queries, the selected SAT or circuit-SAT backend, counterexample
resimulation, and equivalence-to-choice reconstruction. Each remaining class
is independently checked against the original graph before reconstruction;
unknown or false pairs are removed. Native gates are lowered for this path and
validated classes are mapped back to the input graph. `Cec_ComputeChoices`
uses the vector path and the source's circuit-SAT conflict cap.

`Cec_ComputeChoicesNew` uses the separate Cec4 simulation and sweep path,
then independently checks retained classes before constructing choices. The
internal `Cec_ManChoiceComputation_int` entry point requires AND-inverter
inputs; the public vector driver lowers native gates first.

## Remaining fidelity and coverage limits

The naive verification route uses the Go `core.VerifyPairs` cone-CNF/Satoko
adapter, rather than ABC's mapped MF CNF and bsat loop; its conflict behavior,
logging and per-output progress are therefore not exact C reproductions.
Choice candidate validation also deliberately adds independent SAT checks that
ABC does not perform. The `ChoiceSnapshots` and `SatSweep` convenience APIs
remain separate Go algorithms; the source-named vector choice and SAT-sweep
entry points use the source-based paths described above. `Cec_ManChoiceComputation`
uses the DAR/DCH engine; the current DCH port documents that its `UseGia` and
`UseCSat` configuration fields do not switch the old DCH backend.

Original-ABC comparisons cover selected small networks and options, not a broad
real-world corpus or every resource-exhaustion path. In particular, the source
loop's exact verbose timing/statistics, choice-refinement overflow behavior,
and behavior under timeouts and sparse initial classes need further comparison.
Independent original-graph replay and proof checks reduce the risk of an
incorrect published witness or choice but do not establish complete coverage.

`Cec_ManFraSpecReduction` preserves the source level-limit behavior that
indexes the original graph using the representative's reduced-copy ID. If
structural hashing changes IDs, this can suppress an otherwise eligible
query; a regression records this inherited C behavior.
