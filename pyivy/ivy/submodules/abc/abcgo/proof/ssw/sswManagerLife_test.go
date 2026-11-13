package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestManagerCleanupAndStop(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_ManCreate(g, &Ssw_Pars_t{nFramesK: 1})
	p.pFrames = aig.NewGraph()
	p.nConstrTotal, p.nConstrReduced = 3, 2
	p.pNodeToFrames[0] = p.pFrames.Const1()
	Ssw_ManCleanup(p)
	if p.pFrames != nil || p.nConstrTotal != 0 || p.nConstrReduced != 0 || p.pNodeToFrames[0] != aig.InvalidEdge {
		t.Fatal("cleanup retained frame state")
	}
	Ssw_ManStop(p)
	defer aig.Aig_ManFanoutStop(g)
	if p.pAig != nil || p.pNodeToFrames != nil {
		t.Fatal("stop retained manager state")
	}
}
