# Abstraction port review

The `proof/abs` package is translated and builds. Its current test suite runs
against the real VTA, GLA, old GLA, SAT, BDD, PDR, and optional SSW prover
components. Package race tests and `go vet` are required after changes.

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
The port also supports reordered inputs, duplicate AND nodes, and buffers.
Flop-to-gate class conversion visits the control as well as both data inputs
of native MUX nodes; a regression covers the previously omitted control cone.

`Abs_Par_t` and `Gia_ParAbs_t` retain private source-style option fields.
External callers can configure them through `Options` and `LegacyOptions`,
using the corresponding default, setter, and progress methods.

The old GLA `fUseFullProof` option does not select a separate proof-logging
mode in the Go port: `LabeledSolver2` always creates a proof graph. The
resulting proof cores are tested, but memory and runtime behavior can differ
from the C default. Alternate proof modes and their resource behavior remain
an open fidelity and coverage item.

Standalone VTA retains the C algorithm's AND-inverter representation
assumption. Native XOR/MUX nodes are explicitly rejected with an unknown
result and the existing object-class map unchanged; callers can lower those
nodes first. Direct VTA frame-map projection through special-gate lowering is
not yet established, so this guard avoids producing classes under incorrect
AND semantics.

Support and dominator analysis visits all three operands of native MUX nodes;
flop-to-gate class conversion likewise includes the MUX control cone. Other
reparameterization routines inherited from the C AND-inverter implementation
have not been established as native XOR/MUX aware and require separate review.
