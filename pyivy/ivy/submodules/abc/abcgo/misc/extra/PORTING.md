# misc/extra port status

Completed native Go implementations:

- `extraUtilBitMatrix.go`: all 16 bit-matrix functions, with fixed-width words,
  independently stored upper/lower relations, unions, disjointness and clique checks.
- `extraUtilMemory.go`: all 18 fixed/flexible/stepped allocator functions. Mutable
  entries are byte slices, replacing raw C pointers; free lists hold slices rather
  than writing pointers into entry bytes. Allocation preserves chunk accounting,
  recycling and restart behavior. Callers must use slice indexing for entries.

The package is not buildable yet. Other generated files still contain C syntax,
unknown preprocessor branches and unresolved external types. Nothing is excluded
from builds or replaced with placeholders. Important remaining dependencies include
misc/vec, bool/kit, bdd/cudd and base/main's frame/progress configuration.

Focused verification (does not imply package-wide completion):

```
go test extraUtilBitMatrix.go extraUtilBitMatrix_test.go extraUtilMemory.go extraUtilMemory_test.go
```

Full package verification remains `go test ./misc/extra` from `abcgo/`; currently
it fails first on a retained `#` directive in `extra.go`.

`misc/extra/npn` is an independently buildable native subpackage for four-input
NPN canonicalization. It exports `Extra_Truth4VarNPN`, `Permutations`, and
`Transform`. Canonical/phase/permutation tables match all 262144 bytes emitted
by the original C implementation, with the oracle hash recorded in a test.
Exhaustive tests reconstruct all 65536 truth tables and verify all 222 orbits.
Class IDs are completed for every function, including complemented canonical
representatives whose IDs the original C loop leaves unset.

```
go test ./misc/extra/npn
```

This subpackage lets consumers use canonicalization while the larger parent
package's remaining generated source is being ported.

## Required review passes

Readability/idiom/compactness audit (completed for bit matrix, memory managers,
and NPN): kept the externally recognizable C API names while using slices and
standard-library operations internally; split the fixed allocator's oversized
field declaration into logical groups. Removed the NPN complemented-input
traversal because every assignment already fills both output polarities. The
original C table hash remains identical after this simplification.

Separate bug audit (completed for these components): fixed int32 overflow in
matrix word-count rounding; reject fixed-manager entry sizes whose minimum
chunk exceeds int32 accounting and flexible-manager doubling overflow before
allocation. Public NPN transforms now reject duplicate/out-of-range permutation
variables and unsupported phase bits rather than silently returning an invalid
truth table. New regression tests cover invalid mappings and allocator overflow.
The exhaustive reconstruction/orbit tests and original-C table oracle continue
to validate all NPN table entries. These audits do not imply the unfinished
parent package has been ported or reviewed.

Additional completed components:

- `extraUtilReader.go`: all nine file-reader functions now stream through
  `bufio.Reader`, with `[]string` tokens and 64-bit file size/position values.
  Returned token storage remains valid across calls. The reader supports the
  original configurable character categories and special `//` comment rule,
  while also handling long tokens/comments and EOF comments without C buffer
  overrun behavior. `Extra_FileReaderErr` reports read/close failures.
- `extraUtilProgress.go`: progress output accepts `io.Writer` and is configured
  through `Extra_ProgressBarSetMode(enabled, batch)` instead of a dependency on
  base/main's command frame. Main/frame code must connect its settings to this
  setter. The header's duplicate alias/update definition moved into the native
  implementation; no functionality was excluded.

Both components passed separate idiom and bug reviews. Reader review replaced
manual refill/pointer state with buffered streaming and checked token lifetime,
EOF comments, category precedence, position/line accounting, read/close errors,
and idempotent release. Progress review uses a string builder for one output
write instead of per-character writes; uses int64 threshold arithmetic to avoid
overflow, atomic mode settings, and explicit stopped/nil handling. Race-enabled
focused tests cover these cases.

`extraUtilFile.go` is fully ported (all filename/file/string/binary/hex APIs,
including two-file reads). File reads preserve the original appended newline,
but return `(string, error)` and mutable string helpers return owned strings.
`Extra_PrintHexadecimalString` returns its string instead of accepting an output
buffer. Printing accepts `io.Writer`; truth words are `[]uint32`.

Separate required reviews for this group:

- Compactness/idiom/readability: shared hex/binary format helpers replace repeated
  C digit loops; immutable strings replace shared static buffers, io/os APIs
  return errors, and file line transformations use strings/sort without legacy
  fixed buffer sizes.
- Bug review: corrected hex shifts at 32-bit word boundaries; digit validation
  prevents partial `Extra_ReadHex` output changes; PrintSymbols honors its writer
  and emits raw high bytes; filename extension operations do not strip dots in
  parent directories; line numbering handles short/unterminated lines without
  deleting content and reads before writing to support in-place operation.

Race-enabled tests cover all 65536 four-input truth-table hex round trips,
64-bit word boundaries and reverse output, raw-byte symbols, path cases,
invalid digits, file read combinations/newlines, sorting and in-place numbering.

`extraUtilUtil.go` is fully ported: configurable short-option parsing, time/string
helpers, file checks/search, OOM callback and clock wrappers. Independent
`Extra_OptionParser` values are available alongside the original global-parser
wrappers; exported getters expose legacy option argument/index without requiring
cross-package access to C globals. Clock wrappers share misc/util's portable
monotonic elapsed-nanosecond policy (rather than Unix-only process CPU time).

