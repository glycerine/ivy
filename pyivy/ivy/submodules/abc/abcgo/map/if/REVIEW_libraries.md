# LUT and box library review

Scope: `ifLibLut.go`, `ifLibBox.go`, `ifLibraries_test.go`, and the scoped
`test_vectors` fixtures/manifest. Shared IF types, manager/command sources and
cut/truth algorithms remain owned by other workers. No drafts were hidden or
excluded from the package.

## Separate compactness/readability pass

The LUT file/string readers share one line parser and error-aware API. Original
C-named wrappers remain available. Fixed arrays retain the mapper's existing
header layout. Box parsing separates the two source formats while sharing
integer/delay parsing and lifecycle operations. Sparse ID lookup retains the
source vector representation. Immutable Go strings replace owned C strings;
box delay arrays and entire libraries have actual independent deep copies.

Formatting uses supplied `io.Writer` values and propagates errors. Typed
accessors expose timing, attributes and pin layout without exposing storage.
Frame library changes use the supplied framework frame where the original API
accepts one. Companion loading preserves the prior library on failure.

## Separate bug review and corrections

- File/string parsers accept a final line without a trailing newline, comments,
  CRLF and blank lines without running past a C string buffer.
- Validate consecutive LUT sizes, numeric fields, finite nonnegative timing and
  area, exact pin counts, and nondecreasing per-pin delays.
- The maximum LUT row is accepted through `IF_MAX_LUTSIZE` (32); the C reader
  rejected its last valid row after writing it.
- A shared-delay row in a variable-pin library expands to every pin instead of
  leaving zero delays in later pins. Partial variable-pin rows are rejected.
- Discreteness checks every variable-delay pin, fixing the C check of pin zero
  only. Zero delay remains valid under the documented nonnegative rule.
- Float32 output uses sufficient significant digits for exact round trips,
  rather than rounding all area/delay fields to two decimals.
- Box readers reject missing/extra table entries, bad dimensions, duplicate
  IDs/names and invalid/conflicting attributes before installing a library.
- Missing arcs remain negative: classic format uses -1000000000, extended format
  uses -1, matching the source conventions. Writers emit a dash for either.
- The original NULL-returning `If_BoxDup`/`If_LibBoxDup` bodies are replaced by
  real duplication. Library duplication takes a library, correcting its C typo.
- Delay matrix row newlines go to the requested writer, fixing the original
  `printf` to process stdout inside a file-output loop.
- Extended seq/outer attributes are preserved using extended output. Explicit
  classic output rejects attributes it cannot represent rather than losing them.
- Token reads handle EOF comments and transient empty reads without a static
  global buffer. File open/write/close errors propagate through native APIs.
- Companion filenames use the basename's extension; dotted directories are
  preserved. A missing/malformed companion does not replace a valid library.

These validations intentionally turn malformed tables that C merely warned about
or read out of bounds into useful Go errors. The legacy pointer-returning reader
wrappers return nil on failure; native parsing APIs retain the detailed error.

## Fixtures and gates

`test_vectors/library_mixed.lut`, `boxes.classic`, and `boxes.extended` exercise
real file formats, mixed pin delays, sparse IDs, matrix layout, absent arcs and
extended attributes. Regressions cover precision round trips, maximum LUT size,
secondary-pin fractional delays, all simple libraries, deep-copy ownership,
malformed records, writer errors, supplied-frame isolation, companion loading
and preservation after failed replacement.

The combined gate uses the existing IF manager/header/cut/truth/count suite and
its original-C cut oracle alongside the library tests:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 $(cat abcgo/map/if/test_vectors/library-files.txt)
GOCACHE=/tmp/abcgo-cache go -C abcgo vet $(cat abcgo/map/if/test_vectors/library-files.txt)
```

Both gates pass against actual shared source files. Whole `map/if` compilation
still encounters separately owned decomposition drafts (for example ifDec07.go);
this review does not claim completion of those algorithms or the whole package.
