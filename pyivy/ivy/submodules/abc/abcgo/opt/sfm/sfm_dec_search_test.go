package sfm

import (
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestSfmDecompositionSearchBuffer(t *testing.T) {
	solver := satoko.New()
	solver.SetVariables(2)
	solver.AddClause([]int{1, 2})
	solver.AddClause([]int{0, 3})
	s := sfmDecSearch{solver: solver, target: 1, limit: 6, forced: -1}
	s.patterns.setupSimulation(0, 0, []uint64{0}, false)
	truth, support, ok := s.rec(nil, [2][SFM_SIM_WORDS]uint64{}, nil, false, 0)
	if !ok || len(support) != 1 || support[0] != 0 || truth[0] != 0xaaaaaaaaaaaaaaaa {
		t.Fatalf("truth %x support %v ok %v", truth, support, ok)
	}
}

func TestSfmDecompositionSearchAnd(t *testing.T) {
	solver := satoko.New()
	solver.SetVariables(3)
	solver.AddClause([]int{1, 3, 4})
	solver.AddClause([]int{0, 5})
	solver.AddClause([]int{2, 5})
	s := sfmDecSearch{solver: solver, target: 2, limit: 6, forced: -1, useAndOr: true, fallback: []int{0, 1}}
	s.patterns.setupSimulation(0, 0, []uint64{0, 0}, false)
	truth, support, ok := s.rec(nil, [2][SFM_SIM_WORDS]uint64{}, nil, true, 0)
	if !ok {
		t.Fatal("no decomposition")
	}
	for assignment := 0; assignment < 4; assignment++ {
		idx := 0
		for i, id := range support {
			if assignment&(1<<uint(id)) != 0 {
				idx |= 1 << uint(i)
			}
		}
		got := truth[0]&(1<<uint(idx)) != 0
		if got != (assignment == 3) {
			t.Fatalf("assignment %d truth %x support %v", assignment, truth, support)
		}
	}
}
