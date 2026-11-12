# Binary BLIF

This is a standard-library Go port of ABC's standalone binary BLIF interface.
Objects are typed Go records with ordered `Fanins` slices instead of relative
in-memory pointers and packed bitfields. Object/SOP iteration follows insertion
order; original user IDs may be sparse and fanins may refer to later objects.
Functions are deduplicated by canonical SOP strings, with C-compatible byte handles.

The original BBLIF format is supported explicitly as little-endian 32-bit integers:
three length-prefixed pools hold the design name, object records and SOP records.
Objects contain ID, SOP byte handle, packed flags/fanin count, and signed relative
fanin byte offsets. Go machine layout and pointers never enter the file. Files
from the original C implementation on little-endian platforms interoperate in
both directions. Binary readers validate pool sizes, strings, object/function
handles, relative fanins and network consistency before returning a manager.
After reading, `ID` is a sequential index and `OriginalID` retains the user ID.
Rewriting preserves original IDs (unlike the C reader's destructive renumbering).

API differences:

- `Bbl_ManCreateObject` and `Bbl_ManAddFanin` return errors for invalid operations.
  The SOP argument is a string; use the empty string for CI/CO objects.
- `Bbl_ManReadBinaryBlif` returns `(*Bbl_Man_t, error)`; file-writing APIs return
  errors. `ParseBinary`, `WriteBinary(io.Writer)`, and `WriteBLIF(io.Writer)` also
  support in-memory use. Callers should check errors.
- `Check()` returns an explanatory error; `Bbl_ManCheck` retains the C 0/1 result.
  Mixed-polarity or malformed SOPs are rejected rather than interpreted loosely.
- Truth tables are owned `[]uint32` values; SOP conversion returns owned strings.
  Tables smaller than five inputs repeat their function throughout the 32-bit
  word, matching C. At most 16 inputs are supported by truth/SOP conversion.
- `Objects` and `Fanins` provide slice iteration, retaining the C first/next APIs
  for adapters. For repeated fanins, prefer slice iteration: C's pointer-based
  next-fanin API cannot distinguish repeated occurrences of the same fanin.
- `Bbl_ManSimpleDemo` returns its half-adder manager instead of writing fixed
  filenames. `Bbl_ManStop` releases references; Go garbage collection owns storage.

Tests include binary/text roundtrips, sparse IDs, forward edges, function sharing,
markers, malformed input, every truth table through four inputs, random functions
through 16 inputs, and I/O failures. A C-reference test temporarily compiles the
original source when `cc` is available and checks interoperability in both directions.
The package itself has no C compiler, unsafe, cgo or third-party dependency.
