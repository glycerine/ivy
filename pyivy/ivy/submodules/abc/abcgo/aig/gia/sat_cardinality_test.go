package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSbmCardinalitySolver(t *testing.T) {
	var outputs *Vec_Int_t
	s := Sbm_AddCardinSolver(2, &outputs)
	for assignment := 0; assignment < 16; assignment++ {
		count := 0
		assumptions := make([]bsat.Literal, 0, 5)
		for i := int32(0); i < 4; i++ {
			if assignment&(1<<i) != 0 {
				assumptions = append(assumptions, bsat.Positive(i))
				count++
			} else {
				assumptions = append(assumptions, bsat.Negative(i))
			}
		}
		for k := 0; k < 4; k++ {
			trial := append(append([]bsat.Literal(nil), assumptions...), bsat.Negative(vec.Vec_IntEntry(outputs, k)))
			got := s.Solve(trial...) == bsat.Satisfiable
			if got != (count <= k) {
				t.Fatalf("assignment %d bound %d got %v", assignment, k, got)
			}
		}
	}
}
