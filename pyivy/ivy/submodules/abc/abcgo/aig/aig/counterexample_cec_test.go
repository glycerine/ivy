package aig

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCecCounterexampleOwnershipOnSequentialGraph(t *testing.T) {
	g := NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(input)
	g.AddOutput(state)
	Aig_ManSetRegNum(g, 1)
	// The raw CEC witness retains all CIs as PIs, independently of register count.
	cex := util.Abc_CexAlloc(0, 2, 1)
	g.SetCecCounterexample(cex)
	if g.CombinationalCounterexample() != cex {
		t.Fatal("ownership not transferred")
	}
	if g.TakeCombinationalCounterexample() != cex || g.CombinationalCounterexample() != nil {
		t.Fatal("take did not transfer ownership")
	}
	g.SetCecCounterexample(cex)
	g.SetCecCounterexample(nil)
	if g.CombinationalCounterexample() != nil {
		t.Fatal("nil did not clear model")
	}
}
