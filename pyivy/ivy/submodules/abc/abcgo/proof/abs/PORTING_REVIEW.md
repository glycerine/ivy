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

The `Gia_ManPerformGlaOld` entry point lowers native XOR and MUX nodes into
AND-inverter cones before running the original AND-only refinement. It maps
incoming gate classes to the lowered graph, conservatively maps selected cones
back to original gate classes, and verifies any witness against the original
graph. Gate abstraction reconstruction also preserves native XOR and MUX nodes.
The lower-level `Gla_ManStart` and `Gla_ManStart2` interfaces still require an
AND-only graph; replacing only their SAT clauses would leave refinement unsound.
The port also supports reordered inputs, duplicate AND nodes, and buffers.

`Abs_Par_t` retains private source-style option fields, so callers outside this
package cannot currently configure most options directly. The C-style exported
entry points remain usable with default parameters; a Go-facing configuration
API remains future work.

The old GLA `fUseFullProof` option does not select a separate proof-logging
mode in the Go port: `LabeledSolver2` always creates a proof graph. The
resulting proof cores are tested, but memory and runtime behavior can differ
from the C default. Alternate proof modes and their resource behavior remain
an open fidelity and coverage item.
