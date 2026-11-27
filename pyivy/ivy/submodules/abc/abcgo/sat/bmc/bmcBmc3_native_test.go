package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestScalableBMCFindsOneFrameCounterexample(t *testing.T) {
	g := aig.NewGraph()
	pi, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	var params Saig_ParBmc_t
	Saig_ParBmcSetDefaultParams(&params)
	params.NFramesMax = 2
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

func TestScalableBMCBackendSelectionMatchesSourcePriority(t *testing.T) {
	for _, tc := range []struct {
		name            string
		satoko, glucose int32
	}{
		{"bsat", 0, 0}, {"satoko", 1, 0}, {"glucose", 0, 1}, {"satoko_precedes_glucose", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			pi, state := g.NewInput(), g.NewInput()
			g.AddOutput(state)
			g.AddOutput(pi)
			aig.Aig_ManSetRegNum(g, 1)
			params := Saig_ParBmc_t{NFramesMax: 2, FUseSatoko: tc.satoko, FUseGlucose: tc.glucose}
			if got := Saig_ManBmcScalable(g, &params); got != 0 {
				t.Fatalf("result=%d, want witness", got)
			}
			if c := g.SequentialCounterexample(); c == nil || c.IFrame != 1 || c.IPo != 0 {
				t.Fatalf("witness=%+v", c)
			}
		})
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
	params.NFramesMax = 2
	if got := Saig_ManBmcScalable(g, &params); got != -1 {
		t.Fatalf("result=%d, want unknown after finite bound (-1)", got)
	}
}

func TestScalableBMCUnboundedCompletenessCutoff(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state)
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	var params Saig_ParBmc_t
	Saig_ParBmcSetDefaultParams(&params)
	if got := Saig_ManBmcScalable(g, &params); got != 1 {
		t.Fatalf("result=%d, want complete finite-state proof (1)", got)
	}
}

func TestScalableBMCStartAndWitnessReplacement(t *testing.T) {
	g := aig.NewGraph()
	pi, state := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	old := util.Abc_CexAlloc(1, 1, 1)
	old.IFrame, old.IPo = 0, 0
	if err := g.SetSequentialCounterexample(old); err != nil {
		t.Fatal(err)
	}
	params := Saig_ParBmc_t{NStart: 1, NFramesMax: 2}
	if got := Saig_ManBmcScalable(g, &params); got != 0 {
		t.Fatalf("result=%d, want witness", got)
	}
	cex := g.SequentialCounterexample()
	if cex == nil || cex == old || cex.IFrame != 1 || cex.IPo != 0 {
		t.Fatalf("witness did not replace old model with frame-1 output-0 model: old=%p got=%+v", old, cex)
	}
	if !util.Abc_InfoHasBit(cex.PData, 2) {
		t.Fatalf("frame-1 PI absent from witness bits %v", cex.PData)
	}
}

func TestScalableBMCJumpRaisesConflictBudget(t *testing.T) {
	g := aig.NewGraph()
	x := make([]aig.Edge, 12)
	for i := range x {
		x[i] = g.NewInput()
	}
	clause := func(lits ...aig.Edge) aig.Edge {
		v := g.Const0()
		for _, lit := range lits {
			v = g.OR(v, lit)
		}
		return v
	}
	parts := []aig.Edge{}
	for pigeon := 0; pigeon < 4; pigeon++ {
		parts = append(parts, clause(x[3*pigeon], x[3*pigeon+1], x[3*pigeon+2]))
		for a := 0; a < 3; a++ {
			for b := a + 1; b < 3; b++ {
				parts = append(parts, clause(x[3*pigeon+a].Not(), x[3*pigeon+b].Not()))
			}
		}
	}
	for hole := 0; hole < 3; hole++ {
		for p := 0; p < 4; p++ {
			for q := p + 1; q < 4; q++ {
				parts = append(parts, clause(x[3*p+hole].Not(), x[3*q+hole].Not()))
			}
		}
	}
	property := g.Const1()
	for _, part := range parts {
		property = g.AND(property, part)
	}
	g.AddOutput(property)
	params := Saig_ParBmc_t{NFramesMax: 2, NConfLimit: 1, NFramesJump: 1, NConfLimitJump: 0}
	if got := Saig_ManBmcScalable(g, &params); got != -1 {
		t.Fatalf("result=%d, want bounded unknown after jump", got)
	}
	if params.NConfLimit != 0 {
		t.Fatalf("conflict budget after jump=%d, want unlimited", params.NConfLimit)
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
	params.NFramesMax = 4
	params.FSolveAll = 1
	params.FStoreCex = 1
	if result := Saig_ManBmcScalable(g, &params); result != 0 {
		t.Fatalf("result=%d, want failing", result)
	}
	frames := map[int32]int32{}
	for output := int32(0); output < 2; output++ {
		if cex, ok := vec.Vec_PtrEntry(g.SequentialCounterexamples(), int(output)).(*util.Abc_Cex_t); ok && cex != nil {
			if cex.IPo != output {
				t.Fatalf("output %d stored witness %d", output, cex.IPo)
			}
			frames[output] = cex.IFrame
		}
	}
	if len(frames) != 2 || frames[0] != 0 || frames[1] != 2 || params.NFailOuts != 2 {
		t.Fatalf("frames=%v count=%d", frames, params.NFailOuts)
	}
}

func TestScalableBMCSolveAllReturnsFailureWhenSomeOutputsSafe(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const1())
	g.AddOutput(g.Const0())
	params := Saig_ParBmc_t{NFramesMax: 2, FSolveAll: 1, FStoreCex: 1}
	if result := Saig_ManBmcScalable(g, &params); result != 0 || params.NFailOuts != 1 {
		t.Fatalf("result=%d failures=%d", result, params.NFailOuts)
	}
}

func TestScalableBMCSolveAllCallbackRespectsCEXStorageOption(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const1())
	params := Saig_ParBmc_t{NFramesMax: 1, FSolveAll: 1}
	called := false
	params.pFuncOnFail = func(_ int32, cex *util.Abc_Cex_t) int32 {
		called = true
		if cex != nil {
			t.Fatalf("callback received unstored CEX: %+v", cex)
		}
		return 0
	}
	if result := Saig_ManBmcScalable(g, &params); result != 0 || !called {
		t.Fatalf("result=%d callback=%v", result, called)
	}
	// C keeps p->vCexes when solving all outputs; without -x a failed
	// output's entry is the sentinel (Abc_Cex_t *)1, not its CEX
	v := g.SequentialCounterexamples()
	if v == nil || vec.Vec_PtrSize(v) != 1 || vec.Vec_PtrEntry(v, 0) != util.AbcCexSentinel {
		t.Fatalf("CEX vector = %v, want the sentinel for output 0", v)
	}
}
