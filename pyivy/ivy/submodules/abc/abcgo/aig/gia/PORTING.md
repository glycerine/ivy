# GIA port progress

The package is incomplete. Approximately 98,000 generated Go lines still contain
C expressions, pointer arithmetic, and unresolved dependencies. The whole package
must pass `go test ./aig/gia` before being considered ported; selected-file tests
below verify only the completed representation components.

Completed components:

- `object.go` retains `Gia_Obj_t` with the original backward-fanin and edge-bit
  encoding and ports its independent classification, CIO, Value, phase and fanin
  accessors. Corresponding generated definitions were removed from `gia.h.go`.
- `graph.go` supplies a pure Go AND/inverter core, structural hashing, constant
  reductions, input/output tracking, simulation and levels.
- `graph_objects.go` connects that core to original GIA object encoding, including
  explicit CO objects and CI/CO ordinal storage in `iDiff1`. Specialized real XOR
  nodes are explicitly rejected during import; companion MUX arrays and sequential
  manager state remain to be integrated.
- `graph_test.go` tests truth tables, structural sharing, complemented literals,
  levels, invalid inputs and original-object round trips.

Current focused verification:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test aig/gia/object.go aig/gia/graph.go aig/gia/graph_objects.go aig/gia/graph_test.go
```

Remaining integration work includes the other generated GIA algorithms, complete
specialized XOR/MUX and sequential operations, C pointer-expression conversion,
and dependencies on AIG/HOP/SAT/BDD packages. All original entry points must be
ported and the complete package verified before claiming completion. No files
are excluded with build tags or renamed to hide build errors.

Manager integration now lives in `manager_types.go` and `manager.go`: original
manager fields are retained, object pointers remain stable through slice growth,
and original start/stop, object append/lookup, CI/CO/AND/OR/XOR/MUX/majority APIs
are ported. Fanout, sweeper phase, built-in simulation and support callbacks
remain connected to their original functions. `manager_test.go` adds integration
truth tables and object-stability checks; the completed manager files and tests now pass against the actual vec/util
packages using an explicit file list. Full-package tests still fail on remaining
generated GIA syntax. The earlier focused object and graph tests pass too. The manager's dependent types use actual
vec/util aliases, not duplicate containers.


Completed additional algorithms:

- Shared vec API aliases and exported util counterexample fields connect the
  existing manager to the packages being ported by other workers.
- `truth.go` ports original 32-bit word/truth-table primitives using math/bits.
- `aiger_io.go`/`aiger_vec.go` port AIGER fixed-width and unsigned integer codecs
  using encoding/binary, with malformed uint32 encodings rejected.
- `fanout.go` ports dynamic circular-list fanout insertion/removal and ordinary
  static fanout maps, preserving the five-word record layout.
- `manager_auxiliary.go` ports append-time simulation, phase/sweeper updates and
  support union; tests validate these together with object creation.
- `bridge.go` adapts actual managers to/from util.BridgeNetlist with preserved
  PI/register-output and PO/register-input partitions.
- `gia.h.go` is syntactically valid and gofmt-formatted; other generated files
  still prevent building the whole package.

Separate review passes performed:

1. Compactness/readability: canonical object-store slice replaces a duplicated
   cached header; source API helpers stay small, stdlib codecs/bit operations
   replace hand-written C machinery, and containers share vec implementations.
   Legacy object-count/capacity fields remain while unported files reference them.
2. Bug review: regressions cover stale-store replacement, concurrent complemented
   handle identity, malformed terminal/fanin imports, uint32 varint overflow and
   overlong encodings. Focused core and manager suites pass with -race.

Complemented-pointer source APIs use stable paired Go handles to preserve
pointer identity without unsafe. The preferred storage remains integer Literal
edges. Callers of the compatibility APIs must retain the original C discipline
of unwrapping handles with Gia_Regular before direct object field access.

Recent completed integrations and review:

- `giaAgi.go` is fully Go: compact uint64 nodes, native slices, source-compatible
  accessors, append/conversion/support APIs. Size/capacity come from slice len/cap.
  Tests catch the source CO tag-mask ambiguity, absent-third-input handling,
  traversal-epoch rollover, register tags, and stable storage/alias ownership.
- `support.go` ports original traversal/support APIs with iterative walks and
  rollover-safe epochs. Special XOR/MUX support includes all structural fanins.
- `giaMem.go` shares misc/mem pools through thin original-name wrappers. Explicit
  constructors preserve GIA's distinct chunk policies, with no duplicate pools.
- `network.go` exposes the dependency-neutral CEC interface, including real XOR,
  MUX, buffer and output kinds. Its independent evaluator cross-checks SAT models.
- `sat.go` uses actual proof/cec/core + Satoko algorithms. Legacy SolveSat solves
  the conjunction of outputs (source fAndOuts), validates witnesses, records node
  marks and owned counterexample bits, and preserves undecided/cancellation.
  The Cec_ManVerifyTwo compatibility wrapper stores the witness only in its first
  manager, matching source ownership; the second manager's existing model stays
  intact. Sequential free-state SAT requires an explicit native option.
- `references.go` exports manager `Refs`, preserving increment/decrement return
  semantics. 373 exact C free expansions/calls were simplified. Live executable
  free/ABC_FREE tokens and obsolete .pRefs selectors are rejected by a lexical
  regression which ignores comments/string literals. Slice nil assignment drops
  a reference; it does not invalidate other aliases.
- `kit_manager.go` and network traversal methods allow actual kit and extra/CEC
  algorithms to work without dependency cycles. Real kit truth synthesis runs
  against the original GIA manager in the focused suite.
- `aig_conversion.go`/`aig_representatives.go` port original AIG interchange APIs
  against the AIG owner's Edge/Graph representation. The real AIG integration passes `test_aig_bridge.sh -race` after the
  AIG package became buildable (1.754s). Higher-level DAR/DCH driver wrappers remain pending.

Use `GOCACHE=/tmp/abcgo-cache abcgo/aig/gia/test_ported_core.sh -race` from the
repository root for the exact focused coverage. It includes the real vec, util,
mem, kit, Hop, CEC core and Satoko dependencies. This still does not claim that
all original GIA algorithms or the entire GIA package build.

The separate compactness pass removed the duplicated object-store slice header
and replaced redundant complement-cache pointers/Once with one atomic paired
handle. The bug pass added concurrency, malformed object import, varint bounds,
state-layout, source CEC witness ownership and early register metadata tests.


CEC coordination APIs now include masked object marks/direct phase setters,
source phase propagation, representative links and borrowed simulation inputs.
The original multiply-with-carry random sequence lives in `random.go`, with
synchronized process-global state and the original reset semantics (each half
of RandomW resets separately when requested). Public CEX save/read/clear/take
methods validate before replacing state and own their packed input bits.

Readability review used small native methods and explicit borrowed-storage
comments. Bug review checked bitfield truncation, random sequence recurrence,
range-limited bulk filling and model ownership after failed replacement/take.
`test_ported_core.sh -race` passes with these regressions (1.882s).

`choice_miter.go` ports the shared-input, interleaved-output choice miter and its
cone copier, using native manager slices and the existing hash interface. The
readability pass keeps the source Value scratch mapping explicit; the bug pass
checks output interleaving and source input alignment in a focused regression.
The core race suite including this addition passes (1.789s). Real XOR/MUX/buffer
nodes are copied by their semantics rather than misread as ordinary AND nodes.

`object_links.go` now ports the original literal/complement handle adapters,
PI/PO/RO/RI partition and register-pair mappings, real-MUX fanin accessors and
Value-copy helpers. Small native methods retain the original APIs; the review
removed C integer-as-boolean operations and unsigned/signed return mismatches.
Regression coverage checks register pairing, partition boundaries, complemented
handle roundtrips and absent MUX fanins. Scratch mark/value/phase clearing APIs
also use direct slice walks rather than generated iteration macros.

Original real-XOR, real-MUX and buffer construction now use validated literals
and explicit packed fields in `special_gates.go`. Readability review consolidated
fanin ordering and MUX control inversion. Bug review moves validation before
object allocation, preventing invalid fanins from partially mutating a manager.
An exhaustive three-input test checks swapped MUX branches/control polarity,
complemented XOR inputs and buffers against native evaluation. Phase-pattern
regressions also compare real XOR/MUX/CO phases with independent evaluation.

`levels.go` ports gate-level and forward/reverse manager level APIs. The separate
readability review uses native max and slice walks; bug review preserves the
source distinction between weighted forward XOR/MUX levels and unit reverse
edge levels. Dedicated tests cover that distinction and monotonic updates. The
core race suite passes with these tests (1.808s).

AIGER and truth owners completed their files and independent reviews. Their real
files/tests, including the C truth oracle and AIGER interoperability checks, now
run together with the manager core via `test_ported_core.sh -race` (2.574s before
the final level additions). The truth-specific script delegates to that shared
suite to prevent its file list diverging as manager APIs are ported.

`scratch_arrays.go` ports the original per-frame copy arrays, truth numbering
sentinels and timing-channel APIs. Readability review keeps native vector indices
explicit and reuses literal/handle adapters. Bug review checks frame stride,
fanin complement propagation and separate arrival/required/slack channels;
free-and-clear timing stop preserves aliases as expected in Go. Regression suite
passes with integrated truth/AIGER files under race detection (2.681s). The full
real AIG bridge integration also passes with those files together (2.685s).

`class_metadata.go` ports representative-class predicates, colors/visit flags,
proved/failed flags, pair undo and sibling/next helpers. Readability review uses
existing representative methods and explicit boolean-to-int conversion. Bug
review found that native SetNext originally rejected negative links, while the
actual CEC SAT sources deliberately mark temporary tails with -1. The setter
now permits that sentinel; regression covers tail/pair predicates with -1 as
well as color visit previous-value semantics, pair undo and constant classes.

`append_simplified.go` ports the original non-hashing constant-reducing append
family, retaining fGiaSimple's request to preserve redundant AND structure, and
CO driver patching. Reviews check reductions avoid allocation, majority logic
and patch polarity. `ternary.go` ports all original mark-based ternary simulation
helpers, with explicit initialization checks and complemented fanin predicates.
The exhaustive ternary AND regression exercises all 3x3 values and complement
combinations against the independent source Xsim truth rules.

Input-level transfer, weighted reverse levels, required levels and slacks are
also native in `levels.go`. Reviews retain the source's inclusion of dangling
AND cones and separate unit-edge reverse API; regression checks input delays,
weighted reverse paths and requirements. A public idempotent EnableMuxes method
allows external packages to construct real MUXes without exposing storage;
regression checks enabling twice preserves a previously constructed MUX.

The CEC metadata boundary now provides validated representative remapping,
identity/name copying and independent input/output name snapshots. Reviews cover
reordered/collapsed class IDs, independent class storage, names omitted by -1,
and failed-map validation leaving destination metadata unchanged. Core race
regressions pass (2.617s).

MF/Satoko tuning is being ported in importable `mf` and `satcost` leaves so command
packages need not import unfinished GIA drafts. MF native cut enumeration,
truth support minimization, flow/reference blending and depth-eight exact-area
refinement pass race tests, including LUT sizes 3–8 and exhaustive real XOR/MUX
cut truth checks. The source CNF derivation owner supplies ordered clauses and an
original-ABC oracle; source coarsening remains an integration dependency.

The Satoko runner uses the original clause-add/configure/simplify/solve sequence
and reads the per-solve conflict counter. An independent original-C audit first
found source-lifecycle/search differences; the Satoko owner fixed the actual
polarity, watcher, root-trail, heap and simplification behavior. All 100 seeded
original-C status/conflict-cost comparisons now pass, without counter offsets.
The reproducible audit command is
`SATCOST_C_DIFFERENTIAL=1 GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./aig/gia/satcost -run TestOriginalCConflictCost -count=1`.
Arbitrary restart/reduction/incremental trace parity remains unproven; see the
solver's review record. The backend is explicitly injectable and identified.

The complete MF8 generation pipeline is now integrated. Its original-ABC audit
matches variable counts and every clause/literal ordering on 34 complete DAGs,
including reconvergence and greater-than-eight-leaf paths. Separate wide-cut C
cover audits match clause order for both function phases. Native parent
`giaSatoko.go` now provides Init/Create/CallOne/CallOneResult, split/incremental,
DIMACS and mapped-cone drivers; these use actual MF8 CNF, not CEC gate CNF.
Core race tests including driver regressions pass (2.703s).
