package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestStopWhenGoneRetainsSourceIterationCount(t *testing.T) {
	g := aig.NewGraph()
	x, latch := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, latch))
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)

	var pars Ssw_Pars_t
	Ssw_ManSetDefaultParams(&pars)
	pars.FStopWhenGone = 1
	pars.nIters = 37
	result := Ssw_SignalCorrespondence(g, &pars)
	if result == nil {
		t.Fatal("source stop path must return a copy of the original graph")
	}
	if pars.nIters != 37 {
		t.Fatalf("iteration count = %d, want source's unchanged value 37", pars.nIters)
	}
	if aig.Aig_ManNodeNum(result) != aig.Aig_ManNodeNum(g) || aig.Aig_ManRegNum(result) != aig.Aig_ManRegNum(g) {
		t.Fatal("source stop path did not preserve the original graph")
	}
}
