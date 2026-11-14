# `abcTim.c` Go port review

Reviewed against `src/base/abci/abcTim.c`.

- The active topological path is represented by `Abc_NtkTestTimDeriveGia`; the source's `#if 0` pin-boundary experiment is intentionally excluded.
- White-box selection keeps the C rule: node ID divisible by 200, with one through nine fanins. The timing box ID is its fanin count, matching the source delay-table indexing.
- The derived manager preallocates all main and hole CIs before hashing, restores saved fanin literals, emits main outputs followed by box-input outputs, and attaches the hole graph and TIM manager to the cleaned GIA.
- `Abc_NtkTestTimPerformSynthesis` follows the source's AIG conversion, choice/compress selection, and GIA conversion. Timing metadata is cloned when the native graph bridge duplicates it.
- Choice validation and reversal use equivalence representative classes (`pReprs`/`pNexts`), not structural choice siblings. AIGER round-trip uses a deterministic Go TIM payload in the hierarchy extension and preserves any separate timing extension; it preserves terminal, box, delay table, and timing state.
- The optional `abcProve.c` `ABC_USE_CUDD` proving fallback is available under Go build tag `abc_use_cudd`; the default build uses `!abc_use_cudd` and skips that conditional C branch. It is inserted at the same point before the final SAT attempt.

Validation is pending the repository's current `abcDress.go` parser errors; the focused `misc/tim` tests can run independently.
