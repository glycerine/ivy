# bsat port status

Current state: the entire package builds and passes `go test ./sat/bsat`; `go test -race ./sat/bsat` and `go vet ./sat/bsat` also pass. All solver, store, proof, interpolation, allocator, and utility files are included. See “Completed package port and final review” below for the current APIs, representation adaptations, dependencies, and review results.

The following is a chronological development log. Statements about incomplete builds describe the milestone at which they were written.

Completed `satVec.go`: literal/DIMACS encoding, lbool and statistics types,
integer/pointer vector allocation, growth, shrinking, pop and ordered removal.
Vectors own Go slices and pointer-vector removal/shrinking clears references.
`veci_begin` and `vecp_begin` now return live slices instead of C pointers;
solver consumers must adopt slice operations as they are ported.

Validation:

```
GOCACHE=/tmp/abcgo-cache go test abcgo/sat/bsat/satVec.go abcgo/sat/bsat/satVec_test.go
```

Tests cover 2,000 signed-literal round trips, literal validity boundaries,
1,000-element vector growth and removal, vector reuse, pointer ownership and
invalid operations. The full package is not yet buildable: `go -C abcgo test
./sat/bsat` fails on retained C assignments in `satChecker.go`, followed by
other generated drafts. No files are excluded from the package to conceal this.

Next dependency is `satClause.go` (~400 lines): C's paged packed clause words
must become typed Go clauses plus stable integer handles. Preserve learned-bit
handle tagging, two alternating page streams, bookmarks/rollback, hidden clause
IDs and learned-clause compaction/remapping. The solver variants and proof/
interpolation code depend on that representation; port them against the same
handle API rather than introducing unsafe pointer overlays.

Continued progress: `satClause.go` is now fully converted to typed clauses with
stable page-encoded handles, separate problem/learned streams, hidden IDs,
append/growth, restart, shrink, bookmarks/rollback and two-phase compaction.
Multi-page and remapping tests pass. C pointer/end pairs become slice APIs.
`pPages` retains only per-page word limits; actual clauses live in a handle map,
so C byte overlay assumptions must not reappear in solver consumers.

`drand`/`irand` now live in tested `satRandom.go`. The original solver's decision
heap update/pop algorithms now live in tested `satOrder.go`; `order_update`,
`order_unassigned` and `order_select` in satSolver.c.go delegate to them. Heap
positions and activity arrays in the solver header are slices. Remaining solver
activity initialization/rescaling still needs conversion to match these slices.

The isolated test command now includes satVec.go, satClause.go, satRandom.go,
satOrder.go and their four matching test files. Full package compilation remains
unproven and blocked by the other generated files; this is an incomplete port.

Further CDCL/proof ports now wired into original `satSolver.c.go`:

- `satChecker.go`: resolution proof parsing/checking with `io.Reader`, signed
  literal slices, numbered roots/derivations and backwards-parent validation.
- `satAssignments.go`: assignment/trail state, reasons, enqueue, decision,
  backtracking, top-level rollback and phase-saving behavior. Embedded into the
  existing solver struct. Constructor/growth code still needs conversion to
  initialize/grow these slices before the full solver is runnable.
- `satPropagation.go`: original binary and two-watch propagation, watch-list
  compaction, replacement watch movement, conflict records and LBD estimate.
- `satAnalyze.go`: original first-UIP conflict analysis, activity hooks,
  recursive reason minimization, tag restoration and learned backtrack literal.
- `satAnalyzeFinal.go`: backwards assumption-conflict tracing, including tagged
  binary reasons and ordinary reason clauses.

Separate readability/idiomatic review: named typed helpers replace raw pointer
arithmetic, preserve original compatibility entry points, and avoid closures in
solver adapters. Proof pivot search now uses map membership rather than a
quadratic parent cross-product. No disabled/excluded generated files.

Separate bug-review pass and fixes: reject tautological proof parents that could
otherwise manufacture an empty clause from a satisfiable root set; reject signed
literal encoding overflow; preserve remaining watches on early conflict; verify
all-deleted compaction bookmarks; guard invalid decision/rollback/trail states.
Each finding has a regression. Differential evidence includes 500 deterministic
random formulas comparing watched propagation to an independent naive closure,
and 300 conflict-analysis formulas checking every learned clause against all 32
possible assignments. Tests also verify first-UIP units, asserting backtracking,
phase saving, tag rollback/clearing, final assumption cores and proof syntax.

All nine converted files and their tests pass together with:

