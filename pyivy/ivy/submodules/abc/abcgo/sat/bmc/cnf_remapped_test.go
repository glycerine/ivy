package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
)

func TestCnfDeriveGiaRemappedAssertsEveryOutput(t *testing.T) {
	graph := aig.NewGraph()
	input := graph.NewInput()
	graph.AddOutput(input)
	manager := gia.Gia_ManFromAig(graph)

	data := Cnf_DeriveGiaRemapped(manager)
	if data == nil {
		t.Fatal("expected CNF")
	}
	if err := data.Validate(); err != nil {
		t.Fatal(err)
	}
	assertsInput := false
	for _, clause := range data.Clauses {
		if len(clause) == 1 && clause[0] == 0 {
			assertsInput = true
			break
		}
	}
	if data.Variables < 1 || !assertsInput {
		t.Fatalf("CNF=%+v, want an asserted positive input output", data)
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
