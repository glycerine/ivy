# Proof log port review

## Implemented behavior

The original `src/sat/proof/pr.c` format is a custom DIMACS-like file, not plain
DIMACS: its header is `p <variables> <clauses> <roots> [A-clauses]`. The Go reader
parses and validates that format, classifies the root/A prefixes, and checks
literal bounds and declared clause counts. `Pr_ManProofWrite` propagates root
units, checks each learned clause by reverse unit propagation, derives a
resolution chain from recorded implication reasons, and optionally writes the
original `proof.cnf_` root/resolution record format.

The implementation uses native Go slices and the same packed literal format.
It preserves the C proof algorithm's root propagation, learned-clause
assumptions, resolution ordering, unit-clause propagation, and interpolant
combination. A simple unsatisfiable proof is independently checked against
`sat/bsat2` in the tests.

## Bug and readability review

- Clause insertion copies and sorts its input; it rejects repeated variables,
  invalid flags, A clauses outside the initial root prefix, and roots after
  learned clauses instead of relying on C assertions.
- The proof reader reports malformed headers, missing zero terminators,
  out-of-range literals, and clause-count mismatches with line context.
- Learned-clause propagation includes only roots and previously accepted
  learned clauses. Future learned clauses are stored for the proof pass but do
  not leak into earlier RUP checks.
- Resolution pivots and final weakening are checked before a proof ID is
  accepted. A database with a terminal empty clause must derive that clause.
- Interpolation's C implementation indexes a fixed five-entry truth table.
  The Go port returns an error when more than five shared variables occur,
  avoiding the source's out-of-bounds access.
- Proof file ownership and manager cleanup are explicit; input/output paths are
  test-local and no `testdata` directory is used.

Follow-up review rejects any clause after a terminal empty clause without
mutating the proof, records root-level contradictions (including interpolants),
and accepts learned clauses already implied by root assignments. Regressions
exercise repeated recording and direct and propagated root reasons. The full
package passes `go test -race -count=1 ./sat/proof` and `go vet ./sat/proof`
with `GOCACHE=/tmp/abcgo-cache`; finite examples do not prove all proof logs.

## Limits

- The parser intentionally follows the custom ABC header, not the broader
  `p cnf <vars> <clauses>` DIMACS header.
- `Pr_ManPropagate` uses a straightforward clause scan to fixpoint rather than
  pointer-linked two-watch lists. This retains semantics but can be slower on
  large proof logs; the source's 70 MB `stats.txt` proof is a future scale test.
- Interpolants retain the source's 32-bit truth-table representation and thus
  support at most five shared variables.