```
GOCACHE=/tmp/abcgo-cache go test abcgo/sat/bsat/sat{Vec,Clause,Random,Order,Checker,Assignments,Propagation,Analyze,AnalyzeFinal}.go abcgo/sat/bsat/*_test.go
```

Full package build was rerun: the first current syntax failures are now in
`satInter.go` (casts/embedded increments), rather than the converted proof
checker. Substantial original constructor/growth, activities, clause database,
search/restarts, solver2/3, proof recording and interpolation work remains.
No claim of a runnable complete bsat package is made yet.

Main solver continuation (current): the original primary CDCL solver now runs
within the converted-file set. Additional ports include real solver types,
constructors/variable growth/restart, original decision wrappers, binary/general
clause insertion, all variable activity modes (integer, IEEE-double and ABC's
extended-range format), clause activity, learned database selection/compaction,
proof clause storage with file load/dump, input normalization/learned recording,
search/Luby restarts, assumptions, resource limits/deadlines and model access.
These replace corresponding original generated function bodies; they are not a
separate toy solver or excluded-source build.

A further readability pass separated lifecycle, activity, search, storage and
assumption APIs into small typed files, using slices, standard numeric bit
conversions, standard sorting and time. Separate bug review/regressions cover
empty solver selection, empty/zero-ratio database reduction, source proof-reader
ASCII-to-integer accumulation, EOF after numbers, header counts/zero-based IDs,
root/A metadata restoration, and ownership/chunk releases. The original reader
could increment declared clause totals while loading and fail its own count
check; the Go reader tracks declared counts separately.

End-to-end evidence: 500 deterministic random CNFs with assumptions are checked
against exhaustive truth assignments, including returned models. Further tests
cover incremental assumptions/clauses, learned binary IDs, restart budgets,
expired deadlines, Luby sequence, UNSAT proof-clause recording, storage roundtrip,
malformed input, activity arithmetic and representation rescaling. All converted
files and tests pass together. Full package still fails on unported interpolation,
solver2/3 and other legacy drafts. Primary rollback/lex-SAT/minimization/header
utilities also remain to be ported; no full-package completion claim is made.

Current additional milestones:

- Primary satSolver.c.go and satSolver.h.go are fully valid Go; the remaining
  primary functions (rollback, lex-SAT, assumption minimizers, gate encodings,
  controls/statistics/memory reporting) are ported, not disabled. Exhaustive gate
  tests exposed and fixed the original MUX first-clause output-complement omission.
- solver.go provides the public Go interface (`NewSolver`, encoded `Literal`,
  `AddClause`, `Solve`, `Value`/`Model`, conflict core, resource limits, deadlines,
  statistics and reset). A failed public AddClause persists UNSAT, including an
  empty clause; assumption-only UNSAT does not poison the formula.
- Sat_Mm allocators now adapt the shared misc/mem Go pools. Coordinated addition
  of Mem_StepRestart preserves the fixed first-chunk reuse contract; byte-pool
  adapter tests pass. Tests importing that shared package run from abcgo's module.
- satProof.go now holds a typed resolution DAG with word-offset handles,
  tagged root/antecedent references, iterative/recursive reachability, marking,
  compaction/root/pivot remapping and root-clause core derivation. Repeated core
  derivation cleans traversal marks. Invalid forward proof edges are rejected.
- satProof2.go now records dependency bitsets using slices; a borrowed pointer to
  the mapping slice tracks caller reallocations (regression included).
- satTruth.c/h.go now interner truth tables modulo complementation using typed
  records, canonical hash chains and stable tagged handles. Tests cover elementary
  patterns across 256 points, complement identity/input preservation, rehashing
  and word operations.
- satSolver3.c/h.go now share the source-identical typed CDCL core. A comparison
  of original C function bodies (normalizing only prefixes/comments/whitespace)
  found seven differences: priority setup; absence of proof-store/user hooks;
  and the recursive minimizer. The type-1-only priority and solver3 minimizer
  remain distinct; no solver algorithm was replaced with a simpler substitute.
  The minimizer preserves 64-bit conflict limits instead of truncating the C
  temporary to int. Added 150 randomized solver3/exhaustive comparisons and
  variant priority/minimizer regressions. `NewSolver3` is available.

Latest connected test scope includes all converted files (primary solver,
solver3, stores, memory adapters, truth/proof managers and their tests). Full
package still does not compile due to the four interpolation implementations,
solver2/solver2i and remaining utility/trace consumers. Those sources remain
present and the package is not declared complete or ready as a whole.

