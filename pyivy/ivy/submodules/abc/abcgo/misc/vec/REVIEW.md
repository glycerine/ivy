# Vector port review

## Pass 1 — representation and readability

Completed independently of the subsequent correctness review.

- Int, float, pointer, word, byte and integer-level vectors now have a single
  slice whose length is the logical size and whose capacity is the allocation.
  Removed the original cached `nSize` and `nCap` fields. The private `storage`
  accessor is used only where a capacity view is needed for fills/reservations.
- Pointer attribute capacity is the attribute slice length. Queue size follows
  heap length, and its member-order table supplies the supported member range.
- Packed bit vectors retain a bit count because the last word may be partially
  used; their capacity follows the word slice. Paged memory retains entry width,
  page geometry and live entry count because the final page may be partly used.
  Removed cached page count, pointer-array capacity and page mask.
- Level marks are explicit booleans rather than bits mixed into capacities.
- Existing `Vec_*` entry points remain the compatibility API used by other ported
  packages; they no longer require C allocation wrappers or pointer arithmetic.
- Delegated Hsh review also replaced packed unions with stable typed objects,
  removed redundant variable-record size fields, and documented borrowed views.
  Set live/shadow positions are semantic arena state; its records use typed
  pages and explicit byte/word packing.

## Pass 2 — correctness and regressions

Completed after the representation pass.

Fixed and regression-tested:

- Failed integer/word storage reservations return nil without advancing the
  live length. This deliberately fixes C's invalid size-greater-than-capacity
  outcome and preserves valid slice invariants.
- Released arrays contain only live elements, retaining their allocation as
  slice capacity without exposing unused elements as logical entries.
- Pointer shrink/pop/clear and compaction clear dead references so unused
  capacity does not keep objects alive.
- Empty pointer uniquification produces an empty counts vector.
- Typed nil attribute values are treated as absent, matching C null semantics.
- Variable-integer serialization rejects negative inputs and oversized encoded
  values rather than producing malformed data or silently truncating bits.
- Delegated Hsh review corrected large Go table-size truncation. Set review
  added overflow checks for entry and string byte-to-word rounding.

Existing semantic tests additionally cover scalar growth and ownership transfer,
32-bit wraparound, unsigned aliased reservations, record interning and sorted
merge/split, simulation truth tables, word hex roundtrips, binary byte-vector
serialization, partial-word bit counts, level marks, priority updates and stable
paged memory records. Source functions that were originally unimplemented
(`Vec_MemDup`, `Vec_MemFill`, `Vec_MemClean`) are implemented and identified in
`vecMem.go`; source searches found no external C callers of these three helpers.

A subsequent independent container audit also found optional memory-hash indices
could become stale after explicit record writes/zeroing. Writes, cleaning and raw
appends now rebuild the index as needed without changing record IDs; duplicate
records created by mutation resolve deterministically to their lowest ID. Hash
allocation indexes existing records. Regression tests cover each path. Raw record
slices remain borrowed mutable views; callers editing those views directly must
rehash before lookup, as with the source API's mutable storage.
