package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestWindowOutlineCrossesLatch(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi, lo)
	po := g.AddOutput(gate)
	li := g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 1)
	nodes := Saig_ManWindowOutline(g, lo, 2)
	if vec.Vec_PtrSize(nodes) != 2 || vec.Vec_PtrEntry(nodes, 0).(aig.Edge) != lo || vec.Vec_PtrEntry(nodes, 1).(aig.Edge) != gate {
		t.Fatalf("nodes %v", vec.Vec_PtrArray(nodes))
	}
	if !aig.Aig_ObjIsTravIdCurrent(g, li) || !aig.Aig_ObjIsTravIdCurrent(g, lo) {
		t.Fatal("latch endpoints not marked")
	}
	pis := Saig_ManWindowCollectPis(g, nodes)
	if vec.Vec_PtrSize(pis) != 1 || vec.Vec_PtrEntry(pis, 0).(aig.Edge) != pi {
		t.Fatalf("inputs %v", vec.Vec_PtrArray(pis))
	}
	var pointers *vec.Vec_Ptr_t
	pos := Saig_ManWindowCollectPos(g, nodes, &pointers)
	if vec.Vec_PtrSize(pos) != 1 || vec.Vec_PtrEntry(pos, 0).(aig.Edge) != gate || vec.Vec_PtrEntry(pointers, 0).(aig.Edge) != po {
		t.Fatalf("outputs %v pointers %v", vec.Vec_PtrArray(pos), vec.Vec_PtrArray(pointers))
	}
}
func TestWindowOutlinePrimaryInputStopsTraversal(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	aig.Aig_ManSetRegNum(g, 0)
	nodes := Saig_ManWindowOutline(g, a, 3)
	if vec.Vec_PtrSize(nodes) != 1 || vec.Vec_PtrEntry(nodes, 0).(aig.Edge) != a {
		t.Fatalf("nodes %v", vec.Vec_PtrArray(nodes))
	}
}
