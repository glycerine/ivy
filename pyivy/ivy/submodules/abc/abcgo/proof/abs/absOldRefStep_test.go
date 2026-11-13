package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestLegacyRefinementAddsNeededFlop(t *testing.T) {
	g := aig.NewGraph()
	x, lo0, lo1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, g.OR(lo0, lo1)))
	g.AddOutput(lo0)
	g.AddOutput(lo1)
	aig.Aig_ManSetRegNum(g, 2)
	flops := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	classes := vec.Vec_IntAllocArrayCopy([]int32{1, 0}, 2)
	cex := util.Abc_CexAlloc(1, 2, 1)
	util.Abc_InfoSetBit(cex.PData, 1)
	util.Abc_InfoSetBit(cex.PData, 2)
	if Saig_ManCexRefineStep(g, flops, classes, cex, 0, 0, 0, 0) != 1 {
		t.Fatal("expected refinement")
	}
	if got := vec.Vec_IntArray(flops); !reflect.DeepEqual(got, []int32{0, 1}) {
		t.Fatalf("flops %v", got)
	}
}
