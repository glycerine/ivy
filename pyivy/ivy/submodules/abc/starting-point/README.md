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

This creates `starting-point-cpp2go-abc/` in the current working directory:

```text
abc/src/sat/example.c  -> starting-point-cpp2go-abc/src/sat/example.c.go
abc/src/sat/example.h  -> starting-point-cpp2go-abc/src/sat/example.h.go
abc/src/opt/example.cpp -> starting-point-cpp2go-abc/src/opt/example.cpp.go
```

The traversal handles `.c`, `.cpp`, and `.h` files (case-insensitive). Original
extensions are retained before `.go`, preventing same-stem collisions. Directories
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
source-directory name. Files have indexed names such as `00_vec_uint.h.go`.

## Reports and recovery

Tree mode writes `run.json` with source/output paths, packages, hashes, issue
counts, timing, and fallback status. `_starting-point/` contains mirrored original
source snapshots and detailed per-file JSON reports. This metadata directory is
ignored by Go's recursive package discovery.

If analysis fails, the corresponding `.go` file still exists: it contains its
package declaration, an explanatory TODO, and commented original source. The
specific diagnostic is recorded in its report. Other files continue translating.
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
