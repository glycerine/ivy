package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"reflect"
	"sort"
	"testing"
)

func TestOutputSupportAndOverlap(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
	graph.AddOutput(graph.AND(a, b))
	graph.AddOutput(graph.AND(b, c))
	supports := fraOutputSupports(graph)
	for _, support := range supports {
		sort.Ints(support)
	}
	if !reflect.DeepEqual(supports, [][]int{{0, 1}, {1, 2}}) {
		t.Fatalf("supports = %v", supports)
	}
	if got := fraSupportOverlap(supports[0], supports[1]); got != 1 {
		t.Fatalf("overlap = %d", got)
	}
	graph.Assertions = 1
	Fra_ManPartitionTest(graph, 1)
	Fra_ManPartitionTest2(graph)
}

// Both source diagnostics call Aig_ManSupports, which clears the terminal
// pNext/CioId union after computing supports.
func TestPartitionDiagnosticsClearTerminalCioIDs(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*aig.Graph)
	}{{"overlap", func(g *aig.Graph) { Fra_ManPartitionTest(g, 1) }}, {"reverse", Fra_ManPartitionTest2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			g.AddOutput(a)
			g.AddOutput(b)
			aig.Aig_ManSetCioIds(g)
			if aig.Aig_ObjCioId(g, b) != 1 || aig.Aig_ObjCioId(g, g.Outputs()[1]) != 1 {
				t.Fatal("fixture lacks nonzero IDs")
			}
			tc.run(g)
			for _, e := range append(g.Inputs(), g.Outputs()...) {
				if got := aig.Aig_ObjCioId(g, e); got != 0 {
					t.Fatalf("terminal %v CioId=%d, want source-cleared zero", e, got)
				}
			}
		})
	}
}
