package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestFaultAnalyze(t *testing.T) {
	s := bsat.NewSolver()
	for i := 0; i < 3; i++ {
		s.NewVariable()
	}
	s.AddClause(bsat.Negative(0))
	s.AddClause(bsat.Positive(1), bsat.Positive(2))
	pars := vec.Vec_IntAllocArrayCopy([]int32{0, 1, 2}, 3)
	enabled := vec.Vec_IntAllocArrayCopy([]int32{1, 1, 1}, 3)
	scratch := vec.Vec_IntAlloc(3)
	if got := Gia_ManFaultAnalyze(s, pars, enabled, scratch, 1); got != 1 {
		t.Fatalf("unsat count %d", got)
	}
	m := vec.Vec_IntArray(enabled)
	if m[0] != 0 || m[1] != 1 || m[2] != 1 {
		t.Fatalf("fault map %v", m)
	}
}
