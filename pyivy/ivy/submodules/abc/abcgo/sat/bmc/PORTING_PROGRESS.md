# BMC integration progress

## Verified components

`packed2.go` expresses the CEX annotation's sixteen two-bit fields per uint32
word with typed masks and shifts. `Abc_InfoGet2Bits`, `Abc_InfoSet2Bits`, and
`Abc_InfoAdd2Bits` now use it through `Vec_IntEntry` and
`Vec_IntWriteEntry`, avoiding the C `unsigned *` cast. Tests cover every field,
replacement isolation, and the source's OR-based add semantics. Verification:

```
go test -race packed2.go packed2_test.go -count=1
go vet packed2.go packed2_test.go
```

`parsePivotIDs` mirrors `%d` scanning over whitespace-separated pivot IDs and
stops at the first invalid or overflowing token. Tests cover whitespace,
empty input, invalid tokens, and overflow. Verification:

```
go test -race pivots.go pivots_test.go -count=1
go vet pivots.go pivots_test.go
```

The ternary simulation truth functions (`Abs_ManSimInfoNot` and
`Abs_ManSimInfoAnd`) now delegate to `ternary.go`. Tests enumerate all three
NOT cases, all nine AND pairs, and invalid encodings. Packed-frame allocation
stores `[]uint32` in the pointer vector instead of C `calloc` memory; reads and
writes use the shared two-bit packing helper. Run the two focused test files
together with the packed-field helpers using the commands above with
`ternary.go ternary_test.go` added to the file list.

## Current integration boundary

The package remains unbuildable. `go test ./sat/bmc` currently stops while
building its `sat/cnf` dependency, which still has syntax errors. The BMC
header callback fields, pivot parser/loops, first two sequential-frame
unrolling functions, and the full `bmcBmc2.go` syntax have been converted in
the working tree. BMC algorithm type integration against native graph and
solver APIs remains in progress.

The package contains roughly 23,000 generated Go lines. A syntax scan found
hundreds of C-style macro assignment expressions and ternaries across files.
Go package integration also needs explicit aliases/adapters for AIG/GIA,
vectors, SAT solvers, and proof storage; the starting-point drafts refer to
those APIs as if they were in the same C translation unit. The focused helper
gates above do not substitute for a whole-package race/vet pass.
