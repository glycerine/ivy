package int

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestInitialAndAllStateChecks(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, lo.Not()))
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	if Inter_ManCheckInitialState(g) != 1 {
		t.Fatal("expected initial-state counterexample")
	}
	witness := g.SequentialCounterexample()
	if witness == nil || !util.Abc_InfoHasBit(witness.PData, 1) || util.Abc_InfoHasBit(witness.PData, 0) {
		t.Fatalf("invalid initial-state witness %+v", witness)
	}
	if Inter_ManCheckAllStates(g) != 0 {
		t.Fatal("nonzero property reported universally false")
	}
	constant := aig.NewGraph()
	constant.NewInput()
	constant.AddOutput(constant.Const0())
	constant.AddOutput(constant.Const0())
	aig.Aig_ManSetRegNum(constant, 1)
	if Inter_ManCheckInitialState(constant) != 0 || Inter_ManCheckAllStates(constant) != 1 {
		t.Fatal("constant-zero property was not recognized")
	}
}
