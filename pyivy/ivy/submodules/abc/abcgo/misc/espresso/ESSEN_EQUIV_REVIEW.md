# Equivalence and essentiality review

Ported every function in equiv.c and essen.c: output equivalence diagnostics,
mutual-cover equivalence, essential extraction, single-cube essentiality, and
both sharp-consensus cases. Uses the authoritative packed cube/cover, cubeList,
PLA labels, consensus, cofactor and complement implementations.

## Compactness/readability pass

The two containment directions share one small loop. Output diagnostics use an
io.Writer helper, preserving stdout behavior while enabling direct tests without
process-global output interception. Essential extraction preserves source row
order and ACTIVE/RELESSEN/NONESSEN semantics; consensus scratch is local instead
of borrowing global temporary slots. No alternative geometry or graph model was
introduced. Debug-only essential-prime progress printing is omitted.

## Bug-review pass

A direct regression first failed check_equiv(a,a): cofactor excludes a query row
by pointer identity because irredundancy queries need to omit that row. Reusing
that query for equivalence incorrectly removes the only source of coverage.
Equivalence now supplies an independently owned query cube, so identical or
shared-row covers are reflexively equivalent. Separate strict-containment,
empty-cover, and differently decomposed-equivalent cases pass. Both temporary
lists are released on all early exits; the original C returned without doing so.

Checked essential removal/append ordering, don't-care preservation, candidate
NONESSEN suppression, RELESSEN gating, consensus distances zero/one/two,
multiple-valued sharp-consensus masks, and borrowed input row ownership. Tests
also validate positive/negative output equivalence reports with generated labels.

## Original-C differential evidence

160 deterministic mixed binary/multiple-valued original-C cases are recorded in
test_vectors/essen_oracle.json. The reproducible build_essen_oracle.py compiles the
unchanged original Espresso source files with portability header glue. Tests
compare exact ordered consensus rows, essential rows, remaining onset rows,
appended don't-care rows, per-cube essential predicates and equivalence results.
Candidate flags vary to exercise suppression and movement. Source behavior is
preserved on these distinct-input cases; reflexivity is intentionally corrected
by the directed regression above.

The combined actual-source complement/cofactor/irred/essential/equivalence,
cover/geometry/output and sparse/mincov subset passes tests, race tests and vet.
