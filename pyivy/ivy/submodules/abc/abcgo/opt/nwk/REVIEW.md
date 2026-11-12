# NWK port: algorithms, audits and integration gate

## Coverage

Every Go file in this directory now parses normally. The active C module's
families have native implementations: manager/object/adjacency, topology and
support traversals, MFFC references, timing and incremental propagation, AIG
conversion/strashing, retiming residual-flow search, GIA abstraction min-cut,
IF mapping conversion, BDC resynthesis, Shannon-cofactor speedup choices, LUT
merge candidate discovery and degree-prioritized graph matching, statistics,
BLIF output and support/duplicate-fanin reduction.

`nwk_.c` has no algorithms. `ntlnwk.h` declares the separately implemented NTL
package; this package does not pretend to implement NTL. `nwkFlow_depth.c` is an
alternative definition of the same flow symbols and is absent from the original
`module.make`; `nwkFlow.go` ports the active `nwkFlow.c` algorithm. The source
statistics code's commented-out NTL save/dump calls remain absent; best-network
parameter comparison itself is implemented with protected state.

## Readability and compactness pass

Completed separately from the bug pass. Replaced combined flexible-array fanio
storage, reallocating object pointers and duplicate capacity/count fields with
separate typed adjacency slices and stable object IDs. Functions use native
Hop/AIG/TIM/BDC APIs rather than pointer casts. Source-style public names and
integer predicates remain available. Scratch MFFC counts live separately from
adjacency. BLIF output, graph checking and function substitution have native
interfaces. GIA abstraction uses a dependency-neutral interface implemented by
GIA; there is no upward import cycle. The IF bridge uses explicit accessors,
not reflection or layout assumptions.

## Independent bug pass

Reviewed Boolean phases, all-zero support, duplicate edges, MFFC restoration,
choice classes, box boundaries, timing queue ordering and repeated operations.
Fixed and added regression coverage for:

- The original `Nwk_ObjFanout` inline hardcodes index 1. Native indexing honors
  the caller's index and works on one-fanout objects.
- Patching/deleting one duplicate adjacency removes exactly one reciprocal
  occurrence; source patching could remove all occurrences but decrement one.
- MFFC helpers restore temporary reference counts without resizing adjacency.
- Reverse timing traversal crosses CO-to-box-output-CI dependencies. CI levels
  include box-input dependencies; queue insertion accepts the source's `-1`
  initial cursor. Incremental timing propagates through actual TIM boxes.
- Local AIG strashing preserves explicit XOR functionality as well as ANDs and
  complemented edges. Constants and zero-support functions survive reduction.
- IF translation retains actual choice classes. Local cut reconstruction tries
  equivalent realizations and normalizes their phase, rather than assuming the
  cut leaves belong to the representative's structural cone.
- IF power mapping computes real SAIG probabilities and reindexes float-bit
  storage onto IF object IDs. Simulation temporarily treats ROs as free inputs
  while preserving the caller's original register metadata afterward.
- GIA abstraction refinement obtains actual abstraction PPIs and updates removed
  gate classes; ordinary outputs are not substituted for abstraction PPIs.
- Deleting terminals through the node-deletion API fails before mutation.
- Flow verification returns an actual validity result; the original backward
  verification routine unconditionally returned success.

Library ownership follows the source: the network owns its Hop/TIM managers
and borrows its LUT library. Speedup restores borrowed timing/library state.
Support minimization caps truth expansion at twenty inputs (source's unsigned
support masks and fixed work allocations cannot represent arbitrary support).
BDC preserves the source's fifteen-input network resynthesis limit.

## Tests and current gate

`test_native.sh -race -cover` passes, including vet: 59.6% coverage, 2.917s on
the recorded run. This focused integration copies unchanged actual native
sources to a temporary module. It uses the complete native IF library plus its
exact source type declaration, the actual SAIG simulation family and the full
actual supporting packages. It excludes only `nwkMap.go`, whose real IF mapper
currently imports unfinished GIA families. It does not modify or hide product
files, and it is explicitly not the ordinary package completion gate.

Native regressions cover graph/interface checks, DFS/support/MFFC, actual
Hop/AIG round-trip truth, support and duplicate reduction, BDC truth preservation,
unit and variable-pin timing, TIM box propagation, speedup truth preservation,
retiming, matching, and BLIF output.

`test_vectors/reference.c` compiles and runs unchanged original active NWK
Obj/Fanio/Dfs/Flow/Timing/Merge/Util C files plus original AIG memory/TIM code.
Sixty-four deterministic differential cases compare exact forward/backward
retiming cut IDs, exact matching pairs and timing/level results. Only host
logging services are supplied locally. The oracle is skipped only if `cc` is
unavailable; all generated artifacts use temporary directories.

**Ordinary package test/race/vet are not yet green.** The latest normal
`go test ./opt/nwk` reaches external parser failures in `aig/saig/saigRetStep.go` and `aig/saig/saigStrSim.go`
and `aig/gia/giaCTas.go` (the latter through `map/if`). These are owned by other
workers. Once those packages build, run the normal NWK gates, type-check and
exercise the complete IF mapping driver, then update this record. The IF bridge
has not yet been validated by the ordinary compiler end to end.
