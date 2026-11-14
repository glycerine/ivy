# Exact-synthesis store port review

Compared the `Ses_Store*` Go routines with `src/base/abci/abcExact.c` lines
385–871. The implementation preserves the 1024-bucket table, 32-bit hash
accumulation and modulo, truth-key word count, and head insertion order for both
truth entries and arrival-profile entries. Duplicate insertions do not replace
stored solutions, while the synthesized/unsynthesized counters still increment
on every call, matching C.

The database format writes the native-endian 64-bit C `unsigned long` count on
the current 64-bit target, then four 64-bit truth words, 32-bit variable count,
eight 32-bit arrival values, 32-bit resource-limit flag, and either the exact
solution byte sequence or three zero bytes. Read filtering occurs after the
record payload is consumed, preserving file alignment. Synthesized lookups
normalize a copy of the query arrival profile for each candidate network;
unsynthesized lookups compare the stored normalized profile directly.

`test_vectors/ses_store_unsynth_v1.bin` captures a one-record database fixture
in C field order. Tests cover collision-chain order, duplicate/counter behavior,
round-trip I/O, network-normalized lookup, exact synthesized-record field
order, and category filtering. Store allocation and cleanup use the public
`bsat.NewSolver`/`Close` lifecycle that backs CUDD's owned SAT solver.
