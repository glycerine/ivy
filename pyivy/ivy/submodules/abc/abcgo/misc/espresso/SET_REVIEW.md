# Packed-set foundation review

The set layer preserves Espresso's BPI=32 representation: a metadata word,
followed by packed uint32 data words. LOOP uses the low ten metadata bits, flags
use the original PRIME/ACTIVE/etc. positions, and SIZE occupies the high half.
A family retains contiguous storage and its original count/capacity/width fields.
Sets are slices, and family/cube/cover names retain their C relationships.

All original set.c functions are implemented, including weighted restricted
column counts, serialization, active-row compaction, consuming append/column
operations, permutation, compression and matrix transpose. NewSet/NewFamily and
Set/Family methods make the same representation usable by other Go packages.
INLINEset_* call sites should invoke the corresponding direct set_* operation;
control-flow iterator macros must become Go loops, not helper functions.

## Compactness / idiomatic review

- Integer bit-count and trailing-zero operations use math/bits.
- Slice copy handles overlapping sets safely; rows have bounded capacities.
- Family descriptors retain the source's recycling service, protected by a
  mutex, while Go owns their backing storage. sf_cleanup clears that cache.
- Set deallocation requires no manual allocator; set_free documents the
  garbage-collected ownership rather than simulating a C allocator.
- I/O uses io.Reader/io.Writer and formatting/scanning; strings own their
  storage rather than sharing the source's static 120-byte print buffer.
- Sparse-variable cleanup preserves the original cofactor-to-row mapping and
  calls the original irredundancy/expansion stages, rather than replacing them
  with a different minimizer.

## Separate bug review

- Verified metadata/flags survive binary set operations and full metadata is
  copied by set_copy/save.
- Verified aliasing (destination equals either input), zero-sized sets, word
  boundaries, family growth and active-row order.
- Preserved sf_copy_col's OR-only behavior: a zero in the source does not clear
  an existing destination bit.
- Removed newly introduced scanner read-ahead that could consume the following
  serialized family. Sequential-family tests use a reader without RuneScanner.
- Preserved sf_bm_print's original indexed-row format; it intentionally differs
  from sf_bm_read's compact numeric-header matrix format.
- Checked column insertion/deletion consumes only the source family, and row
  storage remains owned by the returned family after append.
- Invalid packed sizes or serialized widths fail explicitly instead of
  overflowing metadata or indexing outside allocated storage.
- Restricted-count rows of cardinality one explicitly panic: the source's
  formula divides by zero for those inputs and its callers must remove singleton
  constraints first. No different weight is silently invented.

The sparse cleanup still uses the package's original global cube configuration.
Concurrent minimization needs the package-level configuration/ownership policy;
independent set primitives do not share scratch storage.

## Differential fixture and validation

`testdata/set_oracle.c` compiles the actual original src/misc/espresso/set.c,
with the original 32-bit packed layout and minimal allocation/platform macros.
It exercises set metadata/algebra, cardinality/intersection/first-bit results,
serialization, column insertion/deletion and transpose. The checked Go fixture
matches its output byte for byte. If a C compiler is available the test recompiles
and reruns that independent oracle; Go runtime code has no compiler dependency.

Focused validation while other Espresso families are being ported:

    go test -race ./misc/espresso/set.go ./misc/espresso/set_types.go ./misc/espresso/set_test.go

This passes, including the original-C oracle. This focused result is not a claim
that the whole Espresso package is complete. make_sparse/mv_reduce integration
also requires the real cube/cost/timing, mark_irredundant and expand families.
