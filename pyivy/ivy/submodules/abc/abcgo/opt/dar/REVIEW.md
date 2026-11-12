# DAR port review

## Implemented scope

The original complete four-input library (43,906 ANDs, 24,772 roots), priority ordering, NPN classification, bounded cuts, and real MFFC-cost rewriting are implemented. Balancing, exact-cofactor refactoring, Rwsat, Compress/Compress2 variants and classic choice orchestration are implemented. NewChoice DCH/CEC engine orchestration and power-driven optimization remain outside the completed API scope. All current Go files compile together.

## Separate design review: completed core

Cuts own a logical slice of sorted edges and one truth word. There is no redundant leaf count, allocation capacity, signature, C bitfield packing, or raw-memory arena. Library preparation produces immutable typed DAGs cached under a mutex; optimization managers own graph mutation and statistics. Context replaces the global deadline. Returned cut boundaries and public library arrays are independent copies. Power mode explicitly reports its missing switching-activity implementation rather than silently changing its meaning.

The original library's special class-one root policy and source priority ordering are retained. Buffer-based stable graph replacements are compacted by the caller's DFS duplication stage. A redundant traversal increment was removed because the AIG cut-MFFC helper already advances and labels its traversal generation.

## Separate bug review: completed core

Exhaustive tests reconstruct all 65,536 truth functions from NPN matches and verify support minimization against every assignment. Exact library array hashes and topology checks prevent transcription errors. Randomized graph rewrites are compared against their original full truth tables, including complemented boundaries and zero-gain candidates. A distributed-logic regression requires an actual node reduction after DFS compaction, ruling out a no-op implementation. Cancellation and parameter bounds are checked.

Review found that structural hashing can rediscover the old root inside a candidate network even when its final edge differs. A dependency check now rejects such a replacement before mutation; a regression verifies detection. This prevents cyclic replacements. Root phase and cuts with internal boundary nodes have separate truth tests.

## Separate design review: balancing, refactoring and orchestration

Balancing uses typed associative boundaries and slice sorting by actual graph level. AND idempotence/contradiction and XOR cancellation/phase are explicit operations, rather than pointer flags. Input graphs are cloned, so ordered interfaces and sequential metadata survive without ownership transfer rules. Refactoring uses exact bounded truth tables and cofactors rather than the original Kit factoring heuristic; the result must pass original MFFC area and requested level acceptance. Extend permits crossing shared fanouts when the support bound allows it. Search limits are validated before allocation. Script orchestration owns independent graph versions and propagates context and writer failures.

## Separate bug review: balancing, refactoring and orchestration

A red eight-input depth regression exposed replacement buffers stopping cone flattening prematurely. Single-fanout buffers now forward cone collection; AND and native XOR depth both reduce seven to three while all 256 assignments are preserved. XOR complemented duplicates cancel to the correct constant. Shared boundaries and extended-boundary limits have regressions. Five-input cofactor reconstruction is compared on every assignment, and absorption must actually reduce area. Classic constructive and partitioned choice paths both preserve the complete truth table. Rwsat's adapter preserves CI/CO order and tests the source stage count. Writer errors and silent short writes propagate explicitly.

## Validation

The connected repository `go -C abcgo test -race ./opt/dar` passes after AIG retiming integration. Earlier focused temporary-module checks used copies of the converted AIG files; no repository files were hidden or altered. Final connected race validation supersedes those focused checks. Exact Kit heuristic/search parity is not claimed, and power mode returns a clear unsupported-activity error instead of silently ignoring the request.
