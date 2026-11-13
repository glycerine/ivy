package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCopyCounterexampleCrossesWordBoundary(t *testing.T) {
	g := aig.NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(g.AND(a, b))
	aig.Aig_ManSetRegNum(g, 1)
	frames := aig.NewGraph()
	for i := 0; i < 40; i++ {
		frames.NewInput()
	}
	var failed aig.Edge
	for i := 0; i < 20; i++ {
		edge := frames.AddOutput(frames.Const0())
		if i == 19 {
			failed = edge
		}
	}
	model := make([]int32, 41)
	model[36], model[37], model[40] = 1, 1, int32(failed.NodeID())
	cex := Fra_SmlCopyCounterExample(g, frames, model)
	if cex == nil || cex.IFrame != 19 || cex.IPo != 0 || cex.NBits != 41 {
		t.Fatalf("counterexample: %+v", cex)
	}
	for i := int32(0); i < cex.NBits; i++ {
		want := i == 37 || i == 38
		if got := util.Abc_InfoHasBit(cex.PData, i); got != want {
			t.Fatalf("bit %d = %v want %v", i, got, want)
		}
	}
	if saig.Saig_ManVerifyCexNoClear(g, cex) != 1 {
		t.Fatal("multiword counterexample does not replay")
	}
}
