# Logic-network retiming port

All active RET source files have native Go implementations. Integration and validation are **in progress**: ordinary package execution currently stops at unrelated syntax errors in the actual `base/abc` dependency. No RET source is excluded and no substitute network implementation is used.

## Algorithms

- `retCore`: mode dispatch, sequential dangling-cone cleanup, timing/area reporting. The diagnostic debug checker uses exhaustive reachable-state exploration of the two concretely initialized machines instead of importing the legacy SEC command stack.
- `retIncrem`: latch-terminal detachment/restoration, compatible latch sharing, enabled-node moves, forward initial-state propagation, backward initial-state constraint construction, one-way iteration and delay replay.
- `retArea`: frontier marking, minimum-cut retiming, initial-state cone construction/composition, latch-boundary replacement with reused terminal objects and explicit box order.
- `retFlow`: directed augmenting paths and residual rerouting, minimum-volume cuts, cut verification, retimable cut updates; alternative original traversal helpers and diagnostics are retained.
- `retDelay`: unit-delay critical-node analysis, best-iteration search on a copy, replay with initial-state preservation.
- `retLvalue`: Pan L-value convergence, period binary search, signed-byte lag output, and optional lag application helper. As in the original active source, mode 6 computes/reports lags without applying them.
- `retInit`: SOP simulation and initial-state SAT. A direct SOP Tseitin encoding uses the native BSAT solver with the source limits of 500000 conflicts and 50000000 propagations. Every SAT model is independently resimulated. The source verifier's unconditional success is replaced by an actual mismatch result. Fixed-seed cycling uses a local glibc-compatible random generator rather than process-global `srand`/`rand`.

## Representation

The package uses the real `base/abc.Network`, `abc.Object`, and explicit complemented `abc.Edge` values. Pointer/int reinterpretations become `Object.Temp` integers; actual copy/path bindings use `Object.Copy`. Vector arguments become typed slices. `LatchBindings` preserves original BI/BO objects during detached retiming. Slice lengths replace independent counts. A real Go object sentinel replaces flow's `(Abc_Obj_t*)1`; the forward area-progress sentinel becomes an explicit boolean.

Prepared retiming graphs connect latches directly and can contain sequential cycles. RET patches incident fanins locally during that phase, then restores BI/BO boundaries and network interface order before validation. It does not invoke the general network transfer routine's combinational-cycle rejection on this temporary representation.

## Validation status

Implemented tests cover direct forward/backward moves, all six modes, exhaustive sequential equivalence on small concrete machines, initial-state SAT and UNSAT, rejected incorrect models, SOP simulation, lag arithmetic, compatible local random sequence, and feedback cycles. A differential driver linking the actual original ABC objects has been compiled and executed independently: all twelve mode/topology cases return latch counts, levels, and exhaustive 256-input four-frame traces for the pending Go comparison.

Pending gates:

- Run normal full RET package tests after `base/abc` builds.
- Resolve all resulting failures, including original-C differential mismatches.
- Run RET race tests and vet.
- Complete separate independent readability/compactness and bug reviews; fix concrete findings with regressions.
