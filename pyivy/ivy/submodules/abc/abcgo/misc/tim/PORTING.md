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
