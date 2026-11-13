package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCexTernaryCareBits(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, y))
	cex, care := util.Abc_CexAlloc(0, 2, 1), util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(cex.PData, 0)
	util.Abc_InfoSetBit(cex.PData, 1)
	util.Abc_InfoSetBit(care.PData, 0)
	if Saig_ManCexVerifyUsingTernary(g, cex, care) != 0 {
		t.Fatal("unknown second input cannot force AND true")
	}
	util.Abc_InfoSetBit(care.PData, 1)
	if Saig_ManCexVerifyUsingTernary(g, cex, care) != 1 {
		t.Fatal("both true care bits must force AND true")
	}
}
func TestCexTernarySequential(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	cex, care := util.Abc_CexAlloc(1, 1, 2), util.Abc_CexAlloc(1, 1, 2)
	cex.IFrame = 1
	care.IFrame = 1
	util.Abc_InfoSetBit(cex.PData, 1)
	util.Abc_InfoSetBit(care.PData, 1)
	if Saig_ManCexVerifyUsingTernary(g, cex, care) != 1 {
		t.Fatal("cared input should force next-frame latch")
	}
}
