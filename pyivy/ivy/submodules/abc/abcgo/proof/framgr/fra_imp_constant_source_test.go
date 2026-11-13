package framgr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// fraImp.c checks coincident representatives without SAT: equal polarities
// are tautological, and opposite constant polarities are decided by the left.
func TestImpCheckCoincidentConstantRepresentatives(t *testing.T) {
	for _, tc := range []struct {
		name                           string
		leftNeg, rightNeg, wantRemoved bool
	}{
		{"one_implies_one", false, false, false},
		{"zero_implies_zero", true, true, false},
		{"zero_implies_one", true, false, false},
		{"one_implies_zero", false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			reduced := aig.NewGraph()
			p := &Fra_Man_t{PManAig: g, PManFraig: reduced, PPars: &Fra_Par_t{}, NFramesAll: 1, PCla: Fra_ClassesStart(g), PMemFraig: make([]aig.Edge, len(g.Nodes()))}
			left, right := reduced.Const1(), reduced.Const1()
			if tc.leftNeg {
				left = left.Not()
			}
			if tc.rightNeg {
				right = right.Not()
			}
			Fra_ObjSetFraig(p, a, 0, left)
			Fra_ObjSetFraig(p, b, 0, right)
			imps := vec.Vec_IntAlloc(1)
			vec.Vec_IntPush(imps, Fra_ImpCreate(int32(a.NodeID()), int32(b.NodeID())))
			if pos := Fra_ImpCheckForNode(p, imps, b, 0); pos != 1 {
				t.Fatalf("position %d", pos)
			}
			removed := vec.Vec_IntEntry(imps, 0) == 0
			if removed != tc.wantRemoved || (p.PCla.FRefinement != 0) != tc.wantRemoved || p.PSat != nil {
				t.Fatalf("removed=%v refinement=%d SAT=%v", removed, p.PCla.FRefinement, p.PSat)
			}
		})
	}
}
