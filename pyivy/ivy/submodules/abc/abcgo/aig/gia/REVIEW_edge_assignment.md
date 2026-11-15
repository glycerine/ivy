# Mapped-edge timing and assignment review

Scope: all 27 entry points in `src/aig/gia/giaEdge.c`, implemented by
`edge_storage.go`, `edge_delay.go` and `edge_mapping.go`. These are the original
experimental mapped-LUT edge heuristics, not a replacement optimization method.

## Repairs backed by red regressions

- Reverse assignment inherited a four-input guard from a shared forward helper.
  C's reverse routine has no such restriction. Keep the generic fanin view
  unbounded; enforce four inputs only in forward evaluation/assignment and Edg
  mapping conversion. Test five/eight/sixteen-input reverse LUTs with both mapping
  representations, retaining explicit forward-limit cross-talk tests.
- Optimizer diagnostics lost field spacing and the number/direction of fixed
  critical fanins/fanouts. Restore the exact source reports after native and
  direct-helper red tests.
- Incremental window evaluation dereferenced a missing primary mapping when only
  the secondary mapping was present. Use the same mapping-selection predicate
  as the full evaluator. C contains this defect too; this is an explicit Go
  repair on a crashing/unsupported source path, verified independently rather
  than claimed identical to C.

## Native and independent coverage

Three fixtures compile the actual original `giaEdge.c`, including private
helpers. They require Linux, GCC and the original ABC objects; native tests skip
when those facilities are unavailable.

| Source group | Evidence |
| --- | --- |
| Slot count/add/remove/clean; adjacency; pair import/export; packing | 28 primitive cases, every one of 1,024 undirected edge subsets over five nonzero nodes, 81 mixed singleton/pair/triple packings; compare slots, pairs, counts, adjacency and capacity diagnostics. |
| Forward evaluation and assignment, reverse assignment | 64 graph/mapping cases over valid two-to-four-input mapped AND cones; 128 forward assignments and reevaluations, 64 reverse assignments, complete edge slots/private delay vectors; six additional wide reverse cases. |
| Secondary conversion, unit forward/reverse evaluation, critical collection, optimizer | 96 optimizer runs for limits zero/one/two, full state and verbose reports; direct helper count/direction checks. |
| Mapping swaps, full/incremental window timing | Eight native full-window and four native incremental-window cases; compare both restored records, temporary traversal order, slots and timing vectors. Independent secondary-only incremental tests cover the source defect repair. |
| Boxed forward timing | 36 native configurations across both mapping forms, one/two-edge modes, input arrivals and box delays; compare assignment, reevaluation, vectors and CI/CO arrival state. |

Independent complete-path enumeration on 32 additional graphs checks the
10/2 forward and 1/0 unit edge costs, per-node and output maxima, reverse
assignment values, critical-node membership, symmetry, legal mapped dependencies,
unique slots and endpoint capacities. This enumerates paths rather than calling
the port's timing recurrence. Direct numerical cases cover first-level skipping,
unique maxima, runner-up domination, two/three-way ties, disabled two-edge mode,
remaining capacity and saturated endpoints. Window tests verify both caller and
manager records after success and after an invalid-window panic. Repeated timing
tests poison old values and change assignments to check complete clearing.

The native context fixture initially used a Go-only CO-arrival getter; inspect
original `timInt.h` fields instead. That was a fixture compilation error, not a
production defect. Boxed native cases use integer delays/arrivals. Fractional
behavior of the separate timing package is outside this family review.

## Compactness and ownership

The original Edg unit evaluators reuse/fill their delay vectors. Restore this
pattern with one small reset helper, removing repeated replacements while
preserving complete zeroing and resizing. A warmed repeated-query benchmark
changes both directions from two allocations/224 bytes to zero allocations/bytes.
Local 100ms samples change forward 841.3 to 446.8 ns and reverse 678.4 to 343.2 ns;
these are observations, not timing assertions. No new persistent cache is added.

Keep the compact common timing walk for primary/secondary mappings and timing
boxes, borrowed bounded mapping views, shared slot helpers, and the small
critical-edge assignment helper. The source's algorithm, selection order,
capacity limits, backup/restore policy and bounded iteration loops remain intact.

## Source boundaries and preserved conventions

Edge slots use zero as the absent-neighbor sentinel; valid represented edge
endpoints are nonzero object IDs. Packing/import capacity overflow preserves C's
partial assignment and diagnostics, including potentially asymmetric retained
slots; symmetry checks apply to algorithm-produced valid assignments, not an
over-capacity imported list.

Primary flat mapping takes precedence when both representations are installed,
as in C. Window swaps affect secondary records; callers requesting evaluation of
those records must use secondary-only mapping. Full native window tests preserve
and verify the primary-precedence convention. Timing/edge side effects describe
the candidate mapping even after the window records are restored, also as in C.
Go retains its existing `defer` restoration on panic; C assertions abort instead.

Forward assignment is a heuristic, not an optimal-delay guarantee. Reverse
assignment requires no timing manager. Edg optimization accepts endpoint limits
zero through two and a nonempty mapped LUT graph; its source assumes that graph
when relating forward and reverse depths. Retain existing fail-fast Go validation
for malformed records and preconditions instead of inventing silent substitutes.

Final focused and module race/vet evidence is recorded in
`../../SEQUENTIAL_AUDIT.md`.