Its required idiom/readability review introduced independent parser state,
standard os/path/string operations and owned strings. The separate bug review
covered argv/index bounds, reset and clustered/required-option behavior,
independent instances, file-open failures, exclusive-write nontruncation and
legacy getter locking. Focused race tests pass. Empty Extra_MemTest is faithful
to its source, whose only allocation is commented out.

`extraUtilTruth.go` is fully ported, retaining the word-level algorithms:
elementary tables, adjacent swaps (both argument-order APIs), stretch/shrink,
support queries, cofactors, existential/universal abstraction, mux, symmetric and
antisymmetric checks, phase changes, minimum cofactor-support overlap, cofactor
counts, hashing and semi-canonicalization. Inputs/outputs use `[]uint32`;
minimum-overlap returns `(minimum, variable)`, and cofactor-count storage is
`[]int32` rather than overflow-prone C short. Stretch/shrink accept storage for
all variables, as in C, and leave the input intact unless it aliases output.

Required readability/idiom audit: factored shared cofactor/quantification and
variable/storage checks, retained explicit word masks and stable adjacent
semi-canonical sorting, used slices/math-bits instead of pointer arithmetic,
and derived the same first 1024 hash primes once rather than copying a large
constant table.

Separate bug audit: fixed original high-variable mux offsets (the C function
advances output but not the cofactor pointers), made output/input overlap safe,
removed shared static scratch tables and their race hazard, validated variable
and placement bounds, and widened cofactor counts to avoid short overflow.

Race tests compare word operations with independent bit-level oracles through
10 variables; cover swap aliases, stretch/shrink round trips, support, symmetry,
quantification, mux block offsets, 17-variable count overflow, and invalid inputs.
All 65536 four-input functions' semi-canonical truth, phase, permutation and
cofactor counts match a compiled original-C oracle, whose SHA256 is recorded in
a Go regression test. No C compiler is required to run the tests.

`extraUtilPerm.go` is fully ported: all original ZDD manager, unique/cache,
family operations, cofactors, path/node counts, threshold/products, transposition
composition/conversion, printing and demonstration entry points are native Go.
Canonical nodes use explicit integer IDs and maps; the operation cache is bounded
and node storage grows lazily up to the caller's original node limit, avoiding
mandatory multi-gigabyte preallocation for the cube diagnostic. Path counts are
uint64 with overflow checks, and shared-node counting uses original mark/unmark
semantics. `Abc_ZddNode` returns a value to avoid dangling pointers after growth.
`Abc_ZddCountNodesArray` accepts []int32. The Gia diagnostic uses a dependency-
neutral graph interface (ObjectCount/InputObjectIDs/AndFanins); Gia implements
these methods without an import cycle. Demonstration output is consolidated.

Required compactness/idiom/readability review: shared union/min-union and product
recurrences avoid duplicate bodies; typed keys, maps and slices replace raw hash
buckets/pointer arithmetic; explicit terminal meanings and ownership are stated.
The separate bug review corrected original intersection's unequal-top-variable
handling and zero-threshold loss of an empty member, bounded memo growth, ensured
node values remain safe across append/reallocation, validated permutation inputs,
and prevented silent path-count overflow. Existing C min-union/cofactor semantics
are otherwise retained rather than relabeled as stronger operations.

Validation: exhaustive all-pairs union/difference/intersection on all 256 families
of three variables; dot-product set oracle and threshold cardinality properties;
all six three-element permutation compositions and group closure; transposition
involutions; 60-variable path counts/shared-node marks; invalid inputs and limits;
mock Gia adapter; and original cube-state counts 1/10/64 for the first rounds.
All-pairs union/difference/dot/min-union results also match a compiled original-C
oracle's SHA256. Race-enabled tests pass. Remaining product drafts stay visible.

`extraUtilCanon.go` is fully ported: the recursive fast input-negation-only
canonicalizer and its three original diagnostics use native values and owned
results. `Extra_TruthCanonFastN(max, real, truth)` returns `(words, phases)`;
phase count is len(phases). Small functions retain the original repeated-word
representation, and real-variable counts below three use the C three-variable
base domain. This operation intentionally does not use the NPN subpackage:
variable permutation/output-negation equivalence is a different operation.

Required compactness/idiom/readability review: replaced large static lookup data
with its exact three-variable phase enumeration; replaced flag-indexed global
scratch buffers and pointer casts with recursive value results; extracted the
phase-bit helper and retained original cofactor branch structure. Native code is
196 lines versus the 691-line C file, while all public entry points remain.

Separate bug review: removed architecture-sensitive byte/unsigned casts and
shared-static-result races. Corrected incomplete phase results for constant
cofactors: C returns only eight phases for a four-variable constant function,
although all sixteen phases minimize it. Go returns every valid minimizing
phase. Canonical words remain identical to C, exhaustively oracle-hash checked
for all 65536 four-variable functions. Tests independently brute-force all
four-variable minima/phase sets and random five/six-variable cases, exercise
constant phase counts, bounds, owned results and concurrent calls. Race tests
pass alongside the already ported truth operations.

`extraUtilCfs.go` preserves its sole original command's genuinely empty body.
Its graph argument is dependency-neutral (`any`), avoiding an unnecessary Gia
cycle. No CF simulation algorithm exists in the source and none is claimed.
The larger parent package still fails first on retained header preprocessing in
extra.go; remaining original API groups stay in the build.
