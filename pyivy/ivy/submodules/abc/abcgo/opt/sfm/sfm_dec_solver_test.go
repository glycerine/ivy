package sfm

import (
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestSfmDecWindowSolver(t *testing.T) {
	// x -> pivot buffer -> root buffer. A pivot flip must change the root.
	cnfs := [][][]int{Sfm_TruthToCnf(0xaaaaaaaaaaaaaaaa, nil, 1)}
	w := sfmDecWindow{gates: []int{-1, 0, 0}, fanins: [][]int{nil, {0}, {1}}, roots: []int{2}, pivot: 1}
	solver, ok := w.prepareSolver(cnfs)
	if !ok {
		t.Fatal("solver rejected window")
	}
	for _, input := range []int{0, 1} {
		if solver.SolveAssumptions([]int{input}) != satoko.SATOKO_SAT {
			t.Fatalf("input literal %d unsatisfiable", input)
		}
	}
}
