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

func TestScalableBMCSolveAllCountsDistinctOutputs(t *testing.T) {
	g := aig.NewGraph()
	r1, r2 := g.NewInput(), g.NewInput()
	g.AddOutput(g.Const1())
	g.AddOutput(r2)
	g.AddOutput(g.Const1())
	g.AddOutput(r1)
	aig.Aig_ManSetRegNum(g, 2)
	var params Saig_ParBmc_t
	Saig_ParBmcSetDefaultParams(&params)
	params.nFramesMax = 4
	params.fSolveAll = 1
	frames := map[int32]int32{}
	params.pFuncOnFail = func(output int32, cex *util.Abc_Cex_t) int32 {
		if cex == nil || cex.IPo != output {
			t.Fatalf("missing witness for output %d", output)
		}
		if _, exists := frames[output]; exists {
			t.Errorf("duplicate witness for output %d", output)
		}
		frames[output] = cex.IFrame
		return 0
	}
	if result := Saig_ManBmcScalable(g, &params); result != 0 {
		t.Fatalf("result=%d, want failing", result)
	}
	if len(frames) != 2 || frames[0] != 0 || frames[1] != 2 || params.nFailOuts != 2 {
		t.Fatalf("frames=%v count=%d", frames, params.nFailOuts)
	}
}

func TestScalableBMCSolveAllReturnsFailureWhenSomeOutputsSafe(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const1())
	g.AddOutput(g.Const0())
	params := Saig_ParBmc_t{nFramesMax: 2, fSolveAll: 1}
	if result := Saig_ManBmcScalable(g, &params); result != 0 || params.nFailOuts != 1 {
		t.Fatalf("result=%d failures=%d", result, params.nFailOuts)
	}
}
