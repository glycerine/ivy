package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// The PDR front end accepts native AIG XOR nodes, including complemented
// edges. Compare multi-output results with independent reachable-state search.
func TestPdrNewJustificationNativeXorAgainstReachability(t *testing.T) {
	for seed := 0; seed < 64; seed++ {
		g := aig.NewGraph()
		x, q0, q1 := g.NewInput(), g.NewInput(), g.NewInput()
		polarity := func(edge aig.Edge, bit uint) aig.Edge {
			if seed&(1<<bit) != 0 {
				return edge.Not()
			}
			return edge
		}
		t0 := g.NewXor(polarity(x, 0), polarity(q0, 1))
		t1 := g.NewXor(polarity(q0, 2), polarity(q1, 3))
		o0 := g.NewXor(polarity(q0, 4), q1)
		o1 := g.NewXor(polarity(x, 5), q1)
		g.AddOutput(o0)
		g.AddOutput(o1)
		g.AddOutput(t0)
		g.AddOutput(t1)
		aig.Aig_ManSetRegNum(g, 2)
		flags := func(b bool, bit uint) bool { return b != (seed&(1<<bit) != 0) }
		reachable := [4]bool{true}
		var fails [2]bool
		for changed := true; changed; {
			changed = false
			for s := 0; s < 4; s++ {
				if !reachable[s] {
					continue
				}
				a, b := s&1 != 0, s&2 != 0
				for i := 0; i < 2; i++ {
					in := i != 0
					fails[0] = fails[0] || (flags(a, 4) != b)
					fails[1] = fails[1] || (flags(in, 5) != b)
					na, nb := flags(in, 0) != flags(a, 1), flags(a, 2) != flags(b, 3)
					next := 0
					if na {
						next |= 1
					}
					if nb {
						next |= 2
					}
					if !reachable[next] {
						reachable[next], changed = true, true
					}
				}
			}
		}
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fSilent, pars.fSolveAll, pars.fStoreCex, pars.fNewXSim = 1, 1, 1, 1
		pars.fMonoCnf = int32(seed & 1)
		Pdr_ManSolve(g, &pars)
		models := g.SequentialCounterexamples()
		for i, expectedFailure := range fails {
			status := vec.Vec_IntEntry(pars.vOutMap, i)
			witness, _ := vec.Vec_PtrEntry(models, i).(*util.Abc_Cex_t)
			if expectedFailure {
				if status != 0 || witness == nil || saig.Saig_ManVerifyCex(g, witness) != 1 {
					t.Fatalf("seed %d output %d: expected replayable failure, status %d witness %+v", seed, i, status, witness)
				}
			} else if status != 1 {
				t.Fatalf("seed %d output %d: expected proof, status %d", seed, i, status)
			}
		}
	}
}
