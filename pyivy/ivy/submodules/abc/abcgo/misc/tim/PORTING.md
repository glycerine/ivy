# Timing package port

`go test ./misc/tim` builds and tests the complete package. The implementation
uses typed slices for objects, boxes, delay tables and former Vec containers;
Go allocation replaces Mem_Flex storage. There is no cgo or unsafe dependency.

All functions declared by the original tim.h are implemented. Public signatures
use Go `int`, `bool`, `[]int`, `[]float32` and `[][]float32`. Save/load use `[]byte`.
`Tim_ManCreate` accepts a `TimingLibrary` interface with `BoxCount` and `BoxTiming`
methods rather than a void pointer to map/if's box library. An adapter for that
library is an integration task for its port. Manager/object/box fields remain
private, matching their placement in the original private timInt.h.

Version-1 serialization preserves ABC's existing format: big-endian int32 fields
and native-endian float32 payloads. Black-box flags were never serialized by that
version and still are not; reload the box library to recover them. The loader
panics on malformed/truncated data, corresponding to the original assertion
contract. The format is not portable across float byte orders.

Timing arithmetic intentionally repairs three apparent source defects, verified
by tests: arrival propagation preserves float fractions rather than passing them
through Abc_MaxInt; required propagation checks connectivity at the current input
column rather than at the output number; Tim_ManGetReqTimes returns required
rather than arrival times. Reduced managers also renumber copied terminal IDs to
their new positions, so save/load and subsequent timing queries address the right
terminal. These should be included in any C-vs-Go differential review.

Printing preserves the same information but uses a simpler readable Go layout;
byte-for-byte terminal output compatibility is not claimed. Trim/reduce retain
the original convention that PIs precede box outputs and POs follow box inputs.

## Separate audits

The compactness/idiomatic/readability pass consolidated timing sentinels, made
zero-port box boundaries explicit instead of recovering them from a flexible
array, and made the stream decoder validate topology before constructing the
manager. It retains the source API names to keep algorithm review straightforward.

The subsequent bug pass fixed and regression-tested:

- Unit-delay duplication now resets constraints in the clone, not the source.
- Duplicating a manager without delay tables preserves nil, allowing later setup.
- Zero-input, zero-output and zero-port boxes retain their boundary positions
  through duplication and version-1 save/load.
- Failed timing queries no longer mark an object as successfully cached.
- Malformed large terminal counts are rejected before allocating terminal arrays.
- Negative terminal-reduction counts are rejected explicitly.

Full package tests, race tests and vet pass after these changes. The unit-delay
clone correction is an additional intentional source defect repair beyond those
listed above. Save/load still preserves the original version-1 binary format.
Manager duplication and reduction reset traversal marks only in the new manager;
they preserve the source manager's timing constraints and traversal cache state.
