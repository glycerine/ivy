# Structured-mapping SAT dependency review (in progress)

This is a required dependency of the open giaIf.c mapping-boundary audit,
not a completed map/if or BSAT package audit. Original audit remains2/131.
The user requires faithful/mechanical ports, not substitute algorithms.

## Plan

1. [x] Identify the substitute and add new failing source-contract regressions.
2. [x] Confirm those contracts by freshly compiling unmodified ifSat.c.
3. [x] Restore original XY/XYZ MUX CNFs, handle destruction, assumption remap,
   actual BSAT solve/reuse and paired model extraction; restore caller ownership.
4. [x] Compare broader native SAT/UNSAT sequences, solver reuse, all K2..6 CNF
   shapes, shared/degenerate/permuted partitions and complete multiword truths.
5. [x] Review search ordering, composition/verification helpers and original
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

## Native reuse and complete-model matrix

The new fresh-C differential matrix passes all420 checks (385 SAT,35 UNSAT),
covering XY/XYZ variable and clause counts for K2..6, shared/empty/permuted
partitions, multiword inputs up to11 variables, single/paired output pointers,
and repeated SAT/UNSAT queries on the same solver. Compare every retained
assumption and both model words with original C. Independently classify each
truth table and verify complete model composition on every assignment.

/tmp/abcgo-if-sat-native-matrix-first.log terminates PASS,2.413s (test2.38s).
This establishes core source equivalence for the exercised cases; search order,
example stubs and the wider mapping integration gate remain open.

## Search ordering and source examples

The source's Test2/Test3 entry points were empty Go stubs. A new native-output
regression fails before repair: Go prints neither the derived LUT tables nor
the selected shared partition. /tmp/abcgo-if-sat-examples-red.log,
handle66570 exit1. Port the original algorithms, prints and cleanup; the direct
C comparison then passes1.861s (handle49598 exit0). map/if race passes9.328s
(handle84991 exit0; /tmp/abcgo-if-sat-examples-full-race.log).

Expanded first-partition and complete-assumption comparisons expose an exact
source quirk in K6/seven-input searches: the final shared-index loop starts at
s[1]+1, not s[2]+1. The generic Go combination search omitted repeated smaller
shared sets, changing final scratch and incremental solver state after UNSAT.
The new105-search matrix fails at case61 before repair (handle79570 exit1;
/tmp/abcgo-if-sat-search-native-first.log). Preserve this unusual source bound
explicitly in the compact recursive enumerator. All105 searches,420 CNF/model
checks and both examples then match original C exactly9.263s (handle95556
exit0; /tmp/abcgo-if-sat-search-compact-repaired.log). No output normalization
or source patch is used to conceal the boundary difference.

A further1,152 native helper comparisons pass before compactness changes:
640 arbitrary-input LUT compositions across arities0..4 and512 signed-table,
permuted/shared-variable complete compositions. /tmp/abcgo-if-sat-compose-before.log,
handle93156 exit0,1.897s. Review fixed scratch and original bitwise composition
next; structured mapping and the final integration gate remain open.

## Composition compactness checkpoint

Use the original bitwise minterm composition and reuse one fixed four-word
scratch array between bound/free LUTs. Read elementary six-variable truths from
the source constants instead of rebuilding them bit by bit. Remove two unused
helpers left by the old cofactor substitute. All native core/search/example/
composition comparisons pass11.140s (handle89682 exit0;
/tmp/abcgo-if-sat-all-native-compact.log).

The local300ms/count3 benchmark changes from48 B and2 allocations per composed
function to0 B and0 allocations. Sample medians are1,383ns before versus135.3ns
after. These numbers describe this small benchmark, not end-to-end mapping;
the after sample ran alongside native tests and is not a general performance
claim. Logs: /tmp/abcgo-if-sat-compose-before-bench.log (handle43821 exit0),
/tmp/abcgo-if-sat-compose-after-bench.log (handle90186 exit0).
Full map/if race passes8.932s (handle13193 exit0); IF/BSAT vet exits0 with no
findings (handle24853). The complete focused native race check passes19.880s (handle44308 exit0;
/tmp/abcgo-if-sat-all-native-compact-race.log). XYZ clause semantics beyond
construction counts remain to be exercised before closing this dependency.
