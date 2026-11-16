# Structured-mapping SAT dependency review (in progress)

This is a required dependency of the open giaIf.c mapping-boundary audit,
not a completed map/if or BSAT package audit. Original audit remains2/131.
The user requires faithful/mechanical ports, not substitute algorithms.

## Plan

1. [x] Identify the substitute and add new failing source-contract regressions.
2. [x] Confirm those contracts by freshly compiling unmodified ifSat.c.
3. [x] Restore original XY/XYZ MUX CNFs, handle destruction, assumption remap,
   actual BSAT solve/reuse and paired model extraction; restore caller ownership.
4. [ ] Compare broader native SAT/UNSAT sequences, solver reuse, all K2..6 CNF
   shapes, shared/degenerate/permuted partitions and complete multiword truths.
5. [ ] Review search ordering, composition/verification helpers and original
   example entry points; repair remaining source-fidelity/ownership gaps.
6. [ ] Review compactness and integration with structured GIA reconstruction,
   then run the mapping-family frozen full-module race/vet gate.

## Source and prior substitute

Original src/map/if/ifSat.c builds parameterized MUX CNFs, maps truth minterms
onto output assumptions, solves the same incremental SAT instance, and extracts
both LUT parameter tables from its model. The old Go BuildXY/XYZ returned
ifSatPlan{lutSize}, Unbuild did nothing, and CheckXY ignored the handle and used
cofactor classification. Even exact Boolean decisions do not make that a
faithful port of the original algorithm.

## Regression-before-repair evidence

New tests fail before production changes:

- Both factories return the placeholder rather than a SAT instance.
- Successful CheckXY clears scratch instead of retaining assumptions
  [17 19 21 23 25 27 29 30].
- Four source-valid zero/one-variable partitions are rejected.
- A single output pointer is modified, although C writes only when both exist.
- UNSAT overwrites caller outputs1234/5678 with zero.

/tmp/abcgo-if-sat-original-flow-red.log (handle64307 exit1) and
/tmp/abcgo-if-sat-unsat-output-red.log (handle38735 exit1).
Fresh original C verifies all contracts, including variable/clause counts
XY16/48 and XYZ28/128 atK2:
/tmp/abcgo-if-sat-native-original-flow.log (handle3378 exit0,1.909s).

## Restored implementation checkpoint

Add thin BSAT methods for its existing original sat_solver_add_mux and mux41
helpers. IF factories construct the same variables and clauses in source order.
CheckXY builds the same remapped minterm array and compact assumptions, calls
the ported BSAT solver with source unlimited resource arguments, retains scratch,
and extracts both model tables only on SAT with both output pointers present.
Keep the source six-input verification path. Nil-handle convenience uses an
owned temporary real SAT instance; a complete nil-handle partition search owns
one instance for reuse. It never calls a cofactor substitute.

Gia_ManFromIfLogic lazily creates one actual XY solver for its structured-DSD
branch and defers its destruction; it passes that handle to FindLut, as C does.
Existing/new SAT tests pass0.235s:
/tmp/abcgo-if-sat-first-restored.log (handle88636 exit0).
BSAT full race suite passes6.945s:
/tmp/abcgo-if-sat-bsAT-bridge-race.log (handle6484 exit0).
Broader native/model/lifecycle coverage and mapping-family integration remain
pending. Save coherent BSAT/IF/GIA checkpoints separately; do not infer family
or package completion from this initial core repair.
