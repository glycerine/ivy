# DSC implementation review

## Scope

Reviewed the Go implementation against `src/opt/dsc/dsc.c`, including initial
truth-table cofactors, AND/XOR grouping, cube specifications, decomposition
results, expression delimiter matching, and AND-equivalent gate counting.

## Findings

- Replaced C pointer arithmetic and fixed node arrays with owned Go slices and
  node records. The optional truth-word pool is reused for initial
  cofactors; grouped cofactors are copied before cube restriction so forming a
  new candidate cannot mutate an earlier candidate's truth table.
- Preserved the source algorithm's operand ordering. It visits each newly
  found node as the left operand and its earlier stored match as the right
  operand, so two-input examples are emitted as `(ba)` and `[ba]`.
- Replaced mutable C output buffers with `(expression, status)` results.
  A failed decomposition returns an empty expression and `-1`; constants and
  complete decompositions return status zero.
- Replaced the shared global match-index array with a per-call slice. This
  avoids races when callers parse expressions concurrently.
- Added strict delimiter and malformed-expression checks to the expression
  helpers.
- Updated the DecType 6 call site in `base/abci/abcDec.go` to use the slice API
  and safely view its existing C-style truth pointers. That package remains
  blocked by unrelated C-derived syntax in `base/abci/abc.go`; no package-wide
  green status is claimed for it.

## Tests

`go test ./opt/dsc`, `go test -race ./opt/dsc`, and `go vet ./opt/dsc` pass.
Tests cover constants, single-variable and negated functions, AND and XOR,
pool reuse, all 256 three-variable truth tables (validating every returned
expression against its input), nested gate counts, and malformed delimiters.
