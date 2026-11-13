package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestClassAddCollectRemove(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	p := Ssw_ClassesStart(g)
	Ssw_ObjAddClass(p, a, []aig.Edge{a, b, c}, 3)
	if Ssw_ClassesClassNum(p) != 1 || Ssw_ClassesLitNum(p) != 2 {
		t.Fatal("wrong class statistics")
	}
	var size int32
	class := Ssw_ClassesReadClass(p, a, &size)
	if size != 3 || class[1] != b {
		t.Fatal("wrong class contents")
	}
	collected := vec.Vec_PtrAlloc(0)
	Ssw_ClassesCollectClass(p, a, collected)
	if vec.Vec_PtrSize(collected) != 2 || vec.Vec_PtrEntry(collected, 0) != b || vec.Vec_PtrEntry(collected, 1) != c {
		t.Fatal("wrong collected class")
	}
	if got := Ssw_ObjRemoveClass(p, a); len(got) != 3 || Ssw_ClassesClassNum(p) != 0 || Ssw_ClassesLitNum(p) != 0 {
		t.Fatal("class removal failed")
	}
	size = 99
	if Ssw_ClassesReadClass(p, a, &size) != nil || size != 99 {
		t.Fatal("absent class changed size")
	}
}

func TestRemoveCandidateFromClassAndConstant(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	p := Ssw_ClassesStart(g)
	Ssw_ObjAddClass(p, a, []aig.Edge{a, b, c}, 3)
	aig.Aig_ObjSetRepr(g, b, a)
	aig.Aig_ObjSetRepr(g, c, a)
	Ssw_ObjSetConst1Cand(g, d)
	p.nCands1 = 1
	Ssw_ClassesRemoveNode(p, b)
	if p.nClasses != 1 || p.nLits != 1 || len(p.pId2Class[a.NodeID()]) != 2 || p.pId2Class[a.NodeID()][1] != c {
		t.Fatal("node was not removed")
	}
	Ssw_ClassesRemoveNode(p, c)
	Ssw_ClassesRemoveNode(p, d)
	if p.nClasses != 0 || p.nLits != 0 || p.nCands1 != 0 || aig.Aig_ObjRepr(g, d) != aig.InvalidEdge {
		t.Fatal("last candidate not removed")
	}
}
