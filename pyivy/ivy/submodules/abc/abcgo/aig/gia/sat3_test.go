package gia

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSat3Init(t *testing.T) {
	sat := Gia_ManSat3Init(&Cnf_Dat_t{Clauses: [][]int32{{0, 2}, {1, 2}}})
	if sat == nil || sat.Solve() != bsat.Satisfiable {
		t.Fatal("expected satisfiable CNF")
	}
	if Gia_ManSat3Init(&Cnf_Dat_t{Clauses: [][]int32{{0}, {1}}}) != nil {
		t.Fatal("expected contradiction")
	}
}
func TestSat3Create(t *testing.T) {
	for _, want := range []bool{false, true} {
		p := Gia_ManStart(4)
		a := Gia_ManAppendCi(p)
		lit := a
		if !want {
			lit = 0
		}
		Gia_ManAppendCo(p, lit)
		solver := Gia_ManSat3Create(p)
		status := bsat.Unsatisfiable
		if solver != nil {
			status = solver.Solve()
		}
		if (status == bsat.Satisfiable) != want {
			t.Fatalf("wanted satisfiable=%v got %v", want, status)
		}
	}
}
