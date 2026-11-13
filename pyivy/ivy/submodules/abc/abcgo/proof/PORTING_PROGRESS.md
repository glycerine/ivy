# Proof package porting progress

The target is a faithful Go port of each compiled C proof path, with source
reviews and tests that cover the algorithms and option branches. Do not treat
Go-only fallback algorithms as completed ports. When the checked-in C path has
a defect or is incomplete, preserve that fact in the porting notes and tests
rather than silently substituting another algorithm.

## Current checkpoint (2026-11-13)

Proof-package work remains under source-by-source audit. The most recent
uncached `GOCACHE=/tmp/abcgo-final-proof-cache go test -race -count=1 -p 8 ./proof/...`
gate passes for all proof packages, and `go vet ./proof/...` passes. The
combined proof/GIA/BSAT race gate and `go vet ./proof/... ./aig/gia ./sat/bsat`
also pass.
The shared dependency gate passes for `./aig/...`, `./opt/dar/...`, and
`./sat/cnf`. These gates do not establish complete source parity: INT2 has
incomplete C routines, CEC's broader sweep/partition review remains partial,
and SAIG retains the documented backward-driver caveat below. Continue the
audit.

The uncached `go test -cover -count=1 -p 8 ./proof/...` snapshot below measures
executed Go statements, not C fidelity or algorithm completeness. Use it to
find untested behavior, not as a completion threshold.

| Package | Statements |
| --- | ---: |
| `abs` | 76.6% |
| `acec` | 76.6% |
| `cec` | 80.3% |
| `cec/core` | 93.0% |
| `dch` | 91.2% |
| `fra` | 74.8% |
| `fracomb` | 92.4% |
| `fraig` | 77.6% |
| `framgr` | 85.6% |
| `int` | 87.2% |
| `int2` | 90.8% |
| `live` | 87.0% |
| `pdr` | 77.4% |
| `ssc` | 88.8% |
| `ssw` | 79.8% |

## Source API inventory

Every checked-in `src/proof` package directory has a Go counterpart. A
tree-wide scan of exported C declarations and definitions found no confirmed
missing active external API. This is an inventory result, not proof that all
implementations or option branches are source-equivalent. C macro iterations
do not require Go functions, and some C internal helpers are consolidated into
Go methods. DCH and SSC class/SAT helpers and CEC solver helpers use
package-internal implementations; INT's optional MiniSat 1.14p interpolation
helpers share a Go implementation. The ABS UNSAT-core refinement helpers in
`absRef.c` are under `#if 0`.

Some headers declare symbols that have no implementation in this C checkout,
including `Fraig_DfsReverse`, `Fraig_CountLevels`, `Ssc_GenerateOneHot`, and
`ZPdr_SetIsInit`. These are upstream declaration-only entries, not completed Go
implementations. INT2's checked-in interpolation driver and suffix/preimage
helpers are incomplete, as detailed below; ABS also has the source-fidelity
and representation limits described in its package notes.

## Package status

* `abs`: source tests cover C's VTA object bound, cancellation result behavior,
  and non-topological flop-ranking order. Native MUX controls are included by
  Go graph helpers where those helpers accept native GIA nodes. The C
  cancellation behavior can expose a stale proof, and is documented. Full
  race tests and vet passed. Standalone VTA rejects native XOR/MUX graphs;
  callers must lower them first. The old GLA `fUseFullProof` option does not
  select a distinct proof-logging mode in Go, so resource behavior can differ.
  Additional flop-transform tests cover word boundaries, non-Boolean initial
  values, duplicate selections and unchanged next-state functions. Bounded
  frame-limit, 257-label UNSAT-core and compiled-C conflict-budget exhaustion
  tests were added; larger sequential resource-limit interactions and
  memory-scale stress remain uncovered. Linux timeout handling now uses the
  source thread-CPU clock and pins the solving goroutine; non-Linux still uses
  wall time where C uses process CPU time. A broader audit remains underway.
