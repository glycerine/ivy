package framgr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
)

// The C model orders PI values by frame and stores the failed CO object ID
// after all unrolled PIs, including those in frames after the failing one.
func TestCopyCounterExampleMultipleInputsAndOutputs(t *testing.T) {
	g := aig.NewGraph()
	a, b, state := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(a.Not())
	g.AddOutput(g.AND(state, b))
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)

	frames := aig.NewGraph()
	for i := 0; i < 6; i++ {
		frames.NewInput()
	}
	var failed aig.Edge
	for i := 0; i < 6; i++ {
		out := frames.AddOutput(frames.Const0())
		if i == 3 {
			failed = out
		} // second output, second frame
	}
	model := []int32{1, 0, 0, 1, 1, 1, int32(failed.NodeID())}
	cex := Fra_SmlCopyCounterExample(g, frames, model)
	if cex == nil || cex.IPo != 1 || cex.IFrame != 1 || cex.NRegs != 1 || cex.NPis != 2 || cex.NBits != 5 {
		t.Fatalf("incorrect counterexample: %+v", cex)
	}
	for i, want := range []bool{false, true, false, false, true} {
		if got := util.Abc_InfoHasBit(cex.PData, int32(i)); got != want {
			t.Fatalf("bit %d = %t, want %t", i, got, want)
		}
	}
	if saig.Saig_ManVerifyCex(g, cex) == 0 {
		t.Fatal("counterexample failed replay")
	}
}
