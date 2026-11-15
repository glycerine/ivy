# Fast-extraction source review

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
8. [x] Complete native/independent retrieve, topology, insertion and driver
   matrices, including large (>6-fanin) factoring, typed/reordered interfaces,
   constants, scratch values, timing ownership and unchanged cases.
9. [x] Review compactness: reusable fanin index vector and source maintained
   levels instead of per-owner map allocation and repeated whole-graph scans.
   Measure before claiming performance improvement.
10. [x] Full source inventory reconciled; production/tests/fixtures frozen
    throughout uncached full-module race testing and GIA/opt-fx vet. Terminal
    results inspected and progress records updated; this family is closed.

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
| Gia_ManFactorGraph | fx_sop.go | 192 native/scalar factored covers |
| Gia_ManFactorNode | fx_sop.go | Native/scalar 3–8-variable factoring and insertion |
| Gia_ManComputeTruths | fx_retrieve.go | Native mapped/multiword matrices, fanin order checks |
| Gia_ManAssignNumbers | fx_retrieve.go | All object Value scratch compared in 192 native cases |
| Gia_ManFxRetrieve | fx_retrieve.go | Native/scalar covers, phases, constants, 2–9-variable cuts |
| Gia_ManFxTopoOrder_rec | fx_order.go | Native exact order/first/count, dependency holes |
| Gia_ManFxTopoOrder | fx_order.go | Native unchanged, new-owner, missing-owner, zero-CI cases |
| Gia_ManFxInsert | fx_insert.go | Native DSD/SOP insertion, owned interface, compact storage |
| Gia_ManPerformFx | fx_driver.go | Native drivers/reports, typed/sequential/no-AND ownership |

## Checkpoints

- `30228642e`: source graph/SOP leaf contract repairs and regressions.
- `d1f8c78ec`: incremental extraction engine and containment regressions.
- `572937416`: source divisor/queue diagnostics and verbosity flags.
- `bb758fdb5`: native backend/caller matrices and interface ownership.
- `8e903bc6c`: final focused source reconciliation and compact storage.

These checkpoints use `red:` because full family review and integration are pending.

## Completed focused reconciliation

All 37 focused tests pass without skips: GIA 16.837s, opt/fx 0.157s
(`/tmp/abcgo-fx-final-complete-focused.log`, exact handle18517 terminal0).
The source inventory above covers all 11 active giaFx.c routines.

- 192 mapped configurations compare retrieved cubes, phases and every object
  Value scratch; their 384 complete driver runs compare C structure counts
  and truth, plus 6,144 independent logical input assignments. Both fanin
  sorting directions, several mapping granularities and reordered CIs occur.
- Original C's unchanged branch calls Gia_ManDup, which loses reordered
  logical ports. The unmodified fixture execution characterizes its wrong
  8080 output mask. A separately requested `--correct-ports` execution uses
  existing C remap helpers only on this duplicate branch. Go retains correct
  logical order; its independent truth check never uses the faulty mask.
- 192 factored covers compare exact returned literal, object count and truth
  with C, followed by 192 insertion runs covering both DSD and >6-fanin SOP
  synthesis. Each stage checks 16,128 independent assignments. The native
  factoring fixture calls Abc_Start/Abc_Stop as required by C's global DEC
  manager; the initial missing initialization was a fixture error.
- Five exact native topology cases cover unchanged results/output pointers,
  grouped multiple cubes, new divisors, holes/missing owners, no CIs and empty
  input. Preserve the source's zero-valued input prefix, rather than treating
  it as an array of input IDs.
- 128 native multiword retrieval cases cover 2–9-input cuts and 16,320
  independent assignments. Two additional constant-function cases verify
  zero/one covers and phases. Constant examples use complete CI cuts rather
  than a constant fanin which C's internal cut collector rejects.
- 512 typed/interleaved/reordered/sequential driver cases check 8,192 input
  assignments, interface names/properties/comments/reset states and ownership
  transfers of timing/name vectors. Stopping the source leaves result-owned
  metadata valid. No-AND cases verify that mapping is unnecessary, including
  zero-input constant outputs and retained registers.
- All four verbosity modes match actual C after clock-only normalization;
  counters, float32 weights, divisor order and field widths remain exact.

## Compactness and ownership

Graph conversion reads existing DEC node payloads directly instead of copying
all leaf signals into a redundant array. FxInsert reuses the source-style
fanin index vector, fanin list and leaf vector, resets only touched indices,
uses six fixed elementary truth patterns, and builds SOP strings only above
six fanins. Destination levels are computed once for each appended node instead
of scanning every prior node for every DSD call. The DSD adapter still uses a
temporary AIG because the existing Go DSD implementation accepts that API;
this is an adapter to the original algorithm, not a new synthesis heuristic.

A single 150ms benchmark run before/after these insertion changes:

| Chain owners | Before time | After time | Before B/op | After B/op | Before allocs | After allocs |
| --- | --- | --- | --- | --- | --- | --- |
| 8 | 59.740us | 49.977us | 27,920 | 26,817 | 464 | 424 |
| 64 | 414.481us | 320.050us | 174,752 | 158,032 | 3,044 | 2,726 |
| 256 | 1,790.226us | 1,278.390us | 780,643 | 609,362 | 11,880 | 10,604 |

Logs: `/tmp/abcgo-fx-insertion-before-bench.log` and `...-after-bench.log`.
These are local short-run observations, not a general speedup guarantee.

Retained logical CI/CO/register order makes interface names, properties and
reset metadata valid; copy these into owned storage. Rebuilding does not copy
stale node-indexed mapping or serialized node payloads. The driver separately
uses the reviewed source timing-transfer operation, including moving names out
of the source. Unchanged duplication preserves actual typed gates; rebuilt
logic uses the original factoring/DSD algorithms.

## Final frozen integration

Exact full-module race handle40924 terminates exit0: 130 tested packages pass,
three command packages have no tests, and there are zero FAIL/DATA RACE records.
GIA 567.287s, ABCI 29.728s, BMC 24.133s, opt/fx 1.475s
(`/tmp/abcgo-fx-final-module-race.log`). Exact GIA/opt-fx vet handle49486
terminates exit0 with an empty log (`/tmp/abcgo-fx-final-vet.log`). Production,
tests and fixtures remained frozen at checkpoint8e903bc6c throughout both gates.

Close all 11 active giaFx.c routines and the required incremental extraction
backend reconciliation. The wider GIA source inventory remains open and
original package audit completion stays2/131. Next source family is giaGen.c
(13 routines).
