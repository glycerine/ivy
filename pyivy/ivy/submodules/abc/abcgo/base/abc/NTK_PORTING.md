# Network lifecycle and transformations

All 38 functions in `src/base/abc/abcNtk.c` are native Go, split between `abcNtk.go` (network transformations) and `ntk_core.go` (allocation, interface setup, duplication, and deletion). `induction_model.go` adds the typed witness handoff used by the narrow `Abc_NtkDarInduction` caller update in `base/abci/abcDar.go`; the rest of that caller family remains visibly unported.

## Ownership and metadata

Networks use actual ABC `Network`/`Object`/`Edge` storage. SOP strings are immutable, Hop/BDD managers are independent on duplication, and strashed managers refer to the destination network. Hierarchy cloning delegates to the same tested duplicate routine. Deleting a duplicate releases its managers without invalidating the source. Source copy bindings retain both regular object pointers and complemented copy edges.

Typed network metadata now includes constraint count, EXDC/EXCARE networks, one-hot groups, sequence witness, object permutation, phases, gate delay, wire-load name, and LTL properties. Mutable slices and witness bits are copied; dependent networks are recursively duplicated. SCL mapping arrays are translated through object copies. CI arrival callbacks are snapshotted by destination CI instead of retaining closures expecting source objects.

`Network.SequenceModel` stores complete or compact induction witnesses. The handoff moves the dedicated `Graph.InductionCounterexample` slot and clears it. Duplication preserves the witness's own register dimensions, so a compact induction witness is not silently expanded into a conventional trace layout. Generic user `Data` remains separate.

C pointer-vector parameters become typed slices. `NtkLiteralNetwork` accepts GIA's public literal/gate interface without requiring private GIA scratch fields. Conversion retains complemented results after structural simplification and supports native XOR/MUX/buffer gates.

## Source fixes and boundaries

- Trim keeps its driver pointer before deleting an output; the source accesses a removed fanin.
- Flop-permutation input validates a full permutation, rejecting duplicates and malformed entries. Stored file mappings are inverted, as in C; a three-cycle regression caught and corrected an initial port error.
- Fanout sorting retains latch terminals in the CI interface rather than dropping them.
- Node splitting accepts limit two; limit one is rejected because the source's zero-size split cannot converge.
- Deferred buffer/inverter creation exposed a preexisting native object-constructor bug. The network owner fixed both constructors and added unary truth regressions.
- The original linked C `MakeOnePo` driver crashes when traversing the returned one-output network after selecting one of two outputs. The native routine is checked against the intended output selection directly; this undefined C case is excluded from the differential oracle.
- Random interface permutation uses Go's random source, preserving valid reversible permutation behavior rather than libc's process-global random stream.

## Validation

`test_ntk.sh` lists actual native source files and all current native ABC tests. It does not rewrite, overlay, exclude, or replace files in the repository. It runs the focused component gate while unrelated ABC drafts remain visible in the ordinary full-package gate.

The original-C oracle links actual ABC objects and compares nine duplication, cone/MFFC, doubling, bottom extraction, one-node, and terminal-conversion variants by interface shape, node count, and complete truth tables. Native tests additionally cover target conjunction, network composition, manager lifetime, metadata/witness isolation, sequential conversion, output edits, fanout splitting, permutation round trips, and typed induction ownership.

```
abcgo/base/abc/test_ntk.sh -race -vet=all -count=1
```

Independent review is in progress. Ordinary `go test ./base/abc` and dependent `go test ./opt/ret` currently stop at unrelated draft syntax (`abcHieGia.go` at the last check). No whole-package success is claimed.

## Independent-review corrections

Separate source review identified three initial port mistakes, each reproduced with a failing regression before correction:

1. Read-start naming now retains the full path and removes only the final dot suffix, matching `Extra_FileNameGeneric`.
2. Graph-changing transforms use fresh metadata. Bottom, Double, and cone extraction discard old witnesses and graph-specific annotations; no-latch starts and paired-miter transforms retain only the fields copied by their respective C routines.
3. File-supplied flop permutations are inverted before storage, matching C and the subsequent unpermute operation.

The full focused actual-source suite passes fresh `-race -vet=all -count=1` after these fixes (3.264 seconds). Source-backed coverage currently exercises the nine variants named above; start/read/finalize, fixed-file permutations, output edits, cone-array/append, and GIA adapter behavior do not yet all have C differential coverage. Native semantic regressions cover several of these, but this is not exhaustive parity testing of every public entry point.

An additional dependency contract fix makes ABC SOP node constructors store actual immutable SOP strings rather than generic callbacks. Exhaustive constructor tests cover constants, unary gates, AND/OR/XOR, and MUX; unspecified-function networks retain their generic callback fallback.
