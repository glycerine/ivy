package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestConstrainedEquivalenceIsNotUnconditional(t *testing.T) {
	source := aig.NewGraph()
	pi, q := source.NewInput(), source.NewInput()
	obj := source.AND(pi, q)
	frames := aig.NewGraph()
	x, y := frames.NewInput(), frames.NewInput()
	candidate := frames.AND(x, y)
	sat := Ssw_SatStart(0)
	defer Ssw_SatStop(sat)
	sat.pAig = frames
	p := &Ssw_Man_t{pAig: source, pFrames: frames, pPars: &Ssw_Pars_t{FConstrs: 1, NBTLimit: 1000}, pMSat: sat, iOutputLit: -1, nFrames: 1, pNodeToFrames: make([]aig.Edge, len(source.Nodes()))}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	Ssw_ObjSetFrame(p, pi, 0, x)
	Ssw_ObjSetFrame(p, obj, 0, candidate)
	aig.Aig_ObjSetRepr(source, obj, pi)
	if got := Ssw_NodesAreEquiv(p, x, candidate); got != 0 {
		t.Fatalf("unconstrained result=%d", got)
	}
	Ssw_NodesAreConstrained(p, x, frames.Const0())
	if got := Ssw_ManSweepNodeConstr(p, obj, 0, 0); got != 0 || Ssw_ObjFrame(p, obj, 0) != x {
		t.Fatalf("constrained merge result=%d mapping=%v", got, Ssw_ObjFrame(p, obj, 0))
	}
}
