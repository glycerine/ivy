# abcIvy prefix port review

Scope: the Go prefix from `Abc_NtkIvyBefore` through `Abc_NtkIvy`, matching the corresponding C routines in `src/base/abci/abcIvy.c`. `Abc_NtkFromIvy` and later conversion helpers are reviewed by the separate tail owner.

## Behavior checks

- `Abc_NtkIvyBefore` retains the source checks and BDD-to-SOP conversion, warns on self-feeding latches and choice nodes, checks the generated Ivy manager, and builds sequential state from collected latch initial values only when requested.
- `Abc_NtkIvyAfter` chooses the sequential or combinational conversion, performs cleanup only for non-HAIG output, copies EXDC, and checks the resulting network.
- The strash, HAIG, cuts, rewrite, sequential rewrite, resyn, SAT miter, and FRAIG wrappers retain their source call order and stop each replaced/temporary Ivy manager at the same point as C. `Abc_NtkIvySat` follows the active C body: its `Ivy_FraigPerform` line is commented, so it calls `Ivy_FraigMiter` directly.
- `Abc_NtkIvyProve` preserves the C proof sequence: initial SAT, optional large-network rewrite pass, Ivy proving, fallback SAT, then the CUDD-guarded BDD route when enabled. SAT and Ivy counterexample payloads are transferred into the network and released from their source manager.
- `Abc_NtkIvy` intentionally returns nil after checking and stopping the converted manager. The experimental rewrite calls and conversion-back path in C are commented out; keeping them active would change the C routine's behavior.
- Edge encoding follows C's `(object ID << 1) | complement` representation. Decode helpers return complemented-edge values, and the Ivy fanin helpers preserve both Ivy edge complement and the network edge complement.

## Compactness and lifecycle

The Go wrappers use slices and manager APIs instead of recreating C pointer arithmetic. The only added cross-package reads are the narrow `Ivy_ManReadHaig` and `Ivy_ManTakeData` accessors. They expose precisely the source fields read or transferred by this integration layer. Edge helper tests cover encoded IDs, complement bits, and regular/complement operations.

## Verification

`go test ./aig/ivy` passes, including compilation of the new accessors. `go test ./base/abci` currently stops at parser errors in `abcMap.go`, before the package can compile the Ivy prefix or its new edge tests. The parser errors reported are in that separate file; no Ivy parser errors remain.
