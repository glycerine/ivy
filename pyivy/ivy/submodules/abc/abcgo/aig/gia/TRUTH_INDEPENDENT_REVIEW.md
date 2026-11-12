# Independent truth-table review

## Compactness and readability

Reviewed `giaTruth.go` independently after the implementation author's review. Removed the unused evaluator dimension field and an unnecessary CI-vector copy; the returned CI slice is explicitly borrowed and only read during evaluation. Evaluator memoization, workspace materialization, Shannon synthesis, and NPN grouping remain separate operations. The typed graph clone is appropriate until the general cone duplicator is available.

## Separate bug review

Checked complemented edges, specialized XOR/MUX evaluation, constant and cut boundaries, support ordering, borrowed workspace results, object numbering and sentinel reset, minimum-base truth compression, phase/permutation grouping, and sequential cone closure. Register inputs and outputs are appended in matched discovery order. NPN grouping intentionally uses the original bounded semi-canonicalizer rather than an exhaustive equivalence oracle. The clone preserves functional cones and sequential interface; hashed construction can change internal object counts and does not reproduce incidental source scratch-Value writes. No pending semantic findings.

The focused real-core truth suite and canon package pass with the race detector. Existing tests include original-C differential evaluation, all three-variable synthesis cases, multiword random graphs through 12 inputs, workspace metadata, specialized nodes, NPN classes, and reachable register-input closure. The complete GIA package still contains unrelated draft files.
