package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// The source visits only the requested class representatives, leaving the
// other classes unchanged until their representative is passed to the helper.
func TestRefineGroupSelectsOnlyRequestedClasses(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	pairs := vec.Vec_IntAllocArrayCopy([]int32{int32(a.NodeID()), int32(b.NodeID()), int32(c.NodeID()), int32(d.NodeID())}, 4)
	p := Ssw_ClassesPreparePairsSimple(g, pairs)
	defer Ssw_ClassesStop(p)
	p.pFuncNodesAreEqual = func(_ any, x, y aig.Edge) int32 {
		if x == y {
			return 1
		}
		return 0
	}
	selected := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(selected, a)
	if got := Ssw_ClassesRefineGroup(p, selected, 0); got != 1 {
		t.Fatalf("refinements=%d", got)
	}
	if Ssw_ClassesClassNum(p) != 1 || aig.Aig_ObjRepr(g, b) != aig.InvalidEdge || aig.Aig_ObjRepr(g, d) != c {
		t.Fatal("incorrect retained classes")
	}
	vec.Vec_PtrClear(selected)
	vec.Vec_PtrPush(selected, c)
	if got := Ssw_ClassesRefineGroup(p, selected, 1); got != 1 || Ssw_ClassesClassNum(p) != 0 {
		t.Fatalf("second refinements=%d classes=%d", got, Ssw_ClassesClassNum(p))
	}
}
