package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestScalableBMCFindsOneFrameCounterexample(t *testing.T) {
	g := aig.NewGraph()
	pi, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	var params Saig_ParBmc_t
	Saig_ParBmcSetDefaultParams(&params)
	params.nFramesMax = 2
	if got := Saig_ManBmcScalable(g, &params); got != 0 {
		t.Fatalf("result=%d, want counterexample (0)", got)
	}
	cex := g.SequentialCounterexample()
	if cex == nil || cex.IFrame != 1 || cex.IPo != 0 {
		t.Fatalf("counterexample=%+v, want frame 1 property 0", cex)
	}
	if !util.Abc_InfoHasBit(cex.PData, 1) {
		t.Fatal("frame 0 PI input missing from counterexample")
	}
}

func TestScalableBMCProvesZeroStateInvariant(t *testing.T) {
	g := aig.NewGraph()
	_, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	var params Saig_ParBmc_t
	Saig_ParBmcSetDefaultParams(&params)
	params.nFramesMax = 2
	if got := Saig_ManBmcScalable(g, &params); got != 1 {
		t.Fatalf("result=%d, want bounded-safe (1)", got)
	}
}
