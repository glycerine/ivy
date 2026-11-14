# Completing the ABCI port

## Contract

Complete `base/abci` against `src/base/abci`. Preserve the C algorithms,
command options, return conventions, graph ownership, literal polarity and
observable results. Use the existing Go implementations in owning packages;
do not substitute algorithms, hide unfinished files behind build exclusions,
or add success stubs. Correct source defects only with explicit evidence and
regression tests. Fixtures use `test_vectors`, never `testdata`.

## Initial assessment (2026-11-14)

The package has 119 production Go files (95,732 lines) and 95 test files
containing 264 test functions. Much prior porting and local review exists,
but the whole package has not yet built. `abc.go` alone contains 47,031 lines
of command dispatch and adapters. Older review documents describe historical
compile frontiers, not current passing results.

`go test ./base/abci -run '^$'` with compiler `-e` reports 3,657 diagnostic
lines. Most are concentrated in `abc.go` (2,665), `abcDar.go` (403),
`abcMiter.go` (307), and `abcQuant.go` (109). Many missing names already exist
in owning packages; the remaining errors include C integer conditions,
pointer/slice mismatches, private fields, untranslated iteration and memory
operations, and invalid option/error paths. Three existing test compilation
errors also prevent exercising previous tests.

## Milestones

- [x] Inventory source, tests, review notes and all compiler diagnostics.
- [x] Resolve API ownership and compatibility bridges against actual Go types.
- [x] Complete command, DAR, miter and quantification translations; compile
      every production file and all existing tests together.
- [x] Run the complete suite; investigate every failure and unexpected
      diagnostic, checking the corresponding C before changing behavior.
- [x] Review algorithms and adapters independently of compilation: ownership,
      complemented edges, sequential interfaces, truth widths, SAT status,
      command parameters and failure handling.
- [x] Add comprehensive behavioral tests for risky and uncovered paths:
      independent exhaustive small-network evaluation, sequential examples,
      command parsing/dispatch, conversion round trips, invalid inputs,
      optimization/verification, and ACB/Ivy/ECO integration.
- [x] Run full verbose, race/checkptr and coverage checks; inspect coverage by
      algorithm family and record gaps honestly.
- [x] Verify affected downstream packages, update source review and progress,
      and commit coherent changes on the current branch.

## Working sequence and verification

First restore the complete compilation boundary. Group missing APIs by owner
and use typed adapters only where representations differ. Then complete the
remaining C expressions and control flow against the source. Run all existing
tests before adding coverage driven by observed risks and coverage reports.
Tests must verify behavior, rather than merely mirror implementation details.

Use `GOCACHE=/tmp/abcgo-gocache` from `abcgo`. The package build and test gates
always include all of `base/abci`; selected-file tests are not evidence that
the package builds. Do not run `experiments/`. Save compact coherent changes
at least about every ten minutes; commits made while the package still fails
are labeled `red:`. Record commands, outcomes and findings below as work
progresses.

## Progress

Initial diagnostic inventory saved to `/tmp/abci-build-initial.log` for this
session. No passing-package claim is made at this stage.

### API and parameter integration

Resolved missing names to the existing owning Go packages. The command layer
also exposed a broad API gap: C-visible solver/mapper parameter fields were
private in their Go types. Exported the required fields and updated their
uses by Go type identity, preserving a single parameter record and avoiding
shadow configuration state. Full tests pass for `aig/gia`, `aig/ivy`,
`map/amap`, `map/if`, `map/mpm`, `proof/abs`, `proof/acec`, `proof/cec`,
`proof/fra`, `proof/pdr`, `proof/ssw`, and `sat/bmc` after these changes.

Restored 368 parameter-qualified flag references dropped by prior ternary
translation, checking the original command bodies. Converted 411 C truth
tests according to their Go operand types. ABCI remains red while command
arguments, representation conversions and graph algorithms are completed.

### Graph and buffer representation repairs

Completed the miter and quantification compile repairs using native slices
and explicit complemented edges. Restored latch iteration and preserved the
linked-list traversal used by reverse DFS. Removed pointer indexing from
truth, switching, and counterexample buffers; fixed affected existing tests.
These changes remain unvalidated at runtime while the whole package is red.
Compiler diagnostics now concentrate in the command and DAR integration
files. The next gate is still a build of every file together, followed by
full behavioral tests and inspection of diagnostics.

### Whole-package compilation and first runtime pass

Every production file and existing test now compiles, with vet enabled.
Completed native DAR/CNF/GIA adapters, SBD mapping and timing translation,
phase mapping and ESOP bridges, and command argument/result conversions.
The full verbose suite is now being exercised. Its first run found a local
DSD worklist that requeued nodes even with recursive decomposition disabled;
it also omitted minimum-base reduction of recursive MUX cofactors.

Supporting tests exposed missing default names and stale/absent CopyEdge
metadata during network duplication. A DAR test confused output objects
with their driver literals; a retiming test expected logic-only timing edges
but C explicitly retains latch and BI terminals at the same level. These
expectations were corrected against source, without suppressing diagnostics.
Runtime verification and comprehensive coverage remain in progress.

### Runtime repairs and regression coverage

