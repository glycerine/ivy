# Proof package porting progress

The target is a faithful Go port of each compiled C proof path, with source
reviews and tests that cover the algorithms and option branches. Do not treat
Go-only fallback algorithms as completed ports. When the checked-in C path has
a defect or is incomplete, preserve that fact in the porting notes and tests
rather than silently substituting another algorithm.

## Current checkpoint (2026-11-13)

Proof-package work has been distributed across parallel lanes. Many packages
have fresh package-level race tests and `go vet` passes, but a repository-wide
proof gate has not yet been rerun after the current source-fidelity changes.
The shared `sat/cnf` port is active, so dependent package checks may change as
that work lands. The integrating thread is committing green package-sized
checkpoints on the current branch.

## Package status

* `abs`: source boundary tests cover the C VTA assertion and retained witness
  behavior. Native MUX controls are included by Go graph helpers where those
  helpers accept native GIA nodes. Full race tests and vet passed; a package
  audit is continuing.
* `acec`: source flow and preconditions are restored. Compiled-ABC differential
  tests cover 76 circuit pairs. The C implementation has reproducible
  false-equivalence cases, documented and tested as source behavior. Package
  race tests and vet passed.
* `cec`: the naive verifier now uses MF CNF/BSAT, with the C output-variable
  mapping. Compiled ABC comparisons confirm source-inherited false speculative
  equivalence pairs and a reordered-output misclassification; tests preserve
  and document those defects. Package race tests and vet passed. Cec4's
  upstream choice builder remains incomplete.
* `cec/core`: source-named node checks and SAT-value accessors were corrected;
  the separate Go equivalence helper remains separate. Mixed-gate truth-table,
  witness, recycling and source-wrapper tests passed under race and vet.
* `dch`: uses the ported BSAT path, source SAT-variable numbering and allocation,
  supergate/MUX clauses, incremental working FRAIG, and source simulation
  behavior. Native XOR/buffer and non-topological inputs outside the C source
  contract are rejected. Package race tests and vet passed.
* `fra`: SEC, clause and BMC paths follow the active source's timeout, budget,
  and solver-reuse behavior. Source-derived partition, implication, one-hot,
  and induction tests passed; broader behavioral audit remains in progress.
* `fracomb`: AIG deadlines now propagate through DAR rewrite stages as in C.
  Its current SAT path still depends on the shared CNF port; the former
  gate-by-gate encoding differs from C's cut-mapped CNF.
* `fraig`: source behavior has been restored for miter checks, external clauses,
  feedback compression, pointer-based hashing/order, and MSAT behavior. Several
  upstream limitations, including witnesses that violate external clauses,
  are covered by source-characterization tests. MSAT review continues.
* `framgr`: duplicate BMC helpers were consolidated only where behavior was
  shown identical. Simulation saturation and callback tests passed; broader
  manager audit remains incomplete.
* `int`: MiniSat 1.14p has exhaustive Craig-interpolant checks and option/error
  coverage. Race tests and vet passed.
* `int2`: the checked-in C interpolation driver is uncompilably incomplete, so
  the Go driver returns undecided and does not invent a bounded reachability
  fallback. Helper tests characterize source behavior. The suffix constructor
  still has inferred mapping around an undeclared C variable and is not yet
  established as fully source-faithful.
* `live`: Go-only reachability and fairness-parser substitutes were removed.
  Command and option behavior follows the source, including its unfinished
  paths. Package race tests and vet passed.
* `pdr`: source solver context, limits, helper edge cases and graph lifecycle
  are covered. Differential tests exercise 64 machines and source-matched
  option combinations, plus invariant-file output. Package race tests and vet
  passed; some cases expose inherited invalid-counterexample behavior in C.
* `ssc`: evolving FRAIG substitutions, BSAT frontier encoding, source-style
  class refinement, native-input lowering, random stream, equivalence
  reduction, and CEC verification have source tests. The shared CNF mapping
  dependency is still being completed; broader source parity remains open.
* `ssw`: source rarity, SAT polarity, prefix, class-count, manager-lifecycle,
  and timeout behavior is restored and characterized. A reachable invalid
  rarity witness is documented as an upstream C defect. Package race tests and
  vet passed.

## Shared dependencies in progress

`abcgo/sat/cnf` is being aligned with ABC's compressed MSOP table, DAR cut
ordering/dominance/eviction, area-flow mapping, mapped CNF writer, and variable
numbering. `abcgo/opt/dar` cut enumeration now follows C's slot ordering and
eviction behavior. `abcgo/sat/bsat` has a source-style cumulative resource
limit API. `abcgo/sat/msat` is undergoing source-by-source comparison with the
MiniSat C implementation. These shared changes need a final `go test -race`
and `go vet` run across `./proof/...` once integration stabilizes.

Run full checks from `abcgo`, excluding experimental directories, for example:

```sh
GOCACHE=/tmp/abcgo-cache go test -race -count=1 -p 8 ./proof/...
GOCACHE=/tmp/abcgo-cache go vet ./proof/...
```
