# Sixteen-variable clustering/decomposition port

Scope: the active algorithms in `ifDec16.c`, ported in `ifDec16.go`; test fixtures are under repository `test_vectors/map_if/`. The shared IF header remains owned by the package lead.

## Design and readability review

Native word slices replace pointer-offset truth storage. Generic bit transforms provide phase changes, cofactoring, adjacent/nonadjacent swaps and support shrinking with explicit variable-to-position maps. Group descriptors retain source variable order and cofactor multiplicity, with explicit nibble/byte encodings for cache and LUT descriptors.

The decomposition search preserves the source two-round add/remove-variable heuristic, easy-cofactor peeling, smaller-bound-set retries, reversed-order retry, start/stop restrictions and shared-variable compatibility checks. Derivation preserves source cofactor class ordering and constructs actual bound/root LUT truth tables. Three-LUT paths include sequential decomposition, structure restrictions and the two-output/four-cofactor alternative. Verifiers reconstruct the complete function, including a second LUT depending on the first LUT's synthesized signal.

The native cache retains source hash arithmetic, hit counts, median-based pruning, frequency ordering and bounded entry policy, using typed Go heap entries in the existing manager bucket maps. Truth operations and composition have no mutable global elementary-table initialization. LUT-structure checks, delay/cost estimation, small input/output decomposition checks and descriptor export are implemented rather than replaced by constant answers.

## Separate bug review

Checked source nibble encoding, endian-explicit hash extraction, truth padding, cofactor class ordering, bounded cofactor counts, shared-variable position, returned LUT phase/order, three-stage synthesized IDs, source search tie order and cache pruning.

A source-review finding prevented an incorrect reuse of KIT semi-canonicalization: this file's algorithm orders cofactors in the opposite direction and keeps input-phase bits in original-variable order, without output normalization. The implementation now ports its own exact rule; the original-C differential verifies phases, permutations and complete truth words.

Native safeguards support in-place truth swaps through copied input storage, explicitly apply bit offsets to wide cofactor extraction, and preserve full uint64 values in six-input decomposition checks (several C temporary arrays were int). Nil managers use the source default unrestricted three-LUT structure rather than dereferencing nil. Invalid dimensions/short mutable descriptors fail explicitly instead of relying on pointer writes.

The source checker expects a minimal/full support when its nVars argument is supplied. C oracle fixtures enforce this precondition to avoid the source's independent-variable assertions; high-level cut checks perform support minimization when configured. Constants and single-LUT cases are handled by the high-level wrapper before the multi-LUT search.

## Validation

- Source examples and independent full-truth reconstruction for seeded two-LUT functions.
- Thirty seeded three-stage functions, verifying all assignments for every discovered decomposition.
- Phase/permutation round trips through three to ten variables; packed cache result retention, pruning and descriptor/delay checks.
- Original C oracle (`test_vectors/map_if/clu_oracle.c`): seeded structured/random two-LUT functions with input-position restrictions, comparing bound/root groups, selected variables, derived truth words and failures; exact semi-canonical phase/permutation/truth comparisons.

The focused production set is the actual shared header/types plus ifCut/ifTruth/ifCount/ifDec16 and assertion helper. Add `ifDec16_native_test.go` and `ifDec16_c_test.go` for the focused race/vet gate. Full package integration remains with the IF lead while other original files are being ported. The C oracle is optional when no C compiler/ELF linker is available; production remains pure Go.
