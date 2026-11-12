# Graph SAT encoding review

## Readability and compactness pass

`GraphSAT` encodes each AIG node with one BSAT variable and uses the public
clause encoder for constants, phased ANDs, and phased wires. XOR nodes use the
four clauses for an equivalence to the exclusive-or of their two fanins. The
node-to-variable table includes deleted graph slots, preserving stable AIG
IDs, and starts after any variables already present in the solver. Constraint
queries use assumptions, so repeated pattern searches reuse one encoding
without permanently adding query clauses.

## Independent bug pass

Checked complemented edges on both AND and XOR fanins, complemented wire
outputs, constant-one polarity, stable IDs with deleted slots, coexistence
with pre-existing solver variables, repeated SAT assumptions, and both
mismatch phases in equivalence checking. Regression tests compare constrained
SAT models against graph evaluation and exercise equivalent and inequivalent
functions.

## Remaining boundary

The old `Res_SatProveUnsat` flow duplicates clauses to make two interpolation
partitions. The C source calls `Sat_SolverDoubleClauses`, whose Go port is
intentionally disabled with a panic because the original C implementation is
behind `assert(0)`. `GraphSAT` does not emulate this operation or claim to
produce an interpolant. A sound partition duplication contract must be
specified and tested before wiring the old proof path to interpolation.

The translated `resSat.go` still contains unfinished C-shaped code and blocks
the full legacy `opt/res` package build. The new graph/SAT APIs are independently
compilable and tested; they do not make the historical `Abc_Ntk_t` entrypoints
ready for use.
