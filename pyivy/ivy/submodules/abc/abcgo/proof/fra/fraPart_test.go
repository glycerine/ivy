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
