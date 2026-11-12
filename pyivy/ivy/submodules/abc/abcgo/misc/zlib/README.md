# Go compression replacement

This package replaces the generated zlib 1.2.5 C drafts with Go standard-library
implementations. It uses `compress/zlib`, `compress/gzip`, `compress/flate`,
`hash/adler32`, and `hash/crc32`. It has no C compiler, cgo, assembly, or external
package dependency. The original zlib notice is retained in LICENSE. This is an
altered Go replacement and is not the original zlib source distribution.

## API mapping

| Original service | Go API |
| --- | --- |
| zlib streaming compression | `NewWriter`, `NewWriterLevel`, `NewWriterLevelDict` |
| zlib streaming decompression | `NewReader`, `NewReaderDict`; `Resetter` |
| raw DEFLATE / gzip framing | `NewFormatWriter`, `NewFormatReader` |
| automatic zlib/gzip detection | `NewFormatReader(r, Auto, dict)` |
| buffer compress/uncompress | `Compress`, `Compress2`, `Uncompress`, `CompressBound` |
| allocating one-shot helpers | `CompressBytes`, `UncompressBytes` |
| gzopen/gzdopen and owned handles | `Open`, `OpenFile`, `Gzopen`, `Gzdopen`; `File` / `GzFile` |
| gzread/gzwrite/gzprintf/gzclose | `Gzread`, `Gzwrite`, `Gzprintf`, `Gzclose`; ordinary `Read`, `Write`, `Close` |
| gzputs/gzgets/gzgetc/gzputc/gzungetc | corresponding exported `Gz...` functions |
| gzseek/gzrewind/gztell/gzoffset | `Gzseek`, `Gzrewind`, `Gztell`, `Gzoffset`; `File.Seek` |
| gzbuffer/gzsetparams/gzflush | corresponding exported functions |
| gzeof/gzdirect/gzerror/gzclearerr | corresponding exported functions |
| adler32/crc32 and combine functions | `Adler32`, `CRC32`, `Adler32Combine`, `CRC32Combine` |
| get_crc_table / zError / zlibVersion | `GetCRCTable`, `ZError`, `ZlibVersion` |

Buffer APIs use slices instead of pointer-plus-source-length arguments, a
`*uint64` destination length, and the original `Z_*` status constants. A gzip
handle implements `io.Reader`, `io.Writer`, `io.Seeker`, and `io.Closer`.
`OpenFile` and `Gzdopen` transfer ownership of the underlying descriptor;
`NewReader` and `NewWriter` do not close their underlying Go streams.

Reading gzip handles transparently accepts ordinary uncompressed files,
including files whose names end in `.gz`. Writing emits gzip regardless of file
extension, matching the original zlib implementation. Explicit `T` mode writes
uncompressed bytes. Concatenated gzip members and append mode are supported;
trailing non-gzip garbage after a member is ignored. Checksums and truncated
compressed input produce errors. Rewind/backward seeking replays decompression
from the descriptor's initial position; nonseekable descriptors remain readable.
Forward writes fill skipped output positions with zero bytes.

`Gzprintf` translates C integer length modifiers and `%u`/`%i`, including unsigned
integer truncation/extension. C string arguments stop at NUL. Unsupported `%n`,
hex-float `%a`/`%A`, and wide-character formats return a recorded error rather
than performing a misleading conversion. `Gzputs` and `Gzgets` preserve C NUL
termination rules; ordinary `Write` and `Read` are binary interfaces.

## Explicit compatibility limits

Go exposes stream readers/writers rather than zlib's mutable `z_stream` structure.
There is no fabricated implementation of `deflate`/`inflate` with caller-managed
`next_in`, `avail_in`, `next_out`, `avail_out`, or progress/bit counters. Streaming
users migrate to the reader/writer APIs above, including preset dictionaries,
flush, and reset. The callback-driven `inflateBack` service maps to a raw reader
and `io.Copy` to a caller-supplied writer.

