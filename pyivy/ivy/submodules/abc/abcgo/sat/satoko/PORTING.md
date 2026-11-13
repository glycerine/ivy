# Satoko Go port

Production code uses only Go and its standard library plus the Go-owned Satoko
utility packages in this subtree. Literal encoding remains `variable*2+sign`;
sign one negates the variable. Public Go entry points are `New`, `AddClause`,
`Solve`, `SolveAssumptions`, `ModelValue`, `ParseDIMACS`, and the lifecycle methods
on `Solver`. Package-local `satoko_*` wrappers retain the original operation
names for ongoing ABC call-site translation.

The solver uses two-watched-literal propagation, first-UIP conflict analysis,
recursive reason-based learnt minimization, binary resolution, nonchronological
backtracking, saved phases, indexed activity ordering with the original packed
activity arithmetic, LBD restarts with trail-based blocking, and learnt-clause
reduction. Clause storage is Go-owned; IDs remain stable and removed literal
payloads become reclaimable without raw arena relocation.

## Evidence

- `go test ./sat/satoko/...` tests the entire subtree, not isolated file subsets.
- `go test -race ./sat/satoko/...` passes.
- Random formulas and assumptions are checked against exhaustive enumeration;
  SAT models are checked against every input clause. Additional random 3-SAT
  and pigeonhole tests exercise conflict learning.
- Incremental solve, assumptions, final cores/minimization, bookmark/rollback,
  limits/cancellation, cone selection, DIMACS round trips and malformed input
  are covered.
- `SATOKO_C_DIFFERENTIAL=1 go test ./sat/satoko -run TestOriginalCDifferential`
  compiles the original C source as an optional test oracle and compares 900
  formula/assumption runs plus an incremental-assumption/reset lifecycle. Production builds do not require a C compiler.

## Fidelity boundaries still under audit

This is a functional Go CDCL port, not a claim of identical internal search
traces. Watch layout, allocation, selection-sort/quicksort implementation,
reference lifetime differ. Recursive/binary minimization, protected-LBD
reduction, special reduction intervals, and blocked restarts are implemented;
secondary search trace details still need comparison on larger benchmarks.
`NoSimplify` currently controls no additional in-place clause rewriting: root
propagation always runs, while clauses are retained in their logical form.
`GarbageRatio` triggers rebuilding watch lists after sufficient removal; Go
reclaims deleted literal buffers without raw arena relocation. These tuning differences do not change the represented CNF.

Go deliberately reports `UNSAT` when a previously added clause made the formula
inconsistent; the original C `satoko_solve` reports `UNDEC` for that state.
The optional C oracle normalizes immediate DIMACS parse conflicts to UNSAT,
and explicitly creates unused header-declared variables before assumptions:
the original C parser otherwise aborts without publishing the solver for an
immediate conflict and does not allocate unused declared variables.

C vector insert's shift count omitted the final element. The Go vector insert
shifts the complete suffix, which is tested. Go vector initialization fills
actual typed values rather than using the C byte-pattern memset behavior.

The initial optional C lifecycle audit also attempted bookmark/add/solve/rollback
and the original C program segfaulted on that sequence. The green C lifecycle
comparison uses reset instead; Go rollback is verified separately against its
restored formula. Original C rollback memory behavior is not reproduced.

## Separate review passes (2026-11-12)

1. Compactness, idiomatic Go and readability: all five vector structs store only
   slices; the bounded queue now stores only its slice, replacement cursor and
   running sum. Removed duplicate capacity/size metadata, grouped imports and
   expanded lifecycle code. Public APIs remain stable; full subtree tests pass.
2. Independent bug review: reviewed implication/watch invariants, assumptions,
   callbacks, limits, reset/rollback, dumping, typed vectors, heap and arithmetic.
   Added failing regressions before fixing temporary assumptions/limits leaking
   after recovered panics, stale final cores after minimization, a spurious
   zero-based dump terminator, rollback literal counters, and reset losing
   cancellation settings. Regression and race suites cover the resulting fixes.

### Source-ordered conflict-cost audit (2026-11-12)

The GIA tuning integration exposed search-order differences despite matching SAT
answers. Instrumented original-C and Go traces identified the first divergent
decisions rather than compensating the conflict counter. Corrections preserve
implicit negative variable polarity, enqueue-time phase saving, reverse trail
reinsertion, binary-before-long watch priority and blocker caching, binary
minimization's tail-swap order, and the additional learnt-reason activity bump.
Root assignments now persist during clause insertion; root-satisfied clauses are
omitted and root-false literals removed. Root simplification rebuilds the heap
bottom-up with the original propagation budget, including the source's reset of
literal statistics at solve entry. Unit records remain Go-owned for dumping and
rollback; they do not contribute to source clause-literal statistics.

Readability pass: kept stable clause IDs and existing API ownership; isolated
watch insertion and root order maintenance into named helpers. The blocker map
keys each literal/clause watch independently, so moving one watch does not alter
the other. Retained original clauses support the existing rollback contract;
`simplified` records removal from source simplification accounting.

Bug/invariant pass: checked implied phase timing, root-unit clause normalization,
heap membership during cancellation, binary propagation before long watches,
root decision order, and propagation accounting on binary versus long conflicts.
Focused regressions cover implicit/explicit polarity, binary watch priority,
persistent root implications, and four original-C conflict costs from seeded
formulas. The independent GIA oracle matches status and conflict count on its
100 seeded source-ordered formulas. Full Satoko subtree race tests, existing
original-C status differential, and vet pass. This evidence covers those cases;
it does not establish complete search-trace equivalence for arbitrary restart,
reduction, assumption, or incremental histories.

### Follow-up incremental review

Clause and cone inputs are fully validated before changing solver state. A
seeded rollback oracle checks 300 formulas and repeated post-rollback
assumptions against exhaustive evaluation. Full subtree race tests and vet pass
with `GOCACHE=/tmp/abcgo-cache`; these finite cases do not establish search-trace
or performance equivalence for all workloads.
