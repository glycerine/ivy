package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSweepNodeProvesEquivalentAndReusesRepresentative(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	left := g.AND(g.AND(a, b), c)
	right := g.AND(a, g.AND(b, c))
	aig.Aig_ManSetPhase(g)
	aig.Aig_ManReprStart(g, int32(len(g.Nodes())))
	aig.Aig_ObjSetRepr(g, right, left)
	sat := Ssw_SatStart(0)
	defer Ssw_SatStop(sat)
	sat.pAig = g
	p := &Ssw_Man_t{pAig: g, pFrames: g, pPars: &Ssw_Pars_t{nBTLimit: 1000}, nFrames: 1, pMSat: sat, pNodeToFrames: make([]aig.Edge, len(g.Nodes())), iOutputLit: -1}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	Ssw_ObjSetFrame(p, left, 0, left)
	Ssw_ObjSetFrame(p, right, 0, right)
	if got := Ssw_ManSweepNode(p, right, 0, 1, nil); got != 0 || Ssw_ObjFrame(p, right, 0) != left {
		t.Fatalf("equivalent sweep: result %d mapping %v", got, Ssw_ObjFrame(p, right, 0))
	}
}

func TestSweepBmcConstructsInitializedFrames(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi, lo)
	g.AddOutput(gate)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	sat := Ssw_SatStart(0)
	defer Ssw_SatStop(sat)
	p := &Ssw_Man_t{pAig: g, pPars: &Ssw_Pars_t{nFramesK: 1, nBTLimit: 1000}, nFrames: 2, pMSat: sat, pNodeToFrames: make([]aig.Edge, 2*len(g.Nodes())), iOutputLit: -1}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	if got := Ssw_ManSweepBmc(p); got != 0 {
		t.Fatalf("unexpected refinement %d", got)
	}
	if p.pMSat.pAig != p.pFrames || Ssw_ObjFrame(p, lo, 0) != p.pFrames.Const0() || Ssw_ObjFrame(p, pi, 0) == aig.InvalidEdge || Ssw_ObjFrame(p, gate, 0) != p.pFrames.Const0() {
		t.Fatal("initial frame construction incorrect")
	}
}

func TestUnassignedComplementedFrameUsesConsistentDefault(t *testing.T) {
	source := aig.NewGraph()
	a, b := source.NewInput(), source.NewInput()
	frames := aig.NewGraph()
	x := frames.NewInput()
	sat := Ssw_SatStart(0)
	sat.pAig = frames
	p := &Ssw_Man_t{pAig: source, pFrames: frames, pMSat: sat, pPars: &Ssw_Pars_t{}, nFrames: 1, pNodeToFrames: make([]aig.Edge, len(source.Nodes()))}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	Ssw_ObjSetFrame(p, a, 0, x)
	Ssw_ObjSetFrame(p, b, 0, x.Not())
	if Ssw_ManGetSatVarValue(p, a, 0) != 0 || Ssw_ManGetSatVarValue(p, b, 0) != 1 {
		t.Fatal("complementary unassigned frame edges received inconsistent defaults")
	}
}
