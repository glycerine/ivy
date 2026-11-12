# Go bzip2 services

This package replaces ABC's generated libbzip2 drafts with Go implementations.
Decoding uses `compress/bzip2`. Go has no standard-library bzip2 encoder, so the
writer implements the standard format directly using Go's standard library.
There is no cgo, subprocess, C compiler, or third-party package dependency at
runtime. The original libbzip2 license is retained in LICENSE; this is an altered
Go replacement, not the original libbzip2 implementation.

## Actual ABC consumers

`src/base/io/ioReadAiger.c` and `ioReadBlifMv.c` use `BZ2_bzReadOpen`, `BZ2_bzRead`,
and `BZ2_bzReadClose`. Their ordinary-file fallback depends on
`BZ_DATA_ERROR_MAGIC`; this status is preserved even for one-byte plain files.

`src/base/io/ioWriteAiger.c` uses `BZ2_bzWriteOpen`, `BZ2_bzWrite`, and
`BZ2_bzWriteClose` with block size 9 to write binary AIGER data. This path has a
real bzip2 encoder, not a stub or an external executable adapter.

## Interfaces

- Idiomatic streaming: `NewReader`, `NewWriter`, `NewWriterLevel`; writers support
  `Write`, `Flush`, and `Close`. Closing these streams does not close the borrowed
  underlying Go reader/writer.
- ABC file APIs: the `BZ2_bzRead*` and `BZ2_bzWrite*` families, including
  `BZ2_bzWriteClose64`, keep the original names and status constants. `*File`
  (`BZFILE`) implements `io.Reader`, `io.Writer`, and `io.Closer`.
- Open helpers: `Open`, `BZ2_bzopen`, `BZ2_bzdopen`, plus the simple
  `BZ2_bzread`, `BZ2_bzwrite`, `BZ2_bzflush`, `BZ2_bzclose`, and `BZ2_bzerror`
  functions. Open helpers own descriptors; ReadOpen/WriteOpen borrow them.
- Buffer APIs: `BZ2_bzBuffToBuffCompress` and `BZ2_bzBuffToBuffDecompress`, using
  Go slices and `*uint32` destination lengths. Short destinations return
  `BZ_OUTBUFF_FULL`; compression writes directly into the bounded destination.
- Allocating helpers: `CompressBytes`, `DecompressBytes`.

Scalar C integers are `int32`; sizes and 32-bit count outputs are `uint32`.
WriteClose64 reports low/high halves of full 64-bit input/output counts. Optional
count pointers can be nil. Input prefix bytes passed to ReadOpen are copied, as
in C. The small/verbosity/work-factor parameters are validated in their original
ranges; see the performance-option differences below.

## Encoder design

The writer emits ordinary non-randomized `BZh1` through `BZh9` streams:

1. First-stage byte run-length encoding, with runs capped at 259 bytes.
2. Cyclic Burrows-Wheeler transform using counting-sort prefix doubling, with
   linear storage and O(n log n) runtime.
3. Move-to-front coding and the bzip2 RUNA/RUNB encoding of front-symbol runs.
4. Frequency-weighted, complete canonical Huffman codes bounded to 20 bits.
   Two identical trees and ordinary selectors keep the implementation compact.
5. Per-block CRC-32/BZIP2 and the rotated combined stream CRC.

Block storage is bounded by the requested 100 kB block-size multiple. The raw
input chunk limit accounts for worst-case expansion of first-stage RLE. Writes
may emit multiple blocks; Flush emits a block without ending the stream; Close
writes the end marker, combined checksum and final byte padding. Input/output
counts refer to real accepted/emitted bytes. Abandon releases pending state
without producing a finishing trailer.

Compressed bytes, compression ratios and block partitioning need not match
libbzip2's optimized multi-tree encoder. Callers must handle `BZ_OUTBUFF_FULL`
instead of assuming a particular implementation's recommended output-size rule.

## Explicit scope differences

- Go's standard decoder reads all concatenated bzip2 streams. It does not expose
  a stop-after-first-member option. Consequently ReadGetUnused describes buffered
  input after all members, rather than C's unused bytes after the first member.
  Arbitrary non-bzip2 trailing data can produce a continuation-header error.
- Go's decoder rejects obsolete randomized bzip2 blocks. Current non-randomized
  streams from ABC/libbzip2 are supported, and this writer never emits randomized
  blocks.
- The C mutable `bz_stream` structure, caller-managed avail-in/avail-out fields,
  and custom allocation callbacks are replaced by Go `io.Reader`/`io.Writer`
  streaming interfaces and garbage-collected storage. There are no fabricated
  `BZ2_bzCompressInit`/`BZ2_bzCompress`/`BZ2_bzCompressEnd` or corresponding
  decompression-state functions.
- `small=1` does not select libbzip2's alternative low-memory decoder. The Go
  standard decoder controls its own workspace. Verbosity does not print C debug
  statistics; workFactor does not tune the deterministic prefix-doubling sorter.
- The original legacy `BZ2_bzflush` was a no-op. This adapter performs the writer's
  actual block flush. Buffered partial output bits remain associated with the
  stream; this is not a byte-alignment or new-member operation.
- ReadOpen/WriteOpen return errors from their Go underlying streams. The encoder
  writes its four-byte header at initialization, so an output error can be
  reported earlier than in a lazily initialized C encoder.

None of these differences is needed by the audited ABC consumers. They are
recorded explicitly so a broader libbzip2 API claim is not inferred from package
build success.

## Separate reviews

Compactness/readability review: streaming and bounded-buffer adapters share the
same real encoder; checksum, cyclic sorting, run-length/MTF and Huffman stages
are separate helpers. Counting-sort state belongs to each block, reusable output
buffering avoids repeated allocations, and file ownership is explicit.

Bug review: checked format completeness, canonical Huffman Kraft equality,
periodic rotations, RLE expansion and block limits, CRC accounting, bit alignment
across Flush, nil/invalid handles, error queries, byte-count halves, borrowed
files, and input-prefix lifetimes. Fixed and regression-tested incomplete Huffman
trees, short-plain-input magic classification, copied unused-prefix storage,
parameter-error status queries, and reuse after a recoverable parameter error.

## Verification

`go test -race ./misc/bzlib` and `go vet ./misc/bzlib` pass. Tests cover empty and
one-byte streams, repeated/periodic text, random binary data, long runs, RLE
expansion, multiple blocks, all boundary block sizes, flush/finish, concatenated
members, file/buffer consumer APIs, short output buffers, bad headers, truncated
streams, checksum corruption, counts and ownership, abandoned writes and failing
underlying writers. CRC-32/BZIP2 and the `banana` BWT use independent reference
vectors. Canonical code lengths are checked for completeness and bounds.

An optional test uses the installed `bzip2` program as an independent oracle:
its decoder accepts Go-generated streams, and Go decodes its output. The test
skips if that executable is unavailable. The implementation itself never starts
an external process.
