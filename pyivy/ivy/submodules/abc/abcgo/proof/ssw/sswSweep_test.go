package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
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

func TestSatPatternUsesSolverPolarity(t *testing.T) {
	source := aig.NewGraph()
	obj := source.NewInput()
	frames := aig.NewGraph()
	frame := frames.NewInput()
	aig.Aig_ObjSetPhase(frames, frame, true)
	sat := Ssw_SatStart(0)
	sat.pAig = frames
	Ssw_CnfNodeAddToSolver(sat, frame)
	p := &Ssw_Man_t{pAig: source, pFrames: frames, pMSat: sat, pPars: &Ssw_Pars_t{fPolarFlip: 1}, nFrames: 1, pNodeToFrames: make([]aig.Edge, len(source.Nodes()))}
	Ssw_ObjSetFrame(p, obj, 0, frame)
	if sat.pSat.Solve(sswLit(sat, frame, false)) != bsat.Satisfiable {
		t.Fatal("expected SAT")
	}
	if Ssw_ManGetSatVarValue(p, obj, 0) != 1 {
		t.Fatal("pattern interpreted using parameter instead of SAT solver polarity")
	}
}

func TestPhaseFlippedSatPatternCollectorsUseActualValues(t *testing.T) {
	frames := aig.NewGraph()
	input := frames.NewInput()
	aig.Aig_ObjSetPhase(frames, input, true)
	sat := Ssw_SatStart(1)
	sat.pAig = frames
	Ssw_CnfNodeAddToSolver(sat, input)
	if sat.pSat.Solve(sswLit(sat, input, false)) != bsat.Satisfiable {
		t.Fatal("expected SAT")
	}
	p := &Ssw_Man_t{pFrames: frames, pMSat: sat, vSimInfo: vec.Vec_PtrAllocSimInfo(1, 1)}
	Ssw_SmlAddPatternDyn(p)
	if vec.Vec_PtrEntry(p.vSimInfo, 0).([]uint32)[0] != 1 {
		t.Fatal("dynamic pattern lost phase flip")
	}
	vec.Vec_PtrCleanSimInfo(p.vSimInfo, 0, 1)
	Ssw_SmlAddPattern(p, 0, 0)
	if vec.Vec_PtrEntry(p.vSimInfo, 0).([]uint32)[0] != 1 {
		t.Fatal("latch pattern lost phase flip")
	}
	if Ssw_CnfGetNodeValue(sat, input) != 1 {
		t.Fatal("CNF model value lost phase flip")
	}
}
