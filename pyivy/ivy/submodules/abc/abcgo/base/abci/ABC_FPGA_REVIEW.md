# FPGA network adapter review

Compared `abcgo/base/abci/abcFpga.go` with `src/base/abci/abcFpga.c` (the complete 288-line translation unit).

- `Abc_NtkFpga` keeps the source sequence: validate STRASH input, warn on choice nodes, compute 4096-pattern switching activity, import to the FPGA manager, set recovery/switching/latch/target controls, map, reconstruct a BDD network, minimize its base, copy EXDC, and check the result.
- `Abc_NtkToFpga` imports CI arrivals and names, applies the latch-path arrival override to PIs, maps constants and CIs, traverses the AIG in DFS order, preserves complemented fanins, transfers complete `Choice` chains into FPGA `NextE`/`Repr`, and connects CO drivers without adding CO phase.
- `Abc_NtkFromFpga` follows the source reconstruction order: start a BDD network, clear FPGA traversal/data marks, bind CI and constant nodes, recursively derive each mapped output cut, finalize outputs, remove an unused constant, and decouple CO drivers.
- `Abc_NodeFromFpga_rec` uses the selected cut's leaf order to create fanins and delegates cut truth construction to the FPGA Boolean builder backed by the destination CUDD manager. It takes and stores a CUDD reference on the node root.
- C's switching simulator allocates random words at CIs and simulates AND nodes in DFS order. The Go adapter does the same at 128 words per CI using the package's source-compatible glibc RNG sequence and computes C's `2*p*(1-p)` score. The RNG is a package-local stream initialized to `srand(1)`; unlike C's process-global `rand`, unrelated random calls elsewhere in Go do not consume this stream.
- Go `Object.CopyData` carries FPGA pointers while the source C uses type-punned `pCopy`. `Abc_NtkStartFrom` clears that scratch data before it installs copies into the destination network, so the two scratch lifetimes do not overlap.
- BDD-aware `Abc_NodeComplement`, `Abc_NodeComplementInput`, and constant/buffer constructors preserve CUDD node values instead of replacing them with generic function closures. These exact `base/abc` support paths are exercised by FPGA output decoupling and constant outputs.
- The mapper manager is freed with a deferred cleanup, including failure returns. C frees it after successful reconstruction but returns early without freeing it if reconstruction fails.

Focused tests cover full-network equivalence for all eight three-input assignments with complemented fanins and switching both disabled/enabled, constant and CI buffer outputs, choice-chain import and mapping, and the exact first C `SIM_RANDOM_UNSIGNED` word. Base/abc tests cover BDD-aware complementation and function-valued data preservation. The tests pass in an isolated harness while full `base/abci` integration is blocked by unrelated C syntax in `abcIvy.go`.