### Resolution reconstruction and interpolation

Ported shared logged-clause RUP reconstruction with explicit watched-clause provenance, backward resolution, independent proof-checker output, deadline/error handling, and dependency recording. Truth interpolation and AIG variants A/B now annotate those actual resolution steps. A uses McMillan OR for A-local pivots and AND otherwise; B preserves shared-pivot multiplexing. Original partition-clause AIG construction is retained. AIG uses the real `aig/aig` graph implementation, with owned result managers.

Validation: connected converted-source suite passes, including 100 randomized truth interpolation Craig checks and 150 partitioned-formula AIG trials checking both variants exhaustively over assignments. Resolution output is independently checked; invalid logged clauses and expired deadlines have regressions. Full package still has solver2, proof-only interpolation, and utility/trace drafts to port.

Separate readability/idiomatic review: factored common propagation/reconstruction and AIG annotation mechanics, retained typed variant entry points and source licenses, removed repeatedly allocated `Graph.Inputs()` copies in favor of indexed API. Separate bug review: checked pivot polarity against original B mux branches, constant partitions, shared-variable ordering, nil writer interface handling, and proof trace validity; regressions exercise these paths.

Proof-only interpolation/core extraction now uses reconstructed antecedent dependencies, preserving root-only and postorder learned-clause modes, core re-solving verification, and BMC variable/frame printing. Regression excludes irrelevant roots and verifies the extracted root subset is UNSAT. Reviewed dependency indexing, root/empty handling, and result ownership separately from readability.

Explicit trace start/write/stop APIs are ported, with returned I/O errors and a 128-byte reserved header (avoids the original fixed 40-byte header overflow for large counters). Round-trip regression checks clause/root counts, signed literals, duplicate start rejection, and idempotent stop. Source solver does not automatically invoke these trace functions; explicit-write behavior is preserved.

### Completed package port and final review

All `sat/bsat` Go files now compile together. The previous incomplete-scope notes above describe intermediate milestones, not the current state. Solver2 now has a direct typed port of its explicit-clause propagation, root-unit proof production, first-UIP analysis, recursive minimization/resolution ordering, assumption conflict analysis, Luby search, limits, learned reduction/proof compaction, and bookmark rollback. Solver2 does not alias the primary CDCL implementation: only representation-independent assignment, heap, memory, and progress helpers are shared. Its activities retain unsigned 32-bit arithmetic.

`NewSolver`, `NewSolver2`, and `NewSolver3` are the usable Go entry points. `Solver2` retains one-based root-clause proof cores, including explicit proofs for contradictory input units and empty clauses. Interpolation managers transfer owned `aig.Graph` results. `AigInterpolationTest` (also exposed under the source `Gia_ManInterTest` name) preserves the two-CNF-copy interpolation harness using the shared AIG graph representation, with no dependency on the currently uncompiled GIA package. This is an intentional representation/API adaptation: callers should pass/receive `*aig.Graph`.

Utility port adds accurate DIMACS output including tagged binary clauses and assumptions, actual header counts, selected-variable model copies, file-directed statistics, and explicit I/O errors. The source-disabled `Sat_SolverDoubleClauses` entry point remains explicitly disabled, as it unconditionally asserts in ABC; its obsolete commented implementation was not revived. Historical `#if 0` proof APIs are not active ABC functionality and were not exposed as pretend implementations.

Separate compactness/idiomatic/readability pass: factored interpolation annotations and proof reconstruction; solver2 retains distinct proof-sensitive code while sharing generic state helpers; sliced arrays and explicit structs replace pointer arithmetic and packed variable fields; errors and ownership are explicit. Source licenses remain in the corresponding ported files. No build tags, hidden drafts, external C, or stubs are used.

Separate bug-review pass: added regressions for source2 unit-proof propagation, first-UIP proof roots, unsigned activity rescaling, learned-unit/proof-handle compaction, bookmark restoration, online interpolation, constant/XOR/AND circuit-relation interpolation, and DIMACS binary/header correctness. A red repeated-UNSAT test exposed consumed-conflict state allowing a later SAT answer; fixed by retaining a base-formula UNSAT result until rollback. Dynamic marked-level storage also handles levels exceeding the variable count under redundant assumptions. The new public solver2 wrapper handles contradictory inputs with actual root-resolution proofs instead of the C API's assertion.

