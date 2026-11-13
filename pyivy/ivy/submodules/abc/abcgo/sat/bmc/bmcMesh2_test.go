package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestMesh2OneHot(t *testing.T) {
	s := bsat.NewSolver()
	for i := 0; i < 3; i++ {
		s.NewVariable()
	}
	for i := int32(0); i < 3; i++ {
		if !s.AddClause(bsat.Positive(i)) {
			t.Fatal("unit clause")
		}
	}
	if s.Solve() != bsat.Satisfiable {
		t.Fatal("expected SAT")
	}
	// A model with multiple selected variables must be excluded, even when it
	// forces an immediate contradiction with the original clauses.
	// Use a separate unconstrained solver to examine all generated clauses.
	s = bsat.NewSolver()
	for i := 0; i < 3; i++ {
		s.NewVariable()
	}
	snapshot := []bool{true, true, true}
	if got := mesh2OneHot(s, snapshot, 0, 3); got != 3 {
		t.Fatalf("got %d clauses", got)
	}
	for mask := 0; mask < 8; mask++ {
		var a []bsat.Literal
		count := 0
		for i := int32(0); i < 3; i++ {
			if mask&(1<<i) != 0 {
				a = append(a, bsat.Positive(i))
				count++
			} else {
				a = append(a, bsat.Negative(i))
			}
		}
		if got := s.Solve(a...) == bsat.Satisfiable; got != (count <= 1) {
			t.Fatalf("mask %d", mask)
		}
	}
}
func TestMesh2SmallGraph(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	Bmc_MeshTest2(p, 3, 3, 2, 0)
}
