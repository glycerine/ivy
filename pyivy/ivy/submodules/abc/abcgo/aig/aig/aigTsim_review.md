# Ternary simulation review

Ported original aigTsim.c into aigTsim.go using the shared Graph/Edge model.
Node.Xsim holds valid 0/1/x encodings (1/2/3); complemented edges invert known
values and preserve x. Recursive per-frame evaluation supports explicit outputs,
buffers, XORs, non-topological rewrites and cycle diagnostics. All latch transfers
are simultaneous. Packed two-bit latch states include the all-zero initial state.
The original round cap, exact-series phase, retirement-to-x phase, repeated-state
termination, history union and constant-register input mapping remain intact.

Aig_ManConstReduce uses typed remapping and a sequential reachability fixed point,
retaining true inputs/outputs and next-state logic of live latches. It preserves
metadata and filters original flop numbers. Optional symbolic/multivalued engines
are injected as MappingSimulator to avoid an aig/saig cycle; requesting that mode
without a nonnil engine is an explicit error. Ternary mode needs no injection.

## Readability/compactness pass

Replaced the C fixed allocator, embedded next-pointer arithmetic, manual hash bins
and duplicate size fields with typed state slices and immutable map keys. State
width is derived from the graph register count. State hash remains available for
legacy callers, but simulation needs no bin-count or allocation-capacity fields.
State printing accepts io.Writer; raw output annotations and duplicated C macro
scaffolding are gone. Existing core duplication helpers centralize metadata and
edge phase handling.

## Separate bug pass

Checked simultaneous updates, inclusion of initial latch zeros, partial final
state words, zero-register graphs, unknown primary inputs, complemented signals,
explicit buffers/XOR/outputs, malformed states and combinational cycles. Tests
cover immutable map-key ownership even when caller state storage mutates.

Review findings fixed: reject a nil injected engine before invocation; reject a
non-progressing mapping rather than looping forever; reject inconsistent flop-
number metadata instead of silently discarding it during register filtering.

Focused actual-source core+Tsim tests and race tests pass. This is a scoped test
selection, not proof that the complete aig package builds; its unrelated draft
files and external package imports remain under parallel conversion.
