# Abstraction port review

The `proof/abs` package is translated and builds. Its current test suite runs
against the real VTA, GLA, old GLA, SAT, BDD, PDR, and optional SSW prover
components. Package race tests and `go vet` are required after changes.

ABS timeout clocks follow ABC's platform choices where supported: Linux uses
thread CPU time; POSIX systems with `getrusage` use process CPU time; and
Windows uses elapsed time, as Microsoft CRT `clock()` does. The Windows
clock does not reproduce the CRT's 32-bit overflow after about 24.8 days.
Targets without an implemented clock reject timed runs explicitly. Existing
ABS driver checks may poll for expiration more often than the C solver.

The review added exhaustive small-state checks for both GLA drivers: every
one-input, one-register transition/property pair is compared against all
three-frame input traces, with returned witnesses replayed on the original
GIA. The old GLA oracle exercises both normal and layer-based refinement.
Additional tests cover proof-core rollback, mapped and immediate-fanin cuts,
complemented edges, duplicate nodes, interleaved inputs, VTA startup, and
counterexample handling. These checks cover small examples; they are not a
proof of correctness for arbitrary networks or all resource-limit settings.

The `Gia_ManPerformGla` and `Gia_ManPerformGlaOld` entry points lower native XOR and MUX nodes into
AND-inverter cones before running the original AND-only refinement. It maps
incoming gate classes to the lowered graph, conservatively maps selected cones
back to original gate classes, and verifies any witness against the original
graph. Old GLA also transfers seeded VTA frame classes to the lowered graph
and consumes them when VTA starts. Gate abstraction reconstruction preserves
native XOR and MUX nodes.
The lower-level `Gla_ManStart` and `Gla_ManStart2` interfaces still require an
AND-only graph; replacing only their SAT clauses would leave refinement unsound.
Old GLA's source `Gia_ManToAigSimple` conversion asserts object-ID identity;
duplicate or simplified AND nodes can violate that assertion. The Go path
preserves the assertion rather than substituting local cuts. The optional
native-gate lowering can itself create a graph that violates this boundary;
lowering therefore does not guarantee that old GLA accepts every native graph.
Flop-to-gate class conversion visits the control as well as both data inputs
of native MUX nodes; a regression covers the previously omitted control cone.

`Abs_Par_t` and `Gia_ParAbs_t` retain private source-style option fields.
External callers can configure them through `Options` and `LegacyOptions`,
using the corresponding default, setter, and progress methods.

The old GLA `fUseFullProof` option does not select a separate proof-logging
mode in the Go port. In C, full mode allocates `pPrf1` at manager creation;
default mode starts the incremental `pPrf2` manager after the first spurious
counterexample, registers each newly added abstraction object, grows the
manager for further refinements, and stops it when the solve ends. Go's
`LabeledSolver2` always creates `pPrf1` and filters its root labels at the
bookmark. Both Go option settings are tested through spurious refinement and
rollback, but this does not establish identical C proof cores, abstraction
classes, memory use, runtime, or resource-limit outcomes. Implementing the
default mode requires the incremental manager lifecycle and object-ID mapping
to be wired through the solver and old GLA; the difference remains open.

Standalone VTA retains the C algorithm's AND-inverter representation
assumption. Native XOR/MUX nodes are explicitly rejected with an unknown
result and the existing object-class map unchanged; callers can lower those
nodes first. Direct VTA frame-map projection through special-gate lowering is
not yet established, so this guard avoids producing classes under incorrect
AND semantics.

The legacy ternary counterexample verifier retains a separate C boundary
behavior: its initial-state assertion repeatedly checks bit zero, rather than
checking each register bit. Later nonzero initial bits can therefore pass that
assertion, although the verifier initializes every register to zero. A
source-derived test records this limitation.

Support and dominator analysis visits all three operands of native MUX nodes;
flop-to-gate class conversion likewise includes the MUX control cone. Other
reparameterization routines inherited from the C AND-inverter implementation
have not been established as native XOR/MUX aware and require separate review.

VTA retains two C source boundary behaviors that callers must account for.
The source asserts that the object count fits in the object-ID mask, which
rejects a power-of-two object count even though the largest ID fits. On a run
that returns without a new counterexample, including trivial UNSAT, C also
leaves any previously stored sequential witness in place. Tests characterize
both behaviors; callers should clear old witnesses before invoking VTA and
should not infer a new witness from a non-SAT return.

With the pthread-enabled C prover, cancellation increments the run ID but does
not clear an already published proof; a superseded worker can also publish a
successful result after cancellation. The Go port preserves that behavior.
As a result, callers must not interpret a proof returned after cancellation as
necessarily belonging to the current abstraction. The compiled C check for a
published proof followed by cancellation returns `1` once, then `0`; Go tests
also cover publication by a superseded worker.

The modern GLA refinement loop mirrors C's root-assignment shortcut: it stops
when the output variable is assigned, even if the output literal is true. A
source-derived two-register fixture reaches this case after layer refinement;
C and Go return unknown without a witness, although the initial state violates
the property. The Go result therefore retains this source false-negative.

Modern GLA also asserts that a nonzero target SAT literal is greater than one.
A register whose next state is constant one reaches literal 1 at a later frame
and aborts in C; the Go port retains that assertion as a panic.

The GLA boundary fixtures are `test_vectors/gla_true_root_oracle.c` and
`test_vectors/gla_const_later_oracle.c`. The optional environment variables
`ABS_GLA_ROOT_ORACLE` and `ABS_GLA_CONST_ORACLE` run their compiled binaries
in the Go tests. The checked results above used a freshly compiled
`src/proof/abs/absGla.c` linked with the remaining ABC object files; those
support objects must match the checkout.
