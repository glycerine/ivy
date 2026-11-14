package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// A deterministic shift register exposes errors in selected-pattern ancestry:
// the assertion becomes true only after several rounds of simulation.
func rarityShiftGraph(n int) *aig.Graph {
	g := aig.NewGraph()
	regs := make([]aig.Edge, n)
	for i := range regs {
		regs[i] = g.NewInput()
	}
	g.AddOutput(regs[n-1])
	g.AddOutput(g.Const1())
	for i := 1; i < n; i++ {
		g.AddOutput(regs[i-1])
	}
	aig.Aig_ManSetRegNum(g, int32(n))
	return g
}

func TestRarityMultiRoundCounterexample(t *testing.T) {
	for _, words := range []int32{1, 2} {
		g := rarityShiftGraph(5)
		pars := &Ssw_RarPars_t{NWords: words, NFrames: 2, NBinSize: 2, NRounds: 4}
		if got := Ssw_RarSimulate(g, pars); got != 0 {
			t.Fatalf("words=%d: result %d", words, got)
		}
		cex := g.SequentialCounterexample()
		if cex == nil || cex.IFrame != 5 || saig.Saig_ManVerifyCex(g, cex) == 0 {
			t.Fatalf("words=%d: invalid counterexample %+v", words, cex)
		}
	}
}

func TestRaritySolveAllOutputsAndCallback(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(b)
	g.AddOutput(g.Const1())
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 2)
	var called []int32
	pars := &Ssw_RarPars_t{NWords: 1, NFrames: 1, NBinSize: 1, NRounds: 4, FSolveAll: 1, FNotVerbose: 1,
		pFuncOnFail: func(po int32, _ *util.Abc_Cex_t) int32 { called = append(called, po); return 0 },
	}
	if got := Ssw_RarSimulate(g, pars); got != 0 {
		t.Fatalf("result %d", got)
	}
	if pars.nSolved != 2 || len(called) != 2 || called[0] != 0 || called[1] != 1 {
		t.Fatalf("solved %d callbacks %v", pars.nSolved, called)
	}
	v := g.SequentialCounterexamples()
	if v == nil || vec.Vec_PtrEntry(v, 0) == nil || vec.Vec_PtrEntry(v, 1) == nil {
		t.Fatalf("missing solved output markers: %v", v)
	}
}

func TestRarityTransferUsesConfiguredBinWidth(t *testing.T) {
	for _, width := range []int32{1, 3, 5, 8, 16} {
		g := aig.NewGraph()
		for i := 0; i < 80; i++ {
			g.NewInput()
		}
		g.AddOutput(g.Const0())
		for i := 0; i < 80; i++ {
			g.AddOutput(g.Const0())
		}
		aig.Aig_ManSetRegNum(g, 80)
		p := Ssw_RarManStart(g, &Ssw_RarPars_t{NWords: 1, NFrames: 1, NBinSize: width})
		// This pattern crosses a 64-bit word boundary at widths 3 and 5.
		for i, li := range g.Outputs()[1:] {
			if (i*17+3)%11 < 5 {
				Ssw_RarObjSim(p, int32(li.NodeID()))[0] = ^uint64(0)
			}
		}
		Ssw_RarTransferPatterns(p, vec.Vec_IntAlloc(80))
		for bin := int32(0); bin < p.nGroups; bin++ {
			var want int32
			for bit := int32(0); bit < width; bit++ {
				i := bin*width + bit
				if (i*17+3)%11 < 5 {
					want |= 1 << bit
				}
			}
			if got := Ssw_RarGetBinPat(p, bin, want); got != 64 {
				t.Fatalf("width %d bin %d pattern %d: count %d", width, bin, want, got)
			}
		}
		Ssw_RarManStop(p)
	}
}

func TestRarityReplaysRandomInputsAcrossRounds(t *testing.T) {
	for seed := int32(0); seed < 10; seed++ {
		g := aig.NewGraph()
		pi := g.NewInput()
		regs := make([]aig.Edge, 5)
		for i := range regs {
			regs[i] = g.NewInput()
		}
		g.AddOutput(g.AND(pi, regs[4]))
		g.AddOutput(pi)
		for i := 1; i < len(regs); i++ {
			g.AddOutput(regs[i-1])
		}
		aig.Aig_ManSetRegNum(g, int32(len(regs)))
		pars := &Ssw_RarPars_t{NWords: 2, NFrames: 2, NBinSize: 2, NRounds: 5, NRandSeed: seed}
		if got := Ssw_RarSimulate(g, pars); got != 0 {
			t.Fatalf("seed=%d: result %d", seed, got)
		}
		cex := g.SequentialCounterexample()
		if cex == nil || cex.IFrame < 5 || saig.Saig_ManVerifyCex(g, cex) == 0 {
			t.Fatalf("seed=%d: invalid counterexample %+v", seed, cex)
		}
	}
}
