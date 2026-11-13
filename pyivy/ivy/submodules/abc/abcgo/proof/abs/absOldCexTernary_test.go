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
	aig.Aig_ManSetRegNum(g, 0)
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

// absOldCex.c:655-665 leaves iBit at zero throughout the register loop.
// Later initial-register bits are therefore ignored by the C assertion.
func TestCexTernarySourceChecksOnlyFirstInitialBit(t *testing.T) {
	g := aig.NewGraph()
	q, r := g.NewInput(), g.NewInput()
	g.AddOutput(g.Const1())
	g.AddOutput(q)
	g.AddOutput(r)
	aig.Aig_ManSetRegNum(g, 2)
	cex, care := util.Abc_CexAlloc(2, 0, 1), util.Abc_CexAlloc(2, 0, 1)
	util.Abc_InfoSetBit(cex.PData, 1)
	util.Abc_InfoSetBit(care.PData, 1)
	if got := Saig_ManCexVerifyUsingTernary(g, cex, care); got != 1 {
		t.Fatalf("source result %d, want true", got)
	}
	util.Abc_InfoSetBit(cex.PData, 0)
	defer func() {
		if recover() == nil {
			t.Fatal("source assertion must reject the first initial bit")
		}
	}()
	Saig_ManCexVerifyUsingTernary(g, cex, care)
}