Dependencies: Go standard library plus `github.com/glycerine/abcgo/misc/mem` (shared owned byte pools) and `github.com/glycerine/abcgo/aig/aig` (shared graph operations). No GIA, CUDD, CNF, cgo, unsafe, or platform runtime dependencies. Full-package tests include 200 randomized solver2/exhaustive formula comparisons and 100 assumption comparisons, with every generated UNSAT proof core checked against exhaustive satisfiability, in addition to the primary/solver3 and interpolation coverage recorded above.

Final bug review also found proof reduction could discard/remap the current final-refutation handle independently of learned clause roots. Reduction now retains and remaps that handle, with a dedicated regression. Full package validation: `GOCACHE=/tmp/abcgo-cache go test ./sat/bsat`, `go test -race ./sat/bsat`, and `go vet ./sat/bsat` pass.

### Dependent-port API bridges

Added public cone activity controls (`ClearActivityFactors`, deduplicating `SetActivityFactor`, `VariableActivity`), actual cumulative `Statistics.Inspections`, and `SolveWithGlobalLimits` with absolute conflict/propagation limits. Propagation and inspection counters remain distinct; dependent clients can enforce their original outer inspection budgets without changing the solver's existing `nInsLimit` meaning. Restart clears active cone metadata to prevent stale variable IDs after reuse.

Added proof-store ownership bridges (`EnableClauseStore`, partition/root markers, last-clause update, typed `ReleaseClauseStore`) and immutable owned clause snapshots/count accessors on `Sto_Man_t` for BMC clients. Regression tests cover actual activity bumps and all activity representations, reset/growth, cumulative budgets, inspection counters, released store proof reconstruction, and snapshot isolation. Separate API/readability review preserved concrete counter units and transfer semantics; bug review covered duplicate cone entries and reset stale IDs. Full bsat tests pass.

BMC bridge also exposes `NeedsPropagation`, `ConflictCount`, `PropagationCount`, consuming `TakeUnitLiterals`, `PropagateUnits`, and `Compress`, with root-conflict persistence and owned queue results covered by tests. Full BSAT and BSAT2 race/vet validation passes together.

`Solver.SolveLexicographic([]Literal) Status` exposes the existing exact lex-SAT algorithm for `misc/extra`. It mutates the ordered assignment vector, updates public model/status state, validates distinct introduced variables, and applies one per-call budget across all recursive subsolves. Empty order uses ordinary solving. The source lower-bound precondition is explicit: start with all-negative literals, then reuse returned assignments while blocking earlier models; arbitrary mixed-sign bounds are not a generic optimization interface. Public regressions enumerate all eight models in variable order (2,0,1), check literal/model agreement and assumption cleanup, then verify UNSAT; empty-order and deadline/resumption paths are also covered. Separate readability review kept this a direct wrapper; bug review checked limits initialization, known input conflicts, public status updates, mutation ownership, and source preconditions.

### Incremental root-conflict review

The public incremental solver and proof-producing Solver2 now propagate pending
root assignments before opening assumption levels. Without this step, an early
contradictory assumption could cause backtracking to skip a pending root
conflict, allowing a later solve to report SAT for an inconsistent formula.
Regressions exercise this sequence in both solvers and verify the proof core;
a deterministic incremental oracle compares 5,000 public-API calls with
exhaustive five-variable evaluation and validates models and assumption cores.

`GOCACHE=/tmp/abcgo-cache go test -race -count=1 ./sat/bsat ./sat/bmc` passes.
Further correctness, simplicity, and coverage review of `sat/bsat` remains in
progress; these tests alone do not complete that review.

The proof-producing `Solver2.NewVariable` now invalidates the preceding solve
status, matching the primary solver. Previously it exposed a stale model
(including a fabricated value for the new variable) until the next solve.
`TestSolver2NewVariableInvalidatesModel` covers invalidation and re-solving;
`GOCACHE=/tmp/abcgo-cache go test -race -count=1 ./sat/bsat` passes.

A repeated-solve proof oracle now validates public `Solver2` models and
root-clause proof cores against exhaustive four-variable assignments. It found
that an assumption contradicting a root unit returned an empty proof core;
the solver now records the unit's proof dependency. A focused regression also
checks that this assumption conflict does not poison a subsequent SAT solve.
`GOCACHE=/tmp/abcgo-cache go test -race -count=1 ./sat/bsat` passes.

The root-conflict review also covered lexicographic solving: an inconsistent
formula with pending root units could lose its conflict across a later ordinary
solve. Both entry points now share root-propagation handling, and a regression
checks repeated lexicographic and ordinary solves. The full race-enabled BSAT
suite passes after this simplification.
