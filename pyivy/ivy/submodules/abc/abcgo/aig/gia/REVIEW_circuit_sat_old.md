# Original chronological circuit SAT (giaCSatOld.c)

`giaCSatOld.go` ports the actual Cbs0 engine: assignment trail, suspended
justification frontiers, top-down AND propagation, highest/lowest/fanin-fanout
heuristics, chronological two-branch search, source backtrack/frontier limits,
partial CI cube witnesses and per-output miter store encoding. It does not
alias the newer CBS or SAToko engines.

Readability pass: queue size derives from the backing slice; logical head/tail
remain because saved frontiers occupy inactive backing storage. Source counters
and predicates keep their recognizable names. Shared assignment consistency is
centralized, and timing statistics use time.Duration.

Separate bug pass: tested scratch cleanup after SAT/UNSAT/timeout, saved queue
frontier isolation, repeated solver use, constants, complemented roots and
partial witness validity. Fixed stale witness exposure after a timeout, duplicate
reference-table initialization, and direct constant-query handling. Rejects
explicit XOR/MUX encodings, matching the source's ordinary-AND precondition.
Assignments through repeated regular fanins detect contradictions instead of
blindly assigning the same object twice. The source default heuristic priority,
strict `>` limits and backtrack counting are retained.

Dedicated regressions exhaustively check 20 deterministic five-input graphs,
three decision heuristics and every generated root in both phases; all SAT cubes
are independently evaluated and every solve must restore marks. Additional tests
exercise limits, reuse, queue store/restore, constant/simplified fanins, and exact
miter status/store layout.

Focused actual-source race gate passed (7.976s):

    bash abcgo/aig/gia/test_aig_bridge.sh -race \
      aig/gia/giaCTas2.go aig/gia/giaCSatOld.go \
      aig/gia/circuit_sat_old_test.go

The additional Tas2 file is required by a concurrently expanded shared test.
No product sources were hidden. Ordinary NWK still stops in other GIA/SAIG
files; see its REVIEW.md for the integration gate distinction.
