package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestManagerConstructionAndEquivCount(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	pars := &Ssw_Pars_t{nFramesK: 2}
	p := Ssw_ManCreate(g, pars)
	if p.nFrames != 3 || p.nPatWords != 1 || p.iOutputLit != -1 {
		t.Fatal("bad manager dimensions")
	}
	for _, edge := range p.pNodeToFrames {
		if edge != aig.InvalidEdge {
			t.Fatal("initial map is not empty")
		}
	}
	aig.Aig_ObjSetRepr(g, lo, g.Const1())
	if Ssw_ManCountEquivs(p) != 1 {
		t.Fatal("constant representative omitted")
	}
	aig.Aig_ManFanoutStop(g)
}