* `acec`: source flow and preconditions are restored. Compiled-ABC differential
  tests cover 76 circuit pairs, and checked-in test vectors exercise Booth
  normalization and tied XOR/MAJ cuts. Negative cut limits and single-output
  box production now match C. The C implementation has reproducible
  false-equivalence cases and a two-output assumption, documented and tested
  as source behavior. Package race tests and vet passed.
  Added compiled-C recognition comparisons across 96 Booth pattern cases and a
  source test inventory; end-to-end complex multiplier, timing-box and larger
  graph differential coverage remain open.
* `cec`: the naive verifier now uses MF CNF/BSAT, with the C output-variable
  mapping. Compiled ABC comparisons confirm source-inherited false speculative
  equivalence pairs and a reordered-output misclassification; tests preserve
  and document those defects. Native XOR/MUX wrapper paths now follow C's
  duplication behavior or reject unsupported input at the source boundary.
  Isomorphism preconditions and constant-miter handling match C. Sweep early
  exits preserve `iOutFail`, the circuit-SAT path does not publish Go-only
  original-ID pairs, and timeout checks occur after a completed reduction
  iteration. Partition traversal now advances the source epoch twice. Linux
  timed sweeps and internal CEC2/3/4 timers use the C thread-CPU clock;
  non-Linux timing retains a wall-clock fallback. Separately managed timed
  calls require creation and destruction on the same OS thread. New compiled-C
  tests cover CEC2/3 SAT and UNSAT witnesses, CEC4 sparse replay, conflict
  limits, UNKNOWN candidate IDs, interleaved inputs, and timer behavior. The
  CEC4 out-of-range failure-bitmap behavior still differs from C; the reviewed
  source matrix remains selective. Package race tests and vet passed.
  The sequential synthesis entry also preserves C's failure boundary for a
  missing clock-domain vector. Tests cover C's incomplete Cec4 builder return
  and partial construction failure instead of substituting a working builder.
  Broader sweep/partition review remains partial.
  The lower `opt/dar` facade wires its source default through the CEC adapter,
  and the GIA default uses the GIA miter/skip-three/DCH path. CEC/Cec4 modes are
  exposed through a higher-layer wrapper to avoid an import cycle.
* `cec/core`: this is a separate Satoko-based Go convenience adapter, not the
  source CEC SAT backend. Source-named node checks and SAT-value accessors now
  use the separate BSAT backend in `cec`. Mixed-gate truth-table, witness,
  recycling and source-wrapper tests have passed under race and vet.
* `dch`: uses the ported BSAT path, source SAT-variable numbering and allocation,
  supergate/MUX clauses, incremental working FRAIG, and source simulation
  behavior. Non-TFO/TFO refinement scope, class counters, SAT-cone values, and
  random completion now match C. Shared RNG side effects and allocated-choice
  representative preconditions also match C. Tests cover support and metadata
  boundaries; recent statement coverage was 88.2%. Native XOR/buffer and
  non-topological inputs outside the C source contract are rejected.
* `fra`: SEC, clause and BMC paths follow the active source's timeout, budget,
  solver-reuse behavior, and native-EXOR-as-AND handling where C hardcodes
  `Aig_And`. Source-derived partition, implication, one-hot, and induction
  tests passed. SEC terminal counts and stored CioIds, SEC setup, and the
  induction no-progress branch now follow C. Source-instrumented tests cover
  its no-progress warning/result and the empty-output partition failure
  boundary. SEC status and witness transfer now match C, including its
  unvalidated witness behavior; callers needing replay use the separate
  `Fra_FraigSecVerified` API. Compiled-C SEC comparisons cover separate-output
  interpolation. Broader behavioral audit remains in progress.
* `fracomb`: AIG deadlines now propagate through DAR rewrite stages as in C.
  Its SAT path uses the source-mapped CNF writer and source Solver2
  simplification. A multi-sweep SAT fallback test replays witnesses on original
  and rewritten graphs. C's zero-model shortcut, raw flipped SAT model, and
  lack of final witness revalidation are now preserved and characterized.
