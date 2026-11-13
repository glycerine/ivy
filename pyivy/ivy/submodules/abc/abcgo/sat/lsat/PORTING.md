# LSAT Go interface

`src/sat/lsat` contains only `solver.h` in this checkout: it declares an opaque
SAT interface but provides no implementation. The Go package implements those
operations through the pure-Go Satoko CDCL engine, avoiding a second duplicated
search implementation or an external SAT dependency.

Public Go entry points include `New`, `MakeLiteral`, `AddClause`, `Solve`,
`ModelValue` and the clause/assumption builders. Every active function declaration
in the original header has a package-local compatibility wrapper in `compat.go`.
Literal representation is `variable*2+negativeSign`; three-valued results are
`True`, `False`, `Undefined`. Decision eligibility suppresses branching only,
never implication propagation. Current root assignments and solved models are
separate queries, and returned conflicts are independent snapshots.

Validation includes exhaustive enumeration of 300 random formulas, builder and
incremental solves, assumptions/conflicts, truth/sign conventions, phase choice,
nondecision propagation, counts, model/root distinction, and buffer ownership.
The implementation dependency is validated by the Satoko subtree's larger
exhaustive and optional original-C differential suites. No original LSAT search
implementation is present for a direct LSAT-specific search-trace comparison.

## Separate review passes

1. Compactness/idiomatic/readability: split the original operation-name wrappers
   from the public interface, use Go slices for builders, use typed Truth values,
   and reuse the already-tested Go engine rather than duplicate its storage or
   algorithms. The original license remains in solver.go.
2. Independent bug review: inspected sign/undefined handling, builder ownership,
   temporary assumptions, model/current-root distinction, clause inconsistency,
   decision eligibility and conflict ownership. Added regressions for the last
   four state/ownership cases; they pass. Engine enhancements needed by this
   frontend (polarity, decision eligibility and root-value queries) are covered
   by frontend propagation and phase-selection tests.

Follow-up tests compare 3,600 incremental queries and their conflict cores
against exhaustive assignments. Invalid clauses are validated before the shared
Satoko engine grows its variable set; a regression checks that the solver
remains usable after the panic. `go test -race -count=1 ./sat/lsat ./sat/satoko`
and `go vet ./sat/lsat ./sat/satoko` pass with `GOCACHE=/tmp/abcgo-cache`.
