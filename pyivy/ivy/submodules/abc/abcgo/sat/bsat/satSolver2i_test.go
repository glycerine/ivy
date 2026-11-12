package bsat

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSolver2OnlineInterpolant(t *testing.T) {
	s := sat_solver2_new()
	sat_solver2_setnvars(s, 3)
	for i, c := range [][]lit{{0, 2}, {1, 2}, {3, 4}, {3, 5}} {
		h := sat_solver2_addclause(s, c, int32(i+1))
		if i < 2 {
			clause2_set_partA(s, h, 1)
		}
	}
	var_set_partA(s, 0, 1)
	manager := Int2_ManStart(s, []int32{1}, 1)
	s.pInt2 = manager
	if got := sat_solver2_solve(s, nil, 0, 0, 0, 0); got != l_False {
		t.Fatal(got)
	}
	graph := Int2_ManReadInterpolant(s)
	if graph == nil || graph.InputCount() != 1 || graph.OutputCount() != 1 {
		t.Fatal("invalid interpolant graph")
	}
	for _, value := range []bool{false, true} {
		out, err := graph.Evaluate([]bool{value})
		if err != nil || out[0] != value {
			t.Fatal(out, value, err)
		}
	}
	if Int2_ManReadInterpolant(s) != nil {
		t.Fatal("result ownership transferred twice")
	}
}

func TestCircuitRelationInterpolationHarness(t *testing.T) {
	for kind := 0; kind < 4; kind++ {
		graph := aig.NewGraph()
		a, b := graph.NewInput(), graph.NewInput()
		driver := graph.AND(a, b)
		switch kind {
		case 1:
			driver = graph.XOR(a, b)
		case 2:
			driver = graph.Const1()
		case 3:
			driver = graph.OR(a.Not(), b)
		}
		graph.AddOutput(driver)
		interpolant := AigInterpolationTest(graph)
		if interpolant == nil || interpolant.InputCount() != 3 {
			t.Fatal("missing relation", kind)
		}
		for model := 0; model < 8; model++ {
			values := []bool{model&1 != 0, model&2 != 0, model&4 != 0}
			expected, err := graph.Evaluate(values[:2])
			if err != nil {
				t.Fatal(err)
			}
			actual, err := interpolant.Evaluate(values)
			if err != nil || actual[0] != (expected[0] == values[2]) {
				t.Fatal(kind, values, expected, actual, err)
			}
		}
	}
}
