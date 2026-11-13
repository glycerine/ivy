package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bmc"
	"reflect"
	"testing"
)

func TestOldCbaAPISelectsPseudoInputs(t *testing.T) {
	g := aig.NewGraph()
	real, pseudo := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(real, pseudo))
	c := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(c.PData, 0)
	util.Abc_InfoSetBit(c.PData, 1)
	care := Saig_ManCbaFindCexCareBits(g, c, 1, 0)
	if care == nil || util.Abc_InfoHasBit(care.PData, 0) || !util.Abc_InfoHasBit(care.PData, 1) {
		t.Fatal("expected pseudo-input care")
	}
	if got := vec.Vec_IntArray(Saig_ManCbaFilterInputs(g, 1, c, 0)); !reflect.DeepEqual(got, []int32{0}) {
		t.Fatalf("pseudo-inputs %v", got)
	}
}
func TestOldCbaPerformFindsPseudoInput(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	var params bmc.Saig_ParBmc_t
	bmc.Saig_ParBmcSetDefaultParams(&params)
	if got := vec.Vec_IntArray(Saig_ManCbaPerform(g, 0, &params)); !reflect.DeepEqual(got, []int32{0}) {
		t.Fatalf("pseudo-inputs %v", got)
	}
}
