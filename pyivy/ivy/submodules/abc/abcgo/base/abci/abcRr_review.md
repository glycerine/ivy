# abcRr source review

`abcRr.go` inventories all routines compiled from `src/base/abci/abcRr.c`: `Abc_NtkRR`, manager start/stop/clean/stats, prove/update/window, TFI/TFO traversal, window extraction, simulation setup/cleanup/refinement, and the three simulation helpers.

The port keeps the source's object-ID traversal order, level-bounded frontier expansion, mark transitions, fanin/fanout candidate order, and proof/update sequencing. Pointer-valued scratch data is represented with `Object.Data` and `Object.CopyData`; complemented graph edges use `abc.Edge`.

The built C source ends `Sim_SimulateCollected` after reading an AND node's two fanins. The later AND-simulation implementation is commented out in C. Go preserves that behavior instead of inventing the missing calculation.

The C routine asserts `pFanout == NULL` in `Abc_NtkRRUpdate` while also containing a fanout-replacement branch and calling the routine with a fanout in `fUseFanouts` mode. The Go port preserves both the assertion and the branch, matching the source's behavior in assertion-enabled builds.
