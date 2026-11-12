# Bidecomposition and circuit SAT review

## Compactness and readability pass

Bidecomposition uses the native bool/bdc engine. Signed vector truth words are
copied to []uint32; there is no unsafe cast. Cut cache offsets remain source
compatible, and the API returns owned truth slices. Native real gates are
handled explicitly rather than accidentally interpreting XOR/MUX as AND.

Circuit SAT is the actual source justification-frontier algorithm: propagation,
source ID/fanout decision heuristics, saved queue frontiers, reason expansion,
conflict resolution and recursive nonchronological return of learned clauses.
It does not route these APIs through the general CDCL solver. Queues keep logical
head/tail positions over slices and reasons retain signed object-ID offsets.

## Separate bug review

Bidecomposition explicitly copies constant one as literal 1 into BDC's integer
copy field. Storing a Go object pointer in its separate generic Copy field would
leave CopyInt zero and corrupt constant fanins. Constant sub-cones return zero
truth directly; a CI outside the declared cut is rejected. Visited truth cache
markers are restored after each query. Metadata survives final cleanup.

Circuit SAT handles constant queries and repeated/opposite root assumptions
before assignment, clears stale models before each call, and restores every
assigned object's marks/Value even after limits or UNSAT. Source complemented
pointer comparisons are represented with actual stable object handles; no
pointer ordering/arithmetic remains. Specialized real gates are explicitly
rejected by this ordinary-AND engine rather than misinterpreted.

## Evidence

- Nine-input truth queries including root complement and both constants match
  all 512 assignments and leave cache markers reset.
- Twenty seeded mapped resynthesis networks match every input assignment.
- Thirty random graphs, all three source decision heuristics, thirty single/pair
  queries per heuristic: 2,700 statuses match exhaustive assignments. SAT partial
  witnesses are independently replayed; object scratch is checked after calls.
- Queue snapshot restoration, justification exhaustion, opposite root assumptions
  and solver reuse have focused regressions.

These are actual selected-source checks. The full GIA package still contains
visible unported source families; no package-level completion is implied.

## CBS2 checkpoint before requested pause

The full giaCSat2.go family is now native, including the distinct integer-ID
assignment encoding, learned clause/cube records, two-watch lists, clause reason
expansion, forward-cone fanouts and both source propagation/frontier modes.
It does not alias the earlier CBS solver or a generic CDCL backend.
Readability pass uses slices and logical queue positions without cached capacity;
watched-list links are reacquired by record index rather than retained across
backing-store growth. Bug pass checks copied clause spans during reason expansion,
watch cleanup, repeated/opposite root handling, partial models and complete
assignment/reason/fanout restoration on solver reuse.

2,400 single/pair queries over 20 random graphs and both propagation modes match
exhaustive 32-assignment evaluation; partial witnesses independently replay.
The combined real-native AIG suite, now including CBS2, passes `-race` in 18.986s.
The earlier CBS source differential compared SAT/UNSAT/unknown counts on twelve
actual original ABC runs and passed. CBS2 has no original-C differential yet.

CCof is native syntax and uses the actual BSAT facade plus the Frames lifecycle
being ported independently. Its whole-family integration/race tests have not yet
run. It must not be described as validated or complete from parser progress.
No whole-package completion is claimed: the last ordinary GIA test stops in the
still-visible giaCTas.go draft (first parser errors 141,142,212,214,...).
