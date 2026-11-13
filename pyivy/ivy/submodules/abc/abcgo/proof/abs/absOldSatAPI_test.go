package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestOldRefinementAPISelectsCounterexampleReasons(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.OR(x, y))
	aig.Aig_ManSetRegNum(g, 0)
	c := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(c.PData, 0)
	util.Abc_InfoSetBit(c.PData, 1)
	care := Saig_ManFindCexCareBits(g, c, 0, 0, 0)
	if care == nil || !util.Abc_InfoHasBit(care.PData, 0) || util.Abc_InfoHasBit(care.PData, 1) {
		t.Fatal("expected the first sufficient reason")
	}
	got := vec.Vec_IntArray(Saig_ManExtendCounterExampleTest3(g, 0, c, 0))
	if !reflect.DeepEqual(got, []int32{0}) {
		t.Fatalf("reasons: %v", got)
	}
}
