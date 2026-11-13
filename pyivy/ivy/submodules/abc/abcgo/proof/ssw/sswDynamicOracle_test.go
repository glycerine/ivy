package ssw

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestDynamicSweepPreservesSequentialBehavior(t *testing.T) {
	for _, local := range []int32{0, 1} {
		rng := rand.New(rand.NewSource(721))
		for trial := 0; trial < 12; trial++ {
			g := aig.NewGraph()
			pi, q := g.NewInput(), g.NewInput()
			nodes := []aig.Edge{pi, q, g.Const1(), g.Const0()}
			for j := 0; j < 8; j++ {
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
			if rng.Intn(2) == 0 {
				output = output.Not()
			}
			if rng.Intn(2) == 0 {
				next = next.Not()
			}
			g.AddOutput(output)
			g.AddOutput(next)
			aig.Aig_ManSetRegNum(g, 1)
			var pars Ssw_Pars_t
			Ssw_ManSetDefaultParams(&pars)
			pars.fDynamic = 1
			pars.fLocalSim = local
			pars.nStepsMax = 3
			reduced := Ssw_SignalCorrespondence(g, &pars)
			if reduced == nil {
				t.Fatalf("local=%d trial=%d no result", local, trial)
			}
			for seq := 0; seq < 16; seq++ {
				oldState, newState := false, false
				for frame := 0; frame < 4; frame++ {
					input := seq&(1<<frame) != 0
					old, err := g.Evaluate([]bool{input, oldState})
					if err != nil {
						t.Fatal(err)
					}
					args := []bool{input}
					if aig.Aig_ManRegNum(reduced) > 0 {
						args = append(args, newState)
					}
					got, err := reduced.Evaluate(args)
					if err != nil {
						t.Fatalf("local=%d trial=%d: %v", local, trial, err)
					}
					if old[0] != got[0] {
						t.Fatalf("local=%d trial=%d sequence=%d frame=%d mismatch", local, trial, seq, frame)
					}
					oldState = old[1]
					if aig.Aig_ManRegNum(reduced) > 0 {
						newState = got[1]
					}
				}
			}
		}
	}
}
