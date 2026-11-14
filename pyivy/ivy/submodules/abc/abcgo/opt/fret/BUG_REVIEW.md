# FRET source and bug review

This is a separate source-comparison pass over the native FRET port. It compares
the timing, initialization, main-loop, and latch-box code with
`src/opt/fret/fretTime.c`, `fretInit.c`, and `fretMain.c`.

## Findings fixed

- Timing DFS order checks cleared temporary marks while still checking later
  neighbors. A valid fanin or fanout could therefore fail the assertion. Both
  checks now clear marks only after checking the complete traversal. The
  regression test covers forward and backward traversal.
- Latch-box repair dropped the complement bit when inserting a BI around a
  complemented latch input. It now transfers the full `Edge`; the regression
  test checks the inserted BI's edge.
- `Abc_FlowRetime_InitState` treated integer flags as Go booleans. It now tests
  the flags against zero.
- Initial-state simulation checked BDD representation before SOP, unlike the C
  dispatch order. It now checks SOP first. The strash simulation also now uses
  the same default input value as the C implementation.
- Init-to-original mapping had discarded constants even though the C code maps
  them to their original constant object. Conflict extraction can select a
  constant, so those mappings are preserved; a regression test covers lookup.
- Backward-init remapping set the latch's generic object data but omitted its
  flow-data mapping. The recursive init-network rebuild reads the latter, so it
  could hit a nil mapping. Both mappings are now set, with a regression test.
- Partial-SAT conflict candidates were ordered as DFS nodes then PIs, although
  the C code rotates the appended PI segment to the front. The order now matches
  C, and each candidate's temporary level index is also written as in C.
- Bias propagation used a distinct visit per object and relative lag. C visits
  each object once; its `lag - isLatch ? 1 : 0` expression also produces a
  zero/one lag toggle for the tracked lag states. The port now mirrors that
  traversal and expression, with tests for the boundary path and lag transition.
- HOP initial-state simulation accepted XOR nodes even though the source only
  handles constants, PIs, POs, and ANDs. Unsupported HOP nodes now follow the
  source assertion path.
- HOP PI initial-state propagation now converts the network's `INIT_CARE`
  meaning (the input is known) into the evaluator's don't-care meaning. The C
  code copies `INIT_CARE` directly into `fMarkB`, then treats `fMarkB` as a
  don't-care bit in `Abc_FlowRetime_EvalHop_rec`; this makes a known HOP input
  appear unknown. The Go evaluator keeps the intended known/unknown behavior,
  covered for known values, DC propagation, controlling zero, and complemented
  outputs in `TestHopInitialStateSimulationTracksCareAndComplement`.
- Named type aliases in `fretime.go` collided with dot imports from the flow
  source files. The flow files now use qualified owning-package APIs, preserving
  the exported aliases without import-scope collisions.
- The `trace2` debug helper from `fretTime.c` was missing. It now follows the
  first fanin on a longest-level path and has a regression test for traversal
  order.
- The minimum-retiming entry point omitted the verbose max-delay setting and
  the `DEBUG_CHECK` network validation enabled by `fretime.h`. Both source
  behaviors are restored; a test covers the rejected-delay report.

## Review coverage

- Compared forward and backward conservative timing propagation and hard-delay
  thresholds with the C implementation.
- Compared exact-constraint traversal direction, latch crossing, level
  propagation, and edge collection with the C implementation. Added tests for
  exact boundaries in both directions.
- Compared SOP, BDD, AIG, strash, and mapped initial-state dispatch and
  simulation with the C implementation. Added SOP and BDD cofactor tests.
- Checked backward-init conflict extraction and partial-SAT polarity against
  the C return convention: SAT is zero at `Abc_NtkMiterSat`; the helper
  `Abc_FlowRetime_PartialSat` returns true for SAT.
- Compared PI/DFS ordering, temporary object levels, latch init mappings, and
  init-bias traversal with `fretInit.c`.
- Checked latch bubble handling, flow-block marking, and latch-box restoration
  against their C counterparts.

## Verification still pending

The Go files parse cleanly under `gofmt`, and `git diff --check` passes for
this directory. `go test ./aig/hop` passes. `go test ./opt/fret` has not yet
reached FRET type-checking: it currently stops on syntax errors in the
transitively imported `base/abci` package (`abcMap.go`, with errors around
lines 424–568 at the latest run). The FRET tests added here remain
unverified until that dependency builds; rerun the complete FRET suite after
the `base/abci` parser errors are resolved.
