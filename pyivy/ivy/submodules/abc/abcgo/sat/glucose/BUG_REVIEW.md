# Separate bug review

Checked legacy behavior against its own unchanged C++ sources, independently of
Glucose2. Native zero-budget/interruption and nondecision-model differences have
specific regressions. Native default-core and preprocessing request-stream hashes
match. Random formula tests compare 1,500 formulas x five assumption queries with
exhaustive evaluation, validate models, exercise aggressive restarts/reduction,
check repeated assumptions, root contradictions, allocator collection and reset.

Inherited common-algorithm fixes retain their regressions: allocator free-after-
mark accounting, opposite/direct clause subsumption, queue resize sums, mutable
pointer map-key hashing, final-element region inverse lookup, UTF-8 byte matching,
malformed integer/header/unterminated DIMACS errors and gzip/read failures.

The source's optional redundant-clause checker can drop a resolvent because its
parents imply it, then remove those parents. Derived replacement clauses bypass
that check. The native bug harness and four-variable red regression remain
visible. Substitution records reconstruct substituted model values correctly.
1,500 random preprocessing cases mix asymmetric branching/redundancy checking
and verify original clauses against reconstructed models.

Bridge review found and fixed two representation/contract issues: traversing the
second fanin of an explicit output reached InvalidEdge (observed red panic), and
overlap must report SAT of both literals, while equivalence reports UNSAT of both
polarities. Tests distinguish an input overlapping itself from its complement.
Constant SOP rows are handled before normal cube decoding. Command-file solving
preserves the frame's existing graph witness rather than replacing it.

Bridge validation enumerates every one of the 256 three-variable Boolean
functions: checks AIG SAT and model, generated SOP truth, every existential
projection mask, and agreement of independent SAT-cube and Shannon-cofactor
quantification paths. Separate regressions cover assumption-core minimality,
AND/XOR phase encodings, callbacks, stop flags, original inert circuit hooks,
command registration/flags, CNF-file handling and witness validity. DRUP output
is parsed and each addition independently checked by RUP, including empty clause.

Follow-up review fixed eliminated assumptions bypassing validation when
preprocessing is off, invalid substitutions mutating the solver, stale models
after preprocessing failure, and DIMACS exports losing assumptions that
contradict root assignments. Regressions cover each case. A 300-formula
incremental preprocessing oracle validates repeated queries and reconstructed
models; a separate 4,000-query oracle validates assumption conflict cores.
These finite tests do not prove all solver or resource-limit behavior.

Passed:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./sat/glucose ./aig/aig ./base/main
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./sat/glucose
```
