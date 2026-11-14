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

`cec2_simulation_api_test.go` now directly covers the source CEC2 helper paths
for register-output signature copies, random CI initialization with pattern
zero reserved, phase-normalized hash keys, and simulation-class splitting.
These are focused source-derived checks; the CEC2 simulation remains subject
to the zero-primary-input limitation above and does not establish coverage of
every sequential sweep branch.

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
source thread-CPU clock. On other targets, cgo calls the same C runtime
`clock()` used by ABC: process CPU time on POSIX and CRT elapsed time on
Windows. A non-Linux build without cgo reports timed sweeps as unsupported. The FRA fallback has its own
stage-boundary limit. Tests compare selected
cases with the original ABC executable and force the unresolved sweep branch
through the genuine FRA fallback. They do not establish equivalence for every
network or exercise every internal FRAIG path.
An end-to-end, preloaded dual-output failure also compares the source's
failing-output diagnostic, output index, and witness presence with compiled C;
the diagnostic is suppressed when the caller requests silent operation.

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
for comparable totals. On non-Linux targets with cgo, diagnostics call the
platform C runtime `clock()`, which measures process CPU time on POSIX and CRT
elapsed time on Windows; matching ABC requires using the same C runtime. Without
cgo, non-Linux CEC2/3/4 timing calls explicitly panic rather than silently
substitute wall time. CEC4
preserves the source assertion when a reduced ID is outside the original
graph's failure bitmap.
The sparse CEC4 justification and the in-range UNKNOWN candidate-ID behavior
have focused comparisons against compiled C; these are not a full mode audit.
The source pattern-packing try helper asserts that the current pattern cursor
is in range, regardless of the separately supplied bit index. The Go helper
preserves that precondition, with an executable C comparison.

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

`Gia_ManRegCreatePart` copies native GIA XOR objects using `Gia_ManAppendAnd`,
so the partition computes AND where the source graph computes XOR. This is an
inherited C behavior, confirmed against compiled ABC; it can change partition
semantics and therefore must not be treated as a faithful XOR reduction.

The CEC2/CEC3 sweep engine matrix checks random combinational graphs with cone
mode both off and on, equivalence merging, miter witnesses, and graph finishing
enabled or disabled. The sequential partition tests cover cleanup, SCORR and
LCORR selection; empty and self-only clock domains; traversal epochs; and
multiple non-contiguous selected registers supplied out of index order. That
partition case checks PI/other-register counts, selected-state ordering, real
output order, and the map-back IDs. Extracted submanagers are released on both
the synthesized and no-PI paths, matching the C loop's cleanup. These checks
are concrete branch coverage, not a complete parameter cross-product.

Remaining sweep gaps include source comparisons for verbose/statistical
reporting, unusual conflict-limit and iteration-boundary combinations, and
exhaustive random seeds/options. Partition gaps include full end-to-end
class-transfer behavior across interacting clock domains, malformed partition
indices (which C treats as assertion failures), and exhaustive combinations of
SCORR/LCORR options and limits. Native XOR partitioning remains the known
source-inherited semantic defect described above.

The zero-round/zero-iteration CEC2/CEC3 case now checks the returned graph on
every input assignment. CEC4 recycle tests exercise the SAT-variable threshold
off/on boundary and the strict call-count comparison while confirming UNSAT
proof queries ran. The CEC4 entry test also checks that an early miter witness
leaves the caller's preexisting output pointer untouched. Overlapping clock
domains cover the A~B then B~C class chain and verify normalization to A. A
compiled-ABC oracle compares that chain and a complemented-phase chain,
confirming that opposite-phase registers are not merged through the shared
register. It also checks constant-valued next-state functions on the two
overlapping endpoints; the middle input-driven register keeps both partitions
on the synthesis path, and C plus Go map the endpoint registers to constant 0.
CEC2's private reduced manager is stopped on every return path, matching
`Cec2_ManDestroy` in C.

The source-backed CEC4 recycle fixture sets `nSatVarMax=1` and
`nCallsRecycle=0`; C and Go both make 56 queries and recycle 55 times. At a
two-conflict limit, both report 45 SAT, one UNSAT, and 10 UNKNOWN results. At
one conflict, C reports 42 SAT, one UNSAT, and 13 UNKNOWN while Go reports 42
SAT and 14 UNKNOWN. The single boundary result is solver-dependent under the
tight conflict budget; the UNKNOWN result is conservative, and no equivalence
is claimed for that candidate. The variable and call threshold boundaries are
also covered separately. CEC2/CEC3 detailed conflict-limit status combinations
and source diagnostic output remain selective. CEC2/3/4 sweep code has no
time-limit branch; timeout parity applies to the separate classic CEC sweep
path documented above.

The source circuit-SAT output solver checks that its entire counterexample
store was consumed even when `fCheckMiter` stops after the first SAT output.
With multiple SAT outputs, the original asserts on trailing records; the Go
entry point preserves that assertion. A compiled-C subprocess test covers it.
