package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// discoverUnfold2Candidates finds polarities contradicted by the property in
// one unconstrained frame, then checks the two induction classifications.
func discoverUnfold2Candidates(g *aig.Graph, frames, conflicts, propagations, verbose int32) (*vec.Vec_Vec_t, unfold2Classes, error) {
	var empty unfold2Classes
	if g == nil || frames != 1 || conflicts < 0 || propagations < 0 || Saig_ManPoNum(g) != 1 {
		return nil, empty, fmt.Errorf("saig: direct implication detection requires one frame and one property")
	}
	unrolled, maps, err := unrollFrames(g, 1)
	if err != nil {
		return nil, empty, err
	}
	property := mappedEdge(maps, aig.Aig_ObjChild0(g, aig.Aig_ManCo(g, 0)), 0)
	unrolled.AddOutput(property)
	formula, err := encodeAIG(unrolled)
	if err != nil {
		return nil, empty, err
	}
	literal := formula.lit(property)
	formula.solver.SetLimits(int64(conflicts), int64(propagations))
	candidates := vec.Vec_VecAlloc(1)
	if formula.solver.Solve(literal) != bsat.Satisfiable {
		return candidates, empty, nil
	}
	seen := make(map[aig.Edge]bool)
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		if seen[e] || (n.Kind != aig.Input && n.Kind != aig.And && n.Kind != aig.ExclusiveOr) {
			continue
		}
		mapped := mappedEdge(maps, e, 0)
		var candidate aig.Edge
		switch {
		case impliedBy(formula, literal, formula.lit(mapped), int64(conflicts), int64(propagations)):
			candidate = e.Not()
		case impliedBy(formula, literal, formula.lit(mapped.Not()), int64(conflicts), int64(propagations)):
			candidate = e
		default:
			continue
		}
		seen[e] = true
		if Saig_ObjIsLo(g, e) {
			seen[aig.Aig_ObjFanin0(g, Saig_ObjLoToLi(g, e))] = true
		}
		vec.Vec_VecPush(candidates, 0, candidate)
	}
	classes, err := filterUnfold2Candidates(g, candidates, conflicts, propagations, verbose)
	return candidates, classes, err
}
