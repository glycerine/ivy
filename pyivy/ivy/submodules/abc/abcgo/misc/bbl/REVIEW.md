# BBL review

## Design/readability pass

Reviewed construction, traversal, function sharing, truth conversion, binary/text
I/O and public compatibility APIs separately from the bug pass. Typed objects and
ordered fanin slices replace private C vector/pointer pools without duplicating
raw memory representations. The binary layout is encoded explicitly with
encoding/binary. io.Reader-style file/error ownership is explicit; WriteBinary
and WriteBLIF support standard io.Writer values. SOP validation is centralized,
and truth expansion enumerates cube don't-care assignments instead of scanning
every cube for every truth bit. Public C spellings remain for surrounding ABC
adapters; API/ownership differences are documented in README.

## Bug-review pass

Reviewed truncation/length bounds, signed relative offsets, original/sequential
IDs, truth bit order and small-table stretching, write failures, function handles
and canonicalization. Found and fixed a reserialization defect for valid imported
SOP records with extra padding words: canonical output shortened their records
but previously retained stale object function handles. WriteBinary now rebuilds
wire handles from the actual emitted pool. TestPaddedFunctionRecordsReserialize
was observed failing before the fix and now passes.

Original-C interoperability tests compile the repository implementation in a
temporary directory, read its half-adder in Go, then read Go's binary output in C
and verify parity/majority truth tables and object count. Production has no C
compiler/runtime requirement.

Validation: `go -C abcgo test ./misc/bbl -count=1` and vet pass. Tests include
exhaustive truth tables through four inputs, random tables through 16 inputs,
constant/off-set SOPs, sparse IDs/forward edges, function sharing, marks,
roundtrips, malformed/truncated records, padding and short-write errors.
No pending review finding remains in the implemented package. C's little-endian
32-bit wire-layout restriction and repeated-fanin pointer-iterator ambiguity are
explicitly documented.