The following low-level C controls have no equivalent exposed by the Go backend
and are deliberately tracked here rather than silently dropped or stubbed:

- `deflateCopy`, `inflateCopy`: cloning internal compressor/decompressor state.
- `deflateTune`, `deflatePrime`, `inflatePrime`, `inflateMark`,
  `inflateSyncPoint`, `inflateUndermine`: internal tuning, bit insertion and
  decoding-position/recovery controls.
- `inflateSync`: recovering at full-flush block markers after corrupt input.
- `deflateInit2` memory-level/strategy/window-size tuning: Go selects its own
  memory layout and 32 KiB window. Formats are selectable, but filtered, RLE and
  fixed-tree strategies are not selectable; gzip modes `f`, `R`, `F` return errors.
- `Z_PARTIAL_FLUSH` and `Z_FULL_FLUSH`: `Gzflush` returns `Z_STREAM_ERROR` because
  Go cannot promise their dictionary/block-boundary semantics. No-flush, sync
  flush, and finish are supported. Finishing then writing starts another member.
- Midstream dictionary installation after a `Z_NEED_DICT` return: supply the
  dictionary when creating/resetting the reader instead.
- C `gz_header` callbacks: use the standard `gzip.Reader`/`gzip.Writer` Header
  fields directly when header customization or inspection is required.
- `zlibCompileFlags`: C ABI/platform flags do not describe the Go implementation.
- `gzseek` past the available decoded input: this Go adapter resolves seeks
  immediately and reports EOF, instead of deferring the skip to the next read.
- Midstream `Gzsetparams` starts a new gzip member to apply the changed level;
  it preserves decoded bytes but does not promise original compressor block layout.

`Z_HUFFMAN_ONLY` is the C strategy value 2. The separate Go compression level
`HuffmanOnly` is `flate.HuffmanOnly` (-2); these are intentionally distinct.
Independent handles can be used concurrently. One mutable gzip handle is used
by a single goroutine at a time, consistent with the original gzFile contract.

## Consumer audit

The actual ABC source consumers outside bundled zlib use gzip file I/O:

- `src/base/io/ioReadAiger.c` and `ioReadBlifMv.c`: `gzopen/read/close`.
- `src/base/io/ioWriteAiger.c`: binary writes and formatted gzip output.
- `src/sat/cnf/cnfMan.c`: `gzprintf` CNF emission.
- `src/sat/{bsat2,glucose,glucose2}/ParseUtils.h`: buffered `gzread`.
- `src/sat/bsat2/Main{Sat,Simp}.cpp`: `gzdopen(0, "rb")` ownership.

No ABC consumer of the C low-level deflate/inflate/prime/tune APIs was found.
The standard streaming interfaces are nevertheless tested for all three formats,
including incremental pipe reads before the writer finishes.

## Separate reviews and verification

Readability/idiomatic review: public streaming services are direct standard
interfaces; only legacy buffer/gzip calls have thin adapters. File ownership is
explicit, checksums are grouped independently from handle state, and compression
levels are separated from strategies. Unneeded generated deflate/Huffman-table
internals were replaced by the maintained Go backend rather than retained as
noncompiling files.

Bug review: checked nil/zero-length handles, signed/unsigned C formatting,
checksum failures, EOF versus errors, append/member boundaries, pushback logical
positions, seek behavior, descriptor ownership, and bounded buffer status returns.
Fixed and regression-tested unwanted empty members after finish, invalid
zero-length handle acceptance, mode/flush validation, and the Huffman
strategy/level constant distinction.

Validation: `go test -race ./misc/zlib` and `go vet ./misc/zlib`. Tests cover
standard-library interoperability, all compression levels, short destination
buffers, invalid headers/checksums, incremental and combined checksums,
dictionaries/reset, zlib/gzip/raw/auto streaming, CNF-style unsigned formatting,
plain input, multiple members, append, binary I/O, seek/rewind/ungetc, pipes and
owned descriptors, corruption, and unsupported requests returning explicit errors.
