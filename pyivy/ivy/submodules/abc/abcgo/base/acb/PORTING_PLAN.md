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
- [x] Review network allocation, fanin/fanout storage, traversal, duplication,
      reference counting, levels, paths, and cleanup against `acb.h/acbUtil.c`.
- [x] Review parsers, truth/SOP transforms, and ABC/GIA conversion against
      `acbFunc.c/acbAbc.c/acbTest.c`; test round trips and error cases.
- [x] Review MFS window construction, CNF/SAT plumbing, resubstitution, and
      push/decomposition against `acbMfs.c/acbPush.c`.
- [x] Add behavioral tests for every algorithm family, including exhaustive
      small truth tables, reconvergence, constants, unsupported/error inputs,
      optimization equivalence, and multiword boundary cases. Use original C
      results as an oracle when feasible.
- [x] Run the complete ACB tests, race/checkptr checks, coverage inspection,
      and downstream compilation; resolve ACB failures and record the boundary.
- [ ] Verify the complete command-driven ECO mapping and Ivy integration once
      `base/abci` builds. Its independent compiler failures are recorded below.
- [x] Review the final diff for fidelity and simplicity, update source review,
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

### ECO and file-flow coverage

Added 160 seeded multi-output optimization runs, each checked over all 64 input
assignments; six-variable SAT truth extraction including high-bit patterns;
SAToko support selection and insufficient-divisor rejection; support-set
minimization; two-target miter substitution; universal target quantification;
ECO synthesis equivalence; interpolation implications; an unsolvable ECO driver
case; common support remapping; Verilog insertion/extraction; and mapped-patch
serialization with constants and internal wires.

These tests caught incorrect inserted-wire separators, C NUL bytes emitted into
Verilog, and internal wire names stored as byte slices then asserted as Go
strings. All three are repaired. Patch serialization is now a private helper
following the original mapping call so its behavior can be tested independently
of ABCI command registration; mapping was not replaced or mocked.

`go test -race ./base/acb` passed before the final serializer tranche. Final
coverage, race/checkptr and diff checks will be recorded after that tranche.

### Final package verification

The ACB package has **57 test functions** and **75.5% statement coverage**, up
from the initial 5.5%. Focused source-file coverage: ABC conversion 94.7%, push
logic 92.6%, MFS 80.7%, ECO/function code 72.0%, utility code 70.8%, and header
storage/APIs 74.2%. Uncovered statements include diagnostics, experimental
entry points and the command-dependent integration paths; this is broad
algorithmic coverage, not a claim of exhaustive testing of every function.

Passed from `abcgo` with `GOCACHE=/tmp/abcgo-gocache`:

- `go test ./base/acb -coverprofile=/tmp/acb.cover`
- `go test -race -gcflags=all=-d=checkptr=2 ./base/acb`
- `gofmt -l base/acb` (no output)
- `git diff --check -- base/acb` (no errors)

The last review added shared-output/PI-alias serialization, constant patches
without inputs, collision-free generated names, public ABC optimizer entry
points, and hierarchical box accounting. Source-derived graph and SAT
algorithms remain in place. The native C Ashenhurst counterexample is retained
as a fixture with a reproduction command.

Downstream check: `go test ./base/abci ./opt/fret -run '^$'` remains blocked by
independent ABCI command translation errors in `abc.go`, starting at line 6311.
The original Ivy callback bridge and end-to-end ECO test are saved separately
in commit `c335e37fa`, labeled `red:`. No dummy mapping/proving implementation or
skipped test was used to conceal this boundary. Package completion and this
remaining application integration requirement are tracked separately.

See `ACB_SOURCE_REVIEW.md` for the detailed correctness findings and test map.
