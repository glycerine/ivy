# Native parent CEC integration

The parent package is still incomplete. Native integration currently covers parameter defaults, SAT/Fra manager lifecycle, incremental node and pair queries, independently evaluated witness values, dual-output verification, simple verification, and two-network verification. The original low-level CNF/activity, sweep/correspondence/choice, sequential and alternative solver files remain migration work. No product source was hidden or excluded from the ordinary package build.

## Readability review

The SAT manager owns one core manager and one last-query witness, rather than duplicate source solver/variable/activity caches. Shared GIA/vector types are aliases; AIG conversion aliases live separately. Queries share one status/error/model path. Verification shares options, deadline and witness attachment logic. Source status values remain 1 proved, 0 disproved, -1 undecided; typed errors are retained in the manager or verification parameters.

The high-level verifier uses the exact SAT equivalence engine rather than reimplementing the original synthesis/simulation optimization pipeline. This preserves proof/witness semantics while resource-limited outcomes and performance can differ. Optimization flags that select the unfinished pipeline do not currently select different native algorithms.

## Separate bug review

Checked regular versus complemented object identities, SAT witness lifetime, mutable conflict budgets, deadline propagation, dual-miter pair indexing and witness ownership. Fixed complemented object queries to encode the regular ID plus complement bit. UNSAT and failed queries clear the previous witness before returning. SAT values are computed independently from the immutable model, not read from phase-flipped CNF literals. Rejected negative conflict budgets preserve the solver's learned clauses. Counterexamples are attached only to the first compared manager and are cleared before a new verification run.

## Validation

`test_native.sh -race` passes actual selected GIA/CEC integration tests in a temporary module. It does not establish that the full parent packages build. `go test -race ./proof/cec/core` passes, including the conflict-budget update regression. The ordinary parent build currently stops in untranslated GIA files beginning with giaBalAig.go before reaching the remaining CEC drafts.

The native group also includes per-output miter solving, status vectors, cone-only pattern cubes, pattern stores, source-shaped counterexample generation, special-case proofs and the old verification entry point. Cube literal polarity follows `Abc_Var2Lit(CI, !value)`; a dedicated negated-input test catches reversal. Empty and stale cubes are cleared before each query. Constant/input/unary special cases check basis assignments before declaring equivalence, avoiding false proofs through buffer aliases. AIG-to-AIG witness transfer, correspondence and choice remain pending.
