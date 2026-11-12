# Lucky port review

Ports the fast, simple, and swap algorithms from ABC `src/bool/lucky`, originally by Jake/Yasha. Truth tables keep the source's low-word-first packed `uint64` layout and high-word-first numerical comparison.

## Compactness, idiomatic Go, and readability pass

Performed separately from the bug review below.

- Pointer arithmetic/type-punning became slices, explicit 32/64-bit packing, byte permutation arrays, and ordinary structs. No cgo or unsafe.
- Shared the eight signed adjacent permutations between the fast six- and sixteen-variable paths. This replaces repeated quarter-comparison implementations while preserving candidate order, cofactor ties, phase flags, and transformation bookkeeping.
- Adjacent swaps and phase changes stay word-parallel. Arbitrary swaps compose adjacent transpositions and restore intervening variables.
- The simple algorithm retains Steinhaus–Johnson–Trotter swaps, Gray-code phases, and mixed-radix grouped traversal, with small helpers for each role.
- Fixed scratch arrays became correctly sized slices; global diagnostic timing is mutex-protected. `math/bits` supplies population counts; buffered standard-library I/O supplies truth-table storage.
- Public lowercase C entry points have exported Go wrappers: `LuckyCanonicizerFinalFast`, `LuckyCanonicizerFinalFast1`, `ResetCanonPermutation`, `NewPermInfo`, `SimpleMinimal`, `SimpleMinimalGroups`, and `ReadTruthStore`. Original internal names remain recognizable for source comparison.
- Truth-file APIs expose errors. `Abc_TtStore_t` exposes `NVars`, `NWords`, `NFuncs`, and `Funcs`.

The fast single-word path intentionally operates in six-variable space, as in C, including for smaller stretched tables; its permutation buffer needs at least six entries. Initialize permutation entries with `ResetCanonPermutation` before calling a canonicalizer. The general 32-bit `Kit_TruthSemiCanonicize_new` interface instead tracks the supplied actual variable count.

## Separate bug review

Two review findings were first demonstrated by failing regressions and then fixed:

1. The simple semi-canonicalizer dereferenced a nil cofactor-count buffer even though the original caller passes NULL. Allocate local counts when nil. Regression: `TestSimpleSemiCanonicalizerNilStore`.
2. Hex output accepted a partial nil-error writer result as success. Return `io.ErrShortWrite`. Regression: `TestHexShortWrite`.

Additional source boundary defects addressed during porting:

- Empty truth-store allocation and sorting no longer dereference element zero.
- Simple minimization with zero/one variables does not attempt an invalid adjacent swap.
- Group and file buffers use their true word dimensions; prefixed multiword output writes each table once rather than walking out of bounds/repeating entire tables at each word.
- Store-line counts describe actual parsed functions rather than appended sentinel newlines. Blank lines and CRLF are handled without C end-of-buffer assumptions.
- Exhaustive permutation metadata rejects counts beyond the source's int32 factorial range (twelve variables); fast algorithms support through sixteen variables.

Reviewed reconstruction metadata, output-complement/tie flags, candidate ties, table aliasing for swaps/copies, grouped input/output phase options, and immutable mask tables. Grouped `flipInput=false` retains the source behavior of disabling phases on singleton groups; larger groups still enumerate their phases.

## Validation

- `GOCACHE=/tmp/abcgo-cache go -C abcgo test ./bool/lucky -count=1` — passes.
- `GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./bool/lucky -count=1` — passes.
- `GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./bool/lucky` — passes.
- Manual C/Go differential: 550 cases across both fast variants, six/seven/eight/ten/sixteen variables, random and constant/symmetric tables. Exact output tables, phase flags, and all sixteen permutation entries match the compiled original C.
- Permanent C-reference golden cases include balanced/symmetric and multiword tables.
- Independent bit-assignment checks cover all swap/phase variable positions and cofactor counts through ten variables.
- All 256 three-variable functions: simple minimization equals an independent exhaustive permutation/input-phase/output-phase reference.
- Grouped minimization checked against independent enumeration; SJT and Gray traversals checked for uniqueness and complete coverage.
- Both fast variants independently reconstruct originals from returned metadata through sixteen variables; the 32-bit public interface is likewise checked.
- Hex/store round trips through sixteen variables, prefixed output formatting, empty files, small-support boundaries, nil scratch storage, and partial writes.

No experiment directories were tested and no source files were hidden or disabled.
