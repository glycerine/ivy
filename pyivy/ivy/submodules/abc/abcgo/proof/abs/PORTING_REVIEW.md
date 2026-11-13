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

Old GLA currently rejects unlowered native XOR and MUX GIA nodes. The C
implementation's simple GIA-to-AIG conversion assumes AND-only internal nodes
and asserts object-ID correspondence. This Go port supports reordered inputs,
duplicate AND nodes, and buffers while preserving object mappings. Supporting
native XOR/MUX safely requires lowering those gates and translating both the
incoming and resulting gate-class maps back to the original objects, or
extending the refinement simulation and justification algorithms. Replacing
only the SAT clauses would leave refinement unsound. Callers must currently
supply an AND-lowered GIA to old GLA.

`Abs_Par_t` retains private source-style option fields, so callers outside this
package cannot currently configure most options directly. The C-style exported
entry points remain usable with default parameters; a Go-facing configuration
API remains future work.
