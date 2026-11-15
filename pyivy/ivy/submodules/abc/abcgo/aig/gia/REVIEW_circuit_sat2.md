# Watched-clause circuit SAT (giaCSat2.c)

`giaCSat2.go` retains CBS2's circuit propagation, learned clauses and watched
literal lists, relative queue frontiers, reason expansion/resolution, backjumping,
fanout propagation and strict resource limits. The default search still uses
source's highest-ID decision policy; the source activity bump is disabled.

Native comparisons cover 24,384 single-root queries in CI and all-object witness
formats, across both propagation modes and resource budgets. Exact verdicts,
ordered models, search/propagation/failure/clause-conflict counters agree with C;
all three outcomes occur, up to 23 backtracks and frontier eight. Every cube is
independently checked over all five-input assignments. Another 3,600 native miter
properties cover logical CI/CO permutations, budgets, stores and unchanged phase
vectors. Source CBS2 disables phase initialization, unlike CBS.

Direct native snapshots exercise watch relocation, unit implication, conflict
learning, forced clause-array growth, reason storage and mark cleanup. Prepared
fanout structure and ordered models match C; solves preserve caller-owned links,
while automatically prepared links are removed on return. Native decision-helper
tests check frontier bounds, ordering/ties and reference/level/activity scores.
Fake-clock statistics, signed cube/clause reports, fanout reports and verbose
propagation counters match source; nondeterministic elapsed values are excluded
from the public verbose comparison while their formatting is validated.

Regression-backed fixes restore miter constants before budget checks, source
reporting, propagation counters and idle sentinels. Both roots and the static
ordinary-AND graph are validated before modifying solver state. Nested queue
snapshots, partial cancellation, CI/all-object encoding, strict limits, activity
cleanup and repeated/active stop have dedicated tests. Solver allocation/stop
preserve graph-owned marks and values.

The Go graph API also accepts raw constant fanins, buffers and repeated/opposite
fanins that usual C hashing simplifies away. Regressions exposed false SAT,
double assignment and stale constant fanout links. Fanin predicates now preserve
fixed values, shared assignments are rechecked, and constant fanout heads are
cleared. A negative reason sentinel distinguishes constant-only implications
from decisions and positive clause handles. Such unconditional implications can
expose a conflict independent of the current branch: reason analysis continues
at the highest remaining antecedent level, following the existing resolution
rule. No alternate solver is introduced.

Independently exhaustive extension tests cover 35,328 raw-gate queries and all
57,832 selected root pairs on the harder five-input graphs. These include
contradictory and repeated roots and the source API's zero second-root sentinel;
every verdict, nonempty witness completion and scratch cleanup is checked.
The original C can assert for multiple same-level paired-root decisions, so its
native baseline deliberately uses one-root queries; Go paired queries are
validated independently rather than claiming C agreement for those cases.

Compactness: restore source reusable mark scratch, use a fixed two-root array,
and reacquire older clause entries by index while storage grows rather than
copying clauses during every analysis. Hard UNSAT queries now allocate zero bytes
in both modes, versus 61/25 allocations previously. Final measurements after the
raw-gate fixes are approximately 5.7/4.3 microseconds; timings are not guarantees.
Logical queue heads/tails remain necessary for suspended frontier storage.

The final focused suite passes in 13.577s. Full uncached module race tests pass
all 130 tested packages (three command packages have no tests), and GIA vet
passes. Logs are /tmp/abcgo-cbs2-final-module-race.log and
/tmp/abcgo-cbs2-final-vet.log; SEQUENTIAL_AUDIT.md records the checkpoints. This note covers CBS2; GIA's other families and the full
131-package sequential audit remain open.
