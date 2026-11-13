package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose2"
	"testing"
)

func TestCec4VerifySATAndSimulationModels(t *testing.T) {
	for mode := int32(0); mode <= 2; mode++ {
		g := gia.Gia_ManStart(16)
		g.EnableMuxes()
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		xor := gia.Gia_ManAppendXorReal(g, x, y^1)
		out := gia.Gia_ManAppendMuxReal(g, z, xor, y)
		gia.Gia_ManAppendCo(g, out)
		cnf := newCec4CNF(g, mode)
		cnf.variable(out >> 1)
		ids := []int32{cnf.variable(x >> 1), cnf.variable(y >> 1), cnf.variable(z >> 1)}
		for mask := 0; mask < 8; mask++ {
			assumptions := make([]int32, 3)
			in := make([]bool, 3)
			for i, v := range ids {
				in[i] = mask&(1<<i) != 0
				assumptions[i] = 2 * v
				if !in[i] {
					assumptions[i]++
				}
			}
			if mode > 0 {
				cnf.solver.SetNewRound()
				for v := 0; v < cnf.solver.NVars(); v++ {
					cnf.solver.MarkCone(glucose2.Var(v))
				}
			}
			if cnf.solver.SolveABCLiterals(assumptions) != glucose2.GLUCOSE_SAT {
				t.Fatalf("mode=%d mask=%d SAT failed", mode, mask)
			}
			expected, err := g.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			gia.Gia_ManIncrementTravId(g)
			got := Cec4_ManVerify_rec(g, out>>1, cnf) != 0
			if got != expected[0] {
				t.Fatalf("mode=%d mask=%d SAT verify=%v want=%v", mode, mask, got, expected[0])
			}
			sim := newCec3Simulation(g.ObjectCount(), 2)
			for i, lit := range []int32{x, y, z} {
				sim.setInputBit(lit>>1, 70, in[i])
			}
			gia.Gia_ManIncrementTravId(g)
			got = Cec4_ManCexVerify_rec(g, out>>1, sim, 70) != 0
			if got != expected[0] {
				t.Fatalf("mode=%d mask=%d sim verify=%v want=%v", mode, mask, got, expected[0])
			}
		}
	}
}
