# MSAT Go port and reviews

This package ports the actual C MSAT implementation into independent Go
algorithms. It does not import or delegate search to Satoko or another solver.
The implementation follows the `module.make` selection of `msatOrderH.c`:
indexed floating-activity ordering, queue-based watched-literal propagation,
first-UIP conflict resolution, nonchronological backtracking, activity decay and
rescaling, geometric restart limits and activity-ranked database reduction.
The unused `msatOrderJ.c` boundary-order alternative is exposed as BoundaryOrder
without introducing duplicate definitions of the configured order functions.

Original `Msat_*` operation names remain for the implemented C APIs. Native Go
entry points are New, AddClause, Solve and ModelValue. Variables/literals retain
ABC's `variable*2+sign` encoding. Model arrays retain encoded assigned literals.
Vectors use slices as authoritative size/capacity; clauses and byte pools have
Go-owned storage. Adjacency vectors are typed integer-vector slices instead of
C's casts between unrelated clause and integer-vector pointers.

## Separate review passes

1. **Compactness, idiomatic Go and readability:** inspected every ported file.
   Kept the original component boundaries and operation names, added a small
   native Go interface, removed redundant vector/queue capacity metadata,
   eliminated raw pointer tagging/flexible-array allocation, used an indexed
   slice heap, and removed unused proof/map state and redundant output work.
   Clause marks/numbers remain because callers can explicitly inspect them.
   Marked-clause removal follows the source clause-vector truncation behavior.
2. **Independent bug review:** checked clause normalization/watch ownership,
   root and assumption levels, conflict learning, repeated solves, models,
   lifecycle/reset/marks, limits, activity arithmetic, pools and containers.
   Failing regressions found forgotten root conflicts across later solves,
   stale models after contradictory clause additions, and marked rollback
   losing its base after units/simplification. Fixed these before integration.
   Marked rollback now follows C and does not restore simplified clauses or
   clear root assignments contributed by later unit clauses.
   Reviewed activity rescaling before a new clause enters the learnt vector,
   total clause/backtrack accessors, and C Prepare's binding-reset semantics.

## Verification

- Full-package tests, race tests and vet pass.
- 1,000 random formulas with four assumption trials are checked against exhaustive
  truth assignments, including every reported SAT model. Another 400 denser
  proper 3-SAT formulas exercise repeated learned-conflict solves.
- Pigeonhole UNSAT exercises learning. Tests cover DIMACS, container ownership,
  ordering, heap update/delete invariants, pool reuse/restart, marked rollback,
  truth-table construction and activity rescaling.
- `MSAT_C_DIFFERENTIAL=1 go test ./sat/msat -run TestOriginalCDifferential`
  compiles the original C source and compares 900 proper 3-SAT/assumption runs.
  The test driver initializes the C decision heap after parsing without calling
  Prepare, which would discard the parser's root unit assignments. Production
  code has no C compiler or cgo requirement.

The C source explicitly documents a tautology-normalization bug. The Go clause
normalization now follows that original in-place algorithm, including its
handling of repeated and opposite literals. A minimized case `(1 OR -6 OR 6 OR
-2 OR -3), (1 OR 3 OR 6), (-1)` under assumptions `-6, 2` is SAT, but the C
implementation reports UNSAT; a source-derived regression retains that result.
Randomized exhaustive formula tests use distinct-variable clauses to avoid
mischaracterizing this known source limitation.

Follow-up review added exhaustive incremental checks and source-derived marked-rollback tests.
Clause and assumption inputs are now validated fully before normalizing or
mutating solver state, including satisfied-clause and mixed-assumption cases.
`GOCACHE=/tmp/abcgo-cache go test -race -count=1 ./sat/msat` and
`GOCACHE=/tmp/abcgo-cache go vet ./sat/msat` pass. These finite tests do not
prove every heuristic or resource-limit path.

## Legacy header declarations

The following public declarations have no definition anywhere in the original
MSAT source in this checkout: Msat_SolverReadSolutions,
Msat_SolverReadSolutionsArray, Msat_SolverReadTruth, Msat_SolverSetProofWriting,
Msat_SolverSetVarMap and Msat_SolverSetVarTypeA. The internal
Msat_ClauseComputeTruth, Msat_ClauseCreateFake, and
Msat_ClauseCreateFakeLit declarations are likewise undefined there. They are not
silently implemented as fake proof, interpolation or enumeration support.
Call sites requiring those obsolete interfaces need a separate specification.
The clear standalone Msat_VarHeap_* header contract is implemented with the same
indexed heap algorithm and tested even though its original standalone bodies
are absent. No package build blockers or excluded generated files remain.

`Msat_SolverPrepare` now matches C and clears root assignments, including
units with no retained clause object. Callers must account for that source
behavior when preparing a new cone.

The source-named AddClause and Solve APIs do not retain a rejected-clause or
root-conflict status and do not clear the previous model when adding a clause;
callers must handle their return values as in C. The existing Go convenience
methods retain their own failure status and model invalidation behavior.

Clause creation reserves and recycles an entry of the C clause allocation size
in the step pool. This preserves the C pool's observable accounting and reuse.
The Go clause and literal objects remain separately allocated, so actual Go
heap usage includes that additional object overhead.
