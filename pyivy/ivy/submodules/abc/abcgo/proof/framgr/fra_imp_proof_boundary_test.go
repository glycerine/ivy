package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// fraImp.c:510-562 skips zero entries, proves noncoincident mapped
// implications with SAT and stops at the first later-node implication.
func TestImpCheckProvesNoncoincidentNodesAndStops(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	later := g.AND(gate, b.Not())
	p := satTestManager(g)
	p.PCla = Fra_ClassesStart(g)
	Fra_ObjSetFraig(p, a, 0, a)
	Fra_ObjSetFraig(p, gate, 0, gate)
	imps := vec.Vec_IntAlloc(3)
	vec.Vec_IntPush(imps, 0)
	proof := Fra_ImpCreate(int32(gate.NodeID()), int32(a.NodeID()))
	vec.Vec_IntPush(imps, proof)
	future := Fra_ImpCreate(int32(a.NodeID()), int32(later.NodeID()))
	vec.Vec_IntPush(imps, future)
	if pos := Fra_ImpCheckForNode(p, imps, gate, 0); pos != 2 {
		t.Fatalf("position = %d, want 2", pos)
	}
	if vec.Vec_IntEntry(imps, 1) != proof || vec.Vec_IntEntry(imps, 2) != future || p.PCla.FRefinement != 0 || p.NSatProof != 1 {
		t.Fatalf("implications %v refinement=%d SAT proofs=%d", vec.Vec_IntArray(imps), p.PCla.FRefinement, p.NSatProof)
	}
}
