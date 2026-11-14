# Network memory accessor review

`Abc_NtkMemoryVectorBytes` mirrors the vector-storage part of `Abc_NtkMemory` in `src/base/abci/abcPrint.c`:

- Manager-vector backing storage uses slice capacities for PI, PO, CI, CO, and object arrays, matching the source's allocated-capacity accounting.
- Per-object fanin and fanout storage uses each slice's capacity and ABC's `sizeof(int)` entry width, matching `Vec_IntCap` rather than logical lengths.
- The caller remains responsible for the network and object struct terms, which it already adds separately. `Network` and `Object` struct sizes include their Go slice/map headers; this helper counts backing entries only.
- The current Go `Network` stores traversal ID and maximum level as scalars and has no `vTravIds` or `vLevelsR` backing vectors. Those C terms have no allocated Go vector capacity to report.
- `Abc_NtkBarBufs2` is the existing read accessor for the secondary count printed by `abcPrint.c`; the focused test verifies its increment/read path.

The estimate deliberately follows ABC's source accounting model rather than attempting to report Go heap allocator overhead.
