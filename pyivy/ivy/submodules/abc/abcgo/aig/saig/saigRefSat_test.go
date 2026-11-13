package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSaigManExtendCounterExampleTest3ReturnsOnlyReasonPIs(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	aig.Aig_ManSetRegNum(g, 0)
	cex := util.Abc_CexAlloc(0, 2, 1)
	cex.IPo, cex.IFrame = 0, 0
	util.Abc_InfoSetBit(cex.PData, 0)
	util.Abc_InfoSetBit(cex.PData, 1)

	got := Saig_ManExtendCounterExampleTest3(g, 1, cex, 0)
	if vec.Vec_IntSize(got) != 1 || vec.Vec_IntEntry(got, 0) != 1 {
		t.Fatalf("essential PIs = %v, want [1]", got)
	}
}

func TestSaigManFindCexCareBitsPreservesFrameSpecificReason(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	aig.Aig_ManSetRegNum(g, 0)
	cex := util.Abc_CexAlloc(0, 2, 2)
	cex.IPo, cex.IFrame = 0, 1
	util.Abc_InfoSetBit(cex.PData, 1) // fixed input a is true in frame zero
	util.Abc_InfoSetBit(cex.PData, 2) // symbolic input a is true in frame one
	util.Abc_InfoSetBit(cex.PData, 3) // symbolic input b is true in frame one

	care := Saig_ManFindCexCareBits(g, cex, 1, 0, 0)
	if care == nil {
		t.Fatal("expected a care CEX")
	}
	if util.Abc_InfoHasBit(care.PData, 0) || util.Abc_InfoHasBit(care.PData, 2) {
		t.Fatal("fixed input a should not be returned as a care bit")
	}
	if !util.Abc_InfoHasBit(care.PData, 3) {
		t.Fatal("frame-one input b is required to preserve the property failure")
	}
}
