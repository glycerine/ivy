package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestRefinementUnrollSequentialCounterexample(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	util.Abc_InfoSetBit(c.PData, 1)
	p := Saig_RefManStart(g, c, 0, 0)
	if got := vec.Vec_IntArray(p.vMapPiF2A); !reflect.DeepEqual(got, []int32{0, 0}) {
		t.Fatalf("mapping %v", got)
	}
	if Saig_RefManSetPhases(p, nil, 0) != 0 {
		t.Fatal("unrolled negated failing output should be false")
	}
	if got := vec.Vec_IntArray(Saig_RefManFindReason(p)); !reflect.DeepEqual(got, []int32{0}) {
		t.Fatalf("reasons %v", got)
	}
	fixed := Saig_RefManStart(g, c, 1, 0)
	if saig.Saig_ManPiNum(fixed.pFrames) != 0 || Saig_RefManSetPhases(fixed, nil, 0) != 0 {
		t.Fatal("fixed input unrolling failed")
	}
}
