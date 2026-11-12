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
