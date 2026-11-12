# Sequential correspondence and resimulation review

Implemented `cecCorr.go`, `cecSeq.go`, and `sat_sequential.go`. The latter replaces the two obsolete raw sequential-SAT definitions removed from `cecSolve.go` under its owner's agreement. GIA identity/name/representative APIs are supplied by the GIA owner; class simulation and incremental SAT are actual neighboring implementations.

## Compactness, idiomatic Go, and readability pass

Frame copies are local `[][]int32`, indexed by frame and original object ID. They replace hidden manager scratch-vector allocation and keep the speculative and real values visibly distinct. A small typed AND-hash builder handles constant simplification and shared logic; XOR/MUX become their actual AND decompositions. Frame zero holds either free initial registers, zero initialization, or the supplied semiformal state. Primary inputs stay in source frame order.

Packing and sequential resimulation use borrowed `[]uint32` rows. The bit helpers make the source cube convention explicit: an even literal requires true, an odd literal requires false. Sequential SAT retains one real incremental manager across output queries, saves only queried-cone CI assignments, grows storage, and reserves bit zero at the source's block boundaries.

Dead-register cleanup computes the fixed point of observable next-state cones. Choice construction retains alternative implementations, excludes transitive-fanin cycles and ordinary-fanout alternatives, and records actual GIA representative/next links. Metadata copies use validated manager APIs. Class progress comparisons use typed slices rather than string formatting. Scratch dependence marks are local instead of changing caller object marks.

## Independent bug review

Reviewed initialization/prefix frame indexing, phase/complement propagation, ring closing edges, initial-register representative mappings, output-pair order, SAT status signed bytes, unknown-result removal, counterexample bounds and batching, register-remap ordering, CI/CO frame order, best-state handoff, register liveness, names, callback ordering, and choice cycles/duplicate links.

Findings corrected with regressions:

- Unknown SAT results remove their candidate member; still-disproved candidates are conservatively removed if packed resimulation makes no progress. The latter prevents silent stalled refinement while retaining soundness.
- A saturated packing batch returns an unconsumed pattern for retry instead of dropping it. Malformed/truncated cube stores fail explicitly.
- Conflict-limit growth uses int64 before clamping, avoiding int32 multiplication wraparound.
- Zero-PI simulation layouts avoid the C division/modulo-by-zero case. Autonomous oscillator correspondence is tested.
- Combinational counterexample resimulation randomizes all CI rows and does not apply sequential register remapping.
- Prefix inverse maps keep the first source image and preserve object zero as constant zero when structural hashing collapses objects.
- Duplicate choice alternatives cannot create a self-linked class. Malformed cyclic ring lists fail instead of looping indefinitely.
- Semiformal models omit free initial-register inputs, honor dual-output cone colors, skip already-proved output checks, and grow frames to the requested minimum check count (or the frame limit).

Source-compatible policy remains: caller-requested early iteration limits return candidate information; prefix invariants may describe states after the requested warmup rather than the original reset behavior. Callbacks retain the source's two initial calls and subsequent per-refinement calls. Source recursion-only helpers are replaced by local unroller/duplicator methods rather than exposing mutable global frame state.

The native incremental CNF solver handles both correspondence solver selections; `fUseCSat` retains its source conflict-budget policy without depending on the original C circuit-solver backend. Verbose statistics use Go timing. Semiformal's unconditional diagnostic AIG dumps are omitted from the library algorithm; computation and state refinement are retained.

## Validation

`test_correspondence.sh` builds the existing native CEC/GIA integration group plus these files in the established temporary-module harness. It does not claim unfinished sibling draft files compile.

- Real original ABC `&lcorr` and `&scorr` runs match native reduced latch counts and exhaustive five-frame traces on equivalent/constant and complementary-transition fixtures.
- All latch/signal × rings/nonrings modes preserve exhaustive traces and merge equal-transition registers.
- 60 seeded random sequential networks preserve 2,400 sampled six-frame traces, with one/two induction frames.
- A rare 12-input conjunction exercises SAT-driven refinement that random simulation is unlikely to discover.
- Prefix, semiformal refinement, autonomous oscillators, implication packing, and internal associative-AND choices have regressions.
- 70 SAT outputs force sequential witness storage to grow across multiple reserved-bit blocks.
- The entire native correspondence integration harness passes `-race`; `-vet=all -run '^$'` passes.
