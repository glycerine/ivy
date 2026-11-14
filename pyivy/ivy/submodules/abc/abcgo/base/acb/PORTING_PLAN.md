# Completing the ACB port

## Contract

Port the algorithms in `src/base/acb` faithfully to Go. Preserve the C
algorithms, ordering, literal polarity, truth-table widths, and observable
results. Do not substitute algorithms, disable files, or add success stubs to
make the compiler pass. Keep useful existing translations and replace broken
sections where a direct source port is clearer.

## Initial assessment (2026-11-14)

The package contains translations of all source files, plus small bridge tests,
but has never passed its complete package test gate. Remaining compilation
errors include cross-package proving dependencies, pointer/slice mismatches,
and incomplete utility conversions. Earlier integration introduced an import
cycle between ACB and ABCI; this must be resolved without changing the solver
used by the C source. Existing tests cover only a small part of the behavior.

## Milestones

- [x] Inventory the C and Go files, existing tests, and current compiler errors.
- [x] Complete compilation of every ACB production file and existing test.
- [ ] Review network allocation, fanin/fanout storage, traversal, duplication,
      reference counting, levels, paths, and cleanup against `acb.h/acbUtil.c`.
- [ ] Review parsers, truth/SOP transforms, and ABC/GIA conversion against
      `acbFunc.c/acbAbc.c/acbTest.c`; test round trips and error cases.
- [ ] Review MFS window construction, CNF/SAT plumbing, resubstitution, and
      push/decomposition against `acbMfs.c/acbPush.c`.
- [ ] Add behavioral tests for every algorithm family, including exhaustive
      small truth tables, reconvergence, constants, unsupported/error inputs,
      optimization equivalence, and multiword boundary cases. Use original C
      results as an oracle when feasible.
- [ ] Run the complete ACB tests, race/checkptr checks, coverage inspection,
      and downstream compilation; resolve failures and record actual coverage.
- [ ] Review the final diff for fidelity and simplicity, update source review,
      and commit coherent changes with messages under 60 characters.

## Verification and progress

Run from `abcgo` using a writable cache, for example:
`GOCACHE=/tmp/abcgo-gocache go test ./base/acb`.
Fixtures belong in `test_vectors`, never `testdata`. Passing compilation is
the first milestone, not evidence of semantic equivalence or complete coverage.
Record test commands, failures fixed, and remaining gaps here as work proceeds.

### Compilation milestone

`go test ./base/acb -coverprofile=/tmp/acb.cover` now passes. Baseline coverage
is only **5.5%**; comprehensive coverage and semantic review remain outstanding.
The ABCI proving call now crosses an explicit callback installed by ABCI;
`Acb_NtkSolve` requires that integration and never substitutes another solver.
The original proving wrapper is in `base/abci/acb_prover.go`; its downstream
verification is still pending. Fixed missing API bridges, concrete pointer and
slice conversions, CNF data import, and floating-point diagnostic calculations.
Removed an invalid empty-name case from a pre-existing test (the source name
manager requires nonempty names).
