# FRA manager test coverage

The package's statement coverage was 85.6% on the current tree, measured with
`go test -coverprofile=/tmp/framgr.cover ./proof/framgr`. Coverage is a guide to
unexercised code, not proof of equivalence to ABC. A source and independent bug/simplicity review found and restored missing
preconditions, but this is not an exhaustive source-equivalence proof.

The tests exercise simulation and prefix rules, equivalence classes and their
refinement, SAT/CNF queries, source assertions, mapped sweeps and BMC frame
unrolling. Source-derived BMC tests cover constant-true and constant-false
outputs and a replayed sequential witness, with rewriting on and off. Model
extraction is tested with a real solver, a permuted 40-variable mapping across
32-bit words, and first-frame PI/final-register counterexample projection.
Other tests cover class representative selection, simulation-weight pruning,
class-derived frames, BMC implication filtering, source SAT constant queries,
failed-node budget policy, and public sweep option adapters. The source
assertions on output phase, FRAIG mapped nodes and SAT activity traversal have
red-first regressions. Further tests directly exercise `Fra_FraigEquivence`, source multi-frame and
saturated distance-one assignment, SAT-proved implications, and actual
local/global conflict and global inspection exhaustion during equivalence
search. These tests are source-derived Go fixtures. Some fixtures compare
against compiled C; the model and BMC outcome tests are source-derived Go
regressions, not compiled-C comparisons.

Meaningful remaining gaps include:

- Additional SAT-equivalence and implication outcomes, including solver-limit
  boundaries, timeout propagation during a complete sweep, and the remaining
  branches in `Fra_ImpCheckForNode`.
- Less common `Fra_FraigSweep`, `Fra_BmcPerform`, and `Fra_BmcPerformSimple`
  branches, including resource-exhaustion and verbose reporting behavior.
- Additional sequential-simulation and counterexample edge cases, especially
  multi-frame pattern reconstruction and additional distance-one boundaries.
- Lifecycle, diagnostic printing, public accessor and trivial adapter paths
  have lower priority and are not comprehensively covered.

The no-internal-node BMC fast path and native EXOR source behavior are documented
in [README.md](README.md); their presence limits what BMC results can establish.

The SAT queries pass local and global conflict and inspection budgets to
BSAT, as in `src/proof/fra/fraSat.c`; unknown is treated as a timed-out proof.
The direct query does not supply a wall-clock deadline.
