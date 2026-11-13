package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestFilterFlopsUsesConcretePrimaryInputs(t *testing.T) {
	g := aig.NewGraph()
	pi, lo0, lo1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(pi)
	g.AddOutput(lo1)
	_ = lo0
	aig.Aig_ManSetRegNum(g, 2)
	cex := util.Abc_CexAlloc(0, 3, 2)
	cex.IFrame = 1
	util.Abc_InfoSetBit(cex.PData, 0) // first frame's real PI drives latch 0 high
	classes := vec.Vec_IntAllocArrayCopy([]int32{0, 0}, 2)
	candidates := vec.Vec_IntAllocArrayCopy([]int32{1, 0}, 2)
	result := Saig_ManCbaFilterFlops(g, cex, classes, candidates, 1)
	if got := vec.Vec_IntArray(result); !reflect.DeepEqual(got, []int32{0}) {
		t.Fatalf("selected %v; expected the inconsistent first latch", got)
	}
}
