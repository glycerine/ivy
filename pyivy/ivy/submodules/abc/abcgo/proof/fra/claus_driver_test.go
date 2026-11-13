package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestClausTopLevelTrivialInduction(t *testing.T) {
	graph := aig.NewGraph()
	graph.NewInput()
	state := graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(graph.Const0())
	aig.Aig_ManSetRegNum(graph, 1)
	// The all-zero register is preserved and the property is one-step inductive.
	if got := Fra_Claus(graph, 1, 0, 32, 4, 8, 8, 1, 0, 1, 0, 1, 0, 0); got != 1 {
		t.Fatalf("got %d", got)
	}
}
