package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCexRemapVerifiesOriginal(t *testing.T) {
	original := aig.NewGraph()
	x := original.NewInput()
	original.AddOutput(x)
	aig.Aig_ManSetRegNum(original, 0)
	abstraction := aig.NewGraph()
	ax, pseudo := abstraction.NewInput(), abstraction.NewInput()
	abstraction.AddOutput(abstraction.OR(ax, pseudo))
	aig.Aig_ManSetRegNum(abstraction, 0)
	cex := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(cex.PData, 0)
	mapped := Saig_ManCexRemap(original, abstraction, cex)
	if mapped == nil || !util.Abc_InfoHasBit(mapped.PData, 0) || mapped.NPis != 1 {
		t.Fatal("failed to remap valid counterexample")
	}
	cex = util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(cex.PData, 1)
	if Saig_ManCexRemap(original, abstraction, cex) != nil {
		t.Fatal("accepted abstraction-only counterexample")
	}
}
