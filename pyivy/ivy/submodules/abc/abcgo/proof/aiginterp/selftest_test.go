package aiginterp

import "testing"

func TestSolver2SelfComparisonAllTwoInputFunctions(t *testing.T) {
	for truth := uint32(0); truth < 16; truth++ {
		graph := predicate(truth)
		nodes := graph.NodeCount()
		if err := Aig_ManInterTest(graph, 0); err != nil {
			t.Fatalf("truth=%x: %v", truth, err)
		}
		if graph.NodeCount() != nodes || graph.RegisterCount() != 0 {
			t.Fatal("self-test changed input graph")
		}
		for bits := 0; bits < 4; bits++ {
			got, err := graph.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
			if err != nil || got[0] != (truth>>bits&1 != 0) {
				t.Fatalf("truth=%x bits=%d got=%v err=%v", truth, bits, got, err)
			}
		}
	}
}
