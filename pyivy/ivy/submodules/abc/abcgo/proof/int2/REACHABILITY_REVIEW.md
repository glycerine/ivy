# INT2 Reachability Review

## Review scope

Reviewed `int2Reach.go`, the `Int2_ManPerformInterpolation` entry point, and
the small transition-system fixtures in `int2_reach_test.go` separately from
implementation.

## Findings

- The checked-in C driver and its image/preimage path have unresolved local
  names and inconsistent parameter use. The Go entry point therefore uses an
  exact finite-state fallback rather than claiming an interpolant proof.
- A proof result is returned only after BFS has exhausted every state reachable
  from the all-zero register state. A reachable PO that evaluates true returns
  the counterexample status. Reaching a configured depth boundary with a new
  successor, exceeding a resource cap, or reaching the time limit returns
  undecided.
- GIA AND evaluation walks object IDs in increasing order and applies each
  stored fanin complement bit, matching the manager's topological fanin
  representation. PIs and ROs are assigned before evaluating nodes; POs and
  RIs are then read from the same valuation.
- Temporary values and next-state buffers are reused. A newly discovered
  state is copied before the scratch next-state buffer is reused.

## Limits

The current exact fallback accepts at most 16 primary inputs, 262,144 distinct
states, and 4,194,304 state/input evaluations. Larger problems return undecided;
they are not approximated. A zero frame limit means continue until closure,
subject to the resource caps.

## Verification status

Regression tests cover a safe two-state cycle, a reachable counterexample one
transition after initialization, and a three-state cycle that is undecided at
a shallow depth bound but proven when explored to closure. They are not yet
executable in the repository because the imported GIA package has parser
errors in its draft files. `gofmt` parses the changed INT2 files successfully.
