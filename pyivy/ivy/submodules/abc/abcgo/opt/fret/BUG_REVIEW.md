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
- Checked latch bubble handling, flow-block marking, and latch-box restoration
  against their C counterparts.

## Verification still pending

`gofmt` and `git diff --check` pass for the FRET files. Package tests, race
tests, and vet cannot yet reach FRET: the current module build stops on raw C
syntax in `base/abc/abcLatch.go` and `base/abci/abc.go`. The SAT API is also
being ported in `base/abci/abcSat.go`; once its package parses, rerun all three
gates and review the resulting compiler diagnostics before treating this port
as complete.
