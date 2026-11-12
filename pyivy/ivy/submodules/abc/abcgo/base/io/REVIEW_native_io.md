# Native I/O checkpoint

Work is paused at the user's request. This is an incomplete package checkpoint, not a full I/O completion claim. Drafts remain visible.

## Implemented source families

- Shared public models/file kinds: `ioAbc.go`, `ioInt.go`; concrete format registration: `ioRegistry.go`; dispatch, net construction, path lookup, and generic read-conversion pipeline: `ioUtil.go`; checked writes: `ioText.go`.
- BLIF/BLIF-MV: `ioReadBlif.go`, `ioReadBlifMv.go`, `ioWriteBlif.go`, `ioWriteBlifMv.go`; combinational BLIF-to-AIG bridge: `ioReadBlifAig.go`.
- BENCH: `ioReadBench.go`, `ioWriteBench.go`.
- EQN/DSD: `ioReadEqn.go`, `ioReadDsd.go`, `ioWriteEqn.go`; shared SOP expression generation: `ioSop.go`.
- BAF: `ioReadBaf.go`, `ioWriteBaf.go`; BBLIF: `ioReadBblif.go`, `ioWriteBblif.go`.
- JSON: `ioJson.go` (separate `REVIEW_json.md`).
- EDIF: `ioReadEdif.go`, `ioEdifSyntax.go`.
- CNF: `ioWriteCnf.go`, `ioCnfModel.go`; SMV: `ioWriteSmv.go`.

AIGER is independently owned by ivy_multi; PLA by map_cov; Dot/GML/List/Book by misc_tim. Their files are not included in this ownership claim.

## Distinct readability review

The readability pass replaced C arenas/pointer arithmetic with owned buffers and typed graph references. BLIF uses separate declaration and connection passes for forward model references. A common set/polarity traversal normalizes symbolic MV terms through the actual ABC MVTable representation. EQN lowers an expression DAG into small SOP gates instead of enumerating its whole truth table. BAF and BBLIF reuse real AIGER/BBLIF model adapters. The registry carries concrete read/write implementations without importing command dispatch into format parsers. SMV uses a per-export name map, and CNF carries explicit model context rather than the source process-global current-network pointer.

## Distinct bug review

The bug pass checked driver ownership, interface order, symbolic domains (including numeric labels and binary symbolic domains), default/set/equality terms, multi-output table projection, reset-state latch/MUX circuitry, best-effort diagnostic locations, BAF lengths/references/constant convention/initial states, BBLIF interface cuts, complemented SOP covers, exact JSON numbers and duplicate keys, EDIF joined connectivity, CNF constants/tautologies, and SMV name collisions. EQN rejects names which cannot round-trip through its grammar; SMV omits empty declaration sections. Exporters check short writes and close files even after errors.

Written network-dependent regressions: `blif_native_test.go`, `equation_bench_test.go`, `baf_native_test.go`, `bblif_native_test.go`, `edif_native_test.go`, `smv_native_test.go`. Fixtures are under `test_vectors/`.

## Gates actually run

The combined independent selected-file race and vet gates pass:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 base/io/ioText.go base/io/ioJson.go base/io/ioEdifSyntax.go base/io/ioCnfModel.go base/io/json_native_test.go base/io/edif_syntax_test.go base/io/cnf_model_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet base/io/ioText.go base/io/ioJson.go base/io/ioEdifSyntax.go base/io/ioCnfModel.go base/io/json_native_test.go base/io/edif_syntax_test.go base/io/cnf_model_test.go
```

CNF tests exhaustively quantify internal SAT variables and compare the projected input relation with independent expected functions, including constants, complements, and sparse IDs. JSON/EDIF syntax tests cover round trips and malformed input.

Ordinary `go -C abcgo test ./base/io` currently fails while compiling the imported `base/abc/abcLatch.go` (raw preprocessor directives at line30 and retained C syntax). Consequently, the network-dependent tests above have not passed an integration gate yet; further type/behavior issues may be exposed when this dependency builds.

Within I/O, remaining parser blockers are `io.go` command dispatch and `ioReadVerilog.go`/`ioWriteVerilog.go`. No files were hidden or excluded from the ordinary gate.

## Remaining behavior work

BLIF currently implements the core .model/.inputs/.outputs/.names/.table/.mv/.latch/.subckt/.blackbox and MV reset directives. Mapped .gate/.flop, timing/attribute directives, and external don't-care networks still require implementation. Unsupported directives fail with a filename and line rather than silently losing semantics. The BLIF-as-AIG bridge currently uses the combinational hierarchy converter; the original sequential fast-reader path still needs lowering. Generic read/write preparation depends on the unfinished public ABC netlist conversion APIs. EDIF semantic/array/hierarchy and all other network-dependent tests require the ordinary integration pass. Command-only external engines and their options have not yet been ported.
