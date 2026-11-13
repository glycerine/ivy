package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestCollapseExpand(t *testing.T) {
	for _, variant := range []bool{false, true} {
		s := bsat.NewSolver()
		s.NewVariable()
		s.NewVariable()
		s.AddClause(bsat.Negative(0))
		lits := vec.Vec_IntAlloc(2)
		vec.Vec_IntPush(lits, bsat.Positive(0))
		vec.Vec_IntPush(lits, bsat.Positive(1))
		nums, temp := vec.Vec_IntAlloc(2), vec.Vec_IntAlloc(2)
		var status int32
		if variant {
			status = Bmc_CollapseExpand2(s, nil, lits, nums, temp, 0, 0, -1)
		} else {
			status = Bmc_CollapseExpand(s, nil, lits, nums, temp, 0, 0, -1)
		}
		if status != 0 || vec.Vec_IntSize(nums) != 1 || vec.Vec_IntEntry(nums, 0) != 0 {
			t.Fatalf("variant=%v status=%d nums=%v", variant, status, vec.Vec_IntArray(nums))
		}
		if variant && vec.Vec_IntEntry(lits, 1) != bsat.Positive(1) {
			t.Fatal("variant changed original literals")
		}
	}
}
func TestComputeCanonical(t *testing.T) {
	for _, alternate := range []bool{false, true} {
		s := bsat.NewSolver()
		s.NewVariable()
		s.NewVariable()
		s.AddClause(bsat.Positive(0), bsat.Positive(1))
		lits := vec.Vec_IntAlloc(2)
		vec.Vec_IntPush(lits, bsat.Negative(0))
		vec.Vec_IntPush(lits, bsat.Negative(1))
		temp := vec.Vec_IntAlloc(2)
		var status int32
		if alternate {
			status = Bmc_ComputeCanonical2(s, lits, temp, 0)
		} else {
			status = Bmc_ComputeCanonical(s, lits, temp, 0)
		}
		if status != int32(bsat.Satisfiable) || vec.Vec_IntEntry(lits, 0) != bsat.Negative(0) || vec.Vec_IntEntry(lits, 1) != bsat.Positive(1) {
			t.Fatalf("alternate=%v status=%d lits=%v", alternate, status, vec.Vec_IntArray(lits))
		}
	}
}
