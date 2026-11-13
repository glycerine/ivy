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
available, also checks 32 generated two-register machines. Other
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
- Bridge output, invariant file output, and rare options such as proof
  obligation reuse and shifting the starting frame are not covered end to end.
- Successful `Pdr_ManSolve` and `IPdr_ManSolve` runs publish their invariant
  in ABC's shared global frame. The frame synchronizes and copies each update,
  but the most recent publication replaces the previous invariant. Callers
  needing the invariant for a particular run must serialize all publishers and
  read it before the next publication. A race test exercises both direct solve
  entry points on independent graphs; it does not establish safety for shared
  graphs, parameters, other global-frame users, or every solver option.
