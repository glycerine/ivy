package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestRefinementReasonChoosesControllingInput(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, y))
	c := util.Abc_CexAlloc(0, 2, 1)
	p := &Saig_RefMan_t{pFrames: g, pCex: c, vMapPiF2A: vec.Vec_IntAllocArrayCopy([]int32{0, 0, 1, 0}, 4)}
	if got := vec.Vec_IntArray(Saig_RefManFindReason(p)); !reflect.DeepEqual(got, []int32{0}) {
		t.Fatalf("equal priorities: %v", got)
	}
	util.Abc_InfoSetBit(c.PData, 0)
	if got := vec.Vec_IntArray(Saig_RefManFindReason(p)); !reflect.DeepEqual(got, []int32{1}) {
		t.Fatalf("controlling zero: %v", got)
	}
}
