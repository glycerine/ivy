# SAIG induction review

## Scope

`Saig_ManInduction` now performs the source's iterative top/bottom support
unrolling with native `aig.Graph` partial duplication, native `cnf.Cnf_Derive`,
and the incremental `bsat.Solver`. It stitches successive latch-state variables,
checks bounded property reachability, and optionally adds pairwise state
uniqueness constraints. A SAT result at the requested last frame returns 0; an
UNSAT induction query returns 1; resource exhaustion or invalid input returns
-1. The successful bounded witness keeps the original compact C layout with
`nRegs-1` register bits and is retained in a dedicated `Graph` induction-model
slot, leaving generic `Graph.Data` unchanged.

## Independent bug pass

- Reviewed the support/partial-duplication ordering against C: each part's
  outputs are the property first, then latch inputs for the bottom latch
  support; the SAT stitch maps those outputs to the prior state's variables.
- Checked property phase for frame zero versus later frames, frame/conflict
  bounds, timeout setup, and the repeated SAT solve after adding uniqueness
  constraints.
- Added SAT/UNSAT bounded-property regressions and a primary-input witness test
  that checks the compact bit layout.
- Added Graph-slot validation tests for the source's `nRegs-1` witness, including
  a rejected full-register layout that must leave the old model intact.
- The uniqueness test exposed a real support bug in `cnf.Cnf_DataAddXorClause`:
  its Go port treated variable IDs as encoded SAT literals. It now translates
  each of the four source clauses through `bsat.Positive`/`Negative`; an
  exhaustive truth table checks all eight XOR assignments, and induction tests
  check that a repeated state is excluded while a distinct state remains SAT.
- The C source's last-frame SAT branch leaves `RetValue` at its initial -1 even
  though `Abc_NtkDarInduction` treats 0 as “not equivalent.” The Go port returns
  0 on this branch, matching the caller's documented result convention and the
  generated witness.

## Compactness and readability pass

- State, top-output, and support data use slices or the existing vector type;
  there is no per-node allocation for the unrolling.
- Partial AIGs and CNFs are released on each completed iteration, and the graph
  stores the optional witness in a typed slot rather than overwriting generic
  metadata.
- State uniqueness logic is isolated in `Saig_ManStatesAreEqual` and
  `Saig_ManAddUniqueness`, with tests covering assignment comparison and XOR
  blocking.

## Verification

Focused source files and the `saigInd`/GLA tests pass normal, race, and vet
checks. `sat/cnf` passes normal, race, and vet checks. `aig/aig` passes normal,
race, and vet checks. Full `go test ./aig/saig` remains blocked by unrelated
syntax errors in `saigIsoFast.go`.
