# Exact synthesis API tail review

Compared the Go tail of `abcExact.go` with `src/base/abci/abcExact.c` lines 2414–3024.

- `Abc_NtkFindExact` and `Gia_ManFindExact` preserve the source argument values, verbosity flags, solution extraction choice, runtime accounting, and cleanup order.
- `Abc_NtkFromTruthTable` passes the C `word*` truth bits as little-endian 32-bit limbs to the SOP cover constructor, then applies short names before releasing the temporary arena.
- The test entry points preserve the source's `0xcafe` fixture, arrival profile, bounds, equivalence checks, and expected minimum size. The Go API checks nil results before dereferencing to make failures diagnosable.
- Mapper APIs retain the 8-input limit, manager configuration, cache key (truth, variable count, normalized arrivals), delay bound formula, `nVars-2` initial gate count, iterative depth improvement, SAT statistics, and store-before-return flow. Pin delays use `[]int8`, matching the source `char*` representation on the target C ABI.
- `Abc_ExactBuildNode` decodes the source solution record in order, reconstructs each 2-input SOP gate, applies final-output polarity to the final gate truth table, and returns the literal-selected output node.
- Go `Abc_ExactStop` clears the global manager after cleaning it. The C tail frees the manager but leaves its static pointer non-NULL, making the stopped manager appear active and preventing a safe restart; clearing it preserves the intended start/stop lifecycle.
- The depth-improvement loop uses `nMaxDepth > 0`; the source's truthy `while (nMaxDepth)` would continue decrementing below zero if a negative maximum were supplied. For nonnegative legal bounds it has the same search behavior and terminates safely.

Package-level execution remains pending because the earlier solver prefix in the same file still has C-style loop syntax; current `go test ./base/abci` reports parser errors before reaching this tail.
