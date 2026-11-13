# DAR port review

## Implemented scope

The original complete four-input library (43,906 ANDs, 24,772 roots), priority ordering, NPN classification, bounded cuts, and real MFFC-cost rewriting are implemented. Balancing, exact-cofactor refactoring, Rwsat, Compress/Compress2 variants and classic choice orchestration are implemented. NewChoice/NewChoiceAig orchestration uses a dependency-neutral engine interface with real native exact partitioned equivalence by default; the separate DCH/CEC implementations can be selected through adapters. Power-driven rewriting is implemented using the original activity stream and power acceptance rule. All current Go files compile together.

## Separate design review: completed core

Cuts own a logical slice of sorted edges and one truth word. There is no redundant leaf count, allocation capacity, signature, C bitfield packing, or raw-memory arena. Library preparation produces immutable typed DAGs cached under a mutex; optimization managers own graph mutation and statistics. Context replaces the global deadline. Returned cut boundaries and public library arrays are independent copies. Power simulation and scoring are run-local typed slices; no graph payload union or shared random state is needed.

The original library's special class-one root policy and source priority ordering are retained. Buffer-based stable graph replacements are compacted by the caller's DFS duplication stage. A redundant traversal increment was removed because the AIG cut-MFFC helper already advances and labels its traversal generation.

## Separate bug review: completed core

Exhaustive tests reconstruct all 65,536 truth functions from NPN matches and verify support minimization against every assignment. Exact library array hashes and topology checks prevent transcription errors. Randomized graph rewrites are compared against their original full truth tables, including complemented boundaries and zero-gain candidates. A distributed-logic regression requires an actual node reduction after DFS compaction, ruling out a no-op implementation. Cancellation and parameter bounds are checked.

Review found that structural hashing can rediscover the old root inside a candidate network even when its final edge differs. A dependency check now rejects such a replacement before mutation; a regression verifies detection. This prevents cyclic replacements. Root phase and cuts with internal boundary nodes have separate truth tests. A follow-up source audit restored constant/single-input support replacements before library matching, with an absorption regression. Cut caches are reused across unmodified roots and invalidated after every graph cleanup/replacement, avoiding repeated whole-cone enumeration and stale boundaries.

## Separate design review: balancing, refactoring and orchestration

Balancing uses typed associative boundaries and slice sorting by actual graph level. AND idempotence/contradiction and XOR cancellation/phase are explicit operations, rather than pointer flags. Input graphs are cloned, so ordered interfaces and sequential metadata survive without ownership transfer rules. Refactoring uses source bounded MFFC/reconvergence cuts, Kit ISOP and SOP factoring in both phases, followed by source MFFC area and required-level acceptance. Extend considers the source extended cut when the support bound permits it. Search limits are validated before allocation. Script orchestration owns independent graph versions and propagates context and writer failures.

## Separate bug review: balancing, refactoring and orchestration

A red eight-input depth regression exposed replacement buffers stopping cone flattening prematurely. Single-fanout buffers now forward cone collection; AND and native XOR depth both reduce seven to three while all 256 assignments are preserved. XOR complemented duplicates cancel to the correct constant. Shared boundaries and extended-boundary limits have regressions. Five-input Kit factoring is compared on every assignment, and absorption must actually reduce area. Classic constructive and partitioned choice paths both preserve the complete truth table. NewChoice and NewChoiceAig also preserve complete truth tables, restore names/specification and ordered interfaces, and record synthesis duration. Explicit engine failures or mismatched result interfaces are rejected; they never silently invoke the default engine. Rwsat's adapter preserves CI/CO order and tests the source stage count. Writer errors and silent short writes propagate explicitly.

## Validation

The connected repository `go -C abcgo test -race ./opt/dar` passes after AIG retiming integration. Earlier focused temporary-module checks used copies of the converted AIG files; no repository files were hidden or altered. Final connected race validation supersedes those focused checks. A compiled original-C oracle compares the complete factored graph structure for 103 five-variable functions in each output phase. Power mode is enabled and checked against original-C activity, MFFC and candidate evaluators.

## Separate design review: power mode

Activity simulation ports `giaSwitch.c` directly to bit-parallel stable-ID graph storage. The default remains ten 32-bit words, 48 frames and 16 settling frames, with zero initial registers. Registers update simultaneously from the last register outputs. The exact ABC multiply-with-carry stream is local to each simulation, including the reset call's discarded first draw and reverse word order. Context and byte-allocation overflow checks precede simulation work. Native XORs and replacement buffers use the graph's real topological order rather than assuming node IDs are topological.

Rewriting retains measured probabilities for existing nodes; newly built ANDs use the source independence approximation. The manager owns these values and releases them on Stop. The source MFFC edge-weighted sum and candidate evaluator's unusual terminal/reused-node probability contribution are preserved, including repeated-DAG visit handling and float/double rounding boundaries. Candidates are rejected when estimated power added exceeds power saved; area and level acceptance still apply.

The public typed simulation reports actual constant-one and CO signal probabilities, including polarity. The legacy Saig-to-GIA result projection dropped the constant literal phase and could round complemented CO probabilities differently; those projection artifacts are not used for the typed API. Internal AND/CI activity, which drives DAR scoring, follows the original simulation kernels exactly.

## Separate bug review: power mode

An actual-C oracle compiles unmodified original `giaSwitch.c`, `darLib.c` and `aigMffc.c` with a small typed-fixture harness. Eighty activity cases vary words, frames, settling prefix, register count, transition counting and input transition-mask factor. Two hundred candidate DAG cases vary complements, sharing, existing-node/MFFC flags and non-dyadic probabilities. One hundred MFFC cases include shared nodes and protected internal boundaries. Every compared float32 bit pattern and area count matches C. The harness counts actual CO vectors and represents the AIG's constant-one explicitly, as documented above.

Review corrected MFFC traversal to recurse at the last removed reference, matching the C dereference order rather than first DFS encounter. Explicit double-expression boundaries preserve source rounding. A byte-count overflow check prevents valid element counts from overflowing the backing allocation. Regression tests cover that bound, cancellation, stable-ID replacement buffers, native XOR polarity and concurrent deterministic simulations.

A fixed sequential toggle-register fixture checks exact original-C probability words. End-to-end tests require power-driven rejection of cost-increasing candidates, actual node reduction of a four-input parity SOP, and complete truth-table preservation. Full connected `go test -race ./opt/dar` and `go vet ./opt/dar` pass, including the live original-C oracle.
