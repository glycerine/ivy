package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestImpCheckForNodeRefinesOppositeRepresentatives(t *testing.T) {
	original := aig.NewGraph()
	a, b, c := original.NewInput(), original.NewInput(), original.NewInput()
	reduced := aig.NewGraph()
	x := reduced.NewInput()
	p := &Fra_Man_t{PManAig: original, PManFraig: reduced, PPars: &Fra_Par_t{}, NFramesAll: 1, PCla: Fra_ClassesStart(original), PMemFraig: make([]aig.Edge, len(original.Nodes()))}
	Fra_ObjSetFraig(p, a, 0, x)
	Fra_ObjSetFraig(p, b, 0, x.Not())
	imps := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(imps, Fra_ImpCreate(int32(a.NodeID()), int32(b.NodeID())))
	later := Fra_ImpCreate(int32(b.NodeID()), int32(c.NodeID()))
	vec.Vec_IntPush(imps, later)
	if pos := Fra_ImpCheckForNode(p, imps, b, 0); pos != 1 {
		t.Fatalf("next position = %d", pos)
	}
	if vec.Vec_IntEntry(imps, 0) != 0 || vec.Vec_IntEntry(imps, 1) != later || p.PCla.FRefinement != 1 {
		t.Fatalf("wrong refinement: %v", vec.Vec_IntArray(imps))
	}
}
