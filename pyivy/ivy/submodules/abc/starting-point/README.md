# starting-point

A Go library and command-line tool for producing readable C/C++-to-Go drafts
for a second repair pass. Source comments and original text are retained. These
are starting points, not automatically verified or compilable ports.

## Install and translate a tree

From this directory:

```sh
go install ./cmd/starting-point
```

From a directory containing `abc/`:

```sh
starting-point abc
```

The same `starting-point abc` command also works from inside a checkout named
`abc`: if there is no `abc/` subdirectory, the bare name denotes the current
checkout, even when it contains an executable also named `abc`. `starting-point .`
is an explicit alternative.

This creates `starting-point-cpp2go-abc/` in the current working directory:

```text
abc/src/sat/example.c  -> starting-point-cpp2go-abc/src/sat/example.c.go
abc/src/sat/example.h  -> starting-point-cpp2go-abc/src/sat/example.h.go
abc/src/opt/example.cpp -> starting-point-cpp2go-abc/src/opt/example.go
```

The traversal handles `.c`, `.cpp`, and `.h` files (case-insensitive). Unique stems use simple `.go` names, such as `giaSim.c` → `giaSim.go`.
When sources in the same output directory share a stem, their extensions are
retained (`example.c.go`, `example.h.go`). Names are checked again to avoid
secondary collisions with dotted stems or destination directories. Directories
containing source files are mirrored. Existing destinations are refused. Symlinks,
version-control metadata directories, and prior `starting-point-cpp2go-*` output
trees are skipped. Inventory happens before output creation, so an output inside
the source tree is not recursively translated into itself.

Each output directory has one Go package derived from its directory name. Names
already valid in Go remain unchanged. Invalid characters become underscores,
leading digits receive `pkg_`, and Go keywords receive a trailing underscore.
The generated root package, if it has files, is `starting_point_cpp2go_abc`.

```sh
starting-point -out /tmp/abc-drafts abc
starting-point -package shared abc
```

`-package` overrides package declarations throughout the tree. `-out` overrides
the destination. Existing files are never overwritten.

## Flat file batches

The earlier file-list mode remains available:

```sh
starting-point -out /tmp/new-draft ../src/sat/satoko/utils/vec/vec_uint.h ../src/sat/satoko/utils/heap.h
```

All files in a flat batch use the explicit `-package` value or the first input's
source-directory name. The same collision-aware naming applies. Identical basenames from separate
source directories receive numeric prefixes to distinguish them in the flat
destination.

## Reports and recovery

Tree mode writes `run.json` with source/output paths, packages, hashes, issue
counts, timing, and fallback status. `_starting-point/` contains mirrored original
source snapshots and detailed per-file JSON reports. This metadata directory is
ignored by Go's recursive package discovery.

If analysis fails, the corresponding `.go` file still exists: it contains its
package declaration, an explanatory TODO, and commented original source. The
specific diagnostic is recorded in its report and in `run.json`. The CLI lists
each fallback source, output, and diagnostic after its summary so failures are
visible immediately. Other files continue translating.
Filesystem errors stop the run and can leave partial output. Use a fresh output
directory for a subsequent run.

Unrecognized constructs may remain as C/C++ in ordinary drafts. A successful run
means output was produced, not that it compiles. Original snapshots retain their encodings. Invalid UTF-8 bytes inside generated
comments become underscores; bytes in code and literals are unchanged and can
still require repair before Go tooling accepts the file.

## Library and validation

The module is `github.com/glycerine/abc/starting-point`; the Go package name is
`startingpoint` (Go identifiers cannot contain hyphens). `TranslateTree` implements
recursive mode. `Analyze`, `AnalyzeWithOptions`, and `ResolvePackage` support
individual files and explicitly coordinated batches. The CLI is a thin wrapper.

```sh
go test ./...
```

There are no third-party Go dependencies. The module declares Go 1.22 and was
tested with Go 1.27.3. Differential tests require `cc` and the surrounding ABC
sources. They compare a repaired Satoko vector/heap port against original C.

Versioned experiments live under `experiments/satoko/_draft-vXX/`, which Go skips
during recursive discovery. Their `.go` files can be passed explicitly to Go
tools. Earlier versions retain their historical `.go.txt` extension. The repaired
experiment and replay script remain separate from generated drafts.

See [PLAN.md](PLAN.md) for transformation rules, decisions, measurements, known
limitations, and progress. Current transformations include source-preserving
lexing, macro function scaffolds, typedef/struct pairs, signature conversion,
if/for braces, while-to-for conversion, ternary lowering, header guard removal,
and commenting out includes and standalone namespace markers. Full C++ parsing
and preprocessor evaluation are not implemented.

Simple integer object macros become Go constants. C integer suffixes determine
explicit types under the current ABC LP64 target: U uses uint32 or uint64 by
value, UL/ULL use uint64, and L/LL use int64 (or uint64 for sufficiently large
nondecimal literals). For example, `#define NUMBER1 3716960521u` becomes
`const NUMBER1 uint32 = 3716960521`. Unsuffixed literals remain untyped constants.

