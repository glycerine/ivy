# CUT port review

The package ports the cut manager, sorted support merges, bucketed cut lists,
dominance filtering, bounded enumeration, truth composition, choices, flow
recording/replay, sequential frontier operations, mapping heuristics, and cell
library construction/loading from src/opt/cut. The ABC frame-dependent entry
points accept a structural Network interface; no process-global ABC objects or
C memory pools are required. All original source-derived filenames remain.

## Compactness, readability, and idiomatic Go pass

Reviewed separately from correctness. Shared list/slice conversion, truth
composition, constant initialization, network inventory, and proof capture are
factored into small helpers. Manager-owned maps replace dense pointer arrays;
explicit slices replace flexible allocation records. No redundant copy of a
network graph is maintained beyond the manager's input snapshot. Cut proof
metadata is optional and immutable; oracle records retain values, not references
to recycled cuts. Explicit lifetime accounting preserves the original allocation
statistics and catches double recycling. C-style exported entry points remain
for consumer migration; normal Go methods expose leaves and truth buffers.

## Independent bug-review pass

Observed failing regressions before corrections:

* An ineligible global cut could delete an eligible existing cut before being
  rejected. Eligibility now precedes dominance mutation.
* Trivial cuts inherited the preceding node's simulation. Compute current-node
  simulation before allocating the trivial cut.
* Complemented choice representatives produced incorrect composed truth.
  Normalize child phase before applying input-edge phase.
* Buffer list sorting moved its own trivial cut away from the first position,
  breaking flow indices. Preserve source bucket insertion order.
* Choice unions were absent from the oracle record, and alternative input
  phases were lost. Records now retain each cut's producer, actual source nodes,
  input phases, operation, indices, and result phase.
* Constant-zero nodes retained input truth 0xaaaaaaaa, and the oracle skipped
  constant setup. Constants now explicitly initialize all truth words and
  simulation; replay traverses actual proof dependencies, including buffers.
* Crossbar truth used the cell canonical phase instead of its own crossbar
  phase. Exhaustive primitive/crossbar tests cover this correction.
* Library loading mishandled a final line without newline. Scanner-based input
  accepts it and reports malformed hex and I/O errors.

Validation: package tests and race tests pass; go vet passes. Independent
minterm evaluation checks randomly generated AND/XOR networks and every
computed cut. Additional tests cover choices/phases, ordinary and buffer oracle
replay, sequential frontier ownership, all 22 primitive cells, crossbar
construction, canonical lookup, dump/reload, lifecycle and cut accounting.

## Scope notes

The source's network-level sequential driver is commented out. This port exposes
an active bounded ten-round implementation plus convergence statistics; its
node-level frontier APIs preserve the source algorithm. The default full cell
precomputation is intentionally expensive, as in the source; tests exercise
parameterized smaller libraries. Package managers and the optional global cell
library are not concurrently mutable.

A final sequential-driver audit found that its new structural adapter treated
explicit XOR nodes as AND and treated buffers as binary nodes. The independent
sequential-network regression reproduced XOR truth 0x88888888 instead of
0x66666666. Dispatch now selects XOR and models a buffer as an idempotent AND
with identical shifted fanins/edge phases; complemented buffer truth is checked
as 0x99999999. Constant seed truth/simulation is initialized explicitly. Full
race tests and vet pass after this correction.
