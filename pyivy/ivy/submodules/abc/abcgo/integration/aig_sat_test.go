package integration_test

import (
	"testing"

	aig "github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/integration"
	"github.com/glycerine/abcgo/sat/satoko"
)

// Exercise graph synthesis and SAT encoding together through public APIs.
// Enumerating every three-input Boolean function also covers constants and
// complemented edges, with repeated assumption changes on one solver.
func TestMiniAIGSatokoTruthParity(t *testing.T) {
	for truth := uint32(0); truth < 256; truth++ {
		p := aig.Mini_AigStart()
		inputs := []int{aig.Mini_AigCreatePi(p), aig.Mini_AigCreatePi(p), aig.Mini_AigCreatePi(p)}
		out := aig.Mini_AigTruth(p, inputs, 3, truth*0x01010101)
		s, err := integration.MiniAIGSolver(p)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			assumptions := make([]int, 0, 4)
			for i, lit := range inputs {
				if assignment&(1<<i) == 0 {
					lit ^= 1
				}
				assumptions = append(assumptions, lit)
			}
			expected := truth>>assignment&1 != 0
			for _, desired := range []bool{false, true} {
				lit := out
				if !desired {
					lit ^= 1
				}
				got := s.SolveAssumptions(append(assumptions, lit))
				want := satoko.SATOKO_UNSAT
				if desired == expected {
					want = satoko.SATOKO_SAT
				}
				if got != want {
					t.Fatalf("truth=%02x assignment=%d desired=%v: got %d want %d", truth, assignment, desired, got, want)
				}
				if got == satoko.SATOKO_SAT {
					value := func(l int) bool {
						v := s.ModelValue(l >> 1)
						if v != 0 && v != 1 {
							t.Fatalf("unassigned model node %d", l>>1)
						}
						return v^(l&1) == 1
					}
					if value(0) || value(out) != desired {
						t.Fatal("model disagrees with circuit output")
					}
					for i, input := range inputs {
						if value(input) != (assignment&(1<<i) != 0) {
							t.Fatal("model violates input assumptions")
						}
					}
					for id := 1; id < aig.Mini_AigNodeNum(p); id++ {
						if aig.Mini_AigNodeIsAnd(p, id) && value(2*id) != (value(aig.Mini_AigNodeFanin0(p, id)) && value(aig.Mini_AigNodeFanin1(p, id))) {
							t.Fatalf("model violates AND node %d", id)
						}
					}
				}
			}
		}
		aig.Mini_AigStop(p)
	}
}
