# Sequential cleanup port and separate reviews

All eleven APIs defined by the original `aigScl.c` are implemented in
`aigScl.go` against `Graph` and `Edge`. There are no stub algorithms or disabled
source files. Inputs/outputs use the manager convention: true primary terminals
first, then paired register outputs/next-state inputs. State initialization is
zero, as in the source sweep algorithms.

| Source API | Go behavior and regression coverage |
| --- | --- |
| `Aig_ManRemap` | Simultaneous `[]Edge` CI replacement, including complemented inputs/constants; metadata and original flop representative numbers |
| `Aig_ManSeqCleanup_rec` | Combinational traversal with a next-state worklist at register boundaries |
| `Aig_ManSeqCleanup` | Sequential reachability fixed point; removes dead paired terminals; filters flop numbers and clock domains |
| `Aig_ManSeqCleanupBasic` | Retains register terminals, ties dead next-state functions to logical zero |
| `Aig_ManCountMergeRegs` | Original diagnostic counts and return value; writer-based test coverage |
| `Aig_ManReduceLachesCount` | Counts repeated identical next-state edges while distinguishing complement phases |
| `Aig_ManReduceLachesOnce` | Stable first-occurrence mapping of equal next-state functions |
| `Aig_ManReduceLaches` | Iterates until no further register equality is exposed by remapping |
| `Aig_ManComputeSccs` | Preserves the source's weakly connected register-dependency components and diagnostic behavior |
| `Aig_ManSclPart` | Actual partition extraction, per-part sweep, mapped-class transfer and representative rebuild; clock-domain and default paths tested |
| `Aig_ManScl` | Actual ternary or multivalued constant sweep plus register-equality reduction and original-manager class reconstruction |

`RegisterComponents` additionally exposes the source diagnostic's component
sets directly. The historical SCC name is misleading: the C routine traverses
both forward and reverse dependency edges, so it does not compute directed SCCs.
The port preserves its actual weak-connectivity behavior.

## Pass 1: compactness, idiomatic Go and readability

- Replaced intrusive temporary `pNext`, pointer casts, mark-bit hashing and
  flexible vectors with explicit edge maps, worklists and typed slices.
- Kept source API names, including the historical `Laches` spelling, to make
  correspondence with the original algorithms straightforward.
- Shared the central traversal-ID helper and isolated layout validation.
- Extracted writer-based diagnostics so tests do not replace process-global
  stdout or need pipe/goroutine machinery.
- Consolidated the earlier private ternary remap/cleanup implementation into
  the public sequential APIs. Optional simulator injection remains useful, but
  the default multivalued path calls the real `Aig_MvManSimulate` implementation;
  it no longer requires an injected engine to function.
- `ClockDomains` and enabled-empty `FlopReprs` storage are real manager metadata,
  deeply cloned by the duplication module. No C union fields or global scratch
  registration are required.

## Pass 2: bug review and fixes

- CI replacement is simultaneous; swapping two CI mappings cannot make both
  point to the second CI through an already-updated source copy binding.
- Layout, flop metadata and clock-domain indices are validated before cleanup
  changes terminal arrays. Invalid tracked classes are rejected before changing
  source copy scratch storage.
- Tracked zero-state classes reject constant-one, complemented-register and
  primary-input mappings, which cannot be represented faithfully by the source
  flop-pair metadata. Generic untracked remapping still supports both constants
  and arbitrary CI complement phases.
- Register cleanup reaches a fixed point through mutually dependent live
  registers and keeps true PIs/POs stable. Clock-domain register indices are
  remapped after removal.
- Basic cleanup respects a complemented explicit output edge when tying its
  logical function to zero.
- Cleanup APIs report removed AND/XOR gates, matching the C return contract,
  rather than counting reclaimed terminal/buffer storage.
- Default `SclPart` repairs the source caller's contradictory zero partition
  size by requesting one all-register partition. The partitioner's positive-size
  contract remains intact.
- Full sweep tests exposed a callee defect in representative duplication:
  initialized CI copies prevented applying representatives at gate fanouts.
  The representative owner fixed this with its own red-to-green regression;
  all four sweep paths now reduce the test graph from three registers to one.

## Validation and boundaries

Tests cover every source API, zero-register behavior, metadata preservation,
invalid arguments, cleanup idempotence, multi-iteration equality reduction,
complement phases and sequential traces from the zero state. Ternary, real
multivalued, explicit clock-domain partition and default-partition sweep paths
are all exercised. The combined converted-AIG focused race suite passes.

The full AIG package still contains unrelated unported files; the focused suite
is evidence for this implementation and its integrated dependencies, not a claim
that the complete AIG directory already builds. Multivalued frame arguments use
the actual engine's positive symbolic/saturation-frame preconditions. Timing
and history metadata use the duplication module's explicit cloning interfaces.
