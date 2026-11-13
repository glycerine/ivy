# PDR port review

The package ports the ordinary and incremental PDR engines, SAT and CNF
integration, ternary justification, abstraction refinement, and invariant
operations. `Solve` exposes options and results to Go callers; the original
C-shaped entry points remain available for internal integration.

The package tests include exhaustive one-register transition/property pairs,
and seeded two-register machines checked against explicit reachability. The
latter exercise monolithic and dynamic CNF, both ternary simulators,
abstraction refinement, shortest-trace mode and incremental restarts. Failing
results are checked with the AIG counterexample verifier. An additional 64
two-register, multi-output native-XOR machines with complemented edges are
checked against explicit reachability and their witnesses are replayed. A
further 160 sampled three-register machines exercise proof-obligation reuse,
shifted frame starts, both CNF modes, abstraction, and incremental restarts
against an eight-state reachability oracle. A local ABC executable, when
available, also checks 64 generated two-register machines with matching C and
Go CNF, abstraction, priority, generalization and frame-management options.
A safe-latch fixture compares the invariant PLA cover with ABC. Bridge tests
check the framed proved/failed-output messages and final abort message for
ordinary and incremental runs. Other
tests cover frame limits, SAT conflict limits, expired global and per-output
deadlines, solver reuse, invariant checks, and SAT core handling. The full
package has passed `go test -race ./proof/pdr` and
`go vet ./proof/pdr`.

Remaining review and coverage gaps:

- The two-register cases are sampled, not exhaustive over all two-register
  machines. Larger sequential designs and deeper refinement chains need
  end-to-end comparisons, including runs against the original ABC PDR.
- Resource exhaustion and per-output deadlines have less coverage than normal
  SAT/UNSAT results. Cancellation is tested, but real timeout, conflict-limit,
  and interruption combinations warrant further tests.
- ABC's updated SAT-based justification (`fNewXSim`, command-line `pdr -u`)
  can produce an invalid counterexample because it derives the justification
  in a frame solver containing reachability clauses. This source behavior is
  preserved. One reproducible three-register machine has transition truth
  tables `[60055, 49603, 19452]` and property table `19712`, indexed by
  `input + 2*state`; local ABC reports that its `pdr -u` counterexample fails
  verification, while ordinary `pdr -m` finds a valid failure. The 160-case
  three-register test therefore uses the default simulator. The public Go
  API does not currently expose `fNewXSim`.
- The original `fNewXSim` implementation also assumes a dynamic-CNF reverse
  map; local ABC crashes on the above machine with `pdr -m -u -p`. Go reports
  this option combination as undecided. As in ABC, updated justification with
  abstraction asserts if a derived cube intersects the initial state.
- ABC compiles the body of its alternative simple generalization out with
  `#if 0`; when enabled for a multi-literal cube, its active caller discards
  the reduced cube. The Go port preserves this source behavior, which can lead
  to a nil-cube failure. The public Go API does not expose this option.
- `Pdr_InvCheck_int` follows ABC on inconclusive SAT queries: it stops that
  check loop and returns failures counted so far, which may be zero. A zero
  result therefore does not establish an invariant if the supplied solver was
  interrupted. `Pdr_InvMinimizeLits` also follows ABC by leaving removed
  literals marked `-1` in its input vector.
- As in ABC, a generalization query with a nonzero generalization conflict
  limit treats an inconclusive SAT result as SAT. If a predecessor is also
  requested, a SAT model may be unavailable; the Go model access then fails
  rather than manufacturing one.
- As in the source, `ZPdr_SetIntersection` does not check required literals
  in the unmatched tail after either sorted input ends. Also,
  `Pdr_SetPrintStr` appends no pattern characters when its flop filter is
  nil. Tests record these source behaviors; they are not corrections to ABC.
- As in ABC, stopping a PDR manager stops the graph's fanout index even if
  that index existed before the manager started. A caller that needs fanout
  data afterward must rebuild it.
- Bridge output beyond trivial properties and invariant files beyond the
  safe-latch fixture need broader end-to-end coverage. ABC can interleave
  ordinary progress text with bridge frames unless `fNotVerbose` is set.
  Proof-obligation reuse and shifted starting frames are varied in the sampled
  state-space tests, but those tests do not establish that every branch is hit.
- Successful `Pdr_ManSolve` and `IPdr_ManSolve` runs publish their invariant
  in ABC's shared global frame. The frame synchronizes and copies each update,
  but the most recent publication replaces the previous invariant. Callers
  needing the invariant for a particular run must serialize all publishers and
  read it before the next publication. A race test exercises both direct solve
  entry points on independent graphs; it does not establish safety for shared
  graphs, parameters, other global-frame users, or every solver option.
