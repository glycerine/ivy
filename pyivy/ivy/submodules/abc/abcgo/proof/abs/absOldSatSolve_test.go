package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRefinementSatCareForAnd(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, y))
	c := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(c.PData, 0)
	util.Abc_InfoSetBit(c.PData, 1)
	for _, order := range []int32{0, 1} {
		p := Saig_RefManStart(g, c, 0, 0)
		care := Saig_RefManRunSat(p, order)
		if care == nil || !util.Abc_InfoHasBit(care.PData, 0) || !util.Abc_InfoHasBit(care.PData, 1) {
			t.Fatalf("order %d: missing care bits: %#v", order, care)
		}
	}
}
func TestRefinementSatRejectsConstantFalseFormula(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	c := util.Abc_CexAlloc(0, 1, 1)
	util.Abc_InfoSetBit(c.PData, 0)
	p := Saig_RefManStart(g, c, 1, 0)
	care := Saig_RefManRunSat(p, 0)
	if care != nil {
		t.Fatal("constant-false unrolled formula is rejected during solver setup")
	}
}

func TestRefinementSatDropsRedundantInput(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.OR(x, y))
	c := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(c.PData, 0)
	util.Abc_InfoSetBit(c.PData, 1)
	p := Saig_RefManStart(g, c, 0, 0)
	inputs := vec.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
	reasons := vec.Vec_IntArray(Saig_RefManRefineWithSat(p, inputs))
	if len(reasons) != 1 {
		t.Fatalf("expected one sufficient input, got %v", reasons)
	}
	care := Saig_RefManReason2Cex(p, vec.Vec_IntAllocArrayCopy(reasons, len(reasons)))
	if Saig_RefManSetPhases(p, care, 0) != 0 {
		t.Fatal("retained reason does not force failure")
	}
}
