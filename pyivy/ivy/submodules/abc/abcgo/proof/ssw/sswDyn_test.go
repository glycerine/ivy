package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSswDynamicCollectConstraintSupport(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	g.AddOutput(gate)
	g.AddOutput(b)
	g.AddOutput(a)
	p := &Ssw_Man_t{pAig: aig.NewGraph(), pFrames: g, nSRMiterMaxId: aig.Aig_ManObjNumMax(g)}
	p.pAig.NewInput()
	p.pAig.NewInput()
	p.pAig.AddOutput(p.pAig.Const0())
	aig.Aig_ManSetRegNum(p.pAig, 1)
	leaves := vec.Vec_PtrAlloc(0)
	Ssw_ManCollectPis_rec(g, gate, leaves)
	if vec.Vec_PtrSize(leaves) != 2 {
		t.Fatalf("support: %v", vec.Vec_PtrArray(leaves))
	}
	Ssw_ManCollectPis_rec(g, gate, leaves)
	if vec.Vec_PtrSize(leaves) != 2 {
		t.Fatal("revisited support")
	}
	outputs := vec.Vec_IntAlloc(0)
	Ssw_ManCollectPos_rec(p, a, outputs)
	Ssw_ManCollectPos_rec(p, b, outputs)
	if vec.Vec_IntSize(outputs) != 2 || vec.Vec_IntEntry(outputs, 0) != 0 || vec.Vec_IntEntry(outputs, 1) != 0 {
		t.Fatalf("constraint pairs: %v", vec.Vec_IntArray(outputs))
	}
	if g.Node(g.Outputs()[2]).MarkB != true {
		t.Fatal("register output should be visited and excluded")
	}
}
func TestSswDynamicLabelPrimaryInputs(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	frames := aig.NewGraph()
	p := &Ssw_Man_t{pAig: g, pFrames: frames, nFrames: 2, pNodeToFrames: make([]aig.Edge, int(aig.Aig_ManObjNumMax(g))*2)}
	x, y := frames.NewInput(), frames.NewInput()
	Ssw_ObjSetFrame(p, pi, 0, x)
	Ssw_ObjSetFrame(p, pi, 1, y)
	Ssw_ManLabelPiNodes(p)
	for _, obj := range []aig.Edge{frames.Const1(), x, y} {
		n := frames.Node(obj)
		if !n.MarkA || !n.MarkB {
			t.Fatal("primary input marks missing")
		}
	}
}
