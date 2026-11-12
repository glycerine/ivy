package sbd

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"math/bits"
	"testing"
)

func TestLutParameterClausesExhaustive(t *testing.T) {
	structure := []Sbd_Str_t{NewLUTStructure([]int32{0, 1})}
	mapping := []int32{0, 1, 2, 3, 4, 5, 6}
	solver := bsat.NewSolver()
	if Sbd_ProblemAddClauses(solver, 2, 1, mapping, structure) != 1 {
		t.Fatal("LUT clause setup")
	}
	for truth := int32(0); truth < 16; truth++ {
		for input := int32(0); input < 4; input++ {
			for output := int32(0); output < 2; output++ {
				assumptions := []bsat.Literal{bsat.Literal(1 - (input & 1)), bsat.Literal(3 - (input >> 1 & 1)), bsat.Literal(5 - output)}
				for row := int32(0); row < 4; row++ {
					assumptions = append(assumptions, bsat.Literal(2*(3+row)+1-(truth>>row&1)))
				}
				status := solver.Solve(assumptions...)
				if (status == bsat.Satisfiable) != (output == truth>>input&1) {
					t.Fatalf("truth%x input%d output%d", truth, input, output)
				}
			}
		}
	}
}
func TestSelectorOneHotClauses(t *testing.T) {
	structures := []Sbd_Str_t{NewSelectorStructure([]int32{0, 1, 2})}
	mapping := []int32{0, 1, 2, 3, 4, 5, 6}
	solver := bsat.NewSolver()
	Sbd_ProblemAddClausesInit(solver, 3, 1, mapping, structures)
	Sbd_ProblemAddClauses(solver, 3, 1, mapping, structures)
	for selector := uint32(0); selector < 8; selector++ {
		for inputs := uint32(0); inputs < 8; inputs++ {
			for output := uint32(0); output < 2; output++ {
				var assumptions []bsat.Literal
				for v := uint32(0); v < 3; v++ {
					assumptions = append(assumptions, bsat.Literal(2*v+1-(inputs>>v&1)), bsat.Literal(2*(v+4)+1-(selector>>v&1)))
				}
				assumptions = append(assumptions, bsat.Literal(7-output))
				want := bits.OnesCount32(selector) == 1 && output == inputs>>uint(bits.TrailingZeros32(selector))&1
				if (solver.Solve(assumptions...) == bsat.Satisfiable) != want {
					t.Fatalf("selector%x input%x output%d", selector, inputs, output)
				}
			}
		}
	}
}
func TestCegisFindsAndAndRejectsMissingDivisor(t *testing.T) {
	network, mirrors, window, mapping, pivot := simpleAndWindow()
	empty := sbdVector()
	structure := []Sbd_Str_t{NewLUTStructure([]int32{0, 1})}
	if Sbd_ProblemSolve(network, mirrors, pivot, window, mapping, empty, empty, sbdVector(1, 2), 1, structure) != 1 || structure[0].Res != 0x8888888888888888 {
		t.Fatalf("AND synthesis result%x", structure[0].Res)
	}
	incomplete := []Sbd_Str_t{NewLUTStructure([]int32{0})}
	if Sbd_ProblemSolve(network, mirrors, pivot, window, mapping, empty, empty, sbdVector(1), 1, incomplete) != 0 {
		t.Fatal("accepted implementation with missing divisor")
	}
	selector := []Sbd_Str_t{NewSelectorStructure([]int32{0, 1})}
	if Sbd_ProblemSolve(network, mirrors, pivot, window, mapping, empty, empty, sbdVector(1, 2), 1, selector) != 0 {
		t.Fatal("AND implemented by a simple selector")
	}
}
func TestCollectStructureSolution(t *testing.T) {
	structures := []Sbd_Str_t{NewLUTStructure([]int32{0, 1}), NewSelectorStructure([]int32{2, 3, 4})}
	literals := vec.Vec_IntAlloc(7)
	for i, value := range []int32{0, 1, 1, 0, 0, 1, 0} {
		vec.Vec_IntPush(literals, 2*int32(i)+1-value)
	}
	Sbd_ProblemCollectSolution(2, structures, literals)
	if structures[0].Res != 0x6666666666666666 || structures[1].Res != 3 {
		t.Fatal("parameter model decoding")
	}
}
