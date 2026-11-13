package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestCbaShrinkDerivesVisitedStateCube(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(0, 1, 2)
	c.IFrame = 1
	util.Abc_InfoSetBit(c.PData, 0)
	p := Saig_ManCbaStart(g, c, 0, 0)
	p.vReg2Frame = vec.Vec_VecAlloc(2)
	p.pFrames = Saig_ManCbaUnrollWithCex(g, c, 0, &p.vMapPiF2A, &p.vReg2Frame)
	Saig_ManCbaFindReason(p)
	Saig_ManCbaShrink(p)
	if got := vec.Vec_IntArray(vec.Vec_VecEntryInt(p.vReg2Value, 0)); !reflect.DeepEqual(got, []int32{0}) {
		t.Fatalf("state cube %v", got)
	}
}
