package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"testing"
)

func TestPhaseFlippedCorrespondencePreservesTraces(t *testing.T) {
	rng := rand.New(rand.NewSource(8703))
	for trial := 0; trial < 40; trial++ {
		g := aig.NewGraph()
		pi, q := g.NewInput(), g.NewInput()
		nodes := []aig.Edge{pi, q, g.Const1(), g.Const0()}
		for i := 0; i < 10; i++ {
			a, b := nodes[rng.Intn(len(nodes))], nodes[rng.Intn(len(nodes))]
			if rng.Intn(2) == 0 {
				a = a.Not()
			}
			if rng.Intn(2) == 0 {
				b = b.Not()
			}
			nodes = append(nodes, g.AND(a, b))
		}
		output, next := nodes[4+rng.Intn(len(nodes)-4)], nodes[4+rng.Intn(len(nodes)-4)]
		g.AddOutput(output)
		g.AddOutput(next)
		aig.Aig_ManSetRegNum(g, 1)
		pars := new(Ssw_Pars_t)
		Ssw_ManSetDefaultParams(pars)
		pars.FPolarFlip = 1
		if trial%3 != 0 {
			pars.FDynamic = 1
			pars.FLocalSim = int32(trial % 2)
		}
		pars.nStepsMax = 3
		out := Ssw_SignalCorrespondence(g, pars)
		if out == nil {
			t.Fatalf("trial %d: nil output", trial)
		}
		for seq := 0; seq < 16; seq++ {
			pattern := make([]bool, 4)
			for i := range pattern {
				pattern[i] = seq&(1<<i) != 0
			}
			old, got := sswTrace(t, g, pattern), sswTrace(t, out, pattern)
			for i := range old {
				if old[i] != got[i] {
					t.Fatalf("trial %d input %v: %v != %v", trial, pattern, old, got)
				}
			}
		}
	}
}