The entire package now builds and passes its full suite, including vet. No
production files were excluded to obtain this result. The existing
`abc_use_cudd` switch selects the upstream optional proving fallback; both
build configurations have been tested.

The runtime review found and repaired the following:

- DSD worklist nontermination: nonrecursive conversion requeued generated
  nodes, and recursive MUX cofactors were not reduced to their minimum base.
  Local and global BDD functions now retain their owned references.
- Strashing wrote temporary signals to CopyData, while the output conversion
  read CopyEdge. Local Hop input polarity was also lost.
- Quantification used Data as scratch storage, overwriting Go node functions
  and latch metadata. Cofactors now use local maps of explicit edges.
- Reachability discarded a constant-one image before incorporating it into
  the reached set. This is also a defect in the C early-exit path; the Go fix
  is checked against independent breadth-first exploration of all 16
  one-bit transition relations.
- Resubstitution overwrote and cleared function Data, and its replacement
  graphs stored raw objects where the Go decomposition adapter requires
  edges. Scratch simulation words now use CopyData. Tests preserve truth
  tables and latch initialization with and without observability don't-cares.
- Mapped-GIA conversion discarded LUT boundaries and assigned SOP functions
  to an AIG-function network. Restored the source's LUT/Hop, explicit XOR/MUX,
  and latch-enable conversion paths.
- Phase mapping constructed separate cell handles for the network and the
  optimizer, silently preventing matches. Canonical handles now allow area
  reduction, checked without changing the mapped network's truth table.
- IF mapping confused the interned truth-table polarity with the cut's
  separate `fCompl` bit. This complemented functions twice. Exhaustive
  multi-output examples now pass for Hop, BDD, SOP, CNF and MV modes.
- Native ABC AND creation omitted zero-input simulation phase and level
  metadata, causing Ivy to report invalid zero-valued witnesses. The owning
  package now maintains both fields, with direct tests of shared AIG APIs.
- Ivy's early SAT exits did not return the replacement network after deleting
  a logic input. This defect is present in C too; every return now updates
  the caller's network pointer. Zero-input clean-model creation also clears
  stale witnesses.
- Fixed Espresso literal decoding, truth-measurement counts for fewer than
  five inputs, source-compatible default timing averages, and error/ownership
  handling in ESOP and clock-gating adapters.

The ECO regression now loads the real command aliases and cell/LUT libraries,
runs the original synthesis script, and fails on internal error diagnostics.
It previously could pass despite command failures inside that script.
Negative command/file tests explicitly check expected failure behavior.

Supporting package changes include exposing C-visible parameter fields by Go
type identity, sharing the existing miter SAT engine through base/abc to
avoid an import cycle, repairing copied edge/name metadata, and preserving
Hop/SOP function types when complementing or constructing simple nodes.
These are integrations of existing algorithms, not replacement solvers.

### Verification status (2026-11-14)

From `abcgo`, with `GOCACHE=/tmp/abcgo-gocache`:

- Full `go test -v ./base/abci -count=1` passes. Tests include the actual
  command-driven ECO flow and help handling for every registered command.
- Full `go test -race ./base/abci -count=1` passes.
- Full `go test -gcflags=all=-d=checkptr=2 ./base/abci -count=1` passes.
- Full `go test -tags abc_use_cudd ./base/abci -count=1` passes.
- Full tests pass for the directly affected `base/abc`, `base/acb`,
  `base/cmd`, `opt/fret`, `map/if`, `aig/ivy`, and `sat/cnf` packages.
  The earlier API integration checks are recorded above.

The expanded behavioral checks include all 256 three-input Boolean functions
under existential/universal quantification; all 16 two-input functions through
SAT/miter/Ivy/ACB proving and QBF synthesis; bounded sequential checking and
PDR with counterexample replay; sequential framing and transition relations;
DAR optimization equivalence; SBD LUT polarity/constant/duplicate handling and
box timing ownership; mapped-GIA conversion; DSD; resubstitution; symmetry,
unateness and autosymmetry; timing; phase mapping; and ESOP file round trips.

The suite now has 291 top-level test functions (initially 264). The final full
coverage run measured **37.5% statement coverage**, with 860 passing tests
including subtests. That broad suite does
not establish exhaustive correctness of every command or option. Large parts
of `abc.go` and `abcDar.go` dispatch specialized engines; help-path coverage
is not equivalent to running those engines on real designs.

### Remaining validation opportunities

These are coverage/validation gaps, not remaining compilation work:

- Large real designs, benchmark performance and differential C/Go results.
- Rare mapper options, choice-heavy multi-output designs, timing libraries,
  and multiple interacting white/black boxes.
- Resource exhaustion deep in rewriting/fraiging/proving, and every optional
  BDD fallback/witness path; a passing tagged build does not exercise every
  branch of that fallback.
- Specialized interpolation, abstraction, retiming and external-tool command
  combinations; diagnostic and reporting branches outside the added checks.

Historical per-file review notes retain the earlier compile frontier for
traceability; this plan and `PORTING_REVIEW.md` supersede those old status
claims. Continue adding tests based on concrete use cases and discrepancies.
