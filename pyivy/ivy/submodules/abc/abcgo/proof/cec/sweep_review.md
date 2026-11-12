# Native SAT sweeping review

`SatSweep` and source-shaped `Cec_ManSatSweeping` perform real combinational SAT sweeping. Deterministic simulations form phase-normalized candidate buckets; incremental SAT checks prove equivalences; only proved representatives enter the rebuilt circuit. SAT witnesses reject candidate relations and undecided queries never authorize a merge. Primary inputs and register outputs remain free variables.

The native implementation uses direct candidate proofs instead of the C speculative-reduction/resimulation pipeline. Simulation scheduling and optimization knobs therefore do not reproduce C run statistics or resource-limited outcomes. `nWords` controls initial samples, `nLevelMax` excludes deeper nodes, zero `nItersMax` disables proofs, conflict/time budgets constrain proofs, and miter/dual-miter flags request counterexample checks. The original alternative-solver/rewrite/color/original-ID tuning fields remain source metadata; the native engine uses the shared Satoko proof backend. Sequential correspondence is a separate unfinished engine.

## Readability pass

One immutable model provides simulation and witness validation; one core manager owns incremental solver state. Candidate buckets hold only unmerged roots. Proven representatives and zero-input phases are the only state passed to reconstruction. Reconstruction uses the existing AIG constructors for simplification and structural hashing, then the actual GIA bridge; no parallel low-level packed-node implementation was introduced. Signature bits use direct byte indexing. Public Fraig wrappers share the native sweep and existing representative transfer.

## Separate bug pass

Checked normalized phase equality, complemented representative reconstruction, constant classes, canonical earlier-ID roots, output/buffer aliases, sequential free-input treatment, miter witness ownership, deadlines, zero iteration limits, level limits, constraint counts and stale metadata. Installed zero-input phases explicitly before exposing representative classes. Preserved constraint and register counts in rebuilt graphs. Fixed the high-level Fraig wrapper to duplicate through proved representatives (`Aig_ManDupRepr`) rather than the ordinary copier, which would retain the redundant gates. Resource-limited proofs leave nodes unmerged.

## Evidence

`test_native.sh -race` passes focused actual GIA/CEC integration. Tests check associativity-cone merging, complemented classes, free register outputs, level/iteration exclusions, miter counterexamples, public AIG verification and public Fraig reduction. `sweep_reference_test.go` executes the original repository ABC binary with `&fraig -J 0`; exhaustive input enumeration confirms that original C and native reduced outputs match and both remove redundant cones. This test skips if the original executable is unavailable. The focused script supplies its exact repository path.

This is scoped evidence: full GIA/CEC packages still contain other unported files, and sequential correspondence/choice engines remain separate integration work.
