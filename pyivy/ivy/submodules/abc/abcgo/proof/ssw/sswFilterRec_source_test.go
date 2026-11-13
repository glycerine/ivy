package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// The recursive filter helper follows the previous-frame latch input and
// retains its complemented edge, then caches each constructed frame node.
func TestFilterRecursiveUnrollingAndCache(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	node := g.AND(pi, lo)
	g.AddOutput(node)
	g.AddOutput(node.Not())
	aig.Aig_ManSetRegNum(g, 1)
	frames := aig.NewGraph()
	x0, x1 := frames.NewInput(), frames.NewInput()
	p := &Ssw_Man_t{pAig: g, pFrames: frames, nFrames: 2, pNodeToFrames: make([]aig.Edge, len(g.Nodes())*2)}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	Ssw_ObjSetFrame(p, pi, 0, x0)
	Ssw_ObjSetFrame(p, pi, 1, x1)
	Ssw_ObjSetFrame(p, lo, 0, frames.Const0())
	if got := Ssw_ManSweepBmcFilter_rec(p, lo, 1); got != frames.Const1() {
		t.Fatalf("latch frame 1 = %v, want constant 1", got)
	}
	if got := Ssw_ManSweepBmcFilter_rec(p, node, 0); got != frames.Const0() {
		t.Fatalf("node frame 0 = %v, want constant 0", got)
	}
	if got := Ssw_ManSweepBmcFilter_rec(p, node, 1); got != x1 {
		t.Fatalf("node frame 1 = %v, want %v", got, x1)
	}
	count := len(frames.Nodes())
	if got := Ssw_ManSweepBmcFilter_rec(p, node, 1); got != x1 || len(frames.Nodes()) != count {
		t.Fatal("cached frame result was not reused")
	}
}

func TestFilterRecursiveUnrollingRejectsMissingInput(t *testing.T) {
	g := aig.NewGraph()
	pi := g.NewInput()
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 0)
	p := &Ssw_Man_t{pAig: g, pFrames: aig.NewGraph(), nFrames: 1, pNodeToFrames: make([]aig.Edge, len(g.Nodes()))}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	defer func() {
		if recover() == nil {
			t.Error("missing input mapping should panic as source asserts")
		}
	}()
	Ssw_ManSweepBmcFilter_rec(p, pi, 0)
}
