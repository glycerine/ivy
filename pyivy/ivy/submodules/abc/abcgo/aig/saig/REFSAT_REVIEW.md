# CEX refinement review

Files reviewed: `saigRefSat.go`, `saigRefSat_test.go`, and the frame-aware reason helper in `saigGlaCba.go`.

- The C source's active `Saig_ManFindCexCareBits` and `Saig_ManExtendCounterExampleTest3` paths use the structural reason and conversion routines. Their SAT minimization/refinement calls are commented out; the Go port keeps that active behavior and does not run the dormant experimental SAT path.
- CEX unrolling fixes the requested leading PIs and initial state from the input CEX, gives later PIs fresh frame inputs, copies graph nodes, and complements the selected final output as the C routine does.
- Reason conversion preserves both original PI IDs and frame indices. The care CEX clears all bits before setting just the frame-specific reasons; the extension API deduplicates original PI IDs in first-reason order.
- The shared frame-aware CBA reason finder handles latch transitions and per-frame PI values, so sequential support is followed through prior LI drivers.
- Invalid graph/CEX dimensions return nil instead of indexing beyond CEX storage.

Focused tests cover the PI IDs returned by the extension routine and a frame-specific care mask. CBA has a separate regression for DFS reason order. This is source-derived; no direct C function oracle was available. The SAT helper routines from the C file were not retained because their only call sites are inside commented-out experimental blocks and none are used in-repo.