* `fraig`: source behavior has been restored for miter checks, external clauses,
  feedback compression, pointer-based hashing/order, and MSAT behavior. Several
  upstream limitations, including witnesses that violate external clauses,
  are covered by source-characterization tests. Node-vector shrink, clear, pop,
  and remove preserve C's backing-array entries observable through
  `Fraig_NodeVecReadArray`. Opt-in compiled-C tests compare allocator behavior
  and compression over covering pairs, overlapping patterns, and partial words.
  Pool-accounting and real-workload differential gaps remain documented.
* `framgr`: duplicate BMC helpers were consolidated only where behavior was
  shown identical. Simulation, BMC callback reset, native-EXOR-as-AND behavior,
  output-ID witness layout, SAT root simplification, constant-implication
  behavior, and the uncomplemented-parent assertion follow C. Tests now cover
  class-copy/latch-correlation state, BMC adapters, sequential simulation,
  counterexample/model extraction and one-hot reporting/RNG behavior. The
  previous coverage snapshot was 80.6%; a new measurement is pending. Remaining
  branches include some BMC/FRAIG sweep and SAT model paths; this is not a proof
  of complete algorithm coverage.
* `int`: MiniSat 1.14p has exhaustive Craig-interpolant checks and option/error
  coverage. Proof-root validation now builds its annotation in one ordered
  pass, preserving clause order. Partition-assembly tests now cover restoration
  of variable/literal state on success and early failures, and C's stop behavior
  after rejecting a B clause. Race tests and vet passed.
* `int2`: the checked-in C interpolation driver is uncompilably incomplete, so
  the Go driver returns undecided and does not invent a bounded reachability
  fallback. Helper tests characterize source behavior. The suffix constructor
  remains unsupported because C references undeclared `k`/`vSop` and reads CNF
  from a different graph than the one into which it derives clauses. The
  preimage helper also returns undecided because its C code refers to an
  undeclared manager and does not establish a reliable mapping contract.
  Refinement compares GIA priorities as unsigned values, matching C's `uint32`
  field. The incremental frame builder reconstructs missing C operations and has no
  unsupported status; it is not source-confirmed and callers must treat the
  whole path as unsupported for source fidelity. `TEST_COVERAGE.md` maps the
  implemented helpers to tests and lists timeout/CNF-error gaps.
* `live`: Go-only reachability and fairness-parser substitutes were removed.
  Graph operations now use explicit AIG arguments as C does; network metadata
  remains sourced from the supplied network. Safe graph-copy and safety-model
  simplifications preserve source order. Source-shaped graph copying maps C
  EXOR nodes through AND as the C constructors do, while native convenience
  APIs retain XOR behavior. Compiled-C tests cover this distinction and the
  liveness, monotone, constrained-cone and disjunctive wrappers. Source builders
  now traverse all nodes in source order and set copy mappings, including for
  unused nodes. Compiled-C tests also cover source parser cursor semantics,
  usage ordering, incomplete-source boundaries and wrapper behavior. Race
  tests and vet pass at 87.0% statement coverage. Compiled-C regressions now
  verify source builders include every PO and retain or clear constraint
  metadata according to the specific C constructor. Native Go convenience
  paths preserve their separate constraint behavior. Command adaptation,
  diagnostics and malformed-input behavior still differ; see `live/PORTING.md`.
* `pdr`: source solver context, limits, helper edge cases and graph lifecycle
  are covered. Differential tests exercise 64 machines and source-matched
  option combinations, plus invariant-file output. Package race tests and vet
  passed; some cases expose inherited invalid-counterexample behavior in C.
  Incremental-restart handling preserves source edge cases, including an
  assertion on a first restart with no saved clauses and an interrupted push
  that can be treated as proof. SAT solvers now restart in place to preserve the
  pointer identity exposed by the C API. Added solver-recycling threshold coverage;
  further timeout/resource-limit combinations need coverage.
* `ssc`: evolving FRAIG substitutions, BSAT frontier encoding, source-style
  class refinement, native-input lowering, random stream, equivalence
  reduction, and CEC verification have source tests. Verbose mode performs the
  source's five-word care estimate and advances the shared RNG. Care-pattern
  transfer uses the source helpers, with tests for simulation, classes, and
  later RNG state. Negative SAT budgets now follow C through pivot and
  assumption checks. Complete option parity remains under review.
