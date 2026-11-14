package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// ABC initializes timeLastSolved to zero and only starts the gap clock after
// finding a failing output in solve-all mode.
func TestRarityGapDeadlineStartsAfterFirstSolvedOutput(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 1)
	pars := &Ssw_RarPars_t{NWords: 1, NFrames: 2, NBinSize: 1, NRounds: 1, TimeOutGap: -1}
	if got := Ssw_RarSimulate(g, pars); got != 0 {
		t.Fatalf("result %d; gap before any solved output should not stop simulation", got)
	}
}

// C treats any nonzero timeout as a deadline, including negative values.
func TestRarity2NegativeTimeoutStopsAfterFirstRound(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(g.AND(input, state.Not()))
	aig.Aig_ManSetRegNum(g, 1)
	if got := Ssw_RarSimulate2(g, 1, 1, 1, 2, -1, 0); got != -1 {
		t.Fatalf("result %d; source stops after first round", got)
	}
}
