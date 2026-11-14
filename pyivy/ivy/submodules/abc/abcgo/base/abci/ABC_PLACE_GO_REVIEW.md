# `abcPlace.c` Go port review

Reviewed against `src/base/abci/abcPlace.c`.

- `Abc_PlaceCreateCell` keeps the ABC object ID, empty label, pad/AND cell type, and the placement database registration.
- `Abc_PlaceUpdateNet` rebuilds each net from the ABC object's fanouts and gives the net unit weight, matching the source's fanout-terminal model.
- `Abc_PlaceEvaluateCut` remains the source placeholder and returns zero.
- `Abc_PlaceUpdate` registers newly added cells, updates their nets, deduplicates touched fanin nets, refreshes modified live nets, and leaves the source's commented incremental `fastPlace` call inactive.
- `Abc_PlaceBegin` allocates five slots per maximum ABC object ID, initializes pad and AND cell types, visits CIs/COs/ANDs in source order, builds CI/node nets, then invokes preplacement at 0.8 utilization and global placement.
- `Abc_PlaceEnd` releases the placement registrations and local records. Go's placement database stores pointers to these records, so unregistering them prevents dangling references after the C routine's array frees.

Focused tests cover initial derivation, incremental fanout-terminal refresh, placeholder cut cost, and cleanup.