* `ssw`: source rarity, SAT polarity, prefix, class-count, manager-lifecycle,
  reduction accounting, stop behavior, and timeout behavior is restored and
  characterized. A reachable invalid rarity witness is documented as an
  upstream C defect. Source CioId setup is restored in manager, BMC,
  latch-sweep, and dynamic-sweep paths. Gap and negative-timeout behavior now
  match C; current race tests and vet passed.

## Shared dependencies in progress

`abcgo/sat/cnf` now uses ABC's compressed MSOP table, DAR cut ordering,
dominance/eviction, area-flow mapping, mapped CNF writer, and variable
numbering. Direct compiled-C comparisons matched exact ordered DIMACS output
for 345 generated asserted, free-output, and sequential cases; exhaustive
three-input truth-function checks also pass. `abcgo/opt/dar` cut enumeration
follows C's slot ordering and eviction behavior. Its former implementation now
lives in `opt/dar/core`, with `opt/dar` as a compatibility facade that avoids
the GIA/CNF/CEC import cycle. The facade's default choice path uses the source
GIA miter and DCH/CEC/Cec4 branches; the lower GIA API exposes CEC/Cec4 through
an explicit higher-layer wrapper. Compiled ABC option-matrix checks pass.

The shared AIG/SAIG migration stores C's true terminal counts and per-object
CioIds instead of deriving them from current terminal slices. Duplication,
retiming, and register hiding/exposure now follow source-specific count and ID
updates; C intentionally leaves stale or zero metadata on some paths. The
backward retiming driver's historical constraint-handling branches remain a
documented fidelity limit. Initialized `Aig_ManFrames` with register outputs
now follows C, including negative true-input counts when no CIs are created.

Timeout accounting remains a cross-package fidelity gap. ABS now uses the C
thread-CPU clock on Linux while pinning the solving goroutine, process CPU via
`getrusage` on supported non-Linux POSIX systems, and elapsed time on Windows
to match the Microsoft CRT; unsupported timed platforms fail explicitly. The
Windows CRT clock's roughly 24.8-day overflow is not emulated. CEC timed
sweeps and internal CEC2/3/4 timers pin their goroutine and use the source
thread-CPU clock on Linux; non-Linux builds retain a wall-time fallback.
Separately managed timed calls need create/destroy operations to stay on the
same OS thread. PDR, INT, SSW, and FRA deadlines and solver contexts still need
a call-site-by-call-site review. Switching the shared helper alone would be
unsafe because goroutines can move between OS threads and SAT/context deadlines
currently use wall time.

FRA/AIG partitioned FRAIG and choice and SAIG timeframe bridges are implemented
through explicit callback APIs, with FRA wrappers supplying the corresponding
source backends. Remaining `base/abci` callers are out-of-scope drafts and
are not evidence of completed application integration. `abcgo/sat/bsat` has a
source-style cumulative resource-limit API. It also preserves an upstream C
assumption-order bug: after certain root conflicts, solver reuse can report
SAT for an inconsistent formula. The behavior is C-oracle tested and callers
are advised to simplify and discard a conflicting solver. `abcgo/sat/msat` now
matches C's chunk pools, clause allocation accounting, learned-clause sorting,
DIMACS/report output, vector ordering and accessors; source-by-source
comparison continues. The original C checkout only declares the unused
variable-heap API, so its Go replacement bodies were removed. The latest
uncached `go test -race -count=1 -p 8 ./proof/...` and `go vet ./proof/...`
both pass. The optional SSC C oracle remains opt-in (`SSC_BUILD_ORACLE=1`);
its fresh-source build passed earlier, but is not run by the default gate.

Run full checks from `abcgo`, excluding experimental directories, for example:

```sh
GOCACHE=/tmp/abcgo-cache go test -race -count=1 -p 8 ./proof/...
GOCACHE=/tmp/abcgo-cache go vet ./proof/...
```
