package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestSswFilterStartingState(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi.Not())
	aig.Aig_ManSetRegNum(g, 1)
	p := &Ssw_Man_t{pAig: g, pPatWords: make([]uint32, 1)}
	c := util.Abc_CexAlloc(1, 1, 3)
	c.IFrame = 2
	util.Abc_InfoSetBit(c.PData, 2)
	Ssw_ManFindStartingState(p, c)
	// At frame 2, the state is !PI(frame 1) = false.
	if util.Abc_InfoHasBit(p.pPatWords, 1) {
		t.Fatal("wrong starting state")
	}
	util.Abc_InfoXorBit(c.PData, 2)
	Ssw_ManFindStartingState(p, c)
	if !util.Abc_InfoHasBit(p.pPatWords, 1) {
		t.Fatal("counterexample state was not updated")
	}
}
func TestSswFilterGrowPreservesMappings(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	p := &Ssw_Man_t{pAig: g, nFrames: 2, pNodeToFrames: make([]aig.Edge, len(g.Nodes())*2)}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	Ssw_ObjSetFrame(p, x, 1, 5)
	sswFilterGrowFrames(p)
	if p.nFrames != 4 || Ssw_ObjFrame(p, x, 1) != 5 || Ssw_ObjFrame(p, x, 2) != aig.InvalidEdge {
		t.Fatal("frame growth corrupted mappings")
	}
}
func TestSswFilterBmcNoCandidates(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(g.AND(pi, lo))
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_ManCreate(g, &Ssw_Pars_t{nFramesK: 2})
	defer Ssw_ManStop(p)
	p.pMSat = Ssw_SatStart(0)
	if got := Ssw_ManSweepBmcFilter(p, 0); got != 1 {
		t.Fatalf("got %d", got)
	}
	Ssw_SatStop(p.pMSat)
	p.pMSat = nil
	Ssw_ManCleanup(p)
}

func TestSswSignalFilterRunsAndRestoresMarks(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	n := g.AND(pi, lo)
	g.AddOutput(n)
	g.AddOutput(n)
	aig.Aig_ManSetRegNum(g, 1)
	Ssw_SignalFilter(g, 2, 100, 1, 0, 0, nil, 0, 0)
	for _, obj := range g.LiveEdges() {
		if g.Node(obj).MarkB {
			t.Fatal("filter left simulation marks behind")
		}
	}
}

func TestSignalFilterDiscardsUnframedRepresentative(t *testing.T) {
	g := aig.NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	early, late := g.AND(a, q), g.AND(b, q)
	aig.Aig_ObjDisconnect(g, early)
	aig.Aig_ObjConnect(g, early, late, b)
	g.AddOutput(early)
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)
	classes := Ssw_ClassesStart(g)
	Ssw_ObjAddClass(classes, early, []aig.Edge{early, late}, 2)
	aig.Aig_ObjSetRepr(g, late, early)
	frames := aig.NewGraph()
	p := &Ssw_Man_t{pAig: g, pFrames: frames, pPars: &Ssw_Pars_t{}, ppClasses: classes, nFrames: 1, pNodeToFrames: make([]aig.Edge, len(g.Nodes()))}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	Ssw_ObjSetFrame(p, late, 0, frames.NewInput())
	if Ssw_ManSweepNodeFilter(p, late, 0) != 0 || classes.nClasses != 0 || aig.Aig_ObjRepr(g, late) != aig.InvalidEdge {
		t.Fatal("unframed representative was not discarded")
	}
}
