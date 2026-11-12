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

`sat/cnf` now builds independently, and `bmcBmc3.go` contains a native bounded
sequential checker using AIG unrolling, the shared CNF encoder, and BSAT. Its
focused regressions cover a one-frame counterexample and an invariant from the
zero initial state. The `Saig_ParBmc_t` option type now lives beside this
implementation, so these tests can run without loading unrelated legacy
unrolling declarations:

```
go test -race ./sat/bmc/bmcBmc3.go ./sat/bmc/aliases.go ./sat/bmc/bmcBmc3_native_test.go
go vet ./sat/bmc/bmcBmc3.go ./sat/bmc/aliases.go ./sat/bmc/bmcBmc3_native_test.go
```

`Gia_ManBmcPerform` now converts its input through the public GIA-to-AIG bridge
and delegates to the scalable checker. It records the returned witness on the
GIA manager through a validated, copy-owning sequential-CEX API. Adapter tests
cover a one-frame counterexample and a bounded-safe invariant.

Whole-package tests still do not pass. `go test ./sat/bmc` is currently blocked
transitively by syntax errors in the imported `aig/gia` package, beginning at
`giaBalAig.go:98`. The BMC package itself still has parser-invalid files such
as `bmcBmcG.go`, `bmcBmcS.go`, and `bmcChain.go`; the syntax scan reports
hundreds of source-conversion errors across these and other legacy algorithm
families. Normal package race/vet verification remains outstanding.

The package contains roughly 23,000 generated Go lines. A syntax scan found
hundreds of C-style macro assignment expressions and ternaries across files.
Go package integration also needs explicit aliases/adapters for AIG/GIA,
vectors, SAT solvers, and proof storage; the starting-point drafts refer to
those APIs as if they were in the same C translation unit. The focused helper
gates above do not substitute for a whole-package race/vet pass.
