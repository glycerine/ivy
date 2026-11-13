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
also uses the source-style sweep. It rejects unsupported native XOR/MUX/buffer
nodes and records original-ID proof pairs for the classic SAT backend, including
deeper speculation, without adding independent verification absent from C.
The circuit-SAT backend does not publish original-ID pairs, matching C. Passing Boolean outcome tests alone does not establish full
equivalence of these algorithms.

## Upstream Cec4 choice limitation

The choice-building path `gia.Cec4_ManSatSolverChoices` is unfinished in
`src/aig/gia/giaEquiv.c`: its recursive AND builder unconditionally returns
literal zero. The Go source-named entry points preserve this stub and the rest
of the source construction sequence. For an independent class whose output
driver remains uninitialized, the original C builder aborts; Go also fails at
that point instead of constructing a substitute graph. A compiled-C subprocess
regression covers both the zero-returning stub and the failing builder path.
When no independent classes exist, both paths duplicate the swept graph.
There are no in-tree Go production callers of this choice builder.

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
The nontrivial source path explicitly rejects native GIA XOR, MUX and buffer
nodes that the current Go sweep cannot handle source faithfully.

`Cec_ManSatSweeping` uses the original speculative sweep and records the
original-ID pairs when requested, without independently rechecking them. The public sweep rejects those unsupported native nodes and retains source IDs
for original AIG nodes. As in the C
pipeline, the sweep time limit is checked after each completed reduction, and
individual simulation or SAT calls can run beyond it, as in C. Linux uses the
source thread-CPU clock; timed sweeps on other platforms report unsupported
until their source CPU clock is implemented. The FRA fallback has its own
stage-boundary limit. Tests compare selected
cases with the original ABC executable and force the unresolved sweep branch
through the genuine FRA fallback. They do not establish equivalence for every
network or exercise every internal FRAIG path.

## Choice computation

The vector choice driver now follows `cecChoice.c`: simulation, ring or direct
speculative queries, the selected SAT or circuit-SAT backend, counterexample
resimulation, and equivalence-to-choice reconstruction. These source-named paths reject unsupported native gates; no extra independent
class verification is added. `Cec_ComputeChoices`
uses the vector path and the source's circuit-SAT conflict cap.

`Cec_ComputeChoicesNew` uses the separate Cec4 simulation and sweep path,
then constructs choices from the retained classes. The
internal `Cec_ManChoiceComputation_int` entry point requires AND-inverter
inputs; the public vector driver rejects unsupported native gates.

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

On Linux CEC2, CEC3 and CEC4 sweep diagnostics use thread CPU time like ABC,
and each sweep and timed helper is pinned to its OS thread. Callers using the
separate manager creation and destruction APIs must keep them on one OS thread
for comparable totals. Other platforms retain wall-clock diagnostics. CEC4
preserves the source assertion when a reduced ID is outside the original
graph's failure bitmap.
The sparse CEC4 justification and the in-range UNKNOWN candidate-ID behavior
have focused comparisons against compiled C; these are not a full mode audit.

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

The public CEC sweep, choice, and nontrivial verification paths reject native
XOR, MUX and buffer nodes rather than silently lowering them. ABC's original
`Gia_ManDup` preserves buffers but copies native XOR and MUX objects as ANDs,
losing their native semantics; a compiled-C regression characterizes that
behavior. A buffer is also rejected where the current Go sweep requires a
strict AND-inverter graph. Phase metadata and the AIG wrapper's GIA cleanup
and witness ownership now follow their C entry points.

The source-named `Cec_ManSatCreate` and node-query wrappers use a separate
BSAT backend with the source frontier allocation and supergate/MUX clauses.
A small compiled-C differential checks status, variable IDs and witness; it
does not establish exact resource-limit, statistics or witness behavior for
every network. The `core` package's Satoko manager remains a separate Go
convenience adapter.

## Split-solver source limits

The split port follows the `ABC_USE_PTHREADS` source path; ABC builds without
that define expose only a `Cec_GiaSplitTest` stub returning -1. The C serial
path can return UNSAT after finding SAT when its pending stack is empty, and
the parallel path records a negative cofactor literal for both branches,
which can produce a false witness. An instrumented compiled-C fixture forces
SAT outcomes while retaining the source scheduling code and compares both
behaviors to Go. The C multi-output driver also reads a temporary manager after
freeing it; Go safely takes ownership of its witness before stopping it, while
retaining the defined output traversal and result behavior. Native XOR, MUX,
and buffer graphs are rejected by the split entry points.

`Cec_SequentialSynthesisPart` requires an attached clock-domain vector. The
original C dereferences a nil vector; the Go entry point rejects that invalid
input with a panic rather than reporting successful synthesis. A compiled-C
subprocess regression characterizes the source behavior.

The source circuit-SAT output solver checks that its entire counterexample
store was consumed even when `fCheckMiter` stops after the first SAT output.
With multiple SAT outputs, the original asserts on trailing records; the Go
entry point preserves that assertion. A compiled-C subprocess test covers it.
