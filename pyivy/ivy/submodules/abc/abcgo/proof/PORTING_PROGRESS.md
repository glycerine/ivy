# Proof package porting progress

The target is a faithful Go port of each compiled C proof path, with source
reviews and tests that cover the algorithms and option branches. Do not treat
Go-only fallback algorithms as completed ports. When the checked-in C path has
a defect or is incomplete, preserve that fact in the porting notes and tests
rather than silently substituting another algorithm.

## Current checkpoint (2026-11-13)

Proof-package work remains under source-by-source audit. The current
uncached race gate passes for all packages in `./proof/...`; vet also passes.
The shared dependency gate passes for `./aig/...`, `./opt/dar/...`, and
`./sat/cnf`. These gates do not establish complete source parity: INT2 has
incomplete C routines, CEC's broader sweep/partition review is unfinished, and
SAIG retains the documented backward-driver caveat below. Continue the audit.

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
* `acec`: source flow and preconditions are restored. Compiled-ABC differential
  tests cover 76 circuit pairs, and checked-in test vectors exercise Booth
  normalization and tied XOR/MAJ cuts. Negative cut limits and single-output
  box production now match C. The C implementation has reproducible
  false-equivalence cases and a two-output assumption, documented and tested
  as source behavior. Package race tests and vet passed.
* `cec`: the naive verifier now uses MF CNF/BSAT, with the C output-variable
  mapping. Compiled ABC comparisons confirm source-inherited false speculative
  equivalence pairs and a reordered-output misclassification; tests preserve
  and document those defects. Native XOR/MUX wrapper paths now follow C's
  duplication behavior or reject unsupported input at the source boundary.
  Isomorphism preconditions and constant-miter handling match C. Sweep early
  exits preserve `iOutFail`, the circuit-SAT path does not publish Go-only
  original-ID pairs, and timeout checks occur after a completed reduction
  iteration. Linux timed sweeps use the C thread-CPU clock; timed sweeps on
  other platforms report unsupported. Package race tests and vet passed.
  Broader sweep/partition review remains partial. Cec4's upstream choice
  builder remains incomplete.
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
  random completion now match C. Native XOR/buffer and non-topological inputs
  outside the C source contract are rejected.
* `fra`: SEC, clause and BMC paths follow the active source's timeout, budget,
  solver-reuse behavior, and native-EXOR-as-AND handling where C hardcodes
  `Aig_And`. Source-derived partition, implication, one-hot, and induction
  tests passed; broader behavioral audit remains in progress.
* `fracomb`: AIG deadlines now propagate through DAR rewrite stages as in C.
  Its SAT path uses the source-mapped CNF writer and source Solver2
  simplification. A multi-sweep SAT fallback test replays witnesses on original
  and rewritten graphs. C's zero-model shortcut, raw flipped SAT model, and
  lack of final witness revalidation are now preserved and characterized.
* `fraig`: source behavior has been restored for miter checks, external clauses,
  feedback compression, pointer-based hashing/order, and MSAT behavior. Several
  upstream limitations, including witnesses that violate external clauses,
  are covered by source-characterization tests. MSAT review continues.
* `framgr`: duplicate BMC helpers were consolidated only where behavior was
  shown identical. Simulation, BMC callback reset, native-EXOR-as-AND behavior,
  and output-ID witness layout follow C; broader manager audit remains
  incomplete.
* `int`: MiniSat 1.14p has exhaustive Craig-interpolant checks and option/error
  coverage. Proof-root validation now builds its annotation in one ordered
  pass, preserving clause order. Race tests and vet passed.
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
  whole path as unsupported for source fidelity.
* `live`: Go-only reachability and fairness-parser substitutes were removed.
  Graph operations now use explicit AIG arguments as C does; network metadata
  remains sourced from the supplied network. Safe graph-copy and safety-model
  simplifications preserve source order. Package race tests and vet passed.
* `pdr`: source solver context, limits, helper edge cases and graph lifecycle
  are covered. Differential tests exercise 64 machines and source-matched
  option combinations, plus invariant-file output. Package race tests and vet
  passed; some cases expose inherited invalid-counterexample behavior in C.
  Incremental-restart handling preserves source edge cases, including an
  assertion on a first restart with no saved clauses and an interrupted push
  that can be treated as proof. Further resource-limit coverage is needed.
* `ssc`: evolving FRAIG substitutions, BSAT frontier encoding, source-style
  class refinement, native-input lowering, random stream, equivalence
  reduction, and CEC verification have source tests. Verbose mode performs the
  source's five-word care estimate and advances the shared RNG. Care-pattern
  transfer uses the source helpers, with tests for simulation, classes, and
  later RNG state. Complete option parity remains under review.
* `ssw`: source rarity, SAT polarity, prefix, class-count, manager-lifecycle,
  reduction accounting, stop behavior, and timeout behavior is restored and
  characterized. A reachable invalid rarity witness is documented as an
  upstream C defect. Source CioId setup is restored in manager, BMC,
  latch-sweep, and dynamic-sweep paths; current race tests and vet passed.

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
backward retiming driver's historical constraint-handling branches and the
initialized-state `Aig_ManFrames` path remain documented fidelity limits.

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
