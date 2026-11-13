# FRA manager test coverage

The package's statement coverage was 84.8% on the current tree, measured with
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
red-first regressions. Some fixtures compare
against compiled C; the model and BMC outcome tests are source-derived Go
regressions, not compiled-C comparisons.

Meaningful remaining gaps include:

- A direct test of the source-named `Fra_FraigEquivence` entry point; its
  mapped underlying engine and the public sweep option adapters have tests.
- Additional SAT-equivalence and implication outcomes, including solver-limit
  boundaries and the remaining branches in `Fra_ImpCheckForNode`.
- Less common `Fra_FraigSweep`, `Fra_BmcPerform`, and `Fra_BmcPerformSimple`
  branches, including resource-exhaustion and verbose reporting behavior.
- Additional sequential-simulation and counterexample edge cases, especially
  multi-frame pattern reconstruction and distance-one boundaries.
- Lifecycle, diagnostic printing, public accessor and trivial adapter paths
  have lower priority and are not comprehensively covered.

The no-internal-node BMC fast path and native EXOR source behavior are documented
in [README.md](README.md); their presence limits what BMC results can establish.
