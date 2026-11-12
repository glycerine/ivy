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

The former first package blocker `extra.go` is now native Go: its twenty truth/
bit helpers use slices, math/bits, integer flags and overflow-safe word rounding.
C-only inline/platform spelling and redundant typedefs are removed; forward
API documentation remains. Type aliases moved beside their owned structures.
Separate readability and bug reviews covered shared word operators, repeated
small-truth conventions, constants/implication, exact/partial aliases and maximum
bit-count rounding. Native memory recycler/restart/release tests were rechecked
after the broad ABC_FREE cleanup; no C deallocation tokens remain in completed
components. Gen, like Cfs, faithfully retains its source-empty behavior.

`extraUtilCube.go` is fully ported: state packing/permutation/bit/signature helpers
and both original BFS enumeration modes. Complete-frontier generation preserves
the nine original turns and state order. Sparse full-state or seven-corner maps
replace raw hash pointers and the 4GB dense signature bit array. Tests match
original cumulative counts 1/10/64/385/2232, validate quarter-turn order, compare
both modes, enforce limits, and check packing and overlapping permutations.
Readability review shares frontier and move logic; separate bug review checks
state ownership, byte/bit indices and fixed-corner signatures. Race tests pass.

`extraUtilDsd.go` and `extra_dsd_classes.go` are fully ported: all595 original DSD6
representatives/costs, exact720-permutation schedules, configuration encoding,
classification/composition/cut remapping, costs/stats/lifecycle/divisors and Rsb
diagnostic. The original required dsdfuncs6.dat, which is absent here. Go accepts
and validates the original binary format, but also supplies exact native orbit
search when the file is absent. This is a complete search, not heuristic matching:
cofactor signatures/phase populations only prune impossible candidates, followed
by full truth-word equality. Query caching is bounded. Full orbit materialization
is available through Sdm_ManBuildHashTable and reproduces all2,866,420 records.

Separate readability/idiom review shares immutable permutation geometry, typed
maps/keys and orbit primitives, avoids compulsory huge global tables for sparse
use, and replaces unsafe binary-pointer casts with explicit encoding/binary.
Separate bug review validates/truncation-checks records, configurations, cuts and
permutations; permits an absent legacy data file using the exact native algorithm;
protects shared query/stats/singleton state; makes release idempotent; resets all
statistics coherently and avoids repeated elementary-table construction. Phase
mask operations have a direct all64-mask bit-oracle regression.

DSD tests parse/evaluate all595 original strings, check all class identities and
random NPN transforms, differential-check lazy results/config tie ordering against
the complete materialized table, and validate full table cardinality, malformed
files, cut reconstruction, partial masks, constants, costs, stats and concurrent
queries. All metadata,720 permutations and595*720 truth words match a compiled
original-C SHA256 oracle. Focused race tests pass (including full materialization).

Whole `go test ./misc/extra` now reaches `extraUtilEnum.go` syntax errors instead
of the old header/Cube/Dsd blockers. Nine source groups remain drafts: Enum,
Macc, Maj, Misc, Mult, Path, Prime, Supp and extraZddTrunc. These remain active
product files; focused file-list tests do not claim whole-package completion.

### Function enumeration (`extraUtilEnum.go`)

Ported all three fixed datasets, formula printing and cost-ordered AND/OR
formula enumeration, and the original reference-counted, level-ordered AIG
enumeration with truth-based redundancy filters. The optional XOR recursion is
retained. Dataset output dimensions are returned as a tuple. Diagnostic output
uses native Go timing and Kit DSD printing. Sparse truth maps replace the
original mandatory 512 MB allocation; gate-cost boundaries grow with the request.

Readability review: kept original API names and traversal order, separated dataset
transposition, formula construction, and graph recursion, and grouped standard
library imports. Fanin order continues to encode AND versus OR in printed
formulas; this non-obvious convention is covered by direct formula evaluation.

Bug review: fixed integer overflow in dimension validation; source zero-gate
input targets are now recognized, and printing an input node produces its letter
rather than recursively reading sentinel fanins. The original fixed 100-element
cost array is dynamically sized. Added regressions for these cases and restored
references/frontiers after every graph enumeration.

Validation: standalone original C oracle counters match all 18 tested combinations
of 2–4 inputs, 1–4 gates, with and without XOR. An independent exhaustive formula
cost oracle checks truth values and minimal costs. Dataset transpose/bit order,
printer, and dimension checks pass. Focused race command:
`GOCACHE=/tmp/abcgo-cache go -C abcgo/misc/extra test -race extraUtilEnum.go extraUtilEnum_test.go`.
The full package was rerun and now first fails in the remaining active
`extraUtilMacc.go` draft; Enum is not being used to hide that blocker.

### Constant multiplier/accumulator generation (`extraUtilMacc.go`)

Ported the complete signed-coefficient cost search, packed decomposition tables,
recursive shift/add/subtract netlist generation, specification emitters, and
file-generation commands. Emitters accept `io.Writer`, tables are owned
`[]uint32`, and file-generation commands return errors. Coefficient widths are
explicitly limited to 2..15, matching the packed second-operand field; malformed
sizes are rejected before shifts/allocation. The original eight-bit examples
remain unchanged. Diagnostic generation no longer depends on Vec containers.

Readability review: factored shared module headers/output modes and sign/name
formatting; retained the original three-operation search order and packed table
layout. A per-module visited set shares coefficient wires instead of recursively
emitting duplicate declarations. Documented why intermediate positive powers of
two may equal the requested signed-coefficient upper bound.

Bug review: original standalone zero-coefficient recursion did not terminate;
zero now emits an explicit zero wire. Shared intermediate wires are emitted
once. Tests exposed and fixed an overly strict internal signed-bound check:
positive boundary powers of two are legal decomposition operands. Dimension
validation uses wider arithmetic to avoid signed overflow. Missing or short
coefficient tables fail explicitly rather than indexing arbitrary storage.

Validation: complete little-endian decomposition arrays for widths 2..10 match
the original compiled C SHA256
`98634aa12c603fb70ad87f852f3e92ffcf119cfd0faf90ad99a16bf32f0ec306`.
A separate evaluator checks every emitted intermediate and final multiplication
for every signed coefficient at widths 2..9 over six input patterns, also
rejecting duplicate or undefined wires. Zero, negative boundary power-of-two,
and invalid dimensions have regressions. All completed native components,
including Enum and Macc, pass the combined race test in 9.398 seconds. Full
`go -C abcgo test ./misc/extra` was rerun and now stops at the active
`extraUtilMaj.go` draft. Remaining drafts are Maj, Misc, Mult, Path, Prime, Supp,
and extraZddTrunc; no file has been excluded from the full package build.
