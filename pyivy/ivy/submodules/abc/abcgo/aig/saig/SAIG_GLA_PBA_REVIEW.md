# SAIG GLA/PBA review

## Scope and behavior

The Gla2 and Gla3 entry points now share one bounded sequential-AIG unroller. It
creates one SAT variable for each reachable graph object and frame, encodes AND
nodes, latch initialization/transitions, output objects, and the OR of the
selected property frames, then maps proof root clause IDs back to the graph
objects constrained by those clauses. Both entry points use `bsat.Solver2`,
whose proof core returns stable one-based input clause IDs. Gla2 preserves the C
behavior of returning nil when clause insertion itself detects a root conflict;
Gla3 retains the conflict in its proof solver and asks it for a core.

The frame relation is `LO(0) = false` and `LO(f) = LI-driver(f-1)` with the
latch-input edge complement applied. The property clause spans frames
`nStart <= f < nFramesMax`, as in the C source. SAT, unknown/resource-limited,
invalid-input, and Gla2 root-conflict outcomes return nil.

## Independent bug pass

- Compared each CNF gate clause to `Aig_Gla2AddConst`, `Aig_Gla2AddBuffer`, and
  `Aig_Gla2AddNode` in C. A focused exhaustive truth-table test checks buffer
  complement and AND fanin-complement behavior for every input/output value.
- Checked sequential latch semantics with a one-register fixture whose next-state
  function is complemented constant zero/one. The selected frame's SAT result
  follows the transition edge polarity.
- Found and fixed a buffer-clause target polarity inversion during the review.
  This had made a direct constant-one property appear UNSAT.
- Found and fixed the distinction between a latch-input CO object and its driver
  edge. The former supplies the complement bit and the latter supplies the SAT
  variable; using the CO object as the driver looked up a nonexistent frame var.
- Found and fixed clause-map drift after a solver root conflict. `Solver2` stops
  accepting clauses after it becomes inconsistent, so later attempted clauses
  must not be added to the local clause-ID map. The manager now stops mapping
  clauses after the first conflict while keeping the conflicting clause ID.
- Verified proof collection retains selected logic nodes as well as the required
  constant node; PI/PO-only clauses are excluded from the returned bitmap.

## Compactness and readability pass

- The frame-variable map is a single dense slice indexed by object ID and frame,
  avoiding one allocation per graph object.
- The Gla2/Gla3 implementations share the encoding and core-to-object mapping,
  keeping the two backend entry points consistent.
- Collection sizes its result from the graph's object bound without copying all
  graph nodes.
- Clause-to-frame IDs remain recorded to preserve the source's mapping metadata;
  the optional per-frame usage report is not exposed because the C source leaves
  `vVec2Use` disabled.

## Verification

The package has unrelated parser errors in `saigInd.go`, so I verified this
scope with an explicit file set:

```sh
GOCACHE=/tmp/abc-go-cache go test ./aig/saig/saig.go ./aig/saig/saigGlaPba.go ./aig/saig/saigGlaPba2.go ./aig/saig/saigGlaPba_test.go
GOCACHE=/tmp/abc-go-cache go test -race ./aig/saig/saig.go ./aig/saig/saigGlaPba.go ./aig/saig/saigGlaPba2.go ./aig/saig/saigGlaPba_test.go
GOCACHE=/tmp/abc-go-cache go vet ./aig/saig/saig.go ./aig/saig/saigGlaPba.go ./aig/saig/saigGlaPba2.go ./aig/saig/saigGlaPba_test.go
```

All three focused gates pass. Full `go test ./aig/saig` remains blocked by
syntax errors in `saigInd.go`, outside this scope.
