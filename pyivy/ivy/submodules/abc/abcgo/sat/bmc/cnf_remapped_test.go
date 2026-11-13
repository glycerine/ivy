package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestCnfDeriveGiaRemappedPreservesOutputsAndIDs(t *testing.T) {
	p := gia.Gia_ManStart(12)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	and := gia.Gia_ManAppendAnd(p, a, b)
	// Interleaving an input with internal nodes makes AIG and GIA IDs diverge.
	c := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, and^1)
	gia.Gia_ManAppendCo(p, c)
	data := Cnf_DeriveGiaRemapped(p)
	if data == nil {
		t.Fatal("expected CNF")
	}
	if err := data.Validate(); err != nil {
		t.Fatal(err)
	}
	s := bsat.NewSolver()
	for s.NumVars() < int32(data.Variables) {
		s.NewVariable()
	}
	for _, clause := range data.Clauses {
		if !s.AddClause(clause...) {
			t.Fatal("inconsistent CNF")
		}
	}
	ids := []int32{a >> 1, b >> 1, c >> 1, gia.Gia_ObjId(p, gia.Gia_ManCo(p, 0)), gia.Gia_ObjId(p, gia.Gia_ManCo(p, 1))}
	for bits := 0; bits < 32; bits++ {
		assumptions := make([]bsat.Literal, len(ids))
		for i, id := range ids {
			if data.VarNums[id] < 0 {
				t.Fatalf("object %d not mapped", id)
			}
			assumptions[i] = bsat.Positive(data.VarNums[id])
			if bits&(1<<i) == 0 {
				assumptions[i] ^= 1
			}
		}
		want := ((bits&8) != 0) == !((bits&1) != 0 && (bits&2) != 0) && ((bits&16) != 0) == ((bits&4) != 0)
		if got := s.Solve(assumptions...) == bsat.Satisfiable; got != want {
			t.Fatalf("bits=%05b got %v want %v", bits, got, want)
		}
	}
}

func TestCnfDeriveGiaRemappedNil(t *testing.T) {
	if got := Cnf_DeriveGiaRemapped(nil); got != nil {
		t.Fatalf("nil graph yielded CNF %v", got)
	}
}

func TestCnfDataLiftGiaPreservesSatisfyingAssignments(t *testing.T) {
	graph := aig.NewGraph()
	x, y := graph.NewInput(), graph.NewInput()
	graph.AddOutput(graph.AND(x, y))
	manager := gia.Gia_ManFromAig(graph)
	data := Cnf_DeriveGiaRemapped(manager)
	if data == nil {
		t.Fatal("missing CNF")
	}
	original := make([][]int32, len(data.Clauses))
	for i, clause := range data.Clauses {
		original[i] = append([]int32(nil), clause...)
	}
	originalVars := data.Variables
	Cnf_DataLiftGia(data, manager, 3)
	if data.Variables != originalVars {
		t.Fatalf("variables=%d want=%d", data.Variables, originalVars)
	}
	for i, clause := range original {
		for j, lit := range clause {
			if got := data.Clauses[i][j]; got != lit+6 {
				t.Fatalf("clause %d literal %d: got %d want %d", i, j, got, lit+6)
			}
		}
	}
	Cnf_DataLiftGia(data, manager, -3)
	if err := data.Validate(); err != nil {
		t.Fatal(err)
	}
}
