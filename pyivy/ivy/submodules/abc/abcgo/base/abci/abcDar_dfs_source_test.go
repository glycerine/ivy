package abci

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestDarNodeWalkInitializesTraversalAndSkipsDeletedSlots(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	g.AddOutput(root.Not())
	dead := g.AND(a, c)
	aig.Aig_ManCleanup(g)
	if g.ValidEdge(dead) {
		t.Fatal("fixture did not delete dangling gate")
	}
	for _, walk := range []func(*aig.Graph) []aig.Edge{darReachableNodes, darAllNodes} {
		for run := 0; run < 2; run++ {
			if got := walk(g); !slices.Equal(got, []aig.Edge{root}) {
				t.Fatalf("run %d: nodes = %v, want %v", run, got, []aig.Edge{root})
			}
		}
	}
}
