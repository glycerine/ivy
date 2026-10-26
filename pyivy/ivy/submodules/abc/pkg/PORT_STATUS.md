# Native Go port status

The full ABC port remains in progress. The detailed roadmap is
[ABC_GO_PORT_PLAN.md](ABC_GO_PORT_PLAN.md). The current library contains native
Go foundations, AIG/AIGER support, an initial native bsat CDCL solver, a simple
Tseitin CNF encoder, and the GIA graph core. Timeframe unrolling and bounded checking now connect these pieces. Automatic four-input optimized CNF now connects DAR cut generation, area-flow
mapping and native-order clause emission on supported AIGs. Other optimized
CNF variants, the ABC command interpreter and a replacement executable for Ivy
remain unfinished. A basic PDR solving loop now returns replayed witnesses
or independently checked inductive invariants. Full ABC PDR feature parity
and larger-design validation remain outstanding.

## Current work

Current inventory: 1,757 tracked source, header, asset and build files, with
677 validated, 573 partially covered and 507 unstarted. These are file coverage
categories, not a percentage of completed functionality. Nineteen supervised
parallel waves and thirty-two strictly serial waves are integrated. Wave 51
ports FRA one-frame inductive-invariant verification.
All implementation and review proceed one task at a time to control token
expenditure. Native scalable engines and explicitly unsupported workflows
remain unfinished. The full port remains at an early stage.

The IVY FRAIG completion classification has been reconciled: `ivyFraig.c` is
fully validated, its complete 87-definition inventory remains auditable in
`IvyFraigDefinitions`, and its coverage-ledger entry contains no residual
partial-symbol marker. A regression assertion now prevents the in-package Ivy
coverage inventory from silently classifying it as partial again. The focused
tests and exactly 5,000 standalone fuzz executions pass, the targeted race gate
passes in 1.049s, vet is clean, and the ordinary full suite passes in 75.225s.

## Fifty-first serial-wave checkpoint

This single-threaded wave ports the sole definition in `fraIndVer.c` as an
owned one-frame invariant verifier over concrete AIG literals. It derives the
native CNF relation once, checks every clause against the initialized base
frame, constructs two uninitialized frames with exact latch transition
equivalences, assumes the complete invariant in frame zero and checks each
clause in frame one. Satisfiable and resource-limited checks remain explicit
per-clause failures; satisfiable models are cloned and replayed against the
formula and assumptions before being returned. An inconsistent induction
premise is reported distinctly rather than mistaken for a verified invariant.

The source supports only K=1. The Go API returns `ErrUnsupported` for other
depths instead of preserving the C routine's accidental success result. It
also adds checked object/variable/clause/literal/work bounds, overflow-safe
accounting and cancellation without mutating caller clauses. Fixed tests use
an independent exhaustive state/input oracle and cover base and induction
failures, empty-clause inconsistency, deterministic SAT unknown, model
ownership, malformed inputs, every limit, cancellation, a retained 10,000-gate
graph and sixteen concurrent readers. The generated sequential-AIG fuzzer
passes its 5,000x single-worker campaign against the same independent semantic
model. The targeted race gate passes in 1.454s, ordinary tests in 75.274s, vet
is clean and the scheduled full-package race passes in 398.681s. The next
scheduled full race is Wave 54. Final source and policy audits remain Go-only
and clean.

## Fiftieth serial-wave checkpoint

This single-threaded wave ports every definition in `saigIsoFast.c` as bounded,
owned structural signatures and candidate primary-output classes. It preserves
the source's packed 16-bit bitfield layout, field truncation, modulo-four
positive and negative fanout counts, complement-first fanin ordering, truncated
level deltas, dynamic latch-output-to-latch-input traversal and postorder
first-signature order. Frequency ordering uses the already validated exact ABC
cost permutation, and vector ordering reproduces the source's little-endian
bytewise `memcmp` rather than substituting numeric lexicographic comparison.

The experimental C driver computes its classes, prints them, frees them and
returns `NULL`; the Go API exposes the same signatures and classes as owned
typed results. Buffers are rejected because the native collector accepts only
CIs and AND nodes. Fixed tests verify manually derived packed words, sequential
classes, constant outputs, ownership, malformed graphs, every limit,
cancellation, a retained 10,000-gate cone and sixteen concurrent frozen
readers. A generated sequential-AIG fuzzer compares against a separate
recursive packed-signature/class model for exactly 5,000 single-core
executions. The targeted race gate passes in 1.072s, ordinary tests pass in
75.652s and vet is clean. Wave 48's 398.659s package-wide race remains current;
Wave 51 is next. Final source and policy audits remain Go-only and clean.

## Forty-ninth serial-wave checkpoint

This single-threaded wave ports every definition in `saigIoa.c` as bounded,
owned input/output for its deliberately restricted structural sequential-BLIF
dialect. Object naming retains the source's padded `n`, `pi`, `po`, `lo` and
`li` namespaces. Marshaling preserves source comments and statistics, CI/CO
partitions, zero-initialized latch declarations, conditional constant-one
emission, source object-number references and complement polarity in cubes.

The reader accepts the exact emitted dialect while replacing static token
buffers, unchecked numeric mapping and assertion/print failures with bounded
owned tokens and typed errors. It validates directive order, terminal names,
latch pairing, fanin references, cube shapes, regular AND results and final
partitions. Path wrappers use ordinary Go files; constraints remain absent
because the source format does not encode them. Fixed tests cover byte-stable
memory and path round trips, exhaustive output semantics, names, metadata,
malformed inputs, every resource limit, cancellation, a retained 10,000-gate
graph and sixteen concurrent frozen readers. A combined arbitrary-parser and
generated sequential-AIG fuzzer passes exactly 5,000 single-core executions.
The targeted race gate passes in 1.278s, ordinary tests pass in 75.431s and vet
is clean. Wave 48's 398.659s package-wide race remains current; Wave 51 is
next. Final source and policy audits remain Go-only and clean.

## Forty-eighth serial-wave checkpoint

This single-threaded wave ports every definition in `saigDual.c` over the
concrete native AIG manager. Dual-rail construction preserves expansion of the
selected PI prefix, ordinary Boolean PI encoding, global or per-register
unknown initialization, complemented-fanin care handling, AND propagation,
zero/one/X violation modes, PO-versus-latch-input selection, optional output
complementation and the doubled latch partition. When zero and one checks are
both requested, the source's zero-check priority is retained.

Property-cycle blocking returns an owned graph instead of partially mutating a
caller on error. It nevertheless preserves the source's transient PO boundary
while delay registers are appended, including gating the selected prefix of
old latch-input outputs. Fixed tests exhaustively compare all option
combinations and raw rail assignments with an independent three-valued model,
compare blocking assignments with an independent sequential model, and cover
ownership, malformed inputs, every resource limit, cancellation, a retained
10,000-gate graph and sixteen concurrent frozen readers. A combined generated-
AIG fuzzer passes exactly 5,000 single-core executions. The targeted race gate
passes in 1.115s, ordinary tests pass in 75.854s, vet is clean, and the
scheduled full-package race passes in 398.659s. Wave 51 is the next scheduled
full race. Final source and policy audits remain Go-only and clean.

## Forty-seventh serial-wave checkpoint

This single-threaded wave ports both definitions in `saigOutDec.c` over the
concrete native AIG, Tseitin-CNF and CDCL SAT layers. Property-prime discovery
preserves native DFS candidate order, phase-relative literals, marking of
single-literal primes, exclusion of their nodes from pairs and the source's
three tested pair polarities. Only UNSAT assumptions produce primes; bounded
unknown answers are counted and skipped instead of being mistaken for proof.

Decomposition creates an owned graph with original primary outputs first,
prime-cube outputs next and latch inputs last, then restores register and
constraint metadata. Fixed tests independently enumerate all assignments to
check prime semantics and exact polarity/order, replay every decomposed output,
check sequential partitions and ownership, cover malformed inputs, every
resource limit, cancellation, a retained 10,000-node graph and sixteen
concurrent frozen readers. An independent exhaustive-model fuzzer passes
exactly 5,000 single-core executions. The targeted race gate passes in 1.179s,
ordinary tests pass in 75.027s and vet is clean. Wave 45's 407.673s
full-package race remains current; Wave 48 is next. Final source and policy
audits remain Go-only and clean.

## Forty-sixth serial-wave checkpoint

This single-threaded wave ports all three definitions in `abcSense.c` over the
concrete native AIG, Tseitin-CNF and CDCL SAT layers. Iterative miter
construction creates the zero and one cofactors of a selected input. It XORs
and ORs every ordinary output difference, then requires the final output—the
source's constraint convention—to hold in both cofactors. Complete native SAT
replaces the bounded Ivy proof pipeline while preserving explicit unknown
results under caller resource limits.

Per-input results own their SAT models and replay each model against its miter.
SAT and resource-limited unknown inputs are classified as sensitive, matching
the source; UNSAT inputs are insensitive. Fixed tests exhaustively validate
every miter assignment, independent sensitivity classification, the two-sided
constraint, replay and ownership, malformed and sequential inputs, every
resource limit, cancellation, a retained 10,000-node cone and sixteen
concurrent readers. An independent generated-AIG cofactor model fuzzer passes
exactly 5,000 single-core executions. The targeted race gate passes in 1.048s,
ordinary tests pass in 75.438s and vet is clean. Wave 45's 407.673s full-package
race remains current; Wave 48 is next. Final source and policy audits remain
Go-only and clean.

## Forty-fifth serial-wave checkpoint

This single-threaded wave ports both definitions in `abcReorder.c` without the
source's optional CUDD runtime dependency. Local nodes use owned fanin IDs and
packed truth tables. Reordering first ties duplicate fanins and removes the
minimum base, then invokes the validated deterministic BDD ordering engine and
relabels both fanins and truth variables so the external Boolean function is
unchanged. The network operation returns a wholly owned result with aggregate
BDD-node and removed-fanin statistics instead of mutating opaque function
pointers or printing diagnostics.

The dense ordering boundary supports up to 20 variables after minimum-base
reduction; wider source nodes with small actual support are accepted, while
larger retained support returns an explicit limit error. Fixed tests prove a
strict mux-tree BDD reduction, exhaustive external-signal semantics, duplicate
fanin tying, unused-input removal, constant collapse, 21-input small-support
handling, ownership, malformed inputs, limits, cancellation and sixteen
concurrent readers. A standalone generated-function model fuzzer passes exactly
5,000 executions. The targeted race gate passes in 1.072s, ordinary tests pass
in 75.418s, vet is clean, and the scheduled full-package race passes in
407.673s. Final source and policy audits remain Go-only and clean; Wave 48 is
the next scheduled full race.

## Forty-fourth serial-wave checkpoint

This single-threaded wave ports all three definitions in `abcBidec.c` over the
concrete native HOP and BDC managers. Per-node resynthesis derives the complete
local truth table, performs bi-decomposition and copies complemented BDC nodes
back through HOP structural hashing. The network-level operation processes an
owned sequence of local roots, retains nodes wider than 15 inputs unchanged,
and reports aggregate gain instead of printing. Whole-batch failure rolls back
every newly constructed HOP node.

The power-aware path preserves the source's transition-probability threshold.
Below it, decomposition retains the incompletely specified care set; at or
above it, don't-cares are deterministically completed to zero or one according
to output probability. Caller roots and care tables remain unmodified.

Fixed tests cover exact semantics, both completion polarities, incomplete-care
semantics, ownership, constants and complemented edges through generated cases,
wide-node retention, malformed inputs, limits, cancellation, batch rollback, a
10,000-node HOP cone and sixteen concurrent independent managers. A standalone
model fuzzer passes exactly 5,000 executions. The targeted race gate passes in
1.070s, ordinary tests pass in 75.182s and vet is clean. Wave 42's 395.760s
full-package race remains current; Wave 45 is next. Final source and policy
audits remain Go-only and clean.

## Forty-third serial-wave checkpoint

This single-threaded wave ports all five definitions in `sbdPath.c`. Checked
critical-path construction covers both ordinary mapped LUT levels and canonical
`TimManager` box layouts, preserving slack-one traversal, output-major DFS and
global first-visit behavior. Timing arcs retain the source's unit arc-presence
semantics. The internal-path pass preserves its observable source behavior: it
marks every LUT boundary before revisiting the already-marked root and is
therefore idempotent. Delay tracing returns owned path and aggregate statistics
instead of printing them.

Fixed tests pin exact unboxed and boxed paths, owned results, malformed inputs,
limits, cancellation, a 10,000-LUT iterative chain, reconvergent output-order
behavior and sixteen concurrent readers of shared frozen graphs and mappings.
An independent generated-DAG model fuzzer passes exactly 5,000 executions. The
targeted race gate passes in 1.071s, ordinary tests pass in 75.100s and vet is
clean. Wave 42's 395.760s full-package race remains current; Wave 45 is next.
Final source and policy audits remain Go-only and clean.

## Forty-second serial-wave checkpoint

This single-threaded wave ports all three definitions in `sbdCnf.c`. Checked
truth-to-CNF conversion preserves ordered direct and complement ISOP covers,
native literal polarity and output-clause position. Formatting returns owned
zero-based literal text instead of writing stdout. Translation substitutes an
owned fanin map and flips literals mapped to the pivot. Structured clauses omit
the source routine's trailing empty vector level, and nonzero-input constants
correctly constrain their output rather than taking its zero-input shortcut.

Fixed tests pin constants, native AND clause/text order, translation order and
ownership, and exhaustively validate zero- through six-input clause semantics.
They also cover malformed truth and CNF data, clause/literal/byte/ISOP limits,
cancellation and sixteen concurrent callers. A standalone truth/translation
fuzzer passes exactly 5,000 executions. The targeted race gate passes in
1.057s, ordinary tests pass in 75.120s, vet is clean, and the scheduled
full-package race passes in 395.760s. Final source and policy audits remain
Go-only and clean; Wave 45 is the next scheduled full race.

## Forty-first serial-wave checkpoint

This single-threaded wave ports both definitions in `llb2Bad.c`. Bad-state
construction traverses only primary-output cones, translates complemented AIG
edges into canonical BDD edges and disjoins every primary output. It preserves
the source's non-obvious BDD variable order—latch outputs first, then primary
inputs—despite the Go AIG's opposite CI partition. PI quantification performs
exact existential abstraction over the second partition. Explicit limits and
cancellation replace the native function's inactive timeout argument.

Fixed tests exhaustively compare both operations with independent AIG
simulation, cover constants, empty output sets, a 10,000-buffer cone, invalid
graphs/managers/literals, dimension and resource limits, cancellation and
sixteen concurrent callers with independent managers. A standalone two-output
truth-table fuzzer passes exactly 5,000 executions. The targeted race gate
passes in 1.070s, ordinary tests pass in 75.853s and vet is clean. Wave 39's
396.785s full-package race remains current; Wave 42 remains the next scheduled
full race. Final source and policy audits remain Go-only and clean.

## Fortieth serial-wave checkpoint

This single-threaded wave ports both definitions in `llb2Dump.c`. Reentrant
dummy-name generation preserves the native zero-padded `ff` convention, while
reached-state dumping composes the native Go BDD manager and BLIF exporter to
return owned bytes with the `Reached` output. The fixed-order manager preserves
Boolean semantics without the source's process-wide file output or in-place
heap reordering, so frozen managers remain safe for concurrent readers.

Fixed tests pin names, exact BLIF structure, exhaustive truth-table semantics,
ownership, malformed managers, roots and models, every resource limit,
cancellation and sixteen concurrent readers. A standalone truth-table fuzzer
passes exactly 5,000 executions. The work also fixes two independently tested
CUDD exporter defects: BLIF output declarations now terminate before the first
`.names` directive, and node limits count discovery rather than only completed
postorder nodes. The combined LLB/CUDD race gate passes in 1.064s, ordinary
tests pass in 75.557s and vet is clean. Wave 39's 396.785s full-package race is
current; Wave 42 remains the next scheduled full race. Final source and policy
audits remain Go-only and clean.

## Thirty-ninth serial-wave checkpoint

This single-threaded wave ports both definitions in `sbd.c`. Checked
configuration-variable counting preserves the native level-vector formula,
degree constraint and feasibility cutoff. Deterministic topology enumeration
generalizes the source's fixed diagnostic, interns level vectors with
`insdict.Dict`, preserves insertion order, and returns owned per-stage records
instead of printing. Arithmetic overflow, dimensions, configurations, retained
level entries, work and cancellation are explicitly bounded.

Fixed tests pin formula cases, malformed vectors, overflow, atomic limits,
ownership and the source's original 9-input, 10-node, degree-3 cumulative totals:
2, 3, 5, 9, 16, 29, 54, 102, 194 and 370. Sixteen concurrent enumerators agree,
and an independent linear-deduplication model passes exactly 5,000 fuzz
executions. The final targeted race gate passes in 1.032s, ordinary tests pass in
74.961s, vet is clean, and the scheduled full-package race passes in 396.785s.
Final source and policy audits remain Go-only and clean; Wave 42 is the next
scheduled full race.

## Thirty-eighth serial-wave checkpoint

This single-threaded wave ports the only definition in `AbcApi.cpp`. The source
constructs an otherwise unused MiniSAT 2.2 solver, never reads `Cnf_Dat_t`, and
returns `-1` unconditionally. `ABCCallMiniSat22` preserves that exact observable
stub contract, including accepting nil, rather than misleading callers by
silently routing the dormant API to the package's different native SAT engine.

Fixed tests pin the native result constant, nil and nonnil calls, and CNF
immutability. Thirty-two goroutines make 4,096 calls against one frozen formula,
and a generated-formula fuzzer passes exactly 5,000 executions. The targeted
race gate passes in 1.024s, the ordinary suite passes in 74.744s, and vet is
clean. Wave 36's 395.682s full race remains current; Wave 39 is next. Final
source and policy audits remain Go-only and clean.

## Thirty-seventh serial-wave checkpoint

This single-threaded wave ports both definitions in `abcLog.c` as owned-byte
marshal and parse operations. The active three-line `snl_SAT`, `snl_UNSAT` and
`snl_UNK` format retains bug-free depth, engine token, optional property and
counterexample frame, initial-register bits and frame-major input bits. Typed
errors replace file-open diagnostics, and returned records own their traces.

Two safe corrections close native round-trip holes: paired `NULL` lines remain
a nil counterexample instead of fabricating a zero-width trace, and unsupported
don't-care or garbage characters are rejected instead of silently deleting bit
positions. Valid zero-input traces with nonempty register state now round-trip.
Strict dimensions, byte/frame/bit/engine limits and responsive cancellation
make the parser suitable for untrusted owned input.

Fixed tests pin exact bytes for every status, optional headers, zero-input
traces, ownership, malformed records, every limit, cancellation and 32
concurrent parsers. Generated-record and arbitrary-byte parser fuzzers each pass
exactly 5,000 executions. The targeted race gate passes in 1.026s, the ordinary
suite passes in 74.983s, and vet is clean. Wave 36's 395.682s full-package race
is current; the next scheduled full race is Wave 39. Final source and policy
audits remain Go-only and clean.

## Thirty-sixth serial-wave checkpoint

This single-threaded wave ports all ten definitions in `abcMini.c`. Both
conversion directions preserve MiniAIG constant polarity, complemented edges,
structural hashing, CI/CO order, dangling cleanup and the convention that the
last register-count PIs and POs are latch outputs and inputs. Private mappings
replace destructive `iTemp` scratch. Converted networks retain the `MiniAig`
name and dummy PI/PO names. Unary buffers are semantically collapsed while
ordinary two-input nodes retain native AND behavior.

A concrete synchronized `ABCMiniAIGFrame` replaces the opaque frame adapters,
including input, output and flop-count repartitioning. It owns its network and
returns snapshots, so simultaneous exporters cannot mutate shared state. The
native process-file smoke test is replaced with a checked in-memory MiniAIG
binary round trip, leaving no permanent files or external dependency.

Fixed tests exhaustively compare combinational and sequential semantics, exact
register partitions, names, cleanup, frame ownership, malformed inputs, limits,
cancellation and sixteen concurrent direct and shared-frame exporters. A
standalone generated-DAG fuzzer passes exactly 5,000 executions with exhaustive
assignment comparison across both conversions. The final targeted race gate
passes in 1.028s, the ordinary suite passes in 75.143s, vet is clean, and the
scheduled full-package race milestone passes in 395.682s. Final source and
policy audits remain Go-only and clean; the next scheduled full race is Wave 39.

## Thirty-fifth serial-wave checkpoint

This single-threaded wave ports all eleven definitions in `abcXsim.c`. The
simulator preserves ABC's numeric zero/one/unknown encoding, complemented-edge
inversion, conservative AND propagation, PI/latch/PO ordering, symbolic input
and state modes, and the exact reset `Gia_ManRandom` stream. Owned frame rows
replace `pCopy` scratch mutation and unconditional stdout. Initial-state cycling
returns an owned cloned network with reached latch values, so frozen source
networks remain safe for concurrent callers. The otherwise unused native
three-valued random helper and 0/1/x formatter are covered directly.

Fixed tests pin exact random bits, frame traces, every latch initialization,
zero-frame behavior, result ownership, malformed structures, dimensions,
resource limits, cancellation and sixteen concurrent callers. An independent
two-bit possibility-set evaluator passes exactly 5,000 generated sequential-DAG
fuzz executions across complemented edges and random/symbolic inputs and state.
The targeted race gate passes in 1.069s, the ordinary suite passes in 74.948s,
and vet is clean. Wave 33's 398.582s full-package race remains the current
milestone gate; the next scheduled full race is Wave 36. Final source and policy
audits remain Go-only and clean.

## Thirty-fourth serial-wave checkpoint

This single-threaded wave completes `ivyFraig.c` rather than retaining its
former bounded-support partial disposition. Deterministic packed simulation
forms complement-normalized candidates, and the native SAT orchestrator
certifies full-support constant, equivalent and complemented merges. The owned
result includes the reduced graph, complete valid mappings, certified classes,
statistics and replayable miter models. Proof mode returns an explicit
proved/counterexample/unknown status; counterexamples are replayed. Native
defaults, limits, cancellation, concurrent frozen callers and all 87 named
translation-unit definitions are covered. The optional CUDD fallback is
replaced semantically by complete unbounded native SAT rather than a runtime
dependency.

The wave also ports `abcOrder.c` with transactional CI permutations, reverse
ordering and stable support-first ordering followed by unused CIs. Network
validation now permits semantic terminal permutations while still requiring
exact duplicate-free terminal sets. Finally, `abcBmc.c` composes the concrete
ABC-to-AIG conversion, sequential timeframe expansion and completed IVY FRAIG
engine. It preserves the source routine as a reduction/reporting pass, returns
owned aggregate and per-frame counts instead of printing, and supports symbolic
or constant-zero initial state without inventing a property verdict.

Standalone Go tests exhaustively compare source and reduced semantics, verify
proof/counterexample replay, ordering, ownership, limits, cancellation and
concurrency. CI ordering, BMC and IVY FRAIG each pass authoritative
single-worker runs of exactly 5,000 fuzz executions.
The combined ordering/FRAIG targeted race tests pass in 1.064s and the BMC race
test itself passes in 1.029s. The ordinary suite passes in 75.533s,
vet is clean, and Wave 33's 398.582s full-package race remains the current
milestone gate; the next scheduled full race is Wave 36. Final source and policy
audits remain Go-only and clean.

## Thirty-third serial-wave checkpoint

This single-threaded wave ports all three definitions in `mfsWin.c`. The API
consumes the source's precomputed object-ID-indexed levels explicitly and walks
the concrete `ABCNetwork` fanouts in stable insertion order. It selects a node
when fanout multiplicity exceeds the limit, a combinational output is reached,
or a fanout lies strictly above the start-level window. A private iterative
visited set replaces network traversal-ID mutation, preserving shared frozen
network state and supporting deep windows and concurrent readers.

Fixed tests pin DFS root order, strict level comparison, CO and fanout cutoffs,
source ownership, malformed level vectors and starts, limits, every cancellation
point, 64 concurrent callers and a 10,000-node nonrecursive chain. An independent
recursive model passes exactly 5,000 generated acyclic-network fuzz executions.
The fuzzer first exposed and then verified the repair of an empty-input harness
slice; its ephemeral reproducer was removed and no `testdata/` remains. The
ordinary suite passes in 75.745s, targeted race tests in 1.062s, vet is clean,
and the scheduled full-package race suite passes in 398.582s. Final source and
policy audits remain Go-only and clean.

## Thirty-second serial-wave checkpoint

This single-threaded wave ports all four functions in `fpgaTruth.c`. A concrete
cut DAG retains both edge complements and each node's output phase. Ordered
leaves form an explicit structural boundary. Native packed truth tables replace
temporary CUDD nodes, while deterministic iterative postorder replaces the
source's pointer-sized sign cache and mutable mark bits. Volume counts each
shared non-leaf cut exactly once and includes the root, matching the source.
Neither operation mutates the DAG, so frozen shared cuts are concurrency-safe.

Fixed tests pin exhaustive three-input semantics, shared-node volume, result
ownership, incomplete nodes, duplicate/nil leaves, root-as-leaf rejection,
cycles, a 10,000-node nonrecursive chain, every resource limit, cancellation
points and 64 concurrent callers. An independent Boolean evaluator passes
exactly 5,000 generated-DAG fuzz executions. The ordinary suite passes in
75.585s, targeted race tests in 1.042s, and vet is clean. Wave 33's 398.582s
full-package race is now the current milestone gate; final source and policy
audits remain Go-only and clean.

## Thirty-first serial-wave checkpoint

This single-threaded wave dispositions `ioReadEdif.c` exactly as shipped. Its
public `Io_ReadEdif` prints that the parser does not work and returns `NULL`
before opening the file, allocating a reader or consulting `fCheck`. The Go
owned-byte counterpart therefore returns typed `ErrUnsupported` before reading
bytes, validating options or invoking cancellation. The static
`Io_ReadEdifNetwork` body appears only after that unconditional return and has
no other caller; it is recorded as unreachable rather than revived into support
that ABC itself does not expose.

Fixed tests cover nil, empty, malformed and plausible EDIF bytes, both check
flags, deliberately invalid limits, untouched cancellation callbacks and 64
concurrent callers. A standalone arbitrary-byte fuzzer passes exactly 5,000
executions. The active source contract is validated without claiming EDIF
parsing. The ordinary suite passes in 74.775s, targeted race tests in 1.025s,
and vet is clean. Wave 30's 395.900s full-package race remains the current
milestone gate; final source and policy audits remain Go-only and clean.

## Thirtieth serial-wave checkpoint

This single-threaded wave ports all four functions in `ioReadDsd.c`. The reader
preserves balanced-end discovery, homogeneous separator splitting, the source's
unusual doubling of every original parenthesis and expansion of commas,
parent-before-child object creation, repeated complement nodes, AND/default
comma groups, special multi-input XOR SOP markers and hexadecimal prime truth
tables in native minterm order. Sparse variables retain the source count and
fake-name rules, including `a0` through `z0` for 26 inputs. The returned owned
logic network is named `dsd` and has the single output `F`.

Fixed tests pin helper results, exact object counts/order/functions, native hex
cube order, sparse names, exhaustive Boolean semantics, malformed formulas,
ownership, every resource limit, cancellation and sixteen concurrent
independent readers. The exact `-fuzztime=5000x` Go-only semantic campaign
completed 5,001 scheduled executions with 46 retained cases. `ioReadDsd.c` is
validated. The ordinary suite passes in 75.080s, targeted race tests in 1.027s,
vet is clean, and the scheduled full-package race suite passes in 395.900s.
The final ledger audit also repairs the Wave 28 JSON metadata onto `ioJson.c`
and restores unrelated `aigInter.c` to unstarted without changing counts.
Final source and policy audits remain Go-only and clean.

## Twenty-ninth serial-wave checkpoint

This single-threaded wave ports `ioReadBaf.c` and `ioWriteBaf.c` together. The
immutable `ABCBAFModel` owns the concrete structurally hashed AIG, separate
PI/PO/latch name vectors and each latch's two-bit initialization. Keeping name
categories separate preserves the format's legal PO-name overlap instead of
forcing it through `ABCNetwork`'s unique-name index. Checked bridges convert
representable models to and from the concrete strashed network. Owned bytes
replace files, explicit little-endian words replace host-endian integers, and a
stable comment replaces the source process-clock timestamp.

Fixed tests independently decode and pin every header field and body word,
exhaustively compare sequential graph semantics, verify exact metadata and
round-trip bytes, exercise network bridges, ownership, cross-category name
overlap, malformed/truncated/forward-reference input, unsupported graph forms,
every resource limit, cancellation and sixteen concurrent frozen readers. A
standalone byte-model fuzzer passes exactly 5,000 executions with semantic
comparison and no C. Both BAF files are validated. The ordinary suite passes in
75.165s, targeted race tests in 1.027s, and vet is clean. Wave 27's 393.575s
full-package race remains the current milestone gate; final source and policy
audits remain Go-only and clean.

## Twenty-eighth serial-wave checkpoint

This single-threaded wave ports all twelve active and static helpers in
`ioJson.c`. A concrete immutable `ABCJSONDocument` retains the source's nested
array/object arena and insertion-ordered scalar interning. Parsing preserves the
brace/bracket preprocessing, whitespace-delimited token subset and
all-scalars-as-strings behavior instead of conflating it with the existing
`IONetwork` JSON schema. Formatting reproduces the source indentation and
ordering, neural-layer extraction preserves its fixed labels and traversal
rules, and owned bytes replace filenames and the process-file round-trip test.

Fixed tests cover preprocessing, exact arena shape, interning, stable
parse/format/parse output, every layer mapping, exact extraction, malformed
nesting and pairs, ownership, every resource limit, cancellation and sixteen
concurrent frozen readers. A standalone parse/format stability fuzzer passes
exactly 5,000 executions without C. `ioJson.c` is therefore validated. The
ordinary suite passes in 75.077s, targeted race tests in 1.026s, and vet is
clean. Wave 27's 393.575s full-package race remains the current milestone gate;
final source and policy audits remain Go-only and clean.

## Twenty-seventh serial-wave checkpoint

This single-threaded wave completes `ioReadBblif.c`. The reader parses owned
binary bytes, factors each distinct function once through the native-order
`Dec_Factor` path, inserts the concrete factor graph into the structurally
hashed AIG, removes dangling logic, and returns a checked strashed ABC network.
The verification wrapper returns the existing combinational-equivalence result
rather than printing it, and the test wrapper performs an in-memory semantic
round trip without retaining process files or timing behavior.

Fixed tests cover exact AND behavior, factor reuse, dangling cleanup, names,
ownership, malformed and corrupt input, every resource limit, cancellation,
equivalent and inequivalent CEC results, counterexamples and concurrent frozen
callers. A standalone Go fuzzer independently evaluates both sides of each
three-input binary round trip and passes exactly 5,000 executions. With every
active source function covered, `ioReadBblif.c` is promoted to validated. The
ordinary suite passes in 74.777s, targeted race tests in 1.079s, vet is clean,
and the scheduled full-package race suite passes in 393.575s. Final source and
policy audits remain Go-only and clean.

## Twenty-sixth serial-wave checkpoint

This single-threaded wave ports `Bbl_ManToAbc`, `Bbl_ManDfs_rec` and
`Bbl_ManDfs` from `ioReadBblif.c`. Reconstruction creates a concrete owned SOP
logic network in BBLIF object order, preserves every function and fanin order,
copies the design name, and assigns the source dummy PI/PO names. It accepts
forward references because all objects precede connection, includes dangling
LUTs like the source all-node DFS, and rejects malformed shapes, boxes and
cycles. The Go traversal uses private state rather than permanently consuming
the manager's mark bits, making repeated frozen calls deterministic.

Fixed tests cover forward references, dangling nodes, exact ordering, names,
functions, ownership, cycles, unsupported box flags, limits, cancellation and
concurrent frozen callers. A standalone fuzzer exhaustively compares every
three-input assignment between the BBLIF object model and an independently
converted AIG. `Bbl_ManToAig`, verification/test harnesses and the top-level
reader remain explicit gaps, so `ioReadBblif.c` is only partial. The frozen
ordinary suite passes in 75.129s, targeted race tests in 1.058s, the fuzzer
passes exactly 5,000 executions, and vet is clean. The wave-24 full-package
race remains the current milestone gate. Final source and policy audits remain
Go-only and clean.

## Twenty-fifth serial-wave checkpoint

This single-threaded wave ports both functions in `ioWriteBblif.c` against the
actual concrete `ABCNetwork` and `BblManager` boundaries. Conversion preserves
original sparse object IDs, CI and CO order, source-order fanins, standard and
native `x`/`n` SOP text, and the reachable-node postorder of
`Abc_NtkDfs(..., 0)`; dangling nodes are intentionally omitted. The writer
returns independently owned deterministic binary bytes instead of mutating a
process file. The separate common `IONetwork` BBLIF format path remains
unsupported rather than being conflated with the native ABC-network entry
point.

Fixed tests pin object order, IDs, SOPs, fanins, dangling-node behavior, exact
bytes against a separately constructed BBLIF manager, parse-back identity,
source ownership, malformed/unsupported representations, every resource limit,
cancellation and concurrent frozen callers. An independent Boolean evaluator
checks every three-input assignment during fuzzing. The frozen ordinary suite
passes in 75.015s, targeted race tests in 1.069s, the fuzzer passes exactly
5,000 executions, and vet is clean. The wave-24 full-package race remains the
current milestone gate. Final source and policy audits remain Go-only and clean.

## Twenty-fourth serial-wave checkpoint

This single-threaded wave ports the portable representative-table portion of
`aigRepr.c`: owned lifecycle, strict lower-ID creation, lower-pair selection,
direct and transitive lookup, clearing, counting and mapped-literal phase
correction. Representative links intentionally use separate storage from the
existing AIG choice-successor chains. Construction snapshots object types and
phases, rejects buffered graphs like the source assertion, and enforces object,
work and retained-byte bounds. Transfer, representative-aware duplication,
remapping, TFI checks, rehash, valid-choice marking and mapped-class transfer
remain unstarted. The exact ordered 31-unit AIG build inventory is validated
without implying coverage of every listed unit.

The frozen ordinary suite passes in 74.840s, targeted race tests in 1.054s,
the standalone model fuzzer passes exactly 5,000 executions, and vet is clean.
The catch-up milestone full-package race suite passes in 404.827s. Final source
and policy audits remain Go-only and clean.

## Twenty-third serial-wave checkpoint

This single-threaded wave ports `Kit_GraphToAigInternal` and `Kit_GraphToAig`
from `kitAig.c`. It validates source and graph shape, maps every leaf to an
explicit AIG literal, preserves graph and fanin complement phases and strashes
internal AND-form nodes into an owned clone of the caller's manager. The clone
is a deliberate transactional safety adaptation: failures never mutate or
partially extend the source manager. Object/work/retained-byte bounds,
cancellation, constants/literals, source ownership and concurrent independent
results are tested. `Kit_TruthToAig` remains unstarted because native ISOP/
factoring truth-to-graph conversion is not yet present.

The frozen ordinary suite passes in 74.953s, targeted race tests in 1.026s,
the new fuzzer passes exactly 5,000 executions, and vet is clean. The periodic
full-package race suite last passed at wave 20. Final source and policy audits
remain Go-only and clean.

## Twenty-second serial-wave checkpoint

This single-threaded wave ports the concrete graph core from `kitGraph.c`:
bounded graph/constant/leaf creation, complemented edges, atomic AND/OR/XOR/MUX
append forms, exact <=5-input truth evaluation, graph level and the native
`-100` absent-leaf depth sentinel. Both native XOR/MUX structural forms and
DeMorgan OR encoding are model-checked. The owned graph uses a mutex for safe
frozen readers, strict literal validation, atomic multi-node limits, cumulative
node/work/temporary-byte bounds and responsive cancellation. ISOP/factoring
truth-to-graph, statistics arrays and file utilities remain unstarted.

The frozen ordinary suite passes in 74.458s, targeted race tests in 1.025s,
the new fuzzer passes its exact 5,000-execution invocation, and vet is clean.
The periodic full-package race suite last passed at wave 20. Final source and
policy audits remain Go-only and clean.

## Twenty-first serial-wave checkpoint

This single-threaded wave connects the existing owned TruthTable and canonical
complemented-edge Cloud manager with source-shaped Shannon conversion from
`kitCloud.c`. Truth-to-Cloud construction preserves variable order and sharing,
and rolls the manager back exactly on validation, node-limit, work-limit or
cancellation failure. Cloud-to-truth reconstruction owns its result and polls
inside BDD traversal. Exhaustive small-function equivalence, complement sharing,
rollback, ownership and concurrent independent-manager tests are Go-only. The
exact twelve-unit KIT build inventory is validated; packed mux serialization,
composition, cofactor-support scratch APIs and the other native KIT engines
remain explicit gaps.

The frozen ordinary suite passes in 74.797s, targeted race tests in 1.070s,
the new fuzzer passes exactly 5,000 executions, and vet is clean. The periodic
full-package race suite last passed at wave 20. Final source and policy audits
remain Go-only and clean.

## Twentieth serial-wave checkpoint

The first single-threaded wave ports source-specific pure helpers from
`monotone.c` and `disjunctiveMonotone.c`: native substring-based pending/hint
discovery, ordered remaining-candidate and consequent filtering, ordered vector
difference and unique insertion. Results are owned and operations enforce
cumulative item/work/byte limits, responsive cancellation and atomic errors.
The exact nine-unit live build inventory is validated. Native AIG tester
construction, PDR proof search, recursive disjunction discovery, reporting and
command/network integration remain explicit unstarted gaps.

The frozen ordinary suite passes in 74.996s, targeted race tests in 1.058s,
the new fuzzer passes exactly 5,000 executions, and vet is clean. As the
scheduled milestone, the full package race suite also passes in 397.104s.
Tests and final source/policy audits remain Go-only and clean.

## Nineteenth supervised parallel-wave checkpoint

The final three parallel lanes received independent cross-review. IF review
fixed complement-preserving truth permutation, merge accounting/cancellation,
NaN rejection and bounded LUT-parser scratch. Mapper review corrected full
64-minterm elementary truths and phase transforms, absent timing arcs, timing
invariants and atomic error results. LLB review added object limits and an
independent two-stage clustering model, then rejected partial credit for
`llbInt.h` because the native manager contract is not represented.

IF now covers bounded normalized/complemented cuts, truth permutation, strict
in-memory LUT-library parsing and variable-pin arrival matching. Mapper adds
bounded cut merge/phase, source-shaped ordering, rise/fall arrival and complete
stretched elementary truths. LLB adds owned supplied incidence matrices,
last-use ordering, greedy scheduling and supplied-partition clustering. Three
build inventories are validated and thirteen C units are partial; all absent
full mapping, library, AIG/CUDD/reachability/SAT/GIA/command workflows remain
unstarted.

The frozen combined ordinary suite passes in 74.519s and vet is clean. Targeted
race tests for all changed lanes pass in 1.111s, and all three fuzz targets pass
coordinator reruns of exactly 5,000 executions. Wave 19 follows the documented
race cadence; wave 20 remains the next full-package race milestone. Tests and
final source/policy audits remain Go-only and clean.

## Eighteenth supervised parallel-wave checkpoint

Three disjoint implementation lanes received independent cross-review. ABCI
review made truth tables immutable, added byte/work preflights and corrected
two native defects: an uninitialized total-pair read and decomposition recursion
over nonexistent cofactor-index bits. GIA review added object-level cancellation,
overflow-safe cumulative accounting, exact literal/object semantics and bounded
sequentially closed cone components. SAIG review strengthened exact inventories,
orbit/storage accounting and concurrent reads, and rejected substitute credit
for the absent ternary `saigPhase.c` workflow.

ABCI now covers exact bounded truth measures, quantification and unateness. GIA
adds owned explicit-pattern simulation, counterexample replay, deterministic
failure discovery, switching statistics and cone partitioning. SAIG adds owned
supplied-trace simulation and activity statistics. Only nine C boundaries are
partial; GIA and SAIG build inventories are validated. The other assigned native
random, packed, ternary, heuristic, command and manager workflows remain
unstarted.

The frozen combined ordinary suite passes in 74.600s and vet is clean. Targeted
race tests for all changed lanes pass in 1.078s, and all three new fuzz targets
pass coordinator reruns of exactly 5,000 executions. Wave 18 follows the
documented cadence; the next periodic full-package race checkpoint is wave 20.
Tests and final source/policy audits remain Go-only and clean.

## Seventeenth supervised parallel-wave checkpoint

Three implementation lanes were replaced one-for-one with independent reviews.
FRA review corrected zero-state and final-divergence replay, bounded-proof
labelling, cumulative work accounting, signature allocation, cancellation and
atomic sweep statistics. ACEC review added cumulative expansion bounds,
deadlock-free lifecycle callbacks, safe concurrent close and arithmetic,
ownership and canonicalization regressions. PDR review enforced exact
observation partitions, non-compounding archive statistics, atomic clause
operations, state-only invariants, malformed-witness rejection and cumulative
preflight accounting.

FRA now provides standalone replay, bounded proof, simulation, equivalence
classes and verified sweep behavior. ACEC adds exact bounded arithmetic
equivalence with owned lifecycle management. The remaining PDR layer adds
explicit generalization, incremental archives, ternary simulation and exact
invariant/witness checks. Native ABC graph, CNF and SAT engines, six distinct
FRA workflows and native PDR CNF construction remain explicit gaps.

The frozen combined ordinary suite passes in 74.289s, the package-wide race
suite in 394.329s, and vet is clean. All three new fuzz targets pass coordinator
reruns of exactly 5,000 executions. Tests remain Go-only; the final source and
policy audits are clean.

## Sixteenth supervised parallel-wave checkpoint

Three implementation lanes were replaced one-for-one with independent reviews.
ABS review closed a proof-soundness hole where a tampered concrete witness could
pass after only abstract replay, then bounded construction, refinement, search,
proof replay and conversions. CEC review tightened trailing-frame rejection,
overflow-safe allocation, witness ownership and output-class accounting. SSW
review strengthened partition/node-kind validation, atomic cancellation,
signature/class/sweep/pattern memory and work accounting, and downgraded three
unsupported workflow claims.

ABS now provides owned explicit-state abstraction, existential reachability,
monotone refinement, real/spurious discrimination, exact witness/proof replay
and bounded conversions. CEC adds exact combinational and paired-state bounded
sequential equivalence, aligned miters, shortest replayable witnesses and output
classes. SSW adds owned sequential graphs, simulation/signatures/classes,
bounded equivalence, replayable counterexamples and phase-aware verified sweep
mappings. Native proof engines and distinct absent workflows remain explicit.

The frozen combined ordinary suite passes in 74.629s, the package-wide race
suite in 395.767s, and vet is clean. All three new fuzz targets pass coordinator
reruns of exactly 5,000 executions. Tests and policy audits remain Go-only and
clean.

## Fifteenth supervised parallel-wave checkpoint

Three implementation lanes were replaced one-for-one with independent reviews.
WLC review hardened source-shape and sequential invariants, topology/list
integrity, cumulative analysis budgets, snapshots, cancellation and lifecycle,
and removed an inaccurate source claim. Ivy review corrected cumulative map/
equivalence/choice budgets, unary forward-edge validation, nil safety, bounded
attributes, sequential/choice memory and balance growth. NWK review corrected
constant nodes, traversal order, filename rules, invariants, support budgets,
timing numerics, derived limits, merge ownership and output-rooted depth.

WLC now provides an owned <=64-bit word/sequential network, exact type values,
checked operators, graph analyses and a loss-rejecting combinational Verilog
bridge. Ivy adds owned attributes, profiles, buffers/assertions, sequential
step, cuts/truth, MFFC, bounded mapping, verified choices, truth canonicalization,
balancing/resynthesis and DOT. NWK adds owned graphs, truth/traversals/support,
timing, normalization, strash, direct cuts and merge matching. Broad native
engines and integrations remain explicit partial or unstarted ledger gaps.

The frozen combined ordinary suite passes in 74.286s, the package-wide race
suite in 392.900s, and vet is clean. All three new fuzz targets pass coordinator
reruns of exactly 5,000 executions. Tests and policy audits remain Go-only and
clean.

## Fourteenth supervised parallel-wave checkpoint

Three implementation lanes were replaced one-for-one with three independent
reviews. Command review corrected disabled history substitution, aggregate
registry/session accounting, atomic multi-name mutations, enumeration polling,
nil contexts and autotuner bounds. Main review added nil safety, atomic mode and
network updates, stable IDs, overflow/copy budgets, snapshot preflights, text
cleanup and source-shaped CLI precedence. I/O review corrected cumulative
budgets, constant latches, BLIF off-set behavior, graph canonicality, parse
depth/structure, bounded onset/CNF generation, constant CNF and SMV semantics.

Command coverage now includes a deterministic registry/session, parser,
aliases, flags, history, nested in-memory sources, builtins and configuration.
Main adds owned frame/runtime lifecycle, network history, libraries,
verification/session/log state, snapshots and CLI utilities. I/O adds an owned
Boolean/latch network and reviewed subsets of BLIF, Bench, equations, PLA,
structural Verilog plus deterministic CNF/DOT/GML/List/SMV projections. Eleven
unsupported I/O units, including the unrelated generic JSON implementation,
remain unstarted rather than receiving substitute credit.

The frozen combined ordinary suite passes in 74.024s, the package-wide race
suite in 392.377s, and vet is clean. All three new fuzz targets pass coordinator
reruns of at least 5,000 executions. Tests and policy audits remain Go-only and
clean.

## Thirteenth supervised parallel-wave checkpoint

Three implementation lanes were replaced one-for-one with three independent
reviews. ACB review corrected interface levels, zero-slack critical paths,
range overflow, aggregate ownership/accounting and bounded set operations. BAC
review added exact binary preflights, atomic cancellable commits, strict
Verilog/BLIF structure, bounded writers, metadata reconciliation and lifecycle
safety while narrowing primitive coverage claims. CBA review corrected range
overflow, exclusive design ownership, stored shapes, direct outputs/constants,
declarations, cube validation, trailing input and cumulative parse budgets.

ACB now provides owned manager/network/object/range/attribute behavior, truth,
traversals, support/MFFC, profiles, normalization and canonical set algebra.
BAC adds stable hierarchical word networks, a concrete primitive subset,
portable binary encoding, bounded structural Verilog/scalar BLIF and cloning.
CBA adds exact type numbering, owned designs/networks, broad word operations,
structural validation and bounded Boolean BLIF/Verilog round trips. All native
frame, full grammar, advanced synthesis and representation gaps remain explicit.

The frozen combined ordinary suite passes in 73.046s, the package-wide race
suite in 412.147s, and vet is clean. All three new fuzz targets pass coordinator
reruns of exactly 5,000 executions. Tests and policy audits remain Go-only and
clean.

## Twelfth supervised parallel-wave checkpoint

Three implementation lanes were replaced one-for-one with three independent
review lanes. BBR review made transition construction bounded and cancellable,
made memory arithmetic overflow-safe, completed iteration-limited onion rings
and preflighted state-set BDD construction. ExtraBDD review corrected half-open
range cubes and empty-family subset behavior, added canonical BDD/ZDD graph
validation, cumulative budgets, non-expanding membership and deterministic
work counts. REO review corrected terminal-unit objectives, compacted unused
support slots, made profile limits cumulative and hardened unit-DAG transfer.

BBR coverage now includes partitioned image/preimage, explicit-state
reachability, onion rings, shortest replayed counterexamples and state-set BDD
construction. ExtraBDD adds support/cubes, n-ary operations, transfer/image,
translation spaces, profiles, symmetry/unateness, threshold checks, Gray/K-map
and extensive fixed-order family algebra. REO adds owned multi-output transfer,
profiles, sift/shuffle/swap and canonical unit-DAG round trips. Exact native
schedulers, encodings, manager mutation and process drivers remain explicit
ledger gaps.

The frozen combined ordinary suite passes in 73.238s, the package-wide race
suite in 385.398s, and vet is clean. All three new fuzz targets pass coordinator
reruns of exactly 5,000 executions. Tests and policy audits remain Go-only and
clean.

## Eleventh supervised parallel-wave checkpoint

Three implementation lanes were again replaced one-for-one with independent
reviews, never exceeding three sub-agents. I/O review corrected additive
Harwell duplicates, dimension and overflow bounds, cumulative transfer
accounting, malformed/cyclic graph rejection, export identifiers and
exponential preflights. Matrix review confirmed all source operations and
added noncontiguous-variable, duplicate-variable and NaN rollback regressions.
ZDD reorder review added complete malformed-DAG validation, bounded canonical
decode, validation-memory accounting, deeper cancellation polling, leaner
symmetry checks and source-aligned path formatting.

Bridge/I/O coverage now includes bounded ADD/BDD conversions, transactional
manager transfer, sparse and RUA matrix parsing, and five deterministic export
formats. Matrix coverage adds arithmetic, times-plus, min-plus triangulation
and pointwise outer-sum relaxation on the concrete ADD manager. ZDD reordering
adds immutable order cost/reconstruction, sift/window/exact/group searches,
interactions, symmetry, GF(2) transforms and owned path/cover utilities.
Native stream, layout, group-tree and table-mutation gaps remain explicit.

The frozen combined ordinary suite passes in 72.309s, the package-wide race
suite in 385.561s, and vet is clean. All three new fuzz targets pass coordinator
reruns of exactly 5,000 executions. Tests and policy audits remain Go-only and
clean.

## Tenth supervised parallel-wave checkpoint

Three implementation lanes were replaced one-for-one by three independent
review lanes, never exceeding the three-sub-agent ceiling. Approximation review
corrected compatible projection, threshold preservation, cumulative budgets,
constant-path cancellation and predicate-build preflights. Reordering review
added aggregate allocation and root bounds, rebuild cancellation, cycle
detection, stable sifting visits, exact range handling and value-owned linear
maps. ZDD review corrected terminal DAG sizing, replaced exponential count and
support expansion with DAG traversal, made family expansion accountable and
made BDD/ZDD transfers cumulative, cancellable and transactional.

Approximation now provides deterministic bounded subset/superset modes,
clipping, priority predicates, projection, Hamming distance and signatures.
Reordering adds fixed-order semantic cost extraction, interactions, symmetry,
groups, linear maps, level queues and bounded sift/window/exact/anneal/genetic
search. ZDD coverage adds a canonical zero-suppressed manager, set algebra,
unate product/quotient, count/support, BDD transfer and correct bounded
full-minterm ISOP construction. Conservative ledger caveats retain every
source-specific or integration-heavy gap.

The frozen combined ordinary suite passes in 70.886s, the package-wide race
suite in 382.453s, and vet is clean. All three new fuzz targets pass coordinator
reruns of exactly 5,000 executions. Tests and policy audits remain Go-only and
clean.

## Ninth supervised parallel-wave checkpoint

Three implementation lanes and three independent reviews again respected the
three-sub-agent ceiling. Hierarchy review corrected absent MV defaults, copied
domains, XOR/XNOR markers, module/instance/port invariants, hidden recursion,
aggregate atomicity and relation/serialization/name budgets. Foundation review
corrected CUDD variable-role values, cache-key canonicality and constructor
cleanup. Analysis review restored true restrict semantics, vacuous essential
literals, constant clause sentinels and literal-set deduplication.

Hierarchy coverage now includes owned MV relations, modules, ports, instances,
recursive validation, deterministic serialization/flattening, trace names,
flat evaluation, GIA mapping and bounded exact CEC. The CUDD foundation adds
concrete BDD/ADD lifecycle integration, variable metadata, references, caches,
unique-table checks and exact build inventories. Analysis adds essential
literals, correlation, decomposition, constrain/restrict, cubes/paths,
unateness, splitting and zero-equation solving.

The frozen combined ordinary suite passes in 53.072s, the package-wide race
suite in 274.193s, and vet is clean. All three new fuzz targets pass coordinator
reruns of at least 5,000 executions; CUDD analysis additionally exhausts all
65,536 four-variable functions. Tests and policy audits remain Go-only and
clean.

## Eighth supervised parallel-wave checkpoint

Three implementation lanes and three independent reviews stayed within the
three-sub-agent ceiling. Functional review corrected malformed latch/box
rejection, exponential duplicate-fanin allocation, round-trip name
preservation and shared-root duplication. Utility review corrected node-only
statistics, source-order crosscuts, dangling-cone reference handling, barrier
rollback and report/DOT limits. ADD review replaced dense query allocations
with bounded memoized DAG traversal and made Walsh/residue construction honor
cumulative operation, byte and cancellation budgets with rollback.

ABC functional coverage now includes owned SOP/network-to-AIG and AIG-to-ABC
conversion, phases, sequential partitions, mappings, combinational
logic/netlist round trips and bounded SOP minimum-base reduction. Utilities add
statistics, collections, crosscuts, terminal ordering, barrier buffers,
library/design metadata and deterministic DOT/reports. The ADD manager adds
canonical numeric terminals, apply/ITE, abstractions, extrema, inverse,
negation, logarithm, rounding and Walsh/residue construction.

The frozen combined ordinary suite passes in 53.261s, the package-wide race
suite in 266.866s, and vet is clean. All three new fuzz targets pass coordinator
reruns of at least 5,000 executions. Tests and policy audits remain Go-only and
clean.

## Seventh supervised parallel-wave checkpoint

Three disjoint implementation lanes and three independent reviews again used
no more than three sub-agents. ABC review corrected terminal partition checks,
sequential self-feed removal, non-strashed constant creation, rollback after
validation/cancellation and SOP size-overflow handling. FXU review corrected
cumulative work accounting, post-update normalization polling and final-step
variable capacity. Gordian review corrected incremental tree reopening and
`Done` validation, changed-core rejection, quadratic edge limits and mid-net
cancellation.

ABC metadata now provides deterministic names, latch triples/initialization,
self-feed handling, phase-preserving fan ordering and owned SOP/cube parsing,
truth, construction, normalization and recognition. FXU flow adds owned cover
conversion, pair preprocessing, deterministic single/double extraction,
transactional semantic updates, cost checks and reporting. Gordian placement
adds owned partition trees, deterministic refinement, QP construction,
center-of-gravity constraints and bounded global/incremental flows.

The frozen combined ordinary suite passes in 52.080s, the package-wide race
suite in 264.661s, and vet is clean. All three new fuzz targets pass coordinator
reruns of at least 5,000 executions. Tests and policy audits remain Go-only and
clean.

## Sixth supervised parallel-wave checkpoint

Three disjoint implementation lanes and three independent reviews again used
no more than three sub-agents. ABC review corrected signal name ordering and
combinational latch comparison, malformed latch triples, transparent-net
reverse traversal/levels and collect-all ordering. FXU review corrected stale
derived divisors, pair preconditions, transactional accounting, scan budgets,
divisor limits and full heap/pair validation. FRAIG review corrected choice
phase/chain/TFI handling, MUX normalization, supergate expansion/uniqueness,
solver/feedback budgets, cancellation composition and SAT statistics.

ABC analysis now provides owned structural/latch/signal/cycle checking,
forward/reverse traversal, support, levels, reference snapshots and MFFC
analysis. FRAIG orchestration adds concrete manager APIs, SAT-backed relations
with replay, choices, feedback simulation, traversal, supergates and MUX/XOR
recognition. FXU adds stable-ID sparse incidence, ordered lists, single/double
divisors, canonical pairs, deterministic heaps and selection. Integration and
native-heuristic gaps remain explicit in the partial ledger entries.

The frozen combined ordinary suite passes in 51.737s, the package-wide race
suite in 263.703s, and vet is clean. All three new fuzz targets pass coordinator
reruns of at least 5,000 executions. Tests and policy audits remain Go-only and
clean.

## Fifth supervised parallel-wave checkpoint

Three disjoint implementation lanes and three independent reviews ran with no
more than three sub-agents active. Review and integration corrected twelve
issue classes: ABC edge-phase composition, insertion phases, cancellation
rollback and constant-one invariants; FRAIG disabled-table accounting,
simulation phase handling, fanout-transfer rollback and pool-size overflow;
DSD terminal-free DFS, source-order/large-prime variable collection and
transactional limit polling; and quadratic exhaustive-test manager growth.
Every correction has focused regression coverage.

The ABC foundation now supplies concrete owned objects/networks, stable IDs,
names, terminal/box/latch partitions, fanin/fanout multiplicity, references,
replacement, cloning and lifecycle. FRAIG adds concrete complemented nodes,
vectors, fixed pools, structural and bounded exact functional interning,
simulation and fanout transfer. DSD adds owned complemented constant, buffer,
OR, XOR and prime nodes, decomposition, support, local truth, reconstruction,
metrics and bounded BDD entry. Conservative partial dispositions preserve the
remaining representation-specific, SAT/choice and full CUDD gaps.

The frozen combined ordinary suite passes in 50.356s, the package-wide race
suite in 256.353s, and vet is clean. All three new fuzz targets pass coordinator
reruns of at least 5,000 executions; DSD additionally reconstructs all 65,536
four-variable functions and passes a 5,000-case six-variable model. Tests and
policy audits remain Go-only and clean.

## Fourth supervised parallel-wave checkpoint

Twenty already-dispatched implementation lanes completed before concurrency
was reduced to one supervisor and three sub-agents. Three independent review
groups then reconciled every assigned source definition and corrected seven
defects: transactional CUT statistics, SATOKO cancellation/budget reporting,
ACEC packed-index overflow, ambiguous MIO leading constants, CUDD adjacent
permutation semantics, CUDD arithmetic option propagation and placement text
spacing. Each correction has a regression test.

New validated coverage includes AIG cuts and factoring; the active Ivy memory
and operation core; SAIG cone handling; KIT truth, permutation, SOP, PLA and
Cloud semantics; five CUT kernel units; MIO parsing/SOP; MPM MIG/truth; COV and
EXOR foundations; placement density/I/O/pads; CUDD Boolean abstraction; live
constraint/arena transformations; and ACEC polynomial utilities. Broader
manager headers, exact native heuristics, frame/network adapters, ADD behavior,
sequential cut paths and subsystem orchestration stay partial with explicit
caveats.

The frozen combined ordinary suite passes in 43.241s, the package-wide race
suite in 218.083s, and vet is clean. All 23 new fuzz targets passed coordinator
reruns of at least 5,000 executions. Source hashes, aggregate source identity,
Go-file references, formatting, package headers, native-source, `testdata/`,
symlink and tool-allowlist policy checks are clean. Tests use only Go-owned
models and datasets.

## Third supervised parallel-wave checkpoint

Twenty disjoint implementation clusters were independently reviewed before
coordinator-owned ledger promotion. Review corrected or strengthened NUL and
limit handling, DEC normalization and atomicity, retiming preimages and PI lag,
rewrite rollback and library traversal, interpolation constraints and budgets,
MiniAIG layouts, CSW allocation and signature order, XSAT parsing and literal
limits, clock-gating polarity and ODC handling, FXCH cleanup and bounds, PLA
consensus and type validation, IOA latch initialization, WLN delay/NDR and
signed-shift behavior, SSC object limits, DCH choice rebuilding, Lucky parsing,
SIM structural symmetry, RES graph validation, HOP transfer/balance/truth
bounds, and AMAP constants, phases, ties and output mappings.

Validated coverage now includes the complete active CSW, IOA and DCH clusters;
most DEC, PLA and HOP utility/manager behavior; MiniAIG operators and build
inventory; concrete WLN object/network/retiming/writer behavior; Lucky fast
paths and swaps; SIM manager/SAT utilities; and the stand-alone portions of
retiming, rewriting, interpolation, SSC, RES and AMAP. Broad headers and
integration-heavy units stay partial where the Go implementation intentionally
uses explicit-state or concrete stand-alone models instead of the absent native
ABC object ecosystem.

The final combined ordinary suite passes in 31.820s, the package-wide race
suite in 162.056s, and vet is clean. All 26 new fuzz targets ran for at least
5,000 executions against the frozen combined tree. Tests remain Go-only, use
independent models or permanent datasets, and add no C/cgo/subprocess oracle or
`testdata/` directory.

## Second supervised parallel-wave checkpoint

Sixteen disjoint implementation lanes and independent review lanes ran under a
20-sub-agent ceiling. Reviewers found and corrected semantic defects in
single-word union storage, DSC match direction, integer-vector atomicity and
native capacity/order, proof-root identity and resolution, selector bit order,
EPD subnormal/exception handling, MTR validation, RWT lifecycle/rollover, and
structural-Verilog formal/blackbox handling. Shared coverage and status files
remained coordinator-owned until all reviews completed.

New validated coverage includes `extra.h`, `vecInt.h` and its empty aggregate
metadata, `deco.h`, the complete DSC cluster, all active `sat/proof` behavior,
the EPD and RPO clusters, RWT's exact 3,562-word static forest and 1,139 retained
decompositions, the MTR cluster and the active six-variable RSB path. The
structural-Verilog formula, parse, stream and build-manifest files are also
validated after mapped-cell, undefined-box and hierarchy finalization work.

Conservative partial classifications remain explicit. CSAT substitutes the Go
CDCL core for IvyProve and does not implement eight declared gate kinds or
byte-identical strashed BENCH. `m114p.h` uses `SatSolver2` because MiniSat 1.14p
is absent. CAS omits CUDD symmetry sifting and non-strict encoding choices.
INT2 supplies checked finite-state reachability rather than the source's broken
SAT-interpolation internals. FRET currently provides only stand-alone lag,
min-cut and forward-frontier scaffolding; `fretInit.c` remains unstarted. These
files were not promoted merely because their safe substitutes are well tested.

The final combined ordinary suite passes in 27.290s, the package-wide race
suite in 138.659s, and vet is clean. Every new fuzz target ran for at least
5,000 executions. Go-only independent models, exhaustive checks, proof replay,
BLIF round trips and fixed static-stream hashes replace native runtime oracles.
Package-header, native-source, `testdata/`, symlink and tool-allowlist scans are
clean; `pkg/tools/` contains only `source_manifest.py`.

## First supervised parallel-wave checkpoint

Fourteen disjoint implementation lanes were followed by independent review,
source-definition reconciliation and coordinator-owned integration. Reviewers
found and corrected source-polarity, SAT-clause accounting, vector growth,
frontier allocation, RNG, cancellation, parser, capacity, sorting and
supergate-accounting defects that worker-local suites did not expose. Broad
BDC and Verilog submissions were not accepted on their initial claims: BDC was
completed and retested; Verilog remains explicitly partial.

Newly validated coverage includes `vecStr.h` (68 definitions), `vecWrd.h` (71),
`vecPtr.h` (67), `extraUtilPrime.c` (31), `cnf.h` and its manifest, six remaining
FSIM files, six BDC files, six supergate files, `abcDarUnfold2.c`, `cgtSat.c`,
`nwkCheck.c`, `rwrTemp.c`, and the two AMAP manager/core units. Permanent tests
are Go-only; no C/cgo/subprocess oracle or new `testdata/` directory was added.

The final ordinary suite passes in 22.357s, the package-wide race suite in
117.660s, and vet is clean. Every new fuzz target ran for at least 5,000
executions, including independent vector, prime, CNF, FSIM scalar, BDC
decomposition/SPFD, supergate, Verilog parser, NWK, unfold, SAT implication,
rewrite and AMAP campaigns. Package-header, native-source, `testdata/`, symlink
and tool-allowlist policy audits remain clean; `pkg/tools/` contains only
`source_manifest.py`.

## Exact-synthesis CEGAR checkpoint

`src/sat/bmc/bmcMaj.c` is validated across all 36 active definitions.
`ExactMajorityValue` and
`ExactMajorityTruth` preserve the source's strict-majority target.
`BuildMajorityExactTopology` preserves its two constant objects, hard-wired
first MAJ3 gate, optional constant candidates, cascade pin, exact connection
variable numbering and fanout-use lists. `BuildLUTExactTopology` consolidates
the two Exa managers while retaining normalized nonzero truth variables,
2-through-6-input candidate ranges, fewer-variable restrictions, exact
function/connection variable order and fanout-use lists.

`ExactNetwork` is an owned topological MAJ3/LUT representation with explicit
constants and output complementation. It returns exhaustive truth, the source's
deterministic first or majority-middle-layer mismatch, owned solution text and
owned generalized BLIF bytes instead of printing or choosing a global output
filename. Validation checks topology, duplicate/forward fanins, truth width,
object/variable/assignment/text bounds and cancellation without interfaces.

`ExactSynthesisSession` builds exact-one fanin selection, adjacent-pin
symmetry, optional node ordering, mandatory-use and normalized two-input
function clauses, including the source `OnlyAND` restriction. `AddMinterm`
atomically appends connection and MAJ3/LUT function CNF at a checked variable
boundary. Every intermediate SAT model decodes into an independently
evaluable network. `SynthesizeExactNetwork` and `SynthesizeExactMajority`
preserve target complementation, initial-minterm choice and iterative
solve/decode/mismatch refinement, returning tri-state status, the exact
network, refinement minterms and solver statistics.

`ExactLazySynthesisSession` completes the alternate Exa3 path. It preallocates
the source's per-minterm/node values, guards function clauses by the fanins
selected in the current model, caches `(minterm,node,fanins)` tuples with a
lookup-only map, preserves tried/added counters, permits the same minterm to
recur under a different topology, and commits clauses/cache entries atomically.
`SynthesizeExactNetworkLazy` exposes this repeated-minterm CEGAR path for two-
and three-input LUTs.

Go-only tests exhaust majority targets through ten inputs, pin exact topology
layouts and SAT-variable IDs, replay majority and LUT networks, complements,
mismatch selection, cloning, formatting, BLIF rows and every error surface.
Projection tests check every normalized two-input truth after every incremental
minterm. Four independent Go-only campaigns passed 5,000 fuzz executions each
for network truth, minterm CNF, full synthesis and lazy synthesis. Cache hits,
limits, repeated minterms and atomic cancellation are pinned without a native
oracle. The complete suite passes in 16.391s, the race suite in 85.268s, and vet
is clean. Inventory is now 263 validated, 39 partial and 1,455 unstarted files.

## Sparse exact-synthesis topology checkpoint

`src/sat/bmc/bmcMaj2.c` is validated across all 36 active definitions. Its
duplicated majority, LUT2/LUT3, CNF and CEGAR definitions reuse the concrete
exact-synthesis implementation above. `GenerateMajorityExactConnections`
adds this source's missing behavior: every non-output MAJ3 gate receives a
later consumer and an explicitly bounded number of distinct extra edges.
Generation uses a caller seed rather than C's global clock-seeded random state,
and impossible requests return an error instead of retrying forever.
`BuildSparseMajorityExactTopology` preserves fixed internal pins, input-only
residual candidates, constants, cascade precedence and source variable order.
The dormant fixed connection table is retained as owned Go data.

`ExactLUTPreferredVariables` preserves the optional Exa3 AND-chain phase hint,
including cyclic input selection and previous-node pin-zero selection;
`InitializeExactLUTPolarity` applies it explicitly to the concrete solver. The
source itself comments this call out, so default synthesis remains unchanged.
Go-only unit tests check deterministic connections, exact pin domains,
mandatory consumers, fixed data, phase choices and complete synthesized truth.
A separate sparse-topology campaign passed 5,000 fuzz executions. No compiler,
C source, generated native executable or runtime native oracle is used, and
`pkg/tools/` still contains only the source-manifest script. Inventory is now
264 validated, 39 partial and 1,454 unstarted files.

## Level-constrained exact-majority checkpoint

The 16-definition Maj3 half of `src/sat/bmc/bmcMaj3.c` is now ported.
`BuildLevelMajorityTopology` preserves the source's reversed level allocation,
fixed first MAJ3, fixed top fanins, inter-level cascade, lower-level candidate
domains and selector numbering. `LevelMajoritySession` owns its concrete solver
and preserves initial fanin/fanout clauses, fresh per-minterm intermediate
values and atomic majority implications. `RefineTopologyModel` adds the same
lazy too-few/too-many fanin blockers until every decoded node has exactly three
fanin objects. `SynthesizeLevelMajority` composes this into the middle-layer
CEGAR loop and returns an independently evaluable `ExactNetwork`.

Go-only tests pin the exact `{1,2,1}` layout, selector domains, both lazy
refinement directions, decoded networks, all 32 minterms and complete
five-input majority synthesis. Limits, invalid distributions, ownership and
atomic cancellation are covered. A separate topology campaign passed 5,000
fuzz executions. This intermediate checkpoint covered 16 definitions; the
remaining Zyx work is completed below.

## Generic subset exact-synthesis completion checkpoint

The remaining 27 Zyx definitions complete `src/sat/bmc/bmcMaj3.c`.
`SubsetExactTopology` preserves the source's contiguous function, topology and
all-minterm evaluation variable regions. `NewSubsetExactSession` fixes unused
function/topology variables, primary-input values and final target values;
pins the majority first gate and final-node predecessor; and installs mandatory
fanin/fanout plus normalized LUT2/`OnlyAND` clauses.

`RefineTopologyModel` preserves lazy exact-k subset cardinality, adjacent-node
fanin/function ordering and symmetric-input first-use ordering. Incremental
functionality caches selected LUT2/LUT3 tuples and majority pairs using a
lookup-only map; wider LUTs preserve the source's uncached emission, and eager
mode emits every bounded LUT2/LUT3 tuple. Topology-only solution blocking,
decode, majority/LUT evaluation, target complementation and bounded enumeration
are composed by `SynthesizeSubsetExact`. Concrete topology and returned
statistics replace variable-map, clause and iteration stdout.

`MarshalExactSolutionDataset` returns the compact records previously appended
to a derived global filename. `ParseExactSolutionName` and
`VerifyExactSolutionDataset` replace the native filename parser and file-based
truth verifier with owned bytes, strict validation, checked limits and
cancellation. The permanent XOR fixture is
`test_vectors/bmcMaj3/6-2-1.bool`; there are no native source/runtime oracles.

Go-only tests cover exact IDs and constants, too-few/too-many topology models,
node and symmetric-input ordering, LUT and majority cache hits, eager clauses,
decode, complementation, topology enumeration, serializers, malformed data,
limits and atomic cancellation. Independent level-topology, subset-synthesis
and dataset-parser campaigns each passed 5,000 fuzz executions. All 43 active
definitions are accounted for, promoting `bmcMaj3.c` to validated. Inventory
is now 265 validated, 39 partial and 1,453 unstarted files. The complete suite
passes in 16.427s, the race suite in 84.628s, and vet is clean.

The `src/sat/bmc/module.make` manifest is also reconciled: 31 of its 32 selected
translation units are validated. It is partial rather than unstarted, with only
the already tracked `bmcMulti.c` dependency substitution preventing promotion.
The authoritative inventory is therefore 265 validated, 40 partial and 1,452
unstarted files.

## Fault-cardinality checkpoint

The default-parameter and cardinality portion of `sat/bmc/bmcFault.c` is now
ported, moving the file from unstarted to partial. `AddSatCardinalityConstraint`
provides the recursive pairwise sorting network, the general odd/even
transposition network and the tree at-most-one construction over the concrete
Go CDCL solver. Exact Boolean gate relations preserve threshold semantics;
checked extensions also define `k=0` and `k=n`. Construction owns its inputs and
outputs, is atomic on invalid input, logical resource exhaustion and synchronous
cancellation, retains active bookmarks, and reports generated variables,
clauses, literals and contradiction without interfaces or process-global state.

Go-only tests exhaust all assignments through eight inputs for both sorting
networks and through nine inputs for the tree, including exact-one behavior.
They also cover existing clauses, contradiction and rollback, determinism,
ownership and every error surface. A fresh independent popcount oracle passed
5,000 fuzz executions. Formula parsing/realization, CNF lifting, all fault
unfold/cofactor transforms, test persistence and the actual fault campaign are
still outstanding, so `bmcFault.c` is not validated. Inventory is now 261
validated, 40 partial and 1,456 unstarted files. The complete suite passes in
16.497s, the race suite in 85.571s, and vet is clean.

## Fault-formula checkpoint

The formula subsystem of `sat/bmc/bmcFault.c` is now represented by the
concrete `FaultFormula` AST, checked parsing, direct evaluation, atomic GIA
realization, structural formula derivation and formula-driven fault unfold.
The parser supports the intended two functional inputs and all eight declared
parameters, plus negation, AND, OR, XOR, mux and constants. It rejects trailing
or ambiguous syntax that the source token recursion could silently ignore, and
group negation keeps derived formulas semantically correct for complemented
internal logic.

`UnfoldGiaFaultFormula` applies the formula to every ordinary AND or only direct
register-input drivers. Parameter inputs are allocated before register outputs,
preserving the sequential PI/RO partition, and the result owns exact source-gate
to parameter-input correspondence. Names, register count and constraints are
preserved. Native XOR/mux/buffer source gates are rejected rather than silently
treated as ordinary ANDs. Bounds cover formula bytes/nodes, graph objects and
synchronous cancellation.

Go-only tests exhaust parser/evaluator and GIA realization truth tables, replay
derived formulas against native GIA evaluation, verify full and register-only
unfold semantics, and cover atomic failures and limits. A fresh independent
formula model passed 5,000 fuzz executions. This accounts for ten more active
definitions while `bmcFault.c` correctly remains partial; CNF lifting, the
remaining fault transforms/cofactor, persistence and campaign are outstanding.
The complete suite passes in 16.478s, the race suite in 85.772s, and vet is
clean.

## Structural fault-transform checkpoint

Five more `sat/bmc/bmcFault.c` definitions are now implemented.
`UnfoldGiaStuckAt`, `UnfoldGiaFlip` and `UnfoldGiaFourOutputFunction` allocate
deterministic gate-order control inputs and preserve inactive mapped slots.
The four-output-function model realizes an arbitrary two-input truth table as
nested muxes while using edge polarity to identify the nominal AND minterm; the
Go port deliberately enables the final gate's fourth control instead of
retaining the source loop's off-by-one disable. `CofactorGiaFaultInputs` fixes a
checked primary-input prefix while retaining the original CI count and
sequential metadata.

`UnfoldGiaTemporalFault` constructs the source's two-frame relation with
explicit frame-zero ROs/PIs, frame-one PIs and gate- or RI-freeze controls.
The result is intentionally combinational. Single-frame transforms place fault
controls between ordinary PIs and ROs so copied register metadata remains
correct. All transforms are owned, bounded and cancellable, preserve names and
constraints where applicable, and reject extended gates rather than silently
applying ordinary-AND semantics.

Go-only tests exhaust every input/control assignment for stuck-at, flip and
four-function transforms, replay PI cofactoring, cover every temporal mode and
verify sequential partitioning. An independent flip model passed 5,000 fuzz
executions. This brings the `bmcFault.c` ledger to 24 covered definitions while
the file remains partial; CNF lifting, persistence/result analysis and the
fault campaign remain outstanding. The complete suite passes in 16.539s, the
race suite in 85.391s, and vet is clean.

## Fault-CNF checkpoint

The two local CNF helpers in `sat/bmc/bmcFault.c` are now implemented.
`DeriveGiaFaultCNF` lowers a checked GIA through the owned AIG conversion and
reference Tseitin encoder, then composes a complete GIA-object-to-SAT-literal
map across structural merging, complements, buffers, native XOR/mux lowering
and CO aliases. Inputs, outputs, clauses and mappings are returned as owned
data; no optimized-cut layout leaks into the helper's semantic contract.

`GiaFaultCNF.Lift` replaces the source's in-place literal/mapping mutation with
an immutable checked transformation. It shifts every clause, object, input and
output literal, expands the declared variable span, preserves polarity and
rejects overflow, resource exhaustion and cancellation atomically. The lifted
relation loads directly into the concrete solver without an interface.

Go-only tests exhaust every assignment and both output polarities on a graph
containing native XOR, mux and complemented signals, then repeat the projection
checks after a nonzero lift. They cover exact shift arithmetic, object mapping,
ownership and all limit/cancellation paths. A fresh SAT-projection fuzzer passed
5,000 executions. The `bmcFault.c` ledger now covers 26 definitions; persistence,
result analysis and campaign orchestration remain. The complete suite passes in
16.669s, the race suite in 84.550s, and vet is clean.

## Fault-pattern and reporting checkpoint

Six more `sat/bmc/bmcFault.c` support definitions are complete.
`MarshalFaultPatterns` and `ParseFaultPatterns` are bounded pure codecs, with
checked `ReadFaultPatterns`/`WriteFaultPatterns` wrappers for explicit caller
paths. `MarshalFaultDelayPatterns` converts `[RO0, PI0, PI1]` test rows into two
conventional `[PI,state]` lines by replaying the graph transition relation,
without borrowing object marks. `FaultRunStatistics` returns an owned record
instead of appending a process-global statistics file.

`DuplicateGiaWithPrimaryInputs` replaces clone-then-append behavior with a
checked rebuild that inserts extra PIs before the RO suffix, preserving names,
constraints, register roles, extended gates and source ownership. Tests cover
exact text, whitespace parsing, checked files in temporary directories,
sequential delay expansion, duplicate semantics, metadata, ownership, limits
and cancellation. A Go-only codec fuzzer passed 5,000 executions. The
`bmcFault.c` ledger now covers 32 definitions; untestable-fault enumeration,
parameter analysis, incremental cofactor addition, new-fault comparison,
preparation and the top-level campaign remain. The complete suite passes in
16.377s, the race suite in 85.338s, and vet is clean.

## Reusable fault-campaign SAT checkpoint

Three more `sat/bmc/bmcFault.c` definitions are complete.
`AnalyzeFaultParameters` preserves source-order positive-assumption checks,
UNSAT deactivation and same-model skipping while returning an owned active map
and exact run/UNSAT/unknown/covered counts. Logical `SatLimits` and synchronous
cancellation replace the fixed conflict constant and printed timing.

`EnumerateFaultAssignments` clones the input solver, enumerates and blocks each
complete parameter projection, and returns enabled parameter indices plus final
tri-state status instead of mutating the campaign solver or writing an ambient
file. `AddFaultPatternInstance` atomically cofactors functional inputs,
derives/lifts/loads a new owned relation, asserts either at least one detected
output or equality of every output, and connects non-functional instance PIs
to existing base parameter variables with exact equivalences.

Focused tests cover UNSAT pruning, model coverage, bounded enumeration, caller
solver preservation, all four one-input XOR pattern/detection cases, atomic
errors and resource cancellation. A Go-only incremental-instance fuzzer passed
5,000 executions. The `bmcFault.c` ledger now covers 35 definitions; only
new-fault comparison, complete preparation and top-level orchestration remain.

## Complete fault-campaign workflow

`src/sat/bmc/bmcFault.c` is now validated across all 38 active definitions.
`PrepareGiaFaultCampaign` selects and builds each of the formula, temporal,
stuck-at, complement and four-output-function fault relations, constructs the
golden/faulty miter, derives and loads its owned CNF, asserts observable
difference, disables analyzed controls, applies basic or general cardinality,
and ingests caller or explicitly named pattern data. It preserves the source's
warm-up and prefix-truncation behavior for seed and all-zero/all-one start
patterns. Deterministic SAT budgets replace the source wall-clock deadline.

`FindGiaFaultNewAssignments` cofactors a candidate pattern, requires detection,
excludes every prior pattern and returns bounded enabled-parameter projections
instead of writing `newfaults.txt`. `RunGiaFaultCampaign` completes iterative
pattern generation, periodic parameter analysis and reconstruction, explicit
iteration stopping, new-fault mode, untestable/rectification queries, bounded
assignment enumeration, pattern encoding and run statistics. Dump options
return owned bytes or records; no ambient file is created. Sources remain
immutable, results own their storage, and all object, variable, clause,
literal, pattern, iteration, assignment and cancellation failures are checked.

Focused Go-only tests cover all five algorithms, both cardinality paths,
seed/start-pattern semantics, periodic deactivation/rebuild, explicit pattern
files, iteration limits, new-fault and untestable enumeration, structured
output, ownership, malformed dimensions, resource limits and cancellation. An
independent exhaustive stuck-at model passed 5,000 fuzz executions, proving
that every detectable allowed assignment is covered by the returned suite.
The complete suite passes in 16.344s, the race suite in 84.921s, and vet is
clean. Inventory is now 262 validated, 39 partial and 1,456 unstarted files.

## FX divisor discovery

`sat/bmc/bmcFx.c` is now validated across all 20 active definitions.
`FindFXDivisors` preserves residual merging, AND/OR/XOR/mux normalization,
accumulated costs and deterministic ranking. `ExtractFXSharing` completes the
source's onset enumeration and repeated cover/discover/add loop: prior derived
signals remain correlated candidates, whole divisor batches receive exact gate
clauses in reverse selection order, and every iteration is returned as owned
data. `AnalyzeFXOutputs` replaces duplicated print-only trials with onset and
offset covers plus exact count tables.

Tests exhaust every gate truth table, classifier, iterative cover, ownership,
all resource surfaces and a 10,000-gate raw chain. Two Go-only fuzz campaigns
passed 5,000 executions, with every intermediate cover replayed against an
exhaustive truth table. The complete suite passes in 16.409s, race in 84.430s
and vet is clean. Inventory is now 261 validated, 39 partial and 1,457
unstarted files.

## SAT-guided SOP collapse

`sat/bmc/bmcClp.c` is now validated across all 18 active definitions.
`OutputSensitivity` and `AnalyzeOutputSensitivity` return exact single-CI
derivatives and caller-seeded statistics without manager scratch, fixed 64-bit
arrays, ambient randomness or stdout. `CollapseGiaOutputToSOP` consolidates the
file's final and experimental solver variants over concrete Go CNF/CDCL while
preserving alternating on/off-set enumeration, first-completed phase,
canonical/reverse ordering, prime-implicant expansion, cube blocking,
constants and full irredundancy.

Unit tests cover exact statistics, both phases, all ordering modes, ownership,
all resource surfaces and a 10,000-gate raw chain. Two Go-only fuzz campaigns
passed 5,000 executions against exhaustive three-input truth tables. The full
suite passes in 16.152s, race in 83.993s and vet is clean. Inventory is now 260
validated, 39 partial and 1,458 unstarted files.

## BMC public header

`sat/bmc/bmc.h` is now validated. Six concrete Go parameter records account for
all five C parameter structures plus the opaque unroller role; constructors pin
the inline exact-synthesis defaults and the distinct `abc9bmc`, `abc9sbmc`,
`abc9bcore` and `abc9mult` command defaults. Concrete function callbacks replace
C function pointers without interfaces. All 22 external declarations map to
the already validated BMC core, scalable, sparse-AND, counterexample,
M-inductiveness and unrolling APIs. Exact-synthesis and fault algorithms remain
accurately unstarted in their own `bmcMaj*.c` and `bmcFault.c` ledger entries;
validating the shared parameter contract does not claim those algorithms.

Focused parameter tests pass, as do the complete suite in 15.862s, the race
suite in 84.004s and vet. Inventory is now 259 validated, 39 partial and 1,459
unstarted files.

## Stand-alone testing policy

The Go module no longer contains retained C oracle programs. The only maintained
tool in `pkg/tools/` is `source_manifest.py`, which inventories the upstream
source tree and does not compile or execute ABC. Permanent expected results and input
corpora live under `test_vectors/`; tests also use independent Go models,
exhaustive small-state checks, round trips, invariants and fuzzing. They do not
invoke an ABC executable, ABC archive, C compiler, cgo or a generated native
helper.

`TestStandaloneOracleDatasets` enforces this boundary on every test run. It
rejects C/C++ sources and headers anywhere in the Go module, rejects a
`testdata/` directory, verifies every permanent oracle dataset against the
reviewed SHA-256 manifest, and rejects untracked files in the oracle dataset
directory. The focused stand-alone dataset and AIGER round-trip checks pass.

The historical C capture helpers have all been replaced by direct Go tests or
immutable datasets under `test_vectors/`. Dataset-backed foundation, GIA mux
and PDR tests also check their rows with exhaustive Boolean semantics or
reachability, so the rows are not trusted merely because of their historical
origin. Stale “native fixture” test names and diagnostics have been removed.
The complete suite passes in 16.014s, focused boundary and replacement checks
pass in 0.120s, vet is clean, and an explicit module scan finds no C/C++ source,
header, test subprocess oracle or `testdata/` directory.

The boundary has now been hardened further: Go tests are rejected if they
import or invoke subprocess machinery or contain generator hooks, module
symlinks are rejected, and `pkg/tools/` may contain maintained Python scripts
but no generated caches, executables or other artifacts. The stray Python
bytecode cache was removed. The plan now explicitly forbids an ongoing
“refresh from C” workflow: historical vectors are immutable reviewed evidence
and every validated slice must have a Go model, property, exhaustive check,
replay/certificate or other stand-alone semantic justification. This
policy-only hardening does not change the source-file coverage counts.
The focused boundary checks pass in 0.077s, the complete Go suite passes in
16.428s, vet is clean, and a fresh filesystem scan reports only
`source_manifest.py` under `pkg/tools/` with no native source or `testdata/`
directory anywhere in the Go module.

`sat/bmc/bmcMaxi.c` is now validated across all four active definitions. The
concrete `GiaInitialClass` preserves all six source encodings, while
`UnfoldMaxInitial` owns its controls, data variables, frame inputs and final
register functions without mutating GIA scratch. `MaximizeInitialState` builds
the fixed/variable unfold miter through native GIA-to-AIG and reference-CNF
conversion, repeatedly shrinks final UNSAT assumption conflicts, and relaxes
exactly the known bits absent from the fixed-point core. Returned status,
reason, statistics and core registers replace stdout and wall-clock reporting;
resource limits, cancellation and unsupported extended-node errors replace
unchecked allocation and ordinary-AND misinterpretation. Unit tests exhaust
all 36 class pairs in both modes over every two-frame assignment, independently
replay transition semantics and generalized cores, and cover lifecycle,
limits, cancellation and a 10,000-gate cone. A fresh Go-only fuzz campaign
passed 5,054 executions. The complete suite passes in 15.721s, race in 82.361s,
and vet is clean.

`sat/bmc/bmcInse.c` is now validated across all ten active definitions,
including its three inline simulation-arena helpers. The concrete insertion
simulation owns separate definitely-zero and definitely-one object bitsets,
preserves the exact reset random stream and earliest maximum-known pattern tie,
and never borrows GIA scratch. `DiscoverInsertionInitialState` makes the
all-zero wrapper assumption explicit when its initial state is nil, while also
supporting the source perform routine's concrete starting state. It returns
the frame-major selected inputs, ternary target, selected patterns, unknown
counts and final `GiaInitialClass` values instead of stdout and timing. Ignored
timeout/simulation arguments are omitted; limits, cancellation, zero-PI
handling and checked errors replace assertions and unsafe allocation. Unit
tests compare every packed round with independent sample-by-sample ternary
evaluation, replay the selected trace concretely, and cover fixed/unknown
results, ownership, malformed options, resource caps, cancellation and a
10,000-gate cone. A fresh Go-only fuzz campaign passed 5,000 executions. The
complete suite passes in 15.756s, race in 82.333s, and vet is clean.

The paired `sat/bmc/bmcMesh.c` and `bmcMesh2.c` files are now validated across
all seven active definitions in each file. They differ only in Satoko versus
bsat calls, so one concrete `EmbedGiaMesh` implementation uses the owned Go SAT
solver without a backend interface. It preserves cell-major time, graph,
configuration and used-variable blocks; boundary restrictions; ordered
neighbor configurations; structural fanin implications; the source's `G-1`
target at cell `(1,1)` even when it is not the CO driver; and iterative
pairwise one-hot refinement. A concrete placement with cells, counts, status,
reason and statistics replaces both stdout reports. Checked canonical GIA
topology, dimensions, resource limits and cancellation replace fixed 102×102
arrays and assertions. Tests cover PI buffering, a two-input node, an
impossible short route, a successful multi-level 4×4 route, exact refinement
clauses, the target quirk and all error surfaces. A structural-oracle fuzzer
passed 5,000 executions while checking every SAT placement cell-by-cell. The
complete suite passes in 15.573s, race in 82.623s, and vet is clean.

`sat/bmc/bmcEnum.c` is now validated across all three active definitions.
`DeriveAssignmentJustification` preserves forward phase propagation, reverse
controlling-value marking and phase-specific AND/OR/passthrough construction
without borrowing object phase, mark or `Value` scratch; the zero-assignment
wrapper is explicit. The incremental enumerator shares original CI variables,
adds exact Tseitin definitions for each owned justification, and carries its
negated selected root into later assumptions. It therefore excludes the whole
covered formula as the source does, rather than merely blocking one model.
Assignments, observed outputs, graphs, completion, SAT status/reason/statistics
and construction counts replace stdout. Checked per-graph, total-corpus, CNF
and assignment limits plus cancellation replace assertions and the fixed
10,000-iteration loop. A marked constant is rejected because the source's
literal-zero temporary value cannot encode the required true equality
predicate. Exhaustive tests prove every formula implies its originating output
phase and that enumeration covers exactly each requested XOR phase; a
10,000-gate cone remains iterative. A fresh Go-only implication fuzzer passed
5,000 executions. The complete suite passes in 15.740s, race in 82.270s, and
vet is clean.

`sat/bmc/bmcBCore.c` is now validated across all six active definitions.
Checked whitespace parsing and file loading replace the unchecked pivot reader;
valid object IDs expand across the deterministic SAT-variable map while invalid
IDs are returned instead of printed. `BuildGiaBoundedCoreCNF` preserves the
source's object/frame variable order, frame-local DFS order, constant-zero
units, paired output and initial-state constraints, RI/RO buffers and ordinary
AND clauses without borrowing traversal marks or `Value` scratch. The traversal
is iterative and extended buffers/XORs/muxes are rejected rather than silently
encoded as ordinary ANDs. `BoundedCore` uses the concrete Go CDCL engine and its
checked resolution DAG; UNSAT results contain both an independently re-solved
root-clause core and every reachable learned clause through the final empty
clause. Owned formatting and checked file output replace stdout. Source pivot
variables remain observable; ABC's bsat stores that list but never reads it.
Unit tests exhaust the small combinational CNF projection, check alternating
sequential reachability, independently solve reported cores, verify exact maps,
parsing, formatting, ownership, limits, cancellation and a 10,000-gate cone.
A fresh Go-only Boolean-model fuzz campaign passed 5,000 executions. The
complete suite passes in 15.883s, race in 82.174s, and vet is clean.

`sat/bmc/bmcExpand.c` is now validated across all three active definitions.
Concrete owned `SOPCover`, `SOPNode` and `SOPNetwork` values replace mutable
character arenas and the unported pointer-rich ABC network. Cube expansion
preserves source cube order, arbitrary input-column mapping and reverse-column
greedy literal removal against the GIA output offset. Constant phase detection,
direct-order full irredundancy and shared output-driver updates are retained.
The network wrapper also reproduces the final five-stage SOP readability sort,
including fanin permutation. Ordinary phase codes and special `n`/`x` parity
codes survive parsing and formatting. The existing concrete Go AIG CNF and CDCL
solver replace optimized native CNF without a solver interface. Resource
limits, construction cancellation and atomic SAT `UNKNOWN` replace assertions,
unchecked storage and the source's partially cleared cube on timeout. Unit
tests check exact removal/order, constants, parsing/parity, full network sorting,
ownership, every failure surface and a 10,000-gate cone. An independent
exhaustive truth-table model checks exact greedy expansion and direct-order
irredundancy; a fresh Go-only fuzz campaign passed 5,000 executions. The
complete suite passes in 15.708s, race in 82.129s, and vet is clean.

The build-disabled `sat/msat/msatOrderJ.c` alternative is now validated across
all 11 definitions. The separate concrete `MsatJBoundaryOrder` owns cone
membership, assignments, activities, adjacency and its circular insertion-
ordered boundary; it does not add an ordering interface or borrow solver
arrays. Assignment and unassignment update affected neighborhoods, selection
uses strict activity comparison with stable boundary-order ties, and an
outside-cone assignment cannot seed the boundary. Unit tests cover exact ring
transitions, ties, quiescent configuration changes, ownership, invalid
operations and lifecycle. A
fresh independent state-machine model passed 5,000 fuzz executions. The full
suite passes in 15.490s, race in 81.962s and vet is clean.

The MSAT module is now closed at the file-ledger level: every translation unit
selected by `sat/msat/module.make` is validated, and the manifest itself plus
`msat.h` and `msatInt.h` are reconciled. All header declarations with linked
definitions map to concrete Go types and operations. The six external solver
declarations, obsolete `Msat_VarHeap` family, and three fake/truth clause
declarations that have no definition anywhere in this source snapshot are
recorded explicitly as stale declaration-only API; the port does not invent
behavior for them. This completes the selected and build-disabled MSAT source
scope without introducing interfaces, native linkage, or borrowed C layouts.

`sat/lsat/solver.h` is also fully dispositioned. It is a declaration-only
compatibility header with no inline logic, build-manifest membership, callers,
or symbol definitions anywhere in this source snapshot. No Go facade is added:
doing so would invent a solver implementation rather than port source behavior;
the validated concrete bsat and MSAT packages already provide native SAT
functionality.

`aig/gia/giaIiff.h` is an empty guard-only header with no declarations, inline
logic, data, includes or callers and is now recorded as a validated no-runtime
disposition.

`aig/gia/giaSatMap.c` is now validated across all 13 active definitions.
`GiaSatMapper` combines the already exact pairwise cardinality network with an
owned `GiaSatMapProblem` that replaces the source's untyped Nf extraction hook.
It checks current-solution closure, builds object/cut and inverter-exclusivity
clauses in source order, adds fanin implications and preferred phases, applies
root/padding/cardinality assumptions, solves repeated bounds, and decodes
objects, input inverters, cuts, gates and checked area. Decreasing optimization
stops on UNSAT or UNKNOWN; this deliberately avoids the source diagnostic
loop's nontermination on UNKNOWN. Tests cover exact emitted clauses, closure,
SAT/UNSAT recovery, optimization, ownership, limits, cancellation and
lifecycle. A fresh independent chain/inverter cardinality model passed 5,000
fuzz executions. The full suite passes in 15.546s, race in 82.292s and vet is
clean.

`sat/bmc/bmcBmc.c` is now validated across all five active definitions.
`BuildBmcFrames` constructs zero-initialized combinational timeframes with
frame-major primary inputs and outputs and preserves the source's unusual
completed-frame reachable-AND threshold. Iterative owned marks count each
shared cone node once across all outputs and frames. The existing concrete
`AigerModel.BoundedCheck` supplies ordered SAT queries and replayed witnesses;
its explicit latch initialization and positive constraint prefixes replace the
C-only rewrite/cofactor, optimized-CNF and alternate-backend switches with one
native Go path. The solver's owned result model subsumes the local Satoko model
copy helper. Tests cover exhaustive sequential semantics, exact cutoff, shared
cones, ownership, limits, cancellation and 10,000-deep graphs; an independent
step model passed 5,000 fuzz executions. The full suite passes in 15.744s,
race in 82.188s and vet is clean.

`sat/bmc/bmcLoad.c` is now validated across all six active definitions. The
concrete owned `GiaLazyCNF` preserves reserved variable zero, constant-zero
initialization, object-variable allocation order and exact positive/negative
directional AND clauses. Eager mode loads complete cones. Interface-free lazy
mode starts with the requested output direction, checks every SAT candidate
against direct graph semantics, loads the direction of any inconsistent gate
and resolves; a directional-subset UNSAT answer is already sound. Atomic
loading, copied clauses and graph ownership, checked resource limits,
cancellation, lifecycle and returned statistics replace graph scratch, solver
callbacks and stdout. Exact clause tests and all 256 three-input Boolean
functions pass, as do 5,000 independently evaluated fuzz cases. The full suite
passes in 15.816s, race in 81.762s and vet is clean.

`sat/bmc/bmcGen.c` is now validated across all six active definitions. Owned
`GiaPackedSimulation` preserves the object-major arena, exact reset
`Gia_ManRandomW` stream, complemented-fanin word propagation and CO copying
without graph scratch. `EnumerateSATDiagnostic` explicitly retains the source's
odd active behavior: simulation-derived polarity candidates are computed but
not applied because its polarity call is commented out, so the same
output-constrained formula is solved for every simulation bit. Go returns bits,
statuses, counts and statistics instead of printing them. Extended node kinds
are rejected rather than silently treated as ordinary ANDs. Tests pin exact RNG
words, compare every packed bit with scalar evaluation, exercise repeated
SAT/UNSAT/UNKNOWN outcomes, limits, cancellation and a 10,000-gate cone. The
independent simulation fuzzer passed 5,000 cases; the full suite passes in
15.778s, race in 81.598s and vet is clean.

The concrete MSAT solver slice now validates all 34 active definitions in
`sat/msat/msatSolverApi.c` and all 13 in `msatSolverSearch.c`, joining the
already validated core, I/O and reader. `MsatSolver` retains MSAT's statuses,
literal-encoded models, allocation/clean/Prepare lifecycle, stored-clause
numbering, Type-A/start marking, factors, progress, statistics, assumptions and
exact checked budgets. Concrete copied state covers decision levels, reasons,
trails, learned clauses, watched IDs, memory, seen epochs, allocation counters,
adjacency, decision cones and used-variable masks. A shared enqueue mask now
preserves source filtering for both direct assumptions and propagated
implications without adding an interface to the CDCL engine.

Direct search operations cover assume/enqueue, watched propagation, conflict
snapshots, first-UIP analysis, learned recording, one/all-level cancellation,
reduction, learned removal and recently-added original rollback. Rebuilds keep
all live original constraints, including unit clauses, and reject malformed
state atomically. Unit tests exercise these transitions and copied nested
views. A fresh independent unit-propagation model passed 5,022 incremental
operation sequences. The complete suite passes in 15.768s, race in 81.839s and
vet is clean.

Assignment, clause/activity and statistics diagnostics return deterministic
owned strings. DIMACS supports injected timestamps, level-zero assertions and
bounded parse/file round trips; an inconsistent solver deliberately emits an
empty clause instead of the C writer's potentially satisfiable output. Unit
tests cover lifecycle, resizing, normalization, accessors, SAT, UNSAT,
assumptions, copied results, truth tables, diagnostics and file/parser paths. A
fresh exhaustive four-variable enumeration model passed 5,008 generated
formulas and re-solved every DIMACS round trip. No native oracle or retained
fixture is used.

`sat/msat/msatMem.c` is now validated across all 15 active definitions. Concrete
owned `MsatFixedArena`, `MsatFlexArena` and `MsatStepArena` types preserve the
fixed manager's 1,024-entry/64-KiB sizing rule, LIFO reuse and first-chunk
restart; MSAT's distinct 4-KiB flexible chunks and two-times-large-request
growth; and the step manager's 8-byte power-of-two classes, exact routing,
immediate oversized release and class-only usage total. Writable stable handles
become invalid after recycle, restart or close. Checked resource limits,
ownership validation, atomic errors, copied statistics and deterministic
diagnostics replace raw pointers, assertions and verbose stop printing. Unit
tests cover sizing boundaries, reuse, growth, accounting, diagnostics, limits,
ownership and lifecycle. Independent fixed-state, flexible-packing and
step-class models each passed 5,000 fuzz executions. The complete suite passes
in 15.482s, race in 81.596s, vet is clean, and no native oracle or retained
fixture is used.

`sat/msat/msatClause.c` is now validated across all 21 active definitions.
`MsatClauseContext` concretely shares assignment, reason, watch, activity and
decision-order state without interfaces. It covers original and learned clause
creation, top-level reduction/enqueue, stable fields, activity, locking,
two-watch propagation, simplification, reasons, watch removal, lifecycle and
both diagnostic formats. The source's explicitly documented duplicate-polarity
bug is intentionally corrected: repeated equal literals collapse, while both
polarities of one variable make an original clause satisfied. Inputs and
literal snapshots do not alias owned storage; checked limits/invariants and
atomic learned-activity rollback replace assertions and partial mutation. Unit
tests cover all operations and error paths. Independent Go normalization/truth
and watched-propagation models each passed 5,000 fuzz executions. The complete
suite passes in 15.615s, race in 81.064s, vet is clean, and no native oracle or
retained generated fixture is used.

`sat/msat/msatActivity.c` is now validated across all six definitions.
`MsatActivityState` concretely joins the owned decision heap and learned-clause
vector without an interface. It preserves negative-decay static ordering,
float64 variable increments/decays, heap updates, the 1e100 variable threshold,
float32 clause additions, clause decay and the 1e20 learned-database rescale.
Finite-arithmetic checks and prevalidated learned clauses make failures atomic;
diagnostic snapshots copy all arrays. Tests cover ordinary and static bumps,
both automatic thresholds, explicit rescaling, rounding, ownership, malformed
databases and lifecycle. A fresh independent recurrence model passed 5,000
mixed activity operations while checking heap invariants. The complete suite
passes in 15.711s, race in 81.450s, vet is clean, and no native oracle or
retained fixture is used.

`sat/msat/msatOrderH.c`, the implementation selected by `module.make`, is now
validated across all 15 definitions. `MsatVariableOrder` owns the one-based
activity max-heap, reverse index, assignment flags and activities required by
the next MSAT stages. It preserves strict-greater ties, verbatim clean layout,
destructive selection, lazy removal of assigned variables, unassignment
reinsertion, upward bump updates and both percolation paths. Uniform positive
rescaling keeps heap order. Checked bounds and atomic validation replace
assertions; snapshots do not alias storage. Tests cover exact tie order,
selection/update/reinsert behavior, lazy assigned removal, resizing, rescaling,
ownership, invariants and lifecycle. A fresh independent heap/activity model
passed 5,000 operation sequences. The complete suite passes in 15.836s, race
in 86.042s, vet is clean, and no native oracle or retained fixture is used.

`sat/msat/msatSort.c` is now validated across all six definitions. Learned
clause sorting operates directly on stable `MsatClause` pointers and
`MsatClauseVector` storage, preserving the fixed MiniSat pseudo-random pivot,
selection-sort cutoff at 15, recursive partitioning, float32 comparisons and
the deterministic reversal of equal-activity runs. NaN behavior remains the
source's comparison behavior; nil and closed clauses reject atomically before
mutation. Tests cover every size through 80, cutoff boundaries, duplicates,
equal runs, infinities, NaN, identity/permutation, vector integration and
errors. A fresh independent ordering/permutation fuzzer passed 5,000 inputs.
The complete suite passes in 15.778s, race in 85.416s, vet is clean, and no
native oracle or retained generated fixture is used.

`sat/msat/msatClauseVec.c` is now validated across all 11 definitions. The
bounded owned `MsatClauseVector` preserves native small-capacity rounding,
exact explicit growth, doubling push, shrink/clear, pop and indexed access.
Snapshots copy the pointer array so callers cannot reorder internal storage,
while the stable clause pointers retain identity for later watched lists and
reason arrays; nil entries remain representable as in the source. Checked
errors replace assertions and idempotent close replaces manual free. Unit tests
cover every operation, capacity transition, identity/alias rule, exact limit
and lifecycle. A fresh independent pointer-slice model passed 5,000 operation
sequences. The complete suite passes in 15.886s, race in 85.125s, vet is clean,
and no native oracle or retained fixture is used.

`sat/msat/msatVec.c` is now validated across all 23 definitions. The bounded
owned `MsatIntVector` preserves native small-capacity rounding, exact growth,
doubling push, fill/shrink/clear, ordered and unordered uniqueness, pop and
signed sorting. Explicit take, transfer and release operations replace raw
array ownership moves; ordinary reads return copies rather than writable
aliases. Checked errors replace every assertion, and idempotent close replaces
manual free. Unit tests cover construction, capacity transitions, ownership,
all accessors/mutators, order modes, exact limits and lifecycle. A fresh
independent slice model passed 5,010 arbitrary operation sequences. The
complete suite passes in 15.355s, race in 86.676s, vet is clean, and the
ephemeral fuzz reproducer was removed rather than retained under `testdata/`.

`sat/msat/msatQueue.c` is now validated across all six active definitions.
`MsatQueue` owns a fixed-capacity propagation epoch and preserves the native
first/last cursor rule: extraction does not make an append slot reusable until
`Clear`. Typed `(literal, found, error)` extraction removes the ambiguous `-1`
empty sentinel while retaining arbitrary integer payloads. Checked capacity,
copied pending snapshots and idempotent close replace assertions and manual
allocation. Unit tests cover FIFO order, zero/exact capacity, clear reuse,
ownership, lifecycle and all failures. A fresh ordered-slice state model passed
5,000 fuzz executions. The complete suite passes in 15.984s, race in 86.554s,
vet is clean, and no native oracle or fixture is retained.

`proof/fraig/fraigPrime.c` is now validated in full. The fixed FRAIG simulation-
hash schedule is represented by the first 1,024 primes, with checked indexed
access and copied snapshots so callers cannot mutate package storage. A
stand-alone sieve reconstructs and compares every entry, including the final
prime 8161; boundary and ownership tests cover the public API. A fresh Go-only
fuzzer independently checks primality and ordinal rank for 5,000 inputs. The
complete suite passes in 15.680s, race in 85.682s, vet is clean, and the tests
retain neither generated C source nor a native runtime oracle.

`bool/bdc/bdcTable.c` is now validated across all four active definitions.
`BdcTable` owns truth functions and support-indexed newest-first chains while
preserving direct-before-complement lookup, the distinguished constant-one
function zero, exact support buckets and occupied-bucket clearing. Checked
dimensions, disjoint ISFs, copied truth storage, function limits and atomic
cancellation replace assertions and borrowed function pointers. Exhaustive
three-variable testing checks every Boolean function against every disjoint
on/off pair; order, complement, constants, lifecycle, aliasing and error paths
are covered separately. A fresh independent Go model passed 5,012 fuzz
executions. The complete suite passes in 15.551s, race in 81.827s, vet is
clean, and no native oracle or retained generated fixture is used.

`base/cmd/cmdHist.c` is now validated across all four active definitions.
`CommandHistory` preserves the exact prefix/question filters, one-trailing-
newline removal, batch suppression and last-ten duplicate move-to-end rule.
Read/write use caller-selected paths and bounded atomic byte loading on every
platform instead of silently accessing a fixed filename only on one platform;
print behavior returns deterministic text. Unit tests cover every filter,
recent and old duplicates, limits, file round trips, ownership and malformed
loads. A fresh 5,000-case Go-only campaign agrees with an independent history
model. The complete suite passes in 10.569s, race in 60.597s, and vet is clean.

`base/cmd/cmdAlias.c` is now validated across all four active definitions.
`CommandAliases` owns copied alias names and argument vectors in a bounded
deterministic `insdict.Dict`. It preserves first-token lookup and unchanged
fallback for unknown commands; formatting retains the source's 15-column name
field and argument order. Duplicate/capacity failures are atomic, snapshots do
not alias storage, and Go lifetime plus table deletion replace standalone free.
Unit tests cover exact formatting, ownership, empty names, invalid expansions,
capacity and delete/reinsert order. A fresh 5,000-case Go-only operation
campaign agrees with an independent ordered-slice model. The complete suite
passes in 10.568s, race in 60.525s, and vet is clean.

`base/cmd/cmdFlag.c` is now validated across its three active definitions.
`CommandFlags` is a concrete bounded `insdict.Dict`, retaining deterministic
insertion order for later command diagnostics. Read, replace, null-value-to-
empty update and delete/missing-delete behavior are represented directly.
Replacement works at exact capacity, new insertion limits are atomic, and
returned ordered entries do not alias owned storage. Unit tests cover empty
keys, all error paths, capacity and delete/reinsert order. A fresh 5,000-case
Go-only operation campaign agrees with an independent ordered-slice model. The
complete suite passes in 10.659s, race in 60.541s, and vet is clean.

`bool/rsb/rsbMan.c` is now validated across all four active definitions.
`RsbManager` owns the exact counterexample, decomposition-pattern, result,
original-fanin and intermediate scratch families required by the later RSB
decomposer. Checked construction covers the source's 20-leaf and 200-divisor
bounds and exact truth-word sizing. Explicit atomic replacement plus copied
fanin snapshots replace mutable vector aliases; idempotent close replaces all
five frees. Unit tests cover every leaf count, boundary/error atomicity,
ownership and lifecycle. A fresh 5,000-case Go-only state/alias fuzz campaign
passes. The complete suite passes in 10.536s, race in 60.705s, vet is clean,
and the standalone-oracle guard remains clean.

`sat/bmc/bmcUnroll.c` is now validated. The existing simple unroller is joined
by a concrete incremental ranked manager with iterative tent/rank discovery,
typed circular value histories, atomic frame appends, immutable owned source
storage, copied rank/statistics results and checked object/rank/mapping limits.
It preserves zero-initialized registers and frame-major primary inputs/outputs;
native buffer, XOR and mux nodes are supported as a Go extension. Exhaustive
four-frame sequential tests compare it with the independently implemented
simple unroller, while lifecycle, ownership, cancellation and exact-limit tests
cover failures without partial state. A fresh Go-only mixed-gate fuzz campaign
passed 5,004 executions. The complete suite passes in 10.792s, race in 60.607s,
vet is clean, and the standalone-oracle guard passes.

`sat/cnf/cnfMan.c` is now validated. Owned mapped and graph-free CNF values
cover allocation/clone/lift/flip/polarity, diagnostics, DIMACS/gzip output,
legacy loading, output OR/AND and XOR constraints. Native-order frame relations
now load atomically into either legacy BSAT or the proof-capable solver2 facade;
solver2 clauses receive stable positive IDs and may be proof-enabled after load.
The solver2 output-OR wrapper is also atomic. All 96 native frame streams run
through exhaustive trajectory checks on both solver generations, and a fresh
5,000-case frame fuzz campaign checks both against the transition model. The
source manager start/stop functions only allocate the SOP table, cut arena and
temporary scratch; Go uses the validated immutable lazy SOP table and owned
per-call cut/truth state instead of retaining an allocation-only manager. The
complete suite passes in 10.539s, race in 60.190s, and vet/formatting are clean.

`sat/cnf/cnfUtil.c` is now validated across all nine active definitions.
Supplied-cut mapping scans now include the source's structural-supergate marks
with exact pre/post mapped-root order, dynamic references, area and bounded
iterative traversal; ordinary cuts remain separately owned for later emission.
The existing DAR-best-cut scanner covers the distinct native path. CI/CO SAT-ID
collection, packed literal polarities, checked strict/native-line DIMACS reading
and formula/text/file solving cover the remaining definitions. Solve options now
apply nonzero learned start/delta/ratio overrides and return copied full patterns
and deterministic formula/solver diagnostics instead of printing. Native result
codes, model suffixes, budgets and cancellation remain intact. Fresh mapping and
solver fuzz campaigns each passed 5,000 executions. The complete suite passes in
10.475s, race in 59.835s, and vet/formatting are clean.

`sat/bsat/satSolver.c` and `satSolver.h` are now validated, closing the legacy
BSAT solver surface through the same concrete core used by solver2 and solver3.
`SatSolverSession` adds the legacy-only variable-growth, projected-model,
run-ID cancellation, clause-diagnostic and proof-store-change operations.
Learned-database policy and Luby restart controls now map caller configuration
directly; zero restart interval implements `fNoRestarts`. Copied finite
nonnegative activity factors are applied on odd search starts, while the
experimental global-variable mask adds its factor-three bias on every search
start. Both use the live heap and checked saturating arithmetic, survive owned
bookmarks and are included in memory accounting. A fresh independent recurrence
fuzzer passed 5,000 factor/mask/search schedules.

The source's `TEST_CNF_LOAD` callback is inactive in the checked configuration
(and its enqueue block is commented out). `fSkipSimplify` needs no exposed mode:
the Go core always completes safe root propagation and never runs the source's
destructive clause-database simplifier. Typed owned slices replace raw arena
handles, and Go lifetime replaces delete-pointer helpers. The session otherwise
covers construction, clauses and learned clauses, propagation/search, models
and final conflicts, push/pop, resource limits, lexical solving, both assumption
minimizers, reduction, restart, bookmark/rollback, activity/polarity/random
controls, counts, memory, every inline gate family, DIMACS/statistics/model
utilities, trace and proof-store lifecycle. Tests are entirely Go-only. The
complete suite passes in 10.504s, race in 59.954s, vet and formatting are clean,
and the package-header/native-oracle policy guard passes.

`sat/bsat/satSolver3.c` and `satSolver3.h` are now validated.
The new concrete `SatSolver3` facade reuses the owned CDCL engine while adding
solver3's distinct pushed-assumption propagation/pop, relative and absolute
work budgets, restart-with-seed, activity/polarity controls, owned facade
bookmark, lexical SAT and deterministic irredundant assumption minimization.
All ten inline gate families are pure ordered-clause constructors with atomic
loading. Exact units cover failed-push atomicity, nested composition, limits,
lifecycle, ownership and every Boolean gate assignment. Lexical search is
checked against all 256 three-variable truth subsets at every lower bound; a
fresh 5,000-case Go-only arbitrary-CNF fuzzer also agrees with direct
enumeration. The specialized minimizer learns only sound entailed units,
correcting the source's complemented-unit insertion and dead loops. Checked
learned-clause insertion, live three-valued assignments, frequency clearing,
copied pivots, owned memory, DIMACS/statistics/model helpers and explicit trace/
proof-store delegation close the rest of the active surface. Fresh minimizer
and stateful push/pop fuzz campaigns passed 5,001 and 5,000 cases. The complete
suite passes in 10.696s, race in 59.280s, and vet is clean.

`sat/bsat/satSolver2.c` and `satSolver2.h` are now validated around a concrete
`SatSolver2` facade and the shared CDCL engine. The facade owns every original
caller clause—even tautologies and clauses already satisfied at root—with
stable one-based handles, exact external IDs, partition bits and copied literal
snapshots. It provides checked pre-solve metadata mutation, variable partitions,
exact original counts, model/conflict queries, controls, owned bookmark/
rollback and idempotent lifecycle. Contradictory single or batch insertion is
atomic. The shared engine supplies watched propagation, first-UIP learning,
Luby restarts, database reduction and assumption conflicts.

Concrete owned `SatSolver2ClauseSpec` builders preserve exact clause/literal
order and the repeated partition/ID metadata for constant, buffer, AND, XOR and
constraint encodings. An atomic loader composes their logical clauses into the
existing concrete CDCL solver while specifications remain available to later
proof/interpolation consumers. Exact units exhaust every Boolean assignment,
check copied ownership and error atomicity. The shared solver now retains owned
SAT models and UNSAT final conflicts and exposes variable/live-original/
learned/conflict counts, model literals, absolute runtime deadlines, learned-
database limits, activity reset and owned bookmark restoration. Formula
mutation invalidates stale results, including the zero-variable SAT case. The
facade closes the earlier exact call-count/handle gap. Root-satisfied watch
cleanup is also concrete and atomic: it removes both watch entries only for
clauses made permanently true at level zero, records the state in invariants
and bookmarks, and returns owned original/learned/entry counts. A fresh Go-only
fuzzer passed 5,000 cases with up to 128 independently stored clauses. Structured
owned memory accounting is now concrete: it uses actual Go type sizes and
backing capacities for solver and
facade structs, clauses/literals, optional watch entries, variable/trail/result
arrays, proof state and both bookmark layers. A fresh Go-only capacity fuzzer
passed 5,000 cases.

Solver2's optional `pPrf2`-style proof-information attachment is now concrete.
`EnableUnsatCoreTracking` freezes the loaded facade formula, maps logical
problem clauses back to stable caller handles and assigns dense proof IDs to
learned clauses. Learned and final conflicts collect the complete implication
dependency closure into the validated `SatProofInfoManager`; repeated solves
discard stale final records, bookmarks own proof snapshots, and learned-database
compaction remaps both records and clause proof IDs. `UnsatCore` returns stable
descending facade handles only after a fresh Go solver independently proves the
selected clauses UNSAT under the last assumptions. Exact tests cover learned
conflicts, assumptions, empty root cores, repeated solves, rollback, limits,
memory and compaction. A fresh Go-only exhaustive small-formula fuzzer passed
5,000 executions. The full suite passes in 10.497s, race in 61.345s, and vet is
clean.

Solver2's packed `pPrf1` behavior is now connected to the validated concrete
`SatProofDAG`. `EnableResolutionProofTracking` freezes and copies exact facade
roots, records variable partition bits and reconstructs every learned or final
clause as an exact resolution chain from the live implication graph. Each step
uses the standalone Go resolver; a node is committed only if the chain reaches
the precise learned clause or final assumption clause. Full replay starts from
the unsimplified caller clauses, so root-shortened CDCL clauses remain valid
proof roots. `VerifyResolutionProof` replays the reachable DAG, while
`ResolutionUnsatCore` extracts stable handles and independently re-solves them.
Repeated solves, assumption-only conflicts, dual pPrf1/pPrf2 attachment,
bookmarks, limits, proof memory and atomic DAG reduction/remapping with learned
database compaction are covered. Exact tests include root simplification and a
forced-reduction pigeonhole proof; a fresh 5,000-case Go-only formula fuzzer
replayed every UNSAT proof and exhaustively checked every core. The full suite
passes in 10.510s, race in 59.548s, and vet is clean. Packed proof checking and
core extraction are no longer gaps.

Solver-attached proof interpolation and the final diagnostics are now concrete.
The reachable exact DAG is replayed into an owned `SatProofStore` with used A
roots first, used B roots second and every verified learned resolvent in
topological order. `ProofAigInterpolant` and `ProofTruthInterpolant` then reuse
the validated concrete interpolators and reject assumption proofs because
assumptions are not partition roots. Exact tests exhaust Craig conditions and
AIG/truth agreement; a fresh 5,000-case fuzzer varies variable IDs, literal
polarity and explicit shared-variable order. Checked `Simplify` and
`VerifyModel` replace the remaining Boolean/printing diagnostics, while copied
facade clause lookup already replaces raw handle inspection. The final full
suite passes in 10.429s, race in 59.416s, and vet is clean. Every active function
and public declaration in `satSolver2.c`/`.h` now has a concrete, tested Go path,
so both files are promoted to validated.

Solver2 learned reduction now follows the source policy rather than the earlier
generic heuristic. Each reduction advances the configured limit, scores
learned clauses by capped LBD and shifted activity, preserves cutoff ties, the
recent suffix, units and locked reasons, and skips deletion unless at least one
sixth is eligible. Successful reduction compacts the arena and remaps every
reason, unit and watch handle while preserving watch-bucket order. Exact tests
pin selection, locked-clause remapping and limit progression; an independent
Go-only fuzzer passed 5,015 cases across ratios, ties and LBD/activity values.
The latest full suite passes in 10.290s, race in 58.855s, and vet is clean.

`sat/bsat/satSolver2i.c` is now validated. Its incremental standard
interpolation rules are absorbed by the concrete `SatAigInterpolator`: copied
global-variable order, A-root disjunctions, B-root one, learned proof reuse,
A-local OR, shared/B AND, cleanup, owned result transfer and lifecycle all
already have exhaustive and fuzz coverage. The file's self-relation diagnostic
is now `GiaManager.InterpolateOutputRelation`. It accepts a combinational
one-output graph and produces an owned, structurally hashed graph over the
original inputs plus a claimed-output input; its sole output is true exactly
when the claim equals the source output. Only reachable logic is copied, with
constants, buffers, ANDs, XORs and muxes supported. Tests cover source
ownership, dangling logic, constants, invalid graph shapes, exact limits and
cancellation. A fresh independent Go-only graph fuzz campaign passed 5,000
executions with exhaustive truth checking. The full suite passes in 10.217s,
race in 58.872s, and vet is clean.

`sat/bsat/satProof.c` is now validated for its complete enabled proof-DAG
surface. Concrete `SatProofDAG` state replaces packed word offsets with typed
node/clause references and dense handles while retaining the partition-pivot
bit, antecedent order, record sizing, label cleanup, reachability, used counts,
core extraction and proof compaction semantics. Traversal is iterative and
deterministic; compaction atomically remaps antecedents, roots and the proof
pivot. Resolution verification composes the existing standalone Go checker.
Tests cover copied ownership, exact sizing, shrinking, invalid states, limits,
cancellation, lifecycle, a 10,000-node chain and resolution failures. A fresh
Go-only independent-model fuzz campaign passed 5,000 executions and checked
that core results survive reduction. The source-commented reduction variants
and blocks under `#if 0` are not live runtime entry points; their useful checker
and interpolation behavior already has concrete validated Go implementations.
The full suite passes in 10.328s, race in 58.537s, and vet is clean. No native
source, compiler, runtime oracle or generated C fixture is retained under
`pkg/`.

`sat/bsat/satUtil.c` is now validated across all ten utility procedures.
Concrete helpers serialize owned solver snapshots in exact one-based DIMACS or
native raw-ID form, including original clauses, root assertions and ordered
assumptions; checked file output replaces stdout/silent failure. Both legacy
solver generations converge on this one Go representation. Exact formatters
cover both statistics layouts and packed counters, native solver2 variable
memory is modeled explicitly, and model projection validates copied SAT
results. The source-disabled clause-doubling routine returns `ErrUnsupported`
without mutation. Go-only round-trip fuzzing passed 5,000 executions by parsing
each snapshot into a fresh solver and comparing satisfiability. The full suite
passes in 10.277s, race in 58.280s, and vet is clean.

`sat/bsat/satInterA.c` and `satInterB.c` are now validated, completing every
declaration family in `satStore.h` and promoting that header to validated.
Concrete `SatAigInterpolator` standard mode preserves A-local OR and shared/B
AND composition; mux mode preserves A-local OR, B-local AND and the exact
polarity-directed Mux0/Mux1 orientation for shared pivots. Both infer or accept
copied shared-variable order, reject omitted common variables, reuse owned
proof replay and return checked one-output `AigManager` graphs. Partition-CNF
derivation preserves the source's reserved gap before local SAT-variable
inputs. Tests cover learned replay, A/B/shared pivots, constants, ownership,
graph invariants, exact input mapping, limits, cancellation and lifecycle.
Fresh Go-only fuzzing passed 5,000 executions across both modes with exhaustive
Craig checks and semantic comparison. The full suite passes in 10.200s, race
in 58.160s, and vet is clean.

`sat/bsat/satInter.c` is now validated across truth-table interpolation and
its shared proof-replay path. `SatTruthInterpolator` infers shared variables
in variable order or accepts an explicit copied order, initializes A roots as
disjunctions of shared literals and B roots as one, then combines proof reasons
with OR for A-local pivots and AND for shared/B pivots. Canonical `uint64` truth
words replace raw `unsigned*` results while preserving assignment-bit order;
the source boundary of eight shared variables is tested as 256 bits across four
words. Exact tests cover constants, learned replay, ownership, formatting,
limits, cancellation and lifecycle. Fresh Go-only fuzzing passed 5,000 cases,
exhaustively checking `A => I` and unsatisfiability of `B && I` under variable
renaming, polarity, clause order and explicit global order. Only the two AIG
interpolators remain in `satStore.h`.
The uncached full suite passes in 9.945s, race in 58.176s, and vet is clean.

`sat/bsat/satInterP.c` is now validated across its complete proof-replay and
UNSAT-core path. Concrete `SatProofCoreManager` state preserves watched-clause
propagation, root assignments, learned-clause assumption replay, proof-number
and antecedent construction, resolution checking and stronger-resolvent
shrinking. It extracts either root or learned cores with iterative deterministic
postorder traversal, verifies root cores with the Go SAT solver and owns exact
debug/BMC formatting. Limits, cancellation and errors replace assertions and
unchecked allocation. Two Go-only fuzz campaigns passed 5,000 executions each,
covering contradictory-root families and randomized learned-unit proofs with
variable renaming, polarity changes, root reordering and independent SAT
verification. `satStore.h` remains partial for only the three truth/AIG
interpolation managers.
The uncached full suite passes in 10.104s, race in 58.680s, and vet is clean.

`sat/bsat/satStore.c` is now validated across all twelve functions, and the
store portion of `satStore.h` is partial only because its four interpolation-
manager declaration families remain unported. `SatProofStore` preserves sorted
encoded literals, same-variable rejection, stable clause IDs/order, variable
maxima, one empty clause, root/A prefixes, last-unit partition transfer,
64-KiB chunk accounting and the four-field text format. Its loader corrects
the source's unusable declared-count insertion bug and restores root/A flags
from the header. Exact/file/model tests cover ownership, malformed formats,
limits and chunk rollover; fresh Go-only round-trip fuzzing passed 5,000
executions without a native oracle or retained fixture.
The uncached full suite passes in 10.097s, race in 58.498s, and vet is clean.

`sat/bsat/satVec.h` is now validated across both legacy vectors, every literal
helper, three-valued Boolean constants and the statistics record. Concrete
`SatIntVector` and `SatPointerVector` preserve capacity four and exact 3/2
growth, ordered push/shrink/pop/removal and indexed mutation; `uintptr` replaces
`void*`, copied views replace backing-pointer exposure, and errors/limits replace
assertions. Existing typed SAT literals now also provide exact signed one-based
DIMACS read/write and variable-bound checks. Boundary/ownership tests pin the
4,6,9,12,18 capacity sequence and literal round trips; an independent randomized
vector model passed 5,013 executions without native code or a fixture.
The uncached full suite passes in 10.381s, race in 58.460s, and vet is clean.

`sat/bsat/satTrace.c` is now validated across all three streaming trace
functions. `SatSolverTrace` creates/truncates a file with the exact 40-space
placeholder, writes ordered signed one-based DIMACS clauses and root counts,
then rewrites the padded `p <vars> <clauses> <roots>` header and closes it.
`StopForSolver` obtains the variable count from the concrete solver. Checked
stdio errors, atomic invalid-stop recovery and clause/literal limits replace
assertions. Exact-byte temporary-file tests cover empty clauses, roots,
lifecycle and errors; fresh standalone formatter fuzzing passed 1,000
executions without a native oracle or retained fixture.
The uncached full suite passes in 9.581s, race in 58.435s, and vet is clean.

`sat/bsat/satProof2.h` is now validated across its complete proof-information
manager. `SatProofInfoManager` owns flat fixed-width root-bit records and
preserves exact word rounding, grow-at-capacity doubling, record growth/shrink,
saved-clause compaction/remapping, learned-record OR chains, direct/mapped
problem roots and descending UNSAT cores. Copied records and mappings replace
raw pointers; explicit `SetProofID` updates replace native mutation through an
aliased caller vector. Errors, limits, capacity statistics and idempotent stop
replace assertions and frees. Unit tests cover direct/mapped chains, mapping
ownership, exact growth, shrink, compaction and lifecycle; a fresh independent
bitset/compaction fuzz model passed 5,000 executions without a C oracle.
The uncached full suite passes in 10.202s, race in 58.837s, and vet is clean.

`sat/bsat/satChecker.c` is now validated across all three functions as a
concrete owned signed-DIMACS resolution checker. Clause formatting returns a
string; resolution preserves exact complementary-pair counting, pivot removal
and first-occurrence antecedent order; proof checking preserves sequential IDs,
root double-zero records, resolution single-zero records, backward references
and final-empty reporting. Errors and byte/clause/literal/line limits replace
assertions and silent file-open failure. Constructed proof/model tests cover
valid derivations, nonempty conclusions, malformed pivots/references/trailers
and file behavior. Fresh Go-only resolver and arbitrary-text fuzzing passed
5,000 and 5,018 executions respectively, with no retained fuzz artifact. The
uncached full suite passes in 10.195s, race in 57.586s, and vet is clean.

`sat/bsat/satClause.h` is now validated across its complete declaration,
inline and iterator surface. `SatClauseStore` preserves native page/offset
handles, tagged binary literals, exact storage rounding and rollover, ordered
problem/learned traversal, metadata/hidden IDs, bookmarks, rollback, restart,
shrink-without-count-repair, memory accounting and two-pass learned compaction.
Owned snapshots replace raw clause pointers, checked errors/limits replace
assertions, and formatting replaces stdout. A lookup-only handle map is never
iterated; ordered slices determine every result. Boundary, multi-page,
bookmark-pivot and independent relocation-model tests are Go-only, and a
fresh 5,000-execution state-machine fuzz run passes. The uncached full suite
passes in 10.144s, race in 57.962s, and vet is clean.

`sat/bsat/satMem.c` and `satMem.h` are now validated as three concrete owned
allocator types. Fixed arenas preserve the native 1,024-entry/64-KiB chunk
rule, LIFO reuse, peaks and first-chunk restart; flexible arenas preserve
64-KiB bump chunks and oversized growth; step arenas preserve power-of-two
classes, retained oversized blocks, restart and the source's class-only memory
usage. Checked blocks make foreign, stale and double recycling explicit, and
resource limits replace allocation failure. Boundary/unit tests and two
stand-alone Go state-machine fuzz targets use neither C code nor a generated
native oracle. Fresh fixed-arena fuzzing passed 84,103 executions and fresh
flex/step fuzzing passed 3,861 executions after its expanded cached corpus.
The combined uncached full suite passes in 10.213s, race in 57.919s, and vet is
clean.

`sat/bsat/satTruth.c` and `satTruth.h` are now validated as one concrete owned
truth store. Exact variable tables through sixteen inputs, zero/one and signed
complement handles, normalized insertion, stable deduplication, copied reads,
statistics, lifecycle and every inline word operation are covered. Dense even
handles replace page offsets without changing XOR-one polarity; a lookup-only
map is never iterated, so ordered slices determine IDs. Unit/model tests cover
all widths, constants, aliasing, bounds and errors, and fresh standalone
fuzzing passed 1,843,374 executions. The combined uncached full suite passes in
10.304s, race in 57.937s, and vet is clean.

`bool/dec/decMan.c` is now validated across both lifecycle functions.
`DecManager` owns concrete MVC state, private factoring scratch and independent
copies of exact four-input canonical, phase, native-permutation and class
tables. Exhaustive Go-only checking covers all 65,536 functions against the
independent NPN enumerator and reconstructs every function from its recorded
transform, including the source's sparse complemented-representative class-map
quirk. Ownership/close tests pass, and fresh orbit fuzzing passed 1,947,272
executions. The combined uncached full suite passes in 10.361s, race in
58.659s, and vet is clean.

`opt/sfm/sfmMit.c` is now validated across all thirteen active default-build
placeholder functions. Concrete Go spellings preserve its unconditional nil
constructor, zero queries/decisions and no-op lifecycle/load/timing mutations;
opaque handles use `uintptr`, while vectors use concrete slices and no
interfaces. Unit tests prove exact caller-storage preservation across nil,
nonzero, duplicate and signed inputs. Fresh alias-heavy standalone fuzzing
passed 1,940,640 executions. The combined uncached full suite passes in
10.385s, race in 58.958s, and vet is clean.

`aig/gia/giaDfs.c` is now validated across all twenty-four definitions.
Owned iterative paths cover forward/reverse grouping, sequential closure,
TFI/TFO, CI support, leaf-bounded/all-gate/combined postorder and union
support/cone counts, including mux controls and extended nodes without manager
scratch mutation. The three timing/stdout-only drivers are superseded by an
explicit Go regression performing their whole-output DFS, small-support count
and per-output sequential traversal. Existing native fixtures pin all ordered
algorithmic results; fresh independent cone-model fuzzing passed 896,179
executions. The combined uncached full suite passes in 10.345s, race in
58.123s, and vet is clean.

`aig/aig/aigObj.c` is now validated across all thirteen active definitions.
The final three mutation paths use an owned topological rebuild in place of
native transient forward-pointing buffers: replacement redirects every old
use, buffer repair/propagation bypasses all chains, structural hashing rebuilds
ANDs, terminal and sequential partitions are retained, and an explicit
old-to-new literal map records every changed ID. Replacement cycles, buffers,
choice graphs, malformed IDs, limits and cancellation fail without changing
the source. Exhaustive phase tests cover higher-ID replacements and unrelated
buffers, a 10,000-node case is iterative, and fresh standalone formula-model
fuzzing passed 2,419,984 executions. The uncached full suite passes in 10.466s,
race in 58.215s, and vet is clean.

`misc/vec/vecMem.h` is now validated across its complete macro/function
surface. `WordMemory` owns bounded fixed-width paged records, deterministic
growth/shrink, cloning, snapshots, interning and truth-table seeds. The final
observable gaps now have an exact native uint32/half-word hash-key operation
and an owned truth-table payload/filename/summary dump. Go lifetime replaces
free helpers, ordinary loops replace the iterator, source-empty `Fill/Clean`
remain explicit no-ops, and map-internal profile/resize replace C bucket
diagnostics without affecting stable IDs. Unit tests cover formatting,
ownership, bounds and hash arithmetic; fresh standalone hash fuzzing passed
2,036,171 executions. The uncached full suite passes in 10.263s, race in
59.059s, and vet is clean.

`aig/saig/saigScl.c` is now validated across both diagnostics. An owned report
replaces CIO scratch and stdout for single-reference LO-to-PO and PI-to-LI
relationships. A second query preserves the source's actual complement-report
behavior: every RI after the first with the same regular driver increments the
count, irrespective of edge phase. Local seen storage replaces `fMarkA`
mutation, with checked zero-register, resource and cancellation behavior.
Exact relationship/duplicate tests and an independent terminal-selection
fuzzer cover the unit; fresh fuzzing passed 1,853,877 executions. The uncached
full suite passes in 10.189s, race in 58.627s, and vet is clean.

`aig/gia/giaPat.c` is now validated across its three external functions and
two inline mark accessors. Owned iterative postorder/value slices replace
object marks; partial assignments use CI ordinals, retain last-duplicate-wins
behavior and leave omitted inputs unknown. Verification returns the ternary
value and definite-assertion result instead of printing failure. Buffers, XORs
and muxes receive explicit semantic handling, including mux-control traversal.
Exhaustive partial-assignment tests, cone-order/duplicate/error/resource checks
and an independent Boolean-completion fuzzer cover the unit; fresh fuzzing
passed 1,861,405 executions. The uncached full suite passes in 10.070s, race in
58.598s, and vet is clean.

`aig/gia/giaEnable.c` is now validated across all ten active definitions.
Owned APIs cover immediate/recursive supergate collection, exact signal
frequency reports, sequential candidate detection, frame transfer, initialized
unroll/cofactor and both enable-removal algorithms. The newer recognizer returns
normalized update controls and data branches per register; the legacy
OR-of-AND recognizer returns one-based control classes explicitly rather than
attaching hidden manager metadata. Nonmatches retain their RI drivers and are
reported without stdout. Exhaustive phase/branch tests, legacy class tests,
ownership/resource/cancellation checks and a standalone Boolean-model fuzzer
cover the transforms; fresh fuzzing passed 2,311,228 executions. The uncached
full suite passes in 10.455s, race in 59.326s, and vet is clean.

`aig/gia/giaFanout.c` is now validated across all ten external functions and
six inline helpers. Owned immutable snapshots cover ordinary and LUT-mapped
fanouts; `GiaDynamicFanouts` covers native two-slot tail insertion, exact removal
and iteration without manager scratch. Mapped order retains duplicate fanins
and appends CO edges after LUT roots. Explicit slot validation removes native
fanin-inference ambiguity, while Go lifetime replaces stop/free operations and
tests replace stdout diagnostics. Fresh standalone model fuzzing passed
1,813,845 cases.

`misc/util/utilSort.c` is now validated across all 24 active definitions.
Concrete slice operations cover integer merge sort, signed/unsigned cost
permutations, low-32-bit word sorting and quick-cost ordering; owned `ABCRandom`
replaces process-global state. Native tie/base-case behavior, unsigned cost
reinterpretation, RNG/reset sequences and the QuickSort3-to-two-way wrapper
quirk are retained. Safe empty handling and direct comparisons correct source
out-of-bounds and subtraction-overflow hazards. Fresh standalone model fuzzing
passed 516,998 cases; benchmark-only drivers are superseded by tests.

`sat/cnf/cnfWrite.c` is now validated across all nine active definitions. The
new owned simple encoders preserve native variable and clause order, asserted
versus exposed outputs, the source's all-CO allocation quirk when exposure is
nonzero, and backward-retiming output-zero units. Exact hand-derived streams,
resource/error tests and solver-backed semantics cover contradictory assertions
and every exposure mode. Fresh standalone fuzzing passed 1,404,802 cases.
Existing mapped, wide-cut, object-ID and mapping-array writers retain their
native fixture and exhaustive SAT-projection coverage.

`sat/cnf/cnfFast.c` is now validated across all twelve definitions. Owned cone
DFS, truth derivation, marking, diagnostic counting and complete fast CNF retain
native traversal, mux/support, variable and clause order without graph scratch
mutation. Native fixtures and exhaustive SAT projection cover direct and
supergate paths; checked errors replace source assertions outside the six-leaf
complex-cut/object model. Fresh standalone fuzzing passed 1,212,559 cone cases,
1,073,059 marking cases and 504,660 complete derivations.

`sat/cnf/cnfData.c` is now validated. Its sole active reader synchronously
decodes the embedded four-input SOP payload into immutable native-ordered data
and returns owned covers. Permanent data matches all 65,536 rows and independent
truth evaluation proves each cover; fresh standalone fuzzing passed 1,994,033
cases. The two `#if 0` table-development/debug programs are superseded by these
integrity and semantic tests, including their stdout/external-file workflow.

`misc/vec/vecFlt.h` is now validated. Owned bounded `FloatVector` accounts for
all 42 definitions; copied constructors and checked access replace adopting and
borrowed pointers, Go lifetime replaces frees, and `Format` replaces stdout.
IEEE payloads, signed zero, infinities, NaN rejection, stable numeric ordering,
limits, ownership and exact finite formatting are covered. Fresh standalone
fuzzing passed 794,222 stateful sequences and 1,983,556 raw IEEE cases.

`misc/vec/vecBit.h` is now validated. `BitVector` accounts for all 27 inline
functions with owned bounded storage; garbage collection replaces free helpers,
owned snapshots/transfer replace mutable array pointers, and ordinary Go loops
replace the five iteration macros. Boundary behavior, allocation limits,
ownership and intentional fixes for C's boundary shifts, unaligned clone and
memory overcount are pinned by unit/model tests. Fresh standalone operation
fuzzing passed 208,121 sequences without C.

`bool/kit/kitIsop.c` is now validated. One owned recursion engine covers exact
and on/don't-care interval ISOP through sixteen variables, including native
direct/complement selection and deterministic PLA-style formatting. The
returned table is derived from the selected cubes, intentionally removing the
source's ambiguous last-attempt table when both polarities are tried. Permanent
equal-set rows retain exact native cube order; exhaustive unit properties cover
interval containment, complements, ownership, constants, malformed inputs,
limits and formatting. Fresh Go-only interval fuzzing passed 1,811,974
executions without invoking or generating C.

`sat/cnf/cnfCut.c` is now validated. Owned cut construction and Go lifetime
replace flex-arena allocation/free; exact formatting replaces stdout; pure
leaf merging/alignment avoids temporary source mutation; postprocessing owns
reference decrement/increment/update state. Correct substitution intentionally
fixes the source truth-pointer, support-movement and packed-cofactor defects,
with reviewed original outputs retained only as regression evidence. Focused
lifecycle/postprocess tests pass and fresh Go-only exhaustive substitution
fuzzing passed 1,539,092 executions. The resulting full suite passes in
10.247s, race in 58.397s, and vet is clean.

`misc/vec/vecQue.h` is now validated. `FloatPriorityQueue` covers all heap,
growth, priority, update, membership and validation operations with native
float32 ordering/ties and stronger sparse-ID invariants. Owned priority copies
replace caller-borrowed mutable pointers, Go lifetime replaces both free
helpers, and `Format` returns the exact level-by-level print layout. The demo is
superseded by ordinary/model tests; fresh Go-only queue fuzzing passed 193,094
executions. The resulting full suite passes in 10.093s, race in 58.763s, and
vet is clean.

`sat/cnf/cnfMap.c` is now validated. Active topological area-flow selection
preserves SOP scaling, integer division, unsigned accumulation, eleven-bit
reference sums and both tie rules. The compiled but uncalled supergate helper
now has a checked literal-regularizing Go API; its arithmetic-model fuzzer
passed 1,934,279 executions. All seven exact-area functions are in one `#if 0`
block, and their only apparent iterative caller is in the separately disabled
old core derivation. Automatic Go CNF construction already composes validated
DAR generation, selection, scan, postprocessing and writing stages. The
resulting full suite passes in 10.131s, race in 58.554s, and vet is clean.

`aig/gia/giaMffc.c` is now validated. Owned counters and dominator slices cover
all private MFFC/dominator helpers without manager scratch mutation; bounded
cone derivation covers the zero-through-five-variable cofactor construction.
The final diagnostic driver now returns source-ordered candidate/root IDs,
strict-selection-sorted boundary/cost pairs and owned `k=0` derived cones
instead of printing, writing fixed-name files and calling an external collapse
experiment. Exhaustive cut replay independently checks every derived graph;
fresh Go-only candidate fuzzing passed 1,648,524 executions. The resulting full
suite passes in 10.290s, race in 58.974s, and vet is clean.

`bool/lucky/luckySimple.c` is now validated. All fourteen active definitions
map to owned Johnson-Trotter/Gray schedules, exact unrestricted/grouped
minimizers, Go lifetime or shared truth-table primitives. Exact hierarchical
canonicalization already consumes the grouped minimizer; unrelated Lucky files
remain separately tracked. The one-input dummy swap is intentionally corrected
to preserve padded truth semantics. Permanent datasets cover 40 unrestricted
and 320 grouped cases, traversal properties prove schedule completeness, and a
fresh independent brute-force orbit fuzz campaign passed 1,035,339 executions.

`aig/gia/giaScl.c` now also covers `Gia_ManCleanupOutputs`. The owned Go API
removes a checked combinational-output prefix while retaining logic reached
only by skipped outputs, because the source marks through every output before
omitting the prefix. It preserves dangling buffer roots, source mappings,
metadata and constraint suffixes; invalid sequential or constraint-splitting
requests return errors. Unit properties exhaust assignments and distinguish
skipped-only cones from unreachable logic; fresh Go-only fuzzing passed
1,722,360 executions. The file remains partial only because native
`pNexts` equivalence and `pSibls` choice reachability metadata is not yet owned
by `GiaManager`. The resulting full suite passes in 10.296s, race in 57.764s,
and vet is clean.

`sat/bmc/bmcMulti.c` has moved from unstarted to partial across all seven active
definitions. Exact owned output processing transfers replayable witnesses into
original output slots, clears transferred input entries and returns both local
leftovers and the compacted original-output map. GIA and AIG constant-zero
counters are exact, and structured stage reports replace stdout and elapsed
time. The concrete `AigerModel.MultiProve` loop performs deterministic
solve-all sequential simulation, per-output bounded checking, checked output
compaction, optional owned synthesis and increasing local frame bounds with
iteration/gap/resource/cancellation stops. It publishes only full
original-coordinate counterexamples independently replayed on the unchanged
original model; this intentionally strengthens the source's solve-all rarity
pointer-one markers and reduced-cone trace coordinates. `GiaManager.MultiProve`
provides the source zero-initialized entry point without adding an interface.

The file is not yet validated: preliminary simulation does not yet implement
`Ssw_RarSimulate`'s rarity bins, pattern ancestry and restart transfer, while
the synthesis boundary currently performs semantics-preserving GIA cleanup and
structural rehash rather than `Gia_ManAigSyn2` area balancing and JF mapping.
Unit tests cover transfer ownership/atomic errors, exact constants/reports,
sequential synthesis equivalence, constraints, original-coordinate replay,
source immutability, gap/iteration/resource/cancellation behavior and the GIA
entry point. Two stand-alone Go fuzz campaigns passed 5,000 executions each.
The full suite passes in 15.703s, the race suite in 81.379s, and vet is clean.

`sat/bmc/bmcEco.c` is now validated across all five active definitions.
`BuildEcoMiter` owns explicit golden/old mappings and preserves the source's
shared CIs, extra replacement PI, last-PO root substitution, onset per-output
XORs, offset mismatch OR, offset-valued divisor outputs, structural hashing,
cleanup and sequential-terminal flattening without borrowing `Value` scratch.
It rejects buffers/native XORs/muxes rather than silently treating every
`Gia_ObjIsAnd` object as an ordinary AND. The remapped-CNF helper uses the
concrete GIA-to-AIG and reference Tseitin path. `SolveEcoCover` preserves the
selector-guarded onset query, divisor-model offset query, UNSAT assumption-core
cube extraction and learned blocking clauses with the source million-conflict
default. `ComputeEcoPatch` adds exact equality/inequality output units and
derives checked root/divisor literals. Owned status, stop reason, cubes and SAT
statistics replace stdout and the hardcoded file driver.

Exhaustive tests synthesize and replay all 16 two-input functions and an
impossible one-divisor projection. Further tests exhaust the miter truth table,
pin sequential flattening, verify source ownership, every graph/CNF/iteration/
cube limit, cancellation and a 10,000-gate cone. A fresh stand-alone Go fuzz
campaign passed 5,000 executions. The full suite passes in 15.698s, the race
suite in 81.802s, and vet is clean.

`sat/bmc/bmcChain.c` is now validated across all ten active definitions.
`ChainFailOneOutput` connects the concrete GIA-to-AIG path to replay-checked
bounded checking. `MoveChainCounterexample` verifies the selected failure,
captures register values at the start of the failing frame and uses exact
RO/RI phase recoding so that reached state becomes zero, without borrowing
marks or `Value` scratch. Initial-state propagation substitutes zero register
outputs, retains only PIs and POs, hashes native extended gates by their actual
semantics and cleans unreachable logic. The incremental concrete CNF/CDCL path
tests every nonconstant output and returns independently replayed one-frame
witnesses. Output removal is atomic and leaves the source graph unchanged.

The outer chain retains the BMC/move/SAT/cleanup order, failing-frame depth
accounting and 10,000-iteration default. Its returned segments explicitly keep
each path in its current phase-recoded coordinates instead of reproducing the
source's ambiguous flat pointer list. Structured stop reasons, statistics and
an owned final graph replace stdout and elapsed time; UNKNOWN never becomes a
failure. Environment constraints are rejected because the native second SAT
phase treats every PO as a property and therefore cannot preserve constraint
semantics. Tests cover exact start-of-frame movement, exhaustive current-state
output satisfiability, witness replay, source ownership, atomic cleanup, a
two-state/two-segment chain, finite and iteration bounds, resource limits,
cancellation and malformed calls. Go-only fuzzing compared 5,039 executions
against exhaustive truth-table enumeration. The final full suite passes in
15.770s, the race suite in 81.941s, and vet is clean.

`sat/bmc/bmcCexTools.c` is now validated across all 21 externally meaningful
definitions. The Go toolbox expands a concrete counterexample into frame-major
PI/RO states and implication masks, derives all-controlling or deterministic
single-controlling structural care across register boundaries, checks care by
ternary replay, computes exact single-bit essentials, minimizes around those
essentials and verifies both published masks. Exact perturbation replaces the
source's cached structural essential-bit approximation. Native buffers, XORs
and muxes have explicit semantic handling in analysis; the specialized
ordinary-AND monotone justification-network builder rejects them instead of
silently misencoding them.

Initialized property unrolling reuses the owned timeframe machinery. Structured
counterexample/AOG statistics and deterministic formatting replace stdout and
append-only files. Reverse care-node statistics return owned per-pattern data,
and the MNIST helper now decodes caller-provided checked standard-format bytes
instead of opening a hardcoded working-directory file. No graph marks or
`Value` scratch are borrowed. Tests cover sequential states and implications,
all ordinary-AND justification cases, native XOR/mux care, insufficient masks,
ownership, diagnostics, malformed image data, limits, cancellation and a
10,000-node iterative cone. Every satisfying assignment of all 255 nonzero
two-input truth tables passes care/minimized replay and exact essential checks;
an additional Go-only fuzz campaign passed 5,000 executions. The final full
suite passes in 15.779s, the race suite in 81.611s, and vet is clean.

`sat/bmc/bmcCexMin2.c` is now validated across all three active definitions
and its packed annotation helpers. `AnnotateCounterexampleUnrolling` owns exact
per-frame object values and backward justification membership, preserving both
the deterministic first-controlling-zero and maximum-controlling-zero policies
without graph marks or `vTruths` scratch. An annotation retains an exact private
structural snapshot, so a mutated same-sized graph cannot consume stale object
indexes. Nonzero initial register states are supported.

`CreateCounterexampleUnate` preserves both source modes: selected start-frame
RO candidates for counterexample analysis and pseudo-PIs after a real-input
prefix for refinement. It retains use-all candidate CIs, fixed/path constants,
frame-to-frame RI transfer and the observed-value AND/OR justification
recurrences. Extended nodes are rejected rather than treated as ordinary ANDs.
The C `Gia_ManCexMin` routine only printed these graphs, optionally overwrote
`temp.aig`, and returned `NULL`; `CounterexampleUnateNetworks` instead returns
the annotation, exact frame indexes and owned networks without filesystem side
effects. Tests cover both modes, all recurrence choices, stale annotations,
ownership, limits, cancellation and a 10,000-node cone. A 5,000-execution
Go-only fuzzer exhaustively checked every completion admitted by each accepted
fixedness set. The final full suite passes in 15.855s, the race suite in
82.361s, and vet is clean.

`sat/bmc/bmcCexMin1.c` is now validated across all 13 active definitions.
Iterative frame-term collection preserves selected-final-PO roots,
next-frame-RO to prior-RI transfer and fanin-zero-first DFS order. Owned
per-frame phase/priority tables reproduce true-AND minimum priority,
single-controlling-zero selection and double-zero maximum priority. Both source
priority-band variants are explicit, including the diagnostic policy that
places registers above all PIs. Nonzero initial state values extend the
zero-reset-only source behavior, and no traversal ID or `iData` scratch is
borrowed.

Reason collection returns ordered frame-local PI or register literals.
Structural provenance prevents reuse after graph mutation. Cube duplication
copies the AIG, emits each post-frame-zero register cube followed by the
original RIs, preserves name/register count, supports buffers and cleans
dangling logic. `MinimizeAigCounterexampleReasons` returns this analysis and
owned graph instead of printing statistics, overwriting `aigcube.aig`, and
returning `NULL`; it does not overclaim a minimized trace. Tests pin exact
sequential terms/priorities/reasons, all controlling-zero cases, both priority
policies, positive and negative cubes, initial states, buffers, provenance,
limits, cancellation and a 10,000-node cone. A 5,000-execution Go-only fuzzer
exhaustively proved that every selected reason forces the original failure.
The final full suite passes in 15.623s, the race suite in 82.022s, and vet is
clean.

All retained tests are now stand-alone Go tests or immutable-dataset tests.
The former C capture programs, native smoke driver and compatibility header
have been removed from `pkg/tools`; no C or C++ source or header remains anywhere
under `pkg`. The package guard rejects `.c`, `.cc`, `.cpp`, `.cxx`, `.h`, `.hh`,
`.hpp` and `.hxx` files, as well as retained `testdata` directories. Tests do
not invoke ABC, a C compiler, cgo or a native oracle. The only remaining tool is
the planning-only source inventory script, which is not read or run by tests.
An isolated copy containing only `pkg/`, with no repository parent or `abc/src`
tree, passes the uncached complete suite in 10.399s and passes vet. The in-tree
uncached complete suite passes in 10.443s, and vet is clean.

The stand-alone boundary was re-audited against the current tree on 2026-10-28.
An explicit whole-module scan found no C/C++ source, header, native object,
archive, shared library, executable or permanent `testdata/` directory. The
only file in `pkg/tools/` remains the non-runtime source inventory script.
Focused immutable-dataset and stand-alone policy tests pass in 0.071s, the full
Go suite passes in 16.225s, the race suite passes in 84.393s, and vet is clean.

`sat/cnf/cnfCore.c` is now validated across its three active derivation forms:
default optimized CNF, packed mapping arrays and object-ID/optional
truth-minimization CNF. Independent owned per-call state replaces the native
global manager singleton and its prepare/read/free API; owned cuts replace
arena transfer/cleanup. The only old derivation is wholly `#if 0`. Native clause,
numbering, range and mapping fixtures plus exhaustive SAT projection cover all
paths; fresh end-to-end Go-only fuzzing passed 558,554 executions.
The resulting full suite passes in 10.308s and vet passes; the unchanged Go
implementation most recently passed race in 58.162s.

`sat/cnf/cnfPost.c` is now validated. The active one-pass postprocessor retains
native root order, stable cost ordering, single-reference absorption, exact
reference updates and nonincreasing cost while using the already validated
corrected Boolean composer. Owned cut slices replace the transfer/free arena
lifecycle. The `_old` routine is a dead diagnostic scaffold with null best-cut
pointers and commented-out producers, so it has no runtime pass to expose.
Forty-eight mappings plus exhaustive SAT projection cover integration; fresh
Go-only semantic fuzzing passed 553,467 executions.

`aig/aig/aigCheck.c` is now validated. `AigManager.Check` covers every C
terminal, internal-fanin, canonical-order, structural-hash and live-count rule,
and additionally verifies topological order, CI/CO order, levels, phases,
references, sequential partitions, tombstones and choice chains. Errors replace
stdout/assertions. Phase checking is integrated into every invocation. The C
MarkA-reset assertion has no runtime counterpart because Go objects contain no
persistent traversal marks; algorithms use owned local slices. Targeted
corruptions exercise each invariant, while the Go-only AIG operations fuzzer
checks after every mutation, cleanup and clone; a fresh campaign passed 219,298
executions.

`misc/util/utilCex.c` and `utilCex.h` are now complete. Owned statistics record
total, register, real-input and auxiliary-input populations; deterministic
formatters replace all three stdout routines and define zero-denominator
behavior. A register-aware combinational-model constructor places model bits
after the zero initial state, correcting C's latent missing offset (its only
call supplies zero registers). Go lifetime replaces both free routines and the
pointer-one sentinel. The checked merge continues to reject nonzero overloaded
`part.iPo`, where C allocates and copies inconsistent frame counts before
failing its assertion. Layout/statistics/format fuzzing passed 1,192,461
Go-only executions, promoting both source and header to validated. The full
suite passes in 10.491s, race in 58.162s, and vet passes.

The build-disabled experimental `opt/dau/dauDsd2.c` is now fully dispositioned.
It is absent from `module.make` and cannot compile as C: invalid preprocessor
directives, undeclared locals, missing semicolons, a missing return, a mismatched
test variable and comment-only object mutations leave the one-/two-variable
manager sketches nonterminating and without executable behavior. No Go runtime
API claims those fragments. Its coherent general-range and low-bound
decomposition algorithms are implemented with exact inverse composition,
atomic limits/cancellation and independent reconstruction checks. Fresh Go-only
fuzzing passed 1,074,337 general-range and 709,741 low-bound executions,
promoting the source to validated without compiling or executing C. Every
tracked artifact under `src/opt/dau/` is now validated. The post-promotion full
suite passes in 10.287s; the unchanged Go code passed race in 58.491s and vet
remains clean.

`opt/dau/dauCanon.c` is now fully accounted for. The final active algorithmic
gap, its direct byte-scanning Slow cofactor population path, has an independent
Go implementation with packed small-table semantics, owned counts and safe
zero-through-two-input extensions. It agrees with the general packed counter
through twelve variables; Go-only fuzzing passed 1,244,318 executions. The
public small-word normalizer now has a checked direct Go surface. A complete
definition audit maps the commented duplicate comparators to `CompareCofactors`,
debug reconstruction/copy/configuration helpers to Go invariants and ownership
tests, and the diagnostic adjacent-swap driver to ordinary unit/fuzz checks.
`Abc_TgResetGroup1` has no caller anywhere in the C tree and intentionally has
no runtime API; the called general reset path is implemented. All 88 ledgered
functions, private helpers, state types and fixed tables have explicit Go
implementations or dispositions, promoting the source to validated without a
C runtime oracle. The resulting full suite passes in 10.362s, the race suite
in 58.491s, and vet passes; JSON, symbol-ledger, package-header and standalone
test-vector policy checks are clean.

The public `opt/dau/dau.h` surface now has checked Go bounds, exact DSD kind
numbering (including the reserved zero value), constant predicates and safe
variable-token decoding. Every declaration now maps to one of the seven
validated defining translation units: canonicalization/hierarchy, Quick
counting, DSD decomposition, GIA lowering, merge, non-DSD set analysis and the
structural manager. Concrete Go option/result methods replace the callback
typedef without an interface; owned managers and deterministic summaries
replace allocation/free and FILE/stdout declarations. The duplicate
hierarchical declaration is represented once. This promotes the header to
validated. The complete default
`opt/dau/module.make` list is
now dispositioned: all twelve selected translation units have explicit ledger
entries, while excluded DAU files remain separately tracked. The full suite
passed in 10.406s, the race suite in 58.444s and vet passed.

`opt/dau/dauTree.c` now has its first native Go structural-manager layer.
Owned nodes preserve complement literals, support counts, the recursive source
comparator and stable leaf ordering; `insdict.Dict` provides deterministic
structural interning. Operation paths flatten AND/XOR, fold XOR phases,
normalize MUXes, own prime truth tables and preserve the distinct AND support
permutation. Complete truth tables replay under checked external leaf mappings;
deterministic formulas round-trip through the Go DSD parser/evaluator, and
prime-subgraph detection replaces the active nondecomposable traversal. Atomic
formula import now replaces the temporary create/transform/rebuild network,
returns depth-first leaf mappings and rejects mux-local bindings and repeated
support before mutation. It preserves zero/one-input primes and matches the Go
parser/evaluator; its fuzzer passed 650,259 executions in addition to 618,078
structural-operation executions. Owned shared-support maps now preserve global
intersection selection, second-leaf order and composed phases; transparent
classification covers the active variable/AND/XOR/MUX/prime recursion. Merge
now handles constants and the complete disjoint-support path, translating its
local support permutation and phases back to owned global mappings. It checks
the exact declared intersection and source leaf-limit equation atomically; its
fuzzer passed 619,808 executions. The optional deterministic cache now stores
structural-local disjoint results, remaps them for each caller, returns owned
slices, enforces its entry cap and reports unshared hits/misses/entries.
Owned manager summaries and deterministic bounded formatters now replace the
stdout/file diagnostics, and independently allocated prime truth datasets are
stretched to manager width. Structural support cannot exceed the declared
manager domain. Shared-support merge now constructs the source local coordinate
system, composes mapping phases, evaluates both operands, runs general DSD
decomposition and atomically imports the result. Shared cache entries own
structural-local mappings and remap each hit for its caller; prime-size limits
reject before manager mutation. The single-leaf shared fuzzer passed 1,883,975
executions and the structured AND/XOR overlap fuzzer passed 1,508,987. Prime
creation now uses the validated `Abc_TtCanonicize` Go path, transferring input
permutations/phases and the output phase before interning. Formula import now
carries mappings through flattening, sorting, MUX normalization and prime
canonicalization; the permanent `[(ac)b]` regression fixes the stale source-
order mapping found by fuzzing, followed by 1,283,850 clean executions. The
same owned `MergeAnd` and cache path represents the MPM wrapper. Manager
summaries replace profiles/reports, active demos are ordinary tests, and the
wholly `#if 0` 6-feasible experiment has no runtime API. Every active function
is ledgered, promoting `dauTree.c` to validated; no C helper or `testdata/`
directory remains. The resulting full suite passes in 10.422s, the race suite
in 58.880s, and vet passes.

`opt/dau/dauNpn2.c` now has a checked immutable formula front end for its
six-variable `~`, `&` and `^` grammar. Parsing has byte/node/depth bounds and
cancellation; exact truth evaluation, deterministic normalization and atomic
GIA lowering replace assertion-driven recursive pointer ranges. In-memory
library verification and construction replace fixed-width line buffers and
ambient file reads. The source example evaluates to
`0x037d037d037d037d`, with every assignment also checked against an independent
Boolean model. The DTT core now preserves cyclic source-order NPN orbit
collection, five-operation candidate batches, additive node and max-delay
frontiers, exact class interning, structural node counts and the source's
per-fanin-pair multiplicity rule. Ordinary enumeration reaches exactly all 14
three-input NPN classes and matches independent exhaustive orbit closure;
delay/multiplicity invariants are replayed from retained fanins. Parser/
normalization fuzzing passed 1,874,986 executions, all-assignment GIA
differential fuzzing passed 1,421,650 and DTT frontier fuzzing passed 464,514.
Packed NPN configuration composition, variable spelling, source-order recovery,
all ten complemented gate types and retained left-fanin transforms now have
concrete checked representations; transform fuzzing passed 396,891. Every
retained implementation can emit a bounded canonical formula. Library
generation preserves first-source unordered fanin-class deduplication and
recursively cross-products every alternative fanin implementation in
deterministic order. Every emitted formula round-trips through the independent
Go evaluator; formula fuzzing passed 553,031 executions previously and another
279,912 after recursive expansion. Limits cover transforms, recursion depth,
individual and aggregate bytes and formula count, with cancellation throughout.
No test or generator compiles or executes C, and the Go module contains no C or
C++ helper source. A package-level guard rejects native helper source anywhere
under `pkg/` and rejects permanent vectors placed under `testdata`; reviewed
datasets remain under `test_vectors/`. The full suite passed in 10.235s, the race suite in 58.581s and vet
passed, promoting `dauNpn2.c` to validated.

`opt/dau/dauNpn.c` now has exact small-function NPN class enumeration, its
enumerator-pruning analysis and exact three/four-input function-by-node
frontiers. `EnumerateDauNPNClasses` returns owned ascending representatives and
canonical lookup for zero through four variables, covering all output phases,
input phases and permutations; it reproduces the known 1/2/4/14/222 class
counts without writing native table files. `AnalyzeDauNPN` returns direct
symmetric-pair counts, the source second-variable pruning mask and per-input
flip equivalence. `EnumerateDauFunctions` preserves candidate order,
symmetry skips, one- and two-gate expansion, reduced-last-level behavior and
first-class node costs while using `insdict.Dict` for deterministic interning;
owned level records replace stdout reports and persistence.
`EnumerateDauNetworks` separately preserves the older exhaustive candidate
frontier, including its optional two-node branch, without `tableW*.data`; it
reaches all 222 four-input NPN orbits and independently cross-checks them
against exhaustive orbit closure. Assignment-model pruning fuzzing passed
1,210,862 executions, canonical-orbit fuzzing passed 138,526,
function-frontier invariant fuzzing passed 21,658 and network-frontier fuzzing
passed 33,352. A bounded versioned little-endian codec now replaces the raw
host-dependent table stream, owned exact-NPN summaries replace report printing,
and deterministic hexadecimal rows replace per-node dump files; malformed
codec fuzzing passed 1,769,604 executions. Five/six-input function enumeration
and exact canonicalization now use bounded exhaustive `MinimizeNPN` above the
four-input lookup domain, with shared work limits and cancellation; wide-orbit
fuzzing passed 9,142 executions. Every source function now has a concrete Go
path or owned side-effect replacement, advancing `dauNpn.c` to validated. The
resulting full suite passes in 10.102s, the race suite in 57.312s, and vet
passes.

`opt/dau/dauNonDsd.c` now has an owned non-DSD set pipeline. Checked two-bit
roles represent free, unique and shared variables; discovery preserves source
bound-mask order, persistent position swaps, shared-subset order, containment
pruning and the at-most-two-cofactors criterion. A selected set produces owned
composition/decomposition truth tables, explicit original-variable mappings,
exhaustive reconstruction and deterministic DSD formulas with a synthetic
selector. Uppercase set parsing corrects the source's undefined negative shift.
Independent set-family fuzzing passed 437,360 executions, truth reconstruction
passed 573,793, formula reconstruction passed 483,584, and complete polarity-
family reconstruction passed 415,284. General decomposition is now used for
both components through twelve inputs. A seven-input composition regression
proves decomposable structural spelling beyond the former six-input boundary;
independent wide-composition fuzzing passed 372,122 executions. Deterministic
formatters and owned results replace the combined stdout reports, while the
three hard-coded source demos are represented by ordinary tests and deliberately
have no runtime API. This accounts for every source function without a C or
stdout oracle and promotes `dauNonDsd.c` to validated.
The resulting full suite passes in 10.321s, the race suite in 59.082s, and vet
passes.

`opt/dau/dauMerge.c` now has an owned two-formula merge pipeline. It validates
and applies both source-to-global permutations, excludes mux-local template
variables from presence analysis, builds shared-first maps, classifies pure
subtrees, extracts ordered definitions, composes the reduced conjunction,
inlines definitions, restores original variables, removes associative braces
and normalizes deterministically. Constants are permuted before identity
return, correcting the source's early-return omission. Reduced functions at
every supported width now use the general decomposer; the temporary
six-variable formatter and wide-prime fallback are gone. All storage is owned
and limits/cancellation expose no partial result. Brace rewrite fuzzing passed
1,304,160 executions, remapping fuzzing passed 1,336,068, and the integrated
assignment-exhaustive merge fuzzer passed 2,201,289. Debug-only printers guarded
by the source's false verbosity flag are omitted as process-global side effects;
the early-return demo has no observable result. This accounts for the complete
source and promotes it to validated. The full suite passes in 10.358s, the race
suite in 58.701s, and vet passes.

The complete `opt/dau/dauCount.c` source is now accounted for without a C
oracle. Its generated 8,192-entry byte-lane table has an owned Go snapshot and
stable language-neutral formatter, while deterministic unit tests and an
independent bit model replace the source's process-global randomized drivers.
The corrected Quick counter retains the source algorithm while fixing its
eight-input total overflow and high-variable sibling overwrite. Focused fuzzing
passed 695,560 executions. The resulting full suite passes in 9.542s, the race
suite in 55.479s, and vet passes.

`opt/dau/dauGia.c` now has an atomic native Go formula-lowering path.
`GiaManager.WithDauDSD` parses once and appends to an owned clone, covering
constants, phases, variables, balanced/direct multi-input AND/XOR, ordinary or
native XOR/MUX nodes, special mux-local bindings and hexadecimal prime functions
through twelve variables. It preserves an existing hash table or uses and
removes a temporary one, checks object limits/cancellation up front and during
recursion, and exposes no partial graph on error. Unit tests exhaust assignments
across all construction/hash modes and check ownership/error atomicity; semantic
fuzzing passed 1,152,035 executions against the independent DSD evaluator.
`WithDauTruth` now adds bounded general DSD dispatch, unused trailing-leaf
support and largest-prime reporting. `DeriveDauGia` validates each mapping,
computes ordered LUT-boundary truth, performs the source's level-balanced
lowering, preserves native-gate selection, buffers, CI/CO order, sequential
partitions and metadata, omits unmapped ANDs and cleans the owned result.
Returned `DauGiaStatistics` and deterministic formatting replace global
counters and reset-on-print behavior. Arbitrary truth lowering passed 898,066
fuzz executions; randomly generated mapped networks passed 814,724 rebuilds
with exhaustive input equivalence. Every source definition is now ledgered,
promoting `dauGia.c` to validated. The resulting full suite passes in 10.244s,
the race suite in 59.386s, and vet passes.

`opt/dau/dauInt.h` is now audited and validated as a no-artifact header. Its
pinned contents are only includes, empty organizational sections and namespace
delimiters; it declares no constants, types, macros, functions or data for Go to
represent. This mirrors the already validated empty `dau.c` translation unit
and advances one inventory entry without manufacturing boilerplate.

The coherent contiguous-bound-set portion of build-disabled
`opt/dau/dauDsd2.c` now has owned Go analysis and exact inverse composition.
General `[first,last)` analysis recognizes a common selector function with at
most two low-variable cofactors and returns the reduced truth table; low-bound
analysis recognizes `D(bound) XOR P(outer)` and returns both functions. Tests
exhaust every nonzero normalized selector mask for a three-variable bound set,
reconstruct all accepted inputs, reject altered and non-decomposable functions,
and cover dimensions, limits and cancellation. Independent reduced-function
fuzzing passed 1,116,884 executions and low-bound phase fuzzing passed 754,683.
The source remains partial: its disabled C text has existing syntax/undeclared-
local defects and its half-unshuffle, unfinished canonicalization and placeholder
object-recording paths are not yet ported. The resulting full suite passes in
9.442s, the race suite in 55.115s, and vet passes.

The complete `opt/dau/dauEnum.c` formula enumerator now returns owned sets in
the source's exact partition, complement, star-marker and insertion order.
Checked variable lifting rejects alphabet overflow; aggregate result and formula
byte limits plus cancellation replace static buffers and unbounded growth; and
deterministic formatting replaces the demo's fixed-path file and stdout writes.
Pinned sets match through three variables, every four-variable result reparses
and evaluates as a valid four-variable function, ownership and all bounds are
tested, and semantic enumeration fuzzing passed 1,324,293 executions. Coverage
advances `dauEnum.c` from unstarted to validated. The resulting full suite
passes in 9.382s, the race suite in 55.775s, and vet passes.

The complete build-disabled `opt/dau/dauArray.c` representation is now an owned
concrete Go pre-order array. It constructs constants, variables, complemented
nodes, AND/XOR/MUX/prime nodes and fanin/depth/minimum-variable metadata;
validates structure instead of asserting; returns copied node storage;
recomputes depths; preserves the source per-node AND cost; flattens nested XORs
and non-complemented nested ANDs; and returns deterministic diagnostics instead
of writing stdout. The source format has no field for mux-local bindings, so
those formulas are explicitly rejected. Focused structure and ownership tests
pass, and generated-formula fuzzing passed 1,303,358 executions. Coverage
advances `dauArray.c` from unstarted to validated. The resulting full suite
passes in 9.481s, the race suite in 55.499s, and vet passes.

The complete active `opt/dau/dauCore.c` template shell is now represented by
`DauTryPerform`. The C source only allocates a two-pointer context, performs no
work, frees it and unconditionally returns one; the Go operation returns the
same success result without exposing redundant allocation/free lifetime or
inventing an optimizer. Tests cover nil and non-nil graphs, extreme parameter
values and graph non-mutation. Coverage advances this source from unstarted to
validated, and the full suite passes in 7.223s.

`opt/dau/dauDsd.c` now has a substantial first native Go slice: owned bounded
parsing, checked delimiter matching, variable discovery, deterministic
normalization, source-compatible AND-cost accounting, hexadecimal prime
functions, ordinary MUXes and the six-variable evaluator's special local-binding
MUX form. Its six-variable truth-to-formula decomposition now also preserves
the source's support minimization, output factoring, two-input substitution and
best-variable MUX fallback order. All 65,536 four-variable Boolean functions
decompose, reparse and exhaustively evaluate to their original truth words;
full-width fuzzing passed 601,659 random 64-bit truth functions. Concrete
indexed nodes avoid interfaces; malformed syntax, excessive
depth/storage, cancellation and truth-table bounds return atomic errors instead
of assertions or partial output. Exhaustive assignment checks cover constants,
complements, AND/XOR/MUX, one- through three-variable primes, nesting and local
bindings. Parser/render fuzzing passed 1,797,753 executions and generated
normalization/truth round trips passed 1,237,470. Variable permutation retains the source mapping and swap
schedule but uses an explicit SplitMix64 seed instead of process-global,
libc-dependent `rand()`; independent assignment-remapping fuzzing passed
1,073,607 executions. The general non-split decomposer now preserves minimum
base reduction and the source output-variable, pair and MUX scan stages through
the full twelve-variable domain, emitting truth-bearing irreducible primes for
independent parser replay. All zero- through three-variable functions pass
exhaustive semantic checks; wide/nested cases, source ownership, every bound
and cancellation are covered, and semantic fuzzing passed 2,128,795 executions.
One-step prime analysis now returns the source's fully decomposable, selected
variable and unsplittable statuses. Split-prime output preserves natural
variable priority, minimum cofactor-support selection, the local MUX binding
form, explicit failure reporting and the source rule that outer disjoint MUX
cofactors disable recursive splitting; semantic fuzzing passed 1,561,400
executions. Level-aware priority now preserves the source selection sort's tie
behavior and recursively accounts for synthesized definitions; the owned brace
pass runs on generated formulas, and the level-aware semantic rerun passed
1,499,751 executions. The duplicated specialized-six and wide manager paths are
explicitly consolidated into the general owned engine; direct support
calculation replaces their private cache, and owned expressions replace fixed
definition/write storage. Formula results replace the print wrappers, while
no-op, fixed-vector and ambient-file benchmark drivers are represented by
ordinary regressions and fuzzing rather than runtime APIs. The active Test44
local-binding and Test888 nine-input formulas also round-trip independently.
Every active source behavior now has a Go definition or explicit side-effect
disposition, promoting `dauDsd.c` to validated. The resulting full suite passes
in 10.251s, the race suite in 59.129s, and vet passes.

The complete `opt/dau/dauDivs.c` pairwise divisor analysis now returns owned
rows and deterministic text rather than printing stdout. It preserves variable
pair order, all five four-bit cofactor classes and function/complement pairing,
with atomic pair/digit limits and cancellation. Unit tests cover every category,
complements and empty pair domains; independent assignment-level fuzzing passed
71,793 executions across zero through eight variables. The source demonstration
entry point has no observable behavior before its early return and is represented
by an explicit side-effect-free Go operation. Coverage advances this source from
unstarted to validated. After both current source closures, the full suite passes
in 7.350s (7.78s wall), vet in 1.44s and the race suite in 49.136s (63.33s wall).

`aig/gia/giaSatLut.c` is now fully accounted for. The existing concrete Go
surface already covered cut generation, mapping-state recovery, structural CNF,
reusable SAT solving, unit/edge and boxed/unboxed timing feedback, checked
replacement, whole-network scheduling, four/six-input search and area-only
mode. `GiaManager.FullWindow` now completes the whole-graph window path, while
owned cut and pass-stat formatting completes the diagnostic surface without
stdout or nondeterministic runtime sampling. Focused tests pass and diagnostic
mask fuzzing passed 101,078 executions in ten seconds. Coverage advanced from
partial to validated for this source.

The permanent test system is now independent of the C implementation. All 287
listed artifacts are hash-verified by the schema-2 stand-alone manifest; semantic
checks combine those immutable datasets with Go-only exhaustive models,
certificate replay, invariants, round trips and fuzzing. All 287 `.c` generators,
the native smoke runner and compatibility header were removed from `pkg/tools`,
and package tests reject their reintroduction. The unused host/binary-specific
`native_smoke.json` transcript was also removed. Focused stand-alone tests pass;
the complete suite passes in 7.337s (7.80s wall) and vet in 1.47s without an ABC
binary/archive, C compiler, C headers, cgo or subprocess oracle. The final race
suite passes in 49.268s (63.40s wall).

The complete `misc/tim/timDump.c` version-1 persistence format now has bounded,
owned Go save/load support. Empty, no-box and two-box managers match native
hierarchy-only and full streams byte-for-byte, including big-endian signed
metadata, pinned-host little-endian float32 payloads, negative copy IDs, delay
tables and primary timing values. Loading rejects truncation, trailing bytes,
bad versions/IDs/dimensions and resource overruns atomically. Saving rejects
noncanonical topology, sparse tables and black-box flags that the native format
would silently lose. Native load/resave fixtures are identical. Malformed-stream
fuzzing passed 2,157,361 executions and independently generated valid-manager
round trips passed 1,305,440. The full suite passed in 7.410s (7.85s wall), race
in 52.445s (66.56s wall), and vet passed in 1.42s. The already implemented
`timTime.c` and `timTrav.c` surfaces were audited function-by-function and are
also now validated; the required-time implementation retains its documented
per-arc guard correction. Inventory advanced by three files.

The final `timBox.c` lookup now preserves `Tim_ManBoxFindFromCiNum` exactly:
primary inputs return -1, outputs before a later box identify the preceding box,
and all outputs of the final box return the native -2 sentinel. A compact
permanent digest covers 294 native rows across zero through four boxes and every
valid CI. Invalid indices return errors instead of reading outside C storage.
Association fuzzing passed 307,916 executions. The final full suite passed in
7.373s (7.85s wall), race in 49.361s (63.35s wall), and vet in 1.43s. With all
creation, association, range, metadata, delay-table, copy, black flag and lookup
operations accounted for, `timBox.c` is validated. Manager duplication,
unit-delay conversion, primary timing extraction, trim/reduce/alignment, atomic
library configuration, deterministic formatting/statistics and Go-owned cleanup
complete `timMan.c`. Its two model fuzzers passed 1,990,861 and 1,940,338
executions; configuration fuzzing passed 2,044,441 and formatting fuzzing passed
1,199,710. The public/internal headers and `module.make` are also validated, so
the complete active `misc/tim` module is accounted for.

A concrete gzip file lifecycle now spans `gzlib.c`, `gzread.c`, `gzwrite.c` and
`gzclose.c`. Readers support gzip/transparent raw input, concatenated members,
rewind, replay and deferred beyond-EOF seek, logical tell, effective compressed
offsets, EOF/error clearing, buffer/byte/line reads and bounded multiple-byte
LIFO pushback. Pushback preserves native negative logical positions, sticky
overflow errors, clearing and EOF visibility; a four-byte buffer accepts eight
bytes when its output area is empty. Writers support truncate/append members,
levels, buffer/byte/string
writes, forward zero-fill seek, tell, effective compressed offsets, flush and
owned close. Path and descriptor constructors both transfer ownership; descriptor
streams preserve a nonzero starting offset. Native lifecycle and descriptor
state-machine fixtures match, including both close paths and far-seek collapse
from requested position 100 to actual decompressed EOF 10. Go's standard encoder
has a different sync-flush byte count, so its reported offset is checked against
the real descriptor position rather than zlib's encoded length. File-backed
round-trip fuzzing passed 17,461 executions, pushback fuzzing passed 23,437 and
concatenated-member/trailing-raw fuzzing passed 19,391. The trailing-raw
increment's full tests passed in 7.300s, race in 49.172s and vet passed. Clean
fixture integrity verification and all hashes match.
`gzclose.c` and `gzread.c` are validated. The separate bounded Go pushback stack
preserves the native guarantee without exposing private output-buffer occupancy.
`gzlib.c` and `gzwrite.c` remain partial pending exact ignored mode and strategy
parsing, formatted writes, strategies and midstream parameter/flush modes.

The pinned non-Windows diagnostic in `misc/util/utilColor.c` now returns owned
ANSI text rather than mutating process-global stdout. Its complete 4,178-byte
output matches the native SHA-256 exactly, including spacing, escape sequences
and the native first-row style formatting. Independent tests check all 18
foreground rows, 144 samples, 150 reset sequences, envelope and determinism.
There is no input domain to fuzz. The full suite passed in 7.252s, the race
suite in 48.667s and vet passed; fixture integrity verification and all hashes match. The separate
Windows console-attribute branch is inactive in the pinned build.

All three process/temp-file wrappers and declarations in `misc/util/utilSignal.c`
and `.h` now have concrete Go behavior. Shell execution preserves the pinned
non-Windows `system()` wait-status encoding. Temporary files preserve prefix,
suffix, uniqueness, mode 0600 and caller ownership, but use one atomic creation
instead of native's unlink/reopen suffix race; removal errors are checked rather
than discarded. Native property records match and filesystem fuzzing passed
39,282 executions. After the invalid-path regression was added, full tests
passed in 7.354s, race in 48.196s and vet passed. Clean fixture integrity verification, source
identity and all hashes match.

The portable public utilities in `misc/zlib/zutil.c` now match the pinned
native configuration: version 1.2.5, compile flags 169 and all nine defined
error strings. Invalid error codes return an explicit validity result rather
than indexing outside native static storage. Go lifetime management replaces
the active allocation/free hooks; debug, legacy 16-bit, Windows CE and fallback
memory branches are inactive in the reference build. Bounds fuzzing passed
445,049 executions. The full suite passed in 7.323s, the race suite in 49.153s
and vet passed; clean dataset integrity verification and every hash match.

The complete `misc/bzlib/crctable.c` data leaf now has an owned Go spelling.
All 256 MSB-first CRC entries are generated from polynomial `0x04c11db7`, match
an independently evaluated recurrence and reproduce the canonical native table
digest and endpoints. Returning the fixed array by value prevents mutation of
package storage. There is no input domain to fuzz. Full tests passed in 7.260s,
race in 49.003s and vet passed. Clean fixture integrity verification, source identity and all
hashes match.

The complete `misc/bzlib/randtable.c` data leaf now returns an owned copy of
all 512 native randomization values. Every entry matches the frozen compatibility dataset and
falls in the native 50-through-999 range; mutating one result cannot affect the
next. There is no input domain to fuzz. Full tests passed in 7.203s, race in
48.881s and vet passed. Clean fixture integrity verification, source identity and all hashes
match.

All three algorithms in `misc/bzlib/huffman.c` now have concrete owned Go
operations. Length construction preserves native heap tie order, zero-frequency
normalization and bounded rescaling; canonical assignment and complete
23-entry limit/base decoding tables preserve native symbol order. Ten native
digests match from two through 258 symbols, including a skewed rescaling case.
Prefix/decode fuzzing passed 655,359 executions. Full tests passed in 7.093s,
race in 49.125s and vet passed. Clean fixture integrity verification, source identity and all
hashes match.

The self-contained cumulative-frequency lookup in `misc/bzlib/bzlib.c` now has
a checked concrete binary search. All 144 native cases match across duplicate
boundaries, an all-zero table, exact edges, negative queries and values beyond
the terminal total. Linear-model fuzzing passed 128,840 executions. Full tests
passed in 7.394s, race in 48.982s and vet passed. Clean fixture integrity verification and all
hashes match. `bzlib.c` moves from
unstarted to partial; its streaming, buffer/file and query APIs remain unported.

The pinned version query and memory-buffer decompressor in `misc/bzlib/bzlib.c`
now have concrete Go forms. The standard-library bzip2 reader decodes all six
native streams from empty through 8,192 deterministic bytes, returns owned
storage, enforces a checked output cap, polls cancellation and returns no partial
bytes on failure. Tests cover undersized caps, malformed and truncated streams,
invalid bounds and ownership; adversarial fuzzing passed 87,211 executions. The
full suite passed in 7.395s, race in 49.601s and vet passed. Clean fixture integrity verification,
source and fixture hashes match. `bzlib.c` remains partial:
buffer compression, streaming/file APIs, error queries, assertions and private
state machines are still unported.

The public update, lookup-table and concatenation surface of
`misc/zlib/crc32.c` is now complete. Go's standard IEEE implementation replaces
the native alignment/endian fast paths, the exposed table is an owned copy, and
the GF(2) concatenation logic supports full 64-bit second lengths. The native
256-entry table digest, 32 updates and 16 combines match exactly, including
lengths beyond one terabyte. Independent standard-library/chunk composition
fuzzing passed 251,956 executions. The full suite passed in 7.251s, the race
suite in 48.868s and vet passed; clean dataset integrity verification and all hashes match.

The complete streaming and concatenation surface of `misc/zlib/adler32.c` now
has concrete Go functions. Arbitrary initial checksums, native nil-buffer
initialization, zero-length normalization, the 5,552-byte reduction boundary,
incremental updates and 64-bit-length combination are preserved. All 32 native
updates and 16 combinations match exactly, including lengths beyond one
terabyte. Independent standard-library and chunk-composition fuzzing passed
159,178 executions. The full suite passed in 7.118s, the race suite in 49.367s
and vet passed; clean dataset integrity verification and every source/fixture hash
match.

The memory-buffer APIs in `misc/zlib/compress_.c` and `uncompr.c` are now
replaced by concrete standard-library zlib operations. Compression supports
native levels -1 and zero through nine, and `compressBound` uses the exact
native formula with overflow detection. Both directions return owned bytes,
enforce explicit output caps, poll cancellation and expose no partial result.
All seven pinned native streams decode exactly; Go streams are format-valid and
round-trip but need not have identical encoded bytes. Compression fuzzing passed
96,510 executions and adversarial decompression fuzzing passed 632,970. The full
suite passed in 7.221s, the race suite in 49.811s and vet passed; clean fixture
fixture integrity verification and all source/fixture hashes match.

Owned truth-table variable exchange now validates all four routines in
`bool/lucky/luckySwapIJ.c`. The existing word-level implementation covers
same-variable identity, within-word masks, mixed word/bit exchange and
whole-block exchange while avoiding the native case-3 temporary allocation
leak. Native aggregate digests match every variable pair through twelve inputs,
364 cases total. Independent assignment-level fuzzing passed 726,434 executions;
the full suite passed in 7.157s, the race suite in 48.844s and vet passed. Clean
fixture integrity verification and source/fixture hashes match.

All three active compatibility routines in `map/if/ifMatch2.c` now have
concrete Go spellings. Setup and setdown preserve their empty bodies; cell
lookup preserves its unconditional success and leaves every argument untouched.
The unused opaque native manager is a `uintptr`, avoiding both a fabricated IF
manager and a Go interface. Sixteen native records spanning varied pointers,
truth words, dimensions and descriptions reproduce byte-for-byte. Argument and
immutability fuzzing passed 787,510 executions; the full suite passed in 7.302s,
the race suite in 48.312s and vet passed. Source and fixture hashes match.

The complete leaf implementation in `map/mpm/mpmLib.c` now has a concrete,
owned Go LUT-library value. All sizes from zero through twelve match native
maximum size, variable-delay flag, 13-entry area table and all 169 delay cells;
out-of-range sizes return a checked error and Go lifetime management replaces
the native free routine. Thirteen native records reproduce byte-for-byte, source
source and fixture hashes match, and ownership/error tests pass.
Deterministic fuzzing passed 1,625,649 executions; the full suite passed in
7.120s, the race suite in 49.025s and vet passed.

Nineteen callable native stubs across fourteen sources now have concrete Go spellings
rather than being misclassified as empty translation units. Placement
legalization, ACB/WLC tests, three `giaSim5.c` lifecycle hooks, CFS, MPM and
simulation hooks preserve explicit no-op bodies; `Gia_Sim4Try` and PLA merge
return zero, and the generators return nil. The expanded batch also covers the
empty ECO/generator hooks, null/empty CBA binary hooks, and zero/null IIFF
placeholders, plus both empty template-package lifecycle hooks. Concrete `*GiaManager` protects nonmutation contracts, fixed arrays
represent filename lists, and `uintptr` is used only for genuinely opaque,
uninspected native pointers. Sixteen varied native records match, and the
expanded fuzz target passed 762,290 executions. The latest full suite passed in 7.247s,
race in 49.528s and vet passed; clean dataset integrity verification and every hash match. The
pinned default archive omits `mpm.c`, so the unchanged pinned source is compiled
directly into the oracle.

Forty-nine additional translation units are now validated as intentional
placeholders. Each contains only one include and namespace delimiters after its
file comment or empty documentation scaffolding, with no declarations,
definitions, data or executable behavior; no fabricated Go API was added. The
batch spans AIG, base, BDD, Boolean, timing,
mapping, optimization, proof, SAT and root package directories. Every local
file hash matches both the inventory and pinned-reference source. Together with
the previously validated `aig_.c`, the ledger now accounts for 50 empty native
translation units explicitly.

Both default initializers in `opt/fsim/fsimCore.c` now return concrete owned Go
value records. Every simulation and switching field matches native ABC,
including 8/32/60 simulation sizing, 1/48/16 switching sizing, enabled verbose
and probability flags, and all remaining zero values. Repeated construction is
mutation-isolated. The routines have no input domain to fuzz. Clean native
fixture integrity verification and all source/fixture hashes match; full tests passed
in 7.247s, race in 49.528s and vet passed.

The complete ternary state engine in `opt/fsim/fsimTsim.c` now operates directly
on concrete AIG managers. It preserves native two-bit values, unknown primary
inputs, zero register initialization, complemented/AND propagation, packed
uint32 hashing and the retained repeated terminal state, while deterministic
slice buckets replace native linked allocations. Four native trajectories match
exactly through a 17-register packing boundary; formatting, ownership, limits,
cancellation and unsupported objects are checked. Graph fuzzing passed 551,851
executions. The legacy sources require a separately hashed compatibility include
for renamed AIG operations and zeroed ternary storage; unchanged local and pinned
sources match. Full tests passed in 7.259s, race in 49.855s and vet passed.

The complete allocation lifecycle in `misc/mvc/mvcMan.c` now has a concrete Go
owner ready for later cube and cover stages. Go allocation replaces the native
four-pool memory manager, and checked close rejects nil or repeated release
instead of permitting invalid frees. Thirty-two simultaneous native managers
confirm non-null, independent managers and pools; Go ownership and reverse/early
close cases match the lifecycle contract. Fuzzing passed 719,269 executions.
Full tests passed in 7.264s, race in 49.300s and vet passed; clean fixture verification,
source and fixture hashes match.

All of `misc/mvc/mvcCube.c` and the allocation constructor in `mvcCover.c` now
provide concrete packed Go values. Cover dimensions preserve 32-bit word and
unused-tail layout; cubes use deterministic zeroed Go storage in place of the
native pool's indeterminate payload and support owned duplication, checked
bit/word access, release and exact adjacent-pair don't-care removal. Ten native
widths from zero through 127 bits match word-for-word across every boundary.
Independent pairwise fuzzing passed 634,155 executions. Full tests passed in
7.287s, race in 49.590s and vet passed; clean dataset integrity verification, pinned source
identity and every hash match.

The complete `misc/mvc/mvcApi.c`, `mvcCover.c` and `mvcList.c` layer now uses a
concrete owned ordered slice rather than exposing native linked-list headers.
It provides checked head/tail insertion, ownership-transferring deletion,
empty clone, ordered deep duplication, close, reusable mask and literal-count
storage, cube-array snapshots and atomic permutation application. Listed cubes
and the retained scratch mask reject direct close, while the two native empty
literal-update hooks remain no-ops. Five native workflows from zero through 65
bits match exact order, clone/duplicate, reuse, deletion, empty, tautology and
binary-buffer results; zero-bit tautology is preserved. Ownership and atomic-
failure tests pass, and ordering/deep-copy fuzzing passed 330,103 executions.
Full tests passed in 7.391s, race in 50.118s and vet passed; clean native
fixture integrity verification, all three pinned source identities and every fixture
hash match.

All five comparators in `misc/mvc/mvcCompare.c` now operate directly on packed
concrete cubes. They preserve high-word-first unsigned integer order,
descending population priority with integer tie-breaking, selected-bit order,
the native OR-with-mask outside order, and outside-then-under partition order.
Population is computed directly rather than kept in the native stale-prone
`nOnes` header; equal partition inputs return zero instead of reaching an
assertion. All 96 native records match over twelve widths through 127 bits.
Tests add zero-width, equality, dimension and lifetime cases, and independent
word-model fuzzing passed 223,139 executions. Full tests passed in 7.290s, race
in 49.374s and vet passed; clean dataset integrity verification, pinned source identity and every
fixture hash match.

The complete `misc/mvc/mvcSort.c` and `mvcContain.c` algorithms now use concrete
ordered slices and an `MvcCubeOrder` enum instead of a callback or interface.
Recursive half splitting and right-half-first selection on ties preserve native
merge order for all five comparison modes. Containment preserves population/
integer sorting, adjacent duplicate deletion, larger-group subset removal,
changed reporting and exact survivor order while closing removed owned cubes.
All 60 native sort orders and twelve containment results match through 127 bits.
Tests cover invalid mode/mask failure atomicity, zero-bit duplicates, lifecycle
and idempotence; independent maximal-set fuzzing passed 281,640 executions. The
direct popcount replacement also partially covers the two size routines in
`mvcUtils.c` without introducing cached state. Full tests passed in 7.265s,
race in 49.275s and vet passed; clean dataset integrity verification, all three pinned source
identities and every fixture hash match.

All seven routines in `misc/mvc/mvcLits.c` now have concrete checked cover
operations. They preserve high-to-low selection for any repeated literal,
lowest-index best/worst ties, first-cube bit order, total population,
best-literal one-cube construction and the native predicate that accepts one
cube with at most one literal, including a zero cube. Owned frequency arrays
replace the dormant native counter storage, and invalid masks or absent best
literals return errors instead of permitting a negative bit shift. Eleven
native records and five predicate cases match through 127 bits; independent
frequency-model fuzzing passed 136,568 executions. Full tests passed in 7.275s,
race in 49.759s and vet passed; clean dataset integrity verification, pinned source identity and
every fixture hash match.

All three operations in `misc/mvc/mvcOpAlg.c` now have concrete checked cover
implementations. Multiplication preserves native Cartesian OR order,
subtraction preserves left-cube order, and equality preserves native
directional membership semantics, including ignored extra right cubes and
multiplicity. `Mvc_CoverSupport` is also implemented as an owned union cube and
recorded against the still-partial `mvcUtils.c`; overlapping multiplication is
supported without reproducing native's incidental stdout warning. Nine native
boundary records through 127 bits match exact counts, digests, support relations
and equality quirks, while independent model fuzzing passed 166,880 executions.
Full tests passed in 7.189s (7.67s wall), race in 49.254s (62.42s wall), and vet
passed in 1.29s. Clean dataset verification is byte-identical, and pinned source,
source and fixture hashes match.

The sole compiled operation in `misc/mvc/mvcOpBool.c`,
`Mvc_CoverBooleanOr`, now returns a checked independent deep-copy
concatenation in exact left-then-right order, retaining duplicates and contained
cubes. The native Boolean AND/equality sketches are inside `#if 0` and are
accounted for without invented Go APIs. Nine native records from zero through
127 bits match complete result and unchanged-input digests; ownership/error
tests pass, and direct concatenation-model fuzzing passed 550,696 executions.
Full tests passed in 7.344s (7.82s wall), race in 49.349s (62.60s wall), and
vet passed in 1.35s. Clean dataset verification is byte-identical, with matching
source and fixture hashes.

All compiled formatting behavior in `misc/mvc/mvcPrint.c` now returns concrete
owned strings without adding a writer interface or mutating global stdout.
Raw bit order, binary `0`/`1`/`-` pair mapping, headings, newlines,
literal-count lines, zero-bit blank cube lines, ignored odd tail bits and the
native iterator's early stop at a `00` pair are preserved. The multivalued
printers under `#if 0` intentionally have no Go API. A complete ten-width
native transcript through 127 bits matches by SHA-256; error tests pass, and
independent bit/pair-model fuzzing passed 642,382 executions. Full tests passed
in 7.221s (7.70s wall), race in 49.791s (63.10s wall), and vet passed in
1.29s. Clean dataset verification and source/fixture hashes all
match.

The division-enabling common-cube/support slice of `misc/mvc/mvcUtils.c` now
has concrete checked operations for support intersection, binary support size
and membership, common cubes, cube-free detection/normalization, common-cube
cover construction and directional support containment. Owned cubes replace
native retained-mask exposure; empty positive-width and odd-width behavior is
preserved. Eight native paired-width records through 126 bits match exact
common/intersection and ordered-cover digests, memberships, normalization and
both containment directions. Independent union/intersection fuzzing passed
427,721 executions. Full tests passed in 7.382s (7.87s wall), race in 48.823s
(62.18s wall), and vet passed in 1.31s; clean dataset integrity verification and all hashes match.
`mvcUtils.c` remains partial with 11 covered routines; pair differences,
remapping, cofactors, quantification, expansion, transpose and related
transforms remain.

All five public routines in `misc/mvc/mvcDivide.c` now implement checked
algebraic division. Literal and cube paths preserve source order; multi-cube
division preserves native in-place sorting, outside-support grouping,
under-support merges, failed-group remainder order and quotient extraction.
Empty/zero-literal divisors are rejected instead of reaching invalid native
indices, and cube containment checks every word instead of only two. Four
native cases through 127 bits match sorted-input and all quotient/remainder
digests plus count/support rejection. Literal/cube model fuzzing passed 343,615
executions and constructed algebraic-reconstruction fuzzing passed 483,532.
Full tests passed in 7.148s (7.63s wall), race in 50.050s (63.48s wall), and
vet passed in 1.45s; clean fixture verification and all hashes match.

`misc/mvc/mvcDivisor.c` now has an iterative, checked quick-divisor kernel that
preserves native initial rejection, independent duplication, lowest-index
least-frequency repeated-literal selection, quotient order and common-cube
removal without recursive stack growth. Seven native multi-stage products from
six through 127 bits match exact result and unchanged-source digests; native
single/no-repeat nil paths match, and independent kernel-model fuzzing passed
520,761 executions. Full tests passed in 7.332s (7.83s wall), race in 49.102s
(62.43s wall), and vet passed in 1.44s; clean fixture verification and all hashes match.

The remaining active transforms in `misc/mvc/mvcUtils.c` are now complete:
ordered uint8 pair differences, fill-and-copy remapping, inversion, don't-care
removal, cofactor, adjacent binary flip, universal quantification, transpose
and unused-tail validation. Checked errors replace native assertions, and tail
validation reports state rather than printing while always returning one.
Seven native transform cases through 126 bits match exact digests, completing
15 records across the two utility fixtures; independent transform-model fuzzing
passed 503,153 executions. Full tests passed in 7.343s (7.86s wall), race in
49.133s (62.62s wall), and vet passed in 1.41s. Both utility helpers reproduce
byte-for-byte with matching hashes. `mvcUtils.c` moves from partial to fully
validated; only routines under native `#if 0` lack Go APIs.

The MVC package manifest is now validated: all fifteen C units named by
`misc/mvc/module.make` are fully validated in the single Go package, with Go's
package build replacing object enumeration. `mvc.h` moves from unstarted to
partial. Its 32-bit packing, active manager/cover/cube/list concepts,
bit/list/iteration operations and implemented declarations are represented,
but the header also advertises missing-source legacy Boolean, complement,
sharp, reshape and merge APIs plus dormant `Mvc_Data_t` and allocator state.
Those declarations remain explicit outstanding scope rather than fabricated
Go APIs.

The complete `misc/nm` name manager is now a concrete checked Go owner backed
by two `insdict.Dict` indexes. It preserves unique IDs, suffix concatenation,
typed and wildcard lookups, two-type fallback, native duplicate-name ring order
and oldest-entry anchor rotation. Owned strings/slices replace arena pointers;
physical bucket hashing and stdout profiles become deterministic dictionary
growth and an owned summary. `NameIDs` intentionally fixes native
`Nm_ManReturnNameIds`, whose pinned 64-entry case returns only 37 bucket heads.
The native workflow reproduces byte-for-byte, and independent operation-model
fuzzing passed 102,829 executions. Full tests passed in 7.373s (7.84s wall),
race in 49.482s (63.16s wall), and vet passed in 1.39s. Both C files, both
headers and `module.make` are validated with matching source and fixture hashes.

The compiled `misc/parse` equation parser now targets the concrete AIG manager
without waiting for HOP. It preserves constants, exact variable lookup,
negation, AND/XOR/OR precedence and associativity, whitespace, parentheses and
the native `a!b` shorthand. Typed internal slices replace both fixed raw stack
families; byte-offset format errors, token/depth limits and cancellation replace
stdout diagnostics, overflow and underflow. Clone-then-commit keeps every
failure atomic. Seventeen native truth tables and nine rejection cases match;
10,000 nested parentheses and 5,000 terms pass, and independent semantic fuzzing
passed 33,630 executions. Full tests passed in 7.332s (7.80s wall), race in
49.947s (63.68s wall), and vet passed in 1.32s. `parseEqn.c`, `parseStack.c`,
`parseInt.h` and the two-unit manifest are validated with exact hashes. The
separate CUDD `parseCore.c` and its public `parse.h` remain unstarted.

Both legacy `misc/st` table variants now share one concrete generic
`StableTable[K,V]` backed by `insdict.Dict`. It preserves insert-versus-replace,
lookup, mutable find/add slots, direct duplicate shadowing, reveal-on-delete,
shallow cloning, callback stop/delete and pooled-table clearing. Concrete slots
and cursors reject stale use; owned snapshots and cursors enumerate every live
duplicate in deterministic insertion order instead of native bucket, rehash and
move-to-front order. Checked helpers retain both string hashes and the numeric
and shifted-address calculations. Both native workflows and hashes match;
duplicate-aware model fuzzing passed 104,762 executions. Full tests passed in
7.286s (7.78s wall), race in 49.211s (63.09s wall), and vet passed in 1.43s.
Both C files, both headers and `module.make` are validated with exact hashes.

Owned sequential ternary-state analysis now covers the deterministic core of
`aigTsim.c`: native two-bit propagation, packed-state hashing, zero-register
initialization, PI unknowns, 300 exact transitions, widening, repeated-state
detection, union and constant substitution mapping. All 32 native sequential
machines match exact mappings, and 32 state hashes match uint32 overflow across
the 128-prime cycle. Tests cover simultaneous latch commits, ownership,
malformed/buffer cases, exact limits, cancellation, combinational graphs and a
10,000-node chain. Exhaustive Boolean-reachability fuzzing passed 1,276,145
executions and independently proves every reported constant. The full suite
passed in 7.190s, race in 49.017s and vet passed; clean fixture verification and all
hashes match. Its ordinary constant-reduction loop now repeatedly composes
ternary analysis, CI remapping and full sequential cleanup while retaining an
original-register projection. All 32 final native arenas match, including
cases where native's stale allocation counters fail its own check while Go
passes stronger validation. Thirty-two-frame trace fuzzing passed 1,248,073
executions. The final suite passed in 7.342s, race in 48.459s and vet passed;
clean dataset integrity verification and hashes match. `aigTsim.c` remains partial only for
multivalued sweeping, print formatting and unused random helpers.

Owned atomic CI remapping now expands `aigScl.c` coverage. All 32 signed-CI,
constant and cross-register substitution cases match native CI/CO order,
register/constraint partitions and every object field. Remapping precreates all
CI slots, collapses buffers, structurally rebuilds ANDs, preserves the source
and rejects choice metadata instead of silently dropping it. Tests add exact
limits, cancellation, buffer collapse and a 10,000-node chain. Exhaustive
five-input semantic fuzzing passed 1,380,510 executions. The full suite passed
in 7.193s, race in 48.569s and vet passed; clean fixture verification is byte-
identical and all hashes match. `aigScl.c` remains partial pending latch
reduction, SCC analysis, partitioned sweeping and the high-level SCL pipeline.

Concrete AIG DOT rendering now covers both functions in `aigShow.c`. All 16
phase/style cases match native output byte-for-byte, including exact ranks,
object order, constant and dangling-node display, complement edge styles and
bold nodes. Temporary native object marks and CO-level mutation are replaced by
owned local state, and errors replace printed diagnostics. The native 200-AND
default is retained alongside object, byte and cancellation limits. The show
wrapper preserves final-extension replacement and explicit file output, but
returns the path instead of launching `Abc_ShowFile`. Tests add file failures,
invalid bold IDs, source immutability and a 10,000-node chain. Deterministic
structural fuzzing passed 432,643 executions. The full suite passed in 7.189s,
the race suite in 48.281s and vet passed; clean fixture verification is byte-
identical and all hashes match. `aigShow.c` is the sixteenth validated item.

Owned atomic forward-frontier retiming now covers all three functions in
`aigRetF.c`. The implementation preserves native reachability marking and
short-circuit order, performs un-hashed intermediate construction, skips
buffers, and dependency-normalizes forward references into a checked manager.
All 32 phase and one-through-three-step cases match every terminal and object
field from a narrowly corrected frozen compatibility dataset. The correction is necessary:
the pinned native setup aliases CIO IDs with a union field asserted null by
retiming, and its final ordered duplicator crashes on the forward CI references
that successful retiming creates. The oracle avoids the unrelated setup alias
and changes only final duplication to precreate CIs and copy cones by DFS.
Independent transition-language tests compare all one-step behaviors after a
warmup frame, and tests also cover ownership, malformed/combinational inputs,
limits, cancellation and 10,000 levels. Retiming fuzzing passed 350,436
executions. The full suite passed in 7.003s, the race suite in 39.985s and vet
passed; fixture integrity verification is byte-identical, and source and fixture hashes match. `aigRetF.c` is the fifteenth validated
inventory item.

Owned atomic AIG sequential cleanup now covers both native modes. Full cleanup
retains precisely the primary terminals and register closure reachable from
primary outputs; basic cleanup retains all slots and patches unreachable latch
inputs to constant zero. All 32 dependency graphs match native terminal arrays,
tombstones and every object field in both modes. Native full cleanup's own
check fails because removed terminals do not update its redundant allocation
counters; the equivalent Go result passes complete structural validation.
Tests cover ownership, errors, limits, cancellation and 10,000 levels. The
independent closure/semantics fuzzer passed 1,572,899 executions. The full suite
passed in 6.912s, race in 39.618s and vet passed; clean fixture verification is byte-
identical with matching hashes. `aigScl.c` moves from unstarted to partial.

The fixed `aigTest.c` procedure now returns an independently owned checked AIG
instead of writing `aig_test_file.blif`. Its complete object arena, CI/CO order
and one-register partition match native ABC exactly, and all eight assignments
independently verify the MUX and AND outputs. The full suite passed in 7.060s,
the race suite in 39.840s and vet passed; fixture integrity verification is byte-
identical with matching hashes. `aigTest.c` is the fourteenth validated item.

Structural AND hashing now has lookup-only, count and owned native-bucket
profile operations backed by the manager's concrete lookup-only Go map. Across
32 phase cases, trivial, existing and missing results plus every nonempty bucket
occupancy match native ABC both before and after its explicit resize. Tests add
lookup nonmutation, owned profiles, deletion cleanup, malformed inputs, limits
and cancellation. Arbitrary construction fuzzing passed 1,544,546 executions.
The full suite passed in 7.013s, the race suite in 40.464s and vet passed; clean
dataset verification is byte-identical and all hashes match. Forward retiming
now supplies the remaining table-clear lifecycle with a private unhashed net
that is normalized into a fresh checked hash before it can escape. This avoids
exposing native's temporarily unusable manager state. `aigTable.c` is therefore
the seventeenth validated item.

Concrete owned AIG timing snapshots now replace native manager-attached direct,
reverse and required-level scratch. All 32 native phase/margin cases match every
forward, reverse, required and freshly recomputed reverse level across shared,
reconvergent, output-driven and dangling cones. Refresh atomically recomputes a
complete snapshot after graph edits; clearing and verification return errors
instead of printing. Tests add buffer-precondition rejection, ownership, exact
limits, cancellation, graph growth and a 10,000-input chain. Independent
recurrence fuzzing passed 1,475,518 executions. The full suite passed in 7.010s,
the race suite in 39.477s and vet passed; immutable fixture digests and independent checks pass. `aigTiming.c` is the thirteenth
validated inventory item; 68 files remain partial and 1,676 unstarted.

`aig_.c` is the twelfth validated inventory item. It is an empty translation-
unit placeholder with no declarations, definitions, data or runtime behavior,
so deliberately has no fabricated Go counterpart. Its local source, inventory
hash and pinned-reference source are identical.

The stable-ID-safe portion of AIG object mutation now covers construction,
connection/disconnection, recursive cone deletion, output deletion and driver
patching, plus owned object formatting. Across 32 phase seeds, three native
mutation scenarios per seed match every object slot, fanin, reference count,
level and phase. Go turns native null arena entries into permanent zero-value
tombstones and removes deleted COs from its authoritative output slice instead
of retaining native's stale `vCos` entry. Mutations are clone-and-commit atomic,
iterative, limitable and cancellable; choice members are protected. Tests also
cover descendant replacement, output/register/constraint accounting, exact
formatting and a 10,000-input cone. The semantic mutation fuzzer passed
1,500,177 executions. The full suite passed in 6.742s, the race suite in
35.929s and vet passed; immutable fixture digests and independent checks pass. `aigObj.c` remains partial because native
buffer-fanin repair, buffer propagation and general object replacement still
need a design compatible with stable IDs and strict topological fanins.

The complete AIG operation surface now includes generic AND/XOR dispatch,
lookup-only AND, direct/default MUX, majority, balanced multi-operations, both
miter forms and atomic N-variable AND/OR/XOR constructors. All 32 native phase
sequences match every returned literal, CI ID and complete object record. The
port also preserves the pinned GCC build's observable right-to-left allocation
order inside native unsequenced XOR, MUX and majority call expressions; CNF
fixtures that intentionally sequence their temporary nodes retain that explicit
shape. Native optional dedicated-XOR and additional-strashing modes are
normalized to the port's single lowered, structurally hashed AND
representation. Tests cover lookup nonmutation, caller-slice ownership, atomic
invalid/limit/cancellation failures, exhaustive zero-through-seven-variable
constructors and 10,000-input construction/miter paths. Miter fuzzing passed
1,488,728 executions and mixed-operation fuzzing passed 253,670 executions.
The full suite passed in 6.680s, the race suite in 35.867s and vet passed; immutable fixture digests and independent checks pass.
`aigOper.c` is the eleventh validated inventory item; 68 files remain partial
and 1,678 unstarted.

Concrete owned AIG pattern packs now preserve the native multiply-with-carry
random stream, bit-zero reservation, packed AND/output simulation, rare-node
selection, care-aware conflict checks, first-compatible pattern placement,
skipped/total counters and symmetric signature histograms. RNG state is local
to each pack rather than process-global, and print-only diagnostics become
copied data and statistics. The completed justification experiment connects
rare candidates to the existing eight-heuristic search and returns its exact
candidate order, counters, literal total and final packed state. All 32 native
phase cases match random/base/final patterns, every object signature, rare
candidates, every attempted literal vector, care masks and both sets of
statistics. Tests add duplicate/conflicting assignments, full-pack skips,
stale graphs, ownership, malformed inputs, buffer rejection, resource limits
and cancellation. Packed semantic fuzzing passed 1,486,804 executions and the
justification fuzzer passed 1,640,077 executions. The final full suite passed
in 6.700s, the race suite in 35.193s and vet passed; immutable fixture digests and independent checks pass. `aigPack.c` is the ninth and `aigJust.c` the tenth validated
inventory item; 69 files remain partial and 1,678 unstarted.

Owned reconvergence-driven AIG window cuts now preserve native frontier order,
zero/one/two-new-leaf cost, cost-two fanout filtering, deeper-node tie breaking,
cost-zero early selection and stable remove/append behavior without touching
object marks. All 32 native size, fanout, phase and external-reference cases
match exact final frontiers and visitation arrays. Tests add malformed roots
and limits, copied-result ownership, typed buffer rejection, cancellation and a
10,000-node reconvergent chain. The independent frontier-model fuzzer passed
1,663,339 executions. The full suite passed in 6.524s, the race suite in
35.496s and vet passed; immutable fixture digests and independent checks pass.
`aigWin.c` is the eighth validated inventory item; 70 files remain partial and
1,679 unstarted.

Owned iterative AIG MFFC transactions now cover depth-limited ordered support,
complete membership labels, protected-cut labels, switching-power accumulation
and one-step cut extension. Copied reference arrays preserve every native
boundary without mutating manager counts or traversal state; cut extension
retains maximum-level selection, first-cheapest tie breaking and skip-node
support traversal. Power accumulation exactly reproduces C's double
intermediates and float stores. All 32 native phase/reference cases match exact
sizes, supports, member sets, cut extensions and float values. Tests add CI
roots, native level-zero buffer boundaries, typed reached-buffer rejection,
malformed inputs, exact limits, cancellation, ownership and a 10,000-node
cone. The independent semantic fuzzer passed 770,052 executions. The full
suite passed in 6.651s, the race suite in 36.420s and vet passed; immutable fixture digests and independent checks pass. `aigMffc.c` is the
seventh validated inventory item; 70 files remain partial and 1,680 unstarted.

Concrete owned fixed, flexible and hierarchical AIG byte arenas now replace
native raw-pointer allocation. Fixed blocks preserve chunk sizing, allocation
order, LIFO recycle, first-chunk restart and lifetime peaks. Flexible blocks
preserve zero-size requests, chunk growth and native nominal accounting while
checking real retained capacity after restart, avoiding native's possible
overflow after a large request. Step arenas preserve every power-of-two size
class, oversized retention/no-op recycle and fixed-only usage reporting.
Borrowed blocks become invalid on recycle, restart or stop; foreign, stale,
double and wrong-size operations return typed errors. The permanent oracle
matches all recorded native states and pointer-reuse observations. Fixed-arena
fuzzing passed 707,244 executions and mixed flexible/step fuzzing passed 132,867
executions. The full suite passed in 6.700s, the race suite in 35.538s and vet
passed; immutable fixture digests and independent checks pass.
`aigMem.c` is the sixth validated inventory item; 70 files remain partial and
1,681 unstarted.

Owned iterative AIG justification now preserves the native eight-heuristic
retry sequence, ordered CI literals, conflict behavior, final failed-attempt
literals and the constant-one success quirk. Owned ternary simulation preserves
Kleene values, fanin-zero short-circuiting and last-wins duplicate assignments.
All 32 native reconvergent, phase and CO-root cases match exact object layouts,
success results, literal order and every partial-assignment value. Tests cover
source ownership, malformed inputs, reached versus disconnected buffers, exact
limits, cancellation and a 10,000-node chain. The semantic fuzzer passed
1,703,306 executions against an independent ternary evaluator and checked every
successful nonconstant justification. The full suite passed in 6.721s, the race
suite in 35.209s and vet passed. Immutable fixture digests and independent checks pass.
`aigJust.c` is now partial: `Aig_ManJustExperiment` remains tied to the
unported `aigPack.c` candidate/pattern subsystem. Inventory is therefore 5
validated, 70 partial and 1,682 unstarted.

Native AIG timeframe expansion now covers all initialization, primary-output,
register-output and enlargement modes with an owned source-object-major map.
All 48 native flag/frame combinations match exact CI/CO partitions, register
counts, complete live/tombstoned object arenas and every cleanup-filtered map
entry. Source graphs are immutable and failures are atomic; native unchecked
zero-frame and buffer cases become typed errors. Tests cover map ownership,
malformed access, exact limits, cancellation and a 10,000-node expansion. The
dedicated semantic fuzzer passed 1,283,807 executions by comparing every output
of three independent frames with direct source evaluation. The final full suite
passed in 6.665s, the race suite in 35.364s and vet passed. Immutable fixture digests and independent checks pass. `aigFrames.c` is the fifth validated inventory item; 69 files remain
partial and 1,683 unstarted.

Dynamic AIG topological order now uses an owned concrete circular list with
object zero as sentinel. It preserves native ascending initialization,
insert-before-next behavior, separate advance/iterator-refresh operations,
stable unlinking, both cursor-repair cases and processed/total counts. Every
state and sequence matches 32 native mutation traces across four cursor
positions. Clone ownership, empty/malformed operations, native buffer rejection,
limits, cancellation and 10,000 nodes are tested. The dedicated fuzzer passed
1,317,802 executions while comparing cursor removals and final state against an
independent ordered-slice model. The final full suite passed in 6.597s, the race
suite in 35.097s and vet passed. Immutable fixture digests and independent checks pass. `aigOrder.c` is
the fourth validated inventory item; 69 files remain partial and 1,684
unstarted.

Owned iterative AIG cut-truth evaluation now supports ordered CI or internal
boundaries, repeated leaves, complemented edges/root literals, constants and
arbitrary-width packed truth tables without using native object-data scratch.
All 32 native 2-, 5-, 8- and 10-leaf cuts match complete node postorders and
truth words. The single-node evaluator also matches every complemented-AND form
and the native dedicated-XOR branch. Malformed/uncovered boundaries, buffers,
limits, cancellation and source/result ownership are tested. The dedicated
fuzzer passed 1,151,347 executions, exhaustively comparing all 64 six-input
assignments with direct AIG simulation. The final full suite passed in 6.439s,
the race suite in 32.773s and vet passed. Immutable fixture digests and independent checks pass.
`aigTruth.c` is the third validated inventory item; 69 files remain partial and
1,685 unstarted.

Ordered AIG fanout storage now uses an owned concrete `AigFanoutIndex` rather
than native manager-attached intrusive lists. Construction preserves object and
fanin order; graph-derived and explicit edge add/remove operations preserve
native pin identity, stable removal and tail reinsertion. All 32 native initial,
removed and restored snapshots match exactly. Queries/clones are independently
owned; malformed, duplicate and missing edges return typed errors; native's
no-buffer build precondition, limits, cancellation and a 10,000-edge wide node
are tested. The dedicated fuzzer passed 1,388,527 executions while independently
reconstructing every structural edge and reference count. The final full suite
passed in 6.487s, the race suite in 32.465s and vet passed. Immutable fixture digests and independent checks pass. `aigFanout.c` is the second validated inventory item; 69 files remain
partial and 1,686 unstarted.

Concrete AIG choice chains now participate in manager growth, cloning, cleanup
and invariant checking. Ordinary DFS preserves native equivalence ordering,
including its mixed preorder/postorder rule; `DFSChoices` preserves fanin-zero,
fanin-one, successor postorder. `ComputeChoiceLevels` returns owned no-box
levels instead of overwriting structural levels. All four ordinary traversal
arrays, dedicated choice order, complete levels and maxima match 32 native
phase cases. The native fixed-16 support diagnostic is now a non-printing
shared-budget count; all 32 native counts match and a 21-input chain crosses
the threshold exactly. Tests cover atomic chain replacement, malformed links, structural
dependency cycles, clone/cleanup ownership, limits/cancellation and a
10,001-alternative iterative chain. The fuzzer passed 1,105,954 executions with
exhaustive semantic equivalence checks. The support/DAG fuzzer passed 1,508,488
executions; the final full race suite passed in 32.210s and vet passed. Inventory
at that checkpoint was 70 partial tracked files and 1,687 unstarted.

Timing-box-aware AIG topology verification and choice leveling now complete
the remaining behavior in `aigDfs.c`. The Go implementation validates exact
CI/CO box ownership, requires every box-input CO to precede its output CIs,
and computes each box-output CI one level beyond the maximum of all box inputs.
It uses owned iterative traversal, preserves graph/timing state, rejects
structural/choice/box dependency cycles and strengthens native assertion cases
into typed errors. All 32 native phase/order cases match complete object levels,
maximum levels and topology answers. Tests cover malformed and overlapping box
metadata, dimension errors, exact limits, cancellation, no-box compatibility
and a 10,000-box chain. The dedicated fuzzer passed 1,198,000 executions while
checking independent topology and level recurrences. The final full suite passed
in 6.464s, the race suite in 32.212s and vet passed. Immutable fixture digests and independent checks pass.
`aigDfs.c` is the first validated inventory item; 69 files remain partial and
1,687 unstarted.

Reverse fanout DFS is now ported with an owned fanout index constructed in
native object-insertion order. It excludes COs, includes every AND/buffer and
returns fanout-first postorder across shared, output-connected and dangling
logic without installing persistent fanout storage or traversal marks. All 32
native orders match. `ConeANDMask` replaces native mark/count/clean/unmark
scratch side effects with an owned Boolean set; all 32 native cone sets and
counts match, and Go has no generic object data requiring a clean pass. Tests
cover explicit buffers, exact object/event limits,
cancellation, source ownership and a 10,000-node iterative chain. The dedicated
fuzzer passed 1,377,241 executions while proving uniqueness, completeness and
fanout-before-fanin order. The refactored cone/DAG path passed another 1,607,499
DFS fuzz executions; the final full race suite passed in 32.222s and vet passed.
At that checkpoint the inventory was 70 partial tracked files and 1,687
unstarted; the timing-box slice described above subsequently validated
`aigDfs.c`.

Native AIG cone transfer and single-CI functional composition are now ported.
Transfer preserves indexed source-to-destination CI mapping, complement phases,
native structural-hash object order, constant and same-manager shortcuts.
Composition preserves native substitution semantics while also safely accepting
an internal replacement function that depends on the replaced CI. Both use
bounded iterative traversal and construct on a clone, committing only after
success so invalid input, limits and cancellation cannot partially mutate a
caller-owned manager. Exact returned literals and complete object layouts match
32 native cases. Tests add exhaustive six-input semantic projection, missing
destination inputs, buffers and malformed literals, atomic errors, an arbitrary
internal replacement and a 10,000-node cone. The dedicated fuzzer passed
1,518,950 executions; the full race suite passed in 32.242s and vet passed.
At that checkpoint the inventory was 70 partial tracked files and 1,687
unstarted; the timing-box slice described above subsequently validated
`aigDfs.c`.

Core AIG DFS, support, cone-size and level traversal is now ported with owned
iterative state. All ordered results match 24 native graphs for node-only/full
preorder and postorder, all-object DFS, explicit-root array/node traversal,
ordered single/multi-root CI support, DAG size, boundary-cut internal postorder,
maximum output level and every level bucket. Topological verification uses the stronger existing Go manager
check. Native assertion cases such as dangling internal objects return typed
errors; buffers, invalid roots, exact object/event caps, cancellation, source
immutability and 10,000-node chains are tested. The final dedicated fuzzer passed
1,644,716 executions with cut-volume consistency; the full race suite passed in
33.606s and vet passed.
At that checkpoint the inventory was 70 partial tracked files and 1,687
unstarted; the timing-box slice described above subsequently validated
`aigDfs.c`.

DAR-best-cut AIG mapping scan with structural-supergate substitution is now
ported. All 32 native area, preorder-root and complete reference arrays match
across root/nested supergates, repeated cut leaves, constants, complemented and
shared outputs. Signed supergate collection preserves boundary order,
same-polarity deduplication and opposite-polarity contradiction collapse while
using the dynamically accumulated mapping references. Results are owned and do
not alter graph marks or structural references. Invalid cuts/marks, exact work
caps, cancellation, a contradictory cone and a 10,000-node iterative mapping
are tested. The dedicated fuzzer passed 1,626,780 executions; the full race
suite passed in 33.439s and vet passed. At that checkpoint the inventory was 70
partial tracked files and 1,687 unstarted. The later traversal slices above
validated the broader AIG DFS surface; CNF-cut supergate mode remains unfinished.

Native CNF diagnostic printing and graph-order SAT-variable collection are now
ported. All four diagnostic byte streams match, including header-only formulas,
empty clauses, omitted DIMACS terminators and readable `-0`. Output is owned and
byte/work bounded with cancellation. CI/PI/CO collection returns owned arrays,
keeps native `-1` unassigned entries and treats the legacy PI spelling as all
CIs, including latch outputs. Graph/map validation, object caps, cancellation,
wrapper coverage and alias checks pass. The formatter fuzzer completed 198,158
executions; the full race suite passed in 34.015s and vet passed. Inventory
remains 69 partial tracked files and 1,688 unstarted; CNF manager lifecycle is
represented by owned Go values, while proof-solver wrappers and other ABC
subsystems remain unfinished.

Complete fast-CNF clause derivation is now connected to structural recognition,
marking, cone collection and truth derivation. All 96 native clause streams,
variable counts and complete object-variable maps match across asserted outputs,
all exposed outputs and sequential register-suffix exposure. The encoder
preserves native reverse-object variable assignment, signed AND-supergate
flattening, ordered direct six-input ISOP covers, output/equivalence clauses and
the final constant unit. Exhaustive six-CI SAT projection proves every fixture
against its original AIG; a 1,000-input supergate checks iterative operation.
It does not mutate the graph, returns owned results and atomically enforces
object, visit, variable, clause, literal and ISOP limits plus cancellation.
Unsupported explicit buffers/XOR objects, invalid marking diagnostics and
larger complex cones return errors. `CountAigFastCNFNodes` also matches all 32
native marked-object/node-clause diagnostics. Its deliberately non-total native
semantics are documented and regression-tested: contradictory signed
supergates can emit more clauses than their regularized truth-count estimate.
The final dedicated fuzzer passed 530,179 derivation/count executions, the full
race suite passed in 33.438s, and vet passed. Inventory remains 69 partial
tracked files and 1,688 unstarted. Large-design/performance validation and the
broader ABC port remain unfinished.

Fast-CNF cone prerequisites now preserve native regular/signed boundary DFS
order, unique internal postorder and six-input truth derivation. All 24 native
cone cases match every ordered field and truth word; independent exhaustive
six-CI simulation also passes. Source graph marks/scratch fields remain
unchanged. Invalid boundaries/dependencies, limits/cancellation and a 10,000-
node chain are checked. 146,012 cone fuzz executions, full race tests (31.363s)
and vet pass. Inventory is now 68 partial tracked files and 1,689 unstarted.
Fast marking, mux recognition, clause generation and complete fast derivation
remain unfinished.

Typed CNF formula/text/file solving now connects parsing, solver loading,
search and verified model extraction. All 33 native default outcomes match;
native and Go SAT models independently satisfy their formulas. Native result
codes and owned final-variable suffixes are preserved. Every returned Go SAT
model is checked; cancellation and logical work limits return UNKNOWN without
partial models. Tests cover suffix ownership, file/parse errors, zero-variable
formulas and load/search cancellation. 150,821 exhaustive-reference fuzz cases,
full race tests (31.174s) and vet pass. Inventory remains 67 partial tracked
files and 1,690 unstarted. Native learned-database tuning, CLI diagnostics and
other solver/CNF variants remain unfinished.

Checked plain CNF reading is now available with strict standard parsing and
explicit NativeLines compatibility. All ten native line-mode results match,
including ignored trailing tokens and actual/declaration count mismatches.
Strict mode supports spanning/multiple clauses and zero counts while rejecting
malformed headers/numbers, missing zeros and wrong counts. Byte/line/variable/
clause/literal caps, bounded temporary-file reads, OS errors and cancellation
are checked. 11,468 arbitrary-input fuzz executions, full race tests (30.952s)
and vet pass. Successful fuzz results validate and roundtrip through standard
serialization. Inventory remains 67 partial tracked files and 1,690 unstarted;
file-solving wrappers and other CNF utilities remain unfinished.

Object-owned CNF clause polarity metadata now matches all 48 native byte
arrays. Masks aggregate negative/positive tail occurrences separately by owner
and head polarity, packing the first four tail positions. Independent mixed-
sign/wider-clause tests, invalid ownership bounds and cancellation are checked.
158,150 arbitrary-clause fuzz executions, full race tests (31.137s) and vet
pass. Inventory remains 67 partial tracked files and 1,690 unstarted. Other
CNF utilities and the broader ABC port remain unfinished.

Compressed CNF output and file-writing wrappers are now implemented. All ten
native gzip fixtures decompress to identical text; repeated Go compression is
deterministic without requiring native zlib byte identity. Independent text/
compressed caps, compression-stage cancellation and formula/mapped/frame wrappers
are checked. Temporary-directory tests cover plain/gzip output, short paths,
case-sensitive suffixes, unchanged destinations on serialization failures and
filesystem errors. 29,232 gzip fuzz executions, the final full race suite
(31.266s) and vet pass. Inventory remains 67 partial tracked files and 1,690
unstarted. Other CNF manager operations and alternate solver variants remain
unfinished.

Bounded DIMACS text serialization now preserves native standard/raw-ID
numbering, quantifier order, spaces, terminators and final blank line. All ten
native byte fixtures match, including zero-variable formulas and empty clauses.
Explicit empty prefixes, invalid IDs, exact byte caps, cancellation and mapped/
frame wrappers are checked. Independent parse/reconstruction passes 142,212 fuzz
executions; full race tests (30.800s) and vet pass. Output is owned bytes; raw-ID
mode retains native variable-zero diagnostic ambiguity. Inventory remains 67
partial tracked files and 1,690 unstarted. Compressed output and filesystem
handling remain unfinished.

Graph-free owned CNF now supports literal-position collection and selected
flip/lift transforms. All 84 native cases match, including absent variables,
repeated/nonmonotonic positions, mixed signs and empty clause boundaries.
Duplicate positions apply sequentially. Go expands the resulting variable span
as needed, while native nVars remains unchanged. Constructor/output ownership,
invalid indices/literals, work/variable limits, mid-operation cancellation and
zero-variable empty/contradictory formula loading are checked. 160,703 transform
fuzz executions, full race tests (30.400s), final focused race checks and vet
pass. Inventory remains 67 partial tracked files and 1,690 unstarted; CNF file
output, alternate solver wrappers and other utilities remain unfinished.

Native output OR/AND constraints and the four-clause XOR relation are now
available as owned clause builders and concrete solver additions. All 36
supported native clause streams match, including all 27 XOR alias triples.
Independent exhaustive solver/truth checks cover output counts zero..four,
phase-transformed signals and base offsets. Empty OR deliberately reports
inconsistency rather than native assertion failure; empty AND succeeds without
clauses. Unassigned COs and invalid XOR variables are rejected. 128,238 XOR
fuzz executions, full race tests (30.784s) and vet pass. The initial test-only
solver-capacity error was corrected before the final full race run. Inventory
remains 67 partial tracked files and 1,690
unstarted; proof-enabled solver2 wrappers and other utilities remain unfinished.

Multiframe mapped CNF is now available with native-order latch links and
optional all-zero initial registers. All 96 native raw clause streams match
across six transition functions, one/two registers, one..four frames and both
initialization modes. Independent exhaustive input-sequence/initial-state
checks verify trajectories and reject corrupted linked state assignments.
Base offsets, phase-transformed relations, ownership, missing LI mappings,
limits and cancellation are checked. 38,453 frame fuzz executions, full race
tests (30.992s) and vet pass. Source CNF remains unchanged; solver ownership
stays with the caller. Inventory remains 67 partial tracked files and 1,690
unstarted. Alternate loader variants and other CNF utilities remain unfinished.

Mapped CNF now supports owned cloning, variable-base shifting, phase polarity
transformation and incremental solver loading. All 48 native polarity/lift cases
match clauses and object-variable numbers. Shifted copies deliberately expand
the variable span (native lift leaves nVars unchanged); transformed signal
literals preserve original AIG meanings. Exhaustive SAT projection, phase
involution, ownership, bounds/cancellation, two independent solver copies and
object-ID variable-zero constants are checked. 35,736 transform fuzz executions,
full race tests (30.803s) and vet pass. Inventory is 67 partial tracked files
and 1,690 unstarted. Native multiframe/latch initialization and other CNF manager
utilities remain unfinished.

CNF postprocessing now performs the native one-pass recovery selection and
mapping-reference updates using corrected composition. All 48 corrected-native
mappings match surviving cuts, truth, area, absorption counts and complete
reference arrays. Exhaustive SAT projection verifies the original AIG relation;
automatic DAR/area-flow/recovery/emission integration also passes semantic checks.
Stable cost ordering, at most one absorption per root, nonincreasing area,
ownership, proposal-limit skips, global limits and cancellation are tested.
44,200 recovery fuzz executions, full race tests (30.732s) and vet pass.
`DeriveAigPostprocessedCNF4` enables recovery explicitly; default derivation
retains its native initial-mapping behavior. Inventory is 66 partial tracked
files and 1,691 unstarted. Supergate marks, other derivation variants and
larger-design validation remain unfinished elsewhere in the CNF subsystem; the
complete `cnfPost.c` source is now explicitly validated.

Cut composition for mapping postprocessing is now implemented with correct
Boolean substitution and owned storage through sixteen inputs. All 240 cases
match the corrected frozen dataset and independent exhaustive evaluation.
The original native function has verified truth-pointer, support-movement and
packed-cofactor defects; its first words are retained as regressions (237 differ),
not adopted as Go semantics. Default nine-input merge caps, exact small-SOP and
wide direct-ISOP costs, 127 cost saturation, malformed shapes, ownership, work/
ISOP limits and cancellation are tested. A sixteen-input AND substitution is
proved exhaustively. 76,934 composition fuzz executions, full race tests
(30.539s), final focused race checks and vet pass. Inventory is 65 partial
tracked files and 1,692 unstarted. Postprocessing's selection/reference-update
pass remains unfinished and is not enabled in automatic derivation.

Object-ID CNF emission and alternate automatic derivation are now ported.
SAT variable IDs equal object IDs, including constant one at variable zero;
all COs receive equivalences. Mapped clauses precede CO pairs and the final
constant unit, with owned per-object clause starts/counts and preorder mapping.
All 48 native automatic cases match every clause, clause range and mapping
entry across both support-minimization modes. Exhaustive original-AIG SAT
projection passes. Supplied wide cuts also pass projection through sixteen
inputs (exhaustive through eight, sampled above eight); native wide object-ID
clause-order comparisons remain pending. Sequential output exposure, empty
constant-only graphs, metadata ownership, bounds and cancellation are tested.
41,813 alternate-encoder fuzz executions, full race tests (29.999s) and vet pass.
Inventory remains 64 partial tracked files and 1,693 unstarted. Other derivation
variants, supergate handling, postprocessing and larger-design validation remain.

Packed CNF mapping serialization now matches the native signed-int format:
object-indexed offsets, repeated 32-bit truth bits and four leaf slots padded
with -1. Automatic mapping-array derivation uses native postorder, separate
from preorder CNF emission. All 24 complete native mapping arrays match.
Explicit cut order, constants, padding, malformed/duplicate roots, allocation
limits, cancellation and owned output are checked. 130,155 fuzz executions,
full race tests (28.949s) and vet pass. Coverage stays 64 partial tracked files
and 1,693 unstarted; this extends already partial CNF core/writer files.

Automatic four-input optimized CNF is now connected through DAR cut
enumeration, area-flow selection, reachable preorder scanning and emission.
All 192 native DAR candidate lists match across 24 six-input graphs, four slot
capacities and both support-minimization settings. Independent exhaustive
simulation verifies every generated cut against its AIG. The 24 default
end-to-end cases match every native CNF clause and object-variable number;
exhaustive SAT projection checks their original two-output AIG relations.
Cleanup preserves IDs on an owned clone. Fixed-slot eviction, dominance,
truth alignment, support minimization, limits and cancellation are covered.
146,256 cut fuzz executions and 46,240 complete-encoder fuzz executions pass.
Full race tests (28.787s), focused final race checks and vet pass. Coverage is
64 partial tracked files and 1,693 unstarted. Explicit buffers/XOR, supergate
marks, alternate derivation variants, postprocessing and
large-design validation remain unfinished; this does not complete ABC's CNF
subsystem or the full ABC port.

The four-input supplied-cut mapping scan now preserves native preorder and
postorder root order, unique reachable cut area and all mapping-edge reference
counts. All 64 native cases match, including constants, shared roots, duplicate
leaves and unreachable cuts. Their emitted CNF passes exhaustive SAT projection.
The supplied-candidate wrapper now connects area-flow selection, preorder scan
and native clause emission without changing structural graph references. Deep
10,000-cut traversal, missing/duplicate cuts, ownership, limits and cancellation
are tested. Full race tests pass (28.792s), as do 116,471 scan fuzz executions.
Coverage is 61 partial source files and 1,696 unstarted. Automatic DAR cut
generation, supergate-mark handling, mapping serialization and optimized
postprocessing remain unfinished.

The initial CNF area-flow mapping pass now selects among supplied nontrivial
four-input cuts in topological order. Native SOP costs/reference sums match for
all 65,536 truth tables. Recursive flows, native reference-sum tie breaks,
owned output, malformed inputs, limits and cancellation have regression checks;
113,663 fuzz executions pass. Full race tests (28.810s) and vet pass, with an
additional focused race run after the final signed-division adjustment. Flow
storage follows native uint32 signatures with signed int32 leaf division.
Coverage is now 60 partial source files and 1,697 unstarted. DAR cut generation,
cut transfer, output-cone scanning and mapping postprocessing remain unfinished.

The supplied-cut CNF writer now supports up to sixteen inputs. Its new Kit
ISOP generator matches 196 native cases in cube order, phase and literal count;
independent exhaustive evaluation verifies these covers through sixteen inputs.
The wider writer matches all clauses and variable numbers in 156 native cases
for 5, 6, 8, 12 and 16 inputs. Mixed small/wide mappings pass SAT projection
checks (exhaustive through eight inputs; deterministic samples above eight).
Resource-limit fallback now recognizes wrapped errors and retains a valid direct
cover when complement construction exceeds its budget; cancellation still
returns no result. The full race suite (28.264s), vet, 106,408 ISOP fuzz executions
and 20,876 wider-writer fuzz executions pass. Source coverage is 59 partial and
1,698 unstarted. Automatic cut derivation and area mapping remain outstanding;
this is still a writer for supplied cuts, not a complete optimized encoder.

The optimized four-input mapped CNF writer is now ported for supplied cuts.
All 72 native cases match complete raw clauses and object-variable numbering,
including mapped-order permutations, constants, output phases, asserted/all-
exposed outputs and sequential latch-input exposure. Exhaustive SAT projection
and direct AIG relation tests pass. The full race suite, vet and 51,862 fuzz
executions pass. Bounds, cancellation and owned metadata are checked. Automatic DAR cut generation/area mapping remains unfinished, so this
is not yet a complete graph-to-optimized-CNF encoder.

Optimized CNF work now includes ABC's exact four-input SOP database and
native cube/literal-count primitives. Every one of 65,536 native covers/counts
matches, and independent exhaustive evaluation proves every represented truth
table. All 153 native cube cases match order/counts; ownership and malformed
input checks pass. The full race suite, vet and 295,524 combined fuzz executions
pass. Coverage is now 58 partial source files and 1,699 unstarted. The complete
optimized CNF writer, cut derivation and mapping remain unfinished.

SAT-LUT option audit added native area-only (`fDelay=false`) behavior across
explicit, whole-network, boxed and packed entry points. Timing exclusions and
application delay conditions are disabled while function/closure checks remain.
Whole edge passes preserve prior rows across improvements and report whether
proposals were timing constrained. All 172 native/corrected area-only cases
match; cases with real unit-delay regressions verify the intended tradeoff.
The full race suite, vet and 24,967 combined fuzz executions pass. The next
roadmap focus is optimized SAT/CNF support; full ABC parity remains unfinished.

Unit remapping now supports native five/six-input feasibility through explicit,
whole-network, boxed and packed paths. `LutSize` preserves native buckets:
requests 1..4 permit four inputs; 5..6 permit six. Edge search retains a checked
four-input limit. All 64 native wider passes match final counts; native truth
words and independent output compositions agree. Boxed packed integration
produces genuine six-input LUTs. Slot-five/six duplicate semantics, input
ownership and infeasible narrower timing cases are checked. The full race
suite, vet and 13,165 fuzz executions pass. Remaining source/API/diagnostic
parity and broader design validation remain unfinished; full ABC parity is
not claimed.

Whole-network boxed unit and edge scheduling is now connected, with packed
entry points and accepted private timing-state carryover. All 36 native/
corrected-reference passes match complete mappings, LUT counts, maxima and
terminal arrival bits across three ten-gate stages. Unlimited passes reduce
30 LUTs to three; capped passes stop at 21 or 12. Output composition, real
session reuse, automatic prior-edge baselines, ownership, atomic failures and
no-box fallback pass. The full race suite, vet and 4,568 fuzz executions pass.
Wider unit LUT search, remaining source/API parity and broader design validation
remain unfinished; the full ABC port is still far from complete.

Explicit boxed unit-window feedback, SAT search and checked replacement are
now connected. All 96 corrected C-reference cases match decisions, candidate
arrivals, original required arrays, clauses and terminal arrivals. Original
native timing bugs remain pinned separately. Application independently checks
full staged global timing; a regression case verifies late proposals cannot
leak through local fixed-CI evaluation. Output composition, source immutability,
cancellation and no-box fallback pass. The full race suite, vet and 30,068 fuzz
executions pass. Whole-network boxed scheduling/baseline propagation remains
unfinished.

Explicit boxed edge-window evaluation, SAT search and checked replacement are
now connected. Candidate timing propagates through staged boxes on an owned
manager clone; all 96 native cases match acceptance, clauses, complete edge/
delay arrays and terminal arrival bits. Accepted replacements preserve graph
outputs, and no-box fallback/ownership/cancellation checks pass. The full race
suite, vet and 29,797 fuzz executions pass. Boxed unit-window feedback and
whole-network boxed baseline handling/scheduling remain unfinished.

Whole-network edge-delay remapping now shares native scheduling and reusable
sessions with the unit pass. Network and packed entry points derive automatic
limits from supplied prior rows, validate accepted replacements and return
owned assignments. All 72 native pass cases match final LUT counts and effective
limits; native/Go mappings preserve outputs. Twelve native no-improvement cases
leak trial rows; regression tests require Go to preserve unchanged prior rows.
The full race suite, vet and 11,512 fuzz executions pass. Boxed remapping and
native arena efficiency remain unfinished.

Explicit edge-delay remapping now has native full-window feedback, shared
SAT search and checked owned replacement. `EvaluateEdgeWindow` matches all 96
native cases for acceptance, clauses, complete delays and edge slots.
`SearchEdgeWindow` retries late candidates with native all-nonempty-node
exclusions; `RemapEdgeWindow` validates functions/closure and applies only the
edge limit. Integration exercises rejection, one-LUT delay-ten acceptance and
zero-limit exhaustion. The full race suite, vet and 35,007 fuzz executions pass.
Whole-network edge scheduling, prior-edge automatic baseline handling, packed
edge outputs and boxed search remain unfinished.

Whole-network unit passes now reuse persistent discovery state as well as
SAT cardinality prefixes. Static fanouts/terminal masks build once, mapped
reference prefixes refresh after replacements, and transient scratch is reset
between pivots. All 480 native discovery vectors match through reused sessions.
Mid-query cancellation recovery, rejected invalid refresh and same-size graph
mutation rejection pass. Whole-pass counters verify actual reuse/refreshes.
The full race suite, vet and 51,159 fuzz executions pass. Native arena efficiency
and boxed/edge remapping remain unfinished.

Whole-network unit passes now reuse one cardinality solver prefix across
active windows, with fixed capacity from the configured window limit. Actual
build/reset counters verify reuse; phase hints and structural/learned/feedback
clauses reset between windows. All 12 native two-cone cases match final counts,
and native/Go mappings reproduce outputs over all eight-input assignments.
An explicit UNSAT-exclusion reset test verifies isolation and failed-reset
atomicity. Vet, the full race suite and 5,844 fuzz executions pass. Persistent discovery scratch, native arena efficiency and boxed/edge
scheduling remain unfinished.

`SatSolver.Bookmark/Rollback` now restore an owned logical prefix, including
clauses, learned state, variables, watches, units and assignment/order metadata.
All 16 native scenarios match restored counts and query statuses. Repeated
restore, contradictory/learned tails, callback guards and Reset are checked.
Window solvers now bookmark cardinality before cut allocation; `ResetWindow`
reuses that prefix across window sizes and preserves the current window on
preparation failure. Vet, the full race suite and 125,026 fuzz executions pass. Snapshots restore heuristic state rather than reproducing
native arena retention. Whole-pass reuse wiring, persistent discovery scratch
and boxed/edge scheduling remain unfinished.

`GiaLUTMapping.AlternateNetwork` and `GiaLUTNetwork.PackedMapping` now
provide native general-width conversion and canonical repacking. All 24 native
cases match complete arrays through width six, including dropped empty/constant
records and footer replacement. The packed `RemapUnitPass` wrapper connects
conversion to checked whole-network unit remapping and returns owned packed
output with an explicit storage cap. Ownership, idempotence and atomic failure
checks pass. The full race suite, vet and 156,946 fuzz executions pass. Native
persistent manager/cardinality-bookmark reuse and boxed/edge modes remain
unfinished.

`GiaLUTNetwork.RemapUnitPass` now schedules whole-network unit remapping:
ascending current mapped pivots, persistent repeated-window suppression,
native small-window skips, checked replacements and improvement limits.
All 36 full native pass cases match final LUT counts; native and Go mappings
independently reproduce all graph outputs. Source immutability, global delay,
pivot ordering and atomic budget/cancellation failures are checked. Vet and
21,423 fuzz executions pass, along with the full race suite. Native
manager/cardinality-bookmark reuse, packed conversion lifecycle and boxed/edge
modes remain unfinished.

`GiaLUTNetwork.DiscoverWindow` now grows native pivot windows with mapped
reference accounting, structural fanout traversal, shared-input candidate
ranking, reverse-growth preference and cone-size limits. All 480 native cases
match complete sorted vectors and oversized-pivot rejection. Discovered
windows feed checked unit remapping and exhaustive replacement-output checks.
Vet, the full race suite and 37,101 fuzz executions pass.
Coverage is now 56 partial source files and 1,701 unstarted. Persistent manager
reuse, native conversion lifecycle, pivot scheduling and boxed/edge remapping
remain unfinished.

`GiaLUTNetwork.RemapUnitWindow` now connects explicit-window SAT search to
checked, owned replacement construction. `ValidateFunctions` verifies complete
mapped dependencies, output coverage and every structural LUT boundary, and
derives truth words including buffer/XOR/mux semantics. Application checks
global unit delay, rebuilds fanouts and leaves graph/source mapping unchanged.
All 24 native UpdateMapping cases match fanins, references and truth words.
Exhaustive composed-output checks, 35,470 remapping fuzz executions and 153,144
mixed-gate function fuzz executions pass. The full race suite and vet pass.
Automatic window discovery, native scheduling/lifecycle, boxed/edge search
and complete SAT-LUT feature parity remain unfinished.

`GiaLUTNetwork.SearchUnitWindow` now joins generated cuts, mapping roots/
preferences, concrete SAT, decoding and unit timing feedback. Late candidates
add native node-path clauses and retry the same bound; accepted candidates
save a proposal and tighten the bound. The default native ten-iteration limit
is exposed. All 32 native timing fixtures replay through the actual feedback
loop; generated-window tests find a one-LUT proposal and preserve source state.
Vet, the full race suite and 46,123 fuzz executions pass.
The native timing exclusions are heuristic, so exhaustion after exclusions is
not an optimality proof. Window discovery, boxed/edge paths, Boolean/closure
validation and mapping application remain unfinished.

`GiaLUTNetwork.WindowMappingState` now derives native root variables,
initial mapped nodes and current-cut phase preferences. All 18 full native
ComputeCuts references match, including missing current cuts, skipped external
fanins, CI roots and duplicate mapped roots. Preferences apply to the concrete
solver with reset semantics and atomic validation. Generated windows solve
with those hints. The full race suite, vet and 89,764 fuzz executions pass.
Timing feedback, mapping closure/equivalence validation, graph replacement
and complete remapping remain unfinished.

`GiaManager.GenerateWindowCuts` now derives native ordered structural
candidates from explicit graph windows. All 24 immutable fixtures match complete
leaf/unit records, masks and ranges through 128 leaf/node ordinals. Generated
candidates feed the concrete window solver. Fuzzing checks proper-cut graph
coverage; 94,732 executions pass. The full race suite and vet pass. Current
mapping cut lookup/polarity, root derivation, timing feedback and complete
remapping remain unfinished.

`GiaWindowSolver` now combines native cardinality and cut-selection clauses
in an owned concrete incremental solver. All 78 native queries across 18
scenarios match ordered assumptions and SAT/UNSAT results. Repeated-bound
checks include SAT recovery after UNSAT; constructor budgets cover both CNF
components. Cancellation and caller-input ownership are checked. Vet and
75,084 fuzz executions pass, along with the full race suite.
Candidate generation, polarity hints, timing feedback, graph updates and
the complete remapping loop remain unfinished.

`BuildGiaCardinalityCNF` now provides the native pairwise half-sorter network.
All eight native networks (1..128 inputs) match variable counts, output indices
and raw clauses. Exhaustive solver checks through eight inputs verify every
assignment/threshold; 68,850 fuzz executions through 16 inputs pass. Vet
and the full race suite pass. Root/padding assumptions,
solver lifecycle and the complete SAT remapping loop remain unfinished.

`BuildGiaWindowCNF` now provides native structural selection clauses in
emission order, excluding final unit cuts and preserving unused variables.
All 24 native reference cases match. Exhaustive small-assignment checks and
a concrete solver test verify the implications and permitted multiple-cut
selection. The race suite, vet and 135,144 fuzz executions pass. Fixture
provenance and partial source coverage are recorded. Polarity hints,
root/cardinality constraints, candidate generation and the complete SAT
remapping loop remain unfinished.

`DecodeGiaWindowCuts` now connects selected cut bitmasks to native ordered
window lists. All 144 native decodings match, and decoder-to-timing integration
returns the expected conflict clause. Cut generation/CNF/search remain.

`EvaluateUnitWindow` now checks decoded unit-window timing and returns native
critical-path selection clauses. All 32 native cases match decisions, arrays
and clauses; a solver integration test verifies clause behavior. Cut decoding
and the complete SAT remapping search remain unfinished.

`UnitTimingWithBoxes` now integrates boxed unit arrivals and reverse required
times with corrected traversal indexing. All 48 corrected C-reference cases
match; original native required arrays remain pinned as regressions.
SAT remapping/window solving remain unfinished.

`GiaLUTNetwork.UnitTiming` now computes native unboxed unit arrivals and
reverse required times for SAT-LUT work. All 32 native cases match complete
arrays. Boxed SAT-LUT timing and remapping remain unfinished.

Alternate LUT networks now support boxed timing evaluation and selection.
Independent native mapping2-only oracles pass 48 evaluations and 32 selections.
Incremental proposal integration and reverse graph timing remain unfinished.

`SelectEdgesWithBoxes` now connects staged timing boxes and native greedy
edge selection. All 32 defined native cases match rows, timing and terminal
arrival bits. Mapping2 boxed integration and broader timing work remain.

`EvaluateDelayWithBoxes` now connects staged GIA traversal, mapped edge
delays and timing-box arrival propagation on an owned timing-manager clone.
All 48 native cases match. Boxed selection and mapping2 integration remain.

`GiaManager.OrderWithBoxes` now provides normalized staged traversal for
timing boxes. Sixteen defined native calls match exact orders/rejections,
with eight additional safe rejections for a native diagnostic crash path.
Graph timing evaluation integration remains unfinished.

`TimManager.CORequired` now propagates boxed required times with a corrected
disabled-arc guard. All 128 corrected reference cases pass; 11 of 80 defined
original C cases expose the pinned guard bug. Graph integration remains.

`TimManager` now owns sparse delay tables and computes boxed CI arrivals,
including native float-to-int truncation and cache epochs. All 96 native
arrival cases match. Required-time propagation and graph integration remain.

`TimManager` now owns box records and terminal associations, with native
primary/black counts, unboxed initialization and box traversal labels. All
72 native box states match; delay-table propagation remains unfinished.

`TimManager` now provides native no-box terminal timing/traversal foundations.
All 128 native states match exact float bits and terminal metadata. Box
creation, delay tables/propagation and graph integration remain unfinished.

`GiaLUTMapping.RecomputeEdges` now ports native incremental packed-mapping
recomputation, with explicit affected order/maximum and retained prior state.
All 24 native cases match. Mapping2 incremental proposal integration and
timing boxes remain unfinished.

`GiaEdgeAssignment.WithEdits` now provides atomic owned directed insert/
remove/clear sequences. All 156 native slot cases match; incremental window
evaluation and timing-box integration remain unfinished.

`GiaLUTNetwork.WithWindow/EvaluateWindow` now provide owned proposals and
native full-window edge/timing evaluation. All 12 native cases match.
Incremental window evaluation and timing-box integration remain unfinished.

`GiaLUTMapping.OptimizeEdges` now implements native critical-node improvement
and its assignment loop, including strict-delay backup restoration. All 24
native cases match restored rows and both timing arrays. Mutable window
update/evaluation and timing-box integration remain unfinished.

`GiaLUTNetwork` now provides immutable alternate mapping lists, unit timing
in both directions and critical-level collection. All 12 native cases match
fanouts, timing arrays/maxima and critical groups. Mutable window improvement
remains unfinished.

`GiaLUTMapping.SelectEdgesReverse` now implements native reverse fast-edge
selection with unit distances and maximum over CIs. All 12 native cases
match; window optimization, timing boxes and mapping2 remain unfinished.

`GiaLUTMapping.SelectEdges` now computes native greedy fast edges and delays
for plain mappings without timing boxes. All 24 native cases match complete
rows and timing arrays; advanced/box/mapping2 paths remain unfinished.

`GiaEdgeAssignment.EvaluateDelay` now computes native plain-mapping/no-box
delays and edge counts. All 12 native cases match complete arrays and output
maxima; box timing, mapping2 and selection remain unfinished.

`GiaEdgeAssignment` now owns native two-slot edges and converts pair lists
and LUT packing groups, preserving native overflow and slot-order export.
All 32 native cases match; edge delay/optimization remain unfinished.

`GiaLUTPacking` now owns checked packing groups, exposes native statistics
and transfers group IDs through explicit correspondence. All 32 native
transfers match; packing selection/conversion and timing transfer remain.

`GiaLUTMapping.Transfer` now remaps native packed LUT records using an
explicit object correspondence; all 32 native cases match complete packed
output. Packing/timing transfer and mapping selection remain unfinished.

`GiaLUTMapping.Verify` now checks native mapped dependency connectivity from
CO and buffer drivers, returning a Boolean without printing diagnostics. All
128 native configurations match; selection and optimization remain unfinished.

`GiaLUTMapping.Analyze` now computes native mapped references, per-LUT cone
sizes and signed overlap with bounded iterative traversal. All 30 native
analyses match both frozen reference-count and overlap implementations.

`GiaLUTMapping` now provides owned native packed records, checked import,
ordered boundary lookup, statistics and levels. All 20 native cases match
packed data, levels and LUT truth. Mapping selection remains unported.

`BoundaryTruth6` now evaluates ordered six-input structural boundaries with
native last-duplicate-slot behavior. LUT boundary evaluation now also connects to owned packed mapping records;
mapping selection remains unported.

`CITruth6` now collects sorted CI-object support and computes a six-input
packed truth word, with explicit oversized-support reporting. All 144 native
cases match; constants and extended-node/CO roots are supported safely.

The latest completed slice is the fixed-graph cut-store lifecycle and batch GIA cut enumeration in
`aig_gia_cut_enumerate.go`: sorted cut merging, dominance filtering, ranking,
truth interning and optional support reduction. All 400 native configurations
match cut sets, function IDs, truth records and statistics. Live-graph incremental updates, allocation reuse and LUT/optimized-CNF integration remain outstanding.

Packed bit-vector foundations now include an owned, bounded `BitVector` with
C-compatible rounded start/fill sizes and safe exact-bit growth. Queue
foundations now include an owned float32 max-heap with C-compatible ties and
owned float32 vectors with explicit IEEE behavior. Paged uint64 record storage
and stable insertion-ID interning now include native-compatible truth-table
seed and mux constructors, plus packed Boolean/cofactor/support and variable
transformation/quantification primitives, support compaction and checked
variable embedding, literal-bound simplification and single-variable output
decomposition, pair-input DSD and checked mux-input DSD recognition for later
truth-table algorithms. Phase-only and population-signature semicanonicalization now match native
passes, and the default heuristic NPN/permutation-only refinement pipeline
now matches native results. Optimized adjacent refinement and pair-cofactor
comparisons now match native raw/guarded candidates through twelve variables.
Individual/group symmetry predicates now match native direct/phased flags.
Initial adaptive tied-group phase normalization and stable signature partitioning
now match native results, with the input permutation explicitly deferred.
Adaptive symmetry reduction now preserves native representative chains, phase
weights, unused-input order and repeated equal-size group merges.
The complete adaptive simple-search branch now matches native words and
configurations, including group permutations and individual phase sweeps.
Full algorithm-0 group enumeration now matches native permutation/Gray-phase
traversal, strict-minimum tie behavior and carried phase state.
Shifted-cofactor coefficients and applied group signatures now match native
results, with uint64 accumulation beyond C signed-int limits.
SCC phase filtering now matches native total-only, signature and cost-first
selection, retained Gray masks and permutation tie-run costs.
SCC-guided enumeration now connects those filters to permutation search and
optional stable cofactor-based group splitting under one shared work budget.
The uncached adaptive dispatcher now matches native encoded thresholds, cost
estimates and algorithm/simple/full branch selection.
Uncached cost-aware search now preserves fitted broadening decisions, sorted
representative resets and replacement cost-first phase filtering.
Hierarchical representative storage now matches native stage insertion/hit
semantics, with explicit bounded, retryable traversal lifetimes.
The hierarchical nonexact heuristic now connects all normalization/refinement
stages to that cache, preserving native early-hit and shallow-cache behavior.
Cached adaptive dispatch now preserves negative-stage aliases, cache/full flags
and the simple prepass before selected full search.
Cached cost-aware search now preserves stage hits, full/hard flags, broadened
search and one-level final storage.
Exact-search permutation/Gray schedules now match native execution order with
checked factorial counts and bounded allocation.
Unrestricted and adjacent-group exact minimization now preserve native
comparison/traversal behavior with shared schedule/candidate bounds.
The exact hierarchical stage now applies native signature-group/phase policy
before final caching; mixed-radix no-op work has a separate transition bound.
Balanced-output wrapping now preserves both-polarity evaluation, shared-cache
order and native winner/tie behavior across concrete canonicalizer modes.
Recursive Fast cofactor populations now preserve native constant-block and
byte-leaf behavior, with explicitly logical small-table counts.
Quick packed populations now preserve all generated lookup lanes while fixing
native total-byte overflow and recursive sibling-count overwrites.
Small truth tables now build native-compatible raw/hashed GIA mux trees, and
GIA output cones produce mapped six-slot truth words without recursive traversal.
Ordered structural cuts now produce owned multiword truth and native postorder,
preserving duplicate/unused leaf slots without touching graph scratch.
Exact/adaptive/hierarchical variants remain unported.

The latest completed slices add general two-design comparison miters with
shared or independent input suffixes and separate sequential state banks,
combinational AND/OR output aggregation, sequential half-state extraction,
combinational paired-output splitting, single-output AND/OR decomposition,
structural XOR-to-dual recovery, zero-state prefix cycling/recoding, and
contiguous output-group extraction with encountered-input projections,
object-order normalization, zero-output copies and independent/shared-CI
graph append, support-renamed structural output classes and constraint-preserving
raw/hashed graph normalization and bounded top-level AND query reduction.
Immutable fixtures and semantic tests cover modes, ordering, empty-output
behavior and explicit state/input/output projections.


Recent work covers GIA graph analysis, cone traversal, MFFCs/dominators,
structural signal detection, explicit signal mappings, and bounded cofactor
construction. Fixed-PI cofactors, multi-pivot Shannon reconstruction and the
initialized unroll/cofactor pipeline now have native and independent semantic
checks. Recent additions cover selected-output and sequential/gate cone extraction,
level/depth cuts, support-based reduction/slicing, positive constraint extraction,
one-hot assumption attachment, graph composition and bounded zero-initialized
sequential structural sweeping and general/dual-output trimming with explicit
source terminal/register projections.
PDR still lacks full native scheduling/mode parity; optimized CNF, broader
large-design validation, command integration and the remaining source tree
remain in scope. Per-source partial coverage is recorded in PORT_COVERAGE.json.

## Implemented and checked

| C area | Go implementation | Current coverage |
| --- | --- | --- |
| `misc/vec/vecMem.h` | `misc_vec_memory.go`, `misc_vec_truth_memory.go` | Owned paged uint64 records, zero-filled gaps, checked reads/writes/growth, corrected shrink, cloning, bounded hex output, stable-ID interning and truth-table/mux seeds. File wrappers and C bucket-layout internals remain absent. |
| `misc/vec/vecFlt.h` | `misc_vec_float.go` | Owned float32 vectors, allocation/copy/move/release, arithmetic/update, fill/growth, ordered/unique insertion, lookup/removal, extrema/equality and stable numeric sorting. NaN storage survives; NaN ordering rejects safely. Borrowed pointers and print helpers remain absent. |
| `misc/vec/vecQue.h` | `misc_vec_queue.go` | Validated: complete owned float32 max-heap with sparse IDs, C-compatible ties, cost updates/replacement, cloning, sparse invariant checks and deterministic source-layout formatting; Go lifetime/tests replace free/demo helpers. |
| `misc/vec/vecBit.h` | `misc_vec_bit.go` | Owned bounded packed bits; rounded starts/fills, exact growth, entry/update/add, push/pop, shrink/reset, count, cloning/release and snapshots. Go corrects unsafe fill-extra boundaries and C memory accounting; borrowed pointers and explicit frees are not exposed. |
| `opt/dau/dauCanon.c` | Truth population, phase, refinement, adaptive/SCC/cost-aware, hierarchy and balanced-wrapper files | Validated: complete direct/Slow/Fast population paths, small-word normalization, heuristic and exact grouped pipelines, cached hierarchy, balanced wrapper, explicit private/debug/unused-helper dispositions, bounded search and owned storage. |
| `opt/dau/dauCount.c` | `misc_util_truth_population_quick.go` | Complete generated lookup, owned snapshot, deterministic formatter and corrected recursive Quick counting. Go-only unit/fuzz tests replace randomized native drivers; historical total-overflow and cofactor-overwrite regressions are pinned. |
| `opt/dau/dauMerge.c` | `opt_dau_merge*.go` | Complete owned permutation-aware shared-variable merge, purity extraction/inlining, general reduced-truth decomposition, associative simplification and normalization; debug-only stdout/timing state is deliberately omitted. |
| `opt/dau/dauNonDsd.c` | `opt_dau_non_dsd_*.go` | Validated: checked set roles, source-order discovery/pruning, owned composition/decomposition truth tables, exhaustive reconstruction, general decomposer formula pairs through twelve component inputs, complete shared-minterm polarity families, deterministic report data and explicit demo disposition. |
| `opt/dau/dauNpn.c` | `opt_dau_npn_*.go` | Complete: exact zero-through-four-input class tables, symmetry/input-flip pruning, complete four-input network frontiers, exact three-through-six-input function frontiers, bounded table codec, owned exact summaries and deterministic dumps with Go-only model/orbit fuzzing. |
| `bool/lucky/luckySimple.c` | `bool_lucky_permutation.go`, `bool_lucky_minimize.go` | Validated: all schedule/allocation helpers and both unrestricted/grouped exact minimizers are represented by owned bounded Go behavior, including no-op transition limits, hierarchical integration and corrected single-input padding. |
| `misc/util/utilTruth.h` | `misc_util_truth*.go` | Canonical packed tables, Boolean operations/comparisons, cofactors/support, flips/swaps/permutations, NPN transforms, quantification, compaction/stretching, checked embeddings, literal simplification, output decomposition, pair-input DSD and corrected mux-input DSD. Owned results replace pointer aliases; larger canonicalization/decomposition algorithms remain. |
| `misc/util/abc_global.h` | `misc_util_literal.go` | Separate checked AIG/SAT literal types, complement operations, packed-word counts and allocation arithmetic. Other utility/header functions remain. |
| `misc/util/utilCex.c/.h` | `misc_util_cex.go` | Validated: owned allocation/full masks, construction, cloning/input duplication, permutations, phase/temporal/unknown-initial transforms, checked merge, register-aware combinational models, packed bits/population, statistics and deterministic trace formatting. Go lifetime replaces free/sentinel APIs. |
| `aig/aig/aig.h`, `aigMan.c`, `aigObj.c`, `aigCheck.c` | `aig_aig_manager.go`, `aig_aig_object.go`, `aig_aig_object_rewrite.go` | Constant-one graph, CI/CO/buffer creation, stable IDs, hashing, reference/level/phase validation, dangling cleanup, independent cloning, atomic recursive deletion, CO deletion/patching, owned object formatting, explicit-map object replacement and topological buffer propagation. Broader `aig.h`/`aigMan.c` lifecycle surface remains partial. |
| `aig/aig/aigOper.c` | `aig_aig_oper.go` | Default-mode AND/OR/XOR, mux, majority and balanced multi-input trees. Extra-strashing, special XOR nodes, miters and demonstration helpers remain. |
| `aig/gia/gia.h`, `giaMan.c`, `giaHash.c`, `giaAig.c`, `giaScl.c`, `giaDup.c` | `aig_gia_manager.go`, `aig_gia_hash.go`, `aig_gia_check.go`, `aig_gia_convert.go` | Three-word compact objects; CIs/COs/buffers/ANDs/native XORs/native muxes; hashing and Boolean constructors; output patching; register/constraint metadata; dense cleanup and explicit maps; AIG conversions, evaluation and phases. The remaining GIA algorithms, choices, extended strashing, mapping/timing metadata and optimized modes remain. |
| `aig/gia/giaDfs.c`, `giaFanout.c` | GIA traversal and fanout files under `aig_gia_*.go` | Validated: complete owned level grouping, reverse order, sequential/TFI/TFO/support/cone traversal plus static, dynamic and mapped fanout behavior; native print/timing drivers are replaced by returned data and Go tests. |
| `aig/gia/giaUtil.c` | GIA analysis files under `aig_gia_*.go` | Checked forward/reverse/required levels and slack, structural mux/XOR recognition, references and cross-cut analysis. Other utility variants remain partial. |
| `aig/gia/giaMffc.c` | `aig_gia_mffc.go`, `aig_gia_dominators.go`, `aig_gia_dom_derive.go`, `aig_gia_mffc_candidates.go` | Validated: owned reference-balanced MFFC boundaries, closest dominators, bounded cofactor cone derivation and deterministic candidate records replace every algorithm and diagnostic side effect. |
| `aig/gia/giaCut.c` | `aig_gia_cut_enumerate.go` | Batch AND/XOR cut enumeration, truth normalization/interning, support reduction and native ranking. Fixed-graph store lifecycle is available; live-graph updates, allocation reuse and diagnostics remain. |
| `misc/tim` active module, headers and manifest | `misc_tim_manager.go`, `misc_tim_box.go`, `misc_tim_delay.go`, `misc_tim_dump.go`, `misc_tim_transform.go`, `misc_tim_configure.go`, `misc_tim_format.go` | Complete owned box/timing/traversal/persistence surface, manager transforms, library configuration and deterministic diagnostics. Go ownership replaces explicit frees; checked errors replace assertions. |
| `aig/gia/giaSatLut.c` | `aig_gia_lut_unit_timing.go` | Unboxed/boxed unit arrival/required timing with corrected reverse traversal. SAT remapping and window machinery remain. |
| `aig/gia/giaEdge.c` | `aig_gia_edges.go` | Owned rows, packing conversion, forward/reverse selection and no-box timing. Timing boxes, mapping2, mutations and window optimization remain. |
| `aig/gia/giaIf.c` | `aig_gia_lut_mapping.go` | LUT/fanin/max-size statistics, object/CO levels, mapped references and structural overlap. Mapping selection, optimization and timing-box integration remain. |
| `aig/gia/giaTruth.c` | `aig_gia_truth.go`, `aig_gia_cut_truth.go` | Owned Shannon mux construction, iterative mapped output words and ordered multiword cut truth with native postorder/leaf slots. Reusable truth managers, LUT mapping and support paths remain. |
| `aig/gia/giaCof.c`, `giaDup.c`, `giaEnable.c` | GIA cofactor, signal detector and transfer files under `aig_gia_*.go` | Single/multi-pivot reconstruction, fixed-PI cofactors and mux composition; thresholded structural sequential candidates; explicit frame signal transfer; initialized unroll/cofactor pipeline. Immutable fixtures, exhaustive truth checks and independent fuzz oracles cover these paths. Other duplication/enable/cofactor modes remain. |
| `base/io/ioReadAiger.c`, `ioWriteAiger.c` | `base_io_aiger.go` | Standard binary AIGER and ABC `aig2` reading, standard binary writing, latch initial values, output/bad/constraint sections, symbols in insertion order using `insdict.Dict`, comments, and bounded checked decoding. Compression, compact writing and additional wrapper APIs remain. ASCII AAG is not yet implemented. |
| `sat/bsat/satSolver.c/.h`, `satClause.h` | BSAT solver/session/order/check/bookmark/gate/compatibility files | `satSolver.c/.h` are validated: owned watched-clause CDCL, first-UIP learning/minimization, assumptions and final conflicts, incremental original/learned clauses, activity heap and factors, phase saving, Luby restarts, budgets/cancellation, projection/callback state, rollback, gates, trace/store/utilities and learned-clause reduction. `satClause.h` remains tracked separately. Exhaustive tiny-formula/gate tests and independent Go-only model fuzzing pass. |
| Simple CNF reference path from the port plan | `sat_cnf_simple.go` | Full Tseitin relation for AIG constants/CIs/ANDs/buffers/COs, explicit input/output mappings and offset loading into the incremental solver. This does not count as porting optimized `sat/cnf` source. |
| `sat/bmc/bmcUnroll.c` | `sat_bmc_unroll.go`, `sat_bmc_ranked_unroll.go` | Validated simple and ranked timeframe unrolling with owned mappings, iterative tent/rank setup, typed circular histories, incremental atomic frames, bounded storage, zero/one/unknown initialization on the simple path, rank/statistics queries and Go-only exhaustive/fuzz equivalence tests. |
| `sat/bmc/bmcBmc.c:Saig_ManBmcSimple` | `sat_bmc_simple.go` | Per-output bounded SAT search, earliest failure, original-model counterexample extraction/replay, property selection, constraint prefixes, zero/one/unknown initialization, budgets/cancellation and resource caps. Uses the reference Tseitin encoder; optimized CNF, rewrite/cofactor options and alternate backends remain. Finite exclusion never claims unbounded safety. |
| `proof/pdr/pdrUtil.c`, `pdrInt.h` | `proof_pdr_cube.go`, `proof_pdr_queue.go` | Immutable register/PI cubes, signatures, containment, reductions, intersections, init tests, ordering, and stable proof-obligation queues. These utilities are tested against C; the basic engine is tracked separately below. |
| `proof/pdr/pdrCore.c`, `pdrSat.c`, `pdrInv.c` | `proof_pdr_core.go`, `proof_pdr_sat.go`, `proof_pdr_invariant.go` | Basic IC3/PDR frame search, full-state obligations, assumption-core reduction, cube deletion generalization, propagation/fixed points, replayed witnesses and independently checked invariants. Dedicated exhaustive/native/fuzz tests pass; advanced ABC scheduling, CTGs, abstraction, optimized CNF, export and remaining APIs are not ported. |
| `base/io/ioWriteAiger.c:Io_WriteAigerCex` | `base_io_aiger_cex.go` | Ivy/ABC four-vector trace writing and checked reading, full transition replay, initialization/positive-constraint checks and selected-property validation. Other witness formats remain. |

All coupled implementation stays in package `abc`; no custom Go interfaces,
cgo dependency or runtime calls to the C executable were added. Ordinary Go
tests read permanent fixtures under `test_vectors/`; no supported test invokes
an ABC executable, archive, C compiler or generated C oracle.

## Explicit differences and representation contracts

* Go source starts with the package declaration. Source references follow it;
  no generated boilerplate is placed before package declarations.
* AIG literal 0 is constant **one** and literal 1 is constant **zero**. External
  AIGER uses the reverse constant convention. Conversion is explicit.
* A manager's `ObjectCount` is its allocated ID span, including tombstones,
  corresponding to `Aig_ManObjNumMax`. Cleanup preserves live IDs and never
  reuses removed ones. Literal ownership is per-manager; numeric literals
  cannot detect accidental use in another manager with the same ID range.
* Graphs are topologically append-only in this slice; arbitrary rewiring,
  compaction/conversion mappings and traversal IDs remain future work.
* Counterexample storage always has consistent dimensions and cleared padding.
  C's full allocator fills padding and intermediate header fields with ones;
  Go fills logical data only. `CountOnes` therefore counts logical bits.
* `Clone` preserves initial register values. `DupInputs` and input permutations
  intentionally retain C's behavior of zeroing initial register values.
* Phase transformation fixes C's inconsistent `nBits` update when the new
  register count differs from the original. Logical input order is unchanged.
* Checked merge accepts only a zero `part.Property` offset; nonzero values in
  C's overloaded `iPo` arithmetic make its allocation and copy lengths disagree.
  Such inputs are rejected rather than causing an out-of-bounds copy.
* `CounterexampleFromCombModel` covers the observed zero-register C call.
  `CounterexampleFromCombModelWithRegisters` additionally applies the intended
  register offset; C writes model bits over initial-state positions when the
  otherwise unused nonzero-register argument is supplied.
* Cleanup also removes dangling buffer nodes. The initial C cleanup scan starts
  only from AND/XOR nodes. Buffer extension behavior has its own tests.
* GIA literal 0 is constant zero, with 29-bit backward fanin differences in
  three 32-bit words. Raw and hashed construction are separate operations.
  Hashing records the first raw duplicate instead of the C assertion on
  duplicates during hash startup. Lookup maps are never enumerated.
* GIA cleanup returns a new graph and an explicit map, preserves all buffers
  and their cones as C does, and leaves source marks/Value untouched. AIG
  conversion lowers native XORs/muxes correctly; the surveyed C simple
  conversion path treats those encodings as ANDs and is not its semantic oracle.
* AIG and GIA conversions now preserve constraint counts explicitly. AIGER
  model validation rejects disagreement between model and graph metadata.
* AIGER constraint signals remain **positive assumptions** in `AigerModel`.
  C's network reader instead inverts its internal constraint outputs. The native PDR
  fixture uses unconstrained zero-initialized models; constrained and arbitrary-init
  Go PDR semantics are checked independently rather than claimed as native parity.
* Justice/fairness sections are rejected, matching the surveyed ABC reader.
  Unknown latch initialization remains distinct from zero and one.
* Four-vector traces retain actual initial state bits; the C writer prints
  zeros unconditionally. Zero-initialized fixtures match C byte for byte.
* SAT uses stable clause handles with ordinary watched clauses instead of C's
  tagged binary handles. Integer activity and Luby algorithms are ported;
  random draws, phase saving and reduction choices are deterministic for a
  fixed configuration, but identical C search traces are not promised.
* Learned unit clauses retain explicit reasons and are reinstalled at root
  level for later solves, avoiding assumption/core lifetime ambiguity.
* The original snapshot can return SAT on repeated calls after unconditional
  UNSAT. Go retains base inconsistency; the observed sequence is preserved in
  `test_vectors/oracle/sat_repeated_unsat_observed.json` and tested against
  exhaustive semantics. The frozen fixture stops after base UNSAT while
  continuing to exercise repeated assumptions on SAT bases.
* Bounded checking reports a counterexample, finite exclusion or unknown.
  Finite exclusion is not an unbounded safety result. Constraints are added
  only for the current prefix, so impossible future constraints cannot hide
  a valid earlier bug. Every returned witness replays on the original model.
* SAT work budgets apply per query, matching the C simple-BMC loop; object,
  mapping and variable caps bound preprocessing. Cancellation is checked
  between phases, unrolling nodes, CNF clauses and SAT work.
* The C simple BMC return code remains -1 after all finite queries are UNSAT.
  Go distinguishes finite exclusion explicitly while preserving the distinction
  from unbounded proof. The basic PDR engine separately reports unbounded proofs only with checked certificates.

## Baseline evidence

Permanent compatibility evidence is now stand-alone. All reviewed inputs and
expected records live under `test_vectors/`; `reference_build.json` schema 2
contains only policy and SHA-256 fixture identities. Package tests verify every
listed artifact, reject untracked files in `test_vectors/oracle`, and reject any
native C/C++ source anywhere in the Go module. They need no ABC executable or
archive, C compiler, C headers, cgo, subprocess oracle, or original source
checkout. The strengthened guard, full suite and race suite pass in 0.046s,
10.317s and 59.405s respectively; vet is clean.

There is no checked-in `testdata/` tree. Go fuzzing may create ephemeral
`testdata/fuzz` output locally, but every retained regression must be promoted
to `test_vectors/` and exercised by an ordinary Go test.

The frozen datasets retain results captured while migration comparisons were
still available, but correctness does not rest on provenance alone. Go tests
also use exhaustive Boolean evaluation, independent state-machine models,
certificate replay, format round trips, invariants and fuzz properties. The
old 287 C data generators, native smoke runner, compatibility header and
optional native CEC path have been removed. `source_manifest.json` remains
planning-only inventory evidence and is itself hash-checked; it is not loaded
by the implementation or required to run tests.

The foundation datasets include counterexamples, 128 four-input AIG operations,
128 SAT formulas with six queries each and 128 mixed GIA operations. SAT answers
are checked by independent enumeration, models and final conflicts; GIA truth
tables are checked by an independent Boolean model. Go-only AIGER tests parse,
serialize, reparse and require a stable canonical second encoding for constants,
a small AND graph and `i10.aig`. A 60-second GIA fuzz campaign passed 542,689
executions.

The current `i10.aig` has 257 inputs, 224 outputs and 2,675 ANDs, confirmed by
native ABC. The upstream README's historic 2,396-AND example is a different
baseline and is not used as a test expectation.

## Validation performed

The normal suite, `go vet`, race checks and ten shuffled repetitions passed.
The latest package-wide statement coverage measurement is 89.0%; coverage is
an aid and does not establish full algorithm parity. Learned-database deletion
and subsequent SAT model recovery also pass a guarded pigeonhole regression.
The three selected AIGER cases pass Go-only semantic and canonical-round-trip
checks. Eleven fuzz targets have passed bounded campaigns:

| Target | Duration | Executions |
| --- | ---: | ---: |
| `FuzzCounterexampleLayout` | 60 seconds | 3,918,514 |
| `FuzzAigOperations` | 60 seconds plus 15-second follow-up | 78,484 total |
| `FuzzAigerRead` | 60 seconds | 3,284,634 |
| `FuzzABCTraceRead` | 60 seconds | 3,219,977 |
| `FuzzSatIncremental` | Two 60-second campaigns | 2,456,666 total |
| `FuzzAigCNFProjection` | 30 seconds | 1,010,523 |
| `FuzzGiaOperations` | 60 seconds | 542,689 |
| `FuzzGiaUnroll` | 60 seconds | 1,895,977 |
| `FuzzBmcSmallMachines` | 60 seconds | 687,307 |
| `FuzzPdrCubes` | 60 seconds | 1,820,198 |
| `FuzzPdrQueue` | 60 seconds | 1,598,232 |

The follow-up AIG run capped minimization at one second to improve campaign
throughput. These finite campaigns are not exhaustive assurance.

Reproduce normal validation from the repository root:

```
go -C pkg test ./...
go -C pkg vet ./...
go -C pkg test -race ./...
go -C pkg test -shuffle=on -count=10 ./...
go -C pkg test -run '^$' -fuzz '^FuzzAigerRead$' -fuzztime=60s -fuzzminimizetime=1s -parallel=2 .
```

## Remaining milestones

M0 still needs a complete configuration/symbol audit and broader reviewed Ivy
examples. M1 needs
the remaining required utility operations and broader lifecycle/ordering tests.
M2 now includes the compact GIA graph core, cleanup maps and AIG conversions;
remaining GIA algorithms, representation metadata, graph operations, format
support and complete I/O parity are still outstanding. M3 now has an initial bsat
CDCL implementation; remaining solver APIs, optimized CNF and all alternate
backends are outstanding. The independent simple CNF relation and offset
instantiation are tested against exhaustive existential projection. M4 now has
checked simple timeframe unrolling and bounded SAT queries/witness extraction;
Basic PDR now has checked witnesses/proofs, while full PDR parity and the remaining sequential verification algorithms are still outstanding. M5–M9
(command integration, synthesis, sequential transformations, mapping, word-level/
hierarchy, BDDs and the remainder of `src`) remain unimplemented. The first Ivy executable milestone
is still ahead; no existing Ivy executable has been replaced.

The latest unrolling fixture matches C for a four-frame, two-register machine
(19 compact objects and eight outputs), plus exhaustive independent traces
with zero, one and unknown initial values. The source graph remains unchanged.

Bounded checking passes independent reachability checks on 256 generated
three-variable truth-table machines, including assumptions and arbitrary
initial states. A frozen dataset adds 256 bounded queries over 64
zero-initialized machines; all agree on finite exclusion and first failure.
The bounded-checker fuzz campaign passed 687,307 executions in 60 seconds.

PDR utility validation now covers 32 native cubes (including signatures,
removal/subset creation, 1,024 containment and 1,024 comparison results) and
64 native proof-obligation insertions/pops. Exhaustive independent tests cover
four-register partial cubes and all zero/one/unknown initial configurations.
Both utility fuzz campaigns passed (1,820,198 cube and 1,598,232 queue
executions over 60 seconds each).

PDR cubes canonicalize duplicate equal assignments and reject contradictory
ones. Go owns their state/PI slices; queue ownership prevents double insertion
and rescheduling while linked. The protected-intersection check inspects tails
that the original loop misses: its observed result is retained as an immutable
regression datum, while Go rejects dropping that protected literal. Allocation
counts are replaced by Go lifetimes; witness successor chains remain intact.

The basic PDR engine agrees with complete reachability on 2,048 generated
machines and 5,376 targeted transition-mask cases, plus 128 cases with
generalization disabled. These cover arbitrary two-register initial values
and positive environment constraints. Every proof is independently checked
by enumerating initiation, induction and safety; every witness is replayed
both through the original model and an independent truth-table simulator.
All 256 native PDR fixture results agree with independent reachability and Go.
The invariant verifier rejects deliberate initiation, induction and safety
failures, invalid property/register references, cancellation and incomplete
checks. PDR reachability fuzzing passed 185,379 executions over 30 seconds.
Full Go tests, full race tests, five shuffled runs and vet pass with this engine.

PDR uses complete-state obligations and full Tseitin frame relations rather
than ABC's ternary shrinking and optimized CNF. Assumption-core reduction followed by deletion generalization is
conservative; advanced CTG scheduling and abstraction remain.
Counterexamples need not be shortest. Proof certificates are rechecked using
separately constructed original-graph Boolean miters before returning proved.
Initiation is guarded by positive environment constraints: a model without an
admissible initial input is vacuously safe. Query budgets include certificate
verification; interrupted or incomplete verification cannot return proved.

Additional PDR regressions cover combinational constant properties, latch
self-loops/toggles, known-one and unknown initialization, impossible constraints
and positive-constraint polarity. Query caps of one through four cannot claim
a proof when the final certificate needs additional queries; five permits the
complete three-condition certificate check on the constant-safe example.

PDR now ports `Pdr_ManReduceClause` before deletion generalization. Negated
SAT assumption conflicts are translated through exact next-state literals,
including alias/complement handling. The canonical first register represents
an aliased assumption. If the resulting cube intersects the reset predicate,
the first original literal excluding the actual zero/one/unknown reset is
restored (C's implementation assumes all-zero reset). Unknown resets are never
silently treated as zero. Unmapped core literals and an original cube without
a reset blocker are rejected. Lookup-only maps preserve deterministic output;
cancellation is checked during translation. Dedicated regression tests cover
these cases. The full reachability/native suite passes; another 30-second PDR
fuzz run passed 182,758 executions after adding core reduction. Full race tests
and vet also pass. Relative-induction activation is now implemented as described below; advanced
generalization remains outstanding.

Relative-induction queries now port the guarded-clause path of
`Pdr_ManCheckCube`: deletion generalization checks
`F & !candidate & T & candidate_next`. A fresh activation variable guards each
current-state exclusion; a unit clause disables the guard after the query,
including cancelled queries. The engine keeps full-state predecessor search
separate, preserving exact witness chains. Targeted tests distinguish relative
induction from ordinary predecessor UNSAT on a self-loop and verify that no
guard survives into later ordinary queries. All activation variables count
against both per-solver and aggregate frame-variable limits, including when
new frames are allocated. Go explicitly disables spent guards and now recycles frame solvers as described below.

After relative-induction integration, full race tests, five shuffled suite
runs, vet, exhaustive reachability and the 256-case immutable fixture pass.
A further 30-second reachability fuzz campaign passed 180,465 executions.

Frame-solver recycling now ports `Pdr_ManFetchSolver`: the default threshold
is ABC's 300 activation variables, configurable using `RecycleVariables`.
Rebuilds reload the complete transition relation, positive constraints, exact
initialization for the initial frame, and blocked cubes in insertion order.
SAT-learned clauses and spent guards are discarded; their work statistics are
accumulated before replacement. Per-solver limits and deterministic seeds are
preserved. Active guarded queries cannot trigger recycling. Tests perform 40
rebuilds across initial and later frames and verify preserved clauses, reset
conditions, SAT answers, variable counts, limits and accumulated statistics.
Another 256 independent reachability cases force recycling after every guard.

Recycling validation passes full race tests, five shuffled runs, vet, and a
30-second forced-recycling fuzz campaign (168,285 executions). A targeted
cancellation test interrupts relation loading during rebuild; the caller
returns UNKNOWN rather than using the partially reconstructed frame.

Invariant PLA export now covers the full-register path of
`Pdr_ManDumpClauses`. `PdrInvariant.MarshalPLA` writes an excluded-state cover:
PLA output one means the state violates the invariant. Export alone does not
certify a proof. Output is deterministic, preserves ABC cube sorting, checks
register references, and bounds complete output allocation before sorting or
building rows. Empty covers, empty cubes and zero-register models have explicit
tests. A pinned immutable fixture contains all 81 four-register partial cubes;
Go output is checked byte for byte against native sorting/printing. Model names,
timestamps and variable labels remain unported; support compression is now covered below.
Export fuzzing passed 631,162 executions in 30 seconds, independently comparing
the emitted cover with invariant evaluation on every four-register state.

`PdrInvariant.ToAig` now ports `Pdr_ManDupAigWithClauses`. One primary
output per blocked cube replaces the original properties and constraints;
register order, PIs, transition functions, name and specification are retained.
These outputs denote excluded states, not safety assertions. Graphs do not
store latch initialization; callers must retain the original AIGER model when
interpreting proofs or traces. Go preserves the source graph and scratch-free
literal mapping, aliases buffers safely, supports tombstone IDs and cleans
new dangling nodes. Object/mapping limits and cancellation are checked.
Independent tests exhaust all eight CI assignments on 256 arbitrary
one-input/two-register transition systems, checking every cube output and next
state. Additional regressions cover empty invariants, invalid references,
source immutability, buffers, tombstones, resource caps and cancellation.
Native byte-level graph parity for this conversion has not been claimed.

Cube-to-AIG fuzzing passed 372,328 executions in 30 seconds. The full race
suite and vet pass with this conversion.

Compact invariant vectors now implement the encoding portion of
`Pdr_ManDeriveInfinityClauses`: signed 32-bit entries contain cube count,
per-cube literal count/literals, and final register count. `ABCVector` exports
owned data; `ReadPdrABCVector` validates the full envelope before cube allocation,
checks signed counts/references and exact consumption, rejects contradictory
literals and canonicalizes duplicates. Both directions have entry caps and
signed-format limits. Empty invariants and empty cubes remain distinct.
Importing or serializing a certificate does not verify it. Tests cover
malformed/truncated records, extra entries, invalid references, resource caps,
input ownership and canonical roundtrips. Vector fuzzing passed 1,361,288
executions in 31 seconds. Native invariant-frame collection remains unported; non-inductive clause
pruning is now covered separately below.

Support-aware invariant export now ports register occurrence counting and the
increasing support map used by `Pdr_InvCounts`, `Pdr_InvMap`,
`Pdr_InvUsedFlopNum` and `Pdr_InvPrintStr`. `RegisterCounts` returns owned
occurrence counts; `MarshalSupportPLA` returns deterministic compressed PLA
and the original register index for each column. Unused registers are omitted;
empty support and repeated cubes retain their semantics. Register-allocation
and output caps are enforced. Native support-printing output matches all 81
partial cubes on four nonadjacent registers in an eight-register domain.
Fuzzing independently checks compressed covers over every eight-register
state and passed 405,956 executions in 30 seconds. Full race tests and vet pass.
The support map must accompany exported data; compressed column numbers alone
do not identify the original model registers.

`PdrInvariant.PruneNonInductive` now ports
`Pdr_ManDeriveMarkNonInductive` and the repeated pruning loop in
`Pdr_ManDeriveInfinityClauses`. Each pass loads all currently retained clauses,
checks each next-state cube, removes violated clauses simultaneously, and
rebuilds until unchanged. Clause order and source data are preserved. Positive
environment constraints guard transitions. Query, clause and SAT-variable caps
bound work; cancellation or SAT UNKNOWN returns no candidate invariant. This
differs from C marking every non-UNSAT query as non-inductive. Completion means
transition closure only, not initiation or safety: callers still use `Verify`.
Independent tests compare every pass and retained clause on 256 truth-table
machines, including cascading removal, empty/false predicates, impossible
constraints, budgets and cancellation. Native frame collection and advanced
PDR modes remain outstanding.

Fixed-point pruning fuzzing passed 236,516 executions in 30 seconds. Full
race tests and vet pass, including a regression showing a transition-closed
false predicate still fails initiation and therefore cannot certify safety.

PDR now applies the containment checks of `Pdr_ManCheckContainment` and
same/next-frame subsumption from `Pdr_ManPushClauses` to its explicit cumulative
frame representation. Insertion skips a clause already implied by a smaller
blocked cube; a new stronger clause removes superseded cubes from the ordered
`insdict.Dict`. Propagation skips candidates already blocked in the next frame.
The SAT solver can retain redundant clauses until recycling because the new
clause implies them. Rebuilding loads only the retained ordered entries.
Containment/subsumption counters distinguish avoided insertions and removed
entries from cumulative SAT clause additions. Targeted tests evaluate every
state before and after recycling, check both directions of containment and
retain unrelated clauses. Exhaustive tiny-machine and native PDR fixture
comparisons continue to pass; full ABC scheduling/CTG modes remain incomplete.

After containment integration, full race tests, five shuffled runs and vet
pass. PDR reachability fuzzing passed 176,112 executions in 30 seconds with
independent witness replay and exhaustive certificate checks.

Default dynamic flop priorities now port `Pdr_ManSortByPriority` and the
priority updates in `Pdr_ManBlockCube`. Generalization tries lower-occurrence
registers first using C's selection-sort swaps, including tie behavior. Cube
storage stays canonical; removal order is tracked by literal identity as cube
sizes change. Each learned cube updates priorities once, independently of how
many cumulative frames receive its clause. Default counters begin at zero;
Go omits C's common power-of-two scale because it cannot affect comparisons.
Unsigned 64-bit counters rescale before overflow instead of wrapping signed C
arithmetic. Regression tests cover exact ordering, ties, source immutability,
updates, overflow, invalid register references and cancellation. Structural priority initialization and CTG/down modes remain unported;
optional analyze-final assumption ordering is now covered below. Full exhaustive/native PDR semantic comparisons still pass.

With dynamic ordering enabled, full race tests and vet pass. Another
30-second independent PDR reachability fuzz run passed 173,396 executions.

Clause propagation now ports the assumption-core reduction step of
`Pdr_ManPushClauses`. A predecessor-UNSAT query can push a smaller cube,
restoring an actual-reset blocker if necessary. The stronger clause is added
to every earlier cumulative frame as well as its destination, preserving frame
implication after strengthening. Propagation traverses a snapshot because a
reduced clause can subsume multiple source-frame entries during the pass.
A targeted regression begins with complementary two-literal cubes, reduces
them to one register literal, removes both source entries and confirms all
four state assignments in both converged frames. Exhaustive reachability and
native PDR fixture comparisons continue to pass. This extends the basic
propagation path; CTG/down scheduling and lazy optimized CNF remain unported.

With push-core reduction, reachability fuzzing passed 172,345 executions in
30 seconds; full race tests and vet pass.

Optional higher-priority-first assumption ordering now ports
`Vec_IntSelectSortPrioReverseLit` from the `fFlopOrder` path.
`PdrOptions.OrderAssumptions` applies its selection-sort order to next-state
SAT queries, including guarded relative-induction queries. Canonical stored
cubes and current-state queries retain their original order. Typed SAT mapping
is applied after register sorting, preserving complements and aliases. This
is a basic-engine heuristic option, not full ABC `fFlopOrder` scheduling parity:
Go applies it consistently to next-state queries rather than duplicating all
of ABC's mode-specific call sites. Targeted tests check reverse priority ties,
complemented mappings, disabled mode and cube immutability. Another 256
independent small-machine reachability/certificate checks exercise ordering
with recycling after every guard.

Ordered-assumption validation passes full race tests, five shuffled suite
runs and vet. Combined ordered-assumption/frequent-recycling fuzzing passed
171,075 executions in 30 seconds against independent complete reachability,
witness replay and certificate enumeration. The first sandboxed fuzz attempt
stopped on a read-only cache write; the completed run used writable cache access.

`GiaManager.ReferenceCounts` now ports `Gia_ManCreateRefs` from `giaUtil.c`,
a prerequisite for PDR structural flop priorities. It counts every incoming
edge, including COs, constant references and dangling cones; buffers contribute
one edge, native XORs two and native muxes three. Results are owned uint64
slices instead of mutable manager scratch. A mixed-node graph matches the
frozen compatibility dataset exactly; independent expected counts, source
immutability, ownership, exact object caps and invalid caps are tested. The
remaining `giaUtil.c` routines and PDR structural priority initialization remain
unported. ASCII AAG is still unsupported: the inspected ABC reader itself
implements binary formats, so adding AAG would be an additional format feature.

`GiaManager.RecognizeMux` now ports `Gia_ObjIsMuxType` and
`Gia_ObjRecognizeMux`, including the C precedence among four matching
complementary-grandchild pairs. Returned control/then/else literals describe
the regular root object (a complemented caller literal requires complementing
the data result). All 576 phased AND-pair patterns over four variables match
native detection and exact decomposition literals; every recognized mux is
also checked independently on all 16 input assignments. Native mux metadata
is supported explicitly. Native XOR encodings and buffers are not treated as
ordinary AND patterns, preventing misinterpretation of extended nodes.
Native mux phases, nonmatches and invalid IDs have dedicated tests. This is
another prerequisite for structural PDR priorities; those priorities are not
yet connected to the engine.

Structural PDR priorities now port `Pdr_ManDeriveFlopPriorities2` and the
active `Abc_MergeSortCost` algorithm. Costs use register reference counts and
an optional mux-control bonus exceeding the largest register fanout. Exact
selection-sort base cases and alternating equal-cost merge behavior preserve
native rankings; the 16-register pinned fixture checks nontrivial ties.
`PdrOptions.StructuralPriorities` connects these ranks to the engine, with
`MaxPriorityObjects` bounding preprocessing. Dynamic increments use a common
power-of-two step larger than the initial ranks, retaining structural tie
breaks. Default priority initialization remains zero. Native mux controls are
handled explicitly in addition to standard AND patterns. Tests cover control
bonus ranking and 256 independent reachability cases combining structural
priorities, assumption ordering and frequent recycling. Full race tests and
vet pass; other structural priority variants and CTG modes remain unported.

`GiaManager.StructuralFlopPrioritiesStable` now ports the alternate
`Pdr_ManDeriveFlopPriorities3` ranking. It shares the reference/mux-control
cost calculation with variant 2 and preserves ascending register order within
each equal-cost bucket. Go sorts register indices stably instead of allocating
one bucket per possible cost; the resulting ranks match C without memory
proportional to the largest sparse cost. A separate pinned 16-register fixture
matches native output and explicitly differs from the merge-sort tie variant.
Tests cover stable ties, empty register sets, mux-control bonuses, invalid
limits and exact object bounds. This method is available separately; engine
initialization retains ABC's selected variant 2. Remaining structural variants,
CTG/down modes and optimized CNF remain unported.

`GiaManager.RecognizeXor` now ports `Gia_ObjRecognizeExor` and also handles
native XOR encodings explicitly. Returned fanins XOR to the regular object,
with complement phases retained. `MuxControlReferenceCounts` ports
`Gia_ManCreateMuxRefs`: counts only recognized control uses and skips XOR
patterns that also admit mux decompositions. It includes dangling cones,
returns owned counts and bounds object allocation. Native counts match all
576 phased AND-pair cases; XOR detection is independently checked against
complementary-grandchild structure, and every recognized XOR is evaluated on
all 16 assignments. Extended-node tests cover native XOR phases, native mux
control counting, nonmatches, invalid IDs and resource limits. Remaining
`giaUtil.c` operations, optimized CNF and advanced PDR modes remain unported.

GIA forward level assignment and grouping now port `Gia_ManLevelNum`,
`Gia_ManLevelize` and `Gia_ManBfsForCrossCut`. `Levels` returns owned per-object
levels and maximum depth; `Levelize` groups AND-like objects by depth in ID
order; `BreadthFirstCrossCutOrder` concatenates those groups. Constants/CIs
start at zero, COs/buffers inherit their driver, ANDs add one, native XORs/muxes
add two. Dangling nodes participate. The native mixed-node fixture and an
independent weighted-depth expectation agree exactly. Tests cover grouping,
ordering, zero-node cases, a 100-level dangling chain, source immutability,
result ownership and object caps. Actual frontier-size computation, DFS traversal and the C fGiaSimple buffer
mode remain unported; reverse levels are now covered below.
Full race tests and vet pass with these additions.

`GiaManager.ReverseLevels` and `ReverseLevelize` now port
`Gia_ManLevelRNum` and `Gia_ManLevelizeR`. The implementation deliberately
preserves ABC's reverse metric: CO drivers start at one, buffers add zero,
AND-like nodes propagate one step through their two packed data fanins, and
native mux control edges are omitted. It is not the weighted forward metric
or timing analysis. Groups include COs at zero and all nonconstant objects
with positive reverse level, in increasing ID order. Native mixed-node levels
match exactly, including extended XOR/mux encodings. Tests cover exact groups,
ignored native control edges, dangling cones, empty graphs and allocation caps.
Remaining traversal/frontier-size algorithms and fGiaSimple mode remain
unported. Full race tests and vet pass.

`DepthFirstCrossCutOrder` and `CrossCut` now port
`Gia_ManDfsForCrossCut` and `Gia_ManCrossCut`. Output cones are visited in
fanin postorder, with optional reversed output order; constant-driven COs
are skipped. Owned reference counters track values still needed by unvisited
consumers, and the peak is sampled before fanin consumption as in C. The DFS
uses an explicit stack instead of recursion and does not mutate traversal IDs
or Value scratch. Plain-AIG order and size match the pinned immutable fixture in
both output orders. Go additionally handles internal constants, single-edge
buffers and native mux control edges safely; these extended-node metrics are
not claimed as parity with C's plain-AIG assumptions. Tests cover shared cones,
constant outputs, extended nodes, source immutability, allocation limits and a
10,000-node cone without recursion. Other GIA traversal APIs remain unported.

`GiaManager.ReverseLevelOrder` now ports `Gia_ManOrderReverse`. It returns
all objects in descending forward level, with COs first within each level and
remaining objects in increasing ID order. The pinned native mixed-node graph
matches exactly. Source scratch is preserved and existing object caps bound
allocation. Empty graphs and zero-level buffer chains have explicit tests.
C's description calls this reverse topological order, but same-level buffers
can follow their drivers; the Go API names the actual ordering and does not
promise strict topology. The C assertion requiring positive levels for every
AND-like node is relaxed for valid zero-delay buffer graphs. Full race tests
and vet pass. Remaining cone/DFS collection APIs are still unported.

`GiaManager.CollectSequential` now ports `Gia_ManCollectSeq` and its recursive
helper using an explicit stack. Constant zero is always first; selected output
cones are collected in fanin postorder. Encountered register outputs append
their corresponding register inputs to a growing root queue, preserving C's
order and terminating on register cycles. Duplicate selections produce no
repeated objects; nil selects no outputs. Native mux control edges are included
explicitly, unlike the C routine's plain-AIG traversal. Owned visitation state
replaces manager traversal-ID writes. Four pinned native output selections
match exactly, covering cycles, multiple roots and duplicates. Additional tests
cover constant-only empty selection, invalid output indices, allocation caps,
source immutability and transitive native mux-control register closure. Full
race tests and vet pass. Other DFS/TFI collection variants remain unported.

`GiaManager.CollectTransitiveFanins` now ports `Gia_ManCollectTfi` and its
helper with an explicit DFS stack. Returned gate IDs include gate roots,
exclude CI/constant boundaries, preserve fanin postorder and contain shared
nodes only once. CO wrappers and native mux control cones are handled safely;
C's original routine assumes CI/AND roots. Owned visitation storage preserves
source scratch. Four pinned native root selections match exactly. Tests cover
shared cones, duplicate roots, CI/constant boundaries, CO wrappers, extended
nodes, invalid roots, resource caps and a 10,000-node cone. Full race tests and
vet pass. Transitive-fanout collection and remaining DFS variants are unported.

`GiaManager.StaticFanouts` now ports `Gia_ManStaticFanoutStart` into an owned
immutable offset/target snapshot. Lists preserve edge multiplicity and
increasing consumer-ID insertion order, including native mux control edges;
Go lifetime replaces explicit fanout-stop cleanup. Snapshot queries return
owned slices and do not change when the source graph grows. Object and edge
caps bound allocation. `CollectTransitiveFanouts` ports `Gia_ManCollectTfo`
using that index and iterative postorder. COs terminate traversal and are
omitted; CI/constant roots are safely accepted beyond C's gate-root premise.
Native lists and three postorders match exactly. Tests cover shared cones,
duplicate roots, data/control edges, CO boundaries, native mux control uses, snapshot/result
ownership, source immutability, invalid IDs and resource limits. Full race
tests and vet pass. Dynamic/LUT fanout variants and other DFS APIs remain
unported.

`GiaManager.CollectInputs` now ports `Gia_ManCollectCis`, returning CI IDs in
first-encounter DFS order. Constants are skipped, CO wrappers are accepted,
and register outputs are support boundaries; this is structural rather than
Boolean functional support. Four native root selections match exactly.
`CollectConeAnds` ports `Gia_ManCollectAnds` with explicit leaf boundaries,
excluding each leaf and its fanin cone. Existing TFI collection delegates to
this implementation without leaves. Both routines use iterative stacks and
fresh per-call visitation, replacing C's implicit caller-managed traversal
state. Native mux control edges participate. Tests cover shared cones,
multiple roots, constant/CI boundaries, leaf cuts, repeated-call independence,
invalid roots/leaves, allocation caps and extended nodes. Full race tests and
vet pass. Other DFS collection variants remain incomplete.

`GiaManager.CollectNodesAndInputs` now ports `Gia_ManCollectNodesCis`, returning
combined CI/gate postorder while excluding constant and CO wrappers. Four
native root selections match exactly. It shares the leaf-aware iterative
traversal used by TFI/AND-cone collection; per-call visitation prevents scratch
state leakage. `AndObjects` ports `Gia_ManCollectAndsAll`, returning every
AND-like ID in order, including dangling buffers/native XORs/muxes. Tests cover
native ordering, shared cones, constant and empty selections, CO wrappers,
dangling/extended nodes, invalid roots and allocation caps. Full race tests and
vet pass with the shared-traversal refactor. Remaining DFS APIs and the full
ABC source tree remain in scope and unfinished.

`GiaManager.Mffc` now ports `Gia_NodeMffcSizeSupp` and its reference/collection
helpers. It returns cone size, DFS-ordered boundary IDs and reference deltas
(remaining minus original, matching C after restoration). Dereferencing uses
owned counters; no graph mutation or restore pass is necessary. Shared fanins
are admitted only when all their incoming contributions disappear. Buffers,
constants and native mux controls use explicit graph edges beyond the C plain
AIG premise. The immutable compatibility dataset covers both shared-cone cases
exactly, before/after an external fanout. Tests cover boundary
size/order/deltas, source immutability, extended nodes, invalid roots and caps.
Full race tests and vet pass. Dominators and remaining MFFC APIs are unported.

`GiaManager.ClosestDominators` now ports `Gia_ManComputeDoms` and
`Gia_ManAddDom`. Owned parent IDs form the closest fanout-dominator forest:
COs point to themselves, incompatible fanout-root intersections point to the
node itself, and objects without consumers retain InvalidGiaID. The native
immutable dataset covers reconvergent graphs with one and two output sinks.
Tests independently check closest chain links, reconvergence, multi-output
splits, dangling branches, native mux-control edges, empty graphs, allocation
caps and source immutability. Missing dangling parents terminate safely rather
than following C's negative parent index; extended nodes use actual edges.
Full race tests and vet pass. Remaining MFFC/statistical APIs and the broader
ABC port remain unfinished.

`GiaManager.DeriveDominatedCone` now ports `Gia_ManDomDerive` and its helper.
An explicit support cut becomes the standalone graph's CIs; the first zero to
five support variables are fixed in assignment order to produce cofactor
outputs. All support CIs remain present as in C, including fixed/unused ones.
Source names/specification are retained and the result is cleaned. Owned
literal mappings replace source Value/traversal mutations. Duplicate, invalid
or incomplete support cuts return checked errors; source-mapping/output graph
allocation is bounded. Buffers, native XORs/muxes and constants are interpreted
explicitly. Pinned native output truth masks match zero-, one- and two-variable
cofactoring; independent exhaustive checks cover all remaining inputs and mux
control cofactors. Tests also cover source immutability and invalid bounds.
Full race tests and vet pass. Experimental dump/collapse drivers and the rest
of ABC remain unfinished.

Cone-analysis fuzz validation now uses constructor-recorded edge ledgers
independent of packed fanin decoding, manager reference counts and static
fanout indexing. A fixed-point consumer-closure oracle checks MFFC membership,
boundary membership and reference deltas. A separate postdominator-set
intersection oracle checks closest dominators without parent climbing. An
independent Boolean evaluator checks every cofactor output on every
four-input assignment, including generated AND, buffer, native XOR and native
mux nodes. The MFFC/cofactor campaign passed 600,133 executions in 30 seconds;
a second campaign including dominators passed 540,560 in 30 seconds. A
five-support-variable regression verifies all 32 cofactor outputs and checks
that an insufficient output-graph allocation cap returns no partial graph.
Full race tests and vet pass. These checks strengthen existing partial source
coverage; they do not establish completion of the full port.

`GiaManager.SupportSize`, `ConeSize` and combined `ConeStatistics` now port
`Gia_ManSuppSize`/`Gia_ManSuppSizeOne` and `Gia_ManConeSize`. A single iterative
union traversal counts distinct CIs and AND-like nodes without materializing
postorder output lists. Constants/CO wrappers are excluded, register outputs
remain CI boundaries, and native mux controls participate. Native cardinalities
match four shared-cone selections, including repeated roots and CI roots.
Tests cover empty/constant selections, shared gate/CO roots, invalid IDs and
allocation caps. The independent cone fuzz oracle now also checks counts from
constructor-recorded edges. Remaining DFS diagnostic/performance drivers and
the full ABC port remain unfinished.

The extended independent cone-analysis fuzz campaign passed 518,411
executions in 30 seconds. Full race tests and vet pass with cone statistics.

`GiaManager.DanglingRoots` and `DanglingRootCount` now port the root-membership
analysis of `Gia_ManGetDangling`, `Gia_ManMarkDangling` and
`Gia_ManHasDangling`. These identify unconsumed AND-like roots, not the full
unreachable-node set; the C Boolean-looking function actually returns a count.
Owned marks preserve caller graph marks instead of exposing C's scratch side
effects. Native mux controls are included consistently, correcting the list/
mark variants' omission while agreeing with C's mux-aware count on supported
graphs. The pinned plain-graph root list/count matches exactly. Tests cover
buffers, control-only uses, mark preservation, empty graphs and object caps.
The independent constructor-edge fuzz oracle now also checks dangling-root
membership. Full race tests and vet pass.

The expanded cone-analysis fuzz campaign passed 246,693 executions in
15 seconds with dangling-root checks enabled.

`GiaManager.LevelsWithInputs` now ports `Gia_ManSetLevels`, accepting one
arrival level per CI (including register outputs), with nil retaining the
zero-arrival default. Isolated CI levels contribute to the maximum. Forward
level addition checks uint32 overflow and returns ErrLimit instead of wrapping.
The mixed-node pinned native seeded levels match exactly. `AverageOutputLevel`
ports `Gia_ManLevelAve` across every CO, including register inputs, using uint64
accumulation and float64 output instead of C's signed sum/float32 return. No
outputs returns ErrDimension rather than a division-by-zero result. Tests cover
seeded native levels, mean depth, explicit-empty/wrong input dimensions,
isolated CIs and both exact uint32 maximum and overflow rejection. Existing
zero-arrival levelization tests still pass. Full race tests and vet pass; other
structural metrics and the full port remain incomplete.

`WeightedReverseLevels`, `RequiredLevels` and `LevelSlacks` now port
`Gia_ManReverseLevel`, `Gia_ManRequiredLevel` and `Gia_ManComputeSlacks`.
Weighted reverse traversal starts at every gate, including dangling gates,
without extra CO distance: ANDs add one, native XORs/muxes two, buffers zero;
mux control edges participate. Required levels subtract reverse distance from
the maximum at CIs; slack subtracts forward and reverse distances from maximum
forward depth. Signed int64 results preserve constant-only negative required
levels without unsigned underflow. This remains distinct from the earlier
unweighted reverse metric. All three vectors match the pinned native mixed-node
fixture. Independent constructor-ledger fuzzing now checks weighted path
lengths and derived values and passed 216,248 executions in 15 seconds.
Tests cover constant-only cones, signed results and allocation caps. Full race
tests and vet pass. Other utilities and the broader source port remain incomplete.

`GiaManager.DetectSequentialSignals` now ports `Gia_ManDetectSeqSignals`
with the active `Gia_CollectSuper` behavior: unique immediate fanin IDs, not
recursive flattening (that call is disabled in C). Returned per-object set,
reset and enable candidate counts are owned data; source marks/references are
preserved. Candidate classification follows RI driver phase, shared immediate
fanin frequencies and phase-insensitive mux feedback to the corresponding RO.
These are structural diagnostics, not certified set/reset/hold functions.
Native summary counts match shared reset-candidate and common-enable feedback
models. Tests also check detailed frequencies, source immutability, invalid
limits and zero-register rejection. Native mux metadata is supported by the
existing recognizer; diagnostic text printing and remaining detector variants
are unported. Full race tests and vet pass.

`SequentialSignalsAboveFanout` now ports
`Gia_ManDetectSeqSignalsWithFanout` by reusing the checked structural detector.
Candidates must have counts strictly greater than the threshold. Selection
preserves set, reset, enable category order with increasing IDs within each
category and first-occurrence deduplication across categories; constant zero
is omitted. Six native lists (three thresholds on two sequential models)
match exactly. A dedicated regression verifies cross-category duplicate
removal and non-global-ID ordering. Negative thresholds return checked errors.
Source storage remains unchanged. Full race tests and vet pass. These are
structural candidates; remaining detector/transfer variants and the full ABC
port remain unfinished.

`TransferGiaFrameSignals` now ports `Gia_ManTransferFrames` by composing the
existing unrolling object map with an explicit frame-to-target map. Source
CI/gate candidates are visited in signal-major/frame-major order, target IDs
are deduplicated on first occurrence, and literal phases are discarded as in C.
Missing/invalid copies and constant-zero target literals are skipped;
constant-one maps retain ID zero. Checked dimensions and target IDs replace
scratch Value assumptions; no graph scratch is changed. Numeric maps cannot
detect accidentally supplying a different same-sized source graph, so callers
must retain source ownership. The pinned initialized three-frame fixture
matches exact IDs/order. Tests cover duplicates, constants, absent copies,
invalid source/target IDs, mapping dimensions and result caps. Full race tests
and vet pass. Remaining unroll/cofactor/transfer variants and the full port
remain incomplete.

`GiaManager.CofactorRebuild` now ports `Gia_ManDupCofInt`: build negative and
positive pivot cofactors and reconstruct each output through the original
pivot function. The result remains functionally equivalent; it does not fix
the pivot in the returned graph. Explicit owned maps retain positive-cofactor
signals and new CO IDs without source Value mutations. Intermediate nodes
remain as in C, and register/constraint/name/spec metadata is preserved.
Object/mapping caps and cancellation bound construction. Native XORs/muxes,
buffers and constant-zero pivots are handled safely (C asserts on a zero
pivot function). Five pinned native output truth tables match exhaustive Go
checks at CI/internal pivots. Tests cover mappings, source immutability,
metadata, native muxes, constant pivots, invalid pivots, bounds and cancellation.
Full race tests and vet pass. Multi-pivot orchestration and the initialized
unroll/cofactor pipeline remain unported; the full objective remains unfinished.

`GiaManager.CofactorRebuildAll` now ports `Gia_ManDupCofAllInt` and its
`Gia_ManTransfer` mapping composition. Each pass selects the highest remaining
object ID, performs Shannon reconstruction, cleans the graph and transfers
remaining candidates through explicit maps with first-occurrence deduplication.
Zero/missing copies disappear; constant-ID candidates are skipped rather than
being passed back as invalid pivots. Empty selection returns an independent
clone. C's 200-signal input cap is retained, with an explicit pass cap plus
object/mapping limits and cancellation. Cleanup-map allocation is also bounded.
Four native output truth-table fixtures match exhaustive Go checks. Tests
cover source immutability, dimensions, pass/resource caps, invalid pivots,
empty selection and cancellation. Remaining multi-pivot variants and the
initialized unroll/cofactor orchestration are not yet implemented.

Multi-pivot reconstruction is included in the independent constructor-ledger
fuzz oracle, checking both outputs on every four-input assignment. The expanded
campaign passed 120,614 executions in 15 seconds. Full race tests and vet pass.

`GiaManager.UnrollAndCofactor` now ports `Gia_ManUnrollAndCofactor`, reusing
the initialized unrolling path for `Gia_ManUnrollInit`. Cleaned object maps
compose with structural threshold selection before bounded multi-pivot Shannon
reconstruction. The function preserves the unrolled Boolean behavior, not a
fixed cofactor. C's 200 selected-frame-signal cap applies after mapping. Graph,
mapping and pass caps plus cancellation span the pipeline. One/unknown reset
values are explicit extensions; unknown initial CIs retain the unroller's
layout. Constraints remain frame-interleaved outputs and global constraint
count stays zero. Native two-frame output masks agree with/without candidates.
Tests replay the machine independently, check all inputs, one/unknown reset,
constraint layout, source immutability, metadata and stage failures. Another
64 arbitrary truth-table machines are checked against independent two-frame
simulation, with known-zero/one initialization and positive constraint outputs.
Full race tests and vet pass. Other cofactor/enable transformations and the
full source port remain unfinished.

`GiaManager.CofactorCandidates` and `CofactorRebuildByFanout` now port
`Gia_ManCofVars` and `Gia_ManDupCofAll`. Selection is inclusive (`refs >= limit`)
and visits CI/gate IDs in increasing order, matching C's executable loop
rather than its reverse-topological comment. Threshold zero includes unused
CIs and dangling gate roots; constants/COs are excluded. Native mux controls
participate in reference counts. The convenience entry point preserves bounded
multi-pivot behavior, source immutability and cancellation. Four pinned native
candidate lists and reconstruction truth masks match exactly. Tests cover
zero/inclusive thresholds, dangling gates, equivalent output truth, independent
cloning, source scratch, invalid thresholds, resource caps and cancellation.
Independent constructor-edge fuzzing also checks candidate membership/order.
Other cofactor variants and the full ABC port remain unfinished.

Full race tests and vet pass with fanout selection. The expanded independent
cone-analysis fuzz campaign passed 120,272 executions in 15 seconds.

`GiaManager.CofactorPrimaryInput` now ports `Gia_ManDupCofactorVar` and its
helper. It fixes one primary input while retaining every CI, including the
unused pivot, and preserves all COs, register/constraint layout, name/spec.
Register outputs cannot be selected as PIs. Reachable cone copying and final
cleanup use owned maps and iterative traversal; native XOR/mux/buffer nodes
retain their meaning. Object/mapping caps and cancellation cover construction.
Six pinned native truth tables match exhaustive input checks. Tests also cover
source immutability, metadata, sequential native muxes, fixed-input independence,
invalid PI indices, register rejection, resource caps and cancellation.
Independent cone fuzzing now checks both outputs against constructor-ledger
truth tables with the selected input bit forced. Other duplication/cofactor
variants and the full port remain unfinished.

Full race tests and vet pass with fixed-PI cofactors. Expanded independent
cone fuzzing passed 109,248 executions in 15 seconds.

`GiaMuxCofactors` now ports `Gia_ManDupMux` for single-output combinational
graphs with matching CI layouts. CI[index] selects the positive/negative output;
all CIs are retained and name/spec comes from the positive cofactor. Constraint
counts must agree and are preserved, extending C's unchecked metadata handling.
Owned reachable-cone maps and a shared checked gate-copy helper replace source
Value scratch. Fixed-PI cofactoring now shares that helper for buffers and
native XOR/mux semantics. Object and combined mapping caps, cancellation and
final cleanup are checked. Three native truth masks match exhaustive Shannon
reconstruction tests, including source immutability and metadata. Tests also
cover invalid controls, incompatible graph shapes, constraints and resource
failures. Full race tests and vet pass. Other duplication variants and the
full source port remain unfinished.

The independent cone/cofactor fuzz campaign passed 109,941 executions in
15 seconds after sharing the checked gate-copy helper.

`GiaManager.CofactorObjectGuarded` now ports `Gia_ManDupCofactorObj` for
combinational CI/gate pivots. Outputs equal the original output AND the original
pivot matching the requested value; this is deliberately different from an
unguarded fixed-PI cofactor. All CIs and constraint/name/spec metadata remain.
Owned cone maps and checked gate copying preserve native XOR/mux/buffer
semantics and source scratch. Outputs preceding the pivot in ID order are
handled safely, rather than using C's still-unset guard. Ten native truth masks
match exhaustive checks. Tests cover guarded semantics, source immutability,
late pivots, metadata, invalid/sequential pivots, resource caps and cancellation.
The constructor-ledger fuzz oracle independently evaluates the original pivot
and both guarded outputs on every four-input assignment. Remaining duplication
variants and the full source port remain unfinished.

Full race tests and vet pass with guarded object cofactors. The expanded
independent cone-analysis fuzz campaign passed 96,333 executions in 15 seconds.
Pinned fixture hashes, coverage destinations, package headers and permanent
test-vector layout were also verified.

`GiaManager.SelectOutputs` and `SelectOutputMask` now port
`Gia_ManDupSelectedOutputs` and `Gia_ManDupOutputVec`. Combinational selections
retain every CI and copy only reachable gates in DFS postorder without hashing;
requested output order and bounded duplicates are preserved. Shared raw gate
copying preserves buffers/native XOR/mux kinds. Source scratch remains unchanged.
Constraint roles are retained as a suffix; invalid ordinary-after-constraint
reorderings return errors instead of silently losing metadata. Four native
object counts and truth-mask lists match exactly, including empty selection.
Tests cover independent output truth, masks, duplicates, constants, native node
kinds, source immutability, invalid indices/masks, constraints, allocation caps
and cancellation. Full race tests and vet pass. Sequential cone extraction and
other output/duplication variants remain unported.

`GiaManager.ExtractOutputCones` now ports `Gia_ManDupCones` and its helper.
Selected PO cones follow transitive register-input closure. Registers retain
C's first-discovery order; optional PI trimming keeps used PIs in original
terminal order. The returned source PI/register index maps make reordering
explicit so callers can project inputs and latch initialization. Raw gate
copying preserves extended node kinds, source scratch and name/spec metadata.
Constraint outputs retain their role when selected as a suffix. Eight pinned
native graph shapes and driver lists match exactly, including register cycles,
PI trimming, reordered roots and empty selections. Exhaustive projected-CI
checks compare every selected output and next-state bit with the source.
Additional tests cover native mux controls, duplicate outputs, constraint
ordering, invalid indices, resource caps and cancellation. Full race tests and
vet pass. Graph extraction does not copy latch reset values: AIGER callers
must use the explicit source-register map. Other duplication variants and the
full port remain incomplete.

`GiaManager.ExtractGateCones` now ports `Gia_ManDupAndCones` for combinational
graphs. Gate, CI and constant roots become outputs in requested order; optional
PI trimming keeps original CI order and exposes source PI indices. Arbitrary
root outputs inherit no constraint roles. Native node kinds and source scratch
are preserved. The C variant disables register creation/finalization, so Go
rejects sequential graphs rather than producing incomplete mappings. Ten pinned
native shapes/driver lists match both trimming modes, including constant and
empty roots. Exhaustive projected-input checks cover all root functions. Tests
also cover invalid/CO roots, source immutability, metadata, constraints,
resource bounds and cancellation. The independent constructor-ledger fuzz
oracle checks extracted root/CI values with trimmed extended-node cones.
Other bounded-cone variants and the full source port remain unfinished.

Full race tests and vet pass with gate-root extraction. Expanded independent
cone fuzzing passed 88,198 executions in 15 seconds.

`ExtractLevelLimitedCones` and `ExtractDepthLimitedCones` now port
`Gia_ManDupAndConesLimit`/`Limit2` and their helpers. Boundary objects become
CIs in first-encounter order; source boundary IDs are returned for projection.
Absolute cuts stop below the forward-level threshold; depth cuts stop when
first-visit remaining depth reaches zero. C's object-only memoization is
preserved: shared nodes keep their first expansion, so later roots can reuse
deeper logic and root order affects the cut. Both variants are iterative,
leave constants constant and return combinational graphs; register outputs
are boundary CIs rather than retained state. Native node kinds/control edges
are handled explicitly. Sixteen native boundary/shape/driver records match
exactly. Exhaustive projection tests cover both root orders/bounds, constants,
register boundaries and native muxes. Invalid roots/depths, resource caps,
cancellation and source immutability are checked. Independent constructor
truth tables now validate bounded-cone projection during fuzzing. Remaining
duplication modes and the full ABC port remain unfinished.

Full race tests and vet pass with bounded cone cuts. Expanded independent
cone fuzzing passed 83,017 executions in 15 seconds.

`NewGiaOneHot` now ports `Gia_ManOneHot`: ignored leading CIs, balanced
power-of-two-padded overlap detection, and a final no-selected-bit output.
All outputs are false exactly when one selected input is true. These are
violation outputs, not positive environment constraints; no constraint role is
assigned. Zero-variable inputs are rejected. Seven pinned native object counts
and exact CO-driver lists match widths 1..7. Exhaustive assignments include
ignored inputs and non-power-of-two widths. Tests check exact object capacity,
scratch limits, invalid dimensions and cancellation with no partial results.
Full race tests and vet pass; dedicated fuzzing passed 584,285 executions in
15 seconds for widths up to 12. Fixture hashes and partial source
coverage are recorded. `Gia_ManDupOneHot` integration, other duplication modes
and the full ABC source port remain unfinished.

`GiaManager.DuplicateLevelized` now ports `Gia_ManDupLevelized`. Every CI,
dangling gate and CO is retained; gates follow increasing weighted level and
source ID. An owned source-object-to-target-literal map includes CO objects.
Raw buffer/XOR/mux kinds and name/spec/register/constraint metadata survive.
The C ordinary-AND path does not safely handle zero-level buffers; Go extends
it with stable ID ordering, preserving same-level dependencies. A pinned native
fixture checks every object's fanin literals after reordering an ordinary AIG.
Exhaustive truth checks cover ordinary and extended/sequential graphs, source
immutability, exact capacity, empty graphs, invalid limits and cancellation.
The independent constructor-ledger fuzz oracle now checks both duplicated
outputs on all four-input assignments. Full race tests and vet pass. Other
duplication variants and the full source port remain incomplete.

Expanded cone-analysis fuzzing with levelized duplication passed 154,870
executions in 15 seconds. The first sandboxed attempt stopped because Go's
fuzz cache was read-only; rerunning with cache write access passed. Native
fixture hashes and all Go package-first headers were verified.

`GiaManager.DuplicateFromLists` now ports `Gia_ManDupFromVecs`. Explicit CI,
gate and CO lists determine object order; all IDs must be unique and correctly
typed, and every gate/CO dependency must already have a mapping. The returned
owned map marks omitted objects invalid. The caller chooses register suffixes;
constraint roles are not inherited, matching this C API. Name/spec and raw
buffer/XOR/mux kinds are retained without changing source scratch. A native
fixture checks every fanin after independent CI, gate and CO permutations.
Exhaustive Boolean checks independently verify reordered inputs and outputs.
Validation covers omitted dependencies, duplicate/wrong/out-of-range IDs,
invalid register counts, empty selections, limits and cancellation. Independent
constructor-ledger fuzzing now checks input-reordered list duplication of mixed
node kinds. Remaining duplication APIs and the full ABC port remain unfinished.

Full race tests and vet pass with explicit-list duplication. The expanded
independent cone fuzz campaign passed 142,437 executions in 15 seconds.

`NodeSupports` and `OutputSupports` now port `Gia_ManCreateNodeSupps` and
`Gia_ManCreateCoSupps`. Results contain sorted CI terminal indices, including
register outputs; these are structural supports, not minimized Boolean
supports. Constants have empty support, phases do not change support, and COs
inherit their drivers. Go additionally handles buffers/native XOR/mux control
edges. Immutable internal sharing avoids copying buffer/CO sets; returned
sets are independently owned even for repeated IDs. Object/query and retained
entry limits bound allocation. Native node/output fixtures match; tests cover
ownership, source immutability, extended gates, invalid/nil inputs and limits.
A nil-wrapper guard caught by the new test was fixed before completion. Full
race tests and vet pass. Expanded independent constructor-ledger fuzzing passed
138,575 executions in 15 seconds. Constraint extraction was still pending at this checkpoint; the next entry
records its positive-polarity implementation.
Other support/duplication routines and the full ABC port remain incomplete.

`ExtractConjunctiveConstraint` now ports `Gia_ManDupWithConstr` and its
conjunction collector. Iterative DFS flattens positive ordinary ANDs, retaining
unique phased boundaries in first encounter order. Highest-level boundary
selection uses C's first-tie rule. The main output and positive assumption
conjoin to the original output. C emits the assumption's complement as a
violation output; Go explicitly adapts this polarity for its verification
APIs and marks the positive output as one constraint. Extended node kinds are
opaque conjunction boundaries and are copied semantically; all CIs/name/spec
survive cleanup. Complemented/constant/CI/extended roots, sequential/multiple
outputs and pre-existing constraints are rejected instead of following unsafe
C assertions. Source scratch is unchanged. Native main/violation truth masks
and cleaned object counts match after polarity adaptation. Exhaustive truth,
metadata, limits, cancellation and invalid-root tests pass. Dedicated fuzzing
passed 597,508 executions in 15 seconds; full race tests and vet pass. Other
duplication modes and the full ABC source port remain incomplete.

`RankOutputSupports` now ports `Gia_ManCoSuppSizeMax`, `Gia_ManCoLargestSupp`
and `Gia_ManSortCoBySuppSize`. It computes checked structural supports, retains
C's minimum maximum of one, returns the first largest-support index (or -1
for no outputs), and orders outputs by decreasing support size. The C
`Abc_QuickSort3` entry point actually delegates to its two-way sorter; Go
preserves its nonstable tie order, selection-sort base case and partition
swaps. Iterating smaller partitions first bounds auxiliary stack growth.
Twenty-five native records cover 0..24 outputs and repeated sizes across the
base-case boundary. Fuzzing independently checks ordering, permutation,
maximum and first-largest properties; 802,958 executions passed in 15 seconds.
Full race tests and vet pass; fixture hash and package-first headers were
verified. Support-remapped structural reduction and the remaining full ABC
source port remain incomplete.

`ReduceBySupport` now ports `Gia_ManDupStrashReduce`. Each combinational CO
cone is hashed onto target CIs ordered by its sorted source support. Returned
support projections and source-to-reduced-output indices make this per-cone
renaming explicit: the transformation does not preserve a single whole-network
input assignment. Optional phased-literal deduplication uses first occurrence
order; a lookup-only map never determines iteration. All source scratch stays
unchanged and name/spec survive cleanup. At least one target CI is retained,
as in C. Actual logic is copied for size-zero/one structural cones rather than
using C's direct-literal shortcut, preserving complemented buffers and extended
node semantics. Sequential and constraint-bearing graphs are rejected pending
explicit caller policy. Native merged/unmerged graph counts and output maps
match. Exhaustive projected-input truth tests cover ordinary and extended
cones, complements, constants, metadata, source immutability, allocation caps
and cancellation. Independent constructor-ledger fuzzing checks each source
output under its own projection; 119,568 executions passed in 15 seconds.
Full race tests and vet pass; fixture hash and package headers were verified.
Other structural-isomorphism/reduction variants and the full ABC source port
remain incomplete.

`DuplicateWithGateOutputs` now ports `Gia_ManDupOuts`. All source logic,
including dangling gates, is raw-copied with native buffer/XOR/mux kinds;
each AND-like object's regular value becomes an additional ordinary PO in
source ID order. Existing ordinary POs remain first; constraints move behind
new gate POs, preserving their role, and RIs remain last. C instead drops the
constraint count. Name/spec/register layout and owned source-object maps are
preserved without source scratch mutation. The pinned native sequential
fixture checks total objects and exact output driver literals. Exhaustive
checks cover ordinary/extended node values, constraints and next-state order;
limits, exact capacity, cancellation, empty graphs and source immutability
are tested. Full race tests and vet pass. Independent fuzzing now compares
every exposed gate against constructor-ledger truth tables. Other duplication
variants and the full ABC source port remain incomplete.

Expanded independent gate/cone fuzzing passed 112,523 executions in 15 seconds.

`SliceBySupport` now ports `Gia_ManDupSliced`. Every gate whose structural CI
support fits the requested limit is retained, including dangling logic.
Retained gates feeding omitted gates or source COs become outputs in source
ID order; direct CI/constant outputs are not copied. Returned regular source
root IDs describe output meaning. This is a combinational slice collection:
all CIs survive, while registers/constraints are not assigned. Raw extended
node kinds and name/spec survive; mux control edges participate in support
and boundary marking, extending C's two-data-fanin path. Five pinned native
object-count/driver fixtures match limits 0..4. Exhaustive tests check slice
values, boundary order, extended gates, source immutability, metadata,
allocation limits and cancellation. Independent constructor-ledger fuzzing
checks boundary selection and every slice output value; 107,215 executions
passed in 15 seconds. Full race tests and vet pass. Remaining duplication
variants and the full ABC source port remain incomplete.

`WithOneHotAssumptions` now ports `Gia_ManDupOneHot` integration. It selects
all PIs for combinational graphs or all register outputs for sequential graphs,
shares the generator's CIs and adds positive exactly-one assumptions. Source
ordinary POs and constraints retain order; new assumptions precede RIs so
next-state layout survives. This explicitly adapts C's violation polarity and
physical append-after-RI behavior. All raw source node kinds, name/spec and
existing constraint/register metadata survive without mutating source scratch.
A native sequential fixture pins physical output truth masks and verifies the
polarity/order adaptation. Exhaustive tests cover register counts 0..3,
existing constraints, native XOR/buffer/mux nodes, original outputs and RI
functions. Limits, cancellation and zero-variable rejection are tested.
Dedicated fuzzing passed 556,338 executions in 15 seconds for widths 1..7 and
all register counts; full race tests and vet pass. Earlier one-hot integration
pending entries are superseded by this implementation. Remaining duplication
and verification algorithms and the full ABC source port remain incomplete.

`ComposeGiaSharedInputs` now ports `Gia_ManDupAppendNew`. Equal-count PIs
share target inputs by index; register outputs/inputs retain distinct first
then second banks. Ordinary POs precede grouped source constraint suffixes and
both RI banks, preserving constraint roles that C drops. Name/spec come from
the first source; owned maps for both sources include CO objects. All logic
is hashed with explicit extended-node lowering and source scratch unchanged.
A native sequential fixture matches object count and exact CO driver literals,
including shared combinational logic. Exhaustive independent checks cover
native XOR/mux semantics, source metadata, distinct states, constraints,
next-state functions, maps, mismatched PIs, caps and cancellation. Dedicated
Boolean-formula fuzzing passed 520,191 executions in 15 seconds. Full race
tests and vet pass; fixture hash and package-first headers were verified.
Other composition/invariant-miter variants and the full source port remain
incomplete.

`GiaInvariantMiter` now ports `Gia_ManDupInvMiter`: a combinational single
output I(state) AND NOT I(nextState), with inputs in transition CI order.
It checks unconstrained induction only; transition properties/constraints
are ignored as in C. Initiation and property implication remain separate
obligations, so an UNSAT miter alone is not a complete safety certificate.
Explicit invariant constant mapping supports constant roots without C scratch
dependence. Extended transition/invariant nodes are lowered semantically;
source scratch, CI order and name/spec are preserved. A immutable fixture pins
cleaned object count and full truth mask. Exhaustive formula tests cover
extended transitions, constant invariants and ignored constraint outputs.
Invalid/nil inputs, dimensions, limits, cancellation and source immutability
are checked. Dedicated independent-formula fuzzing passed 572,410 executions
in 15 seconds. Full race tests and vet pass. Other composition/miter variants
and the full ABC source port remain incomplete.

`AppendPropertyCones` now ports `Gia_ManDupAppendCones`. Single-output
combinational cone inputs map to all source CIs or just register outputs.
Repeated cones and constants work; hashing lowers extended gates and cleanup
retains CI/register order and name/spec. Output order is original ordinary
POs, added cone POs, preserved constraints, RIs; unlike C, constraint roles
survive. Summed source object spans bound mapping storage. Invalid cone roles,
dimensions/nil inputs, resource caps and cancellation return no partial graph.
Two native cleaned-count/driver fixtures cover both input-mapping modes.
Exhaustive checks cover repeated/constant cones, native XOR logic, original
properties, constraints, next-state values, source metadata and immutability.
Dedicated independent-formula fuzzing passed 564,097 executions in 15 seconds.
Full race tests and vet pass; fixture hash and package-first headers were
verified. Remaining graph transformations, broader verification algorithms
and the full ABC source port remain incomplete.

`WithRegisterHoldControl` now ports `Gia_ManDupSelf`. A new first PI selects
original next state when true and current state when false for every register.
Original PIs shift by one; RO/RI ordering and original PO/constraint functions
stay intact. Name/spec and constraint roles survive cleanup; extended gates
are lowered semantically. Combinational/empty graphs still receive the unused
control CI, matching C. Graphs carry no initialization; callers retain reset
lists in unchanged register order. A immutable fixture matches cleaned object
count and exact driver literals. Exhaustive tests cover both control values,
two-register extended logic, constraints, metadata, source immutability,
limits, cancellation and empty graphs. Independent-formula fuzzing passed
502,426 executions in 15 seconds. Full race tests and vet pass; fixture hash
and package-first headers were verified. Remaining transformations and the
full ABC source port remain incomplete.

`ExtractFlopClass` now ports `Gia_ManDupFlopClass` using explicit class labels
rather than source scratch metadata. Selected registers remain state; other
ROs/RIs become independent ordinary inputs/outputs. Returned source CI, CO
and retained-register indices expose input/output projections and reset-list
projection. This is a state abstraction, not whole-machine equivalence without
external assumptions. Raw buffer/XOR/mux kinds, dangling logic and name/spec
survive. Source constraints move after externalized RIs to preserve their
suffix role, which C drops. Three native class fixtures match register counts
and exact driver ordering, including an absent class. Exhaustive source/target
projection tests cover extended gates, metadata, constraints, limits,
cancellation and source immutability. Independent Boolean-formula fuzzing
passed 649,946 executions in 15 seconds. Full race tests and vet pass; fixture
hash and package-first headers were verified. Remaining graph transformations
and the full ABC source port remain incomplete.

`ExistentialInput` and `UniversalFirstOutput` now port `Gia_ManDupExist` and
`Gia_ManDupUniv`. Both retain all original CIs, including the unused quantified
input, and hash negative/positive cofactors with source scratch unchanged.
Existential mode accepts one combinational output and ORs the cofactors.
Universal mode ANDs output zero's cofactors; later outputs are positive
cofactors, matching the source's special behavior. Names make this distinction
explicit. Sequential/constraint-bearing graphs are rejected. Extended node
semantics and name/spec survive cleanup. Six native cases match cleaned
counts and full output truth masks across all three pivot positions.
Exhaustive extended-node tests cover variable independence, later-output
behavior, limits, cancellation and source immutability. Independent-formula
fuzzing passed 491,981 executions in 15 seconds. Full race tests and vet pass;
fixture hash and package-first headers were verified. Remaining quantification
variants and the full ABC source port remain incomplete.

`ExistentialAllOutputs` now ports `Gia_ManDupExist2`, sharing the checked
cofactor kernel with earlier quantification APIs and using reachable DFS gate
order. Every CO independently ORs its two PI cofactors, including constraints
and next-state outputs; all CI/register/constraint roles and name/spec survive.
This is coordinate-wise abstraction: witnesses may differ between outputs,
so it is not existential elimination of the joint transition relation. A
regression explicitly demonstrates independently quantified p and !p becoming
simultaneously true. Native cleaned counts and full sequential CO truth masks
match. Exhaustive extended-node/constraint tests cover source immutability,
unused quantified inputs, empty output sets, PI-only pivots, limits and
cancellation. Independent-formula fuzzing passed 623,449 executions in
15 seconds. Full race tests and vet pass. Other graph transformations and
the full ABC source port remain incomplete.

`FlipRegisters` now ports `Gia_ManDupFlip`. State encoding becomes oldState
XOR phase; all register reads and next-state outputs receive matching phases.
CI/CO ordering, raw extended gate kinds, dangling logic, name/spec and source
constraint roles survive without source scratch mutation. Reset values remain
caller-owned: known resets must be XOR-recoded, unknown resets stay unknown.
Four immutable fixtures match exact driver literals for every two-register mask.
Exhaustive tests verify property/constraint equivalence under state recoding
and correctly recoded next-state bits, including XOR/buffer/mux logic. Limits,
phase dimensions, cancellation and source immutability are checked.
Independent-formula fuzzing passed 679,653 executions in 15 seconds. Full race
tests and vet pass; fixture hash and package-first headers were verified.
Other state transformations and the full ABC source port remain incomplete.

`PermutePrimaryInputs` and `PermuteRegisters` now port `Gia_ManDupPerm` and
`Gia_ManDupPermFlop`. order[i] names the source terminal placed at target i;
validated bijections reject duplicate, missing and invalid indices. Register
permutation moves ROs and matching RIs together; callers must reorder resets
using the same new-to-old order. Raw extended gate kinds, dangling logic,
name/spec and constraint roles survive without source scratch mutation.
Native PI/register three-cycle fixtures pin driver literals and distinguish
permutations from their inverses. Exhaustive tests cover every three-terminal
permutation, all six-input assignments, constraints, next-state projection,
metadata, caps, cancellation and source immutability. Independent-formula
fuzzing passed 598,737 executions in 15 seconds. Full race tests and vet pass;
fixture hash and package-first headers were verified. Other terminal/state
transformations and the full ABC source port remain incomplete.

`SpreadRegisters` and `PermuteRegistersWithGaps` now port
`Gia_ManDupSpreadFlop` and `Gia_ManDupPermFlopGap`. Mask -1 entries create dummy
ROs with constant-zero RIs. Spread consumes each source register in order for
non-gap entries, ignoring their values as C does; gap permutation interprets
non-gap entries as a validated new-to-old bijection. All original POs,
constraint roles, raw extended gates, dangling logic and name/spec survive.
Source scratch stays unchanged. Reset projection and dummy reset choices are
explicit caller responsibilities. Immutable fixtures distinguish both modes and
pin exact CO drivers. Exhaustive extended-node projection tests vary original
and dummy state bits and check next-state order, constraints, metadata,
invalid masks, allocation caps and cancellation. Independent-formula fuzzing
passed 663,312 executions in 15 seconds. Full race tests and vet pass; fixture
hash and package-first headers were verified. Remaining state transformations
and the full ABC source port remain incomplete.

`RetainLastPrimaryInputs` now ports `Gia_ManDupLastPis`. Combinational graphs
keep the requested trailing CIs and raw-copy all gates/COs, including dangling
logic. Any dependency on an omitted CI returns a checked error instead of
using C's unset scratch value. This neither fixes input values nor trims
logic. Constraint roles, extended node kinds and name/spec survive without
source mutation. Three native valid-input cases match counts and exact output
drivers. Exhaustive tests cover trailing-input projection, constraints,
buffer/XOR values, zero/all counts, dangling omitted-input dependencies,
invalid dimensions, allocation caps and cancellation. Independent Boolean
fuzzing passed 656,194 executions in 15 seconds. Full race tests and vet pass;
fixture hash and package-first headers were verified. Other PI-removal modes
and the full ABC source port remain incomplete.

`Replicate` now ports `Gia_ManDupTimes`: independent copies share only constant
zero. Logical CI order groups copy-major PIs before ROs; CO order groups
ordinary POs, preserved constraints, then RIs, with copy-major order in each
bank. Reset lists repeat copy-major. Raw extended gates, dangling logic and
name/spec survive without source mutation. Go uses normalized physical
terminal placement: initial tests caught that C's reordered terminal tables
violate the Go manager invariant, so construction was changed rather than
weakening validation. Immutable fixtures retain CI IDs, physical ANDs and output
drivers; an independent fixture evaluator checks every input assignment
against normalized Go graphs for 1..3 copies. Exhaustive extended-node tests
cover independent state/input banks, constraints, metadata, invalid/overflowing
copy counts, capacity checks, cancellation and empty graphs. The cancellation
regression threshold was adjusted to remain reachable after normalization.
Independent-formula fuzzing passed 662,949 executions in 15 seconds. Full race
tests and vet pass. Other duplication/verification algorithms and the full
ABC source port remain incomplete.

`AbstractRegisters` now ports `Gia_ManDupAbs`. Explicitly ordered source
registers become pseudo-PIs and their RIs are removed; retained registers stay
in original order. Only PO and retained-RI cones are raw-copied in DFS order.
Returned source CI/CO/retained-register indices expose input/output/reset
projections. This is an overapproximation, not an equivalence transform.
Constraints, raw extended node kinds and name/spec survive without source
scratch mutation. Native object/register counts and exact output drivers match
both empty and reordered abstraction lists. Exhaustive extended-node tests
check all projection values, complete externalization, eliminated unused logic,
constraints, invalid mapping indices/duplicates, caps, cancellation and source
immutability. Independent Boolean-formula fuzzing passed 647,792 executions
in 15 seconds. Full race tests and vet pass; fixture hash and package-first
headers were verified. Abstraction refinement orchestration, remaining graph
algorithms and the full ABC source port remain incomplete.

`ComposeGiaSeries` now ports `Gia_ManDupOntop`: bottom COs drive top CIs by
terminal index, yielding bottom CIs and top COs. Both graphs are combinational.
Top constraint roles and bottom name/spec survive; bottom constraints are
rejected because consuming assumptions as data requires explicit caller
policy. Hashing lowers extended gates, cleanup removes unused logic, and
source scratch remains unchanged. A immutable fixture matches cleaned object
count and exact output drivers. Exhaustive tests cover complemented signals,
constants, XOR/buffer/mux nodes, constraints, metadata, caps, cancellation and
source immutability. Independent Boolean-formula fuzzing passed 577,647
executions in 15 seconds. Full race tests and vet pass; fixture hash and
package-first headers were verified. General output trimming still depends
on unported sequential structural sweeping; it remains tracked, along with
other graph algorithms and the full ABC source port.

`WithStateProperties` now ports `Gia_ManDupWithNewPo`. Combinational property
inputs map to transition ROs; new ordinary properties replace original ordinary
POs, while original RIs and CI/register/reset order survive. Property constraints
and original transition constraints are both retained before RIs, extending
C's dropped constraint metadata/original POs. Hashing lowers extended gates,
cleanup removes unused logic, name/spec come from transition, and neither
source's scratch changes. A immutable fixture matches cleaned count and exact
drivers for multiple state properties. Exhaustive tests check property truth,
unchanged transitions, combined constraint roles, extended nodes, metadata,
invalid dimensions/nil inputs, caps, cancellation and source immutability.
Independent-formula fuzzing passed 533,426 executions in 15 seconds. Full race
tests and vet pass; fixture hash and package-first headers were verified.
Remaining transformations and the full ABC source port remain incomplete.

`RemapInputs` now ports `Gia_ManDupDfsCiMap`. Every source CI is retained;
uses are replaced by constants, identity, or phased same/earlier CIs in terminal
order. Substitutions compose, including self-inversion. Later-CI/gate references
are rejected instead of reading unset source scratch. Nil output selection
preserves original CO roles; explicit phased signals define custom outputs,
with enough trailing RIs for the original register count and no implicit
constraint roles. This is substitution rather than an equivalence claim.
Reachable cones are hashed in output DFS order with extended-node lowering;
name/spec and source scratch survive. Two immutable fixtures match object counts
and exact drivers for default/custom outputs. Exhaustive tests cover aliases,
constants, self-phases, custom roles, constraints, source immutability, invalid
references/dimensions, caps and cancellation. Independent-formula fuzzing
passed 647,068 executions in 15 seconds. Full race tests and vet pass; fixture
hash and package-first headers were verified. Remaining graph transformations
and the full ABC source port remain incomplete.

`DuplicateAllOutputPhases` now ports `Gia_ManDupOnsetOffset`. Every source CO,
including RIs, becomes its driver followed by its complement. The original
register count remains, so only the final RegisterCount paired outputs become
RIs. This exactly matches C's raw transformation, changes transition functions,
and intentionally drops source constraint roles; the API makes no state
equivalence claim. Raw extended gates, dangling logic, CI order and name/spec
survive without source mutation. A native sequential fixture matches object
count, register count and exact CO drivers. Exhaustive tests cover every
phase pair, changed RI assignment, constraint removal, extended nodes, metadata,
exact capacity, limits, cancellation and empty graphs. Independent Boolean
fuzzing passed 657,200 executions in 15 seconds. Full race tests and vet pass;
fixture hash and package-first headers were verified. Remaining transformations
and the full ABC source port remain incomplete.

`ExtractInterleavedBlock` now ports `Gia_ManDupBlock`. It retains CI/CO indices
0,stride,2*stride,..., substitutes zero for other CIs, hashes extended gates
and cleans unused logic. Both terminal counts must be divisible by stride.
Register count uses C's floor division and constraint roles are discarded;
this is structural selection/substitution, not machine equivalence. Name/spec
and source scratch survive. Three immutable fixtures match cleaned counts,
register partitions and exact output drivers. Exhaustive projection tests
cover extended nodes, zero substitution, rounded-down register count,
constraint removal, metadata, invalid strides, caps, cancellation and source
immutability. Independent Boolean-formula fuzzing passed 584,685 executions
in 15 seconds. Full race tests and vet pass; fixture hash and package-first
headers were verified. Remaining transformations and the full ABC source
port remain incomplete.

`TrimDirectInputOutputs` now ports `Gia_ManDupTrimmed2`. POs driven by
constants or original PIs are removed independent of phase. PI references
from all POs are discounted before trimming unused PIs; ROs/RIs and every
raw gate survive. Dummy unused PI/zero PO fallbacks ensure at least one of
each. Returned terminal projections use -1 for dummies. Name/spec and extended
node kinds survive, source scratch stays untouched, and constraint roles are
cleared as in C. Dangling source gates are safely retained rather than using
C's final no-dangling assertion. Immutable fixtures match counts, register
partitions and exact drivers, including both dummy fallbacks. Exhaustive tests
cover projection values, constraints, extended nodes, metadata, limits,
cancellation and source immutability. Independent-formula fuzzing passed
629,432 executions in 15 seconds. Full race tests and vet pass; fixture hash
and package-first headers were verified. General/dual-output trimming and the
full ABC source port remain incomplete.

`SequentialCleanup` now ports `Gia_ManSeqMarkUsed`/`Gia_ManSeqCleanup`, using
owned iterative sequential reachability instead of source marks. All original
PIs/POs survive, including constraints; registers and logic outside their
transitive RI/RO closure are removed. Retained registers stay in original
order, with explicit source-register indices for state/reset projection.
Raw extended nodes/control edges, name/spec and source scratch survive.
A immutable fixture matches object/register counts and exact drivers, including
an unobservable cycle and dangling gate. Exhaustive projection tests cover
extended logic, constraints, original register order, unused PIs, no-property
cycles, limits, cancellation and source immutability. Independent dependency
closure and Boolean-value fuzzing passed 626,650 executions in 15 seconds.
Full race tests and vet pass; fixture hash and package-first headers were
verified. Constant/equivalent register reduction, full structural sweep,
general trimming and the full ABC source port remain incomplete.

`ReduceZeroInitializedRegisterInputs` now ports `Gia_ManReduceEquiv`.
ALL registers must start at zero: reads of zero-RI registers become zero and
identical phased RI drivers share the earliest RO. Complementary drivers stay
separate. The returned source-CI substitutions and constant/equivalent counters
make the reduction explicit; register slots remain for subsequent sequential
cleanup. Changed graphs use checked remapping/extended-node lowering; unchanged
graphs are independently cloned instead of aliased as in C. Constraints,
name/spec and source scratch survive. A immutable fixture matches count and exact
CO drivers. Exhaustive eight-frame checks across every eight-bit PI sequence
verify original/reduced traces from zero state, including extended nodes and
constraints. Caps, cancellation, substitution maps and unchanged ownership
are checked. Independent-formula fuzzing passed 555,277 executions in
15 seconds. Full race tests and vet pass. Full constant analysis, iterative
structural sweep and the full ABC source port remain incomplete.

`EvaluateTernary` now ports the three-valued gate/CO round kernel from
`giaTsim.c` and `Gia_XsimNotCond`/`Gia_XsimAndCond` from `gia.h`, using owned
full-object values instead of frontier scratch. Encodings match ABC:
zero=1, one=2, unknown=3; encoding zero is invalid. Conservative native
XOR/mux/buffer extensions include equal known mux branches under unknown
control. Reconvergent correlations are not recovered. Thirty-six native
AND/phase cases match exactly. Exhaustive independent Boolean completions
check every local extended gate table; mixed-graph fuzzing checks soundness
against every compatible Boolean completion. Invalid inputs, dimensions,
resource caps, cancellation and source immutability are tested. Fuzzing passed
696,622 executions in 15 seconds; full race tests and vet pass. Fixture hash
and package-first headers were verified. This is only the simulation kernel:
packed/frontier storage, state history, retirement/convergence, transposition,
constant-register reduction and the full ABC source port remain incomplete.

`SimulateTernaryHistory` now ports all-zero initialization, unknown PIs,
transition transfer, state recording/counts, repeated-state convergence and
changing-register retirement from `Gia_ManTerSimulate`. Owned histories include
initial and final repeated states, matching C's counters. Default retirement
broadens changed registers after rounds 301,311,... before the next transition.
MaxStates/MaxStateEntries and graph limits bound work/storage; exhaustion and
cancellation return no partial history and never claim convergence. Constraints
do not restrict simulation. Lookup-only state maps never determine iteration.
Two immutable fixtures match every state, loop index and zero/unknown count,
including a nine-bit counter with 308 states and loop start 306 that exercises
retirement. Tests cover state ownership, budget exhaustion, cancellation and
source immutability; independent concrete-trace soundness fuzzing passed
651,933 executions in 15 seconds. Full race tests and vet pass; fixture hash
and package-first headers were verified. Packed/frontier storage, history
transposition/map derivation, constant-register reduction, full structural
sweep and the full ABC source port remain incomplete.

`ReduceTernaryZeroInitializedRegisters` now ports history transposition/equality
map derivation and `Gia_ManReduceConst`. All-history-zero register reads become
zero; identical fully known histories use the earliest RO. Histories containing
unknowns cannot justify equality, and complementary histories are not merged.
ALL registers must initially be zero. Direct owned history-column keys replace
C's packed transposition; lookup maps never determine iteration. Source-CI
substitutions/counters are returned and register slots remain for cleanup.
Constraints, name/spec and source scratch survive; limits/cancellation never
return partial reductions. Immutable fixtures match source-CI substitutions,
object count and exact drivers for a constant self-loop and equal toggling
registers. Exhaustive zero-initialized traces cover extended nodes, constraints,
source immutability and unknown-history rejection. A test's accidental PI/state
partition was corrected before completion. Independent-formula fuzzing passed
508,083 executions in 15 seconds. Full race tests and vet pass; fixture hash
and package-first headers were verified. Full structural-sweep orchestration,
remaining graph algorithms and the full ABC source port remain incomplete.

`SweepZeroInitializedRegisters` now ports `Gia_ManSeqStructSweep` orchestration:
source-ordered sequential cleanup, optional ternary history reduction, and
optional repeated cleanup/structural RI equivalence until stable. ALL registers
must initially be zero. Source-register projections compose through every
cleanup; deleted-state witness reconstruction remains a separate obligation.
All PIs/POs, constraint roles and name/spec survive. Constant-only mode retains
substituted register slots as C does. Combinational graphs use ordinary cleanup.
Iteration/history/object/mapping budgets and cancellation fail without partial
results. Four immutable fixtures match object/register counts and exact drivers
for every constant/equivalence mode combination. Exhaustive zero-initialized
traces verify original/reduced properties, constraints and retained-state
projections. Tests cover cascading cleanup after Boolean simplification,
combinational paths, source immutability, budgets and cancellation. Independent
trace fuzzing passed 407,410 executions in 15 seconds; full race tests and vet
pass. General/dual-output trimming can now use this dependency, but remains
unported along with the rest of the unfinished ABC source tree.

`TransformOutputPairs` now ports `Gia_ManTransformMiter`/`Miter2`, XORing
adjacent pairs or corresponding PO halves while retaining RIs. All source
POs participate, and constraint roles are cleared as in C. Hashing lowers
extended nodes, cleanup preserves CI/register order, and name/spec/source
scratch survive. `NonconstantOutputPairIndices` ports the dual-trimming
selector: adjacent XORs are structurally swept from zero initialization and
both constant-zero AND constant-one drivers are excluded. The API explicitly
avoids treating this as an equivalence verdict despite C's NonZero name.
Immutable fixtures match both miter counts/drivers and selected pair indices.
Exhaustive tests cover extended nodes, preserved RIs, constant phase behavior,
sequential constants, dimensions, caps, cancellation and source immutability.
Independent-formula fuzzing passed 495,512 executions in 15 seconds. Full race
tests and vet pass. General trimming orchestration and signal-name metadata
remain unported, along with the rest of the unfinished ABC source tree.

`Trim` now ports `Gia_ManDupTrimmed` orchestration and constant-phase removal.
It supports optional unused-PI trimming, constant-zero/one/both PO removal,
and dual pair selection plus zero-initialized structural sweeping. ALL
registers must initially be zero for dual trimming. Returned CI/CO/register
projections compose through sweeping; -1 denotes dummy unused PIs or zero POs.
Dummy one-PO/two-PO fallbacks, PI references from dangling logic and C's
both-constant-XOR-phase exclusion are preserved. Constraint roles are cleared
as in C; raw gate kinds survive before optional lowering/sweep. Name/spec and
source scratch survive. Thirteen native records match object/input/register
counts and exact drivers: all 12 ordinary flag/phase combinations and dual
sequential sweeping. Exhaustive tests check source projections, extended gates,
constant phases, reset/state maps, dummy fallbacks, metadata, caps and
cancellation. Ordinary fuzzing passed 590,569 executions and dual trace/mapping
fuzzing passed 244,610 executions, each in 15 seconds. Full race tests and vet
pass; fixture hash and package-first headers were verified. Earlier general
trimming pending entries are superseded by this implementation. Source signal
name metadata, remaining graph algorithms and the full ABC source port remain
incomplete.

`ToDualOutputs` and `InterleaveOutputHalves` now port
`Gia_ManTransformToDual` and `Gia_ManTransformTwoWord2DualOutput`. The first
pairs each PO with zero and retains hashed dangling logic; the second
interleaves corresponding PO halves and performs cleanup. Both preserve
original RIs, CI/register order and name/spec while clearing constraint roles
as C does. Extended gates are lowered without source scratch mutation. Native
fixtures match counts and exact drivers, including different dangling-logic
behavior. Exhaustive tests verify output ordering, extended-node truth and
XOR-miter roundtrips, retained transitions, metadata, dimensions, caps,
cancellation and source immutability. Independent-formula fuzzing passed
538,025 executions in 15 seconds. Full race tests and vet pass; fixture hash
and package-first headers were verified. Remaining miter/graph algorithms and
the full ABC source port remain incomplete.

`CollectOutputSide` and `OrderDualOutputSides` now port
`Gia_ManCollectOneSide` and `Gia_ManTransformDualOutput`. Side traversals return
unique gate DFS postorder for even/odd POs, excluding RIs and including native
mux control/buffer dependencies. Combinational dual graphs hash the smaller
side first; adjacent outputs swap only when the even side is larger, with
explicit swap/count metadata. Ties keep original order. Shared logic is hashed
once, extended nodes are lowered, cleanup removes dangling logic, and CIs,
name/spec and source scratch survive. Constraint roles clear as in C. Three
immutable fixtures match both side lists, counts and exact output drivers.
Exhaustive tests cover swapping/ties/sharing, extended control cones, metadata,
invalid sides, caps, cancellation and source immutability. Independent-formula
fuzzing passed 512,469 executions in 15 seconds. Full race tests and vet pass;
fixture hash and package-first headers were verified. Remaining miter/graph
algorithms and the full ABC source port remain incomplete.

`NormalizeInitialState` now ports default raw `Gia_ManDupZeroUndc` behavior.
Known-one resets invert register reads/RIs; unknown resets add initial-value
PIs plus a zero-initialized selector register whose RI is one. Unknown values
are selected at frame zero only; later added-PI values cannot affect state.
Original CI/register ordering, constraints, raw extended gates and name/spec
survive. Returned all-zero reset lists, unknown-register indices and selector
index expose the encoding. Caller-requested extra unused PIs and unused
unknown-register handling match C. Mux-control-only fanouts are included,
extending C's two-data-fanin marking. Three immutable fixtures match object/input/
register counts and exact drivers. Exhaustive multi-frame tests cover all 27
three-register reset patterns, all unknown assignments, varied later added PIs,
constraints, state encoding, extended gates, unused unknowns, mux-control-only
usage, metadata, caps and cancellation. Independent trace fuzzing passed
570,124 executions in 15 seconds. Full race tests and vet pass; fixture hash
and package-first headers were verified. fGiaSimple degenerate-AND mode, further
reset/miter algorithms and the full ABC source port remain incomplete.

`AppendAndSimplified`, `AppendOrSimplified`, `AppendMuxSimplified`,
`AppendMajoritySimplified`, `AppendXorSimplified` and
`AppendXorRealSimplified` now port corresponding `Gia_ManAppend*2` operations
with fGiaSimple=0. They fold constants/equal/complementary literals but do not
strash duplicate nontrivial gates. Native decompositions and allocation order
are preserved; inputs are checked before mutation. Low-level multigate capacity
failures may retain intermediate gates, explicitly documented. Source inspection
corrected the earlier characterization of fGiaSimple: it DISABLES trivial folds
and permits raw degenerate ANDs, rather than providing more simplification.
That representation mode remains unported. Immutable fixtures match exact literal
returns and object counts for 1,296 cases; exhaustive Boolean checks cover all
phases/constants across the six constructors. Tests cover full-capacity folds,
allocation failure, duplicate retention and invalid-input mutation prevention.
Independent mux fuzzing passed 691,625 executions in 15 seconds. Full race tests
and vet pass; fixture hash and package-first headers were verified. Remaining
constructor/representation modes and the full ABC source port remain incomplete.

`ResetComparisonMiter` now ports `Gia_ManMiter2`. CI specification bytes x/X
select shared/independent PIs or unknown reset values; register 0/1 resets are
normalized. Two zero-initialized state banks compare every source PO by XOR,
then retain first/second RI banks. Source constraint POs become comparisons,
not assumptions, matching C. Returned normalized PI maps, unknown-register
indices, zero resets and per-bank register counts expose input/state layout.
Extended gates are lowered, cleanup preserves CI ordering, name/spec survive
and source scratch stays unchanged. Six immutable fixtures match counts and exact
drivers across shared/independent inputs and resets. Multi-frame exhaustive
tests cover initial choices, later unknown-input irrelevance, known-one phase
encoding, constraints, metadata, malformed specifications, caps, cancellation
and source immutability. Independent trace fuzzing passed 365,197 executions
in 15 seconds. Full race tests and vet pass; fixture hash and package-first
headers were verified. Remaining miter/choice algorithms and the full ABC source
port remain incomplete.

`BuildGiaChoiceMiter` now ports `Gia_ManChoiceMiter`: equal-dimension source
implementations share all CIs, hash reachable logic in output-major/source-minor
DFS order and collect each corresponding CO. This is a choice working graph,
not an XOR equivalence query or independent-machine composition. The original
register count remains; only that many final collected outputs become RIs.
Constraint roles clear as in C. Per-source owned maps also preserve repeated
manager-pointer outputs, avoiding C scratch-alias loss. Extended nodes are
lowered, name/spec come from source zero and source scratch stays unchanged.
Immutable fixtures match count and exact drivers for three implementations.
Exhaustive checks cover shared logic/input truth, output order, aliases,
metadata, invalid dimensions/nil sources, summed map/output budgets and
cancellation. Independent-formula fuzzing passed 484,625 executions in
15 seconds. Full race tests and vet pass; fixture hash and package-first
headers were verified. Choice-class derivation, remaining graph algorithms
and the full ABC source port remain incomplete.

`ConvertOutputRoles` now ports `Gia_ManDupWithConstraints`. Explicit types
retain ordinary type-0 POs, complement type-1 violation signals into positive
constraints, and omit all other types. Stable role groups precede unchanged
RIs; existing roles are superseded. Returned source-CO indices and complement
flags expose reordering/phase changes. Raw extended nodes, dangling logic,
CI/register/reset order and name/spec survive without source scratch mutation.
Eighty-one native assignments over four POs match counts and exact drivers.
Exhaustive projection tests cover mixed/all/omitted roles, arbitrary ignored
types, extended gates, constraints, metadata, dimensions, caps, cancellation
and source immutability. Independent Boolean-formula fuzzing passed 578,301
executions in 15 seconds. Full race tests and vet pass; fixture hash and
package-first headers were verified. Remaining graph/verification algorithms
and the full ABC source port remain incomplete.

`ExtractSignalWithInputs` and `WithInsertedCone` now port explicit-support
cone copying and CI/object-bound reinsertion from `Gia_ManDupConeSupp`/
`Gia_ManDupConeBack`/`BackObjs`. Extraction retains caller-ordered unique CI
indices, checks complete support, preserves raw extended nodes and safely
supports constant/CI roots. Reinsertion hashes on an owned target clone with
phased/constant/aliased arbitrary signal bindings, preserves IDs/outputs/hash
state, and returns the resulting literal without implicit output rewiring.
Name/spec and both source graphs' scratch stay unchanged; failures return no
partial graph/literal. Immutable fixtures match extraction counts/drivers and
reinsertion literals/counts for existing-logic reuse and internal-gate binding.
Exhaustive tests cover extended gates, permutations, constants, aliases,
active hashing, invalid/incomplete supports, invalid bindings, caps,
cancellation and source immutability. Independent Boolean-formula fuzzing
passed 558,997 executions in 15 seconds. Full race tests and vet pass; fixture
hash and package-first headers were verified. Remaining graph/verification
algorithms and the full ABC source port remain incomplete.

`ExtractEncounterOrderedCone` and `ExtractSignalsRetainingInputs` now port
`Gia_ManDupDfsNode`/`Gia_ManDupDfsLitArray` without repr/choice metadata. The
first returns source CI ordinals in first DFS encounter order, preserving
CI objects interspersed with gates; phased/constant/CI roots extend C's
AND-only node entry point. The second retains all CIs in original order and
raw-copies shared reachable gates once as ordered phased outputs are visited.
Both return combinational graphs, preserve raw extended nodes and name/spec,
and leave source scratch unchanged. Immutable fixtures match counts, physical
CI positions and output drivers for encounter order, repeated roots and
constants. Exhaustive tests check projected inputs, extended-node truth,
metadata, invalid signals, dimensions, caps, cancellation and source
immutability. Independent Boolean-formula fuzzing passed 526,038 executions
in 15 seconds. Full race tests and vet pass; fixture hash and package-first
headers were verified. Repr/choice metadata, remaining graph algorithms and
the full ABC source port remain incomplete.

`DuplicateDFS` and `SelectPropertyWithState` now port `Gia_ManDupDfs` and
`Gia_ManDupDfsOnePo`. All CIs remain in original order; output-root DFS gates
are raw-copied and unreachable gates are omitted. Single-property selection
retains ALL original RIs/register slots, unlike sequential-closure extraction.
Whole-graph constraint roles survive; a selected PO becomes ordinary and clears
roles as in C. Owned source-object maps expose retained CO objects and invalid
omitted gates/outputs. Raw extended node kinds, name/spec, register/reset order
and source scratch survive. Four native cases match counts/constraint roles
and exact drivers for whole-copy and every property selection. Exhaustive
checks cover extended nodes, unchanged transitions, former constraint outputs,
metadata, maps, invalid indices, limits, cancellation and source immutability.
Independent-formula fuzzing passed 579,308 executions in 15 seconds. Full race
tests and vet pass; fixture hash and package-first headers were verified.
Embedded counterexample/repr/choice metadata and the remaining full ABC source
port remain incomplete.


`BuildGiaDesignMiter` now ports `Gia_ManMiter_rec`/`Gia_ManMiter` for two
distinct graphs. Combinational mode compares all COs and treats every CI as
data; sequential mode requires matching PI/PO counts and nonempty state banks,
while permitting different register counts. Input prefixes share and selected
suffixes are independent; sequential RIs/state banks concatenate first then
second. XOR, first-AND-NOT-second implication violations and adjacent dual
outputs preserve C ordering. Constraints become compared data; result roles
clear, name becomes miter, Spec clears and source scratch stays unchanged.
Owned CI maps expose projections; aliased source pointers work safely.
Extended nodes lower into hashed logic. Twelve native cases match normalized
counts and exact drivers. C combinational implication asserts when visiting
RIs of a graph retaining register roles; Go supports all-CO semantics and
checks that extension independently. Exhaustive tests cover unequal state-bank
sizes, all independent suffix sizes, extended XOR nodes, constraints, maps,
aliases, dimensions, resource caps and cancellation. Independent-formula
fuzzing passed 205,360 executions in 15 seconds. Full tests, race tests and
vet pass. Fixture hashes are pinned. Other duplication algorithms,
full verification parity and the remaining ABC source port stay incomplete.


`AggregateOutputs` now ports `Gia_ManDupAndOr`, used by the sequential
correspondence path. It hashes source logic in physical order, folds all
combinational COs with ordered AND/OR, optionally complements the result,
and emits the requested number of repeated outputs. Empty AND is true,
empty OR false, and zero repeated outputs is valid. All CIs survive cleanup;
Name survives, Spec clears and constraints become ordinary data. Sequential
sources reject explicitly. Extended XOR/mux/buffer nodes lower semantically;
source scratch remains unchanged. Twenty-four native cases match object
counts and exact drivers across empty/nonempty sources, AND/OR, complement
and zero/one/two outputs. Exhaustive truth checks and independent extended-node
formula fuzzing pass; the 15-second campaign completed 264,483 executions.
Dimensions, nil sources, limits and cancellation have negative tests.
Full tests, race tests and vet pass; fixture hashes and package-first headers
are verified. The source file remains partial; sequential correspondence,
remaining GIA algorithms and the full ABC source port remain unfinished.


`SplitRegisterHalf` now ports `Gia_ManDupHalfSeq`: first floor(R/2) versus
remaining registers, with omitted-bank ROs promoted to PIs and retained-bank
ROs last. Every original PO remains, followed by the selected RIs. Owned
CI/register source ordinals expose the input permutation and reset selection.
Raw per-output DFS copying preserves extended gate types and physical CO
interleaving, drops unreachable gates and retains every CI. Name/spec survive;
constraint roles clear as in C. Source scratch remains unchanged. Eight native
cases with zero through three registers match object counts, physical CO IDs,
register counts and exact drivers. Exhaustive projection/transition tests
cover zero through five registers and extended XOR/mux logic. Negative tests
cover nil sources, dimensions, allocation limits and cancellation. A bounded
semantic fuzz campaign passed 156,636 executions in 15 seconds. Full tests,
race tests and vet pass; pinned fixture hashes and package-first headers are
verified. This partition exposes omitted state as unconstrained inputs and
alone proves no source property. Sequential equivalence-class merging,
correspondence driver and the remaining full ABC source port are unfinished.

`SplitPairedOutputs` now ports `Gia_ManDemiterDual`/`Gia_ManDemiterTwoWords`
and their selected-root reachability paths. Combinational even/odd or
first/second-half output selection retains every CI, copies shared cones
independently per result in DFS postorder, preserves raw extended node types
and name/spec, and clears source constraint roles. Explicit source-CO ordinal
projections describe both results. Empty pairs are valid; sequential sources
and odd output counts reject. Failures expose neither partial graph. Eight
native cases match object counts and exact CO drivers; exhaustive tests cover
both layouts, empty outputs, extended XOR/mux/buffer cones, metadata and
independent result/projection ownership. Negative tests cover dimensions,
resource caps and cancellation during either side. Byte-driven structural
fuzzing varies graph construction, root selection, polarity and layout, checks
all eight input assignments and passed 305,868 executions in 15 seconds.
Full tests, race tests and vet pass; pinned fixture hashes and package-first
headers are verified. Structural XOR decomposition, sequential demiter modes,
command integration and the remaining full ABC source port remain unfinished.

`DecomposeOutput` now ports `Gia_ManDupDemiter` for one ordinary-AND-rooted
PO, complemented or not. Distinct phased AND boundaries sort by ascending
source level; ties deliberately retain DFS encounter order. Complemented roots
apply De Morgan, so the returned OR/AND operator and source part literals
expose how outputs reconstruct the source PO. All source CIs and RIs remain
with unchanged register/reset order. Name/spec survive, constraints become
ordinary data and source scratch stays unchanged. Extended nodes are opaque
boundaries, copied through semantic hashed lowering. Non-AND roots and fewer
than two distinct boundaries reject safely instead of entering C's unchecked
traversal. Four native cases match exact object counts and CO drivers for both
root polarities and sequential/combinational roles. Negative tests cover nil,
dimensions, unsupported roots, limits and cancellation. Structural byte-driven
fuzzing varies ordinary/extended gates, phased boundaries and sequential roles,
checks folded outputs and retained transitions for all input assignments, and
passed 151,533 executions in 15 seconds. Full tests, race tests and vet pass;
pinned fixture hashes and package-first headers are verified. XOR-recovery
partitioning, command integration and the remaining full ABC port are pending.


`RecoverDualOutputs` now ports `Gia_ManCollectTopXors`/`Gia_ManDemiterToDual`
with greedy minimum-new-support ordering and cone-overlap side partitioning.
Owned visited/mark sets reproduce shared-cone counts without changing source
scratch. Reversing candidate order during assignment and then reversing the
complete pair vector matches C's pair/side layout. XOR operand phases and
source output polarity are superseded by zero-input phase normalization,
matching C; explicit source signal literals expose that contract. Non-XOR
candidates pair with false. All source CIs (including ROs) become combinational
inputs and source RIs participate as candidate COs; roles clear. C's unusual
first-CO-constant override emits just (false, source constant), even with later
outputs; candidate gates are copied before that override as in C. Name/spec
survive, raw gates copy in selected-cone DFS order and extended gates receive
full semantic support/overlap traversal. Native XORs are recognized; other
extended gates are opaque top boundaries. Nine immutable fixtures match exact
object counts and drivers across shared cones, repeated candidates, constants,
multiple outputs and inverted/phased structures. Structural semantic fuzzing
passed 293,291 executions, then 274,799 executions after adding native XOR and
buffer generation (15 seconds each). All input assignments check returned
source signal projections; nil/empty dimensions, limits and cancellation are
tested. Full tests, race tests and vet pass; hashes/package-first headers are
verified. This structural conversion alone proves no equivalence assertion.
Alternative support-only ordering, command integration and the full remaining
ABC source port remain unfinished.


`CycleFromZero` now ports the zero-initialized `Gia_ManDupCycled` pipeline
and its `Gia_ManCycle` transition simulation. Nil traces use a local uint32
multiply-with-carry generator with C's seeds and discarded reset/advance draw;
explicit traces use frame-major PIs. Trace dimensions are validated and at most
LastFrame transitions are permitted. Initial trace bits and property metadata
are intentionally ignored as in C. Reached register bits are returned, and
raw FlipRegisters recoding makes zero target state represent that reached
source state. Source CI/CO/register order, dangling logic, name/spec survive;
constraint roles clear as in C. Source/trace scratch stays unchanged and
extended gates simulate actual semantics, extending C's ordinary-AND path.
Twelve native cases match reached states and recoded CO drivers for zero
through five transitions using random or trace inputs (trace initial bits
are deliberately one). Exhaustive state/input checks verify recoded outputs
and transitions. Independent XOR/mux Boolean-formula fuzzing passed 228,659
executions in 15 seconds. Negative tests cover nil graphs, frame/trace
mismatches, dimensions, allocation/frame caps and cancellation. Full tests,
race tests and vet pass; hashes/package-first headers are verified. The
caller-mark-initialized standalone C cycle variant, remaining algorithms,
command integration and the full remaining ABC source port are unfinished.


`ExtractOutputGroup` now ports `Gia_ManDupOutputGroup`: selected CO ordinals
[start, stop) become combinational outputs, and CIs appear only at their first
DFS encounter, including encountered source ROs. Explicit source-CI ordinals
expose the resulting input projection. Raw extended gates and physical
per-output gate/CO interleaving survive, as do name/spec; state/constraint
roles clear and unused CIs/unreachable gates disappear. Empty ranges return a
constant-only graph. Source scratch stays unchanged. Fifteen immutable fixtures
cover every contiguous range of a four-output sequential source and match
object counts, source input projections, physical CI/CO IDs and exact drivers.
Structural byte-driven fuzzing varies gate creation, phased roots, ranges and
state roles and checks projected truth for all input assignments; 332,720
executions passed in 15 seconds. Negative tests cover nil graphs, invalid
ranges, dimensions, object/map limits and cancellation. Full tests, race tests
and vet pass; fixture hashes/package-first headers are verified. Other
order-DFS variants and the full remaining ABC source port stay unfinished.


`NormalizeObjectOrder` now ports `Gia_ManDupOrderAiger` without external
choice/equivalence metadata. All CIs precede raw gates in original physical
order (including dangling logic), followed by COs. An owned source-object map
includes CO object literals. Terminal/register/reset order, raw extended node
kinds and name/spec survive; constraints clear as in C. Repeated normalization
is idempotent. `ZeroOutputCopy` ports `Gia_ManDupZero`: only terminal/register
counts and Name survive, all COs including RIs drive false, original gates
are omitted and Spec/constraint roles clear. Both avoid source scratch changes.
Immutable fixtures match counts, physical terminal positions and exact drivers
on an interleaved sequential source with dangling logic. Exhaustive checks
cover source truth, object/terminal mappings, XOR/mux kinds, state roles and
zero-output behavior. Structural fuzzing varies interleaved CI/CO creation,
phases, gates and register roles, checks all input assignments and passed
224,084 executions in 15 seconds. Nil/dimensions, allocation/map caps and
cancellation reject explicitly. Full tests, race tests and vet pass; fixture
hashes/package-first headers are verified. Other ordering/metadata algorithms,
command integration and the full remaining ABC source port remain unfinished.


`WithAppendedGraph` now ports `Gia_ManDupAppend`/`Gia_ManDupAppendShare`
through owned target clones with source-object maps. Independent mode appends
all CIs and raw gates; shared mode binds every CI by ordinal and hashes gates,
lowering extended nodes. Every source CO appends, dangling logic remains,
base IDs/name/spec survive and hashing remains active. Source aliases including
self-append work safely without C scratch aliasing. Roles deliberately match
C's positional append contract: independent clears register count, shared keeps
base register count so final appended COs become RIs, existing constraints stay
positional and source roles are not imported. This operation is distinct from
state-bank-preserving sequential composition. Immutable fixtures match object and
role counts, physical CI IDs and exact drivers for both modes. Exhaustive
truth checks cover independent/shared inputs, self-append, active hashing,
extended XOR logic, maps and independent ownership. Structural fuzzing varies
raw/hashed/XOR/buffer gates and polarity; all input assignments check concatenated
source outputs and maps, passing 187,794 executions in 15 seconds. Negative
tests cover nil/mismatched sources, caps, dimensions and cancellation. Full
tests, race tests and vet pass; fixture hashes/package-first headers are
verified. Remaining duplication/metadata algorithms, command integration and
the full remaining ABC source port stay unfinished.


`ClassifyOutputsBySupport` now ports combinational `Gia_ManIsoStrashReduce`
and its normalized-driver classification path. Classes follow first source
occurrence and retain one representative cone with all original CIs. Explicit
class members, source-CO-to-class indices and per-output sorted CI supports
expose each projection. Unique structural support sizes remain singleton,
matching C. Other cones group by phased normalized hashed driver; equality is
under per-cone support renaming, not a single common source input assignment.
Actual zero/one-support logic is copied, preserving the established Go
correction to C's size-only shortcut. Extended nodes lower semantically during
grouping; representative graphs preserve raw kinds. Source constraints become
ordinary data, result roles clear, name/spec survive and source scratch stays
unchanged. Immutable fixtures match class lists, representative graph counts and
exact drivers for renamed supports, phases, constants, unique support size and
empty outputs. Structural fuzzing checks class coverage/order, representatives
and normalized class truth under every input assignment, passing 173,738
executions in 15 seconds. Nil/sequential sources, dimensions, caps and
cancellation reject explicitly. Full tests, race tests and vet pass; pinned
hashes/package-first headers are verified. The classification is conservative
structural grouping, not a complete functional equivalence procedure. Other
metadata/equivalence algorithms, command integration and the remaining full
ABC source port remain unfinished.


`NormalizeGraph` now ports non-box `Gia_ManDupNormalize` from giaTim.c,
providing the constraint-preserving normalization prerequisite for top-AND
reduction and other pipelines. All CIs precede physical-order gates and COs;
register/reset order, Name/Spec and constraints survive. Raw mode preserves
extended kinds; hashed mode simplifies/merges ordinary logic and semantically
lowers native XOR/mux gates while retaining every explicit buffer object as
in C. Neither mode removes dangling nodes; hashing stops on the result.
Owned maps include all source objects and CO literals, and source scratch/hash
state remains unchanged. Immutable fixtures match raw/hashed counts, constraint
roles, buffer IDs and exact drivers on interleaved sequential graphs with
duplicate ANDs and dangling logic. Structural fuzzing checks every source
object's mapped value under all input assignments, buffer counts and role
preservation, passing 237,666 executions in 15 seconds. Nil/dimensions, limits
and cancellation reject explicitly. Full tests, race tests and vet pass;
fixture hashes/package-first headers are verified. giaTim.c is now partial;
timing-box reshuffling, mapping/timing/equivalence metadata and other timing
routines remain unported. Top-AND reduction itself and the full remaining ABC
port are unfinished.


`ReduceTopAndOnce`/`ReduceTopAnd` now port `Gia_ManDupTopAnd_iter` and
its count-fixed-point pipeline. Positive ordinary AND frontiers expand with
owned visited state; distinct phased boundaries sort by source object ID.
CI boundaries become explicit fixed constants and remaining signal cones copy
raw or hash through substitution. Fixed assignments accumulate by stable CI
ordinal across iterations. A frame-independent iteration cap returns ErrLimit
if no fixed point is reached. This is an all-CO conjunctive-query transform,
not sequential safety: source constraints/RIs participate as data. Actual
reductions clear constraints; no-fixed copies become combinational, while
fixed-input copies keep the source register count only when the output layout
remains valid. Normalized fallbacks preserve roles. Name/spec and source
scratch survive. Extended nodes are opaque boundaries. Explicit contradictions
return normalized fallbacks. Go deliberately corrects C's false-constant CO
loss when another root expands; immutable fixtures capture that bug separately.
Fourteen native parity cases match exact counts/drivers/register roles; two
additional native cases establish the correction. Exhaustive reconstruction
checks and structural fuzzing verify original conjunction = fixed-input
satisfaction AND reduced conjunction (or false on explicit contradiction).
The 15-second fuzz campaign passed 230,142 executions. Dimensions, nil graphs,
limits, iteration exhaustion and cancellation are tested. Full tests, race
tests and vet pass; fixture hashes/package-first headers are verified. Other
metadata/equivalence algorithms, command integration and the full remaining
ABC source port are unfinished.


`BitVector` now ports the packed-vector operations from misc/vec/vecBit.h.
Word layout is uint32 LSB first. Start/full/fill logical sizes round to whole
words as in C, while fill-extra, push and autoextending get/set use exact bit
sizes. Owned storage supports checked reads/writes/add/last, capacity growth,
shrink/clear, reset/count, arbitrary-size cloning, packed snapshots and storage
release. Zero values work; caps bound rounded allocations and failures leave
state unchanged. Shrink/clear stale bits cannot reappear through growth.
Go corrects C fill-extra's unsafe shifts/word-boundary writes and extends Dup
to unaligned sizes. Snapshots own their storage and zero tail padding; release
transfers logical words and clears spare capacity. StorageBytes reports actual
packed payload instead of C's bit-capacity-as-int memory overcount. Go lifetime
management replaces Free/FreeP and explicit loops replace iteration macros.
Nine native operation snapshots match sizes, capacities, packed words and
population counts. A boundary matrix covers zero/partial/aligned word sizes,
prefix preservation, clone/release ownership and limit-error atomicity.
Stateful byte-driven fuzzing checks thirteen operation kinds against an
independent Boolean-array model, passing 15,706 executions in a 15-second
campaign (16 seconds wall time). Full tests, race tests and vet pass; fixture
hashes/package-first headers are verified. The header is tracked partial for
its deliberate safe-storage/lifetime differences. Other vector families,
command integration and the full remaining ABC source port are unfinished.


`FloatPriorityQueue` now ports Vec_Que allocation/growth, priority lookup/setup,
max-heap movement/update, membership, top/pop/push/clear and invariants. Integer
IDs use float32(ID) by default; optional priorities are owned copies. Equal
priorities do not move upward and left children win downward ties as in C.
Upward/downward SetCost repair works for members; nonmembers can receive costs
before insertion. Bulk cost replacement validates first and rebuilds safely.
NaNs reject without mutation; infinities and signed zeros work. Zero values
are usable, ID/allocation caps are checked, clone/diagnostic ID snapshots own
storage, and empty top reports explicit absence. Check handles arbitrary sparse
membership, correcting C's dense-prefix assumption. Four immutable fixtures match
initial/updated heap layouts and complete pop orders for sparse ID fallback,
equal priorities, updates, infinities and signed zeros. Ownership/error tests
cover aliases, cloned priorities, invalid IDs/cost dimensions/duplicates/NaNs,
limits and invariant corruption. Stateful fuzzing checks membership, maximum
priority and heap invariants against an independent model after seven kinds
of operations; the latest campaign passed 193,094 executions. Full tests, race
tests and vet pass; fixture hashes/package-first headers are verified. Pointer-
to-external-cost mutation is deliberately replaced by owned setters; Free/FreeP
use Go lifetime management. `Format` preserves the source print layout and the
demonstration is superseded by ordinary/model tests, completing this header.
Other vector families, command integration and the full remaining ABC source
port are unfinished.


`FloatVector` now ports Vec_Flt allocation/start/full, exact/minimum-capacity
storage, copy/clone/move/release/erase, checked entries/add/max-update/last,
fill/extra/shrink/clear, push/ordered/unique/pop/find/remove, extrema/equality
and sorting. Float32 arithmetic rounds explicitly; all-one StartFull bytes
remain the exact 0xffffffff NaN payload. IEEE lookup/equality treats NaNs as
unequal and signed zeros as equal. Min/max preserve C's first-entry scan
behavior, exposing absence for empty vectors. NaNs reject sorting/ordered
insertion before mutation because C's comparator lacks consistent ordering;
ordinary sorting is stable, preserving signed-zero encounter order. Snapshots
and clones own their data; moves/releases transfer storage and empty sources.
Zero values work, allocation caps are checked and failures leave vectors
unchanged. Six native operation snapshots match sizes, capacities and exact
bits through fill, growth, arithmetic, insertion/removal and sorting. Ownership,
limits, NaN/infinity/zero tests pass. Stateful fuzzing passed 171,622 executions
in a 15-second campaign (16 seconds wall time); arbitrary IEEE-pattern fuzzing
passed 312,703 executions in 15 seconds, including addition checked against
a float64-sum reference rounded to float32. Full tests, race tests and vet
pass; fixture hashes/package-first headers are verified. Borrowed pointer/
adopted-array aliases and formatted print remain absent; Go lifetime handling
replaces Free/FreeP. Other vector families, command integration and the full
remaining ABC source port remain unfinished.


`WordMemory` now ports Vec_Mem fixed-width paged uint64 storage, owned reads,
last/write/get/set/push, page/entry/payload metrics, shrink and hexadecimal dump
semantics. Logical record/page-word limits check before mutation. New gaps
zero-initialize, unlike C's uninitialized records. Exact-page and empty shrink
release the actual unused pages, correcting C bookkeeping; surviving truncated
page payload clears. Clone deeply owns pages (C Dup itself is an empty stub).
Interning uses lookup-only Go keys and stable insertion IDs; no map iteration
controls results. Hash enable checks unique existing records, hashed records
reject direct mutation, shrink removes truncated keys and clone rebuilds an
owned index in ID order. Dumps are bounded, uppercase high-word-first lines.
Immutable fixtures match record values, pages and interning IDs through multiple
C hash resizes and duplicates; separate native page counts establish the shrink
corrections. Ownership, gaps, dump formatting, limit atomicity, dimension and
hash-state errors have unit tests. Stateful model fuzzing varies nine operation
kinds and page sizes, checking all records and invariants after each operation,
passing 203,333 executions in a 15-second campaign (16 seconds wall time).
Full tests, race tests and vet pass; hashes/package-first headers are verified.
Truth-table seed/mux/file wrappers and C hash bucket/key/profile internals
remain unported; C Fill/Clean are empty stubs. Pointer lifetime compatibility
is not exposed. Other vector families, command integration and the full
remaining ABC source port remain unfinished.


Truth-table memory constructors now port `Vec_MemAllocForTTSimple`,
`Vec_MemAllocForTT` and `Vec_MemAddMuxTT`. Checked word counts use one word
through six variables, then powers of two within C's positive signed-int range.
Storage uses C's 4096-record pages with limits checked before allocation.
Seeded tables intern false at ID 0 and variable-zero or complemented-variable
patterns at ID 1; AddMuxTruthTable interns stretched x2 ? x1 : x0 at ID 2.
Every unused/padded small-table bit retains C's repeated packed pattern. Reuse
is idempotent, including after later insertions. Width/hash-state/conflicting
ID layouts reject before mutation, replacing C's post-insertion assertions.
Twenty native cases for zero through nine variables and both phases match
word/page dimensions and every packed entry. Exhaustive assignment checks
verify the actual variable/mux formulas. Limits, wrong layouts, dimensions,
repeated seeds and unchanged state on errors have unit tests. Seed/interning
fuzzing checks stable IDs and complete records against an independent ordered
list, passing 217,122 executions in 15 seconds. Full tests, race tests and vet
pass; fixture hashes/package-first headers are verified. Truth-table file/
stdout reporting wrappers, C hash internals, other vector families, command
integration and the full remaining ABC source port remain unfinished.


`TruthTable` now ports core utilTruth.h operations with owned canonical packed
storage. Variable zero is the low assignment bit and sub-six-variable tables
repeat their logical pattern across one word; invalid padding rejects. Checked
constructors/read/write, constants/variables, clone/complement, phased-input
AND/sharp, OR/XOR, mux/majority, comparison/equality/opposite, implication and
intersection preserve this representation. Logical CountOnes excludes repeated
padding. Cofactors retain width/variable count and replicate selected halves,
using masked shifts within words and block copies across words. HasVariable/
Support compare the relevant paired halves. Every result owns its data and
input graphs/tables stay unchanged; allocation bounds check before creation.
Nine immutable fixtures through nine variables match exact Boolean/cofactor words,
support masks and comparison/implication/intersection results. Unit checks cover
all variable assignments, repeated-bit updates, constants, cloning/snapshots,
malformed padding, dimensions and limits. Boolean-array fuzz oracles check all
assignments, all cofactors and support for varied tables; 170,556 executions
passed in 15 seconds. Full tests, race tests and vet pass; fixture hashes and
package-first headers are verified. The header remains partial: raw digit
helpers, masked/accumulating variants, variable permutations, quantification,
canonicalization/decomposition and many other truth operations remain unported.
Command integration and the full remaining ABC source port are unfinished.


Truth-table variable transformations now port `Abc_TtFlip`, arbitrary/adjacent
`Abc_TtSwapVars`, the permutation semantics of `Abc_Tt6Permute_rec`, explicit
`Abc_TtImplementNpnConfig` application and `Abc_TtExist`. Owned results use
word-level masks/shifts for low variables, bit/word exchange across the 5/6
boundary, and block exchange for high variables. Source-to-target permutation
mapping is explicit: y[i] = x[order[i]], with source-input phases before
permutation and optional output complement. Caller phase/permutation slices
stay unchanged, unlike C's in-place permutation normalization. Existential
OR and universal AND retain variable slots/width without C's fixed 64-word
temporaries. A 13-variable check covers that safety extension. All 165 native
cases (pairs first <= second through nine variables) match packed flip/swap,
cyclic NPN, existential and universal-cofactor results. Independent assignment
oracles fuzz arbitrary permutations/phases and all assignments for transformed
and quantified tables; 258,811 executions passed in 15 seconds. Invalid
indices/permutations/phase lengths, nil tables, identity ownership and argument
immutability have unit tests. Full tests, race tests and vet pass; fixture
hashes/package-first headers are verified. This applies supplied NPN configs;
it does not derive canonical configs. Support compaction/stretching, masked
variants, larger truth decomposition/canonicalization, command integration and
the full remaining ABC source port remain unfinished.


`CompactSupport`, `ProjectVariables`, `Stretch` and `ExpandVariables` now port
truth support minimum-base/shrink/stretch/expand semantics. Compaction returns
ascending source-variable ordinals and C's movement flag (which can stay false
while logical width shrinks for prefix supports/constants). Projection can
retain arbitrary caller-ordered variables, rejecting removal of an active
input. Results narrow physical word count and logical width while preserving
canonical stretched padding. Stretch repeats source words into unused high
variables; expansion embeds inputs at distinct caller-selected positions,
generalizing C's sorted-cut mapping. Sources/mappings remain unchanged, results
own storage and limits carry through. Owned projections replace mutation of C
caller label arrays; larger allowed widths avoid the 16-variable support-mask
restriction. Thirty native cases through nine variables match compact words,
source support indices, movement flags, full restoration and stretched words
for sparse/prefix/constant supports. Assignment-oracle fuzzing checks projection,
compaction/restoration and arbitrary injective embedding, passing 277,626
executions in 15 seconds. Active-input deletion, duplicates/out-of-range
mappings, dimensions, nil sources, storage limits and identity ownership have
unit tests. Full tests, race tests and vet pass; fixture hashes/package-first
headers are verified. Raw helper variants, truth simplification/decomposition/
canonicalization, command integration and the full remaining ABC source port
remain unfinished.


`SimplifyInputs` now ports `Abc_TtSimplify`: constant input literals cofactor
away, equal/complementary symbolic labels merge through representative-relative
cofactors and mux reconstruction, then support compacts. First source occurrences
represent label classes. Returned surviving phased literals and source-slot
indices make result interpretation explicit: result inputs are literal VALUES,
including their phases, under shared symbolic-label assignments. The movement
flag matches MinimumBase, not whether substitutions changed logic. Lookup maps
only choose already-seen representatives and never iterate for ordering. Source
tables and caller literals stay unchanged; result tables/metadata own storage.
General-width cofactors extend C's six-variable alias-only path. Thirty-five
native cases through six variables match packed results, surviving phased
labels and movement flags across constants, sparse labels, repeated inputs
and opposite phases. An eight-variable alias case verifies the width extension;
unit checks cover argument/result ownership, all-constant reduction, dimensions,
nil tables and invalid literal sentinels. Independent bound-assignment fuzzing
checks every assignment of shared symbolic labels against the source table,
passing 317,972 executions in 15 seconds. Full tests, race tests and vet pass;
fixture hashes/package-first headers are verified. Larger truth decomposition/
canonicalization, command integration and the full remaining ABC source port
remain unfinished.


`DecomposeVariable` now ports `Abc_Tt6CheckOutDec`/`Abc_TtCheckOutDec`.
Cofactors test in C's exact order: c0 false, c1 false, c0 true, c1 true, then
opposite cofactors, returning AND/not-AND/not-OR/OR/XOR operator codes and
owned same-width remainders. The remainder is independent of the selected
variable and reconstructs the source with that operator. Unused variables
return no decomposition, replacing C's assertion. General cofactors remove
C's eight-variable/four-word scratch limit. Thirty-six native cases through
eight variables match exact operator codes and remainder words, including
irreducible mux cofactors. Unit checks cover rule precedence for literals,
unused variables, nil/dimension failures, source ownership and a 13-variable
XOR decomposition. Independent Boolean-array fuzzing checks ordered case
classification, remainder independence and exact reconstruction on every
assignment; 353,206 executions passed in 15 seconds. Full tests, race tests
and vet pass; fixture hashes/package-first headers are verified. Multi-stage
seven/eight-variable wrappers, pair-input DSD, larger decomposition/canonical
algorithms, command integration and the remaining full ABC port are unfinished.


`DecomposeInputPair` now ports `Abc_TtCheckDsdAnd` using four general-width
cofactors. Ordered tests return AND with either/both input phases or XOR, plus
an owned same-width remainder whose First slot denotes the pair's replacement
bit and whose Second slot is unused. Substituting the recognized pair function
into First reconstructs the source. C's comments interchange phases for codes
one/two; Go follows the executable cofactor conditions and documents the actual
phase encoding. Degenerate/constant precedence matches C; repeated/invalid
positions reject. Owned cofactors extend C's one-word kernel to larger tables.
All 210 native cases through six variables match exact operator codes and
remainder words for phased pairs, outer functions and irreducible projections.
Unit tests cover 13-variable decomposition, rule precedence, error dimensions,
nil tables and source ownership. Independent Boolean-cofactor fuzzing checks
classification, remainder independence and reconstruction on every assignment,
passing 312,245 executions in 15 seconds. Full tests, race tests and vet pass;
fixture hashes/package-first headers are verified. DSD mux, larger truth
canonicalization/decomposition, command integration and the remaining full ABC
source port remain unfinished.


`DecomposeMuxInput` now ports `Abc_TtCheckDsdMux` with explicit selector,
exclusive data-input positions, low-data phase and an owned remainder using
the selector slot for the mux replacement value. Each selector cofactor must
have exactly one data input absent from the other. Matching cofactor pairs
recognize positive low data; swapped pairs recognize complemented low data.
This corrects C's complemented-cofactor condition: it can report an invalid
factorization or miss a valid phased mux. Immutable fixtures capture 15 false
positives refuted by explicit reconstruction, 15 recovered valid muxes, and
42 exact parity cases (72 total). General cofactors extend the one-word C
kernel; a 13-variable phased mux unit case verifies that path. Remainders
are independent of both data variables and preserve the source function after
mux substitution. Unused selectors, invalid/nil inputs, constants and ownership
have unit checks. Independent Boolean-cofactor/support fuzzing tests candidate
selection, phase classification, data independence and every reconstruction
assignment; 282,985 executions passed in 15 seconds. Full tests, race tests
and vet pass; fixture hashes/package-first headers are verified. Larger truth
decomposition/canonicalization, command integration and the full remaining ABC
source port remain unfinished.


`NormalizePhases` now ports `Abc_TtCanonicizePhase` and its low/high-variable
helpers from dauCanon.c. The output complements when its highest packed bit is
one; one descending variable pass flips an input only if the high-word-first
truth value decreases. This matches C's actual single pass, not an exhaustive
NPN search or idempotence guarantee. Owned result tables and explicit input/
output phases reconstruct through TransformNPN with identity permutation.
Allocation-free `CompareCofactors` ports forward/reverse cofactor ordering
from utilTruth.h using masks and word blocks; full-word flips are equivalent
to C's partial flip after equal leading pairs. Source storage stays unchanged.
Sixty native cases through nine variables match exact words, phase masks and
both cofactor comparison orders. A Boolean-array greedy oracle and independent
configuration reconstruction pass 293,169 fuzz executions in 15 seconds.
Nil/invalid variables and result ownership have unit checks. Full tests, race
tests and vet pass; fixture hashes/package-first headers are verified.
dauCanon.c is now partial. Semicanonicalization, permutation/group refinement,
exact/adaptive/hierarchical canonicalization, command integration and the full
remaining ABC source port remain unfinished.


`CofactorPopulations` and `SemiCanonicalize` now port packed minterm counting
and `Abc_TtSemiCanonicize`'s active signature pass. Counts include stretched
small-table repeats and selected cofactor halves, unlike logical CountOnes.
Full mode orients output to at most half the packed minterms, each input to
zero-cofactor-heavy phase, then sorts zero counts ascending with C's nonstable
selection-sort ties. Permutation-only mode skips all phase changes. Owned
metadata uses source-indexed input phases and source-to-target permutation,
explicitly translating C's target-indexed phases/inverse permutation; sorted
counts remain target-indexed. Both forward and inverse configurations reconstruct
exactly. Dynamic vectors extend C's fixed 16-variable count/permutation buffers.
All 120 native cases through nine variables match initial/sorted populations,
normalized words, permutations and phase masks in both modes. An 18-variable
case checks phase/permutation alignment and source/result ownership. Independent
Boolean population/signature fuzzing verifies every transformed assignment,
metadata, actual result counts and source immutability; 289,799 executions
passed in 15 seconds. Full tests, race tests and vet pass; fixture hashes and
package-first headers are verified. This remains signature normalization,
not complete NPN canonicalization. Pair/cofactor refinement, alternative fast
count variants, exact/adaptive/hierarchical canonicalization, command integration
and the full remaining ABC source port remain unfinished.


`RefineAdjacentCofactors` and `CanonicalizeNPN` now port default
`Abc_TtCanonicize`/`Abc_TtCanonicizePerm` refinement. Adjacent candidates visit
C's Gray order 1,3,2,6,7,5,4, keeping the first strict high-word-first minimum;
configuration bits encode input phases before swap. The complete pipeline
starts from semicanonical signatures and follows C's five-round descending/
ascending sweeps and exact early stops. Equal-signature balanced pairs allow
phases/swaps, other pairs swap only; permutation-only mode never changes phases.
Metadata composes into source-indexed phases and source-to-target permutation.
Owned dynamic scratch replaces static buffers; cancellation/work budgets return
no partial result, and attempts/changes/rounds are reported. All 576 native
pair cases and 160 end-to-end cases through nine variables match configurations,
packed words, inverse metadata and both reconstruction directions. The helper
also cross-checks wrapper behavior including its smart swap-only branch.
An 18-variable case verifies the buffer-size extension. Unit tests cover
budgets, cancellation, invalid/nil inputs and ownership. Independent exhaustive
pair-candidate and assignment/configuration fuzz checks also verify unchanged
signatures and nonincreasing packed value, passing 254,834 executions in
15 seconds. Full tests, race tests and vet pass; fixture hashes/package-first
headers are verified. These are C's default heuristic results, not guaranteed
globally exact NPN representatives. General smart phase candidate selection,
exact/adaptive/hierarchical variants, other algorithms, command integration and
the full remaining ABC source port remain unfinished.


## Optimized adjacent truth-table refinement

`CompareAdjacentCofactors` and `RefineAdjacentCofactorsSmart` now port the
optimized `Abc_TtCofactorPermConfig` branch and its guarded wrapper. Comparisons
handle packed bit fields, adjacent 32-bit lanes across words and higher word
blocks, in forward or reverse word order. Equal/reversed cofactor numbers are
accepted safely. Smart selection preserves C's decision tree, source-phase-
before-swap configurations, nonidentity equal-value ties and rollback when a
candidate increases the packed truth value. The default `CanonicalizeNPN`
pipeline continues to follow C's naive Gray enumeration.

All 1,056 native cases through twelve variables match raw and guarded
configurations/results and all six pair-cofactor comparisons in both orders.
Independent Boolean-model fuzzing verifies comparisons, every assignment under
each returned configuration, rollback, swap-only restrictions and source
ownership: 159,788 executions passed in 15 seconds with widths through twelve
variables. An 18-variable case checks the dynamic-width extension. Invalid
indices, nil input and independent result storage are covered. Full race tests
and vet pass. Fixture hashes are pinned. This supersedes earlier
entries listing general smart candidate selection as pending. Exact/adaptive/
hierarchical canonicalization and the full remaining ABC port are unfinished.


## Individual and group input symmetry

`InputSymmetry` and `GroupSymmetry` now port `Abc_TtIsSymmetric` and
`Abc_TtIsSymmetricHigh`, prerequisites for adaptive canonicalization's symmetry
purging. The bit mask distinguishes direct swap symmetry, swap-plus-input-phase
symmetry, both, and neither. Groups use checked nonempty, disjoint equal-length
input lists. Optional source-indexed phase selections match C's nonzero
`symPhase` semantics, without exposing linked manager internals. All operations
use owned scratch and preserve the source. Immutable fixtures cover 3,664 cases
through ten variables, including differently sized equal-group pairs and four
phase-selection patterns. Independent assignment-model fuzzing through twelve
variables passed 244,374 executions in 15 seconds. Tests also cover all four
symmetry kinds, invalid/duplicate/overlapping indices, nil inputs, phase-vector
size and nonconstant 18-variable cases. Full race tests and vet pass; the added
wide boundary tests pass separately under race detection. Provenance hashes
are pinned. Tied-group creation/purging, adaptive/exact/hierarchical search and
the full remaining ABC port remain unfinished.


## Initial adaptive tied-group signatures

`CreateTiedGroups` now ports `Abc_TgCreateGroups` and its stable
`Abc_TgSplitGroup` dependency. Output polarity and input phases normalize packed
populations; source labels are sorted stably and partitioned by equal
zero-cofactor signatures. The truth permutation is deliberately deferred,
matching C's adaptive manager, and the result records source-indexed phase
metadata, target-order signatures, groups and phase-determination state. This
is distinct from the earlier semicanonicalizer's nonstable selection sort.
All 96 native cases through twelve inputs match words, phases, label order,
group ranges, populations and flags. Tests reconstruct the phase-normalized
truth, exercise nonfirst group splitting with preserved ties and unaffected
ranges, check owned storage, and cover nil input, explicit zero-input empty
partitions and an 18-input nonconstant case beyond C's fixed buffers.
Independent assignment/population-model fuzzing passed 256,795 executions
in a 15-second campaign (16 seconds reported by Go). Full race tests and vet
pass; fixture hashes are pinned. Adaptive manager state, symmetry
purging, permutation/phase enumeration and complete adaptive/exact/hierarchical
search remain unfinished, along with the broader ABC port.


## Adaptive symmetry reduction and chain expansion

`ReduceInputSymmetry` now ports `Abc_TgGroupSymmetry`, `Abc_TgPurgeSymmetry`
and `Abc_TgExpendSymmetry` following initial tied-group creation. It preserves
C's comparison-before-merge order, phase-only group flips, accumulated and
cleared symmetry weights, unsupported-input reverse chains, representative
compaction and signature-group ranges. Optional repeated equal-size group
comparisons find higher-order symmetries beyond individual-input merging.
The result owns its phase-normalized truth, source-indexed phases, deferred
expanded input order, representative groups, links and weights.

All 288 native cases through twelve variables match exact truth/configuration,
representatives, groups, support count, links, weights and expanded order.
Fixtures explicitly exercise higher-order merges using XORs of paired product
and sum terms. Independent Boolean-model fuzzing checks phase reconstruction,
exact support membership, acyclic/disjoint complete expansion and source
ownership: 297,787 executions passed in 15 seconds. Tests cover constants,
zero/nil inputs, independent storage and a nonconstant 18-input case. Full
race tests and vet pass; fixture hashes are pinned. Earlier
notes listing symmetry purging as pending are superseded. Adaptive permutation
application, group resets, enumeration, caches and complete adaptive/exact/
hierarchical search remain unfinished, along with the broader ABC port.


## Adaptive simple enumeration and applied permutation state

`CanonicalizeAdaptiveSimple` now connects stable tied-group creation and
symmetry reduction to expanded permutation application and C's simple search.
Owned permutation/inverse-position state tracks source-indexed phases through
individual/group swaps, including unequal-size chains and unused suffixes.
Pair candidates follow the exact Gray sequence; five rounds retain C's
descending/ascending group sweeps and weighted individual-input phase sweeps.
C's swap-only branch accepts a larger packed value, whereas phase candidates
minimize it. That direction is preserved and explicitly documented; a dedicated
regression checks the larger-valued swap. This is a native heuristic, without
monotonicity, idempotence or globally exact NPN guarantees.

All 288 native complete simple-search cases through twelve inputs match final
words, inverse C permutations and phase masks. Forward metadata reconstructs
the final table. Independent assignment/configuration fuzzing checks every
assignment, valid permutations, deterministic metadata/counters and unchanged
sources: 218,557 executions passed in 15 seconds. Tests cover unequal-size group
swap/phase involutions, zero/nil inputs, an 18-input nonconstant case, work
limits, cancellation at several stages and independent result storage. Full
race tests and vet pass; the final swap-direction regression passes separately
under race detection. Fixture hashes are pinned. Full phase/
permutation enumeration, adaptive threshold selection, group resets, caches
and complete exact/hierarchical algorithms remain unfinished, along with the
broader ABC source port.


## Full non-SCC adaptive group enumeration

`EnumerateAdaptiveGroups` now ports algorithm-0 `Abc_TgFullEnumeration`,
`Abc_TgPhaseEnumeration`, `Abc_TgSaveBest`, first/next permutation setup and
`Abc_NextPermSwapC`. Group permutations follow C's Johnson-Trotter adjacent
swaps, including terminal restoration and carry between signature groups.
First-group phase labels are sorted stably by accumulated symmetry weight,
then visited in Gray order; phase state carries between permutations as in C.
The first strict reverse-word minimum retains matching forward metadata.
Configurations/permutations/changes are reported; dynamic owned storage,
cancellation and configuration budgets replace fixed global scratch. The
native status bit is kept separate from input/output phase data.

All 288 native cases through twelve inputs match final words, permutations
and phases. Independent recursive group-orbit tests compare minimum values
and visit counts without using Johnson-Trotter or Gray traversal, while
immutable fixtures establish metadata tie order. Permutation traversal tests
visit every ordering through eight inputs with adjacent swaps, no duplicates
and exact terminal restoration. Independent orbit fuzzing passed 216,800
executions in 15 seconds. Tests cover zero/nil inputs, cancellation before/
during enumeration, work caps, ownership and an 18-input nonconstant case.
Full race tests and vet pass; fixture hashes are pinned. This
is full enumeration within the normalized signature/symmetry orbit, without
the balanced-output wrapper or an unrestricted brute-force NPN claim. SCC
branches, adaptive threshold selection, group resets, caches and complete
adaptive/hierarchical wrappers remain unfinished, as does the broader ABC port.


## Shifted-cofactor signatures for SCC-guided enumeration

`ShiftedCofactorCoefficients` now ports `shiftFunc`, `Abc_TtScc6/Scc` and
`Abc_TtSccInCofs6/InCofs`. Each packed set minterm contributes two to the power
of its set-bit count; zero-input cofactors collect the corresponding subset.
One traversal computes total and all cofactor signatures without modifying
the truth table or allocating cofactor tables. Small-table repeats contribute
weighted dummy inputs exactly as C. uint64 accumulation safely extends C's
signed-int signatures to every supported truth-table width. Applied adaptive
state now computes `Abc_TgCalcScc` group signatures, checks equal coefficients
within linked symmetry chains and optionally sorts values within signature
groups while preserving truth and representative order.

All 104 native table-signature cases and 288 native group-signature cases
through twelve inputs match. Independent assignment/product-weight fuzzing
passed 244,930 executions in 15 seconds. Closed-form constant tests through
twenty inputs verify powers of three, packed repeats, cofactor weighting and
a result beyond signed 32-bit range. Tests also cover zero/nil inputs, source
ownership and independent result vectors. Full race tests and vet pass;
fixture hashes are pinned. SCC-guided group splitting, cost
models, phase filtering and complete SCC search remain unfinished, together
with adaptive wrappers, caches and the broader ABC port.


## SCC phase filtering and permutation tie costs

`FilterAdaptivePhases` now ports `Abc_TgReorderFGrps`, `Abc_TgPermCostScc`,
`Abc_TgRecordPhase` and `Abc_TgRecordPhase1`. It supports total-only AdjSE,
lexicographic total-plus-group signatures, and permutation-cost-first
selection. First-group labels retain stable symmetry-weight order; a Gray
walk records every minimum mask in native visit order. Fixed signatures
retain C's empty phase vector. Costs preserve C's rounded factorial constants
through sixteen inputs and extend to larger groups. Result vectors are owned;
explicit limits/cancellation return no partial phase selection. Empty/zero-input
groups have a defined single-mask extension instead of C's invalid final flip.

All 540 native cases through ten inputs match mask order, phase-group labels,
phase count, candidate log2 count and permutation costs. Independent Boolean
assignment/model signature selection fuzzing passed 191,874 executions in
15 seconds. Tests cover nil input, invalid modes, negative/work limits,
cancellation during traversal, ownership, zero-input constants and an
18-input nonconstant case plus extended factorial cost. Full race tests and
vet pass; fixture hashes are pinned. SCC-guided group splitting
and permutation search, adaptive threshold dispatch, resets and caches remain
unfinished, along with the broader ABC source port.


## SCC-guided permutation search and stable cofactor splitting

`EnumerateAdaptiveSCC` now ports `Abc_TgSplitGroupsByScc`,
`Abc_TgPermEnumerationScc` and `Abc_TgPhaseEnumerationScc`. Each retained phase
configuration searches normalized signature-group permutations, optionally
splitting groups stably by cofactor coefficients and applying the resulting
expanded truth permutation. Mask order, adjacent traversal, first strict
minimum and forward configuration metadata match native behavior. The SCC
minimum starts with visited filtered candidates, preserving C's all-ones
seed semantics rather than retaining an unfiltered initial table. A shared
configuration budget covers phase signature visits and permutation candidates;
cancellation/exhaustion returns no partial result.

All 1,080 native cases through ten inputs match final words and inverse C
phase/permutation metadata across three phase filters, two symmetry modes
and both splitting modes. Independent Boolean-model recursive group-orbit
fuzzing verifies minimum values, candidate counts, source preservation and
forward reconstruction without using the implementation's group splitting
or permutation traversal: 172,804 executions passed in 15 seconds. Tests
cover invalid/nil inputs, cancellation in filtering/search, shared-budget
boundaries, zero-input constants, ownership and an 18-input nonconstant case.
Full race tests and vet pass; fixture hashes are pinned. Earlier
notes listing SCC-guided search/splitting as pending are superseded. Adaptive
threshold dispatch, group resets, caches and remaining hierarchical/exact
wrappers remain unfinished, along with the broader ABC port.


## Uncached adaptive threshold dispatch

`CanonicalizeAdaptive` now ports the uncached `Abc_TtCanonicizeAda` dispatcher,
`Abc_TgEnumerationCost` and algorithm-0/1 `Abc_SccEnumCost` fits. Encoded
thresholds select ordinary/SCC enumeration and high-symmetry reduction; native
strict cost comparisons and force-full thresholds choose complete simple,
full-group or total-SCC-filtered search. Forward transformation metadata is
retained, and full-enumeration status is separate from phase bits. Estimation
signatures and search visits share a configuration budget; cancellation or
exhaustion returns no partial result. Zero-input constants preserve their
output phase, and SCC all-unused functions use an explicit valid constant
result instead of C's empty-phase-vector path.

All 1,780 native cases through ten inputs and sixteen thresholds match fitted
integer cost, chosen algorithm/path, final words and phase/permutation
metadata. Independent assignment/configuration fuzzing verifies permutations,
threshold decisions, deterministic results/counters and unchanged sources:
172,749 executions passed in 15 seconds. Tests cover nil/invalid inputs,
cancellation during estimation/search, exact shared-budget boundaries,
constant extensions, ownership and an 18-input nonconstant case. Full race
tests and vet pass; fixture hashes are pinned. Immutable fixtures
exclude SCC all-unused inputs and explicitly document why. Cached/hierarchical
branches, cost-aware resets/dispatch and remaining algorithms stay unfinished,
along with the broader ABC source port.


## Uncached cost-aware search and representative reset

`CanonicalizeCostAware` now ports uncached `Abc_TtCanonicizeCA`,
`Abc_TgResetGroup`, `Abc_SccPhaseCost` and the algorithm-2 enumeration fit.
High-symmetry reduction and the native simple prepass precede signature phase
filtering. If enabled and favored by the strict fitted-cost comparison,
representatives sort into one group, linked symmetry/phase weights remain
intact, and cost-first phase filtering replaces the original selection.
Split-SCC permutation enumeration retains native mask, traversal and tie
behavior. Shared work limits count the prepass, both filters when broadened
and final candidates; cancellation/exhaustion yields no partial result.
Constants retain reconstructible phases instead of empty-search failures.

All 202 native cases through ten inputs match final truth/configuration,
fitted estimates and broadened status. A deliberately included eight-input
cycle exercises the hard-search/reset branch. Independent assignment-based
fuzzing checks transformations, permutations, fitted decision boundaries,
deterministic results/counters and unchanged sources: 140,296 executions
passed in 15 seconds. Tests also cover reset invariants and metadata, invalid/
nil inputs, cancellation at several stages, exact combined budget boundaries,
ownership, zero/all-unused constants and an 18-input nonconstant case. Full
race tests and vet pass; fixture hashes are pinned. Earlier
notes listing cost-aware reset/dispatch as pending are superseded. The unused
ResetGroup1 variant, cached/hierarchical canonicalization and remaining APIs
stay unfinished, along with the broader ABC source port.


## Hierarchical representative cache storage

`TruthHierarchyCache` now ports `Abc_TtHieManStart/Stop` and
`Abc_TtHieRetrieveOrInsert` using owned paged `WordMemory`. Zero to five levels,
negative level aliases, stable per-level insertion IDs, intermediate pending
entries, early representative hits, final lookup semantics and earlier-stage
linking match C. Explicit `Begin` traversals prevent stale current IDs from
linking unrelated paths; only visited earlier levels are updated. Pending
entries left by interruptions are safely reusable. Results own truth storage,
clones own memory/link/path state, and Close releases the cache. Per-level
entry/word limits use page sizes adapted to record width and storage bounds.
The cache is single-traversal and requires external synchronization if shared.

All 306 native calls across one through five levels and zero/three/six/nine
inputs match status, representative words and level sizes. Independent stateful
map-model fuzzing verifies completed and abandoned paths, representatives and
sizes: 266,575 executions passed in 15 seconds. Tests cover nil/closed/disabled
caches, invalid dimensions/levels/truth, Begin requirements, interrupted retry,
unvisited stages, independent clones/results, insertion-limit atomicity and
18-input word-budget boundaries. Full race tests and vet pass; fixture hashes are pinned. Cached adaptive/cost-aware algorithm integration,
complete hierarchical canonicalization and the broader ABC port remain
unfinished.


## Cached hierarchical nonexact canonicalization

`CanonicalizeHierarchical` now ports `Abc_TtCanonicizeHie` with `fExact=0`.
Output polarity, input phases, selection-sort permutation, native-default
refinement and final insertion visit the cache in C's exact stage order.
The default refinement loop is shared with `CanonicalizeNPN`. Five cache
levels retain every stage; shallower caches preserve C's earlier terminal
representative behavior. Cache hits return owned class representatives without
unavailable source-specific transformation metadata. Disabled/closed caches
compute the heuristic directly. Work limits and cancellation return no partial
representative; pending cache stages retry safely after Begin.

All 1,536 native calls across zero to five levels and zero/three/six/nine
inputs match exact returned truth and stage sizes. Repeated functions and
output/input-phase/permutation variants exercise early hits. Independent
unrestricted small-input NPN-orbit fuzzing checks that all returned
representatives remain equivalent to the source across sequences and cache
depths: 76,937 executions passed in 15 seconds. Tests cover nil/wrong-width
inputs, cache hit levels, closed/disabled caches, cancellation/budget retries,
result ownership, cache capacity errors and an 18-input nonconstant case.
Full race tests and vet pass; fixture hashes are pinned.
The exact hierarchical group-minimization branch, cached adaptive/cost-aware
integration and the broader ABC source port remain unfinished.


## Cached adaptive dispatch

`CanonicalizeAdaptiveCached` now ports `Abc_TtCanonicizeAda`'s cached branch.
Raw input, phase-normalized truth, applied symmetry reduction, simple
refinement and final representative use native -5/-4/-3/-2/-1 aliases.
Cache hits return early with matching cache/full-search status. Deeper caches
run the native simple prepass even when full search is selected; one-level
caches follow the uncached branch and stay empty. Shared fitted-cost code
keeps uncached/cached threshold decisions consistent. Result representatives
are owned and omit an unavailable source-specific NPN mapping. Early hits
report cost as unknown instead of implying a fitted estimate was computed.
A shared budget counts estimation, prepass and full search; errors return no
partial result and pending stages retry safely.

All 4,608 native calls across zero to five levels, four encoded thresholds
and zero/three/six/nine inputs match words, cache/full flags and stage sizes.
Repeated and transformed functions exercise multiple early equivalences.
Independent unrestricted small-input NPN-orbit fuzzing verifies class
equivalence and source preservation across input sequences/cache depths:
50,729 executions passed in 15 seconds. Tests cover invalid/nil inputs,
closed/one-level caches, exact combined-budget boundaries, pre-estimate
hits, cancellation/retry, owned results, cache capacity failures, constant
extensions and an 18-input nonconstant case. Full race tests and vet pass;
fixture hashes are pinned. Cached cost-aware integration, exact
hierarchical minimization, balanced-output wrappers and remaining algorithms
stay unfinished, along with the broader ABC source port.


## Cached cost-aware canonicalization

`CanonicalizeCostAwareCached` now ports the remaining cached
`Abc_TtCanonicizeCA` path. Raw truth, phase normalization, applied symmetry
reduction, simple refinement and final representative use native negative
stage aliases. Hits return before cost filtering with matching hash-only
status. Misses restore the reduced state after the simple prepass and run the
shared cost-aware search, retaining fitted estimates, broadened reset decisions
and full/hard flags separately from phase metadata. Unlike adaptive dispatch,
one-level caches store the final truth. Owned representatives omit unretained
source-specific NPN mappings. A shared budget counts actual prepass/filter/
search visits, with safe cancellation and pending-stage retry.

All 2,312 native calls across zero through five levels, both broadening modes
and zero/three/six/nine inputs match words, cache/full/hard flags and storage
sizes. The eight-input cycle explicitly exercises broadened cached search.
Independent unrestricted small-input NPN-orbit fuzzing verifies representative
equivalence, fitted decisions and source ownership across transformed-input
sequences: 48,426 executions passed in 15 seconds. Tests cover nil/invalid
inputs, closed/one-level caches, exact combined-budget boundaries, early hits
without estimates, cancellation/retry, cache-capacity errors, constants and
an 18-input nonconstant case. Full race tests and vet pass; fixture hashes are pinned. Earlier cached cost-aware pending notes are
superseded. Exact hierarchical minimization, balanced-output wrappers, the
unused ResetGroup1 variant and remaining algorithms stay unfinished, along
with the broader ABC source port.


## Exact-search permutation and Gray schedules

`TruthPermutationSchedule` now ports the schedule-building dependency from
`bool/lucky/luckySimple.c`: Johnson-Trotter adjacent swaps and reflected Gray
input flips. It preserves native consumption order while exposing owned
forward-order snapshots; garbage collection replaces C free helpers. Checked
factorial/power counts, combined step budgets and integer payload bounds reject
oversized schedules before allocation, avoiding unchecked C factorial growth.
The default cap is one million stored steps; zero-input schedules are valid.

Native schedules match for zero through eight inputs. Independent traversal
tests verify every permutation and phase assignment exactly once; the
eight-input case covers 40,320 permutations and 256 phase masks. Tests also
cover adjacent-swap/phase index ranges, exact budget boundaries, huge-count
rejection, nil/zero-value accessors and snapshot ownership. Schedule/budget
fuzzing passed 186,634 executions in 15 seconds. Full race tests and vet pass;
fixture hashes are pinned. Source coverage now marks this file
complete through the later minimizer and hierarchical checkpoints. The broader
ABC source port remains unfinished.


## Exact unrestricted and adjacent-group truth minimization

`MinimizeNPN` and `MinimizeGroups` now port `simpleMinimal` and
`simpleMinimalGroups` from `bool/lucky/luckySimple.c`, including reverse-word
minimum comparisons and mixed-radix carry transitions. Unrestricted search
covers every input permutation/phase and both output polarities. Group search
keeps adjacent group order fixed; input phases are always explored inside
multivariable groups, while the native fFlipInput flag controls singleton
phases only. Schedules are shared by sizes actually used and have a combined
allocation cap. Candidate budgets/cancellation return no partial result; owned
representatives do not invent source transformation metadata absent from C.
The single-input unrestricted case skips C's invalid swap with dummy input 1,
preserving stretched padding and the actual NPN orbit.

All 40 native unrestricted cases and 320 grouped cases through nine inputs
match exact words, including explicit group permutations across the 64-bit
word boundary. Independent recursive Boolean-model minimization fuzzing checks
both global and restricted-group minima: 85,681 executions passed in 15
seconds. Tests cover zero/single inputs, invalid partitions, shared schedule
bounds, exact candidate limits, cancellation and source/result ownership. Full
race tests and vet pass; the subsequently expanded word-boundary fixtures also
pass targeted race tests. Fixture hashes are pinned. Exact hierarchical
integration is covered by the following checkpoint. Balanced-output policy
belongs to that caller rather than this translation unit. All definitions in
`luckySimple.c` are accounted for; the broader ABC source port remains
unfinished.


## Exact hierarchical group minimization and no-op traversal bounds

`TruthHierarchicalOptions.Exact` now ports `Abc_TtCanonicizeHie`'s exact
branch. Refined adjacent signature groups retain native balanced-output and
balanced-singleton phase policy before grouped minimization and final cache
insertion. Exact-stage work is reported separately; early representative hits
do not claim a stored search mode or source transformation. Schedule, candidate
and transition bounds plus cancellation return no partial representative and
leave pending cache stages safely retryable. This is exact within C's selected
groups/phase policy, not a claim of unrestricted NPN minimization.

Grouped minimization now bounds mixed-radix transitions separately from
compared candidates. Disabled singleton phases can produce many carry steps
without comparisons; those steps now check limits and cancellation too. An
all-singleton eight-input test compares one candidate while visiting exactly
256 transitions, with the preceding 255-step limit rejected safely.

All 1,920 native exact hierarchical calls across zero to five levels and
zero/one/two/four/six inputs match words and cache sizes. Independent recursive
Boolean-model group minima and unrestricted small-input NPN-orbit fuzzing
passed 61,333 executions in 15 seconds. Tests cover exact-stage limit retries,
cache-hit provenance, owned results, invalid bounds, cancellation and no-op
transition boundaries. Full race tests and vet pass; subsequently expanded
single-input immutable fixtures also pass targeted race tests. Fixture hashes are pinned. Earlier exact hierarchical pending notes are
superseded. Balanced-output wrappers, unused ResetGroup1, remaining APIs and
the broader ABC source port remain unfinished.


## Balanced-output canonicalization wrapper

`CanonicalizeBalanced` now ports `Abc_TtCanonicizeWrap` with concrete
hierarchical/adaptive/cost-aware dispatch and no Go interfaces. Precisely half
packed population triggers original-then-complement evaluation using the same
cache; otherwise one run is used. Reverse-word comparison selects a strictly
smaller complement result, retaining the original on ties. Returned class
representatives omit unretained source-specific NPN mappings. Selection and
selected-run cache/full/hard metadata are explicit; checked counters sum both
runs. Selected algorithm limits apply independently per polarity, so total
work is bounded by at most twice those limits. Cancellation/failure in the
second run yields no partial result and leaves completed cache stages reusable.
C hierarchical functions return zero status flags even on cache hits; Go keeps
their actual lookup status while preserving their unmarked full-status bit.

All 4,612 native calls match representative words, evaluation/selection order,
cache sizes and encoded adaptive/cost-aware status flags. Coverage spans zero
to five cache levels, hierarchical heuristic/exact modes, four adaptive
thresholds and both cost-aware modes, including 28 genuine complement wins
and a broadened cycle. Independent NPN-orbit/direct-polarity comparison fuzzing
passed 57,677 executions in 15 seconds. Tests cover per-polarity exact limits,
ties, second-run cancellation/retry, owned results, invalid modes/options/widths
and unbalanced constants. Full race tests and vet pass; fixture
hashes are pinned. Earlier balanced-wrapper pending notes are superseded.
Unused ResetGroup1, specialized population variants and remaining algorithms
stay unfinished, along with the broader ABC source port.


## Recursive Fast logical cofactor populations

`CofactorPopulationsFast` now ports `Abc_TtCountOnesInCofsFast`,
`Abc_TtCountOnesInCofsFast_rec` and `Abc_TtCountOnesInCofsFast6_rec`. Recursive
word/subword splits accumulate zero-cofactor counts, using full-zero/full-one
block shortcuts and byte leaves without allocating cofactor tables. C Fast
counts logical assignments below six inputs, unlike the default packed-repeat
API; the separate result type makes this distinction explicit. uint64 counts
cover supported widths and safely extend C's signed integers. Zero through
two inputs extend C's assertion with valid logical counts. Count vectors own
storage, and source tables stay unchanged.

All 80 native cases for three through twelve inputs match total and cofactor
counts. Independent logical-assignment models and packed-scaling fuzz checks
passed 246,869 executions in 15 seconds. Tests cover nil inputs, zero/small
constants, sparse and nearly constant twenty-input tables, mixed subword/
multiword blocks and independent result storage. Full race tests and vet pass;
fixture hashes are pinned. Slow/Quick specialized counters,
unused ResetGroup1, remaining APIs and the broader ABC source port remain
unfinished.


## Corrected Quick packed cofactor populations

`CofactorPopulationsQuick` now ports `dauCount.c`'s byte-lane lookup and
recursive counter. A generated read-only table matches all 8,192 native values;
explicit numeric byte extraction avoids host byte-order and aliasing dependence.
Two separately accumulated half-blocks keep packed totals below 256 before
decoding, correcting native total-byte overflow. Additive recursive cofactor
updates retain earlier sibling contributions that C overwrites. uint64 counts
extend signed native accumulators. Positive widths preserve packed-repeat
counts; zero inputs follow C's single logical bit. Results own count vectors.

All 104 pinned cases through twelve inputs preserve native outputs alongside
independent bit-by-bit references. Nine native totals and twenty-one cofactor
arrays are incorrect; Go matches every reference count while retaining the
lookup algorithm. Independent assignment, packed-count and logical-scaling
fuzz checks most recently passed 695,560 executions in 10 seconds. Tests also
cover the full-eight-input overflow, mixed twenty-input recursion, nil input,
zero/small conventions, ownership, all 8,192 lookup lanes and stable
formatting. An owned table snapshot and language-neutral formatter replace the
stdout C-initializer generator; deterministic unit tests and the independent
fuzz model replace its process-global randomized drivers. Dataset hashes are
pinned. Coverage now marks the complete `dauCount.c` source validated. Slow
counting belongs to `dauCanon.c`; it is now implemented independently. The
uncalled ResetGroup1 helper and diagnostic-only routines have explicit
dispositions, so that source is validated. The broader ABC source port remains
unfinished.


## Small truth-table/GIA construction and output evaluation

`WithTruthTable6` now ports `Gia_Truth5ToGia`/`Gia_Truth6ToGia` into an owned
copy of a GIA manager. Highest-dependent-input Shannon cofactors recurse in
native order; raw three-AND muxes and strashed muxes preserve exact root/node
IDs and fanin order. Bound inputs can be phased signals, aliases, internal
gates or constants. Raw equal-ID AND cases that C cannot represent fold safely.
Limits/cancellation expose no partial graph/root and preserve source/truth
storage. `OutputTruthWord6` ports simple/mapped six-slot output truth using
iterative postorder, with explicit constants, buffers, native XOR/mux controls,
terminal roles and reachable-CI slot validation. Cone/scratch limits bound work.

All 624 native cases match root literals, object/fanin structure and simple/
reversed-slot output words across five/six-input, raw/hashed and phased/reversed
binding modes. Independent assignment-based composition/evaluation fuzzing
passed 96,182 executions in 15 seconds, including aliased/constant bindings
and mixed native-node sources. Tests cover a 10,000-node cone, invalid widths/
literals/maps, unreachable extra CIs, graph/scratch caps, cancellation and
source ownership. Full race tests and vet pass; fixture hashes
and current source identity are verified. General cut/LUT truth
managers, support collectors, optimized CNF/cut integration and the broader
ABC source port remain unfinished.


## Ordered structural cut truth and postorder collection

`CutTruth` now ports `Gia_ObjCollectInternalCut_rec/CollectInternalCut` and
`Gia_ObjComputeTruthTableCut`. Iterative fanin-first traversal returns native
non-leaf postorder and evaluates arbitrary-width packed truth over ordered
leaf slots. Repeated leaves use their first slot; later duplicates and unused
leaves retain declared input positions. Virtual elementary truths avoid
separate leaf-table allocations. Owned internal scratch/result storage replaces
shared graph numbers and truth buffers, preserving valid object marks/values.
Constant signals remain fixed; root-as-leaf, CO roots, buffers and native
XOR/mux nodes are safe extensions. Invalid boundaries and uncovered CIs reject.
Internal-node/mapping/combined-word limits and cancellation expose no partial
table or order. New graph-truth APIs also reject nil managers safely.

All 136 native cases match exact postorder and packed words through eleven
leaf slots, including internal boundaries, reversed/repeated/unused leaves
and root polarity. Independent free-boundary Boolean assignment/postorder
fuzzing passed 183,023 executions in 15 seconds. Tests cover a 10,000-node
cone, preserved scratch metadata, exact storage bounds, uncovered inputs,
invalid/constant/CO leaves, trivial/constant/CO roots, cancellation and result
ownership. Full race tests and vet pass; fixture hashes are
pinned. Reusable truth managers, cut enumeration/LUT mapping, optimized CNF
integration and the broader ABC source port remain unfinished.


## Batch GIA cut enumeration

`EnumerateCuts` now ports the native batch AND/XOR path: ordered leaf unions,
64-bit signatures, dominance filtering, normalized truth literals and stable
truth IDs, optional support compaction, whole-graph reference costs, native
newest-first ties, retained-cut caps and unit-cut policy. Owned storage and
graph/cut/pair/truth limits preserve source scratch and return no partial result
on failure or cancellation. Buffer/mux inputs reject unsupported native merge
preconditions.

All 400 native configurations match per-object cuts, ranks, signatures,
functions, complete interned truth records and all five statistics. Cases span
sizes 2–8, limits 2–25, truth/support modes, signature collisions above 64 inputs
and true eight-input multiword functions. Independent exhaustive source
assignment checks cover four-input fuzz graphs and all 256 assignments of an
eight-input graph. A native-confirmed permanent regression in
`test_vectors/gia_cut_enumeration_regression.json` records correlated internal
leaves: XOR and structural-boundary OR agree on every realizable source
assignment. The fuzzer explicitly loads this permanent seed.

Cut enumeration fuzzing passed 63,282 executions in 15 seconds after correcting
the independent-leaf oracle for correlated functional cuts. Full race tests
and vet pass. Fixture hashes are pinned.
`PORT_COVERAGE.json` now records 46 partially ported source files and 1,711
unstarted files; no source file is claimed complete. Incremental/reusable cut
managers, LUT mapping, optimized CNF, full PDR parity, the ABC command layer and
the Ivy replacement executable remain unfinished.


### Cut truth metadata validation

Truth lookup now checks exported cut-size metadata against the private truth
record width before reading/slicing records. Mutating a result from a two-input
cut to a seven/eight-input cut previously permitted an out-of-range slice;
lookup now returns `ErrDimension`. Regression tests cover both output phases
and every different-width transition among two, seven and eight inputs, and
verify that the original result remains usable. All native enumeration cases
continue to pass. Permanent cut fixture hashes and package-first Go headers
were rechecked; vet passes.


## Fixed-graph GIA cut-store lifecycle

`NewGiaCutStore`, `Advance`, `Finish` and `Close` now provide owned lifecycle
operations corresponding to `Gia_StoAlloc/Free/RefObj/ComputeCuts*` for a fixed
graph. Construction validates and snapshots the graph, initializes complete
reference costs and constant/CI cuts in native order. Each advance computes
one AND/XOR node; finish computes remaining nodes and transfers the result
exactly once. Explicit close releases abandoned work. Cancellation, pair/cut/
truth limits and computation failure discard private partial state, retain the
failure for subsequent calls and release the callback and graph storage.
The synchronous batch method shares this engine without a snapshot copy.

All 400 native reference configurations now also exercise interrupted and
resumed stores, comparing complete ranked cuts, statistics and stable truth
IDs. Lifecycle tests cover source mutation isolation, repeated completion,
double transfer, nil/zero/empty stores, cancellation between and within nodes,
resource failure and close. Fuzzing with random pause counts passed 43,313
executions in 15 seconds, retaining exhaustive Boolean checks and the permanent
correlated-cut regression. Live-graph append/reference updates, reset and
allocation reuse remain unported; this API does not claim those capabilities.

Final full race tests pass (22.336 seconds), and vet passes using a writable
`/tmp` build cache. Current C source identity and package-first headers were
verified after the refactor.


## CI structural support and six-input truth

`CITruth6` ports `Gia_ObjComputeTruth6CisSupport_rec` and
`Gia_ObjComputeTruth6Cis` with iterative owned traversal and sorted CI object
IDs. An explicit Fits flag disambiguates native zero for oversized support
from a false function; oversized support remains complete. Truth inputs use
structural support, preserving inputs even when logic cancels them. Constants
inside cones, buffers, native XOR/mux and CO roots are safe extensions. Reached
object and arena-scratch caps plus cancellation expose no partial result and
preserve source marks, values and graph storage.

All 144 native cases match sorted IDs and words across zero through eight
inputs and both phases, including oversized seven/eight-input support. Tests
independently propagate support masks and exhaustively evaluate source CI
assignments, verify packed repetition, a 10,000-node chain, source scratch
preservation, limits, cancellation and owned result slices. Fixture binary and fixture hashes and current/reference C source identity are pinned.
Reusable truth managers, LUT mapping and the broader port remain unfinished.

CI-truth fuzzing passed 216,672 executions in 15 seconds. Full race tests
pass (22.382 seconds), and vet passes using the writable `/tmp` build cache.


## Ordered six-input boundary truth

`BoundaryTruth6` ports `Gia_ObjComputeTruth6` and its cone-evaluation helper
using checked cut traversal and variable embedding. Repeated leaves retain
their last slot, matching repeated C scratch writes; unused and earlier
duplicate slots remain in the declared input space. The separate general
`CutTruth` API retains its native first-slot convention. At most six leaves
are accepted. Constants within cones, phased/CI/CO roots and native
buffer/XOR/mux nodes are safe extensions. Invalid/uncovered boundaries, caps
and cancellation reject without changing source scratch.

All 96 native cases match complete packed words, including CI/internal/root
boundaries, both root phases and reversed/unused/duplicate slots. Independent
free-boundary evaluation checks every one of 64 assignments for fuzz-generated
slot orders over extended-node cones. Fuzzing passed 195,853 executions in
15 seconds. Full race tests pass (22.420 seconds), and vet passes. Fixture,
fixture and current source hashes are pinned. Automatic LUT
boundary lookup, mapping storage/selection and the broader port remain
unimplemented.


## Packed LUT mapping storage, lookup and levels

`GiaLUTMapping` now owns the native object-offset header and fanin-count/
ordered-fanin/signed-footer records. Construction preserves insertion order;
iteration follows object IDs. Query results and packed exports own storage.
Checked import caps all packed entries, validates signed native integers,
offsets, lengths, topological fanins and disjoint referenced records, while
allowing unreferenced stale records from native root replacement. Constant
object-zero records remain queryable and are skipped by native iteration.

Native LUT/fanin/max-size statistics retain the empty -1 maximum. Level
calculation retains unmapped zero levels, object-order LUT levels, CO driver
levels and maximum output depth, excluding deeper dangling logic. Stored
boundaries connect to six-input truth evaluation with native repeated-slot
behavior; graph counts and mapped-root types are checked. Mapping identity
uses the source object-ID arena and must be rebuilt after graph edits.

All 20 native cases match packed arrays, statistics, every object level and
truth words from both native LUT evaluators. Tests cover malformed offsets,
lengths, overlapping records, cyclic/invalid fanins, duplicate roots, budgets,
input/export/record ownership and nil/mismatched graphs. Fixture
and current source hashes are pinned. The coverage inventory now
contains 47 partial source files and 1,710 unstarted files. Selection and
optimization algorithms, timing/box-aware levels, alternate mapping2, graph
edit propagation and the broader port remain unfinished.

Packed mapping fuzzing passed 342,557 executions in 15 seconds. Full race
tests pass (22.276 seconds), and vet passes using the writable `/tmp` cache.


## Mapped references and structural LUT overlap

`GiaLUTMapping.Analyze` ports `Gia_ManSetRefsMapped`, `Gia_ManSetLutRefs`
and both native overlap paths through a shared iterative implementation.
Owned references count every CO and each ordered mapped fanin slot, including
duplicates, unused slots and dangling LUTs. Per-root cone sizes count shared
structural nodes once per cone. Signed overlap subtracts the complete source
AND count and intentionally remains negative for partial mappings. Private
epoch marks and reusable stack preserve source marks/values. Constants,
buffers and native XOR/mux are safe extensions; uncovered CIs reject.
Graph and aggregate cone-visit limits and cancellation expose no partial
arrays.

All 30 frozen cases match both compatibility arrays, both overlap algorithms
and each per-root cone size, covering positive/negative overlap and partial/
empty/duplicate-fanin mappings. An independent reverse-topological model
checks references and reachability, including extended nodes. Tests cover a
10,000-node chain, exact work caps, malformed boundaries, cancellation and
source/result ownership. Fuzzing passed 254,879 executions in 15 seconds;
vet passes. Fixture hashes are pinned and checked without external tools.
Mapping selection, optimization and the broader port remain unfinished.

Final full race tests pass (22.192 seconds). Fixture checksum and package-
first Go headers were verified after the analysis implementation.


## LUT mapping connectivity verification

`GiaLUTMapping.Verify` ports native recursive/wrapper connectivity checks
with owned iterative marks. It starts at CO drivers and every buffer driver,
including dangling buffers, and follows mapped fanin records. Reached non-
buffer gates require mapping; constants, CIs and buffers terminate each
dependency path. Unrelated dangling gates may remain unmapped. Incomplete
mapping returns false; malformed graphs/options, graph/visit limits and
cancellation return errors. Verification preserves source scratch and does
not replace structural-boundary/function checks in Analyze/Truth6.

All 128 native cases match across every subset of four mapped gates, direct
versus full-CI boundaries, dangling buffers and CI/gate-driven outputs. An
independent bottom-up model checks dependency availability. Tests cover
buffer chains, a 10,000-node dependency chain, exact work caps, cancellation,
empty/nil/mismatched graphs and source ownership. Fuzzing passed 249,906
executions in 15 seconds. Full race tests pass (23.392 seconds), and vet
passes. Immutable fixture and current source hashes are
pinned. Mapping selection, transfer, optimization and the broader ABC port
remain unfinished.


## LUT mapping transfer

`GiaLUTMapping.Transfer` ports `Gia_ManTransferMapping` without modifying
shared graph Values. It preserves source-root order, skipped constant/dropped
roots, removed dropped fanins, phase-discarded fanin/footer IDs and mux signs.
Root collisions retain last-write behavior and native stale packed records.
Complemented surviving roots, invalid IDs, dropped surviving footers and
non-topological fanins reject. Output storage budgets include stale records;
cancellation and errors expose no partial result. Correspondence functional
correctness remains the caller's obligation.

All 32 native cases match complete packed output across ID shifts, phases,
drops, collisions, mux footers and reversed source storage. Unit tests cover
invalid contexts, root phases, footer failures, exact budgets, cancellation
and source ownership. Fuzzing checks transferred records against an
independent overwrite model and checked packed re-import. Fixture binary, fixture and current source hashes are pinned. Mapping
selection/optimization, packing/timing transfer and the broader port remain
unimplemented.

Transfer fuzzing passed 168,219 executions in 10 seconds. Full race tests
pass (22.525 seconds), and vet passes.


## Owned LUT packing and transfer

`GiaLUTPacking` ports packing storage/statistics and `Gia_ManTransferPacking`
with concrete owned records tied to an immutable source mapping. The native
group count and one-to-three-member groups retain slot/group order and
repeated IDs. Decoding validates complete lengths/counts and mapped source
IDs; exports and group accessors own slices. Statistics preserve native group
size counts, LUT-slot totals and empty maximum -1 without printing.

Transfer discards correspondence phases and preserves aliases, requiring
source and target mapping membership for every packed member. Dropped packed
LUTs and invalid IDs reject instead of indexing invalid native sentinels.
Storage budgets and cancellation expose no partial packing. All 32 native
transfers match complete packed arrays across empty/mixed groups, repeats,
shifted IDs, phases and aliases. Unit tests cover malformed records, missing
members, limits, cancellation and ownership; fuzzing checks malformed import
and valid group substitution against an independent model. Fixture binary, fixture and current source hashes are pinned. Packing
selection/conversion, timing transfer and the broader port remain unfinished.

Packing fuzzing passed 180,940 executions in 10 seconds. Full race tests
pass (22.474 seconds), and vet passes.


## Native two-slot edges and packing conversion

`GiaEdgeAssignment` ports native two-slot insertion/count/lookup, pair-array
import/export and `Gia_ManConvertPackingToEdges`. It preserves duplicate and
self-edge slot use, zero sentinel behavior and independently skipped directed
insertions, including asymmetric rows on overflow. Pair export follows object
and slot order without repairing symmetry or deduplicating. Packing pairs add
one edge; triples add first-to-third then second-to-third edges in native
direction order. Skipped counts are returned without printing. Object/pair
budgets, endpoint validation and cancellation expose no partial assignment.

All 32 native cases match complete rows, exported pairs and overflow counts,
including zero endpoints and repeated packing members. Tests cover self-edge
slots, exact packing budgets, cancellation, invalid endpoints and source/
result ownership. Independent bounded-neighbor-list fuzzing passed 216,738
executions in 10 seconds. Full race tests pass (22.691 seconds), and vet passes.
Fixture and current source hashes are pinned. Coverage
now contains 48 partial files and 1,709 unstarted files. Edge mutation, delay
evaluation/optimization and the broader ABC port remain unfinished.


## Native no-box edge timing

`GiaEdgeAssignment.EvaluateDelay` ports `Gia_ObjEvalEdgeDelay`, the
plain-mapping/no-box branch of `Gia_ManEvalEdgeDelay`, and native edge counts.
Each mapped fanin contributes root-row-directed delay 2 or 10, preserving
asymmetry and zero-sentinel equality. LUTs with more than four slots reject.
Unmapped objects start at zero, CO delay-array entries remain zero, and the
maximum scans CO drivers rather than dangling LUTs. Edge counts truncate the
positive directed-slot count divided by two. Owned scratch, graph and total
fanin-work caps, checked addition and cancellation expose no partial result.

All 12 native cases match every delay entry, output maximum and edge count
across empty/partial/full mappings and fast-edge selections. Unit tests cover
width rejection, exact work caps, cancellation, invalid context and result
ownership. Fuzzing compares iterative results to an independent recursive
model on generated AND/XOR/constant cones. Fixture and current/
reference source hashes are pinned. Box timing, mapping2, edge selection/
optimization and the broader port remain unfinished.

Edge-timing fuzzing passed 168,500 executions in 10 seconds. Full race tests
pass (22.721 seconds), and vet passes.


## Greedy native fast-edge selection

`GiaLUTMapping.SelectEdges` ports `Gia_ObjComputeEdgeDelay` and the
plain-mapping/no-box branch of `Gia_ManComputeEdgeDelay`. It preserves
first-level skipping, fanin-order maximum/runner-up/tie tracking, one- or
optional two-critical-slot selection, capacity checks, duplicate-slot counts
and directed insertion order. LUT widths one through four are accepted;
empty/oversized LUTs reject. Owned rows/delays, graph/aggregate-fanin-work
caps, checked addition and cancellation expose no partial result.

All 24 native cases match full edge rows, every delay entry, CO maximum
and edge count, including both two-critical modes and duplicate fanins.
Tests separately cover distinct tied critical fanins, saturated fanout
capacity, exact work bounds, cancellation, source ownership and a 10,000-node
chain. Fuzzing reevaluates selected assignments through the independent
timing API. Immutable fixture and current source hashes
are pinned. Timing boxes, alternate mapping2, second/global edge optimization
and the broader ABC port remain unfinished.

Selection fuzzing passed 156,457 executions in 10 seconds. Full race tests
pass (22.374 seconds), and vet passes.


## Native reverse fast-edge selection

`GiaLUTMapping.SelectEdgesReverse` ports `Gia_ObjComputeEdgeDelay2` and
`Gia_ManComputeEdgeDelay2` for plain mappings. Descending LUT traversal
preserves zero initial distances, critical-fanout maxima and tie multiplicity,
capacity checks, duplicate-slot insertion and fanin propagation. Every mapped
LUT contributes, including dangling logic; the maximum scans CIs without CO
seeding. A distinct owned reverse-distance result avoids conflating these
unit distances with forward 2/10 timing. Graph/aggregate-fanin caps, checked
addition and cancellation expose no partial rows/distances or source edits.

All 12 native cases match full rows, every distance, CI maximum and edge
count across empty/partial/full mappings and duplicate slots. An independent
weighted fanout model checks generated AND/XOR/constant graphs. A 10,000-node
chain with no COs confirms dangling-logic semantics and iterative traversal.
Tests cover exact work caps, cancellation and source ownership. Fixture binary, fixture and current source hashes are pinned. Timing boxes,
mapping2, window optimization and the broader port remain unfinished.

Reverse-selection fuzzing passed 167,999 executions in 10 seconds. Full race
tests pass (22.780 seconds), and vet passes.


## Alternate mapping lists and critical-level timing

`GiaLUTNetwork` ports `Edg_ManToMapping`, unit timing in both directions and
`Edg_ManCollectCritEdges`. Conversion retains fanin order, duplicate slots
and fanouts in native root/slot order, with owned accessors. Empty fanin lists
are not LUTs, matching mapping2. Unit timing uses root-directed edge costs
0/1, global maxima over mapped LUTs including dangling logic, and zero CI
reverse entries. Critical groups retain object order at each forward level
and safely reject native F+R/positive-level invariant violations. Arena and
slot-work caps plus cancellation expose no partial network/timing.

All 12 native cases match full fanouts, both distance arrays, global maxima
and every critical group. Tests cover duplicate fanouts, accessor ownership,
exact combined-pass budgets, cancellation, width limits and malformed timing.
Independent recursive path models passed 165,129 fuzz executions in 10 seconds;
vet passes. The oracle helper's duplicate maximum key was corrected before
pinning the fixture; unit timing and ordinary LUT depth now have distinct
keys. Immutable fixture and current source hashes are
pinned. Mutable mapping2/window updates, the improvement loop, timing boxes
and the broader ABC port remain unfinished.

Final full race tests pass (22.808 seconds) with the corrected immutable fixture.
Fixture checksum and package-first headers were verified.


## Native unit-edge improvement loop

`GiaLUTMapping.OptimizeEdges` ports `Edg_ObjImprove` and
`Edg_ManAssignEdgeNew`. It starts empty, honors one/two-edge capacity, tries
critical fanins before fanouts and orders critical levels by population.
Stable Go sorting preserves deterministic equal-population order observed
in the pinned native reference. Strict delay decreases save backups; native
50-iteration stagnation and 10,000-iteration bounds are retained. The result
restores the last strict-improvement backup, discarding plateau insertions.
Native assertion preconditions reject safely. Shared work covers construction,
timing passes, candidate slots and final reevaluation; caller iteration caps
and cancellation expose no partial result. Source graph/mapping stay owned
and unchanged.

All 24 native cases match complete restored rows, forward/reverse arrays and
maxima across twelve deterministic twenty-gate graphs and both edge limits.
Tests cover empty/native-assertion preconditions, invalid options, iteration/
work limits and cancellation. Fuzzing passed 122,669 executions in 10 seconds,
checking result reevaluation, capacity, critical invariants and nonregression.
Full race tests pass (22.795 seconds), and vet passes. Fixture and current source hashes are pinned. Mutable window
updates/evaluation, timing boxes and the broader ABC port remain unfinished.


## Owned window proposals and full evaluation

`GiaLUTNetwork.WithWindow` ports mapping-list replacement as an owned proposal
instead of a shared temporary swap. Unique nonconstant roots, topological
fanin IDs, widths up to four, storage caps and cancellation are checked.
Empty lists remove LUTs. Both fanins and derived fanouts remain consistent,
unlike C's temporary fanin-only swap. Network `SelectEdges` shares the checked
greedy engine through an internal packed adapter; `EvaluateWindow` returns
the proposal's fresh rows/timing without changing the graph, source network
or replacement lists. This is explicitly the native mapping2-only, no-box
context. It computes timing and does not prove proposal Boolean equivalence.

All 12 native cases match full selected rows, timing arrays and maxima across
six two-node proposals and both two-critical modes, including removal,
flattening, duplicate and one-input boundaries. Tests cover invalid roots/
slots, width/topology errors, exact storage caps, cancellation and ownership.
Fuzzing independently rebuilds fanouts and proposed packed mappings for
comparison. Immutable fixture and current source hashes
are pinned. Incremental window evaluation, timing boxes and the broader port
remain unfinished.

Window fuzzing passed 144,294 executions in 10 seconds. Full race tests pass
(22.642 seconds), and vet passes.


## Owned directed edge-slot edits

`GiaEdgeAssignment.WithEdits` ports native private remove/clean helpers and
exposes ordered directed edit batches. Insertion preserves duplicate/zero
semantics and overflow history. Removal compacts slot two when slot one
matches, then clears slot two; missing neighbors reject the native assertion.
Clear ignores the unused neighbor field. Reciprocal rows stay unchanged.
Object/edit caps, skipped-count overflow preflight and cancellation expose no
partial result, preserving source state even after earlier successful edits.

All 156 native four-object cases match complete edited rows and overflow
counts, including arbitrary slot shapes and duplicate/zero neighbors. Tests
cover atomic rejection, invalid IDs/kinds, budgets and cancellation. Sequence
fuzzing verifies rows, overflow history and unchanged source assignments.
Immutable fixture and current source hashes are pinned.
Incremental window evaluation, timing boxes and the broader port remain
unfinished.

Slot-edit fuzzing passed 219,211 executions in 10 seconds. Full race tests
pass (22.800 seconds), and vet passes.


## Incremental packed-mapping edge recomputation

`GiaLUTMapping.RecomputeEdges` ports the recomputation phase of
`Gia_ManEvalWindowInc` in the native packed-priority context. It copies prior
rows/delays, reverses structural TFO postorder, skips unmapped gates and
clears/reselects each visited LUT row. Old reciprocal entries in other rows
remain, intentionally preserving native asymmetric/stale-slot behavior. The
native affected-LUT maximum and visit order are returned separately from
global CO-driver timing and edge counts. The full and incremental paths now
share one per-node greedy kernel. Bounds, malformed state and cancellation
expose no partial result or changes to source graph/mapping/prior state.

All 24 native cases match affected order, full rows, retained/updated delays,
affected maximum and edge count. Unit tests cover empty roots, result
ownership, invalid roots/state, work limits and cancellation. Fuzzing checks
untouched delays, affected/global maxima and input ownership. Fixture binary, fixture and current source hashes are pinned. Mapping2-aware
incremental proposal evaluation, timing boxes and the broader ABC port remain
unfinished.

Incremental fuzzing passed 153,487 executions in 10 seconds. Full race tests
pass (23.021 seconds), and vet passes, including the shared-kernel refactor.


## Timing terminal storage and traversal foundations

`TimManager` now ports native no-box CI/CO initialization, exact float32
arrival/required storage, eternity/default metadata, setters/getters, traversal
enable/disable, current checks and 30-bit epoch rollover. Every terminal is
a main PI/PO in this initial layer. Shared traversal marks preserve native
epoch-zero differences: CO required writes may repeat at zero; CI required
and CO arrival writes reject a current mark. Bulk required initialization
preflights all assertions, preventing partial writes. Scalar snapshots and
owned clones preserve storage and float bits; dimension/total-terminal caps
reject before allocation.

All 128 native cases match every terminal field and float bit across zero
through three CIs/COs and eight float patterns, including signed zero,
infinities, NaN payload and subnormal values. Tests cover rollover, disabled
traversal, invalid contexts, repeated writes, atomic bulk rejection and clone
ownership. Immutable fixture and current source hashes
are pinned. Coverage now contains 52 partial source files and 1,705 unstarted
files. Boxes, delay tables/propagation, box traversal operations, native
manager duplication/reduction/dumping and the broader port remain unfinished.

Timing-terminal sequence fuzzing passed 135,078 executions in 10 seconds.
Full race tests pass (22.398 seconds), and vet passes.


## Timing boxes, associations and traversal labels

`TimBox` records now provide native contiguous CO-input/CI-output ranges,
pin positions, insertion-order overlap behavior, table/copy/black metadata,
range queries and counts. Creation validates all dimensions and aggregate
box/pin budgets before association writes. Primary counts preserve native
first/last box boundaries. Bulk PI/PO initialization now skips boxed
terminals; clones own box records and terminal state. Current/previous box
pin labels match native arithmetic, including -1 at epoch zero. Boxed
arrival/required getters reject explicitly until propagation exists.

All 72 native states match complete records, terminal metadata/float bits,
primary/black counts and mixed traversal labels, covering disjoint/overlapping
ranges and all black flags. Tests cover atomic range/budget rejection,
cloning, ownership, previous labels and unsupported propagation. Native
fixture and current source hashes are pinned. Coverage
now contains 53 partial files and 1,704 unstarted files. Delay tables, box
propagation, graph timing integration and the broader port remain unfinished.

Box-association fuzzing passed 194,532 executions in 10 seconds. Full race
tests pass (22.859 seconds), and vet passes.


## Owned delay tables and boxed arrival propagation

`TimDelayTable` and its checked manager storage now own output-major float32
arcs, sparse table slots and disabled-arc values. Shape and aggregate storage
budgets preflight updates; accessors/clones own arrays. Box/table dimensions
are checked. `CIArrival` now evaluates all outputs of a box and marks them
and the box current, preserving native epoch-zero/current cache behavior and
stored-input semantics. C's integer maximum helper truncates float32 sums
toward zero; undefined nonfinite/out-of-int32 conversions reject before any
output/mark writes. Replacing a table does not invalidate cached outputs;
advance the epoch as in native traversal.

All 96 native cases match exact output bits and box epochs over one through
four inputs/outputs, fractional/negative sums and disabled arcs. Tests cover
sparse storage, dimensions/budgets, table/accessor/clone ownership, caching
and atomic undefined-conversion rejection. Immutable fixture and
current source hashes are pinned. The plan records the separate
native required-time disabled-arc indexing concern; required propagation,
box-library loading and graph timing integration remain unfinished.

Box-arrival fuzzing passed 191,772 executions in 10 seconds. Full race tests
pass (23.062 seconds), and vet passes.


## Corrected boxed required-time propagation

`TimManager.CORequired` now propagates all box input requirements with actual
[output,input] arc guarding. This intentionally fixes C's [output,output]
guard, which can skip active arcs, use disabled delays or read beyond a row
when outputs outnumber inputs. Native epoch-zero/current caching, stored
requirements, float32 subtraction and comparison-based minimum selection are
preserved, including IEEE NaN/tie behavior. All input and box marks commit
after validation; missing/mismatched tables expose no partial updates.

The permanent fixture records 128 corrected C matrix references and 80 safe
original native calls. Eleven original cases differ due to the guard bug;
the 48 unsafe shapes skip original calls entirely. Go matches every corrected
result bit. Tests cover a cross-arc regression, more outputs than inputs,
caching, cloning, shape rejection and early/late NaN selection. Fuzzing checks
finite column minima independently. Fixture and current/reference
source hashes are pinned. Graph timing integration, library setup and the
broader ABC port remain unfinished.

Required-time fuzzing passed 186,059 executions in 10 seconds. Full race
tests pass (22.618 seconds), and vet passes.


## Normalized GIA traversal with timing boxes

`IsNormalized` and `OrderWithBoxes` now port native CI-prefix/CO-suffix
checks and staged box traversal. Order includes constant/PIs, each box's
input cones, COs then output CIs, and primary-output cones/COs; dangling gates
are omitted. Iterative private traversal preserves source graph/timing marks
and includes extended gate/control dependencies. Normalized layout and
contiguous box partitions are checked; future-output dependencies, budgets
and cancellation expose no partial order.

The fixture contains 24 scenarios: 16 defined native calls match eight valid
orders and eight dependency rejections. Eight final-box dependency cases
skip C's crashing diagnostic (-2 box lookup used as an index), while Go
rejects safely. Tests cover a 10,000-node cone, ownership, exact entry caps,
normalization/range errors and cancellation. Dependency/reachability fuzzing
passed 162,981 executions in 10 seconds. Full race tests pass (22.915 seconds),
and vet passes. Fixture and current source hashes are
pinned. Choice siblings, graph timing evaluation integration and the broader
ABC port remain unfinished.


## Boxed graph edge-delay evaluation

`GiaEdgeAssignment.EvaluateDelayWithBoxes` now ports the plain-mapping
timing-box branch of native edge evaluation. It advances a private timing
clone, follows staged boxed order, truncates CI arrivals safely to int32,
evaluates mapped LUTs with shared signed 2/10 arithmetic and updates CO
arrival terminals. Graph CO delay entries remain zero; maxima scan all CO
drivers. Dangling LUTs remain unevaluated. No-box fallback retains ignored
terminal overrides and unchanged epochs. Combined table-copy/order/fanin/arc
budgets and cancellation expose no partial result or source changes.

All 48 native cases match complete graph arrays, maxima, edge counts and
CI/CO arrival bits across chained boxes, buffers, fast edges, negative/fractional
PIs and dangling gates. Tests cover exact combined budgets, cancellation,
undefined conversion rejection, no-box fallback and deep result ownership.
Fuzzing independently checks a boxed timing recurrence. Fixture
and current source hashes are pinned. Boxed edge selection,
alternate mapping2 timing and the broader port remain unfinished.

Boxed graph timing fuzzing passed 123,815 executions in 10 seconds. Full
race tests pass (22.732 seconds), and vet passes.


## Boxed graph greedy edge selection

`GiaLUTMapping.SelectEdgesWithBoxes` now ports the plain-mapping timing-box
branch of native greedy selection. Boxed selection/evaluation share staged
traversal, private timing cloning and work accounting; all selection paths
share the node-level tie/capacity kernel. Checked signed arithmetic supports
negative CI arrivals, and C's all-negative-candidate assertion rejects safely.
Selection starts empty, preserves possible critical edges to main CIs and
returns owned rows/timing/manager state. No-box fallback retains ignored
terminal overrides, all-mapped-node scanning and unchanged epochs.

All 32 defined native cases match complete selected rows, graph arrays,
maxima/counts and all CI/CO arrival bits across chained boxes, buffers, both
two-critical modes and signed/fractional arrivals. Tests cover exact combined
budgets, assertion rejection, no-box fallback and source ownership. Fuzzing
checks reevaluation, capacity and delay nonregression. Fixture and current source hashes are pinned. Mapping2 boxed timing
and selection, reverse graph integration and the broader port remain unfinished.

Boxed selection fuzzing passed 103,738 executions in 10 seconds. Full race
tests pass (22.825 seconds), and vet passes, including shared-path refactoring.


## Alternate-mapping boxed timing and selection

`GiaLUTNetwork.EvaluateDelayWithBoxes/SelectEdgesWithBoxes` now connect
immutable alternate fanin lists to checked boxed traversal/timing kernels.
Empty lists remain non-LUTs, and removed nodes stay zero in timing arrays.
Private packed adapters preserve mapping2 semantics without modifying caller
networks or timing managers. The immutable compatibility cases were captured
with packed mappings cleared so they exercise the alternate source branches.

All 48 native evaluations and 32 native selections match graph delays,
maxima/counts, selected rows and terminal float bits. Tests cover LUT removal,
window proposals, duplicate slots, invalid context/budgets/cancellation and
source ownership. Fuzzing compares against independently rebuilt packed
mappings. Fixture and current source hashes are pinned.
Incremental alternate proposals, reverse graph timing and the broader port
remain unfinished.

Alternate boxed timing fuzzing passed 83,950 executions in 10 seconds. Full
race tests pass (22.976 seconds), and vet passes.


## SAT-LUT unboxed arrival and required timing

`GiaLUTNetwork.UnitTiming` ports `Sbl_ManComputeDelay` and unboxed
`Sbl_ManCreateTiming`. It preserves zero CI arrivals, unit fanin costs,
delayStart lower bounds and global mapped-LUT maxima including dangling
logic. Required arrays start at 1e9; CO drivers are downdated by the maximum,
then every mapped LUT propagates required-1, including unobservable cones.
CO array entries retain initial values. Owned arrays, signed arithmetic
checks, combined slot budgets and cancellation preserve source state.

All 32 native runs match full arrival/required arrays and maxima across
full/partial/empty mappings and negative/positive starts. Independent recursive
path models check generated graphs. Tests cover a 10,000-node dangling chain,
exact combined budgets, cancellation and ownership. Fixture and current source hashes are pinned. Coverage now contains
54 partial files and 1,703 unstarted files. The plan records a boxed
reverse traversal-index reuse bug requiring correction before boxed SAT-LUT
integration. SAT remapping, boxed timing and the broader port remain unfinished.

Unit arrival/required fuzzing passed 153,616 executions in 10 seconds. Full
race tests pass (22.871 seconds), and vet passes.


## Corrected boxed SAT-LUT arrival/required timing

`GiaLUTNetwork.UnitTimingWithBoxes` now integrates staged graph traversal,
arrival propagation and reverse required timing over two private manager
epochs. It preserves native visited-LUT maxima/delayStart, integer conversions,
CO-driver required assignment, float terminal values and graph CO initial
values. Separate loop indices correct native reverse traversal clobbering;
required propagation also uses the corrected actual-arc guard. Combined
table-copy/two-pass-order/fanin/arc budgets, checked arithmetic and cancellation
expose no partial result or source changes. No-box fallback leaves epochs
and terminal overrides untouched.

All 48 corrected C-reference cases match full graph arrays and manager
arrival/required bits. Original native required arrays are retained and differ
in all 48 cases, demonstrating the traversal bug. Tests cover exact budgets,
cancellation, fractional required conversion, disabled arcs, cloning and
fallback. Fuzzing independently checks a boxed recurrence. Fixture and current source hashes are pinned. SAT remapping, window
solving and the broader port remain unfinished.

Boxed unit timing fuzzing passed 114,444 executions in 10 seconds. Full
race tests pass (23.208 seconds), and vet passes.


## Decoded SAT-window timing and conflict clauses

`GiaLUTNetwork.EvaluateUnitWindow` ports decoded unboxed/unit mapping
evaluation and critical-fanin selection. It retains original required times
and reference maximum, recomputes proposals in topological window order and
checks roots in caller order. The first violation produces a root-to-boundary
critical path and negated selection clause using window-variable indices.
Independent nodes may be reordered without confusing variables with GIA IDs.
Original/proposal/path work shares a budget; invalid context, assertion
preconditions and cancellation expose no partial result. Boolean validity
and candidate closure remain separate SAT-remapper obligations.

All 32 native runs match decisions, full timing arrays and clause order across
original/proposal variants, root order and delay bounds. Tests cover exact
work limits, reordered variables, invalid order and cancellation. A native
clause passed to the Go solver rejects all implicated selections and permits
a disabled implicated selection. Fuzzing checks independent proposal timing
and first-critical-fanin paths. Fixture and current/reference
source hashes are pinned. Cut decoding/CNF construction, edge-mode checking
and the full SAT remapper remain unfinished.

Decoded-window fuzzing passed 113,684 executions in 10 seconds. Full race
tests pass (22.908 seconds), and vet passes.


## Selected SAT-cut bitmask decoding

`GiaWindowCut` and `DecodeGiaWindowCuts` now port native current-mapping
decode. Destination indices refer to window nodes; two-word leaf and internal
node masks expand in native category/ascending-bit order. Raw duplicate/zero
IDs are preserved, and later cuts replace earlier lists at the same
destination. Unselected records are not interpreted. Empty selection, bad
indices/destinations/bits, input/work budgets and cancellation reject without
partial output. Topology/width/closure/Boolean validity remain separate checks.

All 144 native cases match complete windows through bits 63/64/127, full
masks, repeated values/selections and replacements. Tests cover atomic
rejection, exact bounds and ownership. Decoder-to-timing integration produces
the expected native conflict clause; fuzzing compares an exhaustive bit-scan
model and malformed-mask/work-limit behavior. Fixture and
current source hashes are pinned. Cut generation, CNF construction,
search and the broader port remain unfinished.

Cut-decoder fuzzing passed 102,656 executions in 10 seconds. Full race tests
pass (23.480 seconds), and vet passes.


## BMC counterexample-depth targets

`src/sat/bmc/bmcCexDepth.c` is validated across all six active definitions.
`EnlargeCounterexampleTarget` builds the bounded PO-zero target with the
source's reversed frame-major PI blocks followed by arbitrary initial-register
inputs. It semantically copies ordinary and extended GIA gates, carries each
RI into its matching RO between frames, preserves graph metadata, hashes the
result and returns a cleaned owned graph. `CounterexampleReachabilityTarget`
existentially removes every bounded PI, retains the register variables and
appends the resulting state predicate to the original transition system. It
returns the graph directly and does not retain the source's `miter3.aig` side
effect.

`BuildCounterexampleDepthCone` consolidates `Bmc_CexBuildNetwork2` and its
underscore variant behind the concrete `PruneImplied` option. It exposes one
polarity-adjusted variable per starting register, fixes the later PI suffix to
the expanded trace, transfers RI formulas to ROs, and preserves the exact
observed-one AND, double-zero OR and single-controlling-input recurrences. The
ordinary-AND-only recurrence explicitly rejects buffers, XORs and muxes.
`BuildCounterexampleDepthTargets` preserves descending nonzero start frames
and appends register-only cones; `AnalyzeCounterexampleDepth` composes the
existing verified care analysis and returns all analysis, cone and graph
values without writing `miter2.aig`.

The Go-only tests pin reversed-frame semantics, existential target behavior,
both pruning modes, descending frame selection, ownership and metadata,
extended-gate boundaries, invalid dimensions, limits and cancellation. A
10,000-gate test confirms iterative stack safety. A 5,000-execution fuzzer
independently simulates every register assignment accepted by a generated
cone and proves it reaches the selected failing output. No C oracle, cgo,
compiler or retained native helper is used. The complete suite passes in
15.758s, the race suite in 81.838s, and vet is clean. Inventory is now 249
validated, 39 partial and 1,469 unstarted files.


## BMC counterexample bad-state cuts

`src/sat/bmc/bmcCexCut.c` is validated across all eight active definitions.
`AnnotateCounterexampleStateCut` consolidates the recursive and nonrecursive
source walkers into one iterative owned annotation. It preserves concrete
values, variable initial registers, fixed primary inputs and constants,
RI-to-RO frame transfer, deterministic fanin-zero choice, and the union of
implied nodes with the selected final-property justification. It uses no graph
marks, scratch values or recursion.

`CounterexampleTargetStates` exposes the source's three constructions as a
concrete mode: the justified register cube at the selected end frame, one
terminal monotone cone, or one cone for every terminal frame through the
failure. Results own their annotation, begin-frame initial state, frame order
and combinational property. Sequential mode replaces the source properties
with the state predicate and phase-recodes the returned begin state to zero.
The AIG method uses the validated concrete GIA conversions in both directions.
Verbose printing is omitted, and ordinary-AND marking rejects extended gates
instead of silently assigning them the wrong recurrence.

Go-only tests pin exact value/fixedness/justification rows, deterministic
double-zero choice, all three construction modes, multi-output frame order,
sequential phase semantics, AIG conversion, fully implied terminal failures,
ownership, invalid ranges, resource limits and cancellation. A 10,000-gate
case confirms iterative traversal. A 5,000-execution fuzzer independently
replays every state accepted by generated cubes and terminal cones and proves
the selected failure remains true. The complete suite passes in 16.003s, race
in 81.717s, and vet is clean. Inventory is now 250 validated, 39 partial and
1,468 unstarted files.


## BMC priority counterexample care minimization

`src/sat/bmc/bmcCexCare.c` is validated across all ten active definitions.
`MinimizeCounterexampleCarePriority` preserves the four active PI/PPI orderings,
forward register-priority transfer, the exact true-AND, single-controlling-zero
and double-zero priority recurrences, and backward selection. Reverse traversal
retains CO propagation, RI/RO frame transfer, PI preference, existing-mark
suppression, greater-priority choice and first-sparsest candidate ties. Concrete
initial register values are supported and every published candidate is ternary
replay verified.

`ExtendCounterexampleCareToObjects` returns owned per-frame object knownness
from ternary simulation; it supports nonzero initial states rather than the
source's zero-only scratch setup. `IntersectCounterexampleCare` owns its mask.
Concrete AIG adapters use the validated GIA conversion and return results rather
than printing. The source declares `Bmc_CexCareSatBasedMinimizeAig` but contains
no implementation anywhere in the surveyed tree. Its wrapper is replaced by a
verified native contract: deterministic implication-aware minimization at low
effort and the validated care/essential/minimized pipeline at high effort.

Go-only tests cover every priority recurrence, all four orderings, deterministic
double-zero PI preference, nonzero initial state, per-object knownness,
intersection ownership, both AIG replacement effort modes, invalid dimensions,
all resource limits and cancellation. A 10,000-gate case confirms iterative
operation. A 5,000-execution fuzzer exhaustively enumerates all PI completions
admitted by every result and independently replays the sequential truth table.
The complete suite passes in 15.620s, race in 81.920s, and vet is clean.
Inventory is now 251 validated, 39 partial and 1,467 unstarted files.


## BMC incremental initial-state equivalence

`src/sat/bmc/bmcBmci.c` is validated across all seven active definitions.
`UnfoldBmciFrame` preserves constant and symbolic register literals, shared
primary inputs between the two trajectories, ordinary-AND hashing and RI
transfer. `CompareInitialStateEvolution` unfolds both initial states through
the same input sequence, constructs each register XOR, distinguishes real SAT
differences from UNSAT equivalences, merges the same preferred trajectory as
the source and stops when no SAT difference remains. The zero-state wrapper
preserves `Gia_ManBmciTest`'s useful operation.

The CNF remap/lift and recursive partition helpers are consolidated into a
fresh checked encoding of the owned cumulative graph for each query. This
preserves every Boolean query and the frame/register merge order while trading
incremental clause reuse for simpler ownership and independent resource bounds.
Results expose the cumulative graph, final state literals, per-frame
miter/SAT/merge counts, aggregate statistics, fixed-point status and exact
unknown reason instead of stdout and an ambiguous success integer.

Go-only tests cover absorption-driven UNSAT merging, persistent SAT differences,
shared-input convergence, constants and symbolic states, all resource limits,
SAT-budget unknown results, cancellation and extended-gate rejection. A
10,000-gate test confirms iterative operation. A 5,000-execution fuzzer
independently enumerates both initial values and every shared input sequence and
matches each frame's reported difference and fixed point. The complete suite
passes in 15.806s, race in 82.334s, and vet is clean. Inventory is now 252
validated, 39 partial and 1,466 unstarted files.


## Incremental GIA bounded checking

`src/sat/bmc/bmcBmcG.c` is validated across all 13 active definitions.
`GiaManager.IncrementalBoundedCheck` consolidates the source manager lifecycle,
recursive unfolding and collection, frame/SAT maps, CNF loading, model decoding
and one-process driver over the validated owned unroller, AIG conversion,
Tseitin encoder, CDCL solver and witness replay. It preserves zero initial
state, frame/property query order, constraints, first-SAT termination, finite
bound versus unknown status, logical SAT budgets, cancellation and completion
callbacks.

Frame batching and eliminated-variable CNF remain performance choices rather
than duplicated solver subsystems. Print-only timing/frame functions are
replaced by `BmcResult`, `SatStats` and callback values. Extended GIA gates are
rejected because the source incremental unfolding treats every
`Gia_ObjIsAnd` node as an ordinary AND.

Go-only tests cover exact first witness and frame, independent replay, callback
order, finite safe bounds, constraints, property selection, ownership, all
resource limits and cancellation. A 10,000-gate case confirms iterative
operation. A 5,000-execution fuzzer exhaustively enumerates small sequential
truth tables and matches the earliest failing frame or finite safe result. The
complete suite passes in 15.948s, race in 82.272s, and vet is clean. Inventory
is now 253 validated, 39 partial and 1,465 unstarted files.


## M-inductiveness checking and register-set minimization

`src/sat/bmc/bmcICheck.c` is validated across all eight active definitions.
`CheckMInductiveSet` creates two arbitrary initial states under shared primary
inputs, requires every primary output and selected next-state function to agree
for the first M transitions, and checks for an output or selected-next-state
difference on transition M+1. This preserves the selector assumptions, exact M
meaning and satisfiable/noninductive distinction of `Bmc_DeriveSolver`.

`MinimizeMInductiveCore` preserves the source's repeated UNSAT-conflict
shrinking from either the all-enabled or all-disabled set.
`SearchMInductiveRegisters` carries a selected set through M=1..Frames and
greedily tests removals. Natural, reverse, backward-topological and reversed
backward-topological orders are supported. `BackwardRegisterOrder` preserves
fanin-first PO traversal through recursively discovered register-input cones,
then appends dangling registers in natural order.

Mutable lifted CNF and solver construction are consolidated into owned GIA
self-composition, GIA-to-AIG conversion, Tseitin clauses and concrete CDCL
assumptions. Structured status, stop reason, selected-set ownership, frame and
aggregate SAT statistics replace stdout. Object, mapping, variable, clause and
literal limits plus synchronous cancellation replace assertions and
wall-clock-only timeout. Buffers, XORs and muxes are rejected because the
source path would interpret them as ordinary ANDs.

Go-only tests cover exact M semantics, both core starts, every ordering,
ownership, source immutability, every resource class, SAT-budget unknowns,
cancellation, extended-gate rejection and a 10,000-gate iterative cone. A
5,000-execution fuzzer independently enumerates paired initial states and every
shared input sequence for generated output/transition truth tables. The full
suite passes in 15.958s, race in 83.596s, and vet is clean. Inventory is now
254 validated, 39 partial and 1,464 unstarted files.


## Interval bounded checking and ternary prefix analysis

`src/sat/bmc/bmcBmc2.c` is validated across all 25 active definitions. The
similarly named `Saig_BmcIntervalExplore_rec` text is block-commented and is not
an active definition. `SequentialTernaryPrefix` preserves zero register
initialization, unknown PIs, complemented-AND ternary recurrence, RI/RO state
transfer and the first PO not proved zero. Its owned frame/object matrix eagerly
evaluates live objects instead of retaining the source's zero-valued unvisited
slots.

`IntervalBoundedCheck` preserves frame-major then property-major target order,
start-frame skipping, every-PO target treatment, per-target and cumulative SAT
limits, UNSAT target blocking, first-SAT termination, failed target mapping and
zero-state witness replay. This intentionally includes POs marked as AIGER
constraints because `Saig_BmcPerform` does not give them special semantics.

The pointer manager, frame maps, recursive interval construction, stitched CNF
and lifetime helpers are consolidated over the owned GIA unroller, AIG
conversion, Tseitin encoder and concrete CDCL solver. Node batching remains a
performance hint because the finite Go relation is built in one bounded arena.
The alternate Satoko backend is not duplicated because it has no different
Boolean contract. Debug-only target materialization and printing are replaced
by the frame-major relation, structured results and a target callback.

Go-only tests cover ternary stopping/completion, exact target order, skipped
frames, constraint-marked outputs, callbacks, UNSAT blocking, replay,
ownership, source immutability, local/global SAT limits, every construction
limit, cancellation and a 10,000-gate iterative cone. Two independent
5,000-execution fuzzers compare against exhaustive concrete traces and a
possible-value-set structural ternary simulator. The full suite passes in
15.680s, race in 82.827s, and vet is clean. Inventory is now 255 validated, 39
partial and 1,463 unstarted files.


## Tent/rank analysis and all-output bounded checking

`src/sat/bmc/bmcBmcS.c` is validated across 22 unique active symbols and 23
conditional bodies; `Bmcs_ManPerformMulti` has mutually exclusive pthread and
non-pthread definitions. `AnalyzeTents` preserves PO-rooted traversal,
RI-to-RO expansion, first-tent assignment, cross-tent candidates, the source's
explicit constant-zero candidate clearing, deterministic boundary order and
rank-delay profiles without graph marks or borrowed scratch.

The source tent-network builder is consolidated into the owned analysis plus
`NewGiaRankedUnroller`. Its private source snapshot, circular rank slots,
atomic frame append and semantic snapshots replace the temporary retimed graph
and parallel map vectors. Existing exhaustive and fuzz validation proves ranked
snapshots equal ordinary unrolling. Tent analysis rejects buffers, XORs and
muxes because the source would traverse them as ordinary ANDs.

`SuperBoundedCheck` preserves zero initial state, frame/property target order,
every-PO treatment including constraint-marked outputs, first-SAT termination,
definitive-only callbacks, failed target mapping and replayed witnesses. Manager
lifetime, copies, recursive unfold/collect and incremental clean-CNF loading
are consolidated over the validated owned interval unroller, Tseitin encoder
and concrete CDCL solver. Frame batching and worker count remain validated
performance options; a deterministic solver replaces the nondeterministic
pthread portfolio without changing unlimited-budget Boolean results.

Go-only tests cover exact three-tent metadata, candidate and constant-zero
handling, ownership, source immutability, all-PO order, one/four-worker
equality, batching, callback order and unknown suppression, replay, every
resource limit, cancellation, extended-gate rejection and a 10,000-gate cone.
Separate 5,000-execution fuzzers compare tents with an independent DFS and BMC
with exhaustive sequential truth tables. The full suite passes in 15.856s,
race in 83.338s, and vet is clean. Inventory is now 256 validated, 39 partial
and 1,462 unstarted files.


## Ternary-guided sparse AND BMC

`src/sat/bmc/bmcBmcAnd.c` is validated across all 27 active definitions.
`AndTernaryUnroll` preserves zero-register and unknown-PI ternary simulation,
RI-to-RO transfer, first-unknown discovery, fixed and adaptive frame bounds,
reverse unknown-controlling-fanin selection, register-input leaf propagation,
the source's early backward cutoff, forward structural hashing, frame-major
outputs and original frame/PI provenance. Owned value, needed-node, root and
mapping slices plus iterative walks replace packed state words, graph scratch,
recursive helpers and the pointer manager.

Known-one POs are rejected because the source sparse builder silently turns
them into zero. AIGER constraints, buffers, XORs and muxes are likewise rejected
rather than being silently checked with the wrong source semantics.
`AndBoundedCheck` consolidates both cone-CNF implementations, SAT-variable
assignment, output scans and all three driver variants over concrete GIA/AIG
conversion, Tseitin CNF and the Go CDCL solver. It preserves zero state,
frame/property order, first-SAT termination, finite exclusion versus unknown,
resource/cancellation results and exact witness translation. Reconstructed
early-frame witnesses ignore sparse PIs that belong only to later frames and
are replayed against the original transition system.

Go-only tests cover exact ternary values, adaptive stopping, backward cutoff,
known-one pruning, provenance ownership, source immutability, witness replay,
safe exclusion, limits, cancellation and a 10,000-gate iterative cone. An
independent exhaustive truth-table fuzzer passed 5,000 executions. The complete
suite passes in 16.155s, race in 82.737s, and vet is clean. Inventory is now
257 validated, 39 partial and 1,461 unstarted files.


## Scalable solve-all BMC and structural analysis

`src/sat/bmc/bmcBmc3.c` is validated across all 46 active definitions.
`AnalyzeBMC3Ternary` preserves zero-state/unknown-input simulation and both
source stopping policies: five non-decreasing register-binary observations and
the first unknown PO. `UnknownConeCounts` follows unknown logic across latch
boundaries. `BMC3Sections`, `BMC3Supergate` and `BMC3ReferenceOrder` preserve
sequential postorder layers, complemented-output leaf expansion, reference
counts and deterministic reference/ID ordering without traversal scratch.

Four-input truth cofactoring and reduction preserve constant propagation and
functionally absent inputs. `BMC3Mapping` reuses the validated native-order DAR
mapping array, computes mapping-edge references and emits deterministic owned
BLIF text or a checked file. The old mapping manager, packed ternary words,
recursive visits and diagnostic stdout drivers are represented by owned slices,
iterative traversals and structured results.

`ScalableBoundedCheck` preserves start frames, frame/output order, solve-all
skipping, same-model multi-output discovery, entailed negative units after
UNSAT, conflict-limit jumps, finite-state completion, progress/failure
callbacks and zero-state witness replay. Concrete unrolling, Tseitin CNF and
the Go CDCL solver replace the mapping/literal pointer manager and three native
SAT backends. Logical budgets, explicit unknown dropping and cancellation
replace wall-clock deadlines; dropped outputs deliberately remain unknown
rather than becoming the source timeout path's unsafe success result. The
default-only `fDropSatOuts` and `nPisAbstract` fields are never read by this
source snapshot, so the port does not invent behavior for them.

Go-only tests cover histories, cone counts, sections, supergates, ordering,
truth reduction, mapping/BLIF, ownership, solve-all and same-model behavior,
query order, start/jump/completion, constraints-as-targets, callbacks, all
resource classes, cancellation and 10,000-node cones. Three independent
5,000-execution fuzz campaigns validate ternary possible sets, cofactor
reduction and exhaustive sequential reachability. The complete suite passes in
16.203s, race in 82.921s, and vet is clean. Inventory is now 258 validated, 39
partial and 1,460 unstarted files.

## IF cut-composition cache statistics

`src/map/if/ifCache.c` is validated across both active definitions.
`IfCutCache` preserves canonical DSD-pair ordering, repeated records,
first-occurrence tuple identity and per-shared-input entry and uniqueness
statistics. It uses a concrete `insdict.Dict` with an explicit stable record
hash and returns owned structured results instead of writing diagnostic stdout.

Go-only unit tests cover canonicalization, duplicates, bucket percentages,
empty input, ownership, bounds, atomic failed insertion, cancellation and nil
receivers. A 5,000-execution fuzzer compares every result with an independent
slice model. No generated native fixture or C oracle is retained. The complete
suite passes in 16.368s, the race suite in 84.776s, and vet is clean. Inventory
is now 266 validated, 40 partial and 1,451 unstarted files.

## IF cut decomposition checks

`src/map/if/ifCheck.c` is validated across all five active definitions on the
concrete owned AIG representation. `IfCutReach` performs the source's
cut-boundary reachability iteratively without borrowing object marks.
`IfCutCofactorConstant` fixes each boundary variable through owned truth tables.
`IfCutChecker` preserves cut-leaf order, the zero-before-one cofactor scan and
separate structural, reconvergent-conflict and functional counters while
returning typed results instead of using process-global state and stdout.

Go-only tests cover direct structural decomposition, reconvergence, majority
without a constant cofactor, exact statistics, malformed cuts, every resource
limit, atomic cancellation and a 10,000-node iterative path. A 5,000-execution
fuzzer checks functional results by exhaustive Boolean enumeration and
structural results with an independent graph walk. No generated native fixture
or C oracle is retained. The complete suite passes in 16.401s, the race suite
in 85.264s, and vet is clean. Inventory is now 267 validated, 40 partial and
1,450 unstarted files.

## GIA support-pair experiment

`src/aig/gia/giaSupMin.c` is validated across all four active definitions.
`AnalyzeSupportPairs` consolidates the source's allocation-only Start/Stop
lifecycle and printing drivers into one owned operation. The allocated truth
workspaces in this snapshot are never consumed. Typed records preserve pair
order, regularized complemented output drivers, AND-like care validation,
continuation after an invalid care, and distinct data/care support and cone
cardinalities.

Go-only tests cover exact reports, invalid-care continuation, source
immutability, malformed output indices, pair/object limits, atomic cancellation
and a 10,001-gate iterative cone. A 5,000-execution fuzzer compares every field
with an independent graph traversal. No generated native fixture or C oracle is
retained. The complete suite passes in 16.576s, the race suite in 84.957s, and
vet is clean. Inventory is now 268 validated, 40 partial and 1,449 unstarted
files.

## GIA LUT distance ordering and block packing

`src/aig/gia/giaPack.c` is validated across all three active definitions.
`LUTReachableOrder` preserves output-rooted breadth-first discovery followed by
reversal and intentionally omits dangling LUTs. `LUTDistanceOrder` propagates
CO distance through every mapped LUT, includes dangling roots and reuses the
validated exact `Abc_MergeSortCost` permutation, including its non-stable tie
behavior.

`GiaLUTMapping.Pack` preserves the 32-fanin bound, strict unique-latest-input
test, native selection-sort swaps, block capacity, route/direct delay
recurrence and reverse group fill. It returns owned packing, order, delay and
block arrays instead of mutating graph scratch or printing. The shared packing
format now admits caller-selected wider blocks; operations inherently limited
to three-member clusters reject wider groups explicitly.

Go-only tests cover reachable versus dangling order, exact delays/groups,
four-member blocks, malformed mappings, every resource bound, cancellation and
source ownership. A 5,000-execution fuzzer compares orders, delays and groups
with independent graph and packing models. No generated native fixture or C
oracle is retained. The complete suite passes in 16.481s, the race suite in
85.353s, and vet is clean. Inventory is now 269 validated, 40 partial and 1,448
unstarted files.

## GIA technology-independent mux and MFFC diagnostics

`src/aig/gia/giaTis.c` is validated across all five active definitions on its
ordinary-AND input domain. `DuplicateWithRecognizedMuxes` preserves object-order
structural hashing, recognized mux/XOR construction, the exception that leaves
a mux pattern ordinary when both root children are multiply referenced, graph
metadata and dangling cleanup. Pre-existing native muxes, XORs and buffers are
rejected because the source routine would interpret their compact fields as
ordinary AND logic.

`CollectTisMFFC` uses an iterative owned traversal in place of recursion and
manager traversal IDs. It preserves first-encounter leaf order, fanin
postorder, mux-control traversal and unconditional root inclusion. Constants
are safe leaves rather than an assertion path. `AnalyzeTis` replaces temporary
vectors and diagnostic stdout with an owned converted graph and typed reports
in source AND-then-CO order; terminal CO drivers that have no gate MFFC are
safely omitted.

Go-only tests cover exact mux/XOR truth, doubly shared pattern retention,
metadata, cleanup, report order, multiply referenced roots, malformed roots,
unsupported source shapes, every resource class, atomic cancellation and a
10,001-gate iterative cone. A 5,000-execution fuzzer compares exhaustive
Boolean behavior and every MFFC node/leaf order with independent Go models.
No generated native fixture or C oracle is retained. The complete suite passes
in 16.441s, the race suite in 85.469s, and vet is clean. Inventory is now 270
validated, 40 partial and 1,447 unstarted files.

## GIA MiniSat integer and float-permutation sorting

`src/aig/gia/giaSort.c` is validated across all 15 definitions.
`GiaSortInts` preserves the inlined MiniSat integer sort's 15-entry
selection-sort cutoff, middle pivot, strict comparison and Hoare split.
`GiaSortFloats` preserves the corresponding float32 swaps and synchronized
caller-provided or generated permutation, including deterministic source
behavior for equal values, signed zero, infinities and arbitrary NaN payloads.
Bounded owned work arrays make cancellation atomic without interfaces.

The file-local comparator, duplicate callback/inlined integer implementations,
sortedness assertion, global C-rand generator and 100-million-entry stdout
benchmark are represented by deterministic bounded Go tests rather than
retained runtime-only machinery. Tests pin exact small and partition orders,
custom permutations, malformed input, limits, cancellation, empty input and
100,000 integers. A 5,023-execution Go-only fuzzer compares integer results and
exact float bits/permutations with an independent literal algorithm model. No
generated native fixture or C oracle is retained. The complete suite passes in
16.344s, the race suite in 85.794s, and vet is clean. Inventory is now 271
validated, 40 partial and 1,446 unstarted files.

## GIA structural unateness

`src/aig/gia/giaUnate.c` is validated across all six definitions.
`AnalyzeStructuralUnateness` returns owned positive, negative, binate and
independent classes for ordered selected CI/CO ordinals. Its dense object-order
dataflow preserves the source's phase-literal union as the exact bitwise class
meet, complement involution, output phases, buffers and conservative
reconvergence. Nil selects all terminals; explicit empty and repeated ordered
selections have defined Go behavior.

`StructuralUnate` provides the one-input/one-output operation. Typed report
statistics replace both stdout test drivers. The result is intentionally
structural: opposite-polarity reconvergent paths can report binate even when
the Boolean function cancels that input. Native XOR and mux nodes reject
because the C routine would read them as ordinary two-input ANDs and omit mux
controls.

Go-only tests cover all four classes, selected and repeated terminals,
complemented outputs, reconvergent cancellation, buffers, statistics, malformed
selections, object/entry limits, atomic cancellation, source ownership and a
10,001-gate path. A 5,000-execution fuzzer checks every class with an
independent signed-path traversal and exhaustively proves every reported
positive, negative or independent Boolean implication. No generated native
fixture or C oracle is retained. The complete suite passes in 16.558s, the race
suite in 85.044s, and vet is clean. Inventory is now 272 validated, 40 partial
and 1,445 unstarted files.

## GIA experimental structural-isomorphism signatures

`src/aig/gia/giaIso3.c` is validated across all 12 definitions on ordinary
AND/buffer graphs. `Iso3InitialSignatures` preserves the six fixed object seeds
and fanin-complement classification. `Iso3Refine` preserves saved simultaneous
rounds, uint32 wraparound, complement/direction salts and accumulation.
`AnalyzeIso3` returns owned signatures and distinct-count history using the
source's repeated-adjacent-count stopping rule.

`AnalyzeIso4` reuses validated reverse levels and preserves the exact reset and
discard of the ABC random sequence, per-level complement salts, CO overwrite
and reverse fanin propagation from explicit zero-owned scratch. `CompareIso4`
replaces level data collection, sorting and two-graph stdout diagnostics with
typed multiset intersection counts. Native XOR/mux nodes reject because these
experimental source hashes omit their semantics and mux controls. The results
are structural fingerprints, not a complete isomorphism decision procedure.

Go-only tests cover exact seed/refinement values, convergence history,
isomorphic CI relabeling, fixed RNG propagation, level comparisons, malformed
signatures, every resource class, atomic cancellation, source ownership and a
10,001-gate graph. A 5,000-execution fuzzer compares initialization,
refinement, independently derived reverse levels and a separate ABC-RNG model.
No generated native fixture or C oracle is retained. The complete suite passes
in 16.427s, the race suite in 87.351s, and vet is clean. Inventory is now 273
validated, 40 partial and 1,444 unstarted files.

## GIA most-forward retiming

`src/aig/gia/giaRetime.c` is validated across all six definitions on ordinary
AND/buffer graphs. `RetimeForwardOne` and `RetimeForward` preserve shared
autonomous-state traversal, source object-order cuts, optional equal-class
crossing, phase-normalized new registers, forward duplication, cleanup and
iteration through the first zero-move round. Register classes are explicit
owned slices rather than hidden manager state, and typed round records replace
diagnostic stdout. Native XOR/mux nodes reject because the source treats their
packed representation as ordinary two-fanin AND logic.

The tests are entirely stand-alone Go. Exhaustive small-state checks prove the
reported cut encoding across every old state and primary-input assignment;
cycle-by-cycle tests compare multi-round sequential traces. Additional cases
pin exact cut order, class barriers, complemented phases, autonomous feedback
cycles, metadata, source immutability, malformed inputs, resource bounds,
atomic cancellation and a 10,001-gate iterative cone. A 5,000-execution fuzzer
generates bounded sequential graphs and rechecks the exhaustive state relation.
No native helper, generated C fixture or transient `testdata/` corpus is
retained. The complete suite passes in 16.349s, the race suite in 85.619s, and
vet is clean. Inventory is now 274 validated, 40 partial and 1,443 unstarted
files.

## GIA regular-expression automata

`src/aig/gia/giaRex.c` is validated across all 18 definitions.
`CompileGiaRegex` preserves the byte-oriented `()`, `|`, `*` language,
whitespace removal, doubled-parenthesis preprocessing, insertion or sorted
alphabet order, epsilon-transition construction, source state ordering,
balanced exactly-one inputs, acceptance output, register functions, cleanup
and conversion of the NFA's nonzero initial state to a zero-initialized GIA.
Print-only transition/reachability routines are represented by owned typed
metadata rather than stdout.

`Simulate` returns acceptance after each consumed byte. `GiaTranspose64` and
`AutomatonStep64` preserve the source's reversed-axis bit convention and
evaluate every one-hot input in parallel. `AutomatonInitialMask` retains the
source's complemented-RO structural inference. `WalkAutomaton` replaces global
C randomness and streaming text with caller-seeded deterministic typed walks,
visited states and counts. All operations have explicit dimensions, resource
bounds and atomic cancellation, with no interfaces.

Go-only tests compare accepted prefixes for representative and generated
expressions with the standard Go regexp engine, prove exactly-one logic by
exhaustive assignments, compare transpose with an independent bit model, and
compare all packed lanes with scalar GIA stepping from arbitrary states. They
also cover alphabet orders, whitespace, malformed expressions, metadata,
walk replay/determinism, resource limits, cancellation and a 1,000-symbol
iterative expression. Two independent fuzz campaigns passed 5,000 executions
each for language and packed-step equivalence. No native helper, generated C
fixture or `testdata/` corpus is retained. The complete suite passes in
16.089s, the race suite in 86.053s, and vet is clean. Inventory is now 275
validated, 40 partial and 1,442 unstarted files.

## GIA solver3 integration

`src/aig/gia/giaSat3.c` is validated across all five wrapper definitions.
`GiaSat3Session` concretely owns CNF clauses, the already ported solver3 and
GIA-CI model projection. `SolveGiaSat3` preserves the source's aggregate mode,
which asserts one disjunction over every CO, and split mode, which constructs
and solves each CO's DFS cone in source order while retaining all source CIs.
Typed statuses, stop reasons, models, conflict/statistic counts and problem
dimensions replace timing/status stdout.

This checkpoint deliberately uses the validated standalone reference CNF
encoder. It preserves the asserted Boolean relation but does not claim the
separate eight-input Mf mapping implementation used by the C wrapper; those Mf
internals remain tracked in their own source entries. Sessions support selected
CO disjunctions, simplify before solving, own their lifetime, enforce object,
mapping, variable, clause and literal bounds, and combine construction and SAT
cancellation without interfaces.

Go-only tests exhaustively determine each output and aggregate satisfiability,
replay every returned model through GIA evaluation, exercise ordinary ANDs and
lowered native XOR/mux logic, check source ownership, selected-output sessions,
lifecycle, malformed requests, every resource surface, cancellation at both
construction and solve time, and a 10,001-gate iterative cone. A 5,000-case
fuzzer compares split and aggregate statuses and models against exhaustive
Boolean enumeration. No native helper, C oracle or `testdata/` corpus is
retained. The complete suite passes in 16.598s, the race suite in 86.302s, and
vet is clean. Inventory is now 276 validated, 40 partial and 1,441 unstarted
files.

## Compact AGI support experiment

`src/aig/gia/giaAgi.c` is validated across all 37 definitions on the
experiment's exercised ordinary-AND domain. Concrete `AgiID`, `AgiLit`,
`AgiObject` and `AgiManager` types replace packed untyped words while retaining
object identity, literal complementation, CI/CO order, PI/RO and PO/RI roles,
register count, bounded append operations, metadata and explicit ownership.
`AgiFromGia` converts without borrowing GIA Value or traversal scratch, and
`AgiManager.Evaluate` independently checks the resulting transition relation.

Support traversal is iterative and counts unique reachable CIs for constants,
CIs, ordinary ANDs and CO wrappers. `AnalyzeSupports` replaces timing/stdout
with ordered node IDs, sizes and threshold count. `CompareGiaAgiSupports`
replaces the source test driver with an owned per-node comparison against the
existing independent GIA support-set implementation. The source declares
third-fanin/XOR/mux/majority accessors but never allocates their backing arrays;
the Go conversion rejects native buffers/XORs/muxes instead of making that
dormant unsafe path appear supported.

Go-only tests pin identity mapping, terminal roles, metadata, all assignments,
exact support sets and threshold counts, construction/lifecycle errors, every
resource class, atomic cancellation and a 10,001-gate iterative cone. A
5,000-case fuzzer compares output truth and every node support size with the
independent GIA implementation. No native helper, C oracle or `testdata/`
corpus is retained. The complete suite passes in 16.551s, the race suite in
85.899s, and vet is clean. Inventory is now 277 validated, 40 partial and
1,440 unstarted files.

## GIA Satoko-wrapper integration

`src/aig/gia/giaSatoko.c` is validated across all nine wrapper definitions.
`CollectCNFMappingVariables` replaces the recursive packed-mapping walk with
bounded iterative DFS while preserving unique variable postorder. It accepts
leaf roots safely and reports malformed offsets, invalid variables, cycles,
visit limits and cancellation instead of relying on assertions or recursion.

`GiaSatokoSession` concretely owns the standalone CNF relation, Go CDCL solver,
GIA-CI model projection and lifecycle. `SolveGiaSatoko` preserves the source's
aggregate output disjunction, independent source-order output cones, and shared
incremental relation with one output assumption per solve. Incremental reports
return deterministic cone-variable postorders. Typed status, stop reason,
model, conflict/statistic counts and problem dimensions replace timing and
stdout. `ParseGiaSatokoDIMACS` provides bounded in-memory parsing with comments,
optional checked headers, signed literals, empty clauses, multiline clauses and
strict malformed-input reporting.

This checkpoint uses the validated reference Tseitin encoder and concrete Go
CDCL solver. It proves the wrapper's Boolean, ownership and lifecycle behavior,
but does not claim Satoko's internal branching, clause layout or cone-mark
scheduling heuristic, nor the separately tracked eight-input Mf mapper. Cone
marking is represented by owned cone metadata rather than a false solver-
heuristic parity claim.

Go-only tests compare all three modes with exhaustive truth enumeration and
replay every SAT model through GIA evaluation. They cover selected output
sessions, constants, ordinary and lowered extended gates, exact mapping
postorder, sharing, cycles, malformed records, a 10,001-variable iterative
cone, DIMACS SAT/UNSAT/empty-clause cases, every resource surface, construction
and solve cancellation, ownership and idempotent close. The randomized graph
campaign passed 5,000 executions without a native oracle or retained
`testdata/`. The complete suite passes in 16.241s, the race suite in 85.611s,
vet is clean, and package-header and standalone-oracle audits pass. Inventory
is now 278 validated, 40 partial and 1,439 unstarted files.

## GIA frontier representation

`src/aig/gia/giaFront.c` is validated across all five definitions on its safe
two-distinct-fanin domain. `FrontierCrossCut` owns source-order reference counts
and reports the exact peak number of live values without borrowing manager
scratch. `EncodeGiaFrontier` uses the source's cyclic first-free allocation and
float32 ten-percent width margin, but stores slot fanins and object roles in a
typed owned `GiaFrontierGraph`. `Decode` performs the inverse transformation,
and `FrontierRoundTrip` replaces the print-only source test with an exact result.

Object IDs and compact words, complement phases, constant/CI/CO roles, ordinary
AND and native XOR nodes, dangling logic, sequential partitions, constraints,
Name and Spec all survive the round trip. Native buffers and explicit third-edge
mux controls are rejected: the C routine decrements exactly two fanins for every
AND-like object and has no safe representation for those Go-manager extensions.
This boundary is explicit rather than silently undercounting or corrupting a
frontier.

Go-only tests pin the exact cyclic slot sequence and width for a hand graph,
compare every peak against an independent future-use liveness model, exhaust
truth assignments, verify exact object and metadata recovery, and cover slot
snapshot ownership, dangling nodes, malformed records, bounds, cancellation,
unsupported objects and a 10,001-gate iterative round trip. A 5,000-case fuzzer
checks independent peaks, exact compact objects and exhaustive truth on random
AND/XOR graphs. No native helper or `testdata/` corpus is retained. The complete
suite passes in 16.495s, the race suite in 85.866s, vet is clean, and package-
header and standalone-oracle audits pass. Inventory is now 279 validated, 40
partial and 1,438 unstarted files.

## GIA custom AIGER extension codecs

`src/aig/gia/giaAigerExt.c` is validated across all 14 definitions.
`GiaEquivalenceClasses` owns representative/member records and proved bits;
its codec preserves the constant-class prefix, representative deltas, member
deltas and exact ABC unsigned-varint layout. The compact LUT-mapping codec
preserves ascending root order, ordered fanins and signed delta direction while
reconstructing the source's root footer. Simple mapping preserves the complete
packed array, documented mapping preserves big-endian record counts and maximum
LUT size, and packing vectors preserve signed big-endian int32 values.

Every decoder validates byte lengths, object and entry bounds, truncated and
overflowing varints, delta underflow/overflow, duplicate or non-topological
records, documented maxima and trailing data. All results are owned, encoding
is deterministic, and cancellation is checked during record traversal. The C
simple codec copied host-memory integers; the Go format deliberately specifies
little-endian int32, matching deployed x86 ABC artifacts while removing host-
endian ambiguity.

Go-only tests pin exact equivalence and compact-mapping bytes, independently
decode fixed integers, round-trip all mapping variants and signed packing, and
cover empty sets, ownership, malformed input, every limit and cancellation.
Two independent 5,000-case fuzz campaigns exercise structured equivalence
classes and all mapping codecs, including mutated byte streams. No C helper,
external executable or `testdata/` corpus is retained. The complete suite
passes in 16.134s, the race suite in 85.658s, vet is clean, and package-header
and standalone-oracle audits pass. Inventory is now 280 validated, 40 partial
and 1,437 unstarted files.

## GIA hierarchical LUT packing

`src/aig/gia/giaIff.c` is validated across all ten definitions.
`GiaIffSession` owns cloned graph, mapping, LUT-delay and timing state, with
explicit lifecycle and no interfaces. `Boundary` preserves the source-order
unique cut after absorbing zero, one or two direct LUT fanins. One-, two- and
three-LUT arrival candidates use float32 maxima and input-count delays, examine
children in mapping order, update only on strict improvement, and retain an
already selected pair when no triple improves it.

`Run` propagates CI arrivals and CO arrivals through the concrete timing
manager, returns every object arrival and selected child, and replaces stdout
statistics with maximum delay plus single/pair/triple counts. Packing selection
preserves output order, source dependency order, postorder group emission and
the source rule that absorbed children belong to their parent's group. An
iterative work stack removes recursive depth risk. `BuildGiaIffPacking` also
implements the source command's automatic 4/6-LUT and degree-two/three library
selection without attaching mutable mapping or packing fields to `GiaManager`.

Go-only tests pin exact merged boundaries, pair/triple selection and group
order, timing propagation, strict ties, automatic configuration, ownership,
idempotent runs, lifecycle, malformed dependencies, all limits, cancellation
and a 10,001-LUT iterative chain. A 5,000-case fuzzer compares every arrival
and match with an independent set-based model and validates each packing. No
native helper or `testdata/` corpus is retained. The complete suite passes in
16.399s, the race suite in 85.744s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 281 validated, 40 partial and
1,436 unstarted files.

## GIA packed-simulation and dataset checkpoint

`src/aig/gia/giaGen.c` is validated across all 13 definitions. Caller-supplied
CI-major words and the source-compatible deterministic random stream propagate
through ordinary AND graphs into owned object-major packed simulations. Binary
training/test artifacts and PLA training/validation/test artifacts are returned
as named byte slices instead of being written to the working directory. The
binary format now has an explicit little-endian int32 header, matching deployed
x86 ABC output without retaining host-endian behavior. PLA input is parsed from
bounded caller-owned bytes; comparison and JSON statistics are returned as
typed values rather than printed.

Native buffers, XORs and muxes are rejected because this source treats every
AND-like object as an ordinary two-input AND. Tests independently evaluate all
packed samples with the Go Boolean graph evaluator, decode every generated
binary row, parse and check every generated PLA row, check deterministic stream
advancement and ownership, and cover malformed data, limits and cancellation.
Packed-simulation and PLA-parser fuzz campaigns each passed 5,000 executions;
the latter uses canonical Go-only dataset round trips.

No C capture source, native executable or transient fuzz corpus is retained.
There are no C/C++ sources or headers anywhere under `pkg`, `pkg/tools/`
contains only `source_manifest.py`, and the package policy test enforces this
stand-alone boundary on every run. The complete suite passes in 16.712s, the
race suite in 86.084s, vet is clean, and package-header and standalone-oracle
audits pass. Inventory is now 282 validated, 40 partial and 1,435 unstarted
files.

## Espresso minimum-column-cover checkpoint

`src/misc/espresso/sminterf.c` is validated across its sole operation.
`MvcCover.MinimumColumnCover` converts each owned packed cube into one sparse
matrix row and returns an independently owned bit mask intersecting every row.
The native wrapper requested `sm_minimum_cover` in non-branching heuristic mode;
Go deliberately strengthens the public contract to a deterministic exact
minimum under explicit budgets. Row dominance, a deterministic greedy upper
bound, a disjoint-row lower bound and coverage-ordered branching keep the
search bounded without interfaces or global Espresso state.

Empty matrices return the empty cover; impossible empty rows are diagnosed.
Rows, columns, quadratic dominance checks and search nodes have separate
limits, cancellation returns no partial result, and neither the input cover nor
separate outputs alias. Unit tests pin dominance, duplicate rows, deterministic
ties, zero-width and 128-column cases, ownership and every error surface. Every
row-family through three columns is exhaustively compared with brute-force
enumeration, and a separate 5,000-case fuzz campaign compares random matrices
through eight columns with the same independent model.

No C helper, external executable or retained fuzz corpus is used. The complete
suite passes in 16.743s, the race suite in 86.108s, vet is clean, and package-
header and standalone-oracle audits pass. Inventory is now 283 validated, 40
partial and 1,434 unstarted files.

## Espresso sparse-dominance checkpoint

`src/misc/espresso/dominate.c` is validated across both operations.
`ReduceDominatedRows` preserves sparse row identities and source-order tie
rules while removing strict supersets and equal higher-numbered rows.
`ReduceDominatedColumns` preserves dynamic row/column populations, unit or
caller-supplied weights, superset coverage and the source's equal-length
higher-column preference. Both return an owned reduced MVC cover, an owned
active-column mask and typed removal counts instead of mutating pointer-linked
sparse storage.

Packed-word masks accelerate repeated row population and subset checks while
column membership retains exact dynamic source behavior. Empty rows, malformed
weights, ownership, row/column/work limits and atomic cancellation are checked.
Unit tests pin row and column deletion order, duplicate ties, weighted
non-dominance, active-column pruning and independent results. Every matrix
family through three columns is exhaustively checked to preserve minimum cover
cost, and a 5,000-case fuzz campaign proves both unweighted row reduction and
weighted column reduction against brute-force enumeration through eight
columns.

No native helper or retained fuzz corpus is used. The complete suite passes in
16.525s, the race suite in 86.283s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 284 validated, 40 partial and
1,433 unstarted files.

## Espresso sparse block-partition checkpoint

`src/misc/espresso/part.c` is validated across its public operation and three
static traversal/copy helpers. `MvcCover.BlockPartition` traverses the concrete
row/column bipartite graph iteratively, preserving the first-row component,
source row order and original present-column identities. It also preserves the
source's unusual early connected result when that first component reaches every
present column, including matrices with disconnected empty rows.

Disconnected matrices return independently owned left/right covers plus owned
row and column identity lists; connected and empty matrices allocate no result
covers. Dimensions, sparse entries, visits and cancellation are bounded, and
partial results are closed atomically. Tests pin connected, empty, multi-block,
empty-row and all-columns-reached behavior, ownership, all limits and a
10,001-row iterative component. A 5,000-case fuzz campaign compares every
partition and identity list with an independent fixed-point bipartite graph
model.

No native helper or retained corpus is used. The complete suite passes in
16.612s, the race suite in 86.350s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 285 validated, 40 partial and
1,432 unstarted files.

## Espresso minimum-cover solution checkpoint

`src/misc/espresso/solution.c` is validated across all seven operations.
`MvcCoverSolution` owns checked selected-column state and cost, with explicit
construction, cloning and close. Unit or caller-supplied nonnegative weights
are supported with overflow checks. `Accept` returns an owned matrix with all
covered rows removed and commits the selected column atomically; `Reject`
returns an owned matrix with the column removed and rows newly made empty
discarded. `TakeBetterCoverSolution` preserves lower-cost selection, first-
argument ties, nil alternatives and native-style loser ownership release.

Duplicate selections and invalid sparse-pointer preconditions become errors.
Rows, nonzero output entries and cancellation are bounded, and failed accepts
leave both the solution and input matrix unchanged. Unit tests cover weighted
costs, cloning, lifecycle, tie ownership, covered-row deletion, rejected-column
cleanup, absent columns, overflow, limits and cancellation. A 5,000-case fuzzer
compares accept/reject matrices and selected cost with an independent bit-mask
state model. Its one minimized transient case clarified that rejecting an
absent column is a complete no-op, including for the Go representation's safe
empty-row extension; no fuzz corpus was retained.

The complete suite passes in 16.970s, the race suite in 86.585s, vet is clean,
and package-header and standalone-oracle audits pass. Inventory is now 286
validated, 40 partial and 1,431 unstarted files.

## Espresso Gimpel-reduction checkpoint

`src/misc/espresso/gimpel.c` is validated across its sole operation.
`MvcCover.GimpelReduce` preserves the first eligible two-column pattern,
secondary residual capture, residual merging into every second-column row,
primary/secondary and two-column deletion, reduced solving and the source's
residual-intersection choice between the two reconstructed columns. It clones
optional base solution state and preserves the strict “better than bound”
result rule.

The reduced problem is solved by the validated exact Go minimum-cover engine.
Two fuzz-minimized matrices exposed preconditions supplied implicitly by the
native caller's earlier dominance and essential-column passes: the secondary
row must not also contain the second column, and its residual must be nonempty.
The standalone Go API now checks these conditions and treats unsafe raw
patterns as non-applicable instead of returning an invalid cover. Unit tests
pin both reconstruction branches, base ownership, bounds, non-applicable
patterns, all limits and cancellation. A 5,000-case fuzzer proves every applied
solution covers the original matrix and has the same cardinality as exhaustive
subset enumeration.

No transient minimized case or native corpus is retained. The complete suite
passes in 16.840s, the race suite in 86.226s, vet is clean, and package-header
and standalone-oracle audits pass. Inventory is now 287 validated, 40 partial
and 1,430 unstarted files.

## Espresso maximal-independent-row checkpoint

`src/misc/espresso/indep.c` is validated across both active definitions.
`MvcCover.MaximalIndependentRows` builds an owned row-intersection graph,
repeatedly selects the active row with the smallest active closed-neighborhood
degree, preserves the source's first-row tie rule and removes the selected
closed neighborhood. It returns source-compatible sorted row IDs plus the
otherwise implicit greedy selection order. Cost is the selected-row count for
unit weights or the sum of each selected original row's minimum column weight.

Empty sparse rows become explicit invariant errors rather than unchecked first-
element dereferences. Dimensions, quadratic pair checks, adjacency storage,
visits, integer overflow and cancellation are bounded. Unit tests pin exact
selection order, weighted costs, ownership, atomic failures, all limits and a
129-column case. Every row family through three columns is compared with an
independent graph model and brute-force minimum-cover cost, proving independence,
maximality and the lower-bound property. A separate Go-only fuzz campaign passed
5,000 executions through eight columns against the same independent model.

No native helper or retained fuzz corpus is used. The complete suite passes in
16.675s, the race suite in 85.435s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 288 validated, 40 partial and
1,429 unstarted files.

## Espresso weighted minimum-cover checkpoint

`src/misc/espresso/mincov.c`, `mincov.h` and `mincov_int.h` are validated.
`MvcCover.MinimumWeightedColumnCover` accepts unit or caller-supplied
nonnegative costs and returns owned sorted column IDs, total cost and concrete
search statistics. Exact mode uses deterministic bounded branch-and-bound with
row dominance, a cost-normalized greedy incumbent and a weighted disjoint-row
lower bound. Heuristic mode preserves the source's no-branch valid-cover
contract through that deterministic greedy incumbent. Diagnostic clocks,
global debug state, sparse-pointer ownership and aborts are absent.

Zero-cost columns, maximum integer costs, cost overflow, empty matrices,
impossible empty rows, dimensions, preprocessing work, search nodes and
cancellation are explicit. Unit tests pin a matrix where heuristic cost three
is improved to the exact cost two, weighted selection, result ownership,
atomic errors and every limit. Every matrix through three columns is tested
under every weight vector over `{0,1,2}` against brute-force enumeration. A
separate Go-only campaign fuzzed 5,000 matrices through eight columns, checking
exact cost and independently replaying both exact and heuristic covers.

No native helper or retained fuzz corpus is used. The complete suite passes in
16.744s, the race suite in 86.423s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 291 validated, 40 partial and
1,426 unstarted files.

## Espresso sparse-matrix foundation checkpoint

`src/misc/espresso/matrix.c`, `rows.c`, `cols.c`, `sparse.h` and
`sparse_int.h` are validated. Concrete `EspressoSparseRow`,
`EspressoSparseColumn` and `EspressoSparseMatrix` types preserve sorted unique
indices, ascending row/column traversal, duplicate insertion, bidirectional
membership, row/column deletion, transpose-copy behavior, longest ties,
lexical comparison, containment, intersection and modular hashes. Owned slices
and integer metadata replace intrusive links, allocator free lists and untyped
user pointers without interfaces.

The matrix supports bounded pair-text and compressed parsing, canonical text
writing, source-style display/dump output, cloning and explicit close. It
rejects malformed high bits in a final compressed word rather than inserting
columns outside the declared dimension. Unit tests cover every vector and
matrix operation, exact formatting, metadata, ownership, limits and errors.
One 5,000-case Go-only fuzzer applies arbitrary insert, element removal, row
deletion and column deletion while validating both indexes after every step;
a second 5,000-case campaign fuzzes compressed parsing and canonical Go-only
round trips.

No native helper or retained fuzz corpus is used. The complete suite passes in
16.917s, the race suite in 86.352s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 296 validated, 40 partial and
1,421 unstarted files.

## Espresso packed-set foundation checkpoint

`src/misc/espresso/set.c` is validated across all 50 definitions. Concrete
`EspressoSet` and `EspressoSetFamily` types preserve 32-bit bit numbering,
rank/flag metadata, Boolean operations and predicates, active-set compaction,
replacement-with-last deletion order, joins, consuming append, column counts,
restricted weighting, column insertion/deletion/copy/compression, transpose
and arbitrary permutation. Owned slices replace packed header aliasing,
allocator garbage lists and static formatting buffers without interfaces.

The source native hexadecimal format and bit-matrix format have bounded
byte-slice parsers and canonical writers. Rebuilding column transforms clear
rank/flags exactly as the source does; the same-word end-append fast path alone
preserves them. Malformed packed loops, tail bits, dimensions, metadata,
truncation and resource limits return errors. Unit tests exhaust every pair of
sets through eight bits and pin wide words, zero width, family lifecycle,
metadata, exact formats, all transforms and failures. Separate Go-only packed-
operation and family-round-trip/permutation fuzz campaigns each passed 5,000
executions.

No native helper or retained fuzz corpus is used. The refreshed complete suite
passes in 17.351s, the refreshed race suite in 88.660s, vet is clean, and
package-header and standalone-oracle audits pass. Inventory is now 297
validated, 40 partial and 1,420 unstarted files.

## Espresso variable-aware cube-set checkpoint

`src/misc/espresso/setc.c` is validated across all 13 definitions. Immutable
`EspressoCubeLayout` records binary and multiple-valued part ranges explicitly,
removing the native global cube masks. Its operations preserve full-row checks,
zero/exact/capped distance, force-lower accumulation, per-variable consensus,
sole-active-variable detection and shared-active-variable tests over concrete
owned `EspressoSet` values.

Rank-first descending/ascending comparators, reverse-word lexical comparison,
masked distance-one comparison and nullable descending order preserve native
sort behavior. The layout checks binary arity, total parts and set dimensions;
results never alias inputs. Unit tests pin mixed binary/MV behavior, wide-word
sort precedence, masks, nils, ownership and errors. All 16,384 pairs of
arbitrary seven-part cubes are compared with an independent per-variable
distance/consensus model, and a wider mixed-layout Go-only fuzzer passed 5,000
executions while also checking force-lower.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.106s, the race suite in 92.777s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 298 validated, 40 partial and
1,419 unstarted files.

## Espresso containment checkpoint

`src/misc/espresso/contain.c` is validated across all 18 definitions. Concrete
`EspressoSetFamily` operations preserve cardinality-derived Rank values,
descending containment order by reverse packed words, ascending reverse-
containment order, stable first-survivor duplicate removal, indexed source-row
remapping, contained union and masked distance merging. Distance-one merging
uses the variable-aware cube layout masks. Owned results replace the native
pointer-array list/unlist machinery and consuming input convention without
interfaces.

The implementation uses bounded quadratic subset checks in place of the
source's pointer-optimized internal removal routines. Explicit set/check limits
and cancellation make worst-case work controllable. Unit tests pin ordering,
metadata, duplicates, strict containment, index remapping, cross-family union,
masked merging, ownership and all failures. Independent Go-only containment
and distance-merge fuzz models each passed 5,000 executions.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.363s, the race suite in 87.773s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 299 validated, 40 partial and
1,418 unstarted files.

## Espresso minterm-map checkpoint

`src/misc/espresso/map.c` is validated across all three definitions.
`EspressoCubeLayout.Minterms` expands concrete set families using the source's
mixed-radix numbering with variable zero as the least-significant coordinate.
Owned local recursion replaces the native `Gcube` and `Gminterm` process-global
scratch. `FormatKarnaughMap` returns deterministic owned text instead of
writing stdout while preserving output-major offsets, exact spacing, all 256
coordinate-table entries and the source's linear page count above eight binary
inputs.

Dimensions and product overflow are checked before allocation. Explicit cube,
minterm and traversal limits plus cancellation bound expansion atomically.
Unit tests pin mixed-radix order, ownership, every coordinate, exact formatting
and failures. An independent assignment-enumeration Go model fuzzed arbitrary
binary covers for 5,000 executions.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.159s, the race suite in 88.771s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 300 validated, 40 partial and
1,417 unstarted files.

## Espresso cover-cost checkpoint

`src/misc/espresso/cvrmisc.c` is validated across all eight definitions.
Concrete `EspressoCoverCost` values count missing binary/dense parts, present
sparse/output parts and prime flags exactly as the source's empty-cofactor
`massive_count` path. Ordinary value cloning replaces `copy_cost`. Exact cost,
size and trace strings replace static buffers and stdout. Caller-provided
elapsed milliseconds preserve native two-decimal truncation without hidden
process clocks, and `EspressoCoverTotals` atomically records per-operation time
and call counts.

Checked errors replace fatal process termination. Layout/family dimensions,
cube/check limits, arithmetic overflow and cancellation are explicit. Unit
tests pin sparse and dense accounting, prime counts, both cost formats, exact
size/trace text, totals, ownership, atomic failure and invalid state. An
independent per-part Go model fuzzed arbitrary mixed layouts for 5,000
executions.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.236s, the race suite in 88.683s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 301 validated, 40 partial and
1,416 unstarted files.

## Espresso cofactor checkpoint

`src/misc/espresso/cofactor.c` is validated across all nine definitions.
Concrete owned `EspressoCubeList` values replace pointer/sentinel lists for one,
two or three families. General and single-variable cofactoring preserve
cumulative complement masks, distance-zero filtering and explicit source-cube
exclusion. The specialized path also safely filters malformed empty products
outside its selected variable instead of relying on an unchecked native
invariant. Unlisting returns owned `cube|cofactor` families.

`EspressoCubeCounts` preserves per-part and per-variable zeros, active and
unate counts, binary balance, the executable MV balance value and exact best-
variable ties. Binate splits retain source part order and floor-half division.
Stable masked sorting/deduplication corrects the dormant C simplifier's off-by-
two loop, which otherwise discards entries. Explicit cube/check limits and
cancellation cover traversal and merge comparisons. Fixed tests pin ownership,
metadata, ordering, exclusion, counts, split masks, limits and atomic errors.
An independent Go model fuzzed cofactoring, specialized equivalence, all count
fields, unlisting and simplification for 5,000 executions. Its minimized empty-
product discovery is retained as a Go seed; ephemeral `testdata/` was removed.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.317s, the race suite in 88.005s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 302 validated, 40 partial and
1,415 unstarted files.

## Espresso cube-structure checkpoint

`src/misc/espresso/cubestr.c` is validated across all four definitions.
Immutable owned `EspressoCubeLayout` construction replaces global setup,
including binary-size normalization and absolute sizing plus explicit markers
for negative symbolic MV fields. Concrete accessors expose variable counts,
output identity, symbolic flags, part and packed-word ranges, per-variable,
binary, MV, full and empty masks, default sparsity and the binary input-word
mask. Go word indexes are zero based because metadata is stored in fields
rather than a packed header word.

Go ownership replaces allocation teardown. Independently usable immutable
layouts and deep clones replace the native single-slot save/restore globals,
removing nesting and concurrency hazards; count-analysis values allocate the
former `cdata` arrays on demand. Unit tests pin symbolic normalization, wide
word boundaries, every derived mask/range, sparse defaults, output rules,
clone ownership and invalid inputs. A Go-only structure fuzzer passed 5,000
executions across arbitrary layouts while proving contiguous disjoint variable
masks and exact full-set partitioning.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.404s, the race suite in 88.833s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 303 validated, 40 partial and
1,414 unstarted files.

## Espresso binary cube-sizer checkpoint

`src/misc/espresso/cubehack.c` is validated across all four definitions.
Concrete mutex-protected `EspressoCubeSizer` values replace the process-global
binary layout and hidden static `called_before` flag. Exact immutable all-
binary layouts are reused by logical size and returned as deep clones;
`Undefine` releases the owned cache and `Current` returns an independent
snapshot. The native at-most-100-input oversized allocation is intentionally
omitted because it is solely an unobservable memory-reuse strategy. Counts
above 100 preserve their exact logical layouts.

`DefaultEspressoMinimizeFlags` preserves all ten source flag assignments.
Negative and oversized counts return checked errors. Unit tests pin zero,
ordinary, 100-boundary and larger sizing, repeated reuse, snapshot ownership,
undefinition, every default and failures. Concurrent definition/undefinition is
race-safe. A Go-only operation-sequence fuzzer passed 5,000 executions while
checking exact dimensions and binary/full-mask equality.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.219s, the race suite in 89.933s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 304 validated, 40 partial and
1,413 unstarted files.

## Espresso sharp algebra checkpoint

`src/misc/espresso/sharp.c` is validated across all ten definitions. Concrete
cube-layout methods implement ordinary and disjoint cube sharp, balanced cube-
by-cover subtraction, both cover-sharp orientations, cover intersection and
ordered disjointization. Per-variable difference construction, metadata on
unchanged disjoint cubes, owned results, canonical containment and pairwise-
disjoint guarantees are preserved. Empty malformed products are normalized by
their set semantics rather than retained as meaningless rows.

The source's fixed 500-cube containment batches are replaced by one bounded
canonical pass with identical cover semantics. Hidden debug timing and stdout
are omitted in favor of deterministic returned values. One shared work state
charges every inspected part, pair and generated intermediate cube, with
cancellation and atomic limit failures throughout nested recursion. Unit tests
pin cube order, metadata, ownership, all operation orientations, empty cases,
limits and errors. A Go-only exhaustive-assignment model fuzzed ordinary and
disjoint difference, intersection and disjointization for 5,000 executions,
including the source's pairwise-disjoint guarantees.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.084s, the race suite in 88.787s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 305 validated, 40 partial and
1,412 unstarted files.

## Espresso prime-consensus checkpoint

`src/misc/espresso/primes.c` is validated across all four definitions.
`EspressoCubeLayout.PrimeConsensus` preserves empty, singleton and full-row
cases, common-zero-column factoring, exact active/unate analysis, source best-
variable splitting, single-variable branch cofactors, filtering by split cubes,
distance-one cross-branch consensus and canonical containment. Inputs remain
owned and unchanged rather than being consumed.

Fixed 500-cube merge batches are replaced by one canonical containment pass of
identical cover meaning. A shared recursion-depth, operation and generated-cube
budget covers special-case scans, count analysis, branch cofactors, wide MV
intersections, consensus and containment; cancellation returns no partial
result. Unit tests pin empty, singleton, tautology, factoring, unate, mixed-MV,
recursive parity, ownership and every failure class. An independent Go model
enumerates every valid cube, selects all implicants and retains only maximal
ones. It matched recursive output for 5,000 arbitrary covers through four
variables.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.319s, the race suite in 89.953s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 306 validated, 40 partial and
1,411 unstarted files.

## Espresso essential-prime checkpoint

`src/misc/espresso/essen.c` is validated across all four definitions.
`ConsensusSharp` preserves distance-zero containment suppression, MV per-
variable merging, binary intersection fallback, distance-one consensus and
explicit source-row exclusion. `EssentialCube` implements the source sharp-
consensus plus don't-care coverage criterion. `ExtractEssentials` preserves
`NONESSEN`/`RELESSEN` gating, stable source order and `ACTIVE` metadata while
returning owned essential, remaining-ON and augmented-DC covers instead of
mutating caller pointers.

Tests enforce the documented invariant that ON rows are primes of ON∪DC and
intersect ON. A shared work state absorbs nested sharp traversal and bounds
generated cubes, packed consensus work and cancellation. Unit tests pin all
distance cases, mixed binary/MV construction, flags, ordering, ownership,
limits and errors. An independent assignment model constructs exhaustive prime
covers of arbitrary ON∪DC functions, filters ON-intersecting primes, and checks
unique ON minterms. It matched extraction for 5,000 executions.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.232s, the race suite in 88.335s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 307 validated, 40 partial and
1,410 unstarted files.

## Espresso verification and PLA-core checkpoint

`src/misc/espresso/verify.c` is validated across all four definitions. A
concrete owned `EspressoPLA` core now groups immutable layout, ON/DC/OFF covers,
labels and basic metadata. `VerifyCovers` preserves both source containment
directions modulo the old don't-care set and returns owned counterexample
covers instead of printing. Name-based permutation strictly and
deterministically associates every reference label across all three covers,
rejects missing or ambiguous labels, and returns a new PLA on the reference
layout. PLA verification retains the source comparison direction.

Consistency returns all three pairwise-overlap covers plus unspecified space
instead of stdout diagnostics. One bounded sharp-work state covers nested
differences and intersections with atomic cancellation. Clones preserve
filename/type metadata and independent label ownership. Unit tests pin both
directional witnesses, label permutation, ownership, exact consistency
witnesses, limits and errors. An independent assignment model fuzzed cover
verification, three overlaps, universe coverage and unspecified minterms for
5,000 executions.

No native helper or retained fuzz corpus is used. The complete suite passes in
17.397s, the race suite in 88.967s, vet is clean, and package-header and
standalone-oracle audits pass. Inventory is now 308 validated, 40 partial and
1,409 unstarted files.

## Espresso output-equivalence checkpoint

`src/misc/espresso/equiv.c` is validated across both definitions.
`EquivalentCovers` preserves semantic mutual containment independent of cube
order and duplication. `FindEquivalentOutputs` performs the source output
cofactor by selecting OFF rows and making the output field unrestricted, then
returns stable lexicographic pair records instead of printing. Equal OFF sets
identify direct equivalence; disjoint, jointly exhaustive OFF sets identify
inversion. This is logically the same complement comparison as the source
without introducing a dependency on the still-unported general complement
engine.

Caller labels and source-compatible defaults are returned in owned typed
records. One bounded sharp-work state accounts for every cofactor, containment,
intersection and exhaustiveness operation and supports synchronous
cancellation. Unit tests pin representation-independent equality, pair order,
polarity, labels, ownership, no-match behavior, limits and errors. An
independent assignment-truth model matched arbitrary small multi-output PLAs
for 5,000 fuzz executions without a native oracle or retained corpus.

The complete suite passes in 17.253s, the race suite in 88.101s, vet is clean,
and package-header and standalone-oracle audits pass. Inventory is now 309
validated, 40 partial and 1,408 unstarted files.

## Espresso complement and simplification checkpoint

`src/misc/espresso/compl.c` is validated across all twelve definitions.
`ComplementCover` computes the exact complement as universe sharp cover,
consolidating the private recursive special cases, split/merge, distance-one
merge and three lifting strategies into the already validated deterministic
sharp engine. `SimplifyCover` and `SimplifyAndComplementCover` use exact double
complementation to preserve the semantic contracts of `simplify` and
`simp_comp`. Returned cube order and shape can differ from the C heuristic, but
the represented function and complement are exact and caller inputs remain
owned and unchanged.

One shared work state bounds generated cubes and packed operations across the
paired computation, checks cancellation synchronously and returns no partial
result. Unit tests pin empty/full covers, a single mixed binary/MV cube,
semantic pairing, ownership, dimensions, limits and cancellation. Independent
assignment enumeration matched arbitrary small covers for 5,000 fuzz
executions without a native oracle or retained corpus.

The complete suite passes in 17.269s, the race suite in 88.882s, vet is clean,
and package-header and standalone-oracle audits pass. Inventory is now 310
validated, 40 partial and 1,407 unstarted files.

## Espresso reduction checkpoint

`src/misc/espresso/reduce.c` is validated across all six definitions.
`SmallestCubeContainingComplement` computes an exact complement through the
validated sharp engine and returns its positional supercube, consolidating the
private SCCC split, merge, cube and special-case recursion. `ReduceCube` and
`ReduceCover` use the equivalent essential-region construction: target sharp
the other current ON cubes and DC cover, then take the positional supercube.
Cover reduction preserves sequential replacement, empty-cube deletion and
`ACTIVE`/`PRIME` behavior.

Both distance-from-largest and column-inner-product source orderings are
available explicitly. This replaces the process-global alternating toggle;
stable ties make concurrent calls deterministic. One shared work state bounds
ordering, sharp and supercube work, checks cancellation synchronously and
returns no partial cover. Unit tests pin SCCC, direct reduction, both orders,
flags, ownership, ON-union-DC preservation, errors and limits. An independent
essential-assignment model matched 5,000 arbitrary small covers for both
orders. The earlier output-equivalence fuzzer was also rerun for 5,028
executions after correcting its independent multi-input encoder; both campaigns
remain Go-only with no retained corpus.

The complete suite passes in 17.380s, the race suite in 88.556s, vet is clean,
and package-header and standalone-oracle audits pass. Inventory is now 311
validated, 40 partial and 1,406 unstarted files.

## Espresso unate-family checkpoint

`src/misc/espresso/unate.c` is validated across all nine definitions.
`UnateComplement` replaces the recursive `unate_complement` and three `abs`
helpers with deterministic minimal-transversal enumeration. Stable first
duplicates and inclusion-minimal results preserve its set-family contract.
`ExactMinimumCovers` retains the historical routine's family of every
inclusion-minimal cover rather than narrowing it to minimum cardinality.
`IntersectUnateFamilies` preserves nonempty pairwise intersection,
maximum-cardinality filtering and inclusion-maximal containment.

An owned `EspressoUnateMapping` replaces global `cdata` with its layout,
active-part/unate-variable association and part-zero counts. Mapping to the
missing-part matrix and back rejects binate inputs and does not alias the
source. Bounds, synchronous cancellation and atomic errors replace fixed
recursive/stack assumptions. Unit tests pin empty and impossible matrices,
intersection modes, mapping ownership, complement round trips and failures. A
brute-force model matched minimal hitting sets and intersections for 5,031
fuzz executions; a second monotone-cover model proved mapping round trips for
5,000 executions, both without a native oracle or retained corpus.

The complete suite passes in 17.259s, the race suite in 88.777s, vet is clean,
and package-header and standalone-oracle audits pass. Inventory is now 312
validated, 40 partial and 1,405 unstarted files.

## Espresso exact-minimization checkpoint

`src/misc/espresso/exact.c` is validated across all four definitions.
`MinimizeExact` and `MinimizeExactLiterals` generate every prime of ON∪DC with
the validated consensus engine, remove DC-only primes and select a valid subset
with bounded deterministic search. Cube mode minimizes cube count and then
literal count, subsuming the source's heuristic sparsification post-pass;
literal mode minimizes source-compatible cover-cost literals and then cube
count. When `exactCover` is false, stable deletion returns an irredundant valid
cover and does not claim optimality.

Inputs must be orthogonal ON/DC/OFF covers. Generation, operation, recursion
depth and subset-search limits plus synchronous cancellation return no partial
result. `EspressoExactResult` owns the prime inventory, selected indices,
objective and final cover, replacing filename/global stdout diagnostics. Unit
tests pin both objectives, heuristic and empty functions, ownership, malformed
partitions and resource failures. An independent model enumerated every valid
positional cube, derived maximal primes and examined every prime subset for
5,037 fuzz executions, matching both objectives and Boolean truth without a
native oracle or retained corpus.

The complete suite passes in 17.267s, the race suite in 88.980s, vet is clean,
and package-header and standalone-oracle audits pass. Inventory is now 313
validated, 40 partial and 1,404 unstarted files.

## Espresso irredundancy checkpoint

`src/misc/espresso/irred.c` is validated across all ten definitions.
`CoverIsTautology` and `CubeCoveredByCover` consolidate recursive cofactor and
special-case logic into exact complement/sharp checks. `SplitIrredundant`
preserves the two-stage E/R then Rt/Rp classification and original row indices.
`DeriveIrredundantTable` partitions the uncovered care space by candidate
membership, returning exact stable source-index rows rather than an intrusive
sparse matrix or global `Rp_current`.

`MarkIrredundant` and `Irredundant` return a source-order marked cover and an
owned active subset, preserving `ACTIVE` and `RELESSEN`. Default selection is
deterministic greedy plus deletion, matching the source heuristic intent;
optional exact selection chooses a minimum-cardinality minimal transversal.
All family/index invariants and cube, row, operation and cancellation bounds
are checked atomically. Unit tests pin classification, table rows, flags,
ownership, tautology, selectors and errors. An assignment model matched every
classification and membership row and proved ON∪DC preservation in both modes
for 5,023 fuzz executions without a native oracle or retained corpus.

The complete suite passes in 17.095s, the race suite in 88.624s, vet is clean,
and package-header and standalone-oracle audits pass. Inventory is now 314
validated, 40 partial and 1,403 unstarted files.

## Espresso expansion checkpoint

`src/misc/espresso/expand.c` is validated across all twelve definitions. Exact
prime enumeration of the OFF-set complement replaces mutable blocking and
covering matrices plus the private essential-part, feasible-selection and
minimum-cover scratch routines. `ExpandCube` chooses a containing prime that
covers the most eligible ON rows, breaking ties by source-compatible literal
cost. `ExpandCover` preserves ascending column-inner-product ordering,
caller-marked prime skipping, later nonprime coverage, row removal and
`PRIME`/`ACTIVE`/`NONESSEN` metadata.

Nonsparse mode restricts global primes to the source cube on sparse variables,
producing maximal dense-only expansions. `FindAllPrimesCoveringCube` and
`AllPrimes` return stable owned candidates and preserve marked-prime
passthrough. Cube/operation/depth bounds, cancellation, ON/OFF orthogonality
and ownership are explicit. Unit tests pin exact prime sets, dense-only
maximality, flags, semantics and failures. A Go-only fuzzer proved truth, OFF
avoidance and independent one-part maximality in both modes and compared every
prime covering a minterm with exhaustive cube enumeration for 5,003
executions.

The complete suite passes in 17.266s, the race suite in 87.888s, vet is clean,
and package-header and standalone-oracle audits pass. Inventory is now 315
validated, 40 partial and 1,402 unstarted files.

## Espresso sparse-cleanup checkpoint

`src/misc/espresso/sparse.c` is validated across both definitions. `MVReduce`
performs the source per-part cofactors and irredundancy marking, removes parts
whose cofactor rows are redundant, preserves the full-variable exception, and
drops cubes emptied in a sparse variable. It returns an owned active cover and
preserves the source `ACTIVE` and `PRIME` transitions. Exact or deterministic
heuristic irredundancy is caller-selectable.

`MakeSparse` alternates sparse reduction with nonsparse expansion until literal
cost stabilizes. Explicit cube, operation, row, depth and iteration bounds plus
synchronous cancellation replace implicit resource assumptions and return no
partial result. Unit tests pin mixed binary/MV behavior, truth, OFF exclusion,
cost, flags, ownership, dimensions, limits and cancellation. An independent
truth-table model checked ON-union-DC preservation, OFF exclusion,
nonincreasing cost and nonempty sparse fields for 5,000 fuzz executions, with
no native oracle or retained corpus.

The complete suite passes in 17.259s, the race suite in 88.133s, and vet is
clean. There are no C/C++/header files or `testdata/` directories under `pkg`;
`pkg/tools/` contains only the source-manifest script. Inventory is now 316
validated, 40 partial and 1,401 unstarted files.

## Espresso GASP checkpoint

`src/misc/espresso/gasp.c` is validated across all six definitions.
`ReduceGasp` independently maximally reduces each row against the original ON
and DC covers with one shared work budget, preserving row order and marking
changed rows nonprime. A redundant input row returns an invariant error rather
than invoking the source fatal exit.

`ExpandGaspCube` and `ExpandGasp` use exact OFF-complement prime enumeration to
decide whether one reduced row can expand over another row's essential region.
Candidate unions are deduplicated and expanded to primes. `IrredundantGasp`,
`LastGasp` and `SuperGasp` compose the validated irredundancy and all-primes
engines. DC space may overlap either minimized polarity, while the active ON
cover must avoid its blocking OFF cover.

Cube, operation, row and depth limits plus cancellation are explicit per
phase, and every result is owned and atomic. Unit tests pin independent
reduction, flags, ownership, both workflows and failure surfaces. A standalone
truth-table model began with expanded irredundant covers and proved
ON-union-DC preservation and OFF exclusion for both workflows over 5,001 fuzz
executions without a native oracle or retained corpus.

The complete suite passes in 17.337s, the race suite in 87.887s, vet is clean,
and native-source, package-header and `testdata/` audits pass. Inventory is now
317 validated, 40 partial and 1,400 unstarted files.

## Espresso main minimizer checkpoint

`src/misc/espresso/espresso.c` is validated across its complete minimization
driver. `MinimizeEspressoCover` preserves optional onset simplification and
output unraveling, prime reset, initial expansion and irredundancy, optional
essential extraction, cube-count-stable reduce/expand/irredundant loops,
last- or super-GASP perturbation, essential restoration and sparse cleanup.
The source rerun without output unraveling is retained when unraveling grows
the result; Go additionally returns the original owned cover if any other
strategy unexpectedly grows it.

Concrete `EspressoMinimizeOptions` replace all strategy globals, while the
result returns cost, iteration, unravel and fallback metadata instead of trace
and timing stdout. Iteration, cube, operation, row and depth limits plus
cancellation are explicit per phase. Unit tests exercise every strategy
branch, both reduction orders, output unraveling, ownership and failures. A
standalone truth-table model varied all major strategies for 5,034 fuzz
executions and proved ON-union-DC preservation, OFF exclusion, non-growth and
returned-cost consistency without a native oracle or retained corpus.

The complete suite passes in 17.448s, the race suite in 89.114s, vet is clean,
and native-source, package-header and `testdata/` audits pass. Inventory is now
318 validated, 40 partial and 1,399 unstarted files.

## Espresso cover-manipulation checkpoint

`src/misc/espresso/cvrm.c` is validated across all twenty definitions.
`UnravelRange` and `Unravel` produce the bounded Cartesian product of selected
coordinates. Lexical, size, column-weight and reduction sorting preserve the
source scores with deterministic stable ties. `ShuffleCover` replaces hidden
random state with a caller seed. `PartitionComponents` preserves first-seed
`ccommon` closure and source indices over owned cube lists.

`CofactorOutput` and `UncofactorOutput` select and restore relative output
parts. `MinimizeOutputs` replaces the source callback functions and global
`Fmin`/`phase` scratch with one concrete loop. It supports heuristic and exact
minimization, optional positive/complement comparison, positive tie preference
and returned phase and cube-count metadata.

All expansion, sort comparison, component, cofactor, cube and operation work
is bounded with synchronous cancellation and atomic owned results. Unit tests
pin Cartesian masks, scores, ownership, seeded order, partitions, cofactors,
both minimizers, polarity and every error class. Separate standalone campaigns
passed 5,000 fuzz executions for unravel/cofactor invariants and 5,000 for
per-output positive-or-complement truth, without a native oracle or retained
corpus.

The complete suite passes in 17.421s, the race suite in 88.691s, vet is clean,
and native-source, package-header and `testdata/` audits pass. Inventory is now
319 validated, 40 partial and 1,398 unstarted files.

## Espresso output-phase checkpoint

`src/misc/espresso/opo.c` is validated across all ten active definitions.
`OutputPhaseSetup` builds the doubled characteristic function for a suffix of
outputs, preserving fixed prefixes, ON/OFF polarity, duplicated DC behavior,
labels and cube metadata. `ApplyOutputPhase` splits ON and OFF rows according
to an owned Boolean phase while leaving DC unchanged.

`SelectOutputPhasePrimes` replaces the static recursive unate-product scratch
with bounded deterministic phase enumeration. Maximizing excludable primes is
implemented directly as minimizing the union of carriers for one polarity of
each output, with positive-first ties. `FindOutputPhase` and
`AssignOutputPhase` compose doubled minimization, one-shot or repeated phase
commitment and final minimization. `EnumerateOutputPhases` replaces `opoall`
and retains its strict cube-count improvement rule.

All former OPO flags, global phase storage and recursion level become concrete
options and returned phase, evaluation and cost metadata. Cube, operation,
phase and exact-search limits plus cancellation are explicit. Unit tests pin
doubled truth, phase application, subset selection, heuristic/exact and
repeated assignment, enumeration and failures. Independent standalone
campaigns passed 5,026 transformation fuzz executions and 5,003 complete
assignment executions without a native oracle or retained corpus. The two
definitions under `#if 0` remain correctly outside active coverage.

The complete suite passes in 17.403s, the race suite in 88.578s, vet is clean,
and native-source, package-header and `testdata/` audits pass. Inventory is now
320 validated, 40 partial and 1,397 unstarted files.

## Espresso variable-pairing checkpoint

`src/misc/espresso/pair.c` is validated across all sixteen active definitions.
`PairVariables` replaces disjoint binary pairs with owned four-valued variables
while preserving the source joint-part encoding, unpaired/pair/original-MV
layout, labels and cube metadata. Paired variables are explicitly nonsparse.
The input PLA and pairing remain unchanged.

`EnumeratePairings` preserves deterministic maximum-cardinality recursion order
for even and odd variable counts. `BestPairing` implements exhaustive and greedy
selection over signed edge costs with checked accumulation. `PairingCosts`
implements weak, strong, full and exact evaluation; `FindOptimalPairing` and
`OptimizePairing` compose selection with heuristic, exact and output-phase
minimization. Concrete options and owned results replace process-global pair
arrays, callbacks, trace output and manual allocation/free.

Unit tests pin exact layouts, labels, truth, enumeration order, formatting,
ownership, all cost/minimization strategies, negative costs, overflow, limits
and cancellation. Independent assignment and exhaustive-matching Go models
passed 5,006 and 5,000 fuzz executions. No C source, compiler, executable or
runtime oracle is used: `pkg/tools/` contains only `source_manifest.py`, and the
module-wide integrity test rejects native sources, external-oracle hooks,
untracked permanent vectors and permanent `testdata/` directories.

The complete suite passes in 17.208s, the race suite in 89.505s, vet is clean,
and native-source, package-header and `testdata/` audits pass. Inventory is now
321 validated, 40 partial and 1,396 unstarted files.

## Espresso PLA-output checkpoint

`src/misc/espresso/cvrout.c` is validated across all eighteen active
definitions. `MarshalPLA`, `FormatHeader`, `FormatCube` and
`FormatExpandedCube` preserve standard ON/DC/OFF output, TPLA termination,
type/header directives, binary and multiple-valued fields, output maps, labels
and phase annotations. Equation, PLEASURE and KISS modes preserve their source
encodings, grouping, exact equation wrapping and symbolic-part restrictions.

`DefaultLabels` returns owned generated labels instead of mutating caller
state. Numeric and symbolic constraint reports preserve projected patterns,
singleton/full exclusion, duplicate weights and deterministic source order.
Owned strings replace `FILE`/stdout effects and the `pc1`/`pc2` static buffers;
owned debug renderings replace verbose process-global output. Byte, cube and
part limits plus cancellation are explicit and atomic.

Fixed Go-only tests pin every output mode, exact rows and headers, labels,
wrapping, phases, constraints, debug output, ownership and all error classes.
An independent cube formatter model passed 5,000 fuzz executions, and a
separate marshal/decode model reconstructed every emitted cube for 5,009
executions. No native oracle or retained corpus is used.

The complete suite passes in 17.235s, the race suite in 88.152s, vet is clean,
and native-source, package-header and `testdata/` audits pass. Inventory is now
322 validated, 40 partial and 1,395 unstarted files.

## Espresso PLA-input partial checkpoint

All eleven definition surfaces in `src/misc/espresso/cvrin.c` now have concrete
owned Go paths, but the file remains partial pending the semantic transformations
for `.symbolic` and `.symbolic-output` delegated by the source to `hack.c`.
`ParseEspressoPLA` handles ordinary `.i`/`.o` and `.mv` layouts, product counts,
termination, KISS mode, F/D/R types, input/output/MV labels, phases, variable
pairs, comments and unknown directives. Product parsing preserves binary,
symbolic-MV and ON/DC/OFF character encodings and source-compatible malformed-
product warnings.

Postprocessing constructs missing OFF, DC or ON covers, applies positive or
explicit output phase, performs binary-variable pairing and merges the KISS
next-state/output fields. Symbolic directives are parsed into owned references
and labels when explicitly deferred; the default returns `ErrUnsupported`
instead of silently dropping their meaning. `EspressoPLAParseResult.Summary`
returns the former stdout report, while existing owned PLA construction, labels
and Go lifetime replace allocation/free helpers.

Input bytes, lines, variables, parts, cubes, tokens and operations are bounded,
with atomic cancellation. Fixed tests pin directives, exact cover contents,
labels, phases, KISS, pairing, computed covers, summaries, warnings and errors.
Independent Go-only campaigns passed 5,000 marshal/parse round trips and 5,030
arbitrary-byte robustness executions. Product terms are currently line-scoped;
the source's warning-only physical-line continuation remains to be completed.

The complete suite passes in 17.235s, the race suite in 93.001s, vet is clean,
and native-source, package-header and `testdata/` audits pass. Inventory is now
322 validated, 41 partial and 1,394 unstarted files.

## Espresso PLA-input completion and symbolic-hack checkpoint

`src/misc/espresso/cvrin.c` is now validated across all eleven definitions.
Physical newlines inside products retain the source warning-only continuation,
and `.symbolic` plus `.symbolic-output` apply by default. Explicit deferral
remains available to return the owned pre-mapping PLA and directive metadata.
The refreshed parser campaigns passed 5,000 marshal/parse round trips and 5,015
arbitrary-byte executions.

Seven of nine active definitions in `src/misc/espresso/hack.c` are ported.
`ApplySymbolicInputs` replaces selected binary fields with ordered MV bit-vector
fields, including unconstrained combinations and labels. `ApplySymbolicOutputs`
preserves one-hot group order, F/R cofactor recursion, DC-driven onset
recomputation, removal of old outputs, labels and quick containment/merge
cleanup. The historical omission of an all-zero mapped output when no other ON
output exists is pinned rather than silently corrected.

`MapDontCareVariable` recognizes the same four case-sensitive label prefixes,
derives cofactor disagreement as the DC set, removes dependent ON cubes and the
control variable, and returns a consistent owned OFF cover. Cube, operation and
depth bounds plus cancellation cover all transformations. Independent Go-only
input-symbolic, output-symbolic and named-DC semantic models each passed 5,000
fuzz executions. `fsm_simplify` and `disassemble_fsm` remain before `hack.c`
can be promoted from partial.

The complete suite passes in 17.414s, the race suite in 89.015s, vet is clean,
and native-source, package-header and `testdata/` audits pass. Inventory is now
323 validated, 41 partial and 1,393 unstarted files.

## Espresso symbolic-hack completion checkpoint

`src/misc/espresso/hack.c` is now validated across all nine active definitions.
`DisassembleFSM` completes `fsm_simplify` and `disassemble_fsm` with concrete
owned processing of every-state, nowhere and specific-state transitions. Each
partition is selected in deterministic source order, minimized independently,
and reassembled with separate present-state, next-state and optional external-
output fields. The result owns its statistics, optional diagnostics, PLA and
KISS rendering; it neither writes stdout nor uses process-global scratch.

The Go representation deliberately omits a zero-sized external-output part
when the machine has no external outputs. Bounds cover cubes, operations,
rows, recursion depth, minimizer iterations and emitted bytes, with synchronous
cancellation and no partial result on error. Fixed tests pin partition order,
statistics, exact diagnostics, exact KISS text, ownership, zero-output behavior
and every error class. An independent Go relation model passed 5,000 fuzz
executions over generated valid FSM covers.

The module contains no C/C++ sources, headers, native objects or native
executables. `pkg/tools/` contains only `source_manifest.py`; permanent evidence
is immutable data under `test_vectors/`, while semantic expectations are
checked by Go-only models, exhaustive checks, replay, round trips and
invariants. The integrity test rejects native source, external-oracle hooks,
untracked vectors and permanent `testdata/` directories.

The complete suite passes in 17.248s, the race suite in 88.741s, vet is clean,
and the standalone-oracle audit passes. Inventory is now 324 validated, 40
partial and 1,393 unstarted files.

## Espresso global-state replacement checkpoint

`src/misc/espresso/globals.c` is validated across all 27 declarations without
recreating its writable process globals. `EspressoRuntimeOptions` owns driver
flags per invocation and supplies the source main driver's defaults. Its
minimizer projection maps every active control, including random reduction
order. Random traversal uses the source seed `314973` through owned,
deterministic per-run shuffling instead of libc process-global RNG state.

`EspressoPLATypeDescriptors` returns an owned copy of all 20 format spellings,
and `LookupEspressoPLAType` accepts their command-line forms with or without the
leading dash. `NewEspressoRuntimeTotals` returns independent timing names,
counts and elapsed storage. Existing `EspressoPLA`, input options, immutable
`EspressoCubeLayout` and operation-local work structs own the former filename,
cube, temporary cube and cdata state. `math/bits` replaces the 256-entry byte
population table.

Tests cover defaults, option projection, all type aliases, copy ownership,
concurrent independent instances and every byte value. A separate standalone
Go campaign passed 5,000 fuzz executions. The complete suite passes in 17.551s,
the race suite in 88.774s, vet is clean, and the native-source and permanent-
`testdata/` audits pass. Inventory is now 325 validated, 40 partial and 1,392
unstarted files.

## Espresso driver-table checkpoint

`src/misc/espresso/main.h` is validated. `EspressoCommandKind` and the 43
ordered `EspressoCommandDescriptor` rows preserve every source command name,
alias, PLA arity and OFF/DC input requirement. Lookup remains case-sensitive,
including the source's uppercase default `ESPRESSO` and `PLAverify` spelling.

All 16 debug selectors preserve their exact bit combinations. The eleven
behavior options apply as pure transformations of `EspressoRuntimeOptions`
rather than table entries pointing into writable globals. Descriptor accessors
return owned copies, and unknown options return `ErrUnsupported` without
changing the input value.

Exact-table tests cover order, aliases, masks, arities, requirements, ownership
and failure behavior. A standalone Go campaign passed 5,000 fuzz executions.
The complete suite passes in 17.282s, the race suite in 89.914s, vet is clean,
and standalone-oracle audits pass. Inventory is now 326 validated, 40 partial
and 1,391 unstarted files. `main.c` remains unstarted until command parsing and
dispatch are implemented; table coverage is not being used to claim driver
execution.

## Espresso driver-parser checkpoint

`src/misc/espresso/main.c` moves from unstarted to partial.
`ParseEspressoDriverArguments` replaces `backward_compatibility_hack`,
`delete_arg`, `check_arg`, and the getopt portion of `main` with bounded parsing
that never mutates caller arguments or exits the process. It preserves legacy
`-do`/`-out` pairs, bare behavior switches, the source's `-fdr`/`-fr`/`-f`
precedence, and attached or separate modern option values.

The returned concrete configuration owns command metadata, input/output type,
runtime switches, strategy, range and positional paths. Command arity, numeric
syntax, debug selectors, byte/argument limits and cancellation are checked.
`EspressoDriverUsage` returns deterministic bounded-width text instead of
printing, and includes every registered command.

Unit tests pin defaults, legacy and modern parsing, precedence, ownership,
limits, errors and usage width. A standalone parser campaign passed 5,000 fuzz
executions. The complete suite passes in 17.578s, the race suite in 88.774s,
vet is clean, and standalone-oracle audits pass. Inventory is now 326 validated,
41 partial and 1,390 unstarted files. PLA loading orchestration, actual command
dispatch, runtime summary, output handling and the many-PLA loop remain before
`main.c` can be validated.

## Espresso standalone-driver completion checkpoint

`src/misc/espresso/main.c` is validated across all nine definition surfaces.
`ExecuteEspressoCommand` dispatches every command group to concrete Go
algorithms: heuristic/exact and per-output minimization, phase assignment,
pairing, cover algebra, verification, mapping, FSM processing, individual
Espresso passes and the diagnostic workflow. Results own their PLA, decisions,
verification witnesses, phase/pair metadata and diagnostic text.

`RunEspressoDriver` composes argument parsing, bounded PLA parsing, required-
cover construction, dispatch and bounded output marshaling over caller-owned
bytes. It implements print suppression and the `many` batch loop. A future thin
CLI may adapt files/stdin and exit codes; those process concerns do not leak
into the package. `FormatRuntime` reports caller-supplied elapsed values, so
tests and library users do not depend on hidden clocks.

Fixed tests cover command groups, exact Boolean results, transformations,
phase/pair semantics, verification, legacy/modern arguments, end-to-end bytes,
batch processing, output suppression, runtime formatting, limits, cancellation
and ownership. Independent Boolean dispatcher and complete byte-driver
campaigns each passed 5,000 executions. The complete suite passes in 17.249s,
the race suite in 88.845s, vet is clean, and standalone-oracle audits pass.

All 39 translation units selected by `src/misc/espresso/module.make` are now
individually validated, so that manifest is validated as well. Inventory is now
328 validated, 40 partial and 1,389 unstarted files.

## Espresso header completion checkpoint

`src/misc/espresso/espresso.h` is validated as executable header logic rather
than copied ABI scaffolding. Typed constants preserve all PLA-format, set-flag,
debug, Boolean, cube-input and sixteen runtime-slot numeric values.
`EspressoSet` and `EspressoSetFamily` replace header-word arrays and inline
macros with checked packed words and owned metadata. `EspressoCubeLayout`,
operation-local analysis, cover cost, pairing, symbolic directives and
`EspressoPLA` replace the global structs and linked records. Every algorithm
prototype is accounted for by an individually validated source entry.

Fixed tests pin constants, source normalization, masks, tail words and clone
ownership. A standalone packed-representation campaign passed 5,000 fuzz
executions. Errors, limits and cancellation replace fatal/assert/timing macros;
the Go API intentionally exposes no C layout.

`src/misc/espresso/util_old.h` is also validated with an explicit replacement
disposition. It has no include edge in the current tree and is absent from the
module manifest. Its obsolete platform branches, allocator/libc declarations,
fatal macros, getopt globals and clock aliases are replaced by Go ownership,
standard-library operations, concrete driver parsing, explicit errors and
caller-supplied elapsed values rather than copied.

The complete suite passes in 17.122s, the race suite in 90.156s, vet is clean,
and standalone-oracle audits pass. Every tracked item under
`src/misc/espresso/` is now validated. Inventory is now 330 validated, 40
partial and 1,387 unstarted files.

## Extra progress-bar checkpoint

`src/misc/extra/extraUtilProgress.c` is validated across all five routines.
`ExtraProgressBar` owns the former allocation and update state, preserves the
fixed 78-column geometry, initial position, integer threshold recurrence,
seven-column update steps, completion saturation and one-step-per-call
behavior. Start retains the disabled null-handle result, while batch mode
updates state without producing text.

Frames and the final clearing line are returned as owned strings instead of
retaining a `FILE*`, consulting global frame state or flushing stdout. Output
bounds, checked dimensions and atomic errors replace unchecked formatting and
allocation. Stop explicitly closes the state; later operations report an
invariant error.

Fixed tests pin exact frames, thresholds, disabled and batch modes, lifecycle,
limits and failure atomicity. An independent recurrence-model campaign passed
5,002 Go-only fuzz executions. The full suite passes in 17.451s, the race suite
in 89.482s, vet is clean, and the standalone-oracle audit passes. Inventory is
now 331 validated, 40 partial and 1,386 unstarted files.

## Bar progress facade checkpoint

`src/misc/bar/bar.c` is a behaviorally identical copy of the extra progress
implementation, so `BarProgress` deliberately aliases the same concrete owned
state instead of maintaining a second engine. `StartBarProgress`,
`UpdateBarProgress` and `StopBarProgress` preserve the source-shaped entry
points and disabled nil-handle no-ops. The shared implementation pins the exact
threshold recurrence, text frames, batch behavior and cleanup.

`src/misc/bar/bar.h` is validated as contract rather than copied C layout.
`BarProgressUse` preserves `BAR_PROGRESS_USE == 1`; the opaque record maps to
the concrete alias, and the inline first-field threshold shortcut maps to the
same checked early return in `Update`. Focused tests exercise the constant,
nil facade, fast path, completion and cleanup. The shared independent model
already passed 5,002 fuzz executions.

Because `src/misc/bar/module.make` selects only the now-validated `bar.c`, the
manifest is validated too. The full suite passes in 17.504s, the race suite in
89.620s, vet is clean, and standalone-oracle audits pass. Inventory is now 334
validated, 40 partial and 1,383 unstarted files.

## Glucose constant-contract checkpoint

The identical `src/sat/glucose/Constants.h` and
`src/sat/glucose2/Constants.h` files are validated without claiming either
solver implementation. Typed Go constants preserve the three enabled feature
switches, clause-removal ratio `2`, and blocking-restart lower bound `10000`.
`DefaultGlucoseConstantConfig` returns an independent concrete value for later
solver construction rather than exposing preprocessor state.

`src/sat/glucose2/CGlucose.h` is also validated: its only contract,
`CGLUCOSE_EXP == 1`, is `Glucose2CircuitExperiment`. Exact-value tests cover
the complete finite surface and verify independent returned configurations.
The surrounding Glucose and Glucose2 sources remain at their individual ledger
statuses; these header promotions do not stand in for a SAT solver.

The full suite passes in 17.235s, the race suite in 88.512s, vet is clean, JSON
is valid, and standalone-oracle audits pass. Inventory is now 337 validated,
40 partial and 1,380 unstarted files.

## Extra packed bit-matrix checkpoint

`src/misc/extra/extraUtilBitMatrix.c` is validated across all 16 routines.
`ExtraBitMatrix` uses contiguous `uint32` rows, preserving the source's exact
word geometry, upper- and lower-triangle normalization, diagonal storage,
duplicate mutation counters, raw row union and simultaneous two-row union.
Strict-upper counting, full-storage disjointness and clique-equivalence checks
retain their distinct source semantics.

Triangular printing is now bounded `FormatUpper` text. Size, copied row
snapshots and `uint64` statistics expose owned state; `Clear` retains counters
and `Close` replaces manual freeing. Dimensions, allocation/work/output bounds,
cancellation and scan atomicity are checked. Zero-sized matrices are a safe Go
extension instead of exercising the source's zero-allocation dereference.

Fixed tests cover both storage orientations, exact output, operation counters,
row unions, disjointness, valid and invalid clique relations, ownership,
lifecycle and error paths. An independent Boolean-relation state-machine
campaign passed 5,000 Go-only fuzz executions. The complete suite passes in
17.431s, the race suite in 89.283s, vet is clean, and standalone-oracle audits
pass. Inventory is now 338 validated, 40 partial and 1,379 unstarted files.

## Constant multiplier-accumulator checkpoint

`src/misc/extra/extraUtilMacc.c` is validated across all ten entry points.
`GenerateMaccPlan` preserves the signed low-bit table, trivial and shift
sentinels, source adder-level traversal, first recipe retained for each value,
packed add/subtract operands and negative completion. The plan checks every
signed coefficient by traversing its recipe DAG and returns owned snapshots.
The packed format is explicitly bounded to 15 coefficient bits.

Direct multiply/MACC templates and synthesized multiplier, narrow accumulator
and wide accumulator templates return bounded deterministic text. Timestamps
are caller supplied. Ordered `MaccGeneratedFile` slices replace filesystem-
writing test drivers, and shared recipe declarations are emitted once so the
result is valid deterministic module text. Dimensions, work, entries, output,
cancellation and timestamp syntax are checked instead of asserting or relying
on implicit files.

Fixed tests exhaust every signed coefficient through nine bits, pin an exact
direct module, cover every synthesis mode and zero case, batch order/names,
ownership, limits and cancellation. An independent arithmetic campaign passed
5,000 Go-only fuzz executions. The complete suite passes in 17.165s, the race
suite in 93.508s, vet is clean, and standalone-oracle audits pass. Inventory is
now 339 validated, 40 partial and 1,378 unstarted files.

## Extra canonical BDD checkpoint

`src/misc/extra/extraUtilMult.c` is validated across all 25 routines.
`ExtraBDDLit` preserves false/true and low-bit complement encoding.
`ExtraBDDManager` preserves variable-node IDs, wrapping source hash arithmetic,
chained unique insertion, regular else-edge normalization, direct-mapped cache
replacement, numeric argument ordering, canonical AND and De Morgan OR.
Failed or cancelled operations roll back nodes, unique links, cache entries and
statistics atomically.

Owned visitation replaces the source mark scratch for single-root,
union-shared and per-root-summed counts. The latter deliberately preserves
`Abc_BddGiaTest`'s actual behavior even though its diagnostic calls the result
shared. SOP formatting returns exact else-first text; the four-variable print
test is a returned concrete example, statistics are data, and `Close` replaces
printing plus manual freeing.

`BuildExtraBDDFromGia` converts ordinary topological AND graphs with phases and
ordered COs without touching GIA scratch. Optional SOPs and per-output summed
counts are returned. Buffers and native XOR/mux nodes reject explicitly rather
than being misread as ANDs. Fixed tests exhaust example and GIA truth tables,
pin text/canonical identities/cache/count/lifecycle behavior and prove failure
atomicity. A standalone truth-table state-machine campaign passed 5,000 fuzz
executions. The complete suite passes in 17.358s, the race suite in 88.033s,
vet is clean, and standalone-oracle audits pass. Inventory is now 340
validated, 40 partial and 1,377 unstarted files.

## Stand-alone oracle closure audit

The complete `pkg` tree contains no C/C++ source or header and no `testdata`
directory. No Go test reads the upstream `src` tree, invokes ABC or a compiler,
uses cgo, or contains a generator hook. Dataset-backed tests read permanent,
hash-pinned inputs only from `test_vectors/` and pair those rows with Go models,
bounded exhaustive checks, invariants, replay, or round trips. Consequently the
test suite has no runtime or regeneration dependency on the C implementation.

All historical C oracle programs have been removed from `pkg/tools/`. The only
remaining file there is `source_manifest.py`, a planning-only inventory script
that hashes the upstream source and is never called by Go tests. The package
policy test now uses an exact allowlist for that one file; another Python script
cannot be added as a disguised compiler or ABC wrapper without failing the Go
suite. Native source, symlinks, external-command hooks, and permanent
`testdata/` remain independently forbidden.

## Extra majority-network enumeration checkpoint

`src/misc/extra/extraUtilMaj.c` is validated across all 14 routines.
`GemManager` owns dynamically grown object and truth-table storage instead of
eagerly reserving ten million objects. Full-width stretched truth keys preserve
the source memory table's cross-width deduplication, while concrete
`GemObject` values retain variable count, majority-node count, predecessor,
history and symmetric-group boundary masks.

Expansion directly substitutes a fresh three-input majority for the selected
input. Crossbar reduction ties both selected inputs to one fresh input. Each
result passes through the validated permutation-only truth canonicalizer before
group derivation and insertion. Group remove/insert/unpack operations and the
cumulative one- or three-variable placement search preserve their source
ordering; the latter explicitly preserves little-endian `memcmp` selection.
Fully symmetric functions now return their weight polar and majority status as
structured data.

Progress stages, optional truth dumps, node diagnostics and statistics are
owned bounded results rather than stdout or files. Limits, cancellation,
snapshot ownership, insertion atomicity and explicit close are covered. Fixed
tests pin group algebra, placement, MAJ3 expansion, reduction back to the
buffer, deduplication, metadata, formatting and enumeration. An independent
direct-substitution and exhaustive permutation-orbit campaign passed 5,000
Go-only fuzz executions. No native oracle or retained corpus is used. Inventory
is now 341 validated, 40 partial and 1,376 unstarted files.

## Extra memory-manager checkpoint

`src/misc/extra/extraUtilMemory.c` is validated across all 18 exported
routines. Concrete `ExtraFixedArena`, `ExtraFlexArena` and `ExtraStepArena`
facades reuse the already validated owned byte-arena core rather than copying a
third linked-list allocator. There are no interfaces or borrowed unchecked C
pointers.

Fixed allocation preserves the 1,024-entry/64-KiB/minimum-eight chunk rule,
LIFO address reuse, first-chunk restart, persistent peak and source accounting.
Flexible allocation preserves its distinct 4-KiB initial chunk, doubling to
twice an oversized request and bump allocation. Step allocation preserves
8/16/... power-of-two classes, the `4 << steps` map boundary, nil zero-size
requests, retained large blocks, no-op large recycling and class-only memory
usage. High size classes are initialized without allocating their first chunk;
global byte limits apply when a class is actually used.

Blocks become inaccessible after fixed recycling, restart or arena close.
Cross-arena, double and wrong-size recycling reject instead of corrupting free
lists. Byte, block and output limits plus cancellation replace unchecked
allocation and printing. Fixed tests pin chunk boundaries, two-chunk growth,
restart/peak behavior, exact text, flexible exact-fill and resize behavior,
every step class, large retention, accounting exclusions, deferred high
classes, ownership and lifecycle. A mixed independent state-machine campaign
passed 5,000 Go-only fuzz executions. Inventory is now 342 validated, 40
partial and 1,375 unstarted files.

## Extra token-reader checkpoint

`src/misc/extra/extraUtilReader.c` is validated across all nine routines.
`ExtraFileReader` owns its input and returns concrete copied token/line batches.
An owned whole-input buffer replaces the mutable 4-MiB reload window while
preserving observable comment, stop and clean-character precedence, the
special `//` comment rule, slash fallthrough, empty-group skipping, one-based
line numbers, current byte position, final-token delivery and the source's
unterminated-comment failure.

Both file and byte constructors are available. The byte path makes tests and
embedded callers independent of filesystem state; the file path checks size
before and after loading. NUL bytes reject instead of exposing silently
truncated C strings. Byte, token, token-length and operation limits and
cancellation return without changing parser position. Close invalidates owned
storage, and all returned slices are independent.

Fixed tests cover classification precedence, comments, slash behavior, stops,
clean delimiters, empty batches, lines, positions, EOF, file loading,
ownership, malformed input, limits, cancellation and lifecycle. An independent
parser model passed 5,000 arbitrary-input fuzz executions. No native oracle or
retained corpus is used. Inventory is now 343 validated, 40 partial and 1,374
unstarted files.

## Extra support-minimization checkpoint

`src/misc/extra/extraUtilSupp.c` is validated across all 22 definitions.
Concrete `uint64` rows replace word vectors for M-out-of-N masks, ordered unique
XOR pairs and duplicate-preserving difference matrices. `ExtraSupportGreedy`
preserves the source parity-tag profile, first-maximum ties, strictly improving
merit, selection and filtering. Zero difference rows now reject explicitly
because they make the C minimization loop non-terminating.

Verification returns copied linear codes and the exact equal-pair collision
count. The second solver preserves singleton protection, minimum-double tie
selection and iterative removal, returning its active mask, removal order and
final profile. Low-bit-first mask and profile printing are bounded exact text.
The MIN reader preserves its marker, delimiters, `10`/`01` literals, long cube
terminators, `INPUT` prefix stop, declared count and equal-width checks; file
loading and the old print-only driver are concrete returned workflows.

Fixed tests pin mask and pair order, duplicate/zero behavior, primitive state
transitions, deterministic minimization, injective codes, reduction ties and
irredundancy, exact formatting, MIN files and malformed cases, limits,
cancellation and atomicity. An independent campaign generated unique cube sets
and proved both code injectivity and necessity of every retained variable for
5,000 Go-only fuzz executions. Inventory is now 344 validated, 40 partial and
1,373 unstarted files.

## Extra SIS-utility checkpoint

`src/misc/extra/extraUtilUtil.c` is validated across all 13 definition
surfaces. `ExtraOptionParser` owns a copy of its arguments and parsing state
instead of using the source's process-global argument pointer, index and scan
cursor. It preserves option clusters, attached and separate required
arguments, index advancement, non-option and `--` termination, reset, and
structured unknown/missing-argument results.

The fallback data limit, millisecond text, nullable string copying and the
intentional no-op tilde expansion are preserved. File checking supports the
source modes, including `x` as read and `w` truncation. File search preserves
colon order and dot-component spelling while adding path/candidate bounds and
cancellation. Allocation failure is a returned typed error rather than a
global stream flush, assertion and process exit; the mutable function-pointer
hook is intentionally absent. Runtime user-CPU metrics provide a portable
replacement for the platform-dependent clock implementations, and the memory
test remains a no-op.

Fixed Go-only tests cover every behavior, filesystem side effect, error policy,
ownership property, limit and timing monotonicity. An independent option-parser
model passed 5,000 fuzz executions. No native oracle or retained corpus is
used. The full suite passes in 17.113s, the race suite in 89.547s, vet is clean,
and native-source, package-header, `testdata/` and tool-allowlist audits pass.
Inventory is now 345 validated, 40 partial and 1,372 unstarted files.

## Extra truth-utility checkpoint

`src/misc/extra/extraUtilTruth.c` is validated across all 20 routines. The
owned `TruthTable` core supplies elementary variables, swaps, support,
cofactors, existential and universal quantification, cofactor muxing, input
phase changes and symmetry without introducing another truth representation.
Phase-mask stretch and shrink preserve low-to-high selected-variable order;
shrink additionally rejects silently discarding an active input.

Source-specific helpers preserve first-minimum cofactor-support ties, the
five-input filter and `32/-1` sentinel; the replicated 32-bit small-table
cofactor-counting space; and the exact wrapping prime-weighted uint32 XOR hash.
The Extra semicanonicalizer preserves the older routine's balanced-output
assignment-zero tie, input-phase normalization and stable adjacent signature
sorting. It returns owned tables, both permutation directions and phase vectors
that replay the complete transformation. Dynamic storage replaces the static
16-word scratch arrays and extends safe operation beyond nine variables where
the underlying table limits permit it.

Fixed Go-only tests cover every operation, representation boundary, invalid
dimension, ownership rule, sentinel, exact count/hash and transformation
metadata. An independent assignment-model campaign passed 5,021 fuzz
executions over cofactors, quantifiers, mux reconstruction, flips, swaps,
counts and semicanonical replay. No native oracle or retained corpus is used.
The full suite passes in 17.318s, the race suite in 87.950s, vet is clean, and
native-source, package-header, `testdata/` and tool-allowlist audits pass.

Inventory is now 346 validated, 40 partial and 1,371 unstarted files.

## Extra finite-ZDD truncation checkpoint

`src/misc/extra/extraZddTrunc.c` is validated across its complete public and
private definition surface. Immutable `ExtraZDD` values explicitly own finite
subset families. Construction sorts and uniquifies each subset as unate product
does, removes duplicate subsets as union does, and orders the family by the
source traversal's increasing-index exclude-before-include decisions. A
singleton-variable constructor replaces direct CUDD node creation.

Truncation multiplies probabilities only for included variables, preserves
strict below-limit pruning and equality inclusion, and returns copied subsets
in traversal order plus the source's count/subsets/`-1` flat encoding. Go
slices replace all eight private `Vec_Int_t` allocation and mutation helpers.
Exact former stdout text and the complete built-in example are returned as
owned values. Checked probabilities, dimensions, bounds and cancellation
replace assertions, unchecked allocation, reference counting and global
printing.

Fixed Go-only tests cover canonicalization, duplicate inputs, ordering,
ownership, singleton families, the exact example, flat encoding, formatting,
threshold equality, malformed probabilities and every limit. An independent
exhaustive mask model uses bit-reversed ordinal enumeration to verify ordering
and filtering for 5,000 fuzz executions. No CUDD or native oracle is retained.
The full suite passes in 17.119s, the race suite in 89.851s, vet is clean, and
native-source, package-header, `testdata/` and tool-allowlist audits pass.

Inventory is now 347 validated, 40 partial and 1,370 unstarted files.

## Extra module-manifest checkpoint

`src/misc/extra/module.make` moves from unstarted to partial. Twelve of its 20
selected translation units are validated: `BitMatrix`, `Cfs`, `Gen`, `Macc`,
`Maj`, `Memory`, `Mult`, `Progress`, `Reader`, `Supp`, `Truth` and `Util`.
`Canon`, `Cube`, `Dsd`, `Enum`, `File`, `Misc`, `Path` and `Perm` remain
unstarted. The ledger records that exact boundary; the manifest cannot become
validated until all eight remaining selected sources do. Inventory is now 347
validated, 41 partial and 1,369 unstarted files.

## Extra input-phase canonicalization checkpoint

`src/misc/extra/extraUtilCanon.c` is validated across both canonicalization
routines, both static lookup tables and all three diagnostic drivers.
`ExtraTruthCanonFastN` exhaustively visits the complete input-phase orbit of a
zero-through-six-input function and selects the minimum packed integer value.
It returns every minimizing phase in ascending order. For fewer than three
real inputs it still enumerates three phase bits, preserving the source table's
duplicate phases for unused inputs.

Owned truth tables and phase slices replace recursive static buffers and the
hard-coded 256-row truth/phase arrays. Higher declared inputs must be unused
rather than being silently ignored. Candidate limits and cancellation replace
assertions. Generated three-input rows replace the table printer and its
driver; exhaustive tests replace the four-input diagnostic loop.

Every one of the 256 three-input functions and all 65,536 four-input functions
matches an independent bit-level orbit model. A separate campaign passed 5,000
arbitrary zero-through-six-input fuzz executions, verifying all returned
minimizing phases and truth assignments without a native oracle or corpus.
The Extra manifest now has 13 of 20 selected units validated; `Cube`, `Dsd`,
`Enum`, `File`, `Misc`, `Path` and `Perm` remain. Inventory is now 348
validated, 41 partial and 1,368 unstarted files. The full suite passes in
17.802s, the race suite in 90.669s, vet is clean, and native-source,
package-header, `testdata/` and tool-allowlist audits pass.

## Extra cube-group enumeration checkpoint

`src/misc/extra/extraUtilCube.c` is validated across both enumerators and all
eight private state, bitset, field and signature helpers. `ExtraCubeState` is a
concrete 24-sticker permutation. The three literal nine-swap quarter turns,
their half/three-quarter powers, source composition direction and subsequent
nine-generator order are preserved exactly.

Full-state mode deduplicates comparable permutations. Signature mode preserves
the seven selected positions, corner/orientation maps and five-bit fields but
uses a lookup-only `map[uint64]struct{}` instead of eagerly allocating the
source's 4-GiB bitmap. The BFS preserves identity and first-nine seeding,
frontier boundaries and convergence order, returning layer totals and optional
owned states instead of timing/stdout. Limits, cancellation and errors replace
fixed allocation and assertions.

Fixed tests pin all generator powers, signatures, composition, exact state
formatting, ownership and the published layers through 12,224 states in both
deduplication modes. An opt-in full regression reproduced the complete
`1,10,...,3674160,3674160` sequence in 7.909s. Full-vs-signature fuzzing passed
5,000 executions without a collision. The Extra manifest now has 14 of 20
selected units validated; `Dsd`, `Enum`, `File`, `Misc`, `Path` and `Perm`
remain. The full suite passes in 17.588s, the race suite in 91.028s, vet is
clean, and native-source, package-header, `testdata/` and tool-allowlist audits
pass. Inventory is now 349 validated, 41 partial and 1,367 unstarted files.

## Extra file/text utility checkpoint

`src/misc/extra/extraUtilFile.c` is validated across all 33 definition
surfaces. Owned strings preserve last-dot extension/generic behavior, suffix
composition, unusual `>` path correction, basename/path splitting, first-dot
design extraction, strict suffix tests and ordered similar-file probing.
Concrete `*os.File` and path operations preserve whole-file reads and their
unconditional appended newline. Inspection and size now return explicit errors.

Binary and uint32-packed hexadecimal operations preserve high/low bit order,
nibble order, lower/upper case, prefix/reverse variants and special zero/one
input truth strings. The legacy hex printer rejects widths where its C shift is
undefined. Timestamp, symbol and string helpers return owned text. Sorting
retains CR bytes, removes empty lines and sorts lexically. Line numbering
simulates 999-byte `fgets` chunks and the source's `strlen-2` overwrite—so its
LF-only last-character quirk is tested—while rejecting unsafe short or NUL
chunks. File transforms return errors instead of printing.

Fixed Go-only tests cover all name/path rules, real filesystem reads and
writes, empty/missing files, exact binary/hex/text forms, sorting, numbering,
ownership, limits and cancellation. A round-trip model passed 5,000 fuzz
executions. The Extra manifest now has 15 of 20 selected units validated;
`Dsd`, `Enum`, `Misc`, `Path` and `Perm` remain. Inventory is now 350 validated,
41 partial and 1,366 unstarted files. The full suite passes in 17.686s, the race
suite in 90.410s, vet is clean, and native-source, package-header, `testdata/`
and tool-allowlist audits pass.

## Extra function-enumeration checkpoint

`src/misc/extra/extraUtilEnum.c` is validated across both enumeration engines,
all three fixed datasets and every private manager/helper definition. The fixed
problems preserve row minterm masks, transposed variable truths and target
function masks. `EnumerateExtraFunctions` preserves global truth deduplication,
cost layers, `a+b+1` operand selection, `i<k`, AND-before-OR insertion, target
discovery order and source expression syntax. The built-in second problem
matches the exact prefix `[0,10,88,640,4200,23608,116764]` and discovers target
six as `a+c` at one gate.

The structural engine preserves all `Abc_EnuMan_t` fanin, complement, polarity,
reference, top, level and truth state; new-level traversal; four AND polarities;
optional XOR; NPN restrictions; complement equivalence; constant, prior-node
and cheaper-pair filters; and tries/builds/finished counters. Completed networks
and exact gate text are owned results rather than stdout. Lookup-only maps and
slices replace bitsets/vectors; limits and cancellation replace fixed arrays.

Fixed tests pin dataset transposition, primitive AND/XOR counts, feasibility,
all captured structures, expressions, deterministic counters and every error
path. A 5,000-case campaign independently reevaluated every synthesized truth
and completed network. The Extra manifest now has 16 of 20 selected units
validated; `Dsd`, `Misc`, `Path` and `Perm` remain. Inventory is now 351
validated, 41 partial and 1,365 unstarted files. The full suite passes in
17.930s, the race suite in 91.548s, vet is clean, and native-source,
package-header, `testdata/` and tool-allowlist audits pass.

## Extra grid-path checkpoint

`src/misc/extra/extraUtilPath.c` is validated across all 19 public/private
definitions. `BuildExtraMonotonePathsGIA` preserves the source's horizontal and
vertical variable layout and its right/down dynamic program. `BuildExtraGridGraph`
preserves row-wise edge order, first/last node lifetimes and ordered live
frontiers. `BuildExtraGridFrontier` preserves the packed degree/component
states, endpoint/internal degree rules, early cycle rejection, component
canonicalization, zero-before-one state insertion and terminal transitions.
Counts are exact `big.Int` values rather than potentially rounded doubles.

`ExtraGridFrontier.DeriveGIA` preserves the memoized edge-variable mux graph.
Path rendering and the derived AIG filename are returned as owned values rather
than printed or written implicitly. `SolveExtraGraphPaths` preserves
output-constrained lexicographic SAT enumeration and assignment blocking while
returning every sampled model and defined statistics; a zero span no longer
causes undefined division. Concrete limits, cancellation and errors replace
fixed buffers and assertions, with no interfaces.

Go-only tests independently recognize selected grid edges as a single simple
corner-to-corner path. They exhaust every edge assignment through 3x3 against
the derived GIA, pin the known 2/12/184 counts through 4x4, exhaust the monotone
GIA through size two, and cover exact ordering, formatting, lexical SAT
results, ownership and error paths. A separate campaign passed 5,000
executions without native code or retained fixtures. The Extra manifest now
has 17 of 20 selected units validated; `Dsd`, `Misc` and `Perm` remain.
Inventory is now 352 validated, 41 partial and 1,364 unstarted files. The full
suite passes in 18.042s, the race suite in 90.815s, vet is clean, and
native-source, package-header, `testdata/` and tool-allowlist audits pass.

## Extra permutation-ZDD checkpoint

`src/misc/extra/extraUtilPerm.c` is validated across its complete active
manager, operation, permutation, printing and diagnostic surface.
`ExtraPermutationZDD` preserves terminal and variable IDs, source hash
arithmetic, zero-suppressed unique creation, direct-mapped cache replacement,
canonical node order and permutation pair-variable numbering. Operations are
concrete and bounded; failed or cancelled exported operations are transactional.
Exact `big.Int` path counts replace overflowing native integers.

Difference, union, minimal union, intersection, retained-variable cofactors,
threshold, Cartesian dot product, size-six minimal product, node sharing,
transposition insertion and permutation product follow the literal recursive
definitions. This includes the source's nonstandard `Abc_ZddIntersect`
unmatched-top behavior rather than silently changing ABC semantics. Owned
subset/permutation results and strings replace mutable caller buffers and
stdout. Selection-sort transposition conversion and reverse reconstruction
preserve the source sequence exactly.

`AnalyzeExtraZDDGia` returns the complement-insensitive GIA diagnostic.
`EnumerateExtraCubeStatesZDD` returns the bounded permutation-ZDD cube
experiment. Fixed Go-only tests pin canonical IDs, set families, operation
results, counts, shared nodes, formatting, composition, atomic limits and GIA
analysis. Cube totals 1/10/64/385 and ZDD node totals 0/63/443/2018 exactly
match the source through depth three. An independent explicit-family campaign
passed 5,000 executions without native code or retained fixtures.

The Extra manifest now has 18 of 20 selected units validated; `Dsd` and `Misc`
remain. Inventory is now 353 validated, 41 partial and 1,363 unstarted files.
The full suite passes in 17.725s, the race suite in 94.361s, vet is clean, and
native-source, package-header, `testdata/` and tool-allowlist audits pass.

## Extra six-variable DSD checkpoint

`src/misc/extra/extraUtilDsd.c` is validated across its active class manager,
lookup, composition, statistics and divisor surfaces. The source's missing
runtime `dsdfuncs6.dat` dependency is gone. Its 595 class records are embedded
as the permanent, human-readable `test_vectors/extra_dsd/classes6.tsv`; manager
construction validates the schema, count, support sizes, unique normalized
representatives and permutation cycles.

`ExtraDSDManager` derives exact NPN configurations on demand using the source
adjacent-swap permutation schedule and Gray phase schedule. It preserves
representative output normalization, partial packed-permutation completion,
permuted operand truths, input phases, AND/XOR composition, literal encoding,
atomic cut rewriting, per-class counts, CNF costs and statistics. A concrete
manager replaces the process-global Read/Quit singleton. Structured reports
replace stdout; embedded validation makes the old `CanRead` file probe
unnecessary.

Fixed Go-only tests validate every class support and direct classification,
the schedule cycles, partial permutations, transformed truth replay, compute
and cut semantics, statistics, costs, divisor collection, limits and
cancellation. A 5,000-case campaign transforms arbitrary class
representatives and independently reconstructs every returned configuration.
No C program, native oracle or external binary dataset is used.

The Extra manifest now has 19 of 20 selected units validated; only `Misc`
remains. Inventory is now 354 validated, 41 partial and 1,362 unstarted files.
The full suite passes in 17.691s, the race suite in 93.829s, vet is clean, and
native-source, package-header, `testdata/` and tool-allowlist audits pass.

## Extra miscellaneous utility checkpoint

`src/misc/extra/extraUtilMisc.c` is validated across its complete active
combinatorial, truth-table, canonicalization, scheduling, parsing, diagnostic
and transpose surface. Checked concrete functions return owned arithmetic,
radix, permutation and formatting results. The recursive permutation order,
minterm direction, phase semantics and N/NN/P/NP/NPN minima are preserved.

`ExtraTruthNPhases` owns complete three- and four-input phase orbits.
`ExtraTruth4NPN` owns all 65,536 canonical representatives, phases,
permutation indices and class IDs and reports the exact 222 classes.
`ExtraExpandTruth` algorithmically replaces the source's large literal support
tables and table-print generator while preserving its reverse mapping and
singleton quirks. Owned slices replace `Extra_ArrayAlloc`; Go stable sorting
replaces the C comparison callback; print/test drivers return structured data
or are covered directly by tests.

Exact six-variable minimization traverses both output phases, all 720
adjacent-swap permutations and all 64 Gray phases. The heuristic retains the
source popcount choice and ten-round adjacent-pair descent. Strict hex parsing,
bit/power data, raw reversed-convention 32-by-32 transpose, packed rectangular
transpose, pointer-row input and matrix formatting cover the remaining
surface. Existing validated Go schedules, bit transforms and 64-by-64
transpose are reused instead of duplicated.

Fixed Go-only tests pin arithmetic/errors/ownership, exact permutation order,
sampled transform semantics, every 3N class, every 4NPN row by metadata replay,
all seven-input support masks, stable sorting, transformed exact minima,
parser/formatting quirks and every transpose bit. A standalone model fuzzer
passed 5,000 executions. No native oracle or generated corpus is retained.

All 20 selected Extra translation units and their `module.make` manifest are
now validated. Inventory is 356 validated, 40 partial and 1,361 unstarted
files. The full suite passes in 18.263s, the race suite in 94.263s, vet is
clean, and native-source, package-header, `testdata/` and tool-allowlist audits
pass.

## Header-only hash-map checkpoint

All five headers under `src/misc/hash` and their empty header-only manifest are
validated. `HashDefaultBin` preserves the source polynomial with explicit
32-bit wraparound. `HashStringBin` preserves its bytewise base-6 uint32 hash.
The concrete `HashIntMap`, `HashFloatMap`, `HashPointerMap`, `HashStringMap`
and `HashLongMap` retain tail insertion into collision chains, overwrite
without reordering, zero-valued conditional creation, size and deterministic
bin-then-chain traversal. Opaque C payloads use `uintptr`, not an interface or
unsafe pointer.

`HashGenericMap` preserves caller-supplied hash/comparison behavior through
concrete function types, including comparison-defined key equivalence and
retention of the first inserted key. Immutable strings replace `fFreeKey`.
Checked concrete cells replace native mutable data/entry pointers and become
invalid after removal or close. Entry snapshots are owned; limits, malformed
callback bins and lifecycle failures return errors.

Fixed Go-only tests pin extreme hash arithmetic, string hashes, collision
ordering, all typed operations, custom equivalence callbacks, removal,
ownership, limits and stale handles. An independent int/float state-machine
fuzzer passed 5,000 executions. No interface, native oracle, generated corpus
or randomized Go-map iteration is used.

Inventory is now 362 validated, 40 partial and 1,355 unstarted files. The full
suite passes in 18.269s, the race suite in 95.542s, vet is clean, and
native-source, package-header, `testdata/` and tool-allowlist audits pass.

## Memory-manager checkpoint

`src/misc/mem/mem.c`, `mem.h`, `mem2.h` and their one-unit manifest are
validated. The older byte managers share the already validated concrete
Extra/SAT arena core: fixed geometry, LIFO recycle, retained-first-chunk
restart, peak counts, flexible 4-KiB/twice-request growth, requested-byte
usage, stepped power-of-two classes, large-block retention and class-only
memory reporting remain intact. `ExtraFlexArena.Restart` fills the missing
surface and safely respects the retained chunk's actual size.

The new `MmrFlexArena`, `MmrFixedArena` and `MmrStepArena` preserve page-mask
handle encoding, offset one, source boundary conditions, 0xFF page fill,
fixed handle order, LIFO reuse, packed size tags and current/peak/total
statistics. Flex release intentionally frees the immediately preceding page,
matching the source. Numeric handles intentionally alias a reused allocation.
Checked read/write and owned snapshots replace borrowed word pointers. A
geometry unable to create the mandatory offset-one entry rejects explicitly
instead of falling through to an empty native free-list pop.

Fixed Go-only tests pin exact handles, page release, fill/reuse, stepped
routing, statistics, restart invalidation, ownership, limits, cancellation and
lifecycle. An independent fixed-handle state-machine fuzzer passed 5,000
executions. No native oracle or retained corpus is used.

Inventory is now 366 validated, 40 partial and 1,351 unstarted files. The full
suite passes in 18.305s, the race suite in 94.220s, vet is clean, and
native-source, package-header, `testdata/` and tool-allowlist audits pass.

## Ranked Boolean parser checkpoint

`src/misc/parse/parseCore.c` and `parse.h` are validated across their complete
formula grammar. `ParseRankedBooleanFormula` uses concrete typed stacks and an
owned `TruthTable` instead of CUDD nodes. It preserves constants, prefix and
suffix negation, explicit and implicit AND, XOR, OR, equivalence, both
implications, source precedence and left associativity. Brackets lower every
variable in their subexpression by one rank; bare variables address the
highest rank block.

Errors replace process-output diagnostics. The Go parser strictly rejects
mismatched delimiter kinds and rank underflow instead of retaining ambiguous
native behavior. Token, nesting, node and truth-word limits plus cancellation
bound untrusted input. Variable names and returned truth storage are owned;
lookup-only Go maps do not affect observable order.

Fixed Go-only tests exhaust every operator over all three-variable
assignments, ranked expressions over all 64 six-variable assignments,
precedence, implicit multiplication, duplicate-name behavior, malformed
syntax, ownership, limits and cancellation. An independent expression/model
campaign passes 5,000 executions over all ranks and operators. It uses no C
helper, external executable or generated permanent corpus.

The whole parse directory is now validated. Inventory is 368 validated, 40
partial and 1,349 unstarted files. The full suite passes in 18.406s, the race
suite in 96.117s, vet is clean, and native-source, package-header, `testdata/`
and tool-allowlist audits pass. `pkg/tools/` contains only the source-manifest
script.

## AVL-tree checkpoint

`src/misc/avl/avl.c`, `avl.h` and the API document are validated across their
complete surface. `AVLTree` is a concrete generic balanced tree using a
function comparator rather than an interface. It preserves equal-key right
insertion, first-match lookup and deletion, predecessor replacement, all four
rotation cases, first/last access, mutable find-or-add slots, counts,
bidirectional traversal, generators, callbacks, integer comparison and tree
checking.

Dynamic Go traversal removes the native 50-node path-stack ceiling. Checked
slots become stale after deletion or close. Generators own ordered snapshots:
later insertions remain excluded as documented, while deletion during
generation cannot leave a dangling native node pointer. Optional close
callbacks retain deterministic source postorder, and ordinary Go ownership
requires no cleanup callback. Entry limits, cancellation and errors replace
unchecked allocation and printed invariant diagnostics.

Fixed Go-only tests exercise every rotation shape, ordering, equal keys,
first-match deletion, mutable slots, forward/backward snapshots, mutation
during traversal, postorder close, invalidation, limits, cancellation and
deliberate corruption. An independent multiset state-machine campaign passes
5,000 executions and checks the full AVL invariant after every mutation. No C
helper, executable or generated corpus is used.

The AVL directory is now complete. After the later ledger correction described
below, inventory at this checkpoint is 369 validated, 40 partial and 1,348
unstarted files. The full suite passes in 18.519s, the race suite in
94.943s, vet is clean, and native-source, package-header, `testdata/` and
tool-allowlist audits pass.

## TIM inventory reconciliation

The already completed `misc/tim` implementation has now been reconciled with
its final two stale inventory entries. Every unit selected by `module.make` is
independently validated, and `tim.h` contributes no declaration outside the
owned manager, box, persistence, transform, configuration, timing, traversal
and diagnostic surfaces already covered by the concrete `misc_tim_*.go`
files. Go ownership replaces Stop/StopP and owned data/strings replace raw
pointers and stdout routines; no interface is introduced.

This changes no executable code or prior test evidence. After the later ledger
correction described below, inventory is 371 validated, 40 partial and 1,346
unstarted files. The preceding full, race, vet
and standalone-policy gates remain applicable to the unchanged Go package.

## Binary-BLIF checkpoint and ledger audit

`src/misc/bbl/bblif.c`, `bblif.h` and their one-unit manifest are validated.
`BblManager` owns design names, ordered objects, fanins and canonical SOP
functions. Construction preserves declared fanin order, source selection-sort
cube canonicalization, exact SOP deduplication, CI/CO rules and forward or
backward relative fanin links. Checked object views expose sequential IDs,
retained original IDs, types, marks, fanins, SOPs and function handles.

`MarshalBinary` and `ParseBblBinary` preserve the pinned three-vector stream:
little-endian signed fields, 12-byte object headers, packed flags, word-padded
NUL-terminated functions and signed relative fanin offsets. Parsing is bounded,
owned and exact-consumption only. Read objects receive sequential runtime IDs
while retaining their stored originals, safely implementing the API whose C
body aborts in `Bbl_ObjIdOriginal`. Owned bytes and strings replace file-path
mutation and borrowed pointers. Structured statistics and deterministic BLIF
text replace stdout/files. The half-adder demo returns a checked manager.

`BblTruthToSOP` and `BblSOPToTruth` preserve low-bit variable order,
32-bit stretching below five variables, minterm order, constants and the
source off-set complement convention through sixteen variables. Exact Go-only
tests pin the complete byte stream, read-renumber-write stabilization,
deduplication, forward links, text, ownership, limits, cancellation, every
truncation and truth tables through eight variables. Independent truth and
arbitrary-byte campaigns each pass 5,000 executions, with no native helper or
retained generated corpus.

The coverage ledger was also audited by source path. Five misplaced AVL/TIM
descriptions were moved onto their intended records. Three older false
promotions were corrected: `aigCanon.c`, `aigCuts.c` and `aigFact.c` remain
unstarted, while the completed DSD record belongs to `extraUtilDsd.c`. This
reduces the previously reported historical totals by two without changing
executable coverage.

Inventory is now 374 validated, 40 partial and 1,343 unstarted files. The full
suite passes in 18.518s, the race suite in 94.696s, vet is clean, and the
standalone policy audits pass.

## Packed extended-float and MVC-header checkpoint

`src/misc/util/utilFloat.h` and `utilDouble.h` are validated by the concrete
`ExtendedFloat32` and `ExtendedFloat64` types. Their exact packed exponent and
mantissa layouts, constants, ordering, IEEE conversion, addition,
multiplication, division by a power of two, overflow saturation and underflow
to one are represented without interfaces or C runtime support. The 64-bit
path preserves the source split-product arithmetic rather than silently using
a different rounded multiplication.

Fixed Go-only tests cover every named constant, bit layouts, conversions,
invalid values, saturation, underflow and the exact `1e100` packed value. An
independent integer and `big.Int` model drives a 5,000-execution fuzz campaign
across both widths. No C helper, generated native executable or retained
native-output corpus is used.

The final partial MVC inventory entry, `src/misc/mvc/mvc.h`, is also validated.
Every declaration and packed-cube macro backed by an implementation file in
this ABC snapshot maps to the existing concrete `misc_mvc_*.go` surface and
its tests. The header additionally names routines whose implementation files
are absent from this snapshot; those historical declarations are documented
as inventory-only instead of inventing behavior that is not available to
port. This changes no MVC executable code.

Inventory is now 377 validated, 39 partial and 1,341 unstarted files. The full
suite passes in 18.521s, the race suite in 93.824s, vet is clean, and the
native-source, package-header, `testdata/` and tool-allowlist audits pass.
`pkg/tools/` contains only `source_manifest.py`.

## Utility memory-recorder checkpoint

`src/misc/util/utilMem.c` and `utilMem.h` are validated by a concrete owned
`MemoryRecorder`. Stable caller-assigned numeric identities replace comparison
of unrelated raw pointers. Start, inactive pass-through recording, allocation
and free events, numeric sorting, duplicate multiset pairing, recycle windows,
quit, event limits and interrupt state are all explicit instance operations;
the source's writable process globals are not reproduced.

The source recycle routine is internally impossible: it asserts both recorder
vectors are null immediately before sorting and dereferencing them, and its
header spells the start/quit/recycle names differently from their definitions.
The Go path consistently names instance methods, rejects an unmatched free
atomically and returns an owned sorted list of unreleased identities. Returning
the list replaces the invalid idea of manually freeing an arbitrary Go value
while preserving the recorder's intended leak-detection result.

Fixed Go-only tests cover pass-through behavior, duplicates, ordering,
ownership, invalid-free retry, limits, repeated lifecycle calls and interrupt
state. An independent multiset state-machine fuzzer passes 5,000 executions.
No C helper, executable or generated native corpus is used.

Inventory is now 379 validated, 39 partial and 1,339 unstarted files. The full
suite passes in 18.513s, the race suite in 94.070s, vet is clean, and the
standalone-policy gates pass.

## Utility name-table checkpoint

`src/misc/util/utilNam.c` and `utilNam.h` are validated by the concrete
`StringNameTable`. It remains separate from the existing `misc/nm`
`NameManager`: the latter permits typed namesakes and deletion, whereas this
source is an append-only unique string interner. ID zero is reserved, and IDs
one onward retain first-insertion order across exact and bounded-prefix lookup
and insertion.

An insertion-ordered owned string slice is authoritative; a lookup-only map
replaces the private chained hash bins and is never enumerated. Cross-table ID
maps, duplicate-sensitive common counts, first-unique reporting, owned name
snapshots, logical storage statistics and deterministic source-shaped text are
implemented. Both exposed source hash functions preserve uint32 overflow,
the pinned signed-char arithmetic, 128-prime schedule and the different
bounded-NUL behavior of the original and Jenkins variants.

Go lifetime replaces Ref/Deref/Stop. Owned snapshots replace the mutable
borrowed scratch buffer, Render replaces stdout/file mutation, and callers
compose standard formatting with Intern instead of receiving an any-typed
variadic API. Explicit name/byte limits and errors replace unchecked growth
and assertions.

Fixed Go-only tests cover insertion, duplicates, prefix operations, lookup,
ID/name round trips, rendering, storage counts, ownership, maps, report order,
hash vectors, invalid input, limits and lifecycle. An independent ordered-map
model fuzzer passes 5,000 executions without native code or retained output.

Inventory is now 381 validated, 39 partial and 1,337 unstarted files. The full
suite passes in 18.057s, the race suite in 94.589s, vet is clean, and the
standalone-policy gates pass.

## Binary bridge protocol checkpoint

`src/misc/util/utilBridge.c` is validated by concrete owned `BridgePackage`
values and bounded graph/result codecs. Package marshaling preserves the exact
six-digit type, space, sixteen-digit size, space header followed by raw payload.
Text, abort, progress, bad-abstraction, ordinary netlist and abstract-netlist
messages are constructed without process-global files or flushing behavior.

Netlist serialization preserves the bridge const-one numbering, CI-then-gate
mapping, AIGER unsigned integers, ordinary AND fanins, latch drivers with
zero initialization, property count and complemented property outputs. The
incoming parser preserves its distinct grammar: gate wrapper bits, saved latch
drivers, property complementation, required version one, zero fairness and
zero constraints, then appended register inputs. Extended GIA buffers, XORs
and muxes are rejected because this protocol carries only ordinary ANDs.

The source's outgoing and incoming bodies are intentionally asymmetric: the
outgoing abstraction body stops after property literals, while the incoming
body requires the version/fairness/constraint trailer. Both forms are pinned
separately instead of claiming a false byte-for-byte self-round-trip. The
three-package reader returns the ignored envelopes as owned values while
parsing the middle netlist.

Holds, unknown and counterexample messages preserve status bytes, property
varints, depth, concrete frames, input values and the source's always-zero
initial-register vector. Typed errors and explicit byte/object/input/output
limits replace assertions, short-read diagnostics and `_exit`. Exact fixed
vectors, every truncation, malformed trailers, ownership, limits and all eight
assignments of a sequential fixture are tested. An arbitrary package/netlist
parser campaign passes 5,045 executions without native code or retained fuzz
output.

## Utility file disposition

`src/misc/util/utilFile.c` is validated through maintained Go replacements.
Its `tmpFile` is the implementation called by `utilSignal.c` and maps to the
already tested atomic `CreateSignalTempFile`, removing the native
create-close-unlink-reopen race while preserving prefix/suffix, exclusive
mode-0600 read/write files, ownership and errors. Permanent vectors and the
standalone temporary-file fuzzer cover this path.

The private platform clock only supported Windows temporary naming and needs
no separate surface. `vnsprintf` and `nsprintf` map to standard owned Go
formatting at call sites, avoiding a shared dummy file and an exported
any-typed variadic compatibility API. The commented demonstration has no
runtime behavior.

Inventory is now 383 validated, 39 partial and 1,335 unstarted files. The full
suite passes in 18.187s, the race suite in 95.263s, vet is clean, and the
standalone-policy gates pass.

## Utility ISOP/ESOP completion checkpoint

`src/misc/util/utilIsop.c` is validated jointly by the generalized recursive
`TruthTable.ISOP`/`ISOPInterval` implementation and the utility-specific cover
operations. `BuildCubeCoverTruth` preserves ordinary OR, exclusive XOR and
final-complement semantics. `TruthTable.ESOP` preserves the source's
cube-then-literal cost ordering and exclusive tie/drop precedence.
`GreedyISOPInterval` preserves required-on/permitted-on containment and the
source's whole-tail expansion comparison. `ISOPCNF` preserves the two-polarity
selector-tagged cover while explicitly bounding the selector representation.

All results own their tables and cubes. Explicit cube, element and visit limits
and cancellation replace fixed stacks and assertions. Deterministic formatting
replaces stdout, and no interface abstraction was added.

Go-only evidence covers all 256 three-variable ESOP functions, all 6,561
three-variable on/allowed intervals, every supported three-variable CNF
relation, exact pinned cube ordering, malformed inputs, bounds, limits,
cancellation and ownership. The existing 196 permanent recursive-ISOP vectors
remain under `test_vectors/`, and an independent property campaign passed 5,000
executions. No C helper, executable or runtime oracle is used.

`abc_namespaces.h` is also validated as a compile-time disposition: its only
contents are C/C++ namespace/linkage macros, replaced directly by package
`abc`. `util_hack.h` maps to ordinary Go `nil` plus the already validated
concrete CPU-time, soft-limit and typed out-of-memory operations; its mutable
process-global allocation callback is intentionally absent. `module.make`
selects eight utility translation units, all of which are now independently
validated in the coverage ledger.

Inventory is now 387 validated, 39 partial and 1,331 unstarted files. The full
suite passes in 18.468s, the race suite in 93.673s, vet is clean, and the
native-source, package-header, `testdata/`, symlink and tool-allowlist audits
pass. `pkg/tools/` contains only `source_manifest.py`.

## Integer pair/triple hash checkpoint

`src/misc/vec/vecHash.h` is validated by the concrete, arity-fixed
`IntTupleHash`. Pair managers intern `(Data0, Data1)` while retaining the first
`Data2` payload; triple managers intern all three values. Both preserve stable
one-based IDs in first-insertion order, exact uint32 hash overflow, source
prime-sized growth, collision-chain order, indexed entry access and bin
profiles. Pair payload increment returns the old value, decrement returns the
new value, and explicit assignment supports the source's reference-count and
algorithm-metadata uses.

Arity is fixed at construction because changing `Data2` in a triple table would
invalidate its key. All observed source call sites already use one arity per
manager. Go pointer sharing and garbage collection replace manual Ref/Deref;
`Close` supplies checked explicit lifecycle. Owned snapshots replace borrowed
object/slot pointers, while limits, cancellation, 32-bit value checks and an
atomic resize replace unchecked allocation and assertions.

The header's demonstration incorrectly expects a duplicate insertion to return
zero even though its implementation returns the existing positive ID. The Go
tests preserve the implementation and actual caller behavior. Fixed tests
exercise exact hash arithmetic, pair and triple identity, payload mutation,
forced resizing, profile accounting, ownership, limits, cancellation and
lifecycle. Independent insertion-ordered pair and triple models pass 5,000
fuzz executions without C code, a native executable or retained fuzz output.

Inventory is now 388 validated, 39 partial and 1,330 unstarted files. The full
suite passes in 18.542s, the race suite in 93.903s, vet is clean, and the
native-source, package-header, `testdata/`, symlink and tool-allowlist audits
pass. `pkg/tools/` contains only `source_manifest.py`.

## Managed attribute-vector checkpoint

`src/misc/vec/vecAtt.h` is validated by the generic concrete
`AttributeVector`. Its instantiated storage is `[]*T`; generic constraints are
compile-time only, so the port adds no runtime interface values or dispatch.
Typed function callbacks preserve the optional manager-free, lazy object-start
and object-free hooks.

All 15 source attribute tags, initial/default capacity, exact doubling-or-
index-plus-ten growth, direct writes, lazy construction, entry deletion,
clearing, explicit growth, manager access and retained/freed manager close
paths are represented. Snapshots own their slot slice instead of exposing the
container's mutable array while retaining pointed-to object identity. No source
call site directly consumes `Vec_AttArray` beyond its unused ABC wrapper.

`FreeEntry` deliberately preserves the source's unusual condition that its
free callback runs only when the manager is nonnil; Clear and Close pass the
manager even when it is nil. Checked indices, capacity limits, cancellation,
owned slices and typed errors replace unchecked allocation and assertions.

Fixed Go-only tests cover every tag, direct and lazy use, exact growth,
callback order, overwritten slots, nil managers, retained/freed manager
ownership, limits, cancellation, snapshots and closed state. An independent
slot-state model passes 5,000 fuzz executions without C code, a native
executable or retained fuzz output.

Inventory is now 389 validated, 39 partial and 1,329 unstarted files. The full
suite passes in 18.462s, the race suite in 93.387s, vet is clean, and the
native-source, package-header, `testdata/`, symlink and tool-allowlist audits
pass. `pkg/tools/` contains only `source_manifest.py`.

## Paged primary/shadow word-arena checkpoint

`src/misc/vec/vecSet.h` is validated by the concrete `PagedWordSet`. It
preserves power-of-two page geometry, two metadata words per page, primary and
shadow cursors, page/shift handle encoding, stable record handles, signed-int
packing, strict rollover, explicit entry counts, restart/reuse, byte fetches,
saved strings, shrinking, shadow-limit commit and source-shaped memory
accounting.

Owned word, int and byte snapshots plus checked writes replace borrowed
interior pointers. `CopyWords` snapshots before writing so proof compaction is
safe even when source and destination overlap. Fixed and callback-sized handle
enumeration replace the source iteration macro. Constructor/Close cover both
the embedded and heap allocation helpers because Go ownership has no separate
destruction rule for those cases.

Unwritten words and odd trailing halves are deterministically zero rather than
exposing uninitialized native bytes. A record that would fill a page through
its final word is rejected: the source assertion allows it, but the resulting
current handle aliases offset zero of the next page and is unsafe to shrink or
dereference. All representable source handles retain their exact encoding.
Page, entry and byte bounds, cancellation and typed errors replace unchecked
growth and assertions.

Fixed Go-only tests cover exact handles, signed packing, page rollover, logical
and source-ABI memory measures, bytes, embedded NUL strings, primary/shadow
shrinks, overlap-safe compaction, entry iteration, restart, existing-page
reuse, stale-shadow reset, empty entries, limits and lifecycle. An independent
append/select/compact/replay state model passes 5,000 fuzz executions without
C code, a native executable or retained fuzz output.

Inventory is now 390 validated, 39 partial and 1,328 unstarted files. The full
suite passes in 18.126s, the race suite in 93.475s, vet is clean, and the
native-source, package-header, `testdata/`, symlink and tool-allowlist audits
pass. `pkg/tools/` contains only `source_manifest.py`.

## Vector hash-interner checkpoint

`src/misc/vec/vecHsh.h` is validated by three concrete managers.
`SingleIntInterner` interns individual signed 32-bit values.
`FixedIntRecordInterner` owns fixed-width backing records while retaining the
first source index for each unique value. `IntVectorInterner` owns arbitrary-
length integer vectors. All preserve zero-based stable IDs, first-insertion
order, resize-before-lookup behavior and the source collision-chain rebuild
order.

The source's little-endian uint32 Jenkins hash, alternate weighted-byte hash,
variable-vector prime-sum hash and unusual prime schedule—including table size
one and skipping prime two—are explicit. Word records split into low/high
int32 halves. Empty vectors and the source's header-plus-even-padding storage
count are retained.

Owned backing data and entry snapshots replace borrowed vector slots and the
three mutable scratch headers, preventing callers from mutating an interned key
behind its hash chain. Batch int/word functions reproduce source ID maps, and
the source helper's assertion that a prebuilt word table is all-unique becomes
a typed error. Entry, bin and backing-data limits, cancellation, validation and
Close replace unchecked allocation and assertions.

Fixed Go-only tests pin every hash against independent byte/arithmetic models,
the source integer and fixed-record examples, word splitting, duplicate source
indices, natural variable vectors, empty vectors, padding, ownership, limits
and lifecycle. Independent insertion-ordered models for all three managers
pass 5,000 combined fuzz executions without C code, a native executable or
retained fuzz output.

Inventory is now 391 validated, 39 partial and 1,327 unstarted files. The full
suite passes in 18.613s, the race suite in 94.258s, vet is clean, and the
native-source, package-header, `testdata/`, symlink and tool-allowlist audits
pass. `pkg/tools/` contains only `source_manifest.py`.

## Typed nested-vector checkpoint

`src/misc/vec/vecVec.h` is validated across all 60 tracked definitions by the
concrete `NestedVector[T comparable]`. The generic constraint is compile-time
only; no runtime interface values or dispatch are introduced. Typed pointer
and integer instantiations consolidate the source's duplicated entry points,
with concrete int helpers for first-entry sorting and formatting.

Outer minimum-eight and exact expansion, inner minimum-sixteen/doubling growth,
empty levels, ordered and unique insertion, total counts, deep cloning,
capacity-preserving clearing and source-ABI memory accounting are explicit.
Owned level and traversal snapshots prevent callers from mutating container
storage. `Levels`, `Entries` and `ZipNestedVectorLevels` cover every source
forward/reverse, start/stop, pointer/int and two-vector iterator macro.

Equal-key sorts are stable and deterministic where native `qsort` tie order is
unspecified. Empty first-key levels return a typed error instead of reading out
of bounds, stdout formatting returns owned text, and Close consolidates the
three C destruction forms. Level, entry and storage bounds are checked before
allocation; cancellation and failing pushes are atomic.

Fixed Go-only tests cover capacity schedules, pointer identity, traversal,
ownership, sorting, formatting, memory, limits and lifecycle. An independent
nested-slice model and the Go fuzzer each pass 5,000 mixed-operation executions
without C code, a native executable or retained fuzz output.

Inventory is now 392 validated, 39 partial and 1,326 unstarted files. The full
suite passes in 18.709s, the race suite in 94.895s, vet is clean, and the
native-source, package-header, `testdata/`, symlink and tool-allowlist audits
pass. `pkg/tools/` contains only `source_manifest.py`.

## Embedded integer-level vector checkpoint

`src/misc/vec/vecWec.h` is validated across all 59 tracked definitions by the
concrete `IntLevelVector`. It remains distinct from the generic nested vector
because the source embeds integer-vector headers in every outer-capacity slot.
Ordinary/exact allocation, each outer and inner growth rule, dormant levels
across Init shrinkage, destructive Shrink, Clear-to-zero-size, appended and
inserted levels, unique insertion, active counts, deep cloning and native
memory accounting are represented.

Owned snapshots and integer level IDs replace borrowed headers, raw arrays and
pointer subtraction. Checked level ranges, selected-level traversal, adjacent
pairs and two-vector zipping replace all iterator macros. Size, first-entry and
last-entry sorts use deterministic stable ties; empty boundary-key levels are
rejected. Ordinary and literal printers return owned exact text.

Equivalence classes, nontrivial counts, first-entry collection and pointer-
vector conversion return owned values. Temporary marks are explicit Booleans
instead of stolen capacity bits; bulk mark/unmark preflights duplicates and
state so errors are atomic. Empty compaction preserves nonempty order and
releases abandoned storage. Limits cover outer levels, retained entries and
inner capacity, and cancellation is checked before mutation.

Fixed Go-only tests cover all source behaviors. An independent embedded-slot
model and the Go fuzzer each pass 5,000 mixed-operation executions without C
code, a native executable or retained fuzz output.

Inventory is now 393 validated, 39 partial and 1,325 unstarted files. The full
suite passes in 18.820s, the race suite in 96.043s, vet is clean, and the
native-source, package-header, `testdata/`, symlink and tool-allowlist audits
pass. `pkg/tools/` contains only `source_manifest.py`.
