# ACB source review

## Result

All production Go files in `base/acb` compile together. The port retains the
source graph, CNF/SAT, support-search, decomposition, interpolation and ECO
algorithms. No files were excluded to pass the compiler and no replacement
solver or approximate optimizer was introduced. The inactive command handlers
in `acbCom.c` are inside the original C `#if 0`; they are not missing active
implementations. The other source implementation functions have Go counterparts.

Review compared object storage and mutations against `acb.h`, traversal and
levels against `acbUtil.c`, truth transforms and patch flow against `acbFunc.c`,
ABC conversion against `acbAbc.c`, dual-rail simulation against `acbTest.c`, and
window/SAT/decomposition operations against `acbMfs.c` and `acbPush.c`.

## Repairs verified by regressions

- Object vectors now use logical object count rather than reserved capacity.
  This preserves object IDs when arrays grow and keeps per-object CNF entries
  aligned. Box output enumeration includes the final object; deletion remembers
  the box's kind before clearing it.
- Fanin removal updates the removed source's fanouts. Duplicate input removal,
  constant propagation, buffer/inverter elimination and all supported push
  decomposition kinds preserve truth and exact edge multiplicities.
- Full timing recomputation clears stale levels, paths and queue entries.
  Constants are traversal roots, including cones disconnected from all PIs.
  Pushing logic minimizes support before classifying the residual node.
- CNF constants constrain the output variable even with redundant fanins.
  Conditional buffers encode both polarities correctly. ACB uses the owning
  packages' packed CNF and SAT APIs rather than private layout guesses.
- Multiword masks are mutable slices. The old array-through-interface bridge
  copied write targets. High-bit and multiword tests cover this failure.
- SAT model values are saved before mutating the formula. The BSAT restart
  bridge clears the formula and facade contradiction state, matching C restart.
- An UNSAT one-copy observability window means the pivot cannot affect an
  output. Construction now reports this status and the optimizer explicitly
  replaces the unobservable pivot with zero. This removes the original C's
  misleading error/success handling without changing the optimization result.
  CNF restoration is checked across all clauses for one, two and six copies,
  including contradictions and empty formulas/clauses. An empty clause no
  longer panics while reading a nonexistent first literal.
- A zero budget for additional weighted-support improvement explicitly retains
  the initial feasible support, rather than reporting an immediate timeout.
  This preserves the source behavior and is separate from APIs where zero
  means unlimited search.
- Prime-cube SOP extraction restores C's negated literal polarity; AND patches
  were previously turned into NOR patches. Six-variable truth extraction tests
  include bit 63 and asymmetric patterns.
- Constant SOP output has no redundant ABC fanins. ABC import/export and public
  optimization entry points are checked by independent evaluation.
- Verilog output omits C NUL terminators and uses correct wire separators.
  Generated names are strings, avoid existing/reserved names, and preserve
  output aliases and shared drivers. Constant-only patches omit an empty input
  declaration. Serialization remains downstream of the original mapper; its
  private formatting helpers allow direct testing without mocking mapping.

Some repairs also correct defects present in C (box deletion, duplicate-input
fanout bookkeeping, constant timing assumptions and serializer alias handling).
They are explicit correctness changes rather than silent algorithm substitutions.

## Original C Ashenhurst defect

`test_vectors/ashen_unreachable.blif` reproduces an invalid rewrite in the
repository's native executable. `mfse -d` selects intermediate truth `0x4C` and
output truth `0xB`; native `cec` reports NOT EQUIVALENT and a concrete input
counterexample. See the fixture README for the exact command.

The source support search and construction remain intact. Before installing a
two-node rewrite, Go checks that its output cannot disagree with the original
window relation. SAT counterexamples reject the candidate and remove its unused
intermediate node; the original circuit remains intact. Enumeration blocking
clauses are disabled for this check. Separate exhaustive tests verify that the
check accepts equivalent functions and rejects inequivalent ones.

## Test coverage

60 test functions cover:

- Every Boolean function through three inputs for CNF and SAT extraction;
  all binary functions for SOP patch enumeration and ABC round trips.
- All source push decomposition kinds, duplicate inputs, constants, shared
  inputs, reconvergence, graph mutation, MFFC reference restoration, metadata,
  hierarchical box accounting, and two independent name maps.
- Area and delay optimization, Ashenhurst and push drivers, plus 160 seeded
  multi-output optimizations evaluated exhaustively on all 64 input patterns.
- Weighted/support minimization, cross-word masks, BSAT and SAToko support
  selection, insufficient divisors, interpolation implications, universal
  target quantification, two-target miter substitution, and an unsolvable ECO.
- Window construction status and complete CNF restoration after immediate or
  propagated contradictions; unobservable-pivot replacement with exhaustive
  output-equivalence checks; and zero-budget support feasibility/ownership.
- Verilog parsing, insertion/extraction and round trips; input ordering;
  counterexample files; patch support remapping; mapped-record serialization,
  internal wires, constants, output aliases and file bytes.

Coverage measures exercised statements, not a proof of correctness. Remaining
uncovered code includes diagnostic/experimental drivers and the integration
paths below; the tests do not claim exhaustive coverage of all large circuits.

## Application integration boundary

`Acb_NtkSolve` uses the original Ivy proving pipeline through a callback installed
by `base/abci/acb_prover.go`. This removes the Go import cycle without changing
the solver. Applications using that entry point must install the ABCI integration.
A missing provider fails explicitly instead of returning a fabricated result.

Full command-driven ECO mapping and recursive interpolation synthesis also
require ABCI's command/collapse registration, as in the source architecture.
`base/abci/acb_integration_test.go` records the end-to-end mapped-file test; it is
not skipped or replaced with a dummy mapper. It remains unverified because ABCI
has independent compile errors starting at `abc.go:6311` (`Abc_NtkGetFaninMax`,
`Abc_NtkToSop`, `Abc_NtkOptPush`, and other untranslated command code). `opt/fret`
is blocked by the same dependency. This is separate from the passing ACB package.

## Final verification

With `GOCACHE=/tmp/abcgo-gocache`, `go test ./base/acb -coverprofile=/tmp/acb.cover`
passes with **75.8% statement coverage**. The complete suite also passes
`go test -race -gcflags=all=-d=checkptr=2 ./base/acb`. Formatting and diff checks
are clean. No tests were run on `experiments/`.

The complete `go test -v -count=1` audit initially found 90 misleading SAT
errors (45 unobservable windows reported twice) and one zero-budget timeout.
After the repairs neither appears. The two no-solution diagnostics and the
missing-file diagnostic are asserted by deliberate rejection tests and remain
visible under `Expected rejection` labels in verbose output.
