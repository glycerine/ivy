package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSweepBmcRefinesFalseInputLatchPair(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(pi, lo))
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetPhase(g)
	var pars Ssw_Pars_t
	Ssw_ManSetDefaultParams(&pars)
	p := Ssw_ManCreate(g, &pars)
	defer Ssw_ManStop(p)
	pairs := vec.Vec_IntAlloc(2)
	vec.Vec_IntPushTwo(pairs, int32(pi.NodeID()), int32(lo.NodeID()))
	classes := Ssw_TransformPairsIntoTempClasses(pairs, int32(len(g.Nodes())))
	p.ppClasses = Ssw_ClassesPreparePairs(g, classes)
	p.pSml = Ssw_SmlStart(g, 0, 1, 1)
	Ssw_ClassesSetData(p.ppClasses, p.pSml, func(d any, e aig.Edge) uint32 { return Ssw_SmlObjHashWord(d.(*Ssw_Sml_t), e) }, func(d any, e aig.Edge) int32 { return Ssw_SmlObjIsConstWord(d.(*Ssw_Sml_t), e) }, func(d any, a, b aig.Edge) int32 { return Ssw_SmlObjsAreEqualWord(d.(*Ssw_Sml_t), a, b) })
	p.pMSat = Ssw_SatStart(0)
	if got := Ssw_ManSweepBmc(p); got != 1 {
		t.Fatalf("expected SAT refinement, got %d", got)
	}
	if aig.Aig_ObjRepr(g, lo) == pi {
		t.Fatal("false pair survived BMC")
	}
	Ssw_SatStop(p.pMSat)
	p.pMSat = nil
	Ssw_ManCleanup(p)
}
