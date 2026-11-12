# FPGA mapper port and review

The package ports the active algorithms from `src/map/fpga`: structural hashing,
complemented signals, choices, feasible-cut merging/deduplication/trimming,
delay and area-flow matching, exact reference-counted area and switching
recovery, arrival/required propagation, LUT libraries, vectors, traversal and
library commands. `fpgaGENERIC.c` is empty. The C `#if 0` experiments and disabled
cut-fanout mutation code are not executable source APIs and are not recreated.
Header-only declarations with no C implementation are not advertised as working
features. All package Go files compile normally; no source is hidden by tags.

## Readability and compactness pass

Performed independently of the bug pass. Native slices replace redundant
size/capacity fields and fixed scratch arrays; Go maps replace pointer hashes.
Complemented edges use GC-safe cached proxies. Leaf sets are sorted by stable
node IDs. Each cut owns its leaf-relative `uint64` truth table, removing the
source's pointer-packed truth/signature scratch. Source function names remain
available for callers. `NewManager`, `Map`, `Evaluate`, `ReadLibrary` and library
`Write` provide error-returning native entry points. `BooleanBuilder` permits
real Hop/BDD/AIG construction without importing a BDD implementation.

Refactored area recovery into the public postprocessing routine rather than
leaving a second abbreviated implementation. Kept source vector Union's
concatenation semantics explicit. Levelization retains linked heads per level,
rather than returning an unrelated sorted vector. Structural fanouts and mapped
references have separate storage; public cost queries restore reference counts.

## Separate bug pass

Reviewed graph phases, choice leaf normalization, repeated runs, reference
transitions, truth widths, timing pin assignments, and public traversal behavior.
Fixed and covered:

- A secondary choice's trivial leaf must refer to its representative, with the
  appropriate truth phase. Copying a secondary trivial cut would create a cycle.
- Complemented cut handles must complement the constructed Boolean function.
- Mapping reruns restore structural references and reset required times/statistics.
- Limited DFS counts recursion depth and returns postorder, not an absolute
  level boundary or preorder. Level lists retain the source linkage contract.
- Exact recovery handles unused nodes without dereferencing an inactive cover.
- Negative finite PI arrival times remain valid; maxima must start below zero.
- Cancelled runs, unsupported cut widths, missing arrival arrays, invalid library
  numbers and unset/foreign outputs return errors at native entry points.
- A failed library load preserves the current frame library.

The original mapper explicitly ignores variable pin delays. This port implements
optimal latest-arrival/fastest-pin assignment and propagates the corresponding
required times. `PinOrder` records LUT-pin-to-truth-leaf order; truth table leaf
order never changes. Equal-delay libraries retain source behavior.

The C library parser permits LUT sizes through ten, but C cut storage has only
six leaves (the source comments say seven does not work). Libraries can be read
through ten; manager construction explicitly rejects mapping widths above six.
Constant-only and zero-output covers have area/delay zero instead of the C
maximum-arrival sentinel. Choice equivalence remains a caller precondition,
as in the source.

## Verification

- Exhaustive original graph vs mapped-cover truth checks across deterministic
  random six-input graphs, LUT widths two through six, both area and switching
  modes, complemented outputs, constants, sharing and repeated mapping.
- Choice alternatives, secondary fanout consumption and opposite-phase choices.
- Known LUT cover size/depth, nonuniform pin timing, relaxed required targets,
  reference-cost restoration, bounded DFS/TFO/level lists, hash rebuilding,
  library round trips/ownership/errors, commands and BooleanBuilder via real Hop.
- `test_vectors/reference.c` links the unchanged original FPGA C mapping files
  and memory manager. Only absent host-frame/progress/logging services are
  supplied locally. Ten native-vs-C cases compare exact area, delay and LUT count
  for chains and reconvergent XOR/MUX logic at widths two through six. The
  oracle allocates PI arrival storage, which original callers must supply.
- Normal package tests, race tests with coverage, and `go vet ./map/fpga` pass.

The oracle uses `cc` when available and is skipped only when no C compiler is
installed. All artifacts are created in the test temporary directory.
