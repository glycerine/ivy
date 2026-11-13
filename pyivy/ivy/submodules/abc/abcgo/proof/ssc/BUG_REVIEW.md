# Independent correctness review

Completed after the first functional tests. Checked constraint polarity (all
violations false), both SAT query directions, phase correction, constant literals,
UNKNOWN retention, solver recycling, and source-ID versus rebuilt-literal maps.
Checked that witness refinement separates the failed candidate and never merges
classes on simulation alone. Proved representative links and phases are published
consistently, including conditional constant-one equivalence.

Review/tests caught and corrected care extraction assigning register metadata to
a combinational care graph, the source's trivial-constraint Append behavior, and
missing pivot phases/class links in published representative metadata. Replay-free
word simulation handles zero-input graphs without source division assumptions.
Gate simulation and CNF handle explicit native XOR/MUX/buffer nodes. Construction
validates both source networks using the native circuit-model validator.

Containment retains a sufficient output union as outputs are removed. Ring
constraints are temporary and preserve union rather than each individual output
outside its care set. Concurrent independent-manager tests exercise synchronized shared random,
and independent class and solver state under the race detector. Simultaneous mutation of one
source GIA remains outside the graph API's concurrency contract.

Normal/race focused suites and vet pass with the real native GIA dependency set.
The C oracle actually ran successfully during review; optional skips are visible.
The earlier GIA dependency blocker has since been resolved. The full package
now builds and its race tests and vet pass, including a rerun of the original C
oracle. Additional focused tests cover care masks, witnesses, phase-sensitive
class refinement, and packed simulation primitives.

A follow-up source-branch pass exercises a full 64-pattern counterexample batch
and the final partial batch with 66 independent rare output cones, replaying
all one-block assignments. A separate containment regression checks C's
left-to-right tie behavior: among equivalent output implementations, the last
one survives. Both are source-derived tests; no new implementation mismatch was
found in these paths.
