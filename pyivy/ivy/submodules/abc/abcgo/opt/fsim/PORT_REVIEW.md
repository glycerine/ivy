# Fast sequential simulation port

All original FSIM source files are native Go. Public Fsim entry points remain available with Graph/Edge, slices, exported parameter fields and errors. `RunSimulation` additionally returns explicit frame/timeout/witness results. No GIA/SAIG or proof-package dependency is required; the implementation uses the committed native AIG and utility APIs.

## Separate design and readability review

The simulator preserves the source algorithm's output-root DFS schedule, shuffled CI-to-original-interface mapping, reference-counted frontier liveness, monotone frontier numbers, delta/varint program and word-parallel gates. Logical schedule metadata remains available for verification and switching observations. The simulation buffer scales with frontier width times word count, rather than graph size times word count.

Frontier compilation has explicit operation kinds, supporting native XOR, buffer and output-alias semantics. The compressed stream extends the source tags for these native operations; it is an internal encoding, not an external ABC file format. Both compressed and uncompressed paths execute the same word kernels and expose owned output rows. A frontier verifier reconstructs each original fanin before its slot can be reused.

Sequential simulation uses local instances of the original multiply-with-carry AIG generator. It preserves the source reset draw, descending word generation and shuffled CI traversal without modifying a global RNG. Multi-frame counterexamples regenerate that exact stream, then independently replay through Graph.Evaluate before attachment. Graph stores its own validated copy; returned witness storage remains caller-owned. Ordinary simulation preserves an existing witness.

Ternary simulation uses packed two-bit register states, the original state hash and explicit cycle detection, including the repeated final state. `RunTernary` exposes an optional state limit; unlimited Fsim_ManTerSimulate matches the original convergence API. Native typed links replace pointers hidden after packed state words.

The C switching file contained no implementation despite declaring an entry point. The native implementation provides the advertised activity calculation using the same frontier: one probabilities, transition probabilities, warmup prefix and configurable PI toggle probability. Separate counters record exact sample denominators; probability-only mode avoids allocating transition history.

## Separate bug review

Reviewed slot lifetime and capacity, fanin complement phases, aliasing, constant-first programs, varint bounds, CI permutation, word-zero/high-word failure indices, zero-input widths, random replay, initial register state, witness ownership, packed ternary overwrites and activity denominators.

Four concrete port findings were reproduced by failing regressions and fixed:

1. A compressed program beginning with a constant CO underflowed its predecessor delta. Encoding/decoding now starts from the allocated constant frontier number.
2. Zero-input activity simulation could not infer a word width. Managers have a defined width and high-level simulation explicitly applies its requested word count.
3. Malformed compressed records with zero node IDs were accepted. Decoder validation now rejects zero compute nodes, missing fanins, truncated and overflowing varints.
4. Ordinary simulation cleared a stored witness despite not checking a miter. Witness replacement now occurs only for an explicitly requested miter run.

The port also corrects source defects with dedicated tests: the first-set-bit helper used `32*(w-1)`; ternary AND discarded a controlling zero when its other operand was unknown; packed cycle states assumed shuffled CI storage still had all ROs at the end; and small-frontier allocation needed space for the new node before consuming fanins. Native simulation handles each explicitly. Output-as-fanin aliases retain their complemented phase, and buffers are simulated rather than treated as ANDs.

## Source differences and limits

All registers start at zero, matching the AIG/source model. The Go time limit measures elapsed wall time; zero means unlimited. Constant-miter cases produce actual validated witnesses rather than returning a trivial status without a model. The internal compressed stream includes native operation kinds and is not byte-identical to the old AND-only stream. Ternary constant propagation is more precise where the old AND routine incorrectly returned unknown; it remains a sound Boolean abstraction. Activity estimation is a completed implementation of an otherwise empty source module.

A manager's source graph must remain immutable during use. Independent managers use independent simulation buffers/RNGs. Checking a miter attaches a witness to the supplied graph, so concurrent witness-producing runs should use separate graph instances. Full schedule metadata is retained for validation; frontier compression reduces execution-program and simulation-data storage rather than removing the borrowed source graph.

## Validation

- Sixty seeded native AND/XOR DAGs: both program formats checked against scalar Graph.Evaluate for every bit of two simulation words; frontier reconstruction verified.
- Delayed multi-frame witnesses replayed independently, with Graph/result ownership checks.
- Explicit constant-first program, buffer/output aliases, tiny frontier, malformed codec, ordinary-witness preservation and word-index regressions.
- Ternary oscillator cycles, shuffled CI/RO ordering, controlling-zero abstraction, state-limit and packed-overwrite tests.
- Constant/toggling-register activity, probability flags and quarter-rate PI switching checks.
- Original C oracle: 600 seeded cases comparing word AND kernels, ternary packed storage and the complete state hash. The test-only include forwarding header points the obsolete aig/ssw include to the actual proof/ssw header; no C algorithm is stubbed or removed.

Fixtures are under repository `test_vectors/fsim/`. Production execution is pure Go; only the optional C differential test needs a C compiler and supported ELF linker.

Package gates:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./opt/fsim
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./opt/fsim
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./opt/fsim
```
