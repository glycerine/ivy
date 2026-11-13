package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestFrameMapsAndComplementedFanins(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a.Not(), b)
	manager := &Ssw_Man_t{pAig: g, nFrames: 2, pNodeToFrames: make([]aig.Edge, 2*len(g.Nodes()))}
	for i := range manager.pNodeToFrames {
		manager.pNodeToFrames[i] = aig.InvalidEdge
	}
	frames := &Ssw_Frm_t{pAig: g, nObjs: int32(len(g.Nodes())), vAig2Frm: vec.Vec_PtrAlloc(0)}
	Ssw_ObjSetFrame(manager, a, 1, 0)
	Ssw_ObjSetFrame(manager, b, 1, a)
	Ssw_ObjSetFrame_(frames, a, 1, 0)
	Ssw_ObjSetFrame_(frames, b, 1, a)
	if Ssw_ObjFrame_(frames, b, 0) != aig.InvalidEdge {
		t.Fatal("missing mapping must not be constant one")
	}
	for i, fanin := range g.Node(gate).Fanins {
		mapped := aig.Edge(0)
		if fanin.NodeID() == b.NodeID() {
			mapped = a
		}
		if fanin.Complemented() {
			mapped = mapped.Not()
		}
		var got, got2 aig.Edge
		if i == 0 {
			got = Ssw_ObjChild0Fra(manager, gate, 1)
			got2 = Ssw_ObjChild0Fra_(frames, gate, 1)
		} else {
			got = Ssw_ObjChild1Fra(manager, gate, 1)
			got2 = Ssw_ObjChild1Fra_(frames, gate, 1)
		}
		if got != mapped || got2 != mapped {
			t.Fatalf("fanin %d: got %v %v want %v", i, got, got2, mapped)
		}
	}
	if Ssw_ObjChild0Fra(manager, a, 1) != aig.InvalidEdge {
		t.Fatal("input has no fanin")
	}
}
func TestSatNumbersAndConstantCandidates(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	p := &Ssw_Sat_t{vSatVars: vec.Vec_IntAlloc(0)}
	if Ssw_ObjSatNum(p, x) != 0 {
		t.Fatal("new SAT variable should be zero")
	}
	Ssw_ObjSetSatNum(p, x, 42)
	if Ssw_ObjSatNum(p, x) != 42 {
		t.Fatal("SAT variable not stored")
	}
	if Ssw_ObjIsConst1Cand(g, x) != 0 {
		t.Fatal("unexpected candidate")
	}
	Ssw_ObjSetConst1Cand(g, x)
	if Ssw_ObjIsConst1Cand(g, x) != 1 {
		t.Fatal("candidate not stored")
	}
}
