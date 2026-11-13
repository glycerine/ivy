package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestConstrainedSweepProvesAndMerges(t *testing.T) {
	source := aig.NewGraph()
	pi, q := source.NewInput(), source.NewInput()
	obj := source.AND(pi, q)
	frames := aig.NewGraph()
	a, b := frames.NewInput(), frames.NewInput()
	x := frames.AND(a, b)
	y := frames.AND(x, a)
	if x == y {
		t.Fatal("expected distinct framed nodes")
	}
	sat := Ssw_SatStart(0)
	sat.pAig = frames
	p := &Ssw_Man_t{pAig: source, pFrames: frames, pPars: &Ssw_Pars_t{}, pMSat: sat, iOutputLit: -1, nFrames: 1, pNodeToFrames: make([]aig.Edge, len(source.Nodes()))}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	Ssw_ObjSetFrame(p, pi, 0, x)
	Ssw_ObjSetFrame(p, obj, 0, y)
	aig.Aig_ObjSetRepr(source, obj, pi)
	if Ssw_ManSweepNodeConstr(p, obj, 0, 0) != 0 || Ssw_ObjFrame(p, obj, 0) != x {
		t.Fatal("proven node not merged")
	}
}
