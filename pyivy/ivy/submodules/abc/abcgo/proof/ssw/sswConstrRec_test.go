package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestConstrainedRecUnrollsLatchDriver(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(pi.Not())
	aig.Aig_ManSetRegNum(g, 1)
	frames := aig.NewGraph()
	mapped := frames.NewInput()
	p := &Ssw_Man_t{pAig: g, pFrames: frames, nFrames: 2, pNodeToFrames: make([]aig.Edge, len(g.Nodes())*2)}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	Ssw_ObjSetFrame(p, pi, 0, mapped)
	if got := Ssw_ManSweepBmcConstr_rec(p, lo, 1); got != mapped.Not() {
		t.Fatalf("got %v", got)
	}
	if got := Ssw_FramesWithClasses_rec(p, lo, 1); got != mapped.Not() {
		t.Fatalf("cached got %v", got)
	}
}
