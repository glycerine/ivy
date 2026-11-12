# Output cone decomposition review

Files reviewed: `saigOutDec.go`, `saigOutDec_test.go`.

## Port behavior

- Finds the primary property output cone in depth-first postorder, excluding its root as in `Aig_ManDfsNodes` plus the source's root skip.
- Uses the shared AIG-to-CNF encoder and one persistent incremental SAT solver. Each query assumes the property output false and applies candidate literals in the source phase/mask order.
- Preserves the source's single-literal marking rule and first satisfying two-literal polarity choice.
- Rebuilds original primary outputs, prime outputs, and latch inputs in that order, preserving register and constraint counts.

## Review checks

- Complemented cube literals are translated back to regular-node edges and negated exactly once.
- `known` tracks copied nodes independently of edge value, since constant-one is a valid mapped edge and cannot act as a missing-value sentinel.
- Register outputs remain the final outputs after prime outputs are appended, as required by the graph register convention.
- SAT queries use `Solve` assumptions on the same solver; no clauses or source graph metadata are mutated during prime discovery.
- The semantic regression exhaustively compares original and decomposed outputs for all assignments of a three-input graph and checks the expected pair prime.

## Gates

Explicit-file normal, race, and vet commands passed on the shared SAIG slice; the current file list is recorded in the parent task update. This is a source-based port with exhaustive small-graph tests; no direct invocation of the C helper was available for an output-level oracle comparison.
