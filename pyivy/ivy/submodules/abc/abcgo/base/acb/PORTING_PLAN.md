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

### Behavioral review, first tranche

Added exhaustive CNF truth checks through three variables; graph link,
traversal and path checks; all two-input duplicate/constant/buffer cleanup
cases; every supported fanin/fanout decomposition with shared inputs; seeded
area/delay optimization equivalence checks; small/constant optimizer cases;
Verilog round-trip coverage; and a single-target ECO support/function test.

Fixes exposed by these tests:

- Per-object storage used allocation capacity as its logical length, breaking
  CNF indexing and later growth. Logical lengths now follow object count.
- Duplicate-input removal updated the wrong fanout list after shifting fanins
  (also a C source defect). Save the removed edge and update its source's list.
- Constant CNF used variable zero when redundant fanins were present. Constrain
  the output variable at `nVars`, as for nonconstant functions.
- Full timing refresh retained levels, path counts and queue entries from
  removed cones. Clear those before the existing full recomputation.
- ECO prime-cube generation had lost C's logical negation, producing `00`
  instead of `11` for an AND patch. Restored the source polarity.
- Supporting package repairs: BSAT reset clears its facade's contradiction
  flag; the vector queue returns the C empty sentinel and permits reattaching
  the same stable priority slice header.

The end-to-end ECO mapped-file test and the original Ivy prover bridge live in
`base/abci` because they require that package's command registration. They are
not yet verified: ABCI still has independent compile errors beginning in
`abc.go`. This is a remaining integration requirement, not a passing test.

### Decomposition and SAT review

The full package now passes at **56.6%** statement coverage. Added exhaustive
SAT function extraction, prime-cube patch generation, candidate validation and
conditional-buffer polarity checks; ABC/SOP round trips; graph metadata and box
output tests; and seeded Ashenhurst/push runs checked on all input assignments.

Additional fixes:

- Use mutable slices for multiword support masks; passing an array through an
  interface copied it and either lost writes or panicked beyond one word.
- Snapshot the SAT model before adding Ashenhurst constraints; Go correctly
  invalidates public model access when the formula changes.
- Include constant roots in full timing traversal, and support-minimize nodes
  after pushing logic before classifying them as buffers/inverters.
- Remove redundant constant LUT fanins when returning an ABC SOP network.
- Count a box's last output and remember its kind before deleting its outputs.
- Correct the conditional-buffer encoding for complemented buffers.

**Confirmed original C defect:** the native executable's `mfse -d` produces a
non-equivalent result for `test_vectors/ashen_unreachable.blif`; native `cec`
provides a counterexample. The Go port keeps the original support-search and
construction algorithms but SAT-checks a proposed two-node rewrite before
accepting it. Rejected candidates leave the original circuit and links intact.
This is an intentional correctness repair, not an alternative optimizer.
