package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"reflect"
	"testing"
)

func TestGiaLegacyRefinementUpdatesClasses(t *testing.T) {
	graph := aig.NewGraph()
	x, lo0, lo1 := graph.NewInput(), graph.NewInput(), graph.NewInput()
	graph.AddOutput(graph.AND(x, graph.OR(lo0, lo1)))
	graph.AddOutput(lo0)
	graph.AddOutput(lo1)
	aig.Aig_ManSetRegNum(graph, 2)
	g := gia.Gia_ManFromAig(graph)
	cex := util.Abc_CexAlloc(1, 2, 1)
	util.Abc_InfoSetBit(cex.PData, 1)
	util.Abc_InfoSetBit(cex.PData, 2)
	if Gia_ManCexAbstractionRefine(g, cex, 0, 0, 0, 0) != -1 {
		t.Fatal("missing class map should fail")
	}
	if err := g.SetAbstractionFlopClasses([]int32{1, 0}); err != nil {
		t.Fatal(err)
	}
	if Gia_ManCexAbstractionRefine(g, cex, 0, 0, 0, 0) != -1 {
		t.Fatal("expected refinement")
	}
	if got := g.AbstractionFlopClasses(); !reflect.DeepEqual(got, []int32{1, 1}) {
		t.Fatalf("classes %v", got)
	}
}
