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

The 2026-11-15 sequential audit adds direct actual-C comparisons: 30,432
verdict/counter/ordered-model cases across all decision heuristics and strict
backtrack/frontier limits, including hard reconvergent/parity graphs. The sweep
includes SAT, UNSAT and undecided outcomes, up to 54 backtracks and frontier size
eight. Every partial cube is checked against every completing input assignment.
Another 3,600 public-miter properties compare exact statuses and store records
across CI/CO permutations; fake-clock statistics match C byte-for-byte.

Regression-backed repairs restore source constant-output handling before miter
budget checks, exclusion of constants from solver-call statistics, the original
statistics/constant-one diagnostic, and all four idle queue-head/tail checks.
Lifecycle/guard tests cover active Stop, nested saved queues through growth,
invalid roots/restoration, original allocation/setup, and the Go buffer/constant-
fanin extensions. Stop now clears the complete manager after cancelling its trail.

Valid GIA managers maintain typed-gate counts, so unsupported XOR/MUX checks use
those counts rather than rescanning every object on every query. A 10,000-gate
leaf-query benchmark drops from approximately 21.6 microseconds to 30 nanoseconds
with zero allocations; those figures precede the final queue-head guard checks.
The important improvement is constant-time validation, not a timing guarantee.

The final focused suite passes in 7.486s. Full uncached module race tests pass all
130 tested packages (three command packages have no tests), including GIA in
329.803s; GIA vet passes. Logs are /tmp/abcgo-cbs0-final-focused.log,
/tmp/abcgo-cbs0-final-module-race.log and /tmp/abcgo-cbs0-final-vet.log. No product
sources were hidden. The former isolated bridge-build limitation no longer
represents the current module build. The rest of the GIA sequential audit remains
open; this note covers the chronological Cbs0 engine only.
