package framgr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestFraigSweepLatchCorrelationStopsBeforeInternalNodes(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	internal := g.AND(input, state)
	g.AddOutput(internal)
	aig.Aig_ManSetRegNum(g, 1)
	var pars Fra_Par_t
	Fra_ParamsDefault(&pars)
	pars.FLatchCorr = 1
	p := Fra_ManStart(g, &pars)
	p.PManFraig = Fra_ManPrepareComb(p)
	p.PSml = Fra_SmlStart(g, 0, 1, pars.NSimWords)
	Fra_SmlSimulate(p, 0)
	Fra_FraigSweep(p)
	if Fra_ObjFraig(p, internal, 0) != aig.InvalidEdge {
		t.Fatalf("internal node got mapped during latch-correlation-only sweep: %v", Fra_ObjFraig(p, internal, 0))
	}
	if Fra_ObjFraig(p, state, 0) == aig.InvalidEdge {
		t.Fatal("latch output was not processed before latch-correlation return")
	}
}
