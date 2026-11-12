package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func TestBmcSolverForCNFChecksOutputsIndependently(t *testing.T) {
	graph := aig.NewGraph()
	input := graph.NewInput()
	graph.AddOutput(input)
	graph.AddOutput(graph.Const0())
	data, err := cnf.EncodeAIG(graph, 0)
	if err != nil {
		t.Fatal(err)
	}
	solver, err := bmcSolverForCNF(data)
	if err != nil {
		t.Fatal(err)
	}
	outputs := graph.Outputs()
	positiveInput := data.VarNums[outputs[0].NodeID()]
	constantFalse := data.VarNums[outputs[1].NodeID()]
	if got := solver.Solve(bsat.Positive(positiveInput)); got != bsat.Satisfiable {
		t.Fatalf("positive input result=%v, want satisfiable", got)
	}
	if got := solver.Solve(bsat.Positive(constantFalse)); got != bsat.Unsatisfiable {
		t.Fatalf("false output result=%v, want unsatisfiable", got)
	}
}

func TestBmcSolverForCNFRejectsMalformedLiteral(t *testing.T) {
	_, err := bmcSolverForCNF(&cnf.CNF{Variables: 1, Clauses: [][]int32{{2}}})
	if err == nil {
		t.Fatal("expected invalid literal to be rejected")
	}
}
