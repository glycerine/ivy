# bsat port status

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