The lexer retains multiline block comments inside preprocessor directives, so
comment punctuation does not participate in code delimiter matching.

Fallback diagnostics include original-source line and byte column (both 1-based)
and byte offset (0-based). `run.json` also contains structured `FallbackLocation`
fields. Look these up in the original source, not the generated Go draft. Errors
from later rewriting stages are explicitly labelled as transformed-draft errors.

Conditional-preprocessor regions are preserved as opaque original text, including
all alternatives. Temporary delimiter-only scaffolding lets analysis continue
around regions that supply part of a surrounding call, condition, or block. It
is removed before output; no branch is selected. Incompatible branch structures
cause a larger source span to be retained conservatively. Reports identify these
spans as `conditional-review` issues. Recognized header guards are still removed.

Ordinary top-level C functions now receive the same mechanical signature and
body rewrites as static-inline helpers. Unsupported declarators still require
review. Existing control-flow braces are moved onto their header lines, and
statement traversal includes blocks following iterator macros.

With `-pre` (for example, `starting-point -pre abc`), preserved preprocessor
conditionals are listed on the command line with original
source `path:line:column`, byte offset, directive text, and destination file.
Nested conditional openers are included; removed include guards are not.
These branch-decision entries also appear as `Conditionals` in JSON metadata,
independently of whole-file analysis failures. The flag defaults to false;
JSON locations remain available even without it.

Standalone named struct definitions are translated even when their typedef is
in a separate header. For example, `struct Aig_MmFixed_t_` becomes a Go struct
with that tag name; `typedef struct Aig_MmFixed_t_ Aig_MmFixed_t;` becomes
`type Aig_MmFixed_t = Aig_MmFixed_t_`. Simple pointer/scalar fields are reordered
without deleting comments; complex fields still require review.

Simple direct printf/fprintf calls become fmt.Printf/fmt.Fprintf with a fmt import
added only when needed. Standalone `sprintf(buffer, format, ...)` calls with a
named buffer become `buffer = fmt.Sprintf(format, ...)`; uses of the C count
result remain for review. Format specifiers, FILE pointers, and destination
buffer types can still need semantic repair.

Recognized `char *` types now default to Go `string`, and direct `strlen` calls
become `len`. Mutable buffers, allocation/copy routines, C null/truth-value
semantics, and additional pointer levels still require representation review.

Simple array parameters become slices in their outer dimension (`char *argv[]`
becomes `argv []string`), and fixed local/package arrays retain their bounds
(`pthread_t ids[N]` becomes `var ids [N]pthread_t`). Plain char buffers use byte
arrays. Mixed declarations with initializers are split into separate var lines
in source order, including FILE pointers.

Conventional C main functions become `func main()`. Their argv/argc parameters
map to os.Args/len(os.Args), with an os import when needed. Zero returns become
bare returns; other exit statuses use os.Exit. Package selection still follows
the directory name or `-package` override.


## Standard-library conversion

Generated adapters must never introduce anonymous functions or closures. Direct
Go calls may intentionally leave return-value/type errors for downstream repair;
calls that require structural adaptation remain review sites (prototype-43).


The library pass now has a documented catalog of 302 C-library names, classified
as implemented, conditional on recognizable argument shapes/types, or retained
for review. See [STDLIB.md](STDLIB.md) for the C11 header survey, references,
conversion policy, full rule table, and validation details.

The pass runs automatically during ordinary translation. It covers literal-mode
fopen, fclose, bounded fgets into known byte buffers, atoi/atol/atoll numeric
prefixes, standard streams/status constants, byte I/O and file positioning,
selected string/memory operations, ASCII character classification, and common
float/double math. It manages imports and preserves single argument evaluation.
Cases such as raw allocation, pointer-returning searches, C stream error state,
complex scanning, shell commands, CPU clocks, locale, and sorting remain explicit
review records instead of misleading substitutions.

The Go API exposes `LibraryRules()` and
`RewriteStandardLibrary(draft) (string, []LibraryChange, error)`.
`Result.Library` contains per-call decisions, source-line excerpts, and labeled
pre-library-draft line/column/byte locations. Tree run.json contains per-file
converted/review counts; the directory CLI prints an aggregate summary. Full
review details remain in `_starting-point/<source>.json`, keeping console output
manageable and independent of `-pre`.

Conversions are drafts, not proof of equivalence. The tests compile and execute
small generated programs and compare a translated C consumer against native C.
The latest whole-src coverage is recorded in
[prototype-41-summary.json](experiments/stdlib/prototype-41-summary.json).


Timing queries use the Go time package: time(NULL) becomes time.Now().Unix(),
and time(&value) also stores the timestamp. clock() uses an explicitly recorded
wall-time policy: time.Now().UnixNano(), int64 clock_t, and time.Second for
CLOCKS_PER_SEC. This changes CPU-time semantics; time.Time/time.Since is the
preferred later repair for monotonic elapsed timing. Calendar/struct tm routines
still require representation decisions.
