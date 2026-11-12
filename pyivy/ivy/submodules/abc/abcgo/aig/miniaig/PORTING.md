# MiniAIG, MiniLUT and NDR Go port

The four ABC headers under `src/aig/miniaig/` have executable Go implementations
here. All source functions, including NDR's nine example constructors, are
represented. Allocation macros become slices and ordinary allocation; iterator
macros become loops over the original compact record representation.

Literal and object IDs use Go `int`; truth words use `uint32`. Predicates return
`bool`. Pointer arithmetic APIs return slice views. NDR designs are typed
`*Ndr_Data_t`; textual functions are strings and LUT functions are `uint64`,
removing the C convention of placing a truth table in pointer bits. File paths
are strings, and Verilog writers accept standard-library `io.Writer` values.

Binary AIG, LUT and NDR files use ABC's x86 little-endian 32-bit integer layout.
Loaders reject missing, truncated and inconsistent files. Dump write failures
panic instead of being silently ignored. NDR Verilog emission retains the
original operator coverage, including the original explicit cannot-write text
for operations that header's writer does not support.

Tests exercise truth-table synthesis, literal evaluation, graph topology,
AIG/LUT/NDR binary roundtrips, Verilog generation and every NDR example. Run:

```
go test ./aig/miniaig
go test -race ./aig/miniaig
```

Original implementation: Alan Mishchenko, ABC Logic Synthesis and Verification
System, UC Berkeley. Refer to the corresponding C headers for source history.

## Review passes

A separate compactness/readability pass removed copied `nSize`/`nCap` metadata
from all three representations. Slice `len`/`cap` are the single source of truth;
register counts and LUT width remain because they describe circuit semantics.
Removed generated blank-identifier assignments and the unused assert helper.

A subsequent bug pass added failing regressions and fixes for malformed constant
and PI records, MiniLUT sentinel placement, NDR containers crossing their enclosing
module, MiniAIG Verilog fanins that reference PO nodes, and hierarchical NDR
instances losing their output connections. The complete package and race tests
pass after these fixes. No source functionality was excluded to get a build.
