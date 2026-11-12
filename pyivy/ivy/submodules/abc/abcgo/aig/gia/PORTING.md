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
  against the AIG owner's Edge/Graph representation. Compilation/integration
  tests await the remaining AIG package drafts; these are not yet verified as
  part of the core suite. Higher-level DAR/DCH driver wrappers remain pending.

Use `GOCACHE=/tmp/abcgo-cache abcgo/aig/gia/test_ported_core.sh -race` from the
repository root for the exact focused coverage. It includes the real vec, util,
mem, kit, Hop, CEC core and Satoko dependencies. This still does not claim that
all original GIA algorithms or the entire GIA package build.

The separate compactness pass removed the duplicated object-store slice header
and replaced redundant complement-cache pointers/Once with one atomic paired
handle. The bug pass added concurrency, malformed object import, varint bounds,
state-layout, source CEC witness ownership and early register metadata tests.
