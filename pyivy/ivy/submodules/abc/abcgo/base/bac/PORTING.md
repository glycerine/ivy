# BAC port status

The package is **not yet buildable**. Most files remain C-to-Go drafts; in
particular `bacBlast.go` is the first file currently reported by `go test`.
Do not treat the parsing of one file or isolated tests as package completion.

The binary BAC reader and writer in `bacBac.go` follow the source header and
native-endian vector layout. Object type bytes are ABC literals: PI/PO types
are decoded by shifting right one bit. `test_vectors/simple.bac` was produced
by ABC's `@_read` and `@_write` commands from `test_vectors/simple.blif`; its
read and round-trip test is present but cannot run as a normal package test
until BAC's network model compiles. `header_probe.c` invokes the source C
header parser directly. Its output was compared with the isolated Go header
tests. The codec round-trip was also exercised using the real vector package
and a temporary network-model interface mirror; that does not verify the
unfinished network model.

`bacPrsTrans.go` translates source range-size helpers, including source
`atoi` decimal-prefix behavior. Isolated tests cover ascending and descending
ranges; the full name-manager integration test cannot run until the parser
model compiles.

Other reader, writer, netlist, parser, and bit-blasting files remain
unreviewed drafts. Full package tests, race tests, and vet have not passed.
No source files are excluded from the build.
