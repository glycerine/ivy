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

`Gia_ManBmcPerform`, the BMCg/BMCs entry points, `Bmc_ChainTest`,
`Bmc_ChainFailOneOutput`, and `Gia_ManMultiProve` now use the native scalable
checker. The GIA entry points convert through the public GIA-to-AIG bridge;
chain and multi-output checks project one output cone at a time while retaining
the register interface. Sequential witnesses are copied onto the GIA manager
through validated accessors. Focused tests cover the scalable checker, output
projection, and GIA entry-point behavior. Five duplicate GIA-to-CNF helpers
were consolidated into `cnf_remapped.go`; all had the same behavior of treating
every CO as an asserted output, which the shared encoder now does directly.
The mesh SAT-backend variant now shares the byte-for-byte identical variable
index helpers from the original mesh variant, while retaining its own solver
operations.

`Bmc_LoadTest` now uses the shared CNF encoder and a typed BSAT solver, with
focused tests for satisfiable and constant-false outputs and malformed CNF.
Its focused race/vet gates pass. `bmcUnroll.go` now delegates to the native AIG
frame builder; tests check frame ordering, initial register state, and the
last-frame register-input outputs, and pass under race/vet.

`Gia_ManBmciTest` has been restored as a typed GIA adapter. Its core
`CompareInitialStatesAIG` builds a shared-input two-state unrolling, encodes
the register-state XORs with the common CNF layer, and checks each difference
under BSAT. The adapter preserves the source convention that initial values
above 1 denote independent unconstrained state inputs. Focused tests cover
divergent identity state, reset convergence, shared inputs, symbolic
initialization, zero frames, and invalid initialization. Its focused race/vet
gates pass; the public GIA wrapper remains transitively blocked by GIA syntax.

The C-derived `bmcBmc2.go` object-pointer implementation has been replaced by
the shared native checker at `Saig_BmcPerform`. The adapter preserves its
legacy status convention (`0` for a counterexample, `-1` for bounded-safe or
early stop) and its reported-frame offset. Its ternary simulator now traverses
the typed AIG graph and stores owned packed frames; tests cover both safe and
unknown properties, status/frame mapping, and packed accessor behavior. Focused
race/vet gates pass. The former internal mutable-object helpers had no callers
outside that C implementation; project call sites use `Saig_BmcPerform`.

`Saig_ManBmcSimple` now also delegates to the shared checker and preserves the
legacy `0`/`-1` result and counterexample frame. `Saig_ManFramesBmc` is backed
by the native frame builder. A call-site scan found no use of the removed
private SAT/model helpers or the size-limited frame builder elsewhere in the
project. Focused tests cover safe and failing models and pass under race/vet.

Whole-package tests still do not pass. `go test ./sat/bmc` is currently blocked
transitively by syntax errors in the imported `aig/gia` package, now beginning
at `giaBalLut.go:66`. The BMC syntax scan still reports errors across other
legacy families, including CEX minimization, fault, majority, and
interpolation support. Five duplicate `Cnf_DeriveGiaRemapped` declarations
and three duplicate `Cnf_DataLiftGia` declarations have been consolidated.
Mesh variable helpers were deduplicated after matching their formulas. The
Exa/Maj families in `bmcMaj.go` and `bmcMaj2.go` are distinct solver/topology
variants. Their duplicate helper namespace in bmcMaj2 now carries a `V2`
suffix, and its random-topology public entry point is `Maj_ManExactSynthesis2`,
matching the base/abci call site. These large files remain parser-invalid, so
variant behavior tests and normal package race/vet verification remain
outstanding.

The package contains roughly 23,000 generated Go lines. A syntax scan found
hundreds of C-style macro assignment expressions and ternaries across files.
Go package integration also needs explicit aliases/adapters for AIG/GIA,
vectors, SAT solvers, and proof storage; the starting-point drafts refer to
those APIs as if they were in the same C translation unit. The focused helper
gates above do not substitute for a whole-package race/vet pass.
