# Proof package porting progress

The target is a faithful Go port of each compiled C proof path, with source
reviews and tests that cover the algorithms and option branches. Do not treat
Go-only fallback algorithms as completed ports. When the checked-in C path has
a defect or is incomplete, preserve that fact in the porting notes and tests
rather than silently substituting another algorithm.

## Current checkpoint (2026-11-13)

Proof-package work is under parallel source-by-source audit. Recent green
checkpoints cover CEC, FRA/framgr, ACEC, DCH, ABS, INT2, live, SSW, PDR, BSAT,
MSAT, MiniSat-p, DAR, and SSC. The integrating thread is committing green
package-sized checkpoints on the current branch. A repository-wide race/vet
gate has not yet been rerun after the latest changes; several focused audits
remain active, so do not treat this checkpoint as completion.

## Package status

* `abs`: source boundary tests cover the C VTA assertion and retained witness
  behavior. Native MUX controls are included by Go graph helpers where those
  helpers accept native GIA nodes. Full race tests and vet passed; a package
  audit is continuing.
* `acec`: source flow and preconditions are restored. Compiled-ABC differential
  tests cover 76 circuit pairs, and checked-in test vectors exercise Booth
  normalization and tied XOR/MAJ cuts. The C implementation has reproducible
  false-equivalence cases, documented and tested as source behavior. Package
  race tests and vet passed.
* `cec`: the naive verifier now uses MF CNF/BSAT, with the C output-variable
  mapping. Compiled ABC comparisons confirm source-inherited false speculative
  equivalence pairs and a reordered-output misclassification; tests preserve
  and document those defects. Native XOR/MUX wrapper paths now follow C's
  duplication behavior or reject unsupported input at the source boundary.
  Package race tests and vet passed. Cec4's upstream choice builder remains
  incomplete.
* `cec/core`: source-named node checks and SAT-value accessors were corrected;
  the separate Go equivalence helper remains separate. Mixed-gate truth-table,
  witness, recycling and source-wrapper tests passed under race and vet.
* `dch`: uses the ported BSAT path, source SAT-variable numbering and allocation,
  supergate/MUX clauses, incremental working FRAIG, and source simulation
  behavior. Non-TFO refinement scope and class counters now match C; TFO
  refinement and unconstrained simulation inputs remain under audit. Native
  XOR/buffer and non-topological inputs outside the C source contract are
  rejected.
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
  coverage. Race tests and vet passed.
* `int2`: the checked-in C interpolation driver is uncompilably incomplete, so
  the Go driver returns undecided and does not invent a bounded reachability
  fallback. Helper tests characterize source behavior. The suffix constructor
  remains unsupported because C references undeclared `k`/`vSop` and reads CNF
  from a different graph than the one into which it derives clauses.
* `live`: Go-only reachability and fairness-parser substitutes were removed.
  Command and option behavior follows the source, including its unfinished
  paths. Package race tests and vet passed.
* `pdr`: source solver context, limits, helper edge cases and graph lifecycle
  are covered. Differential tests exercise 64 machines and source-matched
  option combinations, plus invariant-file output. Package race tests and vet
  passed; some cases expose inherited invalid-counterexample behavior in C.
* `ssc`: evolving FRAIG substitutions, BSAT frontier encoding, source-style
  class refinement, native-input lowering, random stream, equivalence
  reduction, and CEC verification have source tests. The constrained sweep
  wrapper now levelizes its subject as C does, and simulation snapshots and
  counters match source behavior. Complete option parity remains under review.
* `ssw`: source rarity, SAT polarity, prefix, class-count, manager-lifecycle,
  and timeout behavior is restored and characterized. A reachable invalid
  rarity witness is documented as an upstream C defect. Package race tests and
  vet passed.

## Shared dependencies in progress

`abcgo/sat/cnf` now uses ABC's compressed MSOP table, DAR cut ordering,
dominance/eviction, area-flow mapping, mapped CNF writer, and variable
numbering. Direct compiled-C comparisons matched exact ordered DIMACS output
for 345 generated asserted, free-output, and sequential cases; exhaustive
three-input truth-function checks also pass. `abcgo/opt/dar` cut enumeration
follows C's slot ordering and eviction behavior. `abcgo/sat/bsat` has a
source-style cumulative resource-limit API. `abcgo/sat/msat` now matches C's
chunk pools, clause allocation accounting, learned-clause sorting, vector
ordering and accessors; source-by-source comparison with MiniSat C continues.
The original C checkout only declares the unused variable-heap API, so its Go
replacement bodies were removed. A final `go test -race` and `go vet` run
across `./proof/...` is still required after active reviews settle.

Run full checks from `abcgo`, excluding experimental directories, for example:

```sh
GOCACHE=/tmp/abcgo-cache go test -race -count=1 -p 8 ./proof/...
GOCACHE=/tmp/abcgo-cache go vet ./proof/...
```
