# xSAT Go port and reviews

This package ports Bruno Schmitt's actual xSAT C implementation into independent
Go algorithms. It imports neither Satoko nor MSAT. Search retains integer
variable/clause activities and rescaling, separate binary and long-clause watch
lists, first-UIP conflict resolution, recursive/binary learnt minimization,
nonchronological backtracking, saved polarity, LBD restart/blocking queues,
protected-clause reduction and handle-updating garbage collection.

All seven public operations in xsat.h have their original xSAT_* spelling in
package-local compatibility functions. Public Go entry points include New,
AddClause, Simplify, Solve and ModelValue. Literal representation remains
`variable*2+sign`; status values remain SAT=1, UNSAT=-1, undecided=0. The original
Vec_Int_t clause interface uses the ported misc/vec package and retains sorting
and shrinking; native AddClause preserves its caller's Go slice.

Storage is Go-owned. Clauses separate literals and activity instead of a union
or flexible array. Handles index clause slots rather than raw arena words;
MemCap/WastedCap report slot counts. Collection rewrites original/learnt clause
lists, binary/long watches and all reason handles. Go slices are authoritative
for vector, heap, watch-list and queue sizes/capacities.

## Separate review passes

1. **Compactness/idiomatic/readability:** reviewed all twelve implementation
   files. Shared normalization avoids divergent native/compatibility rules;
   small typed owners replace raw memory/tagging; queue ring position and sum
   are the only cached state; standard-library sorting replaces pointer casts.
   Removed redundant storage padding, clarified ring backing-array access, and
   preserved original component boundaries and operation names.
2. **Independent bug/invariant review:** checked normalization, empty/invalid
   clauses, binary and long propagation, conflict watch tails, root state across
   repeated solves, phase/order inverse mappings, rescaling, reduction freezes,
   all relocated handles, limits/resumption, models and destruction. Added red
   regressions and fixed singleton low-level allocation partially modifying
   the solver before panicking, and missing watch removal silently succeeding.
   Negative clauses are validated before growing variables. Targeted tests
   cover watch-tail preservation, GC reason/watch rewriting, incremental
   contradictions, queue mixed pop/push wrap, and budget/deadline recovery.

## ABC_FREE audit

Compared the original destroy/release paths in xsatSolverAPI.c, xsatMemory.h,
xsatHeap.h, xsatWatchList.h and xsatBQueue.h with the rewritten owners. No broad
textual ABC_FREE replacement remains in these files. Destroy zeros the entire
solver owner, releasing memory, watch, heap, queue, assignment and model
references. Low-level release functions nil their owned slices or reset typed
owners. Tests verify repeat destruction/release and arena restart. Go aliases
can retain referenced objects safely; they are not forcibly invalidated as C
freed pointers would be. No cgo, unsafe memory access, hidden draft files or
empty production solver callbacks are used.

## Evidence and boundaries

- Full-package tests, race tests and vet pass.
- 1,500 unrestricted random formulas and 400 denser proper 3-SAT formulas are
  compared against exhaustive enumeration, with repeated solves and SAT model
  validation. Pigeonhole UNSAT exercises learning.
- Parser/malformed input, original vector API, normalization ownership,
  incremental units/conflicts, watch manipulation, queue wrap/sum overflow,
  limits and cleanup are covered.
- `XSAT_C_DIFFERENTIAL=1 go test ./sat/xsat -run TestOriginalCDifferential`
  compiles the original C sources and compares 600 mixed binary/3-/4-literal
  formulas. Its driver supplies ABC's ordinary non-bridge diagnostic endpoints
  needed for standalone linkage; solver code is unchanged. Production requires
  no compiler or C runtime dependency.

Search traces and diagnostic formatting are not promised identical: physical
storage and stable handle values differ, and propagation statistics are counted
consistently even when a binary conflict returns early. Go honors the declared
conflict/propagation/deadline limits and returns undecided instead of repeatedly
restarting under an exhausted limit; the original search does not consult those
configuration fields. Empty clauses and repeated root-inconsistent solves have
defined UNSAT behavior. No known package build blockers remain.

Follow-up lifecycle coverage compares 400 seeded incremental formulas with
exhaustive assignments while repeatedly simplifying and collecting clauses.
The full package passes `go test -race -count=1 ./sat/xsat` and
`go vet ./sat/xsat` with `GOCACHE=/tmp/abcgo-cache`. This does not prove all
search heuristics, memory pressure, or resource-limit interactions.
