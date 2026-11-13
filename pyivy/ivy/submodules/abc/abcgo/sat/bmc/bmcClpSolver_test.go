package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func andCollapseSolver() *bsat.Solver {
	s := bsat.NewSolver()
	for i := 0; i < 5; i++ {
		s.NewVariable()
	}
	s.AddClause(bsat.Negative(2), bsat.Positive(3))
	s.AddClause(bsat.Negative(2), bsat.Positive(4))
	s.AddClause(bsat.Positive(2), bsat.Negative(3), bsat.Negative(4))
	return s
}
func collapseCoverValue(sop []byte, in []bool) bool {
	def := len(sop) > len(in)+1 && sop[len(in)+1] == '0'
	for len(sop) > 1 && sop[0] != 0 {
		match := true
		for i, v := range in {
			if sop[i] != '-' && (sop[i] == '1') != v {
				match = false
			}
		}
		if match {
			return !def
		}
		sop = sop[len(in)+3:]
	}
	return def
}
func TestSolverCollapse(t *testing.T) {
	for _, pair := range []bool{false, true} {
		for _, canon := range []int32{0, 1} {
			for _, rev := range []int32{0, 1} {
				a := andCollapseSolver()
				var sop *vec.Vec_Str_t
				if pair {
					sop = Bmc_CollapseOne_int2(a, andCollapseSolver(), 2, 0, 0, canon, rev, 0)
				} else {
					sop = Bmc_CollapseOne_int(a, 2, 0, 0, canon, rev, 0)
				}
				if sop == nil {
					t.Fatalf("missing pair=%v canon=%d rev=%d", pair, canon, rev)
				}
				for bits := 0; bits < 4; bits++ {
					in := []bool{bits&1 != 0, bits&2 != 0}
					if collapseCoverValue(vec.Vec_StrArray(sop), in) != (in[0] && in[1]) {
						t.Fatalf("cover=%q in=%v", vec.Vec_StrArray(sop), in)
					}
				}
			}
		}
	}
}

func TestFourSolverCollapse(t *testing.T) {
	makeSolver := func() *bsat.Solver {
		s := bsat.NewSolver()
		for i := 0; i < 4; i++ {
			s.NewVariable()
		}
		s.AddClause(bsat.Negative(1), bsat.Positive(2))
		s.AddClause(bsat.Negative(1), bsat.Positive(3))
		s.AddClause(bsat.Positive(1), bsat.Negative(2), bsat.Negative(3))
		return s
	}
	for _, canon := range []int32{0, 1} {
		for _, rev := range []int32{0, 1} {
			sop := Bmc_CollapseOne_int3(makeSolver(), makeSolver(), makeSolver(), makeSolver(), 2, 0, 0, canon, rev, 0)
			if sop == nil {
				t.Fatalf("no cover canon=%d rev=%d", canon, rev)
			}
			for bits := 0; bits < 4; bits++ {
				in := []bool{bits&1 != 0, bits&2 != 0}
				if collapseCoverValue(vec.Vec_StrArray(sop), in) != (in[0] && in[1]) {
					t.Fatalf("cover=%q in=%v", vec.Vec_StrArray(sop), in)
				}
			}
		}
	}
}
