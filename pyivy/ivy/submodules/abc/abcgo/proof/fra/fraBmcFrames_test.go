package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestInitializedTimeframes(t *testing.T) {
	source := aig.NewGraph()
	input, latch := source.NewInput(), source.NewInput()
	source.AddOutput(latch)
	source.AddOutput(input)
	aig.Aig_ManSetRegNum(source, 1)
	graph, mapping := fraUnrollInitialized(source, 3, true)
	if graph.InputCount() != 3 || graph.OutputCount() != 3 {
		t.Fatalf("unexpected unrolled interface %d/%d", graph.InputCount(), graph.OutputCount())
	}
	for bits := 0; bits < 8; bits++ {
		values := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, err := graph.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{false, values[0], values[1]}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("inputs %v frame %d = %v, want %v", values, i, got[i], want[i])
			}
		}
	}
	if mapping[latch.NodeID()*3] != graph.Const0() {
		t.Fatal("initial latch was not zero")
	}
}
