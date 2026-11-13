# FRA manager test coverage

The package's statement coverage was 87.2% on the current tree, measured with
`go test -coverprofile=/tmp/framgr.cover ./proof/framgr`. Coverage is a guide to
unexercised code, not proof of equivalence to ABC. A source and independent
bug/simplicity review found and restored missing preconditions, but this is not
an exhaustive source-equivalence proof.

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
red-first regressions. Further tests directly exercise `Fra_FraigEquivence`,
source multi-frame and saturated distance-one assignment, SAT-proved
implications, and actual
local/global conflict and global inspection exhaustion during equivalence
search. These tests are source-derived Go fixtures. Some fixtures compare
against compiled C; the model and BMC outcome tests are source-derived Go
regressions, not compiled-C comparisons.

A complete sweep with a deliberately exhausted SAT budget matches compiled C
for failed calls, recorded node IDs, cleared representatives, and resulting AIG
nodes. A second complete sweep matches C when later disproved nodes are also
added to the timeout list after an earlier SAT limit is hit.
BMC tests also check prefix-frame exclusion and replay a 20-frame counterexample
whose bits cross a 32-bit boundary. Optional checks run with
`ABC_FRA_C_ORACLE=1 go test ./proof/framgr`: the counterexample oracle compiles
all linked ABC units from source; the sweep oracle freshly compiles six FRA
units and links the remaining dependencies from prebuilt ABC objects.

## Algorithm and branch map

| C algorithm / behavior | Go implementation | Focused regression coverage |
| --- | --- | --- |
| `Fra_SmlSimulate`, fixed 0/1 and distance-one patterns, saturation | `fraSim.go`, `fraSimulation.go` | `fraSimulation_test.go`, `fra_empty_patterns_test.go`, `fra_entry_distance_boundary_test.go`, `saturation_test.go` |
| `Fra_FraigNode` SAT proof, disproof, timeout, failed-node handling | `fraCore.go`, `fraSat.go` | `fraCore_test.go`, `fra_sat_failed_node_source_test.go`, `fra_sat_limit_boundary_test.go` |
| `Fra_FraigSweep` whole-pass timeout order and timeout representatives | `fraCore.go` | `fra_sweep_timeout_source_test.go`, `fra_sweep_timeout_followup_test.go`, `fra_sweep_timeout_c_oracle_test.go` |
| Latch-correlation sweep exits after processing latch outputs | `fraCore.go` | `fra_sweep_latch_corr_test.go` |
| `Fra_BmcFrames` initial state, frame transitions, node mappings | `fraBmc.go`, `fraBmcFrames.go` | `fraBmcFrames_test.go`, `fra_bmc_child_source_test.go`, `fra_entry_distance_boundary_test.go` |
| `Fra_BmcFilterImplications` prefix, common representative and polarity cases | `fraBmc.go` | `fra_bmc_prefix_source_test.go`, `fraBmc_test.go` |
| `Fra_BmcPerform` callback filtering, class refinement, cleanup and no-node path | `fraBmc.go` | `fraBmc_test.go`, `fra_adapter_source_test.go` |
| `Fra_BmcPerformSimple` direct assertion, rewritten/unrewritten frames and replayed trace | `fraBmc.go` | `fra_bmc_outcomes_source_test.go`, `fraBmc_test.go` |
| `Fra_BmcStop` releases its owned frame graphs and node maps while preserving the original graph | `fraBmc.go` | `fraBmc_test.go` |
| Solver model packing and counterexample projection | `fraSat.go`, `fraSim.go`, `fraSimulationCex.go` | `fra_model_extraction_source_test.go`, `fra_cex_multiframe_test.go`, `fra_bmc_multiframe_source_test.go`, `fra_sat_model_words_test.go` |
| SAT equivalence, implication, clause, constant, limits and counterexample model | `fraSat.go` | `fraSat_test.go`, `fra_constant_sat_source_test.go`, `fra_sat_limit_boundary_test.go`, `fra_sat_model_words_test.go` |

The current uncached race run and vet pass. Statement coverage is now 87.2%.
The SAT-query regression added this pass drives the solver-model branch itself
and checks CI packing over two 32-bit words; the existing extraction fixture
continues to check permuted SAT-variable IDs and register/PI projection.

Meaningful remaining gaps include:

- Less common `Fra_FraigSweep`, `Fra_BmcPerform`, and `Fra_BmcPerformSimple`
  branches, especially level-limited node ordering and verbose diagnostic text.
- Additional SAT-equivalence and implication combinations beyond the already
  covered proof, counterexample, failed-node, and budget-boundary cases.
- Additional sequential-simulation and counterexample edge cases, especially
  larger multi-output pattern reconstruction and additional distance-one
  boundaries.
- Lifecycle, diagnostic printing, public accessor and trivial adapter paths
  have lower priority and are not comprehensively covered.

The no-internal-node BMC fast path and native EXOR source behavior are documented
in [README.md](README.md); their presence limits what BMC results can establish.

The SAT queries pass local and global conflict and inspection budgets to
BSAT, as in `src/proof/fra/fraSat.c`; unknown is treated as a timed-out proof.
The direct query does not supply a wall-clock deadline.
