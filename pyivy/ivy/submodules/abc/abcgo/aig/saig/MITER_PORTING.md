# Miter family

`saigMiter.go` implements every active function of `src/aig/saig/saigMiter.c` over native `aig.Graph` and `aig.Edge` values. Miter creation shares primary inputs, orders independent latch outputs/inputs by source machine, and supports XOR or directed implication. Combinational miters share all inputs and compare all outputs. The specialized unrolling uses the bottom machine for frame zero and the top machine thereafter, exposing initial latch state as unconstrained inputs and retaining final next-state outputs.

Dual rails keep false/true rails in that order. The dual-rail AND helper uses explicit rail storage instead of borrowed object data/next pointers. Selected-root duplication and register-half extraction retain sequential cleanup and ordinary cleanup respectively. Demitering preserves source register-reachability marks and output-order decisions. Equivalent XOR encodings can yield a common inversion on the two extracted outputs; their XOR is unchanged.

Native extensions handle explicit XOR and buffer gates and topological DFS with non-topological IDs. Maps replace source scratch copy pointers. Source pointer-vector parameters for selected roots become `[]aig.Edge`; counterexample data becomes `[]int32`. Missing edges use `aig.InvalidEdge`, since zero is the valid constant-one edge.

`Ssw_SecSpecial` proves the original combinational unrolling obligation with the acyclic native `proof/cec/core` SAT implementation and the source 100000-conflict limit. It resimulates every returned counterexample and reports the failed frame/output. It does not reproduce FRA's rewriting/heuristic schedule, so bounded-resource undecided behavior may differ.

The general `Saig_ManDemiter` is unfinished in the original source: graph production asserts unconditionally. The port preserves this explicit failure when either output graph is requested, while retaining the diagnostic labeling traversal. It safely treats missing labeled registers/constants instead of dereferencing the source's possible null pointer. Complete simple/different/dual demiter routines remain available.

## Validation

Semantic tests exhaust sequential/combinational miter inputs, both operations, one to three frames, all dual-rail input codes, successful and rejected demitering, cleanup of unobserved registers, quick miter status, counterexample counting, and equivalent/different SAT queries. The C differential oracle links actual ABC objects and compares interfaces and complete truth tables for twelve original miter/unroll/demiter/dual variants. It permits only the common output inversion described above for demitering.

Focused gates (other SAIG drafts stay visible):

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./aig/saig/saig.go ./aig/saig/saigMiter.go ./aig/saig/saigMiter_test.go ./aig/saig/saigMiter_c_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./aig/saig/saig.go ./aig/saig/saigMiter.go ./aig/saig/saigMiter_test.go ./aig/saig/saigMiter_c_test.go
```

Whole-package integration remains with the SAIG owner. No source is excluded or hidden.

## Independent review corrections

A separate source/invariant review found and fixed two inherited source bugs.
`Saig_ManDemiterCheckPo` now uses either half's reachability mark to orient XOR
operands. A PI-only operand paired with a second-half register cone previously
fell back to object-ID ordering, sending that register into the wrong extracted
half and panicking on an unmapped fanin. The regression exhausts all eight
assignments of the extracted machines.

General diagnostic demitering now increments the traversal ID between finding a
labeled register and labeling the newly encountered cone. Previously the search
had already visited the root, making labeling return immediately for every
later output pair. A two-pair regression checks that both later cones acquire
the correct half labels. The unfinished graph-production path still fails
explicitly as documented above.

The independent compactness/readability pass found the shared graph copier,
explicit rail storage, and single selected-root duplication implementation keep
the port smaller than the repeated C scratch-pointer passes while retaining
interface policies. The independent bug pass also checked initial-state and
frame ordering, complement handling, constant adaptation in the SAT backend,
mark cleanup on complete-demiter success/failure, and output counting. Managers
and graph marks are local to an invocation/graph; simultaneous mutation of the
same graph remains outside the native graph API's concurrency contract. Focused
normal/race tests and vet include `saigMiter_review_test.go` in addition to the
files above.
