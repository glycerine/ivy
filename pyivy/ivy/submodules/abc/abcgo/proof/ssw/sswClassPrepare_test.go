package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestClassesFromRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	aig.Aig_ManReprStart(g, 5)
	aig.Aig_ObjSetRepr(g, b, a)
	aig.Aig_ObjSetRepr(g, c, a)
	aig.Aig_ObjSetRepr(g, d, g.Const1())
	p := Ssw_ClassesPrepareFromReprs(g)
	var size int32
	class := Ssw_ClassesReadClass(p, a, &size)
	if size != 3 || class[0] != a || class[1] != b || class[2] != c || p.nCands1 != 1 || p.nLits != 2 {
		t.Fatal("wrong reconstructed classes")
	}
}
func TestPrepareLatchCandidates(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi, lo)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_ClassesPrepareSimple(g, 1, 0)
	if p.nCands1 != 1 || aig.Aig_ObjRepr(g, lo) != g.Const1() || aig.Aig_ObjRepr(g, gate) != aig.InvalidEdge {
		t.Fatal("wrong latch candidates")
	}
}

func TestPrepareClassesFromPairs(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	constants := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(constants, 0)
	vec.Vec_IntPush(constants, int32(c.NodeID()))
	group := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(group, int32(a.NodeID()))
	vec.Vec_IntPush(group, int32(b.NodeID()))
	classes := make([]*vec.Vec_Int_t, len(g.Nodes()))
	classes[0] = constants
	classes[a.NodeID()] = group
	p := Ssw_ClassesPreparePairs(g, classes)
	if p.nCands1 != 1 || p.nClasses != 1 || aig.Aig_ObjRepr(g, b) != a || aig.Aig_ObjRepr(g, c) != g.Const1() {
		t.Fatal("wrong classes")
	}
}
