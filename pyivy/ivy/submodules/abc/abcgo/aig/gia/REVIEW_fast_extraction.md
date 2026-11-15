# Fast-extraction source review (in progress)

Scope: all 11 active routines in `src/aig/gia/giaFx.c`, plus the
`Fx_FastExtract` dependency in `src/base/abci/abcFx.c`. This is a source-family
review within the still-open GIA package audit; original completion is 2/131.
Work is sequential, without agents or branches.

## Plan and progress

1. [x] Map implementations: conversion in `fx_sop.go`; mapped truth/cover
   retrieval in `fx_retrieve.go`; topological order in `fx_order.go`;
   insertion and DSD adapter in `fx_insert.go`; driver in `fx_driver.go`.
2. [x] Add red regressions for conversion faults before repairs. Restore
   root-leaf fast path, ignore unused leaf payloads, remove redundant copied
   signals, accept extra unused leaves in direct SOP conversion.
3. [x] Characterize the documented substitute extraction algorithm against
   actual C. A seed-5 three-selection example produced a different cover and
   an extra divisor. Add both a minimal regression and native comparisons.
4. [x] Port original incremental matrix, literal incidence lists, stable cube
   IDs, insertion-ordered divisor IDs, float32 weights, queue and update order.
   Reuse the existing source-faithful indexed priority queue. The divisor map
   is only used for lookup; its iteration order cannot affect extraction.
5. [x] Preserve the existing Go API's redundant-input normalization around the
   C engine. Original C aborts on a captured general-containment input; extend
   the containment helper to recognize any entirely one-sided remainder.
   Guard recorded containment rows already deleted by a paired rewrite.
6. [x] Add red interface ownership regression, then preserve names, reset
   states, AIGER properties/comments in both insertion return paths.
7. [x] Reconcile source verbose reports and driver behavior. All four verbosity
   modes match C after clock-only normalization; timing field width is checked.
8. [ ] Complete native/independent retrieve, topology, insertion and driver
   matrices, including large (>6-fanin) factoring, typed/reordered interfaces,
   constants, scratch values, timing ownership and unchanged cases.
9. [ ] Review compactness: reusable fanin index vector and source maintained
   levels instead of per-owner map allocation and repeated whole-graph scans.
   Measure before claiming performance improvement.
10. [ ] Reconcile full source inventory, freeze code/tests/fixtures, run the
    uncached full-module race suite and vet, inspect terminal results, and
    update this record and `SEQUENTIAL_AUDIT.md` before closing this family.

## Evidence so far

- Graph leaf failures: `/tmp/abcgo-fx-leaf-red.log` and
  `/tmp/abcgo-fx-unused-leaf-red.log`; repaired graph matrix compares all 148
  exact C records and independently evaluates truth.
- Direct SOP trailing-leaf failure: `/tmp/abcgo-fx-sop-extra-red.log`.
  Permanent native/scalar coverage includes constants, output/leaf phases,
  multiple cubes, don't-cares, trailing leaves and NUL termination.
  Combined focused conversion suite passes in 3.692s
  (`/tmp/abcgo-fx-conversion-focused.log`).
- Backend divergence: `/tmp/abcgo-fx-extract-native-red.log` and
  `/tmp/abcgo-fx-queue-minimal-red.log`; original source returns one new
  divisor where the substitute returned two.
- The incremental port matches 7,680 actual-C extraction records across three
  support-generation modes, 64 seeds, five iteration limits, four literal
  limits and both canonical-only settings. Every case also checks all 64
  input assignments independently (491,520 assignment cases).
  `/tmp/abcgo-fx-extract-wide.log`, 3.674s, no skips.
- `opt/fx`'s 1,000 random-cover equivalence cases pass. Focused uncached race
  test passes in 1.473s; vet log is empty
  (`/tmp/abcgo-fx-incremental-race.log`, `...-vet.log`). These are dependency
  checks, not final module integration.
- General containment C child terminates with assertion failure at abcFx.c:631
  (`/tmp/abcgo-fx-contained-native.log`); core dumps disabled in the probe.
  New containment, tombstone and input-normalization red regressions precede
  their repairs. Input normalization was an existing Go API extension, not a
  replacement extraction heuristic. Irredundant inputs pass through unchanged.
- Interface test fails on both insertion paths before repair
  (`/tmp/abcgo-fx-interface-red.log`) and passes after repair
  (`/tmp/abcgo-fx-interface-repaired.log`). It independently compares logic and
  verifies that modifying the result does not mutate source metadata.

## Source inventory

| C routine | Go implementation | Review status |
| --- | --- | --- |
| Gia_ManGraphToAig | fx_sop.go | Native/scalar conversion matrix |
| Gia_ManSopToAig | fx_sop.go | Native/scalar conversion matrix |
| Gia_ManFactorGraph | fx_sop.go | Further factor-path matrix pending |
| Gia_ManFactorNode | fx_sop.go | Existing tests; native factor matrix pending |
| Gia_ManComputeTruths | fx_retrieve.go | Further mapped matrix pending |
| Gia_ManAssignNumbers | fx_retrieve.go | Scratch checks pending |
| Gia_ManFxRetrieve | fx_retrieve.go | Existing tests; native matrix pending |
| Gia_ManFxTopoOrder_rec | fx_order.go | Further topology matrix pending |
| Gia_ManFxTopoOrder | fx_order.go | Existing tests; native matrix pending |
| Gia_ManFxInsert | fx_insert.go | Interface repaired; synthesis matrix pending |
| Gia_ManPerformFx | fx_driver.go | Incremental backend restored; reports pending |

## Checkpoints

- `30228642e`: source graph/SOP leaf contract repairs and regressions.
- `d1f8c78ec`: incremental extraction engine and containment regressions.
- `572937416`: source divisor/queue diagnostics and verbosity flags.

Both commits use `red:` because full family review and integration are pending.
