package res

import (
	"testing"

	"github.com/glycerine/abcgo/sat/bsat"
)

func literalForValue(variable int32, value bool) bsat.Literal {
	if value {
		return bsat.Positive(variable)
	}
	return bsat.Negative(variable)
}

func solverWithVariables(n int) *bsat.Solver {
	s := bsat.NewSolver()
	for range n {
		s.NewVariable()
	}
	return s
}

func TestResSatAddConst1AllPhases(t *testing.T) {
	for _, phase := range []int32{0, 1} {
		for _, value := range []bool{false, true} {
			s := solverWithVariables(1)
			if got := Res_SatAddConst1(s, 0, phase); got != 1 {
				t.Fatalf("phase %d: helper returned %d", phase, got)
			}
			got := s.Solve(literalForValue(0, value)) == bsat.Satisfiable
			if got != (value == (phase == 0)) {
				t.Fatalf("phase %d value %t: satisfiable=%t", phase, value, got)
			}
		}
	}
}

func TestResSatAddEqualMatchesPhasedEquality(t *testing.T) {
	for _, phase := range []int32{0, 1} {
		for x := 0; x < 2; x++ {
			for y := 0; y < 2; y++ {
				s := solverWithVariables(2)
				if got := Res_SatAddEqual(s, 0, 1, phase); got != 1 {
					t.Fatalf("phase %d: helper returned %d", phase, got)
				}
				assumptions := []bsat.Literal{literalForValue(0, x != 0), literalForValue(1, y != 0)}
				got := s.Solve(assumptions...) == bsat.Satisfiable
				want := (x != 0) == ((y != 0) != (phase != 0))
				if got != want {
					t.Fatalf("phase=%d x=%d y=%d: satisfiable=%t want %t", phase, x, y, got, want)
				}
			}
		}
	}
}

func TestResSatAddAndMatchesPhasedConjunction(t *testing.T) {
	for _, phase0 := range []int32{0, 1} {
		for _, phase1 := range []int32{0, 1} {
			for x0 := 0; x0 < 2; x0++ {
				for x1 := 0; x1 < 2; x1++ {
					for output := 0; output < 2; output++ {
						s := solverWithVariables(3)
						if got := Res_SatAddAnd(s, 2, 0, 1, phase0, phase1); got != 1 {
							t.Fatalf("phases %d,%d: helper returned %d", phase0, phase1, got)
						}
						assumptions := []bsat.Literal{
							literalForValue(0, x0 != 0),
							literalForValue(1, x1 != 0),
							literalForValue(2, output != 0),
						}
						got := s.Solve(assumptions...) == bsat.Satisfiable
						wantOutput := ((x0 != 0) != (phase0 != 0)) && ((x1 != 0) != (phase1 != 0))
						if got != (output != 0 == wantOutput) {
							t.Fatalf("phases=%d,%d inputs=%d,%d output=%d: satisfiable=%t", phase0, phase1, x0, x1, output, got)
						}
					}
				}
			}
		}
	}
}
